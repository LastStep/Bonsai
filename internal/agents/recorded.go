package agents

// What the views read of an agent's state (design/plan-5.md, step 5.5, "How it is counted": "what the studio reads is
// bonsai agents ... the limits and counts as the guard wrote them, never worked out again"; note 5.5.11, 9): an
// agent's counts and limits as recorded (Recorded), the agent that tied a brief (TiedTo), and, for bonsai messages, a
// thread's recipient and the two flags that read an agent's state (Thread).
//
//   - Its limits: as the guard wrote them on its start line, or on its last tie line (the brief's where it gave one).
//   - Its counts: minutes, the figure the guard wrote on its last call line; actions, its call lines since its window's
//     start (the file's length, as the guard counts them); red runs, from its last red or reset line.
//   - Spent: the guard has written a spent line in the window in force.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/message"
)

// Recorded is an agent's state as the guard recorded it.
type Recorded struct {
	Session  string
	Agent    string
	Header   Header
	Limits   Limits    // the limits in force: its last tie line's, else its start line's
	Minutes  *int64    // the minutes figure on its last call line; nil when it has made no call the guard counted
	Actions  int64     // its call lines since its window's start
	RedRuns  int64     // its last red line's count; 0 after a reset, or with none
	Spent    bool      // a spent line in the window in force
	LastCall time.Time // its last call line's time; the zero time when it has none
}

// Recorded reads one agent's state as recorded (ErrNoState when it has none).
func (s Store) Recorded(session, agent string) (*Recorded, error) {
	st, err := s.Read(session, agent)
	if err != nil {
		return nil, err
	}
	w := st.Window()
	r := &Recorded{Session: session, Agent: agent, Header: st.Header, Limits: w.Limits, RedRuns: st.RedRuns(),
		Spent: st.Spent() != nil}
	calls, err := s.Calls(session, agent)
	if err != nil {
		return nil, err
	}
	if r.Actions = calls - w.Calls; r.Actions < 0 {
		r.Actions = 0
	}
	if at, minutes, ok := s.LastCall(session, agent); ok {
		r.Minutes, r.LastCall = &minutes, at
	}
	return r, nil
}

// TiedTo gives the agent that tied a brief, by its tie file: ok false when no agent tied it, or its tie file does not
// name one.
func (s Store) TiedTo(brief string) (session, agent string, ok bool) {
	if validName(brief) != nil {
		return "", "", false
	}
	raw, err := os.ReadFile(filepath.Join(s.Dir(), TiesFolder, brief))
	if err != nil || !bytes.HasSuffix(raw, []byte("\n")) {
		return "", "", false
	}
	session, agent, cut := strings.Cut(strings.TrimSuffix(string(raw), "\n"), "/")
	if !cut || validName(session) != nil || validName(agent) != nil {
		return "", "", false
	}
	return session, agent, true
}

// Thread fills a thread's recipient (bonsai.messages/1's: the agent tied to its newest brief, its counts and limits as
// the guard recorded them, spent or not) and the two flags that read an agent's state: budget-spent, raised while that
// agent's window is spent; and unanswered, which for an open thread whose newest brief an agent tied reads that
// agent's last call, raised once it is the threshold's minutes old or older (an agent that stopped, or went quiet,
// makes no call: Bonsai has no clock of its own). A thread whose newest brief no agent tied keeps the recipient null
// and the folding's flags (unanswered by its brief's age alone). The flags stay in format.MessageFlags' order.
func (s Store) Thread(t *message.Thread, unansweredMinutes int64, now time.Time) {
	newest := ""
	for _, e := range t.Entries {
		if e.Msg.Kind == format.KindBrief {
			newest = e.Msg.ID
		}
	}
	if newest == "" {
		return
	}
	session, agent, ok := s.TiedTo(newest)
	if !ok {
		return
	}
	r, err := s.Recorded(session, agent)
	if err != nil {
		return
	}
	var minutes *json.Number
	if r.Minutes != nil {
		n := json.Number(strconv.FormatInt(*r.Minutes, 10))
		minutes = &n
	}
	lim := func(l Limit) format.AgentLimit { return format.AgentLimit{Value: l.Value, Max: l.Max} }
	t.Recipient = &format.ThreadRecipient{Agent: agent,
		Counts: &format.AgentCounts{Minutes: minutes, Actions: r.Actions, RedRuns: r.RedRuns},
		Limits: &format.AgentLimits{Minutes: lim(r.Limits.Minutes), Actions: lim(r.Limits.Actions), RedRuns: lim(r.Limits.RedRuns)},
		Spent:  r.Spent}
	raised := map[string]bool{}
	for _, f := range t.Flags {
		raised[f] = true
	}
	raised["budget-spent"] = r.Spent
	if t.State == message.Open && !r.LastCall.IsZero() {
		raised["unanswered"] = now.Sub(r.LastCall) >= time.Duration(unansweredMinutes)*time.Minute
	}
	t.Flags = []string{}
	for _, f := range format.MessageFlags {
		if raised[f.Word] {
			t.Flags = append(t.Flags, f.Word)
		}
	}
}
