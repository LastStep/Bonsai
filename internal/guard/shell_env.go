package guard

// CLAUDE_CODE_CHILD_SESSION in a shell call (design/plan-5.md, 5.3.2 note 7; commands.go), read on the delete check's
// words: the splitter's (shell_split.go) with quotes taken off and bash's brace lists written out, each simple command
// reached through the words that run another program (RunnerWords), however deeply nested (bash -c, a here-document fed
// to a shell, -EncodedCommand). A change of the variable, or a program started with an empty or new environment, is a
// person's step (commands.go's personStep): allowed only under a person's grant in command mode, else person-command.
//
// What changes it, letter case of the name ignored as Windows ignores it:
//   - an assignment of it: NAME=... before a command (bash), $env:NAME = ... (PowerShell), set NAME=... (cmd.exe);
//   - its name among the words of a command that sets or unsets variables (envChangers: unset, export, declare, ...);
//   - env -u NAME, an assignment of it env runs a program with, and env -i, env - or --ignore-environment; exec -c;
//     Start-Process -UseNewEnvironment (any prefix PowerShell takes);
//   - its PowerShell provider path (Env:NAME, the Environment:: provider) in a command that only reads is allowed
//     (envReaders); in any other command, a change;
//   - a .NET SetEnvironmentVariable call whose arguments name it, read after the quotes and brackets are taken off.
//
// A read is allowed: $NAME, ${NAME}, %NAME%, $env:NAME in an expression, and the name as another program's argument
// (printenv NAME, grep NAME). What the words cannot show fails closed (envBackstop): a text that holds
// CLAUDE_CODE_CHILD in any quoting (its quote and escape characters taken out) but in no word the check read, the name
// built from parts that way, or a text the splitter cannot split; a comment that names it (# CLAUDE_CODE_CHILD_SESSION)
// is no word, so its line fails closed too, the price of reading no raw text. Not seen, as before: the name built where no part of
// it is in the line, an interpreter's own calls (python -c, node -e), a new session that starts without the variable
// (ssh, su, a terminal multiplexer, a scheduler); the commands' own refusal and the Desk stand behind them.

import "strings"

// envName is the variable's name, lower case, as the words are compared with it.
var envName = strings.ToLower(ChildSessionEnv)

// envChangers are the commands that set or unset a shell's variables by name (bash's, and cmd.exe's set and setx),
// by base name, letter case ignored; this list's one home.
var envChangers = []string{"unset", "export", "declare", "typeset", "local", "readonly", "set", "setx"}

// envReaders are the PowerShell commands that only read an Env: provider path; this list's one home.
var envReaders = []string{"get-item", "gi", "get-childitem", "gci", "ls", "dir", "test-path", "get-content", "gc", "cat",
	"type", "get-itemproperty", "gp", "resolve-path", "rvpa", "get-itempropertyvalue", "gpv"}

// envNewEnv are PowerShell's words for Start-Process, which -UseNewEnvironment starts with a new environment.
var envNewEnv = []string{"start-process", "saps", "start"}

// envStep refuses (unless a person's grant allows) a change of the variable: what says what the line does.
func (r *delRun) envStep(what string) {
	if d := r.j.personStep(r.tool, what); !d.Allow {
		r.refuse(d)
	}
}

// envIs reports a text that is the variable's name, letter case ignored.
func envIs(t string) bool { return strings.EqualFold(t, ChildSessionEnv) }

// envAssigns reports a NAME=value or NAME+=value text that assigns the variable.
func envAssigns(t string) bool {
	name, _, ok := strings.Cut(t, "=")
	return ok && envIs(strings.TrimSuffix(name, "+"))
}

// envSight counts the words that name the variable, for the backstop: a word whose text holds the whole name.
func envSight(t string) bool { return strings.Contains(strings.ToLower(t), envName) }

// envAssignWords judges the NAME=value words before a command's program (bash's, and env's): an assignment of the
// variable, and one whose value holds its name as data (X=CLAUDE_CODE_CHILD_SESSION; unset $X), which the check cannot
// follow.
func (r *delRun) envAssignWords(words []shWord) {
	for _, w := range words {
		if !w.assign {
			return
		}
		_, value, _ := strings.Cut(w.text, "=")
		switch {
		case envAssigns(w.text):
			r.envStep("assigns " + ChildSessionEnv)
			return
		case envAsData(value):
			r.envStep("puts the name " + ChildSessionEnv + " in a variable, which the check cannot follow")
			return
		}
	}
}

// envAsData reports a text that holds the variable's name other than as a read of it ($NAME, ${NAME}, $env:NAME,
// ${env:NAME}, %NAME%).
func envAsData(t string) bool {
	low := strings.ToLower(t)
	for from := 0; ; {
		i := strings.Index(low[from:], envName)
		if i < 0 {
			return false
		}
		i += from
		from = i + len(envName)
		pre := low[:i]
		if !(strings.HasSuffix(pre, "$") || strings.HasSuffix(pre, "${") || strings.HasSuffix(pre, "$env:") ||
			strings.HasSuffix(pre, "${env:") || strings.HasSuffix(pre, "%")) {
			return true
		}
	}
}

// envCommand judges one command's words from its program on, in its shell: a command that sets or unsets variables,
// PowerShell's provider path and Start-Process -UseNewEnvironment, cmd.exe's set.
func (r *delRun) envCommand(x delExec, name string) {
	words := x.words
	args := words[1:]
	switch {
	case listHas(envChangers, name):
		for _, a := range args {
			if envIs(a.text) || envAssigns(a.text) {
				r.envStep("unsets, exports or sets " + ChildSessionEnv + " (" + name + ")")
				return
			}
		}
	case x.d == shPower && listHas(envNewEnv, name):
		for _, a := range args {
			if !a.param {
				continue
			}
			p, _, _ := strings.Cut(strings.ToLower(a.text), ":")
			if len(p) >= len("-usen") && strings.HasPrefix("-usenewenvironment", p) {
				r.envStep("starts a program with a new environment (Start-Process " + oneLine(p) + ")")
				return
			}
		}
	}
	if x.d == shPower && !listHas(envReaders, name) {
		for _, w := range words {
			if envProviderPath(w.text) {
				r.envStep("changes " + ChildSessionEnv + " through its Env: path")
				return
			}
		}
	}
}

// envProviderPath reports a PowerShell provider path to the variable (Env:NAME, Env:\NAME, Env:/NAME, the provider's
// full name Microsoft.PowerShell.Core\Environment::NAME), not a read of it ($env:NAME, ${env:NAME}).
func envProviderPath(t string) bool {
	low := strings.ToLower(t)
	for from := 0; ; {
		i := strings.Index(low[from:], envName)
		if i < 0 {
			return false
		}
		i += from
		from = i + len(envName)
		pre := strings.TrimRight(low[:i], `\/`)
		switch {
		case strings.HasSuffix(pre, "$env:") || strings.HasSuffix(pre, "${env:"):
		case strings.HasSuffix(pre, "env:") || strings.HasSuffix(pre, "environment::"):
			return true
		}
	}
}

// envPSAssign judges PowerShell's $env:NAME = ... (with any assignment operator, the = written with or apart from
// the name), before the assignment's left side is dropped (delPSAssign).
func (r *delRun) envPSAssign(words []shWord) {
	if len(words) == 0 || words[0].quoted {
		return
	}
	t := strings.ToLower(words[0].text)
	left, _, glued := strings.Cut(t, "=")
	left = strings.TrimRight(left, "+-*/%?")
	if left != "$env:"+envName && left != "${env:"+envName+"}" {
		return
	}
	if glued || len(words) >= 2 && strings.HasSuffix(words[1].text, "=") {
		r.envStep("assigns " + ChildSessionEnv)
	}
}

// envPSData judges PowerShell's $x = ... whose right side holds the variable's name as data ($n =
// 'CLAUDE_CODE_CHILD_SESSION'; Remove-Item "Env:$n"), which the check cannot follow.
func (r *delRun) envPSData(words []shWord) {
	if len(words) < 2 || words[0].quoted || !strings.HasPrefix(words[0].text, "$") {
		return
	}
	k := 1
	if strings.Contains(words[0].text, "=") {
		k = 0
	} else if !strings.HasSuffix(words[1].text, "=") {
		return
	}
	for _, w := range words[k:] {
		t := w.text
		if k == 0 {
			_, t, _ = strings.Cut(t, "=")
			k = -1
		}
		if envAsData(t) {
			r.envStep("puts the name " + ChildSessionEnv + " in a variable, which the check cannot follow")
			return
		}
	}
}

// envDotNet judges a .NET SetEnvironmentVariable call: its arguments, in the command's nested lists, read after their
// quotes and brackets are taken off.
func (r *delRun) envDotNet(c *shCommand) {
	called := false
	for _, w := range c.words {
		if strings.Contains(strings.ToLower(w.text), "setenvironmentvariable") {
			called = true
		}
	}
	if !called {
		return
	}
	var walk func(l *shList) bool
	walk = func(l *shList) bool {
		for k := range l.cmds {
			cmd := &l.cmds[k]
			for _, w := range cmd.words {
				if envIs(strings.Trim(w.text, "()'\" ")) {
					return true
				}
			}
			for _, n := range cmd.nested {
				if walk(n.list) {
					return true
				}
			}
			if cmd.group != nil && walk(cmd.group) {
				return true
			}
		}
		return false
	}
	for _, n := range c.nested {
		if walk(n.list) {
			r.envStep("sets " + ChildSessionEnv + " through .NET")
			return
		}
	}
}

// envBackstop says whether a text holds CLAUDE_CODE_CHILD in any quoting: with its quote and escape characters, and
// line continuations, taken out.
func envBackstop(text string) bool {
	low := strings.ToLower(text)
	var b strings.Builder
	for k := 0; k < len(low); k++ {
		switch c := low[k]; c {
		case '\'', '"', '`', '\\', '^', '\n', '\r', '{', '}', ',':
		default:
			b.WriteByte(c)
		}
	}
	return strings.Contains(b.String(), "claude_code_child")
}

// envSightings counts the words and fed texts of a list, at every depth, that name the variable whole.
func envSightings(l *shList) int {
	n := 0
	for k := range l.cmds {
		c := &l.cmds[k]
		for _, w := range c.words {
			texts := []string{w.text}
			if w.alts != nil {
				texts = append(texts, w.alts...)
			}
			for _, t := range texts {
				if envSight(t) {
					n++
				}
			}
		}
		for _, f := range c.feeds {
			if envSight(f.text) {
				n++
			}
			for _, nl := range f.nested {
				n += envSightings(nl.list)
			}
		}
		for _, nl := range c.nested {
			n += envSightings(nl.list)
		}
		if c.group != nil {
			n += envSightings(c.group)
		}
	}
	return n
}
