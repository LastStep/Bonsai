---
tags: [playbook, backlog]
description: Prioritized backlog — bugs, features, debt, research, and improvement ideas. Self-maintained by agents via the backlog-hygiene routine.
---

# Bonsai — Backlog

> [!note]
> This is the intake queue for all work not yet in `Status.md`. Items flow from here into active work.
> For current active work, see `Playbook/Status.md`. For long-term direction, see `Playbook/Roadmap.md`.

---

## How This Works

**Capture:** When you discover a bug, improvement opportunity, tech debt, or idea during a session that is outside your current task scope — add it here instead of fixing it inline. Use the item format below.

**Promote:** When capacity opens, move P0/P1 items into `Playbook/Status.md` as Pending or In Progress. Remove the item from this file when it appears in Status.

**Resolve:** Items completed via Status.md are cleaned up by the backlog-hygiene routine. Items abandoned or made irrelevant should be removed with a note in `Logs/RoutineLog.md`.

**Review:** The backlog-hygiene routine runs periodically to flag stale items, escalate misplaced P0s, remove duplicates, and cross-reference with Status.md and Roadmap.md.

### Item Format

```markdown
- **[category] Short description** — Context or rationale. *(added YYYY-MM-DD, source: routine|session|user)*
```

**Categories:** `bug`, `feature`, `debt`, `security`, `research`, `improvement`

### Priority Guide

| Priority | Meaning | Action |
|----------|---------|--------|
| **P0** | Blocking current work or broken functionality | Must be in Status.md. If a P0 is here, escalate it immediately |
| **P1** | Next up when current work completes | Promote to Status.md when capacity opens |
| **P2** | Planned but not urgent | Review at phase boundaries |
| **P3** | Ideas, nice-to-haves, research topics | Review during roadmap updates |

### Groups

Items that should be worked together are tagged with a group letter. See the group index at the bottom of this file for phasing and dependency info.

---

## P0 — Critical

(none)

## P1 — High

- **[debt] Add `workflow_dispatch:` trigger to `.github/workflows/release.yml`** — Currently only `on: push: tags: v*`. When GoReleaser fails partway (e.g. the v0.2.0 Homebrew step at 401 Bad credentials), there's no clean way to retry without re-tagging. Add `workflow_dispatch:` so failed steps can be retried via `gh workflow run release.yml --ref v0.X.Y`. **Note:** plain retry will run `goreleaser release --clean` which can recreate the GH Release with new timestamps — should pair this with `--skip=publish` or split brew into a separate job. *(added 2026-04-22, source: v0.2.0 release session — Homebrew PAT expired post-tag-push, had to push formula manually via `gh api PUT`)*
- **[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder** — Fine-grained PATs default to 90-day expiry. The `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai` was rotated 2026-04-22 — set calendar reminder for ~2026-07-15 to rotate before next release. Same applies to any other PATs in repo secrets (audit + document expiry dates). Symptom of expired PAT: GoReleaser fails at brew step with `GET https://api.github.com/repos/LastStep/homebrew-tap: 401 Bad credentials` — release otherwise succeeds (binaries published, only formula update missed). *(added 2026-04-22, source: v0.2.0 release session)*
- **[debt] CodeQL Action v3 → v4** — GH deprecation notice: CodeQL Action v3 will be deprecated December 2026. Update `.github/workflows/codeql.yml` `github/codeql-action/{init,autobuild,analyze}` pins from `@v3` to `@v4` when v4 ships and Dependabot opens the bump PR. No urgency — lots of runway. *(added 2026-04-21, source: session — surfaced on PR #38 CI run)*
- **[debt] Testing infrastructure for triggers and sensors** `[Group B]` — No testing infrastructure exists for hook-based triggers, prompt hooks, context-guard regex patterns, path-scoped rules, or skill auto-invocation. Need: (1) unit tests for context-guard regex patterns (positive/negative cases), (2) integration test harness for sensor scripts (mock stdin, verify stdout/exit codes), (3) end-to-end test framework for trigger activation (simulate user prompts, verify correct ability loads), (4) prompt hook evaluation testing (verify Haiku correctly classifies intents). The trigger system is expanding significantly — without test infra, regressions will be invisible. *(added 2026-04-16, source: user)*
- **[debt] Stale agent worktrees + branches accumulating** `[housekeeping]` — 2026-04-21 audit: 17+ `.claude/worktrees/agent-*` (several locked on UNC `//wsl.localhost/...` paths from cross-OS sessions), 20+ stale remote branches on `origin/` (all from merged PRs #1-#27), 18+ local branches. Root cause: `gh pr merge --delete-branch` silently skips branch deletion when its worktree is checked out (memory.md doc'd 5× this session). Linux-side worktrees + branches are safely prunable; UNC ones need Windows-side. Suggested: one-time sweep via `git worktree remove -f -f`, `git branch -D`, `git push origin --delete`. Then add a station routine to prune merged worktrees weekly. *(added 2026-04-20, updated 2026-04-21, source: session)*


## P2 — Medium

### Group B: Code Quality & Testing

> Logical ordering: split the big file first (makes testing easier), then add tests, then fix error handling. The remaining P1 bug (spinner error swallowing) can be fixed independently at any time. (triggerSection frontmatter bug fixed in Plan 17 / PR #24.)

- **[debt] Break up `generate.go` — 1,357 lines, highest churn file** — `internal/generate/generate.go` is both the largest Go file and the most frequently modified. It handles file writing, template rendering, conflict resolution, lock management, sensor/routine wiring, and scaffolding — too many responsibilities in one file. Split along natural seams: (1) template rendering, (2) file writing + conflict resolution, (3) lock management, (4) sensor/routine wiring. Would improve testability and reduce merge friction for agent dispatches. *(added 2026-04-16, source: repo-analytics)*
- **[debt] `internal/catalog/` test coverage — 496 lines, 0%** — Catalog loading (`LoadCatalog()`, `DisplayNameFrom()`, meta.yaml parsing) is the bridge between embedded YAML and the rest of the system. A malformed `meta.yaml` in the catalog would break at runtime with no test to catch it. Basic tests for catalog loading, display name derivation, and agent compatibility filtering would catch regressions cheaply. *(added 2026-04-16, source: repo-analytics)*
- **[debt] CLI command test coverage — `cmd/` package at 0%** — The `cmd/` package contains all user-facing CLI logic (init, add, remove, update, list, catalog, guide) — 1,691 lines across 8 files, zero tests. Priority targets: (1) `cmd/init.go` — happy path e2e test (temp dir, verify output structure), (2) `cmd/add.go` — test that abilities land correctly, (3) `cmd/remove.go` — test clean removal (472 lines, 4th largest file). Table-driven tests with temp dir setup would cover the most ground. *(added 2026-04-16, source: repo-analytics)*
<!-- "Harness composition test for NewConditional(NewLazy(...))" — resolved 2026-04-22 via Plan 23 Phase 3 / PR #64 (3 tests added in internal/tui/harness/steps_test.go: Chromeless forwarding + skipped path + builder-fires-once-per-pass) -->
<!-- "Remove dead post-harness Generate-error warning" — resolved 2026-04-22 via Plan 23 Phase 3 / PR #64 (deleted from BOTH cmd/init_flow.go AND cmd/add.go — symmetric path in cinematic add caught by code review) -->
- **[debt] PTY smoke test for harness-driven CLI commands** — `internal/tui/harness/` reducer tests are TTY-free (`fakeStep` + message injection) which catches logic bugs but can't drive a real `bonsai init`/`add`/`remove`/`update` end-to-end. Add a PTY-based smoke test using `creack/pty` or similar: spawn the built binary, send scripted keystrokes, assert the post-exit filesystem state (config written, workspace generated, lockfile valid). Would catch regressions unit tests miss — huh state transitions, AltScreen entry/exit, embedded form focus. Scope covers iter 1's `bonsai init` + iter 2's `bonsai add` + iter 3's `remove`/`update`. *(added 2026-04-20, source: Plan 15 iter 1 report — out-of-scope followup)*
- **[debt] Routines dashboard table split in `RoutineDashboard()`** — Routines table in `station/CLAUDE.md` (lines 72-83) and `station/agent/Core/routines.md` dashboard (lines 33-42) has a blank row that splits the table into two fragments, breaking GitHub/Obsidian markdown rendering. Second fragment shows without headers. Root cause in `RoutineDashboard()` at `internal/generate/generate.go:884` — frequency-based grouping emits a blank line mid-table. Fix the generator, then also correct the two already-generated files. *(added 2026-04-21, source: routine-digest)*
<!-- "growSucceeded predicate walks prev tail for any error" — resolved 2026-04-22 via Plan 23 Phase 3 / PR #64 (predicate now reads outcome.SpinnerErr + outcome.Ran via closure capture; no prev[] walk, no sentinel needed) -->
<!-- "Unknown-agent path renders YieldTechLeadRequired — copy mismatch" — resolved 2026-04-22 via Plan 23 Phase 3 / PR #64 (new addflow.NewYieldUnknownAgent variant with `bonsai update` CTA + test) -->
- **[bug] context-guard planning-reminder path uses wrong prefix** — `station/agent/Sensors/context-guard.sh` planning-trigger reminder builds paths as `{root}/agent/Workflows/planning.md` but `agent/` lives under `station/agent/`. Result: the injected reminder points to a non-existent path, which is harmless at runtime (agent just can't open the file) but misleading. Fix: change `os.path.join(root, "")` → `os.path.join(root, "station", "")` in the planning reminder section (and verify the same pattern isn't replicated in other inject blocks). Pre-existing — noticed 2026-04-22 during statusLine refactor; out-of-scope so not fixed inline. *(added 2026-04-22, source: session)*
<!-- ".bak write-error silent-discard in both conflict-apply helpers" — resolved 2026-04-22 via Plan 23 Phase 3 / PR #64 (BOTH applyCinematicConflictPicks + applyConflictPicks: failed-backup paths dropped from overwrite list + collected tui.Warning; security review confirmed regression closed; tests use real OS perms) -->
<!-- "confIdx := len(results) - 2 arithmetic" — resolved 2026-04-22 via Plan 23 Phase 3 / PR #64 (replaced with type-scan for map[string]config.ConflictAction) -->
<!-- "No direct unit test for applyCinematicConflictPicks" — resolved 2026-04-22 via Plan 23 Phase 3 / PR #64 (cmd/add_test.go new — 9 table tests including backup-read-fail and backup-write-fail with real OS conditions) -->

- **[debt] `selected[:0]` aliasing in conflict-apply filter loops** `[Plan-23 cosmetic]` — Both `applyCinematicConflictPicks` (`cmd/add.go`) and `applyConflictPicks` (`cmd/root.go`) use `filtered := selected[:0]` then write into the shared backing array while iterating `selected`. Provably safe today (write index ≤ read index per loop ordering) and tests pass — but is a known anti-pattern that would silently break if the loop ordering ever flipped. Suggested swap: `filtered := make([]string, 0, len(selected)-len(dropped))`. Independent code review (PR #64) flagged as minor cosmetic. *(added 2026-04-22, source: PR #64 review)*
- **[debt] `installedSet` shadowing in `cmd/add.go`** `[Plan-23 cosmetic]` — File-scope `installedSet(*config.ProjectConfig)` is shadowed by an inner closure `installedSet := func(items []string)` in `distributeAddItemPicks`. Go-legal (different signatures) but mildly confusing for readers. Rename the inner closure (e.g. `installedItems`). Independent code review (PR #64) flagged as minor cosmetic. *(added 2026-04-22, source: PR #64 review)*
- **[improvement] `generate.FileResult` has no inline-diff field — `ConflictsStage.renderDiffSummary` uses placeholder** — `internal/tui/addflow/conflicts.go:320-327` emits a stable "local has edits · source has changes" placeholder with an inline TODO because `generate.FileResult` exposes no summary/patch. Expose a diff-summary field (or a helper computing line-count delta) from `internal/generate/` and feed it into `renderDiffSummary`. Would make the cinematic conflict picker materially more useful than the legacy Huh list. *(added 2026-04-22, source: PR #62 review)*

### Group C: OSS Readiness

> All support the public repo being contributor-friendly. Small, independent — could knock both out in one session.

- **[improvement] OSS polish — demo GIF/asciinema for README** — Last remaining OSS readiness item after Plan 17. Linter config (`.golangci.yml` with errcheck/govet/unused/misspell/gofmt/goimports) and Makefile `test`/`lint`/`fmt`/`tidy` targets shipped in PR #24. Still need a demo GIF or asciinema recording to add under README hero image — requires user recording (not agent-able). When recording: show `bonsai init` flow, add a skill, run `bonsai list`. *(added 2026-04-16, narrowed 2026-04-17 post-Plan-17, source: RESEARCH-oss-readiness.md cleanup)*
<!-- "Seed GitHub Issues for contributor on-ramp" — resolved 2026-04-22 via Plan 24 Step E (5 issues filed: #53 statusLine port, #54 shell completion, #55 bonsai changelog, #56 catalog umbrella, #57 .bak merge hint — all labeled good first issue + help wanted) -->
<!-- "CHANGELOG.md and richer release notes" — resolved 2026-04-22 via Plan 24 Step A (keep-a-changelog 1.1.0 format, curated v0.1.0–v0.1.3 backfill) -->
<!-- "Run Astro build on PRs touching website/" — resolved 2026-04-22 via Plan 24 Step B (docs.yml pull_request trigger + deploy gated on push, PR #58 / 4ef8271) -->
<!-- "Root Bonsai/CLAUDE.md project-structure tree drift (Group E duplicate)" — resolved 2026-04-22 via Plan 24 Step C (internal/tui/ block refreshed for Plan 15 harness/ + Plan 22 initflow/ + filetree* + styles_test.go, PR #58) -->

<!-- "Consolidate or delineate CHANGELOG backlog items" — resolved 2026-04-22 via Plan 24 (Group C CHANGELOG.md shipped; Group D changelog-generation item kept as future work) -->
<!-- "Pre-release docs audit across all user-facing content" — resolved 2026-04-22 via PR #65 squash `e9b7cba` / v0.2.0 cut. 4-surface audit (website, root community files, catalog, station/) found 3 P0 + 6 P1 + 4 P2; bundle delivered 13 fixes mechanically + agent-synthesized CHANGELOG. 2 deferred: full website concept-page rewrite (heavier framing pivot), skills frontmatter convention — see Plan 26 candidates below. -->
- **[improvement] Plan 26 candidate — full website concept-page rewrite** — Plan 25 README revamp pivoted README to "A workspace for your coding agent" framing; Plan v0.2.0 audit (PR #65) did mechanical tagline replace across `index.mdx`, `astro.config.mjs`, `why-bonsai.mdx` only. Heavier rewrite still pending: `concepts/how-bonsai-works.mdx`, `why-bonsai.mdx` body paragraphs, `astro.config.mjs:28` LLM description ("structured instruction files... so AI agents work like teammates, not tools" — AI-smell), and concepts pages should pivot from "structured language" framing to mechanism/workspace framing matching README's audience-first pitch. Estimate ~1 hr focused copy work. *(added 2026-04-22, source: v0.2.0 audit Q3=C scope split)*
- **[debt] Plan 26 candidate — skills frontmatter convention decision** — 13 of 17 skills lack YAML frontmatter (only `dispatch`, `issue-classification`, `pr-creation`, `workspace-guide` have it). Catalog loader reads triggers from `meta.yaml` not file frontmatter, so this is cosmetic — but a consistent convention should be picked: (a) require frontmatter on all skills, (b) drop frontmatter from the 4 that have it, (c) document that frontmatter is optional and skip-list specific use cases. Once decided, mass-apply across catalog. *(added 2026-04-22, source: v0.2.0 audit P2 deferred)*
<!-- "Pre-release docs audit across all user-facing content" duplicated above — see resolved comment block -->
- **[improvement] Pre-release docs audit across all user-facing content** — Before public launch announce, do a single pass across every user-facing doc: `README.md`, `SECURITY.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, all Starlight pages under `website/src/content/docs/` (concepts, guides, commands, catalog, reference, index/getting-started/installation/why-bonsai/faq/troubleshooting), and the in-terminal `bonsai guide` cheatsheets. Criteria: (1) every doc is still necessary (no legacy migrations lingering), (2) no duplicated content across docs, (3) content is focused — doesn't waste the reader's time, (4) all cross-links resolve. User flagged concern: "lets not update those docs now, we will do one final pass before release, on all the docs, to ensure we are delivering quality content, and not wasting someone's time." Run last before any release-promotion activity. *(added 2026-04-20, source: user — OSS readiness session)* — **RESOLVED 2026-04-22 via PR #65 / v0.2.0 cut. Kept here as historical reference; remove on next backlog-hygiene pass.**

### Group D: Catalog Expansion

> Research first (concept-decisions), then build. The concept-decisions review informs which of the others to prioritize.

- **[research] Revisit concept-decisions research** — Review `station/Research/concept-decisions.md` for unbuilt concepts that may be worth promoting: (1) **Talents** — a new catalog category for innate behavioral aptitudes, (2) **Meta-layer** — runtime system-wide observation layer, (3) **Three-layer catalog ownership model**, (4) **Loading gradient** reasoning. Decide which to build, which to backlog properly, which to discard. *(added 2026-04-16, source: research doc cleanup)*
- **[feature] Unbuilt catalog items — 3 agents, 1 skill, 4 routines** — From the catalog expansion research, 8 items were never built: **agents** `qa`, `reviewer`, `docs`; **skill** `documentation-standards` (blocks `docs` agent); **routines** `test-coverage-check` (qa), `changelog-maintenance` (docs), `api-docs-drift` (docs), `standards-drift` (reviewer). Build order: `documentation-standards` skill first (unblocks `docs`), then agents, then routines. *(added 2026-04-16, source: RESEARCH-catalog-expansion.md cleanup)*
- **[feature] Changelog generation skill + release changelogs** — Add a changelog generation skill that: (1) parses conventional commit messages between tags to generate structured changelogs, (2) outputs CHANGELOG.md, (3) generates release notes for `gh release create --notes`. Current releases (v0.1.0-v0.1.3) shipped with no changelogs — backfill them. Also consider a `bonsai changelog` CLI command. *(added 2026-04-16, source: user)* (refiled as good-first-issue via Plan 24 Step E)
- **[feature] Research scaffolding item + abilities** — Add an optional `Research/` folder to project scaffolding for storing landscape analysis, concept decisions, and design research. Add associated abilities (tech-lead only): a research workflow and/or a research-template skill. *(added 2026-04-16, source: user)*

### Group E: Workspace Improvements

> Small, independent quality-of-life items. Can be done in any order.

- **[improvement] Plan archiving — Active/Archive folder structure** — Plans currently all live in `Plans/Active/`. Completed plans should move to `Plans/Archive/` after merge. Requires: create `Plans/Archive/` in scaffolding manifest, update issue-to-implementation workflow (Phase 10), update planning workflow and planning-template skill, update session-start protocol if it scans for active plans, update CLAUDE.md nav table. *(added 2026-04-16, source: user)*
<!-- "Root Bonsai/CLAUDE.md project-structure tree drift (Group E)" — resolved 2026-04-22 via Plan 24 Step C (PR #58 / 4ef8271). See resolution comment at top of Group C. -->

- **[improvement] Plans Index file** — No Plans Index exists; flagged by 2026-04-20 Status Hygiene and re-surfaced in 2026-04-21 Backlog Hygiene. Decide: add a `Plans/INDEX.md` listing active/archived plans with one-line summaries, or fold into the "Plan archiving" item above as a sub-task. *(added 2026-04-21, source: routine-digest)*
- **[improvement] Consolidate FieldNotes usage** — The current `Logs/FieldNotes.md` file has unclear purpose and overlaps with other state files (memory.md, Status.md, KeyDecisionLog.md). Rethink what it's for, whether it should be merged into another artifact, and how it fits into the session-start context injection. *(added 2026-04-15, source: user)*
- **[improvement] Post-update backup merge hint** — After `bonsai update` creates `.bak` backups during conflict resolution, print a hint telling the user to ask their agent to reconcile customizations. Small change to `cmd/update.go` after `resolveConflicts()` returns. *(added 2026-04-16, source: user)*
- **[feature] Port statusLine to catalog sensor** — filed as issue [#53](https://github.com/LastStep/Bonsai/issues/53) on 2026-04-22 with full background, findings from prototype, acceptance criteria, testing plan, and proposed implementation (in issue comments). Prototype lives at `station/agent/Sensors/statusline.sh` + manual stanza in `station/.claude/settings.json`. Deferred execution — pick up via `/issue-to-implementation` when prioritized. *(added 2026-04-22, source: session)*

### Group F: UI/UX Testing Findings

> Dogfooding findings from the 2026-04-17 `bonsai init` walk-through. Mostly `init` flow polish — visual identity, prompt flow ergonomics, and the review→generate→complete flow.
>
> **Status 2026-04-22:** 9 of 11 init-UX items shipped via Plan 22 + the 2026-04-22 dogfood polish run (full list in `StatusArchive.md` Backlog Resolutions). Group F essentially closed for `init`; UI/UX overhaul now moves on to `add`/`update`/`remove`/`list`/`catalog`/`guide` under a new plan (Phase 2 of the overhaul).

- **[docs] Document AltScreen behavior change in release notes** — Plan 15 migrates `bonsai init` and `bonsai add` into AltScreen, which means the interactive flow no longer accretes to scrollback line-by-line — users who rely on copy/pasting partially-filled screens out of their terminal history will notice. Add a release-note bullet when iter 3 ships and the whole branch merges. *(added 2026-04-20, source: Plan 15 iter 2.1 review)*
- **[docs] Fill "Deviations from Plan" in completion reports more aggressively** — The iter-2 completion report at `Reports/Archive/2026-04-20-plan-15-iter-2-add-migration.md` listed three deviations but missed noting that the iter-2 "pre-harness tech-lead gate" design choice in the plan was itself a regression (fixed in 2.1). Implementing agents should err toward over-documenting plan↔implementation divergence, since the gaps are where post-ship reviewers find bugs. Tweak `planning-template` or `review-checklist` skill to prompt for this. *(added 2026-04-20, source: Plan 15 iter 2.1 review)*

### Ungrouped P2

- **[feature] Developer guide for Bonsai contributors** — Write a `DEVELOPMENT.md` (or docs site page) covering the internal dev workflow: how to build and test locally, `npm run generate:catalog` usage and when to run it, release checklist, catalog structure conventions, testing against a temp dir, and other commands/processes a contributor needs to know. Currently this info is scattered across CLAUDE.md and tribal knowledge. *(added 2026-04-17, source: user)*
- **[feature] Routine report template** — Add a `routine-report-template.md` to `station/Reports/` alongside the existing `report-template.md`. Routine reports have a different shape than plan completion reports. The template in `loop.md` defines the format; this makes it a first-class project artifact. *(added 2026-04-14, source: user)*
- **[improvement] Split design-guide: generic catalog skill + Bonsai-specific station override** — Plan 11 replaces `catalog/skills/design-guide` with Bonsai-specific TUI/CLI rules (paths like `internal/tui/**`, palette tokens). That's useful for dogfooding but irrelevant when external users install the skill. Follow-up: restore a generic Charm/Go CLI design-guide to the catalog (palette patterns, NO_COLOR support, panel vocabulary as principles — not specific paths), and move Bonsai's own rules into `station/agent/Skills/design-guide.md` as a local override. *(added 2026-04-17, source: plan-11 audit)*
- **[improvement] Install semgrep and/or gitleaks for better scanning** — Vulnerability scan and secrets scan routines currently use manual pattern-based Grep scanning. Installing semgrep (SAST) and/or gitleaks (secrets) would improve coverage and reduce false negatives. *(added 2026-04-16, source: routine-digest)*
- **[security] Bump `golang.org/x/net` v0.38.0 → v0.45.0+** — Clears GO-2026-4441 (infinite parsing loop in `golang.org/x/net`) and GO-2026-4440 (quadratic parsing complexity in `golang.org/x/net/html`). Both unreachable package-level CVEs today but easy hygiene. Run `go get golang.org/x/net@latest && go mod tidy`, verify `govulncheck ./...` clean. Should ship alongside or after the P1 Go toolchain upgrade. *(added 2026-04-21, source: routine-digest)*
- **[improvement] Re-plan "Better trigger sections — Phase C"** — Status.md Pending row's Blocked-By condition ("UI/UX Phase 3 ships") was resolved 2026-04-17 via Plan 14 / PR #24. Either re-plan Phase C for execution (promote to Status.md In Progress with a fresh plan) or update the blocker note to a current reason. Long-overdue — two backlog-hygiene cycles have flagged it. *(added 2026-04-21, source: routine-digest)*

## P3 — Ideas & Research

### Future Platform (Roadmap Phase 2+)

- **[feature] Integration scaffolding variants** — Support alternative backends for all PM artifacts (backlog, status, roadmap, reports). During `bonsai init`, user picks a backend per artifact: markdown (default), GitHub Issues, Notion, Jira, etc. Affects: scaffolding manifest, agent instructions, protocols, any sensor/workflow that references PM files. *(added 2026-04-15, source: user)*
- **[feature] Enhanced session-start sensor — project pulse** — Expand `session-context.sh` to inject a project-state summary at session start. Phase 1: check markdown PM files directly. Phase 2: auto-detect from integration variant in `.bonsai.yaml`. *Depends on: integration scaffolding variants.* *(added 2026-04-15, source: user)*
- **[feature] Custom item creator** — Interactive TUI for creating custom items (skill, workflow, protocol, sensor, routine) with frontmatter scaffolding — similar to Claude's skill creator. *(added 2026-04-14, source: user)*
- **[improvement] Self-update mechanism** — Skills and workflows should be able to self-heal or flag when they have issues. *(added 2026-04-13, source: user)*
- **[improvement] Micro-task fast path** — Define an explicit lightweight protocol for trivial changes (< 50 LOC, no architectural impact). Could be a sensor that auto-detects task weight, or a protocol clause agents check before entering full planning mode. *(added 2026-04-15, source: architectural audit)*

### Routine System Enhancements

- **[feature] Scheduled task generation for routines** — Auto-generate Claude scheduled task configs from routine metadata. Maps routine `frequency` to cron expressions. Two tiers: file-only routines → Desktop local tasks; heavier routines → Cloud routines. *Note: Cloud routines are in research preview — API/limits may change.* *(added 2026-04-15, source: user)*
- **[feature] Routine GitHub issue creation** — Routines can create GitHub issues for actionable findings. Per-routine opt-in via `creates_issues: true` in `meta.yaml`. Issues get a `bonsai-routine` label. Requires `gh` CLI. *Dependency: prerequisite for auto-fixer routine.* *(added 2026-04-15, source: user)*
- **[feature] Auto-fixer routine** — New routine that polls GitHub issues labeled `bonsai-routine`, attempts autonomous fixes. *Depends on: routine GitHub issue creation + scheduled task generation.* *(added 2026-04-15, source: user)*
- **[improvement] Add root `Bonsai/CLAUDE.md` check to doc-freshness-check routine** — Recurring drift pattern: whenever `cmd/` or `internal/` layout changes (Plans 15, 16, 18, 09 all affected), the root CLAUDE.md Project Structure tree falls out of date for 1-2 weeks before being noticed. Add a sub-step to `catalog/routines/doc-freshness-check/content.md.tmpl` that diffs the tree block against actual `cmd/` + `internal/` layout. *(added 2026-04-21, source: routine-digest)*
- **[improvement] Reduce `npm audit` cadence in dependency-audit routine** — `website/` npm audit has returned 0 vulnerabilities for multiple consecutive 7-day scans. Consider adding every-other-run skip logic (track last-npm-audit date in routine state) to save scan time. Revisit if a vulnerability surfaces. *(added 2026-04-21, source: routine-digest)*

### Outreach

- **[feature] README case study / blog post from session-transcript metrics** — The 2026-04-16 transcript analysis (archived at `station/Reports/Archive/2026-04-16-session-transcript-analysis.md`) contains compelling quantitative data: 20 sessions over 6 days, ~1,186 user messages (~90% silent tool approvals), ~2,000 substantive words total drove an entire CLI tool from Go rewrite through OSS release. Specific hooks: "75-message silent approval streak," "10+ deliverables in 48 minutes," "14 of 20 sessions had zero user-initiated rework." Could seed a README "Real-World Usage" section, a standalone case study page on the docs site, or a blog post. Parts 5 (session typology) and 7 (metrics) are the primary source material. *(added 2026-04-17, source: session — pending-report review)*

### Research

- **[research] Session-start payload — further optimization** — Plan 21 (2026-04-21, PR #46) cut ~10% (34.3KB→30.9KB) by removing redundant protocol dumps, end-of-session misfire, empty FieldNotes, and Reports full-cat. Remaining ~30KB is mostly identity.md + memory.md + self-awareness.md + INDEX.md + Status.md full-dump on every SessionStart. Ideas: (a) diff-based injection (only dump sections that changed since last session), (b) summary + link pattern (first 10 lines + "read full via tool if needed"), (c) conditional injection based on session type (fresh vs resume). Would need a sensor-side state cache. *(added 2026-04-15, updated 2026-04-21 post-Plan-21, source: architectural audit + Plan 21 findings)*
- **[research] Plan 08 C3 — prompt hook intent classification** — Deferred in original Plan 08 verification because auto-invocation via `.claude/skills/` (Phase A) + context-guard phrase regex (Phase C2 shipped 2026-04-21) were expected to cover the same workflows. Revisit when we have signal that the 3 target workflows (code-review, pr-review, security-audit) are NOT reliably auto-invoked. Trigger: user reports missed activation, or telemetry shows skill-description fuzzy-match misses. Cost: ~$0.001/prompt Haiku classification. *(added 2026-04-21, source: Plan 08 Phase C closeout)*
- **[research] Parallel agent coordination in shared repos** — Research how multiple code agents can work simultaneously on different tasks in the same repository. Key questions: git workflow, file contention, lock/claim protocol, orchestration model, state coherence, tooling. *(added 2026-04-16, source: user)*
- **[research] Archon analysis** — <https://github.com/coleam00/Archon> — research what it does, use cases, overlap with Bonsai, what we can learn. *(added 2026-04-13, source: user)*
- **[debt] Batch refresh outdated Go modules after toolchain upgrade** — 17 modules behind per `go list -m -u all` 2026-04-21: `golang.org/x/crypto v0.36 → v0.50`, `x/tools v0.37 → v0.44`, `x/sys v0.38 → v0.43`, `x/text v0.30 → v0.36`, `x/mod v0.28 → v0.35`, `uax29/v2 v2.5 → v2.7`, `goldmark v1.7.13 → v1.8.2`, plus charm/bubbletea ecosystem. No CVEs beyond govulncheck coverage. Hygiene sweep after P1 Go toolchain upgrade lands. *(added 2026-04-21, source: routine-digest)*
- **[security] Pin `website/package.json` deps to specific versions — drop `"latest"` ranges** — All 5 website deps (`astro`, `@astrojs/starlight`, `js-yaml`, `starlight-links-validator`, `starlight-llms-txt`) use `"latest"` which lets npm pull an arbitrary version on any fresh `npm ci` even though `package-lock.json` pins exact versions. Supply-chain posture: if an upstream package gets compromised, a fresh clone could pick it up. Fix: replace `"latest"` with `"^x.y.z"` caret ranges matching currently-resolved versions; rely on Dependabot for bumps. Caught 2026-04-22 during security sweep (CVE-2026-41067 Astro XSS was auto-fixed only because lockfile happened to resolve 6.1.7 ≥ patched 6.1.6). *(added 2026-04-22, source: session — security sweep)*

### Big Bets

- **[feature] Managed Agents integration** — Cloud deployment via `bonsai deploy`, session management, outcome rubrics in catalog. Build after local foundation is stable. *(added 2026-04-13, source: user)*
- **[feature] Greenhouse companion app** — Desktop app for managing projects + observing AI agents. Design doc: DESIGN-companion-app.md. Stack: Tauri v2 + Svelte 5 + SQLite. Status: Design phase, decisions locked. *(added 2026-04-13, source: user)*
- **[improvement] Catalog display_name audit** — Add explicit `display_name` to all catalog `meta.yaml` files. Research other metadata fields that could be useful (e.g., `version`, `tags`, `dependencies`, `examples`). *(added 2026-04-14, source: user)*

---

## Group Index

| Group | Theme | Phase Order | Notes |
|-------|-------|-------------|-------|
| **A** | Documentation Suite | Quickstart → Concepts → CLI Usage → Multi-topic command | Resolves Roadmap "Usage instructions". Content first, CLI wiring last. |
| **B** | Code Quality & Testing | Split generate.go → catalog tests → cmd tests → trigger test infra → spinner error fix | P1 bugs (frontmatter, spinners) can be fixed independently at any time. |
| **C** | OSS Readiness | Linter + Makefile → seed GitHub Issues | Small, one-session effort. |
| **D** | Catalog Expansion | Concept-decisions review → documentation-standards skill → agents → routines → changelog | Research informs build order. |
| **E** | Workspace Improvements | Any order | Independent quality-of-life items. |
| **F** | UI/UX Testing | Any order | Findings from dogfooding session on 2026-04-17 — CLI polish, install UX, prompt flow issues. |
