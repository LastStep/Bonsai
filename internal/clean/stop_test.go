package clean

// Where a run stops (files.go, "The budget"): the budget is checked before each step of the listing, the ranking and
// the judging, so a huge folder stops at the budget too; a run stopped anywhere leaves each file deleted with its clean
// record, or as it was, and the next run goes on. The test that times a run holds it to a limit the code must meet
// (its budget and stepCeiling), never to this machine's speed; the others move the budget's clock themselves (clock),
// so where a run stops depends on the code alone.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// stepCeiling is how far past its budget a timed run may go: one step (a batch of the listing, one file's age, one
// file judged, deleted and recorded) and the run's return. It is the 3 s the recorder's own budget of 4 s leaves past
// a session end's SessionEndBudget of 1 s: a longer step, and the recorder's limit would cut the cleaning off. A limit
// the code must meet, not a measure of this machine.
const stepCeiling = 3 * time.Second

// heldClock stands the budget's clock still; advance moves it on. Where a run stops then depends only on what the test
// moves it on at.
func heldClock(t *testing.T) (advance func(time.Duration)) {
	t.Helper()
	at := time.Now()
	old := clock
	clock = func() time.Time { return at }
	t.Cleanup(func() { clock = old })
	return func(d time.Duration) { at = at.Add(d) }
}

// tickingClock moves the budget's clock on 1 ms each time a run reads it: a budget of n ms stops a run at its n-th
// check, so a sweep of n meets every place a run can stop.
func tickingClock(t *testing.T) {
	t.Helper()
	at := time.Now()
	old := clock
	clock = func() time.Time {
		at = at.Add(time.Millisecond)
		return at
	}
	t.Cleanup(func() { clock = old })
}

// oldLogFiles writes n old log files holding no span, each going by keep_days 30, the first the oldest. It returns
// their names.
func oldLogFiles(t *testing.T, l workspace.Local, n int) []string {
	t.Helper()
	var out []string
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("d%07d-0000-4000-8000-000000000000", i)
		name := "s-" + id + ".ndjson"
		logFile(t, l, name, rec{event: "guard", at: daysAgo(40).Add(time.Duration(i) * time.Second), session: id})
		out = append(out, name)
	}
	return out
}

// targets are the cleaned items' targets, sorted.
func targets(items []Item) []string {
	var out []string
	for _, it := range items {
		out = append(out, it.Target)
	}
	return sorted(out)
}

// recordTargets are the clean records' targets, sorted.
func recordTargets(t *testing.T, l workspace.Local) []string {
	t.Helper()
	var out []string
	for _, r := range cleanRecords(t, l) {
		out = append(out, r[:strings.Index(r, "=")])
	}
	return sorted(out)
}

// A huge folder, by the wall clock: 4,000 log files, of which 300 go (old, no span), 100 are old but kept (an ended
// session with no row yet) and the rest are new. A run with a budget of 100 ms ends within the budget and
// stepCeiling, and deletes only files it judged, each with its record; a run with time then finishes the job.
func TestBudgetHugeFolder(t *testing.T) {
	const files = 4000
	const budget = 100 * time.Millisecond
	l := project(t, "  log:\n    keep_days: 30\n")
	dir := filepath.Join(l.Main, ".bonsai", "local", "log")
	var goes []string
	for i := 0; i < files; i++ {
		id := fmt.Sprintf("c%07d-0000-4000-8000-000000000000", i)
		name := "s-" + id + ".ndjson"
		switch {
		case i%40 == 0: // old, kept: an ended session with no row
			logFile(t, l, name, session(id, "none", daysAgo(41), daysAgo(40))...)
		case i%10 == 0: // old, goes
			logFile(t, l, name, rec{event: "guard", at: daysAgo(40).Add(time.Duration(i) * time.Second), session: id})
			goes = append(goes, target(Log, name))
		default: // new: kept without a read
			if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	goes = sorted(goes)
	all := names(t, l, "log")
	o := opts(l)
	o.Budget = budget

	// A listing slowed to a floor of 1 ms an entry: listed whole at once, the folder would take 4 s. The run stops at
	// the first batch past its budget, having judged nothing.
	read := readBatch
	t.Cleanup(func() { readBatch = read })
	readBatch = func(fh *os.File, n int) ([]fs.DirEntry, error) {
		b, err := read(fh, n)
		time.Sleep(time.Duration(len(b)) * time.Millisecond)
		return b, err
	}
	start := time.Now()
	d := SessionEnd(o)
	took := time.Since(start)
	readBatch = read
	if took > budget+stepCeiling {
		t.Errorf("a run with a budget of %v on a slow listing took %v, past the budget and %v", budget, took, stepCeiling)
	}
	if !d.Stopped || len(d.Cleaned) != 0 || len(d.Errs) != 0 {
		t.Fatalf("the slow listing: stopped %v, cleaned %d, errors %v", d.Stopped, len(d.Cleaned), d.Errs)
	}
	if got := names(t, l, "log"); !same(got, all) {
		t.Fatalf("the slow listing changed the folder: %d entries, want %d", len(got), len(all))
	}
	t.Logf("a slow listing of %d files: stopped in %v (budget %v)", files, took, budget)

	// The listing at its own pace, each delete slowed to a floor of 1 ms, so no run of 100 ms can clean all 300.
	slow := removeFile
	t.Cleanup(func() { removeFile = slow })
	removeFile = func(name string) error {
		time.Sleep(time.Millisecond)
		return slow(name)
	}
	start = time.Now()
	d = SessionEnd(o)
	took = time.Since(start)
	removeFile = slow
	if took > budget+stepCeiling {
		t.Errorf("a run with a budget of %v took %v, past the budget and %v", budget, took, stepCeiling)
	}
	if !d.Stopped || len(d.Errs) != 0 || len(d.Busy) != 0 {
		t.Fatalf("stopped %v, errors %v, busy %v", d.Stopped, d.Errs, d.Busy)
	}
	// Only what it judged: each file gone is one that goes, each with its record, and every other file is as it was.
	cleaned := targets(d.Cleaned)
	if got := recordTargets(t, l); !same(got, cleaned) {
		t.Errorf("the records name %d files, the run cleaned %d", len(got), len(cleaned))
	}
	want := map[string]bool{}
	for _, g := range goes {
		want[g] = true
	}
	gone := map[string]bool{}
	for _, c := range cleaned {
		if !want[c] {
			t.Errorf("%s was cleaned, and it does not go", c)
		}
		gone[strings.TrimPrefix(c, ".bonsai/local/log/")] = true
	}
	var left []string
	for _, n := range all {
		if !gone[n] {
			left = append(left, n)
		}
	}
	day := "w-" + now.Format("2006-01-02") + ".ndjson" // the clean records' day file
	if len(cleaned) > 0 {
		left = sorted(append(left, day))
	}
	if got := names(t, l, "log"); !same(got, left) {
		t.Errorf("the folder holds %d entries, want %d", len(got), len(left))
	}
	t.Logf("%d files: the run cleaned %d in %v (budget %v)", files, len(cleaned), took, budget)

	// A run with time finishes the job: every file that goes is gone, with its record; the rest stay.
	o.Budget = 0
	d = Files(o, Ladder, Asks, Log)
	if d.Stopped || len(d.Errs) != 0 {
		t.Fatalf("the run with time: stopped %v, errors %v", d.Stopped, d.Errs)
	}
	if got := recordTargets(t, l); !same(got, goes) {
		t.Errorf("after the run with time, %d records, want %d", len(got), len(goes))
	}
	if got := names(t, l, "log"); len(got) != files-len(goes)+1 {
		t.Errorf("after the run with time, the folder holds %d entries, want %d", len(got), files-len(goes)+1)
	}
}

// The listing stops at the budget between two batches: a clock that moves on 1 ms a batch and a budget of 2 ms read
// exactly two of a folder's three batches, and nothing is judged, deleted or noted as an error. The next run goes on.
func TestBudgetStopsTheListing(t *testing.T) {
	l := project(t, "  log:\n    keep_days: 30\n")
	all := oldLogFiles(t, l, 2*listBatch+50)
	advance := heldClock(t)
	batches, ages := 0, 0
	read, age := readBatch, logKind.age
	t.Cleanup(func() { readBatch, logKind.age = read, age })
	readBatch = func(fh *os.File, n int) ([]fs.DirEntry, error) {
		batches++
		advance(time.Millisecond)
		return read(fh, n)
	}
	logKind.age = func(r *run, f *file) time.Time {
		ages++
		return age(r, f)
	}
	o := opts(l)
	o.Budget = 2 * time.Millisecond
	d := Files(o, Log)
	if !d.Stopped || len(d.Cleaned) != 0 || len(d.Errs) != 0 || batches != 2 || ages != 0 {
		t.Errorf("stopped %v, cleaned %d, errors %v; %d batches read, want 2; %d ages read, want 0",
			d.Stopped, len(d.Cleaned), d.Errs, batches, ages)
	}
	if got := names(t, l, "log"); len(got) != len(all) {
		t.Errorf("the folder holds %d, want %d", len(got), len(all))
	}
	o.Budget = 0
	if d := Files(o, Log); d.Stopped || len(d.Cleaned) != len(all) || len(d.Errs) != 0 {
		t.Errorf("the next run: stopped %v, cleaned %d of %d, errors %v", d.Stopped, len(d.Cleaned), len(all), d.Errs)
	}
}

// keep_newest on a folder it cannot rank within the budget deletes nothing of that kind, by either rule: the clock
// moves on 1 ms an age read, and a budget of 50 ms ranks 50 of 300 files and stops. A run with time ranks them all and
// cleans.
func TestKeepNewestUnrankedDeletesNothing(t *testing.T) {
	for _, c := range []struct {
		name, rule string
		cleaned    int
		reason     string
	}{
		{"keep_newest alone", "  log:\n    keep_days: null\n    keep_newest: 10\n", 290, "generated.log.keep_newest=10"},
		{"with keep_days", "  log:\n    keep_days: 30\n    keep_newest: 10\n", 300, "generated.log.keep_days=30"},
	} {
		t.Run(c.name, func(t *testing.T) {
			l := project(t, c.rule)
			all := oldLogFiles(t, l, 300)
			advance := heldClock(t)
			ages := 0
			age := logKind.age
			t.Cleanup(func() { logKind.age = age })
			logKind.age = func(r *run, f *file) time.Time {
				ages++
				advance(time.Millisecond)
				return age(r, f)
			}
			o := opts(l)
			o.Budget = 50 * time.Millisecond
			d := Files(o, Log)
			if !d.Stopped || len(d.Cleaned) != 0 || len(d.Errs) != 0 || ages != 50 {
				t.Errorf("stopped %v, cleaned %d, errors %v, %d ages read, want 50", d.Stopped, len(d.Cleaned), d.Errs, ages)
			}
			if got := names(t, l, "log"); !same(got, sorted(all)) {
				t.Errorf("the folder holds %d, want all %d", len(got), len(all))
			}
			if got := cleanRecords(t, l); len(got) != 0 {
				t.Errorf("records %v", got)
			}
			o.Budget = 0
			d = Files(o, Log)
			if d.Stopped || len(d.Errs) != 0 || len(d.Cleaned) != c.cleaned {
				t.Fatalf("the run with time: stopped %v, cleaned %d, want %d, errors %v", d.Stopped, len(d.Cleaned), c.cleaned, d.Errs)
			}
			for _, it := range d.Cleaned {
				if it.Reason != c.reason {
					t.Errorf("%s: reason %s, want %s", it.Target, it.Reason, c.reason)
				}
			}
		})
	}
}

// A ranking that is whole and a judging the budget cuts short: the run deletes only files ranked beyond the newest,
// oldest first. The clock moves on 1 ms an age read and 1 ms a delete; a budget of 320 ms ranks all 300 and deletes 20.
func TestKeepNewestStoppedInTheJudging(t *testing.T) {
	l := project(t, "  log:\n    keep_days: null\n    keep_newest: 10\n")
	all := oldLogFiles(t, l, 300)
	advance := heldClock(t)
	age, del := logKind.age, removeFile
	t.Cleanup(func() { logKind.age, removeFile = age, del })
	logKind.age = func(r *run, f *file) time.Time {
		advance(time.Millisecond)
		return age(r, f)
	}
	removeFile = func(p string) error {
		advance(time.Millisecond)
		return del(p)
	}
	o := opts(l)
	o.Budget = 320 * time.Millisecond
	d := Files(o, Log)
	if !d.Stopped || len(d.Errs) != 0 || len(d.Cleaned) != 20 {
		t.Fatalf("stopped %v, cleaned %d, want 20, errors %v", d.Stopped, len(d.Cleaned), d.Errs)
	}
	for i, it := range d.Cleaned {
		if want := target(Log, all[i]); it.Target != want || it.Reason != "generated.log.keep_newest=10" {
			t.Errorf("cleaned %d: %s (%s), want %s, the oldest first", i, it.Target, it.Reason, want)
		}
	}
}

// openKeys are the asks that read open.
func openKeys(t *testing.T, l workspace.Local) []string {
	t.Helper()
	book, err := asks.Read(l.Main, "")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range book.Entries(false) {
		out = append(out, e.Key)
	}
	return sorted(out)
}

// folderTargets are the entries of the three file kinds' folders, as targets.
func folderTargets(t *testing.T, l workspace.Local) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	for _, folder := range []string{"log", "asks", "ladder"} {
		for _, n := range names(t, l, folder) {
			out[target(folder, n)] = true
		}
	}
	return out
}

// A run stopped anywhere leaves a whole state: on the fixture of every kind, a sweep of budgets stops a session end's
// run at each of its checks in turn (the listing, the asks' history, the judging, the asks' deletes). Each stopped run
// notes no error, deletes only what an uninterrupted run deletes, each file gone with its record and nothing else
// changed, and no ask reads open that did not; a run with time then leaves exactly what an uninterrupted run leaves.
func TestStoppedAnywhereLeavesAWholeState(t *testing.T) {
	l := project(t, allRules)
	fixture(t, l)
	if d := SessionEnd(opts(l)); d.Stopped || len(d.Errs) != 0 {
		t.Fatalf("the uninterrupted run: stopped %v, errors %v", d.Stopped, d.Errs)
	}
	full := cleanRecords(t, l)
	fullSet := map[string]bool{}
	for _, r := range full {
		fullSet[r] = true
	}
	tickingClock(t)
	day := target(Log, "w-"+now.Format("2006-01-02")+".ndjson")
	stops := 0
	for b := 1; ; b++ {
		if b > 10000 {
			t.Fatal("no budget of up to 10 s let the run finish")
		}
		l := project(t, allRules)
		fixture(t, l)
		before, open := folderTargets(t, l), openKeys(t, l)
		o := opts(l)
		o.Budget = time.Duration(b) * time.Millisecond
		d := SessionEnd(o)
		if len(d.Errs) != 0 || len(d.Busy) != 0 {
			t.Fatalf("budget %d: errors %v, busy %v", b, d.Errs, d.Busy)
		}
		records := cleanRecords(t, l)
		for _, r := range records {
			if !fullSet[r] {
				t.Errorf("budget %d: %s, which an uninterrupted run does not clean", b, r)
			}
		}
		if got, want := recordTargets(t, l), targets(d.Cleaned); !same(got, want) {
			t.Errorf("budget %d: records %v, cleaned %v", b, got, want)
		}
		cleaned := map[string]bool{}
		for _, it := range d.Cleaned {
			cleaned[it.Target] = true
		}
		after := folderTargets(t, l)
		for p := range before {
			if !after[p] && !cleaned[p] {
				t.Errorf("budget %d: %s is gone with no record", b, p)
			}
		}
		for p := range after {
			if !before[p] && p != day {
				t.Errorf("budget %d: %s is new", b, p)
			}
		}
		for p := range cleaned {
			if after[p] {
				t.Errorf("budget %d: %s was cleaned and is still there", b, p)
			}
		}
		if got := openKeys(t, l); !same(got, open) {
			t.Errorf("budget %d: the open asks are %v, were %v", b, got, open)
		}
		o.Budget = 0
		if d := Files(o, Ladder, Asks, Log); d.Stopped || len(d.Errs) != 0 {
			t.Fatalf("budget %d, then a run with time: stopped %v, errors %v", b, d.Stopped, d.Errs)
		}
		if got := cleanRecords(t, l); !same(got, full) {
			t.Errorf("budget %d, then a run with time: records\n  %s\nwant\n  %s", b, strings.Join(got, "\n  "),
				strings.Join(full, "\n  "))
		}
		if !d.Stopped {
			break
		}
		stops++
	}
	t.Logf("%d places a run can stop, each left whole", stops)
}

// Asks go in day order: the earlier day's file, holding a filing, touched later than the next day's, holding its
// answer (a copy, a restore). By modification time the answer's file would go first, and a run stopped between the
// two would leave the filing without its answer: the ask would read open again. A sweep of budgets stops the run at
// each of its checks.
func TestAsksGoInDayOrder(t *testing.T) {
	tickingClock(t)
	for b := 1; ; b++ {
		if b > 10000 {
			t.Fatal("no budget of up to 10 s let the run finish")
		}
		l := project(t, "  asks:\n    keep_days: 30\n")
		filed := asksFile(t, l, "2026-08-01", askLine(t, "file", "agent:k", daysAgo(70)))
		asksFile(t, l, "2026-08-02", askLine(t, "answer", "agent:k", daysAgo(69)))
		stamp(t, filed, daysAgo(50))
		o := opts(l)
		o.Budget = time.Duration(b) * time.Millisecond
		d := Files(o, Asks)
		if len(d.Errs) != 0 {
			t.Fatalf("budget %d: errors %v", b, d.Errs)
		}
		if got := openKeys(t, l); len(got) != 0 {
			t.Fatalf("budget %d: %v reads open again (cleaned %v)", b, got, targets(d.Cleaned))
		}
		if !d.Stopped {
			if len(d.Cleaned) != 2 || d.Cleaned[0].Target != ".bonsai/local/asks/2026-08-01.ndjson" {
				t.Errorf("cleaned %v, want both, the earlier day first", d.Cleaned)
			}
			break
		}
	}
}

// An asks day file that does not go (held open, here) stops that run's asks: the next day's file, holding the answer
// to its filing, stays with it, so the ask does not read open again. A later run cleans both. The hold is stood in
// on every system: a real hold on Windows refuses reads too, and then the asks are not judged at all.
func TestAnAsksFileThatStaysStopsTheAsks(t *testing.T) {
	l := project(t, "  asks:\n    keep_days: 30\n")
	asksFile(t, l, "2026-08-01", askLine(t, "file", "agent:k", daysAgo(70)))
	asksFile(t, l, "2026-08-02", askLine(t, "answer", "agent:k", daysAgo(69)))
	errAskHeld := errors.New("the process cannot access the file because it is being used by another process")
	del, isBusy := removeFile, busy
	t.Cleanup(func() { removeFile, busy = del, isBusy })
	removeFile = func(p string) error {
		if filepath.Base(p) == "2026-08-01.ndjson" {
			return errAskHeld
		}
		return del(p)
	}
	busy = func(err error) bool { return errors.Is(err, errAskHeld) || isBusy(err) }
	d := Files(opts(l), Asks)
	if len(d.Cleaned) != 0 || len(d.Busy) != 1 || len(d.Errs) != 0 {
		t.Fatalf("cleaned %v, busy %v, errors %v", d.Cleaned, d.Busy, d.Errs)
	}
	if got := openKeys(t, l); len(got) != 0 {
		t.Errorf("%v reads open again", got)
	}
	removeFile, busy = del, isBusy
	d = Files(opts(l), Asks)
	if len(d.Cleaned) != 2 || len(d.Busy) != 0 || len(d.Errs) != 0 {
		t.Errorf("the later run: cleaned %v, busy %v, errors %v", d.Cleaned, d.Busy, d.Errs)
	}
}
