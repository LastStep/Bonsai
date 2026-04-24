package hints

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// Render returns the 3-section hints block as a lipgloss-styled string.
// Width clamps to a reasonable range via initflow.PanelWidth; 0 / negative
// widths fall through to an empty string so callers can stack the render
// unconditionally.
//
// Layout:
//
//	─── NEXT STEPS ────────────────────
//	  $ bonsai add backend
//	  $ bonsai catalog
//
//	─── TRY THIS ───────────────────────
//	  Edit station/Playbook/Backlog.md
//
//	─── ASK YOUR AGENT ─────────────────
//	  » Start working
//	  ┌──────────────────────────────┐
//	  │ Hi, get started — read …     │
//	  └──────────────────────────────┘
//
// Each section renders only when its slice is non-empty, so partial
// hints.yaml files produce a partial block rather than empty headers.
//
// Zero-value Block (b.IsZero()) returns "" so callers that lack a hints
// source emit nothing.
func Render(b Block, width int) string {
	if b.IsZero() {
		return ""
	}
	if width <= 0 {
		return ""
	}
	panelW := width
	if panelW > initflow.PanelContentWidth {
		panelW = initflow.PanelContentWidth
	}

	sections := make([]string, 0, 3)
	if len(b.NextCLI) > 0 {
		sections = append(sections, renderCLISection(b.NextCLI, panelW))
	}
	if len(b.NextWorkflow) > 0 {
		sections = append(sections, renderWorkflowSection(b.NextWorkflow, panelW))
	}
	if len(b.AIPrompts) > 0 {
		sections = append(sections, renderPromptsSection(b.AIPrompts, panelW))
	}
	return strings.Join(sections, "\n\n")
}

// renderCLISection renders the NEXT STEPS block — each line prefixed
// with a leaf-accented numeral and the CLI text in accent-white.
func renderCLISection(lines []string, panelW int) string {
	header := initflow.RenderSectionHeader("NEXT STEPS", panelW)
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)

	rendered := []string{header}
	for i, line := range lines {
		numeral := indexToNumeral(i)
		rendered = append(rendered, "  "+bark.Render(numeral)+"  "+white.Render(line))
	}
	return strings.Join(rendered, "\n")
}

// renderWorkflowSection renders the TRY THIS block — workflow
// suggestions rendered as dim muted copy.
func renderWorkflowSection(lines []string, panelW int) string {
	header := initflow.RenderSectionHeader("TRY THIS", panelW)
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()

	rendered := []string{header}
	for _, line := range lines {
		rendered = append(rendered, "  "+bark.Render("·")+"  "+dim.Render(line))
	}
	return strings.Join(rendered, "\n")
}

// renderPromptsSection renders the ASK YOUR AGENT block — each prompt
// as a label + bordered body box so the user can select-copy from the
// terminal.
func renderPromptsSection(prompts []Prompt, panelW int) string {
	header := initflow.RenderSectionHeader("ASK YOUR AGENT", panelW)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	bark := initflow.LabelStyle()

	// Body box budget — PanelContentWidth minus the 2-cell indent + 2
	// cells of border. Clamp to a sane floor so narrow terminals still
	// render something legible.
	const indent = "  "
	boxW := panelW - 4
	if boxW < 20 {
		boxW = 20
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.ColorRule2).
		Padding(0, 1).
		Width(boxW)

	rendered := []string{header}
	for i, p := range prompts {
		labelLine := indent + leaf.Render("» ") + bark.Render(p.Label)
		if i > 0 {
			rendered = append(rendered, "")
		}
		rendered = append(rendered, labelLine)
		// Border-wrap the body so copy-paste works cleanly. Split on
		// newlines manually — lipgloss's Width + Padding already handle
		// wrapping inside the box.
		body := boxStyle.Render(p.Body)
		// Indent each line of the rendered body under the label.
		for _, line := range strings.Split(body, "\n") {
			rendered = append(rendered, indent+line)
		}
	}
	return strings.Join(rendered, "\n")
}

// indexToNumeral returns "1", "2", "3" … up to 9, then falls back to
// the string form of the 1-indexed position for larger counts. Keeps the
// numerals readable in the common 1-3 hint case.
func indexToNumeral(i int) string {
	if i < 0 {
		return "·"
	}
	if i < 9 {
		return string(rune('1' + i))
	}
	return "·"
}
