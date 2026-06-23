---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-23
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~10 minutes
- **Files Read:** 12 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/CLAUDE.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`, `station/agent/Workflows/plan-grilling.md`, `station/agent/Skills/critic-agent-prompts.md`, `station/Logs/RoutineLog.md`, `CHANGELOG.md`, `go.mod`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (routine log entry)
- **Tools Used:** Read, Bash (git log, ls, grep, file existence checks)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation / compare against recent git history
Ran `git log --since="2026-05-04"` to identify all commits since last run (50 commits). Key changes that may affect documentation:

- **Plan 41 (2026-06-16):** Headless CLI contract — added `internal/nonint/` package, `docs/agent-interface.md`, headless `*Result` cores for all mutating commands, `list --json`, `ExitConflict=5`.
- **Plan 40 (2026-06-13):** In-repo memory graph scaffolding, `bonsai validate` project-level pass, `bonsai guide formats`.
- **plan-grilling pipeline added (2026-06-13):** 6-critic adversarial plan review workflow and critic agent prompts skill added directly to `station/agent/`.
- **completion command shipped (PR #78, 2026-05-07):** `bonsai completion [bash|zsh|fish|powershell]` added.
- **Extensive website documentation** added (`website/src/content/docs/`) — outside the `station/` scope for this routine.

### Step 2 — Check INDEX.md accuracy
Verified `station/INDEX.md` tech stack and project description. Found two stale items:

1. **CLI commands count:** INDEX.md states `8 (init, add, remove, list, catalog, update, guide, validate)` but the actual count is **9** — `bonsai completion` was shipped in PR #78 (2026-05-07) and is not listed.
2. **Architecture overview omits `internal/nonint/`:** The architecture block shows `cmd/ → internal/catalog/, internal/config/, internal/generate/, internal/validate/, internal/wsvalidate/, internal/tui/` but the `internal/nonint/` package (the headless exit-code contract layer, Plan 41) is not represented.

Agent types (6), catalog items (~50 — actual count 53), and tech stack are accurate.

### Step 3 — Check navigation links
Verified all links in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References). All linked files exist on disk — **no broken links**.

However, found **two files that exist but are not in the navigation table**:

1. `station/agent/Workflows/plan-grilling.md` — added 2026-06-13, not listed in the Workflows navigation table in `station/CLAUDE.md`. The file has trigger phrases documented ("grill the plan", "review plan NN", "critic pass", "team of agents review this") but no navigation entry to guide the agent to load it.
2. `station/agent/Skills/critic-agent-prompts.md` — added 2026-06-13 (companion to plan-grilling), not listed in the Skills navigation table in `station/CLAUDE.md`. Both files note "full Bonsai-catalog integration pending (Backlog)" but they are already in use in the workspace.

Also verified `station/code-index.md`: **`internal/nonint/` package is missing** — Plan 41 added a substantial new package (`runner.go`, `events.go`, `result.go`, `config.go`, `remove.go`, `update.go`, `nonint.go`) that defines the canonical headless exit-code contract (`ExitConflict=5`, `Exit*` constants). Not documented in code-index.md.

### Step 4 — Report findings
Findings documented below. Per procedure, flagging for user decision — not executing updates.

### Step 5 — Update dashboard
Dashboard row updated: Last Ran → 2026-06-23, Next Due → 2026-06-30, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `bonsai completion` command missing from CLI commands count and architecture overview — INDEX.md says 8 commands, actual is 9 (completion shipped PR #78 2026-05-07) | `station/INDEX.md` lines 33 and 63 | Flagged for user |
| 2 | Medium | `internal/nonint/` package entirely absent from code index — Plan 41's headless contract layer (runner.go, exit codes, events, result shapes) not documented | `station/code-index.md` | Flagged for user |
| 3 | Low | `agent/Workflows/plan-grilling.md` exists but has no entry in `station/CLAUDE.md` Workflows navigation table | `station/CLAUDE.md` Workflows section | Flagged for user |
| 4 | Low | `agent/Skills/critic-agent-prompts.md` exists but has no entry in `station/CLAUDE.md` Skills navigation table | `station/CLAUDE.md` Skills section | Flagged for user |
| 5 | Low | Plans 40 and 41 remain in `Plans/Active/` — both appear only in Recently Done (status-hygiene ran today also flagged this; memory.md has an explicit note to archive Plan 41) | `station/Playbook/Plans/Active/` | Flagged for user (also flagged by status-hygiene today) |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

1. **INDEX.md: completion command** — Update line 33 count `8 → 9` and add `completion` to the command list; update line 63 architecture block to include `bonsai completion`. (Low effort, clear fix.)

2. **code-index.md: `internal/nonint/` section** — Add a section documenting the package's key types and functions: `NonInteractiveConfig`, `Runner`, `ExitConflict` constant, `Result` shape, `Events` shape, `Update*` and `Remove*` headless functions. This is the canonical reference for AI integrators (per `docs/agent-interface.md`). (Moderate effort — requires reading the package to build accurate entries.)

3. **station/CLAUDE.md: plan-grilling and critic-agent-prompts nav entries** — Add a row for `plan-grilling.md` to the Workflows table (trigger: "grill the plan", "review plan NN", "critic pass") and a row for `critic-agent-prompts.md` to the Skills table (activate: when running plan-grilling, loading critic prompts). Note: both files have a comment saying "full Bonsai-catalog integration pending (Backlog)" — user may want to add them to `.bonsai.yaml` as custom items and run `bonsai update` to regenerate CLAUDE.md, rather than hand-editing.

4. **Plans/Active cleanup** — Archive Plans 40 and 41 to `Plans/Archive/`. This is a session-wrapup task already flagged in memory.md and again by the status-hygiene routine today.

## Notes for Next Run
- The gap since last run was 50 days (2026-05-04 → 2026-06-23) covering 50 commits and two major feature plans (40 and 41). Findings were moderate — mainly missing code-index entries for the new nonint package and two undocumented custom workflow/skill files. No broken links found.
- Website documentation (`website/src/content/docs/`) saw extensive additions in this period — if a website-accuracy check is desired, that's a separate scope (not covered by this routine's `station/` focus).
- The `infra-drift-check` routine exists in `catalog/routines/` but is not installed in the station workspace — not relevant here but may be worth evaluating via `bonsai add`.
