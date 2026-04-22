---
tags: [report]
from: tech-lead
to: User
plan: "Plan 22 — `bonsai init` cinematic redesign"
date: 2026-04-21
status: completed
---

# Completion Report — Plan 22

## Status
completed — all 5 phases shipped, default flipped, legacy deleted.

## Scope Delivered

Full cinematic `bonsai init` — persistent chrome (header + enso rail + footer), 6 stages (Vessel → Soil → Branches → Observe → Generate → Planted), responsive to terminal size, CJK glyphs with ASCII fallback. Replaces the prior 11-step `harness.Run` form-stack. Default `bonsai init` runs this flow; legacy code path removed.

## PRs Shipped

| Phase | PR | Squash SHA | Scope |
|-------|----|-----------:|-------|
| 1 | #47 | `7553d43` | `RenderFileTree` widget + palette tokens (`ColorLeafDim`, `ColorRule`, `ColorRule2`) |
| 2 | #48 | `2e2a08c` | `internal/tui/initflow/` package (chrome / enso / fallback / stage / stub) + `harness.Chromeless` interface + `BONSAI_REDESIGN=1` env-flag routing |
| 3 | #49 | `971ee44` | `VesselStage` (3 textinputs) + `SoilStage` (hand-rolled multi-select) + `RenderHeader` stationSubdir strip |
| 3.5 | — | `380dbc5` | Direct-to-main dogfood: Bark→gold `#D4AF37`, Moon token, rail cap 60 cells + centred, 2-col field layout, focus-tinted underlines, copy refresh |
| 4 | #50 | `89c21ba` | `BranchesStage` 5-tab picker + inline-expand + defaults + `BranchesResult` |
| 4 polish | — | `413e360 399fe08 eaee416 6bb74e5 fa0ae64` | Header split `[ 盆 ]`/INITIALIZE, DETAILS box below list, `ColorAccent` ABOUT/FILE, kanji padding, `wrapToWidth` helper |
| 5A | #51 | `6baaf8e` + `3e31967` | Responsive layer: `MinTerminalWidth=70`, `ClampColumns`, hand-rolled `Viewport`, `RenderMinSizeFloor`; Branches/Vessel/Soil retrofit; `ObserveStage` wired; `GenerateStage` + `PlantedStage` packaged (not wired) |
| 5B | #52 | `5916e05` | Wire Generate+Planted into `runInit`; rename `runInitRedesign` → `runInit`; delete 245L legacy init + env-flag routing; harness `Chromeless()` delegation + `ConditionalStep.Init` auto-build of nested Lazy |

Net across all phases: `internal/tui/initflow/` new package (~3,500 LoC), `cmd/init.go` shrunk from 311 → 66 lines, `cmd/init_redesign.go` owns the new flow orchestration (206 lines), one small primitive extension in `internal/tui/harness/steps.go` (`Chromeless` + lazy-build hook, ~35 lines).

## Key Architectural Decisions

1. **Hand-rolled `Viewport` over bubbles/viewport.** 60 LoC of exact control, no new dep, matches the Soil precedent (no bubbles/list). Regression anchor `TestClampColumns_Regression` locks `(24, 44, 12)` so width changes stay auditable.

2. **`GenerateAction = func() error` (no prev injection).** The plan's literal ctor signature (`NewGenerateStage(ctx, cat, agentDef, cwd, configPath, lock, &wr, &cfg, &installed)`) would have forced 5A to know about the full generation pipeline. Instead 5A shipped `NewGenerateStage(ctx, action)` and 5B built the action closure at `cmd/init_redesign.go` where all the state already lives. Consequence: the stage is wrapped in `NewConditional(NewLazy(GenerateStage))` so the closure captures `prev[]` at build time.

3. **`ConditionalStep.Init` auto-builds nested `LazyStep`.** Without it, the `Conditional[Lazy[Generate]]` composition wouldn't fire the Lazy's builder because the harness's top-level loop only sees the Conditional. Small additive hook, guarded by `!c.skip` so a skipped conditional never builds its inner.

4. **Two predicates over one.** Plan prescribed a single `plantedConfirmed` reading `prev[3]`. Splitting into `observeConfirmed` (gates Generate) + `generateSucceeded` (= observeConfirmed + `prev[4]` not a non-nil error; gates conflicts splice + Planted) means Generate-failure cleanly skips Planted instead of displaying an incomplete "Planted" screen after error.

5. **Min-size floor at 70×20.** Below that, every stage short-circuits to `RenderMinSizeFloor` instead of painting a clipped frame. Inclusive: `(70, 20)` renders; `(69, 20)` or `(70, 19)` floors.

## Dogfooding Practice (UX Convention)

6 direct-to-main polish commits (3.5 palette refresh + 5 Phase-4 layout fixes) shipped **without PRs**, per the project's fast-iterate UX convention encoded in `agent/Core/memory.md` Feedback. This is not a security/process violation — it's the explicit tradeoff the user chose for visual-polish iterations where PR overhead outweighs review value. All other substantive code changes went through full PR + CI + independent-review.

## Minor Items Deferred to Backlog

Both surfaced by the independent review of PR #52 as non-blocking:

1. **`NewConditional(NewLazy(...))` composition test.** Covered end-to-end by the init flow but no focused unit test asserts the delegation semantics.
2. **Dead post-harness Generate-error warning.** Unreachable in practice (the in-frame `stateError` panel owns error display); either delete the block or delete the in-frame panel.

Both filed under Backlog P2 Group B (Code Quality & Testing).

## Test Coverage

~50+ new tests landed in Phase 5A alone. All 6 CI checks green on both PRs:
- `test` / `Analyze Go` / `lint` / `govulncheck` / `CodeQL` / `GitGuardian Security Checks`

Regression anchors of note:
- `TestClampColumns_Regression` — `ClampColumns(120) == (24, 44, 12)`
- `TestClampColumns_TagPinned` — tagW==12 across width ≥ 50
- `TestViewport_FollowClampsOffset` — focus always in `[offset, offset+height)`
- `TestBranches_NarrowDoesNotClipTag` / `TestSoil_NarrowDoesNotClipBadge` / `TestVessel_ResponsiveInputWidth`
- `TestPlanted_TreeFromWriteResult` — Created→NEW, Updated→NEW+"UPDATED" note, Unchanged→Normal, Skipped/Conflict omitted
- `TestGenerate_MinHoldEnforced` — 600ms min honored on instant-complete path

## Handoff Artifacts

- Plan archived → `station/Playbook/Plans/Archive/22-init-redesign.md`
- Status.md → Plan 22 row moved In Progress → Recently Done (two rows: 5A + 5B)
- Backlog.md → 2 follow-up items added under P2 Group B
- memory.md → Work State updated (no current task, main at `5916e05`)

## Main SHA at Completion

`5916e05483220c5f4f5b436fa1e8eaa8dd9b394d`
