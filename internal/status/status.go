// Package status builds bonsai status's document, bonsai.status/1 (contract §12): one workspace at a glance.
//
// Every field of bonsai.status/1, in the schema's order (the schema in formats/schemas, embedded by package formats,
// is the one list of fields), each filled where it applies and null or [] only where contract §12 says it does not
// (this package's test holds that list, doesNotApply). What fills each:
//
//   - format, bonsai, mode (offline; full with --full), workspace, home, local, person_only: bonsai.yaml, the home and
//     git (plan part 2);
//   - packs, files, problems: bonsai check's own findings (internal/engine's Check, offline from the lock), problems
//     one sentence each with its next step (contract §12: "the same findings bonsai check reports"); check's warnings
//     are never problems (spec §6). Two need more than the default may spend, so --full alone adds them: approve_first
//     from git history (every commit that touched the task folder) and the plugin drift Claude Code reports. The
//     default does look for Git Bash on Windows (check's git-bash: a few file reads, Options.GitBash);
//   - formats: internal/format's registry (step 5.1.4a);
//   - documents, labels, lanes, status_writes, status_command, active_task: internal/workspace's reads (step 5.1.5);
//   - needs (step 5.1.6): the Claude Code floor (kind tool, name claude-code, the higher of Bonsai's and the packs'
//     needs.claude_code: never a problem, spec §7), then each locked pack (kind pack, with its source and version), or,
//     with --full, kind plugin for a pack whose plugin Claude Code reports not installed for this checkout (spec §5);
//   - checks (--full only, step 5.1.6): newer releases of each pack (git ls-remote, engine.NewerTags), each pack's
//     plugin as Claude Code reports it (claude plugin list), Claude Code's version against the floor (claude --version),
//     and the MCP servers the packs' needs name (none can be named in formats set 4: bonsai.pack/1's needs holds only
//     claude_code, so the list is empty until a later set adds one);
//   - ladder: the active task's last ladder result, stale when it proves another commit than HEAD (lastLadder, step
//     5.4.2);
//   - roles (set 8; step 5.5.0): each role the linked packs declare, with each limit's default and max in force and
//     where each came from (rolesOf: workspace.RolesInForce, the one function every reader of a role's limits uses),
//     from the lock and bonsai.yaml as committed at the main checkout's HEAD, as the guard reads them;
//   - messages (set 8; step 5.5.10): the threads of messages between agents, counted open and flagged (messagesOf:
//     internal/message's reader and its folding, the flags' thresholds from bonsai.yaml as committed). A committed
//     bonsai.yaml that does not read is a problem naming it, roles [] and messages null: neither is read from this
//     checkout's copy in its place (message.CommittedConfig, the reading bonsai message shares);
//   - error: null unless status refused (step 5.1.4b).
//
// The default is cheap and offline (contract §12: for the bridge to run often): it runs git and reads files, and
// never Claude Code or the network; --full asks both. status --active prints active_task alone (Active; contract §13's
// read-only command). The document is held to the schema before it is printed (Encode).
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
	"time"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/stopgate"
	"github.com/LastStep/Bonsai/internal/textui"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Format is the document's format.
const Format = "bonsai.status/1"

// The modes, how a status was gathered (contract §12): offline, cheap, no network and no Claude Code, by default;
// full with --full, which asks both (the schema keeps the field open for it).
const (
	Mode     = "offline"
	ModeFull = "full"
)

// Options are what status is asked beyond the default, and how --full asks the network and Claude Code (tests give
// fakes; nil asks nothing, and its check reads unknown).
type Options struct {
	Full    bool                                  // --full: the checks
	Plugins engine.PluginCLI                      // claude plugin list
	Claude  func() (string, error)                // claude --version's first line
	Tags    func(source string) ([]string, error) // a source's tags (engine.RemoteTags when nil)
	GitBash func() engine.GitBash                 // how check looks for Git Bash (engine.FindGitBash when nil)
}

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

// Build gathers the default status of the workspace holding dir; version is this Bonsai's own. It returns the
// document and its exit code. On exit 3 the document is Refused's, its problem the error's sentence with its next step.
func Build(dir, version string) (schema.Object, int) { return BuildWith(dir, version, Options{}) }

// BuildWith is Build with its options: --full's checks.
func BuildWith(dir, version string, o Options) (schema.Object, int) {
	built, problem := gather(dir, o)
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

// gather reads every field. A problem is an error with its word, its sentence and its next step.
func gather(dir string, o Options) (map[string]any, *engine.Error) {
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
	var checked *engine.CheckResult
	if r, err := engine.CheckWith(dir, home, engine.CheckOptions{History: o.Full, GitBash: o.GitBash}); err == nil {
		checked = r
		lock = r.Lock
		if o.Full {
			// Plugin drift that only Claude Code can report is a problem too (bonsai.status/1), asked with --full only:
			// the default never runs Claude Code.
			engine.ComparePlugins(r, o.Plugins)
		}
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
	mode, checks := Mode, any(nil)
	var installed map[string]bool
	if o.Full {
		mode = ModeFull
		var c schema.Object
		c, installed = fullChecks(checked, cfg, lock, o)
		checks = c
	}
	roles, messages := any([]any{}), any(nil)
	committed, problem := committedConfig(co, cfg) // one git call, for the roles and the messages' thresholds
	if problem != "" {
		problems = append(problems, problem) // neither the roles nor the flags can be read: [] and null
	} else {
		var rolesProblem, messagesProblem string
		if roles, rolesProblem = rolesOf(co, committed, lock); rolesProblem != "" {
			problems = append(problems, rolesProblem)
		}
		if messages, messagesProblem = messagesOf(co, cfg, committed.Full); messagesProblem != "" {
			problems = append(problems, messagesProblem)
		}
	}
	return map[string]any{
		"needs":          needsOf(lock, installed),
		"checks":         checks,
		"documents":      known.documents,
		"labels":         known.labels,
		"lanes":          known.lanes,
		"status_writes":  known.statusWrites,
		"status_command": known.statusCommand,
		"active_task":    known.active,
		"mode":           mode,
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
		"ladder":      lastLadder(co, known.activeID),
		"roles":       roles,
		"messages":    messages,
		// Bonsai's newer releases (set 10): null until step 5.6.5 reads them (plan-5 note 5.6.0, 3).
		"bonsai_release": nil,
	}, nil
}

// lastLadder is the active task's last ladder result (bonsai.status/1's ladder, set 7; plan-5 5.4.2 note 12), so an
// agent reads where its proof stands without a climb: the result at the main checkout's .bonsai/local/ladder/<task>.json
// that the runner wrote, its green, mode, commit and finished, and stale when that commit is not this checkout's HEAD
// (read from its git files, as the stop gate reads it: stopgate.Head). null when there is no active task, no result,
// or a result that does not read as bonsai.ladder/1 or was made for another task.
func lastLadder(co *workspace.Checkout, task string) any {
	if task == "" {
		return nil
	}
	rel := workspace.LocalDir + "/ladder/" + task + ".json"
	raw, err := os.ReadFile(filepath.Join(co.Main, filepath.FromSlash(rel)))
	if err != nil {
		return nil
	}
	l, err := format.ReadLadder(raw)
	if err != nil || l.Task == nil || *l.Task != task {
		return nil
	}
	head, err := stopgate.Head(workspace.ReadGitLayout(co.Root))
	stale := err != nil || !strings.EqualFold(l.Git.SHA, head)
	return schema.Object{{Key: "task", Value: task}, {Key: "path", Value: rel}, {Key: "green", Value: l.Green},
		{Key: "mode", Value: l.Mode}, {Key: "sha", Value: l.Git.SHA}, {Key: "finished", Value: l.Finished}, {Key: "stale", Value: stale}}
}

// rolesOf is status --json's roles (bonsai.status/1, set 8; plan-5 5.5.0 note 10): each role the lock's packs declare,
// in the lock's order, its pack, name, kind and writes, and each of its limits' default and max in force with where
// each came from (pack or bonsai.yaml), from message.RolesInForce (workspace.RolesInForce, the one function every reader
// of a role's limits uses). It reads the lock and bonsai.yaml (w, committedConfig's) as committed at the main
// checkout's HEAD, as the guard reads them (message.CommittedLock, the reading bonsai message shares), so an
// uncommitted change shows only once committed; a file HEAD does not hold (a project not yet committed, a lock not yet
// committed) is read as it stands in this checkout (lock, check's). [] when no linked pack declares a role. A
// committed lock that does not read, or roles whose lock or bonsai.yaml does not read for them, give [] and a problem
// naming that file (bonsai.yaml with its line) and the next step; never the lock for bonsai.yaml's.
func rolesOf(co *workspace.Checkout, w *workspace.Config, lock *workspace.Lock) (any, string) {
	l, held, e := message.CommittedLock(co.Main)
	switch {
	case e != nil:
		return []any{}, problemOf(e)
	case held:
		lock = l
	}
	in, e := message.RolesInForce(lock, w)
	if e != nil {
		return []any{}, problemOf(e)
	}
	out := []any{}
	limit := func(l workspace.LimitInForce) schema.Object {
		n := func(p *int64) any {
			if p == nil {
				return nil
			}
			return *p
		}
		return schema.Object{{Key: "default", Value: n(l.Default)}, {Key: "default_from", Value: l.DefaultFrom},
			{Key: "max", Value: n(l.Max)}, {Key: "max_from", Value: l.MaxFrom}}
	}
	for _, r := range in {
		out = append(out, schema.Object{{Key: "pack", Value: r.Pack}, {Key: "name", Value: r.Role.Name}, {Key: "kind", Value: r.Role.Kind},
			{Key: "writes", Value: r.Role.Writes}, {Key: "minutes", Value: limit(r.Minutes)}, {Key: "actions", Value: limit(r.Actions)},
			{Key: "red_runs", Value: limit(r.RedRuns)}})
	}
	return out, ""
}

// committedConfig is bonsai.yaml as committed at the main checkout's HEAD, as the guard reads it
// (message.CommittedConfig, the reading bonsai message shares); where HEAD holds none, the file as this checkout read
// it (cfg). A committed copy that does not read, or a HEAD git cannot read, gives nil and a problem naming it, never
// this checkout's copy in its place.
func committedConfig(co *workspace.Checkout, cfg *workspace.Config) (*workspace.Config, string) {
	c, held, e := message.CommittedConfig(co.Main)
	switch {
	case e != nil:
		return nil, problemOf(e)
	case !held:
		return cfg, ""
	}
	return c, ""
}

// problemOf is a refusal of internal/message's as a problem's sentence: what is wrong, then its next step.
func problemOf(e *message.Error) string {
	return strings.TrimSuffix(e.What, ".") + "; next: " + e.Next
}

// messagesNow is the clock the flags' ages are read from; a test sets it.
var messagesNow = time.Now

// messagesOf is status --json's messages (bonsai.status/1, set 8; plan-5 5.5.10 note 5): the threads of messages
// between agents counted, open (a brief with no report after it) and flagged (any of bonsai messages' flags), as
// bonsai messages folds them (internal/message, the one reader), with the flags' thresholds from bonsai.yaml as
// committed (w, committedConfig's; workspace.MessageThreshold), as the guard and bonsai messages read them, and the
// two flags that read the agents' state (budget-spent; unanswered by the tied agent's last call) from the guard's
// records, through internal/agents' Store.Thread, as bonsai messages reads them. The
// messages are where every writer of .bonsai/local/ writes (workspace.FindLocal). A messages folder that cannot be
// listed gives null and a problem naming the next step; a file that does not read is check's warning, never a problem.
func messagesOf(co *workspace.Checkout, cfg *workspace.Config, w *format.Workspace) (any, string) {
	main := workspace.FindLocal(co.Root, cfg.ID).Main
	st, err := message.Read(main)
	if err != nil {
		return nil, message.Dir + " cannot be listed: " + ascii(err.Error()) + "; next: check that it can be read, then run: bonsai messages"
	}
	r := message.DefaultRules()
	r.UnansweredMinutes, _ = workspace.MessageThreshold(w, "unanswered_minutes")
	r.CirclingBriefs, _ = workspace.MessageThreshold(w, "circling_briefs")
	var open, flagged int64
	now, store := messagesNow(), agents.Store{Main: main}
	for _, t := range message.Fold(st, r, now) {
		store.Thread(&t, r.UnansweredMinutes, now) // the flags that read the agents' state, as bonsai messages reads them
		if t.State == message.Open {
			open++
		}
		if len(t.Flags) > 0 {
			flagged++
		}
	}
	return schema.Object{{Key: "open", Value: open}, {Key: "flagged", Value: flagged}}, ""
}

// known is what step 5.1.5's reads give status: each field's value, and the problems they found.
type known struct {
	documents, labels, lanes    []any
	statusWrites, statusCommand any
	active                      schema.Object
	activeID                    string // the active task's id, "" for none
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
	a := activeTask(co, cfg)
	k.active, k.activeID = a.Object(), a.ID
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

// Text renders the document for a person, drawn for a pipe (TextTo).
func Text(doc schema.Object) string {
	var d textui.Doc
	TextTo(&d, doc)
	return d.String()
}

// TextTo adds the document for a person: the same facts as the JSON (spec §6's closing words of init, which status
// shows too), what needs a look first. On exit 3 it gives the problems alone.
func TextTo(d *textui.Doc, doc schema.Object) {
	problems, _ := doc.Get("problems")
	ws, _ := doc.Get("workspace")
	w, ok := ws.(schema.Object)
	if !ok {
		d.Title("bonsai status", "cannot read this workspace")
		var rows []textui.Row
		for _, p := range problems.([]any) {
			rows = append(rows, textui.R(textui.T(ascii(p.(string)), textui.Bad)))
		}
		if len(rows) > 0 {
			d.Section("Problems", rows...)
		}
		return
	}
	note := ""
	if cv, _ := doc.Get("checks"); cv != nil {
		note = "with the machine's checks (--full)"
	}
	d.Title("bonsai status", note)
	d.Section("Workspace", textui.R(textui.T(ascii(w.String("name")), textui.Key), textui.T(ascii(w.String("id")), textui.Path),
		textui.T("every clone and worktree of this project shares this id", textui.Dim)))
	list := problems.([]any)
	if len(list) == 0 {
		d.Section("Problems", textui.R(textui.T("none", textui.Ok)))
	} else {
		var rows []textui.Row
		for _, p := range list {
			rows = append(rows, textui.R(textui.T(ascii(p.(string)), textui.Bad)))
		}
		d.Section("Problems", rows...)
	}
	if at, ok := doc.Get("active_task"); ok {
		if a, ok := at.(schema.Object); ok {
			if id := a.String("id"); id != "" {
				d.Section("Active task", textui.R(textui.T(ascii(id), textui.Path), textui.C(ascii(a.String("how")))))
			} else {
				d.Section("Active task", textui.R(textui.T("none", textui.Dim), textui.C(ascii(strings.TrimPrefix(a.String("why"), "no active task: ")))))
			}
		}
	}
	if lv, _ := doc.Get("ladder"); lv != nil {
		l := lv.(schema.Object)
		green, tone := "not green", textui.Bad
		if g, _ := l.Get("green"); g == true {
			green, tone = "green", textui.Ok
		}
		sha := l.String("sha")
		if len(sha) > 12 {
			sha = sha[:12]
		}
		rows := []textui.Row{textui.R(textui.T(green, tone), textui.C(fmt.Sprintf("at commit %s (%s), finished %s", ascii(sha), ascii(l.String("mode")),
			ascii(l.String("finished")))), textui.T(ascii(l.String("path")), textui.Path))}
		if s, _ := l.Get("stale"); s == true {
			rows = append(rows, textui.R(textui.T("stale", textui.Warn), textui.C("it proves another commit than HEAD, so climb again")))
		}
		d.Section("Last ladder", rows...)
	}
	d.Gap()
	pv, _ := doc.Get("packs")
	var packs []textui.Row
	for _, p := range pv.([]any) {
		po := p.(schema.Object)
		commit := po.String("commit")
		if len(commit) > 7 {
			commit = commit[:7]
		}
		state, tone := po.String("state"), textui.Warn
		if state == "ok" {
			tone = textui.Ok
		}
		packs = append(packs, textui.R(textui.T(ascii(po.String("id")), textui.Key), textui.C(ascii(po.String("version"))),
			textui.C("at "+ascii(commit)), textui.T(ascii(state), tone)))
	}
	if len(packs) == 0 {
		packs = []textui.Row{textui.R(textui.T("none locked", textui.Dim))}
	}
	d.Section("Packs", packs...)
	if nv, _ := doc.Get("needs"); nv != nil {
		var needs []textui.Row
		for _, n := range nv.([]any) {
			no := n.(schema.Object)
			switch no.String("kind") {
			case "tool":
				needs = append(needs, textui.R(textui.C(ascii(no.String("name"))), textui.C("tool"), textui.C(ascii(no.String("version")))))
			case "plugin":
				needs = append(needs, textui.R(textui.C(ascii(no.String("id"))), textui.C("plugin"), textui.C(ascii(no.String("version"))),
					textui.T("not installed here: bonsai update installs it", textui.Warn)))
			default:
				needs = append(needs, textui.R(textui.C(ascii(no.String("id"))), textui.C(ascii(no.String("kind"))), textui.C(ascii(no.String("version")))))
			}
		}
		if len(needs) > 0 {
			d.Section("Needs", needs...)
		}
	}
	if cv, _ := doc.Get("checks"); cv != nil {
		checksTo(d, cv.(schema.Object))
	}
	d.Gap()
	hv, _ := doc.Get("home")
	home := hv.(schema.Object)
	engine.WhereTo(d, ascii(w.String("root")))
	d.Section("On this machine", textui.R(textui.T(ascii(home.String("path")), textui.Path), textui.C("Bonsai's home: a secret salt, this "+
		"machine's settings for each project, label files attached on this machine, your personal memory, the pack cache. None of it enters git.")),
		textui.R(textui.T("workspaces/"+ascii(home.String("key")), textui.Path), textui.C("this project's machine folder")))
	po, _ := doc.Get("person_only")
	var globs []textui.Row
	for _, g := range po.([]any) {
		globs = append(globs, textui.R(textui.T(ascii(g.(string)), textui.Path)))
	}
	if len(globs) == 0 {
		globs = []textui.Row{textui.R(textui.T("none", textui.Dim))}
	}
	d.Section("Person-only", globs...)
	if sw, _ := doc.Get("status_writes"); sw != nil {
		line := "agents move task statuses themselves"
		if sw == workspace.StatusWritesCommand {
			sc, _ := doc.Get("status_command")
			line = "moves go through " + schema.Show(sc) + " (a managed workspace)"
		}
		d.Section("Status writes", textui.R(textui.T(ascii(fmt.Sprint(sw)), textui.Key), textui.C(ascii(line))))
	}
	d.Gap()
	d.Hint("every field, documents, labels and lanes among them: bonsai status --json")
	d.Hint("a copy meant as a new project needs its own id: bonsai init --new-id")
}

// checksTo adds --full's checks for a person.
func checksTo(d *textui.Doc, c schema.Object) {
	var rows []textui.Row
	pv, _ := c.Get("packs")
	for _, p := range pv.([]any) {
		po := p.(schema.Object)
		line, tone := "none newer", textui.Ok
		if why := po.String("why"); why != "" {
			line, tone = "newer releases unknown: "+why, textui.Warn
		} else if nv, _ := po.Get("newer"); len(nv.([]any)) > 0 {
			var tags []string
			for _, t := range nv.([]any) {
				tags = append(tags, t.(string))
			}
			line, tone = "newer: "+strings.Join(tags, ", "), textui.Warn
		}
		rows = append(rows, textui.R(textui.C("pack"), textui.C(ascii(po.String("id")+" "+po.String("version"))), textui.T(ascii(line), tone)))
	}
	plv, _ := c.Get("plugins")
	for _, p := range plv.([]any) {
		po := p.(schema.Object)
		state, tone := "unknown: "+po.String("why"), textui.Warn
		switch v, _ := po.Get("installed"); v {
		case true:
			state, tone = "installed for this checkout", textui.Ok
		case false:
			state = "not installed for this checkout (bonsai update installs it)"
		}
		rows = append(rows, textui.R(textui.C("plugin"), textui.C(ascii(po.String("plugin"))), textui.T(ascii(state), tone)))
	}
	cc, _ := c.Get("claude_code")
	co := cc.(schema.Object)
	line := co.String("state")
	tone := textui.Warn
	if line == "ok" {
		tone = textui.Ok
	}
	if v := co.String("version"); v != "" {
		line = v + ", " + line
	}
	if why := co.String("why"); why != "" {
		line += ": " + why
	}
	rows = append(rows, textui.R(textui.C("tool"), textui.C("Claude Code"), textui.T(ascii(line)+" (the floor "+ascii(co.String("floor"))+", from "+
		ascii(co.String("from"))+")", tone)))
	mv, _ := c.Get("mcp")
	if len(mv.([]any)) == 0 {
		rows = append(rows, textui.R(textui.C("mcp"), textui.C("MCP servers"), textui.T("none named by the packs' needs", textui.Dim)))
	}
	d.Section("Checks", rows...)
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
