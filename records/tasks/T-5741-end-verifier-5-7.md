---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5741
title: "5.7's end verifier: plan-5.md's \"5.7 done\" checks 1-17 on the final commit and the published 1.0, both sides"
status: running
lane:
done_when:
  - "a fresh Opus verifier runs plan-5.md's \"#### 5.7 done\" checks 1-17 itself on main's final commit and on the published 1.0, on both sides where a check names them, and passes 5.7 (or each blocking finding is fixed and re-checked, or put after 1.0 on the director's word)"
  - "the Haiku audit of .bonsai/sessions.md against 5.7's run reports (check 17), its report read"
  - "every landed commit of 5.7 matched to a green local ladder record naming the installed bonsai with that result's sha256; every 5.7 task's bonsai.allows read against its section's Owns; every bonsai.yaml change in 5.7 read against CLAUDE.md's stricter-only rule"
depends_on: [T-5740]
blocked_by:
created: 2026-10-25
started: 2026-10-25
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# 5.7's end verifier (T-5741)

plan-5.md's "#### 5.7 done": "A fresh Opus verifier, at the end of 5.7, after 5.7.6, the pack tags and the floors, runs
each check itself on the final commit and on the published 1.0, on both sides where a check names them, and passes or
fails 5.7". Then `.bonsai/STATE.md` rewritten and the director's roadmap updated: Bonsai 1.0.
