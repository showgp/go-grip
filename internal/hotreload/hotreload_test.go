package hotreload

import (
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
			r := testReloader(root, tt.recursive, maxWatchedDirs)
			dirs, _ := r.docDirs(root)

			got := make([]string, 0, len(dirs))
			for _, dir := range dirs {
				rel, err := filepath.Rel(root, dir)
				if err != nil {
					t.Fatalf("relative path of %q: %v", dir, err)
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
	r := testReloader(root, true, maxWatchedDirs)

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

// TestAdoptRespectsWatchLimit covers a directory arriving after startup while
// the directory budget is already full: it must not be registered at all, and
// in particular must not push the watch set past the limit.
func TestAdoptRespectsWatchLimit(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	newDir := filepath.Join(root, "fresh")

	fake := newFakeAdder(-1, nil)
	fake.seed(root) // budget of one is already spent
	r := testReloader(root, true, 1)
	r.adopt(fake, newDir, newDebouncer())

	if got := fake.watched(); !slices.Equal(got, []string{root}) {
		t.Fatalf("watched %q, want only the pre-existing watch", got)
	}
}

// TestAdoptRollsBackOnDescriptorExhaustion covers the descriptor table filling
// while a newly arrived directory is adopted.
func TestAdoptRollsBackOnDescriptorExhaustion(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	newDir := filepath.Join(root, "fresh")

	fake := newFakeAdder(0, &os.PathError{Op: "open", Path: newDir, Err: syscall.EMFILE})
	r := testReloader(root, true, maxWatchedDirs)
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
	r := testReloader(root, true, 10)
	r.adopt(fake, newDir, newDebouncer())

	watched := fake.watched()
	if !slices.Contains(watched, newDir) {
		t.Fatalf("watched %q, want the new root", watched)
	}
	if slices.Contains(watched, filepath.Join(newDir, "deep")) == false {
		t.Fatalf("watched %q, want the nested document directory too", watched)
	}
	if len(watched) != 2 {
		t.Fatalf("watched %q, want each directory registered exactly once", watched)
	}
}

// TestWatchSetLimitTruncatesWatchSet pins the cap against a real watcher: with a
// budget of one only the shallowest directory may be registered.
func TestWatchSetLimitTruncatesWatchSet(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))
	writeTestFile(t, filepath.Join(root, "docs", "a", "b", "deep.md"))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("new watcher: %v", err)
	}
	defer func() { _ = watcher.Close() }()

	r := testReloader(root, true, 1)
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

// TestWatchSetLimitKeepsServingReloads drives a real watcher whose budget only
// covers the root: reloads from the covered directory must keep working while
// documents below the truncated directories stay silent.
func TestWatchSetLimitKeepsServingReloads(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"))
	writeTestFile(t, filepath.Join(root, "docs", "deep.md"))

	r := testReloader(root, true, 1)
	startReloader(t, r)
	messages := dialReloader(t, startReloaderServer(t, r))

	// Beyond the budget: reporting this would prove the cap is not applied.
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

	r := testReloader(root, true, maxWatchedDirs)
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

	r := testReloader(root, true, maxWatchedDirs)
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

	r := testReloader(root, true, maxWatchedDirs)
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

	r := testReloader(root, true, maxWatchedDirs)
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

	r := testReloader(root, true, maxWatchedDirs)
	startReloader(t, r)
	messages := dialReloader(t, startReloaderServer(t, r))

	if err := os.Rename(staging, filepath.Join(root, "node_modules")); err != nil {
		t.Fatalf("move node_modules into place: %v", err)
	}
	expectNoMessage(t, messages, 500*time.Millisecond)

	appendTestFile(t, filepath.Join(root, "README.md"))
	expectMessage(t, messages, "reload:README.md")
}

// testReloader builds an inactive reloader; call startReloader to run it.
func testReloader(rootDir string, recursive bool, maxDirs int) *Reloader {
	r := newReloader(rootDir, recursive, maxDirs)
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

// seed marks paths as already watched without failing, to simulate a watcher
// that has used part of its budget.
func (f *fakeAdder) seed(paths ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.adds = append(f.adds, paths...)
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
