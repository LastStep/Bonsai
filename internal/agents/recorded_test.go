package agents

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// An agent's state as recorded: its limits from its last tie line, else its start line; minutes from its last call
// line; actions the call lines since its window's start; red runs from its last red or reset line; spent in the window
// in force. The agent that tied a brief, by its tie file.
func TestRecorded(t *testing.T) {
	s := store(t)
	if _, err := s.Recorded(sess, agent); err != ErrNoState {
		t.Fatalf("no state: %v", err)
	}
	if _, err := s.Create(header()); err != nil {
		t.Fatal(err)
	}
	r, err := s.Recorded(sess, agent)
	if err != nil || r.Minutes != nil || r.Actions != 0 || !r.LastCall.IsZero() || *r.Limits.Actions.Value != 250 || r.Spent {
		t.Fatalf("before any call: %+v %v", r, err)
	}
	for i := int64(0); i < 3; i++ {
		if _, err := s.Call(sess, agent, t0.Add(time.Duration(i)*time.Minute), i); err != nil {
			t.Fatal(err)
		}
	}
	lim := header().Limits
	lim.Actions.Value = n(5)
	must := func(e Event) {
		t.Helper()
		if err := s.Append(sess, agent, e); err != nil {
			t.Fatal(err)
		}
	}
	must(Event{E: EventTie, At: StampMilli(t0.Add(2 * time.Minute)), Brief: "M-2026-10-15-0000000a", Limits: &lim, Calls: 2})
	must(Event{E: EventRed, At: StampMilli(t0.Add(3 * time.Minute)), Task: "T-0001", N: 1})
	if _, err := s.Call(sess, agent, t0.Add(4*time.Minute), 2); err != nil {
		t.Fatal(err)
	}
	must(Event{E: EventSpent, At: StampMilli(t0.Add(4 * time.Minute)), Measure: "red_runs", Count: 1, Of: 0})
	r, err = s.Recorded(sess, agent)
	if err != nil || *r.Minutes != 2 || r.Actions != 2 || r.RedRuns != 1 || !r.Spent || *r.Limits.Actions.Value != 5 ||
		!r.LastCall.Equal(t0.Add(4*time.Minute)) {
		t.Errorf("recorded: %+v %v", r, err)
	}
	// The agent that tied a brief.
	if _, _, ok := s.TiedTo("M-2026-10-15-0000000a"); ok {
		t.Error("tied before its tie file")
	}
	if ok, err := s.Tie("M-2026-10-15-0000000a", sess, agent); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if se, a, ok := s.TiedTo("M-2026-10-15-0000000a"); !ok || se != sess || a != agent {
		t.Errorf("tied to %s/%s %v", se, a, ok)
	}
	if err := os.WriteFile(filepath.Join(s.Dir(), TiesFolder, "M-2026-10-15-0000000b"), []byte("no-slash\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := s.TiedTo("M-2026-10-15-0000000b"); ok {
		t.Error("a tie file naming no agent read as a tie")
	}
}
