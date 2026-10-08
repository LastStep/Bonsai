//go:build !windows

package guard

// finalPath is Windows' own name for a path (final_windows.go). Elsewhere resolve's symbolic links are all there is
// to follow, so it gives nothing.
func finalPath(string) (string, bool, error) { return "", false, nil }
