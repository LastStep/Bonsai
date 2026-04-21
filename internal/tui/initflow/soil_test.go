package initflow

import (
	"reflect"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// testSoilOptions returns a deterministic fixture mirroring the typical
// bonsai scaffolding catalog — two required items followed by three
// optional ones.
func testSoilOptions() []ScaffoldingOption {
	return []ScaffoldingOption{
		{Name: "claude-md", DisplayName: "CLAUDE.md", Description: "root agent directive", Required: true},
		{Name: "agents-index", DisplayName: "agents-index", Description: "directory of every agent", Required: true},
		{Name: "session-log", DisplayName: "session-log", Description: "rolling log of what each session did"},
		{Name: "readme-stub", DisplayName: "readme-stub", Description: "a starter README"},
		{Name: "editor-config", DisplayName: "editor-config", Description: "editorconfig file"},
	}
}

func newTestSoil() *SoilStage {
	return NewSoilStage(StageContext{
		Version:      "test",
		ProjectDir:   "/tmp/soil-test",
		StationDir:   "station/",
		AgentDisplay: "Tech Lead",
		StartedAt:    time.Now(),
	}, testSoilOptions())
}

func soilPress(s *SoilStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if ss, ok := m.(*SoilStage); ok {
		*s = *ss
	}
}

func soilSpace(s *SoilStage) {
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	if ss, ok := m.(*SoilStage); ok {
		*s = *ss
	}
}

// TestSoil_RequiredPreSelected verifies every Required option starts in the
// selected state on construction.
func TestSoil_RequiredPreSelected(t *testing.T) {
	s := newTestSoil()
	for i, opt := range s.options {
		if opt.Required && !s.selected[i] {
			t.Fatalf("required option %q at index %d not pre-selected", opt.Name, i)
		}
	}
}

// TestSoil_RequiredCannotToggle verifies Space on a Required row does not
// flip its selected flag.
func TestSoil_RequiredCannotToggle(t *testing.T) {
	s := newTestSoil()
	// focus=0 is a Required option per fixture.
	if !s.options[0].Required {
		t.Fatalf("fixture regression: index 0 expected to be Required")
	}
	before := s.selected[0]
	soilSpace(s)
	if s.selected[0] != before {
		t.Fatalf("Required option flipped selected state via Space: %v -> %v",
			before, s.selected[0])
	}
}

// TestSoil_OptionalToggle verifies Space on an optional row flips its
// selected flag both directions.
func TestSoil_OptionalToggle(t *testing.T) {
	s := newTestSoil()
	// Move focus to index 2 (first optional).
	soilPress(s, tea.KeyDown)
	soilPress(s, tea.KeyDown)
	if s.focus != 2 {
		t.Fatalf("focus = %d, want 2", s.focus)
	}
	if s.options[2].Required {
		t.Fatalf("fixture regression: index 2 expected to be optional")
	}

	if s.selected[2] {
		t.Fatalf("fixture regression: optional index 2 should start unselected")
	}

	soilSpace(s)
	if !s.selected[2] {
		t.Fatalf("Space on optional didn't select it")
	}
	soilSpace(s)
	if s.selected[2] {
		t.Fatalf("Space on optional didn't deselect it")
	}
}

// TestSoil_FocusAdvance verifies arrow-key focus movement and wrap-around.
func TestSoil_FocusAdvance(t *testing.T) {
	s := newTestSoil()
	total := len(s.options)
	if s.focus != 0 {
		t.Fatalf("initial focus = %d, want 0", s.focus)
	}
	soilPress(s, tea.KeyDown)
	if s.focus != 1 {
		t.Fatalf("after Down focus = %d, want 1", s.focus)
	}
	// Cycle to the end + one more → wrap to 0.
	for i := 1; i < total; i++ {
		soilPress(s, tea.KeyDown)
	}
	if s.focus != 0 {
		t.Fatalf("wrap focus = %d, want 0", s.focus)
	}
	// Up from 0 → last index.
	soilPress(s, tea.KeyUp)
	if s.focus != total-1 {
		t.Fatalf("reverse wrap focus = %d, want %d", s.focus, total-1)
	}
}

// TestSoil_ResultOrder verifies Result() returns selected item names in the
// original option order, regardless of toggle order.
func TestSoil_ResultOrder(t *testing.T) {
	s := newTestSoil()
	// Fixture: index 0,1 Required → pre-selected. Toggle index 4 first, then
	// index 2 — Result should still order 0,1,2,4 per original option order.
	s.focus = 4
	soilSpace(s)
	s.focus = 2
	soilSpace(s)

	got := s.Result().([]string)
	want := []string{"claude-md", "agents-index", "session-log", "editor-config"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Result order = %v, want %v", got, want)
	}
}

// TestSoil_EnterCompletes verifies ↵ flips the done flag so the harness
// can advance to the next stage.
func TestSoil_EnterCompletes(t *testing.T) {
	s := newTestSoil()
	soilPress(s, tea.KeyEnter)
	if !s.done {
		t.Fatalf("done=false after Enter; expected stage advance")
	}
}

// TestSoil_ResetPreservesSelections verifies Reset clears done but keeps the
// selected slice + focus cursor intact so Esc-and-return doesn't erase the
// user's work.
func TestSoil_ResetPreservesSelections(t *testing.T) {
	s := newTestSoil()
	s.focus = 3
	soilSpace(s)
	picked := make([]bool, len(s.selected))
	copy(picked, s.selected)
	prevFocus := s.focus

	soilPress(s, tea.KeyEnter) // done=true
	s.Reset()

	if s.done {
		t.Fatalf("Reset did not clear done flag")
	}
	if !reflect.DeepEqual(s.selected, picked) {
		t.Fatalf("Reset modified selections: %v -> %v", picked, s.selected)
	}
	if s.focus != prevFocus {
		t.Fatalf("Reset moved focus: %d -> %d", prevFocus, s.focus)
	}
}

// TestSoil_NarrowDoesNotClipBadge verifies the REQUIRED badge stays
// visible on narrow (but ≥floor) widths. Regression guard.
func TestSoil_NarrowDoesNotClipBadge(t *testing.T) {
	s := newTestSoil()
	s.width = 80
	s.height = 30
	row := s.renderRow(0) // claude-md is required in fixture
	if !soilContains(row, "REQUIRED") {
		t.Errorf("80-col render dropped REQUIRED badge; row: %q", row)
	}
	s.width = 70
	row = s.renderRow(0)
	if !soilContains(row, "REQUIRED") {
		t.Errorf("70-col render dropped REQUIRED badge; row: %q", row)
	}
}

// TestSoil_MinSizeFloor verifies tiny terminals route to the floor panel.
func TestSoil_MinSizeFloor(t *testing.T) {
	s := newTestSoil()
	s.width = 60
	s.height = 16
	if !soilContains(s.View(), "please enlarge") {
		t.Errorf("min-size render missing floor panel")
	}
}

// soilContains is a tiny substring helper so the test file stays import-light.
func soilContains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
