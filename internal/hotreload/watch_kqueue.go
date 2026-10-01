//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package hotreload

import "syscall"

// fdReserve is the descriptor headroom left to the HTTP server, the renderer and
// anything else in the process. The watcher must not take the whole table: the
// server still has to accept connections and open the files it serves.
const fdReserve = 1024

// dirCost is what watching a directory consumes. fsnotify's kqueue backend opens
// one descriptor for the directory and one for every entry inside it, so a
// single directory holding hundreds of files costs hundreds of descriptors: the
// number of directories says nothing about the budget they need.
func dirCost(entries int) int { return 1 + entries }

// watchBudget is the descriptor allowance for the watch set. The Go runtime
// already raises the soft limit toward the hard limit, so the current soft limit
// is the descriptor table the process can use.
func watchBudget() int {
	var lim syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim); err != nil {
		return fallbackBudget
	}
	cur := lim.Cur
	return watchBudgetFromLimit(cur, fdReserve)
}
