package addflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// SelectStage is the single-choice agent picker at rail position 0. Renders
// the full agent catalog as a Branches-style list — Leaf "│ " focus border,
// white-bold focused name, ColorRule2 description on unfocused rows — and
// appends an "(installed)" suffix after any agent already present in the
// config.
//
// Result: machine name of the chosen agent (string).
type SelectStage struct {
	initflow.Stage

	options []AgentOption
	focus   int

	// Viewport wraps the list when the catalog exceeds available body rows.
	viewport initflow.Viewport
}

// NewSelectStage constructs the Select stage. options is typically built via
// BuildAgentOptions against the loaded catalog + project config.
func NewSelectStage(ctx initflow.StageContext, options []AgentOption) *SelectStage {
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
		Stage:   base,
		options: options,
		focus:   0,
	}
}

// BuildAgentOptions maps catalog agents into SelectStage's row shape,
// flagging each entry as Installed when the config already contains that
// agent type. Catalog order is preserved verbatim.
func BuildAgentOptions(cat *catalog.Catalog, installed map[string]bool) []AgentOption {
	out := make([]AgentOption, 0, len(cat.Agents))
	for _, a := range cat.Agents {
		display := a.DisplayName
		if display == "" {
			display = catalog.DisplayNameFrom(a.Name)
		}
		out = append(out, AgentOption{
			Name:        a.Name,
			DisplayName: display,
			Description: a.Description,
			Installed:   installed[a.Name],
		})
	}
	return out
}

// Init implements tea.Model — no cmd to fire on entry.
func (s *SelectStage) Init() tea.Cmd { return nil }

// Update handles focus movement + Enter-to-advance. Esc is consumed by the
// harness at the root — stage 0 esc is a no-op (aborts only via Ctrl-C).
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
		{Key: "ctrl-c", Desc: "quit"},
	}
}

// renderBody composes intro + CATEGORY divider + list + counter. Single-
// column layout (no tabs) — mirrors SoilStage's visual rhythm.
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

	intro := strings.Join([]string{
		title,
		white.Render("Pick the agent to add."),
		dim.Render("Each agent ships its own defaults — review and adjust in the next stages."),
	}, "\n")

	divider := initflow.RenderSectionHeader("AGENTS", initflow.PanelWidth(s.Width()))

	// Render each row.
	rows := make([]string, 0, len(s.options))
	for i := range s.options {
		rows = append(rows, s.renderRow(i))
	}

	// Viewport wrap when row count exceeds available budget.
	listH := s.listHeight()
	listBody := strings.Join(rows, "\n")
	if listH > 0 && listH < len(rows) {
		s.viewport.SetLines(rows)
		s.viewport.SetHeight(listH)
		s.viewport.Follow(s.focus)
		listBody = s.viewport.View()
	}

	installed := 0
	for _, o := range s.options {
		if o.Installed {
			installed++
		}
	}
	counter := fmt.Sprintf("%d agents · %d already installed", len(s.options), installed)

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

// listHeight returns the visible row budget for the agent list. Body fixed
// rows: intro (3) + blanks (2) + divider (1) + blank (1) + blank (1) +
// counter (1) = 9 rows. Chrome (10). Floor at 3.
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

// renderRow renders a single agent entry. Matches BranchesStage row layout
// for cross-flow consistency:
//
//	[border 2] [glyph 1] [sp 1] [name + tag] [sp 1] [desc]
//
// Focused rows use FocusBorder + FocusedNameStyle + FocusedDescStyle; the
// "(installed)" suffix renders in bark gold after the name so the badge sits
// inside the name column.
func (s *SelectStage) renderRow(idx int) string {
	opt := s.options[idx]
	focused := idx == s.focus

	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()

	// Single-choice glyph: ▸ on focus, · otherwise. No ◆/◇ — there's no
	// persistent selection state before Enter commits.
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

	nameColW, descColW, tagColW := initflow.ClampColumns(s.Width() - 4)
	// Recover the tag column for desc (no trailing badge) — same move as
	// Branches post-2026-04-22 polish.
	descColW += tagColW - 2

	name := opt.DisplayName
	if name == "" {
		name = opt.Name
	}

	var nameStyle lipgloss.Style
	switch {
	case focused:
		nameStyle = initflow.FocusedNameStyle()
	default:
		nameStyle = initflow.UnfocusedNameStyle()
	}

	// Reserve space for the "(installed)" suffix when present so the name is
	// truncated before the badge rather than after.
	suffix := ""
	if opt.Installed {
		suffix = " " + bark.Render("(installed)")
	}
	suffixW := lipgloss.Width(suffix)

	nameBudget := nameColW - suffixW
	if nameBudget < 6 {
		nameBudget = 6
	}
	if lipgloss.Width(name) > nameBudget {
		rr := []rune(name)
		if len(rr) > nameBudget-1 && nameBudget > 1 {
			name = string(rr[:nameBudget-1]) + "…"
		}
	}
	nameText := nameStyle.Render(name) + suffix
	nameCol := initflow.PadRight(nameText, nameColW)

	// Description column.
	var descCol string
	if descColW > 0 {
		desc := opt.Description
		if lipgloss.Width(desc) > descColW {
			rr := []rune(desc)
			if len(rr) > descColW-1 {
				desc = string(rr[:descColW-1]) + "…"
			}
		}
		descStyle := initflow.UnfocusedDescStyle()
		if focused {
			descStyle = initflow.FocusedDescStyle()
		}
		descCol = " " + descStyle.Render(initflow.PadRight(desc, descColW))
	}

	return border + glyphStyle.Render(glyph) + " " + nameCol + descCol
}

// Result returns the selected agent machine name. Zero value ("") on empty
// option slice or pre-completion.
func (s *SelectStage) Result() any {
	if s.focus < 0 || s.focus >= len(s.options) {
		return ""
	}
	return s.options[s.focus].Name
}

// Reset clears the completion flag so re-entry renders fresh. Focus index +
// option slice are preserved so the user's pick is not scrubbed.
func (s *SelectStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}
