package removeflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// yieldMode distinguishes the two terminal card variants Yield renders.
type yieldMode int

const (
	yieldModeAgentSuccess yieldMode = iota
	yieldModeItemSuccess
)

// YieldStage is the terminal completion card at rail position 3
// (StageIdxYield, 結 YIELD). Renders a chromeless, vertically-centered exit
// card matching addflow.YieldStage's layout beat.
//
// Two modes:
//
//   - agent-success — "N abilities uprooted · <agent> · lock synced"
//   - item-success  — "<item> removed from <target>"
//
// Hints block — Phase E ships a minimal 2-layer placeholder (next CLI +
// workflow tip). The 3-layer contract (catalog-driven next_cli / next_workflow /
// ai_prompts) integrates during the Plan 31 PR2 merge when β's Phase H hints
// infrastructure lands; this stage renders a stable surface either way.
type YieldStage struct {
	initflow.Stage

	mode yieldMode

	agentDisplay string
	workspace    string
	counts       AbilityCounts

	itemDisplay string
	itemType    string
	targets     []AgentOption
}

// NewYieldAgentSuccess renders the agent-remove happy-path card.
func NewYieldAgentSuccess(ctx initflow.StageContext, agentDisplay, workspace string, counts AbilityCounts) *YieldStage {
	y := newYield(ctx, yieldModeAgentSuccess)
	y.agentDisplay = agentDisplay
	y.workspace = workspace
	y.counts = counts
	return y
}

// NewYieldItemSuccess renders the item-remove happy-path card.
func NewYieldItemSuccess(ctx initflow.StageContext, itemDisplay, itemType string, targets []AgentOption) *YieldStage {
	y := newYield(ctx, yieldModeItemSuccess)
	y.itemDisplay = itemDisplay
	y.itemType = itemType
	y.targets = targets
	return y
}

func newYield(ctx initflow.StageContext, mode yieldMode) *YieldStage {
	label := StageLabels[StageIdxYield]
	base := initflow.NewStage(
		StageIdxYield,
		label,
		label.English,
		ctx.Version,
		ctx.ProjectDir,
		ctx.StationDir,
		ctx.AgentDisplay,
		ctx.StartedAt,
	)
	base.ApplyContextHeader(ctx)
	base.SetRailLabels(StageLabels)
	return &YieldStage{Stage: base, mode: mode}
}

// Chromeless matches addflow.YieldStage — the stage renders a full-screen
// chromeless exit card (no header/rail/footer chrome).
func (s *YieldStage) Chromeless() bool { return true }

// Init implements tea.Model.
func (s *YieldStage) Init() tea.Cmd { return nil }

// Update waits for ↵ / q / esc acknowledgement.
func (s *YieldStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(m.Width, m.Height)
	case tea.KeyMsg:
		switch m.String() {
		case "enter", "q", "esc":
			s.MarkDone()
			return s, nil
		}
	}
	return s, nil
}

// View renders the chromeless exit frame.
func (s *YieldStage) View() string {
	w := s.Width()
	h := s.Height()
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	if initflow.TerminalTooSmall(s.Width(), s.Height()) {
		return initflow.RenderMinSizeFloor(s.Width(), s.Height())
	}

	dim := initflow.DimStyle()
	hint := dim.Render("↵  exit  ·  q  quit")
	body := s.renderBody() + "\n\n" + initflow.CenterBlock(hint, w)

	rows := strings.Count(body, "\n") + 1
	topPad := (h - rows) / 2
	if topPad < 1 {
		topPad = 1
	}
	bottomPad := h - rows - topPad
	if bottomPad < 0 {
		bottomPad = 0
	}
	return strings.Repeat("\n", topPad) + body + strings.Repeat("\n", bottomPad)
}

func (s *YieldStage) renderBody() string {
	switch s.mode {
	case yieldModeItemSuccess:
		return s.renderItemSuccess()
	default:
		return s.renderAgentSuccess()
	}
}

func (s *YieldStage) renderAgentSuccess() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	value := initflow.ValueStyle()

	var heroTitle string
	if s.EnsoSafe() {
		heroTitle = leaf.Render(s.Label().Kanji + " · UPROOTED")
	} else {
		heroTitle = leaf.Render("UPROOTED")
	}

	agentDisplay := s.agentDisplay
	if agentDisplay == "" {
		agentDisplay = "agent"
	}
	workspace := s.workspace
	if workspace == "" {
		workspace = "—"
	}

	totalAbilities := s.counts.Total()
	heroSub := white.Render(fmt.Sprintf("%s is uprooted.", agentDisplay))
	heroStats := dim.Render(fmt.Sprintf(
		"%d abilities cleared · %s · lock synced",
		totalAbilities, agentDisplay,
	))

	summaryHeader := initflow.RenderSectionHeader("SUMMARY", initflow.PanelWidth(s.Width()))
	const labelW = 14
	const indent = "  "
	summaryRows := []string{
		summaryHeader,
		indent + bark.Render(initflow.PadRight("AGENT", labelW)) + value.Render(agentDisplay) +
			dim.Render(" → "+workspace),
		indent + bark.Render(initflow.PadRight("ABILITIES", labelW)) + value.Render(fmt.Sprintf("%d uprooted", totalAbilities)),
	}
	if totalAbilities > 0 {
		summaryRows = append(summaryRows,
			indent+strings.Repeat(" ", labelW)+dim.Render(fmt.Sprintf(
				"%d skills · %d workflows · %d protocols · %d sensors · %d routines",
				s.counts.Skills, s.counts.Workflows, s.counts.Protocols,
				s.counts.Sensors, s.counts.Routines,
			)))
	}

	hints := s.renderHints([]hintLine{
		{cmd: "$ bonsai list", caption: "see what's still installed"},
		{cmd: "$ bonsai catalog", caption: "browse replacement abilities any time"},
	})

	body := []string{
		heroTitle,
		heroSub,
		heroStats,
		"",
		"",
		strings.Join(summaryRows, "\n"),
		"",
		"",
		hints,
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

func (s *YieldStage) renderItemSuccess() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	value := initflow.ValueStyle()

	var heroTitle string
	if s.EnsoSafe() {
		heroTitle = leaf.Render(s.Label().Kanji + " · UPROOTED")
	} else {
		heroTitle = leaf.Render("UPROOTED")
	}

	item := s.itemDisplay
	if item == "" {
		item = "item"
	}
	itemType := s.itemType
	if itemType == "" {
		itemType = "item"
	}

	// Targets line — comma-joined agent display names.
	var targetNames []string
	for _, t := range s.targets {
		if t.All {
			continue
		}
		label := t.DisplayName
		if label == "" {
			label = t.Name
		}
		targetNames = append(targetNames, label)
	}
	targetsStr := "—"
	if len(targetNames) > 0 {
		targetsStr = strings.Join(targetNames, ", ")
	}

	heroSub := white.Render(fmt.Sprintf("%s removed.", item))
	heroStats := dim.Render(fmt.Sprintf(
		"%d target(s) updated · lock synced",
		len(targetNames),
	))

	summaryHeader := initflow.RenderSectionHeader("SUMMARY", initflow.PanelWidth(s.Width()))
	const labelW = 14
	const indent = "  "
	summaryRows := []string{
		summaryHeader,
		indent + bark.Render(initflow.PadRight("ITEM", labelW)) + value.Render(item),
		indent + bark.Render(initflow.PadRight("TYPE", labelW)) + value.Render(itemType),
		indent + bark.Render(initflow.PadRight("FROM", labelW)) + value.Render(targetsStr),
	}

	hints := s.renderHints([]hintLine{
		{cmd: "$ bonsai list", caption: "verify the ability is gone"},
		{cmd: "$ bonsai catalog", caption: "browse replacement abilities any time"},
	})

	body := []string{
		heroTitle,
		heroSub,
		heroStats,
		"",
		"",
		strings.Join(summaryRows, "\n"),
		"",
		"",
		hints,
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

// hintLine is a single "$ cmd — caption" hint row for the Yield hints block.
// Plan 31 Phase H replaces this with a 3-layer hints renderer shared across
// all yield stages; until then, Yield renders a 2-layer placeholder so the
// success card has a concrete next-step affordance.
type hintLine struct {
	cmd     string
	caption string
}

// renderHints composes a two-layer hints block under a NEXT divider. Phase H
// integration swaps this for a catalog-driven hints renderer (next CLI +
// workflow + AI-prompt) during the PR2 merge.
func (s *YieldStage) renderHints(lines []hintLine) string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)

	header := initflow.RenderSectionHeader("NEXT", initflow.PanelWidth(s.Width()))
	rows := []string{header}
	for i, l := range lines {
		rows = append(rows,
			"  "+bark.Render(fmt.Sprintf("%d", i+1))+"  "+white.Render(l.cmd))
		rows = append(rows, "     "+dim.Render(l.caption))
	}
	return strings.Join(rows, "\n")
}

// Result returns nil — Yield is terminal.
func (s *YieldStage) Result() any { return nil }

// Reset clears completion on re-entry.
func (s *YieldStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}
