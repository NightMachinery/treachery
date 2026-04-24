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
