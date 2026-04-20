package generate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
)

// buildTestCatalog creates a minimal in-memory FS with shared core + one agent.
func buildTestCatalog(agentCoreFiles map[string]string) (*catalog.Catalog, error) {
	fsys := fstest.MapFS{
		// Shared core files
		"core/memory.md.tmpl":    &fstest.MapFile{Data: []byte("memory for {{ .AgentDisplayName }}")},
		"core/self-awareness.md": &fstest.MapFile{Data: []byte("self-awareness content")},
		// Agent definition
		"agents/test-agent/agent.yaml": &fstest.MapFile{Data: []byte("name: test-agent\ndescription: test\n")},
	}
	// Agent-specific core files
	for name, content := range agentCoreFiles {
		fsys["agents/test-agent/core/"+name] = &fstest.MapFile{Data: []byte(content)}
	}

	return catalog.New(fsys)
}

func TestCoreFilesLayeredResolution(t *testing.T) {
	// Agent has identity only — shared memory + self-awareness should be used
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	err = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr, false)
	if err != nil {
		t.Fatalf("AgentWorkspace: %v", err)
	}

	coreDir := filepath.Join(tmpDir, "agent", "Core")

	// identity.md should come from agent (rendered)
	identity, err := os.ReadFile(filepath.Join(coreDir, "identity.md"))
	if err != nil {
		t.Fatalf("read identity.md: %v", err)
	}
	if !strings.Contains(string(identity), "I am Test Agent") {
		t.Errorf("identity.md unexpected content: %s", identity)
	}

	// memory.md should come from shared (rendered)
	memory, err := os.ReadFile(filepath.Join(coreDir, "memory.md"))
	if err != nil {
		t.Fatalf("read memory.md: %v", err)
	}
	if !strings.Contains(string(memory), "memory for Test Agent") {
		t.Errorf("memory.md unexpected content: %s", memory)
	}

	// self-awareness.md should come from shared (static)
	sa, err := os.ReadFile(filepath.Join(coreDir, "self-awareness.md"))
	if err != nil {
		t.Fatalf("read self-awareness.md: %v", err)
	}
	if string(sa) != "self-awareness content" {
		t.Errorf("self-awareness.md unexpected content: %s", sa)
	}
}

func TestCoreFilesAgentOverride(t *testing.T) {
	// Agent overrides self-awareness.md
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl":  "I am {{ .AgentDisplayName }}",
		"self-awareness.md": "custom self-awareness for this agent",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	err = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr, false)
	if err != nil {
		t.Fatalf("AgentWorkspace: %v", err)
	}

	coreDir := filepath.Join(tmpDir, "agent", "Core")

	// self-awareness should come from agent override, not shared
	sa, err := os.ReadFile(filepath.Join(coreDir, "self-awareness.md"))
	if err != nil {
		t.Fatalf("read self-awareness.md: %v", err)
	}
	if string(sa) != "custom self-awareness for this agent" {
		t.Errorf("self-awareness.md should be agent override, got: %s", sa)
	}

	// memory.md should still come from shared
	memory, err := os.ReadFile(filepath.Join(coreDir, "memory.md"))
	if err != nil {
		t.Fatalf("read memory.md: %v", err)
	}
	if !strings.Contains(string(memory), "memory for Test Agent") {
		t.Errorf("memory.md unexpected content: %s", memory)
	}
}

// ─── Lock-aware tests ─────────────────────────────────────────────────

func TestAgentWorkspaceNewFiles(t *testing.T) {
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	err = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr, false)
	if err != nil {
		t.Fatalf("AgentWorkspace: %v", err)
	}

	// All files should be ActionCreated
	for _, f := range wr.Files {
		if f.Action != ActionCreated {
			t.Errorf("file %s action = %d, want Created", f.RelPath, f.Action)
		}
	}

	// Lock should have entries for all written files
	if len(lock.Files) == 0 {
		t.Error("lock file should have tracked files")
	}
	if len(lock.Files) != len(wr.Files) {
		t.Errorf("lock entries (%d) != write results (%d)", len(lock.Files), len(wr.Files))
	}
}

func TestAgentWorkspaceUnmodifiedUpdate(t *testing.T) {
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr1 WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr1, false)

	// Run again — files are unmodified and content matches, should be Unchanged
	var wr2 WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr2, false)
	for _, f := range wr2.Files {
		if f.Action != ActionUnchanged {
			t.Errorf("file %s action = %d, want Unchanged", f.RelPath, f.Action)
		}
	}
}

func TestAgentWorkspaceModifiedConflict(t *testing.T) {
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr1 WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr1, false)

	// Modify identity.md
	identityPath := filepath.Join(tmpDir, "agent", "Core", "identity.md")
	_ = os.WriteFile(identityPath, []byte("user edited this"), 0644)

	// Run again — identity.md should be Conflict
	var wr2 WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr2, false)

	found := false
	for _, f := range wr2.Files {
		if strings.Contains(f.RelPath, "identity.md") {
			if f.Action != ActionConflict {
				t.Errorf("identity.md action = %d, want Conflict", f.Action)
			}
			found = true
		}
	}
	if !found {
		t.Error("identity.md not found in results")
	}

	// Verify file was NOT overwritten
	data, _ := os.ReadFile(identityPath)
	if !strings.Contains(string(data), "user edited") {
		t.Error("identity.md should not have been overwritten")
	}
}

func TestAgentWorkspaceForceOverwrite(t *testing.T) {
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr1 WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr1, false)

	// Modify identity.md
	identityPath := filepath.Join(tmpDir, "agent", "Core", "identity.md")
	_ = os.WriteFile(identityPath, []byte("user edited this"), 0644)

	// Force overwrite — should succeed
	var wr2 WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr2, true)
	for _, f := range wr2.Files {
		if strings.Contains(f.RelPath, "identity.md") {
			if f.Action != ActionForced {
				t.Errorf("identity.md action = %d, want Forced", f.Action)
			}
		}
	}

	// Verify file was actually overwritten
	data, _ := os.ReadFile(identityPath)
	if strings.Contains(string(data), "user edited") {
		t.Error("identity.md should have been overwritten")
	}
}

func TestForceConflictsReplay(t *testing.T) {
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr1 WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr1, false)

	// Modify identity.md
	identityPath := filepath.Join(tmpDir, "agent", "Core", "identity.md")
	_ = os.WriteFile(identityPath, []byte("user edited this"), 0644)

	// Run without force — get conflicts
	var wr2 WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr2, false)

	if !wr2.HasConflicts() {
		t.Fatal("expected conflicts")
	}

	// Now replay with ForceConflicts
	wr2.ForceConflicts(tmpDir, lock)

	// Verify the conflict was resolved
	for _, f := range wr2.Files {
		if strings.Contains(f.RelPath, "identity.md") {
			if f.Action != ActionForced {
				t.Errorf("after ForceConflicts, identity.md action = %d, want Forced", f.Action)
			}
		}
	}

	// Verify file was overwritten
	data, _ := os.ReadFile(identityPath)
	if strings.Contains(string(data), "user edited") {
		t.Error("identity.md should have been overwritten after ForceConflicts")
	}

	// Verify lock was updated
	exists, modified := lock.IsModified(tmpDir, "agent/Core/identity.md")
	if !exists {
		t.Error("identity.md should exist after ForceConflicts")
	}
	if modified {
		t.Error("identity.md should not be modified after ForceConflicts updated the lock")
	}
}

func TestWriteResultSummary(t *testing.T) {
	wr := WriteResult{
		Files: []FileResult{
			{Action: ActionCreated},
			{Action: ActionCreated},
			{Action: ActionUpdated},
			{Action: ActionSkipped},
			{Action: ActionConflict},
			{Action: ActionForced},
			{Action: ActionUnchanged},
		},
	}
	created, updated, unchanged, skipped, conflicts := wr.Summary()
	if created != 2 {
		t.Errorf("created = %d, want 2", created)
	}
	if updated != 2 { // Updated + Forced
		t.Errorf("updated = %d, want 2", updated)
	}
	if unchanged != 1 {
		t.Errorf("unchanged = %d, want 1", unchanged)
	}
	if skipped != 1 {
		t.Errorf("skipped = %d, want 1", skipped)
	}
	if conflicts != 1 {
		t.Errorf("conflicts = %d, want 1", conflicts)
	}
}

// ─── CLAUDE.md marker tests ──────────────────────────────────────────

func TestClaudeMDHasMarkers(t *testing.T) {
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr, false)

	claudeMD, err := os.ReadFile(filepath.Join(tmpDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(claudeMD)

	if !strings.Contains(content, "<!-- BONSAI_START -->") {
		t.Error("CLAUDE.md missing start marker")
	}
	if !strings.Contains(content, "<!-- BONSAI_END -->") {
		t.Error("CLAUDE.md missing end marker")
	}
}

func TestClaudeMDPreservesUserContent(t *testing.T) {
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr1 WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr1, false)

	// Append user content after end marker
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	existing, _ := os.ReadFile(claudePath)
	userContent := "\n\n### My Custom Section\n\nUser-added content here.\n"
	_ = os.WriteFile(claudePath, append(existing, []byte(userContent)...), 0644)

	// Run again — markers should be found, user content preserved
	var wr2 WriteResult
	_ = WorkspaceClaudeMD(tmpDir, tmpDir, agentDef, installed, cfg, cat, lock, &wr2, false)

	updated, _ := os.ReadFile(claudePath)
	if !strings.Contains(string(updated), "User-added content here.") {
		t.Error("user content after end marker was not preserved")
	}
	if !strings.Contains(string(updated), "<!-- BONSAI_START -->") {
		t.Error("start marker missing after update")
	}
}

func TestClaudeMDMigratesMarkerlessFile(t *testing.T) {
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	// Write a CLAUDE.md without markers — simulates user-customized file
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	oldContent := "# Custom Project — Test Agent\n\nCustom content without markers.\n\n## Navigation\n\nOld nav tables here.\n"
	if err := os.WriteFile(claudePath, []byte(oldContent), 0644); err != nil {
		t.Fatalf("write markerless CLAUDE.md: %v", err)
	}

	lock := config.NewLockFile()
	var wr WriteResult
	err = WorkspaceClaudeMD(tmpDir, tmpDir, agentDef, installed, cfg, cat, lock, &wr, false)
	if err != nil {
		t.Fatalf("WorkspaceClaudeMD: %v", err)
	}

	// Assert: backup file was created with old content
	bakContent, err := os.ReadFile(claudePath + ".bak")
	if err != nil {
		t.Fatalf("CLAUDE.md.bak not created: %v", err)
	}
	if !strings.Contains(string(bakContent), "Custom content without markers") {
		t.Error("backup file does not contain original content")
	}

	// Assert: new file has markers and correct content
	updated, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read updated CLAUDE.md: %v", err)
	}
	content := string(updated)

	if !strings.Contains(content, "<!-- BONSAI_START -->") {
		t.Error("CLAUDE.md missing start marker after migration")
	}
	if !strings.Contains(content, "<!-- BONSAI_END -->") {
		t.Error("CLAUDE.md missing end marker after migration")
	}
	if !strings.Contains(content, "TestProject") {
		t.Error("CLAUDE.md does not contain project name after migration")
	}
}

func TestClaudeMDIncludesCustomItems(t *testing.T) {
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{
		AgentType: "test-agent",
		Workspace: ".",
		Workflows: []string{"my-custom-wf"},
		CustomItems: map[string]*config.CustomItemMeta{
			"my-custom-wf": {
				Description: "A custom workflow for testing",
				DisplayName: "My Custom WF",
			},
		},
	}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	_ = WorkspaceClaudeMD(tmpDir, tmpDir, agentDef, installed, cfg, cat, lock, &wr, false)

	claudeMD, _ := os.ReadFile(filepath.Join(tmpDir, "CLAUDE.md"))
	content := string(claudeMD)

	if !strings.Contains(content, "my-custom-wf") {
		t.Error("CLAUDE.md does not include custom workflow filename")
	}
	if !strings.Contains(content, "A custom workflow for testing") {
		t.Error("CLAUDE.md does not include custom workflow description")
	}
}

// ─── How to Work tests ──────────────────────────────────────────────

func TestHowToWorkSectionExists(t *testing.T) {
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr, false)

	claudeMD, err := os.ReadFile(filepath.Join(tmpDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(claudeMD)

	if !strings.Contains(content, "### How to Work") {
		t.Error("CLAUDE.md missing '### How to Work' section")
	}
	if !strings.Contains(content, "Decision heuristics") {
		t.Error("CLAUDE.md missing How to Work description line")
	}
}

func TestHowToWorkTechLeadHeuristics(t *testing.T) {
	lines := howToWorkLines("tech-lead", "station/", true, false)
	content := strings.Join(lines, "\n")

	if !strings.Contains(content, "orchestrate") {
		t.Error("tech-lead heuristics should contain 'orchestrate'")
	}
	if !strings.Contains(content, "Backlog") {
		t.Error("tech-lead heuristics should contain 'Backlog'")
	}
}

func TestHowToWorkCodeAgentHeuristics(t *testing.T) {
	lines := howToWorkLines("backend", "docs/", false, false)
	content := strings.Join(lines, "\n")

	if !strings.Contains(content, "Plan first") {
		t.Error("backend heuristics should contain 'Plan first'")
	}
	if !strings.Contains(content, "scope") {
		t.Error("backend heuristics should contain 'scope'")
	}
	// Should NOT contain tech-lead heuristics
	if strings.Contains(content, "orchestrate") {
		t.Error("backend heuristics should not contain tech-lead 'orchestrate'")
	}
}

func TestHowToWorkGuidePointer(t *testing.T) {
	// With workspace-guide installed
	lines := howToWorkLines("backend", "", false, true)
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "workspace-guide.md") {
		t.Error("workspace-guide pointer should appear when hasWorkspaceGuide is true")
	}

	// Without workspace-guide installed
	lines = howToWorkLines("backend", "", false, false)
	content = strings.Join(lines, "\n")
	if strings.Contains(content, "workspace-guide.md") {
		t.Error("workspace-guide pointer should NOT appear when hasWorkspaceGuide is false")
	}
}

// ─── Trigger tests ──────────────────────────────────────────────────

// buildTestCatalogWithItems creates a catalog with skills and workflows for trigger testing.
func buildTestCatalogWithItems(extraFiles map[string]string) (*catalog.Catalog, error) {
	fsys := fstest.MapFS{
		// Shared core files
		"core/memory.md.tmpl":    &fstest.MapFile{Data: []byte("memory for {{ .AgentDisplayName }}")},
		"core/self-awareness.md": &fstest.MapFile{Data: []byte("self-awareness content")},
		// Agent definition
		"agents/test-agent/agent.yaml":            &fstest.MapFile{Data: []byte("name: test-agent\ndescription: test\n")},
		"agents/test-agent/core/identity.md.tmpl": &fstest.MapFile{Data: []byte("I am {{ .AgentDisplayName }}")},
	}
	for k, v := range extraFiles {
		fsys[k] = &fstest.MapFile{Data: []byte(v)}
	}
	return catalog.New(fsys)
}

func TestScenariosDescFallback(t *testing.T) {
	// nil triggers → falls back to description
	item := &catalog.CatalogItem{Description: "A test description"}
	result := scenariosDesc(item)
	if result != "A test description" {
		t.Errorf("expected description fallback, got %q", result)
	}

	// non-nil triggers with scenarios → uses scenarios
	item.Triggers = &catalog.Triggers{
		Scenarios: []string{"When doing X"},
	}
	result = scenariosDesc(item)
	if result != "When doing X" {
		t.Errorf("expected scenario, got %q", result)
	}

	// nil item → empty string
	result = scenariosDesc(nil)
	if result != "" {
		t.Errorf("expected empty string for nil item, got %q", result)
	}
}

func TestScenariosDescJoinsScenarios(t *testing.T) {
	item := &catalog.CatalogItem{
		Description: "fallback",
		Triggers: &catalog.Triggers{
			Scenarios: []string{"Scenario A", "Scenario B", "Scenario C"},
		},
	}
	result := scenariosDesc(item)
	expected := "Scenario A; Scenario B; Scenario C"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestClaudeMDUsesScenarios(t *testing.T) {
	cat, err := buildTestCatalogWithItems(map[string]string{
		"skills/test-skill/meta.yaml":     "name: test-skill\ndescription: A test skill\nagents: all\ntriggers:\n  scenarios:\n    - Testing trigger scenarios\n  paths:\n    - \"*.test\"\n",
		"skills/test-skill/test-skill.md": "# Test Skill\n\nSkill content here.\n",
		"workflows/planning/meta.yaml":    "name: planning\ndescription: End-to-end planning\nagents: all\ntriggers:\n  scenarios:\n    - Starting end-to-end planning\n    - Translating requirements into a plan\n  examples:\n    - prompt: \"Plan the caching layer\"\n      action: \"Load planning workflow\"\n",
		"workflows/planning/planning.md":  "# Planning Workflow\n\nWorkflow content here.\n",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{
		AgentType: "test-agent",
		Workspace: ".",
		Skills:    []string{"test-skill"},
		Workflows: []string{"planning"},
	}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	_ = WorkspaceClaudeMD(tmpDir, tmpDir, agentDef, installed, cfg, cat, lock, &wr, false)

	claudeMD, err := os.ReadFile(filepath.Join(tmpDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(claudeMD)

	if !strings.Contains(content, "Activate when...") {
		t.Error("CLAUDE.md should have 'Activate when...' header")
	}
	if !strings.Contains(content, "Testing trigger scenarios") {
		t.Error("CLAUDE.md should contain skill scenario text")
	}
	if !strings.Contains(content, "Starting end-to-end planning") {
		t.Error("CLAUDE.md should contain workflow scenario text")
	}
}

func TestPathScopedRulesGenerated(t *testing.T) {
	cat, err := buildTestCatalogWithItems(map[string]string{
		"skills/test-skill/meta.yaml":     "name: test-skill\ndescription: A test skill\nagents: all\ntriggers:\n  scenarios:\n    - Testing trigger scenarios\n  paths:\n    - \"*.test\"\n    - \"**/test/**\"\n",
		"skills/test-skill/test-skill.md": "# Test Skill\n\nSkill content here.\n",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	installed := &config.InstalledAgent{
		AgentType: "test-agent",
		Workspace: "station/",
		Skills:    []string{"test-skill"},
	}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	err = PathScopedRules(tmpDir, cfg, cat, lock, &wr, false)
	if err != nil {
		t.Fatalf("PathScopedRules: %v", err)
	}

	rulePath := filepath.Join(tmpDir, "station", ".claude", "rules", "skill-test-skill.md")
	data, err := os.ReadFile(rulePath)
	if err != nil {
		t.Fatalf("rule file not created: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "paths:") {
		t.Error("rule file should contain paths: frontmatter")
	}
	if !strings.Contains(content, "*.test") {
		t.Error("rule file should contain the path glob")
	}
	if !strings.Contains(content, "Testing trigger scenarios") {
		t.Error("rule file should contain scenario text")
	}
}

func TestPathScopedRulesSkippedWhenNoPaths(t *testing.T) {
	cat, err := buildTestCatalogWithItems(map[string]string{
		"skills/no-paths/meta.yaml":   "name: no-paths\ndescription: A skill without paths\nagents: all\ntriggers:\n  scenarios:\n    - No paths skill\n",
		"skills/no-paths/no-paths.md": "# No Paths Skill\n\nContent.\n",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	installed := &config.InstalledAgent{
		AgentType: "test-agent",
		Workspace: "station/",
		Skills:    []string{"no-paths"},
	}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	_ = PathScopedRules(tmpDir, cfg, cat, lock, &wr, false)

	rulePath := filepath.Join(tmpDir, "station", ".claude", "rules", "skill-no-paths.md")
	if _, err := os.Stat(rulePath); err == nil {
		t.Error("rule file should NOT be created for skill without paths")
	}
}

func TestWorkflowSkillsGenerated(t *testing.T) {
	cat, err := buildTestCatalogWithItems(map[string]string{
		"workflows/planning/meta.yaml":   "name: planning\ndescription: End-to-end planning\nagents: all\ntriggers:\n  scenarios:\n    - Starting end-to-end planning\n  examples:\n    - prompt: \"Plan the caching layer\"\n      action: \"Load planning workflow\"\n",
		"workflows/planning/planning.md": "# Planning Workflow\n\nWorkflow content here.\n",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	installed := &config.InstalledAgent{
		AgentType: "test-agent",
		Workspace: "station/",
		Workflows: []string{"planning"},
	}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	err = WorkflowSkills(tmpDir, cfg, cat, lock, &wr, false)
	if err != nil {
		t.Fatalf("WorkflowSkills: %v", err)
	}

	skillPath := filepath.Join(tmpDir, "station", ".claude", "skills", "planning", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("SKILL.md not created: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "name: planning") {
		t.Error("SKILL.md should contain workflow name")
	}
	if !strings.Contains(content, "Starting end-to-end planning") {
		t.Error("SKILL.md should contain scenario description")
	}
	if !strings.Contains(content, "Plan the caching layer") {
		t.Error("SKILL.md should contain example prompt")
	}
}

func TestWorkflowSkillsSkippedWhenNotCurated(t *testing.T) {
	cat, err := buildTestCatalogWithItems(map[string]string{
		"workflows/session-logging/meta.yaml":          "name: session-logging\ndescription: End-of-session log\nagents: all\ntriggers:\n  scenarios:\n    - Logging session\n",
		"workflows/session-logging/session-logging.md": "# Session Logging\n\nContent.\n",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	installed := &config.InstalledAgent{
		AgentType: "test-agent",
		Workspace: "station/",
		Workflows: []string{"session-logging"},
	}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	_ = WorkflowSkills(tmpDir, cfg, cat, lock, &wr, false)

	skillPath := filepath.Join(tmpDir, "station", ".claude", "skills", "session-logging", "SKILL.md")
	if _, err := os.Stat(skillPath); err == nil {
		t.Error("SKILL.md should NOT be created for non-curated workflow")
	}
}

func TestTriggerSectionPrepended(t *testing.T) {
	cat, err := buildTestCatalogWithItems(map[string]string{
		"skills/test-skill/meta.yaml":     "name: test-skill\ndescription: A test skill\nagents: all\ntriggers:\n  scenarios:\n    - Testing trigger scenarios\n  paths:\n    - \"*.test\"\n",
		"skills/test-skill/test-skill.md": "# Test Skill\n\nSkill content here.\n",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{
		AgentType: "test-agent",
		Workspace: ".",
		Skills:    []string{"test-skill"},
	}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	err = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr, false)
	if err != nil {
		t.Fatalf("AgentWorkspace: %v", err)
	}

	skillPath := filepath.Join(tmpDir, "agent", "Skills", "test-skill.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	content := string(data)

	if !strings.HasPrefix(content, "## Triggers") {
		t.Error("skill file should start with '## Triggers' section")
	}
	if !strings.Contains(content, "Testing trigger scenarios") {
		t.Error("skill file should contain scenario text in trigger section")
	}
	if !strings.Contains(content, "# Test Skill") {
		t.Error("skill file should still contain original content after trigger section")
	}
}

func TestBackwardCompatNilTriggers(t *testing.T) {
	// Build catalog with NO triggers — should work fine
	cat, err := buildTestCatalogWithItems(map[string]string{
		"skills/plain-skill/meta.yaml":      "name: plain-skill\ndescription: A plain skill\nagents: all\n",
		"skills/plain-skill/plain-skill.md": "# Plain Skill\n\nContent.\n",
		"workflows/plain-wf/meta.yaml":      "name: plain-wf\ndescription: A plain workflow\nagents: all\n",
		"workflows/plain-wf/plain-wf.md":    "# Plain Workflow\n\nContent.\n",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{
		AgentType: "test-agent",
		Workspace: ".",
		Skills:    []string{"plain-skill"},
		Workflows: []string{"plain-wf"},
	}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult

	// AgentWorkspace should not crash
	err = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr, false)
	if err != nil {
		t.Fatalf("AgentWorkspace with nil triggers should not error: %v", err)
	}

	// PathScopedRules should not crash or create files
	var wr2 WriteResult
	err = PathScopedRules(tmpDir, cfg, cat, lock, &wr2, false)
	if err != nil {
		t.Fatalf("PathScopedRules with nil triggers should not error: %v", err)
	}
	for _, f := range wr2.Files {
		if strings.Contains(f.RelPath, "rules") {
			t.Error("should not create rule files when no paths triggers")
		}
	}

	// WorkflowSkills should not crash (plain-wf is not curated)
	var wr3 WriteResult
	err = WorkflowSkills(tmpDir, cfg, cat, lock, &wr3, false)
	if err != nil {
		t.Fatalf("WorkflowSkills with nil triggers should not error: %v", err)
	}

	// Verify skill file does NOT have trigger section
	skillPath := filepath.Join(tmpDir, "agent", "Skills", "plain-skill.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	if strings.HasPrefix(string(data), "## Triggers") {
		t.Error("skill file without triggers should not start with trigger section")
	}
}

// TestWorkspaceClaudeMDUnchangedShortCircuit verifies that calling
// WorkspaceClaudeMD twice with identical inputs hits the short-circuit at
// generate.go:829-833 — the second call records ActionUnchanged and does
// not rewrite the file (mtime is preserved).
func TestWorkspaceClaudeMDUnchangedShortCircuit(t *testing.T) {
	cat, err := buildTestCatalog(map[string]string{
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
	})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{AgentType: "test-agent", Workspace: "."}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()

	// First call — creates CLAUDE.md with markers.
	var wr1 WriteResult
	if err := WorkspaceClaudeMD(tmpDir, tmpDir, agentDef, installed, cfg, cat, lock, &wr1, false); err != nil {
		t.Fatalf("first WorkspaceClaudeMD: %v", err)
	}

	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	info1, err := os.Stat(claudePath)
	if err != nil {
		t.Fatalf("stat CLAUDE.md after first call: %v", err)
	}
	mtime1 := info1.ModTime()

	// Back-date the file so that a rewrite (if it happened) would be
	// detectable via mtime even on filesystems with coarse timestamps.
	past := mtime1.Add(-2 * time.Second)
	if err := os.Chtimes(claudePath, past, past); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	info1, err = os.Stat(claudePath)
	if err != nil {
		t.Fatalf("stat after chtimes: %v", err)
	}
	mtime1 = info1.ModTime()

	// Second call — identical inputs, should short-circuit.
	var wr2 WriteResult
	if err := WorkspaceClaudeMD(tmpDir, tmpDir, agentDef, installed, cfg, cat, lock, &wr2, false); err != nil {
		t.Fatalf("second WorkspaceClaudeMD: %v", err)
	}

	if len(wr2.Files) == 0 {
		t.Fatal("second WorkspaceClaudeMD produced no FileResults")
	}
	last := wr2.Files[len(wr2.Files)-1]
	if last.Action != ActionUnchanged {
		t.Errorf("last FileResult.Action = %v, want ActionUnchanged", last.Action)
	}

	info2, err := os.Stat(claudePath)
	if err != nil {
		t.Fatalf("stat CLAUDE.md after second call: %v", err)
	}
	if !info2.ModTime().Equal(mtime1) {
		t.Errorf("CLAUDE.md mtime changed: before=%v after=%v (expected no rewrite)", mtime1, info2.ModTime())
	}
}

// TestWriteFileChmodRestoresPermOnUnchanged is a regression test for the
// Plan 13 fix: when content is unchanged but the file's exec bit was
// stripped externally, writeFileChmod must still re-apply the declared
// perm. Previously the chmod gate excluded ActionUnchanged, so sensor
// scripts stayed non-executable across bonsai update runs.
func TestWriteFileChmodRestoresPermOnUnchanged(t *testing.T) {
	tmpDir := t.TempDir()
	relPath := "agent/Sensors/test.sh"
	content := []byte("#!/bin/sh\necho hi\n")
	lock := config.NewLockFile()

	// 1. Create the file with 0755.
	r1 := writeFileChmod(tmpDir, relPath, content, "catalog:sensors/test", lock, false, 0755)
	if r1.Action != ActionCreated {
		t.Fatalf("first writeFileChmod action = %v, want ActionCreated", r1.Action)
	}
	absPath := filepath.Join(tmpDir, relPath)
	info, err := os.Stat(absPath)
	if err != nil {
		t.Fatalf("stat after create: %v", err)
	}
	if info.Mode()&0777 != 0755 {
		t.Fatalf("mode after create = %v, want 0755", info.Mode()&0777)
	}

	// 2. Externally strip the exec bit.
	if err := os.Chmod(absPath, 0644); err != nil {
		t.Fatalf("chmod 0644: %v", err)
	}

	// 3. Call again with identical content + perm.
	r2 := writeFileChmod(tmpDir, relPath, content, "catalog:sensors/test", lock, false, 0755)

	// 4. Should report ActionUnchanged (content is identical).
	if r2.Action != ActionUnchanged {
		t.Errorf("second writeFileChmod action = %v, want ActionUnchanged", r2.Action)
	}

	// 5. Mode must be restored to 0755.
	info, err = os.Stat(absPath)
	if err != nil {
		t.Fatalf("stat after second call: %v", err)
	}
	if info.Mode()&0777 != 0755 {
		t.Errorf("mode after unchanged run = %v, want 0755 (perm should be restored)", info.Mode()&0777)
	}
}

// TestShellScriptLF verifies that sensor scripts written during AgentWorkspace
// generation contain no carriage-return bytes, regardless of the source CRLF
// state. Covers the normalizeShellLF belt-and-braces defence for Step 1 of
// Plan 19 — protects against CRLF sneaking in via a git client that ignores
// .gitattributes.
func TestShellScriptLF(t *testing.T) {
	// Build a catalog with shared core, one agent, and one sensor whose
	// script template contains CRLF line endings.
	fsys := fstest.MapFS{
		"core/memory.md.tmpl":                     &fstest.MapFile{Data: []byte("memory for {{ .AgentDisplayName }}")},
		"core/self-awareness.md":                  &fstest.MapFile{Data: []byte("self-awareness content")},
		"agents/test-agent/agent.yaml":            &fstest.MapFile{Data: []byte("name: test-agent\ndescription: test\n")},
		"agents/test-agent/core/identity.md.tmpl": &fstest.MapFile{Data: []byte("I am {{ .AgentDisplayName }}")},
		"sensors/crlf-sensor/meta.yaml": &fstest.MapFile{Data: []byte(
			"name: crlf-sensor\ndescription: test sensor\nagents: all\nevent: SessionStart\n",
		)},
		// Script template with CRLF line endings — simulates a hostile checkout.
		"sensors/crlf-sensor/crlf-sensor.sh.tmpl": &fstest.MapFile{Data: []byte(
			"#!/bin/bash\r\necho {{ .AgentName }}\r\nexit 0\r\n",
		)},
	}
	cat, err := catalog.New(fsys)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	tmpDir := t.TempDir()
	agentDef := cat.GetAgent("test-agent")
	installed := &config.InstalledAgent{
		AgentType: "test-agent",
		Workspace: ".",
		Sensors:   []string{"crlf-sensor"},
	}
	cfg := &config.ProjectConfig{
		ProjectName: "TestProject",
		Agents:      map[string]*config.InstalledAgent{"test-agent": installed},
	}

	lock := config.NewLockFile()
	var wr WriteResult
	if err := AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("AgentWorkspace: %v", err)
	}

	sensorsDir := filepath.Join(tmpDir, "agent", "Sensors")
	entries, err := os.ReadDir(sensorsDir)
	if err != nil {
		t.Fatalf("read sensors dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no sensor scripts written — AgentWorkspace did not install crlf-sensor")
	}

	foundSh := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sh") {
			continue
		}
		foundSh = true
		path := filepath.Join(sensorsDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if bytes.Contains(data, []byte("\r")) {
			t.Errorf("sensor script %s contains carriage return bytes; want LF-only", e.Name())
		}
	}
	if !foundSh {
		t.Fatal("no .sh files found under agent/Sensors")
	}
}

func TestInjectTriggerSection(t *testing.T) {
	ts := "## Triggers\n\n**Slash command:** `/foo`\n\n---\n\n"
	tests := []struct {
		name    string
		ts      string
		content string
		want    string
	}{
		{
			name:    "empty ts returns content unchanged",
			ts:      "",
			content: "---\nfoo: bar\n---\n# Title\n",
			want:    "---\nfoo: bar\n---\n# Title\n",
		},
		{
			name:    "no frontmatter prepends as before",
			ts:      ts,
			content: "# Title\nbody\n",
			want:    ts + "# Title\nbody\n",
		},
		{
			name:    "frontmatter present: ts lands after closing ---",
			ts:      ts,
			content: "---\nfoo: bar\n---\n# Title\nbody\n",
			want:    "---\nfoo: bar\n---\n" + ts + "# Title\nbody\n",
		},
		{
			name:    "opens with --- but no closing fence: prepend",
			ts:      ts,
			content: "---\nfoo: bar\n# Title\n",
			want:    ts + "---\nfoo: bar\n# Title\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := string(injectTriggerSection(tc.ts, []byte(tc.content)))
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
