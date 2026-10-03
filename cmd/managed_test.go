package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const (
	ownerHelperEnv    = "GOGP_TEST_OWNER_HELPER"
	ownerHelperBinary = "GOGP_TEST_OWNER_BINARY"
	ownerHelperTarget = "GOGP_TEST_OWNER_TARGET"
)

// managedBinary is the real go-grip executable built once for the process-level
// managed contract tests.
var managedBinary string

func TestMain(m *testing.M) {
	if os.Getenv(ownerHelperEnv) != "" {
		os.Exit(m.Run())
	}

	dir, err := os.MkdirTemp("", "go-grip-managed-tests")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create test bin dir:", err)
		os.Exit(1)
	}
	bin := filepath.Join(dir, "go-grip")
	build := exec.Command("go", "build", "-o", bin, "github.com/showgp/go-grip")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "build go-grip:", err)
		_ = os.RemoveAll(dir)
		os.Exit(1)
	}
	managedBinary = bin

	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// managedEvent mirrors the v1 protocol messages asserted by the tests.
type managedEvent struct {
	Version    int    `json:"version"`
	Event      string `json:"event"`
	Generation string `json:"generation"`
	URL        string `json:"url"`
	Reload     *struct {
		State  string `json:"state"`
		Reason string `json:"reason"`
	} `json:"reload"`
	Target *struct {
		State  string `json:"state"`
		Reason string `json:"reason"`
	} `json:"target"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestManagedDirectoryPreviewServesContentAndStops(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeProcessDoc(t, filepath.Join(root, "README.md"), "# Root directory marker\n")
	writeProcessDoc(t, filepath.Join(root, "nested", "note.md"), "# Nested directory marker\n")

	child := startGoGrip(t, "--managed", "gen-dir", "-r", "--", root)
	ready := child.nextEvent(t, 10*time.Second)

	if ready.Version != 1 || ready.Event != "ready" || ready.Generation != "gen-dir" {
		t.Fatalf("unexpected ready event %+v (stderr: %s)", ready, child.stderr.String())
	}
	if ready.Reload == nil {
		t.Fatalf("ready must carry the reload snapshot: %+v", ready)
	}
	switch ready.Reload.State {
	case "pending", "active", "degraded":
	default:
		t.Fatalf("reload state = %q, want a real state", ready.Reload.State)
	}

	origin, err := url.Parse(ready.URL)
	if err != nil {
		t.Fatalf("ready URL %q: %v", ready.URL, err)
	}
	if origin.Hostname() != "127.0.0.1" {
		t.Fatalf("ready URL host = %q, want the real loopback address", origin.Hostname())
	}
	if origin.Port() == "" || origin.Port() == "6419" {
		t.Fatalf("ready URL port = %q, want the OS-assigned port", origin.Port())
	}
	base := origin.Scheme + "://" + origin.Host

	if status, body := httpGetBody(t, base+"/README.md"); status != http.StatusOK || !strings.Contains(body, "Root directory marker") {
		t.Fatalf("README preview = %d %q, want the root document", status, body)
	}
	if status, body := httpGetBody(t, base+"/nested/note.md"); status != http.StatusOK || !strings.Contains(body, "Nested directory marker") {
		t.Fatalf("nested preview = %d %q, want the nested document", status, body)
	}
	if status, body := httpGetBody(t, ready.URL); status != http.StatusOK || !strings.Contains(body, "<!doctype html>") {
		t.Fatalf("published URL = %d %q, want a rendered page", status, body)
	}

	readiness := httpRequest(t, http.MethodHead, base+"/__gogrip/ready")
	if readiness.StatusCode != http.StatusNoContent {
		t.Fatalf("readiness status = %d, want 204", readiness.StatusCode)
	}
	if got := readiness.Header.Get("X-GoGrip-Generation"); got != "gen-dir" {
		t.Fatalf("readiness generation = %q, want %q", got, "gen-dir")
	}
	_ = readiness.Body.Close()

	assertLoopbackListener(t, child.cmd.Process.Pid, origin.Port())

	child.closeStdin()
	if code := child.wait(t, 5*time.Second); code != 0 {
		t.Fatalf("owner loss exit code = %d, want 0 (stderr: %s)", code, child.stderr.String())
	}
	for _, event := range child.drainEvents(t) {
		if event.Event == "fatal" {
			t.Fatalf("owner loss produced a fatal event: %+v", event)
		}
	}
	assertPortClosed(t, origin.Host)
}

func TestManagedSingleFileWithSpacesAndNonASCII(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeProcessDoc(t, filepath.Join(dir, "说明 文档.md"), "# 单文件标记\n")
	writeProcessDoc(t, filepath.Join(dir, "other.md"), "# 其他文章标记\n")

	child := startGoGrip(t, "--managed", "gen-file", "--", filepath.Join(dir, "说明 文档.md"))
	ready := child.nextEvent(t, 10*time.Second)
	if ready.Event != "ready" {
		t.Fatalf("unexpected event %+v (stderr: %s)", ready, child.stderr.String())
	}

	origin, err := url.Parse(ready.URL)
	if err != nil {
		t.Fatalf("ready URL %q: %v", ready.URL, err)
	}
	if origin.Path != "/说明 文档.md" {
		t.Fatalf("ready URL path = %q, want the selected single file", origin.Path)
	}
	if escaped := origin.EscapedPath(); !strings.Contains(escaped, "%20") {
		t.Fatalf("ready URL path %q must escape the space", escaped)
	}

	status, body := httpGetBody(t, ready.URL)
	if status != http.StatusOK || !strings.Contains(body, "单文件标记") {
		t.Fatalf("single-file preview = %d %q, want the selected document", status, body)
	}
	if strings.Contains(body, "其他文章标记") {
		t.Fatalf("single-file preview served another article: %q", body)
	}

	child.closeStdin()
	if code := child.wait(t, 5*time.Second); code != 0 {
		t.Fatalf("owner loss exit code = %d, want 0 (stderr: %s)", code, child.stderr.String())
	}
}

func TestManagedEmptyDirectoryPublishesServingReady(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	child := startGoGrip(t, "--managed", "gen-empty", "-r", "--", root)
	ready := child.nextEvent(t, 10*time.Second)
	if ready.Event != "ready" {
		t.Fatalf("unexpected event %+v (stderr: %s)", ready, child.stderr.String())
	}

	if status, body := httpGetBody(t, ready.URL); status != http.StatusOK || !strings.Contains(body, "docs-empty") {
		t.Fatalf("empty directory preview = %d %q, want the explicit empty state page", status, body)
	}

	child.closeStdin()
	if code := child.wait(t, 5*time.Second); code != 0 {
		t.Fatalf("owner loss exit code = %d, want 0 (stderr: %s)", code, child.stderr.String())
	}
}

func TestManagedStartupFailureEmitsFatalWithoutReady(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "missing-target")
	child := startGoGrip(t, "--managed", "gen-fail", "--", missing)

	event := child.nextEvent(t, 10*time.Second)
	if event.Event != "fatal" {
		t.Fatalf("event = %+v, want fatal", event)
	}
	if event.Code != "target-unavailable" || event.Message == "" {
		t.Fatalf("fatal = %+v, want a classifiable target failure", event)
	}
	if code := child.wait(t, 5*time.Second); code == 0 {
		t.Fatalf("fatal startup exit code = 0, want non-zero (stderr: %s)", child.stderr.String())
	}
	for _, event := range child.drainEvents(t) {
		if event.Event == "ready" {
			t.Fatalf("failed startup published ready: %+v", event)
		}
	}
}

func TestManagedRejectsInapplicableInvocations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeProcessDoc(t, filepath.Join(dir, "README.md"), "# Marker\n")
	doc := filepath.Join(dir, "README.md")
	missing := filepath.Join(t.TempDir(), "missing")

	cases := []struct {
		name string
		args []string
		code string
	}{
		{name: "empty generation", args: []string{"--managed", "", "--", dir}, code: "invalid-generation"},
		{name: "json output", args: []string{"--managed", "g", "--json", "--", dir}, code: "inapplicable-flags"},
		{name: "export", args: []string{"--managed", "g", "--export", doc, "--", dir}, code: "inapplicable-flags"},
		{name: "output", args: []string{"--managed", "g", "--output", filepath.Join(dir, "out.html"), "--", dir}, code: "inapplicable-flags"},
		{name: "port override", args: []string{"--managed", "g", "--port", "1234", "--", dir}, code: "inapplicable-flags"},
		{name: "host override", args: []string{"--managed", "g", "--host", "0.0.0.0", "--", dir}, code: "inapplicable-flags"},
		// Rejection must precede target I/O: a missing target would otherwise
		// report target-unavailable instead.
		{name: "rejected before target access", args: []string{"--managed", "g", "--json", "--", missing}, code: "inapplicable-flags"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			child := startGoGrip(t, tt.args...)
			event := child.nextEvent(t, 10*time.Second)
			if event.Event != "fatal" || event.Code != tt.code {
				t.Fatalf("event = %+v, want fatal %q (stderr: %s)", event, tt.code, child.stderr.String())
			}
			if code := child.wait(t, 5*time.Second); code == 0 {
				t.Fatal("inapplicable invocation exited with success")
			}
			for _, event := range child.drainEvents(t) {
				if event.Event == "ready" {
					t.Fatalf("inapplicable invocation published ready: %+v", event)
				}
			}
		})
	}
}

func TestManagedOwnerLossBeforeReadyExitsBounded(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeProcessDoc(t, filepath.Join(root, "README.md"), "# Early loss marker\n")

	child := startGoGrip(t, "--managed", "gen-early", "-r", "--", root)
	// Close the ownership pipe immediately, without waiting for readiness.
	child.closeStdin()

	if code := child.wait(t, 6*time.Second); code != 0 {
		t.Fatalf("early owner loss exit code = %d, want 0 (stderr: %s)", code, child.stderr.String())
	}

	for _, event := range child.drainEvents(t) {
		switch event.Event {
		case "fatal":
			t.Fatalf("early owner loss produced a fatal event: %+v", event)
		case "ready":
			// The service was already up when ownership ended; it must still
			// have released its listener by the time the process exited.
			origin, err := url.Parse(event.URL)
			if err != nil {
				t.Fatalf("ready URL %q: %v", event.URL, err)
			}
			assertPortClosed(t, origin.Host)
		}
	}
}

func TestManagedExitsWhenOwnerIsForceKilled(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeProcessDoc(t, filepath.Join(root, "README.md"), "# Force-killed owner marker\n")

	helper := exec.Command(os.Args[0], "-test.run=^TestOwnerHelperProcess$")
	helper.Env = append(os.Environ(),
		ownerHelperEnv+"=1",
		ownerHelperBinary+"="+managedBinary,
		ownerHelperTarget+"="+root,
	)
	stdout, err := helper.StdoutPipe()
	if err != nil {
		t.Fatalf("helper stdout: %v", err)
	}
	stderr, err := helper.StderrPipe()
	if err != nil {
		t.Fatalf("helper stderr: %v", err)
	}
	if err := helper.Start(); err != nil {
		t.Fatalf("start owner helper: %v", err)
	}
	helperExited := make(chan struct{})
	go func() {
		_ = helper.Wait()
		close(helperExited)
	}()
	t.Cleanup(func() {
		_ = helper.Process.Kill()
		<-helperExited
	})

	pidCh := make(chan int, 1)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if rest, ok := strings.CutPrefix(scanner.Text(), "CHILD_PID="); ok {
				if pid, err := strconv.Atoi(strings.TrimSpace(rest)); err == nil {
					select {
					case pidCh <- pid:
					default:
					}
				}
			}
		}
	}()

	var childPID int
	select {
	case childPID = <-pidCh:
	case <-time.After(10 * time.Second):
		t.Fatal("owner helper did not report the managed child PID")
	}

	readyLine := readLine(t, stdout, 10*time.Second)
	var ready managedEvent
	if err := json.Unmarshal([]byte(readyLine), &ready); err != nil {
		t.Fatalf("ready line %q: %v", readyLine, err)
	}
	if ready.Event != "ready" {
		t.Fatalf("event = %+v, want ready", ready)
	}
	origin, err := url.Parse(ready.URL)
	if err != nil {
		t.Fatalf("ready URL %q: %v", ready.URL, err)
	}
	if status, body := httpGetBody(t, ready.URL); status != http.StatusOK || !strings.Contains(body, "Force-killed owner marker") {
		t.Fatalf("pre-kill preview = %d %q, want the rendered document", status, body)
	}

	// Force-terminate the owner. The kernel closes the ownership pipe; the
	// managed process must end and release its listener without the owner
	// performing any cleanup.
	if err := helper.Process.Kill(); err != nil {
		t.Fatalf("kill owner helper: %v", err)
	}
	<-helperExited

	deadline := time.Now().Add(6 * time.Second)
	for {
		pidGone := syscall.Kill(childPID, 0) == syscall.ESRCH
		portGone := !portAccepts(origin.Host)
		if pidGone && portGone {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("managed child %d (pid gone: %v) still holds %s (port gone: %v)", childPID, pidGone, origin.Host, portGone)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestManagedStopLeavesOtherSessionsServing(t *testing.T) {
	t.Parallel()

	dirA := t.TempDir()
	dirB := t.TempDir()
	dirC := t.TempDir()
	writeProcessDoc(t, filepath.Join(dirA, "README.md"), "# Session A marker\n")
	writeProcessDoc(t, filepath.Join(dirB, "README.md"), "# Session B marker\n")
	writeProcessDoc(t, filepath.Join(dirC, "README.md"), "# Standalone C marker\n")

	childA := startGoGrip(t, "--managed", "gen-a", "-r", "--", dirA)
	childB := startGoGrip(t, "--managed", "gen-b", "-r", "--", dirB)
	readyA := childA.nextEvent(t, 10*time.Second)
	readyB := childB.nextEvent(t, 10*time.Second)
	if readyA.Event != "ready" || readyB.Event != "ready" {
		t.Fatalf("unexpected ready events %+v %+v", readyA, readyB)
	}
	originA, err := url.Parse(readyA.URL)
	if err != nil {
		t.Fatalf("ready A URL: %v", err)
	}
	originB, err := url.Parse(readyB.URL)
	if err != nil {
		t.Fatalf("ready B URL: %v", err)
	}

	// A standalone CLI with --json keeps its existing startup behavior and
	// must not enter the ownership path: closing its stdin changes nothing.
	standalone := startGoGrip(t, "--json", "--port", "0", "--browser=false", "--", dirC)
	infoLine := standalone.nextLine(t, 10*time.Second)
	var info struct {
		Port int    `json:"port"`
		Host string `json:"host"`
		URL  string `json:"url"`
	}
	if err := json.Unmarshal([]byte(infoLine), &info); err != nil {
		t.Fatalf("standalone --json line %q: %v", infoLine, err)
	}
	if info.Port <= 0 || info.URL == "" {
		t.Fatalf("standalone info = %+v, want a real port and URL", info)
	}
	standalone.closeStdin()

	if status, body := httpGetBody(t, info.URL); status != http.StatusOK || !strings.Contains(body, "Standalone C marker") {
		t.Fatalf("standalone preview = %d %q, want the standalone document", status, body)
	}

	// Stopping session A must release only A's resources.
	childA.closeStdin()
	if code := childA.wait(t, 5*time.Second); code != 0 {
		t.Fatalf("session A exit code = %d, want 0 (stderr: %s)", code, childA.stderr.String())
	}
	assertPortClosed(t, originA.Host)

	if status, body := httpGetBody(t, originB.Scheme+"://"+originB.Host+"/README.md"); status != http.StatusOK || !strings.Contains(body, "Session B marker") {
		t.Fatalf("session B after A stopped = %d %q, want B still serving", status, body)
	}
	if status, body := httpGetBody(t, info.URL); status != http.StatusOK || !strings.Contains(body, "Standalone C marker") {
		t.Fatalf("standalone after A stopped = %d %q, want C still serving", status, body)
	}

	childB.closeStdin()
	if code := childB.wait(t, 5*time.Second); code != 0 {
		t.Fatalf("session B exit code = %d, want 0 (stderr: %s)", code, childB.stderr.String())
	}
}

// TestManagedEmptyRootReloadsNewDocument drives the real managed process over
// an empty directory: once the watch set is active, a Markdown file created in
// it must produce a real reload signal and be served by the same service
// without a restart.
func TestManagedEmptyRootReloadsNewDocument(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	child := startGoGrip(t, "--managed", "gen-watch", "-r", "--", root)
	ready := child.nextEvent(t, 10*time.Second)
	if ready.Event != "ready" || ready.Reload == nil {
		t.Fatalf("first event = %+v, want ready with a reload snapshot (stderr: %s)", ready, child.stderr.String())
	}
	origin, err := url.Parse(ready.URL)
	if err != nil {
		t.Fatalf("ready URL %q: %v", ready.URL, err)
	}
	base := origin.Scheme + "://" + origin.Host

	if status, body := httpGetBody(t, ready.URL); status != http.StatusOK || !strings.Contains(body, "docs-empty") {
		t.Fatalf("empty root preview = %d %q, want the empty state", status, body)
	}

	// The initial snapshot may still be pending; wait for real coverage before
	// relying on the watcher.
	if ready.Reload.State != "active" {
		child.waitEvent(t, func(e managedEvent) bool {
			return e.Event == "reload-status" && e.Reload != nil && e.Reload.State == "active"
		}, 10*time.Second)
	}

	wsURL := "ws://" + origin.Host + "/reload_ws?v=2"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", wsURL, err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	writeProcessDoc(t, filepath.Join(root, "first.md"), "# First document marker\n")

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read reload message: %v (stderr: %s)", err, child.stderr.String())
	}
	if string(message) != "reload:first.md" {
		t.Fatalf("reload message = %q, want the new document", message)
	}

	if status, body := httpGetBody(t, base+"/first.md"); status != http.StatusOK || !strings.Contains(body, "First document marker") {
		t.Fatalf("new document preview = %d %q", status, body)
	}
	if status, body := httpGetBody(t, ready.URL); status != http.StatusOK || !strings.Contains(body, "First document marker") {
		t.Fatalf("published URL after the new document = %d %q", status, body)
	}

	child.closeStdin()
	if code := child.wait(t, 5*time.Second); code != 0 {
		t.Fatalf("owner loss exit code = %d, want 0 (stderr: %s)", code, child.stderr.String())
	}
}

// TestManagedDegradedWatchStillServesAndRefreshes pins the real degraded
// service: when the watch budget leaves coverage incomplete, the ready
// snapshot (or a reload-status transition) reports degraded with a reason, and
// the accessible document is still served, including after a manual refresh
// picks up new content.
func TestManagedDegradedWatchStillServesAndRefreshes(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "darwin" {
		t.Skip("the descriptor-derived watch budget applies to kqueue platforms")
	}

	root := t.TempDir()
	doc := filepath.Join(root, "README.md")
	writeProcessDoc(t, doc, "# Degraded initial marker\n")

	child := startGoGripWithNoFileLimit(t, 128, "--managed", "gen-degraded", "-r", "--", root)
	ready := child.nextEvent(t, 10*time.Second)
	if ready.Event != "ready" || ready.Reload == nil {
		t.Fatalf("first event = %+v, want ready with a reload snapshot (stderr: %s)", ready, child.stderr.String())
	}

	degraded := ready
	if ready.Reload.State != "degraded" {
		degraded = child.waitEvent(t, func(e managedEvent) bool {
			return e.Reload != nil && e.Reload.State == "degraded"
		}, 10*time.Second)
	}
	if degraded.Reload.Reason == "" {
		t.Fatalf("degraded report %+v has no reason", degraded)
	}

	if status, body := httpGetBody(t, ready.URL); status != http.StatusOK || !strings.Contains(body, "Degraded initial marker") {
		t.Fatalf("degraded preview = %d %q, want the accessible document", status, body)
	}

	writeProcessDoc(t, doc, "# Degraded refreshed marker\n")
	if status, body := httpGetBody(t, ready.URL); status != http.StatusOK || !strings.Contains(body, "Degraded refreshed marker") {
		t.Fatalf("degraded manual refresh = %d %q, want the new content", status, body)
	}

	child.closeStdin()
	if code := child.wait(t, 5*time.Second); code != 0 {
		t.Fatalf("owner loss exit code = %d, want 0 (stderr: %s)", code, child.stderr.String())
	}
}

// TestManagedSingleFileReportsUnavailableAndRecovery pins the single-file
// target lifecycle: removing the selected file is reported unavailable with a
// reason while the HTTP error stays its own, and restoring the path reports
// available again.
func TestManagedSingleFileReportsUnavailableAndRecovery(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	doc := filepath.Join(dir, "doc.md")
	writeProcessDoc(t, doc, "# Single file marker\n")

	child := startGoGrip(t, "--managed", "gen-file-loss", "--", doc)
	ready := child.nextEvent(t, 10*time.Second)
	if ready.Event != "ready" {
		t.Fatalf("first event = %+v, want ready (stderr: %s)", ready, child.stderr.String())
	}
	if status, body := httpGetBody(t, ready.URL); status != http.StatusOK || !strings.Contains(body, "Single file marker") {
		t.Fatalf("single-file preview = %d %q", status, body)
	}

	if err := os.Remove(doc); err != nil {
		t.Fatalf("remove %s: %v", doc, err)
	}
	if status, _ := httpGetBody(t, ready.URL); status != http.StatusNotFound {
		t.Fatalf("removed selected file response = %d, want 404", status)
	}
	unavailable := child.waitEvent(t, func(e managedEvent) bool {
		return e.Event == "target-status" && e.Target != nil && e.Target.State == "unavailable"
	}, 10*time.Second)
	if unavailable.Target.Reason == "" {
		t.Fatalf("unavailable report %+v has no reason", unavailable)
	}

	writeProcessDoc(t, doc, "# Single file restored marker\n")
	if status, body := httpGetBody(t, ready.URL); status != http.StatusOK || !strings.Contains(body, "Single file restored marker") {
		t.Fatalf("restored selected file preview = %d %q", status, body)
	}
	child.waitEvent(t, func(e managedEvent) bool {
		return e.Event == "target-status" && e.Target != nil && e.Target.State == "available"
	}, 10*time.Second)

	child.closeStdin()
	if code := child.wait(t, 5*time.Second); code != 0 {
		t.Fatalf("owner loss exit code = %d, want 0 (stderr: %s)", code, child.stderr.String())
	}
}

// TestOwnerHelperProcess is the temporary owner the SIGKILL test terminates.
// It starts the real managed service, holds the ownership pipe write end for
// its whole lifetime, and reports the managed child PID to the parent.
func TestOwnerHelperProcess(t *testing.T) {
	if os.Getenv(ownerHelperEnv) != "1" {
		t.Skip("owner helper process")
	}

	bin := os.Getenv(ownerHelperBinary)
	target := os.Getenv(ownerHelperTarget)

	child := exec.Command(bin, "--managed", "gen-owner-kill", "-r", "--", target)
	stdin, err := child.StdinPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "helper stdin:", err)
		os.Exit(2)
	}
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	if err := child.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "helper start:", err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "CHILD_PID=%d\n", child.Process.Pid)

	// Hold the ownership pipe (stdin) open until this process is force-killed.
	time.Sleep(2 * time.Minute)
	runtime.KeepAlive(stdin)
}

type childProcess struct {
	t       *testing.T
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	lines   chan string
	stderr  *lockedBuffer
	exited  chan struct{}
	exitErr error
}

func startGoGrip(t *testing.T, args ...string) *childProcess {
	t.Helper()

	return startChild(t, exec.Command(managedBinary, args...))
}

// startGoGripWithNoFileLimit starts the real binary through a shell that lowers
// the soft descriptor limit first. On kqueue platforms the hot-reload watch
// budget derives from RLIMIT_NOFILE, so a low limit forces a degraded watch set
// with a real process. The shell execs the binary, so ownership and lifecycle
// semantics stay those of a directly started process.
func startGoGripWithNoFileLimit(t *testing.T, limit int, args ...string) *childProcess {
	t.Helper()

	script := fmt.Sprintf("ulimit -n %d; exec \"$0\" \"$@\"", limit)
	return startChild(t, exec.Command("/bin/sh", append([]string{"-c", script, managedBinary}, args...)...))
}

func startChild(t *testing.T, cmd *exec.Cmd) *childProcess {
	t.Helper()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderr := &lockedBuffer{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start %v: %v", cmd.Args, err)
	}

	lines := make(chan string, 256)
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()

	child := &childProcess{
		t:      t,
		cmd:    cmd,
		stdin:  stdin,
		lines:  lines,
		stderr: stderr,
		exited: make(chan struct{}),
	}
	go func() {
		<-drained
		child.exitErr = cmd.Wait()
		close(child.exited)
	}()

	t.Cleanup(func() {
		_ = stdin.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		select {
		case <-child.exited:
		case <-time.After(5 * time.Second):
			t.Errorf("child %v did not exit during cleanup", cmd.Args)
		}
	})
	return child
}

func (c *childProcess) closeStdin() {
	_ = c.stdin.Close()
}

func (c *childProcess) nextLine(t *testing.T, timeout time.Duration) string {
	t.Helper()

	select {
	case line, ok := <-c.lines:
		if !ok {
			t.Fatalf("stdout closed before the expected line (stderr: %s)", c.stderr.String())
		}
		return line
	case <-time.After(timeout):
		t.Fatalf("timed out waiting for a stdout line (stderr: %s)", c.stderr.String())
		return ""
	}
}

func (c *childProcess) nextEvent(t *testing.T, timeout time.Duration) managedEvent {
	t.Helper()

	line := c.nextLine(t, timeout)
	var event managedEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		t.Fatalf("stdout line %q is not managed protocol JSON: %v", line, err)
	}
	return event
}

// waitEvent reads events until one matches want.
func (c *childProcess) waitEvent(t *testing.T, want func(managedEvent) bool, timeout time.Duration) managedEvent {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		event := c.nextEvent(t, time.Until(deadline))
		if want(event) {
			return event
		}
	}
	t.Fatal("timed out waiting for the managed event")
	return managedEvent{}
}

func (c *childProcess) drainEvents(t *testing.T) []managedEvent {
	t.Helper()

	var events []managedEvent
	for line := range c.lines {
		var event managedEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("stdout line %q is not managed protocol JSON: %v", line, err)
		}
		events = append(events, event)
	}
	return events
}

func (c *childProcess) wait(t *testing.T, timeout time.Duration) int {
	t.Helper()

	select {
	case <-c.exited:
	case <-time.After(timeout):
		t.Fatalf("process did not exit within %s (stderr: %s)", timeout, c.stderr.String())
	}
	return c.cmd.ProcessState.ExitCode()
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func writeProcessDoc(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readLine(t *testing.T, r io.Reader, timeout time.Duration) string {
	t.Helper()

	type result struct {
		line string
		err  error
	}
	lines := make(chan result, 1)
	go func() {
		reader := bufio.NewReader(r)
		line, err := reader.ReadString('\n')
		lines <- result{line: line, err: err}
	}()

	select {
	case res := <-lines:
		if res.err != nil {
			t.Fatalf("read line: %v", res.err)
		}
		return strings.TrimSpace(res.line)
	case <-time.After(timeout):
		t.Fatal("timed out waiting for a line")
		return ""
	}
}

func httpGetBody(t *testing.T, rawURL string) (int, string) {
	t.Helper()

	resp := httpRequest(t, http.MethodGet, rawURL)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s: %v", rawURL, err)
	}
	return resp.StatusCode, string(body)
}

func httpRequest(t *testing.T, method, rawURL string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, rawURL, nil)
	if err != nil {
		t.Fatalf("new request %s %s: %v", method, rawURL, err)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, rawURL, err)
	}
	return resp
}

func portAccepts(host string) bool {
	conn, err := net.DialTimeout("tcp", host, 300*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func assertPortClosed(t *testing.T, host string) {
	t.Helper()

	if portAccepts(host) {
		t.Fatalf("expected %s to stop accepting connections", host)
	}
}

// assertLoopbackListener observes the managed process's real listening socket
// and checks it is bound to 127.0.0.1 rather than the wildcard address. Dialing
// a local non-loopback address from the same host is not usable evidence: the
// kernel delivers those connections locally regardless of the bind address.
func assertLoopbackListener(t *testing.T, pid int, port string) {
	t.Helper()

	out, err := exec.Command("lsof", "-nP", "-a", "-p", strconv.Itoa(pid), "-iTCP", "-sTCP:LISTEN").Output()
	if err != nil {
		t.Logf("lsof unavailable (%v); loopback bind evidence limited to the listener unit test", err)
		return
	}
	bound := string(out)
	want := "127.0.0.1:" + port
	if !strings.Contains(bound, want) {
		t.Fatalf("managed listener is not bound to %s:\n%s", want, bound)
	}
	if strings.Contains(bound, "*:"+port) {
		t.Fatalf("managed listener is bound to the wildcard address:\n%s", bound)
	}
}
