package initflow

import (
	"strings"
	"testing"
)

// TestClampColumns_Regression locks the 120-col anchor — the row widths
// shipped in Phase 4 polish must be preserved so typical-terminal rendering
// doesn't regress when the responsive math changes.
func TestClampColumns_Regression(t *testing.T) {
	name, desc, tag := ClampColumns(120)
	if name != 24 || desc != 44 || tag != 12 {
		t.Fatalf("ClampColumns(120) = (%d, %d, %d), want (24, 44, 12)", name, desc, tag)
	}
}

// TestClampColumns_InvariantFits verifies nameW+descW+tagW+6 <= availableWidth
// at every width from the min-size floor upward. The +6 accounts for the
// fixed overhead (border 2 + glyph 1 + three gaps = 6 cells).
func TestClampColumns_InvariantFits(t *testing.T) {
	for w := 40; w <= 200; w++ {
		name, desc, tag := ClampColumns(w)
		if total := name + desc + tag + 6; total > w {
			t.Errorf("ClampColumns(%d) = (%d, %d, %d) — total %d exceeds available %d",
				w, name, desc, tag, total, w)
		}
	}
}

// TestClampColumns_TagPinned verifies tag stays at 12 across the full range
// above the min-size floor. Shrinking tag would hide DEFAULT / (required).
func TestClampColumns_TagPinned(t *testing.T) {
	for w := 50; w <= 200; w++ {
		_, _, tag := ClampColumns(w)
		if tag != 12 {
			t.Errorf("ClampColumns(%d) tag = %d, want 12 (pinned)", w, tag)
		}
	}
}

// TestClampColumns_NameCap verifies name never exceeds its 24-cell cap.
func TestClampColumns_NameCap(t *testing.T) {
	for w := 50; w <= 300; w++ {
		name, _, _ := ClampColumns(w)
		if name > 24 {
			t.Errorf("ClampColumns(%d) name = %d, exceeds cap 24", w, name)
		}
	}
}

// TestClampColumns_DescFloorDrop verifies desc drops to 0 when the budget
// cannot accommodate the desc floor of 20. Callers treat descW=0 as the
// "render name+tag only" signal.
func TestClampColumns_DescFloorDrop(t *testing.T) {
	// At w=40: overhead 6 + tag 12 = 18; rem=22. rem/3=7, nameFloor=12,
	// desc=22-12=10 < 20 floor → drop to 0.
	_, desc, _ := ClampColumns(40)
	if desc != 0 {
		t.Errorf("ClampColumns(40) desc = %d, want 0 (below floor)", desc)
	}
}

// TestClampColumns_DescAbsorbs verifies extra width past the 24+44 max is
// NOT pushed into descW — desc caps at 44 even on wide terminals. Keeps
// rows from sprawling past a comfortable reading width.
func TestClampColumns_DescAbsorbs(t *testing.T) {
	_, desc, _ := ClampColumns(200)
	if desc != 44 {
		t.Errorf("ClampColumns(200) desc = %d, want 44 (max-cap)", desc)
	}
}

// TestTerminalTooSmall covers the three predicate zones. width=0 / height=0
// is treated as unknown and returns false so the first paint before a
// WindowSizeMsg doesn't falsely show the floor.
func TestTerminalTooSmall(t *testing.T) {
	cases := []struct {
		w, h int
		want bool
	}{
		{0, 0, false},
		{0, 30, false},
		{120, 0, false},
		{120, 40, false},
		{70, 20, false}, // floor (inclusive)
		{69, 20, true},
		{70, 19, true},
		{40, 10, true},
	}
	for _, c := range cases {
		if got := TerminalTooSmall(c.w, c.h); got != c.want {
			t.Errorf("TerminalTooSmall(%d, %d) = %v, want %v", c.w, c.h, got, c.want)
		}
	}
}

// TestViewport_FollowClampsOffset is the required regression: the focused
// line must always be within [offset, offset+height) after Follow().
func TestViewport_FollowClampsOffset(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	v := Viewport{}
	v.SetLines(lines)
	v.SetHeight(5)

	// Focus near the top — offset stays at 0.
	v.Follow(0)
	if v.Offset() != 0 {
		t.Errorf("Follow(0) offset = %d, want 0", v.Offset())
	}
	// Focus well past the visible window — offset slides down to keep it in
	// view: focus=10, height=5 → offset=6 (focus at position height-1).
	v.Follow(10)
	if v.Offset() != 6 {
		t.Errorf("Follow(10) offset = %d, want 6", v.Offset())
	}
	// Focus back above — offset clamps up.
	v.Follow(3)
	if v.Offset() != 3 {
		t.Errorf("Follow(3) offset = %d, want 3 (focus becomes top of window)", v.Offset())
	}
	// Focus at last line — offset = len-height = 15.
	v.Follow(19)
	if v.Offset() != 15 {
		t.Errorf("Follow(19) offset = %d, want 15", v.Offset())
	}
	// Out-of-range negative — clamps to 0.
	v.Follow(-5)
	if v.Offset() != 0 {
		t.Errorf("Follow(-5) offset = %d, want 0 (neg clamp)", v.Offset())
	}
	// Out-of-range past-end — clamps to max.
	v.Follow(9999)
	if v.Offset() != 15 {
		t.Errorf("Follow(9999) offset = %d, want 15 (past-end clamp)", v.Offset())
	}
}

// TestViewport_ViewLines verifies View returns exactly `height` lines when
// the content is long enough, clipped to the window.
func TestViewport_ViewLines(t *testing.T) {
	v := Viewport{}
	v.SetLines([]string{"a", "b", "c", "d", "e", "f"})
	v.SetHeight(3)
	v.Follow(4)
	got := v.View()
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("View() returned %d lines, want 3 (height)", len(lines))
	}
	// Focus was 4 → offset=2, window = [c, d, e].
	want := []string{"c", "d", "e"}
	for i, ln := range lines {
		if ln != want[i] {
			t.Errorf("line %d = %q, want %q", i, ln, want[i])
		}
	}
}

// TestViewport_ShortContent verifies that when content fits within height,
// View returns all lines and offset stays at 0.
func TestViewport_ShortContent(t *testing.T) {
	v := Viewport{}
	v.SetLines([]string{"a", "b"})
	v.SetHeight(10)
	v.Follow(1)
	if v.Offset() != 0 {
		t.Errorf("short content offset = %d, want 0", v.Offset())
	}
	if got := v.View(); got != "a\nb" {
		t.Errorf("View() = %q, want %q", got, "a\nb")
	}
}

// TestViewport_SetLinesClampsOffset verifies shrinking the line slice pulls
// a past-end offset back into range.
func TestViewport_SetLinesClampsOffset(t *testing.T) {
	v := Viewport{}
	v.SetLines(make([]string, 50))
	v.SetHeight(5)
	v.Follow(45) // offset=41
	v.SetLines([]string{"a", "b", "c"})
	if v.Offset() != 0 {
		t.Errorf("after shrink offset = %d, want 0", v.Offset())
	}
}
