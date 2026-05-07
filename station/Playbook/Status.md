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

## Pending

| Task | Plan | Agent | Blocked By |
|------|------|-------|------------|
| **[research] Trial sentrux on Bonsai repo** — one-shot eval, build from source `/tmp/sentrux-trial/`, run `sentrux check .` + `sentrux scan .`, judge actionable/noise/wrong, adopt vs drop. [Backlog P0](Backlog.md#p0--critical) | — | tl | Rust toolchain (cargo/rustc) not installed — needs rustup install before trial |
<!-- Plan 26 candidates (skills frontmatter convention) filed in Backlog — pick up as next sweep -->

## Recently Done

| Task | Plan | Agent | Date |
|------|------|-------|------|
| **v0.4.0 release shipped** — Plan 36 release prep + hotfix: x/net v0.53.0 + Go 1.25.9 + workflow_dispatch retry + doc sweep + CHANGELOG + Windows cross-compile fix. 6 platform binaries + Homebrew v0.4.0 published. [plan](Plans/Archive/36-v04-release-prep.md) · [PR #94](https://github.com/LastStep/Bonsai/pull/94) · [hotfix #95](https://github.com/LastStep/Bonsai/pull/95) | 36 | gp×2 + tl | 2026-05-04 |
| Plan 35 — `bonsai validate` command: read-only ability-state audit, 6 issue categories, --json + --agent flags. v0.4.0 headline. [plan](Plans/Archive/35-bonsai-validate-command.md) · [PR #93](https://github.com/LastStep/Bonsai/pull/93) | 35 | gp + tl | 2026-05-04 |
| Plan 34 — custom-ability discovery bug bundle: orphan-registration recovery in scan, bash-shebang sensor frontmatter parser, non-TTY invalid-file stderr warning, bonsai-model.md hand-edit ban. [plan](Plans/Archive/34-custom-ability-discovery-bug-bundle.md) · [PR #92](https://github.com/LastStep/Bonsai/pull/92) | 34 | gp + tl | 2026-05-04 |
| Plan 32 — followup bundle: wsvalidate extract, Validate() chokepoint, O_NOFOLLOW snapshot. 13/17 review items closed. [plan](Plans/Archive/32-followup-bundle.md) · [PR #80](https://github.com/LastStep/Bonsai/pull/80) | 32 | gp×2 + tl | 2026-04-25 |
| Plan 33 — website concept-page rewrite: 3 files, README-aligned mechanism-led voice, 7 banned phrases scrubbed. [plan](Plans/Archive/33-website-concept-page-rewrite.md) · [PR #79](https://github.com/LastStep/Bonsai/pull/79) | 33 | gp + tl | 2026-04-25 |
| Plan 31 — v0.3 release readiness: peer-awareness refresh + bonsai-model skill + catalog.json snapshot, removeflow + updateflow cinematics, hints package. 3 PRs. [plan](Plans/Archive/31-v03-release-readiness.md) · [#75](https://github.com/LastStep/Bonsai/pull/75) · [#76](https://github.com/LastStep/Bonsai/pull/76) · [#77](https://github.com/LastStep/Bonsai/pull/77) | 31 | tl + gp×5 | 2026-04-24 |
| Plan 30 — guide viewer perf + view-cmds polish: glamour renderer cache + pre-warm, 18 NIT cleanups. [plan](Plans/Archive/30-guide-perf-and-view-polish.md) · [PR #74](https://github.com/LastStep/Bonsai/pull/74) | 30 | gp×3 + tl | 2026-04-23 |
| Archive-reconcile sweep: 20 shipped plans Active→Archive, 14 stale frontmatters synced. Commit `45df4cd`. | — | tl | 2026-04-23 |
| Plan 29 — init+add bug bundle: 8 phases, 8 bugs + 4 test gaps, workspace path-escape validators. [plan](Plans/Archive/29-init-add-bug-bundle.md) · [PR #72](https://github.com/LastStep/Bonsai/pull/72) | 29 | gp + tl | 2026-04-23 |
| Plan 28 — `bonsai list` + `bonsai guide` cinematics shipped in parallel. [plan](Plans/Archive/28-view-cmds-cinematic.md) · [#71](https://github.com/LastStep/Bonsai/pull/71) · [#70](https://github.com/LastStep/Bonsai/pull/70) | 28 | gp×4 + tl | 2026-04-23 |

> Done items older than 2026-04-23 moved to [StatusArchive.md](StatusArchive.md).
