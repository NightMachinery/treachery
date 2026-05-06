package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type AssetPackCacheResult struct {
	CrimePackID string
	AssetPackID string
	ImageCount  int
	CacheDir    string
	Geometry    assetGeometry
	Images      []assetImage
}

type AssetPackInfo struct {
	CrimePackID         string
	CrimePackName       string
	AssetPackID         string
	AssetPackName       string
	DefaultAssetPackID  string
	FallbackAssetPackID string
	ImageCount          int
	AssetDir            string
	AssetDirExists      bool
	Geometry            assetGeometry
	CategoryCounts      map[string]int
}

func ListAssetPacks(wordpacksDir string) ([]AssetPackInfo, error) {
	wordpacksDir = strings.TrimSpace(wordpacksDir)
	if wordpacksDir == "" {
		wordpacksDir = "wordpacks"
	}
	crimeRoot := filepath.Join(wordpacksDir, "crime")
	entries, err := os.ReadDir(crimeRoot)
	if err != nil {
		return nil, err
	}
	infos := []AssetPackInfo{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		packDir := filepath.Join(crimeRoot, entry.Name())
		metaPath, err := findExistingFile(filepath.Join(packDir, "pack"), []string{".json5", ".json"})
		if err != nil {
			continue
		}
		var meta crimePackMeta
		if err := decodeJSON5File(metaPath, &meta); err != nil {
			return nil, err
		}
		if meta.ID == "" {
			meta.ID = entry.Name()
		}
		if meta.Name == "" {
			meta.Name = meta.ID
		}
		for assetPackID, setMeta := range meta.AssetSets {
			if setMeta.Name == "" {
				setMeta.Name = assetPackID
			}
			geometry, err := geometryFromAspectAndWidth(setMeta.AspectRatio, setMeta.Width)
			if err != nil {
				return nil, fmt.Errorf("asset pack %s/%s geometry: %w", meta.ID, assetPackID, err)
			}
			assetDir := filepath.Join(packDir, "assets", assetPackID)
			info := AssetPackInfo{
				CrimePackID:         meta.ID,
				CrimePackName:       meta.Name,
				AssetPackID:         assetPackID,
				AssetPackName:       setMeta.Name,
				DefaultAssetPackID:  meta.DefaultAssetSetID,
				FallbackAssetPackID: meta.FallbackAssetSet,
				AssetDir:            assetDir,
				Geometry:            geometry,
				CategoryCounts:      map[string]int{},
			}
			if err := countAssetImages(assetDir, &info); err != nil {
				return nil, fmt.Errorf("count asset pack %s/%s images: %w", meta.ID, assetPackID, err)
			}
			infos = append(infos, info)
		}
	}
	sort.Slice(infos, func(i, j int) bool {
		if infos[i].CrimePackID != infos[j].CrimePackID {
			return infos[i].CrimePackID < infos[j].CrimePackID
		}
		return infos[i].AssetPackID < infos[j].AssetPackID
	})
	return infos, nil
}

func CompleteAssetPackCache(ctx context.Context, wordpacksDir, cacheDir, assetPackName string) (*AssetPackCacheResult, error) {
	packDir, packID, setMeta, err := findCrimeAssetPack(wordpacksDir, assetPackName)
	if err != nil {
		return nil, err
	}
	geometry, err := geometryFromAspectAndWidth(setMeta.AspectRatio, setMeta.Width)
	if err != nil {
		return nil, err
	}
	cache := newAssetImageCache(cacheDir)
	assetDir := filepath.Join(packDir, "assets", assetPackName)
	if err := walkImageFiles(assetDir, func(path string) error {
		if !supportedCacheImagePath(path) {
			if filepath.Ext(path) == "" {
				ok, err := sniffSupportedImage(path)
				if err != nil || !ok {
					return err
				}
			} else {
				return nil
			}
		}
		_, err := cache.Register(path, geometry)
		return err
	}); err != nil {
		return nil, err
	}
	if err := cache.EnsureAll(ctx); err != nil {
		return nil, err
	}
	images := cache.All()
	sort.Slice(images, func(i, j int) bool { return images[i].SourcePath < images[j].SourcePath })
	return &AssetPackCacheResult{CrimePackID: packID, AssetPackID: assetPackName, ImageCount: len(images), CacheDir: cache.cacheDir, Geometry: geometry, Images: images}, nil
}

func countAssetImages(assetDir string, info *AssetPackInfo) error {
	stat, err := os.Stat(assetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !stat.IsDir() {
		return nil
	}
	info.AssetDirExists = true
	return walkImageFiles(assetDir, func(path string) error {
		if !isCountableAssetImage(path) {
			return nil
		}
		info.ImageCount++
		rel, err := filepath.Rel(assetDir, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		category := strings.Split(rel, "/")[0]
		if category == "." || category == "" || !strings.Contains(rel, "/") {
			category = "(root)"
		}
		info.CategoryCounts[category]++
		return nil
	})
}

func isCountableAssetImage(path string) bool {
	if supportedCacheImagePath(path) {
		return true
	}
	if filepath.Ext(path) != "" {
		return false
	}
	ok, err := sniffSupportedImage(path)
	return err == nil && ok
}

type MigrateAssetPackOptions struct {
	WordpacksDir      string
	CacheDir          string
	RepoDir           string
	AssetPackName     string
	NoCleanGitHistory bool
}

func MigrateAssetPack(ctx context.Context, opts MigrateAssetPackOptions) (*AssetPackCacheResult, error) {
	repoDir := strings.TrimSpace(opts.RepoDir)
	if repoDir == "" {
		var err error
		repoDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	if err := requireCleanGit(repoDir); err != nil {
		return nil, err
	}
	if err := createGitMirrorBackup(repoDir); err != nil {
		return nil, err
	}
	result, err := CompleteAssetPackCache(ctx, opts.WordpacksDir, opts.CacheDir, opts.AssetPackName)
	if err != nil {
		return nil, err
	}
	originalPaths := []string{}
	for _, img := range result.Images {
		newPath := strings.TrimSuffix(img.SourcePath, filepath.Ext(img.SourcePath)) + ".avif"
		cacheBytes, err := os.ReadFile(img.CachePath)
		if err != nil {
			return nil, fmt.Errorf("read cached avif %s: %w", img.CachePath, err)
		}
		if err := os.WriteFile(newPath, cacheBytes, 0o644); err != nil {
			return nil, fmt.Errorf("write migrated avif %s: %w", newPath, err)
		}
		if filepath.Clean(newPath) != filepath.Clean(img.SourcePath) {
			if err := os.Remove(img.SourcePath); err != nil {
				return nil, fmt.Errorf("remove original asset %s: %w", img.SourcePath, err)
			}
			if rel, err := filepath.Rel(repoDir, img.SourcePath); err == nil {
				originalPaths = append(originalPaths, filepath.ToSlash(rel))
			}
		}
	}
	if err := runGit(repoDir, "add", "wordpacks"); err != nil {
		return nil, err
	}
	if err := runGit(repoDir, "commit", "-m", fmt.Sprintf("Migrate %s assets to AVIF", opts.AssetPackName)); err != nil {
		return nil, err
	}
	if !opts.NoCleanGitHistory {
		originURL, _ := gitOutput(repoDir, "remote", "get-url", "origin")
		if _, err := exec.LookPath("git-filter-repo"); err != nil {
			return nil, fmt.Errorf("git history cleanup requires git-filter-repo: %w", err)
		}
		args := []string{"filter-repo", "--force", "--invert-paths"}
		for _, path := range originalPaths {
			args = append(args, "--path", path)
		}
		if len(originalPaths) > 0 {
			if err := runGit(repoDir, args...); err != nil {
				return nil, err
			}
		}
		if strings.TrimSpace(originURL) != "" {
			if _, err := gitOutput(repoDir, "remote", "get-url", "origin"); err != nil {
				if err := runGit(repoDir, "remote", "add", "origin", strings.TrimSpace(originURL)); err != nil {
					return nil, err
				}
			}
		}
		if err := runGit(repoDir, "push", "--force-with-lease"); err != nil {
			return nil, err
		}
	} else {
		if err := runGit(repoDir, "push"); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func findCrimeAssetPack(wordpacksDir, assetPackName string) (packDir, packID string, setMeta crimeAssetSetMeta, err error) {
	wordpacksDir = strings.TrimSpace(wordpacksDir)
	if wordpacksDir == "" {
		wordpacksDir = "wordpacks"
	}
	assetPackName = strings.TrimSpace(assetPackName)
	if assetPackName == "" {
		return "", "", crimeAssetSetMeta{}, fmt.Errorf("asset pack name is required")
	}
	crimeRoot := filepath.Join(wordpacksDir, "crime")
	entries, err := os.ReadDir(crimeRoot)
	if err != nil {
		return "", "", crimeAssetSetMeta{}, err
	}
	type match struct {
		packDir string
		packID  string
		meta    crimeAssetSetMeta
	}
	matches := []match{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(crimeRoot, entry.Name())
		metaPath, err := findExistingFile(filepath.Join(dir, "pack"), []string{".json5", ".json"})
		if err != nil {
			continue
		}
		var meta crimePackMeta
		if err := decodeJSON5File(metaPath, &meta); err != nil {
			return "", "", crimeAssetSetMeta{}, err
		}
		if meta.ID == "" {
			meta.ID = entry.Name()
		}
		if set, ok := meta.AssetSets[assetPackName]; ok {
			matches = append(matches, match{packDir: dir, packID: meta.ID, meta: set})
		}
	}
	if len(matches) == 0 {
		return "", "", crimeAssetSetMeta{}, fmt.Errorf("asset pack %q not found", assetPackName)
	}
	if len(matches) > 1 {
		ids := make([]string, 0, len(matches))
		for _, m := range matches {
			ids = append(ids, m.packID)
		}
		sort.Strings(ids)
		return "", "", crimeAssetSetMeta{}, fmt.Errorf("asset pack %q is ambiguous across crime packs: %s", assetPackName, strings.Join(ids, ", "))
	}
	return matches[0].packDir, matches[0].packID, matches[0].meta, nil
}

func requireCleanGit(repoDir string) error {
	out, err := exec.Command("git", "-C", repoDir, "status", "--porcelain").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git status: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if strings.TrimSpace(string(out)) != "" {
		return fmt.Errorf("git worktree must be clean before migration")
	}
	return nil
}

func createGitMirrorBackup(repoDir string) error {
	backupRoot := expandHome("~/tmp/backups")
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		return err
	}
	backupPath := filepath.Join(backupRoot, "treachery-"+time.Now().UTC().Format("20060102T150405Z")+".git")
	out, err := exec.Command("git", "clone", "--mirror", repoDir, backupPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("create git mirror backup: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func runGit(repoDir string, args ...string) error {
	_, err := gitOutput(repoDir, args...)
	return err
}

func gitOutput(repoDir string, args ...string) (string, error) {
	cmdArgs := append([]string{"-C", repoDir}, args...)
	out, err := exec.Command("git", cmdArgs...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}
