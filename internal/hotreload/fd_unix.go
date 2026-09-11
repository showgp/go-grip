//go:build unix

package hotreload

import (
	"errors"
	"syscall"
)

// isFdExhausted reports whether err means the process or the system descriptor
// table is full.
//
// The Go runtime already raises the soft RLIMIT_NOFILE limit for the process
// (syscall.init), so the watcher only has to degrade gracefully when the
// descriptor budget runs out.
func isFdExhausted(err error) bool {
	return errors.Is(err, syscall.EMFILE) || errors.Is(err, syscall.ENFILE)
}
