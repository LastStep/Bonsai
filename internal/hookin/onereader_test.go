package hookin

// One reader (design/plan-5.md, 5.3.0 note 2): no package but this one decodes a hook payload, so a field Claude Code
// adds or renames is read in one place.
//
// What "decodes a payload" means here: a package's code (its .go files, tests aside) names a key of a hook payload's
// top level that only a payload has, in any of the ways Go's decoders find a key:
//   - as a string literal (a map key, a lookup), or a concatenation of string literals ("tool_"+"input"), folded;
//   - as the JSON name in a struct field's tag;
//   - as an untagged struct field's own name (or one whose tag gives no name), which encoding/json matches to a key
//     regardless of letter case: a field Hook_event_name or Tool_input reads hook_event_name or tool_input.
// A key built at run time any other way (from a constant declared elsewhere, strings.Join, fmt.Sprintf, a variable)
// is not caught: telling those would take the type checker over the whole module on every test run, for a reader no
// one writes by accident. The key is one of this
// package's own table (fields) or one of the payload keys it passes over (passedOver), less every property any of
// Bonsai's formats names (formats/schemas, their one home) and every word of the log's events (format.LogEvents),
// since a format's own key or word (session, source, tool_use_id; the event prompt) is no sign of a payload. A grep of schema.Decode would not do: the lock, the settings and the log are read with it
// too. A package that read a payload would have to name hook_event_name, tool_input or the like to reach its fields.
//
// The whole module is walked from its root (go.mod), every package but this one; testdata and hidden folders aside.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// passedOver are top-level keys of Claude Code's hook payloads (the hooks reference, 2.1.296) that Bonsai does not
// read: a package naming one is reading a payload all the same.
var passedOver = []string{"tool_response", "last_assistant_message", "permission_suggestions", "background_tasks",
	"session_crons", "prompt_id", "scratchpad_dir", "session_title", "duration_ms", "mcp_server"}

// payloadOnly lists the keys that mark a payload's reader (this file's comment), sorted.
func payloadOnly(t *testing.T) []string {
	t.Helper()
	ours := map[string]bool{}
	err := fs.WalkDir(formats.Schemas(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".schema.json") {
			return err
		}
		raw, err := fs.ReadFile(formats.Schemas(), path)
		if err != nil {
			return err
		}
		v, err := schema.Decode(raw)
		if err != nil {
			return err
		}
		properties(v, ours)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ours) < 50 {
		t.Fatalf("the formats name %d properties: the schemas were not read", len(ours))
	}
	for _, w := range format.LogEvents {
		ours[w.Word] = true
	}
	var keys []string
	for _, f := range fields {
		keys = append(keys, f.key)
	}
	keys = append(keys, passedOver...)
	var out []string
	for _, k := range keys {
		if !ours[k] {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// properties adds every name under a "properties" object in a schema, at any depth.
func properties(v any, into map[string]bool) {
	switch x := v.(type) {
	case schema.Object:
		for _, m := range x {
			if p, ok := m.Value.(schema.Object); ok && m.Key == "properties" {
				for _, k := range p.Keys() {
					into[k] = true
				}
			}
			properties(m.Value, into)
		}
	case []any:
		for _, e := range x {
			properties(e, into)
		}
	}
}

// moduleRoot finds the folder holding go.mod, above this package.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod above this package")
		}
		dir = parent
	}
}

// payloadReaders lists, as "file:line: key", every place outside this package that names a payload-only key.
func payloadReaders(t *testing.T, root string, keys []string) []string {
	t.Helper()
	marks := map[string]bool{}
	for _, k := range keys {
		marks[k] = true
	}
	self, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	var found []string
	fset := token.NewFileSet()
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && (strings.HasPrefix(name, ".") || name == "testdata" || path == self) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		mark := func(s string, at token.Pos) {
			if marks[s] {
				rel, _ := filepath.Rel(root, fset.Position(at).Filename)
				found = append(found, filepath.ToSlash(rel)+":"+strconv.Itoa(fset.Position(at).Line)+": "+s)
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.StructType:
				for _, field := range x.Fields.List {
					name := ""
					if field.Tag != nil {
						if tag, err := strconv.Unquote(field.Tag.Value); err == nil {
							name = strings.Split(reflect.StructTag(tag).Get("json"), ",")[0]
						}
					}
					switch name {
					case "-":
					case "":
						// No JSON name: encoding/json matches the field's own name, letter case aside.
						for _, id := range field.Names {
							mark(strings.ToLower(id.Name), id.Pos())
						}
					default:
						mark(name, field.Tag.Pos())
					}
				}
			case *ast.BinaryExpr:
				if s, ok := literal(x); ok {
					mark(s, x.Pos())
					return false // a concatenation of literals, read whole: its parts are not counted again
				}
			case *ast.BasicLit:
				if s, ok := literal(x); ok {
					mark(s, x.Pos())
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}

// literal gives a string literal's text, or a concatenation of string literals folded (in parentheses too).
func literal(e ast.Expr) (string, bool) {
	switch x := e.(type) {
	case *ast.BasicLit:
		if x.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(x.Value)
		return s, err == nil
	case *ast.ParenExpr:
		return literal(x.X)
	case *ast.BinaryExpr:
		if x.Op != token.ADD {
			return "", false
		}
		a, ok := literal(x.X)
		if !ok {
			return "", false
		}
		b, ok := literal(x.Y)
		return a + b, ok
	}
	return "", false
}

func TestOneReader(t *testing.T) {
	keys := payloadOnly(t)
	for _, k := range []string{"hook_event_name", "tool_input", "tool_name", "session_id", "tool_response"} {
		if i := sort.SearchStrings(keys, k); i == len(keys) || keys[i] != k {
			t.Fatalf("%s is not among the payload-only keys %v", k, keys)
		}
	}
	if found := payloadReaders(t, moduleRoot(t), keys); len(found) != 0 {
		t.Errorf("a package other than internal/hookin decodes a hook payload (it names a key only a payload has); "+
			"read the payload through internal/hookin instead:\n  %s", strings.Join(found, "\n  "))
	}
}

// The walk sees a payload's reader, by a literal, a concatenation of literals, a struct tag, or an untagged field's
// name in any letter case, and a literal compared inside a larger expression: a folder of its own holding each is
// found. A function's parameter of such a name, a field the decoder skips (json:"-") and a concatenation with a
// variable are not readers it can tell.
func TestOneReaderSeesAReader(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := "package x\n\ntype p struct {\n\tTool string `json:\"tool_name,omitempty\"`\n}\n\nvar k = map[string]int{\"hook_event_name\": 1}\n" +
		"\ntype q struct {\n\tHook_event_name string\n\tTool_Input      map[string]any `json:\",omitempty\"`\n" +
		"\tTool_name       string `json:\"-\"`\n\tSessionID, Session_id string\n}\n" +
		"\nvar m = map[string]any{\"tool_\" + \"input\": 1, (\"transcript\" + \"_\") + \"path\": 2}\n" +
		"\nfunc f(tool_input string, v string) bool { return \"agent_\"+v == \"x\" || v == \"stop_hook_active\" }\n"
	if err := os.MkdirAll(filepath.Join(root, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", "x.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	// A test file and a format's own key are not counted.
	if err := os.WriteFile(filepath.Join(root, "a", "x_test.go"), []byte("package x\n\nvar q = \"tool_input\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	found := payloadReaders(t, root, payloadOnly(t))
	want := []string{"a/x.go:4: tool_name", "a/x.go:7: hook_event_name", "a/x.go:10: hook_event_name",
		"a/x.go:11: tool_input", "a/x.go:13: session_id", "a/x.go:16: tool_input", "a/x.go:16: transcript_path",
		"a/x.go:18: stop_hook_active"}
	if !reflect.DeepEqual(found, want) {
		t.Errorf("found %v, want %v", found, want)
	}
}
