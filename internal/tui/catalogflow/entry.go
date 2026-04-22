package catalogflow

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// renderEntry renders a single catalog row. Focused rows receive the
// leaf-coloured `│ ` left border (same pattern as initflow's
// BranchesStage.renderRow) and white-bold name styling;
// unfocused rows render muted.
//
// The row shape:
//
//	<border> <name>  <description>
//
// Required entries (Required != "") get a trailing gold "*" glyph
// after the name — matches the required-marker convention used by
// Soil and Branches in the init flow.
func renderEntry(e Entry, focused bool) string {
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)

	border := initflow.UnfocusBorder()
	if focused {
		border = initflow.FocusBorder()
	}

	name := e.DisplayName
	if name == "" {
		name = e.Name
	}

	// Name style: white-bold on focus, subtle otherwise. Description
	// lifts from dim to accent-white on focus so the whole row reads as
	// a single bright block under the cursor.
	var nameStyle lipgloss.Style
	var descStyle lipgloss.Style
	if focused {
		nameStyle = initflow.FocusedNameStyle()
		descStyle = initflow.FocusedDescStyle()
	} else {
		nameStyle = initflow.UnfocusedNameStyle()
		descStyle = dim
	}

	nameText := nameStyle.Render(name)
	if e.Required != "" {
		nameText += " " + initflow.RequiredGlyph()
	}

	// Description clamped to a soft budget — browser rows are wider than
	// the Branches list (no tag column), so we allow up to 60 cells of
	// description before truncating.
	const descBudget = 60
	desc := e.Description
	if lipgloss.Width(desc) > descBudget {
		rr := []rune(desc)
		if len(rr) > descBudget-1 {
			desc = string(rr[:descBudget-1]) + "…"
		}
	}

	row := border + nameText
	if desc != "" {
		row += "  " + descStyle.Render(desc)
	}
	return row
}

// renderDetailsBlock renders the inline-expand block shown when `?`
// is toggled on. Labels render bark-gold bold (LabelStyle); values
// render accent-white (ValueStyle). Rendered fields, in order:
//
//   - Agents      — always shown when set.
//   - Required    — shown when non-empty.
//   - Meta keys   — rendered in sorted order so the output is stable
//     across runs (map iteration is unordered in Go).
//
// panelW is the effective width of the details panel. Rows are
// tail-truncated with an ellipsis to keep the block from wrapping.
func renderDetailsBlock(e Entry, panelW int) string {
	label := initflow.LabelStyle()
	value := lipgloss.NewStyle().Foreground(tui.ColorAccent)

	header := initflow.RenderSectionHeader("DETAILS", panelW)

	const labelW = 12
	const indent = "  "

	rows := []string{header}
	add := func(key, val string) {
		if val == "" {
			return
		}
		// Tail-truncate long values so the DETAILS block never wraps.
		budget := panelW - labelW - len(indent) - 2
		if budget < 10 {
			budget = 10
		}
		if lipgloss.Width(val) > budget {
			rr := []rune(val)
			if len(rr) > budget-1 {
				val = string(rr[:budget-1]) + "…"
			}
		}
		rows = append(rows,
			indent+label.Render(padRight(key, labelW))+value.Render(val))
	}

	if e.Agents != "" {
		add("AGENTS", e.Agents)
	}
	if e.Required != "" {
		add("REQUIRED", e.Required)
	}

	// Sort Meta keys so output is deterministic.
	metaKeys := make([]string, 0, len(e.Meta))
	for k := range e.Meta {
		metaKeys = append(metaKeys, k)
	}
	sort.Strings(metaKeys)
	for _, k := range metaKeys {
		add(strings.ToUpper(k), e.Meta[k])
	}

	if len(rows) == 1 {
		// Nothing to show beyond the header.
		dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
		rows = append(rows, indent+dim.Render("(no extra metadata)"))
	}

	return strings.Join(rows, "\n")
}

// padRight right-pads s with spaces so its visible width reaches w.
// Kept local to catalogflow so the package stays self-contained without
// leaking into initflow's internal helper namespace.
func padRight(s string, w int) string {
	cur := lipgloss.Width(s)
	if cur >= w {
		return s
	}
	return s + strings.Repeat(" ", w-cur)
}
