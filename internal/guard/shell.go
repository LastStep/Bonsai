package guard

// The delete check (spec §7, "Shell commands"; plan-5 5.3.3): in a Bash, Monitor or PowerShell call, a recursive or
// bulk delete that does not name what it deletes is refused (delete-unnamed), and a delete or move of a guarded
// path is judged by the guard's rules as an edit of that path (delete-protected). Rohan's answer to question A
// (spec §19) is no shell reader: the check splits a command line into words and reads only what it needs to find
// a delete word and its operands. Nothing is expanded or run.
//
// For a reviewer, in the order a line is read:
//   - shell_split.go: the line cut into simple commands and words by the quoting of the shell that runs it (bash's
//     for Bash and Monitor, PowerShell's for PowerShell, cmd.exe's inside cmd /c), nested text split again to a
//     depth of three. A line it cannot split is refused only when it holds a delete word (delete-unsplit).
//   - shell_words.go: past NAME=value words and the words that run another program (delRunners: sudo, env, xargs,
//     ...), the program. A shell given text (bash -c, eval, cmd /c, powershell -Command, a here-document fed to
//     bash) has that text split again; a delete word (delWords, the list's one home) gives what it deletes.
//   - this file: the list walked in order, the folder each command runs in followed through cd (note 4: a cd to a
//     name moves it; one that cannot be told leaves it unknown; a cd that may fail leaves both folders), and each
//     delete judged. Rule 3 (delete-unnamed): a recursive or bulk delete with an operand that is not a plain name
//     (shell_paths.go's delPlain), a delete of several files by a glob, a delete fed its operands by xargs, a
//     pipeline or a loop's variable, git clean with no path, and always git clean -x or -X, git stash --all and
//     rm --no-preserve-root. Rule 4 (delete-protected): every operand, and both sides of a move, read in the
//     command's folder and judged by the guard's function for an edit of that path (delJudge); a recursive delete or
//     a move of a folder also against the guarded paths under it, found from the guarded globs without reading the
//     folder's tree (shell_paths.go).
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
// The answer (note 6) names the command, the operand and the rule, and the next step; the guard's record keeps the
// command's head only (redact.CommandHead) and the deciding operand, project-relative, as its target.

import (
	"fmt"
	"path/filepath"
	"strings"
)

// The delete check's rules, the guard record's rule field (contract §8.1). Until the check is wired into the
// guard's rules table (plan-5 5.3, "The order", item 5), delRules below is their one home.
const (
	delRuleUnnamed   = "delete-unnamed"   // deny: a recursive or bulk delete that does not name what it deletes
	delRuleProtected = "delete-protected" // deny: a delete or move of a guarded path, or of a folder holding one
	delRuleUnsplit   = "delete-unsplit"   // deny: a line the check cannot split, holding a delete word
	delRuleNamed     = "delete-named"     // allow: every delete names what it deletes, and none is guarded
	delRuleNone      = "shell-no-delete"  // allow: no delete word (a line that cannot be split holds none either)
)

// delRules is the delete check's part of the guard's rules table: each rule's name, whether it allows, and one line
// on it. Every rule has a row in the delete check's test table (TestShellRulesEachHaveARow).
var delRules = []struct {
	Name  string
	Allow bool
	Line  string
}{
	{delRuleUnnamed, false, "a recursive or bulk delete that does not name what it deletes (a glob, a variable, " +
		"., .., /, ~, a drive's root; xargs or a pipeline feeding it; git clean with no path), and always git clean " +
		"-x or -X, git stash --all and rm --no-preserve-root"},
	{delRuleProtected, false, "a delete or move of a guarded path, or of a folder holding one, judged as an edit of " +
		"that path"},
	{delRuleUnsplit, false, "a command line the check cannot split (a quote or bracket left open, nesting past " +
		"three levels) that holds a delete word"},
	{delRuleNamed, true, "a shell call whose deletes and moves all name what they take, none of it guarded"},
	{delRuleNone, true, "a shell call with no delete word"},
}

// delJudge is what the delete check asks of the guard's rules (plan-5 5.3.2 notes 1 and 8): the one function that
// answers "may this agent change this path now", and the projects a shell call's deletes are judged in.
type delJudge interface {
	// delMay judges a delete or move of the path p, read against the folder dir as a file tool's path is read
	// against the payload's cwd, as an edit of that path (5.3.2 note 1). For deletes, the floor's .git is the
	// project's own top one only (5.3.2 note 2): a .git deeper down is not guarded here.
	delMay(dir, p string) Decision
	// delRoots gives the folders whose guarded paths a delete can take: the projects a shell call is judged in (the
	// session's own, from a worktree its verified main too, and the other linked projects on this machine's record,
	// 5.3.2 rule 8), and Bonsai's home.
	delRoots() []delRoot
}

// delRoot is one folder a delete is judged in.
type delRoot struct {
	Dir   string   // its top folder, absolute
	Globs []string // its guarded paths as folder-relative globs with forward slashes: the lists, the packs' declared
	//                paths, the floor's top paths; a glob with no fixed start (**/x) is judged against the
	//                operand itself only, never against what a deleted folder holds (5.3.3 note 4)
	Whole bool //     everything in it is guarded (Bonsai's home): a delete of a folder holding it is refused
	Own   bool //     the session's own project: a refusal's target is kept relative to it
}

// delOp is one operand of a delete or a move: one word, or one word its brace list writes out.
type delOp struct {
	text    string
	expand  bool    // a variable or substitution
	glob    bool    // a glob character
	literal bool    // given as -LiteralPath: its glob characters are literal
	git     bool    // a git pathspec: a leading : is pathspec magic
	pipeVar bool    // PowerShell's $_ or $PSItem: an item a pipeline or loop gives
	item    bool    // find's {}: judged as find's starting points are, not here
	word    *shWord // the word it came from, when it holds an expansion that may stand for literal text
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
}

// shellDecide judges one shell call by the delete check: tool is Bash, Monitor or PowerShell (any other name is read
// as bash), command its command line, cwd the payload's folder. No input makes it panic (FuzzShellDecide).
func shellDecide(tool, command, cwd string, j delJudge) Decision {
	d := shBash
	if tool == "PowerShell" {
		d = shPower
	}
	r := &delRun{tool: tool, j: j, roots: j.delRoots()}
	for _, root := range r.roots {
		if root.Own {
			r.own = root.Dir
			break
		}
	}
	switch {
	case cwd != "" && filepath.IsAbs(cwd):
		r.base = filepath.Clean(cwd)
	case r.own != "":
		r.base = r.own
	}
	start := delDirs{}
	if r.base != "" {
		start.known = []string{r.base}
	} else {
		start.unknown = true
	}
	r.text(command, d, 0, start, false)
	if r.deny != nil {
		return *r.deny
	}
	if r.found {
		return allow(delRuleNamed, r.first)
	}
	return allow(delRuleNone, "")
}

// delRun is one judgment of a command line.
type delRun struct {
	tool     string
	j        delJudge
	roots    []delRoot
	own      string    // the session's own project's top folder, "" for none
	base     string    // the payload's folder: where an operand is read when its folder cannot be told
	found    bool      // a delete was judged
	first    string    // the first operand judged, project-relative, for an allow's target
	firstSet bool      // first is set (it may be "": an operand outside the project)
	deny     *Decision // the first refusal
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
		if delHoldsWord(s) {
			r.refuse(Decision{Rule: delRuleUnsplit, Target: "",
				Why: fmt.Sprintf("%s refused: the command %s holds a delete word, and the check cannot split it (%s), "+
					"so it cannot tell what it deletes (rule %s)", r.tool, quote(s, 120), why, delRuleUnsplit),
				Next: "close every quote and bracket, nest at most three levels deep, and name what the delete removes"})
		}
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
	words []shWord
	shown string // the simple command as written, for an answer
	d     shDialect
	dirs  delDirs
	depth int      // the depth of the list it stands in
	loop  bool     // inside a loop's body
	feed  string   // set when a runner feeds the program its operands from input ("xargs"), which the check cannot see
	find  bool     // run by find -exec: its own operands are a bulk delete's
	piped bool     // fed by a pipe
	stdin []string // text fed to it: here-documents, here-strings, a literal piped in
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

// delPSAssign drops a PowerShell assignment's left side ($x = ..., $x += ...), so the program after it is read.
func delPSAssign(words []shWord) []shWord {
	if len(words) == 0 || words[0].quoted || !strings.HasPrefix(words[0].text, "$") {
		return words
	}
	t := words[0].text
	if k := strings.IndexByte(t, '='); k > 0 {
		rest := strings.TrimLeft(t[k+1:], "=")
		if rest == "" {
			return words[1:]
		}
		w := words[0]
		w.text, w.raw, w.expand = rest, rest, strings.Contains(rest, "$")
		return append([]shWord{w}, words[1:]...)
	}
	if len(words) >= 2 {
		switch words[1].text {
		case "=", "+=", "-=", "*=", "/=", "%=", "??=":
			return words[2:]
		}
	}
	return words
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
		r.refuse(Decision{Rule: delRuleUnnamed, Why: fmt.Sprintf("%s refused: %s does not name what it deletes: %s "+
			"(rule %s)", r.tool, quote(x.shown, 120), a.unnamed, delRuleUnnamed), Next: a.next})
		return
	}
	verb := "deletes"
	if a.move {
		verb = "moves"
	}
	if x.feed != "" {
		r.refuse(Decision{Rule: delRuleUnnamed, Why: fmt.Sprintf("%s refused: %s does not name what it %s: its "+
			"operands come from %s, so the check cannot see them (rule %s)", r.tool, quote(x.shown, 120), verb, x.feed,
			delRuleUnnamed), Next: "name what to " + delVerb(a.move) + " in the command itself, not through " + x.feed})
		return
	}
	a.ops = delLiterals(a.ops, x)
	a.dests = delLiterals(a.dests, x)
	bulk := a.bulk != "" || x.find
	for _, op := range a.ops {
		if op.item {
			continue
		}
		why, next := "", ""
		switch {
		case op.pipeVar || x.loop && op.expand:
			why = fmt.Sprintf("%s takes %s from a pipeline or a loop, one item after another, so the check "+
				"cannot see what it %s", a.word, quote(op.text, 60), verb)
			next = "name the files in the command, or " + delVerb(a.move) + " their folder by its name"
		case (a.recursive || bulk) && !a.move:
			if p := delPlain(op); p != "" {
				why = fmt.Sprintf("its operand %s %s", quote(op.text, 60), p)
				next = delNextFor(p)
			} else if x.dirs.unknown && !delIsAbs(op.text) {
				why = fmt.Sprintf("its operand %s is read in a folder the check cannot tell (after a cd it cannot "+
					"follow)", quote(op.text, 60))
				next = "cd to a plain folder name first, or give the operand as an absolute path"
			}
		case op.glob && !op.literal && !a.move:
			why = fmt.Sprintf("its operand %s holds a glob, which deletes several files it does not name", quote(op.text, 60))
			next = "name the files, or delete the folder by its name"
		}
		if why != "" {
			r.refuse(Decision{Rule: delRuleUnnamed, Target: r.target(x.dirs, op),
				Why:  fmt.Sprintf("%s refused: %s does not name what it %s: %s (rule %s)", r.tool, quote(x.shown, 120), verb, why, delRuleUnnamed),
				Next: next})
			return
		}
	}
	for _, op := range a.ops {
		if op.text == "" || op.expand || op.item {
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
	for _, dir := range r.dirsFor(x.dirs, text) {
		if !r.firstSet {
			r.first, r.firstSet = delRel(r.own, dir, text), true
		}
		if !glob {
			if dec := r.j.delMay(dir, text); !dec.Allow {
				r.protected(x, dir, text, "", dec, verb)
				return
			}
		}
		if !tree && !glob {
			continue
		}
		for _, root := range r.roots {
			rels, whole, err := delPlace(root.Dir, dir, text)
			if err != nil {
				r.refuse(Decision{Rule: RuleBadPath, Target: delRel(r.own, dir, text),
					Why: fmt.Sprintf("%s refused: %s %s %s, which Windows opens but cannot name (%s), so the check "+
						"cannot tell what it %s (rule %s)", r.tool, quote(x.shown, 120), verb, quote(text, 60),
						oneLine(err.Error()), verb, RuleBadPath),
					Next: "give the path in its plain form inside the project (a drive letter and folders), or tell the person"})
				return
			}
			if root.Whole && tree && (whole || len(rels) > 0) {
				dec := r.j.delMay(root.Dir, ".")
				if dec.Allow {
					dec = deny("", "everything in "+quote(filepath.ToSlash(root.Dir), 80)+" is Bonsai's own", "leave it as it is")
				}
				r.protected(x, dir, text, filepath.ToSlash(root.Dir), dec, verb)
				return
			}
			places := rels
			if whole {
				places = append([][]string{{}}, rels...)
			}
			fold := caseInsensitive(root.Dir)
			lastLiteral := !strings.ContainsAny(delBase(text), "*?[")
			for _, g := range root.Globs {
				fixed := delFixed(g)
				if !fixed && (!glob || !lastLiteral) {
					// Judged against the operand itself only (delMay, above): never against what a folder holds,
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
					if dec := r.j.delMay(root.Dir, shownRep); !dec.Allow {
						r.protected(x, dir, text, shownRep, dec, verb)
						return
					}
				}
			}
		}
	}
}

// protected refuses a delete of a guarded path: the operand, and the guarded path under it when that decided.
func (r *delRun) protected(x delExec, dir, text, under string, dec Decision, verb string) {
	what := quote(text, 60)
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
	r.refuse(Decision{Rule: delRuleProtected, Target: delRel(r.own, dir, text),
		Why: fmt.Sprintf("%s refused: %s %s %s, a guarded path; the guard's answer for changing it: %s (rule %s)",
			r.tool, quote(x.shown, 120), verb, what, why, delRuleProtected),
		Next: next})
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
