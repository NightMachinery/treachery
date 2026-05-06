package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type WordpackCatalog struct {
	CrimePacks []CrimePackCatalogEntry `json:"crimePacks"`
	HintPacks  []HintPackCatalogEntry  `json:"hintPacks"`
}

type PackLanguageOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PackAssetSetOption struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	HasAnyImages bool   `json:"hasAnyImages"`
	AspectRatio  string `json:"aspectRatio"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

type CrimePackCatalogEntry struct {
	ID                string               `json:"id"`
	Name              string               `json:"name"`
	DefaultLanguage   string               `json:"defaultLanguage"`
	DefaultAssetSetID string               `json:"defaultAssetSetId"`
	FallbackAssetSet  string               `json:"fallbackAssetSetId,omitempty"`
	MeansCount        int                  `json:"meansCount"`
	ClueCount         int                  `json:"clueCount"`
	HasAnyImages      bool                 `json:"hasAnyImages"`
	Languages         []PackLanguageOption `json:"languages"`
	AssetSets         []PackAssetSetOption `json:"assetSets"`
}

type HintPackCatalogEntry struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	DefaultLanguage string               `json:"defaultLanguage"`
	CauseCount      int                  `json:"causeCount"`
	LocationCount   int                  `json:"locationCount"`
	OtherCount      int                  `json:"otherCount"`
	Languages       []PackLanguageOption `json:"languages"`
}

type CrimePackResource struct {
	PackID          string               `json:"packId"`
	PackName        string               `json:"packName"`
	Language        string               `json:"language"`
	AssetSetID      string               `json:"assetSetId"`
	DefaultAssetSet string               `json:"defaultAssetSetId,omitempty"`
	HasAnyImages    bool                 `json:"hasAnyImages"`
	Languages       []PackLanguageOption `json:"languages,omitempty"`
	AssetSets       []PackAssetSetOption `json:"assetSets,omitempty"`
	ClueCards       []Card               `json:"clueCards"`
	MeansCards      []Card               `json:"meansCards"`
}

type HintPackResource struct {
	PackID        string               `json:"packId"`
	PackName      string               `json:"packName"`
	Language      string               `json:"language"`
	Languages     []PackLanguageOption `json:"languages,omitempty"`
	ForensicCards ForensicCardResource `json:"forensicCards"`
}

type crimePackMeta struct {
	ID                string                       `json:"id"`
	Name              string                       `json:"name"`
	DefaultLanguage   string                       `json:"defaultLanguage"`
	DefaultAssetSetID string                       `json:"defaultAssetSetId"`
	FallbackAssetSet  string                       `json:"fallbackAssetSetId"`
	Languages         map[string]string            `json:"languages"`
	AssetSets         map[string]crimeAssetSetMeta `json:"assetSets"`
}

type crimeAssetSetMeta struct {
	Name        string `json:"name"`
	AspectRatio string `json:"aspectRatio"`
	Width       int    `json:"width"`
}

func (m *crimeAssetSetMeta) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		m.Name = name
		return nil
	}
	var raw struct {
		Name        string `json:"name"`
		AspectRatio string `json:"aspectRatio"`
		Width       int    `json:"width"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.Name = raw.Name
	m.AspectRatio = raw.AspectRatio
	m.Width = raw.Width
	return nil
}

type hintPackMeta struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	DefaultLanguage string            `json:"defaultLanguage"`
	Languages       map[string]string `json:"languages"`
}

type crimePack struct {
	id                string
	name              string
	rootDir           string
	defaultLanguage   string
	defaultAssetSetID string
	fallbackAssetSet  string
	languages         map[string]string
	assetSetNames     map[string]string
	meansIDs          []string
	clueIDs           []string
	meansLabels       map[string]map[string]string
	clueLabels        map[string]map[string]string
	meansLabelToID    map[string]string
	clueLabelToID     map[string]string
	assetSets         map[string]*crimeAssetSet
}

type crimeAssetSet struct {
	id           string
	name         string
	geometry     assetGeometry
	means        map[string]string
	clues        map[string]string
	meansDefault string
	cluesDefault string
}

type hintPack struct {
	id              string
	name            string
	rootDir         string
	defaultLanguage string
	languages       map[string]string
	data            map[string]*hintPackLanguage
	legacyCardKeys  map[string]string
}

type hintPackLanguage struct {
	causeCards    []hintCardDefinition
	locationCards []hintCardDefinition
	otherCards    []hintCardDefinition
	cardByID      map[string]hintCardDefinition
}

type hintPackLanguageFile struct {
	CauseCards    []hintCardDefinition `json:"causeCards"`
	LocationCards []hintCardDefinition `json:"locationCards"`
	OtherCards    []hintCardDefinition `json:"otherCards"`
}

type hintCardDefinition struct {
	CardID   string                 `json:"cardId"`
	CardName string                 `json:"cardName"`
	Choices  []hintChoiceDefinition `json:"choices"`
}

type hintChoiceDefinition struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

func loadWordpacks(rootDir string, imageCache *assetImageCache) (map[string]*crimePack, map[string]*hintPack, *WordpackCatalog, error) {
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		return nil, nil, nil, fmt.Errorf("wordpacks dir is required")
	}
	crimeRoot := filepath.Join(rootDir, "crime")
	hintRoot := filepath.Join(rootDir, "hint")
	crimePacks, err := loadCrimePacks(crimeRoot, imageCache)
	if err != nil {
		return nil, nil, nil, err
	}
	hintPacks, err := loadHintPacks(hintRoot)
	if err != nil {
		return nil, nil, nil, err
	}
	catalog := &WordpackCatalog{
		CrimePacks: make([]CrimePackCatalogEntry, 0, len(crimePacks)),
		HintPacks:  make([]HintPackCatalogEntry, 0, len(hintPacks)),
	}
	for _, pack := range crimePacks {
		catalog.CrimePacks = append(catalog.CrimePacks, pack.catalogEntry())
	}
	for _, pack := range hintPacks {
		catalog.HintPacks = append(catalog.HintPacks, pack.catalogEntry())
	}
	sort.Slice(catalog.CrimePacks, func(i, j int) bool { return catalog.CrimePacks[i].ID < catalog.CrimePacks[j].ID })
	sort.Slice(catalog.HintPacks, func(i, j int) bool { return catalog.HintPacks[i].ID < catalog.HintPacks[j].ID })
	return crimePacks, hintPacks, catalog, nil
}

func loadCrimePacks(rootDir string, imageCache *assetImageCache) (map[string]*crimePack, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]*crimePack{}, nil
		}
		return nil, err
	}
	packs := map[string]*crimePack{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		packDir := filepath.Join(rootDir, entry.Name())
		pack, err := loadCrimePack(packDir, imageCache)
		if err != nil {
			return nil, fmt.Errorf("load crime pack %s: %w", entry.Name(), err)
		}
		packs[pack.id] = pack
	}
	return packs, nil
}

func loadCrimePack(dir string, imageCache *assetImageCache) (*crimePack, error) {
	metaPath, err := findExistingFile(filepath.Join(dir, "pack"), []string{".json5", ".json"})
	if err != nil {
		return nil, err
	}
	var meta crimePackMeta
	if err := decodeJSON5File(metaPath, &meta); err != nil {
		return nil, err
	}
	if meta.ID == "" {
		meta.ID = filepath.Base(dir)
	}
	if meta.Name == "" {
		meta.Name = meta.ID
	}
	if meta.DefaultLanguage == "" {
		return nil, fmt.Errorf("defaultLanguage is required")
	}
	if len(meta.Languages) == 0 || meta.Languages[meta.DefaultLanguage] == "" {
		return nil, fmt.Errorf("default language %q must exist in languages", meta.DefaultLanguage)
	}
	meansIDs, err := loadOrderedIDs(filepath.Join(dir, "means", "ids.txt"))
	if err != nil {
		return nil, fmt.Errorf("load means ids: %w", err)
	}
	clueIDs, err := loadOrderedIDs(filepath.Join(dir, "clues", "ids.txt"))
	if err != nil {
		return nil, fmt.Errorf("load clue ids: %w", err)
	}
	assetSetNames := map[string]string{}
	for id, setMeta := range meta.AssetSets {
		assetSetNames[id] = setMeta.Name
	}
	pack := &crimePack{
		id:                meta.ID,
		name:              meta.Name,
		rootDir:           dir,
		defaultLanguage:   meta.DefaultLanguage,
		defaultAssetSetID: meta.DefaultAssetSetID,
		fallbackAssetSet:  meta.FallbackAssetSet,
		languages:         meta.Languages,
		assetSetNames:     assetSetNames,
		meansIDs:          meansIDs,
		clueIDs:           clueIDs,
		meansLabels:       map[string]map[string]string{},
		clueLabels:        map[string]map[string]string{},
		meansLabelToID:    map[string]string{},
		clueLabelToID:     map[string]string{},
		assetSets:         map[string]*crimeAssetSet{},
	}
	for lang := range meta.Languages {
		meansLabels, err := loadOrderedLabels(filepath.Join(dir, "means", "languages"), lang)
		if err != nil {
			return nil, fmt.Errorf("load means language %s: %w", lang, err)
		}
		if len(meansLabels) != len(meansIDs) {
			return nil, fmt.Errorf("means language %s line count mismatch", lang)
		}
		pack.meansLabels[lang] = map[string]string{}
		for idx, id := range meansIDs {
			pack.meansLabels[lang][id] = meansLabels[idx]
			if lang == meta.DefaultLanguage {
				pack.meansLabelToID[strings.TrimSpace(meansLabels[idx])] = id
			}
		}

		clueLabels, err := loadOrderedLabels(filepath.Join(dir, "clues", "languages"), lang)
		if err != nil {
			return nil, fmt.Errorf("load clue language %s: %w", lang, err)
		}
		if len(clueLabels) != len(clueIDs) {
			return nil, fmt.Errorf("clue language %s line count mismatch", lang)
		}
		pack.clueLabels[lang] = map[string]string{}
		for idx, id := range clueIDs {
			pack.clueLabels[lang][id] = clueLabels[idx]
			if lang == meta.DefaultLanguage {
				pack.clueLabelToID[strings.TrimSpace(clueLabels[idx])] = id
			}
		}
	}
	if pack.defaultAssetSetID == "" && len(meta.AssetSets) > 0 {
		for id := range meta.AssetSets {
			pack.defaultAssetSetID = id
			break
		}
	}
	for assetSetID, setMeta := range meta.AssetSets {
		set, err := loadCrimeAssetSet(dir, assetSetID, setMeta, imageCache)
		if err != nil {
			return nil, fmt.Errorf("load asset set %s: %w", assetSetID, err)
		}
		pack.assetSets[assetSetID] = set
	}
	if pack.defaultAssetSetID != "" {
		if _, ok := pack.assetSets[pack.defaultAssetSetID]; !ok {
			return nil, fmt.Errorf("default asset set %q does not exist", pack.defaultAssetSetID)
		}
	}
	if pack.fallbackAssetSet != "" {
		if _, ok := pack.assetSets[pack.fallbackAssetSet]; !ok {
			return nil, fmt.Errorf("fallback asset set %q does not exist", pack.fallbackAssetSet)
		}
	}
	return pack, nil
}

func loadCrimeAssetSet(packDir, assetSetID string, meta crimeAssetSetMeta, imageCache *assetImageCache) (*crimeAssetSet, error) {
	geometry, err := geometryFromAspectAndWidth(meta.AspectRatio, meta.Width)
	if err != nil {
		return nil, err
	}
	name := meta.Name
	if name == "" {
		name = assetSetID
	}
	set := &crimeAssetSet{
		id:       assetSetID,
		name:     name,
		geometry: geometry,
		means:    map[string]string{},
		clues:    map[string]string{},
	}
	meansDir := filepath.Join(packDir, "assets", assetSetID, "means")
	if err := loadAssetDir(meansDir, filepath.ToSlash(filepath.Join("/wordpacks", "crime", filepath.Base(packDir), "assets", assetSetID, "means")), set.means, &set.meansDefault, imageCache, geometry); err != nil {
		return nil, err
	}
	cluesDir := filepath.Join(packDir, "assets", assetSetID, "clues")
	if err := loadAssetDir(cluesDir, filepath.ToSlash(filepath.Join("/wordpacks", "crime", filepath.Base(packDir), "assets", assetSetID, "clues")), set.clues, &set.cluesDefault, imageCache, geometry); err != nil {
		return nil, err
	}
	return set, nil
}

func loadAssetDir(dir, urlBase string, cards map[string]string, defaultURL *string, imageCache *assetImageCache, geometry assetGeometry) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		stem := strings.TrimSuffix(name, filepath.Ext(name))
		url := filepath.ToSlash(filepath.Join(urlBase, name))
		if imageCache != nil && supportedCacheImagePath(filepath.Join(dir, name)) {
			registeredURL, err := imageCache.Register(filepath.Join(dir, name), geometry)
			if err != nil {
				return err
			}
			url = registeredURL
		}
		if stem == "default" {
			*defaultURL = url
			continue
		}
		cards[stem] = url
	}
	return nil
}

func loadHintPacks(rootDir string) (map[string]*hintPack, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]*hintPack{}, nil
		}
		return nil, err
	}
	packs := map[string]*hintPack{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		packDir := filepath.Join(rootDir, entry.Name())
		pack, err := loadHintPack(packDir)
		if err != nil {
			return nil, fmt.Errorf("load hint pack %s: %w", entry.Name(), err)
		}
		packs[pack.id] = pack
	}
	return packs, nil
}

func loadHintPack(dir string) (*hintPack, error) {
	metaPath, err := findExistingFile(filepath.Join(dir, "pack"), []string{".json5", ".json"})
	if err != nil {
		return nil, err
	}
	var meta hintPackMeta
	if err := decodeJSON5File(metaPath, &meta); err != nil {
		return nil, err
	}
	if meta.ID == "" {
		meta.ID = filepath.Base(dir)
	}
	if meta.Name == "" {
		meta.Name = meta.ID
	}
	if meta.DefaultLanguage == "" {
		return nil, fmt.Errorf("defaultLanguage is required")
	}
	if len(meta.Languages) == 0 || meta.Languages[meta.DefaultLanguage] == "" {
		return nil, fmt.Errorf("default language %q must exist in languages", meta.DefaultLanguage)
	}
	pack := &hintPack{
		id:              meta.ID,
		name:            meta.Name,
		rootDir:         dir,
		defaultLanguage: meta.DefaultLanguage,
		languages:       meta.Languages,
		data:            map[string]*hintPackLanguage{},
		legacyCardKeys:  map[string]string{},
	}
	var defaultData *hintPackLanguage
	for lang := range meta.Languages {
		path, err := findExistingFile(filepath.Join(dir, "languages", lang), []string{".json5", ".json"})
		if err != nil {
			return nil, err
		}
		var file hintPackLanguageFile
		if err := decodeJSON5File(path, &file); err != nil {
			return nil, err
		}
		data, err := buildHintPackLanguage(file)
		if err != nil {
			return nil, fmt.Errorf("language %s: %w", lang, err)
		}
		if lang == meta.DefaultLanguage {
			defaultData = data
		} else if err := validateHintLanguageShape(defaultData, data); err != nil {
			return nil, fmt.Errorf("language %s shape mismatch: %w", lang, err)
		}
		pack.data[lang] = data
	}
	if defaultData != nil {
		for _, card := range append(append([]hintCardDefinition{}, defaultData.causeCards...), append(defaultData.locationCards, defaultData.otherCards...)...) {
			pack.legacyCardKeys[legacyHintCardKey(card.CardName, choiceLabels(card.Choices))] = card.CardID
		}
	}
	return pack, nil
}

func buildHintPackLanguage(file hintPackLanguageFile) (*hintPackLanguage, error) {
	lang := &hintPackLanguage{
		causeCards:    file.CauseCards,
		locationCards: file.LocationCards,
		otherCards:    file.OtherCards,
		cardByID:      map[string]hintCardDefinition{},
	}
	seen := map[string]struct{}{}
	for _, card := range append(append([]hintCardDefinition{}, lang.causeCards...), append(lang.locationCards, lang.otherCards...)...) {
		if card.CardID == "" {
			return nil, fmt.Errorf("cardId is required")
		}
		if _, ok := seen[card.CardID]; ok {
			return nil, fmt.Errorf("duplicate cardId %q", card.CardID)
		}
		seen[card.CardID] = struct{}{}
		choiceSeen := map[string]struct{}{}
		for _, choice := range card.Choices {
			if choice.ID == "" {
				return nil, fmt.Errorf("card %s has blank choice id", card.CardID)
			}
			if _, ok := choiceSeen[choice.ID]; ok {
				return nil, fmt.Errorf("card %s has duplicate choice id %q", card.CardID, choice.ID)
			}
			choiceSeen[choice.ID] = struct{}{}
		}
		lang.cardByID[card.CardID] = card
	}
	return lang, nil
}

func validateHintLanguageShape(def, candidate *hintPackLanguage) error {
	if def == nil || candidate == nil {
		return nil
	}
	if err := validateHintCardSlice(def.causeCards, candidate.causeCards, "causeCards"); err != nil {
		return err
	}
	if err := validateHintCardSlice(def.locationCards, candidate.locationCards, "locationCards"); err != nil {
		return err
	}
	return validateHintCardSlice(def.otherCards, candidate.otherCards, "otherCards")
}

func validateHintCardSlice(def, candidate []hintCardDefinition, label string) error {
	if len(def) != len(candidate) {
		return fmt.Errorf("%s length mismatch", label)
	}
	for idx := range def {
		if def[idx].CardID != candidate[idx].CardID {
			return fmt.Errorf("%s[%d] cardId mismatch", label, idx)
		}
		if len(def[idx].Choices) != len(candidate[idx].Choices) {
			return fmt.Errorf("%s[%d] choice count mismatch", label, idx)
		}
		for choiceIdx := range def[idx].Choices {
			if def[idx].Choices[choiceIdx].ID != candidate[idx].Choices[choiceIdx].ID {
				return fmt.Errorf("%s[%d] choice id mismatch", label, idx)
			}
		}
	}
	return nil
}

func choiceLabels(choices []hintChoiceDefinition) []string {
	result := make([]string, 0, len(choices))
	for _, choice := range choices {
		result = append(result, choice.Label)
	}
	return result
}

func decodeJSON5File(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	normalized, err := normalizeJSON5ToJSON(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(normalized, target)
}

var (
	json5BareKeyPattern = regexp.MustCompile(`([\{,]\s*)([A-Za-z_][A-Za-z0-9_-]*)(\s*:)`)
	json5TrailingComma  = regexp.MustCompile(`,(\s*[}\]])`)
)

func normalizeJSON5ToJSON(input []byte) ([]byte, error) {
	stripped, err := stripJSON5CommentsAndNormalizeStrings(input)
	if err != nil {
		return nil, err
	}
	stripped = json5BareKeyPattern.ReplaceAll(stripped, []byte(`${1}"${2}"${3}`))
	for {
		next := json5TrailingComma.ReplaceAll(stripped, []byte(`${1}`))
		if bytes.Equal(next, stripped) {
			break
		}
		stripped = next
	}
	return stripped, nil
}

func stripJSON5CommentsAndNormalizeStrings(input []byte) ([]byte, error) {
	var out bytes.Buffer
	inString := false
	stringQuote := byte(0)
	escaped := false
	lineComment := false
	blockComment := false
	for i := 0; i < len(input); i++ {
		ch := input[i]
		if lineComment {
			if ch == '\n' {
				lineComment = false
				out.WriteByte(ch)
			}
			continue
		}
		if blockComment {
			if ch == '*' && i+1 < len(input) && input[i+1] == '/' {
				blockComment = false
				i++
			}
			continue
		}
		if inString {
			if escaped {
				escaped = false
				if stringQuote == '\'' && ch == '"' {
					out.WriteString(`\"`)
				} else {
					out.WriteByte(ch)
				}
				continue
			}
			if ch == '\\' {
				escaped = true
				out.WriteByte(ch)
				continue
			}
			if ch == stringQuote {
				inString = false
				out.WriteByte('"')
				continue
			}
			if stringQuote == '\'' && ch == '"' {
				out.WriteString(`\"`)
				continue
			}
			out.WriteByte(ch)
			continue
		}
		if ch == '/' && i+1 < len(input) {
			next := input[i+1]
			if next == '/' {
				lineComment = true
				i++
				continue
			}
			if next == '*' {
				blockComment = true
				i++
				continue
			}
		}
		if ch == '\'' || ch == '"' {
			inString = true
			stringQuote = ch
			out.WriteByte('"')
			continue
		}
		out.WriteByte(ch)
	}
	if inString || blockComment {
		return nil, fmt.Errorf("invalid json5 input")
	}
	return out.Bytes(), nil
}

func findExistingFile(base string, exts []string) (string, error) {
	for _, ext := range exts {
		candidate := base + ext
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	return "", fmt.Errorf("missing file for %s", base)
}

func loadOrderedIDs(path string) ([]string, error) {
	return loadOrderedTextFile(path)
}

func loadOrderedLabels(dir, lang string) ([]string, error) {
	if path, err := findExistingFile(filepath.Join(dir, lang), []string{".txt"}); err == nil {
		return loadOrderedTextFile(path)
	}
	path, err := findExistingFile(filepath.Join(dir, lang), []string{".jsonl"})
	if err != nil {
		return nil, err
	}
	return loadOrderedJSONLLines(path)
}

func loadOrderedTextFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result = append(result, strings.TrimRight(scanner.Text(), "\r"))
	}
	return result, scanner.Err()
}

func loadOrderedJSONLLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var value any
		if err := json.Unmarshal([]byte(line), &value); err != nil {
			return nil, err
		}
		switch typed := value.(type) {
		case string:
			result = append(result, typed)
		case map[string]any:
			for _, key := range []string{"label", "name", "value"} {
				if raw, ok := typed[key]; ok {
					if str, ok := raw.(string); ok {
						result = append(result, str)
						goto nextLine
					}
				}
			}
			return nil, fmt.Errorf("jsonl objects must contain a string label/name/value")
		default:
			return nil, fmt.Errorf("unsupported jsonl row type %T", typed)
		}
	nextLine:
	}
	return result, scanner.Err()
}

func (p *crimePack) catalogEntry() CrimePackCatalogEntry {
	entry := CrimePackCatalogEntry{
		ID:                p.id,
		Name:              p.name,
		DefaultLanguage:   p.defaultLanguage,
		DefaultAssetSetID: p.defaultAssetSetID,
		FallbackAssetSet:  p.fallbackAssetSet,
		MeansCount:        len(p.meansIDs),
		ClueCount:         len(p.clueIDs),
		HasAnyImages:      p.hasAnyImages(),
		Languages:         p.languageOptions(),
		AssetSets:         p.assetSetOptions(),
	}
	return entry
}

func (p *hintPack) catalogEntry() HintPackCatalogEntry {
	defaultData := p.mustLanguage(p.defaultLanguage)
	return HintPackCatalogEntry{
		ID:              p.id,
		Name:            p.name,
		DefaultLanguage: p.defaultLanguage,
		CauseCount:      len(defaultData.causeCards),
		LocationCount:   len(defaultData.locationCards),
		OtherCount:      len(defaultData.otherCards),
		Languages:       p.languageOptions(),
	}
}

func (p *crimePack) languageOptions() []PackLanguageOption {
	options := make([]PackLanguageOption, 0, len(p.languages))
	for id, name := range p.languages {
		options = append(options, PackLanguageOption{ID: id, Name: name})
	}
	sort.Slice(options, func(i, j int) bool { return options[i].ID < options[j].ID })
	return options
}

func (p *hintPack) languageOptions() []PackLanguageOption {
	options := make([]PackLanguageOption, 0, len(p.languages))
	for id, name := range p.languages {
		options = append(options, PackLanguageOption{ID: id, Name: name})
	}
	sort.Slice(options, func(i, j int) bool { return options[i].ID < options[j].ID })
	return options
}

func (p *crimePack) assetSetOptions() []PackAssetSetOption {
	options := make([]PackAssetSetOption, 0, len(p.assetSets))
	for id, set := range p.assetSets {
		options = append(options, PackAssetSetOption{ID: id, Name: set.name, HasAnyImages: set.hasAnyImages(), AspectRatio: set.geometry.AspectRatio, Width: set.geometry.Width, Height: set.geometry.Height})
	}
	sort.Slice(options, func(i, j int) bool { return options[i].ID < options[j].ID })
	return options
}

func (s *crimeAssetSet) hasAnyImages() bool {
	return len(s.means) > 0 || len(s.clues) > 0 || s.meansDefault != "" || s.cluesDefault != ""
}

func (p *crimePack) hasAnyImages() bool {
	for _, set := range p.assetSets {
		if set.hasAnyImages() {
			return true
		}
	}
	return false
}

func (p *crimePack) normalizedLanguage(lang string) (string, error) {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		lang = p.defaultLanguage
	}
	if _, ok := p.languages[lang]; !ok {
		return "", fmt.Errorf("%w: unknown crime pack language %q", ErrBadInput, lang)
	}
	return lang, nil
}

func (p *crimePack) normalizedAssetSet(assetSetID string) (string, error) {
	assetSetID = strings.TrimSpace(assetSetID)
	if assetSetID == "" {
		assetSetID = p.defaultAssetSetID
	}
	if assetSetID == "" {
		return "", nil
	}
	if _, ok := p.assetSets[assetSetID]; !ok {
		return "", fmt.Errorf("%w: unknown crime pack asset set %q", ErrBadInput, assetSetID)
	}
	return assetSetID, nil
}

func (p *hintPack) normalizedLanguage(lang string) (string, error) {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		lang = p.defaultLanguage
	}
	if _, ok := p.languages[lang]; !ok {
		return "", fmt.Errorf("%w: unknown hint pack language %q", ErrBadInput, lang)
	}
	return lang, nil
}

func (p *crimePack) mustLabels(deck, lang string) map[string]string {
	lang = p.mustLanguage(lang)
	switch deck {
	case "means":
		return p.meansLabels[lang]
	default:
		return p.clueLabels[lang]
	}
}

func (p *crimePack) mustLanguage(lang string) string {
	if _, ok := p.languages[lang]; ok {
		return lang
	}
	return p.defaultLanguage
}

func (p *crimePack) resolveCard(deck, lang, assetSetID string, card Card, forceTextOnly bool) Card {
	resolved := card
	if set := p.assetSets[assetSetID]; set != nil {
		resolved.AspectRatio = set.geometry.AspectRatio
		resolved.Width = set.geometry.Width
		resolved.Height = set.geometry.Height
	}
	resolved.ID = p.normalizeCardID(deck, card)
	labels := p.mustLabels(deck, lang)
	if label, ok := labels[resolved.ID]; ok {
		resolved.Name = label
	}
	if forceTextOnly {
		resolved.ImgURL = ""
		resolved.AltImgURL = ""
		resolved.HasImage = false
		return resolved
	}
	primary, secondary := p.resolveImageURLs(deck, resolved.ID, assetSetID)
	resolved.ImgURL = primary
	resolved.AltImgURL = secondary
	resolved.HasImage = primary != ""
	return resolved
}

func (p *crimePack) resolveCards(deck, lang, assetSetID string, ids []string) []Card {
	cards := make([]Card, 0, len(ids))
	forceTextOnly := !p.hasAnyImages()
	for _, id := range ids {
		cards = append(cards, p.resolveCard(deck, lang, assetSetID, Card{ID: id}, forceTextOnly))
	}
	return cards
}

func (p *crimePack) resolveImageURLs(deck, cardID, assetSetID string) (string, string) {
	candidates := make([]string, 0, 4)
	addCandidate := func(url string) {
		if url == "" {
			return
		}
		for _, existing := range candidates {
			if existing == url {
				return
			}
		}
		candidates = append(candidates, url)
	}
	selected := p.assetSets[assetSetID]
	fallback := p.assetSets[p.fallbackAssetSet]
	addSetCandidates := func(set *crimeAssetSet) {
		if set == nil {
			return
		}
		switch deck {
		case "means":
			addCandidate(set.means[cardID])
			addCandidate(set.meansDefault)
		default:
			addCandidate(set.clues[cardID])
			addCandidate(set.cluesDefault)
		}
	}
	addSetCandidates(selected)
	if fallback != selected {
		addSetCandidates(fallback)
	}
	primary := ""
	secondary := ""
	if len(candidates) > 0 {
		primary = candidates[0]
	}
	if len(candidates) > 1 {
		secondary = candidates[1]
	}
	return primary, secondary
}

func (p *crimePack) normalizeCardID(deck string, card Card) string {
	if card.ID != "" {
		return card.ID
	}
	name := strings.TrimSpace(card.Name)
	if name == "" {
		return ""
	}
	switch deck {
	case "means":
		return p.meansLabelToID[name]
	default:
		return p.clueLabelToID[name]
	}
}

func (p *crimePack) cardsResource(lang, assetSetID string) (*CrimePackResource, error) {
	lang, err := p.normalizedLanguage(lang)
	if err != nil {
		return nil, err
	}
	assetSetID, err = p.normalizedAssetSet(assetSetID)
	if err != nil {
		return nil, err
	}
	forceTextOnly := !p.hasAnyImages()
	means := make([]Card, 0, len(p.meansIDs))
	for _, id := range p.meansIDs {
		means = append(means, p.resolveCard("means", lang, assetSetID, Card{ID: id}, forceTextOnly))
	}
	clues := make([]Card, 0, len(p.clueIDs))
	for _, id := range p.clueIDs {
		clues = append(clues, p.resolveCard("clues", lang, assetSetID, Card{ID: id}, forceTextOnly))
	}
	return &CrimePackResource{
		PackID:          p.id,
		PackName:        p.name,
		Language:        lang,
		AssetSetID:      assetSetID,
		DefaultAssetSet: p.defaultAssetSetID,
		HasAnyImages:    p.hasAnyImages(),
		Languages:       p.languageOptions(),
		AssetSets:       p.assetSetOptions(),
		MeansCards:      means,
		ClueCards:       clues,
	}, nil
}

func (p *hintPack) mustLanguage(lang string) *hintPackLanguage {
	if data, ok := p.data[lang]; ok {
		return data
	}
	return p.data[p.defaultLanguage]
}

func (p *hintPack) cardsResource(lang string) (*HintPackResource, error) {
	lang, err := p.normalizedLanguage(lang)
	if err != nil {
		return nil, err
	}
	data := p.mustLanguage(lang)
	return &HintPackResource{
		PackID:    p.id,
		PackName:  p.name,
		Language:  lang,
		Languages: p.languageOptions(),
		ForensicCards: ForensicCardResource{
			CauseCards:    p.resolveCards(data.causeCards),
			LocationCards: p.resolveCards(data.locationCards),
			OtherCards:    p.resolveCards(data.otherCards),
		},
	}, nil
}

func (p *hintPack) resolveCards(defs []hintCardDefinition) []ForensicCard {
	cards := make([]ForensicCard, 0, len(defs))
	for _, def := range defs {
		card := ForensicCard{CardID: def.CardID, CardName: def.CardName, Replaced: false}
		for _, choice := range def.Choices {
			card.ChoiceIDs = append(card.ChoiceIDs, choice.ID)
			card.Choices = append(card.Choices, choice.Label)
		}
		cards = append(cards, card)
	}
	return cards
}

func (p *hintPack) resolveCard(card ForensicCard, lang string) ForensicCard {
	resolved := card
	data := p.mustLanguage(lang)
	cardID := p.normalizeCardID(card)
	if def, ok := data.cardByID[cardID]; ok {
		resolved.CardID = def.CardID
		resolved.CardName = def.CardName
		resolved.Choices = resolved.Choices[:0]
		resolved.ChoiceIDs = resolved.ChoiceIDs[:0]
		for _, choice := range def.Choices {
			resolved.ChoiceIDs = append(resolved.ChoiceIDs, choice.ID)
			resolved.Choices = append(resolved.Choices, choice.Label)
		}
		if resolved.SelectedChoiceID == "" && resolved.SelectedChoice != "" {
			resolved.SelectedChoiceID = p.choiceIDFromLabel(def, resolved.SelectedChoice)
		}
		resolved.SelectedChoice = p.choiceLabelFromID(def, resolved.SelectedChoiceID)
	}
	return resolved
}

func (p *hintPack) normalizeCardID(card ForensicCard) string {
	if card.CardID != "" {
		return card.CardID
	}
	if card.CardName == "" {
		return ""
	}
	return p.legacyCardKeys[legacyHintCardKey(card.CardName, card.Choices)]
}

func legacyHintCardKey(name string, choices []string) string {
	return strings.TrimSpace(name) + "|" + strings.Join(choices, "\x1f")
}

func (p *hintPack) choiceIDFromLabel(def hintCardDefinition, label string) string {
	for _, choice := range def.Choices {
		if choice.Label == label {
			return choice.ID
		}
	}
	return ""
}

func (p *hintPack) choiceLabelFromID(def hintCardDefinition, id string) string {
	for _, choice := range def.Choices {
		if choice.ID == id {
			return choice.Label
		}
	}
	return ""
}

func (a *App) resolveGameContent(game *Game, players []Player, guesses []Guess) {
	if game == nil {
		return
	}
	crimePack := a.getCrimePack(game.CrimePackID)
	if crimePack != nil {
		language, err := crimePack.normalizedLanguage(game.CrimePackLanguage)
		if err != nil {
			language = crimePack.defaultLanguage
		}
		assetSetID, err := crimePack.normalizedAssetSet(game.CrimePackAssetSetID)
		if err != nil {
			assetSetID = crimePack.defaultAssetSetID
		}
		game.CrimePackID = crimePack.id
		game.CrimePackLanguage = language
		game.CrimePackAssetSetID = assetSetID
		forceTextOnly := game.MeansCluesTextOnly || !crimePack.hasAnyImages()
		if !crimePack.hasAnyImages() {
			game.MeansCluesTextOnly = true
		}
		for playerIdx := range players {
			for clueIdx := range players[playerIdx].ClueCards {
				players[playerIdx].ClueCards[clueIdx] = crimePack.resolveCard("clues", language, assetSetID, players[playerIdx].ClueCards[clueIdx], forceTextOnly)
			}
			for meansIdx := range players[playerIdx].MeansCards {
				players[playerIdx].MeansCards[meansIdx] = crimePack.resolveCard("means", language, assetSetID, players[playerIdx].MeansCards[meansIdx], forceTextOnly)
			}
		}
		game.MurdererClueCardID = crimePack.normalizeCardID("clues", Card{ID: game.MurdererClueCardID, Name: game.MurdererClueCardName})
		if game.MurdererClueCardID != "" {
			game.MurdererClueCardName = crimePack.mustLabels("clues", language)[game.MurdererClueCardID]
		}
		game.MurdererMeansCardID = crimePack.normalizeCardID("means", Card{ID: game.MurdererMeansCardID, Name: game.MurdererMeansCardName})
		if game.MurdererMeansCardID != "" {
			game.MurdererMeansCardName = crimePack.mustLabels("means", language)[game.MurdererMeansCardID]
		}
		for idx := range guesses {
			guesses[idx].ClueCardID = crimePack.normalizeCardID("clues", Card{ID: guesses[idx].ClueCardID, Name: guesses[idx].ClueCardName})
			if guesses[idx].ClueCardID != "" {
				guesses[idx].ClueCardName = crimePack.mustLabels("clues", language)[guesses[idx].ClueCardID]
			}
			guesses[idx].MeansCardID = crimePack.normalizeCardID("means", Card{ID: guesses[idx].MeansCardID, Name: guesses[idx].MeansCardName})
			if guesses[idx].MeansCardID != "" {
				guesses[idx].MeansCardName = crimePack.mustLabels("means", language)[guesses[idx].MeansCardID]
			}
		}
	}
	hintPack := a.getHintPack(game.HintPackID)
	if hintPack != nil {
		lang, err := hintPack.normalizedLanguage(game.HintPackLanguage)
		if err != nil {
			lang = hintPack.defaultLanguage
		}
		game.HintPackID = hintPack.id
		game.HintPackLanguage = lang
		if game.CauseCard != nil {
			resolved := hintPack.resolveCard(*game.CauseCard, lang)
			game.CauseCard = &resolved
		}
		if game.LocationCard != nil {
			resolved := hintPack.resolveCard(*game.LocationCard, lang)
			game.LocationCard = &resolved
		}
		for idx := range game.OtherCards {
			game.OtherCards[idx] = hintPack.resolveCard(game.OtherCards[idx], lang)
		}
	}
}

func (a *App) getCrimePack(id string) *crimePack {
	if len(a.crimePacks) == 0 {
		return nil
	}
	if pack, ok := a.crimePacks[id]; ok {
		return pack
	}
	if pack, ok := a.crimePacks[defaultCrimePackID]; ok {
		return pack
	}
	for _, pack := range a.crimePacks {
		return pack
	}
	return nil
}

func (a *App) getHintPack(id string) *hintPack {
	if len(a.hintPacks) == 0 {
		return nil
	}
	if pack, ok := a.hintPacks[id]; ok {
		return pack
	}
	if pack, ok := a.hintPacks[defaultHintPackID]; ok {
		return pack
	}
	for _, pack := range a.hintPacks {
		return pack
	}
	return nil
}

func (a *App) GetWordpackCatalog() *WordpackCatalog {
	return a.wordpackCatalog
}

func (a *App) GetCrimePackResource(packID, language, assetSetID string) (*CrimePackResource, error) {
	pack := a.getCrimePack(packID)
	if pack == nil {
		return nil, fmt.Errorf("%w: unknown crime pack %q", ErrNotFound, packID)
	}
	return pack.cardsResource(language, assetSetID)
}

func (a *App) GetHintPackResource(packID, language string) (*HintPackResource, error) {
	pack := a.getHintPack(packID)
	if pack == nil {
		return nil, fmt.Errorf("%w: unknown hint pack %q", ErrNotFound, packID)
	}
	return pack.cardsResource(language)
}

func (a *App) serveWordpackAsset(w http.ResponseWriter, r *http.Request) {
	cleanPath := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/wordpacks/"))
	if cleanPath == "." || strings.HasPrefix(cleanPath, "..") {
		http.NotFound(w, r)
		return
	}
	if a.wordpacksDir == "" {
		http.NotFound(w, r)
		return
	}
	candidate := filepath.Join(a.wordpacksDir, cleanPath)
	rel, err := filepath.Rel(a.wordpacksDir, candidate)
	if err != nil || strings.HasPrefix(rel, "..") {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(candidate)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, candidate)
}
