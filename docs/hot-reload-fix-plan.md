# Hot Reload Fix Plan

## Objective

Replace `aarol/reload v1.2.0`'s all-file watcher with a custom watcher that:
1. Only watches `.md` files
2. Sends the changed file path through WebSocket as `"reload:relative/path/file.md"`
3. Client-side: skips page refresh when the changed file is the currently editing file (shows Toast instead); refreshes for non-editing .md changes
4. Replaces `saveContent()`'s `window.location.reload()` with live re-render via AJAX

---

## Phase 1 — Server-side: Custom fsnotify watcher + path-aware WebSocket

### What changes

| File | Action |
|---|---|
| `internal/hotreload/hotreload.go` | **Create** — new package with custom Reloader |
| `internal/server.go` | **Modify** — replace `reload.New()` and `reloadMiddleware.Handle()` with custom Reloader |
| `go.mod` | **Modify** — remove `aarol/reload`, add `bep/debounce` as direct dependency |

### New file: `internal/hotreload/hotreload.go`

```go
package hotreload

import (
	"fmt"
	"log"
	"net/http"
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

		body := &bytes.Buffer{}
		wrap := newWrapResponseWriter(w, req.ProtoMajor)
		wrap.Tee(body)
		next.ServeHTTP(wrap, req)

		ct := w.Header().Get("Content-Type")
		if ct == "" {
			ct = http.DetectContentType(body.Bytes())
		}
		if strings.HasPrefix(ct, "text/html") {
			w.Write([]byte(script))
		}
	})
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

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
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

	_, _, err = conn.ReadMessage() // blocks until client disconnects
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
```

**Note:** `newWrapResponseWriter` is already defined in `aarol/reload` internally and not exported. We need to write a minimal `wrapResponseWriter` in this package, or avoid the Tee pattern. A simpler approach: buffer the response body ourselves.

**Simpler approach for Handle middleware** (recommended):

```go
func (r *Reloader) Handle(next http.Handler) http.Handler {
	go r.watch()
	script := r.injectedScript()

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == r.endpoint {
			r.serveWS(w, req)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")

		// Sniff content-type, then append script
		buf := &bytes.Buffer{}
		mw := io.MultiWriter(w, buf)
		next.ServeHTTP(&responseSniffer{ResponseWriter: w, writer: mw}, req)

		ct := w.Header().Get("Content-Type")
		if ct == "" {
			ct = http.DetectContentType(buf.Bytes())
		}
		if strings.HasPrefix(ct, "text/html") {
			w.Write([]byte(script))
		}
	})
}
```

This requires a small `responseSniffer` helper in the same file. See `aarol/reload/reload.go:124` for reference — the library uses a `wrapResponseWriter` that sniffs the content type by buffering the body.

### Modification to `internal/server.go` (lines 83–100)

**Remove** (delete these lines):
- `"github.com/aarol/reload"` import (line 18)
- `reloadMiddleware := reload.New(target.rootDir)` (line 85)
- `reloadMiddleware.DebugLog = log.New(io.Discard, "", 0)` (line 86)
- `reloadMiddleware.Upgrader.CheckOrigin = ...` (lines 87–90)
- `handler = reloadMiddleware.Handle(handler)` (line 96)
- `fmt.Printf("📡 Auto-reload..."` line (97)

**Add** import:
```go
"github.com/showgp/go-grip/internal/hotreload"
```

**Replace** the reload setup block (lines 83–100) with:

```go
var reloadMiddleware *hotreload.Reloader
if s.enableReload {
	reloadMiddleware = hotreload.New(filepath.Clean(target.rootDir))
	reloadMiddleware.Upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
}

handler := s.newHandlerForTarget(target)

if s.enableReload {
	handler = reloadMiddleware.Handle(handler)
	fmt.Printf("📡 Auto-reload enabled. Only .md files will trigger browser refresh.\n")
} else {
	fmt.Printf("🔄 Auto-reload disabled. Use F5 to manually refresh.\n")
}
```

### `go.mod` changes

- Remove line: `github.com/aarol/reload v1.2.0`
- Run `go mod tidy` — this will promote `github.com/bep/debounce` to a direct dependency and remove `github.com/fsnotify/fsnotify` and `github.com/gorilla/websocket` from indirect. If they become orphaned (they shouldn't — our new code imports them), add them explicitly.

---

## Phase 2 — Client-side: path-aware WebSocket handling

The injected script (Phase 1) already dispatches a `CustomEvent("reload-changed", { detail: { file: "..." } })` when `data-editing === "true"`. Phase 2 adds the listener in `editor.js`.

### Modification to `defaults/static/js/editor.js`

**Add** a new section after the `init()` call (after line 53, before `enterEditMode`):

```javascript
function handleExternalReload(e) {
	var changedFile = e.detail.file;
	if (changedFile === currentFile) {
		// The file we're editing was saved externally (likely our own save).
		// No action needed — the content is already current.
		showToast("Saved", "success");
	} else {
		// A different .md file changed externally.
		showToast("File updated: " + changedFile, "info");
	}
}
```

**Add** in `init()`, after `interceptSidebarLinks()` (after line 50), before `restoreScrollPosition()`:

```javascript
window.addEventListener("reload-changed", handleExternalReload);
```

**Add** clean-up in `exitEditMode()` (after line 193, after `pollTimer` clear):

```javascript
// Remove the reload-changed listener when exiting edit mode
// so stale events don't fire after exit.
```

Actually, since `exitEditMode` doesn't clean up the listener (which is on `window`), and the listener is harmless after exit (no toasts to show since `data-editing` is now `false`), we can skip cleanup. But for completeness, store the handler reference for removal.

**Better approach** — use a named function with addEventListener/removeEventListener in the editor IIFE scope:

```javascript
// At top of IIFE, add:
var reloadChangedHandler = null;

// In init():
reloadChangedHandler = handleExternalReload;
window.addEventListener("reload-changed", reloadChangedHandler);

// In exitEditMode(), after pollTimer cleanup:
if (reloadChangedHandler) {
	window.removeEventListener("reload-changed", reloadChangedHandler);
	reloadChangedHandler = null;
}
```

### Path comparison logic

`currentFile` is read from `body.getAttribute("data-current-file")` at line 17. This contains the URL-decoded file path with `/` separators (e.g., `docs/foo.md`).

The WebSocket message sends `reload:docs/foo.md` (with `/` separators — normalized via `filepath.ToSlash` on the server). The substring after `reload:` is extracted on the client.

Comparison is a simple `===` string match. On Windows, the server normalizes `\` to `/`, so the comparison is cross-platform safe.

### Edge case — user exits edit mode while Toast is showing

If the user exits edit mode (presses Cancel or Escape) while a Toast is showing:
- `exitEditMode()` runs, removes `data-editing` attribute
- `window.location.reload()` is NOT called (the Toast was from a different-file change, which only shows Toast, not reload)
- When the user later navigates to the changed file or refreshes, they'll see the updated content

This is acceptable. The Toast auto-removes after 1.5–3 seconds. If the user is fast enough to exit before the Toast disappears, the Toast element is harmless as a DOM orphan (it'll be removed on next page navigation or after its own `setTimeout`).

---

## Phase 3 — Replace saveContent page reload with AJAX re-render

### Modification to `defaults/static/js/editor.js` (lines 140–156)

**Current code (lines 140–156):**
```javascript
.then(function () {
	originalContent = content;
	isDirty = false;
	saveSidebarState();
	sessionStorage.setItem("go-grip-scrollTop", document.documentElement.scrollTop.toString());
	sessionStorage.setItem("go-grip-editor-saved", "true");
	showToast("Saved", "success");
	setTimeout(function () {
		window.location.reload();
	}, 600);
})
```

**Replace with:**
```javascript
.then(function () {
	originalContent = content;
	isDirty = false;
	saveSidebarState();
	showToast("Saved", "success");
	// Re-render the server-side preview area with fresh rendered HTML
	fetch("/api/raw/" + encodePath(currentFile))
		.then(function(resp) { return resp.text(); })
		.then(function(raw) {
			if (typeof marked !== "undefined") {
				var rendered = marked.parse(raw);
				var previewContent = document.querySelector(".preview-content");
				if (previewContent) {
					previewContent.innerHTML = rendered;
				}
				// Also update the live split preview if visible
				var splitPreview = document.querySelector(".editor-split-preview");
				if (splitPreview) {
					splitPreview.innerHTML = rendered;
				}
			}
		})
		.catch(function() {
			// Silent: re-render is best-effort
		});
})
```

Remove `sessionStorage` lines for `scrollTop` and `editor-saved` — they were only needed to preserve state across the page reload, which no longer happens.

**Remove** from `restoreScrollPosition()` (lines 391–403):
```javascript
sessionStorage.removeItem("go-grip-editor-saved");
```
This line becomes unnecessary since we never set `go-grip-editor-saved` anymore.

**Keep** `restoreScrollPosition()` for regular page navigations (sidebar links) — the `go-grip-scrollTop` set/remove flow still applies there.

---

## Summary of all file changes

| File | Change type | Lines affected |
|---|---|---|
| `internal/hotreload/hotreload.go` | **CREATE** | Entire file (new package) |
| `internal/server.go` | MODIFY | Lines 3–21 (imports), 83–100 (reload setup) |
| `go.mod` | MODIFY | Remove `aarol/reload` dependency |
| `defaults/static/js/editor.js` | MODIFY | Add `handleExternalReload` + listener (~25 lines); replace `saveContent` reload (~15 lines) |

## Migration steps

1. Create `internal/hotreload/hotreload.go`
2. Modify `internal/server.go` — swap imports, swap reloader creation
3. Modify `defaults/static/js/editor.js` — Phase 2 + Phase 3 changes
4. Run `go mod tidy`
5. Run `go build` to verify compilation
6. Test: start server with `--reload`, edit a `.md` file, verify only `.md` changes trigger WebSocket messages (check browser console for `reload:` messages). Verify CSS/JS changes do NOT trigger reload. Verify save in editor does not reload page.
