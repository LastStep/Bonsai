package removeflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// SelectStage is the item-remove agent picker at rail position 0. Rendered
// only when the caller has computed that more than one installed agent
// carries the target ability — single-match removals skip this stage.
//
// Result: machine name of the chosen agent ("_all_" for the aggregate row).
type SelectStage struct {
	initflow.Stage

	itemDisplay string
	itemType    string
	options     []AgentOption
	focus       int

	viewport initflow.Viewport
}

// NewSelectStage constructs the remove-flow agent picker. options must carry
// at least two rows (the caller gates on len(matches) > 1 before instantiating)
// and typically ends with an aggregate "All agents" entry so the user can
// remove from every installed target in one pass.
func NewSelectStage(ctx initflow.StageContext, itemDisplay, itemType string, options []AgentOption) *SelectStage {
	label := StageLabels[StageIdxSelect]
	base := initflow.NewStage(
		StageIdxSelect,
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
	return &SelectStage{
		Stage:       base,
		itemDisplay: itemDisplay,
		itemType:    itemType,
		options:     options,
		focus:       0,
	}
}

// Init implements tea.Model — no cmd on entry.
func (s *SelectStage) Init() tea.Cmd { return nil }

// Update handles focus movement + Enter-to-advance.
func (s *SelectStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(m.Width, m.Height)
	case tea.KeyMsg:
		switch m.String() {
		case "up", "k":
			if len(s.options) == 0 {
				return s, nil
			}
			s.focus--
			if s.focus < 0 {
				s.focus = 0
			}
		case "down", "j":
			if len(s.options) == 0 {
				return s, nil
			}
			s.focus++
			if s.focus >= len(s.options) {
				s.focus = len(s.options) - 1
			}
		case "enter":
			s.MarkDone()
			return s, nil
		}
	}
	return s, nil
}

// View composes the Select stage body inside the shared frame.
func (s *SelectStage) View() string {
	return s.RenderFrame(s.renderBody(), s.keyHints())
}

func (s *SelectStage) keyHints() []initflow.KeyHint {
	return []initflow.KeyHint{
		{Key: "↑↓", Desc: "move"},
		{Key: "↵", Desc: "pick"},
		{Key: "esc", Desc: "back"},
		{Key: "ctrl-c", Desc: "quit"},
	}
}

func (s *SelectStage) renderBody() string {
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()

	var title string
	if s.EnsoSafe() {
		title = bark.Render(s.Label().Kanji) + " " + white.Render(s.Label().English)
	} else {
		title = white.Render(s.Label().English)
	}

	introLine2 := white.Render("Pick the agent to uproot from.")
	if s.itemDisplay != "" {
		introLine2 = white.Render(fmt.Sprintf("%s is installed on multiple agents — pick one.", s.itemDisplay))
	}
	intro := strings.Join([]string{
		title,
		introLine2,
		dim.Render("Each row is an agent with this ability installed."),
	}, "\n")

	divider := initflow.RenderSectionHeader("AGENTS", initflow.PanelWidth(s.Width()))

	rows := make([]string, 0, len(s.options))
	for i := range s.options {
		rows = append(rows, s.renderRow(i))
	}

	listH := s.listHeight()
	listBody := strings.Join(rows, "\n")
	if listH > 0 && listH < len(rows) {
		s.viewport.SetLines(rows)
		s.viewport.SetHeight(listH)
		s.viewport.Follow(s.focus)
		listBody = s.viewport.View()
	}

	counter := fmt.Sprintf("%d targets", len(s.options))

	body := []string{
		intro,
		"",
		"",
		divider,
		"",
		listBody,
		"",
		dim.Render(counter),
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

// listHeight returns the visible row budget. Mirrors addflow.SelectStage's
// chrome-accounting: intro (3) + blanks (2) + divider (1) + blank (1) +
// blank-after-list (1) + counter (1) = 9 body rows + 10 chrome rows. Floor at 3.
func (s *SelectStage) listHeight() int {
	h := s.Height()
	if h <= 0 {
		return 0
	}
	const chromeRows = 10
	const fixedBodyRows = 9
	v := h - chromeRows - fixedBodyRows
	if v < 3 {
		v = 3
	}
	return v
}

// renderRow composes a single agent row. Layout mirrors addflow.SelectStage:
//
//	[border 2] [glyph 1] [sp 1] [name] [sp 2] [workspace muted]
func (s *SelectStage) renderRow(idx int) string {
	opt := s.options[idx]
	focused := idx == s.focus

	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	dim := initflow.DimStyle()

	glyph := " "
	glyphStyle := dim
	if focused {
		glyph = "▸"
		glyphStyle = leaf
	}

	border := initflow.UnfocusBorder()
	if focused {
		border = initflow.FocusBorder()
	}

	name := opt.DisplayName
	if name == "" {
		name = opt.Name
	}
	var nameStyle lipgloss.Style
	if focused {
		nameStyle = initflow.FocusedNameStyle()
	} else {
		nameStyle = initflow.UnfocusedNameStyle()
	}

	rowTotal := s.Width() - 4
	if rowTotal < 40 {
		rowTotal = 40
	}
	const nameColW = 22
	if lipgloss.Width(name) > nameColW {
		rr := []rune(name)
		if len(rr) > nameColW-1 && nameColW > 1 {
			name = string(rr[:nameColW-1]) + "…"
		}
	}
	namePadded := initflow.PadRight(nameStyle.Render(name), nameColW)

	// Workspace column — muted, truncated if wide.
	ws := opt.Workspace
	if opt.All {
		ws = "every installed target"
	}
	descBudget := rowTotal - 2 - 1 - 1 - nameColW - 2
	if descBudget < 10 {
		descBudget = 10
	}
	if lipgloss.Width(ws) > descBudget {
		rr := []rune(ws)
		if len(rr) > descBudget-1 {
			ws = string(rr[:descBudget-1]) + "…"
		}
	}
	descStyle := initflow.UnfocusedDescStyle()
	if focused {
		descStyle = initflow.FocusedDescStyle()
	}
	descPadded := initflow.PadRight(descStyle.Render(ws), descBudget)

	return border + glyphStyle.Render(glyph) + " " + namePadded + "  " + descPadded
}

// Result returns the selected agent's machine name, or "" when the stage
// has no options (defensive — ctor is gated on len > 1).
func (s *SelectStage) Result() any {
	if s.focus < 0 || s.focus >= len(s.options) {
		return ""
	}
	return s.options[s.focus].Name
}

// Reset clears the completion flag so re-entry renders fresh.
func (s *SelectStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}
