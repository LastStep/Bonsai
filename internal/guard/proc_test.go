package guard

// Paths into /proc (V1's finding on 5.3.1): /proc/self/... and /proc/<pid>/... name one process's files, and the
// guard would resolve them in its own, so a file tool's write there, or through a link that leads there, is refused
// as bad-path. Linux only: no other system has /proc, and on Windows a link needs a privilege (the string half runs
// everywhere).

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

func TestIntoProc(t *testing.T) {
	for _, c := range []struct {
		a, goos string
		want    bool
	}{{"/proc/self/cwd/x", "linux", true}, {"/proc", "linux", true}, {"/procfs/x", "linux", false},
		{"/home/x/proc/y", "linux", false}, {"/proc/self/cwd/x", "windows", false}, {"/proc/1/root/etc", "darwin", false}} {
		if got := intoProc(c.a, c.a, c.goos); got != c.want {
			t.Errorf("intoProc(%s, %s) = %v", c.a, c.goos, got)
		}
	}
}

func TestAPathIntoProcIsRefused(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("/proc is Linux's; TestIntoProc reads the paths on every system")
	}
	dir := withTask(t, "running", "**")
	for _, p := range []string{"/proc/self/cwd/free.txt", "/proc/1/root/tmp/x", "/proc"} {
		in := toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: p}, {Key: "content", Value: "x"}})
		if s := wantRule(t, dir, in, RuleBadPath); !strings.Contains(s, "it leads into /proc") || strings.Contains(s, "on Windows") {
			t.Errorf("%s: %s", p, s)
		}
	}
	// A link in the project that leads into /proc.
	if err := os.Symlink("/proc/self/cwd", filepath.Join(dir, "here")); err != nil {
		t.Fatal(err)
	}
	wantRule(t, dir, writeOf(dir, "here/free.txt", "x\n"), RuleBadPath)
	// A link that leads elsewhere in the project is followed as before.
	if err := os.Symlink(filepath.Join(dir, "assets"), filepath.Join(dir, "pics")); err != nil {
		t.Fatal(err)
	}
	wantRule(t, dir, writeOf(dir, "pics/b.png", "x\n"), RuleGranted)
	// A delete through /proc too.
	if d := judgeIn(dir).mayDelete(dir, "/proc/self/cwd/.git"); d.Rule != RuleBadPath {
		t.Errorf("a delete through /proc: %+v", d)
	}
}
