package clean

// The budget: a run on a folder of many files stops when its budget runs out, cleans what it judged before that, and
// the next runs go on from there until the folder is clean. The budget's clock is the test's (heldClock): each delete
// moves it on by a fixed pace and nothing else does, so each session end's run cleans exactly its budget's worth, the
// oldest first, on any machine. Timed by the wall clock, this test failed on a loaded Windows runner (listing 600 files
// once took longer than a 100 ms budget; later, under load, a 1 s budget ran out before a run reached its first file),
// on the code before the budget was checked during the listing as on the code after. The wall clock's own limit, a run
// ending within its budget and one step, is TestBudgetHugeFolder's (stop_test.go).

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBudgetStopsTheRun(t *testing.T) {
	const files = 600
	const budget = SessionEndBudget
	const pace = 5 * time.Millisecond // each delete moves the budget's clock on this much: 200 files a run
	const perRun = int(budget / pace)
	l := project(t, "  log:\n    keep_days: 30\n")
	dir := filepath.Join(l.Main, ".bonsai", "local", "log")
	var oldest []string // the targets, oldest first
	for i := 0; i < files; i++ {
		id := fmt.Sprintf("f%07d-0000-4000-8000-000000000000", i)
		at := daysAgo(40).Add(time.Duration(i) * time.Second)
		p := filepath.Join(dir, "s-"+id+".ndjson")
		if err := os.WriteFile(p, logLine(t, rec{event: "guard", at: at, session: id}), 0o644); err != nil {
			t.Fatal(err)
		}
		stamp(t, p, at)
		oldest = append(oldest, target(Log, "s-"+id+".ndjson"))
	}
	advance := heldClock(t)
	del := removeFile
	t.Cleanup(func() { removeFile = del })
	removeFile = func(name string) error {
		advance(pace)
		return del(name)
	}
	o := opts(l)
	o.Budget = budget
	// Each run stops when its budget runs out, having cleaned its budget's worth; the next goes on from there, the
	// oldest first, until the folder is clean.
	var cleaned []Item
	runs := 0
	for {
		runs++
		d := SessionEnd(o)
		want := min(perRun, files-len(cleaned))
		if len(d.Errs) > 0 || len(d.Cleaned) != want {
			t.Fatalf("run %d: cleaned %d, want %d (%d of %d cleaned before it); errors %v", runs, len(d.Cleaned), want,
				len(cleaned), files, d.Errs)
		}
		cleaned = append(cleaned, d.Cleaned...)
		if !d.Stopped {
			break
		}
		if runs > files {
			t.Fatalf("%d runs and the folder is not clean", runs)
		}
	}
	if len(cleaned) != files {
		t.Errorf("%d runs cleaned %d of %d", runs, len(cleaned), files)
	}
	for i, it := range cleaned {
		if it.Target != oldest[i] {
			t.Fatalf("cleaned %d is %s, want %s: the oldest first", i, it.Target, oldest[i])
		}
	}
}
