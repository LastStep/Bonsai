---
tags: [session-log]
description: Plan 20 security scanning infra + Dependabot Group G triage — 15 PRs shipped, main at 5248212.
---

# 2026-04-21 — Plan 20 + Group G (15 PRs)

## Headline

Shipped **15 PRs** in one day: 6 Plan 20 security-scanning PRs, 1 Plan 20 wrap-up, 8 Dependabot Group G weekly-drift PRs, 1 Group G closeout. Zero rollbacks. Main advanced `afbcd6d` → `5248212`.

## Plan 20 — Security Scanning Infrastructure

Goal: layered automated scanning stack for OSS launch — dep CVEs, Go stdlib CVEs, SAST, secrets.

**Merged PRs** (sequential, each on fresh main):
- #30 `chore(station)` pre-flight commit (caught local-only main state)
- #29 `ci: migrate golangci-lint v1 to v2` (config schema + action pin v8)
- #28 `build(deps): bump Go toolchain to 1.25.8` (bundled lint pin v2.11.4)
- #31 `ci: add Dependabot config` (gomod + github-actions, weekly Mon, 5 PR limit)
- #40 `ci: add govulncheck job`
- #41 `ci: add CodeQL workflow for Go SAST`
- #42 `chore(station) Plan 20 wrap-up`

**Post-PRs:** gitleaks history audit — **0 findings across 156 commits**.
**govulncheck on main:** 0 reachable findings (previously 2 reachable stdlib CVEs GO-2026-4602 + GO-2026-4601, cleared by Go 1.25.8).

## Group G — Dependabot Weekly Drift

PRs #31 Dependabot landed → immediately opened 8 drift PRs. Merged all 8 same day.

**gomod (3):**
- #39 charmbracelet/x/ansi 0.11.6→0.11.7
- #38 mattn/go-isatty 0.0.20→0.0.21 (1× rebase for go.sum conflict)
- #37 golang.org/x/term 0.36.0→0.42.0 (1× rebase)

**github-actions (5):**
- #35 actions/setup-go v5→v6
- #34 golangci/golangci-lint-action v8→v9
- #36 actions/setup-node v4→v6
- #33 actions/upload-pages-artifact v3→v5
- #32 goreleaser/goreleaser-action v6→v7

**Closeout:** #43 `chore(station) Group G Dependabot closeout`.

**Release-notes review findings (batch 3, not CI-exercised):**
- setup-node v5 auto-caching keyed on `packageManager` field → **inactive** (website/package.json has no such field; explicit `cache: npm` unchanged).
- upload-pages-artifact v4 drops dotfiles from artifact → **safe** (website/dist has no dotfiles; `_astro`/`_llms-txt` are underscore not dot).
- goreleaser-action v7 = node 24 runtime, ESM, inputs unchanged → **safe**.

## Gotchas captured this session

All promoted to memory.md for cross-session persistence:

1. **golangci-lint binary Go-version coupling.** Moving go.mod to Go 1.25 required golangci-lint v2.11.4+ action pin. v2.1 binary was built on Go 1.24 and errored: `the Go language version (go1.24) used to build golangci-lint is lower than the targeted Go version (1.25.8)`. Rule: when bumping Go in go.mod, pin golangci-lint-action to a version whose notes say "built with Go ≥ target".

2. **Agent worktrees base off origin/main, not local main.** Agent(isolation: worktree) fetches origin/main. Any uncommitted/unpushed local main state is invisible to dispatched agents. For Plan 20 pre-flight this meant a local-only chore commit had to ship as its own PR #30 first.

3. **`gh pr merge --delete-branch` silent half-failure.** Confirmed 5× more this session (#28 #29 #31 #40 #41 + all 8 Group G). Always follow with `git push origin --delete <br>` + `git worktree remove -f -f <path>` + `git branch -D <br>`.

4. **Dependabot weekly opens fire immediately on config landing.** PR #31 merged → 8 PRs opened within ~60s, not waiting for next Monday. Plan scope creep: triage had to happen today.

5. **CodeQL Action v3 deprecation** (Dec 2026) — notice surfaced on PR #38 CI output. Backlog item queued.

## Session-end state

- **Main:** `5248212c9e4f4b61087f35845a3c7e3be39e3bca`
- **Open PRs:** 0
- **Uncommitted:** none in station/
- **Stale artifacts discovered:** 17+ agent worktrees, 20+ remote branches, 18+ local branches (all from merged PRs) — Backlog item updated.
- **Status.md:** Plan 20 row in Recently Done (via PR #42).

## Next

- Housekeeping sweep of stale worktrees/branches (Backlog P1 debt item, updated this session).
- CodeQL v3→v4 bump when Dependabot opens it (P1 debt, long runway).
- Phase C of Plan 08 (sensors) still paused — resume once UI/UX wraps.
