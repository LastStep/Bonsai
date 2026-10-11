package guard

import (
	"os"
	"path/filepath"
	"testing"
)

// ruleBadPath is bad-path's test (TestEveryRule), on Windows only (rules_other_test.go says why): a write through a
// junction whose target is gone is refused, with its reason and step, and recorded.
func ruleBadPath(t *testing.T) {
	dir := withTask(t, "running", "**")
	target := filepath.Join(t.TempDir(), "gone")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	mklinkJ(t, filepath.Join(dir, "link"), target)
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	wantRule(t, dir, writeOf(dir, "link/x.txt", "x\n"), RuleBadPath)
}
