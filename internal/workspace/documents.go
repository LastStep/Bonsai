package workspace

// Document kinds (contract §7.3): every kind of document a workspace holds is declared, by Bonsai for its own kinds
// and by the packs for theirs, so no reader hard-codes a project's folders. Bonsai's own kinds and their places are
// the table bonsaiKinds below, their one home: task, run, state, answers, memory, and the two generated tables tasks
// and sessions. A pack's kinds come from the lock's declares (format.Declares), each at the place bonsai.yaml's
// documents names under the kind's name, else at the pack's default (spec §6: init writes that default into
// bonsai.yaml). DocKinds lists them all, Bonsai's first, then each pack's in the lock's order (bonsai.yaml's), and
// DocFiles lists the documents of a folder kind. status --json prints them (documents); check, the tables, the
// active task (active.go) and, from step 5.3, the guard read them.

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// DocKind is one document kind, with contract §7.3's fields.
type DocKind struct {
	Kind      string        // the kind's name
	From      string        // who declared it: bonsai, or a pack's id
	Path      string        // the folder holding one file per document (top level only), project-relative; "" for a file kind
	File      string        // the single file, project-relative; "" for a folder kind
	ID        string        // the id pattern a document's id matches, "" for none
	Format    string        // its bonsai.* format, for Bonsai's kinds ("bonsai.task"); "" for a pack's
	Statuses  []string      // its statuses, the first where a new document starts; none for a kind without
	Person    [][]string    // the status moves a person makes, as [from, to] pairs
	Agent     [][]string    // the status moves agents make
	Stamp     schema.Object // fields set to the day's date when a document reaches a status: {status: field}
	TaskField string        // the field naming the task a document belongs to, "" for none
}

// bonsaiKind is one of Bonsai's own kinds: where it lives (a key of bonsai.yaml's documents, a folder or a file, or
// a fixed file), its format, and its id pattern.
type bonsaiKind struct {
	Kind   string
	Key    string // bonsai.yaml's documents key naming its place; "" for a fixed file
	IsFile bool   // the place is one file (answers), not a folder
	Fixed  string // its fixed file (STATE and the two tables, contract §3), "" when bonsai.yaml names it
	Format string // its format's short name, "" for none (answers holds answers, not a Bonsai format)
	ID     string // its id pattern (the format's schema's, held equal by TestBonsaiKindsAreTheSchemas)
}

// The id patterns of Bonsai's kinds that have ids, as their schemas hold them (formats/schemas, their one home): kept
// here as text so the task folder's reader (active.go, which the guard uses from step 5.3) parses no schema, and held
// equal to the schemas by a test.
const (
	TaskIDPattern   = `^T-[0-9]{4,6}$`
	runIDPattern    = `^R-[0-9]{4}-[0-9]{2}-[0-9]{2}-T-[0-9]{4,6}`
	memoryIDPattern = `^M-[a-z0-9][a-z0-9-]*$`
)

// bonsaiKinds are Bonsai's own document kinds, in contract §7.3's order: their one home.
var bonsaiKinds = []bonsaiKind{
	{Kind: "task", Key: "task", Format: "task", ID: TaskIDPattern},
	{Kind: "run", Key: "run", Format: "run", ID: runIDPattern},
	{Kind: "state", Fixed: StateFile, Format: "state"},
	{Kind: "answers", Key: "answers", IsFile: true},
	{Kind: "memory", Key: "memory", Format: "memory", ID: memoryIDPattern},
	{Kind: "tasks", Fixed: TasksTableFile, Format: "tasks"},
	{Kind: "sessions", Fixed: SessionsTableFile, Format: "sessions"},
}

// The fixed files of Bonsai's kinds (contract §3, §7.2, §7.5).
const (
	StateFile         = ".bonsai/STATE.md"
	TasksTableFile    = ".bonsai/tasks.md"
	SessionsTableFile = ".bonsai/sessions.md"
)

// taskStamp is the task kind's stamp (contract §7.3: "started only when empty").
var taskStamp = schema.Object{{Key: "running", Value: "started"}, {Key: "done", Value: "finished"}, {Key: "cut", Value: "finished"}}

// BonsaiKind reports whether a kind's name is one of Bonsai's own, which no pack may declare.
func BonsaiKind(name string) bool {
	for _, k := range bonsaiKinds {
		if k.Kind == name {
			return true
		}
	}
	return false
}

// Format0Kinds are the kinds bonsai check holds to format 0 (contract §2.3): Bonsai's own markdown kinds that had a
// format 0, task, run and state.
var Format0Kinds = []string{"task", "run", "state"}

// DocKinds lists the workspace's document kinds: Bonsai's (bonsaiKinds) at the places w's documents names, then
// each locked pack's from its declares, at the place w's documents names under the kind's name or else the pack's
// default. lock may be nil (no packs). Its error is a lock pack whose declares Bonsai does not read.
func DocKinds(w *format.Workspace, lock *Lock) ([]DocKind, error) {
	places := map[string]string{"task": w.Documents.Task, "run": w.Documents.Run, "answers": w.Documents.Answers,
		"memory": w.Documents.Memory}
	for _, m := range w.Documents.Extra {
		if s, ok := m.Value.(string); ok {
			places[m.Key] = s
		}
	}
	var out []DocKind
	for _, b := range bonsaiKinds {
		k := DocKind{Kind: b.Kind, From: "bonsai", ID: b.ID}
		switch {
		case b.Fixed != "":
			k.File = b.Fixed
		case b.IsFile:
			k.File = places[b.Key]
		default:
			k.Path = places[b.Key]
		}
		if b.Format != "" {
			k.Format = format.MustLookup(b.Format).ID()
		}
		if b.Kind == "task" {
			k.Statuses = taskStatuses()
			k.Stamp = taskStamp
		}
		out = append(out, k)
	}
	if lock == nil {
		return out, nil
	}
	for _, lp := range lock.Packs {
		d, err := lp.Declared()
		if err != nil {
			return nil, lockError("the pack %s's %v", showValue(lp.ID), err)
		}
		for _, pk := range d.Documents {
			k := DocKind{Kind: pk.Kind, From: lp.ID, Statuses: pk.Statuses, Person: pk.Person, Agent: pk.Agent, Stamp: pk.Stamp}
			if pk.Path != nil {
				k.Path = *pk.Path
			}
			if pk.File != nil {
				k.File = *pk.File
			}
			if place, ok := places[pk.Kind]; ok {
				if k.File != "" {
					k.File = place
				} else {
					k.Path = place
				}
			}
			if pk.ID != nil {
				k.ID = *pk.ID
			}
			if pk.TaskField != nil {
				k.TaskField = *pk.TaskField
			}
			out = append(out, k)
		}
	}
	return out, nil
}

var (
	statusesOnce sync.Once
	statuses     []string
)

// taskStatuses is the task's closed list of statuses, from bonsai.task/1's schema (its one home).
func taskStatuses() []string {
	statusesOnce.Do(func() {
		props, _ := format.MustLookup("task").Schema().Get("properties")
		st, _ := props.(schema.Object).Get("status")
		enum, _ := st.(schema.Object).Get("enum")
		for _, v := range enum.([]any) {
			statuses = append(statuses, v.(string))
		}
	})
	return append([]string{}, statuses...)
}

// Object gives the kind as status --json's documents entry holds it (contract §7.3): kind and from, then the
// fields that apply, in the contract's order; a field that does not apply is left out.
func (k DocKind) Object() schema.Object {
	o := schema.Object{{Key: "kind", Value: k.Kind}, {Key: "from", Value: k.From}}
	add := func(key string, v any, applies bool) {
		if applies {
			o = append(o, schema.Member{Key: key, Value: v})
		}
	}
	add("path", k.Path, k.Path != "")
	add("file", k.File, k.File != "")
	add("id", k.ID, k.ID != "")
	add("format", k.Format, k.Format != "")
	add("statuses", texts(k.Statuses), len(k.Statuses) > 0)
	add("person", pairs(k.Person), len(k.Person) > 0)
	add("agent", pairs(k.Agent), len(k.Agent) > 0)
	add("stamp", k.Stamp, len(k.Stamp) > 0)
	add("task_field", k.TaskField, k.TaskField != "")
	return o
}

func texts(l []string) []any {
	out := make([]any, len(l))
	for i, s := range l {
		out[i] = s
	}
	return out
}

func pairs(l [][]string) []any {
	out := make([]any, len(l))
	for i, p := range l {
		out[i] = texts(p)
	}
	return out
}

// DocFiles lists the documents of a folder kind in the checkout at root, by project-relative path, sorted: each file
// at the folder's top level (contract §7.3: one file per document, top level only) named <id>.md or <id>-<slug>.md,
// its id matching the kind's id pattern (contract §4: a task file is named <id>-<slug>.md), or, for a kind with no id
// pattern, every .md file there. A file kind gives its file when it exists. A folder or file that is not there gives
// none; one that cannot be read is an error.
func DocFiles(root string, k DocKind) ([]string, error) {
	if k.File != "" {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(k.File)))
		if err != nil || !info.Mode().IsRegular() {
			if err != nil && !os.IsNotExist(err) {
				return nil, err
			}
			return nil, nil
		}
		return []string{k.File}, nil
	}
	if k.Path == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(k.Path)))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var idp *regexp.Regexp
	if k.ID != "" {
		if idp, err = regexp.Compile(k.ID); err != nil {
			return nil, err
		}
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if !e.Type().IsRegular() || !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, ".") {
			continue
		}
		if idp != nil && NameID(name, idp) == "" {
			continue
		}
		out = append(out, strings.TrimSuffix(k.Path, "/")+"/"+name)
	}
	sort.Strings(out)
	return out, nil
}

// NameID is the id a document's file name gives, <id>.md or <id>-<slug>.md, by the kind's id pattern: the shortest
// part of the name before .md, ending at a dash or at .md, that the pattern matches whole; "" for none.
func NameID(name string, idp *regexp.Regexp) string {
	stem := strings.TrimSuffix(name, ".md")
	for i := 0; i <= len(stem); i++ {
		if i < len(stem) && stem[i] != '-' {
			continue
		}
		if id := stem[:i]; id != "" {
			if loc := idp.FindStringIndex(id); loc != nil && loc[0] == 0 && loc[1] == len(id) {
				return id
			}
		}
	}
	return ""
}
