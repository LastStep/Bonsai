package main

// bonsai hook <name>: the hook entry point (spec §4, §7). This build answers `bonsai hook guard` (plan part 5,
// internal/guard), `bonsai hook stop` (step 5.3.4, internal/stopgate), `bonsai hook start` and `bonsai hook record`
// (step 5.2.4, internal/recorder). Its tables are hookWord and its sub-words'. A hook speaks Claude Code's hook format:
// it takes no --json, and its refusals are lines on stderr.

import (
	"os"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/recorder"
	"github.com/LastStep/Bonsai/internal/stopgate"
)

func init() { register(hookWord) }

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
		stopWord,
		recordWord,
	},
}

var guardWord = &Word{
	Name:    "hook guard",
	Title:   "the PreToolUse guard that bonsai init writes as the hook line",
	Summary: "the PreToolUse guard",
	About: `  bonsai hook guard || exit 2
on PreToolUse (Edit, Write, MultiEdit, NotebookEdit, Bash, PowerShell), timeout 10 s.
It reads Claude Code's PreToolUse payload on stdin and the session's project from CLAUDE_PROJECT_DIR, and refuses an
Edit, Write, MultiEdit or NotebookEdit of a path on the protected or person_only list of a project holding it, or
among the protected paths its packs declare (the lock). A path is judged by every folder at or above it holding
bonsai.yaml, past any .git; the nearest alone when it is a verified checkout. A worktree judges by its main
checkout's bonsai.yaml and lock once the main checkout is verified: git's own link back names the worktree, and
this machine's record (written by bonsai init and update in the main checkout) is the first to claim the id. Where
nothing holding a refused path is verified, no task's grant can count (rule main-not-verified), and the refusal names
the step: a person runs bonsai update in the main checkout. Bash and PowerShell are allowed in this build. It blocks
whatever it cannot read or decide, and blocks itself after 5 s, before the hook's timeout could let the call through.
When neither the session's folder nor the path is in a linked project, it allows and records nothing.
Each decision is a guard record (bonsai.log/1) in the .bonsai/local/log/ of the project whose rules decided (a
verified worktree's main checkout's), with the input's hash, and on the record that made its file the binary's own
path (from ~/ under the home) and SHA-256.
` + engine.SessionFolderWords,
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

var stopWord = &Word{
	Name:    "hook stop",
	Title:   "the Stop gate that bonsai init writes as the hook line",
	Summary: "the Stop gate",
	About: `  bonsai hook stop || exit 2
on Stop (no matcher), timeout 10 s.
It engages only when BONSAI_TASK names a task and the session's own project (from CLAUDE_PROJECT_DIR) is linked;
otherwise it lets the session end and records nothing. Engaged, it keeps the session going (exit 2, the reason and
the next step on stderr, which Claude Code gives the agent as the reason to go on) until all of these hold: the
checkout's main checkout is verified (git's own link back, and this machine's record written by bonsai init and
update there); the named task is one task file there that parses; this checkout holds a run report for the task
(documents.run, R-<date>-<task id>.md); and the main checkout's .bonsai/local/ladder/<task id>.json reads as
bonsai.ladder/1 with mode local, green, git.sha this checkout's HEAD (read from its git files: a branch's ref file,
packed-refs or a detached HEAD), git.dirty false, and the workspace's ladder_floor and the task's bonsai.ladder each
among its requested rungs and green. The next step names bonsai ladder --task <id>.
This checkout is the one holding the session's working folder (the payload's cwd) when it is a verified checkout of
the same project, such as a worktree the session entered with EnterWorktree; else CLAUDE_PROJECT_DIR's.
It does not read stop_hook_active: Claude Code itself keeps a turn going eight times in a row at most, with no tool
call between them, and ends it at the ninth block (CLAUDE_CODE_STOP_HOOK_BLOCK_CAP); each block is a guard record
(bonsai.log/1, agent_event Stop) in the main checkout's .bonsai/local/log/, as each allow is. It blocks whatever it
cannot read or decide, and blocks itself after 5 s, before the hook's timeout could let the session end unchecked.
`,
	FlagNext: "use the hook line bonsai init writes: bonsai hook stop || exit 2",
	Exits: []Exit{
		{Code: 0, Means: "allow: the session ends (nothing printed)"},
		{Code: 2, Means: "block: the session goes on, with the reason and the next step on stderr"},
	},
	Examples: []string{"BONSAI_TASK=T-0001 CLAUDE_PROJECT_DIR=. bonsai hook stop < payload.json"},
	Run: func(c *call) int {
		if len(c.rest) > 0 {
			return c.refuse(c.flagError("hook stop takes no %+q", c.rest[0]))
		}
		return stopgate.Main(stopgate.Options{Stdin: os.Stdin, Stderr: c.stderr, StdinTerminal: isTerminal(os.Stdin)})
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
