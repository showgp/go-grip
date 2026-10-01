package hotreload

// fallbackBudget is used when the platform resource cannot be read.
const fallbackBudget = 65536

// watchBudgetFromLimit converts a platform resource limit into the watch budget
// by holding back reserve units for the rest of the process. The limit type
// varies by platform (BSD reports RLIMIT_NOFILE as int64, Darwin as uint64).
func watchBudgetFromLimit[T ~int | ~int64 | ~uint64](limit T, reserve int) int {
	if limit <= 0 {
		return fallbackBudget
	}
	usable := uint64(limit)
	const maxInt = uint64(^uint(0) >> 1)
	if usable > maxInt {
		return fallbackBudget
	}
	if budget := int(usable) - reserve; budget > 1 {
		return budget
	}
	return 1
}
