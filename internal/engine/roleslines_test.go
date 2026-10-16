package engine

// The guard's two lines where the lock declares roles (plan-5 5.5.11 note 2): a first link of a pack with roles writes
// them on --yes alone, as Bonsai's own lines; check holds a project to the form its lock calls for (own-hooks); and an
// update that adds a project's first pack with roles changes the guard's line and adds the second, both under Runs
// code, so --yes alone is refused and --allow-exec writes them; taking the roles out again goes back to the one line.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
)

func TestTheGuardsTwoLines(t *testing.T) {
	e := setup(t)
	src, shas := protoPack(t, e.tmp, "workflow", workflowRoles)

	// A first link: the roles' own lines, on --yes alone.
	root := testpack.Project(t, e.tmp, "roles")
	p := e.linkTo(t, root, "demo", src, shas[0], false)
	var own []string
	for _, c := range p.OwnHooks {
		own = append(own, c.Line)
	}
	if strings.Join(own, "\n") != strings.Join(OwnRolesHookLines(), "\n") || len(own) != len(OwnHookLines())+1 || p.NeedsExec() {
		t.Fatalf("own hooks %+v, runs code %+v", p.OwnHooks, p.RunsCode)
	}
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}
	s := read(t, root, SettingsFile)
	for _, want := range []string{jsonCmd(GuardCommand), jsonCmd(SecondCommand), `"matcher": "` + GuardRolesMatcher + `"`,
		`"matcher": "` + SecondMatcher + `"`} {
		if !strings.Contains(s, want) {
			t.Errorf("the settings file has no %s:\n%s", want, s)
		}
	}
	if f := e.settingsFindings(t, root); f != "" {
		t.Errorf("check after the link: %s", f)
	}

	// A project linked to a pack with no roles, whose next commit declares them: update adds the second line and
	// changes the guard's, each under Runs code.
	plainSrc, plainShas := testpack.Fixture(t, e.tmp, "growing", map[string]string{
		".claude-plugin/plugin.json": pluginJSON("growing"),
		"bonsai/pack.yaml":           "format: bonsai.pack/1\nid: growing\nversion: \"0.5.0\"\nfiles: []\nhooks: []\ndeny: []\nroles: []\n",
	}, map[string]string{
		"bonsai/pack.yaml": "format: bonsai.pack/1\nid: growing\nversion: \"0.6.0\"\nfiles: []\nhooks: []\ndeny: []\nroles:\n" + workflowRoles,
	})
	two := testpack.Project(t, e.tmp, "growing")
	e.apply(t, two, Request{Command: "init", Init: &InitValues{Name: "demo", Source: plainSrc, Ref: plainShas[0]}})
	if s := read(t, two, SettingsFile); strings.Contains(s, jsonCmd(SecondCommand)) || !strings.Contains(s, `"matcher": "`+GuardMatcher+`"`) {
		t.Fatalf("a pack with no roles wrote the roles' lines:\n%s", s)
	}
	if f := e.settingsFindings(t, two); f != "" {
		t.Errorf("check with no roles: %s", f)
	}
	before := read(t, two, SettingsFile)
	testpack.SetRef(t, two, plainShas[0], plainShas[1])
	p = e.plan(t, two, Request{})
	got := items(p)
	if !p.NeedsExec() || !strings.Contains(got, "hook add "+ownRolesHooks[0].Text()+" (bonsai)") ||
		!strings.Contains(got, "hook add "+ownRolesHooks[1].Text()+" (bonsai)") {
		t.Fatalf("runs code: %s", got)
	}
	if pv := p.Preview(false); !strings.Contains(flat(pv), flat(ownRolesHooks[1].Text())) {
		t.Errorf("the preview does not name the second line:\n%s", pv)
	}
	if err := Apply(p); err == nil {
		t.Fatal("update --yes wrote the guard's two lines without --allow-exec")
	}
	if read(t, two, SettingsFile) != before {
		t.Error("a refused update wrote the settings file")
	}
	e.apply(t, two, Request{AllowExec: true})
	if s := read(t, two, SettingsFile); !strings.Contains(s, jsonCmd(SecondCommand)) || strings.Contains(s, `"matcher": "`+GuardMatcher+`"`) {
		t.Errorf("the update did not write the roles' lines:\n%s", s)
	}
	if f := e.settingsFindings(t, two); f != "" {
		t.Errorf("check after the update: %s", f)
	}

	// The second line taken out by hand: check names it to add.
	s = read(t, two, SettingsFile)
	cut := strings.Replace(s, jsonCmd(SecondCommand), `"echo not bonsai"`, 1)
	writeFile(t, two, SettingsFile, cut)
	if f := e.settingsFindings(t, two); !strings.Contains(f, "changed") {
		t.Errorf("check does not see the second line gone: %s", f)
	}
	writeFile(t, two, SettingsFile, s)

	// Back to a commit with no roles: the one line again, the second taken out.
	testpack.SetRef(t, two, plainShas[1], plainShas[0])
	e.apply(t, two, Request{AllowExec: true})
	if s := read(t, two, SettingsFile); strings.Contains(s, jsonCmd(SecondCommand)) || !strings.Contains(s, `"matcher": "`+GuardMatcher+`"`) {
		t.Errorf("taking the roles out left the roles' lines:\n%s", s)
	}
}

// settingsFindings runs Check on a project with the test's home and gives its findings and warnings on the settings
// file, "" for none.
func (e *env) settingsFindings(t *testing.T, root string) string {
	t.Helper()
	r, err := Check(root, e.home)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, f := range append(r.Findings, r.Warnings...) {
		if f.File == SettingsFile {
			out = append(out, f.Code+": "+f.Message)
		}
	}
	return strings.Join(out, "; ")
}
