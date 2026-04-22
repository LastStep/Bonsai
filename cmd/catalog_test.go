package cmd

import (
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
