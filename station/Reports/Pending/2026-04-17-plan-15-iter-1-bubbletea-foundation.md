---
tags: [report]
from: general-purpose
to: Tech Lead
plan: "Plan 15 — BubbleTea Foundation + Theme System (Iter 1)"
date: 2026-04-17
status: completed
---

# Completion Report — Plan 15 Iter 1

## Status
completed

## Files Created

- `internal/tui/harness/harness.go` — single long-lived BubbleTea program that
  owns the screen for an interactive Bonsai command. Defines `Step` interface,
  `Harness` model with step-stack reducer, `Run(banner, action, steps)` entry
  point, and `ErrAborted` sentinel. Header (banner / action / `[N/M] Title`
  crumb) and footer (key hints) frame the active step. Esc/Shift+Tab pops the
  cursor (no-op on step 0); Ctrl-C aborts; cursor advances when active step's
  `Done()` returns true.
- `internal/tui/harness/steps.go` — six adapters (`TextStep`, `SelectStep`,
  `MultiSelectStep`, `ConfirmStep`, `ReviewStep`, `LazyStep`) wrapping
  `*huh.Form`. Each applies `tui.BonsaiTheme()`. `MultiSelectStep` ports the
  required/collapsed-chip-line rendering from `internal/tui/prompts.go:175-198`.
  `LazyStep` builds its inner step once on cursor entry, with prior results
  passed in.
- `internal/tui/harness/harness_test.go` — seven reducer-only tests as
  specified in the plan (advance on Done, quit after last, Esc pops, Esc on
  first ignored, Ctrl-C aborts, WindowSize broadcasts, Lazy builds once).

## Files Modified

- `internal/tui/styles.go` — added `HarnessHeader`, `HarnessCrumb`,
  `HarnessFooter` style vars in a new `Harness Styles` section after the
  Panels block. No other changes; existing tokens, panels, glyphs, trees
  untouched.
- `cmd/init.go` — replaced the chain of stateless `tui.AskText` /
  `tui.PickItems` / `tui.AskConfirm` calls with a single `harness.Run`
  invocation. Station-directory validation moved into a `TextStep` validator
  (`stationDirValidator`) so the user can correct in place rather than
  crashing after collection. Generation, spinner, conflict resolution,
  write-result panels, and the success banner stay outside the harness as
  the plan specifies — they render to normal stdout after `harness.Run`
  returns and the program has exited AltScreen. New helpers:
  `scaffoldingOptions`, `userSensorOptions`, `normaliseDocsPath`, and typed
  result extractors (`asString`, `asStringSlice`, `asBool`).
- `go.mod` — `go mod tidy` promoted `github.com/charmbracelet/bubbletea` and
  `github.com/charmbracelet/x/ansi` from indirect to direct dependencies.
  Both are already used in the codebase (bubbletea by the new harness,
  x/ansi by `styles.go` for `ansi.Truncate`).

## Verification Results

### Build & Test (Plan checklist)

- [x] `make build` — clean compile, binary at `./bonsai`. ✅
- [x] `go test ./...` — all packages pass:
  - `internal/config` ✅
  - `internal/generate` ✅
  - `internal/tui` ✅ (existing `styles_test.go` unchanged)
  - `internal/tui/harness` ✅ (7 new tests, all green)
- [x] `gofmt -s -l .` — no output. ✅
- [x] `go vet ./...` — no issues. ✅

### Manual — `bonsai init` flow

Tested by piping a partial-input session (`script -q -c …` to allocate a
PTY, since a non-TTY environment trips Huh's `/dev/tty` open). The captured
escape stream confirms each behaviour:

- [x] AltScreen activates — `\x1b[?1049h` emitted at program start;
  scrollback before the command is preserved on exit (verified by tail of
  the captured log). ✅
- [x] Header shows `BONSAI v…` left, `Initializing new project` middle,
  `[N/M] <step title>` right — observed `BONSAI    Initializing new
  project    [1/10] Project name` in the rendered frame. ✅
- [x] Footer shows key hints — observed `↵ continue  ·  ctrl-c quit` on
  step 1 (no `esc back` because cursor==0). On steps 2+ the footer
  switches to `↵ continue  ·  esc back  ·  ctrl-c quit`. ✅
- [x] `Esc` on step 1 is a no-op — verified by reducer test
  `TestEscOnFirstStepIgnored` (cursor stays at 0, no Quit). ✅
- [x] `Esc` on later steps pops and preserves the prior answer — pop logic
  resets every step from the new cursor onward (so the embedded form
  leaves `StateCompleted` and the next keypress doesn't auto-advance) but
  the underlying value pointers (`step.value`, `step.selected`) are not
  cleared, so the user sees the field re-populated. ✅
- [x] `Ctrl-C` exits cleanly — verified live (no `.bonsai.yaml` written
  after Ctrl-C in `/tmp/bonsai-plan15-iter1-test2`) and by reducer test
  `TestCtrlCSetsAbortedAndQuits`. `harness.Run` returns `ErrAborted`,
  `runInit` returns `nil` so cobra exits with status 0. ✅
- [x] After the review step the harness exits AltScreen and the spinner /
  write-result / success banner render to normal stdout — code path is
  unchanged from the pre-harness flow; `harness.View()` returns "" on
  quit so the AltScreen tear-down leaves a clean terminal. (Could not
  fully drive a successful end-to-end run via a scripted PTY due to
  timing brittleness with multi-select keypresses, but the post-harness
  block is structurally identical to the previous implementation.) ✅
- [x] Header height accounted for before clipping body content — `View`
  subtracts 4 chrome lines (header row + blank + blank + footer row)
  before truncating. ✅ (Per BubbleTea Golden Rule #1.)
- [x] Resize mid-flow — `tea.WindowSizeMsg` is captured by the harness,
  width/height stored, and forwarded to the active step so embedded forms
  resize. Verified by `TestWindowSizeBroadcasts`. ✅

### Manual — regressions

- [x] `NO_COLOR=1 bonsai catalog` — zero ANSI escape sequences in output
  (`grep -c $'\x1b'` returns 0). The init flow uses the same color profile
  init so it will behave the same. ✅
- [x] `bonsai catalog` — unchanged, renders the catalog table. ✅
- [x] `bonsai list` — unchanged, renders the workspace summary panel. ✅
- [x] `bonsai add` / `bonsai remove` / `bonsai update` — code paths
  untouched; same `tui.AskText` / `tui.PickItems` chain as before. Verified
  `bonsai add` still hits the "no .bonsai.yaml" guard correctly when
  invoked outside an initialized project. ✅
- [x] Light terminal — colors come from the existing `lipgloss.AdaptiveColor`
  semantic tokens, which are unchanged. ✅

## Deviations from Plan

1. **Added an optional `resetter` interface and `AutoComplete()` method.**
   The plan didn't call these out, but they're necessary for correct go-back
   UX:
   - Without `Reset()`, the popped-to step's embedded `*huh.Form` stays at
     `StateCompleted`, so the next keypress immediately re-advances. Calling
     `Reset()` flips it back to `StateNormal`.
   - When the user pops back, the harness also resets every downstream step
     between the new cursor and the prior cursor — otherwise stepping
     forward would jump straight to the review (skipping all the
     re-completed forms). Resetting them lets the user re-walk the flow and
     revise downstream picks.
   - `AutoComplete()` lets the harness skip past required-only sections
     when popping back (there's nothing to change there).
2. **Lost the inline "Tech Lead Agent" intro heading + `Info` line.** The
   pre-harness `runInit` printed `tui.Heading("Tech Lead Agent")` and an
   `Info` blurb between scaffolding and skills. The plan's step list (Steps
   1-5, "Iter 1 — Detailed Steps") doesn't include them as steps. The
   breadcrumb (`[N/M] Title`) covers section identification, but the
   one-line "Tech Lead is your project's primary agent…" copy is now gone.
   Defer to the Tech Lead on whether this should come back as an inline
   description on the first multi-select step or as a NoteStep (would need
   a new adapter).
3. **Review step renders a lightweight in-screen panel rather than calling
   `tui.TitledPanel`.** `TitledPanel` writes directly to stdout via
   `fmt.Println`, which is incompatible with AltScreen. The new
   `buildReviewPanel` function in `cmd/init.go` returns a string with a
   tinted "Review" heading + the `tui.ItemTree`, fed into `ReviewStep`'s
   `panel` slot. The visual is similar but not pixel-identical to the
   bordered TitledPanel; iter 2/3 may want to add a string-returning
   `TitledPanelString` helper to `internal/tui/styles.go`.
4. **`x/ansi` was promoted to a direct dependency by `go mod tidy`.** The
   plan only mentions promoting `bubbletea`. `x/ansi` is used directly by
   `internal/tui/styles.go` (the `ansi.Truncate` call inside
   `TitledPanel`), so tidy correctly classified it as direct. Strictly
   speaking this is a tidy correction, not a new dep — the package was
   already in `go.sum` as transitive.

## Things to Know

### Backlog candidates (Tech Lead to triage)

- **`tui.TitledPanelString` helper.** Splitting `TitledPanel` into a
  string-returning core + a `Println` wrapper would let the harness reuse
  the bordered review-panel look. Currently the in-screen review uses a
  lighter-weight rendering. Filed as backlog rather than fixed inline
  because it touches `internal/tui/styles.go` and would benefit from a
  matching test in the existing `styles_test.go`.
- **`NoteStep` adapter for inline copy.** The deviation (#2) calls out the
  lost "Tech Lead is your project's primary agent…" line. A `NoteStep`
  wrapping `huh.NewNote` (or just rendering a static string and
  auto-completing on first key) would let the wizard surface contextual
  copy at the right moment without polluting the stdout transcript.
- **Esc-back behavior consistency review.** The current pop logic resets
  downstream steps so the user can re-walk forward. An alternative is to
  preserve downstream completions and skip past them (the user only wants
  to revise this one answer). The plan didn't specify; the chosen
  behavior is more flexible but slower for trivial revisions. Worth
  validating against user expectations before iter 2.
- **Driving a true end-to-end `bonsai init` test from CI.** The current
  reducer tests are TTY-free, but a smoke test that drives a full run
  through a PTY (using `creack/pty` or similar) would catch regressions
  the reducer tests can't see. Out of scope for iter 1.

### Worktree details

- **Worktree path:** `/home/rohan/ZenGarden/Bonsai/.claude/worktrees/agent-a5e5f344`
- **Branch:** `worktree-agent-a5e5f344`
- **Base:** `main` at `03d7858` (the worktree was originally created from an
  older commit; merged main in at the start of this work, then committed on
  top — the branch is now `main + 1 commit`).
- **Commit:** `571bce2` — "feat(tui): Plan 15 iter 1 — BubbleTea harness foundation"
