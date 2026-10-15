package stopgate

// The stall stops' exit (5.5.11 note 7): a named session whose state reads its budget spent and its report written
// may stop, its ladder red or not (stop-budget), with a record; spent but not reported, or reported in an earlier
// window, or a session that is not a named one, the gate is as before. On step 5.4's named-session fixture: the
// project gated for T-0901, its result made red.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/agents"
)

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
	gateRed := func() {
		t.Helper()
		code, stderr := gateFor(t, root, "T-0901")
		blocks(t, code, stderr, root, RuleRed, "is not green", installed+" ladder --task T-0901")
	}
	// Not spent: as before. Reported in an earlier window (a tie after it): as before.
	gateRed()
	add(agents.Event{E: agents.EventSpent, Measure: "actions", Count: 601, Of: 600})
	gateRed() // spent, not reported
	add(agents.Event{E: agents.EventReported})
	code, stderr := gateFor(t, root, "T-0901")
	if code != 0 || stderr != "" {
		t.Fatalf("a spent, reported named session: exit %d %q", code, stderr)
	}
	if r := lastRecord(t, root); r.String("decision") != "allow" || r.String("rule") != RuleBudget {
		t.Errorf("the record: %v", r)
	}
	add(agents.Event{E: agents.EventTie, Brief: "M-2026-10-15-0000000b"})
	gateRed() // re-briefed: a fresh window, not spent

	// A session that is not a named one (the person's own, an agent's state under another key) is as before.
	other := project(t)
	writeResult(t, other, "T-0901", red)
	o := agents.Store{Main: other}
	if _, err := o.Create(agents.Header{Agent: "s1", Session: "s1", Role: "builder", Kind: "lead", First: at}); err != nil {
		t.Fatal(err)
	}
	for _, e := range []agents.Event{{E: agents.EventSpent, At: at}, {E: agents.EventReported, At: at}} {
		if err := o.Append("s1", "s1", e); err != nil {
			t.Fatal(err)
		}
	}
	code, stderr = gateFor(t, other, "T-0901")
	if code != 2 || !strings.Contains(stderr, "is not green") {
		t.Errorf("a state that is not a named session's let the session stop: %d %q", code, stderr)
	}
}
