---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5691
title: "Bonsai's own packs moved: workflow to a42a4f4, base to main at 4b41790b (5.6's first ref moves)"
status: running
lane:
done_when:
  - "bonsai.yaml's refs: workflow a42a4f4 (5.6.7 and 5.6.10), base 4b41790b (5.6.9, 5.6.8's --explain lines, 5.6.0's settings and labels lines), on the director's yes (16 Oct)"
  - "bonsai update --json read line by line (no wall, deny rule, label definition or protected path removed), run with --yes, and everything it wrote committed in one commit"
  - "bonsai check clean; a climb green on main after it"
depends_on: [T-5600, T-5610]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: ["bonsai.yaml", ".bonsai/lock.json", ".claude/**", "CLAUDE.md"]
  bonsai.wants: ["bonsai.yaml", ".bonsai/lock.json", ".claude/**"]
---

# Bonsai's own packs moved (T-5691)

The ref moves of `design/plan-5.md`'s "Added 16 Oct, before the build" ("The ref moves"). The director's yes, 16 Oct:
"Yes, both", then "a42a4f4, the newest". Done in the main checkout by the orchestrator, as 5.5.7's `T-5507`.

- **Owns:** `bonsai.yaml` (the two `ref` lines), `.bonsai/lock.json`, `.claude/settings.json`, `CLAUDE.md`'s block.
- **Grants:** those; the person-only ones in `bonsai.wants`, granted by the director's yes.
- **Proof:** the lines of `done_when`.
