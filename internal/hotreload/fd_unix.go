//go:build unix

package hotreload

import (
	"errors"
	"syscall"
)

// isFdExhausted reports whether err means the watch resource is used up: the
// descriptor table on kqueue platforms (EMFILE/ENFILE), or the per-user inotify
// watch quota on Linux (ENOSPC).
func isFdExhausted(err error) bool {
	return errors.Is(err, syscall.EMFILE) ||
		errors.Is(err, syscall.ENFILE) ||
		errors.Is(err, syscall.ENOSPC)
}
