package engine

// The moved tag (spec §5: "A tag that later resolves to another commit is refused"): update exits 4, writes
// nothing, and refuses again on the next run; a new tag, a commit ref and a tag that never moved go through.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
)

// tagPack is a pack with two commits at one version (0.1.0), a file changed between them; its tag v0.1.0 starts at
// the first.
func tagPack(t *testing.T, tmp string) (string, []string) {
	t.Helper()
	src, shas := testpack.Fixture(t, tmp, "tagged",
		map[string]string{
			".claude-plugin/plugin.json": pluginJSON("tagged"),
			"bonsai/pack.yaml":           packYAML("tagged", "\n  - path: tagged/readme.md\n    from: readme.md\n    kind: pack", " []"),
			"bonsai/files/readme.md":     "one\n",
		},
		map[string]string{"bonsai/files/readme.md": "two\n"})
	testpack.Tag(t, src, "v0.1.0", shas[0])
	testpack.Tag(t, src, "stable", shas[0])
	return src, shas
}

func wantMoved(t *testing.T, err error) {
	t.Helper()
	var ee *Error
	if !errors.As(err, &ee) || ee.Code != "tag-moved" || ee.Exit != ExitState || ee.Who != "person" ||
		!strings.Contains(ee.What, "a tag that moves is refused") || !strings.Contains(ee.Next, "git ls-remote -- ") ||
		!strings.Contains(ee.Next, "then run bonsai update") {
		t.Fatalf("want tag-moved (exit 4, a person's step naming git ls-remote and bonsai update), got %v", err)
	}
}

func TestMovedTagIsRefused(t *testing.T) {
	e := setup(t)
	src, shas := tagPack(t, e.tmp)
	for _, tag := range []string{"v0.1.0", "stable"} {
		t.Run(tag, func(t *testing.T) {
			testpack.Tag(t, src, tag, shas[0])
			root := testpack.Project(t, e.tmp, "moved-"+tag)
			e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: tag}})
			testpack.Tag(t, src, tag, shas[1])
			before := snapshot(t, root)
			for run := 0; run < 2; run++ { // a refused run leaves the cache's record: the next one refuses too
				_, err := e.try(root, Request{})
				wantMoved(t, err)
				_, err = e.try(root, Request{Command: "init"})
				wantMoved(t, err)
			}
			sameSnapshot(t, "a moved tag", before, snapshot(t, root))
			// Taking the new commit is a person's edit of bonsai.yaml: the commit itself goes through.
			testpack.SetRef(t, root, `ref: "`+tag+`"`, `ref: "`+shas[1]+`"`)
			e.apply(t, root, Request{})
			if read(t, root, "tagged/readme.md") != "two\n" {
				t.Errorf("the new commit was not taken")
			}
		})
	}
}

// A release tag that names the locked version is caught on a machine whose cache never resolved it (a clone, CI): the
// lock's version says what the tag was.
func TestMovedTagOnAnotherMachine(t *testing.T) {
	e := setup(t)
	src, shas := tagPack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "moved-elsewhere")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: "v0.1.0"}})
	testpack.Tag(t, src, "v0.1.0", shas[1])
	if err := os.RemoveAll(filepath.Join(e.home, "cache")); err != nil {
		t.Fatal(err)
	}
	_, err := e.try(root, Request{})
	wantMoved(t, err)
}

// What is no moved tag: a new tag in bonsai.yaml (a person's choice), a tag that resolves where it did, and the
// ref-names rule's edges.
func TestTagNotMoved(t *testing.T) {
	e := setup(t)
	src, shas := tagPack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "new-tag")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: "v0.1.0"}})
	if p := e.plan(t, root, Request{}); !p.Nothing() {
		t.Errorf("a tag that did not move: %s", p.Preview(false))
	}
	testpack.Tag(t, src, "v0.2.0", shas[1])
	testpack.SetRef(t, root, `ref: "v0.1.0"`, `ref: "v0.2.0"`)
	e.apply(t, root, Request{})
	if read(t, root, "tagged/readme.md") != "two\n" {
		t.Errorf("a new tag was not taken")
	}
	for _, c := range []struct {
		tag, version string
		want         bool
	}{
		{"v1.0.0", "1.0.0", true}, {"1.0.0", "1.0.0", true}, {"base-v1.0.0", "1.0.0", true}, {"packs/base/v1.0.0", "1.0.0", true},
		{"v1.0.0", "1.0.1", false}, {"v11.0.0", "1.0.0", false}, {"xv1.0.0", "1.0.0", false}, {"stable", "1.0.0", false},
		{"v1.0.0", "", false},
	} {
		if got := tagNames(c.tag, c.version); got != c.want {
			t.Errorf("tagNames(%q, %q) = %v", c.tag, c.version, got)
		}
	}
}
