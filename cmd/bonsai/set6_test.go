package main

// Formats set 6's two additions to existing outputs (step 5.2.0), each checked against its schema through the
// command line: check --json's notes and unlink --json's left in place.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
)

func TestCheckJSONHasNotesAndUnlinkJSONHasLeft(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.commit("link")

	// A home with no install record: check says so as a note, which --json now carries.
	code, out, _ := c.run("", "check", "--json")
	doc := fits(t, out, "check")
	notes, ok := doc.Get("notes")
	list, _ := notes.([]any)
	if code > 1 || !ok || len(list) == 0 || !strings.Contains(list[0].(string), "install.json") {
		t.Errorf("check --json: %d, notes %v\n%s", code, notes, out)
	}
	if _, text, _ := c.run("", "check"); !strings.Contains(text, "note: "+list[0].(string)) {
		t.Errorf("the text output does not say the note the JSON carries:\n%s", text)
	}

	// unlink's preview lists what it leaves in place, the lines its text output prints under "Left in place".
	market := engine.MarketplaceName("demo", []string{c.pack.A})
	defer func(p engine.PluginCLI) { pluginCLI = p }(pluginCLI)
	pluginCLI = &fakePlugins{list: []engine.InstalledPlugin{{ID: "demo-pack@" + market, Version: c.pack.A[:12], Scope: "project", ProjectPath: c.root}}}
	_, text, _ := c.run("", "unlink")
	_, out, _ = c.run("", "unlink", "--json")
	doc = fits(t, out, "changes")
	left, ok := doc.Get("left")
	items, _ := left.([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("unlink --json has no left lines:\n%s", out)
	}
	for _, l := range items {
		if !strings.Contains(text, "  "+l.(string)+"\n") {
			t.Errorf("the left line %q is not in the text output:\n%s", l, text)
		}
	}
	// init and update leave nothing: left is [].
	if _, out, _ := c.run("", "update", "--json"); !strings.Contains(out, `"left": []`) {
		t.Errorf("update --json has no empty left:\n%s", out)
	}

	// status --json names both new formats, each read and written at major 1.
	_, out, _ = c.run("", "status", "--json")
	for _, name := range []string{"bonsai.asks", "bonsai.logs"} {
		if !strings.Contains(strings.Join(strings.Fields(out), ""), `"`+name+`":{"read":[1],"write":1}`) {
			t.Errorf("status --json's formats do not name %s at major 1:\n%s", name, out)
		}
	}
}
