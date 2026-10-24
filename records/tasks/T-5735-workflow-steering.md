---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5735
title: "The workflow pack steers planning and decision work: run report, STATE, memory, running, one task at a time"
status: done
lane:
done_when:
  - "LastStep/bonsai-workflow: session-start, orchestrate and lanes say a run report opens when any work on a task starts (planning, research, decisions included), that the orchestrator keeps STATE true as it goes and its notes in the project's memory (base:memory), that a task reads running when any agent starts on it, and what two running tasks cost; ci/check.sh check and validate clean; CI green; main pushed by the orchestrator"
  - "Bonsai's bonsai.yaml pins workflow at that commit; bonsai update in the main checkout; bonsai check clean; the records commit lands by CI"
depends_on: []
blocked_by:
created: 2026-10-24
started: 2026-10-24
finished: 2026-10-24
labels:
  bonsai.ladder: []
  bonsai.allows: [".bonsai/lock.json", "bonsai.yaml"]
  bonsai.wants: ["bonsai.yaml", ".bonsai/lock.json"]
  workflow.model: sonnet
---

# The workflow pack steers planning and decision work (T-5735)

T-5734 item 1: BUG-103, BUG-104, BUG-105 and the build stage's two tasks `running` at once (`records/trial-rc2.md`).
The pack's change in `~/bonsai-checks/bonsai-workflow` (a builder, worktree `~/bonsai-checks/bonsai-workflow-T-5735`,
branch T-5735; the orchestrator pushes); then Bonsai's pin and `bonsai update` in the main checkout, as `T-5732`
(`bonsai.yaml` and the lock granted then, on the director's 24 Oct word for rc.3).
