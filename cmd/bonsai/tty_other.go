//go:build !linux && !windows

package main

import "os"

// isTerminal reports whether f is a character device. On these systems (macOS builds, untested, spec §3) that
// takes /dev/null for a terminal too; a command there still never waits once --yes or --json is given.
func isTerminal(f *os.File) bool {
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}
