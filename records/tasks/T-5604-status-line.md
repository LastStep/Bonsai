---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5604
title: "`status --line` (5.6.4)"
status: done
lane:
done_when:
  - "status --line: one ASCII line of Bonsai's part (the active task, its proof, verify and asks counts), --json as bonsai.line/1, never failing, no git, Claude Code or network"
  - "Go tests on fixture projects (each part; two tasks running; not linked; a broken bonsai.yaml; a title with escape codes; its input held open); its p50 and p95 within budget on WSL (Windows timing to the next session's check)"
  - "check 10 both sides, the ladder, CI"
depends_on: [T-5600]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished: 2026-10-18
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5604
---

# `status --line` (5.6.4)

Piece 5.6.4 of step 5.6: its row in `design/plan-5.md`'s "#### The pieces and their order" and its note in "#### Notes
per piece" (lines 6608-6643). One of the four run side by side after 5.6.0 ("The order", step 2).

- **Owns:** `cmd/bonsai/status.go` (`--line` built; the help's "step 5.6" wording) and `cmd/bonsai/statusflags_test.go`; a new `internal/status/line.go`, calling the package's functions and changing no line of `status.go`; the not-built cases in `cmd/bonsai/main_test.go`, `outputs_test.go` and `operating_test.go` moved to a fake Word with `Later` set (added 16 Oct, the orchestrator, on the builder's question: every not-built example there is `status --line` or a 5.6.1/5.6.2 sub-word).
- **Grants:** none while others run.
- **Proof:** the lines of `done_when`.
