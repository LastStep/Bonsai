package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/nonint"
)

// initFreshProject is a test shorthand: spin up a tmp dir, run the
// non-interactive init with a tech-lead-only fixture, and return the
// project root. Used by every Phase C test that exercises an
// already-initialised project.
func initFreshProject(t *testing.T) string {
	t.Helper()
	setupListTestCatalog(t)
	tmp := t.TempDir()
	cfgPath := writeYAMLFixture(t, tmp, "init-cfg.yaml", "agents:\n  tech-lead: {}\n")
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	var stdout, stderr bytes.Buffer
	code, err := runInitNonInteractive(tmp, configPath, true, cfgPath, &stdout, &stderr)
	if err != nil || code != nonint.ExitOK {
		t.Fatalf("seed RunInit failed: code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	return tmp
}

// TestRunAddNonInteractive_NewAgent: init then add backend overlay. Backend
// workspace must materialise; JSONL summary must appear; cfg post-load has
// both agents.
func TestRunAddNonInteractive_NewAgent(t *testing.T) {
	tmp := initFreshProject(t)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	overlayPath := writeYAMLFixture(t, tmp, "overlay.yaml", "agents:\n  backend: {}\n")

	var stdout, stderr bytes.Buffer
	code, err := runAddNonInteractive(tmp, configPath, true, overlayPath, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runAddNonInteractive: %v (stderr=%s)", err, stderr.String())
	}
	if code != nonint.ExitOK {
		t.Fatalf("exit code: want %d, got %d", nonint.ExitOK, code)
	}
	if _, err := os.Stat(filepath.Join(tmp, "backend")); err != nil {
		t.Errorf("backend/ workspace not created: %v", err)
	}
	post, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("reload .bonsai.yaml: %v", err)
	}
	if _, ok := post.Agents["backend"]; !ok {
		t.Errorf("backend agent missing from post-add .bonsai.yaml")
	}

	// JSONL summary present
	var sawSummary bool
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Errorf("bad JSON: %q (%v)", line, err)
			continue
		}
		if rec["event"] == "summary" {
			sawSummary = true
		}
	}
	if !sawSummary {
		t.Errorf("missing summary event in:\n%s", stdout.String())
	}
}

// TestRunAddNonInteractive_AddItems: init then add tech-lead overlay with an
// extra skill. Skill must end up in the post-load tech-lead.Skills.
func TestRunAddNonInteractive_AddItems(t *testing.T) {
	tmp := initFreshProject(t)
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	// Find an extra tech-lead skill not in the defaults.
	setupListTestCatalog(t)
	cat := loadCatalog()
	def := cat.GetAgent("tech-lead")
	defaultSet := make(map[string]bool)
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
		t.Skip("no extra tech-lead skill in catalog")
	}

	overlayPath := writeYAMLFixture(t, tmp, "overlay.yaml",
		"agents:\n  tech-lead:\n    skills:\n      - "+extra+"\n")

	var stdout, stderr bytes.Buffer
	code, err := runAddNonInteractive(tmp, configPath, true, overlayPath, &stdout, &stderr)
	if err != nil || code != nonint.ExitOK {
		t.Fatalf("runAddNonInteractive: code=%d err=%v stderr=%s", code, err, stderr.String())
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
		t.Errorf("extra skill %q present %d times in post-add skills %v; want 1", extra, count, tl.Skills)
	}
}

// TestRunAddNonInteractive_AllInstalledShortCircuit: re-running the same
// minimal overlay against an already-installed agent emits a single
// all-zero summary line and nothing else.
func TestRunAddNonInteractive_AllInstalledShortCircuit(t *testing.T) {
	tmp := initFreshProject(t)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	overlayPath := writeYAMLFixture(t, tmp, "overlay.yaml", "agents:\n  tech-lead: {}\n")

	var stdout, stderr bytes.Buffer
	code, err := runAddNonInteractive(tmp, configPath, true, overlayPath, &stdout, &stderr)
	if err != nil || code != nonint.ExitOK {
		t.Fatalf("runAddNonInteractive: code=%d err=%v stderr=%s", code, err, stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	var records []map[string]any
	for _, line := range lines {
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Errorf("bad JSON: %q (%v)", line, err)
			continue
		}
		records = append(records, rec)
	}
	if len(records) != 1 {
		t.Fatalf("expected exactly 1 record (summary only); got %d:\n%s", len(records), stdout.String())
	}
	if records[0]["event"] != "summary" {
		t.Errorf("only record must be summary; got %v", records[0])
	}
	for _, k := range []string{"created", "updated", "unchanged", "skipped", "conflicts"} {
		if v, ok := records[0][k]; !ok || v.(float64) != 0 {
			t.Errorf("%s: want 0, got %v (present=%v)", k, v, ok)
		}
	}
}

// TestRunAddNonInteractive_MultiAgentOverlay: overlay with two agents → exit 2.
func TestRunAddNonInteractive_MultiAgentOverlay(t *testing.T) {
	tmp := initFreshProject(t)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	overlayPath := writeYAMLFixture(t, tmp, "overlay.yaml",
		"agents:\n  tech-lead: {}\n  backend: {}\n")

	var stdout, stderr bytes.Buffer
	code, err := runAddNonInteractive(tmp, configPath, true, overlayPath, &stdout, &stderr)
	if code != nonint.ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", nonint.ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "exactly one agent") {
		t.Errorf("error must mention exactly one agent; got %v", err)
	}
}

// TestRunAddNonInteractive_OverlayMismatch: overlay project_name differs
// from existing → exit 2.
func TestRunAddNonInteractive_OverlayMismatch(t *testing.T) {
	tmp := initFreshProject(t)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	overlayPath := writeYAMLFixture(t, tmp, "overlay.yaml",
		"project_name: WRONG\nagents:\n  backend: {}\n")

	var stdout, stderr bytes.Buffer
	code, err := runAddNonInteractive(tmp, configPath, true, overlayPath, &stdout, &stderr)
	if code != nonint.ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", nonint.ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "project_name") {
		t.Errorf("error must mention project_name; got %v", err)
	}
}

// TestRunAddNonInteractive_TechLeadRequired: hand-roll a .bonsai.yaml with
// NO tech-lead, then add a non-tech-lead overlay. Must exit 2 with the
// "requires a tech-lead" message.
func TestRunAddNonInteractive_TechLeadRequired(t *testing.T) {
	setupListTestCatalog(t)
	tmp := t.TempDir()
	base := filepath.Base(tmp)
	body := "project_name: " + base + "\ndocs_path: station/\nscaffolding:\n  - index\n  - playbook\n  - logs\nagents:\n  backend:\n    workspace: backend/\n    skills: []\n    workflows: []\n    protocols: []\n    sensors: []\n    routines: []\n"
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	if err := os.WriteFile(configPath, []byte(body), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	overlayPath := writeYAMLFixture(t, tmp, "overlay.yaml", "agents:\n  security: {}\n")

	var stdout, stderr bytes.Buffer
	code, err := runAddNonInteractive(tmp, configPath, true, overlayPath, &stdout, &stderr)
	if code != nonint.ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", nonint.ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "tech-lead") {
		t.Errorf("error must mention tech-lead; got %v", err)
	}
}

// TestRunAddNonInteractive_UnknownAgent: overlay names a phantom agent →
// exit 2 with "unknown agent type".
func TestRunAddNonInteractive_UnknownAgent(t *testing.T) {
	tmp := initFreshProject(t)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	overlayPath := writeYAMLFixture(t, tmp, "overlay.yaml", "agents:\n  does-not-exist: {}\n")

	var stdout, stderr bytes.Buffer
	code, err := runAddNonInteractive(tmp, configPath, true, overlayPath, &stdout, &stderr)
	if code != nonint.ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", nonint.ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "unknown agent type") {
		t.Errorf("error must mention unknown agent type; got %v", err)
	}
}

// TestRunAddNonInteractive_MissingProjectConfig: tmp dir without .bonsai.yaml
// → exit 4.
func TestRunAddNonInteractive_MissingProjectConfig(t *testing.T) {
	setupListTestCatalog(t)
	tmp := t.TempDir()
	overlayPath := writeYAMLFixture(t, tmp, "overlay.yaml", "agents:\n  tech-lead: {}\n")
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	var stdout, stderr bytes.Buffer
	code, err := runAddNonInteractive(tmp, configPath, true, overlayPath, &stdout, &stderr)
	if code != nonint.ExitWrongCWDForInit {
		t.Errorf("exit code: want %d, got %d", nonint.ExitWrongCWDForInit, code)
	}
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("error must mention not found; got %v", err)
	}
}

// TestRunAddNonInteractive_FlagAloneIsUsageError: --non-interactive alone
// or --from-config alone yields the "must be set together" error via
// cobra-handled exit-code-0 path.
func TestRunAddNonInteractive_FlagAloneIsUsageError(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	// --non-interactive alone
	code, err := runAddNonInteractive(tmp, configPath, true, "", io.Discard, io.Discard)
	if code != 0 {
		t.Errorf("--non-interactive alone: want code 0, got %d", code)
	}
	if err == nil || !strings.Contains(err.Error(), "must be set together") {
		t.Errorf("--non-interactive alone: want 'must be set together'; got %v", err)
	}

	// --from-config alone
	cfgPath := writeYAMLFixture(t, tmp, "cfg.yaml", "agents:\n  tech-lead: {}\n")
	code, err = runAddNonInteractive(tmp, configPath, false, cfgPath, io.Discard, io.Discard)
	if code != 0 {
		t.Errorf("--from-config alone: want code 0, got %d", code)
	}
	if err == nil || !strings.Contains(err.Error(), "must be set together") {
		t.Errorf("--from-config alone: want 'must be set together'; got %v", err)
	}
}

// TestAddCmd_FlagsRegistered: cobra has the flag definitions wired.
func TestAddCmd_FlagsRegistered(t *testing.T) {
	for _, name := range []string{"non-interactive", "from-config"} {
		if f := addCmd.Flags().Lookup(name); f == nil {
			t.Errorf("flag --%s not registered on addCmd", name)
		}
	}
}
