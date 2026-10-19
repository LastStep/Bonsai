---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5695
title: "5.6's floors: go_tests and go_tests_tagged raised to main's climb"
status: done
lane:
done_when:
  - "main climbed once by the installed bonsai's ladder --task T-5695, green, clean, at main's top"
  - "bonsai.yaml's ratchets raised to that climb's two counts (stricter only), committed on main; a climb after it green"
depends_on: [T-5700]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: ["bonsai.yaml"]
  bonsai.wants: []
---

# 5.6's floors (T-5695)

CLAUDE.md, "The floors, once a part": the orchestrator climbs `main` once and raises the floors to that climb's counts
under a task granting `bonsai.yaml` (a stricter change, no person needed). Before the 5.6 end verifier (plan-5.md, "5.6
done": "after 5.6.5 and the floors").

- **Owns:** `bonsai.yaml`'s `ratchets` counts only.
- **Grants:** `bonsai.yaml`.
