package status

// status --json's roles (set 8; step 5.5.0, note 10): each role the linked packs declare, its pack, kind and writes,
// and each limit's default and max in force with where each came from, held to bonsai.status/1's schema; read from the
// lock and bonsai.yaml as committed at the main checkout's HEAD, so a change counts once committed.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// rolesFixture is a pack declaring a session role and a lead, as workflow ships them.
const rolesFixture = `format: bonsai.pack/1
id: workflow
version: "0.1.0"
files: []
hooks: []
deny: []
roles:
  - name: orchestrator
    kind: lead
    session: true
    writes: true
    max_turns: null
    minutes:
      default: null
      max: null
    actions:
      default: null
      max: null
    red_runs:
      default: null
      max: null
  - name: verifier
    kind: lead
    session: false
    writes: false
    max_turns: 1600
    minutes:
      default: 120
      max: 120
    actions:
      default: 400
      max: 400
    red_runs:
      default: null
      max: null
`

func TestStatusRoles(t *testing.T) {
	tmp := t.TempDir()
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = r
	}
	home := testpack.Isolate(t, tmp)
	t.Setenv("BONSAI_TASK", "")
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	src, shas := testpack.Fixture(t, tmp, "workflow", map[string]string{
		".claude-plugin/plugin.json": "{\"name\": \"workflow\", \"description\": \"A fixture pack with roles.\"}\n",
		"bonsai/pack.yaml":           rolesFixture,
	})
	root := testpack.Project(t, tmp, "roles")
	p, err := engine.Build(engine.Request{Command: "init", Dir: root, Home: home, Version: "test",
		Init: &engine.InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Apply(p); err != nil {
		t.Fatal(err)
	}
	roles := func(t *testing.T) string {
		t.Helper()
		doc, code := Build(root, "dev")
		if code != ExitOK {
			t.Fatalf("exit %d: %s", code, schema.Show(doc))
		}
		checkShape(t, doc, false)
		v, _ := doc.Get("roles")
		return schema.Show(v)
	}
	limit := func(d, df, m, mf string) string {
		return `{"default":` + d + `,"default_from":"` + df + `","max":` + m + `,"max_from":"` + mf + `"}`
	}
	none := limit("null", "pack", "null", "pack")
	fromPack := `[{"pack":"workflow","name":"orchestrator","kind":"lead","writes":true,"minutes":` + none + `,"actions":` + none +
		`,"red_runs":` + none + `},{"pack":"workflow","name":"verifier","kind":"lead","writes":false,"minutes":` +
		limit("120", "pack", "120", "pack") + `,"actions":` + limit("400", "pack", "400", "pack") + `,"red_runs":` + none + `}]`
	// Not yet committed: the files as they stand.
	if got := roles(t); got != fromPack {
		t.Errorf("roles from the pack:\n%s\nwant\n%s", got, fromPack)
	}
	testpack.Git(t, root, "add", "-A")
	testpack.Git(t, root, "commit", "-q", "-m", "linked")
	// The project raises the verifier's minutes (default and max) and its max of actions.
	raw, err := os.ReadFile(filepath.Join(root, "bonsai.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	set := strings.Replace(string(raw), "    roles: null\n", "    roles:\n      verifier:\n        minutes:\n          default: 240\n"+
		"          max: 480\n        actions:\n          max: 600\n", 1)
	if err := os.WriteFile(filepath.Join(root, "bonsai.yaml"), []byte(set), 0o644); err != nil {
		t.Fatal(err)
	}
	// Committed at HEAD: an uncommitted change counts only once committed, as the guard reads it.
	if got := roles(t); got != fromPack {
		t.Errorf("an uncommitted change shows:\n%s", got)
	}
	testpack.Git(t, root, "commit", "-q", "-am", "the verifier's limits")
	want := strings.Replace(fromPack, `"minutes":`+limit("120", "pack", "120", "pack")+`,"actions":`+limit("400", "pack", "400", "pack"),
		`"minutes":`+limit("240", "bonsai.yaml", "480", "bonsai.yaml")+`,"actions":`+limit("400", "pack", "600", "bonsai.yaml"), 1)
	if got := roles(t); got != want {
		t.Errorf("roles once committed:\n%s\nwant\n%s", got, want)
	}
}
