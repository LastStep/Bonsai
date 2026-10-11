package guard

// Bonsai's own commands in a shell call (design/plan-5.md, 5.3.2 note 7; spec §4, §7; contract §10.6), under the
// answer to step 5.3's second question, (ii): who may consent to code and take the guard out.
//
//   - bonsai settings set, bonsai labels attach and detach are refused in an agent session, always (home-command):
//     they write this machine's settings and label files, which only a person's own terminal changes (the commands
//     refuse themselves too, step 5.6).
//   - bonsai with --allow-exec (the consent to code), bonsai unlink (which takes the guard out), and a line that unsets
//     or changes CLAUDE_CODE_CHILD_SESSION (which hides an agent's session from those commands' own refusal) are a
//     person's commands (person-command): allowed to an agent only with a person's grant in a project in command
//     mode, that is, the project's running task's bonsai.allows holds bonsai.yaml and .bonsai/lock.json (rule 3's
//     grant; in command mode only a person's tap puts a person-only path there, tripwire.go).
//   - bonsai init, update and unlink are judged as one edit of bonsai.yaml and .bonsai/lock.json at the top of the
//     checkout they run in (may), so an agent's update needs a grant of both, as any edit of them does.
//   - Every other word is allowed (bonsai-command).
//
// The command itself refuses too (step 5.3.6 note 8), since a variable, xargs, a script or a subprocess can hide a
// flag from the guard: the guard is the second layer. ownCommand judges one bonsai call's words, found by step 5.3.3's
// splitter whatever folder bonsai is called from; ownLine judges a command line's text for the variable, and is called
// for every Bash, PowerShell and Monitor call (and may be called again for each text the splitter finds nested).
//
// What ownLine reads (no shell reader: a scan of the text for the variable's name, letter case ignored as Windows
// ignores it): an assignment of it (NAME=..., export NAME=..., set NAME=..., $env:NAME = ...,
// [Environment]::SetEnvironmentVariable('NAME', ...)); the name in a command that unsets or sets variables (unset,
// export, declare, typeset, local, readonly, env, set, setx); the name as a PowerShell provider path (Env:NAME) in a
// command that is not one of the read cmdlets; and, whatever it names, a command that starts a program with an empty
// or new environment (env -i, env -, env --ignore-environment, exec -c, Start-Process -UseNewEnvironment). A read of
// the variable ($NAME, ${NAME}, %NAME%, $env:NAME, printenv NAME) and the name as an argument of another program
// (grep NAME) are allowed. Not seen: the name built from parts or decoded at run time, an interpreter's own calls
// (python -c, node -e), and a new session that starts without the variable (ssh, su, a terminal multiplexer, a
// scheduler); the command's own refusal and the Desk stand behind it, as question A says for every shell write.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// ChildSessionEnv names the variable Claude Code sets in every process its tools and hooks start (its env-vars
// reference), by which Bonsai's commands tell an agent's session from a person's terminal (step 5.3.6 note 8).
const ChildSessionEnv = "CLAUDE_CODE_CHILD_SESSION"

// personNext is the next step of a person's command refused to an agent.
const personNext = "a person runs this command in a terminal of their own, not one Claude Code started"

// ownCommand judges one bonsai call: argv its words as the shell gives them to the program (argv[0] the program),
// dir the folder it runs in (absolute; the session's project folder when it cannot be told). Its record goes in the
// session's project (rule 8: a shell call is judged there).
func (j *judge) ownCommand(dir string, argv []string) Decision {
	if len(argv) < 2 {
		return allow(RuleBonsaiCommand, "")
	}
	args := argv[1:]
	word := ""
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			word = strings.ToLower(a)
			break
		}
	}
	shown := "bonsai " + word // the word alone: the call's other words may name paths or values (contract §2.6)
	if allowExec(args) {
		shown += " --allow-exec"
	}
	shown = strings.TrimSpace(shown)
	later := func(subs ...string) bool {
		seen := false
		for _, a := range args {
			if seen {
				for _, s := range subs {
					if strings.EqualFold(a, s) {
						return true
					}
				}
			}
			if strings.ToLower(a) == word {
				seen = true
			}
		}
		return false
	}
	switch {
	case word == "settings" && later("set"), word == "labels" && later("attach", "detach"):
		return deny(RuleHomeCommand, fmt.Sprintf("%s refused: it writes this machine's settings or label files in Bonsai's "+
			"home, which never change from an agent's session", shown), personNext)
	case word != "init" && word != "update" && word != "unlink" && !allowExec(args):
		return allow(RuleBonsaiCommand, "")
	}
	top := checkoutTop(dir)
	if allowExec(args) || word == "unlink" {
		what := "--allow-exec consents to code that runs on this machine"
		if word == "unlink" {
			what = "bonsai unlink takes Bonsai's guard out of the project"
		}
		if why := j.personGrant(top); why != "" {
			d := deny(RulePersonCommand, fmt.Sprintf("%s refused: %s, a person's command: an agent may run it only in a "+
				"project in command mode under a running task whose bonsai.allows a person gave bonsai.yaml and "+
				".bonsai/lock.json, and %s", shown, what, why), personNext)
			d.Target = target(j.session, []string{filepath.Join(top, workspace.ConfigFile)})
			return d
		}
	}
	if word == "init" || word == "update" || word == "unlink" {
		for _, rel := range []string{workspace.ConfigFile, workspace.LockFile} {
			d, _ := j.may(change{tool: "bonsai " + word, path: filepath.Join(top, filepath.FromSlash(rel)), cwd: top, base: top})
			if !d.Allow {
				d.Why = "bonsai " + word + " writes bonsai.yaml and .bonsai/lock.json, so it is judged as an edit of them: " + d.Why
				return d
			}
		}
	}
	return allow(RuleBonsaiCommand, "")
}

// allowExec reports the consent flag among a bonsai call's words, in any spelling the guard can see (--allow-exec,
// -allow-exec, --allow-exec=..., any letter case): the command takes the first alone, and refuses the others itself.
func allowExec(args []string) bool {
	for _, a := range args {
		if strings.HasPrefix(strings.ToLower(strings.TrimLeft(a, "-")), "allow-exec") && strings.HasPrefix(a, "-") {
			return true
		}
	}
	return false
}

// personGrant says why a person's command may not run for the project at top: "" when the project reads command and
// its running task grants bonsai.yaml and .bonsai/lock.json (a person's grant), else the reason, ASCII words that
// follow "and".
func (j *judge) personGrant(top string) string {
	chain := j.chain(filepath.Join(top, workspace.ConfigFile))
	g := j.grantor(chain)
	switch {
	case g == nil && len(chain) > 0:
		return "the project it runs in is not verified on this machine (" + j.notVerified(chain[0]).Why + ")"
	case g == nil:
		return "no linked project holds the folder it runs in"
	}
	ms, err := j.modeOf(g)
	switch {
	case err != nil:
		return "this machine's settings for the project cannot be read (" + oneLine(err.Error()) + ")"
	case ms.StatusWrites != workspace.StatusWritesCommand:
		return "the project reads status_writes agents (no studio manages it), where only a person runs it"
	}
	for _, rel := range []string{workspace.ConfigFile, workspace.LockFile} {
		r, ok := relTo(g.root, filepath.Join(top, filepath.FromSlash(rel)))
		if !ok {
			return "the folder it runs in is not the project's top"
		}
		if yes, why := j.grantOf(g).grants(r, caseInsensitive(g.root)); !yes {
			return "for " + rel + ": " + why
		}
	}
	return ""
}

// checkoutTop is the top of the checkout holding dir, as git finds it: the nearest folder at or above it holding a
// .git; dir itself when none does.
func checkoutTop(dir string) string {
	dir = filepath.Clean(dir)
	for d := dir; ; {
		if _, err := os.Lstat(filepath.Join(d, ".git")); err == nil {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			return dir
		}
		d = parent
	}
}

// ownLine judges a command line's text for CLAUDE_CODE_CHILD_SESSION (this file's comment): tool is Bash, PowerShell
// or Monitor. A line that changes it is a person's command, allowed only with a person's grant in the session's
// project in command mode.
func (j *judge) ownLine(tool, text string) Decision {
	what := childSessionChange(text)
	if what == "" {
		return allow(RuleShell, "")
	}
	why := "the session is in no linked project"
	if j.session != nil {
		why = j.personGrant(j.session.root)
	}
	if why == "" {
		return allow(RuleShell, "")
	}
	return deny(RulePersonCommand, fmt.Sprintf("%s refused: the command line %s, which hides an agent's session from Bonsai's "+
		"commands; that is a person's step, allowed to an agent only in a project in command mode under a person's grant "+
		"of bonsai.yaml and .bonsai/lock.json, and %s", tool, what, why), personNext)
}

// envChangers are the commands that set or unset a shell's variables, by base name, letter case ignored.
var envChangers = []string{"unset", "export", "declare", "typeset", "local", "readonly", "env", "set", "setx"}

// envReaders are the PowerShell commands that only read an Env: provider path.
var envReaders = []string{"get-item", "gi", "get-childitem", "gci", "ls", "dir", "test-path", "get-content", "gc", "cat",
	"type", "get-itemproperty", "gp", "resolve-path", "rvpa", "get-itempropertyvalue", "gpv"}

// envPrinters are the commands that only print their words.
var envPrinters = []string{"echo", "printf", "print", "write-host", "write-output", "write"}

// childSessionChange says how a line's text changes CLAUDE_CODE_CHILD_SESSION or starts a program without it ("" for
// neither), in words that follow "the line ...".
func childSessionChange(text string) string {
	low := strings.ToLower(text)
	if what := emptyEnvironment(low); what != "" {
		return what
	}
	name := strings.ToLower(ChildSessionEnv)
	for from := 0; ; {
		i := strings.Index(low[from:], name)
		if i < 0 {
			return ""
		}
		i += from
		from = i + len(name)
		pre, post := low[:i], low[i+len(name):]
		prog := programAt(low, i)
		switch {
		case strings.HasSuffix(pre, "$env:") || strings.HasSuffix(pre, "${env:"):
			if assigns(strings.TrimPrefix(post, "}"), true) {
				return "assigns " + ChildSessionEnv
			}
		case strings.HasSuffix(pre, "$") || strings.HasSuffix(pre, "${") || strings.HasSuffix(pre, "%"):
			// a read
		case strings.HasSuffix(pre, "env:") || strings.HasSuffix(pre, `env:\`) || strings.HasSuffix(pre, "env:/"):
			if !listHas(envReaders, prog) {
				return "changes " + ChildSessionEnv + " through its Env: path"
			}
		case setEnvironmentCall(pre):
			return "sets " + ChildSessionEnv + " through .NET"
		case assigns(post, false) && !strings.HasSuffix(pre, `"`) && !strings.HasSuffix(pre, "'") && !listHas(envPrinters, prog):
			return "assigns " + ChildSessionEnv
		case listHas(envChangers, prog):
			return "unsets or exports " + ChildSessionEnv + " (" + prog + ")"
		}
	}
}

// assigns reports an assignment's operator at the start of the text after a variable's name: = (not ==), and with
// spaced (PowerShell) spaces allowed before it.
func assigns(post string, spaced bool) bool {
	if spaced {
		post = strings.TrimLeft(post, " \t")
	}
	return strings.HasPrefix(post, "=") && !strings.HasPrefix(post, "==")
}

// setEnvironmentCall reports text ending in a .NET call that sets a variable: SetEnvironmentVariable(, then a quote.
func setEnvironmentCall(pre string) bool {
	p := strings.TrimRight(pre, `'" `+"\t")
	p = strings.TrimRight(strings.TrimSuffix(p, "("), " \t")
	return strings.HasSuffix(p, "setenvironmentvariable")
}

// programAt gives the program of the simple command holding the text at i (lower case): its first word after the
// last command boundary before i, past NAME=value words and the words that run another program, by base name,
// without .exe.
func programAt(low string, i int) string {
	start := strings.LastIndexAny(low[:i], ";&|\n\r(){}`") + 1
	words := strings.Fields(low[start:i])
	for len(words) > 0 {
		w := strings.Trim(words[0], `"'`)
		switch {
		case w == "!" || w == "if" || w == "then" || w == "do" || w == "else" || w == "&" || w == "." || w == "@":
		case listHas(lineRunners, base(w)):
		case strings.HasPrefix(w, "-") || (len(w) == 2 && w[0] == '/'):
		case strings.Contains(w, "=") && !strings.HasPrefix(w, "="):
			eq := strings.IndexByte(w, '=')
			if strings.ContainsAny(w[:eq], "/\\$") {
				return base(w)
			}
		default:
			return base(w)
		}
		words = words[1:]
	}
	return ""
}

// lineRunners are the words that run the program or the text after them (programAt passes over them, and over their
// options): sudo and its kind, and the shells given a command's text (bash -c "unset ...", cmd /c set ...).
var lineRunners = []string{"sudo", "command", "builtin", "exec", "nohup", "nice", "time", "eval", "xargs", "bash", "sh",
	"zsh", "dash", "ksh", "cmd", "powershell", "pwsh", "wsl", "invoke-expression", "iex"}

// base is a program word's base name, without .exe.
func base(w string) string {
	w = strings.ReplaceAll(w, `\`, "/")
	if k := strings.LastIndexByte(w, '/'); k >= 0 {
		w = w[k+1:]
	}
	return strings.TrimSuffix(w, ".exe")
}

// emptyEnvironment says how a line starts a program with an empty or new environment ("" for none): a command whose
// program is env with -i, - or --ignore-environment (any unambiguous prefix), exec with -c, or Start-Process with
// -UseNewEnvironment (any prefix PowerShell takes). The program is the command's first word past NAME=value words and
// the words that run another program (lineRunners), as programAt finds it.
func emptyEnvironment(low string) string {
	for _, cmd := range splitRough(low) {
		words := strings.Fields(cmd)
		k := 0
		for k < len(words) {
			w := strings.Trim(words[k], `"'`)
			b := base(w)
			switch {
			case b == "exec":
				for _, a := range words[k+1:] {
					a = strings.Trim(a, `"'`)
					if !strings.HasPrefix(a, "-") {
						break
					}
					if !strings.HasPrefix(a, "--") && strings.ContainsRune(a[1:], 'c') {
						return "starts a program with an empty environment (exec " + a + ")"
					}
				}
				k++
				continue
			case listHas(lineRunners, b) || strings.HasPrefix(w, "-") || (len(w) == 2 && w[0] == '/') ||
				(strings.Contains(w, "=") && !strings.HasPrefix(w, "=") && !strings.ContainsAny(w[:strings.IndexByte(w, '=')], "/\\$")):
				k++
				continue
			case b == "env":
				if a := envClears(words[k+1:]); a != "" {
					return "starts a program with an empty environment (env " + a + ")"
				}
			case b == "start-process" || b == "saps" || b == "start":
				for _, a := range words[k+1:] {
					a, _, _ = strings.Cut(strings.Trim(a, `"'`), ":")
					if strings.HasPrefix(a, "-usen") && strings.HasPrefix("-usenewenvironment", a) {
						return "starts a program with a new environment (Start-Process " + a + ")"
					}
				}
			}
			break
		}
	}
	return ""
}

// envClears gives the option of env's words that clears its environment (-i, a bundle holding it, -, or
// --ignore-environment by any unambiguous prefix), "" for none: env's own words end at the program it runs.
func envClears(words []string) string {
	for m := 0; m < len(words); m++ {
		a := strings.Trim(words[m], `"'`)
		long := strings.HasPrefix(a, "--")
		switch {
		case a == "-" || (long && len(a) > 2 && strings.HasPrefix("--ignore-environment", a)):
			return a
		case !long && strings.HasPrefix(a, "-") && strings.ContainsRune(shortOpts(a), 'i'):
			return a
		case a == "-u" || a == "-c" || a == "-s" || a == "-p" || a == "--unset" || a == "--chdir" || a == "--split-string":
			m++ // its value is the next word
		case !strings.HasPrefix(a, "-") && !strings.Contains(a, "="):
			return "" // the program: env's own words end
		}
	}
	return ""
}

// shortOpts gives a bundle of short options' letters (-iu -> iu), stopping at an option that takes a value in the same
// word (env's -u, -C, -S, -P take one).
func shortOpts(a string) string {
	var b strings.Builder
	for _, r := range a[1:] {
		b.WriteRune(r)
		if strings.ContainsRune("ucsp", r) {
			break
		}
	}
	return b.String()
}

// splitRough cuts a line's text at the shell's command boundaries, without reading quotes: a rough cut, enough to
// find a command's own words.
func splitRough(low string) []string {
	return strings.FieldsFunc(low, func(r rune) bool { return strings.ContainsRune(";&|\n\r(){}`", r) })
}
