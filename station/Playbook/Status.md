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
| _(none)_ | | | |

## Pending

| Task | Plan | Agent | Blocked By |
|------|------|-------|------------|
| **[research] Trial sentrux on Bonsai repo** — one-shot eval, build from source `/tmp/sentrux-trial/`, run `sentrux check .` + `sentrux scan .`, judge actionable/noise/wrong, adopt vs drop. [Backlog P0](Backlog.md#p0--critical) | — | tl | Rust toolchain (cargo/rustc) not installed — needs rustup install before trial |
<!-- Plan 26 candidates (skills frontmatter convention) filed in Backlog — pick up as next sweep -->

## Recently Done

| Task | Plan | Agent | Date |
|------|------|-------|------|
| _(none — all completed items aged out; see StatusArchive.md)_ | | | |

> Done items older than 14 days (≤ 2026-08-18) moved to [StatusArchive.md](StatusArchive.md). Last sweep: 2026-09-01 (status-hygiene routine, 16 items archived).
