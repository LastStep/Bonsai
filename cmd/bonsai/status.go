package main

// bonsai status (spec §4, contract §12): one workspace at a glance. Its table of flags and exit codes is statusWord.

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/status"
	"github.com/LastStep/Bonsai/internal/textui"
)

func init() { register(statusWord) }

// releaseTags reads Bonsai's own release tags for status --full. This package's tests replace it before any test runs
// (release_test.go's init), so no test reaches the network.
var releaseTags = engine.RemoteTags

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
		utcFlag,
		{Name: "--full", Help: "add checks (mode full): each pack's newer release tags (git ls-remote), its plugin as Claude\n" +
			"Code reports it (claude plugin list), Claude Code's version against the floor (claude --version),\n" +
			"and the MCP servers the packs' needs name; a pack whose plugin is not installed here becomes a\nneeds entry of kind plugin. " +
			"It also reads Bonsai's own release tags (git ls-remote, at most daily, 10 s) into the home's\ncache/release.json, which " +
			"every status, check and --line then read offline (bonsai_release)"},
		{Name: "--active", Help: "print only the active task (contract section 13): its id and how it was found (named, or the\n" +
			"one task reading running), or none and why; it reads bonsai.yaml and the task files only"},
		{Name: "--line", Help: "print the statusline's workspace half: one line of ASCII (the active task and its title, its ladder, verify:N,\n" +
			"asks:N, bonsai:vX.Y.Z when a newer release is recorded), or with --json the bonsai.line/1 document. It runs in the\n" +
			"folder it is started in, never reads stdin,\n" +
			"runs no git, Claude Code or network, and never fails: outside a linked folder it prints nothing, and a\n" +
			"bonsai.yaml it cannot read prints `bonsai.yaml unreadable` (exit 0 either way). To use it alone, add to your own\n" +
			"Claude settings: {\"statusLine\": {\"type\": \"command\", \"command\": \"bonsai status --line\"}}. Claude Code starts the\n" +
			"command in the session's folder; a script that runs it elsewhere cd's into the project first. Bonsai writes no\n" +
			"statusLine setting itself. Not with --full or --active"},
	},
	Exits: []Exit{
		{Code: 0, Means: "the status was read (a check --full could not make reads unknown, never a failure)"},
		{Code: 2, Means: "bad input (a flag status does not take, one not built yet, or --active with --full); --json prints\nformat, bonsai, problems [] and error, and every other field null"},
		{Code: 3, Means: "the workspace cannot be read at all (not in a git checkout, no bonsai.yaml, a bonsai.yaml Bonsai refuses);\n--json then prints format, bonsai, problems and error, and every other field null (with --active, the\nerror object alone)"},
	},
	Examples: []string{"bonsai status --json", "bonsai status --full --json", "bonsai status --active --json", "bonsai status --line"},
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
	if c.has("--line") && (c.has("--full") || c.has("--active")) {
		e := c.flagError("status --line prints one line of its own, which --full and --active add nothing to")
		e.Next = "run `bonsai status --line` alone, or `bonsai status --full` or `--active` without --line"
		return c.refuse(e)
	}
	if c.has("--active") && c.has("--full") {
		e := c.flagError("status --active prints only the active task, which --full adds nothing to")
		e.Next = "run `bonsai status --active`, or `bonsai status --full` for every field"
		return c.refuse(e)
	}
	if c.has("--line") {
		return runStatusLine(c)
	}
	dir, err := os.Getwd()
	if err != nil {
		return c.fail(&engine.Error{Code: "read-failed", Exit: exitRuntime, What: "the current folder cannot be read",
			Next: "run it from a folder inside the project"})
	}
	if c.has("--active") {
		return runActive(c, dir)
	}
	o := status.Options{Full: c.has("--full"), GitBash: gitBash}
	if o.Full {
		o.Plugins, o.Claude, o.Release = pluginCLI, claudeVersion, releaseTags
		o.NotKept = func(line string) { _, _ = io.WriteString(c.stderr, engine.ASCII(line)+"\n") }
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
	var d textui.Doc
	status.TextToIn(&d, doc, shownZone(), c.has("--utc"))
	if code != status.ExitOK {
		if e, ok := doc.Get("error"); ok && e != nil {
			eo := e.(schema.Object)
			next, _ := eo.Get("next")
			no, _ := next.(schema.Object)
			d.Refuse(refusalOf(c.word, &engine.Error{Code: eo.String("code"), Exit: code, What: eo.String("message"), Next: no.String("do"),
				Who: no.String("who")}))
		}
		_, _ = io.WriteString(c.stderr, d.Render(textui.For(c.stderr)))
		return code
	}
	if write(c.stdout, d.Render(textui.For(c.stdout))) != exitOK {
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
		_, _ = io.WriteString(c.stderr, refusalText(c.word, e, c.stderr))
		return e.Exit
	}
	if c.json {
		return c.printDoc(activeDoc(a.Object()), exitOK)
	}
	var d textui.Doc
	if a.ID != "" {
		d.Section("Active task", textui.R(textui.T(a.ID, textui.Path), textui.C(a.How)))
	} else {
		d.Section("Active task", textui.R(textui.T("none", textui.Dim), textui.C(engine.ASCII(strings.TrimPrefix(a.Sentence(), "no active task: ")))))
	}
	return c.show(&d)
}

// activeDoc is status --active --json's object, held to bonsai.status/1's active_task before it is printed.
type activeDoc schema.Object

func (d activeDoc) Encode() ([]byte, error) {
	if msgs := schema.Validate(status.ActiveSchema(), schema.Object(d)); len(msgs) > 0 {
		return nil, fmt.Errorf("active_task does not fit bonsai.status/1: %v", msgs)
	}
	return schema.Encode(schema.Object(d))
}

// lineDoc is status --line --json's document, held to bonsai.line/1 when it is encoded.
type lineDoc struct{ l *format.Line }

func (d lineDoc) Encode() ([]byte, error) { return d.l.Encode() }

// runStatusLine is status --line: it never fails (an unreadable current folder reads as not linked), and it reads no stdin.
func runStatusLine(c *call) int {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	l := status.BuildLineWith(dir, version)
	if c.json {
		return c.printDoc(lineDoc{l}, exitOK)
	}
	if l.Line == "" {
		return exitOK
	}
	if write(c.stdout, l.Line+"\n") != exitOK {
		return exitOK // a status line must not fail because its reader went away
	}
	return exitOK
}
