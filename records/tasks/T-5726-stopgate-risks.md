---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-5726
title: "The stop gate's three risks from the trial's verifier"
status: done
lane: null
done_when:
  - "every stop-gate rule word passes the same word filter as the rest of a block's text, or a test proves every rule constant is safe, failing on an unsafe constant"
  - "base's operating-bonsai skill says correctly which blocks the log keeps and which it does not"
  - "bonsai hook stop given a bad argument refuses through the stop gate's frame with its rule line, by test"
  - "the climb at the branch's HEAD green, CI on the branch green"
depends_on: []
blocked_by: null
created: 2026-10-23
started: 2026-10-23
finished: 2026-10-23
labels:
  bonsai.branch: T-5726
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# The stop gate's three risks from the trial's verifier

## Context

T-5710 item 6, the trial verifier's risks (`M-2026-10-20-57b77ed9`, "Risks worth knowing"): the rule word skips
`safe()`; SKILL.md's "the one the log keeps" (no-project and an early over-time keep no record); `cmd/bonsai/hook.go`
~144's flag-error path refuses without the frame.
