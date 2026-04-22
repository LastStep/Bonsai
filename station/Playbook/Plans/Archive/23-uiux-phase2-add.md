# Plan 23 — UI/UX Overhaul Phase 2: `bonsai add` Cinematic Port

**Tier:** 2 (Feature)
**Status:** Draft
**Agent:** general-purpose (worktree isolation)
**Source:** User request 2026-04-22 — Plan 22 complete; UI overhaul moves from `init` to the remaining user-facing commands. `bonsai add` is iteration 1 (closest shape match to init, highest post-init traffic).

---

## Goal

Replace the `bonsai add` harness+Huh flow with a cinematic AltScreen flow that matches `bonsai init`'s visual language: persistent chrome, enso progress rail, chromeless stages, responsive resize, kanji/kana labels with ASCII fallback, and a terminal completion stage. All behavioural guarantees of the current `runAdd` (two branches — new-agent vs add-items, conflict resolution, all-installed short-circuit, tech-lead gate) must be preserved bit-for-bit.

Ship in **3 PRs**. Phases 1 and 2 land behind `BONSAI_ADD_REDESIGN=1` so the default binary keeps the current `runAdd` throughout development; Phase 3 flips the default and deletes the legacy path.

---

## Context

### Why

Plan 22 shipped the cinematic init flow (6 stages + chromeless harness + `internal/tui/initflow/` package + design-system tokens in `internal/tui/styles.go` + `internal/tui/initflow/design.go`). The remaining user-facing commands (`add`, `update`, `remove`, `list`, `catalog`, `guide`) still use the old palette-consistent-but-flat Huh flow. Dogfooding feedback and Group F backlog closure (2026-04-22) confirm the init pattern is the target. `add` is first because it's the most-trafficked command post-init and its shape (agent pick → workspace/branches → review → generate → completion ± conflicts) maps directly onto init's stage sequence.

### Current state (anchored to code read 2026-04-22)

- `cmd/add.go` composes four `harness.Step`s: `NewSelect` (agent) → `NewLazyGroup` (agent-flow splicer) → `NewConditional(NewSpinnerWithPrior)` (generate) → `NewLazyGroup` (conflict picker). Two branches:
  - **new-agent** (`buildNewAgentSteps`): workspace `NewText` → 5 `NewMultiSelect` pickers (Skills/Workflows/Protocols/Sensors/Routines) → `NewLazy(Review)` → spinner → optional conflict picker.
  - **add-items** (`buildAddItemsSteps`): intro `NewNote` → filtered `NewMultiSelect` pickers (only categories with uninstalled items) → `NewLazy(Review)` → spinner → optional conflict picker.
- `runAdd` also handles: all-installed short-circuit (`availableAddItems(cat, installed).Total() == 0` → `EmptyPanel` post-harness), tech-lead-required gate (non-tech-lead agent pick with no tech-lead installed → `ErrorDetail` post-harness), and post-harness success banner + tree (`showWriteResults` + `tui.Success` + `tui.Hint`).
- `cmd/remove.go` and `cmd/update.go` use `buildConflictSteps` / `applyConflictPicks` (defined in `cmd/root.go`). Same primitives must be reached by the cinematic conflict stage.
- Design system available for reuse (verified 2026-04-22, commit `027ddd0`):
  - `internal/tui/styles.go` — palette (Leaf/Bark/Stone/Water/Moss/Ember/Amber/Sand/Petal/Moon/Slate) + semantic tokens (ColorPrimary/Secondary/Accent/Subtle/Muted/Success/Danger/Warning/Info/Surface) + chrome tokens (ColorLeafDim/Rule/Rule2) + Banner/Panels/Glyphs helpers.
  - `internal/tui/initflow/design.go` — `PanelContentWidth=84`, `PanelWidth(termW)`, `RenderSectionHeader`, `RequiredGlyph`, `FocusBorder/UnfocusBorder`, `FocusedNameStyle/UnfocusedNameStyle/FocusedDescStyle/UnfocusedDescStyle`, `ValueStyle`, `LabelStyle`, `DimStyle`, `FocusedPrimaryStyle`, `FocusedAccentStyle`.
  - `internal/tui/initflow/chrome.go` — `RenderHeader(title, stationSubdir, width)`, enso rail, footer, `RenderMinSizeFloor(w,h)`.
  - `internal/tui/initflow/layout.go` — `MinTerminalWidth=70`, `MinTerminalHeight=20`, `TerminalTooSmall`, `ClampColumns(termW)`, `Viewport`.
  - `internal/tui/harness/` — `Chromeless` variant for AltScreen full-takeover stages; `LazyStep` / `ConditionalStep` / `LazyGroup` compose as in init (Plan 22 Phase 5B additions).

### Design decisions locked

| # | Decision | Picked |
|---|----------|--------|
| 1 | Package location | **B** — new `internal/tui/addflow/` package that imports `initflow/` for chrome + design tokens. Rationale: flows share chrome but differ in stage count/shape; keeping them in separate packages prevents test/resize matrix bloat in `initflow/`. If a third command later shares identical chrome, lift to `internal/tui/stages/` as a rename pass. |
| 2 | Kanji / kana labels | **A** — six labels in bonsai-raising metaphor: 選 Select · 地 Ground · 接 Graft · 観 Observe · 育 Grow · 結 Yield. Same runewidth fallback as init. Two-char ASCII fallback: `SEL`, `GND`, `GFT`, `OBS`, `GRW`, `YLD`. |
| 3 | Stage sequence (new-agent branch) | Select → Ground → Graft → Observe → Grow → [Conflicts (conditional)] → Yield. 7 slots (6 default + 1 conditional). |
| 4 | Stage sequence (add-items branch) | Select → Graft (filtered) → Observe → Grow → [Conflicts] → Yield. 6 slots (5 default + 1 conditional). Ground (workspace) is skipped. Enso rail adapts its segment count to the active branch. |
| 5 | Select stage | Single-choice picker with inline description on focus (same leaf-border/white-name treatment as Branches). List of all catalog agents plus an "already installed" indicator. |
| 6 | Ground stage | Textinput for workspace path (Vessel-like). Tech-lead auto-fills `cfg.DocsPath` or `station/` and the stage auto-completes with no user input. Skipped entirely on the add-items branch. |
| 7 | Graft stage | Branches-like tabbed picker with the same 5 tabs (Skills/Workflows/Protocols/Sensors/Routines). Filtering: on new-agent branch show all available with agentDef defaults; on add-items branch show only uninstalled + drop empty tabs entirely from the tab row. Per-tab counts update live. |
| 8 | Observe stage | ReviewPanel layout (NAME / WORKSPACE / AGENT + 5 category blocks + plant CTA) — same pattern as init's Observe. The `priorAware.SetPrior` hook captures agent/workspace/picks from prior stages exactly as init does. |
| 9 | Grow stage | Same `GenerateStage` as init — reuse the struct by parameterising the action closure. Keep the `max(actual, 600ms)` hold. The action closure runs `runAddSpinner`'s body (renamed / factored out of `cmd/add.go`). |
| 10 | Conflicts stage | **New.** Tabbed-picker variant — one tab per conflict file, each showing diff summary + [Keep / Overwrite / Backup] radio. New design token: `ConflictRowStyle` in `initflow/design.go` (exposed for addflow to import). Conflict stage is a `ConditionalStep` gated on `wr.HasConflicts()`. |
| 11 | Yield stage | Planted-like completion — shows what was added (tree of installed items under the agent + workspace path) + 3 next-step commands (`bonsai list`, `cd <workspace>`, session-start hint). Reuse `PlantedStage` struct by parameterising its summary builder. |
| 12 | All-installed short-circuit | When `availableAddItems(cat, installed).Total() == 0` on the add-items branch, skip every stage after Select and render a dedicated "already full" Yield variant (still in AltScreen, still cinematic) instead of the post-harness EmptyPanel. |
| 13 | Tech-lead gate | When user picks a non-tech-lead agent and no tech-lead exists, skip to a "missing tech-lead" Yield variant with a "Run: bonsai init" CTA. Still in AltScreen. |
| 14 | Post-harness output | All user-facing output moves inside AltScreen (Yield stage). `showWriteResults` + `Success` + `Hint` are subsumed. Error paths (spinner failure, save failure) surface in the Grow stage's error panel (same pattern as init's GenerateStage). |
| 15 | Esc semantics | Stage-level back — same as init. Graft → Ground (new-agent) or → Select (add-items). Observe → Graft. Grow does not accept back (in-flight). Yield is terminal. |
| 16 | Ship strategy | **B** — phased, 3 PRs, env-flag gated until Phase 3. |
| 17 | Env flag | `BONSAI_ADD_REDESIGN=1` selects the cinematic path; default = legacy `runAdd` through Phase 1 + 2. Phase 3 flips default and deletes legacy. |

---

## Phases

Each phase is an independent PR. Phases 1 and 2 land behind `BONSAI_ADD_REDESIGN=1`. Phase 3 flips the default and deletes the legacy path.

### Phase 1 — `addflow/` foundations + Select + Ground + new-agent Graft/Observe/Grow/Yield

**Scope:** Create the `internal/tui/addflow/` package, implement the new-agent branch end-to-end behind the env flag. Add-items branch still routes through legacy `runAdd`.

#### Files touched

- `internal/tui/addflow/addflow.go` — **new.** Package doc comment + any shared types (e.g., `AgentOption`, `GraftResult{Skills,Workflows,Protocols,Sensors,Routines []string}`).
- `internal/tui/addflow/select.go` — **new.** `SelectStage` single-choice agent picker. Mirrors `BranchesStage`'s focus/border/kanji conventions but with one tab only.
- `internal/tui/addflow/select_test.go` — **new.** Cover: focus clamp, enter completes, esc aborts (top of flow), installed indicator rendering, result exposes agent name.
- `internal/tui/addflow/ground.go` — **new.** `GroundStage` workspace textinput (Vessel pattern). Tech-lead auto-complete path (skips the stage with no keystroke via an `AutoComplete()` method mirroring `GenerateStage.AutoComplete()`).
- `internal/tui/addflow/ground_test.go` — **new.** Cover: textinput accepts valid workspace, validator rejects duplicate (via injected `existingWorkspaces`), tech-lead branch auto-completes with DocsPath or `station/`.
- `internal/tui/addflow/graft.go` — **new.** `GraftStage` tabbed picker. Ctor signatures:
  ```go
  func NewNewAgentGraft(ctx *GraftContext) *GraftStage       // all categories, agentDef defaults
  func NewAddItemsGraft(ctx *GraftContext) *GraftStage        // filtered to uninstalled, no defaults
  ```
  Context struct holds catalog, agent type, agentDef, installed (nil on new-agent), and computed tab metadata.
- `internal/tui/addflow/graft_test.go` — **new.** Cover: new-agent uses defaults, add-items filters uninstalled, empty-category tabs dropped (add-items path), per-tab count updates live, Reset preserves state across esc-back.
- `internal/tui/addflow/observe.go` — **new.** `ObserveStage` — same ReviewPanel layout as `initflow/observe.go` but with AGENT block instead of STATION, and the same `priorAware.SetPrior` hook.
- `internal/tui/addflow/observe_test.go` — **new.** Cover: SetPrior capture (agent/workspace/picks), [GRAFT]/[BACK] buttons, 2-col layout at ≥100 cols.
- `internal/tui/addflow/grow.go` — **new.** Thin wrapper around `initflow.NewGenerateStage` that parameterises the action closure. Keeps Kanji as 育 Grow (not 生 Generate).
- `internal/tui/addflow/yield.go` — **new.** `YieldStage` Planted-equivalent. Three modes: success (installed items + next steps), all-installed (warm "already full" copy + `bonsai catalog` CTA), tech-lead-required (error + `bonsai init` CTA).
- `internal/tui/addflow/yield_test.go` — **new.** Cover: 3 modes render distinct bodies, next-steps rendered inline.
- `internal/tui/initflow/design.go` — extend with `ConflictRowStyle()` helper (unused in Phase 1 but exposed so Phase 2 doesn't pad the diff).
- `cmd/add_redesign.go` — **new.** `runAddRedesign` entry point — builds the cinematic step list and routes through `harness.Run`. Branch picker reads the Select result and chooses new-agent vs add-items splicer. Gated by `BONSAI_ADD_REDESIGN=1`.
- `cmd/add.go` — add the env-flag branch at top of `runAdd`: `if os.Getenv("BONSAI_ADD_REDESIGN") == "1" { return runAddRedesign(cmd, args) }`. No other changes in Phase 1.

#### Steps

1. **Create `internal/tui/addflow/` package.** Mirror `initflow/`'s layout: `addflow.go` (types), `select.go`, `ground.go`, `graft.go`, `observe.go`, `grow.go`, `yield.go`, plus `_test.go` for each. Every stage implements `harness.Step` and `harness.Chromeless`. Every stage imports `initflow` for chrome (`RenderHeader`, enso rail, `RenderMinSizeFloor`, `ClampColumns`, `Viewport`, `PanelContentWidth`, `RenderSectionHeader`, `FocusedNameStyle`, etc.). Do NOT recreate any design primitive — consume `initflow/design.go` tokens exclusively.

2. **Select stage.** Single-choice picker. One column, same `FocusedNameStyle` / description pattern as Branches. Render an "(installed)" suffix after agents already in `cfg.Agents`. Expose `Result() string` returning the chosen agent machine name.

3. **Ground stage.** Textinput with the same Vessel styling (focus-tinted underline, ColorRule2 placeholder, white-bold input text, stable input cell width via `lipgloss.PlaceHorizontal`). Inject `existingWorkspaces map[string]bool` + `workspaceUniqueValidator` (factored out of `cmd/add.go` — move to `internal/tui/addflow/ground.go` as an unexported helper, or keep in `cmd/add.go` and pass into stage ctor). Tech-lead path: implement `AutoComplete() (string, bool)` that returns the resolved workspace + true when agent type is "tech-lead", so `harness.LazyStep` can skip the stage entirely.

4. **Graft stage.** Full tabbed picker. `NewNewAgentGraft` seeds defaults from `agentDef.DefaultSkills`/etc and shows all 5 tabs. `NewAddItemsGraft` filters each category to uninstalled items (use `availableAddItems` factored out of `cmd/add.go` — move to `internal/tui/addflow/graft.go` or re-export). Empty tabs (all items installed in that category) are dropped from the tab row entirely. Required items: same `*` glyph + always-in-result rule as init Branches. Per-tab counts update live: `"Skills (2)"` → `"Skills (3)"` on toggle.

5. **Observe stage.** Layout: NAME (agent display name), WORKSPACE (resolved path), AGENT (agent type). Below: 5 category blocks showing picks. Below: CTA — `[GRAFT ~N items]` + `[BACK]` buttons. Reuse init's `priorAware.SetPrior(...)` pattern exactly — Observe reads agent + workspace + GraftResult from prev[] and renders without side effects.

6. **Grow stage.** Wrap `initflow.NewGenerateStage` with an addflow-specific action closure that runs `runAddSpinner`'s body. Rename kanji to 育 (GRW). Keep `max(actual, 600ms)` hold. Error panel surfaces `addOutcome.spinnerErr` with the same copy as init.

7. **Yield stage.** Three modes selected at construction:
   - `NewYieldSuccess(installed *config.InstalledAgent, cat *catalog.Catalog, isNewAgent bool, totalSelected int)` — renders a tree of what was added (same ItemTree pattern as init's Planted but rooted at the agent display name + workspace), followed by 3 next-step lines.
   - `NewYieldAllInstalled(agentDef *catalog.AgentDef)` — renders an "already full" body with `bonsai catalog` CTA.
   - `NewYieldTechLeadRequired(agentType string)` — renders an error body with `bonsai init` CTA.

8. **Wire in `cmd/add_redesign.go`.** Step list:
   ```
   [0] Select
   [1] LazyGroup — splices either:
         new-agent: [Ground (chromeless-lazy to skip for tech-lead), NewAgentGraft, Observe]
         add-items: [AddItemsGraft, Observe]                      (no Ground)
         all-installed: [YieldAllInstalled]                        (single stage, terminates here)
         tech-lead-required: [YieldTechLeadRequired]               (single stage, terminates here)
   [2] Conditional(Lazy(Grow))  — gated on observeConfirmed
   [3] LazyGroup — splices conflict picker iff wr.HasConflicts()   (Phase 2)
   [4] Conditional(Lazy(Yield))  — gated on growSucceeded
   ```
   Phase 1 leaves slot [3] as a no-op splice (always empty) — conflict picker lands in Phase 2.

9. **Env flag in `cmd/add.go`.** First line of `runAdd`: `if os.Getenv("BONSAI_ADD_REDESIGN") == "1" { return runAddRedesign(cmd, args) }`.

10. **Tests.** Each stage's `_test.go` covers the same surface shape as the init-package tests. Where a stage is a thin wrapper around an init stage (`Grow`, `Yield`), the test asserts the ctor wiring and the kanji override — not the underlying stage body.

#### Verification

- [ ] `make build && go test ./...` green.
- [ ] `BONSAI_ADD_REDESIGN=1 ./bonsai add` in a fresh scaffold runs the full new-agent flow end-to-end (pick non-tech-lead agent → Ground → Graft → Observe → Grow → Yield). File output identical to legacy `runAdd` for same inputs.
- [ ] `BONSAI_ADD_REDESIGN=1 ./bonsai add` for tech-lead skips Ground.
- [ ] `BONSAI_ADD_REDESIGN=1 ./bonsai add` when all abilities installed terminates at YieldAllInstalled (no Graft / Observe / Grow).
- [ ] `BONSAI_ADD_REDESIGN=1 ./bonsai add` when picking a non-tech-lead agent with no tech-lead installed terminates at YieldTechLeadRequired.
- [ ] Without `BONSAI_ADD_REDESIGN`, `./bonsai add` runs the legacy flow unchanged (no test regressions from the env-flag branch).
- [ ] Esc from Graft returns to Ground (new-agent) / Select (add-items); esc from Observe returns to Graft; esc from Select aborts.
- [ ] `internal/tui/addflow/` package has no hex literals — audit via `grep -nE '#[0-9A-Fa-f]{6}' internal/tui/addflow/` returns zero hits.

### Phase 2 — add-items branch + conflict picker

**Scope:** Complete the add-items branch (filtered Graft) and the Conflicts stage. Still env-flag gated.

#### Files touched

- `internal/tui/addflow/graft.go` — enable the `NewAddItemsGraft` path (stub was in Phase 1 but returned zero tabs — now fully wired to `availableAddItems`).
- `internal/tui/addflow/conflicts.go` — **new.** `ConflictsStage` — one tab per conflict file, each tab shows a diff summary (source vs local) with a 3-way radio (Keep / Overwrite / Backup). Result: `map[string]config.ConflictAction`.
- `internal/tui/addflow/conflicts_test.go` — **new.** Cover: tab per file, radio toggle per tab, default is Keep, result map populated correctly.
- `internal/tui/initflow/design.go` — implement `ConflictRowStyle` body (Phase 1 exposed it as a stub).
- `cmd/add_redesign.go` — wire add-items splicer (Phase 1 had the branch placeholder), wire Conflicts ConditionalStep to replace the empty slot at index [3].

#### Steps

1. **Finish add-items Graft.** When branch = add-items, the Graft ctor uses `NewAddItemsGraft`. Tabs with zero uninstalled items are dropped before the tab row renders.
2. **Conflicts stage.** Build on `BranchesStage`'s tab model. Each tab corresponds to a `generate.FileResult` with conflict action. Keystroke model: `← →` cycle tabs, `↑ ↓ / j k` cycle radio, `↵` advance to next tab or complete if last.
3. **Wire LazyGroup at index [3].** Splicer returns `[ConflictsStage]` iff `wr.HasConflicts()`, else nil.
4. **Apply picks.** After Yield completes, `applyConflictPicks` reads from the Conflicts stage result and mutates `wr` + `lock` (same primitive currently used in `cmd/add.go:256`).
5. **Tests.**

#### Verification

- [ ] Add-items branch: `BONSAI_ADD_REDESIGN=1 ./bonsai add` on an existing agent with uninstalled items in 2 of 5 categories shows only those 2 tabs.
- [ ] Conflicts branch: make a scaffold, hand-edit a generated file, rerun `BONSAI_ADD_REDESIGN=1 ./bonsai add` with more selections — Conflicts stage appears, each tab shows the affected file, radio selection is respected in the final write result.
- [ ] `make build && go test ./...` green.

### Phase 3 — flip default + delete legacy + bundled cleanup

**Scope:** Promote cinematic to default + delete legacy `runAdd`/Phase-1-deferred remnants/env gate + rename `cmd/init_redesign.go` → `cmd/init_flow.go`. **Plus** seven absorbed Backlog items (PR #52/#59/#62 review fallout) that all touch the same files we are already rewriting — bundling avoids reopening the same code twice.

#### Bundled Backlog items (resolved by this PR)

| # | Source | Item | Fix shape |
|---|--------|------|-----------|
| 1 | PR #52 review | Dead post-harness Generate-error warning in `cmd/init_redesign.go:190-198` | Delete the block in the renamed `cmd/init_flow.go`. `GenerateStage.stateError` already prints the in-frame panel; the post-harness `tui.Warning` fires into a cleared terminal. |
| 2 | PR #52 review | No harness composition test for `NewConditional(NewLazy(...))` | Add `internal/tui/harness/steps_test.go` cases asserting (a) `Chromeless()==true` when inner Lazy stage is chromeless and Conditional is not skipped, (b) Lazy builder fires exactly once per active Conditional pass. |
| 3 | PR #59 review | `growSucceeded` predicate walks `prev` tail for any `error` (`cmd/add_redesign.go:84-94`) | Predicate already shares scope with the `outcome addflow.Outcome` struct (`outcome.SpinnerErr` is the only error source on the happy path). Replace the prev-walk with `outcome.SpinnerErr == nil && outcome.Ran` and document the closure capture. No sentinel type needed. Caveat: predicate is evaluated once per harness tick, but `outcome` is only mutated by Grow's action closure which runs to completion before the conditional re-evaluates — safe. |
| 4 | PR #59 review | Unknown-agent path renders `YieldTechLeadRequired` (`cmd/add_redesign.go:104-106`) | Add `addflow.NewYieldUnknownAgent(ctx, agentType)` variant in `internal/tui/addflow/yield.go` with copy `Unknown agent type — catalog may be stale, run \`bonsai update\``. Replace the call. Add a yield_test.go case. |
| 5 | PR #62 security review | `.bak` write-error silent-discard in **both** conflict-apply helpers (`cmd/add_redesign.go:282-288` cinematic + `cmd/root.go:153-161` legacy, used by `remove`/`update`/legacy add) | On `os.ReadFile` failure OR `os.WriteFile` failure for `.bak`, drop that path from `toOverwrite` and surface a single `tui.Warning` listing the dropped paths. Apply the same fix in both helpers. |
| 6 | PR #62 review | `confIdx := len(results) - 2` arithmetic in `cmd/add_redesign.go:232` | Type-scan `results` for `map[string]config.ConflictAction` instead of length math. Mirrors the existing type-scan pattern Observe uses for `GraftResult`. |
| 7 | PR #62 review | No direct unit test for `applyCinematicConflictPicks` | Add `cmd/add_test.go` (new) covering: Keep=noop, Overwrite=ForceSelected only, Backup=.bak+ForceSelected, mixed-action map, empty map=false, backup-read-fail drops + warns, backup-write-fail drops + warns. |

#### Files touched

**Cutover (the original Phase 3 scope):**

- `cmd/add.go` — **replace.** Body becomes `runAddRedesign` (renamed back to `runAdd`). Delete:
  - `runAdd` legacy body (lines 100-279, includes the `BONSAI_ADD_REDESIGN` env-flag branch lines 104-108)
  - `addOutcome` struct (lines 281-289) — only legacy uses it; cinematic uses `addflow.Outcome`
  - `runAddSpinner` (lines 292-399)
  - `buildAddItemsSteps` (lines 616-708)
  - `buildNewAgentSteps` (lines 469-558)
  - `normaliseWorkspace` (lines 94-101) — only legacy `buildNewAgentSteps` uses it; cinematic uses `addflow.NormaliseWorkspace`
  - `workspaceUniqueValidator` (lines 76-92) — only legacy `buildNewAgentSteps` uses it; cinematic Ground stage has its own validator
  - `newDescriber` (lines 58-74) — only legacy `buildNewAgentSteps` consumes its result
  - `userSensorOptions` (in `cmd/init.go:22-31`) — only `buildNewAgentSteps` calls it; verify with `grep` after deletion
- **Keep** in `cmd/add.go`: `availableAddItems` + `distributeAddItemPicks` (still consumed by `buildAddGrowAction` in the cinematic path).
- **Keep** in `cmd/init.go`: `asString` / `asStringSlice` / `asBool` (still used by `cmd/root.go` `applyConflictPicks` + `cmd/remove.go` + `cmd/update.go`).
- `cmd/add_redesign.go` — **delete** (contents moved to `cmd/add.go`). Apply cleanup fixes #3, #5, #6 in the moved body before saving.
- `cmd/init_redesign.go` — **rename** `git mv` → `cmd/init_flow.go`. Apply cleanup fix #1 in the renamed file.

**Bundled-cleanup files (new or edited):**

- `internal/tui/addflow/yield.go` — add `NewYieldUnknownAgent` variant + `yieldModeUnknownAgent` enum value + `renderUnknownAgent()` body. Delete `NewYieldAddItemsDeferred` + `yieldModeAddItemsDeferred` + `renderAddItemsDeferred()` (all unreachable post-Phase 2).
- `internal/tui/addflow/yield_test.go` — drop the `NewYieldAddItemsDeferred` test case (line 137); add `NewYieldUnknownAgent` test case.
- `internal/tui/addflow/grow.go` — unchanged. (Predicate hardening uses closure capture, not a sentinel.)
- `internal/tui/harness/steps_test.go` — add 2 composition tests (item #2).
- `cmd/add_test.go` — **new.** `applyCinematicConflictPicks` table tests (item #7) including backup-failure paths.
- `cmd/root.go` — modify `applyConflictPicks` per item #5 (drop-path-on-backup-fail + warn).

#### Steps

1. **Audit dead-helper grep** before any deletion — re-run `grep -rn "buildNewAgentSteps\|buildAddItemsSteps\|runAddSpinner\|addOutcome\|workspaceUniqueValidator\|normaliseWorkspace\|newDescriber\|userSensorOptions\|NewYieldAddItemsDeferred\|yieldModeAddItemsDeferred"` against `cmd/` + `internal/` to confirm what is safely removable. If any survives outside the legacy `runAdd` body, leave it.
2. **Add `NewYieldUnknownAgent` variant** in `internal/tui/addflow/yield.go`. Update `yield_test.go`.
3. **Move `runAddRedesign` body into `cmd/add.go`**, replacing legacy `runAdd`. Apply inline:
   - Cleanup #3: `growSucceeded` predicate reads `outcome.SpinnerErr` via closure capture instead of walking `prev` for any error.
   - Cleanup #4: unknown-agent branch calls `addflow.NewYieldUnknownAgent`.
   - Cleanup #5: cinematic `applyCinematicConflictPicks` drops paths on backup-read OR backup-write failure + collected `tui.Warning`.
   - Cleanup #6: post-harness conflict slot resolved by type-scan, not `len(results)-2`.
4. **Delete `cmd/add_redesign.go`** (moved). Delete `NewYieldAddItemsDeferred` + `yieldModeAddItemsDeferred` + `renderAddItemsDeferred()` from `internal/tui/addflow/yield.go` + their test case.
5. **Delete legacy helpers in `cmd/add.go`** per file-touched table. Verify with `go build ./...` + `golangci-lint run` (catch anything missed by the unused linter).
6. **`git mv cmd/init_redesign.go cmd/init_flow.go`.** Apply cleanup #1 (delete dead post-harness Generate-error warning lines 190-198).
7. **Modify `cmd/root.go` `applyConflictPicks`** for cleanup #5 (drop-path-on-fail + warn).
8. **Add `cmd/add_test.go`** with `applyCinematicConflictPicks` table tests including backup-failure paths.
9. **Add harness composition tests** in `internal/tui/harness/steps_test.go` (item #2).
10. **Verify env-flag dead** — `grep -rn "BONSAI_ADD_REDESIGN"` returns zero hits.
11. **Build + test + smoke.**

#### Verification

- [ ] `make build && go test ./...` green.
- [ ] `gofmt -s -l .` returns empty.
- [ ] `golangci-lint run` (project-CI version) clean — no `unused` warnings.
- [ ] `grep -rn "BONSAI_ADD_REDESIGN" .` → zero hits.
- [ ] `grep -rn "buildNewAgentSteps\|buildAddItemsSteps\|runAddSpinner\|addOutcome\|NewYieldAddItemsDeferred\|yieldModeAddItemsDeferred"` → zero hits in non-archived files.
- [ ] `cmd/init_redesign.go` removed; `cmd/init_flow.go` present.
- [ ] `./bonsai add` smoke (no env var) — all five paths still correct:
  - new-agent (non-tech-lead) → Ground → Graft → Observe → Grow → Yield
  - tech-lead → Ground auto-completed → Graft → Observe → Grow → Yield
  - add-items → Graft (filtered) → Observe → Grow → Yield (or YieldAllInstalled)
  - tech-lead-required guard → YieldTechLeadRequired
  - unknown-agent path → **YieldUnknownAgent** (new copy)
  - conflicts present → ConflictsStage tabs → backup writes succeed and `.bonsai-lock.yaml` reflects ForceSelected.
- [ ] **Backup-fail dogfood:** make a scaffold, hand-edit a file to trigger conflict, set its parent dir read-only, rerun `bonsai add` with Backup pick — file is dropped from overwrite + warning surfaces; original file untouched.
- [ ] `applyCinematicConflictPicks` test covers all 7 cases listed in item #7.
- [ ] Harness composition test asserts both cases listed in item #2.
- [ ] `addflow.GrowResult` sentinel test passes for both nil and non-nil error.
- [ ] `NewYieldUnknownAgent` test asserts copy mentions `bonsai update`.

---

## Dependencies

- `internal/tui/initflow/` package (built in Plan 22, already imported across the codebase).
- `internal/tui/styles.go` tokens (verified current as of commit `027ddd0`).
- `internal/tui/harness/` — `LazyStep`, `ConditionalStep`, `LazyGroup`, `Chromeless` delegation (all exist post-Plan-22 Phase 5B).
- `internal/generate.WriteResult` / `config.LockFile` — unchanged.
- `go.mod` — no new dependencies.

---

## Security

> [!warning]
> Refer to `Playbook/Standards/SecurityStandards.md` for all security requirements.

- No new external dependencies; no new network calls; no file I/O outside existing `generate.*` primitives.
- `ConflictsStage` must not mutate filesystem state — only collect user picks into `map[string]config.ConflictAction`. The mutation happens post-harness in `applyConflictPicks` exactly as today.
- Env-flag branch `BONSAI_ADD_REDESIGN=1` is a developer affordance only — no user data reaches the env; delete it in Phase 3.

---

## Verification (overall)

- [ ] All three phases' individual Verification blocks green.
- [ ] Post-Phase-3: legacy add harness is fully removed; cinematic flow is default; `cmd/init_redesign.go` renamed.
- [ ] Dogfood smoke on the Bonsai repo itself (`bonsai add` a second tech-lead route or new skill) — matches or improves on the Plan 22 init experience.
- [ ] Independent code review pass on each PR per `agent/Workflows/code-review.md` + `agent/Skills/review-checklist.md`.
