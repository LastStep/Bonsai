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
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
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

	divider := leaf.Render(strings.Repeat("─", 3)) + " " +
		bark.Render("SCAFFOLDING") + " " +
		dim.Render(strings.Repeat("─", 55))

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

// renderRow renders a single scaffolding option at index idx. Focused rows
// get a Leaf left-border + leaf-tint background; selected rows use the ◆
// glyph in Leaf, unselected use ◇ in a dimmer muted color. The REQUIRED
// badge (Bark) is right-aligned.
func (s *SoilStage) renderRow(idx int) string {
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	label := lipgloss.NewStyle().Foreground(tui.ColorSubtle)

	opt := s.options[idx]
	selected := s.selected[idx]

	// Glyph: ◆ when selected (Leaf), ◇ when not (dim).
	glyph := "◇"
	glyphStyle := dim
	if selected {
		glyph = "◆"
		glyphStyle = leaf
	}

	// Left-border: Leaf "│ " for the focused row, two spaces otherwise.
	// The focus tint background is approximated via the Leaf-dim foreground
	// on the border glyph — LipGloss AdaptiveColor degrades predictably on
	// 256-color terminals without forcing an ANSI background (keeps the row
	// readable on both dark and light themes).
	border := "  "
	if idx == s.focus {
		border = lipgloss.NewStyle().Foreground(tui.ColorPrimary).Render("│ ")
	}

	// Name column — bold when focused, regular otherwise. Keep it padded so
	// descriptions align across rows.
	name := opt.DisplayName
	if name == "" {
		name = opt.Name
	}
	nameCol := label.Render(padRight(name, 20))
	if idx == s.focus {
		nameCol = bark.Render(padRight(name, 20))
	}

	// Description column — muted, truncated at ~50 cols to keep the row
	// from wrapping in narrow terminals.
	desc := opt.Description
	if len(desc) > 50 {
		desc = desc[:49] + "…"
	}
	descCol := dim.Render(desc)

	// Required badge — right-aligned after the description, in Bark.
	badge := ""
	if opt.Required {
		badge = "  " + bark.Render("REQUIRED")
	}

	return border + glyphStyle.Render(glyph) + " " + nameCol + " " + descCol + badge
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
