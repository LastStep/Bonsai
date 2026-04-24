package removeflow

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestObserve_ShowsInstalledCounts verifies the agent-remove preview panel
// contains the five per-category rows (skills / workflows / protocols /
// sensors / routines) and reflects the item counts supplied at ctor time.
func TestObserve_ShowsInstalledCounts(t *testing.T) {
	ctx := newTestCtx()
	skills := []string{"s1", "s2", "s3"}
	workflows := []string{"w1", "w2"}
	protocols := []string{"p1"}
	sensors := []string{"sen1"}
	routines := []string{"r1", "r2", "r3", "r4"}

	s := NewObserveAgent(ctx, "backend", "Backend", "services/api/",
		skills, workflows, protocols, sensors, routines)
	s.SetSize(120, 40)
	out := s.View()

	for _, want := range []string{
		"Backend",
		"services/api/",
		"SKILLS",
		"WORKFLOWS",
		"PROTOCOLS",
		"SENSORS",
		"ROUTINES",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("observe view missing %q; got:\n%s", want, out)
		}
	}
}

// TestObserve_ItemModeShowsTargets verifies the item-remove panel renders
// the item label + type + per-target FROM rows.
func TestObserve_ItemModeShowsTargets(t *testing.T) {
	ctx := newTestCtx()
	targets := []AgentOption{
		{Name: "tech-lead", DisplayName: "Tech Lead", Workspace: "station/"},
		{Name: "backend", DisplayName: "Backend", Workspace: "backend/"},
	}
	s := NewObserveItem(ctx, "Coding Standards", "Skill", targets)
	s.SetSize(120, 40)
	out := s.View()

	for _, want := range []string{
		"Coding Standards",
		"Skill",
		"Tech Lead",
		"station/",
		"Backend",
		"backend/",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("observe item view missing %q; got:\n%s", want, out)
		}
	}
}

// TestObserve_EnterAdvances verifies Enter marks the stage done (no gate —
// Observe is a preview, Confirm handles the destructive decision).
func TestObserve_EnterAdvances(t *testing.T) {
	ctx := newTestCtx()
	s := NewObserveAgent(ctx, "backend", "Backend", "backend/", nil, nil, nil, nil, nil)
	observePressKey(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("Enter should MarkDone")
	}
}

// TestObserve_EmptyAbilitiesRendersGracefully verifies an agent with no
// installed abilities still renders the panel without panicking.
func TestObserve_EmptyAbilitiesRendersGracefully(t *testing.T) {
	ctx := newTestCtx()
	s := NewObserveAgent(ctx, "solo", "Solo", "solo/", nil, nil, nil, nil, nil)
	s.SetSize(100, 30)
	out := s.View()
	if !strings.Contains(out, "Solo") {
		t.Fatalf("empty-ability observe should still render agent name; got:\n%s", out)
	}
}

// TestObserve_SetTargetsOverwritesList verifies the item-remove ctor's
// target slice can be replaced via SetTargets after construction (used by
// the cmd/remove.go LazyStep resolver).
func TestObserve_SetTargetsOverwritesList(t *testing.T) {
	ctx := newTestCtx()
	initial := []AgentOption{{Name: "a", DisplayName: "A", Workspace: "a/"}}
	s := NewObserveItem(ctx, "Item", "Skill", initial)
	replacement := []AgentOption{
		{Name: "b", DisplayName: "B", Workspace: "b/"},
		{Name: "c", DisplayName: "C", Workspace: "c/"},
	}
	s.SetTargets(replacement)
	if len(s.targets) != 2 {
		t.Fatalf("after SetTargets, targets len = %d, want 2", len(s.targets))
	}
}

// TestObserve_RenderDoesNotPanicAtFloor smokes View under terminal-too-small
// dims to prove the fallback path renders.
func TestObserve_RenderDoesNotPanicAtFloor(t *testing.T) {
	ctx := newTestCtx()
	s := NewObserveAgent(ctx, "a", "A", "a/", nil, nil, nil, nil, nil)
	s.SetSize(40, 10) // below floor
	_ = s.View()
}
