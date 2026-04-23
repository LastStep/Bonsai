package guideflow

import (
	"strings"
	"testing"
)

// TestStripFrontmatter_NoBlock verifies content without leading
// `---` delimiters is returned unchanged.
func TestStripFrontmatter_NoBlock(t *testing.T) {
	input := "# Heading\n\nbody text\n"
	got := stripFrontmatter(input)
	if got != input {
		t.Fatalf("stripFrontmatter should be a no-op without a frontmatter block;\n got=%q\nwant=%q", got, input)
	}
}

// TestStripFrontmatter_WithBlock verifies a well-formed frontmatter
// block is stripped including the closing delimiter line.
func TestStripFrontmatter_WithBlock(t *testing.T) {
	input := "---\ndescription: demo\ntag: x\n---\n# Heading\n\nbody\n"
	got := stripFrontmatter(input)
	if strings.Contains(got, "description:") || strings.Contains(got, "tag:") {
		t.Fatalf("stripFrontmatter left YAML keys in place:\n%s", got)
	}
	if !strings.HasPrefix(got, "# Heading") {
		t.Fatalf("stripFrontmatter should leave body starting at heading;\n got=%q", got)
	}
}

// TestStripFrontmatter_Malformed verifies an unterminated
// frontmatter block (leading `---` but no closer) returns the
// original content so the renderer still has something to work
// with.
func TestStripFrontmatter_Malformed(t *testing.T) {
	input := "---\ndescription: demo\nno closer here\n"
	got := stripFrontmatter(input)
	if got != input {
		t.Fatalf("stripFrontmatter on malformed input should return original;\n got=%q\nwant=%q", got, input)
	}
}

// TestStripFrontmatter_EmptyBody verifies an empty YAML block
// (two adjacent `---` lines) strips cleanly.
func TestStripFrontmatter_EmptyBody(t *testing.T) {
	input := "---\n---\n# Heading\n"
	got := stripFrontmatter(input)
	if !strings.HasPrefix(got, "# Heading") {
		t.Fatalf("stripFrontmatter on empty frontmatter should leave heading;\n got=%q", got)
	}
}

// TestRenderMarkdown_NarrowVsWide verifies the renderer honours
// the width hint — a narrow width produces a taller render
// (more wrapped lines) than a wide one for the same long-prose
// input.
func TestRenderMarkdown_NarrowVsWide(t *testing.T) {
	input := "# Heading\n\n" + strings.Repeat("The quick brown fox jumps over the lazy dog. ", 20)
	narrow, err := renderMarkdown(input, 40)
	if err != nil {
		t.Fatalf("narrow render: %v", err)
	}
	wide, err := renderMarkdown(input, 120)
	if err != nil {
		t.Fatalf("wide render: %v", err)
	}
	narrowLines := strings.Count(narrow, "\n")
	wideLines := strings.Count(wide, "\n")
	if narrowLines <= wideLines {
		t.Fatalf("narrow render should produce more lines than wide; narrow=%d wide=%d", narrowLines, wideLines)
	}
}

// TestRenderMarkdown_ZeroWidthClamps verifies passing width=0
// doesn't error and produces non-empty output — the viewer may
// call this before its first WindowSizeMsg and should not
// explode.
func TestRenderMarkdown_ZeroWidthClamps(t *testing.T) {
	input := "# Heading\n\nsome body text.\n"
	out, err := renderMarkdown(input, 0)
	if err != nil {
		t.Fatalf("zero-width render: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatalf("zero-width render produced empty output")
	}
}

// TestRenderMarkdown_NegativeWidthClamps verifies negative widths
// also clamp instead of panicking inside glamour.
func TestRenderMarkdown_NegativeWidthClamps(t *testing.T) {
	out, err := renderMarkdown("# Heading\n", -10)
	if err != nil {
		t.Fatalf("negative-width render: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatalf("negative-width render produced empty output")
	}
}

// TestRenderMarkdown_EmptyContent verifies empty input renders
// without error. Glamour handles this gracefully; the viewer's
// cache warmup benefits from not needing a special case.
func TestRenderMarkdown_EmptyContent(t *testing.T) {
	out, err := renderMarkdown("", 80)
	if err != nil {
		t.Fatalf("empty render: %v", err)
	}
	// Empty-in → empty-ish out (glamour may emit trailing
	// whitespace). Assert no panic + no error; contents
	// immaterial.
	_ = out
}

// TestRenderMarkdown_StripsFrontmatter verifies the renderer
// integrates the frontmatter strip — YAML keys should not leak
// into the rendered output.
func TestRenderMarkdown_StripsFrontmatter(t *testing.T) {
	input := "---\ndescription: demo\n---\n# Heading\n\nbody text\n"
	out, err := renderMarkdown(input, 80)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(out, "description:") {
		t.Fatalf("rendered output leaked frontmatter YAML:\n%s", out)
	}
}
