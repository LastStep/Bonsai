//go:build !linux && !windows

package textui

import "os"

// termSize reports whether f is a character device, as cmd/bonsai's isTerminal does on these systems (macOS
// builds, untested, spec §3), and no width: the layout takes Fixed.
func termSize(f *os.File) (bool, int) {
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0, 0
}

// enableColor reports that a terminal here draws colour codes as they are.
func enableColor(*os.File) bool { return true }
