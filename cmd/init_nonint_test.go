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

// TestRunInitNonInteractive_StreamSeparation is the C5 stream-hygiene
// invariant (Plan 41 Phase 1.7): driving the HELPER (not os.Exit), stdout
// must be pure JSONL where every non-empty line is a known event shape
// (file/summary), and stderr must carry no `{`-leading JSON line. This is the
// hard prerequisite for the Plan 42 stdio MCP server — stdout is pure
// protocol, all diagnostics go to stderr.
func TestRunInitNonInteractive_StreamSeparation(t *testing.T) {
	setupListTestCatalog(t)
	tmp := t.TempDir()
	cfgPath := writeYAMLFixture(t, tmp, "cfg.yaml", "agents:\n  tech-lead: {}\n")
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	var stdout, stderr bytes.Buffer
	code, err := runInitNonInteractive(tmp, configPath, true, cfgPath, &stdout, &stderr)
	if err != nil || code != nonint.ExitOK {
		t.Fatalf("runInitNonInteractive: code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	assertStreamSeparation(t, stdout.String(), stderr.String())
}

// assertStreamSeparation enforces the C5 invariant on a captured stdout/stderr
// pair: every non-empty stdout line unmarshals to a file/summary event, and
// no stderr line starts with `{` (no JSON leaked onto the diagnostic stream).
func assertStreamSeparation(t *testing.T, stdout, stderr string) {
	t.Helper()
	sawSummary := false
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Errorf("stdout line is not JSON: %q (%v)", line, err)
			continue
		}
		ev, _ := rec["event"].(string)
		if ev != "file" && ev != "summary" {
			t.Errorf("stdout carries unknown event %q; only file/summary allowed: %q", ev, line)
		}
		if ev == "summary" {
			sawSummary = true
		}
	}
	if !sawSummary {
		t.Errorf("stdout missing terminal summary event:\n%s", stdout)
	}
	for _, line := range strings.Split(strings.TrimSpace(stderr), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "{") {
			t.Errorf("stderr carries a JSON line — data leaked onto the diagnostic stream: %q", line)
		}
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

// TestRunInitNonInteractive_MultiAgentOverlay_Rejected: overlay with
// tech-lead PLUS extras must exit 2 (Decision §1 exclusivity). Stderr
// must hint at the single-entry rule and the .bonsai.yaml must not be
// written.
func TestRunInitNonInteractive_MultiAgentOverlay_Rejected(t *testing.T) {
	setupListTestCatalog(t)
	tmp := t.TempDir()
	cfgPath := writeYAMLFixture(t, tmp, "cfg.yaml", "agents:\n  tech-lead: {}\n  backend: {}\n")
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	var stdout, stderr bytes.Buffer
	code, err := runInitNonInteractive(tmp, configPath, true, cfgPath, &stdout, &stderr)
	if code != nonint.ExitInvalidConfig {
		t.Errorf("exit code: want %d, got %d", nonint.ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "single") {
		t.Errorf("error must mention exclusivity; got %v", err)
	}
	if !strings.Contains(stderr.String(), "single") {
		t.Errorf("stderr must mention exclusivity; got %q", stderr.String())
	}
	if _, statErr := os.Stat(configPath); statErr == nil {
		t.Errorf(".bonsai.yaml must not be written on rejection")
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
