---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-20-T-5701
task: T-5701
role: orchestrator
model: opus
started: 2026-10-20 13:02
finished: 2026-10-20 13:57
outcome: merged
commits: ["0a9115c9", "f6468594", "4a0aad8b", "a55cf3ad", "93d65f0e", "67c44e3e", "137e419f"]
labels: {}
---

# Run: T-5701, 5.7.1, the release build and its rehearsal

- Who: the orchestrator, with the builder it briefs
- Why: `design/plan-5.md`, step 5.7, piece 5.7.1 (approved 10 Oct; the director's word for its `.goreleaser.yaml` changes)

> Opened before the first edit and appended to as the work goes.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Builder M-2026-10-20-93898dd8 (audit, build, WSL proof, two climbs) | opus | 2026-10-20 13:01 | 2026-10-20 13:35 | 34 |
| Orchestrator, the Windows dist run | opus | 2026-10-20 13:24 | 2026-10-20 13:27 | 3 |

## Log

- 13:02 T-5701 `running` (grants .goreleaser.yaml, ci.yml, the notes script; the one running task). Builder
  (opus: the release build) briefed, `M-2026-10-20-93898dd8`, audit first; base the batch's top 4ebee8cc; its
  Windows dist run to be asked as "Windows ready".
- 13:10 Audit `M-2026-10-20-c0f5e680` (`~/bonsai-checks/T-5701/audit.md`): GoReleaser v2.18.3 by go run (the action checks only a same-page checksums.txt); ignore_tags misses a base-v tag on HEAD silently, so GORELEASER_CURRENT_TAG in the job and the dist test wants a release-shaped version; Owns amended by the orchestrator: internal/selfcheck/doc.go, CONTRIBUTING.md's job list, cmd/bonsai/hook_test.go's markers. Tar owners carry the builder's user name locally: a later item. Go given.
- 13:23 Report `M-2026-10-20-ea619256` (`~/bonsai-checks/T-5701/windows-ready.md`), needs-input: branch at 137e419f, 7
  commits on 4ebee8cc, within the amended Owns; climb red at rung 0 only (.gitattributes from T-5704 below its base,
  main still dba4e7af), to climb again once main holds 4ebee8cc. WSL by hand: check 10, golangci-lint plain and tagged
  0 issues, actionlint, goreleaser check (v2.18.3), the job's steps in a fresh clone (TestDist PASS, notes 39 lines),
  five doctored copies each failing for its reason, base-v9.9.9 failing without the job's tag line and passing with
  it, the notes script's three failures, install.sh's new row failing first.
- 13:21 **The Windows dist run, by the orchestrator** (the one Windows run): the WSL snapshot and the bundle to
  %USERPROFILE%\bonsai-checks\T-5701, cloned with Windows git (137e419), Windows Go go1.27.2 with a narrowed PATH:
  `go test -count=1 -run TestDist -v ./internal/selfcheck/`: PASS (checksums, six archives, installer: install.ps1
  from the windows_amd64 zip to a scratch target), ok 4.346 s, exit 0.
- 13:35 Report `M-2026-10-20-b04787e1`, done: climb at 137e419f on main 4ebee8cc green, sha256
  953c6924f0cd03818480a840133caae4bba19ffd8ddd109ac007000bb271bd89 (rungs 0 3.7 s, 1 0.9 s, 2 0.4 s, 3 60.7 s 4778, 4
  57.9 s 2125, 5 0.04 s). The orchestrator read the diff: .goreleaser.yaml as note 5.7.1, 2 (no brews, no token,
  release and changelog disabled, trimpath, installers at the top, base-v* ignored); the release-check job (go run
  v2.18.3, the current tag set, notes, TestDist); paths within the amended Owns. Branch T-5701 pushed; CI run
  39390834758 on 137e419f.
- 13:57 CI run 39390834758 on 137e419f (branch T-5701): every job success, release-check among them (its
  first run on GitHub). `main` fast-forwarded 4ebee8cc..137e419f and pushed; the branch deleted on GitHub; the
  worktree removed by `git worktree remove`. Landing: git.sha 137e419f, sha256
  953c6924f0cd03818480a840133caae4bba19ffd8ddd109ac007000bb271bd89, green. Task `done`.

## Processes left running

None of this task's (the builder reported none; the CI watch exited; ps -eo pid,etime,cmd).

## What is left

A later item, not 5.7's: the tar archives record the building user's name as each file's owner (on GitHub, the
runner's); `archives.files[].info.owner/group`, a task after 1.0. The Windows scratch `%USERPROFILE%\bonsai-checks\T-5701`
can go at the part's cleanup.
