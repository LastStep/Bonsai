---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-2026-10-23-T-5727
task: T-5727
role: builder
model: opus
started: 2026-10-23 07:36
finished: null
outcome: running
commits: ["083c0e6c", "7eced158", "e4fb2cfb", "3b2d6000", "e83f4abc"]
labels: {}
---

# Run: Other agents' markers refuse person-only commands

> Opened at the start of the session and added to as it goes. Not written at the end.

## What is different now

Bonsai's person-only commands (`update --allow-exec`, `unlink`, recording looser checks or another base branch,
`settings set`, `labels attach`/`detach`) now refuse when any of eight agent markers is set, not only Claude Code's:
`CLAUDE_CODE_CHILD_SESSION`, `AI_AGENT`, `PI_CODING_AGENT`, `CODEX_THREAD_ID`, `GEMINI_CLI`, `CURSOR_AGENT`,
`OPENCODE`, `GOOSE_TERMINAL`. The refusal names the marker it found and the agent ("not one pi started"). No y/N
question is asked with any of them set. `CLAUDECODE` alone counts for nothing. The guard's watch on a shell line now
refuses unsetting or assigning any of the eight by the exact forms it already read for Claude's (choice C); its
catch-all text match stays Claude's, so `cd opencode-tests` passes.

## Deviations from the plan

None from the audit as approved (M-2026-10-23-955578bb: go, watch C). Rule texts in `internal/guard/rules.go`
(person-command, shell-not-judged rows) and the guard's help in `cmd/bonsai/hook.go` were reworded to say "an agent
marker": both were named in the audit as staying Claude-only under A; under C they describe the wider watch, so they
changed with it (inside the guard files C added to Owns).

## Log

- 07:36 Brief M-2026-10-23-a75d000f read (orchestrator; this builder on Opus). Worktree ~/Servers/Bonsai-T-5727, branch
  T-5727 at 7e172040, clean. Common rules: ~/bonsai-checks/briefs/common-2026-10-23.md.
- 07:39 Audit (no code changed). `grep -rn CLAUDE_CODE_CHILD_SESSION cmd internal`, then each reader:
  - Person-only gates that refuse (get the list): `cmd/bonsai/agent.go:42` `agentSession()` (init/update
    `--allow-exec`, unlink, an update changing Bonsai's own lines, a new base branch, looser checks: all through
    `personCommand`); `agent.go:104` `personOnlyAt` (settings set, labels attach/detach); `cmd/bonsai/engine.go:34`
    `interactive` (no y/N question to an agent: a refusal to ask).
  - Claude-only, left as they are: `internal/guard/commands.go:193` the guard's preview rule (an allow, not a
    refusal: widening it would let the guard allow more); `internal/guard/shell_env.go`, `shell_words.go:581-611`,
    `rules.go:149,156` the unset-watch (Claude Code always sets CLAUDE_CODE_CHILD_SESSION in tools and hooks, so
    removing another marker leaves it set; the backstop's substring match on names like OPENCODE would cry wolf);
    `internal/format/check.go:133` HookEnvKeys (Claude settings env; a blank value there changed nothing);
    `internal/asks/asks.go:33`, `internal/ladder/ladder.go:123`, `internal/recorder/record.go:88` (session ids, log
    identity: not markers of an agent at all).
  - This shell carries CLAUDE_CODE_CHILD_SESSION=1, AI_AGENT=claude-code_2-1-309_agent, CLAUDECODE=1: every test that
    clears only Claude's marker to run as a person would flip in a local climb once AI_AGENT counts
    (`agent_test.go:24,463,467`, `hookline_test.go:251,255`, `settings_test.go:199`).
  - Sent as needs-input (Audit); waiting for go.
- 07:40 Go received, choice C (M-2026-10-23-955578bb). Built in order: the list, the cli, the test helpers, the
  guard.
- Fail-first, on 7e172040's code (a `git archive` copy in the scratchpad with the new test files dropped in):
  - `internal/guard/markers_test.go` TestEveryMarkerWatched: 16 forms x 7 markers = 112 subtests, **112 FAIL** on the
    old guard (the forms not refused, or a refusal naming no marker); TestOtherMarkersNoWolf 14 subtests pass on old
    and new (the cry-wolf guard: `cd opencode-tests`, `OPENCODE_HOME`, `unset CLAUDECODE`, reads).
  - `cmd/bonsai/markers_test.go` TestEveryMarkerRefusesPersonCommands: CLAUDE_CODE_CHILD_SESSION PASS, the other
    **7 FAIL**; TestEveryMarkerRefusesHomeCommands **FAIL**; TestClaudeCodeMarkerAloneRefusesNothing PASS (guard
    against adding it).
  - TestEveryMarkerIsAskedNothing (`askable`) and `internal/format` TestAgentMarker test new functions: they cannot
    compile on the old code, so not shown failing (said so in the audit).
- New code: all of the above pass. `go vet ./...` clean; `go test ./... -count=1` exit 0, run from this agent's shell
  (CLAUDE_CODE_CHILD_SESSION=1 and AI_AGENT set), so no test that runs as a person flipped.
- `go run ./internal/reference/gen <root>`: docs/reference/lists.md, 3 lines (person-command's error and the two rule
  rows); packs unchanged.
- 07:49 Five commits, 083c0e6c..e83f4abc. Climbing next.

## What did not work

- My first scratch copy command used `rm -rf $S/old` and the guard refused it (delete-unnamed: a variable in rm's
  operand). Rewritten with the path spelled out and no rm.
- `go build ./cmd/bonsai` for a compile check wrote `./bonsai` into the worktree (ignored by .gitignore); removed.
  `go vet` used after.

## Files changed

- `internal/format/markers.go` (new): AgentMarkers, the one list, and AgentMarker(getenv); `error.go`: person-command
  names the list; `docs/reference/lists.md` regenerated.
- `cmd/bonsai/agent.go`, `engine.go`: agentSession, personOnlyAt and interactive (via askable) read the list; the
  texts name the marker and its agent (byte-identical for Claude's). Help in init, update, unlink, settings, labels,
  hook.
- `internal/guard/shell_env.go`, `shell_words.go`: envMarker/envAssigned/envPSLeft/envProviderPath give the marker an
  exact form names; the backstop and name-as-data stay Claude's. Comments in shell.go, commands.go, guard.go; rule
  texts in rules.go.
- `internal/testpack/testpack.go`: Isolate clears every marker (AsPerson), PersonEnv for subprocesses; the test
  helpers in agent_test.go, hookline_test.go, settings_test.go use them. No assertion changed.
- `docs/features.md`: the person's commands row.

## Processes left running

Not yet checked (session running).
