package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/catalog"
)

// fakeCatalogForStatic returns a minimal in-memory catalog covering
// all 7 sections so renderCatalogStatic exercises every branch.
func fakeCatalogForStatic() *catalog.Catalog {
	return &catalog.Catalog{
		Agents: []catalog.AgentDef{
			{Name: "tech-lead", DisplayName: "Tech Lead", Description: "orchestrator"},
		},
		Skills: []catalog.CatalogItem{
			{Name: "planning-template", DisplayName: "Planning Template", Description: "tiered plans", Agents: catalog.AgentCompat{All: true}},
		},
		Workflows: []catalog.CatalogItem{
			{Name: "code-review", DisplayName: "Code Review", Description: "review pipeline", Agents: catalog.AgentCompat{All: true}},
		},
		Protocols: []catalog.CatalogItem{
			{Name: "memory", DisplayName: "Memory", Description: "memory protocol", Agents: catalog.AgentCompat{All: true}},
		},
		Sensors: []catalog.SensorItem{
			{Name: "status-bar", DisplayName: "Status Bar", Description: "status line", Event: "Stop", Agents: catalog.AgentCompat{All: true}},
		},
		Routines: []catalog.RoutineItem{
			{Name: "backlog-hygiene", DisplayName: "Backlog Hygiene", Description: "weekly sweep", Frequency: "7 days", Agents: catalog.AgentCompat{All: true}},
		},
		Scaffolding: []catalog.ScaffoldingItem{
			{Name: "playbook", DisplayName: "Playbook", Description: "plans + roadmap", Required: true, Affects: "planning"},
		},
	}
}

// TestRenderCatalogStatic_AllSevenSectionsPresent verifies the non-TTY
// static render produces a header for each of the 7 catalog sections
// and embeds the entry count.
func TestRenderCatalogStatic_AllSevenSectionsPresent(t *testing.T) {
	cat := fakeCatalogForStatic()

	out := captureStdout(t, func() {
		if err := renderCatalogStatic(cat, ""); err != nil {
			t.Fatalf("renderCatalogStatic: %v", err)
		}
	})

	// Each section header renders as "<Name> (<count>)" with a trailing
	// rule. Assert presence of each headline + embedded count.
	for _, want := range []string{
		"Agents (1)",
		"Skills (1)",
		"Workflows (1)",
		"Protocols (1)",
		"Sensors (1)",
		"Routines (1)",
		"Scaffolding (1)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("static catalog output missing %q:\n%s", want, out)
		}
	}
}

// TestRenderCatalogStatic_AgentFilterChangesSuffix verifies the
// agent-filter branch appends the "for <agent>" suffix on filterable
// sections.
func TestRenderCatalogStatic_AgentFilterChangesSuffix(t *testing.T) {
	cat := fakeCatalogForStatic()

	out := captureStdout(t, func() {
		if err := renderCatalogStatic(cat, "tech-lead"); err != nil {
			t.Fatalf("renderCatalogStatic: %v", err)
		}
	})

	// Skills/Workflows/Protocols/Sensors/Routines show the filter suffix.
	// Agents and Scaffolding do not (per the existing static-render
	// contract).
	for _, want := range []string{
		"Skills (1 for tech-lead)",
		"Workflows (1 for tech-lead)",
		"Protocols (1 for tech-lead)",
		"Sensors (1 for tech-lead)",
		"Routines (1 for tech-lead)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("filtered catalog output missing %q:\n%s", want, out)
		}
	}
}

// TestCatalog_JSONFlagEmitsStableSchema verifies renderCatalogJSON emits
// a JSON envelope with the six catalog categories populated from the
// supplied catalog. Locks the shape so downstream agent readers can rely
// on a stable top-level schema.
func TestCatalog_JSONFlagEmitsStableSchema(t *testing.T) {
	cat := fakeCatalogForStatic()

	out := captureStdout(t, func() {
		if err := renderCatalogJSON(cat, ""); err != nil {
			t.Fatalf("renderCatalogJSON: %v", err)
		}
	})

	// Round-trip into the stable envelope shape.
	var snapshot struct {
		Version   string           `json:"version"`
		Agents    []map[string]any `json:"agents"`
		Skills    []map[string]any `json:"skills"`
		Workflows []map[string]any `json:"workflows"`
		Protocols []map[string]any `json:"protocols"`
		Sensors   []map[string]any `json:"sensors"`
		Routines  []map[string]any `json:"routines"`
	}
	if err := json.Unmarshal([]byte(out), &snapshot); err != nil {
		t.Fatalf("JSON unmarshal: %v\noutput:\n%s", err, out)
	}

	if len(snapshot.Agents) != 1 || snapshot.Agents[0]["name"] != "tech-lead" {
		t.Fatalf("agents[0].name = %v, want tech-lead", snapshot.Agents)
	}
	if len(snapshot.Skills) != 1 || snapshot.Skills[0]["name"] != "planning-template" {
		t.Fatalf("skills[0].name = %v, want planning-template", snapshot.Skills)
	}
	if len(snapshot.Workflows) != 1 || len(snapshot.Protocols) != 1 ||
		len(snapshot.Sensors) != 1 || len(snapshot.Routines) != 1 {
		t.Fatalf("category count mismatch: %+v", snapshot)
	}
	// Sensors carry an event field.
	if snapshot.Sensors[0]["event"] != "Stop" {
		t.Fatalf("sensors[0].event = %v, want Stop", snapshot.Sensors[0])
	}
	// Routines carry a frequency field.
	if snapshot.Routines[0]["frequency"] != "7 days" {
		t.Fatalf("routines[0].frequency = %v, want 7 days", snapshot.Routines[0])
	}
}

// TestCatalog_JSONFlagRespectsAgentFilter verifies the --json path reduces
// the catalog view when -a <agent> is set. Skills compatible only with a
// non-matching agent should be excluded from the output.
func TestCatalog_JSONFlagRespectsAgentFilter(t *testing.T) {
	cat := &catalog.Catalog{
		Agents: []catalog.AgentDef{
			{Name: "tech-lead", DisplayName: "Tech Lead"},
			{Name: "backend", DisplayName: "Backend"},
		},
		Skills: []catalog.CatalogItem{
			{Name: "planning-template", DisplayName: "Planning Template",
				Agents: catalog.AgentCompat{Names: []string{"tech-lead"}}},
			{Name: "coding-standards", DisplayName: "Coding Standards",
				Agents: catalog.AgentCompat{All: true}},
		},
	}

	out := captureStdout(t, func() {
		if err := renderCatalogJSON(cat, "backend"); err != nil {
			t.Fatalf("renderCatalogJSON with filter: %v", err)
		}
	})

	var snapshot struct {
		Skills []map[string]any `json:"skills"`
	}
	if err := json.Unmarshal([]byte(out), &snapshot); err != nil {
		t.Fatalf("JSON unmarshal: %v\noutput:\n%s", err, out)
	}

	// "planning-template" should be filtered out (tech-lead only);
	// "coding-standards" should pass (agents: all).
	var names []string
	for _, s := range snapshot.Skills {
		names = append(names, s["name"].(string))
	}
	for _, name := range names {
		if name == "planning-template" {
			t.Fatalf("-a backend should exclude tech-lead-only skill; got: %v", names)
		}
	}
	if len(names) != 1 || names[0] != "coding-standards" {
		t.Fatalf("-a backend skills = %v, want [coding-standards]", names)
	}
}

// TestFilterCatalog_NoFilterReturnsOriginal verifies the empty-filter
// short-circuit returns the input unchanged.
func TestFilterCatalog_NoFilterReturnsOriginal(t *testing.T) {
	cat := fakeCatalogForStatic()
	got := filterCatalog(cat, "")
	if got != cat {
		t.Fatal("empty filter should return the original catalog pointer")
	}
}
