---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5611
title: "5.6's end verifier (T-5690-T-5699 being full, the next free number after 5.6's pieces)"
status: running
lane:
done_when:
  - "a fresh Opus verifier runs plan-5.md's \"5.6 done\" checks 1-17 itself on main a4dfa5ed, both sides where a check names them, and passes or fails 5.6 with evidence"
  - "the director's runbook steps 4, 5, 7 and 10 (~/bonsai-checks/end-5.6/runbook.md) named open, not counted as passed"
  - "a climb of main by the installed ladder under T-5611 green"
depends_on: [T-5695, T-5699]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# 5.6's end verifier (T-5611)

plan-5.md, step 5.6, "5.6 done". The verifier and runner roles cannot make scratch repositories (BUG-013): the
verifier writes the setup it needs as scripts and asks; the orchestrator or a builder runs them.

- **Owns:** nothing in the repository.
- **Grants:** none.
