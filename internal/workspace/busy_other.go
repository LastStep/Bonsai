//go:build !windows

package workspace

// isBusy reports whether a rename failed because another process holds the file. Linux and macOS replace a file
// another process holds open, so no rename there is ever busy.
func isBusy(error) bool { return false }
