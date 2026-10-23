---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5798
title: "Bonsai under pi: the contract plan (the pi side in pi-harness, the envelope bonsai.hook/1)"
status: done
lane:
done_when:
  - "design/pi-contract.md written: everything the pi side (one plain-JavaScript extension in ~/Servers/pi-harness) must do, its stopgap mode against today's bonsai, its envelope mode against bonsai.hook/1, and how it is proved, precise enough that an agent with no other context builds it"
  - "a draft contract section for bonsai.hook/1 (request and answer for guard, stop, start and record) in the plan, for the director to approve"
  - "a fresh Opus agent reviewed the plan; the director approved it"
depends_on: [T-5797]
blocked_by:
created: 2026-10-23
started: 2026-10-23
finished: 2026-10-23
labels:
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# Bonsai under pi: the contract plan (T-5798)

From the walkthrough of `design/agent-neutral.md` (23 Oct, the director's answers in
`records/runs/R-2026-10-22-T-5797-any-agent-research.md`): 1.0 first (C1); translate (C2); pi builds the pi side in
the director's pi-harness repo now, in parallel, with a stopgap mode against today's binary, while the Go side is
built here after 1.0 (C3 reshaped); the wider tripwire with the markers in rc.2 (C4, C8; `T-5710` item 13); plain
JavaScript, its own package last in pi's `packages` (C5, C6); the envelope only for now (C7); the after-1.0 proof on
the director's pi login (C9).

No code in this repo. The plan is a document the director hands to pi; nothing here writes into pi-harness.
