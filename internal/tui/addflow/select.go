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

// renderRow renders a single agent entry (Plan 27 PR2 §C8 layout). Row:
//
//		[border 2] [glyph 1] [sp 1] [name] [sp 2] [description fill...] [sp 2] [installed-badge right-aligned]
//
//	  - The word "Agent" (case-insensitive) is stripped from DisplayName at
//	    render time so "Tech Lead Agent" → "Tech Lead". opt.DisplayName is not
//	    mutated; machine-name opt.Name is untouched.
//	  - The full description is rendered; truncation only fires if it would
//	    overflow the row's visible width.
//	  - The "(installed)" badge is pad-right aligned at the trailing edge of
//	    the row — outside the name column, outside the description column.
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

	// Display name — strip trailing " Agent" / "Agent" (case-insensitive)
	// so "Tech Lead Agent" → "Tech Lead". Fall back to machine name when
	// the stripped display name is empty.
	name := stripAgentSuffix(opt.DisplayName)
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

	// Trailing badge — right-aligned at the row's visible edge.
	badgeText := ""
	if opt.Installed {
		badgeText = bark.Render("(installed)")
	}
	badgeW := lipgloss.Width(badgeText)

	// Row budget — everything between the left border and the right edge.
	rowTotal := s.Width() - 4
	if rowTotal < 40 {
		rowTotal = 40
	}

	// Name column — fixed budget so rows align vertically.
	const nameColW = 22
	if lipgloss.Width(name) > nameColW {
		rr := []rune(name)
		if len(rr) > nameColW-1 && nameColW > 1 {
			name = string(rr[:nameColW-1]) + "…"
		}
	}
	namePadded := initflow.PadRight(nameStyle.Render(name), nameColW)

	// Description budget = row total - border(2) - glyph(1) - sp(1) -
	// name(nameColW) - sp(2) - badge(badgeW) - sp-before-badge(2 iff badge).
	badgeGap := 0
	if badgeW > 0 {
		badgeGap = 2
	}
	descBudget := rowTotal - 2 - 1 - 1 - nameColW - 2 - badgeW - badgeGap
	if descBudget < 10 {
		descBudget = 10
	}

	desc := opt.Description
	if lipgloss.Width(desc) > descBudget {
		rr := []rune(desc)
		if len(rr) > descBudget-1 {
			desc = string(rr[:descBudget-1]) + "…"
		}
	}
	descStyle := initflow.UnfocusedDescStyle()
	if focused {
		descStyle = initflow.FocusedDescStyle()
	}
	descPadded := initflow.PadRight(descStyle.Render(desc), descBudget)

	row := border + glyphStyle.Render(glyph) + " " + namePadded + "  " + descPadded
	if badgeW > 0 {
		row += "  " + badgeText
	}
	return row
}

// stripAgentSuffix drops a trailing " Agent" / "Agent" token (case-
// insensitive) from s. The replacement is display-only; the caller stores
// the original DisplayName unchanged. Used by SelectStage.renderRow per
// Plan 27 PR2 §C8.
func stripAgentSuffix(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return trimmed
	}
	lower := strings.ToLower(trimmed)
	const suffix = "agent"
	if strings.HasSuffix(lower, " "+suffix) {
		return strings.TrimSpace(trimmed[:len(trimmed)-len(suffix)-1])
	}
	if lower == suffix {
		return ""
	}
	return trimmed
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
