package workspace

// Tests of the lock: it reads the set's example, writes it back byte for byte, validates against the lock schema,
// keeps unknown fields, refuses what it should, and writes whole.

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// schemaAt walks a schema by keys and shows the value there.
func schemaAt(t *testing.T, s schema.Object, keys ...string) string {
	t.Helper()
	var v any = s
	for _, k := range keys {
		o, ok := v.(schema.Object)
		if !ok {
			t.Fatalf("schema: no %v", keys)
		}
		v, _ = o.Get(k)
	}
	return schema.Show(v)
}

func exampleLock(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "formats", "examples", "lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestLockReadsAndWritesTheSetsExample(t *testing.T) {
	raw := exampleLock(t)
	l, err := ReadLock(raw)
	if err != nil {
		t.Fatal(err)
	}
	if l.WrittenBy != "1.0.0" || len(l.Packs) != 2 || l.Packs[1].ID != "workflow" || len(l.Files) != 2 ||
		l.Files["CLAUDE.md"].Kind != "block" || len(l.Format0) != 1 || len(l.Packs[1].Declares) != 1 {
		t.Errorf("read %+v", l)
	}
	out, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, raw) {
		t.Errorf("Encode(ReadLock(example)) differs from formats/examples/lock.json:\n%s", out)
	}
}

// What the writer writes validates against part 0's lock schema and keeps its field order, at every depth.
func TestLockOutputFitsTheSchema(t *testing.T) {
	l := &Lock{
		WrittenBy: "dev",
		Packs: []LockedPack{{ID: "test-pack", Source: "https://example.com/test-pack.git", Version: "0.1.0",
			Commit: strings.Repeat("a", 40), SHA256: strings.Repeat("b", 64), Path: "packs/test", PathSet: true}},
		Files: map[string]LockedFile{
			"z/last.md":             {Kind: "pack", Pack: "test-pack", SHA256: strings.Repeat("c", 64)},
			".claude/settings.json": {Kind: "keys", Pack: "test-pack", SHA256: strings.Repeat("d", 64)},
			"a/first.md":            {Kind: "once", Pack: "test-pack", SHA256: strings.Repeat("e", 64)},
		},
	}
	out, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := schema.Decode(out)
	if err != nil {
		t.Fatal(err)
	}
	if msgs := schema.Validate(LockSchema(), doc); len(msgs) != 0 {
		t.Errorf("the written lock does not validate: %v", msgs)
	}
	if msgs := schema.CheckOrder(LockSchema(), doc); len(msgs) != 0 {
		t.Errorf("the written lock is out of the schema's order: %v", msgs)
	}
	files, _ := doc.(schema.Object).Get("files")
	if got := strings.Join(files.(schema.Object).Keys(), " "); got != ".claude/settings.json a/first.md z/last.md" {
		t.Errorf("files not sorted: %s", got)
	}
	if !strings.Contains(string(out), `"declares": {}`) || !strings.Contains(string(out), `"format0": {}`) {
		t.Errorf("empty declares and format0 are not written as {}:\n%s", out)
	}
	for i := 0; i < 10; i++ {
		again, _ := l.Encode()
		if !bytes.Equal(again, out) {
			t.Fatal("two encodings differ")
		}
	}
}

// Formats set 4 gives each locked pack its folder, path (null for the repository's root). A lock written before has
// none: it reads as unknown (PathSet false, step 5.1.3's verifier, F7), not as the root, and the writer refuses to write
// a pack whose folder it does not know; a path, null or a folder, is read and written back; and a written path is held
// to the schema.
func TestLockPackPath(t *testing.T) {
	commit, sum := strings.Repeat("0", 40), strings.Repeat("1", 64)
	pack := func(path string) string {
		return `{
      "id": "p",
      "source": "https://example.com/p.git",
      "version": "1.0.0",
      "commit": "` + commit + `",
      "sha256": "` + sum + `",
      "declares": {}` + path + `
    }`
	}
	doc := func(path string) string {
		return `{
  "format": "bonsai.lock/1",
  "written_by": "dev",
  "packs": [
    ` + pack(path) + `
  ],
  "files": {},
  "format0": {}
}
`
	}
	for _, c := range []struct {
		name, path, want string
		set              bool
	}{
		{"a lock from before set 4", "", "", false},
		{"the repository's root", ",\n      \"path\": null", "", true},
		{"a folder", ",\n      \"path\": \"packs/p\"", "packs/p", true},
	} {
		raw := []byte(doc(c.path))
		l, err := ReadLock(raw)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if l.Packs[0].Path != c.want || l.Packs[0].PathSet != c.set {
			t.Errorf("%s: path %q set %v, want %q %v", c.name, l.Packs[0].Path, l.Packs[0].PathSet, c.want, c.set)
		}
		out, err := l.Encode()
		if !c.set {
			if err == nil || !strings.Contains(err.Error(), "does not record the folder of the pack") {
				t.Errorf("%s: a pack whose folder is unknown was written (%v):\n%s", c.name, err, out)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !bytes.Equal(out, raw) {
			t.Errorf("%s: written back as\n%s\nwant\n%s", c.name, out, raw)
		}
		back, _ := schema.Decode(out)
		if msgs := schema.Validate(LockSchema(), back); len(msgs) != 0 {
			t.Errorf("%s: the written lock does not fit the schema: %v", c.name, msgs)
		}
	}
	// A path that is not a project-relative folder does not fit the schema, read or written.
	for _, bad := range []string{`"/abs"`, `"c:/x"`, `"a\\b"`, `7`, `"../up"`, `"x/.git/y"`} {
		if _, err := ReadLock([]byte(doc(",\n      \"path\": " + bad))); err == nil {
			t.Errorf("a lock whose pack path is %s was read", bad)
		}
	}
	l := &Lock{WrittenBy: "dev", Packs: []LockedPack{{ID: "p", Source: "s", Version: "1", Commit: commit, SHA256: sum,
		Path: "/abs", PathSet: true}}}
	if _, err := l.Encode(); err == nil {
		t.Errorf("a lock whose pack path is absolute was written")
	}
}

func TestLockKeepsUnknownFields(t *testing.T) {
	in := `{"format": "bonsai.lock/1", "written_by": "9.0.0", "later": {"x": 1},
	"packs": [{"id": "p", "source": "s", "version": "1", "commit": "` + strings.Repeat("0", 40) + `",
	  "sha256": "` + strings.Repeat("1", 64) + `", "new_pack_field": [1, 2], "declares": {"lanes": []}, "path": null}],
	"files": {"a.md": {"kind": "pack", "new_file_field": true, "pack": "p", "sha256": "` + strings.Repeat("2", 64) + `"}},
	"format0": {}, "last": null}`
	l, err := ReadLock([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	out, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	doc, _ := schema.Decode(out)
	o := doc.(schema.Object)
	if got := strings.Join(o.Keys(), " "); got != "format written_by packs files format0 later last" {
		t.Errorf("top-level keys %s", got)
	}
	packs, _ := o.Get("packs")
	p := packs.([]any)[0].(schema.Object)
	if got := strings.Join(p.Keys(), " "); got != "id source version commit sha256 declares path new_pack_field" {
		t.Errorf("pack keys %s", got)
	}
	files, _ := o.Get("files")
	f, _ := files.(schema.Object).Get("a.md")
	if got := strings.Join(f.(schema.Object).Keys(), " "); got != "kind pack sha256 new_file_field" {
		t.Errorf("file keys %s", got)
	}
	if v, _ := o.Get("later"); !schema.Equal(v, schema.Object{{Key: "x", Value: jsonNumber("1")}}) {
		t.Errorf("later = %s", schema.Show(v))
	}
}

func TestLockRefuses(t *testing.T) {
	good := string(exampleLock(t))
	cases := []struct {
		name, in, want string
	}{
		{"a duplicate key", strings.Replace(good, `"written_by": "1.0.0",`, `"written_by": "1.0.0", "written_by": "1.0.0",`, 1), "duplicate key"},
		{"a BOM", "\xEF\xBB\xBF" + good, "byte order mark"},
		{"trailing data", good + "{}", "something follows"},
		{"not an object", "[]", "is not a JSON object"},
		{"a newer major", strings.Replace(good, "bonsai.lock/1", "bonsai.lock/2", 1), "format too new"},
		{"another format", strings.Replace(good, "bonsai.lock/1", "bonsai.task/1", 1), `its format is "bonsai.task/1"`},
		{"a kind not in the schema", strings.Replace(good, `"kind": "block"`, `"kind": "copy"`, 1), `"copy" is not one of`},
		{"a short commit", strings.Replace(good, `"0a1b2c3d4e5f60718293a4b5c6d7e8f901a2b3c4"`, `"0a1b2c3"`, 1), "does not match"},
		{"a missing format", strings.Replace(good, `"format": "bonsai.lock/1",`, ``, 1), "its format is null"},
		{"a field of the wrong kind", strings.Replace(good, `"version": "1.0.0",`, `"version": 1,`, 1), "is integer, want string"},
		{"two paths in letter case", strings.Replace(good, `"CLAUDE.md": {`, `"claude.md": {"kind": "pack", "pack": "base", "sha256": "`+strings.Repeat("0", 64)+`"}, "CLAUDE.md": {`, 1), `differ only in letter case`},
		{"an absolute path", strings.Replace(good, `"CLAUDE.md": {`, `"/etc/CLAUDE.md": {`, 1), "does not match"},
		{"a path out", strings.Replace(good, `"CLAUDE.md": {`, `"a/../../CLAUDE.md": {`, 1), "files: the path"},
		{"a format0 path out", strings.Replace(good, `"work/runs/`, `"../runs/`, 1), "format0: the path"},
		{"a pack twice", strings.Replace(good, `"id": "workflow"`, `"id": "base"`, 1), `the pack "base" is locked twice`},
		{"declares not an object", strings.Replace(good, `"declares": {}`, `"declares": []`, 1), "is array, want object"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ReadLock([]byte(c.in))
			if err == nil {
				t.Fatal("read with no error")
			}
			if msg := err.Error(); !strings.HasPrefix(msg, ".bonsai/lock.json: ") || !strings.Contains(msg, c.want) || !strings.Contains(msg, "; next: ") {
				t.Errorf("error %q, want it to hold %q and name a next step", msg, c.want)
			}
			if msg := err.Error(); strings.Contains(msg, `\u`) && c.name != "a BOM" {
				t.Errorf("an ASCII file gave a message with an escape (our own text is not ASCII): %q", msg)
			}
		})
	}
	// The writer refuses what the reader would.
	bad := []*Lock{
		{WrittenBy: "dev", Packs: []LockedPack{{ID: "p", Commit: "short", SHA256: strings.Repeat("0", 64)}}},
		{WrittenBy: "dev", Files: map[string]LockedFile{"a.md": {Kind: "copy", Pack: "p", SHA256: strings.Repeat("0", 64)}}},
		{WrittenBy: "dev", Files: map[string]LockedFile{"a\\b.md": {Kind: "pack", Pack: "p", SHA256: strings.Repeat("0", 64)}}},
		{WrittenBy: "dev", Format0: map[string]string{"nul.md": strings.Repeat("0", 64)}},
	}
	for i, l := range bad {
		if _, err := l.Encode(); err == nil {
			t.Errorf("bad lock %d encoded", i)
		}
		if err := WriteLock(t.TempDir(), l); err == nil {
			t.Errorf("bad lock %d written", i)
		}
	}
}

func TestWriteLockAndLoadLock(t *testing.T) {
	root := t.TempDir()
	if _, err := LoadLock(root); !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "next: run bonsai update") {
		t.Fatalf("a missing lock: %v", err)
	}
	l, err := ReadLock(exampleLock(t))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ { // the second write replaces the first
		if err := WriteLock(root, l); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(root, ".bonsai", "lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, exampleLock(t)) {
		t.Errorf("the written lock differs from what was read")
	}
	entries, _ := os.ReadDir(filepath.Join(root, ".bonsai"))
	if len(entries) != 1 {
		t.Errorf(".bonsai holds %d files, want the lock alone (no temporary file left)", len(entries))
	}
	back, err := LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := back.Encode()
	if !bytes.Equal(again, raw) {
		t.Errorf("LoadLock does not read back what WriteLock wrote")
	}
}

// Contract §2.2: a reader treats a missing field, or a null one, as null, an older writer of the same major; the
// writer still writes every field.
func TestLockReadsMissingFieldsAsNull(t *testing.T) {
	in := `{"format": "bonsai.lock/1", "packs": [{"id": "p", "source": null, "version": "1"}],
	  "files": {"a.md": {"kind": "pack"}}, "format0": null}`
	l, err := ReadLock([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	if l.WrittenBy != "" || len(l.Packs) != 1 || l.Packs[0].Source != "" || l.Packs[0].Commit != "" ||
		len(l.Packs[0].Declares) != 0 || l.Files["a.md"].SHA256 != "" || len(l.Format0) != 0 {
		t.Errorf("read %+v", l)
	}
	if _, err := l.Encode(); err == nil {
		t.Errorf("a lock with missing fields was written: the writer writes every field")
	}
	l2, err := ReadLock([]byte(`{"format": "bonsai.lock/1"}`))
	if err != nil || len(l2.Packs) != 0 || len(l2.Files) != 0 || l2.Files == nil {
		t.Errorf("a lock with only its format: %+v, %v", l2, err)
	}
}

func TestLockRefusesTwoPathsInLetterCase(t *testing.T) {
	l := &Lock{WrittenBy: "dev", Format0: map[string]string{"work/A.md": strings.Repeat("0", 64), "work/a.md": strings.Repeat("1", 64)}}
	if _, err := l.Encode(); err == nil || !strings.Contains(err.Error(), "differ only in letter case") {
		t.Errorf("Encode: %v", err)
	}
}
