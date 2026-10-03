package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/showgp/go-grip/internal/hotreload"
)

const (
	// managedProtocolVersion is the host integration protocol version.
	managedProtocolVersion = 1

	// managedReadyPath is the startup confirmation route. It answers from the
	// managed path only and never touches the target or renders markdown.
	managedReadyPath = "/__gogrip/ready"

	// managedGenerationHeader carries the startup generation on the readiness
	// response.
	managedGenerationHeader = "X-GoGrip-Generation"

	// managedCleanupDeadline bounds the graceful aftermath after owner loss.
	// An independent watchdog force-exits the process beyond it, so a blocked
	// startup or file-system operation cannot leave the service unmanaged.
	managedCleanupDeadline = 2 * time.Second

	// managedShutdownTimeout bounds the graceful HTTP shutdown inside the
	// cleanup deadline.
	managedShutdownTimeout = 1 * time.Second
)

// managedReloadDisabled is the reload snapshot reported when reload is off.
const managedReloadDisabled hotreload.State = "disabled"

// ManagedOptions configures the managed preview contract the macOS host uses.
type ManagedOptions struct {
	// Generation is the host-assigned startup generation, echoed in every
	// protocol message.
	Generation string
	// Target is the previewed directory or Markdown file ("" = current dir).
	Target string
	// Recursive enables nested directory browsing.
	Recursive bool
	// EnableReload enables the hot-reload watcher.
	EnableReload bool
	// BoundingBox adds the debugging bounding box to rendered pages.
	BoundingBox bool

	// Standalone-only modes requested alongside --managed. Managed previews
	// always bind loopback on an OS-assigned port, never open a browser, and
	// reserve stdout for the protocol, so these combinations are rejected
	// before any target I/O.
	ExportRequested bool
	OutputRequested bool
	JSONRequested   bool
	HostRequested   bool
	PortRequested   bool
}

// validate reports the fatal code and reason for an inapplicable invocation.
func (opts ManagedOptions) validate() (string, error) {
	if opts.Generation == "" {
		return "invalid-generation", errors.New("managed preview requires a non-empty generation")
	}
	switch {
	case opts.ExportRequested || opts.OutputRequested:
		return "inapplicable-flags", errors.New("managed preview cannot be combined with --export/--output")
	case opts.JSONRequested:
		return "inapplicable-flags", errors.New("managed preview reserves stdout for the v1 NDJSON protocol; --json is not applicable")
	case opts.HostRequested || opts.PortRequested:
		return "inapplicable-flags", errors.New("managed preview always binds 127.0.0.1 on an OS-assigned port; --host/--port are not applicable")
	}
	return "", nil
}

// managedEvent is one v1 NDJSON protocol message.
type managedEvent struct {
	Version    int                    `json:"version"`
	Event      string                 `json:"event"`
	Generation string                 `json:"generation"`
	URL        string                 `json:"url,omitempty"`
	Reload     *managedReloadSnapshot `json:"reload,omitempty"`
	Code       string                 `json:"code,omitempty"`
	Message    string                 `json:"message,omitempty"`
}

// managedReloadSnapshot is the reload coverage at the moment ready is
// published: pending while the initial scan runs, active when the watch set is
// complete, degraded when setup or the budget left it incomplete.
type managedReloadSnapshot struct {
	State hotreload.State `json:"state"`
}

// managedWriter serializes protocol messages so concurrent emitters never
// interleave a JSON frame.
type managedWriter struct {
	mu         sync.Mutex
	enc        *json.Encoder
	generation string
}

func newManagedWriter(out io.Writer, generation string) *managedWriter {
	return &managedWriter{enc: json.NewEncoder(out), generation: generation}
}

func (w *managedWriter) emit(event managedEvent) error {
	event.Version = managedProtocolVersion
	event.Generation = w.generation
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.enc.Encode(event)
}

func (w *managedWriter) ready(url string, reload *managedReloadSnapshot) error {
	return w.emit(managedEvent{Event: "ready", URL: url, Reload: reload})
}

func (w *managedWriter) fatal(code, message string) error {
	return w.emit(managedEvent{Event: "fatal", Code: code, Message: message})
}

// ownerLiveness records that ownership already ended, so ready publication can
// be ordered against owner loss. watchOwner closes the loss channel before
// marking liveness, so arming the watchdog never waits for a blocked protocol
// write; the mutex orders the mark against publication.
type ownerLiveness struct {
	mu   sync.Mutex
	gone bool
}

func (l *ownerLiveness) markGone() {
	l.mu.Lock()
	l.gone = true
	l.mu.Unlock()
}

func (l *ownerLiveness) ended() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.gone
}

// publish runs fn unless ownership already ended. The check and the write are
// serialized against markGone: once loss is marked, a ready write that has not
// started cannot proceed. A write already in flight when the loss happens
// cannot be retracted; that case is covered by the host's readiness validation
// and by process exit.
func (l *ownerLiveness) publish(fn func() error) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.gone {
		return false, nil
	}
	return true, fn()
}

// managedRuntime owns the managed process lifecycle: the ownership watch, the
// cleanup watchdog, and the startup work.
type managedRuntime struct {
	opts          ManagedOptions
	stdin         io.Reader
	stdout        io.Writer
	cleanDeadline time.Duration
	forceExit     func(code int)
	// owner is set by run; serve uses it to suppress ready after owner loss.
	owner *ownerLiveness
}

// RunManaged runs the managed preview contract used by the macOS host. Protocol
// messages (v1 NDJSON) go to stdout; logs and diagnostics stay on stderr. Owner
// loss ends the process gracefully; a fatal startup failure is reported as a
// fatal event and returned as an error.
func RunManaged(opts ManagedOptions) error {
	rt := &managedRuntime{
		opts:          opts,
		stdin:         os.Stdin,
		stdout:        os.Stdout,
		cleanDeadline: managedCleanupDeadline,
		forceExit:     os.Exit,
	}
	return rt.run(rt.serve)
}

// run validates the invocation, establishes the ownership watch before any
// target I/O, and runs startup. Owner loss cancels startup and arms the
// watchdog; startup may substitute test work in package tests.
func (rt *managedRuntime) run(startup func(context.Context, *managedWriter) error) error {
	writer := newManagedWriter(rt.stdout, rt.opts.Generation)

	if code, err := rt.opts.validate(); err != nil {
		_ = writer.fatal(code, err.Error())
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	liveness := &ownerLiveness{}
	rt.owner = liveness
	ownerLost := watchOwner(rt.stdin, liveness)
	done := make(chan struct{})
	go func() {
		select {
		case <-ownerLost:
			cancel()
			select {
			case <-done:
			case <-time.After(rt.cleanDeadline):
				rt.forceExit(1)
			}
		case <-done:
		}
	}()

	err := startup(ctx, writer)
	close(done)
	return err
}

// watchOwner reports owner loss: EOF or a read error on the dedicated
// ownership pipe closes the returned channel. The host holds the write end,
// closing it on a normal stop, and the kernel closes it when the host crashes
// or is force-terminated.
//
// The loss channel is closed before liveness is marked: arming the watchdog
// must not wait for a blocked protocol write, while the mark still serializes
// with publication.
func watchOwner(stdin io.Reader, liveness *ownerLiveness) <-chan struct{} {
	lost := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, stdin)
		close(lost)
		liveness.markGone()
	}()
	return lost
}

// newManagedHandler adds the managed startup confirmation route in front of the
// preview handler. The route answers without touching the target.
func newManagedHandler(generation string, next http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(managedReadyPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(managedGenerationHeader, generation)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.Handle("/", next)
	return mux
}

func managedReloadSnapshotFor(reloader *hotreload.Reloader, enabled bool) *managedReloadSnapshot {
	if !enabled {
		return &managedReloadSnapshot{State: managedReloadDisabled}
	}
	return &managedReloadSnapshot{State: reloader.State()}
}

// verifyTargetAccess confirms the target can actually be served before ready is
// published. Directory access is already verified by the article scan that
// resolves the initial path; a single-file target must be a regular readable
// file, because a present path can still deny reading or not be a document at
// all (opening a FIFO could block).
func verifyTargetAccess(target serveTarget) error {
	if target.mode != modeSingleFile {
		return nil
	}
	path := filepath.Join(target.rootDir, target.initialFile)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	return file.Close()
}

// startupFailure reports a fatal startup failure, unless the owner is already
// gone: cancellation is a normal stop, not a failure.
func (rt *managedRuntime) startupFailure(ctx context.Context, writer *managedWriter, code string, err error) error {
	if ctx.Err() != nil {
		return nil
	}
	_ = writer.fatal(code, err.Error())
	return err
}

func (rt *managedRuntime) ownershipEnded() bool {
	return rt.owner != nil && rt.owner.ended()
}

func (rt *managedRuntime) publishReady(writer *managedWriter, url string, reloader *hotreload.Reloader) (bool, error) {
	emit := func() error {
		return writer.ready(url, managedReloadSnapshotFor(reloader, rt.opts.EnableReload))
	}
	if rt.owner == nil {
		return true, emit()
	}
	return rt.owner.publish(emit)
}

// serve runs the managed preview: resolve the target, start the watcher, bind
// the real loopback listener, publish ready with the actual URL, and shut down
// within the cleanup deadline when the owner is lost.
func (rt *managedRuntime) serve(ctx context.Context, writer *managedWriter) error {
	if ctx.Err() != nil || rt.ownershipEnded() {
		return nil
	}

	target, err := resolveServeTarget(rt.opts.Target)
	if err != nil {
		return rt.startupFailure(ctx, writer, "target-unavailable", fmt.Errorf("resolve target: %w", err))
	}
	if err := verifyTargetAccess(target); err != nil {
		return rt.startupFailure(ctx, writer, "target-unavailable", fmt.Errorf("access target: %w", err))
	}

	server := NewServerWithOptions(ServerOptions{
		BoundingBox:  rt.opts.BoundingBox,
		Browser:      false,
		EnableReload: rt.opts.EnableReload,
		Recursive:    rt.opts.Recursive,
		Parser:       NewParser(),
	})

	var reloader *hotreload.Reloader
	if rt.opts.EnableReload {
		reloader = hotreload.New(filepath.Clean(target.rootDir), rt.opts.Recursive)
		reloader.Upgrader.CheckOrigin = func(*http.Request) bool { return true }
	}
	stopReloader := func() {
		if reloader != nil {
			reloader.Stop()
		}
	}

	handler := server.newHandlerForTarget(target)
	if reloader != nil {
		handler = reloader.Handle(handler)
	}

	listener, port, err := listenOn("127.0.0.1", 0, true)
	if err != nil {
		stopReloader()
		return rt.startupFailure(ctx, writer, "listen-failed", fmt.Errorf("bind managed preview: %w", err))
	}

	initialPath, err := initialPathForTarget(target, rt.opts.Recursive)
	if err != nil {
		_ = listener.Close()
		stopReloader()
		return rt.startupFailure(ctx, writer, "target-unavailable", fmt.Errorf("resolve initial preview path: %w", err))
	}

	previewURL := fmt.Sprintf("http://127.0.0.1:%d%s", port, initialPath)
	if ctx.Err() != nil || rt.ownershipEnded() {
		_ = listener.Close()
		stopReloader()
		return nil
	}

	httpServer := &http.Server{Handler: newManagedHandler(rt.opts.Generation, handler)}
	serveErr := make(chan error, 1)
	go func() { serveErr <- httpServer.Serve(listener) }()

	published, publishErr := rt.publishReady(writer, previewURL, reloader)
	if publishErr != nil {
		_ = httpServer.Close()
		_ = listener.Close()
		<-serveErr
		stopReloader()
		return fmt.Errorf("publish ready: %w", publishErr)
	}
	if !published {
		_ = httpServer.Close()
		_ = listener.Close()
		<-serveErr
		stopReloader()
		return nil
	}

	select {
	case err := <-serveErr:
		stopReloader()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			_ = writer.fatal("serve-failed", err.Error())
			return fmt.Errorf("managed preview server: %w", err)
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), managedShutdownTimeout)
		shutdownErr := httpServer.Shutdown(shutdownCtx)
		cancel()
		if shutdownErr != nil {
			_ = httpServer.Close()
		}
		_ = listener.Close()
		<-serveErr
		stopReloader()
		server.closeInitializedPDF()
		return nil
	}
}
