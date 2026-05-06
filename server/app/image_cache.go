package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

const (
	avifQuality = "80"
	avifSpeed   = "6"
	pipelineV1  = "v1"
)

var validTreacheryAVIFCacheBasename = regexp.MustCompile(`^[0-9a-f]{64}\.avif$`)

type assetGeometry struct {
	AspectRatio string
	Width       int
	Height      int
}

type assetImageCache struct {
	cacheDir string
	mu       sync.RWMutex
	images   map[string]assetImage
}

type assetImage struct {
	ID         string
	SourcePath string
	CachePath  string
	Geometry   assetGeometry
}

func DefaultImageCacheDir() string {
	if value := strings.TrimSpace(os.Getenv("TREACHERY_IMAGE_CACHE_DIR")); value != "" {
		return expandHome(value)
	}
	return expandHome("~/.cache/treachery/images")
}

func expandHome(path string) string {
	path = strings.TrimSpace(path)
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}

func newAssetImageCache(cacheDir string) *assetImageCache {
	cacheDir = expandHome(cacheDir)
	if strings.TrimSpace(cacheDir) == "" {
		cacheDir = DefaultImageCacheDir()
	}
	return &assetImageCache{cacheDir: cacheDir, images: map[string]assetImage{}}
}

func (c *assetImageCache) Register(sourcePath string, geometry assetGeometry) (string, error) {
	if c == nil || strings.TrimSpace(sourcePath) == "" {
		return "", nil
	}
	if !supportedCacheImagePath(sourcePath) {
		return filepath.ToSlash(sourcePath), nil
	}
	bytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("read asset image %s: %w", sourcePath, err)
	}
	id := treacheryImageID(bytes, geometry)
	image := assetImage{ID: id, SourcePath: sourcePath, CachePath: filepath.Join(c.cacheDir, id+".avif"), Geometry: geometry}
	c.mu.Lock()
	if _, exists := c.images[id]; !exists {
		c.images[id] = image
	}
	c.mu.Unlock()
	return "/api/assets/images/" + id, nil
}

func (c *assetImageCache) Get(id string) (assetImage, bool) {
	if c == nil || !validTreacheryAVIFCacheBasename.MatchString(id+".avif") {
		return assetImage{}, false
	}
	c.mu.RLock()
	img, ok := c.images[id]
	c.mu.RUnlock()
	return img, ok
}

func (c *assetImageCache) All() []assetImage {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]assetImage, 0, len(c.images))
	for _, img := range c.images {
		out = append(out, img)
	}
	return out
}

func (c *assetImageCache) EnsureAll(ctx context.Context) error {
	for _, img := range c.All() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := ensureTreacheryAVIFCache(img); err != nil {
			return err
		}
	}
	return nil
}

func (c *assetImageCache) ServeHTTP(w http.ResponseWriter, r *http.Request, imageID string) bool {
	img, ok := c.Get(imageID)
	if !ok {
		return false
	}
	if err := ensureTreacheryAVIFCache(img); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	w.Header().Set("Content-Type", "image/avif")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, img.CachePath)
	return true
}

func treacheryImageID(sourceBytes []byte, g assetGeometry) string {
	sourceSum := sha256.Sum256(sourceBytes)
	descriptor := fmt.Sprintf("source=%s|ratio=%s|width=%d|output=%dx%d|fmt=avif|backend=native|quality=%s|speed=%s|threads=auto|channels=rgb|pipeline=%s", hex.EncodeToString(sourceSum[:]), g.AspectRatio, g.Width, g.Width, g.Height, avifQuality, avifSpeed, pipelineV1)
	idSum := sha256.Sum256([]byte(descriptor))
	return hex.EncodeToString(idSum[:])
}

func ensureTreacheryAVIFCache(img assetImage) error {
	if ok, err := validCachedAVIF(img.CachePath, img.Geometry); err == nil && ok {
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(img.CachePath), 0o755); err != nil {
		return fmt.Errorf("create image cache dir: %w", err)
	}
	if strings.EqualFold(filepath.Ext(img.SourcePath), ".avif") {
		if ok, err := validCachedAVIF(img.SourcePath, img.Geometry); err != nil {
			return err
		} else if ok {
			_ = os.Remove(img.CachePath)
			if err := os.Symlink(img.SourcePath, img.CachePath); err != nil {
				return fmt.Errorf("symlink avif cache: %w", err)
			}
			return nil
		}
	}
	return buildTreacheryAVIFCache(img)
}

func validCachedAVIF(path string, g assetGeometry) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, nil
	}
	identify, err := exec.LookPath("identify")
	if err != nil {
		return false, fmt.Errorf("check avif cache: missing identify command: %w", err)
	}
	out, err := exec.Command(identify, "-format", "%wx%h", path).CombinedOutput()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) == fmt.Sprintf("%dx%d", g.Width, g.Height), nil
}

func buildTreacheryAVIFCache(img assetImage) error {
	convert, err := exec.LookPath("convert")
	if err != nil {
		return fmt.Errorf("build avif cache: missing convert command: %w", err)
	}
	avifenc, err := exec.LookPath("avifenc")
	if err != nil {
		return fmt.Errorf("build avif cache: missing avifenc command: %w", err)
	}
	tmpDir := filepath.Dir(img.CachePath)
	tmpPNG, err := os.CreateTemp(tmpDir, ".treachery-*.png")
	if err != nil {
		return fmt.Errorf("create temp png: %w", err)
	}
	tmpPNGPath := tmpPNG.Name()
	_ = tmpPNG.Close()
	defer os.Remove(tmpPNGPath)
	tmpAVIF, err := os.CreateTemp(tmpDir, ".treachery-*.avif")
	if err != nil {
		return fmt.Errorf("create temp avif: %w", err)
	}
	tmpAVIFPath := tmpAVIF.Name()
	_ = tmpAVIF.Close()
	defer os.Remove(tmpAVIFPath)
	output := fmt.Sprintf("%dx%d", img.Geometry.Width, img.Geometry.Height)
	if out, err := exec.Command(convert, img.SourcePath, "-auto-orient", "-resize", output+"^", "-gravity", "center", "-extent", output, tmpPNGPath).CombinedOutput(); err != nil {
		return fmt.Errorf("convert image to %s png: %w: %s", img.Geometry.AspectRatio, err, strings.TrimSpace(string(out)))
	}
	if out, err := exec.Command(avifenc, "-q", avifQuality, "--speed", avifSpeed, tmpPNGPath, tmpAVIFPath).CombinedOutput(); err != nil {
		return fmt.Errorf("encode avif cache: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if err := os.Rename(tmpAVIFPath, img.CachePath); err != nil {
		return fmt.Errorf("install avif cache: %w", err)
	}
	if ok, err := validCachedAVIF(img.CachePath, img.Geometry); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("built avif cache has wrong dimensions: %s", img.CachePath)
	}
	return nil
}

func supportedCacheImagePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".avif", ".svg":
		return true
	default:
		ct := mime.TypeByExtension(ext)
		return strings.HasPrefix(ct, "image/")
	}
}

type fileIdentity struct {
	dev uint64
	ino uint64
}

func walkImageFiles(path string, visitFile func(string) error) error {
	visited := map[fileIdentity]bool{}
	return walkImageFilesInner(path, visited, visitFile)
}

func walkImageFilesInner(path string, visited map[fileIdentity]bool, visitFile func(string) error) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return visitFile(path)
	}
	if identity, ok := identityForFile(info); ok {
		if visited[identity] {
			return nil
		}
		visited[identity] = true
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := walkImageFilesInner(filepath.Join(path, entry.Name()), visited, visitFile); err != nil {
			return err
		}
	}
	return nil
}

func identityForFile(info os.FileInfo) (fileIdentity, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fileIdentity{}, false
	}
	return fileIdentity{dev: uint64(stat.Dev), ino: uint64(stat.Ino)}, true
}

func sniffSupportedImage(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("open image for sniffing %s: %w", path, err)
	}
	defer f.Close()
	buf := make([]byte, 16)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return false, nil
	}
	buf = buf[:n]
	return bytes.HasPrefix(buf, []byte{0xff, 0xd8, 0xff}) || bytes.HasPrefix(buf, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) || (len(buf) >= 12 && string(buf[:4]) == "RIFF" && string(buf[8:12]) == "WEBP") || (len(buf) >= 12 && strings.Contains(string(buf[4:12]), "ftyp")), nil
}

func parseAspectRatio(value string) (int, int, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "7:10"
	}
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == ':' || r == '/' })
	if len(parts) != 2 {
		return 0, 0, "", fmt.Errorf("invalid aspect ratio %q", value)
	}
	w, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || w <= 0 {
		return 0, 0, "", fmt.Errorf("invalid aspect ratio width %q", value)
	}
	h, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || h <= 0 {
		return 0, 0, "", fmt.Errorf("invalid aspect ratio height %q", value)
	}
	return w, h, fmt.Sprintf("%d:%d", w, h), nil
}

func geometryFromAspectAndWidth(aspect string, width int) (assetGeometry, error) {
	aw, ah, normalized, err := parseAspectRatio(aspect)
	if err != nil {
		return assetGeometry{}, err
	}
	if width <= 0 {
		width = 1050
	}
	height := (width*ah + aw/2) / aw
	if height <= 0 {
		return assetGeometry{}, fmt.Errorf("invalid output height for %s @ %d", normalized, width)
	}
	return assetGeometry{AspectRatio: normalized, Width: width, Height: height}, nil
}
