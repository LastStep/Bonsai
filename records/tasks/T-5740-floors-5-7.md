---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5740
title: "5.7's floors: go_tests and go_tests_tagged raised to main's climb by the installed 1.0"
status: running
lane:
done_when:
  - "main climbed once by the installed bonsai 1.0's ladder --task T-5740, green, clean, at main's top"
  - "bonsai.yaml's ratchets raised to that climb's two counts (stricter only), committed on main; a climb after it green"
depends_on: [T-5706]
blocked_by:
created: 2026-10-25
started: 2026-10-25
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: ["bonsai.yaml"]
  bonsai.wants: []
  workflow.model: opus
---

# 5.7's floors (T-5740)

CLAUDE.md, "The floors, once a part": the orchestrator climbs `main` once and raises the floors to that climb's counts
under a task granting `bonsai.yaml` (a stricter change, no person needed); plan-5.md's order for 5.7's end: the floors
(climbed by the installed 1.0), the Haiku audit of `.bonsai/sessions.md`, then the 5.7 end verifier. T-5790 to T-5799
are all taken, so this takes the next free number.

- **Owns:** `bonsai.yaml`'s `ratchets` counts only.
- **Grants:** `bonsai.yaml`.
