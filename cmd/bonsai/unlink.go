package main

// bonsai unlink (spec section 4; plan-5, step 5.1.7): take Bonsai out of this checkout. Its table of flags and exit
// codes is unlinkWord; its plan is internal/engine/unlink.go's BuildUnlink, written by engine.Apply; with --json it
// prints the changes output (bonsai.changes/1, command unlink). The consent rule is init's and update's (engine.go):
// it previews first, writes with --yes, asks y/N only at a terminal, and without one exits 4. It needs no
// --allow-exec: it adds no code, and a removed hook line runs nothing. In an agent's session it is a person's command,
// refused but under a person's grant in a project in command mode (agent.go).

import (
	"os"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/textui"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func init() { register(unlinkWord) }

var unlinkWord = &Word{
	Name:    "unlink",
	Order:   3,
	Title:   "take Bonsai out of this checkout (spec section 4).",
	Summary: "take Bonsai out of this checkout, leaving edited files",
	Args:    "[flags]",
	About: `It reads the lock (.bonsai/lock.json) for what Bonsai wrote, and previews everything first. It removes the pack
files nobody edited, Bonsai's block in CLAUDE.md, Bonsai's lines in .claude/settings.json (each file too when
Bonsai's part was all it held), the tables .bonsai/tasks.md and .bonsai/sessions.md, .bonsai/.gitignore while
.bonsai/local/ is empty, bonsai.yaml, and the lock, last. bonsai.yaml goes even when you edited it: with no
bonsai.yaml, Bonsai's hook lets every call through, a session still running included.
It leaves, and names: pack files edited here and files written once (now the project's), Bonsai's block if you
edited it, .bonsai/STATE.md, .bonsai/local/ (with .bonsai/.gitignore while it holds anything), and this machine's
folder for the checkout in the Bonsai home.
Then it runs claude plugin uninstall <pack>@<marketplace> --scope project in this checkout for each pack, removing
Claude Code's record of the plugin's install here; this step never changes the exit code. That record is per
checkout: a worktree's install record is its own, so run bonsai unlink in each checkout. A plain git revert of the
link commit takes Bonsai's files out but leaves Claude Code's record (installed_plugins.json, in Claude Code's own
folder, outside git); to remove it, run claude plugin uninstall <pack>@<marketplace> --scope project in the checkout.
Run again, it changes nothing; bonsai init links the project again.
In an agent's session (an agent marker set: Claude Code's CLAUDE_CODE_CHILD_SESSION, AI_AGENT, pi's
PI_CODING_AGENT and the others the reference lists under person-command) it is a person's command, refused (exit 4,
nothing read or written) but in a project in command mode under a running task whose bonsai.allows a person gave
bonsai.yaml and .bonsai/lock.json: a person runs it in a terminal of their own.
`,
	Flags: []Flag{
		{Name: "--yes", Help: "write without asking"},
		{Name: "--json", Help: "print the changes document (bonsai.changes/1) instead of text; never asks"},
	},
	Exits: []Exit{
		{Code: 0, Means: "unlinked (or nothing to change: no bonsai.yaml and no lock here)"},
		{Code: 2, Means: "bad input: a flag wrong, or a bonsai.yaml, .claude/settings.json or CLAUDE.md block Bonsai does not\nread (nothing written)"},
		{Code: 3, Means: "runtime: git or the Bonsai home missing, a file that cannot be read or written (nothing written); or\nthe write stopped part-way (--json: result failed): run the same command again"},
		{Code: 4, Means: "wrong state (not in a git checkout, a Bonsai 0.4.3 workspace, bonsai.yaml with no lock, a lock Bonsai\ndoes not read, a test build outside its scratch folder), no --yes (the preview was printed; nothing written), or an\nagent's session (a person's command)"},
	},
	Examples: []string{"bonsai unlink --yes"},
	Refused:  changesRefused,
	Run:      runUnlink,
}

func runUnlink(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("unlink takes no %+q", c.rest[0]))
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
	// In an agent's session, unlink is a person's command (agent.go), refused before anything is read.
	if e := personCommand(dir, home, "bonsai unlink", "it takes Bonsai's guard and stop gate out of the project", "bonsai unlink --yes"); e != nil {
		return c.fail(e)
	}
	plan, err := engine.BuildUnlink(dir, home)
	if err != nil {
		return c.fail(engine.Unexpected(err))
	}
	out := func(result string, exit int, refusal *engine.Error, d *textui.Doc) int {
		if refusal != nil {
			noted(c.word, refusal)
		}
		if c.json {
			return c.printDoc(plan.Changes(result, refusal), exit)
		}
		if refusal != nil {
			if !d.Empty() {
				d.Gap()
			}
			d.Refuse(refusalOf(c.word, refusal))
		}
		if write(c.stdout, d.Render(textui.For(c.stdout))) != exitOK {
			return exitRuntime
		}
		return exit
	}
	if plan.Nothing() {
		d := &textui.Doc{}
		d.Title("bonsai unlink", "nothing to change: no bonsai.yaml and no lock here, so this checkout is not linked")
		plan.LeftTo(d)
		return out("nothing", engine.ExitOK, nil, d)
	}
	preview := func() *textui.Doc {
		d := &textui.Doc{}
		d.Title("bonsai unlink", "the preview")
		plan.PreviewTo(d, false, true)
		return d
	}
	yes := c.has("--yes")
	if !yes {
		if c.json || !interactive() {
			next := "to write it, run: bonsai unlink --yes"
			e := &engine.Error{Code: "needs-yes", Exit: engine.ExitState, What: "no --yes, and no terminal to ask at: nothing was written", Next: next}
			return out("preview", e.Exit, e, preview())
		}
		if write(c.stdout, preview().Render(textui.For(c.stdout))+"Take Bonsai out of this checkout? [y/N] ") != exitOK {
			return exitRuntime
		}
		if !yesAnswer() {
			d := &textui.Doc{}
			d.Title("bonsai unlink", "nothing written")
			d.Hint("run the command again when you want it written")
			return out("declined", engine.ExitState, nil, d)
		}
	}
	if err := applyPlan(plan); err != nil {
		e := engine.Unexpected(err)
		result := "refused"
		if e.Code == "partly-written" {
			result = "failed"
		}
		if c.json {
			return out(result, e.Exit, e, &textui.Doc{})
		}
		return c.fail(e)
	}
	name := ""
	if plan.Config != nil {
		name = plan.Config.Name
	}
	plan.Plugins = engine.UninstallPlugins(plan.Root, name, plan.OldMarket, plan.Removed, pluginCLI)
	d := &textui.Doc{}
	if yes {
		d = preview()
		d.Gap()
	}
	d.Title("bonsai unlink", "done")
	plan.AppliedTo(d)
	plan.PluginsTo(d)
	d.Hint("to link this project again: bonsai init (it reads no bonsai.yaml now, so give it --name, --source and --ref)")
	return out("applied", engine.ExitOK, nil, d)
}
