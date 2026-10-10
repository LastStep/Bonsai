package main

// bonsai check (spec §4, §6): findings on this checkout, or with --schema a format. Its table of flags and exit codes
// is checkWord; with --json it prints bonsai.check/1 (format.Check), its error object filled when check could not run.

import (
	"fmt"
	"os"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func init() { register(checkWord) }

// claudeVersion gives claude --version's first line (engine.ClaudeVersion of the claude on the PATH): main sets it;
// nil, as in tests, asks nothing.
var claudeVersion func() (string, error)

var checkWord = &Word{
	Name:    "check",
	Order:   5,
	Title:   "findings on this checkout (spec section 6).",
	Summary: "findings on the lock and the files; --schema F prints a format",
	Args:    "[flags]",
	About: `It checks the project against spec section 6: the lock against the files and bonsai.yaml; the files of
Bonsai's kinds (format 0 new or changed, each read under its format, labels against their definitions, no
absolute path); approve_first from git history; the block, memory index and note budgets; the paths CLAUDE.md,
STATE and memory notes name; Claude Code's settings in the project (each permission rule, disableAllHooks, no
version on a pack's plugin); this machine's record of the checkout and the bonsai on the PATH; .bonsai/.gitignore
and .bonsai/local/ in git; and what Claude Code reports: the packs' plugins (claude plugin list --json) and its
version against the floor (claude --version; Bonsai's ` + engine.ClaudeCodeFloor + `, or a pack's needs.claude_code). Each
finding and warning has a code (bonsai check --schema bonsai.check lists them all) and a next step: the exact
command, written "run: <command>", wherever one fixes it. It fetches nothing, and writes nothing but with --write,
which rebuilds the tasks table (.bonsai/tasks.md) from the task files, in the main checkout only: in a worktree it
refuses, exit 4, naming the main checkout. A tasks table that differs from a rebuild is a warning, never a finding.
Findings exit 1; warnings never change the exit code; with --write the exit code says only whether it wrote.
`,
	Flags: []Flag{
		{Name: "--json", Help: "print the bonsai.check/1 document (for programs) instead of text; with --schema, the format's\nJSON Schema"},
		{Name: "--schema", Value: "<format>", Need: "a format's name", NeedNext: schemaNext,
			Help: "print a format instead: every field in the order a writer writes it, its type and allowed\n" +
				"values, and an open list's known words from Bonsai's table. It reads no project, so it runs\n" +
				"anywhere. <format> is a name (bonsai.task), a short name (task) or a name and major\n" +
				"(bonsai.task/1), one of:",
			More: func() string { return strings.Join(format.Names(), ", ") }},
		{Name: "--write", Help: "rebuild the tasks table in .bonsai/ (the sessions table joins in step 5.2.3), in the main checkout\nonly; the exit code then says only whether it wrote (0 written, 3 could not); findings are still\nlisted"},
		{Name: "--pack", Value: "P", Help: "check a pack folder", Later: "step 5.1.9"},
	},
	Exits: []Exit{
		{Code: 0, Means: "no findings (or the format printed); with --write, the table was written (or already current)"},
		{Code: 1, Means: "findings (never with --write)"},
		{Code: 2, Means: "bad input (a flag check does not take, one not built yet, or a format Bonsai does not know: the\nrefusal lists every name)"},
		{Code: 3, Means: "runtime (git is not on the PATH, the Bonsai home cannot be found, a file cannot be read; with --write,\nthe table could not be written)"},
		{Code: 4, Means: "not a linked checkout (not in a git checkout, or no bonsai.yaml), or --write in a worktree"},
	},
	Examples: []string{"bonsai check --json", "bonsai check --schema bonsai.task"},
	Refused:  func(c *call, e *engine.Error) encoder { return engine.CheckRefused(e) },
	Run:      runCheck,
}

func runCheck(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("check takes no %+q", c.rest[0]))
	}
	if c.has("--schema") {
		return runSchema(c, c.value("--schema"))
	}
	dir, err := os.Getwd()
	if err != nil {
		return c.fail(&engine.Error{Code: "read-failed", Exit: exitRuntime, What: "check cannot read the current folder",
			Next: "run it from a folder inside the project"})
	}
	home, err := workspace.Home()
	if err != nil {
		return c.fail(engine.HomeError(err))
	}
	// --write first, so the check that follows sees the rebuilt table. A refusal (a worktree, a folder not linked)
	// stops here; a write that failed is listed after the findings, and is the exit code.
	var wrote *engine.TablesResult
	var werr *engine.Error
	if c.has("--write") {
		if wrote, werr = engine.WriteTables(dir); werr != nil && werr.Exit != exitRuntime {
			return c.fail(werr)
		}
	}
	r, err := engine.Check(dir, home)
	if err != nil {
		return c.fail(engine.Unexpected(err))
	}
	engine.ComparePlugins(r, pluginCLI)
	engine.CompareClaudeCode(r, claudeVersion)
	exit := engine.ExitOK
	if len(r.Findings) > 0 {
		exit = engine.ExitFindings
	}
	if c.has("--write") {
		// With --write the exit code says only whether it wrote (spec section 6); findings are still listed.
		exit = engine.ExitOK
		if werr != nil {
			exit = werr.Exit
		}
		r.Notes = append(r.Notes, wrote.Notes()...)
	}
	if c.json {
		doc := r.Doc()
		if werr != nil {
			doc.Error = werr.Object()
		}
		return c.printDoc(doc, exit)
	}
	if write(c.stdout, r.Text()) != exitOK {
		return exitRuntime
	}
	if werr != nil {
		noted(c.word, werr)
		_, _ = fmt.Fprintf(c.stderr, "bonsai check: %s.\nnext: %s\n", strings.TrimSuffix(werr.What, "."), werr.Next)
	}
	return exit
}

// runSchema is check --schema: a format for a person, or with --json its schema (contract §2.2).
func runSchema(c *call, name string) int {
	f, ok := format.Lookup(name)
	if !ok {
		return c.refuse(&engine.Error{Code: "unknown-format", Exit: exitInput, What: fmt.Sprintf("%+q is not one of Bonsai's formats", name),
			Next: schemaNext()})
	}
	if c.json {
		out, err := schema.Encode(f.Schema())
		if err != nil {
			return c.fail(engine.Unexpected(err))
		}
		return c.print(string(out))
	}
	return c.print(f.Describe())
}

// schemaNext is the next step of a check --schema refusal: every format's name.
func schemaNext() string {
	return "run `bonsai check --schema <format>` with one of: " + strings.Join(format.Names(), ", ")
}
