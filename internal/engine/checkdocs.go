package engine

// check's findings on the documents of Bonsai's own kinds (contract §2.3, §5, §7; spec §6, §10), read from this
// checkout at the places bonsai.yaml's documents names (workspace.DocKinds, their one home):
//
//   - format0-new: a task, run report or STATE file with no format: line that the lock's format0 list does not hold
//     (contract §2.3: "a new or changed one is a finding (give it format: bonsai.task/1 first)"); with no lock, there
//     is no list to hold it to, and the lock's own finding says so;
//   - document: a file of one of Bonsai's kinds that does not read under its format, format 0 or 1 (task, run, state,
//     memory and its index, the two tables);
//   - label: a format-1 document's label whose value is not the kind its definition says, or that is on a kind of
//     document its definition does not name (contract §5.2). A label no definition in force names is left alone:
//     Bonsai's own bonsai.* definitions arrive with the base pack (step 5.5);
//   - label-twice: two sources of the labels in force define one name (contract §5.1);
//   - absolute-path: an absolute path in a field of bonsai.yaml or of a format-1 document (contract §2.6: "no
//     absolute paths in committed formats"): a value that is one (/x, C:\x, \\host, a file: URL), or a word of a value
//     under a well-known root (/home/, /Users/, /tmp/, C:\ ...). A pack's source in bonsai.yaml is held to it too
//     (spec §14 check 2: "the lock's source is the pack's remote URL"): a remote URL (https://, http://, ssh://,
//     git://, or git's user@host:path) is left alone; a local absolute path or a file: URL holding one is a finding,
//     whose step is a person's (only a person knows the pack's remote URL). Left out: the lock, Bonsai's copy of
//     bonsai.yaml and the packs, so a path there is found at its source; a format-0 file, frozen as it is (contract
//     §2.3); and markdown bodies, which are prose, not fields;
//   - block-size, memory-index-size, memory-note-size: the fixed budgets (spec §6, §10; the memory caps' one home is
//     MemoryCaps);
//   - missing-path: a project path that CLAUDE.md, STATE or a memory note names and that does not exist (doc
//     freshness), or that exists here only because git ignores it, so a clone does not have it: an @import, a
//     markdown link, or a path in backticks whose first part is in the project. Bonsai's own
//     block in CLAUDE.md is left out: the lock checks it, its protocol imports are pack files the lock lists, and its
//     memory-index import is written before the index exists, by design (spec §6, §10: the memory skill writes the
//     index), so a project linked a moment ago has no finding for it;
//   - run-reports (a warning): run reports past bonsai.yaml's generated.run rule, by the date in their id, listed for a
//     person to delete (spec §6: Bonsai cleans no run report).

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// MemoryCap is one of the memory caps: the most a memory file of one place may hold. Lines 0 means no line cap.
type MemoryCap struct {
	Name  string // the cap's name, as the reference page lists it
	What  string // the file it holds, in words
	Lines int    // the most lines, 0 for none
	Bytes int    // the most bytes
}

// MemoryCaps are the memory caps (spec section 10; plan-5 5.6.3, 3): their one home. check reads them (the project's
// memory-index-size and memory-note-size findings, the personal layer's personal-budget warning), the reference page
// of lists prints them, and a test holds base's memory skill and the check words' meanings to them. Fixed, not
// settings. The personal index is 40 lines at the project index's own rate (120 lines in 12 KB, about 100 bytes a
// line), so a few overlong lines cannot hide under its line count.
var MemoryCaps = []MemoryCap{
	{Name: "project-index", What: "the project's memory index (INDEX.md in the folder bonsai.yaml names under documents.memory)", Lines: 120, Bytes: 12 * 1024},
	{Name: "personal-index", What: "the personal layer's index, INDEX.md in the Bonsai home's personal/", Lines: 40, Bytes: 4 * 1024},
	{Name: "note", What: "a memory note, in either layer", Bytes: 4 * 1024},
}

// memoryCap finds a cap by its name; a name not in the table is a bug.
func memoryCap(name string) MemoryCap {
	for _, c := range MemoryCaps {
		if c.Name == name {
			return c
		}
	}
	panic("engine: no memory cap " + name)
}

// KB writes a cap's bytes in KB, as people read them: "12 KB".
func KB(n int) string { return strconv.Itoa(n/1024) + " KB" }

// capLines counts a file's lines as the caps count them: line ends, and a last line with none.
func capLines(raw []byte) int {
	n := bytes.Count(raw, []byte("\n"))
	if len(raw) > 0 && !bytes.HasSuffix(raw, []byte("\n")) {
		n++
	}
	return n
}

// now is check's clock (a variable, so a test can set the day).
var now = time.Now

// isFormat0 reports whether a markdown file has no format: line (contract §2.3).
func isFormat0(raw []byte) bool { return reader.ReadMarkdown(raw).Outcome == reader.Format0 }

// format0Kind names the format-0 kind of a file the lock's format0 list holds: state for STATE, run in the run
// reports' folder, else task.
func format0Kind(cfg *workspace.Config, p string) string {
	switch {
	case p == workspace.StateFile:
		return "state"
	case cfg != nil && cfg.Full != nil && cfg.Full.Documents.Run != "" && strings.HasPrefix(p, strings.TrimSuffix(cfg.Full.Documents.Run, "/")+"/"):
		return "run"
	}
	return "task"
}

// docRead reads one document of a Bonsai kind under its format: its labels and its fields as read (format 1 only; nil
// for a format-0 file), or the refusal.
func docRead(kind string, raw []byte) (labels schema.Object, fields schema.Object, err error) {
	f := format.MustLookup(kind)
	switch kind {
	case "task":
		var t *format.Task
		if t, err = format.ReadTask(raw); err == nil && t.Format0 == nil {
			labels = t.Labels
		}
	case "run":
		var x *format.Run
		if x, err = format.ReadRun(raw); err == nil && x.Format0 == nil {
			labels = x.Labels
		}
	case "state":
		var x *format.State
		if x, err = format.ReadState(raw); err == nil && x.Format0 == nil {
			labels = x.Labels
		}
	case "memory":
		var x *format.Memory
		if x, err = format.ReadMemory(raw); err == nil {
			labels = x.Labels
		}
	case "tasks":
		_, err = format.ReadTasks(raw)
	case "sessions":
		_, err = format.ReadSessions(raw)
	default:
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if !isFormat0(raw) {
		fields, _, _ = f.ReadYAML(raw)
	}
	return labels, fields, nil
}

// checkDocuments checks the documents of Bonsai's kinds, the labels in force, the budgets, the paths the
// instruction file, STATE and the memory notes name, bonsai.yaml's fields and the run reports.
func (r *CheckResult) checkDocuments() {
	cfg := r.Config
	kinds, err := workspace.DocKinds(cfg.Full, r.Lock)
	if err != nil {
		kinds, _ = workspace.DocKinds(cfg.Full, nil)
	}
	sets, _ := workspace.LabelsInForce(r.Lock, r.Home, r.Main)
	defs := r.checkLabelsTwice(sets)
	r.checkAbsolute(workspace.ConfigFile, configFields(cfg))
	var named []string // the files whose paths are checked: STATE and the memory notes
	for _, k := range kinds {
		if k.From != "bonsai" || k.Format == "" {
			continue
		}
		short := strings.TrimPrefix(k.Format, "bonsai.")
		files, err := workspace.DocFiles(r.Root, k)
		if err != nil {
			r.add("document", strings.TrimSuffix(k.Path+k.File, "/"), "", "the "+k.Kind+" documents ("+k.Path+k.File+") cannot be read: "+oneLine(err.Error()),
				"check the folder's permissions, then run: bonsai check")
			continue
		}
		if short == "memory" && k.Path != "" {
			index := strings.TrimSuffix(k.Path, "/") + "/" + MemoryIndex
			if _, exists, _ := readFile(r.Root, index); exists {
				files = append(files, index)
			}
		}
		for _, p := range files {
			raw, exists, err := readFile(r.Root, p)
			if err != nil || !exists {
				continue
			}
			r.checkDocument(short, k.Kind, p, raw, defs)
			if short == "memory" || short == "state" {
				named = append(named, p)
			}
		}
	}
	r.checkBudgets()
	r.checkNamedPaths(named)
	r.checkRunReports(kinds)
}

// checkDocument checks one document of a Bonsai kind: format 0 new, its format, its labels and its fields' paths.
func (r *CheckResult) checkDocument(short, kind, p string, raw []byte, defs map[string]labelDef) {
	f0 := isFormat0(raw)
	if f0 && contains(workspace.Format0Kinds, short) && r.Lock != nil {
		if _, listed := r.Lock.Format0[p]; !listed {
			r.add("format0-new", p, "", p+" has no format: line and is not on the lock's format0 list: a "+kind+" file new since the link "+
				"is written in format 1 (contract section 2.3)", "give it a format: line first, format: bonsai."+short+"/1, with the fields "+
				"this lists, run: bonsai check --schema bonsai."+short)
		}
	}
	labels, fields, err := docRead(short, raw)
	if err != nil {
		next := "fix it (the fields and their allowed values: bonsai check --schema bonsai." + short + "), then run: bonsai check"
		if short == "tasks" || short == "sessions" {
			next = "the table is generated, never edited by hand: restore it, run: git checkout -- " + ShellArg(p)
		}
		if re, ok := err.(*format.ReadError); ok && re.TooNew {
			next = "read it with a newer Bonsai, or run: git checkout -- " + ShellArg(p)
		}
		r.add("document", p, "", p+" does not read as "+formatOf(short, f0)+": "+docError(err), next)
		return
	}
	r.checkLabels(kind, p, labels, defs)
	if fields != nil {
		r.checkAbsolute(p, fields)
	}
	if note := memoryCap("note"); short == "memory" && len(raw) > note.Bytes && path.Base(p) != MemoryIndex {
		r.add("memory-note-size", p, "", fmt.Sprintf("the memory note %s is %d bytes, over its %d (spec section 10)", p, len(raw), note.Bytes),
			"shorten the note to at most "+KB(note.Bytes)+", or split it into two notes (and their two lines in "+MemoryIndex+")")
	}
}

// formatOf names the format a document was read under, for a person.
func formatOf(short string, f0 bool) string {
	if f0 {
		return "bonsai." + short + " format 0"
	}
	return "bonsai." + short + "/1"
}

// docError is a reader's refusal without its own next step (the finding names one).
func docError(err error) string {
	if re, ok := err.(*format.ReadError); ok {
		where := ""
		if re.Line > 0 {
			where = "line " + strconv.Itoa(re.Line) + ": "
		}
		code := ""
		if re.Code != "" {
			code = " (" + re.Code + ")"
		}
		return ascii(where + re.Msg + code)
	}
	return ascii(oneLine(err.Error()))
}

func oneLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		return s[:i]
	}
	return s
}

// labelDef is a label's definition in force and where it came from.
type labelDef struct {
	format.LabelDef
	from string // a pack's id, or machine
}

// checkLabelsTwice finds a label name two sources define (contract §5.1: a definition may be added, never redefined)
// and gives the definitions in force by name, the first source's for a name defined twice.
func (r *CheckResult) checkLabelsTwice(sets []workspace.LabelSet) map[string]labelDef {
	defs := map[string]labelDef{}
	for _, s := range sets {
		for _, l := range s.Labels {
			first, twice := defs[l.Name]
			if !twice {
				defs[l.Name] = labelDef{LabelDef: l, from: s.From}
				continue
			}
			where := func(from string) string {
				if from == workspace.FromMachine {
					return "a file attached on this machine"
				}
				return "the pack " + from
			}
			next := "a person takes one of the two definitions out: the pack's maker drops it and releases the pack again, " +
				"or the pack leaves bonsai.yaml"
			if s.From == workspace.FromMachine {
				next = "a person takes the attached definitions off this machine: delete " + ascii(filepath.ToSlash(s.File)) +
					" (its namespace may not be a pack's, contract section 5.1)"
			}
			r.add("label-twice", "", "", "the label "+l.Name+" is defined twice, by "+where(first.from)+" and by "+where(s.From)+
				" (contract section 5.1: a definition is added, never redefined)", next)
		}
	}
	return defs
}

// checkLabels holds a document's labels to their definitions (contract §5.2).
func (r *CheckResult) checkLabels(kind, p string, labels schema.Object, defs map[string]labelDef) {
	for _, m := range labels {
		d, ok := defs[m.Key]
		if !ok {
			continue
		}
		who := "agent"
		if d.SetBy == "outside" {
			who = "person"
		}
		if !contains(d.Kinds, kind) {
			r.add("label", p, who, fmt.Sprintf("%s carries the label %s, which goes only on %s (its definition, from %s)", p, m.Key,
				strings.Join(d.Kinds, ", "), d.from), "take "+m.Key+" out of "+p+"'s labels")
			continue
		}
		if why := labelValueProblem(d.LabelDef, m.Value); why != "" {
			r.add("label", p, who, fmt.Sprintf("%s's label %s is %s: %s (its definition, from %s)", p, m.Key, schema.Show(m.Value), why, d.from),
				"set "+m.Key+" in "+p+" to "+labelKindText(d.LabelDef)+", or take it out")
		}
	}
}

// labelValueProblem says why a value does not fit its definition's kind (contract §5.2), "" when it fits.
func labelValueProblem(d format.LabelDef, v any) string {
	switch d.Kind {
	case "choice":
		s, ok := v.(string)
		if !ok || !contains(d.Values, s) {
			return "not one of " + strings.Join(d.Values, ", ")
		}
	case "text":
		s, ok := v.(string)
		switch {
		case !ok:
			return "not text"
		case strings.ContainsAny(s, "\r\n"):
			return "not one line"
		case d.Max != nil && int64(len([]rune(s))) > *d.Max:
			return fmt.Sprintf("longer than %d characters", *d.Max)
		case d.Pattern != nil:
			re, err := regexp.Compile(*d.Pattern)
			if err == nil && !fullMatch(re, s) {
				return "not matching " + *d.Pattern
			}
		}
	case "number":
		if _, ok := v.(json.Number); !ok {
			return "not a number"
		}
	case "list":
		l, ok := v.([]any)
		if !ok {
			return "not a list"
		}
		if d.Max != nil && int64(len(l)) > *d.Max {
			return fmt.Sprintf("more than %d items", *d.Max)
		}
		items := "text"
		if d.Items != nil {
			items = *d.Items
		}
		for _, it := range l {
			if _, isNum := it.(json.Number); items == "number" && !isNum {
				return "an item is not a number"
			}
			if s, isText := it.(string); items == "text" && (!isText || strings.ContainsAny(s, "\r\n")) {
				return "an item is not one line of text"
			}
		}
	}
	return ""
}

func fullMatch(re *regexp.Regexp, s string) bool {
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}

// labelKindText says what a definition takes, as the block's label lines do.
func labelKindText(d format.LabelDef) string {
	line := labelLine(d)
	line = strings.TrimPrefix(line, "- "+d.Name+": ")
	if i := strings.Index(line, "; on "); i >= 0 {
		line = line[:i]
	}
	return line
}

// configFields is bonsai.yaml as a document, for the absolute-path finding.
func configFields(cfg *workspace.Config) schema.Object {
	if cfg == nil || cfg.Full == nil {
		return nil
	}
	doc, err := format.MustLookup("workspace").Document(cfg.Full)
	if err != nil {
		return nil
	}
	return doc
}

// checkAbsolute finds an absolute path in a document's fields (contract §2.6), each field once.
func (r *CheckResult) checkAbsolute(file string, doc schema.Object) {
	var walk func(field string, v any)
	walk = func(field string, v any) {
		switch x := v.(type) {
		case schema.Object:
			for _, m := range x {
				name := m.Key
				if field != "" {
					name = field + "." + m.Key
				}
				walk(name, m.Value)
			}
		case []any:
			for i, e := range x {
				walk(field+"["+strconv.Itoa(i)+"]", e)
			}
		case string:
			source := file == workspace.ConfigFile && packSource().MatchString(field)
			if source && remoteURL().MatchString(x) {
				return // a remote URL, which every clone fetches
			}
			if abs, ok := absoluteIn(x); ok {
				next := "write it project-relative, or as ~/... for a place under a home folder"
				switch {
				case source:
					next = "a person names the pack's remote URL (https://, ssh:// or git@host:...) as its source in bonsai.yaml, " +
						"then run: bonsai update --yes"
				case file == workspace.ConfigFile:
					next = "a person edits bonsai.yaml: " + next + "; then run: bonsai update --yes"
				}
				r.add("absolute-path", file, map[bool]string{true: "person", false: "agent"}[file == workspace.ConfigFile],
					file+"'s field "+field+" holds the absolute path "+ascii(abs)+" (contract section 2.6: no absolute path in a committed file; "+
						"it names one machine's folders)", next)
			}
		}
	}
	walk("", doc)
}

var (
	packSource = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^packs\[[0-9]+\]\.source$`) })
	remoteURL  = sync.OnceValue(func() *regexp.Regexp {
		return regexp.MustCompile(`^((https?|ssh|git|git\+ssh)://[^/]|[A-Za-z0-9._-]+@[A-Za-z0-9.-]+:)`)
	})
	absRoots = sync.OnceValue(func() *regexp.Regexp {
		return regexp.MustCompile(`^(/(home|Users|root|mnt|tmp|var|srv|opt|usr|etc|private|Volumes)/|[A-Za-z]:[\\/]|\\\\[^\\]|file:/)`)
	})
	absWholeVal = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^/[A-Za-z0-9._-]`) })
)

// absoluteIn finds an absolute path in a value: the whole value one (/x, C:\x, \\host, a file: URL), or a word of
// it under a well-known root (/home/, /Users/, /tmp/, C:\ ...), so prose such as "the /api/users endpoint" is not
// taken for one. ~/ is a home's place, never absolute here.
func absoluteIn(s string) (string, bool) {
	v := strings.TrimSpace(s)
	if absWholeVal().MatchString(v) || absRoots().MatchString(v) {
		return firstWord(v), true
	}
	for _, w := range strings.FieldsFunc(s, func(r rune) bool { return strings.ContainsRune(" \t\"'`()[]{}<>,;=", r) }) {
		if absRoots().MatchString(w) {
			return w, true
		}
	}
	return "", false
}

func firstWord(s string) string {
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i]
	}
	return s
}

// checkBudgets finds the block over its 40 lines and the memory index over its cap (MemoryCaps; spec §6, §10).
func (r *CheckResult) checkBudgets() {
	if bd, err := readBlock(r.Root); err == nil && bd.found {
		if n := strings.Count(bd.region(), "\n"); n > MaxBlockLines {
			r.add("block-size", BlockFile, "", fmt.Sprintf("Bonsai's block in %s is %d lines, over its fixed %d (spec section 6)", BlockFile, n, MaxBlockLines),
				"take Bonsai's block back (your copy is saved in the Bonsai home), run: bonsai update --yes --adopt "+BlockFile)
		}
	}
	mem := ""
	if r.Config.Full != nil {
		mem = strings.TrimSuffix(r.Config.Full.Documents.Memory, "/")
	}
	if mem == "" {
		return
	}
	index := mem + "/" + MemoryIndex
	raw, exists, _ := readFile(r.Root, index)
	if !exists {
		return
	}
	lines, c := capLines(raw), memoryCap("project-index")
	if lines > c.Lines || len(raw) > c.Bytes {
		r.add("memory-index-size", index, "", fmt.Sprintf("the memory index %s is %d lines and %d bytes, over its fixed %d lines or %s (spec section 10)",
			index, lines, len(raw), c.Lines, KB(c.Bytes)), fmt.Sprintf("shorten it to one short line per note, at most %d lines and %s: merge notes, or move a note's detail into the note", c.Lines, KB(c.Bytes)))
	}
}

var (
	importRef   = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`(?:^|\s)@([^\s]+)`) })
	linkRef     = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`\]\(([^)\s]+)\)`) })
	backtickRef = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile("`([^`\\s]+)`") })
	lineSuffix  = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`:[0-9]+(-[0-9]+)?$`) })
)

// checkNamedPaths finds project paths that CLAUDE.md (outside Bonsai's block), STATE and the memory notes name and
// that do not exist, or exist only here because git ignores them (gitIgnored): a clone, a fresh worktree or CI does
// not have such a path, so it is missing there.
func (r *CheckResult) checkNamedPaths(files []string) {
	type named struct {
		file, ref, target string
		line              int
		there             bool
	}
	var refsFound []named
	check := func(file string, text string, firstLine int) {
		dir := path.Dir(file)
		inFence := false
		sc := bufio.NewScanner(strings.NewReader(text))
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for n := firstLine; sc.Scan(); n++ {
			line := sc.Text()
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inFence = !inFence
				continue
			}
			if inFence {
				continue
			}
			var refs []struct {
				p        string
				fromFile bool
			}
			for _, m := range importRef().FindAllStringSubmatch(line, -1) {
				refs = append(refs, struct {
					p        string
					fromFile bool
				}{strings.TrimRight(m[1], ".,;:)"), true})
			}
			for _, m := range linkRef().FindAllStringSubmatch(line, -1) {
				refs = append(refs, struct {
					p        string
					fromFile bool
				}{m[1], true})
			}
			for _, m := range backtickRef().FindAllStringSubmatch(line, -1) {
				if strings.Contains(m[1], "/") {
					refs = append(refs, struct {
						p        string
						fromFile bool
					}{m[1], false})
				}
			}
			for _, ref := range refs {
				target, ok := projectRef(r.Root, dir, ref.p, ref.fromFile)
				if !ok {
					continue
				}
				_, err := os.Stat(filepath.Join(r.Root, filepath.FromSlash(target)))
				refsFound = append(refsFound, named{file: file, ref: ref.p, target: target, line: n, there: err == nil})
			}
		}
	}
	if raw, exists, _ := readFile(r.Root, BlockFile); exists {
		text := strings.ReplaceAll(string(raw), "\r\n", "\n")
		bd, err := readBlock(r.Root)
		if err == nil && bd.found {
			before := strings.ReplaceAll(string(bd.raw[:bd.start]), "\r\n", "\n")
			after := strings.ReplaceAll(string(bd.raw[bd.end:]), "\r\n", "\n")
			check(BlockFile, before, 1)
			check(BlockFile, after, 1+strings.Count(before, "\n")+strings.Count(bd.region(), "\n"))
		} else if err == nil {
			check(BlockFile, text, 1)
		}
	}
	sort.Strings(files)
	for _, f := range files {
		if raw, exists, _ := readFile(r.Root, f); exists {
			check(f, strings.ReplaceAll(string(raw), "\r\n", "\n"), 1)
		}
	}
	var there []string
	for _, f := range refsFound {
		if f.there {
			there = append(there, f.target)
		}
	}
	ignored := gitIgnored(r.Root, there)
	for _, f := range refsFound {
		switch {
		case !f.there:
			r.add("missing-path", f.file, "", fmt.Sprintf("%s line %d names %s, which does not exist in the project", f.file, f.line, ascii(f.ref)),
				fmt.Sprintf("fix line %d of %s: name the path that exists now, or take the reference out", f.line, f.file))
		case ignored[f.target]:
			r.add("missing-path", f.file, "", fmt.Sprintf("%s line %d names %s, which git ignores, so a clone of the project does not have it",
				f.file, f.line, ascii(f.ref)),
				fmt.Sprintf("fix line %d of %s: name a path the project commits, or say the file in words", f.line, f.file))
		}
	}
}

// gitIgnored gives the paths among paths (project-relative, forward slashes) that git ignores in the checkout at root,
// by one git check-ignore: a path git tracks is never among them. When git cannot answer (no git, no checkout), none:
// the paths are judged by the files alone, as before.
func gitIgnored(root string, paths []string) map[string]bool {
	out := map[string]bool{}
	if len(paths) == 0 {
		return out
	}
	cmd := exec.Command("git", "-C", root, "check-ignore", "-z", "--stdin")
	cmd.Env = gitEnv()
	cmd.Stdin = strings.NewReader(strings.Join(paths, "\x00") + "\x00")
	raw, err := cmd.Output()
	var ee *exec.ExitError
	if err != nil && !(errors.As(err, &ee) && ee.ExitCode() == 1) { // 1: none of them is ignored
		return out
	}
	for _, p := range bytes.Split(raw, []byte{0}) {
		if len(p) > 0 {
			out[filepath.ToSlash(string(p))] = true
		}
	}
	return out
}

// projectRef gives a reference's project-relative path when it names a project path Bonsai can judge: no URL, no
// anchor alone, no home (~) or absolute path, no placeholder or glob. An @import or a link is read from the file's own
// folder (as Claude Code reads an import); a path in backticks from the project's top, and only when its first part
// is there, so a path of another repository's is not taken for this one's.
func projectRef(root, dir, ref string, fromFile bool) (string, bool) {
	if i := strings.IndexByte(ref, '#'); i >= 0 {
		ref = ref[:i]
	}
	ref = lineSuffix().ReplaceAllString(ref, "")
	switch {
	case ref == "" || strings.Contains(ref, "://") || strings.HasPrefix(ref, "mailto:"),
		strings.HasPrefix(ref, "~") || strings.HasPrefix(ref, "/") || strings.HasPrefix(ref, `\`),
		len(ref) > 1 && ref[1] == ':',
		strings.ContainsAny(ref, "<>*?{}$[]|\"'%"):
		return "", false
	}
	var p string
	if fromFile {
		p = path.Clean(path.Join(dir, ref))
	} else {
		p = path.Clean(strings.TrimPrefix(ref, "./"))
		first := strings.SplitN(p, "/", 2)[0]
		if _, err := os.Stat(filepath.Join(root, first)); err != nil {
			return "", false
		}
	}
	if p == "." || p == ".." || strings.HasPrefix(p, "../") {
		return "", false
	}
	return p, true
}

var runDate = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^R-([0-9]{4}-[0-9]{2}-[0-9]{2})-`) })

// checkRunReports lists the run reports past bonsai.yaml's generated.run rule (spec §6): keep_days by the date in a
// report's id, keep_newest by that date, newest first. One warning for them all, for a person: Bonsai cleans no run
// report.
func (r *CheckResult) checkRunReports(kinds []workspace.DocKind) {
	if r.Config.Full == nil {
		return
	}
	rule := r.Config.Full.Generated.Run
	if rule.KeepDays == nil && rule.KeepNewest == nil {
		return
	}
	type report struct{ path, day string }
	var reports []report
	for _, k := range kinds {
		if k.From != "bonsai" || k.Kind != "run" {
			continue
		}
		files, _ := workspace.DocFiles(r.Root, k)
		for _, f := range files {
			if m := runDate().FindStringSubmatch(path.Base(f)); m != nil {
				reports = append(reports, report{f, m[1]})
			}
		}
	}
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].day != reports[j].day {
			return reports[i].day > reports[j].day
		}
		return reports[i].path > reports[j].path
	})
	cutoff := ""
	if rule.KeepDays != nil {
		cutoff = now().AddDate(0, 0, -int(*rule.KeepDays)).Format("2006-01-02")
	}
	var past []string
	for i, rep := range reports {
		if (cutoff != "" && rep.day < cutoff) || (rule.KeepNewest != nil && int64(i) >= *rule.KeepNewest) {
			past = append(past, rep.path)
		}
	}
	if len(past) == 0 {
		return
	}
	sort.Strings(past)
	args := make([]string, len(past))
	for i, p := range past {
		args[i] = ShellArg(p)
	}
	r.add("run-reports", "", "", fmt.Sprintf("%d run %s past bonsai.yaml's generated.run rule (%s): %s", len(past),
		plural(len(past), "report is", "reports are"), keepText(rule), strings.Join(past, ", ")),
		"a person deletes them if they agree (Bonsai cleans no run report), run: git rm -q -- "+strings.Join(args, " ")+"; then commit")
}

func keepText(k format.Keep) string {
	var parts []string
	if k.KeepDays != nil {
		parts = append(parts, fmt.Sprintf("keep_days %d", *k.KeepDays))
	}
	if k.KeepNewest != nil {
		parts = append(parts, fmt.Sprintf("keep_newest %d", *k.KeepNewest))
	}
	return strings.Join(parts, ", ")
}
