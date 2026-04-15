package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

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

	// Run again — files are unmodified, should be Updated
	var wr2 WriteResult
	_ = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat, lock, &wr2, false)
	for _, f := range wr2.Files {
		if f.Action != ActionUpdated {
			t.Errorf("file %s action = %d, want Updated", f.RelPath, f.Action)
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
		},
	}
	created, updated, skipped, conflicts := wr.Summary()
	if created != 2 {
		t.Errorf("created = %d, want 2", created)
	}
	if updated != 2 { // Updated + Forced
		t.Errorf("updated = %d, want 2", updated)
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
