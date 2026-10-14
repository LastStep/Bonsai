---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5512
title: bonsai agents, the studio's view of every agent, and the usage records (5.5.12)
status: todo
lane:
done_when:
  - "bonsai agents [--json] (bonsai.agents/1) lists every agent seen in the project, the person's sessions and every subagent, with recorded facts only and its usage token counts; it reads only and never acts, derives no state and reads no setting"
  - "Go tests on a fixture project: a lead and two helpers, one recorded spent and one with no recent call, beside the person's session; an ended agent; a resumed agent with two ends; a brief's limit below the role's default shown as given; a report from another address not taken; a malformed state file skipped and named; each entry equal to what was recorded"
  - "A transcript in the measured layout summed exactly, one in another layout recorded unknown with its reason, the agent's end never blocked"
  - "--json against its schema, byte-stable; the tree drawn from parent alone; scripted runs on both sides"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; VG passes"
depends_on: [T-5511]
blocked_by:
created: 2026-10-14
started:
finished:
labels:
  bonsai.branch: "t5512"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# bonsai agents, the studio's view of every agent, and the usage records (5.5.12)

Piece 5.5.12 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece". Rohan, 14 Oct: "Add it to 5.5"; it shows raw recorded facts only, the studio builds the interface.

- **Owns:** `cmd/bonsai/agents.go` (new), its word in the registry and the flag table; the view in `internal/agents/`; the `usage` record in `internal/recorder/` (at SubagentStop and SessionEnd); their tests and test data.
- **Grants:** none: no protected path among its files, so `bonsai.allows` is empty.
- **Proof:** the lines of `done_when`; the installed `bonsai`'s ladder green at the branch's HEAD; the landing rule; CI. VG.
