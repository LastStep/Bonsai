package main

// BUG-106 (T-5736): bonsai check warns, config-uncommitted, while the main checkout's repository holds no commit, or
// its bonsai.yaml is not in HEAD or differs from HEAD's, which a climb (config-uncommitted) and the stop gate
// (stop-config-uncommitted) refuse; read as they read it (stopgate.CommittedConfig). A warning, never a finding: a
// fresh link is uncommitted by nature. BUG-107's words reach init's and update's output, text and --json.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// uncommittedWarning runs check --json and gives its config-uncommitted warning's message and next step, "" for
// none; it fails the test when the word is among the findings, or check's exit is not the findings' alone.
func uncommittedWarning(t *testing.T, c *cli) (msg, next, who string) {
	t.Helper()
	code, out, _ := c.run("", "check", "--json")
	doc := fits(t, out, "check")
	findings, _ := doc.Get("findings")
	if want := map[bool]int{true: 1, false: 0}[len(findings.([]any)) > 0]; code != want {
		t.Errorf("check exits %d with %d findings:\n%s", code, len(findings.([]any)), out)
	}
	if hasCode(findings.([]any), "config-uncommitted") {
		t.Errorf("config-uncommitted is a finding:\n%s", out)
	}
	warnings, _ := doc.Get("warnings")
	for _, w := range warnings.([]any) {
		wo := w.(schema.Object)
		if wo.String("code") == "config-uncommitted" {
			n, _ := wo.Get("next")
			return wo.String("message"), n.(schema.Object).String("do"), n.(schema.Object).String("who")
		}
	}
	return "", "", ""
}

func TestCheckUncommitted(t *testing.T) {
	c := newCLI(t)
	c.link()
	msg, next, who := uncommittedWarning(t, c)
	if !strings.Contains(msg, "no commit yet") || !strings.Contains(msg, "a climb and the stop gate refuse") ||
		!strings.Contains(next, "commits what bonsai init wrote") || who != "person" {
		t.Errorf("no commit: %q %q %q", msg, next, who)
	}
	if _, text, _ := c.run("", "check"); !strings.Contains(text, "config-uncommitted") {
		t.Errorf("check's text has no config-uncommitted:\n%s", text)
	}

	c.commit("link")
	if msg, _, _ := uncommittedWarning(t, c); msg != "" {
		t.Errorf("committed: %q", msg)
	}

	c.write("bonsai.yaml", c.read("bonsai.yaml")+"# an edit\n")
	msg, next, who = uncommittedWarning(t, c)
	if !strings.Contains(msg, "differs from the one committed at its HEAD") || !strings.Contains(next, "reverts it") || who != "person" {
		t.Errorf("an edit: %q %q %q", msg, next, who)
	}
	// Line ends read as LF, as the ladder and the gate read them: a CRLF copy of the committed file is the same file.
	testpack.Git(t, c.root, "checkout", "--", "bonsai.yaml")
	c.write("bonsai.yaml", strings.ReplaceAll(c.read("bonsai.yaml"), "\n", "\r\n"))
	if msg, _, _ := uncommittedWarning(t, c); msg != "" {
		t.Errorf("CRLF line ends: %q", msg)
	}

	// A repository with commits, none holding bonsai.yaml: linked after them.
	d := newCLI(t)
	testpack.Git(t, d.root, "commit", "-q", "--allow-empty", "-m", "before")
	d.link()
	msg, next, _ = uncommittedWarning(t, d)
	if !strings.Contains(msg, "never committed") || !strings.Contains(next, "commits what bonsai init wrote") {
		t.Errorf("not in HEAD: %q %q", msg, next)
	}
}

// init and update, when they install a pack's plugin under a new marketplace name, say that a session already open
// in the checkout loads its skills only after a reload or a fresh start: in the text, and in --json's plugins.
func TestReloadWordsInOutput(t *testing.T) {
	c := newCLI(t)
	setPlugins(t, &fakePlugins{install: engine.InstallResult{Outcome: "ok"}})
	code, out, _ := c.run("", append(c.linkArgs(c.pack.A), "--yes")...)
	if code != 0 || !strings.Contains(flat(out), "/reload-plugins") || !strings.Contains(flat(out), "a fresh start of Claude Code") {
		t.Errorf("init --yes: %d\n%s", code, out)
	}
	c.commit("link")
	if code, out, _ := c.run("", "update", "--yes"); code != 0 || strings.Contains(out, "reload-plugins") {
		t.Errorf("an update that keeps the marketplace's name: %d\n%s", code, out)
	}
	testpack.SetRef(t, c.root, c.pack.A, c.pack.B)
	code, out, _ = c.run("", "update", "--yes", "--json")
	doc := fits(t, out, "changes")
	plugins, _ := doc.Get("plugins")
	said := false
	for _, p := range plugins.([]any) {
		po := p.(schema.Object)
		if po.String("result") == "installed" && strings.HasSuffix(po.String("message"), "; "+engine.ReloadWords) {
			said = true
		}
	}
	if code != 0 || !said {
		t.Errorf("update --yes --json to another commit: %d\n%s", code, out)
	}
}
