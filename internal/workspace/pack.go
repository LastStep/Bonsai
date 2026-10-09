package workspace

// A pack's bonsai/pack.yaml (bonsai.pack/1, spec §5): what the engine reads from a pack and writes into a project.
//
// Read here: format, id, version, needs.claude_code, block, files (path, from, kind), hooks (event, matcher,
// command, runs, why) and deny (rule, why): every field the test pack's pack.yaml holds (its commits A to F), which is
// what the walking skeleton's engine (part 3), hook path (part 5) and consent to code (step 5.1.1) use. Any other
// key, and any other key under needs, is kept in Pack.Doc as read and not checked; bonsai.pack/1's schema, its
// documentation check (bonsai check --pack) and the rule that a hook line never calls bash by name come later in
// step 5.1.
//
// A hook entry's runs (step 5.1.1, plan-5 rule 4) lists the pack's own files its command runs, each by the path its
// files entry gives it in the project (test-pack/run.sh), so a change to one of them is "runs code" even when the
// hook line stays the same (spec §6: "a change to a hook line or to a file a hook runs"). [] for none; a hook entry
// with no runs reads as [] (the reader rule for a missing field). Each item must be one of the files entries' paths:
// a hook may name only files the pack writes, and Bonsai consents to no file it cannot see. Example:
//
//	hooks:
//	  - event: SessionStart
//	    matcher: startup
//	    command: "sh test-pack/run.sh"
//	    runs: ["test-pack/run.sh"]
//	    why: "Prints the pack's greeting when a session starts; it blocks nothing."
//
// bonsai check --pack (step 5.1.9) refuses a hook command naming a pack file that runs does not list.

import (
	"regexp"
	"strings"

	"github.com/LastStep/Bonsai/internal/reader"
)

// PackFile is a pack's manifest's path inside the pack.
const PackFile = "bonsai/pack.yaml"

// PackFormat is the format: line a pack.yaml carries.
const PackFormat = "bonsai.pack/1"

// packIDPattern is a pack id: lower-case letters, digits and dashes, starting with a letter. It equals the
// plugin's name in .claude-plugin/plugin.json, so roles load as <id>:<role> (spec §5).
var packIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// shortName matches a path segment shaped like an NTFS short (8.3) name: a tilde, then a digit (CLAUDE~1,
// PROTEC~1.TXT).
var shortName = regexp.MustCompile(`~[0-9]`)

// PackFileKinds are the kinds a pack gives its files: pack (Bonsai's; an edit in the project is a conflict) and
// once (written once, then the project's). The lock's other kinds come from elsewhere: block from block, keys from
// hooks and deny, kept from the engine (the test pack's pack.yaml documents this). A test holds each to the lock
// schema's list of kinds, their one home.
var PackFileKinds = []string{"pack", "once"}

// Pack is what Bonsai reads from a pack's bonsai/pack.yaml.
type Pack struct {
	ID         string      // the pack's id
	Version    string      // the pack's version; a release tag must equal it
	ClaudeCode string      // needs.claude_code: the oldest Claude Code the pack works with, "" for none
	Block      string      // the pack's part of the instruction block, a file in bonsai/, "" for none
	Files      []FileEntry // the files the engine writes into a project, from bonsai/files/
	Hooks      []HookEntry // hook lines for the project's .claude/settings.json
	Deny       []DenyEntry // deny rules for the project's .claude/settings.json
	Doc        *reader.Map // the whole file as read, every key kept
}

// FileEntry is one of a pack's files.
type FileEntry struct {
	Path string // where it goes in the project: project-relative, forward slashes
	From string // its source in the pack's bonsai/files/
	Kind string // pack or once
	Line int
}

// HookEntry is one hook line.
type HookEntry struct {
	Event   string   // the Claude Code hook event
	Matcher string   // the event's matcher, "" for none
	Command string   // the shell line Claude Code runs
	Runs    []string // the pack's files the command runs, by their path in the project; [] for none
	Why     string   // one plain sentence: what the line does, as update's preview prints it
	Line    int
}

// DenyEntry is one deny rule.
type DenyEntry struct {
	Rule string // a Claude Code permission rule
	Why  string // one plain sentence: what the rule stops
	Line int
}

const packNext = "fix that line of the pack's bonsai/pack.yaml, then release the pack again"

// ReadPack reads a pack's bonsai/pack.yaml bytes.
func ReadPack(raw []byte) (*Pack, error) {
	m, err := readYAML(PackFile, raw, PackFormat, packNext)
	if err != nil {
		return nil, err
	}
	f := &fields{file: PackFile, next: packNext}
	p := &Pack{Doc: m}
	p.ID = f.text(m, "id", "the file", 0, true)
	if e, _ := m.Entry("id"); f.err == nil && !packIDPattern.MatchString(p.ID) {
		f.fail(e.Line, "the id %s is not a pack id: a lower-case letter, then lower-case letters, digits and dashes", showValue(p.ID))
	}
	p.Version = f.text(m, "version", "the file", 0, true)
	if needs := f.mapping(m, "needs", "the file"); needs != nil {
		p.ClaudeCode = f.text(needs, "claude_code", "needs", 0, false)
	}
	p.Block = f.text(m, "block", "the file", 0, false)
	if e, _ := m.Entry("block"); f.err == nil && p.Block != "" {
		if err := CheckRelPath(p.Block); err != nil {
			f.fail(e.Line, "block: %v", err)
		}
	}

	items, line := f.list(m, "files", "the file")
	p.Files = []FileEntry{}
	seen := map[string]bool{}
	for i, it := range items {
		fm, first := entryOf(f, it, line, "files", i, "path, from and kind")
		if fm == nil {
			break
		}
		where := "files item " + itoa(i+1)
		fe := FileEntry{Line: first}
		fe.Path = f.text(fm, "path", where, first, true)
		fe.From = f.text(fm, "from", where, first, true)
		fe.Kind = f.text(fm, "kind", where, first, true)
		if f.err != nil {
			break
		}
		if err := CheckRelPath(fe.Path); err != nil {
			f.fail(first, "%s's path: %v", where, err)
		} else if err := checkPackTarget(fe.Path); err != nil {
			f.fail(first, "%s's path: %v", where, err)
		} else if err := CheckRelPath(fe.From); err != nil {
			f.fail(first, "%s's from: %v", where, err)
		} else if !contains(PackFileKinds, fe.Kind) {
			f.fail(first, "%s's kind %s is not pack or once", where, showValue(fe.Kind))
		} else if folded := strings.ToLower(fe.Path); seen[folded] {
			f.fail(first, "%s's path %s is listed twice (letter case aside, as Windows sees it)", where, showValue(fe.Path))
		} else {
			seen[folded] = true
		}
		p.Files = append(p.Files, fe)
	}

	items, line = f.list(m, "hooks", "the file")
	p.Hooks = []HookEntry{}
	for i, it := range items {
		hm, first := entryOf(f, it, line, "hooks", i, "event, matcher, command, runs and why")
		if hm == nil {
			break
		}
		where := "hooks item " + itoa(i+1)
		h := HookEntry{Line: first}
		h.Event = f.text(hm, "event", where, first, true)
		h.Matcher = f.text(hm, "matcher", where, first, false)
		h.Command = f.text(hm, "command", where, first, true)
		h.Runs = f.texts(hm, "runs", where)
		h.Why = f.text(hm, "why", where, first, true)
		if f.err == nil {
			runsLine := first
			if e, ok := hm.Entry("runs"); ok {
				runsLine = e.Line
			}
			for _, r := range h.Runs {
				if !isFilePath(p.Files, r) {
					f.fail(runsLine, "%s's runs names %s, which no files entry writes: a hook's runs lists only the pack's own files, "+
						"by their path in the project", where, showValue(r))
					break
				}
			}
		}
		p.Hooks = append(p.Hooks, h)
	}

	items, line = f.list(m, "deny", "the file")
	p.Deny = []DenyEntry{}
	for i, it := range items {
		dm, first := entryOf(f, it, line, "deny", i, "rule and why")
		if dm == nil {
			break
		}
		where := "deny item " + itoa(i+1)
		d := DenyEntry{Line: first}
		d.Rule = f.text(dm, "rule", where, first, true)
		d.Why = f.text(dm, "why", where, first, true)
		p.Deny = append(p.Deny, d)
	}
	if f.err != nil {
		return nil, f.err
	}
	return p, nil
}

// isFilePath reports whether path is one of the pack's files entries' paths.
func isFilePath(files []FileEntry, path string) bool {
	for _, fe := range files {
		if fe.Path == path {
			return true
		}
	}
	return false
}

// entryOf takes a list item that must be a mapping, and its first line.
func entryOf(f *fields, it any, listLine int, list string, i int, holds string) (*reader.Map, int) {
	m, ok := it.(*reader.Map)
	if !ok || m.Len() == 0 {
		f.fail(listLine, "%s item %d is %s, not a mapping of %s", list, i+1, kindOf(it), holds)
		return nil, 0
	}
	return m, m.Entries()[0].Line
}

// checkPackTarget refuses a pack file aimed at a path Bonsai keeps for itself (letter case aside): bonsai.yaml and
// .bonsai/, and ReservedTargets; and any path with a segment shaped like an NTFS short (8.3) name, <name>~<digit>,
// which on Windows can be another name for one of those (CLAUDE~1/settings.json is .claude/settings.json).
// CheckRelPath refuses GIT~1 the same way.
func checkPackTarget(p string) error {
	for _, seg := range strings.Split(p, "/") {
		if shortName.MatchString(seg) {
			return pathError(p, "has a segment shaped like a Windows short (8.3) name ("+asciiOnly(seg)+
				"), which can stand for another file or folder; name it in full")
		}
	}
	l := strings.ToLower(p)
	if l == "bonsai.yaml" || l == ".bonsai" || strings.HasPrefix(l, ".bonsai/") {
		return pathError(p, "is Bonsai's own file; a pack cannot write it")
	}
	for _, reserved := range ReservedTargets {
		if l == strings.ToLower(reserved) {
			return pathError(p, "is a file the engine writes only its own part of; a pack cannot write it whole")
		}
	}
	return nil
}

// ReservedTargets are the project files a pack's files entry may not name, beside bonsai.yaml and .bonsai/: the
// engine keeps the instruction block in CLAUDE.md (kind block) and its own entries in .claude/settings.json (kind
// keys), and .claude/settings.local.json is a person's own file.
var ReservedTargets = []string{"CLAUDE.md", ".claude/settings.json", ".claude/settings.local.json"}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
