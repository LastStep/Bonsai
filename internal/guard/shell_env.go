package guard

// CLAUDE_CODE_CHILD_SESSION in a shell call (design/plan-5.md, 5.3.2 note 7; commands.go), read on the delete check's
// words: the splitter's (shell_split.go) with quotes taken off and bash's brace lists written out, each simple command
// reached through the words that run another program (RunnerWords), however deeply nested (bash -c, a here-document fed
// to a shell, -EncodedCommand). A change of the variable, or a program started with an empty or new environment, is a
// person's step (commands.go's personStep): allowed only under a person's grant in command mode, else person-command.
//
// The other agents' markers (format.AgentMarkers; the director, 23 Oct, T-5727's choice C): every exact form below (an
// assignment, a changer's word, env -u, env's assignment, a PowerShell assignment, a provider path naming the whole
// name, a .NET argument) reads each of them as it reads CLAUDE_CODE_CHILD_SESSION, and the refusal names the one it
// found. The catch-all text match (a name held as data, the backstop) stays Claude Code's alone: another marker's name
// inside a word (cd opencode-tests, OPENCODE_HOME) is no change of it. So a pi agent's unset PI_CODING_AGENT is
// refused where Bonsai's guard reads pi's lines.
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
//
// Known limits, fixed after 1.0 (the director, 11 Oct, "Scoped"); the command's own refusal (5.3.6b) is the first layer
// behind them and rung 0 (5.4) the after-the-fact one:
//   - a name changer the check does not read: let, : $((N=0)), read, printf -v, mapfile, for N in, getopts, a nameref;
//   - the name as data: declare v=NAME, a loop, a function argument, a command substitution, xargs -I{};
//   - the name built from parts in a line that also names it whole elsewhere, so the backstop stays quiet;
//   - PowerShell's provider and .NET forms: New-Item -Path Env:, Set-Location Env:, ('Env:'+'NAME'), [string]'NAME',
//     .ToUpper(), Start-Process -Environment @{...}, a ProcessStartInfo;
//   - cmd /c "env -i x.sh";
//   - PowerShell casts, parentheses and comma lists on $env: ([string]$env:NAME = "", ($env:NAME) = "",
//     $env:NAME, $x = "", 1);
//   - env's own -S splitting rules (\_ as env's space, a -S text holding ; read as a command line), as in the
//     delete check's limits (shell.go);
//   - a wolf: the backstop refuses a text that names part of the variable (env | grep CLAUDE_CODE_CHILD, a comment).

import (
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
)

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

// envMarker gives the agent marker (format.AgentMarkers: CLAUDE_CODE_CHILD_SESSION and the other agents') a text is
// the name of, exactly, letter case ignored; "" for none. The exact forms (an assignment, a changer's word, env -u, a
// provider path, .NET's argument) read every marker; the catch-all text match (envSight, envAsData, envBackstop)
// reads Claude Code's alone, so a word that only holds another marker's name (cd opencode-tests) passes (T-5727).
func envMarker(t string) string {
	for _, m := range format.AgentMarkers {
		if strings.EqualFold(t, m.Word) {
			return m.Word
		}
	}
	return ""
}

// envAssigned gives the marker a NAME=value or NAME+=value text assigns, "" for none.
func envAssigned(t string) string {
	name, _, ok := strings.Cut(t, "=")
	if !ok {
		return ""
	}
	return envMarker(strings.TrimSuffix(name, "+"))
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
		switch m := envAssigned(w.text); {
		case m != "":
			r.envStep("assigns " + m)
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
			m := envMarker(a.text)
			if m == "" {
				m = envAssigned(a.text)
			}
			if m != "" {
				r.envStep("unsets, exports or sets " + m + " (" + name + ")")
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
			if m := envProviderPath(w.text); m != "" {
				r.envStep("changes " + m + " through its Env: path")
				return
			}
		}
	}
}

// envProviderPath gives the marker a PowerShell provider path names (Env:NAME, Env:\NAME, Env:/NAME, the provider's
// full name Microsoft.PowerShell.Core\Environment::NAME), not a read of it ($env:NAME, ${env:NAME}); "" for none.
// Claude Code's name counts wherever it stands after the prefix, as before; another marker's only as the whole name,
// so Env:OPENCODE_HOME is not OPENCODE's.
func envProviderPath(t string) string {
	low := strings.ToLower(t)
	for _, m := range format.AgentMarkers {
		name := strings.ToLower(m.Word)
		for from := 0; ; {
			i := strings.Index(low[from:], name)
			if i < 0 {
				break
			}
			i += from
			from = i + len(name)
			if m.Word != ChildSessionEnv && from < len(low) && shNameChar(low[from]) {
				continue
			}
			pre := strings.TrimRight(low[:i], `\/`)
			switch {
			case strings.HasSuffix(pre, "$env:") || strings.HasSuffix(pre, "${env:"):
			case strings.HasSuffix(pre, "env:") || strings.HasSuffix(pre, "environment::"):
				return m.Word
			}
		}
	}
	return ""
}

// psAssign is one PowerShell assignment among a command's words: its left side (lower case, as written: $env:name,
// ${env:name}, $x) and the texts of what it assigns.
type psAssign struct {
	left  string
	right []string
}

// envPSAssigns finds PowerShell's assignments among a command's words, wherever they stand, chained too ($x =
// $env:NAME = 'v', $x=$env:NAME='v', $x =$env:NAME =""): a word whose written form starts with a variable, then an
// assignment operator (=, +=, -=, *=, /=, %=, ??=) glued to it, or starting the next word, alone or glued to what it
// assigns (= "", ="", =$null). What follows the operator, quoted or not, is what it assigns; a word that starts with
// an operator holds the right side of the word before it, and may be a chain's next left side.
func envPSAssigns(words []shWord) []psAssign {
	var out []psAssign
	for k, w := range words {
		raw, text := strings.ToLower(w.raw), w.text
		if op := envPSOperator(raw); op != "" && k > 0 {
			raw = raw[len(op):]
			_, text, _ = strings.Cut(text, "=")
		}
		for strings.HasPrefix(raw, "$") {
			v := envPSVariable(raw)
			if v == "" {
				break
			}
			rest := raw[len(v):]
			if op := envPSOperator(rest); op != "" {
				_, after, _ := strings.Cut(text, "=")
				out = append(out, psAssign{left: v, right: envPSRight(after, words[k+1:])})
				raw, text = rest[len(op):], after
				continue
			}
			if rest == "" && k+1 < len(words) && envPSOperator(strings.ToLower(words[k+1].raw)) != "" {
				_, after, _ := strings.Cut(words[k+1].text, "=")
				out = append(out, psAssign{left: v, right: envPSRight(after, words[k+2:])})
			}
			break
		}
	}
	return out
}

// envPSRight gives an assignment's right side: the text after its operator, then the texts of the words after it.
func envPSRight(after string, words []shWord) []string {
	right := []string{after}
	for _, x := range words {
		right = append(right, x.text)
	}
	return right
}

// envPSVariable gives the variable a written PowerShell word starts with: ${...} up to its brace, or $ and the
// characters of a name (letters, digits, _, and : for a scope or drive such as env:); "" for none.
func envPSVariable(raw string) string {
	if strings.HasPrefix(raw, "${") {
		if end := strings.IndexByte(raw, '}'); end > 0 {
			return raw[:end+1]
		}
		return ""
	}
	k := 1
	for k < len(raw) && (shNameChar(raw[k]) || raw[k] == ':') {
		k++
	}
	if k == 1 {
		return ""
	}
	return raw[:k]
}

// envPSOperator gives the assignment operator a text starts with, "" for none (== is a comparison).
func envPSOperator(t string) string {
	for _, op := range []string{"??=", "+=", "-=", "*=", "/=", "%=", "="} {
		if strings.HasPrefix(t, op) && !strings.HasPrefix(t, "==") {
			return op
		}
	}
	return ""
}

// envPSAssign judges PowerShell's assignments among a command's words, before its left side is dropped
// (delPSAssign): one of the variable ($env:NAME, ${env:NAME}, any letter case, any operator, the value quoted or
// not); one of another variable whose value holds the name as data ($n = 'CLAUDE_CODE_CHILD_SESSION'; Remove-Item
// "Env:$n"), which the check cannot follow.
func (r *delRun) envPSAssign(words []shWord) {
	for _, a := range envPSAssigns(words) {
		if m := envPSLeft(a.left); m != "" {
			r.envStep("assigns " + m)
			return
		}
		for _, t := range a.right {
			if envAsData(t) {
				r.envStep("puts the name " + ChildSessionEnv + " in a variable, which the check cannot follow")
				return
			}
		}
	}
}

// envPSLeft gives the marker an assignment's left side names ($env:NAME, ${env:NAME}, lower case as envPSAssigns
// gives it), "" for none.
func envPSLeft(left string) string {
	if v, ok := strings.CutPrefix(left, "$env:"); ok {
		return envMarker(v)
	}
	if v, ok := strings.CutPrefix(left, "${env:"); ok && strings.HasSuffix(v, "}") {
		return envMarker(strings.TrimSuffix(v, "}"))
	}
	return ""
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
	var walk func(l *shList) string
	walk = func(l *shList) string {
		for k := range l.cmds {
			cmd := &l.cmds[k]
			for _, w := range cmd.words {
				if m := envMarker(strings.Trim(w.text, "()'\" ")); m != "" {
					return m
				}
			}
			for _, n := range cmd.nested {
				if m := walk(n.list); m != "" {
					return m
				}
			}
			if cmd.group != nil {
				if m := walk(cmd.group); m != "" {
					return m
				}
			}
		}
		return ""
	}
	for _, n := range c.nested {
		if m := walk(n.list); m != "" {
			r.envStep("sets " + m + " through .NET")
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
