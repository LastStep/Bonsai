package engine

// The bug files and the bugs page (set 14; plan P-T-6023, steps 3 and 4; contract sections 7.3 and 7.5).
//
// The bug files (bonsai.bug/1) are documents of Bonsai's kind bug, one per bug, in the folder bonsai.yaml names under
// documents.bug (workspace.DocKinds gives it, with its default). check reads each one under its format (checkDocuments,
// the finding document), and holds them together to two rules a schema cannot say (checkBugFiles): no two files hold one
// id, and a file's name names its own id (bug-id); a fixed or wontfix bug has its closed day, unless it was carried
// over from an old register (was set) whose row named none (bug-closed).
//
// The bugs page (bonsai.bugs/1) lives at documents.bugs's place, the path a project's hand-kept register had. The
// marker rule: check --write builds it from the main checkout's bug files only while it is missing or carries the
// generated marker (its format line, format.BugsMarked); a page without it is a register not moved yet and is never
// overwritten, never read as the page, never a finding: check warns bugs-not-moved, naming base's bug skill. A marked
// page that differs from a rebuild is the warning tables, as the tasks table is. The page is byte-stable: rows by
// section (format.BugSections), newest id first in each, then by path; LF; no map order.
//
// A pack that still declares a kind named bugs (the workflow pack before its 1.1.0) is set aside, never refused
// (workspace.SetAsideKind): kind-set-aside names it and the step to a release without it.

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// bugDoc is one bug file as read: its path, and its frontmatter, nil when it does not read.
type bugDoc struct {
	path string
	bug  *format.Bug
}

// readBugFiles reads the bug files in main's folder dir (project-relative), sorted by path.
func readBugFiles(main, dir string) ([]bugDoc, error) {
	files, err := workspace.DocFiles(main, workspace.DocKind{Kind: "bug", Path: dir, ID: workspace.BugIDPattern})
	if err != nil {
		return nil, err
	}
	var out []bugDoc
	for _, p := range files {
		raw, exists, err := readFile(main, p)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}
		b, err := format.ReadBug(raw)
		if err != nil {
			b = nil
		}
		out = append(out, bugDoc{path: p, bug: b})
	}
	return out, nil
}

// bugNumber is a bug id's number: BUG-0145 is 145.
func bugNumber(id string) int {
	n, _ := strconv.Atoi(strings.TrimPrefix(id, "BUG-"))
	return n
}

// BuildBugsPage gives the bytes of the bugs page for the bug files in main's folder dir: one row per file that reads,
// in its status's section, newest id first, then by path. A file that does not read gives no row (check names it).
func BuildBugsPage(main, dir string) ([]byte, error) {
	docs, err := readBugFiles(main, dir)
	if err != nil {
		return nil, err
	}
	section := map[string]int{}
	for i, s := range format.BugSections {
		section[s.Status] = i
	}
	type row struct {
		sec, n int
		p      string
		row    format.BugRow
	}
	var rows []row
	for _, d := range docs {
		if d.bug == nil {
			continue
		}
		b := d.bug
		cell := func(p *string) *string {
			if p == nil {
				return nil
			}
			s := oneLineCell(*p)
			if s == "" {
				return nil
			}
			return &s
		}
		rows = append(rows, row{sec: section[b.Status], n: bugNumber(b.ID), p: d.path, row: format.BugRow{
			ID: b.ID, Title: oneLineCell(b.Title), Status: b.Status, Area: cell(b.Area), Where: cell(b.Where),
			FoundBy: cell(b.FoundBy), FixedIn: cell(b.FixedIn), Task: cell(b.Task), Closed: cell(b.Closed), Why: cell(b.Why)}})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].sec != rows[j].sec {
			return rows[i].sec < rows[j].sec
		}
		if rows[i].n != rows[j].n {
			return rows[i].n > rows[j].n
		}
		return rows[i].p < rows[j].p
	})
	page := &format.Bugs{Bugs: []format.BugRow{}}
	for _, r := range rows {
		page.Bugs = append(page.Bugs, r.row)
	}
	return page.Encode()
}

// bugPlaces are the bug files' folder and the bugs page's file of the checkout holding them: main's bonsai.yaml, else
// the checkout's own (as taskDir).
func bugPlaces(root, main string, cfg *workspace.Config) (folder, file string) {
	d := cfg.Full.Documents
	if main != root {
		if m, err := workspace.LoadConfigFull(main); err == nil {
			d = m.Full.Documents
		}
	}
	return d.BugFolder(), d.BugsFile()
}

// bugsPageState is what the main checkout's bugs page is: missing, marked (Bonsai's, rebuilt by check --write),
// unmarked (a hand-kept register, left alone), or too new (marked with a newer major, left alone).
type bugsPageState int

const (
	bugsMissing bugsPageState = iota
	bugsMarked
	bugsUnmarked
	bugsTooNew
)

func pageState(raw []byte, exists bool) bugsPageState {
	switch {
	case !exists:
		return bugsMissing
	case !format.BugsMarked(raw):
		return bugsUnmarked
	}
	if _, err := format.ReadBugs(raw); err != nil {
		if re, ok := err.(*format.ReadError); ok && re.TooNew {
			return bugsTooNew
		}
	}
	return bugsMarked
}

// notMovedNext is bugs-not-moved's next step.
const notMovedNext = "move the register to bug files, one per bug, by base's bug skill (skill base:bug); until then " +
	"it stays an ordinary file, never overwritten"

// checkBugsPage is the warning tables for the bugs page (a marked page that differs from a rebuild of the main
// checkout's bug files, or a missing one while there are bug files), and bugs-not-moved for a page without the marker.
func (r *CheckResult) checkBugsPage(next string) {
	folder, file := bugPlaces(r.Root, r.Main, r.Config)
	raw, exists, err := readFile(r.Main, file)
	if err != nil {
		return
	}
	whose := file
	if r.Root != r.Main {
		whose = "the main checkout's " + file
	}
	switch pageState(raw, exists) {
	case bugsUnmarked:
		r.add("bugs-not-moved", file, "", whose+" has no generated marker (format: "+format.MustLookup("bugs").Versioned()+
			"): a hand-kept bugs register, not moved to bug files yet, so check --write leaves it as it is", notMovedNext)
		return
	case bugsTooNew:
		return // the document finding names it
	case bugsMissing:
		if docs, err := readBugFiles(r.Main, folder); err != nil || len(docs) == 0 {
			return
		}
		r.add("tables", file, "", whose+" is missing, and there are bug files in "+folder, next)
		return
	}
	want, err := BuildBugsPage(r.Main, folder)
	if err != nil {
		return
	}
	if _, err := format.ReadBugs(raw); err != nil {
		return // the document finding names it
	}
	if string(lfBytes(raw)) != string(want) {
		r.add("tables", file, "", whose+" differs from a rebuild of the bug files (the page lags between moves; it grants nothing)", next)
	}
}

func lfBytes(b []byte) []byte { return []byte(strings.ReplaceAll(string(b), "\r\n", "\n")) }

// checkBugFiles holds the bug files of this checkout together: bug-id (two files hold one id; a file's name names
// another id than its own) and bug-closed (a fixed or wontfix bug with no closed day, its was not set). A file that
// does not read is the document finding's.
func (r *CheckResult) checkBugFiles(k workspace.DocKind) {
	docs, err := readBugFiles(r.Root, k.Path)
	if err != nil {
		return
	}
	idp := regexp.MustCompile(workspace.BugIDPattern)
	byID := map[string][]string{}
	var ids []string
	for _, d := range docs {
		if d.bug == nil {
			continue
		}
		b := d.bug
		if named := workspace.NameID(path.Base(d.path), idp); named != b.ID {
			r.add("bug-id", d.path, "", fmt.Sprintf("%s is named for %s and holds the id %s: a bug file is named after its own id", d.path, named, b.ID),
				"name the file "+b.ID+".md (or "+b.ID+"-<slug>.md), or give it the id its name says, then run: bonsai check")
		}
		if _, seen := byID[b.ID]; !seen {
			ids = append(ids, b.ID)
		}
		byID[b.ID] = append(byID[b.ID], d.path)
		if (b.Status == "fixed" || b.Status == "wontfix") && b.Closed == nil && b.Was == nil {
			r.add("bug-closed", d.path, "", d.path+" is "+b.Status+" with no closed day: only a bug carried over from an old register (was set) may lack one",
				"set closed to the day it was "+b.Status+" (bonsai now --day prints today, in UTC), then run: bonsai check")
		}
	}
	for _, id := range ids {
		if paths := byID[id]; len(paths) > 1 {
			r.add("bug-id", paths[1], "", fmt.Sprintf("%d bug files hold the id %s (%s): an id is one bug's, never reused", len(paths), id, strings.Join(paths, ", ")),
				"give all but one of them the next free id (the highest number in "+k.Path+" plus one), and its file that name, then run: bonsai check")
		}
	}
}

// checkSetAside is kind-set-aside: a locked pack declares a document kind that is Bonsai's own now (bugs, from set
// 14), which every reader of the kinds in force leaves out.
func (r *CheckResult) checkSetAside() {
	if r.Lock == nil {
		return
	}
	for _, lp := range r.Lock.Packs {
		d, err := lp.Declared()
		if err != nil {
			continue
		}
		for _, k := range d.Documents {
			if workspace.SetAsideKind(k.Kind) {
				r.add("kind-set-aside", workspace.LockFile, "", "the pack "+lp.ID+" declares the document kind "+k.Kind+
					", which is Bonsai's own from Bonsai 1.1.0: its declaration is set aside, and Bonsai's "+k.Kind+" is in force",
					"move "+lp.ID+" to a release that no longer declares it: bonsai pack outdated, then bonsai pack set "+lp.ID+"@<ref> (a preview; --yes writes it)")
			}
		}
	}
}
