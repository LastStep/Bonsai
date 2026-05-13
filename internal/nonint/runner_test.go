package nonint

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
)

// parseJSONL splits a JSONL stream into discrete records, deserialising
// each into a map. Used by every test that walks the runner's stdout.
func parseJSONL(t *testing.T, raw string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("parse JSONL line %q: %v", line, err)
		}
		out = append(out, rec)
	}
	return out
}

// findSummary walks parsed events and returns the summary record (or nil
// if absent).
func findSummary(records []map[string]any) map[string]any {
	for _, r := range records {
		if r["event"] == "summary" {
			return r
		}
	}
	return nil
}

// minimalInitCfg loads the tech-lead-only fixture used by most init-path
// tests.
func minimalInitCfg(t *testing.T, cwd string) *config.ProjectConfig {
	t.Helper()
	cat := loadTestCatalog(t)
	body := "agents:\n  tech-lead: {}\n"
	cfgPath := writeYAML(t, cwd, "cfg.yaml", body)
	cfg, err := LoadConfig(cfgPath, cwd, cat)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	return cfg
}

// TestRunInit_Smoke runs the runner end-to-end against a fresh tmp dir with
// the tech-lead-only fixture. Asserts: .bonsai.yaml is written; station/
// tree is materialised; every emitted line is valid JSON; a summary line
// appears.
func TestRunInit_Smoke(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	cfg := minimalInitCfg(t, tmp)
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	var buf bytes.Buffer
	code, err := RunInit(tmp, configPath, cfg, cat, "test", &buf)
	if err != nil {
		t.Fatalf("RunInit: %v", err)
	}
	if code != ExitOK {
		t.Fatalf("exit code: want %d, got %d", ExitOK, code)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf(".bonsai.yaml not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "station")); err != nil {
		t.Fatalf("station/ tree not created: %v", err)
	}
	records := parseJSONL(t, buf.String())
	if len(records) == 0 {
		t.Fatalf("no JSONL records emitted")
	}
	summary := findSummary(records)
	if summary == nil {
		t.Fatalf("no summary event in stream:\n%s", buf.String())
	}
	if summary["created"].(float64) == 0 {
		t.Errorf("expected created > 0; got summary %v", summary)
	}
}

// TestRunInit_ConfigExists_ExitCode4 pre-creates .bonsai.yaml and asserts the
// runner refuses to overwrite — exit 4 + stderr-style error.
func TestRunInit_ConfigExists_ExitCode4(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	cfg := minimalInitCfg(t, tmp)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	if err := os.WriteFile(configPath, []byte("project_name: prior\nagents: {}\n"), 0o644); err != nil {
		t.Fatalf("seed .bonsai.yaml: %v", err)
	}

	code, err := RunInit(tmp, configPath, cfg, cat, "test", io.Discard)
	if code != ExitWrongCWDForInit {
		t.Errorf("exit code: want %d, got %d", ExitWrongCWDForInit, code)
	}
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error must mention already exists; got %v", err)
	}
}

// TestRunInit_MissingTechLead is a defence-in-depth test: a caller that
// bypassed the cmd-layer's tech-lead check shouldn't crash the runner —
// it must return ExitInvalidConfig.
func TestRunInit_MissingTechLead(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	// Build a cfg with a non-tech-lead agent only. LoadConfig will accept
	// this (it's the init-caller's responsibility to enforce tech-lead).
	body := "agents:\n  backend: {}\n"
	cfgPath := writeYAML(t, tmp, "cfg.yaml", body)
	cfg, err := LoadConfig(cfgPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	code, err := RunInit(tmp, filepath.Join(tmp, ".bonsai.yaml"), cfg, cat, "test", io.Discard)
	if code != ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "tech-lead") {
		t.Errorf("error must mention tech-lead; got %v", err)
	}
}

// TestRunAdd_NewAgent_Smoke initialises a project then runs RunAdd with a
// backend-agent overlay. The backend workspace must be materialised and
// the overlay agent must appear in .bonsai.yaml.
func TestRunAdd_NewAgent_Smoke(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()

	// init
	initCfg := minimalInitCfg(t, tmp)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	if code, err := RunInit(tmp, configPath, initCfg, cat, "test", io.Discard); err != nil || code != ExitOK {
		t.Fatalf("RunInit failed (code=%d err=%v)", code, err)
	}

	// overlay: a backend agent
	overlayBody := "agents:\n  backend: {}\n"
	overlayPath := writeYAML(t, tmp, "overlay.yaml", overlayBody)
	overlay, err := LoadConfig(overlayPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig overlay: %v", err)
	}

	var buf bytes.Buffer
	code, err := RunAdd(tmp, configPath, overlay, cat, "test", &buf)
	if err != nil {
		t.Fatalf("RunAdd: %v", err)
	}
	if code != ExitOK {
		t.Errorf("exit code: want %d, got %d", ExitOK, code)
	}
	// backend workspace materialised
	if _, err := os.Stat(filepath.Join(tmp, "backend")); err != nil {
		t.Errorf("backend/ workspace not created: %v", err)
	}
	// .bonsai.yaml now has both agents
	post, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("reload .bonsai.yaml: %v", err)
	}
	if _, ok := post.Agents["backend"]; !ok {
		t.Errorf(".bonsai.yaml does not contain backend agent")
	}
	// summary line in output
	records := parseJSONL(t, buf.String())
	if findSummary(records) == nil {
		t.Errorf("missing summary event")
	}
}

// TestRunAdd_AddItems_Smoke: RunInit with tech-lead, then RunAdd targeting
// tech-lead with an extra skill that wasn't in the defaults. The extra
// skill must be appended without duplicating existing entries.
func TestRunAdd_AddItems_Smoke(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()

	initCfg := minimalInitCfg(t, tmp)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	if code, err := RunInit(tmp, configPath, initCfg, cat, "test", io.Discard); err != nil || code != ExitOK {
		t.Fatalf("RunInit failed (code=%d err=%v)", code, err)
	}

	// Find a skill the tech-lead supports but doesn't have by default.
	def := cat.GetAgent("tech-lead")
	defaultSet := make(map[string]bool, len(def.DefaultSkills))
	for _, s := range def.DefaultSkills {
		defaultSet[s] = true
	}
	var extra string
	for _, s := range cat.SkillsFor("tech-lead") {
		if !defaultSet[s.Name] {
			extra = s.Name
			break
		}
	}
	if extra == "" {
		t.Skip("no extra tech-lead skill available — every skill in defaults")
	}

	// Overlay with the extra skill — also include the defaults so explicit
	// non-nil ability lists prevent the defaulting walk from re-populating.
	overlayBody := "agents:\n  tech-lead:\n    skills:\n      - " + extra + "\n"
	overlayPath := writeYAML(t, tmp, "overlay.yaml", overlayBody)
	overlay, err := LoadConfig(overlayPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig overlay: %v", err)
	}

	if code, err := RunAdd(tmp, configPath, overlay, cat, "test", io.Discard); err != nil || code != ExitOK {
		t.Fatalf("RunAdd failed (code=%d err=%v)", code, err)
	}

	post, _ := config.Load(configPath)
	tl := post.Agents["tech-lead"]
	count := 0
	for _, s := range tl.Skills {
		if s == extra {
			count++
		}
	}
	if count != 1 {
		t.Errorf("extra skill %q appears %d times in post-add skills %v; want 1", extra, count, tl.Skills)
	}
}

// TestRunAdd_AllInstalled_ShortCircuit: re-running the same overlay against
// an already-installed agent with no additional abilities must emit a
// single all-zero summary event and exit 0. Decision §6.
func TestRunAdd_AllInstalled_ShortCircuit(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()

	initCfg := minimalInitCfg(t, tmp)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	if code, err := RunInit(tmp, configPath, initCfg, cat, "test", io.Discard); err != nil || code != ExitOK {
		t.Fatalf("RunInit failed (code=%d err=%v)", code, err)
	}

	// Re-run with the SAME minimal overlay — every default ability is
	// already installed.
	overlayPath := writeYAML(t, tmp, "overlay.yaml", "agents:\n  tech-lead: {}\n")
	overlay, err := LoadConfig(overlayPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig overlay: %v", err)
	}

	var buf bytes.Buffer
	code, err := RunAdd(tmp, configPath, overlay, cat, "test", &buf)
	if err != nil {
		t.Fatalf("RunAdd: %v", err)
	}
	if code != ExitOK {
		t.Errorf("exit code: want %d, got %d", ExitOK, code)
	}
	records := parseJSONL(t, buf.String())
	if len(records) != 1 {
		t.Fatalf("expected exactly 1 record (summary only); got %d:\n%s", len(records), buf.String())
	}
	summary := records[0]
	if summary["event"] != "summary" {
		t.Errorf("expected single summary event; got %v", summary)
	}
	for _, k := range []string{"created", "updated", "unchanged", "skipped", "conflicts"} {
		if v, ok := summary[k]; !ok || v.(float64) != 0 {
			t.Errorf("%s: want 0, got %v (present=%v)", k, v, ok)
		}
	}
}

// TestRunAdd_MultiAgentOverlay_Rejected: overlay with two agents must be
// rejected with ExitInvalidConfig. Decision §2.
func TestRunAdd_MultiAgentOverlay_Rejected(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()

	initCfg := minimalInitCfg(t, tmp)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	if code, err := RunInit(tmp, configPath, initCfg, cat, "test", io.Discard); err != nil || code != ExitOK {
		t.Fatalf("RunInit failed (code=%d err=%v)", code, err)
	}

	overlayBody := "agents:\n  tech-lead: {}\n  backend: {}\n"
	overlayPath := writeYAML(t, tmp, "overlay.yaml", overlayBody)
	overlay, err := LoadConfig(overlayPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig overlay: %v", err)
	}

	code, err := RunAdd(tmp, configPath, overlay, cat, "test", io.Discard)
	if code != ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "exactly one agent") {
		t.Errorf("error must mention exactly one agent; got %v", err)
	}
}

// TestRunAdd_OverlayMismatchedProjectName_Rejected: overlay project_name
// must match existing or be empty. Decision §3.
func TestRunAdd_OverlayMismatchedProjectName_Rejected(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()

	initCfg := minimalInitCfg(t, tmp)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	if code, err := RunInit(tmp, configPath, initCfg, cat, "test", io.Discard); err != nil || code != ExitOK {
		t.Fatalf("RunInit failed (code=%d err=%v)", code, err)
	}

	// Overlay with a different project_name. LoadConfig will accept it
	// (the basename-default doesn't fire when the user supplies a value);
	// RunAdd must reject.
	overlayBody := "project_name: NOT-THE-SAME\nagents:\n  backend: {}\n"
	overlayPath := writeYAML(t, tmp, "overlay.yaml", overlayBody)
	overlay, err := LoadConfig(overlayPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig overlay: %v", err)
	}

	code, err := RunAdd(tmp, configPath, overlay, cat, "test", io.Discard)
	if code != ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "project_name") {
		t.Errorf("error must mention project_name; got %v", err)
	}
}

// TestRunAdd_TechLeadRequired_Errors: a non-tech-lead overlay against a
// project with no tech-lead must exit 2 (Decision §4). The existing config
// must have a project_name that matches the cwd-basename default the overlay
// inherits, so the §3 project_name match passes and the §4 tech-lead check
// is the rule that trips.
func TestRunAdd_TechLeadRequired_Errors(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	base := filepath.Base(tmp)

	// Hand-roll a .bonsai.yaml with NO tech-lead. config.Load enforces
	// shell-metachar + workspace rules but not the tech-lead-presence rule,
	// so this is a legal existing config that RunAdd must guard against.
	// Include the required scaffolding so the overlay's applyDefaults result
	// matches; we want §4 (tech-lead-required) to fire, not §3 (scaffolding
	// mismatch).
	body := "project_name: " + base + "\ndocs_path: station/\nscaffolding:\n  - index\n  - playbook\n  - logs\nagents:\n  backend:\n    workspace: backend/\n    skills: []\n    workflows: []\n    protocols: []\n    sensors: []\n    routines: []\n"
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	if err := os.WriteFile(configPath, []byte(body), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	overlayPath := writeYAML(t, tmp, "overlay.yaml", "agents:\n  security: {}\n")
	overlay, err := LoadConfig(overlayPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig overlay: %v", err)
	}

	code, err := RunAdd(tmp, configPath, overlay, cat, "test", io.Discard)
	if code != ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "tech-lead") {
		t.Errorf("error must mention tech-lead; got %v", err)
	}
}

// TestRunAdd_UnknownAgent_Errors: overlay names a non-existent agent type.
// LoadConfig accepts unknown agent types (it cannot tell what the catalog
// knows); RunAdd checks against the loaded catalog.
func TestRunAdd_UnknownAgent_Errors(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()

	initCfg := minimalInitCfg(t, tmp)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	if code, err := RunInit(tmp, configPath, initCfg, cat, "test", io.Discard); err != nil || code != ExitOK {
		t.Fatalf("RunInit failed (code=%d err=%v)", code, err)
	}

	// Overlay with a phantom agent. LoadConfig's applyDefaults sets the
	// workspace from the type name; cfg.Validate accepts it as long as the
	// path is project-relative.
	overlayPath := writeYAML(t, tmp, "overlay.yaml", "agents:\n  does-not-exist: {}\n")
	overlay, err := LoadConfig(overlayPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig overlay: %v", err)
	}
	code, err := RunAdd(tmp, configPath, overlay, cat, "test", io.Discard)
	if code != ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "unknown agent type") {
		t.Errorf("error must mention unknown agent type; got %v", err)
	}
}

// TestRunAdd_MissingProjectConfig_ExitCode4: RunAdd in a directory without
// .bonsai.yaml must exit 4.
func TestRunAdd_MissingProjectConfig_ExitCode4(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	overlayPath := writeYAML(t, tmp, "overlay.yaml", "agents:\n  tech-lead: {}\n")
	overlay, err := LoadConfig(overlayPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig overlay: %v", err)
	}
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	code, err := RunAdd(tmp, configPath, overlay, cat, "test", io.Discard)
	if code != ExitWrongCWDForInit {
		t.Errorf("exit code: want %d, got %d", ExitWrongCWDForInit, code)
	}
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("error must mention not found; got %v", err)
	}
}

// TestRunInit_ConflictEmittedNotForced is the key safety test: a pre-existing
// user-edited file at a path the generator targets must produce
// action=conflict in the JSONL stream AND the file on disk must remain
// byte-identical to the user's edit. Plan 39 Security §3.
//
// station/agent/Core/identity.md is a stable generator output for the
// tech-lead init path (every init writes it via the agent-workspace pipeline)
// AND it flows through the standard writeFile lock-aware path — unlike
// station/CLAUDE.md, which has its own bonsai-markers migration branch.
func TestRunInit_ConflictEmittedNotForced(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()

	userBody := []byte("# USER EDIT — must not be overwritten\n")
	relPath := "station/agent/Core/identity.md"
	target := filepath.Join(tmp, relPath)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(target, userBody, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// IsModified returns exists+modified=true when a file exists on disk
	// with no lock entry (conservative: treat as user-owned). So no
	// pre-seed of the lock file is needed — writeFile will see an
	// untracked existing file and emit ActionConflict.

	cfg := minimalInitCfg(t, tmp)
	var buf bytes.Buffer
	code, err := RunInit(tmp, filepath.Join(tmp, ".bonsai.yaml"), cfg, cat, "test", &buf)
	if err != nil {
		t.Fatalf("RunInit: %v", err)
	}
	if code != ExitOK {
		t.Fatalf("exit code: %d", code)
	}

	records := parseJSONL(t, buf.String())
	var foundConflict bool
	for _, r := range records {
		if r["event"] == "file" && r["path"] == relPath && r["action"] == "conflict" {
			foundConflict = true
			break
		}
	}
	if !foundConflict {
		t.Errorf("expected conflict event for %s; records=%+v", relPath, records)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read post: %v", err)
	}
	if !bytes.Equal(got, userBody) {
		t.Errorf("user file overwritten; want %q, got %q", userBody, got)
	}
}

// TestAssertOverlayMatchesExisting_EmptyFieldsPass: the §3 contract accepts
// empty / nil non-`agents` fields without comparison. Pure-function test.
func TestAssertOverlayMatchesExisting_EmptyFieldsPass(t *testing.T) {
	existing := &config.ProjectConfig{
		ProjectName: "abc",
		DocsPath:    "station/",
		Scaffolding: []string{"index", "playbook", "logs"},
	}
	overlay := &config.ProjectConfig{}
	if err := assertOverlayMatchesExisting(overlay, existing); err != nil {
		t.Errorf("empty overlay must pass; got %v", err)
	}
}

// TestAssertOverlayMatchesExisting_ScaffoldingOrderTolerated: a non-empty
// scaffolding overlay must match existing up to ordering (sorted-equal).
func TestAssertOverlayMatchesExisting_ScaffoldingOrderTolerated(t *testing.T) {
	existing := &config.ProjectConfig{
		ProjectName: "abc",
		DocsPath:    "station/",
		Scaffolding: []string{"index", "playbook", "logs"},
	}
	overlay := &config.ProjectConfig{
		Scaffolding: []string{"logs", "index", "playbook"},
	}
	if err := assertOverlayMatchesExisting(overlay, existing); err != nil {
		t.Errorf("sorted-equal scaffolding must pass; got %v", err)
	}
	// And a non-match must fail.
	overlay.Scaffolding = []string{"index"}
	if err := assertOverlayMatchesExisting(overlay, existing); err == nil {
		t.Errorf("scaffolding mismatch must fail")
	}
}
