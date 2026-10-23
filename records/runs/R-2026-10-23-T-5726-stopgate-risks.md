---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-2026-10-23-T-5726
task: T-5726
role: builder
model: opus
started: 2026-10-23 07:37
finished: 2026-10-23 08:22
outcome: verify
commits: ["30750502", "a3cdf030", "74241bac", "43f2fa64", "1a427344"]
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

- 08:15-08:21 UTC climb 1 (`/usr/local/bin/bonsai ladder --task T-5726 --json`, after a wait on the ladder's lock):
  green at 1a427344, dirty false, sha256 c9bdbc5f623950dc6a7be90290cc47f17e357b3c6b7054d9340129bef40d99a6. This report's
  last commit follows, so the final climb is at the new HEAD (below and in the report to the orchestrator).

## What is different now

An agent stopped by the stop gate now always gets the same three lines, ending with the rule's word, even when
`bonsai hook stop` is given an argument or a flag it does not take (before, that case printed the command's own
refusal with no rule line). Every rule word the stop gate and the guard can print is now checked by a test against the
words that would make Claude Code treat the block as a failed hook, so a future rule named with one fails the build.
Base's operating skill now says truly which blocks leave no log record.

## Deviations from the plan

- outputs_test.go (cmd/bonsai) held two specs of the old refusal for hook stop; they now accept, for hook stop only,
  the gate's exact frame (stopFrame). Not named in the brief's Owns; it is hook.go's test by subject.
- The unsafe-word test over the guard's rule words lives in internal/stopgate/stopgate_test.go (which already imports
  the guard's table), so no guard test file changed.

## Ladder

Climb 1, at 1a427344: green, git.dirty false, sha256 c9bdbc5f623950dc6a7be90290cc47f17e357b3c6b7054d9340129bef40d99a6.

| Rung | Check | Result | Evidence |
|---|---|---|---|
| 0 | guard | green | 10 checked, 4.9 s |
| 1 | vet | green | 1.5 s |
| 2 | schemas | green | 0.4 s |
| 3 | tests | green | 4886 seen, 9 skipped (each OS-only or release-only, reasons in the result), 62.0 s |
| 4 | tests with the fault switch | green | 2179 seen, 3 skipped, 57.5 s |
| 5 | git integrity | green | 0.035 s |

The final climb at the HEAD holding this report is quoted in the report to the orchestrator.

## What did not work

- The first BadLine wording ("the command line hook stop takes no ...") read badly; changed before the hook was
  wired, so the reason now starts with the command's own words.
- The cmd/bonsai suite went red once after the change (TestWordTables/hook_stop, TestHookRefusals): both asserted the
  old refusal for hook stop. Fixed by holding them to the frame's exact lines, not by dropping the check.
- My first climb's output file holds a stray, cut-off T-5727 header before my own document (another builder's climb
  sharing the lock); the result file itself is mine and whole.

## Evidence for a human

`printf '{}' | bonsai hook stop now` (once this lands and is installed) prints three lines ending `rule: bad-input.`
and exits 2.

## Files changed

- internal/stopgate/stopgate.go: Options.BadLine, blocked on first (bad-input, drained, no record); comments.
- internal/stopgate/stopgate_test.go: TestRuleWordsSafe (stop gate and guard rule words), TestBadLine.
- cmd/bonsai/word.go: Word.LineRefusal and refuseLine, used by the flag parser's refusals only.
- cmd/bonsai/hook.go: hook stop's bad argument and bad flags go through stopLine (the gate's frame).
- cmd/bonsai/hook_test.go, outputs_test.go: the frame's output for hook stop; only hook stop has LineRefusal.
- internal/guard/rules.go, docs/reference/lists.md: bad-input's description names hook stop's command line.
- packs/base/skills/operating-bonsai/SKILL.md: which blocks keep no log record.

## Processes left running

None of mine. Checked with `ps -eo pid,etime,cmd` at 08:21 UTC: only T-5728's climb (`bonsai ladder --task T-5728`)
and its `go test`, which are not mine.

## What it cost

About 45 minutes wall clock (07:37 to 08:22 UTC), most of it waiting on the ladder's lock. Spend not known.

## Left for next time

- A bug after 1.0: `bonsai hook guard <bad arg>` (cmd/bonsai/hook.go:107) refuses through the command's own refusal,
  not the guard's frame (no rule line); LineRefusal is the way to fix it.

## Verifier

> Added by the verifier, in a fresh context. Pass or fail, with reasons.
