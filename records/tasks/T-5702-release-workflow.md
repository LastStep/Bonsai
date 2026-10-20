---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5702
title: "5.7.2, the release workflow: release.yml in two jobs, the tag check, the templates, CLAUDE.md's release lines"
status: done
lane:
done_when:
  - "release.yml as note 5.7.2, 1-2: started only by the two tag patterns; build (read-only, no environment) and publish (environment release, runs nothing from the build: attestation, draft, draft checked, publish, published release checked, its own draft deleted on failure); every action pinned"
  - ".github/scripts/release-tag-check.sh as note 5.7.2, 2.2, run locally against Bonsai's history as note 5.7.2, 7 says (main passes, a scratch branch fails, v1.0 and v1.0.0-beta fail; a running named job fails; its own suite left out)"
  - "the issue and pull request templates (note 5.7.2, 5); CLAUDE.md's GitHub lines (note 5.7.2, 6)"
  - "actionlint and the pin test green; a green climb; CI green on the branch"
depends_on: [T-5701]
blocked_by:
created: 2026-10-20
started: 2026-10-20
finished: 2026-10-20
labels:
  bonsai.ladder: []
  bonsai.allows: [".github/workflows/release.yml", ".github/scripts/release-tag-check.sh", ".github/ISSUE_TEMPLATE/**", ".github/pull_request_template.md", "CLAUDE.md"]
  bonsai.wants: [".github/workflows/release.yml"]
  bonsai.branch: T-5702
---

# 5.7.2, the release workflow (T-5702)

Plan: `design/plan-5.md`, step 5.7, note "5.7.2" and its row; the director approved the section on 10 Oct, the word
for the changes it names to `release.yml`. **Owns:** `.github/workflows/release.yml`;
`.github/scripts/release-tag-check.sh` (new); `.github/ISSUE_TEMPLATE/` and `.github/pull_request_template.md`;
`CLAUDE.md` (its GitHub lines).
