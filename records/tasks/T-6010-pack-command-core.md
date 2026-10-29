---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6010
title: "T-6009's stage 1, the pack command's core: bonsai pack outdated, set, add and remove, the status cache, BUG-139, the preview's --json and its fingerprint"
status: done
lane: null
done_when:
  - "Steps 1-8 of records/plans/P-T-6009.md (approved by the director 29 Oct) built: the byte-preserving bonsai.yaml editor, Request.Config, the loosening gate for pack and update, the pack cache and status's ref and newer, BUG-139, the pack word with outdated, set, add and remove and its guard rules, formats set 13 (bonsai.changes/1's new fields, bonsai.packs/1, status), the fingerprint and --expect, and the docs, spec section 4, contract section 12 and base's operating-setup skill"
  - "Every row of the plan's proof table for stage 1 holds, each by its named test; the floor's rungs 0-5 green at the branch HEAD by the installed bonsai (bonsai ladder --task T-6010)"
  - "A fresh Opus verifier passes the branch's diff (tests and correctness, then security on the guard words); landed by the landing rule, CI green on the branch, both windows jobs among them"
depends_on: [T-6009]
blocked_by: null
created: 2026-10-29
started: 2026-10-29
finished: 2026-10-29
labels:
  bonsai.ladder: []
  bonsai.allows: ["design/bonsai-spec.md", "design/contract.md", ".github/workflows/ci.yml"]
  bonsai.branch: T-6010-pack-command-core
  workflow.model: opus
---

# The pack command's core (T-6009's stage 1)

The director, 29 Oct: "approved. now build it end to end in this session". The plan is `records/plans/P-T-6009.md`,
steps 1-8 and the stage 1 rows of its proof table; stage 2 (steps 9-17) is its own task after this one lands.

## Scope

Steps 1-8 of the plan, as written there. Its approval covers the edits of `design/bonsai-spec.md` section 4,
`design/contract.md` section 12 and the changes format's words.

## Out of scope

- Stage 2, the director's eight picks (steps 9-17).
- `unlink`'s grant rule (the plan's "Out of this task").
