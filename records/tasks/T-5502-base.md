---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5502
title: "base, Bonsai's own pack: the walls, the labels, the block and the template skills (5.5.2)"
status: running
lane:
done_when:
  - "base in packs/base/: the plugin manifest; pack.yaml with the walls, each with its why, and base's protected paths; labels.yaml with contract section 5.6's four labels; block.md; README; the template skills"
  - "check --pack and claude plugin validate; Bonsai's packs job; packs/packs_test.go"
  - "A scratch project linked to base with --yes alone (nothing under Runs code), its block and deny rules read back"
  - "P0's privacy read of the task and run-report templates (after a Haiku grep) before it lands; the orchestrator's read"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; V1 passes"
depends_on: [T-5500, T-5501, T-5591]
blocked_by:
created: 2026-10-14
started: 2026-10-15
finished:
labels:
  bonsai.branch: "t5502"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# base, Bonsai's own pack: the walls, the labels, the block and the template skills (5.5.2)

Piece 5.5.2 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece".

- **Owns:** `packs/base/` but 5.5.0's three generated skill folders and `skills/operating-bonsai/`; `packs/packs_test.go`.
- **Grants:** none: no protected path among its files, so `bonsai.allows` is empty. Its files are LF through the root's `.gitattributes` line (5.5.0).
- **Proof:** the lines of `done_when`; the installed `bonsai`'s ladder green at the branch's HEAD; the landing rule; CI. P0; V1.
