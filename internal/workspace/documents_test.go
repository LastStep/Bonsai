package workspace

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// lockWith is a lock holding one pack, base, whose declares are d.
func lockWith(t *testing.T, d *format.Declares) *Lock {
	t.Helper()
	o, err := d.Object()
	if err != nil {
		t.Fatal(err)
	}
	return &Lock{Packs: []LockedPack{{ID: "base", Source: "s", Version: "1.0.0", Commit: strings.Repeat("0", 40),
		SHA256: strings.Repeat("1", 64), Declares: o, PathSet: true}}}
}

func strp(s string) *string { return &s }

// The id patterns kept as text here are their schemas' (formats/schemas, their one home).
func TestBonsaiKindsAreTheSchemas(t *testing.T) {
	for _, b := range bonsaiKinds {
		if b.Format == "" {
			continue
		}
		props, _ := format.MustLookup(b.Format).Schema().Get("properties")
		id, _ := props.(schema.Object).Get("id")
		want := ""
		if io, ok := id.(schema.Object); ok {
			if p, ok := io.Get("pattern"); ok {
				want = p.(string)
			}
		}
		if b.ID != want {
			t.Errorf("%s's id pattern %q, its schema's %q", b.Kind, b.ID, want)
		}
	}
	if strings.Join(Format0Kinds, " ") != "task run state" {
		t.Errorf("format-0 kinds %v", Format0Kinds)
	}
	for _, f := range format.All {
		if f.Read0 != (f.Name == "task" || f.Name == "run" || f.Name == "state") {
			t.Errorf("%s: Read0 %v and Format0Kinds disagree", f.Name, f.Read0)
		}
	}
}

// Bonsai's kinds at bonsai.yaml's places and their fixed files, then a pack's from its declares: at its default
// place, or at the one bonsai.yaml names under its kind's name.
func TestDocKinds(t *testing.T) {
	w := &format.Workspace{Documents: format.Documents{Task: "work/tasks", Run: "work/runs", Answers: "work/answers.md",
		Memory: "work/memory", Protocols: "work/protocols", Extra: schema.Object{{Key: "bugs", Value: "notes/bugs.md"}}}}
	lock := lockWith(t, &format.Declares{Documents: []format.PackDocument{
		{Kind: "plan", Path: strp("work/plans"), ID: strp(`^P-T-[0-9]{4,6}$`), Statuses: []string{"draft", "done"},
			Person: [][]string{{"draft", "done"}}, Stamp: schema.Object{{Key: "done", Value: "done"}}, TaskField: strp("task")},
		{Kind: "bugs", File: strp("work/bugs.md")},
	}})
	kinds, err := DocKinds(w, lock)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, k := range kinds {
		got = append(got, schema.Show(k.Object()))
	}
	want := []string{
		`{"kind":"task","from":"bonsai","path":"work/tasks","id":"^T-[0-9]{4,6}$","format":"bonsai.task","statuses":["todo","plan","approved","running","verify","done","blocked","cut"],"stamp":{"running":"started","done":"finished","cut":"finished"}}`,
		`{"kind":"run","from":"bonsai","path":"work/runs","id":"^R-[0-9]{4}-[0-9]{2}-[0-9]{2}-T-[0-9]{4,6}","format":"bonsai.run"}`,
		`{"kind":"state","from":"bonsai","file":".bonsai/STATE.md","format":"bonsai.state"}`,
		`{"kind":"answers","from":"bonsai","file":"work/answers.md"}`,
		`{"kind":"memory","from":"bonsai","path":"work/memory","id":"^M-[a-z0-9][a-z0-9-]*$","format":"bonsai.memory"}`,
		`{"kind":"tasks","from":"bonsai","file":".bonsai/tasks.md","format":"bonsai.tasks"}`,
		`{"kind":"sessions","from":"bonsai","file":".bonsai/sessions.md","format":"bonsai.sessions"}`,
		`{"kind":"plan","from":"base","path":"work/plans","id":"^P-T-[0-9]{4,6}$","statuses":["draft","done"],"person":[["draft","done"]],"stamp":{"done":"done"},"task_field":"task"}`,
		`{"kind":"bugs","from":"base","file":"notes/bugs.md"}`,
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("document kinds:\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	// Each entry fits status --json's documents item.
	props, _ := format.MustLookup("status").Schema().Get("properties")
	docs, _ := props.(schema.Object).Get("documents")
	items, _ := docs.(schema.Object).Get("items")
	for _, k := range kinds {
		if msgs := schema.Validate(items.(schema.Object), k.Object()); len(msgs) > 0 {
			t.Errorf("%s: %v", k.Kind, msgs)
		}
	}
	for _, name := range []string{"task", "run", "state", "answers", "memory", "tasks", "sessions"} {
		if !BonsaiKind(name) {
			t.Errorf("%s is not Bonsai's", name)
		}
	}
	if BonsaiKind("plan") || BonsaiKind("protocols") {
		t.Errorf("plan or protocols taken for Bonsai's kind")
	}
}

// A folder kind's documents: top-level .md files named <id>.md or <id>-<slug>.md by the kind's id pattern.
func TestDocFiles(t *testing.T) {
	root := t.TempDir()
	for _, f := range []string{"work/tasks/T-0901-first.md", "work/tasks/T-0902.md", "work/tasks/README.md",
		"work/tasks/T-09-short.md", "work/tasks/notes.txt", "work/tasks/.T-0903-hidden.md", "work/tasks/sub/T-0904-deep.md",
		"work/tasks/T-0905-x.md.bak", ".bonsai/STATE.md"} {
		p := filepath.Join(root, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := DocFiles(root, DocKind{Kind: "task", Path: "work/tasks", ID: TaskIDPattern})
	if err != nil || strings.Join(got, " ") != "work/tasks/T-0901-first.md work/tasks/T-0902.md" {
		t.Errorf("task files %v %v", got, err)
	}
	if got, err := DocFiles(root, DocKind{Kind: "notes", Path: "work/tasks"}); err != nil || len(got) != 4 {
		t.Errorf("a kind with no id pattern: %v %v", got, err)
	}
	if got, _ := DocFiles(root, DocKind{Kind: "state", File: StateFile}); strings.Join(got, "") != StateFile {
		t.Errorf("a file kind: %v", got)
	}
	if got, err := DocFiles(root, DocKind{Kind: "run", Path: "work/runs", ID: runIDPattern}); err != nil || got != nil {
		t.Errorf("a folder that is not there: %v %v", got, err)
	}
	run := regexp.MustCompile(runIDPattern)
	for name, want := range map[string]string{"R-2026-10-09-T-0901.md": "R-2026-10-09-T-0901",
		"R-2026-10-09-T-0901-2.md": "R-2026-10-09-T-0901", "R-2026-10-09.md": "", "T-0901-x.md": ""} {
		if got := NameID(name, run); got != want {
			t.Errorf("NameID(%s) = %q, want %q", name, got, want)
		}
	}
}

// Lanes in force: the packs' from the lock's declares, each with its pack.
func TestLanes(t *testing.T) {
	lock := lockWith(t, &format.Declares{Lanes: []format.Lane{{Name: "light", Close: "agent", Description: "x"},
		{Name: "full", ApproveFirst: true, Close: "person", Description: "y"}}})
	l, err := Lanes(lock)
	if err != nil || len(l) != 2 || l[1].Name != "full" || !l[1].ApproveFirst || l[1].From != "base" {
		t.Errorf("lanes %+v %v", l, err)
	}
	if l, err := Lanes(nil); err != nil || len(l) != 0 {
		t.Errorf("no lock: %+v %v", l, err)
	}
}
