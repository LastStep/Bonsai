package main

// bonsai unlink from the command line (step 5.1.7): the preview and exit 4 with no --yes and no terminal, y/N at one,
// --json in bonsai.changes/1, the plugin step after the write, a second run, and init after it.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
)

func TestUnlinkCommand(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.commit("link")
	market := engine.MarketplaceName("demo", []string{c.pack.A})
	f := &fakePlugins{list: []engine.InstalledPlugin{{ID: "demo-pack@" + market, Version: c.pack.A[:12], Scope: "project", ProjectPath: c.root}}}
	defer func(p engine.PluginCLI) { pluginCLI = p }(pluginCLI)
	pluginCLI = f
	files := strings.Join(c.files(), " ")

	// No --yes and no terminal: the preview, exit 4, nothing written, nothing asked of Claude Code.
	code, out, _ := c.run("", "unlink")
	if code != 4 || !strings.HasPrefix(out, "bonsai unlink: the preview.\nPacks:\n  demo-pack 0.1.0  "+c.pack.A[:7]+" -> taken out") ||
		!strings.Contains(out, "  removed      bonsai.yaml: removed even if a person edited it") || !strings.Contains(out, "\nLeft in place:\n") ||
		!strings.HasSuffix(out, "Nothing written yet.\nnext: to write it, run: bonsai unlink --yes\n") || len(f.asked) != 0 ||
		strings.Join(c.files(), " ") != files {
		t.Errorf("unlink with no --yes: %d %v\n%s", code, f.asked, out)
	}
	code, out, _ = c.run("", "unlink", "--json")
	doc := fits(t, out, "changes")
	if lock, _ := doc.Get("lock"); code != 4 || doc.String("command") != "unlink" || doc.String("result") != "preview" || lock != "removed" ||
		!strings.Contains(out, `"to": null`) || whoOf(errorIn(doc, "changes")) != "person" {
		t.Errorf("unlink --json: %d\n%s", code, out)
	}
	// At a terminal: n declines; y writes.
	if code, out, _ := c.run("n\n", "unlink"); code != 4 || !strings.Contains(out, "Take Bonsai out of this checkout? [y/N] Nothing written.") ||
		strings.Join(c.files(), " ") != files {
		t.Errorf("unlink, n: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "unlink", "--yes", "--json")
	doc = fits(t, out, "changes")
	plugins, _ := doc.Get("plugins")
	if code != 0 || doc.String("result") != "applied" || c.exists("bonsai.yaml") || c.exists(".bonsai/lock.json") || c.exists("demo/guide.md") ||
		!c.exists("demo/start.md") || strings.Join(f.asked, ",") != "list,uninstall demo-pack@"+market ||
		!strings.Contains(schema.Show(plugins), `"result":"uninstalled","message":"Claude Code's record of the install for this checkout removed","next":null`) {
		t.Errorf("unlink --yes: %d %v\n%s", code, f.asked, out)
	}
	// Again: nothing to change, nothing asked.
	f.asked = nil
	code, out, _ = c.run("", "unlink")
	if code != 0 || out != "bonsai unlink: nothing to change: no bonsai.yaml and no lock here, so this checkout is not linked.\n" || len(f.asked) != 0 {
		t.Errorf("a second unlink: %d %v\n%s", code, f.asked, out)
	}
	if code, out, _ := c.run("", "unlink", "--json"); code != 0 || fits(t, out, "changes").String("result") != "nothing" {
		t.Errorf("a second unlink --json: %d\n%s", code, out)
	}
	// init links again.
	if code, out, errOut := c.run("", append(c.linkArgs(c.pack.A), "--yes")...); code != 0 || !strings.Contains(out, "found        demo/start.md") {
		t.Errorf("init after unlink: %d\n%s%s", code, out, errOut)
	}
	// And unlink at a terminal, y: the text names what was removed and the plugin step.
	f.asked = nil
	code, out, _ = c.run("y\n", "unlink")
	if code != 0 || !strings.Contains(out, "bonsai unlink: done.\n") || !strings.Contains(out, "  removed      .bonsai/lock.json\n") ||
		!strings.Contains(out, "This machine's plugins (Claude Code, scope project: this checkout's .claude/settings.json):\n  uninstalled  demo-pack@"+market) ||
		!strings.Contains(out, "To link this project again: bonsai init") {
		t.Errorf("unlink, y: %d\n%s", code, out)
	}
}
