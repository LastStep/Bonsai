---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5714
title: "Time, piece C: printed times in local time with their offset"
status: done
lane:
done_when:
  - "every printed time in the inventory's part 3 through timefmt.Show, --utc on each command, --json unchanged (plan step 3)"
  - "bonsai logs --day as a local day, filtered to its instants; message --show's trust line; the SessionStart lines"
  - "climbed green; landed by CI on the branch"
depends_on: [T-5713]
blocked_by:
created: 2026-10-21
started: 2026-10-21
finished: 2026-10-21
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.branch: T-5714
  workflow.model: sonnet
---

# Time, piece C (T-5714)

A piece of `T-5795`, planned in `records/plans/P-T-5795.md` (approved 21 Oct by the director): read its step for
piece C, its parallel-work checkpoint and its proof table; the contract is `design/contract.md` section 23
"Time". Evidence: `~/bonsai-checks/T-5795/inventory.md`, `~/bonsai-checks/T-5795/research.md`.
