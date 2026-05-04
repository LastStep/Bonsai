---
tags: [playbook, status]
description: Live task tracker. Update this file at the start and end of every working session.
---

# Bonsai — Status

> [!note]
> Move items between tables as work progresses. Link to the relevant plan in `Plans/Active/` or `Plans/Archive/`. Older Done items age out to `StatusArchive.md`.
>
> **Brevity rule:** every row follows [Standards/NoteStandards.md](Standards/NoteStandards.md) — 3 lines max, link out for detail. Phase walkthroughs go in the plan; commit walkthroughs go in the PR; process narrative goes in `Logs/`.

---

## In Progress

| Task | Plan | Agent | Notes |
|------|------|-------|-------|
| Plan 34 — custom-ability discovery bug bundle: orphaned-registration recovery, bash-shebang sensor frontmatter, non-TTY invalid-file warning, bonsai-model.md doc fix. [plan](Plans/Active/34-custom-ability-discovery-bug-bundle.md) | 34 | gp + tl | Tier 1, 4 fixes, single worktree |

## Pending

| Task | Plan | Agent | Blocked By |
|------|------|-------|------------|
<!-- Plan 26 candidates (skills frontmatter convention) filed in Backlog — pick up as next sweep -->

## Recently Done

| Task | Plan | Agent | Date |
|------|------|-------|------|
| Plan 32 — followup bundle: wsvalidate extract, Validate() chokepoint, O_NOFOLLOW snapshot. 13/17 review items closed. [plan](Plans/Archive/32-followup-bundle.md) · [PR #80](https://github.com/LastStep/Bonsai/pull/80) | 32 | gp×2 + tl | 2026-04-25 |
| Plan 33 — website concept-page rewrite: 3 files, README-aligned mechanism-led voice, 7 banned phrases scrubbed. [plan](Plans/Archive/33-website-concept-page-rewrite.md) · [PR #79](https://github.com/LastStep/Bonsai/pull/79) | 33 | gp + tl | 2026-04-25 |
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

> Done items older than 2026-04-22 moved to [StatusArchive.md](StatusArchive.md).
