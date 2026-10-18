---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5696
title: "CI's windows job red since 43e8b5fb: RealPath and the guard on 8.3 short names"
status: done
lane:
done_when:
  - "CI's windows job green at the fix's commit: TestFolderHoldsPastTheOldLimit, TestARelativeLinkTargetIsReadFromTheLinksRealFolder (internal/engine) and TestAShellWriteTheGuardCannotPlaceIsRefused (internal/guard) pass where the temp folder is an 8.3 short name (the CI runner's user folder)"
  - "the cause named: a test expecting the path as spelled, or the code reading a short-name path as leading nowhere; a code fault fixed in the code, never by loosening the test; a short-name case held by a test that runs on any Windows (a short name made or read for the temp folder), shown failing first"
  - "joined from the audit M-2026-10-18-39aecfea: may.go's bad-path refusal text names the walk's real cause on every system (not only Windows); TestEveryRule holds bad-path on Linux by a link loop as well as /proc; WalkLinks reads a drive-relative link target (C:foo) as relative to that drive's current folder or refuses it, never as from the drive's root"
  - "WSL check 10, lint, a green climb"
depends_on: [T-5694]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5696
---

# CI's windows job on 8.3 short names (T-5696)

CI's `windows` job has failed on `43e8b5fb`, `192a29dc` and `91d8a972` (run 39124123297): the runner's temp folder is
an 8.3 short name; `engine.RealPath` returns the long form where the tests want the short one, and
the guard refuses a deep shell write with `bad-path` ("leads nowhere that exists") where `bonsai-files` is wanted.

- **Owns:** `internal/guard/may.go`, `rules.go` and `rules_*_test.go` (the audit's items); `internal/engine/hookline.go`, `internal/engine/realpath*.go` and the engine's walk files T-5694 made, and
  their tests; `internal/guard/paths.go`, `internal/guard/deeppath_test.go` and the guard's files T-5694 touched, and
  their tests.
- **Grants:** none.
