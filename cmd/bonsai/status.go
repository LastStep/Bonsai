package main

// bonsai status (spec §4, contract §12): one workspace at a glance. Its table of flags and exit codes is statusWord.

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/status"
)

func init() { register(statusWord) }

var statusWord = &Word{
	Name:    "status",
	Order:   4,
	Title:   "one workspace at a glance (contract section 12), from the folder you run it in.",
	Summary: "one workspace at a glance",
	Args:    "[flags]",
	About: `By default it is cheap and offline (it runs git and reads files, never Claude Code or the network), for a
program to run often. problems are the findings bonsai check reports, one sentence each with its next step; needs
lists the Claude Code floor and each locked pack. Fields that do not apply print null or [] (contract section 12).
`,
	Flags: []Flag{
		{Name: "--json", Help: "print the bonsai.status/1 document (for programs) instead of plain text; with --active,\nthe active_task object alone"},
		{Name: "--full", Help: "add checks (mode full): each pack's newer release tags (git ls-remote), its plugin as Claude\n" +
			"Code reports it (claude plugin list), Claude Code's version against the floor (claude --version),\n" +
			"and the MCP servers the packs' needs name; a pack whose plugin is not installed here becomes a\nneeds entry of kind plugin"},
		{Name: "--active", Help: "print only the active task (contract section 13): its id and how it was found (named, or the\n" +
			"one task reading running), or none and why; it reads bonsai.yaml and the task files only"},
		{Name: "--line", Help: "print the statusline's workspace half", Later: "step 5.6"},
	},
	Exits: []Exit{
		{Code: 0, Means: "the status was read (a check --full could not make reads unknown, never a failure)"},
		{Code: 2, Means: "bad input (a flag status does not take, one not built yet, or --active with --full); --json prints\nformat, bonsai, problems [] and error, and every other field null"},
		{Code: 3, Means: "the workspace cannot be read at all (not in a git checkout, no bonsai.yaml, a bonsai.yaml Bonsai refuses);\n--json then prints format, bonsai, problems and error, and every other field null (with --active, the\nerror object alone)"},
	},
	Examples: []string{"bonsai status --json", "bonsai status --full --json", "bonsai status --active --json"},
	Refused: func(c *call, e *engine.Error) encoder {
		return statusDoc(status.Refused(version, e, []any{}))
	},
	Run: runStatus,
}

// statusDoc is a status document for printDoc: status.Encode holds it to bonsai.status/1.
type statusDoc schema.Object

func (d statusDoc) Encode() ([]byte, error) { return status.Encode(schema.Object(d)) }

func runStatus(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("status takes no %+q", c.rest[0]))
	}
	if c.has("--active") && c.has("--full") {
		e := c.flagError("status --active prints only the active task, which --full adds nothing to")
		e.Next = "run `bonsai status --active`, or `bonsai status --full` for every field"
		return c.refuse(e)
	}
	dir, err := os.Getwd()
	if err != nil {
		return c.fail(&engine.Error{Code: "read-failed", Exit: exitRuntime, What: "the current folder cannot be read",
			Next: "run it from a folder inside the project"})
	}
	if c.has("--active") {
		return runActive(c, dir)
	}
	o := status.Options{Full: c.has("--full")}
	if o.Full {
		o.Plugins, o.Claude = pluginCLI, claudeVersion
	}
	doc, code := status.BuildWith(dir, version, o)
	if code != status.ExitOK {
		if e, ok := doc.Get("error"); ok && e != nil {
			noted(c.word, &engine.Error{Code: e.(schema.Object).String("code"), Exit: code})
		}
	}
	if c.json {
		return c.printDoc(statusDoc(doc), code)
	}
	text := status.Text(doc)
	if code != status.ExitOK {
		_, _ = io.WriteString(c.stderr, text)
		return code
	}
	if write(c.stdout, text) != exitOK {
		return exitRuntime
	}
	return code
}

// runActive is status --active (contract §13's read-only command): the active task alone, with --json the
// active_task object held to its schema, or, when Bonsai cannot read the workspace, the error object alone (exit 3).
func runActive(c *call, dir string) int {
	a, e := status.Active(dir)
	if e != nil {
		e.Exit = exitRuntime
		noted(c.word, e)
		if c.json {
			return c.printDoc(errorOnly{e}, e.Exit)
		}
		_, _ = fmt.Fprintf(c.stderr, "bonsai status: %s.\nnext: %s\n", strings.TrimSuffix(e.What, "."), e.Next)
		return e.Exit
	}
	if c.json {
		return c.printDoc(activeDoc(a.Object()), exitOK)
	}
	if a.ID != "" {
		return c.print(fmt.Sprintf("Active task: %s (%s)\n", a.ID, a.How))
	}
	return c.print("Active task: none (" + engine.ASCII(strings.TrimPrefix(a.Sentence(), "no active task: ")) + ")\n")
}

// activeDoc is status --active --json's object, held to bonsai.status/1's active_task before it is printed.
type activeDoc schema.Object

func (d activeDoc) Encode() ([]byte, error) {
	if msgs := schema.Validate(status.ActiveSchema(), schema.Object(d)); len(msgs) > 0 {
		return nil, fmt.Errorf("active_task does not fit bonsai.status/1: %v", msgs)
	}
	return schema.Encode(schema.Object(d))
}
