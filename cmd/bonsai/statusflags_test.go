package main

// status --full and --active from the command line (step 5.1.6): --full asks Claude Code through the same seams as
// check (pluginCLI, claudeVersion: fakes here, never the real claude) and fits bonsai.status/1; --active prints the
// active task alone, its --json held to active_task's schema, and the error object alone when the workspace cannot
// be read.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/status"
)

func TestStatusFullCommand(t *testing.T) {
	c := newCLI(t)
	c.link()
	fakeClaude(t, "2.1.300 (Claude Code)", nil)
	setPlugins(t, fakeList{})
	code, out, _ := c.run("", "status", "--full", "--json")
	doc := fits(t, out, "status")
	checks, _ := doc.Get("checks")
	needs, _ := doc.Get("needs")
	if code != 0 || doc.String("mode") != status.ModeFull || !strings.Contains(schema.Show(checks), `"claude_code":{"version":"2.1.300"`) ||
		!strings.Contains(schema.Show(needs), `"kind":"plugin","id":"demo-pack"`) {
		t.Errorf("status --full --json: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "status", "--full")
	if code != 0 || !strings.Contains(out, "Checks (--full):\n") || !strings.Contains(out, "  Claude Code: 2.1.300, ok (the floor "+engine.ClaudeCodeFloor) ||
		!strings.Contains(out, "Needs from this machine: claude-code >="+engine.ClaudeCodeFloor) {
		t.Errorf("status --full: %d\n%s", code, out)
	}
	// Without --full, Claude Code is never asked.
	asked := false
	claudeVersion = func() (string, error) { asked = true; return "", nil }
	if code, out, _ := c.run("", "status", "--json"); code != 0 || asked || fits(t, out, "status").String("mode") != status.Mode {
		t.Errorf("status --json asked Claude Code (%v) or was not offline: %d", asked, code)
	}
}

func TestStatusActiveCommand(t *testing.T) {
	c := newCLI(t)
	c.link()
	t.Setenv("BONSAI_TASK", "")
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	c.task("T-0901", "running", "", "")
	code, out, _ := c.run("", "status", "--active")
	if code != 0 || out != "Active task: T-0901 (running)\n" {
		t.Errorf("status --active: %d %q", code, out)
	}
	code, out, _ = c.run("", "status", "--active", "--json")
	if code != 0 || strings.TrimSpace(out) != "{\n  \"id\": \"T-0901\",\n  \"how\": \"running\",\n  \"why\": null\n}" {
		t.Errorf("status --active --json: %d %q", code, out)
	}
	c.task("T-0902", "running", "", "")
	if code, out, _ := c.run("", "status", "--active"); code != 0 || out != "Active task: none (two or more tasks read running (T-0901, T-0902))\n" {
		t.Errorf("two running: %d %q", code, out)
	}
	// Outside a checkout: exit 3, the error object alone with --json, a next step on stderr without.
	outside := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(outside))
	t.Chdir(outside)
	code, out, _ = c.run("", "status", "--active", "--json")
	e := fits(t, out, "error")
	if code != 3 || e.String("code") != "not-a-checkout" {
		t.Errorf("outside: %d\n%s", code, out)
	}
	if code, out, errOut := c.run("", "status", "--active"); code != 3 || out != "" || !strings.Contains(errOut, "\nnext: ") {
		t.Errorf("outside, text: %d %q %q", code, out, errOut)
	}
}
