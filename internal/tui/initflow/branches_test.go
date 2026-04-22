package initflow

import (
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/catalog"
)

// newTestBranches constructs a BranchesStage from a minimal in-memory
// catalog + AgentDef fixture. Keeps the test independent of the on-disk
// catalog loader — we want to assert behaviour over a known item shape.
func newTestBranches() *BranchesStage {
	skillReq := catalog.AgentCompat{Names: []string{"tech-lead"}}
	none := catalog.AgentCompat{}
	all := catalog.AgentCompat{All: true}

	cat := &catalog.Catalog{
		Skills: []catalog.CatalogItem{
			{Name: "alpha-skill", DisplayName: "Alpha Skill", Description: "first skill", Agents: all, Required: skillReq, ContentPath: "skills/alpha/alpha.md"},
			{Name: "beta-skill", DisplayName: "Beta Skill", Description: "second skill", Agents: all, Required: none, ContentPath: "skills/beta/beta.md"},
			{Name: "gamma-skill", DisplayName: "Gamma Skill", Description: "third skill", Agents: all, Required: none, ContentPath: "skills/gamma/gamma.md"},
		},
		Workflows: []catalog.CatalogItem{
			{Name: "wf-one", DisplayName: "WF One", Description: "workflow one", Agents: all, Required: none, ContentPath: "workflows/one/one.md"},
			{Name: "wf-two", DisplayName: "WF Two", Description: "workflow two", Agents: all, Required: none, ContentPath: "workflows/two/two.md"},
		},
		Protocols: []catalog.CatalogItem{
			{Name: "proto-req", DisplayName: "Proto Req", Description: "required protocol", Agents: all, Required: all, ContentPath: "protocols/req/req.md"},
		},
		Sensors: []catalog.SensorItem{
			{Name: "sensor-one", DisplayName: "Sensor One", Description: "first sensor", Agents: all, Required: none, Event: "SessionStart", ContentPath: "sensors/one/one.sh"},
			{Name: "sensor-two", DisplayName: "Sensor Two", Description: "second sensor", Agents: all, Required: none, Event: "SessionStart", ContentPath: "sensors/two/two.sh"},
		},
		Routines: []catalog.RoutineItem{
			{Name: "routine-a", DisplayName: "Routine A", Description: "first routine", Agents: all, Required: none, Frequency: "7 days", ContentPath: "routines/a/a.md"},
		},
	}

	agentDef := &catalog.AgentDef{
		Name:             "tech-lead",
		DisplayName:      "Tech Lead",
		DefaultSkills:    []string{"beta-skill"},
		DefaultWorkflows: []string{"wf-one"},
		DefaultProtocols: nil,
		DefaultSensors:   []string{"sensor-two"},
		DefaultRoutines:  nil,
	}

	return NewBranchesStage(StageContext{
		Version:      "test",
		ProjectDir:   "/tmp/branches-test",
		StationDir:   "station/",
		AgentDisplay: "Tech Lead",
		StartedAt:    time.Now(),
	}, cat, agentDef)
}

func branchesPressKey(s *BranchesStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if bs, ok := m.(*BranchesStage); ok {
		*s = *bs
	}
}

func branchesPressRune(s *BranchesStage, r rune) {
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	if bs, ok := m.(*BranchesStage); ok {
		*s = *bs
	}
}

// TestBranches_DefaultsApplied verifies that required items and default items
// are both selected on first render. Required = "proto-req"; defaults =
// beta-skill, wf-one, sensor-two.
func TestBranches_DefaultsApplied(t *testing.T) {
	s := newTestBranches()

	// skills (idx 0): beta-skill default → selected.
	if !s.selected[0]["beta-skill"] {
		t.Fatalf("beta-skill default not pre-selected")
	}
	if s.selected[0]["alpha-skill"] {
		// alpha-skill is required per fixture → MUST be selected too.
	} else {
		t.Fatalf("alpha-skill required not pre-selected")
	}
	// gamma-skill has neither flag → NOT selected.
	if s.selected[0]["gamma-skill"] {
		t.Fatalf("gamma-skill (no default/required) was pre-selected")
	}
	// workflows (idx 1): wf-one default.
	if !s.selected[1]["wf-one"] {
		t.Fatalf("wf-one default not pre-selected")
	}
	// protocols (idx 2): proto-req required (agents: all).
	if !s.selected[2]["proto-req"] {
		t.Fatalf("proto-req required not pre-selected")
	}
	// sensors (idx 3): sensor-two default.
	if !s.selected[3]["sensor-two"] {
		t.Fatalf("sensor-two default not pre-selected")
	}
	// routines (idx 4): nothing default / required.
	if len(s.selected[4]) != 0 {
		t.Fatalf("routines tab expected empty pre-selection, got %v", s.selected[4])
	}
}

// TestBranches_TabCyclingRight verifies Right cycles through the 5 tabs and
// wraps back to 0.
func TestBranches_TabCyclingRight(t *testing.T) {
	s := newTestBranches()
	if s.catIdx != 0 {
		t.Fatalf("initial catIdx = %d, want 0", s.catIdx)
	}
	for want := 1; want < len(s.categories); want++ {
		branchesPressKey(s, tea.KeyRight)
		if s.catIdx != want {
			t.Fatalf("after %d Right press, catIdx = %d, want %d", want, s.catIdx, want)
		}
	}
	// One more Right → wrap to 0.
	branchesPressKey(s, tea.KeyRight)
	if s.catIdx != 0 {
		t.Fatalf("after wrap Right, catIdx = %d, want 0", s.catIdx)
	}
}

// TestBranches_TabCyclingLeft verifies Left wraps from 0 to the last tab.
func TestBranches_TabCyclingLeft(t *testing.T) {
	s := newTestBranches()
	branchesPressKey(s, tea.KeyLeft)
	if s.catIdx != len(s.categories)-1 {
		t.Fatalf("after Left from 0, catIdx = %d, want %d", s.catIdx, len(s.categories)-1)
	}
}

// TestBranches_ItemFocusClamps verifies Down clamps at the last item (no
// wrap) and Up clamps at 0.
func TestBranches_ItemFocusClamps(t *testing.T) {
	s := newTestBranches()
	// skills tab has 3 items (indexes 0,1,2).
	total := len(s.categories[0].items)
	if total != 3 {
		t.Fatalf("fixture regression: skills tab expected 3 items, got %d", total)
	}
	// Press Down past the end — should clamp at total-1.
	for i := 0; i < total+5; i++ {
		branchesPressKey(s, tea.KeyDown)
	}
	if s.itemIdx[0] != total-1 {
		t.Fatalf("after many Down, itemIdx = %d, want %d (clamp)", s.itemIdx[0], total-1)
	}
	// Press Up past the start — should clamp at 0.
	for i := 0; i < total+5; i++ {
		branchesPressKey(s, tea.KeyUp)
	}
	if s.itemIdx[0] != 0 {
		t.Fatalf("after many Up, itemIdx = %d, want 0 (clamp)", s.itemIdx[0])
	}
}

// TestBranches_ToggleNonRequired verifies Space on a non-required item flips
// its selected state in both directions.
func TestBranches_ToggleNonRequired(t *testing.T) {
	s := newTestBranches()
	// skills tab: move focus to gamma-skill (idx 2, neither default nor required).
	branchesPressKey(s, tea.KeyDown)
	branchesPressKey(s, tea.KeyDown)
	if s.itemIdx[0] != 2 {
		t.Fatalf("focus not at 2, got %d", s.itemIdx[0])
	}
	if s.selected[0]["gamma-skill"] {
		t.Fatalf("fixture regression: gamma-skill should start unselected")
	}
	branchesPressRune(s, ' ')
	if !s.selected[0]["gamma-skill"] {
		t.Fatalf("Space did not select gamma-skill")
	}
	branchesPressRune(s, ' ')
	if s.selected[0]["gamma-skill"] {
		t.Fatalf("Space did not deselect gamma-skill")
	}
}

// TestBranches_ToggleRequiredNoOp verifies Space on a required item leaves it
// selected (cannot be toggled off).
func TestBranches_ToggleRequiredNoOp(t *testing.T) {
	s := newTestBranches()
	// skills tab, item 0 = alpha-skill (required in fixture).
	if !s.categories[0].items[0].required {
		t.Fatalf("fixture regression: alpha-skill expected to be required")
	}
	branchesPressRune(s, ' ')
	if !s.selected[0]["alpha-skill"] {
		t.Fatalf("Space on required item deselected it")
	}
}

// TestBranches_ExpandToggle verifies `?` flips the expanded flag.
func TestBranches_ExpandToggle(t *testing.T) {
	s := newTestBranches()
	if s.expanded {
		t.Fatalf("initial expanded = true, want false")
	}
	branchesPressRune(s, '?')
	if !s.expanded {
		t.Fatalf("? did not set expanded=true")
	}
	branchesPressRune(s, '?')
	if s.expanded {
		t.Fatalf("? did not toggle expanded back to false")
	}
}

// TestBranches_ResultShape verifies Result() returns a BranchesResult with
// per-category slices in catalog order, and required items are always
// present even if never explicitly toggled.
func TestBranches_ResultShape(t *testing.T) {
	s := newTestBranches()
	// Skills defaults: alpha (required), beta (default) — gamma unpicked.
	// Toggle gamma ON to verify it joins the result.
	branchesPressKey(s, tea.KeyDown)
	branchesPressKey(s, tea.KeyDown)
	branchesPressRune(s, ' ')

	// Switch to workflows tab, toggle wf-two ON.
	branchesPressKey(s, tea.KeyRight)
	branchesPressKey(s, tea.KeyDown)
	branchesPressRune(s, ' ')

	res, ok := s.Result().(BranchesResult)
	if !ok {
		t.Fatalf("Result() not a BranchesResult: %T", s.Result())
	}
	wantSkills := []string{"alpha-skill", "beta-skill", "gamma-skill"}
	if !reflect.DeepEqual(res.Skills, wantSkills) {
		t.Fatalf("Skills = %v, want %v", res.Skills, wantSkills)
	}
	wantWF := []string{"wf-one", "wf-two"}
	if !reflect.DeepEqual(res.Workflows, wantWF) {
		t.Fatalf("Workflows = %v, want %v", res.Workflows, wantWF)
	}
	wantProto := []string{"proto-req"}
	if !reflect.DeepEqual(res.Protocols, wantProto) {
		t.Fatalf("Protocols = %v, want %v", res.Protocols, wantProto)
	}
	wantSensors := []string{"sensor-two"}
	if !reflect.DeepEqual(res.Sensors, wantSensors) {
		t.Fatalf("Sensors = %v, want %v", res.Sensors, wantSensors)
	}
	if len(res.Routines) != 0 {
		t.Fatalf("Routines = %v, want empty", res.Routines)
	}
}

// TestBranches_ResultCatalogOrder verifies Result() slices preserve the
// catalog's iteration order, not the user's toggle order.
func TestBranches_ResultCatalogOrder(t *testing.T) {
	s := newTestBranches()
	// Toggle gamma first (idx 2), then re-select by pressing Space on beta
	// (default). This proves order is by catalog iteration, not toggle order.
	branchesPressKey(s, tea.KeyDown)
	branchesPressKey(s, tea.KeyDown)
	branchesPressRune(s, ' ') // gamma ON
	branchesPressKey(s, tea.KeyUp)
	branchesPressRune(s, ' ') // beta OFF
	branchesPressRune(s, ' ') // beta ON

	res := s.Result().(BranchesResult)
	want := []string{"alpha-skill", "beta-skill", "gamma-skill"}
	if !reflect.DeepEqual(res.Skills, want) {
		t.Fatalf("Skills order = %v, want %v (must follow catalog order)", res.Skills, want)
	}
}

// TestBranches_EnterCompletes verifies ↵ flips the done flag.
func TestBranches_EnterCompletes(t *testing.T) {
	s := newTestBranches()
	branchesPressKey(s, tea.KeyEnter)
	if !s.done {
		t.Fatalf("done=false after Enter; expected stage advance")
	}
}

// TestBranches_ResetPreservesState verifies Reset clears done but keeps
// catIdx, per-tab selected, per-tab itemIdx, and expanded all intact — so
// Esc-and-return restores the stage verbatim.
func TestBranches_ResetPreservesState(t *testing.T) {
	s := newTestBranches()
	// Build up some non-default state: jump to sensors tab, toggle sensor-one,
	// move focus down, turn on expand.
	branchesPressKey(s, tea.KeyRight)
	branchesPressKey(s, tea.KeyRight)
	branchesPressKey(s, tea.KeyRight) // now on sensors (idx 3)
	if s.catIdx != 3 {
		t.Fatalf("catIdx = %d, want 3", s.catIdx)
	}
	branchesPressRune(s, ' ') // toggle focused item (sensor-one — idx 0 — now ON)
	branchesPressKey(s, tea.KeyDown)
	branchesPressRune(s, '?') // expand

	// Snapshot state pre-reset.
	savedCat := s.catIdx
	savedExpand := s.expanded
	savedItemIdx := make(map[int]int, len(s.itemIdx))
	for k, v := range s.itemIdx {
		savedItemIdx[k] = v
	}
	savedSelected := make(map[int]map[string]bool, len(s.selected))
	for k, v := range s.selected {
		cp := make(map[string]bool, len(v))
		for kk, vv := range v {
			cp[kk] = vv
		}
		savedSelected[k] = cp
	}

	// Enter to flip done, then Reset.
	branchesPressKey(s, tea.KeyEnter)
	if !s.done {
		t.Fatalf("done=false after Enter")
	}
	s.Reset()

	if s.done {
		t.Fatalf("Reset did not clear done")
	}
	if s.catIdx != savedCat {
		t.Fatalf("Reset changed catIdx: %d -> %d", savedCat, s.catIdx)
	}
	if s.expanded != savedExpand {
		t.Fatalf("Reset changed expanded: %v -> %v", savedExpand, s.expanded)
	}
	if !reflect.DeepEqual(s.itemIdx, savedItemIdx) {
		t.Fatalf("Reset changed itemIdx: %v -> %v", savedItemIdx, s.itemIdx)
	}
	if !reflect.DeepEqual(s.selected, savedSelected) {
		t.Fatalf("Reset changed selected: %v -> %v", savedSelected, s.selected)
	}
}

// TestBranches_SelectionPersistsAcrossTabs verifies toggles in one tab don't
// affect other tabs, and switching back restores the tab's selection.
func TestBranches_SelectionPersistsAcrossTabs(t *testing.T) {
	s := newTestBranches()
	// Toggle gamma-skill ON in skills tab.
	branchesPressKey(s, tea.KeyDown)
	branchesPressKey(s, tea.KeyDown)
	branchesPressRune(s, ' ')
	if !s.selected[0]["gamma-skill"] {
		t.Fatalf("gamma-skill not selected after Space")
	}

	// Switch to sensors tab, toggle sensor-one.
	branchesPressKey(s, tea.KeyRight)
	branchesPressKey(s, tea.KeyRight)
	branchesPressKey(s, tea.KeyRight)
	branchesPressRune(s, ' ')

	// Skills selection is untouched.
	if !s.selected[0]["gamma-skill"] {
		t.Fatalf("gamma-skill lost after tab-switch")
	}
	if !s.selected[0]["alpha-skill"] {
		t.Fatalf("alpha-skill (required) lost after tab-switch")
	}

	// Come back to skills, focus should still be at gamma.
	branchesPressKey(s, tea.KeyLeft)
	branchesPressKey(s, tea.KeyLeft)
	branchesPressKey(s, tea.KeyLeft)
	if s.itemIdx[0] != 2 {
		t.Fatalf("skills focus not restored: itemIdx[0] = %d, want 2", s.itemIdx[0])
	}
}

// TestBranches_NarrowDoesNotClipRequiredGlyph verifies the required `*`
// glyph stays visible on the row across narrow (but ≥floor) widths.
// Replaces the pre-2026-04-22 tag-column assertion — "(required)" /
// "DEFAULT" text was dropped in favour of an inline "*" after the name.
func TestBranches_NarrowDoesNotClipRequiredGlyph(t *testing.T) {
	s := newTestBranches()
	s.width = 80
	s.height = 30
	row := s.renderRow(0)
	if !strings.Contains(row, "*") {
		t.Errorf("80-col render dropped required * glyph; row: %q", row)
	}
	s.width = 70
	row = s.renderRow(0)
	if !strings.Contains(row, "*") {
		t.Errorf("70-col render dropped required * glyph; row: %q", row)
	}
}

// TestBranches_MinSizeFloor verifies tiny terminals show the floor.
func TestBranches_MinSizeFloor(t *testing.T) {
	s := newTestBranches()
	s.width = 60
	s.height = 16
	if !strings.Contains(s.View(), "please enlarge") {
		t.Error("min-size render missing floor panel")
	}
}

// TestBranches_ListScrollsWhenLong verifies the Viewport wraps the list
// when catalog entries exceed the available body height. Seeds a tab
// with 20 items and checks that Follow(19) slides the viewport down.
func TestBranches_ListScrollsWhenLong(t *testing.T) {
	s := newTestBranches()
	s.width = 120
	s.height = 30 // listHeight → 30-10-22 = -2 → floor 3

	// Replace the skills tab items with 20 synthetic entries so the
	// viewport has something to scroll.
	fake := make([]branchItem, 20)
	for i := range fake {
		fake[i] = branchItem{
			name:        "sk-" + string(rune('a'+i%26)),
			displayName: "Skill " + string(rune('A'+i%26)),
			description: "fake",
		}
	}
	s.categories[0].items = fake
	s.itemIdx[0] = 19 // focus at the last row
	// renderList should produce exactly listHeight() lines.
	rendered := s.renderList()
	lines := strings.Split(rendered, "\n")
	if len(lines) != s.listHeight() {
		t.Errorf("renderList lines = %d, want %d (listHeight)", len(lines), s.listHeight())
	}
	// The focused row's display name must appear in the rendered output.
	if !strings.Contains(rendered, "Skill T") { // idx 19 % 26 = 19 → 'T'
		t.Errorf("focused row Skill T not visible after Follow; output:\n%s", rendered)
	}
}
