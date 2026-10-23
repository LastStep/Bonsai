package ladder

import (
	"path/filepath"
	"testing"
)

// T-5729: the header names the checkout relative to the folder the person runs the climb in, not in full.
func TestShownRoot(t *testing.T) {
	base := filepath.Join(string(filepath.Separator), "w", "proj")
	for _, c := range []struct{ root, cwd, want string }{
		{base, base, "."},
		{base, filepath.Join(base, "sub", "deep"), "../.."},
		{base, filepath.Join(string(filepath.Separator), "w", "proj-wt"), "../proj"},
		{base, filepath.Join(string(filepath.Separator), "elsewhere", "a", "b"), "/w/proj"},
		{base, "", "/w/proj"},
	} {
		if got := shownRoot(c.root, c.cwd); got != c.want {
			t.Errorf("shownRoot(%q, %q) = %q, want %q", c.root, c.cwd, got, c.want)
		}
	}
}
