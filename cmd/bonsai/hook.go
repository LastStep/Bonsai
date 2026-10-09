package main

// bonsai hook <name>: the hook entry point (spec §4, §7). This build answers `bonsai hook guard` (plan part 5,
// internal/guard); start, stop and record come with steps 5.2 and 5.3, and part 3 writes no line that calls them.
// Its tables are hookWord and its sub-words'. A hook speaks Claude Code's hook format: it takes no --json, and its
// refusals are lines on stderr.

import (
	"os"

	"github.com/LastStep/Bonsai/internal/guard"
)

func init() { register(hookWord) }

const notBuiltHookNext = "take that hook line out of .claude/settings.json; bonsai init writes only the guard's line in this build"

var hookWord = &Word{
	Name:    "hook",
	Order:   6,
	Title:   "the hook entry point Claude Code's hook lines call (spec section 7).",
	Summary: "a hook Claude Code's hook lines call",
	Args:    "<name>",
	About:   "A hook speaks Claude Code's hook format: it takes no --json.\n",
	Exits: []Exit{
		{Code: 0, Means: "allow (or the help was printed)"},
		{Code: 2, Means: "block, or bad input (no hook name, a hook this build does not have)"},
	},
	Examples: []string{"bonsai hook guard --help"},
	Subs: []*Word{
		guardWord,
		{Name: "hook start", Later: "step 5.2.4", LaterNext: notBuiltHookNext},
		{Name: "hook stop", Later: "step 5.3", LaterNext: notBuiltHookNext},
		{Name: "hook record", Later: "step 5.2.4", LaterNext: notBuiltHookNext},
	},
}

var guardWord = &Word{
	Name:    "hook guard",
	Title:   "the PreToolUse guard that bonsai init writes as the hook line",
	Summary: "the PreToolUse guard",
	About: `  bonsai hook guard || exit 2
on PreToolUse (Edit, Write, MultiEdit, NotebookEdit, Bash, PowerShell), timeout 10 s.
It reads Claude Code's PreToolUse payload on stdin and the project from CLAUDE_PROJECT_DIR, and refuses an Edit,
Write, MultiEdit or NotebookEdit of a path on bonsai.yaml's protected or person_only list. Bash and PowerShell are
allowed in this build. It blocks whatever it cannot read or decide, and blocks itself after 5 s, before the hook's
timeout could let the call through. In a project with no bonsai.yaml it allows and records nothing.
Each decision is a guard record (bonsai.log/1) in .bonsai/local/log/, naming the binary's own path, and on a
session's first record its SHA-256.
`,
	FlagNext: "use the hook line bonsai init writes: bonsai hook guard || exit 2",
	Exits: []Exit{
		{Code: 0, Means: "allow: the call goes on (nothing printed)"},
		{Code: 2, Means: "block: the reason and the next step on stderr, which Claude Code shows to the agent"},
	},
	Examples: []string{"CLAUDE_PROJECT_DIR=. bonsai hook guard < payload.json"},
	Run: func(c *call) int {
		if len(c.rest) > 0 {
			return c.refuse(c.flagError("hook guard takes no %+q", c.rest[0]))
		}
		return guard.Main(guard.Options{Stdin: os.Stdin, Stderr: c.stderr, StdinTerminal: isTerminal(os.Stdin)})
	},
}
