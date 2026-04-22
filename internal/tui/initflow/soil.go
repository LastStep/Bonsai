package initflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
)

// ScaffoldingOption is the per-row shape consumed by SoilStage. It is
// deliberately minimal — the stage only needs a machine name (returned via
// Result), a display name + description (rendered in the row), and a
// Required flag (pinned, non-toggleable). Callers map their catalog types
// into this shape so the stage package stays independent of catalog.
type ScaffoldingOption struct {
	Name        string // machine identifier — returned verbatim in Result
	DisplayName string // human-readable label shown in each row
	Description string // one-line caption rendered muted after the label
	Required    bool   // pinned selection — cannot be toggled off
}

// SoilStage presents the scaffolding options as a hand-rolled multi-select
// list. Rationale (plan §Design decisions resolved, 2026-04-21):
// scaffolding catalog is ~4–8 items; bubbles/list ships its own styling
// that's hard to match the design's exact row layout. Hand-rolled list is
// ~120 lines of exact control over glyphs, focus highlight, and badges.
type SoilStage struct {
	Stage

	options  []ScaffoldingOption
	selected []bool // parallel to options; true = selected
	focus    int    // index of the currently-highlighted row
}

// NewSoilStage constructs the Soil stage at rail position 1. Required
// options are pre-selected and cannot be toggled off.
func NewSoilStage(ctx StageContext, options []ScaffoldingOption) *SoilStage {
	label := StageLabels[1]
	base := NewStage(
		1,
		label,
		label.English,
		ctx.Version,
		ctx.ProjectDir,
		ctx.StationDir,
		ctx.AgentDisplay,
		ctx.StartedAt,
	)

	selected := make([]bool, len(options))
	for i, opt := range options {
		if opt.Required {
			selected[i] = true
		}
	}

	return &SoilStage{
		Stage:    base,
		options:  options,
		selected: selected,
		focus:    0,
	}
}

// Init implements tea.Model — no cursor/cmd to fire on entry.
func (s *SoilStage) Init() tea.Cmd { return nil }

// Update handles arrow-key focus, space-to-toggle, enter-to-advance.
func (s *SoilStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
		s.height = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "up", "k":
			if len(s.options) == 0 {
				return s, nil
			}
			s.focus--
			if s.focus < 0 {
				s.focus = len(s.options) - 1 // wrap to bottom
			}
		case "down", "j":
			if len(s.options) == 0 {
				return s, nil
			}
			s.focus++
			if s.focus >= len(s.options) {
				s.focus = 0 // wrap to top
			}
		case " ":
			if s.focus < 0 || s.focus >= len(s.options) {
				return s, nil
			}
			// Required options ignore toggles — they stay selected.
			if s.options[s.focus].Required {
				return s, nil
			}
			s.selected[s.focus] = !s.selected[s.focus]
		case "enter":
			s.done = true
			return s, nil
		}
	}
	return s, nil
}

// View composes the Soil stage body inside the shared frame.
func (s *SoilStage) View() string {
	return s.renderFrame(s.renderBody(), s.keyHints())
}

// keyHints builds the footer key row for this stage.
func (s *SoilStage) keyHints() []KeyHint {
	return []KeyHint{
		{Key: "↑↓", Desc: "move"},
		{Key: "␣", Desc: "toggle"},
		{Key: "↵", Desc: "next"},
		{Key: "esc", Desc: "back"},
		{Key: "ctrl-c", Desc: "quit"},
	}
}

// renderBody renders the stage intro + the list of scaffolding rows. The
// body is centred inside the current terminal width via centerBlock so the
// list sits visually balanced.
func (s *SoilStage) renderBody() string {
	dim := DimStyle()
	bark := LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)

	var title string
	if s.ensoSafe {
		title = bark.Render(s.label.Kanji) + " " + white.Render(s.label.English)
	} else {
		title = white.Render(s.label.English)
	}

	intro := strings.Join([]string{
		title,
		white.Render("Tend the soil."),
		dim.Render("Shared scaffolding every agent can see — required items always on."),
	}, "\n")

	divider := RenderSectionHeader("SCAFFOLDING", PanelWidth(s.width))

	// Render each option row.
	rows := make([]string, 0, len(s.options))
	for i := range s.options {
		rows = append(rows, s.renderRow(i))
	}

	// Counter — "X of N selected · R required, always on".
	total := len(s.options)
	selCount := 0
	reqCount := 0
	for i, opt := range s.options {
		if s.selected[i] {
			selCount++
		}
		if opt.Required {
			reqCount++
		}
	}
	counter := fmt.Sprintf("%d of %d selected · %d required, always on", selCount, total, reqCount)

	body := []string{
		intro,
		"",
		"",
		divider,
		"",
	}
	body = append(body, rows...)
	body = append(body, "", dim.Render(counter))
	return centerBlock(strings.Join(body, "\n"), s.width)
}

// renderRow renders a single scaffolding option at index idx. Layout
// matches the Branches ability row for cross-stage consistency:
//
//	[border 2] [glyph 1] [sp 1] [name + * W] [sp 1] [desc W]
//
// Focus state lifts the name to ColorAccent bold (FocusedNameStyle) and
// the description from ColorRule2 to ColorAccent (FocusedDescStyle) so
// the focused row reads as one bright block. Required scaffolding items
// get an inline bark-gold "*" after the name instead of a trailing
// "REQUIRED" badge (2026-04-22 UX pass — matches Branches).
//
// Scaffolding catalog is small (≤8 items today) so no Viewport wrap —
// if the catalog ever grows past ~12 entries, revisit and reach for the
// Viewport helper in layout.go.
func (s *SoilStage) renderRow(idx int) string {
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	dim := DimStyle()
	opt := s.options[idx]
	selected := s.selected[idx]
	focused := idx == s.focus

	// Glyph: ◆ when selected (Leaf), ◇ when not (dim).
	glyph := "◇"
	glyphStyle := dim
	if selected {
		glyph = "◆"
		glyphStyle = leaf
	}

	// Left-border — shared focus token so Soil and Branches match exactly.
	border := UnfocusBorder()
	if focused {
		border = FocusBorder()
	}

	// Column widths — match Branches via ClampColumns so Soil rows line up
	// under Branches rows on the same terminal. nameW/descW absorb the
	// old tag column (no trailing REQUIRED badge any more).
	nameColW, descColW, tagColW := ClampColumns(s.width - 4)
	descColW += tagColW - 2

	// Name column — truncate against (nameW-2) when required so the " *"
	// glyph always has room.
	name := opt.DisplayName
	if name == "" {
		name = opt.Name
	}
	nameBudget := nameColW
	if opt.Required {
		nameBudget = nameColW - 2
	}
	if lipgloss.Width(name) > nameBudget {
		rr := []rune(name)
		if len(rr) > nameBudget-1 && nameBudget > 1 {
			name = string(rr[:nameBudget-1]) + "…"
		}
	}
	// Selected items render leaf-green (matches glyph + Branches precedent);
	// focus lifts unselected rows to white bold. Combined:
	//   selected+focused   → leaf bold
	//   selected           → leaf
	//   unselected+focused → white bold
	//   else               → subtle
	var nameStyle lipgloss.Style
	switch {
	case selected && focused:
		nameStyle = leaf.Bold(true)
	case selected:
		nameStyle = leaf
	case focused:
		nameStyle = FocusedNameStyle()
	default:
		nameStyle = UnfocusedNameStyle()
	}
	nameText := nameStyle.Render(name)
	if opt.Required {
		nameText += " " + RequiredGlyph()
	}
	nameCol := padRight(nameText, nameColW)

	// Description column — leaf-white when focused, dim otherwise.
	var descCol string
	if descColW > 0 {
		desc := opt.Description
		if lipgloss.Width(desc) > descColW {
			rr := []rune(desc)
			if len(rr) > descColW-1 {
				desc = string(rr[:descColW-1]) + "…"
			}
		}
		descStyle := UnfocusedDescStyle()
		if focused {
			descStyle = FocusedDescStyle()
		}
		descCol = " " + descStyle.Render(padRight(desc, descColW))
	}

	return border + glyphStyle.Render(glyph) + " " + nameCol + descCol
}

// Result returns the machine names of every selected option, preserving the
// input order. Required items are always present in the result.
func (s *SoilStage) Result() any {
	out := make([]string, 0, len(s.options))
	for i, opt := range s.options {
		if s.selected[i] {
			out = append(out, opt.Name)
		}
	}
	return out
}

// Reset clears the completion flag so re-entry behaves correctly, preserving
// the user's selections + focus cursor position.
func (s *SoilStage) Reset() tea.Cmd {
	s.done = false
	return nil
}
