package catalogflow

import (
	"strings"
	"testing"
)

// TestRenderEntry_NameAppears verifies the display name is rendered in
// both focused and unfocused states.
func TestRenderEntry_NameAppears(t *testing.T) {
	e := Entry{Name: "foo", DisplayName: "Foo Bar", Description: "a thing"}
	for _, focused := range []bool{true, false} {
		out := renderEntry(e, focused)
		if !strings.Contains(out, "Foo Bar") {
			t.Fatalf("focused=%v: display name missing, got:\n%s", focused, out)
		}
	}
}

// TestRenderEntry_FallsBackToName verifies an empty DisplayName falls
// back to the machine Name field.
func TestRenderEntry_FallsBackToName(t *testing.T) {
	e := Entry{Name: "foo-bar", Description: "a thing"}
	out := renderEntry(e, false)
	if !strings.Contains(out, "foo-bar") {
		t.Fatalf("expected machine name fallback, got:\n%s", out)
	}
}

// TestRenderEntry_RequiredGlyph verifies entries with a non-empty
// Required field get a trailing "*" after the name.
func TestRenderEntry_RequiredGlyph(t *testing.T) {
	e := Entry{Name: "foo", DisplayName: "Foo", Required: "all"}
	out := renderEntry(e, false)
	if !strings.Contains(out, "*") {
		t.Fatalf("expected required glyph after name, got:\n%s", out)
	}
}

// TestRenderEntry_FocusBorder verifies the focused row carries the
// leaf `│ ` border prefix while the unfocused row is plain-padded.
func TestRenderEntry_FocusBorder(t *testing.T) {
	e := Entry{Name: "foo", DisplayName: "Foo", Description: "thing"}
	focused := renderEntry(e, true)
	unfocused := renderEntry(e, false)
	if !strings.Contains(focused, "│") {
		t.Fatalf("focused row missing leaf border, got:\n%s", focused)
	}
	if strings.Contains(unfocused, "│") {
		t.Fatalf("unfocused row must not carry leaf border, got:\n%s", unfocused)
	}
}

// TestRenderDetailsBlock_RendersMetaKeys verifies that Meta entries
// appear as labelled rows in sorted order.
func TestRenderDetailsBlock_RendersMetaKeys(t *testing.T) {
	e := Entry{
		Name: "foo", DisplayName: "Foo", Description: "thing",
		Agents: "all", Required: "all",
		Meta: map[string]string{"Event": "Stop", "Matcher": "Edit"},
	}
	out := renderDetailsBlock(e, 80)
	for _, want := range []string{"AGENTS", "REQUIRED", "EVENT", "MATCHER", "Stop", "Edit"} {
		if !strings.Contains(out, want) {
			t.Fatalf("details block missing %q, got:\n%s", want, out)
		}
	}
}

// TestRenderDetailsBlock_MinRowsWithoutMeta verifies that an entry
// with no metadata renders the header + the "(no extra metadata)"
// placeholder so the block always reads as a labelled section.
func TestRenderDetailsBlock_MinRowsWithoutMeta(t *testing.T) {
	e := Entry{Name: "bare", DisplayName: "Bare"}
	out := renderDetailsBlock(e, 80)
	if !strings.Contains(out, "DETAILS") {
		t.Fatalf("header missing, got:\n%s", out)
	}
	if !strings.Contains(out, "no extra metadata") {
		t.Fatalf("expected fallback copy for bare entry, got:\n%s", out)
	}
}
