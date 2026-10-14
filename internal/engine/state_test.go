package engine

// STATE at init (step 5.5.0, note 2; state.go): written in format 1 when there is none, kind once in the lock, named
// in the preview; never written over, at init or update; left missing by update; left by unlink; a project linked
// before step 5.5 gets none at update.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestStateAtInit(t *testing.T) {
	e := setup(t)
	was := today
	today = func() string { return "2026-10-15" }
	t.Cleanup(func() { today = was })

	root := testpack.Project(t, e.tmp, "state")
	p := e.plan(t, root, Request{Command: "init", Init: e.values(e.pack.A), AllowExec: true})
	if !strings.Contains(p.Preview(false), "  created      .bonsai/STATE.md [once]: where the project stands, from Bonsai's template") {
		t.Errorf("the preview does not name STATE:\n%s", p.Preview(false))
	}
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}
	raw := read(t, root, workspace.StateFile)
	if raw != string(StateTemplate("2026-10-15")) {
		t.Errorf("STATE:\n%s", raw)
	}
	s, err := format.ReadState([]byte(raw))
	if err != nil || s.Format0 != nil || s.Updated != "2026-10-15" || s.UpdatedBy != "bonsai init" || len(s.Labels) != 0 || s.Labels == nil {
		t.Errorf("STATE reads as %+v, %v", s, err)
	}
	if !strings.HasPrefix(raw, "---\nformat: bonsai.state/1   # fields: bonsai check --schema bonsai.state; how to fill: skill base:state\n") {
		t.Errorf("the format line has no pointer:\n%s", raw)
	}
	l, _ := workspace.LoadLock(root)
	if lf := l.Files[workspace.StateFile]; lf.Kind != "once" || lf.Pack != testpack.ID || lf.SHA256 != workspace.HashLF([]byte(raw)) {
		t.Errorf("the lock's entry %+v", lf)
	}
	if len(l.Format0) != 0 {
		t.Errorf("format0 %v", l.Format0)
	}
	if r, err := checkLocal(t, root, e.home); err != nil || len(r.Findings) != 0 {
		t.Errorf("check: %v %+v", err, r.Findings)
	}

	// Never written over: edited, then update to another commit and init again.
	writeFile(t, root, workspace.StateFile, strings.Replace(raw, "## Now\n", "## Now\n\nOur own words.\n", 1))
	edited := read(t, root, workspace.StateFile)
	testpack.SetRef(t, root, e.pack.A, e.pack.B)
	p = e.apply(t, root, Request{})
	if got := result(t, p, workspace.StateFile); got.Result != Unchanged || got.Writes() {
		t.Errorf("update: %+v", got)
	}
	e.apply(t, root, Request{Command: "init", AllowExec: true})
	if read(t, root, workspace.StateFile) != edited {
		t.Error("an edited STATE was written over")
	}
	// Deleted: update leaves it missing, its lock entry kept; init writes a new one.
	if err := os.Remove(filepath.Join(root, ".bonsai", "STATE.md")); err != nil {
		t.Fatal(err)
	}
	p = e.apply(t, root, Request{})
	if got := result(t, p, workspace.StateFile); got.Result != LeftMissing || got.Writes() {
		t.Errorf("update with STATE deleted: %+v", got)
	}
	if _, exists, _ := readFile(root, workspace.StateFile); exists {
		t.Error("update wrote STATE")
	}
	if l, _ := workspace.LoadLock(root); l.Files[workspace.StateFile].Kind != "once" {
		t.Error("the lock lost STATE's entry")
	}
	e.apply(t, root, Request{Command: "init", AllowExec: true})
	if read(t, root, workspace.StateFile) != raw {
		t.Error("init did not write a new STATE")
	}
	// unlink leaves it, and names it.
	up := unlinkPlan(t, e, root)
	if !strings.Contains(up.Preview(false), "Left in place:\n  .bonsai/STATE.md: the project's STATE\n") || up.lists(workspace.StateFile) {
		t.Errorf("unlink's preview:\n%s", up.Preview(false))
	}
	if err := Apply(up); err != nil {
		t.Fatal(err)
	}
	if read(t, root, workspace.StateFile) != raw {
		t.Error("unlink took STATE")
	}
}

// A STATE already in the project when it is first linked is found: left as it is, format 0 among them (then the lock's
// format0 lists it too), and locked as found.
func TestStateFoundAtInit(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "state-found")
	own := "---\nupdated: 2026-10-01\n---\n\n# Our state\n"
	writeFile(t, root, workspace.StateFile, own)
	p := e.link(t, root, e.pack.A)
	if got := result(t, p, workspace.StateFile); got.Result != Found || got.Writes() {
		t.Errorf("init: %+v", got)
	}
	if read(t, root, workspace.StateFile) != own {
		t.Error("init wrote over the project's STATE")
	}
	l, _ := workspace.LoadLock(root)
	if l.Files[workspace.StateFile].SHA256 != workspace.HashLF([]byte(own)) || l.Format0[workspace.StateFile] == "" {
		t.Errorf("the lock: %+v, format0 %v", l.Files[workspace.StateFile], l.Format0)
	}
}

// A project linked before step 5.5 (its lock has no STATE): update writes none and locks none, so the plan says
// nothing of it; base's state skill is how such a project adds one.
func TestStateNotAddedByUpdate(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "state-before")
	e.link(t, root, e.pack.A)
	l, _ := workspace.LoadLock(root)
	delete(l.Files, workspace.StateFile)
	if err := workspace.WriteLock(root, l); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, ".bonsai", "STATE.md")); err != nil {
		t.Fatal(err)
	}
	p := e.apply(t, root, Request{})
	if p.lists(workspace.StateFile) {
		t.Errorf("the plan names STATE:\n%s", p.Preview(false))
	}
	if _, exists, _ := readFile(root, workspace.StateFile); exists {
		t.Error("update wrote STATE")
	}
	if l, _ := workspace.LoadLock(root); l.Files[workspace.StateFile].Kind != "" {
		t.Error("update locked a STATE")
	}
}

// STATE's template documents its fields once (StateFields, held to bonsai.state/1's schema both ways) and is read as
// a format-1 STATE by Bonsai's own reader.
func TestStateTemplate(t *testing.T) {
	props, _ := format.MustLookup("state").Schema().Get("properties")
	var fields []string
	for _, f := range StateFields {
		if f.Meaning == "" {
			t.Errorf("%s has no meaning", f.Field)
		}
		fields = append(fields, f.Field)
	}
	want := props.(schema.Object).Keys()
	if strings.Join(fields, " ") != strings.Join(want, " ") {
		t.Errorf("StateFields %v, want the schema's %v", fields, want)
	}
	if _, err := format.ReadState(StateTemplate("2026-01-02")); err != nil {
		t.Errorf("the template does not read: %v", err)
	}
	if strings.ContainsAny(string(StateTemplate("2026-01-02")), "`@[") {
		t.Error("the template names a path: check's missing-path finding would read it")
	}
}
