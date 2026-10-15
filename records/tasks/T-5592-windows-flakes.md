---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5592
title: "Two Windows CI flakes on e1437d0: an exited process read as a job member, and a reused pid in TestSweep"
status: done
lane:
done_when:
  - "internal/rungjob's member skips a pid whose process has exited (its exit code is no longer STILL_ACTIVE) as well as one OpenProcess can no longer open, and still reports a live member it cannot describe; a Windows Go test stages an exited process whose handle is still held and shows it skipped"
  - "internal/newtests' TestSweep holds an open handle to its finished process for the whole check, so Windows cannot give its pid to another process; shown failing first natively on Windows with the pid reused or the handle let go, where it can be staged, else the run report says why"
  - "Each test 20 runs green natively on Windows; check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI's windows job green at the pushed commit"
depends_on: [T-5508]
blocked_by:
created: 2026-10-15
started: 2026-10-15
finished: 2026-10-15
labels:
  bonsai.branch: "t5592"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# Two Windows CI flakes on e1437d0

Part 5.5's work outside its pieces: CI's `windows` job on `e1437d0` (5.5.0's landing, which touched neither package)
went red on two Windows process-timing tests. `internal/rungjob`'s `TestOutputAndExit` ("ended [{8632  }], want an
empty list"): 5.5.8's `member` skips a pid `OpenProcess` refuses with ERROR_INVALID_PARAMETER, but this one opened and
had exited, so it was reported with no name. `internal/newtests`' `TestSweep` ("the killed check's worktree is still
there"): its "dead" owner is a finished `git --version` whose pid Windows may give to another process before the sweep
reads it. Red CI is fixed forward (`CLAUDE.md`).

- **Owns:** `member` in `internal/rungjob/job_windows.go` and its test; `TestSweep` in `internal/newtests/fixtures_test.go`.
- **Grants:** none: no protected path, so `bonsai.allows` is empty.
- **Proof:** the lines of `done_when`; the orchestrator's read of the diff; CI.
