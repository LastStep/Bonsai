package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// CustomItemMeta holds metadata for user-created custom items (parsed from frontmatter).
type CustomItemMeta struct {
	DisplayName string `yaml:"display_name,omitempty"`
	Description string `yaml:"description"`
	Event       string `yaml:"event,omitempty"`     // sensors only
	Matcher     string `yaml:"matcher,omitempty"`    // sensors only
	Frequency   string `yaml:"frequency,omitempty"`  // routines only
}

// InstalledAgent represents an agent installed in a project.
type InstalledAgent struct {
	AgentType   string                      `yaml:"agent_type"`
	Workspace   string                      `yaml:"workspace"`
	Skills      []string                    `yaml:"skills"`
	Workflows   []string                    `yaml:"workflows"`
	Protocols   []string                    `yaml:"protocols"`
	Sensors     []string                    `yaml:"sensors"`
	Routines    []string                    `yaml:"routines"`
	CustomItems map[string]*CustomItemMeta  `yaml:"custom_items,omitempty"`
}

// ProjectConfig is the root project config serialized to .bonsai.yaml.
type ProjectConfig struct {
	ProjectName string                     `yaml:"project_name"`
	Description string                     `yaml:"description,omitempty"`
	DocsPath    string                     `yaml:"docs_path,omitempty"`
	Scaffolding []string                   `yaml:"scaffolding,omitempty"`
	Agents      map[string]*InstalledAgent `yaml:"agents,omitempty"`
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
	return &cfg, nil
}
