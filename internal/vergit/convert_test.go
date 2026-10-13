package vergit

// The attributes that convert a file's bytes on their way into git (V2b's H1): a tracked file under ident or
// working-tree-encoding reads as unchanged only when its bytes, line ends read as LF, are HEAD's blob's, wherever the
// attribute is set.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixture's Go file as HEAD holds it, and the edit git's ident attribute reads as the same: the code between
// "$Id:" and "$" is dropped.
const (
	identHead = "package pkg\n\nfunc Sub(a, b int) int { return a /*$Id$*/ + 0*b }\n"
	identEdit = "package pkg\n\nfunc Sub(a, b int) int { return a /*$Id: */ - b /* $*/ + 0*b }\n"
)

// hiddenOnly holds a read to exactly one hidden path with its reason, nothing shown.
func hiddenOnly(t *testing.T, st TreeState, path, why string) {
	t.Helper()
	if !st.Dirty || len(st.Shown) != 0 || len(st.Hidden) != 1 || st.Hidden[0].Path != path || st.Hidden[0].Why != why {
		t.Errorf("want %s hidden (%s), got %+v", path, why, st)
	}
}

func TestConvertingAttributes(t *testing.T) {
	isolate(t)
	byIdent := "its bytes differ from HEAD's, though the ident attribute makes git read them as HEAD's"
	// commitSub commits pkg/sub.go as HEAD holds it, with more files.
	commitSub := func(t *testing.T, more map[string]string) string {
		t.Helper()
		root, _ := repo(t)
		write(t, root, "pkg/sub.go", identHead)
		for rel, body := range more {
			write(t, root, rel, body)
		}
		git(t, root, "add", ".")
		git(t, root, "commit", "-q", "-m", "a buggy Sub")
		return root
	}
	// hide edits pkg/sub.go under the ident attribute and stages it, so the index holds the edited file's times.
	hide := func(t *testing.T, root string) {
		t.Helper()
		write(t, root, "pkg/sub.go", identEdit)
		git(t, root, "add", "pkg/sub.go")
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain git status shows the change, so the fixture proves nothing: %s", s)
		}
	}
	t.Run("ident in info/attributes", func(t *testing.T) {
		root := commitSub(t, nil)
		write(t, root, ".git/info/attributes", "pkg/sub.go ident\n")
		hide(t, root)
		hiddenOnly(t, tree(t, root), "pkg/sub.go", byIdent)
		if d, err := Dirty(Input{Root: root}); !d || err != nil {
			t.Errorf("Dirty = %v, %v", d, err)
		}
	})
	t.Run("ident in a committed .gitattributes", func(t *testing.T) {
		root := commitSub(t, map[string]string{".gitattributes": "pkg/sub.go ident\n"})
		hide(t, root)
		hiddenOnly(t, tree(t, root), "pkg/sub.go", byIdent)
	})
	t.Run("ident in a .gitattributes a committed .gitignore ignores", func(t *testing.T) {
		root := commitSub(t, map[string]string{".gitignore": "build/\npkg/.gitattributes\n"})
		write(t, root, "pkg/.gitattributes", "sub.go ident\n")
		hide(t, root)
		hiddenOnly(t, tree(t, root), "pkg/sub.go", byIdent)
	})
	t.Run("working-tree-encoding in info/attributes", func(t *testing.T) {
		root := commitSub(t, map[string]string{"pkg/want.txt": "caf\xc3\xa9\n"})
		write(t, root, ".git/info/attributes", "pkg/want.txt working-tree-encoding=ISO-8859-1\n")
		write(t, root, "pkg/want.txt", "caf\xe9\n")
		git(t, root, "add", "pkg/want.txt")
		hiddenOnly(t, tree(t, root), "pkg/want.txt",
			"its bytes differ from HEAD's, though the working-tree-encoding attribute makes git read them as HEAD's")
	})
	t.Run("both attributes", func(t *testing.T) {
		root := commitSub(t, nil)
		write(t, root, ".git/info/attributes", "pkg/sub.go ident working-tree-encoding=ISO-8859-1\n")
		hide(t, root)
		hiddenOnly(t, tree(t, root), "pkg/sub.go",
			"its bytes differ from HEAD's, though the ident and working-tree-encoding attributes make git read them as HEAD's")
	})
	t.Run("HEAD's own bytes under the attributes read clean", func(t *testing.T) {
		root := commitSub(t, map[string]string{"pkg/want.txt": "caf\xc3\xa9\n"})
		write(t, root, ".git/info/attributes", "pkg/sub.go ident\npkg/want.txt working-tree-encoding=UTF-16\n")
		if st := tree(t, root); st.Dirty {
			t.Errorf("%+v", st)
		}
	})
	t.Run("HEAD's bytes with CRLF line ends read clean", func(t *testing.T) {
		root := commitSub(t, nil)
		write(t, root, ".git/info/attributes", "pkg/sub.go ident text\n")
		write(t, root, "pkg/sub.go", strings.ReplaceAll(identHead, "\n", "\r\n"))
		git(t, root, "add", "pkg/sub.go") // the index takes the file's size: git reads a size it did not record as a change
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain git status shows the line ends: %s", s)
		}
		if st := tree(t, root); st.Dirty {
			t.Errorf("%+v", st)
		}
	})
	t.Run("a keyword git expanded reads as changed (a known limit)", func(t *testing.T) {
		root := commitSub(t, map[string]string{".gitattributes": "pkg/sub.go ident\n"})
		must(t, os.Remove(filepath.Join(root, "pkg", "sub.go")))
		git(t, root, "checkout", "--", "pkg/sub.go")
		b, err := os.ReadFile(filepath.Join(root, "pkg", "sub.go"))
		must(t, err)
		if !strings.Contains(string(b), "$Id: ") || plainStatus(t, root) != "" {
			t.Fatalf("git did not expand the keyword: %q", b)
		}
		hiddenOnly(t, tree(t, root), "pkg/sub.go", byIdent)
	})
	t.Run("a change git status shows is not named again", func(t *testing.T) {
		root := commitSub(t, nil)
		write(t, root, ".git/info/attributes", "pkg/sub.go ident\n")
		write(t, root, "pkg/sub.go", "package pkg\n")
		if st := tree(t, root); !st.Dirty || strings.Join(st.Shown, ",") != "pkg/sub.go" || len(st.Hidden) != 0 {
			t.Errorf("%+v", st)
		}
	})
	t.Run("inside a submodule", func(t *testing.T) {
		root := commitSub(t, nil)
		sub := commitSub(t, nil)
		git(t, root, "-c", "protocol.file.allow=always", "submodule", "add", "-q", filepath.ToSlash(sub), "lib")
		git(t, root, "commit", "-q", "-m", "a submodule")
		common := strings.TrimSpace(git(t, filepath.Join(root, "lib"), "rev-parse", "--git-common-dir"))
		if !filepath.IsAbs(common) {
			common = filepath.Join(root, "lib", common)
		}
		must(t, os.MkdirAll(filepath.Join(common, "info"), 0o755))
		must(t, os.WriteFile(filepath.Join(common, "info", "attributes"), []byte("pkg/sub.go ident\n"), 0o644))
		write(t, root, "lib/pkg/sub.go", identEdit)
		git(t, filepath.Join(root, "lib"), "add", "pkg/sub.go")
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain git status shows the change: %s", s)
		}
		hiddenOnly(t, tree(t, root), "lib/pkg/sub.go", byIdent+" (inside the submodule lib)")
	})
}

// blobID is git's own object id, for both of git's hashes.
func TestBlobID(t *testing.T) {
	isolate(t)
	root, _ := repo(t)
	b, err := os.ReadFile(filepath.Join(root, "pkg", "a.go"))
	must(t, err)
	if got, want := blobID(b, 40), git(t, root, "hash-object", "pkg/a.go"); got != want {
		t.Errorf("SHA-1: %s, git: %s", got, want)
	}
	// git hash-object of "a\n" in a repository made with --object-format=sha256
	if got, want := blobID([]byte("a\n"), 64), "f8625e43f9e04f24291f77cdbe4c71b3c2a3b0003f60419b3ed06a058d766c8b"; got != want {
		t.Errorf("SHA-256: %s, git: %s", got, want)
	}
}
