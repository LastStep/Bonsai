package main

// bonsai hook <name>: the hook entry point (spec §4, §7). This build answers `bonsai hook guard` (plan part 5,
// internal/guard), `bonsai hook start` and `bonsai hook record` (step 5.2.4, internal/recorder); stop comes with step
// 5.3, and the engine writes no line that calls it. Its tables are hookWord and its sub-words'. A hook speaks Claude
// Code's hook format: it takes no --json, and its refusals are lines on stderr.

import (
	"os"

	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/recorder"
)

func init() { register(hookWord) }

const notBuiltHookNext = "take that hook line out of .claude/settings.json; bonsai init writes no line that calls it in this build"

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
		startWord,
		{Name: "hook stop", Later: "step 5.3", LaterNext: notBuiltHookNext},
		recordWord,
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

var startWord = &Word{
	Name:    "hook start",
	Title:   "the SessionStart hook that prints a session's opening context and records its start",
	Summary: "the SessionStart hook",
	About: `  bonsai hook start
on SessionStart (no matcher: startup, resume, clear, compact and fork alike), timeout 10 s, synchronous.
It reads Claude Code's SessionStart payload on stdin and the project from CLAUDE_PROJECT_DIR (else the payload's cwd).
It first writes the session's session_start record (bonsai.log/1) in the main checkout's .bonsai/local/log/, with the
active task it finds, the model and the source, and the bonsai binary's path (from ~/ under the home) and SHA-256,
hashed on every start. Then it prints the opening context on stdout, which Claude Code adds to the session: the
workspace, the active task (status, lane, branch, grants), its last ladder result, and the label definitions attached
on this machine; at most 60 lines. In a project with no bonsai.yaml it prints and records nothing.
`,
	FlagNext: "use the hook line bonsai init writes: bonsai hook start",
	Exits: []Exit{
		{Code: 0, Means: "always: what it could read is printed, and the session goes on"},
	},
	Examples: []string{"CLAUDE_PROJECT_DIR=. bonsai hook start < payload.json"},
	Run: func(c *call) int {
		return recorder.Start(recorder.Options{Stdin: os.Stdin, StdinTerminal: isTerminal(os.Stdin)}, c.stdout)
	},
}

var recordWord = &Word{
	Name:    "hook record",
	Title:   "the recorder: one log record for each of ten of Claude Code's hook events",
	Summary: "the recorder's hook",
	About: `  bonsai hook record
on UserPromptSubmit, PreToolUse and PostToolUse (every tool), PermissionRequest, PostToolUseFailure, Notification,
SubagentStart, SubagentStop and Stop (async, timeout 10 s), and on SessionEnd (synchronous, timeout 5 s, so the
line is not lost at the session's exit).
It reads the event's payload on stdin and the project from CLAUDE_PROJECT_DIR (else the payload's cwd), and appends one
bonsai.log/1 record to the session's file in the main checkout's .bonsai/local/log/: every free-text string redacted,
a tool call reduced to its target, and a prompt's words never kept. It prints nothing and exits 0 whatever it is
handed; a payload it cannot read, or a project with no bonsai.yaml, writes nothing.
`,
	FlagNext: "use the hook line bonsai init writes: bonsai hook record",
	Exits: []Exit{
		{Code: 0, Means: "always: nothing is printed, and the session goes on"},
	},
	Examples: []string{"CLAUDE_PROJECT_DIR=. bonsai hook record < payload.json"},
	Run: func(c *call) int {
		return recorder.Record(recorder.Options{Stdin: os.Stdin, StdinTerminal: isTerminal(os.Stdin)})
	},
}
