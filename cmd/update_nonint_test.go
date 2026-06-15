package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/nonint"
)

// update_nonint_test.go drives the runUpdateNonInteractive helper directly so
// we can observe the exit code + stdout/stderr buffers without trapping
// os.Exit. The helper is the same code path the cobra headless gate invokes.

// initForUpdate materialises a tech-lead project via nonint.RunInit so the
// update tests have a real workspace + .bonsai.yaml + lock on disk. Returns
// the project cwd.
func initForUpdate(t *testing.T, cwd string) {
	t.Helper()
	cat := loadCatalog()
	cfgPath := writeYAMLFixture(t, cwd, "init.yaml", "agents:\n  tech-lead: {}\n")
	cfg, err := nonint.LoadConfig(cfgPath, cwd, cat)
	if err != nil {
		t.Fatalf("LoadConfig (init): %v", err)
	}
	if _, code, err := nonint.RunInit(cwd, filepath.Join(cwd, configFile), cfg, cat, Version); err != nil || code != nonint.ExitOK {
		t.Fatalf("RunInit (setup): code=%d err=%v", code, err)
	}
}

// reloadForUpdate loads the project config + lock the way runUpdate does before
// invoking the headless adapter.
func reloadForUpdate(t *testing.T, cwd string) (*config.ProjectConfig, *config.LockFile) {
	t.Helper()
	cfg, err := config.Load(filepath.Join(cwd, configFile))
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	lock, _ := config.LoadLockFile(cwd)
	if lock == nil {
		lock = config.NewLockFile()
	}
	return cfg, lock
}

// TestRunUpdateNonInteractive_StreamSeparation drives the headless adapter and
// asserts the C5 invariant: every non-empty stdout line unmarshals to a
// file/summary event, and no stderr line carries `{`-leading JSON. A custom
// skill with valid frontmatter exercises the apply path; an invalid one
// exercises the warnings→stderr path.
func TestRunUpdateNonInteractive_StreamSeparation(t *testing.T) {
	setupListTestCatalog(t)
	tmp := t.TempDir()
	initForUpdate(t, tmp)

	skillsDir := filepath.Join(tmp, "station", "agent", "Skills")
	validBody := "---\ndescription: A user-authored custom skill\ndisplay_name: My Custom Skill\n---\n\n# My Custom Skill\n"
	if err := os.WriteFile(filepath.Join(skillsDir, "my-custom.md"), []byte(validBody), 0o644); err != nil {
		t.Fatalf("write valid custom skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "broken.md"), []byte("# no frontmatter\n"), 0o644); err != nil {
		t.Fatalf("write invalid custom skill: %v", err)
	}

	cfg, lock := reloadForUpdate(t, tmp)
	var stdout, stderr bytes.Buffer
	code := runUpdateNonInteractive(tmp, cfg, loadCatalog(), lock, false, &stdout, &stderr)
	if code != nonint.ExitOK {
		t.Fatalf("exit code: want %d, got %d (stderr=%s)", nonint.ExitOK, code, stderr.String())
	}
	assertUpdateStreamSeparation(t, stdout.String(), stderr.String())
	// The invalid file must surface on stderr (a warning), never on stdout.
	if !strings.Contains(stderr.String(), "broken.md") {
		t.Errorf("invalid discovery must surface on stderr; got %q", stderr.String())
	}
	if strings.Contains(stdout.String(), "broken.md") {
		t.Errorf("invalid discovery leaked onto stdout:\n%s", stdout.String())
	}
}

// assertUpdateStreamSeparation enforces the C5 invariant on a stdout/stderr
// pair: every non-empty stdout line is a known file/summary event and no
// stderr line starts with `{`.
func assertUpdateStreamSeparation(t *testing.T, stdout, stderr string) {
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

// TestRunUpdateNonInteractive_ConflictExit5 is the C2 negative control: a
// user-edited generator-target file with no lock entry surfaces as a conflict.
// Without --skip-conflicts the adapter returns ExitConflict (5) and the JSONL
// stream carries an action=conflict file event for that path. With
// --skip-conflicts it returns 0 and the file is counted skipped (untouched).
func TestRunUpdateNonInteractive_ConflictExit5(t *testing.T) {
	setupListTestCatalog(t)
	tmp := t.TempDir()
	initForUpdate(t, tmp)

	// Overwrite a stable generator output with user bytes and drop its lock
	// entry — mirrors TestRunInit_ConflictEmittedNotForced.
	userBody := []byte("# USER EDIT — must not be overwritten\n")
	relPath := "station/agent/Core/identity.md"
	target := filepath.Join(tmp, relPath)
	if err := os.WriteFile(target, userBody, 0o644); err != nil {
		t.Fatalf("seed conflict: %v", err)
	}
	lock0, _ := config.LoadLockFile(tmp)
	delete(lock0.Files, relPath)
	if err := lock0.Save(tmp); err != nil {
		t.Fatalf("re-save lock: %v", err)
	}

	// skipConflicts=false → exit 5, conflict event present.
	cfg, lock := reloadForUpdate(t, tmp)
	var stdout, stderr bytes.Buffer
	code := runUpdateNonInteractive(tmp, cfg, loadCatalog(), lock, false, &stdout, &stderr)
	if code != nonint.ExitConflict {
		t.Fatalf("exit code: want %d (ExitConflict), got %d (stderr=%s)", nonint.ExitConflict, code, stderr.String())
	}
	if !hasFileEvent(t, stdout.String(), relPath, "conflict") {
		t.Errorf("expected action=conflict file event for %q in stdout:\n%s", relPath, stdout.String())
	}
	if got, _ := os.ReadFile(target); !bytes.Equal(got, userBody) {
		t.Errorf("conflict file overwritten; want %q got %q", userBody, got)
	}

	// skipConflicts=true → exit 0, file counted skipped in the summary.
	cfg2, lock2 := reloadForUpdate(t, tmp)
	var stdout2, stderr2 bytes.Buffer
	code2 := runUpdateNonInteractive(tmp, cfg2, loadCatalog(), lock2, true, &stdout2, &stderr2)
	if code2 != nonint.ExitOK {
		t.Fatalf("exit code (skip): want %d, got %d (stderr=%s)", nonint.ExitOK, code2, stderr2.String())
	}
	if skipped := summarySkipped(t, stdout2.String()); skipped == 0 {
		t.Errorf("skipConflicts must count the conflict as skipped; summary skipped=%d\n%s", skipped, stdout2.String())
	}
	if got, _ := os.ReadFile(target); !bytes.Equal(got, userBody) {
		t.Errorf("conflict file overwritten under skip; want %q got %q", userBody, got)
	}
}

// hasFileEvent reports whether the JSONL stream contains a file event for path
// with the given action.
func hasFileEvent(t *testing.T, stream, path, action string) bool {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(stream), "\n") {
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		if rec["event"] == "file" && rec["path"] == path && rec["action"] == action {
			return true
		}
	}
	return false
}

// summarySkipped extracts the `skipped` count from the terminal summary event.
func summarySkipped(t *testing.T, stream string) int {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(stream), "\n") {
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		if rec["event"] == "summary" {
			if v, ok := rec["skipped"].(float64); ok {
				return int(v)
			}
		}
	}
	t.Fatalf("no summary event with skipped count in stream:\n%s", stream)
	return 0
}

// TestUpdateCmd_FlagsRegistered confirms the new headless flags are wired so
// `--help` picks them up. Cheap regression guard against accidental removal.
func TestUpdateCmd_FlagsRegistered(t *testing.T) {
	for _, name := range []string{"non-interactive", "skip-conflicts"} {
		if f := updateCmd.Flags().Lookup(name); f == nil {
			t.Errorf("flag --%s not registered on updateCmd", name)
		}
	}
}
