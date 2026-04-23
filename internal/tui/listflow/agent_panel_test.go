package listflow

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/config"
)

// newAgent is a minimal InstalledAgent factory used across tests. Only
// fields the panel actually reads are populated — empty ability slices
// stay empty so pairs assertions stay tight.
func newAgent(workspace string) *config.InstalledAgent {
	return &config.InstalledAgent{
		AgentType: "tech-lead",
		Workspace: workspace,
	}
}

// TestRenderAgentPanel_MissingWorkspaceShowsHint covers case (a) from the
// plan: a config whose workspace path does not exist on disk surfaces the
// D3 CTA — Workspace missing — run: bonsai update — and no tree.
func TestRenderAgentPanel_MissingWorkspaceShowsHint(t *testing.T) {
	projectDir := t.TempDir()
	agent := newAgent("station") // no dir created

	out := RenderAgentPanel("tech-lead", agent, nil, projectDir, 120)

	if !strings.Contains(out, "Workspace missing — run: bonsai update") {
		t.Fatalf("expected D3 hint in output, got:\n%s", out)
	}
	// No tree markers when the dir doesn't exist.
	if strings.Contains(out, "├─") || strings.Contains(out, "└─") {
		t.Fatalf("did not expect tree glyphs, got:\n%s", out)
	}
}

// TestRenderAgentPanel_EmptyWorkspaceShowsMarker covers case (b): the
// workspace dir exists but is empty (or contains only hidden entries).
// The tree should render with a single "(empty)" leaf.
func TestRenderAgentPanel_EmptyWorkspaceShowsMarker(t *testing.T) {
	projectDir := t.TempDir()
	ws := filepath.Join(projectDir, "station")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Add a hidden file — it should be filtered out and the dir treated as empty.
	if err := os.WriteFile(filepath.Join(ws, ".hidden"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	agent := newAgent("station")
	out := RenderAgentPanel("tech-lead", agent, nil, projectDir, 120)

	if !strings.Contains(out, "(empty)") {
		t.Fatalf("expected (empty) marker in tree, got:\n%s", out)
	}
}

// TestRenderAgentPanel_UnderCapShowsAllFiles covers case (c): a workspace
// with 10 files renders all 10 in the tree without a truncation row.
func TestRenderAgentPanel_UnderCapShowsAllFiles(t *testing.T) {
	projectDir := t.TempDir()
	ws := filepath.Join(projectDir, "station")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for i := 0; i < 10; i++ {
		name := "file" + strconv.Itoa(i) + ".md"
		if err := os.WriteFile(filepath.Join(ws, name), []byte("x"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	agent := newAgent("station")
	out := RenderAgentPanel("tech-lead", agent, nil, projectDir, 120)

	for i := 0; i < 10; i++ {
		name := "file" + strconv.Itoa(i) + ".md"
		if !strings.Contains(out, name) {
			t.Fatalf("expected %s in output, got:\n%s", name, out)
		}
	}
	if strings.Contains(out, "more)") {
		t.Fatalf("did not expect truncation row under cap, got:\n%s", out)
	}
}

// TestRenderAgentPanel_OverCapTruncatesWithSummary covers case (d): 60
// files → first 50 entries + synthetic "... (10 more)" row.
func TestRenderAgentPanel_OverCapTruncatesWithSummary(t *testing.T) {
	projectDir := t.TempDir()
	ws := filepath.Join(projectDir, "station")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for i := 0; i < 60; i++ {
		// Zero-pad so the sort order matches numeric order and the
		// "(10 more)" assertion is deterministic.
		name := "file" + fmtInt(i, 2) + ".md"
		if err := os.WriteFile(filepath.Join(ws, name), []byte("x"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	agent := newAgent("station")
	out := RenderAgentPanel("tech-lead", agent, nil, projectDir, 120)

	if !strings.Contains(out, "(10 more)") {
		t.Fatalf("expected '... (10 more)' truncation row, got:\n%s", out)
	}
	// The 51st file (file50.md, sorted) should NOT appear in output.
	if strings.Contains(out, "file50.md") {
		t.Fatalf("did not expect file50 when cap=50; got:\n%s", out)
	}
	// The 50th file (file49.md) should appear.
	if !strings.Contains(out, "file49.md") {
		t.Fatalf("expected file49.md in truncated output, got:\n%s", out)
	}
}

// TestRenderAgentPanel_EscapePathRefused covers case (e): a workspace
// path containing ".." is refused with a warning line and no walk.
func TestRenderAgentPanel_EscapePathRefused(t *testing.T) {
	projectDir := t.TempDir()
	agent := newAgent("../outside")

	out := RenderAgentPanel("tech-lead", agent, nil, projectDir, 120)

	if !strings.Contains(out, "escapes project root") {
		t.Fatalf("expected escape-warning line, got:\n%s", out)
	}
	if strings.Contains(out, "├─") || strings.Contains(out, "└─") {
		t.Fatalf("did not expect tree glyphs for escape path, got:\n%s", out)
	}
}

// TestRenderAgentPanel_SymlinkLoopTerminates covers case (f): a symlink
// cycle must not produce infinite recursion. A 2s deadline is the
// smoke-test budget — the SkipDir-on-symlink policy should return in
// milliseconds.
func TestRenderAgentPanel_SymlinkLoopTerminates(t *testing.T) {
	if _, err := os.Lstat("/"); err != nil {
		t.Skip("filesystem lstat unavailable")
	}
	projectDir := t.TempDir()
	ws := filepath.Join(projectDir, "station")
	if err := os.MkdirAll(filepath.Join(ws, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create a symlink loop: ws/sub/loop -> ws
	if err := os.Symlink(ws, filepath.Join(ws, "sub", "loop")); err != nil {
		t.Skipf("symlink unsupported on this platform: %v", err)
	}
	// Add a regular file so we have something non-symlink to render.
	if err := os.WriteFile(filepath.Join(ws, "real.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	done := make(chan string, 1)
	go func() {
		agent := newAgent("station")
		done <- RenderAgentPanel("tech-lead", agent, nil, projectDir, 120)
	}()
	select {
	case out := <-done:
		if !strings.Contains(out, "real.md") {
			t.Fatalf("expected real.md to render alongside the symlink, got:\n%s", out)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("symlink loop did not terminate within 2s — SkipDir policy regressed")
	}
}

// TestRenderAgentPanel_PairsIncludeInstalledAbilities covers the panel
// field list: a populated agent renders Skills/Workflows/Protocols/
// Sensors/Routines rows. Workspace row is always present.
func TestRenderAgentPanel_PairsIncludeInstalledAbilities(t *testing.T) {
	projectDir := t.TempDir()
	ws := filepath.Join(projectDir, "station")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	agent := &config.InstalledAgent{
		AgentType: "tech-lead",
		Workspace: "station",
		Skills:    []string{"planning-template"},
		Workflows: []string{"code-review"},
		Protocols: []string{"memory"},
		Sensors:   []string{"status-bar"},
		Routines:  []string{"backlog-hygiene"},
	}

	out := RenderAgentPanel("tech-lead", agent, nil, projectDir, 120)

	for _, want := range []string{
		"Workspace",
		"Skills", "Planning Template",
		"Workflows", "Code Review",
		"Protocols", "Memory",
		"Sensors", "Status Bar",
		"Routines", "Backlog Hygiene",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in panel output, got:\n%s", want, out)
		}
	}
}

// fmtInt zero-pads n to at least width digits. Avoids pulling strconv in
// again (already imported above, but we keep this local so if the import
// slims later the test still compiles).
func fmtInt(n, width int) string {
	s := strconv.Itoa(n)
	for len(s) < width {
		s = "0" + s
	}
	return s
}
