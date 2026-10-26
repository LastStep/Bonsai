package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/testpack"
)

// A stamp ahead of the UTC now (BUG-123): a run report's finished an hour ahead is a future-stamp finding, as is a
// task's day after the UTC today; 5 minutes ahead, the past, and today are not; a report in format 0 is judged too.
func TestFutureStamp(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "stamps")
	e.link(t, root, e.pack.A)
	at := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	was := now
	now = func() time.Time { return at.In(time.FixedZone("IST", 19800)) }
	t.Cleanup(func() { now = was })
	run := func(started, finished string) string {
		return "---\nformat: bonsai.run/1\nid: R-2026-10-26-T-0901\ntask: T-0901\nrole: builder\nmodel: a model\n" +
			"started: " + started + "\nfinished: " + finished + "\noutcome: merged\ncommits: []\nlabels: {}\n---\n"
	}
	task := func(finished string) string {
		return "---\nformat: bonsai.task/1\nid: T-0901\ntitle: \"A task\"\nstatus: done\nlane: null\ndone_when: [\"it is done\"]\n" +
			"depends_on: []\nblocked_by: null\ncreated: 2026-10-25\nstarted: 2026-10-25\nfinished: " + finished + "\nlabels: {}\n---\n"
	}
	const report, taskFile = "work/runs/R-2026-10-26-T-0901.md", "work/tasks/T-0901-a-task.md"
	for _, c := range []struct {
		name, run, task string
		want            []string // the files with a future-stamp finding
	}{
		{"an hour ahead", run(`"2026-10-26 11:00"`, `"2026-10-26 13:00"`), task("2026-10-26"), []string{report}},
		{"5 minutes ahead", run(`"2026-10-26 11:00"`, `"2026-10-26 12:05"`), task("2026-10-26"), nil},
		{"the past", run(`"2026-10-25 09:00"`, `"2026-10-25 10:00"`), task("2026-10-25"), nil},
		{"a day ahead", run(`"2026-10-26 11:00"`, `null`), task("2026-10-27"), []string{taskFile}},
		{"started ahead, a task's tomorrow", run(`"2026-10-26 18:30"`, `null`), task("2026-10-27"), []string{taskFile, report}},
	} {
		writeFile(t, root, report, c.run)
		writeFile(t, root, taskFile, c.task)
		r, err := checkLocal(t, root, e.home)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		var got []string
		for _, f := range r.Findings {
			switch f.Code {
			case "future-stamp":
				got = append(got, f.File)
				if !strings.Contains(f.Message, "UTC") || f.Next == "" {
					t.Errorf("%s: %+v", c.name, f)
				}
			case "document":
				t.Errorf("%s: the fixture does not read: %+v", c.name, f)
			}
		}
		if strings.Join(got, " ") != strings.Join(c.want, " ") {
			t.Errorf("%s: future-stamp on %q, want %q", c.name, got, c.want)
		}
	}
}

// A pack kind's stamp field (the declaring pack's plan, stamp approved) is judged the same: a day after the UTC
// today is a finding, today is not.
func TestFutureStampPackKind(t *testing.T) {
	e := setup(t)
	src, shas := testpack.DeclaringPack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "plans")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "plans", Source: src, Ref: shas[0]}, AllowExec: true})
	at := time.Date(2026, 10, 26, 23, 30, 0, 0, time.UTC)
	was := now
	now = func() time.Time { return at }
	t.Cleanup(func() { now = was })
	const plan = "work/plans/P-T-0901.md"
	for day, want := range map[string]bool{"2026-10-27": true, "2026-10-26": false} {
		writeFile(t, root, plan, "---\nid: P-T-0901\ntask: T-0901\nstatus: approved\napproved: "+day+"\n---\n\n# A plan\n")
		r, err := checkLocal(t, root, e.home)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, f := range r.Findings {
			found = found || (f.Code == "future-stamp" && f.File == plan)
		}
		if found != want {
			t.Errorf("approved %s: future-stamp %v, want %v: %+v", day, found, want, r.Findings)
		}
	}
}
