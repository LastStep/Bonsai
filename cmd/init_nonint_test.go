package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/nonint"
)

// init_nonint_test.go drives the runInitNonInteractive helper directly so we
// can observe (exitCode, error) without trapping os.Exit. The helper is the
// same code path the cobra RunE branch invokes; the test contract is:
// helper-result + stdout/stderr ←→ user-observed behaviour.

// writeYAMLFixture drops a YAML config file in dir and returns its absolute
// path. Helper duplicated here (rather than imported from internal/nonint)
// so the cmd test stays standalone.
func writeYAMLFixture(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

// TestRunInitNonInteractive_MinimalSuccess: tech-lead-only fixture in a fresh
// tmp dir. Expect exit 0, .bonsai.yaml + station/ present, JSONL parseable,
// terminating summary event.
func TestRunInitNonInteractive_MinimalSuccess(t *testing.T) {
	setupListTestCatalog(t)
	tmp := t.TempDir()
	cfgPath := writeYAMLFixture(t, tmp, "cfg.yaml", "agents:\n  tech-lead: {}\n")
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	var stdout, stderr bytes.Buffer
	code, err := runInitNonInteractive(tmp, configPath, true, cfgPath, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runInitNonInteractive: %v (stderr=%s)", err, stderr.String())
	}
	if code != nonint.ExitOK {
		t.Fatalf("exit code: want %d, got %d (stderr=%s)", nonint.ExitOK, code, stderr.String())
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf(".bonsai.yaml missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "station")); err != nil {
		t.Errorf("station/ tree missing: %v", err)
	}

	// Validate JSONL stream — every line must parse, summary must be present.
	var sawSummary bool
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Errorf("bad JSON line %q: %v", line, err)
			continue
		}
		if rec["event"] == "summary" {
			sawSummary = true
		}
	}
	if !sawSummary {
		t.Errorf("no summary event in stdout:\n%s", stdout.String())
	}
}

// TestRunInitNonInteractive_MissingTechLead: overlay without tech-lead must
// exit 2 (Decision §1) and stderr must mention tech-lead.
func TestRunInitNonInteractive_MissingTechLead(t *testing.T) {
	setupListTestCatalog(t)
	tmp := t.TempDir()
	cfgPath := writeYAMLFixture(t, tmp, "cfg.yaml", "agents:\n  backend: {}\n")
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	var stdout, stderr bytes.Buffer
	code, err := runInitNonInteractive(tmp, configPath, true, cfgPath, &stdout, &stderr)
	if code != nonint.ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", nonint.ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "tech-lead") {
		t.Errorf("error must mention tech-lead; got %v", err)
	}
	if !strings.Contains(stderr.String(), "tech-lead") {
		t.Errorf("stderr must mention tech-lead; got %q", stderr.String())
	}
}

// TestRunInitNonInteractive_PreExistingConfig: a project with .bonsai.yaml
// already present must exit 4. Mirrors the legacy interactive "Skipping
// init" branch's intent but with a non-zero exit so the caller knows.
func TestRunInitNonInteractive_PreExistingConfig(t *testing.T) {
	setupListTestCatalog(t)
	tmp := t.TempDir()
	cfgPath := writeYAMLFixture(t, tmp, "cfg.yaml", "agents:\n  tech-lead: {}\n")
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	// Pre-seed an existing .bonsai.yaml — content irrelevant, just the
	// presence triggers exit 4.
	if err := os.WriteFile(configPath, []byte("project_name: prior\nagents: {}\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code, err := runInitNonInteractive(tmp, configPath, true, cfgPath, &stdout, &stderr)
	if code != nonint.ExitWrongCWDForInit {
		t.Errorf("exit code: want %d, got %d", nonint.ExitWrongCWDForInit, code)
	}
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error must mention 'already exists'; got %v", err)
	}
}

// TestRunInitNonInteractive_NonInteractiveAlone: --non-interactive without
// --from-config must yield exit-code 0 (cobra-handled error path) with a
// "must be set together" error, no os.Exit. Caller's cobra wrapper surfaces
// it via the usage banner.
func TestRunInitNonInteractive_NonInteractiveAlone(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	var stdout, stderr bytes.Buffer
	code, err := runInitNonInteractive(tmp, configPath, true, "", &stdout, &stderr)
	if code != 0 {
		t.Errorf("exit code: want 0 (cobra-handled), got %d", code)
	}
	if err == nil || !strings.Contains(err.Error(), "must be set together") {
		t.Errorf("error: want 'must be set together'; got %v", err)
	}
}

// TestRunInitNonInteractive_FromConfigAlone: --from-config without
// --non-interactive is symmetric — must yield the same usage error.
func TestRunInitNonInteractive_FromConfigAlone(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := writeYAMLFixture(t, tmp, "cfg.yaml", "agents:\n  tech-lead: {}\n")
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	var stdout, stderr bytes.Buffer
	code, err := runInitNonInteractive(tmp, configPath, false, cfgPath, &stdout, &stderr)
	if code != 0 {
		t.Errorf("exit code: want 0 (cobra-handled), got %d", code)
	}
	if err == nil || !strings.Contains(err.Error(), "must be set together") {
		t.Errorf("error: want 'must be set together'; got %v", err)
	}
}

// TestRunInitNonInteractive_BadYAML: malformed YAML must exit 2 + stderr
// "parse YAML". Plan 39 §B test list.
func TestRunInitNonInteractive_BadYAML(t *testing.T) {
	setupListTestCatalog(t)
	tmp := t.TempDir()
	cfgPath := writeYAMLFixture(t, tmp, "cfg.yaml", "agents: [this: is: not valid yaml")
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	var stdout, stderr bytes.Buffer
	code, err := runInitNonInteractive(tmp, configPath, true, cfgPath, &stdout, &stderr)
	if code != nonint.ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", nonint.ExitInvalidConfig, code)
	}
	if err == nil {
		t.Errorf("expected an error for bad YAML")
	}
	if !strings.Contains(stderr.String(), "from-config") {
		t.Errorf("stderr must mention from-config; got %q", stderr.String())
	}
}

// TestInitCmd_FlagsRegistered confirms the cobra flags are wired so
// `--help` rendering picks them up. Cheap regression guard against
// accidental flag removal.
func TestInitCmd_FlagsRegistered(t *testing.T) {
	for _, name := range []string{"non-interactive", "from-config"} {
		if f := initCmd.Flags().Lookup(name); f == nil {
			t.Errorf("flag --%s not registered on initCmd", name)
		}
	}
}
