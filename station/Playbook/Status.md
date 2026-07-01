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
| **v0.4.3 hotfix shipped** — sensor hook commands now bake install-time absolute paths in `.claude/settings.json` (vs `$PWD`-walk-up which drifted across sibling Bonsai projects). Surfaced during Bonsai-Eval bootstrap. Existing projects need `bonsai update` to refresh. [release](https://github.com/LastStep/Bonsai/releases/tag/v0.4.3) · [PR #105](https://github.com/LastStep/Bonsai/pull/105) · [PR #106](https://github.com/LastStep/Bonsai/pull/106) · commit `584b82b` | — | tl + gp | 2026-05-13 |
| **Plan 38 handoff to Bonsai-Eval tech-lead** — bootstrapped `LastStep/Bonsai-Eval` via `bonsai init --non-interactive` (commit `057f378`). Plan moved to that repo's `Plans/Active/`; this station archives the local copy. P2/P3 owned there going forward. [Bonsai-Eval station](https://github.com/LastStep/Bonsai-Eval/tree/main/station) | 38 | tl | 2026-05-13 |
| **v0.4.2 release shipped** — `bonsai init`/`add` `--non-interactive --from-config <path>` (JSONL stdout, hard-skip conflicts, exit codes 0/2/3/4). Unblocks Plan 38 P2 rung-3. Plan 39 sequential A→B→C→D + 3 review-driven fix-ups. [release](https://github.com/LastStep/Bonsai/releases/tag/v0.4.2) · [PR #102](https://github.com/LastStep/Bonsai/pull/102) · commit `410a5f1` | 39 | gp + tl | 2026-05-13 |
| **PR triage sweep** — closed 9 stale routine bot PRs (#86–91, #96–98) superseded by local routine-digests (`dcc9143`, `39ee362`); merged 4 Dependabot bumps: codeql-action v3→v4 (#85), checkout v4→v6 (#81), deploy-pages v4→v5 (#82), go-isatty 0.0.21→0.0.22 (#84). Closes 2 P1 backlog rows (CodeQL v3→v4, Node 20→24). [Backlog](Backlog.md) bot pile-up follow-up filed. | — | tl | 2026-05-07 |
| **First external contribution merged** — `bonsai completion [bash|zsh|fish|powershell]` from @mvanhorn. Closes #54. CI green, squash-merged. Stale-comment fixup follow-up. [PR #78](https://github.com/LastStep/Bonsai/pull/78) · commit `2eae9d4` | — | tl | 2026-05-07 |
| **v0.4.1 release shipped** — quiet patch: Windows cross-compile CI gate + root CLAUDE.md Go drift fix. CHANGELOG entry, 6 platform binaries + Homebrew published. [release](https://github.com/LastStep/Bonsai/releases/tag/v0.4.1) · commit `533d112` | — | tl | 2026-05-07 |
| Windows cross-compile CI gate — `GOOS=windows GOARCH=amd64 go build` step added to `ci.yml` test job. Catches POSIX-only divergence (v0.4.0 `syscall.O_NOFOLLOW` class). [Backlog](Backlog.md) P2 row cleared. | — | tl | 2026-05-07 |
| Root CLAUDE.md Go drift fix — `Go 1.24+ → 1.25+` one-liner. Plan 37 followup. [Backlog](Backlog.md) row cleared. | — | tl | 2026-05-07 |

> Done items older than 14 days (≤ 2026-05-07) moved to [StatusArchive.md](StatusArchive.md).
