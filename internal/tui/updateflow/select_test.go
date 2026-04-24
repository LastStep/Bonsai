package updateflow

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// buildSelectFixture returns a SelectStage over 2 agents, each with 2
// valid discoveries — matches the plan's "per-agent tab switching" test
// shape.
func buildSelectFixture() *SelectStage {
	agents := []AgentDiscoveries{
		{
			AgentName:  "tech-lead",
			AgentLabel: "Tech Lead",
			Installed:  &config.InstalledAgent{AgentType: "tech-lead", Workspace: "station/"},
			Valid: []generate.DiscoveredFile{
				{Name: "skill-a", Type: "skill", RelPath: "station/agent/Skills/skill-a.md", Meta: &config.CustomItemMeta{Description: "alpha"}},
				{Name: "skill-b", Type: "skill", RelPath: "station/agent/Skills/skill-b.md", Meta: &config.CustomItemMeta{Description: "beta"}},
			},
		},
		{
			AgentName:  "backend",
			AgentLabel: "Backend",
			Installed:  &config.InstalledAgent{AgentType: "backend", Workspace: "backend/"},
			Valid: []generate.DiscoveredFile{
				{Name: "skill-c", Type: "skill", RelPath: "backend/agent/Skills/skill-c.md", Meta: &config.CustomItemMeta{Description: "gamma"}},
				{Name: "skill-d", Type: "skill", RelPath: "backend/agent/Skills/skill-d.md", Meta: &config.CustomItemMeta{Description: "delta"}},
			},
		},
	}
	s := NewSelectStage(initflow.StageContext{StartedAt: time.Now()}, agents)
	s.SetSize(120, 40)
	return s
}

func selectPressKey(s *SelectStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if ss, ok := m.(*SelectStage); ok {
		*s = *ss
	}
}
func selectPressRune(s *SelectStage, r rune) {
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	if ss, ok := m.(*SelectStage); ok {
		*s = *ss
	}
}

// TestSelect_DefaultsAllSelected — on construction every valid file is
// pre-selected (mirrors legacy huh.Selected(true) contract).
func TestSelect_DefaultsAllSelected(t *testing.T) {
	s := buildSelectFixture()
	for i, row := range s.selected {
		for j, v := range row {
			if !v {
				t.Fatalf("agent %d row %d = false, want true (default all-selected)", i, j)
			}
		}
	}
}

// TestSelect_PerAgentTabSwitching — ←→ cycles between agent tabs and
// the per-tab focus resets to 0.
func TestSelect_PerAgentTabSwitching(t *testing.T) {
	s := buildSelectFixture()
	if s.tab != 0 {
		t.Fatalf("initial tab = %d, want 0", s.tab)
	}
	// Move to second agent.
	selectPressKey(s, tea.KeyRight)
	if s.tab != 1 {
		t.Fatalf("after right tab = %d, want 1", s.tab)
	}
	if s.focus != 0 {
		t.Fatalf("tab change should reset focus to 0; got %d", s.focus)
	}
	// Clamp at end.
	selectPressKey(s, tea.KeyRight)
	if s.tab != 1 {
		t.Fatalf("right at end should clamp; got %d", s.tab)
	}
	// Left back to first.
	selectPressKey(s, tea.KeyLeft)
	if s.tab != 0 {
		t.Fatalf("after left tab = %d, want 0", s.tab)
	}
}

// TestSelect_SpaceToggles — space on the focused file flips its
// selection bit; the lock contract says toggling on/off only affects
// the focused row.
func TestSelect_SpaceToggles(t *testing.T) {
	s := buildSelectFixture()
	// Initial: all selected. Space on row 0 → deselect.
	selectPressRune(s, ' ')
	if s.selected[0][0] {
		t.Fatal("space should deselect focused row")
	}
	// Row 1 untouched.
	if !s.selected[0][1] {
		t.Fatal("space should NOT touch non-focused row")
	}
	// Second space → reselect.
	selectPressRune(s, ' ')
	if !s.selected[0][0] {
		t.Fatal("second space should reselect focused row")
	}
}

// TestSelect_FocusWithinTabIndependent — moving focus inside one tab
// does not change the other tab's focus state (each tab stores its own).
func TestSelect_FocusWithinTabIndependent(t *testing.T) {
	s := buildSelectFixture()
	// Tab 0 — move focus down.
	selectPressKey(s, tea.KeyDown)
	if s.focus != 1 {
		t.Fatalf("tab 0 focus = %d, want 1", s.focus)
	}
	// Tab change — focus resets to 0 on new tab.
	selectPressKey(s, tea.KeyRight)
	if s.focus != 0 {
		t.Fatalf("tab-change focus = %d, want 0", s.focus)
	}
}

// TestSelect_EnterCompletes — enter marks the stage done.
func TestSelect_EnterCompletes(t *testing.T) {
	s := buildSelectFixture()
	selectPressKey(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("enter should flip Done")
	}
}

// TestSelect_SelectedKeysShapesResult — the SelectedKeys map groups the
// selected "type:name" keys per agent name.
func TestSelect_SelectedKeysShapesResult(t *testing.T) {
	s := buildSelectFixture()
	// Deselect tech-lead row 1.
	selectPressKey(s, tea.KeyDown)
	selectPressRune(s, ' ')

	keys := s.SelectedKeys()
	if len(keys["tech-lead"]) != 1 {
		t.Fatalf("tech-lead keys len = %d, want 1 (got %v)", len(keys["tech-lead"]), keys["tech-lead"])
	}
	if keys["tech-lead"][0] != "skill:skill-a" {
		t.Fatalf("tech-lead[0] = %q, want skill:skill-a", keys["tech-lead"][0])
	}
	if len(keys["backend"]) != 2 {
		t.Fatalf("backend keys len = %d, want 2", len(keys["backend"]))
	}
}

// TestSelect_ToggleAllInTab — 'a' deselects every row in the current
// tab when all are selected, then selects all when any is unselected.
func TestSelect_ToggleAllInTab(t *testing.T) {
	s := buildSelectFixture()
	// All selected → 'a' deselects all.
	selectPressRune(s, 'a')
	for _, v := range s.selected[0] {
		if v {
			t.Fatal("'a' on all-selected should deselect all")
		}
	}
	// 'a' again → selects all.
	selectPressRune(s, 'a')
	for _, v := range s.selected[0] {
		if !v {
			t.Fatal("'a' on all-deselected should select all")
		}
	}
	// Other tab untouched.
	for _, v := range s.selected[1] {
		if !v {
			t.Fatal("'a' should NOT affect other tabs")
		}
	}
}
