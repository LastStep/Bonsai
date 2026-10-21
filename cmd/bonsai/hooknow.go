package main

// bonsai hook now: the time line (plan P-T-5795 step 5, piece E; contract section 23). Claude Code runs it, synchronous,
// on each prompt and at each subagent's start, and the model reads its one line as context, so an agent records times
// from the clock, never from memory or a guess. Its line never blocks (engine.OpenHookCommand), and the word itself
// exits 0 whatever happens.

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/timefmt"
)

var hookNowWord = &Word{
	Name:    "hook now",
	Title:   "the time line: the current time told to the model on each prompt and each subagent's start",
	Summary: "the time line's hook",
	About: lineWords + engine.ShownLine(engine.NowCommand) + `on UserPromptSubmit and on SubagentStart (no matcher), timeout 5 s, synchronous, each.
It prints one line, the time in the form Bonsai records (UTC) with the person's local time beside it:
  Now: 2026-10-21 08:19 UTC (local 13:49 +05:30)
with the local date too when the local day is not the UTC day:
  Now: 2026-10-21 23:50 UTC (local 2026-10-22 05:20 +05:30)
On a prompt the line is plain stdout, which Claude Code adds to the model's context. At a subagent's start (read
from the payload's hook_event_name) it is the same line as one JSON object on stdout,
{"hookSpecificOutput":{"hookEventName":"SubagentStart","additionalContext":"Now: ..."}}, the one way that event
gives a subagent context. Any other payload, or none, gets the plain line.
It never blocks: it exits 0 whatever it is handed, and its line ends || exit 0, so a bonsai that predates this
word (which answers exit 2), a fault, a timeout or a missing bonsai lets the prompt run with no line. It reads
no project and records nothing (the recorder's lines record both events). A long claude -p run with no new
prompt gets no new line: run bonsai now there.
`,
	FlagNext: "use the hook line bonsai init writes (bonsai hook now --help quotes it)",
	Exits: []Exit{
		{Code: 0, Means: "always: the line is printed, and the prompt or the subagent goes on"},
	},
	Examples: []string{"bonsai hook now < payload.json"},
	Run: func(c *call) int {
		return hookNow(os.Stdin, isTerminal(os.Stdin), c.stdout, nowClock(), nowZone())
	},
}

// hookNow writes the time line for the payload on stdin (terminal: none is waited for) to stdout, and returns 0
// always: a panic is caught, and then nothing more is printed.
func hookNow(stdin io.Reader, terminal bool, stdout io.Writer, t time.Time, loc *time.Location) (code int) {
	defer func() {
		if recover() != nil {
			code = 0
		}
	}()
	line := nowLine(t, loc)
	p, _ := hookin.Read(stdin, terminal)
	if p != nil && p.Event == engine.NowSubagentEvent {
		out := struct {
			Hook struct {
				Event   string `json:"hookEventName"`
				Context string `json:"additionalContext"`
			} `json:"hookSpecificOutput"`
		}{}
		out.Hook.Event, out.Hook.Context = engine.NowSubagentEvent, line
		b, err := json.Marshal(out)
		if err != nil {
			return 0
		}
		_, _ = stdout.Write(append(b, '\n'))
		return 0
	}
	_, _ = io.WriteString(stdout, line+"\n")
	return 0
}

// nowLine is the time line: the record form in UTC, then the local time with its offset, and the local date only
// when the local day differs from the UTC day.
func nowLine(t time.Time, loc *time.Location) string {
	local := timefmt.Doc(t, loc) // 2006-01-02 15:04 -07:00
	if timefmt.Day(t, loc) == timefmt.RecordDay(t) {
		local = local[len("2006-01-02 "):]
	}
	return "Now: " + timefmt.Record(t) + " UTC (local " + local + ")"
}
