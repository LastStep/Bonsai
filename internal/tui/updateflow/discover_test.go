package updateflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// minimalCatalog returns a Catalog with a single agent def so tests can
// resolve agent display names without embedding the full fs.FS.
func minimalCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	// Use the embedded catalog from the root package via a tiny init-style
	// loader — tests live inside the module so we can just ReadFile from
	// the repo checkout. Fall back to a hand-built Catalog if necessary.
	c, err := catalog.New(os.DirFS("../../../catalog"))
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	return c
}

// writeCustomSkill writes a custom skill file with well-formed frontmatter
// into the given agent's workspace. Returns the absolute path to the file.
func writeCustomSkill(t *testing.T, root, workspace, name, desc string) string {
	t.Helper()
	dir := filepath.Join(root, workspace, "agent", "Skills")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\ndescription: " + desc + "\n---\n\n# " + name + "\n"
	path := filepath.Join(dir, name+".md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// writeBadSkill writes a custom skill with missing frontmatter so the
// scan reports an Error on the returned DiscoveredFile.
func writeBadSkill(t *testing.T, root, workspace, name string) string {
	t.Helper()
	dir := filepath.Join(root, workspace, "agent", "Skills")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# no frontmatter\n\njust a heading.\n"
	path := filepath.Join(dir, name+".md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// newFixtureCfg returns a tiny project config with one agent. The
// workspace is relative to the temp root so ScanCustomFiles can walk it.
func newFixtureCfg(workspace string) *config.ProjectConfig {
	return &config.ProjectConfig{
		ProjectName: "fixture",
		DocsPath:    workspace,
		Agents: map[string]*config.InstalledAgent{
			"tech-lead": {
				AgentType: "tech-lead",
				Workspace: workspace,
			},
		},
	}
}

// TestDiscover_InvalidFilesSurface — a file with bad frontmatter shows
// up in the DiscoveredFile.Invalid slice (NOT Valid) and the stage's
// renderWarnings() output mentions the file's relpath.
func TestDiscover_InvalidFilesSurface(t *testing.T) {
	root := t.TempDir()
	cfg := newFixtureCfg("station/")
	cat := minimalCatalog(t)
	lock := config.NewLockFile()

	_ = writeBadSkill(t, root, "station/", "broken-skill")

	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewDiscoverStage(ctx, root, cfg, cat, lock)

	if len(s.Discoveries()) != 1 {
		t.Fatalf("discoveries len = %d, want 1", len(s.Discoveries()))
	}
	d := s.Discoveries()[0]
	if len(d.Invalid) != 1 {
		t.Fatalf("invalid len = %d, want 1", len(d.Invalid))
	}
	if len(d.Valid) != 0 {
		t.Fatalf("valid len = %d, want 0", len(d.Valid))
	}

	// Render should not crash and the warnings block should contain
	// the relpath of the bad file.
	s.SetSize(100, 30)
	view := s.View()
	if !strings.Contains(view, "broken-skill") {
		t.Fatalf("View should surface broken-skill relpath; got:\n%s", view)
	}
	if !strings.Contains(view, "WARNINGS") {
		t.Fatalf("View should surface WARNINGS header; got:\n%s", view)
	}
}

// TestDiscover_EmptyWorkspaceRendersGracefully — a workspace with no
// custom files produces an empty discoveries slice and the stage
// renders a "nothing to promote" placeholder.
func TestDiscover_EmptyWorkspaceRendersGracefully(t *testing.T) {
	root := t.TempDir()
	cfg := newFixtureCfg("station/")
	cat := minimalCatalog(t)
	lock := config.NewLockFile()

	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewDiscoverStage(ctx, root, cfg, cat, lock)

	if len(s.Discoveries()) != 0 {
		t.Fatalf("empty workspace should produce 0 discoveries; got %d", len(s.Discoveries()))
	}
	if s.HasValidDiscoveries() {
		t.Fatal("HasValidDiscoveries should be false")
	}
	s.SetSize(100, 30)
	_ = s.View() // must not panic
}

// TestDiscover_EnterCompletes — Enter marks the stage done.
func TestDiscover_EnterCompletes(t *testing.T) {
	root := t.TempDir()
	cfg := newFixtureCfg("station/")
	cat := minimalCatalog(t)
	lock := config.NewLockFile()

	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewDiscoverStage(ctx, root, cfg, cat, lock)
	s.SetSize(100, 30)

	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ds, ok := m.(*DiscoverStage); ok {
		*s = *ds
	}
	if !s.Done() {
		t.Fatal("Enter should flip Done")
	}
}

// TestDiscover_ValidDiscoveryFlagged — a file with well-formed
// frontmatter shows up in Valid and HasValidDiscoveries returns true.
func TestDiscover_ValidDiscoveryFlagged(t *testing.T) {
	root := t.TempDir()
	cfg := newFixtureCfg("station/")
	cat := minimalCatalog(t)
	lock := config.NewLockFile()

	_ = writeCustomSkill(t, root, "station/", "my-custom", "does something useful")

	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewDiscoverStage(ctx, root, cfg, cat, lock)

	if !s.HasValidDiscoveries() {
		t.Fatal("HasValidDiscoveries should be true with one valid file")
	}
	if len(s.Discoveries()) != 1 {
		t.Fatalf("discoveries len = %d, want 1", len(s.Discoveries()))
	}
	if got := len(s.Discoveries()[0].Valid); got != 1 {
		t.Fatalf("valid len = %d, want 1", got)
	}
}
