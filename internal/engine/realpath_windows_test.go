package engine

// RealPath on Windows follows junctions (FinalPath), which Go's own link calls no longer do (since Go 1.23), and which
// any user may make without a privilege: a test build's scratch root holds nothing reached through a junction out of
// it, at any depth; a junction that leads nowhere fails closed.

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

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
