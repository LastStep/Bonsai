package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/nonint"
)

// remove_nonint_test.go drives the headless remove adapters
// (runRemoveAgentNonInteractive / runRemoveItemNonInteractive) directly so we
// observe (exitCode, error) + stdout/stderr without trapping os.Exit. The
// adapters are the same code path the cobra gate invokes.

// addAgentNonInt adds a single agent to an already-initialised project via the
// add adapter. Used to build multi-agent fixtures for the cmd-level tests.
func addAgentNonInt(t *testing.T, tmp, agentType string) {
	t.Helper()
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	overlayPath := writeYAMLFixture(t, tmp, agentType+"-overlay.yaml", "agents:\n  "+agentType+": {}\n")
	var stdout, stderr bytes.Buffer
	code, err := runAddNonInteractive(tmp, configPath, true, overlayPath, &stdout, &stderr)
	if err != nil || code != nonint.ExitOK {
		t.Fatalf("add %s setup: code=%d err=%v stderr=%s", agentType, code, err, stderr.String())
	}
}

// ─── Flag registration ──────────────────────────────────────────────────

// TestRemoveCmd_FlagsRegistered asserts --non-interactive / --yes (+ -y) live
// on removeCmd and --from is reachable. The headless flags are persistent so
// the item subcommands inherit them.
func TestRemoveCmd_FlagsRegistered(t *testing.T) {
	for _, name := range []string{"non-interactive", "yes", "from"} {
		if f := removeCmd.PersistentFlags().Lookup(name); f == nil {
			t.Errorf("flag --%s not registered on removeCmd (persistent)", name)
		}
	}
	if f := removeCmd.PersistentFlags().ShorthandLookup("y"); f == nil {
		t.Errorf("shorthand -y for --yes not registered")
	}
	// Item subcommands must inherit the persistent flags. cobra exposes parent
	// persistent flags via InheritedFlags() (it merges them into Flags() lazily
	// during Execute; InheritedFlags is the stable pre-execute view).
	for _, sub := range []string{"skill", "workflow", "protocol", "sensor", "routine"} {
		c, _, err := removeCmd.Find([]string{sub})
		if err != nil {
			t.Fatalf("find subcommand %s: %v", sub, err)
		}
		for _, name := range []string{"non-interactive", "yes", "from"} {
			if f := c.InheritedFlags().Lookup(name); f == nil {
				t.Errorf("subcommand %s missing inherited flag --%s", sub, name)
			}
		}
	}
}

// ─── Agent removal: tech-lead guard ─────────────────────────────────────

// TestRunRemoveAgentNonInteractive_TechLeadGuard: init + backend, removing
// tech-lead → exit 2 with stderr mentioning tech-lead. Remove backend then
// tech-lead → exit 0.
func TestRunRemoveAgentNonInteractive_TechLeadGuard(t *testing.T) {
	tmp := initFreshProject(t)
	addAgentNonInt(t, tmp, "backend")

	var stdout, stderr bytes.Buffer
	code, err := runRemoveAgentNonInteractive(tmp, "tech-lead", false, &stdout, &stderr)
	if code != nonint.ExitInvalidConfig {
		t.Fatalf("remove tech-lead in use: want exit %d, got %d", nonint.ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "tech-lead") {
		t.Errorf("error must mention tech-lead; got %v", err)
	}
	if !strings.Contains(stderr.String(), "tech-lead") {
		t.Errorf("stderr must carry the tech-lead message; got %q", stderr.String())
	}

	// Remove backend, then tech-lead.
	stdout.Reset()
	stderr.Reset()
	if code, err := runRemoveAgentNonInteractive(tmp, "backend", false, &stdout, &stderr); err != nil || code != nonint.ExitOK {
		t.Fatalf("remove backend: code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code, err := runRemoveAgentNonInteractive(tmp, "tech-lead", false, &stdout, &stderr); err != nil || code != nonint.ExitOK {
		t.Fatalf("remove sole tech-lead: code=%d err=%v stderr=%s", code, err, stderr.String())
	}
}

// ─── Stream separation ──────────────────────────────────────────────────

// TestRunRemoveAgentNonInteractive_StreamSeparation drives the agent-removal
// adapter and asserts stdout is pure JSONL (file/summary) and stderr carries
// no `{`-leading JSON. Reuses assertStreamSeparation from init_nonint_test.go.
func TestRunRemoveAgentNonInteractive_StreamSeparation(t *testing.T) {
	tmp := initFreshProject(t)
	addAgentNonInt(t, tmp, "backend")

	var stdout, stderr bytes.Buffer
	code, err := runRemoveAgentNonInteractive(tmp, "backend", false, &stdout, &stderr)
	if err != nil || code != nonint.ExitOK {
		t.Fatalf("remove backend: code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	assertStreamSeparation(t, stdout.String(), stderr.String())
}

// TestRunRemoveItemNonInteractive_StreamSeparation drives the item-removal
// adapter (single-owner skill) and asserts the same stream invariant.
func TestRunRemoveItemNonInteractive_StreamSeparation(t *testing.T) {
	tmp := initFreshProject(t)
	addAgentNonInt(t, tmp, "backend")

	// database-conventions is a backend-only default (not on tech-lead, not
	// required) — a clean single-owner removal.
	var stdout, stderr bytes.Buffer
	code, err := runRemoveItemNonInteractive(tmp, "skill", "database-conventions", "", &stdout, &stderr)
	if err != nil || code != nonint.ExitOK {
		t.Fatalf("remove skill: code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	assertStreamSeparation(t, stdout.String(), stderr.String())
}

// ─── Multi-owner disambiguation ─────────────────────────────────────────

// TestRunRemoveItemNonInteractive_MultiOwner: coding-standards is shared by
// backend + security. No --from → exit 2, message names both owners, and the
// error must NOT be JSON on stderr. With --from backend → exit 0.
func TestRunRemoveItemNonInteractive_MultiOwner(t *testing.T) {
	tmp := initFreshProject(t)
	addAgentNonInt(t, tmp, "backend")
	addAgentNonInt(t, tmp, "security")

	var stdout, stderr bytes.Buffer
	code, err := runRemoveItemNonInteractive(tmp, "skill", "coding-standards", "", &stdout, &stderr)
	if code != nonint.ExitInvalidConfig {
		t.Fatalf("multi-owner no --from: want exit %d, got %d", nonint.ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "backend") || !strings.Contains(err.Error(), "security") {
		t.Errorf("error must name both owners; got %v", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(stderr.String()), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "{") {
			t.Errorf("stderr carries JSON — diagnostics leaked: %q", line)
		}
	}
	// stdout must be empty on the reject path (no Result emitted).
	if strings.TrimSpace(stdout.String()) != "" {
		t.Errorf("stdout must be empty on a rejected removal; got %q", stdout.String())
	}

	// --from backend → exit 0.
	stdout.Reset()
	stderr.Reset()
	if code, err := runRemoveItemNonInteractive(tmp, "skill", "coding-standards", "backend", &stdout, &stderr); err != nil || code != nonint.ExitOK {
		t.Fatalf("--from backend: code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	post, _ := config.Load(filepath.Join(tmp, ".bonsai.yaml"))
	if itemInSliceCmd(post.Agents["backend"].Skills, "coding-standards") {
		t.Errorf("coding-standards still on backend")
	}
	if !itemInSliceCmd(post.Agents["security"].Skills, "coding-standards") {
		t.Errorf("coding-standards wrongly removed from security")
	}
}

// ─── Required item + --from (H1 control) ────────────────────────────────

// TestRunRemoveItemNonInteractive_RequiredFrom: removing the required security
// protocol with --from tech-lead → exit 2, ZERO filesystem mutation.
func TestRunRemoveItemNonInteractive_RequiredFrom(t *testing.T) {
	tmp := initFreshProject(t)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	before, _ := os.ReadFile(configPath)
	protoFile := filepath.Join(tmp, "station", "agent", "Protocols", "security.md")
	if _, err := os.Stat(protoFile); err != nil {
		t.Fatalf("fixture: security protocol file missing: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code, err := runRemoveItemNonInteractive(tmp, "protocol", "security", "tech-lead", &stdout, &stderr)
	if code != nonint.ExitInvalidConfig {
		t.Fatalf("required --from: want exit %d, got %d", nonint.ExitInvalidConfig, code)
	}
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Errorf("error must mention required; got %v", err)
	}
	after, _ := os.ReadFile(configPath)
	if string(after) != string(before) {
		t.Errorf("config mutated removing a required item via --from")
	}
	if _, err := os.Stat(protoFile); err != nil {
		t.Errorf("protocol file deleted despite required refusal: %v", err)
	}
}

// ─── Empty / wildcard targets ───────────────────────────────────────────

// TestRunRemoveNonInteractive_UnsafeTargets: empty and "*" targets are
// rejected with exit 2 and zero mutation, for both agent and item removal.
func TestRunRemoveNonInteractive_UnsafeTargets(t *testing.T) {
	tmp := initFreshProject(t)
	addAgentNonInt(t, tmp, "backend")
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	before, _ := os.ReadFile(configPath)

	for _, bad := range []string{"", "*"} {
		var so, se bytes.Buffer
		code, _ := runRemoveAgentNonInteractive(tmp, bad, true, &so, &se)
		if code != nonint.ExitInvalidConfig {
			t.Errorf("agent target %q: want exit %d, got %d", bad, nonint.ExitInvalidConfig, code)
		}
		so.Reset()
		se.Reset()
		code, _ = runRemoveItemNonInteractive(tmp, "skill", bad, "", &so, &se)
		if code != nonint.ExitInvalidConfig {
			t.Errorf("item target %q: want exit %d, got %d", bad, nonint.ExitInvalidConfig, code)
		}
	}
	after, _ := os.ReadFile(configPath)
	if string(after) != string(before) {
		t.Errorf("config mutated by a rejected unsafe target")
	}
}

// ─── Symlink refusal ────────────────────────────────────────────────────

// TestRunRemoveAgentNonInteractive_SymlinkRefused: with --delete-files, each
// of the three delete targets replaced by a symlink in turn → exit 2, zero
// deletion. Drives the adapter (the cobra gate's real entry point).
func TestRunRemoveAgentNonInteractive_SymlinkRefused(t *testing.T) {
	rels := map[string]string{
		"agentDir":  "agent",
		"CLAUDE.md": "CLAUDE.md",
		".claude":   ".claude",
	}
	for label, rel := range rels {
		t.Run(label, func(t *testing.T) {
			tmp := initFreshProject(t)
			addAgentNonInt(t, tmp, "backend")
			cfg, _ := config.Load(filepath.Join(tmp, ".bonsai.yaml"))
			ws := cfg.Agents["backend"].Workspace

			sentinel := t.TempDir()
			linkPath := filepath.Join(tmp, ws, rel)
			_ = os.RemoveAll(linkPath)
			if err := os.Symlink(sentinel, linkPath); err != nil {
				t.Fatalf("symlink: %v", err)
			}

			var so, se bytes.Buffer
			code, err := runRemoveAgentNonInteractive(tmp, "backend", true, &so, &se)
			if code != nonint.ExitInvalidConfig {
				t.Fatalf("symlinked %s: want exit %d, got %d (err=%v)", label, nonint.ExitInvalidConfig, code, err)
			}
			if err == nil || !strings.Contains(err.Error(), "symlink") {
				t.Errorf("error must mention symlink; got %v", err)
			}
			if _, lerr := os.Lstat(linkPath); lerr != nil {
				t.Errorf("symlink deleted despite refusal: %v", lerr)
			}
			if _, serr := os.Stat(sentinel); serr != nil {
				t.Errorf("symlink target followed and deleted: %v", serr)
			}
		})
	}
}

// itemInSliceCmd is a local copy so the cmd test stays standalone.
func itemInSliceCmd(list []string, name string) bool {
	for _, s := range list {
		if s == name {
			return true
		}
	}
	return false
}
