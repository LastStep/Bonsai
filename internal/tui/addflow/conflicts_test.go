package addflow

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// newTestConflicts builds a ConflictsStage over a synthetic WriteResult with
// three conflict files spanning different categories. Used as the common
// fixture for every test in this file.
func newTestConflicts() *ConflictsStage {
	wr := &generate.WriteResult{}
	wr.Files = []generate.FileResult{
		{RelPath: "station/agent/Skills/foo.md", Action: generate.ActionConflict, Source: "skills/foo/foo.md"},
		{RelPath: "station/agent/Protocols/bar.md", Action: generate.ActionConflict, Source: "protocols/bar/bar.md"},
		{RelPath: "station/agent/Core/identity.md", Action: generate.ActionConflict, Source: "agents/tech-lead/core/identity.md.tmpl"},
	}
	return NewConflictsStage(initflow.StageContext{StartedAt: time.Now()}, wr)
}

func conflictsPressKey(s *ConflictsStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if cs, ok := m.(*ConflictsStage); ok {
		*s = *cs
	}
}

// TestConflicts_TabCount verifies the stage surfaces one tab per conflict
// file.
func TestConflicts_TabCount(t *testing.T) {
	s := newTestConflicts()
	if len(s.files) != 3 {
		t.Fatalf("tabs = %d, want 3 (one per conflict file)", len(s.files))
	}
}

// TestConflicts_DefaultIsKeep verifies every conflict file defaults to
// ConflictActionKeep — the destructive action (overwrite) must be opt-in.
func TestConflicts_DefaultIsKeep(t *testing.T) {
	s := newTestConflicts()
	for _, f := range s.files {
		act, ok := s.action[f.RelPath]
		if !ok {
			t.Fatalf("no action recorded for %q", f.RelPath)
		}
		if act != config.ConflictActionKeep {
			t.Fatalf("default action for %q = %v, want ConflictActionKeep", f.RelPath, act)
		}
	}
}

// TestConflicts_TabCycleWraps verifies ← / → wraps around the end of the
// tab list.
func TestConflicts_TabCycleWraps(t *testing.T) {
	s := newTestConflicts()
	if s.catIdx != 0 {
		t.Fatalf("initial catIdx = %d, want 0", s.catIdx)
	}
	// Right across the end: 0 → 1 → 2 → 0.
	for i := 0; i < len(s.files); i++ {
		conflictsPressKey(s, tea.KeyRight)
	}
	if s.catIdx != 0 {
		t.Fatalf("after full cycle, catIdx = %d, want 0 (wrap)", s.catIdx)
	}
	// Left from 0 wraps to the last tab.
	conflictsPressKey(s, tea.KeyLeft)
	if s.catIdx != len(s.files)-1 {
		t.Fatalf("after left from 0, catIdx = %d, want %d", s.catIdx, len(s.files)-1)
	}
}

// TestConflicts_RadioCycleClamps verifies ↑/↓ cycles the radio focus without
// wrapping.
func TestConflicts_RadioCycleClamps(t *testing.T) {
	s := newTestConflicts()
	key := s.currentKey()
	// Start at row 0 (Keep).
	if s.radio[key] != 0 {
		t.Fatalf("initial radio = %d, want 0", s.radio[key])
	}
	// Up from 0 clamps.
	conflictsPressKey(s, tea.KeyUp)
	if s.radio[key] != 0 {
		t.Fatalf("up from 0 radio = %d, want 0", s.radio[key])
	}
	// Down → 1 (Overwrite) → 2 (Backup) → 2 (clamp).
	conflictsPressKey(s, tea.KeyDown)
	if s.radio[key] != 1 || s.action[key] != config.ConflictActionOverwrite {
		t.Fatalf("after 1x down radio=%d action=%v, want 1 + Overwrite", s.radio[key], s.action[key])
	}
	conflictsPressKey(s, tea.KeyDown)
	if s.radio[key] != 2 || s.action[key] != config.ConflictActionBackup {
		t.Fatalf("after 2x down radio=%d action=%v, want 2 + Backup", s.radio[key], s.action[key])
	}
	conflictsPressKey(s, tea.KeyDown)
	if s.radio[key] != 2 {
		t.Fatalf("down from last radio=%d, want clamp at 2", s.radio[key])
	}
}

// TestConflicts_EnterAdvancesTabsThenCompletes verifies Enter cycles through
// the tabs in order and flips Done only on the last tab.
func TestConflicts_EnterAdvancesTabsThenCompletes(t *testing.T) {
	s := newTestConflicts()
	// First Enter: tab 0 → tab 1, not done.
	conflictsPressKey(s, tea.KeyEnter)
	if s.catIdx != 1 {
		t.Fatalf("after 1st Enter catIdx = %d, want 1", s.catIdx)
	}
	if s.Done() {
		t.Fatal("Done flipped on intermediate Enter")
	}
	// Second Enter: tab 1 → tab 2, not done.
	conflictsPressKey(s, tea.KeyEnter)
	if s.catIdx != 2 || s.Done() {
		t.Fatalf("after 2nd Enter catIdx=%d done=%v, want 2 + !done", s.catIdx, s.Done())
	}
	// Third Enter (from last tab): stays at 2, flips Done.
	conflictsPressKey(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("3rd Enter on last tab should flip Done")
	}
}

// TestConflicts_ResultMapPopulated verifies Result returns a map with one
// entry per conflict file carrying the user's current pick.
func TestConflicts_ResultMapPopulated(t *testing.T) {
	s := newTestConflicts()
	// Pick Overwrite on tab 0 + Backup on tab 1.
	conflictsPressKey(s, tea.KeyDown) // tab 0 → Overwrite
	conflictsPressKey(s, tea.KeyRight)
	conflictsPressKey(s, tea.KeyDown) // tab 1 → Overwrite
	conflictsPressKey(s, tea.KeyDown) // tab 1 → Backup

	res, ok := s.Result().(map[string]config.ConflictAction)
	if !ok {
		t.Fatalf("Result type = %T, want map[string]config.ConflictAction", s.Result())
	}
	if len(res) != 3 {
		t.Fatalf("result size = %d, want 3", len(res))
	}
	if res[s.files[0].RelPath] != config.ConflictActionOverwrite {
		t.Fatalf("tab 0 action = %v, want Overwrite", res[s.files[0].RelPath])
	}
	if res[s.files[1].RelPath] != config.ConflictActionBackup {
		t.Fatalf("tab 1 action = %v, want Backup", res[s.files[1].RelPath])
	}
	if res[s.files[2].RelPath] != config.ConflictActionKeep {
		t.Fatalf("tab 2 action = %v, want Keep (untouched)", res[s.files[2].RelPath])
	}
}

// TestConflicts_EmptyWriteResultRendersGracefully verifies constructing over
// a WriteResult with no conflicts produces a usable stage that completes on
// Enter and returns an empty map.
func TestConflicts_EmptyWriteResultRendersGracefully(t *testing.T) {
	wr := &generate.WriteResult{}
	s := NewConflictsStage(initflow.StageContext{StartedAt: time.Now()}, wr)
	if len(s.files) != 0 {
		t.Fatalf("empty wr should produce 0 tabs; got %d", len(s.files))
	}
	// Render should not panic even with no tabs. View depends on SetSize.
	s.SetSize(100, 30)
	_ = s.View()
	// Enter completes.
	conflictsPressKey(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("Enter on empty conflicts should flip Done")
	}
	// Result is a non-nil empty map.
	res, ok := s.Result().(map[string]config.ConflictAction)
	if !ok {
		t.Fatalf("Result type = %T, want map[string]config.ConflictAction", s.Result())
	}
	if len(res) != 0 {
		t.Fatalf("empty stage result size = %d, want 0", len(res))
	}
}

// TestConflicts_ResetPreservesPicks verifies Reset clears done but keeps tab
// focus, radio focus, and per-file action picks.
func TestConflicts_ResetPreservesPicks(t *testing.T) {
	s := newTestConflicts()
	// Pick Backup on tab 0, advance to tab 1.
	conflictsPressKey(s, tea.KeyDown) // Overwrite
	conflictsPressKey(s, tea.KeyDown) // Backup
	conflictsPressKey(s, tea.KeyRight)
	// Complete + reset.
	s.MarkDone()
	s.Reset()
	if s.Done() {
		t.Fatal("Reset should clear Done")
	}
	if s.catIdx != 1 {
		t.Fatalf("Reset changed catIdx = %d, want 1", s.catIdx)
	}
	if s.action[s.files[0].RelPath] != config.ConflictActionBackup {
		t.Fatalf("Reset changed action = %v, want Backup", s.action[s.files[0].RelPath])
	}
}

// TestConflicts_RenderDoesNotPanic smokes View at a few dims to prove the
// layout helpers don't index out of range.
func TestConflicts_RenderDoesNotPanic(t *testing.T) {
	s := newTestConflicts()
	for _, dim := range []struct{ w, h int }{
		{80, 24},
		{120, 40},
		{70, 20}, // min floor
		{40, 10}, // below floor — should return the min-size floor view
	} {
		s.SetSize(dim.w, dim.h)
		_ = s.View()
	}
}
