package internal

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDirectoryRootRedirectsToInitialArticle(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Hello\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandler(http.Dir(tmpDir))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-Modified-Since", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); !strings.Contains(got, "no-store") {
		t.Fatalf("expected Cache-Control to disable storage, got %q", got)
	}
	if got := recorder.Header().Get("Location"); got != "/README.md" {
		t.Fatalf("expected redirect to README.md, got %q", got)
	}
}

func TestRecursiveDirectoryRootRedirectsToNestedInitialArticle(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("mkdir nested dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "README.md"), []byte("# Nested\n"), 0o644); err != nil {
		t.Fatalf("write nested README.md: %v", err)
	}

	server := NewServerWithOptions(ServerOptions{
		Host:      "localhost",
		Port:      6419,
		Recursive: true,
		Parser:    NewParser(),
	})
	handler := server.newHandlerForTarget(serveTarget{
		mode:    modeDirectory,
		rootDir: tmpDir,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, recorder.Code)
	}
	if got := recorder.Header().Get("Location"); got != "/docs/README.md" {
		t.Fatalf("expected redirect to nested README.md, got %q", got)
	}
}

func TestDirectoryMarkdownResponseIncludesPreviousAndNextArticleNavigation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	files := map[string]string{
		"README.md": "# Readme\n",
		"guide.md":  "# Guide\n",
		"zeta.md":   "# Zeta\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandlerForTarget(serveTarget{
		mode:    modeDirectory,
		rootDir: tmpDir,
	})

	req := httptest.NewRequest(http.MethodGet, "/guide.md", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		`data-prev-article="/README.md"`,
		`data-next-article="/zeta.md"`,
		`docs-page-nav-prev" href="/README.md"`,
		`docs-page-nav-next" href="/zeta.md"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected body to contain %q, got %q", want, body)
		}
	}
}

func TestDirectorySidebarTitleUsesRootDirectoryName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	rootDir := filepath.Join(tmpDir, "reference")
	if err := os.Mkdir(rootDir, 0o755); err != nil {
		t.Fatalf("mkdir root dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "README.md"), []byte("# Hello\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandlerForTarget(serveTarget{
		mode:    modeDirectory,
		rootDir: rootDir,
	})

	req := httptest.NewRequest(http.MethodGet, "/README.md", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `class="docs-sidebar-title">reference</div>`) {
		t.Fatalf("expected sidebar title to use root directory name, got %q", recorder.Body.String())
	}
}

func TestRecursiveDirectoryMarkdownResponseIncludesNestedSidebar(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("mkdir nested dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "中文 guide.md"), []byte("# Nested Guide\n"), 0o644); err != nil {
		t.Fatalf("write nested guide: %v", err)
	}

	server := NewServerWithOptions(ServerOptions{
		Host:      "localhost",
		Port:      6419,
		Recursive: true,
		Parser:    NewParser(),
	})
	handler := server.newHandlerForTarget(serveTarget{
		mode:    modeDirectory,
		rootDir: tmpDir,
	})

	req := httptest.NewRequest(http.MethodGet, "/docs/%E4%B8%AD%E6%96%87%20guide.md", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		`<ul class="docs-sidebar-list">`,
		`<details class="docs-sidebar-details" open>`,
		`class="docs-sidebar-folder"`,
		`class="docs-sidebar-folder-icon"`,
		`class="docs-sidebar-label">docs</span>`,
		`class="docs-sidebar-chevron"`,
		`href="/docs/%E4%B8%AD%E6%96%87%20guide.md"`,
		`aria-current="page"`,
		`Nested Guide`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected body to contain %q, got %q", want, body)
		}
	}
	if strings.Contains(body, `<ol class="docs-sidebar-list">`) {
		t.Fatalf("expected sidebar article tree to avoid ordered lists, got %q", body)
	}
}

func TestRecursiveDirectorySidebarDefaultsToCollapsed(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("mkdir nested dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "doc.md"), []byte("# Single\nContent\n"), 0o644); err != nil {
		t.Fatalf("write doc.md: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandlerForTarget(serveTarget{
		mode:        modeSingleFile,
		rootDir:     tmpDir,
		initialFile: "doc.md",
	})

	req := httptest.NewRequest(http.MethodGet, "/export?file=doc.md", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "Single") {
		t.Fatalf("expected body to contain rendered content, got %q", body)
	}
	if cd := recorder.Header().Get("Content-Disposition"); !strings.Contains(cd, `attachment; filename="doc.html"`) {
		t.Fatalf("expected Content-Disposition with doc.html, got %q", cd)
	}
}

func TestExportRouteMissingFileParam(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandler(http.Dir(tmpDir))

	req := httptest.NewRequest(http.MethodGet, "/export", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestExportRouteNonexistentFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandler(http.Dir(tmpDir))

	req := httptest.NewRequest(http.MethodGet, "/export?file=nonexistent.md", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestExportRouteDirectoryTraversal(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandler(http.Dir(tmpDir))

	req := httptest.NewRequest(http.MethodGet, "/export?file=../etc/passwd", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for directory traversal attempt, got %d", http.StatusBadRequest, recorder.Code)
	}
}
