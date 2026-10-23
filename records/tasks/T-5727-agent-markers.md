---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-5727
title: "Other agents' markers refuse person-only commands"
status: done
lane: null
done_when:
  - "with any one of CLAUDE_CODE_CHILD_SESSION, AI_AGENT, PI_CODING_AGENT, CODEX_THREAD_ID, GEMINI_CLI, CURSOR_AGENT, OPENCODE, GOOSE_TERMINAL set, every command Claude's child marker refuses today is refused, by a test per marker"
  - "the refusal names the marker it found"
  - "CLAUDECODE alone refuses nothing; the test helper clears every marker"
  - "the climb at the branch's HEAD green, CI on the branch green"
depends_on: []
blocked_by: null
created: 2026-10-23
started: 2026-10-23
finished: 2026-10-23
labels:
  bonsai.branch: T-5727
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# Other agents' markers refuse person-only commands

## Context

T-5710 item 13; the director, 23 Oct, C4 and C8 of `design/agent-neutral-walkthrough.md`. Today
`CLAUDE_CODE_CHILD_SESSION` refuses `update --yes --allow-exec`, `unlink --yes` and accepting looser settings
(`cmd/bonsai/agent.go` and its callers). Before 1.0, since pi runs with a stopgap extension before then.
