package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LastStep/Bonsai/internal/catalog"
)

// CatalogSnapshot is the stable JSON shape written to .bonsai/catalog.json.
// Agent-consumable — provides a filesystem-discoverable listing of every
// agent/skill/workflow/protocol/sensor/routine that the installed Bonsai
// binary ships.
//
// This shape is deliberately decoupled from the internal catalog types in
// internal/catalog — those may evolve. The JSON schema here is a stable
// contract for downstream agent readers.
type CatalogSnapshot struct {
	Version   string         `json:"version"`
	Agents    []AgentEntry   `json:"agents"`
	Skills    []AbilityEntry `json:"skills"`
	Workflows []AbilityEntry `json:"workflows"`
	Protocols []AbilityEntry `json:"protocols"`
	Sensors   []SensorEntry  `json:"sensors"`
	Routines  []RoutineEntry `json:"routines"`
}

// AgentEntry describes an agent type shipped in the catalog.
type AgentEntry struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// AbilityEntry describes a skill, workflow, or protocol. Agents is the list
// of agent types the ability is compatible with ("all" or specific names);
// Required is the subset the ability is forcibly installed on.
type AbilityEntry struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Agents      []string `json:"agents"`
	Required    []string `json:"required,omitempty"`
}

// SensorEntry extends AbilityEntry with event + matcher fields unique to sensors.
type SensorEntry struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Agents      []string `json:"agents"`
	Required    []string `json:"required,omitempty"`
	Event       string   `json:"event"`
	Matcher     string   `json:"matcher,omitempty"`
}

// RoutineEntry extends AbilityEntry with the frequency field.
type RoutineEntry struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Agents      []string `json:"agents"`
	Required    []string `json:"required,omitempty"`
	Frequency   string   `json:"frequency"`
}

// compatToSlice converts an AgentCompat into a JSON-friendly slice.
// "all" collapses to ["all"]. When omitEmpty is true, an empty Names
// list returns nil so JSON `omitempty` drops the field; when false,
// it returns an empty slice (always-emit).
func compatToSlice(a catalog.AgentCompat, omitEmpty bool) []string {
	if a.All {
		return []string{"all"}
	}
	if len(a.Names) == 0 {
		if omitEmpty {
			return nil
		}
		return []string{}
	}
	out := make([]string, len(a.Names))
	copy(out, a.Names)
	return out
}

// SerializeCatalog builds a CatalogSnapshot from the in-memory catalog and
// returns its JSON representation (2-space indented). Single source of truth
// for both WriteCatalogSnapshot (on-disk) and `bonsai catalog --json` (stdout,
// PR2).
func SerializeCatalog(cat *catalog.Catalog, version string) ([]byte, error) {
	if cat == nil {
		return nil, fmt.Errorf("nil catalog")
	}
	snap := CatalogSnapshot{
		Version:   version,
		Agents:    make([]AgentEntry, 0, len(cat.Agents)),
		Skills:    make([]AbilityEntry, 0, len(cat.Skills)),
		Workflows: make([]AbilityEntry, 0, len(cat.Workflows)),
		Protocols: make([]AbilityEntry, 0, len(cat.Protocols)),
		Sensors:   make([]SensorEntry, 0, len(cat.Sensors)),
		Routines:  make([]RoutineEntry, 0, len(cat.Routines)),
	}

	for _, a := range cat.Agents {
		snap.Agents = append(snap.Agents, AgentEntry{
			Name:        a.Name,
			DisplayName: a.DisplayName,
			Description: a.Description,
		})
	}
	for _, s := range cat.Skills {
		snap.Skills = append(snap.Skills, AbilityEntry{
			Name:        s.Name,
			DisplayName: s.DisplayName,
			Description: s.Description,
			Agents:      compatToSlice(s.Agents, false),
			Required:    compatToSlice(s.Required, true),
		})
	}
	for _, w := range cat.Workflows {
		snap.Workflows = append(snap.Workflows, AbilityEntry{
			Name:        w.Name,
			DisplayName: w.DisplayName,
			Description: w.Description,
			Agents:      compatToSlice(w.Agents, false),
			Required:    compatToSlice(w.Required, true),
		})
	}
	for _, p := range cat.Protocols {
		snap.Protocols = append(snap.Protocols, AbilityEntry{
			Name:        p.Name,
			DisplayName: p.DisplayName,
			Description: p.Description,
			Agents:      compatToSlice(p.Agents, false),
			Required:    compatToSlice(p.Required, true),
		})
	}
	for _, s := range cat.Sensors {
		snap.Sensors = append(snap.Sensors, SensorEntry{
			Name:        s.Name,
			DisplayName: s.DisplayName,
			Description: s.Description,
			Agents:      compatToSlice(s.Agents, false),
			Required:    compatToSlice(s.Required, true),
			Event:       s.Event,
			Matcher:     s.Matcher,
		})
	}
	for _, r := range cat.Routines {
		snap.Routines = append(snap.Routines, RoutineEntry{
			Name:        r.Name,
			DisplayName: r.DisplayName,
			Description: r.Description,
			Agents:      compatToSlice(r.Agents, false),
			Required:    compatToSlice(r.Required, true),
			Frequency:   r.Frequency,
		})
	}

	return json.MarshalIndent(snap, "", "  ")
}

// WriteCatalogSnapshot serializes the catalog to .bonsai/catalog.json at the
// project root. Creates the .bonsai/ directory (mode 0755) if missing. The
// file is written with mode 0644 and terminated with a trailing newline.
//
// NOT lock-tracked — the snapshot is regenerated on every init/add/update
// and is agent-read-only; no user edits are expected. A path outcome is
// appended to result (ActionCreated for fresh writes, ActionUpdated when
// content differed from an existing snapshot, ActionUnchanged otherwise)
// so callers can display it alongside other generated files.
func WriteCatalogSnapshot(projectRoot string, version string, cat *catalog.Catalog, result *WriteResult) error {
	data, err := SerializeCatalog(cat, version)
	if err != nil {
		return fmt.Errorf("serialize catalog: %w", err)
	}
	data = append(data, '\n')

	bonsaiDir := filepath.Join(projectRoot, ".bonsai")
	if err := os.MkdirAll(bonsaiDir, 0755); err != nil {
		return fmt.Errorf("create .bonsai/: %w", err)
	}

	relPath := filepath.Join(".bonsai", "catalog.json")
	absPath := filepath.Join(projectRoot, relPath)

	// Short-circuit to Unchanged if the on-disk content already matches.
	action := ActionCreated
	if existing, rerr := os.ReadFile(absPath); rerr == nil {
		if string(existing) == string(data) {
			result.Add(FileResult{RelPath: relPath, Action: ActionUnchanged, Source: "generated:catalog-snapshot"})
			return nil
		}
		action = ActionUpdated
	}

	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return fmt.Errorf("write catalog.json: %w", err)
	}
	result.Add(FileResult{RelPath: relPath, Action: action, Source: "generated:catalog-snapshot"})
	return nil
}
