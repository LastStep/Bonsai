package engine

// The path walk's failing closed (V1's B1 on 5.6, M-2026-10-18-8ba15229): a path deeper than the old 64-step limit
// behind a link out of the folder, a link loop, and a relative link target not made yet whose `..` must be read from
// the link's real folder (V1's N1). The links are symbolic links, which need a privilege on Windows: each case skips
// where none can be made (realpath_windows_test.go holds Windows' junctions, which need none). Each spells its temporary
// folder by its short (8.3) name where Windows gives it one (shortTemp), as CI's runner does.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// symlinkOr skips the test where no symbolic link can be made (on Windows it needs a privilege).
func symlinkOr(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("no link can be made here (on Windows it needs a privilege): %v", err)
	}
}

// deep is base with n folders named d0, d1, ... below it.
func deep(base string, n int) string {
	parts := []string{base}
	for i := 0; i < n; i++ {
		parts = append(parts, "d"+strings.Repeat("x", i%3))
	}
	return filepath.Join(parts...)
}

func TestFolderHoldsPastTheOldLimit(t *testing.T) {
	tmp := shortTemp(t)
	root, outside := filepath.Join(tmp, "root"), filepath.Join(tmp, "outside")
	for _, d := range []string{root, outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	link := filepath.Join(root, "link")
	symlinkOr(t, outside, link)
	for _, n := range []int{1, 60, 63, 64, 65, 70, 200} {
		if FolderHolds(root, deep(link, n)) {
			t.Errorf("a link out of the root, then %d folders not made yet: held inside the root", n)
		}
		if !FolderHolds(root, deep(root, n)) {
			t.Errorf("%d folders not made yet inside the root: not held", n)
		}
		got, err := RealPath(deep(link, n))
		if err != nil || got != deep(outside, n) {
			t.Errorf("RealPath of a link out, then %d folders: %q, %v; want %q", n, got, err, deep(outside, n))
		}
	}
}

func TestRealPathFailsClosed(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "root")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	// A loop of two links: the walk never ends, so it fails, and the folder holds nothing reached through it.
	symlinkOr(t, filepath.Join(root, "b"), filepath.Join(root, "a"))
	symlinkOr(t, filepath.Join(root, "a"), filepath.Join(root, "b"))
	for _, p := range []string{filepath.Join(root, "a"), filepath.Join(root, "a", "new"), deep(filepath.Join(root, "a"), 70)} {
		if FolderHolds(root, p) {
			t.Errorf("%s, through a link loop: held inside the root", p)
		}
		if got, err := RealPath(p); err == nil {
			t.Errorf("RealPath(%s) through a link loop: %q, no error", p, got)
		}
	}
	// A chain of 65 links, each to the next, ending in a folder inside the root: past the limit, refused; a chain of
	// 64 is followed.
	end := filepath.Join(root, "end")
	if err := os.Mkdir(end, 0o755); err != nil {
		t.Fatal(err)
	}
	chain := func(name string, n int) string {
		next := end
		for i := n; i > 0; i-- {
			l := filepath.Join(root, name+strings.Repeat("c", i))
			symlinkOr(t, next, l)
			next = l
		}
		return next
	}
	if p := chain("ok", 64); !FolderHolds(root, filepath.Join(p, "new")) {
		t.Errorf("a chain of 64 links inside the root: not held")
	}
	if p := chain("long", 65); FolderHolds(root, filepath.Join(p, "new")) {
		t.Errorf("a chain of 65 links: held, want refused past the limit")
	}
	// The folder's own side: a root that cannot be walked holds nothing.
	if FolderHolds(filepath.Join(root, "a"), filepath.Join(root, "a", "x")) {
		t.Error("a root through a link loop held a path")
	}
}

// V1's N1: root/L -> outside/deep, and outside/deep/dang -> ../x, x not made. The write reaches outside/x; read by
// spelling from the link's folder as written, it was root/x.
func TestARelativeLinkTargetIsReadFromTheLinksRealFolder(t *testing.T) {
	tmp := shortTemp(t)
	root, outside := filepath.Join(tmp, "root"), filepath.Join(tmp, "outside")
	if err := os.MkdirAll(filepath.Join(outside, "deep"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	symlinkOr(t, filepath.Join(outside, "deep"), filepath.Join(root, "L"))
	symlinkOr(t, filepath.Join("..", "x"), filepath.Join(outside, "deep", "dang"))
	for _, p := range []string{filepath.Join(root, "L", "dang"), filepath.Join(root, "L", "dang", "new")} {
		if FolderHolds(root, p) {
			t.Errorf("%s: held inside the root, but the write reaches %s", p, filepath.Join(outside, "x"))
		}
	}
	if got, err := RealPath(filepath.Join(root, "L", "dang", "new")); err != nil || got != filepath.Join(outside, "x", "new") {
		t.Errorf("RealPath: %q, %v; want %q", got, err, filepath.Join(outside, "x", "new"))
	}
	// A target's own `..` after a link in it: outside/deep/up -> sub/../.. where sub is a link to outside/deep/a/b:
	// the kernel reads sub/.. as outside/deep/a, so the target is outside/deep; by spelling it would be outside.
	if err := os.MkdirAll(filepath.Join(outside, "deep", "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	symlinkOr(t, filepath.Join(outside, "deep", "a", "b"), filepath.Join(outside, "deep", "sub"))
	symlinkOr(t, strings.Join([]string{"sub", "..", "..", "gone"}, string(filepath.Separator)), filepath.Join(outside, "deep", "up"))
	if got, err := RealPath(filepath.Join(outside, "deep", "up")); err != nil || got != filepath.Join(outside, "deep", "gone") {
		t.Errorf("RealPath through sub/..: %q, %v; want %q", got, err, filepath.Join(outside, "deep", "gone"))
	}
}
