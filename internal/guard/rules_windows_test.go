package guard

import (
	"os"
	"path/filepath"
	"strings"
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
	// Its record names the path from the project, never from the user's folder (the temporary folder is in it).
	recs := records(t, dir, "s-r1.ndjson")
	text := recs[len(recs)-1].String("text")
	if user, err := os.UserHomeDir(); err == nil && strings.Contains(strings.ToLower(filepath.ToSlash(text)), strings.ToLower(filepath.ToSlash(user))) {
		t.Errorf("the record's text names the user's folder: %q", text)
	}
	if !strings.Contains(text, `"link/x.txt"`) {
		t.Errorf("the record's text: %q", text)
	}
}
