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

### Phase 3 — flip default + delete legacy

**Scope:** Promote the cinematic flow to default. Delete the legacy `runAdd` body (everything replaced by `runAddRedesign`), the env flag branch, and the `buildNewAgentSteps` / `buildAddItemsSteps` helpers that the cinematic flow doesn't use. Rename `cmd/add_redesign.go` → `cmd/add.go` (replacing the old file contents).

#### Files touched

- `cmd/add.go` — **replace.** Move `runAddRedesign` body here as the new `runAdd`. Delete every helper that is only called by the legacy path (`buildNewAgentSteps`, `buildAddItemsSteps`, `distributeAddItemPicks` if unused by cinematic path, the `BONSAI_ADD_REDESIGN` env-flag branch).
- `cmd/add_redesign.go` — **delete** (contents moved to `cmd/add.go`).
- `cmd/init_redesign.go` — **rename** to `cmd/init_flow.go`. This is a housekeeping rename (file name no longer reflects reality post-Plan-22). Bundle here rather than a separate commit.
- `cmd/add_test.go` (if any) — update references.
- Any `cmd/add.go` helper that was used by both paths (e.g., `userSensorOptions`, `asString`, `asStringSlice`, `newDescriber`, `workspaceUniqueValidator`, `normaliseWorkspace`) stays — likely consumed by `remove`/`update`/the cinematic path.

#### Steps

1. **Move `runAddRedesign` body into `cmd/add.go`** as the new `runAdd`.
2. **Delete legacy helpers** (`buildNewAgentSteps`, `buildAddItemsSteps`, anything else that becomes dead code). Run `go build ./...` + `golangci-lint run` to catch anything missed.
3. **Rename `cmd/init_redesign.go` → `cmd/init_flow.go`.** Pure `git mv`. Update imports / build tags if any.
4. **Remove env-flag branch.**
5. **Sanity smoke** — `./bonsai add` (no env var) runs the cinematic flow; no regression vs Phase 2 behavior.

#### Verification

- [ ] `make build && go test ./...` green.
- [ ] `grep -r "BONSAI_ADD_REDESIGN" .` returns zero hits.
- [ ] `grep -r "buildNewAgentSteps\|buildAddItemsSteps" .` returns zero hits in non-archived files.
- [ ] `./bonsai add` smoke: new-agent path, add-items path, all-installed path, tech-lead-required path, conflicts path — each still produces correct file output.
- [ ] `cmd/init_redesign.go` no longer exists; `cmd/init_flow.go` does.

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
