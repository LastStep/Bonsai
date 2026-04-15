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

## P1 — High

<!-- "Better trigger sections" promoted to Status.md Pending — removed 2026-04-14, backlog-hygiene routine -->
<!-- "Selective file update" implemented — removed 2026-04-15, issue-to-implementation workflow -->
<!-- "Doubled path prefix" fixed — removed 2026-04-15, issue-to-implementation workflow -->
<!-- "Workspace artifact sync" — marker migration implemented in PR #1, removed 2026-04-15, issue-to-implementation workflow -->

## P2 — Medium

<!-- "Rename catalog items to abilities" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #2 -->
- **[feature] Integration scaffolding variants** — Support alternative backends for all PM artifacts (backlog, status, roadmap, reports). During `bonsai init`, user picks a backend per artifact: markdown (default), GitHub Issues, Notion, Jira, etc. Generates variant-specific agent instructions (e.g., "use `gh issue create`" instead of "append to Backlog.md"). Each variant is a scaffolding template set — no runtime sync, pure swap at init time. Affects: scaffolding manifest, agent instructions, protocols, any sensor/workflow that references PM files. *Dependency: foundational — session-start auto-detect and routine GitHub issues build on this.* *(added 2026-04-15, source: user)*
- **[feature] Enhanced session-start sensor — project pulse** — Expand `session-context.sh` to inject a project-state summary at session start: backlog item count & top priorities, active status items, roadmap progress, and open items from external sources. Phase 1: check markdown PM files directly (no dependencies). Phase 2: auto-detect external sources from chosen integration variant in `.bonsai.yaml` (e.g., if backlog uses GitHub Issues, check open issues). *Dependency: full auto-detect requires integration scaffolding variants.* *(added 2026-04-15, source: user)*
- **[feature] Scheduled task generation for routines** — Auto-generate Claude scheduled task configs from routine metadata. Maps routine `frequency` to cron expressions. Two tiers: file-only routines (doc-freshness, memory-consolidation) → Desktop local tasks; heavier routines (vulnerability-scan, dependency-audit) → Cloud routines. Generated during `bonsai init`/`bonsai add`. Desktop tasks need the app running; Cloud routines run unattended on fresh clones (min 1hr interval). Add tier hint to routine `meta.yaml` (e.g., `schedule_tier: local|cloud`). *Note: Cloud routines are in research preview — API/limits may change.* *(added 2026-04-15, source: user)*
- **[feature] Routine GitHub issue creation** — Routines can create GitHub issues for actionable findings instead of (or alongside) reports. Per-routine opt-in via `creates_issues: true` in `meta.yaml`. Issues get a `bonsai-routine` label + routine name label. Issue body includes finding details, severity, suggested fix. Requires `gh` CLI. Best candidates: dependency-audit, vulnerability-scan, backlog-hygiene. *Dependency: prerequisite for auto-fixer routine (P3).* *(added 2026-04-15, source: user)*
<!-- "UI overhaul" promoted to Status.md Pending — removed 2026-04-14, backlog-hygiene routine -->
- **[feature] Usage instructions** — Add guidance on best practices: running routines at first start to fix references, how to trigger specific effects, proper explanation of each catalog item. *(added 2026-04-13, source: user)*
<!-- "Custom item detection" completed and shipped as bonsai update — removed 2026-04-14, backlog-hygiene routine -->
- **[improvement] Self-update mechanism** — Skills and workflows should be able to self-heal or flag when they have issues. *(added 2026-04-13, source: user)*
<!-- "bonsai guide command" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #3 -->
- **[feature] Custom item creator** — Interactive TUI for creating custom items (skill, workflow, protocol, sensor, routine) with frontmatter scaffolding — similar to Claude's skill creator. *(added 2026-04-14, source: user)*
- **[improvement] Catalog display_name audit** — Add explicit `display_name` to all catalog `meta.yaml` files. Research other metadata fields that could be useful (e.g., `version`, `tags`, `dependencies`, `examples`). *(added 2026-04-14, source: user)*
- **[feature] Routine report template** — Add a `routine-report-template.md` to `station/Reports/` alongside the existing `report-template.md`. Routine reports have a different shape than plan completion reports — they need execution metadata (duration, files read/modified, errors), step-by-step procedure walkthrough, findings summary table, and notes for next run. The template in `loop.md` defines the format; this makes it a first-class project artifact that subagents and manual runs can both reference. *(added 2026-04-14, source: user)*
- **[feature] Routine report digest routine** — New routine that scans `Reports/Pending/` for routine reports, extracts all actionable items (flagged for user review, errors, persistent findings) and notable observations across reports, and presents a consolidated digest to the user. Clears the signal-to-noise problem of having 8 individual reports — the human reads one digest instead of eight files. Should move processed reports from `Pending/` to an archive after digesting. *(added 2026-04-14, source: user)*

## P3 — Ideas & Research

- **[feature] Auto-fixer routine** — New routine that polls GitHub issues labeled `bonsai-routine`, attempts autonomous fixes. If fixable: creates a branch + PR with the fix, links the original issue. If human judgment needed: creates a new issue labeled `human-needed` with context on what's blocked and why, links the original. Forms a closed loop with routine GitHub issue creation (P2). Good candidate for Cloud routine scheduling. *Depends on: routine GitHub issue creation + scheduled task generation.* *(added 2026-04-15, source: user)*
- **[research] Archon analysis** — https://github.com/coleam00/Archon — research what it does, use cases, overlap with Bonsai, what we can learn. *(added 2026-04-13, source: user)*
- **[feature] Managed Agents integration** — Cloud deployment via `bonsai deploy`, session management, outcome rubrics in catalog. Build after local foundation is stable. *(added 2026-04-13, source: user)*
- **[feature] Greenhouse companion app** — Desktop app for managing projects + observing AI agents. Design doc: DESIGN-companion-app.md. Stack: Tauri v2 + Svelte 5 + SQLite. Status: Design phase, decisions locked. *(added 2026-04-13, source: user)*
