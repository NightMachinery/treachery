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

const defaultRoomTimerDurationSeconds = 40

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
		`CREATE TABLE IF NOT EXISTS profiles (
			uid TEXT PRIMARY KEY,
			display_name TEXT NOT NULL,
			updated_timestamp TEXT NOT NULL
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
			scientist_uid TEXT,
			marked_scientist_uid TEXT,
			cause_card_json TEXT,
			location_card_json TEXT,
			other_cards_json TEXT NOT NULL DEFAULT '[]',
			finished INTEGER NOT NULL DEFAULT 0,
			means_cards_per_player INTEGER NOT NULL DEFAULT 4,
			clue_cards_per_player INTEGER NOT NULL DEFAULT 4,
			link_clue_count_to_means INTEGER NOT NULL DEFAULT 1,
			means_clues_text_only INTEGER NOT NULL DEFAULT 0,
			accomplice_count INTEGER NOT NULL DEFAULT 0,
			witness_count INTEGER NOT NULL DEFAULT 0,
			witnesses_to_find INTEGER NOT NULL DEFAULT 0,
			pending_witness_selection INTEGER NOT NULL DEFAULT 0,
			winner TEXT NOT NULL DEFAULT 'none',
			finished_reason TEXT,
			result_message TEXT,
			room_timer_duration_seconds INTEGER NOT NULL DEFAULT 0,
			room_timer_expires_at TEXT,
			room_timer_paused_remaining_seconds INTEGER NOT NULL DEFAULT 0,
			room_timer_run_id INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS participants (
			game_id TEXT NOT NULL,
			uid TEXT NOT NULL,
			name TEXT NOT NULL,
			role TEXT NOT NULL,
			joined_timestamp TEXT NOT NULL,
			PRIMARY KEY (game_id, uid),
			FOREIGN KEY (game_id) REFERENCES games(game_id) ON DELETE CASCADE
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
		`CREATE TABLE IF NOT EXISTS room_auth_tokens (
			token TEXT PRIMARY KEY,
			game_id TEXT NOT NULL,
			uid TEXT NOT NULL,
			created_timestamp TEXT NOT NULL,
			UNIQUE(game_id, uid),
			FOREIGN KEY (game_id) REFERENCES games(game_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS player_roles (
			game_id TEXT NOT NULL,
			uid TEXT NOT NULL,
			role TEXT NOT NULL,
			PRIMARY KEY (game_id, uid),
			FOREIGN KEY (game_id) REFERENCES games(game_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS witness_selection_prompts (
			game_id TEXT NOT NULL,
			uid TEXT NOT NULL,
			required_selections INTEGER NOT NULL,
			dismissible INTEGER NOT NULL DEFAULT 0,
			creator_initiated INTEGER NOT NULL DEFAULT 0,
			active INTEGER NOT NULL DEFAULT 1,
			PRIMARY KEY (game_id, uid),
			FOREIGN KEY (game_id) REFERENCES games(game_id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_participants_game_id ON participants(game_id);`,
		`CREATE INDEX IF NOT EXISTS idx_players_game_id ON players(game_id);`,
		`CREATE INDEX IF NOT EXISTS idx_player_roles_game_id ON player_roles(game_id);`,
		`CREATE INDEX IF NOT EXISTS idx_guesses_game_id ON guesses(game_id);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_game_id ON messages(game_id, id);`,
		`CREATE INDEX IF NOT EXISTS idx_room_auth_game_id ON room_auth_tokens(game_id);`,
		`CREATE INDEX IF NOT EXISTS idx_witness_selection_prompts_game_id ON witness_selection_prompts(game_id);`,
	}

	for _, stmt := range stmts {
		if _, err := a.db.Exec(stmt); err != nil {
			return err
		}
	}

	legacyAlterStatements := []string{
		`ALTER TABLE games ADD COLUMN scientist_uid TEXT;`,
		`ALTER TABLE games ADD COLUMN marked_scientist_uid TEXT;`,
		`ALTER TABLE games ADD COLUMN means_cards_per_player INTEGER NOT NULL DEFAULT 4;`,
		`ALTER TABLE games ADD COLUMN clue_cards_per_player INTEGER NOT NULL DEFAULT 4;`,
		`ALTER TABLE games ADD COLUMN link_clue_count_to_means INTEGER NOT NULL DEFAULT 1;`,
		`ALTER TABLE games ADD COLUMN means_clues_text_only INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE games ADD COLUMN accomplice_count INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE games ADD COLUMN witness_count INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE games ADD COLUMN witnesses_to_find INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE games ADD COLUMN pending_witness_selection INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE games ADD COLUMN winner TEXT NOT NULL DEFAULT 'none';`,
		`ALTER TABLE games ADD COLUMN finished_reason TEXT;`,
		`ALTER TABLE games ADD COLUMN result_message TEXT;`,
		`ALTER TABLE games ADD COLUMN room_timer_duration_seconds INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE games ADD COLUMN room_timer_expires_at TEXT;`,
		`ALTER TABLE games ADD COLUMN room_timer_paused_remaining_seconds INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE games ADD COLUMN room_timer_run_id INTEGER NOT NULL DEFAULT 0;`,
	}
	for _, stmt := range legacyAlterStatements {
		_, _ = a.db.Exec(stmt)
	}

	// Backfill any pre-existing player rows into participants for older databases.
	_, _ = a.db.Exec(`INSERT OR IGNORE INTO participants(game_id, uid, name, role, joined_timestamp)
		SELECT p.game_id, p.uid, p.name, 'player', COALESCE(g.created_timestamp, '')
		FROM players p
		JOIN games g ON g.game_id = p.game_id`)
	_, _ = a.db.Exec(`INSERT OR IGNORE INTO player_roles(game_id, uid, role)
		SELECT p.game_id, p.uid,
			CASE
				WHEN g.murderer_uid IS NOT NULL AND g.murderer_uid = p.uid THEN 'murderer'
				ELSE 'investigator'
			END
		FROM players p
		JOIN games g ON g.game_id = p.game_id`)

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

func normalizeParticipantRole(role ParticipantRole) (ParticipantRole, error) {
	switch role {
	case ParticipantRolePlayer, ParticipantRoleObserver:
		return role, nil
	case "":
		return ParticipantRolePlayer, nil
	default:
		return "", fmt.Errorf("%w: invalid participant role", ErrBadInput)
	}
}

type GameSettingsInput struct {
	MeansCardsPerPlayer  int  `json:"meansCardsPerPlayer"`
	ClueCardsPerPlayer   int  `json:"clueCardsPerPlayer"`
	LinkClueCountToMeans bool `json:"linkClueCountToMeans"`
	MeansCluesTextOnly   bool `json:"meansCluesTextOnly"`
	AccompliceCount      int  `json:"accompliceCount"`
	WitnessCount         int  `json:"witnessCount"`
	WitnessesToFind      int  `json:"witnessesToFind"`
}

type GameRoomModsInput struct {
	MeansCluesTextOnly bool `json:"meansCluesTextOnly"`
}

func normalizeSettings(input GameSettingsInput) (GameSettingsInput, error) {
	if input.MeansCardsPerPlayer <= 0 {
		return input, fmt.Errorf("%w: means cards per player must be at least 1", ErrBadInput)
	}
	if input.ClueCardsPerPlayer <= 0 {
		return input, fmt.Errorf("%w: clue cards per player must be at least 1", ErrBadInput)
	}
	if input.AccompliceCount < 0 || input.AccompliceCount > 10 {
		return input, fmt.Errorf("%w: accomplice count must be between 0 and 10", ErrBadInput)
	}
	if input.WitnessCount < 0 || input.WitnessCount > 10 {
		return input, fmt.Errorf("%w: witness count must be between 0 and 10", ErrBadInput)
	}
	if input.LinkClueCountToMeans {
		input.ClueCardsPerPlayer = input.MeansCardsPerPlayer
	}
	if input.WitnessCount == 0 {
		input.WitnessesToFind = 0
	} else if input.WitnessesToFind <= 0 || input.WitnessesToFind > input.WitnessCount {
		return input, fmt.Errorf("%w: witnesses to find must be between 1 and witness count", ErrBadInput)
	}
	return input, nil
}

func (a *App) EnsureSession(existingToken string) (*Session, error) {
	if existingToken != "" {
		session, err := a.GetSession(existingToken)
		if err == nil {
			if profile, profileErr := a.GetProfile(session.UID); profileErr == nil {
				session.DisplayName = profile
			}
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
	if profile, err := a.GetProfile(session.UID); err == nil {
		session.DisplayName = profile
	}
	return &session, nil
}

func (a *App) GetProfile(uid string) (string, error) {
	row := a.db.QueryRow(`SELECT display_name FROM profiles WHERE uid = ?`, uid)
	var displayName string
	if err := row.Scan(&displayName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return displayName, nil
}

func (a *App) UpdateProfile(uid, displayName string) error {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return fmt.Errorf("%w: display name is required", ErrBadInput)
	}
	updatedAt := nowTimestamp(a.timeNow())
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		if _, err := tx.Exec(`INSERT INTO profiles(uid, display_name, updated_timestamp) VALUES (?, ?, ?)
			ON CONFLICT(uid) DO UPDATE SET display_name = excluded.display_name, updated_timestamp = excluded.updated_timestamp`, uid, displayName, updatedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE participants SET name = ? WHERE uid = ?`, displayName, uid); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE players SET name = ? WHERE uid = ?`, displayName, uid); err != nil {
			return err
		}
		return nil
	})
}

func (a *App) CreateGame(creatorUID, gameID string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	if gameID == "" {
		return fmt.Errorf("%w: game ID is required", ErrBadInput)
	}
	displayName, err := a.GetProfile(creatorUID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: display name is required", ErrBadInput)
		}
		return err
	}
	createdAt := nowTimestamp(a.timeNow())
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO games(game_id, creator_uid, created_timestamp, other_cards_json, finished, murderer_cards_selected) VALUES (?, ?, ?, '[]', 0, 0)`, gameID, creatorUID, createdAt)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				return fmt.Errorf("%w: game already exists", ErrBadInput)
			}
			return err
		}
		_, err = tx.Exec(`INSERT INTO participants(game_id, uid, name, role, joined_timestamp) VALUES (?, ?, ?, ?, ?)`, gameID, creatorUID, displayName, string(ParticipantRolePlayer), createdAt)
		return err
	})
}

func (a *App) UpdateGameSettings(gameID, actorUID string, input GameSettingsInput) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	settings, err := normalizeSettings(input)
	if err != nil {
		return err
	}
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != actorUID {
			return fmt.Errorf("%w: only the creator can update game settings", ErrForbidden)
		}
		if game.StartedOn != "" {
			return fmt.Errorf("%w: game already started", ErrBadInput)
		}
		game.MeansCardsPerPlayer = settings.MeansCardsPerPlayer
		game.ClueCardsPerPlayer = settings.ClueCardsPerPlayer
		game.LinkClueCountToMeans = settings.LinkClueCountToMeans
		game.MeansCluesTextOnly = settings.MeansCluesTextOnly
		game.AccompliceCount = settings.AccompliceCount
		game.WitnessCount = settings.WitnessCount
		game.WitnessesToFind = settings.WitnessesToFind
		return a.saveGameTx(tx, game)
	})
}

func (a *App) UpdateRoomMods(gameID, actorUID string, input GameRoomModsInput) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != actorUID {
			return fmt.Errorf("%w: only the creator can update room mods", ErrForbidden)
		}
		if game.StartedOn == "" {
			return fmt.Errorf("%w: room mods are only available after the game starts", ErrBadInput)
		}
		game.MeansCluesTextOnly = input.MeansCluesTextOnly
		return a.saveGameTx(tx, game)
	})
}

func (a *App) UpsertParticipant(gameID, uid string, requestedRole ParticipantRole) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	role, err := normalizeParticipantRole(requestedRole)
	if err != nil {
		return err
	}
	displayName, err := a.GetProfile(uid)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: display name is required", ErrBadInput)
		}
		return err
	}
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		existing, err := a.getParticipantTx(tx, gameID, uid)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
		if game.StartedOn != "" {
			switch {
			case existing == nil && role == ParticipantRoleObserver:
				_, err = tx.Exec(`INSERT INTO participants(game_id, uid, name, role, joined_timestamp) VALUES (?, ?, ?, ?, ?)`, gameID, uid, displayName, string(ParticipantRoleObserver), nowTimestamp(a.timeNow()))
				return err
			case existing == nil:
				return fmt.Errorf("%w: game already started", ErrBadInput)
			case existing.Role == ParticipantRolePlayer:
				_, err = tx.Exec(`UPDATE participants SET name = ? WHERE game_id = ? AND uid = ?`, displayName, gameID, uid)
				return err
			default:
				_, err = tx.Exec(`UPDATE participants SET name = ?, role = ? WHERE game_id = ? AND uid = ?`, displayName, string(ParticipantRoleObserver), gameID, uid)
				return err
			}
		}
		if existing != nil {
			_, err = tx.Exec(`UPDATE participants SET name = ?, role = ? WHERE game_id = ? AND uid = ?`, displayName, string(role), gameID, uid)
			return err
		}
		_, err = tx.Exec(`INSERT INTO participants(game_id, uid, name, role, joined_timestamp) VALUES (?, ?, ?, ?, ?)`, gameID, uid, displayName, string(role), nowTimestamp(a.timeNow()))
		return err
	})
}

func (a *App) SetParticipantRole(gameID, actorUID, targetUID string, requestedRole ParticipantRole) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	role, err := normalizeParticipantRole(requestedRole)
	if err != nil {
		return err
	}
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != actorUID {
			return fmt.Errorf("%w: only the creator can update participant roles", ErrForbidden)
		}
		if game.StartedOn != "" {
			return fmt.Errorf("%w: game already started", ErrBadInput)
		}
		participant, err := a.getParticipantTx(tx, gameID, targetUID)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE participants SET role = ? WHERE game_id = ? AND uid = ?`, string(role), gameID, targetUID); err != nil {
			return err
		}
		if role == ParticipantRoleObserver && game.MarkedScientistUID == targetUID {
			game.MarkedScientistUID = ""
			if err := a.saveGameTx(tx, game); err != nil {
				return err
			}
		}
		if role == ParticipantRoleObserver && participant.UID == game.CreatorUID {
			return nil
		}
		return nil
	})
}

func (a *App) ToggleScientistMark(gameID, actorUID, targetUID string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != actorUID {
			return fmt.Errorf("%w: only the creator can mark the forensic scientist", ErrForbidden)
		}
		if game.StartedOn != "" {
			return fmt.Errorf("%w: game already started", ErrBadInput)
		}
		participant, err := a.getParticipantTx(tx, gameID, targetUID)
		if err != nil {
			return err
		}
		if participant.Role != ParticipantRolePlayer {
			return fmt.Errorf("%w: only players can be marked as scientist", ErrBadInput)
		}
		if game.MarkedScientistUID == targetUID {
			game.MarkedScientistUID = ""
		} else {
			game.MarkedScientistUID = targetUID
		}
		return a.saveGameTx(tx, game)
	})
}

func (a *App) CreateOrGetRoomAuthToken(gameID, uid string) (*RoomAuthToken, error) {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	row := a.db.QueryRow(`SELECT token, game_id, uid, created_timestamp FROM room_auth_tokens WHERE game_id = ? AND uid = ?`, gameID, uid)
	var token RoomAuthToken
	if err := row.Scan(&token.Token, &token.GameID, &token.UID, &token.CreatedTimestamp); err == nil {
		return &token, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	generated, err := randomToken(24)
	if err != nil {
		return nil, err
	}
	createdAt := nowTimestamp(a.timeNow())
	_, err = a.db.Exec(`INSERT INTO room_auth_tokens(token, game_id, uid, created_timestamp) VALUES (?, ?, ?, ?)`, generated, gameID, uid, createdAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return a.CreateOrGetRoomAuthToken(gameID, uid)
		}
		return nil, err
	}
	return &RoomAuthToken{Token: generated, GameID: gameID, UID: uid, CreatedTimestamp: createdAt}, nil
}

func (a *App) ResolveRoomAuthToken(gameID, token string) (string, error) {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	row := a.db.QueryRow(`SELECT uid FROM room_auth_tokens WHERE game_id = ? AND token = ?`, gameID, token)
	var uid string
	if err := row.Scan(&uid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return uid, nil
}

func (a *App) ListGames() ([]Game, error) {
	rows, err := a.db.Query(`SELECT game_id, creator_uid, created_timestamp, started_on, murderer_cards_selected, murderer_uid, murderer_clue_card_name, murderer_means_card_name, scientist_uid, marked_scientist_uid, cause_card_json, location_card_json, other_cards_json, finished, means_cards_per_player, clue_cards_per_player, link_clue_count_to_means, means_clues_text_only, accomplice_count, witness_count, witnesses_to_find, pending_witness_selection, winner, finished_reason, result_message, room_timer_duration_seconds, room_timer_expires_at, room_timer_paused_remaining_seconds, room_timer_run_id FROM games ORDER BY created_timestamp DESC`)
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
	participants, err := a.getParticipants(gameID)
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
	playerRoles, err := a.getPlayerRoles(gameID)
	if err != nil {
		return nil, err
	}
	witnessPrompts, err := a.getWitnessSelectionPrompts(gameID)
	if err != nil {
		return nil, err
	}

	viewer := Viewer{UID: viewerUID}
	if game.CreatorUID == viewerUID {
		viewer.IsCreator = true
	}
	if participant := findParticipant(participants, viewerUID); participant != nil {
		viewer.IsParticipant = true
		viewer.Role = participant.Role
		viewer.Name = participant.Name
		viewer.IsScientist = participant.IsScientist
	}

	snapshot := &GameSnapshot{
		Game:              game,
		Participants:      participants,
		Players:           players,
		Guesses:           guesses,
		Messages:          messages,
		Viewer:            viewer,
		PlayerPrivateData: PlayerPrivateData{},
		ServerTimestamp:   nowTimestamp(a.timeNow()),
	}

	if viewerUID == "" {
		return snapshot, nil
	}

	if role, ok := playerRoles[viewerUID]; ok {
		snapshot.PlayerPrivateData.Role = role
		snapshot.PlayerPrivateData.IsMurderer = role == SecretRoleMurderer
		if role == SecretRoleMurderer {
			snapshot.PlayerPrivateData.ClueCardName = game.MurdererClueCardName
			snapshot.PlayerPrivateData.MeansCardName = game.MurdererMeansCardName
		}
		if role == SecretRoleMurderer || role == SecretRoleAccomplice || role == SecretRoleWitness {
			snapshot.PlayerPrivateData.KnownMurdererTeam = buildKnownMurdererTeam(players, playerRoles)
		}
		if role == SecretRoleAccomplice {
			snapshot.PlayerPrivateData.KnownMurdererClueCardName = game.MurdererClueCardName
			snapshot.PlayerPrivateData.KnownMurdererMeansCardName = game.MurdererMeansCardName
		}
		if prompt, ok := witnessPrompts[viewerUID]; ok && prompt.Active {
			copyPrompt := prompt
			snapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt = &copyPrompt
		}
	}

	if game.ScientistUID == viewerUID {
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

	if viewer.IsCreator && game.PendingWitnessSelection {
		snapshot.ModeratorPrivateData = &ModeratorPrivateData{
			WitnessPromptCandidates: buildWitnessPromptCandidates(players),
		}
	}

	if game.Finished {
		snapshot.RoleReveal = buildRoleReveal(participants, playerRoles, game.ScientistUID)
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
			return fmt.Errorf("%w: only the creator can start the game", ErrForbidden)
		}
		if game.StartedOn != "" {
			return fmt.Errorf("%w: game already started", ErrBadInput)
		}
		participants, err := a.getParticipantsTx(tx, gameID)
		if err != nil {
			return err
		}
		playerParticipants := filterParticipantsByRole(participants, ParticipantRolePlayer)
		if len(playerParticipants) < 4 {
			return fmt.Errorf("%w: need at least 4 players to start", ErrBadInput)
		}
		if len(a.cards.ForensicCards.OtherCards) < 6 {
			return fmt.Errorf("%w: missing forensic cards", ErrBadInput)
		}

		scientist := a.selectScientist(game, playerParticipants)
		suspects := make([]Participant, 0, len(playerParticipants)-1)
		for _, participant := range playerParticipants {
			if participant.UID == scientist.UID {
				continue
			}
			suspects = append(suspects, participant)
		}
		if len(suspects) < 3 {
			return fmt.Errorf("%w: need at least 3 suspects after choosing a scientist", ErrBadInput)
		}
		if err := validateStartSettings(game, len(suspects), len(a.cards.ClueCards), len(a.cards.MeansCards)); err != nil {
			return err
		}

		otherCards := cloneForensicCards(sampleRandom(a.rand, a.cards.ForensicCards.OtherCards, 6))
		clueCards := cloneCards(sampleRandom(a.rand, a.cards.ClueCards, len(suspects)*game.ClueCardsPerPlayer))
		meansCards := cloneCards(sampleRandom(a.rand, a.cards.MeansCards, len(suspects)*game.MeansCardsPerPlayer))
		if err := a.replaceGamePlayersTx(tx, gameID, suspects, clueCards, meansCards, game.ClueCardsPerPlayer, game.MeansCardsPerPlayer); err != nil {
			return err
		}

		roleAssignments := assignSecretRoles(a.rand, suspects, game.AccompliceCount, game.WitnessCount)
		murderer := findParticipantByRole(roleAssignments, SecretRoleMurderer)
		if murderer == nil {
			return fmt.Errorf("missing murderer assignment")
		}
		if err := a.replacePlayerRolesTx(tx, gameID, roleAssignments); err != nil {
			return err
		}
		if err := a.clearWitnessSelectionPromptsTx(tx, gameID); err != nil {
			return err
		}
		startedOn := nowTimestamp(a.timeNow())
		game.StartedOn = startedOn
		game.StartedTimestamp = startedOn
		game.ScientistUID = scientist.UID
		game.MurdererUID = murderer.UID
		game.MurdererSelected = true
		game.MurdererCardsSelected = false
		game.OtherCards = otherCards
		game.PendingWitnessSelection = false
		game.Finished = false
		game.Winner = WinnerNone
		game.FinishedReason = ""
		game.ResultMessage = ""
		game.CauseCard = nil
		game.LocationCard = nil
		game.MurdererClueCardName = ""
		game.MurdererMeansCardName = ""
		clearRoomTimer(game)
		if err := a.saveGameTx(tx, game); err != nil {
			return err
		}
		if err := a.sendForensicMessageTx(tx, gameID, "Game started! The forensic scientist has been chosen. Murderer, select your cards without revealing them."); err != nil {
			return err
		}
		return nil
	})
}

func (a *App) selectScientist(game *Game, playerParticipants []Participant) Participant {
	if game.MarkedScientistUID != "" {
		for _, participant := range playerParticipants {
			if participant.UID == game.MarkedScientistUID {
				return participant
			}
		}
	}
	return playerParticipants[a.rand.Intn(len(playerParticipants))]
}

func (a *App) StartRoomTimer(gameID, actorUID string, seconds *int) error {
	durationSeconds := defaultRoomTimerDurationSeconds
	if seconds != nil {
		durationSeconds = *seconds
	}
	if durationSeconds <= 0 {
		return fmt.Errorf("%w: room timer duration must be a positive whole number of seconds", ErrBadInput)
	}
	return a.updateRoomTimer(gameID, actorUID, func(game *Game) error {
		nextRunID := currentRoomTimerRunID(game) + 1
		game.RoomTimerRunID = nextRunID
		game.RoomTimer = newRunningRoomTimer(a.timeNow(), durationSeconds, nextRunID)
		return nil
	})
}

func (a *App) PauseRoomTimer(gameID, actorUID string) error {
	return a.updateRoomTimer(gameID, actorUID, func(game *Game) error {
		if game.RoomTimer == nil {
			return fmt.Errorf("%w: room timer has not been started", ErrBadInput)
		}
		remainingSeconds, err := roomTimerRemainingSeconds(game.RoomTimer.ExpiresAt, a.timeNow())
		if err != nil {
			return err
		}
		if game.RoomTimer.ExpiresAt == "" || remainingSeconds <= 0 {
			return fmt.Errorf("%w: room timer is not currently running", ErrBadInput)
		}
		game.RoomTimer.PausedRemainingSeconds = remainingSeconds
		game.RoomTimer.ExpiresAt = ""
		return nil
	})
}

func (a *App) ResumeRoomTimer(gameID, actorUID string) error {
	return a.updateRoomTimer(gameID, actorUID, func(game *Game) error {
		if game.RoomTimer == nil {
			return fmt.Errorf("%w: room timer has not been started", ErrBadInput)
		}
		if game.RoomTimer.ExpiresAt != "" || game.RoomTimer.PausedRemainingSeconds <= 0 {
			return fmt.Errorf("%w: room timer is not paused", ErrBadInput)
		}
		game.RoomTimer.ExpiresAt = roomTimerExpiryTimestamp(a.timeNow(), game.RoomTimer.PausedRemainingSeconds)
		game.RoomTimer.PausedRemainingSeconds = 0
		return nil
	})
}

func (a *App) ResetRoomTimer(gameID, actorUID string) error {
	return a.updateRoomTimer(gameID, actorUID, func(game *Game) error {
		if game.RoomTimer == nil || game.RoomTimer.DurationSeconds <= 0 {
			return fmt.Errorf("%w: room timer has not been started", ErrBadInput)
		}
		nextRunID := currentRoomTimerRunID(game) + 1
		durationSeconds := game.RoomTimer.DurationSeconds
		game.RoomTimerRunID = nextRunID
		game.RoomTimer = newRunningRoomTimer(a.timeNow(), durationSeconds, nextRunID)
		return nil
	})
}

func (a *App) ClearRoomTimer(gameID, actorUID string) error {
	return a.updateRoomTimer(gameID, actorUID, func(game *Game) error {
		if game.RoomTimer == nil {
			return fmt.Errorf("%w: room timer has not been started", ErrBadInput)
		}
		clearRoomTimer(game)
		return nil
	})
}

func (a *App) updateRoomTimer(gameID, actorUID string, fn func(*Game) error) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != actorUID {
			return fmt.Errorf("%w: only the creator can control the room timer", ErrForbidden)
		}
		if game.StartedOn == "" {
			return fmt.Errorf("%w: room timer is only available after the game starts", ErrBadInput)
		}
		if game.Finished {
			return fmt.Errorf("%w: room timer is not available after the game ends", ErrBadInput)
		}
		if err := fn(game); err != nil {
			return err
		}
		return a.saveGameTx(tx, game)
	})
}

func (a *App) replaceGamePlayersTx(tx *sql.Tx, gameID string, suspects []Participant, clueCards, meansCards []Card, clueCardsPerPlayer, meansCardsPerPlayer int) error {
	if err := clearPlayersTx(tx, gameID); err != nil {
		return err
	}
	for i, participant := range suspects {
		player := Player{
			UID:        participant.UID,
			Name:       participant.Name,
			ClueCards:  cloneCards(clueCards[i*clueCardsPerPlayer : (i+1)*clueCardsPerPlayer]),
			MeansCards: cloneCards(meansCards[i*meansCardsPerPlayer : (i+1)*meansCardsPerPlayer]),
		}
		clueJSON, err := json.Marshal(player.ClueCards)
		if err != nil {
			return err
		}
		meansJSON, err := json.Marshal(player.MeansCards)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO players(game_id, uid, name, clue_cards_json, means_cards_json) VALUES (?, ?, ?, ?, ?)`, gameID, player.UID, player.Name, string(clueJSON), string(meansJSON)); err != nil {
			return err
		}
	}
	return nil
}

func clearGameRoundDataTx(tx *sql.Tx, gameID string) error {
	if err := clearPlayersTx(tx, gameID); err != nil {
		return err
	}
	if err := clearPlayerRolesTx(tx, gameID); err != nil {
		return err
	}
	if err := clearGuessesTx(tx, gameID); err != nil {
		return err
	}
	if err := clearMessagesTx(tx, gameID); err != nil {
		return err
	}
	return clearWitnessSelectionPromptsTx(tx, gameID)
}

func clearPlayersTx(tx *sql.Tx, gameID string) error {
	_, err := tx.Exec(`DELETE FROM players WHERE game_id = ?`, gameID)
	return err
}

func clearPlayerRolesTx(tx *sql.Tx, gameID string) error {
	_, err := tx.Exec(`DELETE FROM player_roles WHERE game_id = ?`, gameID)
	return err
}

func clearGuessesTx(tx *sql.Tx, gameID string) error {
	_, err := tx.Exec(`DELETE FROM guesses WHERE game_id = ?`, gameID)
	return err
}

func clearMessagesTx(tx *sql.Tx, gameID string) error {
	_, err := tx.Exec(`DELETE FROM messages WHERE game_id = ?`, gameID)
	return err
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
		if game.Finished || game.PendingWitnessSelection {
			return fmt.Errorf("%w: game is no longer accepting murderer card changes", ErrBadInput)
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
		if game.ScientistUID != uid {
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
		if game.ScientistUID == uid {
			return fmt.Errorf("%w: forensic scientist cannot guess", ErrForbidden)
		}
		if game.Finished {
			return fmt.Errorf("%w: game already finished", ErrBadInput)
		}
		if game.PendingWitnessSelection {
			return fmt.Errorf("%w: witness selection is already in progress", ErrBadInput)
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
		var duplicateGuessCount int
		if err := tx.QueryRow(
			`SELECT COUNT(1) FROM guesses WHERE game_id = ? AND murderer_uid = ? AND means_card_name = ? AND clue_card_name = ?`,
			gameID,
			murdererUID,
			meansCardName,
			clueCardName,
		).Scan(&duplicateGuessCount); err != nil {
			return err
		}
		if duplicateGuessCount > 0 {
			return fmt.Errorf("%w: that exact guess was already submitted", ErrBadInput)
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
			return a.handleCorrectGuessTx(tx, game)
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
		participant, err := a.getParticipantTx(tx, gameID, uid)
		if err != nil {
			return fmt.Errorf("%w: only game participants can send messages", ErrForbidden)
		}
		if participant.Role != ParticipantRolePlayer {
			return fmt.Errorf("%w: observers cannot send messages", ErrForbidden)
		}
		if game.StartedOn != "" && game.ScientistUID == uid {
			return fmt.Errorf("%w: forensic scientist cannot send chat after the game starts", ErrForbidden)
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
			return fmt.Errorf("%w: only the creator can end the game", ErrForbidden)
		}
		finishGame(game, WinnerNone, "moderator-ended", "The moderator ended the game.")
		if err := a.clearWitnessSelectionPromptsTx(tx, gameID); err != nil {
			return err
		}
		return a.saveGameTx(tx, game)
	})
}

func (a *App) RestartGame(gameID, uid string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != uid {
			return fmt.Errorf("%w: only the creator can restart the game", ErrForbidden)
		}
		if !game.Finished {
			return fmt.Errorf("%w: game must be finished before it can be restarted", ErrBadInput)
		}
		if err := clearGameRoundDataTx(tx, gameID); err != nil {
			return err
		}

		game.StartedOn = ""
		game.StartedTimestamp = ""
		game.ScientistUID = ""
		game.MurdererUID = ""
		game.MurdererSelected = false
		game.MurdererCardsSelected = false
		game.MurdererClueCardName = ""
		game.MurdererMeansCardName = ""
		game.CauseCard = nil
		game.LocationCard = nil
		game.OtherCards = []ForensicCard{}
		game.PendingWitnessSelection = false
		game.Finished = false
		game.Winner = WinnerNone
		game.FinishedReason = ""
		game.ResultMessage = ""
		clearRoomTimer(game)

		return a.saveGameTx(tx, game)
	})
}

func (a *App) ShowWitnessSelectionPrompt(gameID, actorUID, targetUID string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.CreatorUID != actorUID {
			return fmt.Errorf("%w: only the creator can show witness prompts", ErrForbidden)
		}
		if !game.PendingWitnessSelection || game.Finished {
			return fmt.Errorf("%w: witness selection is not active", ErrBadInput)
		}
		players, err := a.getPlayersTx(tx, gameID)
		if err != nil {
			return err
		}
		if findPlayer(players, targetUID) == nil {
			return fmt.Errorf("%w: target must be a current suspect", ErrBadInput)
		}
		playerRoles, err := a.getPlayerRolesTx(tx, gameID)
		if err != nil {
			return err
		}
		role, ok := playerRoles[targetUID]
		if !ok {
			return fmt.Errorf("%w: target must be a current suspect", ErrBadInput)
		}
		prompts, err := a.getWitnessSelectionPromptsTx(tx, gameID)
		if err != nil {
			return err
		}
		if role != SecretRoleAccomplice {
			return nil
		}
		if existing, ok := prompts[targetUID]; ok && existing.Active && !existing.Dismissible {
			return nil
		}
		return a.upsertWitnessSelectionPromptTx(tx, gameID, targetUID, WitnessSelectionPromptState{
			RequiredSelections: game.WitnessesToFind,
			Dismissible:        true,
			CreatorInitiated:   true,
		})
	})
}

func (a *App) SubmitWitnessSelection(gameID, uid string, selectedUIDs []string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if !game.PendingWitnessSelection || game.Finished {
			return fmt.Errorf("%w: witness selection is not active", ErrBadInput)
		}
		playerRoles, err := a.getPlayerRolesTx(tx, gameID)
		if err != nil {
			return err
		}
		role, ok := playerRoles[uid]
		if !ok || (role != SecretRoleMurderer && role != SecretRoleAccomplice) {
			return fmt.Errorf("%w: only the murderer team can submit witness selections", ErrForbidden)
		}
		prompts, err := a.getWitnessSelectionPromptsTx(tx, gameID)
		if err != nil {
			return err
		}
		prompt, ok := prompts[uid]
		if !ok || !prompt.Active {
			return fmt.Errorf("%w: no active witness-selection prompt", ErrForbidden)
		}
		selectedUIDs, err = normalizeSelectedWitnessUIDs(selectedUIDs, game.WitnessesToFind)
		if err != nil {
			return err
		}
		for _, selectedUID := range selectedUIDs {
			selectedRole, ok := playerRoles[selectedUID]
			if !ok {
				return fmt.Errorf("%w: selected player is not a suspect", ErrBadInput)
			}
			if selectedRole == SecretRoleMurderer || selectedRole == SecretRoleAccomplice {
				return fmt.Errorf("%w: murderer-team players cannot be selected as witnesses", ErrBadInput)
			}
		}
		allWitnesses := true
		for _, selectedUID := range selectedUIDs {
			if playerRoles[selectedUID] != SecretRoleWitness {
				allWitnesses = false
				break
			}
		}
		if err := a.clearWitnessSelectionPromptsTx(tx, gameID); err != nil {
			return err
		}
		if allWitnesses {
			finishGame(game, WinnerMurdererTeam, "witnesses-found", "The murderer team correctly identified every witness and escaped.")
		} else {
			finishGame(game, WinnerInvestigatorTeam, "witnesses-missed", "The murderer team failed to identify the witnesses. The investigator team wins.")
		}
		return a.saveGameTx(tx, game)
	})
}

func (a *App) DismissWitnessSelectionPrompt(gameID, uid string) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if !game.PendingWitnessSelection || game.Finished {
			return fmt.Errorf("%w: witness selection is not active", ErrBadInput)
		}
		prompts, err := a.getWitnessSelectionPromptsTx(tx, gameID)
		if err != nil {
			return err
		}
		prompt, ok := prompts[uid]
		if !ok || !prompt.Active {
			return fmt.Errorf("%w: no active witness-selection prompt", ErrForbidden)
		}
		if !prompt.Dismissible {
			return fmt.Errorf("%w: this prompt cannot be dismissed", ErrForbidden)
		}
		return a.deleteWitnessSelectionPromptTx(tx, gameID, uid)
	})
}

func (a *App) updateForensicCard(gameID, uid string, mutator func(*Game) error) error {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	return a.withTx(context.Background(), func(tx *sql.Tx) error {
		game, err := a.getGameTx(tx, gameID)
		if err != nil {
			return err
		}
		if game.ScientistUID != uid {
			return fmt.Errorf("%w: only the forensic scientist can update clue cards", ErrForbidden)
		}
		if game.Finished || game.PendingWitnessSelection {
			return fmt.Errorf("%w: game is already resolving", ErrBadInput)
		}
		if err := mutator(game); err != nil {
			return err
		}
		return a.saveGameTx(tx, game)
	})
}

func (a *App) handleCorrectGuessTx(tx *sql.Tx, game *Game) error {
	if game.WitnessCount == 0 {
		finishGame(game, WinnerInvestigatorTeam, "correct-guess", "The investigators correctly identified the murderer and solved the case.")
		return a.saveGameTx(tx, game)
	}
	game.PendingWitnessSelection = true
	game.Winner = WinnerNone
	game.FinishedReason = ""
	game.ResultMessage = ""
	if err := a.clearWitnessSelectionPromptsTx(tx, game.GameID); err != nil {
		return err
	}
	if err := a.upsertWitnessSelectionPromptTx(tx, game.GameID, game.MurdererUID, WitnessSelectionPromptState{
		RequiredSelections: game.WitnessesToFind,
		Dismissible:        false,
		CreatorInitiated:   false,
	}); err != nil {
		return err
	}
	if err := a.sendForensicMessageTx(tx, game.GameID, "The murderer team now has one chance to identify the witnesses."); err != nil {
		return err
	}
	return a.saveGameTx(tx, game)
}

func finishGame(game *Game, winner Winner, reason, message string) {
	game.Finished = true
	game.PendingWitnessSelection = false
	game.Winner = winner
	game.FinishedReason = reason
	game.ResultMessage = message
	clearRoomTimer(game)
}

func validateStartSettings(game *Game, suspectCount, totalClueCards, totalMeansCards int) error {
	if game.MeansCardsPerPlayer <= 0 || game.ClueCardsPerPlayer <= 0 {
		return fmt.Errorf("%w: cards per player must be at least 1", ErrBadInput)
	}
	if game.WitnessCount < 0 || game.AccompliceCount < 0 {
		return fmt.Errorf("%w: role counts cannot be negative", ErrBadInput)
	}
	if 1+game.AccompliceCount+game.WitnessCount > suspectCount {
		return fmt.Errorf("%w: too many special roles for the available suspects", ErrBadInput)
	}
	if game.WitnessCount == 0 {
		if game.WitnessesToFind != 0 {
			return fmt.Errorf("%w: witnesses to find must be 0 when no witnesses are enabled", ErrBadInput)
		}
	} else if game.WitnessesToFind <= 0 || game.WitnessesToFind > game.WitnessCount {
		return fmt.Errorf("%w: witnesses to find must be between 1 and witness count", ErrBadInput)
	}
	if suspectCount*game.ClueCardsPerPlayer > totalClueCards {
		return fmt.Errorf("%w: not enough clue cards for the configured lobby settings", ErrBadInput)
	}
	if suspectCount*game.MeansCardsPerPlayer > totalMeansCards {
		return fmt.Errorf("%w: not enough means cards for the configured lobby settings", ErrBadInput)
	}
	return nil
}

func assignSecretRoles(rng *mathrand.Rand, suspects []Participant, accompliceCount, witnessCount int) map[string]SecretRole {
	roles := make(map[string]SecretRole, len(suspects))
	if len(suspects) == 0 {
		return roles
	}
	order := rng.Perm(len(suspects))
	roles[suspects[order[0]].UID] = SecretRoleMurderer
	idx := 1
	for i := 0; i < accompliceCount && idx < len(order); i++ {
		roles[suspects[order[idx]].UID] = SecretRoleAccomplice
		idx++
	}
	for i := 0; i < witnessCount && idx < len(order); i++ {
		roles[suspects[order[idx]].UID] = SecretRoleWitness
		idx++
	}
	for _, participant := range suspects {
		if _, ok := roles[participant.UID]; !ok {
			roles[participant.UID] = SecretRoleInvestigator
		}
	}
	return roles
}

func findParticipantByRole(assignments map[string]SecretRole, role SecretRole) *Participant {
	for uid, assignedRole := range assignments {
		if assignedRole == role {
			return &Participant{UID: uid}
		}
	}
	return nil
}

func normalizeSelectedWitnessUIDs(selectedUIDs []string, required int) ([]string, error) {
	seen := make(map[string]struct{}, len(selectedUIDs))
	normalized := make([]string, 0, len(selectedUIDs))
	for _, uid := range selectedUIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		if _, ok := seen[uid]; ok {
			return nil, fmt.Errorf("%w: witness selections must be unique", ErrBadInput)
		}
		seen[uid] = struct{}{}
		normalized = append(normalized, uid)
	}
	if len(normalized) != required {
		return nil, fmt.Errorf("%w: select exactly %d witness(es)", ErrBadInput, required)
	}
	return normalized, nil
}

func (a *App) checkAndEndGameTx(tx *sql.Tx, game *Game) error {
	players, err := a.getPlayersTx(tx, game.GameID)
	if err != nil {
		return err
	}
	playerRoles, err := a.getPlayerRolesTx(tx, game.GameID)
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
		role, ok := playerRoles[player.UID]
		if !ok {
			continue
		}
		if role == SecretRoleMurderer || role == SecretRoleAccomplice {
			continue
		}
		if _, ok := guessedBy[player.UID]; !ok {
			return nil
		}
	}
	finishGame(game, WinnerMurdererTeam, "all-guesses-used", "No correct good-team guess was made before every eligible good-team player ran out of guesses. The murderer team wins.")
	if err := a.sendForensicMessageTx(tx, game.GameID, "All good-team guesses are exhausted. The murderer team wins."); err != nil {
		return err
	}
	return a.saveGameTx(tx, game)
}

func (a *App) getGame(gameID string) (*Game, error) {
	row := a.db.QueryRow(`SELECT game_id, creator_uid, created_timestamp, started_on, murderer_cards_selected, murderer_uid, murderer_clue_card_name, murderer_means_card_name, scientist_uid, marked_scientist_uid, cause_card_json, location_card_json, other_cards_json, finished, means_cards_per_player, clue_cards_per_player, link_clue_count_to_means, means_clues_text_only, accomplice_count, witness_count, witnesses_to_find, pending_witness_selection, winner, finished_reason, result_message, room_timer_duration_seconds, room_timer_expires_at, room_timer_paused_remaining_seconds, room_timer_run_id FROM games WHERE game_id = ?`, gameID)
	game, err := scanGame(row)
	if err != nil {
		return nil, err
	}
	return game, nil
}

func (a *App) getGameTx(tx *sql.Tx, gameID string) (*Game, error) {
	row := tx.QueryRow(`SELECT game_id, creator_uid, created_timestamp, started_on, murderer_cards_selected, murderer_uid, murderer_clue_card_name, murderer_means_card_name, scientist_uid, marked_scientist_uid, cause_card_json, location_card_json, other_cards_json, finished, means_cards_per_player, clue_cards_per_player, link_clue_count_to_means, means_clues_text_only, accomplice_count, witness_count, witnesses_to_find, pending_witness_selection, winner, finished_reason, result_message, room_timer_duration_seconds, room_timer_expires_at, room_timer_paused_remaining_seconds, room_timer_run_id FROM games WHERE game_id = ?`, gameID)
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
		scientistUID, markedScientistUID                                    sql.NullString
		causeCardJSON, locationCardJSON, roomTimerExpiresAt                 sql.NullString
		finishedReason, resultMessage                                       sql.NullString
		otherCardsJSON, winner                                              string
		murdererCardsSelected, finished                                     int
		meansCardsPerPlayer, clueCardsPerPlayer                             int
		linkClueCountToMeans, meansCluesTextOnly                            int
		accompliceCount, witnessCount, witnessesToFind                      int
		pendingWitnessSelection                                             int
		roomTimerDurationSeconds, roomTimerPausedRemainingSeconds           int
		roomTimerRunID                                                      int
	)
	if err := scanner.Scan(
		&gameID,
		&creatorUID,
		&createdTimestamp,
		&startedOn,
		&murdererCardsSelected,
		&murdererUID,
		&murdererClueCardName,
		&murdererMeansCardName,
		&scientistUID,
		&markedScientistUID,
		&causeCardJSON,
		&locationCardJSON,
		&otherCardsJSON,
		&finished,
		&meansCardsPerPlayer,
		&clueCardsPerPlayer,
		&linkClueCountToMeans,
		&meansCluesTextOnly,
		&accompliceCount,
		&witnessCount,
		&witnessesToFind,
		&pendingWitnessSelection,
		&winner,
		&finishedReason,
		&resultMessage,
		&roomTimerDurationSeconds,
		&roomTimerExpiresAt,
		&roomTimerPausedRemainingSeconds,
		&roomTimerRunID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	game := &Game{
		GameID:                  gameID,
		CreatorUID:              creatorUID,
		CreatedTimestamp:        createdTimestamp,
		StartedTimestamp:        startedOn.String,
		StartedOn:               startedOn.String,
		MurdererCardsSelected:   murdererCardsSelected == 1,
		MurdererSelected:        murdererUID.Valid,
		MurdererUID:             murdererUID.String,
		MurdererClueCardName:    murdererClueCardName.String,
		MurdererMeansCardName:   murdererMeansCardName.String,
		ScientistUID:            scientistUID.String,
		MarkedScientistUID:      markedScientistUID.String,
		Finished:                finished == 1,
		MeansCardsPerPlayer:     meansCardsPerPlayer,
		ClueCardsPerPlayer:      clueCardsPerPlayer,
		LinkClueCountToMeans:    linkClueCountToMeans == 1,
		MeansCluesTextOnly:      meansCluesTextOnly == 1,
		AccompliceCount:         accompliceCount,
		WitnessCount:            witnessCount,
		WitnessesToFind:         witnessesToFind,
		PendingWitnessSelection: pendingWitnessSelection == 1,
		Winner:                  Winner(winner),
		FinishedReason:          finishedReason.String,
		ResultMessage:           resultMessage.String,
		RoomTimerRunID:          roomTimerRunID,
	}
	if game.Winner == "" {
		game.Winner = WinnerNone
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
	if roomTimerDurationSeconds > 0 {
		game.RoomTimer = &RoomTimer{
			DurationSeconds:        roomTimerDurationSeconds,
			ExpiresAt:              roomTimerExpiresAt.String,
			PausedRemainingSeconds: roomTimerPausedRemainingSeconds,
			RunID:                  roomTimerRunID,
		}
	}
	return game, nil
}

func (a *App) getParticipants(gameID string) ([]Participant, error) {
	rows, err := a.db.Query(`SELECT p.uid, p.name, p.role, g.creator_uid, g.scientist_uid, g.marked_scientist_uid
		FROM participants p
		JOIN games g ON g.game_id = p.game_id
		WHERE p.game_id = ?
		ORDER BY p.joined_timestamp ASC, p.rowid ASC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanParticipants(rows)
}

func (a *App) getParticipantsTx(tx *sql.Tx, gameID string) ([]Participant, error) {
	rows, err := tx.Query(`SELECT p.uid, p.name, p.role, g.creator_uid, g.scientist_uid, g.marked_scientist_uid
		FROM participants p
		JOIN games g ON g.game_id = p.game_id
		WHERE p.game_id = ?
		ORDER BY p.joined_timestamp ASC, p.rowid ASC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanParticipants(rows)
}

func scanParticipants(rows *sql.Rows) ([]Participant, error) {
	participants := []Participant{}
	for rows.Next() {
		var uid, name, role, creatorUID string
		var scientistUID, markedScientistUID sql.NullString
		if err := rows.Scan(&uid, &name, &role, &creatorUID, &scientistUID, &markedScientistUID); err != nil {
			return nil, err
		}
		participants = append(participants, Participant{
			UID:               uid,
			Name:              name,
			Role:              ParticipantRole(role),
			IsCreator:         uid == creatorUID,
			IsScientist:       scientistUID.Valid && scientistUID.String == uid,
			IsMarkedScientist: markedScientistUID.Valid && markedScientistUID.String == uid,
		})
	}
	return participants, rows.Err()
}

func (a *App) getParticipantTx(tx *sql.Tx, gameID, uid string) (*Participant, error) {
	row := tx.QueryRow(`SELECT p.uid, p.name, p.role, g.creator_uid, g.scientist_uid, g.marked_scientist_uid
		FROM participants p
		JOIN games g ON g.game_id = p.game_id
		WHERE p.game_id = ? AND p.uid = ?`, gameID, uid)

	var participant Participant
	var role, creatorUID string
	var scientistUID, markedScientistUID sql.NullString
	if err := row.Scan(&participant.UID, &participant.Name, &role, &creatorUID, &scientistUID, &markedScientistUID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	participant.Role = ParticipantRole(role)
	participant.IsCreator = participant.UID == creatorUID
	participant.IsScientist = scientistUID.Valid && scientistUID.String == participant.UID
	participant.IsMarkedScientist = markedScientistUID.Valid && markedScientistUID.String == participant.UID
	return &participant, nil
}

func filterParticipantsByRole(participants []Participant, role ParticipantRole) []Participant {
	result := make([]Participant, 0, len(participants))
	for _, participant := range participants {
		if participant.Role == role {
			result = append(result, participant)
		}
	}
	return result
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

func (a *App) getPlayerRoles(gameID string) (map[string]SecretRole, error) {
	rows, err := a.db.Query(`SELECT uid, role FROM player_roles WHERE game_id = ?`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlayerRoles(rows)
}

func (a *App) getPlayerRolesTx(tx *sql.Tx, gameID string) (map[string]SecretRole, error) {
	rows, err := tx.Query(`SELECT uid, role FROM player_roles WHERE game_id = ?`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlayerRoles(rows)
}

func scanPlayerRoles(rows *sql.Rows) (map[string]SecretRole, error) {
	result := map[string]SecretRole{}
	for rows.Next() {
		var uid, role string
		if err := rows.Scan(&uid, &role); err != nil {
			return nil, err
		}
		result[uid] = SecretRole(role)
	}
	return result, rows.Err()
}

func (a *App) replacePlayerRolesTx(tx *sql.Tx, gameID string, roles map[string]SecretRole) error {
	if err := clearPlayerRolesTx(tx, gameID); err != nil {
		return err
	}
	for uid, role := range roles {
		if _, err := tx.Exec(`INSERT INTO player_roles(game_id, uid, role) VALUES (?, ?, ?)`, gameID, uid, string(role)); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) getWitnessSelectionPrompts(gameID string) (map[string]WitnessSelectionPromptState, error) {
	rows, err := a.db.Query(`SELECT uid, required_selections, dismissible, creator_initiated, active FROM witness_selection_prompts WHERE game_id = ?`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWitnessSelectionPrompts(rows)
}

func (a *App) getWitnessSelectionPromptsTx(tx *sql.Tx, gameID string) (map[string]WitnessSelectionPromptState, error) {
	rows, err := tx.Query(`SELECT uid, required_selections, dismissible, creator_initiated, active FROM witness_selection_prompts WHERE game_id = ?`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWitnessSelectionPrompts(rows)
}

func scanWitnessSelectionPrompts(rows *sql.Rows) (map[string]WitnessSelectionPromptState, error) {
	result := map[string]WitnessSelectionPromptState{}
	for rows.Next() {
		var (
			uid                string
			requiredSelections int
			dismissible        int
			creatorInitiated   int
			active             int
		)
		if err := rows.Scan(&uid, &requiredSelections, &dismissible, &creatorInitiated, &active); err != nil {
			return nil, err
		}
		result[uid] = WitnessSelectionPromptState{
			RequiredSelections: requiredSelections,
			Dismissible:        dismissible == 1,
			CreatorInitiated:   creatorInitiated == 1,
			Active:             active == 1,
		}
		if active == 0 {
			// Keep inactive rows out of the map.
			delete(result, uid)
		}
	}
	return result, rows.Err()
}

func (a *App) upsertWitnessSelectionPromptTx(tx *sql.Tx, gameID, uid string, prompt WitnessSelectionPromptState) error {
	_, err := tx.Exec(`INSERT INTO witness_selection_prompts(game_id, uid, required_selections, dismissible, creator_initiated, active)
		VALUES (?, ?, ?, ?, ?, 1)
		ON CONFLICT(game_id, uid) DO UPDATE SET
			required_selections = excluded.required_selections,
			dismissible = excluded.dismissible,
			creator_initiated = excluded.creator_initiated,
			active = 1`, gameID, uid, prompt.RequiredSelections, boolToInt(prompt.Dismissible), boolToInt(prompt.CreatorInitiated))
	return err
}

func clearWitnessSelectionPromptsTx(tx *sql.Tx, gameID string) error {
	_, err := tx.Exec(`DELETE FROM witness_selection_prompts WHERE game_id = ?`, gameID)
	return err
}

func (a *App) clearWitnessSelectionPromptsTx(tx *sql.Tx, gameID string) error {
	return clearWitnessSelectionPromptsTx(tx, gameID)
}

func (a *App) deleteWitnessSelectionPromptTx(tx *sql.Tx, gameID, uid string) error {
	_, err := tx.Exec(`DELETE FROM witness_selection_prompts WHERE game_id = ? AND uid = ?`, gameID, uid)
	return err
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
	roomTimerRunID := currentRoomTimerRunID(game)
	game.RoomTimerRunID = roomTimerRunID
	roomTimerDurationSeconds := 0
	roomTimerPausedRemainingSeconds := 0
	var roomTimerExpiresAt any
	if game.RoomTimer != nil {
		roomTimerDurationSeconds = game.RoomTimer.DurationSeconds
		roomTimerPausedRemainingSeconds = game.RoomTimer.PausedRemainingSeconds
		roomTimerRunID = game.RoomTimer.RunID
		game.RoomTimerRunID = roomTimerRunID
		roomTimerExpiresAt = nullableString(game.RoomTimer.ExpiresAt)
	}
	_, err = tx.Exec(`UPDATE games SET started_on = ?, murderer_cards_selected = ?, murderer_uid = ?, murderer_clue_card_name = ?, murderer_means_card_name = ?, scientist_uid = ?, marked_scientist_uid = ?, cause_card_json = ?, location_card_json = ?, other_cards_json = ?, finished = ?, means_cards_per_player = ?, clue_cards_per_player = ?, link_clue_count_to_means = ?, means_clues_text_only = ?, accomplice_count = ?, witness_count = ?, witnesses_to_find = ?, pending_witness_selection = ?, winner = ?, finished_reason = ?, result_message = ?, room_timer_duration_seconds = ?, room_timer_expires_at = ?, room_timer_paused_remaining_seconds = ?, room_timer_run_id = ? WHERE game_id = ?`,
		nullableString(game.StartedOn),
		boolToInt(game.MurdererCardsSelected),
		nullableString(game.MurdererUID),
		nullableString(game.MurdererClueCardName),
		nullableString(game.MurdererMeansCardName),
		nullableString(game.ScientistUID),
		nullableString(game.MarkedScientistUID),
		causeCardJSON,
		locationCardJSON,
		string(otherCardsJSON),
		boolToInt(game.Finished),
		game.MeansCardsPerPlayer,
		game.ClueCardsPerPlayer,
		boolToInt(game.LinkClueCountToMeans),
		boolToInt(game.MeansCluesTextOnly),
		game.AccompliceCount,
		game.WitnessCount,
		game.WitnessesToFind,
		boolToInt(game.PendingWitnessSelection),
		string(game.Winner),
		nullableString(game.FinishedReason),
		nullableString(game.ResultMessage),
		roomTimerDurationSeconds,
		roomTimerExpiresAt,
		roomTimerPausedRemainingSeconds,
		roomTimerRunID,
		game.GameID,
	)
	return err
}

func (a *App) sendMessageTx(tx *sql.Tx, gameID, playerUID, message string, messageType MessageType) error {
	_, err := tx.Exec(`INSERT INTO messages(game_id, player_uid, message, type, timestamp) VALUES (?, ?, ?, ?, ?)`, gameID, nullableString(playerUID), message, string(messageType), nowTimestamp(a.timeNow()))
	return err
}

func (a *App) sendForensicMessageTx(tx *sql.Tx, gameID, message string) error {
	return a.sendMessageTx(tx, gameID, "", message, MessageTypeForensic)
}

func findParticipant(participants []Participant, uid string) *Participant {
	for i := range participants {
		if participants[i].UID == uid {
			return &participants[i]
		}
	}
	return nil
}

func findPlayer(players []Player, uid string) *Player {
	for i := range players {
		if players[i].UID == uid {
			return &players[i]
		}
	}
	return nil
}

func buildKnownMurdererTeam(players []Player, playerRoles map[string]SecretRole) []KnownRolePlayer {
	known := []KnownRolePlayer{}
	for _, player := range players {
		role := playerRoles[player.UID]
		if role != SecretRoleMurderer && role != SecretRoleAccomplice {
			continue
		}
		known = append(known, KnownRolePlayer{
			UID:  player.UID,
			Name: player.Name,
			Role: role,
		})
	}
	sort.Slice(known, func(i, j int) bool {
		if known[i].Role != known[j].Role {
			return secretRoleSortOrder(known[i].Role) < secretRoleSortOrder(known[j].Role)
		}
		return known[i].Name < known[j].Name
	})
	return known
}

func buildWitnessPromptCandidates(players []Player) []WitnessPromptCandidate {
	candidates := make([]WitnessPromptCandidate, 0, len(players))
	for _, player := range players {
		candidates = append(candidates, WitnessPromptCandidate{
			UID:  player.UID,
			Name: player.Name,
		})
	}
	return candidates
}

func buildRoleReveal(participants []Participant, playerRoles map[string]SecretRole, scientistUID string) []RoleRevealEntry {
	result := make([]RoleRevealEntry, 0, len(participants))
	for _, participant := range participants {
		revealedRole := string(participant.Role)
		switch {
		case participant.UID == scientistUID:
			revealedRole = "scientist"
		case participant.Role == ParticipantRoleObserver:
			revealedRole = "observer"
		case playerRoles[participant.UID] != "":
			revealedRole = string(playerRoles[participant.UID])
		default:
			revealedRole = "player"
		}
		result = append(result, RoleRevealEntry{
			UID:  participant.UID,
			Name: participant.Name,
			Role: revealedRole,
		})
	}
	return result
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

func newRunningRoomTimer(now time.Time, durationSeconds, runID int) *RoomTimer {
	return &RoomTimer{
		DurationSeconds: durationSeconds,
		ExpiresAt:       roomTimerExpiryTimestamp(now, durationSeconds),
		RunID:           runID,
	}
}

func roomTimerExpiryTimestamp(now time.Time, durationSeconds int) string {
	return nowTimestamp(now.Add(time.Duration(durationSeconds) * time.Second))
}

func roomTimerRemainingSeconds(expiresAt string, now time.Time) (int, error) {
	if strings.TrimSpace(expiresAt) == "" {
		return 0, nil
	}
	expiresAtTime, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return 0, err
	}
	remainingDuration := expiresAtTime.Sub(now)
	if remainingDuration <= 0 {
		return 0, nil
	}
	return int((remainingDuration + time.Second - 1) / time.Second), nil
}

func currentRoomTimerRunID(game *Game) int {
	if game == nil {
		return 0
	}
	if game.RoomTimer != nil && game.RoomTimer.RunID > game.RoomTimerRunID {
		return game.RoomTimer.RunID
	}
	return game.RoomTimerRunID
}

func clearRoomTimer(game *Game) {
	if game == nil {
		return
	}
	game.RoomTimerRunID = currentRoomTimerRunID(game)
	game.RoomTimer = nil
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

func secretRoleSortOrder(role SecretRole) int {
	switch role {
	case SecretRoleMurderer:
		return 0
	case SecretRoleAccomplice:
		return 1
	case SecretRoleWitness:
		return 2
	default:
		return 3
	}
}
