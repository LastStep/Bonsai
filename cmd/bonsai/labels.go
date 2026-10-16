package main

// bonsai labels attach and detach (spec section 4; contract sections 5.1-5.3; plan-5 5.6.0 and 5.6.2): label
// definitions attached on this machine for one workspace, a bonsai.labels/1 file copied into the workspace's machine
// folder (labels/<namespace>.yaml in the Bonsai home), or taken out. Step 5.6.0 places the word: its table, its
// sub-words' full flag tables, exit codes and examples, so bonsai --help and bonsai --help --json show its final
// shape; both sub-words are marked Later until step 5.6.2 builds them, and refused as not-built till then. With --json
// each prints bonsai.attach/1 (format.Attach). Both refuse themselves in an agent's session (personOnly, agent.go).

import (
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
)

func init() { register(labelsWord) }

// labelsLater names the step that builds both sub-words.
const labelsLater = "step 5.6.2"

var labelsWord = &Word{
	Name:    "labels",
	Order:   17,
	Title:   "label definitions attached on this machine for this workspace (contract section 5.3).",
	Summary: "attach or detach label definitions on this machine",
	Args:    "<action>",
	About: "bonsai labels attach copies a bonsai.labels/1 file into the workspace's machine folder; bonsai labels detach\n" +
		"takes it out. Both are a person's commands. bonsai status --json lists what is attached (labels, from machine).\n",
	Exits: []Exit{
		{Code: 0, Means: "done (or the help was printed)"},
		{Code: 2, Means: "bad input: no action, one this build does not have, or the action's own (its --help)"},
		{Code: 3, Means: "a file could not be read or written (the action's)"},
		{Code: 4, Means: "not linked, a namespace taken or a definition redefined, or an agent's session (the action's)"},
	},
	Examples: []string{"bonsai labels attach --help"},
	Subs:     []*Word{labelsAttachWord, labelsDetachWord},
}

// labelsRefused gives a labels sub-word's --json document for a refusal.
func labelsRefused(command string) func(c *call, e *engine.Error) encoder {
	return func(c *call, e *engine.Error) encoder { return &format.Attach{Command: command, Error: e.Object()} }
}

var labelsAttachWord = &Word{
	Name:    "labels attach",
	Title:   "attach a file of label definitions on this machine: a person's command",
	Summary: "attach a bonsai.labels/1 file (a person's command)",
	Args:    "<file> [flags]",
	About: `The file is read as bonsai.labels/1 and copied, its line endings made LF, to the workspace's machine folder
as labels/<namespace>.yaml. A namespace that is bonsai or a locked pack's is refused (namespace-taken). Attached
again, the same content is unchanged, and a newer version that only adds definitions is updated; anything else (a
definition changed or removed, other content at the same or an older version) is refused (label-redefined): a
definition may be added, never redefined, so a person detaches the namespace first.
In an agent's session (CLAUDE_CODE_CHILD_SESSION set) it refuses itself (person-command, exit 4) and writes nothing:
a person types it in a terminal of their own, not inside Claude (its ! lines count as an agent's session).
`,
	Flags: []Flag{{Name: "--json", Help: "print the bonsai.attach/1 document (for programs) instead of plain text"}},
	Exits: []Exit{
		{Code: 0, Means: "attached, updated, or unchanged"},
		{Code: 2, Means: "bad input: no file, or one that does not read as bonsai.labels/1 (bad-value, naming the field),\n" +
			"or a flag it does not take. Nothing was written"},
		{Code: 3, Means: "the file could not be read, or the machine folder's copy written (read-failed, write-failed)"},
		{Code: 4, Means: "not linked (not-linked), a namespace taken (namespace-taken), a definition redefined\n" +
			"(label-redefined), or an agent's session (person-command). Nothing was written"},
	},
	Examples: []string{"bonsai labels attach studio-labels.yaml --json"},
	Later:    labelsLater,
	Refused:  labelsRefused("attach"),
	Run: func(c *call) int {
		return runLabelsPerson(c, "bonsai labels attach", "it attaches label definitions on this machine", "labels attach")
	},
}

var labelsDetachWord = &Word{
	Name:    "labels detach",
	Title:   "detach a namespace's label definitions from this machine: a person's command",
	Summary: "detach a namespace's definitions (a person's command)",
	Args:    "<namespace> [flags]",
	About: `It removes the workspace's machine folder's labels/<namespace>.yaml (detached), or does nothing when there is
none (nothing). It runs in a checkout whose bonsai.yaml is gone too, so a project taken out can be unregistered.
In an agent's session (CLAUDE_CODE_CHILD_SESSION set) it refuses itself (person-command, exit 4) and removes nothing:
a person types it in a terminal of their own, not inside Claude (its ! lines count as an agent's session).
`,
	Flags: []Flag{{Name: "--json", Help: "print the bonsai.attach/1 document (for programs) instead of plain text"}},
	Exits: []Exit{
		{Code: 0, Means: "detached, or nothing to detach"},
		{Code: 2, Means: "bad input: no namespace, one not in a namespace's form (bad-value), or a flag it does not take"},
		{Code: 3, Means: "the file could not be removed (write-failed)"},
		{Code: 4, Means: "not in a git checkout (not-a-checkout), or an agent's session (person-command). Nothing was removed"},
	},
	Examples: []string{"bonsai labels detach studio --json"},
	Later:    labelsLater,
	Refused:  labelsRefused("detach"),
	Run: func(c *call) int {
		return runLabelsPerson(c, "bonsai labels detach", "it detaches label definitions from this machine", "labels detach")
	},
}

// runLabelsPerson is a labels sub-word's run: the agent's-session refusal first, then the command, which step 5.6.2
// builds. It is not reached while the sub-word is marked Later (the not-built refusal comes first).
func runLabelsPerson(c *call, shown, what, name string) int {
	if e := personOnly(shown, what, c.line()); e != nil {
		return c.refuse(e)
	}
	return c.refuse(notBuiltYet(name, labelsLater))
}
