package generate

import (
	"fmt"
	"strings"

	"github.com/LastStep/Bonsai/internal/config"
	"gopkg.in/yaml.v3"
)

// ParseFrontmatter extracts YAML frontmatter from file content.
// Returns nil if no frontmatter is found.
func ParseFrontmatter(data []byte) (*config.CustomItemMeta, error) {
	content := string(data)
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return nil, fmt.Errorf("no frontmatter found")
	}

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
