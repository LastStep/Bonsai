package initflow

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
)

// KeyHint is a single entry in the footer key row: one glyph/token plus a
// short descriptor. Kept as a small value type so callers can build
// []KeyHint slices inline from each stage's View().
type KeyHint struct {
	Key  string // e.g. "↵", "␣", "?", "esc"
	Desc string // e.g. "continue", "toggle", "details"
}

// RenderHeader renders the two-column, two-row top banner shown above every
// stage. Both columns are two rows — left stacks the service badge above the
// process label so the brand and the action sit on distinct lines; right
// stacks "PLANTING INTO" above the project path so the destination reads as
// its own block.
//
//	Left row 1:  [盆] BONSAI
//	Left row 2:  INITIALIZE · v<version>
//	Right row 1: PLANTING INTO
//	Right row 2: ~/.../<project>/
//
// version "dev" / "" hides the version segment on row 2. projectDir is the
// absolute path to the project root — the only path segment rendered in the
// right block; earlier iterations also rendered a "station/" suffix, but the
// station subdir doesn't exist yet at any point before the Generate stage,
// so showing it was misleading. safe gates the single wide-char glyph so
// ASCII-only terminals get a safe substitute.
func RenderHeader(version, projectDir string, width int, safe bool) string {
	if width <= 0 {
		width = 80
	}

	primary := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary)

	// ── Left block (2 rows) ─────────────────────────────────────────
	mark := "盆"
	if !safe {
		mark = "o"
	}
	// Pad inside brackets so the wide-char kanji (and its ASCII fallback)
	// reads as visually centred rather than left-hugging the `[` — terminals
	// left-anchor CJK glyphs inside their 2-cell slot, so "[盆]" looks off.
	leftRow1 := muted.Render("[ ") + primary.Render(mark) + muted.Render(" ]") +
		" " + primary.Render("BONSAI")

	leftRow2Parts := []string{muted.Render("INITIALIZE")}
	if version != "" && version != "dev" {
		leftRow2Parts = append(leftRow2Parts,
			muted.Render("·"),
			muted.Render("v"+version),
		)
	}
	leftRow2 := strings.Join(leftRow2Parts, " ")

	// ── Right block (2 rows) ────────────────────────────────────────
	projectDisplay := collapseHome(projectDir)
	projectName := filepath.Base(projectDir)
	parent := filepath.Dir(projectDisplay)
	if parent == "." || parent == "" {
		parent = ""
	} else if !strings.HasSuffix(parent, "/") {
		parent += "/"
	}
	rightRow1 := muted.Render("PLANTING INTO")
	rightRow2 := muted.Render(parent) + bark.Render(projectName) + muted.Render("/")

	// ── Compose ─────────────────────────────────────────────────────
	// Pick the wider of each block as the column anchor so both rows
	// align consistently on the shared gap.
	left1W := lipgloss.Width(leftRow1)
	left2W := lipgloss.Width(leftRow2)
	leftW := left1W
	if left2W > leftW {
		leftW = left2W
	}
	right1W := lipgloss.Width(rightRow1)
	right2W := lipgloss.Width(rightRow2)
	rightW := right1W
	if right2W > rightW {
		rightW = right2W
	}
	gap := width - leftW - rightW - 2 // -2 for 1-col padding each side
	if gap < 2 {
		gap = 2
	}

	row1 := leftRow1 + strings.Repeat(" ", leftW-left1W+gap) +
		strings.Repeat(" ", rightW-right1W) + rightRow1
	row2 := leftRow2 + strings.Repeat(" ", leftW-left2W+gap) +
		strings.Repeat(" ", rightW-right2W) + rightRow2

	row1 = padRight(row1, width)
	row2 = padRight(row2, width)
	return row1 + "\n" + row2
}

// RenderFooter draws a two-row bottom unit: a subtle full-width rule
// (ColorRule2) above the brand/hints row. The rule gives the footer a clean
// visual separation from the stage body per design.
//
//	──────────────────────────────────────────────────────────────────
//	一 BONSAI 一               ↵ continue · esc back · ctrl-c quit
func RenderFooter(hints []KeyHint, width int) string {
	if width <= 0 {
		width = 80
	}

	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	primary := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	ruleStyle := lipgloss.NewStyle().Foreground(tui.ColorRule2)

	rule := ruleStyle.Render(strings.Repeat("─", width))

	left := muted.Render("一 ") + primary.Render("BONSAI") + muted.Render(" 一")
	// Render each hint as "<KEY> desc" with the key in muted-emphasis.
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		key := muted.Render(h.Key)
		desc := muted.Render(h.Desc)
		parts = append(parts, key+" "+desc)
	}
	sep := muted.Render("  " + tui.GlyphDot + "  ")
	right := strings.Join(parts, sep)

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	gap := width - leftW - rightW - 2
	if gap < 2 {
		gap = 2
	}
	row := " " + left + strings.Repeat(" ", gap) + right + " "
	return rule + "\n" + padRight(row, width)
}

// RenderMinSizeFloor renders a centred "please enlarge terminal" panel
// shown in place of any stage body when TerminalTooSmall reports true.
// The frame never attempts the persistent chrome at tiny dims — header /
// footer / rail would themselves clip below the floor. Instead the whole
// AltScreen is filled with a single centred block carrying brand + hint
// + current dims so the user knows they hit the floor.
//
// width/height are the live terminal dims (as seen by the stage). The
// rendered output always occupies `height` rows so AltScreen doesn't leave
// stale content below the panel.
func RenderMinSizeFloor(width, height int) string {
	if width <= 0 {
		width = 40
	}
	if height <= 0 {
		height = 10
	}

	primary := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)

	title := primary.Render("BONSAI")
	subtitle := muted.Render("please enlarge your terminal")
	hint := dim.Render(
		"minimum " + itoa(MinTerminalWidth) + " × " + itoa(MinTerminalHeight) +
			"   ·   current " + itoa(width) + " × " + itoa(height),
	)

	lines := []string{title, "", subtitle, hint}

	// Measure the widest line for centering.
	maxW := 0
	for _, l := range lines {
		if w := lipgloss.Width(l); w > maxW {
			maxW = w
		}
	}
	leftPad := (width - maxW) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	prefix := strings.Repeat(" ", leftPad)
	rendered := make([]string, len(lines))
	for i, l := range lines {
		if l == "" {
			rendered[i] = ""
		} else {
			rendered[i] = prefix + l
		}
	}

	// Vertically centre inside height.
	topPad := (height - len(lines)) / 2
	if topPad < 0 {
		topPad = 0
	}
	bottomPad := height - topPad - len(lines)
	if bottomPad < 0 {
		bottomPad = 0
	}
	top := strings.Repeat("\n", topPad)
	bottom := strings.Repeat("\n", bottomPad)
	return top + strings.Join(rendered, "\n") + bottom
}

// centerBlock left-pads every line in block so the widest line is
// horizontally centred inside width. Used by stage bodies to sit visually
// balanced inside the AltScreen rather than flush-left. Trailing whitespace
// on each line is preserved because focused-input underlines rely on exact
// column width.
func centerBlock(block string, width int) string {
	if width <= 0 {
		return block
	}
	lines := strings.Split(block, "\n")
	maxW := 0
	for _, l := range lines {
		if w := lipgloss.Width(l); w > maxW {
			maxW = w
		}
	}
	pad := (width - maxW) / 2
	if pad < 2 {
		pad = 2
	}
	prefix := strings.Repeat(" ", pad)
	out := make([]string, len(lines))
	for i, l := range lines {
		if l == "" {
			out[i] = ""
			continue
		}
		out[i] = prefix + l
	}
	return strings.Join(out, "\n")
}

// padRight right-pads s with spaces so its visible width reaches w.
func padRight(s string, w int) string {
	cur := lipgloss.Width(s)
	if cur >= w {
		return s
	}
	return s + strings.Repeat(" ", w-cur)
}

// collapseHome replaces a leading $HOME prefix with "~" for display. Fails
// safely by returning the original path if the lookup errors.
func collapseHome(abs string) string {
	home, err := homeDir()
	if err != nil || home == "" {
		return abs
	}
	if strings.HasPrefix(abs, home) {
		return "~" + abs[len(home):]
	}
	return abs
}
