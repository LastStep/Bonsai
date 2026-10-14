---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5590
title: Step 5.5 re-planned on the agent-pipelines redesign and Rohan's six decisions of 14 Oct
status: verify
lane:
done_when:
  - "design/plan-5.md's Step 5.5 section re-planned on Rohan's 14 Oct decisions (six roles, the new message format, the guard's role rules and stall stops with budgets under the pack's entry in bonsai.yaml, the facts lens, Bonsai linking workflow) and 5.4's carry-overs (the two CI flakes in the first piece, the runner clearing BONSAI_TASK and BONSAI_ROLE, D2, O1), with its pieces, order, Owns, proof and hours re-costed"
  - "A fresh Opus reviewer's report on the section: ready, or ready after fixes with every blocking finding fixed"
  - "Rohan's approval of the re-planned section under his (B), with its hours and re-ask line"
  - "Only design/plan-5.md changes on the branch; landed by the landing rule; CI after the push (the two known flakes aside, fixed in 5.5's first piece)"
depends_on: [T-5491]
blocked_by:
created: 2026-10-14
started: 2026-10-14
finished:
labels:
  bonsai.branch: "plan5.5r"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# Step 5.5 re-planned on the agent-pipelines redesign

Part 5.5's planning: its approved section (9 Oct) re-planned before any piece is built, on Rohan's 12 Oct
agent-pipelines redesign, his 13 Oct stall stops and his six decisions of 14 Oct (the run report holds them, with his
words). The section comes back to him under his (B): it changes the hours, his steps and a format.

- **Owns:** `design/plan-5.md`, its Step 5.5 section; nothing else.
- **Grants:** none: `design/plan-5.md` is not protected, so `bonsai.allows` is empty. Dated notes the re-plan needs in
  the spec or the contract (both protected) are placed in a piece that grants them, not made here.
- **Proof:** a fresh Opus reviewer's report and its fixes; Rohan's approval; the landing rule.
