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

// buildBonsaiRefTestCatalog creates a minimal catalog adequate for exercising
// WorkspaceClaudeMD's Bonsai Reference block rendering (Plan 31 Phase D).
func buildBonsaiRefTestCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	fsys := fstest.MapFS{
		"core/memory.md.tmpl":         &fstest.MapFile{Data: []byte("mem")},
		"core/self-awareness.md":      &fstest.MapFile{Data: []byte("sa")},
		"core/identity.md.tmpl":       &fstest.MapFile{Data: []byte("id {{ .AgentDisplayName }}")},
		"agents/tech-lead/agent.yaml": &fstest.MapFile{Data: []byte("name: tech-lead\ndisplay_name: Tech Lead\ndescription: lead\n")},
		"agents/backend/agent.yaml":   &fstest.MapFile{Data: []byte("name: backend\ndisplay_name: Backend\ndescription: backend\n")},
	}
	cat, err := catalog.New(fsys)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	return cat
}

// TestBonsaiReference_TechLeadWorkspace — for tech-lead, bonsai-model path
// resolves to "agent/Skills/bonsai-model.md" (workspace-local) and
// catalog.json resolves to "../.bonsai/catalog.json".
func TestBonsaiReference_TechLeadWorkspace(t *testing.T) {
	cat := buildBonsaiRefTestCatalog(t)
	tmpDir := t.TempDir()

	installed := &config.InstalledAgent{
		AgentType: "tech-lead",
		Workspace: "station/",
	}
	cfg := &config.ProjectConfig{
		ProjectName: "Test",
		DocsPath:    "station/",
		Agents:      map[string]*config.InstalledAgent{"tech-lead": installed},
	}
	lock := config.NewLockFile()
	var wr WriteResult

	if err := AgentWorkspace(tmpDir, cat.GetAgent("tech-lead"), installed, cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("AgentWorkspace: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "station", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "## Bonsai Reference") {
		t.Errorf("CLAUDE.md missing Bonsai Reference heading:\n%s", content)
	}
	// Tech-lead's own workspace-relative path.
	if !strings.Contains(content, "agent/Skills/bonsai-model.md") {
		t.Errorf("CLAUDE.md missing bonsai-model path:\n%s", content)
	}
	// .bonsai/catalog.json — from station/ to project root's .bonsai/ → ../.bonsai/catalog.json
	if !strings.Contains(content, "../.bonsai/catalog.json") {
		t.Errorf("CLAUDE.md missing expected ../.bonsai/catalog.json:\n%s", content)
	}
	if !strings.Contains(content, "../.bonsai.yaml") {
		t.Errorf("CLAUDE.md missing ../.bonsai.yaml:\n%s", content)
	}
}

// TestBonsaiReference_NonTechLeadPointsToTechLead — non-tech-lead agents'
// CLAUDE.md points to tech-lead's workspace for bonsai-model.md.
func TestBonsaiReference_NonTechLeadPointsToTechLead(t *testing.T) {
	cat := buildBonsaiRefTestCatalog(t)
	tmpDir := t.TempDir()

	techLead := &config.InstalledAgent{AgentType: "tech-lead", Workspace: "station/"}
	backend := &config.InstalledAgent{AgentType: "backend", Workspace: "backend/"}
	cfg := &config.ProjectConfig{
		ProjectName: "Test",
		DocsPath:    "station/",
		Agents: map[string]*config.InstalledAgent{
			"tech-lead": techLead,
			"backend":   backend,
		},
	}
	lock := config.NewLockFile()
	var wr WriteResult

	if err := AgentWorkspace(tmpDir, cat.GetAgent("backend"), backend, cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("backend workspace: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "backend", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read backend/CLAUDE.md: %v", err)
	}
	content := string(data)

	// Backend's CLAUDE.md should reference ../station/agent/Skills/bonsai-model.md
	if !strings.Contains(content, "../station/agent/Skills/bonsai-model.md") {
		t.Errorf("backend CLAUDE.md missing pointer to tech-lead's bonsai-model:\n%s", content)
	}
}

// TestBonsaiReference_OrderedBeforeQuickTriggers — the Bonsai Reference block
// must come after Core nav table and before Quick Triggers so agents load it
// in the natural reading order.
func TestBonsaiReference_OrderedBeforeQuickTriggers(t *testing.T) {
	cat := buildBonsaiRefTestCatalog(t)
	tmpDir := t.TempDir()

	installed := &config.InstalledAgent{AgentType: "tech-lead", Workspace: "station/"}
	cfg := &config.ProjectConfig{
		ProjectName: "Test",
		DocsPath:    "station/",
		Agents:      map[string]*config.InstalledAgent{"tech-lead": installed},
	}
	lock := config.NewLockFile()
	var wr WriteResult
	if err := AgentWorkspace(tmpDir, cat.GetAgent("tech-lead"), installed, cfg, cat, lock, &wr, false); err != nil {
		t.Fatalf("workspace: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "station", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	content := string(data)

	coreIdx := strings.Index(content, "### Core (load first")
	bonsaiIdx := strings.Index(content, "## Bonsai Reference")
	qtIdx := strings.Index(content, "### Quick Triggers")
	if coreIdx < 0 || bonsaiIdx < 0 || qtIdx < 0 {
		t.Fatalf("missing one of core/bonsai/quick-triggers: core=%d bonsai=%d qt=%d", coreIdx, bonsaiIdx, qtIdx)
	}
	if !(coreIdx < bonsaiIdx && bonsaiIdx < qtIdx) {
		t.Errorf("order wrong: core=%d bonsai=%d qt=%d (want core < bonsai < qt)", coreIdx, bonsaiIdx, qtIdx)
	}
}

// TestBonsaiReferenceLines_EmptyDocsPathDegrades locks the degrade-don't-
// crash contract after the empty-DocsPath fallback was removed from
// bonsaiReferenceLines. cmd/init_flow.go:233 normalises DocsPath to non-
// empty + trailing "/" before saving cfg, so an empty DocsPath here means
// an invariant was already broken upstream — the renderer must still
// produce output (degenerate but non-crashing), not panic.
func TestBonsaiReferenceLines_EmptyDocsPathDegrades(t *testing.T) {
	cfg := &config.ProjectConfig{
		ProjectName: "Test",
		DocsPath:    "", // violates the init_flow invariant on purpose
	}

	// Call the unexported helper directly — same contract as via
	// WorkspaceClaudeMD, without requiring a full template render setup.
	lines := bonsaiReferenceLines("/proj", "/proj/station", cfg)

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "## Bonsai Reference") {
		t.Errorf("expected Bonsai Reference heading even with empty DocsPath:\n%s", joined)
	}
	if !strings.Contains(joined, "bonsai-model.md") {
		t.Errorf("expected bonsai-model.md row even with empty DocsPath:\n%s", joined)
	}
}
