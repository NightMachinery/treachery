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

func TestStartGameBuildsSnapshot(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	if err := app.CreateGame(creator, "ABCD"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	for i, uid := range []string{"p1", "p2", "p3"} {
		if err := app.AddPlayer("ABCD", uid, string(rune('A'+i))); err != nil {
			t.Fatalf("add player %s: %v", uid, err)
		}
	}
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
	if snapshot.Game.MurdererUID == "" {
		t.Fatalf("expected murderer to be assigned")
	}
	if len(snapshot.Players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(snapshot.Players))
	}
	for _, player := range snapshot.Players {
		if len(player.ClueCards) != 4 || len(player.MeansCards) != 4 {
			t.Fatalf("expected 4 clue and 4 means cards for %s, got %d and %d", player.UID, len(player.ClueCards), len(player.MeansCards))
		}
	}
	if snapshot.ForensicPrivateData == nil || snapshot.ForensicPrivateData.Murderer == nil {
		t.Fatalf("expected forensic private data with murderer, got %+v", snapshot.ForensicPrivateData)
	}

	murdererSnapshot, err := app.GetSnapshot("ABCD", snapshot.Game.MurdererUID)
	if err != nil {
		t.Fatalf("murderer snapshot: %v", err)
	}
	if !murdererSnapshot.PlayerPrivateData.IsMurderer {
		t.Fatalf("expected murderer private data, got %+v", murdererSnapshot.PlayerPrivateData)
	}
}

func TestCorrectGuessFinishesGame(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	creator := "creator"
	if err := app.CreateGame(creator, "WXYZ"); err != nil {
		t.Fatalf("create game: %v", err)
	}
	players := []string{"p1", "p2", "p3", "p4"}
	for _, uid := range players {
		if err := app.AddPlayer("WXYZ", uid, uid); err != nil {
			t.Fatalf("add player %s: %v", uid, err)
		}
	}
	if err := app.StartGame("WXYZ", creator); err != nil {
		t.Fatalf("start game: %v", err)
	}

	forensicSnapshot, err := app.GetSnapshot("WXYZ", creator)
	if err != nil {
		t.Fatalf("forensic snapshot: %v", err)
	}
	murderer := forensicSnapshot.ForensicPrivateData.Murderer
	if murderer == nil {
		t.Fatalf("expected murderer in forensic private data")
	}
	clue := murderer.ClueCards[0].Name
	means := murderer.MeansCards[0].Name
	if err := app.SelectMurdererCards("WXYZ", murderer.UID, clue, means); err != nil {
		t.Fatalf("select murderer cards: %v", err)
	}

	guesser := players[0]
	if guesser == murderer.UID {
		guesser = players[1]
	}
	if err := app.MakeGuess("WXYZ", guesser, murderer.UID, clue, means); err != nil {
		t.Fatalf("make guess: %v", err)
	}

	snapshot, err := app.GetSnapshot("WXYZ", creator)
	if err != nil {
		t.Fatalf("snapshot after guess: %v", err)
	}
	if !snapshot.Game.Finished {
		t.Fatalf("expected finished game after correct guess")
	}
	if len(snapshot.Guesses) != 1 || !snapshot.Guesses[0].Correct {
		t.Fatalf("expected one correct guess, got %+v", snapshot.Guesses)
	}
}
