---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-13-T-5491
task: T-5491
role: orchestrator
model: opus
started: 2026-10-13 13:36
finished: 2026-10-13 14:24
outcome: merged
commits: ['d7037dc']
labels: {}
---

# Run: T-5491, tests that read the caller's BONSAI_TASK set their own (D1)

- Who: the orchestrator; one Sonnet builder (a test-only, mechanical change)
- Branch: `t5491`, from `main` at `81cb8d7`, in the plain worktree `~/Servers/Bonsai-t5491`
- Why: the 5.4 end verifier's D1 (`R-2026-10-13-T-5490-verify-5.4-end.md`, 13:31); Rohan, 13 Oct, asked "fix it now,
  or note it as a known limit and close 5.4?": **"Fix now"**
- Proof: green tests and CI and the orchestrator's read of the diff (a test-only change); landed by the landing rule

> Opened before the first edit and appended to as the work goes.

## Log

- 13:36 Opened. `T-5490` to `done` (its verifier reported; its climb green at `c0f6778`, CI green); `T-5491` running,
  the only running task, granting nothing. Builder brief: `~/bonsai-checks/briefs/brief-T-5491.md`.
- 13:40 **CI on `81cb8d7` (records only) red on Linux:** `internal/ladder`'s `TestLooserKindsRefused` failed in its
  temp folder's cleanup, "TempDir RemoveAll cleanup: unlinkat .../project/.git/objects: directory not empty": a write
  into the test repository's objects still going on when the test ended (most likely a git process outliving the test's
  last climb or gate call, or git's own background upkeep). The same Go code was green on Linux at `c0f6778`,
  `6f25dd6`, `5a38f9a`, `c1bb21f` and before, in the verifier's WSL check 10 and in four climbs. **A flake, noted as a
  known limit beside the rungjob Windows flake** (Rohan, 13 Oct: issues found now are noted); not chased here (a small
  fix stays small); T-5491's push runs CI again. Fix after 1.0 or in 5.5: find which process writes there after the
  test (each fixture repository could also turn off git's background upkeep, `gc.auto 0`, `maintenance.auto false`).
- 13:44 **The builder reported** (Sonnet, 13:35-13:44): one commit, `d7037dc` "tests: two tests set their own
  BONSAI_TASK and BONSAI_ROLE", `cmd/bonsai/agent_test.go` +2 and `internal/status/ladder_test.go` +2, test files only.
  Both tests now `t.Setenv` the two variables to "" (the code reads "" as none named: `status.go` through
  `os.Getenv(workspace.TaskEnv)`, `internal/record`'s `fromEnv`, `guard.PersonGrant(..., os.Getenv)`); the sweep found
  no other test. Through `~/bonsai-checks/scripts/T-5491/sweep.sh`: with `BONSAI_TASK=T-5491 BONSAI_ROLE=builder` plain
  4188 pass, tagged 4194, 0 fail, 2 skips each; unset the same; `go vet` plain and tagged exit 0. Climb 1 (plain
  environment) green 6 of 6, result sha256 `a1a6c6f8...4f750f56`; climb 2 with `BONSAI_TASK=T-5491 BONSAI_ROLE=builder`
  in its own environment (through `climb2.sh`, the named-session case D1 broke) green 6 of 6, `go_tests 4188->4188`,
  `go_tests_tagged 4194->4194`, rungs 0.12/0.75/0.38/45.7/58.6/0.03 s. The guard refused its `cp` of the result file
  out of the main checkout (`.bonsai/local/` is Bonsai's); it did not work round it.
- 13:44 **The orchestrator's read and the landing.** The diff: four lines in two `_test.go` files, inside the task's
  Owns; nothing in `records/`, `design/`, `STATE.md`. Both tests run here with `BONSAI_TASK=T-5491 BONSAI_ROLE=builder`:
  ok (they failed so at `c0f6778`). **The landing rule:** `.bonsai/local/ladder/T-5491.json` in the main checkout reads
  `mode: local`, green, `git.sha` `d7037dc25c544da96575939aee7d17e04a315129` (the branch's HEAD), `dirty: false`,
  rungs 0-5 requested and green (the floor; `bonsai.ladder` empty), ratchets ok; **sha256
  `f28e1015f732044f6980674d1424379b71095503b3468dd75fb05fba3466c285`**; its `ladder` record names
  `/usr/local/bin/bonsai`, `bonsai_sha256` `356f787c...afeef644`, and that sha256. Check 10's Windows half: CI, by
  Rohan's word until 5.4 closes. `main` fast-forwarded to `d7037dc` and pushed; the worktree removed with `git worktree
  remove`, the branch deleted. `T-5491` to `verify`; `done` once CI is green on `d7037dc`.
- 13:51 **Rohan asked to run the two CI flakes in isolation**, then, part way, "okay just wait for current ci". Done
  before that: `internal/ladder`'s `TestLooserKindsRefused`, 30 runs in a row in a scratch clone at `d7037dc`, all pass
  (11 s). Its likely cause, read: CI's runner has git 2.55 (this machine's WSL git is 2.34); since git 2.47
  `maintenance.autoDetach` defaults to true, so the `git maintenance run --auto` that commands such as `git commit` start
  after writing (traced here: `run_command: git maintenance run --auto --quiet`) runs detached, and "each `git
  maintenance run` command takes a lock on the repository's object database" (git-scm.com, git-maintenance); the test
  commits about twenty times, and on a loaded runner the last detached run can still be writing in `.git/objects` as the
  temp folder is removed. Not reproducible on 2.34, which runs it in the foreground. Fix (test-only, not made): the
  fixtures' git calls set `maintenance.auto=false` (`testpack.Git`'s `-c` list, and wherever else tests run git). The
  Windows flake (`TestOutputAndExit`) was not run.
- 14:23 CI on `d7037dc` green on every job, Windows included (`81cb8d7`'s Linux red was the git flake above).
  `T-5491` to `done`. 5.4 closes: the end verifier's one must-fix is fixed and landed green; its other findings are
  noted (`STATE.md`, known limits). Next: the records commit, then one climb of `main` for the floors.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Builder | Sonnet | 13:35 | 13:44 | 9 |
| Helper: Claude Code's docs on stopping subagents (Rohan's question) | default (`claude-code-guide`) | 13:43 | 13:44 | 1 |
| Helper: git's docs on background maintenance | default (`web-fetch`) | 13:51 | 13:52 | 1 |
