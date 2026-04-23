package listflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// TestRenderAll_EmptyConfigShowsEmptyState covers case (a): zero agents
// → empty-state panel + zero counts footer.
func TestRenderAll_EmptyConfigShowsEmptyState(t *testing.T) {
	cfg := &config.ProjectConfig{ProjectName: "demo"}

	out := RenderAll(cfg, nil, "0.1.2", t.TempDir(), 120, 40)

	if !strings.Contains(out, "No agents installed") {
		t.Fatalf("expected empty-state CTA, got:\n%s", out)
	}
	if !strings.Contains(out, "bonsai add") {
		t.Fatalf("expected bonsai add CTA, got:\n%s", out)
	}
	if !strings.Contains(out, "0 agents") {
		t.Fatalf("expected '0 agents' in counts footer, got:\n%s", out)
	}
}

// TestRenderAll_OneAgentNoScaffolding covers case (b): one installed
// agent, no scaffolding — one panel, no scaffolding row.
func TestRenderAll_OneAgentNoScaffolding(t *testing.T) {
	projectDir := t.TempDir()
	ws := filepath.Join(projectDir, "station")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	cfg := &config.ProjectConfig{
		ProjectName: "demo",
		Agents: map[string]*config.InstalledAgent{
			"tech-lead": {
				AgentType: "tech-lead",
				Workspace: "station",
			},
		},
	}

	out := RenderAll(cfg, nil, "0.1.2", projectDir, 120, 40)

	if strings.Contains(out, "Scaffolding:") {
		t.Fatalf("did not expect scaffolding row when cfg.Scaffolding is empty, got:\n%s", out)
	}
	if !strings.Contains(out, "Tech Lead") {
		t.Fatalf("expected Tech Lead panel title, got:\n%s", out)
	}
}

// TestRenderAll_AgentsSortedAlphabetically covers case (c): two agents
// in the config map, rendered in alphabetical order regardless of the
// map key iteration order. Uses "alpha" and "tech-lead" so the test
// catches a naive range-over-map ordering bug.
func TestRenderAll_AgentsSortedAlphabetically(t *testing.T) {
	projectDir := t.TempDir()

	cfg := &config.ProjectConfig{
		ProjectName: "demo",
		Agents: map[string]*config.InstalledAgent{
			"tech-lead": {AgentType: "tech-lead", Workspace: "station"},
			"alpha":     {AgentType: "alpha", Workspace: "alpha-ws"},
		},
	}

	out := RenderAll(cfg, nil, "0.1.2", projectDir, 120, 40)

	alphaIdx := strings.Index(out, "Alpha")
	techIdx := strings.Index(out, "Tech Lead")
	if alphaIdx < 0 || techIdx < 0 {
		t.Fatalf("expected both panel titles, got:\n%s", out)
	}
	if alphaIdx >= techIdx {
		t.Fatalf("expected 'Alpha' before 'Tech Lead' (alphabetical), got alpha=%d tech=%d:\n%s", alphaIdx, techIdx, out)
	}
}

// TestRenderAll_CountsFooterMatchesTotals covers case (d): counts footer
// reflects the sum across all agents.
func TestRenderAll_CountsFooterMatchesTotals(t *testing.T) {
	projectDir := t.TempDir()

	cfg := &config.ProjectConfig{
		ProjectName: "demo",
		Agents: map[string]*config.InstalledAgent{
			"a": {
				AgentType: "a",
				Workspace: "ws-a",
				Skills:    []string{"s1", "s2"},
				Workflows: []string{"w1"},
				Protocols: []string{"p1", "p2", "p3"},
				Sensors:   []string{"sen1"},
				Routines:  []string{"r1", "r2"},
			},
			"b": {
				AgentType: "b",
				Workspace: "ws-b",
				Skills:    []string{"s3"},
				Workflows: []string{},
				Protocols: []string{},
				Sensors:   []string{"sen2", "sen3"},
				Routines:  []string{},
			},
		},
	}

	out := RenderAll(cfg, nil, "0.1.2", projectDir, 120, 40)

	for _, want := range []string{
		"2 agents",
		"3 skills",
		"1 workflow",
		"3 protocols",
		"3 sensors",
		"2 routines",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in counts footer, got:\n%s", want, out)
		}
	}
}

// TestRenderAll_MinSizeFloor covers case (e): below the 70×20 threshold,
// the output collapses to RenderMinSizeFloor and nothing else.
func TestRenderAll_MinSizeFloor(t *testing.T) {
	cfg := &config.ProjectConfig{
		ProjectName: "demo",
		Agents: map[string]*config.InstalledAgent{
			"tech-lead": {AgentType: "tech-lead", Workspace: "station"},
		},
	}

	out := RenderAll(cfg, nil, "0.1.2", t.TempDir(), 40, 10)
	want := initflow.RenderMinSizeFloor(40, 10)

	if out != want {
		t.Fatalf("expected min-size floor output exactly, got mismatch\n--- want ---\n%s\n--- got ---\n%s", want, out)
	}
	if strings.Contains(out, "No agents installed") {
		t.Fatalf("min-size floor must suppress agent body; got:\n%s", out)
	}
}

// TestRenderAll_ScaffoldingRowRendered covers the scaffolding CTA line.
// When scaffolding is non-empty, each item renders as its display-name
// in a single muted row above the agent panels.
func TestRenderAll_ScaffoldingRowRendered(t *testing.T) {
	cfg := &config.ProjectConfig{
		ProjectName: "demo",
		Scaffolding: []string{"playbook", "logs"},
	}

	out := RenderAll(cfg, nil, "0.1.2", t.TempDir(), 120, 40)

	if !strings.Contains(out, "Scaffolding:") {
		t.Fatalf("expected 'Scaffolding:' label, got:\n%s", out)
	}
	if !strings.Contains(out, "Playbook") || !strings.Contains(out, "Logs") {
		t.Fatalf("expected scaffolding display names, got:\n%s", out)
	}
}
