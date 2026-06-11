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
| — | — | — | — |

## Pending

| Task | Plan | Agent | Blocked By |
|------|------|-------|------------|
| **[research] Trial sentrux on Bonsai repo** — one-shot eval, build from source `/tmp/sentrux-trial/`, run `sentrux check .` + `sentrux scan .`, judge actionable/noise/wrong, adopt vs drop. [Backlog P0](Backlog.md#p0--critical) | — | tl | Rust toolchain (cargo/rustc) not installed — needs rustup install before trial |
<!-- Plan 26 candidates (skills frontmatter convention) filed in Backlog — pick up as next sweep -->

## Recently Done

| Task | Plan | Agent | Date |
|------|------|-------|------|
| **Plan 38 handoff to Bonsai-Eval tech-lead** — bootstrapped `LastStep/Bonsai-Eval` via `bonsai init --non-interactive` (commit `057f378`). Plan moved to that repo's `Plans/Active/`; this station archives the local copy. P2/P3 owned there going forward. [Bonsai-Eval station](https://github.com/LastStep/Bonsai-Eval/tree/main/station) | 38 | tl | 2026-05-13 |
| **v0.4.2 release shipped** — `bonsai init`/`add` `--non-interactive --from-config <path>` (JSONL stdout, hard-skip conflicts, exit codes 0/2/3/4). Unblocks Plan 38 P2 rung-3. Plan 39 sequential A→B→C→D + 3 review-driven fix-ups. [release](https://github.com/LastStep/Bonsai/releases/tag/v0.4.2) · [PR #102](https://github.com/LastStep/Bonsai/pull/102) · commit `410a5f1` | 39 | gp + tl | 2026-05-13 |
| **PR triage sweep** — closed 9 stale routine bot PRs (#86–91, #96–98) superseded by local routine-digests (`dcc9143`, `39ee362`); merged 4 Dependabot bumps: codeql-action v3→v4 (#85), checkout v4→v6 (#81), deploy-pages v4→v5 (#82), go-isatty 0.0.21→0.0.22 (#84). Closes 2 P1 backlog rows (CodeQL v3→v4, Node 20→24). [Backlog](Backlog.md) bot pile-up follow-up filed. | — | tl | 2026-05-07 |
| **First external contribution merged** — `bonsai completion [bash|zsh|fish|powershell]` from @mvanhorn. Closes #54. CI green, squash-merged. Stale-comment fixup follow-up. [PR #78](https://github.com/LastStep/Bonsai/pull/78) · commit `2eae9d4` | — | tl | 2026-05-07 |
| **v0.4.1 release shipped** — quiet patch: Windows cross-compile CI gate + root CLAUDE.md Go drift fix. CHANGELOG entry, 6 platform binaries + Homebrew published. [release](https://github.com/LastStep/Bonsai/releases/tag/v0.4.1) · commit `533d112` | — | tl | 2026-05-07 |
| Windows cross-compile CI gate — `GOOS=windows GOARCH=amd64 go build` step added to `ci.yml` test job. Catches POSIX-only divergence (v0.4.0 `syscall.O_NOFOLLOW` class). [Backlog](Backlog.md) P2 row cleared. | — | tl | 2026-05-07 |
| Root CLAUDE.md Go drift fix — `Go 1.24+ → 1.25+` one-liner. Plan 37 followup. [Backlog](Backlog.md) row cleared. | — | tl | 2026-05-07 |
| Plan 37 — doc refresh bundle: `code-index.md` 50+ line refs synced across cmd/ + internal/; 2 stale rows fixed (GraftStage→BranchesStage, NormaliseWorkspace dropped); INDEX.md `Go 1.24+ → 1.25+` drift. [plan](Plans/Archive/37-doc-refresh-bundle.md) | 37 | tl | 2026-05-07 |
| **v0.4.0 release shipped** — Plan 36 release prep + hotfix: x/net v0.53.0 + Go 1.25.9 + workflow_dispatch retry + doc sweep + CHANGELOG + Windows cross-compile fix. 6 platform binaries + Homebrew v0.4.0 published. [plan](Plans/Archive/36-v04-release-prep.md) · [PR #94](https://github.com/LastStep/Bonsai/pull/94) · [hotfix #95](https://github.com/LastStep/Bonsai/pull/95) | 36 | gp×2 + tl | 2026-05-04 |
| Plan 35 — `bonsai validate` command: read-only ability-state audit, 6 issue categories, --json + --agent flags. v0.4.0 headline. [plan](Plans/Archive/35-bonsai-validate-command.md) · [PR #93](https://github.com/LastStep/Bonsai/pull/93) | 35 | gp + tl | 2026-05-04 |

> Done items older than 14 days (≤ 2026-05-27) moved to [StatusArchive.md](StatusArchive.md).
