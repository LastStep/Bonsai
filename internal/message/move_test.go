package message

// The move that files a message: retried while Windows reports the file busy (an antivirus scan, an indexer, an
// editor), to moveWait's schedule against a clock of the test's own; any other error, and a taken name, at once; and
// on each system a file held open with no sharing (Windows refuses the move until it is let go; Linux and macOS have no
// share modes, so the move is made at once there).

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fakeMove stands in a move that is busy for its first busyTries tries, with a clock that sleeps no real time.
type fakeMove struct {
	tries, busyTries int
	clock            time.Time
	slept            []time.Duration
	last             error
}

var errBusy = errors.New("busy: the process cannot access the file")

func (f *fakeMove) install(t *testing.T) {
	t.Helper()
	m, b, n, s := moveFile, busy, now, sleep
	t.Cleanup(func() { moveFile, busy, now, sleep = m, b, n, s })
	f.clock = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	moveFile = func(string, string) error {
		f.tries++
		if f.busyTries < 0 || f.tries <= f.busyTries {
			return errBusy
		}
		return f.last
	}
	busy = func(err error) bool { return errors.Is(err, errBusy) }
	now = func() time.Time { return f.clock }
	sleep = func(d time.Duration) { f.slept = append(f.slept, d); f.clock = f.clock.Add(d) }
}

func TestMoveRetriesWhileBusy(t *testing.T) {
	// Busy three times, then moved: four tries, waiting 5, 10 and 20 ms.
	f := &fakeMove{busyTries: 3}
	f.install(t)
	if err := moveRetry("a", "b"); err != nil || f.tries != 4 || len(f.slept) != 3 || f.slept[0] != 5*time.Millisecond || f.slept[2] != 20*time.Millisecond {
		t.Errorf("busy three times: %v, %d tries, slept %v", err, f.tries, f.slept)
	}
	// Busy forever: it gives up once moveWait has passed, the waits doubling to 160 ms and staying there.
	f = &fakeMove{busyTries: -1}
	f.install(t)
	err := moveRetry("a", "b")
	var total time.Duration
	for _, d := range f.slept {
		total += d
		if d > 160*time.Millisecond {
			t.Errorf("a wait of %v", d)
		}
	}
	if !errors.Is(err, errBusy) || total <= moveWait || total-f.slept[len(f.slept)-1] > moveWait {
		t.Errorf("busy forever: %v after %v of waits (%d tries)", err, total, f.tries)
	}
	// Any other error, and a taken name, at once.
	for _, want := range []error{errors.New("no such folder"), errTaken} {
		f = &fakeMove{last: want}
		f.install(t)
		if err := moveRetry("a", "b"); err != want || f.tries != 1 || len(f.slept) != 0 {
			t.Errorf("%v: %v after %d tries", want, err, f.tries)
		}
	}
}

// The real move: never onto a file that is there; and a staged file held open with no sharing is moved once it is let
// go (on Windows only after the hold ends, through the busy retry; elsewhere at once).
func TestMoveNeverReplacesAndWaitsOutAHold(t *testing.T) {
	dir := t.TempDir()
	from, to := filepath.Join(dir, ".staged"), filepath.Join(dir, "M.md")
	for path, text := range map[string]string{from: "new\n", to: "old\n"} {
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := moveRetry(from, to); err != errTaken {
		t.Fatalf("a move onto a file that is there: %v", err)
	}
	if raw, _ := os.ReadFile(to); string(raw) != "old\n" {
		t.Errorf("the file there was replaced: %q", raw)
	}
	if err := os.Remove(to); err != nil {
		t.Fatal(err)
	}
	release := holdFile(t, from)
	done := make(chan struct{})
	go func() {
		time.Sleep(150 * time.Millisecond)
		release()
		close(done)
	}()
	err := moveRetry(from, to)
	<-done
	if err != nil {
		t.Fatalf("the held file was not moved once let go: %v", err)
	}
	if raw, _ := os.ReadFile(to); string(raw) != "new\n" {
		t.Errorf("moved: %q", raw)
	}
	if _, err := os.Stat(from); !os.IsNotExist(err) {
		t.Errorf("the staged name is still there: %v", err)
	}
}
