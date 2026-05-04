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

<!-- "workflow_dispatch trigger on release.yml" — resolved 2026-05-04 via Plan 36 / PR #94 (workflow_dispatch with tag input + checkout ref override). -->
- **[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder** — Fine-grained PATs default to 90-day expiry. The `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai` was rotated 2026-04-22 — set calendar reminder for ~2026-07-15 to rotate before next release. Same applies to any other PATs in repo secrets (audit + document expiry dates). Symptom of expired PAT: GoReleaser fails at brew step with `GET https://api.github.com/repos/LastStep/homebrew-tap: 401 Bad credentials` — release otherwise succeeds (binaries published, only formula update missed). *(added 2026-04-22, source: v0.2.0 release session)*
- **[debt] CodeQL Action v3 → v4** — GH deprecation notice: CodeQL Action v3 will be deprecated December 2026. Update `.github/workflows/codeql.yml` `github/codeql-action/{init,autobuild,analyze}` pins from `@v3` to `@v4` when v4 ships and Dependabot opens the bump PR. No urgency — lots of runway. *(added 2026-04-21, source: session — surfaced on PR #38 CI run)*
- **[debt] Testing infrastructure for triggers and sensors** `[Group B]` — No testing infrastructure exists for hook-based triggers, prompt hooks, context-guard regex patterns, path-scoped rules, or skill auto-invocation. Need: (1) unit tests for context-guard regex patterns (positive/negative cases), (2) integration test harness for sensor scripts (mock stdin, verify stdout/exit codes), (3) end-to-end test framework for trigger activation (simulate user prompts, verify correct ability loads), (4) prompt hook evaluation testing (verify Haiku correctly classifies intents). The trigger system is expanding significantly — without test infra, regressions will be invisible. *(added 2026-04-16, source: user)*
- **[debt] Stale agent worktrees + branches accumulating** `[housekeeping]` — 2026-04-21 audit: 17+ `.claude/worktrees/agent-*` (several locked on UNC `//wsl.localhost/...` paths from cross-OS sessions), 20+ stale remote branches on `origin/` (all from merged PRs #1-#27), 18+ local branches. Root cause: `gh pr merge --delete-branch` silently skips branch deletion when its worktree is checked out (memory.md doc'd 5× this session). Linux-side worktrees + branches are safely prunable; UNC ones need Windows-side. Suggested: one-time sweep via `git worktree remove -f -f`, `git branch -D`, `git push origin --delete`. Then add a station routine to prune merged worktrees weekly. *(added 2026-04-20, updated 2026-04-21, source: session)*
<!-- "Re-archive Plan 29 file" — resolved 2026-04-23 via archive-reconcile sweep. All 20 shipped plans moved Active → Archive + Status frontmatter synced from StatusArchive. -->


## P2 — Medium

### Group A: Bookkeeping

- **[bookkeeping] Retroactively trim Backlog entries to NoteStandards** — current entries embed file:line references, multi-paragraph rationales, and inline code blocks that belong in the linked source artifacts. New rule at [Standards/NoteStandards.md](Standards/NoteStandards.md) caps each entry at 3 lines + link out. Sweep all P0–P3 bullets, replace verbose prose with `[tag] Title — one-liner. *(source: link)*` shape. Same for `StatusArchive.md` Recently Done table. *(added 2026-04-25, source: Plan 32 wrap-up — Status row hit ~3KB single-row before NoteStandards rule)*

### Group B: Code Quality & Testing

> Logical ordering: split the big file first (makes testing easier), then add tests, then fix error handling. The remaining P1 bug (spinner error swallowing) can be fixed independently at any time. (triggerSection frontmatter bug fixed in Plan 17 / PR #24.)

- **[debt] Break up `generate.go` — 1,357 lines, highest churn file** — `internal/generate/generate.go` is both the largest Go file and the most frequently modified. It handles file writing, template rendering, conflict resolution, lock management, sensor/routine wiring, and scaffolding — too many responsibilities in one file. Split along natural seams: (1) template rendering, (2) file writing + conflict resolution, (3) lock management, (4) sensor/routine wiring. Would improve testability and reduce merge friction for agent dispatches. *(added 2026-04-16, source: repo-analytics)*
- **[debt] `internal/catalog/` test coverage — 496 lines, 0%** — Catalog loading (`LoadCatalog()`, `DisplayNameFrom()`, meta.yaml parsing) is the bridge between embedded YAML and the rest of the system. A malformed `meta.yaml` in the catalog would break at runtime with no test to catch it. Basic tests for catalog loading, display name derivation, and agent compatibility filtering would catch regressions cheaply. *(added 2026-04-16, source: repo-analytics)*
- **[debt] CLI command test coverage — `cmd/` package at 0%** — The `cmd/` package contains all user-facing CLI logic (init, add, remove, update, list, catalog, guide) — 1,691 lines across 8 files, zero tests. Priority targets: (1) `cmd/init.go` — happy path e2e test (temp dir, verify output structure), (2) `cmd/add.go` — test that abilities land correctly, (3) `cmd/remove.go` — test clean removal (472 lines, 4th largest file). Table-driven tests with temp dir setup would cover the most ground. *(added 2026-04-16, source: repo-analytics)*
- **[debt] PTY smoke test for harness-driven CLI commands** — `internal/tui/harness/` reducer tests are TTY-free (`fakeStep` + message injection) which catches logic bugs but can't drive a real `bonsai init`/`add`/`remove`/`update` end-to-end. Add a PTY-based smoke test using `creack/pty` or similar: spawn the built binary, send scripted keystrokes, assert the post-exit filesystem state (config written, workspace generated, lockfile valid). Would catch regressions unit tests miss — huh state transitions, AltScreen entry/exit, embedded form focus. Scope covers iter 1's `bonsai init` + iter 2's `bonsai add` + iter 3's `remove`/`update`. *(added 2026-04-20, source: Plan 15 iter 1 report — out-of-scope followup)*
<!-- "Routines dashboard table split" — resolved 2026-04-22 via Plan 26 / PR #66 (file was stale output; generator already clean; regenerated + TestRoutineDashboardNoBlankRows regression test added) -->
<!-- "context-guard planning-reminder path wrong prefix" — resolved 2026-04-22 via Plan 26 / PR #66 (swapped `os.path.join(root, "")` → existing `docs_path` var at lines 156-157) -->
<!-- "selected[:0] aliasing in conflict-apply filter loops [Plan-23 cosmetic]" — resolved 2026-04-22 via Plan 26 / PR #66 (both cmd/root.go:173 + cmd/add.go:329 now use `make([]string, 0, len(slice)-len(dropped))`) -->
<!-- "installedSet shadowing in cmd/add.go [Plan-23 cosmetic]" — resolved 2026-04-22 via Plan 26 / PR #66 (both inner closures at lines 532 + 619 renamed to `installedItems`; file-scope func preserved) -->
<!-- "generate.FileResult has no inline-diff field / renderDiffSummary placeholder" — stale 2026-04-23 (renderDiffSummary was removed in Plan 27 PR2 / PR #69 when conflicts.go was rewritten into the vertical list form; re-file if inline diffs are re-scoped) -->
<!-- "[Plan-27-cosmetic] ConflictsStage render polish bundle" — items 1/2/5 resolved 2026-04-23 via Plan 29 / PR #72 (conflicts.go dead `w` dropped, listHeight fixedRows 15→14 + comment reconciled, grow.go SetRailHidden redundant call dropped). Items 3/4/6 remain deferred: fixedRows magic-number rename (cosmetic), renderList viewport in View() vs Update() (repo-wide nit), renderKeyHintsInline vs RenderFooter join duplication (different seps). -->
<!-- "[Plan-27-test-gap] Add direct unit tests for PR2 additions" — all 4 resolved 2026-04-23 via Plan 29 / PR #72 (TestGenerateStage_BodyOnlyDropsChrome in initflow/generate_test.go, TestConflicts_ViewportFollowsFocus + TestConflicts_ColorTonesDifferPerAction + TestConflicts_LowercaseKMovesFocus in addflow/conflicts_test.go) -->
<!-- "[Plan-29-cosmetic] init+add bug bundle review minors" — all 3 items resolved 2026-04-25 via Plan 32 / PR #80 (Phase C Keep-vs-Backup tone assertion + shortName→conflictsShortName rename, Phase A error string "absolute paths not allowed (no leading / or drive letter)"). -->
- **[Plan-29-test-gap] happy-path validator coverage for Phase H** — (item 1 resolved 2026-04-25 via Plan 32 PR #80 Phase C — TestGround_AcceptsNestedRelative + TestVessel_AcceptsCleanRelative). Remaining: (2) `TestGenerateStage_BodyOnlyDropsChrome` doesn't verify inverse (chrome IS present when `SetBodyOnly(false)`). Add positive-chrome companion test or concrete pointer. *(added 2026-04-23, source: PR #72 review)*
- **[Plan-29-security-hardening] Phase H validator hardening** — PR #72 security review minors (all non-blocking, defence-in-depth polish). (items 1/2/3 resolved 2026-04-25 via Plan 32 PR #80: wsvalidate package extraction, backslash + pure-root rejection). Remaining: (4) Unicode lookalikes (`․․` U+2024, `．．` U+FF0E) pass as literal directory names. Filesystems don't resolve — no traversal — but NFKC normalisation before `..` scan would flag homoglyph attempts. Purely speculative. *(added 2026-04-23, source: PR #72 security review)*
<!-- "Cross-agent OtherAgents template staleness on bonsai add" — resolved 2026-04-24 via Plan 31 PR1 / PR #75 (generate.RefreshPeerAwareness re-renders identity.md + scope-guard-files.sh + dispatch-guard.sh for all peers excluding the newly-added agent; called from both cmd/add.go branches). -->
- **[Plan-31-cosmetic] PR #75 review minors** — non-blocking, low-value. (items 4/5 resolved 2026-04-25 via Plan 32 PR #80 Phase D: hasAbility → slices.Contains, agentsToSlice + requiredToSlice → compatToSlice). Remaining: (1) `internal/generate/catalog_snapshot.go:192, 209` uses legacy `0755`/`0644` octal not Go-1.13 `0o755`/`0o644` prescribed in plan; neighbors use legacy octal so local convention wins, leaving for awareness. (2) Same file — `WriteCatalogSnapshot` duplicates the "read-existing + bytes.Equal + ActionUnchanged" logic already present in `writeFile:297-300`. Could be `writeFileNoLock` helper but current inline readable enough; skip unless third caller needs it. (3) `bonsai_reference_test.go` — rendered Markdown table emits `[path](path)` where link-text = URL; neighboring Core/How-to-Work tables use human-readable labels. Cosmetic fix: `[bonsai-model.md](relative-path)` or prose label. (6) Every `bonsai add <ability>` pays the cost of 6×3 template re-renders per peer just to emit `ActionUnchanged`. Low-cost; worth a benchmark if dogfooding feels slower. *(added 2026-04-24, source: PR #75 review)*
<!-- "[Plan-31-test-gap] PR #75 coverage gaps" — both items resolved 2026-04-25 via Plan 32 PR #80 Phase E (TestWriteCatalogSnapshot_TrailingNewline + TestSerializeCatalog_VersionPassThrough table for "" / dev / v0.3.0). -->
- **[Plan-31-security-hardening] PR #75 defence-in-depth** — non-blocking, PRE-EXISTING weaknesses. (items 1/2 resolved 2026-04-25 via Plan 32 PR #80 Phase F: ProjectConfig.Validate() chokepoint wired into Load() + O_NOFOLLOW symlink-resistant write at WriteCatalogSnapshot. Item 4 also closed — DocsPath shell-metachar scan in Validate() catches `]` / `)` / `[` / `(` in the injection payload `x) [evil](http://attacker`). Remaining: (3) TOCTOU on `.bonsai/` dir perms — `os.MkdirAll(..., 0755)` silently succeeds if dir exists with different mode; Go doesn't chmod pre-existing dirs. Minor (contents non-secret). *(added 2026-04-24, source: PR #75 security review)*
<!-- "[Plan-28-cosmetic] Phase 1 review NITs" — resolved 2026-04-23 via Plan 30 / PR #74 (timeZero wrapper dropped, stale filterHidZero docstring fixed, expanded:false dropped, Agents Meta:{Kind:agent} added, greyed-tab style test added, dead-code sweep clean). -->
<!-- "[Plan-28-cosmetic] Phase 2+3 post-fix NITs" — resolved 2026-04-23 via Plan 30 / PR #74 (itoa → strconv.Itoa in both listflow + initflow, TestRenderWorkspaceBlock_EmptyString + TestRunList_NoAgents added, chromeRows magic → Stage.ChromeHeights() helper, deriveLabel unifies labelFor/shortFor fallback, NewStage("GUIDE") label args trimmed). -->

### Group C: OSS Readiness

> All support the public repo being contributor-friendly. Small, independent — could knock both out in one session.

- **[improvement] OSS polish — demo GIF/asciinema for README** — Last remaining OSS readiness item after Plan 17. Linter config (`.golangci.yml` with errcheck/govet/unused/misspell/gofmt/goimports) and Makefile `test`/`lint`/`fmt`/`tidy` targets shipped in PR #24. Still need a demo GIF or asciinema recording to add under README hero image — requires user recording (not agent-able). When recording: show `bonsai init` flow, add a skill, run `bonsai list`. *(added 2026-04-16, narrowed 2026-04-17 post-Plan-17, source: RESEARCH-oss-readiness.md cleanup)*

<!-- Closed 2026-04-25 by Plan 33 (PR #79 squash `b5345d6`) — full website concept-page rewrite shipped: how-bonsai-works.mdx + why-bonsai.mdx + astro.config.mjs LLM description, banned-phrase grep clean -->
- **[debt] Plan 26 candidate — skills frontmatter convention decision** — 13 of 17 skills lack YAML frontmatter (only `dispatch`, `issue-classification`, `pr-creation`, `workspace-guide` have it). Catalog loader reads triggers from `meta.yaml` not file frontmatter, so this is cosmetic — but a consistent convention should be picked: (a) require frontmatter on all skills, (b) drop frontmatter from the 4 that have it, (c) document that frontmatter is optional and skip-list specific use cases. Once decided, mass-apply across catalog. *(added 2026-04-22, source: v0.2.0 audit P2 deferred)*

### Group D: Catalog Expansion

> Research first (concept-decisions), then build. The concept-decisions review informs which of the others to prioritize.

- **[research] Revisit concept-decisions research** — Review `station/Research/concept-decisions.md` for unbuilt concepts that may be worth promoting: (1) **Talents** — a new catalog category for innate behavioral aptitudes, (2) **Meta-layer** — runtime system-wide observation layer, (3) **Three-layer catalog ownership model**, (4) **Loading gradient** reasoning. Decide which to build, which to backlog properly, which to discard. *(added 2026-04-16, source: research doc cleanup)*
- **[feature] Unbuilt catalog items — 3 agents, 1 skill, 4 routines** — From the catalog expansion research, 8 items were never built: **agents** `qa`, `reviewer`, `docs`; **skill** `documentation-standards` (blocks `docs` agent); **routines** `test-coverage-check` (qa), `changelog-maintenance` (docs), `api-docs-drift` (docs), `standards-drift` (reviewer). Build order: `documentation-standards` skill first (unblocks `docs`), then agents, then routines. *(added 2026-04-16, source: RESEARCH-catalog-expansion.md cleanup)*
- **[feature] Changelog generation skill + release changelogs** — Add a changelog generation skill that: (1) parses conventional commit messages between tags to generate structured changelogs, (2) outputs CHANGELOG.md, (3) generates release notes for `gh release create --notes`. Current releases (v0.1.0-v0.1.3) shipped with no changelogs — backfill them. Also consider a `bonsai changelog` CLI command. *(added 2026-04-16, source: user)* (refiled as good-first-issue via Plan 24 Step E)
- **[feature] Research scaffolding item + abilities** — Add an optional `Research/` folder to project scaffolding for storing landscape analysis, concept decisions, and design research. Add associated abilities (tech-lead only): a research workflow and/or a research-template skill. *(added 2026-04-16, source: user)*

### Group E: Workspace Improvements

> Small, independent quality-of-life items. Can be done in any order.

- **[improvement] Plan archiving — Active/Archive folder structure** — Plans currently all live in `Plans/Active/`. Completed plans should move to `Plans/Archive/` after merge. Requires: create `Plans/Archive/` in scaffolding manifest, update issue-to-implementation workflow (Phase 10), update planning workflow and planning-template skill, update session-start protocol if it scans for active plans, update CLAUDE.md nav table. *(added 2026-04-16, source: user)*

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

- **[improvement] Add root `Bonsai/CLAUDE.md` tree-drift check to doc-freshness-check routine** — Promoted from P3 → P2 after 3rd-cycle recurrence (2026-04-14, 2026-04-21, 2026-05-04 routine-digests all flagged the same drift class). Plans 15/16/18/22/23/27/28/30/31/32/35 each compounded the gap. Add a sub-step to `catalog/routines/doc-freshness-check/content.md.tmpl` that diffs the tree block in root `Bonsai/CLAUDE.md` against actual `cmd/` + `internal/` layout. Bake in permanently so future plans don't compound silently. *(added 2026-04-21, promoted 2026-05-04, source: routine-digest)*
- **[feature] Developer guide for Bonsai contributors** — Write a `DEVELOPMENT.md` (or docs site page) covering the internal dev workflow: how to build and test locally, `npm run generate:catalog` usage and when to run it, release checklist, catalog structure conventions, testing against a temp dir, and other commands/processes a contributor needs to know. Currently this info is scattered across CLAUDE.md and tribal knowledge. *(added 2026-04-17, source: user)*
- **[feature] Routine report template** — Add a `routine-report-template.md` to `station/Reports/` alongside the existing `report-template.md`. Routine reports have a different shape than plan completion reports. The template in `loop.md` defines the format; this makes it a first-class project artifact. *(added 2026-04-14, source: user)*
- **[improvement] Split design-guide: generic catalog skill + Bonsai-specific station override** — Plan 11 replaces `catalog/skills/design-guide` with Bonsai-specific TUI/CLI rules (paths like `internal/tui/**`, palette tokens). That's useful for dogfooding but irrelevant when external users install the skill. Follow-up: restore a generic Charm/Go CLI design-guide to the catalog (palette patterns, NO_COLOR support, panel vocabulary as principles — not specific paths), and move Bonsai's own rules into `station/agent/Skills/design-guide.md` as a local override. *(added 2026-04-17, source: plan-11 audit)*
- **[improvement] Install semgrep for better SAST scanning** — `gitleaks` shipped (closed half of this item 2026-05-04 routine-digest); `semgrep` still pending. Vulnerability scan SAST currently falls back to Grep patterns — semgrep restores AST-aware OWASP-grade coverage. *(added 2026-04-16, narrowed 2026-05-04, source: routine-digest)*
<!-- "Bump golang.org/x/net v0.38 → v0.53" — resolved 2026-05-04 via Plan 36 / PR #94. -->
<!-- "Bump Go toolchain 1.25.8 → 1.25.9" — resolved 2026-05-04 via Plan 36 / PR #94. -->
<!-- "Plan 36 docs sweep ST-1+ST-2+PW-1" — resolved 2026-05-04 via Plan 36 / PR #94 (root CLAUDE.md tree, INDEX.md arch diagram, code-index.md sweep). -->
## P3 — Ideas & Research

### Validate command followups

- **[debt] `bonsai validate` — flag ownerless stale lock entries** — `internal/validate/validate.go` `auditStaleLockEntries` filters lock entries by per-agent workspace prefix to avoid double-reporting. Lock entries whose path lies outside ALL installed agents' workspaces are silently skipped — would surface only if an agent was uninstalled and lock entries lingered. Add a final post-loop pass that flags such entries with `AgentName=""`. *(added 2026-05-04, source: PR #93 review nit)*

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
<!-- "Add root Bonsai/CLAUDE.md check to doc-freshness-check routine" — promoted to P2 on 2026-05-04 after 3rd-cycle recurrence. See P2 Group B. -->
- **[improvement] Reduce `npm audit` cadence in dependency-audit routine** — `website/` npm audit has returned 0 vulnerabilities for multiple consecutive 7-day scans. Consider adding every-other-run skip logic (track last-npm-audit date in routine state) to save scan time. Revisit if a vulnerability surfaces. *(added 2026-04-21, source: routine-digest)*

### Outreach

- **[feature] README case study / blog post from session-transcript metrics** — The 2026-04-16 transcript analysis (archived at `station/Reports/Archive/2026-04-16-session-transcript-analysis.md`) contains compelling quantitative data: 20 sessions over 6 days, ~1,186 user messages (~90% silent tool approvals), ~2,000 substantive words total drove an entire CLI tool from Go rewrite through OSS release. Specific hooks: "75-message silent approval streak," "10+ deliverables in 48 minutes," "14 of 20 sessions had zero user-initiated rework." Could seed a README "Real-World Usage" section, a standalone case study page on the docs site, or a blog post. Parts 5 (session typology) and 7 (metrics) are the primary source material. *(added 2026-04-17, source: session — pending-report review)*

### Research

- **[research] Session-start payload — further optimization** — Plan 21 (2026-04-21, PR #46) cut ~10% (34.3KB→30.9KB) by removing redundant protocol dumps, end-of-session misfire, empty FieldNotes, and Reports full-cat. Remaining ~30KB is mostly identity.md + memory.md + self-awareness.md + INDEX.md + Status.md full-dump on every SessionStart. Ideas: (a) diff-based injection (only dump sections that changed since last session), (b) summary + link pattern (first 10 lines + "read full via tool if needed"), (c) conditional injection based on session type (fresh vs resume). Would need a sensor-side state cache. *(added 2026-04-15, updated 2026-04-21 post-Plan-21, source: architectural audit + Plan 21 findings)*
- **[research] Plan 08 C3 — prompt hook intent classification** — Deferred in original Plan 08 verification because auto-invocation via `.claude/skills/` (Phase A) + context-guard phrase regex (Phase C2 shipped 2026-04-21) were expected to cover the same workflows. Revisit when we have signal that the 3 target workflows (code-review, pr-review, security-audit) are NOT reliably auto-invoked. Trigger: user reports missed activation, or telemetry shows skill-description fuzzy-match misses. Cost: ~$0.001/prompt Haiku classification. *(added 2026-04-21, source: Plan 08 Phase C closeout)*
- **[research] Parallel agent coordination in shared repos** — Research how multiple code agents can work simultaneously on different tasks in the same repository. Key questions: git workflow, file contention, lock/claim protocol, orchestration model, state coherence, tooling. *(added 2026-04-16, source: user)*
- **[research] Archon analysis** — <https://github.com/coleam00/Archon> — research what it does, use cases, overlap with Bonsai, what we can learn. *(added 2026-04-13, source: user)*
- **[debt] Batch refresh outdated Go modules after toolchain upgrade** — 23 modules behind per `go list -m -u all` 2026-05-04 (was 17 on 2026-04-21): `golang.org/x/crypto v0.36 → v0.50`, `x/tools v0.37 → v0.44`, `x/sys v0.38 → v0.43`, `x/text v0.30 → v0.36`, `x/mod v0.28 → v0.35`, `x/sync v0.17 → v0.20`, `chroma/v2 v2.20 → v2.24`, `goldmark v1.7.13 → v1.8.2`, `go-udiff v0.3.1 → v0.4.1`, `regexp2 v1.11.5 → v1.12.0`, `pflag v1.0.9 → v1.0.10`, plus charm `x/exp/*` pseudo-versions. No CVEs beyond govulncheck. Hygiene sweep after Plan 36 Go toolchain bump lands. *(added 2026-04-21, updated 2026-05-04, source: routine-digest)*
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
