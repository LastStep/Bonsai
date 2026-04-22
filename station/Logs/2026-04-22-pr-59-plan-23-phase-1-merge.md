---
tags: [log, session, plan-23, phase-1, pr-review, cross-track]
date: 2026-04-22
plan: 23
phase: 1
pr: 59
squash: 163bc45
---

# PR #59 review + fixes + merge — Plan 23 Phase 1 (`bonsai add` cinematic new-agent path)

## What shipped

PR #59 `163bc45` — squash-merged to main. Final diff net +~2900 / −~350 across ~30 files (addflow package + initflow exports + cmd wiring + 46 tests).

### New

- `internal/tui/addflow/` — 7 stage files mirroring initflow shape:
  - `addflow.go` — `Outcome` scratchpad + shared helpers.
  - `select.go` — 選 agent picker with BuildAgentOptions + installedSet filter.
  - `ground.go` — 地 workspace input with existing-workspace collision guard.
  - `graft.go` — 接 abilities multi-tab picker; `GraftResult{Skills,Workflows,Protocols,Sensors,Routines []string}` payload.
  - `observe.go` — 観 review card with y/n/Enter/tab wiring + 2-col at ≥100 cols.
  - `grow.go` — 育 spinner stage wrapping a `GenerateAction` closure.
  - `yield.go` — 結 terminal card with 4 variants (`yieldModeSuccess` / `yieldModeAllInstalled` / `yieldModeTechLeadRequired` / `yieldModeAddItemsDeferred`).
- 46 unit tests across the 7 stages.
- `cmd/add_redesign.go` (311 lines) — `runAddRedesign` harness wiring with Select → LazyGroup(agent flow splicer) → Conditional[Lazy[Grow]] → LazyGroup[conflicts] → Conditional[Lazy[Yield]].
- 7-line env-flag branch at top of `cmd/add.go:runAdd`: `if os.Getenv("BONSAI_ADD_REDESIGN") == "1" { return runAddRedesign(cmd, args) }`.

### Modified (initflow additive exports — no behaviour change to init)

- `CenterBlock` / `PadRight` package-level aliases exposed for addflow consumers.
- `Stage.SetRailLabels` / `SetRailIndex` / `SetLabel` / `SetSize` / `RenderFrame` + read-accessors.
- `GenerateStage.SetBodyTitle` kanji override.
- `RenderEnsoRail(labels []string)` signature — nil labels fall back to the init 6-stage default.

## Process — not a clean agent-dispatch, user took over

Plan 23 Phase 1 was agent-dispatched from a parallel session earlier the same day. I inherited the branch at review time. User authorization: *"git should be clean now. merge the pr to main, and do the fixes."*

### Blockers surfaced in review

| ID | Category | Issue |
|----|----------|-------|
| B1 | correctness | Splicer routed BOTH `installed with items available` AND `installed already full` to `YieldAllInstalled` — "already full" copy lied to the user when Phase 2 was simply not built. |
| B2 | lint | `gofmt -s` drift in `graft.go:117` + `observe.go:30` — CI lint fail. |
| B3 | process | Branch forked from `7c1dd49` before `f24750d` + `674f987` landed on main. PR diff reintroduced the "Plant ~N files into" misleading CTA (removed deliberately on `674f987`) and reverted 4 station/ bookkeeping files. |

### Fixes applied on branch

1. **Rebase on origin/main** — clean, auto-dropped 6 cross-track drift files (B3 resolved with zero conflicts because main's versions of the drifted files already held the correct content).
2. **`babb3df`** — `gofmt -s -w internal/tui/addflow/{graft,observe}.go` (B2 resolved).
3. **`885d128`** — B1 resolved:
   - `internal/tui/addflow/yield.go` — added `yieldModeAddItemsDeferred` enum variant + `NewYieldAddItemsDeferred(ctx, agentDef)` ctor + `renderAddItemsDeferred()` body with amber "ADD-ITEMS COMING IN PHASE 2" hero and two-line CTA: `$ unset BONSAI_ADD_REDESIGN` then `$ bonsai add`. `renderBody()` switch extended with the new case.
   - `cmd/add_redesign.go` splicer — partial-installed arm switched from `NewYieldAllInstalled` to `NewYieldAddItemsDeferred`. Only the `installedAgent + availableAddItems.Total() == 0` case still routes to `YieldAllInstalled`.
   - `internal/tui/addflow/yield_test.go` — new test `TestYield_AddItemsDeferredRendersLegacyCTA` asserts hero + unset-env CTA + bonsai-add CTA + agent display name.

### Push + merge

- Force-pushed `--force-with-lease` (safe — single-author branch, no upstream collaborators).
- CI: 6/6 green in 2m window (test / Analyze Go / lint / govulncheck / CodeQL / GitGuardian).
- Squash-merged via `gh pr merge --squash --delete-branch`; squash SHA `163bc45`.
- Local + remote branch cleanup — clean.

## Review misread worth recording

C3 in the initial review was filed as "YieldStage lacks `Result()`" — turned out to already exist at `yield.go:318` (`func (s *YieldStage) Result() any { return nil }`). First-scan reading missed it. Lesson: before filing a "missing method" finding, grep the file for the method name explicitly.

## Deferred — filed to Backlog

- **C1 (P2)** — `growSucceeded` predicate walks `prev` tail for any error value; fragile if an earlier stage's Result ever contains an error type. Tighten to explicitly target the Grow slot index, or have Grow publish a sentinel struct.
- **C2 (P2)** — unknown-agent path in splicer renders `YieldTechLeadRequired` — copy lies to the user (the card is about a missing tech-lead, not an unknown agent). Add a `NewYieldUnknownAgent` variant or collapse the catch to a fatal panel.

Both filed under Backlog Group B with `(added 2026-04-22, source: PR #59 review)`.

## Remaining Plan 23 work

- **Phase 2** — add-items branch (filtered Graft + Observe + Grow for existing agents) + `ConflictsStage` real body (currently LazyGroup placeholder no-ops).
- **Phase 3** — flip default (remove `BONSAI_ADD_REDESIGN` gate) + delete legacy `runAdd` body.

No active parallel session owning these — open for a future dispatch.

## Cross-reference

- Status.md Recently Done top row updated with PR #59 line.
- memory.md Work State refreshed: `Main at: 163bc45`, `In flight` notes Plan 23 Phase 2 + Phase 3 pending with no active parallel session.
- Plan file: `Playbook/Plans/Active/23-uiux-phase2-add.md` (stays Active — Phase 2 + 3 pending).
