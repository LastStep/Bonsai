---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-15-T-5595
task: T-5595
role: orchestrator
model: opus
started: 2026-10-15 07:46
finished: 2026-10-15 13:01
outcome: merged
commits: ['ad85f24']
labels: {}
---

# Run: T-5595, Windows timing tests that fail under load

- Who: the orchestrator; one Sonnet builder (small and named exactly)
- Branch: `t5595`, from `main` at `80b5bf7`, in the plain worktree `~/Servers/Bonsai-t5595`
- Why: five red Windows check-10 runs on 15 Oct from wall-clock tests under load (T-5594's run1 and run3, T-5511's
  second run, T-5502's run5): each landing then needs a rerun and a judgement, and CI will meet them next
- Proof: the task's `done_when`; the landing rule; CI

> Opened before the first edit and appended to as the work goes.

## Log

- 07:46 `T-5595` written and to `running`; worktree from `80b5bf7`. The gate's own 4 s git limit stays a known limit
  (Rohan, 13 Oct); only the tests get their own limits. Brief `~/bonsai-checks/briefs/brief-T-5595.md`.
- 10:02 **Builder's report** (Sonnet, 07:46-10:02, no helpers): one commit, `01ae3c1` on `7615ec7`, nine files (tests,
  and `internal/stopgate/stopgate.go`'s `gitWait` made a variable with the same value and comment, "only tests change
  it"), inside Owns. `internal/clean`: a `patient(l)` helper (the cleaning with a one-hour budget) used by every
  `SessionEnd` test that left the 1 s budget to chance (eight calls, among them `TestDefaultsWhenLeftOut`,
  `TestAgentsNullAndSessionEnd`, `TestMessagesNullAndSessionEnd`, `TestStoppedAnywhereLeavesAWholeState`); the budget's
  own tests untouched. `internal/ladder`'s `TestLockOfADeadHolder`: the 5 s wall clock gone; exit 0 and an empty
  notice (the lock taken at once). `internal/stopgate`: `patientGit(t)` (one minute, restored by `t.Cleanup`) in the
  shared `gate()` helper, covering the timeout tests; the failing `TestAnAllowItCannotRecordIsBlocked` was the stop
  gate's, not the guard's. **Failing first**: a copy of `main` with the budget and `gitWait` at 1 ns fails
  `TestDefaultsWhenLeftOut`, the two `NullAndSessionEnd` tests and ten `TestStopGateFixtures` cases; with the fix, all
  pass at 1 ns; `TestLockOfADeadHolder` not shown failing (its bound was only a wall clock). 20 runs natively on
  Windows: `internal/clean`, `internal/ladder` and `internal/stopgate`'s tests all ok. Check 10 green on both sides,
  plain and tagged; lint clean. Climb green at `01ae3c1` (1.4/0.9/0.5/126/118/0.05 s; `go_tests` 4188->4479, tagged
  4194->4485), `sha256:b3f266f6...f2c4`.
- 10:03 **Held, not landed:** `base` (5.5.2) lands first (a plan piece 5.5.3 waits for), and landing this first would
  send `base` through another rebase and Windows run in the wrap-up Rohan asked for. `t5595` stays as it is, in its
  worktree; the next session rebases it onto `main` (tests only), re-runs check 10 and the climb, and lands it.
- 10:49 **The next session:** a Sonnet landing helper rebases `t5595` onto `main` at `2b9d09f`, re-runs check 10 on
  both sides and climbs (brief `~/bonsai-checks/briefs/brief-T-5595-land.md`). Landed alone, before `T-5598`'s
  Windows check, so the fix round's Windows halves meet these budgets, not the flakes.
- 11:37 **The helper's report** (Sonnet, 10:47-11:37): rebase clean, `ad85f24` on `2b9d09f`. Check 10: WSL plain 4485
  pass, 0 fail, 3 skips (Windows-only, each saying so), tagged 4491, 0 fail; vet clean; lint 0 issues. Windows
  natively from a CRLF checkout at `ad85f24`: plain and tagged 29 packages ok, no FAIL; vet clean; Defender nothing.
  Climb green (2.4/1.1/0.5/72.1/92.8/0.06 s; `go_tests` 4188->4485, tagged 4194->4491; `new_tests` 0, base `2b9d09f`).
- 11:40 **Landed by the landing rule** (a batch of one: `T-5598` not ready): `.bonsai/local/ladder/T-5595.json` mode
  local, green, `git.sha` `ad85f248ff754e9782a2da4c4d17d3fc919dae93` (the branch's HEAD), dirty false, rungs 0-5
  green; its `ladder` record (05:30:26Z) names `/usr/local/bin/bonsai` and
  `sha256:c35c822f2fe84faa50aca70970c29fcbc64bfcdb394d1af0065897d8484ed271`. The diff: nine files, inside Owns.
  `main` fast-forwarded `2b9d09f..ad85f24` and pushed; the worktree removed, the branch deleted. `T-5595` to `verify`
  until CI is green at `ad85f24`.
- 13:01 **CI green at `ad85f24`** (run 38692545835: `test`, `lint`, `packs`, `govulncheck`, `windows` 26 min, CodeQL).
  **`T-5595` to `done`.**

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Builder | Sonnet | 07:46 | 10:02 | 136 |
| Landing helper: rebase, check 10, climb | Sonnet | 10:47 | 11:37 | 50 |
