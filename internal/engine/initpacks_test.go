package engine

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Several packs at init (BUG-122): a first link of a pack with no roles and one with roles writes Bonsai's own lines
// in their roles' form on --yes alone (a first link's own lines), both packs in bonsai.yaml and the lock. A project
// linked to the first alone that adds the second later still needs --allow-exec for the changed guard line, and the
// preview says why: the pack brings roles, so the guard then judges agent calls too.
func TestInitSeveralPacks(t *testing.T) {
	e := setup(t)
	quiet, qc := testpack.QuietPack(t, e.tmp)
	wf, wc := protoPack(t, e.tmp, "workflow", workflowRoles)
	root := testpack.Project(t, e.tmp, "two")
	v := &InitValues{Name: "two", Source: quiet, Ref: qc[0], More: []InitPack{{Source: wf, Ref: wc[0]}}}
	p := e.plan(t, root, Request{Command: "init", Init: v})
	if p.NeedsExec() {
		t.Fatalf("a first link of two packs with no code needs --allow-exec: %s", items(p))
	}
	var own []string
	for _, c := range p.OwnHooks {
		own = append(own, c.Line)
	}
	if strings.Join(own, "\n") != strings.Join(OwnRolesHookLines(), "\n") {
		t.Errorf("own hooks %q", own)
	}
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}
	cfg, err := workspace.LoadConfig(root)
	if err != nil || len(cfg.Packs) != 2 || cfg.Packs[0].ID != "quiet-pack" || cfg.Packs[1].ID != "workflow" {
		t.Fatalf("bonsai.yaml's packs: %+v, %v", cfg, err)
	}
	if s := read(t, root, ".bonsai/lock.json"); !strings.Contains(s, `"workflow"`) || !strings.Contains(s, `"quiet-pack"`) {
		t.Errorf("the lock:\n%s", s)
	}
	if s := read(t, root, SettingsFile); !strings.Contains(s, jsonCmd(SecondCommand)) {
		t.Errorf("no second guard line:\n%s", s)
	}
	if p := e.plan(t, root, Request{Command: "init", Init: v}); len(p.Settings) != 0 || p.NeedsExec() {
		t.Errorf("init again changes settings: %+v", p.Settings)
	}

	// The same two, the second added later: --allow-exec, with the reason.
	later := testpack.Project(t, e.tmp, "later")
	e.apply(t, later, Request{Command: "init", Init: &InitValues{Name: "later", Source: quiet, Ref: qc[0]}})
	yaml := read(t, later, "bonsai.yaml")
	i := strings.Index(yaml, "documents:")
	writeFile(t, later, "bonsai.yaml", yaml[:i]+"  - id: workflow\n    source: \""+wf+"\"\n    path: null\n    ref: \""+wc[0]+"\"\n    roles: null\n"+yaml[i:])
	p = e.plan(t, later, Request{})
	if !p.NeedsExec() {
		t.Fatalf("adding a pack with roles later needs no --allow-exec: %s", items(p))
	}
	for _, c := range p.RunsCode {
		if c.Pack == "bonsai" && !strings.Contains(c.Why, "the pack workflow brings roles") {
			t.Errorf("Bonsai's line %s says %q", c.Item, c.Why)
		}
	}
}
