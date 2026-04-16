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

- **[bug] triggerSection() prepends before YAML frontmatter** `[Group B]` — `triggerSection()` in `internal/generate/generate.go` blindly prepends the `## Triggers` block to file content. When the source file has YAML frontmatter (`---` block), the triggers end up above it, breaking metadata parsing. Fix: detect existing frontmatter and insert the trigger section after the closing `---` instead of before the full content. Affects all skills and workflows with frontmatter + triggers. Quick fix (~10 lines). *(added 2026-04-16, source: session — found during `bonsai update` dogfooding)*
- **[bug] Silent error swallowing in spinner callbacks** `[Group B]` — 41 errors are discarded in production code because the Huh spinner callback signature is `func()` — it can't return errors. Generation failures during `bonsai init/add/remove/update` are invisible to users. A corrupted template, missing file, or permission error succeeds silently and leaves the workspace broken. Fix: collect errors inside the callback via closure (e.g., append to `[]error`), then check after the spinner completes. Affects: `cmd/add.go` (10), `cmd/init.go` (5), `cmd/update.go` (5), `cmd/remove.go` (14), `internal/generate/generate.go` (6), `cmd/root.go` (1). *(added 2026-04-16, source: repo-analytics)*
- **[security] Upgrade Go toolchain from 1.24.3 to 1.24.13+** — 3 symbol-level stdlib vulnerabilities: GO-2025-3956 (os/exec LookPath, medium), GO-2025-3750 (O_CREATE|O_EXCL Windows-only, low), GO-2026-4602 (FileInfo Root escape, medium — requires go1.25.8). Upgrading to 1.24.13+ resolves the first two. Third to monitor until go1.25 is stable. *(added 2026-04-16, source: routine-digest)*
- **[debt] Testing infrastructure for triggers and sensors** `[Group B]` — No testing infrastructure exists for hook-based triggers, prompt hooks, context-guard regex patterns, path-scoped rules, or skill auto-invocation. Need: (1) unit tests for context-guard regex patterns (positive/negative cases), (2) integration test harness for sensor scripts (mock stdin, verify stdout/exit codes), (3) end-to-end test framework for trigger activation (simulate user prompts, verify correct ability loads), (4) prompt hook evaluation testing (verify Haiku correctly classifies intents). The trigger system is expanding significantly — without test infra, regressions will be invisible. *(added 2026-04-16, source: user)*

## P2 — Medium

### Group A: Documentation Suite

> Resolves Roadmap Phase 1 "Usage instructions". The multi-topic command is the delivery mechanism; the three docs are the content. Ship incrementally — content first, CLI wiring last.

- **[feature] Quickstart guide** — Write `docs/quickstart.md` as a post-install walkthrough: what `bonsai init` generated, what to read first, how to add your first code agent, running your first session with the Tech Lead, understanding the generated CLAUDE.md, and when to run routines. Distinct from the README quick start (which is 2 commands) — this is the "now what?" guide for after installation. *(added 2026-04-16, source: user)*
- **[feature] Concepts guide** — Write `docs/concepts.md` explaining Bonsai's mental model for newcomers: station vs workspace, the 6-layer instruction stack, agents as team members, sensors as automated enforcement, routines as self-maintenance, the Playbook as project state, and how everything connects. Extract and reorganize content from HANDBOOK.md into a standalone conceptual overview aimed at someone evaluating whether to adopt Bonsai. *(added 2026-04-16, source: user)*
- **[feature] CLI usage guide** — Write `docs/cli-usage.md` covering every command in depth: `init` walkthrough (what each prompt means, scaffolding choices, agent defaults), `add` (component selection, compatibility filtering, what gets generated), `remove` (clean vs file-preserving removal), `update` (conflict resolution flow, custom file detection), `list` and `catalog` (reading the output). Include practical scenarios: first-time setup, adding a second agent, upgrading after a Bonsai version bump, recovering from a bad update. Link from README guides table and `bonsai guide` multi-topic command. *(added 2026-04-16, source: user)*
- **[feature] `bonsai guide` multi-topic command** — Expand `bonsai guide` from a single-doc renderer into a multi-topic CLI guide with an interactive Huh topic picker. Topics: **quickstart**, **concepts**, **catalog**, **custom-files** (existing). Implementation: each topic is a separate `docs/{topic}.md` file; `cmd/guide.go` adds a Huh select form when called without args; `bonsai guide <topic>` skips the picker. *(added 2026-04-16, source: plan-05 split)*

### Group B: Code Quality & Testing

> Logical ordering: split the big file first (makes testing easier), then add tests, then fix error handling. The two P1 bugs (triggerSection frontmatter, spinner error swallowing) can be fixed independently at any time.

- **[debt] Break up `generate.go` — 1,357 lines, highest churn file** — `internal/generate/generate.go` is both the largest Go file and the most frequently modified. It handles file writing, template rendering, conflict resolution, lock management, sensor/routine wiring, and scaffolding — too many responsibilities in one file. Split along natural seams: (1) template rendering, (2) file writing + conflict resolution, (3) lock management, (4) sensor/routine wiring. Would improve testability and reduce merge friction for agent dispatches. *(added 2026-04-16, source: repo-analytics)*
- **[debt] `internal/catalog/` test coverage — 496 lines, 0%** — Catalog loading (`LoadCatalog()`, `DisplayNameFrom()`, meta.yaml parsing) is the bridge between embedded YAML and the rest of the system. A malformed `meta.yaml` in the catalog would break at runtime with no test to catch it. Basic tests for catalog loading, display name derivation, and agent compatibility filtering would catch regressions cheaply. *(added 2026-04-16, source: repo-analytics)*
- **[debt] CLI command test coverage — `cmd/` package at 0%** — The `cmd/` package contains all user-facing CLI logic (init, add, remove, update, list, catalog, guide) — 1,691 lines across 8 files, zero tests. Priority targets: (1) `cmd/init.go` — happy path e2e test (temp dir, verify output structure), (2) `cmd/add.go` — test that abilities land correctly, (3) `cmd/remove.go` — test clean removal (472 lines, 4th largest file). Table-driven tests with temp dir setup would cover the most ground. *(added 2026-04-16, source: repo-analytics)*

### Group C: OSS Readiness

> All support the public repo being contributor-friendly. Small, independent — could knock both out in one session.

- **[improvement] OSS polish — linter config, demo visual, Makefile targets** — Three remaining items from the OSS readiness audit: (1) add `.golangci.yml` with standard checks (unused code, formatting, error handling, shadowing) and wire into CI, (2) add a demo GIF/asciinema recording to README (placeholder comment exists at line 24), (3) add `test` and `lint` targets to the Makefile. *(added 2026-04-16, source: RESEARCH-oss-readiness.md cleanup)*
- **[improvement] Seed GitHub Issues for contributor on-ramp** — The repo is public with community health files and a polished README, but has 0 issues, 0 stars, 0 forks — no entry point for potential contributors. File 3-5 well-scoped issues labeled `good first issue` (e.g., add `test`/`lint` Makefile targets, add `.golangci.yml`, add catalog tests). Also consider adding a `help wanted` label for medium-complexity items. *(added 2026-04-16, source: repo-analytics)*

### Group D: Catalog Expansion

> Research first (concept-decisions), then build. The concept-decisions review informs which of the others to prioritize.

- **[research] Revisit concept-decisions research** — Review `station/Research/concept-decisions.md` for unbuilt concepts that may be worth promoting: (1) **Talents** — a new catalog category for innate behavioral aptitudes, (2) **Meta-layer** — runtime system-wide observation layer, (3) **Three-layer catalog ownership model**, (4) **Loading gradient** reasoning. Decide which to build, which to backlog properly, which to discard. *(added 2026-04-16, source: research doc cleanup)*
- **[feature] Unbuilt catalog items — 3 agents, 1 skill, 4 routines** — From the catalog expansion research, 8 items were never built: **agents** `qa`, `reviewer`, `docs`; **skill** `documentation-standards` (blocks `docs` agent); **routines** `test-coverage-check` (qa), `changelog-maintenance` (docs), `api-docs-drift` (docs), `standards-drift` (reviewer). Build order: `documentation-standards` skill first (unblocks `docs`), then agents, then routines. *(added 2026-04-16, source: RESEARCH-catalog-expansion.md cleanup)*
- **[feature] Changelog generation skill + release changelogs** — Add a changelog generation skill that: (1) parses conventional commit messages between tags to generate structured changelogs, (2) outputs CHANGELOG.md, (3) generates release notes for `gh release create --notes`. Current releases (v0.1.0-v0.1.3) shipped with no changelogs — backfill them. Also consider a `bonsai changelog` CLI command. *(added 2026-04-16, source: user)*
- **[feature] Research scaffolding item + abilities** — Add an optional `Research/` folder to project scaffolding for storing landscape analysis, concept decisions, and design research. Add associated abilities (tech-lead only): a research workflow and/or a research-template skill. *(added 2026-04-16, source: user)*

### Group E: Workspace Improvements

> Small, independent quality-of-life items. Can be done in any order.

- **[improvement] Plan archiving — Active/Archive folder structure** — Plans currently all live in `Plans/Active/`. Completed plans should move to `Plans/Archive/` after merge. Requires: create `Plans/Archive/` in scaffolding manifest, update issue-to-implementation workflow (Phase 10), update planning workflow and planning-template skill, update session-start protocol if it scans for active plans, update CLAUDE.md nav table. *(added 2026-04-16, source: user)*
- **[improvement] Consolidate FieldNotes usage** — The current `Logs/FieldNotes.md` file has unclear purpose and overlaps with other state files (memory.md, Status.md, KeyDecisionLog.md). Rethink what it's for, whether it should be merged into another artifact, and how it fits into the session-start context injection. *(added 2026-04-15, source: user)*
- **[improvement] Post-update backup merge hint** — After `bonsai update` creates `.bak` backups during conflict resolution, print a hint telling the user to ask their agent to reconcile customizations. Small change to `cmd/update.go` after `resolveConflicts()` returns. *(added 2026-04-16, source: user)*

### Ungrouped P2

- **[feature] Routine report template** — Add a `routine-report-template.md` to `station/Reports/` alongside the existing `report-template.md`. Routine reports have a different shape than plan completion reports. The template in `loop.md` defines the format; this makes it a first-class project artifact. *(added 2026-04-14, source: user)*
- **[improvement] Install semgrep and/or gitleaks for better scanning** — Vulnerability scan and secrets scan routines currently use manual pattern-based Grep scanning. Installing semgrep (SAST) and/or gitleaks (secrets) would improve coverage and reduce false negatives. *(added 2026-04-16, source: routine-digest)*

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

### Research

- **[research] Session-start payload optimization** — Investigate whether the session-context sensor payload can be made leaner. Current payload is ~600-700 lines — could free ~200-300 tokens with formatting cleanup. Low priority since layered loading already defers most content. *(added 2026-04-15, source: architectural audit)*
- **[research] Parallel agent coordination in shared repos** — Research how multiple code agents can work simultaneously on different tasks in the same repository. Key questions: git workflow, file contention, lock/claim protocol, orchestration model, state coherence, tooling. *(added 2026-04-16, source: user)*
- **[research] Archon analysis** — <https://github.com/coleam00/Archon> — research what it does, use cases, overlap with Bonsai, what we can learn. *(added 2026-04-13, source: user)*

### Big Bets

- **[feature] Managed Agents integration** — Cloud deployment via `bonsai deploy`, session management, outcome rubrics in catalog. Build after local foundation is stable. *(added 2026-04-13, source: user)*
- **[feature] Greenhouse companion app** — Desktop app for managing projects + observing AI agents. Design doc: DESIGN-companion-app.md. Stack: Tauri v2 + Svelte 5 + SQLite. Status: Design phase, decisions locked. *(added 2026-04-13, source: user)*
- **[improvement] Catalog display_name audit** — Add explicit `display_name` to all catalog `meta.yaml` files. Research other metadata fields that could be useful (e.g., `version`, `tags`, `dependencies`, `examples`). *(added 2026-04-14, source: user)*

---

## Removed Items

<!-- Items resolved or removed during backlog hygiene. Keep for audit trail. -->
<!-- "Case-insensitive file collision" fixed — removed 2026-04-16, issue-to-implementation workflow, PR #8 -->
<!-- "Code index line number drift" fixed — removed 2026-04-16, issue-to-implementation workflow, PR #12 -->
<!-- "CI workflow + branch protection" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #6 -->
<!-- "Release pipeline" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #5 -->
<!-- "Better trigger sections" promoted to Status.md Pending — removed 2026-04-14, backlog-hygiene routine -->
<!-- "Selective file update" implemented — removed 2026-04-15, issue-to-implementation workflow -->
<!-- "Doubled path prefix" fixed — removed 2026-04-15, issue-to-implementation workflow -->
<!-- "Workspace artifact sync" — marker migration implemented in PR #1, removed 2026-04-15, issue-to-implementation workflow -->
<!-- "Rename catalog items to abilities" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #2 -->
<!-- "Custom item detection" completed and shipped as bonsai update — removed 2026-04-14, backlog-hygiene routine -->
<!-- "UI overhaul" promoted to Status.md Pending — removed 2026-04-14, backlog-hygiene routine -->
<!-- "Usage instructions" partially implemented as AI operational intelligence (Plan 05, PR pending) — split to guide items in Group A -->
<!-- "Human-AI interaction guide" implemented as docs/working-with-agents.md — removed 2026-04-16, session work -->
<!-- "bonsai guide command" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #3 -->
<!-- "Community health files" implemented — removed 2026-04-16, issue-to-implementation workflow, PR #9 -->
<!-- "Routine report digest" implemented as custom workflow agent/Workflows/routine-digest.md — removed 2026-04-16, manual creation -->
<!-- "Clean up stale remote branches" — done 2026-04-16, backlog hygiene (deleted 8 merged branches, enabled prune) -->
<!-- "Remove infra-drift-check routine" — done 2026-04-16, backlog hygiene (no cloud infra to check) -->
<!-- "Consolidate Usage instructions roadmap item" — resolved by grouping guide items into Group A, 2026-04-16 -->

---

## Group Index

| Group | Theme | Phase Order | Notes |
|-------|-------|-------------|-------|
| **A** | Documentation Suite | Quickstart → Concepts → CLI Usage → Multi-topic command | Resolves Roadmap "Usage instructions". Content first, CLI wiring last. |
| **B** | Code Quality & Testing | Split generate.go → catalog tests → cmd tests → trigger test infra → spinner error fix | P1 bugs (frontmatter, spinners) can be fixed independently at any time. |
| **C** | OSS Readiness | Linter + Makefile → seed GitHub Issues | Small, one-session effort. |
| **D** | Catalog Expansion | Concept-decisions review → documentation-standards skill → agents → routines → changelog | Research informs build order. |
| **E** | Workspace Improvements | Any order | Independent quality-of-life items. |
