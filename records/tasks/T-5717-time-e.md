---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5717
title: "Time, piece E: now told on each prompt and each subagent start"
status: approved
lane:
done_when:
  - "a synchronous bonsai hook now on UserPromptSubmit and, by JSON additionalContext, SubagentStart, never blocking (plan step 5)"
  - "scripted sessions (a runner) show both reach the model; bonsai missing lets the prompt run; the cost measured on both sides, under twice the guard's call"
  - "climbed green; landed by CI on the branch; then the director's bonsai update --allow-exec --yes in the main checkout"
depends_on: [T-5716]
blocked_by:
created: 2026-10-21
started:
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.branch: T-5717
  workflow.model: opus
---

# Time, piece E (T-5717)

A piece of `T-5795`, planned in `records/plans/P-T-5795.md` (approved 21 Oct by the director): read its step for
piece E, its parallel-work checkpoint and its proof table; the contract is `design/contract.md` section 23
"Time". Evidence: `~/bonsai-checks/T-5795/inventory.md`, `~/bonsai-checks/T-5795/research.md`.
