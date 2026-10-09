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
//     under a well-known root (/home/, /Users/, /tmp/, C:\ ...). Left out: a pack's source in bonsai.yaml, which is
//     where git fetches the pack (a URL as git reads it, or a folder git reads for a pack in the making), not a place
//     in the project, and which the bridge never forwards; the lock, Bonsai's copy of bonsai.yaml and the packs, so a
//     path there is found at its source; a format-0 file, frozen as it is (contract §2.3); and markdown bodies, which
//     are prose, not fields;
//   - block-size, memory-index-size, memory-note-size: the fixed budgets (spec §6, §10);
//   - missing-path: a project path that CLAUDE.md, STATE or a memory note names and that does not exist (doc
//     freshness): an @import, a markdown link, or a path in backticks whose first part is in the project. Bonsai's own
//     block in CLAUDE.md is left out: the lock checks it, its protocol imports are pack files the lock lists, and its
//     memory-index import is written before the index exists, by design (spec §6, §10: the memory skill writes the
//     index), so a project linked a moment ago has no finding for it;
//   - run-reports (a warning): run reports past bonsai.yaml's generated.run rule, by the date in their id, listed for a
//     person to delete (spec §6: Bonsai cleans no run report).

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// The budgets (spec §6, §10): fixed, not settings.
const (
	MaxIndexLines = 120
	MaxIndexBytes = 12 * 1024
	MaxNoteBytes  = 4 * 1024
)

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
				"bonsai check --schema bonsai."+short+" lists (run: bonsai check --schema bonsai."+short+")")
		}
	}
	labels, fields, err := docRead(short, raw)
	if err != nil {
		next := "fix it, then run: bonsai check (the fields and their allowed values: bonsai check --schema bonsai." + short + ")"
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
	if short == "memory" && len(raw) > MaxNoteBytes && path.Base(p) != MemoryIndex {
		r.add("memory-note-size", p, "", fmt.Sprintf("the memory note %s is %d bytes, over its %d (spec section 10)", p, len(raw), MaxNoteBytes),
			"shorten the note to at most 4 KB, or split it into two notes (and their two lines in "+MemoryIndex+")")
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
			if file == workspace.ConfigFile && packSource.MatchString(field) {
				return // where git fetches the pack, not a place in the project
			}
			if abs, ok := absoluteIn(x); ok {
				next := "write it project-relative, or as ~/... for a place under a home folder"
				if file == workspace.ConfigFile {
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
	packSource  = regexp.MustCompile(`^packs\[[0-9]+\]\.source$`)
	absRoots    = regexp.MustCompile(`^(/(home|Users|root|mnt|tmp|var|srv|opt|usr|etc|private|Volumes)/|[A-Za-z]:[\\/]|\\\\[^\\]|file:/)`)
	absWholeVal = regexp.MustCompile(`^/[A-Za-z0-9._-]`)
)

// absoluteIn finds an absolute path in a value: the whole value one (/x, C:\x, \\host, a file: URL), or a word of
// it under a well-known root (/home/, /Users/, /tmp/, C:\ ...), so prose such as "the /api/users endpoint" is not
// taken for one. ~/ is a home's place, never absolute here.
func absoluteIn(s string) (string, bool) {
	v := strings.TrimSpace(s)
	if absWholeVal.MatchString(v) || absRoots.MatchString(v) {
		return firstWord(v), true
	}
	for _, w := range strings.FieldsFunc(s, func(r rune) bool { return strings.ContainsRune(" \t\"'`()[]{}<>,;=", r) }) {
		if absRoots.MatchString(w) {
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

// checkBudgets finds the block over its 40 lines and the memory index over 120 lines or 12 KB (spec §6, §10).
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
	lines := bytes.Count(raw, []byte("\n"))
	if len(raw) > 0 && !bytes.HasSuffix(raw, []byte("\n")) {
		lines++
	}
	if lines > MaxIndexLines || len(raw) > MaxIndexBytes {
		r.add("memory-index-size", index, "", fmt.Sprintf("the memory index %s is %d lines and %d bytes, over its fixed %d lines or 12 KB (spec section 10)",
			index, lines, len(raw), MaxIndexLines), "shorten it to one short line per note, at most 120 lines and 12 KB: merge notes, or move a note's detail into the note")
	}
}

var (
	importRef   = regexp.MustCompile(`(?:^|\s)@([^\s]+)`)
	linkRef     = regexp.MustCompile(`\]\(([^)\s]+)\)`)
	backtickRef = regexp.MustCompile("`([^`\\s]+)`")
	lineSuffix  = regexp.MustCompile(`:[0-9]+(-[0-9]+)?$`)
)

// checkNamedPaths finds project paths that CLAUDE.md (outside Bonsai's block), STATE and the memory notes name and
// that do not exist.
func (r *CheckResult) checkNamedPaths(files []string) {
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
			for _, m := range importRef.FindAllStringSubmatch(line, -1) {
				refs = append(refs, struct {
					p        string
					fromFile bool
				}{strings.TrimRight(m[1], ".,;:)"), true})
			}
			for _, m := range linkRef.FindAllStringSubmatch(line, -1) {
				refs = append(refs, struct {
					p        string
					fromFile bool
				}{m[1], true})
			}
			for _, m := range backtickRef.FindAllStringSubmatch(line, -1) {
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
				if _, err := os.Stat(filepath.Join(r.Root, filepath.FromSlash(target))); err == nil {
					continue
				}
				r.add("missing-path", file, "", fmt.Sprintf("%s line %d names %s, which does not exist in the project", file, n, ascii(ref.p)),
					fmt.Sprintf("fix line %d of %s: name the path that exists now, or take the reference out", n, file))
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
}

// projectRef gives a reference's project-relative path when it names a project path Bonsai can judge: no URL, no
// anchor alone, no home (~) or absolute path, no placeholder or glob. An @import or a link is read from the file's own
// folder (as Claude Code reads an import); a path in backticks from the project's top, and only when its first part
// is there, so a path of another repository's is not taken for this one's.
func projectRef(root, dir, ref string, fromFile bool) (string, bool) {
	if i := strings.IndexByte(ref, '#'); i >= 0 {
		ref = ref[:i]
	}
	ref = lineSuffix.ReplaceAllString(ref, "")
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

var runDate = regexp.MustCompile(`^R-([0-9]{4}-[0-9]{2}-[0-9]{2})-`)

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
			if m := runDate.FindStringSubmatch(path.Base(f)); m != nil {
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
