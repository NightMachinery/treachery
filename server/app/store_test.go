package app

import (
	mathrand "math/rand"
	"path/filepath"
	"testing"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	app, err := New(Config{
		DataDir:   t.TempDir(),
		CardsPath: filepath.Join("..", "..", "UI", "src", "assets", "cards.json"),
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
	if snapshot.Game.MeansCardsPerPlayer != 5 || snapshot.Game.ClueCardsPerPlayer != 3 {
		t.Fatalf("unexpected settings on game: %+v", snapshot.Game)
	}
	for _, player := range snapshot.Players {
		if len(player.MeansCards) != 5 || len(player.ClueCards) != 3 {
			t.Fatalf("unexpected dealt counts for %s: %d means, %d clue", player.UID, len(player.MeansCards), len(player.ClueCards))
		}
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
	if len(updatedCreator.ModeratorPrivateData.WitnessPromptTargets) != 3 {
		t.Fatalf("expected moderator prompt targets for murderer team, got %+v", updatedCreator.ModeratorPrivateData)
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
