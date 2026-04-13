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

	err = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat)
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
		"identity.md.tmpl": "I am {{ .AgentDisplayName }}",
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

	err = AgentWorkspace(tmpDir, agentDef, installed, cfg, cat)
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
