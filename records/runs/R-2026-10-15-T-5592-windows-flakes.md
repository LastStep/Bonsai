---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-15-T-5592
task: T-5592
role: orchestrator
model: opus
started: 2026-10-15 03:21
finished:
outcome:
commits: []
labels: {}
---

# Run: T-5592, two Windows CI flakes on e1437d0

- Who: the orchestrator; one Sonnet builder (small and named exactly, as 5.5.8's), its Windows runs its own or a Sonnet
  helper's
- Branch: `t5592`, from `main` at `c641c15`, in the plain worktree `~/Servers/Bonsai-t5592`
- Why: CI's `windows` job red on `e1437d0` (CI run 38640084667); red CI is fixed forward (`CLAUDE.md`)
- Proof: the task's `done_when`; the landing rule; CI's `windows` job

> Opened before the first edit and appended to as the work goes.

## Log

- 03:21 **CI on `e1437d0`** (5.5.0's landing): `test`, `lint`, `govulncheck`, CodeQL green; **`windows` red** on two
  tests in packages 5.5.0 did not touch. `internal/rungjob`'s `TestOutputAndExit` ("ended [{8632  }], want an empty
  list: nothing was left"): the flake 5.5.8's item 3 met, not wholly closed: `member` skips a pid `OpenProcess`
  refuses with ERROR_INVALID_PARAMETER (87), but a process that has exited and still opens is described with no name.
  `internal/newtests`' `TestSweep` ("the killed check's worktree is still there: <nil>"; "no note of the sweep"):
  its "dead" owner is a finished `git --version`, whose pid a busy runner can give to a new process before the sweep
  asks `alive` (which already reads the exit code). The orchestrator's read: one cause, a Windows pid's life; fixes:
  `member` reads the exit code after a successful open (as `newtests`' `alive` does); `TestSweep` holds a handle to
  its finished process so its pid cannot be reused. `T-5592` written and to `running`; worktree from `c641c15`.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
