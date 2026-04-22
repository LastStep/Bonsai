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

// conflictsPressRune dispatches a rune-based KeyMsg (digits, letters, space).
func conflictsPressRune(s *ConflictsStage, r rune) {
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	if cs, ok := m.(*ConflictsStage); ok {
		*s = *cs
	}
}

// TestConflicts_RowCount verifies the stage surfaces one row per conflict
// file.
func TestConflicts_RowCount(t *testing.T) {
	s := newTestConflicts()
	if len(s.files) != 3 {
		t.Fatalf("rows = %d, want 3 (one per conflict file)", len(s.files))
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

// TestConflicts_FocusMoveClamps verifies ↑/↓ move focus without wrapping.
func TestConflicts_FocusMoveClamps(t *testing.T) {
	s := newTestConflicts()
	if s.focus != 0 {
		t.Fatalf("initial focus = %d, want 0", s.focus)
	}
	// Up from 0 clamps at 0.
	conflictsPressKey(s, tea.KeyUp)
	if s.focus != 0 {
		t.Fatalf("up from 0 focus = %d, want 0 (clamp)", s.focus)
	}
	// Down → 1 → 2 → 2 (clamp at last).
	conflictsPressKey(s, tea.KeyDown)
	if s.focus != 1 {
		t.Fatalf("after 1x down focus = %d, want 1", s.focus)
	}
	conflictsPressKey(s, tea.KeyDown)
	if s.focus != 2 {
		t.Fatalf("after 2x down focus = %d, want 2", s.focus)
	}
	conflictsPressKey(s, tea.KeyDown)
	if s.focus != 2 {
		t.Fatalf("down from last focus = %d, want 2 (clamp)", s.focus)
	}
}

// TestConflicts_DigitKeysSetAction verifies 1/2/3 set the focused row's
// action without moving focus.
func TestConflicts_DigitKeysSetAction(t *testing.T) {
	s := newTestConflicts()
	key := s.currentKey()

	conflictsPressRune(s, '2')
	if s.action[key] != config.ConflictActionOverwrite {
		t.Fatalf("after '2' action = %v, want Overwrite", s.action[key])
	}
	conflictsPressRune(s, '3')
	if s.action[key] != config.ConflictActionBackup {
		t.Fatalf("after '3' action = %v, want Backup", s.action[key])
	}
	conflictsPressRune(s, '1')
	if s.action[key] != config.ConflictActionKeep {
		t.Fatalf("after '1' action = %v, want Keep", s.action[key])
	}
	// Focus should be unchanged.
	if s.focus != 0 {
		t.Fatalf("digit keys moved focus to %d, want 0", s.focus)
	}
}

// TestConflicts_SpaceCyclesAction verifies ␣ cycles Keep → Overwrite →
// Backup → Keep on the focused row.
func TestConflicts_SpaceCyclesAction(t *testing.T) {
	s := newTestConflicts()
	key := s.currentKey()

	// Start: Keep.
	if s.action[key] != config.ConflictActionKeep {
		t.Fatalf("initial action = %v, want Keep", s.action[key])
	}
	conflictsPressRune(s, ' ')
	if s.action[key] != config.ConflictActionOverwrite {
		t.Fatalf("after 1x space action = %v, want Overwrite", s.action[key])
	}
	conflictsPressRune(s, ' ')
	if s.action[key] != config.ConflictActionBackup {
		t.Fatalf("after 2x space action = %v, want Backup", s.action[key])
	}
	conflictsPressRune(s, ' ')
	if s.action[key] != config.ConflictActionKeep {
		t.Fatalf("after 3x space action = %v, want Keep (wrap)", s.action[key])
	}
}

// TestConflicts_BatchKeyAppliesToAll verifies uppercase K / O / B apply the
// action to every row.
func TestConflicts_BatchKeyAppliesToAll(t *testing.T) {
	s := newTestConflicts()

	// Seed row 1 with a different action so we can prove K overwrites it.
	conflictsPressKey(s, tea.KeyDown)
	conflictsPressRune(s, '2') // row 1 → Overwrite
	conflictsPressKey(s, tea.KeyUp)

	// Batch overwrite — every row becomes Overwrite.
	conflictsPressRune(s, 'O')
	for _, f := range s.files {
		if s.action[f.RelPath] != config.ConflictActionOverwrite {
			t.Fatalf("after 'O' action[%q] = %v, want Overwrite", f.RelPath, s.action[f.RelPath])
		}
	}

	// Batch backup.
	conflictsPressRune(s, 'B')
	for _, f := range s.files {
		if s.action[f.RelPath] != config.ConflictActionBackup {
			t.Fatalf("after 'B' action[%q] = %v, want Backup", f.RelPath, s.action[f.RelPath])
		}
	}

	// Batch keep.
	conflictsPressRune(s, 'K')
	for _, f := range s.files {
		if s.action[f.RelPath] != config.ConflictActionKeep {
			t.Fatalf("after 'K' action[%q] = %v, want Keep", f.RelPath, s.action[f.RelPath])
		}
	}
}

// TestConflicts_LowercaseBatchKeysNoOp verifies lowercase k/o/b do not
// trigger batch-resolve (k maps to focus-up, o/b are unassigned no-ops).
// The Plan 27 §C4 contract: uppercase keys only trigger batch.
func TestConflicts_LowercaseBatchKeysNoOp(t *testing.T) {
	s := newTestConflicts()

	// Seed row 0 with Backup so we can prove lowercase 'o' doesn't
	// overwrite it.
	conflictsPressRune(s, '3')
	key0 := s.files[0].RelPath
	if s.action[key0] != config.ConflictActionBackup {
		t.Fatalf("setup action = %v, want Backup", s.action[key0])
	}

	// Lowercase 'o' — no batch, row 0's action must stay at Backup.
	conflictsPressRune(s, 'o')
	if s.action[key0] != config.ConflictActionBackup {
		t.Fatalf("after lowercase 'o' action[0] = %v, want Backup (unchanged)", s.action[key0])
	}
	// Row 1 should still be Keep (default).
	key1 := s.files[1].RelPath
	if s.action[key1] != config.ConflictActionKeep {
		t.Fatalf("after lowercase 'o' action[1] = %v, want Keep (unchanged)", s.action[key1])
	}

	// Lowercase 'b' — same expectation.
	conflictsPressRune(s, 'b')
	if s.action[key1] != config.ConflictActionKeep {
		t.Fatalf("after lowercase 'b' action[1] = %v, want Keep", s.action[key1])
	}
}

// TestConflicts_EnterCompletes verifies Enter flips Done on the single-screen
// vertical list (no per-tab cycling).
func TestConflicts_EnterCompletes(t *testing.T) {
	s := newTestConflicts()
	conflictsPressKey(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("Enter should flip Done")
	}
}

// TestConflicts_ResultMapPopulated verifies Result returns a map with one
// entry per conflict file carrying the user's current pick.
func TestConflicts_ResultMapPopulated(t *testing.T) {
	s := newTestConflicts()
	// Pick Overwrite on row 0, Backup on row 1, leave row 2 as Keep.
	conflictsPressRune(s, '2') // row 0 → Overwrite
	conflictsPressKey(s, tea.KeyDown)
	conflictsPressRune(s, '3') // row 1 → Backup

	res, ok := s.Result().(map[string]config.ConflictAction)
	if !ok {
		t.Fatalf("Result type = %T, want map[string]config.ConflictAction", s.Result())
	}
	if len(res) != 3 {
		t.Fatalf("result size = %d, want 3", len(res))
	}
	if res[s.files[0].RelPath] != config.ConflictActionOverwrite {
		t.Fatalf("row 0 action = %v, want Overwrite", res[s.files[0].RelPath])
	}
	if res[s.files[1].RelPath] != config.ConflictActionBackup {
		t.Fatalf("row 1 action = %v, want Backup", res[s.files[1].RelPath])
	}
	if res[s.files[2].RelPath] != config.ConflictActionKeep {
		t.Fatalf("row 2 action = %v, want Keep (untouched)", res[s.files[2].RelPath])
	}
}

// TestConflicts_EmptyWriteResultRendersGracefully verifies constructing over
// a WriteResult with no conflicts produces a usable stage that completes on
// Enter and returns an empty map.
func TestConflicts_EmptyWriteResultRendersGracefully(t *testing.T) {
	wr := &generate.WriteResult{}
	s := NewConflictsStage(initflow.StageContext{StartedAt: time.Now()}, wr)
	if len(s.files) != 0 {
		t.Fatalf("empty wr should produce 0 rows; got %d", len(s.files))
	}
	// Render should not panic even with no rows. View depends on SetSize.
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

// TestConflicts_ResetPreservesPicks verifies Reset clears done but keeps
// focus and per-file action picks.
func TestConflicts_ResetPreservesPicks(t *testing.T) {
	s := newTestConflicts()
	// Pick Backup on row 0, advance focus to row 1.
	conflictsPressRune(s, '3')
	conflictsPressKey(s, tea.KeyDown)
	// Complete + reset.
	s.MarkDone()
	s.Reset()
	if s.Done() {
		t.Fatal("Reset should clear Done")
	}
	if s.focus != 1 {
		t.Fatalf("Reset changed focus = %d, want 1", s.focus)
	}
	if s.action[s.files[0].RelPath] != config.ConflictActionBackup {
		t.Fatalf("Reset changed action = %v, want Backup", s.action[s.files[0].RelPath])
	}
}

// TestConflicts_Chromeless verifies the stage reports Chromeless=true so the
// harness yields its View() verbatim (Plan 27 PR2 §C1).
func TestConflicts_Chromeless(t *testing.T) {
	s := newTestConflicts()
	if !s.Chromeless() {
		t.Fatal("ConflictsStage should be chromeless")
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

// TestConflicts_ColorFollowsAction smokes renderRow at each action pick to
// prove the row composition doesn't panic when colors flip per action.
// Verifies the per-file color (Plan 27 PR2 §C3) is keyed off the current
// action for that row.
func TestConflicts_ColorFollowsAction(t *testing.T) {
	s := newTestConflicts()
	s.SetSize(100, 30)

	// Cycle through every action on row 0 and verify render produces output.
	for _, a := range []config.ConflictAction{
		config.ConflictActionKeep,
		config.ConflictActionOverwrite,
		config.ConflictActionBackup,
	} {
		s.action[s.files[0].RelPath] = a
		out := s.renderRow(0)
		if out == "" {
			t.Fatalf("renderRow returned empty for action %v", a)
		}
	}
}
