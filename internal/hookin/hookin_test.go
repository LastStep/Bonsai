package hookin

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// The input hash: keyed by the salt, over the input with keys sorted at every depth and numbers as written, so key
// order does not count and a number's text does; 16 hex.
func TestInputHash(t *testing.T) {
	dec := func(raw string) any {
		v, err := schema.Decode([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	a := InputHash("k1", dec(`{"b":{"y":1,"x":[{"q":2,"p":1}]},"a":"é"}`))
	b := InputHash("k1", dec(`{"a":"é","b":{"x":[{"p":1,"q":2}],"y":1}}`))
	if a != b || !regexp.MustCompile(`^[0-9a-f]{16}$`).MatchString(a) {
		t.Errorf("key order counted: %s %s", a, b)
	}
	if InputHash("k2", dec(`{"a":1}`)) == InputHash("k1", dec(`{"a":1}`)) || InputHash("k1", dec(`{"a":1.0}`)) == InputHash("k1", dec(`{"a":1}`)) ||
		InputHash("k1", dec(`{"a":[1,2]}`)) == InputHash("k1", dec(`{"a":[2,1]}`)) {
		t.Error("the salt, a number's text or a list's order did not count")
	}
	// The hash of a payload's tool_input is the hash of the input itself: a call's events pair.
	p, err := Parse([]byte(`{"hook_event_name":"PostToolUse","tool_name":"Read","tool_input":{"limit":50,"file_path":"/p/a"},"tool_response":{}}`))
	if err != nil || p.Fault != nil {
		t.Fatal(err, p.Fault)
	}
	if InputHash("k1", p.Input) != InputHash("k1", dec(`{"file_path":"/p/a","limit":50}`)) {
		t.Error("a payload's tool_input hashes differently from the input")
	}
}

// FindProject and FindProjects (design/plan-5.md, 5.3.1 note 1): every folder at or above holding bonsai.yaml,
// nearest first, walking past any .git to the volume's root, so a nested repository or a .git placed in a subfolder
// does not hide the project around it; FindProject is the nearest.
func TestFindProject(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	mk := func(rel string) string {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}
	linkAt := func(p string) {
		if err := os.WriteFile(filepath.Join(p, "bonsai.yaml"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	proj, sub, plain, plainSub := mk("proj"), mk("proj/a/b"), mk("plain"), mk("plain/x")
	nested, nestedSub, planted := mk("proj/vendor/lib"), mk("proj/vendor/lib/src"), mk("proj/a/dotgit")
	mk("proj/.git")
	mk("plain/.git")
	mk("proj/vendor/lib/.git") // a nested repository, not linked
	mk("proj/a/dotgit/.git")   // a .git someone placed in a subfolder
	linkAt(proj)
	for _, c := range []struct {
		from, root string
		linked     bool
	}{
		{proj, proj, true}, {sub, proj, true}, {plain, plain, false}, {plainSub, plainSub, false},
		{nested, proj, true}, {nestedSub, proj, true}, {planted, proj, true},
	} {
		root, linked, err := FindProject(c.from)
		if err != nil || root != c.root || linked != c.linked {
			t.Errorf("from %s: %s %v %v; want %s %v", c.from, root, linked, err, c.root, c.linked)
		}
	}
	// A bonsai.yaml placed in the nested repository, and one above every checkout: each holds the folder too, the
	// nearest first.
	linkAt(nested)
	linkAt(dir)
	roots, abs, err := FindProjects(nestedSub)
	if err != nil || abs != nestedSub || len(roots) < 3 || roots[0] != nested || roots[1] != proj || roots[2] != dir {
		t.Errorf("FindProjects from the nested repository: %v %s %v", roots, abs, err)
	}
	if root, linked, err := FindProject(plainSub); err != nil || !linked || root != dir {
		t.Errorf("past a checkout's top: %s %v %v", root, linked, err)
	}
	for _, bad := range []string{filepath.Join(proj, "bonsai.yaml"), filepath.Join(dir, "missing")} {
		if _, linked, err := FindProject(bad); err == nil || linked {
			t.Errorf("%s: linked %v, err %v", bad, linked, err)
		}
	}
	// Holding: for a folder that need not exist (a new file's), each folder looked up once with seen.
	seen := map[string]bool{}
	got := Holding(filepath.Join(nestedSub, "new", "deeper"), seen)
	if len(got) < 3 || got[0] != nested || got[1] != proj || got[2] != dir {
		t.Errorf("Holding: %v", got)
	}
	if !seen[nested] || seen[nestedSub] || !seen[dir] {
		t.Errorf("seen: %v", seen)
	}
	if again := Holding(filepath.Join(nestedSub, "other"), seen); len(again) != len(got) || again[0] != nested {
		t.Errorf("Holding again: %v", again)
	}
}
