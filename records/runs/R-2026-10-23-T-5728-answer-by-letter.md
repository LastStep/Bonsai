---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-23-T-5728
task: T-5728
role: builder
model: sonnet
started: 2026-10-23 07:38
finished: 2026-10-23 07:56
outcome: verify
commits: []
labels: {}
---

# Run: T-5728, answer a Decide by its option's letter

- Who: the builder (sonnet)
- Why: the director, 20 Oct: answer by A, B etc instead of words. Briefs M-2026-10-23-eace70c4, go M-2026-10-23-7d5f942c.

## Log

- 2026-10-23 07:38 Audit sent (M-2026-10-23-f76efbeb); go received. Building as audited.
- 07:43 Tests first: TestAnswerByLetter failed to build (Letter undefined), TestAskCommands failed on "option B" and the letter answer (exit 2). Built: asks.Letter, byPlace/lettered in file.go, statusText letters, answer help. Two existing assertions moved with the spec: ask_test.go option line (option 2 -> option B), and the order/count in TestAskCommands (7 -> 9 records, agent:size added by the new letter case). The asks_test.go:432 prefix kept. docs/features.md rows, operating-bonsai skill line; internal/reference/gen produced no diff.

## Left

Nothing. Processes started: go test runs and the ladder climb, all finished; checked with ps. No process left running.
Rebase on main at landing and regenerate docs/reference (it produced no diff here).
