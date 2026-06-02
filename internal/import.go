package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var importLocks sync.Map

func isAllowedImageExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".tiff", ".tif", ".avif":
		return true
	}
	return false
}

func isAllowedImageMIME(mime string) bool {
	switch strings.ToLower(mime) {
	case "image/jpeg", "image/png", "image/gif", "image/webp",
		"image/svg+xml", "image/bmp", "image/tiff", "image/avif":
		return true
	}
	return false
}

func writeImageWithDedup(imagesDir string, filename string, data []byte) (string, error) {
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return "", err
	}

	cleanName := filepath.Base(filename)
	if cleanName == "" || cleanName == "." {
		cleanName = "image.png"
	}

	targetPath := filepath.Join(imagesDir, cleanName)
	if info, err := os.Stat(targetPath); err == nil {
		if info.Size() == int64(len(data)) {
			return cleanName, nil
		}
		cleanName = uniqueFilename(imagesDir, cleanName)
		targetPath = filepath.Join(imagesDir, cleanName)
	}

	if err := writeImageToDir(targetPath, data); err != nil {
		return "", err
	}
	return cleanName, nil
}

func uniqueFilename(dir string, filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	for i := 1; i < 1000; i++ {
		candidate := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}
	return fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext)
}

func writeImageToDir(absPath string, data []byte) error {
	muRaw, _ := importLocks.LoadOrStore(absPath, &sync.Mutex{})
	mu := muRaw.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	tmpFile := absPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	return os.Rename(tmpFile, absPath)
}
