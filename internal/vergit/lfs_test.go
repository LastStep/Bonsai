package vergit

// Files under Git LFS (BUG-054, lfs.go): a file HEAD's committed attributes put under the filter lfs, whose bytes are
// those HEAD's pointer names, reads as unchanged; any other byte, an executable bit, a pointer this package does not
// read, or an attribute no commit shows keeps it changed. No test needs git-lfs: the pointer is written by git
// hash-object, and a stand-in clean filter in the repository's configuration (git cat-file of that pointer) makes git
// add store it, as git-lfs's would.

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// lfsAttributes is the line git lfs track writes into .gitattributes.
const lfsAttributes = "*.png filter=lfs diff=lfs merge=lfs -text\n"

// logoBytes is the real file a pointer names in these tests: 3000 bytes, not text.
var logoBytes = func() string {
	b := make([]byte, 3000)
	for i := range b {
		b[i] = byte(i*7 + i/13)
	}
	return string(b)
}()

// pointerOf is the Git LFS pointer of content, as git-lfs writes it.
func pointerOf(content string) string {
	return fmt.Sprintf("version https://git-lfs.github.com/spec/v1\noid sha256:%x\nsize %d\n", sha256.Sum256([]byte(content)), len(content))
}

// storeBlob writes text as a blob into the repository at root, with no filter, and gives its id.
func storeBlob(t *testing.T, root, text string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "blob")
	must(t, os.WriteFile(f, []byte(text), 0o644))
	return git(t, root, "hash-object", "-w", "--no-filters", filepath.ToSlash(f))
}

// lfsAdd stages rel holding content as Git LFS does: the folder holds content, the index the blob stored (pointer,
// usually content's pointer), through a stand-in clean filter lfs in the repository's configuration that gives that
// blob, which is taken out again unless keep (as with git lfs install, which writes the user's global configuration).
// The file is aged first, so the index holds its stat times from then on.
func lfsAdd(t *testing.T, root, rel, content, pointer string, keep bool) {
	t.Helper()
	id := storeBlob(t, root, pointer)
	git(t, root, "config", "filter.lfs.clean", "git cat-file blob "+id)
	write(t, root, rel, content)
	aged(t, root, rel)
	git(t, root, "add", "--", rel)
	if !keep {
		git(t, root, "config", "--unset", "filter.lfs.clean")
	}
}

// plant commits the blob pointer at rel as plain text, through the index (no attribute or filter involved), and puts
// content in the folder.
func plant(t *testing.T, root, rel, content, pointer string) {
	t.Helper()
	id := storeBlob(t, root, pointer)
	git(t, root, "update-index", "--add", "--cacheinfo", "100644,"+id+","+rel)
	git(t, root, "commit", "-q", "-m", "planted "+rel)
	write(t, root, rel, content)
}

// lfsRepo makes a repository whose commit holds .gitattributes (lfsAttributes), a.txt and logo.png under Git LFS, the
// folder holding logoBytes and plain git status showing nothing; keep keeps the stand-in filter in its configuration.
func lfsRepo(t *testing.T, keep bool) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "r")
	must(t, os.MkdirAll(root, 0o755))
	git(t, root, "init", "-q")
	write(t, root, ".gitattributes", lfsAttributes)
	write(t, root, "a.txt", "a\n")
	git(t, root, "add", ".")
	lfsAdd(t, root, "logo.png", logoBytes, pointerOf(logoBytes), keep)
	git(t, root, "commit", "-q", "-m", "lfs")
	if got := git(t, root, "cat-file", "blob", "HEAD:logo.png"); got+"\n" != pointerOf(logoBytes) {
		t.Fatalf("HEAD's logo.png is not its pointer: %q", got)
	}
	if s := plainStatus(t, root); s != "" {
		t.Fatalf("plain git status: %s", s)
	}
	return root
}

// cleanTree holds a read to clean, with the paths verified.
func cleanTree(t *testing.T, st TreeState, verified string) {
	t.Helper()
	if st.Dirty || len(st.Shown) != 0 || len(st.Hidden) != 0 || strings.Join(st.Verified, ", ") != verified {
		t.Errorf("want clean with %q verified, got %+v", verified, st)
	}
}

// dirtyWith holds a read to dirty, rel shown or hidden, and, when why is not "", rel hidden with a reason holding it.
func dirtyWith(t *testing.T, st TreeState, rel, why string) {
	t.Helper()
	verified := false
	for _, v := range st.Verified {
		verified = verified || v == rel
	}
	if !st.Dirty || verified {
		t.Errorf("want %s dirty, got %+v", rel, st)
		return
	}
	for _, h := range st.Hidden {
		if h.Path == rel && strings.Contains(h.Why, why) {
			return
		}
	}
	if why == "" {
		for _, s := range st.Shown {
			if s == rel {
				return
			}
		}
	}
	t.Errorf("want %s dirty (%q), got %+v", rel, why, st)
}

func TestLFSPointer(t *testing.T) {
	oid := strings.Repeat("ab", 32)
	good := "version https://git-lfs.github.com/spec/v1\noid sha256:" + oid + "\nsize 3000\n"
	for _, c := range []struct {
		name, blob string
		ok         bool
		size       int64
	}{
		{"the spec's form", good, true, 3000},
		{"size zero", "version https://git-lfs.github.com/spec/v1\noid sha256:" + oid + "\nsize 0\n", true, 0},
		{"no final line feed", strings.TrimSuffix(good, "\n"), false, 0},
		{"CRLF", strings.ReplaceAll(good, "\n", "\r\n"), false, 0},
		{"an extension line", "version https://git-lfs.github.com/spec/v1\next-0-foo sha256:" + oid + "\noid sha256:" + oid + "\nsize 3000\n", false, 0},
		{"the old hawser version", strings.Replace(good, "https://git-lfs.github.com/spec/v1", "http://git-media.io/v/2", 1), false, 0},
		{"keys out of order", "version https://git-lfs.github.com/spec/v1\nsize 3000\noid sha256:" + oid + "\n", false, 0},
		{"upper-case hex", strings.Replace(good, oid, strings.ToUpper(oid), 1), false, 0},
		{"a short oid", strings.Replace(good, oid, oid[:62], 1), false, 0},
		{"another hash", strings.Replace(good, "sha256:", "sha512:", 1), false, 0},
		{"a negative size", strings.Replace(good, "size 3000", "size -1", 1), false, 0},
		{"a leading zero", strings.Replace(good, "size 3000", "size 03000", 1), false, 0},
		{"a size past int64", strings.Replace(good, "size 3000", "size 99999999999999999999", 1), false, 0},
		{"no size", strings.Replace(good, "size 3000\n", "", 1), false, 0},
		{"a line more", good + "x y\n", false, 0},
		{"a blank line more", good + "\n", false, 0},
		{"a space at an end", strings.Replace(good, "size 3000", "size 3000 ", 1), false, 0},
		{"1024 bytes", good + strings.Repeat("x", 1024-len(good)), false, 0},
		{"not a pointer", "hello\n", false, 0},
		{"empty", "", false, 0},
	} {
		gotOID, size, ok := lfsPointer([]byte(c.blob))
		if ok != c.ok || (ok && (gotOID != oid || size != c.size)) {
			t.Errorf("%s: %q %d %v", c.name, gotOID, size, ok)
		}
	}
}

func TestTreeReadsLFS(t *testing.T) {
	isolate(t)
	t.Run("committed and untouched, the filter in the repository's configuration: clean", func(t *testing.T) {
		root := lfsRepo(t, true)
		cleanTree(t, tree(t, root), "logo.png")
	})
	t.Run("committed and untouched, the filter only where these reads do not look: clean", func(t *testing.T) {
		root := lfsRepo(t, false)
		cleanTree(t, tree(t, root), "logo.png")
	})
	t.Run("touched, so git status reads its bytes against the pointer: clean", func(t *testing.T) {
		root := lfsRepo(t, false)
		write(t, root, "logo.png", logoBytes)
		if s := plainStatus(t, root); s != "M logo.png" {
			t.Fatalf("plain git status: %q", s)
		}
		cleanTree(t, tree(t, root), "logo.png")
	})
	t.Run("just added through update-index, no stat times kept: clean", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "r")
		must(t, os.MkdirAll(root, 0o755))
		git(t, root, "init", "-q")
		write(t, root, ".gitattributes", lfsAttributes)
		git(t, root, "add", ".")
		id := storeBlob(t, root, pointerOf(logoBytes))
		git(t, root, "update-index", "--add", "--cacheinfo", "100644,"+id+",img/logo.png")
		git(t, root, "commit", "-q", "-m", "lfs")
		write(t, root, "img/logo.png", logoBytes)
		cleanTree(t, tree(t, root), "img/logo.png")
	})
	t.Run("the folder holding the pointer itself (not smudged): clean, as git reads it", func(t *testing.T) {
		root := lfsRepo(t, false)
		write(t, root, "logo.png", pointerOf(logoBytes))
		cleanTree(t, tree(t, root), "logo.png")
	})
	for _, c := range []struct{ name, body, why string }{
		{"edited, the same size", strings.Repeat("x", len(logoBytes)), "under Git LFS, its bytes differ from those HEAD's pointer names"},
		{"edited, another size", logoBytes + "more", ""}, // git status shows a file whose size changed, filter or not
		{"emptied", "", ""},
	} {
		t.Run(c.name+", the filter in the configuration giving the pointer whatever it reads: dirty, named", func(t *testing.T) {
			root := lfsRepo(t, true)
			write(t, root, "logo.png", c.body)
			aged(t, root, "logo.png")
			dirtyWith(t, tree(t, root), "logo.png", c.why)
		})
		t.Run(c.name+", no filter: dirty, shown", func(t *testing.T) {
			root := lfsRepo(t, false)
			write(t, root, "logo.png", c.body)
			dirtyWith(t, tree(t, root), "logo.png", "")
		})
	}
	t.Run("a staged change: dirty", func(t *testing.T) {
		root := lfsRepo(t, false)
		other := strings.Repeat("y", 100)
		lfsAdd(t, root, "logo.png", other, pointerOf(other), false)
		st := tree(t, root)
		if !st.Dirty || strings.Join(st.Shown, ",") != "logo.png" || len(st.Verified) != 0 {
			t.Errorf("%+v", st)
		}
	})
	t.Run("deleted: dirty", func(t *testing.T) {
		root := lfsRepo(t, false)
		must(t, os.Remove(filepath.Join(root, "logo.png")))
		dirtyWith(t, tree(t, root), "logo.png", "")
	})
	t.Run("replaced by a folder: dirty", func(t *testing.T) {
		root := lfsRepo(t, false)
		must(t, os.Remove(filepath.Join(root, "logo.png")))
		write(t, root, "logo.png/x", logoBytes)
		dirtyWith(t, tree(t, root), "logo.png", "")
	})
	t.Run("replaced by a link to its real bytes: dirty", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("making a link on Windows needs a privilege a test cannot count on, and Git for Windows checks links out as files")
		}
		root := lfsRepo(t, false)
		real := filepath.Join(t.TempDir(), "real.png")
		must(t, os.WriteFile(real, []byte(logoBytes), 0o644))
		must(t, os.Remove(filepath.Join(root, "logo.png")))
		must(t, os.Symlink(real, filepath.Join(root, "logo.png")))
		dirtyWith(t, tree(t, root), "logo.png", "")
	})
	t.Run("made executable, its bytes unchanged: dirty", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows keeps no executable bit, so git there cannot see one either")
		}
		root := lfsRepo(t, false)
		must(t, os.Chmod(filepath.Join(root, "logo.png"), 0o755))
		dirtyWith(t, tree(t, root), "logo.png", "")
	})
	t.Run("HEAD's blob under filter=lfs is no pointer: dirty", func(t *testing.T) {
		root := lfsRepo(t, false)
		lfsAdd(t, root, "plain.png", "not lfs\n", "not lfs either\n", false)
		git(t, root, "commit", "-q", "-m", "plain")
		dirtyWith(t, tree(t, root), "plain.png", "HEAD's blob is no Git LFS pointer this package reads")
	})
	for _, c := range []struct{ name, pointer string }{
		{"a pointer with an extension line", strings.Replace(pointerOf(logoBytes), "\noid", "\next-0-foo sha256:"+strings.Repeat("0", 64)+"\noid", 1)},
		{"a pointer in upper-case hex", strings.Replace(pointerOf(logoBytes), fmt.Sprintf("%x", sha256.Sum256([]byte(logoBytes))),
			strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256([]byte(logoBytes)))), 1)},
	} {
		t.Run(c.name+": dirty", func(t *testing.T) {
			root := lfsRepo(t, false)
			plant(t, root, "odd.png", logoBytes, c.pointer)
			dirtyWith(t, tree(t, root), "odd.png", "")
		})
	}
	t.Run("a pointer committed at a path no attribute puts under Git LFS: dirty", func(t *testing.T) {
		root := lfsRepo(t, false)
		plant(t, root, "data.bin", logoBytes, pointerOf(logoBytes))
		dirtyWith(t, tree(t, root), "data.bin", "")
	})
	t.Run("in a SHA-256 repository: clean; edited: dirty", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "r")
		must(t, os.MkdirAll(root, 0o755))
		if err := runGit(root, "init", "-q", "--object-format=sha256"); err != nil {
			t.Skip("this git makes no SHA-256 repository: " + err.Error())
		}
		write(t, root, ".gitattributes", lfsAttributes)
		git(t, root, "add", ".")
		lfsAdd(t, root, "logo.png", logoBytes, pointerOf(logoBytes), false)
		git(t, root, "commit", "-q", "-m", "lfs")
		cleanTree(t, tree(t, root), "logo.png")
		write(t, root, "logo.png", strings.Repeat("z", len(logoBytes)))
		dirtyWith(t, tree(t, root), "logo.png", "")
	})
}

// runGit runs git in dir and gives its error, for a step a test may skip on.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, firstLine(string(out)))
	}
	return nil
}

// TestLFSAttributesFromHEADOnly: which files are tried comes from HEAD's committed attributes only (the plan review's
// attack): a pointer committed as plain text at a path HEAD's .gitattributes does not put under Git LFS, the real bytes
// in the folder, and filter=lfs added where no commit shows it, reads dirty.
func TestLFSAttributesFromHEADOnly(t *testing.T) {
	isolate(t)
	// planted commits a pointer as plain text at secret.bin (no attribute of HEAD's names it) and puts its real bytes
	// in the folder.
	planted := func(t *testing.T) string {
		root := lfsRepo(t, false)
		plant(t, root, "secret.bin", logoBytes, pointerOf(logoBytes))
		return root
	}
	t.Run("filter=lfs in info/attributes", func(t *testing.T) {
		root := planted(t)
		write(t, root, ".git/info/attributes", "secret.bin filter=lfs\n")
		dirtyWith(t, tree(t, root), "secret.bin", "")
	})
	t.Run("filter=lfs in an ignored .gitattributes no commit holds", func(t *testing.T) {
		root := planted(t)
		write(t, root, ".git/info/exclude", ".gitattributes\n")
		git(t, root, "rm", "-q", "--cached", ".gitattributes")
		git(t, root, "commit", "-q", "-m", "no attributes")
		write(t, root, ".gitattributes", lfsAttributes+"secret.bin filter=lfs\n")
		st := tree(t, root)
		dirtyWith(t, st, "secret.bin", "")
		dirtyWith(t, st, "logo.png", "")
	})
	t.Run("filter=lfs in an ignored .gitattributes in a folder", func(t *testing.T) {
		root := lfsRepo(t, false)
		write(t, root, ".gitignore", "deep/.gitattributes\n")
		git(t, root, "add", ".gitignore")
		plant(t, root, "deep/secret.bin", logoBytes, pointerOf(logoBytes))
		write(t, root, "deep/.gitattributes", "secret.bin filter=lfs\n")
		dirtyWith(t, tree(t, root), "deep/secret.bin", "")
	})
	t.Run("filter=lfs in a core.attributesFile the repository's configuration names", func(t *testing.T) {
		root := planted(t)
		attrs := filepath.Join(t.TempDir(), "attributes")
		must(t, os.WriteFile(attrs, []byte("secret.bin filter=lfs\n"), 0o644))
		git(t, root, "config", "core.attributesFile", filepath.ToSlash(attrs))
		dirtyWith(t, tree(t, root), "secret.bin", "")
	})
	t.Run("filter=lfs in a tree attr.tree names (git 2.42 and later)", func(t *testing.T) {
		root := planted(t)
		write(t, root, ".gitattributes", lfsAttributes+"secret.bin filter=lfs\n")
		git(t, root, "add", ".gitattributes")
		git(t, root, "commit", "-q", "-m", "attributes elsewhere")
		other := git(t, root, "rev-parse", "HEAD")
		git(t, root, "reset", "-q", "--hard", "HEAD~1")
		write(t, root, "secret.bin", logoBytes)
		git(t, root, "config", "attr.tree", other)
		dirtyWith(t, tree(t, root), "secret.bin", "")
	})
	t.Run("an honest Git LFS file while info/attributes holds a rule: dirty, the reason saying why", func(t *testing.T) {
		root := lfsRepo(t, false)
		write(t, root, ".git/info/attributes", "# a comment\n\n*.txt -diff\n")
		dirtyWith(t, tree(t, root), "logo.png", "not checked while the repository's info/attributes holds a rule")
	})
	t.Run("info/attributes holding only comments and blank lines: clean", func(t *testing.T) {
		root := lfsRepo(t, false)
		write(t, root, ".git/info/attributes", "# a comment\n\n   \n")
		cleanTree(t, tree(t, root), "logo.png")
	})
}

// TestOtherFilterStaysDirty: a file under another filter (a stand-in for git-crypt: its clean gives a blob other than
// its bytes) is not checked, and reads dirty with a reason naming the filter.
func TestOtherFilterStaysDirty(t *testing.T) {
	isolate(t)
	root := filepath.Join(t.TempDir(), "r")
	must(t, os.MkdirAll(root, 0o755))
	git(t, root, "init", "-q")
	write(t, root, ".gitattributes", "*.key filter=crypt\n")
	git(t, root, "add", ".")
	id := storeBlob(t, root, "\x00GITCRYPT\x00 ciphertext")
	git(t, root, "config", "filter.crypt.clean", "git cat-file blob "+id)
	write(t, root, "a.key", "the plain text\n")
	aged(t, root, "a.key")
	git(t, root, "add", "a.key")
	git(t, root, "commit", "-q", "-m", "crypt")
	if s := plainStatus(t, root); s != "" {
		t.Fatalf("plain git status: %s", s)
	}
	st := tree(t, root)
	if !st.Dirty || len(st.Hidden) != 1 || st.Hidden[0].Path != "a.key" || st.Hidden[0].Filter != "crypt" ||
		st.Hidden[0].Why != "under the filter crypt, which Bonsai cannot check; a checkout with it reads dirty" {
		t.Errorf("%+v", st)
	}
}

func TestLFSInASubmodule(t *testing.T) {
	isolate(t)
	make := func(t *testing.T) string {
		tmp := t.TempDir()
		sub := filepath.Join(tmp, "sub")
		must(t, os.MkdirAll(sub, 0o755))
		git(t, sub, "init", "-q")
		write(t, sub, ".gitattributes", lfsAttributes)
		git(t, sub, "add", ".")
		id := storeBlob(t, sub, pointerOf(logoBytes))
		git(t, sub, "update-index", "--add", "--cacheinfo", "100644,"+id+",logo.png")
		git(t, sub, "commit", "-q", "-m", "lib")
		root := filepath.Join(tmp, "top")
		must(t, os.MkdirAll(root, 0o755))
		git(t, root, "init", "-q")
		write(t, root, "a.txt", "a\n")
		git(t, root, "-c", "protocol.file.allow=always", "submodule", "add", "-q", filepath.ToSlash(sub), "lib")
		git(t, root, "add", ".")
		git(t, root, "commit", "-q", "-m", "top")
		// as git-lfs's smudge would have checked it out: its real bytes, the index refreshed through the filter
		lfsAdd(t, filepath.Join(root, "lib"), "logo.png", logoBytes, pointerOf(logoBytes), true)
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain git status: %s", s)
		}
		return root
	}
	t.Run("its real bytes: clean", func(t *testing.T) {
		root := make(t)
		cleanTree(t, tree(t, root), "lib/logo.png")
	})
	t.Run("edited, the filter there giving the pointer whatever it reads: dirty, named from the top", func(t *testing.T) {
		root := make(t)
		write(t, root, "lib/logo.png", strings.Repeat("q", len(logoBytes)))
		aged(t, root, "lib/logo.png")
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain git status: %s", s)
		}
		st := tree(t, root)
		if !st.Dirty || hiddenPaths(st) != "lib/logo.png" || !strings.HasSuffix(st.Hidden[0].Why, "(inside the submodule lib)") {
			t.Errorf("%+v", st)
		}
	})
}
