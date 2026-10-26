package ladder

// A checkout with Git LFS (BUG-054): a file HEAD's committed attributes put under Git LFS, whose bytes are those its
// committed pointer names, reads as unchanged through the runner: the result is clean, and rung 0 does not judge it,
// even under a protected path, whether git status shows it (stale stat times) or not. An edit still counts.

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/testpack"
)

// lfsLogo is the real file of these tests' pointer.
var lfsLogo = strings.Repeat("\x89PNG logo bytes ", 200)

// lfsPointerOf is the Git LFS pointer of content, as git-lfs writes it.
func lfsPointerOf(content string) string {
	return fmt.Sprintf("version https://git-lfs.github.com/spec/v1\noid sha256:%x\nsize %d\n", sha256.Sum256([]byte(content)), len(content))
}

// lfsFile commits rel (under a protected path in r0Project) as Git LFS keeps it: .gitattributes puts *.png under the
// filter lfs, HEAD holds the pointer, the folder its real bytes. With fresh, the index is refreshed through a stand-in
// clean filter (as git-lfs's own), taken out again, so git status shows nothing; else the pointer goes in by
// update-index, so the index keeps no stat times and git status shows the file changed.
func lfsFile(t *testing.T, f *fx, rel string, fresh bool) {
	t.Helper()
	f.write(".gitattributes", "*.png filter=lfs diff=lfs merge=lfs -text\n")
	f.commit("lfs")
	blob := filepath.Join(t.TempDir(), "pointer")
	if err := os.WriteFile(blob, []byte(lfsPointerOf(lfsLogo)), 0o644); err != nil {
		t.Fatal(err)
	}
	id := testpack.Git(t, f.root, "hash-object", "-w", "--no-filters", filepath.ToSlash(blob))
	if fresh {
		testpack.Git(t, f.root, "config", "filter.lfs.clean", "git cat-file blob "+id)
		f.write(rel, lfsLogo)
		past := time.Now().Add(-time.Hour)
		if err := os.Chtimes(filepath.Join(f.root, filepath.FromSlash(rel)), past, past); err != nil {
			t.Fatal(err)
		}
		testpack.Git(t, f.root, "add", "--", rel)
		testpack.Git(t, f.root, "config", "--unset", "filter.lfs.clean")
		testpack.Git(t, f.root, "commit", "-q", "-m", "a logo")
		if s := testpack.Git(t, f.root, "status", "--porcelain"); s != "" {
			t.Fatalf("plain git status: %q", s)
		}
		return
	}
	testpack.Git(t, f.root, "update-index", "--add", "--cacheinfo", "100644,"+id+","+rel)
	testpack.Git(t, f.root, "commit", "-q", "-m", "a logo")
	f.write(rel, lfsLogo)
	if s := testpack.Git(t, f.root, "status", "--porcelain"); s != "M "+rel {
		t.Fatalf("plain git status: %q", s)
	}
}

func TestLFSCheckoutClimbsClean(t *testing.T) {
	for _, fresh := range []bool{true, false} {
		t.Run(fmt.Sprintf("the index's stat times fresh %v: green and clean", fresh), func(t *testing.T) {
			f := r0Project(t)
			lfsFile(t, f, "assets/old/logo.png", fresh)
			a := f.climb("", Options{Task: "T-0901"})
			l := f.result("T-0901")
			if a.Exit != 0 || l.Git.Dirty || statuses(l) != "0:green 1:green" || showOf(l.Rungs[0].Captures, "checked") != "0" {
				t.Fatalf("exit %d, dirty %v, %s, checked %s\n%s", a.Exit, l.Git.Dirty, statuses(l),
					showOf(l.Rungs[0].Captures, "checked"), f.out.String())
			}
			if code, stderr := gate(t, f.root, "T-0901"); code != 0 {
				t.Errorf("the stop gate: %d %q", code, stderr)
			}
		})
	}
	t.Run("edited to the same size: rung 0 judges it, red, and the result is dirty", func(t *testing.T) {
		f := r0Project(t)
		lfsFile(t, f, "assets/old/logo.png", true)
		f.write("assets/old/logo.png", strings.Repeat("x", len(lfsLogo)))
		a := f.climb("", Options{Task: "T-0901"})
		l := f.result("T-0901")
		if a.Exit != ExitRed || !l.Git.Dirty || statuses(l) != "0:red 1:skipped" {
			t.Fatalf("exit %d, dirty %v, %s\n%s", a.Exit, l.Git.Dirty, statuses(l), f.out.String())
		}
		if r := l.Rungs[0]; r.Reason == nil || !strings.Contains(*r.Reason, `"assets/old/logo.png" (protected)`) {
			t.Errorf("rung 0: %v", r.Reason)
		}
	})
}

// TestFilteredFileNextStep: a file under a filter Bonsai cannot check reads dirty, and the human lines say so instead
// of the index's next step, which does nothing for it.
func TestFilteredFileNextStep(t *testing.T) {
	f := r0Project(t)
	f.write(".gitattributes", "*.key filter=crypt\n")
	f.commit("crypt")
	blob := filepath.Join(t.TempDir(), "cipher")
	if err := os.WriteFile(blob, []byte("\x00GITCRYPT\x00 ciphertext"), 0o644); err != nil {
		t.Fatal(err)
	}
	id := testpack.Git(t, f.root, "hash-object", "-w", "--no-filters", filepath.ToSlash(blob))
	testpack.Git(t, f.root, "config", "filter.crypt.clean", "git cat-file blob "+id)
	f.write("src/a.key", "the plain text\n")
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(f.root, "src", "a.key"), past, past); err != nil {
		t.Fatal(err)
	}
	f.commit("a key")
	if s := testpack.Git(t, f.root, "status", "--porcelain"); s != "" {
		t.Fatalf("plain git status: %q", s)
	}
	f.climb("", Options{Task: "T-0901"})
	l := f.result("T-0901")
	out := f.out.String()
	if !l.Git.Dirty || !strings.Contains(out, `"src/a.key": under the filter crypt, which Bonsai cannot check; a checkout with it reads dirty`) ||
		!strings.Contains(out, "a file under another filter (git-crypt and the like) cannot be checked by Bonsai") ||
		strings.Contains(out, "git rm -q --cached") {
		t.Errorf("dirty %v:\n%s", l.Git.Dirty, out)
	}
}
