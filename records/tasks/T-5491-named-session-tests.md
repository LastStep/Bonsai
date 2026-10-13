---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5491
title: Tests that read the caller's BONSAI_TASK set their own, so a named session's climb is green (D1)
status: done
lane:
done_when:
  - "internal/status's TestStatusLadder and cmd/bonsai's TestAgentUnderAPersonsGrant set their own BONSAI_TASK and BONSAI_ROLE, and pass with the caller's set to any task"
  - "go test ./... plain and tagged, with BONSAI_TASK=T-5491 and BONSAI_ROLE=builder set, has no failure, and as many pass as with them unset"
  - "The installed bonsai's ladder --task T-5491 is green at the branch's HEAD, and green again when the climb's own environment names BONSAI_TASK=T-5491"
  - "Only test files change; landed by the landing rule; CI green"
depends_on: [T-5490]
blocked_by:
created: 2026-10-13
started: 2026-10-13
finished: 2026-10-13
labels:
  bonsai.branch: "t5491"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# Tests that read the caller's BONSAI_TASK set their own (D1)

Part 5.4's work outside its pieces: the 5.4 end verifier's D1 (`records/runs/R-2026-10-13-T-5490-verify-5.4-end.md`,
13:31), fixed on Rohan's word ("Fix now"). Two tests read the `BONSAI_TASK` of whoever runs them, so a climb from a
named session (one started with `BONSAI_TASK=<id>`) is red on rungs 3 and 4 and the stop gate never lets it end.

- **Owns:** the `_test.go` files that read the caller's `BONSAI_TASK` or `BONSAI_ROLE`; nothing else.
- **Grants:** none: test files are not protected, so `bonsai.allows` is empty.
- **Proof:** the lines of `done_when`; the orchestrator's read of the diff, green tests and CI (a test-only change).
