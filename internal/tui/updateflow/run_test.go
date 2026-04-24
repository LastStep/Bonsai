package updateflow

import (
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
)

// TestAppendUnique covers the small helper used by applyCustomFileSelection
// to guard against duplicate entries on repeated `bonsai update` runs.
// Ported from cmd/update_test.go (pre-Plan 31 Phase F) when the helpers
// moved into internal/tui/updateflow/run.go.
func TestAppendUnique(t *testing.T) {
	t.Run("appends when absent", func(t *testing.T) {
		got := appendUnique([]string{"a", "b"}, "c")
		want := []string{"a", "b", "c"}
		if !equalStringSlice(got, want) {
			t.Errorf("appendUnique = %v, want %v", got, want)
		}
	})

	t.Run("skips when present", func(t *testing.T) {
		got := appendUnique([]string{"a", "b", "c"}, "b")
		want := []string{"a", "b", "c"}
		if !equalStringSlice(got, want) {
			t.Errorf("appendUnique = %v, want %v", got, want)
		}
	})

	t.Run("appends to empty slice", func(t *testing.T) {
		got := appendUnique(nil, "a")
		want := []string{"a"}
		if !equalStringSlice(got, want) {
			t.Errorf("appendUnique = %v, want %v", got, want)
		}
	})
}

// TestApplyCustomFileSelectionDedupes verifies that re-running
// applyCustomFileSelection with the same selections does not accumulate
// duplicate entries in the installed agent's ability lists. Ported from
// cmd/update_test.go (pre-Plan 31 Phase F).
func TestApplyCustomFileSelectionDedupes(t *testing.T) {
	installed := &config.InstalledAgent{
		AgentType: "test-agent",
		Workspace: ".",
		Skills:    []string{},
		Workflows: []string{},
	}

	valid := []generate.DiscoveredFile{
		{Name: "custom-skill", Type: "skill", RelPath: "agent/Skills/custom-skill.md", Meta: &config.CustomItemMeta{Description: "custom skill"}},
		{Name: "custom-workflow", Type: "workflow", RelPath: "agent/Workflows/custom-workflow.md", Meta: &config.CustomItemMeta{Description: "custom workflow"}},
	}
	selected := []string{"skill:custom-skill", "workflow:custom-workflow"}

	lock := config.NewLockFile()
	cwd := t.TempDir()

	// First call — both get added.
	_ = applyCustomFileSelection(installed, valid, selected, lock, cwd)
	// Second call — same selection; neither should duplicate.
	_ = applyCustomFileSelection(installed, valid, selected, lock, cwd)

	if len(installed.Skills) != 1 {
		t.Errorf("Skills len = %d, want 1 (%v)", len(installed.Skills), installed.Skills)
	}
	if len(installed.Workflows) != 1 {
		t.Errorf("Workflows len = %d, want 1 (%v)", len(installed.Workflows), installed.Workflows)
	}
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
