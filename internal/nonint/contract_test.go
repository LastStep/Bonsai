package nonint

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
)

// contract_test.go is the Plan 41 Phase 5 CROSS-command sweep. The per-command
// test files (runner_test / update_test / remove_test and the cmd/*_nonint_test
// adapters) each pin ONE command's behaviour. This file asserts the contract
// holds UNIFORMLY across every mutating core in one place — so a future core
// added to internal/nonint that forgets the JSONL/warnings discipline fails
// here even if its own per-command test is missing.

// runMutatingCore is one headless mutating core invoked through the same
// signature shape every adapter uses: it returns the (*Result, exitCode, error)
// triple. Each case materialises its own project so the cores run against a
// real workspace + lock on disk.
type mutatingCase struct {
	name string
	// run produces the core's Result. A non-nil error here is a SETUP failure
	// (t.Fatal), distinct from the core's own returned error.
	run func(t *testing.T) (*Result, int)
}

// allMutatingCores enumerates init / add / update / remove(agent) /
// remove(item). Each closure builds a fresh tmp project and drives exactly one
// core, returning its happy-path Result. Keeping all five in one slice is the
// point: the sweep below iterates them so the invariant is asserted identically
// for every command.
func allMutatingCores() []mutatingCase {
	return []mutatingCase{
		{
			name: "init",
			run: func(t *testing.T) (*Result, int) {
				cat := loadTestCatalog(t)
				tmp := t.TempDir()
				cfg := minimalInitCfg(t, tmp)
				res, code, err := RunInit(tmp, filepath.Join(tmp, ".bonsai.yaml"), cfg, cat, "test")
				if err != nil {
					t.Fatalf("RunInit setup: %v", err)
				}
				return res, code
			},
		},
		{
			name: "add",
			run: func(t *testing.T) (*Result, int) {
				cat := loadTestCatalog(t)
				tmp := t.TempDir()
				configPath := filepath.Join(tmp, ".bonsai.yaml")
				cfg := minimalInitCfg(t, tmp)
				if _, code, err := RunInit(tmp, configPath, cfg, cat, "test"); err != nil || code != ExitOK {
					t.Fatalf("RunInit (add setup): code=%d err=%v", code, err)
				}
				overlay, err := LoadConfig(writeYAML(t, tmp, "backend.yaml", "agents:\n  backend: {}\n"), tmp, cat)
				if err != nil {
					t.Fatalf("LoadConfig backend overlay: %v", err)
				}
				res, code, err := RunAdd(tmp, configPath, overlay, cat, "test")
				if err != nil {
					t.Fatalf("RunAdd setup: %v", err)
				}
				return res, code
			},
		},
		{
			name: "update",
			run: func(t *testing.T) (*Result, int) {
				cat := loadTestCatalog(t)
				tmp := t.TempDir()
				configPath := filepath.Join(tmp, ".bonsai.yaml")
				cfg := minimalInitCfg(t, tmp)
				if _, code, err := RunInit(tmp, configPath, cfg, cat, "test"); err != nil || code != ExitOK {
					t.Fatalf("RunInit (update setup): code=%d err=%v", code, err)
				}
				reloaded, err := config.Load(configPath)
				if err != nil {
					t.Fatalf("reload config: %v", err)
				}
				lock, _ := config.LoadLockFile(tmp)
				if lock == nil {
					lock = config.NewLockFile()
				}
				res, code, err := RunUpdate(tmp, reloaded, cat, lock, "test", false)
				if err != nil {
					t.Fatalf("RunUpdate setup: %v", err)
				}
				return res, code
			},
		},
		{
			name: "remove-agent",
			run: func(t *testing.T) (*Result, int) {
				cat := loadTestCatalog(t)
				tmp, configPath := initWithAgents(t, "backend")
				cfg := reloadConfig(t, configPath)
				lock, _ := config.LoadLockFile(tmp)
				res, code, err := RunRemoveAgent(tmp, cfg, cat, lock, "test", "backend", false)
				if err != nil {
					t.Fatalf("RunRemoveAgent setup: %v", err)
				}
				return res, code
			},
		},
		{
			name: "remove-item",
			run: func(t *testing.T) (*Result, int) {
				cat := loadTestCatalog(t)
				tmp, configPath := initWithAgents(t, "backend")
				cfg := reloadConfig(t, configPath)
				lock, _ := config.LoadLockFile(tmp)
				// coding-standards is a non-required skill backend installs by
				// default — a safe single-owner item to remove.
				res, code, err := RunRemoveItem(tmp, cfg, cat, lock, "test", "skill", "coding-standards", "backend")
				if err != nil {
					t.Fatalf("RunRemoveItem setup: %v", err)
				}
				return res, code
			},
		},
	}
}

// TestContract_AllMutatingCoresEmitPureJSONL is the cross-command stdout-purity
// sweep. For every mutating core, EmitJSONL over its happy-path Result must
// produce ONLY file/summary events (every line valid JSON, no other event
// kind) and exactly one terminal summary line. This is the single guard that a
// new core can't ship a non-conforming stdout shape.
func TestContract_AllMutatingCoresEmitPureJSONL(t *testing.T) {
	for _, tc := range allMutatingCores() {
		t.Run(tc.name, func(t *testing.T) {
			res, code := tc.run(t)
			if code != ExitOK {
				t.Fatalf("%s: want ExitOK, got %d", tc.name, code)
			}

			var buf bytes.Buffer
			if err := EmitJSONL(&buf, res); err != nil {
				t.Fatalf("%s: EmitJSONL: %v", tc.name, err)
			}

			records := parseJSONL(t, buf.String())
			summaries := 0
			for _, rec := range records {
				ev, _ := rec["event"].(string)
				if ev != "file" && ev != "summary" {
					t.Errorf("%s: stdout carries event %q; only file/summary allowed", tc.name, ev)
				}
				if ev == "summary" {
					summaries++
				}
			}
			if summaries != 1 {
				t.Errorf("%s: want exactly one terminal summary line, got %d", tc.name, summaries)
			}
			// The summary line must carry all five count keys (stable shape).
			summary := findSummary(records)
			if summary == nil {
				t.Fatalf("%s: no summary event emitted", tc.name)
			}
			for _, key := range []string{"created", "updated", "unchanged", "skipped", "conflicts"} {
				if _, ok := summary[key]; !ok {
					t.Errorf("%s: summary missing count key %q", tc.name, key)
				}
			}
		})
	}
}

// TestContract_AllMutatingCoresKeepWarningsOffStdout asserts the warnings
// discipline uniformly: for EVERY core, any Result.Warnings text must NOT leak
// into the EmitJSONL stdout stream. The cores route warnings to Result.Warnings
// (which the CLI adapter prints to stderr); the JSONL writer must never see them.
// We inject a synthetic warning into each core's real Result and confirm the
// serialized stream stays clean — proving the boundary holds regardless of how
// the core was reached.
func TestContract_AllMutatingCoresKeepWarningsOffStdout(t *testing.T) {
	const sentinel = "CONTRACT-SWEEP-SENTINEL-WARNING"
	for _, tc := range allMutatingCores() {
		t.Run(tc.name, func(t *testing.T) {
			res, code := tc.run(t)
			if code != ExitOK {
				t.Fatalf("%s: want ExitOK, got %d", tc.name, code)
			}
			res.Warnings = append(res.Warnings, sentinel)

			var buf bytes.Buffer
			if err := EmitJSONL(&buf, res); err != nil {
				t.Fatalf("%s: EmitJSONL: %v", tc.name, err)
			}
			if strings.Contains(buf.String(), sentinel) {
				t.Errorf("%s: warning text leaked onto the JSONL stdout stream:\n%s", tc.name, buf.String())
			}
			// And no line is a `warning` event.
			for _, rec := range parseJSONL(t, buf.String()) {
				if ev, _ := rec["event"].(string); ev == "warning" {
					t.Errorf("%s: warning event must never appear on stdout", tc.name)
				}
			}
		})
	}
}

// TestContract_ExitConstantsAreStable pins the mutating exit-code constants to
// their documented values. docs/agent-interface.md names internal/nonint/runner.go
// as the canonical source; this test is the machine-checked half of that
// contract — if a constant's value drifts the doc table is wrong and CI fails
// here (no markdown-parsing doc-drift test, by Plan 41 design).
func TestContract_ExitConstantsAreStable(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"ExitOK", ExitOK, 0},
		{"ExitInvalidConfig", ExitInvalidConfig, 2},
		{"ExitRuntime", ExitRuntime, 3},
		{"ExitWrongCWDForInit", ExitWrongCWDForInit, 4},
		{"ExitConflict", ExitConflict, 5},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d — docs/agent-interface.md exit-code table is now stale", c.name, c.got, c.want)
		}
	}
}
