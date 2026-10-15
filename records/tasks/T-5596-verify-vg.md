---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5596
title: "VG: a fresh verifier on the guard's role rules and stall stops, bonsai agents, and 5.5.8's gate and runner lines (5.5.11, 5.5.12, 5.5.8)"
status: blocked
lane:
done_when:
  - "A fresh Opus verifier, its bar adversarial (agents trying to break the system on purpose), reads note 5.5.11, 5.5.12 and 5.5.8 items 4 and 5, the section's rules and the diffs, re-runs check 10 itself on both sides, runs 5.5 done's checks 1 (the gate and runner lines) and 5 and check 15 for their refusals, and breaks the rules on purpose"
  - "Its verdict, PASS or FAIL for each of 5.5.11, 5.5.12 and 5.5.8's lines, with evidence; each must-fix fixed forward and passed by the same verifier, at most two rounds a piece"
depends_on: [T-5511, T-5512, T-5508]
blocked_by: "Waits for VG's fix round (A-F, H) before its second round."
created: 2026-10-15
started: 2026-10-15
finished:
labels:
  bonsai.branch: null
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# VG: a fresh verifier on 5.5.11, 5.5.12 and 5.5.8's gate and runner lines

Part 5.5's work outside its pieces: the section's verifier VG ("after 5.5.12: the guard's role rules and stall stops,
the stop gate's budget exit, 5.5.12's view of the guard's state, and 5.5.8's gate and runner lines: guard code, a
verifier of its own"). It fixes nothing; a must-fix goes back as a fix round.

- **Owns:** no repository file: its scratch in `~/bonsai-checks/verify-5.5-VG/`.
- **Grants:** none.
- **Proof:** its report, in this task's run report.
