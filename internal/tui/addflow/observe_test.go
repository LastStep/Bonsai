package addflow

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

func newTestObserve() *ObserveStage {
	cat := &catalog.Catalog{
		Agents: []catalog.AgentDef{
			{Name: "backend", DisplayName: "Backend", Description: "b"},
		},
	}
	return NewObserveStage(initflow.StageContext{
		Version:      "test",
		ProjectDir:   "/tmp/observe-test",
		StationDir:   "station/",
		AgentDisplay: "",
		StartedAt:    time.Now(),
	}, cat)
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

// TestObserve_SetPriorCapturesFields verifies SetPrior pulls the three
// upstream results into the stage.
func TestObserve_SetPriorCapturesFields(t *testing.T) {
	s := newTestObserve()
	graft := GraftResult{
		Skills:    []string{"s1"},
		Workflows: []string{"w1"},
	}
	s.SetPrior([]any{"backend", "services/api/", graft})
	if s.agent != "backend" {
		t.Fatalf("agent = %q, want backend", s.agent)
	}
	if s.workspace != "services/api/" {
		t.Fatalf("workspace = %q, want services/api/", s.workspace)
	}
	if len(s.graft.Skills) != 1 || s.graft.Skills[0] != "s1" {
		t.Fatalf("graft.Skills = %v, want [s1]", s.graft.Skills)
	}
	if s.agentDef == nil || s.agentDef.Name != "backend" {
		t.Fatal("agentDef should be resolved from catalog")
	}
	if s.agentDisplay != "Backend" {
		t.Fatalf("agentDisplay = %q, want Backend", s.agentDisplay)
	}
}

// TestObserve_SetPriorMissingAgentTolerant verifies SetPrior tolerates an
// unknown agent without panicking.
func TestObserve_SetPriorMissingAgentTolerant(t *testing.T) {
	s := newTestObserve()
	s.SetPrior([]any{"unknown-agent", "x/", GraftResult{}})
	if s.agent != "unknown-agent" {
		t.Fatalf("agent = %q, want unknown-agent", s.agent)
	}
	if s.agentDef != nil {
		t.Fatal("unknown agent should leave agentDef nil")
	}
}

// TestObserve_SetPriorAddItemsShape verifies SetPrior handles the add-items
// branch's shorter prev[] slice (no Ground stage → no workspace string).
// GraftResult is the second slot and must still be captured by type.
func TestObserve_SetPriorAddItemsShape(t *testing.T) {
	s := newTestObserve()
	s.SetDefaultWorkspace("backend/")
	graft := GraftResult{
		Skills:   []string{"s1"},
		Sensors:  []string{"sensor-a"},
		Routines: []string{"routine-a"},
	}
	// prev layout on add-items: [agent, graft, ...].
	s.SetPrior([]any{"backend", graft})
	if s.agent != "backend" {
		t.Fatalf("agent = %q, want backend", s.agent)
	}
	if s.workspace != "backend/" {
		t.Fatalf("workspace = %q, want backend/ (preserved from SetDefaultWorkspace)", s.workspace)
	}
	if len(s.graft.Skills) != 1 || s.graft.Skills[0] != "s1" {
		t.Fatalf("graft.Skills = %v, want [s1]", s.graft.Skills)
	}
	if len(s.graft.Sensors) != 1 || len(s.graft.Routines) != 1 {
		t.Fatalf("graft not fully captured: %+v", s.graft)
	}
}

// TestObserve_DefaultFocusGraft verifies btnFocus starts on GRAFT (1).
func TestObserve_DefaultFocusGraft(t *testing.T) {
	s := newTestObserve()
	if s.btnFocus != 1 {
		t.Fatalf("btnFocus = %d, want 1 (GRAFT default)", s.btnFocus)
	}
}

// TestObserve_TabTogglesButtons verifies tab flips button focus.
func TestObserve_TabTogglesButtons(t *testing.T) {
	s := newTestObserve()
	observePressKey(s, tea.KeyTab)
	if s.btnFocus != 0 {
		t.Fatalf("after Tab btnFocus = %d, want 0", s.btnFocus)
	}
	observePressKey(s, tea.KeyTab)
	if s.btnFocus != 1 {
		t.Fatalf("after 2× Tab btnFocus = %d, want 1", s.btnFocus)
	}
}

// TestObserve_YConfirmsGraft verifies y shortcut sets confirmed + done.
func TestObserve_YConfirmsGraft(t *testing.T) {
	s := newTestObserve()
	observePressRune(s, 'y')
	if !s.Done() {
		t.Fatal("y should MarkDone")
	}
	if got, _ := s.Result().(bool); !got {
		t.Fatal("y should set Result=true")
	}
}

// TestObserve_NCancels verifies n shortcut cancels.
func TestObserve_NCancels(t *testing.T) {
	s := newTestObserve()
	observePressRune(s, 'n')
	if !s.Done() {
		t.Fatal("n should MarkDone")
	}
	if got, _ := s.Result().(bool); got {
		t.Fatal("n should set Result=false")
	}
}

// TestObserve_EnterOnGraftConfirms verifies Enter with GRAFT focus confirms.
func TestObserve_EnterOnGraftConfirms(t *testing.T) {
	s := newTestObserve()
	// btnFocus defaults to 1 (GRAFT).
	observePressKey(s, tea.KeyEnter)
	if got, _ := s.Result().(bool); !got {
		t.Fatal("Enter on GRAFT should confirm")
	}
}

// TestObserve_EnterOnBackCancels verifies Enter with BACK focus cancels.
func TestObserve_EnterOnBackCancels(t *testing.T) {
	s := newTestObserve()
	observePressKey(s, tea.KeyTab) // move to BACK
	observePressKey(s, tea.KeyEnter)
	if got, _ := s.Result().(bool); got {
		t.Fatal("Enter on BACK should cancel")
	}
}

// TestObserve_ViewRendersWorkspace verifies the workspace path appears in
// the rendered view.
func TestObserve_ViewRendersWorkspace(t *testing.T) {
	s := newTestObserve()
	s.SetSize(120, 40)
	s.SetPrior([]any{"backend", "services/api/", GraftResult{Skills: []string{"s1"}}})
	out := s.View()
	if !strings.Contains(out, "services/api/") {
		t.Fatal("view should contain workspace path")
	}
}

// TestObserve_ResetClearsConfirmation verifies Reset clears done and
// confirmation state but preserves prior.
func TestObserve_ResetClearsConfirmation(t *testing.T) {
	s := newTestObserve()
	s.SetPrior([]any{"backend", "ws/", GraftResult{}})
	observePressRune(s, 'y')
	s.Reset()
	if s.Done() {
		t.Fatal("Reset should clear done")
	}
	if got, _ := s.Result().(bool); got {
		t.Fatal("Reset should clear confirmed")
	}
	if s.agent != "backend" {
		t.Fatal("Reset should preserve prior")
	}
}
