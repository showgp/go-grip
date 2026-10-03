package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestManagedWriterSerializesConcurrentMessages pins the single serialized
// output path: concurrent emitters must produce one complete JSON frame per
// line, never interleaved fragments.
func TestManagedWriterSerializesConcurrentMessages(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	writer := newManagedWriter(&buf, "gen-1")

	var wg sync.WaitGroup
	const emitters = 24
	for i := range emitters {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := writer.ready(fmt.Sprintf("http://127.0.0.1:%d/", 10000+i), nil); err != nil {
				t.Errorf("ready %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != emitters {
		t.Fatalf("got %d protocol lines, want %d", len(lines), emitters)
	}
	for _, line := range lines {
		var event managedEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("protocol line %q is not one complete JSON frame: %v", line, err)
		}
		if event.Version != managedProtocolVersion || event.Event != "ready" || event.Generation != "gen-1" {
			t.Fatalf("unexpected event %+v", event)
		}
	}
}

// TestOwnerLossForcesExitWhenStartupIsBlocked pins that the watchdog is armed
// by owner loss alone: it must force exit while startup work is still
// unfinished instead of waiting for the startup function to return.
func TestOwnerLossForcesExitWhenStartupIsBlocked(t *testing.T) {
	t.Parallel()

	owner, ownerWriter := io.Pipe()
	var stdout bytes.Buffer
	forced := make(chan int, 1)

	rt := &managedRuntime{
		opts:          ManagedOptions{Generation: "gen-blocked"},
		stdin:         owner,
		stdout:        &stdout,
		cleanDeadline: 100 * time.Millisecond,
		forceExit:     func(code int) { forced <- code },
	}

	started := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		finished <- rt.run(func(context.Context, *managedWriter) error {
			close(started)
			<-release // startup work that cannot observe cancellation, like blocked kernel I/O
			return nil
		})
	}()

	<-started
	if err := ownerWriter.Close(); err != nil {
		t.Fatalf("close owner writer: %v", err)
	}

	select {
	case code := <-forced:
		if code == 0 {
			t.Fatal("watchdog reported success instead of a forced exit")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("watchdog did not force exit while startup was blocked")
	}

	close(release)
	select {
	case err := <-finished:
		if err != nil {
			t.Fatalf("run returned %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not return after the blocked startup was released")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("owner loss wrote %q to the protocol stream, want nothing", got)
	}
}

// TestOwnerLossRunsBoundedCleanupWithoutForceExit pins the graceful path: owner
// loss cancels startup, and cleanup finishing inside the deadline exits without
// the watchdog.
func TestOwnerLossRunsBoundedCleanupWithoutForceExit(t *testing.T) {
	t.Parallel()

	owner, ownerWriter := io.Pipe()
	forced := make(chan int, 1)

	rt := &managedRuntime{
		opts:          ManagedOptions{Generation: "gen-graceful"},
		stdin:         owner,
		stdout:        io.Discard,
		cleanDeadline: 5 * time.Second,
		forceExit:     func(code int) { forced <- code },
	}

	canceled := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		finished <- rt.run(func(ctx context.Context, _ *managedWriter) error {
			<-ctx.Done()
			close(canceled)
			return nil
		})
	}()

	if err := ownerWriter.Close(); err != nil {
		t.Fatalf("close owner writer: %v", err)
	}

	select {
	case <-canceled:
	case <-time.After(2 * time.Second):
		t.Fatal("owner loss did not cancel startup")
	}
	select {
	case err := <-finished:
		if err != nil {
			t.Fatalf("run returned %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not return after graceful cleanup")
	}
	select {
	case code := <-forced:
		t.Fatalf("graceful cleanup was force-exited with code %d", code)
	default:
	}
}

// TestManagedReadinessRouteAnswersWithoutTargetAccess pins the startup
// confirmation route: HEAD returns 204 and the generation header without
// touching the target handler.
func TestManagedReadinessRouteAnswersWithoutTargetAccess(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the preview handler must not run for the readiness route")
	})
	handler := newManagedHandler("gen-7", next)

	req := httptest.NewRequest(http.MethodHead, managedReadyPath, nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("readiness status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get(managedGenerationHeader); got != "gen-7" {
		t.Fatalf("generation header = %q, want %q", got, "gen-7")
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("readiness body = %q, want empty", recorder.Body.String())
	}
}

// TestManagedServePublishesReadyAndReleasesPortOnOwnerLoss drives the real
// managed startup: loopback listener with an OS-assigned port, ready with the
// full URL and a real reload snapshot, HTTP content, and bounded shutdown on
// owner loss.
//
// Not parallel: constructing the reloader reads os.Stderr, which the legacy
// TestServeJSONOutput replaces globally while running in parallel.
func TestManagedServePublishesReadyAndReleasesPortOnOwnerLoss(t *testing.T) {
	root := t.TempDir()
	writeManagedDoc(t, filepath.Join(root, "README.md"), "# Root managed marker\n")
	writeManagedDoc(t, filepath.Join(root, "nested", "note.md"), "# Nested managed marker\n")

	owner, ownerWriter := io.Pipe()
	stdout, stdoutWriter := io.Pipe()
	finished := make(chan error, 1)
	exited := make(chan struct{})
	rt := &managedRuntime{
		opts: ManagedOptions{
			Generation:   "gen-serve",
			Target:       root,
			Recursive:    true,
			EnableReload: true,
		},
		stdin:         owner,
		stdout:        stdoutWriter,
		cleanDeadline: 2 * time.Second,
		forceExit:     func(code int) { t.Errorf("unexpected forced exit %d", code) },
	}
	go func() {
		finished <- rt.run(rt.serve)
		close(exited)
	}()
	t.Cleanup(func() {
		_ = ownerWriter.Close()
		_ = stdoutWriter.Close()
		select {
		case <-exited:
		case <-time.After(5 * time.Second):
			t.Errorf("managed run did not return during cleanup")
		}
	})

	line := readManagedLine(t, bufio.NewReader(stdout), 10*time.Second)
	var ready managedEvent
	if err := json.Unmarshal([]byte(line), &ready); err != nil {
		t.Fatalf("ready line %q is not JSON: %v", line, err)
	}
	if ready.Event != "ready" || ready.Version != managedProtocolVersion || ready.Generation != "gen-serve" {
		t.Fatalf("unexpected ready event %+v", ready)
	}
	if ready.Reload == nil {
		t.Fatal("ready must carry the current reload snapshot")
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
		t.Fatalf("ready URL host = %q, want the loopback address", origin.Hostname())
	}
	if origin.Port() == "" || origin.Port() == "6419" {
		t.Fatalf("ready URL port = %q, want the OS-assigned port", origin.Port())
	}

	if status, body := managedGet(t, ready.URL); status != http.StatusOK || !strings.Contains(body, "<!doctype html>") {
		t.Fatalf("published preview URL = %d %q, want a rendered page", status, body)
	}
	if status, body := managedGet(t, "http://"+origin.Host+"/README.md"); status != http.StatusOK || !strings.Contains(body, "Root managed marker") {
		t.Fatalf("root document preview = %d %q, want the rendered README", status, body)
	}
	if status, body := managedGet(t, "http://"+origin.Host+"/nested/note.md"); status != http.StatusOK || !strings.Contains(body, "Nested managed marker") {
		t.Fatalf("nested preview = %d %q, want the nested document", status, body)
	}

	readiness := managedRequest(t, http.MethodHead, "http://"+origin.Host+managedReadyPath)
	if readiness.StatusCode != http.StatusNoContent {
		t.Fatalf("readiness status = %d, want 204", readiness.StatusCode)
	}
	if got := readiness.Header.Get(managedGenerationHeader); got != "gen-serve" {
		t.Fatalf("readiness generation = %q, want %q", got, "gen-serve")
	}
	_ = readiness.Body.Close()

	if err := ownerWriter.Close(); err != nil {
		t.Fatalf("close owner writer: %v", err)
	}
	select {
	case err := <-finished:
		if err != nil {
			t.Fatalf("managed run returned %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("managed run did not return after owner loss")
	}

	if conn, err := net.DialTimeout("tcp", origin.Host, 500*time.Millisecond); err == nil {
		_ = conn.Close()
		t.Fatal("the preview listener still accepts connections after owner loss cleanup")
	}
}

// TestManagedServeDoesNotPublishReadyWhenOwnerIsGone pins the startup window:
// with the owner already lost, no ready may be published for a preview the
// host cannot own.
func TestManagedServeDoesNotPublishReadyWhenOwnerIsGone(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeManagedDoc(t, filepath.Join(root, "README.md"), "# Gone owner marker\n")

	var stdout bytes.Buffer
	writer := newManagedWriter(&stdout, "gen-gone")
	rt := &managedRuntime{
		opts:          ManagedOptions{Generation: "gen-gone", Target: root, Recursive: true, EnableReload: true},
		stdin:         strings.NewReader(""),
		stdout:        &stdout,
		cleanDeadline: time.Second,
		forceExit:     func(int) {},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := rt.serve(ctx, writer); err != nil {
		t.Fatalf("canceled startup returned %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("canceled startup published %q, want nothing", stdout.String())
	}
}

// TestManagedServeRejectsUnreadableSingleFile pins that managed startup
// verifies the selected file can actually be served: a present but unreadable
// Markdown file must fail with a fatal event instead of publishing ready.
func TestManagedServeRejectsUnreadableSingleFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	path := filepath.Join(dir, "secret.md")
	writeManagedDoc(t, path, "# Secret marker\n")
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("chmod %s: %v", path, err)
	}
	if file, err := os.Open(path); err == nil {
		_ = file.Close()
		t.Skip("file permissions are not enforced for this user")
	}

	var stdout bytes.Buffer
	rt := &managedRuntime{
		opts:          ManagedOptions{Generation: "gen-secret", Target: path},
		stdin:         strings.NewReader(""),
		stdout:        &stdout,
		cleanDeadline: time.Second,
		forceExit:     func(int) {},
	}

	if err := rt.serve(context.Background(), newManagedWriter(&stdout, "gen-secret")); err == nil {
		t.Fatal("expected unreadable target to fail startup")
	}

	var event managedEvent
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &event); err != nil {
		t.Fatalf("fatal line %q is not JSON: %v", stdout.String(), err)
	}
	if event.Event != "fatal" || event.Code != "target-unavailable" {
		t.Fatalf("event = %+v, want fatal target-unavailable", event)
	}
}

// TestManagedServeRejectsNonRegularSingleFile pins that a `.md`-named FIFO or
// device is not accepted as a preview target: ready must not advertise a URL
// whose request would block or fail.
func TestManagedServeRejectsNonRegularSingleFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "pipe.md")
	if err := exec.Command("mkfifo", path).Run(); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}

	var stdout bytes.Buffer
	rt := &managedRuntime{
		opts:          ManagedOptions{Generation: "gen-pipe", Target: path},
		stdin:         strings.NewReader(""),
		stdout:        &stdout,
		cleanDeadline: time.Second,
		forceExit:     func(int) {},
	}

	if err := rt.serve(context.Background(), newManagedWriter(&stdout, "gen-pipe")); err == nil {
		t.Fatal("expected non-regular target to fail startup")
	}

	var event managedEvent
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &event); err != nil {
		t.Fatalf("fatal line %q is not JSON: %v", stdout.String(), err)
	}
	if event.Event != "fatal" || event.Code != "target-unavailable" {
		t.Fatalf("event = %+v, want fatal target-unavailable", event)
	}
}

// TestOwnerLivenessSuppressesPublicationAfterLoss pins the ordering primitive:
// ownership loss observed before publication must suppress the ready write,
// while a publication that wins the race still runs.
func TestOwnerLivenessSuppressesPublicationAfterLoss(t *testing.T) {
	t.Parallel()

	liveness := &ownerLiveness{}
	called := false
	published, err := liveness.publish(func() error { called = true; return nil })
	if !published || !called || err != nil {
		t.Fatalf("publication before loss = (%v, %v, called=%v), want the write to run", published, err, called)
	}

	liveness.markGone()
	called = false
	published, err = liveness.publish(func() error { called = true; return nil })
	if published || called || err != nil {
		t.Fatalf("publication after loss = (%v, %v, called=%v), want it suppressed", published, err, called)
	}
}

// TestManagedServeDoesNotPublishReadyWhenOwnershipEnded pins the startup
// window: with ownership already ended, no ready may be published for a
// preview the host cannot own.
func TestManagedServeDoesNotPublishReadyWhenOwnershipEnded(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeManagedDoc(t, filepath.Join(root, "README.md"), "# Ended ownership marker\n")

	liveness := &ownerLiveness{}
	liveness.markGone()

	var stdout bytes.Buffer
	rt := &managedRuntime{
		opts:          ManagedOptions{Generation: "gen-ended", Target: root, Recursive: true, EnableReload: true},
		owner:         liveness,
		stdin:         strings.NewReader(""),
		stdout:        &stdout,
		cleanDeadline: time.Second,
		forceExit:     func(int) {},
	}

	if err := rt.serve(context.Background(), newManagedWriter(&stdout, "gen-ended")); err != nil {
		t.Fatalf("serve returned %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("ownership ended before publication but ready was written: %q", stdout.String())
	}
}

// TestCloseInitializedPDFReleasesInitializedGenerator pins that a stop really
// releases an initialized PDF generator, observed through the allocator
// context cancellation.
func TestCloseInitializedPDFReleasesInitializedGenerator(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	server := NewServerWithOptions(ServerOptions{})
	server.pdfGenMu.Lock()
	server.pdfGen = &PDFGenerator{allocCtx: ctx, allocCancel: cancel}
	server.pdfGenMu.Unlock()

	server.closeInitializedPDF()

	select {
	case <-ctx.Done():
	default:
		t.Fatal("initialized PDF generator was not released on stop")
	}
}

// TestCloseInitializedPDFDoesNotInitializeGenerator pins that stopping must not
// start the headless browser pool: no PDF request was served, so the
// constructor must not run.
//
// Not parallel: it swaps the constructor seam.
func TestCloseInitializedPDFDoesNotInitializeGenerator(t *testing.T) {
	calls := 0
	original := newPDFGenerator
	newPDFGenerator = func(maxConcurrent int) (*PDFGenerator, error) {
		calls++
		return original(maxConcurrent)
	}
	defer func() { newPDFGenerator = original }()

	server := NewServerWithOptions(ServerOptions{})
	server.closeInitializedPDF()
	if calls != 0 {
		t.Fatal("stopping a preview initialized the PDF generator")
	}
}

// TestOwnerLossArmsWatchdogWhileProtocolWriteBlocks pins that the loss path
// never waits for an in-flight protocol write: with the ready frame stuck in a
// blocked writer, owner loss must still close the loss channel and arm the
// watchdog.
func TestOwnerLossArmsWatchdogWhileProtocolWriteBlocks(t *testing.T) {
	t.Parallel()

	owner, ownerWriter := io.Pipe()
	stdout := newBlockingWriter()
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(stdout.release) }) }
	t.Cleanup(unblock)

	forced := make(chan int, 1)
	rt := &managedRuntime{
		opts:          ManagedOptions{Generation: "gen-stuck"},
		stdin:         owner,
		stdout:        stdout,
		cleanDeadline: 100 * time.Millisecond,
		forceExit:     func(code int) { forced <- code },
	}
	finished := make(chan error, 1)
	go func() {
		finished <- rt.run(func(_ context.Context, writer *managedWriter) error {
			_, err := rt.publishReady(writer, "http://127.0.0.1:1/", nil)
			return err
		})
	}()

	select {
	case <-stdout.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("the ready write never reached the blocked writer")
	}
	if err := ownerWriter.Close(); err != nil {
		t.Fatalf("close owner writer: %v", err)
	}

	select {
	case code := <-forced:
		if code == 0 {
			t.Fatal("watchdog reported success instead of a forced exit")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("watchdog was not armed while the protocol write was blocked")
	}

	unblock()
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("run did not return after the blocked write was released")
	}
}

// TestOwnerLossTreatsReadErrorAsLoss pins that an ownership read error, not
// only EOF, ends the managed run through the same cancellation path.
func TestOwnerLossTreatsReadErrorAsLoss(t *testing.T) {
	t.Parallel()

	forced := make(chan int, 1)
	rt := &managedRuntime{
		opts:          ManagedOptions{Generation: "gen-readerr"},
		stdin:         errorReader{},
		stdout:        io.Discard,
		cleanDeadline: 5 * time.Second,
		forceExit:     func(code int) { forced <- code },
	}
	canceled := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		finished <- rt.run(func(ctx context.Context, _ *managedWriter) error {
			<-ctx.Done()
			close(canceled)
			return nil
		})
	}()

	select {
	case <-canceled:
	case <-time.After(2 * time.Second):
		t.Fatal("ownership read error did not cancel startup")
	}
	select {
	case err := <-finished:
		if err != nil {
			t.Fatalf("run returned %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not return after an ownership read error")
	}
	select {
	case code := <-forced:
		t.Fatalf("graceful cleanup was force-exited with code %d", code)
	default:
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("ownership read failed") }

type blockingWriter struct {
	once    sync.Once
	entered chan struct{}
	release chan struct{}
}

func newBlockingWriter() *blockingWriter {
	return &blockingWriter{entered: make(chan struct{}), release: make(chan struct{})}
}

func (w *blockingWriter) Write(p []byte) (int, error) {
	w.once.Do(func() { close(w.entered) })
	<-w.release
	return len(p), nil
}

func writeManagedDoc(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readManagedLine(t *testing.T, reader *bufio.Reader, timeout time.Duration) string {
	t.Helper()

	type result struct {
		line string
		err  error
	}
	lines := make(chan result, 1)
	go func() {
		line, err := reader.ReadString('\n')
		lines <- result{line: line, err: err}
	}()

	select {
	case res := <-lines:
		if res.err != nil {
			t.Fatalf("read managed protocol line: %v", res.err)
		}
		return strings.TrimSpace(res.line)
	case <-time.After(timeout):
		t.Fatal("timed out waiting for a managed protocol line")
		return ""
	}
}

func managedGet(t *testing.T, rawURL string) (int, string) {
	t.Helper()

	resp := managedRequest(t, http.MethodGet, rawURL)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s: %v", rawURL, err)
	}
	return resp.StatusCode, string(body)
}

func managedRequest(t *testing.T, method, rawURL string) *http.Response {
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

// managedFrames decodes the v1 NDJSON frames of one managed process, failing
// the test on any malformed or off-generation line.
func managedFrames(t *testing.T, out io.Reader, generation string) <-chan managedEvent {
	t.Helper()

	frames := make(chan managedEvent, 32)
	go func() {
		defer close(frames)
		reader := bufio.NewReader(out)
		for {
			line, err := reader.ReadString('\n')
			trimmed := strings.TrimSpace(line)
			if trimmed == "" && err != nil {
				return
			}
			var event managedEvent
			if err := json.Unmarshal([]byte(trimmed), &event); err != nil {
				t.Errorf("managed stdout line %q is not JSON: %v", line, err)
				return
			}
			if event.Version != managedProtocolVersion || event.Generation != generation {
				t.Errorf("managed frame %+v is not v1 for generation %q", event, generation)
				return
			}
			frames <- event
		}
	}()
	return frames
}

// waitForManagedFrame waits for a frame matching want, skipping the others.
func waitForManagedFrame(t *testing.T, frames <-chan managedEvent, want func(managedEvent) bool, timeout time.Duration) managedEvent {
	t.Helper()

	deadline := time.After(timeout)
	for {
		select {
		case event, ok := <-frames:
			if !ok {
				t.Fatal("managed stdout closed before the expected frame")
			}
			if want(event) {
				return event
			}
		case <-deadline:
			t.Fatal("timed out waiting for the expected managed frame")
			return managedEvent{}
		}
	}
}

// expectNoManagedFrame fails when a queued frame matching want arrives within
// wait.
func expectNoManagedFrame(t *testing.T, frames <-chan managedEvent, want func(managedEvent) bool, wait time.Duration) {
	t.Helper()

	deadline := time.After(wait)
	for {
		select {
		case event, ok := <-frames:
			if !ok {
				return
			}
			if want(event) {
				t.Fatalf("unexpected managed frame %+v", event)
			}
		case <-deadline:
			return
		}
	}
}

// TestManagedTargetStatusTracksRealAccess pins the access-driven target state:
// a failing request for an unrelated asset or a document that disappeared
// inside an accessible directory does not flip the target; losing the served
// root reports unavailable with the real path, and serving it again reports
// available.
//
// Not parallel: constructing the reloader reads os.Stderr, which the legacy
// TestServeJSONOutput replaces globally while running in parallel.
func TestManagedTargetStatusTracksRealAccess(t *testing.T) {
	root := t.TempDir()
	writeManagedDoc(t, filepath.Join(root, "README.md"), "# Root marker\n")
	writeManagedDoc(t, filepath.Join(root, "nested", "note.md"), "# Nested marker\n")

	owner, ownerWriter := io.Pipe()
	stdout, stdoutWriter := io.Pipe()
	exited := make(chan struct{})
	rt := &managedRuntime{
		opts: ManagedOptions{
			Generation:   "gen-target",
			Target:       root,
			Recursive:    true,
			EnableReload: true,
		},
		stdin:         owner,
		stdout:        stdoutWriter,
		cleanDeadline: 2 * time.Second,
		forceExit:     func(code int) { t.Errorf("unexpected forced exit %d", code) },
	}
	go func() {
		_ = rt.run(rt.serve)
		close(exited)
	}()
	t.Cleanup(func() {
		_ = ownerWriter.Close()
		_ = stdoutWriter.Close()
		select {
		case <-exited:
		case <-time.After(5 * time.Second):
			t.Errorf("managed run did not return during cleanup")
		}
	})

	frames := managedFrames(t, stdout, "gen-target")
	ready := waitForManagedFrame(t, frames, func(e managedEvent) bool { return e.Event == "ready" }, 10*time.Second)
	origin, err := url.Parse(ready.URL)
	if err != nil {
		t.Fatalf("ready URL %q: %v", ready.URL, err)
	}
	base := origin.Scheme + "://" + origin.Host
	isTargetStatus := func(e managedEvent) bool { return e.Event == "target-status" }

	if status, _ := managedGet(t, base+"/missing.png"); status != http.StatusNotFound {
		t.Fatalf("missing asset status = %d, want 404", status)
	}
	expectNoManagedFrame(t, frames, isTargetStatus, 300*time.Millisecond)

	if err := os.Remove(filepath.Join(root, "nested", "note.md")); err != nil {
		t.Fatalf("remove nested document: %v", err)
	}
	if status, _ := managedGet(t, base+"/nested/note.md"); status != http.StatusNotFound {
		t.Fatalf("removed document status = %d, want 404", status)
	}
	expectNoManagedFrame(t, frames, isTargetStatus, 300*time.Millisecond)

	if err := os.RemoveAll(root); err != nil {
		t.Fatalf("remove served root: %v", err)
	}
	// The normal browser refresh path: the already-open article URL.
	if status, _ := managedGet(t, base+"/README.md"); status != http.StatusNotFound {
		t.Fatalf("open article after root removal = %d, want its own HTTP error", status)
	}
	unavailable := waitForManagedFrame(t, frames, func(e managedEvent) bool {
		return isTargetStatus(e) && e.Target != nil && e.Target.State == "unavailable"
	}, 10*time.Second)
	if !strings.Contains(unavailable.Target.Reason, root) {
		t.Fatalf("unavailable reason = %q, want the original path %q", unavailable.Target.Reason, root)
	}
	// The directory index reports the same failure without an empty state.
	if status, body := managedGet(t, base+"/"); status < 400 || strings.Contains(body, "docs-empty") {
		t.Fatalf("removed root response = %d %q, want an error without the empty state", status, body)
	}

	writeManagedDoc(t, filepath.Join(root, "README.md"), "# Recreated marker\n")
	if status, body := managedGet(t, base+"/README.md"); status != http.StatusOK || !strings.Contains(body, "Recreated marker") {
		t.Fatalf("recreated target preview = %d %q, want the new document", status, body)
	}
	waitForManagedFrame(t, frames, func(e managedEvent) bool {
		return isTargetStatus(e) && e.Target != nil && e.Target.State == "available"
	}, 10*time.Second)
}

// TestManagedReloadStatusSuppressesSnapshotDuplicate pins the delivery gate
// across ready publication: a transition whose state and reason the ready
// snapshot already carried is not reported a second time, while a later
// transition is. The delayed-callback window is timing-dependent, so the gate
// itself is the regression pinned here.
func TestManagedReloadStatusSuppressesSnapshotDuplicate(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	rt := &managedRuntime{
		opts:            ManagedOptions{Generation: "gen-dedupe"},
		stdout:          &stdout,
		published:       true,
		publishedReload: &managedReloadSnapshot{State: "active"},
	}
	writer := newManagedWriter(&stdout, "gen-dedupe")

	rt.emitReloadStatus(writer, "active", "")
	if stdout.Len() != 0 {
		t.Fatalf("a transition already carried by the ready snapshot was emitted again: %q", stdout.String())
	}

	rt.emitReloadStatus(writer, "degraded", "watch error")
	var event managedEvent
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &event); err != nil {
		t.Fatalf("emitted line %q is not JSON: %v", stdout.String(), err)
	}
	if event.Event != "reload-status" || event.Reload == nil || event.Reload.State != "degraded" || event.Reload.Reason != "watch error" {
		t.Fatalf("event = %+v, want the later degraded transition", event)
	}
}

// TestManagedNoReloadSnapshotIsDisabled pins the decided reload-off semantics:
// the snapshot is disabled, the service still serves, and no coverage events
// are fabricated.
func TestManagedNoReloadSnapshotIsDisabled(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeManagedDoc(t, filepath.Join(root, "README.md"), "# Marker\n")

	owner, ownerWriter := io.Pipe()
	stdout, stdoutWriter := io.Pipe()
	exited := make(chan struct{})
	rt := &managedRuntime{
		opts:          ManagedOptions{Generation: "gen-noreload", Target: root, Recursive: true, EnableReload: false},
		stdin:         owner,
		stdout:        stdoutWriter,
		cleanDeadline: 2 * time.Second,
		forceExit:     func(code int) { t.Errorf("unexpected forced exit %d", code) },
	}
	go func() {
		_ = rt.run(rt.serve)
		close(exited)
	}()
	t.Cleanup(func() {
		_ = ownerWriter.Close()
		_ = stdoutWriter.Close()
		select {
		case <-exited:
		case <-time.After(5 * time.Second):
			t.Errorf("managed run did not return during cleanup")
		}
	})

	frames := managedFrames(t, stdout, "gen-noreload")
	ready := waitForManagedFrame(t, frames, func(e managedEvent) bool { return e.Event == "ready" }, 10*time.Second)
	if ready.Reload == nil || ready.Reload.State != "disabled" || ready.Reload.Reason != "" {
		t.Fatalf("ready reload snapshot = %+v, want disabled without a reason", ready.Reload)
	}
	if status, _ := managedGet(t, ready.URL); status != http.StatusOK {
		t.Fatalf("disabled-reload preview = %d, want 200", status)
	}
	expectNoManagedFrame(t, frames, func(e managedEvent) bool { return e.Event == "reload-status" }, 300*time.Millisecond)
}
