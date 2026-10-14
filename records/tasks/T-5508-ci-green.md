---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5508
title: "CI green and 5.4's carry-overs: three flakes, the rungs' environment, the stop gate's installed path (5.5.8)"
status: todo
lane:
done_when:
  - "internal/ladder's TestLooserKindsRefused: every git call through testpack.Git passes -c maintenance.auto=false -c gc.auto=0; shown failing first in a temporary copy with git 2.47 or later (or the run report says why it could not be made to fail), then 20 runs green"
  - "internal/clean's TestEveryKind sets its own Budget; shown failing first with a 1 ns budget in a temporary copy, then 20 runs green natively on Windows"
  - "internal/rungjob's member (job_windows.go) skips a pid whose process is gone and still reports one it cannot open for another reason; a Go test where Windows lets the case be staged, else the run report says why it is read, not run"
  - "The ladder runner clears BONSAI_TASK and BONSAI_ROLE from every rung's environment: a fixture rung printing its environment holds neither while the caller's holds both"
  - "Every next step of the stop gate that names bonsai ladder --task names the installed bonsai its hook line names, read by its tests on both sides"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI's test and windows jobs green at the pushed commit; VG re-reads items 4 and 5"
depends_on: [T-5590]
blocked_by:
created: 2026-10-14
started:
finished:
labels:
  bonsai.branch: "t5508"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# CI green and 5.4's carry-overs: three flakes, the rungs' environment, the stop gate's installed path (5.5.8)

Piece 5.5.8 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece". Rohan, 14 Oct: leave `main`'s CI red and "fix both in 5.5's first piece"; the third flake on his "A. Fix it too".

- **Owns:** the fixtures' git settings in `internal/testpack/`; `internal/clean/`'s tests; `member` in `internal/rungjob/job_windows.go` and its test; the rungs' environment in `internal/ladder/`; the next steps' text in `internal/stopgate/`; their tests.
- **Grants:** none: no protected path among its files, so `bonsai.allows` is empty.
- **Proof:** the lines of `done_when`; the installed `bonsai`'s ladder green at the branch's HEAD; the landing rule; CI. VG re-reads the gate and runner lines.
