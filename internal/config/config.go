package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/LastStep/Bonsai/internal/wsvalidate"
	"gopkg.in/yaml.v3"
)

// forbiddenShellChars is the deny-list of characters that must never appear
// in user-controlled strings emitted into shell scripts (sensor templates,
// status-bar prompts, etc.). The set is intentionally small and explicit —
// quote/backtick/dollar enable command substitution, backslash + newline
// break out of single-line contexts, and the bracket/paren pairs guard
// against array/process-substitution surprises in bash.
var forbiddenShellChars = []rune{'"', '`', '$', '\\', '\n', ']', ')', '[', '('}

// CustomItemMeta holds metadata for user-created custom items (parsed from frontmatter).
type CustomItemMeta struct {
	DisplayName string `yaml:"display_name,omitempty"`
	Description string `yaml:"description"`
	Event       string `yaml:"event,omitempty"`     // sensors only
	Matcher     string `yaml:"matcher,omitempty"`   // sensors only
	Frequency   string `yaml:"frequency,omitempty"` // routines only
}

// InstalledAgent represents an agent installed in a project.
type InstalledAgent struct {
	AgentType   string                     `yaml:"agent_type"`
	Workspace   string                     `yaml:"workspace"`
	Skills      []string                   `yaml:"skills"`
	Workflows   []string                   `yaml:"workflows"`
	Protocols   []string                   `yaml:"protocols"`
	Sensors     []string                   `yaml:"sensors"`
	Routines    []string                   `yaml:"routines"`
	CustomItems map[string]*CustomItemMeta `yaml:"custom_items,omitempty"`
}

// ProjectConfig is the root project config serialized to .bonsai.yaml.
type ProjectConfig struct {
	ProjectName string                     `yaml:"project_name"`
	Description string                     `yaml:"description,omitempty"`
	DocsPath    string                     `yaml:"docs_path,omitempty"`
	Scaffolding []string                   `yaml:"scaffolding,omitempty"`
	Agents      map[string]*InstalledAgent `yaml:"agents,omitempty"`
}

// Validate checks the project config for required fields, valid workspace
// paths, and the absence of shell metacharacters in user-controlled strings.
// Run after YAML unmarshalling so callers fail fast on hand-edited configs.
//
// Workspace + DocsPath are validated through wsvalidate (same rules as the
// init/add TUI flows). The shell-metachar scan covers ProjectName, every
// agent name (map key), every agent.Workspace, and DocsPath — these strings
// flow into shell scripts via sensor templates, so a stray `"` or `$` would
// either break the script or open a substitution channel.
func (c *ProjectConfig) Validate() error {
	if c.ProjectName == "" {
		return errors.New("project_name is required")
	}

	for name, agent := range c.Agents {
		if agent == nil {
			continue
		}
		ws := wsvalidate.Normalise(agent.Workspace)
		if reason := wsvalidate.InvalidReason(ws); reason != "" {
			return fmt.Errorf("agent %q: workspace %q invalid: %s", name, agent.Workspace, reason)
		}
	}

	if c.DocsPath != "" {
		dp := wsvalidate.Normalise(c.DocsPath)
		if reason := wsvalidate.InvalidReason(dp); reason != "" {
			return fmt.Errorf("docs_path %q invalid: %s", c.DocsPath, reason)
		}
	}

	if err := scanShellMetachars("project_name", c.ProjectName); err != nil {
		return err
	}
	for name, agent := range c.Agents {
		if err := scanShellMetachars("agent name", name); err != nil {
			return err
		}
		if agent == nil {
			continue
		}
		if err := scanShellMetachars(fmt.Sprintf("agent %q workspace", name), agent.Workspace); err != nil {
			return err
		}
	}
	if c.DocsPath != "" {
		if err := scanShellMetachars("docs_path", c.DocsPath); err != nil {
			return err
		}
	}
	return nil
}

// scanShellMetachars returns an error on the first forbidden rune in v.
// Field is the human-readable label included in the error message so the
// caller can pinpoint which YAML key tripped the check.
func scanShellMetachars(field, v string) error {
	for _, r := range v {
		for _, bad := range forbiddenShellChars {
			if r == bad {
				return fmt.Errorf("field %q contains forbidden character %q", field, string(r))
			}
		}
	}
	return nil
}

// Save writes the config to a YAML file.
func (c *ProjectConfig) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads a ProjectConfig from a YAML file.
func Load(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Agents == nil {
		cfg.Agents = make(map[string]*InstalledAgent)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return &cfg, nil
}
