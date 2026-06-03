package internal

import (
	"encoding/json"
	"io"
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

func TestServeJSONOutput(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Hello\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	// Capture stdout and stderr
	origStdout := os.Stdout
	origStderr := os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stderr pipe: %v", err)
	}
	os.Stdout = outW
	os.Stderr = errW
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	server := NewServerWithOptions(ServerOptions{
		Host:       "127.0.0.1",
		Port:       0,
		Browser:    false,
		JSONOutput: true,
		Parser:     NewParser(),
	})

	// Serve in a goroutine since it blocks
	done := make(chan error, 1)
	go func() {
		done <- server.Serve(filepath.Join(tmpDir, "README.md"))
	}()

	// Read stdout/stderr in background goroutines
	stdoutCh := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(outR)
		stdoutCh <- data
	}()
	stderrCh := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(errR)
		stderrCh <- data
	}()

	// Close write ends to signal readers after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		outW.Close()
		errW.Close()
	}()

	// Wait for stdout with timeout
	var stdoutBytes []byte
	select {
	case stdoutBytes = <-stdoutCh:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for stdout")
	}

	// Wait for stderr with timeout
	var stderrBytes []byte
	select {
	case stderrBytes = <-stderrCh:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for stderr")
	}

	// Verify stdout contains valid JSON with "port" key
	stdoutStr := strings.TrimSpace(string(stdoutBytes))
	if stdoutStr == "" {
		t.Fatal("expected JSON output on stdout, got empty")
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdoutStr), &result); err != nil {
		t.Fatalf("expected valid JSON on stdout, got error: %v\nraw: %q", err, stdoutStr)
	}

	portVal, ok := result["port"].(float64)
	if !ok {
		t.Fatalf("expected port to be a JSON number, got type %T: %v", result["port"], result)
	}
	if int(portVal) <= 0 {
		t.Fatalf("expected port to be positive, got %v", portVal)
	}
	if _, ok := result["host"]; !ok {
		t.Fatalf("expected JSON to contain 'host' key, got: %v", result)
	}
	if _, ok := result["url"]; !ok {
		t.Fatalf("expected JSON to contain 'url' key, got: %v", result)
	}

	// Verify stderr contains status messages
	stderrStr := string(stderrBytes)
	if !strings.Contains(stderrStr, "Starting server") && !strings.Contains(stderrStr, "Auto-reload") {
		t.Fatalf("expected stderr to contain status messages, got: %q", stderrStr)
	}

	// Stop the server
	_ = server.Stop()
}

func TestStopWithoutServe(t *testing.T) {
	t.Parallel()

	server := NewServerWithOptions(ServerOptions{
		Host: "localhost",
		Port: 6419,
	})
	// Should not panic or error when Stop() called without Serve()
	err := server.Stop()
	if err != nil {
		t.Fatalf("expected no error from Stop() without Serve(), got: %v", err)
	}
}
