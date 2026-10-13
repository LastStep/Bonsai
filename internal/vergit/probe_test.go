package vergit

// What the file system can hold, pinned (V2b's H2): a stale core.ignorecase, core.fileMode or core.symlinks in the
// repository's own configuration hides nothing from the reads where the file system can hold the difference, and is
// left as git has it where the file system cannot (Windows: no case twins, no executable bits, links checked out as
// files); and no hook the repository names runs during the contents' read.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The probe: on Linux every setting is pinned (a case-sensitive file system with modes and links); on Windows none
// (an NTFS folder is case-blind, keeps no executable bit, and Git for Windows checks links out as files). It leaves
// nothing behind; a folder it cannot write is taken at its strictest on Linux and macOS, as git has it on Windows.
func TestFileSystemProbe(t *testing.T) {
	dir := t.TempDir()
	got := strings.Join(probeFS(dir), " ")
	switch runtime.GOOS {
	case "linux":
		if got != "core.ignorecase=false core.fileMode=true core.symlinks=true" {
			t.Errorf("Linux: %q", got)
		}
	case "windows":
		if got != "" {
			t.Errorf("Windows (git's own core.ignorecase, core.fileMode and core.symlinks must stand): %q", got)
		}
	default:
		t.Logf("%s: %q", runtime.GOOS, got)
	}
	if left, err := os.ReadDir(dir); err != nil || len(left) != 0 {
		t.Errorf("the probe left %v (%v)", left, err)
	}
	strict := strings.Join(probeFS(filepath.Join(dir, "gone")), " ")
	if want := "core.ignorecase=false core.fileMode=true core.symlinks=true"; runtime.GOOS == "windows" && strict != "" ||
		runtime.GOOS != "windows" && strict != want {
		t.Errorf("a folder that cannot be probed: %q", strict)
	}
	if a, b := strings.Join(fsPins(dir), " "), strings.Join(fsPins(dir), " "); a != got || b != got {
		t.Errorf("fsPins: %q then %q, the probe %q", a, b, got)
	}
}

// Each stale setting, set in the repository's configuration after the checkout was made: plain git status shows
// nothing, the reads show the change. Each needs what the setting would hide, which Windows cannot hold (a case twin,
// an executable bit, a link), so each runs on the systems whose file system the probe finds holds it.
func TestStaleFileSystemSettings(t *testing.T) {
	isolate(t)
	pinned := " " + strings.Join(probeFS(t.TempDir()), " ") + " "
	t.Run("core.ignorecase=true hides an untracked case twin", func(t *testing.T) {
		if !strings.Contains(pinned, " core.ignorecase=false ") {
			t.Skip("needs a case-sensitive file system (not Windows, not macOS's default)")
		}
		root, _ := repo(t)
		write(t, root, "pkg/zz_sub.GO", "notes\n")
		git(t, root, "add", ".")
		git(t, root, "commit", "-q", "-m", "notes")
		git(t, root, "config", "core.ignorecase", "true")
		write(t, root, "pkg/zz_sub.go", "package pkg\n\nfunc Sub(a, b int) int { return a - b }\n")
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain status shows it: %s", s)
		}
		if st := tree(t, root); !st.Dirty || strings.Join(st.Shown, ",") != "pkg/zz_sub.go" {
			t.Errorf("%+v", st)
		}
		if got := strings.Join(FileSystemSettings(root), " "); !strings.Contains(got, "core.ignorecase=false") {
			t.Errorf("FileSystemSettings: %q", got)
		}
	})
	t.Run("core.fileMode=false hides an executable bit", func(t *testing.T) {
		if !strings.Contains(pinned, " core.fileMode=true ") {
			t.Skip("needs a file system that keeps executable bits (not Windows)")
		}
		root, _ := repo(t)
		write(t, root, "check.sh", "#!/bin/sh\nexit 0\n")
		git(t, root, "add", ".")
		git(t, root, "commit", "-q", "-m", "a script HEAD cannot run by name")
		git(t, root, "config", "core.fileMode", "false")
		must(t, os.Chmod(filepath.Join(root, "check.sh"), 0o755))
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain status shows it: %s", s)
		}
		if st := tree(t, root); !st.Dirty || strings.Join(st.Shown, ",") != "check.sh" {
			t.Errorf("%+v", st)
		}
	})
	t.Run("core.symlinks=false hides a link replaced by a file of its text", func(t *testing.T) {
		if !strings.Contains(pinned, " core.symlinks=true ") {
			t.Skip("needs a file system with symbolic links (not Windows)")
		}
		root, _ := repo(t)
		must(t, os.Symlink("a.txt", filepath.Join(root, "link.txt")))
		git(t, root, "add", ".")
		git(t, root, "commit", "-q", "-m", "a link")
		git(t, root, "config", "core.symlinks", "false")
		must(t, os.Remove(filepath.Join(root, "link.txt")))
		write(t, root, "link.txt", "a.txt")
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain status shows it: %s", s)
		}
		if st := tree(t, root); !st.Dirty || strings.Join(st.Shown, ",") != "link.txt" {
			t.Errorf("%+v", st)
		}
	})
}

// Where the file system cannot hold a case twin or an executable bit, git's own values stand: a repository with
// core.ignorecase=true and core.fileMode=false (what git init writes on Windows) and a checkout git made with CRLF
// line ends (core.autocrlf=true, Git for Windows' default) read clean, and a file whose name differs from the index's
// only in case reads as the tracked file there. Run on every system; the twin's part only where the folder is
// case-blind.
func TestGitsOwnDetectionStands(t *testing.T) {
	isolate(t)
	root, _ := repo(t)
	git(t, root, "config", "core.autocrlf", "true")
	for _, p := range []string{"a.txt", "pkg/a.go", "pkg/a_test.go", ".gitignore"} {
		must(t, os.Remove(filepath.Join(root, filepath.FromSlash(p))))
	}
	cmd := []string{"-c", "core.autocrlf=true", "checkout", "--", "."}
	git(t, root, cmd...)
	b, err := os.ReadFile(filepath.Join(root, "a.txt"))
	must(t, err)
	if string(b) != "a\r\n" {
		t.Fatalf("git did not check out CRLF: %q", b)
	}
	if st := tree(t, root); st.Dirty {
		t.Errorf("a CRLF checkout: %+v", st)
	}
	if strings.Contains(" "+strings.Join(fsPins(filepath.Join(root, ".git")), " ")+" ", " core.ignorecase=false ") {
		return // a case-sensitive folder: git's core.ignorecase is pinned there, and a twin can exist
	}
	git(t, root, "config", "core.ignorecase", "true")
	git(t, root, "config", "core.fileMode", "false")
	must(t, os.Rename(filepath.Join(root, "a.txt"), filepath.Join(root, "A.TXT")))
	if st := tree(t, root); st.Dirty {
		t.Errorf("a case-blind folder, the file renamed in case only: %+v", st)
	}
}

// A post-index-change hook (one the repository names, in .git/hooks or by core.hooksPath) runs whenever git writes an
// index: the contents' read writes its fresh one, so such a hook could mark the edited file there as assume-unchanged
// and the read would trust it. The read names no hooks folder of the repository's.
func TestNoHookDuringTheContentsRead(t *testing.T) {
	isolate(t)
	root, _ := repo(t)
	hook := "#!/bin/sh\n[ -n \"$BONSAI_TEST_HOOKED\" ] && exit 0\nexport BONSAI_TEST_HOOKED=1\ngit update-index --assume-unchanged a.txt\n"
	for _, dir := range []string{".git/hooks", ".git/more-hooks"} {
		write(t, root, dir+"/post-index-change", hook)
		must(t, os.Chmod(filepath.Join(root, filepath.FromSlash(dir), "post-index-change"), 0o755))
	}
	write(t, root, "a.txt", "changed\n")
	git(t, root, "update-index", "--assume-unchanged", "a.txt")
	for _, set := range []string{"", ".git/more-hooks"} {
		if set != "" {
			git(t, root, "config", "core.hooksPath", set)
		}
		if s := plainStatus(t, root); s != "" {
			t.Fatalf("plain status shows the change: %s", s)
		}
		st := tree(t, root)
		if !st.Dirty || hiddenPaths(st) != "a.txt" {
			t.Errorf("hooks %q: %+v", set, st)
		}
	}
}
