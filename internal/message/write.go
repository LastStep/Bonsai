package message

// Writing one message (contract §22; design/plan-5.md, "Messages between agents", note 5.5.10, 2). Everything is
// checked before anything is written, in this order, each refusal naming its next step:
//
//  1. --from, --to and --kind are given (missing-value) and of their form (bad-value: an address, CheckAddress; a
//     kind, brief or report).
//  2. A report has a --status (report-no-status), done, blocked or needs-input (bad-value); a brief has none, and a
//     report no --budget (bad-flag).
//  3. --budget is minutes=N,actions=N,red_runs=N, any of the three, each a whole number (bad-value); --task a task
//     id, --thread and --reply-to message ids (bad-value).
//  4. The body: given (missing-value), UTF-8 with no hidden or control character but a line feed and a tab (refused,
//     not stripped, as an ask's text is: asks.Hidden), every secret taken out by the redactor, then at most MaxBody
//     bytes (bad-value, never cut). A byte-order mark at its start (a Windows editor's) is dropped, CRLF is made LF,
//     and a body that does not end in a line feed gets one.
//  5. The rules (Rules): --from's and --to's roles among the declared roles, where any are declared
//     (role-not-declared); each limit a brief's budget gives at most its max in force for --to's role
//     (budget-over-max).
//  6. Each --pointer names a file that is there, inside the checkout or under the home (pointer-missing), and is
//     fingerprinted then (pointer.go).
//  7. The thread: --thread names a thread's first brief, --reply-to a message (message-not-found; a --reply-to of
//     another thread than --thread is bad-value). A brief with neither starts a thread of its own; a report needs one
//     (missing-value). A report goes only into a thread holding a brief to its --from, and only to that brief's
//     from; or, from a brief's own writer closing its dropped thread, blocked, to that brief's to
//     (report-not-in-thread). Its reply_to is that brief (the one --reply-to names, else the newest that fits); a
//     follow-up brief's is --reply-to, else its thread's newest report, else its newest message. A message in a
//     thread with no --task serves the thread's task.
//
// Then the file: .bonsai/.gitignore restored when missing (every writer of .bonsai/local/ does), the recipient's
// folder made on first use, the message's id made from the clock and 4 random bytes, its bytes (format's EncodeFile)
// staged under .<id>.md.tmp-<random> by an exclusive create in that folder, and moved to <id>.md only where no file
// of that name is there (move.go), retried while Windows reports the file busy. Two writers at once never collide:
// each stages a name of its own, and a move never replaces a file (a taken name gets a new id, at most three times).
// Then one message log record (contract §8.2): target the id, text its kind, recipient and a report's status, task
// the message's task, in the writing session's log file or its day's. A log record that cannot be written after the
// message is partly-written: the message stands.
//
// Who may write a brief to whom, and whether --from is the writer's own address, are the guard's rules (step
// 5.5.11), on the shell line that runs bonsai message; nothing here knows which agent is calling.

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// The statuses a report takes (bonsai.message/1's closed list).
const (
	StatusDone       = "done"
	StatusBlocked    = "blocked"
	StatusNeedsInput = "needs-input"
)

var statuses = []string{StatusDone, StatusBlocked, StatusNeedsInput}

// Writing is one message as its command line gives it; "" for a flag not given.
type Writing struct {
	From, To, Kind                string
	Task, Thread, ReplyTo, Status string
	Budget                        *string  // --budget as given, nil when not given
	Pointers                      []string // each --pointer as given
	Body                          string   // the words, as given (--body, --body-file's file, or standard input)
	Cwd                           string   // the folder the command runs in, where a relative --pointer is read from
}

// Written is a message Write wrote.
type Written struct {
	Message *format.Message
	Path    string // its file, project-relative in the main checkout, forward slashes
	Thread  Thread // its thread as it stands after the write
}

// The writes Write makes, variables for the tests.
var (
	ensureGitignore = workspace.EnsureGitignore
	random          = rand.Read
	writeLog        = record.WriteLog
)

// Write checks one message and writes it (this file's comment). The store it reads the threads from is read here.
func Write(p Place, w Writing, r Rules) (*Written, *Error) {
	msg, e := w.check(r)
	if e != nil {
		return nil, e
	}
	for _, given := range w.Pointers {
		ptr, e := Pointer(p, given, w.Cwd)
		if e != nil {
			return nil, e
		}
		if !hasPointer(msg.Pointers, ptr.Path) {
			msg.Pointers = append(msg.Pointers, ptr)
		}
	}
	st, err := Read(p.Local.Main)
	if err != nil {
		return nil, &Error{Code: "read-failed", Exit: ExitRuntime, What: Dir + " cannot be read: " + oneLine(err),
			Next: "check that " + Dir + "/ can be read, then run the command again"}
	}
	if e := place(st, msg, w); e != nil {
		return nil, e
	}
	path, e := file(p, st, msg)
	if e != nil {
		return nil, e
	}
	entry := Entry{Msg: msg, Path: path}
	st.Messages = append(st.Messages, entry)
	sortEntries(st.Messages)
	written := &Written{Message: msg, Path: path}
	for _, t := range Fold(st, r, p.now()) {
		if t.ID == msg.Thread {
			written.Thread = t
		}
	}
	if err := writeRecord(p, msg); err != nil {
		return written, &Error{Code: "partly-written", Exit: ExitRuntime,
			What: "the message " + msg.ID + " was written, but its log record was not: " + oneLine(err),
			Next: "check that .bonsai/local/log/ can be written; the message stands (bonsai message --show " + msg.ID + ")"}
	}
	return written, nil
}

// check holds the command line to steps 1 to 5 and gives the message as far as it can be made without the disk.
func (w Writing) check(r Rules) (*format.Message, *Error) {
	for _, f := range []struct{ name, value string }{{"--from", w.From}, {"--to", w.To}, {"--kind", w.Kind}} {
		if f.value == "" {
			return nil, &Error{Code: "missing-value", Exit: ExitInput, What: "bonsai message needs " + f.name,
				Next: "give --from <address> --to <address> --kind brief|report (bonsai message --help)"}
		}
	}
	for _, f := range []struct{ name, value string }{{"--from", w.From}, {"--to", w.To}} {
		if why := CheckAddress(f.value); why != "" {
			return nil, badValue(f.name+" "+strconv.QuoteToASCII(f.value)+" "+why,
				"give "+f.name+" as <role>, <role>@<task> or <role>@<task>#<n>, such as builder@T-0901")
		}
	}
	msg := &format.Message{From: w.From, To: w.To, Kind: w.Kind, Pointers: []format.MessagePointer{}, Labels: schema.Object{}}
	switch w.Kind {
	case format.KindBrief:
		if w.Status != "" {
			return nil, badFlag("a brief takes no --status: only a report says done, blocked or needs-input", "leave --status out of a brief")
		}
		msg.Budget = &format.MessageBudget{}
	case format.KindReport:
		if w.Status == "" {
			return nil, &Error{Code: "report-no-status", Exit: ExitInput, What: "a report needs --status: done, blocked or needs-input",
				Next: "give --status done, --status blocked or --status needs-input"}
		}
		if w.Budget != nil {
			return nil, badFlag("a report takes no --budget: only a brief gives one", "leave --budget out of a report")
		}
		if !contains(statuses, w.Status) {
			return nil, badValue("--status "+strconv.QuoteToASCII(w.Status)+" is not done, blocked or needs-input",
				"give --status done, --status blocked or --status needs-input")
		}
		msg.Status = ptr(w.Status)
	default:
		return nil, badValue("--kind "+strconv.QuoteToASCII(w.Kind)+" is not brief or report", "give --kind brief or --kind report")
	}
	if w.Budget != nil {
		if e := parseBudget(*w.Budget, msg.Budget); e != nil {
			return nil, e
		}
	}
	if w.Task != "" && !IsTask(w.Task) {
		return nil, badValue("--task "+strconv.QuoteToASCII(w.Task)+" is not a task id (T- and 4 to 6 digits)", "give --task as T-0901, or leave it out")
	}
	msg.Task = ptr(w.Task)
	for _, f := range []struct{ name, value string }{{"--thread", w.Thread}, {"--reply-to", w.ReplyTo}} {
		if f.value != "" && !IsID(f.value) {
			return nil, badValue(f.name+" "+strconv.QuoteToASCII(f.value)+" is not a message id (M-, the UTC date, a dash and 8 hex)",
				"give "+f.name+" as an id bonsai messages lists, such as M-2026-10-16-3f9a01bc")
		}
	}
	body, e := cleanBody(w.Body)
	if e != nil {
		return nil, e
	}
	msg.Body = body
	for _, f := range []struct{ name, value string }{{"--from", w.From}, {"--to", w.To}} {
		if role := RoleOf(f.value); !r.declares(role) {
			return nil, &Error{Code: "role-not-declared", Exit: ExitInput,
				What: fmt.Sprintf("%s names the role %s, which the linked packs do not declare (they declare %s)", f.name, role, strings.Join(r.Roles, ", ")),
				Next: "address the message to a declared role, as bonsai status --json lists them"}
		}
	}
	if msg.Budget != nil && r.Max != nil {
		role := RoleOf(w.To)
		for _, l := range []struct {
			name  string
			value *int64
		}{{"minutes", msg.Budget.Minutes}, {"actions", msg.Budget.Actions}, {"red_runs", msg.Budget.RedRuns}} {
			if max := r.Max(role, l.name); l.value != nil && max != nil && *l.value > *max {
				return nil, &Error{Code: "budget-over-max", Exit: ExitInput,
					What: fmt.Sprintf("--budget gives %s %d, over its max in force for the role %s, %d", l.name, *l.value, role, *max),
					Next: fmt.Sprintf("give %s=%d or less, or leave it out for the default in force; a person may raise the max in bonsai.yaml", l.name, *max)}
			}
		}
	}
	return msg, nil
}

// parseBudget reads --budget into b: minutes=N,actions=N,red_runs=N, any of the three once each, each a whole number.
func parseBudget(s string, b *format.MessageBudget) *Error {
	bad := func(why string) *Error {
		return badValue("--budget "+strconv.QuoteToASCII(s)+" "+why, "give --budget as minutes=N,actions=N,red_runs=N, any of the three, each a whole number")
	}
	if strings.TrimSpace(s) == "" {
		return bad("is empty")
	}
	seen := map[string]bool{}
	for _, part := range strings.Split(s, ",") {
		name, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		var slot **int64
		switch name {
		case "minutes":
			slot = &b.Minutes
		case "actions":
			slot = &b.Actions
		case "red_runs":
			slot = &b.RedRuns
		default:
			return bad("names " + strconv.QuoteToASCII(name) + ", which is not minutes, actions or red_runs")
		}
		n, err := strconv.ParseInt(value, 10, 64)
		switch {
		case !ok || err != nil || n < 0 || strings.TrimLeft(value, "0123456789") != "":
			return bad("gives " + name + " " + strconv.QuoteToASCII(value) + ", which is not a whole number")
		case seen[name]:
			return bad("gives " + name + " twice")
		}
		seen[name] = true
		*slot = &n
	}
	return nil
}

// ReadBody reads a body from a reader (--body-file's file, or standard input with --body -), at most what cleanBody
// could take: more is refused, never cut.
func ReadBody(r io.Reader, from string) (string, *Error) {
	raw, err := io.ReadAll(io.LimitReader(r, 2*MaxBody+1))
	if err != nil {
		return "", &Error{Code: "read-failed", Exit: ExitRuntime, What: from + " cannot be read: " + oneLine(err),
			Next: "check that it can be read, then run the command again"}
	}
	if len(raw) > 2*MaxBody {
		return "", tooLong(from)
	}
	return string(raw), nil
}

func tooLong(what string) *Error {
	return badValue(fmt.Sprintf("%s is over %d bytes, more than a message's body may hold (nothing is cut)", what, MaxBody),
		"put the long output in a file and give it with --pointer, with a few lines in the body")
}

// cleanBody gives the body as written (this file's comment, step 4).
func cleanBody(s string) (string, *Error) {
	s = strings.ReplaceAll(strings.TrimPrefix(s, "\uFEFF"), "\r\n", "\n")
	if strings.TrimSpace(s) == "" {
		return "", &Error{Code: "missing-value", Exit: ExitInput, What: "a message needs words: its body is empty",
			Next: "give --body <text>, --body-file <path> or --body - (standard input)"}
	}
	if !utf8.ValidString(s) {
		return "", badValue("the body is not UTF-8 text", "give the body as UTF-8 text")
	}
	for _, r := range s {
		if r != '\n' && r != '\t' && asks.Hidden(r) {
			return "", badValue(fmt.Sprintf("the body holds a hidden or control character (U+%04X), which is refused, not stripped", r),
				"give the body again without it")
		}
	}
	s = redact.Text(s)
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	if len(s) > MaxBody {
		return "", tooLong("the body")
	}
	return s, nil
}

// place puts the message in its thread (this file's comment, step 7): thread, reply_to and task.
func place(st *Store, msg *format.Message, w Writing) *Error {
	notFound := func(flag, id string) *Error {
		e := &Error{Code: "message-not-found", Exit: ExitState, What: flag + " " + id + " names no message under " + Dir + "/",
			Next: "run bonsai messages to list the threads and their messages"}
		if _, bad := st.Find(id); bad != nil {
			e.What = flag + " " + id + " names " + bad.Path + ", which does not read as a message: " + bad.Reason
		}
		return e
	}
	thread := w.Thread
	var replyTo *Entry
	if thread != "" {
		head, _ := st.Find(thread)
		switch {
		case head == nil:
			return notFound("--thread", thread)
		case head.Msg.Thread != thread:
			return badValue("--thread "+thread+" is a message of the thread "+head.Msg.Thread+", not a thread's first brief",
				"give --thread "+head.Msg.Thread)
		}
	}
	if w.ReplyTo != "" {
		if replyTo, _ = st.Find(w.ReplyTo); replyTo == nil {
			return notFound("--reply-to", w.ReplyTo)
		}
		if thread != "" && replyTo.Msg.Thread != thread {
			return badValue("--reply-to "+w.ReplyTo+" is a message of the thread "+replyTo.Msg.Thread+", not of --thread "+thread,
				"give --thread "+replyTo.Msg.Thread+", or leave --thread out")
		}
		thread = replyTo.Msg.Thread
	}
	if thread == "" {
		if msg.Kind == format.KindReport {
			return &Error{Code: "missing-value", Exit: ExitInput, What: "a report needs --thread or --reply-to: the thread of the brief it answers",
				Next: "give --thread <id> (bonsai messages --to " + w.From + " lists the threads)"}
		}
		return nil // a thread of its own: its id, and so its thread, is made when it is written
	}
	msg.Thread = thread
	var in []Entry
	for _, e := range st.Messages {
		if e.Msg.Thread == thread {
			in = append(in, e)
		}
	}
	if msg.Task == nil {
		for _, e := range in {
			if e.Msg.ID == thread {
				msg.Task = e.Msg.Task
			}
		}
	}
	if msg.Kind == format.KindBrief {
		switch {
		case replyTo != nil:
			msg.ReplyTo = &replyTo.Msg.ID
		default:
			last := in[len(in)-1].Msg
			for _, e := range in {
				if e.Msg.Kind == format.KindReport {
					last = e.Msg
				}
			}
			msg.ReplyTo = &last.ID
		}
		return nil
	}
	fits := func(b *format.Message) bool {
		if b.Kind != format.KindBrief {
			return false
		}
		answers := b.To == msg.From && msg.To == b.From
		closes := b.From == msg.From && msg.To == b.To && *msg.Status == StatusBlocked
		return answers || closes
	}
	var brief *format.Message
	if replyTo != nil {
		if fits(replyTo.Msg) {
			brief = replyTo.Msg
		}
	} else {
		for i := len(in) - 1; i >= 0 && brief == nil; i-- {
			if fits(in[i].Msg) {
				brief = in[i].Msg
			}
		}
	}
	if brief == nil {
		return notInThread(in, msg, thread, replyTo)
	}
	msg.ReplyTo = &brief.ID
	return nil
}

// notInThread is a report refused for its thread, saying what the thread holds.
func notInThread(in []Entry, msg *format.Message, thread string, replyTo *Entry) *Error {
	e := &Error{Code: "report-not-in-thread", Exit: ExitState,
		Next: "answer a brief to " + msg.From + " back to its from (bonsai messages --to " + msg.From + "), or close a brief " +
			msg.From + " wrote with --status blocked to its to"}
	var to, from []string
	for _, x := range in {
		if x.Msg.Kind != format.KindBrief {
			continue
		}
		if x.Msg.To == msg.From && !contains(to, x.Msg.From) {
			to = append(to, x.Msg.From)
		}
		if x.Msg.From == msg.From && !contains(from, x.Msg.To) {
			from = append(from, x.Msg.To)
		}
	}
	switch {
	case replyTo != nil && replyTo.Msg.Kind != format.KindBrief:
		e.What = "--reply-to " + replyTo.Msg.ID + " is a report: a report answers a brief"
	case replyTo != nil:
		e.What = "--reply-to " + replyTo.Msg.ID + " is a brief from " + replyTo.Msg.From + " to " + replyTo.Msg.To +
			": a report to it goes from its to back to its from (or, blocked, from its from to its to)"
	case len(to) == 0 && len(from) == 0:
		e.What = "the thread " + thread + " holds no brief to " + msg.From + ", so " + msg.From + " has no report to write in it"
	case len(to) > 0 && (len(from) == 0 || *msg.Status != StatusBlocked):
		e.What = "a report from " + msg.From + " in the thread " + thread + " goes back to the brief's from (" + strings.Join(to, ", ") + "), not to " + msg.To
	default:
		e.What = msg.From + " wrote the thread " + thread + "'s brief to " + strings.Join(from, ", ") +
			": its writer closes it only with --status blocked, to that brief's to"
	}
	return e
}

// file writes the message's file (this file's comment): its id made, staged, and moved to a name no file has.
func file(p Place, st *Store, msg *format.Message) (string, *Error) {
	main := p.Local.Main
	if _, err := ensureGitignore(main); err != nil {
		return "", &Error{Code: "write-failed", Exit: ExitRuntime,
			What: workspace.GitignoreFile + " is missing and cannot be written, so no message is written: " + oneLine(err),
			Next: "check that .bonsai/ can be written, then run the command again"}
	}
	folder := FolderOf(msg.To)
	dir := filepath.Join(main, filepath.FromSlash(Dir), folder)
	failed := func(err error) *Error {
		return &Error{Code: "write-failed", Exit: ExitRuntime, What: Dir + "/" + folder + " cannot be written: " + oneLine(err),
			Next: "check that " + Dir + "/ can be written, then run the command again"}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", failed(err)
	}
	own := msg.Thread == ""
	for try := 0; try < 3; try++ {
		at := p.now().UTC()
		var b [4]byte
		if _, err := random(b[:]); err != nil {
			return "", unexpected(fmt.Errorf("no random number from the system: %v", err))
		}
		msg.ID = "M-" + at.Format("2006-01-02") + "-" + hex.EncodeToString(b[:])
		msg.Created = at.Format(record.AtLayout)
		if own {
			msg.Thread = msg.ID
		}
		if e, bad := st.Find(msg.ID); e != nil || bad != nil {
			continue // an id another file has, in any folder: a new one
		}
		raw, err := msg.EncodeFile()
		if err != nil {
			return "", unexpected(err)
		}
		if len(raw) > MaxFile {
			return "", badValue(fmt.Sprintf("the message is %d bytes once written, over the %d a message file may hold", len(raw), MaxFile),
				"give fewer pointers")
		}
		err = stage(dir, msg.ID, raw)
		switch {
		case err == nil:
			return Dir + "/" + folder + "/" + msg.ID + ".md", nil
		case !errors.Is(err, errTaken):
			return "", failed(err)
		}
	}
	return "", &Error{Code: "write-failed", Exit: ExitRuntime, What: "three ids in a row were another message's, so no message is written",
		Next: "run the command again"}
}

// stage writes raw to a new file of its own in dir, by an exclusive create, and moves it to <id>.md (move.go); the
// staged file is removed on any failure.
func stage(dir, id string, raw []byte) error {
	f, err := os.CreateTemp(dir, "."+id+".md.tmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = f.Write(raw)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = moveRetry(tmp, filepath.Join(dir, id+".md"))
	}
	if err != nil {
		_ = workspace.RemoveFile(tmp)
	}
	return err
}

// writeRecord writes the message's log record (contract §8.2's message): target its id; text its kind, its
// recipient and a report's status; task its task; in the writing session's file, or its day's.
func writeRecord(p Place, msg *format.Message) error {
	at, err := time.Parse(record.AtLayout, msg.Created)
	if err != nil {
		return err
	}
	session, agent := p.Session, ""
	if session != "" {
		if _, err := record.LogFileName(session, at); err != nil {
			session = "" // not an id Claude Code gives: the day's file
		} else {
			agent = asks.Agent
		}
	}
	l := record.New("message", record.Common{Workspace: p.Workspace, Local: p.Local, Session: session, Agent: agent,
		Getenv: p.Getenv, Redact: redact.Text, At: at})
	id, text := msg.ID, msg.Kind+" to "+msg.To
	if msg.Status != nil {
		text += ", " + *msg.Status
	}
	l.Target, l.Text, l.Task = &id, &text, nil
	if msg.Task != nil {
		task := *msg.Task
		l.Task = &task
	}
	_, err = writeLog(p.Local.Main, l)
	return err
}

func hasPointer(ps []format.MessagePointer, path string) bool {
	for _, p := range ps {
		if p.Path == path {
			return true
		}
	}
	return false
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
