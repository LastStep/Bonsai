package engine

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// never_read (BUG-121): each path becomes the deny rule Read(/<path>), named in the preview with its sentence and
// written on --yes alone (a deny rule runs no code); a path taken out loses its rule with the sentence that says so; a
// Read rule the project wrote itself stays whatever never_read does.
func TestNeverReadRules(t *testing.T) {
	e := setup(t)
	quiet, qc := testpack.QuietPack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "secret")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "secret", Source: quiet, Ref: qc[0]}})
	setNeverRead := func(list string) {
		t.Helper()
		cfg := read(t, root, "bonsai.yaml")
		line := "never_read: " + list
		if i := strings.Index(cfg, "\nnever_read: ["); i >= 0 {
			end := strings.Index(cfg[i+1:], "\n")
			cfg = cfg[:i+1] + line + cfg[i+1+end:]
		} else {
			cfg += line + "\n"
		}
		writeFile(t, root, "bonsai.yaml", cfg)
	}
	setNeverRead(`[".env", "secrets/**"]`)
	p := e.apply(t, root, Request{})
	if p.NeedsExec() || len(p.RunsCode) != 0 {
		t.Errorf("never_read's rules run code: %s", items(p))
	}
	var got []string
	for _, c := range p.Settings {
		got = append(got, c.Change+" "+c.Line+": "+c.Why)
	}
	want := "add Read(/.env): Agents can never read .env: it is in never_read in bonsai.yaml.; " +
		"add Read(/secrets/**): Agents can never read secrets/**: it is in never_read in bonsai.yaml."
	if strings.Join(got, "; ") != want {
		t.Errorf("settings lines\n  %s\nwant\n  %s", strings.Join(got, "; "), want)
	}
	deny, _ := getList(settingsDocOf(t, root), "permissions", "deny")
	if s := schema.Show(deny); !strings.Contains(s, `"Read(/.env)","Read(/secrets/**)"`) {
		t.Errorf("deny list %s", s)
	}
	// Linked again with nothing changed: nothing to write.
	if p := e.plan(t, root, Request{}); len(p.Settings) != 0 {
		t.Errorf("a second update changes settings: %+v", p.Settings)
	}
	// A path taken out, beside a Read rule of the project's own.
	s := read(t, root, SettingsFile)
	writeFile(t, root, SettingsFile, strings.Replace(s, `"Read(/secrets/**)"`, `"Read(/secrets/**)", "Read(mine.txt)"`, 1))
	setNeverRead(`["secrets/**"]`)
	p = e.apply(t, root, Request{})
	got = nil
	for _, c := range p.Settings {
		got = append(got, c.Change+" "+c.Line+": "+c.Why)
	}
	want = "remove Read(/.env): This file is no longer in never_read in bonsai.yaml, so agents may read it again."
	if strings.Join(got, "; ") != want {
		t.Errorf("settings lines\n  %s\nwant\n  %s", strings.Join(got, "; "), want)
	}
	if s := read(t, root, SettingsFile); !strings.Contains(s, `"Read(mine.txt)"`) || strings.Contains(s, `"Read(/.env)"`) {
		t.Errorf("settings after:\n%s", s)
	}
	// init --never-read: written into bonsai.yaml and as its rule on the first link; init again with other values is
	// refused as values-differ.
	other := testpack.Project(t, e.tmp, "secret-init")
	v := &InitValues{Name: "secret", Source: quiet, Ref: qc[0], NeverRead: []string{".env"}}
	e.apply(t, other, Request{Command: "init", Init: v})
	if cfg := read(t, other, "bonsai.yaml"); !strings.Contains(cfg, "\nnever_read: [\".env\"]\n") {
		t.Errorf("init's bonsai.yaml:\n%s", cfg)
	}
	if s := read(t, other, SettingsFile); !strings.Contains(s, `"Read(/.env)"`) {
		t.Errorf("init's settings:\n%s", s)
	}
	v.NeverRead = []string{"other.env"}
	if _, err := e.try(other, Request{Command: "init", Init: v}); err == nil || err.(*Error).Code != "values-differ" {
		t.Errorf("init with another never_read: %v", err)
	}
}
