# Session Log — Plan 13 ActionUnchanged Follow-ups

**Date:** 2026-04-17
**Plan:** 13
**PR:** #22 (squashed into `86e7adf`)
**Status:** Shipped

---

## Scope

Closed two small gaps left by Plan 12 (PR #20) once `ActionUnchanged` landed:

1. **Test coverage** — `TestWriteResultSummary` swallowed the new `unchanged` return via `_`; no dedicated test covered the `WorkspaceClaudeMD` short-circuit at `internal/generate/generate.go:829-833`.
2. **Bug** — `writeFileChmod` only chmod'd on `{Created, Updated, Forced}`. Plan 12's short-circuit made `ActionUnchanged` the new normal, so a sensor script whose exec bit was stripped externally would stay non-executable across `bonsai update` runs.

## Changes

- `internal/generate/generate.go` (+1/-1): added `ActionUnchanged` to the chmod gate in `writeFileChmod`.
- `internal/generate/generate_test.go` (+124/-1):
  - Expanded `TestWriteResultSummary` with an `ActionUnchanged` fixture + `unchanged == 1` assertion.
  - Added `TestWorkspaceClaudeMDUnchangedShortCircuit` — back-dates the file's mtime via `os.Chtimes` before the second call so rewrites are detectable on coarse-mtime filesystems.
  - Added `TestWriteFileChmodRestoresPermOnUnchanged` — regression test: create at 0755 → strip to 0644 externally → re-run with identical content → assert `ActionUnchanged` + mode restored to 0755.

## Flow

- Planning → commit → dispatch general-purpose agent in a worktree → review diff → push → PR → CI pass → squash-merge.
- Agent stayed exactly in scope. Noted (but correctly did not fix) that `gofmt -s -l .` from the repo root walks into `.claude/worktrees/agent-*/` and flags unrelated files — potential `.gitignore`-adjacent improvement, out of scope.

## Backlog Impact

Removed:
- `[debt] Plan 12 follow-up — ActionUnchanged test coverage gaps` (P2 Group B)
- `[bug] writeFileChmod skips chmod on ActionUnchanged` (P2 Group B)

No new items discovered this session.

## Next

Plan 08 Phase C (new sensors: `compact-recovery` sensor + `context-guard` verification/planning triggers). Remains the top carry-forward task — unchanged by this session.
