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
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Preview is the plan as text: what the command would write, and why.
func (p *Plan) Preview(diff bool) string {
	var b strings.Builder
	if len(p.Packs) > 0 {
		b.WriteString("Packs:\n")
		for _, m := range p.Packs {
			move := "at " + short(m.To)
			switch {
			case m.From == "":
				move = "new, at " + short(m.To)
			case m.From != m.To:
				move = short(m.From) + " -> " + short(m.To)
			}
			fmt.Fprintf(&b, "  %s %s  %s  (%s)\n", m.ID, ascii(m.Version), move, ascii(m.Source))
		}
	}
	b.WriteString("Files:\n")
	for _, f := range p.Files {
		kind := ""
		if f.Kind != "" {
			kind = " [" + f.Kind + "]"
		}
		why := ""
		if f.Why != "" {
			why = ": " + f.Why
		}
		fmt.Fprintf(&b, "  %-12s %s%s%s\n", f.Result, ascii(f.Path), kind, ascii(why))
	}
	if p.LockWrite {
		fmt.Fprintf(&b, "  %-12s %s: written last\n", map[bool]string{true: Created, false: Updated}[p.FirstLink], workspace.LockFile)
	} else {
		fmt.Fprintf(&b, "  %-12s %s\n", Unchanged, workspace.LockFile)
	}
	if p.Format0 > 0 {
		fmt.Fprintf(&b, "  %-12s %d %s with no format: line (format 0), each fixed in the lock by its hash: an agent gives one\n"+
			"               format: bonsai.<kind>/1 before changing it (bonsai check finds a format-0 file new or changed)\n",
			"listed", p.Format0, plural(p.Format0, "task, run report or STATE file", "task, run report and STATE files"))
	}
	if p.EmptyLocal > 0 {
		fmt.Fprintf(&b, "  %-12s %s/: its %d files removed (init --new-id: a copy keeps none of the original's log, asks and ladder results)\n",
			"emptied", LocalDir, p.EmptyLocal)
	}
	if len(p.Settings) > 0 {
		fmt.Fprintf(&b, "%s: %d %s\n", SettingsFile, len(p.Settings), plural(len(p.Settings), "line", "lines"))
		for _, c := range p.Settings {
			fmt.Fprintf(&b, "  %-7s %-11s %s\n", c.Change, c.Kind, ascii(c.Line))
			if c.Was != "" {
				fmt.Fprintf(&b, "          was: %s\n", ascii(c.Was))
			}
			fmt.Fprintf(&b, "          %s\n", ascii(c.Why))
		}
	}
	b.WriteString(p.RunsCodeText())
	if diff {
		for _, f := range p.Files {
			if f.Writes() {
				var nb []byte
				if !f.remove {
					nb = f.write
				}
				b.WriteString(Diff(f.Path, f.old, nb))
			}
		}
	}
	if len(p.Conflicts) > 0 {
		fmt.Fprintf(&b, "Conflicts: %d. Nothing can be written until each is settled:\n", len(p.Conflicts))
		for _, f := range p.Conflicts {
			fmt.Fprintf(&b, "  %s: %s\n", ascii(f.Path), ascii(f.Why))
		}
	}
	return b.String()
}

// RunsCodeText is the preview's "Runs code" part (consent.go): each item, and whether --allow-exec is needed; at a
// first link, Bonsai's own hook lines, which --yes writes.
func (p *Plan) RunsCodeText() string {
	var b strings.Builder
	if len(p.OwnHooks) > 0 {
		fmt.Fprintf(&b, "Bonsai's own hook %s, written with --yes (linking the project is your consent to %s):\n",
			plural(len(p.OwnHooks), "line", "lines"), plural(len(p.OwnHooks), "it", "them"))
		for _, c := range p.OwnHooks {
			fmt.Fprintf(&b, "  %-7s %-11s %s\n", c.Change, c.Kind, ascii(c.Line))
		}
	}
	for _, id := range p.Unverified {
		fmt.Fprintf(&b, "Unverified: the pack %s's folder in bonsai.yaml is not the lock's, or the lock's record of it does not match "+
			"its locked commit (a lock edited by hand), so its hook lines and plugin count as at a first link.\n", ascii(id))
	}
	if len(p.LeftHooks) > 0 {
		why := "the lock is missing"
		if !p.FirstLink {
			why = "the lock does not record them as its own"
		}
		b.WriteString("Left in place as the project's own (" + why + ", so Bonsai cannot tell them from its own; " +
			"take out by hand any you do not want):\n")
		for _, l := range p.LeftHooks {
			fmt.Fprintf(&b, "  %-7s %-11s %s\n", "keep", "hook", ascii(l))
		}
	}
	n := len(p.RunsCode)
	switch {
	case n == 0:
		b.WriteString("Runs code: nothing that needs --allow-exec (no pack hook line, file a hook runs or plugin code part is " +
			"added or changed, and no change to a hook line).\n")
		return b.String()
	case p.AllowExec:
		fmt.Fprintf(&b, "Runs code: %d %s, consented to with --allow-exec (spec section 6):\n", n, plural(n, "item", "items"))
	default:
		fmt.Fprintf(&b, "Runs code: %d %s, which %s --allow-exec as well as --yes (spec section 6):\n", n,
			plural(n, "item", "items"), plural(n, "needs", "need"))
	}
	for _, c := range p.RunsCode {
		fmt.Fprintf(&b, "  %-7s %-7s %s  (%s)\n", c.Change, c.Kind, ascii(c.Item), ascii(c.Pack))
		if c.Was != "" {
			fmt.Fprintf(&b, "          was: %s\n", ascii(c.Was))
		}
		fmt.Fprintf(&b, "          %s\n", ascii(c.Why))
	}
	return b.String()
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

// Applied is the text after a plan was written: each file that changed, and where each adopted copy went.
func (p *Plan) Applied() string {
	var b strings.Builder
	n := 0
	for _, f := range p.Files {
		if f.Result == Unchanged {
			continue
		}
		n++
		line := fmt.Sprintf("  %-12s %s", f.Result, ascii(f.Path))
		if f.Kind == "keys" && len(p.Settings) > 0 {
			line += fmt.Sprintf(" (%d %s)", len(p.Settings), plural(len(p.Settings), "line", "lines"))
		}
		if f.Saved != "" {
			line += "; your copy: " + ascii(f.Saved)
		}
		b.WriteString(line + "\n")
	}
	if p.LockWrite {
		fmt.Fprintf(&b, "  %-12s %s\n", "written", workspace.LockFile)
	}
	if p.Format0 > 0 {
		fmt.Fprintf(&b, "  %-12s %d %s with no format: line (format 0), each fixed in the lock by its hash: an agent gives one\n"+
			"               format: bonsai.<kind>/1 before changing it (bonsai check finds a format-0 file new or changed)\n",
			"listed", p.Format0, plural(p.Format0, "task, run report or STATE file", "task, run report and STATE files"))
	}
	if p.EmptyLocal > 0 {
		fmt.Fprintf(&b, "  %-12s %s/ (%d files)\n", "emptied", LocalDir, p.EmptyLocal)
	}
	if n == 0 && !p.LockWrite && p.EmptyLocal <= 0 {
		return ""
	}
	return b.String()
}

// LeftAlone is what a plan leaves as it is that bonsai check reports as a finding: each file edited here that the
// pack did not change (spec §6's table: "left alone; check reports it changed"), with how to settle it, so update's
// "nothing to change" and check's findings agree (step 5.1.1's verifier: a forged lock left update saying "nothing to
// change" beside check's findings). "" when there is none.
func (p *Plan) LeftAlone() string {
	var b strings.Builder
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
		fmt.Fprintf(&b, "  %s [%s]: %s\n    next: %s\n", ascii(f.Path), f.Kind, ascii(f.Why), ascii(next))
	}
	if b.Len() == 0 {
		return ""
	}
	return "Left as they are, edited here (bonsai check reports each as changed until it is settled):\n" + b.String()
}

// ClosingWords are what init ends with and status shows (spec §6, "Said in plain words").
func ClosingWords(cfg *workspace.Config, home string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Linked %s. Workspace id: %s (every clone and worktree of this project shares it).\n", ascii(cfg.Name), ascii(cfg.ID))
	b.WriteString("In this project:\n")
	b.WriteString("  bonsai.yaml      this project's Bonsai settings. You edit it; it is committed.\n")
	b.WriteString("  .bonsai/         the lock, STATE and two tables Bonsai rebuilds (tasks; sessions and hours). Committed.\n")
	b.WriteString("  .bonsai/local/   the log, questions for you and their answers, ladder results. Never committed.\n")
	b.WriteString("On this machine:\n")
	fmt.Fprintf(&b, "  %s\n", ascii(filepath.ToSlash(home)))
	b.WriteString("    Bonsai's home: a secret salt, this machine's settings for each project, label files attached\n")
	b.WriteString("    on this machine, your personal memory, the pack cache. None of it enters git.\n")
	b.WriteString("A copy meant as a new project needs its own id: bonsai init --new-id\n")
	return b.String()
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
		to := m.To
		c.Packs = append(c.Packs, format.ChangesPack{ID: m.ID, Source: m.Source, Version: m.Version, From: orNull(m.From), To: &to})
	}
	for _, f := range p.Files {
		c.Files = append(c.Files, format.ChangesFile{Path: f.Path, Kind: orNull(f.Kind), Pack: orNull(f.Pack), Result: f.Result,
			Why: orNull(f.Why), Saved: orNull(f.Saved)})
	}
	lock := map[bool]string{true: "written", false: "unchanged"}[p.LockWrite]
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

// PluginsText is what InstallPlugins did, for a person: one line per pack's plugin, and its next step when this
// machine still needs one. "" when nothing was asked of Claude Code.
func (p *Plan) PluginsText() string {
	if len(p.Plugins) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "This machine's plugins (Claude Code, scope %s: this checkout's %s):\n", PluginScope, SettingsFile)
	for _, r := range p.Plugins {
		fmt.Fprintf(&b, "  %-12s %s: %s\n", r.Result, ascii(r.Plugin), ascii(r.Message))
		if r.Next != "" {
			fmt.Fprintf(&b, "               next: %s\n", ascii(r.Next))
		}
	}
	return b.String()
}

// Doc is check's result as check --json prints it (bonsai.check/1, written by format.Check.Encode): every finding
// and warning with its next step and who takes it; error null.
func (r *CheckResult) Doc() *format.Check {
	list := func(fs []Finding) []format.Finding {
		out := []format.Finding{}
		for _, f := range fs {
			out = append(out, format.Finding{Code: f.Code, File: orNull(f.File), Message: ascii(strings.TrimSuffix(f.Message, ".")),
				Next: format.Next{Do: ascii(f.Next), Who: whoOr(f.Who)}})
		}
		return out
	}
	return &format.Check{Findings: list(r.Findings), Warnings: list(r.Warnings)}
}

// CheckRefused is check --json when check could not run: no findings, no warnings, and the error.
func CheckRefused(e *Error) *format.Check { return &format.Check{Error: e.Object()} }

// Text is check's result for a person.
func (r *CheckResult) Text() string {
	var b strings.Builder
	if len(r.Findings) == 0 {
		b.WriteString("bonsai check: no findings.\n")
	} else {
		fmt.Fprintf(&b, "bonsai check: %d %s.\n", len(r.Findings), plural(len(r.Findings), "finding", "findings"))
		for _, f := range r.Findings {
			fmt.Fprintf(&b, "  %s\n    next: %s\n", ascii(f.Message), ascii(f.Next))
		}
	}
	for _, w := range r.Warnings {
		fmt.Fprintf(&b, "warning: %s\n    next: %s\n", ascii(w.Message), ascii(w.Next))
	}
	for _, n := range r.Notes {
		fmt.Fprintf(&b, "note: %s\n", ascii(n))
	}
	return b.String()
}
