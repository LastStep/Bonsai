//go:build !windows

package message

// A pointer to a FIFO or to a device (VM's D1), on the systems that have them at a path (Windows has neither: its
// device names are pointer_test.go's): refused at writing as pointer-missing, saying what it is, and never opened to
// wait or read without end; a pointer whose file is replaced afterwards by either is missing at --show. The old read
// hung on the FIFO and read /dev/zero until it ran out of memory.

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// fifo makes a FIFO at rel in the fixture's checkout.
func (f *fixture) fifo(rel string) string {
	f.t.Helper()
	path := filepath.Join(f.main, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		f.t.Skipf("no FIFO on this system: %v", err)
	}
	return path
}

// link makes a symbolic link at rel in the fixture's checkout to target.
func (f *fixture) link(rel, target string) string {
	f.t.Helper()
	path := filepath.Join(f.main, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		f.t.Skipf("no symbolic link on this system: %v", err)
	}
	return path
}

func devZero(t *testing.T) {
	t.Helper()
	if fi, err := os.Stat("/dev/zero"); err != nil || fi.Mode()&os.ModeDevice == 0 {
		t.Skipf("no /dev/zero device on this system: %v", err)
	}
}

// refusedSettles writes a brief pointing to rel, which must be refused pointer-missing with why in its words, and
// fails the test if the write has not returned within a minute (settles), never waiting on it in the test's own
// goroutine.
func (f *fixture) refusedSettles(rel, why string) {
	f.t.Helper()
	var e *Error
	settles(f.t, "--pointer "+rel, func() {
		_, e = Write(f.p, Writing{From: "orchestrator", To: "builder@T-0901", Kind: "brief", Body: "Objective: one thing.",
			Pointers: []string{rel}, WorkDir: f.main}, DefaultRules())
	})
	if e == nil || e.Code != "pointer-missing" || !strings.Contains(e.What, why) || e.Next == "" {
		f.t.Errorf("--pointer %s: %v, want pointer-missing saying %q", rel, e, why)
	}
	if files := f.files(); len(files) != 0 {
		f.t.Errorf("--pointer %s wrote %v", rel, files)
	}
}

func TestPointerFIFORefused(t *testing.T) {
	f := newFixture(t)
	f.fifo("work/pipe.md")
	f.refusedSettles("work/pipe.md", "is a pipe (a FIFO), or a link to one, not a regular file")
	f.link("work/to-pipe.md", "pipe.md")
	f.refusedSettles("work/to-pipe.md", "is a pipe (a FIFO), or a link to one, not a regular file")
}

func TestPointerLinkToDeviceRefused(t *testing.T) {
	devZero(t)
	f := newFixture(t)
	f.link("work/zero.md", "/dev/zero")
	f.refusedSettles("work/zero.md", "is a device, or a link to one, not a regular file")
}

// A pointer written to a regular file, which is then replaced by a FIFO, then by a link to /dev/zero: --show reads it
// as missing each time, at once.
func TestPointerReplacedAfterwards(t *testing.T) {
	devZero(t)
	f := newFixture(t)
	f.put("work/a.md", "words\n")
	r := DefaultRules()
	w := f.write(Writing{From: "orchestrator", To: "builder@T-0901", Kind: "brief", Pointers: []string{"work/a.md"}}, r)
	path := filepath.Join(f.main, "work", "a.md")
	show := func(what string) {
		t.Helper()
		st := f.read()
		var shown *Shown
		var e *Error
		settles(t, what, func() { shown, e = Show(f.p, st, w.Message.ID, r, f.clock) })
		if e != nil || strings.Join(shown.Pointers, ",") != PointerMissing {
			t.Errorf("%s: %v %v", what, shown, e)
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	f.fifo("work/a.md")
	show("--show with the pointer's file a FIFO now")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	f.link("work/a.md", "/dev/zero")
	show("--show with the pointer's file a link to /dev/zero now")
}

// A FIFO put in a regular file's place between the look and the open is opened without waiting and refused by its
// open handle, never read.
func TestPointerFIFOSwappedIn(t *testing.T) {
	f := newFixture(t)
	a := f.put("work/a.md", "words\n")
	fi, err := os.Stat(a)
	if err != nil {
		t.Fatal(err)
	}
	p := f.fifo("work/p.md")
	var sum, why string
	settles(t, "a FIFO swapped in", func() { sum, why, err = fingerprint(p, fi) })
	if err != nil || sum != "" || !strings.Contains(why, "replaced while it was read: it is a pipe") {
		t.Errorf("%q %q %v", sum, why, err)
	}
}
