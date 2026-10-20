---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5703
title: "5.7.3, the stronger check in the install lines: gh attestation verify before unpacking, and docs/install.md"
status: running
lane:
done_when:
  - "5.6.5's lines table gains, on each side after the fingerprint line and before unpacking, gh attestation verify <archive> -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v<version> --deny-self-hosted-runners; its exact-text test built only from the checked version (plan-5.md note 5.7.3, 1-2)"
  - "docs/install.md (new) as note 5.7.3, 3 says, its test holding its two blocks to the table for 1.0.0 byte for byte and failing on a changed line"
  - "the operating skill's install line names docs/install.md and bonsai_release.lines (note 5.7.3, 4); its test green"
  - "the line's form run on both sides against gh's own signed release file, the wrong repository and --deny-self-hosted-runners, words and exit codes recorded (note 5.7.3, 5); WSL check 10; a green climb; CI green on the branch"
depends_on: [T-5792]
blocked_by:
created: 2026-10-20
started: 2026-10-20
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.branch: T-5703
---

# 5.7.3, the stronger check in the install lines (T-5703)

Plan: `design/plan-5.md`, step 5.7, note "5.7.3" and its row. **Owns:** `internal/engine/release.go`'s lines table and
its test; `docs/install.md` (new) and its test; `packs/base/skills/operating-bonsai/SKILL.md`'s install line.
