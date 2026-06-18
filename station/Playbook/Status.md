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
| **Plan 41 — Headless CLI Contract + MCP-ready cores SHIPPED** — all 5 phases merged (PRs #120/#122/#123/#121/#125, main `ab202c3`). Every mutating cmd (init/add/update/remove) has a pure `*Result` headless core + JSONL/exit contract (`ExitConflict=5`); `list --json`; `docs/agent-interface.md` contract doc. Phase 1 solo (sig change) then 2/3/4 parallel, each reviewed + gates green (byte-identity oracle, symlink/multi-owner/required guards). Folded in pre-existing errcheck fix (main red since #116). MCP server = fast-follow Plan 42. Debt filed: unify remove cinematic/headless logic (Backlog P2). [plan](Plans/Active/41-headless-cli-contract.md) | 41 | tl + gp×5 | 2026-06-16 |
| **Plan 40 Phases 1–3 merged (v0.5.0, untagged)** — frozen v1 schemas + root-relative scaffolding (manifest + memory) [PR #114], project-level `validate` pass w/ adversarial path/symlink hardening [PR #116], memory-routing docs + guide Formats page [PR #115]. 1 blocking sec-bug caught in review (out-of-tree read via traversing `memory_dir`) + fixed. **Phase 4 HELD**, **dogfood deferred** (blocked: no update-delivery until P4 + repo gitignores `.bonsai-lock.yaml`), **tag held** (user). [plan](Plans/Active/40-odysseus-platform-integration.md) | 40 | gp×5 + tl | 2026-06-13 |

> Done items older than 14 days (≤ 2026-05-31) moved to [StatusArchive.md](StatusArchive.md).
