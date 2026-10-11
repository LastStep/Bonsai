package guard

// The delete check's words (plan-5 5.3.3 notes 1 and 2): the delete words and how each is read (DeleteWords), the
// words that run another program (RunnerWords), each table its list's one home and on the reference page
// (docs/reference/lists.md, go generate ./...); the shells whose text is split again, the folder words (cd and its
// kin), and Bonsai's own program, whose calls go to commands.go's ownCommand.
//
// A program is matched by its base name, case-folded, with .exe dropped (/bin/rm, RM.EXE, C:\...\robocopy.exe). In
// PowerShell a bare rm, del or mv is the cmdlet's alias; a program named with a folder or .exe is the program
// itself. Options are read as the program reads them: GNU programs take short options bundled in any order (-rf,
// -fR) and long ones by any unambiguous prefix (--rec); PowerShell takes a parameter by any unambiguous prefix,
// case-blind (-r, -Rec), and an ambiguous one counts as each it may be (fail closed); cmd.exe takes /x options,
// glued or not (rd/s/q).

import (
	"encoding/base64"
	"fmt"
	"path"
	"runtime"
	"strings"
)

// Where a word is read.
const (
	delInBash = 1 << iota // by bash; for a delete word, as a program from PowerShell or cmd.exe too
	delInPS               // by PowerShell, as a cmdlet or its alias
	delInCmd              // by cmd.exe, as its own command
)

// DeleteWord is one delete word.
type DeleteWord struct {
	Word    string   // as the reference page shows it, e.g. "git clean"
	Names   []string // the program names it answers to: base names, case-folded, .exe dropped ("git" for git's words)
	Sub     string   // git's subcommand, for git's words; "" for others
	ByName  bool     // its name alone makes it a delete (an unsplittable line holding it is refused)
	Deletes string   // what it deletes or moves, and what makes it recursive or bulk: the reference page's words
	in      int      // where it is read (delInBash, delInPS, delInCmd)
}

// Shells names where a delete word is read, for the reference page: a program's name in every shell (in PowerShell
// a bare name that is also a cmdlet's alias is the alias), a cmdlet in PowerShell, cmd.exe's own command in cmd.exe.
func (w DeleteWord) Shells() string {
	switch {
	case w.in&delInBash != 0:
		return "every shell, as a program"
	case w.in&delInPS != 0:
		return "PowerShell, as a cmdlet or its alias"
	}
	return "cmd.exe, as its own command"
}

// DeleteWords is the delete check's table of delete words (plan-5 5.3.3 note 2), its one home.
var DeleteWords = []DeleteWord{
	{Word: "rm", Names: []string{"rm"}, in: delInBash, ByName: true,
		Deletes: "its operands; recursive with -r, -R or --recursive (bundled, any prefix); always refused with --no-preserve-root"},
	{Word: "rmdir", Names: []string{"rmdir"}, in: delInBash, ByName: true,
		Deletes: "its operands (empty folders)"},
	{Word: "unlink", Names: []string{"unlink"}, in: delInBash, ByName: true,
		Deletes: "its operand"},
	{Word: "shred", Names: []string{"shred"}, in: delInBash, ByName: true,
		Deletes: "its operands (overwritten, and deleted with -u)"},
	{Word: "mv", Names: []string{"mv"}, in: delInBash, ByName: true,
		Deletes: "moves its sources away (with all they hold) onto its destination"},
	{Word: "git rm", Names: []string{"git"}, Sub: "rm", in: delInBash, ByName: false,
		Deletes: "its paths; recursive with -r; :/x and :(top)x at the checkout's top, other pathspec magic unnamed; its paths from --pathspec-from-file are unnamed; not with -n"},
	{Word: "git clean", Names: []string{"git"}, Sub: "clean", in: delInBash,
		Deletes: "the untracked files under its paths (bulk); with no path unnamed; pathspec magic read as git rm's; always refused with -x or -X; not with -n"},
	{Word: "git stash", Names: []string{"git"}, Sub: "stash", in: delInBash,
		Deletes: "with --all or -a, ignored files moved away (.bonsai/local/ among them): always refused"},
	{Word: "git mv", Names: []string{"git"}, Sub: "mv", in: delInBash,
		Deletes: "moves its sources away onto its destination; pathspec magic read as git rm's; not with -n"},
	{Word: "find", Names: []string{"find"}, in: delInBash,
		Deletes: "with -delete, or -exec, -execdir, -ok or -okdir running a delete word: what it finds under its starting points (bulk)"},
	{Word: "rsync", Names: []string{"rsync"}, in: delInBash,
		Deletes: "with any --delete option (--del): files in its destination (bulk); with --remove-source-files: moves its sources"},
	{Word: "robocopy", Names: []string{"robocopy"}, in: delInBash,
		Deletes: "with /MIR or /PURGE: files in its destination (bulk); with /MOV or /MOVE: moves its source; not with /L"},
	{Word: "Remove-Item", Names: []string{"remove-item", "rm", "ri", "del", "erase", "rd", "rmdir"}, in: delInPS, ByName: true,
		Deletes: "its -Path or -LiteralPath (-LiteralPath names it as written); recursive with -Recurse (any prefix); bulk when piped into or with -Include, -Exclude or -Filter; not with -WhatIf"},
	{Word: "Move-Item", Names: []string{"move-item", "mv", "mi", "move"}, in: delInPS, ByName: true,
		Deletes: "moves its -Path or -LiteralPath onto its -Destination; piped into, unnamed"},
	{Word: "Rename-Item", Names: []string{"rename-item", "ren", "rni"}, in: delInPS, ByName: true,
		Deletes: "moves its -Path or -LiteralPath onto -NewName beside it; piped into, unnamed"},
	{Word: "del", Names: []string{"del", "erase"}, in: delInCmd, ByName: true,
		Deletes: "the files in its operands; recursive with /s"},
	{Word: "rd", Names: []string{"rd", "rmdir"}, in: delInCmd, ByName: true,
		Deletes: "its folders; recursive with /s"},
	{Word: "move", Names: []string{"move"}, in: delInCmd, ByName: true,
		Deletes: "moves its sources onto its destination"},
	{Word: "ren", Names: []string{"ren", "rename"}, in: delInCmd, ByName: true,
		Deletes: "moves its source onto a new name beside it"},
}

// delSplitWords splits env -S's text into words, as one command's (bash's quoting): ok is false when it does not
// split into one simple command.
func delSplitWords(text string, depth int) ([]shWord, bool) {
	l, why := shSplit(text, shBash, depth+1)
	if why != "" || len(l.cmds) != 1 || l.cmds[0].group != nil || len(l.cmds[0].nested) > 0 {
		return nil, false
	}
	return delExpandAlts(l.cmds[0].words), true
}

// delEnvSplit reads env's first -S (--split-string), the option g.on[k] of words (env's words after its name, env),
// as GNU env does: its text split into words that take the place of the option and its value, ahead of the words
// after them, for env's options to be read again. The words are rebuilt from delParse's reading, however the -S was
// written (bundled after other options as in -vS, its value glued as in -Srm, long, or abbreviated as in --spl), with
// the options before it kept for their effect; a -S in its text or after it is read on the next pass, one at a time.
// It gives env's words anew, nil when the -S has no text (env runs nothing); ok is false when the text does not split
// into one simple command.
func delEnvSplit(env shWord, words []shWord, g delGot, k, depth int) ([]shWord, bool) {
	vals := g.vals[g.on[k]]
	if len(vals) == 0 {
		return nil, true
	}
	more, ok := delSplitWords(vals[0].text, depth) // the first -S's text: the first value of its name
	if !ok {
		return nil, false
	}
	at := g.at[k]
	out := append([]shWord{env}, words[:at]...)
	if !strings.HasPrefix(words[at].text, "--") {
		// The short options bundled before the S (-vS): each is an option of its own, its letter as delParse read it.
		var b strings.Builder
		for j := 0; j < k; j++ {
			if g.at[j] == at {
				b.WriteString(g.on[j])
			}
		}
		if b.Len() > 0 {
			out = append(out, shWord{text: "-" + b.String(), raw: "-" + b.String()})
		}
	}
	return append(append(out, more...), words[g.end[k]:]...), true
}

// delParseWord gives the reading of a delete word of the table (git's are read by delRun.git).
func delParseWord(word string) func(r *delRun, x delExec, words []shWord) []delAct {
	switch word {
	case "rm":
		return delRm
	case "rmdir":
		return delRmdir
	case "unlink":
		return delUnlink
	case "shred":
		return delShred
	case "mv":
		return delMv
	case "find":
		return delFind
	case "rsync":
		return delRsync
	case "robocopy":
		return delRobocopy
	case "Remove-Item":
		return delRemoveItem
	case "Move-Item":
		return delMoveItem
	case "Rename-Item":
		return delRenameItem
	case "del":
		return delCmdDel
	case "rd":
		return delCmdRd
	case "move":
		return delCmdMove
	case "ren":
		return delCmdRen
	}
	return func(*delRun, delExec, []shWord) []delAct { return nil }
}

// RunnerWord is a word that runs another program (plan-5 5.3.3 note 1): the check skips it, its options and their
// values, to reach the program it runs.
type RunnerWord struct {
	Word string   // as the reference page shows it
	Does string   // what the check reads in it: the reference page's words
	in   int      // where it is read
	opt  delOpt   // its options (options stop at the program it runs)
	dir  []string // the options naming the folder the program runs in
}

// Shells names where a runner word is read, for the reference page.
func (w RunnerWord) Shells() string { return delShells(w.in) }

// delShells names the shells of a where-read mask.
func delShells(in int) string {
	var out []string
	if in&delInBash != 0 {
		out = append(out, "bash")
	}
	if in&delInPS != 0 {
		out = append(out, "PowerShell")
	}
	if in&delInCmd != 0 {
		out = append(out, "cmd.exe")
	}
	return strings.Join(out, ", ")
}

// RunnerWords is the delete check's table of words that run another program, its one home.
var RunnerWords = []RunnerWord{
	{Word: "sudo", in: delInBash, Does: "its options skipped; -D (--chdir) names the folder",
		opt: delOpt{short: "AbEeHiKklnPSsVv", shortVal: "ugpChDrtTUR", long: []string{"askpass", "background", "edit",
			"help", "login", "remove-timestamp", "reset-timestamp", "list", "non-interactive", "preserve-groups", "shell",
			"stdin", "validate", "version"}, longVal: []string{"user", "group", "host", "prompt", "close-from", "chdir",
			"chroot", "role", "type", "command-timeout", "other-user"}, longOpt: []string{"preserve-env"}},
		dir: []string{"D", "chdir"}},
	{Word: "doas", in: delInBash, Does: "its options skipped",
		opt: delOpt{short: "nLs", shortVal: "uC"}},
	{Word: "env", in: delInBash, Does: "its options and NAME=value words skipped; -C (--chdir) names the folder; -S (--split-string) is split again",
		opt: delOpt{short: "i0v", shortVal: "uCSa", long: []string{"ignore-environment", "null", "debug", "help", "version",
			"list-signal-handling"}, longVal: []string{"unset", "chdir", "split-string", "argv0"},
			longOpt: []string{"ignore-signal", "default-signal", "block-signal"}},
		dir: []string{"C", "chdir"}},
	{Word: "command", in: delInBash, Does: "skipped; with -v or -V it runs nothing",
		opt: delOpt{short: "pvV"}},
	{Word: "builtin", in: delInBash, Does: "skipped"},
	{Word: "exec", in: delInBash, Does: "its options skipped",
		opt: delOpt{short: "cl", shortVal: "a"}},
	{Word: "nice", in: delInBash, Does: "its options skipped",
		opt: delOpt{shortVal: "n", longVal: []string{"adjustment"}}},
	{Word: "ionice", in: delInBash, Does: "its options skipped",
		opt: delOpt{short: "t", shortVal: "cnpPu"}},
	{Word: "nohup", in: delInBash, Does: "skipped"},
	{Word: "setsid", in: delInBash, Does: "its options skipped",
		opt: delOpt{short: "cfw", long: []string{"ctty", "fork", "wait"}}},
	{Word: "stdbuf", in: delInBash, Does: "its options skipped",
		opt: delOpt{shortVal: "ioe", longVal: []string{"input", "output", "error"}}},
	{Word: "time", in: delInBash, Does: "its options skipped",
		opt: delOpt{short: "pvaq", shortVal: "fo", long: []string{"portability", "append", "verbose", "quiet", "help",
			"version"}, longVal: []string{"format", "output"}}},
	{Word: "timeout", in: delInBash, Does: "its options and its duration skipped",
		opt: delOpt{short: "v", shortVal: "sk", long: []string{"preserve-status", "foreground", "verbose"},
			longVal: []string{"signal", "kill-after"}}},
	{Word: "xargs", in: delInBash, Does: "its options skipped; a delete it runs takes its operands from input, so it does not name them; " +
		"a bonsai call it runs is a person's step unless its own words settle a command that is not a person's",
		opt: delOpt{short: "0prtxo", shortOpt: "eil", shortVal: "adEILnPs", long: []string{"null", "no-run-if-empty",
			"interactive", "verbose", "exit", "open-tty", "show-limits"}, longVal: []string{"arg-file", "delimiter",
			"max-args", "max-procs", "max-chars", "process-slot-var"}, longOpt: []string{"eof", "replace", "max-lines"}}},
	{Word: "busybox", in: delInBash, Does: "skipped: the word after it is the program"},
	{Word: "chronic", in: delInBash, Does: "skipped"},
	{Word: "unbuffer", in: delInBash, Does: "skipped"},
	{Word: "fakeroot", in: delInBash, Does: "its options skipped",
		opt: delOpt{shortVal: "lsi", long: []string{"unknown-is-real", "version", "help"}, longVal: []string{"lib", "faked"}}},
	{Word: "watch", in: delInBash, Does: "its options skipped; its words joined and split again, as sh -c runs them, unless -x (--exec)",
		opt: delOpt{short: "bcdegptwx", shortVal: "nq", long: []string{"beep", "color", "no-color", "errexit", "chgexit",
			"precise", "no-title", "no-wrap", "exec", "help", "version"}, longVal: []string{"interval", "equexit"},
			longOpt: []string{"differences"}}},
	{Word: "wsl", in: delInBash | delInPS | delInCmd, Does: "its options skipped; the program runs in bash, its words joined and split again unless -e (--exec); --cd names the folder",
		opt: delOpt{short: "e", shortVal: "du", long: []string{"exec", "system"}, longVal: []string{"distribution",
			"user", "cd", "shell-type"}},
		dir: []string{"cd"}},
	{Word: "start", in: delInCmd, Does: "its /options and a quoted title skipped; /d names the folder"},
	{Word: "call", in: delInCmd, Does: "skipped"},
}

// delBashNames are the programs read as bash-like shells: their -c text, or their input, is split again as bash.
var delBashNames = map[string]bool{"bash": true, "sh": true, "zsh": true, "dash": true, "ksh": true, "ash": true,
	"mksh": true, "yash": true, "rbash": true, "fish": true}

// delByName holds every name of a delete word that is a delete by its name alone.
var delByName = func() map[string]bool {
	m := map[string]bool{}
	for _, w := range DeleteWords {
		if w.ByName {
			for _, n := range w.Names {
				m[n] = true
			}
		}
	}
	return m
}()

// delWordFor finds the delete word a program name is, where it is read; nil for none. git's words are found by
// their subcommand (delRun.git).
func delWordFor(name string, d shDialect, native bool) *DeleteWord {
	pick := func(in int) *DeleteWord {
		for k := range DeleteWords {
			w := &DeleteWords[k]
			if w.in&in == 0 || w.Sub != "" {
				continue
			}
			for _, n := range w.Names {
				if n == name {
					return w
				}
			}
		}
		return nil
	}
	switch {
	case d == shPower && !native:
		if w := pick(delInPS); w != nil {
			return w
		}
	case d == shCmd && !native:
		if w := pick(delInCmd); w != nil {
			return w
		}
	}
	return pick(delInBash)
}

// delRunnerFor finds the runner a program name is, where it is read; nil for none.
func delRunnerFor(name string, d shDialect) *RunnerWord {
	in := delInBash
	switch d {
	case shPower:
		in = delInPS
	case shCmd:
		in = delInCmd
	}
	for k := range RunnerWords {
		if RunnerWords[k].Word == name && RunnerWords[k].in&in != 0 {
			return &RunnerWords[k]
		}
	}
	return nil
}

// delName gives a program's name as the tables match it: its base name, case-folded, .exe dropped; native is
// set when it is named as a program file (with a folder, a drive or .exe), which in PowerShell is the program and
// not a cmdlet's alias.
func delName(t string, d shDialect) (string, bool) {
	base := t
	if k := strings.LastIndexAny(base, `/\`); k >= 0 {
		base = base[k+1:]
	}
	native := base != t || strings.Contains(t, ":")
	low := strings.ToLower(base)
	if strings.HasSuffix(low, ".exe") {
		low = low[:len(low)-4]
		native = true
	} else if d == shPower && !strings.ContainsAny(t, "/:") && !strings.HasPrefix(t, ".") {
		native = false // Module\Cmdlet
	}
	return low, native
}

// delCdKind tells a folder word: delCdTo, delCdPush, delCdPop, or delCdNone.
func delCdKind(name string, d shDialect) int {
	switch name {
	case "cd", "chdir":
		return delCdTo
	case "pushd":
		return delCdPush
	case "popd":
		return delCdPop
	}
	if d == shPower {
		switch name {
		case "sl", "set-location":
			return delCdTo
		case "push-location":
			return delCdPush
		case "pop-location":
			return delCdPop
		}
	}
	return delCdNone
}

// delSteps bounds the walk past the words that run another program (each runner word, each env -S text, one step):
// a command exec reaches it in is refused (delete-unsplit), and one delProbe reaches it in counts as a delete; neither
// is allowed unread.
const delSteps = 256

// exec reads one command from its program on: past assignments and runners to the program, then a folder word, a
// shell given text, or a delete word, each judged. It gives what the command did to the folder.
func (r *delRun) exec(x delExec) delCd {
	for steps := 0; steps < delSteps; steps++ {
		if r.deny != nil {
			return delCd{}
		}
		w := x.words
		if x.d == shBash {
			r.envAssignWords(w)
		}
		for len(w) > 0 && w[0].assign {
			w = w[1:]
		}
		if len(w) == 0 {
			return delCd{}
		}
		x.words = w
		if w[0].expand || w[0].splat {
			r.unknownProgram(x)
			return delCd{}
		}
		if x.d == shCmd {
			x.words = delCmdGlued(x.words)
		}
		name, native := delName(x.words[0].text, x.d)
		if rn := delRunnerFor(name, x.d); rn != nil {
			next, ok := r.runner(rn, x)
			if !ok {
				return delCd{}
			}
			x = next
			continue
		}
		if r.envCommand(x, name); r.deny != nil {
			return delCd{}
		}
		if kind := delCdKind(name, x.d); kind != delCdNone && (x.d != shPower || !native) {
			return r.cd(x, kind)
		}
		if r.shell(name, x) {
			return delCd{}
		}
		if delBonsai(x.words[0].text) {
			r.bonsaiCall(x)
			return delCd{}
		}
		if name == "git" {
			nx, acts := r.git(x)
			for _, a := range acts {
				r.judge(a, nx)
			}
			return delCd{}
		}
		if dw := delWordFor(name, x.d, native); dw != nil {
			for _, a := range delParseWord(dw.Word)(r, x, x.words[1:]) {
				r.judge(a, x)
			}
		}
		return delCd{}
	}
	r.refuse(Decision{Rule: RuleDeleteUnsplit,
		Why: fmt.Sprintf("%s refused: a command runs its program through more than %d words that run another program "+
			"(sudo, env, nice and their kind, env -S texts among them), more than the check reads, so it cannot tell what "+
			"it runs (rule %s)", r.tool, delSteps, RuleDeleteUnsplit),
		Next: "run the program with fewer words before it that run another program"})
	return delCd{}
}

// delBonsai reports a program word that names Bonsai's own program: bonsai, by its base name with any of the
// endings a stand-in takes (StandInNames), letter case ignored.
func delBonsai(t string) bool {
	if k := strings.LastIndexAny(t, `/\`); k >= 0 {
		t = t[k+1:]
	}
	for _, n := range StandInNames {
		if strings.EqualFold(t, n) {
			return true
		}
	}
	return false
}

// bonsaiCall gives one bonsai call to the guard's judgment of Bonsai's own commands (commands.go's ownCommand),
// in each folder it may run in, or in the session's project folder when that cannot be told (5.3.2 note 7). Under
// xargs, directly or through a shell's text xargs runs, it is allowed only when its own words settle a command that
// is not a person's (delFedBonsai); else it is a person's step.
func (r *delRun) bonsaiCall(x delExec) {
	argv := make([]string, len(x.words))
	for k, w := range x.words {
		argv[k] = w.text
	}
	dirs := append([]string(nil), x.dirs.known...)
	if (x.dirs.unknown || len(dirs) == 0) && r.own != "" {
		dirs = appendNew(dirs, r.own)
	}
	for _, dir := range dirs {
		if d := r.j.ownCommand(dir, argv); !d.Allow {
			r.refuse(d)
			return
		}
	}
	if x.feed != "" || r.feed != "" {
		if why := delFedBonsai(x.words, argv, append(append([]string(nil), x.places...), r.places...)); why != "" {
			if r.envStep("runs bonsai under xargs " + why + ", so what xargs's input adds could make it a person's " +
				"command"); r.deny != nil {
				return
			}
		}
	}
	r.bonsai = true
}

// delFedBonsai says why a bonsai call xargs runs is open to its input, "" when its own words settle a command that
// is not a person's: a word holding a text xargs's input replaces (places: -I's, -i's and --replace's), or a variable
// or substitution, stands for one of its words; or its words leave the command open (commands.go's ownOpen).
func delFedBonsai(words []shWord, argv, places []string) string {
	for _, w := range words {
		if w.expand || w.splat {
			return "with a word a variable or substitution stands for"
		}
		for _, p := range places {
			if p != "" && strings.Contains(w.text, p) {
				return "with a word its input replaces (" + oneLine(p) + ")"
			}
		}
	}
	return ownOpen(argv)
}

// delCmdGlued splits a cmd.exe program glued to its options (rd/s/q) into its words.
func delCmdGlued(words []shWord) []shWord {
	t := words[0].text
	k := strings.IndexByte(t, '/')
	if k <= 0 || words[0].quoted {
		return words
	}
	switch strings.ToLower(t[:k]) {
	case "del", "erase", "rd", "rmdir", "move", "ren", "rename", "cd", "chdir", "start":
	default:
		return words
	}
	out := []shWord{{text: t[:k], raw: t[:k]}}
	out = append(out, shWord{text: t[k:], raw: t[k:]})
	return append(out, words[1:]...)
}

// runner skips a runner word to the program it runs. ok is false when it runs none, or the check has split what
// it runs already (env -S, wsl).
func (r *delRun) runner(rn *RunnerWord, x delExec) (delExec, bool) {
	words := x.words[1:]
	if rn.Word == "start" {
		for len(words) > 0 && strings.HasPrefix(words[0].text, "/") && !words[0].quoted {
			if strings.EqualFold(words[0].text, "/d") && len(words) > 1 {
				x.dirs = r.cdTo(x.dirs, words[1], x.d)
				words = words[1:]
			}
			words = words[1:]
		}
		if len(words) > 0 && words[0].quoted {
			words = words[1:] // the window's title
		}
		x.words = words
		return x, true
	}
	g := delParse(words, rn.opt)
	switch rn.Word {
	case "command":
		if g.has("v", "V") {
			return x, false
		}
	case "exec":
		if g.has("c") {
			r.envStep("starts a program with an empty environment (exec -c)")
			return x, false
		}
	}
	if rn.Word == "env" {
		for _, u := range g.val("u", "unset") {
			if envIs(u.text) {
				r.envStep("unsets " + ChildSessionEnv + " (env -u)")
				return x, false
			}
		}
		if g.has("i", "ignore-environment") || g.rest < len(words) && words[g.rest].text == "-" {
			r.envStep("starts a program with an empty environment (env -i)")
			return x, false
		}
	}
	switch rn.Word {
	case "env":
		if k := g.first("S", "split-string"); k >= 0 {
			// -S's text is more of env's own words (its options, assignments, then the program).
			more, ok := delEnvSplit(x.words[0], words, g, k, x.depth)
			if !ok {
				r.text(g.vals[g.on[k]][0].text+" "+delJoin(words[g.rest:], false), shBash, x.depth+1, x.dirs, x.loop)
				return x, false
			}
			x.words = more
			return x, more != nil
		}
		// env's own NAME=value words, up to its program, on their quote-free text (env reads 'NAME=' as one too).
		k := g.rest
		for k < len(words) && strings.Contains(words[k].text, "=") && !strings.HasPrefix(words[k].text, "=") {
			_, value, _ := strings.Cut(words[k].text, "=")
			switch {
			case envAssigns(words[k].text):
				r.envStep("assigns " + ChildSessionEnv + " (env)")
				return x, false
			case envAsData(value):
				r.envStep("puts the name " + ChildSessionEnv + " in a variable, which the check cannot follow")
				return x, false
			}
			k++
		}
		g.rest = k
	}
	for _, o := range rn.dir {
		for _, v := range g.vals[o] {
			x.dirs = r.cdTo(x.dirs, v, x.d)
		}
	}
	rest := words[g.rest:]
	switch rn.Word {
	case "timeout":
		if len(rest) > 0 {
			rest = rest[1:] // the duration
		}
	case "xargs":
		x.feed = "xargs"
		bare := 0 // -i and --replace given with no text of their own: {} is theirs
		for _, o := range g.on {
			if o == "i" || o == "replace" {
				bare++
			}
		}
		for _, v := range g.val("I", "i", "replace") {
			x.places = append(x.places, v.text)
		}
		if bare > len(g.val("i", "replace")) {
			x.places = append(x.places, "{}")
		}
	case "wsl", "watch":
		x.d = shBash
		if !g.has("e", "exec") && rn.Word == "wsl" || !g.has("x", "exec") && rn.Word == "watch" {
			r.text(delJoin(rest, false), shBash, x.depth+1, x.dirs, x.loop)
			return x, false
		}
	}
	x.words = rest
	return x, true
}

// unknownProgram reads a command whose program is a variable or a substitution ($RM -rf x, "$(which rm)" -r x):
// with -c it may be a shell, whose text is split again; with a recursive flag it is judged as a recursive delete,
// failing closed; else it is not judged.
func (r *delRun) unknownProgram(x delExec) {
	if text, ok := delLiteral(x.words[0], x.d, x.depth); ok {
		// "$(echo rm) -rf x": the program is literal text after all; unquoted, bash splits it into words, and the
		// check reads it as a command line again, failing closed.
		if strings.HasPrefix(strings.TrimSpace(x.words[0].raw), `"`) {
			w := x.words[0]
			w.text, w.raw, w.expand = text, text, false
			x.words = append([]shWord{w}, x.words[1:]...)
			r.exec(x)
			return
		}
		r.text(text+" "+delJoin(x.words[1:], false), x.d, x.depth+1, x.dirs, x.loop)
		return
	}
	words := x.words[1:]
	for k, w := range words {
		if w.text == "-c" && k+1 < len(words) {
			r.text(words[k+1].text, shBash, x.depth+1, x.dirs, x.loop)
			return
		}
	}
	recursive := false
	var ops []shWord
	for _, w := range words {
		t := w.text
		switch {
		case w.expand || t == "" || t[0] != '-' && t[0] != '/':
			ops = append(ops, w)
		case t == "--recursive" || len(t) >= 5 && strings.HasPrefix("--recursive", t):
			recursive = true
		case x.d == shPower && len(t) >= 2 && strings.HasPrefix("-recurse", strings.ToLower(t)):
			recursive = true
		case x.d == shCmd && strings.EqualFold(t, "/s"):
			recursive = true
		case len(t) > 1 && t[0] == '-' && t[1] != '-' && strings.Trim(t[1:], "fiIrRdv") == "" && strings.ContainsAny(t, "rR"):
			recursive = true
		}
	}
	if !recursive {
		if delHoldsWord(x.words[0].raw) != "" {
			r.judge(delAct{word: "a program named by a variable or substitution", unnamed: "its program is a variable " +
				"or substitution that holds a delete word, so the check cannot tell what it runs",
				next: "name the program plainly (rm, Remove-Item), not through a variable or substitution"}, x)
		}
		return
	}
	r.judge(delAct{word: "a program named by a variable", recursive: true, tree: true, ops: delOps(ops, x.d, false, false)}, x)
}

// shell splits again the text a shell is given; false when the program is not a shell. A shell run by xargs passes
// xargs's input on to the deletes in its text.
func (r *delRun) shell(name string, x delExec) bool {
	if x.feed != "" && r.feed == "" {
		r.feed, r.places = x.feed, x.places
		defer func() { r.feed, r.places = "", nil }()
	}
	switch {
	case delBashNames[name]:
		r.bashShell(x)
	case name == "eval" && x.d == shBash:
		args := x.words[1:]
		if len(args) == 1 {
			if text, ok := delLiteral(args[0], x.d, x.depth); ok {
				r.text(text, shBash, x.depth+1, x.dirs, x.loop)
				return true
			}
		}
		r.text(delJoin(args, false), shBash, x.depth+1, x.dirs, x.loop)
	case name == "trap" && x.d == shBash:
		args := x.words[1:]
		if len(args) > 0 && args[0].text == "--" {
			args = args[1:]
		}
		if len(args) >= 2 && args[0].text != "-" {
			r.shellText(args[0], shBash, x)
		}
	case (name == "su" || name == "script" || name == "flock" || name == "runuser") && x.d == shBash:
		args := x.words[1:]
		for k, w := range args {
			if (w.text == "-c" || w.text == "--command") && k+1 < len(args) {
				r.shellText(args[k+1], shBash, x)
			} else if strings.HasPrefix(w.text, "--command=") {
				r.text(strings.TrimPrefix(w.text, "--command="), shBash, x.depth+1, x.dirs, x.loop)
			}
		}
	case (name == "source" || name == ".") && x.d == shBash:
		if args := x.words[1:]; len(args) > 0 {
			if text, ok := delLiteral(args[0], x.d, x.depth); ok {
				r.text(text, shBash, x.depth+1, x.dirs, x.loop)
			}
		}
	case name == "cmd":
		r.cmdShell(x)
	case name == "powershell" || name == "pwsh" || name == "powershell_ise":
		r.psShell(x)
	case x.d == shPower && (name == "invoke-expression" || name == "iex"):
		var parts []string
		for _, w := range x.words[1:] {
			if !w.param {
				parts = append(parts, w.text)
			}
		}
		if len(parts) > 0 {
			r.text(strings.Join(parts, " "), shPower, x.depth+1, x.dirs, x.loop)
		} else {
			for _, s := range x.stdin {
				r.text(s, shPower, x.depth+1, x.dirs, x.loop)
			}
		}
	default:
		return false
	}
	return true
}

// shellText splits again the text a shell is given in one word: the word's own text, or, when the word is a
// substitution of literal output ("$(cat <<EOF ... EOF)"), that output.
func (r *delRun) shellText(w shWord, d shDialect, x delExec) {
	text := w.text
	if lit, ok := delLiteral(w, x.d, x.depth); ok {
		text = lit
	}
	r.text(text, d, x.depth+1, x.dirs, x.loop)
}

// bashShell reads bash's own options: with -c its text is split again; with no script (or -s) its input is.
func (r *delRun) bashShell(x delExec) {
	words := x.words[1:]
	i := 0
	cflag, sflag := false, false
	for i < len(words) {
		t := words[i].text
		if t == "--" || t == "-" {
			i++
			break
		}
		if strings.HasPrefix(t, "--") {
			if t == "--rcfile" || t == "--init-file" {
				i++
			}
			i++
			continue
		}
		if len(t) > 1 && (t[0] == '-' || t[0] == '+') {
			takes := 0
			for _, ch := range t[1:] {
				switch ch {
				case 'c':
					cflag = true
				case 's':
					sflag = true
				case 'o', 'O':
					takes++
				}
			}
			i += 1 + takes
			continue
		}
		break
	}
	if cflag {
		if i < len(words) {
			r.shellText(words[i], shBash, x)
		}
		return
	}
	if i < len(words) && !sflag {
		if text, ok := delLiteral(words[i], x.d, x.depth); ok {
			r.text(text, shBash, x.depth+1, x.dirs, x.loop) // bash <(echo ...): a script of literal text
		}
		return // a script run by name: not seen (note 5)
	}
	for _, s := range x.stdin {
		r.text(s, shBash, x.depth+1, x.dirs, x.loop)
	}
}

// cmdShell reads cmd.exe's options: the text after /c, /k or /r is split again as cmd.exe's; with none, its input.
func (r *delRun) cmdShell(x delExec) {
	words := x.words[1:]
	for k, w := range words {
		t := strings.ToLower(w.text)
		if !strings.HasPrefix(t, "/") && !strings.HasPrefix(t, "-") {
			continue
		}
		o := strings.TrimLeft(t, "/-")
		if o == "" || strings.IndexByte("ckr", o[0]) < 0 || strings.Contains(o, ":") {
			continue
		}
		var text string
		if len(o) == 1 {
			text = delJoin(words[k+1:], true)
		} else {
			// glued: /crd /s x
			at := strings.IndexAny(w.text, "cCkKrR")
			text = strings.TrimSpace(w.text[at+1:] + " " + delJoin(words[k+1:], true))
		}
		// cmd.exe's /c rule: a line starting with a quote loses that quote and the last one.
		if strings.HasPrefix(text, `"`) {
			text = text[1:]
			if q := strings.LastIndexByte(text, '"'); q >= 0 {
				text = text[:q] + text[q+1:]
			}
		}
		r.text(text, shCmd, x.depth+1, x.dirs, x.loop)
		return
	}
	for _, s := range x.stdin {
		r.text(s, shCmd, x.depth+1, x.dirs, x.loop)
	}
}

// psShell reads powershell's and pwsh's parameters: -Command's text (or input, for -Command -) and
// -EncodedCommand's decoded text are split again as PowerShell's; with no parameter the words are the command.
func (r *delRun) psShell(x delExec) {
	words := x.words[1:]
	for i := 0; i < len(words); i++ {
		w := words[i]
		t := strings.ToLower(w.text)
		if t == "-" {
			break // its input is the command
		}
		if len(t) < 2 || t[0] != '-' && t[0] != '/' {
			if len(words[i:]) == 1 {
				r.shellText(words[i], shPower, x)
				return
			}
			r.text(delJoin(words[i:], false), shPower, x.depth+1, x.dirs, x.loop)
			return
		}
		name, _, _ := strings.Cut(t[1:], ":")
		switch kind := delPSShellParam(name); kind {
		case "command":
			rest := words[i+1:]
			if len(rest) == 1 && rest[0].text == "-" || len(rest) == 0 {
				for _, s := range x.stdin {
					r.text(s, shPower, x.depth+1, x.dirs, x.loop)
				}
				return
			}
			if len(rest) == 1 {
				r.shellText(rest[0], shPower, x)
				return
			}
			r.text(delJoin(rest, false), shPower, x.depth+1, x.dirs, x.loop)
			return
		case "encodedcommand":
			if i+1 < len(words) {
				if raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(words[i+1].text)); err == nil {
					if text, ok := shDecodeUTF16(raw); ok {
						r.text(text, shPower, x.depth+1, x.dirs, x.loop)
					}
				}
			}
			return
		case "file":
			return // a script run by name: not seen (note 5)
		case "workingdirectory":
			if i+1 < len(words) {
				x.dirs = r.cdTo(x.dirs, words[i+1], shPower)
			}
			i++
		case "value":
			i++
		}
	}
	for _, s := range x.stdin {
		r.text(s, shPower, x.depth+1, x.dirs, x.loop)
	}
}

// delPSShellParam reads one of powershell's or pwsh's own parameters: "command", "encodedcommand", "file",
// "workingdirectory", "value" (one taking a value) or "switch". An ambiguous prefix that may be -Command is
// -Command, failing closed.
func delPSShellParam(name string) string {
	aliases := map[string]string{"c": "command", "e": "encodedcommand", "ec": "encodedcommand", "en": "encodedcommand",
		"enc": "encodedcommand", "f": "file", "ex": "executionpolicy", "ep": "executionpolicy", "w": "windowstyle",
		"wd": "workingdirectory", "nop": "noprofile", "noni": "noninteractive", "nol": "nologo", "l": "login",
		"i": "interactive", "v": "version", "o": "outputformat", "if": "inputformat"}
	full := []string{"command", "encodedcommand", "encodedarguments", "file", "executionpolicy", "windowstyle",
		"workingdirectory", "configurationname", "custompipename", "settingsfile", "inputformat", "outputformat",
		"version", "psconsolefile", "noprofile", "nologo", "noninteractive", "noexit", "sta", "mta", "login",
		"interactive", "noprofileloadtime", "help"}
	pick := aliases[name]
	if pick == "" {
		var cands []string
		for _, f := range full {
			if strings.HasPrefix(f, name) {
				cands = append(cands, f)
			}
		}
		for _, c := range cands {
			if c == "command" {
				pick = c
			}
		}
		if pick == "" && len(cands) > 0 {
			pick = cands[0]
		}
	}
	switch pick {
	case "command", "encodedcommand", "file", "workingdirectory":
		return pick
	case "encodedarguments", "executionpolicy", "windowstyle", "configurationname", "custompipename", "settingsfile",
		"inputformat", "outputformat", "version", "psconsolefile":
		return "value"
	}
	return "switch"
}

// delJoin joins words' texts with spaces; with quoteSpaced, a text holding a space is put in double quotes, as a
// program passes it on to cmd.exe.
func delJoin(words []shWord, quoteSpaced bool) string {
	parts := make([]string, 0, len(words))
	for _, w := range words {
		t := w.text
		if quoteSpaced && strings.ContainsAny(t, " \t") {
			t = `"` + t + `"`
		}
		parts = append(parts, t)
	}
	return strings.Join(parts, " ")
}

// cd reads a folder word and gives the folders the commands after it may run in.
func (r *delRun) cd(x delExec, kind int) delCd {
	if kind == delCdPop {
		return delCd{kind: delCdPop}
	}
	args := x.words[1:]
	var target *shWord
	for k := 0; k < len(args) && target == nil; k++ {
		t := args[k].text
		switch x.d {
		case shBash:
			if len(t) > 1 && t[0] == '-' && !args[k].expand {
				continue // -L, -P, -e, -@
			}
		case shCmd:
			if strings.EqualFold(t, "/d") {
				continue
			}
		case shPower:
			if args[k].param {
				name := strings.ToLower(strings.TrimPrefix(t, "-"))
				name, _, _ = strings.Cut(name, ":")
				if strings.HasPrefix("path", name) || strings.HasPrefix("literalpath", name) || strings.HasPrefix("stackname", name) {
					if strings.HasPrefix("stackname", name) {
						k++
					}
					continue
				}
				continue
			}
		}
		target = &args[k]
	}
	if target == nil {
		h := delHome()
		switch {
		case x.d == shCmd:
			return delCd{} // cd alone prints the folder
		case h == "":
			return delCd{kind: kind, to: delDirs{unknown: true}}
		case x.d == shPower:
			return delCd{kind: kind, to: x.dirs.union(delDirs{known: []string{h}})} // PowerShell 7 goes home, 5.1 stays
		}
		return delCd{kind: kind, to: delDirs{known: []string{h}}}
	}
	return delCd{kind: kind, to: r.cdTo(x.dirs, *target, x.d)}
}

// cdTo gives the folders a cd to the word w leads to from dirs: each known folder moved, or one that cannot be told
// for a variable, a substitution, a glob, -, or a stack entry.
func (r *delRun) cdTo(dirs delDirs, w shWord, d shDialect) delDirs {
	t := w.text
	if w.expand || w.glob || w.splat || len(w.alts) > 0 || t == "" || t == "-" || d == shBash && (strings.HasPrefix(t, "+") ||
		strings.HasPrefix(t, "-")) {
		return delDirs{unknown: true}
	}
	if d == shBash && strings.HasPrefix(t, "~") {
		head, rest, _ := strings.Cut(t[1:], "/")
		switch here, ok := delTildeStack(head); {
		case ok && here:
			t = "./" + rest // ~+, ~0: the folder it runs in
		case ok && delTildeBottom(head):
			// The stack's bottom: the folder it runs in until a pushd, then the one before it, which the check cannot tell.
			out := r.cdTo(dirs, shWord{text: "./" + rest, raw: "./" + rest}, d)
			out.unknown = true
			return out
		case ok:
			return delDirs{unknown: true} // ~-, ~N: a folder the check cannot tell
		}
	}
	if d != shBash && strings.ContainsRune(t, '\\') && !strings.ContainsRune(t, ':') && delSlashOS() {
		t = strings.ReplaceAll(t, `\`, "/")
	}
	if d == shPower {
		t = delProvider(t)
	}
	if delIsAbs(t) {
		base := r.base
		if len(dirs.known) > 0 {
			base = dirs.known[0]
		}
		if base == "" {
			base = string('/')
		}
		return delDirs{known: []string{delAbs(base, t)}}
	}
	out := delDirs{unknown: dirs.unknown}
	for _, k := range dirs.known {
		out.known = appendNew(out.known, delAbs(k, t))
	}
	if len(out.known) == 0 {
		out.unknown = true
	}
	return out
}

// delOpt is a GNU-style program's options.
type delOpt struct {
	short    string   // short options with no value
	shortOpt string   // short options whose value is optional: the rest of the bundle only (GNU xargs's -iR, -l5)
	shortVal string   // short options that take a value (the rest of the bundle, or the next word)
	long     []string // long options with no value
	longVal  []string // long options that take a value (after =, or the next word)
	longOpt  []string // long options whose value is optional (after = only)
	permute  bool     // options may follow operands, as GNU's do; else the first operand ends them
}

// delGot is a command's options as read.
type delGot struct {
	on   []string            // the options given: a short one by its letter, a long one by its full name
	at   []int               // per option of on, the index of the word it is written in
	end  []int               // per option of on, the index of the word after its word and the value that word takes
	vals map[string][]shWord // the values of those that take one
	ops  []shWord            // the operands, in order
	rest int                 // without permute: the index of the first operand
}

func (g delGot) has(names ...string) bool {
	for _, o := range g.on {
		for _, n := range names {
			if o == n {
				return true
			}
		}
	}
	return false
}

// first gives the index in on of the first of the options names, -1 for none.
func (g delGot) first(names ...string) int {
	for k, o := range g.on {
		for _, n := range names {
			if o == n {
				return k
			}
		}
	}
	return -1
}

func (g delGot) val(names ...string) []shWord {
	var out []shWord
	for _, n := range names {
		out = append(out, g.vals[n]...)
	}
	return out
}

// resolve gives the long options a written name may be: itself, else every option it is a prefix of; unknown, it
// is kept as written.
func (o delOpt) resolve(name string) []string {
	all := make([]string, 0, len(o.long)+len(o.longVal)+len(o.longOpt))
	all = append(append(append(all, o.long...), o.longVal...), o.longOpt...)
	for _, n := range all {
		if n == name {
			return []string{n}
		}
	}
	var out []string
	for _, n := range all {
		if name != "" && strings.HasPrefix(n, name) {
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		return []string{name}
	}
	return out
}

func delHasName(list []string, n string) bool {
	for _, x := range list {
		if x == n {
			return true
		}
	}
	return false
}

// delParse reads a GNU-style command's words after its program.
func delParse(words []shWord, o delOpt) delGot {
	g := delGot{vals: map[string][]shWord{}, rest: len(words)}
	for i := 0; i < len(words); i++ {
		w := words[i]
		t := w.text
		if t == "--" && !w.expand {
			if !o.permute {
				g.rest = i + 1
				g.ops = words[i+1:]
				return g
			}
			g.ops = append(g.ops, words[i+1:]...)
			return g
		}
		if len(t) > 2 && strings.HasPrefix(t, "--") && !w.expand {
			name, v, hasVal := strings.Cut(t[2:], "=")
			took, first := false, len(g.on)
			for _, full := range o.resolve(name) {
				g.on, g.at = append(g.on, full), append(g.at, i)
				switch {
				case hasVal:
					g.vals[full] = append(g.vals[full], delValue(w, v))
				case delHasName(o.longVal, full) && i+1 < len(words) && !took:
					g.vals[full] = append(g.vals[full], words[i+1])
					took = true
				}
			}
			if took {
				i++
			}
			for range g.on[first:] {
				g.end = append(g.end, i+1)
			}
			continue
		}
		if len(t) > 1 && t[0] == '-' && !w.expand {
			first := len(g.on)
			for k := 1; k < len(t); k++ {
				ch := t[k : k+1]
				g.on, g.at = append(g.on, ch), append(g.at, i)
				if strings.Contains(o.shortVal, ch) {
					if k+1 < len(t) {
						g.vals[ch] = append(g.vals[ch], delValue(w, t[k+1:]))
					} else if i+1 < len(words) {
						g.vals[ch] = append(g.vals[ch], words[i+1])
						i++
					}
					break
				}
				if strings.Contains(o.shortOpt, ch) {
					if k+1 < len(t) {
						g.vals[ch] = append(g.vals[ch], delValue(w, t[k+1:]))
					}
					break
				}
			}
			for range g.on[first:] {
				g.end = append(g.end, i+1)
			}
			continue
		}
		if !o.permute {
			g.rest = i
			g.ops = words[i:]
			return g
		}
		g.ops = append(g.ops, w)
	}
	return g
}

// delValue makes a word of a value written inside an option's word (--target-directory=x, -tx).
func delValue(w shWord, v string) shWord {
	return shWord{text: v, raw: v, quoted: w.quoted, glob: strings.ContainsAny(v, "*?["), expand: w.expand && strings.ContainsAny(v, "$`%")}
}

// delOps makes operands of words, one per word its brace list writes out.
func delOps(words []shWord, d shDialect, literal, git bool) []delOp {
	globs := "*?["
	if d == shCmd {
		globs = "*?"
	}
	var out []delOp
	for _, w := range words {
		texts := []string{w.text}
		if w.alts != nil {
			texts = w.alts
		}
		for _, t := range texts {
			op := delOp{text: t, expand: w.expand || w.splat, literal: literal, git: git, glob: w.glob, item: w.item}
			if w.expand && w.alts == nil {
				wc := w
				op.word = &wc
			}
			if w.alts != nil {
				op.glob = strings.ContainsAny(t, globs)
			}
			if d == shPower {
				low := strings.ToLower(t)
				op.pipeVar = strings.Contains(low, "$_") || strings.Contains(low, "$psitem")
			}
			out = append(out, op)
		}
	}
	return out
}

// delSlashOS reports a system whose own separator is /, where PowerShell and cmd.exe's \ is read as / too.
func delSlashOS() bool { return runtime.GOOS != "windows" }

func delRm(r *delRun, x delExec, words []shWord) []delAct {
	g := delParse(words, delOpt{short: "fiIrRdv", long: []string{"force", "one-file-system", "no-preserve-root",
		"recursive", "dir", "verbose", "help", "version"}, longOpt: []string{"interactive", "preserve-root"}, permute: true})
	if g.has("help", "version") {
		return nil
	}
	a := delAct{word: "rm", ops: delOps(g.ops, x.d, false, false)}
	a.recursive = g.has("r", "R", "recursive")
	a.tree = a.recursive
	if g.has("no-preserve-root") {
		a.unnamed = "--no-preserve-root lets it delete the root folder"
		a.next = "leave --no-preserve-root out, and name what to delete"
	}
	return []delAct{a}
}

func delRmdir(r *delRun, x delExec, words []shWord) []delAct {
	g := delParse(words, delOpt{short: "pv", long: []string{"parents", "verbose", "ignore-fail-on-non-empty", "help",
		"version"}, permute: true})
	if g.has("help", "version") {
		return nil
	}
	return []delAct{{word: "rmdir", ops: delOps(g.ops, x.d, false, false)}}
}

func delUnlink(r *delRun, x delExec, words []shWord) []delAct {
	g := delParse(words, delOpt{long: []string{"help", "version"}, permute: true})
	if g.has("help", "version") {
		return nil
	}
	return []delAct{{word: "unlink", ops: delOps(g.ops, x.d, false, false)}}
}

func delShred(r *delRun, x delExec, words []shWord) []delAct {
	g := delParse(words, delOpt{short: "fuvxz", shortVal: "ns", long: []string{"force", "verbose", "exact", "zero",
		"help", "version"}, longVal: []string{"iterations", "size", "random-source"}, longOpt: []string{"remove"}, permute: true})
	if g.has("help", "version") {
		return nil
	}
	return []delAct{{word: "shred", ops: delOps(g.ops, x.d, false, false)}}
}

func delMv(r *delRun, x delExec, words []shWord) []delAct {
	g := delParse(words, delOpt{short: "bfinuvTZ", shortVal: "St", long: []string{"force", "interactive", "no-clobber",
		"strip-trailing-slashes", "no-target-directory", "verbose", "exchange", "no-copy", "debug", "help", "version"},
		longVal: []string{"suffix", "target-directory"}, longOpt: []string{"backup", "update", "context"}, permute: true})
	if g.has("help", "version") {
		return nil
	}
	ops := delOps(g.ops, x.d, false, false)
	a := delAct{word: "mv", move: true, tree: true}
	if t := g.val("t", "target-directory"); len(t) > 0 {
		a.ops, a.dests, a.destDir = ops, delOps(t, x.d, false, false), true
	} else if len(ops) >= 2 {
		a.ops, a.dests, a.destDir = ops[:len(ops)-1], ops[len(ops)-1:], !g.has("T", "no-target-directory")
	} else {
		a.ops = ops
	}
	if g.has("exchange") {
		a.ops = append(a.ops, a.dests...)
	}
	return []delAct{a}
}

// delFind reads find: its starting points, and its expression for -delete or an -exec running a delete word.
func delFind(r *delRun, x delExec, words []shWord) []delAct {
	i := 0
	for i < len(words) {
		t := words[i].text
		if t == "-H" || t == "-L" || t == "-P" || strings.HasPrefix(t, "-O") && len(t) <= 3 {
			i++
		} else if t == "-D" {
			i += 2
		} else {
			break
		}
	}
	var starts []shWord
	for i < len(words) {
		t := words[i].text
		if len(t) > 1 && t[0] == '-' || t == "(" || t == ")" || t == "!" || t == "," {
			break
		}
		starts = append(starts, words[i])
		i++
	}
	deletes := false
	var expr []shWord
	if i < len(words) {
		expr = words[i:]
	}
	for k := 0; k < len(expr); k++ {
		switch expr[k].text {
		case "-delete":
			deletes = true
		case "-exec", "-execdir", "-ok", "-okdir":
			end := k + 1
			for end < len(expr) && expr[end].text != ";" && expr[end].text != "+" {
				end++
			}
			run := expr[k+1 : end]
			if delProbe(run, x.d) {
				deletes = true
			}
			// What -exec runs is read as any command is (a delete among find's items, a bonsai call, a shell's text).
			sub := x
			sub.find = true
			sub.words = delDropBraces(run)
			r.exec(sub)
			k = end
		}
	}
	if !deletes {
		return nil
	}
	if len(starts) == 0 {
		starts = []shWord{{text: ".", raw: "."}} // GNU find's default
	}
	return []delAct{{word: "find", bulk: "find", tree: true, ops: delOps(starts, x.d, false, false)}}
}

// delDropBraces marks find's {} in the command it runs: an item find gives, judged as find's starting points are.
// A word holding {} among other text cannot be named.
func delDropBraces(words []shWord) []shWord {
	out := make([]shWord, 0, len(words))
	for _, w := range words {
		switch {
		case w.text == "{}":
			w.item = true
		case strings.Contains(w.text, "{}"):
			w.expand = true
		}
		out = append(out, w)
	}
	return out
}

// delProbe reports whether a command, past its runners (env -S's text read as exec reads it), is a delete word or a
// shell (which may delete), without judging it: what makes find a bulk delete. What it cannot read (runners past the
// walk's bound, a -S text that is not one simple command) counts as a delete.
func delProbe(words []shWord, d shDialect) bool {
	for steps := 0; len(words) > 0; steps++ {
		if steps == delSteps {
			return true
		}
		for len(words) > 0 && words[0].assign {
			words = words[1:]
		}
		if len(words) == 0 {
			return false
		}
		name, native := delName(words[0].text, d)
		if rn := delRunnerFor(name, d); rn != nil {
			g := delParse(words[1:], rn.opt)
			if k := g.first("S", "split-string"); rn.Word == "env" && k >= 0 {
				more, ok := delEnvSplit(words[0], words[1:], g, k, 0)
				if !ok {
					return true
				}
				words = more
				continue
			}
			words = words[1:][g.rest:]
			if rn.Word == "timeout" && len(words) > 0 {
				words = words[1:]
			}
			continue
		}
		if delBashNames[name] || name == "cmd" || name == "powershell" || name == "pwsh" || name == "eval" {
			return true
		}
		if name == "git" {
			for _, w := range words[1:] {
				switch w.text {
				case "rm", "clean", "mv":
					return true
				case "stash":
					return false
				}
			}
			return false
		}
		return delWordFor(name, d, native) != nil
	}
	return false
}

func delRsync(r *delRun, x delExec, words []shWord) []delAct {
	g := delParse(words, delOpt{shortVal: "efTBM", long: []string{"dry-run"}, longVal: []string{"exclude", "include",
		"exclude-from", "include-from", "files-from", "filter", "rsh", "rsync-path", "temp-dir", "compare-dest",
		"copy-dest", "link-dest", "backup-dir", "suffix", "chmod", "chown", "usermap", "groupmap", "max-size", "min-size",
		"max-delete", "partial-dir", "log-file", "log-file-format", "out-format", "password-file", "port", "sockopts",
		"timeout", "contimeout", "bwlimit", "modify-window", "block-size", "compress-level", "compress-choice",
		"checksum-choice", "skip-compress", "info", "debug", "iconv", "protocol", "address", "write-batch",
		"only-write-batch", "read-batch", "early-input", "stop-after", "stop-at", "outbuf", "remote-option", "copy-as",
		"checksum-seed", "max-alloc"}, permute: true})
	if g.has("n", "dry-run") {
		return nil
	}
	del, rm := false, false
	for _, o := range g.on {
		if o == "del" || strings.HasPrefix(o, "dele") {
			del = true
		}
		if strings.HasPrefix(o, "remove") {
			rm = true
		}
	}
	if len(g.ops) < 2 {
		return nil
	}
	ops := delOps(g.ops, x.d, false, false)
	var acts []delAct
	if del {
		acts = append(acts, delAct{word: "rsync --delete", bulk: "rsync --delete", tree: true, ops: ops[len(ops)-1:]})
	}
	if rm {
		acts = append(acts, delAct{word: "rsync --remove-source-files", move: true, tree: true, ops: ops[:len(ops)-1]})
	}
	return acts
}

func delRobocopy(r *delRun, x delExec, words []shWord) []delAct {
	var ops []shWord
	var opts []string
	started := false
	for _, w := range words {
		t := w.text
		name := strings.ToLower(strings.TrimLeft(t, "/-"))
		if (strings.HasPrefix(t, "/") || strings.HasPrefix(t, "-")) && name != "" && !strings.ContainsAny(name, `/\`) {
			started = true
			name, _, _ = strings.Cut(name, ":")
			opts = append(opts, name)
			continue
		}
		if !started {
			ops = append(ops, w)
		}
	}
	has := func(names ...string) bool {
		for _, o := range opts {
			for _, n := range names {
				if o == n {
					return true
				}
			}
		}
		return false
	}
	if has("l") || len(ops) < 2 {
		return nil
	}
	all := delOps(ops, x.d, false, false)
	var acts []delAct
	if has("mir", "purge") {
		acts = append(acts, delAct{word: "robocopy /MIR", bulk: "robocopy /MIR", tree: true, ops: all[1:2]})
	}
	if has("mov", "move") {
		acts = append(acts, delAct{word: "robocopy /MOVE", move: true, tree: true, ops: all[:1]})
	}
	return acts
}

// git reads git's global options (-C, -c, --git-dir, --work-tree, ...) and then its delete words.
func (r *delRun) git(x delExec) (delExec, []delAct) {
	words := x.words[1:]
	i := 0
	for i < len(words) {
		t := words[i].text
		if len(t) < 2 || t[0] != '-' {
			break
		}
		switch t {
		case "-C":
			if i+1 < len(words) {
				x.dirs = r.cdTo(x.dirs, words[i+1], x.d)
			}
			i += 2
		case "-c", "--namespace", "--super-prefix", "--config-env", "--attr-source", "--list-cmds":
			i += 2
		case "--git-dir", "--work-tree":
			x.dirs.unknown = true
			i += 2
		case "--icase-pathspecs":
			x.icase = true
			i++
		default:
			if strings.HasPrefix(t, "--git-dir=") || strings.HasPrefix(t, "--work-tree=") {
				x.dirs.unknown = true
			}
			i++
		}
	}
	if i >= len(words) {
		return x, nil
	}
	args := words[i+1:]
	gitOps := func(ws []shWord) []delOp { return delOps(ws, x.d, false, true) }
	dry := func(g delGot) bool { return g.has("n", "dry-run") && !g.has("no-dry-run") }
	switch words[i].text {
	case "rm":
		g := delParse(args, delOpt{short: "fnrq", long: []string{"force", "dry-run", "no-dry-run", "cached",
			"ignore-unmatch", "sparse", "quiet", "pathspec-file-nul"}, longVal: []string{"pathspec-from-file"}, permute: true})
		if dry(g) {
			return x, nil
		}
		a := delAct{word: "git rm", recursive: g.has("r"), tree: g.has("r"), ops: gitOps(g.ops)}
		if len(g.val("pathspec-from-file")) > 0 {
			a.unnamed = "it takes its paths from a file (--pathspec-from-file)"
			a.next = "name the paths in the command"
		}
		return x, []delAct{a}
	case "clean":
		g := delParse(args, delOpt{short: "dfinqxX", shortVal: "e", long: []string{"dry-run", "no-dry-run", "force",
			"interactive", "quiet"}, longVal: []string{"exclude"}, permute: true})
		if dry(g) {
			return x, nil
		}
		a := delAct{word: "git clean", bulk: "git clean", tree: true, ops: gitOps(g.ops)}
		switch {
		case g.has("x", "X"):
			a.unnamed = "git clean -x or -X deletes ignored files, .bonsai/local/ among them"
			a.next = "leave -x and -X out (git clean without them keeps ignored files), and name the folder to clean"
		case len(g.ops) == 0:
			a.unnamed = "git clean with no path cleans the whole checkout"
			a.next = "name the folder to clean (git clean -fd build/), or delete the files by name"
		}
		return x, []delAct{a}
	case "stash":
		if len(args) > 0 && !strings.HasPrefix(args[0].text, "-") {
			switch args[0].text {
			case "push", "save":
				args = args[1:]
			default:
				return x, nil // list, show, pop, apply, drop, branch, clear, create, store
			}
		}
		g := delParse(args, delOpt{short: "akpqSu", shortVal: "m", long: []string{"all", "include-untracked",
			"keep-index", "no-keep-index", "patch", "staged", "quiet", "pathspec-file-nul", "no-include-untracked"},
			longVal: []string{"message", "pathspec-from-file"}, permute: true})
		if g.has("a", "all") {
			return x, []delAct{{word: "git stash", unnamed: "git stash --all moves ignored files away, .bonsai/local/ among them",
				next: "use git stash or git stash -u, which leave ignored files in place"}}
		}
		return x, nil
	case "mv":
		g := delParse(args, delOpt{short: "fknv", long: []string{"force", "dry-run", "no-dry-run", "verbose", "sparse"},
			permute: true})
		if dry(g) {
			return x, nil
		}
		ops := gitOps(g.ops)
		a := delAct{word: "git mv", move: true, tree: true, ops: ops}
		if len(ops) >= 2 {
			a.ops, a.dests, a.destDir = ops[:len(ops)-1], ops[len(ops)-1:], true
		}
		return x, []delAct{a}
	}
	return x, nil
}

// delPSCmdlet is a PowerShell cmdlet's parameters, lower case.
type delPSCmdlet struct {
	params     []string
	switches   []string
	aliases    map[string]string
	positional []string
}

// delPSCommon are PowerShell's common parameters and their aliases.
var (
	delPSCommon = []string{"verbose", "debug", "erroraction", "warningaction", "informationaction", "progressaction",
		"errorvariable", "warningvariable", "informationvariable", "outvariable", "outbuffer", "pipelinevariable"}
	delPSCommonAliases = map[string]string{"vb": "verbose", "db": "debug", "ea": "erroraction", "wa": "warningaction",
		"infa": "informationaction", "proga": "progressaction", "ev": "errorvariable", "wv": "warningvariable",
		"iv": "informationvariable", "ov": "outvariable", "ob": "outbuffer", "pv": "pipelinevariable"}
)

var (
	delPSRemove = delPSCmdlet{
		params:     []string{"path", "literalpath", "filter", "include", "exclude", "recurse", "force", "credential", "stream", "whatif", "confirm"},
		switches:   []string{"recurse", "force", "whatif", "confirm", "verbose", "debug"},
		aliases:    map[string]string{"pspath": "literalpath", "lp": "literalpath", "wi": "whatif", "cf": "confirm"},
		positional: []string{"path"},
	}
	delPSMove = delPSCmdlet{
		params:     []string{"path", "literalpath", "destination", "filter", "include", "exclude", "force", "passthru", "credential", "whatif", "confirm"},
		switches:   []string{"force", "passthru", "whatif", "confirm", "verbose", "debug"},
		aliases:    map[string]string{"pspath": "literalpath", "lp": "literalpath", "wi": "whatif", "cf": "confirm"},
		positional: []string{"path", "destination"},
	}
	delPSRename = delPSCmdlet{
		params:     []string{"path", "literalpath", "newname", "force", "passthru", "credential", "whatif", "confirm"},
		switches:   []string{"force", "passthru", "whatif", "confirm", "verbose", "debug"},
		aliases:    map[string]string{"pspath": "literalpath", "lp": "literalpath", "wi": "whatif", "cf": "confirm"},
		positional: []string{"path", "newname"},
	}
)

// resolve gives the parameters a written name may be: an alias's, the one it names, else every one it is a prefix
// of (ambiguous counts as each); none for an unknown name.
func (c delPSCmdlet) resolve(name string) []string {
	if a := c.aliases[name]; a != "" {
		return []string{a}
	}
	if a := delPSCommonAliases[name]; a != "" {
		return []string{a}
	}
	all := append(append([]string(nil), c.params...), delPSCommon...)
	for _, p := range all {
		if p == name {
			return []string{p}
		}
	}
	var out []string
	for _, p := range all {
		if name != "" && strings.HasPrefix(p, name) {
			out = append(out, p)
		}
	}
	return out
}

// delPSBound is a cmdlet call's parameters as bound.
type delPSBound struct {
	vals  map[string][]shWord
	on    map[string]bool
	splat bool
}

// delPSBind binds a cmdlet's words to its parameters as PowerShell does: by name (any prefix), -Name:value, a
// switch's :$false, and by position; words joined by commas are one array.
func delPSBind(words []shWord, c delPSCmdlet) delPSBound {
	b := delPSBound{vals: map[string][]shWord{}, on: map[string]bool{}}
	pos := 0
	for k := 0; k < len(words); k++ {
		w := words[k]
		if w.splat {
			b.splat = true
			continue
		}
		if w.param {
			orig := w.text[1:]
			nameO, inline, hasInline := strings.Cut(orig, ":")
			names := c.resolve(strings.ToLower(nameO))
			if len(names) == 0 {
				continue // an unknown parameter: PowerShell refuses the call
			}
			isSwitch := true
			for _, n := range names {
				if !delHasName(c.switches, n) {
					isSwitch = false
				}
			}
			if isSwitch {
				off := hasInline && (strings.EqualFold(inline, "$false") || inline == "0")
				for _, n := range names {
					if !off {
						b.on[n] = true
					}
				}
				continue
			}
			var vals []shWord
			if hasInline && inline != "" {
				vals = append(vals, delValue(w, inline))
			} else {
				for k+1 < len(words) {
					k++
					vals = append(vals, words[k])
					if !words[k].comma {
						break
					}
				}
			}
			for _, n := range names {
				b.vals[n] = append(b.vals[n], vals...)
			}
			continue
		}
		vals := []shWord{w}
		for w.comma && k+1 < len(words) {
			k++
			w = words[k]
			vals = append(vals, w)
		}
		name := c.positional[0]
		if pos < len(c.positional) {
			name = c.positional[pos]
		}
		pos++
		b.vals[name] = append(b.vals[name], vals...)
	}
	return b
}

// delPSOps gives a cmdlet's -Path operands (globs read) and -LiteralPath ones (as written).
func delPSOps(b delPSBound) []delOp {
	return append(delOps(b.vals["path"], shPower, false, false), delOps(b.vals["literalpath"], shPower, true, false)...)
}

// delPSFed marks a cmdlet's act unnamed when its items come from a pipeline or a splat.
func delPSFed(a *delAct, x delExec, b delPSBound) {
	switch {
	case b.splat:
		a.unnamed = "its parameters come from a splat (@name), which the check cannot read"
		a.next = "write the parameters out in the command"
	case x.piped:
		a.unnamed = "it takes its items from a pipeline, which the check cannot see"
		a.next = "name the paths with -LiteralPath (or -Path) in the command itself, not through a pipeline"
	}
}

func delRemoveItem(r *delRun, x delExec, words []shWord) []delAct {
	b := delPSBind(words, delPSRemove)
	if b.on["whatif"] {
		return nil
	}
	a := delAct{word: "Remove-Item", recursive: b.on["recurse"], tree: true, ops: delPSOps(b)}
	for _, k := range []string{"include", "exclude", "filter"} {
		if len(b.vals[k]) > 0 {
			a.bulk = "-" + strings.ToUpper(k[:1]) + k[1:]
			break
		}
	}
	delPSFed(&a, x, b)
	return []delAct{a}
}

func delMoveItem(r *delRun, x delExec, words []shWord) []delAct {
	b := delPSBind(words, delPSMove)
	if b.on["whatif"] {
		return nil
	}
	a := delAct{word: "Move-Item", move: true, tree: true, ops: delPSOps(b),
		dests: delOps(b.vals["destination"], shPower, false, false), destDir: true}
	delPSFed(&a, x, b)
	return []delAct{a}
}

func delRenameItem(r *delRun, x delExec, words []shWord) []delAct {
	b := delPSBind(words, delPSRename)
	if b.on["whatif"] {
		return nil
	}
	a := delAct{word: "Rename-Item", move: true, tree: true, ops: delPSOps(b), rename: true}
	a.dests = delBeside(a.ops, delOps(b.vals["newname"], shPower, false, false))
	delPSFed(&a, x, b)
	return []delAct{a}
}

// delBeside gives the paths a rename leads to: each new name in its source's folder.
func delBeside(srcs, names []delOp) []delOp {
	var out []delOp
	for _, s := range srcs {
		dir := path.Dir(strings.ReplaceAll(s.text, `\`, "/"))
		for _, n := range names {
			d := n
			if dir != "." && !delIsAbs(n.text) {
				d.text = dir + "/" + n.text
			}
			out = append(out, d)
		}
	}
	return out
}

// delCmdOpts splits cmd.exe's words into its /options (lower case, glued ones apart) and its operands.
func delCmdOpts(words []shWord) ([]string, []shWord) {
	var opts []string
	var ops []shWord
	for _, w := range words {
		if strings.HasPrefix(w.text, "/") && !w.quoted {
			for _, part := range strings.Split(w.text[1:], "/") {
				if part != "" {
					opts = append(opts, strings.ToLower(part))
				}
			}
			continue
		}
		ops = append(ops, w)
	}
	return opts, ops
}

func delCmdDel(r *delRun, x delExec, words []shWord) []delAct {
	opts, ops := delCmdOpts(words)
	return []delAct{{word: "del", recursive: delHasName(opts, "s"), tree: true, ops: delOps(ops, shCmd, false, false)}}
}

func delCmdRd(r *delRun, x delExec, words []shWord) []delAct {
	opts, ops := delCmdOpts(words)
	s := delHasName(opts, "s")
	return []delAct{{word: "rd", recursive: s, tree: s, ops: delOps(ops, shCmd, false, false)}}
}

func delCmdMove(r *delRun, x delExec, words []shWord) []delAct {
	_, ws := delCmdOpts(words)
	ops := delOps(ws, shCmd, false, false)
	a := delAct{word: "move", move: true, tree: true, ops: ops}
	if len(ops) >= 2 {
		a.ops, a.dests, a.destDir = ops[:len(ops)-1], ops[len(ops)-1:], true
	}
	return []delAct{a}
}

func delCmdRen(r *delRun, x delExec, words []shWord) []delAct {
	_, ws := delCmdOpts(words)
	ops := delOps(ws, shCmd, false, false)
	a := delAct{word: "ren", move: true, tree: true, rename: true, ops: ops}
	if len(ops) >= 2 {
		a.ops = ops[:1]
		a.dests = delBeside(ops[:1], ops[1:2])
	}
	return []delAct{a}
}

// delHoldsBonsai gives the kind of refused bonsai call a text the check cannot split holds, "" for none: "home" for
// bonsai settings set and bonsai labels attach or detach (always refused), "person" for bonsai with --allow-exec or
// unlink (a person's step). Words are cut and matched as delHoldsWord cuts them.
func delHoldsBonsai(s string) string {
	low := strings.ToLower(s)
	toks := strings.FieldsFunc(low, func(c rune) bool {
		return !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || strings.ContainsRune("._-/\\:", c) || c >= 0x80)
	})
	seen, kind := false, ""
	for k, t := range toks {
		if delBonsai(t) {
			seen = true
			continue
		}
		if !seen {
			continue
		}
		switch {
		case t == "settings" && k+1 < len(toks) && toks[k+1] == "set", t == "labels":
			return "home"
		case t == "unlink" || strings.HasPrefix(strings.TrimLeft(t, "-"), "allow-exec") && strings.HasPrefix(t, "-"):
			kind = "person"
		}
	}
	return kind
}

// unsplitBonsai refuses a line the check cannot split that holds a refused bonsai call (delHoldsBonsai's kind).
func (r *delRun) unsplitBonsai(kind string) {
	if kind == "home" {
		r.refuse(deny(RuleHomeCommand, r.tool+" refused: the command line holds bonsai settings set or bonsai labels, "+
			"and the check cannot split it; those write this machine's settings or label files in Bonsai's home, which "+
			"never change from an agent's session", personNext))
		return
	}
	r.envStep("holds a bonsai call with --allow-exec or unlink that the check cannot split")
}

// delHoldsWord gives the delete word a text the check cannot split holds, "" for none: a word that deletes by its
// name, git with rm, clean, mv or stash, find with -delete or -exec, rsync with --del or --remove, robocopy with /MIR,
// /PURGE or /MOV. Words are cut at anything but letters, digits and . _ - / \ : and matched by base name, .exe
// dropped, case-folded.
func delHoldsWord(s string) string {
	low := strings.ToLower(s)
	toks := strings.FieldsFunc(low, func(c rune) bool {
		return !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || strings.ContainsRune("._-/\\:", c) || c >= 0x80)
	})
	git := false
	for _, t := range toks {
		base := t
		if k := strings.LastIndexAny(base, `/\`); k >= 0 {
			base = base[k+1:]
		}
		base = strings.TrimSuffix(base, ".exe")
		if delByName[base] {
			return base
		}
		switch base {
		case "git":
			git = true
		case "clean", "stash":
			if git {
				return "git " + base
			}
		case "find":
			if strings.Contains(low, "-delete") || strings.Contains(low, "-exec") || strings.Contains(low, "-ok") {
				return base
			}
		case "rsync":
			if strings.Contains(low, "--del") || strings.Contains(low, "--remove") {
				return base
			}
		case "robocopy":
			if strings.Contains(low, "mir") || strings.Contains(low, "purge") || strings.Contains(low, "mov") {
				return base
			}
		}
	}
	return ""
}
