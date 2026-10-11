package guard

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// A recursive delete of a folder by its name is judged without reading its tree: rm -rf node_modules on a folder of
// about 10,000 entries with nested .git folders is allowed within the guard's own budget (Budget). The limit is
// that hard one, never this machine's speed (5.2's lesson); the times are logged for the run report.
func TestShellBigTreeWithinBudget(t *testing.T) {
	s := delFixture(t)
	nm := filepath.Join(s.own, "node_modules")
	for p := 0; p < 100; p++ {
		dir := filepath.Join(nm, fmt.Sprintf("pkg%03d", p), "lib")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for f := 0; f < 98; f++ {
			if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("f%02d.js", f)), nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if p%10 == 0 {
			git := filepath.Join(nm, fmt.Sprintf("pkg%03d", p), ".git")
			if err := os.MkdirAll(git, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(git, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	entries := 0
	if err := filepath.WalkDir(nm, func(string, fs.DirEntry, error) error { entries++; return nil }); err != nil {
		t.Fatal(err)
	}
	if entries < 10000 {
		t.Fatalf("node_modules holds %d entries, want about 10,000", entries)
	}
	for _, c := range []struct{ tool, line string }{
		{tB, "rm -rf node_modules"},
		{tB, "find node_modules -delete"},
		{tB, "rm -rf node_modules/pkg000/.git"},
		{tP, "Remove-Item -Recurse -Force node_modules"},
		{tP, `cmd /c "rd /s /q node_modules"`},
	} {
		start := time.Now()
		d := shellDecide(c.tool, c.line, s.own, s)
		took := time.Since(start)
		if !d.Allow {
			t.Errorf("%s %q on %d entries refused: %s", c.tool, c.line, entries, d.Why)
		}
		if took >= Budget {
			t.Errorf("%s %q on %d entries took %v, past the guard's budget of %v", c.tool, c.line, entries, took, Budget)
		}
		t.Logf("%s %q on %d entries: %v", c.tool, c.line, entries, took)
	}
}
