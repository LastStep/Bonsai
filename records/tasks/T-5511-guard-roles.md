---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5511
title: The guard's role rules and stall stops (5.5.11)
status: done
lane:
done_when:
  - "The hook payloads measured first, on every route an agent can be started by (note 5.5.11, 1), written in the run report"
  - "In a project whose packs declare roles, the guard's two lines hold who may call whom, the no-change roles, writes into Bonsai's records and the lock refused, a --from naming another address refused, and the stall budgets, with Go tests on the measured payloads and the budget table as the section's proof lists"
  - "A live lock edit changes no rule; the second line lets reads through with bonsai missing and refuses past a budget with it present; the stop gate's exit only after a report"
  - "Scripted sessions on both sides with the stamped build: a lead calling a helper, a helper's call refused, a budget spent, a spent helper re-briefed and working again; the guard's added cost per call measured on both sides"
  - "Its last commit (the brief's budget through internal/message, bonsai messages' agent column, the agents kind in the cleaning, check's own-hooks reading) after T-5510 has landed; docs/reference/lists.md regenerated"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; VG's break-it passes"
depends_on: [T-5500, T-5508]
blocked_by:
created: 2026-10-14
started: 2026-10-15
finished: 2026-10-15
labels:
  bonsai.branch: "t5511"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# The guard's role rules and stall stops (5.5.11)

Piece 5.5.11 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece". The rules are the section's "Role rules and stall stops, written once". Guard code: an Opus builder and its own verifier, VG.

- **Owns:** `internal/guard/` (the role rules, the no-change roles, the budgets, the shell judge's new words, its rows in `guard.Rules`); `internal/agents/` (new: each agent's state and parent link); the `spawn` record's writer; `internal/hookin/`; the hook lines where the engine writes them (`engine.GuardTools`, the second line, `internal/engine/hookline.go` and `settings.go`) and the second line's word on `bonsai hook guard`; the budget exit and its `stop-budget` row in `internal/stopgate/`; `docs/reference/lists.md`; its last commit after 5.5.10 lands, as the section's Owns says.
- **Grants:** none: no protected path among its files, so `bonsai.allows` is empty. The hook lines change in the engine's code, not in this repo's `.claude/settings.json` (that is 5.5.7's, by Rohan's update).
- **Proof:** the lines of `done_when`; the installed `bonsai`'s ladder green at the branch's HEAD; the landing rule; CI. VG.
