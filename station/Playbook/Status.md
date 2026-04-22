---
tags: [playbook, status]
description: Live task tracker. Update this file at the start and end of every working session.
---

# Bonsai — Status

> [!note]
> Move items between tables as work progresses. Link to the relevant plan in `Plans/Active/` or `Plans/Archive/`.

---

## In Progress

| Task | Plan | Agent | Notes |
|------|------|-------|-------|
<!-- Plan 22 (`bonsai init` cinematic redesign) — 6 PRs shipped 2026-04-21; plan archived to Plans/Archive/22-init-redesign.md -->
<!-- Plan 19 (pre-launch bug sweep) — merged PR #27 squash a44e447 2026-04-21, moved to Recently Done -->
<!-- Documentation site (Starlight) — Plan 10 complete, all phases shipped (PRs #13-19) -->
<!-- UI/UX overhaul Phase 2 — merged PR #20, moved to Recently Done 2026-04-17 -->
<!-- ActionUnchanged follow-ups — merged PR #22, moved to Recently Done 2026-04-17 -->
<!-- Plan 15 (BubbleTea foundation) — merged PR #26 squash 2ce63f6 2026-04-20, moved to Recently Done -->
<!-- Plan 18 (`bonsai guide` multi-topic) — merged PR #25 2026-04-20, moved to Recently Done -->
<!-- Plan 16 (`go install` binary name) — merged PR #23 2026-04-20, moved to Recently Done -->

<!-- UI/UX Overhaul Phase 3 (Plan 14) — merged via PR #24, moved to Recently Done 2026-04-17 -->
<!-- Release prep (Plan 17) — merged PR #24, moved to Recently Done 2026-04-17 -->
<!-- Code index refresh — completed 2026-04-16, PR #12 -->
<!-- Cross-link moved to Recently Done — merged 2026-04-16, PR #4 -->
<!-- P0 file collision fix — merged 2026-04-16, PR #8 -->
<!-- Docs & repo polish — merged 2026-04-16, PR #9 -->

## Pending

| Task | Plan | Agent | Blocked By |
|------|------|-------|------------|
<!-- Better trigger sections promoted to In Progress — 2026-04-16, planning session -->
<!-- UI overhaul (P2) — promoted to In Progress as Plan 12, removed 2026-04-17 -->
<!-- Plan 08 Phase C shipped via Plan 21 / PR #46 (2026-04-21): compact-recovery sensor + context-guard verify/plan patterns. C3 deferred permanently per original plan note. -->

## Recently Done

| Task | Plan | Agent | Date |
|------|------|-------|------|
| Pre-launch security sweep — (commits `ef55f4f` + hotfix `7c1dd49`, direct-to-main): audited Dependabot + CodeQL + secret-scanning. Dependabot alert #1 (Astro XSS via `define:vars`, CVE-2026-41067 < 6.1.6) auto-fixed by lockfile resolving `astro@6.1.7`. Two low-severity CodeQL `go/useless-assignment-to-local` alerts silenced: `internal/tui/styles.go:211` (`val := value` → `var val string`; both branches overwrite) and `cmd/add.go:690` (drop trailing `offset++` after last conditional read). Secret scanning clean. Widened `ci.yml` to trigger on `push: branches: [main]` in addition to `pull_request` — closes the "gofmt drift on main silently hides because CI is PR-only" pattern (doc'd in memory). Branch-protection ruleset `main-protection` kept active with admin bypass (fast-iter UX convention preserved; external contributors still hit required `test` check). First push-CI run immediately caught a follow-on breakage in `ef55f4f`: `git add cmd/add.go` had accidentally staged Plan 23 WIP hunks (os import + `BONSAI_ADD_REDESIGN` gate calling undefined `runAddRedesign`) along with the intended offset++ drop — hotfix `7c1dd49` removed the WIP hunks; Plan 23 WIP restored to working tree uncommitted for parallel-session owner. Net lesson: push-CI paid for itself within 5 min of shipping. | — | tech-lead | 2026-04-22 |
| Pre-launch polish bundle — **Plan 24** (PR #58 squash `4ef8271`): (A) `CHANGELOG.md` at repo root in keep-a-changelog 1.1.0 format with curated v0.1.0–v0.1.3 backfill + `[Unreleased]` stub + 5 link-reference rows; (B) `.github/workflows/docs.yml` gains `pull_request` trigger (same `paths:` as `push`) + `if: github.event_name == 'push'` guards on both `deploy` job and `upload-pages-artifact` step — broken MDX now fails at PR time, deploy stays push-to-main (prevents the PR #25 post-merge MDX incident pattern); (C) root `Bonsai/CLAUDE.md` `internal/tui/` block refreshed to include `styles_test.go`, `filetree.go`, `filetree_test.go`, `harness/` (Plan 15), `initflow/` (Plan 22) — subdirs listed only, no per-file enumeration for 22-file initflow; (D) Backlog Group C/D consolidation — two duplicate CHANGELOG entries replaced with HTML-comment markers, Group D changelog-generation feature retained as future work with "(refiled as good-first-issue)" suffix. Also (E) 5 `good first issue` + `help wanted` issues filed: [#53](https://github.com/LastStep/Bonsai/issues/53) statusLine port (relabeled), [#54](https://github.com/LastStep/Bonsai/issues/54) shell completion, [#55](https://github.com/LastStep/Bonsai/issues/55) `bonsai changelog` + skill, [#56](https://github.com/LastStep/Bonsai/issues/56) catalog umbrella, [#57](https://github.com/LastStep/Bonsai/issues/57) `.bak` merge hint. +72/−5 across 4 files. CI: pre-existing gofmt drift in `observe.go:466` + `planted.go:423` (left by earlier direct-to-main polish commits) caught on PR lint — fixed on same branch via `gofmt -s -w` before merge; exposed "PR CI ≠ main CI" gotcha again. 6/6 CI green post-fix. | 24 | general-purpose | 2026-04-22 |
| statusLine redesign (project scope) — new `station/agent/Sensors/statusline.sh` persistent bar: `cave:<mode> · <workspace> · <model> · <branch>[*<dirty>] · ctx N% · 5h N% · 7d N% · <elapsed> · $<cost>` with sage/sand/rose 256-color tiers + `NO_COLOR` + `BONSAI_STATUSLINE_HIDE` toggles. Caveman badge wrapped natively (reads `$CLAUDE_CONFIG_DIR/.caveman-active` with symlink refuse + 64B cap + whitelist). Wired via top-level `statusLine` stanza in project `.claude/settings.json` (walk-up-to-`.bonsai.yaml` bash wrapper), `padding: 0`, `refreshInterval: 30`. Rewrote `status-bar.sh` → warnings-only (silent on zero; uncommitted count, memory-stale via `git log`, overdue-routines via dashboard parse). Rewrote `context-guard.sh` → self-sufficient (computes ctx% from `transcript_path` directly, no `/tmp` state file dep) — all 30/50/70/85 tier injections + wrap-up/verify/plan triggers preserved. Direct-to-main per fast-iter UX convention. Catalog port filed as Backlog P2 Group E (Phase 2). | — | tech-lead | 2026-04-22 |
| Plan 22 **complete** — Phase 5B: wire Generate+Planted into `runInit`, flip default, delete legacy init. `runInitRedesign` → `runInit`, `BONSAI_REDESIGN` env-flag removed, 245 lines of legacy `cmd/init.go` deleted (`buildReviewPanel`, old `runInit` body). Step order: Vessel → Soil → Branches → Observe → Conditional[Lazy[Generate]] → LazyGroup[conflicts] → Conditional[Lazy[Planted]]. Harness extension: `LazyStep.Chromeless()` + `ConditionalStep.Chromeless()` delegation + `ConditionalStep.Init` auto-builds nested Lazy (necessary for `NewConditional(NewLazy(GenerateStage))` given `GenerateAction = func() error` needs prev-capture). `buildGenerateAction` closes over cat/agentDef/cwd/configPath/lock/wr/cfg/installed; `plantedSummary(installed)` built lazily so post-`EnsureRoutineCheckSensor` counts render correctly. +260/−297 across 8 files. PR #52 squash `5916e05`. 6/6 CI green; independent review PASS + 2 minors deferred (composition test + dead-path warning — Backlog). | 22 | general-purpose | 2026-04-21 |
| Plan 22 Phase 5A — responsive resize foundation + Observe stage wired + Generate/Planted packaged. `internal/tui/initflow/layout.go` new: `MinTerminalWidth=70`, `MinTerminalHeight=20`, `TerminalTooSmall`, `ClampColumns(120) → (24, 44, 12)` regression anchor, hand-rolled `Viewport` (no bubbles/viewport dep — matches Soil precedent). `RenderMinSizeFloor(w,h)` in chrome.go routes from `stage.renderFrame` before body composition. Branches/Vessel/Soil retrofit to consume `ClampColumns` + `Viewport` scroll. `ObserveStage` replaces idx=3 stub — 4-block review (vessel / soil / branches / CTA), `[PLANT]`/`[BACK]` buttons, y/n/Enter/tab wiring, 2-col at ≥100 cols; `priorAware.SetPrior` captures vessel map + soil []string + BranchesResult. `GenerateStage` + `PlantedStage` packaged (not wired — 5B splices them). 50+ new tests. PR #51 squash `6baaf8e` + lint-fix `3e31967`. Process: dispatched via isolated worktree agent, independent review PASS. | 22 | general-purpose | 2026-04-21 |
| Plan 22 Phase 4 polish run — 5 direct-to-main dogfood commits against Branches stage (no PRs per fast-iter UX memory): `413e360` header split (`[ 盆 ]` + INITIALIZE stacked), per-tab 2-line intro above list, DETAILS moved to fixed-height box below list (no `?`-toggle jitter), right-aligned DEFAULT/(required) tags; `399fe08` density cut + word-wrapped details + `lipgloss.PlaceHorizontal(Center)` on tab counts; `eaee416` wider layout (row 84 cells, nameColW 24 + descColW 44 + tab colW 16 + divider trails 60) + rune-aware name truncation keeps DEFAULT tag in its column even for future long names; `6bb74e5` ABOUT + FILE values → `ColorAccent` (white) + 3-row wrap × 70 cells (absorbs dispatch-guard-length ~111-char descriptions that previously clipped at 2×60); `fa0ae64` kanji padding `[ 盆 ]` (terminals left-anchor CJK in their 2-cell slot) + extra blank line between DETAILS and counter. New helper `wrapToWidth` (word-break w/ rune-fallback hard-wrap). All 12 tests green across the pass. | 22 | tech-lead | 2026-04-21 |
| Plan 22 Phase 4 — `BranchesStage` tabbed picker across Skills/Workflows/Protocols/Sensors/Routines with inline-expand (`?`), per-category multi-select, defaults pre-seeded from agentDef, required items pinned. Hand-rolled 5-tab UI, `BranchesResult` shape, Reset() preserves state across esc-back. 12 tests cover tab cycling, focus clamp, toggles, expand, Result, defaults, Reset preservation (PR #50 squash `89c21ba`) | 22 | general-purpose | 2026-04-21 |
| Plan 22 Phase 3.5 — dogfood polish pass (direct-to-main, no PR per fast-iter memory): Bark→gold `#D4AF37`, new Moon white token + ColorAccent=Moon, drop kana row + secondary kana appends, current-stage anchor gold, done-anchor bright Primary, rail capped 60 cells + centred, footer muted rule, 2-col field layout (LABEL/subtitle + input/underline), focus-tinted underline, placeholder dim Rule2, white-bold input text, helper copy dimmed Rule2, stable input cell width via `lipgloss.PlaceHorizontal` (fixes typed-shift from textinput View width asymmetry), copy refresh (drop verbose hints, `Tend the soil.` headline, `Three quick answers — a name, a purpose, a place to grow.`) | 22 | tech-lead | 2026-04-21 |
| Plan 22 Phase 3 — `VesselStage` (3 textinputs) + `SoilStage` (hand-rolled multi-select) + `RenderHeader` signature strip `stationSubdir` (bug fix: header no longer shows non-existent `station/` subdir) (PR #49 squash `971ee44`) | 22 | general-purpose | 2026-04-21 |
| Plan 22 Phase 2 — `internal/tui/initflow/` package (chrome/enso/fallback/stage/stub) + `harness.Chromeless` + env-flag `runInitRedesign` (PR #48 squash `2e2a08c`) | 22 | general-purpose | 2026-04-21 |
| Plan 22 Phase 1 — `RenderFileTree` widget + `ColorLeafDim`/`ColorRule`/`ColorRule2` palette tokens (PR #47 squash `7553d43`) | 22 | general-purpose | 2026-04-21 |
| Session-start context dedup + Phase C sensors — session-context redundancy cuts, context-guard verify/plan patterns, new compact-recovery sensor, UX prefs moved to memory.md Feedback (PR #46 squash `d14edbe`) | 21 | general-purpose | 2026-04-21 |
| Better trigger sections — Phase C (new sensors: compact-recovery + context-guard expansion) **shipped via Plan 21 / PR #46**; C3 deferred | 08 | general-purpose | 2026-04-21 |
| Security scanning infra — Go 1.25.8 bump, golangci-lint v2 + pin, Dependabot, govulncheck CI, CodeQL workflow, gitleaks history audit clean (PRs #28 #29 #30 #31 #40 #41) | 20 | general-purpose | 2026-04-21 |
| Pre-launch bug sweep — 8 OSS-blocker fixes: CRLF, cross-workspace tree, dedup, spinner errors.Join, harness polish (PR #27 squash `a44e447`) | 19 | general-purpose | 2026-04-21 |
| BubbleTea foundation + theme system — harness migration across init/add/remove/update (PR #26 squash `2ce63f6`) | 15 | general-purpose | 2026-04-20 |
| `bonsai guide` multi-topic — 3 cheatsheets + delete 3 orphan docs (1,213L) + CLI refactor (PR #25) | 18 | general-purpose | 2026-04-20 |
| Fix `go install` binary name — main.go → cmd/bonsai/ + root embed package (PR #23) | 16 | general-purpose | 2026-04-20 |
| Release prep — Go toolchain 1.24.13 + triggerSection frontmatter fix + OSS polish (PR #24) | 17 | general-purpose | 2026-04-17 |
| UI/UX overhaul — Phase 3 visual identity + init polish (merged via PR #24 bundle) | 14 | general-purpose | 2026-04-17 |
| ActionUnchanged follow-ups — chmod bug + test gaps (PR #22) | 13 | tech-lead | 2026-04-17 |
| UI/UX overhaul — Phase 2 (consistency: hints, counts, no-op detection, structured errors) | 12 | tech-lead | 2026-04-17 |
| Documentation site (Starlight) — Phase D (deploy & CI) — **Plan 10 complete** | 10 | tech-lead | 2026-04-17 |
| Documentation site (Starlight) — Phase C (catalog auto-generation + LLM layer) | 10 | tech-lead | 2026-04-17 |
| Documentation site (Starlight) — Phase B (fill content gaps, 30 pages) | 10 | tech-lead | 2026-04-17 |
| UI/UX overhaul — Phase 1 (adaptive palette, NO_COLOR, FatalPanel, version banner) | 11 | tech-lead | 2026-04-17 |
| Code index refresh — fix line drift + missing entries | 09 | tech-lead | 2026-04-16 |
| Better trigger sections — Phase B (trigger documentation) | 08 | tech-lead | 2026-04-16 |
| Better trigger sections — Phase A (trigger metadata system) | 08 | tech-lead | 2026-04-16 |
| Community health files + README polish for public release | 07 | tech-lead | 2026-04-16 |
| Fix case-insensitive file collision (index.md / INDEX.md) | 06 | tech-lead | 2026-04-16 |
| README rewrite for open-source release | — | tech-lead | 2026-04-16 |
| Cross-link generated files — Obsidian-compatible markdown links | 03 | tech-lead | 2026-04-16 |
| AI operational intelligence — How to Work + workspace-guide skill | 05 | tech-lead | 2026-04-16 |
| Release pipeline — GoReleaser + GitHub Actions + Homebrew Tap | 04 | tech-lead | 2026-04-15 |
| Rename "catalog items" to "abilities" — CLI, TUI, README, docs, comments | — | tech-lead | 2026-04-15 |
| `bonsai guide` command — render custom files guide in terminal | 02 | tech-lead | 2026-04-15 |
| CLAUDE.md marker migration — backup + overwrite for marker-less files | 01 | tech-lead | 2026-04-15 |
| Selective file update — multi-select conflict picker | — | tech-lead | 2026-04-15 |
| Fix doubled path prefix in `bonsai add` output panels | — | tech-lead | 2026-04-15 |
| `bonsai update` — custom file detection + workspace sync | — | tech-lead | 2026-04-14 |
| Dogfooding — station workspace setup + customization | — | tech-lead | 2026-04-14 |
| Session wrap-up workflow + context-guard wiring | — | tech-lead | 2026-04-14 |
| Stale artifact cleanup (Python refs, index.md rewrite) | — | tech-lead | 2026-04-14 |
| Awareness Framework — status-bar + context-guard sensors | — | tech-lead | 2026-04-13 |
| Lock file conflict handling | — | tech-lead | 2026-04-13 |
| Catalog expansion — all 3 phases, 6 agent types | — | tech-lead | 2026-04-13 |
| Go rewrite from Python | — | tech-lead | 2026-04-12 |
