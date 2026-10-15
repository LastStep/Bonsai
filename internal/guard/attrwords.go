package guard

// Changes of a path's permissions, owner, ACL or attributes on an agent's shell line (VG round 2's M1, 15 Oct: an
// agent's chmod a-w on its own calls file left it uncounted, so its budget never ran out; chmod 000 on its state gave
// it a fresh window each time). Each such word's paths are judged as the shell writes are (rolecmds.go's mayWrite):
// Bonsai's own files refused to every agent but the person's session (bonsai-files), and a path inside a linked project
// refused to a role that changes nothing (role-writes-nothing). Beyond a write, such a change reaches what a folder
// holds, so each path is judged against Bonsai's own files too (attrReach): one that is one of them, a folder that
// holds one (the main checkout's .bonsai/local/, .bonsai/.gitignore, the generated tables and the lock, and Bonsai's
// home: a folder's mode stops Bonsai reaching what is in it, so chmod 000 .bonsai or chmod -R a-w . is refused), or a
// glob that may name either is refused as bonsai-files.
//
// The words, read in every shell (Windows' programs from Git Bash, PowerShell or cmd; the Unix ones from pwsh too):
//   - chmod (a mode written -w is the mode, not an option; --reference=<file> names no mode), chown and chgrp (the
//     owner or group first; --from, --reference), chattr (+i, -i, =i; -R, -V, -f; -v and -p take a value), setfacl
//     (-m, -x, -M, -X, --set, --modify and the rest: every operand a path), setfattr (-n, -v, -x), chcon (a context
//     first, unless -u, -r, -l, -t or --reference gives it) and BSD's chflags (its flags first);
//   - Windows' attrib (with a +x or -x attribute; without one it only lists), icacls (with /grant, /deny, /remove,
//     /reset, /setowner, /setintegritylevel, /inheritance, /restore or /substitute; without one it only shows, and
//     /save's file is a write), cacls (with /e, /g, /r, /p, /d or /s:<sddl>) and takeown (/f's file), a switch also as
//     Git Bash passes it, //x;
//   - PowerShell's Set-ItemProperty (sp), Clear-ItemProperty (clp) and Set-Acl, by -Path or -LiteralPath, and an
//     assignment to a member of a Get-Item (gi) or Get-ChildItem (gci, ls, dir) result written in parentheses
//     ((Get-Item x).IsReadOnly = $true, (gi x).Attributes += 'ReadOnly', with $( ) too), or a call of its set_... or
//     Set... method ((gi x).SetAccessControl($acl)).
//
// Known limits, beside rolecmds.go's (a path in a variable or a substitution, an interpreter's change, a script run by
// name): an item a pipeline gives a cmdlet (Get-Item x | Set-ItemProperty -Name IsReadOnly -Value $true), a member set
// on a variable ($f.IsReadOnly = $true, $_ in a loop), .NET's own calls ([IO.File]::SetAttributes, SetAccessControl
// on a variable), the items find -exec or xargs gives a word (find . -exec chmod a-w {} +), the paths setfacl
// --restore and setfattr --restore read from a file, and macOS's xattr. For Bonsai's own counts these are backed by the
// count itself (budget.go): a calls or state file the guard cannot write or read has its owner's access set back, and
// where that fails the call is refused as one past a budget is, so no change of a mode or an ACL lets a budget escape.

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// The options of the Unix words, as delParse reads them (chmod and chattr are read by hand: a mode may start with -).
var (
	chownOpts = delOpt{short: "cfvhRHLP", long: []string{"changes", "dereference", "no-dereference", "preserve-root",
		"no-preserve-root", "quiet", "silent", "recursive", "verbose", "help", "version"}, longVal: []string{"from", "reference"},
		permute: true}
	setfaclOpts = delOpt{short: "bkndRLPvhs", shortVal: "mxMX", long: []string{"remove-all", "remove-default", "no-mask", "mask",
		"default", "recursive", "logical", "physical", "test", "version", "help"}, longVal: []string{"modify", "remove", "set",
		"modify-file", "remove-file", "set-file", "restore"}, permute: true}
	setfattrOpts = delOpt{short: "h", shortVal: "nvx", long: []string{"no-dereference", "raw", "version", "help"},
		longVal: []string{"name", "value", "remove", "restore"}, permute: true}
	chconOpts = delOpt{short: "hRHLPv", shortVal: "urlt", long: []string{"dereference", "no-dereference", "preserve-root",
		"no-preserve-root", "recursive", "verbose", "help", "version"}, longVal: []string{"user", "role", "range", "type",
		"reference"}, permute: true}
	chflagsOpts = delOpt{short: "fhvRHLPx", permute: true}
)

// PowerShell's attribute cmdlets, their parameters as delPSBind reads them.
var (
	psItemProperty = delPSCmdlet{params: []string{"path", "literalpath", "name", "value", "inputobject", "passthru", "force",
		"filter", "include", "exclude", "credential", "type", "whatif", "confirm"}, switches: []string{"passthru", "force",
		"whatif", "confirm", "verbose", "debug"}, aliases: map[string]string{"pspath": "literalpath", "lp": "literalpath",
		"psproperty": "name", "wi": "whatif", "cf": "confirm"}, positional: []string{"path", "name", "value"}}
	psSetACL = delPSCmdlet{params: []string{"path", "literalpath", "aclobject", "inputobject", "passthru", "filter", "include",
		"exclude", "clearcentralaccesspolicy", "whatif", "confirm"}, switches: []string{"passthru", "clearcentralaccesspolicy",
		"whatif", "confirm", "verbose", "debug"}, aliases: map[string]string{"pspath": "literalpath", "lp": "literalpath",
		"wi": "whatif", "cf": "confirm"}, positional: []string{"path", "aclobject"}}
	psGetItem = delPSCmdlet{params: []string{"path", "literalpath", "filter", "include", "exclude", "force", "credential",
		"stream", "recurse", "depth", "name", "attributes", "directory", "file", "hidden", "readonly", "system"},
		switches: []string{"force", "recurse", "name", "directory", "file", "hidden", "readonly", "system", "verbose", "debug"},
		aliases:  map[string]string{"pspath": "literalpath", "lp": "literalpath"}, positional: []string{"path", "filter"}}
)

// The Windows programs' switches (winSwitch), and those that make icacls and cacls change what they name.
var (
	attribSwitches  = []string{"s", "d", "l"}
	icaclsSwitches  = []string{"grant", "deny", "remove", "setintegritylevel", "inheritance", "reset", "setowner", "findsid", "verify", "save", "restore", "substitute", "t", "c", "l", "q"}
	icaclsChanges   = []string{"grant", "deny", "remove", "setintegritylevel", "inheritance", "reset", "setowner", "restore", "substitute"}
	caclsSwitches   = []string{"t", "m", "l", "s", "e", "c", "g", "r", "p", "d"}
	caclsChanges    = []string{"e", "g", "r", "p", "d"}
	takeownSwitches = []string{"s", "u", "p", "f", "a", "r", "d"}
)

// attrTargets gives the paths a permission, owner, ACL or attribute word changes (this file's comment), each marked
// attr, and any file it writes (icacls /save); none for a command that is no such word, or one that only reads.
func attrTargets(name string, native bool, x delExec) []writeTarget {
	words := x.words[1:]
	targets := func(word string, ws []shWord, attr bool) []writeTarget {
		var out []writeTarget
		for _, op := range delOps(ws, x.d, false, false) {
			out = append(out, writeTarget{op: op, word: word, attr: attr})
		}
		return out
	}
	psPaths := func(b delPSBound) []shWord {
		return append(append([]shWord(nil), b.vals["path"]...), b.vals["literalpath"]...)
	}
	if x.d == shPower && !native {
		switch name {
		case "set-itemproperty", "sp":
			return targets("Set-ItemProperty", psPaths(delPSBind(words, psItemProperty)), true)
		case "clear-itemproperty", "clp":
			return targets("Clear-ItemProperty", psPaths(delPSBind(words, psItemProperty)), true)
		case "set-acl":
			return targets("Set-Acl", psPaths(delPSBind(words, psSetACL)), true)
		}
	}
	switch name {
	case "chmod":
		return targets("chmod", chmodFiles(words), true)
	case "chown", "chgrp":
		g := delParse(words, chownOpts)
		return targets(name, firstLeft(g.ops, !g.has("reference")), true)
	case "chattr":
		return targets("chattr", chattrFiles(words), true)
	case "setfacl":
		return targets("setfacl", delParse(words, setfaclOpts).ops, true)
	case "setfattr":
		return targets("setfattr", delParse(words, setfattrOpts).ops, true)
	case "chcon":
		g := delParse(words, chconOpts)
		return targets("chcon", firstLeft(g.ops, !g.has("u", "r", "l", "t", "user", "role", "range", "type", "reference")), true)
	case "chflags":
		return targets("chflags", firstLeft(delParse(words, chflagsOpts).ops, true), true)
	case "attrib":
		change := false
		var files []shWord
		for _, w := range words {
			t := w.text
			switch {
			case len(t) == 2 && (t[0] == '+' || t[0] == '-') && isLetter(t[1]):
				change = true
			case winSwitch(t, attribSwitches) == "":
				files = append(files, w)
			}
		}
		if !change {
			return nil // it lists the attributes
		}
		return targets("attrib", files, true)
	case "icacls":
		change := false
		var file, saves []shWord
		for i := 0; i < len(words); i++ {
			switch sw := winSwitch(words[i].text, icaclsSwitches); {
			case sw == "" && file == nil:
				file = words[i : i+1] // its one name comes first; what follows a switch is that switch's
			case sw == "save" && i+1 < len(words):
				saves = append(saves, words[i+1])
				i++
			case listHas(icaclsChanges, sw):
				change = true
			}
		}
		out := targets("icacls /save", saves, false)
		if change {
			out = append(out, targets("icacls", file, true)...)
		}
		return out
	case "cacls":
		change := false
		var files []shWord
		named := true
		for _, w := range words {
			sw := winSwitch(w.text, caclsSwitches)
			switch {
			case sw == "" && named:
				files = append(files, w)
			case sw != "":
				named = false
				change = change || listHas(caclsChanges, sw) || sw == "s" && strings.Contains(w.text, ":")
			}
		}
		if !change {
			return nil // it shows the ACL
		}
		return targets("cacls", files, true)
	case "takeown":
		var files []shWord
		for i := 0; i < len(words); i++ {
			sw := winSwitch(words[i].text, takeownSwitches)
			if (sw == "f" || sw == "s" || sw == "u" || sw == "p" || sw == "d") && i+1 < len(words) {
				if sw == "f" {
					files = append(files, words[i+1])
				}
				i++
			}
		}
		return targets("takeown", files, true)
	}
	return nil
}

// firstLeft drops a word's first operand (its mode, owner, context or flags) when it has one: drop false keeps all.
func firstLeft(ops []shWord, drop bool) []shWord {
	if drop && len(ops) > 0 {
		return ops[1:]
	}
	return ops
}

// chmodFiles gives the files a chmod call changes: its operands after the mode (none is the mode with --reference),
// its options (-c, -f, -v, -R and the long ones) passed over, and a word such as -w or -rx read as the mode, as chmod
// itself reads it.
func chmodFiles(words []shWord) []shWord {
	var files []shWord
	mode, ref, dashes := false, false, false
	for i := 0; i < len(words); i++ {
		w := words[i]
		t := w.text
		if !dashes && !w.expand && len(t) > 1 && t[0] == '-' {
			switch {
			case t == "--":
				dashes = true
				continue
			case strings.HasPrefix(t, "--"):
				name, _, hasVal := strings.Cut(t[2:], "=")
				if len(name) >= 3 && strings.HasPrefix("reference", name) {
					ref = true
					if !hasVal {
						i++ // the reference file is read, not changed
					}
				}
				continue
			case strings.Trim(t[1:], "cfvR") == "":
				continue
			case !mode && !ref:
				mode = true // -w, -rx: the mode
				continue
			}
		}
		if !mode && !ref {
			mode = true
			continue
		}
		files = append(files, w)
	}
	return files
}

// chattrFiles gives the files a chattr call changes: its operands, past its modes (+x, -x, =x) and its options (-R, -V,
// -f; -v and -p with their values).
func chattrFiles(words []shWord) []shWord {
	var files []shWord
	dashes := false
	for i := 0; i < len(words); i++ {
		t := words[i].text
		switch {
		case dashes || words[i].expand || t == "" || t == "-":
			files = append(files, words[i])
		case t == "--":
			dashes = true
		case t == "-v" || t == "-p":
			i++
		case t[0] == '+' || t[0] == '=' || t[0] == '-':
			// a mode, or -R, -V, -f
		default:
			files = append(files, words[i])
		}
	}
	return files
}

// winSwitch gives a Windows program's switch a word is (/name, or //name as Git Bash passes it; any letter case; a
// :value after it), lower case, when its name is one of names; "" for any other word (a path among them).
func winSwitch(t string, names []string) string {
	s, ok := strings.CutPrefix(t, "//")
	if !ok {
		if s, ok = strings.CutPrefix(t, "/"); !ok {
			return ""
		}
	}
	name, _, _ := strings.Cut(s, ":")
	if name = strings.ToLower(name); listHas(names, name) {
		return name
	}
	return ""
}

// programWords drops PowerShell's call operator (&) before a program named by a quoted path, so the program is the
// first word; other shells' words are as split.
func programWords(d shDialect, words []shWord) []shWord {
	if d == shPower && len(words) > 1 && words[0].text == "&" && !words[0].quoted {
		return words[1:]
	}
	return words
}

// psGetters are the cmdlets whose result a member assignment changes (psMemberSet).
var psGetters = []string{"get-item", "gi", "get-childitem", "gci", "ls", "dir"}

// psMemberSet reads a PowerShell command that sets a member of a Get-Item result written in parentheses, or calls its
// set_... or Set... method (this file's comment): the paths it names, as attribute changes; none for any other.
func psMemberSet(words []shWord, depth int) []writeTarget {
	if len(words) == 0 || words[0].quoted {
		return nil
	}
	raw := strings.TrimPrefix(words[0].raw, "$")
	if !strings.HasPrefix(raw, "(") {
		return nil
	}
	end := psClose(raw)
	if end < 0 || end+1 >= len(raw) || raw[end+1] != '.' {
		return nil
	}
	inner, rest := raw[1:end], raw[end+2:]
	k := 0
	for k < len(rest) && shNameChar(rest[k]) {
		k++
	}
	member, after := strings.ToLower(rest[:k]), rest[k:]
	sets := envPSOperator(after) != "" || after == "" && len(words) > 1 && envPSOperator(strings.ToLower(words[1].raw)) != ""
	opens := strings.HasPrefix(after, "(") || after == "" && len(words) > 1 && strings.HasPrefix(words[1].raw, "(")
	calls := opens && (strings.HasPrefix(member, "set") || member == "encrypt" || member == "decrypt")
	if member == "" || !sets && !calls {
		return nil
	}
	l, why := shSplit(inner, shPower, depth+1)
	if why != "" || l == nil || len(l.cmds) != 1 || len(l.cmds[0].words) == 0 {
		return nil
	}
	ws := programWords(shPower, l.cmds[0].words)
	name, native := delName(ws[0].text, shPower)
	if native || !listHas(psGetters, name) {
		return nil
	}
	b := delPSBind(ws[1:], psGetItem)
	var out []writeTarget
	what := "an assignment to a member of " + cmdletShown(name) + "'s result (." + rest[:k] + ")"
	for _, op := range append(delOps(b.vals["path"], shPower, false, false), delOps(b.vals["literalpath"], shPower, true, false)...) {
		out = append(out, writeTarget{op: op, word: what, attr: true})
	}
	return out
}

// cmdletShown names a getter cmdlet by its full name, for a refusal.
func cmdletShown(name string) string {
	if name == "get-item" || name == "gi" {
		return "Get-Item"
	}
	return "Get-ChildItem"
}

// psClose gives the index of the ) that closes the ( a PowerShell word starts with, outside quotes; -1 for none.
func psClose(raw string) int {
	depth := 0
	var quote byte
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '\'' || c == '"':
			quote = c
		case c == '(':
			depth++
		case c == ')':
			if depth--; depth == 0 {
				return i
			}
		}
	}
	return -1
}

// bonsaiOwn gives Bonsai's own files a change of permissions or attributes may reach (attrReach): the floor's
// bonsai-files entries and the lock at the main checkout's top, and Bonsai's home, each as written and with its links
// resolved; read once a run.
func (s *shellRole) bonsaiOwn() []string {
	if s.own != nil {
		return s.own
	}
	s.own = []string{}
	add := func(p string) {
		s.own = appendNew(s.own, filepath.Clean(p))
		s.own = appendNew(s.own, resolve(filepath.Clean(p)))
	}
	for _, f := range Floor {
		if f.Class == RuleBonsaiFiles || f.Glob == ".bonsai/lock.json" {
			add(filepath.Join(s.rr.main, filepath.FromSlash(strings.TrimSuffix(f.Glob, "/"))))
		}
	}
	for _, forms := range s.rr.j.homes() {
		for _, h := range forms {
			add(h)
		}
	}
	return s.own
}

// attrReach judges a change of the permissions, owner, ACL or attributes of the path p, read in dir, against Bonsai's
// own files it reaches (this file's comment): refused as bonsai-files when p is one of them, or a folder holding one,
// or, a glob, may name either (each segment matched as the shell matches it: in bash a * or ? names no dot file).
func (s *shellRole) attrReach(dir, p, word string, glob bool, d shDialect) Decision {
	forms, err := pathForms(dir, dir, p)
	if err != nil {
		return allow(RuleShell, "") // judged by mayWrite already
	}
	fold := caseInsensitive(s.rr.main)
	for _, own := range s.bonsaiOwn() {
		for _, f := range forms {
			if !mayReach(f, own, glob, fold, d == shBash) {
				continue
			}
			dec := deny(RuleBonsaiFiles, fmt.Sprintf("%s refused: %s changes the permissions, owner, ACL or attributes of %s, "+
				"which is or holds one of Bonsai's own files (%s): no agent changes them by a shell line, so no count, brief, "+
				"report or log line is kept from being written or read", s.rr.in.Tool, word, quote(s.rr.j.shown(f), 120),
				quote(s.rr.j.shown(own), 120)), "change only your own files, named one by one; leave Bonsai's files and the folders "+
				"that hold them as they are")
			dec.Target = target(s.rr.j.session, []string{f})
			return dec
		}
	}
	return allow(RuleShell, "")
}

// mayReach reports whether a path (a glob with glob set) is, holds or lies in the folder own: each of the segments the
// two share matched, a glob's by path.Match (a pattern that does not read matches: fail closed), letter case folded
// with fold, and with dots a bash glob's * or ? never stands for a leading dot.
func mayReach(p, own string, glob, fold, dots bool) bool {
	ps := strings.Split(filepath.ToSlash(filepath.Clean(p)), "/")
	owns := strings.Split(filepath.ToSlash(filepath.Clean(own)), "/")
	for i := 0; i < len(ps) && i < len(owns); i++ {
		a, b := ps[i], owns[i]
		if fold {
			a, b = strings.ToLower(a), strings.ToLower(b)
		}
		if !glob || !strings.ContainsAny(a, "*?[") {
			if a != b {
				return false
			}
			continue
		}
		if dots && strings.HasPrefix(b, ".") && !strings.HasPrefix(a, ".") {
			return false
		}
		if ok, err := path.Match(a, b); err == nil && !ok {
			return false
		}
	}
	return true
}
