package main

// bonsai messages (contract §22; design/plan-5.md, step 5.5, "Messages between agents", note 5.5.10, 4): the threads
// of messages between agents in the main checkout's .bonsai/local/messages/, headers only and never bodies, each with
// its state (open or answered) and its flags, and every file that does not read. It reads and writes nothing else,
// and reports, never acts. With --json it prints bonsai.messages/1 (format.Messages). Its table is messagesWord.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/redact"
)

func init() { register(messagesWord) }

var messagesWord = &Word{
	Name:    "messages",
	Order:   14,
	Title:   "the threads of messages between agents (contract section 22).",
	Summary: "list the threads of messages between agents",
	Args:    "[flags]",
	About: `It reads the main checkout's .bonsai/local/messages/ (a worktree reads its main checkout's) and writes nothing.
Each thread is one line, oldest first: its state (open while its newest brief has no report after it, else
answered), its id (its first brief's), its caller and recipient, its task, its messages and its flags. Headers
only: bonsai message --show prints a message's words. A file that does not read is listed as malformed and
skipped; it never blocks the rest. It reports, and never acts on an agent.
`,
	Flags: []Flag{
		{Name: "--to", Value: "ADDRESS", Help: "only threads holding a message to this address"},
		{Name: "--task", Value: "ID", Help: "only threads holding a message serving this task"},
		{Name: "--open", Help: "only open threads: a brief with no report after it"},
		{Name: "--json", Help: "print the bonsai.messages/1 document (for programs) instead of plain text"},
	},
	Notes: "Its flags: " + messageFlagWords() + " (bonsai check --schema bonsai.messages\n" +
		"lists them). unanswered: an open thread's newest brief as old as messages.unanswered_minutes in bonsai.yaml;\n" +
		"circling: more briefs one way between two addresses than messages.circling_briefs.\n",
	Exits: []Exit{
		{Code: 0, Means: "the threads were read (none prints a line saying so)"},
		{Code: 2, Means: "bad input (a flag messages does not take, an address or task id of the wrong form); bonsai.yaml\nrefused"},
		{Code: 3, Means: "the messages cannot be read, or the answer cannot be printed"},
		{Code: 4, Means: "not in a git checkout, or not linked (bonsai init)"},
	},
	Examples: []string{"bonsai messages --open", "bonsai messages --to builder@T-0901 --json"},
	Refused:  messagesRefused,
	Run:      runMessages,
}

// messageFlagWords are the known flags, in their table's order.
func messageFlagWords() string {
	var w []string
	for _, f := range format.MessageFlags {
		w = append(w, f.Word)
	}
	return strings.Join(w, ", ")
}

func runMessages(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("messages takes no %+q", redact.Text(c.rest[0])))
	}
	f := message.Filter{To: c.value("--to"), Task: c.value("--task"), Open: c.has("--open")}
	if why := message.CheckAddress(f.To); c.has("--to") && why != "" {
		return c.refuse(&engine.Error{Code: "bad-value", Exit: exitInput, What: "--to " + strconv.QuoteToASCII(f.To) + " " + why,
			Next: "give --to as <role>, <role>@<task> or <role>@<task>#<n>, such as builder@T-0901"})
	}
	if c.has("--task") && !message.IsTask(f.Task) {
		return c.refuse(&engine.Error{Code: "bad-value", Exit: exitInput, What: "--task " + strconv.QuoteToASCII(f.Task) + " is not a task id (T- and 4 to 6 digits)",
			Next: "give --task as T-0901, or leave it out"})
	}
	site, e := askSite()
	if e != nil {
		return c.fail(e)
	}
	p := messagePlace(site)
	st, e := readMessages(p)
	if e != nil {
		return c.fail(e)
	}
	rules, e := messageRules(site, false)
	if e != nil {
		return c.fail(e)
	}
	var kept []message.Thread
	for _, t := range message.Fold(st, rules, messageNow()) {
		if f.Keep(t) {
			kept = append(kept, t)
		}
	}
	if c.json {
		doc := &format.Messages{Workspace: site.ref, Threads: []format.MessageThread{}, Malformed: st.Malformed}
		for _, t := range kept {
			doc.Threads = append(doc.Threads, t.MessageThread)
		}
		return c.printDoc(doc, exitOK)
	}
	var b strings.Builder
	for _, t := range kept {
		task := "-"
		if t.Task != nil {
			task = *t.Task
		}
		flags := ""
		if len(t.Flags) > 0 {
			flags = "  flags: " + strings.Join(t.Flags, ", ")
		}
		fmt.Fprintf(&b, "%-8s  %s  %s -> %s  %s  %d messages, last %s%s\n", t.State, t.ID, t.From, t.To, task,
			len(t.Messages), t.LastAt, flags)
	}
	if len(kept) == 0 {
		which := "threads"
		if f.Open {
			which = "open threads"
		}
		if f.To != "" || f.Task != "" {
			which += " that the filters keep"
		}
		b.WriteString("no " + which + " in " + message.Dir + "/\n")
	}
	for _, m := range st.Malformed {
		fmt.Fprintf(&b, "malformed  %s: %s\n", m.Path, m.Reason)
	}
	return c.print(engine.ASCII(b.String()))
}
