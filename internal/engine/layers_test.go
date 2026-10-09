package engine

// Layers and the old workspace (spec §6): packs apply in bonsai.yaml's order; a path two packs write stops the
// command (letter case aside, as Windows sees it), at a first link and at an update; and a Bonsai 0.4.3 workspace
// (.bonsai.yaml or .bonsai-lock.yaml) is refused by init and update alike. Each refusal writes nothing.

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// layerPack is a pack with a block line and one pack file at path.
func layerPack(t *testing.T, tmp, id, path string) (string, string) {
	t.Helper()
	src, shas := testpack.Fixture(t, tmp, id, map[string]string{
		".claude-plugin/plugin.json": pluginJSON(id),
		"bonsai/pack.yaml": "format: bonsai.pack/1\nid: " + id + "\nversion: \"1.0.0\"\nblock: block.md\nfiles:\n  - path: " + path +
			"\n    from: f.md\n    kind: pack\nhooks: []\ndeny: []\n",
		"bonsai/block.md":   "The " + id + " pack's line.\n",
		"bonsai/files/f.md": "# " + id + "\n",
	})
	return src, shas[0]
}

func packsYAML(packs ...[3]string) string {
	s := "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\npacks:\n"
	for _, p := range packs {
		s += "  - id: " + p[0] + "\n    source: \"" + filepath.ToSlash(p[1]) + "\"\n    ref: \"" + p[2] + "\"\n"
	}
	return s
}

func TestLayersApplyInOrder(t *testing.T) {
	e := setup(t)
	s1, c1 := layerPack(t, e.tmp, "first", "one/a.md")
	s2, c2 := layerPack(t, e.tmp, "second", "two/a.md")
	root := testpack.Project(t, e.tmp, "layers")
	writeFile(t, root, "bonsai.yaml", packsYAML([3]string{"first", s1, c1}, [3]string{"second", s2, c2}))
	e.apply(t, root, Request{Command: "init"})
	claude := read(t, root, "CLAUDE.md")
	if !strings.Contains(claude, "its packs: first 1.0.0, second 1.0.0.") || strings.Index(claude, "The first pack's line.") > strings.Index(claude, "The second pack's line.") {
		t.Errorf("the block is not in bonsai.yaml's order:\n%s", claude)
	}
	l, _ := workspace.LoadLock(root)
	if l.Packs[0].ID != "first" || l.Packs[1].ID != "second" {
		t.Errorf("the lock's packs: %+v", l.Packs)
	}
	writeFile(t, root, "bonsai.yaml", packsYAML([3]string{"second", s2, c2}, [3]string{"first", s1, c1}))
	e.apply(t, root, Request{})
	claude = read(t, root, "CLAUDE.md")
	if !strings.Contains(claude, "its packs: second 1.0.0, first 1.0.0.") || strings.Index(claude, "The second pack's line.") > strings.Index(claude, "The first pack's line.") {
		t.Errorf("after the order changed, the block:\n%s", claude)
	}
}

func wantRefused(t *testing.T, err error, code string, exit int, words string) {
	t.Helper()
	var ee *Error
	if !errors.As(err, &ee) || ee.Code != code || ee.Exit != exit || !strings.Contains(ee.What, words) || ee.Next == "" {
		t.Fatalf("want %s (exit %d) holding %q, got %v", code, exit, words, err)
	}
}

func TestTwoPacksWriteOnePath(t *testing.T) {
	e := setup(t)
	s1, c1 := layerPack(t, e.tmp, "first", "docs/guide.md")
	s2, c2 := layerPack(t, e.tmp, "second", "docs/guide.md")
	s3, c3 := layerPack(t, e.tmp, "third", "Docs/Guide.md")
	for _, c := range []struct {
		name  string
		other [3]string
	}{{"the same path", [3]string{"second", s2, c2}}, {"the path in another letter case", [3]string{"third", s3, c3}}} {
		t.Run(c.name, func(t *testing.T) {
			// At a first link.
			root := testpack.Project(t, e.tmp, "overlap-first-"+c.other[0])
			writeFile(t, root, "bonsai.yaml", packsYAML([3]string{"first", s1, c1}, c.other))
			before := snapshot(t, root)
			_, err := e.try(root, Request{Command: "init"})
			wantRefused(t, err, "packs-overlap", ExitInput, "two packs write")
			sameSnapshot(t, "a refused first link", before, snapshot(t, root))
			// At an update that adds the second pack.
			root = testpack.Project(t, e.tmp, "overlap-update-"+c.other[0])
			writeFile(t, root, "bonsai.yaml", packsYAML([3]string{"first", s1, c1}))
			e.apply(t, root, Request{Command: "init"})
			writeFile(t, root, "bonsai.yaml", packsYAML([3]string{"first", s1, c1}, c.other))
			before = snapshot(t, root)
			_, err = e.try(root, Request{})
			wantRefused(t, err, "packs-overlap", ExitInput, "two packs write")
			sameSnapshot(t, "a refused update", before, snapshot(t, root))
		})
	}
}

func TestOldWorkspaceIsRefused(t *testing.T) {
	e := setup(t)
	for _, old := range []string{".bonsai.yaml", ".bonsai-lock.yaml"} {
		t.Run(old, func(t *testing.T) {
			root := testpack.Project(t, e.tmp, "old"+old)
			writeFile(t, root, old, "agents: {}\n")
			writeFile(t, root, ".bonsai/catalog.json", "{}\n") // 0.4.3's own .bonsai/
			before := snapshot(t, root)
			_, err := e.try(root, Request{Command: "init", Init: e.values(e.pack.A), AllowExec: true})
			wantRefused(t, err, "old-workspace", ExitState, "Bonsai 0.4.3 workspace ("+old+")")
			sameSnapshot(t, "a refused init", before, snapshot(t, root))
			// A linked project in which the old file appears: update refuses too.
			linked := testpack.Project(t, e.tmp, "linked"+old)
			e.link(t, linked, e.pack.A)
			writeFile(t, linked, old, "agents: {}\n")
			before = snapshot(t, linked)
			_, err = e.try(linked, Request{AllowExec: true})
			wantRefused(t, err, "old-workspace", ExitState, "Bonsai 0.4.3 workspace")
			sameSnapshot(t, "a refused update", before, snapshot(t, linked))
		})
	}
}
