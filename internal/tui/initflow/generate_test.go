package initflow

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestGenerate(action GenerateAction) *GenerateStage {
	s := NewGenerateStage(StageContext{
		Version:      "test",
		ProjectDir:   "/tmp/gen",
		StationDir:   "station/",
		AgentDisplay: "Tech Lead",
		StartedAt:    time.Now(),
	}, action)
	s.width = 120
	s.height = 40
	return s
}

// TestGenerate_MinHoldEnforced verifies an instant action still blocks the
// stage in stateMinHold until `minGenerateHold` has elapsed. We drive the
// state machine directly via the message types — no real-time Tick waiting
// so the test runs fast.
func TestGenerate_MinHoldEnforced(t *testing.T) {
	var ran atomic.Int32
	s := newTestGenerate(func() error {
		ran.Add(1)
		return nil
	})

	// Init starts the goroutine; in tests we don't run the tea.Program, so
	// simulate the lifecycle by driving msgs through Update manually.
	// startedAt is initialized via Init.
	s.startedAt = time.Now()
	// Instantly-finished action: state transitions to minHold because
	// elapsed < minGenerateHold.
	m, _ := s.Update(generateDoneMsg{err: nil, elapsed: 10 * time.Millisecond})
	s = m.(*GenerateStage)
	if s.state != stateMinHold {
		t.Fatalf("state after instant done = %v, want stateMinHold", s.state)
	}
	if s.done {
		t.Fatalf("Done() flipped before minHold elapsed")
	}

	// Simulate a tick before the floor → still not done.
	m, _ = s.Update(generateTickMsg{})
	s = m.(*GenerateStage)
	if s.done {
		t.Fatalf("Done() flipped on first tick — min hold not enforced")
	}

	// Force startedAt into the past and tick again — must transition to
	// stateDone.
	s.startedAt = time.Now().Add(-2 * minGenerateHold)
	m, _ = s.Update(generateTickMsg{})
	s = m.(*GenerateStage)
	if s.state != stateDone {
		t.Fatalf("state after min-hold-elapsed tick = %v, want stateDone", s.state)
	}
	if !s.done {
		t.Fatalf("Done() not true after min hold elapsed")
	}
}

// TestGenerate_ActionErrorRoutesToStateError verifies a failing action
// moves the stage to stateError — not stateDone.
func TestGenerate_ActionErrorRoutesToStateError(t *testing.T) {
	sentinel := errors.New("boom")
	s := newTestGenerate(func() error { return sentinel })

	s.startedAt = time.Now()
	m, _ := s.Update(generateDoneMsg{err: sentinel, elapsed: time.Millisecond})
	s = m.(*GenerateStage)
	if s.state != stateError {
		t.Fatalf("state after err = %v, want stateError", s.state)
	}
	if s.done {
		t.Fatalf("Done() flipped on error (should wait for Enter)")
	}
	if got := s.Result(); got != sentinel {
		t.Fatalf("Result = %v, want sentinel error", got)
	}

	// Enter acknowledges and advances.
	m, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	s = m.(*GenerateStage)
	if !s.done {
		t.Fatalf("Done() not set after Enter on stateError")
	}
}

// TestGenerate_ArcScales verifies arc size scales down at narrow widths.
// We render the body at wide and narrow and count non-blank lines in the
// arc portion — 12-row arc should be taller than 8-row arc.
func TestGenerate_ArcScales(t *testing.T) {
	s := newTestGenerate(func() error { return nil })

	s.width = 120
	s.height = 40
	wideBody := s.renderBody()
	wideLines := len(strings.Split(wideBody, "\n"))

	s.width = 80
	narrowBody := s.renderBody()
	narrowLines := len(strings.Split(narrowBody, "\n"))

	if wideLines <= narrowLines {
		t.Errorf("wide render has %d lines, narrow %d — expected wide > narrow", wideLines, narrowLines)
	}
}

// TestGenerate_BodyContainsKanjiOrFallback verifies the centre character
// is present based on ensoSafe. Don't hard-assert which — fallback depends
// on env.
func TestGenerate_BodyContainsKanjiOrFallback(t *testing.T) {
	s := newTestGenerate(func() error { return nil })
	body := s.renderBody()
	if s.ensoSafe {
		if !strings.Contains(body, "生") {
			t.Errorf("safe render missing 生 kanji")
		}
	} else {
		if !strings.Contains(body, "O") {
			t.Errorf("ascii render missing centre")
		}
	}
}

// TestGenerate_ResultNilOnSuccess verifies Result returns nil after a
// successful action path (used as the proceed/abort signal for the
// Phase 5B Conditional step).
func TestGenerate_ResultNilOnSuccess(t *testing.T) {
	s := newTestGenerate(func() error { return nil })
	s.startedAt = time.Now().Add(-2 * minGenerateHold)
	m, _ := s.Update(generateDoneMsg{err: nil, elapsed: time.Millisecond})
	s = m.(*GenerateStage)
	// Tick to promote minHold → done.
	m, _ = s.Update(generateTickMsg{})
	s = m.(*GenerateStage)
	if got := s.Result(); got != nil {
		t.Errorf("Result on success = %v, want nil", got)
	}
}

// TestGenerate_TickAnimatesArc verifies successive ticks advance the lit
// count (more cells are "●" / "#" on later frames). Compares rendered
// arc at tick=1 vs tick=50.
func TestGenerate_TickAnimatesArc(t *testing.T) {
	s := newTestGenerate(func() error { return nil })
	s.ticks = 1
	early := s.renderBody()
	s.ticks = 50
	late := s.renderBody()
	if early == late {
		t.Errorf("arc did not change between tick=1 and tick=50")
	}
}

// TestGenerateStage_BodyOnlyDropsChrome verifies that when the stage is
// flipped into body-only mode (Plan 27 PR2 §C7 — Grow), the rendered View
// omits both the enso-rail glyphs and the footer BONSAI brand. The full-
// chrome View() path is covered by the inverse assertion in other tests.
func TestGenerateStage_BodyOnlyDropsChrome(t *testing.T) {
	s := NewGenerateStage(StageContext{
		Version:      "test",
		ProjectDir:   "/tmp/gen",
		StationDir:   "station/",
		AgentDisplay: "Tech Lead",
		StartedAt:    time.Now(),
	}, func() error { return nil })
	s.SetBodyOnly(true)
	s.SetSize(120, 30)

	out := s.View()
	// Enso rail glyphs: ● pending-dot, ○ done-dot, ─ connector. Any of them
	// appearing would mean rail chrome leaked through.
	for _, glyph := range []string{ensoDone, ensoPending} {
		if strings.Contains(out, glyph) {
			t.Errorf("body-only View should not contain rail glyph %q; got:\n%s", glyph, out)
		}
	}
	// Footer brand — the "一 BONSAI 一" pattern from RenderFooter.
	if strings.Contains(out, "BONSAI 一") {
		t.Errorf("body-only View should not contain footer brand; got:\n%s", out)
	}
}
