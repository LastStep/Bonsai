---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5701
title: "5.7.1, the release build and its rehearsal: .goreleaser.yaml, the release-check job, the notes script, the dist checks"
status: done
lane:
done_when:
  - "GoReleaser pinned to an exact v2 version, its fetch checked (or go run by the module path); goreleaser check passes; .goreleaser.yaml as note 5.7.1, 2 (trimpath, installers at each archive's top, changelog and release disabled, base-v* ignored, no brews)"
  - ".github/scripts/release-notes.sh as note 5.7.1, 3, failing on a missing or empty section"
  - "internal/selfcheck/dist_test.go and markers.go as note 5.7.1, 4, skipped without BONSAI_CHECK_DIST; green on a local snapshot in WSL and natively on Windows, failing on each doctored copy of note 5.7.1, 7; a snapshot with scratch base-v9.9.9 tags passing"
  - "the release-check job in ci.yml (note 5.7.1, 5) green; install.sh refusing off Linux with its test (note 5.7.1, 6); 5.6.5's name test and 5.6.6's installer tests green"
  - "check 10; a green climb; CI green on the branch, release-check among it"
depends_on: [T-5792, T-5704]
blocked_by:
created: 2026-10-20
started: 2026-10-20
finished: 2026-10-20
labels:
  bonsai.ladder: []
  bonsai.allows: [".goreleaser.yaml", ".github/workflows/ci.yml", ".github/scripts/release-notes.sh"]
  bonsai.wants: [".goreleaser.yaml"]
  bonsai.branch: T-5701
---

# 5.7.1, the release build and its rehearsal (T-5701)

Plan: `design/plan-5.md`, step 5.7, note "5.7.1" and its row; the director approved the section on 10 Oct, which is
the word for the changes it names to `.goreleaser.yaml`. **Owns:** `.goreleaser.yaml`; `ci.yml`'s `release-check` job;
`.github/scripts/release-notes.sh` (new); `internal/selfcheck/dist_test.go` and `markers.go` (moved from
`cmd/bonsai/hook_test.go`, which then reads it); `install/install.sh` and its test.
