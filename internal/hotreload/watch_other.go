//go:build !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !linux

package hotreload

// dirCost is what watching a directory consumes: the remaining platforms hold
// one handle per watched directory.
func dirCost(entries int) int { return 1 }

// watchBudget falls back to a conservative ceiling; these platforms bound their
// handles far above what a documentation tree needs.
func watchBudget() int { return fallbackBudget }
