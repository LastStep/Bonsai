// Package bugs is bonsai bug's work (plan P-T-6023, steps 6 to 8; contract sections 7.3, 8.2 and 10): reading the bug
// files (bonsai.bug/1) of a project's bug folder, and writing them: add files a new bug at the next free id, the moves
// (fix, wontfix, watch, reopen) change where one stands, edit changes its text. cmd/bonsai/bug.go is the thin command
// over it.
//
// Where: every write goes to the main checkout's bug folder (bonsai.yaml's documents.bug, else bugs/ beside
// documents.bugs's page), wherever the command runs: a worktree's copy is never written, as tasks are held to main
// (Place.Main is the main checkout the caller found and verified).
//
// The lock: every writing subcommand holds .bonsai/local/bug.lock in the main checkout while it reads the folder and
// writes one file, so two adds never take one id and two moves of one bug never lose one. The lock file is made with
// O_EXCL; one older than StaleLock is taken over (its writer died); while a live one is held, a writer waits up to
// LockWait, then refuses (bug-busy). It is removed when the write ends (its removal retried while Windows reports it
// busy).
//
// The moves are the bug kind's (workspace.BugMoves, their one home): open to fixed, wontfix or watched; watched to open
// or fixed; fixed and wontfix back to open. bonsai bug writes the status, its stamp (closed, the UTC day, set by fixed
// and wontfix and cleared by the others) and the move's fields itself, whatever this machine's status_writes reads: the
// switch is for tasks only (the director, 19 Nov). Every write appends one bug record to the log (contract section 8.2):
// target the bug's id, kind the subcommand, reason the move, task and role whoever ran it.
package bugs

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// LockFile is the id lock, relative to the main checkout (format.GeneratedKinds' bug-lock).
const LockFile = workspace.LocalDir + "/bug.lock"

// StaleLock is the age past which a lock file is taken over, and LockWait how long a writer waits for a live one.
var (
	StaleLock = time.Minute
	LockWait  = 5 * time.Second
)

// Place is where the bug files live and who writes.
type Place struct {
	Main      string          // the main checkout's top folder, absolute: every write lands here
	Folder    string          // the bug folder, project-relative, forward slashes (documents.bug)
	Workspace string          // the workspace id, for the log record
	Local     workspace.Local // where the writer runs: the record's checkout and branch
	Session   string          // CLAUDE_CODE_SESSION_ID, "" for none: the record's file and session
	Getenv    func(string) string
	Now       func() time.Time
}

func (p Place) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}

// Error is a refusal or failure: its error word (format.ErrorWords), the exit code, what went wrong and the next step.
type Error struct {
	Code string
	Exit int // 2 bad input, 3 runtime, 4 wrong state
	What string
	Next string
	Who  string // "person" for a person's step; "" is the agent's
}

func (e *Error) Error() string { return e.What }

func refuse(code string, exit int, next, what string, args ...any) *Error {
	return &Error{Code: code, Exit: exit, What: fmt.Sprintf(what, args...), Next: next}
}

// Entry is one bug file that reads: its path (project-relative), its frontmatter and its body.
type Entry struct {
	Path string
	Bug  *format.Bug
	Body string
}

// Malformed is a file in the folder named by a bug id that does not read as bonsai.bug/1.
type Malformed struct {
	Path   string
	Reason string
}

var idPattern = regexp.MustCompile(workspace.BugIDPattern)

// Number is a bug id's number: BUG-0145 is 145; -1 for no bug id.
func Number(id string) int {
	if !idPattern.MatchString(id) {
		return -1
	}
	n, err := strconv.Atoi(strings.TrimPrefix(id, "BUG-"))
	if err != nil {
		return -1
	}
	return n
}

// ID is the bug id of a number: three digits at least (BUG-007), six at most.
func ID(n int) string { return fmt.Sprintf("BUG-%03d", n) }

// Read reads the bug folder of the main checkout: every file named by a bug id, in id order (by number, then path).
// A missing folder holds no bugs.
func Read(p Place) ([]Entry, []Malformed, error) {
	files, err := workspace.DocFiles(p.Main, workspace.DocKind{Kind: "bug", Path: p.Folder, ID: workspace.BugIDPattern})
	if err != nil {
		return nil, nil, err
	}
	var out []Entry
	var bad []Malformed
	for _, rel := range files {
		raw, err := os.ReadFile(filepath.Join(p.Main, filepath.FromSlash(rel)))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		b, body, err := format.ReadBugFile(raw)
		if err != nil {
			bad = append(bad, Malformed{Path: rel, Reason: oneLine(err.Error())})
			continue
		}
		out = append(out, Entry{Path: rel, Bug: b, Body: body})
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := Number(out[i].Bug.ID), Number(out[j].Bug.ID)
		if a != b {
			return a < b
		}
		return out[i].Path < out[j].Path
	})
	return out, bad, nil
}

// Find gives the bug a key names: its id (BUG-145), or the id it had in an old register (its was: TB-012). An id
// two files hold is refused (bonsai check names them).
func Find(entries []Entry, key string) (*Entry, *Error) {
	var hits []*Entry
	for i := range entries {
		if entries[i].Bug.ID == key {
			hits = append(hits, &entries[i])
		}
	}
	if len(hits) == 0 {
		for i := range entries {
			if w := entries[i].Bug.Was; w != nil && *w == key {
				hits = append(hits, &entries[i])
			}
		}
	}
	switch len(hits) {
	case 0:
		return nil, refuse("bug-not-found", 4, "list the bugs to find its id: bonsai bug list",
			"no bug file in the bug folder has the id (or the old id) %s", strconv.QuoteToASCII(key))
	case 1:
		return hits[0], nil
	}
	var paths []string
	for _, h := range hits {
		paths = append(paths, h.Path)
	}
	return nil, refuse("bug-not-found", 4, "run bonsai check, which names the files that share an id, and give all but one another id",
		"%d bug files answer to %s (%s), so bonsai bug cannot tell which one is meant", len(hits), strconv.QuoteToASCII(key),
		strings.Join(paths, ", "))
}

// ---------------------------------------------------------------- who

// By is who writes: by as given (an agent's address, or person), else BONSAI_ROLE and BONSAI_TASK (role@task, or the
// role), else person in a person's terminal (no agent marker set). An agent's session with neither is refused: an
// agent says who it is with --by, and never writes person.
func By(by string, getenv func(string) string) (string, *Error) {
	marker, _ := format.AgentMarker(getenv)
	if by == "" {
		if role := getenv("BONSAI_ROLE"); role != "" {
			if task := getenv("BONSAI_TASK"); task != "" {
				return role + "@" + task, nil
			}
			return role, nil
		}
		if marker != "" {
			return "", refuse("missing-value", 2, "give your own address, as your brief's to names it: --by <role>@<task> (or --by <role>)",
				"an agent's session (%s is set) must say who writes the bug: no --by, and no BONSAI_ROLE", marker)
		}
		return "person", nil
	}
	if strings.ContainsAny(by, "\r\n") || len(by) > 100 {
		return "", refuse("bad-value", 2, "give --by an address on one line: <role>@<task>, <role>, or person",
			"--by %s is not an address on one line of at most 100 characters", strconv.QuoteToASCII(by))
	}
	if by == "person" && marker != "" {
		return "", refuse("bad-value", 2, "give your own address: --by <role>@<task> (or --by <role>)",
			"--by person in an agent's session (%s is set): an agent never writes as a person", marker)
	}
	return by, nil
}

// ---------------------------------------------------------------- the lock

// lock takes the id lock in the main checkout, or refuses (bug-busy) once a live one has been held for LockWait. Its
// release removes the file.
func lock(p Place) (func(), *Error) {
	file := filepath.Join(p.Main, filepath.FromSlash(LockFile))
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return nil, refuse("write-failed", 3, "check that the main checkout's .bonsai/local/ can be written, then run the command again",
			"the bug lock's folder cannot be made: %s", oneLine(err.Error()))
	}
	deadline := time.Now().Add(LockWait) // the real clock: the lock is a file's age, never a test's clock
	delay := 5 * time.Millisecond
	for {
		f, err := os.OpenFile(file, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = fmt.Fprintf(f, "pid %d at %s\n", os.Getpid(), time.Now().UTC().Format(record.AtLayout))
			_ = f.Close()
			return func() { _ = workspace.RemoveFile(file) }, nil
		}
		if !errors.Is(err, fs.ErrExist) && !workspace.IsBusy(err) && !(runtime.GOOS == "windows" && errors.Is(err, fs.ErrPermission)) {
			return nil, refuse("write-failed", 3, "check that the main checkout's .bonsai/local/ can be written, then run the command again",
				"the bug lock %s cannot be made: %s", LockFile, oneLine(err.Error()))
		}
		if st, serr := os.Stat(file); serr == nil && time.Since(st.ModTime()) > StaleLock {
			_ = workspace.RemoveFile(file) // stale: its writer died; the next try takes it
			continue
		}
		if time.Now().After(deadline) {
			return nil, refuse("bug-busy", 3, "run the command again in a few seconds; a lock older than a minute is taken over",
				"another bonsai bug is writing in this project (%s is held)", LockFile)
		}
		time.Sleep(delay)
		if delay < 160*time.Millisecond {
			delay *= 2
		}
	}
}

// ---------------------------------------------------------------- writes

// Written is what a write did: the bug as written, and the log record's reason.
type Written struct {
	Entry  Entry
	Reason string
	Logged bool // its log record was written
}

// Fields are a bug's text fields as a command gives them: nil leaves one as it is (edit) or null (add); a pointer to
// "" sets a nullable one to null.
type Fields struct {
	Title, Area, Where, FoundBy, Task, Why *string
	Body                                   *string
}

// AddInput is what add files. Was, Number, Status, Commit, Created and Closed are the carry-over's (a bug moved from
// an old register): Number, Status, Created and Closed are taken only with Was.
type AddInput struct {
	Fields
	By      string
	Was     string
	Number  int    // 0: the next free number
	Status  string // "": open
	Commit  string // the carry-over's fixed_in, as the old row gave it
	Created string // YYYY-MM-DD, "" for today
	Closed  string // YYYY-MM-DD, "" for none
}

var dayPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

// Add files a bug: the next free id (the highest number in the folder plus one, never one a file holds) under the
// lock, its file written whole, and its log record.
func Add(p Place, in AddInput) (*Written, *Error) {
	if in.Title == nil || strings.TrimSpace(*in.Title) == "" {
		return nil, refuse("missing-value", 2, "give the bug a title: --title <what goes wrong, on one line>", "add needs --title")
	}
	carried := in.Was != ""
	if !carried && (in.Number != 0 || in.Status != "" || in.Commit != "" || in.Created != "" || in.Closed != "") {
		return nil, refuse("bad-flag", 2, "file a new bug with --title and its text only; carry an old one over with --was <old id> (then --number, --status, --commit, --created and --closed)",
			"--number, --status, --commit, --created and --closed carry a bug over from an old register, and are taken only with --was")
	}
	status := "open"
	if in.Status != "" {
		status = in.Status
	}
	if !contains(statuses(), status) {
		return nil, refuse("bad-value", 2, "give --status one of "+strings.Join(statuses(), ", "), "--status %s is not a bug's status", strconv.QuoteToASCII(status))
	}
	for _, d := range []struct{ flag, v string }{{"--created", in.Created}, {"--closed", in.Closed}} {
		if d.v != "" && !validDay(d.v) {
			return nil, refuse("bad-value", 2, "give "+d.flag+" a day as YYYY-MM-DD (the UTC day)", "%s %s is not a day YYYY-MM-DD", d.flag, strconv.QuoteToASCII(d.v))
		}
	}
	if in.Closed != "" && status != "fixed" && status != "wontfix" {
		return nil, refuse("bad-flag", 2, "give --closed only with --status fixed or wontfix", "--closed with --status %s: only a fixed or wontfix bug has a closed day", status)
	}
	if in.Number < 0 || in.Number > 999999 {
		return nil, refuse("bad-value", 2, "give --number a number from 1 to 999999", "--number %d is not a bug's number", in.Number)
	}
	if carried && strings.ContainsAny(in.Was, "\r\n") {
		return nil, refuse("bad-value", 2, "give --was the old id on one line", "--was %s is not on one line", strconv.QuoteToASCII(in.Was))
	}
	release, e := lock(p)
	if e != nil {
		return nil, e
	}
	defer release()
	entries, bad, err := Read(p)
	if err != nil {
		return nil, readError(p, err)
	}
	high := 0
	taken := map[int]bool{}
	note := func(n int) {
		if n > high {
			high = n
		}
		taken[n] = true
	}
	for _, en := range entries {
		note(Number(en.Bug.ID))
		note(Number(workspace.NameID(path.Base(en.Path), idPattern)))
	}
	for _, m := range bad {
		note(Number(workspace.NameID(path.Base(m.Path), idPattern)))
	}
	n := in.Number
	switch {
	case n == 0:
		n = high + 1
	case taken[n]:
		return nil, refuse("bug-id-taken", 4, "carry it over at another number, or leave --number out for the next free one ("+ID(high+1)+")",
			"%s is held by a file in %s already: an id is one bug's, never reused", ID(n), p.Folder)
	}
	if carried {
		for _, en := range entries {
			if en.Bug.Was != nil && *en.Bug.Was == in.Was {
				return nil, refuse("bug-id-taken", 4, "show it: bonsai bug show "+in.Was,
					"the old id %s was carried over already, as %s (%s)", strconv.QuoteToASCII(in.Was), en.Bug.ID, en.Path)
			}
		}
	}
	if n > 999999 {
		return nil, refuse("bug-id-taken", 4, "tell the person: the bug folder holds the highest id a bug may have", "no bug id is left above %s", ID(high))
	}
	today := p.now().UTC().Format("2006-01-02")
	b := &format.Bug{ID: ID(n), Title: oneLine(*in.Title), Status: status, Created: today, Labels: schema.Object{}}
	b.Area, b.Where, b.FoundBy, b.Task, b.Why = opt(in.Area), opt(in.Where), opt(in.FoundBy), opt(in.Task), opt(in.Why)
	if e := taskValue(b.Task); e != nil {
		return nil, e
	}
	by := in.By
	b.FiledBy = &by
	reason := "filed (" + status + ")"
	if carried {
		was := in.Was
		b.Was = &was
		if in.Created != "" {
			b.Created = in.Created
		}
		if in.Closed != "" {
			c := in.Closed
			b.Closed = &c
		}
		if in.Commit != "" {
			c := in.Commit
			b.FixedIn = &c
		}
		reason = "carried over from " + in.Was + " (" + status + ")"
	}
	body := ""
	if in.Body != nil {
		body = *in.Body
	}
	rel := strings.TrimSuffix(p.Folder, "/") + "/" + b.ID + ".md"
	w, e := write(p, rel, b, body, true)
	if e != nil {
		return nil, e
	}
	w.Reason = reason
	w.Logged = logRecord(p, "add", b, reason) == nil
	return w, nil
}

// MoveInput is a move's fields: the commits that fixed it (fix), why (wontfix, watch; reopen may say), the task.
type MoveInput struct {
	By      string
	Commits []string // fix: each a commit the repository holds
	How     string   // fix: a few words on how, after the commits
	Why     *string
	Task    *string
}

// Move moves the bug key names to the status to (fixed, wontfix, watched, open), if the bug kind allows it from where
// it stands, writing the status, its stamp and the move's fields, under the lock, then the log record.
func Move(p Place, sub, key, to string, in MoveInput) (*Written, *Error) {
	switch to {
	case "fixed":
		if len(in.Commits) == 0 {
			return nil, refuse("missing-value", 2, "name the commit that fixed it: --commit <sha> (more than one: --commit again)", "fix needs --commit")
		}
	case "wontfix", "watched":
		if in.Why == nil || strings.TrimSpace(*in.Why) == "" {
			return nil, refuse("missing-value", 2, "say why, in one sentence: --why <text>", "%s needs --why", sub)
		}
	}
	var full []string
	for _, c := range in.Commits {
		if to != "fixed" {
			break // only a fix names commits
		}
		sha, e := commitOf(p.Main, c)
		if e != nil {
			return nil, e
		}
		full = append(full, sha)
	}
	release, e := lock(p)
	if e != nil {
		return nil, e
	}
	defer release()
	entries, _, err := Read(p)
	if err != nil {
		return nil, readError(p, err)
	}
	en, e := Find(entries, key)
	if e != nil {
		return nil, e
	}
	b := *en.Bug
	from := b.Status
	if !moveAllowed(from, to) {
		return nil, refuse("bug-move", 4, "see where it stands: bonsai bug show "+b.ID+"; the moves: "+movesText(),
			"%s is %s, and a bug does not move from %s to %s", b.ID, from, from, to)
	}
	b.Status = to
	today := p.now().UTC().Format("2006-01-02")
	switch to {
	case "fixed":
		by := in.By
		b.FixedBy = &by
		fixed := strings.Join(full, ", ")
		if h := oneLine(in.How); h != "" {
			fixed += " (" + h + ")"
		}
		b.FixedIn = &fixed
		b.Closed = &today
		if in.Why != nil {
			b.Why = opt(in.Why)
		}
	case "wontfix":
		b.Closed = &today
		b.Why = opt(in.Why)
	case "watched":
		b.Closed = nil
		b.Why = opt(in.Why)
	case "open":
		b.Closed, b.FixedBy, b.FixedIn = nil, nil, nil
		b.Why = opt(in.Why)
	}
	if in.Task != nil {
		b.Task = opt(in.Task)
		if e := taskValue(b.Task); e != nil {
			return nil, e
		}
	}
	w, e := write(p, en.Path, &b, en.Body, false)
	if e != nil {
		return nil, e
	}
	w.Reason = from + " to " + to
	w.Logged = logRecord(p, sub, &b, w.Reason) == nil
	return w, nil
}

// Edit changes a bug's text fields (never its status: that is a move's), under the lock, then the log record.
func Edit(p Place, key string, in Fields) (*Written, *Error) {
	if in.Title != nil && strings.TrimSpace(*in.Title) == "" {
		return nil, refuse("bad-value", 2, "give --title the new title, on one line", "a bug's title cannot be empty")
	}
	release, e := lock(p)
	if e != nil {
		return nil, e
	}
	defer release()
	entries, _, err := Read(p)
	if err != nil {
		return nil, readError(p, err)
	}
	en, e := Find(entries, key)
	if e != nil {
		return nil, e
	}
	b := *en.Bug
	var changed []string
	set := func(name string, dst **string, v *string) {
		if v == nil {
			return
		}
		nv := opt(v)
		if show(*dst) != show(nv) {
			changed = append(changed, name)
			*dst = nv
		}
	}
	if in.Title != nil && oneLine(*in.Title) != b.Title {
		b.Title = oneLine(*in.Title)
		changed = append(changed, "title")
	}
	set("area", &b.Area, in.Area)
	set("where", &b.Where, in.Where)
	set("found_by", &b.FoundBy, in.FoundBy)
	set("task", &b.Task, in.Task)
	set("why", &b.Why, in.Why)
	if e := taskValue(b.Task); e != nil {
		return nil, e
	}
	body := en.Body
	if in.Body != nil && strings.ReplaceAll(*in.Body, "\r\n", "\n") != body {
		body = *in.Body
		changed = append(changed, "body")
	}
	if len(changed) == 0 {
		return nil, refuse("bad-flag", 2, "give edit the fields to change: --title, --area, --where, --found-by, --task, --why, --body",
			"edit changes nothing: every value given is the bug's already")
	}
	w, e := write(p, en.Path, &b, body, false)
	if e != nil {
		return nil, e
	}
	w.Reason = "edited " + strings.Join(changed, ", ")
	w.Logged = logRecord(p, "edit", &b, w.Reason) == nil
	return w, nil
}

// write writes a bug file whole (through a staged file and a rename, retried while Windows reports it busy); fresh
// refuses a file already there.
func write(p Place, rel string, b *format.Bug, body string, fresh bool) (*Written, *Error) {
	raw, err := b.EncodeFile(body)
	if err != nil {
		return nil, refuse("bad-value", 2, "give the fields in the form bonsai check --schema bonsai.bug lists",
			"the bug would not pass its format: %s", oneLine(err.Error()))
	}
	abs := filepath.Join(p.Main, filepath.FromSlash(rel))
	if fresh {
		if _, err := os.Lstat(abs); err == nil {
			return nil, refuse("bug-id-taken", 4, "run the command again: the next free id is read again under the lock",
				"%s is there already", rel)
		}
	}
	if err := workspace.WriteFileAtomic(abs, raw); err != nil {
		return nil, refuse("write-failed", 3, "check that the bug folder can be written, then run the command again",
			"%s cannot be written: %s", rel, oneLine(err.Error()))
	}
	back, readBody, err := format.ReadBugFile(raw)
	if err != nil {
		back, readBody = b, body
	}
	return &Written{Entry: Entry{Path: rel, Bug: back, Body: readBody}}, nil
}

// logRecord appends the bug record (contract section 8.2): target the id, kind the subcommand, reason the change, text
// the title, task and role from BONSAI_TASK and BONSAI_ROLE, in the session's file or the day's.
func logRecord(p Place, sub string, b *format.Bug, reason string) error {
	at := p.now()
	session, agent := p.Session, ""
	if session != "" {
		if _, err := record.LogFileName(session, at); err != nil {
			session = ""
		} else {
			agent = "claude-code"
		}
	}
	l := record.New("bug", record.Common{Workspace: p.Workspace, Local: p.Local, Session: session, Agent: agent,
		Getenv: p.Getenv, Redact: redact.Text, At: at})
	id, kind, why, title := b.ID, sub, redact.Text(reason), redact.Text(b.Title)
	l.Target, l.Kind, l.Reason, l.Text = &id, &kind, &why, &title
	_, err := record.WriteLog(p.Main, l)
	return err
}

// ---------------------------------------------------------------- helpers

// commitOf gives the full name of a commit the main checkout's repository holds, or refuses one it lacks. Only a hex
// name of 4 to 64 digits is taken, never a ref or an option.
func commitOf(main, c string) (string, *Error) {
	c = strings.ToLower(strings.TrimSpace(c))
	if !regexp.MustCompile(`^[0-9a-f]{4,64}$`).MatchString(c) {
		return "", refuse("bad-value", 2, "give --commit the commit's hex name (git log --oneline shows it)",
			"--commit %s is not a commit's hex name", strconv.QuoteToASCII(c))
	}
	out, ok := git(main, "rev-parse", "--verify", "--quiet", "--end-of-options", c+"^{commit}")
	if !ok {
		return "", refuse("commit-not-found", 2, "name a commit this repository holds (git log --oneline shows them; fetch first if it is on another clone)",
			"the repository holds no commit %s", c)
	}
	// The name as given when it is long enough to stay unique, else the commit's first eight digits.
	if full := strings.TrimSpace(out); len(c) < 8 && len(full) >= 8 {
		return full[:8], nil
	}
	return c, nil
}

// git runs git in dir with git's variables that would point it at another repository taken out; ok is false when it
// fails.
func git(dir string, args ...string) (string, bool) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var env []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if runtime.GOOS == "windows" {
			k = strings.ToUpper(k)
		}
		switch k {
		case "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES",
			"GIT_COMMON_DIR", "GIT_NAMESPACE", "GIT_TERMINAL_PROMPT", "GIT_OPTIONAL_LOCKS":
			continue
		}
		env = append(env, kv)
	}
	cmd.Env = append(env, "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", false
	}
	return out.String(), true
}

// statuses is the bug kind's closed list (bonsai.bug/1's enum, its one home, through workspace).
func statuses() []string {
	k, _ := workspace.BonsaiDocKind("bug")
	return k.Statuses
}

// moveAllowed reports a move the bug kind allows (workspace's bugMoves, its one home).
func moveAllowed(from, to string) bool {
	k, _ := workspace.BonsaiDocKind("bug")
	for _, m := range k.Agent {
		if len(m) == 2 && m[0] == from && m[1] == to {
			return true
		}
	}
	return false
}

func movesText() string {
	k, _ := workspace.BonsaiDocKind("bug")
	var out []string
	for _, m := range k.Agent {
		out = append(out, m[0]+" to "+m[1])
	}
	return strings.Join(out, ", ")
}

func taskValue(t *string) *Error {
	if t != nil && !regexp.MustCompile(workspace.TaskIDPattern).MatchString(*t) {
		return refuse("bad-value", 2, "give --task a task id, T- and 4 to 6 digits (or \"\" for none)", "--task %s is not a task id", strconv.QuoteToASCII(*t))
	}
	return nil
}

func validDay(d string) bool {
	if !dayPattern.MatchString(d) {
		return false
	}
	_, err := time.Parse("2006-01-02", d)
	return err == nil
}

func readError(p Place, err error) *Error {
	return refuse("read-failed", 3, "check that the bug folder "+p.Folder+" can be read, then run the command again",
		"the bug folder %s cannot be read: %s", p.Folder, oneLine(err.Error()))
}

// opt gives a nullable field's value, on one line: nil and "" (after trimming) are null.
func opt(v *string) *string {
	if v == nil {
		return nil
	}
	s := oneLine(*v)
	if s == "" {
		return nil
	}
	return &s
}

func show(v *string) string {
	if v == nil {
		return "\x00"
	}
	return *v
}

func oneLine(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(strings.ReplaceAll(strings.ReplaceAll(s, "\r", " "), "\n", " ")), " "))
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
