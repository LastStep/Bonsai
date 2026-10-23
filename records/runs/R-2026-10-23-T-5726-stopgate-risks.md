---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-2026-10-23-T-5726
task: T-5726
role: builder
model: opus
started: 2026-10-23 07:37
finished: null
outcome: running
commits: []
labels: {}
---

# Run: The stop gate's three risks from the trial's verifier

> Opened at the start of the session and added to as it goes. Not written at the end.

## Log

- 07:37 UTC, builder (opus), brief M-2026-10-23-6376559e. Worktree ~/Servers/Bonsai-T-5726, branch T-5726 at
  7e172040. Read the task, the brief, the common brief of 23 Oct, the trial verifier's risks (M-2026-10-20-57b77ed9).
- Audit of the code (no edit yet):
  - Risk 1: `Decision.answer()` (internal/stopgate/stopgate.go:233) passes `d.Rule` unfiltered; `safe()` covers Why
    and Next only. `ruleConsts(t)` (stopgate_test.go:823) already reads every `Rule...` constant from the package's
    source, so a test over it covers every rule, not only those a test blocks with.
  - Risk 2: SKILL.md lines 103-104 say the rule word is "the one the log keeps". Blocks that keep no record in the stop
    gate: no-project (both, stopgate.go:257, 265), bad-config of the session's own bonsai.yaml (277), over-time before
    the workspace is known (899), log-failed (its record is what failed). The guard: bad-config (session's own) and
    no-project, by its rules table (internal/guard/rules.go:199-200).
  - Risk 3: two paths refuse `hook stop`'s command line outside the frame: a positional argument (cmd/bonsai/hook.go:144)
    and any flag but --help (the generic parser, cmd/bonsai/word.go:278), e.g. `hook stop --json`. The existing row
    in cmd/bonsai/hook_test.go:433 asserts today's bad-flag output and must change with the fix.
- 07:45 UTC: Audit sent (M-2026-10-23-be9f7f23, needs-input).
- Go: M-2026-10-23-7b81140d (orchestrator): Q1 A (cmd/bonsai/word.go granted, the hook only), Q2 A
  (internal/guard/rules.go's bad-input text widened, docs/reference regenerated); add the unsafe-word test over the
  guard's rule words too, test only. A later note: T-5727's agent markers; no test of mine runs a person-only
  command, so nothing to change.
- Fail-first, TestRuleWordsSafe and TestBadLine (scratch constant `RuleScratch = "stop-file-missing"` and a bare
  `BadLine` field added to stopgate.go, then the file restored from a copy):

  ```
  --- FAIL: TestRuleNames (0.00s)
      stopgate_test.go:813: the stop gate's "stop-file-missing" is not in the guard's rules table
  --- FAIL: TestRuleWordsSafe (0.00s)
      stopgate_test.go:925: the stop gate's rule word "stop-file-missing" is not safe to print: it says a file is missing as an operating system says it, or it is not one plain ASCII word
  --- FAIL: TestBadLine (0.02s)
      stopgate_test.go:961: exit 0, want 2 (bad-input): ""
  ```

  Every real constant (stop gate's and guard's) passes the filter: nothing unsafe found.
- Fail-first, TestHookCommand's rows on the old cmd/bonsai (hook.go and word.go unchanged):

  ```
  hook_test.go:442: [hook stop now]: exit 2, stdout "", stderr "bonsai hook stop: refused  [bad-flag]\n  what  exit 2, block\n  why   hook stop takes no \"now\"\n  next  use the hook line ..."
  hook_test.go:442: [hook stop --json]: ... [bad-flag] ...
  hook_test.go:442: [hook stop --x=1 now]: ... [bad-flag] ...
  ```
- Fail-first, TestOnlyHookStopRefusesItsLineItsOwnWay (scratch: LineRefusal set on hook guard, then hook.go restored):
  `hook_test.go:462: hook guard: LineRefusal set true`.
- Two existing specs in cmd/bonsai/outputs_test.go asserted the old refusal for hook stop (TestWordTables'
  `--not-in-the-table` row, TestHookRefusals' `hook stop now`): both now accept, for hook stop only, the frame's
  exact three lines (stopFrame: reason "hook stop takes no ...", a next step, `rule: bad-input.`). The changed spec,
  as the go named for hook_test.go:433; flagged here since outputs_test.go is hook.go's test by subject, not by name.
- Commits: 30750502 (stopgate), a3cdf030 (hook + word.go), 74241bac (guard rules text + lists.md, regenerated with
  `go run ./internal/reference/gen .`), 43f2fa64 (SKILL.md).
- `go vet ./...` clean; `go test ./... -count=1`: every package ok.

## What is different now

(not yet)

## Deviations from the plan

(not yet)

## What did not work

(not yet)

## Processes left running

(not yet)

## What it cost

(not yet)
