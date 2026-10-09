// Package status builds bonsai status's document, bonsai.status/1 (contract §12): one workspace at a glance.
//
// Plan part 2 builds the partial status: every field of bonsai.status/1, in the schema's order (the schema in
// formats/schemas, embedded by package formats, is the one list of fields), filled where the walking skeleton has
// built it and null (or [] for a list) where it has not. Built here: format, bonsai, mode, workspace, home, local,
// person_only, and from part 3's check (internal/engine) packs, files and problems: the same findings bonsai check
// reports (contract §12); from step 5.1.4a, formats (internal/format's registry: every format with the majors this
// Bonsai reads and writes); from step 5.1.5, internal/workspace's reads: documents (the declared document kinds),
// labels and lanes (from the lock's declares and the machine folder), status_writes and status_command (this
// machine's settings), and active_task (the active-task function, contract §13, for this checkout and environment).
// The rest (needs, checks) wait for step 5.1.6; this package's test holds them in a named list. The document is held
// to the schema before it is printed (Encode).
//
// Exit codes (contract §12, spec §3): 0, or 3 when Bonsai cannot read the workspace at all; the document then
// fills format, bonsai, problems and error (step 5.1.4b: the error object, with its word from format.ErrorWords), and
// every other field is null.
package status

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Format is the document's format.
const Format = "bonsai.status/1"

// Mode is how this status is gathered: offline, cheap, no network (contract §12).
const Mode = "offline"

var (
	schemaOnce   sync.Once
	statusSchema schema.Object
)

// Schema is bonsai.status/1's schema, from formats/schemas (embedded).
func Schema() schema.Object {
	schemaOnce.Do(func() {
		raw, err := formats.Schema("status")
		if err != nil {
			panic(err)
		}
		if statusSchema, err = schema.Parse(raw); err != nil {
			panic(err)
		}
	})
	return statusSchema
}

// Exit codes.
const (
	ExitOK      = 0
	ExitRuntime = 3 // Bonsai cannot read the workspace at all (contract §12)
)

// Build gathers the status of the workspace holding dir; version is this Bonsai's own. It returns the document and
// its exit code. On exit 3 the document is Refused's, its problem the error's sentence with its next step.
func Build(dir, version string) (schema.Object, int) {
	built, problem := gather(dir)
	if problem != nil {
		return Refused(version, problem, []any{problem.Error()}), ExitRuntime
	}
	built["format"], built["bonsai"] = Format, version
	return document(built, false), ExitOK
}

// Refused is the document of a status that refused or failed (bonsai.status/1): format, bonsai, problems and the
// error object (spec §3, §16 row 29), every other field null. problems holds the error's sentence when Bonsai cannot
// read the workspace (exit 3), and nothing when the command line was refused (exit 2).
func Refused(version string, e *engine.Error, problems []any) schema.Object {
	errDoc, err := format.MustLookup("error").Document(e.Object())
	if err != nil {
		panic(err) // the error object's Go type is held to its schema by internal/format's tests
	}
	return document(map[string]any{"format": Format, "bonsai": version, "problems": problems, "error": errDoc}, true)
}

// gather reads what part 2 builds. A problem is an error with its word, its sentence and its next step.
func gather(dir string) (map[string]any, *engine.Error) {
	co, err := workspace.Find(dir)
	if err != nil {
		return nil, engine.FindError(err)
	}
	cfg, err := workspace.LoadConfigFull(co.Root)
	if err != nil {
		return nil, engine.ConfigError(err)
	}
	home, err := workspace.Home()
	if err != nil {
		return nil, engine.HomeError(err)
	}
	key, err := workspace.MachineKey(co.Main)
	if err != nil {
		return nil, &engine.Error{Code: "read-failed", Exit: ExitRuntime,
			What: fmt.Sprintf("the main checkout %s cannot be resolved: %v", filepath.ToSlash(co.Main), err), Next: "check that it exists"}
	}
	root := filepath.ToSlash(co.Main)
	packs, files, problems := []any{}, schema.Object{{Key: "changed", Value: 0}, {Key: "missing", Value: 0},
		{Key: "format0_changed", Value: 0}}, []any{}
	var lock *workspace.Lock
	if r, err := engine.Check(dir, home); err == nil {
		lock = r.Lock
		for _, p := range r.Packs {
			packs = append(packs, schema.Object{{Key: "id", Value: p.ID}, {Key: "version", Value: p.Version},
				{Key: "commit", Value: p.Commit}, {Key: "state", Value: p.State}})
		}
		files = schema.Object{{Key: "changed", Value: r.Changed}, {Key: "missing", Value: r.Missing},
			{Key: "format0_changed", Value: r.Format0Changed}}
		for _, f := range r.Findings {
			problems = append(problems, f.Sentence())
		}
	}
	personOnly := make([]any, len(cfg.PersonOnly))
	for i, g := range cfg.PersonOnly {
		personOnly[i] = g
	}
	known := knownFields(co, cfg, lock, home)
	problems = append(problems, known.problems...)
	return map[string]any{
		"documents":      known.documents,
		"labels":         known.labels,
		"lanes":          known.lanes,
		"status_writes":  known.statusWrites,
		"status_command": known.statusCommand,
		"active_task":    known.active,
		"mode":           Mode,
		"formats":        format.StatusFormats(),
		"workspace": schema.Object{
			{Key: "id", Value: cfg.ID}, {Key: "name", Value: cfg.Name}, {Key: "root", Value: root},
		},
		"home": schema.Object{{Key: "path", Value: filepath.ToSlash(home)}, {Key: "key", Value: key}},
		"local": schema.Object{
			{Key: "log", Value: root + "/.bonsai/local/log"},
			{Key: "asks", Value: root + "/.bonsai/local/asks"},
			{Key: "ladder", Value: root + "/.bonsai/local/ladder"},
		},
		"packs":       packs,
		"files":       files,
		"person_only": personOnly,
		"problems":    problems,
	}, nil
}

// known is what step 5.1.5's reads give status: each field's value, and the problems they found.
type known struct {
	documents, labels, lanes    []any
	statusWrites, statusCommand any
	active                      schema.Object
	problems                    []any
}

// knownFields reads the document kinds, labels and lanes in force, this machine's settings and the active task. A
// read that fails is a problem (its sentence names the next step), and its field shows what can still be read.
func knownFields(co *workspace.Checkout, cfg *workspace.Config, lock *workspace.Lock, home string) known {
	k := known{documents: []any{}, labels: []any{}, lanes: []any{}, statusWrites: workspace.StatusWritesAgents}
	problem := func(err error) { k.problems = append(k.problems, ascii(err.Error())) }
	kinds, err := workspace.DocKinds(cfg.Full, lock)
	if err != nil {
		problem(err)
		kinds, _ = workspace.DocKinds(cfg.Full, nil)
	}
	for _, d := range kinds {
		k.documents = append(k.documents, d.Object())
	}
	sets, problems := workspace.LabelsInForce(lock, home, co.Main)
	for _, err := range problems {
		problem(err)
	}
	for _, l := range sets {
		k.labels = append(k.labels, schema.Object{{Key: "namespace", Value: l.Namespace}, {Key: "from", Value: l.From},
			{Key: "version", Value: l.Version}})
	}
	if lanes, err := workspace.Lanes(lock); err != nil {
		problem(err)
	} else {
		for _, l := range lanes {
			k.lanes = append(k.lanes, schema.Object{{Key: "name", Value: l.Name}, {Key: "approve_first", Value: l.ApproveFirst},
				{Key: "close", Value: l.Close}, {Key: "from", Value: l.From}})
		}
	}
	settings, err := workspace.LoadMachineSettings(home, co.Main)
	if err != nil {
		problem(err)
	}
	k.statusWrites = settings.StatusWrites
	if settings.StatusCommand != "" {
		k.statusCommand = settings.StatusCommand
	}
	k.active = activeTask(co, cfg).Object()
	return k
}

// activeTask is the active task for this checkout and the process's environment (contract §13): tasks read from
// the main checkout's task folder (its bonsai.yaml's documents.task; the checkout's own when main's cannot be read),
// BONSAI_TASK counting only when this project is the session's own (the project holding CLAUDE_PROJECT_DIR, when
// Claude Code set it). A task folder that cannot be read gives none, saying why.
func activeTask(co *workspace.Checkout, cfg *workspace.Config) workspace.ActiveTask {
	dir := cfg.Full.Documents.Task
	if co.Main != co.Root {
		if main, err := workspace.LoadConfigFull(co.Main); err == nil {
			dir = main.Full.Documents.Task
		}
	}
	counts := true
	if session := os.Getenv(sessionEnv); session != "" {
		s, err := workspace.Find(session)
		counts = err == nil && samePlace(s.Main, co.Main)
	}
	a, err := workspace.Active(workspace.ActiveInput{Main: co.Main, TaskDir: dir, Env: os.Getenv(workspace.TaskEnv), EnvCounts: counts})
	if err != nil {
		return workspace.ActiveTask{Why: "unreadable", Detail: ascii(err.Error())}
	}
	return a
}

// sessionEnv names the variable Claude Code sets to the session's project folder.
const sessionEnv = "CLAUDE_PROJECT_DIR"

func samePlace(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

// document writes every field of the schema, in its order: a built field's value, else [] for a field whose type
// allows a list and null for any other (contract §2.2: a writer writes every field, null or [] where it does not
// apply). On exit 3 every field not built is null.
func document(built map[string]any, failed bool) schema.Object {
	props, _ := Schema().Get("properties")
	doc := schema.Object{}
	for _, p := range props.(schema.Object) {
		v, ok := built[p.Key]
		if !ok && !failed && allowsList(p.Value.(schema.Object)) {
			v = []any{}
		}
		doc = append(doc, schema.Member{Key: p.Key, Value: v})
	}
	return doc
}

func allowsList(prop schema.Object) bool {
	t, _ := prop.Get("type")
	switch x := t.(type) {
	case string:
		return x == "array"
	case []any:
		for _, e := range x {
			if e == "array" {
				return true
			}
		}
	}
	return false
}

// Text renders the document for a person: plain ASCII, the same facts as the JSON (spec §6's closing words of
// init, which status shows too). On exit 3 it gives the problems alone.
func Text(doc schema.Object) string {
	var b strings.Builder
	problems, _ := doc.Get("problems")
	ws, _ := doc.Get("workspace")
	w, ok := ws.(schema.Object)
	if !ok {
		b.WriteString("bonsai status: cannot read this workspace.\n")
		for _, p := range problems.([]any) {
			b.WriteString("  " + ascii(p.(string)) + "\n")
		}
		return b.String()
	}
	hv, _ := doc.Get("home")
	home := hv.(schema.Object)
	fmt.Fprintf(&b, "Workspace %s, id %s (every clone and worktree of this project shares it).\n",
		ascii(w.String("name")), ascii(w.String("id")))
	fmt.Fprintf(&b, "In this project, %s:\n", ascii(w.String("root")))
	b.WriteString("  bonsai.yaml      this project's Bonsai settings. You edit it; it is committed.\n")
	b.WriteString("  .bonsai/         the lock, STATE and two tables Bonsai rebuilds (tasks; sessions and hours). Committed.\n")
	b.WriteString("  .bonsai/local/   the log, questions for you and their answers, ladder results. Never committed.\n")
	b.WriteString("On this machine:\n")
	fmt.Fprintf(&b, "  %s\n", ascii(home.String("path")))
	b.WriteString("    Bonsai's home: a secret salt, this machine's settings for each project, label files attached\n")
	b.WriteString("    on this machine, your personal memory, the pack cache. None of it enters git.\n")
	fmt.Fprintf(&b, "    This project's machine folder: workspaces/%s\n", ascii(home.String("key")))
	po, _ := doc.Get("person_only")
	var globs []string
	for _, g := range po.([]any) {
		globs = append(globs, ascii(g.(string)))
	}
	if len(globs) == 0 {
		globs = []string{"none"}
	}
	fmt.Fprintf(&b, "Person-only paths: %s\n", strings.Join(globs, ", "))
	pv, _ := doc.Get("packs")
	var packs []string
	for _, p := range pv.([]any) {
		po := p.(schema.Object)
		commit := po.String("commit")
		if len(commit) > 7 {
			commit = commit[:7]
		}
		packs = append(packs, ascii(po.String("id")+" "+po.String("version")+" at "+commit+" ("+po.String("state")+")"))
	}
	if len(packs) == 0 {
		packs = []string{"none locked"}
	}
	fmt.Fprintf(&b, "Packs: %s\n", strings.Join(packs, ", "))
	list := problems.([]any)
	if len(list) == 0 {
		b.WriteString("Problems: none.\n")
	} else {
		b.WriteString("Problems:\n")
		for _, p := range list {
			b.WriteString("  " + ascii(p.(string)) + "\n")
		}
	}
	if at, ok := doc.Get("active_task"); ok {
		if a, ok := at.(schema.Object); ok {
			if id := a.String("id"); id != "" {
				fmt.Fprintf(&b, "Active task: %s (%s)\n", ascii(id), ascii(a.String("how")))
			} else {
				fmt.Fprintf(&b, "Active task: none (%s)\n", ascii(strings.TrimPrefix(a.String("why"), "no active task: ")))
			}
		}
	}
	if sw, _ := doc.Get("status_writes"); sw != nil {
		line := "agents move task statuses themselves"
		if sw == workspace.StatusWritesCommand {
			sc, _ := doc.Get("status_command")
			line = "moves go through " + schema.Show(sc) + " (a managed workspace)"
		}
		fmt.Fprintf(&b, "Status writes: %s (%s)\n", ascii(fmt.Sprint(sw)), ascii(line))
	}
	b.WriteString("A copy meant as a new project needs its own id: bonsai init --new-id\n")
	b.WriteString("Every field, documents, labels and lanes among them: bonsai status --json\n")
	return b.String()
}

// ascii keeps printable ASCII and writes any other character as a Go escape, so the text prints unbroken in
// PowerShell 5.1 (spec §3).
func ascii(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 0x20 && r < 0x7f {
			b.WriteRune(r)
			continue
		}
		q := fmt.Sprintf("%+q", string(r))
		b.WriteString(q[1 : len(q)-1])
	}
	return b.String()
}

// Encode writes the document as bonsai.status/1's writer does: held to the schema, every field in its order, then
// encoded byte-stable (internal/format).
func Encode(doc schema.Object) ([]byte, error) {
	return format.MustLookup("status").Encode(doc)
}
