//go:build !windows

package guard

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// ruleBadPath is bad-path's test (TestEveryRule), off Windows (rules_windows_test.go has Windows' junction that leads
// nowhere): a Write through a loop of symbolic links, whose walk never ends, on every system; and on Linux a write into
// /proc (TestAPathIntoProcIsRefused has the rest). The refusal names the walk's own cause, never Windows.
func ruleBadPath(t *testing.T) {
	dir := withTask(t, "running", "**")
	away, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range [][2]string{{"b", "a"}, {"a", "b"}} {
		if err := os.Symlink(filepath.Join(away, l[0]), filepath.Join(away, l[1])); err != nil {
			t.Fatal(err)
		}
	}
	stderr := wantRule(t, dir, toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: filepath.Join(away, "a", "x.txt")},
		{Key: "content", Value: "x"}}), RuleBadPath)
	if strings.Contains(stderr, "Windows") || !strings.Contains(stderr, "a loop of links") {
		t.Errorf("a Write through a link loop: the refusal names Windows, or not the loop: %s", stderr)
	}
	if runtime.GOOS != "linux" {
		return
	}
	stderr = wantRule(t, dir, toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: "/proc/self/cwd/x.txt"},
		{Key: "content", Value: "x"}}), RuleBadPath)
	if strings.Contains(stderr, "Windows") || !strings.Contains(stderr, "/proc") {
		t.Errorf("a Write into /proc: the refusal names Windows, or not /proc: %s", stderr)
	}
}
