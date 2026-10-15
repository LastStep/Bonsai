package engine

// A pack file's <protocols>/ start (step 5.5.0, note 1; protocols.go): a project whose folder is init's default, one
// whose folder is its own, a folder moved, and each refusal.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// protoPack is a fixture pack (its id given) that writes an always-on protocol file through <protocols>/
// (session-start.md, kind pack) and a plain one (proto/notes.md), and declares roles (YAML items, "" for roles: []).
// Its second commit changes only the protocol file.
func protoPack(t *testing.T, tmp, id, roles string) (string, []string) {
	t.Helper()
	list := " []\n"
	if roles != "" {
		list = "\n" + roles
	}
	yaml := "format: bonsai.pack/1\nid: " + id + "\nversion: \"0.5.0\"\nfiles:\n" +
		"  - path: \"<protocols>/" + id + "-start.md\"\n    from: start.md\n    kind: pack\n" +
		"  - path: " + id + "/notes.md\n    from: notes.md\n    kind: pack\n" +
		"hooks: []\ndeny: []\nroles:" + list
	return testpack.Fixture(t, tmp, id,
		map[string]string{
			".claude-plugin/plugin.json": pluginJSON(id),
			"bonsai/pack.yaml":           yaml,
			"bonsai/files/start.md":      "# Session start\n\nRead STATE first.\n",
			"bonsai/files/notes.md":      "# Notes\n",
		},
		map[string]string{"bonsai/files/start.md": "# Session start\n\nRead STATE first, then the task.\n"})
}

// protoConfig is the bonsai.yaml init would write for a pack, its documents.protocols line set to folder (the line
// left out for "-").
func protoConfig(t *testing.T, id, source, ref, folder string) string {
	t.Helper()
	raw, err := ConfigYAML("ws-aaaaaaaaaaaaaaaaaaaaaaaaaa", InitValues{Name: "demo", Source: source, Ref: ref},
		&PackData{Manifest: &workspace.Pack{ID: id}})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	i := strings.Index(s, "  protocols: work/protocols")
	if i < 0 {
		t.Fatalf("no protocols line in\n%s", s)
	}
	j := i + strings.Index(s[i:], "\n") + 1
	k := strings.LastIndex(s[:i], "\n  # ") + 1 // the comment line above it
	switch folder {
	case "-":
		return s[:k] + s[j:]
	default:
		return s[:i] + "  protocols: " + folder + "\n" + s[j:]
	}
}

// A project linked at init's default folder: the protocol file is written under work/protocols, the lock records the
// path it resolved to, the block imports it, and check finds nothing of its own.
func TestProtocolsAtTheDefaultFolder(t *testing.T) {
	e := setup(t)
	src, shas := protoPack(t, e.tmp, "proto", "")
	root := testpack.Project(t, e.tmp, "proto-default")
	p := e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	if got := read(t, root, "work/protocols/proto-start.md"); got != "# Session start\n\nRead STATE first.\n" {
		t.Errorf("the protocol file: %q", got)
	}
	for _, f := range p.Files {
		if strings.Contains(f.Path, "<") {
			t.Errorf("the plan names %s", f.Path)
		}
	}
	l, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if lf, ok := l.Files["work/protocols/proto-start.md"]; !ok || lf.Kind != "pack" || lf.Pack != "proto" {
		t.Errorf("the lock's entry: %+v (%v)", lf, ok)
	}
	for path := range l.Files {
		if strings.Contains(path, "<") {
			t.Errorf("the lock holds %s", path)
		}
	}
	if claude := read(t, root, "CLAUDE.md"); !strings.Contains(claude, "Always-on protocols, read in full: @work/protocols/proto-start.md\n") {
		t.Errorf("the block does not import the protocol file:\n%s", claude)
	}
	r, err := checkLocal(t, root, e.home)
	if err != nil || len(r.Findings) != 0 {
		t.Errorf("check: %v %+v", err, r.Findings)
	}
}

// A project whose protocols folder is its own (team/protocols here): the file goes there at a first link, and nothing
// under init's default.
func TestProtocolsAtTheProjectsFolder(t *testing.T) {
	e := setup(t)
	src, shas := protoPack(t, e.tmp, "proto", "")
	root := testpack.Project(t, e.tmp, "proto-own")
	writeFile(t, root, "bonsai.yaml", protoConfig(t, "proto", src, shas[0], "team/protocols"))
	e.apply(t, root, Request{Command: "init"})
	if got := read(t, root, "team/protocols/proto-start.md"); !strings.HasPrefix(got, "# Session start") {
		t.Errorf("the protocol file: %q", got)
	}
	if _, exists, _ := readFile(root, "work/protocols/proto-start.md"); exists {
		t.Error("a file under init's default folder")
	}
	if claude := read(t, root, "CLAUDE.md"); !strings.Contains(claude, "@team/protocols/proto-start.md") || strings.Contains(claude, "work/protocols") {
		t.Errorf("the block:\n%s", claude)
	}
	// The pack's next commit changes the file, at the project's folder.
	testpack.SetRef(t, root, shas[0], shas[1])
	e.apply(t, root, Request{})
	if got := read(t, root, "team/protocols/proto-start.md"); !strings.Contains(got, "then the task") {
		t.Errorf("after update: %q", got)
	}
}

// A project that moves its folder: at the next update the old file is handled as one the pack no longer has (removed
// when unedited, a conflict when edited) and the new one written; the lock and the block follow.
func TestProtocolsFolderMoved(t *testing.T) {
	e := setup(t)
	src, shas := protoPack(t, e.tmp, "proto", "")
	move := func(t *testing.T, root string) {
		t.Helper()
		raw := read(t, root, "bonsai.yaml")
		writeFile(t, root, "bonsai.yaml", strings.Replace(raw, "  protocols: work/protocols", "  protocols: docs/always", 1))
	}

	root := testpack.Project(t, e.tmp, "proto-moved")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	move(t, root)
	p := e.plan(t, root, Request{})
	results := map[string]string{}
	for _, f := range p.Files {
		results[f.Path] = f.Result
	}
	if results["work/protocols/proto-start.md"] != Removed || results["docs/always/proto-start.md"] != Created || results["proto/notes.md"] != Unchanged {
		t.Errorf("the plan for a moved folder: %v", results)
	}
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}
	if _, exists, _ := readFile(root, "work/protocols/proto-start.md"); exists {
		t.Error("the old file is still there")
	}
	if got := read(t, root, "docs/always/proto-start.md"); !strings.HasPrefix(got, "# Session start") {
		t.Errorf("the new file: %q", got)
	}
	l, _ := workspace.LoadLock(root)
	if _, ok := l.Files["work/protocols/proto-start.md"]; ok {
		t.Error("the lock still lists the old path")
	}
	if _, ok := l.Files["docs/always/proto-start.md"]; !ok {
		t.Error("the lock does not list the new path")
	}
	if claude := read(t, root, "CLAUDE.md"); !strings.Contains(claude, "@docs/always/proto-start.md") || strings.Contains(claude, "@work/protocols") {
		t.Errorf("the block:\n%s", claude)
	}

	// Edited before the move: the old file is a conflict, nothing is written, and --keep leaves it the project's.
	root = testpack.Project(t, e.tmp, "proto-moved-edited")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	writeFile(t, root, "work/protocols/proto-start.md", "# Session start\n\nOur own words.\n")
	move(t, root)
	p = e.plan(t, root, Request{})
	if len(p.Conflicts) != 1 || p.Conflicts[0].Path != "work/protocols/proto-start.md" {
		t.Fatalf("conflicts: %+v", p.Conflicts)
	}
	e.apply(t, root, Request{Keep: []string{"work/protocols/proto-start.md"}})
	if got := read(t, root, "work/protocols/proto-start.md"); !strings.Contains(got, "Our own words.") {
		t.Errorf("the kept file: %q", got)
	}
	if _, exists, _ := readFile(root, "docs/always/proto-start.md"); !exists {
		t.Error("the new file was not written")
	}
}

// Each refusal of a <protocols>/ file, at a first link and at a folder change alike: no folder (missing, or null), a
// folder that puts the file where Bonsai, Claude Code or git keeps its own, and two files of one pack made one by the
// folder. Each is bad-config, a person's, naming the line of bonsai.yaml to set, and writes nothing.
func TestProtocolsRefused(t *testing.T) {
	e := setup(t)
	src, shas := protoPack(t, e.tmp, "proto", "")
	n := 0
	refused := func(t *testing.T, root string, req Request, words string) {
		t.Helper()
		before := snapshot(t, root)
		_, err := e.try(root, req)
		wantRefused(t, err, "bad-config", ExitInput, words)
		ee := err.(*Error)
		if !strings.Contains(ee.Next, "documents.protocols in bonsai.yaml") || ee.Object().Next.Who != "person" {
			t.Errorf("next %q, who %s", ee.Next, ee.Object().Next.Who)
		}
		sameSnapshot(t, "a refusal", before, snapshot(t, root))
	}
	project := func(t *testing.T, folder string) string {
		t.Helper()
		n++
		root := testpack.Project(t, e.tmp, "proto-refused-"+strings.Repeat("x", n))
		writeFile(t, root, "bonsai.yaml", protoConfig(t, "proto", src, shas[0], folder))
		return root
	}
	t.Run("no folder", func(t *testing.T) {
		refused(t, project(t, "-"), Request{Command: "init"}, "names no documents.protocols folder")
		// null: bonsai.yaml's own reader refuses it first, naming the line.
		root := project(t, "null")
		_, err := e.try(root, Request{Command: "init"})
		wantRefused(t, err, "bad-config", ExitInput, `bonsai.yaml line 29: documents' "protocols" is null`)
	})
	for _, c := range []struct{ folder, words string }{
		{".claude", "is inside a .claude folder"},
		{".git", "is inside a .git folder"},
		{".bonsai", "is Bonsai's own file"},
		{"..", "has a . or .. segment"},
		{"CLAUDE~1", "short (8.3) name"},
		{"work/.claude/x", "is inside a .claude folder"},
	} {
		t.Run(c.folder, func(t *testing.T) {
			refused(t, project(t, `"`+c.folder+`"`), Request{Command: "init"}, c.words)
			// At a folder change: linked at the default, then moved there.
			root := project(t, "work/protocols")
			e.apply(t, root, Request{Command: "init"})
			raw := read(t, root, "bonsai.yaml")
			writeFile(t, root, "bonsai.yaml", strings.Replace(raw, "  protocols: work/protocols", `  protocols: "`+c.folder+`"`, 1))
			refused(t, root, Request{}, c.words)
		})
	}
	t.Run("two files made one", func(t *testing.T) {
		twoSrc, twoShas := testpack.Fixture(t, e.tmp, "twice", map[string]string{
			".claude-plugin/plugin.json": pluginJSON("twice"),
			"bonsai/pack.yaml": "format: bonsai.pack/1\nid: twice\nversion: \"1.0.0\"\nfiles:\n" +
				"  - path: \"<protocols>/a.md\"\n    from: a.md\n    kind: pack\n  - path: Work/Protocols/A.md\n    from: a.md\n    kind: pack\n" +
				"hooks: []\ndeny: []\n",
			"bonsai/files/a.md": "# A\n",
		})
		root := testpack.Project(t, e.tmp, "proto-twice")
		writeFile(t, root, "bonsai.yaml", protoConfig(t, "twice", twoSrc, twoShas[0], "work/protocols"))
		refused(t, root, Request{Command: "init"}, `the pack twice writes one path twice under documents.protocols "work/protocols": "<protocols>/a.md" and "Work/Protocols/A.md"`)
	})
}

// A documents.protocols that names a file of the project, or a path through one, is refused as bad-config before
// anything is written, at a first link and at update (V1's should-fix 3, 16 Oct: CLAUDE.md there left a link
// partly written and bonsai.yaml failed as read-failed).
func TestProtocolsFolderThatIsAFile(t *testing.T) {
	e := setup(t)
	src, shas := protoPack(t, e.tmp, "proto", "")
	n := 0
	refused := func(t *testing.T, root string, req Request, words string) {
		t.Helper()
		before := snapshot(t, root)
		_, err := e.try(root, req)
		wantRefused(t, err, "bad-config", ExitInput, words)
		ee := err.(*Error)
		if !strings.Contains(ee.Next, "documents.protocols in bonsai.yaml") || ee.Object().Next.Who != "person" {
			t.Errorf("next %q, who %s", ee.Next, ee.Object().Next.Who)
		}
		sameSnapshot(t, "a refusal", before, snapshot(t, root))
	}
	project := func(t *testing.T, folder string) string {
		t.Helper()
		n++
		root := testpack.Project(t, e.tmp, "proto-file-"+strings.Repeat("x", n))
		writeFile(t, root, "bonsai.yaml", protoConfig(t, "proto", src, shas[0], folder))
		writeFile(t, root, "CLAUDE.md", "# Demo\n")
		writeFile(t, root, "docs/notes.txt", "notes\n")
		return root
	}
	for _, c := range []struct{ folder, words string }{
		{"CLAUDE.md", `names "CLAUDE.md", a file of the project, not a folder`},
		{"bonsai.yaml", `names "bonsai.yaml", a file of the project, not a folder`},
		{"docs/notes.txt", `names "docs/notes.txt", a file of the project, not a folder`},
		{"docs/notes.txt/protocols", `names "docs/notes.txt/protocols", inside "docs/notes.txt", a file of the project`},
	} {
		t.Run(c.folder, func(t *testing.T) {
			refused(t, project(t, `"`+c.folder+`"`), Request{Command: "init"}, c.words)
			root := project(t, "work/protocols")
			e.apply(t, root, Request{Command: "init"})
			raw := read(t, root, "bonsai.yaml")
			writeFile(t, root, "bonsai.yaml", strings.Replace(raw, "  protocols: work/protocols", `  protocols: "`+c.folder+`"`, 1))
			refused(t, root, Request{}, c.words)
		})
	}
	// An existing folder is no file: linked there.
	root := project(t, "docs")
	e.apply(t, root, Request{Command: "init"})
	if got := read(t, root, "docs/proto-start.md"); !strings.HasPrefix(got, "# Session start") {
		t.Errorf("the protocol file: %q", got)
	}
}

// A documents.protocols that is, or runs through, a file Bonsai writes at the project's top (CLAUDE.md, bonsai.yaml),
// in any letter case, is refused as bad-config before anything is written, even where that file is not there yet: at a
// first link with no CLAUDE.md (V1's second round, proto-cases2.sh case r1-first: the link was left written part-way
// and CLAUDE.md made a folder), and at update with CLAUDE.md taken away.
func TestProtocolsFolderThatIsATopTarget(t *testing.T) {
	e := setup(t)
	src, shas := protoPack(t, e.tmp, "proto", "")
	n := 0
	refused := func(t *testing.T, root string, req Request, words string) {
		t.Helper()
		before := snapshot(t, root)
		_, err := e.try(root, req)
		wantRefused(t, err, "bad-config", ExitInput, words)
		ee := err.(*Error)
		if !strings.Contains(ee.Next, "documents.protocols in bonsai.yaml") || ee.Object().Next.Who != "person" {
			t.Errorf("next %q, who %s", ee.Next, ee.Object().Next.Who)
		}
		sameSnapshot(t, "a refusal", before, snapshot(t, root))
	}
	project := func(t *testing.T, folder string) string {
		t.Helper()
		n++
		root := testpack.Project(t, e.tmp, "proto-top-"+strings.Repeat("x", n))
		writeFile(t, root, "bonsai.yaml", protoConfig(t, "proto", src, shas[0], folder))
		return root
	}
	for _, c := range []struct{ folder, words string }{
		{"CLAUDE.md", `names "CLAUDE.md", which is "CLAUDE.md", a file Bonsai writes at the project's top`},
		{"Claude.md", `names "Claude.md", which is "CLAUDE.md", a file Bonsai writes at the project's top`},
		{"claude.md/protocols", `names "claude.md/protocols", inside "CLAUDE.md", a file Bonsai writes at the project's top`},
		// Where letter case is aside (Windows, macOS), bonsai.yaml is found as the file it is: refused either way.
		{"Bonsai.yaml", `names "Bonsai.yaml", `},
	} {
		t.Run(c.folder, func(t *testing.T) {
			// A first link, with no CLAUDE.md in the project yet.
			refused(t, project(t, `"`+c.folder+`"`), Request{Command: "init"}, c.words)
			// At update, CLAUDE.md taken away after the link.
			root := project(t, "work/protocols")
			e.apply(t, root, Request{Command: "init"})
			if err := os.Remove(filepath.Join(root, "CLAUDE.md")); err != nil {
				t.Fatal(err)
			}
			raw := read(t, root, "bonsai.yaml")
			writeFile(t, root, "bonsai.yaml", strings.Replace(raw, "  protocols: work/protocols", `  protocols: "`+c.folder+`"`, 1))
			refused(t, root, Request{}, c.words)
		})
	}
	// A folder of that name below the top is the project's own: linked there.
	root := project(t, "docs/CLAUDE.md")
	e.apply(t, root, Request{Command: "init"})
	if got := read(t, root, "docs/CLAUDE.md/proto-start.md"); !strings.HasPrefix(got, "# Session start") {
		t.Errorf("the protocol file: %q", got)
	}
}
