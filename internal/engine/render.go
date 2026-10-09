package engine

// What init, update and check print: plain ASCII text for a person (spec §3), and JSON for programs (--json).
//
// The preview names every file with its result and every settings line the plan adds, changes or removes, each
// with its sentence (spec §6, "The preview names every settings line"); the JSON carries the same, one entry per
// settings line (file, change, line, why). The JSON documents are not yet one of the contract's formats: step 5.1
// gives the commands' outputs and the error object (spec §16) their formats.

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
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
		fmt.Fprintf(&b, "Unverified: the lock's record of the pack %s does not match its locked commit read at bonsai.yaml's "+
			"folder (a folder changed in bonsai.yaml, or a lock edited by hand), so its hook lines and plugin count as at a first link.\n", ascii(id))
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
	if p.EmptyLocal > 0 {
		fmt.Fprintf(&b, "  %-12s %s/ (%d files)\n", "emptied", LocalDir, p.EmptyLocal)
	}
	if n == 0 && !p.LockWrite && p.EmptyLocal <= 0 {
		return ""
	}
	return b.String()
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

// JSON is the plan as a document for programs: result is preview, applied, nothing, conflict, refused or declined.
// runs_code lists what the plan writes that runs code, one entry per item (kind, change, pack, item, was, why), and
// allow_exec says whether --allow-exec was given: without it, a plan with any item is refused. left_hooks lists, at
// a link again with the lock missing, the hook lines left in place as the project's own.
func (p *Plan) JSON(result string, exit int, refusal *Error) schema.Object {
	packs := []any{}
	for _, m := range p.Packs {
		var from any
		if m.From != "" {
			from = m.From
		}
		packs = append(packs, objectOf("id", m.ID, "source", m.Source, "version", m.Version, "from", from, "to", m.To))
	}
	files := []any{}
	for _, f := range p.Files {
		var kind, pack, saved any
		if f.Kind != "" {
			kind = f.Kind
		}
		if f.Pack != "" {
			pack = f.Pack
		}
		if f.Saved != "" {
			saved = f.Saved
		}
		files = append(files, objectOf("path", f.Path, "kind", kind, "pack", pack, "result", f.Result, "why", f.Why, "saved", saved))
	}
	settings := []any{}
	for _, c := range p.Settings {
		var was any
		if c.Was != "" {
			was = c.Was
		}
		settings = append(settings, objectOf("file", SettingsFile, "change", c.Change, "kind", c.Kind, "line", c.Line,
			"was", was, "why", c.Why, "runs_code", c.RunsCode))
	}
	conflicts := []any{}
	for _, f := range p.Conflicts {
		conflicts = append(conflicts, f.Path)
	}
	unverified := []any{}
	for _, id := range p.Unverified {
		unverified = append(unverified, id)
	}
	leftHooks := []any{}
	for _, l := range p.LeftHooks {
		leftHooks = append(leftHooks, l)
	}
	runsCode := []any{}
	for _, c := range p.RunsCode {
		var was any
		if c.Was != "" {
			was = c.Was
		}
		runsCode = append(runsCode, objectOf("kind", c.Kind, "change", c.Change, "pack", c.Pack, "item", c.Item, "was", was, "why", c.Why))
	}
	var ws any
	if p.Config != nil {
		ws = objectOf("name", p.Config.Name, "id", p.Config.ID, "root", filepath.ToSlash(p.Root))
	}
	plugins := []any{}
	for _, r := range p.Plugins {
		var next any
		if r.Next != "" {
			next = r.Next
		}
		plugins = append(plugins, objectOf("pack", r.Pack, "plugin", r.Plugin, "commit", r.Commit, "result", r.Result,
			"message", r.Message, "next", next))
	}
	return objectOf("command", p.Command, "result", result, "exit", exit, "workspace", ws, "packs", packs,
		"files", files, "lock", map[bool]string{true: "written", false: "unchanged"}[p.LockWrite],
		"settings", settings, "runs_code", runsCode, "allow_exec", p.AllowExec, "left_hooks", leftHooks, "unverified", unverified, "conflicts", conflicts, "plugins", plugins, "error", errorJSON(refusal))
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

// errorJSON is a refusal for programs: what is wrong and the next step (spec §3); null for none.
func errorJSON(e *Error) any {
	if e == nil {
		return nil
	}
	return objectOf("exit", e.Exit, "message", e.What, "next", e.Next)
}

// ErrorJSON is a refusal that came before any plan, for programs.
func ErrorJSON(command string, e *Error) schema.Object {
	return objectOf("command", command, "result", "refused", "exit", e.Exit, "error", errorJSON(e))
}

// CheckJSON is check's result for programs.
func (r *CheckResult) JSON(exit int) schema.Object {
	list := func(fs []Finding) []any {
		out := []any{}
		for _, f := range fs {
			out = append(out, objectOf("code", f.Code, "file", f.File, "message", f.Message, "next", f.Next))
		}
		return out
	}
	result := "ok"
	if len(r.Findings) > 0 {
		result = "findings"
	}
	return objectOf("command", "check", "result", result, "exit", exit, "findings", list(r.Findings),
		"warnings", list(r.Warnings), "error", nil)
}

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
	return b.String()
}
