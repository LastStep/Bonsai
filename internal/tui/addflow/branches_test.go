package addflow

import (
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// newTestGraftCatalog builds a fixture catalog covering all five ability
// categories with a mix of required / default / plain items.
func newTestGraftCatalog() (*catalog.Catalog, *catalog.AgentDef) {
	none := catalog.AgentCompat{}
	all := catalog.AgentCompat{All: true}

	cat := &catalog.Catalog{
		Skills: []catalog.CatalogItem{
			{Name: "alpha-skill", DisplayName: "Alpha", Description: "a", Agents: all, Required: none, ContentPath: "skills/alpha/alpha.md"},
			{Name: "beta-skill", DisplayName: "Beta", Description: "b", Agents: all, Required: none, ContentPath: "skills/beta/beta.md"},
		},
		Workflows: []catalog.CatalogItem{
			{Name: "wf-one", DisplayName: "WF1", Description: "w1", Agents: all, Required: none, ContentPath: "workflows/one/one.md"},
		},
		Protocols: []catalog.CatalogItem{
			{Name: "proto-req", DisplayName: "Proto Req", Description: "p", Agents: all, Required: all, ContentPath: "protocols/req/req.md"},
		},
		Sensors: []catalog.SensorItem{
			{Name: "sensor-a", DisplayName: "Sensor A", Description: "s", Agents: all, Required: none, Event: "SessionStart", ContentPath: "sensors/a/a.sh"},
			{Name: "routine-check", DisplayName: "Routine Check", Description: "auto", Agents: all, Required: none, Event: "SessionStart", ContentPath: "sensors/rc/rc.sh"},
		},
		Routines: []catalog.RoutineItem{
			{Name: "routine-a", DisplayName: "Routine A", Description: "r", Agents: all, Required: none, Frequency: "7 days", ContentPath: "routines/a/a.md"},
		},
	}
	agentDef := &catalog.AgentDef{
		Name:          "backend",
		DisplayName:   "Backend",
		DefaultSkills: []string{"beta-skill"},
	}
	return cat, agentDef
}

func newTestGraft() *BranchesStage {
	cat, agentDef := newTestGraftCatalog()
	return NewNewAgentBranches(initflow.StageContext{
		StartedAt: time.Now(),
	}, BranchesContext{
		Cat:       cat,
		AgentType: "backend",
		AgentDef:  agentDef,
	})
}

func graftPressKey(s *BranchesStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if gs, ok := m.(*BranchesStage); ok {
		*s = *gs
	}
}

func graftPressRune(s *BranchesStage, r rune) {
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	if gs, ok := m.(*BranchesStage); ok {
		*s = *gs
	}
}

// TestGraft_FiltersRoutineCheck verifies the routine-check sensor is dropped
// from the sensors tab regardless of filter mode.
func TestGraft_FiltersRoutineCheck(t *testing.T) {
	s := newTestGraft()
	for _, c := range s.categories {
		if c.key != branchCatSensors {
			continue
		}
		for _, it := range c.items {
			if it.name == "routine-check" {
				t.Fatal("routine-check should be filtered from sensors tab")
			}
		}
	}
}

// TestGraft_RequiredPreSelected verifies required items are pre-selected and
// cannot be toggled off.
func TestGraft_RequiredPreSelected(t *testing.T) {
	s := newTestGraft()
	// proto-req is in category protocols.
	var protoIdx int = -1
	for i, c := range s.categories {
		if c.key == branchCatProtocols {
			protoIdx = i
			break
		}
	}
	if protoIdx < 0 {
		t.Fatal("protocols category missing")
	}
	if !s.selected[protoIdx]["proto-req"] {
		t.Fatal("proto-req should be pre-selected as required")
	}
	// Navigate to protocols tab + try to toggle.
	s.catIdx = protoIdx
	graftPressRune(s, ' ')
	if !s.selected[protoIdx]["proto-req"] {
		t.Fatal("␣ on required item should be no-op")
	}
}

// TestGraft_DefaultPreSelectedButToggleable verifies defaults are
// pre-selected and can be toggled off.
func TestGraft_DefaultPreSelectedButToggleable(t *testing.T) {
	s := newTestGraft()
	// beta-skill is default in category skills.
	var skillIdx int = -1
	for i, c := range s.categories {
		if c.key == branchCatSkills {
			skillIdx = i
			break
		}
	}
	if skillIdx < 0 {
		t.Fatal("skills category missing")
	}
	if !s.selected[skillIdx]["beta-skill"] {
		t.Fatal("beta-skill should be pre-selected as default")
	}
	// Focus beta-skill (second row, idx=1) and toggle off.
	s.catIdx = skillIdx
	s.itemIdx[skillIdx] = 1
	graftPressRune(s, ' ')
	if s.selected[skillIdx]["beta-skill"] {
		t.Fatal("␣ on default should toggle off")
	}
}

// TestGraft_TabCycles verifies ← → cycles tabs with wrap.
func TestGraft_TabCycles(t *testing.T) {
	s := newTestGraft()
	start := s.catIdx
	graftPressKey(s, tea.KeyRight)
	if s.catIdx != start+1 {
		t.Fatalf("right tab = %d, want %d", s.catIdx, start+1)
	}
	// Walk to end + wrap.
	for range s.categories {
		graftPressKey(s, tea.KeyRight)
	}
	if s.catIdx != start+1 {
		t.Fatalf("after full cycle, catIdx = %d, want %d (wrap)", s.catIdx, start+1)
	}
}

// TestGraft_ExpandToggle verifies ? flips the expanded flag.
func TestGraft_ExpandToggle(t *testing.T) {
	s := newTestGraft()
	if s.expanded {
		t.Fatal("expanded should start false")
	}
	graftPressRune(s, '?')
	if !s.expanded {
		t.Fatal("? should set expanded=true")
	}
	graftPressRune(s, '?')
	if s.expanded {
		t.Fatal("?? should toggle back to false")
	}
}

// TestGraft_EnterAdvances verifies Enter flips done.
func TestGraft_EnterAdvances(t *testing.T) {
	s := newTestGraft()
	if s.Done() {
		t.Fatal("should not be done before Enter")
	}
	graftPressKey(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("should be done after Enter")
	}
}

// TestGraft_ResultShape verifies Result returns a BranchesResult with the
// expected pre-selected items (required + defaults).
func TestGraft_ResultShape(t *testing.T) {
	s := newTestGraft()
	res, ok := s.Result().(BranchesResult)
	if !ok {
		t.Fatalf("Result type = %T, want BranchesResult", s.Result())
	}
	if !reflect.DeepEqual(res.Skills, []string{"beta-skill"}) {
		t.Fatalf("Skills = %v, want [beta-skill]", res.Skills)
	}
	if !reflect.DeepEqual(res.Protocols, []string{"proto-req"}) {
		t.Fatalf("Protocols = %v, want [proto-req]", res.Protocols)
	}
	if len(res.Workflows) != 0 {
		t.Fatalf("Workflows = %v, want empty", res.Workflows)
	}
}

// TestGraft_AddItemsFiltersInstalled verifies add-items mode filters out
// items already in the installed agent + drops empty tabs.
func TestGraft_AddItemsFiltersInstalled(t *testing.T) {
	cat, agentDef := newTestGraftCatalog()
	installed := &config.InstalledAgent{
		AgentType: "backend",
		Workspace: "backend/",
		Skills:    []string{"alpha-skill", "beta-skill"},
		Workflows: []string{"wf-one"},
		Protocols: []string{"proto-req"},
		Sensors:   []string{"sensor-a"},
		Routines:  []string{"routine-a"},
	}
	s := NewAddItemsBranches(initflow.StageContext{StartedAt: time.Now()}, BranchesContext{
		Cat:       cat,
		AgentType: "backend",
		AgentDef:  agentDef,
		Installed: installed,
	})
	// Every tab should be dropped — each category had all items installed.
	if len(s.categories) != 0 {
		t.Fatalf("add-items with everything installed: len(categories) = %d, want 0", len(s.categories))
	}
}

// TestGraft_AddItemsPartialDropsEmpty verifies add-items mode drops only the
// tabs that go empty after filtering.
func TestGraft_AddItemsPartialDropsEmpty(t *testing.T) {
	cat, agentDef := newTestGraftCatalog()
	installed := &config.InstalledAgent{
		AgentType: "backend",
		Workspace: "backend/",
		Skills:    []string{"alpha-skill"}, // leaves beta-skill available
		Workflows: []string{"wf-one"},      // empty after filter
		Protocols: []string{"proto-req"},   // empty after filter
		Sensors:   []string{"sensor-a"},    // empty after filter
		Routines:  []string{"routine-a"},   // empty after filter
	}
	s := NewAddItemsBranches(initflow.StageContext{StartedAt: time.Now()}, BranchesContext{
		Cat:       cat,
		AgentType: "backend",
		AgentDef:  agentDef,
		Installed: installed,
	})
	if len(s.categories) != 1 {
		t.Fatalf("len(categories) = %d, want 1 (only skills should remain)", len(s.categories))
	}
	if s.categories[0].key != branchCatSkills {
		t.Fatalf("remaining tab = %q, want skills", s.categories[0].key)
	}
	if len(s.categories[0].items) != 1 || s.categories[0].items[0].name != "beta-skill" {
		t.Fatalf("skills tab items = %v, want [beta-skill]", s.categories[0].items)
	}
}

// TestGraft_TabCountUpdatesLive verifies the per-tab "(N)" counter rendered
// in the tab strip reflects toggles as the user makes them — not a captured
// ctor-time snapshot. Render the tab row before/after a toggle and compare
// the focused tab's rendered count.
func TestGraft_TabCountUpdatesLive(t *testing.T) {
	s := newTestGraft()
	// Find skills tab.
	var skillIdx = -1
	for i, c := range s.categories {
		if c.key == branchCatSkills {
			skillIdx = i
			break
		}
	}
	if skillIdx < 0 {
		t.Fatal("skills tab missing")
	}

	// Initial: beta-skill is a default (pre-selected). Count = 1.
	before := len(s.selected[skillIdx])
	if before != 1 {
		t.Fatalf("initial skills count = %d, want 1", before)
	}

	// Focus skills + toggle alpha-skill on (idx 0).
	s.catIdx = skillIdx
	s.itemIdx[skillIdx] = 0
	graftPressRune(s, ' ')
	after := len(s.selected[skillIdx])
	if after != 2 {
		t.Fatalf("after toggle skills count = %d, want 2", after)
	}

	// Rendered tab row should contain "SKILLS (2)" now, not "(1)".
	row := s.renderTabs()
	if !strings.Contains(row, "SKILLS (2)") {
		t.Fatalf("tab row should contain SKILLS (2); got: %q", row)
	}

	// Toggle alpha back off — count returns to 1.
	graftPressRune(s, ' ')
	if len(s.selected[skillIdx]) != 1 {
		t.Fatalf("after second toggle skills count = %d, want 1", len(s.selected[skillIdx]))
	}
	row = s.renderTabs()
	if !strings.Contains(row, "SKILLS (1)") {
		t.Fatalf("tab row should contain SKILLS (1); got: %q", row)
	}
}

// TestGraft_NewAgentUnaffectedByFilter is a regression guard — the new-agent
// ctor must NOT filter by Installed even when a non-nil Installed slice is
// passed (Plan 23 Phase 2 added the filter path inside a shared helper; this
// test proves the new-agent entry still seeds every catalog item regardless).
func TestGraft_NewAgentUnaffectedByFilter(t *testing.T) {
	cat, agentDef := newTestGraftCatalog()
	// Installed carries every item in every category — if the new-agent path
	// were mistakenly filtering, the tabs would all be empty.
	installed := &config.InstalledAgent{
		AgentType: "backend",
		Workspace: "backend/",
		Skills:    []string{"alpha-skill", "beta-skill"},
		Workflows: []string{"wf-one"},
		Protocols: []string{"proto-req"},
		Sensors:   []string{"sensor-a"},
		Routines:  []string{"routine-a"},
	}
	s := NewNewAgentBranches(initflow.StageContext{StartedAt: time.Now()}, BranchesContext{
		Cat:       cat,
		AgentType: "backend",
		AgentDef:  agentDef,
		Installed: installed, // deliberately non-nil to prove the new-agent path ignores it
	})
	// All five tabs should survive.
	if len(s.categories) != 5 {
		t.Fatalf("new-agent ctor tabs = %d, want 5 (filter leaked into new-agent path)", len(s.categories))
	}
	// Skills tab should carry both alpha + beta — filter leak would drop them.
	var skillIdx = -1
	for i, c := range s.categories {
		if c.key == branchCatSkills {
			skillIdx = i
			break
		}
	}
	if skillIdx < 0 || len(s.categories[skillIdx].items) != 2 {
		t.Fatalf("new-agent skills tab items = %v, want both alpha + beta",
			s.categories[skillIdx].items)
	}
}

// TestGraft_ResetPreservesState verifies Reset clears done but preserves
// selections + cursor + expanded.
func TestGraft_ResetPreservesState(t *testing.T) {
	s := newTestGraft()
	// Move around + toggle.
	graftPressKey(s, tea.KeyRight)
	graftPressRune(s, '?')
	graftPressKey(s, tea.KeyEnter)
	catIdx := s.catIdx
	expanded := s.expanded
	s.Reset()
	if s.Done() {
		t.Fatal("Reset should clear done")
	}
	if s.catIdx != catIdx {
		t.Fatalf("Reset changed catIdx — got %d, want %d", s.catIdx, catIdx)
	}
	if s.expanded != expanded {
		t.Fatalf("Reset changed expanded — got %v, want %v", s.expanded, expanded)
	}
}
