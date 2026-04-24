package removeflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// removeMode distinguishes the two remove-flow shapes Observe renders.
type removeMode int

const (
	ModeAgentRemove removeMode = iota
	ModeItemRemove
)

// ObserveStage is the pre-confirm preview at rail position 1 (観 OBSERVE).
// Agent-remove: shows the full ability tree installed on the target agent.
// Item-remove: shows the item name + type + targets.
//
// Unlike addflow's Observe, this stage is display-only — it does not gate on
// user input. The Confirm stage handles the destructive yes/no decision.
//
// Result: nil. The harness advances on Enter; SelectStage's result (if any)
// stays in the prev[] slice for Confirm to consume.
type ObserveStage struct {
	initflow.Stage

	mode removeMode

	// Agent-remove inputs.
	agentDisplay string
	agentName    string
	workspace    string
	skills       []string
	workflows    []string
	protocols    []string
	sensors      []string
	routines     []string

	// Item-remove inputs.
	itemDisplay string
	itemType    string
	targets     []AgentOption // agent name + workspace + "all" flag per target
}

// NewObserveAgent constructs Observe in agent-remove mode.
func NewObserveAgent(ctx initflow.StageContext, agentName, agentDisplay, workspace string, skills, workflows, protocols, sensors, routines []string) *ObserveStage {
	s := newObserve(ctx, ModeAgentRemove)
	s.agentName = agentName
	s.agentDisplay = agentDisplay
	s.workspace = workspace
	s.skills = skills
	s.workflows = workflows
	s.protocols = protocols
	s.sensors = sensors
	s.routines = routines
	return s
}

// NewObserveItem constructs Observe in item-remove mode. targets may be
// overwritten on entry via SetTargets when the upstream Select stage resolves
// the picker choice into a concrete target list.
func NewObserveItem(ctx initflow.StageContext, itemDisplay, itemType string, targets []AgentOption) *ObserveStage {
	s := newObserve(ctx, ModeItemRemove)
	s.itemDisplay = itemDisplay
	s.itemType = itemType
	s.targets = targets
	return s
}

func newObserve(ctx initflow.StageContext, mode removeMode) *ObserveStage {
	label := StageLabels[StageIdxObserve]
	base := initflow.NewStage(
		StageIdxObserve,
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
	return &ObserveStage{Stage: base, mode: mode}
}

// SetTargets overwrites the item-remove target list — used when the upstream
// Select stage resolves the picker into a concrete subset.
func (s *ObserveStage) SetTargets(targets []AgentOption) { s.targets = targets }

// Init implements tea.Model — no cmd on entry.
func (s *ObserveStage) Init() tea.Cmd { return nil }

// Update handles Enter (advance) and Esc (back).
func (s *ObserveStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(m.Width, m.Height)
	case tea.KeyMsg:
		switch m.String() {
		case "enter":
			s.MarkDone()
			return s, nil
		}
	}
	return s, nil
}

// View composes the Observe body inside the shared frame.
func (s *ObserveStage) View() string {
	return s.RenderFrame(s.renderBody(), s.keyHints())
}

func (s *ObserveStage) keyHints() []initflow.KeyHint {
	return []initflow.KeyHint{
		{Key: "↵", Desc: "continue"},
		{Key: "esc", Desc: "back"},
		{Key: "ctrl-c", Desc: "quit"},
	}
}

func (s *ObserveStage) renderBody() string {
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()

	var title string
	if s.EnsoSafe() {
		title = bark.Render(s.Label().Kanji) + " " + white.Render(s.Label().English)
	} else {
		title = white.Render(s.Label().English)
	}

	var introLine2 string
	switch s.mode {
	case ModeAgentRemove:
		introLine2 = white.Render("About to uproot this agent.")
	case ModeItemRemove:
		introLine2 = white.Render("About to uproot this ability.")
	}
	intro := strings.Join([]string{
		title,
		introLine2,
		dim.Render("Nothing is deleted until you confirm on the next stage."),
	}, "\n")

	var summary string
	if s.mode == ModeAgentRemove {
		summary = s.renderAgentSummary()
	} else {
		summary = s.renderItemSummary()
	}

	body := []string{
		intro,
		"",
		"",
		summary,
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

func (s *ObserveStage) renderAgentSummary() string {
	bark := initflow.LabelStyle()
	dim := initflow.DimStyle()
	value := initflow.ValueStyle()

	panelW := initflow.PanelWidth(s.Width())
	targetHeader := initflow.RenderSectionHeader("TARGET", panelW)
	abilitiesHeader := initflow.RenderSectionHeader("ABILITIES", panelW)

	const labelW = 14
	const indent = "  "

	display := s.agentDisplay
	if display == "" {
		display = s.agentName
	}
	if display == "" {
		display = "—"
	}
	workspace := s.workspace
	if workspace == "" {
		workspace = "—"
	}

	renderRow := func(label string, items []string) string {
		row := indent + bark.Render(initflow.PadRight(label, labelW))
		if len(items) == 0 {
			return row + dim.Render("(none)")
		}
		listStr := strings.Join(items, ", ")
		const maxList = 36
		if lipgloss.Width(listStr) > maxList {
			runes := []rune(listStr)
			cut := maxList - 1
			if cut > len(runes) {
				cut = len(runes)
			}
			listStr = string(runes[:cut]) + "…"
		}
		return row + value.Render(fmt.Sprintf("%d", len(items))) + dim.Render("  "+listStr)
	}

	rows := []string{
		targetHeader,
		indent + bark.Render(initflow.PadRight("AGENT", labelW)) + value.Render(display),
		indent + bark.Render(initflow.PadRight("WORKSPACE", labelW)) + value.Render(workspace),
		"",
		abilitiesHeader,
		renderRow("SKILLS", s.skills),
		renderRow("WORKFLOWS", s.workflows),
		renderRow("PROTOCOLS", s.protocols),
		renderRow("SENSORS", s.sensors),
		renderRow("ROUTINES", s.routines),
	}
	return strings.Join(rows, "\n")
}

func (s *ObserveStage) renderItemSummary() string {
	bark := initflow.LabelStyle()
	dim := initflow.DimStyle()
	value := initflow.ValueStyle()

	panelW := initflow.PanelWidth(s.Width())
	targetHeader := initflow.RenderSectionHeader("TARGET", panelW)
	fromHeader := initflow.RenderSectionHeader("FROM", panelW)

	const labelW = 14
	const indent = "  "

	itemType := s.itemType
	if itemType == "" {
		itemType = "—"
	}
	display := s.itemDisplay
	if display == "" {
		display = "—"
	}

	rows := []string{
		targetHeader,
		indent + bark.Render(initflow.PadRight("ITEM", labelW)) + value.Render(display),
		indent + bark.Render(initflow.PadRight("TYPE", labelW)) + value.Render(itemType),
		"",
		fromHeader,
	}

	if len(s.targets) == 0 {
		rows = append(rows, indent+dim.Render("(no targets)"))
	} else {
		for _, t := range s.targets {
			if t.All {
				// Aggregate rows shouldn't reach observe; skip defensively.
				continue
			}
			label := t.DisplayName
			if label == "" {
				label = t.Name
			}
			rows = append(rows,
				indent+bark.Render(initflow.PadRight(label, labelW))+
					value.Render(t.Workspace))
		}
	}
	return strings.Join(rows, "\n")
}

// Result returns nil — Observe is a preview, not a decision point.
func (s *ObserveStage) Result() any { return nil }

// Reset clears completion on re-entry.
func (s *ObserveStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}
