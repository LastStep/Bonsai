package ladder

// The human lines (note 9), ASCII (spec §3: PowerShell 5.1 garbles UTF-8): a header, one line per rung as it ends with
// its status, its time and its ratchet, the red rungs' reasons and failing tests, the leftovers line, the result's path
// and the next step. --json prints none of them: the climb document says the same.

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// windows reports a climb whose rungs run in cmd.
func windows() bool { return runtime.GOOS == "windows" }

// say prints one line, made ASCII.
func (c *Climb) say(format string, a ...any) {
	_, _ = fmt.Fprintln(c.o.Out, engine.ASCII(fmt.Sprintf(format, a...)))
}

// printNote prints a note: what the climb could not do that changes no colour.
func (c *Climb) printNote(s string) { c.say("  %s", s) }

// printHeader is the climb's first lines: the project, the task, the mode, the rungs and the commit.
func (c *Climb) printHeader() {
	task := "no task"
	if c.Task != nil {
		task = c.Task.ID
	}
	var rungs []string
	for _, n := range c.Requested {
		rungs = append(rungs, strconv.FormatInt(n, 10))
	}
	c.say("bonsai ladder: %s, %s (%s), rungs %s", c.Config.Name, task, c.mode(), strings.Join(rungs, ", "))
	dirty := ""
	if c.git.Dirty {
		dirty = " (dirty)"
	}
	c.say("  %s at %s %s%s", shown(c.Root), c.git.Branch, shortSHA(c.git.SHA), dirty)
	if c.unverified != "" {
		c.say("  this checkout's main checkout is not verified (%s): the ci result is written in this checkout", c.unverified)
	}
	if c.priorNote != "" {
		c.say("  note: %s", c.priorNote)
	}
	if c.checksNote != "" {
		c.say("  note: %s", c.checksNote)
	}
}

// printRung is one rung's line as it ends, with the job's notes under it.
func (c *Climb) printRung(r format.Rung, e *format.LadderRung, notes []string) {
	dur := ""
	if e.DurationMS != nil {
		dur = fmt.Sprintf("%.1fs", float64(*e.DurationMS)/1000)
	}
	rat := ""
	if e.Ratchet != nil {
		rat = "  " + e.Ratchet.Name + " " + string(e.Ratchet.Was) + "->" + string(e.Ratchet.Now)
	}
	c.say("  %2d %-30s %-7s %7s%s", e.Rung, e.Name, e.Status, dur, rat)
	for _, n := range notes {
		for _, line := range strings.Split(strings.TrimSpace(n), "\n") {
			c.say("       note: %s", line)
		}
	}
}

// printSummary is the climb's last lines: green or red, each red rung's reason and failing tests, the leftovers line,
// the result and the next step.
func (c *Climb) printSummary(l *format.Ladder, w *Written, d *format.Climb) {
	n := 0
	for _, r := range l.Rungs {
		if r.Status == statusGreen {
			n++
		}
	}
	colour := "GREEN"
	if !l.Green {
		colour = "RED"
	}
	ci := ""
	if c.CI {
		ci = "; ci: never a task's proof"
	}
	dirty := ""
	if l.Git.Dirty {
		dirty = "; the tree was dirty (a dirty climb proves no commit)"
	}
	c.say("")
	c.say("%s: %d of %d rungs green, %.1fs%s%s", colour, n, len(l.Rungs), c.Now().Sub(c.started).Seconds(), ci, dirty)
	if c.headAfter != c.git.SHA {
		c.say("  HEAD moved during the climb (%s to %s): the result counts the tree as dirty", shortSHA(c.git.SHA), shortSHA(c.headAfter))
	}
	if len(c.hides) > 0 {
		c.say("  git integrity found a change git hides (rung %s): the result counts the tree as dirty", strings.Join(c.hides, ", "))
	}
	if c.configMoved {
		c.say("  the main checkout's bonsai.yaml changed during the climb: the result counts the tree as dirty")
	}
	if c.treeErr != "" {
		c.say("  the tree could not be read, so the result counts it as dirty: %s", c.treeErr)
	}
	if len(c.hidden) > 0 {
		c.say("  %d changed %s git status does not show, so the result counts the tree as dirty (each read by its contents "+
			"against HEAD):", len(c.hidden), plural(len(c.hidden), "path", "paths"))
		for i, h := range c.hidden {
			if i == FailedPrinted {
				c.say("    ... and %d more", len(c.hidden)-FailedPrinted)
				break
			}
			c.say("    %s: %s", strconv.QuoteToASCII(h.Path), h.Why)
		}
		c.say("  git keeps a file it saw unchanged in its index; to make git status show each again: git rm -q --cached -- <path> " +
			"&& git add -- <path> (a submodule's file: inside the submodule), then commit or revert it")
	}
	for _, r := range l.Rungs {
		if r.Status == statusGreen || r.Status == statusSkipped {
			continue
		}
		c.say("")
		c.say("  rung %d %s: %s", r.Rung, r.Name, r.Status)
		if r.Reason != nil {
			for _, line := range strings.Split(*r.Reason, "\n") {
				c.say("    %s", line)
			}
		}
		c.printFailed(r.Tests)
	}
	if c.CI && len(l.Skipped) > 0 {
		c.say("")
		c.say("  skipped under --ci (%d):", len(l.Skipped))
		for _, s := range l.Skipped {
			ok := "NOT ACCEPTED"
			if s.Accepted {
				ok = "accepted"
			}
			c.say("    rung %d  %s  [%s]  %s", s.Rung, s.Name, s.Reason, ok)
		}
	}
	if len(c.leftLines) > 0 {
		c.say("")
		for _, line := range c.leftLines {
			c.say("  %s", line)
		}
	}
	c.say("")
	c.say("  result: %s (sha256 %s)", w.Rel, w.SHA256)
	for _, p := range d.Cleaned {
		c.say("  cleaned: %s", p)
	}
	if d.Next != nil {
		c.say("  next: %s", d.Next.Do)
	}
}

// printFailed prints a rung's failing tests, at most FailedPrinted, each with its file, its error and its command.
func (c *Climb) printFailed(tests schema.Object) {
	fv, ok := tests.Get("failed")
	list, isList := fv.([]any)
	if !ok || !isList || len(list) == 0 {
		return
	}
	total, _ := tests.Get("failed_count")
	c.say("    failing tests (%v):", total)
	for i, f := range list {
		if i == FailedPrinted {
			c.say("    ... and more (the first %d are in the result)", FailedKept)
			break
		}
		o := f.(schema.Object)
		line := "    x " + o.String("name")
		if file := o.String("file"); file != "" {
			line += "  [" + file + "]"
		}
		c.say("%s", line)
		if e := o.String("error"); e != "" {
			c.say("        %s", e)
		}
		if run := o.String("run"); run != "" {
			c.say("        run: %s", run)
		}
	}
}

// shortSHA is a commit's first 8 characters.
func shortSHA(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	if s == "" {
		return "(no commit)"
	}
	return s
}
