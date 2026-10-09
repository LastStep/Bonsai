package engine

// The instruction block complete (spec §6): the workspace line, the import of the always-on protocol files, the
// import of the project's memory index, the label definitions agents see, the packs' block.md texts; at most 40 lines.

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
)

// baseYAML is a bonsai.yaml linking the declaring fixture pack, its documents in work/.
func baseYAML(source, ref string) string {
	return "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\npacks:\n  - id: base\n    source: \"" +
		filepath.ToSlash(source) + "\"\n    ref: \"" + ref + "\"\ndocuments:\n  task: work/tasks\n  run: work/runs\n" +
		"  answers: work/answers.md\n  memory: work/memory\n  protocols: work/protocols\n"
}

func TestBlockComplete(t *testing.T) {
	e := setup(t)
	src, shas := testpack.DeclaringPack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "block")
	writeFile(t, root, "bonsai.yaml", baseYAML(src, shas[0]))
	writeFile(t, root, "CLAUDE.md", "# My project\n\nOur own words.\n")
	e.apply(t, root, Request{Command: "init"})
	got := read(t, root, "CLAUDE.md")
	want := "# My project\n\nOur own words.\n\n" + blockStart + "\n" +
		"This project is linked to Bonsai as workspace demo; its packs: base 1.0.0.\n" +
		"Always-on protocols, read in full: @work/protocols/session-start.md\n" +
		"Project memory, an index of notes kept in work/memory: @work/memory/INDEX.md\n" +
		"Labels a document may carry under its labels: field (name: the value it takes; the documents it goes on; who sets it):\n" +
		"- bonsai.allows: a list of text; on task; agents set it; it grants a right. Protected paths this task may change, as globs.\n" +
		"- bonsai.branch: text matching ^[A-Za-z0-9._/-]{1,100}$; on task; agents set it. The branch the task is built on; empty means the base branch.\n" +
		"The base fixture is linked.\n" + blockEnd + "\n"
	if got != want {
		t.Errorf("CLAUDE.md:\n%s\nwant\n%s", got, want)
	}
	if n := strings.Count(got[strings.Index(got, blockStart):], "\n"); n > MaxBlockLines {
		t.Errorf("the block is %d lines", n)
	}
	// An outside label says agents never write it; a choice names its values; a number and a capped text read plainly.
	for _, c := range []struct{ yaml, want string }{
		{"kind: choice\n    values: [\"a\", \"b\"]\n    items: null\n    pattern: null\n    max: null\n    kinds: [\"task\", \"run\"]\n    set_by: outside",
			"- x.y: one of a, b; on task, run; set outside agent sessions: never write it. D."},
		{"kind: number\n    values: []\n    items: null\n    pattern: null\n    max: null\n    kinds: [\"run\"]\n    set_by: agent",
			"- x.y: a number; on run; agents set it. D."},
		{"kind: text\n    values: []\n    items: null\n    pattern: null\n    max: 60\n    kinds: [\"task\"]\n    set_by: agent",
			"- x.y: text, at most 60 characters; on task; agents set it. D."},
	} {
		raw := "format: bonsai.labels/1\nnamespace: x\nversion: 1\nlabels:\n  - name: x.y\n    " + c.yaml + "\n    grants: false\n    description: \"D.\"\n"
		l, err := formatLabels(raw)
		if err != nil {
			t.Fatal(err)
		}
		if got := labelLine(l.Labels[0]); got != c.want {
			t.Errorf("label line %q, want %q", got, c.want)
		}
	}
}

// A block over 40 lines is refused, naming the packs' makers' step; nothing is written.
func TestBlockOverItsCap(t *testing.T) {
	e := setup(t)
	long := strings.Repeat("A line of the pack's block.\n", MaxBlockLines-2)
	src, shas := testpack.Fixture(t, e.tmp, "long-block", map[string]string{
		".claude-plugin/plugin.json": pluginJSON("long-block"),
		"bonsai/pack.yaml":           "format: bonsai.pack/1\nid: long-block\nversion: \"1.0.0\"\nblock: block.md\nfiles: []\nhooks: []\ndeny: []\n",
		"bonsai/block.md":            long,
	})
	root := testpack.Project(t, e.tmp, "long")
	before := snapshot(t, root)
	_, err := e.try(root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	var ee *Error
	if !errors.As(err, &ee) || ee.Code != "bad-pack" || ee.Exit != ExitInput || !strings.Contains(ee.What, "would be 41 lines, over its fixed 40") {
		t.Fatalf("a block of 41 lines: %v", err)
	}
	sameSnapshot(t, "a refused block", before, snapshot(t, root))
	// 40 lines is within the cap.
	src2, shas2 := testpack.Fixture(t, e.tmp, "full-block", map[string]string{
		".claude-plugin/plugin.json": pluginJSON("full-block"),
		"bonsai/pack.yaml":           "format: bonsai.pack/1\nid: full-block\nversion: \"1.0.0\"\nblock: block.md\nfiles: []\nhooks: []\ndeny: []\n",
		"bonsai/block.md":            strings.Repeat("A line of the pack's block.\n", MaxBlockLines-3),
	})
	root2 := testpack.Project(t, e.tmp, "full")
	e.apply(t, root2, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src2, Ref: shas2[0]}})
}
