package initflow

import (
	"strings"
)

// Min dimensions below which every stage shows a "please enlarge terminal"
// floor panel instead of attempting to lay out its body. Below these
// thresholds the persistent chrome + any of the four stage bodies cannot
// render without clipping vital elements (tag column, underline, CTA),
// so we short-circuit at renderFrame rather than painting a broken frame.
const (
	MinTerminalWidth  = 70
	MinTerminalHeight = 20
)

// TerminalTooSmall reports whether the given dims fall below the floor.
// width=0 / height=0 (pre-WindowSizeMsg) is treated as unknown-but-OK —
// the very first frame renders at the hard-coded 80x24 default, then the
// stage recomputes on the first real size msg.
func TerminalTooSmall(width, height int) bool {
	if width <= 0 || height <= 0 {
		return false
	}
	return width < MinTerminalWidth || height < MinTerminalHeight
}

// ClampColumns returns the per-column cell budget for the Branches ability
// row given the available row width. Layout order (left→right):
//
//	[border 2] [glyph 1] [sp 1] [name W] [sp 1] [desc W] [sp 1] [tag W]
//
// Fixed overhead per row = 6 cells (border + glyph + three gaps). The
// remaining budget is split across three columns with these rules:
//
//   - tagW is pinned at 12 — "(required)" is 10 cells and DEFAULT is 7;
//     12 keeps a 2-cell right-margin so tags never hug the edge.
//   - nameW caps at 24 (longest display name today is "Issue To
//     Implementation" at 22 cells); shrinks proportionally below that when
//     the row can't spare 24.
//   - descW absorbs the remainder. Floor at 20 — below that, the caller
//     drops desc entirely and renders name + tag only (see renderRow).
//
// Invariant: `nameW + descW + tagW + 6 <= availableWidth` (strict ≤).
//
// Regression anchor: ClampColumns(120) returns (24, 44, 12) — the row
// widths shipped in Phase 4 polish for a typical 120-col terminal.
func ClampColumns(availableWidth int) (nameW, descW, tagW int) {
	// 84-cell target budget per Phase 4 polish: 24 + 44 + 12 + 6 overhead.
	const overhead = 6
	const maxName = 24
	const maxDesc = 44
	const pinTag = 12
	const floorDesc = 20

	// Pathologically narrow — below the min-size floor, the stage body
	// won't render anyway; return zeros so the caller can short-circuit.
	if availableWidth < overhead+pinTag {
		return 0, 0, 0
	}

	tagW = pinTag

	// Available for name+desc after overhead and tag.
	rem := availableWidth - overhead - tagW
	if rem <= 0 {
		return 0, 0, tagW
	}

	// Prefer the max-width layout when there's room for both.
	if rem >= maxName+maxDesc {
		return maxName, maxDesc, tagW
	}

	// Shrink desc first — it's the most compressible column because we
	// truncate copy with an ellipsis. Keep name at maxName if the budget
	// still admits it plus the desc floor.
	if rem >= maxName+floorDesc {
		return maxName, rem - maxName, tagW
	}

	// Tight — allocate name proportionally (1/3 of remaining), floor at 12.
	nameW = rem / 3
	if nameW < 12 {
		nameW = 12
	}
	if nameW > maxName {
		nameW = maxName
	}
	descW = rem - nameW
	if descW < floorDesc {
		// Drop desc entirely — caller renders name+tag only. Return 0 so
		// the caller can detect this mode.
		descW = 0
	}
	return nameW, descW, tagW
}

// Viewport is a minimal hand-rolled vertical scroll: holds a slice of
// pre-rendered lines and an offset. Focus-follows-cursor: caller supplies
// the focused-line index and viewport clamps offset so that line sits
// within the visible window.
//
// Rationale: matches Soil's hand-roll precedent (no bubbles/list) and
// sidesteps the bubbles/viewport dependency for ~60 LoC of exact control.
type Viewport struct {
	lines  []string
	offset int
	height int
}

// SetLines replaces the rendered line slice. Offset is clamped to the new
// bounds so SetLines after a content change doesn't leave the viewport
// scrolled past the end.
func (v *Viewport) SetLines(lines []string) {
	v.lines = lines
	v.clamp()
}

// SetHeight sets the visible window height. A height ≤ 0 collapses the
// viewport to a single line — defensive against tiny terminals.
func (v *Viewport) SetHeight(h int) {
	if h < 1 {
		h = 1
	}
	v.height = h
	v.clamp()
}

// Follow adjusts offset so the line at focusIdx is visible inside the
// current window. Silent no-op when lines or height are unset.
func (v *Viewport) Follow(focusIdx int) {
	if len(v.lines) == 0 || v.height <= 0 {
		return
	}
	if focusIdx < 0 {
		focusIdx = 0
	}
	if focusIdx >= len(v.lines) {
		focusIdx = len(v.lines) - 1
	}
	if focusIdx < v.offset {
		v.offset = focusIdx
	} else if focusIdx >= v.offset+v.height {
		v.offset = focusIdx - v.height + 1
	}
	v.clamp()
}

// View returns the visible slice joined by newlines. Never exceeds height
// rows; never returns the full slice when height < len(lines).
func (v *Viewport) View() string {
	if len(v.lines) == 0 || v.height <= 0 {
		return ""
	}
	end := v.offset + v.height
	if end > len(v.lines) {
		end = len(v.lines)
	}
	start := v.offset
	if start < 0 {
		start = 0
	}
	if start > end {
		start = end
	}
	return strings.Join(v.lines[start:end], "\n")
}

// Offset returns the current scroll offset — used by tests to assert
// Follow() behaviour without poking the unexported field.
func (v *Viewport) Offset() int { return v.offset }

// ScrollBy shifts the viewport offset by delta (positive = down). Clamps
// inside the valid range. Used by stages that expose explicit keyboard
// scroll (Planted's WRITTEN tree, Observe's PLANTING tree).
func (v *Viewport) ScrollBy(delta int) {
	v.offset += delta
	v.clamp()
}

// HasMore reports whether there are lines above or below the current
// viewport window. Callers use this to decide whether to surface "more"
// indicators or scroll key hints.
func (v *Viewport) HasMore() (up, down bool) {
	if len(v.lines) == 0 || v.height <= 0 {
		return false, false
	}
	return v.offset > 0, v.offset+v.height < len(v.lines)
}

// clamp bounds offset to [0, max(0, len(lines)-height)].
func (v *Viewport) clamp() {
	if len(v.lines) == 0 || v.height <= 0 {
		v.offset = 0
		return
	}
	maxOffset := len(v.lines) - v.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if v.offset > maxOffset {
		v.offset = maxOffset
	}
	if v.offset < 0 {
		v.offset = 0
	}
}
