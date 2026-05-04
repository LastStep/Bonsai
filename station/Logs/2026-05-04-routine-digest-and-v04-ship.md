---
tags: [log, session, release, v0.4]
description: 2026-05-04 PM session — routine digest + v0.4.0 release prep + Windows cross-compile hotfix.
---

# 2026-05-04 — Routine Digest + v0.4.0 Ship + Hotfix

**Session:** afternoon block. Caveman mode active throughout.

## Outcomes

- **v0.4.0 released.** 6 platform binaries + checksums.txt published, Homebrew formula bumped to `0.4.0`. [Release](https://github.com/LastStep/Bonsai/releases/tag/v0.4.0).
- **Plan 36** drafted, dispatched, merged ([#94](https://github.com/LastStep/Bonsai/pull/94)). 4 file-disjoint phases: x/net v0.38→v0.53, Go toolchain 1.25.8→1.25.9, `workflow_dispatch` retry hook, doc-drift sweep, CHANGELOG.
- **Hotfix [#95](https://github.com/LastStep/Bonsai/pull/95)** — Windows cross-compile broken since Plan 32 (`syscall.O_NOFOLLOW` POSIX-only). Split into platform files. Tag force-moved (zero blast radius — no release artifacts existed yet).
- **Routine digest** for 3 overdue routines (dep-audit, vuln-scan, doc-freshness) — cleanest cycle on record (0 reachable CVEs, 0 secrets, 5 prior flags resolved).
- **`bonsai-model` skill** installed for tech-lead workspace (closed digest's broken-link finding).

## Key decisions / non-obvious moves

- **Plan 36 was Tier 2 multi-domain but file-disjoint** — single-agent bundled dispatch instead of parallel. Single PR, less coordination overhead, ~30 min plan-to-merge.
- **Force-move `v0.4.0` tag** instead of cutting `v0.4.1`. Justified because no release artifacts existed yet (GoReleaser failed pre-publish). Preserved planned version number.
- **Dogfooded Plan 36's `workflow_dispatch` retry hook within hours of shipping it** — `gh workflow run release.yml --ref v0.4.0 -f tag=v0.4.0` recovered the failed release without re-tagging. Validated the design.

## Cycles

- Plan 36: 1 dispatch, 1 review (PASS), merge. 0 fix-iterations.
- Hotfix #95: 1 dispatch (no review agent — surgical fix), merge. 0 fix-iterations.
- Total: 2 worktree dispatches, 1 review agent, 1 audit agent (post-ship verification).

## Memory updates

- **Durable gotcha:** `syscall.O_NOFOLLOW` is POSIX-only — Windows cross-compile breaks. CI Linux-only doesn't catch. Memory note added with mitigation pattern (platform-split files).
- **Stale memory cleared:** prior note implying "v0.3.0 not tagged" was wrong; tag was at `ac59f8b` (2026-04-24). Plan 36 research surfaced this.

## Backlog deltas

**Closed (4):** `[debt] workflow_dispatch on release.yml` P1; `[security] x/net bump` P2; `[security] Go 1.25.9 bump` P2; `[debt] Plan 36 docs sweep` P2.

**Filed (1):** `[ops] Windows cross-compile gate to ci.yml` P2 — prevents the v0.4.0 incident class.

**Promoted (1):** root `Bonsai/CLAUDE.md` tree-drift check P3 → P2 (3rd-cycle recurrence).

**Narrowed (1):** `[improvement] Install semgrep` — gitleaks half done; semgrep still pending.

## Caveats / followups

- **Stale worktree + branch sweep** (P1 backlog line 57) untouched — ~17 worktrees, ~20 stale branches. Pre-existing housekeeping.
- **CodeQL Action v3 → v4** (P1) deferred — v4 not yet released.
- **23-module hygiene refresh** (P3) — separate post-release sweep PR when prioritized.
- **HOMEBREW_TAP_TOKEN PAT rotation calendar** (P1) — due ~2026-07-15.

## Commits this session

```
864241b chore: post-ship audit cleanup — CHANGELOG hotfix mention + RoutineLog + ci backlog
7414a88 chore(station): v0.4.0 ship-out logging + Windows cross-compile gotcha
c5f6dc1 fix(generate): split catalog_snapshot O_NOFOLLOW for Windows compat (#95)
81826b1 chore(station): archive Plan 36 — v0.4.0 shipped
b9ade62 chore(release): v0.4.0 prep — toolchain bump, retry hook, doc sweep, CHANGELOG (#94)
dcc9143 chore(station): routine digest 2026-05-04 + bonsai-model install + Plan 36 draft
```
