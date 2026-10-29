package main

// bonsai init and update (plan part 3; spec §4, §6): the run they share, the consent rule and what they print. Their
// tables of flags and exit codes are initWord (init.go) and updateWord (update.go); with --json both print the changes
// output, bonsai.changes/1 (engine.Plan.Changes), its error object filled on a refusal. unlink (step 5.1.7) prints the
// same output: its word, in a file of its own, reuses changesRefused and the plan's Changes.
//
// The consent rule (spec §3): a command that writes prints its preview first. It writes with --yes; at a terminal
// without --yes it asks y/N; without a terminal, in an agent session (an agent marker set, format.AgentMarkers) or
// with --json it never waits: it prints the preview and exits 4, naming the command with --yes.
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
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/textui"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// interactive reports whether a person can answer a y/N question: stdin and stdout are terminals and this is no
// agent session. A variable, so tests stand in a terminal.
var interactive = func() bool {
	return askable(os.Getenv) && isTerminal(os.Stdin) && isTerminal(os.Stdout)
}

// askable reports no agent marker (format.AgentMarkers) set in getenv: an agent is never asked y/N, whatever its
// terminal, in a stamped build too.
func askable(getenv func(string) string) bool {
	marker, _ := format.AgentMarker(getenv)
	return marker == ""
}

// stdin is where a y/N answer is read from (a variable for tests).
var stdin io.Reader = os.Stdin

// pluginCLI is Claude Code's plugin commands: init and update install each locked pack's plugin on this machine,
// and check compares what is installed with the lock (internal/engine/plugins.go). main sets the real one; nil,
// as in tests, asks Claude Code nothing.
var pluginCLI engine.PluginCLI

// engineFlags are init's or update's flags, read from the word's table.
type engineFlags struct {
	yes, diff, explain, newID, allowExec bool
	expect                               string // --expect: the fingerprint the plan must have (T-6009)
	keep, adopt                          []string
	values                               engine.InitValues
	unpaired                             string   // why the packs' flags do not pair up, "" when they do
	typed                                []string // the flags as typed that a next step repeats (not --yes, --diff, --explain or --json)
}

func readEngineFlags(c *call) *engineFlags {
	f := &engineFlags{yes: c.has("--yes"), diff: c.has("--diff"), explain: c.has("--explain"), newID: c.has("--new-id"), allowExec: c.has("--allow-exec"),
		keep: c.all("--keep"), adopt: c.all("--adopt"), expect: c.value("--expect")}
	f.values = engine.InitValues{Name: c.value("--name"), NeverEdit: c.all("--never-edit"), NeverRead: c.all("--never-read")}
	// The packs (BUG-122): the n-th --ref, and the n-th --path when --path is given, go with the n-th --source.
	sources, refs, paths := c.all("--source"), c.all("--ref"), c.all("--path")
	switch {
	case len(sources) <= 1 && len(refs) <= 1 && len(paths) <= 1:
		f.values.Source, f.values.Ref, f.values.Path = c.value("--source"), c.value("--ref"), c.value("--path")
	case len(refs) != len(sources) || (len(paths) != 0 && len(paths) != len(sources)):
		f.unpaired = fmt.Sprintf("init takes --ref once per --source, and --path not at all or once per --source "+
			"(\"\" for a repository's top): here --source %d, --ref %d, --path %d", len(sources), len(refs), len(paths))
	default:
		for i := range sources {
			p := engine.InitPack{Source: sources[i], Ref: refs[i]}
			if len(paths) > 0 {
				p.Path = paths[i]
			}
			if i == 0 {
				f.values.Source, f.values.Ref, f.values.Path = p.Source, p.Ref, p.Path
			} else {
				f.values.More = append(f.values.More, p)
			}
		}
	}
	exec := false
	for _, g := range c.given {
		switch g.name {
		case "--yes", "--diff", "--explain", "--json", "--expect":
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

// packRun is what bonsai pack hands the update run (pack.go): bonsai.yaml's edited bytes, the edit, the
// subcommand, and the arguments as typed, for the command lines it names.
type packRun struct {
	edit *engine.ConfigEdit
	sub  string
	args []string
}

// runEngine runs init or update: build the plan, print it, ask or refuse, write.
func runEngine(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("%s takes no %+q", c.word.Name, c.rest[0]))
	}
	return runEngineAs(c, nil)
}

// runEngineAs is runEngine, or with pack set bonsai pack's run: an update with bonsai.yaml's edited bytes
// (engine.Request.Config), its titles and command lines the pack word's.
func runEngineAs(c *call, pack *packRun) int {
	word := c.word.Name // init, update or unlink's word; for bonsai pack, the request is update's
	f := readEngineFlags(c)
	if f.unpaired != "" && pack == nil {
		return c.refuse(c.flagError("%s", f.unpaired))
	}
	cmdline := "bonsai " + word
	if pack != nil {
		for _, a := range pack.args {
			cmdline += " " + engine.ShellArg(a)
		}
	}
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
	// In an agent's session, --allow-exec is a person's command (agent.go), refused before anything is read or fetched.
	if f.allowExec {
		if e := personCommand(dir, home, cmdline, "--allow-exec consents to code that runs on this machine", cmdline+" --yes"); e != nil {
			return c.fail(e)
		}
	}
	req := engine.Request{Command: word, Dir: dir, Home: home, Version: version, NewID: f.newID, Keep: f.keep, Adopt: f.adopt,
		AllowExec: f.allowExec}
	if word == "init" {
		req.Init = &f.values
	}
	if pack != nil {
		req.Command, req.Config, req.Edit = "update", pack.edit.Raw, pack.edit
	}
	plan, err := engine.Build(req)
	if err != nil {
		return c.fail(engine.Unexpected(err))
	}
	if pack != nil {
		plan.WordName, plan.WordSub = "pack", pack.sub
	}
	// out prints the plan's outcome: with --json the changes output (its error object filled from refusal), else
	// the Doc on stdout, ending in the refusal's block when there is one.
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
	// closing adds init's closing words.
	closing := func(d *textui.Doc) {
		if word == "init" {
			d.Gap()
			engine.ClosingTo(d, plan.Config, home)
		}
	}
	// preview is the plan's preview, its title first.
	preview := func() *textui.Doc {
		d := &textui.Doc{}
		d.Title("bonsai "+word, "the preview")
		plan.PreviewTo(d, f.diff, f.explain)
		return d
	}
	// In an agent's session, recording another base branch for the workspace is a person's command (agent.go): the
	// ladder and the stop gate hold the main checkout to the base this machine recorded, and an agent that could
	// record the branch it switched to would make that check moot. Refused before anything is written.
	if was, now, changes := engine.BaseChange(plan); changes {
		if e := personCommand(dir, home, cmdline, "it would record the main checkout's branch "+strconv.QuoteToASCII(now)+
			" as the workspace's base branch in place of "+strconv.QuoteToASCII(was)+", the branch the ladder and the stop gate "+
			"hold the main checkout to", cmdline+" --yes (or switch the main checkout back: git switch "+was+")"); e != nil {
			return out("refused", e.Exit, e, &textui.Doc{})
		}
	}
	// So is recording settings for the checks looser than the ones this machine accepted (workspace/checks.go): the
	// ladder and the stop gate refuse bonsai.yaml's committed settings while they are looser, and an agent that could
	// record them would make that check moot. Refused before anything is written.
	if looser := engine.ChecksLooser(plan); len(looser) > 0 {
		if e := personCommand(dir, home, cmdline, "it would record settings for the checks looser than the ones this machine "+
			"last accepted for the workspace: "+looserList(looser), cmdline+" --yes (or put bonsai.yaml's settings back)"); e != nil {
			return out("refused", e.Exit, e, &textui.Doc{})
		}
	}
	// A plan that takes protections away (engine.Plan.Loosening: a pack taken out, or a deny rule or protected path no
	// planned pack holds) is a person's to write, whatever word applies it (update after a hand edit of bonsai.yaml as
	// well as pack): in an agent's session its --yes is refused by personOnly, which no grant opens (T-6009, the
	// director's answer of 28 Oct). Its preview is anyone's, and lists each protection that goes.
	if loose := plan.Loosening(); len(loose) > 0 && f.yes {
		if e := personOnly(cmdline+" --yes", "it takes protections away: "+looseningList(loose), cmdline+" --yes"); e != nil {
			return out("refused", e.Exit, e, preview())
		}
	}
	// --expect <fingerprint>: the plan built now must be the one the person read (engine.Plan.Fingerprint), else
	// nothing is written. It opens no gate: every check above and below holds as without it.
	if f.expect != "" {
		if fp := plan.Fingerprint(); fp != f.expect {
			e := &engine.Error{Code: "plan-changed", Exit: engine.ExitState,
				What: "the plan built now has the fingerprint " + fp + ", not the " + engine.ASCII(f.expect) + " --expect names: " +
					"something it would change differs from the preview read (a file, a pack, bonsai.yaml): nothing was written",
				Next: "read the preview again (" + cmdline + " --json), then run it with --expect and the fingerprint it prints"}
			return out("refused", e.Exit, e, preview())
		}
	}
	// The command that installs a plugin carrying code on this machine, a person's step: once the files are written
	// (or have nothing to change) the same word with --allow-exec does it, without the values or the flags that
	// settled conflicts.
	again := "bonsai " + word + " --allow-exec"
	if plan.Nothing() {
		plan.Plugins = pluginStep(plan, again, false)
		d := &textui.Doc{}
		lead := "nothing to change"
		if len(engine.ChecksLooser(plan)) > 0 {
			// recorded (below) is about to accept looser settings for the checks: that is what this run did.
			lead = "recorded the settings for the checks, looser than the ones accepted before"
		}
		d.Title("bonsai "+word, lead+packsAt(plan))
		plan.LeftAloneTo(d)
		recorded(d, plan)
		if f.yes {
			personalTo(d, home, plan)
		}
		plan.PluginsTo(d)
		if req.Command == "update" && f.yes {
			cacheCleanTo(d, home, plan)
		}
		closing(d)
		return out("nothing", engine.ExitOK, nil, d)
	}
	// In an agent's session, a change or removal of Bonsai's own guard or stop line is a person's command (agent.go):
	// refused with the preview, whatever else the command says.
	if touched := plan.OwnGuardChanges(); len(touched) > 0 {
		person := cmdline
		if plan.NeedsExec() && !f.allowExec {
			person += " --allow-exec"
		}
		lines := make([]string, len(touched))
		for i, t := range touched {
			lines[i] = t.Change + " " + strings.SplitN(t.Line, ":", 2)[0]
		}
		if e := personCommand(dir, home, cmdline, "this "+word+" changes or takes out Bonsai's own guard or stop line ("+
			strings.Join(lines, "; ")+"), which is what guards this project", person+" --yes"); e != nil {
			return out("refused", e.Exit, e, preview())
		}
	}
	// Code is consented to separately (spec §6): without --allow-exec, a plan that runs code is refused whole, at a
	// terminal too. The next step repeats the command with --allow-exec, and settles any conflict in the same run.
	if plan.NeedsExec() && !f.allowExec {
		base := cmdline + " --allow-exec"
		next := "read each item under Runs code; to write them with the rest, run: " + base + " --yes"
		if len(plan.Conflicts) > 0 {
			next = "read each item under Runs code, and settle the conflicts: " + plan.ConflictNext(base)
		}
		if marker, agent := agentSession(); marker != "" {
			// --allow-exec is a person's command here (agent.go): the next step is the person's.
			next = "read each item under Runs code; " + personNext(agent, base+" --yes")
			if len(plan.Conflicts) > 0 {
				next = "read each item under Runs code; settle the conflicts, and " + personNext(agent, plan.ConflictNext(base))
			}
		}
		n := len(plan.RunsCode)
		e := &engine.Error{Code: "needs-allow-exec", Exit: engine.ExitState,
			What: fmt.Sprintf("this %s writes code that runs on this machine (%d %s under Runs code), which needs --allow-exec as well as --yes: nothing was written",
				word, n, map[bool]string{true: "item", false: "items"}[n == 1]),
			Next: next}
		return out("refused", e.Exit, e, preview())
	}
	conflict := func(printed bool) int {
		e := &engine.Error{Code: "conflicts", Exit: engine.ExitConflict, What: conflictWhat(plan) + ": nothing was written",
			Next: plan.ConflictNext(cmdline)}
		d := &textui.Doc{}
		if !printed {
			d = preview()
		}
		return out("conflict", e.Exit, e, d)
	}
	if !f.yes {
		if c.json || !interactive() {
			next := "to write it, run: " + cmdline + " --yes"
			if len(plan.Conflicts) > 0 {
				next = plan.ConflictNext(cmdline)
			}
			e := &engine.Error{Code: "needs-yes", Exit: engine.ExitState, What: "no --yes, and no terminal to ask at: nothing was written", Next: next}
			return out("preview", e.Exit, e, preview())
		}
		// At a terminal (never with --json): the person answers.
		if write(c.stdout, preview().Render(textui.For(c.stdout))+"Write these changes? [y/N] ") != exitOK {
			return exitRuntime
		}
		if !yesAnswer() {
			d := &textui.Doc{}
			d.Title("bonsai "+word, "nothing written")
			d.Hint("run the command again when you want it written")
			return out("declined", engine.ExitState, nil, d)
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
			return out(result, e.Exit, e, &textui.Doc{})
		}
		return c.fail(e)
	}
	plan.Plugins = pluginStep(plan, again, true)
	d := &textui.Doc{}
	if f.yes {
		d = preview()
		d.Gap()
	}
	d.Title("bonsai "+word, "written"+packsAt(plan))
	plan.AppliedTo(d)
	plan.LeftAloneTo(d)
	recorded(d, plan)
	personalTo(d, home, plan)
	plan.PluginsTo(d)
	if req.Command == "update" {
		cacheCleanTo(d, home, plan)
	}
	closing(d)
	return out("applied", engine.ExitOK, nil, d)
}

// cacheCleanTo cleans the pack cache after update --yes, written or with nothing to change, never a preview
// (engine.CleanCache: the home's cache_keep_days, never a clone the lock or bonsai.yaml names, never an --adopt copy)
// and adds its one line to update's text and to the plan's Cache, which --json prints (set 13, BUG-017). It never
// changes the exit code or the result.
func cacheCleanTo(d *textui.Doc, home string, plan *engine.Plan) {
	var keep []string
	for _, p := range plan.Config.Packs {
		keep = append(keep, p.Source)
	}
	if lock := plan.NewLock(); lock != nil {
		for _, p := range lock.Packs {
			keep = append(keep, p.Source)
		}
	}
	cc := engine.CleanCache(home, keep)
	if line := cc.Line(); line != "" {
		tone := textui.Dim
		if cc.Problem != "" || len(cc.Failed) > 0 {
			tone = textui.Warn
		}
		d.Para("Cache", line, tone)
		plan.Cache = &format.ChangesLine{Line: engine.ASCII(line), Warn: tone == textui.Warn}
	}
}

// personalTo runs init's and update's step on the personal memory layer once they wrote, or had nothing to change
// with --yes, never in a preview (engine.PersonalAfterWrite: the Windows copy refreshed from personal_from, then the
// empty index when the home has none), and adds its lines to the text and to the plan's Memory, which --json prints
// (set 13, BUG-017). It never changes the exit code or the result.
func personalTo(d *textui.Doc, home string, plan *engine.Plan) {
	for _, l := range engine.PersonalAfterWrite(home, time.Now) {
		tone := textui.Dim
		if l.Warn {
			tone = textui.Warn
		}
		d.Para("Memory", l.Text, tone)
		plan.Memory = append(plan.Memory, format.ChangesLine{Line: engine.ASCII(l.Text), Warn: l.Warn})
	}
}

// pluginStep brings this machine's plugins to the plan once it is written (or has nothing to write;
// internal/engine/plugins.go): it removes Claude Code's record of each pack the plan took out of bonsai.yaml, after
// Bonsai's own lines left .claude/settings.json, then asks Claude Code to install each locked pack's plugin, and,
// when the command wrote, removes this checkout's stale records of the locked packs under an older marketplace name
// of this workspace. Its results never change the exit code; a preview or a refusal never reaches it.
func pluginStep(plan *engine.Plan, again string, written bool) []engine.PluginResult {
	out := engine.UninstallPlugins(plan.Root, plan.Config.Name, plan.OldMarket, plan.Removed, pluginCLI)
	out = append(out, engine.InstallPlugins(plan.Root, plan.Config, plan.NewLock(), pluginCLI, plan.PluginConsent(again))...)
	if written {
		out = append(out, engine.PruneStalePlugins(plan.Root, plan.Config, plan.NewLock(), pluginCLI)...)
	}
	return out
}

// recorded writes this machine's record of the checkout (engine.RecordCheckout) once init or update ended without a
// refusal, and adds a note when it could not: the project's files are written either way. When it recorded settings
// for the checks looser than the ones accepted before (a person's update: an agent's is refused first), the note
// names each loosening.
func recorded(d *textui.Doc, plan *engine.Plan) {
	looser := engine.ChecksLooser(plan)
	if err := engine.RecordCheckout(plan); err != nil {
		d.Para("note", "this machine's record of the checkout was not written: "+engine.ASCII(err.Error()), textui.Warn)
		return
	}
	if len(looser) > 0 {
		d.Para("note", "this machine now accepts bonsai.yaml's settings for the checks, looser than the ones it accepted before: "+
			looserList(looser), textui.Warn)
	}
}

// maxLoosenings caps the loosenings a line names; the rest are counted.
const maxLoosenings = 8

// looserList joins loosenings for a line, at most maxLoosenings of them, the rest counted.
func looserList(l []string) string {
	if len(l) > maxLoosenings {
		l = append(l[:maxLoosenings:maxLoosenings], "and "+strconv.Itoa(len(l)-maxLoosenings)+" more")
	}
	return engine.ASCII(strings.Join(l, "; "))
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
		if m.To == "" {
			parts = append(parts, m.ID+" taken out")
			continue
		}
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

// looseningList names each protection a plan takes away, for a refusal's sentence.
func looseningList(loose []engine.Loosening) string {
	parts := make([]string, len(loose))
	for i, l := range loose {
		switch l.Kind {
		case engine.LoosePack:
			parts[i] = "the pack " + l.Value + " taken out"
		case engine.LooseDeny:
			parts[i] = "the deny rule " + engine.ASCII(l.Value) + " (" + l.Pack + ")"
		default:
			parts[i] = "the protected path " + engine.ASCII(l.Value) + " (" + l.Pack + ")"
		}
	}
	return strings.Join(parts, "; ")
}
