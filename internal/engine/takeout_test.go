package engine

// Taking a pack out of a project (plan-5, step 5.1.7): a pack the lock holds and bonsai.yaml no longer lists is taken
// out by update. Its files nobody edited go, an edited one and its once file stay (released, named); its part of the
// block, its hook line, deny rule and plugin wiring go; its lock entry and declares go; --yes is enough, since a
// removed hook line runs nothing; and check is clean afterwards.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// twoPacks is a bonsai.yaml naming the test pack and the side pack.
func twoPacks(e *env, side, sideRef string) string {
	return "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\npacks:\n" +
		"  - id: demo-pack\n    source: \"" + filepath.ToSlash(e.pack.Source) + "\"\n    ref: \"" + e.pack.A + "\"\n" +
		"  - id: side-pack\n    source: \"" + filepath.ToSlash(side) + "\"\n    ref: \"" + sideRef + "\"\n"
}

func TestTakeAPackOut(t *testing.T) {
	e := setup(t)
	side, sc := testpack.SidePack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "two")
	writeFile(t, root, "bonsai.yaml", twoPacks(e, side, sc[0]))
	e.apply(t, root, Request{Command: "init", Init: &InitValues{}, AllowExec: true})
	if !strings.Contains(read(t, root, SettingsFile), "echo side hook") || !strings.Contains(read(t, root, "CLAUDE.md"), "The side pack is linked.") {
		t.Fatalf("the side pack is not linked:\n%s", read(t, root, SettingsFile))
	}
	oldMarket := MarketplaceName("demo", []string{e.pack.A, sc[0]})
	writeFile(t, root, "side/b.md", "# Side B, edited here\n")

	// bonsai.yaml without the side pack: check names update as the step; update takes it out.
	cfg := read(t, root, "bonsai.yaml")
	writeFile(t, root, "bonsai.yaml", cfg[:strings.Index(cfg, "  - id: side-pack")])
	r, err := checkLocal(t, root, e.home)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range r.Findings {
		if f.Code == "packs" && strings.Contains(f.Message, "the lock holds the pack side-pack") {
			found = f.Next == "run: bonsai update --yes" && f.Who == "person"
		}
	}
	if !found {
		t.Errorf("check's packs finding: %+v", r.Findings)
	}

	p := e.plan(t, root, Request{})
	if len(p.Removed) != 1 || p.Removed[0].ID != "side-pack" || p.OldMarket != oldMarket || p.NeedsExec() || len(p.Conflicts) != 0 {
		t.Fatalf("the plan: removed %+v, old market %s, runs code %v, conflicts %v", p.Removed, p.OldMarket, p.RunsCode, p.Conflicts)
	}
	for path, want := range map[string]string{"side/a.md": Removed, "side/b.md": Released, "side/once.md": Released,
		"demo/guide.md": Unchanged, "CLAUDE.md": Updated, SettingsFile: Updated} {
		if got := result(t, p, path); got.Result != want {
			t.Errorf("%s: %s (%s), want %s", path, got.Result, got.Why, want)
		}
	}
	var lines []string
	for _, c := range p.Settings {
		lines = append(lines, c.Change+" "+c.Kind+" "+c.Line)
		if c.RunsCode {
			t.Errorf("a settings line runs code: %+v", c)
		}
	}
	got := strings.Join(lines, "\n")
	for _, want := range []string{"remove deny Read(side/secret.txt)", "remove hook SessionStart (startup): echo side hook",
		"remove plugin side-pack@" + oldMarket + " enabled", "change marketplace bonsai-demo-", "change plugin demo-pack@bonsai-demo-"} {
		if !strings.Contains(got, want) {
			t.Errorf("the settings lines lack %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "demo hook A") || strings.Contains(got, "demo/never.txt") {
		t.Errorf("the remaining pack's lines changed:\n%s", got)
	}
	preview := p.Preview(false)
	for _, want := range []string{"  side-pack 0.2.0  " + sc[0][:7] + " -> taken out  (", "Taken out of bonsai.yaml: side-pack. Its files nobody edited go (side/a.md);",
		"its lock entry and declares go;", "Left in place, the project's now (edited here, or written once): side/b.md, side/once.md",
		"released     side/b.md [pack]: its pack is taken out; you edited it, so it stays"} {
		if !strings.Contains(preview, want) {
			t.Errorf("the preview lacks %q:\n%s", want, preview)
		}
	}
	doc, err := p.Changes("preview", nil).Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(doc), `"id": "side-pack",`) || !strings.Contains(string(doc), `"to": null`) {
		t.Errorf("the changes output:\n%s", doc)
	}
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}

	// What is left: the edited and once files; no line, part of the block or lock entry of the side pack.
	files := snapshot(t, root)
	for _, gone := range []string{"side/a.md"} {
		if _, ok := files[gone]; ok {
			t.Errorf("%s is still there", gone)
		}
	}
	for _, kept := range []string{"side/b.md", "side/once.md", "demo/guide.md", "demo/start.md"} {
		if _, ok := files[kept]; !ok {
			t.Errorf("%s is gone", kept)
		}
	}
	settings := read(t, root, SettingsFile)
	if strings.Contains(settings, "side") || !strings.Contains(settings, "echo demo hook A") || !strings.Contains(settings, "Edit(demo/never.txt)") {
		t.Errorf("settings:\n%s", settings)
	}
	if strings.Contains(read(t, root, "CLAUDE.md"), "side") {
		t.Errorf("CLAUDE.md:\n%s", read(t, root, "CLAUDE.md"))
	}
	lock, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Packs) != 1 || lock.Packs[0].ID != "demo-pack" {
		t.Errorf("lock packs %+v", lock.Packs)
	}
	for path := range lock.Files {
		if strings.HasPrefix(path, "side/") {
			t.Errorf("the lock still lists %s", path)
		}
	}
	if r, err := checkLocal(t, root, e.home); err != nil || len(r.Findings) != 0 {
		t.Errorf("check after: %v %+v", err, r.Findings)
	}
	// Again: nothing to change.
	if p := e.plan(t, root, Request{}); !p.Nothing() {
		t.Errorf("a second update:\n%s", p.Preview(false))
	}
}

// Taking a pack out while the same run changes code (the remaining pack's hook line, C to D) needs --allow-exec, as
// any run that adds or changes code does; the removal alone does not.
func TestTakeAPackOutWithCode(t *testing.T) {
	e := setup(t)
	side, sc := testpack.SidePack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "two-code")
	writeFile(t, root, "bonsai.yaml", strings.Replace(twoPacks(e, side, sc[0]), e.pack.A, e.pack.C, 1))
	e.apply(t, root, Request{Command: "init", Init: &InitValues{}, AllowExec: true})
	cfg := read(t, root, "bonsai.yaml")
	writeFile(t, root, "bonsai.yaml", strings.Replace(cfg[:strings.Index(cfg, "  - id: side-pack")], e.pack.C, e.pack.D, 1))
	p := e.plan(t, root, Request{})
	if !p.NeedsExec() || len(p.RunsCode) != 1 || p.RunsCode[0].Pack != "demo-pack" || len(p.Removed) != 1 {
		t.Errorf("runs code %+v, removed %+v", p.RunsCode, p.Removed)
	}
	if err := Apply(p); err == nil {
		t.Errorf("applied without --allow-exec")
	}
	e.apply(t, root, Request{AllowExec: true})
	if s := read(t, root, SettingsFile); strings.Contains(s, "side") || !strings.Contains(s, "echo demo hook D") {
		t.Errorf("settings:\n%s", s)
	}
}

// The changes output of a pack taken out fits bonsai.changes/1: to null, the files released and removed.
func TestTakeOutChangesFit(t *testing.T) {
	e := setup(t)
	side, sc := testpack.SidePack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "fit")
	writeFile(t, root, "bonsai.yaml", twoPacks(e, side, sc[0]))
	e.apply(t, root, Request{Command: "init", Init: &InitValues{}, AllowExec: true})
	cfg := read(t, root, "bonsai.yaml")
	writeFile(t, root, "bonsai.yaml", cfg[:strings.Index(cfg, "  - id: side-pack")])
	p := e.plan(t, root, Request{})
	raw, err := p.Changes("preview", &Error{Code: "needs-yes", Exit: ExitState, What: "no --yes", Next: "run: bonsai update --yes"}).Encode()
	if err != nil {
		t.Fatal(err)
	}
	v, err := schema.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if msgs := schema.Validate(format.MustLookup("changes").Schema(), v); len(msgs) > 0 {
		t.Errorf("does not fit bonsai.changes/1: %v\n%s", msgs, raw)
	}
	packs, _ := v.(schema.Object).Get("packs")
	side0, _ := packs.([]any)[1].(schema.Object).Get("to")
	if list := packs.([]any); len(list) != 2 || side0 != nil {
		t.Errorf("packs %s", schema.Show(packs))
	}
}
