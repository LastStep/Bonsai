---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6011
title: "T-6009's stage 2, the director's eight picks: what changed, the session nudge, set --all, the downgrade guard, the compatibility check, the old registration's hint, commit lines, pack show"
status: todo
lane: null
done_when:
  - "Steps 9-17 of records/plans/P-T-6009.md (approved by the director 29 Oct) built on main with stage 1 (T-6010): PackChanges in set's preview and outdated --changes, the start hook's packs-behind line, pack set --all, pack-downgrade and --downgrade, pack-needs before any write and spec section 7's line rewritten, the old marketplace's hint, the commit lines in text and commit_lines, pack show and bonsai.packs/1 filled, the docs"
  - "Every stage-2 row of the plan's proof table holds, each by its named test; no format change (stage 1 defined every field); the floor's rungs 0-5 green at the branch HEAD by the installed bonsai (bonsai ladder --task T-6011)"
  - "A fresh Opus verifier passes the branch's diff (tests and correctness, then security on the gates); landed by the landing rule, CI green on the branch, both windows jobs among them"
depends_on: [T-6010]
blocked_by: null
created: 2026-10-29
started: null
finished: null
labels:
  bonsai.ladder: []
  bonsai.allows: ["design/bonsai-spec.md"]
  bonsai.branch: T-6011-pack-command-picks
  workflow.model: opus
---

# The pack command's eight picks (T-6009's stage 2)

The director, 29 Oct: "approved. now build it end to end in this session". Steps 9-17 of `records/plans/P-T-6009.md`
and the stage-2 rows of its proof table, from `main` once T-6010 has landed. The plan's approval covers spec section
7's rewrite (pick 5, "Refuse, change the rule").

## Out of scope

- Any format change: stage 1 (T-6010) defined every field stage 2 fills.
- BUG-141 and BUG-142 (filed 29 Oct for the director).
