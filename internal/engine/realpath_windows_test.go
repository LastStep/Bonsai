package engine

// RealPath on Windows follows junctions (FinalPath), which Go's own link calls no longer do (since Go 1.23), and which
// any user may make without a privilege: a test build's scratch root holds nothing reached through a junction out of
// it, at any depth; a junction that leads nowhere fails closed.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// shortTemp is a new temporary folder spelled by its short (8.3) name where the volume gives it one, as CI's Windows
// runner spells its own (C:\Users\RUNNER~1\...): GetShortPathName shortens every name of the folder that has a short
// name. Where 8.3 names are off, it is the folder as made, and the tests read it as on any other system.
func shortTemp(t *testing.T) string {
	t.Helper()
	long := t.TempDir()
	in, err := syscall.UTF16PtrFromString(long)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]uint16, 1024)
	n, err := syscall.GetShortPathName(in, &buf[0], uint32(len(buf)))
	if err != nil || n == 0 || int(n) >= len(buf) {
		return long
	}
	return syscall.UTF16ToString(buf[:n])
}

// hasShort reports whether p is spelled with a short name: shortTemp found one for it.
func hasShort(p string) bool { return strings.Contains(p, "~") }

// junction makes a directory junction with cmd's mklink /J, which needs no privilege (unlike a symbolic link).
func junction(t *testing.T, link, target string) {
	t.Helper()
	if out, err := exec.Command("cmd.exe", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
		t.Fatalf("mklink /J %s %s: %v %s", link, target, err, out)
	}
}

func TestFolderHoldsFollowsAJunction(t *testing.T) {
	tmp := t.TempDir()
	root, outside, inner := filepath.Join(tmp, "root"), filepath.Join(tmp, "outside"), filepath.Join(tmp, "root", "inner")
	for _, d := range []string{root, outside, inner} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	junction(t, filepath.Join(root, "out"), outside)
	junction(t, filepath.Join(root, "in"), inner)
	for _, n := range []int{0, 1, 70} {
		if FolderHolds(root, deep(filepath.Join(root, "out"), n)) {
			t.Errorf("a junction out of the root, then %d folders not made yet: held inside the root", n)
		}
		if !FolderHolds(root, deep(filepath.Join(root, "in"), n)) {
			t.Errorf("a junction inside the root, then %d folders: not held", n)
		}
	}
	// A junction whose target is gone: the walk fails, and nothing holds it.
	gone := filepath.Join(tmp, "gone")
	if err := os.Mkdir(gone, 0o755); err != nil {
		t.Fatal(err)
	}
	junction(t, filepath.Join(root, "dang"), gone)
	if err := os.Remove(gone); err != nil {
		t.Fatal(err)
	}
	if FolderHolds(root, filepath.Join(root, "dang", "new")) {
		t.Error("a junction that leads nowhere: held inside the root")
	}
	if p, err := RealPath(filepath.Join(root, "dang", "new")); err == nil {
		t.Errorf("RealPath through a junction that leads nowhere: %q, no error", p)
	}
}

// A folder spelled by its short (8.3) name, and a path not made yet inside it: RealPath names both by their long
// names, so a folder and a path in it compare in one form however either is spelled (CI's runner spells its temporary
// folder C:\Users\RUNNER~1\...). It needs a volume with 8.3 names on, and skips where they are off.
func TestRealPathReadsAShortName(t *testing.T) {
	short := shortTemp(t)
	if !hasShort(short) {
		t.Skipf("short (8.3) names are off on this volume: %s has none", short)
	}
	long, err := filepath.EvalSymlinks(short)
	if err != nil {
		t.Fatal(err)
	}
	if hasShort(long) {
		t.Fatalf("EvalSymlinks(%s) = %s, still short", short, long)
	}
	if err := os.Mkdir(filepath.Join(long, "made"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"", "made", filepath.Join("made", "new", "x.txt")} {
		got, err := RealPath(filepath.Join(short, rel))
		if want := filepath.Join(long, rel); err != nil || got != want {
			t.Errorf("RealPath(%s): %q, %v; want %q", filepath.Join(short, rel), got, err, want)
		}
	}
	if !FolderHolds(short, filepath.Join(long, "made", "new")) || !FolderHolds(long, filepath.Join(short, "made", "new")) {
		t.Error("a folder and a path not made yet in it, one spelled short and one long: not held")
	}
	if FolderHolds(filepath.Join(short, "made"), filepath.Join(long, "other")) {
		t.Error("a path beside the folder, spelled long: held inside it, spelled short")
	}
}
