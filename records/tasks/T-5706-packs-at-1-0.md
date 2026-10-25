---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5706
title: "5.7.6, the packs at 1.0: the template's and the workflow repo's CI take bonsai from the 1.0 archive by its SHA-256; base and workflow at version 1.0.0"
status: done
lane:
done_when:
  - "packs/template's pack.yml takes bonsai from the 1.0 release archive for the runner's system (bonsai_1.0.0_linux_amd64.tar.gz; bonsai_1.0.0_windows_amd64.zip), from the fixed release URL, checked against the SHA-256 written beside it (from checksums.txt), each with a comment saying how to move it; ci/check.sh takes the bonsai it is given; base's ci skill the same for Linux"
  - "packs/base/bonsai/pack.yaml at version: \"1.0.0\"; check --pack on base green"
  - "the template's checks green with the archive on both sides, and a temporary copy with a wrong SHA-256 fails before running anything"
  - "Bonsai's packs job (still building bonsai from the commit it checks) and every CI check green on the branch; landed by the landing rule"
  - "in the workflow repo: its pack.yml as the template's new one and bonsai/pack.yaml at version: \"1.0.0\", one commit, pushed by the orchestrator, its CI green on both sides"
  - "the two tag lines sent to the director naming those two commits; after their tags, the workflow repo's release job green and its release made, and no run started by base-v1.0.0 in Bonsai's repo"
depends_on: [T-5739]
blocked_by:
created: 2026-10-25
started: 2026-10-25
finished: 2026-10-25
labels:
  bonsai.ladder: []
  bonsai.allows: ["packs/base/bonsai/pack.yaml"]
  bonsai.branch: T-5706-packs-1-0
  workflow.model: sonnet
---

# 5.7.6, the packs at 1.0 (T-5706)

plan-5.md's note 5.7.6 (spec §5, the pack's CI, step 2: "The pinned `bonsai` comes from a Bonsai release archive
checked by its SHA-256"). Owns: `packs/template/.github/workflows/pack.yml`, `packs/template/ci/check.sh`,
`packs/base/bonsai/pack.yaml` (protected, granted here), `packs/base/skills/ci/SKILL.md`; in the workflow repo,
`.github/workflows/pack.yml` and `bonsai/pack.yaml`. Bonsai's own `packs` job still builds `bonsai` from the commit it
checks. Bonsai's repo keeps naming `base` by its commit (moving it to the tag waits for the director's word).
