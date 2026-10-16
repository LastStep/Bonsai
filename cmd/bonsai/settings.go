package main

// bonsai settings show and set (spec section 4, section 6; contract section 3, section 10.1; plan-5 5.6.0 and 5.6.1):
// this machine's settings for the workspace, in its machine folder's settings.json in the Bonsai home, or with
// --machine the home's own, in the home's settings.json. Step 5.6.0 places the word: its table, its sub-words' full
// flag tables, exit codes and examples, so bonsai --help and bonsai --help --json show its final shape; both sub-words
// are marked Later until step 5.6.1 builds them, and refused as not-built till then. With --json each prints
// bonsai.settings/1 (format.Settings). settings show never refuses in an agent's session; settings set refuses itself
// there (personOnly, agent.go), since a machine's settings are a person's.

import (
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
)

func init() { register(settingsWord) }

// settingsLater names the step that builds both sub-words.
const settingsLater = "step 5.6.1"

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
	Later:    settingsLater,
	Refused:  settingsRefused,
	Run:      func(c *call) int { return c.refuse(notBuiltYet("settings show", settingsLater)) },
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
	Later:   settingsLater,
	Refused: settingsRefused,
	Run:     runSettingsSet,
}

// runSettingsSet is settings set's run: the agent's-session refusal first, then the command, which step 5.6.1 builds.
// It is not reached while the sub-word is marked Later (the not-built refusal comes first).
func runSettingsSet(c *call) int {
	if e := personOnly("bonsai settings set", "it changes this machine's settings", c.line()); e != nil {
		return c.refuse(e)
	}
	return c.refuse(notBuiltYet("settings set", settingsLater))
}

// notBuiltYet is the refusal of a sub-word whose step has not built it yet.
func notBuiltYet(name, step string) *engine.Error {
	return &engine.Error{Code: "not-built", Exit: exitInput,
		What: "bonsai " + name + " is not built yet (it comes with a later part of the rebuild: " + step + ")",
		Next: "run `bonsai " + name + " --help`"}
}
