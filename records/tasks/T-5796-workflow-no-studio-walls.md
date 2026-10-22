---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5796
title: "The workflow pack names no studio's files: its twelve Trinetra deny rules taken out"
status: done
lane: light
done_when:
  - "LastStep/bonsai-workflow: the twelve Read deny rules on ~/.trinetra/ and their README section gone, every other mention checked; the pack's own checks (ci/check.sh) and CI green; main pushed by the orchestrator"
  - "Bonsai's bonsai.yaml pins workflow at that commit; bonsai update in the main checkout drops the twelve rules from .claude/settings.json and the lock; bonsai check clean; the records commit lands by CI"
depends_on: []
blocked_by:
created: 2026-10-22
started: 2026-10-22
finished: 2026-10-22
labels:
  bonsai.ladder: []
  bonsai.allows: ["bonsai.yaml", ".bonsai/lock.json"]
  bonsai.wants: ["bonsai.yaml", ".bonsai/lock.json"]
  workflow.model: sonnet
---

# The workflow pack names no studio's files (T-5796)

The director, 22 Oct: "remove the trinetra paths from the workflow pack". The feature audit for T-5711 found them:
`bonsai/pack.yaml:29-58` (twelve `Read(...)` rules on `~/.trinetra/token`, `deploy_ed25519` and `salt`, in four home
forms) and `README.md`, "The walls"; Bonsai's link carries them into `.bonsai/lock.json` and `.claude/settings.json`.
A pack for everyone names no one studio's files; the studio keeps its own walls (its own pack or settings).

Two halves. The pack's change in `~/bonsai-checks/bonsai-workflow` (a builder, its worktree
`~/bonsai-checks/bonsai-workflow-T-5796`, branch T-5796; the orchestrator pushes). Then Bonsai's link: the ref in
`bonsai.yaml` and `bonsai update` in the main checkout, by the orchestrator, once no other task reads `running`
(this task's `bonsai.allows` then gains `bonsai.yaml` and `.bonsai/lock.json`, on the director's word above). Taking
deny rules out loosens the walls: the director asked for it.
