---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5490
title: The end verification of 5.4, a fresh verifier runs the 5.4 done checks 16 to 23
status: running
lane:
done_when:
  - "A fresh Opus verifier runs plan-5's 5.4 done checks 16 to 23 on the final commit and the installed bonsai, and passes or fails 5.4; its report is in this task's run report"
  - "The installed bonsai's ladder --task T-5490 is green on the final commit (check 22)"
  - "Its findings are noted, not fixed, by Rohan's word of 13 Oct, unless he says otherwise"
depends_on: [T-5407]
blocked_by:
created: 2026-10-13
started: 2026-10-13
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# The end verification of 5.4

Part 5.4's work outside its pieces, in `design/plan-5.md`, step 5.4: "5.4 done", checks 16 to 23 (checks 1 to 15 are
the verifiers' reports V1 to V2d, in `records/runs/R-2026-10-13-5.4-verify.md`). The task the end verifier's session
runs under; it grants nothing.

- **Owns:** nothing in the repo; the verifier's report goes in this task's run report, written by the orchestrator.
- **Grants:** none: `bonsai.allows` is empty.
- **Proof:** the verifier's report, and `/usr/local/bin/bonsai ladder --task T-5490` green on the final commit.
