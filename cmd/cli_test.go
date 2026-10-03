package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// standaloneInfo is the --json startup line of the independent CLI.
type standaloneInfo struct {
	Port int    `json:"port"`
	Host string `json:"host"`
	URL  string `json:"url"`
}

func parseStandaloneInfo(t *testing.T, line string) standaloneInfo {
	t.Helper()

	if strings.Contains(line, `"event"`) || strings.Contains(line, `"generation"`) {
		t.Fatalf("standalone startup line %q looks like a managed protocol frame", line)
	}
	var info standaloneInfo
	if err := json.Unmarshal([]byte(line), &info); err != nil {
		t.Fatalf("standalone startup line %q is not JSON: %v", line, err)
	}
	if info.Port <= 0 || info.URL == "" {
		t.Fatalf("standalone startup info = %+v, want a real port and URL", info)
	}
	parsed, err := url.Parse(info.URL)
	if err != nil {
		t.Fatalf("standalone URL %q: %v", info.URL, err)
	}
	if parsed.Port() != strconv.Itoa(info.Port) {
		t.Fatalf("standalone URL %q does not use the reported port %d", info.URL, info.Port)
	}
	return info
}

// logStandaloneListener records the actual bind of a standalone CLI listener.
// The CLI keeps its existing wildcard policy; this fails only when lsof
// positively shows a loopback-only bind, which would silently narrow that
// policy, and stays quiet when lsof is unavailable.
func logStandaloneListener(t *testing.T, pid, port int) {
	t.Helper()

	out, err := exec.Command("lsof", "-nP", "-a", "-p", strconv.Itoa(pid), "-iTCP", "-sTCP:LISTEN").Output()
	if err != nil {
		t.Logf("lsof unavailable (%v); standalone listener policy not observed here", err)
		return
	}
	bound := string(out)
	wildcard := "*:" + strconv.Itoa(port)
	dialable := "[::]:" + strconv.Itoa(port)
	loopback := "127.0.0.1:" + strconv.Itoa(port) + " (LISTEN)"
	if strings.Contains(bound, loopback) && !strings.Contains(bound, wildcard) && !strings.Contains(bound, dialable) {
		t.Fatalf("standalone CLI listener is loopback-only, which changes its existing network policy:\n%s", bound)
	}
	t.Logf("standalone listener bind observed:\n%s", bound)
}

// TestStandaloneRecursiveDirectoryPreview pins the independent directory mode
// through the real binary: recursive content is served, --json reports the
// actual port and full URL instead of managed frames, and closing the
// standalone stdin does not end the service.
func TestStandaloneRecursiveDirectoryPreview(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeProcessDoc(t, filepath.Join(root, "README.md"), "# Root CLI marker\n")
	writeProcessDoc(t, filepath.Join(root, "nested", "note.md"), "# Nested CLI marker\n")

	child := startGoGrip(t, "--json", "--browser=false", "-r", "--port", "0", "--", root)
	info := parseStandaloneInfo(t, child.nextLine(t, 10*time.Second))
	base := fmt.Sprintf("http://%s:%d", info.Host, info.Port)

	// The startup URL points at a real article of this directory (the initial
	// article order is existing behavior), and both the root and the recursive
	// document are served with their own content.
	if status, body := httpGetBody(t, info.URL); status != http.StatusOK || !strings.Contains(body, "CLI marker") {
		t.Fatalf("standalone startup URL = %d %q, want a served article", status, body)
	}
	if status, body := httpGetBody(t, base+"/README.md"); status != http.StatusOK || !strings.Contains(body, "Root CLI marker") {
		t.Fatalf("standalone root document = %d %q, want the root document", status, body)
	}
	if status, body := httpGetBody(t, base+"/nested/note.md"); status != http.StatusOK || !strings.Contains(body, "Nested CLI marker") {
		t.Fatalf("standalone nested document = %d %q, want the nested document", status, body)
	}

	logStandaloneListener(t, child.cmd.Process.Pid, info.Port)

	// The standalone CLI never consumes ownership stdin.
	child.closeStdin()
	if err := child.cmd.Process.Signal(syscall.Signal(0)); err != nil {
		t.Fatalf("standalone CLI ended after its stdin closed: %v", err)
	}
	if status, body := httpGetBody(t, base+"/README.md"); status != http.StatusOK || !strings.Contains(body, "Root CLI marker") {
		t.Fatalf("standalone preview after stdin close = %d %q, want it still serving", status, body)
	}
}

// TestStandaloneSingleFilePreviewWithSpacesAndNonASCII pins the independent
// single-file mode through the real binary, including the URL that locates a
// file whose path contains spaces and Chinese characters.
func TestStandaloneSingleFilePreviewWithSpacesAndNonASCII(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	doc := filepath.Join(dir, "说明 文档.md")
	writeProcessDoc(t, doc, "# 独立单文件标记\n")
	writeProcessDoc(t, filepath.Join(dir, "other.md"), "# 其他文章标记\n")

	child := startGoGrip(t, "--json", "--browser=false", "--port", "0", "--", doc)
	info := parseStandaloneInfo(t, child.nextLine(t, 10*time.Second))

	parsed, err := url.Parse(info.URL)
	if err != nil {
		t.Fatalf("standalone URL %q: %v", info.URL, err)
	}
	if parsed.Path != "/说明 文档.md" {
		t.Fatalf("standalone URL path = %q, want the selected file", parsed.Path)
	}
	if !strings.Contains(parsed.EscapedPath(), "%20") {
		t.Fatalf("standalone URL path %q must escape the space", parsed.EscapedPath())
	}

	status, body := httpGetBody(t, info.URL)
	if status != http.StatusOK || !strings.Contains(body, "独立单文件标记") {
		t.Fatalf("standalone single-file preview = %d %q, want the selected document", status, body)
	}
	if strings.Contains(body, "其他文章标记") {
		t.Fatalf("standalone single-file preview served another article: %q", body)
	}
}

// TestStandaloneDefaultPortFallback pins the existing default-port policy
// through the real binary. When this test can hold 6419 itself, the CLI must
// fall back; when another service already holds it, that service is left alone
// and the CLI must still serve on whatever port it obtained.
func TestStandaloneDefaultPortFallback(t *testing.T) {
	t.Parallel()

	held, holdErr := net.Listen("tcp", ":6419")
	if holdErr == nil {
		t.Cleanup(func() { _ = held.Close() })
	} else {
		t.Logf("port 6419 is held by another service and will not be stopped: %v", holdErr)
	}

	root := t.TempDir()
	writeProcessDoc(t, filepath.Join(root, "README.md"), "# Fallback marker\n")

	child := startGoGrip(t, "--json", "--browser=false", "--", root)
	info := parseStandaloneInfo(t, child.nextLine(t, 10*time.Second))

	if holdErr == nil && info.Port == 6419 {
		t.Fatalf("standalone CLI claimed the default port 6419 while this test held it")
	}
	if status, body := httpGetBody(t, info.URL); status != http.StatusOK || !strings.Contains(body, "Fallback marker") {
		t.Fatalf("standalone fallback preview = %d %q, want the document", status, body)
	}
}

// TestStandaloneExplicitPortConflictIsStrict pins the strict explicit-port
// policy through the real binary: an occupied explicit port fails the startup
// instead of silently moving to another port.
func TestStandaloneExplicitPortConflictIsStrict(t *testing.T) {
	t.Parallel()

	held, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer func() { _ = held.Close() }()
	port := held.Addr().(*net.TCPAddr).Port

	root := t.TempDir()
	writeProcessDoc(t, filepath.Join(root, "README.md"), "# Conflict marker\n")

	child := startGoGrip(t, "--json", "--browser=false", "--port", strconv.Itoa(port), "--", root)
	if code := child.wait(t, 10*time.Second); code == 0 {
		t.Fatalf("explicit port conflict exited with success (stderr: %s)", child.stderr.String())
	}
	if !strings.Contains(child.stderr.String(), strconv.Itoa(port)) {
		t.Fatalf("strict port failure does not mention port %d (stderr: %s)", port, child.stderr.String())
	}
	for line := range child.lines {
		var info standaloneInfo
		if err := json.Unmarshal([]byte(line), &info); err == nil && info.Port != 0 {
			t.Fatalf("conflicted startup still reported a server on port %d", info.Port)
		}
	}
	if !portAccepts(fmt.Sprintf("localhost:%d", port)) {
		t.Fatalf("the held port %d stopped accepting connections", port)
	}
}
