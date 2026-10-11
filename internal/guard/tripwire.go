package guard

// command mode's tripwires (design/plan-5.md, 5.3.2 note 6; contract §10.1), only where the path's project reads
// command (this machine's status_writes for its main checkout, workspace.LoadMachineSettings): in a workspace the
// studio manages, status moves, lane changes, grants and outside labels go through its one write path (the
// workspace's status_command), so the guard refuses an agent's file-tool edit that would make one:
//
//   - task-outside-main: an edit of a task file outside the main checkout (a worktree's copy: tasks change on main
//     only, contract §10.1);
//   - status-line: an edit of a declared document's status line; a new document may be created at its kind's first
//     status only;
//   - lane-line: an edit of a task's lane line;
//   - grant-person-only: an edit that puts a person-only path into a task's bonsai.allows (bonsai.yaml's person_only
//     list and the floor's person-only paths; a glob that covers one, or lies inside one);
//   - outside-label: an edit that sets, changes or removes a label whose definition reads set_by: outside (contract
//     §5.2: a pack's, from the lock's declares, or one attached on this machine);
//   - bad-format: a write of a task, a run report or STATE that fails its format (one broken task file makes the
//     active task none for everyone, contract §13), read under format 0 or 1.
//
// Each names the workspace's status_command as the next step. The guard judges the file as it would be after the call:
// Write's content; Edit's and MultiEdit's replacements applied to the file on disk as Claude Code applies them (line
// endings read as LF, curly quotes matched as straight when the text holds none as written); an old_string not found,
// a NotebookEdit, or a delete or move (step 5.3.3) leaves what the file holds after it unknown, which refuses an
// existing document's status line (it cannot be shown to stay) and passes a new one (a shell write, question A).
//
// Only documents are read this way (documentOf: a task or run report in bonsai.yaml's task or run folder, named by
// its kind's id; .bonsai/STATE.md; a document of a kind a pack declares, at its place), so every other call pays
// nothing; the mode is read only for a document. The lock is read in full, and this machine's labels, only when a
// document's labels change, or a new document of a pack's kind needs its kind's first status.

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// docKind is the kind of a document the tripwires read.
type docKind struct {
	kind   string // task, run, state, or a pack's kind
	pack   string // the pack that declares it; "" for one of Bonsai's
	status bool   // its documents carry a status line that moves change (a task; a pack's kind)
}

// bonsaiDocKinds are Bonsai's own document kinds as workspace.DocKinds gives them (their one home: id patterns, the
// task's statuses), read once a process and only when a document is judged.
var bonsaiDocKinds = sync.OnceValue(func() map[string]workspace.DocKind {
	out := map[string]workspace.DocKind{}
	kinds, err := workspace.DocKinds(&format.Workspace{}, nil)
	if err != nil {
		return out
	}
	for _, k := range kinds {
		out[k.Kind] = k
	}
	return out
})

// idPatterns compiles each Bonsai kind's id pattern once.
var idPatterns = sync.OnceValue(func() map[string]*regexp.Regexp {
	out := map[string]*regexp.Regexp{}
	for name, k := range bonsaiDocKinds() {
		if k.ID != "" {
			if re, err := regexp.Compile(k.ID); err == nil {
				out[name] = re
			}
		}
	}
	return out
})

// documentOf says which document a judged path is in its project, nil for none.
func (j *judge) documentOf(x judged) *docKind {
	l := j.listsOf(x.p)
	if l.err != nil || l.cfg == nil {
		return nil
	}
	for _, kind := range []string{"task", "run"} {
		if dir := docPlace(l.cfg, kind); dir != "" && inFolder(x.rel, dir, x.fold) {
			if re := idPatterns()[kind]; re != nil && workspace.NameID(path.Base(x.rel), re) != "" {
				return &docKind{kind: kind, status: kind == "task"}
			}
		}
	}
	if same(x.rel, workspace.StateFile, x.fold) {
		return &docKind{kind: "state"}
	}
	for _, d := range l.declared {
		for _, dp := range d.Documents {
			place := docPlace(l.cfg, dp.Kind)
			if place == "" {
				place = strings.TrimSuffix(dp.Path, "/")
			}
			if same(x.rel, place, x.fold) || inFolder(x.rel, place, x.fold) {
				return &docKind{kind: dp.Kind, pack: d.Pack, status: true}
			}
		}
	}
	return nil
}

// inFolder reports a markdown file at the top level of a folder (contract §7.3: one file per document, top level
// only), its name not starting with a dot.
func inFolder(rel, dir string, fold bool) bool {
	d, base := path.Split(rel)
	return same(strings.TrimSuffix(d, "/"), dir, fold) && !strings.HasPrefix(base, ".") &&
		strings.HasSuffix(strings.ToLower(base), ".md")
}

// same compares two project-relative paths, letter case folded with fold.
func same(a, b string, fold bool) bool {
	if fold {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// machineMains are the main checkouts whose machine settings say a project's mode: the verified main checkout (or the
// project's own folder), and the main checkout its .git names when that differs (verified or not), so a worktree whose
// main cannot be verified still meets its project's mode.
func (j *judge) machineMains(p *linkedProject) []string {
	out := []string{j.local(p).Main}
	if !samePlace(p.layout.Main, out[0]) {
		out = append(out, p.layout.Main)
	}
	return out
}

// modeOf reads this machine's settings for a project, once a run: command when any of its machine mains reads command.
func (j *judge) modeOf(p *linkedProject) (workspace.MachineSettings, error) {
	out := workspace.MachineSettings{StatusWrites: workspace.StatusWritesAgents}
	if j.home == "" {
		return out, nil
	}
	for _, m := range j.machineMains(p) {
		got, ok := j.modes[m]
		if !ok {
			ms, err := workspace.LoadMachineSettings(j.home, m)
			if err != nil {
				return out, err
			}
			got = ms
			j.modes[m] = got
		}
		if got.StatusWrites == workspace.StatusWritesCommand {
			return got, nil
		}
	}
	return out, nil
}

// tripwire judges one judged form of a change by command mode's tripwires; false with the refusal when one trips.
func (j *judge) tripwire(c change, x judged) (Decision, bool) {
	doc := j.documentOf(x)
	if doc == nil {
		return Decision{}, true
	}
	ms, err := j.modeOf(x.p)
	if err != nil {
		next := "tell the person: a person fixes this machine's settings for the workspace"
		var we *workspace.Error
		if errors.As(err, &we) && we.Next != "" {
			next = "tell the person: " + we.Next
		}
		return j.trip(c, x, RuleBadSettings, "this machine's settings for the workspace cannot be read ("+oneLine(err.Error())+
			"), so the guard cannot tell whether command mode's tripwires apply to this document", oneLine(next)), false
	}
	if ms.StatusWrites != workspace.StatusWritesCommand {
		return Decision{}, true
	}
	next := "run this workspace's status command (" + quote(ms.StatusCommand, 80) + ") for status moves, lane changes, grants " +
		"and outside labels, and keep the edit to the file's other lines"
	if doc.kind == "task" && x.p.layout.Worktree() {
		return j.trip(c, x, RuleTaskOutsideMain, "a task file outside the main checkout: in command mode tasks change on "+
			"the main checkout only, through the status command", next), false
	}
	before, err := os.ReadFile(x.abs)
	exists := err == nil
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		if doc.status {
			return j.trip(c, x, RuleStatusLine, "the guard cannot read the document ("+oneLine(err.Error())+"), so it cannot "+
				"tell whether the change keeps its status line", next), false
		}
		return Decision{}, true
	}
	after, known := afterOf(c.in, before, exists)
	if !known {
		if exists && doc.status && readDoc(before).hasStatus {
			return j.trip(c, x, RuleStatusLine, "the guard cannot tell what the change leaves in the document's status line "+
				"(a delete or move, a NotebookEdit, or an old_string not in the file as written)", next), false
		}
		return Decision{}, true
	}
	if doc.pack == "" && doc.kind != "" {
		if err := bonsaiFormat(doc.kind, after); err != nil {
			return j.trip(c, x, RuleBadFormat, "the "+doc.kind+" would fail its format ("+oneLine(err.Error())+"), and in "+
				"command mode one broken file makes the active task none for everyone", "fix the text so it passes its format "+
				"(bonsai check --schema bonsai."+doc.kind+" lists every field), or "+next), false
		}
	}
	var bf docFields
	if exists {
		bf = readDoc(before)
	}
	af := readDoc(after)
	if doc.status {
		switch {
		case exists && bf.status != af.status:
			return j.trip(c, x, RuleStatusLine, fmt.Sprintf("it changes the document's status line (%s to %s)",
				quote(bf.status, 40), quote(af.status, 40)), next), false
		case !exists && af.hasStatus:
			first, err := j.firstStatus(x, doc)
			if err != nil {
				return j.trip(c, x, RuleBadLock, "the lock cannot be read to tell the first status of a new "+doc.kind+" ("+
					oneLine(err.Error())+")", "tell the person: a person fixes the lock (bonsai check names the problem)"), false
			}
			if af.status != first {
				return j.trip(c, x, RuleStatusLine, fmt.Sprintf("a new %s starts at its first status, %s, not %s", doc.kind,
					quote(first, 40), quote(af.status, 40)), next), false
			}
		}
	}
	if doc.kind == "task" && doc.pack == "" {
		if exists && bf.lane != af.lane {
			return j.trip(c, x, RuleLaneLine, fmt.Sprintf("it changes the task's lane line (%s to %s)", quote(bf.lane, 40),
				quote(af.lane, 40)), next), false
		}
		if g := j.personOnlyGrant(x, bf.allows, af.allows); g != "" {
			return j.trip(c, x, RuleGrantPersonOnly, "it puts the person-only path "+quote(g, 80)+" into the task's "+
				"bonsai.allows, which only a person's grant may do in command mode", next), false
		}
	}
	if changed := changedLabels(bf.labels, af.labels); len(changed) > 0 && !allBonsai(changed) {
		name, err := j.outsideLabel(x, changed)
		switch {
		case err != nil:
			return j.trip(c, x, RuleOutsideLabel, "it changes the labels "+strings.Join(quoteAll(changed, 60), ", ")+", and the "+
				"label definitions cannot be read to tell whether one is set from outside ("+oneLine(err.Error())+")", next), false
		case name != "":
			return j.trip(c, x, RuleOutsideLabel, "it changes the label "+quote(name, 60)+", whose definition reads set_by: "+
				"outside, which agents never write", next), false
		}
	}
	return Decision{}, true
}

// trip is the refusal of a tripwire.
func (j *judge) trip(c change, x judged, rule, why, next string) Decision {
	shown := quote(x.rel, 120)
	if len(x.chain) > 1 {
		shown += " (in the project " + quote(path.Base(strings.ReplaceAll(x.p.root, `\`, "/")), 60) + ")"
	}
	d := deny(rule, fmt.Sprintf("%s of %s refused: in command mode (this machine's status_writes), %s", c.tool, shown, why), next)
	d.Target = target(j.recordIn(x.chain), []string{x.abs})
	return d
}

// afterOf gives what a file holds after a file tool's call, its lines ending in LF; false when that cannot be told
// (no call, a NotebookEdit, an old_string not in the file).
func afterOf(in *hookin.Payload, before []byte, exists bool) ([]byte, bool) {
	if in == nil {
		return nil, false
	}
	switch in.Tool {
	case "Write":
		return lfBytes([]byte(in.Content)), true
	case "Edit", "MultiEdit":
	default:
		return nil, false
	}
	edits := in.Edits
	text := ""
	if exists {
		text = lf(string(before))
	} else {
		if len(edits) == 0 || edits[0].Old != "" {
			return nil, false
		}
		text, edits = lf(edits[0].New), edits[1:]
	}
	for _, e := range edits {
		old, repl := lf(e.Old), lf(e.New)
		if old == "" {
			return nil, false
		}
		if !strings.Contains(text, old) {
			text, old = straightQuotes(text), straightQuotes(old)
			if !strings.Contains(text, old) {
				return nil, false
			}
		}
		if e.ReplaceAll {
			text = strings.ReplaceAll(text, old, repl)
		} else {
			text = strings.Replace(text, old, repl, 1)
		}
	}
	return []byte(text), true
}

func lf(s string) string      { return strings.ReplaceAll(s, "\r\n", "\n") }
func lfBytes(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }
func straightQuotes(s string) string {
	return strings.NewReplacer("‘", "'", "’", "'", "“", `"`, "”", `"`).Replace(s)
}

// bonsaiFormat holds a document of one of Bonsai's kinds to its format, under format 0 or 1.
func bonsaiFormat(kind string, raw []byte) error {
	var err error
	switch kind {
	case "task":
		_, err = format.ReadTask(raw)
	case "run":
		_, err = format.ReadRun(raw)
	case "state":
		_, err = format.ReadState(raw)
	}
	return err
}

// docFields are the lines of a document's frontmatter the tripwires compare.
type docFields struct {
	status    string
	hasStatus bool
	lane      string
	labels    schema.Object // label name to value, in the file's order; nil when there are none or none could be read
	allows    []string      // the task's bonsai.allows (format 0: allows_assets), as text
}

// readDoc reads a document's frontmatter under format 1, or format 0 when it has no format: line; when it does not
// read, its status: and lane: lines are read as text and its labels count as none.
func readDoc(raw []byte) docFields {
	var m *reader.Map
	r := reader.ReadMarkdown(raw)
	switch r.Outcome {
	case reader.Accepted:
		m = r.Value
	case reader.Format0:
		if m0, ok := reader.ReadFormat0Markdown(raw).Map(); ok {
			m = m0
		}
	}
	if m == nil {
		return scanDoc(raw)
	}
	var f docFields
	if v, ok := m.Get("status"); ok {
		f.status, f.hasStatus = scalarText(v), true
	}
	if v, ok := m.Get("lane"); ok {
		f.lane = scalarText(v)
	}
	if v, ok := m.Get("labels"); ok {
		if lm, ok := v.(*reader.Map); ok {
			f.labels, _ = reader.JSON(lm).(schema.Object)
		}
	}
	allows, has := f.labels.Get("bonsai.allows")
	if !has {
		allows, has = m.Get("allows_assets")
		allows = reader.JSON(allows)
	}
	if has {
		if list, ok := allows.([]any); ok {
			for _, e := range list {
				f.allows = append(f.allows, scalarText(e))
			}
		} else if allows != nil {
			f.allows = []string{scalarText(allows)}
		}
	}
	return f
}

// scanDoc reads the status: and lane: lines of a frontmatter that does not read, as text.
func scanDoc(raw []byte) docFields {
	var f docFields
	lines := strings.Split(lf(string(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], " ") != "---" {
		return f
	}
	for _, l := range lines[1:] {
		if strings.TrimRight(l, " ") == "---" {
			break
		}
		key, val, ok := strings.Cut(l, ":")
		if !ok || strings.TrimSpace(key) != key {
			continue
		}
		if i := strings.Index(val, " #"); i >= 0 {
			val = val[:i]
		}
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		switch strings.Trim(key, `"'`) {
		case "status":
			f.status, f.hasStatus = val, true
		case "lane":
			f.lane = val
		}
	}
	return f
}

// scalarText gives a frontmatter value as text: a string as itself, null as "", any other value as JSON.
func scalarText(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	}
	return schema.Show(reader.JSON(v))
}

// changedLabels lists the labels set, changed or taken away between two frontmatters, in the after file's order then
// the before file's.
func changedLabels(before, after schema.Object) []string {
	var out []string
	for _, m := range after {
		if b, ok := before.Get(m.Key); !ok || schema.Show(b) != schema.Show(m.Value) {
			out = append(out, m.Key)
		}
	}
	for _, m := range before {
		if _, ok := after.Get(m.Key); !ok {
			out = append(out, m.Key)
		}
	}
	return out
}

// allBonsai reports labels that are all Bonsai's own (bonsai.*), which the label definitions need not be read for.
func allBonsai(names []string) bool {
	for _, n := range names {
		if !strings.HasPrefix(n, "bonsai.") {
			return false
		}
	}
	return true
}

// personOnlyGrant gives the first glob an edit adds to a task's bonsai.allows that is a person-only path (""
// for none): one the floor counts as person-only (floorOf, in the floor's order, so .bonsai/STATE.md is not), one
// inside a glob of bonsai.yaml's person_only list or covering one, or one covering a floor person-only glob (**,
// *.yaml).
func (j *judge) personOnlyGrant(x judged, before, after []string) string {
	listed := j.listsOf(x.p).cfg.PersonOnly
	floor := floorPersonOnly()
	for _, g := range after {
		if listHas(before, g) {
			continue
		}
		if f := floorOf(g, x.fold, false); f != nil && f.Class == RulePersonOnly {
			return g
		}
		for _, p := range listed {
			if matchGlob(p, g, x.fold) || matchGlob(g, p, x.fold) {
				return g
			}
		}
		for _, p := range floor {
			if matchGlob(g, p, x.fold) {
				return g
			}
		}
	}
	return ""
}

func listHas(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// outsideLabel gives the first of the changed labels whose definition in force reads set_by: outside, "" for none;
// Bonsai's own (bonsai.*) are never set from outside.
// The definitions are the locked packs' (the lock read in full) and those attached on this machine; a changed label
// with no definition, while a definition file could not be read, is refused (its definition may be the one missing).
func (j *judge) outsideLabel(x judged, changed []string) (string, error) {
	l := j.listsOf(x.p)
	lock, err := workspace.LoadLock(l.from)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) || len(l.cfg.Packs) > 0 {
			return "", err
		}
		lock = nil
	}
	defs := map[string]string{}
	var problems []error
	for _, m := range j.machineMains(x.p) {
		sets, probs := workspace.LabelsInForce(lock, j.home, m)
		problems = append(problems, probs...)
		for _, s := range sets {
			for _, d := range s.Labels {
				if prev, ok := defs[d.Name]; !ok || prev != "outside" {
					defs[d.Name] = d.SetBy
				}
			}
		}
	}
	for _, name := range changed {
		if strings.HasPrefix(name, "bonsai.") {
			continue // Bonsai's own labels (contract §5.6) are set by agents
		}
		set, ok := defs[name]
		switch {
		case ok && set == "outside":
			return name, nil
		case !ok && len(problems) > 0:
			return "", problems[0]
		}
	}
	return "", nil
}

// firstStatus is the first status of a document's kind: the task's from Bonsai's kinds, a pack's from its declares in
// the lock (read in full).
func (j *judge) firstStatus(x judged, doc *docKind) (string, error) {
	if doc.pack == "" {
		if st := bonsaiDocKinds()[doc.kind].Statuses; len(st) > 0 {
			return st[0], nil
		}
		return "", nil
	}
	lock, err := workspace.LoadLock(j.listsOf(x.p).from)
	if err != nil {
		return "", err
	}
	for _, lp := range lock.Packs {
		if lp.ID != doc.pack {
			continue
		}
		d, err := lp.Declared()
		if err != nil {
			return "", err
		}
		for _, k := range d.Documents {
			if k.Kind == doc.kind && len(k.Statuses) > 0 {
				return k.Statuses[0], nil
			}
		}
	}
	return "", nil
}
