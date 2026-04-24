package removeflow

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestYield_AgentSuccessShowsCounts verifies the agent-remove Yield card
// shows the target agent display name, workspace, and per-category counts.
func TestYield_AgentSuccessShowsCounts(t *testing.T) {
	ctx := newTestCtx()
	counts := AbilityCounts{Skills: 3, Workflows: 2, Protocols: 1, Sensors: 4, Routines: 1}
	s := NewYieldAgentSuccess(ctx, "Backend", "services/api/", counts)
	s.SetSize(120, 40)
	out := s.View()

	for _, want := range []string{
		"UPROOTED",
		"Backend",
		"services/api/",
		"11", // total abilities
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("yield view missing %q; got:\n%s", want, out)
		}
	}
}

// TestYield_ItemSuccessShowsTargets verifies the item-remove Yield card
// lists the target agents the item was removed from.
func TestYield_ItemSuccessShowsTargets(t *testing.T) {
	ctx := newTestCtx()
	targets := []AgentOption{
		{Name: "tech-lead", DisplayName: "Tech Lead", Workspace: "station/"},
		{Name: "backend", DisplayName: "Backend", Workspace: "backend/"},
	}
	s := NewYieldItemSuccess(ctx, "Coding Standards", "Skill", targets)
	s.SetSize(120, 40)
	out := s.View()

	for _, want := range []string{
		"UPROOTED",
		"Coding Standards",
		"Tech Lead",
		"Backend",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("yield item view missing %q; got:\n%s", want, out)
		}
	}
}

// TestYield_HintsBlockPresent verifies the NEXT hints section renders with
// at least one next-step CLI suggestion. Plan 31 Phase E ships a 2-layer
// placeholder until Phase H lands.
func TestYield_HintsBlockPresent(t *testing.T) {
	ctx := newTestCtx()
	s := NewYieldAgentSuccess(ctx, "Backend", "backend/", AbilityCounts{})
	s.SetSize(100, 30)
	out := s.View()
	if !strings.Contains(out, "NEXT") {
		t.Fatalf("yield should include NEXT hints; got:\n%s", out)
	}
	if !strings.Contains(out, "$ bonsai") {
		t.Fatalf("yield hints should include at least one $ bonsai command; got:\n%s", out)
	}
}

// TestYield_EnterMarksDone verifies Enter / q / esc flip Done (terminal).
func TestYield_EnterMarksDone(t *testing.T) {
	ctx := newTestCtx()
	s := NewYieldAgentSuccess(ctx, "Backend", "backend/", AbilityCounts{})
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ys, ok := m.(*YieldStage); ok {
		*s = *ys
	}
	if !s.Done() {
		t.Fatal("Enter should MarkDone")
	}
}

// TestYield_Chromeless verifies Yield renders chromelessly.
func TestYield_Chromeless(t *testing.T) {
	ctx := newTestCtx()
	s := NewYieldAgentSuccess(ctx, "X", "x/", AbilityCounts{})
	if !s.Chromeless() {
		t.Fatal("YieldStage should be chromeless")
	}
}
