package initflow

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
)

// Enso progress glyphs — referenced from Plan 22 decision Q5.
const (
	ensoDone    = "●"
	ensoPending = "○"
	railChar    = "─"
)

// RenderEnsoRail draws the N-stage progress rail used above every stage's
// body. Rendered as two rows:
//
//	row 1: ● ─── ● ─── [枝] ─── ○        (dots + connector segments)
//	row 2: VESSEL  SOIL  BRANCHES  OBSERVE   (stage labels)
//
// When safe=false (WideCharSafe() reported false or BONSAI_ASCII_ONLY=1 is
// set), the dots collapse to bracketed ASCII: [x] for done, [ ] for
// pending, [N] for current.
//
// labels is the stage-label slice to render against. When nil or empty,
// the function falls back to the package-level StageLabels (the init flow's
// 4-stage set). Other flows (e.g. addflow's 6-stage set) pass their own
// slice so the rail adapts to the active flow's length without reimplementing
// this primitive.
//
// stageIdx is the 0-based current stage. width is the terminal column count
// the rail should occupy. Layout tries to centre the rail inside width; the
// dot/bracket glyphs sit at fixed anchors with equal-length connector runs
// between them. Labels are centred under each anchor.
func RenderEnsoRail(stageIdx int, labels []StageLabel, width int, safe bool) string {
	if len(labels) == 0 {
		labels = StageLabels[:]
	}
	if width <= 0 {
		width = 80
	}

	// Clamp index to the valid range so callers don't have to guard.
	if stageIdx < 0 {
		stageIdx = 0
	}
	if stageIdx > len(labels)-1 {
		stageIdx = len(labels) - 1
	}

	// Glyph + styled rendering for each anchor and connector.
	numStages := len(labels)
	anchors := make([]string, numStages)
	anchorWidths := make([]int, numStages)
	for i := range labels {
		glyph, w := anchorGlyph(i, stageIdx, labels, safe)
		anchors[i] = glyph
		anchorWidths[i] = w
	}

	// Cap the rail's visible width so the checkpoints sit tight and centred
	// rather than stretching the full terminal. maxRail is the target total
	// width (anchors + connectors). When the terminal is narrower than
	// maxRail we fall back to filling the terminal minus sidePad.
	const maxRail = 60
	const minConn = 3
	const sidePad = 2
	sumAnchor := 0
	for _, w := range anchorWidths {
		sumAnchor += w
	}
	connectors := numStages - 1
	// Target rail width: smaller of maxRail and (width - 2*sidePad).
	target := width - sidePad*2
	if target > maxRail {
		target = maxRail
	}
	connLen := (target - sumAnchor) / connectors
	if connLen < minConn {
		connLen = minConn
	}
	railWidth := sumAnchor + connLen*connectors

	// Centre the whole rail inside `width` by left-padding.
	leftPad := (width - railWidth) / 2
	if leftPad < sidePad {
		leftPad = sidePad
	}

	// Build row 1: dot ─── dot ─── dot ─── dot
	var row1 strings.Builder
	row1.WriteString(strings.Repeat(" ", leftPad))
	for i := 0; i < numStages; i++ {
		row1.WriteString(anchors[i])
		if i < numStages-1 {
			row1.WriteString(railSegment(i, stageIdx, connLen))
		}
	}

	// Compute column positions of each anchor so row 2 labels can
	// centre underneath them.
	colPositions := make([]int, numStages)
	col := leftPad
	for i := 0; i < numStages; i++ {
		colPositions[i] = col + anchorWidths[i]/2
		col += anchorWidths[i]
		if i < numStages-1 {
			col += connLen
		}
	}

	// Row 2: stage English labels, centred on each anchor.
	row2 := placeLabels(colPositions, stageLabelTexts(labels, safe, stageIdx), width, stageIdx, false)

	return row1.String() + "\n" + row2
}

// anchorGlyph returns the styled glyph (and the visible-width-count) for the
// i-th stage anchor. The current stage gets a bracketed kanji in Bark gold;
// completed stages get a bright Primary ● (user-requested green-done badge);
// pending stages get a muted ○.
func anchorGlyph(i, current int, labels []StageLabel, safe bool) (string, int) {
	goldStyle := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	doneStyle := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	muted := lipgloss.NewStyle().Foreground(tui.ColorRule2)

	if safe {
		switch {
		case i < current:
			return doneStyle.Render(ensoDone), 1
		case i == current:
			// Render kanji in a small boxed form: [K] — 4 visible cells
			// (bracket + 2-cell kanji + bracket). Kanji + brackets are Bark
			// gold so the current stage reads as the active accent.
			kanji := labels[i].Kanji
			return goldStyle.Render("[") + goldStyle.Render(kanji) + goldStyle.Render("]"), 4
		default:
			return muted.Render(ensoPending), 1
		}
	}

	// ASCII fallback: [x] done (bright green), [N] current (gold), [ ] pending.
	switch {
	case i < current:
		return muted.Render("[") + doneStyle.Render("x") + muted.Render("]"), 3
	case i == current:
		return goldStyle.Render("[") + goldStyle.Render(itoa(i+1)) + goldStyle.Render("]"), 3
	default:
		return muted.Render("[ ]"), 3
	}
}

// railSegment renders the "─────" between anchors i and i+1.
// Segments before the current stage are tinted LeafDim; segments at or
// beyond the current stage are muted Rule2.
func railSegment(i, current, length int) string {
	if length < 1 {
		length = 1
	}
	seg := strings.Repeat(railChar, length)
	if i < current {
		return lipgloss.NewStyle().Foreground(tui.ColorLeafDim).Render(seg)
	}
	return lipgloss.NewStyle().Foreground(tui.ColorRule2).Render(seg)
}

// placeLabels renders a row where each non-empty label in `labels` is
// horizontally centred on the corresponding column in `colPositions`.
// Labels are padded with spaces to fill the row up to `width`.
//
// If kana is true, the current-stage label is rendered muted. Otherwise
// the current-stage label is rendered leaf-bold while other labels are
// muted.
func placeLabels(colPositions []int, labels []string, width, current int, kana bool) string {
	leafStyle := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)

	// Build a character-indexed buffer; we overlay each label centred on
	// its anchor column. Use runes to keep multibyte labels addressable.
	buf := make([]rune, width)
	for i := range buf {
		buf[i] = ' '
	}
	for i, text := range labels {
		if text == "" {
			continue
		}
		runes := []rune(text)
		// centre the label on colPositions[i]; half of the width to the left.
		// Use runeWidth-unaware arithmetic here because the labels are
		// English ASCII ("VESSEL" etc.) OR kana rendered on safe terminals
		// where kana occupies 2 cells per char — which we compensate for by
		// treating rune count as char count; perfect centring is not
		// critical for a three-row rail.
		start := colPositions[i] - len(runes)/2
		if start < 0 {
			start = 0
		}
		for j, r := range runes {
			if start+j >= len(buf) {
				break
			}
			buf[start+j] = r
		}
	}

	// Rebuild the row, applying per-label styling by splitting on the
	// original anchor positions. Simplest: render the whole buffer muted,
	// then overlay the current label in leaf-bold by substituting at the
	// right range.
	out := string(buf)
	// Trim trailing whitespace to keep the row tidy for small widths.
	out = strings.TrimRight(out, " ")
	if kana {
		return muted.Render(out)
	}
	// For the English-label row, apply leaf-bold to the current stage only.
	if current >= 0 && current < len(labels) && labels[current] != "" {
		target := labels[current]
		idx := strings.Index(out, target)
		if idx >= 0 {
			prefix := out[:idx]
			suffix := out[idx+len(target):]
			return muted.Render(prefix) + leafStyle.Render(target) + muted.Render(suffix)
		}
	}
	return muted.Render(out)
}

// stageLabelTexts returns the English-label slice for the rail's row 2.
// Safe vs ASCII has no effect on this row — the English labels always
// render (the ASCII fallback is on the glyph row above).
func stageLabelTexts(labels []StageLabel, safe bool, current int) []string {
	_ = safe // reserved for future styling toggles; labels are uniform today
	out := make([]string, len(labels))
	for i, l := range labels {
		out[i] = l.English
	}
	_ = current
	return out
}

// itoa is a tiny stdlib-free int-to-string used only for the ASCII fallback
// rail where the current stage is rendered as "[N]". Avoids pulling strconv
// into this file just for one callsite.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
