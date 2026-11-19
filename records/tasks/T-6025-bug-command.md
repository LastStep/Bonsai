---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6025
title: "Bugs stage 2: the bonsai bug command, its log records, the guard's hold on bug status lines and the marked page, the operating-skill line"
status: done
lane: null
done_when:
  - "The plan's steps 6 to 10 built on a branch from stage 1's, each test the plan names for them green, on Linux and Windows"
  - "The ladder green at the branch's HEAD by the installed bonsai, CI green on the branch, both windows jobs; a fresh Opus verifier's PASS (tests and correctness, security)"
depends_on: [T-6024]
blocked_by: null
created: 2026-11-19
started: 2026-11-19
finished: 2026-11-19
labels:
  bonsai.ladder: []
  bonsai.allows: ["design/contract.md"]
  bonsai.branch: t-6025-bug-command
  workflow.model: opus
---

# Bugs stage 2: the bonsai bug command, its log records, the guard's hold on bug status lines and the marked page, the operating-skill line

Part of T-6023, by its approved plan `records/plans/P-T-6023.md` (the director, 19 Nov: "Approve, start building").
