package cmd

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bonsai "github.com/LastStep/Bonsai"
	"github.com/LastStep/Bonsai/internal/config"
)

// TestMain wires the embedded catalog into the cmd package's package-level
// catalogFS variable so runList's loadCatalog() call can find a populated
// catalog during tests. The production wiring happens in cmd/bonsai/main.go;
// the test binary skips that entry point and needs an equivalent here.
func TestMain(m *testing.M) {
	sub, err := fs.Sub(bonsai.CatalogFS, "catalog")
	if err == nil {
		catalogFS = sub
	}
	os.Exit(m.Run())
}

// TestRunList_HappyPath sets up a minimal project with one agent + one
// workspace file and asserts the cinematic output contains the expected
// anchors: LIST action label, agent display-name panel, counts footer,
// and the live workspace file. Uses captureStdout (defined in
// add_test.go) for output redirection.
func TestRunList_HappyPath(t *testing.T) {
	// Isolate the test from the developer's home dir / Bonsai config —
	// runList calls mustCwd + requireConfig which read the real filesystem.
	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "demo-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir projectDir: %v", err)
	}
	ws := filepath.Join(projectDir, "station")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ws, "CLAUDE.md"), []byte("# hello"), 0o644); err != nil {
		t.Fatalf("write workspace file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ws, "INDEX.md"), []byte("# index"), 0o644); err != nil {
		t.Fatalf("write index file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ws, "memory.md"), []byte("# memory"), 0o644); err != nil {
		t.Fatalf("write memory file: %v", err)
	}

	cfg := &config.ProjectConfig{
		ProjectName: "demo-project",
		Agents: map[string]*config.InstalledAgent{
			"tech-lead": {
				AgentType: "tech-lead",
				Workspace: "station",
			},
		},
	}
	if err := cfg.Save(filepath.Join(projectDir, configFile)); err != nil {
		t.Fatalf("save config: %v", err)
	}

	// chdir into projectDir so mustCwd resolves to the right root.
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origCwd) }()

	// Load the bundled catalog so GetAgent("tech-lead") resolves the
	// display name. The test binary has catalogFS/guideContents set to
	// nil by default; init-time bundled load happens in main.go — we
	// bypass it here by loading via a nil-safe catalog inside listflow
	// (display name falls back to DisplayNameFrom("tech-lead")).

	stdout := captureStdout(t, func() {
		if err := runList(nil, nil); err != nil {
			t.Fatalf("runList: %v", err)
		}
	})

	// Header should include the LIST action label.
	if !strings.Contains(stdout, "LIST") {
		t.Fatalf("expected 'LIST' in header, got:\n%s", stdout)
	}
	// Agent display-name (catalog may be nil in test — DisplayNameFrom
	// still produces "Tech Lead").
	if !strings.Contains(stdout, "Tech Lead") {
		t.Fatalf("expected 'Tech Lead' panel title, got:\n%s", stdout)
	}
	// Counts footer — single agent, zero abilities installed here.
	if !strings.Contains(stdout, "1 agent") {
		t.Fatalf("expected '1 agent' count, got:\n%s", stdout)
	}
	// At least one of the workspace files should appear in the tree.
	if !strings.Contains(stdout, "CLAUDE.md") {
		t.Fatalf("expected CLAUDE.md in workspace tree, got:\n%s", stdout)
	}
}
