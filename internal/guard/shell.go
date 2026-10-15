package guard

// The delete check (spec §7, "Shell commands"; plan-5 5.3.3): in a Bash, Monitor or PowerShell call, a recursive or
// bulk delete that does not name what it deletes is refused (delete-unnamed), and a delete or move of a guarded
// path is judged by the guard's rules as an edit of that path (delete-protected). The director's answer to question A
// (spec §19) is no shell reader: the check splits a command line into words and reads only what it needs to find
// a delete word and its operands. Nothing is expanded or run.
//
// For a reviewer, in the order a line is read:
//   - shell_split.go: the line cut into simple commands and words by the quoting of the shell that runs it (bash's
//     for Bash and Monitor, PowerShell's for PowerShell, cmd.exe's inside cmd /c), nested text split again to a
//     depth of three. A line it cannot split is refused only when it holds a delete word (delete-unsplit).
//   - shell_words.go: past NAME=value words and the words that run another program (RunnerWords: sudo, env, xargs,
//     ...), the program; a command still not read after 256 of them (delSteps) is refused (delete-unsplit). A shell
//     given text (bash -c, eval, cmd /c, powershell -Command, a here-document fed to bash, and script, su, runuser,
//     flock and sg's command strings: delCmdRunners) has that text split again;
//     a delete word (DeleteWords, the list's one home) gives what it deletes; a bonsai call goes to the guard's
//     judgment of Bonsai's own commands (commands.go's ownCommand).
//   - shell_env.go: CLAUDE_CODE_CHILD_SESSION read on the words of every simple command, at every depth; a change of
//     it, or a program started with an empty or new environment, goes to commands.go's personStep.
//   - this file: every text split (the line, and each text nested in it or decoded); the list walked in order, the
//     folder each command runs in followed
//     through cd (note 4: a cd to a name moves it; one that cannot be told leaves it unknown; a cd that may fail
//     leaves both folders), and each delete judged. Rule 3 (delete-unnamed): a recursive or bulk delete with an
//     operand that is not a plain name (shell_paths.go's delPlain), a delete of several files by a glob, a delete fed
//     its operands by xargs, a pipeline or a loop's variable, git clean with no path, a git pathspec with magic other
//     than :/, :(top) or :(literal) (delGitMagic, which reads those three), and always git clean -x or -X, git stash
//     --all and rm --no-preserve-root. bash's ~+/x is read in the command's folder, ~-/x in one that cannot be told. Rule 4 (delete-protected): every operand, and both sides of a move,
//     read in the command's folder and judged by the guard's one function for a delete (may.go's mayDelete); a
//     recursive delete or a move of a folder also against the guarded paths under it, found from the guarded globs
//     (may.go's judgedRoots, read only once a delete needs them: it reads this machine's records) without reading
//     the folder's tree (shell_paths.go). One exception: git rm --cached of .bonsai/local/, or a path in it, at its
//     checkout's top, named plainly, with engine.UntrackLocal's own flags only (-r, -q, --cached, --: delUntracks,
//     since a later --no-cached deletes from disk), changes git's index only and is let through (localIndex): check's
//     finding for a tracked one names it (engine.UntrackLocal).
//
// What it leaves (note 5; question A): a shell write to a guarded path (a redirect, cp or ln over it, an
// interpreter such as python -c or node -e) is not judged, nor a delete inside a script run by name, nor, in
// another linked project, a named delete below its top and floor folders (5.3.2 rule 8); rung 0 and the Desk catch
// them after the fact. Nor, being a splitter and not a shell, does it see: deletes made by a program's own calls
// ([IO.File]::Delete, $item.Delete(), Start-Process with arguments, shutil.rmtree) or by programs outside its
// tables (rimraf, trash, GNU parallel, forfiles); one file deleted through a variable outside a loop (rm "$tmp"),
// which is judged as written; a command assembled in a variable across the line (c=rm; $c x) or named by an alias
// or function defined outside it; and a link made earlier in the same line and deleted through (ln -s .git g && rm
// -rf g/), since the check reads the disk as it is before the line runs.
//
// Known limits, fixed after 1.0 (the director, 11 Oct; V2c's items 3 and 6), each adversarial; the commands' own refusal
// (5.3.6b) and rung 0 (5.4) stand behind them:
//   - env's own -S splitting rules: \_ (env's space) is read as bash reads it, and a -S text holding ;, |, && or ( is
//     read as a command line, not as env's words;
//   - env -v echo '...' | bash: a literal piped into a shell is read only when echo, printf or a here-document's cat
//     writes it, not through a runner word;
//   - the git icase variable named with quotes or encoded (GIT_ICASE_PATH""SPECS=1), read on the line's raw text;
//   - a planted .git folder holding HEAD, objects and refs whose HEAD git does not take: it ends the walk up to a
//     checkout's top (gitEntry) where git walks on;
//   - parallel as a runner word (GNU parallel ... ::: x): what it runs is not read;
//   - other runners that start a command of their own, confirmed unread by V1's second round (16 Oct): tmux and
//     screen (a session's command or send-keys), expect -c (its script's spawn), systemd-run, nsenter, ssh (to this
//     machine or another), strace and ltrace (the traced command), at (the queued job), firejail and bwrap (the
//     sandboxed command): each is a program outside the tables, its command not read. They are left, not chased: the
//     guard is a tripwire against an agent's accidents, not a wall against intent (spec section 7, decision D), and
//     rung 0 and the landing rule stand behind it;
//   - a git alias defined in the line (git -c 'alias.x=!rm -rf .bonsai/local' x, or one set through
//     GIT_CONFIG_PARAMETERS): the alias's shell text is not read, and x is judged as a git subcommand of that name;
//     left for the same reason;
//   - bonsai's words from a variable, a substitution, or a positional parameter outside xargs (bonsai $X);
//   - env -C or sudo -D given twice, or env -C and a -S text's -C: read together, where the program takes the last;
//   - ~-, ~N and ~-N after a line's own cd or pushd: judged in the payload's folder only, not every folder the line
//     visited (a plain delete of an operand that cannot be placed stays allowed, rule 3's scope);
//   - the program itself from xargs's input (echo bonsai | xargs -I{} sh -c '{} unlink', echo bonsai | xargs sh -c
//     '$0 unlink'): a program the check cannot see is not read;
//   - cd and pushd through CDPATH, or under shopt -s cdable_vars: a bare cd <name> is read as ./<name>, where bash may
//     go elsewhere (CDPATH=.. cd s && rm -rf .git deletes the checkout's own .git); the 5.3 end verifier, 12 Oct;
//   - Git Bash's /proc/cygdrive/<drive>/... (a link to /): read as a plain path on no project, not as the drive's
//     (rm -rf /proc/cygdrive/c/.../.git); the 5.3 end verifier, 12 Oct;
//   - printf or echo with escapes (\xNN, echo -e) or a format (%s) piped into a shell: the text the shell runs is not
//     rebuilt; Start-Process -ArgumentList's command text is not read.
//
// Known wrong refusal: rm of a link to itself is refused (fail closed).
//
// The answer (note 6) names the command by its head only (redact.CommandHead: the program and, for git and its kind,
// one subcommand word), the operand and the rule, and the next step; it quotes no command line, and an operand that
// holds a variable or a substitution is not quoted either (contract §2.6: the record keeps no part of the line beyond
// the head). The deciding operand is the record's target, relative to the session's project ("" outside it).
//
// The rules (rules.go's table, their one home): delete-unnamed, delete-protected and delete-unsplit refuse;
// delete-named allows a call whose deletes all name what they take; a call with no delete is allowed as
// bonsai-command when it ran a bonsai command, else as shell-not-judged.

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// delJudge is what the delete check asks of the guard's rules (plan-5 5.3.2 notes 1, 7 and 8); the guard's judge
// (rule.go) is one, a test's stub another.
type delJudge interface {
	// mayDelete judges a delete or move of the path p, read against the folder dir (absolute), as an edit of that
	// path: the one function (may.go), for which the floor's .git is the project's own top one only.
	mayDelete(dir, p string) Decision
	// judgedRoots gives the folders whose guarded paths a delete can take (judgedRoot): the projects a shell call is
	// judged in and Bonsai's home. It reads this machine's records, so it is asked only once a delete needs it.
	judgedRoots() []judgedRoot
	// ownCommand judges one bonsai call: argv its words, dir the folder it runs in, literal true when every word
	// after the program is literal (no variable, substitution, glob or brace list) (commands.go).
	ownCommand(dir string, argv []string, literal bool) Decision
	// personStep judges a person's step the check found (commands.go): a change of CLAUDE_CODE_CHILD_SESSION
	// (shell_env.go), a bonsai call whose words it cannot see.
	personStep(tool, what string) Decision
	// roleShell gives the roles' part of the call's judgment (rolecmds.go), nil where no role rule applies.
	roleShell() *shellRole
}

// delOp is one operand of a delete or a move: one word, or one word its brace list writes out.
type delOp struct {
	text     string
	expand   bool    // a variable or substitution
	glob     bool    // a glob character
	literal  bool    // given as -LiteralPath: its glob characters are literal
	git      bool    // a git pathspec: a leading : is pathspec magic
	pipeVar  bool    // PowerShell's $_ or $PSItem: an item a pipeline or loop gives
	item     bool    // find's {}: judged as find's starting points are, not here
	word     *shWord // the word it came from, when it holds an expansion that may stand for literal text
	top      bool    // read against the checkout's top, not the command's folder (git's :/x, :(top)x)
	away     bool    // read in a folder the check cannot tell (bash's ~-/x, the previous folder)
	alsoHere bool    // with away: read in the command's folder too (bash's ~-0, the command's folder until a pushd)
}

// delAct is one delete or move a command makes.
type delAct struct {
	word      string  // the delete word as the table names it: "rm", "git clean", "Remove-Item"
	move      bool    // a move: its operands are taken away, and its destinations receive them
	recursive bool    // a recursive delete (-r, -Recurse, /s)
	bulk      string  // a bulk form, named ("find", "git clean", "rsync --delete", "-Include"); "" for none
	tree      bool    // each operand goes with everything it holds
	unnamed   string  // set: refused as it stands; the reason
	next      string  // the refusal's next step, with unnamed
	ops       []delOp // what it deletes or moves away
	dests     []delOp // a move's destinations
	destDir   bool    // a destination may be a folder the sources go into (dest/<name> is judged too)
	rename    bool    // Rename-Item, ren: the destination is a new name beside the source
	cached    bool    // git rm --cached: it takes its operands out of git's index only, and the files stay on disk
}

// shellDecide judges one shell call by the delete check: tool is Bash, Monitor or PowerShell (any other name is read
// as bash), command its command line, cwd the payload's folder, base the session's project folder (where the line
// runs when cwd is not absolute, and what a refusal's target is relative to). No input makes it panic
// (FuzzShellDecide).
func shellDecide(tool, command, cwd, base string, j delJudge) Decision {
	d := shBash
	if tool == "PowerShell" {
		d = shPower
	}
	r := &delRun{tool: tool, j: j, own: base, base: base, icase: strings.Contains(strings.ToUpper(command), "GIT_ICASE_PATHSPECS")}
	if cwd != "" && filepath.IsAbs(cwd) {
		r.base = filepath.Clean(cwd)
	}
	start := delDirs{}
	if r.base != "" {
		start.known = []string{r.base}
	} else {
		start.unknown = true
	}
	r.text(command, d, 0, start, false)
	switch {
	case r.deny != nil:
		return *r.deny
	case r.found:
		return allow(RuleDeleteNamed, r.first)
	case r.bonsai:
		return allow(RuleBonsaiCommand, "")
	}
	return allow(RuleShell, "")
}

// delRun is one judgment of a command line.
type delRun struct {
	tool     string
	j        delJudge
	own      string    // the session's project folder: a refusal's target is relative to it, "" for none
	base     string    // the payload's folder: where an operand is read when its folder cannot be told
	found    bool      // a delete was judged
	bonsai   bool      // a bonsai command was judged, and allowed
	first    string    // the first operand judged, project-relative, for an allow's target
	firstSet bool      // first is set (it may be "": an operand outside the project)
	feed     string    // set while splitting text a shell runs under xargs: its deletes' operands come from input
	places   []string  // with feed: the texts xargs's input replaces in the words (delExec's places)
	icase    bool      // the line names GIT_ICASE_PATHSPECS: git may read every pathspec case-blind
	deny     *Decision // the first refusal

	roots     []judgedRoot // the folders a delete is judged in, read once, when a delete first needs them
	rootsSet  bool         // roots has been read
	rootForms [][]string   // per root, its forms (delRootForms), read once per line; nil until read
	rootFold  []bool       // per root, whether letter case is folded there
	rootErr   error        // why a root's forms could not be read
}

// rootsRead reads the folders a delete is judged in, and each one's forms, once per line: only when a recursive
// delete, a move or a glob needs them.
func (r *delRun) rootsRead() error {
	if r.rootForms != nil || r.rootErr != nil {
		return r.rootErr
	}
	if !r.rootsSet {
		r.roots, r.rootsSet = r.j.judgedRoots(), true
	}
	r.rootForms = make([][]string, len(r.roots))
	r.rootFold = make([]bool, len(r.roots))
	for k, root := range r.roots {
		f, err := delRootForms(root.Dir)
		if err != nil {
			r.rootErr = err
			return err
		}
		r.rootForms[k], r.rootFold[k] = f, caseInsensitive(root.Dir)
	}
	return nil
}

func (r *delRun) refuse(d Decision) {
	if r.deny == nil {
		r.deny = &d
	}
}

// text splits and judges a text read by dialect d at nesting depth depth.
func (r *delRun) text(s string, d shDialect, depth int, dirs delDirs, loop bool) delDirs {
	if r.deny != nil {
		return dirs
	}
	l, why := shSplit(s, d, depth)
	if why != "" {
		if envBackstop(s) {
			r.envStep("holds CLAUDE_CODE_CHILD in a form the check cannot split")
			return dirs
		}
		if b := delHoldsBonsai(s); b != "" {
			r.unsplitBonsai(b)
			return dirs
		}
		if word := delHoldsWord(s); word != "" {
			r.refuse(Decision{Rule: RuleDeleteUnsplit,
				Why: fmt.Sprintf("%s refused: the command line holds the delete word %s, and the check cannot split it (%s), "+
					"so it cannot tell what it deletes (rule %s)", r.tool, quote(word, 40), why, RuleDeleteUnsplit),
				Next: "close every quote and bracket, nest at most three levels deep, and name what the delete removes"})
		}
		return dirs
	}
	if envBackstop(s) && envSightings(l) == 0 {
		r.envStep("holds CLAUDE_CODE_CHILD in a form the check cannot read (the name built from parts)")
		return dirs
	}
	return r.list(l, d, dirs, loop)
}

// list walks a command list in order, following the folder through cd, and gives the folders it may end in.
func (r *delRun) list(l *shList, d shDialect, dirs delDirs, loop bool) delDirs {
	cur := dirs
	var chain []delDirs // the folders before each cd in an && chain: the chain may stop at any of them
	var stack []delDirs // pushd's
	loops := 0          // bash's do ... done
	var prevOut []string
	for k := range l.cmds {
		if r.deny != nil {
			return cur
		}
		c := &l.cmds[k]
		inLoop := loop || loops > 0
		nestLoop := inLoop || delLoopCmd(c, d)
		for _, n := range c.nested {
			end := r.list(n.list, d, cur, nestLoop)
			if !n.sub {
				cur = cur.union(end)
			}
		}
		for _, f := range c.feeds {
			for _, n := range f.nested {
				r.list(n.list, d, cur, inLoop)
			}
		}
		if d == shPower {
			r.envDotNet(c)
		}
		if rs := r.j.roleShell(); rs.agent() {
			if r.redirects(rs, c, d, cur, l.depth); r.deny != nil {
				return cur
			}
		}
		var cd delCd
		if c.group != nil {
			end := r.list(c.group, d, cur, inLoop)
			if !c.sub {
				cd = delCd{kind: delCdTo, to: end}
			}
		} else {
			var stdin []string
			for _, f := range c.feeds {
				stdin = append(stdin, f.text)
			}
			if c.piped {
				stdin = append(stdin, prevOut...)
			}
			cd = r.simple(c, d, l.depth, cur, inLoop, stdin, &loops)
		}
		prevOut = delOutput(c, d)
		switch cd.kind {
		case delCdTo:
			next := cd.to
			if c.op == "&&" {
				chain = append(chain, cur)
				cur = next
			} else {
				cur = cur.union(next)
			}
		case delCdPush:
			stack = append(stack, cur)
			if c.op == "&&" {
				chain = append(chain, cur)
				cur = cd.to
			} else {
				cur = cur.union(cd.to)
			}
		case delCdPop:
			if n := len(stack); n > 0 {
				cur = cur.union(stack[n-1])
				stack = stack[:n-1]
			} else {
				cur = cur.union(delDirs{unknown: true})
			}
		}
		if c.op != "&&" {
			for _, ch := range chain {
				cur = cur.union(ch)
			}
			chain = nil
		}
	}
	return cur
}

// delCd is what a command did to the folder.
type delCd struct {
	kind int
	to   delDirs
}

const (
	delCdNone = iota
	delCdTo   // cd, Set-Location; a brace group's end
	delCdPush // pushd, Push-Location
	delCdPop  // popd, Pop-Location
)

// delOutput gives the text a command writes when it is a literal one (echo 'x', printf, cat of a here-document, a
// PowerShell string), for a shell it pipes into; nil for any other.
func delOutput(c *shCommand, d shDialect) []string {
	if c.group != nil || len(c.words) == 0 {
		return nil
	}
	w := c.words
	name := strings.ToLower(w[0].text)
	switch {
	case d == shPower && len(w) == 1 && w[0].quoted && !w[0].param:
		return []string{w[0].text}
	case name == "echo" || name == "printf" || d == shPower && (name == "write-output" || name == "write"):
		var parts []string
		for _, x := range w[1:] {
			if d == shBash && name == "echo" && len(parts) == 0 && (x.text == "-n" || x.text == "-e" || x.text == "-E") {
				continue
			}
			parts = append(parts, x.text)
		}
		return []string{strings.Join(parts, " ")}
	case name == "cat" && len(w) == 1:
		var out []string
		for _, f := range c.feeds {
			out = append(out, f.text)
		}
		return out
	}
	return nil
}

// delExec is a command from its program on, as it runs.
type delExec struct {
	words  []shWord
	shown  string // the simple command as written, for an answer
	d      shDialect
	dirs   delDirs
	depth  int      // the depth of the list it stands in
	loop   bool     // inside a loop's body
	feed   string   // set when a runner feeds the program its operands from input ("xargs"), which the check cannot see
	places []string // with feed: the texts xargs's input replaces in the words (-I's, -i's and --replace's)
	find   bool     // run by find -exec: its own operands are a bulk delete's
	icase  bool     // git was given --icase-pathspecs
	piped  bool     // fed by a pipe
	stdin  []string // text fed to it: here-documents, here-strings, a literal piped in
}

// simple judges one simple command, after the shell's own words that come before a program (bash's if, do, time;
// PowerShell's assignment and call operator; cmd's @, if and for), and gives what it did to the folder.
func (r *delRun) simple(c *shCommand, d shDialect, depth int, dirs delDirs, loop bool, stdin []string, loops *int) delCd {
	words := c.words
	var shown []string
	for _, w := range words {
		shown = append(shown, w.raw)
	}
	x := delExec{shown: strings.Join(shown, " "), d: d, dirs: dirs, depth: depth, loop: loop, piped: c.piped, stdin: stdin}
	switch d {
	case shBash:
		words = delExpandAlts(words) // {rm,-rf,x} is the command rm -rf x
	prelude:
		for len(words) > 0 && !words[0].quoted {
			switch words[0].text {
			case "!", "if", "then", "elif", "else", "while", "until", "{", "}", "coproc":
				words = words[1:]
			case "do":
				*loops++
				x.loop = true
				words = words[1:]
			case "done":
				if *loops > 0 {
					*loops--
				}
				words = words[1:]
			case "fi", "esac":
				words = words[1:]
			case "time":
				words = words[1:]
				if len(words) > 0 && words[0].text == "-p" {
					words = words[1:]
				}
			case "for", "select", "case", "function", "[[", "((":
				return delCd{}
			default:
				break prelude
			}
		}
	case shPower:
		if rs := r.j.roleShell(); rs.agent() {
			// (Get-Item x).IsReadOnly = $true: a change of x's attributes (attrwords.go), read before the left side goes.
			if ts := psMemberSet(words, depth); len(ts) > 0 {
				if r.writes(rs, x, ts); r.deny != nil {
					return delCd{}
				}
			}
		}
		r.envPSAssign(words)
		words = delPSAssign(words)
		for len(words) > 0 && !words[0].quoted && (words[0].text == "&" || words[0].text == ".") {
			words = words[1:]
		}
	case shCmd:
		words = delCmdPrelude(words, &x)
	}
	x.words = words
	return r.exec(x)
}

// delLoopCmd reports a command whose blocks are a loop's body: PowerShell's foreach, for, while, do and
// ForEach-Object (%), cmd.exe's for.
func delLoopCmd(c *shCommand, d shDialect) bool {
	words := c.words
	if d == shPower {
		words = delPSAssign(words)
	}
	if len(words) == 0 || words[0].quoted {
		return false
	}
	switch strings.ToLower(words[0].text) {
	case "foreach", "for", "while", "do", "foreach-object", "%":
		return d == shPower || d == shCmd && strings.EqualFold(words[0].text, "for")
	}
	return false
}

// delPSAssign drops PowerShell assignments' left sides, a chain's too ($x = ..., $x += ..., $x =..., $x = $y = ...),
// so the program after them is read.
func delPSAssign(words []shWord) []shWord {
	for {
		next, ok := delPSAssignOne(words)
		if !ok {
			return words
		}
		words = next
	}
}

// delPSAssignOne drops one assignment's left side, its operator glued to it, written as the next word, or glued to
// what it assigns after a space ($x =Remove-Item x runs Remove-Item); ok is false when words start with none.
func delPSAssignOne(words []shWord) ([]shWord, bool) {
	if len(words) == 0 || words[0].quoted || !strings.HasPrefix(words[0].text, "$") {
		return words, false
	}
	t := words[0].text
	if k := strings.IndexByte(t, '='); k > 0 {
		rest := strings.TrimLeft(t[k+1:], "=")
		if rest == "" {
			return words[1:], true
		}
		w := words[0]
		w.text, w.raw, w.expand = rest, rest, strings.Contains(rest, "$")
		return append([]shWord{w}, words[1:]...), true
	}
	if len(words) < 2 {
		return words, false
	}
	op := envPSOperator(strings.ToLower(words[1].raw))
	if op == "" {
		return words, false
	}
	_, rest, _ := strings.Cut(words[1].text, "=")
	if rest == "" {
		return words[2:], true
	}
	w := words[1]
	w.text, w.raw, w.expand = rest, w.raw[len(op):], strings.Contains(rest, "$")
	return append([]shWord{w}, words[2:]...), true
}

// delCmdPrelude reads cmd.exe's words before a program: @, if's condition, for's loop up to do.
func delCmdPrelude(words []shWord, x *delExec) []shWord {
	if len(words) > 0 && strings.HasPrefix(words[0].text, "@") && !words[0].quoted {
		w := words[0]
		w.text = strings.TrimLeft(w.text, "@")
		if w.text == "" {
			words = words[1:]
		} else {
			words = append([]shWord{w}, words[1:]...)
		}
	}
	for len(words) > 0 {
		switch strings.ToLower(words[0].text) {
		case "if":
			words = words[1:]
			for len(words) > 0 && (strings.EqualFold(words[0].text, "/i") || strings.EqualFold(words[0].text, "not")) {
				words = words[1:]
			}
			if len(words) == 0 {
				return nil
			}
			switch strings.ToLower(words[0].text) {
			case "exist", "defined", "errorlevel", "cmdextversion":
				words = words[min(2, len(words)):]
			default:
				if len(words) >= 3 && delCmdCompare(words[1].text) {
					words = words[3:]
				} else {
					words = words[1:] // a==b in one word
				}
			}
		case "for":
			at := -1
			for k, w := range words {
				if strings.EqualFold(w.text, "do") {
					at = k
					break
				}
			}
			if at < 0 {
				return nil
			}
			x.loop = true
			words = words[at+1:]
		case "else", "do":
			words = words[1:]
		default:
			return words
		}
	}
	return words
}

func delCmdCompare(op string) bool {
	switch strings.ToLower(op) {
	case "==", "equ", "neq", "lss", "leq", "gtr", "geq":
		return true
	}
	return false
}

// judge judges one delete or move, in the folders the command may run in.
func (r *delRun) judge(a delAct, x delExec) {
	if r.deny != nil {
		return
	}
	r.found = true
	if a.unnamed != "" {
		r.refuse(Decision{Rule: RuleDeleteUnnamed, Why: fmt.Sprintf("%s refused: %s does not name what it deletes: %s "+
			"(rule %s)", r.tool, r.head(x), a.unnamed, RuleDeleteUnnamed), Next: a.next})
		return
	}
	verb := "deletes"
	if a.move {
		verb = "moves"
	}
	if x.feed == "" {
		x.feed = r.feed
	}
	if x.feed != "" {
		r.refuse(Decision{Rule: RuleDeleteUnnamed, Why: fmt.Sprintf("%s refused: %s does not name what it %s: its "+
			"operands come from %s, so the check cannot see them (rule %s)", r.tool, r.head(x), verb, x.feed,
			RuleDeleteUnnamed), Next: "name what to " + delVerb(a.move) + " in the command itself, not through " + x.feed})
		return
	}
	a.ops = delLiterals(a.ops, x)
	a.dests = delLiterals(a.dests, x)
	for _, list := range [][]delOp{a.ops, a.dests} {
		for k := range list {
			op := &list[k]
			if x.d == shBash {
				delTildeOp(op)
			}
			if op.git && (r.icase || x.icase) {
				r.refuse(Decision{Rule: RuleDeleteUnnamed, Target: r.target(x.dirs, *op),
					Why: fmt.Sprintf("%s refused: %s does not name what it %s: git reads its pathspecs case-blind here "+
						"(--icase-pathspecs or GIT_ICASE_PATHSPECS), which the check does not read (rule %s)", r.tool,
						r.head(x), verb, RuleDeleteUnnamed),
					Next: "run the git command without --icase-pathspecs or GIT_ICASE_PATHSPECS, naming the paths as they are"})
				return
			}
			if !op.git || !strings.HasPrefix(op.text, ":") {
				continue
			}
			if why := delGitMagic(op); why != "" {
				r.refuse(Decision{Rule: RuleDeleteUnnamed, Target: r.target(x.dirs, *op),
					Why: fmt.Sprintf("%s refused: %s does not name what it %s: its operand %s %s (rule %s)", r.tool,
						r.head(x), verb, r.shownText(op.text), why, RuleDeleteUnnamed),
					Next: "name the paths plainly, or with :/ or :(top) for the checkout's top"})
				return
			}
		}
	}
	bulk := a.bulk != "" || x.find
	for _, op := range a.ops {
		if op.item {
			continue
		}
		why, next := "", ""
		switch {
		case op.pipeVar || x.loop && op.expand:
			why = fmt.Sprintf("%s takes its operand from a pipeline or a loop's variable, one item after another, so "+
				"the check cannot see what it %s", a.word, verb)
			next = "name the files in the command, or " + delVerb(a.move) + " their folder by its name"
		case (a.recursive || bulk) && !a.move:
			if p := delPlain(op); p != "" {
				why = fmt.Sprintf("its operand %s %s", r.shownOp(op), p)
				next = delNextFor(p)
			} else if (x.dirs.unknown || op.away) && !delIsAbs(op.text) {
				why = fmt.Sprintf("its operand %s is read in a folder the check cannot tell (after a cd it cannot "+
					"follow)", r.shownOp(op))
				next = "cd to a plain folder name first, or give the operand as an absolute path"
			}
		case op.glob && !op.literal && !a.move:
			why = fmt.Sprintf("its operand %s holds a glob, which deletes several files it does not name", r.shownOp(op))
			next = "name the files, or delete the folder by its name"
		}
		if why != "" {
			r.refuse(Decision{Rule: RuleDeleteUnnamed, Target: r.target(x.dirs, op),
				Why:  fmt.Sprintf("%s refused: %s does not name what it %s: %s (rule %s)", r.tool, r.head(x), verb, why, RuleDeleteUnnamed),
				Next: next})
			return
		}
	}
	for _, op := range a.ops {
		if op.text == "" || op.expand || op.item || a.cached && r.localIndex(x, op) {
			continue
		}
		r.path(x, op, a.tree, verb)
		if r.deny != nil {
			return
		}
	}
	if a.move {
		for _, dest := range a.dests {
			if dest.text == "" || dest.expand || dest.item {
				continue
			}
			r.path(x, dest, false, "moves onto")
			if a.destDir {
				for _, src := range a.ops {
					base := delBase(src.text)
					if base == "" || src.expand || src.glob || src.item {
						continue // a glob's names are not known: what it may land on is a shell write (note 5)
					}
					in := dest
					in.text = strings.TrimRight(dest.text, `/\`) + "/" + base
					in.glob = dest.glob || src.glob
					r.path(x, in, false, "moves onto")
				}
			}
			if r.deny != nil {
				return
			}
		}
	}
	// A role that changes nothing deletes and moves nothing inside a linked project (rolecmds.go), after the floor's
	// own judgment of each path above.
	if rs := r.j.roleShell(); rs.changesNothing() {
		for _, op := range append(append([]delOp(nil), a.ops...), a.dests...) {
			if r.writePath(rs, x, writeTarget{op: op, word: a.word}, true); r.deny != nil {
				return
			}
		}
	}
}

// delLiterals reads in place of an operand that is a substitution of literal output ("$(echo x)", PowerShell's
// ('x')) the text it stands for.
func delLiterals(ops []delOp, x delExec) []delOp {
	out := make([]delOp, 0, len(ops))
	for _, op := range ops {
		if !op.expand || op.word == nil {
			out = append(out, op)
			continue
		}
		text, ok := delLiteral(*op.word, x.d, x.depth)
		if !ok {
			out = append(out, op)
			continue
		}
		texts := []string{text}
		if x.d == shBash && !strings.HasPrefix(strings.TrimSpace(op.word.raw), `"`) {
			texts = strings.Fields(text) // unquoted, bash splits the output into words
		}
		for _, t := range texts {
			n := op
			n.text, n.expand, n.word, n.pipeVar = t, false, nil, false
			n.glob = strings.ContainsAny(t, "*?[")
			out = append(out, n)
		}
	}
	return out
}

// delLiteral gives the text a word stands for when the word is one substitution of a command whose output is
// literal (delOutput): "$(echo x)", `printf x`, <(cat <<EOF ... EOF), PowerShell's ('x') and $('x'). The check reads
// that text in the word's place; any other word gives false.
func delLiteral(w shWord, d shDialect, depth int) (string, bool) {
	raw := strings.TrimSpace(w.raw)
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	var inner string
	switch {
	case len(raw) >= 3 && (strings.HasPrefix(raw, "$(") || strings.HasPrefix(raw, "<(") || strings.HasPrefix(raw, ">(")) &&
		!strings.HasPrefix(raw, "$((") && strings.HasSuffix(raw, ")"):
		inner = raw[2 : len(raw)-1]
	case d == shBash && len(raw) >= 2 && raw[0] == '`' && raw[len(raw)-1] == '`':
		inner = strings.NewReplacer("\\`", "`", "\\$", "$", "\\\\", "\\").Replace(raw[1 : len(raw)-1])
	case d == shPower && len(raw) >= 2 && (raw[0] == '(' || strings.HasPrefix(raw, "@(")) && raw[len(raw)-1] == ')':
		inner = strings.TrimPrefix(raw, "@")[1 : len(strings.TrimPrefix(raw, "@"))-1]
	default:
		return "", false
	}
	l, why := shSplit(inner, d, depth+1)
	if why != "" || len(l.cmds) != 1 {
		return "", false
	}
	out := delOutput(&l.cmds[0], d)
	if len(out) != 1 {
		return "", false
	}
	return out[0], true
}

// delExpandAlts writes out bash's brace lists in place: each word a list gives is a word of the command.
func delExpandAlts(words []shWord) []shWord {
	var out []shWord
	for _, w := range words {
		if w.alts == nil {
			out = append(out, w)
			continue
		}
		for _, a := range w.alts {
			out = append(out, shWord{text: a, raw: a, quoted: w.quoted, glob: strings.ContainsAny(a, "*?["), expand: w.expand})
		}
	}
	return out
}

// delGitMagic reads a git pathspec that starts with a colon (its magic) in place: :/x, :/:x and :(top)x name x at the
// checkout's top (top); :x, ::x and :(literal)x name x in the command's folder (:(literal)'s glob characters as
// written). Any other magic (icase, glob, exclude and its ! and ^, attr, or more than one) gives why it is refused:
// the check cannot tell which paths it takes.
func delGitMagic(op *delOp) string {
	t := op.text
	const other = "is a git pathspec with magic the check does not read (icase, glob, exclude, attr, or more than one)"
	var magic []string
	rest := ""
	if strings.HasPrefix(t, ":(") {
		end := strings.IndexByte(t, ')')
		if end < 0 {
			return other
		}
		for _, m := range strings.Split(t[2:end], ",") {
			if m = strings.ToLower(strings.TrimSpace(m)); m != "" {
				magic = append(magic, m)
			}
		}
		rest = t[end+1:]
	} else {
		k := 1
		for k < len(t) && strings.IndexByte("/!^", t[k]) >= 0 {
			if t[k] == '/' {
				magic = appendNew(magic, "top")
			} else {
				magic = appendNew(magic, "exclude")
			}
			k++
		}
		if k < len(t) && t[k] == ':' {
			k++
		}
		rest = t[k:]
	}
	if len(magic) > 1 {
		return other
	}
	if len(magic) == 1 {
		switch magic[0] {
		case "top":
			op.top = true
		case "literal":
			op.literal = true
		default:
			return other
		}
	}
	if rest == "" {
		rest = "."
	}
	op.text = rest
	op.glob = !op.literal && strings.ContainsAny(rest, "*?[")
	return ""
}

// delTildeOp reads bash's tilde forms in an operand: ~+ is the command's folder ($PWD), so ~+/x is x there; ~- is the
// previous folder ($OLDPWD), which the check cannot tell (away); of the folder stack's forms, ~0 and ~+0 are the
// command's folder, ~-0 (the stack's bottom: the command's folder until a pushd, then the folder before it) both the
// command's folder and one the check cannot tell, and ~N, ~+N, ~-N for any other N folders the check cannot tell;
// ~name/x is x in the home folder when name is the home folder's own (delTildeUser).
func delTildeOp(op *delOp) {
	t := op.text
	if !strings.HasPrefix(t, "~") {
		return
	}
	head, rest, slashed := strings.Cut(t[1:], "/")
	rel := "."
	if slashed {
		rel = "./" + rest // never the root: ~+//x is x in the folder
	}
	switch here, ok := delTildeStack(head); {
	case ok && here:
		op.text = rel
	case ok:
		op.text, op.away, op.alsoHere = rel, true, delTildeBottom(head)
	case slashed && head != "":
		if abs := delTildeUser("", t); abs != "" {
			op.text = abs
		}
	}
}

// delTildeStack reads a tilde's word as bash's folder stack does: ok for +, -, +N, -N and N, N read as a number
// (00 is 0); here when it is the command's folder (+, 0, +0), not when it is another the check cannot tell (-, -0 the
// stack's bottom, which a pushd leaves behind, and any other N).
func delTildeStack(head string) (here, ok bool) {
	switch head {
	case "+":
		return true, true
	case "-":
		return false, true
	}
	n := strings.TrimLeft(head, "+-")
	if n == "" || len(head)-len(n) > 1 || strings.Trim(n, "0123456789") != "" {
		return false, false
	}
	return strings.Trim(n, "0") == "" && !strings.HasPrefix(head, "-"), true
}

// delTildeBottom reports bash's ~-0 (any count of zeros): the folder stack's bottom, the command's folder until a
// pushd.
func delTildeBottom(head string) bool {
	return len(head) > 1 && head[0] == '-' && strings.Trim(head[1:], "0") == ""
}

func delVerb(move bool) string {
	if move {
		return "move"
	}
	return "delete"
}

// delBase is a path's last name, "" when it has none.
func delBase(p string) string {
	s := strings.TrimRight(strings.ReplaceAll(p, `\`, "/"), "/")
	if k := strings.LastIndexByte(s, '/'); k >= 0 {
		s = s[k+1:]
	}
	if s == "." || s == ".." || s == "~" {
		return ""
	}
	return s
}

// delNextFor gives the next step for an operand that is not a plain name.
func delNextFor(reason string) string {
	switch {
	case strings.Contains(reason, "glob"):
		return "name the files, or delete the folder by its name"
	case strings.Contains(reason, "variable"):
		return "write the path out in the command (no variable or substitution), or delete the folder by its plain name"
	case strings.Contains(reason, "magic"):
		return "name the paths plainly, without git's pathspec magic"
	}
	return "name the folder to delete by its own name (rm -r build), not by ., .., ~ or a root"
}

// dirsFor gives the folders an operand is read in: the command's known ones, and the payload's when one cannot be
// told (note 4: "judged as if in the payload's folder"); one folder for an operand that needs none.
func (r *delRun) dirsFor(dirs delDirs, p string) []string {
	var out []string
	if delIsAbs(p) {
		if len(dirs.known) > 0 {
			return dirs.known[:1]
		}
		if r.base != "" {
			return []string{r.base}
		}
		return []string{string(filepath.Separator)}
	}
	out = append(out, dirs.known...)
	if (dirs.unknown || len(out) == 0) && r.base != "" {
		out = appendNew(out, r.base)
	}
	return out
}

// path judges one operand by rule 4: as an edit of its path, and with tree against the guarded paths under it.
func (r *delRun) path(x delExec, op delOp, tree bool, verb string) {
	text := op.text
	if x.d != shBash && filepath.Separator == '/' {
		text = strings.ReplaceAll(text, `\`, "/") // PowerShell and cmd read \ as a separator on every system
	}
	if x.d == shPower {
		text = delProvider(text)
	}
	glob := op.glob && !op.literal
	dirs := r.dirsFor(x.dirs, text)
	switch {
	case op.away && r.base != "" && op.alsoHere:
		dirs = appendNew(dirs, r.base)
	case op.away && r.base != "":
		dirs = []string{r.base} // judged as if in the payload's folder (note 4)
	case op.top:
		var tops []string
		for _, d := range dirs {
			tops = appendNew(tops, checkoutTop(d))
		}
		dirs = tops
	}
	for _, dir := range dirs {
		if !r.firstSet {
			r.first, r.firstSet = delRel(r.own, dir, text), true
		}
		if !glob {
			if dec := r.j.mayDelete(dir, text); !dec.Allow {
				r.protected(x, dir, text, "", dec, verb)
				return
			}
		}
		if !tree && !glob || tree && !glob && !delFolder(dir, text) {
			continue // a file holds nothing, and what does not exist is not deleted: no folder to meet the globs with
		}
		forms, err := delForms(dir, text)
		if err == nil {
			err = r.rootsRead()
		}
		if err != nil {
			r.refuse(Decision{Rule: RuleBadPath, Target: delRel(r.own, dir, text),
				Why: fmt.Sprintf("%s refused: %s %s %s, and the check cannot tell where it leads (%s), so it cannot "+
					"tell what it %s (rule %s)", r.tool, r.head(x), verb, r.shownText(text), oneLine(err.Error()), verb, RuleBadPath),
				Next: "give the path in its plain form inside the project (a drive letter and folders), or tell the person"})
			return
		}
		for k, root := range r.roots {
			fold := r.rootFold[k]
			rels, whole := delPlace(r.rootForms[k], forms, fold)
			if root.Whole && tree && (whole || len(rels) > 0) {
				dec := r.j.mayDelete(root.Dir, root.Dir)
				if dec.Allow {
					dec = deny(RuleBonsaiFiles, "everything in "+quote(fromHome(root.Dir), 80)+" is Bonsai's own",
						"leave Bonsai's home as it is")
				}
				r.protected(x, dir, text, fromHome(root.Dir), dec, verb)
				return
			}
			places := rels
			if whole {
				places = append([][]string{{}}, rels...)
			}
			lastLiteral := !strings.ContainsAny(delBase(text), "*?[")
			for _, g := range root.Globs {
				fixed := delFixed(g)
				if !fixed && (!glob || !lastLiteral) {
					// Judged against the operand itself only (mayDelete, above): never against what a folder holds,
					// nor the names a glob may stand for (mv dist/* . is not refused for a dist/bonsai.yaml it may
					// hold); a glob that names the file (*/bonsai.yaml) is met.
					continue
				}
				segs := delGlobSegs(g)
				for _, p := range places {
					rep, ok := delMeet(p, segs, tree && fixed, fold)
					if !ok || len(rep) == 0 {
						continue
					}
					shownRep := strings.Join(rep, "/")
					if dec := r.j.mayDelete(root.Dir, shownRep); !dec.Allow {
						r.protected(x, dir, text, shownRep, dec, verb)
						return
					}
				}
			}
		}
	}
}

// localIndex reports an operand of git rm --cached that is .bonsai/local/, or a path inside it, at the top of the
// checkout of each folder it may be read in: taking it out of git's index changes no file (they stay on disk), and
// .bonsai/local/ is never committed (contract §3), so check's finding for a tracked one names exactly this command
// (T-5505's S1, 16 Oct: the guard refused it, and an agent could not clear the finding). Any other guarded path is
// judged as before: untracked, then committed, it is a deletion in git. A plain operand only (no glob, variable, git
// pathspec magic or folder the check cannot tell), in letters as written.
func (r *delRun) localIndex(x delExec, op delOp) bool {
	text := op.text
	if x.d != shBash {
		text = strings.ReplaceAll(text, `\`, "/")
	}
	if op.glob || op.away || op.top || op.pipeVar || delPlain(op) != "" || strings.HasPrefix(text, ":") || x.dirs.unknown {
		return false
	}
	dirs := r.dirsFor(x.dirs, text)
	if len(dirs) == 0 {
		return false
	}
	for _, dir := range dirs {
		abs := filepath.FromSlash(text)
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(dir, abs)
		}
		rel, err := filepath.Rel(checkoutTop(dir), filepath.Clean(abs))
		if rel = filepath.ToSlash(rel); err != nil || rel != workspace.LocalDir && !strings.HasPrefix(rel, workspace.LocalDir+"/") {
			return false
		}
	}
	return true
}

// protected refuses a delete of a guarded path: the operand, and the guarded path under it when that decided.
func (r *delRun) protected(x delExec, dir, text, under string, dec Decision, verb string) {
	what := r.shownText(text)
	if under != "" {
		what += ", which takes " + quote(under, 80) + " with it"
	}
	why := strings.TrimSpace(dec.Why)
	if why == "" {
		why = "the guard refuses changing it (rule " + dec.Rule + ")"
	}
	next := dec.Next
	if next == "" {
		next = "leave it as it is and tell the person"
	}
	r.refuse(Decision{Rule: RuleDeleteProtected, Target: delRel(r.own, dir, text),
		Why: fmt.Sprintf("%s refused: %s %s %s, a guarded path; the guard's answer for changing it: %s (rule %s)",
			r.tool, r.head(x), verb, what, why, RuleDeleteProtected),
		Next: next})
}

// head names a command in an answer by its head alone (redact.CommandHead), never its line: "rm", "git clean",
// "Remove-Item"; "the command" when no head is safe to keep.
func (r *delRun) head(x delExec) string {
	if h := redact.CommandHead(x.shown, x.d == shPower); h != "" {
		return quote(h, 60)
	}
	return "the command"
}

// shownOp names an operand in an answer: as shownText does, or, when it holds a variable or a substitution (a part of
// the line beyond a path), not at all.
func (r *delRun) shownOp(op delOp) string {
	if op.expand {
		return "(a variable or substitution)"
	}
	return r.shownText(op.text)
}

// shownText names a path in an answer: as written when it is relative, else from ~/ under the user's home folder
// (fromHome), quoted.
func (r *delRun) shownText(text string) string {
	if delIsAbs(text) && !strings.HasPrefix(text, "~") {
		return quote(fromHome(text), 80)
	}
	return quote(text, 80)
}

// target gives an operand's target for a refusal: project-relative in the command's first folder, or "".
func (r *delRun) target(dirs delDirs, op delOp) string {
	dir := r.base
	if len(dirs.known) > 0 {
		dir = dirs.known[0]
	}
	if dir == "" {
		return ""
	}
	return delRel(r.own, dir, op.text)
}

// delProvider drops a PowerShell provider's prefix from a path: FileSystem::C:\x is C:\x.
func delProvider(p string) string {
	if k := strings.Index(p, "::"); k > 0 && strings.Contains(strings.ToLower(p[:k]), "filesystem") {
		return p[k+2:]
	}
	return p
}
