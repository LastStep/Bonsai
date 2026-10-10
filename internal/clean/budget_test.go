package clean

// The budget: a run on a folder of many files stops when its budget runs out, cleans what it judged before that, and
// the next runs go on from there until the folder is clean.

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBudgetStopsTheRun(t *testing.T) {
	const files = 600
	const budget = 100 * time.Millisecond
	l := project(t, "  log:\n    keep_days: 30\n")
	dir := filepath.Join(l.Main, ".bonsai", "local", "log")
	for i := 0; i < files; i++ {
		id := fmt.Sprintf("f%07d-0000-4000-8000-000000000000", i)
		at := daysAgo(40).Add(time.Duration(i) * time.Second)
		p := filepath.Join(dir, "s-"+id+".ndjson")
		if err := os.WriteFile(p, logLine(t, rec{event: "guard", at: at, session: id}), 0o644); err != nil {
			t.Fatal(err)
		}
		stamp(t, p, at)
	}
	o := opts(l)
	o.Budget = budget
	start := time.Now()
	d := SessionEnd(o)
	took := time.Since(start)
	if !d.Stopped || len(d.Cleaned) == 0 || len(d.Cleaned) >= files || len(d.Errs) > 0 {
		t.Fatalf("stopped %v, cleaned %d of %d, errors %v", d.Stopped, len(d.Cleaned), files, d.Errs)
	}
	// It stops within its budget and the one file it was judging; the slack is for a loaded machine.
	if took > budget+400*time.Millisecond {
		t.Errorf("a run with a budget of %v took %v", budget, took)
	}
	t.Logf("%d files: the first run cleaned %d in %v (budget %v)", files, len(d.Cleaned), took, budget)
	// The oldest go first, so what is left is the newest.
	if d.Cleaned[0].Target != ".bonsai/local/log/s-f0000000-0000-4000-8000-000000000000.ndjson" {
		t.Errorf("the first cleaned is %s, not the oldest", d.Cleaned[0].Target)
	}
	// The next runs go on until the folder is clean, each cleaning some.
	total, runs := len(d.Cleaned), 1
	for d.Stopped {
		d = SessionEnd(o)
		total += len(d.Cleaned)
		runs++
		if d.Stopped && len(d.Cleaned) == 0 {
			t.Fatalf("run %d stopped having cleaned nothing (%d of %d cleaned before it)", runs, total, files)
		}
	}
	if total != files || len(d.Errs) > 0 {
		t.Errorf("after %d runs, %d of %d cleaned, errors %v", runs, total, files, d.Errs)
	}
	t.Logf("%d runs cleaned all %d", runs, files)
}
