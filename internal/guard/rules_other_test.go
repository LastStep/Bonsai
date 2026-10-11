//go:build !windows

package guard

import (
	"runtime"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// ruleBadPath is bad-path's test (TestEveryRule). On Linux, a write into /proc (TestAPathIntoProcIsRefused has the
// rest); on Windows, a junction that leads nowhere (rules_windows_test.go). Elsewhere (macOS) neither exists, so it is
// proved on those two.
func ruleBadPath(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("bad-path comes from /proc on Linux and from Windows' own name for a path; neither exists here")
	}
	dir := withTask(t, "running", "**")
	wantRule(t, dir, toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: "/proc/self/cwd/x.txt"},
		{Key: "content", Value: "x"}}), RuleBadPath)
}
