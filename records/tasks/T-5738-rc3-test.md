---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5738
title: "rc.3 tested by the orchestrator: the release read back, a fresh project from the README, each fix seen working"
status: running
lane:
done_when:
  - "the v1.0.0-rc.3 release read back on both sides as rc.2's was: release verify, verify-asset, attestation, checksums, TestDist, the installers to a scratch target, every program rebuilt byte-identical"
  - "the director's WSL install read back: the installed bonsai prints 1.0.0-rc.3 at the tag's commit, its sha256 the release's"
  - "the README's first project followed on the installed rc.3 by an agent in scratch, to its last step, every refusal and warning as the README says"
  - "each rc.3 fix seen working on the installed rc.3: config-uncommitted before the first commit and gone after; the reload line after a pack is added; the guard's words after a cd it cannot follow; the stop gate holding a named session on a run report check rejects (B3); the workflow pack's new words loaded"
  - "LinkSink read again, read-only; check 10's Windows half (vet and test, plain and tagged, natively on Windows) green at the tag"
depends_on: [T-5734]
blocked_by:
created: 2026-10-24
started: 2026-10-24
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# rc.3 tested by the orchestrator (T-5738)

The director, 24 Oct: "once rc3 lands and you have tested it, then we go to 1.0"; "okay go with rc.3, tell me when to
tag". The test replaces their own second trial (STATE's question, answered by these words). A failure here is a fix
and an rc.4, told to the director first.
