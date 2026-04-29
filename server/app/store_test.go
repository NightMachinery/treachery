package app

import (
	"context"
	"database/sql"
	"errors"
	mathrand "math/rand"
	"path/filepath"
	"testing"
	"time"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	app, err := New(Config{
		DataDir:      t.TempDir(),
		WordpacksDir: filepath.Join("..", "..", "wordpacks"),
	})
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	app.rand = mathrand.New(mathrand.NewSource(1))
	return app
}

func setProfile(t *testing.T, app *App, uid, name string) {
	t.Helper()
	if err := app.UpdateProfile(uid, name); err != nil {
		t.Fatalf("update profile %s: %v", uid, err)
	}
}

func joinPlayer(t *testing.T, app *App, gameID, uid, name string) {
	t.Helper()
	setProfile(t, app, uid, name)
	if err := app.UpsertParticipant(gameID, uid, ParticipantRolePlayer); err != nil {
		t.Fatalf("join player %s: %v", uid, err)
	}
}

func joinPlayers(t *testing.T, app *App, gameID string, uids ...string) {
	t.Helper()
	for _, uid := range uids {
		joinPlayer(t, app, gameID, uid, uid)
	}
}

func startTestGame(t *testing.T, app *App, gameID, creator string, otherUIDs ...string) {
	t.Helper()
	setProfile(t, app, creator, "Creator")
	if err := app.CreateGame(creator, gameID); err != nil {
		t.Fatalf("create game: %v", err)
	}
	joinPlayers(t, app, gameID, otherUIDs...)
	if err := app.StartGame(gameID, creator); err != nil {
		t.Fatalf("start game: %v", err)
	}
}

func TestCreateGameAutoJoinsCreator(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	setProfile(t, app, "creator", "Creator")
	if err := app.CreateGame("creator", "ABCD"); err != nil {
		t.Fatalf("create game: %v", err)
	}

	snapshot, err := app.GetSnapshot("ABCD", "creator")
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.Participants) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(snapshot.Participants))
	}
	participant := snapshot.Participants[0]
	if participant.UID != "creator" || participant.Role != ParticipantRolePlayer || !participant.IsCreator {
		t.Fatalf("unexpected creator participant: %+v", participant)
	}
	if !snapshot.Viewer.IsParticipant || !snapshot.Viewer.IsCreator {
		t.Fatalf("unexpected viewer info: %+v", snapshot.Viewer)
	}
}

func TestCreateGameUsesPackDefaults(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	setProfile(t, app, "creator", "Creator")
	if err := app.CreateGame("creator", "DFLT"); err != nil {
		t.Fatalf("create game: %v", err)
	}

	snapshot, err := app.GetSnapshot("DFLT", "creator")
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snapshot.Game.CrimePackLanguage != "fa" {
		t.Fatalf("expected Persian crime pack language, got %q", snapshot.Game.CrimePackLanguage)
	}
	if snapshot.Game.CrimePackAssetSetID != "gouache-treachery" {
		t.Fatalf("expected gouache asset set, got %q", snapshot.Game.CrimePackAssetSetID)
	}
	if snapshot.Game.HintPackLanguage != "fa" {
		t.Fatalf("expected Persian hint pack language, got %q", snapshot.Game.HintPackLanguage)
	}
}

func TestStartGameBuildsSnapshotWithScientist(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	setProfile(t, app, creator, "Creator")
	if err := app.CreateGame(creator, "ABCD"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	joinPlayer(t, app, "ABCD", "p1", "A")
	joinPlayer(t, app, "ABCD", "p2", "B")
	joinPlayer(t, app, "ABCD", "p3", "C")

	if err := app.StartGame("ABCD", creator); err != nil {
		t.Fatalf("start game: %v", err)
	}

	snapshot, err := app.GetSnapshot("ABCD", creator)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snapshot.Game == nil || snapshot.Game.StartedOn == "" {
		t.Fatalf("expected started game, got %+v", snapshot.Game)
	}
	if snapshot.Game.ScientistUID == "" {
		t.Fatalf("expected scientist to be assigned")
	}
	if len(snapshot.Players) != 3 {
		t.Fatalf("expected 3 suspects, got %d", len(snapshot.Players))
	}
	for _, player := range snapshot.Players {
		if player.UID == snapshot.Game.ScientistUID {
			t.Fatalf("scientist should not receive suspect cards: %+v", player)
		}
		if len(player.ClueCards) != 4 || len(player.MeansCards) != 4 {
			t.Fatalf("expected 4 clue and 4 means cards for %s, got %d and %d", player.UID, len(player.ClueCards), len(player.MeansCards))
		}
	}

	scientistSnapshot, err := app.GetSnapshot("ABCD", snapshot.Game.ScientistUID)
	if err != nil {
		t.Fatalf("scientist snapshot: %v", err)
	}
	if !scientistSnapshot.Viewer.IsScientist {
		t.Fatalf("expected scientist viewer, got %+v", scientistSnapshot.Viewer)
	}
	if scientistSnapshot.ForensicPrivateData == nil || scientistSnapshot.ForensicPrivateData.Murderer == nil {
		t.Fatalf("expected forensic private data with murderer, got %+v", scientistSnapshot.ForensicPrivateData)
	}
}

func TestMarkedScientistSelectedOnStart(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	setProfile(t, app, creator, "Creator")
	if err := app.CreateGame(creator, "MARK"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	joinPlayer(t, app, "MARK", "p1", "A")
	joinPlayer(t, app, "MARK", "p2", "B")
	joinPlayer(t, app, "MARK", "p3", "C")

	if err := app.ToggleScientistMark("MARK", creator, "p2"); err != nil {
		t.Fatalf("mark scientist: %v", err)
	}
	if err := app.StartGame("MARK", creator); err != nil {
		t.Fatalf("start game: %v", err)
	}

	snapshot, err := app.GetSnapshot("MARK", creator)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snapshot.Game.ScientistUID != "p2" {
		t.Fatalf("expected marked scientist p2, got %s", snapshot.Game.ScientistUID)
	}
}

func TestRoleChangeClearsScientistMark(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	setProfile(t, app, creator, "Creator")
	if err := app.CreateGame(creator, "ROLE"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	joinPlayer(t, app, "ROLE", "p1", "A")
	joinPlayer(t, app, "ROLE", "p2", "B")
	joinPlayer(t, app, "ROLE", "p3", "C")

	if err := app.ToggleScientistMark("ROLE", creator, "p1"); err != nil {
		t.Fatalf("mark scientist: %v", err)
	}
	if err := app.SetParticipantRole("ROLE", creator, "p1", ParticipantRoleObserver); err != nil {
		t.Fatalf("set role: %v", err)
	}

	snapshot, err := app.GetSnapshot("ROLE", creator)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snapshot.Game.MarkedScientistUID != "" {
		t.Fatalf("expected scientist mark to clear, got %q", snapshot.Game.MarkedScientistUID)
	}
}

func TestRoomAuthTokenIsRoomScoped(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	setProfile(t, app, "creator1", "Creator 1")
	setProfile(t, app, "creator2", "Creator 2")
	if err := app.CreateGame("creator1", "AAAA"); err != nil {
		t.Fatalf("create game AAAA: %v", err)
	}
	if err := app.CreateGame("creator2", "BBBB"); err != nil {
		t.Fatalf("create game BBBB: %v", err)
	}
	token, err := app.CreateOrGetRoomAuthToken("AAAA", "creator1")
	if err != nil {
		t.Fatalf("create room auth: %v", err)
	}
	uid, err := app.ResolveRoomAuthToken("AAAA", token.Token)
	if err != nil {
		t.Fatalf("resolve room auth in same room: %v", err)
	}
	if uid != "creator1" {
		t.Fatalf("expected creator1 uid, got %s", uid)
	}
	if _, err := app.ResolveRoomAuthToken("BBBB", token.Token); err == nil {
		t.Fatalf("expected room auth token to fail in other room")
	}
}

func TestCorrectGuessFinishesGame(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	setProfile(t, app, creator, "Creator")
	if err := app.CreateGame(creator, "WXYZ"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	for _, uid := range []string{"p1", "p2", "p3"} {
		joinPlayer(t, app, "WXYZ", uid, uid)
	}
	if err := app.StartGame("WXYZ", creator); err != nil {
		t.Fatalf("start game: %v", err)
	}

	snapshot, err := app.GetSnapshot("WXYZ", creator)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	murdererUID := snapshot.Game.MurdererUID
	scientistSnapshot, err := app.GetSnapshot("WXYZ", snapshot.Game.ScientistUID)
	if err != nil {
		t.Fatalf("scientist snapshot: %v", err)
	}
	murderer := scientistSnapshot.ForensicPrivateData.Murderer
	if murderer == nil {
		t.Fatalf("expected murderer in forensic private data")
	}
	clue := murderer.ClueCards[0].Name
	means := murderer.MeansCards[0].Name
	if err := app.SelectMurdererCards("WXYZ", murdererUID, clue, means); err != nil {
		t.Fatalf("select murderer cards: %v", err)
	}

	guesser := "creator"
	if guesser == murdererUID || snapshot.Game.ScientistUID == guesser {
		guesser = "p1"
	}
	if guesser == murdererUID || snapshot.Game.ScientistUID == guesser {
		guesser = "p2"
	}
	if err := app.MakeGuess("WXYZ", guesser, murdererUID, clue, means); err != nil {
		t.Fatalf("make guess: %v", err)
	}

	updated, err := app.GetSnapshot("WXYZ", creator)
	if err != nil {
		t.Fatalf("snapshot after guess: %v", err)
	}
	if !updated.Game.Finished {
		t.Fatalf("expected finished game after correct guess")
	}
	if len(updated.Guesses) != 1 || !updated.Guesses[0].Correct {
		t.Fatalf("expected one correct guess, got %+v", updated.Guesses)
	}
}

func TestDuplicateExactGuessRejectedButDifferentComboAllowed(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	setProfile(t, app, creator, "Creator")
	if err := app.CreateGame(creator, "DUPL"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	for _, uid := range []string{"p1", "p2", "p3", "p4"} {
		joinPlayer(t, app, "DUPL", uid, uid)
	}
	if err := app.StartGame("DUPL", creator); err != nil {
		t.Fatalf("start game: %v", err)
	}

	creatorSnapshot, err := app.GetSnapshot("DUPL", creator)
	if err != nil {
		t.Fatalf("creator snapshot: %v", err)
	}
	scientistSnapshot, err := app.GetSnapshot("DUPL", creatorSnapshot.Game.ScientistUID)
	if err != nil {
		t.Fatalf("scientist snapshot: %v", err)
	}
	murderer := scientistSnapshot.ForensicPrivateData.Murderer
	murdererUID := murderer.UID
	clue1 := murderer.ClueCards[0].Name
	clue2 := murderer.ClueCards[1].Name
	means1 := murderer.MeansCards[0].Name
	means2 := murderer.MeansCards[1].Name
	if err := app.SelectMurdererCards("DUPL", murdererUID, clue1, means1); err != nil {
		t.Fatalf("select murderer cards: %v", err)
	}

	guessers := []string{}
	for _, player := range creatorSnapshot.Players {
		if player.UID != murdererUID {
			guessers = append(guessers, player.UID)
		}
	}
	if len(guessers) < 2 {
		t.Fatalf("expected at least 2 non-murderer guessers, got %v", guessers)
	}

	if err := app.MakeGuess("DUPL", guessers[0], murdererUID, clue2, means2); err != nil {
		t.Fatalf("first guess: %v", err)
	}
	if err := app.MakeGuess("DUPL", guessers[1], murdererUID, clue2, means2); !errors.Is(err, ErrBadInput) {
		t.Fatalf("expected duplicate exact guess to fail with bad input, got %v", err)
	}
	if err := app.MakeGuess("DUPL", guessers[1], murdererUID, clue2, means1); err != nil {
		t.Fatalf("expected different combo for same suspect to be allowed, got %v", err)
	}
}

func TestPlayerStillGetsOnlyOneGuess(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	setProfile(t, app, creator, "Creator")
	if err := app.CreateGame(creator, "ONCE"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	for _, uid := range []string{"p1", "p2", "p3", "p4"} {
		joinPlayer(t, app, "ONCE", uid, uid)
	}
	if err := app.StartGame("ONCE", creator); err != nil {
		t.Fatalf("start game: %v", err)
	}

	creatorSnapshot, err := app.GetSnapshot("ONCE", creator)
	if err != nil {
		t.Fatalf("creator snapshot: %v", err)
	}
	scientistSnapshot, err := app.GetSnapshot("ONCE", creatorSnapshot.Game.ScientistUID)
	if err != nil {
		t.Fatalf("scientist snapshot: %v", err)
	}
	murderer := scientistSnapshot.ForensicPrivateData.Murderer
	murdererUID := murderer.UID
	clue1 := murderer.ClueCards[0].Name
	clue2 := murderer.ClueCards[1].Name
	means1 := murderer.MeansCards[0].Name
	means2 := murderer.MeansCards[1].Name
	if err := app.SelectMurdererCards("ONCE", murdererUID, clue1, means1); err != nil {
		t.Fatalf("select murderer cards: %v", err)
	}

	var guesserUID string
	for _, player := range creatorSnapshot.Players {
		if player.UID != murdererUID {
			guesserUID = player.UID
			break
		}
	}
	if guesserUID == "" {
		t.Fatalf("expected non-murderer guesser")
	}

	if err := app.MakeGuess("ONCE", guesserUID, murdererUID, clue2, means2); err != nil {
		t.Fatalf("first guess: %v", err)
	}
	if err := app.MakeGuess("ONCE", guesserUID, murdererUID, clue2, means1); !errors.Is(err, ErrBadInput) {
		t.Fatalf("expected second guess to fail with bad input, got %v", err)
	}
}

func TestRestartGameKeepsLobbyRosterAndSettingsButClearsRoundState(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	setProfile(t, app, creator, "Creator")
	if err := app.CreateGame(creator, "RSET"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	joinPlayer(t, app, "RSET", "p1", "Player 1")
	joinPlayer(t, app, "RSET", "p2", "Player 2")
	joinPlayer(t, app, "RSET", "p3", "Player 3")
	joinPlayer(t, app, "RSET", "p4", "Player 4")
	if err := app.SetParticipantRole("RSET", creator, "p4", ParticipantRoleObserver); err != nil {
		t.Fatalf("set observer: %v", err)
	}
	if err := app.ToggleScientistMark("RSET", creator, "p2"); err != nil {
		t.Fatalf("mark scientist: %v", err)
	}
	if err := app.UpdateGameSettings("RSET", creator, GameSettingsInput{
		MeansCardsPerPlayer:  5,
		ClueCardsPerPlayer:   3,
		LinkClueCountToMeans: false,
		MeansCluesTextOnly:   true,
		AccompliceCount:      1,
		WitnessCount:         0,
		WitnessesToFind:      0,
	}); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	roomAuthToken, err := app.CreateOrGetRoomAuthToken("RSET", creator)
	if err != nil {
		t.Fatalf("create room auth: %v", err)
	}
	if err := app.StartGame("RSET", creator); err != nil {
		t.Fatalf("start game: %v", err)
	}

	startedSnapshot, err := app.GetSnapshot("RSET", creator)
	if err != nil {
		t.Fatalf("started snapshot: %v", err)
	}
	scientistSnapshot, err := app.GetSnapshot("RSET", startedSnapshot.Game.ScientistUID)
	if err != nil {
		t.Fatalf("scientist snapshot: %v", err)
	}
	murderer := scientistSnapshot.ForensicPrivateData.Murderer
	murdererUID := murderer.UID
	if err := app.SelectMurdererCards("RSET", murdererUID, murderer.ClueCards[0].Name, murderer.MeansCards[0].Name); err != nil {
		t.Fatalf("select murderer cards: %v", err)
	}

	var guesserUID string
	for _, player := range startedSnapshot.Players {
		if player.UID != murdererUID {
			guesserUID = player.UID
			break
		}
	}
	if guesserUID == "" {
		t.Fatalf("expected non-murderer guesser")
	}
	if err := app.MakeGuess("RSET", guesserUID, murdererUID, murderer.ClueCards[1].Name, murderer.MeansCards[1].Name); err != nil {
		t.Fatalf("make guess: %v", err)
	}
	if err := app.SendChatMessage("RSET", guesserUID, "still thinking"); err != nil {
		t.Fatalf("send chat message: %v", err)
	}
	if err := app.StartRoomTimer("RSET", creator, nil); err != nil {
		t.Fatalf("start room timer: %v", err)
	}
	if err := app.withTx(context.Background(), func(tx *sql.Tx) error {
		return app.upsertWitnessSelectionPromptTx(tx, "RSET", murdererUID, WitnessSelectionPromptState{
			RequiredSelections: 1,
			Dismissible:        true,
			CreatorInitiated:   true,
		})
	}); err != nil {
		t.Fatalf("seed witness prompt: %v", err)
	}
	if err := app.EndGame("RSET", creator); err != nil {
		t.Fatalf("end game: %v", err)
	}

	if err := app.RestartGame("RSET", creator); err != nil {
		t.Fatalf("restart game: %v", err)
	}

	restartedSnapshot, err := app.GetSnapshot("RSET", creator)
	if err != nil {
		t.Fatalf("restarted snapshot: %v", err)
	}
	if restartedSnapshot.Game.GameID != "RSET" {
		t.Fatalf("expected same game id, got %+v", restartedSnapshot.Game)
	}
	if restartedSnapshot.Game.StartedOn != "" || restartedSnapshot.Game.StartedTimestamp != "" {
		t.Fatalf("expected lobby game after restart, got %+v", restartedSnapshot.Game)
	}
	if restartedSnapshot.Game.Finished || restartedSnapshot.Game.PendingWitnessSelection {
		t.Fatalf("expected unfinished lobby game after restart, got %+v", restartedSnapshot.Game)
	}
	if restartedSnapshot.Game.MeansCardsPerPlayer != 5 || restartedSnapshot.Game.ClueCardsPerPlayer != 3 || !restartedSnapshot.Game.MeansCluesTextOnly {
		t.Fatalf("expected settings to persist, got %+v", restartedSnapshot.Game)
	}
	if restartedSnapshot.Game.MarkedScientistUID != "p2" {
		t.Fatalf("expected marked scientist to persist, got %q", restartedSnapshot.Game.MarkedScientistUID)
	}
	if restartedSnapshot.Game.ScientistUID != "" || restartedSnapshot.Game.MurdererUID != "" || restartedSnapshot.Game.MurdererCardsSelected {
		t.Fatalf("expected round assignments to clear, got %+v", restartedSnapshot.Game)
	}
	if restartedSnapshot.Game.CauseCard != nil || restartedSnapshot.Game.LocationCard != nil || len(restartedSnapshot.Game.OtherCards) != 0 {
		t.Fatalf("expected forensic cards cleared, got %+v", restartedSnapshot.Game)
	}
	if restartedSnapshot.Game.RoomTimer != nil {
		t.Fatalf("expected room timer cleared, got %+v", restartedSnapshot.Game.RoomTimer)
	}
	if len(restartedSnapshot.Players) != 0 {
		t.Fatalf("expected dealt suspect state cleared, got %+v", restartedSnapshot.Players)
	}
	if len(restartedSnapshot.Guesses) != 0 {
		t.Fatalf("expected guesses cleared, got %+v", restartedSnapshot.Guesses)
	}
	if len(restartedSnapshot.Messages) != 0 {
		t.Fatalf("expected messages cleared, got %+v", restartedSnapshot.Messages)
	}
	if restartedSnapshot.ModeratorPrivateData != nil {
		t.Fatalf("expected moderator private data cleared in lobby, got %+v", restartedSnapshot.ModeratorPrivateData)
	}
	if len(restartedSnapshot.RoleReveal) != 0 {
		t.Fatalf("expected role reveal cleared after restart, got %+v", restartedSnapshot.RoleReveal)
	}
	if resolvedUID, err := app.ResolveRoomAuthToken("RSET", roomAuthToken.Token); err != nil || resolvedUID != creator {
		t.Fatalf("expected room auth token to survive restart, got uid=%q err=%v", resolvedUID, err)
	}

	roleByUID := map[string]ParticipantRole{}
	for _, participant := range restartedSnapshot.Participants {
		roleByUID[participant.UID] = participant.Role
	}
	if len(roleByUID) != 5 {
		t.Fatalf("expected full participant roster to persist, got %+v", restartedSnapshot.Participants)
	}
	if roleByUID["creator"] != ParticipantRolePlayer || roleByUID["p1"] != ParticipantRolePlayer || roleByUID["p2"] != ParticipantRolePlayer || roleByUID["p3"] != ParticipantRolePlayer || roleByUID["p4"] != ParticipantRoleObserver {
		t.Fatalf("expected participant roles to persist, got %+v", restartedSnapshot.Participants)
	}
}

func TestUpdateGameSettingsAndDealConfiguredCounts(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	setProfile(t, app, "creator", "Creator")
	if err := app.CreateGame("creator", "SETT"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	joinPlayers(t, app, "SETT", "p1", "p2", "p3", "p4")

	if err := app.UpdateGameSettings("SETT", "creator", GameSettingsInput{
		MeansCardsPerPlayer:  5,
		ClueCardsPerPlayer:   3,
		LinkClueCountToMeans: false,
		MeansCluesTextOnly:   true,
		AccompliceCount:      1,
		WitnessCount:         0,
		WitnessesToFind:      0,
	}); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if err := app.StartGame("SETT", "creator"); err != nil {
		t.Fatalf("start game: %v", err)
	}

	snapshot, err := app.GetSnapshot("SETT", "creator")
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snapshot.Game.MeansCardsPerPlayer != 5 || snapshot.Game.ClueCardsPerPlayer != 3 || !snapshot.Game.MeansCluesTextOnly {
		t.Fatalf("unexpected settings on game: %+v", snapshot.Game)
	}
	if snapshot.Game.RandomMurdererCardSelection {
		t.Fatalf("expected random murderer card selection to default off")
	}
	for _, player := range snapshot.Players {
		if len(player.MeansCards) != 5 || len(player.ClueCards) != 3 {
			t.Fatalf("unexpected dealt counts for %s: %d means, %d clue", player.UID, len(player.MeansCards), len(player.ClueCards))
		}
	}
}

func TestRandomMurdererCardSelectionOnStart(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	setProfile(t, app, "creator", "Creator")
	if err := app.CreateGame("creator", "RAND"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	joinPlayers(t, app, "RAND", "p1", "p2", "p3")

	if err := app.UpdateGameSettings("RAND", "creator", GameSettingsInput{
		MeansCardsPerPlayer:         4,
		ClueCardsPerPlayer:          4,
		LinkClueCountToMeans:        true,
		MeansCluesTextOnly:          false,
		RandomMurdererCardSelection: true,
		AccompliceCount:             0,
		WitnessCount:                0,
		WitnessesToFind:             0,
	}); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if err := app.StartGame("RAND", "creator"); err != nil {
		t.Fatalf("start game: %v", err)
	}

	snapshot, err := app.GetSnapshot("RAND", "creator")
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if !snapshot.Game.RandomMurdererCardSelection || !snapshot.Game.MurdererCardsSelected {
		t.Fatalf("expected random murderer card selection to complete on start, got %+v", snapshot.Game)
	}
	if snapshot.Game.MurdererClueCardID == "" || snapshot.Game.MurdererMeansCardID == "" {
		t.Fatalf("expected selected murderer card ids, got %+v", snapshot.Game)
	}
	murderer := findPlayer(snapshot.Players, snapshot.Game.MurdererUID)
	if murderer == nil {
		t.Fatalf("expected murderer player %q in snapshot", snapshot.Game.MurdererUID)
	}
	if !hasCard(murderer.ClueCards, snapshot.Game.MurdererClueCardID) || !hasCard(murderer.MeansCards, snapshot.Game.MurdererMeansCardID) {
		t.Fatalf("expected random cards to belong to murderer, got clue=%q means=%q murderer=%+v", snapshot.Game.MurdererClueCardID, snapshot.Game.MurdererMeansCardID, murderer)
	}
}

func TestRoomModsDefaultOffAndCanBeUpdatedMidGame(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	startTestGame(t, app, "MODS", "creator", "p1", "p2", "p3")

	beforeSnapshot, err := app.GetSnapshot("MODS", "creator")
	if err != nil {
		t.Fatalf("snapshot before room mod update: %v", err)
	}
	if beforeSnapshot.Game.MeansCluesTextOnly {
		t.Fatalf("expected means/clues text-only room mod to default off")
	}

	if err := app.UpdateRoomMods("MODS", "creator", GameRoomModsInput{MeansCluesTextOnly: true}); err != nil {
		t.Fatalf("update room mods: %v", err)
	}

	creatorSnapshot, err := app.GetSnapshot("MODS", "creator")
	if err != nil {
		t.Fatalf("creator snapshot after room mod update: %v", err)
	}
	if !creatorSnapshot.Game.MeansCluesTextOnly {
		t.Fatalf("expected creator snapshot to include updated room mod")
	}

	otherSnapshot, err := app.GetSnapshot("MODS", "p1")
	if err != nil {
		t.Fatalf("other snapshot after room mod update: %v", err)
	}
	if !otherSnapshot.Game.MeansCluesTextOnly {
		t.Fatalf("expected all players to see updated room mod")
	}
}

func TestWitnessSelectionFlowAndAccompliceKnowledge(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	setProfile(t, app, creator, "Creator")
	if err := app.CreateGame(creator, "ROLE2"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	joinPlayers(t, app, "ROLE2", "p1", "p2", "p3", "p4", "p5")

	if err := app.UpdateGameSettings("ROLE2", creator, GameSettingsInput{
		MeansCardsPerPlayer:  4,
		ClueCardsPerPlayer:   4,
		LinkClueCountToMeans: true,
		AccompliceCount:      2,
		WitnessCount:         1,
		WitnessesToFind:      1,
	}); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if err := app.StartGame("ROLE2", creator); err != nil {
		t.Fatalf("start game: %v", err)
	}

	creatorSnapshot, err := app.GetSnapshot("ROLE2", creator)
	if err != nil {
		t.Fatalf("creator snapshot: %v", err)
	}
	scientistUID := creatorSnapshot.Game.ScientistUID

	roleByUID := map[string]SecretRole{}
	var murdererUID string
	var accompliceUIDs []string
	var witnessUID string
	for _, participant := range creatorSnapshot.Participants {
		if participant.UID == scientistUID || participant.Role != ParticipantRolePlayer {
			continue
		}
		snapshot, err := app.GetSnapshot("ROLE2", participant.UID)
		if err != nil {
			t.Fatalf("player snapshot for %s: %v", participant.UID, err)
		}
		roleByUID[participant.UID] = snapshot.PlayerPrivateData.Role
		switch snapshot.PlayerPrivateData.Role {
		case SecretRoleMurderer:
			murdererUID = participant.UID
		case SecretRoleAccomplice:
			accompliceUIDs = append(accompliceUIDs, participant.UID)
			if len(snapshot.PlayerPrivateData.KnownMurdererTeam) != 3 {
				t.Fatalf("expected accomplice %s to know full murderer team, got %+v", participant.UID, snapshot.PlayerPrivateData.KnownMurdererTeam)
			}
		case SecretRoleWitness:
			witnessUID = participant.UID
		}
	}
	if murdererUID == "" || len(accompliceUIDs) != 2 || witnessUID == "" {
		t.Fatalf("unexpected role breakdown: murderer=%s accomplices=%v witness=%s roles=%+v", murdererUID, accompliceUIDs, witnessUID, roleByUID)
	}

	scientistSnapshot, err := app.GetSnapshot("ROLE2", scientistUID)
	if err != nil {
		t.Fatalf("scientist snapshot: %v", err)
	}
	murderer := scientistSnapshot.ForensicPrivateData.Murderer
	clue := murderer.ClueCards[0].Name
	means := murderer.MeansCards[0].Name
	if err := app.SelectMurdererCards("ROLE2", murdererUID, clue, means); err != nil {
		t.Fatalf("select murderer cards: %v", err)
	}

	guesserUID := creator
	if guesserUID == murdererUID || guesserUID == scientistUID || roleByUID[guesserUID] == SecretRoleAccomplice || roleByUID[guesserUID] == SecretRoleWitness {
		for uid, role := range roleByUID {
			if uid != murdererUID && uid != scientistUID && role == SecretRoleInvestigator {
				guesserUID = uid
				break
			}
		}
	}
	if err := app.MakeGuess("ROLE2", guesserUID, murdererUID, clue, means); err != nil {
		t.Fatalf("make guess: %v", err)
	}

	updatedCreator, err := app.GetSnapshot("ROLE2", creator)
	if err != nil {
		t.Fatalf("creator snapshot after guess: %v", err)
	}
	if updatedCreator.Game.Finished || !updatedCreator.Game.PendingWitnessSelection {
		t.Fatalf("expected pending witness selection after correct guess, got %+v", updatedCreator.Game)
	}
	if updatedCreator.ModeratorPrivateData == nil {
		t.Fatalf("expected moderator private data during witness selection")
	}
	if len(updatedCreator.ModeratorPrivateData.WitnessPromptCandidates) != len(updatedCreator.Players) {
		t.Fatalf("expected blind moderator witness prompt candidates for every suspect, got %+v with players %+v", updatedCreator.ModeratorPrivateData, updatedCreator.Players)
	}
	candidateByUID := map[string]string{}
	for _, candidate := range updatedCreator.ModeratorPrivateData.WitnessPromptCandidates {
		candidateByUID[candidate.UID] = candidate.Name
	}
	creatorIsSuspect := false
	for _, player := range updatedCreator.Players {
		if candidateByUID[player.UID] != player.Name {
			t.Fatalf("expected blind candidate for player %+v, got %+v", player, updatedCreator.ModeratorPrivateData.WitnessPromptCandidates)
		}
		if player.UID == creator {
			creatorIsSuspect = true
		}
	}
	if creatorIsSuspect {
		if _, ok := candidateByUID[creator]; !ok {
			t.Fatalf("expected creator to appear in blind witness prompt candidates when they are a suspect, got %+v", updatedCreator.ModeratorPrivateData.WitnessPromptCandidates)
		}
	}

	murdererSnapshot, err := app.GetSnapshot("ROLE2", murdererUID)
	if err != nil {
		t.Fatalf("murderer snapshot before blind prompt: %v", err)
	}
	if murdererSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt == nil || murdererSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt.Dismissible {
		t.Fatalf("expected murderer to keep the built-in non-dismissible prompt, got %+v", murdererSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt)
	}
	if err := app.ShowWitnessSelectionPrompt("ROLE2", creator, murdererUID); err != nil {
		t.Fatalf("show witness selection prompt for murderer should quietly succeed: %v", err)
	}
	murdererSnapshot, err = app.GetSnapshot("ROLE2", murdererUID)
	if err != nil {
		t.Fatalf("murderer snapshot after blind prompt: %v", err)
	}
	if murdererSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt == nil || murdererSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt.Dismissible {
		t.Fatalf("expected murderer prompt to remain non-dismissible after blind creator click, got %+v", murdererSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt)
	}
	if err := app.ShowWitnessSelectionPrompt("ROLE2", creator, witnessUID); err != nil {
		t.Fatalf("show witness selection prompt for witness should quietly succeed: %v", err)
	}
	witnessSnapshot, err := app.GetSnapshot("ROLE2", witnessUID)
	if err != nil {
		t.Fatalf("witness snapshot after blind prompt: %v", err)
	}
	if witnessSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt != nil {
		t.Fatalf("expected witness blind prompt click to remain a no-op, got %+v", witnessSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt)
	}

	accompliceUID := accompliceUIDs[0]
	if err := app.ShowWitnessSelectionPrompt("ROLE2", creator, accompliceUID); err != nil {
		t.Fatalf("show witness selection prompt: %v", err)
	}
	accompliceSnapshot, err := app.GetSnapshot("ROLE2", accompliceUID)
	if err != nil {
		t.Fatalf("accomplice snapshot: %v", err)
	}
	if accompliceSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt == nil || !accompliceSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt.Dismissible {
		t.Fatalf("expected dismissible accomplice prompt, got %+v", accompliceSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt)
	}
	if err := app.DismissWitnessSelectionPrompt("ROLE2", accompliceUID); err != nil {
		t.Fatalf("dismiss accomplice prompt: %v", err)
	}
	accompliceSnapshot, err = app.GetSnapshot("ROLE2", accompliceUID)
	if err != nil {
		t.Fatalf("accomplice snapshot after dismiss: %v", err)
	}
	if accompliceSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt != nil {
		t.Fatalf("expected accomplice prompt to be dismissed, got %+v", accompliceSnapshot.PlayerPrivateData.ActiveWitnessSelectionPrompt)
	}
	if err := app.ShowWitnessSelectionPrompt("ROLE2", creator, accompliceUID); err != nil {
		t.Fatalf("show witness selection prompt again: %v", err)
	}
	if err := app.SubmitWitnessSelection("ROLE2", accompliceUID, []string{witnessUID}); err != nil {
		t.Fatalf("submit witness selection: %v", err)
	}

	finalSnapshot, err := app.GetSnapshot("ROLE2", creator)
	if err != nil {
		t.Fatalf("final snapshot: %v", err)
	}
	if !finalSnapshot.Game.Finished || finalSnapshot.Game.Winner != WinnerMurdererTeam {
		t.Fatalf("expected murderer team to win after finding the witness, got %+v", finalSnapshot.Game)
	}
}

func TestBadTeamWinsOnceAllGoodTeamGuessesAreUsed(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	setProfile(t, app, creator, "Creator")
	if err := app.CreateGame(creator, "EXHA"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	joinPlayers(t, app, "EXHA", "p1", "p2", "p3", "p4")
	if err := app.UpdateGameSettings("EXHA", creator, GameSettingsInput{
		MeansCardsPerPlayer:  4,
		ClueCardsPerPlayer:   4,
		LinkClueCountToMeans: true,
		AccompliceCount:      1,
		WitnessCount:         1,
		WitnessesToFind:      1,
	}); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if err := app.StartGame("EXHA", creator); err != nil {
		t.Fatalf("start game: %v", err)
	}

	creatorSnapshot, err := app.GetSnapshot("EXHA", creator)
	if err != nil {
		t.Fatalf("creator snapshot: %v", err)
	}
	scientistUID := creatorSnapshot.Game.ScientistUID

	var murdererUID string
	var accompliceUID string
	goodTeamUIDs := []string{}
	for _, participant := range creatorSnapshot.Participants {
		if participant.UID == scientistUID || participant.Role != ParticipantRolePlayer {
			continue
		}
		snapshot, err := app.GetSnapshot("EXHA", participant.UID)
		if err != nil {
			t.Fatalf("player snapshot for %s: %v", participant.UID, err)
		}
		switch snapshot.PlayerPrivateData.Role {
		case SecretRoleMurderer:
			murdererUID = participant.UID
		case SecretRoleAccomplice:
			accompliceUID = participant.UID
		default:
			goodTeamUIDs = append(goodTeamUIDs, participant.UID)
		}
	}
	if murdererUID == "" || accompliceUID == "" || len(goodTeamUIDs) != 2 {
		t.Fatalf("unexpected role assignment: murderer=%q accomplice=%q good=%v", murdererUID, accompliceUID, goodTeamUIDs)
	}

	scientistSnapshot, err := app.GetSnapshot("EXHA", scientistUID)
	if err != nil {
		t.Fatalf("scientist snapshot: %v", err)
	}
	murderer := scientistSnapshot.ForensicPrivateData.Murderer
	if err := app.SelectMurdererCards("EXHA", murdererUID, murderer.ClueCards[0].Name, murderer.MeansCards[0].Name); err != nil {
		t.Fatalf("select murderer cards: %v", err)
	}

	firstWrongTarget := accompliceUID
	secondWrongTarget := goodTeamUIDs[0]
	if secondWrongTarget == firstWrongTarget {
		secondWrongTarget = goodTeamUIDs[1]
	}
	if err := app.MakeGuess("EXHA", goodTeamUIDs[0], firstWrongTarget, murderer.ClueCards[0].Name, murderer.MeansCards[0].Name); err != nil {
		t.Fatalf("first wrong guess: %v", err)
	}

	midSnapshot, err := app.GetSnapshot("EXHA", creator)
	if err != nil {
		t.Fatalf("mid snapshot: %v", err)
	}
	if midSnapshot.Game.Finished {
		t.Fatalf("expected game to continue until every good-team guess is used, got %+v", midSnapshot.Game)
	}

	if err := app.MakeGuess("EXHA", goodTeamUIDs[1], secondWrongTarget, murderer.ClueCards[1].Name, murderer.MeansCards[1].Name); err != nil {
		t.Fatalf("second wrong guess: %v", err)
	}

	finalSnapshot, err := app.GetSnapshot("EXHA", creator)
	if err != nil {
		t.Fatalf("final snapshot: %v", err)
	}
	if !finalSnapshot.Game.Finished || finalSnapshot.Game.Winner != WinnerMurdererTeam {
		t.Fatalf("expected murderer team win after all good-team guesses are used, got %+v", finalSnapshot.Game)
	}
	if finalSnapshot.Game.FinishedReason != "all-guesses-used" {
		t.Fatalf("expected all-guesses-used finish reason, got %+v", finalSnapshot.Game)
	}
}

func TestRoomTimerLifecycle(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	currentTime := time.Date(2026, time.April, 20, 12, 0, 0, 0, time.UTC)
	app.timeNow = func() time.Time {
		return currentTime
	}

	startTestGame(t, app, "TIME", "creator", "p1", "p2", "p3")

	if err := app.StartRoomTimer("TIME", "creator", nil); err != nil {
		t.Fatalf("start room timer: %v", err)
	}

	snapshot, err := app.GetSnapshot("TIME", "creator")
	if err != nil {
		t.Fatalf("snapshot after start: %v", err)
	}
	if snapshot.ServerTimestamp != nowTimestamp(currentTime) {
		t.Fatalf("expected server timestamp %s, got %s", nowTimestamp(currentTime), snapshot.ServerTimestamp)
	}
	if snapshot.Game.RoomTimer == nil {
		t.Fatalf("expected room timer to exist")
	}
	if snapshot.Game.RoomTimer.DurationSeconds != defaultRoomTimerDurationSeconds {
		t.Fatalf("expected default timer duration, got %+v", snapshot.Game.RoomTimer)
	}
	if snapshot.Game.RoomTimer.ExpiresAt != nowTimestamp(currentTime.Add(40*time.Second)) {
		t.Fatalf("unexpected timer expiry after start: %+v", snapshot.Game.RoomTimer)
	}
	if snapshot.Game.RoomTimer.RunID != 1 {
		t.Fatalf("expected first run id 1, got %+v", snapshot.Game.RoomTimer)
	}

	currentTime = currentTime.Add(10 * time.Second)
	if err := app.PauseRoomTimer("TIME", "creator"); err != nil {
		t.Fatalf("pause room timer: %v", err)
	}
	snapshot, err = app.GetSnapshot("TIME", "creator")
	if err != nil {
		t.Fatalf("snapshot after pause: %v", err)
	}
	if snapshot.Game.RoomTimer.ExpiresAt != "" || snapshot.Game.RoomTimer.PausedRemainingSeconds != 30 || snapshot.Game.RoomTimer.RunID != 1 {
		t.Fatalf("unexpected paused timer state: %+v", snapshot.Game.RoomTimer)
	}

	currentTime = currentTime.Add(15 * time.Second)
	if err := app.ResumeRoomTimer("TIME", "creator"); err != nil {
		t.Fatalf("resume room timer: %v", err)
	}
	snapshot, err = app.GetSnapshot("TIME", "creator")
	if err != nil {
		t.Fatalf("snapshot after resume: %v", err)
	}
	if snapshot.Game.RoomTimer.PausedRemainingSeconds != 0 {
		t.Fatalf("expected resumed timer to clear paused seconds, got %+v", snapshot.Game.RoomTimer)
	}
	if snapshot.Game.RoomTimer.ExpiresAt != nowTimestamp(currentTime.Add(30*time.Second)) {
		t.Fatalf("unexpected timer expiry after resume: %+v", snapshot.Game.RoomTimer)
	}
	if snapshot.Game.RoomTimer.RunID != 1 {
		t.Fatalf("resume should preserve run id, got %+v", snapshot.Game.RoomTimer)
	}

	currentTime = currentTime.Add(5 * time.Second)
	if err := app.ResetRoomTimer("TIME", "creator"); err != nil {
		t.Fatalf("reset room timer: %v", err)
	}
	snapshot, err = app.GetSnapshot("TIME", "creator")
	if err != nil {
		t.Fatalf("snapshot after reset: %v", err)
	}
	if snapshot.Game.RoomTimer.DurationSeconds != 40 || snapshot.Game.RoomTimer.ExpiresAt != nowTimestamp(currentTime.Add(40*time.Second)) {
		t.Fatalf("unexpected timer after reset: %+v", snapshot.Game.RoomTimer)
	}
	if snapshot.Game.RoomTimer.RunID != 2 {
		t.Fatalf("expected reset to increment run id, got %+v", snapshot.Game.RoomTimer)
	}

	if err := app.ClearRoomTimer("TIME", "creator"); err != nil {
		t.Fatalf("clear room timer: %v", err)
	}
	snapshot, err = app.GetSnapshot("TIME", "creator")
	if err != nil {
		t.Fatalf("snapshot after clear: %v", err)
	}
	if snapshot.Game.RoomTimer != nil {
		t.Fatalf("expected timer to clear, got %+v", snapshot.Game.RoomTimer)
	}

	if err := app.StartRoomTimer("TIME", "creator", intPtr(15)); err != nil {
		t.Fatalf("restart room timer after clear: %v", err)
	}
	snapshot, err = app.GetSnapshot("TIME", "creator")
	if err != nil {
		t.Fatalf("snapshot after restart: %v", err)
	}
	if snapshot.Game.RoomTimer == nil || snapshot.Game.RoomTimer.RunID != 3 || snapshot.Game.RoomTimer.DurationSeconds != 15 {
		t.Fatalf("expected new timer run after clear, got %+v", snapshot.Game.RoomTimer)
	}
}

func TestRoomTimerRejectsInvalidStatesAndActors(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	setProfile(t, app, "creator", "Creator")
	if err := app.CreateGame("creator", "LOCK"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	joinPlayers(t, app, "LOCK", "p1", "p2", "p3")

	if err := app.StartRoomTimer("LOCK", "creator", nil); !errors.Is(err, ErrBadInput) {
		t.Fatalf("expected bad input before game start, got %v", err)
	}
	if err := app.UpdateRoomMods("LOCK", "creator", GameRoomModsInput{MeansCluesTextOnly: true}); !errors.Is(err, ErrBadInput) {
		t.Fatalf("expected bad input for room mods before game start, got %v", err)
	}
	if err := app.StartRoomTimer("LOCK", "creator", intPtr(0)); !errors.Is(err, ErrBadInput) {
		t.Fatalf("expected bad input for zero-second timer, got %v", err)
	}

	if err := app.StartGame("LOCK", "creator"); err != nil {
		t.Fatalf("start game: %v", err)
	}
	if err := app.UpdateRoomMods("LOCK", "p1", GameRoomModsInput{MeansCluesTextOnly: true}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden for non-creator room mod control, got %v", err)
	}
	if err := app.StartRoomTimer("LOCK", "p1", nil); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden for non-creator timer control, got %v", err)
	}
	if err := app.ResetRoomTimer("LOCK", "creator"); !errors.Is(err, ErrBadInput) {
		t.Fatalf("expected bad input when resetting missing timer, got %v", err)
	}
	if err := app.EndGame("LOCK", "creator"); err != nil {
		t.Fatalf("end game: %v", err)
	}
	if err := app.StartRoomTimer("LOCK", "creator", nil); !errors.Is(err, ErrBadInput) {
		t.Fatalf("expected bad input after game finish, got %v", err)
	}
}

func TestRoomTimerClearsWhenGameEnds(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	startTestGame(t, app, "DONE", "creator", "p1", "p2", "p3")

	if err := app.StartRoomTimer("DONE", "creator", intPtr(20)); err != nil {
		t.Fatalf("start room timer: %v", err)
	}
	if err := app.EndGame("DONE", "creator"); err != nil {
		t.Fatalf("end game: %v", err)
	}

	snapshot, err := app.GetSnapshot("DONE", "creator")
	if err != nil {
		t.Fatalf("snapshot after end: %v", err)
	}
	if snapshot.Game.RoomTimer != nil {
		t.Fatalf("expected room timer to clear when game ends, got %+v", snapshot.Game.RoomTimer)
	}
}

func intPtr(value int) *int {
	return &value
}

func TestCreatorWitnessPromptCandidatesIncludeCreatorWhenTheyAreASuspect(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	setProfile(t, app, creator, "Creator")
	if err := app.CreateGame(creator, "BLND"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	joinPlayers(t, app, "BLND", "p1", "p2", "p3", "p4")
	if err := app.ToggleScientistMark("BLND", creator, "p1"); err != nil {
		t.Fatalf("mark scientist: %v", err)
	}
	if err := app.UpdateGameSettings("BLND", creator, GameSettingsInput{
		MeansCardsPerPlayer:  4,
		ClueCardsPerPlayer:   4,
		LinkClueCountToMeans: true,
		AccompliceCount:      1,
		WitnessCount:         1,
		WitnessesToFind:      1,
	}); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if err := app.StartGame("BLND", creator); err != nil {
		t.Fatalf("start game: %v", err)
	}

	creatorSnapshot, err := app.GetSnapshot("BLND", creator)
	if err != nil {
		t.Fatalf("creator snapshot after start: %v", err)
	}
	if creatorSnapshot.Game.ScientistUID != "p1" {
		t.Fatalf("expected marked scientist p1, got %s", creatorSnapshot.Game.ScientistUID)
	}
	if len(creatorSnapshot.Players) != 4 {
		t.Fatalf("expected 4 suspects, got %d", len(creatorSnapshot.Players))
	}

	roleByUID := map[string]SecretRole{}
	var murdererUID, witnessUID, investigatorUID string
	for _, player := range creatorSnapshot.Players {
		snapshot, err := app.GetSnapshot("BLND", player.UID)
		if err != nil {
			t.Fatalf("player snapshot for %s: %v", player.UID, err)
		}
		roleByUID[player.UID] = snapshot.PlayerPrivateData.Role
		switch snapshot.PlayerPrivateData.Role {
		case SecretRoleMurderer:
			murdererUID = player.UID
		case SecretRoleWitness:
			witnessUID = player.UID
		case SecretRoleInvestigator:
			if investigatorUID == "" {
				investigatorUID = player.UID
			}
		}
	}
	if roleByUID[creator] == "" {
		t.Fatalf("expected creator to remain a suspect, roles=%+v", roleByUID)
	}
	if murdererUID == "" || witnessUID == "" || investigatorUID == "" {
		t.Fatalf("unexpected role breakdown murderer=%s witness=%s investigator=%s roles=%+v", murdererUID, witnessUID, investigatorUID, roleByUID)
	}

	scientistSnapshot, err := app.GetSnapshot("BLND", "p1")
	if err != nil {
		t.Fatalf("scientist snapshot: %v", err)
	}
	murderer := scientistSnapshot.ForensicPrivateData.Murderer
	clue := murderer.ClueCards[0].Name
	means := murderer.MeansCards[0].Name
	if err := app.SelectMurdererCards("BLND", murdererUID, clue, means); err != nil {
		t.Fatalf("select murderer cards: %v", err)
	}
	if err := app.MakeGuess("BLND", investigatorUID, murdererUID, clue, means); err != nil {
		t.Fatalf("make guess: %v", err)
	}

	updatedCreator, err := app.GetSnapshot("BLND", creator)
	if err != nil {
		t.Fatalf("creator snapshot after correct guess: %v", err)
	}
	if !updatedCreator.Game.PendingWitnessSelection {
		t.Fatalf("expected pending witness selection, got %+v", updatedCreator.Game)
	}
	if updatedCreator.ModeratorPrivateData == nil {
		t.Fatalf("expected moderator private data")
	}
	candidateByUID := map[string]bool{}
	for _, candidate := range updatedCreator.ModeratorPrivateData.WitnessPromptCandidates {
		candidateByUID[candidate.UID] = true
	}
	if !candidateByUID[creator] {
		t.Fatalf("expected creator to appear in blind witness prompt candidates, got %+v", updatedCreator.ModeratorPrivateData.WitnessPromptCandidates)
	}
}
