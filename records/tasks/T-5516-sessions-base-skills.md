---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5516
title: "5.5.5's fix round for base's skills (5.5.5c): the workspace skill's roles example in block form; the operating skill's person's steps held on a chat request"
status: running
lane:
done_when:
  - "packs/base/skills/workspace/SKILL.md's packs[].roles example written as indented key: value lines Bonsai's YAML reads, and every other example in base's skills read by Bonsai's own reader (a Go test that parses each yaml example, failing first on the flow mapping)"
  - "packs/base/skills/operating-bonsai/SKILL.md: a check loosened (a floor lowered, a rung taken out, a ratchet emptied) or a pack taken out or its version moved is never edited on a chat request: the agent hands the person the exact line or a Decide ask, as note 5.5.5's S4 and the section's \"Moving a pack's version\" say; the skill's test green"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; the orchestrator's read of the diff"
depends_on: [T-5505]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished:
labels:
  bonsai.branch: "t5516"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# 5.5.5's fix round for base's skills (5.5.5c)

Piece 5.5.5c: 5.5.5's real sessions (`T-5505`, findings 3 and 4) found base's workspace skill teaching a form Bonsai's
YAML cannot hold (an agent copying it locked itself out), and the operating skill letting a chat request stand for the
person's step. 5.5.2 and 5.5.3 own the files.

- **Owns:** `packs/base/skills/` and their tests in `packs/`; `internal/engine/skills.go` and its test (the workspace skill is generated from it; the orchestrator, 01:55).
- **Grants:** none: no protected path.
- **Proof:** the lines of `done_when`.
