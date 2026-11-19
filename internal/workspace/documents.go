package workspace

// Document kinds (contract §7.3): every kind of document a workspace holds is declared, by Bonsai for its own kinds
// and by the packs for theirs, so no reader hard-codes a project's folders. Bonsai's own kinds and their places are
// the table bonsaiKinds below, their one home: task, run, state, answers, memory, the two generated tables tasks
// and sessions, and (set 14) bug, one file per bug, with bugs, the page generated from them. A pack's kinds come from the lock's declares (format.Declares), each at the place bonsai.yaml's
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
// a fixed file), its format, its id pattern, and for a kind with statuses (task, bug) where they come from, the moves
// agents make and the stamp.
type bonsaiKind struct {
	Kind     string
	Key      string // bonsai.yaml's documents key naming its place; "" for a fixed file
	IsFile   bool   // the place is one file (answers, the bugs page), not a folder
	Fixed    string // its fixed file (STATE and the two tables, contract §3), "" when bonsai.yaml names it
	Format   string // its format's short name, "" for none (answers holds answers, not a Bonsai format)
	ID       string // its id pattern (the format's schema's, held equal by TestBonsaiKindsAreTheSchemas)
	Statuses bool   // its statuses are its format's status enum (its one home), the first where a new one starts
	Agent    [][]string
	Stamp    schema.Object
}

// The id patterns of Bonsai's kinds that have ids, as their schemas hold them (formats/schemas, their one home): kept
// here as text so the task folder's reader (active.go, which the guard uses from step 5.3) parses no schema, and held
// equal to the schemas by a test.
const (
	TaskIDPattern   = `^T-[0-9]{4,6}$`
	BugIDPattern    = `^BUG-[0-9]{3,6}$`
	runIDPattern    = `^R-[0-9]{4}-[0-9]{2}-[0-9]{2}-T-[0-9]{4,6}`
	memoryIDPattern = `^M-[a-z0-9][a-z0-9-]*$`
)

// bonsaiKinds are Bonsai's own document kinds, in contract §7.3's order: their one home.
var bonsaiKinds = []bonsaiKind{
	{Kind: "task", Key: "task", Format: "task", ID: TaskIDPattern, Statuses: true, Stamp: taskStamp},
	{Kind: "run", Key: "run", Format: "run", ID: runIDPattern},
	{Kind: "state", Fixed: StateFile, Format: "state"},
	{Kind: "answers", Key: "answers", IsFile: true},
	{Kind: "memory", Key: "memory", Format: "memory", ID: memoryIDPattern},
	{Kind: "tasks", Fixed: TasksTableFile, Format: "tasks"},
	{Kind: "sessions", Fixed: SessionsTableFile, Format: "sessions"},
	{Kind: "bug", Key: "bug", Format: "bug", ID: BugIDPattern, Statuses: true, Agent: bugMoves, Stamp: bugStamp},
	{Kind: "bugs", Key: "bugs", IsFile: true, Format: "bugs"},
}

// bugMoves are the bug kind's status moves, every one an agent's (plan P-T-6023, step 3; from Bonsai 1.1.0 made by
// bonsai bug alone): open to fixed, wontfix or watched; watched to open or fixed; fixed and wontfix back to open.
var bugMoves = [][]string{{"open", "fixed"}, {"open", "wontfix"}, {"open", "watched"}, {"watched", "open"},
	{"watched", "fixed"}, {"fixed", "open"}, {"wontfix", "open"}}

// bugStamp is the bug kind's stamp: closed is the day a bug was fixed or set wontfix.
var bugStamp = schema.Object{{Key: "fixed", Value: "closed"}, {Key: "wontfix", Value: "closed"}}

// The fixed files of Bonsai's kinds (contract §3, §7.2, §7.5).
const (
	StateFile         = ".bonsai/STATE.md"
	TasksTableFile    = ".bonsai/tasks.md"
	SessionsTableFile = ".bonsai/sessions.md"
)

// taskStamp is the task kind's stamp (contract §7.3: "started only when empty").
var taskStamp = schema.Object{{Key: "running", Value: "started"}, {Key: "done", Value: "finished"}, {Key: "cut", Value: "finished"}}

// BonsaiKind reports whether a kind's name is one of Bonsai's own, which no pack may declare (but SetAsideKind's).
func BonsaiKind(name string) bool {
	for _, k := range bonsaiKinds {
		if k.Kind == name {
			return true
		}
	}
	return false
}

// BonsaiDocKind gives one of Bonsai's own kinds as DocKinds gives it with no bonsai.yaml and no lock (its id pattern,
// statuses, agent moves and stamp: the bug kind's moves for bonsai bug); false for a name that is not Bonsai's.
func BonsaiDocKind(name string) (DocKind, bool) {
	kinds, err := DocKinds(&format.Workspace{}, nil)
	if err != nil {
		return DocKind{}, false
	}
	for _, k := range kinds {
		if k.Kind == name {
			return k, true
		}
	}
	return DocKind{}, false
}

// SetAsideKind reports whether a pack's document kind is set aside rather than refused: bugs, which the workflow pack
// declared (one file, its hand-kept register) until it became Bonsai's own in set 14. A pack that still declares it
// links and checks as before; its declaration is left out of the kinds in force (DocKinds, the guard), and bonsai
// check warns, naming bonsai pack set (plan P-T-6023, step 3). Any other of Bonsai's names a pack takes is refused.
func SetAsideKind(name string) bool { return name == "bugs" }

// BonsaiKindNames are Bonsai's own document kinds' names, in contract §7.3's order: the list the reference page of
// lists (step 5.1.10) prints.
func BonsaiKindNames() []string {
	out := make([]string, len(bonsaiKinds))
	for i, k := range bonsaiKinds {
		out[i] = k.Kind
	}
	return out
}

// Format0Kinds are the kinds bonsai check holds to format 0 (contract §2.3): Bonsai's own markdown kinds that had a
// format 0, task, run and state.
var Format0Kinds = []string{"task", "run", "state"}

// DocKinds lists the workspace's document kinds: Bonsai's (bonsaiKinds) at the places w's documents names, then
// each locked pack's from its declares, at the place w's documents names under the kind's name or else the pack's
// default. lock may be nil (no packs). Its error is a lock pack whose declares Bonsai does not read.
func DocKinds(w *format.Workspace, lock *Lock) ([]DocKind, error) {
	places := map[string]string{}
	for _, m := range w.Documents.Extra {
		if s, ok := m.Value.(string); ok {
			places[m.Key] = s
		}
	}
	for k, v := range map[string]string{"task": w.Documents.Task, "run": w.Documents.Run, "answers": w.Documents.Answers,
		"memory": w.Documents.Memory, "bug": w.Documents.BugFolder(), "bugs": w.Documents.BugsFile()} {
		places[k] = v
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
		if b.Statuses {
			k.Statuses = enumOf(b.Format)
		}
		k.Agent, k.Stamp = b.Agent, b.Stamp
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
			if SetAsideKind(pk.Kind) {
				continue // Bonsai's own from set 14 (SetAsideKind): listed once, as Bonsai's
			}
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

var statuses sync.Map // a format's short name -> its status enum, []string

// enumOf is a kind's closed list of statuses, from its format's schema (its one home: bonsai.task/1's, bonsai.bug/1's),
// read once a process.
func enumOf(short string) []string {
	if v, ok := statuses.Load(short); ok {
		return append([]string{}, v.([]string)...)
	}
	var out []string
	props, _ := format.MustLookup(short).Schema().Get("properties")
	st, _ := props.(schema.Object).Get("status")
	enum, _ := st.(schema.Object).Get("enum")
	for _, v := range enum.([]any) {
		out = append(out, v.(string))
	}
	statuses.Store(short, out)
	return append([]string{}, out...)
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

// LaneDef is one lane in force, and the pack that defines it (status --json's lanes).
type LaneDef struct {
	format.Lane
	From string // the pack's id
}

// Lanes lists the lanes the locked packs define (contract §6), from the lock's declares, in the lock's order. lock
// may be nil (none).
func Lanes(lock *Lock) ([]LaneDef, error) {
	var out []LaneDef
	if lock == nil {
		return out, nil
	}
	for _, lp := range lock.Packs {
		d, err := lp.Declared()
		if err != nil {
			return nil, lockError("the pack %s's %v", showValue(lp.ID), err)
		}
		for _, l := range d.Lanes {
			out = append(out, LaneDef{Lane: l, From: lp.ID})
		}
	}
	return out, nil
}
