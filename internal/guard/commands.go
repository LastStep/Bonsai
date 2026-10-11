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
// splitter whatever folder bonsai is called from (through sudo, env, bash -c, find -exec and the rest); a bonsai call
// whose words come from xargs's input, and a line the splitter cannot split that holds a refused bonsai call, are
// judged as a person's step (personStep). What changes the variable is read on the splitter's words, not the line's
// text (shell_env.go), and judged here by personStep too.

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
	shown := "bonsai " + oneLine(word) // the word alone: the call's other words may name paths or values (contract §2.6)
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
			d, _ := j.may(change{tool: "bonsai " + word, path: filepath.Join(top, filepath.FromSlash(rel)), dir: top, base: top})
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
// .git git takes as one (gitEntry); dir itself when none does.
func checkoutTop(dir string) string {
	dir = filepath.Clean(dir)
	for d := dir; ; {
		if gitEntry(filepath.Join(d, ".git")) {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			return dir
		}
		d = parent
	}
}

// gitEntry reports a .git git itself takes as one: a file naming the repository (gitdir: ...), or a folder holding HEAD,
// objects and refs. Any other .git (an empty folder, a stray file) does not end the walk up to a checkout's top.
func gitEntry(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	if !fi.IsDir() {
		b := make([]byte, 8)
		f, err := os.Open(p)
		if err != nil {
			return false
		}
		n, _ := f.Read(b)
		_ = f.Close()
		return string(b[:n]) == "gitdir: "
	}
	for _, need := range []string{"HEAD", "objects", "refs"} {
		if _, err := os.Stat(filepath.Join(p, need)); err != nil {
			return false
		}
	}
	return true
}

// personStep judges a person's step the delete check found in a shell call (shell_env.go: a change of
// CLAUDE_CODE_CHILD_SESSION, a program started with an empty or new environment; a bonsai call whose words the check
// cannot see): allowed only with a person's grant in the session's project in command mode, else person-command. what
// says what the line does, in ASCII words that follow "the command line" (the line itself is never quoted).
func (j *judge) personStep(tool, what string) Decision {
	why := "the session is in no linked project"
	if j.session != nil {
		why = j.personGrant(j.session.root)
	}
	if why == "" {
		return allow(RuleShell, "")
	}
	return deny(RulePersonCommand, fmt.Sprintf("%s refused: the command line %s, which hides an agent's session or its "+
		"words from Bonsai's commands; that is a person's step, allowed to an agent only in a project in command mode "+
		"under a person's grant of bonsai.yaml and .bonsai/lock.json, and %s", tool, what, why), personNext)
}
