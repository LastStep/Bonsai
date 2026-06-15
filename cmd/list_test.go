package cmd

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bonsai "github.com/LastStep/Bonsai"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
)

// setupListTestCatalog wires the embedded catalog into the cmd package's
// package-level catalogFS variable so runList's loadCatalog() call can
// find a populated catalog. The production wiring happens in
// cmd/bonsai/main.go; the test binary skips that entry point, so each
// test that needs the catalog opts in explicitly via this helper. Prior
// value is restored via t.Cleanup so other tests in the package (which
// previously ran with catalogFS nil) are unaffected.
func setupListTestCatalog(t *testing.T) {
	t.Helper()
	sub, err := fs.Sub(bonsai.CatalogFS, "catalog")
	if err != nil {
		t.Fatalf("fs.Sub(CatalogFS): %v", err)
	}
	prev := catalogFS
	catalogFS = sub
	t.Cleanup(func() { catalogFS = prev })
}

// TestRunList_HappyPath sets up a minimal project with one agent + one
// workspace file and asserts the cinematic output contains the expected
// anchors: LIST action label, agent display-name panel, counts footer,
// and the live workspace file. Uses captureStdout (defined in
// add_test.go) for output redirection.
func TestRunList_HappyPath(t *testing.T) {
	setupListTestCatalog(t)

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

// TestRunList_NoAgents covers the zero-agents branch — a project
// config with an empty Agents map must still render the cinematic
// LIST header + a 0-agent count + return nil (no panic on the
// empty-map loop). Mirrors TestRunList_HappyPath's scaffold but
// swaps in an empty Agents map.
//
// The non-TTY fallback case isn't a separate test: captureStdout
// redirects os.Stdout to a pipe, so runList's `isTerminal(os.Stdout)`
// check returns false and the happy-path + no-agents tests already
// exercise the non-TTY render. Mentioned in the PR body.
func TestRunList_NoAgents(t *testing.T) {
	setupListTestCatalog(t)

	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "demo-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir projectDir: %v", err)
	}

	cfg := &config.ProjectConfig{
		ProjectName: "demo-project",
		Agents:      map[string]*config.InstalledAgent{},
	}
	if err := cfg.Save(filepath.Join(projectDir, configFile)); err != nil {
		t.Fatalf("save config: %v", err)
	}

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origCwd) }()

	stdout := captureStdout(t, func() {
		if err := runList(nil, nil); err != nil {
			t.Fatalf("runList: %v", err)
		}
	})

	if !strings.Contains(stdout, "LIST") {
		t.Fatalf("expected 'LIST' in header, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "0 agent") {
		t.Fatalf("expected '0 agent' count, got:\n%s", stdout)
	}
}

// TestRunList_JSONSchema is the Plan 41 Phase 4 B2 gate: a two-agent project
// rendered through `bonsai list --json` must Unmarshal cleanly into the pinned
// ListSnapshot struct with EVERY field populated and field names/types exactly
// matching the contract. Drives runList with the --json flag set on listCmd so
// the real flag-read short-circuit is exercised end to end.
func TestRunList_JSONSchema(t *testing.T) {
	setupListTestCatalog(t)

	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "demo-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir projectDir: %v", err)
	}

	cfg := &config.ProjectConfig{
		ProjectName: "demo-project",
		DocsPath:    "docs",
		Agents: map[string]*config.InstalledAgent{
			"tech-lead": {
				AgentType: "tech-lead",
				Workspace: "station",
				Skills:    []string{"planning-template", "review-checklist"},
				Workflows: []string{"code-review"},
				Protocols: []string{"memory", "security"},
				Sensors:   []string{"scope-guard-files"},
				Routines:  []string{"backlog-hygiene"},
			},
			"backend": {
				AgentType: "backend",
				Workspace: "backend-ws",
				Skills:    []string{"coding-standards"},
				Workflows: []string{"pr-review"},
				Protocols: []string{"scope-boundaries"},
				Sensors:   []string{"context-guard"},
				Routines:  []string{"dependency-audit"},
			},
		},
	}
	if err := cfg.Save(filepath.Join(projectDir, configFile)); err != nil {
		t.Fatalf("save config: %v", err)
	}

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origCwd) }()

	// Set the --json flag on the real listCmd so runList's flag-read
	// short-circuit fires; reset it afterwards so other tests see the
	// default value.
	if err := listCmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("set --json: %v", err)
	}
	defer func() { _ = listCmd.Flags().Set("json", "false") }()

	out := captureStdout(t, func() {
		if err := runList(listCmd, nil); err != nil {
			t.Fatalf("runList --json: %v", err)
		}
	})

	var snap generate.ListSnapshot
	if err := json.Unmarshal([]byte(out), &snap); err != nil {
		t.Fatalf("unmarshal ListSnapshot: %v\noutput:\n%s", err, out)
	}

	if snap.Version == "" {
		t.Errorf("snap.Version empty — want the build version string")
	}
	if snap.DocsPath != "docs" {
		t.Errorf("snap.DocsPath = %q, want %q", snap.DocsPath, "docs")
	}
	if len(snap.Agents) != 2 {
		t.Fatalf("len(snap.Agents) = %d, want 2", len(snap.Agents))
	}

	// Agents are emitted alphabetically by type: backend then tech-lead.
	if snap.Agents[0].Type != "backend" || snap.Agents[1].Type != "tech-lead" {
		t.Fatalf("agent order = [%q %q], want [backend tech-lead]",
			snap.Agents[0].Type, snap.Agents[1].Type)
	}

	// Every field of every ListAgent must be populated — proves no field
	// dropped out of the contract.
	for _, a := range snap.Agents {
		if a.Type == "" {
			t.Errorf("agent.Type empty")
		}
		if a.Workspace == "" {
			t.Errorf("agent %q: Workspace empty", a.Type)
		}
		if len(a.Skills) == 0 {
			t.Errorf("agent %q: Skills empty", a.Type)
		}
		if len(a.Workflows) == 0 {
			t.Errorf("agent %q: Workflows empty", a.Type)
		}
		if len(a.Protocols) == 0 {
			t.Errorf("agent %q: Protocols empty", a.Type)
		}
		if len(a.Sensors) == 0 {
			t.Errorf("agent %q: Sensors empty", a.Type)
		}
		if len(a.Routines) == 0 {
			t.Errorf("agent %q: Routines empty", a.Type)
		}
	}

	// Spot-check the tech-lead record fully — every slice value preserved.
	tl := snap.Agents[1]
	if tl.Workspace != "station" {
		t.Errorf("tech-lead Workspace = %q, want %q", tl.Workspace, "station")
	}
	if got, want := strings.Join(tl.Skills, ","), "planning-template,review-checklist"; got != want {
		t.Errorf("tech-lead Skills = %q, want %q", got, want)
	}
	if got, want := strings.Join(tl.Protocols, ","), "memory,security"; got != want {
		t.Errorf("tech-lead Protocols = %q, want %q", got, want)
	}

	// Output must be indent-2 JSON (parity with catalog --json / validate
	// --json) — assert the canonical 2-space indent appears.
	if !strings.Contains(out, "\n  \"version\":") {
		t.Errorf("expected indent-2 JSON (2-space), got:\n%s", out)
	}
}

// TestListCmd_FlagsRegistered confirms the --json flag is wired onto listCmd
// so `--help` picks it up and the JSON short-circuit is reachable. Mirrors
// TestInitCmd_FlagsRegistered.
func TestListCmd_FlagsRegistered(t *testing.T) {
	for _, name := range []string{"json"} {
		if f := listCmd.Flags().Lookup(name); f == nil {
			t.Errorf("flag --%s not registered on listCmd", name)
		}
	}
}
