package guard

// The roles' rules on an agent's shell line (design/plan-5.md, step 5.5, "Role rules and stall stops, written once";
// 5.5.11 notes 3, 3a and 4), read by the shell judge's own word splitting (shell.go): nothing is expanded or run.
//
//   - A command starting Claude Code (ClaudeStartWords: claude, claude-here, by any path and with .exe, .cmd, .bat or
//     .ps1; a program run by a path in Claude Code's native versions folder, ClaudeVersions; npx, pnpx or bunx of
//     @anthropic-ai/claude-code, npm exec, pnpm dlx and yarn dlx too, any version; node running Claude Code's cli.js),
//     from every role but the session's: refused (call-refused), since a top-level session started from an agent's
//     shell carries no agent_id and would read as the person's own.
//   - bonsai message (commands.go hands each call here, messageCall): one that writes a message from a caller with no
//     address, or with a --from other than its own address, refused; a --kind brief to a role the writer may not call,
//     refused (call-refused). --show and bonsai messages read, and are free. Past a budget, only the installed bonsai
//     at the place the hook lines run it from (installedBonsai; in PowerShell after its call operator, &) is the report
//     or a brief's read; a bare bonsai is not (it runs whatever the PATH finds first).
//   - bonsai ladder: noted, once the call is allowed, so the agent's next call reads the climb's result (budget.go's
//     readClimbs), for its --task <id>, or, with none the guard can read, for BONSAI_TASK as the ladder reads it (an
//     assignment of it on the line, export's, else the session's own).
//   - Shell writes the guard can read (writeTargets): a redirect (>, >>, >|, &>, &>>, <> and their numbered forms; in
//     PowerShell *> too), tee, cp, ln and sed -i or --in-place onto a path (and install, touch, truncate, mkfifo, mknod
//     and dd's of=), git's own output files (diff, log and show --output, archive -o, format-patch -o), PowerShell's
//     Set-Content, Add-Content, Clear-Content, Set-Item, Clear-Item, Out-File, Tee-Object, Copy-Item, New-Item,
//     Export-Csv, Start-Transcript and Invoke-WebRequest or Invoke-RestMethod -OutFile, and cmd's copy and xcopy; a
//     change of a path's permissions, owner, ACL or attributes (attrwords.go: chmod, chown, chgrp, chattr, setfacl,
//     setfattr, chcon, chflags, attrib, icacls, cacls, takeown, Set-ItemProperty, Clear-ItemProperty, Set-Acl, and a
//     member set on a Get-Item result), judged by what it reaches too (a folder holding Bonsai's files, a glob that may
//     name them); with a role that changes nothing, rm and mv too (the delete check's words, shell.go), git's words
//     that change the checkout or its history (GitWriteWords), and bonsai init, update, unlink and check --write
//     (commands.go). Each written path is judged in every folder the command may run in (mayWrite): from every agent
//     (all but the person's session), one in Bonsai's own files (the floor's bonsai-files entries in any linked
//     project, .bonsai/local/ in a worktree's main checkout among them, the lock at a checkout's top, Bonsai's home) is
//     refused (bonsai-files); from a role that changes nothing, one inside any linked project is refused
//     (role-writes-nothing). Writes outside every linked project (a scratch build, a Windows checkout under the user's
//     folder) stay allowed, as do Bonsai's own message and ladder.
//
// Known limits for the roles that change nothing, said again here as shell.go says them for the delete check: what the
// splitter cannot read is not judged: an interpreter's write (python -c, node -e), a script run by name, a path built
// in a variable or a substitution whose output is not a literal, a program's own writes (make, a test run, git's hooks,
// a git alias already in git's config files), a link made earlier in the line and written through, and the programs
// and words these tables do not hold, each a write the guard lets through: rsync, unzip, tar -x, patch, curl -o, wget
// -O, scp, a compiler's -o, a formatter's -w (gofmt -w, prettier --write), git's words outside GitWriteWords that
// write only its own store (gc, prune, repack, fetch, hash-object -w, commit-tree, mktag, write-tree, pack-refs,
// reflog expire), git -C <folder> before an output option (the file is read in the command's folder), and
// PowerShell cmdlets outside the list above (Rename-Item and Move-Item are the delete check's); and an MCP tool
// whose server writes. They are caught after: by the lead's read of the diff, and by the ladder's git.dirty and rung 0
// at the next climb. A BONSAI_TASK changed where the guard cannot read it (a script, PowerShell's $env:, a nested
// shell) is not followed: a climb so is noted for the session's own task, and counts once a result for it appears.

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// shellRole is the roles' part of one shell call's judgment: nil where no role rule applies (the roles do not apply,
// or the caller is the person's session, which keeps today's guard but for bonsai message's --from).
type shellRole struct {
	rr *roleRun
	w  *who
	// taskAssign is a literal BONSAI_TASK=<id> before the command being read (bash), the task a bonsai ladder with no
	// --task climbs for; "" for none.
	taskAssign string
	own        []string // Bonsai's own files a change of permissions may reach (attrwords.go's bonsaiOwn); nil until read
}

// noteTaskAssign keeps a line's literal assignment of BONSAI_TASK (bash), for a bonsai ladder with no --task
// (commands.go's roleCommand): among the assignments before a command's program, or as an argument of export; the
// line's last one is kept (an assignment alone changes the exported variable for the commands after it).
func (s *shellRole) noteTaskAssign(words []shWord) {
	if s == nil {
		return
	}
	set := func(w shWord) {
		if name, val, ok := strings.Cut(w.text, "="); ok && strings.EqualFold(name, workspace.TaskEnv) && !w.expand {
			s.taskAssign = val
		}
	}
	for i, w := range words {
		switch {
		case w.assign:
			set(w)
		case w.text == "export" || w.text == "declare" || w.text == "typeset":
			for _, a := range words[i+1:] {
				set(a)
			}
			return
		default:
			return
		}
	}
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
	// ClaudeVersions is the folder Claude Code's native installer keeps each version's binary in, named by its
	// version (~/.local/share/claude/versions/2.1.300, the file claude links to): a program run by a path holding it
	// starts Claude Code, whatever its name.
	ClaudeVersions = "/claude/versions/"
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
	if p := strings.ToLower(strings.ReplaceAll(words[0].text, `\`, "/")); strings.Contains(p, ClaudeVersions) {
		return "it starts Claude Code's native binary by its path (" + ClaudeVersions + "<version>)"
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
// nothing, each with the forms that only read, which stay allowed (gitWrite): their one home, on the reference page.
// Measured on git 2.x (15 Oct): git branch -v <name> and git tag --sort=<key> <name> make the ref, so a listing flag
// counts only where it puts the call in its listing mode. git symbolic-ref given two names sets the first, whatever
// else is there (gitWrite), so it only reads with one name and --short or -q.
var GitWriteWords = []struct {
	Sub   string
	Reads string // the words after the subcommand that make the call only read ("" for none): a listing or a dry run
	First bool   // Reads count only as the first word after the subcommand (its own subcommand: stash list); a word
	// of Reads that is no option counts there in every row (config get)
	Writes string // options that make the call write whatever Reads it holds ("" for none)
	Lists  bool   // with no word of Reads or Writes, it only reads when it names nothing: options alone list (branch -v)
}{
	{Sub: "add"}, {Sub: "stage"}, {Sub: "am"},
	{Sub: "apply", Reads: "--check --stat --numstat --summary", Writes: "--apply"},
	{Sub: "checkout"}, {Sub: "cherry-pick"},
	{Sub: "clean", Reads: "-n --dry-run"},
	{Sub: "commit"}, {Sub: "filter-branch"}, {Sub: "merge"},
	{Sub: "mv", Reads: "-n --dry-run"},
	{Sub: "notes", Reads: "list show", First: true},
	{Sub: "pull"},
	{Sub: "push", Reads: "--dry-run -n"},
	{Sub: "rebase"},
	{Sub: "replace", Reads: "-l --list", Writes: "-d --delete -f --force -e --edit --graft --convert-graft-file"},
	{Sub: "reset"}, {Sub: "restore"}, {Sub: "revert"},
	{Sub: "rm", Reads: "-n --dry-run"},
	{Sub: "stash", Reads: "list show", First: true},
	{Sub: "switch"}, {Sub: "update-index"}, {Sub: "update-ref"},
	{Sub: "worktree", Reads: "list", First: true},
	{Sub: "bisect", Reads: "log view visualize terms", First: true},
	{Sub: "clone"}, {Sub: "init"},
	{Sub: "submodule", Reads: "status summary", First: true},
	{Sub: "sparse-checkout", Reads: "list", First: true},
	{Sub: "read-tree"}, {Sub: "checkout-index"}, {Sub: "fast-import"},
	{Sub: "symbolic-ref", Reads: "--short -q --quiet", Writes: "-d --delete -m"},
	{Sub: "merge-file", Reads: "-p --stdout"},
	{Sub: "format-patch", Reads: "--stdout"},
	{Sub: "rerere", Reads: "status diff remaining", First: true},
	{Sub: "config", Reads: "--get --get-all --get-regexp --get-urlmatch --get-color --get-colorbool --list -l get list",
		Writes: "--add --unset --unset-all --replace-all --rename-section --remove-section --edit -e"},
	{Sub: "branch", Reads: "-l --list -a --all -r --remotes --contains --no-contains --merged --no-merged --points-at --show-current",
		Writes: "-d -D --delete -m -M --move -c -C --copy -f --force -u --set-upstream-to --unset-upstream --edit-description " +
			"-t --track --no-track --create-reflog", Lists: true},
	{Sub: "tag", Reads: "-l --list -v --verify --contains --no-contains --merged --no-merged --points-at",
		Writes: "-d --delete -f --force -a --annotate -s --sign -u --local-user -m --message -F --file -e --edit --cleanup", Lists: true},
}

// gitWrite says which of git's write words a git call's subcommand is, "" for one that only reads (GitWriteWords'
// comment). sub is its subcommand, rest the words after it. A word is matched as written, its value after = cut off
// (--set-upstream-to=x), and a short cluster (-dr) by each of its letters. Anything the table does not read as a
// listing or a dry run is a write.
func gitWrite(sub string, rest []shWord) string {
	for _, g := range GitWriteWords {
		if g.Sub != sub {
			continue
		}
		reads, writes := strings.Fields(g.Reads), strings.Fields(g.Writes)
		if len(rest) == 0 && (sub == "branch" || sub == "tag") {
			return "" // a bare branch or tag lists
		}
		in := func(list []string, w string) bool {
			name, _, _ := strings.Cut(w, "=")
			if listHas(list, name) {
				return true
			}
			if len(w) > 2 && w[0] == '-' && w[1] != '-' && !strings.Contains(w, "=") {
				for _, c := range w[1:] {
					if listHas(list, "-"+string(c)) {
						return true
					}
				}
			}
			return false
		}
		named, read, dashes := false, false, false
		for i, w := range rest {
			t := w.text
			switch {
			case dashes || !strings.HasPrefix(t, "-") || t == "-":
				named = true
				if i == 0 && listHas(reads, t) {
					read = true
				}
			case t == "--":
				dashes = true
			case in(writes, t):
				return "git " + sub
			case !g.First && in(reads, t):
				read = true
			}
		}
		if sub == "symbolic-ref" && gitOperands(rest) > 1 {
			return "git " + sub // two names: it sets the first, --short or not
		}
		if read || (g.Lists && !named) {
			return ""
		}
		return "git " + sub
	}
	return ""
}

// gitOperands counts a git subcommand's words that are not options (all of them after --).
func gitOperands(rest []shWord) int {
	n, dashes := 0, false
	for _, w := range rest {
		switch {
		case dashes || !strings.HasPrefix(w.text, "-") || w.text == "-":
			n++
		case w.text == "--":
			dashes = true
		}
	}
	return n
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
	installOpts = delOpt{short: "bcCdDpsTv", shortVal: "gmoStZ", long: []string{"compare", "directory", "preserve-timestamps",
		"strip", "no-target-directory", "verbose", "preserve-context", "debug"}, longVal: []string{"group", "mode", "owner",
		"suffix", "target-directory", "strip-program"}, longOpt: []string{"backup", "context"}, permute: true}
	touchOpts = delOpt{short: "acfhm", shortVal: "dtr", long: []string{"no-create", "no-dereference"},
		longVal: []string{"date", "reference", "time"}, permute: true}
	truncateOpts = delOpt{short: "co", shortVal: "rs", long: []string{"no-create", "io-blocks"},
		longVal: []string{"reference", "size"}, permute: true}
	mkfifoOpts = delOpt{short: "Z", shortVal: "m", long: []string{"help", "version"}, longVal: []string{"mode"},
		longOpt: []string{"context"}, permute: true}
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
	psItem = delPSCmdlet{params: []string{"path", "literalpath", "value", "filter", "include", "exclude", "force", "passthru",
		"credential", "stream", "whatif", "confirm"}, switches: []string{"force", "passthru", "whatif", "confirm", "verbose", "debug"},
		aliases: map[string]string{"pspath": "literalpath", "lp": "literalpath", "wi": "whatif", "cf": "confirm"}, positional: []string{"path", "value"}}
	psExportCsv = delPSCmdlet{params: []string{"path", "literalpath", "inputobject", "force", "noclobber", "encoding", "append",
		"delimiter", "useculture", "notypeinformation", "includetypeinformation", "quotefields", "usequotes", "noheader", "whatif",
		"confirm"}, switches: []string{"force", "noclobber", "append", "useculture", "notypeinformation", "includetypeinformation",
		"noheader", "whatif", "confirm", "verbose", "debug"}, aliases: map[string]string{"pspath": "literalpath", "lp": "literalpath",
		"nti": "notypeinformation", "wi": "whatif", "cf": "confirm"}, positional: []string{"path", "delimiter"}}
	psTranscript = delPSCmdlet{params: []string{"path", "literalpath", "outputdirectory", "append", "force", "noclobber",
		"includeinvocationheader", "useminimalheader", "whatif", "confirm"}, switches: []string{"append", "force", "noclobber",
		"includeinvocationheader", "useminimalheader", "whatif", "confirm", "verbose", "debug"},
		aliases: map[string]string{"pspath": "literalpath", "lp": "literalpath", "wi": "whatif", "cf": "confirm"}, positional: []string{"path"}}
	psWeb = delPSCmdlet{params: []string{"uri", "outfile", "method", "body", "headers", "contenttype", "infile", "usebasicparsing",
		"usedefaultcredentials", "credential", "certificate", "certificatethumbprint", "useragent", "disablekeepalive",
		"timeoutsec", "maximumredirection", "proxy", "proxycredential", "proxyusedefaultcredentials", "sessionvariable",
		"websession", "passthru", "transferencoding", "skipcertificatecheck", "skipheadervalidation", "resume", "form",
		"authentication", "token", "maximumretrycount", "retryintervalsec", "statuscodevariable", "responseheaderstvariable"},
		switches: []string{"usebasicparsing", "usedefaultcredentials", "disablekeepalive", "proxyusedefaultcredentials", "passthru",
			"skipcertificatecheck", "skipheadervalidation", "resume", "verbose", "debug"},
		aliases: map[string]string{"sv": "sessionvariable"}, positional: []string{"uri"}}
)

// writeTarget is one path a command writes, and whether it may be a folder its sources land in (cp, ln, Copy-Item:
// dest/<a source's name> is judged too).
type writeTarget struct {
	op      delOp
	srcs    []delOp
	destDir bool
	word    string // the write word, as a refusal names it
	attr    bool   // a change of its permissions, owner, ACL or attributes (attrwords.go): judged by what it reaches too
}

// writeTargets gives the paths a write word writes (this file's comment), for a command whose program name and
// native-ness delName gives; none for a command that is no write word.
func writeTargets(name string, native bool, x delExec) []writeTarget {
	if ts := attrTargets(name, native, x); len(ts) > 0 {
		return ts
	}
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
		case "clear-content", "clc", "set-item", "si", "clear-item", "cli":
			b := delPSBind(words, psItem)
			return one(cmdletName(name), append(append([]shWord(nil), b.vals["path"]...), b.vals["literalpath"]...))
		case "export-csv", "epcsv":
			b := delPSBind(words, psExportCsv)
			return one("Export-Csv", append(append([]shWord(nil), b.vals["path"]...), b.vals["literalpath"]...))
		case "start-transcript":
			b := delPSBind(words, psTranscript)
			return one("Start-Transcript", append(append(append([]shWord(nil), b.vals["path"]...), b.vals["literalpath"]...),
				b.vals["outputdirectory"]...))
		case "invoke-webrequest", "iwr", "invoke-restmethod", "irm", "curl", "wget":
			return one(cmdletName(name)+" -OutFile", delPSBind(words, psWeb).vals["outfile"])
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
			var out []writeTarget
			for _, op := range ops(ws[len(ws)-1:]) {
				out = append(out, writeTarget{op: op, srcs: ops(ws[:len(ws)-1]), destDir: true, word: name})
			}
			return out
		}
		return nil
	}
	switch name {
	case "tee":
		return one("tee", delParse(words, teeOpts).ops)
	case "touch":
		return one("touch", delParse(words, touchOpts).ops)
	case "truncate":
		return one("truncate", delParse(words, truncateOpts).ops)
	case "install":
		if g := delParse(words, installOpts); g.has("d", "directory") {
			return one("install -d", g.ops)
		}
		return copyLike("install", installOpts)
	case "mkfifo":
		return one("mkfifo", delParse(words, mkfifoOpts).ops)
	case "mknod":
		if g := delParse(words, mkfifoOpts); len(g.ops) > 0 {
			return one("mknod", g.ops[:1]) // its name; the type and numbers after it are no path
		}
		return nil
	case "git":
		return gitOutputs(words, ops)
	case "dd":
		var outs []shWord
		for _, w := range words {
			if v, ok := strings.CutPrefix(w.text, "of="); ok {
				outs = append(outs, delValue(w, v))
			}
		}
		return one("dd of=", outs)
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

// gitOutputs gives the files a git call writes by an option of its own, outside its write words: --output=<file> of
// the diff family (diff, log, show and their kind) and of archive (and archive's -o <file>), and format-patch's -o or
// --output-directory <folder>. words are git's own from its options on.
func gitOutputs(words []shWord, ops func([]shWord) []delOp) []writeTarget {
	sub, rest := gitSub(words)
	var outs []shWord
	take := func(names ...string) {
		for i := 0; i < len(rest); i++ {
			t := rest[i].text
			if t == "--" {
				return
			}
			for _, n := range names {
				switch {
				case t == n && i+1 < len(rest):
					outs = append(outs, rest[i+1])
					i++
				case strings.HasPrefix(n, "--") && strings.HasPrefix(t, n+"="):
					outs = append(outs, delValue(rest[i], strings.TrimPrefix(t, n+"=")))
				case len(n) == 2 && strings.HasPrefix(t, n) && len(t) > 2 && !strings.HasPrefix(t, "--"):
					outs = append(outs, delValue(rest[i], t[2:]))
				}
			}
		}
	}
	switch sub {
	case "diff", "diff-files", "diff-index", "diff-tree", "log", "show", "whatchanged", "range-diff":
		take("--output")
	case "archive":
		take("--output", "-o")
	case "format-patch":
		take("--output-directory", "-o")
	default:
		return nil
	}
	var out []writeTarget
	for _, op := range ops(outs) {
		out = append(out, writeTarget{op: op, word: "git " + sub + " --output"})
	}
	return out
}

// cmdletName names a content cmdlet by its full name, for a refusal.
func cmdletName(name string) string {
	switch name {
	case "set-content", "sc":
		return "Set-Content"
	case "add-content", "ac":
		return "Add-Content"
	case "clear-content", "clc":
		return "Clear-Content"
	case "set-item", "si":
		return "Set-Item"
	case "clear-item", "cli":
		return "Clear-Item"
	case "invoke-webrequest", "iwr", "curl", "wget":
		return "Invoke-WebRequest"
	case "invoke-restmethod", "irm":
		return "Invoke-RestMethod"
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
			if r.writePath(rs, x, t, role); r.deny != nil {
				return
			}
		}
	}
}

// writePath judges one written path in each folder the command may run in, and a change of its permissions or
// attributes by what it reaches too (attrReach); a path held in a variable or a substitution the check cannot read is
// not judged (a known limit, this file's comment).
func (r *delRun) writePath(rs *shellRole, x delExec, t writeTarget, role bool) {
	word := t.word
	ops := delLiterals([]delOp{t.op}, x)
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
			if !t.attr || role {
				continue
			}
			if d := rs.attrReach(dir, text, word, op.glob && !op.literal, x.d); !d.Allow {
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
		w := gitWrite(sub, rest)
		if gitInlineAlias(x.words[1:]) {
			w = "git -c alias.<name>=..." // an alias given on the line can name any subcommand
		}
		if w != "" {
			r.refuse(Decision{Rule: RuleRoleWritesNothing, Why: fmt.Sprintf("%s refused: %s changes the checkout or its history, "+
				"and %s changes nothing (its role's writes: false)", r.tool, w, rs.w.shown()),
				Next: "put what you would change in your report; reading git (status, diff, log, show) stays allowed"})
		}
	}
}

// gitInlineAlias reports a git call that defines an alias on its own line (-c alias.<name>=..., or --config-env
// alias.<name>=...), before its subcommand: the subcommand it runs could be any of git's words, so a role that changes
// nothing is refused it. An alias already in git's own config files is a known limit (this file's comment).
func gitInlineAlias(words []shWord) bool {
	for i := 0; i < len(words); i++ {
		t := words[i].text
		v := ""
		switch {
		case (t == "-c" || t == "--config-env") && i+1 < len(words):
			v = words[i+1].text
			i++
		case strings.HasPrefix(t, "-c") && len(t) > 2:
			v = t[2:]
		case strings.HasPrefix(t, "--config-env="):
			v = strings.TrimPrefix(t, "--config-env=")
		case !strings.HasPrefix(t, "-"):
			return false // the subcommand: options after it are its own
		case listHas([]string{"-C", "--git-dir", "--work-tree", "--namespace", "--super-prefix"}, t):
			i++
		}
		if strings.HasPrefix(strings.ToLower(v), "alias.") {
			return true
		}
	}
	return false
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

// installedBonsai reports a program word that runs the installed bonsai by the very place the hook lines run it from
// (engine's PlaceWindows and PlaceLinux, or a test build's stamped place), in any form a shell names it: forward or
// back slashes, a Windows place as C:/... or /c/..., its letter case and its .exe ignored. Past a budget only it is the
// report or a brief's read (VG's item D; VG round 2's S2, 15 Oct: a bare bonsai runs whatever the PATH finds first,
// ~/go/bin or ~/.local/bin before /usr/local/bin, so it is no longer taken). Any other program named bonsai is not the
// installed one. Its word is literal (no variable, no substitution, no glob).
func installedBonsai(w shWord) bool {
	if w.expand || w.glob || w.splat || w.alts != nil {
		return false
	}
	t := w.text
	norm := func(p string, windows bool) string {
		p = strings.ReplaceAll(p, `\`, "/")
		if len(p) >= 2 && p[1] == ':' && isLetter(p[0]) {
			p = "/" + strings.ToLower(p[:1]) + p[2:]
		}
		if windows {
			p = strings.TrimSuffix(strings.ToLower(p), ".exe")
		}
		return p
	}
	for _, p := range hookPlaces() {
		windows := strings.HasSuffix(strings.ToLower(p), ".exe") || (len(p) >= 3 && p[0] == '/' && p[2] == '/') ||
			(len(p) >= 2 && p[1] == ':')
		if norm(t, windows) == norm(p, windows) {
			return true
		}
	}
	return false
}

// hookPlaces are the places the hook lines run bonsai from: a test build's stamped place alone, else engine's two.
func hookPlaces() []string {
	if stamped, _ := engine.TestBuild(); stamped != "" {
		return []string{filepath.ToSlash(stamped)}
	}
	return []string{engine.PlaceWindows, engine.PlaceLinux}
}

// hookPlace is the installed bonsai as a next step names it (VG round 2's S2, as the stop gate's next steps name it):
// the place this system's hook lines run it from (engine.PlaceLinux; engine.PlaceWindows written C:/...; a test
// build's stamped place), in double quotes when it holds a space, after PowerShell's call operator (& ) for ps.
func hookPlace(ps bool) string {
	windows := runtime.GOOS == "windows"
	place := engine.PlaceLinux
	if windows {
		place = engine.PlaceWindows
	}
	if stamped, _ := engine.TestBuild(); stamped != "" {
		place = filepath.ToSlash(stamped)
	}
	if windows && len(place) >= 3 && place[0] == '/' && place[2] == '/' {
		place = strings.ToUpper(place[1:2]) + ":" + place[2:]
	}
	if strings.Contains(place, " ") {
		place = `"` + place + `"`
	}
	if ps {
		place = "& " + place
	}
	return place
}

// bareBonsai reports a program word that is a bare bonsai (or bonsai.exe), as the shell finds it on its PATH: a
// brief's read by it ties the brief (the tie reads the brief's own file, so which program ran does not count there),
// but past a budget it is neither the report nor an exempt read (installedBonsai).
func bareBonsai(w shWord) bool {
	return !w.expand && !w.glob && !w.splat && w.alts == nil && (w.text == "bonsai" || strings.EqualFold(w.text, "bonsai.exe"))
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
	words := programWords(d, c.words)
	// One command: a line end (or ;) may close it, nothing may follow it.
	if c.group != nil || len(c.nested) > 0 || len(c.redirs) > 0 || (c.op != "" && c.op != ";") || c.piped || len(words) < 2 {
		return false
	}
	if len(c.feeds) > 1 || (len(c.feeds) == 1 && (!c.feeds[0].doc || !c.feeds[0].quoted || len(c.feeds[0].nested) > 0)) {
		return false
	}
	if !installedBonsai(words[0]) {
		return false
	}
	kind, from, body, bodyDash := "", "", false, false
	for i := 1; i < len(words); i++ {
		w := words[i]
		if w.expand || w.glob || w.splat || w.alts != nil {
			return false
		}
		name, val, hasVal := strings.Cut(w.text, "=")
		next := func() (string, bool) {
			if hasVal {
				return val, true
			}
			if i+1 >= len(words) {
				return "", false
			}
			i++
			nw := words[i]
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
			if i+1 >= len(words) {
				return false
			}
			i++
			bw := words[i]
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

// briefShow gives the id a shell call reads with bonsai message --show <id> as its one command, "" for any other call,
// and whether it runs the installed bonsai at its place (installedBonsai); a bare bonsai reads too (bareBonsai).
func briefShow(tool, command string) (id string, atPlace bool) {
	d := shBash
	if tool == "PowerShell" {
		d = shPower
	}
	l, why := shSplit(command, d, 0)
	if why != "" || l == nil || len(l.cmds) != 1 {
		return "", false
	}
	c := l.cmds[0]
	words := programWords(d, c.words)
	if c.group != nil || len(c.nested) > 0 || len(c.redirs) > 0 || len(c.feeds) > 0 || (c.op != "" && c.op != ";") || len(words) < 3 {
		return "", false
	}
	atPlace = installedBonsai(words[0])
	if !atPlace && !bareBonsai(words[0]) || words[1].text != "message" {
		return "", false
	}
	for i := 2; i < len(words); i++ {
		w := words[i]
		if w.expand || w.glob {
			return "", false
		}
		switch {
		case w.text == "--show" && i+1 < len(words):
			i++
			id = words[i].text
		case strings.HasPrefix(w.text, "--show="):
			id = strings.TrimPrefix(w.text, "--show=")
		case w.text == "--json":
		default:
			return "", false
		}
	}
	if !message.IsID(id) {
		return "", false
	}
	return id, atPlace
}
