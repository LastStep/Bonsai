---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5713
title: "Time, piece B: bonsai now and internal/timefmt"
status: approved
lane:
done_when:
  - "internal/timefmt as plan P-T-5795 step 2 says, every function taking a *time.Location, tests with injected locations (Asia/Kolkata, UTC, a DST zone) via time/tzdata in tests only"
  - "bonsai now [--utc|--day] [--json], its --json held to bonsai.now/1; help and the reference page"
  - "climbed green at its tip, rebased on T-5712; landed by CI on the branch (windows jobs green)"
depends_on: [T-5712]
blocked_by:
created: 2026-10-21
started:
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.branch: T-5713
  workflow.model: sonnet
---

# Time, piece B (T-5713)

A piece of `T-5795`, planned in `records/plans/P-T-5795.md` (approved 21 Oct by the director): read its step for
piece B, its parallel-work checkpoint and its proof table; the contract is `design/contract.md` section 23
"Time". Evidence: `~/bonsai-checks/T-5795/inventory.md`, `~/bonsai-checks/T-5795/research.md`.
