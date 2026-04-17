# Plan 13 — ActionUnchanged Follow-ups from Plan 12

**Tier:** 1 (Bug fix + test coverage)
**Status:** Ready
**Source:** Plan 12 (PR #20) review follow-ups — Backlog Group B

---

## Goal

Close two small gaps left behind after Plan 12 introduced `ActionUnchanged`:

1. Two test coverage gaps — `TestWriteResultSummary` swallows the new `unchanged` return value; no dedicated test covers the `WorkspaceClaudeMD` short-circuit at `internal/generate/generate.go:829-833`.
2. `writeFileChmod` skips chmod when the inner `writeFile` returns `ActionUnchanged`. If a sensor script's exec bit was stripped externally (e.g. `chmod -x`) but its content is untouched, `bonsai update` reports "Up to date" and leaves the file non-executable.

### Success Criteria

- `TestWriteResultSummary` asserts the new `unchanged` count.
- A new test covers the `WorkspaceClaudeMD` short-circuit path — verifies both `ActionUnchanged` and that no write occurred.
- `writeFileChmod` re-applies perms when the inner result is `ActionUnchanged`.
- A regression test proves `writeFileChmod` restores the exec bit on an identical-content run where the file was externally `chmod -x`'d.
- `go build ./...`, `go vet ./...`, `go test ./...`, `gofmt -s -l .` all clean.

---

## Steps

### Step 1 — Expand `TestWriteResultSummary` (`internal/generate/generate_test.go:352`)

Add an `{Action: ActionUnchanged}` fixture to the `Files` slice. Change the destructure at line 363 from `created, updated, _, skipped, conflicts := wr.Summary()` to name the unchanged value, and add an `if unchanged != 1 { ... }` assertion. Keep existing assertions unchanged.

### Step 2 — Add `TestWorkspaceClaudeMDUnchangedShortCircuit` (`internal/generate/generate_test.go`)

Cover `internal/generate/generate.go:829-833`:

1. Build a test catalog + config, call `WorkspaceClaudeMD()` once — creates the file with markers.
2. Record the file's `ModTime()`.
3. Sleep a few ms (or use a monotonic-safe comparison) to ensure mtime would differ on rewrite.
4. Call `WorkspaceClaudeMD()` again with identical inputs.
5. Assert the most recent `FileResult` in the `WriteResult` for the workspace CLAUDE.md path has `Action == ActionUnchanged`.
6. Assert the file's `ModTime()` is unchanged from step 2 (no rewrite occurred).

Match the style of existing workspace-CLAUDE.md tests (see `TestClaudeMDHasMarkers` around `generate_test.go:380`).

### Step 3 — Fix `writeFileChmod` gate (`internal/generate/generate.go:301-310`)

Add `ActionUnchanged` to the gate so chmod runs on identical-content runs:

```go
if result.Action == ActionCreated || result.Action == ActionUpdated || result.Action == ActionForced || result.Action == ActionUnchanged {
    absPath := filepath.Join(projectRoot, relPath)
    _ = os.Chmod(absPath, perm)
}
```

Rationale: re-applying the generator-declared perm on `ActionUnchanged` is a no-op when the file was never tampered with, and a correct recovery when it was. No scope widening to `ActionConflict`/`ActionSkipped` — those cases intentionally skip writes.

### Step 4 — Add `TestWriteFileChmodRestoresPermOnUnchanged` (`internal/generate/generate_test.go`)

Regression test for Step 3:

1. Create a file via `writeFileChmod(..., perm=0755)` in a temp dir.
2. Externally strip the exec bit: `os.Chmod(absPath, 0644)`.
3. Call `writeFileChmod(...)` again with the **same** content + perm `0755`.
4. Assert returned `FileResult.Action == ActionUnchanged`.
5. Stat the file and assert `mode & 0777 == 0755`.

---

## Dependencies

- No new Go modules.
- All changes within `internal/generate/` (generate.go + generate_test.go).

---

## Security

> [!warning]
> Refer to SecurityStandards.md for all security requirements.

- No secrets or credentials handled.
- The chmod widening only includes `ActionUnchanged`, which still represents generator-managed files. Perm value comes from the generator's own declaration (e.g. `0755` for sensor scripts) — not user input. No privilege escalation path.
- No new file-system surface; no new callers.

---

## Verification

### Build & Test

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `gofmt -s -l .` (expect empty output)

### Manual

- [ ] In a dogfooded workspace: `chmod -x station/agent/Sensors/status-bar.sh`, then run `bonsai update`. Expect:
  - Panel still shows "Up to date" (content unchanged).
  - `ls -l station/agent/Sensors/status-bar.sh` shows exec bit restored.
- [ ] Run `bonsai update` twice on a clean workspace: second run shows "Up to date"; workspace CLAUDE.md mtime did not change between runs.

---

## Dispatch

| Phase | Agent | Isolation | Notes |
|-------|-------|-----------|-------|
| All | general-purpose | worktree | ~40 lines total: 1 production-code change, 2 new tests, 1 expanded test. Single commit. |
