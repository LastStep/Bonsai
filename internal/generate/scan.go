package generate

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/config"
)

// DiscoveredFile represents a custom file found in a workspace that isn't tracked by Bonsai.
type DiscoveredFile struct {
	Name    string                 // kebab-case name derived from filename
	Type    string                 // "skill", "workflow", "protocol", "sensor", "routine"
	RelPath string                 // relative path from project root
	Meta    *config.CustomItemMeta // parsed frontmatter (nil if parsing failed)
	Error   string                 // validation error, if any
}

// ScanCustomFiles finds untracked custom files in an agent's workspace directories.
// Only scans top-level files (no subdirectories).
func ScanCustomFiles(projectRoot string, installed *config.InstalledAgent, lock *config.LockFile) ([]DiscoveredFile, error) {
	workspaceRoot := filepath.Join(projectRoot, installed.Workspace)
	agentDir := filepath.Join(workspaceRoot, "agent")

	type categoryDef struct {
		dir   string
		ext   string
		items []string
	}

	categories := map[string]categoryDef{
		"skill":    {dir: "Skills", ext: ".md", items: installed.Skills},
		"workflow": {dir: "Workflows", ext: ".md", items: installed.Workflows},
		"protocol": {dir: "Protocols", ext: ".md", items: installed.Protocols},
		"sensor":   {dir: "Sensors", ext: ".sh", items: installed.Sensors},
		"routine":  {dir: "Routines", ext: ".md", items: installed.Routines},
	}

	var discovered []DiscoveredFile

	for itemType, info := range categories {
		dirPath := filepath.Join(agentDir, info.dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue // directory might not exist
		}

		// Build set of known items (already in config)
		known := make(map[string]bool)
		for _, name := range info.items {
			known[name] = true
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue // skip subdirectories
			}
			if !strings.HasSuffix(entry.Name(), info.ext) {
				continue
			}

			name := strings.TrimSuffix(entry.Name(), info.ext)
			if known[name] {
				continue // already tracked in config
			}

			relPath, _ := filepath.Rel(projectRoot, filepath.Join(dirPath, entry.Name()))

			// Skip if already in lock file (tracked but somehow not in config — be safe)
			if _, tracked := lock.Files[relPath]; tracked {
				continue
			}

			// Read file and parse frontmatter
			data, err := os.ReadFile(filepath.Join(dirPath, entry.Name()))
			if err != nil {
				continue
			}

			meta, _ := ParseFrontmatter(data)

			df := DiscoveredFile{
				Name:    name,
				Type:    itemType,
				RelPath: relPath,
				Meta:    meta,
			}

			// Validate
			if meta == nil || meta.Description == "" {
				df.Error = "missing frontmatter (description required)"
			} else if itemType == "sensor" && meta.Event == "" {
				df.Error = "sensor missing 'event' in frontmatter"
			} else if itemType == "routine" && meta.Frequency == "" {
				df.Error = "routine missing 'frequency' in frontmatter"
			}

			discovered = append(discovered, df)
		}
	}

	return discovered, nil
}
