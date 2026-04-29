package app

import (
	"strings"
	"testing"
)

func findCatalogCrimePackByID(t *testing.T, catalog *WordpackCatalog, id string) CrimePackCatalogEntry {
	t.Helper()
	for _, pack := range catalog.CrimePacks {
		if pack.ID == id {
			return pack
		}
	}
	t.Fatalf("crime pack %q not found in catalog", id)
	return CrimePackCatalogEntry{}
}

func findAssetSetByID(t *testing.T, sets []PackAssetSetOption, id string) PackAssetSetOption {
	t.Helper()
	for _, set := range sets {
		if set.ID == id {
			return set
		}
	}
	t.Fatalf("asset set %q not found", id)
	return PackAssetSetOption{}
}

func findCardByID(t *testing.T, cards []Card, id string) Card {
	t.Helper()
	for _, card := range cards {
		if card.ID == id {
			return card
		}
	}
	t.Fatalf("card %q not found", id)
	return Card{}
}

func TestStorybookCelAssetSetAppearsInCatalog(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	catalog := app.GetWordpackCatalog()
	pack := findCatalogCrimePackByID(t, catalog, "treachery")

	if pack.DefaultAssetSetID != "treachery" {
		t.Fatalf("expected treachery default asset set, got %q", pack.DefaultAssetSetID)
	}
	if pack.FallbackAssetSet != "treachery" {
		t.Fatalf("expected treachery fallback asset set, got %q", pack.FallbackAssetSet)
	}

	gouache := findAssetSetByID(t, pack.AssetSets, "gouache-treachery")
	if gouache.Name != "Gouache Treachery" {
		t.Fatalf("unexpected gouache-treachery name: %q", gouache.Name)
	}
	if !gouache.HasAnyImages {
		t.Fatalf("expected gouache-treachery to report images")
	}

	storybook := findAssetSetByID(t, pack.AssetSets, "storybook-cel")
	if storybook.Name != "Storybook Cel" {
		t.Fatalf("unexpected storybook-cel name: %q", storybook.Name)
	}
	if !storybook.HasAnyImages {
		t.Fatalf("expected storybook-cel to report images")
	}
}

func TestStorybookCelResourceUsesDeckDefaultsAndFallsBackToTreachery(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	resource, err := app.GetCrimePackResource("treachery", "en", "storybook-cel")
	if err != nil {
		t.Fatalf("get crime pack resource: %v", err)
	}

	if resource.AssetSetID != "storybook-cel" {
		t.Fatalf("expected storybook-cel resource, got %q", resource.AssetSetID)
	}

	defaultMeans := findCardByID(t, resource.MeansCards, "006-bamboo-tip")
	if !strings.Contains(defaultMeans.ImgURL, "/assets/storybook-cel/means/default.png") {
		t.Fatalf("expected storybook means default art, got %q", defaultMeans.ImgURL)
	}
	if !strings.Contains(defaultMeans.AltImgURL, "/assets/treachery/means/006-bamboo-tip") {
		t.Fatalf("expected treachery fallback alt image, got %q", defaultMeans.AltImgURL)
	}

	defaultClue := findCardByID(t, resource.ClueCards, "015-briefs")
	if !strings.Contains(defaultClue.ImgURL, "/assets/storybook-cel/clues/default.png") {
		t.Fatalf("expected storybook clue default art, got %q", defaultClue.ImgURL)
	}
	if !strings.Contains(defaultClue.AltImgURL, "/assets/treachery/clues/015-briefs") {
		t.Fatalf("expected treachery fallback alt image, got %q", defaultClue.AltImgURL)
	}

	anotherClue := findCardByID(t, resource.ClueCards, "101-lock")
	if !strings.Contains(anotherClue.ImgURL, "/assets/storybook-cel/clues/default.png") {
		t.Fatalf("expected storybook clue default art, got %q", anotherClue.ImgURL)
	}
	if !strings.Contains(anotherClue.AltImgURL, "/assets/treachery/clues/101-lock") {
		t.Fatalf("expected treachery fallback alt image, got %q", anotherClue.AltImgURL)
	}
}

func TestGouacheTreacheryResourceUsesOwnDefaultsAndFallsBackToTreachery(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	resource, err := app.GetCrimePackResource("treachery", "en", "gouache-treachery")
	if err != nil {
		t.Fatalf("get crime pack resource: %v", err)
	}

	if resource.AssetSetID != "gouache-treachery" {
		t.Fatalf("expected gouache-treachery resource, got %q", resource.AssetSetID)
	}

	generatedMeans := findCardByID(t, resource.MeansCards, "001-alcohol")
	if !strings.Contains(generatedMeans.ImgURL, "/assets/gouache-treachery/means/001-alcohol.png") {
		t.Fatalf("expected gouache alcohol art, got %q", generatedMeans.ImgURL)
	}
	if !strings.Contains(generatedMeans.AltImgURL, "/assets/gouache-treachery/means/default.png") {
		t.Fatalf("expected gouache means default alt image, got %q", generatedMeans.AltImgURL)
	}

	generatedClue := findCardByID(t, resource.ClueCards, "001-air-conditioning")
	if !strings.Contains(generatedClue.ImgURL, "/assets/gouache-treachery/clues/001-air-conditioning.png") {
		t.Fatalf("expected gouache air-conditioning art, got %q", generatedClue.ImgURL)
	}
	if !strings.Contains(generatedClue.AltImgURL, "/assets/gouache-treachery/clues/default.png") {
		t.Fatalf("expected gouache clue default alt image, got %q", generatedClue.AltImgURL)
	}

	defaultClue := findCardByID(t, resource.ClueCards, "101-lock")
	if !strings.Contains(defaultClue.ImgURL, "/assets/gouache-treachery/clues/default.png") {
		t.Fatalf("expected gouache clue default art, got %q", defaultClue.ImgURL)
	}
	if !strings.Contains(defaultClue.AltImgURL, "/assets/treachery/clues/101-lock") {
		t.Fatalf("expected treachery fallback alt image, got %q", defaultClue.AltImgURL)
	}
}
