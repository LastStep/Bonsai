---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6026
title: "Bugs stage 3: base:bug and base 1.1.0, the workflow pack's bugs kind out and v1.1.0, CHANGELOG 1.1.0"
status: todo
lane: null
done_when:
  - "The plan's steps 11 to 13 (up to the tags) built, in packs/base and in the workflow repo's local copy"
  - "A scratch project holding both drifting registers moved by base:bug's steps by a Sonnet runner, nothing dropped, check clean"
  - "The ladder and CI green; the packs' text read against the command as built"
depends_on: [T-6025]
blocked_by: null
created: 2026-11-19
started: null
finished: null
labels:
  bonsai.ladder: []
  bonsai.allows: ["packs/base/bonsai/pack.yaml"]
  workflow.model: sonnet
---

# Bugs stage 3: base:bug and base 1.1.0, the workflow pack's bugs kind out and v1.1.0, CHANGELOG 1.1.0

Part of T-6023, by its approved plan `records/plans/P-T-6023.md` (the director, 19 Nov: "Approve, start building").
