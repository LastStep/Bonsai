package main

// bonsai init and update (plan part 3; spec §4, §6): the run they share, the consent rule and what they print. Their
// tables of flags and exit codes are initWord (init.go) and updateWord (update.go); with --json both print the changes
// output, bonsai.changes/1 (engine.Plan.Changes), its error object filled on a refusal. unlink (step 5.1.7) prints the
// same output: its word, in a file of its own, reuses changesRefused and the plan's Changes.
//
// The consent rule (spec §3): a command that writes prints its preview first. It writes with --yes; at a terminal
// without --yes it asks y/N; without a terminal, in an agent session (CLAUDE_CODE_CHILD_SESSION set) or with --json
// it never waits: it prints the preview and exits 4, naming the command with --yes.
//
// Code is consented to separately (spec §6; internal/engine/consent.go): a plan whose preview lists anything under
// "Runs code" needs --allow-exec as well as --yes. Without --allow-exec it prints the preview and exits 4, naming
// --allow-exec, at a terminal too: code is never a y/N question. With --allow-exec and no --yes it asks y/N at a
// terminal, and elsewhere exits 4 naming --yes. --allow-exec with nothing under "Runs code" changes nothing.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// interactive reports whether a person can answer a y/N question: stdin and stdout are terminals and this is no
// agent session. A variable, so tests stand in a terminal.
var interactive = func() bool {
	return os.Getenv("CLAUDE_CODE_CHILD_SESSION") == "" && isTerminal(os.Stdin) && isTerminal(os.Stdout)
}

// stdin is where a y/N answer is read from (a variable for tests).
var stdin io.Reader = os.Stdin

// pluginCLI is Claude Code's plugin commands: init and update install each locked pack's plugin on this machine,
// and check compares what is installed with the lock (internal/engine/plugins.go). main sets the real one; nil,
// as in tests, asks Claude Code nothing.
var pluginCLI engine.PluginCLI

// engineFlags are init's or update's flags, read from the word's table.
type engineFlags struct {
	yes, diff, newID, allowExec bool
	keep, adopt                 []string
	values                      engine.InitValues
	typed                       []string // the flags as typed that a next step repeats (not --yes, --diff or --json)
}

func readEngineFlags(c *call) *engineFlags {
	f := &engineFlags{yes: c.has("--yes"), diff: c.has("--diff"), newID: c.has("--new-id"), allowExec: c.has("--allow-exec"),
		keep: c.all("--keep"), adopt: c.all("--adopt")}
	f.values = engine.InitValues{Name: c.value("--name"), Source: c.value("--source"), Ref: c.value("--ref"), Path: c.value("--path"),
		NeverEdit: c.all("--never-edit")}
	exec := false
	for _, g := range c.given {
		switch g.name {
		case "--yes", "--diff", "--json":
		case "--new-id":
			f.typed = append(f.typed, g.name)
		case "--allow-exec":
			if !exec {
				f.typed = append(f.typed, g.name)
			}
			exec = true
		default:
			f.typed = append(f.typed, g.name, g.value)
		}
	}
	return f
}

// changesRefused is init's and update's --json for a refusal before a plan (bonsai.changes/1).
func changesRefused(c *call, e *engine.Error) encoder {
	return engine.RefusedChanges(c.word.Name, c.has("--allow-exec"), e)
}

// applyPlan writes a plan (a variable, so a test can stop it part-way).
var applyPlan = engine.Apply

// runEngine runs init or update: build the plan, print it, ask or refuse, write.
func runEngine(c *call) int {
	word := c.word.Name
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("%s takes no %+q", word, c.rest[0]))
	}
	f := readEngineFlags(c)
	cmdline := "bonsai " + word
	for _, t := range f.typed {
		cmdline += " " + engine.ShellArg(t)
	}
	dir, err := os.Getwd()
	if err != nil {
		return c.fail(&engine.Error{Code: "read-failed", Exit: engine.ExitRuntime, What: "the current folder cannot be read",
			Next: "run it from a folder inside the project"})
	}
	home, err := workspace.Home()
	if err != nil {
		return c.fail(engine.HomeError(err))
	}
	req := engine.Request{Command: word, Dir: dir, Home: home, Version: version, NewID: f.newID, Keep: f.keep, Adopt: f.adopt,
		AllowExec: f.allowExec}
	if word == "init" {
		req.Init = &f.values
	}
	plan, err := engine.Build(req)
	if err != nil {
		return c.fail(engine.Unexpected(err))
	}
	// out prints the plan's outcome: with --json the changes output (its error object filled from refusal), else
	// text on stdout.
	out := func(result string, exit int, refusal *engine.Error, text string) int {
		if refusal != nil {
			noted(c.word, refusal)
		}
		if c.json {
			return c.printDoc(plan.Changes(result, refusal), exit)
		}
		if write(c.stdout, text) != exitOK {
			return exitRuntime
		}
		return exit
	}
	closing := ""
	if word == "init" {
		closing = engine.ClosingWords(plan.Config, home)
	}
	// The command that installs a plugin carrying code on this machine, a person's step: once the files are written
	// (or have nothing to change) the same word with --allow-exec does it, without the values or the flags that
	// settled conflicts.
	again := "bonsai " + word + " --allow-exec"
	if plan.Nothing() {
		plan.Plugins = engine.InstallPlugins(plan.Root, plan.Config, plan.NewLock(), pluginCLI, plan.PluginConsent(again))
		return out("nothing", engine.ExitOK, nil, "bonsai "+word+": nothing to change"+packsAt(plan)+".\n"+plan.PluginsText()+closing)
	}
	head := "bonsai " + word + ": the preview.\n"
	preview := head + plan.Preview(f.diff)
	// Code is consented to separately (spec §6): without --allow-exec, a plan that runs code is refused whole, at a
	// terminal too. The next step repeats the command with --allow-exec, and settles any conflict in the same run.
	if plan.NeedsExec() && !f.allowExec {
		base := cmdline + " --allow-exec"
		next := "read each item under Runs code; to write them with the rest, run: " + base + " --yes"
		if len(plan.Conflicts) > 0 {
			next = "read each item under Runs code, and settle the conflicts: " + plan.ConflictNext(base)
		}
		n := len(plan.RunsCode)
		e := &engine.Error{Code: "needs-allow-exec", Exit: engine.ExitState,
			What: fmt.Sprintf("this %s writes code that runs on this machine (%d %s under Runs code), which needs --allow-exec as well as --yes: nothing was written",
				word, n, map[bool]string{true: "item", false: "items"}[n == 1]),
			Next: next}
		return out("refused", e.Exit, e, preview+"Refused: "+e.What+".\nnext: "+e.Next+"\n")
	}
	conflict := func(printed bool) int {
		e := &engine.Error{Code: "conflicts", Exit: engine.ExitConflict, What: conflictWhat(plan) + ": nothing was written",
			Next: plan.ConflictNext(cmdline)}
		text := "Stopped: " + e.What + ".\nnext: " + e.Next + "\n"
		if !printed {
			text = preview + text
		}
		return out("conflict", e.Exit, e, text)
	}
	if !f.yes {
		if c.json || !interactive() {
			next := "to write it, run: " + cmdline + " --yes"
			if len(plan.Conflicts) > 0 {
				next = plan.ConflictNext(cmdline)
			}
			e := &engine.Error{Code: "needs-yes", Exit: engine.ExitState, What: "no --yes, and no terminal to ask at: nothing was written", Next: next}
			return out("preview", e.Exit, e, preview+"Nothing written yet.\nnext: "+next+"\n")
		}
		// At a terminal (never with --json): the person answers.
		if write(c.stdout, preview+"Write these changes? [y/N] ") != exitOK {
			return exitRuntime
		}
		if !yesAnswer() {
			return out("declined", engine.ExitState, nil, "Nothing written.\nnext: run the command again when you want it written\n")
		}
		if len(plan.Conflicts) > 0 {
			return conflict(true)
		}
	} else if len(plan.Conflicts) > 0 {
		return conflict(false)
	}
	// Apply refuses before any rename (refused, nothing written), or stops part-way through the renames (failed: some
	// files written, the lock not; the same command again finishes the rest).
	if err := applyPlan(plan); err != nil {
		e := engine.Unexpected(err)
		result := "refused"
		if e.Code == "partly-written" {
			result = "failed"
		}
		if c.json {
			return out(result, e.Exit, e, "")
		}
		return c.fail(e)
	}
	plan.Plugins = engine.InstallPlugins(plan.Root, plan.Config, plan.NewLock(), pluginCLI, plan.PluginConsent(again))
	text := ""
	if f.yes {
		text = preview
	}
	text += "bonsai " + word + ": written" + packsAt(plan) + ".\n" + plan.Applied() + plan.PluginsText() + closing
	return out("applied", engine.ExitOK, nil, text)
}

func conflictWhat(p *engine.Plan) string {
	var names []string
	for _, c := range p.Conflicts {
		names = append(names, c.Path)
	}
	return fmt.Sprintf("%d %s (%s)", len(names), map[bool]string{true: "conflict", false: "conflicts"}[len(names) == 1],
		strings.Join(names, ", "))
}

func packsAt(p *engine.Plan) string {
	var parts []string
	for _, m := range p.Packs {
		parts = append(parts, m.ID+" at "+m.To[:7])
	}
	if len(parts) == 0 {
		return ""
	}
	return "; " + strings.Join(parts, ", ")
}

func yesAnswer() bool {
	line, _ := bufio.NewReader(stdin).ReadString('\n')
	a := strings.ToLower(strings.TrimSpace(line))
	return a == "y" || a == "yes"
}
