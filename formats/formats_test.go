// Package formats holds Bonsai's formats set: one JSON Schema per format, one example document per format, the
// trick files with their expected outcomes, and a manifest of every file's bytes (README.md says what each is).
//
// This test needs no reader: it checks that the set is whole and consistent. Standard library only, so it runs on
// any machine that has Go, and it survives the clear-out of the old product (plan part 1). It checks that:
//   - manifest.json matches every file's raw bytes, lists every file and nothing more, sorted by path;
//   - the CRLF case holds CRLF line endings and the BOM cases start with EF BB BF, as checked out;
//   - every rule of contract §2.4 has a case (the hand list below), and every case has both outcomes in expect.json;
//   - every schema is valid JSON, declares draft 2020-12, and documents itself (a description, and a description and
//     examples on every property); each example validates under the small checker below, keeps the schema's field
//     order, and each YAML or markdown example's format-1 value in expect.json equals its <name>.json;
//   - no file in the set holds a private string: an absolute home path, a drive letter, an email address, a tailnet
//     host.
package formats

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"
)

// The ten formats of contract §2 that the set describes, each with a schema and an example.
var formatNames = []string{"task", "labels", "lanes", "run", "state", "log", "ask", "ladder", "status", "lock"}

// The five formats whose files are YAML or markdown: their example source is also a case in expect.json.
var sourceExamples = map[string]string{
	"task":   "examples/task.md",
	"labels": "examples/labels.yaml",
	"lanes":  "examples/lanes.yaml",
	"run":    "examples/run.md",
	"state":  "examples/state.md",
}

// The rules of contract §2.4 (and the format-0 oddities and dispatch cases plan part 0 names). Every rule here needs
// at least one case in expect.json, and every case names one of these rules (or "example" for the five sources).
var rules = []string{
	// Lines.
	"lines-lf", "lines-crlf", "bom-frontmatter", "bom-definition", "tab-indent", "doc-marker", "every-line-read",
	// Keys.
	"key-pattern", "key-label", "key-uppercase", "key-hyphen", "key-quoted", "key-complex", "key-merge", "key-twice",
	"key-reserved", "dup-key-quoted",
	// Structure.
	"nested-mapping", "block-sequence", "seq-dash-space", "flow-sequence", "flow-empty", "anchor", "alias", "tag",
	"flow-mapping", "flow-nested", "flow-multiline", "block-scalar", "block-indicator", "block-hash-line",
	"folded-deeper",
	// Comments.
	"comment-line-start", "comment-after-space", "hash-no-space", "quote-in-plain",
	// Quoted scalars.
	"quoted-one-line", "quoted-single-escape", "quoted-escapes", "quoted-bad-escape", "quoted-unescaped-quote",
	"quoted-after-close", "quoted-multiline",
	// Plain scalars: every row of the table, each listed refusal, and each read as text once quoted.
	"plain-null", "plain-bool", "plain-int-15", "plain-int-16", "plain-decimal", "plain-date", "plain-datetime",
	"plain-text", "plain-other", "plain-listed-refusal", "plain-listed-quoted", "plain-word", "plain-ends-colon",
	"flow-item-char",
	// Today's format-0 oddities.
	"format0-oddity",
	// Dispatch.
	"dispatch-first", "dispatch-pointer", "dispatch-comment-first", "dispatch-none", "dispatch-nested",
	"dispatch-not-first",
}

// Files whose bytes are the point of their case, checked as checked out.
const crlfCase = "trick/yaml-1/lines-crlf/case.md"

var bomCases = []string{"trick/yaml-1/bom-frontmatter/case.md", "trick/yaml-1/bom-definition/case.yaml"}

// ---------------------------------------------------------------- the set's files and manifest

// setFiles lists every file under formats/ that the manifest covers: all but manifest.json and the Go files,
// forward-slash paths sorted by path.
func setFiles(t *testing.T) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := filepath.ToSlash(p)
		if rel == "manifest.json" || strings.HasSuffix(rel, ".go") {
			return nil
		}
		out = append(out, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walking formats/: %v", err)
	}
	sort.Strings(out)
	return out
}

func TestManifestMatchesEveryFile(t *testing.T) {
	m := mustObject(t, readJSON(t, "manifest.json"), "manifest.json")
	if _, ok := m.get("set"); !ok {
		t.Errorf("manifest.json: no set version")
	}
	filesV, _ := m.get("files")
	files, ok := filesV.([]any)
	if !ok {
		t.Fatalf("manifest.json: files is not a list")
	}
	var listed []string
	for i, f := range files {
		e := mustObject(t, f, fmt.Sprintf("manifest.json files[%d]", i))
		p, _ := e.get("path")
		h, _ := e.get("sha256")
		path, _ := p.(string)
		want, _ := h.(string)
		listed = append(listed, path)
		raw, err := os.ReadFile(filepath.FromSlash(path))
		if err != nil {
			t.Errorf("manifest.json lists %s, which cannot be read: %v", path, err)
			continue
		}
		sum := sha256.Sum256(raw)
		if got := hex.EncodeToString(sum[:]); got != want {
			t.Errorf("manifest.json: %s has sha256 %s, the manifest says %s (regenerate the manifest in the same commit)", path, got, want)
		}
	}
	if !sort.StringsAreSorted(listed) {
		t.Errorf("manifest.json: files are not sorted by path")
	}
	for i := 1; i < len(listed); i++ {
		if listed[i] == listed[i-1] {
			t.Errorf("manifest.json lists %s twice", listed[i])
		}
	}
	onDisk := setFiles(t)
	if strings.Join(onDisk, "\n") != strings.Join(listed, "\n") {
		t.Errorf("manifest.json does not list exactly the set's files.\nmissing from the manifest: %v\nlisted but not on disk: %v",
			minus(onDisk, listed), minus(listed, onDisk))
	}
}

// sortedKeys lists a set's keys in order, so failure messages come out byte-stable (no map-order iteration).
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func minus(a, b []string) []string {
	in := map[string]bool{}
	for _, x := range b {
		in[x] = true
	}
	var out []string
	for _, x := range a {
		if !in[x] {
			out = append(out, x)
		}
	}
	return out
}

func TestLineEndingsAndBOMSurviveCheckout(t *testing.T) {
	raw, err := os.ReadFile(filepath.FromSlash(crlfCase))
	if err != nil {
		t.Fatalf("the CRLF case: %v", err)
	}
	if !bytes.Contains(raw, []byte("\r\n")) {
		t.Errorf("%s holds no CRLF: a checkout changed its bytes (see .gitattributes, formats/** -text)", crlfCase)
	}
	if n := bytes.Count(raw, []byte("\n")); n != bytes.Count(raw, []byte("\r\n")) {
		t.Errorf("%s has a line ending in a bare LF: every line must end in CRLF", crlfCase)
	}
	for _, p := range bomCases {
		raw, err := os.ReadFile(filepath.FromSlash(p))
		if err != nil {
			t.Errorf("a BOM case: %v", err)
			continue
		}
		if !bytes.HasPrefix(raw, []byte{0xEF, 0xBB, 0xBF}) {
			t.Errorf("%s does not start with the BOM EF BB BF: a checkout changed its bytes", p)
		}
	}
}

// ---------------------------------------------------------------- expect.json

func TestEveryRuleHasACaseWithBothOutcomes(t *testing.T) {
	root := mustObject(t, readJSON(t, "expect.json"), "expect.json")
	casesV, _ := root.get("cases")
	cases, ok := casesV.([]any)
	if !ok {
		t.Fatalf("expect.json: cases is not a list")
	}
	codes := reasonCodes(t)
	known := map[string]bool{"example": true}
	for _, r := range rules {
		known[r] = true
	}
	seenRule := map[string]bool{}
	usedCode := map[string]bool{}
	var paths []string
	for i, c := range cases {
		where := fmt.Sprintf("expect.json cases[%d]", i)
		o := mustObject(t, c, where)
		path := str(o, "path")
		where += " (" + path + ")"
		paths = append(paths, path)
		if _, err := os.Stat(filepath.FromSlash(path)); err != nil {
			t.Errorf("%s: no such file", where)
		}
		rule := str(o, "rule")
		if !known[rule] {
			t.Errorf("%s: rule %q is not in the test's list of rules", where, rule)
		}
		seenRule[rule] = true
		if s := str(o, "settled_by"); s != "rule" && s != "libraries" {
			t.Errorf("%s: settled_by is %q, want rule or libraries", where, s)
		}
		if str(o, "about") == "" {
			t.Errorf("%s: no about", where)
		}
		f0v, ok0 := o.get("format0")
		f1v, ok1 := o.get("format1")
		if !ok0 || !ok1 {
			t.Errorf("%s: a case needs both outcomes, format0 and format1", where)
			continue
		}
		f0 := mustObject(t, f0v, where+" format0")
		switch str(f0, "outcome") {
		case "accepted":
			if _, ok := f0.get("value"); !ok {
				t.Errorf("%s: format0 accepted with no value", where)
			}
		case "refused":
			if str(f0, "message") == "" {
				t.Errorf("%s: format0 refused with no message", where)
			}
		default:
			t.Errorf("%s: format0 outcome %q, want accepted or refused", where, str(f0, "outcome"))
		}
		f1 := mustObject(t, f1v, where+" format1")
		switch str(f1, "outcome") {
		case "accepted":
			if _, ok := f1.get("value"); !ok {
				t.Errorf("%s: format1 accepted with no value", where)
			}
		case "refused":
			code := str(f1, "reason")
			if !codes[code] {
				t.Errorf("%s: format1 reason %q is not in README.md's reason codes", where, code)
			}
			usedCode[code] = true
		case "format-0":
		default:
			t.Errorf("%s: format1 outcome %q, want accepted, refused or format-0", where, str(f1, "outcome"))
		}
	}
	for _, r := range rules {
		if !seenRule[r] {
			t.Errorf("rule %s has no case in expect.json", r)
		}
	}
	for _, code := range sortedKeys(codes) {
		if !usedCode[code] {
			t.Errorf("README.md lists reason code %s, which no case uses", code)
		}
	}
	if !sort.StringsAreSorted(paths) {
		t.Errorf("expect.json: cases are not sorted by path")
	}
	for i := 1; i < len(paths); i++ {
		if paths[i] == paths[i-1] {
			t.Errorf("expect.json: %s appears twice", paths[i])
		}
	}
	// Every input file is a case: each file under trick/, and each YAML or markdown example.
	want := map[string]bool{}
	for _, p := range setFiles(t) {
		if strings.HasPrefix(p, "trick/") {
			want[p] = true
		}
	}
	for _, p := range sourceExamples {
		want[p] = true
	}
	have := map[string]bool{}
	for _, p := range paths {
		have[p] = true
	}
	for _, p := range sortedKeys(want) {
		if !have[p] {
			t.Errorf("%s has no case in expect.json", p)
		}
	}
	for _, p := range sortedKeys(have) {
		if !want[p] {
			t.Errorf("expect.json has a case for %s, which is neither a trick file nor an example source", p)
		}
	}
	// Each trick file sits alone in its case folder, as trick/yaml-1/<case>/case.yaml or case.md.
	caseFile := regexp.MustCompile(`^trick/yaml-[01]/[a-z0-9-]+/case\.(yaml|md)$`)
	for _, p := range sortedKeys(want) {
		if strings.HasPrefix(p, "trick/") && !caseFile.MatchString(p) {
			t.Errorf("%s is not laid out as trick/yaml-<0|1>/<case>/case.<yaml|md>", p)
		}
	}
}

// reasonCodes reads the format-1 reason codes from README.md's "## Reason codes" table: their one home.
func reasonCodes(t *testing.T) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("README.md: %v", err)
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	start := strings.Index(text, "\n## Reason codes\n")
	if start < 0 {
		t.Fatalf("README.md has no \"## Reason codes\" section")
	}
	section := text[start+1:]
	if end := strings.Index(section[3:], "\n## "); end >= 0 {
		section = section[:end+3]
	}
	row := regexp.MustCompile("(?m)^\\| `([a-z0-9-]+)` \\|")
	codes := map[string]bool{}
	for _, m := range row.FindAllStringSubmatch(section, -1) {
		codes[m[1]] = true
	}
	if len(codes) == 0 {
		t.Fatalf("README.md's reason codes table is empty")
	}
	return codes
}

// ---------------------------------------------------------------- the schemas and examples

const draft202012 = "https://json-schema.org/draft/2020-12/schema"

func TestSchemasDocumentThemselvesAndValidateTheirExamples(t *testing.T) {
	expect := mustObject(t, readJSON(t, "expect.json"), "expect.json")
	casesV, _ := expect.get("cases")
	byPath := map[string]object{}
	if cases, ok := casesV.([]any); ok {
		for _, c := range cases {
			if o, ok := c.(object); ok {
				byPath[str(o, "path")] = o
			}
		}
	}
	for _, name := range formatNames {
		t.Run(name, func(t *testing.T) {
			file := "schemas/" + name + ".schema.json"
			schema := mustObject(t, readJSON(t, file), file)
			if str(schema, "$schema") != draft202012 {
				t.Errorf("%s: $schema is not %s", file, draft202012)
			}
			if str(schema, "description") == "" {
				t.Errorf("%s: no top-level description", file)
			}
			for _, msg := range checkSchema(schema, "#", true) {
				t.Errorf("%s: %s", file, msg)
			}
			if t.Failed() {
				return
			}
			example := readJSON(t, "examples/"+name+".json")
			for _, msg := range validate(schema, example, "#") {
				t.Errorf("examples/%s.json does not validate: %s", name, msg)
			}
			for _, msg := range checkOrder(schema, example, "#") {
				t.Errorf("examples/%s.json: %s", name, msg)
			}
			src, ok := sourceExamples[name]
			if !ok {
				return
			}
			c, ok := byPath[src]
			if !ok {
				t.Fatalf("%s has no case in expect.json", src)
			}
			f1v, _ := c.get("format1")
			f1, _ := f1v.(object)
			if str(f1, "outcome") != "accepted" {
				t.Fatalf("%s: format1 outcome is %q, want accepted", src, str(f1, "outcome"))
			}
			v, _ := f1.get("value")
			if !equal(v, example) {
				t.Errorf("%s: its format-1 value in expect.json differs from examples/%s.json", src, name)
			}
		})
	}
}

// The keywords the checker implements. Annotations are skipped; any other keyword fails the schema.
var (
	annotations = map[string]bool{"title": true, "description": true, "examples": true}
	keywords    = map[string]bool{
		"type": true, "properties": true, "required": true, "additionalProperties": true, "propertyNames": true,
		"items": true, "enum": true, "const": true, "pattern": true, "maxLength": true, "minimum": true,
		"maximum": true, "maxItems": true,
	}
)

// checkSchema checks a schema's own form: only known keywords, every property documented with a description and
// examples that validate against it, required naming properties in their order (all of them at the top level).
func checkSchema(s object, at string, top bool) []string {
	var out []string
	for _, m := range s {
		switch {
		case annotations[m.key], keywords[m.key]:
		case top && m.key == "$schema":
		default:
			out = append(out, fmt.Sprintf("%s: keyword %q is not one the checker implements", at, m.key))
		}
	}
	if p, ok := s.get("pattern"); ok {
		if ps, ok := p.(string); !ok {
			out = append(out, at+": pattern is not a string")
		} else if _, err := regexp.Compile(ps); err != nil {
			out = append(out, fmt.Sprintf("%s: pattern %q: %v", at, ps, err))
		}
	}
	var propNames []string
	if pv, ok := s.get("properties"); ok {
		props, ok := pv.(object)
		if !ok {
			return append(out, at+": properties is not an object")
		}
		for _, m := range props {
			here := at + "/properties/" + m.key
			propNames = append(propNames, m.key)
			ps, ok := m.val.(object)
			if !ok {
				out = append(out, here+": not a schema object")
				continue
			}
			if d, _ := ps.get("description"); d == nil || d == "" {
				out = append(out, here+": no description")
			}
			ev, _ := ps.get("examples")
			if ex, ok := ev.([]any); !ok || len(ex) == 0 {
				out = append(out, here+": no examples")
			} else {
				for i, e := range ex {
					for _, msg := range validate(ps, e, fmt.Sprintf("examples[%d]", i)) {
						out = append(out, here+": its own example "+msg)
					}
				}
			}
			out = append(out, checkSchema(ps, here, false)...)
		}
	}
	if rv, ok := s.get("required"); ok {
		req, _ := rv.([]any)
		var names []string
		for _, r := range req {
			name, _ := r.(string)
			names = append(names, name)
		}
		if top && strings.Join(names, ",") != strings.Join(propNames, ",") {
			out = append(out, at+": a writer writes every field: required must list every property, in order")
		}
		j := 0
		for _, n := range names {
			for j < len(propNames) && propNames[j] != n {
				j++
			}
			if j == len(propNames) {
				out = append(out, fmt.Sprintf("%s: required names %q out of the properties' order, or not a property", at, n))
				break
			}
		}
	} else if top {
		out = append(out, at+": no required list")
	}
	for _, k := range []string{"items", "additionalProperties", "propertyNames"} {
		if v, ok := s.get(k); ok {
			if sub, ok := v.(object); ok {
				out = append(out, checkSchema(sub, at+"/"+k, false)...)
			} else if _, ok := v.(bool); !ok || k != "additionalProperties" {
				out = append(out, fmt.Sprintf("%s/%s: not a schema object", at, k))
			}
		}
	}
	return out
}

// validate checks an instance against a schema, for the keywords listed above, with draft 2020-12's meaning: each
// keyword applies only to instances of its own type.
func validate(s object, v any, at string) []string {
	var out []string
	if tv, ok := s.get("type"); ok {
		var types []string
		switch x := tv.(type) {
		case string:
			types = []string{x}
		case []any:
			for _, e := range x {
				name, _ := e.(string)
				types = append(types, name)
			}
		}
		match := false
		for _, ty := range types {
			if hasType(v, ty) {
				match = true
			}
		}
		if !match {
			return append(out, fmt.Sprintf("%s: is %s, want %s", at, typeOf(v), strings.Join(types, " or ")))
		}
	}
	if c, ok := s.get("const"); ok && !equal(c, v) {
		out = append(out, fmt.Sprintf("%s: is %s, want %s", at, show(v), show(c)))
	}
	if ev, ok := s.get("enum"); ok {
		in := false
		list, _ := ev.([]any)
		for _, e := range list {
			if equal(e, v) {
				in = true
			}
		}
		if !in {
			out = append(out, fmt.Sprintf("%s: %s is not one of %s", at, show(v), show(ev)))
		}
	}
	switch x := v.(type) {
	case string:
		if p, ok := s.get("pattern"); ok {
			ps, _ := p.(string)
			if re, err := regexp.Compile(ps); err != nil || !re.MatchString(x) {
				out = append(out, fmt.Sprintf("%s: %q does not match %s", at, x, ps))
			}
		}
		if n, ok := num(s, "maxLength"); ok && big.NewRat(int64(utf8.RuneCountInString(x)), 1).Cmp(n) > 0 {
			out = append(out, fmt.Sprintf("%s: longer than %s characters", at, n.RatString()))
		}
	case json.Number:
		r, _ := new(big.Rat).SetString(string(x))
		if n, ok := num(s, "minimum"); ok && r.Cmp(n) < 0 {
			out = append(out, fmt.Sprintf("%s: %s is below %s", at, x, n.RatString()))
		}
		if n, ok := num(s, "maximum"); ok && r.Cmp(n) > 0 {
			out = append(out, fmt.Sprintf("%s: %s is above %s", at, x, n.RatString()))
		}
	case []any:
		if n, ok := num(s, "maxItems"); ok && big.NewRat(int64(len(x)), 1).Cmp(n) > 0 {
			out = append(out, fmt.Sprintf("%s: more than %s items", at, n.RatString()))
		}
		if iv, ok := s.get("items"); ok {
			is, _ := iv.(object)
			for i, e := range x {
				out = append(out, validate(is, e, fmt.Sprintf("%s/%d", at, i))...)
			}
		}
	case object:
		if rv, ok := s.get("required"); ok {
			req, _ := rv.([]any)
			for _, r := range req {
				name, _ := r.(string)
				if _, ok := x.get(name); !ok {
					out = append(out, fmt.Sprintf("%s: required field %q is missing", at, name))
				}
			}
		}
		props, _ := s.get("properties")
		po, _ := props.(object)
		addl, hasAddl := s.get("additionalProperties")
		names, hasNames := s.get("propertyNames")
		for _, m := range x {
			here := at + "/" + m.key
			if hasNames {
				ns, _ := names.(object)
				out = append(out, validate(ns, m.key, here+" (its name)")...)
			}
			if ps, ok := po.get(m.key); ok {
				pso, _ := ps.(object)
				out = append(out, validate(pso, m.val, here)...)
				continue
			}
			if hasAddl {
				switch a := addl.(type) {
				case bool:
					if !a {
						out = append(out, here+": not allowed")
					}
				case object:
					out = append(out, validate(a, m.val, here)...)
				}
			}
		}
	}
	return out
}

// checkOrder checks that an instance's fields follow its schema's properties order (contract §2.2: a writer writes
// every field in a fixed order), at every depth the schema describes.
func checkOrder(s object, v any, at string) []string {
	var out []string
	switch x := v.(type) {
	case []any:
		if iv, ok := s.get("items"); ok {
			is, _ := iv.(object)
			for i, e := range x {
				out = append(out, checkOrder(is, e, fmt.Sprintf("%s/%d", at, i))...)
			}
		}
	case object:
		props, _ := s.get("properties")
		po, _ := props.(object)
		addl, _ := s.get("additionalProperties")
		ao, _ := addl.(object)
		last := -1
		for _, m := range x {
			idx := po.index(m.key)
			if idx >= 0 {
				if idx < last {
					out = append(out, fmt.Sprintf("%s/%s: out of the schema's field order", at, m.key))
				}
				last = idx
				ps, _ := po[idx].val.(object)
				out = append(out, checkOrder(ps, m.val, at+"/"+m.key)...)
			} else if ao != nil {
				out = append(out, checkOrder(ao, m.val, at+"/"+m.key)...)
			}
		}
	}
	return out
}

func hasType(v any, ty string) bool {
	switch ty {
	case "null":
		return v == nil
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "string":
		_, ok := v.(string)
		return ok
	case "number":
		_, ok := v.(json.Number)
		return ok
	case "integer":
		n, ok := v.(json.Number)
		if !ok {
			return false
		}
		r, ok := new(big.Rat).SetString(string(n))
		return ok && r.IsInt()
	case "array":
		_, ok := v.([]any)
		return ok
	case "object":
		_, ok := v.(object)
		return ok
	}
	return false
}

func typeOf(v any) string {
	for _, ty := range []string{"null", "boolean", "string", "integer", "number", "array", "object"} {
		if hasType(v, ty) {
			return ty
		}
	}
	return "unknown"
}

func num(s object, key string) (*big.Rat, bool) {
	v, ok := s.get(key)
	if !ok {
		return nil, false
	}
	n, ok := v.(json.Number)
	if !ok {
		return nil, false
	}
	return new(big.Rat).SetString(string(n))
}

// equal compares two JSON values: objects by their fields whatever the order, numbers by value.
func equal(a, b any) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case string:
		y, ok := b.(string)
		return ok && x == y
	case json.Number:
		y, ok := b.(json.Number)
		if !ok {
			return false
		}
		rx, okx := new(big.Rat).SetString(string(x))
		ry, oky := new(big.Rat).SetString(string(y))
		return okx && oky && rx.Cmp(ry) == 0
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !equal(x[i], y[i]) {
				return false
			}
		}
		return true
	case object:
		y, ok := b.(object)
		if !ok || len(x) != len(y) {
			return false
		}
		for _, m := range x {
			w, ok := y.get(m.key)
			if !ok || !equal(m.val, w) {
				return false
			}
		}
		return true
	}
	return false
}

func show(v any) string {
	b, err := json.Marshal(plain(v))
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}

// plain turns an ordered object back into Go values json.Marshal prints (fields sorted).
func plain(v any) any {
	switch x := v.(type) {
	case object:
		m := map[string]any{}
		for _, e := range x {
			m[e.key] = plain(e.val)
		}
		return m
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = plain(e)
		}
		return out
	}
	return v
}

// ---------------------------------------------------------------- nothing private

var privatePatterns = []struct {
	name string
	re   *regexp.Regexp
}{
	{"an absolute home path", regexp.MustCompile(`/home/`)},
	{"a WSL drive path", regexp.MustCompile(`/mnt/`)},
	{"a path under a home folder", regexp.MustCompile(`~/`)},
	{"a drive letter", regexp.MustCompile(`(^|[^A-Za-z0-9])[A-Za-z]:[\\/]`)},
	{"an email address", regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}`)},
	{"a tailnet host", regexp.MustCompile(`\.ts\.net`)},
}

func TestNoPrivateStrings(t *testing.T) {
	for _, p := range setFiles(t) {
		raw, err := os.ReadFile(filepath.FromSlash(p))
		if err != nil {
			t.Errorf("%s: %v", p, err)
			continue
		}
		for _, pat := range privatePatterns {
			if loc := pat.re.FindIndex(raw); loc != nil {
				t.Errorf("%s holds %s: %q", p, pat.name, raw[loc[0]:loc[1]])
			}
		}
	}
}

// ---------------------------------------------------------------- an ordered JSON reader

// object is a JSON object that keeps its fields' order; arrays are []any, numbers json.Number, null nil.
type object []member

type member struct {
	key string
	val any
}

func (o object) get(key string) (any, bool) {
	if i := o.index(key); i >= 0 {
		return o[i].val, true
	}
	return nil, false
}

func (o object) index(key string) int {
	for i, m := range o {
		if m.key == key {
			return i
		}
	}
	return -1
}

func str(o object, key string) string {
	v, _ := o.get(key)
	s, _ := v.(string)
	return s
}

func mustObject(t *testing.T, v any, where string) object {
	t.Helper()
	o, ok := v.(object)
	if !ok {
		t.Fatalf("%s: not a JSON object", where)
	}
	return o
}

// readJSON reads one JSON document, refusing a duplicate key (contract §2.5) and anything after the document.
func readJSON(t *testing.T, path string) any {
	t.Helper()
	raw, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	v, err := decodeValue(dec)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		t.Fatalf("%s: something follows the JSON document", path)
	}
	return v
}

func decodeValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch x := tok.(type) {
	case json.Delim:
		switch x {
		case '{':
			var o object
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, _ := kt.(string)
				if o.index(key) >= 0 {
					return nil, fmt.Errorf("duplicate key %q", key)
				}
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				o = append(o, member{key, v})
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			if o == nil {
				o = object{}
			}
			return o, nil
		case '[':
			arr := []any{}
			for dec.More() {
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v)
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			return arr, nil
		}
		return nil, fmt.Errorf("unexpected %v", x)
	default:
		return tok, nil
	}
}
