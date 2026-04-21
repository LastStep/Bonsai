package initflow

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
)

// PlantedStage is the post-write celebration frame. Renders a file tree
// sourced from generate.WriteResult alongside a summary panel (agent +
// ability counts) and a three-step "next" callout.
//
// Result: nil (terminal stage — Done() flips on ↵/q and the harness exits).
type PlantedStage struct {
	Stage

	// Populated by the caller via the ctor or a post-hoc setter. Kept as
	// pointer so the stage can be built before generate runs and refresh
	// once the WriteResult is populated in-place.
	wr *generate.WriteResult

	// Summary inputs.
	skillCount    int
	workflowCount int
	protocolCount int
	sensorCount   int
	routineCount  int

	viewport Viewport
}

// PlantedSummary carries the ability counts rendered in the summary
// panel. Kept as a small record so the constructor stays legible — the
// caller populates it from BranchesResult + SoilStage.Result.
type PlantedSummary struct {
	Skills    int
	Workflows int
	Protocols int
	Sensors   int
	Routines  int
}

// NewPlantedStage constructs the terminal Planted stage at rail position 3
// (shares the Observe slot visually; see GenerateStage's note). wr may be
// nil at ctor time — the stage renders a "(nothing written)" placeholder
// if wr is still nil at render time.
func NewPlantedStage(ctx StageContext, wr *generate.WriteResult, summary PlantedSummary) *PlantedStage {
	label := StageLabels[3]
	base := NewStage(
		3,
		label,
		label.English,
		ctx.Version,
		ctx.ProjectDir,
		ctx.StationDir,
		ctx.AgentDisplay,
		ctx.StartedAt,
	)
	return &PlantedStage{
		Stage:         base,
		wr:            wr,
		skillCount:    summary.Skills,
		workflowCount: summary.Workflows,
		protocolCount: summary.Protocols,
		sensorCount:   summary.Sensors,
		routineCount:  summary.Routines,
	}
}

// Init implements tea.Model — no cursor / goroutine on entry.
func (s *PlantedStage) Init() tea.Cmd { return nil }

// Update handles ↵ / q as the terminal acknowledgement that advances Done.
func (s *PlantedStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
		s.height = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "enter", "q", "esc":
			s.done = true
			return s, nil
		}
	}
	return s, nil
}

// View composes the Planted stage body inside the shared frame.
func (s *PlantedStage) View() string {
	return s.renderFrame(s.renderBody(), s.keyHints())
}

// keyHints builds the footer key row.
func (s *PlantedStage) keyHints() []KeyHint {
	return []KeyHint{
		{Key: "↵", Desc: "exit"},
		{Key: "q", Desc: "quit"},
	}
}

// renderBody renders the full post-plant frame.
func (s *PlantedStage) renderBody() string {
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)

	created, updated, unchanged, skipped, conflicts := s.counts()
	total := created + updated + unchanged + skipped + conflicts

	// Hero block.
	var heroTitle string
	if s.ensoSafe {
		heroTitle = leaf.Render("生 · PLANTED")
	} else {
		heroTitle = leaf.Render("PLANTED")
	}
	project := filepath.Base(s.projectDir)
	heroSub := white.Render(project + " is ready.")
	heroStats := dim.Render(fmt.Sprintf(
		"%d files written · %d conflicts · lock synced",
		created+updated, conflicts,
	))
	_ = total

	divider := dim.Render(strings.Repeat("─", s.dividerWidth()))

	// Left — WRITTEN.
	writtenBlock := s.renderWrittenBlock()

	// Right — SUMMARY.
	summaryBlock := s.renderSummaryBlock()

	// Responsive: 2-column ≥100 cols; stacked otherwise.
	var grid string
	if s.width >= 100 {
		grid = joinHoriz(writtenBlock, summaryBlock, 4)
	} else {
		grid = writtenBlock + "\n\n" + summaryBlock
	}

	// Footer nextsteps + brand.
	nextBlock := s.renderNextSteps()

	elapsed := formatElapsed(time.Since(s.startedAt))
	heroLine := heroTitle + "   " + bark.Render("ELAPSED "+elapsed)

	body := strings.Join([]string{
		heroLine,
		heroSub,
		heroStats,
		"",
		divider,
		"",
		grid,
		"",
		divider,
		"",
		nextBlock,
	}, "\n")
	return centerBlock(body, s.width)
}

// counts summarises s.wr — returns zero counts when wr is nil.
func (s *PlantedStage) counts() (created, updated, unchanged, skipped, conflicts int) {
	if s.wr == nil {
		return
	}
	created, updated, unchanged, skipped, conflicts = s.wr.Summary()
	return
}

// dividerWidth returns the horizontal rule width — capped at 80 cells to
// keep wide terminals visually framed.
func (s *PlantedStage) dividerWidth() int {
	w := s.width - 4
	if w > 80 {
		w = 80
	}
	if w < 20 {
		w = 20
	}
	return w
}

// renderWrittenBlock renders the tree of written files. NodeStatus mapping:
//
//	ActionCreated              → NodeNew
//	ActionUpdated / ActionForced → NodeNew (plus "UPDATED" badge via Note)
//	ActionUnchanged            → NodeNormal
//	ActionSkipped / ActionConflict → omitted
func (s *PlantedStage) renderWrittenBlock() string {
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)

	header := leaf.Render(strings.Repeat("─", 3)) + " " +
		bark.Render("WRITTEN") + " "
	if s.ensoSafe {
		header += bark.Render("書 ")
	}
	header += dim.Render(strings.Repeat("─", 20))

	if s.wr == nil || len(s.wr.Files) == 0 {
		return header + "\n  " + dim.Render("(nothing written)")
	}

	tree := buildPlantedTree(s.wr, s.projectDir, s.stationDir)
	// Cap rendered width so this column doesn't hog the right side in the
	// 2-col layout.
	maxW := s.width/2 - 4
	if s.width < 100 {
		maxW = s.width - 4
	}
	if maxW < 30 {
		maxW = 30
	}
	rendered := tui.RenderFileTree(tree, tui.FileTreeOpts{
		Dense:    true,
		MaxWidth: maxW,
	})

	// Wrap in Viewport when the rendered tree exceeds a reasonable body
	// budget. Budget heuristic: body height - chrome - summary/next blocks.
	lines := strings.Split(rendered, "\n")
	budget := s.listHeight()
	if budget > 0 && len(lines) > budget {
		s.viewport.SetLines(lines)
		s.viewport.SetHeight(budget)
		return header + "\n" + s.viewport.View()
	}
	return header + "\n" + rendered
}

// listHeight returns the visible-row budget for the WRITTEN tree. Keeps
// the summary + next-steps blocks on screen even at 20-row terminals.
func (s *PlantedStage) listHeight() int {
	if s.height <= 0 {
		return 0
	}
	// chrome (10) + hero (4) + dividers+blanks (6) + next (6) + summary (~6) = 32
	h := s.height - 28
	if h < 5 {
		h = 5
	}
	return h
}

// renderSummaryBlock renders the right-hand summary panel.
func (s *PlantedStage) renderSummaryBlock() string {
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent)

	header := leaf.Render(strings.Repeat("─", 3)) + " " +
		bark.Render("SUMMARY") + " "
	if s.ensoSafe {
		header += bark.Render("概要 ")
	}
	header += dim.Render(strings.Repeat("─", 15))

	total := s.skillCount + s.workflowCount + s.protocolCount + s.sensorCount + s.routineCount

	const labelW = 11
	rows := []string{
		header,
		bark.Render(padRight("AGENT", labelW)) + white.Render(s.agentDisplay) +
			dim.Render(" → "+s.stationDir),
		bark.Render(padRight("ABILITIES", labelW)) + white.Render(fmt.Sprintf("%d wired", total)),
		bark.Render(padRight("···", labelW)) + dim.Render(fmt.Sprintf(
			"%d skills · %d flows · %d rules · %d sense · %d habit",
			s.skillCount, s.workflowCount, s.protocolCount, s.sensorCount, s.routineCount,
		)),
	}
	return strings.Join(rows, "\n")
}

// renderNextSteps renders the three-step "next" callout.
func (s *PlantedStage) renderNextSteps() string {
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent)

	header := leaf.Render(strings.Repeat("─", 3)) + " " +
		bark.Render("NEXT") + " "
	if s.ensoSafe {
		header += bark.Render("次へ ")
	}
	header += dim.Render(strings.Repeat("─", 20))

	numerals := []string{"一", "二", "三"}
	if !s.ensoSafe {
		numerals = []string{"1", "2", "3"}
	}

	steps := []struct {
		num, cmd, caption string
	}{
		{numerals[0], "$ claude", "open the workspace · say \"hi, get started\""},
		{numerals[1], "$ bonsai add", "add a code agent — backend, frontend, devops"},
		{numerals[2], "$ bonsai dashboard", "tend the garden — inspect + adjust abilities"},
	}
	lines := []string{header}
	for _, st := range steps {
		lines = append(lines, "  "+bark.Render(st.num)+"  "+white.Render(st.cmd))
		lines = append(lines, "     "+dim.Render(st.caption))
	}
	return strings.Join(lines, "\n")
}

// Result returns nil — Planted is the terminal stage.
func (s *PlantedStage) Result() any { return nil }

// Reset clears the completion flag. Preserves all counts + wr.
func (s *PlantedStage) Reset() tea.Cmd {
	s.done = false
	return nil
}

// buildPlantedTree turns a WriteResult into a tree of tui.TreeNode values
// rooted at the project dir. Only files with actions that reflect an
// actual write (Created / Updated / Forced / Unchanged) are included;
// Skipped and Conflict entries are omitted per plan.
//
// Directory nodes carry NodeCurrent when their path equals the agent's
// stationDir so the agent subtree renders with the green border + tint.
func buildPlantedTree(wr *generate.WriteResult, projectDir, stationDir string) []tui.TreeNode {
	// Collect kept files into a path-sorted list so sibling order is
	// deterministic across renders.
	type entry struct {
		path   string
		status tui.NodeStatus
		note   string
	}
	entries := make([]entry, 0, len(wr.Files))
	for _, f := range wr.Files {
		switch f.Action {
		case generate.ActionCreated:
			entries = append(entries, entry{path: f.RelPath, status: tui.NodeNew})
		case generate.ActionUpdated, generate.ActionForced:
			entries = append(entries, entry{path: f.RelPath, status: tui.NodeNew, note: "UPDATED"})
		case generate.ActionUnchanged:
			entries = append(entries, entry{path: f.RelPath, status: tui.NodeNormal})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })

	// Build trie of TreeNodes by walking each path's segments.
	type trieNode struct {
		name     string
		kind     tui.NodeKind
		status   tui.NodeStatus
		note     string
		children map[string]*trieNode
		order    []string // preserves insertion order
	}
	root := &trieNode{children: map[string]*trieNode{}}

	stationClean := strings.TrimSuffix(stationDir, "/")
	for _, e := range entries {
		parts := strings.Split(e.path, "/")
		cur := root
		for i, part := range parts {
			leaf := i == len(parts)-1
			child, ok := cur.children[part]
			if !ok {
				kind := tui.NodeDir
				if leaf {
					kind = tui.NodeFile
				}
				child = &trieNode{
					name:     part,
					kind:     kind,
					children: map[string]*trieNode{},
				}
				cur.children[part] = child
				cur.order = append(cur.order, part)
			}
			if leaf {
				child.status = e.status
				child.note = e.note
			} else {
				// Mark the station subtree with NodeCurrent.
				prefix := strings.Join(parts[:i+1], "/")
				if prefix == stationClean {
					child.status = tui.NodeCurrent
				}
			}
			cur = child
		}
	}

	// Walk trie → []tui.TreeNode using insertion order.
	var walk func(n *trieNode) []tui.TreeNode
	walk = func(n *trieNode) []tui.TreeNode {
		out := make([]tui.TreeNode, 0, len(n.order))
		for _, name := range n.order {
			c := n.children[name]
			node := tui.TreeNode{
				Name:   c.name,
				Kind:   c.kind,
				Status: c.status,
				Note:   c.note,
			}
			if c.kind == tui.NodeDir {
				node.Children = walk(c)
			}
			out = append(out, node)
		}
		return out
	}
	return walk(root)
}

// formatElapsed returns "MM:SS.d" formatting used by the ELAPSED chip.
func formatElapsed(d time.Duration) string {
	total := d.Seconds()
	mins := int(total) / 60
	secs := total - float64(mins*60)
	return fmt.Sprintf("%02d:%04.1f", mins, secs)
}
