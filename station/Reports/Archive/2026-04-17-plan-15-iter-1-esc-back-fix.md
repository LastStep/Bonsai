---
tags: [report]
from: general-purpose
to: Tech Lead
plan: "Plan 15 — BubbleTea Foundation (iter 1 follow-up)"
date: 2026-04-17
status: completed
---

# Completion Report — Plan 15 iter 1 Esc-back fix

## Status
completed

## Files Created
- `internal/tui/harness/steps_test.go` — per-adapter guards for the Reset() bug. Each test forces `form.State = StateCompleted` (the post-submit state), calls Reset(), and asserts View() is non-empty. Direct guard against future regressions of the same shape.

## Files Modified
- `internal/tui/harness/steps.go` — refactored every adapter (TextStep, SelectStep, MultiSelectStep, ConfirmStep, ReviewStep) to factor form construction into a private `buildForm()` method. Reset() now rebuilds the *huh.Form rather than flipping `form.State` back. Construction args (prompt, validators, options, defaultVal) promoted to struct fields so the builder can re-run. For MultiSelectStep, the previously closure-captured `optionalSelected` slice was promoted to a struct field (`s.optionalSelected`) so `buildForm` and `SubmitCmd` both reach the same destination. Re-entry re-applies `Selected(true)` to options matching prior picks (or defaults on first build).
- `internal/tui/harness/harness_test.go` — added `fakeResetStep` (embeds `fakeStep`, satisfies `resetter`, counts calls, returns sentinel cmd) and `TestEscPopReinitsActiveStep`. The new test verifies the reducer calls Reset() exactly once on the popped-onto step (and not on steps before the new cursor or at the original cursor) and batches the returned tea.Cmd into the harness's return.

## Verification Results
- `make build` — clean
- `go test ./...` — all packages pass; harness package now has 13 tests (7 original reducer + 1 new reducer + 5 new step adapter tests), all green
- `gofmt -s -l .` — no findings
- `go vet ./...` — no findings
- Manual `bonsai init` walkthrough — **could not perform** in this worktree session because the agent shell has no TTY (`Error: could not open a new TTY: open /dev/tty: no such device or address`). Same caveat as the original iter 1 verification. The reducer test (`TestEscPopReinitsActiveStep`) plus the per-adapter form-rebuild tests in `steps_test.go` provide the strongest evidence we can give without a PTY: each adapter's Reset() now produces a non-empty View() even when the prior form was in StateCompleted (which is exactly the failure mode at the huh layer).

## Deviations from Plan
- The fix-instruction expected `Reset()` to be the single touchpoint. Per-adapter implementation needed *additional* struct fields (TextStep.prompt/required/validators, SelectStep.prompt/options, ConfirmStep.prompt, ReviewStep.prompt, MultiSelectStep.optionalSelected) so the builder can re-run with the original args. No changes to public API, the Step interface, or `cmd/init.go`.
- Added `internal/tui/harness/steps_test.go` (5 adapter tests) on top of the harness_test.go addition the instructions specified. Spec said "the reducer-level test is sufficient" — I agreed but added the per-adapter guards because they directly assert the observable contract (View() non-empty post-Reset) and would catch any future adapter that forgets to wire the rebuild path. They take <50ms total. Removable if you'd prefer the slimmer test set.
- Discovered (via reading huh's `field_multiselect.go`) that huh's MultiSelect calls `updateValue()` on Focus, which writes the currently-selected options back into the value-pointer slice. This means after the harness fires the post-Reset `tea.Cmd` (Init -> Focus), `s.optionalSelected` is restored to the prior picks automatically. `TestMultiSelectStepResetPreservesPicks` documents this contract.

## Things to Know
- Worktree path: `/home/rohan/ZenGarden/Bonsai/.claude/worktrees/agent-a3aa71c2`
- Branch: `worktree-agent-a3aa71c2`
- Commits on top of main:
  - `a26c333` — cherry-picked iter 1 feature commit (originally `571bce2` on `worktree-agent-a5e5f344`)
  - `8f55129` — this fix
- Total tests in `internal/tui/harness/` now: 13 (was 7 — all 7 originals still pass).
- No changes to `cmd/init.go`, `internal/tui/harness/harness.go` reducer logic, or any catalog/generator code.
- The TTY caveat means the only outstanding verification is "user runs `bonsai init` in a real terminal, walks to step 3+, presses Esc, sees the prior step's form re-render with prior input visible." If this fails for any reason in your own check, please report back — the test suite proves the form-View() output is non-empty post-Reset, but a real terminal interaction can still surface something the unit tests miss.
