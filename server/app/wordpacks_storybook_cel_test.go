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

	if pack.DefaultAssetSetID != "gouache-treachery" {
		t.Fatalf("expected gouache-treachery default asset set, got %q", pack.DefaultAssetSetID)
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

func TestCrimePackResolveImageURLsUsesSelectedDefaultsBeforeFallbackAssetSet(t *testing.T) {
	pack := &crimePack{
		fallbackAssetSet: "fallback",
		assetSets: map[string]*crimeAssetSet{
			"selected": {
				means:        map[string]string{"generated": "/selected/means/generated.png"},
				clues:        map[string]string{},
				meansDefault: "/selected/means/default.png",
				cluesDefault: "/selected/clues/default.png",
			},
			"selected-no-default": {
				means: map[string]string{},
				clues: map[string]string{},
			},
			"fallback": {
				means:        map[string]string{"missing": "/fallback/means/missing.png"},
				clues:        map[string]string{"missing": "/fallback/clues/missing.png"},
				meansDefault: "/fallback/means/default.png",
				cluesDefault: "/fallback/clues/default.png",
			},
		},
	}

	primary, secondary := pack.resolveImageURLs("means", "generated", "selected")
	if primary != "/selected/means/generated.png" || secondary != "/selected/means/default.png" {
		t.Fatalf("expected selected means art then selected default, got primary=%q secondary=%q", primary, secondary)
	}

	primary, secondary = pack.resolveImageURLs("clues", "missing", "selected")
	if primary != "/selected/clues/default.png" || secondary != "/fallback/clues/missing.png" {
		t.Fatalf("expected selected clue default then fallback art, got primary=%q secondary=%q", primary, secondary)
	}

	primary, secondary = pack.resolveImageURLs("means", "missing", "selected-no-default")
	if primary != "/fallback/means/missing.png" || secondary != "/fallback/means/default.png" {
		t.Fatalf("expected fallback means art then fallback default, got primary=%q secondary=%q", primary, secondary)
	}
}

func TestGouacheTreacheryResourceUsesGeneratedImagesAndOwnDefaults(t *testing.T) {
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

}
