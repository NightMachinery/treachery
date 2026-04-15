package app

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	mathrand "math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
	ErrBadInput  = errors.New("bad input")
)

type Config struct {
	DistDir   string
	DataDir   string
	CardsPath string
}

type App struct {
	db      *sql.DB
	distDir string
	cards   CardsResource
	hub     *Hub
	rand    *mathrand.Rand
	timeNow func() time.Time
}

func New(cfg Config) (*App, error) {
	if cfg.DataDir == "" {
		return nil, fmt.Errorf("data dir is required")
	}
	if cfg.CardsPath == "" {
		return nil, fmt.Errorf("cards path is required")
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(cfg.DataDir, "treachery.sqlite")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode = WAL; PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000;`); err != nil {
		_ = db.Close()
		return nil, err
	}

	cardsBytes, err := os.ReadFile(cfg.CardsPath)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	var cards CardsResource
	if err := json.Unmarshal(cardsBytes, &cards); err != nil {
		_ = db.Close()
		return nil, err
	}

	a := &App{
		db:      db,
		distDir: cfg.DistDir,
		cards:   cards,
		hub:     NewHub(),
		rand:    mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
		timeNow: time.Now,
	}
	if err := a.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return a, nil
}

func (a *App) Close() error {
	if a.db == nil {
		return nil
	}
	return a.db.Close()
}

func (a *App) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			uid TEXT NOT NULL,
			created_timestamp TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS games (
			game_id TEXT PRIMARY KEY,
			creator_uid TEXT NOT NULL,
			created_timestamp TEXT NOT NULL,
			started_on TEXT,
			murderer_cards_selected INTEGER NOT NULL DEFAULT 0,
			murderer_uid TEXT,
			murderer_clue_card_name TEXT,
			murderer_means_card_name TEXT,
			cause_card_json TEXT,
			location_card_json TEXT,
			other_cards_json TEXT NOT NULL DEFAULT '[]',
			finished INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS players (
			game_id TEXT NOT NULL,
			uid TEXT NOT NULL,
			name TEXT NOT NULL,
			clue_cards_json TEXT NOT NULL DEFAULT '[]',
			means_cards_json TEXT NOT NULL DEFAULT '[]',
			PRIMARY KEY (game_id, uid),
			FOREIGN KEY (game_id) REFERENCES games(game_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS guesses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id TEXT NOT NULL,
			guessed_by_uid TEXT NOT NULL,
			murderer_uid TEXT NOT NULL,
			means_card_name TEXT NOT NULL,
			clue_card_name TEXT NOT NULL,
			correct INTEGER NOT NULL,
			created_timestamp TEXT NOT NULL,
			FOREIGN KEY (game_id) REFERENCES games(game_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id TEXT NOT NULL,
			player_uid TEXT,
			message TEXT NOT NULL,
			type TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			FOREIGN KEY (game_id) REFERENCES games(game_id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_players_game_id ON players(game_id);`,
		`CREATE INDEX IF NOT EXISTS idx_guesses_game_id ON guesses(game_id);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_game_id ON messages(game_id, id);`,
	}

	for _, stmt := range stmts {
		if _, err := a.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func randomToken(bytesLen int) (string, error) {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func randomReadableID(rng *mathrand.Rand, length int) string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	result := make([]rune, length)
	for i := range result {
		result[i] = chars[rng.Intn(len(chars))]
	}
	return string(result)
}

func newUID() (string, error) {
	token, err := randomToken(16)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("u-%s", token), nil
}

func (a *App) EnsureSession(existingToken string) (*Session, error) {
	if existingToken != "" {
		session, err := a.GetSession(existingToken)
		if err == nil {
			return session, nil
		}
		if !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}

	uid, err := newUID()
	if err != nil {
		return nil, err
	}
	token, err := randomToken(24)
	if err != nil {
		return nil, err
	}
	session := &Session{Token: token, UID: uid, CreatedTimestamp: nowTimestamp(a.timeNow())}
	_, err = a.db.Exec(`INSERT INTO sessions(token, uid, created_timestamp) VALUES (?, ?, ?)`, session.Token, session.UID, session.CreatedTimestamp)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (a *App) GetSession(token string) (*Session, error) {
	row := a.db.QueryRow(`SELECT token, uid, created_timestamp FROM sessions WHERE token = ?`, token)
	var session Session
	if err := row.Scan(&session.Token, &session.UID, &session.CreatedTimestamp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (a *App) CreateGame(creatorUID, gameID string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	if gameID == "" {
		return fmt.Errorf("%w: game ID is required", ErrBadInput)
	}
	createdAt := nowTimestamp(a.timeNow())
	_, err := a.db.Exec(`INSERT INTO games(game_id, creator_uid, created_timestamp, other_cards_json, finished, murderer_cards_selected) VALUES (?, ?, ?, '[]', 0, 0)`, gameID, creatorUID, createdAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return fmt.Errorf("%w: game already exists", ErrBadInput)
		}
		return err
	}
	return nil
}

func (a *App) AddPlayer(gameID, uid, playerName string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	playerName = strings.TrimSpace(playerName)
	if playerName == "" {
		return fmt.Errorf("%w: player name is required", ErrBadInput)
	}
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID == uid {
			return fmt.Errorf("%w: creator cannot join as player", ErrForbidden)
		}

		existing, err := a.getPlayerTx(tx, gameID, uid)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
		if game.StartedOn != "" && errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: game already started", ErrBadInput)
		}
		if existing != nil {
			_, err = tx.Exec(`UPDATE players SET name = ? WHERE game_id = ? AND uid = ?`, playerName, gameID, uid)
			return err
		}
		_, err = tx.Exec(`INSERT INTO players(game_id, uid, name, clue_cards_json, means_cards_json) VALUES (?, ?, ?, '[]', '[]')`, gameID, uid, playerName)
		return err
	})
}

func (a *App) ListGames() ([]Game, error) {
	rows, err := a.db.Query(`SELECT game_id, creator_uid, created_timestamp, started_on, murderer_cards_selected, murderer_uid, murderer_clue_card_name, murderer_means_card_name, cause_card_json, location_card_json, other_cards_json, finished FROM games ORDER BY created_timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		game, err := scanGame(rows)
		if err != nil {
			return nil, err
		}
		games = append(games, *game)
	}
	return games, rows.Err()
}

func (a *App) GetSnapshot(gameID, viewerUID string) (*GameSnapshot, error) {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	game, err := a.getGame(gameID)
	if err != nil {
		return nil, err
	}
	players, err := a.getPlayers(gameID)
	if err != nil {
		return nil, err
	}
	guesses, err := a.getGuesses(gameID)
	if err != nil {
		return nil, err
	}
	messages, err := a.getMessages(gameID)
	if err != nil {
		return nil, err
	}

	snapshot := &GameSnapshot{
		Game:              game,
		Players:           players,
		Guesses:           guesses,
		Messages:          messages,
		PlayerPrivateData: PlayerPrivateData{},
	}

	if viewerUID == "" {
		return snapshot, nil
	}

	if game.MurdererUID != "" && game.MurdererUID == viewerUID {
		snapshot.PlayerPrivateData = PlayerPrivateData{
			IsMurderer:    true,
			ClueCardName:  game.MurdererClueCardName,
			MeansCardName: game.MurdererMeansCardName,
		}
	}

	if game.CreatorUID == viewerUID {
		privateData := &ForensicPrivateData{
			MurdererClueCardName:  game.MurdererClueCardName,
			MurdererMeansCardName: game.MurdererMeansCardName,
		}
		if murderer := findPlayer(players, game.MurdererUID); murderer != nil {
			copyPlayer := *murderer
			privateData.Murderer = &copyPlayer
		}
		snapshot.ForensicPrivateData = privateData
	}

	return snapshot, nil
}

func (a *App) StartGame(gameID, creatorUID string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != creatorUID {
			return fmt.Errorf("%w: only the forensic scientist can start the game", ErrForbidden)
		}
		if game.StartedOn != "" {
			return fmt.Errorf("%w: game already started", ErrBadInput)
		}
		players, err := a.getPlayersTx(tx, gameID)
		if err != nil {
			return err
		}
		if len(players) < 3 {
			return fmt.Errorf("%w: need at least 3 players to start", ErrBadInput)
		}
		if len(a.cards.ForensicCards.OtherCards) < 6 {
			return fmt.Errorf("%w: missing forensic cards", ErrBadInput)
		}

		otherCards := cloneForensicCards(sampleRandom(a.rand, a.cards.ForensicCards.OtherCards, 6))
		clueCards := cloneCards(sampleRandom(a.rand, a.cards.ClueCards, len(players)*4))
		meansCards := cloneCards(sampleRandom(a.rand, a.cards.MeansCards, len(players)*4))

		for i := range players {
			players[i].ClueCards = cloneCards(clueCards[i*4 : (i+1)*4])
			players[i].MeansCards = cloneCards(meansCards[i*4 : (i+1)*4])
			if err := a.savePlayerCardsTx(tx, gameID, players[i]); err != nil {
				return err
			}
		}

		murderer := players[a.rand.Intn(len(players))]
		startedOn := nowTimestamp(a.timeNow())
		game.StartedOn = startedOn
		game.StartedTimestamp = startedOn
		game.MurdererUID = murderer.UID
		game.MurdererSelected = true
		game.MurdererCardsSelected = false
		game.OtherCards = otherCards
		if err := a.saveGameTx(tx, game); err != nil {
			return err
		}
		if err := a.sendForensicMessageTx(tx, gameID, "Game started! Murderer, select your cards. Don't let anyone else find out!"); err != nil {
			return err
		}
		return nil
	})
}

func (a *App) SelectMurdererCards(gameID, uid, clueCardName, meansCardName string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.MurdererUID == "" || game.MurdererUID != uid {
			return fmt.Errorf("%w: only the murderer can select cards", ErrForbidden)
		}
		player, err := a.getPlayerTx(tx, gameID, uid)
		if err != nil {
			return err
		}
		if !hasCard(player.ClueCards, clueCardName) || !hasCard(player.MeansCards, meansCardName) {
			return fmt.Errorf("%w: selected cards do not belong to this player", ErrBadInput)
		}

		game.MurdererClueCardName = clueCardName
		game.MurdererMeansCardName = meansCardName
		game.MurdererCardsSelected = true
		if err := a.saveGameTx(tx, game); err != nil {
			return err
		}
		return a.sendForensicMessageTx(tx, gameID, "Murderer selected their cards. Time to figure out who they are. Good luck!")
	})
}

func (a *App) SelectForensicCauseCard(gameID, uid string, card ForensicCard) error {
	return a.updateForensicCard(gameID, uid, func(game *Game) error {
		game.CauseCard = &card
		return nil
	})
}

func (a *App) SelectForensicLocationCard(gameID, uid string, card ForensicCard) error {
	return a.updateForensicCard(gameID, uid, func(game *Game) error {
		game.LocationCard = &card
		return nil
	})
}

func (a *App) SelectForensicOtherCard(gameID, uid string, card ForensicCard, replaceCardName string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != uid {
			return fmt.Errorf("%w: only the forensic scientist can update clue cards", ErrForbidden)
		}

		selectedCount := 0
		newCardIndex := -1
		for i := range game.OtherCards {
			if game.OtherCards[i].SelectedChoice != "" {
				selectedCount++
			}
			if game.OtherCards[i].CardName == card.CardName {
				newCardIndex = i
			}
		}
		if newCardIndex == -1 {
			return fmt.Errorf("%w: card not found", ErrBadInput)
		}
		if selectedCount >= 4 {
			if replaceCardName == "" {
				return fmt.Errorf("%w: replacement card required", ErrBadInput)
			}
			replaceIndex := -1
			for i := range game.OtherCards {
				if game.OtherCards[i].CardName == replaceCardName {
					replaceIndex = i
					break
				}
			}
			if replaceIndex == -1 {
				return fmt.Errorf("%w: replacement card not found", ErrBadInput)
			}
			game.OtherCards[replaceIndex].Replaced = true
		}
		game.OtherCards[newCardIndex] = card
		return a.saveGameTx(tx, game)
	})
}

func (a *App) MakeGuess(gameID, uid, murdererUID, clueCardName, meansCardName string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID == uid {
			return fmt.Errorf("%w: forensic scientist cannot guess", ErrForbidden)
		}
		if game.Finished {
			return fmt.Errorf("%w: game already finished", ErrBadInput)
		}
		if !game.MurdererCardsSelected {
			return fmt.Errorf("%w: murderer has not selected cards yet", ErrBadInput)
		}
		if _, err := a.getPlayerTx(tx, gameID, uid); err != nil {
			return fmt.Errorf("%w: only players can guess", ErrForbidden)
		}
		var existingCount int
		if err := tx.QueryRow(`SELECT COUNT(1) FROM guesses WHERE game_id = ? AND guessed_by_uid = ?`, gameID, uid).Scan(&existingCount); err != nil {
			return err
		}
		if existingCount > 0 {
			return fmt.Errorf("%w: each player only gets one guess", ErrBadInput)
		}
		guessedPlayer, err := a.getPlayerTx(tx, gameID, murdererUID)
		if err != nil {
			return err
		}

		correct := game.MurdererUID == murdererUID && game.MurdererClueCardName == clueCardName && game.MurdererMeansCardName == meansCardName
		created := nowTimestamp(a.timeNow())
		if _, err := tx.Exec(`INSERT INTO guesses(game_id, guessed_by_uid, murderer_uid, means_card_name, clue_card_name, correct, created_timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`, gameID, uid, murdererUID, meansCardName, clueCardName, boolToInt(correct), created); err != nil {
			return err
		}
		guessMessage := fmt.Sprintf("I think '%s' is the murderer, with clue '%s' and means '%s'.", guessedPlayer.Name, clueCardName, meansCardName)
		if err := a.sendMessageTx(tx, gameID, uid, guessMessage, MessageTypeGuess); err != nil {
			return err
		}
		if err := a.sendForensicMessageTx(tx, gameID, map[bool]string{true: "That is correct!", false: "Nope."}[correct]); err != nil {
			return err
		}
		if correct {
			game.Finished = true
			if err := a.saveGameTx(tx, game); err != nil {
				return err
			}
			return nil
		}
		return a.checkAndEndGameTx(tx, game)
	})
}

func (a *App) SendChatMessage(gameID, uid, message string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	message = strings.TrimSpace(message)
	if message == "" {
		return fmt.Errorf("%w: message is required", ErrBadInput)
	}
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != uid {
			if _, err := a.getPlayerTx(tx, gameID, uid); err != nil {
				return fmt.Errorf("%w: only game participants can send messages", ErrForbidden)
			}
		}
		return a.sendMessageTx(tx, gameID, uid, message, MessageTypeChat)
	})
}

func (a *App) EndGame(gameID, uid string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != uid {
			return fmt.Errorf("%w: only the forensic scientist can end the game", ErrForbidden)
		}
		game.Finished = true
		return a.saveGameTx(tx, game)
	})
}

func (a *App) updateForensicCard(gameID, uid string, mutator func(*Game) error) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != uid {
			return fmt.Errorf("%w: only the forensic scientist can update clue cards", ErrForbidden)
		}
		if err := mutator(game); err != nil {
			return err
		}
		return a.saveGameTx(tx, game)
	})
}

func (a *App) checkAndEndGameTx(tx *sql.Tx, game *Game) error {
	players, err := a.getPlayersTx(tx, game.GameID)
	if err != nil {
		return err
	}
	rows, err := tx.Query(`SELECT guessed_by_uid FROM guesses WHERE game_id = ?`, game.GameID)
	if err != nil {
		return err
	}
	defer rows.Close()

	guessedBy := map[string]struct{}{}
	for rows.Next() {
		var guessedByUID string
		if err := rows.Scan(&guessedByUID); err != nil {
			return err
		}
		guessedBy[guessedByUID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, player := range players {
		if player.UID == game.MurdererUID {
			continue
		}
		if _, ok := guessedBy[player.UID]; !ok {
			return nil
		}
	}
	game.Finished = true
	return a.saveGameTx(tx, game)
}

func (a *App) getGame(gameID string) (*Game, error) {
	row := a.db.QueryRow(`SELECT game_id, creator_uid, created_timestamp, started_on, murderer_cards_selected, murderer_uid, murderer_clue_card_name, murderer_means_card_name, cause_card_json, location_card_json, other_cards_json, finished FROM games WHERE game_id = ?`, gameID)
	game, err := scanGame(row)
	if err != nil {
		return nil, err
	}
	return game, nil
}

func (a *App) getGameTx(tx *sql.Tx, gameID string) (*Game, error) {
	row := tx.QueryRow(`SELECT game_id, creator_uid, created_timestamp, started_on, murderer_cards_selected, murderer_uid, murderer_clue_card_name, murderer_means_card_name, cause_card_json, location_card_json, other_cards_json, finished FROM games WHERE game_id = ?`, gameID)
	game, err := scanGame(row)
	if err != nil {
		return nil, err
	}
	return game, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanGame(scanner rowScanner) (*Game, error) {
	var (
		gameID, creatorUID, createdTimestamp                                string
		startedOn, murdererUID, murdererClueCardName, murdererMeansCardName sql.NullString
		causeCardJSON, locationCardJSON                                     sql.NullString
		otherCardsJSON                                                      string
		murdererCardsSelected                                               int
		finished                                                            int
	)
	if err := scanner.Scan(&gameID, &creatorUID, &createdTimestamp, &startedOn, &murdererCardsSelected, &murdererUID, &murdererClueCardName, &murdererMeansCardName, &causeCardJSON, &locationCardJSON, &otherCardsJSON, &finished); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	game := &Game{
		GameID:                gameID,
		CreatorUID:            creatorUID,
		CreatedTimestamp:      createdTimestamp,
		StartedTimestamp:      startedOn.String,
		StartedOn:             startedOn.String,
		MurdererCardsSelected: murdererCardsSelected == 1,
		MurdererSelected:      murdererUID.Valid,
		MurdererUID:           murdererUID.String,
		MurdererClueCardName:  murdererClueCardName.String,
		MurdererMeansCardName: murdererMeansCardName.String,
		Finished:              finished == 1,
	}
	if otherCardsJSON != "" {
		if err := json.Unmarshal([]byte(otherCardsJSON), &game.OtherCards); err != nil {
			return nil, err
		}
	}
	if causeCardJSON.Valid && causeCardJSON.String != "" {
		var card ForensicCard
		if err := json.Unmarshal([]byte(causeCardJSON.String), &card); err != nil {
			return nil, err
		}
		game.CauseCard = &card
	}
	if locationCardJSON.Valid && locationCardJSON.String != "" {
		var card ForensicCard
		if err := json.Unmarshal([]byte(locationCardJSON.String), &card); err != nil {
			return nil, err
		}
		game.LocationCard = &card
	}
	return game, nil
}

func (a *App) getPlayers(gameID string) ([]Player, error) {
	rows, err := a.db.Query(`SELECT uid, name, clue_cards_json, means_cards_json FROM players WHERE game_id = ? ORDER BY rowid ASC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlayers(rows)
}

func (a *App) getPlayersTx(tx *sql.Tx, gameID string) ([]Player, error) {
	rows, err := tx.Query(`SELECT uid, name, clue_cards_json, means_cards_json FROM players WHERE game_id = ? ORDER BY rowid ASC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlayers(rows)
}

func scanPlayers(rows *sql.Rows) ([]Player, error) {
	players := []Player{}
	for rows.Next() {
		var (
			uid, name                     string
			clueCardsJSON, meansCardsJSON string
		)
		if err := rows.Scan(&uid, &name, &clueCardsJSON, &meansCardsJSON); err != nil {
			return nil, err
		}
		player := Player{Name: name, UID: uid, ClueCards: []Card{}, MeansCards: []Card{}}
		if clueCardsJSON != "" {
			if err := json.Unmarshal([]byte(clueCardsJSON), &player.ClueCards); err != nil {
				return nil, err
			}
		}
		if meansCardsJSON != "" {
			if err := json.Unmarshal([]byte(meansCardsJSON), &player.MeansCards); err != nil {
				return nil, err
			}
		}
		players = append(players, player)
	}
	return players, rows.Err()
}

func (a *App) getPlayerTx(tx *sql.Tx, gameID, uid string) (*Player, error) {
	row := tx.QueryRow(`SELECT uid, name, clue_cards_json, means_cards_json FROM players WHERE game_id = ? AND uid = ?`, gameID, uid)
	var (
		scannedUID, name              string
		clueCardsJSON, meansCardsJSON string
	)
	if err := row.Scan(&scannedUID, &name, &clueCardsJSON, &meansCardsJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	player := &Player{Name: name, UID: scannedUID}
	if err := json.Unmarshal([]byte(clueCardsJSON), &player.ClueCards); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(meansCardsJSON), &player.MeansCards); err != nil {
		return nil, err
	}
	return player, nil
}

func (a *App) getGuesses(gameID string) ([]Guess, error) {
	rows, err := a.db.Query(`SELECT guessed_by_uid, murderer_uid, means_card_name, clue_card_name, correct, created_timestamp FROM guesses WHERE game_id = ? ORDER BY id ASC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	guesses := []Guess{}
	for rows.Next() {
		var guess Guess
		var correct int
		if err := rows.Scan(&guess.GuessedByUID, &guess.MurdererUID, &guess.MeansCardName, &guess.ClueCardName, &correct, &guess.CreatedAt); err != nil {
			return nil, err
		}
		guess.Correct = correct == 1
		guesses = append(guesses, guess)
	}
	return guesses, rows.Err()
}

func (a *App) getMessages(gameID string) ([]Message, error) {
	rows, err := a.db.Query(`SELECT COALESCE(player_uid, ''), message, type, timestamp FROM messages WHERE game_id = ? ORDER BY id ASC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := []Message{}
	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.PlayerUID, &message.Message, &message.Type, &message.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (a *App) saveGameTx(tx *sql.Tx, game *Game) error {
	otherCardsJSON, err := json.Marshal(game.OtherCards)
	if err != nil {
		return err
	}
	var causeCardJSON any
	if game.CauseCard != nil {
		data, err := json.Marshal(game.CauseCard)
		if err != nil {
			return err
		}
		causeCardJSON = string(data)
	}
	var locationCardJSON any
	if game.LocationCard != nil {
		data, err := json.Marshal(game.LocationCard)
		if err != nil {
			return err
		}
		locationCardJSON = string(data)
	}
	_, err = tx.Exec(`UPDATE games SET started_on = ?, murderer_cards_selected = ?, murderer_uid = ?, murderer_clue_card_name = ?, murderer_means_card_name = ?, cause_card_json = ?, location_card_json = ?, other_cards_json = ?, finished = ? WHERE game_id = ?`,
		nullableString(game.StartedOn),
		boolToInt(game.MurdererCardsSelected),
		nullableString(game.MurdererUID),
		nullableString(game.MurdererClueCardName),
		nullableString(game.MurdererMeansCardName),
		causeCardJSON,
		locationCardJSON,
		string(otherCardsJSON),
		boolToInt(game.Finished),
		game.GameID,
	)
	return err
}

func (a *App) savePlayerCardsTx(tx *sql.Tx, gameID string, player Player) error {
	clueJSON, err := json.Marshal(player.ClueCards)
	if err != nil {
		return err
	}
	meansJSON, err := json.Marshal(player.MeansCards)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE players SET clue_cards_json = ?, means_cards_json = ? WHERE game_id = ? AND uid = ?`, string(clueJSON), string(meansJSON), gameID, player.UID)
	return err
}

func (a *App) sendMessageTx(tx *sql.Tx, gameID, playerUID, message string, messageType MessageType) error {
	_, err := tx.Exec(`INSERT INTO messages(game_id, player_uid, message, type, timestamp) VALUES (?, ?, ?, ?, ?)`, gameID, nullableString(playerUID), message, string(messageType), nowTimestamp(a.timeNow()))
	return err
}

func (a *App) sendForensicMessageTx(tx *sql.Tx, gameID, message string) error {
	return a.sendMessageTx(tx, gameID, "", message, MessageTypeForensic)
}

func findPlayer(players []Player, uid string) *Player {
	for i := range players {
		if players[i].UID == uid {
			return &players[i]
		}
	}
	return nil
}

func hasCard(cards []Card, name string) bool {
	for _, card := range cards {
		if card.Name == name {
			return true
		}
	}
	return false
}

func cloneCards(cards []Card) []Card {
	result := make([]Card, len(cards))
	copy(result, cards)
	return result
}

func cloneForensicCards(cards []ForensicCard) []ForensicCard {
	result := make([]ForensicCard, len(cards))
	copy(result, cards)
	return result
}

func sampleRandom[T any](rng *mathrand.Rand, source []T, count int) []T {
	if count > len(source) {
		panic("sample count exceeds source length")
	}
	idxs := rng.Perm(len(source))[:count]
	sort.Ints(idxs)
	result := make([]T, 0, count)
	for _, idx := range idxs {
		result = append(result, source[idx])
	}
	return result
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
