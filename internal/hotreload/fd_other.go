//go:build !unix

package hotreload

// isFdExhausted always reports false on platforms without descriptor limits.
func isFdExhausted(err error) bool { return false }
