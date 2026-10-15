---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-15-T-5595
task: T-5595
role: orchestrator
model: opus
started: 2026-10-15 07:46
finished:
outcome:
commits: []
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

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Builder | Sonnet | 07:46 | 10:02 | 136 |
