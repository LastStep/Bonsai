package main

// bonsai init, update and check (plan part 3; spec §4, §6): their flags, the consent rule and what they print.
//
// The consent rule (spec §3): a command that writes prints its preview first. It writes with --yes; at a terminal
// without --yes it asks y/N; without a terminal, in an agent session (CLAUDE_CODE_CHILD_SESSION set) or with --json
// it never waits: it prints the preview and exits 4, naming the command with --yes.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
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

const initUsage = `bonsai init: link this project to Bonsai (spec sections 4 and 6).
It writes bonsai.yaml (a comment on every line), then the packs' files, the instruction block in CLAUDE.md,
Bonsai's lines in .claude/settings.json (hook line, deny rules, plugin wiring), each pack's plugin turned on in this
checkout's .claude/settings.local.json (never committed), .bonsai/.gitignore, and the lock .bonsai/lock.json last.
It previews every file and settings line first, and writes with --yes (or y at a terminal). Then it asks Claude Code
to install each pack's plugin on this machine at the locked commit (claude plugin install --scope local), as update
does.
Run it in the project's checkout. In a project that already has bonsai.yaml it needs no values and works as
bonsai update does; run again with nothing changed, it changes no byte.
Flags:
  --name N          the workspace's name: a lower-case letter, then lower-case letters, digits and dashes
  --source URL      the pack's git repository (fetched with this machine's git, into the Bonsai home's cache)
  --ref R           the pack's release tag, or a 40-character commit
  --path P          the pack's folder inside its repository (default: the repository's top)
  --never-edit P    a path no agent may ever edit, written as a deny rule; give it once per path
  --new-id          give this project a new workspace id and empty .bonsai/local/: for a copy meant as a new project
  --yes             write without asking
  --diff            show each file's changes as a diff in the preview
  --keep P          settle a conflict on P by keeping your edit (the file becomes kind kept)
  --adopt P         settle a conflict on P by taking the pack's copy; yours is saved in the Bonsai home's cache
  --json            print a JSON document instead of text; never asks
  --help            print this help
Without bonsai.yaml, --name, --source and --ref are required. Not built yet: --allow-exec (step 5.1).
Exit codes: 0 linked (or nothing to change), 2 bad input (a value missing or wrong), 3 runtime (a fetch failed),
  4 wrong state or no --yes (the preview was printed; nothing written), 5 conflicts (nothing written).
Example: bonsai init --name demo --source https://github.com/LastStep/bonsai-test-pack --ref 506205354b7589f82f849820987aad17dba3309d --yes
`

const updateUsage = `bonsai update: bring this project's packs to the refs in bonsai.yaml (spec section 6).
It fetches each pack, then decides per file from three fingerprints (what the lock says Bonsai wrote, what is on
disk, what the pack now gives): unchanged, updated, created, adopted, changed (your edit, left alone) or conflict.
The preview names every file and every line of .claude/settings.json and .claude/settings.local.json it adds,
changes or removes, with a sentence. All or nothing: every file is staged, then renamed, the lock last.
Then, written or with nothing to change, it brings this machine's plugins to the lock: for each pack it runs
claude plugin install <pack>@<marketplace> --scope local (a no-op once installed). A marketplace Claude Code has not
registered yet (a new commit, or a folder not yet trusted) waits for the next Claude Code session in this checkout,
which fetches the plugin itself; the output says so, with the next step. This step never changes the exit code.
Flags:
  --yes         write without asking
  --diff        show each file's changes as a diff in the preview
  --keep P      settle a conflict on P by keeping your edit (the file becomes kind kept; a later pack change to it
                is a conflict again)
  --adopt P     settle a conflict on P by taking the pack's copy; yours is saved in the Bonsai home's cache
  --json        print a JSON document instead of text; never asks
  --help        print this help
Not built yet: --allow-exec (step 5.1). A change to a hook line runs code, so this build refuses it (exit 4).
Exit codes: 0 updated (or nothing to change), 2 bad input, 3 runtime (a fetch failed), 4 wrong state, no --yes or
  a hook-line change (the preview was printed; nothing written), 5 conflicts (nothing written).
Example: bonsai update --yes
`

const checkUsage = `bonsai check: findings on this checkout (spec section 6): the lock against the files (a pack file, the
block in CLAUDE.md or Bonsai's lines in .claude/settings.json edited or missing), bonsai.yaml against the lock,
.bonsai/.gitignore, any file from .bonsai/local/ (or .claude/settings.local.json) that git tracks or has staged, and
this machine's plugins against the lock: it asks Claude Code (claude plugin list --json) and reports a pack's plugin
that sessions here load at another commit (a finding), or the locked one not installed yet (a warning). It writes
nothing and fetches nothing. Warnings never change the exit code.
Flags:
  --json    print a JSON document instead of text
  --help    print this help
Not built yet: --write, --schema, --pack (step 5.1).
Exit codes: 0 no findings, 1 findings, 2 bad input, 4 not a linked checkout.
Example: bonsai check --json
`

type engineFlags struct {
	yes, diff, asJSON, newID bool
	keep, adopt              []string
	values                   engine.InitValues
	valuesGiven              bool
	typed                    []string // the value flags as typed, for the command a next step repeats
}

// parseEngineFlags reads init's or update's flags. --flag value and --flag=value both work.
func parseEngineFlags(word string, args []string, stdout, stderr io.Writer) (*engineFlags, int, bool) {
	f := &engineFlags{}
	takesValue := map[string]bool{"--keep": true, "--adopt": true}
	if word == "init" {
		for _, k := range []string{"--name", "--source", "--ref", "--path", "--never-edit"} {
			takesValue[k] = true
		}
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		name, value, hasValue := a, "", false
		if j := strings.IndexByte(a, '='); j > 0 && strings.HasPrefix(a, "--") {
			name, value, hasValue = a[:j], a[j+1:], true
		}
		if takesValue[name] && !hasValue {
			if i+1 >= len(args) {
				return nil, refuse(stderr, word+" "+name+" needs a value", "run `bonsai "+word+" --help` to see its flags"), true
			}
			i++
			value, hasValue = args[i], true
		}
		switch {
		case name == "--help" || name == "-h":
			if word == "init" {
				return nil, write(stdout, initUsage), true
			}
			return nil, write(stdout, updateUsage), true
		case hasValue && !takesValue[name]:
			return nil, refuse(stderr, fmt.Sprintf("%s takes no value for %+q", word, name), "run `bonsai "+word+" --help` to see its flags"), true
		case name == "--yes":
			f.yes = true
		case name == "--diff":
			f.diff = true
		case name == "--json":
			f.asJSON = true
		case name == "--new-id" && word == "init":
			f.newID = true
			f.typed = append(f.typed, "--new-id")
		case name == "--keep":
			f.keep = append(f.keep, value)
		case name == "--adopt":
			f.adopt = append(f.adopt, value)
		case name == "--allow-exec":
			return nil, refuse(stderr, word+" --allow-exec is not built yet: it comes with step 5.1, and until then this build refuses every hook-line change",
				"run `bonsai "+word+"` without --allow-exec; a hook-line change waits for step 5.1"), true
		case word == "init" && takesValue[name]:
			f.valuesGiven = true
			f.typed = append(f.typed, name, value)
			switch name {
			case "--name":
				f.values.Name = value
			case "--source":
				f.values.Source = value
			case "--ref":
				f.values.Ref = value
			case "--path":
				f.values.Path = value
			case "--never-edit":
				f.values.NeverEdit = append(f.values.NeverEdit, value)
			}
		default:
			return nil, refuse(stderr, fmt.Sprintf("%s takes no %+q", word, a), "run `bonsai "+word+" --help` to see its flags"), true
		}
	}
	return f, 0, false
}

func runInit(args []string, stdout, stderr io.Writer) int {
	return runEngine("init", args, stdout, stderr)
}

func runUpdate(args []string, stdout, stderr io.Writer) int {
	return runEngine("update", args, stdout, stderr)
}

// runEngine runs init or update: build the plan, print it, ask or refuse, write.
func runEngine(word string, args []string, stdout, stderr io.Writer) int {
	f, code, done := parseEngineFlags(word, args, stdout, stderr)
	if done {
		return code
	}
	cmdline := "bonsai " + word
	for _, t := range f.typed {
		cmdline += " " + engine.ShellArg(t)
	}
	fail := func(e *engine.Error) int {
		if f.asJSON {
			return printJSON(stdout, engine.ErrorJSON(word, e), e.Exit)
		}
		_, _ = fmt.Fprintf(stderr, "bonsai %s: %s.\nnext: %s\n", word, strings.TrimSuffix(e.What, "."), e.Next)
		return e.Exit
	}
	dir, err := os.Getwd()
	if err != nil {
		return fail(&engine.Error{Exit: engine.ExitRuntime, What: "the current folder cannot be read", Next: "run it from a folder inside the project"})
	}
	home, err := workspace.Home()
	if err != nil {
		return fail(&engine.Error{Exit: engine.ExitRuntime, What: err.Error(), Next: "set BONSAI_HOME to the folder Bonsai should use"})
	}
	req := engine.Request{Command: word, Dir: dir, Home: home, Version: version, NewID: f.newID, Keep: f.keep, Adopt: f.adopt}
	if word == "init" {
		req.Init = &f.values
	}
	plan, err := engine.Build(req)
	if err != nil {
		e, ok := err.(*engine.Error)
		if !ok {
			e = &engine.Error{Exit: engine.ExitRuntime, What: err.Error(), Next: "run the command again"}
		}
		return fail(e)
	}
	out := func(result string, exit int, refusal *engine.Error, text string) int {
		if f.asJSON {
			return printJSON(stdout, plan.JSON(result, exit, refusal), exit)
		}
		if write(stdout, text) != exitOK {
			return exitRuntime
		}
		return exit
	}
	closing := ""
	if word == "init" {
		closing = engine.ClosingWords(plan.Config, home)
	}
	if plan.Nothing() {
		plan.Plugins = engine.InstallPlugins(plan.Root, plan.Config, plan.NewLock(), pluginCLI)
		return out("nothing", engine.ExitOK, nil, "bonsai "+word+": nothing to change"+packsAt(plan)+".\n"+plan.PluginsText()+closing)
	}
	head := "bonsai " + word + ": the preview.\n"
	preview := head + plan.Preview(f.diff)
	if plan.HookChange {
		e := &engine.Error{Exit: engine.ExitState,
			What: "this update changes a hook line, and this build of Bonsai refuses every hook-line change: nothing was written",
			Next: "a hook-line change needs --allow-exec as well as --yes: " + cmdline + " --yes --allow-exec (--allow-exec comes with step 5.1; until then the change cannot be applied, and the project stays as it is)"}
		return out("refused", e.Exit, e, preview+"Refused: "+e.What+".\nnext: "+e.Next+"\n")
	}
	conflict := func(printed bool) int {
		e := &engine.Error{Exit: engine.ExitConflict, What: conflictWhat(plan) + ": nothing was written", Next: plan.ConflictNext(cmdline)}
		text := "Stopped: " + e.What + ".\nnext: " + e.Next + "\n"
		if !printed {
			text = preview + text
		}
		return out("conflict", e.Exit, e, text)
	}
	if !f.yes {
		if f.asJSON || !interactive() {
			next := "to write it, run: " + cmdline + " --yes"
			if len(plan.Conflicts) > 0 {
				next = plan.ConflictNext(cmdline)
			}
			e := &engine.Error{Exit: engine.ExitState, What: "no --yes, and no terminal to ask at: nothing was written", Next: next}
			return out("preview", e.Exit, e, preview+"Nothing written yet.\nnext: "+next+"\n")
		}
		if write(stdout, preview+"Write these changes? [y/N] ") != exitOK {
			return exitRuntime
		}
		if !yesAnswer() {
			return out("declined", engine.ExitState, nil, "Nothing written.\n")
		}
		if len(plan.Conflicts) > 0 {
			return conflict(true)
		}
	} else if len(plan.Conflicts) > 0 {
		return conflict(false)
	}
	if err := engine.Apply(plan); err != nil {
		e, ok := err.(*engine.Error)
		if !ok {
			e = &engine.Error{Exit: engine.ExitRuntime, What: err.Error(), Next: "run the command again"}
		}
		return fail(e)
	}
	plan.Plugins = engine.InstallPlugins(plan.Root, plan.Config, plan.NewLock(), pluginCLI)
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

func printJSON(stdout io.Writer, doc schema.Object, exit int) int {
	b, err := schema.Encode(doc)
	if err != nil {
		return exitRuntime
	}
	if write(stdout, string(b)) != exitOK {
		return exitRuntime
	}
	return exit
}

func runCheck(args []string, stdout, stderr io.Writer) int {
	asJSON := false
	for _, a := range args {
		switch a {
		case "--json":
			asJSON = true
		case "--help", "-h":
			return write(stdout, checkUsage)
		case "--write", "--pack", "--schema":
			return refuse(stderr, "check "+a+" is not built yet (it comes with step 5.1)", "run `bonsai check` without "+a)
		default:
			return refuse(stderr, fmt.Sprintf("check takes no %+q", a), "run `bonsai check --help` to see its flags")
		}
	}
	dir, err := os.Getwd()
	if err != nil {
		return refuse(stderr, "check cannot read the current folder", "run it from a folder inside the project")
	}
	home, err := workspace.Home()
	if err != nil {
		return refuse(stderr, err.Error(), "set BONSAI_HOME to the folder Bonsai should use")
	}
	r, err := engine.Check(dir, home)
	if err != nil {
		e := err.(*engine.Error)
		if asJSON {
			return printJSON(stdout, engine.ErrorJSON("check", e), e.Exit)
		}
		_, _ = fmt.Fprintf(stderr, "bonsai check: %s.\nnext: %s\n", strings.TrimSuffix(e.What, "."), e.Next)
		return e.Exit
	}
	engine.ComparePlugins(r, pluginCLI)
	exit := engine.ExitOK
	if len(r.Findings) > 0 {
		exit = engine.ExitFindings
	}
	if asJSON {
		return printJSON(stdout, r.JSON(exit), exit)
	}
	if write(stdout, r.Text()) != exitOK {
		return exitRuntime
	}
	return exit
}
