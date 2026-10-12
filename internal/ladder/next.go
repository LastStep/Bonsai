package ladder

// The next steps (note 12, agents first): a refusal's, each a command an agent runs as written with every value Bonsai
// knows filled in (the task's id, the rung, the path, the holder's process), or where the fix is an edit, the file and
// the change and then the command that confirms it; and a red climb's, the first red rung's fix and then the climb
// again. Who is agent but where a safety reason makes the step a person's (format.ErrorWords says which words).

import (
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/testout"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// noTask is the refusal of a local climb with no task (note 1): none named, or the named task is none. named says a
// task was named (--task, or BONSAI_TASK counting in this project).
func noTask(c *Climb, a workspace.ActiveTask, named bool) *engine.Error {
	again := " --json"
	if c.CI {
		again = " --ci --json"
	}
	taskDir := c.Config.Full.Documents.Task
	if !named {
		what := "a local climb proves one task, and none was named (--task, or " + workspace.TaskEnv + " in this project's session), " +
			"so nothing was run"
		if a.How == workspace.HowRunning {
			return &engine.Error{Code: "no-task", Exit: ExitInput, What: what + "; " + a.ID + " is the one task reading running",
				Next: "run: bonsai ladder --task " + a.ID + again}
		}
		return &engine.Error{Code: "no-task", Exit: ExitInput, What: what + " (" + a.Sentence() + ")",
			Next: "find the task: run: bonsai status --active --json; then, with its id, run: bonsai ladder --task <id>" + again}
	}
	id := strings.TrimSpace(strings.Split(a.Detail, " ")[0])
	if c.o.Task != "" {
		id = c.o.Task
	}
	what := "the task named is none: " + a.Sentence() + " in the main checkout's task folder " + taskDir + ", so nothing was run"
	switch a.Why {
	case workspace.ReasonNamedDiffer:
		return &engine.Error{Code: "no-task", Exit: ExitInput, What: what,
			Next: "name one task: where " + workspace.TaskEnv + " is unset or names " + id + ", run: bonsai ladder --task " + id + again}
	case workspace.ReasonNamedDuplicate:
		return &engine.Error{Code: "no-task", Exit: ExitInput, What: what,
			Next: "give each of those files its own id in the main checkout (one keeps " + id + "); then run: bonsai ladder --task " + id + again}
	case workspace.ReasonNamedBroken:
		return &engine.Error{Code: "no-task", Exit: ExitInput, What: what,
			Next: "find the line: run: bonsai check --json; fix the task file in the main checkout so it parses; then run: bonsai ladder " +
				"--task " + id + again}
	}
	return &engine.Error{Code: "no-task", Exit: ExitInput, What: what,
		Next: "check the id: its file belongs in the main checkout's " + taskDir + ", which this reads too: run: bonsai status --active " +
			"--json; then, with the task's id, run: bonsai ladder --task <id>" + again}
}

// next is a red climb's next step (note 12): the first red rung's fix, then the climb again.
func (c *Climb) next() *format.Next {
	again := "then run: bonsai ladder " + c.taskFlag() + " --json"
	for _, s := range c.steps {
		e := s.Entry
		if e.Status == statusGreen || e.Status == statusSkipped {
			continue
		}
		label := "rung " + strconv.FormatInt(e.Rung, 10)
		if e.Name != "" {
			label = "the " + e.Name + " rung (rung " + strconv.FormatInt(e.Rung, 10) + ")"
		}
		who := "agent"
		if s.who == "person" {
			who = "person"
		}
		return &format.Next{Do: "fix " + label + ": " + c.fix(s) + " ; " + again, Who: who}
	}
	return &format.Next{Do: "run: bonsai ladder " + c.taskFlag() + " --json", Who: "agent"}
}

// fix is one red rung's fix, as next names it.
func (c *Climb) fix(s *Step) string {
	if s.next != "" {
		return s.next
	}
	if s.Tests != nil {
		for _, b := range s.Tests.Broken {
			if b.What != "" && b.Build && s.Tests.Form == format.TestsGoJSON {
				return "the package " + b.What + " does not build (the rung's reason shows the error); run: go vet " + b.What
			}
		}
		n := failedCount(*s.Tests)
		in := testout.RunInput{Root: c.Root, Windows: windows()}
		if s.Rung.Command != nil {
			in.Command = *s.Rung.Command
		}
		for _, t := range s.Tests.Tests {
			if t.Outcome != testout.Failed || t.Parent {
				continue
			}
			first := "the test " + t.Name + " fails"
			if n > 1 {
				first = strconv.Itoa(n) + " tests fail (tests.failed names each with its command); the first, " + t.Name + ","
				if run := testout.Run(s.Tests.Form, t, in); run != "" {
					return first + " runs alone with: run: " + run
				}
				return first + " is in the rung's tests"
			}
			if run := testout.Run(s.Tests.Form, t, in); run != "" {
				return first + "; it runs alone with: run: " + run
			}
			return first
		}
		if s.Entry.Reason != nil && strings.Contains(*s.Entry.Reason, "this climb cannot accept") {
			return "its reason names the skipped or todo tests this climb cannot accept (a local climb runs every marked test); make " +
				"each run, or take its skip out"
		}
	}
	if s.timed {
		return "it ran past its limit of " + strconv.FormatInt(int64(rungTimeout(s.Rung).Seconds()), 10) + " s; run its command to see " +
			"where it stops: run: " + commandOf(s.Rung) + " (a longer limit is bonsai.yaml's timeout_s, a person's change)"
	}
	if s.Rung.Kind == KindCommand && commandOf(s.Rung) != "" {
		return "its reason shows what failed; run: " + commandOf(s.Rung)
	}
	return "its reason in the result names what to change"
}

// commandOf is a command rung's command, "" for none.
func commandOf(r format.Rung) string {
	if r.Command == nil {
		return ""
	}
	return strings.TrimSpace(*r.Command)
}
