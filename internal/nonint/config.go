package nonint

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/wsvalidate"
)

const (
	// defaultDocsPath mirrors the Vessel stage default — `bonsai init` places
	// the tech-lead workspace at `station/` unless the user overrides it.
	defaultDocsPath = "station/"

	// techLeadType is the canonical machine name for the orchestrating agent.
	// Hard-coded here to mirror cmd/init_flow.go; if Bonsai ever renames it,
	// these two constants need to update in lockstep.
	techLeadType = "tech-lead"
)

// LoadConfig reads <path> as a `.bonsai.yaml`-shaped YAML document, applies
// the lenient defaults from Plan 39 Locked Decision Q3, validates the result
// (shell-metachar scan + workspace normalisation), and returns a fully
// resolved *config.ProjectConfig ready to drive `RunInit` or `RunAdd`.
//
// Defaulting walk (Q3):
//   - ProjectName       → filepath.Base(cwd) when empty
//   - DocsPath          → "station/" when empty
//   - Scaffolding       → required scaffolding from catalog when nil
//   - per-agent fields  → agent.yaml `defaults` when every ability list is nil
//   - per-agent Workspace → cfg.DocsPath for tech-lead, else `<agentType>/`
//     (then normalised through wsvalidate.Normalise)
//
// Errors:
//   - "from-config: read <path>: ..."  on I/O failure
//   - "from-config: parse YAML: ..."   on bad YAML
//   - "from-config: missing required field 'agents' (need at least one entry)"
//   - validation errors from config.Validate (shell metachars, bad workspace)
//
// Note: the caller (`cmd/init.go`, `cmd/add.go`) is responsible for any
// command-specific guards on top of these — e.g. RunAdd enforces the
// "exactly one agent in overlay" rule that init does not.
func LoadConfig(path, cwd string, cat *catalog.Catalog) (*config.ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("from-config: read %s: %w", path, err)
	}

	var cfg config.ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("from-config: parse YAML: %w", err)
	}

	if len(cfg.Agents) == 0 {
		return nil, fmt.Errorf("from-config: missing required field 'agents' (need at least one entry)")
	}

	applyDefaults(&cfg, cwd, cat)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("from-config: %w", err)
	}
	return &cfg, nil
}

// applyDefaults walks an unmarshalled config and fills in every Q3-lenient
// default. Pure function on cfg — no I/O, no error path — so unit tests can
// exercise the defaulting independently of the YAML reader.
func applyDefaults(cfg *config.ProjectConfig, cwd string, cat *catalog.Catalog) {
	if cfg.ProjectName == "" {
		cfg.ProjectName = filepath.Base(cwd)
	}
	if cfg.DocsPath == "" {
		cfg.DocsPath = defaultDocsPath
	}
	// `nil` means "user omitted the key"; an explicit empty list (e.g.
	// `scaffolding: []`) is treated as the user opting out of scaffolding.
	// yaml.v3 reports both as `nil` after Unmarshal, but distinguishing
	// would require a custom unmarshaller — Plan 39 §A.1 calls for the
	// nil-only check and the user can pass an empty-required scaffolding
	// list in their YAML by listing the items explicitly.
	if cfg.Scaffolding == nil {
		for _, item := range cat.Scaffolding {
			if item.Required {
				cfg.Scaffolding = append(cfg.Scaffolding, item.Name)
			}
		}
	}

	for agentType, agent := range cfg.Agents {
		if agent == nil {
			agent = &config.InstalledAgent{}
			cfg.Agents[agentType] = agent
		}
		// AgentType field is normally redundant with the map key, but the
		// generator reads agent.AgentType directly (see generate.AgentWorkspace
		// and template context wiring), so mirror it in lock-step with the key.
		agent.AgentType = agentType

		if agent.Workspace == "" {
			if agentType == techLeadType {
				agent.Workspace = cfg.DocsPath
			} else {
				agent.Workspace = wsvalidate.Normalise(agentType + "/")
			}
		} else {
			agent.Workspace = wsvalidate.Normalise(agent.Workspace)
		}

		// If every ability list is nil, fall back to the agent.yaml defaults.
		// "Every nil" is the user signalling "give me the defaults"; an
		// explicit empty list opts out of that category. Mirrors the
		// interactive BranchesStage's initial-selection logic.
		if agent.Skills == nil && agent.Workflows == nil &&
			agent.Protocols == nil && agent.Sensors == nil &&
			agent.Routines == nil {
			if def := cat.GetAgent(agentType); def != nil {
				agent.Skills = append([]string(nil), def.DefaultSkills...)
				agent.Workflows = append([]string(nil), def.DefaultWorkflows...)
				agent.Protocols = append([]string(nil), def.DefaultProtocols...)
				agent.Sensors = append([]string(nil), def.DefaultSensors...)
				agent.Routines = append([]string(nil), def.DefaultRoutines...)
			}
		}

		// EnsureRoutineCheckSensor adds `routine-check` iff any routines are
		// installed and it is not already in the sensor list. Mirrors the
		// interactive flow's wiring so the SettingsJSON generator sees the
		// hook registration exactly as it would under the TUI.
		generate.EnsureRoutineCheckSensor(agent)
	}
}
