package vergit

// The tree's state (Tree): the contents read against HEAD through a fresh index, the repository's own filters off,
// so neither the index's stat times nor its flags nor a filter keep a change from the result's dirty (V2's M1); every
// submodule read the same way, and judged for the findings that hide a change (V2's S1).

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// plainStatus is git status as a person runs it in the checkout.
func plainStatus(t *testing.T, root string) string {
	t.Helper()
	cmd := exec.Command("git", "-c", "core.autocrlf=false", "status", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	must(t, err)
	return strings.TrimSpace(string(out))
}

// aged sets a file's times an hour back, so the index git writes next is newer than the file (not racily clean) and
// its stat times are trusted from then on.
func aged(t *testing.T, root, rel string) {
	t.Helper()
	past := time.Now().Add(-time.Hour)
	must(t, os.Chtimes(filepath.Join(root, filepath.FromSlash(rel)), past, past))
}

// tree reads the tree and fails on an error.
func tree(t *testing.T, root string) TreeState {
	t.Helper()
	st, err := Tree(Input{Root: root})
	if err != nil {
		t.Fatalf("Tree: %v", err)
	}
	return st
}

// hiddenPaths are a state's hidden paths, joined by ", ".
func hiddenPaths(st TreeState) string {
	var p []string
	for _, h := range st.Hidden {
		p = append(p, h.Path)
	}
	return strings.Join(p, ", ")
}

// filterTrick hides a same-size edit of a.txt and pkg/a.go behind a filter in the repository's own configuration
// whose clean output is HEAD's copy, and refreshes the index while it is on, so the index holds the edited files' stat
// times; with gone, the filter and its attributes are then taken out, leaving no trace in .git/config or info/.
func filterTrick(t *testing.T, root string, gone bool) {
	t.Helper()
	git(t, root, "config", "filter.h.clean", "git show HEAD:%f")
	write(t, root, ".git/info/attributes", "a.txt filter=h\npkg/a.go filter=h\n")
	write(t, root, "a.txt", "A\n")
	write(t, root, "pkg/a.go", "package PKG\n")
	aged(t, root, "a.txt")
	aged(t, root, "pkg/a.go")
	_ = plainStatus(t, root) // refreshes the index while the filter makes both read as HEAD's
	if gone {
		git(t, root, "config", "--unset", "filter.h.clean")
		must(t, os.Remove(filepath.Join(root, ".git", "info", "attributes")))
	}
	if s := plainStatus(t, root); s != "" {
		t.Fatalf("plain git status shows the change, so the fixture proves nothing: %s", s)
	}
}

func TestTreeReadsContents(t *testing.T) {
	isolate(t)
	t.Run("clean", func(t *testing.T) {
		root, _ := repo(t)
		write(t, root, "build/out.bin", "x")
		if st := tree(t, root); st.Dirty || len(st.Shown) != 0 || len(st.Hidden) != 0 {
			t.Errorf("%+v", st)
		}
	})
	t.Run("a change git status shows", func(t *testing.T) {
		root, _ := repo(t)
		write(t, root, "a.txt", "changed\n")
		if st := tree(t, root); !st.Dirty || strings.Join(st.Shown, ",") != "a.txt" || len(st.Hidden) != 0 {
			t.Errorf("%+v", st)
		}
	})
	t.Run("stat times refreshed under a filter since taken out", func(t *testing.T) {
		root, _ := repo(t)
		filterTrick(t, root, true)
		st := tree(t, root)
		if !st.Dirty || len(st.Shown) != 0 || hiddenPaths(st) != "a.txt, pkg/a.go" ||
			st.Hidden[0].Why != "its contents differ from HEAD's, though git status shows no change" {
			t.Errorf("%+v", st)
		}
		if d, err := Dirty(Input{Root: root}); !d || err != nil {
			t.Errorf("Dirty = %v, %v", d, err)
		}
	})
	t.Run("a filter still on whose clean output is HEAD's copy", func(t *testing.T) {
		root, _ := repo(t)
		filterTrick(t, root, false)
		if st := tree(t, root); !st.Dirty || hiddenPaths(st) != "a.txt, pkg/a.go" {
			t.Errorf("%+v", st)
		}
	})
	t.Run("a filter in an included configuration file", func(t *testing.T) {
		root, _ := repo(t)
		write(t, root, ".git/more.config", "[filter \"h\"]\n\tclean = git show HEAD:%f\n")
		git(t, root, "config", "include.path", "more.config")
		write(t, root, ".git/info/attributes", "a.txt filter=h\n")
		write(t, root, "a.txt", "A\n")
		aged(t, root, "a.txt")
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain git status shows the change: %s", s)
		}
		if st := tree(t, root); !st.Dirty || hiddenPaths(st) != "a.txt" {
			t.Errorf("%+v", st)
		}
	})
	t.Run("an honest filter: files that differ from their blobs count as dirty (the known limit)", func(t *testing.T) {
		root, _ := repo(t)
		// A clean filter that writes the blob in capitals, as Git LFS writes a pointer: the working file differs from
		// its blob, which only the filter reconciles.
		git(t, root, "config", "filter.up.clean", "tr a-z A-Z")
		write(t, root, ".gitattributes", "*.up filter=up\n")
		write(t, root, "c.up", "lower\n")
		git(t, root, "add", ".")
		git(t, root, "commit", "-q", "-m", "a filtered file")
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain git status: %s", s)
		}
		if st := tree(t, root); !st.Dirty || hiddenPaths(st) != "c.up" {
			t.Errorf("%+v", st)
		}
	})
	t.Run("line ends git converts are no change", func(t *testing.T) {
		root, _ := repo(t)
		write(t, root, ".gitattributes", "*.txt text eol=crlf\n")
		git(t, root, "add", ".")
		git(t, root, "commit", "-q", "-m", "crlf")
		must(t, os.Remove(filepath.Join(root, "a.txt")))
		git(t, root, "checkout", "--", "a.txt")
		if raw, err := os.ReadFile(filepath.Join(root, "a.txt")); err != nil || string(raw) != "a\r\n" {
			t.Fatalf("a.txt is %q, %v", raw, err)
		}
		if st := tree(t, root); st.Dirty {
			t.Errorf("%+v", st)
		}
	})
	t.Run("an assume-unchanged entry and a skip-worktree one", func(t *testing.T) {
		root, _ := repo(t)
		git(t, root, "update-index", "--assume-unchanged", "a.txt")
		git(t, root, "update-index", "--skip-worktree", "b.txt")
		write(t, root, "a.txt", "changed behind git's back\n")
		write(t, root, "b.txt", "changed too\n")
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain git status: %s", s)
		}
		if st := tree(t, root); !st.Dirty || hiddenPaths(st) != "a.txt, b.txt" {
			t.Errorf("%+v", st)
		}
	})
	t.Run("a staged change shows, a staged change undone in the tree too", func(t *testing.T) {
		root, _ := repo(t)
		write(t, root, "a.txt", "staged\n")
		git(t, root, "add", "a.txt")
		write(t, root, "a.txt", "a\n")
		if st := tree(t, root); !st.Dirty || len(st.Hidden) != 0 || strings.Join(st.Shown, ",") != "a.txt" {
			t.Errorf("%+v", st)
		}
	})
	t.Run("a repository with no commit", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "r")
		must(t, os.MkdirAll(root, 0o755))
		git(t, root, "init", "-q")
		if st := tree(t, root); st.Dirty {
			t.Errorf("empty: %+v", st)
		}
		write(t, root, "a.txt", "a\n")
		if st := tree(t, root); !st.Dirty || len(st.Hidden) != 0 {
			t.Errorf("a file: %+v", st)
		}
	})
	t.Run("the fresh index leaves the checkout's own index as it was", func(t *testing.T) {
		root, _ := repo(t)
		git(t, root, "update-index", "--assume-unchanged", "a.txt")
		before, err := os.ReadFile(filepath.Join(root, ".git", "index"))
		must(t, err)
		tree(t, root)
		after, err := os.ReadFile(filepath.Join(root, ".git", "index"))
		must(t, err)
		if string(before) != string(after) {
			t.Error("the checkout's index changed")
		}
	})
}

// submoduleRepo makes a repository with a submodule at lib (its own repository beside it, lib.txt committed) and gives
// the top checkout's folder.
func submoduleRepo(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	must(t, os.MkdirAll(sub, 0o755))
	git(t, sub, "init", "-q")
	write(t, sub, "lib.txt", "two\n")
	git(t, sub, "add", ".")
	git(t, sub, "commit", "-q", "-m", "lib")
	root := filepath.Join(tmp, "top")
	must(t, os.MkdirAll(root, 0o755))
	git(t, root, "init", "-q")
	write(t, root, "a.txt", "a\n")
	git(t, root, "-c", "protocol.file.allow=always", "submodule", "add", "-q", filepath.ToSlash(sub), "lib")
	git(t, root, "add", ".")
	git(t, root, "commit", "-q", "-m", "top")
	return root
}

func TestTreeReadsSubmodules(t *testing.T) {
	isolate(t)
	t.Run("clean", func(t *testing.T) {
		root := submoduleRepo(t)
		if st := tree(t, root); st.Dirty {
			t.Errorf("%+v", st)
		}
	})
	t.Run("a change in a submodule git status shows", func(t *testing.T) {
		root := submoduleRepo(t)
		write(t, root, "lib/lib.txt", "three\n")
		if st := tree(t, root); !st.Dirty || strings.Join(st.Shown, ",") != "lib" {
			t.Errorf("%+v", st)
		}
	})
	t.Run("a change a submodule's own assume-unchanged entry hides", func(t *testing.T) {
		root := submoduleRepo(t)
		write(t, root, "lib/lib.txt", "six\n")
		git(t, filepath.Join(root, "lib"), "update-index", "--assume-unchanged", "lib.txt")
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain git status: %s", s)
		}
		st := tree(t, root)
		if !st.Dirty || hiddenPaths(st) != "lib/lib.txt" || !strings.HasSuffix(st.Hidden[0].Why, "(inside the submodule lib)") {
			t.Errorf("%+v", st)
		}
		o := judge(t, Input{Root: root})
		if n, ok := count(t, o, "assume-unchanged"); n != 1 || !ok || !o.Hides {
			t.Errorf("assume-unchanged %d %v, hides %v: %s", n, ok, o.Hides, o.Reason)
		}
		has(t, o, "assume-unchanged 1: lib/lib.txt")
	})
	t.Run("a file only a submodule's info/exclude ignores", func(t *testing.T) {
		root := submoduleRepo(t)
		gitDir := git(t, filepath.Join(root, "lib"), "rev-parse", "--absolute-git-dir")
		must(t, os.MkdirAll(filepath.Join(gitDir, "info"), 0o755))
		must(t, os.WriteFile(filepath.Join(gitDir, "info", "exclude"), []byte("notes.txt\n"), 0o644))
		write(t, root, "lib/notes.txt", "read by a test\n")
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain git status: %s", s)
		}
		o := judge(t, Input{Root: root})
		if n, _ := count(t, o, "exclude-hidden"); n != 1 || !o.Hides {
			t.Errorf("exclude-hidden %d, hides %v: %s", n, o.Hides, o.Reason)
		}
		has(t, o, "exclude-hidden 1: lib/notes.txt (by info/exclude)")
	})
	t.Run("a file in a submodule's folder that is not checked out", func(t *testing.T) {
		root := submoduleRepo(t)
		git(t, root, "submodule", "deinit", "-q", "-f", "lib")
		write(t, root, "lib/lib.go", "package lib\n")
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain git status: %s", s)
		}
		st := tree(t, root)
		if !st.Dirty || hiddenPaths(st) != "lib/lib.go" || !strings.Contains(st.Hidden[0].Why, "not checked out") {
			t.Errorf("%+v", st)
		}
	})
	t.Run("a submodule whose .git cannot be read: git status itself fails, which counts as dirty", func(t *testing.T) {
		root := submoduleRepo(t)
		write(t, root, "lib/.git", "gitdir: ../nowhere\n")
		st, err := Tree(Input{Root: root})
		if err == nil || !st.Dirty {
			t.Errorf("%+v, %v", st, err)
		}
		if d, err := Dirty(Input{Root: root}); !d || err == nil {
			t.Errorf("Dirty = %v, %v", d, err)
		}
	})
}

// setTemp points the system's temporary folder at dir, as os.TempDir reads it on each system.
func setTemp(t *testing.T, dir string) {
	t.Helper()
	for _, k := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(k, dir)
	}
}

// The fresh index's temporary folder (V2b's S7): removed after the read, each file through workspace.RemoveFile's
// retry; an error making it says why in words and names no folder (on Windows, a path with backslashes and the
// user's own folder).
func TestTreeTemporaryFolder(t *testing.T) {
	isolate(t)
	root, _ := repo(t)
	tmp := t.TempDir()
	setTemp(t, tmp)
	if st := tree(t, root); st.Dirty {
		t.Fatalf("%+v", st)
	}
	if left, err := os.ReadDir(tmp); err != nil || len(left) != 0 {
		t.Errorf("the read left %v behind (%v)", left, err)
	}
	setTemp(t, filepath.Join(tmp, "gone"))
	st, err := Tree(Input{Root: root})
	if want := "no temporary folder for a fresh index: it is not there"; err == nil || err.Error() != want || !st.Dirty {
		t.Errorf("Tree = %+v, %v; want dirty and %q", st, err, want)
	}
}
