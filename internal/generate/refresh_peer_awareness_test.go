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

// buildPeerTestCatalog creates an in-memory catalog with two agent types
// (lead, peer-a, peer-b) plus scope-guard-files and dispatch-guard sensors
// whose templates range over OtherAgents. identity.md.tmpl is shared.
func buildPeerTestCatalog() (*catalog.Catalog, error) {
	scopeGuard := `#!/usr/bin/env bash
# Blocks {{ .AgentDisplayName }} from editing files outside {{ .Workspace }}
input=$(cat)
{{ range .OtherAgents }}
# Block writes to {{ .Workspace }}
if [[ "$file_path" == *"{{ .Workspace }}"* ]]; then
  echo "BLOCKED: {{ $.AgentDisplayName }} cannot modify {{ .Workspace }}"
  exit 2
fi
{{ end }}
exit 0
`
	dispatchGuard := `#!/usr/bin/env bash
# Dispatch guard for {{ .AgentDisplayName }}
declare -A workspaces
{{ range .OtherAgents }}
workspaces["{{ .Workspace }}"]="{{ .AgentType }}"
{{ end }}
exit 0
`
	identity := `# {{ .AgentDisplayName }}
Peers:
{{ range .OtherAgents }}
- {{ .AgentType }} @ {{ .Workspace }}
{{ end }}
`

	fsys := fstest.MapFS{
		"core/memory.md.tmpl":                                 &fstest.MapFile{Data: []byte("memory")},
		"core/self-awareness.md":                              &fstest.MapFile{Data: []byte("sa")},
		"core/identity.md.tmpl":                               &fstest.MapFile{Data: []byte(identity)},
		"agents/lead/agent.yaml":                              &fstest.MapFile{Data: []byte("name: lead\ndescription: lead agent\n")},
		"agents/peer-a/agent.yaml":                            &fstest.MapFile{Data: []byte("name: peer-a\ndescription: peer a\n")},
		"agents/peer-b/agent.yaml":                            &fstest.MapFile{Data: []byte("name: peer-b\ndescription: peer b\n")},
		"sensors/scope-guard-files/meta.yaml":                 &fstest.MapFile{Data: []byte("name: scope-guard-files\ndescription: scope guard\nevent: PreToolUse\nmatcher: Edit\nagents: all\n")},
		"sensors/scope-guard-files/scope-guard-files.sh.tmpl": &fstest.MapFile{Data: []byte(scopeGuard)},
		"sensors/dispatch-guard/meta.yaml":                    &fstest.MapFile{Data: []byte("name: dispatch-guard\ndescription: dispatch guard\nevent: PreToolUse\nmatcher: Agent\nagents: all\n")},
		"sensors/dispatch-guard/dispatch-guard.sh.tmpl":       &fstest.MapFile{Data: []byte(dispatchGuard)},
	}
	return catalog.New(fsys)
}

// setupPeerFixture installs two agents (lead + peer-a) through the full
// AgentWorkspace pipeline so the filesystem reflects a real multi-agent
// project. Returns the catalog, cfg, lock, and tmpDir for test bodies.
func setupPeerFixture(t *testing.T) (*catalog.Catalog, *config.ProjectConfig, *config.LockFile, string) {
	t.Helper()
	cat, err := buildPeerTestCatalog()
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	tmpDir := t.TempDir()

	lead := &config.InstalledAgent{
		AgentType: "lead",
		Workspace: "station/",
		Sensors:   []string{"scope-guard-files", "dispatch-guard"},
	}
	peerA := &config.InstalledAgent{
		AgentType: "peer-a",
		Workspace: "peer-a/",
		Sensors:   []string{"scope-guard-files", "dispatch-guard"},
	}
	cfg := &config.ProjectConfig{
		ProjectName: "Test",
		Agents: map[string]*config.InstalledAgent{
			"lead":   lead,
			"peer-a": peerA,
		},
	}
	lock := config.NewLockFile()
	var wr WriteResult
	if err := AgentWorkspace(tmpDir, cat.GetAgent("lead"), lead, cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("lead workspace: %v", err)
	}
	if err := AgentWorkspace(tmpDir, cat.GetAgent("peer-a"), peerA, cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("peer-a workspace: %v", err)
	}
	return cat, cfg, lock, tmpDir
}

// TestRefreshPeerAwareness_UpdatesSiblingScopeGuard — install lead + peer-a,
// then add peer-b to cfg and call Refresh excluding peer-b. Asserts the lead
// and peer-a scope-guard scripts now reference peer-b's workspace.
func TestRefreshPeerAwareness_UpdatesSiblingScopeGuard(t *testing.T) {
	cat, cfg, lock, tmpDir := setupPeerFixture(t)

	// Simulate `bonsai add peer-b`: register in cfg, run AgentWorkspace for it,
	// then call Refresh. Before the refresh, lead's scope-guard does NOT yet
	// reference peer-b/ (because AgentWorkspace only renders peer-b's own files).
	peerB := &config.InstalledAgent{
		AgentType: "peer-b",
		Workspace: "peer-b/",
		Sensors:   []string{"scope-guard-files", "dispatch-guard"},
	}
	cfg.Agents["peer-b"] = peerB
	var wr WriteResult
	if err := AgentWorkspace(tmpDir, cat.GetAgent("peer-b"), peerB, cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("peer-b workspace: %v", err)
	}

	// Sanity: lead's scope-guard currently references peer-a/ but NOT peer-b/.
	leadGuardPath := filepath.Join(tmpDir, "station", "agent", "Sensors", "scope-guard-files.sh")
	before, err := os.ReadFile(leadGuardPath)
	if err != nil {
		t.Fatalf("read lead guard: %v", err)
	}
	if !strings.Contains(string(before), "peer-a/") {
		t.Fatalf("lead guard should reference peer-a/ before refresh: %s", before)
	}
	if strings.Contains(string(before), "peer-b/") {
		t.Fatalf("lead guard should NOT reference peer-b/ before refresh: %s", before)
	}

	// Exercise: Refresh peers (excluding peer-b — the newly-added agent).
	if err := RefreshPeerAwareness(tmpDir, "peer-b", cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("RefreshPeerAwareness: %v", err)
	}

	after, err := os.ReadFile(leadGuardPath)
	if err != nil {
		t.Fatalf("read lead guard after: %v", err)
	}
	if !strings.Contains(string(after), "peer-b/") {
		t.Errorf("lead guard missing peer-b/ after refresh: %s", after)
	}

	peerAGuardPath := filepath.Join(tmpDir, "peer-a", "agent", "Sensors", "scope-guard-files.sh")
	peerAAfter, err := os.ReadFile(peerAGuardPath)
	if err != nil {
		t.Fatalf("read peer-a guard: %v", err)
	}
	if !strings.Contains(string(peerAAfter), "peer-b/") {
		t.Errorf("peer-a guard missing peer-b/ after refresh: %s", peerAAfter)
	}
}

// TestRefreshPeerAwareness_SkipsExcludedAgent — the excluded agent's files
// are NOT touched by the refresh pass. Verifies by checking WriteResult does
// not include any entries under the excluded workspace.
func TestRefreshPeerAwareness_SkipsExcludedAgent(t *testing.T) {
	cat, cfg, lock, tmpDir := setupPeerFixture(t)

	peerB := &config.InstalledAgent{
		AgentType: "peer-b",
		Workspace: "peer-b/",
		Sensors:   []string{"scope-guard-files", "dispatch-guard"},
	}
	cfg.Agents["peer-b"] = peerB
	var wr WriteResult
	if err := AgentWorkspace(tmpDir, cat.GetAgent("peer-b"), peerB, cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("peer-b workspace: %v", err)
	}

	var refreshWR WriteResult
	if err := RefreshPeerAwareness(tmpDir, "peer-b", cfg, cat, lock, &refreshWR, false); err != nil {
		t.Fatalf("RefreshPeerAwareness: %v", err)
	}

	for _, f := range refreshWR.Files {
		if strings.HasPrefix(f.RelPath, "peer-b/") {
			t.Errorf("excluded agent should not appear in refresh WriteResult: %q", f.RelPath)
		}
	}
}

// TestRefreshPeerAwareness_SkipsAgentsMissingSensor — a peer without
// dispatch-guard installed does not crash or emit a FileResult for that sensor.
func TestRefreshPeerAwareness_SkipsAgentsMissingSensor(t *testing.T) {
	cat, err := buildPeerTestCatalog()
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	tmpDir := t.TempDir()

	// peer-a has scope-guard-files but NOT dispatch-guard.
	lead := &config.InstalledAgent{
		AgentType: "lead",
		Workspace: "station/",
		Sensors:   []string{"scope-guard-files", "dispatch-guard"},
	}
	peerA := &config.InstalledAgent{
		AgentType: "peer-a",
		Workspace: "peer-a/",
		Sensors:   []string{"scope-guard-files"}, // no dispatch-guard
	}
	cfg := &config.ProjectConfig{
		ProjectName: "Test",
		Agents: map[string]*config.InstalledAgent{
			"lead":   lead,
			"peer-a": peerA,
		},
	}
	lock := config.NewLockFile()
	var wr WriteResult
	if err := AgentWorkspace(tmpDir, cat.GetAgent("lead"), lead, cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if err := AgentWorkspace(tmpDir, cat.GetAgent("peer-a"), peerA, cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("peer-a: %v", err)
	}

	var refreshWR WriteResult
	if err := RefreshPeerAwareness(tmpDir, "lead", cfg, cat, lock, &refreshWR, false); err != nil {
		t.Fatalf("RefreshPeerAwareness: %v", err)
	}

	// Peer-a should have its identity.md + scope-guard refreshed but NOT a
	// dispatch-guard file — that sensor was never installed.
	peerADispatchPath := filepath.Join("peer-a", "agent", "Sensors", "dispatch-guard.sh")
	for _, f := range refreshWR.Files {
		if f.RelPath == peerADispatchPath {
			t.Errorf("peer-a should not have dispatch-guard refreshed (not installed): %+v", f)
		}
	}
}

// TestRefreshPeerAwareness_TracksInWriteResult — refreshed files appear in
// result.Files with an Updated/Unchanged/Created action (not conflict).
func TestRefreshPeerAwareness_TracksInWriteResult(t *testing.T) {
	cat, cfg, lock, tmpDir := setupPeerFixture(t)

	// Add peer-b so lead/peer-a see fresh OtherAgents.
	peerB := &config.InstalledAgent{
		AgentType: "peer-b",
		Workspace: "peer-b/",
		Sensors:   []string{"scope-guard-files", "dispatch-guard"},
	}
	cfg.Agents["peer-b"] = peerB
	var wr WriteResult
	if err := AgentWorkspace(tmpDir, cat.GetAgent("peer-b"), peerB, cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("peer-b: %v", err)
	}

	var refreshWR WriteResult
	if err := RefreshPeerAwareness(tmpDir, "peer-b", cfg, cat, lock, &refreshWR, false); err != nil {
		t.Fatalf("RefreshPeerAwareness: %v", err)
	}

	// We expect refreshes for lead + peer-a. Each has 3 files: identity.md,
	// scope-guard-files.sh, dispatch-guard.sh = 6 total entries.
	if len(refreshWR.Files) != 6 {
		t.Errorf("expected 6 file results, got %d: %+v", len(refreshWR.Files), refreshWR.Files)
	}
	for _, f := range refreshWR.Files {
		if f.Action == ActionConflict {
			t.Errorf("unexpected conflict on unmodified refresh: %+v", f)
		}
	}
}
