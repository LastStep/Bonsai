---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5793
title: "Scratch folders, declared in bonsai.yaml: what one is, where it lives, which roles may use it and with what authority"
status: todo
lane: full
done_when:
  - "a plan (workflow:plan) the director approves: the bonsai.yaml field and its schema, the folder's place (the director: \"bonsai local maybe?\"), each role's authority in it, what the guard refuses, cleaning, and the formats and contract changes it brings"
  - "after approval: built by an Opus builder, checked by a fresh Opus verifier (guards), landed by the landing rule"
depends_on: []
blocked_by:
created: 2026-10-20
started:
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
---

# Scratch folders, declared (T-5793)

The director, 20 Oct, on the auditor's options (`M-2026-10-20-68e4b9a0`): "normally an agent shouldn't make changes
outside its repo at all. and now that we are giving these 'read only' sub agents a scratch folder (even that i'm
worried about to be honest), then that should be done with deliberation. we need a proper structure which decides what
a scratch folder is and what type of authority it can contain. it should be declared in bonsai.yaml. and also try to
find the ideal place where this should live, bonsai local maybe?" Supersedes the 19 Oct answer 3's "guard unchanged"
shape. BUG-013.

**20 Oct, later: after 1.0** (the director, asked whether now or after 1.0: "okay we will tackle this idea after 1.0").
Planning facts kept for then: the audits `M-2026-10-20-68e4b9a0` and `M-2026-10-20-d4f09499`, the research
`M-2026-10-20-6a934eca`, and the orchestrator's draft shape in `~/bonsai-checks/T-5793/notes.md`. Until then the
write-nothing roles keep today's rule (no file-tool writes; shell writes outside projects only; BUG-011's `~/` form).
