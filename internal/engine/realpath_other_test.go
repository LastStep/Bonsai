//go:build !windows

package engine

// shortTemp is a new temporary folder: short (8.3) names are Windows' (realpath_windows_test.go spells the folder by
// its short name there), so elsewhere it is the folder as made.

import "testing"

func shortTemp(t *testing.T) string { return t.TempDir() }
