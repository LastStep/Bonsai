---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5602
title: "`labels attach` and `detach` (5.6.2)"
status: done
lane:
done_when:
  - "labels attach and detach: the file checked, its namespace never a pack's or bonsai, a re-attach additive only; the refusal in an agent session"
  - "Go tests: attach, re-attach (unchanged, updated, a redefinition refused), detach; each namespace refusal; the variable set; what the attach then reaches (status --json's labels, check's values)"
  - "Human text in style C; check 10 both sides, the ladder, CI"
depends_on: [T-5600]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished: 2026-10-18
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5602
---

# `labels attach` and `detach` (5.6.2)

Piece 5.6.2 of step 5.6: its row in `design/plan-5.md`'s "#### The pieces and their order" and its note in "#### Notes
per piece" (lines 6494-6512). One of the four run side by side after 5.6.0 ("The order", step 2).

- **Owns:** `cmd/bonsai/labels.go` (the run functions); a new `internal/workspace/attach.go`.
- **Grants:** none while others run.
- **Proof:** the lines of `done_when`.
