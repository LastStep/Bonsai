package initflow

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

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
	base.applyContextHeader(ctx)
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
		case "j", "down":
			s.viewport.ScrollBy(1)
			return s, nil
		case "k", "up":
			s.viewport.ScrollBy(-1)
			return s, nil
		case "pgdown", " ":
			step := s.viewport.height
			if step < 1 {
				step = 1
			}
			s.viewport.ScrollBy(step)
			return s, nil
		case "pgup", "b":
			step := s.viewport.height
			if step < 1 {
				step = 1
			}
			s.viewport.ScrollBy(-step)
			return s, nil
		}
	}
	return s, nil
}

// View renders the Planted stage as a chromeless, full-screen celebration
// frame. Unlike the other stages, Planted deliberately bypasses the shared
// header + rail + footer chrome (2026-04-22 UX pass) — the flow has
// finished and the visual framing should feel like an exit card rather
// than another in-flow step. The body is centred vertically inside the
// live terminal height so the hero lands mid-screen.
func (s *PlantedStage) View() string {
	width := s.width
	if width <= 0 {
		width = 80
	}
	height := s.height
	if height <= 0 {
		height = 24
	}
	if TerminalTooSmall(s.width, s.height) {
		return RenderMinSizeFloor(s.width, s.height)
	}

	body := s.renderBody()

	// Vertically center inside the AltScreen.
	rows := strings.Count(body, "\n") + 1
	topPad := (height - rows) / 2
	if topPad < 1 {
		topPad = 1
	}
	bottomPad := height - rows - topPad
	if bottomPad < 0 {
		bottomPad = 0
	}
	return strings.Repeat("\n", topPad) + body + strings.Repeat("\n", bottomPad)
}

// renderBody renders the full post-plant frame. No hero/stats dividers —
// each major block carries its own section header (WRITTEN / SUMMARY /
// NEXT) via the shared RenderSectionHeader helper so the whole stage
// reads as a single centred card.
func (s *PlantedStage) renderBody() string {
	dim := DimStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)

	created, updated, _, _, conflicts := s.counts()

	// Hero block — no ELAPSED chip (dropped 2026-04-22 UX pass).
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

	writtenBlock := s.renderWrittenBlock()
	summaryBlock := s.renderSummaryBlock()
	nextBlock := s.renderNextSteps()

	hintText := "↵  exit  ·  q  quit"
	if up, down := s.viewport.HasMore(); up || down {
		hintText = "j/k  scroll  ·  ↵  exit  ·  q  quit"
	}
	hint := dim.Render(hintText)

	body := strings.Join([]string{
		heroTitle,
		heroSub,
		heroStats,
		"",
		"",
		writtenBlock,
		"",
		"",
		summaryBlock,
		"",
		"",
		nextBlock,
		"",
		hint,
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

// renderWrittenBlock renders the tree of written files. NodeStatus mapping:
//
//	ActionCreated              → NodeNew
//	ActionUpdated / ActionForced → NodeNew (plus "UPDATED" badge via Note)
//	ActionUnchanged            → NodeNormal
//	ActionSkipped / ActionConflict → omitted
func (s *PlantedStage) renderWrittenBlock() string {
	dim := DimStyle()

	panelW := PanelWidth(s.width)
	header := RenderSectionHeader("WRITTEN", panelW)

	if s.wr == nil || len(s.wr.Files) == 0 {
		return header + "\n  " + dim.Render("(nothing written)")
	}

	tree := buildPlantedTree(s.wr, s.projectDir, s.stationDir)
	// Tree spans the full panel width now that Planted is a single-column
	// layout (no 2-col grid split). Clamp to PanelContentWidth so wide
	// terminals don't spread the tree across the whole row.
	rendered := tui.RenderFileTree(tree, tui.FileTreeOpts{
		Dense:    true,
		MaxWidth: panelW,
	})

	// Wrap in Viewport when the rendered tree exceeds a reasonable body
	// budget. Always push content + height so HasMore() stays accurate when
	// the terminal resizes across the overflow threshold.
	lines := strings.Split(rendered, "\n")
	budget := s.listHeight()
	if budget <= 0 {
		budget = len(lines)
	}
	s.viewport.SetLines(lines)
	s.viewport.SetHeight(budget)
	if len(lines) > budget {
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

// renderSummaryBlock renders the SUMMARY panel — agent + abilities counts.
// Single-column inline layout so wide terminals don't spread values far
// from their labels (2026-04-22 UX pass — matched Observe's PROJECT block).
func (s *PlantedStage) renderSummaryBlock() string {
	dim := DimStyle()
	bark := LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent)

	header := RenderSectionHeader("SUMMARY", PanelWidth(s.width))

	total := s.skillCount + s.workflowCount + s.protocolCount + s.sensorCount + s.routineCount

	const labelW = 14
	const indent = "  "
	rows := []string{
		header,
		indent + bark.Render(padRight("AGENT", labelW)) + white.Render(s.agentDisplay) +
			dim.Render(" → "+s.stationDir),
		indent + bark.Render(padRight("ABILITIES", labelW)) + white.Render(fmt.Sprintf("%d wired", total)),
		indent + strings.Repeat(" ", labelW) + dim.Render(fmt.Sprintf(
			"%d skills · %d workflows · %d protocols · %d sensors · %d routines",
			s.skillCount, s.workflowCount, s.protocolCount, s.sensorCount, s.routineCount,
		)),
	}
	return strings.Join(rows, "\n")
}

// renderNextSteps renders the three-step "next" callout.
func (s *PlantedStage) renderNextSteps() string {
	dim := DimStyle()
	bark := LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent)

	header := RenderSectionHeader("NEXT", PanelWidth(s.width))

	numerals := []string{"1", "2", "3"}

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
