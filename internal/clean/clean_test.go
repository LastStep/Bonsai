package clean

// The fixture project of every file kind (design/plan-5.md, 5.2.6b's proof; "5.2 done" check 14): old, new and
// protected files of each, and a decoy of each (a name Bonsai does not write, a sub-folder, on Linux a link). Exactly
// the unprotected old ones go, each with one clean record whose target and reason are right; a second run removes
// nothing and writes no record.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const allRules = `  log:
    keep_days: 30
    keep_newest: null
  asks:
    keep_days: 30
    keep_newest: null
  ladder:
    keep_days: 7
    keep_newest: null
  run:
    keep_days: null
    keep_newest: null
  sessions:
    keep_days: 30
    keep_newest: null
`

// sid is a made-up session id whose first 8 characters are its row's.
func sid(n string) string { return "a000000" + n + "-0000-4000-8000-000000000000" }

// fixture writes the project's files; links reports whether the Linux-only links were made.
func fixture(t *testing.T, l workspace.Local) (links bool) {
	t.Helper()
	old, older := daysAgo(40), daysAgo(41)
	// The log.
	logFile(t, l, "s-"+sid("1")+".ndjson", session(sid("1"), "T-0001", older, old)...) // ended, its row in: goes
	logFile(t, l, "s-"+sid("2")+".ndjson", session(sid("2"), "none", older, old)...)   // ended, no row: kept
	three := append(session(sid("3"), "none", older, old)[:2],
		rec{event: "subagent_start", at: older.Add(2 * time.Minute), session: sid("3"), subagent: "b0000003-x"},
		rec{event: "subagent_stop", at: older.Add(3 * time.Minute), session: sid("3"), subagent: "b0000003-x"},
		rec{event: "session_end", at: old, session: sid("3")})
	logFile(t, l, "s-"+sid("3")+".ndjson", three...) // its session's row in, its subagent run's not: kept
	logFile(t, l, "s-"+sid("4")+".ndjson",           // no span at all (a guard's records only): goes
		rec{event: "guard", at: older, session: sid("4")}, rec{event: "guard", at: old, session: sid("4")})
	logFile(t, l, "s-"+sid("5")+".ndjson", session(sid("5"), "none", daysAgo(6), daysAgo(5))...) // new: stays
	logFile(t, l, "s-"+sid("6")+".ndjson",                                                       // open, two hours old: stays
		rec{event: "session_start", at: now.Add(-2 * time.Hour), session: sid("6"), source: "startup"},
		rec{event: "tool_end", at: now.Add(-time.Hour), session: sid("6")})
	logFile(t, l, "w-2026-08-01.ndjson", rec{event: "event", at: time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)}) // old day: goes
	logFile(t, l, "w-2026-10-09.ndjson", rec{event: "event", at: daysAgo(1)})                                  // new day: stays
	logFile(t, l, "s-abc.ndjson", session("abc", "none", older, old)...)                                       // a session id no row can hold: goes
	torn := logFile(t, l, "s-"+sid("7")+".ndjson", session(sid("7"), "T-0001", older, old)...)                 // a torn last line: goes
	appendTo(t, torn, `{"format":"bonsai.log/1","id":"torn`)
	stamp(t, torn, old)
	sessionsTable(t, l, row(sid("1"), "T-0001", older, old), row(sid("3"), "none", older, old),
		row(sid("5"), "none", daysAgo(6), daysAgo(5)), row(sid("7"), "T-0001", older, old))
	// The log's decoys, each old and holding no span, so each would go were it Bonsai's.
	guards := []rec{{event: "guard", at: older, session: sid("8")}, {event: "guard", at: old, session: sid("8")}}
	for _, name := range []string{"notes.txt", "s-" + sid("8") + ".ndjson.bak", "w-2026-02-30.ndjson", "S-" + sid("8") + ".ndjson"} {
		logFile(t, l, name, guards...)
	}
	logFile(t, l, "old/s-"+sid("8")+".ndjson", guards...)
	mkdir(t, l, ".bonsai/local/log/s-"+sid("9")+".ndjson")
	// The asks: A open, filed with D; B filed and answered; D answered later; E filed, answered after the cutoff.
	asksFile(t, l, "2026-08-01", askLine(t, "file", "agent:a", daysAgo(70)), askLine(t, "file", "agent:d", daysAgo(70)))
	asksFile(t, l, "2026-08-02", askLine(t, "file", "agent:b", daysAgo(69)), askLine(t, "answer", "agent:b", daysAgo(69)))
	asksFile(t, l, "2026-08-05", askLine(t, "answer", "agent:d", daysAgo(66)))
	asksFile(t, l, "2026-08-06", askLine(t, "file", "agent:e", daysAgo(65)))
	asksFile(t, l, "2026-10-08", askLine(t, "answer", "agent:e", daysAgo(2)))
	asksFile(t, l, "2026-10-09", askLine(t, "file", "agent:f", daysAgo(1)))
	for _, name := range []string{"2026-08-07.ndjson.tmp", "notes.ndjson", "2026-13-45.ndjson"} {
		p := write(t, l.Main, ".bonsai/local/asks/"+name, string(askLine(t, "file", "agent:x", daysAgo(64))))
		stamp(t, p, daysAgo(64))
	}
	mkdir(t, l, ".bonsai/local/asks/2026-08-08.ndjson")
	// The ladder and the tasks.
	for id, status := range map[string]string{"T-0001": "done", "T-0002": "running", "T-0004": "cut", "T-0005": "done",
		"T-0006": "done", "T-0007": "broken"} {
		taskFile(t, l, id, status)
	}
	for _, id := range []string{"T-0001", "T-0002", "T-0003", "T-0004", "T-0007"} {
		ladderFile(t, l, id+".json", ptr(id), daysAgo(20))
	}
	ladderFile(t, l, "T-0005.json", ptr("T-0005"), daysAgo(2))  // new: stays
	ladderFile(t, l, "T-0006.json", ptr("T-0002"), daysAgo(20)) // its task field names a running task: kept
	for _, name := range []string{"ci.json", "T-0008.json.bak", "notes.json", "t-0009.json"} {
		ladderFile(t, l, name, nil, daysAgo(20))
	}
	taskFile(t, l, "T-0010", "done") // a folder with a done task's result's name: an empty folder a delete would take
	mkdir(t, l, ".bonsai/local/ladder/T-0010.json")
	stamp(t, filepath.Join(l.Main, ".bonsai", "local", "ladder", "T-0010.json"), daysAgo(20))
	// On Linux, a link of each kind's name to an old file outside the project, which must stay as it is.
	if runtime.GOOS != "windows" {
		outside := filepath.Join(filepath.Dir(l.Main), "outside")
		for _, p := range []struct{ folder, name, src string }{
			{"log", "s-" + sid("9") + "x.ndjson", "s-" + sid("1") + ".ndjson"},
			{"asks", "2026-08-03.ndjson", "2026-08-02.ndjson"},
			{"ladder", "T-0011.json", "T-0001.json"},
		} {
			raw, err := os.ReadFile(filepath.Join(l.Main, ".bonsai", "local", p.folder, p.src))
			if err != nil {
				t.Fatal(err)
			}
			target := write(t, outside, p.folder+"-"+p.name, string(raw))
			stamp(t, target, daysAgo(70))
			link := filepath.Join(l.Main, ".bonsai", "local", p.folder, p.name)
			if err := os.Symlink(target, link); err != nil {
				t.Fatal(err)
			}
			lstamp(t, link, daysAgo(70))
		}
		links = true
	}
	return links
}

func appendTo(t *testing.T, p, s string) {
	t.Helper()
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(s); err != nil {
		t.Fatal(err)
	}
}

func mkdir(t *testing.T, l workspace.Local, rel string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(l.Main, filepath.FromSlash(rel)), 0o755); err != nil {
		t.Fatal(err)
	}
}

// The links are Linux's only: making one on Windows needs a privilege or developer mode, so the Windows run holds
// every other decoy and the link is held on Linux.
func TestEveryKind(t *testing.T) {
	l := project(t, allRules)
	links := fixture(t, l)
	d := SessionEnd(opts(l))
	if len(d.Errs) > 0 || d.Stopped || len(d.Busy) > 0 {
		t.Fatalf("errors %v, stopped %v, busy %v", d.Errs, d.Stopped, d.Busy)
	}
	wantLog := []string{"S-" + sid("8") + ".ndjson", "notes.txt", "old/", "s-" + sid("2") + ".ndjson",
		"s-" + sid("3") + ".ndjson", "s-" + sid("5") + ".ndjson", "s-" + sid("6") + ".ndjson",
		"s-" + sid("8") + ".ndjson.bak", "s-" + sid("9") + ".ndjson/", "w-2026-02-30.ndjson", "w-2026-10-09.ndjson",
		"w-2026-10-10.ndjson"}
	wantAsks := []string{"2026-08-01.ndjson", "2026-08-05.ndjson", "2026-08-07.ndjson.tmp", "2026-08-08.ndjson/",
		"2026-10-08.ndjson", "2026-10-09.ndjson", "2026-13-45.ndjson", "notes.ndjson"}
	wantLadder := []string{"T-0002.json", "T-0003.json", "T-0005.json", "T-0006.json", "T-0007.json", "T-0008.json.bak",
		"T-0010.json/", "ci.json", "notes.json", "t-0009.json"}
	if links {
		wantLog = append(wantLog, "s-"+sid("9")+"x.ndjson")
		wantAsks = append(wantAsks, "2026-08-03.ndjson")
		wantLadder = append(wantLadder, "T-0011.json")
	}
	for folder, want := range map[string][]string{"log": wantLog, "asks": wantAsks, "ladder": wantLadder} {
		if got := names(t, l, folder); !same(got, sorted(want)) {
			t.Errorf("%s holds\n  %s\nwant\n  %s", folder, strings.Join(got, "\n  "), strings.Join(sorted(want), "\n  "))
		}
	}
	wantRecords := sorted([]string{
		".bonsai/local/ladder/T-0001.json=generated.ladder.keep_days=7",
		".bonsai/local/ladder/T-0004.json=generated.ladder.keep_days=7",
		".bonsai/local/asks/2026-08-02.ndjson=generated.asks.keep_days=30",
		".bonsai/local/asks/2026-08-06.ndjson=generated.asks.keep_days=30",
		".bonsai/local/log/s-" + sid("1") + ".ndjson=generated.log.keep_days=30",
		".bonsai/local/log/s-" + sid("4") + ".ndjson=generated.log.keep_days=30",
		".bonsai/local/log/s-" + sid("7") + ".ndjson=generated.log.keep_days=30",
		".bonsai/local/log/s-abc.ndjson=generated.log.keep_days=30",
		".bonsai/local/log/w-2026-08-01.ndjson=generated.log.keep_days=30",
	})
	if got := cleanRecords(t, l); !same(got, wantRecords) {
		t.Errorf("clean records\n  %s\nwant\n  %s", strings.Join(got, "\n  "), strings.Join(wantRecords, "\n  "))
	}
	if len(d.Cleaned) != len(wantRecords) {
		t.Errorf("Done names %d cleaned, want %d", len(d.Cleaned), len(wantRecords))
	}
	// The asks that stay read as they did: A open, D answered (its answer kept with its filing), F open.
	book, err := asks.Read(l.Main, "")
	if err != nil {
		t.Fatal(err)
	}
	for key, state := range map[string]string{"agent:a": "open", "agent:d": "answered", "agent:f": "open"} {
		if e := book.Get(key); e == nil || e.State != state {
			t.Errorf("%s reads %v, want %s", key, e, state)
		}
	}
	// The links' targets are as they were.
	if links {
		entries, _ := os.ReadDir(filepath.Join(filepath.Dir(l.Main), "outside"))
		if len(entries) != 3 {
			t.Errorf("the links' targets: %d left of 3", len(entries))
		}
	}
	// A second run removes nothing and writes no record.
	before := cleanRecords(t, l)
	d = SessionEnd(opts(l))
	if len(d.Cleaned) != 0 || len(d.Errs) != 0 {
		t.Errorf("a second run cleaned %v, errors %v", d.Cleaned, d.Errs)
	}
	if got := cleanRecords(t, l); !same(got, before) {
		t.Errorf("a second run wrote records: %v", got)
	}
}

func sorted(s []string) []string {
	out := append([]string{}, s...)
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// keep_newest with protected files among the newest: a protected file still counts among them, so the newest two
// here are a protected one and one other, and the third goes even though only one unprotected file is newer.
func TestKeepNewestCountsTheProtected(t *testing.T) {
	l := project(t, "  log:\n    keep_days: null\n    keep_newest: 2\n")
	logFile(t, l, "s-"+sid("1")+".ndjson", session(sid("1"), "none", daysAgo(1.2), daysAgo(1))...) // newest, no row: kept
	logFile(t, l, "s-"+sid("2")+".ndjson", session(sid("2"), "none", daysAgo(2.2), daysAgo(2))...) // second
	logFile(t, l, "s-"+sid("3")+".ndjson", session(sid("3"), "none", daysAgo(3.2), daysAgo(3))...) // third: goes
	logFile(t, l, "s-"+sid("4")+".ndjson", session(sid("4"), "none", daysAgo(4.2), daysAgo(4))...) // fourth, no row: kept
	sessionsTable(t, l, row(sid("2"), "none", daysAgo(2.2), daysAgo(2)), row(sid("3"), "none", daysAgo(3.2), daysAgo(3)))
	d := Files(opts(l), Log)
	if len(d.Errs) > 0 {
		t.Fatal(d.Errs)
	}
	want := []string{"s-" + sid("1") + ".ndjson", "s-" + sid("2") + ".ndjson", "s-" + sid("4") + ".ndjson", "w-2026-10-10.ndjson"}
	if got := names(t, l, "log"); !same(got, want) {
		t.Errorf("log holds %v, want %v", got, want)
	}
	if got := cleanRecords(t, l); !same(got, []string{".bonsai/local/log/s-" + sid("3") + ".ndjson=generated.log.keep_newest=2"}) {
		t.Errorf("records %v", got)
	}
}

// Either rule cleans: keep_newest takes what keep_days leaves, and keep_days names the reason when both would.
func TestEitherRuleCleans(t *testing.T) {
	l := project(t, "  ladder:\n    keep_days: 10\n    keep_newest: 1\n")
	taskFile(t, l, "T-0001", "done")
	taskFile(t, l, "T-0002", "done")
	taskFile(t, l, "T-0003", "done")
	ladderFile(t, l, "T-0001.json", ptr("T-0001"), daysAgo(1))
	ladderFile(t, l, "T-0002.json", ptr("T-0002"), daysAgo(2))
	ladderFile(t, l, "T-0003.json", ptr("T-0003"), daysAgo(20))
	if d := Files(opts(l), Ladder); len(d.Errs) > 0 {
		t.Fatal(d.Errs)
	}
	want := sorted([]string{".bonsai/local/ladder/T-0002.json=generated.ladder.keep_newest=1",
		".bonsai/local/ladder/T-0003.json=generated.ladder.keep_days=10"})
	if got := cleanRecords(t, l); !same(got, want) {
		t.Errorf("records %v, want %v", got, want)
	}
	if got := names(t, l, "ladder"); !same(got, []string{"T-0001.json"}) {
		t.Errorf("ladder holds %v", got)
	}
}

// null keeps: every kind's rule null, or generated: null, cleans nothing, however old.
func TestNullKeeps(t *testing.T) {
	for name, gen := range map[string]string{
		"every rule null": strings.NewReplacer("30", "null", "7", "null").Replace(allRules),
		"each kind null":  "  log: null\n  asks: null\n  ladder: null\n  run: null\n  sessions: null\n",
	} {
		t.Run(name, func(t *testing.T) {
			l := project(t, gen)
			fixture(t, l)
			d := SessionEnd(opts(l))
			if len(d.Cleaned) != 0 || len(d.Errs) != 0 {
				t.Errorf("cleaned %v, errors %v", d.Cleaned, d.Errs)
			}
		})
	}
	t.Run("generated null", func(t *testing.T) {
		l := project(t, "")
		write(t, l.Main, "bonsai.yaml", "format: bonsai.workspace/1\nid: "+testID+"\nname: demo\ndocuments:\n  task: work/tasks\ngenerated: null\n")
		fixture(t, l)
		if d := SessionEnd(opts(l)); len(d.Cleaned) != 0 || len(d.Errs) != 0 {
			t.Errorf("cleaned %v, errors %v", d.Cleaned, d.Errs)
		}
	})
}

// generated: left out takes the defaults: the log after 30 days, ladder results after 7, asks kept.
func TestDefaultsWhenLeftOut(t *testing.T) {
	l := project(t, "")
	fixture(t, l)
	d := SessionEnd(opts(l))
	if len(d.Errs) > 0 {
		t.Fatal(d.Errs)
	}
	var kinds []string
	for _, it := range d.Cleaned {
		kinds = append(kinds, it.Kind)
	}
	if got := strings.Join(kinds, " "); got != "ladder ladder log log log log log" {
		t.Errorf("cleaned kinds %q", got)
	}
}

// A rule that does not read cleans nothing of its kind, and the others go on.
func TestABadRuleCleansNothingOfItsKind(t *testing.T) {
	l := project(t, "  log:\n    keep_days: soon\n  ladder:\n    keep_days: 7\n")
	fixture(t, l)
	d := SessionEnd(opts(l))
	if len(d.Errs) != 1 || !strings.Contains(d.Errs[0].Error(), "generated.log.keep_days") {
		t.Errorf("errors %v", d.Errs)
	}
	for _, it := range d.Cleaned {
		if it.Kind != Ladder {
			t.Errorf("cleaned %v", it)
		}
	}
	if len(d.Cleaned) != 2 {
		t.Errorf("cleaned %d ladder results, want 2", len(d.Cleaned))
	}
}

// An open span is kept whatever the rule: keep_days 0 cleans every ended file whose rows are in, at once, but not
// a session still going.
func TestOpenSpanIsKept(t *testing.T) {
	l := project(t, "  log:\n    keep_days: 0\n")
	logFile(t, l, "s-"+sid("1")+".ndjson", // open: kept
		rec{event: "session_start", at: now.Add(-time.Hour), session: sid("1"), source: "startup"},
		rec{event: "subagent_start", at: now.Add(-30 * time.Minute), session: sid("1"), subagent: "b0000001-x"},
		rec{event: "subagent_stop", at: now.Add(-20 * time.Minute), session: sid("1"), subagent: "b0000001-x"})
	logFile(t, l, "s-"+sid("2")+".ndjson", session(sid("2"), "none", now.Add(-time.Hour), now.Add(-time.Minute))...)
	sessionsTable(t, l, row(sid("2"), "none", now.Add(-time.Hour), now.Add(-time.Minute)))
	if d := Files(opts(l), Log); len(d.Errs) > 0 {
		t.Fatal(d.Errs)
	}
	if got := names(t, l, "log"); !same(got, []string{"s-" + sid("1") + ".ndjson", "w-2026-10-10.ndjson"}) {
		t.Errorf("log holds %v", got)
	}
}

// No sessions table (or one that does not read): every file holding an ended span waits for check --write.
func TestNoTableKeepsEndedSpans(t *testing.T) {
	for name, table := range map[string]string{"missing": "", "broken": "---\nformat: bonsai.sessions/1\n---\nnot a table\n"} {
		t.Run(name, func(t *testing.T) {
			l := project(t, "  log:\n    keep_days: 30\n")
			logFile(t, l, "s-"+sid("1")+".ndjson", session(sid("1"), "none", daysAgo(41), daysAgo(40))...)
			logFile(t, l, "s-"+sid("2")+".ndjson", rec{event: "guard", at: daysAgo(40), session: sid("2")})
			if table != "" {
				write(t, l.Main, ".bonsai/sessions.md", table)
			}
			if d := Files(opts(l), Log); len(d.Errs) > 0 {
				t.Fatal(d.Errs)
			}
			if got := names(t, l, "log"); !same(got, []string{"s-" + sid("1") + ".ndjson", "w-2026-10-10.ndjson"}) {
				t.Errorf("log holds %v", got)
			}
		})
	}
}

// A folder of a kind that is a link (or .bonsai/local itself) is never followed: nothing in it is cleaned. A link
// needs a privilege on Windows, so this runs on Linux and macOS.
func TestLinkedFolderIsNotFollowed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("making a link on Windows needs a privilege or developer mode; the Linux run holds this")
	}
	l := project(t, allRules)
	outside := filepath.Join(filepath.Dir(l.Main), "elsewhere")
	taskFile(t, l, "T-0001", "done")
	elsewhere := projectAt(t, outside)
	ladderFile(t, elsewhere, "T-0001.json", ptr("T-0001"), daysAgo(20))
	ladder := filepath.Join(l.Main, ".bonsai", "local", "ladder")
	if err := os.Remove(ladder); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, ".bonsai", "local", "ladder"), ladder); err != nil {
		t.Fatal(err)
	}
	if d := Files(opts(l), Ladder); len(d.Cleaned) != 0 || len(d.Errs) != 0 {
		t.Errorf("cleaned %v, errors %v", d.Cleaned, d.Errs)
	}
	if _, err := os.Stat(filepath.Join(outside, ".bonsai", "local", "ladder", "T-0001.json")); err != nil {
		t.Errorf("the result behind the link: %v", err)
	}
}

// The ladder runner's own result is never cleaned by the run that wrote it.
func TestKeepNamesTheRunnersResult(t *testing.T) {
	l := project(t, "  ladder:\n    keep_days: 0\n")
	taskFile(t, l, "T-0001", "done")
	taskFile(t, l, "T-0002", "done")
	ladderFile(t, l, "T-0001.json", ptr("T-0001"), now.Add(-time.Minute))
	ladderFile(t, l, "T-0002.json", ptr("T-0002"), now.Add(-time.Hour))
	o := opts(l)
	o.Keep = []string{".bonsai/local/ladder/T-0001.json"}
	if d := LadderResults(o); len(d.Errs) > 0 || len(d.Cleaned) != 1 || d.Cleaned[0].Target != ".bonsai/local/ladder/T-0002.json" {
		t.Errorf("cleaned %v, errors %v", d.Cleaned, d.Errs)
	}
}

// A delete that finds the file gone (another session's cleaner took it) writes nothing.
func TestGoneBeforeItsDeleteWritesNoRecord(t *testing.T) {
	l := project(t, "  ladder:\n    keep_days: 7\n")
	taskFile(t, l, "T-0001", "done")
	ladderFile(t, l, "T-0001.json", ptr("T-0001"), daysAgo(20))
	old := removeFile
	removeFile = func(p string) error {
		_ = os.Remove(p) // the other cleaner, first
		return os.Remove(p)
	}
	t.Cleanup(func() { removeFile = old })
	d := Files(opts(l), Ladder)
	if len(d.Cleaned) != 0 || len(d.Errs) != 0 {
		t.Errorf("cleaned %v, errors %v", d.Cleaned, d.Errs)
	}
	if got := cleanRecords(t, l); len(got) != 0 {
		t.Errorf("records %v", got)
	}
}

// A file that changed after it was judged (a resumed session writing it) is left for a later run.
func TestChangedSinceListedIsLeft(t *testing.T) {
	l := project(t, "  log:\n    keep_days: 30\n")
	p := logFile(t, l, "s-"+sid("1")+".ndjson", rec{event: "guard", at: daysAgo(40), session: sid("1")})
	old := removeFile
	removeFile = func(string) error { t.Error("a changed file was deleted"); return nil }
	t.Cleanup(func() { removeFile = old })
	judged := logKind.protect
	logKind.protect = func(r *run, f *file) (bool, error) {
		protected, err := judged(r, f)
		appendTo(t, p, string(logLine(t, rec{event: "session_start", at: now, session: sid("1"), source: "resume"})))
		return protected, err
	}
	t.Cleanup(func() { logKind.protect = judged })
	if d := Files(opts(l), Log); len(d.Cleaned) != 0 || len(d.Errs) != 0 {
		t.Errorf("cleaned %v, errors %v", d.Cleaned, d.Errs)
	}
}

// A file another process holds open: on Windows a real hold (no sharing), elsewhere the same error stood in, since
// Linux and macOS let a held file be deleted. It is skipped after one try, with no record, and cleaned at a later run.
func TestHeldFileIsCleanedLater(t *testing.T) {
	l := project(t, "  ladder:\n    keep_days: 7\n")
	taskFile(t, l, "T-0001", "done")
	p := ladderFile(t, l, "T-0001.json", ptr("T-0001"), daysAgo(20))
	orig := removeFile
	t.Cleanup(func() { removeFile = orig })
	release := holdFile(t, p)
	tries := 0
	held := removeFile
	removeFile = func(p string) error { tries++; return held(p) }
	d := Files(opts(l), Ladder)
	release()
	if len(d.Busy) != 1 || len(d.Cleaned) != 0 || tries != 1 {
		t.Fatalf("busy %v, cleaned %v, tries %d", d.Busy, d.Cleaned, tries)
	}
	if got := cleanRecords(t, l); len(got) != 0 {
		t.Errorf("records %v", got)
	}
	d = Files(opts(l), Ladder)
	if len(d.Cleaned) != 1 || len(d.Busy) != 0 {
		t.Errorf("the later run: cleaned %v, busy %v", d.Cleaned, d.Busy)
	}
}

// A clean record that cannot be written stops the run: the file it was for is gone, and no other goes unrecorded.
func TestARecordThatFailsStopsTheRun(t *testing.T) {
	l := project(t, "  ladder:\n    keep_days: 7\n")
	for _, id := range []string{"T-0001", "T-0002", "T-0003"} {
		taskFile(t, l, id, "done")
		ladderFile(t, l, id+".json", ptr(id), daysAgo(20))
	}
	old := writeLog
	writeLog = func(string, *format.Log) (bool, error) { return false, errors.New("the disk is full") }
	t.Cleanup(func() { writeLog = old })
	d := Files(opts(l), Ladder)
	if len(d.Errs) != 1 || len(d.Cleaned) != 0 {
		t.Errorf("errors %v, cleaned %v", d.Errs, d.Cleaned)
	}
	if got := names(t, l, "ladder"); len(got) != 2 {
		t.Errorf("ladder holds %v, want two left", got)
	}
}
