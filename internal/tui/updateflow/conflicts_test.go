package updateflow

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// Mirror of addflow/conflicts_test.go — the stage body is an identical
// copy per Plan 31 F hard constraint #2 ("Copy conflicts.go verbatim"),
// so the behaviour under test is the same. These smoke tests guarantee
// the cross-package copy compiles and wires up correctly.

func newTestConflictsUF() *ConflictsStage {
	wr := &generate.WriteResult{}
	wr.Files = []generate.FileResult{
		{RelPath: "station/agent/Skills/foo.md", Action: generate.ActionConflict},
		{RelPath: "station/agent/Protocols/bar.md", Action: generate.ActionConflict},
	}
	return NewConflictsStage(initflow.StageContext{StartedAt: time.Now()}, wr)
}

func conflictsPressKeyUF(s *ConflictsStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if cs, ok := m.(*ConflictsStage); ok {
		*s = *cs
	}
}
func conflictsPressRuneUF(s *ConflictsStage, r rune) {
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	if cs, ok := m.(*ConflictsStage); ok {
		*s = *cs
	}
}

// TestConflictsUF_DefaultIsKeep — every row defaults to Keep (destructive
// action is opt-in).
func TestConflictsUF_DefaultIsKeep(t *testing.T) {
	s := newTestConflictsUF()
	for _, f := range s.files {
		if s.action[f.RelPath] != config.ConflictActionKeep {
			t.Fatalf("default action for %q = %v, want Keep", f.RelPath, s.action[f.RelPath])
		}
	}
}

// TestConflictsUF_DigitKeysSetAction — 1/2/3 set Keep/Overwrite/Backup.
func TestConflictsUF_DigitKeysSetAction(t *testing.T) {
	s := newTestConflictsUF()
	key := s.currentKey()
	conflictsPressRuneUF(s, '2')
	if s.action[key] != config.ConflictActionOverwrite {
		t.Fatalf("after '2' action = %v, want Overwrite", s.action[key])
	}
	conflictsPressRuneUF(s, '3')
	if s.action[key] != config.ConflictActionBackup {
		t.Fatalf("after '3' action = %v, want Backup", s.action[key])
	}
}

// TestConflictsUF_EnterCompletes — Enter flips Done.
func TestConflictsUF_EnterCompletes(t *testing.T) {
	s := newTestConflictsUF()
	conflictsPressKeyUF(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("Enter should flip Done")
	}
}

// TestConflictsUF_BatchKeys — K/O/B apply action to every row.
func TestConflictsUF_BatchKeys(t *testing.T) {
	s := newTestConflictsUF()
	conflictsPressRuneUF(s, 'O')
	for _, f := range s.files {
		if s.action[f.RelPath] != config.ConflictActionOverwrite {
			t.Fatalf("'O' batch action[%q] = %v, want Overwrite", f.RelPath, s.action[f.RelPath])
		}
	}
}

// TestConflictsUF_ResultMapShape — Result returns a map keyed by
// RelPath.
func TestConflictsUF_ResultMapShape(t *testing.T) {
	s := newTestConflictsUF()
	res, ok := s.Result().(map[string]config.ConflictAction)
	if !ok {
		t.Fatalf("Result type = %T, want map[string]config.ConflictAction", s.Result())
	}
	if len(res) != 2 {
		t.Fatalf("result size = %d, want 2", len(res))
	}
}
