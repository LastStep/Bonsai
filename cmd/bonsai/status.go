package main

// bonsai status (spec §4, contract §12): one workspace at a glance. Its table of flags and exit codes is statusWord.

import (
	"io"
	"os"

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
	Flags: []Flag{
		{Name: "--json", Help: "print the bonsai.status/1 document (for programs) instead of plain text"},
		{Name: "--full", Help: "add checks: newer pack tags, Claude Code's version, MCP servers reachable", Later: "step 5.1.6"},
		{Name: "--active", Help: "print only the active task", Later: "step 5.1.6"},
		{Name: "--line", Help: "print the statusline's workspace half", Later: "step 5.6"},
	},
	Exits: []Exit{
		{Code: 0, Means: "the status was read"},
		{Code: 2, Means: "bad input (a flag status does not take, or one not built yet); --json prints format, bonsai,\nproblems [] and error, and every other field null"},
		{Code: 3, Means: "the workspace cannot be read at all (not in a git checkout, no bonsai.yaml, a bonsai.yaml Bonsai refuses);\n--json then prints format, bonsai, problems and error, and every other field null"},
	},
	Examples: []string{"bonsai status --json"},
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
	dir, err := os.Getwd()
	if err != nil {
		return c.fail(&engine.Error{Code: "read-failed", Exit: exitRuntime, What: "the current folder cannot be read",
			Next: "run it from a folder inside the project"})
	}
	doc, code := status.Build(dir, version)
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
