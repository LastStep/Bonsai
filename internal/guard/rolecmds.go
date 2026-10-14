package guard

// The roles' rules on an agent's shell line (design/plan-5.md, step 5.5, "Role rules and stall stops, written once";
// 5.5.11 notes 3, 3a and 4), read by the shell judge's own word splitting (shell.go): nothing is expanded or run.
//
//   - A command starting Claude Code (ClaudeStartWords: claude, claude-here, by any path and with .exe, .cmd, .bat or
//     .ps1; npx, pnpx or bunx of @anthropic-ai/claude-code, npm exec, pnpm dlx and yarn dlx too, any version; node
//     running Claude Code's cli.js), from every role but the session's: refused (call-refused), since a top-level
//     session started from an agent's shell carries no agent_id and would read as the person's own.
//   - bonsai message (commands.go hands each call here, messageCall): one that writes a message from a caller with no
//     address, or with a --from other than its own address, refused; a --kind brief to a role the writer may not call,
//     refused (call-refused). --show and bonsai messages read, and are free.
//   - bonsai ladder --task <id>: noted, once the call is allowed, so the agent's next call reads the climb's result
//     (budget.go's readClimbs).
//   - Shell writes the guard can read (writeTargets): a redirect (>, >>, >|, &>, &>>, <> and their numbered forms; in
//     PowerShell *> too), tee, cp, ln and sed -i or --in-place onto a path, PowerShell's Set-Content, Add-Content,
//     Out-File, Tee-Object, Copy-Item and New-Item, and cmd's copy and xcopy; with a role that changes nothing, rm and mv
//     too (the delete check's words, shell.go), git's words that change the checkout or its history (GitWriteWords),
//     and bonsai init, update, unlink and check --write (commands.go). Each written path is judged in every folder the
//     command may run in (mayWrite): from every agent (all but the person's session), one in Bonsai's own files (the
//     floor's bonsai-files entries in any linked project, .bonsai/local/ in a worktree's main checkout among them, the
//     lock at a checkout's top, Bonsai's home) is refused (bonsai-files); from a role that changes nothing, one inside
//     any linked project is refused (role-writes-nothing). Writes outside every linked project (a scratch build, a
//     Windows checkout under the user's folder) stay allowed, as do Bonsai's own message and ladder.
//
// Known limits for the roles that change nothing, said again here as shell.go says them for the delete check: what the
// splitter cannot read is not judged: an interpreter's write (python -c, node -e), a script run by name, a path built
// in a variable or a substitution whose output is not a literal, a program's own writes (make, a test run, git's hooks),
// a link made earlier in the line and written through, and programs outside the tables (install, rsync's copies,
// dd's of=, unzip, tar -x, patch); and an MCP tool whose server writes. They are caught after: by the lead's read of
// the diff, and by the ladder's git.dirty and rung 0 at the next climb.

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// shellRole is the roles' part of one shell call's judgment: nil where no role rule applies (the roles do not apply,
// or the caller is the person's session, which keeps today's guard but for bonsai message's --from).
type shellRole struct {
	rr *roleRun
	w  *who
}

// agent reports a caller other than the person's session.
func (s *shellRole) agent() bool { return s != nil && !s.w.h.SessRole }

// changesNothing reports a caller whose role changes nothing (writes: false, every undeclared type).
func (s *shellRole) changesNothing() bool { return s.agent() && !s.w.h.Writes }

// ClaudeStartWords are the program names that start Claude Code from a shell (base names, case-folded, with .exe,
// .cmd, .bat or .ps1 dropped), and ClaudePackage its npm package: their one home, on the reference page.
var (
	ClaudeStartWords = []string{"claude", "claude-here"}
	ClaudePackage    = "@anthropic-ai/claude-code"
)

// claudeName drops a program name's Windows endings, for ClaudeStartWords.
func claudeName(name string) string {
	for _, ext := range []string{".exe", ".cmd", ".bat", ".ps1"} {
		name = strings.TrimSuffix(name, ext)
	}
	return name
}

// claudeStart says how a command starts Claude Code, "" when it does not (this file's comment). name is its program's
// name as delName gives it, words its words from the program on.
func claudeStart(name string, words []shWord) string {
	n := claudeName(name)
	if listHas(ClaudeStartWords, n) {
		return "it starts Claude Code (" + n + ")"
	}
	args := words[1:]
	pkg := func(ws []shWord) bool {
		for _, w := range ws {
			t := strings.ToLower(w.text)
			if strings.HasPrefix(t, "-") {
				continue
			}
			return t == ClaudePackage || strings.HasPrefix(t, ClaudePackage+"@")
		}
		return false
	}
	anyPkg := func(ws []shWord) bool {
		for _, w := range ws {
			t := strings.ToLower(w.text)
			if t == ClaudePackage || strings.HasPrefix(t, ClaudePackage+"@") || strings.HasPrefix(t, "--package="+ClaudePackage) {
				return true
			}
		}
		return false
	}
	if listHas([]string{"npx", "pnpx", "bunx"}, n) && (pkg(args) || anyPkg(args)) {
		return "it starts Claude Code through " + n + " (" + ClaudePackage + ")"
	}
	if listHas([]string{"npm", "pnpm", "yarn", "bun"}, n) && len(args) > 0 &&
		listHas([]string{"exec", "x", "dlx"}, strings.ToLower(args[0].text)) && anyPkg(args[1:]) {
		return "it starts Claude Code through " + n + " " + strings.ToLower(args[0].text) + " (" + ClaudePackage + ")"
	}
	if listHas([]string{"node", "nodejs", "bun", "deno"}, n) {
		for _, w := range args {
			t := strings.ToLower(strings.ReplaceAll(w.text, `\`, "/"))
			if strings.HasPrefix(t, "-") {
				continue
			}
			if strings.HasSuffix(t, "/cli.js") && (strings.Contains(t, "claude-code") || strings.Contains(t, ClaudePackage)) {
				return "it runs Claude Code's cli.js with " + n
			}
			break
		}
	}
	return ""
}

// GitWriteWords are git's subcommands that change the checkout or its history, refused to a role that changes
// nothing; each with the forms that only read, which stay allowed: their one home, on the reference page.
var GitWriteWords = []struct {
	Sub   string
	Reads string // the first words after the subcommand that only read ("" for none): a listing form
}{
	{"add", ""}, {"am", ""}, {"apply", "--check --stat --numstat --summary"}, {"checkout", ""}, {"cherry-pick", ""},
	{"clean", "-n --dry-run"}, {"commit", ""}, {"filter-branch", ""}, {"merge", ""}, {"mv", "-n --dry-run"},
	{"notes", "list show"}, {"pull", ""}, {"push", "--dry-run -n"}, {"rebase", ""}, {"replace", "-l --list"},
	{"reset", ""}, {"restore", ""}, {"revert", ""}, {"rm", "-n --dry-run"}, {"stash", "list show"}, {"switch", ""},
	{"update-index", ""}, {"update-ref", ""}, {"worktree", "list"}, {"branch", "-l --list -a --all -r --remotes -v -vv --show-current --contains --merged --no-merged"},
	{"tag", "-l --list -v --verify --contains"},
}

// gitWrite says which of git's write words a git call's subcommand is, "" for one that only reads (this file's
// comment). sub is its subcommand, rest the words after it.
func gitWrite(sub string, rest []shWord) string {
	for _, g := range GitWriteWords {
		if g.Sub != sub {
			continue
		}
		reads := strings.Fields(g.Reads)
		if len(rest) == 0 && (sub == "branch" || sub == "tag" || sub == "stash" || sub == "worktree" || sub == "notes") {
			if sub == "stash" || sub == "worktree" || sub == "notes" {
				return "git " + sub // a bare stash stashes; a bare worktree or notes call is refused by git: judged as written
			}
			return "" // a bare branch or tag lists
		}
		for _, w := range rest {
			if listHas(reads, w.text) {
				return ""
			}
		}
		return "git " + sub
	}
	return ""
}

// writeOpts are the GNU options of the write words: tee, cp, ln, sed.
var (
	teeOpts = delOpt{short: "aip", long: []string{"append", "ignore-interrupts"}, longOpt: []string{"output-error"}, permute: true}
	cpOpts  = delOpt{short: "abdfHilLnPpRrsTuvx", shortVal: "St", long: []string{"archive", "attributes-only", "force",
		"interactive", "link", "dereference", "no-clobber", "no-dereference", "parents", "recursive", "remove-destination",
		"strip-trailing-slashes", "symbolic-link", "no-target-directory", "update", "verbose", "one-file-system", "debug",
		"copy-contents", "keep-directory-symlink"}, longVal: []string{"suffix", "target-directory", "no-preserve"},
		longOpt: []string{"backup", "preserve", "reflink", "sparse", "context"}, permute: true}
	lnOpts = delOpt{short: "bdFfiLnPrsTv", shortVal: "St", long: []string{"directory", "force", "interactive", "logical",
		"no-dereference", "physical", "relative", "symbolic", "no-target-directory", "verbose"},
		longVal: []string{"suffix", "target-directory"}, longOpt: []string{"backup"}, permute: true}
	sedOpts = delOpt{short: "nrEsuz", shortVal: "efl", shortOpt: "i", long: []string{"quiet", "silent", "debug",
		"posix", "regexp-extended", "separate", "sandbox", "unbuffered", "null-data", "follow-symlinks"},
		longVal: []string{"expression", "file", "line-length"}, longOpt: []string{"in-place"}, permute: true}
)

// PowerShell's write cmdlets, their parameters as delPSBind reads them.
var (
	psWriteContent = delPSCmdlet{params: []string{"path", "literalpath", "value", "passthru", "filter", "include", "exclude",
		"force", "credential", "whatif", "confirm", "nonewline", "encoding", "stream", "asbytestream"},
		switches: []string{"passthru", "force", "whatif", "confirm", "nonewline", "asbytestream", "verbose", "debug"},
		aliases:  map[string]string{"pspath": "literalpath", "lp": "literalpath", "wi": "whatif", "cf": "confirm"}, positional: []string{"path", "value"}}
	psOutFile = delPSCmdlet{params: []string{"filepath", "literalpath", "encoding", "append", "force", "noclobber", "width",
		"nonewline", "inputobject", "whatif", "confirm"}, switches: []string{"append", "force", "noclobber", "nonewline", "whatif", "confirm", "verbose", "debug"},
		aliases: map[string]string{"path": "filepath", "pspath": "literalpath", "lp": "literalpath", "wi": "whatif", "cf": "confirm"}, positional: []string{"filepath", "encoding"}}
	psTee = delPSCmdlet{params: []string{"filepath", "literalpath", "append", "variable", "inputobject", "encoding"},
		switches: []string{"append", "verbose", "debug"}, aliases: map[string]string{"path": "filepath", "pspath": "literalpath", "lp": "literalpath"},
		positional: []string{"filepath"}}
	psCopy = delPSCmdlet{params: []string{"path", "literalpath", "destination", "container", "force", "filter", "include",
		"exclude", "recurse", "passthru", "credential", "whatif", "confirm", "fromsession", "tosession"},
		switches: []string{"container", "force", "recurse", "passthru", "whatif", "confirm", "verbose", "debug"},
		aliases:  map[string]string{"pspath": "literalpath", "lp": "literalpath", "wi": "whatif", "cf": "confirm"}, positional: []string{"path", "destination"}}
	psNewItem = delPSCmdlet{params: []string{"path", "name", "itemtype", "value", "force", "credential", "whatif", "confirm"},
		switches: []string{"force", "whatif", "confirm", "verbose", "debug"}, aliases: map[string]string{"type": "itemtype", "wi": "whatif", "cf": "confirm"},
		positional: []string{"path"}}
)

// writeTarget is one path a command writes, and whether it may be a folder its sources land in (cp, ln, Copy-Item:
// dest/<a source's name> is judged too).
type writeTarget struct {
	op      delOp
	srcs    []delOp
	destDir bool
	word    string // the write word, as a refusal names it
}

// writeTargets gives the paths a write word writes (this file's comment), for a command whose program name and
// native-ness delName gives; none for a command that is no write word.
func writeTargets(name string, native bool, x delExec) []writeTarget {
	words := x.words[1:]
	d := x.d
	ops := func(ws []shWord) []delOp { return delOps(ws, d, false, false) }
	one := func(word string, ws []shWord) []writeTarget {
		var out []writeTarget
		for _, op := range ops(ws) {
			out = append(out, writeTarget{op: op, word: word})
		}
		return out
	}
	copyLike := func(word string, o delOpt) []writeTarget {
		g := delParse(words, o)
		if t := g.val("t", "target-directory"); len(t) > 0 {
			var out []writeTarget
			for _, op := range ops(t) {
				out = append(out, writeTarget{op: op, srcs: ops(g.ops), destDir: true, word: word})
			}
			return out
		}
		if len(g.ops) < 2 {
			if word == "ln" && len(g.ops) == 1 {
				// ln with one operand makes the link in the current folder, under the target's name.
				src := ops(g.ops)
				if len(src) == 1 {
					src[0].text = path.Base(strings.ReplaceAll(src[0].text, `\`, "/"))
					return []writeTarget{{op: src[0], word: word}}
				}
			}
			return nil
		}
		dest := ops(g.ops[len(g.ops)-1:])
		var out []writeTarget
		for _, op := range dest {
			out = append(out, writeTarget{op: op, srcs: ops(g.ops[:len(g.ops)-1]), destDir: !g.has("T", "no-target-directory"), word: word})
		}
		return out
	}
	switch {
	case d == shPower && !native:
		switch name {
		case "set-content", "sc", "add-content", "ac":
			b := delPSBind(words, psWriteContent)
			return one(cmdletName(name), append(append([]shWord(nil), b.vals["path"]...), b.vals["literalpath"]...))
		case "out-file":
			b := delPSBind(words, psOutFile)
			return one("Out-File", append(append([]shWord(nil), b.vals["filepath"]...), b.vals["literalpath"]...))
		case "tee-object", "tee":
			b := delPSBind(words, psTee)
			return one("Tee-Object", append(append([]shWord(nil), b.vals["filepath"]...), b.vals["literalpath"]...))
		case "copy-item", "copy", "cp", "cpi":
			b := delPSBind(words, psCopy)
			var out []writeTarget
			srcs := delPSOps(b)
			for _, op := range ops(b.vals["destination"]) {
				out = append(out, writeTarget{op: op, srcs: srcs, destDir: true, word: "Copy-Item"})
			}
			return out
		case "new-item", "ni":
			b := delPSBind(words, psNewItem)
			var out []writeTarget
			names := ops(b.vals["name"])
			for _, op := range ops(b.vals["path"]) {
				if len(names) == 0 {
					out = append(out, writeTarget{op: op, word: "New-Item"})
				}
				for _, n := range names {
					in := op
					in.text = strings.TrimRight(op.text, `/\`) + "/" + n.text
					in.expand = op.expand || n.expand
					out = append(out, writeTarget{op: in, word: "New-Item"})
				}
			}
			if len(b.vals["path"]) == 0 {
				out = append(out, one("New-Item", b.vals["name"])...)
			}
			return out
		}
		return nil
	case d == shCmd && !native:
		switch name {
		case "copy", "xcopy":
			var ws []shWord
			for _, w := range words {
				if !strings.HasPrefix(w.text, "/") {
					ws = append(ws, w)
				}
			}
			if len(ws) < 2 {
				return nil
			}
			return []writeTarget{{op: ops(ws[len(ws)-1:])[0], srcs: ops(ws[:len(ws)-1]), destDir: true, word: name}}
		}
		return nil
	}
	switch name {
	case "tee":
		return one("tee", delParse(words, teeOpts).ops)
	case "cp":
		return copyLike("cp", cpOpts)
	case "ln":
		return copyLike("ln", lnOpts)
	case "sed":
		g := delParse(words, sedOpts)
		if !g.has("i", "in-place") {
			return nil
		}
		files := g.ops
		if !g.has("e", "expression", "f", "file") && len(files) > 0 {
			files = files[1:] // the first operand is the script
		}
		return one("sed -i", files)
	}
	return nil
}

// cmdletName names a content cmdlet by its full name, for a refusal.
func cmdletName(name string) string {
	switch name {
	case "set-content", "sc":
		return "Set-Content"
	case "add-content", "ac":
		return "Add-Content"
	}
	return name
}

// mayWrite judges one path a shell call writes, read against the folder dir (this file's comment): bonsai-files from
// every agent, then, with role set, role-writes-nothing from a role that changes nothing; else allowed.
func (s *shellRole) mayWrite(dir, p, word string, role bool) Decision {
	j := s.rr.j
	forms, err := pathForms(dir, dir, p)
	if err != nil {
		if role && s.changesNothing() {
			return deny(RuleRoleWritesNothing, fmt.Sprintf("%s refused: %s writes %s, which the guard cannot place (%s), and this "+
				"role changes nothing", s.rr.in.Tool, word, quote(j.shown(p), 120), oneLine(err.Error())),
				"put what you would write in your report instead")
		}
		return allow(RuleShell, "")
	}
	for _, abs := range forms {
		if devicePath(abs) || adminShare(abs) != "" {
			continue
		}
		if home := j.inHome(abs); home != "" {
			return deny(RuleBonsaiFiles, fmt.Sprintf("%s refused: %s writes into Bonsai's home (%s), which only Bonsai's own "+
				"commands write", s.rr.in.Tool, word, quote(j.shown(home), 120)), "leave Bonsai's home as it is")
		}
		chain := j.chain(abs)
		for _, proj := range chain {
			rel, ok := relTo(proj.root, abs)
			if !ok {
				continue
			}
			if f := floorOf(rel, caseInsensitive(proj.root), false); f != nil && (f.Class == RuleBonsaiFiles || f.Glob == ".bonsai/lock.json") {
				d := deny(RuleBonsaiFiles, fmt.Sprintf("%s refused: %s writes %s, one of Bonsai's own files (%s): no agent writes "+
					"there by a shell line, whatever its task grants, so no count, brief, report or log line is reset or planted",
					s.rr.in.Tool, word, quote(rel, 120), f.Why), "write through Bonsai's own command (bonsai message, bonsai "+
					"ladder), or put it in your report")
				d.Target = target(j.session, []string{abs})
				return d
			}
		}
		if role && s.changesNothing() && len(chain) > 0 {
			d := deny(RuleRoleWritesNothing, fmt.Sprintf("%s refused: %s writes %s, inside a linked project, and %s changes nothing "+
				"(its role's writes: false)", s.rr.in.Tool, word, quote(j.shown(abs), 120), s.w.shown()),
				"put what you would change in your report; write scratch files outside the project")
			d.Target = target(j.session, []string{abs})
			return d
		}
	}
	return allow(RuleShell, "")
}

// writes judges the paths a write word writes, in each folder the command may run in: every path for Bonsai's own
// files first, then every path for a role that changes nothing.
func (r *delRun) writes(rs *shellRole, x delExec, ts []writeTarget) {
	var all []writeTarget
	for _, t := range ts {
		all = append(all, t)
		if !t.destDir {
			continue
		}
		for _, src := range t.srcs {
			base := delBase(src.text)
			if base == "" || src.expand || src.glob {
				continue
			}
			in := t
			in.op.text = strings.TrimRight(t.op.text, `/\`) + "/" + base
			all = append(all, in)
		}
	}
	for _, role := range []bool{false, true} {
		for _, t := range all {
			if r.writePath(rs, x, t.op, t.word, role); r.deny != nil {
				return
			}
		}
	}
}

// writePath judges one written path in each folder the command may run in; a path held in a variable or a
// substitution the check cannot read is not judged (a known limit, this file's comment).
func (r *delRun) writePath(rs *shellRole, x delExec, op delOp, word string, role bool) {
	ops := delLiterals([]delOp{op}, x)
	for _, op := range ops {
		if op.text == "" || op.expand || op.item {
			continue
		}
		if x.d == shBash {
			delTildeOp(&op)
		}
		text := op.text
		if x.d != shBash && filepath.Separator == '/' {
			text = strings.ReplaceAll(text, `\`, "/")
		}
		if x.d == shPower {
			text = delProvider(text)
		}
		dirs := r.dirsFor(x.dirs, text)
		if op.away && r.base != "" {
			dirs = appendNew(dirs, r.base)
		}
		for _, dir := range dirs {
			if d := rs.mayWrite(dir, text, word, role); !d.Allow {
				r.refuse(d)
				return
			}
		}
	}
}

// redirects judges a command's redirections that write (shRedir), in the folders it runs in.
func (r *delRun) redirects(rs *shellRole, c *shCommand, d shDialect, dirs delDirs, depth int) {
	if len(c.redirs) == 0 {
		return
	}
	x := delExec{d: d, dirs: dirs, depth: depth}
	var ts []writeTarget
	for _, rd := range c.redirs {
		for _, op := range delOps([]shWord{rd.target}, d, true, false) {
			ts = append(ts, writeTarget{op: op, word: "a redirect (" + rd.op + ")"})
		}
	}
	r.writes(rs, x, ts)
}

// roleExec judges one command's program for the roles' rules (this file's comment), before the delete check's own
// reading of it: a start of Claude Code, a write word, and, for a role that changes nothing, git's write words.
func (r *delRun) roleExec(rs *shellRole, name string, native bool, x delExec) {
	if rs == nil {
		return
	}
	if rs.agent() {
		if how := claudeStart(name, x.words); how != "" {
			r.refuse(Decision{Rule: RuleCallRefused, Why: fmt.Sprintf("%s refused: %s, and %s may not start a session from its "+
				"shell: a session started so carries no agent id and would read as the person's own (who may call whom)",
				r.tool, how, rs.w.shown()), Next: "call the role you need through the Agent tool, or hand the work back to your caller in your report"})
			return
		}
	}
	if !rs.agent() {
		return
	}
	if ts := writeTargets(name, native, x); len(ts) > 0 {
		r.writes(rs, x, ts)
		if r.deny != nil {
			return
		}
	}
	if name == "git" && rs.changesNothing() {
		sub, rest := gitSub(x.words[1:])
		if w := gitWrite(sub, rest); w != "" {
			r.refuse(Decision{Rule: RuleRoleWritesNothing, Why: fmt.Sprintf("%s refused: %s changes the checkout or its history, "+
				"and %s changes nothing (its role's writes: false)", r.tool, w, rs.w.shown()),
				Next: "put what you would change in your report; reading git (status, diff, log, show) stays allowed"})
		}
	}
}

// gitSub gives a git call's subcommand and the words after it, past git's own options (-C <dir>, -c <name=value>,
// --git-dir=..., and the rest).
func gitSub(words []shWord) (string, []shWord) {
	for i := 0; i < len(words); i++ {
		t := words[i].text
		switch {
		case listHas([]string{"-C", "-c", "--git-dir", "--work-tree", "--namespace", "--super-prefix", "--config-env"}, t):
			i++ // its value is the next word
		case strings.HasPrefix(t, "-"):
		default:
			return strings.ToLower(t), words[i+1:]
		}
	}
	return "", nil
}

// messageCall judges a bonsai message call's words for the roles' rules (this file's comment); argv its words from
// the program on. Allowed is RuleBonsaiCommand.
func (s *shellRole) messageCall(argv []string) Decision {
	if s == nil {
		return allow(RuleBonsaiCommand, "")
	}
	flags := map[string]string{}
	show := false
	for i := 1; i < len(argv); i++ {
		a := argv[i]
		name, val, hasVal := strings.Cut(a, "=")
		switch name {
		case "--show":
			show = true
		case "--from", "--to", "--kind":
			if !hasVal && i+1 < len(argv) {
				val = argv[i+1]
				i++
			}
			if _, seen := flags[name]; seen {
				flags[name] = "\x00twice"
			} else {
				flags[name] = val
			}
		}
	}
	if show && flags["--from"] == "" && flags["--kind"] == "" {
		return allow(RuleBonsaiCommand, "") // a read
	}
	addr := s.rr.address(s.w)
	refuse := func(why string) Decision {
		return deny(RuleCallRefused, "bonsai message refused: "+why, "write your message from your own address ("+
			orNone(addr)+"); a brief only to a role you may call; or put it in your report")
	}
	from := flags["--from"]
	switch {
	case addr == "":
		return refuse(s.w.shown() + " has no address (no brief is tied to it), so it writes no message: its last reply is its report")
	case strings.Contains(from, "\x00"):
		return refuse("it gives --from twice")
	case from != addr:
		return refuse("its --from " + quote(from, 60) + " is not the writer's own address " + quote(addr, 60))
	}
	if kind := flags["--kind"]; kind == "brief" {
		rs, err := s.rr.roles()
		if err != nil {
			return rolesFault(err)
		}
		to := roleOf(flags["--to"])
		c := s.rr.resolveCallee(rs, to)
		if r := rs.byType(to); r == nil {
			c = callee{shown: "the role " + quote(to, 40) + ", which no pack declares"}
		}
		if why := mayCall(s.w, c); why != "" {
			return refuse("a brief to " + c.shown + ": " + why)
		}
	}
	return allow(RuleBonsaiCommand, "")
}

// orNone writes an address for a refusal: itself, or "none".
func orNone(addr string) string {
	if addr == "" {
		return "none"
	}
	return addr
}

// reportCall reports a shell call that is an agent's report past its budget in its exact form: one command and nothing
// else (no ;, &&, ||, pipe, &, group, command or process substitution, redirection), bonsai message --kind report
// --from <its own address>, every word literal (no variable, no substitution, no glob), its body by --body in single
// quotes or --body - fed by one here-document whose delimiter is quoted (<<'EOF'), nothing expanded or run. PowerShell
// takes the single-quoted body alone.
func reportCall(tool, command, addr string) bool {
	d := shBash
	if tool == "PowerShell" {
		d = shPower
	}
	l, why := shSplit(command, d, 0)
	if why != "" || l == nil || len(l.cmds) != 1 {
		return false
	}
	c := l.cmds[0]
	// One command: a line end (or ;) may close it, nothing may follow it.
	if c.group != nil || len(c.nested) > 0 || len(c.redirs) > 0 || (c.op != "" && c.op != ";") || c.piped || len(c.words) < 2 {
		return false
	}
	if len(c.feeds) > 1 || (len(c.feeds) == 1 && (!c.feeds[0].doc || !c.feeds[0].quoted || len(c.feeds[0].nested) > 0)) {
		return false
	}
	if !delBonsai(c.words[0].text) || c.words[0].expand {
		return false
	}
	kind, from, body, bodyDash := "", "", false, false
	for i := 1; i < len(c.words); i++ {
		w := c.words[i]
		if w.expand || w.glob || w.splat || w.alts != nil {
			return false
		}
		name, val, hasVal := strings.Cut(w.text, "=")
		next := func() (string, bool) {
			if hasVal {
				return val, true
			}
			if i+1 >= len(c.words) {
				return "", false
			}
			i++
			nw := c.words[i]
			if nw.expand || nw.splat || nw.alts != nil {
				return "", false
			}
			return nw.text, true
		}
		switch {
		case i == 1 && w.text != "message":
			return false
		case i == 1:
		case name == "--kind":
			v, ok := next()
			if !ok || kind != "" {
				return false
			}
			kind = v
		case name == "--from":
			v, ok := next()
			if !ok || from != "" {
				return false
			}
			from = v
		case name == "--body":
			if body {
				return false
			}
			body = true
			if hasVal {
				return false // --body=... : written as two words, its quoting seen
			}
			if i+1 >= len(c.words) {
				return false
			}
			i++
			bw := c.words[i]
			if bw.text == "-" && !bw.quoted {
				bodyDash = true
				continue
			}
			if !strings.HasPrefix(bw.raw, "'") || !strings.HasSuffix(bw.raw, "'") || len(bw.raw) < 2 || bw.expand {
				return false
			}
		case name == "--body-file", name == "--show":
			return false
		case strings.HasPrefix(name, "--") && !hasVal:
			// another flag: its value, when it takes one, is a literal word read on the next turn
		}
	}
	switch {
	case kind != "report" || from != addr || !body:
		return false
	case bodyDash != (len(c.feeds) == 1):
		return false
	}
	return true
}

// briefShow gives the id a shell call reads with bonsai message --show <id> as its one command, "" for any other call.
func briefShow(tool, command string) string {
	d := shBash
	if tool == "PowerShell" {
		d = shPower
	}
	l, why := shSplit(command, d, 0)
	if why != "" || l == nil || len(l.cmds) != 1 {
		return ""
	}
	c := l.cmds[0]
	if c.group != nil || len(c.nested) > 0 || len(c.redirs) > 0 || len(c.feeds) > 0 || (c.op != "" && c.op != ";") || len(c.words) < 3 {
		return ""
	}
	if !delBonsai(c.words[0].text) || c.words[1].text != "message" {
		return ""
	}
	id := ""
	for i := 2; i < len(c.words); i++ {
		w := c.words[i]
		if w.expand || w.glob {
			return ""
		}
		switch {
		case w.text == "--show" && i+1 < len(c.words):
			i++
			id = c.words[i].text
		case strings.HasPrefix(w.text, "--show="):
			id = strings.TrimPrefix(w.text, "--show=")
		case w.text == "--json":
		default:
			return ""
		}
	}
	if briefPattern().FindString(id) != id {
		return ""
	}
	return id
}
