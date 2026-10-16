---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5522
title: "5.5's end verifier: \"5.5 done\", checks 1 to 17"
status: blocked
lane:
done_when:
  - "A fresh Opus verifier (workflow:verifier) runs design/plan-5.md's \"5.5 done\" checks 1 to 17 itself on the final commit, on both sides where a check names them, and passes or fails 5.5"
  - "Its report saved and read; each must-fix fixed forward and re-run by the same verifier (two rounds at most, a third only if necessary)"
  - "Then T-5500 to T-5522 done, STATE rewritten for 5.6, and the director given 5.5's close with the floors"
depends_on: [T-5507, T-5521]
blocked_by: "Paused while T-5507 changes CLAUDE.md (the end verifier's must-fix 2)."
created: 2026-10-16
started: 2026-10-16
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# 5.5's end verifier

`design/plan-5.md`, "#### 5.5 done" (from 5726): a fresh Opus verifier after 5.5.7 and the floors. It changes nothing
in the repository: its scratch is under `~/bonsai-checks/verify-5.5-end/`.

- **Owns:** no repository file.
- **Grants:** none.
- **Proof:** the lines of `done_when`.
