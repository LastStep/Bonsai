---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5517
title: "5.5.5's fix round for workflow (5.5.5d): the orchestrate skill's line on who starts a role"
status: done
lane:
done_when:
  - "workflow's skills/orchestrate/SKILL.md line 38 (\"You start every role; never ask another agent to start one for you\") squared with a lead calling a helper, as the pack's role rules allow"
  - "check --pack and validate green; the commit grepped by Haiku and read by the orchestrator before its push; workflow's CI green on both sides"
depends_on: [T-5505]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished: 2026-10-16
labels:
  bonsai.branch: null
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# 5.5.5's fix round for workflow (5.5.5d)

Piece 5.5.5d: 5.5.5's R7 (`T-5505`, finding 6): the orchestrator read the orchestrate skill's line as forbidding what
the role rules allow. 5.5.4 owns the file.

- **Owns:** `skills/orchestrate/SKILL.md` in the `workflow` repository; nothing in Bonsai's repo.
- **Grants:** none.
- **Proof:** the lines of `done_when`.
