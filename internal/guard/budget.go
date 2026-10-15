package guard

// The stall stops (design/plan-5.md, step 5.5, "Role rules and stall stops, written once": "How it is counted",
// "Past a budget"; 5.5.11 notes 5 and 6): each agent's budget, counted by the guard in its state (internal/agents).
//
// Counted: every call the guard sees from an agent or a named session (refused ones included), one call line each,
// appended, so calls made side by side lose no count; minutes since its first call or since it tied the brief now in
// force; red runs, read from each climb's own result at a later call, never from the shell's exit code. The person's
// own session is never counted.
//
// A brief ties (tie) only if it is sound: a message file holds it, read through internal/message (readBrief), it is a
// brief, it is to the agent's address (its role at a first tie, <role>@<BONSAI_TASK> for a named session; at a later
// tie the address it has, in the tied brief's thread, written after it), its from names a role that may call the
// agent's, bonsai message wrote it (its message log record is there, found by its id: messageRecorded, read only at a
// tie), and no other agent tied it (an exclusive tie file). Anything else ties nothing, with a guard record saying why
// (brief-not-tied). A brief ties at the agent's first call from its starting call's prompt (link, roles.go), when a
// SendMessage resuming it names one (an offer, tried at its next call), and when it reads one (bonsai message --show
// <id>, or a Read of the brief's file). Its budget is then in force, each limit the brief's value up to the max in
// force, else the role's default, counted afresh from the tie.
//
// Past a limit (more minutes, actions or red runs than its value): every call is refused (budget-spent) but three
// (budget-exempt): its report in its exact form (rolecmds.go's reportCall), reading its tied brief or a newer one to
// its address, each by the installed bonsai (a bare bonsai or the hook lines' place, never another program of that
// name), and Claude Code's hand-back tool. The first refusal says what to report and how; later ones are short; each
// names a newer brief to its address already waiting, if one is (waitingBrief). Of calls side by side past it, one
// alone writes the window's spent line (agents.MarkSpent) and gives the first refusal. The report's call let through
// is noted (reported) and never trusted: the stop gate finds the report itself (agents.ReportAfter).
// An agent with no address (no tied brief, or an undeclared type) has no report call: its last reply is its report.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// brief is a message as a tie reads it.
type brief struct {
	id, thread, from, to, kind string
	created                    time.Time
	budget                     *format.MessageBudget
}

// readBrief reads a message by its id from the main checkout's messages, through internal/message's own reading of
// one file (message.Parse: bonsai.message/1, held to its recipient folder and its file name): the file
// .bonsai/local/messages/<folder>/<id>.md, a regular file in a plain folder, at most message.MaxFile bytes, in one
// folder only; why it cannot be read otherwise ("" when it can). A variable for the tests.
var readBrief = func(main, id string) (*brief, string) {
	if !message.IsID(id) {
		return nil, "it is not a message's id"
	}
	root := filepath.Join(main, filepath.FromSlash(message.Dir))
	folders, err := os.ReadDir(root)
	if err != nil {
		return nil, "no message file holds it"
	}
	var found []string
	for _, f := range folders {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		if _, err := os.Lstat(filepath.Join(root, f.Name(), id+".md")); err == nil {
			found = append(found, f.Name())
		}
	}
	switch len(found) {
	case 0:
		return nil, "no message file holds it"
	case 1:
	default:
		return nil, "two message files hold it"
	}
	folder := found[0]
	if fi, err := os.Lstat(filepath.Join(root, folder)); err != nil || fi.Mode().Type() != fs.ModeDir {
		return nil, "its recipient's folder is not a plain folder"
	}
	path := filepath.Join(root, folder, id+".md")
	fi, err := os.Lstat(path)
	switch {
	case err != nil:
		return nil, "its message file cannot be read"
	case !fi.Mode().IsRegular():
		return nil, "its message file is not a regular file"
	case fi.Size() > message.MaxFile:
		return nil, "its message file is larger than any bonsai message writes"
	}
	raw, err := agents.ReadRegular(path, message.MaxFile)
	if err != nil {
		return nil, "its message file cannot be read"
	}
	m, _, why := message.Parse(raw, folder, id+".md")
	if why != "" {
		return nil, "its message file does not read (" + why + ")"
	}
	created, err := time.Parse(time.RFC3339Nano, m.Created)
	if err != nil {
		return nil, "its created time does not read"
	}
	return &brief{id: m.ID, thread: m.Thread, from: m.From, to: m.To, kind: m.Kind, created: created, budget: m.Budget}, ""
}

// messageRecorded finds the message record bonsai message wrote with a message, read only at a tie
// (agents.MessageRecorded, which the stop gate reads a report by too). A variable for the tests.
var messageRecorded = func(main, session string, b *brief) bool {
	return agents.MessageRecorded(main, session, b.id, b.created)
}

// waitingBrief gives the newest brief to the agent's address in the thread of the brief it has tied, written after it
// and tied by no agent: the newer brief waiting for it, named in the refusal past its budget; "" for none. It reads
// its address's own recipient folder only, each file through message.Parse.
func (rr *roleRun) waitingBrief(w *who) string {
	if w.st == nil {
		return ""
	}
	t := w.st.Window().Tie
	if t == nil || t.To == "" || message.CheckAddress(t.To) != "" {
		return ""
	}
	tied, why := readBrief(rr.main, t.Brief)
	if why != "" {
		return ""
	}
	folder := message.FolderOf(t.To)
	dir := filepath.Join(rr.main, filepath.FromSlash(message.Dir), folder)
	if fi, err := os.Lstat(dir); err != nil || fi.Mode().Type() != fs.ModeDir {
		return ""
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	best, bestAt := "", tied.created
	for _, e := range entries {
		id, ok := strings.CutSuffix(e.Name(), ".md")
		if !ok || !message.IsID(id) || id == t.Brief || !e.Type().IsRegular() {
			continue
		}
		fi, err := e.Info()
		if err != nil || fi.Size() > message.MaxFile {
			continue
		}
		raw, err := agents.ReadRegular(filepath.Join(dir, e.Name()), message.MaxFile)
		if err != nil {
			continue
		}
		m, _, why := message.Parse(raw, folder, e.Name())
		if why != "" || m.Kind != format.KindBrief || m.To != t.To || m.Thread != tied.thread {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, m.Created)
		if err != nil || !at.After(bestAt) || rr.store.TiedBy(id) != "" {
			continue
		}
		best, bestAt = id, at
	}
	return best
}

// roleOf gives an address's role: the part before its @.
func roleOf(addr string) string {
	r, _, _ := strings.Cut(addr, "@")
	return r
}

// address gives the caller's own address (step 5.5's "an agent's own address"): its tied brief's to; a named
// session's <role>@<BONSAI_TASK>; the person's session's, its role's name; "" for none (no tied brief, or an
// undeclared type).
func (rr *roleRun) address(w *who) string {
	switch {
	case w.h.SessRole:
		if rs, err := rr.roles(); err == nil {
			if r := rs.sessionRole(); r != nil {
				return r.Role.Name
			}
		}
		return ""
	case w.h.Role == "":
		return ""
	}
	if w.st != nil {
		return w.st.Address()
	}
	if w.h.Named {
		if w.h.Task == "" {
			return w.h.Role
		}
		return w.h.Role + "@" + w.h.Task
	}
	return ""
}

// tie ties a brief to the agent if it is sound (this file's comment), saying how it was named; true when it is tied
// to it (now, or already). A brief that ties nothing is noted (brief-not-tied).
func (rr *roleRun) tie(w *who, id, how string) bool {
	in := rr.in
	not := func(why string) bool {
		rr.note(RuleBriefNotTied, fmt.Sprintf("the brief %s, named by %s, ties nothing: %s", id, how, why))
		return false
	}
	if w.h.Role == "" || w.h.SessRole {
		return not("an agent of a type no pack declares has no address, so no brief is to it")
	}
	b, why := readBrief(rr.main, id)
	switch {
	case why != "":
		return not(why)
	case b.kind != format.KindBrief:
		return not("it is a " + b.kind + ", not a brief")
	}
	var tied *agents.Event
	if w.st != nil {
		tied = w.st.Window().Tie
	}
	switch {
	case tied != nil && tied.Brief == id:
		return true
	case tied != nil:
		old, why := readBrief(rr.main, tied.Brief)
		switch {
		case why != "":
			return not("the brief tied now cannot be read to compare (" + why + ")")
		case b.to != tied.To:
			return not("it is to " + quote(b.to, 60) + ", not to this agent's address " + quote(tied.To, 60))
		case b.thread != old.thread:
			return not("it is in another thread than the brief tied now")
		case !b.created.After(old.created):
			return not("it was not written after the brief tied now")
		}
	case w.h.Named:
		if want := rr.address(w); b.to != want {
			return not("it is to " + quote(b.to, 60) + ", not to this named session's address " + quote(want, 60))
		}
	case roleOf(b.to) != w.h.Role:
		return not("it is to " + quote(b.to, 60) + ", not to this agent's role " + quote(w.h.Role, 40))
	}
	rs, err := rr.roles()
	if err != nil {
		return not("the roles in force cannot be read (" + oneLine(err.Error()) + ")")
	}
	from := rs.byType(roleOf(b.from))
	switch {
	case from == nil:
		return not("its from " + quote(b.from, 60) + " names no role the lock declares")
	case from.Role.Session:
	case from.Role.Kind == "lead" && w.h.Kind == "helper":
	default:
		return not("its from " + quote(b.from, 60) + " is a role that may not call this agent's")
	}
	if !messageRecorded(rr.main, in.Session, b) {
		return not("the log holds no message record of it, so bonsai message did not write it")
	}
	ok, err := rr.store.Tie(id, in.Session, w.key)
	switch {
	case err != nil:
		return not("its tie cannot be written (" + oneLine(err.Error()) + ")")
	case !ok:
		return not("another agent tied it")
	}
	limits := w.h.Limits
	if r := rs.byType(w.h.Pack + ":" + w.h.Role); r != nil {
		limits = limitsOf(r)
	}
	if b.budget != nil {
		limits.Minutes.Value = briefValue(b.budget.Minutes, limits.Minutes)
		limits.Actions.Value = briefValue(b.budget.Actions, limits.Actions)
		limits.RedRuns.Value = briefValue(b.budget.RedRuns, limits.RedRuns)
	}
	calls, _ := rr.store.Calls(in.Session, w.key)
	e := agents.Event{E: agents.EventTie, At: agents.StampMilli(rr.now), Brief: id, To: b.to, Thread: b.thread, From: b.from,
		Limits: &limits, Calls: calls}
	if err := rr.store.Append(in.Session, w.key, e); err != nil {
		return not("its tie line cannot be written (" + oneLine(err.Error()) + ")")
	}
	if st, err := rr.store.Read(in.Session, w.key); err == nil {
		w.st = st
	}
	return true
}

// briefValue is a brief's value for a limit, held to the max in force (bonsai message refuses one above it; a message
// written another way is held here); nil keeps the default in force.
func briefValue(v *int64, l agents.Limit) *int64 {
	switch {
	case v == nil:
		return l.Value
	case l.Max != nil && *v > *l.Max:
		return l.Max
	}
	return v
}

// counted is what the count of one call gave.
type counted struct {
	spent    *agents.Event // the window's spent line; nil while within the budget
	newSpent bool          // this call found it spent
}

// count ties the briefs offered or read, reads the open climbs' results, appends this call's line and judges the
// budget (this file's comment). reads is the brief this call reads, "" for none.
func (rr *roleRun) count(w *who, reads string) counted {
	in := rr.in
	if w.st != nil {
		for _, o := range w.st.Offers() {
			rr.tie(w, o.Brief, "a SendMessage resuming this agent")
			_ = rr.store.Append(in.Session, w.key, agents.Event{E: agents.EventTried, At: agents.StampMilli(rr.now), Brief: o.Brief})
		}
	}
	if reads != "" {
		rr.tie(w, reads, "this agent's read of it")
	}
	if st, err := rr.store.Read(in.Session, w.key); err == nil {
		w.st = st
	}
	rr.readClimbs(w)
	if w.st == nil {
		return counted{}
	}
	win := w.st.Window()
	minutes := minutesSince(win.Start, rr.now)
	total, err := rr.store.Call(in.Session, w.key, rr.now, minutes)
	if err != nil {
		rr.note(RuleStateUnreadable, "this call's line cannot be written ("+oneLine(err.Error())+"), so it is not counted")
		return counted{spent: w.st.Spent()}
	}
	if sp := w.st.Spent(); sp != nil {
		return counted{spent: sp}
	}
	actions := total - win.Calls
	red := w.st.RedRuns()
	var e *agents.Event
	for _, m := range []struct {
		name  string
		count int64
		limit *int64
	}{{"minutes", minutes, win.Limits.Minutes.Value}, {"actions", actions, win.Limits.Actions.Value},
		{"red_runs", red, win.Limits.RedRuns.Value}} {
		if m.limit != nil && m.count > *m.limit {
			e = &agents.Event{E: agents.EventSpent, At: agents.StampMilli(rr.now), Measure: m.name, Count: m.count, Of: *m.limit}
			break
		}
	}
	if e == nil {
		return counted{}
	}
	if made, err := rr.store.MarkSpent(in.Session, w.key, w.st.Ties()); err == nil && !made {
		return counted{spent: e} // another call side by side found this window spent first, and writes its one line
	}
	if err := rr.store.Append(in.Session, w.key, *e); err != nil {
		rr.note(RuleStateUnreadable, "this agent's spent line cannot be written ("+oneLine(err.Error())+")")
	}
	if st, err := rr.store.Read(in.Session, w.key); err == nil {
		w.st = st
	}
	return counted{spent: e, newSpent: true}
}

// maxResult caps a ladder result the guard reads for an agent's red runs, in bytes.
const maxResult = 64 << 20

// readClimbs reads, at a later call, the result of each climb the guard passed for this agent: a ladder result for
// its task written after the note (.bonsai/local/ladder/<task>.json, its started time not before the note's) appends
// a red line (the count in a row) or a reset; with none yet, the note stays open, and a later call looks again.
func (rr *roleRun) readClimbs(w *who) {
	if w.st == nil {
		return
	}
	red := w.st.RedRuns()
	for _, c := range w.st.OpenClimbs() {
		noted, err := agents.ParseStamp(c.At)
		if err != nil {
			continue
		}
		raw, err := agents.ReadRegular(filepath.Join(rr.main, filepath.FromSlash(workspace.LocalDir), "ladder", c.Task+".json"), maxResult)
		if err != nil {
			continue
		}
		l, err := format.ReadLadder(raw)
		if err != nil || l.Task == nil || *l.Task != c.Task {
			continue
		}
		started, err := time.Parse(time.RFC3339Nano, l.Started)
		if err != nil || started.Before(noted.Truncate(time.Millisecond)) {
			continue
		}
		e := agents.Event{E: agents.EventReset, At: agents.StampMilli(rr.now), Task: c.Task}
		if !l.Green {
			red++
			e = agents.Event{E: agents.EventRed, At: agents.StampMilli(rr.now), Task: c.Task, N: red}
		} else {
			red = 0
		}
		_ = rr.store.Append(rr.in.Session, w.key, e)
	}
	if st, err := rr.store.Read(rr.in.Session, w.key); err == nil {
		w.st = st
	}
}

// noteClimbs notes, once the call is allowed, each bonsai ladder --task <id> its shell line runs.
func (rr *roleRun) noteClimbs(w *who) {
	if w == nil || !w.counted {
		return
	}
	for _, task := range rr.climbs {
		_ = rr.store.Append(rr.in.Session, w.key, agents.Event{E: agents.EventClimb, At: agents.StampMilli(rr.now), Task: task})
	}
}

// spentRefusal is the refusal past a budget: the first one in full, later ones short; each names a newer brief to
// the agent's address already waiting (waitingBrief).
func (rr *roleRun) spentRefusal(w *who, c counted) Decision {
	e := c.spent
	measure := strings.ReplaceAll(e.Measure, "_", " ")
	addr := rr.address(w)
	waiting := ""
	if addr != "" {
		if id := rr.waitingBrief(w); id != "" {
			waiting = "; a newer brief to you is waiting: read it with bonsai message --show " + id
		}
	}
	if !c.newSpent {
		why := "budget spent (" + measure + ": " + fmt.Sprint(e.Count) + " of " + fmt.Sprint(e.Of) + "): only your report, " +
			"reading your brief, and the hand-back are allowed now" + waiting
		next := "write your report as the first refusal said, then stop"
		if waiting != "" {
			next = "read the newer brief, which gives you a fresh window; or write your report as the first refusal said, then stop"
		}
		if addr == "" {
			why = "budget spent (" + measure + ": " + fmt.Sprint(e.Count) + " of " + fmt.Sprint(e.Of) + "): every call is refused"
			next = "end your turn now with your last reply, which is your report"
		}
		return deny(RuleBudgetSpent, why, next)
	}
	why := fmt.Sprintf("budget spent (%s: %d of %d): report what you have and what is left", measure, e.Count, e.Of)
	if addr == "" {
		return deny(RuleBudgetSpent, why+"; you have no address (no brief tied to you), so every call is refused and your "+
			"last reply is your report", "end your turn now with a reply saying what you have and what is left")
	}
	to, thread := "<the brief's from>", "<the brief's thread>"
	if w.st != nil {
		if t := w.st.Window().Tie; t != nil {
			to, thread = t.From, t.Thread
		}
	}
	cmd := "bonsai message --kind report --from " + addr + " --to " + to + " --thread " + thread +
		" --status blocked --body '<what you have and what is left>'"
	why += "; your report is one shell call holding only: " + cmd + " (the body in single quotes, or --body - with a " +
		"here-document whose delimiter is quoted, <<'EOF')"
	why += waiting
	next := "write that report now; your caller re-briefs you, hands the work on, or asks the person"
	if waiting != "" {
		next = "read the newer brief, which gives you a fresh window; or write that report now"
	}
	return deny(RuleBudgetSpent, why, next)
}

// briefPath reports the brief id a Read of a path reads, "" for a path that is no message file in the main checkout's
// .bonsai/local/messages/.
func (rr *roleRun) briefPath(p string) string {
	if p == "" {
		return ""
	}
	base := filepath.Base(filepath.FromSlash(strings.ReplaceAll(p, `\`, "/")))
	id, ok := strings.CutSuffix(base, ".md")
	if !ok || !message.IsID(id) {
		return ""
	}
	abs := filepath.FromSlash(strings.ReplaceAll(p, `\`, "/"))
	if !filepath.IsAbs(abs) && rr.in.Cwd != "" {
		abs = filepath.Join(rr.in.Cwd, abs)
	}
	folder := filepath.Join(rr.main, filepath.FromSlash(message.Dir))
	rel, err := filepath.Rel(folder, filepath.Clean(abs))
	if err != nil || strings.HasPrefix(filepath.ToSlash(rel), "../") || strings.Count(filepath.ToSlash(rel), "/") != 1 {
		return ""
	}
	return id
}

// exempt reports whether a call past the budget is one of the three it may still make (budget-exempt), and whether it
// is the report: the hand-back tool; reading a brief to its own address (its tied brief, or a newer one, which ties at
// its read before the budget is judged); the report in its exact form (reportCall).
func (rr *roleRun) exempt(w *who, reads string) (ok, report bool) {
	in := rr.in
	if listHas(HandBackTools, in.Tool) {
		return true, false
	}
	addr := rr.address(w)
	if addr == "" {
		return false, false
	}
	if reads != "" {
		if b, why := readBrief(rr.main, reads); why == "" && b.kind == format.KindBrief && b.to == addr {
			return true, false // its tied brief, or another to its address (a newer one ties at its read, count)
		}
	}
	if shellTool(in.Tool) && in.Tool != "Monitor" && reportCall(in.Tool, in.Command, addr) {
		return true, true
	}
	return false, false
}
