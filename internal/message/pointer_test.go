package message

// A pointer is read only from a regular file, streamed (VM's D1): its fingerprint streamed with the LF rule as
// workspace.HashLF reads it whole; a large file hashed with flat memory; a folder, and a file put in another's place
// while it was read, never fingerprinted; on Windows, nul and the other names it keeps for devices refused before
// anything is opened. A FIFO and a link to /dev/zero are pointer_unix_test.go's (Windows has neither at a path).

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// settles runs f and fails the test if it has not returned within a minute: a pointer read that waits on a FIFO's
// writer or reads a device without end never returns (the old read did both), and the test says so rather than hang.
func settles(t *testing.T, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(time.Minute):
		t.Fatalf("%s: still reading after a minute (a read that waits on a FIFO, or reads a device without end)", what)
	}
}

// The streamed fingerprint is workspace.HashLF's, whatever the reads' sizes: a CRLF split across two reads, a CR at
// the very end, a CR before a CRLF, a lone CR, an LF before a CR.
func TestHashLFStreamed(t *testing.T) {
	big := bytes.Repeat([]byte("line\r\n"), hashBuffer/6+7) // past one buffer, a CRLF across its edge somewhere
	edge := append(bytes.Repeat([]byte("x"), hashBuffer-1), "\r\nnext\r"...)
	for _, in := range [][]byte{nil, []byte("\r"), []byte("\r\n"), []byte("a\r\nb\r\n"), []byte("a\rb"), []byte("\r\r\n"),
		[]byte("\n\r"), []byte("end\r"), []byte("\r\r\r"), big, edge} {
		want := workspace.HashLF(in)
		for name, r := range map[string]io.Reader{"whole": bytes.NewReader(in), "one byte": iotest.OneByteReader(bytes.NewReader(in)),
			"halves": iotest.HalfReader(bytes.NewReader(in)), "data with EOF": iotest.DataErrReader(bytes.NewReader(in))} {
			got, err := hashLF(r)
			if err != nil || got != want {
				t.Errorf("%q (%d bytes) read %s: %s, %v; want %s", trim(in), len(in), name, got, err, want)
			}
		}
	}
	if _, err := hashLF(iotest.ErrReader(io.ErrUnexpectedEOF)); err == nil {
		t.Error("a read that fails gives a fingerprint")
	}
}

func trim(b []byte) string {
	if len(b) > 20 {
		return string(b[:20]) + "..."
	}
	return string(b)
}

// A large regular file is fingerprinted with flat memory: what the read allocates stays a small fixed amount, far
// below the file's size (the old read held the whole file); its fingerprint is HashLF's, and --show reads it so too.
func TestPointerLargeFileFlatMemory(t *testing.T) {
	f := newFixture(t)
	const size = 48 << 20
	block := bytes.Repeat([]byte("forty-eight megabytes of words\r\n"), 1<<15) // 1 MiB, CRLF line ends
	path := filepath.Join(f.main, "work", "big.log")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	for written := 0; written < size; written += len(block) {
		if _, err := out.Write(block); err != nil {
			t.Fatal(err)
		}
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	ptr, e := Pointer(f.p, "work/big.log", f.main)
	runtime.ReadMemStats(&after)
	if e != nil {
		t.Fatal(e)
	}
	if grew := after.TotalAlloc - before.TotalAlloc; grew > 4<<20 {
		t.Errorf("fingerprinting a %d MiB file allocated %d KiB: the file was held, not streamed", size>>20, grew>>10)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if ptr.Path != "work/big.log" || ptr.SHA256 != workspace.HashLF(raw) {
		t.Errorf("the pointer: %+v, want the sha256 %s", ptr, workspace.HashLF(raw))
	}
	if got := PointerState(f.p, ptr); got != PointerUnchanged {
		t.Errorf("--show reads it as %s", got)
	}
}

// A folder is no file: refused at writing as pointer-missing, saying so, and missing at --show.
func TestPointerFolder(t *testing.T) {
	f := newFixture(t)
	f.put("work/a.md", "words\n")
	e := f.refused(Writing{From: "orchestrator", To: "builder@T-0901", Kind: "brief", Pointers: []string{"work"}}, DefaultRules(), "pointer-missing")
	if !strings.Contains(e.What, "is a folder") {
		t.Errorf("the refusal: %s", e.What)
	}
	if got := PointerState(f.p, format.MessagePointer{Path: "work", SHA256: strings.Repeat("0", 64)}); got != PointerMissing {
		t.Errorf("a folder at --show: %s", got)
	}
}

// A file put in the place of the one the path named, between the look and the open, is never read: the open file is
// not the file looked at.
func TestPointerReplacedWhileRead(t *testing.T) {
	f := newFixture(t)
	a, b := f.put("work/a.md", "one\n"), f.put("work/b.md", "two\n")
	fi, err := os.Stat(a)
	if err != nil {
		t.Fatal(err)
	}
	sum, why, err := fingerprint(b, fi)
	if err != nil || sum != "" || !strings.Contains(why, "replaced by another file") {
		t.Errorf("another file in its place: %q %q %v", sum, why, err)
	}
	if sum, why, err := fingerprint(a, fi); err != nil || why != "" || sum != workspace.HashLF([]byte("one\n")) {
		t.Errorf("the file itself: %q %q %v", sum, why, err)
	}
}

// On Windows nul, and every name it keeps for a device, in any case and with any extension, is refused at writing as
// pointer-missing before anything is opened, and is missing at --show; the old read took --pointer nul as an empty
// file. Elsewhere such a name is a file's like any other.
func TestPointerWindowsDevices(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only Windows keeps names for devices: elsewhere nul is a file's name like any other, read if it is a regular file")
	}
	f := newFixture(t)
	for _, name := range []string{"nul", "NUL", "nul.txt", "con", "aux.md", "prn", "com1", "LPT9.log", "conin$", "CONOUT$", "work/nul"} {
		e := f.refused(Writing{From: "orchestrator", To: "builder@T-0901", Kind: "brief", Pointers: []string{name}}, DefaultRules(), "pointer-missing")
		if !strings.Contains(e.What, "a name Windows keeps for a device") {
			t.Errorf("%s: %s", name, e.What)
		}
		if got := PointerState(f.p, format.MessagePointer{Path: name, SHA256: workspace.HashLF(nil)}); got != PointerMissing {
			t.Errorf("%s at --show: %s", name, got)
		}
	}
	f.put("work/nullish.md", "a file\n") // a name that only starts like one is a file
	if _, e := Pointer(f.p, "work/nullish.md", f.main); e != nil {
		t.Errorf("nullish.md: %v", e)
	}
}

// A relative pointer into .bonsai/local/ sent from a linked worktree is read in the main checkout, where Bonsai keeps
// those files (the ladder's result, the messages); any other relative pointer keeps reading against the worktree
// (BUG-125). A worktree's own .bonsai/local is never looked at for such a path.
func TestPointerIntoLocalFromWorktree(t *testing.T) {
	f := newFixture(t)
	wt := t.TempDir()
	f.p.Local = workspace.Local{Root: wt, Main: f.main, Branch: "t-0901"}
	write := func(root, rel, body string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(f.main, ".bonsai/local/ladder/T-0901.json", "{\"green\":true}\n")
	write(wt, "work/a.md", "words\n")
	ptr, e := Pointer(f.p, ".bonsai/local/ladder/T-0901.json", wt)
	if e != nil {
		t.Fatalf("a pointer to the ladder result from a worktree: %v", e)
	}
	if ptr.Path != ".bonsai/local/ladder/T-0901.json" {
		t.Errorf("the path as written: %q", ptr.Path)
	}
	if got := PointerState(f.p, ptr); got != PointerUnchanged {
		t.Errorf("read at --show from the worktree: %s", got)
	}
	// Another relative pointer is read against the worktree, as before.
	if ptr, e := Pointer(f.p, "work/a.md", wt); e != nil || ptr.Path != "work/a.md" {
		t.Errorf("a worktree file: %+v %v", ptr, e)
	}
	// A file named like the main's but only in the worktree's own tree outside .bonsai/local is unchanged: not found
	// in main, still found in the worktree.
	if _, e := Pointer(f.p, ".bonsai/other/x.json", wt); e == nil {
		t.Errorf("a path outside .bonsai/local that is in neither checkout was accepted")
	}
}
