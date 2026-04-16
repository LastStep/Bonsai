---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-04-14
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** _never_
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~5 min
- **Files Read:** 12 — `station/INDEX.md`, `station/CLAUDE.md`, `station/index.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`, `station/agent/Routines/doc-freshness-check.md`, `CLAUDE.md` (root), `cmd/*.go` (directory listing), `internal/generate/*.go` (directory listing), `catalog/` (item counts)
- **Files Modified:** 3 — `station/Reports/Pending/2026-04-14-doc-freshness-check.md` (this report), `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** `git log --oneline --since`, `git log --stat`, `ls`, `grep -n`, `find -name meta.yaml | wc -l`, file existence checks
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, and root `CLAUDE.md`. Compared against git log from last 7 days (5 commits: `433315f`, `fe3ad0d`, `c1c988d`, `4dfd3f4`, `64d5dec`).
- **Result:** Found 3 areas of documentation drift:
  1. The `bonsai update` command (added in `fe3ad0d`) is not reflected in INDEX.md (CLI commands count says 5, should be 6) or in the root CLAUDE.md project structure (missing `cmd/update.go`).
  2. INDEX.md says "Catalog items: ~30" but actual count is 46 (13 skills, 9 workflows, 4 protocols, 12 sensors, 8 routines).
  3. Root CLAUDE.md project structure is missing new files in `internal/generate/`: `frontmatter.go`, `frontmatter_test.go`, `scan.go`, `scan_test.go`.
- **Issues:** Drift found, flagged for user review.

### Step 2: Check INDEX.md accuracy
- **Action:** Verified tech stack table, folder structure, project description, and key metrics against actual codebase state.
- **Result:**
  - Tech stack table: accurate, no drift.
  - Project description: accurate.
  - Key Metrics table: **stale** — CLI commands says "5 (init, add, remove, list, catalog)" but there are now 6 (+ update). Catalog items says "~30" but actual is 46.
  - Document Registry table: all paths verified, all files exist.
  - Architecture Overview: accurate, no drift.
  - Agent Handoff Notes: accurate.
- **Issues:** Key Metrics row needs updating.

### Step 3: Check navigation links
- **Action:** Extracted every file path referenced in `station/CLAUDE.md` navigation tables (39 paths total across Core, Protocols, Workflows, Skills, Routines, Sensors, Code Index, and External References sections). Verified each resolves to an existing file.
- **Result:** All 39 referenced files exist. No broken links.
- **Issues:** None.

### Step 4: Check code index (station/index.md) accuracy
- **Action:** Compared every line number reference in `station/index.md` against actual function locations in the Go source using `grep -n`. Also checked for missing entries.
- **Result:**
  - **Missing entry:** `bonsai update` command (`cmd/update.go:28`, `runUpdate()`) not listed in CLI Commands table.
  - **Missing entries:** `internal/generate/frontmatter.go` (`ParseFrontmatter` at `:13`) and `internal/generate/scan.go` (`ScanCustomFiles` at `:22`, `DiscoveredFile` type at `:12`) not listed.
  - **Line number drift in `internal/generate/generate.go`:** 12 of 13 functions that live below line ~180 have drifted (by +6 to +102 lines) due to new code added for the `update` command. Specifically:
    - `ForceConflicts()`: index says `:181`, actual `:187`
    - `writeFile()`: index says `:205`, actual `:211`
    - `writeFileChmod()`: index says `:237`, actual `:243`
    - `renderContent()`: index says `:250`, actual `:256`
    - `Scaffolding()`: index says `:279`, actual `:285`
    - `SettingsJSON()`: index says `:385`, actual `:391`
    - `WorkspaceClaudeMD()`: index says `:452`, actual `:474`
    - `EnsureRoutineCheckSensor()`: index says `:583`, actual `:654`
    - `parseFrequencyDays()`: index says `:609`, actual `:680`
    - `RoutineDashboard()`: index says `:621`, actual `:692`
    - `AgentWorkspace()`: index says `:737`, actual `:839`
  - **Signature drift:** `descFor()` now takes an additional `customItems map[string]*config.CustomItemMeta` parameter not reflected in index description.
  - All other files (cmd/, catalog/, config/, tui/) have accurate line numbers.
- **Issues:** Substantial drift in generate.go line numbers and missing new files/functions.

### Step 5: Report findings and update dashboard
- **Action:** Compiled findings into this report, flagged all issues for user decision.
- **Result:** 5 findings total, all flagged for user review (no autonomous fixes per procedure).
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | CLI commands count says "5" — should be "6" (missing `update`) | `station/INDEX.md` line 34 | Flagged for user |
| 2 | medium | Catalog items count says "~30" — actual is 46 | `station/INDEX.md` line 35 | Flagged for user |
| 3 | medium | Root CLAUDE.md project structure missing `cmd/update.go` and 4 new `internal/generate/` files | `CLAUDE.md` lines 23-37 | Flagged for user |
| 4 | medium | Code index missing `bonsai update`, `frontmatter.go`, `scan.go` entries | `station/index.md` | Flagged for user |
| 5 | low | 11 line number references in code index drifted for `internal/generate/generate.go` functions | `station/index.md` lines 137-162 | Flagged for user |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review
- **INDEX.md Key Metrics stale** (`station/INDEX.md`): CLI commands count should be updated from "5 (init, add, remove, list, catalog)" to "6 (init, add, remove, list, catalog, update)". Catalog items count should be updated from "~30" to "~46".
- **Root CLAUDE.md project structure incomplete** (`CLAUDE.md`): The `cmd/` listing needs `cmd/update.go` added. The `internal/generate/` listing needs `frontmatter.go`, `frontmatter_test.go`, `scan.go`, `scan_test.go` added.
- **Code index missing new entries** (`station/index.md`): Needs a row for `bonsai update` in the CLI Commands table, and new sections for `frontmatter.go` and `scan.go` in the Generator section.
- **Code index line numbers drifted** (`station/index.md`): 11 line references for `generate.go` functions need updating. All drifted by +6 to +102 lines. Consider whether to update now or wait for code to stabilize.
- **`descFor()` signature changed** (`station/index.md`): The description should note the new `customItems` parameter added for the `update` command.

## Notes for Next Run
- The `bonsai update` command is brand new (added same day). The next run should verify whether its code index entries have been added and whether line numbers are stable.
- The root `CLAUDE.md` is a project-level file (not station-owned). Updates to it should be coordinated with the user since it serves as the primary project instructions file.
- All navigation links are currently clean (39/39 resolve). This is a good baseline.
- Catalog item count (46) may grow — consider changing INDEX.md to use "~45" or keeping it as an exact count that gets updated each routine run.
