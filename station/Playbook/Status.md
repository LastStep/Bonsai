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
| `bonsai init` cinematic redesign — 5 phases / 5 PRs, behind `BONSAI_REDESIGN=1` until Phase 5 flips default | 22 | general-purpose | Phase 1 shipped (PR #47). Next: Phase 2 (initflow package + chrome + env-flag entrypoint). |
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
