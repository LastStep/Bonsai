package main

// bonsai settings show and set (spec section 4, section 6; contract section 3, section 10.1; plan-5 5.6.0 and 5.6.1):
// this machine's settings for the workspace, in its machine folder's settings.json in the Bonsai home, or with
// --machine the home's own, in the home's settings.json. Step 5.6.0 placed the word; step 5.6.1 builds both
// sub-words on internal/workspace's settings files (homesettings.go: the keys' table, their one home, its rules,
// the reader and the writer). With --json each prints bonsai.settings/1 (format.Settings). settings show never
// refuses in an agent's session; settings set refuses itself there (personOnly, agent.go), since a machine's
// settings are a person's.

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/textui"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func init() {
	register(settingsWord)
	settingsShowWord.Notes = settingsKeysText()
	settingsSetWord.Notes = settingsKeysText()
}

var settingsWord = &Word{
	Name:    "settings",
	Order:   16,
	Title:   "this machine's settings, for the workspace or for the Bonsai home (spec section 4, section 6).",
	Summary: "read or set this machine's settings for the workspace, or the home's",
	Args:    "<action>",
	About: "bonsai settings show reads them (agents may run it); bonsai settings set changes them, a person's command.\n" +
		"They live in the Bonsai home, never in the project: the workspace's in its machine folder's settings.json,\n" +
		"the home's own (--machine) in the home's settings.json.\n",
	Exits: []Exit{
		{Code: 0, Means: "done (or the help was printed)"},
		{Code: 2, Means: "bad input: no action, one this build does not have, or the action's own (its --help)"},
		{Code: 3, Means: "a settings file could not be read or written (the action's)"},
		{Code: 4, Means: "not linked, or an agent's session (the action's)"},
	},
	Examples: []string{"bonsai settings show --help"},
	Subs:     []*Word{settingsShowWord, settingsSetWord},
}

// settingsFlags are the flags both sub-words take.
func settingsFlags(json string) []Flag {
	return []Flag{
		{Name: "--machine", Help: "the Bonsai home's own settings (cache_keep_days; personal_from on Windows), not the workspace's"},
		{Name: "--json", Help: json},
	}
}

// settingsRefused is a settings sub-word's --json document for a refusal: the scope the command line asked for, the
// lists empty.
func settingsRefused(c *call, e *engine.Error) encoder {
	scope := "workspace"
	if c.has("--machine") {
		scope = "machine"
	}
	return &format.Settings{Scope: scope, Error: e.Object()}
}

var settingsShowWord = &Word{
	Name:    "settings show",
	Title:   "every key of this machine's settings, with its value in force and where it came from",
	Summary: "every key with its value in force and where it came from",
	Args:    "[flags]",
	About: `It lists every key of the scope with its value in force and where it came from (the file, or the default),
and the file's place. Without --machine it reads the workspace's keys (status_writes, status_command) from the main
checkout's machine folder, from a worktree too; with --machine the home's (cache_keep_days; personal_from on
Windows). It writes nothing, and agents may run it.
`,
	Flags: settingsFlags("print the bonsai.settings/1 document (for programs) instead of plain text"),
	Exits: []Exit{
		{Code: 0, Means: "the settings were printed"},
		{Code: 2, Means: "bad input: a flag or argument it does not take (bad-flag)"},
		{Code: 3, Means: "a settings file could not be read, or holds a value outside its key's rule (read-failed, naming\n" +
			"the file and the key)"},
		{Code: 4, Means: "outside a linked checkout without --machine (not-linked)"},
	},
	Examples: []string{"bonsai settings show --json", "bonsai settings show --machine --json"},
	Refused:  settingsRefused,
	Run:      runSettingsShow,
}

var settingsSetWord = &Word{
	Name:    "settings set",
	Title:   "change this machine's settings, all or nothing: a person's command",
	Summary: "change keys, all or nothing (a person's command)",
	Args:    "<key>=<value> [<key>=<value> ...] [flags]",
	About: `Every value is checked first, then all are written or none. status_writes=command needs a status_command,
given or already set; status_writes=agents drops status_command. A value already in force writes nothing (changed is
then []). It asks nothing and takes no --yes: each key=value is the change, typed.
In an agent's session (CLAUDE_CODE_CHILD_SESSION set) it refuses itself (person-command, exit 4) and writes nothing:
a person types it in a terminal of their own, not inside Claude (its ! lines count as an agent's session).
`,
	Flags: settingsFlags("print the bonsai.settings/1 document (for programs) instead of plain text; changed lists the keys written"),
	Exits: []Exit{
		{Code: 0, Means: "written, or every value was already in force"},
		{Code: 2, Means: "bad input: an unknown key or a value outside its rule (bad-value, naming the allowed values or\n" +
			"listing the keys), not key=value, or a flag it does not take. Nothing was written"},
		{Code: 3, Means: "a settings file could not be read or written (read-failed, write-failed)"},
		{Code: 4, Means: "not linked without --machine (not-linked), or an agent's session (person-command)"},
	},
	Examples: []string{
		`bonsai settings set status_writes=command "status_command=studio task move"`,
		"bonsai settings set cache_keep_days=90 --machine --json",
	},
	Refused: settingsRefused,
	Run:     runSettingsSet,
}

// runSettingsSet is settings set's run: the agent's-session refusal first, then every value checked, then all
// written or none.
func runSettingsSet(c *call) int {
	if e := personOnly("bonsai settings set", "it changes this machine's settings", c.line()); e != nil {
		return c.refuse(e)
	}
	f, e := settingsFile(c)
	if e != nil {
		return c.refuse(e)
	}
	changed, err := f.Set(c.rest)
	if err != nil {
		return c.refuse(settingsError(err))
	}
	return settingsOut(c, f, changed)
}

// runSettingsShow is settings show's run: it reads and prints, and writes nothing.
func runSettingsShow(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("settings show takes no argument, and %+q was given", c.rest[0]))
	}
	f, e := settingsFile(c)
	if e != nil {
		return c.refuse(e)
	}
	return settingsOut(c, f, []string{})
}

// settingsFile opens the call's settings file: the home's with --machine, else the workspace's, in the main
// checkout's machine folder (from a worktree too), which needs a linked checkout.
func settingsFile(c *call) (*workspace.SettingsFile, *engine.Error) {
	home, err := workspace.Home()
	if err != nil {
		e := engine.HomeError(err)
		e.Exit = exitRuntime
		return nil, e
	}
	main := ""
	scope := workspace.ScopeMachine
	if !c.has("--machine") {
		scope = workspace.ScopeWorkspace
		co, err := workspace.Find(".")
		if err != nil {
			e := engine.FindError(err)
			e.Code, e.Exit = "not-linked", engine.ExitState
			return nil, e
		}
		if _, err := os.Stat(filepath.Join(co.Root, workspace.ConfigFile)); err != nil {
			return nil, &engine.Error{Code: "not-linked", Exit: engine.ExitState,
				What: "bonsai.yaml is not in this checkout, so it is not linked to Bonsai: the workspace's settings need a linked checkout",
				Next: "run it in a linked project's checkout, or give --machine for the Bonsai home's own settings"}
		}
		main = co.Main
	}
	f, err := workspace.OpenSettings(scope, home, main)
	if err != nil {
		return nil, settingsError(err)
	}
	return f, nil
}

// settingsError gives a settings file's error, or settings set's refusal of its input, with its word.
func settingsError(err error) *engine.Error {
	var se *workspace.SettingError
	if errors.As(err, &se) {
		return &engine.Error{Code: "bad-value", Exit: exitInput, What: engine.ASCII(se.Msg) + ": nothing was written", Next: engine.ASCII(se.Next)}
	}
	var we *workspace.Error
	if errors.As(err, &we) {
		code := we.Code
		if code != "write-failed" {
			code = "read-failed"
		}
		return &engine.Error{Code: code, Exit: exitRuntime, What: engine.ASCII(we.File + ": " + we.Msg), Next: engine.ASCII(we.Next)}
	}
	e := engine.Unexpected(err)
	return e
}

// settingsOut prints the scope's settings in force, and what set changed.
func settingsOut(c *call, f *workspace.SettingsFile, changed []string) int {
	in, err := f.InForce()
	if err != nil {
		return c.refuse(settingsError(err))
	}
	show := f.Show()
	if c.json {
		doc := &format.Settings{Scope: f.Scope, File: &show, Settings: []format.SettingValue{}, Changed: changed}
		for _, s := range in {
			doc.Settings = append(doc.Settings, format.SettingValue{Key: s.Key, Value: s.Value, From: s.From})
		}
		return c.printDoc(doc, exitOK)
	}
	d := &textui.Doc{}
	whose := "this workspace on this machine"
	if f.Scope == workspace.ScopeMachine {
		whose = "the Bonsai home (--machine)"
	}
	if c.word.Name == "settings set" {
		note := "nothing changed: every value was already in force"
		if len(changed) > 0 {
			note = "written: " + strings.Join(changed, ", ")
		}
		d.Title("bonsai settings set", note)
	} else {
		d.Title("bonsai settings show", whose)
	}
	var rows []textui.Row
	for _, s := range in {
		value, tone := "none", textui.Dim
		switch v := s.Value.(type) {
		case string:
			value, tone = v, textui.Path
		case int64:
			value, tone = strconv.FormatInt(v, 10), textui.Path
		}
		from := "from the file"
		if s.From == workspace.FromDefault {
			from = "the default"
		}
		rows = append(rows, textui.R(textui.T(s.Key, textui.Key), textui.T(engine.ASCII(value), tone), textui.T(from, textui.Dim)))
	}
	d.Section("Settings", rows...)
	where := "not written yet: every key holds its default"
	if f.Exists {
		where = "this machine's settings for " + whose
	}
	d.Section("File", textui.R(textui.T(engine.ASCII(show), textui.Path), textui.C(where)))
	d.Gap()
	d.Hint("every key, its values and its default: bonsai settings set --help")
	if c.word.Name == "settings show" {
		d.Hint("a person changes one in a terminal of their own: bonsai settings set <key>=<value>" + map[bool]string{true: " --machine", false: ""}[f.Scope == workspace.ScopeMachine])
	}
	return c.show(d)
}

// settingsKeysText is the keys' table (workspace.SettingKeys, their one home) for both sub-words' --help.
func settingsKeysText() string {
	var b strings.Builder
	for _, scope := range []string{workspace.ScopeWorkspace, workspace.ScopeMachine} {
		title := "The workspace's keys (without --machine):"
		if scope == workspace.ScopeMachine {
			title = "The home's own keys (with --machine):"
		}
		b.WriteString(title + "\n")
		for _, k := range workspace.SettingKeys {
			if k.Scope != scope {
				continue
			}
			only := ""
			if k.Windows {
				only = " (Windows only)"
			}
			wrapTo(&b, "  ", k.Key+only+": "+k.Values+"; default "+k.Default+".")
			wrapTo(&b, "      ", k.Means+".")
		}
	}
	return b.String()
}

// wrapTo writes text in lines of at most 116 characters, each starting with indent.
func wrapTo(b *strings.Builder, indent, text string) {
	line := indent
	for _, w := range strings.Fields(text) {
		if len(line) > len(indent) && len(line)+1+len(w) > 116 {
			b.WriteString(line + "\n")
			line = indent
		}
		if len(line) > len(indent) {
			line += " "
		}
		line += w
	}
	b.WriteString(line + "\n")
}
