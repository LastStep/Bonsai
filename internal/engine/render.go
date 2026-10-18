package engine

// What init, update and check print: plain ASCII text for a person (spec §3), and for programs (--json) the
// documents of formats set 4: init and update the changes output (bonsai.changes/1), check its own (bonsai.check/1),
// each with the error object (bonsai.error) when the command refused or failed (step 5.1.4b). They are written by
// internal/format's writers, which hold each document to its schema before a byte is printed.
//
// The preview names every file with its result and every settings line the plan adds, changes or removes, each
// with its sentence (spec §6, "The preview names every settings line"); the JSON carries the same, one entry per
// settings line (file, change, kind, line, was, why, runs_code).

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/textui"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// resultTone is what a file's or a plugin's result means for the reader.
func resultTone(result string) textui.Tone {
	switch result {
	case Created, "added", "installed", "uninstalled", "written":
		return textui.New
	case Unchanged, "skipped":
		return textui.Dim
	case Conflict, "failed":
		return textui.Bad
	case Changed, Removed, Released, Dropped, LeftMissing, "waiting":
		return textui.Warn
	}
	return textui.Plain
}

// flat is text with its line breaks and runs of spaces made single spaces, for a paragraph the Doc wraps itself.
func flat(s string) string { return strings.Join(strings.Fields(s), " ") }

// PreviewTo adds the plan to d: what the command would write, and why. explain lists every settings line with
// where it comes from and why (unlink's preview always does); else they are counted by change and kind.
func (p *Plan) PreviewTo(d *textui.Doc, diff, explain bool) {
	if len(p.Packs) > 0 {
		var rows []textui.Row
		for _, m := range p.Packs {
			move, tone := "at "+short(m.To), textui.Plain
			switch {
			case m.To == "":
				move, tone = short(m.From)+" -> taken out", textui.Warn
			case m.From == "":
				move, tone = "new, at "+short(m.To), textui.New
			case m.From != m.To:
				move, tone = short(m.From)+" -> "+short(m.To), textui.New
			}
			rows = append(rows, textui.R(textui.T(m.ID, textui.Key), textui.C(ascii(m.Version)), textui.T(move, tone), textui.T(ascii(m.Source), textui.Path)))
		}
		d.Section("Packs", rows...)
	}
	p.takenOutTo(d)
	var files []textui.Row
	for _, f := range p.Files {
		kind := ""
		if f.Kind != "" {
			kind = "[" + f.Kind + "]"
		}
		files = append(files, textui.R(textui.T(f.Result, resultTone(f.Result)), textui.T(ascii(f.Path), textui.Path), textui.T(kind, textui.Dim),
			textui.C(ascii(f.Why))))
	}
	switch {
	case p.LockRemove:
		files = append(files, textui.R(textui.T(Removed, resultTone(Removed)), textui.T(workspace.LockFile, textui.Path), textui.C(""), textui.C("removed last")))
	case p.LockWrite:
		r := map[bool]string{true: Created, false: Updated}[p.FirstLink]
		files = append(files, textui.R(textui.T(r, resultTone(r)), textui.T(workspace.LockFile, textui.Path), textui.C(""), textui.C("written last")))
	default:
		files = append(files, textui.R(textui.T(Unchanged, textui.Dim), textui.T(workspace.LockFile, textui.Path)))
	}
	files = append(files, p.format0Rows()...)
	if p.EmptyLocal > 0 {
		files = append(files, textui.R(textui.T("emptied", textui.Warn), textui.T(LocalDir+"/", textui.Path), textui.C(""),
			textui.C(fmt.Sprintf("its %d files removed (init --new-id: a copy keeps none of the original's log, asks and ladder results)", p.EmptyLocal))))
	}
	d.Section("Files", files...)
	if len(p.Settings) > 0 {
		p.settingsTo(d, explain || p.Command == "unlink")
	}
	if p.Command == "unlink" {
		p.LeftTo(d)
	} else {
		p.RunsCodeTo(d)
	}
	if diff {
		for _, f := range p.Files {
			if f.Writes() {
				var nb []byte
				if !f.remove {
					nb = f.write
				}
				d.Raw(Diff(f.Path, f.old, nb))
			}
		}
	}
	if len(p.Conflicts) > 0 {
		d.Para("Conflicts", fmt.Sprintf("%d. Nothing can be written until each is settled:", len(p.Conflicts)), textui.Bad)
		var rows []textui.Row
		for _, f := range p.Conflicts {
			rows = append(rows, textui.R(textui.T(ascii(f.Path), textui.Path), textui.C(ascii(f.Why))))
		}
		d.Section("", rows...)
	}
}

// Preview is the plan as text, every settings line listed: what the command would write, and why.
func (p *Plan) Preview(diff bool) string {
	var d textui.Doc
	p.PreviewTo(&d, diff, true)
	return d.String()
}

// format0Rows is the Files row for the format-0 files the lock lists; none when there are none.
func (p *Plan) format0Rows() []textui.Row {
	if p.Format0 <= 0 {
		return nil
	}
	return []textui.Row{textui.R(textui.C("listed"), textui.T(fmt.Sprintf("%d %s", p.Format0,
		plural(p.Format0, "task, run report or STATE file", "task, run report and STATE files")), textui.Path), textui.C(""),
		textui.C("with no format: line (format 0), each fixed in the lock by its hash: an agent gives one format: bonsai.<kind>/1 "+
			"before changing it (bonsai check finds a format-0 file new or changed)"))}
}

// settingsTo adds the settings lines: with explain each line, where it comes from and why (and the line it
// replaces); else one row per change and kind, with how many lines each origin gives, then each line changed or taken
// out in full.
func (p *Plan) settingsTo(d *textui.Doc, explain bool) {
	n := len(p.Settings)
	if explain {
		d.Para("Settings", fmt.Sprintf("%s: %d %s, each with where it comes from and why", SettingsFile, n, plural(n, "line", "lines")), textui.Plain)
		var rows []textui.Row
		for _, c := range p.Settings {
			rows = append(rows, textui.R(textui.T(c.Change, changeTone(c.Change)), textui.C(c.Kind), textui.T(origin(c.Origin), textui.Dim),
				textui.T(ascii(c.Line), textui.Path)))
			if c.Was != "" {
				rows = append(rows, textui.R(textui.C(""), textui.C(""), textui.C(""), textui.T("was: "+ascii(c.Was), textui.Dim)))
			}
			rows = append(rows, textui.R(textui.C(""), textui.C(""), textui.C(""), textui.C(ascii(c.Why))))
		}
		d.Table("", textui.R(textui.C("change"), textui.C("kind"), textui.C("from"), textui.C("line, and why")), rows...)
		return
	}
	d.Para("Settings", fmt.Sprintf("%s: %d %s; bonsai %s --explain lists each with where it comes from and why", SettingsFile, n,
		plural(n, "line", "lines"), p.Command), textui.Plain)
	type group struct {
		change, kind string
		n            int
		origins      []string
		count        map[string]int
	}
	var groups []*group
	byKey := map[string]*group{}
	for _, c := range p.Settings {
		g := byKey[c.Change+" "+c.Kind]
		if g == nil {
			g = &group{change: c.Change, kind: c.Kind, count: map[string]int{}}
			byKey[c.Change+" "+c.Kind] = g
			groups = append(groups, g)
		}
		o := origin(c.Origin)
		if g.count[o] == 0 {
			g.origins = append(g.origins, o)
		}
		g.count[o]++
		g.n++
	}
	var rows []textui.Row
	for _, g := range groups {
		from := make([]string, len(g.origins))
		for i, o := range g.origins {
			from[i] = fmt.Sprintf("%s %d", o, g.count[o])
		}
		rows = append(rows, textui.R(textui.T(g.change, changeTone(g.change)), textui.C(g.kind), textui.C(fmt.Sprint(g.n)),
			textui.T(ascii(strings.Join(from, ", ")), textui.Dim)))
	}
	d.Table("", textui.R(textui.C("change"), textui.C("kind"), textui.C("lines"), textui.C("from")), rows...)
	// A line changed or taken out is named whatever the count: few, and each one matters.
	rows = nil
	for _, c := range p.Settings {
		if c.Change == "add" {
			continue
		}
		rows = append(rows, textui.R(textui.T(c.Change, changeTone(c.Change)), textui.C(c.Kind), textui.T(ascii(c.Line), textui.Path)))
		if c.Was != "" {
			rows = append(rows, textui.R(textui.C(""), textui.C(""), textui.T("was: "+ascii(c.Was), textui.Dim)))
		}
		rows = append(rows, textui.R(textui.C(""), textui.C(""), textui.C(ascii(c.Why))))
	}
	if len(rows) > 0 {
		d.Para("", "Each line changed or taken out:", textui.Plain)
		d.Section("", rows...)
	}
}

// origin names where a settings line comes from: Bonsai, bonsai.yaml, a pack by its id, or the project's own file.
func origin(o string) string {
	switch o {
	case "":
		return "the project"
	case "bonsai":
		return "Bonsai"
	}
	return o
}

func changeTone(change string) textui.Tone {
	switch change {
	case "add":
		return textui.New
	case "change", "remove":
		return textui.Warn
	}
	return textui.Plain
}

// takenOutTo names, for each pack the plan takes out (Plan.Removed), everything it wrote and what becomes of it
// (step 5.1.7): its files (each listed under Files: removed when nobody edited it, else released and named here), its
// part of the block, its settings lines (listed under .claude/settings.json), its lock entry and declares, and its
// plugin's install record, removed after the write. Nothing when the plan takes none out.
func (p *Plan) takenOutTo(d *textui.Doc) {
	if len(p.Removed) == 0 || p.Command == "unlink" {
		return
	}
	var rows []textui.Row
	for _, lp := range p.Removed {
		var removed, left []string
		for _, f := range p.Files {
			if f.Pack != lp.ID {
				continue
			}
			switch f.Result {
			case Removed:
				removed = append(removed, f.Path)
			case Released:
				left = append(left, f.Path)
			}
		}
		rows = append(rows, textui.R(textui.T(ascii(lp.ID), textui.Warn), textui.C(fmt.Sprintf("Taken out of bonsai.yaml. Its files nobody edited go (%s); its part of Bonsai's block in %s goes; its hook lines, deny rules "+
			"and plugin wiring leave %s; its lock entry and declares go; once written, Claude Code's record of its plugin's install "+
			"for this checkout is removed (claude plugin uninstall, scope %s).", listOrNone(removed), BlockFile, SettingsFile, PluginScope))))
		if len(left) > 0 {
			rows = append(rows, textui.R(textui.C(""), textui.C("Left in place, the project's now (edited here, or written once): "+ascii(strings.Join(left, ", ")))))
		}
	}
	d.Section("Taken out", rows...)
}

func listOrNone(paths []string) string {
	if len(paths) == 0 {
		return "none"
	}
	return ascii(strings.Join(paths, ", "))
}

// RunsCodeTo adds the preview's "Runs code" part (consent.go): each item, and whether --allow-exec is needed; at a
// first link, Bonsai's own hook lines, which --yes writes.
func (p *Plan) RunsCodeTo(d *textui.Doc) {
	if len(p.OwnHooks) > 0 {
		d.Para("Own hooks", fmt.Sprintf("Bonsai's own hook %s, written with --yes (linking the project is your consent to %s):",
			plural(len(p.OwnHooks), "line", "lines"), plural(len(p.OwnHooks), "it", "them")), textui.Plain)
		var rows []textui.Row
		for _, c := range p.OwnHooks {
			rows = append(rows, textui.R(textui.T(c.Change, changeTone(c.Change)), textui.C(c.Kind), textui.T(ascii(c.Line), textui.Path)))
		}
		d.Section("", rows...)
	}
	for _, id := range p.Unverified {
		d.Para("Unverified", fmt.Sprintf("the pack %s's folder in bonsai.yaml is not the lock's, or the lock's record of it does not match "+
			"its locked commit (a lock edited by hand), so its hook lines and plugin count as at a first link.", ascii(id)), textui.Warn)
	}
	if len(p.LeftHooks) > 0 {
		why := "the lock is missing"
		if !p.FirstLink {
			why = "the lock does not record them as its own"
		}
		d.Para("Left hooks", "Left in place as the project's own ("+why+", so Bonsai cannot tell them from its own; "+
			"take out by hand any you do not want):", textui.Warn)
		var rows []textui.Row
		for _, l := range p.LeftHooks {
			rows = append(rows, textui.R(textui.C("keep"), textui.C("hook"), textui.T(ascii(l), textui.Path)))
		}
		d.Section("", rows...)
	}
	n := len(p.RunsCode)
	switch {
	case n == 0:
		d.Para("Runs code", "nothing that needs --allow-exec (no pack hook line, file a hook runs or plugin code part is "+
			"added or changed, and no change to a hook line)", textui.Ok)
		return
	case p.AllowExec:
		d.Para("Runs code", fmt.Sprintf("%d %s, consented to with --allow-exec (spec section 6):", n, plural(n, "item", "items")), textui.Warn)
	default:
		d.Para("Runs code", fmt.Sprintf("%d %s, which %s --allow-exec as well as --yes (spec section 6):", n,
			plural(n, "item", "items"), plural(n, "needs", "need")), textui.Warn)
	}
	var rows []textui.Row
	for _, c := range p.RunsCode {
		rows = append(rows, textui.R(textui.T(c.Change, changeTone(c.Change)), textui.C(c.Kind), textui.T(ascii(c.Item)+"  ("+ascii(c.Pack)+")", textui.Path)))
		if c.Was != "" {
			rows = append(rows, textui.R(textui.C(""), textui.C(""), textui.T("was: "+ascii(c.Was), textui.Dim)))
		}
		rows = append(rows, textui.R(textui.C(""), textui.C(""), textui.C(ascii(c.Why))))
	}
	d.Section("", rows...)
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// ConflictNext is the next step for a plan's conflicts: the command again with --keep for each (a file Bonsai
// owns part of takes --adopt), and the same with --adopt.
func (p *Plan) ConflictNext(cmdline string) string {
	var keep, adopt []string
	for _, f := range p.Conflicts {
		path := ShellArg(f.Path)
		if f.Kind == "block" || f.Kind == "keys" {
			keep = append(keep, "--adopt "+path)
		} else {
			keep = append(keep, "--keep "+path)
		}
		adopt = append(adopt, "--adopt "+path)
	}
	k := cmdline + " --yes " + strings.Join(keep, " ")
	a := cmdline + " --yes " + strings.Join(adopt, " ")
	if k == a {
		return "to take Bonsai's copy (yours is saved in the Bonsai home's cache, never in the repo), run: " + a
	}
	return "to keep your edits, run: " + k + "\n  or, to take the pack's copies (yours are saved in the Bonsai home's cache, never in the repo), run: " + a
}

// ShellArg quotes an argument for PowerShell and sh alike when it holds a space or a quote character.
func ShellArg(s string) string {
	if s != "" && !strings.ContainsAny(s, " \t'\"$`;&|<>()") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// AppliedTo adds what was written: each file that changed, and where each adopted copy went. Nothing when nothing
// was.
func (p *Plan) AppliedTo(d *textui.Doc) {
	var rows []textui.Row
	for _, f := range p.Files {
		if f.Result == Unchanged {
			continue
		}
		note := ""
		if f.Kind == "keys" && len(p.Settings) > 0 {
			note = fmt.Sprintf("%d %s", len(p.Settings), plural(len(p.Settings), "line", "lines"))
		}
		if f.Saved != "" {
			note = strings.TrimPrefix(note+"; your copy: "+ascii(f.Saved), "; ")
		}
		rows = append(rows, textui.R(textui.T(f.Result, resultTone(f.Result)), textui.T(ascii(f.Path), textui.Path), textui.C(note)))
	}
	if p.LockWrite {
		rows = append(rows, textui.R(textui.T("written", textui.New), textui.T(workspace.LockFile, textui.Path)))
	}
	if p.LockRemove {
		rows = append(rows, textui.R(textui.T(Removed, resultTone(Removed)), textui.T(workspace.LockFile, textui.Path)))
	}
	rows = append(rows, p.format0Rows()...)
	if p.EmptyLocal > 0 {
		rows = append(rows, textui.R(textui.T("emptied", textui.Warn), textui.T(LocalDir+"/", textui.Path), textui.C(fmt.Sprintf("%d files", p.EmptyLocal))))
	}
	if len(rows) > 0 {
		d.Section("Written", rows...)
	}
}

// LeftAloneTo adds what a plan leaves as it is that bonsai check reports as a finding: each file edited here that the
// pack did not change (spec §6's table: "left alone; check reports it changed"), with how to settle it, so update's
// "nothing to change" and check's findings agree (step 5.1.1's verifier: a forged lock left update saying "nothing to
// change" beside check's findings). Nothing when there is none.
func (p *Plan) LeftAloneTo(d *textui.Doc) {
	var rows []textui.Row
	for _, f := range p.Files {
		if f.Result != Changed {
			continue
		}
		path := ShellArg(f.Path)
		next := "to take Bonsai's copy back (yours is saved in the Bonsai home), run: bonsai " + p.Command + " --yes --adopt " + path
		if f.Kind == "pack" {
			next = "to keep the edit, run: bonsai " + p.Command + " --yes --keep " + path + "; to take the pack's copy back, run: bonsai " +
				p.Command + " --yes --adopt " + path
		}
		rows = append(rows, textui.R(textui.T(ascii(f.Path), textui.Path), textui.T("["+f.Kind+"]", textui.Dim), textui.C(ascii(f.Why))),
			textui.R(textui.C(""), textui.C(""), textui.T("next: "+ascii(next), textui.Path)))
	}
	if len(rows) == 0 {
		return
	}
	d.Para("Edited here", "Left as they are, edited here (bonsai check reports each as changed until it is settled):", textui.Warn)
	d.Section("", rows...)
}

// SessionFolderWords say where a session must start for Bonsai's hook lines to run (design/plan-5.md, 5.3.1 note 1):
// Claude Code reads a project's .claude/settings.json from the folder a session starts in, and reads it again from a
// folder a person's /cd moves the session to, so a session started below the project's top folder, or moved out of
// it, runs none of them, the guard's among them. Measured 11 Oct with a scratch project's own hooks: on Claude Code
// 2.1.296 in WSL, a session started in a subfolder of the checkout ran none, and in a session started at the top none
// ran after a /cd into a subfolder until a /cd back; on 2.1.294 on Windows, the subfolder start alike. bonsai hook
// guard --help says the same.
const SessionFolderWords = "Start Claude Code in the project's top folder, and do not /cd out of it: a session started or\n" +
	"moved elsewhere loads none of the project's Claude Code settings, so no Bonsai hook guards it.\n"

// ClosingTo adds what init ends with and status shows (spec §6, "Said in plain words").
func ClosingTo(d *textui.Doc, cfg *workspace.Config, home string) {
	d.Section("Linked", textui.R(textui.T(ascii(cfg.Name), textui.Key), textui.T(ascii(cfg.ID), textui.Path),
		textui.T("every clone and worktree of this project shares this workspace id", textui.Dim)))
	WhereTo(d, "")
	d.Section("On this machine", textui.R(textui.T(ascii(filepath.ToSlash(home)), textui.Path), textui.C("Bonsai's home: a secret salt, this "+
		"machine's settings for each project, label files attached on this machine, your personal memory, the pack cache. None of it enters git.")))
	d.Para("Sessions", flat(SessionFolderWords), textui.Plain)
	d.Para("Hook lines", flat(HookPlaceWords()), textui.Plain)
	d.Hint("a copy meant as a new project needs its own id: bonsai init --new-id")
}

// WhereTo adds the project's Bonsai files and folders and what each holds, under the label "In this project" (with
// the checkout's path when root is not "").
func WhereTo(d *textui.Doc, root string) {
	var rows []textui.Row
	if root != "" {
		rows = append(rows, textui.R(textui.T(ascii(root), textui.Path)))
	}
	rows = append(rows,
		textui.R(textui.T("bonsai.yaml", textui.Path), textui.C("this project's Bonsai settings. You edit it; it is committed.")),
		textui.R(textui.T(".bonsai/", textui.Path), textui.C("the lock, STATE and two tables Bonsai rebuilds (tasks; sessions and hours). Committed.")),
		textui.R(textui.T(".bonsai/local/", textui.Path), textui.C("the log, questions for you and their answers, ladder results, messages between agents. Never committed.")))
	d.Section("In this project", rows...)
}

// WorkspaceRef is the plan's workspace as the changes output names it: bonsai.yaml's id and name (init --new-id:
// the new id) and the checkout's absolute path, forward slashes; nil before bonsai.yaml was read.
func (p *Plan) WorkspaceRef() *format.WorkspaceRef {
	if p.Config == nil {
		return nil
	}
	return &format.WorkspaceRef{ID: p.Config.ID, Name: p.Config.Name, Root: filepath.ToSlash(p.Root)}
}

// Changes is the plan as init's and update's --json prints it (bonsai.changes/1, written by format.Changes.Encode):
// result is preview, applied, nothing, conflict, refused or failed, and refusal what stopped it (nil for none). A
// preview, a conflict, a refusal or a failure gives the whole plan, though nothing (or, failed, not all) was written.
// A settings line's runs_code is true only for a line among runs_code's items: Bonsai's own hook lines at a first
// link are own_hooks, which --yes writes.
func (p *Plan) Changes(result string, refusal *Error) *format.Changes {
	c := &format.Changes{Command: p.Command, Result: result, Workspace: p.WorkspaceRef(), AllowExec: p.AllowExec,
		LeftHooks: p.LeftHooks, Unverified: p.Unverified}
	for _, m := range p.Packs {
		c.Packs = append(c.Packs, format.ChangesPack{ID: m.ID, Source: m.Source, Version: m.Version, From: orNull(m.From), To: orNull(m.To)})
	}
	for _, f := range p.Files {
		c.Files = append(c.Files, format.ChangesFile{Path: f.Path, Kind: orNull(f.Kind), Pack: orNull(f.Pack), Result: f.Result,
			Why: orNull(f.Why), Saved: orNull(f.Saved)})
	}
	lock := map[bool]string{true: "written", false: "unchanged"}[p.LockWrite]
	if p.LockRemove {
		lock = LockRemoved
	}
	c.Lock = &lock
	for _, s := range p.Settings {
		c.Settings = append(c.Settings, format.ChangesSetting{File: SettingsFile, Change: s.Change, Kind: s.Kind, Line: s.Line,
			Was: orNull(s.Was), Why: s.Why, RunsCode: s.RunsCode && !p.ownHook(s)})
	}
	for _, it := range p.RunsCode {
		c.RunsCode = append(c.RunsCode, format.ChangesCode{Kind: it.Kind, Change: it.Change, Pack: it.Pack, Item: it.Item,
			Was: orNull(it.Was), Why: it.Why})
	}
	for _, o := range p.OwnHooks {
		c.OwnHooks = append(c.OwnHooks, o.Line)
	}
	for _, f := range p.Conflicts {
		c.Conflicts = append(c.Conflicts, f.Path)
	}
	for _, r := range p.Plugins {
		var next *format.Next
		if r.Next != "" {
			next = &format.Next{Do: r.Next, Who: whoOr(r.Who)}
		}
		c.Plugins = append(c.Plugins, format.ChangesPlugin{Pack: r.Pack, Plugin: r.Plugin, Commit: r.Commit, Result: r.Result,
			Message: r.Message, Next: next})
	}
	for _, l := range p.Left {
		c.Left = append(c.Left, ascii(l))
	}
	if refusal != nil {
		c.Error = refusal.Object()
	}
	return c
}

// ownHook reports whether a settings change is one of Bonsai's own hook lines added at a first link (OwnHooks),
// which --yes writes.
func (p *Plan) ownHook(s SettingsChange) bool {
	return p.FirstLink && s.Own && s.Change == "add"
}

// RefusedChanges is the changes output of init or update stopped before a plan: result refused, every field after
// it null or [] but the workspace (once bonsai.yaml was read), allow_exec and the error (bonsai.changes/1).
func RefusedChanges(command string, allowExec bool, e *Error) *format.Changes {
	return &format.Changes{Command: command, Result: "refused", Workspace: e.Workspace, AllowExec: allowExec, Error: e.Object()}
}

func orNull(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// whoOr gives who, or agent when none was named.
func whoOr(who string) string {
	if who == "person" {
		return who
	}
	return "agent"
}

// PluginsTo adds what InstallPlugins did, for a person: one row per pack's plugin, and its next step when this
// machine still needs one, an empty line between one plugin's rows and the next's (so one plugin's last words never
// run on into the next plugin's result). Nothing when nothing was asked of Claude Code.
func (p *Plan) PluginsTo(d *textui.Doc) {
	if len(p.Plugins) == 0 {
		return
	}
	d.Para("Plugins", fmt.Sprintf("This machine's plugins (Claude Code, scope %s: this checkout's %s):", PluginScope, SettingsFile), textui.Plain)
	var rows []textui.Row
	for i, r := range p.Plugins {
		if i > 0 {
			rows = append(rows, textui.R())
		}
		rows = append(rows, textui.R(textui.T(r.Result, resultTone(r.Result)), textui.T(ascii(r.Plugin), textui.Path), textui.C(ascii(r.Message))))
		if r.Next != "" {
			rows = append(rows, textui.R(textui.C(""), textui.C(""), textui.T("next: "+ascii(r.Next), textui.Path)))
		}
	}
	d.Section("", rows...)
}

// Doc is check's result as check --json prints it (bonsai.check/1, written by format.Check.Encode): every finding
// and warning with its next step and who takes it, and the notes check says to a person; error null.
func (r *CheckResult) Doc() *format.Check {
	list := func(fs []Finding) []format.Finding {
		out := []format.Finding{}
		for _, f := range fs {
			out = append(out, format.Finding{Code: f.Code, File: orNull(f.File), Message: ascii(strings.TrimSuffix(f.Message, ".")),
				Next: format.Next{Do: ascii(f.Next), Who: whoOr(f.Who)}})
		}
		return out
	}
	notes := []string{}
	for _, n := range r.Notes {
		notes = append(notes, ascii(n))
	}
	return &format.Check{Findings: list(r.Findings), Warnings: list(r.Warnings), Notes: notes}
}

// CheckRefused is check --json when check could not run: no findings, no warnings, and the error.
func CheckRefused(e *Error) *format.Check { return &format.Check{Error: e.Object()} }

// TextTo adds check's result for a person: the count first, then each finding and warning with its next step, and
// the notes.
func (r *CheckResult) TextTo(d *textui.Doc) {
	if len(r.Findings) == 0 {
		d.Title("bonsai check", "no findings")
	} else {
		d.Title("bonsai check", fmt.Sprintf("%d %s", len(r.Findings), plural(len(r.Findings), "finding", "findings")))
	}
	add := func(label string, fs []Finding, tone textui.Tone) {
		var rows []textui.Row
		for _, f := range fs {
			rows = append(rows, textui.R(textui.T(ascii(f.Message), tone)), textui.R(textui.T("  next: "+ascii(f.Next), textui.Path)))
		}
		if len(rows) > 0 {
			d.Section(label, rows...)
		}
	}
	add("Findings", r.Findings, textui.Bad)
	add("Warnings", r.Warnings, textui.Warn)
	if len(r.Notes) > 0 {
		var rows []textui.Row
		for _, n := range r.Notes {
			rows = append(rows, textui.R(textui.C(ascii(n))))
		}
		d.Section("Notes", rows...)
	}
}

// Text is check's result for a person, drawn for a pipe.
func (r *CheckResult) Text() string {
	var d textui.Doc
	r.TextTo(&d)
	return d.String()
}
