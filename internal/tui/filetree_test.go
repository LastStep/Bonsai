package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

// ansiVisible strips ANSI escape sequences so string assertions can focus on
// the rendered glyph layer without caring about styling. Uses the same ansi
// helper used by the rest of the TUI package.
func ansiVisible(s string) string {
	return ansi.Strip(s)
}

// forceColorProfile turns on a truecolor profile for the duration of a test,
// then restores the prior profile. Useful for asserting ANSI escape presence
// (NodeCurrent background) which otherwise gets stripped under the ASCII
// profile installed by init() when stdout isn't a TTY.
func forceColorProfile(t *testing.T, p termenv.Profile) {
	t.Helper()
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(p)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(prev)
	})
}

func TestRenderFileTree_FlatBranchGlyphs(t *testing.T) {
	nodes := []TreeNode{
		{Name: "one.md", Kind: NodeFile},
		{Name: "two.md", Kind: NodeFile},
		{Name: "three.md", Kind: NodeFile},
	}
	out := ansiVisible(RenderFileTree(nodes, FileTreeOpts{Dense: true, MaxWidth: 60}))
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), out)
	}
	if !strings.HasPrefix(lines[0], "├─ ") {
		t.Errorf("line 0 should start with intermediate branch glyph; got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "├─ ") {
		t.Errorf("line 1 should start with intermediate branch glyph; got %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "└─ ") {
		t.Errorf("final line should start with last-branch glyph; got %q", lines[2])
	}
}

func TestRenderFileTree_NestedContinuationPrefix(t *testing.T) {
	nodes := []TreeNode{
		{
			Name: "station", Kind: NodeDir,
			Children: []TreeNode{
				{Name: "alpha.md", Kind: NodeFile},
				{Name: "beta.md", Kind: NodeFile},
			},
		},
		{Name: "README.md", Kind: NodeFile},
	}
	out := ansiVisible(RenderFileTree(nodes, FileTreeOpts{Dense: true, MaxWidth: 60}))
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// Expect the station/ dir (non-last) to use │  continuation for its
	// children, since it is followed by README.md at the top level.
	want := []string{
		"├─ station/",
		"│  ├─ alpha.md",
		"│  └─ beta.md",
		"└─ README.md",
	}
	for i, w := range want {
		if i >= len(lines) {
			t.Fatalf("missing line %d; got:\n%s", i, out)
		}
		if !strings.HasPrefix(lines[i], w) {
			t.Errorf("line %d: want prefix %q, got %q", i, w, lines[i])
		}
	}
}

func TestRenderFileTree_NestedLastDirSpaceContinuation(t *testing.T) {
	// When the last top-level node is a directory, its grandchild continuation
	// prefix should be three spaces ("   ") instead of "│  ".
	nodes := []TreeNode{
		{Name: "README.md", Kind: NodeFile},
		{
			Name: "station", Kind: NodeDir,
			Children: []TreeNode{
				{Name: "alpha.md", Kind: NodeFile},
				{Name: "beta.md", Kind: NodeFile},
			},
		},
	}
	out := ansiVisible(RenderFileTree(nodes, FileTreeOpts{Dense: true, MaxWidth: 60}))
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")

	// "   ├─ alpha.md" and "   └─ beta.md" under a last-dir parent.
	if !strings.HasPrefix(lines[2], "   ├─ alpha.md") {
		t.Errorf("expected three-space continuation before alpha; got %q", lines[2])
	}
	if !strings.HasPrefix(lines[3], "   └─ beta.md") {
		t.Errorf("expected three-space continuation before beta; got %q", lines[3])
	}
}

func TestRenderFileTree_NewBadgeAndLeafName(t *testing.T) {
	forceColorProfile(t, termenv.TrueColor)
	nodes := []TreeNode{
		{Name: "log.md", Kind: NodeFile, Status: NodeNew, Note: "session rolling log"},
	}
	raw := RenderFileTree(nodes, FileTreeOpts{Dense: true, MaxWidth: 80})
	visible := ansiVisible(raw)

	if !strings.Contains(visible, "NEW") {
		t.Errorf("expected NEW badge in output; got %q", visible)
	}
	if !strings.Contains(visible, "session rolling log") {
		t.Errorf("expected note in output; got %q", visible)
	}
	// Badge ends exactly at MaxWidth (80), with a trailing space as part of
	// the 9-col block. The trimmed line length must equal width - trailing
	// spaces after the badge.
	line := strings.Split(strings.TrimRight(visible, "\n"), "\n")[0]
	if w := lipgloss.Width(line); w != 80 {
		t.Errorf("expected rendered line width %d (MaxWidth), got %d (line=%q)", 80, w, line)
	}
	// Leaf foreground color ANSI should appear at least once (name + badge).
	if !strings.Contains(raw, "\x1b[") {
		t.Errorf("expected ANSI escape sequences when truecolor profile is forced; got %q", raw)
	}
}

func TestRenderFileTree_RequiredBadge(t *testing.T) {
	nodes := []TreeNode{
		{Name: "CLAUDE.md", Kind: NodeFile, Status: NodeRequired},
	}
	out := ansiVisible(RenderFileTree(nodes, FileTreeOpts{Dense: true, MaxWidth: 60}))
	if !strings.Contains(out, "REQUIRED") {
		t.Errorf("expected REQUIRED badge in output; got %q", out)
	}
}

func TestRenderFileTree_CurrentBorderAndBackground(t *testing.T) {
	forceColorProfile(t, termenv.TrueColor)
	nodes := []TreeNode{
		{Name: "agent", Kind: NodeDir, Status: NodeCurrent, Children: []TreeNode{
			{Name: "identity.md", Kind: NodeFile},
		}},
	}
	raw := RenderFileTree(nodes, FileTreeOpts{Dense: true, MaxWidth: 60})
	visible := ansiVisible(raw)

	// Current node prepends "│ " (two visible cols) in Leaf.
	lines := strings.Split(strings.TrimRight(visible, "\n"), "\n")
	if !strings.HasPrefix(lines[0], "│ ") {
		t.Errorf("NodeCurrent should prepend a 2-col leaf border; got %q", lines[0])
	}
	// Background ANSI sequence must be present (SGR code for background).
	if !strings.Contains(raw, "\x1b[") {
		t.Fatalf("expected ANSI escapes under truecolor profile; got %q", raw)
	}
	// Background code for lipgloss is either "48;" (truecolor) or "48;5;"
	// (256-color). Either substring satisfies the assertion.
	if !strings.Contains(raw, "\x1b[48;") && !strings.Contains(raw, "48;5;") {
		t.Errorf("expected a background SGR code on NodeCurrent row; raw=%q", raw)
	}
}

func TestRenderFileTree_DenseHasNoBlankSeparators(t *testing.T) {
	nodes := []TreeNode{
		{Name: "a.md", Kind: NodeFile},
		{Name: "b.md", Kind: NodeFile},
		{Name: "c.md", Kind: NodeFile},
	}
	dense := ansiVisible(RenderFileTree(nodes, FileTreeOpts{Dense: true, MaxWidth: 60}))
	if strings.Contains(strings.TrimRight(dense, "\n"), "\n\n") {
		t.Errorf("Dense mode should not contain blank separator lines; got:\n%s", dense)
	}

	spaced := ansiVisible(RenderFileTree(nodes, FileTreeOpts{Dense: false, MaxWidth: 60}))
	if !strings.Contains(spaced, "\n\n") {
		t.Errorf("Non-dense mode should insert blank separator lines between siblings; got:\n%s", spaced)
	}
}

func TestRenderFileTree_MaxWidthTrimsNote(t *testing.T) {
	longNote := "this note is intentionally quite long and should be truncated with an ellipsis"
	nodes := []TreeNode{
		{Name: "f.md", Kind: NodeFile, Note: longNote},
	}
	out := ansiVisible(RenderFileTree(nodes, FileTreeOpts{Dense: true, MaxWidth: 40}))
	if !strings.Contains(out, "…") {
		t.Errorf("expected truncation ellipsis when MaxWidth forces trim; got %q", out)
	}
	line := strings.Split(strings.TrimRight(out, "\n"), "\n")[0]
	if w := lipgloss.Width(line); w > 40 {
		t.Errorf("line should not exceed MaxWidth=40; got width=%d line=%q", w, line)
	}
}

func TestRenderFileTree_RootLabel(t *testing.T) {
	root := &TreeNode{Name: "station", Kind: NodeDir, Note: "workspace root"}
	nodes := []TreeNode{
		{Name: "alpha.md", Kind: NodeFile},
		{Name: "beta.md", Kind: NodeFile},
	}
	out := ansiVisible(RenderFileTree(nodes, FileTreeOpts{Root: root, Dense: true, MaxWidth: 60}))
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")

	if len(lines) < 3 {
		t.Fatalf("expected at least root + 2 children; got %d:\n%s", len(lines), out)
	}
	// Root line first with station/ + note.
	if !strings.Contains(lines[0], "station/") || !strings.Contains(lines[0], "workspace root") {
		t.Errorf("root line missing name or note; got %q", lines[0])
	}
	// Children hang under the root with the │  continuation prefix.
	if !strings.HasPrefix(lines[1], "│  ├─ alpha.md") {
		t.Errorf("first child should hang under root with │  prefix; got %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "│  └─ beta.md") {
		t.Errorf("last child should hang under root with │  prefix; got %q", lines[2])
	}
}

// TestRenderFileTree_Demo prints a fully-populated tree so the rendered
// output can be eyeballed against the design reference. Not an assertion
// test — it is kept alongside the rule tests as a visual regression aid.
// Run with:  go test ./internal/tui/... -v -run TestRenderFileTree_Demo
func TestRenderFileTree_Demo(t *testing.T) {
	forceColorProfile(t, termenv.TrueColor)

	root := &TreeNode{
		Name: "~/code/voyager-api",
		Kind: NodeDir,
		Note: "workspace root",
	}
	nodes := []TreeNode{
		{Name: "CLAUDE.md", Kind: NodeFile, Status: NodeRequired, Note: "root-level agent directive"},
		{Name: ".bonsai.yaml", Kind: NodeFile, Status: NodeNew, Note: "project config"},
		{
			Name: "station", Kind: NodeDir, Status: NodeCurrent,
			Note: "agent workspace",
			Children: []TreeNode{
				{Name: "agents-index.md", Kind: NodeFile, Status: NodeRequired, Note: "directory of every agent"},
				{Name: "session-log.md", Kind: NodeFile, Status: NodeNew, Note: "rolling per-session log"},
				{
					Name: "protocols", Kind: NodeDir,
					Children: []TreeNode{
						{Name: "memory.md", Kind: NodeFile, Status: NodeNew},
						{Name: "security.md", Kind: NodeFile, Status: NodeNew},
					},
				},
			},
		},
		{Name: "readme.md", Kind: NodeFile, Status: NodeNew, Note: "starter README"},
	}

	spaced := RenderFileTree(nodes, FileTreeOpts{Root: root, MaxWidth: 80})
	dense := RenderFileTree(nodes, FileTreeOpts{Root: root, Dense: true, MaxWidth: 80})

	t.Log("\n── DEMO (spaced, MaxWidth=80) ──\n" + spaced)
	t.Log("\n── DEMO (dense,  MaxWidth=80) ──\n" + dense)
}
