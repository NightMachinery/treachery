package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"treachery/server/app"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "cache":
		cacheCmd(os.Args[2:])
	case "list", "ls":
		listCmd(os.Args[2:])
	case "migrate":
		migrateCmd(os.Args[2:])
	default:
		usage()
	}
}

func cacheCmd(args []string) {
	fs := flag.NewFlagSet("cache", flag.ExitOnError)
	wordpacksDir := fs.String("wordpacks-dir", envOrDefault("TREACHERY_WORDPACKS_DIR", "wordpacks"), "path to wordpacks directory")
	cacheDir := fs.String("cache-dir", envOrDefault("TREACHERY_IMAGE_CACHE_DIR", ""), "path to AVIF image cache directory")
	_ = fs.Parse(flagsFirst(args))
	if fs.NArg() != 1 {
		log.Fatalf("Usage: treachery-assets cache <asset-pack-name>")
	}
	result, err := app.CompleteAssetPackCache(context.Background(), *wordpacksDir, *cacheDir, fs.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	printCacheResult("Cache complete", result)
}

func listCmd(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	wordpacksDir := fs.String("wordpacks-dir", envOrDefault("TREACHERY_WORDPACKS_DIR", "wordpacks"), "path to wordpacks directory")
	_ = fs.Parse(flagsFirst(args))
	if fs.NArg() != 0 {
		log.Fatalf("Usage: treachery-assets list")
	}
	infos, err := app.ListAssetPacks(*wordpacksDir)
	if err != nil {
		log.Fatal(err)
	}
	if len(infos) == 0 {
		fmt.Printf("No asset packs found under %s.\n", *wordpacksDir)
		return
	}
	fmt.Printf("Found %d asset pack(s) under %s:\n", len(infos), *wordpacksDir)
	for _, info := range infos {
		fmt.Printf("\n- %s/%s\n", info.CrimePackID, info.AssetPackID)
		fmt.Printf("  crime pack: %s (%s)\n", info.CrimePackName, info.CrimePackID)
		fmt.Printf("  asset pack: %s (%s)\n", info.AssetPackName, info.AssetPackID)
		fmt.Printf("  geometry: %s, %dx%d\n", info.Geometry.AspectRatio, info.Geometry.Width, info.Geometry.Height)
		fmt.Printf("  images: %d\n", info.ImageCount)
		if len(info.CategoryCounts) > 0 {
			fmt.Printf("  categories: %s\n", formatCategoryCounts(info.CategoryCounts))
		}
		fmt.Printf("  asset dir: %s", info.AssetDir)
		if !info.AssetDirExists {
			fmt.Printf(" (missing)")
		}
		fmt.Printf("\n")
		flags := []string{}
		if info.AssetPackID == info.DefaultAssetPackID {
			flags = append(flags, "default")
		}
		if info.AssetPackID == info.FallbackAssetPackID {
			flags = append(flags, "fallback")
		}
		if len(flags) > 0 {
			fmt.Printf("  roles: %s\n", strings.Join(flags, ", "))
		}
	}
}

func migrateCmd(args []string) {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	wordpacksDir := fs.String("wordpacks-dir", envOrDefault("TREACHERY_WORDPACKS_DIR", "wordpacks"), "path to wordpacks directory")
	cacheDir := fs.String("cache-dir", envOrDefault("TREACHERY_IMAGE_CACHE_DIR", ""), "path to AVIF image cache directory")
	repoDir := fs.String("repo-dir", envOrDefault("TREACHERY_REPO_DIR", mustGetwd()), "path to git repository")
	noCleanHistory := fs.Bool("no-clean-git-history", false, "skip git history rewrite and push normally")
	_ = fs.Parse(flagsFirst(args))
	if fs.NArg() != 1 {
		log.Fatalf("Usage: treachery-assets migrate <asset-pack-name> [--no-clean-git-history]")
	}
	result, err := app.MigrateAssetPack(context.Background(), app.MigrateAssetPackOptions{WordpacksDir: *wordpacksDir, CacheDir: *cacheDir, RepoDir: *repoDir, AssetPackName: fs.Arg(0), NoCleanGitHistory: *noCleanHistory})
	if err != nil {
		log.Fatal(err)
	}
	printCacheResult("Migration complete", result)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n  treachery-assets list|ls\n  treachery-assets cache <asset-pack-name>\n  treachery-assets migrate <asset-pack-name> [--no-clean-git-history]\n")
	os.Exit(2)
}

func printCacheResult(title string, result *app.AssetPackCacheResult) {
	fmt.Printf("%s:\n", title)
	fmt.Printf("  crime pack: %s\n", result.CrimePackID)
	fmt.Printf("  asset pack: %s\n", result.AssetPackID)
	fmt.Printf("  cache dir: %s\n", result.CacheDir)
	fmt.Printf("  geometry: %s, %dx%d\n", result.Geometry.AspectRatio, result.Geometry.Width, result.Geometry.Height)
	fmt.Printf("  images: %d\n", result.ImageCount)
	if len(result.Images) == 0 {
		return
	}
	fmt.Printf("  cached images:\n")
	for _, img := range result.Images {
		fmt.Printf("    - source: %s\n", img.SourcePath)
		fmt.Printf("      cache: %s\n", img.CachePath)
		fmt.Printf("      id: %s\n", img.ID)
	}
}

func formatCategoryCounts(counts map[string]int) string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, counts[key]))
	}
	return strings.Join(parts, ", ")
}

func flagsFirst(args []string) []string {
	flags := []string{}
	positionals := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) > 0 && arg[0] == '-' {
			flags = append(flags, arg)
			if arg != "--no-clean-git-history" && i+1 < len(args) && len(args[i+1]) > 0 && args[i+1][0] != '-' {
				flags = append(flags, args[i+1])
				i++
			}
		} else {
			positionals = append(positionals, arg)
		}
	}
	return append(flags, positionals...)
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return filepath.Clean(".")
	}
	return wd
}
