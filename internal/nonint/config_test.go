package nonint

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	bonsai "github.com/LastStep/Bonsai"
	"github.com/LastStep/Bonsai/internal/catalog"
)

// loadTestCatalog mounts the embedded catalog/ subtree the same way
// cmd/bonsai/main.go does at runtime. Each test that touches LoadConfig
// or RunInit needs this — the catalog is the source of truth for default
// abilities + required scaffolding.
func loadTestCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	sub, err := fs.Sub(bonsai.CatalogFS, "catalog")
	if err != nil {
		t.Fatalf("fs.Sub(CatalogFS): %v", err)
	}
	cat, err := catalog.New(sub)
	if err != nil {
		t.Fatalf("catalog.New: %v", err)
	}
	return cat
}

// writeYAML drops a YAML fixture in tmp and returns its absolute path.
func writeYAML(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", p, err)
	}
	return p
}

// TestLoadConfig_Minimal_AppliesAllDefaults: input has only the tech-lead
// agent. Every Q3 default — ProjectName from cwd basename, DocsPath=station/,
// Scaffolding=required, agent.Workspace=DocsPath, agent ability lists from
// agent.yaml defaults — must be filled in.
func TestLoadConfig_Minimal_AppliesAllDefaults(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	cfgPath := writeYAML(t, tmp, "cfg.yaml", "agents:\n  tech-lead: {}\n")
	cwd := tmp

	cfg, err := LoadConfig(cfgPath, cwd, cat)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.ProjectName != filepath.Base(cwd) {
		t.Errorf("ProjectName: want %q, got %q", filepath.Base(cwd), cfg.ProjectName)
	}
	if cfg.DocsPath != "station/" {
		t.Errorf("DocsPath: want station/, got %q", cfg.DocsPath)
	}
	// Required-only scaffolding is what we expect.
	var wantScaff []string
	for _, item := range cat.Scaffolding {
		if item.Required {
			wantScaff = append(wantScaff, item.Name)
		}
	}
	gotScaff := append([]string(nil), cfg.Scaffolding...)
	sort.Strings(gotScaff)
	sort.Strings(wantScaff)
	if !reflect.DeepEqual(gotScaff, wantScaff) {
		t.Errorf("Scaffolding: want %v, got %v", wantScaff, gotScaff)
	}

	agent, ok := cfg.Agents["tech-lead"]
	if !ok || agent == nil {
		t.Fatalf("tech-lead agent missing after defaulting")
	}
	if agent.Workspace != cfg.DocsPath {
		t.Errorf("Workspace: want %q (DocsPath), got %q", cfg.DocsPath, agent.Workspace)
	}
	def := cat.GetAgent("tech-lead")
	if def == nil {
		t.Fatalf("catalog missing tech-lead def")
	}
	if !reflect.DeepEqual(agent.Skills, def.DefaultSkills) {
		t.Errorf("Skills: want %v, got %v", def.DefaultSkills, agent.Skills)
	}
	if !reflect.DeepEqual(agent.Workflows, def.DefaultWorkflows) {
		t.Errorf("Workflows: want %v, got %v", def.DefaultWorkflows, agent.Workflows)
	}
	if !reflect.DeepEqual(agent.Protocols, def.DefaultProtocols) {
		t.Errorf("Protocols: want %v, got %v", def.DefaultProtocols, agent.Protocols)
	}
}

// TestLoadConfig_NonTechLeadDefaultWorkspace: a non-tech-lead agent with no
// workspace set gets `<agentType>/` after Normalise. This is required for
// the new-agent branch under bonsai add.
func TestLoadConfig_NonTechLeadDefaultWorkspace(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	// Pick the first non-tech-lead agent from the catalog so this test
	// stays honest as new agent types ship.
	var otherAgent string
	for _, a := range cat.Agents {
		if a.Name != "tech-lead" {
			otherAgent = a.Name
			break
		}
	}
	if otherAgent == "" {
		t.Skip("no non-tech-lead agent in catalog — cannot exercise this branch")
	}
	body := "agents:\n  " + otherAgent + ": {}\n"
	cfgPath := writeYAML(t, tmp, "cfg.yaml", body)

	cfg, err := LoadConfig(cfgPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	agent := cfg.Agents[otherAgent]
	if agent == nil {
		t.Fatalf("agent missing")
	}
	want := otherAgent + "/"
	if agent.Workspace != want {
		t.Errorf("Workspace: want %q, got %q", want, agent.Workspace)
	}
}

// TestLoadConfig_MissingAgents_Errors: empty agents map → exit 2 candidate.
func TestLoadConfig_MissingAgents_Errors(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	cfgPath := writeYAML(t, tmp, "cfg.yaml", "project_name: foo\n")

	_, err := LoadConfig(cfgPath, tmp, cat)
	if err == nil {
		t.Fatalf("expected error for empty agents:")
	}
	if !strings.Contains(err.Error(), "missing required field 'agents'") {
		t.Errorf("error message: want 'missing required field 'agents''; got %q", err.Error())
	}
}

// TestLoadConfig_InvalidWorkspace_Errors: a workspace pointing outside the
// project (../etc/passwd) must be rejected by the wsvalidate-driven
// config.Validate path.
func TestLoadConfig_InvalidWorkspace_Errors(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	body := "agents:\n  tech-lead:\n    workspace: ../etc/passwd\n"
	cfgPath := writeYAML(t, tmp, "cfg.yaml", body)

	_, err := LoadConfig(cfgPath, tmp, cat)
	if err == nil {
		t.Fatalf("expected error for ../etc/passwd workspace")
	}
	if !strings.Contains(err.Error(), "workspace") && !strings.Contains(err.Error(), "escape") {
		t.Errorf("error must mention workspace/escape; got %q", err.Error())
	}
}

// TestLoadConfig_ShellMetachar_Errors: project_name containing a shell
// metacharacter (here `$`) must be rejected — these strings flow into shell
// scripts via sensor templates.
func TestLoadConfig_ShellMetachar_Errors(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	body := "project_name: proj$danger\nagents:\n  tech-lead: {}\n"
	cfgPath := writeYAML(t, tmp, "cfg.yaml", body)

	_, err := LoadConfig(cfgPath, tmp, cat)
	if err == nil {
		t.Fatalf("expected error for shell metachar in project_name")
	}
	if !strings.Contains(err.Error(), "forbidden character") {
		t.Errorf("error: want 'forbidden character'; got %q", err.Error())
	}
}

// TestLoadConfig_BadYAML_Errors: unparseable YAML must produce a parse-YAML
// prefixed error.
func TestLoadConfig_BadYAML_Errors(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	cfgPath := writeYAML(t, tmp, "cfg.yaml", "agents: [this is: not valid yaml")

	_, err := LoadConfig(cfgPath, tmp, cat)
	if err == nil {
		t.Fatalf("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse YAML") && !strings.Contains(err.Error(), "from-config") {
		t.Errorf("error: want 'parse YAML' or 'from-config'; got %q", err.Error())
	}
}

// TestLoadConfig_MissingFile_Errors: --from-config pointing at a non-existent
// path must yield a read-prefixed error.
func TestLoadConfig_MissingFile_Errors(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	missing := filepath.Join(tmp, "does-not-exist.yaml")

	_, err := LoadConfig(missing, tmp, cat)
	if err == nil {
		t.Fatalf("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "from-config: read") {
		t.Errorf("error: want 'from-config: read'; got %q", err.Error())
	}
}

// TestLoadConfig_ExplicitOverridesAreRespected: user-supplied values must
// survive the defaulting walk untouched. Round-trip: write a hand-tuned
// cfg, load it, verify equality on every non-Required field.
func TestLoadConfig_ExplicitOverridesAreRespected(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	body := `project_name: my-proj
description: an example
docs_path: docs/
scaffolding:
  - playbook
agents:
  tech-lead:
    workspace: docs/
    skills:
      - planning-template
    workflows: []
    protocols: []
    sensors: []
    routines: []
`
	cfgPath := writeYAML(t, tmp, "cfg.yaml", body)
	cfg, err := LoadConfig(cfgPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.ProjectName != "my-proj" {
		t.Errorf("ProjectName: want my-proj, got %q", cfg.ProjectName)
	}
	if cfg.Description != "an example" {
		t.Errorf("Description: want 'an example', got %q", cfg.Description)
	}
	if cfg.DocsPath != "docs/" {
		t.Errorf("DocsPath: want docs/, got %q", cfg.DocsPath)
	}
	if len(cfg.Scaffolding) != 1 || cfg.Scaffolding[0] != "playbook" {
		t.Errorf("Scaffolding: want [playbook], got %v", cfg.Scaffolding)
	}
	agent := cfg.Agents["tech-lead"]
	if agent == nil {
		t.Fatalf("tech-lead agent missing")
	}
	if !reflect.DeepEqual(agent.Skills, []string{"planning-template"}) {
		t.Errorf("Skills: want [planning-template], got %v", agent.Skills)
	}
	// Explicit empty list opts the user out of the category — must remain
	// empty (not get populated by agent.yaml defaults).
	if len(agent.Workflows) != 0 {
		t.Errorf("Workflows must remain empty; got %v", agent.Workflows)
	}
}

// TestApplyDefaults_TechLeadGetsDocsPath: pure-function test of the defaulting
// walk. tech-lead with no workspace gets DocsPath; non-tech-lead gets
// `<type>/`.
func TestApplyDefaults_TechLeadGetsDocsPath(t *testing.T) {
	cat := loadTestCatalog(t)
	cfg := &struct{}{}
	_ = cfg
	// Test through LoadConfig to keep this end-to-end with yaml parsing.
	tmp := t.TempDir()
	cfgPath := writeYAML(t, tmp, "cfg.yaml", "docs_path: workspace/\nagents:\n  tech-lead: {}\n")
	got, err := LoadConfig(cfgPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got.Agents["tech-lead"].Workspace != "workspace/" {
		t.Errorf("tech-lead workspace defaulted from DocsPath: want workspace/, got %q", got.Agents["tech-lead"].Workspace)
	}
}

// TestLoadOverlay_LeavesProjectFieldsEmpty: LoadOverlay must NOT default
// project_name / docs_path / scaffolding so the §3 "leave empty or match
// exactly" contract holds against the user's literal YAML. Critical for
// the bonsai add headless path — the runner compares overlay fields to
// the existing .bonsai.yaml, and a defaulted basename would always
// mismatch unless the user happens to invoke `bonsai add` from a dir
// whose basename equals the original project_name.
func TestLoadOverlay_LeavesProjectFieldsEmpty(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	cfgPath := writeYAML(t, tmp, "cfg.yaml", "agents:\n  backend: {}\n")
	overlay, err := LoadOverlay(cfgPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadOverlay: %v", err)
	}
	if overlay.ProjectName != "" {
		t.Errorf("ProjectName: want empty, got %q", overlay.ProjectName)
	}
	if overlay.DocsPath != "" {
		t.Errorf("DocsPath: want empty, got %q", overlay.DocsPath)
	}
	if len(overlay.Scaffolding) != 0 {
		t.Errorf("Scaffolding: want empty, got %v", overlay.Scaffolding)
	}
	// Per-agent defaults still apply.
	agent := overlay.Agents["backend"]
	if agent == nil {
		t.Fatalf("backend agent missing after LoadOverlay")
	}
	if agent.Workspace == "" {
		t.Errorf("Workspace must be defaulted; got empty")
	}
}

// TestLoadOverlay_StillValidatesWorkspaces: ../etc/passwd as a workspace
// must still be rejected by LoadOverlay's wsvalidate path even though
// project_name is allowed to be empty.
func TestLoadOverlay_StillValidatesWorkspaces(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	body := "agents:\n  backend:\n    workspace: ../etc/passwd\n"
	cfgPath := writeYAML(t, tmp, "cfg.yaml", body)
	_, err := LoadOverlay(cfgPath, tmp, cat)
	if err == nil {
		t.Fatalf("expected error for ../etc/passwd workspace")
	}
	if !strings.Contains(err.Error(), "escape") && !strings.Contains(err.Error(), "workspace") {
		t.Errorf("error must mention workspace/escape; got %q", err.Error())
	}
}

// TestLoadConfig_RoutineCheckSensorAutoAdded: an agent overlay with routines
// but no routine-check sensor must have routine-check appended by
// EnsureRoutineCheckSensor (mirrors interactive flow).
func TestLoadConfig_RoutineCheckSensorAutoAdded(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	body := `agents:
  tech-lead:
    routines:
      - backlog-hygiene
    sensors: []
`
	cfgPath := writeYAML(t, tmp, "cfg.yaml", body)
	cfg, err := LoadConfig(cfgPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	hasRoutineCheck := false
	for _, s := range cfg.Agents["tech-lead"].Sensors {
		if s == "routine-check" {
			hasRoutineCheck = true
			break
		}
	}
	if !hasRoutineCheck {
		t.Errorf("routine-check sensor must be auto-added when routines present; got sensors=%v", cfg.Agents["tech-lead"].Sensors)
	}
}
