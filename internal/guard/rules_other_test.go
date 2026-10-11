//go:build !windows

package guard

import "testing"

// ruleBadPath is bad-path's test (TestEveryRule). The rule comes from Windows' own name for a path (final_windows.go:
// a part it opens but cannot name, a junction or link that does not resolve), which no other system gives, so it is
// proved on Windows (rules_windows_test.go, and TestJunctionLimits).
func ruleBadPath(t *testing.T) {
	t.Skip("bad-path comes only from Windows' own name for a path; rules_windows_test.go proves it on Windows")
}
