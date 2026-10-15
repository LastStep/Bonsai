// Package agents keeps each agent's state for the guard's role rules and stall stops (design/plan-5.md, step 5.5,
// "Role rules and stall stops, written once"; 5.5.11 note 5): Bonsai's own internal files with no format, which may
// change. What the studio reads is bonsai agents (bonsai.agents/1, step 5.5.12), which reads them through this package.
//
// Where: the main checkout's .bonsai/local/agents/ (Folder), one folder per session (its session id), written only by
// Bonsai as a program (the guard): the engine's deny rule stops the file tools there, the delete check an rm of it,
// and the guard, from every agent but the person's session, the shell writes it can read (internal/guard).
//
// In a session's folder (SessionDir):
//   - <agent>.state, one per agent (an agent's id, or a top-level session's own id for a session playing a role): its
//     first line its Header (its role, the role's facts and the limits in force as read, its type, the time of its
//     first call), made by an exclusive create, so of calls made side by side exactly one writes it; then one Event a
//     line, appended: the link to the call that started it, a tie (its brief, the limits in force for it, a fresh
//     window), a climb noted, a red or a reset line, spent, reported, a newer brief offered by a SendMessage.
//   - <agent>.calls: one fixed-length line a tool call the guard saw from it (CallLine: its time and the minutes figure
//     the guard judged it by), appended whole, so calls made side by side lose no count; the count is the file's
//     length over the line's, read with one stat, so a long session costs no more per call.
//   - spawns.ndjson: one Spawn a line, each Agent, Task or Workflow call the guard allowed from this session (its
//     caller, its tool_use_id, the type it asked for and the role that is, a name it gave, the briefs its prompt names).
//   - claims/<tool_use_id>: made by an exclusive create when an agent is joined to the Agent call that started it,
//     naming the agent, so one call starts one agent.
//
// Beside the session folders, ties/<brief id>: made by an exclusive create when an agent ties a brief, naming it
// (<session>/<agent>), so a brief is tied by one agent only.
//
// Every write goes through internal/record's one append path (Windows' busy errors retried, a missing
// .bonsai/.gitignore restored) or an exclusive create retried the same way, each after a look that the file is not
// there yet or is a regular file; every read opens regular files only (regular.go), so a FIFO, a device, a folder or a
// link planted here is named, never opened. An agent's state that cannot be read (MoveAside) is moved aside under
// a name of its own in its session's folder, kept and listed (bonsai agents names it in skipped), and the agent is
// counted afresh from its next call (the guard says so in a guard record): a budget bounds time, not proof (step
// 5.5's "What it cannot do").
package agents

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Folder is the agents folder inside .bonsai/local/; TiesFolder the brief ties' folder inside it; ClaimsFolder the
// claims' folder inside a session's.
const (
	Folder       = "agents"
	TiesFolder   = "ties"
	ClaimsFolder = "claims"
	spawnsFile   = "spawns.ndjson"
	stateExt     = ".state"
	callsExt     = ".calls"
)

// maxLine caps a state, spawn or tie line, in bytes; maxState a state file, and maxSpawns a session's spawn calls,
// as a reader reads them whole.
const (
	maxLine   = 4096
	maxState  = 1 << 20
	maxSpawns = 4 << 20
)

// Store is one main checkout's agents folder.
type Store struct {
	Main string // the main checkout's top folder, whose .bonsai/local/ holds it
}

// Dir is the agents folder's path.
func (s Store) Dir() string {
	return filepath.Join(s.Main, filepath.FromSlash(workspace.LocalDir), Folder)
}

// Exists reports whether the agents folder is there: the guard has kept agents' state in this checkout.
func (s Store) Exists() bool {
	fi, err := os.Stat(s.Dir())
	return err == nil && fi.IsDir()
}

// rel is a path inside .bonsai/local/, forward slashes, for record.Append.
func rel(parts ...string) string { return Folder + "/" + strings.Join(parts, "/") }

// validName holds a session's, an agent's or a call's id to what can name a file (letters, digits, _ and -, at most
// 128), so none can reach outside its folder.
func validName(s string) error {
	if s == "" || len(s) > 128 {
		return fmt.Errorf("the id %q cannot name a state file", s)
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return fmt.Errorf("the id %q cannot name a state file", s)
		}
	}
	return nil
}

// Limit is one limit in force: its value (what the agent is held to) and its max; nil for no limit.
type Limit struct {
	Value *int64 `json:"value"`
	Max   *int64 `json:"max"`
}

// Limits are an agent's three limits in force, in format.RoleLimits' order.
type Limits struct {
	Minutes Limit `json:"minutes"`
	Actions Limit `json:"actions"`
	RedRuns Limit `json:"red_runs"`
}

// Header is an agent's first line: who it is and what holds it, as the guard read it at its first call.
type Header struct {
	Agent    string `json:"agent"`        // Claude Code's agent id, or a top-level session's own id
	Session  string `json:"session"`      // the session it runs in (a subagent's is its parent session's)
	Type     string `json:"type"`         // the payload's agent type, as written ("" for none)
	Role     string `json:"role"`         // the declared role it plays, by name; "" for an undeclared type
	Pack     string `json:"pack"`         // the pack declaring that role; ""
	Kind     string `json:"kind"`         // lead or helper (an undeclared type is held as a helper)
	SessRole bool   `json:"session_role"` // it plays the role the person's own session plays: a top-level session, never counted
	Writes   bool   `json:"writes"`       // the role's writes
	MaxTurns *int64 `json:"max_turns"`    // the role's max_turns; nil for none
	Limits   Limits `json:"limits"`       // the role's limits in force (its default as value), before any brief
	Named    bool   `json:"named"`        // a top-level session playing a role (--agent, or BONSAI_ROLE)
	Task     string `json:"task"`         // a named session's BONSAI_TASK; ""
	First    string `json:"first"`        // the time of its first call, UTC (Stamp)
	ReadFrom string `json:"read_from"`    // where its role facts were read: "HEAD" (as committed) or "files" (HEAD held neither)
	Note     string `json:"note"`         // why the facts were read from the files, or ""
	// Unbacked: its type names a declared role, but no call the guard allowed for that role started it, so it is held as
	// a type no pack declares (Role ""); false for every other agent.
	Unbacked bool `json:"unbacked,omitempty"`
}

// The events of an agent's state, after its header.
const (
	EventLink     = "link"     // the call that started it: Parent, Call, Briefs
	EventTie      = "tie"      // a brief tied: Brief, To, Thread, Limits, Calls (the window starts after them)
	EventOffer    = "offer"    // a newer brief named by a SendMessage resuming it: Brief, From (the sender's role)
	EventClimb    = "climb"    // the guard passed a bonsai ladder --task <Task>: its result is read at a later call
	EventRed      = "red"      // a climb's result read red: N, the count in a row
	EventReset    = "reset"    // a climb's result read green: the count back to 0
	EventSpent    = "spent"    // past a limit: Measure, Count, Of
	EventReported = "reported" // its report call passed the guard past its budget: a note only, which the stop gate never trusts (it finds the report itself, ReportAfter)
	EventTried    = "tried"    // an offered brief was tried at a later call (tied or not): Brief
)

// Event is one line of an agent's state after its header.
type Event struct {
	E       string   `json:"e"`
	At      string   `json:"at"`
	Parent  string   `json:"parent,omitempty"`
	Call    string   `json:"call,omitempty"`
	Briefs  []string `json:"briefs,omitempty"`
	Brief   string   `json:"brief,omitempty"`
	To      string   `json:"to,omitempty"`
	Thread  string   `json:"thread,omitempty"`
	From    string   `json:"from,omitempty"`
	Limits  *Limits  `json:"limits,omitempty"`
	Calls   int64    `json:"calls,omitempty"`
	Task    string   `json:"task,omitempty"`
	N       int64    `json:"n,omitempty"`
	Measure string   `json:"measure,omitempty"`
	Count   int64    `json:"count,omitempty"`
	Of      int64    `json:"of,omitempty"`
}

// Stamp writes a time as the state files keep it: UTC, to the second. StampMilli writes one to the millisecond (an
// event a later result is compared with: a climb, an offer, a spawn call).
func Stamp(t time.Time) string { return t.UTC().Format(time.RFC3339) }

// StampMilli writes a time to the millisecond, UTC.
func StampMilli(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000Z07:00") }

// ParseStamp reads a Stamp or a StampMilli.
func ParseStamp(s string) (time.Time, error) { return time.Parse(time.RFC3339Nano, s) }

// SessionDir is a session's folder.
func (s Store) SessionDir(session string) string { return filepath.Join(s.Dir(), session) }

func (s Store) statePath(session, agent string) string {
	return filepath.Join(s.SessionDir(session), agent+stateExt)
}

func (s Store) callsPath(session, agent string) string {
	return filepath.Join(s.SessionDir(session), agent+callsExt)
}

// Create makes an agent's state with its header, by an exclusive create: made is false when another call made it
// first (its header then stands).
func (s Store) Create(h Header) (made bool, err error) {
	if err := validName(h.Session); err != nil {
		return false, err
	}
	if err := validName(h.Agent); err != nil {
		return false, err
	}
	line, err := json.Marshal(h)
	if err != nil {
		return false, err
	}
	return s.exclusive(s.statePath(h.Session, h.Agent), append(line, '\n'))
}

// exclusive makes a file holding body, failing with made false when it is there already; a busy error on Windows is
// retried as the log's appends are, and a missing .bonsai/.gitignore restored first.
func (s Store) exclusive(path string, body []byte) (made bool, err error) {
	if len(body) > maxLine {
		return false, fmt.Errorf("a state line of %d bytes is over the cap of %d", len(body), maxLine)
	}
	_, _ = workspace.EnsureGitignore(s.Main)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	deadline := time.Now().Add(time.Second)
	delay := 5 * time.Millisecond
	for {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			_, werr := f.Write(body)
			cerr := f.Close()
			return true, errors.Join(werr, cerr)
		}
		if errors.Is(err, fs.ErrExist) {
			return false, nil
		}
		if !workspace.IsBusy(err) || time.Now().After(deadline) {
			return false, err
		}
		time.Sleep(delay)
		if delay < 80*time.Millisecond {
			delay *= 2
		}
	}
}

// Append adds one event to an agent's state.
func (s Store) Append(session, agent string, e Event) error {
	if err := validName(session); err != nil {
		return err
	}
	if err := validName(agent); err != nil {
		return err
	}
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if err := appendable(s.statePath(session, agent)); err != nil {
		return err
	}
	_, err = record.Append(s.Main, rel(session, agent+stateExt), append(line, '\n'), maxLine)
	return err
}

// CallLen is the length of one call line: its time (Stamp, 20 characters), a space, the minutes figure in six digits,
// and a line feed.
const CallLen = 28

// maxMinutes is the largest minutes figure a call line holds (six digits).
const maxMinutes = 999999

// CallLine is one call line.
func CallLine(at time.Time, minutes int64) []byte {
	if minutes < 0 {
		minutes = 0
	}
	if minutes > maxMinutes {
		minutes = maxMinutes
	}
	return []byte(fmt.Sprintf("%s %06d\n", Stamp(at), minutes))
}

// Call appends one call line for an agent and gives the count of call lines in its file after it (its own among
// them, and any written side by side with it).
func (s Store) Call(session, agent string, at time.Time, minutes int64) (int64, error) {
	if err := validName(session); err != nil {
		return 0, err
	}
	if err := validName(agent); err != nil {
		return 0, err
	}
	if err := appendable(s.callsPath(session, agent)); err != nil {
		return 0, err
	}
	if _, err := record.Append(s.Main, rel(session, agent+callsExt), CallLine(at, minutes), CallLen); err != nil {
		return 0, err
	}
	return s.Calls(session, agent)
}

// Calls gives the count of an agent's call lines: its calls file's length over CallLen (0 when there is none), with
// one look and no open; a calls file that is not a regular file is an error.
func (s Store) Calls(session, agent string) (int64, error) {
	fi, err := os.Lstat(s.callsPath(session, agent))
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if !fi.Mode().IsRegular() {
		return 0, notRegular(fi.Mode())
	}
	return fi.Size() / CallLen, nil
}

// LastCall gives an agent's last call line's time and minutes figure; ok false when it has none that reads.
func (s Store) LastCall(session, agent string) (at time.Time, minutes int64, ok bool) {
	f, size, err := openRegular(s.callsPath(session, agent))
	if err != nil {
		return time.Time{}, 0, false
	}
	defer func() { _ = f.Close() }()
	if size < CallLen {
		return time.Time{}, 0, false
	}
	b := make([]byte, CallLen)
	if _, err := f.ReadAt(b, (size/CallLen-1)*CallLen); err != nil && !errors.Is(err, io.EOF) {
		return time.Time{}, 0, false
	}
	at, err = ParseStamp(string(b[:20]))
	if err != nil {
		return time.Time{}, 0, false
	}
	n, err := strconv.ParseInt(string(b[21:27]), 10, 64)
	if err != nil {
		return time.Time{}, 0, false
	}
	return at, n, true
}

// State is an agent's state as read: its header and its events, in order.
type State struct {
	Header Header
	Events []Event
}

// ErrNoState is Read's answer for an agent with no state yet.
var ErrNoState = errors.New("no state")

// Read reads an agent's state: ErrNoState when it has none, another error when its file does not read (its header
// is not whole yet when another call is still writing it: the caller may read again), is not a regular file or is
// larger than maxState, or when its calls file is there and is not a regular file (regular.go). An event line that
// does not read is passed over.
func (s Store) Read(session, agent string) (*State, error) {
	if err := validName(session); err != nil {
		return nil, err
	}
	if err := validName(agent); err != nil {
		return nil, err
	}
	raw, err := ReadRegular(s.statePath(session, agent), maxState)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNoState
	}
	if err != nil {
		return nil, err
	}
	if err := appendable(s.callsPath(session, agent)); err != nil {
		return nil, fmt.Errorf("its calls file: %v", err)
	}
	first, rest, ok := bytes.Cut(raw, []byte("\n"))
	if !ok {
		return nil, errors.New("its header is not whole")
	}
	st := &State{}
	if err := json.Unmarshal(first, &st.Header); err != nil {
		return nil, fmt.Errorf("its header does not read: %v", err)
	}
	sc := bufio.NewScanner(bytes.NewReader(rest))
	sc.Buffer(make([]byte, 0, 1024), maxLine+1)
	for sc.Scan() {
		var e Event
		if json.Unmarshal(sc.Bytes(), &e) == nil && e.E != "" {
			st.Events = append(st.Events, e)
		}
	}
	return st, nil
}

// ReadWait reads an agent's state, waiting up to wait while its header is not whole (made side by side by another
// call): ErrNoState when it has none.
func (s Store) ReadWait(session, agent string, wait time.Duration) (*State, error) {
	deadline := time.Now().Add(wait)
	for {
		st, err := s.Read(session, agent)
		if err == nil || errors.Is(err, ErrNoState) || time.Now().After(deadline) {
			return st, err
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// Window is an agent's budget window: since its last tie, else since its first call.
type Window struct {
	Start  time.Time // its start: the tie's time, else the first call's
	Calls  int64     // the call lines before it
	Limits Limits    // the limits in force in it
	Tie    *Event    // the tie that opened it; nil for the first window
}

// Window gives the window in force.
func (st *State) Window() Window {
	start, _ := ParseStamp(st.Header.First)
	w := Window{Start: start, Limits: st.Header.Limits}
	for i := range st.Events {
		e := &st.Events[i]
		if e.E != EventTie {
			continue
		}
		at, err := ParseStamp(e.At)
		if err != nil {
			continue
		}
		w.Start, w.Calls, w.Tie = at, e.Calls, e
		if e.Limits != nil {
			w.Limits = *e.Limits
		}
	}
	return w
}

// Link gives the event that joined the agent to the call that started it; nil when none did.
func (st *State) Link() *Event {
	for i := range st.Events {
		if st.Events[i].E == EventLink {
			return &st.Events[i]
		}
	}
	return nil
}

// RedRuns gives the failed climbs in a row: the last red line's count, 0 after a reset or with none.
func (st *State) RedRuns() int64 {
	var n int64
	for _, e := range st.Events {
		switch e.E {
		case EventRed:
			n = e.N
		case EventReset:
			n = 0
		}
	}
	return n
}

// Spent gives the spent line in the window in force, nil when the window is not spent.
func (st *State) Spent() *Event {
	var spent *Event
	for i := range st.Events {
		switch st.Events[i].E {
		case EventTie:
			spent = nil
		case EventSpent:
			if spent == nil {
				spent = &st.Events[i]
			}
		}
	}
	return spent
}

// Reported reports a reported line after the spent line of the window in force: that the report's call was let
// through, not that the report was written (the stop gate reads the report itself, ReportAfter).
func (st *State) Reported() bool {
	spent := false
	reported := false
	for _, e := range st.Events {
		switch e.E {
		case EventTie:
			spent, reported = false, false
		case EventSpent:
			spent = true
		case EventReported:
			reported = reported || spent
		}
	}
	return reported
}

// OpenClimbs gives the climbs noted with no red or reset line after them for their task, oldest first.
func (st *State) OpenClimbs() []Event {
	var open []Event
	for _, e := range st.Events {
		switch e.E {
		case EventClimb:
			open = append(open, e)
		case EventRed, EventReset:
			var keep []Event
			for _, c := range open {
				if c.Task != e.Task {
					keep = append(keep, c)
				}
			}
			open = keep
		}
	}
	return open
}

// Offers gives the briefs offered by a SendMessage and not tried yet, in order.
func (st *State) Offers() []Event {
	var out []Event
	for _, e := range st.Events {
		switch e.E {
		case EventOffer:
			out = append(out, e)
		case EventTried, EventTie:
			var keep []Event
			for _, o := range out {
				if o.Brief != e.Brief {
					keep = append(keep, o)
				}
			}
			out = keep
		}
	}
	return out
}

// Spawn is one Agent, Task or Workflow call the guard allowed from a session.
type Spawn struct {
	At     string   `json:"at"`
	Tool   string   `json:"tool"`   // Agent, Task or Workflow
	Use    string   `json:"use"`    // its tool_use_id
	Caller string   `json:"caller"` // the calling agent's id; "" for the person's session; a named session's own id
	Parent string   `json:"parent"` // as a spawn record names it: "session", the calling agent's id, a named session's id
	Type   string   `json:"type"`   // the agent type a child of it carries (Claude Code's, as measured): the type asked
	Role   string   `json:"role"`   // the declared role that type plays, by name; "" for an undeclared type
	Name   string   `json:"name"`   // the name it gave, ""
	Briefs []string `json:"briefs"` // the brief ids its prompt names, in order
}

// AddSpawn records a spawn call of a session.
func (s Store) AddSpawn(session string, sp Spawn) error {
	if err := validName(session); err != nil {
		return err
	}
	line, err := json.Marshal(sp)
	if err != nil {
		return err
	}
	if len(line)+1 > maxLine {
		sp.Briefs = sp.Briefs[:min(len(sp.Briefs), 4)]
		if line, err = json.Marshal(sp); err != nil {
			return err
		}
	}
	if err := appendable(filepath.Join(s.SessionDir(session), spawnsFile)); err != nil {
		return err
	}
	_, err = record.Append(s.Main, rel(session, spawnsFile), append(line, '\n'), maxLine)
	return err
}

// Spawns reads a session's spawn calls, in the order they were allowed (none when it has none).
func (s Store) Spawns(session string) ([]Spawn, error) {
	if err := validName(session); err != nil {
		return nil, err
	}
	raw, err := ReadRegular(filepath.Join(s.SessionDir(session), spawnsFile), maxSpawns)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Spawn
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 0, 1024), maxLine+1)
	for sc.Scan() {
		var sp Spawn
		if json.Unmarshal(sc.Bytes(), &sp) == nil && sp.Use != "" {
			out = append(out, sp)
		}
	}
	return out, nil
}

// Claim joins an agent to the call that started it, by an exclusive create of the call's claim: ok is true when the
// claim is this agent's (made now, or made before for it), false when another agent holds it.
func (s Store) Claim(session, use, agent string) (ok bool, err error) {
	for _, n := range []string{session, use, agent} {
		if err := validName(n); err != nil {
			return false, err
		}
	}
	path := filepath.Join(s.SessionDir(session), ClaimsFolder, use)
	made, err := s.exclusive(path, []byte(agent+"\n"))
	if err != nil || made {
		return made, err
	}
	return s.holder(path) == agent, nil
}

// Claimed reports whether a call has been claimed by an agent.
func (s Store) Claimed(session, use string) bool {
	if validName(session) != nil || validName(use) != nil {
		return false
	}
	_, err := os.Lstat(filepath.Join(s.SessionDir(session), ClaimsFolder, use))
	return err == nil
}

// MarkSpent marks an agent's budget window spent, by an exclusive create of claims/spent-<agent>-<window> (window
// the count of ties before it, State.Ties), so of calls made side by side exactly one writes the window's spent line
// (VG's item K): made is false when another call marked it first.
func (s Store) MarkSpent(session, agent string, window int) (made bool, err error) {
	if err := validName(session); err != nil {
		return false, err
	}
	if err := validName(agent); err != nil {
		return false, err
	}
	name := fmt.Sprintf("spent-%s-%d", agent, window)
	if len(name) > 128 {
		name = fmt.Sprintf("spent-%s-%d", agent[:128-len(name)+len(agent)], window)
	}
	return s.exclusive(filepath.Join(s.SessionDir(session), ClaimsFolder, name), []byte(agent+"\n"))
}

// Ties counts the tie lines in an agent's state: the windows before the one in force.
func (st *State) Ties() int {
	n := 0
	for _, e := range st.Events {
		if e.E == EventTie {
			n++
		}
	}
	return n
}

// ClaimHolder gives the agent that claimed a call, "" when none has (or its claim does not read).
func (s Store) ClaimHolder(session, use string) string {
	if validName(session) != nil || validName(use) != nil || !s.Claimed(session, use) {
		return ""
	}
	return s.holder(filepath.Join(s.SessionDir(session), ClaimsFolder, use))
}

// holder reads the agent a claim or tie file names, waiting briefly for one being written; "" for a file that is not
// a regular file (regular.go), or does not read.
func (s Store) holder(path string) string {
	deadline := time.Now().Add(200 * time.Millisecond)
	for {
		raw, err := ReadRegular(path, maxLine)
		if err == nil && bytes.HasSuffix(raw, []byte("\n")) {
			return strings.TrimSuffix(string(raw), "\n")
		}
		if errors.Is(err, ErrNotRegular) || time.Now().After(deadline) {
			return ""
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// Tie claims a brief for one agent (session/agent), by an exclusive create of its tie file: ok is true when the tie
// is this agent's (made now, or made before for it), false when another agent tied it.
func (s Store) Tie(brief, session, agent string) (ok bool, err error) {
	for _, n := range []string{brief, session, agent} {
		if err := validName(n); err != nil {
			return false, err
		}
	}
	path := filepath.Join(s.Dir(), TiesFolder, brief)
	who := session + "/" + agent
	made, err := s.exclusive(path, []byte(who+"\n"))
	if err != nil || made {
		return made, err
	}
	return s.holder(path) == who, nil
}

// TiedBy gives the agent (session/agent) that tied a brief, "" for none.
func (s Store) TiedBy(brief string) string {
	if validName(brief) != nil {
		return ""
	}
	path := filepath.Join(s.Dir(), TiesFolder, brief)
	if _, err := os.Lstat(path); err != nil {
		return ""
	}
	return s.holder(path)
}

// Agents lists the agents with state in a session's folder, by name order (none when it has none): every name
// <agent>.state there, whatever it is, so a reader names one that is not a regular file (a folder, a FIFO) rather than
// pass it over unseen.
func (s Store) Agents(session string) ([]string, error) {
	if err := validName(session); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.SessionDir(session))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if name, ok := strings.CutSuffix(e.Name(), stateExt); ok && validName(name) == nil {
			out = append(out, name)
		}
	}
	return out, nil
}
