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

// RenderHeader renders the two-column top banner shown above every stage.
//
//	Left:  [盆] BONSAI · INITIALIZE · v<version>
//	Right: PLANTING INTO
//	       ~/.../<project>/<station>
//
// version "dev" / "" hides the version segment. projectDir is the absolute
// path to the project root; stationSubdir is the subdir under it (e.g.
// "station/"). safe gates the single wide-char glyph so ASCII-only terminals
// get a safe substitute.
func RenderHeader(version, projectDir, stationSubdir string, width int, safe bool) string {
	if width <= 0 {
		width = 80
	}

	primary := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)

	// ── Left block ───────────────────────────────────────────────────
	mark := "盆"
	if !safe {
		mark = "o"
	}
	leftParts := []string{
		muted.Render("[") + primary.Render(mark) + muted.Render("]"),
		primary.Render("BONSAI"),
		muted.Render("·"),
		muted.Render("INITIALIZE"),
	}
	if version != "" && version != "dev" {
		leftParts = append(leftParts,
			muted.Render("·"),
			muted.Render("v"+version),
		)
	}
	left := strings.Join(leftParts, " ")

	// ── Right block ──────────────────────────────────────────────────
	// "PLANTING INTO" headline above "~/path/<project>/<station>"
	projectDisplay := collapseHome(projectDir)
	projectName := filepath.Base(projectDir)
	parent := filepath.Dir(projectDisplay)
	// Render parent muted, project name bark, station leaf.
	station := stationSubdir
	if station != "" && !strings.HasSuffix(station, "/") {
		station += "/"
	}
	if parent == "." || parent == "" {
		parent = ""
	} else if !strings.HasSuffix(parent, "/") {
		parent += "/"
	}
	pathRow := muted.Render(parent) + bark.Render(projectName) + muted.Render("/") + leaf.Render(station)
	rightRow1 := muted.Render("PLANTING INTO")

	// ── Compose two-row layout with left-padded right block ─────────
	// Row 1: left + spaces + rightRow1
	// Row 2: spaces (matching left width) + pathRow
	leftW := lipgloss.Width(left)
	right1W := lipgloss.Width(rightRow1)
	right2W := lipgloss.Width(pathRow)
	// Pick the wider right side to use as the target anchor so both rows
	// right-align consistently.
	rightW := right1W
	if right2W > rightW {
		rightW = right2W
	}
	gap := width - leftW - rightW - 2 // -2 for 1-col padding each side
	if gap < 2 {
		gap = 2
	}
	// Right-pad row 1 / row 2 so their right columns align.
	row1 := left + strings.Repeat(" ", gap) + strings.Repeat(" ", rightW-right1W) + rightRow1
	row2Left := strings.Repeat(" ", leftW+gap)
	row2 := row2Left + strings.Repeat(" ", rightW-right2W) + pathRow

	// Pad both rows to width so AltScreen doesn't see ragged edges.
	row1 = padRight(row1, width)
	row2 = padRight(row2, width)
	return row1 + "\n" + row2
}

// RenderFooter draws the bottom hairline row: brand on the left, the passed
// key hints on the right, separated by a muted dot.
//
//	一 BONSAI 一               ↵ continue · esc back · ctrl-c quit
func RenderFooter(hints []KeyHint, width int) string {
	if width <= 0 {
		width = 80
	}

	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	primary := lipgloss.NewStyle().Foreground(tui.ColorPrimary)

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
	return padRight(row, width)
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
