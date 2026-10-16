---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5521
title: "5.5's floors raised to the last climb's counts"
status: running
lane:
done_when:
  - "After 5.5's last records commit, main climbed once under T-5507 by the installed bonsai, green"
  - "bonsai.yaml's ratchets go_tests and go_tests_tagged raised to that climb's counts (stricter only), committed on main with this task running alone; the Bless numbers in the run report"
  - "A climb of T-5521 at the floors' commit green; CI green"
depends_on: [T-5507]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: ["bonsai.yaml"]
  bonsai.wants: ["bonsai.yaml"]
---

# 5.5's floors raised to the last climb's counts

`CLAUDE.md`, "The floors, once a part": after the part's last records commit, the orchestrator climbs `main` once under
the part's last task and raises the floors to that climb's counts under a task granting `bonsai.yaml`. The
outside-piece numbers `T-5590`-`T-5599` are all taken, so it takes the next free number.

- **Owns:** `bonsai.yaml`'s `ratchets` only.
- **Grants:** `bonsai.yaml` (a stricter change: floors raised).
- **Proof:** the lines of `done_when`.
