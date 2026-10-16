---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5593
title: "VM: a fresh verifier on formats set 8, the engine's roles and limits, and messages between agents (5.5.9, 5.5.0, 5.5.10)"
status: done
lane:
done_when:
  - "A fresh Opus verifier, its bar adversarial, reads note 5.5.9, 5.5.0 and 5.5.10, the section's shapes and the three diffs, re-runs check 10 itself on both sides, and runs 5.5 done's checks 2, 3 (roles and limits) and 4, and check 15 for their refusals"
  - "Its verdict, PASS or FAIL for each of 5.5.9, 5.5.0 and 5.5.10, with evidence; each must-fix fixed forward by the piece's builder and passed by the same verifier, at most two rounds a piece"
depends_on: [T-5509, T-5500, T-5510]
blocked_by:
created: 2026-10-15
started: 2026-10-15
finished: 2026-10-15
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# VM: a fresh verifier on 5.5.9, 5.5.0 and 5.5.10

Part 5.5's work outside its pieces: the section's verifier VM ("after 5.5.10: the formats set, 5.5.0's roles and
budgets, the message reader and its commands; `CLAUDE.md`'s rule for a reader"). It fixes nothing; a must-fix goes back
to the piece's builder as a fix round.

- **Owns:** no repository file: its scratch in `~/bonsai-checks/verify-5.5-VM/`.
- **Grants:** none.
- **Proof:** its report, in this task's run report.
