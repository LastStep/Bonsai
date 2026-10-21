---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5715
title: "Time, piece D: the sessions table: new rows with an offset, compared by instant"
status: approved
lane:
done_when:
  - "new rows in timefmt.Doc, old rows untouched; Key, Missing, the sort and clean/rows.go compare by instant (plan step 4)"
  - "a mixed-table test: no row re-added, order by instant, check clean after check --write"
  - "climbed green; landed by CI on the branch"
depends_on: [T-5713]
blocked_by:
created: 2026-10-21
started:
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.branch: T-5715
  workflow.model: sonnet
---

# Time, piece D (T-5715)

A piece of `T-5795`, planned in `records/plans/P-T-5795.md` (approved 21 Oct by the director): read its step for
piece D, its parallel-work checkpoint and its proof table; the contract is `design/contract.md` section 23
"Time". Evidence: `~/bonsai-checks/T-5795/inventory.md`, `~/bonsai-checks/T-5795/research.md`.
