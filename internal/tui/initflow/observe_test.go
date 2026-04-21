package initflow

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/catalog"
)

func newTestObserve() *ObserveStage {
	all := catalog.AgentCompat{All: true}
	none := catalog.AgentCompat{}
	cat := &catalog.Catalog{
		Skills: []catalog.CatalogItem{
			{Name: "alpha", DisplayName: "Alpha", Agents: all, Required: none},
		},
	}
	agentDef := &catalog.AgentDef{Name: "tech-lead", DisplayName: "Tech Lead"}
	s := NewObserveStage(StageContext{
		Version:      "test",
		ProjectDir:   "/tmp/obs",
		StationDir:   "station/",
		AgentDisplay: "Tech Lead",
		StartedAt:    time.Now(),
	}, cat, agentDef)
	// Stamp a plausible terminal size so the renderer doesn't short-circuit.
	s.width = 120
	s.height = 40
	return s
}

func observePressKey(s *ObserveStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if os, ok := m.(*ObserveStage); ok {
		*s = *os
	}
}

func observePressRune(s *ObserveStage, r rune) {
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	if os, ok := m.(*ObserveStage); ok {
		*s = *os
	}
}

// TestObserve_DefaultFocusPlant verifies the CTA starts on PLANT so a bare
// ↵ ships the happy path per plan.
func TestObserve_DefaultFocusPlant(t *testing.T) {
	s := newTestObserve()
	if s.btnFocus != 1 {
		t.Fatalf("initial btnFocus = %d, want 1 (PLANT)", s.btnFocus)
	}
}

// TestObserve_EnterConfirmsPlant verifies ↵ with PLANT focused → Result=true.
func TestObserve_EnterConfirmsPlant(t *testing.T) {
	s := newTestObserve()
	observePressKey(s, tea.KeyEnter)
	if !s.done {
		t.Fatalf("done=false after Enter")
	}
	if got, _ := s.Result().(bool); !got {
		t.Fatalf("Result = %v, want true", s.Result())
	}
}

// TestObserve_EnterCancels verifies ↵ with CANCEL focused → Result=false.
func TestObserve_EnterCancels(t *testing.T) {
	s := newTestObserve()
	observePressKey(s, tea.KeyTab) // toggle to CANCEL
	if s.btnFocus != 0 {
		t.Fatalf("Tab did not move focus to CANCEL")
	}
	observePressKey(s, tea.KeyEnter)
	if got, _ := s.Result().(bool); got {
		t.Fatalf("Result = true, want false (cancelled)")
	}
}

// TestObserve_YConfirms verifies y shortcut confirms regardless of focus.
func TestObserve_YConfirms(t *testing.T) {
	s := newTestObserve()
	observePressKey(s, tea.KeyTab) // focus CANCEL
	observePressRune(s, 'y')
	if got, _ := s.Result().(bool); !got {
		t.Fatalf("Result = false after y, want true")
	}
}

// TestObserve_NCancels verifies n shortcut cancels regardless of focus.
func TestObserve_NCancels(t *testing.T) {
	s := newTestObserve()
	observePressRune(s, 'n')
	if got, _ := s.Result().(bool); got {
		t.Fatalf("Result = true after n, want false")
	}
}

// TestObserve_TabTogglesFocus verifies Tab and Left/Right cycle the
// CANCEL / PLANT focus.
func TestObserve_TabTogglesFocus(t *testing.T) {
	s := newTestObserve()
	observePressKey(s, tea.KeyTab)
	if s.btnFocus != 0 {
		t.Fatalf("after Tab from PLANT, btnFocus = %d, want 0", s.btnFocus)
	}
	observePressKey(s, tea.KeyRight)
	if s.btnFocus != 1 {
		t.Fatalf("after Right, btnFocus = %d, want 1", s.btnFocus)
	}
	observePressKey(s, tea.KeyLeft)
	if s.btnFocus != 0 {
		t.Fatalf("after Left, btnFocus = %d, want 0", s.btnFocus)
	}
}

// TestObserve_SetPriorSnapshot verifies SetPrior captures all three prior
// stage results into the observe summary state.
func TestObserve_SetPriorSnapshot(t *testing.T) {
	s := newTestObserve()
	prev := []any{
		map[string]string{"name": "voyager", "description": "api svc", "station": "station/"},
		[]string{"agents-index", "session-log"},
		BranchesResult{
			Skills:    []string{"alpha", "beta"},
			Workflows: []string{"planning"},
			Protocols: []string{"memory"},
			Sensors:   []string{"scope-guard"},
			Routines:  []string{"backlog-hygiene"},
		},
	}
	s.SetPrior(prev)

	if s.vessel["name"] != "voyager" {
		t.Errorf("vessel name = %q, want voyager", s.vessel["name"])
	}
	if len(s.soil) != 2 {
		t.Errorf("soil len = %d, want 2", len(s.soil))
	}
	if len(s.branches.Skills) != 2 {
		t.Errorf("branches skills len = %d, want 2", len(s.branches.Skills))
	}
}

// TestObserve_BodyIncludesSummary verifies the rendered body contains the
// expected summary markers when the three prior stages have been captured.
func TestObserve_BodyIncludesSummary(t *testing.T) {
	s := newTestObserve()
	s.SetPrior([]any{
		map[string]string{"name": "voyager", "description": "svc", "station": "station/"},
		[]string{"agents-index"},
		BranchesResult{Skills: []string{"alpha"}},
	})
	v := s.View()
	// Strip ANSI (styles) for robust substring check.
	for _, want := range []string{"VESSEL", "SOIL", "BRANCHES", "voyager", "agents-index", "alpha"} {
		if !strings.Contains(v, want) {
			t.Errorf("View missing %q", want)
		}
	}
}

// TestObserve_ResponsiveStackedNarrow verifies the body uses a stacked
// (single-column) layout below 100 cols. We check by confirming the
// left block's content appears above the right block's content in the
// rendered output at narrow widths.
func TestObserve_ResponsiveStackedNarrow(t *testing.T) {
	s := newTestObserve()
	s.width = 80
	s.height = 30
	s.SetPrior([]any{
		map[string]string{"name": "voyager", "description": "svc", "station": "station/"},
		[]string{"agents-index"},
		BranchesResult{Skills: []string{"alpha"}},
	})
	v := s.View()
	// In stacked mode VESSEL must appear before BRANCHES (which appears
	// before SOIL in our stacked order).
	vIdx := strings.Index(v, "VESSEL")
	bIdx := strings.Index(v, "BRANCHES")
	if vIdx < 0 || bIdx < 0 {
		t.Fatalf("missing VESSEL/BRANCHES markers in narrow render")
	}
	if vIdx >= bIdx {
		t.Errorf("narrow-width layout: VESSEL (idx %d) should precede BRANCHES (idx %d)", vIdx, bIdx)
	}
}

// TestObserve_MinSizeFloor verifies that below the min-size threshold the
// stage falls through to the floor panel — renderFrame short-circuits.
func TestObserve_MinSizeFloor(t *testing.T) {
	s := newTestObserve()
	s.width = 60
	s.height = 16
	v := s.View()
	if !strings.Contains(v, "please enlarge") {
		t.Errorf("min-size render missing floor panel text; got:\n%s", v)
	}
}

// TestObserve_ResetClearsConfirm verifies Reset wipes confirmation so a
// follow-up Enter re-advances freshly (matches the harness resetter
// contract).
func TestObserve_ResetClearsConfirm(t *testing.T) {
	s := newTestObserve()
	observePressKey(s, tea.KeyEnter)
	if !s.done || !s.confirmed {
		t.Fatalf("pre-reset state: done=%v confirmed=%v", s.done, s.confirmed)
	}
	s.Reset()
	if s.done || s.confirmed {
		t.Errorf("Reset left done=%v confirmed=%v, want both false", s.done, s.confirmed)
	}
}
