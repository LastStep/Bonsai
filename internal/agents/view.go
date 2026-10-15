package agents

// The view: bonsai agents (bonsai.agents/1; design/plan-5.md, note 5.5.12; Rohan, 14 Oct: "it will not interpret
// anything, give any prediction, do any unnessary calculations"), a read-only fold of what Bonsai has recorded about
// each agent, copied as recorded, field by field. It reads three things and writes nothing:
//
//   - The guard's state under the main checkout's .bonsai/local/agents/ (this package): one entry for each state file
//     the agents kind still keeps, its role, type, start, tied brief, parent, limits, counts and spent as the guard wrote
//     them (Recorded's rules: the limits on its start or last tie line, minutes from its last call line, actions its
//     call lines since its window's start, red runs from its last red or reset line). Read strictly: a state that does
//     not read whole (its first line not a start line, an event line that does not read, a line cut short, a calls file
//     that is not whole lines or holds fewer calls than its last tie line counts, a value bonsai.agents/1 cannot show) is
//     skipped and named in Skipped, never shown with counts it does not hold, and never an error for the rest.
//   - The log, .bonsai/local/log/s-<session>.ndjson: a session's first session_start (its start, for a session with no
//     state) and its newest (its model and BONSAI_TASK); its newest session_end; each subagent's newest subagent_stop
//     (its end); each agent's newest tool_start (its last call: at, tool, category); each agent's newest spawn record
//     (its parent); and each agent's newest usage records (those of its newest end, which share one at: copied, never
//     summed). Records are read leanly, by their event first: only those six events are decoded, each field held to the
//     log schema's own pattern, and a line that does not read is passed over (record.ReadLog's whole-schema check of
//     every record costs about 90 microseconds a record, more than a view polled often may spend). "Newest" is by at,
//     a later line winning a tie. A session's own records are those with no subagent_id.
//   - The messages (internal/message, as read): the tied brief's task; and the agent's report, the newest report in its
//     tied brief's thread whose from is its own address (its tied brief's to). A report from any other address is never
//     taken.
//
// Each person's session, a session with a log file and no state of its own, is an entry from the log alone: its
// counts, limits, role, brief and report null (the guard keeps no state for it). A subagent with no state (in a project
// whose packs declare no roles, or one that made no call) is not listed. Nothing is derived: no state such as working
// or quiet, no share of a budget, no minutes worked out against the clock; and no setting is read, so the limits shown
// are the ones that bound each agent, never today's. Entries are sorted by started (null last), then id, then session;
// Skipped by path. What was recorded under an agent's name is attributed, not authenticated.

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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Viewed is what the view read: every agent, and every state file it skipped.
type Viewed struct {
	Agents  []format.Agent
	Skipped []format.AgentSkipped
}

// maxState caps a state file the view reads, in bytes.
const maxState = 1 << 20

// View reads every agent of the main checkout (this file's comment). msgs is its messages as read (message.Read). Only
// an agents or log folder that is there and cannot be listed is an error; everything below them that does not read is
// skipped.
func (s Store) View(msgs *message.Store) (*Viewed, error) {
	logs, err := s.readLogs()
	if err != nil {
		return nil, err
	}
	v := &Viewed{Agents: []format.Agent{}, Skipped: []format.AgentSkipped{}}
	seen := map[string]bool{} // session/agent of each state file, read or skipped
	var sessions []string
	switch fi, err := os.Lstat(s.Dir()); {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return nil, err
	case !Plain(s.Dir()) || !fi.IsDir():
		v.skip(local(Folder), "it is not a plain folder (a link or a file where Bonsai keeps its folder)")
	default:
		if sessions, err = s.Sessions(); err != nil {
			return nil, err
		}
	}
	for _, session := range sessions {
		names, err := s.Agents(session)
		if err != nil {
			v.skip(local(rel(session)), "the session's folder cannot be listed ("+fault(err)+")")
			continue
		}
		for _, agent := range names {
			seen[session+"/"+agent] = true
			path := local(rel(session, agent+stateExt))
			a, why := s.viewOne(session, agent, logs[session], msgs)
			if why != "" {
				v.skip(path, why)
				continue
			}
			v.add(a, path)
		}
	}
	for session, lf := range logs { // sorted below
		if !seen[session+"/"+session] { // a session with no state of its own (a named session's, read or skipped, stands)
			v.add(personOf(session, lf), lf.path)
		}
	}
	sort.Slice(v.Agents, func(i, j int) bool { return less(&v.Agents[i], &v.Agents[j]) })
	sort.Slice(v.Skipped, func(i, j int) bool { return v.Skipped[i].Path < v.Skipped[j].Path })
	return v, nil
}

// skip names a file the view skipped, by its project-relative path.
func (v *Viewed) skip(path, why string) {
	v.Skipped = append(v.Skipped, format.AgentSkipped{Path: path, Reason: why})
}

// local is a path inside .bonsai/local/ as a project-relative path.
func local(rel string) string { return workspace.LocalDir + "/" + rel }

// add keeps an entry that fits bonsai.agents/1, and skips one with a value it cannot show (a forged line).
func (v *Viewed) add(a format.Agent, from string) {
	if _, err := (&format.Agents{Agents: []format.Agent{a}, Skipped: []format.AgentSkipped{}}).Encode(); err != nil {
		v.skip(from, "a value it holds does not fit bonsai.agents/1: "+oneLine(err))
		return
	}
	v.Agents = append(v.Agents, a)
}

// less orders entries by started (null last), then id, then session.
func less(a, b *format.Agent) bool {
	switch {
	case (a.Started == nil) != (b.Started == nil):
		return a.Started != nil
	case a.Started != nil && *a.Started != *b.Started:
		return *a.Started < *b.Started
	case a.ID != b.ID:
		return a.ID < b.ID
	}
	return a.Session < b.Session
}

// viewOne reads one agent's state strictly and lays out its entry, or says why it is skipped.
func (s Store) viewOne(session, agent string, lf *logFacts, msgs *message.Store) (format.Agent, string) {
	st, calls, minutes, why := s.strict(session, agent)
	if why != "" {
		return format.Agent{}, why
	}
	h := st.Header
	w := st.Window()
	if calls < w.Calls {
		return format.Agent{}, fmt.Sprintf("its calls file holds %d calls, fewer than the %d its last tie line counts before "+
			"its window", calls, w.Calls)
	}
	started, _ := ParseStamp(h.First) // strict read it
	a := format.Agent{ID: agent, Session: session, Role: orNull(h.Role), Type: orNull(h.Type),
		Started: ptr(started.UTC().Format(record.AtLayout)), Spent: st.Spent() != nil}
	if !h.SessRole { // the session's role is never counted, and has no limits
		a.Counts = &format.AgentCounts{Actions: calls - w.Calls, RedRuns: st.RedRuns()}
		if minutes != nil {
			n := json.Number(strconv.FormatInt(*minutes, 10))
			a.Counts.Minutes = &n
		}
		lim := func(l Limit) format.AgentLimit { return format.AgentLimit{Value: l.Value, Max: l.Max} }
		a.Limits = &format.AgentLimits{Minutes: lim(w.Limits.Minutes), Actions: lim(w.Limits.Actions), RedRuns: lim(w.Limits.RedRuns)}
	}
	if w.Tie != nil {
		a.Brief, a.Thread = ptr(w.Tie.Brief), ptr(w.Tie.Thread)
		if e, _ := msgs.Find(w.Tie.Brief); e != nil && e.Msg.Task != nil && message.IsTask(*e.Msg.Task) {
			a.Task = ptr(*e.Msg.Task)
		}
		a.Report = reportOf(msgs, w.Tie.Thread, w.Tie.To)
	}
	if a.Task == nil && message.IsTask(h.Task) {
		a.Task = ptr(h.Task)
	}
	key := agent
	if agent == session {
		key = "" // a top-level session playing a role: the session's own records
	}
	if link := st.Link(); link != nil && key != "" {
		a.Parent = orNull(link.Parent)
	}
	if lf != nil {
		a.Log = ptr(lf.path)
		lf.fill(&a, key)
	}
	return a, ""
}

// personOf lays out a session's entry from its log alone: the person's own session (or a session in a project whose
// packs declare no roles), which the guard keeps no state for.
func personOf(session string, lf *logFacts) format.Agent {
	a := format.Agent{ID: session, Session: session, Log: ptr(lf.path)}
	if lf.firstStart != nil {
		a.Started = ptr(lf.firstStart.At)
	}
	if lf.lastStart != nil && lf.lastStart.Task != nil && message.IsTask(*lf.lastStart.Task) {
		a.Task = ptr(*lf.lastStart.Task)
	}
	lf.fill(&a, "")
	return a
}

// reportOf gives the newest report in a thread whose from is the agent's own address, nil for none.
func reportOf(msgs *message.Store, thread, own string) *format.AgentReport {
	if msgs == nil || thread == "" || own == "" {
		return nil
	}
	var r *format.AgentReport
	for _, e := range msgs.Messages { // by created, then id: the last match is the newest
		m := e.Msg
		if m.Kind == format.KindReport && m.Thread == thread && m.From == own && m.Status != nil {
			r = &format.AgentReport{ID: m.ID, Status: *m.Status}
		}
	}
	return r
}

// strict reads an agent's state whole, or says why it does not read (this file's comment): its state, its call lines'
// count, and the minutes figure on its last call line (nil for none).
func (s Store) strict(session, agent string) (*State, int64, *int64, string) {
	path := s.statePath(session, agent)
	fi, err := os.Lstat(path)
	switch {
	case err != nil:
		return nil, 0, nil, "it cannot be read (" + fault(err) + ")"
	case !fi.Mode().IsRegular():
		return nil, 0, nil, "it is not a regular file"
	case fi.Size() > maxState:
		return nil, 0, nil, fmt.Sprintf("it is larger than %d bytes", maxState)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, nil, "it cannot be read (" + fault(err) + ")"
	}
	if !bytes.HasSuffix(raw, []byte("\n")) {
		return nil, 0, nil, "its last line is not whole (cut short)"
	}
	lines := bytes.Split(bytes.TrimSuffix(raw, []byte("\n")), []byte("\n"))
	st := &State{}
	if err := json.Unmarshal(lines[0], &st.Header); err != nil {
		return nil, 0, nil, "its first line is not a start line the guard writes"
	}
	h := st.Header
	switch {
	case h.Agent != agent || h.Session != session:
		return nil, 0, nil, "its first line names another agent or session than its file"
	case h.Role != "" && !rolePattern().MatchString(h.Role):
		return nil, 0, nil, "its first line names a role that is not a role's name"
	}
	if _, err := ParseStamp(h.First); err != nil {
		return nil, 0, nil, "its first line holds no time of its first call"
	}
	if limitsBelow0(h.Limits) {
		return nil, 0, nil, "its first line holds a limit below 0"
	}
	for i, l := range lines[1:] {
		var e Event
		if json.Unmarshal(l, &e) != nil || e.E == "" {
			return nil, 0, nil, fmt.Sprintf("its line %d is not an event line the guard writes", i+2)
		}
		if why := eventWhy(&e); why != "" {
			return nil, 0, nil, fmt.Sprintf("its line %d %s", i+2, why)
		}
		st.Events = append(st.Events, e)
	}
	calls, minutes, why := s.strictCalls(session, agent)
	if why != "" {
		return nil, 0, nil, why
	}
	return st, calls, minutes, ""
}

// eventWhy says why an event's values cannot be shown, "" when they can.
func eventWhy(e *Event) string {
	if _, err := ParseStamp(e.At); err != nil {
		return "holds no time"
	}
	switch e.E {
	case EventTie:
		if !message.IsID(e.Brief) || !message.IsID(e.Thread) {
			return "ties no message id"
		}
		if e.Calls < 0 {
			return "counts calls below 0"
		}
		if e.Limits != nil && limitsBelow0(*e.Limits) {
			return "holds a limit below 0"
		}
	case EventRed:
		if e.N < 0 {
			return "counts red runs below 0"
		}
	}
	return ""
}

// limitsBelow0 reports limits that cannot be shown: a value or a max below 0.
func limitsBelow0(l Limits) bool {
	for _, v := range []*int64{l.Minutes.Value, l.Minutes.Max, l.Actions.Value, l.Actions.Max, l.RedRuns.Value, l.RedRuns.Max} {
		if v != nil && *v < 0 {
			return true
		}
	}
	return false
}

// strictCalls reads an agent's calls file: its count of call lines and the minutes figure on its last one; none when it
// has no calls file.
func (s Store) strictCalls(session, agent string) (int64, *int64, string) {
	path := s.callsPath(session, agent)
	fi, err := os.Lstat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return 0, nil, ""
	case err != nil:
		return 0, nil, "its calls file cannot be read (" + fault(err) + ")"
	case !fi.Mode().IsRegular():
		return 0, nil, "its calls file is not a regular file"
	case fi.Size()%CallLen != 0:
		return 0, nil, fmt.Sprintf("its calls file is not whole call lines (%d bytes)", fi.Size())
	case fi.Size() == 0:
		return 0, nil, ""
	}
	_, minutes, ok := s.LastCall(session, agent)
	if !ok {
		return 0, nil, "its calls file's last line does not read"
	}
	return fi.Size() / CallLen, &minutes, ""
}

// logFacts is what one session's log file holds for the view (this file's comment).
type logFacts struct {
	path       string   // the file, project-relative, forward slashes
	firstStart *logLine // its first session_start
	lastStart  *logLine // its newest session_start
	end        *logLine // its newest session_end
	stops      map[string]*logLine
	calls      map[string]*logLine   // by agent: a subagent's id, "" for the session's own
	usage      map[string][]*logLine // by agent: the records of its newest end
	spawns     map[string]*logLine   // by the subagent started
}

// fill copies an agent's facts from its session's log: key is a subagent's id, "" for the session's own records.
func (lf *logFacts) fill(a *format.Agent, key string) {
	if key == "" {
		if lf.end != nil {
			a.Ended = ptr(lf.end.At)
		}
		if lf.lastStart != nil && lf.lastStart.Model != nil {
			a.Model = ptr(*lf.lastStart.Model)
		}
	} else {
		if stop := lf.stops[key]; stop != nil {
			a.Ended = ptr(stop.At)
		}
		if sp := lf.spawns[key]; sp != nil && sp.Parent != nil {
			a.Parent = ptr(*sp.Parent)
		}
	}
	if c := lf.calls[key]; c != nil {
		a.LastCall = &format.AgentCall{At: c.At, Tool: *c.Tool, Category: c.Category}
	}
	recs := lf.usage[key]
	if len(recs) == 0 {
		return
	}
	a.Usage = []format.AgentUsage{}
	for _, u := range recs {
		item := format.AgentUsage{Kind: u.Kind, Reason: u.Reason}
		if u.Usage != nil {
			item.Model = u.Model
			item.Input, item.Output, item.CacheRead, item.CacheWrite = u.Usage.Input, u.Usage.Output, u.Usage.CacheRead, u.Usage.CacheWrite
		}
		a.Usage = append(a.Usage, item)
	}
	if a.Model == nil {
		a.Model = recs[len(recs)-1].Model // the newest record: the model of its last reply (internal/recorder's order)
	}
}

// logLine is a log record as the view decodes it: the fields it reads.
type logLine struct {
	Format     string  `json:"format"`
	At         string  `json:"at"`
	Session    *string `json:"session"`
	Event      string  `json:"event"`
	SubagentID *string `json:"subagent_id"`
	Task       *string `json:"task"`
	Tool       *string `json:"tool"`
	Category   *string `json:"category"`
	Kind       *string `json:"kind"`
	Model      *string `json:"model"`
	Reason     *string `json:"reason"`
	Parent     *string `json:"parent"`
	Usage      *struct {
		Input      *int64 `json:"input"`
		Output     *int64 `json:"output"`
		CacheRead  *int64 `json:"cache_read"`
		CacheWrite *int64 `json:"cache_write"`
	} `json:"usage"`
}

// viewEvents are the log's events the view reads.
var viewEvents = map[string]bool{"session_start": true, "session_end": true, "subagent_stop": true, "tool_start": true,
	"spawn": true, "usage": true}

var (
	atPattern = sync.OnceValue(func() *regexp.Regexp {
		return regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$`)
	})
	rolePattern = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^[a-z][a-z0-9-]*$`) })
	eventKey    = []byte(`"event":"`)
	recordStart = []byte(`{"format":"bonsai.log/`)
)

// readLogs reads every session's log file in the main checkout's log folder: its facts, by session.
func (s Store) readLogs() (map[string]*logFacts, error) {
	dir := filepath.Join(s.Main, filepath.FromSlash(workspace.LocalDir), record.LogFolder)
	files, err := record.ListLog(dir)
	if err != nil {
		return nil, err
	}
	out := map[string]*logFacts{}
	for _, name := range files.Sessions {
		session := record.SessionOf(name)
		lf := &logFacts{path: workspace.LocalDir + "/" + record.LogFolder + "/" + name, stops: map[string]*logLine{},
			calls: map[string]*logLine{}, usage: map[string][]*logLine{}, spawns: map[string]*logLine{}}
		if err := lf.read(filepath.Join(dir, name), session); err != nil {
			continue // a file that cannot be read gives what it gave; its session is still listed
		}
		out[session] = lf
	}
	return out, nil
}

// read reads one session's log file, line by line.
func (lf *logFacts) read(path, session string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReaderSize(f, 64<<10)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			if l := decode(line, session); l != nil {
				lf.note(l)
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// decode reads one line as a record of the events the view reads, its fields held to the log schema's patterns; nil
// for any other line. A torn line glued to a record gives the record (record.ReadLog's rule).
func decode(line []byte, session string) *logLine {
	i := bytes.Index(line, eventKey)
	if i < 0 {
		return nil
	}
	rest := line[i+len(eventKey):]
	j := bytes.IndexByte(rest, '"')
	if j < 0 || !viewEvents[string(rest[:j])] {
		return nil
	}
	l := &logLine{}
	if json.Unmarshal(line, l) != nil {
		k := bytes.LastIndex(line, recordStart)
		if k <= 0 {
			return nil
		}
		l = &logLine{}
		if json.Unmarshal(line[k:], l) != nil {
			return nil
		}
	}
	switch {
	case l.Format != "bonsai.log/1" && !strings.HasPrefix(l.Format, "bonsai.log/1."),
		!atPattern().MatchString(l.At), l.Session == nil || *l.Session != session, !viewEvents[l.Event],
		l.Event == "tool_start" && (l.Tool == nil || *l.Tool == ""),
		l.Event == "usage" && !usageReads(l):
		return nil
	}
	return l
}

// usageReads reports a usage record that reads: its four counts, each a whole number from 0, and its model; or kind
// unknown with no counts.
func usageReads(l *logLine) bool {
	if l.Usage == nil {
		return l.Kind != nil && *l.Kind == "unknown"
	}
	u := l.Usage
	for _, v := range []*int64{u.Input, u.Output, u.CacheRead, u.CacheWrite} {
		if v == nil || *v < 0 {
			return false
		}
	}
	return l.Model != nil && l.Kind == nil
}

// note keeps a record's facts: the first and newest of each kind, by at, a later line winning a tie.
func (lf *logFacts) note(l *logLine) {
	newer := func(cur *logLine) bool { return cur == nil || l.At >= cur.At }
	key := ""
	if l.SubagentID != nil {
		key = *l.SubagentID
	}
	switch l.Event {
	case "session_start":
		if key != "" {
			return
		}
		if lf.firstStart == nil || l.At < lf.firstStart.At {
			lf.firstStart = l
		}
		if newer(lf.lastStart) {
			lf.lastStart = l
		}
	case "session_end":
		if key == "" && newer(lf.end) {
			lf.end = l
		}
	case "subagent_stop":
		if key != "" && newer(lf.stops[key]) {
			lf.stops[key] = l
		}
	case "tool_start":
		if newer(lf.calls[key]) {
			lf.calls[key] = l
		}
	case "spawn":
		if key != "" && newer(lf.spawns[key]) {
			lf.spawns[key] = l
		}
	case "usage":
		cur := lf.usage[key]
		switch {
		case len(cur) == 0 || l.At > cur[0].At:
			lf.usage[key] = []*logLine{l}
		case l.At == cur[0].At:
			lf.usage[key] = append(cur, l)
		}
	}
}

// orNull gives "" as nil, any other text as a pointer to it.
func orNull(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptr(s string) *string { return &s }

// oneLine keeps an error's words on one line, its paths' backslashes as forward slashes.
func oneLine(err error) string {
	return strings.ReplaceAll(strings.Join(strings.Fields(err.Error()), " "), `\`, "/")
}

// fault names a file error without its path (a home folder's path holds the user's name).
func fault(err error) string {
	switch {
	case errors.Is(err, fs.ErrPermission):
		return "permission denied"
	case errors.Is(err, fs.ErrNotExist):
		return "not there"
	}
	return "an error of the file system"
}
