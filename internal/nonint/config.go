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
// resolved *config.ProjectConfig ready to drive `RunInit`.
//
// For overlay configs consumed by `RunAdd`, use `LoadOverlay` instead — that
// variant skips the project-level defaults (project_name, docs_path,
// scaffolding) so the §3 match contract ("leave empty or match exactly")
// holds against the user's literal YAML, not the cwd-basename fallback.
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

// LoadOverlay reads an overlay YAML for `bonsai add`. Same parsing path as
// LoadConfig, but only per-agent defaults are applied — project-level
// fields (project_name, docs_path, scaffolding) remain at their literal
// YAML values (typically empty) so the §3 contract holds against the
// user's input rather than a defaulted fallback.
//
// Defaulting applied here:
//   - per-agent Workspace → cfg.DocsPath when techLead, else `<agentType>/`
//     (only if cfg.DocsPath is set; otherwise blank workspace defers to
//     RunAdd's existing-cfg-aware logic)
//   - per-agent ability lists → agent.yaml defaults when every list is nil
//   - per-agent routine-check sensor → wired iff routines present
//
// Validation differs from LoadConfig: project_name="" is accepted (overlays
// don't carry that field for the typical case), but every other shell-
// metachar / workspace rule still runs.
func LoadOverlay(path, cwd string, cat *catalog.Catalog) (*config.ProjectConfig, error) {
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
	applyAgentDefaults(&cfg, cat)
	// Run the wsvalidate + shell-metachar checks the same way
	// config.Validate does, but tolerate an empty project_name (we want
	// the §3 "leave empty or match exactly" contract to hold; the matched
	// existing config's name is the one that ultimately gets persisted).
	if err := validateOverlay(&cfg); err != nil {
		return nil, fmt.Errorf("from-config: %w", err)
	}
	return &cfg, nil
}

// validateOverlay runs the subset of config.Validate appropriate for an
// overlay: every check except the required-project-name rule. Inlined here
// rather than refactored into config/ so the project_name-is-required
// behaviour for normal config load paths stays unchanged.
func validateOverlay(c *config.ProjectConfig) error {
	if c.ProjectName != "" {
		// Re-use the shell-metachar scan via config.Validate by setting a
		// placeholder, running, and reverting on success. Simpler though:
		// rely on the documented sentinel — config.Validate scans every
		// field including project_name. Run it; if it errors with the
		// project_name-required message we suppress.
		if err := c.Validate(); err != nil {
			return err
		}
		return nil
	}
	// project_name is empty — temporarily set a benign placeholder so
	// config.Validate's shell-metachar scan doesn't trip on it, then
	// restore. The placeholder uses only safe characters so it can't
	// itself trip the scan.
	c.ProjectName = "_overlay_placeholder_"
	defer func() { c.ProjectName = "" }()
	return c.Validate()
}

// applyAgentDefaults is the overlay-only subset of applyDefaults: it
// populates per-agent workspaces + ability lists but leaves project-level
// fields untouched. Pure function; used by LoadOverlay.
func applyAgentDefaults(cfg *config.ProjectConfig, cat *catalog.Catalog) {
	for agentType, agent := range cfg.Agents {
		if agent == nil {
			agent = &config.InstalledAgent{}
			cfg.Agents[agentType] = agent
		}
		agent.AgentType = agentType
		if agent.Workspace == "" {
			if agentType == techLeadType && cfg.DocsPath != "" {
				agent.Workspace = cfg.DocsPath
			} else {
				agent.Workspace = wsvalidate.Normalise(agentType + "/")
			}
		} else {
			agent.Workspace = wsvalidate.Normalise(agent.Workspace)
		}
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
		generate.EnsureRoutineCheckSensor(agent)
	}
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
