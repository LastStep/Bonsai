package main

// bonsai pack (T-6009's plan, steps 7 and 7b; the director's answers of 28 Oct, approved 29 Oct): add, remove and
// move a project's packs, and see which are behind, in one command each instead of a hand edit of bonsai.yaml and an
// update.
//
//   - pack outdated: each locked pack's ref, the newest release tag of its source and its state. It reads every
//     source with git ls-remote (the network, PackTagsTimeout each), writes the home's cache/packs.json (which plain
//     bonsai status then shows, offline) and exits 0 whatever is behind; --json prints bonsai.packs/1.
//   - pack set <id>@<ref>, pack add --source <url> [--path <folder>] --ref <ref>, pack remove <id>: each edits
//     bonsai.yaml's bytes in memory (engine's configedit.go: only the ref's text, or one whole entry; every comment
//     kept), then runs update's own run on those bytes (engine.Request.Config): one preview, the bonsai.yaml line and
//     everything the update changes, and with --yes one all-or-nothing write. --json prints bonsai.changes/1, command
//     update and word {pack, <sub>}. update's flags pass through (--allow-exec, --keep, --adopt, --diff, --explain,
//     --expect), and update's gates hold unchanged.
//
// Who may apply: a plan that takes protections away (a pack taken out, or a deny rule or protected path the new
// version drops) is a person's in an agent's session, whatever the task grants (engine.go's loosening gate, for
// update too); pack remove always takes a pack out. The guard judges the line first (internal/guard/commands.go).

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/textui"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func init() { register(packWord) }

var packWord = &Word{
	Name:    "pack",
	Order:   19,
	Title:   "add, remove and move this project's packs, and see which are behind (spec section 4).",
	Summary: "move a pack to another version, add or remove one, or list the packs behind",
	Args:    "<action>",
	About: "bonsai pack outdated asks each pack's source for its newest release (agents may run it); bonsai pack set,\n" +
		"add and remove change bonsai.yaml's packs and run update on it, in one preview and one --yes.\n",
	Exits: []Exit{
		{Code: 0, Means: "done (or the help was printed)"},
		{Code: 2, Means: "bad input: no action, one this build does not have, or the action's own (its --help)"},
		{Code: 3, Means: "the action's: a fetch or a read failed"},
		{Code: 4, Means: "the action's: not linked, no lock, no --yes, a person's command in an agent's session"},
		{Code: 5, Means: "conflicts (the action's)"},
	},
	Examples: []string{"bonsai pack outdated", "bonsai pack set base@base-v1.0.1"},
	Subs:     []*Word{packOutdatedWord, packSetWord, packAddWord, packRemoveWord},
}

// packApplyFlags are the flags pack set, add and remove take: update's.
func packApplyFlags(more ...Flag) []Flag {
	return append(more,
		Flag{Name: "--yes", Help: "write without asking"},
		Flag{Name: "--allow-exec", Help: "consent to what the preview lists under \"Runs code\" (needed as well as --yes), as update's"},
		Flag{Name: "--diff", Help: "show each file's changes as a diff in the preview"},
		Flag{Name: "--explain", Help: "list every settings line in the preview with where it comes from and why (else they are counted)"},
		Flag{Name: "--keep", Value: "P", Many: true, Help: "settle a conflict on P by keeping your edit, as update's"},
		Flag{Name: "--adopt", Value: "P", Many: true, Help: "settle a conflict on P by taking the pack's copy, as update's"},
		Flag{Name: "--expect", Value: "FP", Help: "write only when the plan's fingerprint is FP, the one a --json preview printed (else plan-changed,\nnothing written); it opens no gate"},
		Flag{Name: "--json", Help: "print the changes document (bonsai.changes/1: command update, word pack) instead of text; never asks"},
	)
}

// packApplyExits are pack set's, add's and remove's exit codes: update's.
var packApplyExits = []Exit{
	{Code: 0, Means: "written (or nothing to change)"},
	{Code: 2, Means: "bad input: a flag or value wrong, a pack bonsai.yaml does not name (or already names), a ref this command does\nnot write, or a bonsai.yaml Bonsai does not read (nothing written)"},
	{Code: 3, Means: "runtime: a fetch failed, git or the Bonsai home missing (nothing written); or the write stopped part-way:\nrun update"},
	{Code: 4, Means: "wrong state (not linked, no lock, a moved tag), no --yes, code without --allow-exec, a person's command in\nan agent's session (a plan that takes protections away among them), or a plan whose fingerprint is not\n--expect's (nothing written)"},
	{Code: 5, Means: "conflicts (nothing written)"},
}

const packApplyNotes = `One preview, all or nothing: the bonsai.yaml line (only the ref's text changes; the line's own comment stays as it
is), then every file, settings line and plugin the update changes. Without --yes nothing is written, bonsai.yaml
included; a fetch or a check that fails writes nothing either.
Who may apply: an agent under a task whose bonsai.allows holds bonsai.yaml and .bonsai/lock.json, as with update.
A change that takes protections away (pack remove, or a version that drops a deny rule or a protected path) is a
person's: the preview lists each protection that goes (Loosens; --json: loosening), and in an agent's session --yes
is refused (person-command, exit 4), whatever the task grants. The same holds for update after a hand edit.
`

// packRefused is a pack action's --json for a refusal before a plan: bonsai.changes/1, command update, word pack.
func packRefused(sub string) func(c *call, e *engine.Error) encoder {
	return func(c *call, e *engine.Error) encoder {
		return engine.RefusedChangesWord("update", format.ChangesWord{Name: "pack", Sub: &sub}, c.has("--allow-exec"), e)
	}
}

var packSetWord = &Word{
	Name:    "pack set",
	Title:   "move a pack to another version: its ref in bonsai.yaml, then update, in one preview",
	Summary: "move a pack to another tag or commit",
	Args:    "<id>@<ref> [flags]",
	About: `It sets the pack id's ref in bonsai.yaml to ref (a release tag such as base-v1.0.1, a branch, or a 40-character
commit), changing only the ref's text on its line, and runs update on the result: the preview shows the line before
and after and everything the update changes; --yes writes it all, bonsai.yaml first and the lock last.
`,
	Flags:    packApplyFlags(),
	Notes:    packApplyNotes,
	Exits:    packApplyExits,
	Examples: []string{"bonsai pack set base@base-v1.0.1", "bonsai pack set workflow@v1.0.2 --yes", "bonsai pack set base@base-v1.0.1 --json"},
	Refused:  packRefused("set"),
	Run:      runPackSet,
}

var packAddWord = &Word{
	Name:    "pack add",
	Title:   "link one more pack: a new entry in bonsai.yaml, then update, in one preview",
	Summary: "link a new pack from its repository",
	Args:    "--source <url> --ref <ref> [--path <folder>] [flags]",
	About: `It fetches the pack at ref to read its id from its bonsai/pack.yaml, appends its entry to bonsai.yaml's packs in
init's form (each line with init's comment), and runs update on the result. A project's first pack with roles adds
Bonsai's own guard lines for them, which is a person's command (as update's preview says).
`,
	Flags: packApplyFlags(
		Flag{Name: "--source", Value: "URL", Help: "the git repository the pack comes from"},
		Flag{Name: "--ref", Value: "REF", Help: "the release tag or 40-character commit to take"},
		Flag{Name: "--path", Value: "FOLDER", Help: "the pack's folder inside that repository (none: its top)"}),
	Notes:    packApplyNotes,
	Exits:    packApplyExits,
	Examples: []string{"bonsai pack add --source https://github.com/LastStep/bonsai-workflow.git --ref v1.0.1"},
	Refused:  packRefused("add"),
	Run:      runPackAdd,
}

var packRemoveWord = &Word{
	Name:    "pack remove",
	Title:   "take a pack out: its entry out of bonsai.yaml, then update, in one preview (a person's command)",
	Summary: "take a pack out of the project (a person's command)",
	Args:    "<id> [flags]",
	About: `It takes the pack id's whole entry out of bonsai.yaml (the comment lines above it aligned with its dash go with it;
those of the next entry stay) and runs update on the result: the preview names everything the pack wrote that goes,
and every protection with it (its deny rules and protected paths). Taking a pack out always takes protections away,
so in an agent's session --yes is refused (person-command), whatever the task grants; the preview is anyone's.
`,
	Flags:    packApplyFlags(),
	Notes:    packApplyNotes,
	Exits:    packApplyExits,
	Examples: []string{"bonsai pack remove workflow", "bonsai pack remove workflow --json"},
	Refused:  packRefused("remove"),
	Run:      runPackRemove,
}

var packOutdatedWord = &Word{
	Name:    "pack outdated",
	Title:   "each locked pack's version and the newest release of its source",
	Summary: "list each pack's version and its newest release (reads the network)",
	Args:    "[flags]",
	About: `For each locked pack: its ref, its version, the newest release tag its source holds by the pack's own prefix (the
ref's, else its id's) and its state: newer, current, or stale (not read, or the read older than the lock).
It reads every source with git ls-remote (the network, up to 30 seconds each) and keeps the tags in the Bonsai
home's cache/packs.json, so plain bonsai status shows the same offline. It changes no project file, exits 0 whatever
is behind, and agents may run it.
`,
	Flags: []Flag{{Name: "--json", Help: "print the bonsai.packs/1 document (for programs) instead of plain text"}},
	Exits: []Exit{
		{Code: 0, Means: "the packs were listed, whatever is behind (a source that could not be read is named)"},
		{Code: 2, Means: "bad input: an argument it does not take, or a bonsai.yaml Bonsai does not read"},
		{Code: 3, Means: "the Bonsai home cannot be found"},
		{Code: 4, Means: "not linked, or no lock (nothing is locked yet: bonsai update)"},
	},
	Examples: []string{"bonsai pack outdated", "bonsai pack outdated --json"},
	Refused: func(c *call, e *engine.Error) encoder {
		return &format.Packs{Command: "outdated", Packs: []format.PacksPack{}, Error: e.Object()}
	},
	Run: runPackOutdated,
}

// packTags reads a source's tags for pack outdated (a variable, so tests never reach the network).
var packTags = func(source string) ([]string, error) { return engine.RemoteTags(source, engine.PackTagsTimeout) }

// packConfig reads the checkout's bonsai.yaml bytes, from the folder the command runs in.
func packConfig() (root string, raw []byte, e *engine.Error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", nil, &engine.Error{Code: "read-failed", Exit: engine.ExitRuntime, What: "the current folder cannot be read",
			Next: "run it from a folder inside the project"}
	}
	co, err := workspace.Find(dir)
	if err != nil {
		return "", nil, engine.FindError(err)
	}
	raw, err = os.ReadFile(filepath.Join(co.Root, workspace.ConfigFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil, &engine.Error{Code: "not-linked", Exit: engine.ExitState, Who: "person",
				What: "this checkout has no bonsai.yaml, so it is not linked to Bonsai",
				Next: "link the project first: bonsai init --name <name> --source <pack repository> --ref <tag or commit>"}
		}
		return "", nil, &engine.Error{Code: "read-failed", Exit: engine.ExitRuntime, What: "bonsai.yaml cannot be read: " + engine.ASCII(err.Error()),
			Next: "check the file's permissions, then run the command again"}
	}
	return co.Root, raw, nil
}

// editError is a bonsai.yaml edit refused (nothing written).
func editError(err error) *engine.Error {
	var e *engine.Error
	if errors.As(err, &e) {
		return e
	}
	var we *workspace.Error
	if errors.As(err, &we) {
		return engine.ConfigError(err)
	}
	return &engine.Error{Code: "bad-value", Exit: exitInput, What: engine.ASCII(err.Error()) + ": nothing was written",
		Next: "run bonsai pack outdated to see the packs bonsai.yaml names, then run the command again"}
}

func runPackSet(c *call) int {
	if len(c.rest) != 1 {
		return c.refuse(c.flagError("pack set takes one <id>@<ref>, such as base@base-v1.0.1"))
	}
	id, ref, ok := strings.Cut(c.rest[0], "@")
	if !ok || id == "" || ref == "" {
		return c.refuse(c.flagError("pack set takes <id>@<ref>, such as base@base-v1.0.1, not %+q", c.rest[0]))
	}
	_, raw, e := packConfig()
	if e != nil {
		return c.fail(e)
	}
	edit, err := engine.SetRef(raw, id, ref)
	if err != nil {
		return c.fail(editError(err))
	}
	return runEngineAs(c, &packRun{edit: &edit, sub: "set", args: c.rest})
}

func runPackAdd(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("pack add takes no %+q: it takes --source, --ref and --path", c.rest[0]))
	}
	source, ref, p := c.value("--source"), c.value("--ref"), c.value("--path")
	if source == "" || ref == "" {
		return c.refuse(&engine.Error{Code: "missing-value", Exit: exitInput, What: "pack add needs --source and --ref",
			Next: "run bonsai pack add --source <pack repository> --ref <tag or commit>"})
	}
	_, raw, e := packConfig()
	if e != nil {
		return c.fail(e)
	}
	home, err := workspace.Home()
	if err != nil {
		return c.fail(engine.HomeError(err))
	}
	id, err := engine.PackIDAt(home, source, p, ref)
	if err != nil {
		return c.fail(engine.Unexpected(err))
	}
	edit, err := engine.AddPack(raw, engine.NewPack{ID: id, Source: source, Path: p, Ref: ref})
	if err != nil {
		return c.fail(editError(err))
	}
	return runEngineAs(c, &packRun{edit: &edit, sub: "add"})
}

func runPackRemove(c *call) int {
	if len(c.rest) != 1 {
		return c.refuse(c.flagError("pack remove takes one pack id"))
	}
	_, raw, e := packConfig()
	if e != nil {
		return c.fail(e)
	}
	edit, err := engine.RemovePack(raw, c.rest[0])
	if err != nil {
		return c.fail(editError(err))
	}
	return runEngineAs(c, &packRun{edit: &edit, sub: "remove", args: c.rest})
}

func runPackOutdated(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("pack outdated takes no %+q", c.rest[0]))
	}
	root, raw, e := packConfig()
	if e != nil {
		return c.fail(e)
	}
	cfg, err := workspace.ReadConfigFull(raw)
	if err != nil {
		return c.fail(engine.ConfigError(err))
	}
	home, err := workspace.Home()
	if err != nil {
		return c.fail(engine.HomeError(err))
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(workspace.LockFile))); err != nil {
		return c.fail(&engine.Error{Code: "no-lock", Exit: engine.ExitState, Who: "person",
			What: "bonsai.yaml is here but .bonsai/lock.json is not, so no pack is locked yet",
			Next: "run bonsai update to lock the packs bonsai.yaml names"})
	}
	lock, err := workspace.LoadLock(root)
	if err != nil {
		return c.fail(engine.Unexpected(err))
	}
	// Every source once, in the lock's order, then the cache written (a write that fails is said, never a failure).
	var reads []engine.PackTagRead
	why := map[string]string{}
	for _, lp := range lock.Packs {
		seen := false
		for _, r := range reads {
			seen = seen || r.Source == lp.Source
		}
		if seen {
			continue
		}
		tags, err := packTags(lp.Source)
		reads = append(reads, engine.PackTagRead{Source: lp.Source, Tags: tags, Err: err})
		if err != nil {
			why[lp.Source] = engine.ASCII(err.Error())
		}
	}
	cache, werr := engine.RecordPackTags(home, reads)
	doc := &format.Packs{Command: "outdated", Packs: []format.PacksPack{}}
	for _, lp := range lock.Packs {
		doc.Packs = append(doc.Packs, packsPack(home, cfg, lp, cache, why[lp.Source]))
	}
	if c.json {
		if werr != nil {
			_, _ = fmt.Fprintln(c.stderr, "the packs' tags were not kept in "+engine.PackCachePath(home)+" ("+engine.ASCII(werr.Error())+
				"): plain bonsai status will not show them")
		}
		return c.printDoc(doc, exitOK)
	}
	d := &textui.Doc{}
	d.Title("bonsai pack outdated", fmt.Sprintf("%d %s", len(doc.Packs), map[bool]string{true: "pack", false: "packs"}[len(doc.Packs) == 1]))
	var rows []textui.Row
	behind := 0
	for _, p := range doc.Packs {
		ref := "(not in bonsai.yaml)"
		if p.Ref != nil {
			ref = *p.Ref
		}
		state, tone := "newer releases unknown: "+engine.ASCII(strOr(p.Why, "not read")), textui.Warn
		if p.Newer != nil {
			switch p.Newer.State {
			case engine.PackNewer:
				behind++
				state, tone = p.Newer.Tag+" is released: bonsai pack set "+p.ID+"@"+p.Newer.Tag, textui.Warn
			case engine.PackCurrent:
				state, tone = "no newer release", textui.Ok
			default:
				state, tone = "stale: the newest release read is below the locked version, or the read is old", textui.Dim
			}
			if p.Why != nil {
				state += " (this run's read failed: " + *p.Why + "; the cache's last read, " + p.Newer.Read + ")"
			}
		}
		rows = append(rows, textui.R(textui.T(p.ID, textui.Key), textui.C(engine.ASCII(p.Version)), textui.T(engine.ASCII(ref), textui.Path),
			textui.T(engine.ASCII(state), tone)))
	}
	if len(rows) == 0 {
		rows = []textui.Row{textui.R(textui.T("none locked", textui.Dim))}
	}
	d.Section("Packs", rows...)
	if werr != nil {
		d.Para("note", "the packs' tags were not kept in "+engine.PackCachePath(home)+" ("+engine.ASCII(werr.Error())+
			"): plain bonsai status will not show them", textui.Warn)
	}
	if behind > 0 {
		d.Hint("each move is one preview first; --yes writes it (an agent's needs a task granting bonsai.yaml and .bonsai/lock.json)")
	}
	return write(c.stdout, d.Render(textui.For(c.stdout)))
}

func strOr(s *string, or string) string {
	if s == nil {
		return or
	}
	return *s
}

// packsPack is one locked pack as bonsai.packs/1 holds it: the lock's fields and bonsai.yaml's ref, the skills from
// the home's clone at the locked commit (null when it is not there), the cache's newer record and this run's read
// failure.
func packsPack(home string, cfg *workspace.Config, lp workspace.LockedPack, cache engine.PackCache, why string) format.PacksPack {
	p := format.PacksPack{ID: lp.ID, Source: lp.Source, Commit: lp.Commit, Version: lp.Version, Roles: []string{}, Deny: []string{},
		Protected: []string{}, Kinds: []string{}}
	if lp.Path != "" {
		path := lp.Path
		p.Path = &path
	}
	ref := ""
	for _, r := range cfg.Packs {
		if r.ID == lp.ID {
			ref = r.Ref
			p.Ref = &ref
		}
	}
	d, err := format.ReadDeclares(lp.Declares)
	if err == nil && d != nil {
		for _, r := range d.Roles {
			p.Roles = append(p.Roles, r.Name)
		}
		for _, r := range d.Deny {
			p.Deny = append(p.Deny, r.Rule)
		}
		p.Protected = append(p.Protected, d.Protected...)
		for _, k := range d.Documents {
			p.Kinds = append(p.Kinds, k.Kind)
		}
		if d.Needs != nil && d.Needs.ClaudeCode != nil {
			cc := *d.Needs.ClaudeCode
			p.Needs = &format.PacksNeeds{ClaudeCode: &cc}
		}
	}
	if skills, ok := engine.PackSkills(home, lp.Source, lp.Path, lp.Commit); ok {
		p.Skills = skills
	}
	if n := engine.PackNewerOf(cache, lp.Source, lp.ID, ref, lp.Path, lp.Version); n != nil && ref != "" {
		p.Newer = &format.StatusNewer{Tag: n.Tag, Read: n.Read, State: n.State}
	}
	if why != "" {
		p.Why = &why
	}
	return p
}
