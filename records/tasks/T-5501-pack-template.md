---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5501
title: The pack template, its checks and Bonsai's CI job for packs (5.5.1)
status: todo
lane:
done_when:
  - "packs/template/ with one role declared with its budget, one skill, one documented template, pack.yaml, labels.yaml, README and .gitattributes; its checks in ci/check.sh; its pack.yml"
  - "check --pack and claude plugin validate --json on the template report only the missing version; the script green on both sides locally"
  - "The release subcommand run locally with made-up tag names: one unlike the version fails, the matching one passes; no tag, repository or release made"
  - "Its last commit (packs/template/.gitattributes, and the packs job and one step of the windows job in .github/workflows/ci.yml) made while T-5501 is the only task running, after T-5500 has landed; Bonsai's packs and windows jobs green on CI; validate's need of a login measured"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; V1 passes"
depends_on: [T-5509]
blocked_by:
created: 2026-10-14
started:
finished:
labels:
  bonsai.branch: "t5501"
  bonsai.ladder: []
  bonsai.allows: ["packs/template/.gitattributes", ".github/workflows/ci.yml"]
  bonsai.wants: []
---

# The pack template, its checks and Bonsai's CI job for packs (5.5.1)

Piece 5.5.1 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece".

- **Owns:** `packs/template/` (all but its `.gitattributes`); its last commit: `packs/template/.gitattributes`, and the `packs` job and one step of the `windows` job in `.github/workflows/ci.yml`.
- **Grants:** `packs/template/.gitattributes` (protected by `**/.gitattributes`) and `.github/workflows/ci.yml` (protected by `.github/**`, not person-only), for its last commit alone, made while it is the only task `running`.
- **Proof:** the lines of `done_when`; the installed `bonsai`'s ladder green at the branch's HEAD; the landing rule; CI. V1.
