package initflow

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
)

// Design-system tokens for the init-flow stages. Every stage body reaches
// for these instead of hand-rolling its own styles so the five screens
// render with identical panel width, section dividers, focus semantics,
// and required-item glyphs. Adding a new stage or polishing existing copy
// should mean adding/using a token here — never recreating the pattern
// inline. 2026-04-22 polish run.

// PanelContentWidth is the target content width for every stage's centered
// body. Chosen so a 120-col terminal leaves a comfortable ~18-cell margin
// on each side (84 cells of content) — matches the Branches row budget
// from ClampColumns(120)=(24,44,12)+6 overhead. Narrower terminals fall
// back to (s.width - 4).
const PanelContentWidth = 84

// PanelWidth returns the effective content width for a stage body given the
// live terminal width. Clamps to PanelContentWidth on wide terms, falls back
// to (w - 4) on narrow terms so centerBlock's padding still has headroom.
// Returns 0 on degenerate widths so callers can short-circuit.
func PanelWidth(termWidth int) int {
	if termWidth <= 0 {
		return 0
	}
	avail := termWidth - 4
	if avail < 0 {
		return 0
	}
	if avail > PanelContentWidth {
		return PanelContentWidth
	}
	return avail
}

// SectionHeaderWidth is the total cell width of a rendered section header
// ("─── LABEL ───..."). Anchored to the panel width so every section in
// every stage lines up vertically regardless of label length.
const SectionHeaderWidth = PanelContentWidth

// RenderSectionHeader renders a uniform section divider used as the header
// of every labelled block in the init flow (CATEGORIES, DETAILS, VESSEL,
// SOIL, BRANCHES, WRITTEN, SUMMARY, NEXT, etc.). Format:
//
//	─── LABEL ────────────────────────────────────
//	^ leaf       ^ bark-bold    ^ dim fill
//
// The trailing fill is sized so the total visible width equals w (clamped
// to a sane minimum). Consistent width across every section is the whole
// point — callers must NOT tack on their own trailing dashes.
func RenderSectionHeader(label string, w int) string {
	if w < 20 {
		w = 20
	}
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)

	const leadN = 3
	lead := strings.Repeat("─", leadN)
	labelCells := lipgloss.Width(label)
	// 2 cells of gap around the label (one on each side).
	fillN := w - leadN - labelCells - 2
	if fillN < 3 {
		fillN = 3
	}
	fill := strings.Repeat("─", fillN)
	return leaf.Render(lead) + " " + bark.Render(label) + " " + dim.Render(fill)
}

// RequiredGlyph is the single-char marker rendered after a name column entry
// that cannot be toggled off. Bark-gold to distinguish it from the selected-
// state green ◆/◇ glyph. Used by both Soil and Branches.
func RequiredGlyph() string {
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	return bark.Render("*")
}

// FocusBorder is the left-edge glyph rendered on the focused row of a list.
// Leaf vertical bar + space. Unfocused rows render two plain spaces to hold
// the column position.
func FocusBorder() string {
	return lipgloss.NewStyle().Foreground(tui.ColorPrimary).Render("│ ")
}

// UnfocusBorder is the two-space filler for unfocused rows so they align
// under the focus bar. Always two cells wide.
func UnfocusBorder() string { return "  " }

// FocusedNameStyle returns the style applied to an ability/scaffolding
// name on the currently-focused row: white-bold (ColorAccent). High
// legibility — the visual anchor the user's eye lands on.
func FocusedNameStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
}

// UnfocusedNameStyle returns the style for name columns outside the focus.
// Subtle neutral so the focused row reads brighter by comparison.
func UnfocusedNameStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tui.ColorSubtle)
}

// FocusedDescStyle returns the style applied to an ability/scaffolding
// description on the focused row: white (ColorAccent), not bold. Matches
// the name's colour family so focus reads as a single bright block.
// Unfocused rows render descriptions in ColorRule2 (dim).
func FocusedDescStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tui.ColorAccent)
}

// UnfocusedDescStyle returns the dim style for unfocused description text.
func UnfocusedDescStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tui.ColorRule2)
}

// ValueStyle renders property values (Observe's NAME / STATION / AGENT,
// Planted's summary numbers, etc.) in leaf-green — emphasises the
// living-plant identity the flow is building.
func ValueStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tui.ColorPrimary)
}

// LabelStyle renders property labels (NAME, DESCRIPTION, STATION, ABOUT,
// FILE, etc.) in bark-gold bold — establishes visual rhythm against the
// leaf-green values on the same row.
func LabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
}

// DimStyle returns the universal dim helper used for placeholder copy and
// decorative fill. Centralises the ColorRule2 lookup so swapping rails on
// theme changes is a single edit.
func DimStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tui.ColorRule2)
}
