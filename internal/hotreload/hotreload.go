package hotreload

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

const wsVersion = "2"

type client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type Reloader struct {
	rootDir   string
	endpoint  string
	errorLog  *log.Logger
	Upgrader  websocket.Upgrader
	clients   map[*client]bool
	clientsMu sync.RWMutex
}

func New(rootDir string) *Reloader {
	r := &Reloader{
		rootDir:  rootDir,
		endpoint: "/reload_ws",
		errorLog: log.New(os.Stderr, "HotReload: ", log.Lmsgprefix|log.Ltime),
		Upgrader: websocket.Upgrader{},
		clients:  make(map[*client]bool),
	}
	go r.watch()
	return r
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
		return
	}
	defer func() { _ = watcher.Close() }()

	absRoot := r.rootDir
	if !filepath.IsAbs(absRoot) {
		abs, err := filepath.Abs(absRoot)
		if err != nil {
			r.errorLog.Printf("abs path error: %s\n", err)
			return
		}
		absRoot = abs
	}

	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			r.errorLog.Printf("walk error at %s: %s\n", path, err)
			return nil
		}
		if d.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		r.errorLog.Printf("walk error: %s\n", err)
		return
	}

	deb := newDebouncer()

	for {
		select {
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
				fi, fiErr := os.Stat(e.Name)
				if fiErr == nil && fi.IsDir() {
					_ = filepath.WalkDir(e.Name, func(path string, d os.DirEntry, walkErr error) error {
						if walkErr != nil {
							return nil
						}
						if d.IsDir() {
							_ = watcher.Add(path)
						}
						return nil
					})
				} else {
					_ = watcher.Add(e.Name)
				}
				r.handleEvent(e.Name, deb)
			case e.Has(fsnotify.Write):
				r.handleEvent(e.Name, deb)
			case e.Has(fsnotify.Rename), e.Has(fsnotify.Remove):
				_ = watcher.Remove(e.Name)
			}
		}
	}
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
	if !strings.HasSuffix(strings.ToLower(name), ".md") {
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

	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Minute))
	_, _, err = conn.ReadMessage()
	if err != nil {
		r.errorLog.Printf("read error: %s\n", err)
	}

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
    if(isRetry && document.body.getAttribute("data-editing") !== "true") { window.location.reload() }
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
