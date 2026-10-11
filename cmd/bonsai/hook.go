package main

// bonsai hook <name>: the hook entry point (spec §4, §7). This build answers `bonsai hook guard` (plan part 5,
// internal/guard), `bonsai hook stop` (step 5.3.4, internal/stopgate), `bonsai hook start` and `bonsai hook record`
// (step 5.2.4, internal/recorder). Its tables are hookWord and its sub-words'. A hook speaks Claude Code's hook format:
// it takes no --json, and its refusals are lines on stderr. Each sub-word's help quotes its line as bonsai init writes
// it (engine.GuardCommand and the others, in Bonsai's form: internal/engine/hookline.go), so help and line cannot drift.

import (
	"os"
	"strings"

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
	About: "A hook speaks Claude Code's hook format: it takes no --json. The lines bonsai init writes run bonsai at its installed\n" +
		"place, never the PATH.\n" + engine.HookPlaceWords(),
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

// lineWords open each hook word's help, before its line.
const lineWords = "Its line, as bonsai init writes it (one line in .claude/settings.json, broken across lines here):\n"

var guardWord = &Word{
	Name:    "hook guard",
	Title:   "the PreToolUse guard that bonsai init writes as the hook line",
	Summary: "the PreToolUse guard",
	About: lineWords + engine.ShownLine(engine.GuardCommand) + "on PreToolUse (" + strings.ReplaceAll(engine.GuardMatcher, "|", ", ") +
		"), timeout 10 s, onFailure block.\n" + engine.HookPlaceWords() + `The onFailure field (Claude Code 2.1.295 and later) makes a line that cannot start, or that PowerShell runs on a
Windows machine without Git Bash, refuse the call too; an older Claude Code reads the file and ignores the field.
It reads Claude Code's PreToolUse payload on stdin and the session's project from CLAUDE_PROJECT_DIR, and refuses an
Edit, Write, MultiEdit or NotebookEdit of a path on the protected or person_only list of a project holding it, among
the protected paths its packs declare (the lock), or on the floor of paths guarded whatever the lists say
(bonsai.yaml, .bonsai/, .claude/settings*.json and .git), unless the active task's grant (bonsai.allows, while the
task reads running) opens it; Bonsai's own files (.bonsai/local/, the tables, Bonsai's home) are never granted, and
in command mode the tripwires of contract section 10.1 are refused. A path is judged by every folder at or above it
holding bonsai.yaml, past any .git; the nearest alone when it is a verified checkout. A worktree judges by its main
checkout's bonsai.yaml and lock once the main checkout is verified: git's own link back names the worktree, and
this machine's record (written by bonsai init and update in the main checkout) is the first to claim the id. Where
nothing holding a refused path is verified, no task's grant can count (rule main-not-verified), and the refusal names
the step: a person runs bonsai update in the main checkout.
A Bash, PowerShell or Monitor call is judged too: the delete check refuses a recursive or bulk delete that does not
name what it deletes (git clean -x or -X and git stash --all among them) and a delete or move of a guarded path;
bonsai settings set and labels are refused; and bonsai with --allow-exec, bonsai unlink, and a line that unsets or
changes CLAUDE_CODE_CHILD_SESSION are a person's commands, allowed to an agent only under a person's grant in a
project in command mode (the commands refuse themselves too). It blocks whatever it cannot read or decide, and blocks
itself after 5 s, before the hook's timeout could let the call through. When neither the session's folder nor the
path is in a linked project, it allows and records nothing.
Each decision is a guard record (bonsai.log/1) in the .bonsai/local/log/ of the project whose rules decided (a
verified worktree's main checkout's), with the input's hash, and on the record that made its file the binary's own
path (from ~/ under the home) and SHA-256.
` + engine.SessionFolderWords,
	FlagNext: "use the hook line bonsai init writes (bonsai hook guard --help quotes it)",
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
	About: lineWords + engine.ShownLine(engine.StopCommand) + "on Stop (no matcher), timeout 10 s.\n" + engine.HookPlaceWords() + `It engages only when BONSAI_TASK names a task and the session's own project (from CLAUDE_PROJECT_DIR) is linked;
otherwise it lets the session end and records nothing. Engaged, it keeps the session going (exit 2, the reason and
the next step on stderr, which Claude Code gives the agent as the reason to go on) until all of these hold: the
checkout's main checkout is verified (git's own link back, and this machine's record written by bonsai init and
update there); the named task is one task file there that parses; this checkout holds a run report for the task
(documents.run, R-<date>-<task id>.md); and the main checkout's .bonsai/local/ladder/<task id>.json reads as
bonsai.ladder/1, made for the task (its task that id, its workspace this project's), with mode local, green, git.sha
this checkout's HEAD (read from its git files: a branch's ref file, packed-refs or a detached HEAD), git.dirty false,
and the workspace's ladder_floor and the task's bonsai.ladder each among its requested rungs and green. The next step
names bonsai ladder --task <id>.
This checkout is the git checkout holding the session's working folder (the payload's cwd) when git's own link ties
it to the project (the main checkout, or a worktree of it, such as one the session entered with EnterWorktree);
else CLAUDE_PROJECT_DIR's. A working folder in a checkout that link does not tie to the project (another repository
inside it, a worktree whose link back does not hold) needs the result to prove that checkout's HEAD too.
It does not read stop_hook_active: Claude Code itself keeps a turn going eight times in a row at most, with no tool
call between them, and ends it at the ninth block (CLAUDE_CODE_STOP_HOOK_BLOCK_CAP); each block is a guard record
(bonsai.log/1, agent_event Stop) in the main checkout's .bonsai/local/log/, as each allow is. It blocks whatever it
cannot read or decide, and blocks itself after 5 s, before the hook's timeout could let the session end unchecked.
`,
	FlagNext: "use the hook line bonsai init writes (bonsai hook stop --help quotes it)",
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
	About: lineWords + engine.ShownLine(engine.StartCommand) + `on SessionStart (no matcher: startup, resume, clear, compact and fork alike), timeout 10 s, synchronous.
It reads Claude Code's SessionStart payload on stdin and the project from CLAUDE_PROJECT_DIR (else the payload's cwd).
It first writes the session's session_start record (bonsai.log/1) in the main checkout's .bonsai/local/log/, with the
active task it finds, the model and the source, and the bonsai binary's path (from ~/ under the home) and SHA-256,
hashed on every start. Then it prints the opening context on stdout, which Claude Code adds to the session: the
workspace, the active task (status, lane, branch, grants), its last ladder result, and the label definitions attached
on this machine; at most 60 lines. In a project with no bonsai.yaml it prints and records nothing.
`,
	FlagNext: "use the hook line bonsai init writes (bonsai hook start --help quotes it)",
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
	About: lineWords + engine.ShownLine(engine.RecordCommand) + `on UserPromptSubmit, PreToolUse and PostToolUse (every tool), PermissionRequest, PostToolUseFailure, Notification,
SubagentStart, SubagentStop and Stop (async, timeout 10 s), and on SessionEnd (synchronous, timeout 5 s, so the
line is not lost at the session's exit).
It reads the event's payload on stdin and the project from CLAUDE_PROJECT_DIR (else the payload's cwd), and appends one
bonsai.log/1 record to the session's file in the main checkout's .bonsai/local/log/: every free-text string redacted,
a tool call reduced to its target, and a prompt's words never kept. It prints nothing and exits 0 whatever it is
handed; a payload it cannot read, or a project with no bonsai.yaml, writes nothing.
`,
	FlagNext: "use the hook line bonsai init writes (bonsai hook record --help quotes it)",
	Exits: []Exit{
		{Code: 0, Means: "always: nothing is printed, and the session goes on"},
	},
	Examples: []string{"CLAUDE_PROJECT_DIR=. bonsai hook record < payload.json"},
	Run: func(c *call) int {
		return recorder.Record(recorder.Options{Stdin: os.Stdin, StdinTerminal: isTerminal(os.Stdin)})
	},
}
