package formats

// The schema-compare test (contract §2.2: "A CI rung in Bonsai's repo fails on any schema change but additions";
// README.md, "How the set changes"). It compares each schema with the same file at the set's base commit, read with
// git show, and fails on anything but an addition:
//   - a new schema file;
//   - a new property at the end of an object's properties, its name added at the end of required;
//   - a change to description, examples or title, anywhere.
//
// Anything else fails: a schema file removed; a property removed, renamed or moved; a type, const, enum, pattern or
// bound changed, added or taken away; a name taken out of required, or required changed but for the new properties'
// names; any other keyword added or removed. A checkout without the base commit (git missing, or a shallow clone)
// skips with that reason, except under CI (the CI variable set), where it fails: CI checks out the history it needs.
// From step 5.4 the same test is a rung of Bonsai's own ladder.

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// schemaBase is the set's base: the commit of the set before this one (set 4), whose schemas this set may only add
// to. Each set's commit moves it forward to the commit of the set before it.
const schemaBase = "22dd08af9cf3691a61aee71e3aeec5d3fde71b38"

// annotationKeys may change freely: they document, they do not constrain.
var annotationKeys = map[string]bool{"description": true, "examples": true, "title": true}

// atBase gives a git runner for the repository, once it has checked the base commit is there. It skips the test (or
// fails it under CI) when the base cannot be read.
func atBase(t *testing.T, base string) func(args ...string) ([]byte, error) {
	t.Helper()
	unavailable := func(why string) {
		t.Helper()
		if os.Getenv("CI") != "" {
			t.Fatalf("the schema-compare test cannot read the base commit %s: %s; CI must check out the history it needs (fetch-depth: 0)", base, why)
		}
		t.Skipf("the base commit %s cannot be read here (%s), so the schemas are not compared with it", base, why)
	}
	if _, err := exec.LookPath("git"); err != nil {
		unavailable("git is not on the PATH")
	}
	git := func(args ...string) ([]byte, error) {
		cmd := exec.Command("git", append([]string{"-C", ".."}, args...)...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
		}
		return out, nil
	}
	if _, err := git("cat-file", "-e", base+"^{commit}"); err != nil {
		unavailable("no such commit in this checkout, as in a shallow clone")
	}
	return git
}

// baseSchemas reads every schema file at the base commit, by file name.
func baseSchemas(t *testing.T, base string) map[string][]byte {
	t.Helper()
	git := atBase(t, base)
	list, err := git("ls-tree", "--name-only", base, "formats/schemas/")
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]byte{}
	for _, p := range strings.Fields(string(list)) {
		raw, err := git("show", base+":"+p)
		if err != nil {
			t.Fatal(err)
		}
		out[strings.TrimPrefix(p, "formats/schemas/")] = raw
	}
	if len(out) == 0 {
		t.Fatalf("the base commit %s holds no schema", base)
	}
	return out
}

func TestSchemasChangeOnlyByAdditions(t *testing.T) {
	compareWithBase(t, schemaBase, "schemas")
}

// The base is the set before this one: its manifest's set is this manifest's minus one. A set's commit that forgets
// to move schemaBase forward fails here (README.md, "How the set changes").
func TestSchemaBaseIsTheSetBefore(t *testing.T) {
	git := atBase(t, schemaBase)
	raw, err := git("show", schemaBase+":formats/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	setOf := func(raw []byte, where string) int {
		t.Helper()
		v, err := schema.Decode(raw)
		if err != nil {
			t.Fatalf("%s: %v", where, err)
		}
		o, _ := v.(schema.Object)
		n, _ := o.Get("set")
		var set int
		if _, err := fmt.Sscan(fmt.Sprint(n), &set); err != nil {
			t.Fatalf("%s: set %v is not a number", where, n)
		}
		return set
	}
	now, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	was, is := setOf(raw, "manifest.json at "+schemaBase), setOf(now, "manifest.json")
	if was != is-1 {
		t.Errorf("schemaBase %s holds set %d, and this is set %d: the base is the set before (move schemaBase to the "+
			"commit of set %d)", schemaBase, was, is, is-1)
	}
}

// compareWithBase compares the schema files in dir with the base commit's, failing on every change but an addition.
func compareWithBase(t *testing.T, base, dir string) {
	t.Helper()
	old := baseSchemas(t, base)
	names := make([]string, 0, len(old))
	for n := range old {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		raw, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			t.Errorf("schemas/%s is gone: a schema is never removed inside a major (contract §2.2)", n)
			continue
		}
		was, err := schema.Decode(old[n])
		if err != nil {
			t.Fatalf("schemas/%s at %s: %v", n, base, err)
		}
		now, err := schema.Decode(raw)
		if err != nil {
			t.Fatalf("schemas/%s: %v", n, err)
		}
		for _, msg := range schemaChanges(was, now, "#") {
			t.Errorf("schemas/%s: %s: only additions are allowed inside a major (contract §2.2)", n, msg)
		}
	}
}

// schemaChanges lists every change from was to now that is not an addition, by JSON pointer into the schema.
func schemaChanges(was, now any, at string) []string {
	wo, okW := was.(schema.Object)
	no, okN := now.(schema.Object)
	if !okW || !okN {
		if schema.Show(was) != schema.Show(now) {
			return []string{fmt.Sprintf("%s changed from %s to %s", at, schema.Show(was), schema.Show(now))}
		}
		return nil
	}
	var out []string
	// The properties: the old ones, in their order, then any new ones; required: the old list, then the new names.
	var added []string
	if wp, ok := wo.Get("properties"); ok {
		wprops, _ := wp.(schema.Object)
		np, _ := no.Get("properties")
		nprops, _ := np.(schema.Object)
		for i, m := range wprops {
			here := at + "/properties/" + m.Key
			switch {
			case nprops.Index(m.Key) < 0:
				out = append(out, here+" was removed or renamed")
			case nprops.Index(m.Key) != i:
				out = append(out, fmt.Sprintf("%s moved from place %d to %d", here, i+1, nprops.Index(m.Key)+1))
			default:
				v, _ := nprops.Get(m.Key)
				out = append(out, schemaChanges(m.Value, v, here)...)
			}
		}
		if len(out) == 0 {
			for _, m := range nprops[len(wprops):] {
				added = append(added, m.Key)
			}
		}
	} else if _, ok := no.Get("properties"); ok {
		out = append(out, at+"/properties was added: it would narrow an open object")
	}
	wantRequired := []string{}
	if wr, ok := wo.Get("required"); ok {
		list, _ := wr.([]any)
		for _, r := range list {
			s, _ := r.(string)
			wantRequired = append(wantRequired, s)
		}
	}
	_, hadRequired := wo.Get("required")
	if hadRequired || len(added) > 0 {
		wantRequired = append(wantRequired, added...)
		nr, _ := no.Get("required")
		list, _ := nr.([]any)
		var got []string
		for _, r := range list {
			s, _ := r.(string)
			got = append(got, s)
		}
		if strings.Join(got, ",") != strings.Join(wantRequired, ",") {
			out = append(out, fmt.Sprintf("%s/required is %v, want the old list then the new properties' names %v", at, got, wantRequired))
		}
	} else if _, ok := no.Get("required"); ok {
		out = append(out, at+"/required was added")
	}
	// Every other keyword: unchanged, or (an annotation) anything.
	for _, m := range wo {
		if annotationKeys[m.Key] || m.Key == "properties" || m.Key == "required" {
			continue
		}
		v, ok := no.Get(m.Key)
		if !ok {
			out = append(out, fmt.Sprintf("%s/%s was removed", at, m.Key))
			continue
		}
		out = append(out, schemaChanges(m.Value, v, at+"/"+m.Key)...)
	}
	for _, m := range no {
		if annotationKeys[m.Key] || m.Key == "properties" || m.Key == "required" {
			continue
		}
		if _, ok := wo.Get(m.Key); !ok {
			out = append(out, fmt.Sprintf("%s/%s was added", at, m.Key))
		}
	}
	return out
}

// The comparison itself, on made-up schemas: what passes as an addition and what fails, so the rule does not rest
// on the committed files alone.
func TestSchemaChangesRule(t *testing.T) {
	base := `{"$schema": "x", "title": "t", "description": "d", "type": "object", "required": ["a", "b"],
	  "properties": {
	    "a": {"description": "a", "examples": ["x"], "type": "string", "pattern": "^x"},
	    "b": {"description": "b", "examples": [{}], "type": ["object", "null"], "required": ["c"],
	          "properties": {"c": {"description": "c", "examples": [1], "enum": [1, 2]}}},
	    "d": {"description": "d", "examples": [[]], "type": "array", "items": {"type": "string"}}}}`
	// A property added at the end of the top level's properties, with its name added to required, or not.
	withE := strings.Replace(base, `"items": {"type": "string"}}}}`,
		`"items": {"type": "string"}}, "e": {"description": "e", "examples": [1], "type": "integer"}}}`, 1)
	// Two properties swapped.
	swapped := `{"$schema": "x", "title": "t", "description": "d", "type": "object", "required": ["a", "b"],
	  "properties": {
	    "b": {"description": "b", "examples": [{}], "type": ["object", "null"], "required": ["c"],
	          "properties": {"c": {"description": "c", "examples": [1], "enum": [1, 2]}}},
	    "a": {"description": "a", "examples": ["x"], "type": "string", "pattern": "^x"},
	    "d": {"description": "d", "examples": [[]], "type": "array", "items": {"type": "string"}}}}`
	cases := []struct {
		name, now string
		pass      bool
	}{
		{"the same", base, true},
		{"words changed", strings.NewReplacer(`"description": "a"`, `"description": "A, said better"`,
			`"title": "t"`, `"title": "T"`, `"examples": ["x"]`, `"examples": ["xy", "xz"]`).Replace(base), true},
		{"a property added at the end, its name added to required",
			strings.Replace(withE, `"required": ["a", "b"]`, `"required": ["a", "b", "e"]`, 1), true},
		{"a property added at the end, not to required", withE, false},
		{"a nested property added at the end, its name added to required", strings.NewReplacer(
			`"required": ["c"]`, `"required": ["c", "f"]`,
			`"enum": [1, 2]}}}`, `"enum": [1, 2]}, "f": {"description": "f", "examples": [true], "type": "boolean"}}}`).Replace(base), true},
		{"a property removed", strings.Replace(base, `"a": {"description": "a", "examples": ["x"], "type": "string", "pattern": "^x"},`, ``, 1), false},
		{"two properties swapped", swapped, false},
		{"a pattern changed", strings.Replace(base, `"pattern": "^x"`, `"pattern": "^y"`, 1), false},
		{"a type widened", strings.Replace(base, `"type": "string", "pattern"`, `"type": ["string", "null"], "pattern"`, 1), false},
		{"an enum grown", strings.Replace(base, `"enum": [1, 2]`, `"enum": [1, 2, 3]`, 1), false},
		{"a name taken out of required", strings.Replace(base, `"required": ["c"]`, `"required": []`, 1), false},
		{"a bound added", strings.Replace(base, `"pattern": "^x"`, `"pattern": "^x", "maxLength": 9`, 1), false},
		{"items changed", strings.Replace(base, `"items": {"type": "string"}`, `"items": {"type": "integer"}`, 1), false},
	}
	was, err := schema.Decode([]byte(base))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		now, err := schema.Decode([]byte(c.now))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		got := schemaChanges(was, now, "#")
		if (len(got) == 0) != c.pass {
			t.Errorf("%s: changes %v, want pass %v", c.name, got, c.pass)
		}
	}
}
