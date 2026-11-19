package main

// bonsai bug (plan P-T-6023, steps 6 to 8; contract sections 7.3, 8.2 and 10): file a bug, move one, change its text,
// list them and show one. A thin command over internal/bugs, which holds every rule; with --json it prints
// bonsai.buglist/1 (format.BugList). Every writing subcommand writes the main checkout's bug folder wherever it runs,
// whatever this machine's status_writes reads (the switch is for tasks only: the director, 19 Nov), and logs one bug
// record; list and show read the main checkout's folder too.

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/bugs"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func init() { register(bugWord) }

var bugWord = &Word{
	Name:    "bug",
	Order:   20,
	Title:   "file a bug, move one, change its text, list them or show one (contract sections 7.3 and 10).",
	Summary: "file, move, edit, list and show the project's bugs, one file each",
	Args:    "<action>",
	About: `Each bug is its own file (bonsai.bug/1) in the bug folder bonsai.yaml names under documents.bug, named after its
id, BUG- and a number. bonsai bug is the one way a bug's status changes, in every project: add files one at the next
free id, fix, wontfix, watch and reopen move it, edit changes its text, list and show print them (agents: --json).
Every write goes to the main checkout's folder, wherever it runs, and is logged with who made it. bonsai check
--write builds the bugs page from the files.
`,
	Exits:    bugExits,
	Examples: []string{`bonsai bug add --title "The page forgets the dark mode switch" --area settings`, "bonsai bug fix BUG-012 --commit 0a1b2c3d", "bonsai bug list --status open --json"},
	Subs:     []*Word{bugAddWord, bugFixWord, bugWontfixWord, bugWatchWord, bugReopenWord, bugEditWord, bugListWord, bugShowWord},
}

var bugExits = []Exit{
	{Code: 0, Means: "written (its id and file are printed), or listed or shown"},
	{Code: 2, Means: "bad input: a flag or value wrong or missing, a commit the repository does not hold\n(commit-not-found), a bonsai.yaml Bonsai does not read. Nothing was written"},
	{Code: 3, Means: "the bug folder or the lock could not be read or written, or another bonsai bug held the lock\n(bug-busy). Nothing was written"},
	{Code: 4, Means: "not in a git checkout or not linked; no bug with that id (bug-not-found); a move the bug kind does\nnot make (bug-move); an id held already (bug-id-taken); a worktree whose main checkout is not verified\n(main-not-verified). Nothing was written"},
}

// bugBy is the --by flag every writing subcommand takes.
var bugBy = Flag{Name: "--by", Value: "ADDRESS", Help: "who writes: your address (role@task, or a role), or person; default BONSAI_ROLE@BONSAI_TASK,\nelse person in a person's terminal (an agent's session must give it)"}

var bugJSON = Flag{Name: "--json", Help: "print the bonsai.buglist/1 document (for programs) instead of plain text"}

// bugTextFlags are the text flags add and edit take.
var bugTextFlags = []Flag{
	{Name: "--area", Value: "TEXT", Help: "the part of the project it is in, a few words (\"\" for none)"},
	{Name: "--where", Value: "TEXT", Help: "where it goes wrong: a file and line, a command, a screen (\"\" for none)"},
	{Name: "--found-by", Value: "TEXT", Help: "who or what found it: a role and task, a message id, person, a CI run (\"\" for none)"},
	{Name: "--task", Value: "ID", Help: "the task it is fixed under (\"\" for none)"},
	{Name: "--why", Value: "TEXT", Help: "why, one sentence (\"\" for none)"},
	{Name: "--body", Value: "TEXT", Help: "its body, markdown: how to repeat it, what was seen; - reads it from standard input"},
	{Name: "--body-file", Value: "PATH", Help: "or its body from a file"},
}

const bugWriteNotes = `Every write lands in the main checkout's bug folder, wherever the command runs (a worktree's copy is never
written), under the lock .bonsai/local/bug.lock, and appends one bug record to the log: the bug's id, the subcommand,
what changed and who. The status, its closed day and the move's fields are bonsai bug's to write in every project,
whatever this machine's status_writes reads; the guard refuses a hand edit of a bug's status line.
`

var bugAddWord = &Word{
	Name:    "bug add",
	Title:   "file a bug: the next free id, its file in the main checkout's bug folder, at open",
	Summary: "file a new bug, or carry one over from an old register (--was)",
	Args:    "--title <text> [flags]",
	About: `It takes the highest number in the bug folder plus one (never one a file holds), writes <id>.md at status open
with created today (the UTC day) and filed_by who writes, and prints the id and the file.
To carry a bug over from an old register (base:bug), give --was its old id: then --number keeps its number,
--status where it stood, --commit what fixed it (as the old row gave it), --created and --closed its days.
`,
	Flags: append(append([]Flag{{Name: "--title", Value: "TEXT", Help: "what goes wrong, as a person sees it, on one line"}}, bugTextFlags...),
		Flag{Name: "--was", Value: "ID", Help: "carry-over: the bug's id in the old register (TB-012)"},
		Flag{Name: "--number", Value: "N", Help: "carry-over: the number to keep (12 is BUG-012); refused when a file holds it"},
		Flag{Name: "--status", Value: "STATUS", Help: "carry-over: where it stood: open, fixed, wontfix or watched"},
		Flag{Name: "--commit", Value: "TEXT", Help: "carry-over: what fixed it, as the old row gave it (its fixed_in)"},
		Flag{Name: "--created", Value: "DAY", Help: "carry-over: the day the old row names, YYYY-MM-DD (else today)"},
		Flag{Name: "--closed", Value: "DAY", Help: "carry-over: the day it was fixed or set wontfix, YYYY-MM-DD (else none)"},
		bugBy, bugJSON),
	Notes: bugWriteNotes,
	Exits: bugExits,
	Examples: []string{`bonsai bug add --title "The page forgets the dark mode switch" --area settings --where "src/settings/store.ts:88"`,
		"bonsai bug add --was TB-012 --number 12 --status fixed --commit 0a1b2c3d --title \"The switch is not saved\""},
	Refused: bugRefused,
	Run:     runBugAdd,
}

var bugFixWord = &Word{
	Name:    "bug fix",
	Title:   "mark a bug fixed: the commit that fixed it, who, and the day",
	Summary: "mark a bug fixed by a commit",
	Args:    "<id> --commit <sha> [flags]",
	About: `It moves the bug (open or watched) to fixed: fixed_in the commits, each one the repository holds, with --how after
them; fixed_by who writes; closed today (the UTC day). The id may be the bug's old id (its was).
`,
	Flags: []Flag{{Name: "--commit", Value: "SHA", Many: true, Help: "a commit that fixed it, one the repository holds (again for more)"},
		{Name: "--how", Value: "TEXT", Help: "a few words on how, written after the commits"},
		{Name: "--task", Value: "ID", Help: "the task that fixed it"}, bugBy, bugJSON},
	Notes:    bugWriteNotes,
	Exits:    bugExits,
	Examples: []string{"bonsai bug fix BUG-012 --commit 0a1b2c3d --how \"the switch is saved before the page closes\" --task T-0904"},
	Refused:  bugRefused,
	Run:      func(c *call) int { return runBugMove(c, "fix", "fixed") },
}

var bugWontfixWord = &Word{
	Name:     "bug wontfix",
	Title:    "mark a bug wontfix: known, not a bug, or not to be fixed, and why",
	Summary:  "mark a bug wontfix, saying why",
	Args:     "<id> --why <text> [flags]",
	About:    "It moves the bug (open) to wontfix, why its reason, closed today (the UTC day).\n",
	Flags:    []Flag{{Name: "--why", Value: "TEXT", Help: "why, one sentence"}, bugBy, bugJSON},
	Notes:    bugWriteNotes,
	Exits:    bugExits,
	Examples: []string{`bonsai bug wontfix BUG-013 --why "A deliberate over-refusal, documented in the guard's tests"`},
	Refused:  bugRefused,
	Run:      func(c *call) int { return runBugMove(c, "wontfix", "wontfix") },
}

var bugWatchWord = &Word{
	Name:     "bug watch",
	Title:    "mark a bug watched: seen once and not repeated, and why",
	Summary:  "mark a bug watched (seen once, a flaky test among them), saying why",
	Args:     "<id> --why <text> [flags]",
	About:    "It moves the bug (open) to watched, why its reason.\n",
	Flags:    []Flag{{Name: "--why", Value: "TEXT", Help: "why, one sentence"}, bugBy, bugJSON},
	Notes:    bugWriteNotes,
	Exits:    bugExits,
	Examples: []string{`bonsai bug watch BUG-014 --why "Seen once in CI's windows job, not repeated in 20 runs"`},
	Refused:  bugRefused,
	Run:      func(c *call) int { return runBugMove(c, "watch", "watched") },
}

var bugReopenWord = &Word{
	Name:     "bug reopen",
	Title:    "reopen a bug: it came back, or watching it is over",
	Summary:  "move a fixed, wontfix or watched bug back to open",
	Args:     "<id> [flags]",
	About:    "It moves the bug back to open: fixed_by, fixed_in and closed cleared, why --why's (else none).\n",
	Flags:    []Flag{{Name: "--why", Value: "TEXT", Help: "why it is open again, one sentence"}, bugBy, bugJSON},
	Notes:    bugWriteNotes,
	Exits:    bugExits,
	Examples: []string{`bonsai bug reopen BUG-012 --why "Seen again after the 1.1 build"`},
	Refused:  bugRefused,
	Run:      func(c *call) int { return runBugMove(c, "reopen", "open") },
}

var bugEditWord = &Word{
	Name:     "bug edit",
	Title:    "change a bug's text: its title, area, where, found_by, task, why or body, never its status",
	Summary:  "change a bug's text fields or body",
	Args:     "<id> [flags]",
	About:    "It changes the fields given (\"\" sets one to none) and leaves the rest as they are; a move is fix,\nwontfix, watch or reopen.\n",
	Flags:    append(append([]Flag{{Name: "--title", Value: "TEXT", Help: "the new title, on one line"}}, bugTextFlags...), bugJSON),
	Notes:    bugWriteNotes,
	Exits:    bugExits,
	Examples: []string{`bonsai bug edit BUG-012 --where "src/settings/store.ts:91 (pointed)"`},
	Refused:  bugRefused,
	Run:      runBugEdit,
}

var bugListWord = &Word{
	Name:     "bug list",
	Title:    "list the bugs: one line each, oldest id first",
	Summary:  "list the project's bugs (--status to keep some)",
	Args:     "[flags]",
	About:    "It reads the main checkout's bug folder and prints each bug's id, status and title; a file that does not read is named.\n",
	Flags:    []Flag{{Name: "--status", Value: "STATUS", Many: true, Help: "keep only the bugs at this status (again for more)"}, bugJSON},
	Exits:    bugExits,
	Examples: []string{"bonsai bug list --status open", "bonsai bug list --json"},
	Refused:  bugRefused,
	Run:      runBugList,
}

var bugShowWord = &Word{
	Name:     "bug show",
	Title:    "show one bug: its file's fields and body",
	Summary:  "show one bug, by its id or its old id",
	Args:     "<id> [flags]",
	About:    "It prints the bug's file as it stands in the main checkout. The id may be its old id (its was).\n",
	Flags:    []Flag{bugJSON},
	Exits:    bugExits,
	Examples: []string{"bonsai bug show BUG-012", "bonsai bug show TB-012 --json"},
	Refused:  bugRefused,
	Run:      runBugShow,
}

// bugSub is the subcommand a bug word is: "add" for bug add.
func bugSub(w *Word) string { return strings.TrimPrefix(w.Name, "bug ") }

// bugRefused is bonsai.buglist/1 for a refusal before bonsai.yaml was read: the workspace null.
func bugRefused(c *call, e *engine.Error) encoder {
	sub := bugSub(c.word)
	doc := &format.BugList{Bugs: []format.BugEntry{}, Malformed: []format.BugMalformed{}, Error: e.Object()}
	if sub != c.word.Name {
		doc.Command = &sub
	}
	return doc
}

// bugSite is where a bug command reads and writes: the main checkout's bug folder (bugs.Place) and the workspace.
type bugSite struct {
	site  *asksSite
	place bugs.Place
}

// bugPlace finds the checkout, its main checkout (a worktree's must be verified: its copy is never written) and the
// bug folder its bonsai.yaml names (main's, as check reads it).
func bugPlace() (*bugSite, *engine.Error) {
	site, e := askSite()
	if e != nil {
		return nil, e
	}
	local := site.place.Local
	if workspace.ReadGitLayout(site.co.Root).Worktree() && samePath(local.Main, local.Root) {
		return nil, &engine.Error{Code: "main-not-verified", Exit: engine.ExitState, Who: "person",
			What: "this is a worktree whose main checkout is not verified on this machine, and bonsai bug writes and reads only the main checkout's bug folder",
			Next: "run bonsai bug from the main checkout, or have a person verify it: bonsai status names the step"}
	}
	cfg := site.cfg
	if !samePath(local.Main, site.co.Root) {
		if m, err := workspace.LoadConfigFull(local.Main); err == nil {
			cfg = m
		}
	}
	return &bugSite{site: site, place: bugs.Place{Main: local.Main, Folder: cfg.Full.Documents.BugFolder(), Workspace: site.cfg.ID,
		Local: local, Session: os.Getenv(asks.SessionEnv), Now: bugClock}}, nil
}

// bugClock is the clock bonsai bug reads its days and records from; a test sets it.
var bugClock = time.Now

// samePath compares two absolute folders, cleaned, letter case ignored on Windows.
func samePath(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// bugError is an internal/bugs refusal as the command's.
func bugError(e *bugs.Error) *engine.Error {
	return &engine.Error{Code: e.Code, Exit: e.Exit, What: e.What, Next: e.Next, Who: e.Who}
}

// bugFail is c.fail once bonsai.yaml was read: with --json the document names the workspace.
func (c *call) bugFail(s *bugSite, e *engine.Error) int {
	if !c.json {
		return c.fail(e)
	}
	noted(c.word, e)
	doc := bugRefused(c, e).(*format.BugList)
	doc.Workspace = s.site.ref
	return c.printDoc(doc, e.Exit)
}

// bugEntry is one bug as --json prints it; with body, its body too.
func bugEntry(en bugs.Entry, body bool) (format.BugEntry, error) {
	obj, err := en.Bug.Object()
	if err != nil {
		return format.BugEntry{}, err
	}
	out := format.BugEntry{Path: en.Path, Bug: obj}
	if body {
		b := en.Body
		out.Body = &b
	}
	return out, nil
}

// bugOut prints what a subcommand read or wrote: --json's document, else lines for a person.
func (c *call) bugOut(s *bugSite, entries []bugs.Entry, bad []bugs.Malformed, body bool, text string) int {
	if !c.json {
		return c.print(text)
	}
	sub := bugSub(c.word)
	doc := &format.BugList{Workspace: s.site.ref, Command: &sub, Bugs: []format.BugEntry{}, Malformed: []format.BugMalformed{}}
	for _, en := range entries {
		e, err := bugEntry(en, body)
		if err != nil {
			return c.bugFail(s, &engine.Error{Code: "unexpected", Exit: exitRuntime, What: engine.ASCII(err.Error()),
				Next: "run the command again; if it fails again, report it to Bonsai's maintainers"})
		}
		doc.Bugs = append(doc.Bugs, e)
	}
	for _, m := range bad {
		doc.Malformed = append(doc.Malformed, format.BugMalformed{Path: m.Path, Reason: engine.ASCII(m.Reason)})
	}
	return c.printDoc(doc, exitOK)
}

// written prints a write's answer: the id and the file, and a note when its log record could not be written.
func (c *call) bugWritten(s *bugSite, w *bugs.Written) int {
	text := w.Entry.Bug.ID + "\n" + engine.ASCII(w.Entry.Path) + "\n"
	if !w.Logged && !c.json {
		_, _ = fmt.Fprintf(c.stderr, "bonsai %s: the bug was written, but its log record could not be: bonsai check names the log's folder\n", c.word.Name)
	}
	return c.bugOut(s, []bugs.Entry{w.Entry}, nil, true, text)
}

// bugFlag reads a text flag: nil when not given.
func (c *call) bugFlag(name string) *string {
	if !c.has(name) {
		return nil
	}
	v := c.value(name)
	return &v
}

// bugFields reads add's and edit's text flags.
func (c *call) bugFields() (bugs.Fields, *engine.Error) {
	f := bugs.Fields{Title: c.bugFlag("--title"), Area: c.bugFlag("--area"), Where: c.bugFlag("--where"),
		FoundBy: c.bugFlag("--found-by"), Task: c.bugFlag("--task"), Why: c.bugFlag("--why")}
	if c.has("--body") && c.has("--body-file") {
		e := c.flagError("%s takes --body or --body-file, not both", c.word.Name)
		e.Next = "give the body once: --body <text>, --body - (standard input) or --body-file <path>"
		return f, e
	}
	switch {
	case c.has("--body") && c.value("--body") == "-":
		raw, err := io.ReadAll(stdin)
		if err != nil {
			return f, &engine.Error{Code: "read-failed", Exit: exitRuntime, What: "standard input cannot be read: " + engine.ASCII(err.Error()),
				Next: "give the body with --body <text> or --body-file <path>"}
		}
		s := string(raw)
		f.Body = &s
	case c.has("--body"):
		f.Body = c.bugFlag("--body")
	case c.has("--body-file"):
		path := c.value("--body-file")
		raw, err := os.ReadFile(path)
		if err != nil {
			e := &engine.Error{Code: "read-failed", Exit: exitRuntime, What: "--body-file " + engine.ASCII(path) + " cannot be read: " + engine.ASCII(err.Error()),
				Next: "check the file's permissions, then run the command again"}
			if errors.Is(err, fs.ErrNotExist) {
				e = &engine.Error{Code: "bad-value", Exit: exitInput, What: "--body-file " + engine.ASCII(path) + " names no file that is there",
					Next: "give --body-file the path of the file holding the body, or --body <text>"}
			}
			return f, e
		}
		s := string(raw)
		f.Body = &s
	}
	return f, nil
}

// bugByOf is who writes, from --by or the environment.
func (c *call) bugByOf() (string, *engine.Error) {
	by, e := bugs.By(c.value("--by"), os.Getenv)
	if e != nil {
		return "", bugError(e)
	}
	return by, nil
}

func runBugAdd(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("bug add takes no %+q (the title goes in --title)", c.rest[0]))
	}
	fields, e := c.bugFields()
	if e != nil {
		return c.refuse(e)
	}
	by, e := c.bugByOf()
	if e != nil {
		return c.refuse(e)
	}
	in := bugs.AddInput{Fields: fields, By: by, Was: c.value("--was"), Status: c.value("--status"), Commit: c.value("--commit"),
		Created: c.value("--created"), Closed: c.value("--closed")}
	if c.has("--number") {
		n, err := strconv.Atoi(c.value("--number"))
		if err != nil || n <= 0 {
			return c.refuse(&engine.Error{Code: "bad-value", Exit: exitInput, What: "--number " + engine.ASCII(c.value("--number")) + " is not a number from 1 to 999999",
				Next: "give --number the old id's number (TB-012: --number 12)"})
		}
		in.Number = n
	}
	if c.has("--was") && in.Was == "" {
		return c.refuse(&engine.Error{Code: "bad-value", Exit: exitInput, What: "--was is empty", Next: "give --was the bug's id in the old register (TB-012)"})
	}
	s, e := bugPlace()
	if e != nil {
		return c.fail(e)
	}
	s.place.Getenv = os.Getenv
	w, be := bugs.Add(s.place, in)
	if be != nil {
		return c.bugFail(s, bugError(be))
	}
	return c.bugWritten(s, w)
}

func runBugMove(c *call, sub, to string) int {
	if len(c.rest) != 1 {
		return c.refuse(c.flagError("bug %s takes one bug id", sub))
	}
	by, e := c.bugByOf()
	if e != nil {
		return c.refuse(e)
	}
	s, e := bugPlace()
	if e != nil {
		return c.fail(e)
	}
	s.place.Getenv = os.Getenv
	in := bugs.MoveInput{By: by, Commits: c.all("--commit"), How: c.value("--how"), Why: c.bugFlag("--why"), Task: c.bugFlag("--task")}
	w, be := bugs.Move(s.place, sub, c.rest[0], to, in)
	if be != nil {
		return c.bugFail(s, bugError(be))
	}
	return c.bugWritten(s, w)
}

func runBugEdit(c *call) int {
	if len(c.rest) != 1 {
		return c.refuse(c.flagError("bug edit takes one bug id"))
	}
	fields, e := c.bugFields()
	if e != nil {
		return c.refuse(e)
	}
	s, e := bugPlace()
	if e != nil {
		return c.fail(e)
	}
	s.place.Getenv = os.Getenv
	w, be := bugs.Edit(s.place, c.rest[0], fields)
	if be != nil {
		return c.bugFail(s, bugError(be))
	}
	return c.bugWritten(s, w)
}

func runBugList(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("bug list takes no %+q", c.rest[0]))
	}
	keep := c.all("--status")
	for _, st := range keep {
		if k, _ := workspace.BonsaiDocKind("bug"); !slices.Contains(k.Statuses, st) {
			return c.refuse(&engine.Error{Code: "bad-value", Exit: exitInput, What: "--status " + engine.ASCII(st) + " is not a bug's status",
				Next: "give --status one of " + strings.Join(k.Statuses, ", ")})
		}
	}
	s, e := bugPlace()
	if e != nil {
		return c.fail(e)
	}
	all, bad, err := bugs.Read(s.place)
	if err != nil {
		return c.bugFail(s, &engine.Error{Code: "read-failed", Exit: exitRuntime, What: "the bug folder " + s.place.Folder + " cannot be read: " + engine.ASCII(err.Error()),
			Next: "check the folder's permissions, then run the command again"})
	}
	var kept []bugs.Entry
	for _, en := range all {
		if len(keep) == 0 || slices.Contains(keep, en.Bug.Status) {
			kept = append(kept, en)
		}
	}
	var b strings.Builder
	if len(kept) == 0 {
		b.WriteString("no bugs in " + s.place.Folder + "\n")
	}
	for _, en := range kept {
		fmt.Fprintf(&b, "%s  %-7s  %s\n", en.Bug.ID, en.Bug.Status, engine.ASCII(en.Bug.Title))
	}
	for _, m := range bad {
		fmt.Fprintf(&b, "%s does not read: %s (bonsai check names it)\n", engine.ASCII(m.Path), engine.ASCII(m.Reason))
	}
	return c.bugOut(s, kept, bad, false, b.String())
}

func runBugShow(c *call) int {
	if len(c.rest) != 1 {
		return c.refuse(c.flagError("bug show takes one bug id"))
	}
	s, e := bugPlace()
	if e != nil {
		return c.fail(e)
	}
	all, _, err := bugs.Read(s.place)
	if err != nil {
		return c.bugFail(s, &engine.Error{Code: "read-failed", Exit: exitRuntime, What: "the bug folder " + s.place.Folder + " cannot be read: " + engine.ASCII(err.Error()),
			Next: "check the folder's permissions, then run the command again"})
	}
	en, be := bugs.Find(all, c.rest[0])
	if be != nil {
		return c.bugFail(s, bugError(be))
	}
	raw, err := en.Bug.EncodeFile(en.Body)
	if err != nil {
		raw = []byte(en.Bug.ID + "\n")
	}
	return c.bugOut(s, []bugs.Entry{*en}, nil, true, engine.ASCII(en.Path)+"\n"+engine.ASCII(string(raw)))
}
