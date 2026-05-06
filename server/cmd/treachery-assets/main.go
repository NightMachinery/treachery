package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	fmt.Printf("Cached %d %s/%s images in %s (%s %dx%d)\n", result.ImageCount, result.CrimePackID, result.AssetPackID, result.CacheDir, result.Geometry.AspectRatio, result.Geometry.Width, result.Geometry.Height)
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
	fmt.Printf("Migrated %d %s/%s images to AVIF\n", result.ImageCount, result.CrimePackID, result.AssetPackID)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n  treachery-assets cache <asset-pack-name>\n  treachery-assets migrate <asset-pack-name> [--no-clean-git-history]\n")
	os.Exit(2)
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
