package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"treachery/server/app"
)

func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("getwd: %v", err)
	}

	addr := flag.String("addr", envOrDefault("TREACHERY_ADDR", "127.0.0.1:18083"), "listen address")
	distDir := flag.String("dist-dir", envOrDefault("TREACHERY_DIST_DIR", filepath.Join(rootDir, "UI", "dist", "deceptiongame")), "path to built UI dist directory")
	dataDir := flag.String("data-dir", envOrDefault("TREACHERY_DATA_DIR", filepath.Join(rootDir, ".self_host", "data")), "path to runtime data directory")
	wordpacksDir := flag.String("wordpacks-dir", envOrDefault("TREACHERY_WORDPACKS_DIR", filepath.Join(rootDir, "wordpacks")), "path to wordpacks directory")
	imageCacheDir := flag.String("image-cache-dir", envOrDefault("TREACHERY_IMAGE_CACHE_DIR", app.DefaultImageCacheDir()), "path to AVIF image cache directory")
	flag.Parse()

	application, err := app.New(app.Config{
		DistDir:       *distDir,
		DataDir:       *dataDir,
		WordpacksDir:  *wordpacksDir,
		ImageCacheDir: *imageCacheDir,
	})
	if err != nil {
		log.Fatalf("create app: %v", err)
	}
	defer application.Close()

	server := &http.Server{
		Addr:    *addr,
		Handler: application.Handler(),
	}

	log.Printf("treachery server listening on %s", *addr)
	log.Printf("serving dist dir %s", *distDir)
	log.Printf("using data dir %s", *dataDir)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
