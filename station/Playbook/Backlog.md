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

```
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

---

## P0 — Critical

<!-- "Case-insensitive file collision" fixed — removed 2026-04-16, issue-to-implementation workflow, PR #8 -->

## P1 — High

- **[debt] Testing infrastructure for triggers and sensors** — No testing infrastructure exists for hook-based triggers, prompt hooks, context-guard regex patterns, path-scoped rules, or skill auto-invocation. Need: (1) unit tests for context-guard regex patterns (positive/negative cases), (2) integration test harness for sensor scripts (mock stdin, verify stdout/exit codes), (3) end-to-end test framework for trigger activation (simulate user prompts, verify correct ability loads), (4) prompt hook evaluation testing (verify Haiku correctly classifies intents). The trigger system is expanding significantly — without test infra, regressions will be invisible. *(added 2026-04-16, source: user)*
- **[debt] Code index line number drift + missing entries** — `station/code-index.md` is missing entries for `bonsai update`, `bonsai guide`, `frontmatter.go`, `scan.go`. 11 line number references for `internal/generate/generate.go` functions have drifted by +6 to +102 lines. `descFor()` signature description is stale (missing `customItems` parameter). *(added 2026-04-16, source: routine-digest)*
- **[security] Upgrade Go toolchain from 1.24.3 to 1.24.13+** — 3 symbol-level stdlib vulnerabilities: GO-2025-3956 (os/exec LookPath, medium), GO-2025-3750 (O_CREATE|O_EXCL Windows-only, low), GO-2026-4602 (FileInfo Root escape, medium — requires go1.25.8). Upgrading to 1.24.13+ resolves the first two. Third to monitor until go1.25 is stable. *(added 2026-04-16, source: routine-digest)*
<!-- "CI workflow + branch protection" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #6 -->
<!-- "Release pipeline" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #5 -->
<!-- "Better trigger sections" promoted to Status.md Pending — removed 2026-04-14, backlog-hygiene routine -->
<!-- "Selective file update" implemented — removed 2026-04-15, issue-to-implementation workflow -->
<!-- "Doubled path prefix" fixed — removed 2026-04-15, issue-to-implementation workflow -->
<!-- "Workspace artifact sync" — marker migration implemented in PR #1, removed 2026-04-15, issue-to-implementation workflow -->

## P2 — Medium

- **[improvement] OSS polish — linter config, demo visual, Makefile targets** — Three remaining items from the OSS readiness audit: (1) add `.golangci.yml` with standard checks (unused code, formatting, error handling, shadowing) and wire into CI, (2) add a demo GIF/asciinema recording to README (placeholder comment exists at line 24), (3) add `test` and `lint` targets to the Makefile. *(added 2026-04-16, source: RESEARCH-oss-readiness.md cleanup)*
- **[feature] Unbuilt catalog items — 3 agents, 1 skill, 4 routines** — From the catalog expansion research, 8 items were never built: **agents** `qa`, `reviewer`, `docs`; **skill** `documentation-standards` (blocks `docs` agent); **routines** `test-coverage-check` (qa), `changelog-maintenance` (docs), `api-docs-drift` (docs), `standards-drift` (reviewer). Build order: `documentation-standards` skill first (unblocks `docs`), then agents, then routines. *(added 2026-04-16, source: RESEARCH-catalog-expansion.md cleanup)*
- **[feature] Research scaffolding item + abilities** — Add an optional `Research/` folder to project scaffolding (in `catalog/scaffolding/manifest.yaml`) for storing landscape analysis, concept decisions, and design research. Add associated abilities (tech-lead only): a research workflow (structured research process — question framing, landscape scan, comparison, synthesis, position-taking) and/or a research-template skill (format and sections for research docs). Only relevant for tech-lead agent — research is an architectural activity, not an implementation one. *(added 2026-04-16, source: user)*
- **[research] Revisit concept-decisions research** — Review `station/Research/concept-decisions.md` for unbuilt concepts that may be worth promoting: (1) **Talents** — a new catalog category for innate behavioral aptitudes (ambient, soft, closer to identity than protocols), (2) **Meta-layer** — runtime system-wide observation layer outside the agent paradigm, (3) **Three-layer catalog ownership model** — upstream/generated/customized distinction is undocumented outside this research, (4) **Loading gradient** — the reasoning behind Core > Protocols > Workflows > Skills ordering. Decide which to build, which to backlog properly, which to discard. *(added 2026-04-16, source: research doc cleanup)*
- **[improvement] Plan archiving — Active/Archive folder structure** — Plans currently all live in `Plans/Active/`. Completed plans should move to `Plans/Archive/` after merge. Requires: create `Plans/Archive/` in scaffolding manifest, update issue-to-implementation workflow (Phase 10 logging should move plan to Archive), update planning workflow and planning-template skill to reference the folder structure, update session-start protocol if it scans for active plans, update any sensors or routines that reference plan paths. Also update the station CLAUDE.md nav table to show both folders. *(added 2026-04-16, source: user)*
- **[improvement] Consolidate FieldNotes usage** — The current `Logs/FieldNotes.md` file has unclear purpose and overlaps with other state files (memory.md, Status.md, KeyDecisionLog.md). Rethink what it's for, whether it should be merged into another artifact, and how it fits into the session-start context injection. *(added 2026-04-15, source: user)*
<!-- "Rename catalog items to abilities" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #2 -->
- **[feature] Integration scaffolding variants** — Support alternative backends for all PM artifacts (backlog, status, roadmap, reports). During `bonsai init`, user picks a backend per artifact: markdown (default), GitHub Issues, Notion, Jira, etc. Generates variant-specific agent instructions (e.g., "use `gh issue create`" instead of "append to Backlog.md"). Each variant is a scaffolding template set — no runtime sync, pure swap at init time. Affects: scaffolding manifest, agent instructions, protocols, any sensor/workflow that references PM files. *Dependency: foundational — session-start auto-detect and routine GitHub issues build on this.* *(added 2026-04-15, source: user)*
- **[feature] Enhanced session-start sensor — project pulse** — Expand `session-context.sh` to inject a project-state summary at session start: backlog item count & top priorities, active status items, roadmap progress, and open items from external sources. Phase 1: check markdown PM files directly (no dependencies). Phase 2: auto-detect external sources from chosen integration variant in `.bonsai.yaml` (e.g., if backlog uses GitHub Issues, check open issues). *Dependency: full auto-detect requires integration scaffolding variants.* *(added 2026-04-15, source: user)*
- **[feature] Scheduled task generation for routines** — Auto-generate Claude scheduled task configs from routine metadata. Maps routine `frequency` to cron expressions. Two tiers: file-only routines (doc-freshness, memory-consolidation) → Desktop local tasks; heavier routines (vulnerability-scan, dependency-audit) → Cloud routines. Generated during `bonsai init`/`bonsai add`. Desktop tasks need the app running; Cloud routines run unattended on fresh clones (min 1hr interval). Add tier hint to routine `meta.yaml` (e.g., `schedule_tier: local|cloud`). *Note: Cloud routines are in research preview — API/limits may change.* *(added 2026-04-15, source: user)*
- **[feature] Routine GitHub issue creation** — Routines can create GitHub issues for actionable findings instead of (or alongside) reports. Per-routine opt-in via `creates_issues: true` in `meta.yaml`. Issues get a `bonsai-routine` label + routine name label. Issue body includes finding details, severity, suggested fix. Requires `gh` CLI. Best candidates: dependency-audit, vulnerability-scan, backlog-hygiene. *Dependency: prerequisite for auto-fixer routine (P3).* *(added 2026-04-15, source: user)*
<!-- "UI overhaul" promoted to Status.md Pending — removed 2026-04-14, backlog-hygiene routine -->
<!-- "Usage instructions" partially implemented as AI operational intelligence (Plan 05, PR pending) — the AI-facing half (How to Work CLAUDE.md section + workspace-guide skill). Human-facing half split to separate item below. removed 2026-04-16, issue-to-implementation workflow -->
- **[feature] `bonsai guide` multi-topic command** — Expand `bonsai guide` from a single-doc renderer into a multi-topic CLI guide with an interactive Huh topic picker. Topics: **quickstart** (post-init walkthrough — what was generated, what to read first, how to add agents, when to run routines), **concepts** (mental model — station/workspace, agents, sensors, routines, dispatch, Playbook, how they connect), **catalog** (annotated catalog with per-item explanations — goes beyond `bonsai catalog` tables by explaining *why* and *when* to use each item), **custom-files** (existing `docs/custom-files.md` — moves to subtopic). Implementation: each topic is a separate `docs/{topic}.md` file; `cmd/guide.go` adds a Huh select form when called without args; `bonsai guide <topic>` skips the picker. The HANDBOOK.md content can be mined for the concepts and quickstart topics. Pairs with AI operational intelligence (Plan 05) to complete the full usage instructions feature. *(added 2026-04-16, source: plan-05 split)*
<!-- "Custom item detection" completed and shipped as bonsai update — removed 2026-04-14, backlog-hygiene routine -->
- **[improvement] Self-update mechanism** — Skills and workflows should be able to self-heal or flag when they have issues. *(added 2026-04-13, source: user)*
<!-- "bonsai guide command" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #3 -->
<!-- "Community health files" implemented — removed 2026-04-16, issue-to-implementation workflow, PR #9 -->
- **[feature] CLI usage guide** — Write `docs/cli-usage.md` covering every command in depth: `init` walkthrough (what each prompt means, scaffolding choices, agent defaults), `add` (component selection, compatibility filtering, what gets generated), `remove` (clean vs file-preserving removal), `update` (conflict resolution flow, custom file detection), `list` and `catalog` (reading the output). Include practical scenarios: first-time setup, adding a second agent, upgrading after a Bonsai version bump, recovering from a bad update. Link from README guides table and `bonsai guide` multi-topic command. *(added 2026-04-16, source: user)*
<!-- "Human-AI interaction guide" implemented as docs/working-with-agents.md — removed 2026-04-16, session work -->
- **[feature] Concepts guide** — Write `docs/concepts.md` explaining Bonsai's mental model for newcomers: station vs workspace, the 6-layer instruction stack, agents as team members, sensors as automated enforcement, routines as self-maintenance, the Playbook as project state, and how everything connects. Extract and reorganize content from HANDBOOK.md into a standalone conceptual overview aimed at someone evaluating whether to adopt Bonsai. *(added 2026-04-16, source: user)*
- **[feature] Quickstart guide** — Write `docs/quickstart.md` as a post-install walkthrough: what `bonsai init` generated, what to read first, how to add your first code agent, running your first session with the Tech Lead, understanding the generated CLAUDE.md, and when to run routines. Distinct from the README quick start (which is 2 commands) — this is the "now what?" guide for after installation. *(added 2026-04-16, source: user)*
- **[improvement] Remove infra-drift-check routine** — Bonsai has no cloud infrastructure, containers, or CI/CD pipelines. This routine will always find nothing. Remove from `.bonsai.yaml` and station workspace; re-add via `bonsai add` if infra is added later. *(added 2026-04-16, source: routine-digest)*
- **[improvement] Install semgrep and/or gitleaks for better scanning** — Vulnerability scan and secrets scan routines currently use manual pattern-based Grep scanning. Installing semgrep (SAST) and/or gitleaks (secrets) would improve coverage and reduce false negatives. *(added 2026-04-16, source: routine-digest)*
- **[improvement] Consolidate "Usage instructions" roadmap item with `bonsai guide`** — Roadmap Phase 1 has "Usage instructions" unchecked, while `bonsai guide` (P2 backlog) partially addresses it. Clarify whether `bonsai guide` multi-topic command fully covers this or if additional work is needed, then update Roadmap accordingly. *(added 2026-04-16, source: routine-digest)*
- **[feature] Custom item creator** — Interactive TUI for creating custom items (skill, workflow, protocol, sensor, routine) with frontmatter scaffolding — similar to Claude's skill creator. *(added 2026-04-14, source: user)*
- **[improvement] Catalog display_name audit** — Add explicit `display_name` to all catalog `meta.yaml` files. Research other metadata fields that could be useful (e.g., `version`, `tags`, `dependencies`, `examples`). *(added 2026-04-14, source: user)*
- **[feature] Routine report template** — Add a `routine-report-template.md` to `station/Reports/` alongside the existing `report-template.md`. Routine reports have a different shape than plan completion reports — they need execution metadata (duration, files read/modified, errors), step-by-step procedure walkthrough, findings summary table, and notes for next run. The template in `loop.md` defines the format; this makes it a first-class project artifact that subagents and manual runs can both reference. *(added 2026-04-14, source: user)*
<!-- "Routine report digest" implemented as custom workflow agent/Workflows/routine-digest.md — removed 2026-04-16, manual creation -->

## P3 — Ideas & Research

- **[improvement] Micro-task fast path** — Define an explicit lightweight protocol for trivial changes (< 50 LOC, no architectural impact). When a task is classified as micro, skip the planning pipeline and let the agent execute directly. Could be a sensor that auto-detects task weight, or a protocol clause agents check before entering full planning mode. Related to "Better trigger sections" (P1) but distinct — trigger sections control *what activates*; fast path controls *how much ceremony*. *(added 2026-04-15, source: architectural audit)*
- **[research] Session-start payload optimization** — Investigate whether the session-context sensor payload can be made leaner: strip redundant markdown formatting, collapse whitespace, or pre-render a minified version. Current payload is ~600-700 lines — not critical, but could free ~200-300 tokens with formatting cleanup. Low priority since the layered loading already defers most content. *(added 2026-04-15, source: architectural audit)*
- **[feature] Auto-fixer routine** — New routine that polls GitHub issues labeled `bonsai-routine`, attempts autonomous fixes. If fixable: creates a branch + PR with the fix, links the original issue. If human judgment needed: creates a new issue labeled `human-needed` with context on what's blocked and why, links the original. Forms a closed loop with routine GitHub issue creation (P2). Good candidate for Cloud routine scheduling. *Depends on: routine GitHub issue creation + scheduled task generation.* *(added 2026-04-15, source: user)*
- **[research] Parallel agent coordination in shared repos** — Research how multiple code agents can work simultaneously on different tasks in the same repository. Key questions: (1) git workflow — worktrees, feature branches, merge conflict resolution when agents finish concurrently, (2) file contention — what happens when two agents need to touch overlapping files (shared config, package.json, migration files), (3) lock/claim protocol — should agents declare intent before starting, or resolve conflicts at merge time, (4) orchestration model — does the tech lead dispatch and wait serially, or dispatch N agents in parallel and review as they finish, (5) state coherence — how do Status.md, backlog, and memory stay consistent when multiple agents update them, (6) tooling — do we need a coordinator process, or can convention-based rules (worktree isolation + merge protocol) handle it. Look at how GSD handles wave-based parallelization and how real engineering teams handle concurrent feature work. This is foundational for scaling beyond single-agent-at-a-time execution. *(added 2026-04-16, source: user)*
- **[research] Archon analysis** — https://github.com/coleam00/Archon — research what it does, use cases, overlap with Bonsai, what we can learn. *(added 2026-04-13, source: user)*
- **[feature] Managed Agents integration** — Cloud deployment via `bonsai deploy`, session management, outcome rubrics in catalog. Build after local foundation is stable. *(added 2026-04-13, source: user)*
- **[feature] Greenhouse companion app** — Desktop app for managing projects + observing AI agents. Design doc: DESIGN-companion-app.md. Stack: Tauri v2 + Svelte 5 + SQLite. Status: Design phase, decisions locked. *(added 2026-04-13, source: user)*
