---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6003
title: "Five small open bugs, built beside T-6002 and landed in its batch: BUG-091, BUG-112, BUG-114's next step, BUG-022, BUG-113"
status: done
lane: light
done_when:
  - "BUG-091: every next-step check in cmd/bonsai/engine_test.go that used HasSuffix(nextOf(out), ...) holds the end with endsWith, and each still passes"
  - "BUG-112: TestDecideReadsDevicePaths's setup fails (or skips) with its own words when QueryDosDevice gives an empty device, instead of building a path with no device"
  - "BUG-114 (its second half only): config-uncommitted's first-commit next step (UncommittedInitNext) names the exact command, git add -A && git commit -m \"link Bonsai\", with the tests that read it updated"
  - "BUG-022: read; closed in records/bugs.md if the personal-copy warning covers it, else its row says what is missing"
  - "BUG-113: README's first-project steps 5 and 6 agree: the small task's rung passes on what step 5 asks for, and step 6 says the task must read running"
  - "A green climb of the branch, and CI green in T-6002's batch"
depends_on: []
blocked_by: null
created: 2026-10-26
started: 2026-10-26
finished: 2026-10-26
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: t-6003-small-fixes
  workflow.model: sonnet
  workflow.estimate_h: 1.5
---

# Five small open bugs, beside T-6002

## Context

The director (26 Oct), approving T-6002: "also check if you can club fix for any of the known bugs on bonsai, which
does not add to the overall time here". The orchestrator chose open bugs in `records/bugs.md` that touch no protected
path, no guard decision and no format, each under an hour, built by a second builder at the same time as T-6002 and
landed in its batch (one CI run).

## Scope

`cmd/bonsai/engine_test.go`, `internal/guard/guard_windows_test.go`, `internal/engine/checkcommitted.go` and the tests
that read its next step, `README.md`, `records/bugs.md` (rows for these five only, through the orchestrator). Nothing
else; a fix that grows past this scope stops and reports.
