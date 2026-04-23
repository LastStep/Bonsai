package guideflow

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

// defaultRenderWidth is the clamp applied when the caller passes a
// zero or negative width. Matches glamour's own default sensibility
// so narrow-terminal callers still get readable output rather than
// a single long line.
const defaultRenderWidth = 80

// renderMarkdown strips an optional YAML frontmatter block from
// content (fenced by leading/trailing `---` lines) and renders the
// remainder through glamour's auto-style terminal renderer.
//
// width drives glamour's word-wrap. Values ≤ 0 clamp to
// defaultRenderWidth so the viewer can call this before the first
// WindowSizeMsg lands without producing a broken render. Empty
// content returns an empty string and no error — glamour handles
// the empty case gracefully and callers (tab cache warmup) benefit
// from not having to special-case it.
func renderMarkdown(content string, width int) (string, error) {
	if width <= 0 {
		width = defaultRenderWidth
	}
	body := stripFrontmatter(content)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", fmt.Errorf("guideflow: create renderer: %w", err)
	}
	out, err := renderer.Render(body)
	if err != nil {
		return "", fmt.Errorf("guideflow: render markdown: %w", err)
	}
	return out, nil
}

// stripFrontmatter removes a leading YAML frontmatter block from
// content, bounded by the first two `---` delimiter lines. Inputs
// without a frontmatter block (no leading `---`) are returned
// unchanged. Malformed blocks (leading `---` but no closing
// delimiter) are also returned unchanged so the renderer has
// something to work with instead of an empty string.
//
// The trim treats the delimiter strictly — it must appear on a
// line of its own (leading `---` followed by a newline). A stray
// `---` embedded in prose mid-file is not a frontmatter boundary
// and will not trigger the strip.
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---\n") && content != "---" && !strings.HasPrefix(content, "---\r\n") {
		// Not a frontmatter block — return as-is. The `content == "---"`
		// guard is pedantic but keeps the function symmetric for the
		// degenerate single-line input.
		return content
	}
	// Find the end of the first delimiter line so we know where to
	// start searching for the closer.
	firstNL := strings.Index(content, "\n")
	if firstNL < 0 {
		return content
	}
	rest := content[firstNL+1:]
	// Closing `---` must start a line — search for "\n---" and
	// handle the edge case where the closer is at offset 0 (empty
	// frontmatter body).
	var closeIdx int
	if strings.HasPrefix(rest, "---\n") || strings.HasPrefix(rest, "---\r\n") || rest == "---" {
		closeIdx = 0
	} else {
		nl := strings.Index(rest, "\n---")
		if nl < 0 {
			return content
		}
		closeIdx = nl + 1 // skip the preceding \n so closeIdx points at `---`
	}
	// Advance past the closer's own line.
	tail := rest[closeIdx:]
	tailNL := strings.Index(tail, "\n")
	if tailNL < 0 {
		// Frontmatter closer is the last line of the file with no
		// trailing newline — return empty body.
		return ""
	}
	return strings.TrimLeft(tail[tailNL+1:], "\n")
}
