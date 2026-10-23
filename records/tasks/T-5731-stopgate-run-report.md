---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-5731
title: "The stop gate reads the run report as bonsai check does (the pi brief's B3)"
status: done
lane: null
done_when:
  - "a named session whose run report bonsai check rejects (no frontmatter; a started without a time; any finding check makes on the report file itself) is held by the stop gate, with a rule of its own and a next step naming the field, by tests shown failing on the old code"
  - "a run report check accepts lets the gate go as before; every existing stop-gate test green unchanged"
  - "base's operating skill and docs/features.md say it; the reference page regenerated"
  - "the climb at the branch's HEAD green, CI on the branch green; a fresh verifier passes it"
depends_on: [T-5730]
blocked_by: null
created: 2026-10-23
started: 2026-10-23
finished: 2026-10-23
labels:
  bonsai.branch: T-5731
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# The stop gate reads the run report as bonsai check does

## Context

`design/pi-bonsai-work.md` (the pi side's brief, 23 Oct), B3: the gate only checks that `R-<date>-<task>.md`
exists (`internal/stopgate/stopgate.go:50`, `stop-no-run-report`); in the pi trial it twice let go on a run report
`bonsai check` then rejected. The same under Claude Code. The director, 23 Oct: "B3 only" before 1.0, for rc.3.

## Out of scope

B1, B4, B5 (after 1.0); whether the gate wants the ladder's rungs to cover `done_when` (a design question for the
director, not asked).
