package guard

// Bonsai's own commands in a shell call (design/plan-5.md, 5.3.2 note 7; spec §4, §7; contract §10.6), under the
// answer to step 5.3's second question, (ii): who may consent to code and take the guard out.
//
//   - bonsai settings set, bonsai labels attach and detach are refused in an agent session, always (home-command):
//     they write this machine's settings and label files, which only a person's own terminal changes (the commands
//     refuse themselves too, step 5.6).
//   - bonsai with --allow-exec (the consent to code), bonsai unlink (which takes the guard out), and a line that unsets
//     or changes an agent marker (format.AgentMarkers, which hide an agent's session from those commands' own
//     refusal) are a person's commands (person-command): allowed to an agent only with a person's grant in a project in command
//     mode, that is, the project's running task's bonsai.allows holds bonsai.yaml and .bonsai/lock.json (rule 3's
//     grant; in command mode only a person's tap puts a person-only path there, tripwire.go).
//   - bonsai init, update and unlink are judged as one edit of bonsai.yaml and .bonsai/lock.json at the top of the
//     checkout they run in (may), so an agent's update needs a grant of both, as any edit of them does.
//   - Every other word is allowed (bonsai-command), and so is a call that writes nothing whatever its word (T-5505's
//     findings, 16 Oct: bonsai message --help and a preview of bonsai update were refused as writes): one holding
//     --help or -h among its words, which prints the word's help and runs nothing (runWord stops at it, and no flag
//     takes it as a value: a flag's value may not start with a dash), and a preview, bonsai init or update with
//     neither --yes nor --allow-exec, its words literal, in an agent's session (CLAUDE_CODE_CHILD_SESSION set, so the
//     command asks no y/N and writes neither bonsai.yaml nor the lock: it prints the preview and exits 4). Such a
//     preview still writes outside the project's files (V1's should-fix 2, 16 Oct): its fetch fills the Bonsai home's
//     cache, and with nothing to change it records this machine's record of the checkout and asks Claude Code to
//     install the locked packs' plugins (claude plugin install), as update's help says; in Bonsai's own repo that is
//     the person's own plugin state. None of it changes bonsai.yaml, the lock or a guarded path.
//
// The command itself refuses too (step 5.3.6 note 8), since a variable, xargs, a script or a subprocess can hide a
// flag from the guard: the guard is the second layer. ownCommand judges one bonsai call's words, found by step 5.3.3's
// splitter whatever folder bonsai is called from (through sudo, env, bash -c, find -exec and the rest); a bonsai call
// xargs runs whose own words leave it open to what xargs's input adds (ownOpen, or a word its input or a variable
// stands for), and a line the splitter cannot split that holds a refused bonsai call, are judged as a person's step
// (personStep). What changes the variable is read on the splitter's words, not the line's
// text (shell_env.go), and judged here by personStep too.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// ChildSessionEnv names the variable Claude Code sets in every process its tools and hooks start (its env-vars
// reference), by which Bonsai's commands tell an agent's session from a person's terminal (step 5.3.6 note 8).
const ChildSessionEnv = "CLAUDE_CODE_CHILD_SESSION"

// personNext is the next step of a person's command refused to an agent.
const personNext = "a person runs this command in a terminal of their own, not one Claude Code started"

// ownHome are Bonsai's commands that write this machine's settings or label files, each with the subcommands that do
// (home-command); ownWrites are those judged as an edit of bonsai.yaml and .bonsai/lock.json: init and update, which
// take the consent flag --allow-exec, and unlink, which takes the guard out. These lists' one home: ownCommand reads
// a bonsai call by them, and ownOpen a call whose last words xargs gives.
var (
	ownHome = []struct {
		word string
		subs []string
	}{{"settings", []string{"set"}}, {"labels", []string{"attach", "detach"}}}
	ownWrites = []string{"init", "update", "unlink"}
)

// ownWord gives a bonsai call's command word, lower case (the first of its words after the program that is not an
// option), and its index in args; "" and -1 for none.
func ownWord(args []string) (string, int) {
	for k, a := range args {
		if !strings.HasPrefix(a, "-") {
			return strings.ToLower(a), k
		}
	}
	return "", -1
}

// ownSubs gives the subcommands of a command word that write this machine's files (ownHome), nil for none.
func ownSubs(word string) []string {
	for _, h := range ownHome {
		if h.word == word {
			return h.subs
		}
	}
	return nil
}

// ownOpen says why words appended to a bonsai call's own, as xargs appends its input, could still make it a person's
// or a home command, "" when its own words settle a command that is neither: argv as ownCommand takes it. Open are a
// call with no command word, a command of ownHome with no subcommand given, and a command of ownWrites (the consent
// flag may follow it).
func ownOpen(argv []string) string {
	word, k := ownWord(argv[1:])
	switch {
	case word == "":
		return "with no command word of its own"
	case listHas(ownWrites, word):
		return "with the command " + oneLine(word) + ", which its input may give --allow-exec"
	case ownSubs(word) != nil:
		if sub, _ := ownWord(argv[2+k:]); sub == "" {
			return "with the command " + oneLine(word) + " and no subcommand of its own"
		}
	}
	return ""
}

// ownCommand judges one bonsai call: argv its words as the shell gives them to the program (argv[0] the program),
// dir the folder it runs in (absolute; the session's project folder when it cannot be told). Its record goes in the
// session's project (rule 8: a shell call is judged there).
func (j *judge) ownCommand(dir string, argv []string, literal bool) Decision {
	if len(argv) < 2 || helpArg(argv[1:]) {
		return allow(RuleBonsaiCommand, "")
	}
	args := argv[1:]
	word, _ := ownWord(args)
	if j.preview(word, args, literal) {
		return allow(RuleBonsaiCommand, "")
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
	if rs := j.sh; rs != nil {
		if d := j.roleCommand(rs, word, argv); !d.Allow {
			return d
		}
	}
	switch subs := ownSubs(word); {
	case subs != nil && later(subs...):
		return deny(RuleHomeCommand, fmt.Sprintf("%s refused: it writes this machine's settings or label files in Bonsai's "+
			"home, which never change from an agent's session", shown), personNext)
	case !listHas(ownWrites, word) && !allowExec(args):
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
	if listHas(ownWrites, word) {
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

// helpArg reports --help or -h among a bonsai call's words (as the shell gives them, so literal): the command prints
// its help and runs nothing (cmd/bonsai's runWord and runSub; bonsai alone with more words refuses), so the call writes
// nothing whatever its other words say.
func helpArg(args []string) bool {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return true
		}
	}
	return false
}

// preview reports a bonsai init or update that only previews: no --yes (in any form) and no --allow-exec among its
// words, every word literal (a variable could stand for --yes), in an agent's session (CLAUDE_CODE_CHILD_SESSION set in
// the hook's environment, so the command has no terminal to ask y/N at: cmd/bonsai's interactive). It prints the
// preview and writes neither bonsai.yaml nor the lock (a line that changes the variable is a person's step, judged
// apart: shell_env.go).
func (j *judge) preview(word string, args []string, literal bool) bool {
	if word != "init" && word != "update" || !literal || allowExec(args) || j.getenv == nil || j.getenv(ChildSessionEnv) == "" {
		return false
	}
	for _, a := range args {
		if strings.HasPrefix(strings.ToLower(strings.TrimLeft(a, "-")), "yes") && strings.HasPrefix(a, "-") {
			return false
		}
	}
	return true
}

// roleCommand judges a bonsai call for the roles' rules (rolecmds.go): bonsai message's --from and --kind brief; bonsai
// init, update, unlink and check --write refused to a role that changes nothing; and bonsai ladder --task <id> noted for
// the agent's red runs once the call is allowed.
func (j *judge) roleCommand(rs *shellRole, word string, argv []string) Decision {
	args := argv[1:]
	switch {
	case word == "message":
		return rs.messageCall(argv)
	case rs.changesNothing() && (listHas(ownWrites, word) || word == "check" && hasArg(args, "--write")):
		return deny(RuleRoleWritesNothing, fmt.Sprintf("bonsai %s refused: it writes the project's files, and %s changes nothing "+
			"(its role's writes: false)", oneLine(word), rs.w.shown()), "put what you would change in your report")
	case word == "ladder" && rs.w.counted:
		note := func(task string) {
			if validTask(task) && !listHas(rs.rr.climbs, task) {
				rs.rr.climbs = append(rs.rr.climbs, task)
			}
		}
		literal := false
		for i, a := range args {
			task, named := "", false
			switch {
			case a == "--task" && i+1 < len(args):
				task, named = args[i+1], true
			case strings.HasPrefix(a, "--task="):
				task, named = strings.TrimPrefix(a, "--task="), true
			}
			if named && validTask(task) {
				literal = true
				note(task)
			}
		}
		if !literal {
			// No --task the guard can read (none, or one in a variable): the ladder climbs for BONSAI_TASK, as it reads
			// it (VG's item E): an assignment of it before the command on this line, else the session's own.
			task := rs.taskAssign
			if task == "" {
				task = j.getenv(workspace.TaskEnv)
			}
			note(task)
		}
	}
	return allow(RuleBonsaiCommand, "")
}

// hasArg reports a word among a call's words, its = form too.
func hasArg(args []string, name string) bool {
	for _, a := range args {
		if a == name || strings.HasPrefix(a, name+"=") {
			return true
		}
	}
	return false
}

// validTask reports a task id the ladder's result can be named by (T- and digits).
func validTask(t string) bool {
	if !strings.HasPrefix(t, "T-") || len(t) < 3 || len(t) > 12 {
		return false
	}
	for _, c := range t[2:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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

// PersonGrant is personGrant for bonsai init, update and unlink themselves (step 5.3.6 note 8: the command refuses a
// person's command in an agent's session too, since a variable, xargs, a script or a subprocess can hide a flag from
// the guard), so the judgment of a person's grant has one home: "" when the checkout holding dir is a project in
// command mode whose running task grants bonsai.yaml and .bonsai/lock.json, else the reason, ASCII words that follow
// "and". home is the Bonsai home; getenv the command's environment. The session's own project, for BONSAI_TASK, is
// CLAUDE_PROJECT_DIR's when it is set, as in the guard; else the project holding dir, since Claude Code gives a tool's
// commands no CLAUDE_PROJECT_DIR (measured at step 5.3.6a) and the command's checkout stands for the session's.
func PersonGrant(home, dir string, getenv func(string) string) string {
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	j := newJudge(home)
	j.getenv = getenv
	at := getenv(hookin.ProjectEnv)
	if at == "" {
		at = dir
	}
	if roots, _, err := hookin.FindProjects(at); err == nil && len(roots) > 0 {
		j.session = j.project(roots[0])
	}
	return j.personGrant(checkoutTop(dir))
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
