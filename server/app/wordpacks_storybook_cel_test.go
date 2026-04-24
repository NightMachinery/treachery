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

func TestStorybookCelResourceUsesPilotArtAndFallsBackToTreachery(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()

	resource, err := app.GetCrimePackResource("treachery", "en", "storybook-cel")
	if err != nil {
		t.Fatalf("get crime pack resource: %v", err)
	}

	if resource.AssetSetID != "storybook-cel" {
		t.Fatalf("expected storybook-cel resource, got %q", resource.AssetSetID)
	}

	pilotMeans := findCardByID(t, resource.MeansCards, "006-bamboo-tip")
	if !strings.Contains(pilotMeans.ImgURL, "/assets/storybook-cel/means/006-bamboo-tip.png") {
		t.Fatalf("expected pilot means art, got %q", pilotMeans.ImgURL)
	}
	if !strings.Contains(pilotMeans.AltImgURL, "/assets/treachery/means/006-bamboo-tip") {
		t.Fatalf("expected treachery fallback alt image, got %q", pilotMeans.AltImgURL)
	}

	pilotClue := findCardByID(t, resource.ClueCards, "015-briefs")
	if !strings.Contains(pilotClue.ImgURL, "/assets/storybook-cel/clues/015-briefs.png") {
		t.Fatalf("expected pilot clue art, got %q", pilotClue.ImgURL)
	}
	if !strings.Contains(pilotClue.AltImgURL, "/assets/treachery/clues/015-briefs") {
		t.Fatalf("expected treachery fallback alt image, got %q", pilotClue.AltImgURL)
	}

	fallbackClue := findCardByID(t, resource.ClueCards, "101-lock")
	if strings.Contains(fallbackClue.ImgURL, "/assets/storybook-cel/") {
		t.Fatalf("expected fallback clue art, got storybook image %q", fallbackClue.ImgURL)
	}
	if !strings.Contains(fallbackClue.ImgURL, "/assets/treachery/clues/101-lock") {
		t.Fatalf("expected treachery fallback clue image, got %q", fallbackClue.ImgURL)
	}
}
