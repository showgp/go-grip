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

type Reloader struct {
	rootDir   string
	endpoint  string
	debugLog  *log.Logger
	errorLog  *log.Logger
	Upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex
}

func New(rootDir string) *Reloader {
	return &Reloader{
		rootDir:  rootDir,
		endpoint: "/reload_ws",
		errorLog: log.New(os.Stderr, "HotReload: ", log.Lmsgprefix|log.Ltime),
		Upgrader: websocket.Upgrader{},
		clients:  make(map[*websocket.Conn]bool),
	}
}

func (r *Reloader) Endpoint() string { return r.endpoint }

func (r *Reloader) Handle(next http.Handler) http.Handler {
	go r.watch()
	script := r.injectedScript()

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == r.endpoint {
			r.serveWS(w, req)
			return
		}

		w.Header().Set("Cache-Control", "no-cache")

		buf := &bytes.Buffer{}
		sw := &sniffResponseWriter{ResponseWriter: w, buf: buf}
		next.ServeHTTP(sw, req)

		ct := w.Header().Get("Content-Type")
		if ct == "" {
			ct = http.DetectContentType(buf.Bytes())
		}
		if strings.HasPrefix(ct, "text/html") {
			w.Write([]byte(script))
		}
	})
}

type sniffResponseWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (w *sniffResponseWriter) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

func (r *Reloader) watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		r.errorLog.Printf("fsnotify error: %s\n", err)
		return
	}
	defer watcher.Close()

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
			return err
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

	deb := debounce.New(100 * time.Millisecond)

	for {
		select {
		case err := <-watcher.Errors:
			r.errorLog.Printf("watch error: %s\n", err)
		case e := <-watcher.Events:
			switch {
			case e.Has(fsnotify.Create):
				dir := filepath.Dir(e.Name)
				_ = watcher.Add(dir)
				r.handleEvent(e.Name, deb)
			case e.Has(fsnotify.Write):
				r.handleEvent(e.Name, deb)
			case e.Has(fsnotify.Rename), e.Has(fsnotify.Remove):
				watcher.Remove(e.Name)
			}
		}
	}
}

func (r *Reloader) handleEvent(name string, deb func(func())) {
	if !strings.HasSuffix(strings.ToLower(name), ".md") {
		return
	}
	rel, err := filepath.Rel(r.rootDir, name)
	if err != nil {
		return
	}
	msg := fmt.Sprintf("reload:%s", filepath.ToSlash(rel))
	deb(func() {
		r.broadcast(msg)
	})
}

func (r *Reloader) broadcast(msg string) {
	r.clientsMu.RLock()
	defer r.clientsMu.RUnlock()
	for conn := range r.clients {
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			r.errorLog.Printf("write error: %s\n", err)
			conn.Close()
			delete(r.clients, conn)
		}
	}
}

func (r *Reloader) serveWS(w http.ResponseWriter, req *http.Request) {
	version := req.URL.Query().Get("v")
	if version != wsVersion {
		r.errorLog.Printf("script version mismatch: v%s vs v%s\n", version, wsVersion)
	}

	conn, err := r.Upgrader.Upgrade(w, req, nil)
	if err != nil {
		r.errorLog.Printf("upgrade error: %s\n", err)
		return
	}

	r.clientsMu.Lock()
	r.clients[conn] = true
	r.clientsMu.Unlock()

	_, _, err = conn.ReadMessage()
	if err != nil {
		// client disconnected
	}
	r.clientsMu.Lock()
	delete(r.clients, conn)
	r.clientsMu.Unlock()
	conn.Close()
}

func (r *Reloader) injectedScript() string {
	return fmt.Sprintf(`
<script>
function retry() {
  setTimeout(function(){ listen(true) }, 1000)
}
function listen(isRetry) {
  var protocol = location.protocol === "https:" ? "wss://" : "ws://"
  var ws = new WebSocket(protocol + location.host + "%s?v=%s")
  if(isRetry) {
    ws.onopen = function(){ window.location.reload() }
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
