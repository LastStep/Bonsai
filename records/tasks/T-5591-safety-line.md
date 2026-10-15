---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5591
title: CLAUDE.md's Safety line names the workflow pack's repository among the places Bonsai's work writes
status: done
lane:
done_when:
  - "CLAUDE.md's Safety line names LastStep/bonsai-workflow (its local copy in the scratch folder, then GitHub from P on) among the places Bonsai's work writes, and nothing else in CLAUDE.md changes"
  - "Changed by the orchestrator in the main checkout with no other task running, before T-5502 and T-5504 start, and landed as a records commit"
depends_on: [T-5590]
blocked_by:
created: 2026-10-14
started: 2026-10-15
finished: 2026-10-15
labels:
  bonsai.branch: null
  bonsai.ladder: []
  bonsai.allows: ["CLAUDE.md"]
  bonsai.wants: []
---

# CLAUDE.md's Safety line names the workflow pack's repository among the places Bonsai's work writes

Part 5.5's work outside its pieces, named in the section's order, item 4: before `base` and `workflow` start, Bonsai's work also writes the workflow pack's repository.

- **Owns:** `CLAUDE.md`'s Safety line only.
- **Grants:** `CLAUDE.md` (protected, not person-only): a change the approved section names.
- **Proof:** the orchestrator's read of the one-line diff; CI.
