//go:build !windows

package engine

// FinalPath is Windows' own name for a path (final_windows.go). Elsewhere WalkLinks' symbolic links are all there is
// to follow, so it gives nothing.
func FinalPath(string) (string, bool, error) { return "", false, nil }
