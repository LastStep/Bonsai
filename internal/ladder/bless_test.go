package ladder

// The ratchet's judgment, the rise step and Bless through the runner (plan-5 5.4.4's proof; kinds.go's afterRung and
// afterResult): a count below its floor red, a count unread red, equal green, under --ci the accepted skips named; one
// Bless ask filed by a green, clean climb in the main checkout on its base branch at HEAD, and none with each condition
// of the offer false in turn; the ask and its ask log record held to their schemas; a second filing replacing the
// first's numbers; the ask resolved when HEAD moved, the climb went red or dirty, or no rise is left; the climb's bless
// listing the records written; the rise step's new_tests carried into the result, the ask's why and its data, and the
// step called only on a local, green, clean climb. Every project is a real git repository in a temporary folder.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/ratchet"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// countRung is rung 1 of a Bless fixture: it prints count.txt, whose "passed N" its capture reads as the ratchet found.
func countRung(command string) rung {
	return rung{n: 1, name: "count", command: command, required: true, ratchet: "found", capture: "\n      found: \"passed ([0-9]+)\""}
}

// blessProject is a fixture whose floor is rung 1 (countRung, the ratchet found at its floor given in ratchets), with
// count.txt reading "passed 5", the extra rungs given, and three tasks: T-0901 running, T-0902, and T-0903 also
// climbing rung 2.
func blessProject(t *testing.T, ratchets string, rungs ...rung) *fx {
	t.Helper()
	if len(rungs) == 0 {
		rungs = []rung{countRung(show("count.txt"))}
	}
	return project(t, yamlOf("[1]", ratchets, "[]", rungs...), map[string]string{
		"count.txt":              "passed 5\n",
		"work/tasks/T-0901-a.md": taskText("T-0901", "running", ""),
		"work/tasks/T-0902-b.md": taskText("T-0902", "todo", ""),
		"work/tasks/T-0903-c.md": taskText("T-0903", "todo", "[2]"),
	})
}

// blessOps are the climb's bless records' ops and keys: "file ladder:bless-T-0901".
func blessOps(a *Answer) string {
	var out []string
	for _, r := range a.Doc.Bless {
		out = append(out, r.String("op")+" "+r.String("key"))
	}
	return strings.Join(out, ", ")
}

// askLines are the asks files' records, each held to bonsai.ask/1 as written.
func askLines(t *testing.T, main string) []schema.Object {
	t.Helper()
	var out []schema.Object
	files, _ := filepath.Glob(filepath.Join(asks.Folder(main), "*.ndjson"))
	for _, p := range files {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			out = append(out, fits(t, "ask", []byte(line)))
		}
	}
	return out
}

// logEvents are the log's records of the events given, each held to bonsai.log/1 as written, in file order.
func logEvents(t *testing.T, main string, events ...string) []schema.Object {
	t.Helper()
	var out []schema.Object
	files, _ := filepath.Glob(filepath.Join(main, ".bonsai", "local", "log", "*.ndjson"))
	for _, p := range files {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			doc := fits(t, "log", []byte(line))
			for _, e := range events {
				if doc.String("event") == e {
					out = append(out, doc)
				}
			}
		}
	}
	return out
}

// openBless is the open Bless asks' keys, sorted.
func openBless(t *testing.T, main string) string {
	t.Helper()
	b, err := asks.Read(main, "")
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, e := range b.Entries(false) {
		keys = append(keys, e.Key)
	}
	sort.Strings(keys)
	return strings.Join(keys, " ")
}

// The offer: a green, clean climb in the main checkout on its base branch at HEAD, its ratchet above its floor, files
// one Bless ask with source ladder and contract §9.1's data, held to bonsai.ask/1, with its ask log record after the
// climb's ladder record, in the climbing session's file; the climb's bless lists it; the human lines say so.
func TestBlessOffered(t *testing.T) {
	f := blessProject(t, "\n  found: 3")
	f.env[SessionEnv] = "6d1e2f3a-4b5c-4d6e-8f7a-9b0c1d2e3f4a"
	a := f.climb("", Options{Task: "T-0901"})
	if a.Exit != 0 || blessOps(a) != "file ladder:bless-T-0901" {
		t.Fatalf("exit %d, bless %s\n%s", a.Exit, blessOps(a), f.out.String())
	}
	rec := a.Doc.Bless[0]
	data, _ := rec.Get("data")
	want := `{"task":"T-0901","head":"` + f.head() + `","result_sha":"` + *a.Doc.SHA256 + `","rises":[{"name":"found","was":3,"now":5,` +
		`"new_tests":{"base":"","count":2,"failed_on_base":0,"passed_on_base":0,"not_run":2}}]}` // a capture's count: its tests cannot be named
	if schema.Show(data) != want || rec.String("source") != "ladder" || rec.String("type") != "Bless" || rec.String("task") != "T-0901" ||
		rec.String("title") != "Bless T-0901's ratchets" || rec.String("then_text") != ratchet.Then || rec.String("workspace") != testID ||
		rec.String("session") != f.env[SessionEnv] {
		t.Errorf("the ask: %s", schema.Show(rec))
	}
	if lines := askLines(t, f.root); len(lines) != 1 || schema.Show(lines[0]) != schema.Show(rec) {
		t.Errorf("the asks file: %v", lines)
	}
	logs := logEvents(t, f.root, "ladder", "ask")
	if len(logs) != 2 || logs[0].String("event") != "ladder" || logs[1].String("target") != "ladder:bless-T-0901" ||
		logs[1].String("kind") != "file" || logs[1].String("session") != f.env[SessionEnv] {
		t.Errorf("the log: %v", logs)
	}
	if !strings.Contains(f.out.String(), "bless: offered ladder:bless-T-0901 (found 3 -> 5)") {
		t.Errorf("the human lines:\n%s", f.out.String())
	}
}

// Each condition of the offer false in turn, through the runner: no ask filed, the climb's bless [], and a climb that
// rose says why it was not offered.
func TestBlessNotOffered(t *testing.T) {
	type setup func(t *testing.T, f *fx) (dir string, o Options)
	main := func(task string) setup {
		return func(*testing.T, *fx) (string, Options) { return "", Options{Task: task} }
	}
	for _, c := range []struct {
		name     string
		ratchets string
		rungs    []rung
		before   setup
		says     string // the human lines hold this ("" for no line)
	}{
		{name: "--ci", ratchets: "\n  found: 3", before: func(*testing.T, *fx) (string, Options) { return "", Options{CI: true} }},
		{name: "red", ratchets: "\n  found: 3", rungs: []rung{countRung(show("count.txt")), {n: 2, name: "fails", command: "exit 1"}},
			before: main("T-0903"), says: "bless: not offered: the last climb for T-0903 was not green"},
		{name: "a rise on a red rung", ratchets: "\n  found: 3", rungs: []rung{countRung(show("count.txt") + " && exit 1")}, before: main("T-0901")},
		{name: "dirty", ratchets: "\n  found: 3", before: func(t *testing.T, f *fx) (string, Options) {
			f.write("loose.txt", "not committed\n")
			return "", Options{Task: "T-0901"}
		}, says: "bless: not offered: the last climb for T-0901 was dirty"},
		{name: "not the base branch", ratchets: "\n  found: 3", before: func(t *testing.T, f *fx) (string, Options) {
			wt := filepath.Join(f.tmp, "wt")
			testpack.Git(t, f.root, "worktree", "add", "-q", "-b", "t0901", wt)
			return wt, Options{Task: "T-0901"}
		}, says: `bless: not offered: the last climb for T-0901 was on "t0901", not the base branch main`},
		{name: "not the main checkout", ratchets: "\n  found: 3", before: func(t *testing.T, f *fx) (string, Options) {
			wt := filepath.Join(f.tmp, "wt")
			testpack.Git(t, f.root, "worktree", "add", "-q", "--force", wt, "main")
			return wt, Options{Task: "T-0901"}
		}, says: "bless: not offered: the last climb for T-0901 ran in a worktree, not the main checkout"},
		{name: "no rise", ratchets: "\n  found: 5", before: main("T-0901")},
		{name: "a name not a key of ratchets", ratchets: "\n  other: 1", before: main("T-0901"),
			says: "bless: not offered: the last climb for T-0901 measured no ratchet above its floor"},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := blessProject(t, c.ratchets, c.rungs...)
			dir, o := c.before(t, f)
			a := f.climb(dir, o)
			if a.Error != nil || len(a.Doc.Bless) != 0 || len(askLines(t, f.root)) != 0 {
				t.Fatalf("exit %d, %v, bless %s\n%s", a.Exit, a.Error, blessOps(a), f.out.String())
			}
			if c.says != "" && !strings.Contains(f.out.String(), c.says) || c.says == "" && strings.Contains(f.out.String(), "bless:") {
				t.Errorf("the human lines, want %q:\n%s", c.says, f.out.String())
			}
		})
	}
}

// The ask resolved when its offer goes, by a climb on the base branch whatever its task: HEAD moved (and that climb's
// own offer filed, its bless listing both records), the climb went red, the climb was dirty, no rise left (a floor
// raised); a second filing replacing the first's numbers.
func TestBlessResolved(t *testing.T) {
	offered := func(t *testing.T) *fx {
		f := blessProject(t, "\n  found: 3")
		if a := f.climb("", Options{Task: "T-0901"}); blessOps(a) != "file ladder:bless-T-0901" {
			t.Fatalf("not offered: %s\n%s", blessOps(a), f.out.String())
		}
		f.out.Reset()
		return f
	}
	t.Run("HEAD moved", func(t *testing.T) {
		f := offered(t)
		before := f.head()
		f.commit("HEAD moves")
		a := f.climb("", Options{Task: "T-0902"})
		if blessOps(a) != "resolve ladder:bless-T-0901, file ladder:bless-T-0902" {
			t.Fatalf("bless %s\n%s", blessOps(a), f.out.String())
		}
		if !strings.Contains(f.out.String(), "bless: resolved ladder:bless-T-0901: its offer is gone: the last climb for T-0901 was made at "+
			before[:8]+", and HEAD is "+f.head()[:8]+" (HEAD moved)") {
			t.Errorf("the human lines:\n%s", f.out.String())
		}
		if got := openBless(t, f.root); got != "ladder:bless-T-0902" {
			t.Errorf("open: %s", got)
		}
		if logs := logEvents(t, f.root, "ask"); len(logs) != 3 || logs[1].String("kind") != "resolve" || logs[2].String("kind") != "file" {
			t.Errorf("the ask log records: %v", logs)
		}
		askLines(t, f.root)
	})
	t.Run("the climb went red", func(t *testing.T) {
		f := offered(t)
		f.write("count.txt", "passed 2\n")
		f.commit("fewer")
		a := f.climb("", Options{Task: "T-0901"})
		if a.Exit != ExitRed || blessOps(a) != "resolve ladder:bless-T-0901" ||
			!strings.Contains(f.out.String(), "its offer is gone: the last climb for T-0901 was not green") {
			t.Fatalf("exit %d, bless %s\n%s", a.Exit, blessOps(a), f.out.String())
		}
		if openBless(t, f.root) != "" {
			t.Error("still open")
		}
	})
	t.Run("the climb was dirty", func(t *testing.T) {
		f := offered(t)
		f.write("loose.txt", "not committed\n")
		a := f.climb("", Options{Task: "T-0901"})
		if a.Exit != 0 || blessOps(a) != "resolve ladder:bless-T-0901" ||
			!strings.Contains(f.out.String(), "its offer is gone: the last climb for T-0901 was dirty") {
			t.Fatalf("exit %d, bless %s\n%s", a.Exit, blessOps(a), f.out.String())
		}
	})
	t.Run("no rise left", func(t *testing.T) {
		f := offered(t)
		// A person raises the floor to the count and commits it (the ladder reads bonsai.yaml as committed); the task's
		// next climb, at the new HEAD, measures no rise, and resolves the ask.
		f.write("bonsai.yaml", yamlOf("[1]", "\n  found: 5", "[]", countRung(show("count.txt"))))
		f.commit("the floor raised to the blessed count")
		a := f.climb("", Options{Task: "T-0901"})
		if a.Exit != 0 || blessOps(a) != "resolve ladder:bless-T-0901" ||
			!strings.Contains(f.out.String(), "its offer is gone: the last climb for T-0901 measured no ratchet above its floor") {
			t.Fatalf("exit %d, bless %s\n%s", a.Exit, blessOps(a), f.out.String())
		}
	})
	t.Run("a second filing", func(t *testing.T) {
		f := offered(t)
		f.write("count.txt", "passed 8\n")
		f.commit("more")
		if a := f.climb("", Options{Task: "T-0901"}); blessOps(a) != "file ladder:bless-T-0901" {
			t.Fatalf("bless %s\n%s", blessOps(a), f.out.String())
		}
		b, _ := asks.Read(f.root, "")
		e := b.Get("ladder:bless-T-0901")
		if data, _ := e.Docs[0].Get("data"); e.State != asks.Open || !strings.Contains(schema.Show(data), `"was":3,"now":8`) ||
			!strings.Contains(schema.Show(data), `"head":"`+f.head()+`"`) {
			t.Errorf("the numbers: %s", schema.Show(e.Docs[0]))
		}
		if n := len(askLines(t, f.root)); n != 2 {
			t.Errorf("%d records", n)
		}
	})
}

// The ratchet's judgment through the runner: below its floor the rung is red (exit 1), its reason "found went
// backwards: 3 -> 2", and next names the fix; a count unread is red; equal is green; under --ci the accepted marked
// skip is named; a ver-git rung's ratchet measures nothing.
func TestRatchetJudged(t *testing.T) {
	f := blessProject(t, "\n  found: 3")
	f.write("count.txt", "passed 2\n")
	f.commit("fewer")
	a := f.climb("", Options{Task: "T-0901"})
	l := f.result("T-0901")
	r := l.Rungs[0]
	if a.Exit != ExitRed || r.Status != "red" || r.Ratchet == nil || r.Ratchet.OK || r.Ratchet.Was != "3" || r.Ratchet.Now != "2" ||
		r.Reason == nil || !strings.HasPrefix(*r.Reason, "found went backwards: 3 -> 2") {
		t.Fatalf("exit %d, %+v, %v", a.Exit, r.Ratchet, r.Reason)
	}
	if a.Doc.Next == nil || !strings.HasPrefix(a.Doc.Next.Do, "fix the count rung (rung 1): find why found fell from 3 to 2") ||
		a.Doc.Next.Who != "agent" {
		t.Errorf("next: %+v", a.Doc.Next)
	}
	if !strings.Contains(f.out.String(), "found went backwards: 3 -> 2") {
		t.Errorf("the human lines:\n%s", f.out.String())
	}
	f.write("count.txt", "nothing counted\n")
	f.commit("none")
	a = f.climb("", Options{Task: "T-0901"})
	r = f.result("T-0901").Rungs[0]
	if a.Exit != ExitRed || r.Status != "red" || r.Ratchet != nil || r.Reason == nil ||
		!strings.Contains(*r.Reason, "the ratchet found measured nothing: its capture found matched no line of the output") {
		t.Fatalf("unread: exit %d, %+v, %v", a.Exit, r.Ratchet, r.Reason)
	}
	f.write("count.txt", "passed 3\n")
	f.commit("equal")
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 || f.result("T-0901").Rungs[0].Ratchet == nil || !f.result("T-0901").Rungs[0].Ratchet.OK {
		t.Errorf("equal: exit %d\n%s", a.Exit, f.out.String())
	}

	// Under --ci, an accepted marked skip counts and the red reason says so: 1 passed + 1 accepted = 2, below 3.
	ci := strings.ReplaceAll(goJSON, `{"Action":"run","Package":"example.com/demo/good","Test":"TestPlainSkip"}
{"Action":"output","Package":"example.com/demo/good","Test":"TestPlainSkip","Output":"    a_test.go:9: not on this system\n"}
{"Action":"skip","Package":"example.com/demo/good","Test":"TestPlainSkip","Elapsed":0}
`, "")
	tr := rung{n: 1, name: "tests", command: show("out.json"), required: true, ratchet: "go_tests", tst: "go-test-json"}
	g := project(t, yamlOf("[1]", "\n  go_tests: 3", `["example.com/demo/good.TestMarked"]`, tr), map[string]string{"out.json": ci,
		"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	a = g.climb("", Options{CI: true})
	r = g.result("ci").Rungs[0]
	if a.Exit != ExitRed || r.Ratchet == nil || r.Ratchet.Now != "2" || r.Reason == nil ||
		*r.Reason != "go_tests went backwards: 3 -> 2 (its floor is bonsai.yaml's ratchets.go_tests; a count may only rise); under --ci "+
			"the count holds 1 accepted marked skip" {
		t.Errorf("--ci: exit %d, %+v, %v", a.Exit, r.Ratchet, r.Reason)
	}

	// A ver-git rung counts nothing: its ratchet measures nothing, and the rung is not green.
	h := project(t, yamlOf("[1, 5]", "{}", "[]", countRung(show("count.txt")), rung{n: 5, name: "git", kind: "ver-git", ratchet: "hidden"}),
		map[string]string{"count.txt": "passed 5\n", "work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	h.climb("", Options{Task: "T-0901"})
	r = h.result("T-0901").Rungs[1]
	if r.Status == "green" || r.Reason == nil || !strings.Contains(*r.Reason, "the ratchet hidden measured nothing: a ver-git rung counts nothing") {
		t.Errorf("ver-git: %s %v", r.Status, r.Reason)
	}
}

// The rise step: on a local, green, clean climb, each ratchet above its floor (a key of ratchets or not) gets its
// new_tests from ratchet.NewTests, carried into the result, the Bless ask's why and its data; on a red, dirty or --ci
// climb, or with no rise, it is not called and new_tests stays null. A branch climb, green and clean, gets it too.
func TestRiseStep(t *testing.T) {
	defer func(saved func(ratchet.RiseInput) ratchet.Checked) { newTests = saved }(newTests)
	var calls []ratchet.RiseInput
	newTests = func(in ratchet.RiseInput) ratchet.Checked {
		calls = append(calls, in)
		summary := "2 new tests, run on the base " + in.MergeBase[:12] + ": 2 failed there, 0 prove nothing new, 0 check not run"
		return ratchet.Checked{NewTests: &format.LadderNewTests{Base: in.MergeBase, Count: 2, FailedOnBase: 2}, Summary: summary,
			Lines: []string{summary, "  failed on the base: TestA, TestB"}}
	}
	f := blessProject(t, "\n  found: 3")
	a := f.climb("", Options{Task: "T-0901"})
	r := f.result("T-0901").Rungs[0]
	if len(calls) != 1 || r.Ratchet == nil || r.Ratchet.NewTests == nil || r.Ratchet.NewTests.Count != 2 || r.Ratchet.NewTests.Base != f.head() {
		t.Fatalf("%d calls, %+v", len(calls), r.Ratchet)
	}
	in := calls[0]
	if in.Name != "found" || in.Was != "3" || in.Now != "5" || in.Task != "T-0901" || in.Head != f.head() || in.Base != "main" ||
		in.MergeBase != f.head() || in.Root != f.root || in.Main != f.root || in.Rung.Rung != 1 || in.Tests != nil || in.Timeout <= 0 || in.Stop == nil ||
		len(in.Env) == 0 {
		t.Errorf("the input: %+v", in)
	}
	rec := a.Doc.Bless[0]
	data, _ := rec.Get("data")
	if !strings.Contains(rec.String("why"), "found rose from 3 to 5, 2 new tests, run on the base "+f.head()[:12]+": 2 failed there, 0 prove "+
		"nothing new, 0 check not run.") ||
		!strings.Contains(schema.Show(data), `"new_tests":{"base":"`+f.head()+`","count":2,"failed_on_base":2,"passed_on_base":0,"not_run":0}`) {
		t.Errorf("the ask: %s", schema.Show(rec))
	}
	// The check's lines print after the rungs': a line before it starts, then its summary and its names.
	if out := f.out.String(); !strings.Contains(out, "  new tests, rung 1 count (found 3->5): checking them on the base") ||
		!strings.Contains(out, "s: 2 new tests, run on the base "+f.head()[:12]+": 2 failed there") ||
		!strings.Contains(out, "\n    failed on the base: TestA, TestB\n") ||
		strings.Index(out, "count") > strings.Index(out, "new tests, rung 1") {
		t.Errorf("the human lines:\n%s", out)
	}
	for _, c := range []struct {
		name   string
		before func(f *fx) (string, Options)
		calls  int
	}{
		{"--ci", func(*fx) (string, Options) { return "", Options{CI: true} }, 0},
		{"dirty", func(f *fx) (string, Options) { f.write("loose.txt", "x\n"); return "", Options{Task: "T-0901"} }, 0},
		{"red", func(*fx) (string, Options) { return "", Options{Task: "T-0903"} }, 0},
		{"no rise", func(f *fx) (string, Options) {
			f.write("count.txt", "passed 3\n")
			f.commit("equal")
			return "", Options{Task: "T-0901"}
		}, 0},
		{"a branch", func(f *fx) (string, Options) {
			wt := filepath.Join(f.tmp, "wt")
			testpack.Git(f.t, f.root, "worktree", "add", "-q", "-b", "t0901", wt)
			return wt, Options{Task: "T-0901"}
		}, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			calls = nil
			rungs := []rung{countRung(show("count.txt")), {n: 2, name: "fails", command: "exit 1"}}
			if c.name == "--ci" {
				rungs = rungs[:1] // a --ci climb takes every rung: without the failing one it is green
			}
			g := blessProject(t, "\n  found: 3", rungs...)
			dir, o := c.before(g)
			g.climb(dir, o)
			name := "T-0901"
			if o.CI {
				name = "ci"
			} else if o.Task != "" {
				name = o.Task
			}
			l := g.result(name)
			nt := l.Rungs[0].Ratchet
			if c.name == "--ci" && !l.Green {
				t.Fatalf("the --ci climb is not green:\n%s", g.out.String())
			}
			if len(calls) != c.calls || nt == nil || (nt.NewTests != nil) != (c.calls > 0) {
				t.Errorf("%d calls, want %d; %+v\n%s", len(calls), c.calls, nt, g.out.String())
			}
		})
	}
	// A ratchet that is no key of ratchets rose from 0: the rise step runs for it, though Bless offers it not.
	calls = nil
	k := blessProject(t, "{}")
	if a := k.climb("", Options{Task: "T-0901"}); len(calls) != 1 || len(a.Doc.Bless) != 0 {
		t.Errorf("%d calls, bless %s", len(calls), blessOps(a))
	}
}

// A climb whose ladder record cannot be written (write-failed, exit 3) runs no Bless step: its result has no
// fingerprint, so nothing is filed from it, an open ask its HEAD move would resolve stays open, and the human lines
// say so.
func TestBlessNeedsTheRecord(t *testing.T) {
	f := blessProject(t, "\n  found: 3")
	defer func(saved func(string, *format.Log) (bool, error)) { writeLog = saved }(writeLog)
	writeLog = func(string, *format.Log) (bool, error) { return false, os.ErrPermission }
	a := f.climb("", Options{Task: "T-0901"})
	if a.Exit != ExitRuntime || a.Error == nil || a.Error.Code != "write-failed" || len(a.Doc.Bless) != 0 || len(askLines(t, f.root)) != 0 {
		t.Fatalf("exit %d, %v, bless %s\n%s", a.Exit, a.Error, blessOps(a), f.out.String())
	}
	says := "bless: not run: the result's ladder record is not in the log, so this climb files and resolves no Bless ask"
	if !strings.Contains(f.out.String(), says) || strings.Contains(f.out.String(), "bless: offered") {
		t.Errorf("the human lines:\n%s", f.out.String())
	}
	// An offer filed by a climb with its record; HEAD moves; a climb with no record resolves nothing.
	writeLog = record.WriteLog
	if a := f.climb("", Options{Task: "T-0901"}); blessOps(a) != "file ladder:bless-T-0901" {
		t.Fatalf("not offered: %s\n%s", blessOps(a), f.out.String())
	}
	f.commit("HEAD moves")
	writeLog = func(string, *format.Log) (bool, error) { return false, os.ErrPermission }
	f.out.Reset()
	if a := f.climb("", Options{Task: "T-0902"}); a.Exit != ExitRuntime || len(a.Doc.Bless) != 0 || openBless(t, f.root) != "ladder:bless-T-0901" {
		t.Errorf("exit %d, bless %s, open %q\n%s", a.Exit, blessOps(a), openBless(t, f.root), f.out.String())
	}
	// Under --ci there is no Bless step to skip, and nothing is said.
	f.out.Reset()
	if a := f.climb("", Options{CI: true}); a.Exit != ExitRuntime || strings.Contains(f.out.String(), "bless:") {
		t.Errorf("--ci: exit %d\n%s", a.Exit, f.out.String())
	}
}

// A floor that is not a number is never read as 0, which would let a typo lower it: the rung is red, its reason and
// its next naming the fix in bonsai.yaml, and no ratchet written; a floor that is a number, written as one or as
// text, is judged. bonsai.yaml's own read refuses such a floor before anything runs (the second half), so the runner's
// judgment guards a Config built otherwise.
func TestFloorNotANumber(t *testing.T) {
	cmd := "cat count.txt"
	r := format.Rung{Rung: 1, Name: strPtr("count"), Kind: KindCommand, Command: &cmd, Required: true, Ratchet: "found"}
	for _, c := range []struct {
		floor any
		red   bool
	}{
		{"many", true}, {true, true}, {nil, true}, {[]any{json.Number("3")}, true}, {json.Number("x"), true},
		{json.Number("3"), false}, {"3", false}, {int64(3), false}, {3.5, false},
	} {
		climb := &Climb{Config: &workspace.Config{Full: &format.Workspace{Ratchets: schema.Object{{Key: "found", Value: c.floor}}}}}
		e := &format.LadderRung{Rung: 1, Name: "count", Kind: KindCommand, Status: statusGreen}
		floor, _ := floorOf(climb.Config.Full.Ratchets, "found")
		s := &Step{Rung: r, Entry: e, Count: &Count{Name: "found", Floor: floor, Now: "5", From: "capture"}}
		judgeRatchet(climb, s)
		why := ""
		if e.Reason != nil {
			why = *e.Reason
		}
		switch {
		case c.red && (e.Status != statusRed || e.Ratchet != nil || s.who != "person" ||
			!strings.HasPrefix(why, "the ratchet found cannot be judged: its floor in bonsai.yaml's ratchets is "+schema.Show(c.floor)+", which is not a number") ||
			!strings.Contains(why, "the floor in bonsai.yaml's ratchets must be a number") ||
			!strings.HasPrefix(s.next, "a person writes ratchets.found in bonsai.yaml as a plain number")):
			t.Errorf("%v: %s, %+v, %q, next %q (%s)", c.floor, e.Status, e.Ratchet, why, s.next, s.who)
		case !c.red && (e.Status != statusGreen || e.Ratchet == nil || !e.Ratchet.OK || e.Ratchet.Was == "0"):
			t.Errorf("%v: %s, %+v, %q", c.floor, e.Status, e.Ratchet, why)
		}
	}
	// Through bonsai.yaml: a floor written as text is refused before a rung runs.
	f := blessProject(t, "\n  found: many")
	a := f.climb("", Options{Task: "T-0901"})
	if a.Exit != ExitInput || a.Error == nil || !strings.Contains(a.Error.What, "ratchets.found: is string, want number") ||
		a.Doc.Result != nil {
		t.Errorf("exit %d, %+v", a.Exit, a.Error)
	}
}
