---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-2026-10-23-T-5731
task: T-5731
role: builder
model: opus
started: 2026-10-23 09:36
finished: null
outcome: running
commits: ["1ec9b0b4", "51403d52", "7fce5791"]
labels: {}
---

# Run: the stop gate reads the run report as bonsai check does (T-5731)

> Opened at the start of the session and added to as it goes. Not written at the end.

## Log

- 09:36 UTC, builder (opus), brief M-2026-10-23-5e0a1b83. Worktree ~/Servers/Bonsai-T-5731 on T-5731 at 1cbd0aa4
  (T-5730's top). Read: the brief, the task, the pi brief's B3, internal/stopgate/stopgate.go (runReport),
  internal/engine/checkdocs.go (checkDocument, docRead), internal/format/task.go (ReadRun), formats/schemas/run.schema.json.
- Probe (a throwaway test in internal/format, deleted): ReadRun on a report with no frontmatter reads as format 0 with
  no error (check's finding there is format0-new, from the lock's format0 list); `started: 2026-10-21` is refused
  (`line 7: field started ... does not match`, check's finding document); a format-1 report with no `outcome` line
  reads with no error: readers hold documents to the lenient schema (schema.Lenient), so check does not find a missing
  field. Null `started`/`finished` and `outcome: running` read fine.
- Every existing gate fixture (internal/stopgate, internal/guard/stop_test.go, internal/ladder) writes a format-0
  report ("# R-...") in a project with no lock: check finds nothing there, so a gate that judges as check does keeps
  them green unchanged.
- Audit sent as needs-input (M-2026-10-23-c615b676); waiting for go.
- Go (M-2026-10-23-2eaca622): option A (the gate judges exactly as check does; done_when reworded by the
  orchestrator to "any finding check makes on the report file itself"); every run report of the task judged; no lock:
  a format-0 report passes, as in check. internal/engine/checkdocs.go and checkdocs_gate_test.go added to Owns.
- Tests first. With only the rule word added (constant and rules.go row, no behaviour), the new stopgate tests on the
  old gate: `--- FAIL: TestARunReportCheckRejectsHoldsTheGate` (6 subtests, each `exit 0, want 2
  (stop-bad-run-report)`), `--- FAIL: TestEveryRunReportOfTheTaskIsJudged` (2 subtests, the same). The four
  allow subtests (good format 1; format 0 on the lock's list; no lock; another task's bad report) pass before and
  after: they guard against over-blocking. The engine parity test failed first to build (`undefined:
  DocumentFindings`).
- First parity run: the absolute-path case read as `document` (an unquoted /path is a plain value YAML refuses);
  quoted the value in both tests; then green. Test paths use /opt/model, no home-folder path.
- internal/guard's TestEveryRule wants a test per rule row: added one case to internal/guard/stop_test.go's table
  (a report with `started: 2026-10-11`). That file is not in the brief's Owns list: an added case, nothing changed or
  weakened.
- go vet ./... clean, gofmt clean, go test ./... all ok (WSL, once). go generate ./internal/reference: lists.md +1 line.
- 09:44 UTC, climb 1 at 899982a1 (sha256 13c60c7b...): red at rung 0 only, "2 of 80 changed paths refused":
  .gitignore and CLAUDE.md, both T-5730's changes (my branch starts on T-5730; rung 0 judges against main at
  7e172040, under T-5731's empty grant). Nothing of mine; rungs 1-5 skipped. Waiting for main to reach T-5730
  (its CI was finishing), then climbing again. Not reverting T-5730's work.

## What is different now

A named session (Claude Code or pi) can no longer stop on a run report `bonsai check` would reject. The gate reads
each of the task's run reports in the session's checkout through check's own reading, and the first finding check makes
on a report sends the session back (`stop-bad-run-report`). The reason quotes check's finding, which names the field
and line; the next step is check's own, then `bonsai check`, then the ladder. Examples: a report with no `format:`
line (format0-new), or `started: 2026-10-11` with no time (document). A report check accepts lets the gate go as
before. A report with `finished: null` and `outcome: running` passes too.

## Deviations from the plan

One file outside the brief's Owns: internal/guard/stop_test.go gains one case for the new rule (guard's TestEveryRule
needs a test for every rules.go row). Nothing in it was changed or removed.

## Ladder

Not climbed yet.

## What did not work

- The first absolute-path fixture was an unquoted value, which YAML reads as a document finding, not absolute-path;
  quoted.
- Seen, not fixed (not this task's): the gate's noPaths swallows the space before "(" after a path it hides
  ("a path on this machine(contract section 2.6 ..."), a cosmetic flaw of the old helper.
- Not done, by the director's choice (A): a format-1 report with a field missing passes, as it passes check.

## Files changed

- internal/engine/checkdocs.go: DocumentFindings, check's checkDocument on one file with the lock and labels in force.
- internal/engine/checkdocs_gate_test.go (new): parity with Check for 5 reports.
- internal/stopgate/stopgate.go: RuleBadRunReport; runReport judges every report of the task (checkedReport); step 5
  of the package comment.
- internal/stopgate/runreport_test.go (new): 6 block cases, 2 every-report cases, 4 allow cases.
- internal/guard/rules.go: the rule's row; internal/guard/stop_test.go: its case.
- docs/features.md (the stop gate's row), docs/reference/lists.md (regenerated), packs/base/skills/operating-bonsai/SKILL.md
  (the ladder paragraph).

## Processes left running

(after the build)

## What it cost

Not known.

## Left for next time

(after the build)
