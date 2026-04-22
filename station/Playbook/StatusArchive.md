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
