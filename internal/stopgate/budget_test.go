package stopgate

// The stall stops' exit (5.5.11 note 7): a named session whose state reads its budget spent, and which has written
// its report since, may stop, its ladder red or not (stop-budget), with a record naming the report. The report is the
// message file itself, of the kind report, from the session's own address, created at or after its spent line, with
// its message record in the log (VG's item C: the guard's note that the report's call was let through is never
// enough, so a report command that failed leaves the session held). Otherwise, and for a session that is not a named
// one, the gate is as before. On step 5.4's named-session fixture: the project gated for T-0901, its result made red.

import (
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// reportFile is a report's message file, bonsai.message/1, from from to the orchestrator.
func reportFile(id, from string, created time.Time) string {
	return "---\nformat: bonsai.message/1   # fields: bonsai check --schema bonsai.message\nid: " + id + "\nthread: " + id +
		"\nreply_to: null\ntask: T-0901\nkind: report\nfrom: " + from + "\nto: orchestrator\nstatus: blocked\nbudget: null\n" +
		"pointers: []\ncreated: \"" + created.UTC().Format(record.AtLayout) + "\"\nlabels: {}\n---\nWhat I have, what is left.\n"
}

// writeReport writes a report as bonsai message would: its file, and with recorded its message record in the
// session's log.
func writeReport(t *testing.T, root, id, from string, created time.Time, recorded bool) {
	t.Helper()
	write(t, root, ".bonsai/local/messages/orchestrator/"+id+".md", reportFile(id, from, created))
	if !recorded {
		return
	}
	l := record.New("message", record.Common{Workspace: testID, Local: workspace.Local{Root: root, Main: root}, Session: "s1",
		Agent: "claude-code", Getenv: func(string) string { return "" }, At: created})
	l.Target = &id
	if _, err := record.WriteLog(root, l); err != nil {
		t.Fatal(err)
	}
}

func TestTheBudgetExit(t *testing.T) {
	root := project(t)
	red := greenResult(headSHA)
	red.Green = false
	writeResult(t, root, "T-0901", red)
	s := agents.Store{Main: root}
	at := agents.Stamp(testNow)
	if _, err := s.Create(agents.Header{Agent: "s1", Session: "s1", Role: "builder", Kind: "lead", Writes: true, Named: true,
		Task: "T-0901", First: at}); err != nil {
		t.Fatal(err)
	}
	add := func(e agents.Event) {
		t.Helper()
		e.At = at
		if err := s.Append("s1", "s1", e); err != nil {
			t.Fatal(err)
		}
	}
	gateRed := func(why string) {
		t.Helper()
		code, stderr := gateFor(t, root, "T-0901")
		if code != 2 || !strings.Contains(stderr, "is not green") {
			t.Fatalf("%s: exit %d %q, want the red ladder's block", why, code, stderr)
		}
	}
	gateRed("not spent")
	add(agents.Event{E: agents.EventSpent, Measure: "actions", Count: 601, Of: 600})
	gateRed("spent, no report")
	// VG's reproducer: the report's call was let through (the guard's reported line), but bonsai message refused it
	// (a thread that does not exist), so no report was written: held.
	add(agents.Event{E: agents.EventReported})
	gateRed("spent, its report's call let through, no report written")
	// Reports that are not its own report past its budget: one written before its spent line, one from another
	// address, one bonsai message did not write (no message record), and a brief.
	writeReport(t, root, "M-2026-10-11-0000000a", "builder@T-0901", testNow.Add(-time.Second), true)
	writeReport(t, root, "M-2026-10-11-0000000b", "builder@T-0902", testNow.Add(time.Second), true)
	writeReport(t, root, "M-2026-10-11-0000000c", "builder@T-0901", testNow.Add(time.Second), false)
	write(t, root, ".bonsai/local/messages/orchestrator/M-2026-10-11-0000000d.md",
		strings.Replace(reportFile("M-2026-10-11-0000000d", "builder@T-0901", testNow.Add(time.Second)), "kind: report", "kind: brief", 1))
	gateRed("no report of its own past its budget")
	// Its report, written since, recorded: it may stop, the record naming the report.
	writeReport(t, root, "M-2026-10-11-0000000e", "builder@T-0901", testNow.Add(2*time.Second), true)
	code, stderr := gateFor(t, root, "T-0901")
	if code != 0 || stderr != "" {
		t.Fatalf("a spent, reported named session: exit %d %q", code, stderr)
	}
	if r := lastRecord(t, root); r.String("decision") != "allow" || r.String("rule") != RuleBudget ||
		r.String("target") != ".bonsai/local/messages/orchestrator/M-2026-10-11-0000000e.md" {
		t.Errorf("the record: %v", r)
	}
	add(agents.Event{E: agents.EventTie, Brief: "M-2026-10-15-0000000b", To: "builder@T-0901"})
	gateRed("re-briefed: a fresh window, not spent")

	// A session that is not a named one (the person's own, an agent's state under another key) is as before.
	other := project(t)
	writeResult(t, other, "T-0901", red)
	o := agents.Store{Main: other}
	if _, err := o.Create(agents.Header{Agent: "s1", Session: "s1", Role: "builder", Kind: "lead", Task: "T-0901", First: at}); err != nil {
		t.Fatal(err)
	}
	if err := o.Append("s1", "s1", agents.Event{E: agents.EventSpent, At: at}); err != nil {
		t.Fatal(err)
	}
	writeReport(t, other, "M-2026-10-11-0000000e", "builder@T-0901", testNow.Add(2*time.Second), true)
	code, stderr = gateFor(t, other, "T-0901")
	if code != 2 || !strings.Contains(stderr, "is not green") {
		t.Errorf("a state that is not a named session's let the session stop: %d %q", code, stderr)
	}
}
