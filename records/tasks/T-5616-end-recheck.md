---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5616
title: "5.6's end verifier, the short re-check of checks 8, 11, 14, 15 and 17 and the director's Windows half"
status: done
lane:
done_when:
  - "each item T-5611 (M-2026-10-19-c7306085) left open on checks 8, 14, 15 and 17, and the runbook steps it left to the director (5, 7, 10), read against what closes it now, PASS or FAIL each with evidence"
  - "every landing since b3cb8641 matched to a green local ladder record naming the installed bonsai with its sha256; every task's bonsai.allows read against its Owns"
  - "a verdict on 5.6 as a whole"
depends_on: [T-5611, T-5612, T-5613, T-5614, T-5615]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# 5.6's end verifier, the short re-check (T-5616)

T-5611 failed 5.6 on records and the director's questions, not on code. Since then: T-5612 (check 15's 11 assertions),
the director's answers on checks 14 (b) and 15 (a), T-5613 (step 7a-7d in CI; 7e to 5.7, the director's word, 19 Oct:
this computer has no Windows Sandbox), T-5614 (the director's steps 4, 5, 10 and check 8's WSL half), T-5615 (the
tally, the sessions table, correct's first run).

- **Owns:** nothing; a fresh verifier reads and runs, and changes nothing.
- **Grants:** none.
