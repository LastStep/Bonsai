package message

// The threads (contract §22; design/plan-5.md, "Messages between agents", note 5.5.10, 4): folded from the files each
// time they are read, since every message is a file of its own and no shared file is ever rewritten (the research's
// lesson from mailboxes that broke under many writers). A thread is every message carrying one thread id, its first
// brief's.
//
//   - Its caller and recipient (from, to) and its task are its first brief's: the message whose id is the thread's,
//     or where that does not read, its oldest brief; a thread holding only reports takes its oldest report's, turned
//     round (a report goes back to its caller).
//   - Its state: open while its newest brief has no report after it (by created, then id, or replying to it: a report
//     written in the brief's own millisecond still answers it); answered once a report comes after it, its caller's own
//     blocked report closing a dropped thread among them. A thread with no brief left is answered.
//   - Its flags, in format.MessageFlags' order, reported and never acted on:
//     unanswered, an open thread whose newest brief is the threshold's minutes old or older, compared without overflow
//     however large the threshold (AtLeastMinutes); where an agent tied that brief, internal/agents' Store.Thread reads
//     that agent's last call in its place, from the guard's records;
//     circling, more briefs one way between the same two addresses (from one to the other) than the threshold, so
//     the threshold is how many each way a thread may hold before it is flagged: at the default 3, the fourth brief
//     one way flags it;
//     budget-spent, never raised by the folding: internal/agents' Store.Thread raises it while the agent tied to the
//     newest brief is past its budget, as the guard recorded it;
//     malformed, a file naming the thread does not read (it is listed apart, and check warns of it).
//   - Its messages, headers only (each frontmatter, never a body), oldest first; its recipient null from the folding,
//     filled by internal/agents' Store.Thread from the guard's records of the agent tied to the newest brief. Every
//     reader that prints a thread (bonsai messages, bonsai message --json, status's messages) calls both.
//
// Threads are listed oldest first, by their first brief's created, then the thread's id.

import (
	"math"
	"sort"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
)

// The states of a thread (bonsai.messages/1's closed list).
const (
	Open     = "open"
	Answered = "answered"
)

// Thread is one thread as folded: bonsai.messages/1's view of it, and the files it rests on.
type Thread struct {
	format.MessageThread
	Entries   []Entry  // its messages that read, oldest first
	Malformed []string // the files naming it that do not read, by path
}

// Fold gives every thread of a store, oldest first, with its state and its flags as of now under the rules.
func Fold(st *Store, r Rules, now time.Time) []Thread {
	byID := map[string]*Thread{}
	var ids []string
	for _, e := range st.Messages {
		t, ok := byID[e.Msg.Thread]
		if !ok {
			t = &Thread{}
			byID[e.Msg.Thread] = t
			ids = append(ids, e.Msg.Thread)
		}
		t.Entries = append(t.Entries, e)
	}
	for _, m := range st.Malformed {
		if m.Thread == nil {
			continue
		}
		if t, ok := byID[*m.Thread]; ok {
			t.Malformed = append(t.Malformed, m.Path)
		}
	}
	threads := make([]Thread, 0, len(ids))
	heads := map[string]string{} // a thread's id to its first brief's created
	for _, id := range ids {
		t := byID[id]
		head := t.fold(id, r, now)
		heads[id] = head
		threads = append(threads, *t)
	}
	sort.Slice(threads, func(i, j int) bool {
		a, b := threads[i].ID, threads[j].ID
		if heads[a] != heads[b] {
			return heads[a] < heads[b]
		}
		return a < b
	})
	return threads
}

// fold fills the thread's view from its entries (already oldest first) and gives its first brief's created.
func (t *Thread) fold(id string, r Rules, now time.Time) string {
	var head *format.Message
	for _, e := range t.Entries {
		if e.Msg.ID == id && e.Msg.Kind == format.KindBrief {
			head = e.Msg
			break
		}
	}
	for _, e := range t.Entries {
		if head == nil && e.Msg.Kind == format.KindBrief {
			head = e.Msg
		}
	}
	from, to := "", ""
	if head != nil {
		from, to = head.From, head.To
	} else {
		head = t.Entries[0].Msg
		from, to = head.To, head.From
	}
	t.ID, t.Task, t.From, t.To = id, head.Task, from, to
	t.LastAt = t.Entries[len(t.Entries)-1].Msg.Created
	t.Messages = make([]schema.Object, 0, len(t.Entries))
	for _, e := range t.Entries {
		t.Messages = append(t.Messages, Header(e.Msg))
	}
	t.State, t.Flags = Answered, []string{}
	newest := -1
	for i, e := range t.Entries {
		if e.Msg.Kind == format.KindBrief {
			newest = i
		}
	}
	if newest >= 0 {
		t.State = Open
		for i, e := range t.Entries {
			if e.Msg.Kind == format.KindReport && (i > newest || (e.Msg.ReplyTo != nil && *e.Msg.ReplyTo == t.Entries[newest].Msg.ID)) {
				t.State = Answered
				break
			}
		}
	}
	raised := map[string]bool{}
	if t.State == Open {
		at, err := time.Parse(record.AtLayout, t.Entries[newest].Msg.Created)
		raised["unanswered"] = err == nil && AtLeastMinutes(now.Sub(at), r.UnansweredMinutes)
	}
	briefs := map[[2]string]int64{}
	for _, e := range t.Entries {
		if e.Msg.Kind == format.KindBrief {
			k := [2]string{e.Msg.From, e.Msg.To}
			briefs[k]++
			if briefs[k] > r.CirclingBriefs {
				raised["circling"] = true
			}
		}
	}
	raised["malformed"] = len(t.Malformed) > 0
	for _, f := range format.MessageFlags {
		if raised[f.Word] {
			t.Flags = append(t.Flags, f.Word)
		}
	}
	return head.Created
}

// AtLeastMinutes reports whether d is n minutes or more, compared without overflow: time.Duration(n)*time.Minute
// wraps past about 292 years (math.MaxInt64 nanoseconds), so a threshold beyond any duration (bonsai.yaml may set up
// to 15 digits) is never reached, never read as a negative one.
func AtLeastMinutes(d time.Duration, n int64) bool {
	const most = math.MaxInt64 / int64(time.Minute)
	switch {
	case n > most:
		return false
	case n < -most:
		return true
	}
	return d >= time.Duration(n)*time.Minute
}

// Header gives a message's frontmatter as bonsai messages lists it: bonsai.message/1's fields from format to labels,
// then any field this Bonsai does not know, never the body or the trust line.
func Header(msg *format.Message) schema.Object {
	doc, err := msg.Object()
	if err != nil { // a message that read is held to its schema when read (Parse), so this is never reached
		doc = schema.Object{{Key: "format", Value: format.MustLookup("message").Versioned()}, {Key: "id", Value: msg.ID}}
	}
	out := schema.Object{}
	for _, m := range doc {
		if m.Key != "body" && m.Key != "trust" {
			out = append(out, m)
		}
	}
	return out
}

// Filter is what bonsai messages keeps: a thread holding a message to an address, one with a message serving a task,
// an open thread; each "" or false keeps every thread.
type Filter struct {
	To   string
	Task string
	Open bool
}

// Keep reports whether the filter keeps the thread.
func (f Filter) Keep(t Thread) bool {
	if f.Open && t.State != Open {
		return false
	}
	to, task := f.To == "", f.Task == ""
	for _, e := range t.Entries {
		to = to || e.Msg.To == f.To
		task = task || (e.Msg.Task != nil && *e.Msg.Task == f.Task)
	}
	return to && task
}
