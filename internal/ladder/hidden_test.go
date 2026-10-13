package ladder

// Changes hidden from git status, through the runner and the stop gate (V2's reproducers m1-filter-refresh and
// s1-submodule): the result's dirty reads the checkout's contents against HEAD, so a change git status does not show
// makes the result dirty and is named in the human lines; rung 0 judges each such path by the guard's function; the
// stop gate refuses the result.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/testpack"
)

// hideByFilterRefresh edits files to the same size behind a filter in the repository's own configuration whose clean
// output is HEAD's copy, refreshes the index while it is on (the index then holds the edited files' stat times), and
// takes the filter and its attributes out: plain git status shows nothing, and nothing in .git/config or info/ is left.
func hideByFilterRefresh(t *testing.T, root string, files map[string]string) {
	t.Helper()
	testpack.Git(t, root, "config", "filter.h.clean", "git show HEAD:%f")
	var attrs strings.Builder
	past := time.Now().Add(-time.Hour)
	for rel, body := range files {
		attrs.WriteString(rel + " filter=h\n")
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, past, past); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "info", "attributes"), []byte(attrs.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	testpack.Git(t, root, "status", "--porcelain")
	testpack.Git(t, root, "config", "--unset", "filter.h.clean")
	if err := os.Remove(filepath.Join(root, ".git", "info", "attributes")); err != nil {
		t.Fatal(err)
	}
	if s := testpack.Git(t, root, "status", "--porcelain"); s != "" {
		t.Fatalf("plain git status shows the change, so the fixture proves nothing: %q", s)
	}
}

func TestHiddenByAFilterRefresh(t *testing.T) {
	t.Run("a protected file and the code: rung 0 judges both, red", func(t *testing.T) {
		f := r0Project(t)
		hideByFilterRefresh(t, f.root, map[string]string{"assets/old/a.txt": "AN ASSET\n", "src/main.txt": "THE CODE\n"})
		a := f.climb("", Options{Task: "T-0901"})
		l := f.result("T-0901")
		if a.Exit != ExitRed || statuses(l) != "0:red 1:skipped" || !l.Git.Dirty {
			t.Fatalf("exit %d, %s, dirty %v\n%s", a.Exit, statuses(l), l.Git.Dirty, f.out.String())
		}
		if r := l.Rungs[0]; r.Reason == nil || !strings.Contains(*r.Reason, `"assets/old/a.txt" (protected)`) ||
			showOf(r.Captures, "checked") != "2" {
			t.Errorf("rung 0: %v, captures %s", r.Reason, showOf(r.Captures, "checked"))
		}
		for _, line := range []string{
			"  2 changed paths git status does not show, so the result counts the tree as dirty (each read by its contents against HEAD):\n",
			`    "assets/old/a.txt": its contents differ from HEAD's, though git status shows no change` + "\n",
			`    "src/main.txt": its contents differ from HEAD's, though git status shows no change` + "\n",
		} {
			if !strings.Contains(f.out.String(), line) {
				t.Errorf("the human lines lack %q:\n%s", line, f.out.String())
			}
		}
		if code, stderr := gate(t, f.root, "T-0901"); code != 2 {
			t.Errorf("the stop gate let the session end: %d %q", code, stderr)
		}
	})
	t.Run("the code alone: green, but dirty, which the stop gate refuses", func(t *testing.T) {
		f := r0Project(t)
		hideByFilterRefresh(t, f.root, map[string]string{"src/main.txt": "THE CODE\n"})
		a := f.climb("", Options{Task: "T-0901"})
		l := f.result("T-0901")
		if a.Exit != 0 || !l.Git.Dirty || showOf(l.Rungs[0].Captures, "checked") != "1" {
			t.Fatalf("exit %d, dirty %v, checked %s\n%s", a.Exit, l.Git.Dirty, showOf(l.Rungs[0].Captures, "checked"), f.out.String())
		}
		if code, stderr := gate(t, f.root, "T-0901"); code != 2 || !strings.Contains(stderr, "uncommitted changes") {
			t.Errorf("the stop gate: %d %q", code, stderr)
		}
	})
}

func TestHiddenInASubmodule(t *testing.T) {
	f := vgProject(t)
	sub := filepath.Join(f.tmp, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	testpack.Git(t, sub, "init", "-q")
	f.writeIn(sub, "lib.txt", "two\n")
	f.commitIn(sub, "lib")
	testpack.Git(t, f.root, "-c", "protocol.file.allow=always", "submodule", "add", "-q", filepath.ToSlash(sub), "lib")
	f.commit("a submodule")
	f.write("lib/lib.txt", "six\n")
	testpack.Git(t, filepath.Join(f.root, "lib"), "update-index", "--assume-unchanged", "lib.txt")
	if s := testpack.Git(t, f.root, "status", "--porcelain"); s != "" {
		t.Fatalf("plain git status shows the change: %q", s)
	}
	a := f.climb("", Options{Task: "T-0901"})
	l := f.result("T-0901")
	r := gitRung(t, f)
	if a.Exit != 0 || !l.Git.Dirty || showOf(r.Captures, "assume-unchanged") != "1" {
		t.Fatalf("exit %d, dirty %v, %s\n%s", a.Exit, l.Git.Dirty, showOf(r.Captures, "assume-unchanged"), f.out.String())
	}
	for _, s := range []string{"assume-unchanged 1: lib/lib.txt (hides a change from git",
		`"lib/lib.txt": its contents differ from HEAD's, though git status shows no change (inside the submodule lib)`} {
		if !strings.Contains(f.out.String(), s) {
			t.Errorf("the human lines lack %q:\n%s", s, f.out.String())
		}
	}
	if code, stderr := gate(t, f.root, "T-0901"); code != 2 || !strings.Contains(stderr, "uncommitted changes") {
		t.Errorf("the stop gate: %d %q", code, stderr)
	}
}
