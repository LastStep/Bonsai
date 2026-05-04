package generate

import (
	"fmt"
	"strings"

	"github.com/LastStep/Bonsai/internal/config"
	"gopkg.in/yaml.v3"
)

// ParseFrontmatter extracts YAML frontmatter from file content.
// Returns nil if no frontmatter is found.
//
// Two delimiter styles are accepted:
//
//  1. Markdown-style — content begins with `---\n` (or `---\r\n`),
//     closes with `\n---`. Used by .md ability files.
//  2. Bash-comment style — content optionally begins with a `#!` shebang,
//     then a `# ---` (or `#---`) opener, comment-prefixed YAML body, and
//     a `# ---` (or `#---`) closer. Used by sensor `.sh` files where
//     byte 0 must be the shebang for the file to be executable.
func ParseFrontmatter(data []byte) (*config.CustomItemMeta, error) {
	content := string(data)

	// Markdown-style fast path — keeps the existing behaviour for .md
	// files unchanged.
	if strings.HasPrefix(content, "---\n") || strings.HasPrefix(content, "---\r\n") {
		return parseMarkdownFrontmatter(content)
	}

	// Bash-comment path — used for sensor .sh files.
	return parseBashCommentFrontmatter(content)
}

// parseMarkdownFrontmatter handles the `---\n...---\n` delimiter style.
func parseMarkdownFrontmatter(content string) (*config.CustomItemMeta, error) {
	// Skip opening delimiter (handle both \n and \r\n line endings)
	skip := strings.Index(content, "\n") + 1
	rest := content[skip:]

	// Find closing delimiter
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return nil, fmt.Errorf("unterminated frontmatter")
	}

	fmContent := strings.ReplaceAll(rest[:end], "\r", "")

	var meta config.CustomItemMeta
	if err := yaml.Unmarshal([]byte(fmContent), &meta); err != nil {
		return nil, fmt.Errorf("invalid frontmatter YAML: %w", err)
	}

	return &meta, nil
}

// parseBashCommentFrontmatter handles `# ---` opener / `# ---` closer
// with comment-prefixed YAML body. Optional `#!` shebang at byte 0 is
// skipped. Returns an error when no opener is found or no closer matches.
func parseBashCommentFrontmatter(content string) (*config.CustomItemMeta, error) {
	// Normalise line endings for line-by-line scanning. The original
	// content is not mutated in place — we operate on a slice of lines.
	normalised := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalised, "\n")

	idx := 0

	// Optional shebang on line 0.
	if idx < len(lines) && strings.HasPrefix(lines[idx], "#!") {
		idx++
	}

	// Skip blank lines until the opener.
	for idx < len(lines) && strings.TrimSpace(lines[idx]) == "" {
		idx++
	}

	if idx >= len(lines) || !isBashFrontmatterDelim(lines[idx]) {
		return nil, fmt.Errorf("no frontmatter found")
	}
	idx++ // past opener

	// Read body until closer.
	var body []string
	closed := false
	for ; idx < len(lines); idx++ {
		line := lines[idx]
		if isBashFrontmatterDelim(line) {
			closed = true
			break
		}
		body = append(body, stripBashCommentPrefix(line))
	}

	if !closed {
		return nil, fmt.Errorf("unterminated frontmatter")
	}

	var meta config.CustomItemMeta
	if err := yaml.Unmarshal([]byte(strings.Join(body, "\n")), &meta); err != nil {
		return nil, fmt.Errorf("invalid frontmatter YAML: %w", err)
	}
	return &meta, nil
}

// isBashFrontmatterDelim reports whether line is `# ---` or `#---`
// (trimmed). Trailing whitespace is tolerated.
func isBashFrontmatterDelim(line string) bool {
	t := strings.TrimRight(line, " \t")
	return t == "# ---" || t == "#---"
}

// stripBashCommentPrefix removes the leading `# ` or `#` from a comment
// line. Lines without a leading `#` are returned as-is so YAML parsing
// can flag them rather than this function silently swallowing the issue.
func stripBashCommentPrefix(line string) string {
	if strings.HasPrefix(line, "# ") {
		return line[2:]
	}
	if strings.HasPrefix(line, "#") {
		return line[1:]
	}
	return line
}
