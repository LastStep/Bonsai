// Package format holds a Go type for each of Bonsai's eighteen formats (contract §2; formats set 4), held to the
// format's schema in formats/schemas, embedded by package formats: the schemas are the one home of every field and
// of every closed list, and this package keeps no second copy of either.
//
// For a reviewer, in the order a document goes through it:
//   - format.go: the registry, All, one Format per format in formats.Names' order: its name, its major, the shape
//     of its file, the majors Bonsai reads and whether Bonsai writes it, where its Go type lives, and the open lists
//     whose known words live in a Go table (List). A later piece adds a format, or an open list's table, by adding
//     an entry here (and its Go file); `bonsai check --schema`, status --json's formats and the registry's test all
//     read the registry, so nothing else needs to learn of it.
//   - read.go: reading, the same for every format (contract §2.2): a newer major is refused as "format too new" and
//     nothing else is read; a missing field reads as null (as does a present null where the schema allows none); an
//     unknown field is kept; a present field whose type or value the schema does not allow is refused, naming the
//     field (and, in a YAML or markdown file, its line) and the next step. Bonsai's own kinds (task, run, state) also
//     read under format 0, through internal/reader's format-0 mode, as they are: never held to format 1's schema.
//   - bind.go: a document into its Go type and back, by the schema's field order; a field this Bonsai does not know
//     is kept in the type's Extra and written after the known ones.
//   - write.go: writing, where Bonsai writes the format (Format.Writes): every field in the schema's order, the
//     document held to the full schema (every field required) before a byte is written, byte-stable (schema.Encode:
//     two-space indent or one line, LF, ASCII). yaml.go writes format 1's YAML (bonsai.yaml), table.go the two
//     generated tables.
//   - one file per format (task.go to changes.go): its Go type and its read and write functions.
//   - describe.go: `bonsai check --schema <format>`, a format with every field and allowed value, for a person.
//
// The lock's Go type is internal/workspace's Lock (lock.go), which this package's registry names; bonsai.yaml and
// pack.yaml are read here in full (Workspace, Pack) and by internal/workspace, whose Config and Pack carry the full
// read beside the fields the engine and the guard use. The guard reads bonsai.yaml through workspace.LoadConfig
// alone, never through this package's full read (spec §3: a hook starts in under 5 ms; plan-5 5.1.4a).
//
// Nothing here runs at start-up: schemas are parsed on first use, once (sync.Once), so a binary that never asks
// for a format pays nothing for it. Standard library only.
package format

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/schema"
)

// Shape is how a format's documents are stored.
type Shape int

const (
	YAMLFile Shape = iota // a YAML definition file: bonsai.yaml, pack.yaml, a pack's labels and lanes
	Markdown              // a markdown file whose YAML frontmatter is the document: a task, a run report, STATE, memory
	Table                 // a generated markdown table: frontmatter, then a body Bonsai writes and reads (table.go)
	JSONDoc               // a JSON document: status, the lock, a ladder result, the check and changes outputs
	JSONLine              // a JSON-lines record, one per line: the log, asks
	Part                  // part of other documents, with no format line of its own: the error object
)

var shapeNames = map[Shape]string{YAMLFile: "a YAML file", Markdown: "a markdown file's frontmatter",
	Table: "a generated markdown table", JSONDoc: "a JSON document", JSONLine: "a JSON line",
	Part: "an object inside other documents"}

// String names the shape for a person.
func (s Shape) String() string { return shapeNames[s] }

// Format is one of Bonsai's formats.
type Format struct {
	Name   string // the schema's short name, formats/schemas/<Name>.schema.json: "task"
	Major  int    // the major this Bonsai reads and writes; 0 for the error object, which has none
	Shape  Shape  // how its documents are stored
	Read0  bool   // Bonsai also reads it under format 0 (contract §2.3: Bonsai's own kinds task, run and state)
	Writes bool   // Bonsai writes it (plan-5 5.1.4a's list): the writer is held to the schema
	GoType string // where its Go type lives, for a reader of this code and of check --schema
	Next   string // the next step when a document of this format is refused for a field
	Lists  []List // the open lists whose known words live in a Go table, not in the schema
}

// List is an open list whose known words live in Bonsai's code, in one Go table (formats/README.md: a list of fixed
// words that code reads at many places); bonsai check --schema prints them. Every other list lives in its schema.
type List struct {
	Field string  // where the list is, as describe.go names a field: "code", "findings[].code"
	Table string  // the table's name, its one home: "format.ErrorWords"
	Words *[]Word // the table
}

// Word is one known word of an open list.
type Word struct {
	Word  string // the word itself
	Means string // one line: what it means
	Who   string // for a word that comes with a next step (the error object's, check's): who usually takes it, agent or person; else ""
	Kind  string // for check's words: finding (exit 1) or warning (never the exit code); else ""
}

// The next steps a refusal names, by who fixes the document.
const (
	nextPersonFile = "fix that field in the file (bonsai check --schema %s lists every field and its allowed values)"
	nextPack       = "fix that field in the pack's file and release the pack again (bonsai check --schema %s lists every field)"
	nextConfig     = "fix that line of bonsai.yaml, or restore the file from git (bonsai check --schema %s lists every field)"
	nextTable      = "rebuild it: run bonsai check --write in the main checkout (the table is generated, never edited by hand)"
	nextRecord     = "leave the file as it is and tell the person: Bonsai's writer never wrote that record (bonsai check --schema %s lists every field)"
	nextLadder     = "run the ladder again to write a new result (bonsai check --schema %s lists every field)"
	nextOutput     = "run the command again with --json: Bonsai's writer never prints that document (bonsai check --schema %s lists every field)"
	nextLock       = "restore .bonsai/lock.json from git (git checkout -- .bonsai/lock.json); the lock is Bonsai's to write"
)

// errorCode is the error object's code in the formats that carry it as their error field: the open list of
// ErrorWords.
var errorCode = []List{{Field: "error.code", Table: "format.ErrorWords", Words: &ErrorWords}}

// All is every format, in formats.Names' order (the ten of contract §2, then the eight of set 4). Its test holds it
// to formats.Names and to the schemas, so a schema with no entry here, or an entry with no schema, fails.
//
// To add a format: its schema in formats/ (a new set), its Go file here, and an entry below. To add an open list's
// table (step 5.2.0's log events and categories): the table beside its format's Go type, and a List in its entry.
var All = []*Format{
	{Name: "task", Major: 1, Shape: Markdown, Read0: true, GoType: "format.Task", Next: nextPersonFile},
	{Name: "labels", Major: 1, Shape: YAMLFile, GoType: "format.Labels", Next: nextPack},
	{Name: "lanes", Major: 1, Shape: YAMLFile, GoType: "format.Lanes", Next: nextPack},
	{Name: "run", Major: 1, Shape: Markdown, Read0: true, GoType: "format.Run", Next: nextPersonFile},
	{Name: "state", Major: 1, Shape: Markdown, Read0: true, GoType: "format.State", Next: nextPersonFile},
	{Name: "log", Major: 1, Shape: JSONLine, Writes: true, GoType: "format.Log", Next: nextRecord},
	{Name: "ask", Major: 1, Shape: JSONLine, Writes: true, GoType: "format.Ask", Next: nextRecord},
	{Name: "ladder", Major: 1, Shape: JSONDoc, Writes: true, GoType: "format.Ladder", Next: nextLadder},
	{Name: "status", Major: 1, Shape: JSONDoc, Writes: true, GoType: "format.Status", Next: nextOutput, Lists: errorCode},
	{Name: "lock", Major: 1, Shape: JSONDoc, Writes: true, GoType: "workspace.Lock (internal/workspace/lock.go)", Next: nextLock},
	{Name: "workspace", Major: 1, Shape: YAMLFile, Writes: true, GoType: "format.Workspace", Next: nextConfig},
	{Name: "pack", Major: 1, Shape: YAMLFile, GoType: "format.Pack", Next: nextPack},
	{Name: "tasks", Major: 1, Shape: Table, Writes: true, GoType: "format.Tasks", Next: nextTable},
	{Name: "sessions", Major: 1, Shape: Table, Writes: true, GoType: "format.Sessions", Next: nextTable},
	{Name: "memory", Major: 1, Shape: Markdown, GoType: "format.Memory", Next: nextPersonFile},
	{Name: "error", Major: 0, Shape: Part, Writes: true, GoType: "format.ErrorObject", Next: nextOutput,
		Lists: []List{{Field: "code", Table: "format.ErrorWords", Words: &ErrorWords}}},
	{Name: "check", Major: 1, Shape: JSONDoc, Writes: true, GoType: "format.Check", Next: nextOutput,
		Lists: append([]List{
			{Field: "findings[].code", Table: "format.CheckWords", Words: &CheckWords},
			{Field: "findings[].code", Table: "format.PackCheckWords", Words: &PackCheckWords},
			{Field: "warnings[].code", Table: "format.CheckWords", Words: &CheckWords},
		}, errorCode...)},
	{Name: "changes", Major: 1, Shape: JSONDoc, Writes: true, GoType: "format.Changes", Next: nextOutput, Lists: errorCode},
}

// Lookup finds a format by its short name ("task"), its name ("bonsai.task") or its name and major
// ("bonsai.task/1").
func Lookup(name string) (*Format, bool) {
	short := strings.TrimPrefix(name, "bonsai.")
	if i := strings.IndexByte(short, '/'); i >= 0 {
		short = short[:i]
	}
	for _, f := range All {
		if f.Name == short && (name == f.Name || name == f.ID() || name == f.Versioned()) {
			return f, true
		}
	}
	return nil, false
}

// MustLookup is Lookup for a name this package's own code gives; an unknown one is a bug.
func MustLookup(name string) *Format {
	f, ok := Lookup(name)
	if !ok {
		panic("format: no format " + name)
	}
	return f
}

// Names lists every format's name ("bonsai.task"), in All's order: what an unknown name's refusal lists.
func Names() []string {
	out := make([]string, len(All))
	for i, f := range All {
		out[i] = f.ID()
	}
	return out
}

// ID is the format's name, bonsai.<name> (contract §2.1).
func (f *Format) ID() string { return "bonsai." + f.Name }

// Versioned is the format's name and major as its format line holds it, "bonsai.task/1"; the error object, which
// has no major, is its name alone.
func (f *Format) Versioned() string {
	if f.Major == 0 {
		return f.ID()
	}
	return f.ID() + "/" + strconv.Itoa(f.Major)
}

// ReadMajors lists the majors Bonsai reads: 0 (format 0, today's files) for task, run and state, then the major.
func (f *Format) ReadMajors() []int {
	if f.Read0 {
		return []int{0, f.Major}
	}
	return []int{f.Major}
}

// next is the format's next step for a refused field.
func (f *Format) next() string {
	if strings.Contains(f.Next, "%s") {
		return fmt.Sprintf(f.Next, f.ID())
	}
	return f.Next
}

// schemas holds each format's parsed schema, parsed on first use.
type schemas struct {
	once   sync.Once
	full   schema.Object // what a writer writes: every field required
	read   schema.Object // what a reader holds a document to: no field required (schema.Lenient)
	failed error
}

var parsed sync.Map // name -> *schemas

func (f *Format) load() *schemas {
	v, _ := parsed.LoadOrStore(f.Name, &schemas{})
	s := v.(*schemas)
	s.once.Do(func() {
		raw, err := formats.Schema(f.Name)
		if err == nil {
			s.full, err = schema.Parse(raw)
		}
		if err != nil {
			s.failed = fmt.Errorf("the embedded schema of %s cannot be read: %v", f.ID(), err)
			return
		}
		s.read = schema.Lenient(s.full)
	})
	return s
}

// Schema is the format's schema, from formats/schemas (embedded): what a writer writes. It panics if the embedded
// schema cannot be read, which the formats test rules out.
func (f *Format) Schema() schema.Object {
	s := f.load()
	if s.failed != nil {
		panic(s.failed)
	}
	return s.full
}

// ReadSchema is the schema a reader holds a document to: the format's schema with no field required (contract
// §2.2: a reader reads a missing field as null).
func (f *Format) ReadSchema() schema.Object {
	f.Schema()
	return f.load().read
}

// StatusFormats is status --json's formats (contract §12): every format with a major, by name, with the majors this
// Bonsai reads and the one it writes, in All's order. The error object has no major, so it is not listed.
func StatusFormats() schema.Object {
	out := schema.Object{}
	for _, f := range All {
		if f.Major == 0 {
			continue
		}
		read := []any{}
		for _, m := range f.ReadMajors() {
			read = append(read, jsonInt(int64(m)))
		}
		out = append(out, schema.Member{Key: f.ID(), Value: schema.Object{
			{Key: "read", Value: read}, {Key: "write", Value: jsonInt(int64(f.Major))}}})
	}
	return out
}
