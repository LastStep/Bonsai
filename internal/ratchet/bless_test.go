package ratchet

// The Bless step (design/plan-5.md, 5.4.4 notes 2, 3 and 5) on a main checkout's folder (git's answers given in
// Where): never under --ci nor on a branch; on the base branch the climb's own offer filed, a second filing replacing
// the first's numbers; every other open Bless ask worked out again from its task's result and the floors now, and
// resolved when HEAD moved, its climb went red or dirty, no rise is left or its result is gone, one whose offer stands
// left open; the climb's own open ask resolved when its climb went red; every record as written held to bonsai.ask/1.

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const testSession = "6d1e2f3a-4b5c-4d6e-8f7a-9b0c1d2e3f4a"

// mainPlace is a main checkout's folder and where its asks go, with a clock a second a call from 2026-10-13 10:00 UTC.
func mainPlace(t *testing.T) asks.Place {
	t.Helper()
	dir := t.TempDir()
	tick := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)
	return asks.Place{Workspace: testWS, Local: workspace.Local{Root: dir, Main: dir, Branch: "main"}, Session: testSession,
		Now: func() time.Time { tick = tick.Add(time.Second); return tick }}
}

// writeResult writes l as task's result in main's ladder folder, and gives it as a Result.
func writeResult(t *testing.T, main, task string, l *format.Ladder) Result {
	t.Helper()
	raw, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(main, ".bonsai", "local", "ladder", task+".json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return readResult(main, task)
}

// fileOffer files task's Bless ask as a climb on main at headA did, from its result written.
func fileOffer(t *testing.T, p asks.Place, task string) {
	t.Helper()
	r := writeResult(t, p.Local.Main, task, greenResult(task, rung(1, "green", "go_tests", "812", "815")))
	o, why := OfferFor(task, r, here())
	if o == nil {
		t.Fatal(why)
	}
	if _, e := asks.FileBless(p, asks.Bless{Task: task, Title: Title(task), Why: Why(o), Then: Then, Data: Data(o)}); e != nil {
		t.Fatal(e)
	}
}

// openKeys are the open Bless asks' keys, sorted.
func openKeys(t *testing.T, main string) string {
	t.Helper()
	b, err := asks.Read(main, "")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range b.Entries(false) {
		out = append(out, e.Key)
	}
	sort.Strings(out)
	return strings.Join(out, " ")
}

// ops are the records' ops and keys: "file ladder:bless-T-0901".
func ops(records []schema.Object) string {
	var out []string
	for _, r := range records {
		out = append(out, r.String("op")+" "+r.String("key"))
	}
	return strings.Join(out, ", ")
}

// fitsAsk holds each record to bonsai.ask/1 as a writer writes it: every field, in order.
func fitsAsk(t *testing.T, records []schema.Object) {
	t.Helper()
	s := format.MustLookup("ask").Schema()
	for _, r := range records {
		if msgs := append(schema.Validate(s, r), schema.CheckOrder(s, r)...); len(msgs) > 0 {
			t.Errorf("not bonsai.ask/1 as written: %v\n%s", msgs, schema.Show(r))
		}
	}
}

// Never under --ci, nor on a branch: nothing filed or resolved; on a branch, a climb that rose says why it was not
// offered.
func TestBlessNever(t *testing.T) {
	p := mainPlace(t)
	fileOffer(t, p, "T-0800")
	own := writeResult(t, p.Local.Main, "T-0901", greenResult("T-0901", rung(1, "green", "go_tests", "812", "815")))
	w := here()
	w.Head = headB // T-0800's offer is gone, but these climbs resolve nothing
	if d := Bless(Step{Place: p, CI: true, Task: "T-0901", Result: own, Where: w}); len(d.Records)+len(d.Lines)+len(d.Notes) != 0 {
		t.Errorf("--ci: %+v", d)
	}
	branch := greenResult("T-0901", rung(1, "green", "go_tests", "812", "815"))
	branch.Git.Branch = "t0901"
	d := Bless(Step{Place: p, Task: "T-0901", Result: Result{Ladder: branch, SHA256: resultSH}, Where: w})
	if len(d.Records) != 0 || len(d.Notes) != 0 || len(d.Lines) != 1 ||
		d.Lines[0] != `bless: not offered: the last climb for T-0901 was on "t0901", not the base branch main` {
		t.Errorf("a branch: %+v", d)
	}
	// A detached main checkout has no base branch: nothing either.
	w.Base, w.Branch = "", ""
	if d := Bless(Step{Place: p, Task: "T-0901", Result: own, Where: w}); len(d.Records) != 0 {
		t.Errorf("no base: %+v", d)
	}
	// A branch climb that rose nothing says nothing.
	flat := greenResult("T-0901", rung(1, "green", "go_tests", "812", "812"))
	flat.Git.Branch = "t0901"
	if d := Bless(Step{Place: p, Task: "T-0901", Result: Result{Ladder: flat}, Where: here()}); len(d.Lines)+len(d.Records) != 0 {
		t.Errorf("a flat branch climb: %+v", d)
	}
	if got := openKeys(t, p.Local.Main); got != "ladder:bless-T-0800" {
		t.Errorf("open: %s", got)
	}
}

// On the base branch: the climb's own offer filed, held to bonsai.ask/1 with contract §9.1's data; a later climb's
// numbers replace the first's; a red climb resolves its own task's ask, files nothing.
func TestBlessOwn(t *testing.T) {
	p := mainPlace(t)
	r := writeResult(t, p.Local.Main, "T-0901", greenResult("T-0901", rung(1, "green", "go_tests", "812", "815")))
	d := Bless(Step{Place: p, Task: "T-0901", Result: r, Where: here()})
	if ops(d.Records) != "file ladder:bless-T-0901" || len(d.Notes) != 0 || len(d.Lines) != 1 ||
		d.Lines[0] != "bless: offered ladder:bless-T-0901 (go_tests 812 -> 815): a person blesses it, or raises the floor in "+
			"bonsai.yaml's ratchets to this climb's count" {
		t.Fatalf("%+v", d)
	}
	fitsAsk(t, d.Records)
	rec := d.Records[0]
	want := `{"task":"T-0901","head":"` + headA + `","result_sha":"` + r.SHA256 + `","rises":[{"name":"go_tests","was":812,"now":815,"new_tests":null}]}`
	if data, _ := rec.Get("data"); schema.Show(data) != want || rec.String("source") != "ladder" || rec.String("type") != "Bless" ||
		rec.String("task") != "T-0901" || rec.String("title") != "Bless T-0901's ratchets" || rec.String("then_text") != Then ||
		rec.String("session") != testSession ||
		rec.String("why") != "A green, clean climb of main at 5d6e7f80: go_tests rose from 812 to 815, its new tests not checked on the base." {
		t.Errorf("the record: %s", schema.Show(rec))
	}
	// A later climb at the same HEAD with a higher count: a new file record, whose numbers stand.
	r = writeResult(t, p.Local.Main, "T-0901", greenResult("T-0901", rung(1, "green", "go_tests", "812", "820")))
	if d := Bless(Step{Place: p, Task: "T-0901", Result: r, Where: here()}); ops(d.Records) != "file ladder:bless-T-0901" {
		t.Fatalf("%+v", d)
	}
	b, _ := asks.Read(p.Local.Main, "")
	if e := b.Get("ladder:bless-T-0901"); e.State != asks.Open || !strings.Contains(schema.Show(e.Filed.Data), `"now":820`) {
		t.Errorf("the second filing does not stand: %s", schema.Show(e.Filed.Data))
	}
	// The same task climbs red: its ask resolved with the reason, nothing filed.
	red := greenResult("T-0901", rung(1, "red", "go_tests", "812", "805"))
	red.Green = false
	d = Bless(Step{Place: p, Task: "T-0901", Result: Result{Ladder: red, SHA256: resultSH}, Where: here()})
	if ops(d.Records) != "resolve ladder:bless-T-0901" || len(d.Lines) != 1 ||
		d.Lines[0] != "bless: resolved ladder:bless-T-0901: its offer is gone: the last climb for T-0901 was not green" {
		t.Fatalf("%+v", d)
	}
	fitsAsk(t, d.Records)
	// Red again with nothing open, and a climb that rose nothing: nothing written, nothing said.
	if d := Bless(Step{Place: p, Task: "T-0901", Result: Result{Ladder: red}, Where: here()}); len(d.Records)+len(d.Lines) != 0 {
		t.Errorf("%+v", d)
	}
	flat := greenResult("T-0901", rung(1, "green", "go_tests", "812", "812"))
	if d := Bless(Step{Place: p, Task: "T-0901", Result: Result{Ladder: flat}, Where: here()}); len(d.Records)+len(d.Lines) != 0 {
		t.Errorf("%+v", d)
	}
	// A rise not offered on the base branch (its name not a key of ratchets) says why.
	other := greenResult("T-0901", rung(1, "green", "other", "0", "9"))
	if d := Bless(Step{Place: p, Task: "T-0901", Result: Result{Ladder: other}, Where: here()}); len(d.Records) != 0 || len(d.Lines) != 1 ||
		!strings.HasPrefix(d.Lines[0], "bless: not offered: the last climb for T-0901 measured no ratchet above its floor") {
		t.Errorf("%+v", d)
	}
}

// A climb on the base branch, whatever its task, works every other open Bless ask out again from its task's result and
// the floors now: resolved when HEAD moved, its climb went red or dirty, no rise is left, its result is gone; left open
// when its offer stands. The records come in key order, then the climb's own.
func TestBlessOthers(t *testing.T) {
	p := mainPlace(t)
	main := p.Local.Main
	for _, task := range []string{"T-0800", "T-0801", "T-0802", "T-0803", "T-0804", "T-0805", "T-0806"} {
		fileOffer(t, p, task)
	}
	moved := greenResult("T-0800", rung(1, "green", "go_tests", "812", "815"))
	moved.Git.SHA = headB
	writeResult(t, main, "T-0800", moved)
	red := greenResult("T-0801", rung(1, "green", "go_tests", "812", "815"))
	red.Green = false
	writeResult(t, main, "T-0801", red)
	dirty := greenResult("T-0802", rung(1, "green", "go_tests", "812", "815"))
	dirty.Git.Dirty = true
	writeResult(t, main, "T-0802", dirty)
	writeResult(t, main, "T-0803", greenResult("T-0803", rung(1, "green", "go_tests", "812", "813")))
	if err := os.Remove(filepath.Join(main, ".bonsai", "local", "ladder", "T-0804.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(main, ".bonsai", "local", "ladder", "T-0806.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The floors as they now are: go_tests raised to 813, so T-0803's 813 is no rise; T-0805's 815 still is.
	w := here()
	w.Floors = floors("go_tests", "813")
	own := writeResult(t, main, "T-0901", greenResult("T-0901", rung(1, "green", "go_tests", "813", "816")))
	d := Bless(Step{Place: p, Task: "T-0901", Result: own, Where: w})
	if got := ops(d.Records); got != "resolve ladder:bless-T-0800, resolve ladder:bless-T-0801, resolve ladder:bless-T-0802, "+
		"resolve ladder:bless-T-0803, resolve ladder:bless-T-0804, resolve ladder:bless-T-0806, file ladder:bless-T-0901" {
		t.Fatalf("records: %s\n%s", got, strings.Join(d.Lines, "\n"))
	}
	fitsAsk(t, d.Records)
	want := []string{
		"bless: resolved ladder:bless-T-0800: its offer is gone: the last climb for T-0800 was made at 0f1e2d3c, and HEAD is 5d6e7f80 (HEAD moved)",
		"bless: resolved ladder:bless-T-0801: its offer is gone: the last climb for T-0801 was not green",
		"bless: resolved ladder:bless-T-0802: its offer is gone: the last climb for T-0802 was dirty (a dirty climb proves no commit)",
		"bless: resolved ladder:bless-T-0803: its offer is gone: the last climb for T-0803 measured no ratchet above its floor in bonsai.yaml's ratchets",
		"bless: resolved ladder:bless-T-0804: its offer is gone: there is no ladder result for T-0804 (.bonsai/local/ladder/T-0804.json)",
		"bless: resolved ladder:bless-T-0806: its offer is gone: .bonsai/local/ladder/T-0806.json is not a ladder result (",
		"bless: offered ladder:bless-T-0901 (go_tests 813 -> 816)",
	}
	if len(d.Lines) != len(want) {
		t.Fatalf("lines:\n%s", strings.Join(d.Lines, "\n"))
	}
	for i, l := range d.Lines {
		if !strings.HasPrefix(l, want[i]) {
			t.Errorf("line %d:\n%s\nwant\n%s", i, l, want[i])
		}
	}
	if got := openKeys(t, main); got != "ladder:bless-T-0805 ladder:bless-T-0901" {
		t.Errorf("open: %s", got)
	}
	// A record written by hand under the prefix with a key Bonsai never makes (a path) is no Bless ask: left alone.
	line := `{"format":"bonsai.ask/1","id":"9c8b7a6f-5e4d-4c3b-8a29-1f0e9d8c7b6a","at":"2026-10-13T10:00:00.000Z","op":"file",` +
		`"key":"ladder:bless-../../x","workspace":"` + testWS + `","source":"ladder","session":null,"type":"Bless","task":null,"doc":null,` +
		`"title":"t","why":"w","then_text":null,"options":[],"verdict":null,"data":null,"answer":null}` + "\n"
	f, err := os.OpenFile(filepath.Join(asks.Folder(main), "2026-10-13.ndjson"), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(line); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	// The next climb with nothing changed resolves nothing more: the standing offers stay, the hand-written one too.
	if d := Bless(Step{Place: p, Task: "T-0902", Result: writeResult(t, main, "T-0902", greenResult("T-0902")), Where: w}); len(d.Records) != 0 ||
		len(d.Notes) != 0 || len(d.Lines) != 0 {
		t.Errorf("a second climb: %s %v %v", ops(d.Records), d.Lines, d.Notes)
	}
	if got := openKeys(t, main); got != "ladder:bless-../../x ladder:bless-T-0805 ladder:bless-T-0901" {
		t.Errorf("open: %s", got)
	}
}

// What the step cannot do is a note, never a panic: the asks that cannot be read, a task that cannot make a key.
func TestBlessNotes(t *testing.T) {
	p := mainPlace(t)
	own := writeResult(t, p.Local.Main, "T-0901", greenResult("T-0901", rung(1, "green", "go_tests", "812", "815")))
	if err := os.MkdirAll(filepath.Join(p.Local.Main, ".bonsai", "local"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(asks.Folder(p.Local.Main), []byte("a file, not a folder"), 0o644); err != nil {
		t.Fatal(err)
	}
	d := Bless(Step{Place: p, Task: "T-0901", Result: own, Where: here()})
	// A file where the asks folder goes cannot be read on Linux; on Windows it reads as no folder (Go takes the system's
	// "path not found" for not-there), so there the filing's note is the only one.
	want := []string{"the open Bless asks cannot be read (", "the Bless ask ladder:bless-T-0901 cannot be filed: "}
	if runtime.GOOS == "windows" {
		want = want[1:]
	}
	if len(d.Records) != 0 || len(d.Notes) != len(want) {
		t.Fatalf("%+v", d)
	}
	for i, w := range want {
		if !strings.HasPrefix(d.Notes[i], w) {
			t.Errorf("note %d: %q, want %q", i, d.Notes[i], w)
		}
	}
	p = mainPlace(t)
	odd := greenResult("T 0901", rung(1, "green", "go_tests", "812", "815"))
	d = Bless(Step{Place: p, Task: "T 0901", Result: Result{Ladder: odd, SHA256: resultSH}, Where: here()})
	if len(d.Records) != 0 || len(d.Notes) != 1 || !strings.Contains(d.Notes[0], "cannot make a Bless ask's key") {
		t.Errorf("%+v", d)
	}
}

// The climb's own offer carries its new-tests lines into the ask's why; the data keeps the counts.
func TestBlessChecked(t *testing.T) {
	p := mainPlace(t)
	r := rung(1, "green", "go_tests", "812", "815")
	r.Ratchet.NewTests = &format.LadderNewTests{Base: headB, Count: 1, FailedOnBase: 1}
	own := writeResult(t, p.Local.Main, "T-0901", greenResult("T-0901", r))
	own.Checks = []string{"1 new test, run on the base 0f1e2d3c4b5a: 1 failed there, 0 prove nothing new, 0 check not run"}
	d := Bless(Step{Place: p, Task: "T-0901", Result: own, Where: here()})
	if ops(d.Records) != "file ladder:bless-T-0901" || len(d.Notes) != 0 {
		t.Fatalf("%+v", d)
	}
	fitsAsk(t, d.Records)
	rec := d.Records[0]
	data, _ := rec.Get("data")
	if rec.String("why") != "A green, clean climb of main at 5d6e7f80: go_tests rose from 812 to 815, "+own.Checks[0]+"." ||
		!strings.Contains(schema.Show(data), `"new_tests":{"base":"`+headB+`","count":1,"failed_on_base":1,"passed_on_base":0,"not_run":0}`) {
		t.Errorf("the ask: %s", schema.Show(rec))
	}
}
