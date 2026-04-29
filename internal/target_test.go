package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveServeTargetDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	target, err := resolveServeTarget(tmpDir)
	if err != nil {
		t.Fatalf("resolve target: %v", err)
	}

	if target.mode != modeDirectory {
		t.Fatalf("expected directory mode, got %q", target.mode)
	}
	wantRoot, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("abs tmp dir: %v", err)
	}
	if target.rootDir != wantRoot {
		t.Fatalf("expected root %q, got %q", wantRoot, target.rootDir)
	}
	if target.initialFile != "" {
		t.Fatalf("expected no initial file, got %q", target.initialFile)
	}
}

func TestResolveServeTargetSingleFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(file, []byte("# Hello\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	target, err := resolveServeTarget(file)
	if err != nil {
		t.Fatalf("resolve target: %v", err)
	}

	if target.mode != modeSingleFile {
		t.Fatalf("expected single-file mode, got %q", target.mode)
	}
	if target.rootDir != tmpDir {
		t.Fatalf("expected root %q, got %q", tmpDir, target.rootDir)
	}
	if target.initialFile != "README.md" {
		t.Fatalf("expected README.md, got %q", target.initialFile)
	}
}
