---
tags: [playbook, status-archive]
description: Archive of Done items moved from Status.md once they age past 14 days. Appended to by the status-hygiene routine.
---

# Bonsai — Status Archive

> [!note]
> Done items older than 14 days are moved here from `Status.md` by the status-hygiene routine (runs every 5 days). Newest entries at the top.

<!-- status-hygiene routine appends archived rows below this marker. -->

## Archived

| Task | Plan | Agent | Date |
|------|------|-------|------|
| Plan 31 — v0.3 release readiness: peer-awareness refresh + bonsai-model skill + catalog.json snapshot, removeflow + updateflow cinematics, hints package. 3 PRs. [plan](Plans/Archive/31-v03-release-readiness.md) · [#75](https://github.com/LastStep/Bonsai/pull/75) · [#76](https://github.com/LastStep/Bonsai/pull/76) · [#77](https://github.com/LastStep/Bonsai/pull/77) | 31 | tl + gp×5 | 2026-04-24 |
| Plan 30 — guide viewer perf + view-cmds polish: glamour renderer cache + pre-warm, 18 NIT cleanups. [plan](Plans/Archive/30-guide-perf-and-view-polish.md) · [PR #74](https://github.com/LastStep/Bonsai/pull/74) | 30 | gp×3 + tl | 2026-04-23 |
| Archive-reconcile sweep: 20 shipped plans Active→Archive, 14 stale frontmatters synced. Commit `45df4cd`. | — | tl | 2026-04-23 |
| Plan 29 — init+add bug bundle: 8 phases, 8 bugs + 4 test gaps, workspace path-escape validators. [plan](Plans/Archive/29-init-add-bug-bundle.md) · [PR #72](https://github.com/LastStep/Bonsai/pull/72) | 29 | gp + tl | 2026-04-23 |
| Plan 28 — `bonsai list` + `bonsai guide` cinematics shipped in parallel. [plan](Plans/Archive/28-view-cmds-cinematic.md) · [#71](https://github.com/LastStep/Bonsai/pull/71) · [#70](https://github.com/LastStep/Bonsai/pull/70) | 28 | gp×4 + tl | 2026-04-23 |
| Plan 27 PR2 — addflow Phase C polish + Phase D verification: chromeless conflicts, vertical list, per-file action color. [plan](Plans/Archive/27-add-flow-polish.md) · [PR #69](https://github.com/LastStep/Bonsai/pull/69) | 27 | gp + tl | 2026-04-22 |
| Plan 28 Phase 1 — cinematic `bonsai catalog` + RenderHeader extension + hide `completion`. [plan](Plans/Archive/28-view-cmds-cinematic.md) · [PR #68](https://github.com/LastStep/Bonsai/pull/68) | 28 | gp×2 + tl | 2026-04-22 |
| Plan 27 PR1 — addflow Phase A foundations + Phase B bug fixes: rail shrink, graft→branches rename, harness re-splice, *ForAgent variants. [plan](Plans/Archive/27-add-flow-polish.md) · [PR #67](https://github.com/LastStep/Bonsai/pull/67) | 27 | gp×2 + tl | 2026-04-22 |
| Plan 26 — P2 knock-off bundle: 4 cleanups + routine-dashboard regression test. [plan](Plans/Archive/26-p2-knockoff-bundle.md) · [PR #66](https://github.com/LastStep/Bonsai/pull/66) | 26 | gp + tl | 2026-04-22 |
| v0.2.0 release: pre-release docs audit bundle + GoReleaser binaries + Homebrew formula (manual push after PAT rotation). [PR #65](https://github.com/LastStep/Bonsai/pull/65) | — | tl + gp | 2026-04-22 |
| Plan 23 Phase 3 — addflow cutover + 7 bundled cleanups: env gate deleted, legacy paths gone, init_redesign→init_flow rename. [plan](Plans/Archive/23-uiux-phase2-add.md) · [PR #64](https://github.com/LastStep/Bonsai/pull/64) | 23 | gp×2 + tl | 2026-04-22 |
| Plan 23 Phase 2 — `bonsai add` cinematic add-items branch + ConflictsStage. [plan](Plans/Archive/23-uiux-phase2-add.md) · [PR #62](https://github.com/LastStep/Bonsai/pull/62) | 23 | gp + tl | 2026-04-22 |
| Plan 25 — README revamp: audience-first rewrite, new tagline, AI-readme smells cut. [plan](Plans/Archive/25-readme-revamp.md) · [PR #61](https://github.com/LastStep/Bonsai/pull/61) | 25 | tl | 2026-04-22 |
| Plan 23 Phase 1 — `bonsai add` cinematic new-agent path: addflow package + 6-stage flow behind env gate. [plan](Plans/Archive/23-uiux-phase2-add.md) · [PR #59](https://github.com/LastStep/Bonsai/pull/59) | 23 | gp + tl | 2026-04-22 |
| Pre-launch security sweep: Astro XSS auto-fix, 2 CodeQL alerts silenced, push-CI widened. Commits `ef55f4f` + hotfix `7c1dd49`. | — | tl | 2026-04-22 |
| Plan 24 — pre-launch polish: CHANGELOG curated, docs.yml PR trigger, 5 `good-first-issue` filed. [plan](Plans/Archive/24-pre-launch-polish.md) · [PR #58](https://github.com/LastStep/Bonsai/pull/58) | 24 | gp | 2026-04-22 |
| statusLine redesign: persistent bar with caveman badge + tier coloring. Direct-to-main. | — | tl | 2026-04-22 |
| Plan 22 **complete** — Phase 5B: wire Generate+Planted into `runInit`, flip default, delete legacy init. `runInitRedesign` → `runInit`, `BONSAI_REDESIGN` env-flag removed, 245 lines of legacy `cmd/init.go` deleted. Step order: Vessel → Soil → Branches → Observe → Conditional[Lazy[Generate]] → LazyGroup[conflicts] → Conditional[Lazy[Planted]]. Harness extension: `LazyStep.Chromeless()` + `ConditionalStep.Chromeless()` delegation + `ConditionalStep.Init` auto-builds nested Lazy. +260/−297 across 8 files. PR #52 squash `5916e05`. 6/6 CI green. | 22 | general-purpose | 2026-04-21 |
| Plan 22 Phase 5A — responsive resize foundation + Observe stage wired + Generate/Planted packaged. `internal/tui/initflow/layout.go` new: `MinTerminalWidth=70`, `MinTerminalHeight=20`, `ClampColumns(120) → (24, 44, 12)` regression anchor, hand-rolled `Viewport`. Branches/Vessel/Soil retrofit. `ObserveStage` replaces idx=3 stub. `GenerateStage` + `PlantedStage` packaged (not wired). 50+ new tests. PR #51 squash `6baaf8e` + lint-fix `3e31967`. | 22 | general-purpose | 2026-04-21 |
| Plan 22 Phase 4 polish run — 5 direct-to-main dogfood commits against Branches stage (no PRs per fast-iter UX memory): header split, per-tab 2-line intro, DETAILS moved to fixed-height box below list, right-aligned DEFAULT tags; density cut + word-wrapped details; wider layout + rune-aware truncation; ABOUT+FILE values → `ColorAccent` white + 3-row wrap; kanji padding + blank line. | 22 | tech-lead | 2026-04-21 |
| Plan 22 Phase 4 — `BranchesStage` tabbed picker across Skills/Workflows/Protocols/Sensors/Routines with inline-expand (`?`), per-category multi-select, defaults pre-seeded, required items pinned. 5-tab UI, `BranchesResult` shape, Reset() preserves state. 12 tests (PR #50 squash `89c21ba`). | 22 | general-purpose | 2026-04-21 |
| Plan 22 Phase 3.5 — dogfood polish pass (direct-to-main): Bark→gold `#D4AF37`, new Moon white token + ColorAccent=Moon, drop kana row, rail capped 60 cells, footer muted rule, 2-col field layout, focus-tinted underline, stable input cell width, copy refresh. | 22 | tech-lead | 2026-04-21 |
| Plan 22 Phase 3 — `VesselStage` (3 textinputs) + `SoilStage` (hand-rolled multi-select) + `RenderHeader` signature strip `stationSubdir` (PR #49 squash `971ee44`). | 22 | general-purpose | 2026-04-21 |
| Plan 22 Phase 2 — `internal/tui/initflow/` package (chrome/enso/fallback/stage/stub) + `harness.Chromeless` + env-flag `runInitRedesign` (PR #48 squash `2e2a08c`). | 22 | general-purpose | 2026-04-21 |
| Plan 22 Phase 1 — `RenderFileTree` widget + `ColorLeafDim`/`ColorRule`/`ColorRule2` palette tokens (PR #47 squash `7553d43`). | 22 | general-purpose | 2026-04-21 |
| Session-start context dedup + Phase C sensors — session-context redundancy cuts, context-guard verify/plan patterns, new compact-recovery sensor, UX prefs moved to memory.md Feedback (PR #46 squash `d14edbe`). | 21 | general-purpose | 2026-04-21 |
| Better trigger sections — Phase C (new sensors: compact-recovery + context-guard expansion) shipped via Plan 21 / PR #46; C3 deferred. | 08 | general-purpose | 2026-04-21 |
| Security scanning infra — Go 1.25.8 bump, golangci-lint v2 + pin, Dependabot, govulncheck CI, CodeQL workflow, gitleaks history audit clean (PRs #28 #29 #30 #31 #40 #41). | 20 | general-purpose | 2026-04-21 |
| Pre-launch bug sweep — 8 OSS-blocker fixes: CRLF, cross-workspace tree, dedup, spinner errors.Join, harness polish (PR #27 squash `a44e447`). | 19 | general-purpose | 2026-04-21 |
| BubbleTea foundation + theme system — harness migration across init/add/remove/update (PR #26 squash `2ce63f6`). | 15 | general-purpose | 2026-04-20 |
| `bonsai guide` multi-topic — 3 cheatsheets + delete 3 orphan docs (1,213L) + CLI refactor (PR #25). | 18 | general-purpose | 2026-04-20 |
| Fix `go install` binary name — main.go → cmd/bonsai/ + root embed package (PR #23). | 16 | general-purpose | 2026-04-20 |
| Release prep — Go toolchain 1.24.13 + triggerSection frontmatter fix + OSS polish (PR #24). | 17 | general-purpose | 2026-04-17 |
| UI/UX overhaul — Phase 3 visual identity + init polish (merged via PR #24 bundle). | 14 | general-purpose | 2026-04-17 |
| ActionUnchanged follow-ups — chmod bug + test gaps (PR #22). | 13 | tech-lead | 2026-04-17 |
| UI/UX overhaul — Phase 2 (consistency: hints, counts, no-op detection, structured errors). | 12 | tech-lead | 2026-04-17 |
| Documentation site (Starlight) — Phase D (deploy & CI) — Plan 10 complete. | 10 | tech-lead | 2026-04-17 |
| Documentation site (Starlight) — Phase C (catalog auto-generation + LLM layer). | 10 | tech-lead | 2026-04-17 |
| Documentation site (Starlight) — Phase B (fill content gaps, 30 pages). | 10 | tech-lead | 2026-04-17 |
| UI/UX overhaul — Phase 1 (adaptive palette, NO_COLOR, FatalPanel, version banner). | 11 | tech-lead | 2026-04-17 |
| Code index refresh — fix line drift + missing entries. | 09 | tech-lead | 2026-04-16 |
| Better trigger sections — Phase B (trigger documentation). | 08 | tech-lead | 2026-04-16 |
| Better trigger sections — Phase A (trigger metadata system). | 08 | tech-lead | 2026-04-16 |
| Community health files + README polish for public release. | 07 | tech-lead | 2026-04-16 |
| Fix case-insensitive file collision (index.md / INDEX.md). | 06 | tech-lead | 2026-04-16 |
| README rewrite for open-source release. | — | tech-lead | 2026-04-16 |
| Cross-link generated files — Obsidian-compatible markdown links. | 03 | tech-lead | 2026-04-16 |
| AI operational intelligence — How to Work + workspace-guide skill. | 05 | tech-lead | 2026-04-16 |
| Release pipeline — GoReleaser + GitHub Actions + Homebrew Tap. | 04 | tech-lead | 2026-04-15 |
| Rename "catalog items" to "abilities" — CLI, TUI, README, docs, comments. | — | tech-lead | 2026-04-15 |
| `bonsai guide` command — render custom files guide in terminal. | 02 | tech-lead | 2026-04-15 |
| CLAUDE.md marker migration — backup + overwrite for marker-less files. | 01 | tech-lead | 2026-04-15 |
| Selective file update — multi-select conflict picker. | — | tech-lead | 2026-04-15 |
| Fix doubled path prefix in `bonsai add` output panels. | — | tech-lead | 2026-04-15 |
| `bonsai update` — custom file detection + workspace sync. | — | tech-lead | 2026-04-14 |
| Dogfooding — station workspace setup + customization. | — | tech-lead | 2026-04-14 |
| Session wrap-up workflow + context-guard wiring. | — | tech-lead | 2026-04-14 |
| Stale artifact cleanup (Python refs, index.md rewrite). | — | tech-lead | 2026-04-14 |
| Awareness Framework — status-bar + context-guard sensors. | — | tech-lead | 2026-04-13 |
| Lock file conflict handling. | — | tech-lead | 2026-04-13 |
| Catalog expansion — all 3 phases, 6 agent types. | — | tech-lead | 2026-04-13 |
| Go rewrite from Python. | — | tech-lead | 2026-04-12 |

---

## Resolved Backlog Items

> Items removed from `Backlog.md` once shipped. Newest first.

### 2026-04-22 — v0.2.0 cycle resolutions

- **Pre-release docs audit across all user-facing content** `[Group C]` — Resolved via PR #65 squash `e9b7cba` / v0.2.0 cut. 4-surface audit (website, root community files, catalog, station/) found 3 P0 + 6 P1 + 4 P2; bundle delivered 13 fixes mechanically + agent-synthesized CHANGELOG. 2 deferred (filed as Plan 26 candidates): full website concept-page rewrite (heavier framing pivot), skills frontmatter convention.
- **Consolidate or delineate CHANGELOG backlog items** `[Group C/D]` — Resolved via Plan 24 (Group C CHANGELOG.md shipped; Group D changelog-generation item kept as future work).
- **Root `Bonsai/CLAUDE.md` project-structure tree drift** `[Group C/E]` — Resolved via Plan 24 Step C (PR #58 / `4ef8271`). `internal/tui/` block refreshed for Plan 15 `harness/` + Plan 22 `initflow/` + `filetree*` + `styles_test.go`.
- **Run Astro build on PRs touching `website/`** `[Group C]` — Resolved via Plan 24 Step B (PR #58 / `4ef8271`). `docs.yml` `pull_request` trigger + deploy job gated on `push`.
- **CHANGELOG.md and richer release notes** `[Group C]` — Resolved via Plan 24 Step A. Keep-a-changelog 1.1.0 format, curated v0.1.0–v0.1.3 backfill.
- **Seed GitHub Issues for contributor on-ramp** `[Group C]` — Resolved via Plan 24 Step E. 5 issues filed: [#53](https://github.com/LastStep/Bonsai/issues/53) statusLine port, [#54](https://github.com/LastStep/Bonsai/issues/54) shell completion, [#55](https://github.com/LastStep/Bonsai/issues/55) `bonsai changelog`, [#56](https://github.com/LastStep/Bonsai/issues/56) catalog umbrella, [#57](https://github.com/LastStep/Bonsai/issues/57) `.bak` merge hint — all `good first issue` + `help wanted`.
- **Harness composition test for `NewConditional(NewLazy(...))`** `[Group B]` — Resolved via Plan 23 Phase 3 / PR #64. 3 tests added in `internal/tui/harness/steps_test.go`: Chromeless forwarding + skipped path + builder-fires-once-per-pass.
- **Remove dead post-harness Generate-error warning** `[Group B]` — Resolved via Plan 23 Phase 3 / PR #64. Deleted from BOTH `cmd/init_flow.go` AND `cmd/add.go` — symmetric path in cinematic add caught by code review.
- **`growSucceeded` predicate walks `prev` tail for any error** `[Group B]` — Resolved via Plan 23 Phase 3 / PR #64. Predicate now reads `outcome.SpinnerErr` + `outcome.Ran` via closure capture; no `prev[]` walk, no sentinel needed.
- **Unknown-agent path renders `YieldTechLeadRequired` — copy mismatch** `[Group B]` — Resolved via Plan 23 Phase 3 / PR #64. New `addflow.NewYieldUnknownAgent` variant with `bonsai update` CTA + test.
- **`.bak` write-error silent-discard in both conflict-apply helpers** `[Group B]` — Resolved via Plan 23 Phase 3 / PR #64. BOTH `applyCinematicConflictPicks` + `applyConflictPicks`: failed-backup paths dropped from overwrite list + collected `tui.Warning`; security review confirmed regression closed; tests use real OS perms.
- **`confIdx := len(results) - 2` arithmetic** `[Group B]` — Resolved via Plan 23 Phase 3 / PR #64. Replaced with type-scan for `map[string]config.ConflictAction`.
- **No direct unit test for `applyCinematicConflictPicks`** `[Group B]` — Resolved via Plan 23 Phase 3 / PR #64. `cmd/add_test.go` new — 9 table tests including backup-read-fail and backup-write-fail with real OS conditions.
- **Re-plan "Better trigger sections — Phase C"** `[Ungrouped P2]` — Resolved via Plan 21 / PR #46 squash `d14edbe` (2026-04-21). New `compact-recovery` sensor + `context-guard` verify/plan patterns shipped; C3 (Haiku prompt-hook intent classification) deferred to P3 Research per Plan 08 closeout.
