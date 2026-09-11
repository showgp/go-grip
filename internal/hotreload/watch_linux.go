//go:build linux

package hotreload

import (
	"os"
	"strconv"
	"strings"
)

// dirCost is what watching a directory consumes. inotify holds one watch per
// directory and reports changes to the files inside it through that watch, so
// only the directory itself is counted.
func dirCost(entries int) int { return 1 }

// watchBudget is the per-user inotify watch allowance, which is the resource
// that actually runs out here: Add reports it as ENOSPC rather than EMFILE. The
// quota is shared with every other process of the user, so only half is claimed.
func watchBudget() int {
	raw, err := os.ReadFile("/proc/sys/fs/inotify/max_user_watches")
	if err != nil {
		return fallbackBudget
	}
	quota, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil || quota <= 0 {
		return fallbackBudget
	}
	return watchBudgetFromLimit(quota, quota/2)
}
