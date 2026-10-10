package main

// bonsai asks (spec section 4, section 8; contract section 9; plan-5 5.2.5): the open asks of this project, newest
// first, or with --all every key and where it stands. It reads the main checkout's .bonsai/local/asks/ and writes
// nothing. With --json it prints bonsai.asks/1 (format.Asks). Its table is asksWord.

import (
	"fmt"
	"strings"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/redact"
)

func init() { register(asksWord) }

var asksWord = &Word{
	Name:    "asks",
	Order:   10,
	Title:   "the asks of this project (contract section 9).",
	Summary: "list the open asks",
	Args:    "[flags]",
	About: `It reads the main checkout's .bonsai/local/asks/ (a worktree reads its main checkout's) and writes nothing.
Each ask is one line, newest first: its state, when it was filed, its type, its key and its title. An ask's state
is its latest record: open once filed, answered or resolved once closed. A line that does not read is skipped and
counted.
`,
	Flags: []Flag{
		{Name: "--all", Help: "every key, answered and resolved too, not only the open asks"},
		{Name: "--json", Help: "print the bonsai.asks/1 document (for programs) instead of plain text"},
	},
	Exits: []Exit{
		{Code: 0, Means: "the asks were read (none prints a line saying so)"},
		{Code: 2, Means: "bad input (a flag asks does not take); bonsai.yaml refused"},
		{Code: 3, Means: "the asks cannot be read, or the answer cannot be printed"},
		{Code: 4, Means: "not in a git checkout, or not linked (bonsai init)"},
	},
	Examples: []string{"bonsai asks", "bonsai asks --all --json"},
	Refused:  asksRefused,
	Run:      runAsks,
}

func runAsks(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("asks takes no %+q", redact.Text(c.rest[0])))
	}
	site, e := askSite()
	if e != nil {
		return c.fail(e)
	}
	entries, unreadable, ae := asks.List(site.place, c.has("--all"))
	if ae != nil {
		return c.fail(askError(ae))
	}
	if c.json {
		doc := &format.Asks{Workspace: site.ref, Asks: []format.AskEntry{}}
		for _, en := range entries {
			doc.Asks = append(doc.Asks, en.Entry())
		}
		return c.printDoc(doc, exitOK)
	}
	var b strings.Builder
	for _, en := range entries {
		typ := "-"
		if en.Filed.Type != nil {
			typ = *en.Filed.Type
		}
		fmt.Fprintf(&b, "%-8s  %s  %-6s  %s  %s\n", en.State, en.Filed.At, typ, en.Key, text0(en.Filed.Title))
	}
	if len(entries) == 0 {
		which := "open asks"
		if c.has("--all") {
			which = "asks"
		}
		b.WriteString("no " + which + " in .bonsai/local/asks/\n")
	}
	if unreadable > 0 {
		fmt.Fprintf(&b, "%d lines of .bonsai/local/asks/ did not read as ask records and were skipped\n", unreadable)
	}
	return c.print(engine.ASCII(b.String()))
}
