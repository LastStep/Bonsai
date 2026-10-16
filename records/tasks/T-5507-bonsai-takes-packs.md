---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5507
title: "Bonsai takes base and workflow: the new pre-release, Rohan's sitting, CLAUDE.md shrinks (5.5.7)"
status: done
lane:
done_when:
  - "The new pre-release built from a clean clone once every code fix has landed, the current one first kept in ~/bonsai-checks/prerelease/previous/, its install lines handed to Rohan in one batch"
  - "The preview read before Rohan's line: nothing of either pack under Runs code, Bonsai's two guard lines the only change to its own lines"
  - "bonsai.yaml gains base (on this section's word) and workflow (on Rohan's yes) and Bonsai's limits, in the main checkout with T-5507 the only task running; Rohan's bonsai update --allow-exec --yes; the link committed and pushed, CI read"
  - "check has no finding; a green climb of T-5507 at the link's commit; status --json shows the four labels, workflow's lanes and labels and the roles' limits in force"
  - "An auditor subagent started in Bonsai's repo meets a wall's refusal, its call refused and its actions counted"
  - "CLAUDE.md shrinks, every line's new home audited by a Sonnet agent; D2 written in; STATE.md moved to .bonsai/STATE.md; the end verifier's checks on the real repo"
depends_on: [T-5506]
blocked_by:
created: 2026-10-14
started: 2026-10-16
finished: 2026-10-16
labels:
  bonsai.ladder: []
  bonsai.allows: ["bonsai.yaml", ".bonsai/lock.json", ".claude/**", "CLAUDE.md", ".bonsai/.gitignore"]
  bonsai.wants: ["bonsai.yaml", ".bonsai/lock.json", ".claude/**"]
---

# Bonsai takes base and workflow: the new pre-release, Rohan's sitting, CLAUDE.md shrinks (5.5.7)

Piece 5.5.7 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece". Rohan's decision 6 ("option A obviously"); `workflow` only on his yes.

- **Owns:** `bonsai.yaml`, `.bonsai/lock.json`, `.claude/settings.json` and `.bonsai/.gitignore` (by Rohan's update), the block in `CLAUDE.md` and `workflow`'s session-start file under `records/protocols/` (by the update), `CLAUDE.md`, `STATE.md` moved to `.bonsai/STATE.md`, the two tables (by `check --write`).
- **Grants:** `bonsai.yaml`, `.bonsai/lock.json`, `.claude/**`, `CLAUDE.md` and `.bonsai/.gitignore`, the person-only ones also in `bonsai.wants`: a change this section, which Rohan approved, names (the stricter-only rule).
- **Proof:** the lines of `done_when`; the end verifier.
