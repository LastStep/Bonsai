---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5739
title: "Bonsai 1.0: the notes' date, the v1.0.0 tag, 1.0's files checked, the director's installs and memory line"
status: running
lane:
done_when:
  - "CHANGELOG.md's heading reads ## [1.0.0] - <the date>, in a commit of its own; git diff --stat from 474ae7f7 holds only that line, records/ and .bonsai/STATE.md; every CI check run on that commit complete and green"
  - "the director's v1.0.0 tag on that commit and their approval; the release published by release.yml, a full release (not a pre-release), six archives and checksums.txt"
  - "1.0's files read back on both sides as rc.3's were (note 5.7.5, 2): release verify, verify-asset, attestation, checksums, TestDist, the installers to a scratch target, every program rebuilt byte-identical"
  - "the director's WSL and Windows installs read back (notes 5.7.5, 6 and 8); their memory line's tail read and the memory check run (note 5.7.5, 7)"
depends_on: [T-5738]
blocked_by:
created: 2026-10-25
started: 2026-10-25
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# Bonsai 1.0 (T-5739)

The director, 25 Oct, after rc.3's test (`T-5738`) came back green: "lets go", their word for 1.0 (plan-5.md, batch
5). Note 5.7.5, 4's date commit, then the tag, 1.0's files checked, the installs and the memory line. 5.7.6 (the packs
at 1.0) and the pack tags follow under their own task; then the floors and the 5.7 end verifier.
