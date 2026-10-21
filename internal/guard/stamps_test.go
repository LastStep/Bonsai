package guard

// Fresh stamps checked (design/contract.md section 23; P-T-5795 step 6, amendment 2; stamps.go): the proof table's
// cases, each judged by the guard at a set clock, in both status_writes modes where the mode could matter.

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// stampYAML is grantYAML with a memory folder.
var stampYAML = strings.Replace(grantYAML, "  run: work/runs\n", "  run: work/runs\n  memory: work/memory\n", 1)

// stampNow is the guard's clock in these tests: 21 Oct 2026, 09:16 UTC (14:46 in India, UTC+5:30).
var stampNow = time.Date(2026, 10, 21, 9, 16, 30, 0, time.UTC)

const (
	stampTask  = "work/tasks/T-0001-the-task.md"
	stampRun   = "work/runs/R-2026-10-21-T-0001.md"
	stampState = ".bonsai/STATE.md"
	stampNote  = "work/memory/M-a-note.md"
)

// runText is a run report started at started and finished at finished ("" for empty).
func runText(started, finished string) string {
	return "---\nformat: bonsai.run/1\nid: R-2026-10-21-T-0001\ntask: T-0001\nrole: builder\nmodel: a-model\nstarted: " +
		started + "\nfinished: " + finished + "\noutcome: running\ncommits: []\nlabels: {}\n---\n\n# Run\n"
}

func stateText(updated string) string {
	return "---\nformat: bonsai.state/1\nupdated: " + updated + "\nupdated_by: an agent\nlabels: {}\n---\n\n# Where it stands\n"
}

func noteText(updated string) string {
	return "---\nformat: bonsai.memory/1\nid: M-a-note\ntitle: A note\nkind: project\nupdated: " + updated +
		"\nsource: a test\nlabels: {}\n---\n\nThe note.\n"
}

// stampProject is a verified project holding a task (created 2026-10-11, started empty), a run report started at
// 08:00 with finished empty, STATE and a memory note updated 2026-10-01; in command mode when command.
func stampProject(t *testing.T, command bool) string {
	t.Helper()
	dir := grantProject(t, stampYAML, map[string]string{
		stampTask:  taskText("T-0001", "todo", "light"),
		stampRun:   runText("2026-10-21 08:00", ""),
		stampState: stateText("2026-10-01"),
		stampNote:  noteText("2026-10-01"),
	})
	if command {
		commandMode(t, dir)
	}
	return dir
}

// stampAnswer runs the guard at now and checks the rule it records (time-not-now refused, else allowed); it returns
// the refusal's text.
func stampAnswer(t *testing.T, dir, payload string, now time.Time, want string) string {
	t.Helper()
	code, s := runAt(t, dir, payload, now, false)
	wantCode := 0
	if !ruleAllows(t, want) {
		wantCode = 2
	}
	if code != wantCode {
		t.Fatalf("exit %d, want %d (%s): %s\n%s", code, wantCode, want, s, payload)
	}
	if got := lastRule(t, dir, payload); got != want {
		t.Fatalf("rule %q, want %q: %s", got, want, s)
	}
	if code == 2 {
		asciiOnly(t, s)
	}
	return s
}

// The proof table's cases, in both modes: a refused fresh stamp is refused in each, and a stamp that passes passes in
// each (in command mode the task's other tripwires do not trip on these edits).
func TestFreshStamps(t *testing.T) {
	cases := []struct {
		name    string
		payload func(dir string) string
		now     time.Time
		refused bool
		says    []string
	}{
		// Times: a run report's started and finished.
		{"a run's finished from bonsai now", func(d string) string {
			return editIn(d, stampRun, "finished: \n", "finished: 2026-10-21 09:16\n")
		}, stampNow, false, nil},
		{"a run's finished 9 minutes ahead (the window's edge)", func(d string) string {
			return editIn(d, stampRun, "finished: \n", "finished: 2026-10-21 09:25\n")
		}, stampNow, false, nil},
		{"a stale finished, 30 minutes old", func(d string) string {
			return editIn(d, stampRun, "finished: \n", "finished: 2026-10-21 08:46\n")
		}, stampNow, true, []string{`finished "2026-10-21 08:46"`, "30 min from Bonsai's clock", "Bonsai's UTC now: 2026-10-21 09:16",
			"next: run: bonsai now, and write the UTC time it prints"}},
		{"a guess in local time, 5 h 30 min off (IST typed as UTC)", func(d string) string {
			return editIn(d, stampRun, "finished: \n", "finished: 2026-10-21 14:46\n")
		}, stampNow, true, []string{"5 h 29 min from Bonsai's clock"}},
		{"a time with a zone name", func(d string) string {
			return editIn(d, stampRun, "finished: \n", "finished: 2026-10-21 09:16 IST\n")
		}, stampNow, true, []string{"not a recorded time (YYYY-MM-DD HH:MM, in UTC)"}},
		{"a quoted time with a zone name", func(d string) string {
			return editIn(d, stampRun, "finished: \n", "finished: \"2026-10-21 09:16 +05:30\"\n")
		}, stampNow, true, []string{"not a recorded time"}},
		{"a new run report started from bonsai now", func(d string) string {
			return writeOf(d, "work/runs/R-2026-10-21-T-0001-b.md", runText("2026-10-21 09:15", "null"))
		}, stampNow, false, nil},
		{"a new run report started an hour ago", func(d string) string {
			return writeOf(d, "work/runs/R-2026-10-21-T-0001-b.md", runText("2026-10-21 08:16", "null"))
		}, stampNow, true, []string{`started "2026-10-21 08:16"`, "1 h 0 min"}},
		{"a time across UTC midnight, 3 minutes before it", func(d string) string {
			return editIn(d, stampRun, "finished: \n", "finished: 2026-10-21 23:58\n")
		}, time.Date(2026, 10, 22, 0, 1, 0, 0, time.UTC), false, nil},
		// Corrections, null and values left as they were.
		{"a correction of a time already there", func(d string) string {
			return editIn(d, stampRun, "started: 2026-10-21 08:00", "started: 2026-10-21 07:58")
		}, stampNow, false, nil},
		{"null written in an empty field", func(d string) string {
			return editIn(d, stampRun, "finished: \n", "finished: null\n")
		}, stampNow, false, nil},
		{"a value taken back to null", func(d string) string {
			return editIn(d, stampTask, "created: 2026-10-11", "created: null")
		}, stampNow, false, nil},
		{"an untouched legacy value under another edit", func(d string) string {
			return editIn(d, stampTask, "title: A made-up task", "title: A clearer title")
		}, stampNow, false, nil},
		{"a field outside the list", func(d string) string {
			return editIn(d, stampTask, "blocked_by:\n", "blocked_by: 2026-01-01\n")
		}, stampNow, false, nil},
		// Days: a task's started and created.
		{"a task's started from bonsai now --day", func(d string) string {
			return editIn(d, stampTask, "started:\n", "started: 2026-10-21\n")
		}, stampNow, false, nil},
		{"a day 2 days back (the window's edge)", func(d string) string {
			return editIn(d, stampTask, "started:\n", "started: 2026-10-19\n")
		}, stampNow, false, nil},
		{"a day 3 days back", func(d string) string {
			return editIn(d, stampTask, "started:\n", "started: 2026-10-18\n")
		}, stampNow, true, []string{`started "2026-10-18"`, "more than 2 days before the UTC today", "Bonsai's UTC now: 2026-10-21",
			"next: run: bonsai now --day, and write the UTC day it prints"}},
		{"a future day", func(d string) string {
			return editIn(d, stampTask, "started:\n", "started: 2026-10-22\n")
		}, stampNow, true, []string{"after the UTC today"}},
		{"a local day ahead of UTC just after local midnight", func(d string) string {
			return editIn(d, stampTask, "started:\n", "started: 2026-10-22\n")
		}, time.Date(2026, 10, 21, 18, 40, 0, 0, time.UTC), true, []string{"after the UTC today"}},
		{"yesterday's day just after UTC midnight", func(d string) string {
			return editIn(d, stampTask, "started:\n", "started: 2026-10-21\n")
		}, time.Date(2026, 10, 22, 0, 2, 0, 0, time.UTC), false, nil},
		{"a day that is no day", func(d string) string {
			return editIn(d, stampTask, "started:\n", "started: 21 Oct\n")
		}, stampNow, true, []string{"not a day (YYYY-MM-DD, the UTC day)"}},
		{"a value that opens a structure is the format's, not a stamp", func(d string) string {
			return editIn(d, stampTask, "started:\n", "started: [\n")
		}, stampNow, false, nil},
		{"a new task created today", func(d string) string {
			return writeOf(d, "work/tasks/T-0002-new.md", strings.Replace(taskText("T-0002", "todo", "light"),
				"created: 2026-10-11", "created: 2026-10-21", 1))
		}, stampNow, false, nil},
		{"a new task created 10 days back", func(d string) string {
			return writeOf(d, "work/tasks/T-0002-new.md", taskText("T-0002", "todo", "light"))
		}, stampNow, true, []string{`created "2026-10-11"`}},
		// STATE's and a memory note's updated: every change is fresh.
		{"STATE's updated changed to today", func(d string) string {
			return editIn(d, stampState, "updated: 2026-10-01", "updated: 2026-10-21")
		}, stampNow, false, nil},
		{"STATE's updated changed to a stale day", func(d string) string {
			return editIn(d, stampState, "updated: 2026-10-01", "updated: 2026-10-02")
		}, stampNow, true, []string{`state's updated "2026-10-02"`}},
		{"STATE rewritten, its old updated left", func(d string) string {
			return writeOf(d, stampState, strings.Replace(stateText("2026-10-01"), "# Where it stands", "# Where it stands now", 1))
		}, stampNow, false, nil},
		{"a memory note's updated changed to a stale day", func(d string) string {
			return editIn(d, stampNote, "updated: 2026-10-01", "updated: 2026-10-05")
		}, stampNow, true, []string{`memory's updated "2026-10-05"`}},
		{"a memory note's updated changed to today", func(d string) string {
			return editIn(d, stampNote, "updated: 2026-10-01", "updated: 2026-10-21")
		}, stampNow, false, nil},
		{"the memory index is no note", func(d string) string {
			return writeOf(d, "work/memory/INDEX.md", "---\nupdated: 2020-01-01\n---\n")
		}, stampNow, false, nil},
	}
	// A stamp in a wrong form fails its kind's format as well (each schema's pattern): in command mode bad-format
	// answers for it (tripwire.go), in agents mode time-not-now.
	wrongForm := map[string]bool{"a time with a zone name": true, "a quoted time with a zone name": true,
		"a day that is no day": true}
	// A file that does not read at all (an open [) is its format's alone: agents mode allows it as before, command
	// mode refuses it as bad-format, so the guard and rung 0 still agree (internal/rung0's agree_test.go).
	const brokenFile = "a value that opens a structure is the format's, not a stamp"
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, mode := range []string{"agents", "command"} {
				dir := stampProject(t, mode == "command")
				want := RuleNotProtected
				if c.refused {
					want = RuleTimeNotNow
				}
				if mode == "command" && (c.refused && wrongForm[c.name] || c.name == brokenFile) {
					want = RuleBadFormat // the file fails its format too: command mode's plainer refusal answers
				}
				s := stampAnswer(t, dir, c.payload(dir), c.now, want)
				if want == RuleBadFormat {
					continue
				}
				for _, w := range c.says {
					if !strings.Contains(s, w) {
						t.Errorf("%s mode: the refusal does not say %q: %s", mode, w, s)
					}
				}
			}
		})
	}
}

// A pack kind's stamp field (a plan's approved, as the workflow pack declares it) is a day judged as a task's; the
// kind's other fields are not judged.
func TestFreshStampOfAPacksKind(t *testing.T) {
	yaml := strings.Replace(grantYAML, "packs: []", "packs:\n  - id: base\n    source: \"https://example.com/base.git\"\n    ref: v1.0.0", 1)
	declares := `{"documents":[{"kind":"plan","path":"work/plans","file":null,"id":"^P-","statuses":["draft","approved"],` +
		`"person":[["draft","approved"]],"agent":[],"stamp":{"approved":"approved"},"task_field":"task"}]}`
	plan := "---\nid: P-0001\ntask: T-0001\nstatus: draft\ncreated: 2026-01-01\napproved:\nlabels: {}\n---\n\n# Plan\n"
	for _, mode := range []string{"agents", "command"} {
		dir := grantProject(t, yaml, map[string]string{
			".bonsai/lock.json":             lockWith(declares, `{}`),
			"work/tasks/T-0001-the-task.md": taskText("T-0001", "todo", "light"),
			"work/plans/P-0001.md":          plan,
		})
		if mode == "command" {
			commandMode(t, dir)
		}
		s := stampAnswer(t, dir, editIn(dir, "work/plans/P-0001.md", "approved:\n", "approved: 2026-10-01\n"), stampNow, RuleTimeNotNow)
		if !strings.Contains(s, `the plan's approved "2026-10-01"`) {
			t.Errorf("%s mode: the refusal: %s", mode, s)
		}
		stampAnswer(t, dir, editIn(dir, "work/plans/P-0001.md", "approved:\n", "approved: 2026-10-21\n"), stampNow, RuleNotProtected)
		stampAnswer(t, dir, editIn(dir, "work/plans/P-0001.md", "# Plan\n", "# Plan\n\nA note.\n"), stampNow, RuleNotProtected)
	}
}

// Command mode's tripwires still ignore a memory note (the orchestrator's Q3 ruling): a label set from outside on a
// note, refused on a task in command mode, passes on a note; its stamp alone is judged.
func TestCommandModeIgnoresMemory(t *testing.T) {
	dir := stampProject(t, true)
	home, _ := workspace.Home()
	md, _ := workspace.MachineDir(home, dir)
	writeFile(t, filepath.Join(md, workspace.MachineLabelsDir, "acme.yaml"), "format: bonsai.labels/1\nnamespace: acme\n"+
		"version: 1\nlabels:\n  - name: acme.cost\n    kind: number\n    values: []\n    items:\n    pattern:\n    max:\n"+
		"    kinds: [task, memory]\n    set_by: outside\n    grants: false\n    description: \"Dollars this work cost.\"\n")
	stampAnswer(t, dir, editIn(dir, stampNote, "labels: {}", "labels:\n  acme.cost: 3"), stampNow, RuleNotProtected)
	if code, s := runAt(t, dir, editIn(dir, stampTask, "labels:\n", "labels:\n  acme.cost: 3\n"), stampNow, false); code != 2 ||
		!strings.Contains(s, "rule: outside-label.") {
		t.Errorf("the same label on a task in command mode: exit %d, %s", code, s)
	}
}

// Rung 0 judges finished work hours after it was written: a branch whose new task, run report and STATE carry stamps
// far from the climb's clock passes; the fresh-stamp check is the guard's, as a call is made.
func TestRungZeroSkipsTheClock(t *testing.T) {
	dir := stampProject(t, false)
	j := NewChangeJudge(guardHome(), dir, NamedGrants{Why: "no task is named"}, false)
	for path, after := range map[string]string{
		"work/runs/R-2026-10-01-T-0001.md": strings.Replace(runText("2026-10-01 03:00", "2026-10-01 05:00"), "R-2026-10-21", "R-2026-10-01", 1),
		"work/tasks/T-0002-old.md":         strings.Replace(taskText("T-0002", "todo", "light"), "started:\n", "started: 2026-10-01\n", 1),
		stampState:                         stateText("2026-09-01"),
		stampNote:                          noteText("2026-09-01"),
	} {
		d := j.Judge(Change{Path: path, After: []byte(after), AfterKnown: true})
		if d.Rule == RuleTimeNotNow {
			t.Errorf("rung 0 judged %s by the clock: %s", path, d.Why)
		}
	}
}
