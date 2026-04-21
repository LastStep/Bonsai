package initflow

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
)

func newTestPlanted(wr *generate.WriteResult) *PlantedStage {
	s := NewPlantedStage(StageContext{
		Version:      "test",
		ProjectDir:   "/tmp/planted",
		StationDir:   "station/",
		AgentDisplay: "Tech Lead",
		StartedAt:    time.Now().Add(-3 * time.Second),
	}, wr, PlantedSummary{
		Skills: 2, Workflows: 3, Protocols: 1, Sensors: 2, Routines: 1,
	})
	s.width = 120
	s.height = 40
	return s
}

// TestPlanted_EnterCompletes verifies ↵ flips done.
func TestPlanted_EnterCompletes(t *testing.T) {
	s := newTestPlanted(&generate.WriteResult{})
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	s = m.(*PlantedStage)
	if !s.done {
		t.Fatalf("done=false after Enter")
	}
}

// TestPlanted_QCompletes verifies q also exits.
func TestPlanted_QCompletes(t *testing.T) {
	s := newTestPlanted(&generate.WriteResult{})
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	s = m.(*PlantedStage)
	if !s.done {
		t.Fatalf("done=false after q")
	}
}

// TestPlanted_ResultNil verifies Result is nil (terminal stage).
func TestPlanted_ResultNil(t *testing.T) {
	s := newTestPlanted(&generate.WriteResult{})
	if s.Result() != nil {
		t.Fatalf("Result = %v, want nil", s.Result())
	}
}

// TestPlanted_TreeFromWriteResult verifies the rendered tree includes
// NEW'd files and excludes Skipped / Conflict entries.
func TestPlanted_TreeFromWriteResult(t *testing.T) {
	wr := &generate.WriteResult{
		Files: []generate.FileResult{
			{RelPath: "CLAUDE.md", Action: generate.ActionCreated},
			{RelPath: "station/agent/Core/identity.md", Action: generate.ActionCreated},
			{RelPath: "station/INDEX.md", Action: generate.ActionUpdated},
			{RelPath: "station/existing.md", Action: generate.ActionSkipped},  // omitted
			{RelPath: "station/conflict.md", Action: generate.ActionConflict}, // omitted
		},
	}
	s := newTestPlanted(wr)
	body := s.renderWrittenBlock()

	for _, want := range []string{"CLAUDE.md", "identity.md", "INDEX.md"} {
		if !strings.Contains(body, want) {
			t.Errorf("tree missing %q", want)
		}
	}
	for _, omit := range []string{"existing.md", "conflict.md"} {
		if strings.Contains(body, omit) {
			t.Errorf("tree should not contain %q", omit)
		}
	}
}

// TestPlanted_TreeStationNodeCurrent verifies buildPlantedTree marks the
// station directory with NodeCurrent status so it renders with the leaf
// border.
func TestPlanted_TreeStationNodeCurrent(t *testing.T) {
	wr := &generate.WriteResult{
		Files: []generate.FileResult{
			{RelPath: "station/INDEX.md", Action: generate.ActionCreated},
		},
	}
	tree := buildPlantedTree(wr, "/tmp/proj", "station/")
	found := false
	var walk func(nodes []tui.TreeNode)
	walk = func(nodes []tui.TreeNode) {
		for _, n := range nodes {
			if n.Name == "station" && n.Status == tui.NodeCurrent {
				found = true
			}
			walk(n.Children)
		}
	}
	walk(tree)
	if !found {
		t.Errorf("station dir not marked NodeCurrent in tree")
	}
}

// TestPlanted_UpdatedBadge verifies updated files carry the UPDATED note.
func TestPlanted_UpdatedBadge(t *testing.T) {
	wr := &generate.WriteResult{
		Files: []generate.FileResult{
			{RelPath: "station/INDEX.md", Action: generate.ActionUpdated},
		},
	}
	tree := buildPlantedTree(wr, "/tmp/proj", "station/")
	var got string
	var walk func(nodes []tui.TreeNode)
	walk = func(nodes []tui.TreeNode) {
		for _, n := range nodes {
			if n.Name == "INDEX.md" {
				got = n.Note
			}
			walk(n.Children)
		}
	}
	walk(tree)
	if got != "UPDATED" {
		t.Errorf("INDEX.md note = %q, want UPDATED", got)
	}
}

// TestPlanted_ResponsiveStacked verifies the stacked layout at <100 cols
// puts the WRITTEN block before SUMMARY linearly in the rendered body.
func TestPlanted_ResponsiveStacked(t *testing.T) {
	s := newTestPlanted(&generate.WriteResult{
		Files: []generate.FileResult{
			{RelPath: "CLAUDE.md", Action: generate.ActionCreated},
		},
	})
	s.width = 80
	body := s.View()
	wIdx := strings.Index(body, "WRITTEN")
	sIdx := strings.Index(body, "SUMMARY")
	if wIdx < 0 || sIdx < 0 {
		t.Fatalf("missing WRITTEN/SUMMARY markers")
	}
	if wIdx >= sIdx {
		t.Errorf("narrow-width: WRITTEN (idx %d) should precede SUMMARY (idx %d)", wIdx, sIdx)
	}
}

// TestPlanted_MinSizeFloor verifies <floor dimensions show the floor panel.
func TestPlanted_MinSizeFloor(t *testing.T) {
	s := newTestPlanted(&generate.WriteResult{})
	s.width = 60
	s.height = 16
	if !strings.Contains(s.View(), "please enlarge") {
		t.Errorf("min-size render missing floor panel")
	}
}

// TestPlanted_ElapsedRendered verifies ELAPSED row is present and formatted.
func TestPlanted_ElapsedRendered(t *testing.T) {
	s := newTestPlanted(&generate.WriteResult{})
	body := s.View()
	if !strings.Contains(body, "ELAPSED") {
		t.Errorf("body missing ELAPSED label")
	}
}

// TestPlanted_SummaryAbilitiesTotal verifies the wired-count summary line
// sums the per-category counts correctly.
func TestPlanted_SummaryAbilitiesTotal(t *testing.T) {
	s := newTestPlanted(&generate.WriteResult{})
	body := s.renderSummaryBlock()
	// Totals: 2+3+1+2+1 = 9
	if !strings.Contains(body, "9 wired") {
		t.Errorf("summary missing '9 wired' total; got:\n%s", body)
	}
}
