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
<!-- Documentation site (Starlight) — Plan 10 complete, all phases shipped (PRs #13-19) -->
<!-- UI/UX overhaul Phase 2 — merged PR #20, moved to Recently Done 2026-04-17 -->
<!-- ActionUnchanged follow-ups — merged PR #22, moved to Recently Done 2026-04-17 -->
| BubbleTea foundation + theme system (migrate init/add/remove/update onto shared harness) | 15 | general-purpose | On `ui-ux-testing` branch. Iter 1 (init) + iter 2 (add) + iter 2.1 (reviewer fixes) shipped locally @ `d0e6256`. Iter 3 remains: `remove.go` + `update.go` + carry-forward nits |
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
| Better trigger sections — Phase C (new sensors) | 08 | tech-lead | Phases A+B shipped; Phase C paused while UI/UX Phase 3 ships |

## Recently Done

| Task | Plan | Agent | Date |
|------|------|-------|------|
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
