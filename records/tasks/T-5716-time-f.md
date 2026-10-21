---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5716
title: "Time, piece F: the guard checks fresh stamps against Bonsai's clock"
status: done
lane:
done_when:
  - "the check before the tripwire's status_writes gate (internal/guard/tripwire.go:178), in both modes, over the fields plan step 6 lists, memory added to documentOf; rule time-not-now with next bonsai now"
  - "check flags a zone-less time only in a field an uncommitted change wrote; base:task, base:run, base:state, base:memory updated"
  - "guard tests as the plan's proof table lists, and a scripted session (a runner) refused on a guessed stamp then accepted from bonsai now; climbed green; landed by CI on the branch"
depends_on: [T-5713]
blocked_by:
created: 2026-10-21
started: 2026-10-21
finished: 2026-10-21
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.branch: T-5716
  workflow.model: opus
---

# Time, piece F (T-5716)

A piece of `T-5795`, planned in `records/plans/P-T-5795.md` (approved 21 Oct by the director): read its step for
piece F, its parallel-work checkpoint and its proof table; the contract is `design/contract.md` section 23
"Time". Evidence: `~/bonsai-checks/T-5795/inventory.md`, `~/bonsai-checks/T-5795/research.md`.
