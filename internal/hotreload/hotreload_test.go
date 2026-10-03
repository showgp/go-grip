package hotreload

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestDocDirs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for _, name := range []string{
		"README.md",
		"docs/a/b/deep.md",
		"docs/other/note.md",
		"node_modules/pkg/readme.md",
		"node_modules/pkg/nested/more.md",
		".git/notes.md",
		"assets",
	} {
		writeTestFile(t, filepath.Join(root, filepath.FromSlash(name)))
	}

	tests := []struct {
		name      string
		recursive bool
		want      []string
	}{
		{
			name:      "recursive watches documents and their ancestors",
			recursive: true,
			want:      []string{".", "docs", "docs/a", "docs/other", "docs/a/b"},
		},
		{
			name:      "non recursive watches the root only",
			recursive: false,
			want:      []string{"."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := testReloader(root, tt.recursive, fallbackBudget)
			plans, _ := r.docPlans(root)

			got := make([]string, 0, len(plans))
			for _, plan := range plans {
				rel, err := filepath.Rel(root, plan.path)
				if err != nil {
					t.Fatalf("relative path of %q: %v", plan.path, err)
				}
				got = append(got, filepath.ToSlash(rel))
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("watched directories\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// TestAddDirsStopsAtDescriptorExhaustion pins the liveness guarantee: when the
// descriptor table fills mid-registration, the watch set is trimmed and the
// partially watched directory is rolled back rather than left registered.
func TestAddDirsStopsAtDescriptorExhaustion(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))
	writeTestFile(t, filepath.Join(root, "docs", "a", "b", "deep.md"))

	docsDir := filepath.Join(root, "docs")
	fake := newFakeAdder(1, &os.PathError{Op: "open", Path: docsDir, Err: syscall.EMFILE})
	r := testReloader(root, true, fallbackBudget)

	added, doc := r.addDirectories(fake, root)
	if added != 1 {
		t.Fatalf("registered %d directories, want 1 before exhaustion", added)
	}
	if doc == "" {
		t.Fatalf("expected the discoverable markdown to be reported")
	}
	if got := fake.watched(); !slices.Equal(got, []string{root}) {
		t.Fatalf("watched %q, want only the root", got)
	}
	if got := fake.removed(); !slices.Equal(got, []string{docsDir}) {
		t.Fatalf("removed %q, want the partially registered directory", got)
	}
}

// TestAdoptRespectsWatchBudget covers a directory arriving after startup while
// the budget is already spent: it must not be registered at all.
func TestAdoptRespectsWatchBudget(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))
	newDir := filepath.Join(root, "fresh")
	writeTestFile(t, filepath.Join(newDir, "fresh.md"))

	fake := newFakeAdder(-1, nil)
	r := testReloader(root, true, fallbackBudget)
	r.watchCost = r.maxFDs // budget already spent

	r.adopt(fake, newDir, newDebouncer())

	if got := fake.watched(); len(got) != 0 {
		t.Fatalf("watched %q, want nothing once the budget is spent", got)
	}
}

// TestAdoptRollsBackOnDescriptorExhaustion covers the descriptor table filling
// while a newly arrived directory is adopted.
func TestAdoptRollsBackOnDescriptorExhaustion(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	newDir := filepath.Join(root, "fresh")
	writeTestFile(t, filepath.Join(newDir, "fresh.md"))

	fake := newFakeAdder(0, &os.PathError{Op: "open", Path: newDir, Err: syscall.EMFILE})
	r := testReloader(root, true, fallbackBudget)
	r.adopt(fake, newDir, newDebouncer())

	if got := fake.watched(); len(got) != 0 {
		t.Fatalf("watched %q, want nothing after exhaustion", got)
	}
	if got := fake.removed(); !slices.Equal(got, []string{newDir}) {
		t.Fatalf("removed %q, want the partially registered directory", got)
	}
}

// TestAdoptRegistersEachDirectoryOnce covers the budget arithmetic for a
// subtree whose root was already registered before the scan.
func TestAdoptRegistersEachDirectoryOnce(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	newDir := filepath.Join(root, "fresh")
	writeTestFile(t, filepath.Join(newDir, "deep", "note.md"))

	fake := newFakeAdder(-1, nil)
	r := testReloader(root, true, fallbackBudget)
	r.adopt(fake, newDir, newDebouncer())

	watched := fake.watched()
	if !slices.Contains(watched, newDir) {
		t.Fatalf("watched %q, want the new root", watched)
	}
	if !slices.Contains(watched, filepath.Join(newDir, "deep")) {
		t.Fatalf("watched %q, want the nested document directory too", watched)
	}
	if len(watched) != 2 {
		t.Fatalf("watched %q, want each directory registered exactly once", watched)
	}

	// The budget must account for both registrations, not just the directories.
	wantCost, err := dirWatchCost(newDir)
	if err != nil {
		t.Fatalf("cost of %s: %v", newDir, err)
	}
	deepCost, err := dirWatchCost(filepath.Join(newDir, "deep"))
	if err != nil {
		t.Fatalf("cost of nested directory: %v", err)
	}
	if r.watchCost != wantCost+deepCost {
		t.Fatalf("watchCost = %d, want %d", r.watchCost, wantCost+deepCost)
	}
}

// TestWatchSetBudgetCountsDirectoryEntries pins that the budget is spent on what
// watching a directory really costs — one unit per entry inside it on kqueue —
// and not on the number of directories.
func TestWatchSetBudgetCountsDirectoryEntries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))
	heavy := filepath.Join(root, "heavy")
	for i := range 20 {
		writeTestFile(t, filepath.Join(heavy, fmt.Sprintf("doc%02d.md", i)))
	}

	fake := newFakeAdder(-1, nil)
	// Room for the root plus everything but the last unit of the heavy directory.
	rootCost, err := dirWatchCost(root)
	if err != nil {
		t.Fatalf("cost of root: %v", err)
	}
	heavyCost, err := dirWatchCost(heavy)
	if err != nil {
		t.Fatalf("cost of heavy directory: %v", err)
	}
	r := testReloader(root, true, rootCost+heavyCost-1)

	added, _ := r.addDirectories(fake, root)
	if added != 1 {
		t.Fatalf("registered %d directories, want 1 (the heavy one must not fit)", added)
	}
	if got := fake.watched(); !slices.Equal(got, []string{root}) {
		t.Fatalf("watched %q, want only the root", got)
	}
}

// TestWatchSetBudgetTruncatesWatchSet pins the budget against a real watcher.
func TestWatchSetBudgetTruncatesWatchSet(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))
	writeTestFile(t, filepath.Join(root, "docs", "a", "b", "deep.md"))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("new watcher: %v", err)
	}
	defer func() { _ = watcher.Close() }()

	rootCost, err := dirWatchCost(root)
	if err != nil {
		t.Fatalf("cost of root: %v", err)
	}
	r := testReloader(root, true, rootCost)
	if added, _ := r.addDirectories(watcher, root); added != 1 {
		t.Fatalf("registered %d directories, want 1", added)
	}

	watched := watcher.WatchList()
	if len(watched) != 1 {
		t.Fatalf("watched %d directories, want 1: %q", len(watched), watched)
	}
	if watched[0] != root {
		t.Fatalf("watched %q, want the shallowest directory %q", watched, root)
	}
}

// TestWatchSetBudgetKeepsServingReloads drives a real watcher whose budget only
// covers the root: reloads from the covered directory must keep working while
// documents below the truncated directories stay silent.
func TestWatchSetBudgetKeepsServingReloads(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))
	writeTestFile(t, filepath.Join(root, "docs", "deep.md"))

	rootCost, err := dirWatchCost(root)
	if err != nil {
		t.Fatalf("cost of root: %v", err)
	}
	r := testReloader(root, true, rootCost)
	startReloader(t, r)
	messages := dialReloader(t, startReloaderServer(t, r))

	// Beyond the budget: reporting this would prove the budget is not applied.
	appendTestFile(t, filepath.Join(root, "docs", "deep.md"))
	expectNoMessage(t, messages, 500*time.Millisecond)

	appendTestFile(t, filepath.Join(root, "README.md"))
	expectMessage(t, messages, "reload:README.md")
}

func TestReloadMessages(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "docs", "guide.md"))
	writeTestFile(t, filepath.Join(root, "node_modules", "pkg", "readme.md"))

	r := testReloader(root, true, fallbackBudget)
	startReloader(t, r)
	messages := dialReloader(t, startReloaderServer(t, r))

	appendTestFile(t, filepath.Join(root, "docs", "guide.md"))
	expectMessage(t, messages, "reload:docs/guide.md")

	appendTestFile(t, filepath.Join(root, "node_modules", "pkg", "readme.md"))
	expectNoMessage(t, messages, 500*time.Millisecond)
}

// TestReloadWhenDocumentArrivesAfterStartup covers a root that holds no
// markdown when the server starts: the root must still be watched so the first
// document created in it is noticed.
func TestReloadWhenDocumentArrivesAfterStartup(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	r := testReloader(root, true, fallbackBudget)
	startReloader(t, r)
	messages := dialReloader(t, startReloaderServer(t, r))

	writeTestFile(t, filepath.Join(root, "first.md"))
	expectMessage(t, messages, "reload:first.md")
}

// TestReloadForNewDirectory covers a directory arriving with its markdown: no
// per-file event is emitted, so the reloader announces the new document itself.
func TestReloadForNewDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))

	fresh := filepath.Join(t.TempDir(), "fresh")
	writeTestFile(t, filepath.Join(fresh, "fresh.md"))

	r := testReloader(root, true, fallbackBudget)
	startReloader(t, r)
	messages := dialReloader(t, startReloaderServer(t, r))

	if err := os.Rename(fresh, filepath.Join(root, "fresh")); err != nil {
		t.Fatalf("move directory into place: %v", err)
	}
	expectMessage(t, messages, "reload:fresh/fresh.md")
}

// TestReloadForDocumentInNewEmptyDirectory covers the package-manager-style
// sequence of creating an empty directory and filling it immediately after: the
// document must be caught by the scan or by the watch installed before it.
func TestReloadForDocumentInNewEmptyDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))

	r := testReloader(root, true, fallbackBudget)
	startReloader(t, r)
	messages := dialReloader(t, startReloaderServer(t, r))

	if err := os.Mkdir(filepath.Join(root, "later"), 0o755); err != nil {
		t.Fatalf("create directory: %v", err)
	}
	writeTestFile(t, filepath.Join(root, "later", "note.md"))
	expectMessage(t, messages, "reload:later/note.md")
}

// TestIgnoredDirectoryCreatedAfterStartup guards against a package manager
// materialising node_modules after startup and restoring the descriptor
// pressure this watcher exists to avoid.
func TestIgnoredDirectoryCreatedAfterStartup(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))

	staging := filepath.Join(t.TempDir(), "node_modules")
	writeTestFile(t, filepath.Join(staging, "pkg", "readme.md"))

	r := testReloader(root, true, fallbackBudget)
	startReloader(t, r)
	messages := dialReloader(t, startReloaderServer(t, r))

	if err := os.Rename(staging, filepath.Join(root, "node_modules")); err != nil {
		t.Fatalf("move node_modules into place: %v", err)
	}
	expectNoMessage(t, messages, 500*time.Millisecond)

	appendTestFile(t, filepath.Join(root, "README.md"))
	expectMessage(t, messages, "reload:README.md")
}

// TestStateReportsPendingThenActive pins the initial watch snapshot the managed
// ready event carries: pending before the initial watch set is registered,
// active once it is complete.
func TestStateReportsPendingThenActive(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))

	r := testReloader(root, true, fallbackBudget)
	if got, _ := r.Status(); got != StatePending {
		t.Fatalf("state before the watch loop = %q, want %q", got, StatePending)
	}

	startReloader(t, r)
	if got, reason := r.Status(); got != StateActive || reason != "" {
		t.Fatalf("state after registration = (%q, %q), want active without a reason", got, reason)
	}
}

// TestStateReportsDegradedWhenWatchSetIsTruncated pins that an incomplete
// initial watch set is not reported as active coverage.
func TestStateReportsDegradedWhenWatchSetIsTruncated(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))
	writeTestFile(t, filepath.Join(root, "docs", "note.md"))

	// A zero budget registers no directory at all, leaving the tree unwatched.
	r := testReloader(root, true, 0)
	startReloader(t, r)

	if got, _ := r.Status(); got != StateDegraded {
		t.Fatalf("state with a truncated watch set = %q, want %q", got, StateDegraded)
	}
}

// TestStateReportsDegradedWhenRootCannotBeRead pins that a watch set which
// could not even be discovered is not reported as active coverage.
func TestStateReportsDegradedWhenRootCannotBeRead(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })
	if err := os.Chmod(root, 0o000); err != nil {
		t.Fatalf("chmod %s: %v", root, err)
	}
	if entries, err := os.ReadDir(root); err == nil {
		_ = entries
		t.Skip("directory permissions are not enforced for this user")
	}

	r := testReloader(root, false, fallbackBudget)
	startReloader(t, r)

	if got, reason := r.Status(); got != StateDegraded || reason == "" {
		t.Fatalf("state after an undiscoverable watch set = (%q, %q), want degraded with a reason", got, reason)
	}
}

// TestStopWaitsForWatchLoopAndClosesClients pins the managed shutdown
// contract: Stop returns only after the watch loop ended and every connected
// WebSocket client was closed, rather than merely signalling the loop.
func TestStopWaitsForWatchLoopAndClosesClients(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))

	r := testReloader(root, true, fallbackBudget)
	startReloader(t, r)

	server := httptest.NewServer(r.Handle(http.NotFoundHandler()))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/reload_ws?v=" + wsVersion
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", url, err)
	}
	defer func() { _ = conn.Close() }()
	waitForClients(t, r, 1)

	stopped := make(chan struct{})
	go func() {
		r.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not return after the watch loop ended")
	}

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("expected Stop to close the connected WebSocket client")
	}

	// Stop stays idempotent once the loop is gone.
	r.Stop()
}

func waitForClients(t *testing.T, r *Reloader, want int) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		r.clientsMu.RLock()
		got := len(r.clients)
		r.clientsMu.RUnlock()
		if got == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d connected clients", want)
}

// testReloader builds an inactive reloader; call startReloader to run it.
func testReloader(rootDir string, recursive bool, maxDirs int) *Reloader {
	return testReloaderReporting(rootDir, recursive, maxDirs, nil)
}

// testReloaderReporting builds an inactive reloader with a coverage reporter.
func testReloaderReporting(rootDir string, recursive bool, maxDirs int, report StateReporter) *Reloader {
	r := newReloader(rootDir, recursive, maxDirs, report)
	r.errorLog = log.New(io.Discard, "", 0)
	return r
}

// startReloader runs the watch loop and waits until the initial watch set is
// registered, then stops it when the test ends.
func startReloader(t *testing.T, r *Reloader) {
	t.Helper()

	go r.watch()
	t.Cleanup(r.stop)
	select {
	case <-r.ready:
	case <-time.After(10 * time.Second):
		t.Fatalf("watcher did not become ready")
	}
}

func startReloaderServer(t *testing.T, r *Reloader) string {
	t.Helper()

	server := httptest.NewServer(r.Handle(http.NotFoundHandler()))
	t.Cleanup(server.Close)
	return server.URL
}

func dialReloader(t *testing.T, serverURL string) <-chan string {
	t.Helper()

	url := "ws" + strings.TrimPrefix(serverURL, "http") + "/reload_ws?v=" + wsVersion
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", url, err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	messages := make(chan string, 8)
	go func() {
		defer close(messages)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			messages <- string(data)
		}
	}()
	return messages
}

// expectMessage waits for want. A directory arriving with its markdown can be
// reported twice — once by the directory adoption and once by the file event —
// so identical repeats are ignored; any other message fails the test.
func expectMessage(t *testing.T, messages <-chan string, want string) {
	t.Helper()

	deadline := time.After(5 * time.Second)
	for {
		select {
		case got, ok := <-messages:
			if !ok {
				t.Fatalf("connection closed before %q", want)
			}
			if got == want {
				return
			}
			t.Fatalf("got message %q, want %q", got, want)
		case <-deadline:
			t.Fatalf("timed out waiting for %q", want)
		}
	}
}

func expectNoMessage(t *testing.T, messages <-chan string, wait time.Duration) {
	t.Helper()

	select {
	case got := <-messages:
		t.Fatalf("unexpected message %q", got)
	case <-time.After(wait):
	}
}

func writeTestFile(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("# Doc\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func appendTestFile(t *testing.T, path string) {
	t.Helper()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	if _, err := f.WriteString("<!-- changed -->\n"); err != nil {
		_ = f.Close()
		t.Fatalf("append to %s: %v", path, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close %s: %v", path, err)
	}
}

// fakeAdder fails the Add call at index failAt with err and records the calls.
type fakeAdder struct {
	failAt  int
	err     error
	mu      sync.Mutex
	adds    []string
	removes []string
}

func newFakeAdder(failAt int, err error) *fakeAdder {
	return &fakeAdder{failAt: failAt, err: err}
}

func (f *fakeAdder) Add(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.adds) == f.failAt {
		return f.err
	}
	f.adds = append(f.adds, path)
	return nil
}

func (f *fakeAdder) Remove(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.removes = append(f.removes, path)
	return nil
}

func (f *fakeAdder) WatchList() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.adds)
}

func (f *fakeAdder) watched() []string { return f.WatchList() }

func (f *fakeAdder) removed() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.removes)
}

// recordingReporter captures the coverage transitions handed to it.
type recordingReporter struct {
	mu     sync.Mutex
	events []reportedState
}

type reportedState struct {
	state  State
	reason string
}

func (r *recordingReporter) report(state State, reason string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, reportedState{state: state, reason: reason})
}

func (r *recordingReporter) snapshot() []reportedState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.events)
}

// waitForReport waits until a recorded transition matches want.
func waitForReport(t *testing.T, r *recordingReporter, want func(reportedState) bool) reportedState {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, ev := range r.snapshot() {
			if want(ev) {
				return ev
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for a reported state; got %+v", r.snapshot())
	return reportedState{}
}

// stubWatcher is a watchAdder whose registration outcome and event channels the
// test controls.
type stubWatcher struct {
	err    error
	events chan fsnotify.Event
	errors chan error
}

func newStubWatcher() *stubWatcher {
	return &stubWatcher{events: make(chan fsnotify.Event), errors: make(chan error)}
}

func (s *stubWatcher) Add(string) error    { return s.err }
func (s *stubWatcher) Remove(string) error { return nil }
func (s *stubWatcher) WatchList() []string { return nil }

// stubSource returns the openSource seam serving a stub watcher.
func stubSource(s *stubWatcher) func() (*watchIO, error) {
	return func() (*watchIO, error) {
		return &watchIO{watchAdder: s, events: s.events, errors: s.errors, close: func() error { return nil }}, nil
	}
}

// TestReporterReportsWatchSetupFailure pins that a watcher which cannot be
// created at all is reported as degraded with the setup error, so the ready
// snapshot can never claim active coverage.
func TestReporterReportsWatchSetupFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))

	recorder := &recordingReporter{}
	r := testReloaderReporting(root, true, fallbackBudget, recorder.report)
	r.openSource = func() (*watchIO, error) { return nil, errors.New("no watch backend") }

	go r.watch()
	t.Cleanup(r.stop)
	select {
	case <-r.ready:
	case <-time.After(10 * time.Second):
		t.Fatal("watcher setup failure did not finish the initial snapshot")
	}

	got := waitForReport(t, recorder, func(ev reportedState) bool { return ev.state == StateDegraded })
	if !strings.Contains(got.reason, "no watch backend") {
		t.Fatalf("reason = %q, want the watcher setup error", got.reason)
	}
	if state, reason := r.Status(); state != StateDegraded || reason == "" {
		t.Fatalf("Status() = (%q, %q), want degraded with a reason", state, reason)
	}
}

// TestReporterReportsWalkError pins that a served tree the initial scan cannot
// read is reported degraded with the failing path.
func TestReporterReportsWalkError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))
	sealed := filepath.Join(root, "sealed")
	writeTestFile(t, filepath.Join(sealed, "note.md"))
	if err := os.Chmod(sealed, 0o000); err != nil {
		t.Fatalf("chmod %s: %v", sealed, err)
	}
	t.Cleanup(func() { _ = os.Chmod(sealed, 0o755) })
	if entries, err := os.ReadDir(sealed); err == nil {
		_ = entries
		t.Skip("directory permissions are not enforced for this user")
	}

	recorder := &recordingReporter{}
	r := testReloaderReporting(root, true, fallbackBudget, recorder.report)
	startReloader(t, r)

	got := waitForReport(t, recorder, func(ev reportedState) bool { return ev.state == StateDegraded })
	if !strings.Contains(got.reason, sealed) {
		t.Fatalf("reason = %q, want the unscanned directory %q", got.reason, sealed)
	}
}

// TestReporterReportsBudgetTruncation pins that a watch set truncated by the
// budget is reported degraded with a reason instead of a false active.
func TestReporterReportsBudgetTruncation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))

	recorder := &recordingReporter{}
	r := testReloaderReporting(root, true, 0, recorder.report)
	startReloader(t, r)

	got := waitForReport(t, recorder, func(ev reportedState) bool { return ev.state == StateDegraded })
	if got.reason == "" {
		t.Fatal("degraded state was reported without a reason")
	}
}

// TestReporterReportsRegistrationExhaustion pins that a watch whose
// registration fails (resource exhaustion) is reported degraded with the
// failure instead of being ignored.
func TestReporterReportsRegistrationExhaustion(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))

	recorder := &recordingReporter{}
	r := testReloaderReporting(root, true, fallbackBudget, recorder.report)
	stub := newStubWatcher()
	stub.err = &os.PathError{Op: "open", Path: root, Err: syscall.EMFILE}
	r.openSource = stubSource(stub)
	startReloader(t, r)

	got := waitForReport(t, recorder, func(ev reportedState) bool { return ev.state == StateDegraded })
	if got.reason == "" {
		t.Fatal("degraded state was reported without a reason")
	}
}

// TestReporterReportsRuntimeWatchErrorOnce pins the running transition: a known
// watch error moves an active watch set to degraded with the error text, and a
// repeated error does not produce another state event.
func TestReporterReportsRuntimeWatchErrorOnce(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))

	recorder := &recordingReporter{}
	r := testReloaderReporting(root, true, fallbackBudget, recorder.report)
	stub := newStubWatcher()
	r.openSource = stubSource(stub)
	startReloader(t, r)

	waitForReport(t, recorder, func(ev reportedState) bool { return ev.state == StateActive })

	stub.errors <- errors.New("watch stream collapsed")

	got := waitForReport(t, recorder, func(ev reportedState) bool { return ev.state == StateDegraded })
	if !strings.Contains(got.reason, "watch stream collapsed") {
		t.Fatalf("reason = %q, want the watch error", got.reason)
	}

	count := len(recorder.snapshot())
	stub.errors <- errors.New("watch stream collapsed again")
	time.Sleep(200 * time.Millisecond)
	if got := len(recorder.snapshot()); got != count {
		t.Fatalf("reported %d transitions after a repeated error, want %d (a degraded state is reported once)", got, count)
	}
}

// TestReporterReportsRuntimeCoverageShortfall pins that coverage lost after
// startup — a new directory that does not fit the watch budget — is reported
// degraded instead of silently staying active, driven through the real watch
// loop dispatch of a create event.
func TestReporterReportsRuntimeCoverageShortfall(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))
	rootCost, err := dirWatchCost(root)
	if err != nil {
		t.Fatalf("cost of root: %v", err)
	}

	recorder := &recordingReporter{}
	// The initial watch set fits exactly, so the service starts active.
	r := testReloaderReporting(root, true, rootCost, recorder.report)
	stub := newStubWatcher()
	r.openSource = stubSource(stub)
	startReloader(t, r)
	waitForReport(t, recorder, func(ev reportedState) bool { return ev.state == StateActive })

	fresh := filepath.Join(root, "fresh")
	writeTestFile(t, filepath.Join(fresh, "note.md"))
	stub.events <- fsnotify.Event{Name: fresh, Op: fsnotify.Create}

	got := waitForReport(t, recorder, func(ev reportedState) bool { return ev.state == StateDegraded })
	if got.reason == "" {
		t.Fatal("degraded state was reported without a reason")
	}
	if state, _ := r.Status(); state != StateDegraded {
		t.Fatalf("Status() = %q, want degraded after the adopted directory did not fit", state)
	}
}
