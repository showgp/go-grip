package hotreload

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

const (
	wsVersion = "2"

	// maxWatchedDirs bounds the number of directory watches. fsnotify holds one
	// descriptor per watched directory plus one per file inside it, so
	// registering an unbounded recursive tree exhausts the process descriptor
	// table (macOS defaults to a soft limit of 256).
	maxWatchedDirs = 4096
)

// ignoredDirs are dependency and build directories that never hold served
// documentation. Skipping them keeps the watch set — and its descriptor cost —
// proportional to the actual documents.
var ignoredDirs = map[string]struct{}{
	"node_modules": {}, "bower_components": {}, "vendor": {},
	".git": {}, ".hg": {}, ".svn": {},
	".pnpm": {}, ".pnpm-store": {}, ".yarn": {},
	"dist": {}, "build": {}, "out": {}, "target": {},
	".next": {}, ".nuxt": {}, ".svelte-kit": {}, ".output": {},
	".venv": {}, "venv": {}, "__pycache__": {}, ".tox": {},
	".mypy_cache": {}, ".pytest_cache": {}, ".ruff_cache": {},
	".cache": {}, ".turbo": {}, ".parcel-cache": {}, "coverage": {},
}

type client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type Reloader struct {
	rootDir   string
	recursive bool
	endpoint  string
	errorLog  *log.Logger
	Upgrader  websocket.Upgrader
	clients   map[*client]bool
	clientsMu sync.RWMutex

	// maxDirs caps the directory watches; the zero value watches nothing.
	maxDirs int

	// ready is closed once the initial watch set is registered, and done is
	// closed by stop to end the watch loop.
	ready    chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

// watchAdder is the part of fsnotify.Watcher used to build the watch set.
type watchAdder interface {
	Add(path string) error
	Remove(path string) error
	WatchList() []string
}

// New builds a Reloader for rootDir. Only markdown reachable through the server
// needs to trigger a reload, so recursive mirrors the server's --recursive mode:
// without it, subdirectories are neither served nor watched.
func New(rootDir string, recursive bool) *Reloader {
	r := newReloader(rootDir, recursive, maxWatchedDirs)
	go r.watch()
	return r
}

func newReloader(rootDir string, recursive bool, maxDirs int) *Reloader {
	return &Reloader{
		rootDir:   rootDir,
		recursive: recursive,
		endpoint:  "/reload_ws",
		errorLog:  log.New(os.Stderr, "HotReload: ", log.Lmsgprefix|log.Ltime),
		Upgrader:  websocket.Upgrader{},
		clients:   make(map[*client]bool),
		maxDirs:   maxDirs,
		ready:     make(chan struct{}),
		done:      make(chan struct{}),
	}
}

func (r *Reloader) Endpoint() string { return r.endpoint }

func (r *Reloader) Handle(next http.Handler) http.Handler {
	script := r.injectedScript()

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == r.endpoint {
			r.serveWS(w, req)
			return
		}

		rrw := &reloadResponseWriter{header: make(http.Header)}
		rrw.header.Set("Cache-Control", "no-cache")
		next.ServeHTTP(rrw, req)

		for k, v := range rrw.header {
			if k == "Content-Length" {
				continue
			}
			w.Header()[k] = v
		}

		body := rrw.buf.Bytes()
		ct := w.Header().Get("Content-Type")
		if ct == "" {
			ct = http.DetectContentType(body)
			w.Header().Set("Content-Type", ct)
		}
		if strings.HasPrefix(ct, "text/html") {
			scriptBytes := []byte(script)
			if idx := bytes.LastIndex(body, []byte("</body>")); idx != -1 {
				body = bytes.Join([][]byte{body[:idx], scriptBytes, body[idx:]}, nil)
			} else {
				body = append(body, scriptBytes...)
			}
		}
		w.WriteHeader(rrw.code)
		_, _ = w.Write(body)
	})
}

type reloadResponseWriter struct {
	buf         bytes.Buffer
	header      http.Header
	code        int
	wroteHeader bool
}

func (w *reloadResponseWriter) Header() http.Header { return w.header }

func (w *reloadResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.code = code
}

func (w *reloadResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.buf.Write(b)
}

func (r *Reloader) watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		r.errorLog.Printf("fsnotify error: %s\n", err)
		close(r.ready)
		return
	}
	defer func() { _ = watcher.Close() }()

	r.register(watcher)
	close(r.ready)

	deb := newDebouncer()

	for {
		select {
		case <-r.done:
			return
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			r.errorLog.Printf("watch error: %s\n", err)
		case e, ok := <-watcher.Events:
			if !ok {
				return
			}
			switch {
			case e.Has(fsnotify.Create):
				r.handleCreate(watcher, e.Name, deb)
			case e.Has(fsnotify.Write):
				r.handleEvent(e.Name, deb)
			case e.Has(fsnotify.Rename), e.Has(fsnotify.Remove):
				_ = watcher.Remove(e.Name)
			}
		}
	}
}

// register adds the initial watch set for the served root.
func (r *Reloader) register(watch watchAdder) {
	root := r.rootDir
	if !filepath.IsAbs(root) {
		abs, err := filepath.Abs(root)
		if err != nil {
			r.errorLog.Printf("abs path error: %s\n", err)
			return
		}
		root = abs
	}
	_, _ = r.addDirectories(watch, root)
}

// stop ends the watch loop and releases its descriptors.
func (r *Reloader) stop() {
	r.stopOnce.Do(func() { close(r.done) })
}

// Stop ends the watch loop and releases its descriptors. It is safe to call
// more than once.
func (r *Reloader) Stop() { r.stop() }

// handleCreate picks up directories and markdown files that appeared after the
// initial walk.
func (r *Reloader) handleCreate(watch *fsnotify.Watcher, name string, deb *debouncer) {
	fi, err := os.Stat(name)
	if err != nil {
		return
	}
	if fi.IsDir() {
		if r.recursive && !isIgnoredDir(filepath.Base(name)) {
			r.adopt(watch, name, deb)
		}
		return
	}
	if !isMarkdown(name) {
		return
	}
	// The parent directory watch already reports writes to its markdown, and
	// fsnotify registers per-entry watches itself where the platform needs them.
	// Adding the file by hand would spend another descriptor on macOS and count
	// against the directory budget on Linux.
	r.handleEvent(name, deb)
}

// adopt registers a subtree that appeared after startup. A directory created
// together with its markdown emits no per-file event — fsnotify marks entries
// present at registration as already seen — so the documents found by the scan
// are announced here.
//
// The directory watch is installed before the scan: a document landing between
// the scan and the registration would fall into neither (the scan has already
// passed it, and registration suppresses its create event). Scanning afterwards
// means such a document is seen by the scan, or by an event from the
// already-installed watch.
func (r *Reloader) adopt(watch watchAdder, name string, deb *debouncer) {
	if len(watch.WatchList()) >= r.maxDirs {
		r.errorLog.Printf("watch %s: the %d directory limit is reached; markdown below it will not trigger a reload\n",
			name, r.maxDirs)
		return
	}
	if err := watch.Add(name); err != nil {
		if isFdExhausted(err) {
			_ = watch.Remove(name)
			r.errorLog.Printf("too many open files: %s is not watched; markdown below it will not trigger a reload\n", name)
			return
		}
		r.errorLog.Printf("watch error at %s: %s\n", name, err)
		return
	}

	dirs, firstDoc := r.docDirs(name)
	// name is already registered, so it must not be counted again.
	rest := slices.DeleteFunc(dirs, func(dir string) bool { return dir == name })
	r.registerDirs(watch, name, rest)

	if firstDoc == "" {
		// The scan found no document, but one can have landed during it. Ask
		// once more; a later arrival is covered by the watches just installed.
		firstDoc = firstDocument(name)
	}
	if firstDoc != "" {
		r.handleEvent(firstDoc, deb)
	}
}

// addDirectories registers a watch on every directory that can hold a served
// document and reports the first document found below root. Running out of
// descriptors or hitting the watch-set limit degrades the watch set instead of
// killing the reloader: the directories registered so far keep working.
func (r *Reloader) addDirectories(watch watchAdder, root string) (int, string) {
	dirs, firstDoc := r.docDirs(root)
	return r.registerDirs(watch, root, dirs), firstDoc
}

// registerDirs installs a watch per directory, stopping at the directory limit
// or when the descriptor table fills. Registered directories keep working.
func (r *Reloader) registerDirs(watch watchAdder, root string, dirs []string) int {
	added := 0
	for i, dir := range dirs {
		if len(watch.WatchList()) >= r.maxDirs {
			r.errorLog.Printf("watch %s: %d of %d directories exceed the %d watch limit; markdown below the rest will not trigger a reload\n",
				root, len(dirs)-i, len(dirs), r.maxDirs)
			break
		}
		if err := watch.Add(dir); err != nil {
			if isFdExhausted(err) {
				// fsnotify's kqueue backend registers the directory and every
				// entry it managed to open before the failing one, and does not
				// roll them back on error. Drop the partial watch so those
				// descriptors return to the process instead of pinning it at
				// the limit.
				_ = watch.Remove(dir)
				r.errorLog.Printf("too many open files: %d of %d directories under %s are unwatched; markdown below them will not trigger a reload\n",
					len(dirs)-i, len(dirs), root)
				break
			}
			r.errorLog.Printf("watch error at %s: %s\n", dir, err)
			continue
		}
		added++
	}
	return added
}

// firstDocument returns any markdown directly inside dir.
func firstDocument(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() && isMarkdown(e.Name()) {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}

// docDirs lists the directories worth watching, plus the first markdown file
// found. The non-recursive mode serves only the root. The recursive mode serves
// every markdown file below the root, so their directories and the ancestors
// needed to notice new subdirectories are watched; dependency and build
// directories are skipped.
//
// The scan root is always watched, even while it holds no markdown: otherwise a
// document created in it later would go unnoticed, which is exactly how a
// directory that arrives empty gets adopted.
func (r *Reloader) docDirs(root string) ([]string, string) {
	if !r.recursive {
		return []string{root}, ""
	}

	var dirs []string
	firstDoc := ""
	walkErrors := 0
	var collect func(dir string) bool
	collect = func(dir string) bool {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if walkErrors == 0 {
				r.errorLog.Printf("walk error at %s: %s\n", dir, err)
			}
			walkErrors++
			return false
		}
		hasDocs := false
		for _, e := range entries {
			name := e.Name()
			if !e.IsDir() {
				if isMarkdown(name) {
					hasDocs = true
					if firstDoc == "" {
						firstDoc = filepath.Join(dir, name)
					}
				}
				continue
			}
			if isIgnoredDir(name) {
				continue
			}
			if collect(filepath.Join(dir, name)) {
				hasDocs = true
			}
		}
		if hasDocs {
			dirs = append(dirs, dir)
		}
		return hasDocs
	}
	if !collect(root) {
		dirs = append(dirs, root)
	}
	if walkErrors > 1 {
		r.errorLog.Printf("%d directories could not be read; markdown below them is not watched\n", walkErrors)
	}

	// Shallowest first, so a truncated watch set still covers the documents
	// closest to the root.
	slices.SortFunc(dirs, func(a, b string) int {
		sep := string(filepath.Separator)
		if da, db := strings.Count(a, sep), strings.Count(b, sep); da != db {
			return da - db
		}
		return strings.Compare(a, b)
	})
	return dirs, firstDoc
}

func isMarkdown(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".md")
}

// isIgnoredDir reports whether a directory is dependency or build output.
func isIgnoredDir(name string) bool {
	_, skip := ignoredDirs[name]
	return skip
}

type debouncer struct {
	mu  sync.Mutex
	deb map[string]func(func())
}

func newDebouncer() *debouncer {
	return &debouncer{
		deb: make(map[string]func(func())),
	}
}

func (d *debouncer) call(key string, fn func()) {
	d.mu.Lock()
	f, ok := d.deb[key]
	if !ok {
		f = debounce.New(100 * time.Millisecond)
		d.deb[key] = f
	}
	d.mu.Unlock()
	f(fn)
}

func (r *Reloader) handleEvent(name string, deb *debouncer) {
	if !isMarkdown(name) {
		return
	}
	rel, err := filepath.Rel(r.rootDir, name)
	if err != nil {
		return
	}
	msg := fmt.Sprintf("reload:%s", filepath.ToSlash(rel))
	deb.call(name, func() {
		r.broadcast(msg)
	})
}

func (r *Reloader) broadcast(msg string) {
	r.clientsMu.RLock()
	dead := make([]*client, 0)
	for cl := range r.clients {
		cl.mu.Lock()
		err := cl.conn.WriteMessage(websocket.TextMessage, []byte(msg))
		cl.mu.Unlock()
		if err != nil {
			r.errorLog.Printf("write error: %s\n", err)
			_ = cl.conn.Close()
			dead = append(dead, cl)
		}
	}
	r.clientsMu.RUnlock()

	if len(dead) > 0 {
		r.clientsMu.Lock()
		for _, cl := range dead {
			delete(r.clients, cl)
		}
		r.clientsMu.Unlock()
	}
}

func (r *Reloader) serveWS(w http.ResponseWriter, req *http.Request) {
	version := req.URL.Query().Get("v")
	if version != wsVersion {
		r.errorLog.Printf("warning: script version mismatch: client v%s != server v%s\n", version, wsVersion)
	}

	conn, err := r.Upgrader.Upgrade(w, req, nil)
	if err != nil {
		r.errorLog.Printf("upgrade error: %s\n", err)
		return
	}

	cl := &client{conn: conn}
	r.clientsMu.Lock()
	r.clients[cl] = true
	r.clientsMu.Unlock()

	// No read deadline: the client never sends messages, so we keep the
	// connection open indefinitely. ReadMessage blocks until the browser
	// tab closes, navigates away, or the page is refreshed — at which
	// point it returns a websocket.CloseError (typically code 1001
	// "going away"). This is normal WebSocket lifecycle, not a failure.
	_, _, _ = conn.ReadMessage()

	r.clientsMu.Lock()
	delete(r.clients, cl)
	r.clientsMu.Unlock()
	_ = conn.Close()
}

func (r *Reloader) injectedScript() string {
	return fmt.Sprintf(`
<script>
var retryDelay = 1000
function retry() {
  setTimeout(function(){ listen(true) }, retryDelay)
  retryDelay = Math.min(retryDelay * 2, 30000)
}
function listen(isRetry) {
  var protocol = location.protocol === "https:" ? "wss://" : "ws://"
  var ws = new WebSocket(protocol + location.host + "%s?v=%s")
  ws.onopen = function() {
    retryDelay = 1000
  }
  ws.onmessage = function(msg) {
    if(msg.data.startsWith("reload:")) {
      var filePath = msg.data.substring(7)
      if(document.body.getAttribute("data-editing") === "true") {
        window.dispatchEvent(new CustomEvent("reload-changed", { detail: { file: filePath } }))
        return
      }
      window.location.reload()
    }
  }
  ws.onclose = retry
}
listen(false)
</script>`, r.endpoint, wsVersion)
}
