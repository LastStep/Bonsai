---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5705
title: "5.7.5, the releases checked: batch 1 read back, each pre-release's and 1.0's files checked on both sides, V2, the director's installs read back"
status: verify
lane:
done_when:
  - "batch 1's settings read back by an agent, changing nothing, each as the plan's batch 1 says"
  - "each pre-release (rc.1 and each fix round's) and 1.0: every published file checked on both sides (checksums, the signed record, the archives' contents, the installers), V2 on rc.1, the director's installs read back, check in Bonsai's repo on each"
depends_on: [T-5702]
blocked_by:
created: 2026-10-20
started: 2026-10-20
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
---

# 5.7.5, the releases checked (T-5705)

Plan: `design/plan-5.md`, step 5.7, note "5.7.5" and its row. Owns: `records/` only.
