package generate

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
)

// ListSnapshot is the stable JSON shape emitted by `bonsai list --json`.
// Agent-consumable — a single-document snapshot of the project's installed
// agents and the abilities each carries, read directly from .bonsai.yaml.
//
// The shape mirrors validate.Report's serialization conventions (explicit
// struct, indent-2, no map-vs-list ambiguity) so an AI agent / CI script
// driving Bonsai headlessly gets a pinned contract. Plan 41 Phase 4.
//
// This sits beside SerializeCatalog (catalog_snapshot.go) in internal/generate
// precisely because that package is TUI-free: the list serializer must pull in
// zero chrome (huh/bubbletea/lipgloss/glamour/charm) so a future bonsai mcp
// server (Plan 42) can wrap it directly. A TUI-free import scan guards this.
type ListSnapshot struct {
	Version  string      `json:"version"`
	DocsPath string      `json:"docs_path"`
	Agents   []ListAgent `json:"agents"`
}

// ListAgent describes one installed agent and its registered abilities. Type
// is the agent-type machine name (the .bonsai.yaml agents-map key); the
// ability slices are always emitted (never nil) so the JSON shape is stable
// — an agent with no skills serializes "skills": [] rather than null.
type ListAgent struct {
	Type      string   `json:"type"`
	Workspace string   `json:"workspace"`
	Skills    []string `json:"skills"`
	Workflows []string `json:"workflows"`
	Protocols []string `json:"protocols"`
	Sensors   []string `json:"sensors"`
	Routines  []string `json:"routines"`
}

// SerializeJSON builds a ListSnapshot from the project config and returns its
// indent-2 JSON representation (matching SerializeCatalog / validate --json).
// Agents are emitted alphabetically by type for deterministic output.
//
// The signature mirrors the cinematic list renderer (cfg, cat, version, cwd)
// for call-site parity and Plan 42 readiness. cat and cwd are accepted but not
// consulted today — all snapshot data lives in the config; display-name and
// workspace-tree concerns belong to the TTY path, not the headless contract.
func SerializeJSON(cfg *config.ProjectConfig, _ *catalog.Catalog, version, _ string) ([]byte, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil config")
	}

	snap := ListSnapshot{
		Version:  version,
		DocsPath: cfg.DocsPath,
		Agents:   make([]ListAgent, 0, len(cfg.Agents)),
	}

	names := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		agent := cfg.Agents[name]
		if agent == nil {
			continue
		}
		snap.Agents = append(snap.Agents, ListAgent{
			Type:      name,
			Workspace: agent.Workspace,
			Skills:    nonNil(agent.Skills),
			Workflows: nonNil(agent.Workflows),
			Protocols: nonNil(agent.Protocols),
			Sensors:   nonNil(agent.Sensors),
			Routines:  nonNil(agent.Routines),
		})
	}

	return json.MarshalIndent(snap, "", "  ")
}

// nonNil returns an empty (non-nil) slice for a nil input so the field
// marshals to [] rather than null — keeps the JSON shape uniform.
func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
