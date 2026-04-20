---
tags: [report]
from: general-purpose
to: Tech Lead
plan: "Plan 15 — BubbleTea Foundation + Theme System (Iter 2)"
date: 2026-04-20
status: completed
---

# Completion Report — Plan 15 Iter 2

## Status
completed

## Scope Delivered

Migrated `cmd/add.go` onto the iter-1 BubbleTea harness. Both the "configure
new agent" and "add items to existing agent" flows now live in a single
`harness.Run` invocation with a `LazyGroup` splice selecting the branch at
runtime. Added the two primitives called out by iter 1 follow-ups
(`NoteStep`, `TitledPanelString`) plus the harness extension they require
(`splicer` interface + `expandSplicer` reducer step).

## Files Modified

- `cmd/add.go` — full rewrite onto harness. Tech-lead-required check stays
  pre-harness (`add.go:120-124`). In-harness step list is
  `SelectStep` (agent type) + `LazyGroup` (branch). Two builders:
  `buildNewAgentSteps` (workspace `LazyStep` branching on tech-lead +
  5 multi-selects + review) and `buildAddItemsSteps` (filter counts →
  either single "all installed" `NoteStep` or intro-note + N multi-selects
  + review). Post-harness finalisers `finaliseNewAgent` / `finaliseAddItems`
  preserve the exact write pipeline (`EnsureRoutineCheckSensor` → save →
  spinner → generate → `resolveConflicts` → `lock.Save` → `showWriteResults`
  → success banner).
- `cmd/init.go` — `buildReviewPanel` now uses `tui.TitledPanelString("Review",
  tree, tui.Water)`; shared helpers (`newDescriber`, `userSensorOptions`,
  `asString`/`asStringSlice`/`asBool`) lifted out so add.go can reuse them.
- `internal/tui/harness/harness.go` — new `splicer` interface + `expandSplicer`
  helper invoked in `Init()` (first-step-is-splicer case) and after each
  `cursor++` in the advance loop. Idempotent via `Spliced()` guard.
- `internal/tui/harness/steps.go` — `LazyGroup` adapter (Step interface stub;
  `Splice(prev) []Step` does the work) + `NoteStep` adapter following the
  `buildForm()`+rebuild-on-Reset pattern that guards the iter-1 huh-quitting
  regression.
- `internal/tui/styles.go` — `TitledPanel` split into `TitledPanelString`
  (returns string, for in-AltScreen rendering) + `TitledPanel`
  (one-line stdout wrapper). Existing callers' behaviour preserved.
- Tests: `internal/tui/harness/harness_test.go` (+3 LazyGroup reducer tests),
  `internal/tui/harness/steps_test.go` (+2 NoteStep tests including
  Reset-restores-view), `internal/tui/styles_test.go` (+3 TitledPanelString
  tests including byte-for-byte parity with the stdout wrapper).

## Verification Results

Run on `ui-ux-testing` at `4011882`:

- `go build ./...` — clean ✅
- `go vet ./...` — clean ✅
- `gofmt -s -l .` — no output ✅
- `go test ./... -count=1` — all 4 test packages pass ✅

### Reviewer findings (independent pass, APPROVE)

Four non-blocking nits, all unreachable from current callers:

1. Nested splicers silently unsupported — if `Splice()` returns a slice
   whose head is itself a splicer, the second splicer is driven as a live
   step. Not hit by any caller; worth a docstring note.
2. Empty-splice re-entry — if `Splice()` returns `[]`, the cursor lands on
   former `cursor+1` without another `expandSplicer()` call. Same shape as
   #1, safe for current callers.
3. `WindowSizeMsg` not re-broadcast after splice — huh forms adapt on their
   next update cycle, so first-frame layout on the spliced step *may* be
   briefly wrong at narrow terminal widths. Not a regression from iter 1.
4. No unit test for `LazyStep` nested inside a `LazyGroup` result (the
   workspace `NewLazy` exercises this at runtime via `cmd/add.go`, but
   there's no isolated harness test).

Filed as iter-3 follow-ups in the plan's "Iteration 3" section — none block
iter 2 shipping.

## Deviations from Plan

1. **`expandSplicer` as a dedicated helper.** Plan sketched the splicer
   reducer inline in the advance loop; the implementation factors it into a
   named helper so it can also be called from `Init()` (covers the case
   where step 0 is a splicer — otherwise the harness would quit immediately
   without rendering). Cleaner than the plan's inline sketch.
2. **Helper extraction is tighter than described.** Plan listed `newDescriber`
   as the one factoring-out; the implementation also lifted
   `userSensorOptions`, `asString`, `asStringSlice`, `asBool` out of
   `cmd/init.go` into (effectively) package-shared helpers so `cmd/add.go`
   can reuse them. All within the "eliminate duplication that's now
   duplicated between init.go and add.go" spirit of the plan.
3. **`TitledPanelString` got its own test file entry (`styles_test.go`
   additions).** Plan didn't explicitly call this out, but since the split
   is a refactor of user-facing rendering, defensive test coverage
   (byte-for-byte parity with `TitledPanel`) is cheap and valuable.

## Manual Smoke

Skipped by the implementing agent — no PTY in the dispatched environment.
`bonsai --help` ran cleanly; the binary builds. **Follow-up for Tech Lead:**
run `bonsai init && bonsai add` locally to walk the flow before merging
the whole `ui-ux-testing` branch back to main (at iter 3 completion).

### Worktree Details

- Worktree path: `/home/rohan/ZenGarden/Bonsai/.claude/worktrees/agent-a12e0a9b`
- Branch: `worktree-agent-a12e0a9b`
- Base: `ui-ux-testing` at `ae86c24` (fast-forwarded after worktree creation)
- Commit: `4011882` — "feat(tui): Plan 15 iter 2 — migrate bonsai add onto harness"
- Merged into `ui-ux-testing` via `git merge --ff-only`; worktree branch
  should be pruned after this report is filed.
