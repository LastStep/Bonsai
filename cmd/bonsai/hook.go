package main

// bonsai hook <name>: the hook entry point (spec §4, §7). This build answers `bonsai hook guard` (plan part 5,
// internal/guard); start, stop and record come with steps 5.2 and 5.3, and part 3 writes no line that calls them.

import (
	"fmt"
	"io"
	"os"

	"github.com/LastStep/Bonsai/internal/guard"
)

const hookUsage = `bonsai hook <name>: the hook entry point Claude Code's hook lines call (spec section 7).
This build answers:
  bonsai hook guard     the PreToolUse guard (bonsai hook guard --help)
Not built yet: start, stop, record.
Exit codes: 0 allow, 2 block. A hook speaks Claude Code's hook format: it takes no --json.
Example: bonsai hook guard --help
`

const guardUsage = `bonsai hook guard: the PreToolUse guard that bonsai init writes as the hook line
  bonsai hook guard || exit 2
on PreToolUse (Edit, Write, MultiEdit, NotebookEdit, Bash, PowerShell), timeout 10 s.
It reads Claude Code's PreToolUse payload on stdin and the project from CLAUDE_PROJECT_DIR, and refuses an Edit,
Write, MultiEdit or NotebookEdit of a path on bonsai.yaml's protected or person_only list. Bash and PowerShell are
allowed in this build. It blocks whatever it cannot read or decide, and blocks itself after 5 s, before the hook's
timeout could let the call through. In a project with no bonsai.yaml it allows and records nothing.
Each decision is a guard record (bonsai.log/1) in .bonsai/local/log/, naming the binary's own path, and on a
session's first record its SHA-256.
Exit codes:
  0  allow: the call goes on (nothing printed)
  2  block: the reason and the next step on stderr, which Claude Code shows to the agent
Example: CLAUDE_PROJECT_DIR=. bonsai hook guard < payload.json
`

// runHook answers bonsai hook.
func runHook(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		return write(stdout, hookUsage)
	}
	if len(args) == 0 {
		return refuse(stderr, "bonsai hook needs a hook name", "run `bonsai hook --help`")
	}
	switch args[0] {
	case "guard":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			return write(stdout, guardUsage)
		}
		if len(args) > 1 {
			return refuse(stderr, fmt.Sprintf("hook guard takes no %+q", args[1]),
				"use the hook line bonsai init writes: bonsai hook guard || exit 2")
		}
		return guard.Main(guard.Options{Stdin: os.Stdin, Stderr: stderr, StdinTerminal: isTerminal(os.Stdin)})
	case "start", "stop", "record":
		return refuse(stderr, "bonsai hook "+args[0]+" is not built yet (it comes with a later part of the rebuild)",
			"take that hook line out of .claude/settings.json; bonsai init writes only the guard's line in this build")
	}
	return refuse(stderr, fmt.Sprintf("%+q is not a hook", args[0]), "run `bonsai hook --help`")
}
