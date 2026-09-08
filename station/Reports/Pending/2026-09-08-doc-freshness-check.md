---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-08
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~10 min
- **Files Read:** 8 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/CLAUDE.md` (via system context), `station/code-index.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, file existence checks, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation vs. recent git history
- **Action:** Read `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`; ran `git log --oneline --since="2026-05-04"` to identify commits since the last doc freshness check.
- **Result:** 48 commits since 2026-05-04. Key code changes: Plan 39 (`bonsai init/add --non-interactive`), Plan 40 Phases 1–3 (frozen schemas, project-level validate, memory-routing docs, guide Formats page), Plan 41 Phases 1–5 (headless CLI contract + `internal/nonint/` package, `list --json`, `agent-interface.md`). Also: `bonsai completion` command from external contribution (PR #78, 2026-05-07).
- **Issues:** `code-index.md` has significant drift against these changes — see Findings.

### Step 2: Check INDEX.md accuracy
- **Action:** Compared INDEX.md tech stack, folder structure, CLI command count, and catalog item count against current codebase.
- **Result:** Tech stack is accurate. Architecture overview is accurate. One metric is stale: "CLI commands | 8" — the `completion` command was added in PR #78, making the real count 9.
- **Issues:** Minor numeric drift (8 → 9 commands; catalog item count says ~50, actual directory count is ~57 — within acceptable approximation range).

### Step 3: Check navigation links
- **Action:** Enumerated all 48 linked file paths from `station/CLAUDE.md` navigation tables and checked each against the filesystem.
- **Result:** All 48 links resolve correctly. No broken navigation links.
- **Issues:** None.

### Step 4: Report findings
- **Action:** Identified documentation drift in `code-index.md` relative to Plans 39–41 and the `completion` command addition. Catalogued findings below.
- **Result:** 5 findings, all flagged for user review per procedure (no edits executed).
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for Doc Freshness Check.
- **Result:** `Last Ran` → 2026-09-08, `Next Due` → 2026-09-15, `Status` → done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `internal/nonint/` package (Plans 39 + 41) is entirely absent from the code index — no section, no types, no public functions | `station/code-index.md` | Flagged for user — no edit made |
| 2 | Medium | `bonsai completion` command (PR #78, 2026-05-07) is missing from the CLI Commands table | `station/code-index.md` | Flagged for user — no edit made |
| 3 | Low | `internal/generate/list_snapshot.go` (Plan 41 Phase 4) is not documented — exports `ListSnapshot`, `ListAgent`, `SerializeJSON()` | `station/code-index.md` | Flagged for user — no edit made |
| 4 | Low | `internal/validate/project.go` (Plan 40 Phase 2) is not reflected — adds the project-level validate pass (memory-tree audit, manifest checking) | `station/code-index.md` | Flagged for user — no edit made |
| 5 | Low | KEY METRICS table: "CLI commands \| 8" is one short — should be 9 after `completion` was added | `station/INDEX.md` | Flagged for user — no edit made |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

All 5 findings above require user decision on whether to update documentation. Key details:

**Finding 1 (High Impact) — `internal/nonint/` package missing from code-index.md**

This is the headless CLI contract package (Plans 39 + 41). It has substantial public surface:
- `config.go` — `LoadConfig()`, `applyDefaults()`
- `result.go` — `Result`, `Counts`
- `events.go` — `EmitJSONL()`, `EmitFile()`, `EmitSummary()`
- `runner.go` — `RunInit()`, `RunAdd()`, exit code constants (`ExitOK=0`, `ExitConflict=5`, etc.)
- `remove.go` — headless remove core
- `update.go` — headless update core

Suggested action: add a new `## Headless / Non-Interactive (`internal/nonint/`)` section after the Validate section in `code-index.md`.

**Finding 2 — `bonsai completion` missing from CLI Commands table**

Add row: `bonsai completion | cmd/completion.go | completionCmd → generate bash/zsh/fish/powershell completion scripts`

**Finding 3 — `list_snapshot.go` not in Generator section**

Add row under `internal/generate/`: `SerializeJSON()` — render installed agent state as structured JSON for `list --json`.

**Finding 4 — `project.go` not noted in Validate section**

`project.go` implements the project-level audit (memory-tree, manifest, permalink hygiene, relation links). All functions are unexported, so documentation is a prose note rather than a function table. Add a note in the Validate section.

**Finding 5 — INDEX.md metric: CLI commands count**

Update "8" → "9" in the Key Metrics table.

## Notes for Next Run

- All 5 findings are documentation-only and non-blocking. Address with a dedicated `code-index.md` + `INDEX.md` refresh pass.
- `code-index.md` should get a post-Plan-41 sweep as a one-time doc catch-up rather than piecemeal edits.
- Next doc freshness check: verify `docs/agent-interface.md` (Plan 41 Phase 5) is reachable and linked from somewhere in station (it lives in the project root, not station, so may not need a station link).
