//go:build !unix

package hotreload

// isFdExhausted always reports false on platforms without watch quotas.
func isFdExhausted(err error) bool { return false }
