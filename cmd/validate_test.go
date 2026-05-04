package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/validate"
)

// TestValidate_RenderText_NoIssues verifies the human-readable renderer's
// clean-state branch. Assertion is on the user-facing "No issues found"
// + the agent-count breadcrumb.
func TestValidate_RenderText_NoIssues(t *testing.T) {
	report := &validate.Report{
		Issues:        nil,
		AgentsScanned: []string{"tech-lead", "backend"},
	}
	out := captureStdout(t, func() { renderValidateText(report) })
	if !strings.Contains(out, "No issues found") {
		t.Fatalf("expected 'No issues found' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "2 agent") {
		t.Fatalf("expected '2 agent' breadcrumb, got:\n%s", out)
	}
}

// TestValidate_JSONOutput_Shape exercises renderValidateJSON end-to-end:
// build a Report with one orphan issue, marshal, parse the stdout back
// through encoding/json, and assert the stable schema. Locks the
// public JSON shape that downstream agents will rely on.
func TestValidate_JSONOutput_Shape(t *testing.T) {
	report := &validate.Report{
		AgentsScanned: []string{"tech-lead"},
		Issues: []validate.Issue{
			{
				Category:    validate.CategoryOrphanedRegistration,
				Severity:    validate.SeverityError,
				AgentName:   "tech-lead",
				AbilityType: "skill",
				Name:        "foo",
				Path:        "station/agent/Skills/foo.md",
				Detail:      "registered but lock entry missing — run `bonsai update` to recover",
			},
		},
	}
	out := captureStdout(t, func() {
		if err := renderValidateJSON(report); err != nil {
			t.Fatalf("renderValidateJSON: %v", err)
		}
	})

	var snapshot struct {
		Issues []struct {
			Category    string `json:"category"`
			Severity    string `json:"severity"`
			Agent       string `json:"agent"`
			AbilityType string `json:"ability_type"`
			Name        string `json:"name"`
			Path        string `json:"path"`
			Detail      string `json:"detail"`
		} `json:"issues"`
		AgentsScanned []string `json:"agents_scanned"`
	}
	if err := json.Unmarshal([]byte(out), &snapshot); err != nil {
		t.Fatalf("parse JSON: %v\n%s", err, out)
	}
	if len(snapshot.Issues) != 1 {
		t.Fatalf("issues len = %d, want 1", len(snapshot.Issues))
	}
	got := snapshot.Issues[0]
	if got.Category != "orphaned_registration" || got.Severity != "error" ||
		got.Agent != "tech-lead" || got.AbilityType != "skill" || got.Name != "foo" {
		t.Fatalf("issue payload mismatch: %+v", got)
	}
	if len(snapshot.AgentsScanned) != 1 || snapshot.AgentsScanned[0] != "tech-lead" {
		t.Fatalf("agents_scanned = %v, want [tech-lead]", snapshot.AgentsScanned)
	}
}

// TestValidate_E2E_OrphanProject is a thin smoke test that wires
// validate.Run through a temp project. Mirrors the sandbox `orphan`
// scenario from the plan's Verification section. Ensures the package
// composes correctly: config.Load → validate.Run → Report.HasIssues.
func TestValidate_E2E_OrphanProject(t *testing.T) {
	tmp := t.TempDir()
	for _, sub := range []string{"agent/Skills", "agent/Workflows", "agent/Protocols", "agent/Sensors", "agent/Routines"} {
		if err := os.MkdirAll(filepath.Join(tmp, "station", sub), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	// Drop a custom skill with valid frontmatter but DO NOT register it
	// in the lock or custom_items — Plan 34's repro pattern.
	body := "---\ndescription: orphan skill\n---\nbody\n"
	if err := os.WriteFile(filepath.Join(tmp, "station/agent/Skills/foo.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg := &config.ProjectConfig{
		ProjectName: "demo",
		Agents: map[string]*config.InstalledAgent{
			"tech-lead": {
				AgentType: "tech-lead",
				Workspace: "station",
				Skills:    []string{"foo"},
			},
		},
	}
	if err := cfg.Save(filepath.Join(tmp, configFile)); err != nil {
		t.Fatalf("save cfg: %v", err)
	}

	loaded, err := config.Load(filepath.Join(tmp, configFile))
	if err != nil {
		t.Fatalf("load cfg: %v", err)
	}
	report, err := validate.Run(tmp, loaded, nil, config.NewLockFile(), "")
	if err != nil {
		t.Fatalf("validate.Run: %v", err)
	}
	if !report.HasIssues() {
		t.Fatalf("expected issues for orphan project, got none")
	}
	if !report.HasErrors() {
		t.Fatalf("expected an error-level issue (orphan)")
	}
}
