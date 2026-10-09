// The formats set's own test (README.md, "The Go test"). It needs no YAML reader: it checks that the set is whole
// and consistent. Standard library plus Bonsai's schema checker (internal/schema, moved out of this file in plan part
// 2 so Bonsai's code uses the same one), so it runs on any machine that has Go. It checks that:
//   - manifest.json matches every file's raw bytes, lists every file and nothing more, sorted by path;
//   - the CRLF case holds CRLF line endings, the lone-CR case CR alone, and the BOM cases start with EF BB BF, as
//     checked out;
//   - every rule of contract §2.4 has a case (the hand list below), and every case has both outcomes in expect.json;
//   - every schema is valid JSON, declares draft 2020-12, and documents itself (a description, and a description and
//     examples on every property); each example validates under the schema checker, keeps the schema's field
//     order, and each YAML or markdown example's format-1 value in expect.json equals its <name>.json (for the two
//     tables, the fields of <name>.json before what a reader reads from the body);
//   - the error object held inline in status, check and changes, and every next object, are error.schema.json's,
//     and a rung's kind in bonsai.yaml is the ladder result's list (copies held to their one home);
//   - contract §13's fixtures and their answers file are well formed and agree (fixtures_test.go);
//   - each schema changed from the set's base commit only by additions (compare_test.go);
//   - the embedded schemas (embed.go) are exactly the files in schemas/, byte for byte;
//   - no file in the set holds a private string: an absolute home path, a drive letter, an email address, a tailnet
//     host.
package formats

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// The formats whose files are YAML or markdown: their example source is also a case in expect.json.
var sourceExamples = map[string]string{
	"task":      "examples/task.md",
	"labels":    "examples/labels.yaml",
	"lanes":     "examples/lanes.yaml",
	"run":       "examples/run.md",
	"state":     "examples/state.md",
	"workspace": "examples/workspace.yaml",
	"pack":      "examples/pack.yaml",
	"tasks":     "examples/tasks.md",
	"sessions":  "examples/sessions.md",
	"memory":    "examples/memory.md",
}

// bodyFields are, for the two generated tables, the fields a reader reads from the markdown body rather than the
// frontmatter (README.md, "The examples"): they come after the frontmatter's, and the case's format-1 value is the
// example without them.
var bodyFields = map[string][]string{
	"tasks":    {"active", "tasks"},
	"sessions": {"sessions", "hours"},
}

// The rules of contract §2.4 (and the format-0 oddities and dispatch cases plan part 0 names). Every rule here needs
// at least one case in expect.json, and every case names one of these rules (or "example" for the example sources).
var rules = []string{
	// Lines.
	"lines-lf", "lines-crlf", "lines-cr", "bom-frontmatter", "bom-definition", "tab-indent", "doc-marker", "every-line-read",
	"not-text",
	// Keys.
	"key-pattern", "key-label", "key-uppercase", "key-hyphen", "key-quoted", "key-complex", "key-merge", "key-twice",
	"key-reserved", "dup-key-quoted", "key-length",
	// Structure.
	"nested-mapping", "block-sequence", "seq-dash-space", "flow-sequence", "flow-empty", "anchor", "alias", "tag",
	"flow-mapping", "flow-nested", "flow-multiline", "flow-trailing-comma", "block-scalar", "block-indicator",
	"block-hash-line", "folded-deeper", "block-space-line",
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
	"dispatch-not-first", "dispatch-tab", "format-too-new",
	// The order of codes on one line.
	"code-order",
}

// Files whose bytes are the point of their case, checked as checked out.
const (
	crlfCase = "trick/yaml-1/lines-crlf/case.md"
	crCase   = "trick/yaml-1/lines-cr/case.yaml"
)

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
	if _, ok := m.Get("set"); !ok {
		t.Errorf("manifest.json: no set version")
	}
	filesV, _ := m.Get("files")
	files, ok := filesV.([]any)
	if !ok {
		t.Fatalf("manifest.json: files is not a list")
	}
	var listed []string
	for i, f := range files {
		e := mustObject(t, f, fmt.Sprintf("manifest.json files[%d]", i))
		p, _ := e.Get("path")
		h, _ := e.Get("sha256")
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
	raw, err = os.ReadFile(filepath.FromSlash(crCase))
	if err != nil {
		t.Fatalf("the lone-CR case: %v", err)
	}
	if bytes.Contains(raw, []byte("\n")) || bytes.Count(raw, []byte("\r")) < 2 {
		t.Errorf("%s must end its lines in CR alone, with no LF: a checkout changed its bytes", crCase)
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
	casesV, _ := root.Get("cases")
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
		f0v, ok0 := o.Get("format0")
		f1v, ok1 := o.Get("format1")
		if !ok0 || !ok1 {
			t.Errorf("%s: a case needs both outcomes, format0 and format1", where)
			continue
		}
		f0 := mustObject(t, f0v, where+" format0")
		switch str(f0, "outcome") {
		case "accepted":
			if _, ok := f0.Get("value"); !ok {
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
			if _, ok := f1.Get("value"); !ok {
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

func TestSchemasDocumentThemselvesAndValidateTheirExamples(t *testing.T) {
	expect := mustObject(t, readJSON(t, "expect.json"), "expect.json")
	casesV, _ := expect.Get("cases")
	byPath := map[string]schema.Object{}
	if cases, ok := casesV.([]any); ok {
		for _, c := range cases {
			if o, ok := c.(schema.Object); ok {
				byPath[o.String("path")] = o
			}
		}
	}
	for _, name := range Names {
		t.Run(name, func(t *testing.T) {
			file := "schemas/" + name + ".schema.json"
			s := mustObject(t, readJSON(t, file), file)
			if s.String("$schema") != schema.Draft202012 {
				t.Errorf("%s: $schema is not %s", file, schema.Draft202012)
			}
			if s.String("description") == "" {
				t.Errorf("%s: no top-level description", file)
			}
			for _, msg := range schema.CheckSchema(s) {
				t.Errorf("%s: %s", file, msg)
			}
			if t.Failed() {
				return
			}
			example := readJSON(t, "examples/"+name+".json")
			for _, msg := range schema.Validate(s, example) {
				t.Errorf("examples/%s.json does not validate: %s", name, msg)
			}
			for _, msg := range schema.CheckOrder(s, example) {
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
			f1v, _ := c.Get("format1")
			f1, _ := f1v.(schema.Object)
			if f1.String("outcome") != "accepted" {
				t.Fatalf("%s: format1 outcome is %q, want accepted", src, f1.String("outcome"))
			}
			v, _ := f1.Get("value")
			want := example
			if body := bodyFields[name]; len(body) > 0 {
				want = frontmatterOf(t, example.(schema.Object), body, name)
			}
			if !schema.Equal(v, want) {
				t.Errorf("%s: its format-1 value in expect.json differs from examples/%s.json", src, name)
			}
		})
	}
}

// frontmatterOf gives a table's example without the fields a reader reads from its body, which must be its last.
func frontmatterOf(t *testing.T, example schema.Object, body []string, name string) schema.Object {
	t.Helper()
	keys := example.Keys()
	if len(keys) < len(body) || strings.Join(keys[len(keys)-len(body):], " ") != strings.Join(body, " ") {
		t.Fatalf("examples/%s.json: its last fields are %v, want the body's %v", name, keys, body)
	}
	return example[:len(keys)-len(body)]
}

// ---------------------------------------------------------------- copies held to their one home

// schemaAt walks a schema by keys, failing the test where a key is missing.
func schemaAt(t *testing.T, s schema.Object, where string, keys ...string) schema.Object {
	t.Helper()
	var v any = s
	for _, k := range keys {
		o, ok := v.(schema.Object)
		if !ok {
			t.Fatalf("%s: no %s", where, strings.Join(keys, "/"))
		}
		v, ok = o.Get(k)
		if !ok {
			t.Fatalf("%s: no %s", where, strings.Join(keys, "/"))
		}
	}
	o, ok := v.(schema.Object)
	if !ok {
		t.Fatalf("%s: %s is not an object", where, strings.Join(keys, "/"))
	}
	return o
}

func loadSchema(t *testing.T, name string) schema.Object {
	t.Helper()
	file := "schemas/" + name + ".schema.json"
	return mustObject(t, readJSON(t, file), file)
}

// sameShape holds a copy's required list and properties equal to its home's, key order and every word included.
func sameShape(t *testing.T, copyOf, home schema.Object, where string) {
	t.Helper()
	for _, k := range []string{"required", "properties"} {
		a, _ := copyOf.Get(k)
		b, _ := home.Get(k)
		if b == nil || schema.Show(a) != schema.Show(b) {
			t.Errorf("%s: its %s differ from its one home's: change the home and every copy together", where, k)
		}
	}
}

// The schema checker has no $ref, so status, check and changes hold the error object inline, and findings, warnings
// and the plugin step hold its next object: each copy is error.schema.json's, its one home (README.md). And a rung's
// kind in bonsai.yaml is the ladder result's closed list.
func TestCopiesAreTheirHomes(t *testing.T) {
	errorSchema := loadSchema(t, "error")
	next := schemaAt(t, errorSchema, "error", "properties", "next")
	for _, name := range []string{"status", "check", "changes"} {
		sameShape(t, schemaAt(t, loadSchema(t, name), name, "properties", "error"), errorSchema, name+"'s error")
	}
	check := loadSchema(t, "check")
	for _, list := range []string{"findings", "warnings"} {
		sameShape(t, schemaAt(t, check, "check", "properties", list, "items", "properties", "next"), next,
			"check's "+list+"[].next")
	}
	sameShape(t, schemaAt(t, loadSchema(t, "changes"), "changes", "properties", "plugins", "items", "properties", "next"),
		next, "changes' plugins[].next")
	rung := schemaAt(t, loadSchema(t, "workspace"), "workspace", "properties", "ladder", "items", "properties", "kind")
	home := schemaAt(t, loadSchema(t, "ladder"), "ladder", "properties", "rungs", "items", "properties", "kind")
	a, _ := rung.Get("enum")
	b, _ := home.Get("enum")
	if b == nil || schema.Show(a) != schema.Show(b) {
		t.Errorf("workspace's ladder[].kind is %s, the ladder result's rung kind %s: one list", schema.Show(a), schema.Show(b))
	}
}

// The embedded schemas (embed.go) are what Bonsai's code reads: they must be exactly the files in schemas/, and
// Names must name each of them once.
func TestEmbeddedSchemasAreTheFiles(t *testing.T) {
	var embedded []string
	err := fs.WalkDir(Schemas(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		embedded = append(embedded, p)
		return nil
	})
	if err != nil {
		t.Fatalf("walking the embedded schemas: %v", err)
	}
	var onDisk []string
	for _, p := range setFiles(t) {
		if strings.HasPrefix(p, "schemas/") {
			onDisk = append(onDisk, p)
		}
	}
	var named []string
	for _, n := range Names {
		named = append(named, "schemas/"+n+".schema.json")
	}
	sort.Strings(named)
	if strings.Join(embedded, "\n") != strings.Join(onDisk, "\n") || strings.Join(named, "\n") != strings.Join(onDisk, "\n") {
		t.Fatalf("embedded %v, on disk %v, named by Names %v: all three must be the same files", embedded, onDisk, named)
	}
	for _, n := range Names {
		got, err := Schema(n)
		if err != nil {
			t.Errorf("Schema(%q): %v", n, err)
			continue
		}
		want, err := os.ReadFile(filepath.FromSlash("schemas/" + n + ".schema.json"))
		if err != nil {
			t.Fatalf("%v", err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("Schema(%q) differs from schemas/%s.schema.json", n, n)
		}
		if _, err := schema.Parse(got); err != nil {
			t.Errorf("schema.Parse(Schema(%q)): %v", n, err)
		}
	}
	if _, err := Schema("no-such-format"); err == nil {
		t.Errorf("Schema of an unknown name returned no error")
	}
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

// ---------------------------------------------------------------- reading JSON

func str(o schema.Object, key string) string {
	return o.String(key)
}

func mustObject(t *testing.T, v any, where string) schema.Object {
	t.Helper()
	o, ok := v.(schema.Object)
	if !ok {
		t.Fatalf("%s: not a JSON object", where)
	}
	return o
}

// readJSON reads one JSON document with the schema package's reader, which refuses a duplicate key (contract §2.5)
// and anything after the document.
func readJSON(t *testing.T, path string) any {
	t.Helper()
	raw, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	v, err := schema.Decode(raw)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return v
}
