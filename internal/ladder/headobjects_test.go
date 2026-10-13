package ladder

// V2b's S6: a main checkout whose HEAD names a commit git cannot read in (an object not in the repository, as in a
// clone whose borrowed objects are gone) refuses the climb and blocks the stop gate, each saying so and naming git fsck
// as the person's step: there is no line of bonsai.yaml to fix, and git status reports nothing to fix either.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
)

// dropObject takes the object rev names out of the checkout's repository (a loose object, as a fresh commit leaves it).
func dropObject(t *testing.T, root, rev string) {
	t.Helper()
	sha := testpack.Git(t, root, "rev-parse", rev)
	p := filepath.Join(root, ".git", "objects", sha[:2], sha[2:])
	// git writes a loose object read-only, and Windows removes a read-only file only once it is made writable.
	if err := os.Chmod(p, 0o644); err != nil {
		t.Fatalf("%s is not a loose object: %v", rev, err)
	}
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
}

func TestHeadObjectsNotThere(t *testing.T) {
	const next = "a person runs git fsck in the main checkout, which names each object that is not there or is damaged, then " +
		"fetches them (git fetch) or makes a fresh clone of the project"
	for _, c := range []struct{ name, rev, why string }{
		{"bonsai.yaml's object", "HEAD:bonsai.yaml", ", but git cannot read bonsai.yaml there (HEAD's bonsai.yaml is not a file git can read"},
		{"the commit's object", "HEAD", ", which git cannot read (the commit is not in the repository, or is damaged)"},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "one", command: "echo one", required: true}),
				map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
			head := f.head()
			dropObject(t, f.root, c.rev)
			a := f.climb("", Options{Task: "T-0901"})
			if a.Exit != ExitRuntime || a.Error == nil || a.Error.Code != "read-failed" || a.Error.Who != "person" ||
				!strings.Contains(a.Error.What, "the main checkout's HEAD names commit "+head[:12]+c.why) ||
				a.Error.Next != next+"; then run: bonsai ladder --task T-0901 --json" {
				t.Errorf("the climb: exit %d, %+v", a.Exit, a.Error)
			}
			if code, stderr := gate(t, f.root, "T-0901"); code != 2 || !strings.Contains(stderr, "the main checkout's HEAD names commit "+
				head[:12]+c.why) || !strings.Contains(stderr, "next: "+next) || strings.Contains(stderr, "bonsai check names the line") {
				t.Errorf("the stop gate: %d %q", code, stderr)
			}
		})
	}
}
