---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5597
title: "CLAUDE.md's landing rule: pieces ready together land as one batch, with one Windows check at the batch's top"
status: done
lane:
done_when:
  - "CLAUDE.md's landing rule says that pieces ready together land as a batch: each rebased in turn onto the one before, each climbed at its own tip, check 10's Windows half run once at the batch's top commit, main fast-forwarded to it and pushed once, CI after it; and nothing else in CLAUDE.md changes"
  - "Changed by the orchestrator in the main checkout with no other task running, and landed as a records commit"
depends_on: [T-5590]
blocked_by:
created: 2026-10-15
started: 2026-10-15
finished: 2026-10-15
labels:
  bonsai.branch: null
  bonsai.ladder: []
  bonsai.allows: ["CLAUDE.md"]
  bonsai.wants: []
---

# CLAUDE.md's landing rule: land in batches

Rohan, 15 Oct, on a side agent's note that each landing made the other in-flight pieces rebase and redo their 35-60
minute Windows check (`base` ran it seven times, 574 minutes against 3-4.5 h): **"land in batches"**.

- **Owns:** `CLAUDE.md`'s landing rule only.
- **Grants:** `CLAUDE.md` (protected, not person-only): a change Rohan decided.
- **Proof:** the orchestrator's read of the diff; CI.
