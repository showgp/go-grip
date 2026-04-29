package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

type serverMode string

const (
	modeSingleFile serverMode = "single-file"
	modeDirectory  serverMode = "directory"
)

type serveTarget struct {
	mode        serverMode
	rootDir     string
	initialFile string
}

func resolveServeTarget(input string) (serveTarget, error) {
	if input == "" {
		rootDir, err := filepath.Abs(".")
		if err != nil {
			return serveTarget{}, err
		}
		return serveTarget{mode: modeDirectory, rootDir: rootDir}, nil
	}

	info, err := os.Stat(input)
	if err != nil {
		return serveTarget{}, fmt.Errorf("resolve target %q: %w", input, err)
	}

	if info.IsDir() {
		rootDir, err := filepath.Abs(input)
		if err != nil {
			return serveTarget{}, err
		}
		return serveTarget{mode: modeDirectory, rootDir: rootDir}, nil
	}

	rootDir, err := filepath.Abs(filepath.Dir(input))
	if err != nil {
		return serveTarget{}, err
	}
	return serveTarget{
		mode:        modeSingleFile,
		rootDir:     rootDir,
		initialFile: filepath.Base(input),
	}, nil
}
