package main

// bonsai agents (contract §12's note of 14 Oct; design/plan-5.md, step 5.5, note 5.5.12; Rohan, 14 Oct: "bonsai agent
// status is just a quick way to poll data created by bonsai itself"): every agent Bonsai has recorded in the project,
// the person's sessions and every subagent the guard kept state for, each copied as recorded from the guard's state
// under .bonsai/local/agents/, the log and the messages (internal/agents' View). It reads only, never acts (it never
// starts, wakes or stops an agent), derives nothing (no state such as working or quiet, no share of a budget, no
// estimate, no minutes worked out against the clock) and reads no setting, so the limits it shows are the ones that
// bound each agent. With --json it prints bonsai.agents/1 (format.Agents). Its table is agentsWord.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/redact"
)

func init() { register(agentsWord) }

var agentsWord = &Word{
	Name:    "agents",
	Order:   15,
	Title:   "every agent Bonsai has recorded in this project, as recorded (bonsai.agents/1).",
	Summary: "list every agent Bonsai has recorded, as recorded",
	Args:    "[flags]",
	About: `It reads the main checkout's .bonsai/local/ (a worktree reads its main checkout's) and writes nothing: the
guard's state of each agent (agents/), the log (log/) and the messages (messages/). One line an agent, oldest
first: its id, its session and its parent's id (session for an agent the person's own session started; the
tree is drawn from parent alone), its role and type, its task, tied brief and thread, its model, when it
started and ended, its last call, its counts and limits as the guard wrote them, spent where the guard
recorded it, its report (only one from its own address), its log file and its token counts (usage) from
its newest end. The person's own sessions are listed from the log, with no counts or limits. Nothing is
worked out: no state such as working or quiet, no share of a budget, no estimate; the limits are the ones
that bound each agent, never today's settings. A state file that does not read is listed as skipped and
never blocks the rest. What is recorded under an agent's name is attributed, not authenticated. It never
starts, wakes or stops an agent.
`,
	Flags: []Flag{
		{Name: "--json", Help: "print the bonsai.agents/1 document (for programs) instead of plain text"},
	},
	Exits: []Exit{
		{Code: 0, Means: "the agents were read (none prints a line saying so; a state file that does not read is listed as skipped)"},
		{Code: 2, Means: "bad input (a flag agents does not take); bonsai.yaml refused"},
		{Code: 3, Means: "the agents' folder, the log or the messages cannot be read, or the answer cannot be printed"},
		{Code: 4, Means: "not in a git checkout, or not linked (bonsai init)"},
	},
	Examples: []string{"bonsai agents", "bonsai agents --json"},
	Refused:  agentsRefused,
	Run:      runAgents,
}

func agentsRefused(c *call, e *engine.Error) encoder {
	return &format.Agents{Agents: []format.Agent{}, Skipped: []format.AgentSkipped{}, Error: e.Object()}
}

func runAgents(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("agents takes no %+q", redact.Text(c.rest[0])))
	}
	site, e := askSite()
	if e != nil {
		return c.fail(e)
	}
	main := site.place.Local.Main
	msgs, err := message.Read(main)
	if err != nil {
		return c.fail(&engine.Error{Code: "read-failed", Exit: exitRuntime, What: message.Dir + " cannot be read: " + engine.ASCII(err.Error()),
			Next: "check that " + message.Dir + "/ can be read, then run the command again"})
	}
	v, err := agents.Store{Main: main}.View(msgs)
	if err != nil {
		return c.fail(&engine.Error{Code: "read-failed", Exit: exitRuntime,
			What: "the agents' state or the log under .bonsai/local/ cannot be read: " + engine.ASCII(err.Error()),
			Next: "check that .bonsai/local/agents/ and .bonsai/local/log/ can be read, then run the command again"})
	}
	if c.json {
		return c.printDoc(&format.Agents{Workspace: site.ref, Agents: v.Agents, Skipped: v.Skipped}, exitOK)
	}
	var b strings.Builder
	for i := range v.Agents {
		b.WriteString(agentLine(&v.Agents[i]))
		b.WriteByte('\n')
	}
	if len(v.Agents) == 0 {
		b.WriteString("no agents recorded in .bonsai/local/agents/ or .bonsai/local/log/\n")
	}
	for _, s := range v.Skipped {
		fmt.Fprintf(&b, "skipped  %s: %s\n", s.Path, s.Reason)
	}
	return c.print(engine.ASCII(b.String()))
}

// agentLine is one agent's line: every field of its entry, in bonsai.agents/1's order, "-" for null, as recorded.
func agentLine(a *format.Agent) string {
	or := func(p *string) string {
		if p == nil {
			return "-"
		}
		return *p
	}
	num := func(p *int64) string {
		if p == nil {
			return "-"
		}
		return strconv.FormatInt(*p, 10)
	}
	f := []string{a.ID, "session " + a.Session, "parent " + or(a.Parent), "role " + or(a.Role), "type " + or(a.Type),
		"task " + or(a.Task), "brief " + or(a.Brief), "model " + or(a.Model), "started " + or(a.Started), "ended " + or(a.Ended)}
	if c := a.LastCall; c != nil {
		f = append(f, "last_call "+c.At+" "+c.Tool+" "+or(c.Category))
	} else {
		f = append(f, "last_call -")
	}
	if c := a.Counts; c != nil {
		minutes := "-"
		if c.Minutes != nil {
			minutes = c.Minutes.String()
		}
		f = append(f, fmt.Sprintf("counts minutes %s actions %d red_runs %d", minutes, c.Actions, c.RedRuns))
	} else {
		f = append(f, "counts -")
	}
	if l := a.Limits; l != nil {
		lim := func(x format.AgentLimit) string { return num(x.Value) + " (max " + num(x.Max) + ")" }
		f = append(f, "limits minutes "+lim(l.Minutes)+" actions "+lim(l.Actions)+" red_runs "+lim(l.RedRuns))
	} else {
		f = append(f, "limits -")
	}
	f = append(f, "spent "+strconv.FormatBool(a.Spent))
	if r := a.Report; r != nil {
		f = append(f, "report "+r.ID+" "+r.Status)
	} else {
		f = append(f, "report -")
	}
	f = append(f, "log "+or(a.Log), "thread "+or(a.Thread))
	switch {
	case a.Usage == nil:
		f = append(f, "usage -")
	default:
		var items []string
		for _, u := range a.Usage {
			if u.Kind != nil {
				items = append(items, *u.Kind+" ("+or(u.Reason)+")")
				continue
			}
			items = append(items, fmt.Sprintf("%s input %s output %s cache_read %s cache_write %s", or(u.Model), num(u.Input),
				num(u.Output), num(u.CacheRead), num(u.CacheWrite)))
		}
		f = append(f, "usage "+strings.Join(items, "; "))
	}
	return strings.Join(f, "  ")
}
