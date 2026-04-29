---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-04-29
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-04-21
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~10 min
- **Files Read:** 14 — `station/agent/Routines/doc-freshness-check.md`, `station/CLAUDE.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Playbook/Backlog.md`, `station/code-index.md`, `station/Logs/RoutineLog.md`, `internal/generate/generate.go`, `internal/catalog/catalog.go`, `internal/config/config.go`, `internal/config/lockfile.go`, `cmd/init_flow.go`, `cmd/add.go`, `cmd/remove.go`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** `git log --since="7 days ago" --oneline --name-only`, `ls` on `cmd/`, `internal/`, `internal/tui/`, `internal/generate/`, `catalog/`, `catalog/agents/`, `catalog/skills/`, `catalog/sensors/`, `catalog/routines/`; `grep -n` on multiple source files; `find` for meta.yaml counts
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history (last 7 days)
- **Action:** Ran `git log --since="7 days ago"` and enumerated changed files. Cross-referenced against station/ docs.
- **Result:** 16 commits in the 7-day window. Key changes: NoteStandards scaffolding + skill added (Plans 32/33); wsvalidate package extracted; `internal/tui/updateflow` + `internal/tui/hints` added in Plan 31; `catalog_snapshot.go` added in Plan 31/32; `RefreshPeerAwareness()` function added. None of these structural additions are reflected in `station/code-index.md` or the root `CLAUDE.md` project structure tree.
- **Issues:** Significant drift between codebase and code-index documentation — see Findings.

### Step 2: Check INDEX.md accuracy
- **Action:** Compared INDEX.md tech stack, folder structure, project description, and key metrics against reality.
- **Result:** Tech stack is accurate. Agent count (6) is accurate. Catalog item count shown as "~50" vs actual 53 (18 skills, 10 workflows, 4 protocols, 13 sensors, 8 routines) — within "~50" range but could be updated. CLI command count (7) is accurate for top-level commands. Architecture diagram is accurate. The INDEX.md is in good shape.
- **Issues:** Minor — catalog count "~50" is technically 53. Acceptable approximation.

### Step 3: Check navigation links
- **Action:** Checked every link in `station/CLAUDE.md` — Core (3 files), Protocols (4), Workflows (9), Skills (5), Routines (7), Sensors (9), and External References (7 paths).
- **Result:** All 44 links resolve to real files. No broken navigation links found.
- **Issues:** None.

### Step 4: Report findings
- **Action:** Documented all drift instances below.
- **Result:** 3 findings of material impact — all in `station/code-index.md` and root `CLAUDE.md` project structure tree. No findings require immediate action to unblock work, but code-index line numbers are far enough off (up to 143 lines) that they reduce rather than aid navigation.
- **Issues:** Findings flagged for user review — see below. Not autonomously updated per routine procedure ("propose updates but don't execute — flag for user decision").

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` — Doc Freshness Check row `Last Ran` → 2026-04-29, `Next Due` → 2026-05-06, `Status` → done.
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `code-index.md` line numbers are stale across all major source files — drifted by 1–143 lines after Plans 31/32 added code. Key offsets: `remove.go` functions off by ~99, `generate.go` functions off by 7–143, `init_flow.go` off by ~24, `add.go` off by ~39. `catalog.go New()` off by 22 (`:220` → actual `:242`). | `station/code-index.md` | Flagged for user — update requires a pass through every function entry in the code-index. |
| 2 | Medium | `code-index.md` missing entries for new packages and files added in Plans 31/32: `internal/wsvalidate/` (Normalise, InvalidReason), `internal/generate/catalog_snapshot.go` (SerializeCatalog, WriteCatalogSnapshot), `internal/generate/generate.go:RefreshPeerAwareness()` (`:1590`), `internal/tui/hints/` package, `internal/tui/updateflow/` package, `internal/tui/catalogflow/` package, `internal/tui/listflow/` package, `internal/tui/removeflow/` package, `internal/tui/guideflow/` package. | `station/code-index.md` | Flagged for user — new sections needed. |
| 3 | Low | Root `CLAUDE.md` project structure tree documents only two TUI subpackages (`harness/`, `initflow/`) and does not list the 7 new flow packages added since Plan 15 (`addflow/`, `updateflow/`, `removeflow/`, `catalogflow/`, `listflow/`, `guideflow/`, `hints/`). Also missing `internal/wsvalidate/` entirely. `generate/catalog_snapshot.go` not listed in the tree. | `/home/user/Bonsai/CLAUDE.md` (project root) | Flagged — noted in Backlog as improvement item (2026-04-21). Not escalated further. |

## Errors & Warnings

No errors encountered.

Note: Finding #3 (root CLAUDE.md project structure drift) was already filed in Backlog.md as a P3 improvement item on 2026-04-21: `[improvement] Add root Bonsai/CLAUDE.md check to doc-freshness-check routine`. The finding is consistent with that existing entry — no new Backlog item needed.

## Items Flagged for User Review

**Finding 1 + 2 — code-index.md refresh:** The code-index has drifted significantly enough to be misleading for navigation. Recommend a targeted update pass:
- Update line numbers for all existing entries (many are off by 20–140+)
- Add sections for: `internal/wsvalidate/`, `internal/generate/catalog_snapshot.go`, `internal/tui/updateflow/`, `internal/tui/removeflow/`, `internal/tui/catalogflow/`, `internal/tui/listflow/`, `internal/tui/guideflow/`, `internal/tui/hints/`
- Add `RefreshPeerAwareness()` (`:1590`) to the generate.go Core Generation Functions table

**Finding 3 — root CLAUDE.md tree:** Already tracked in Backlog — no new action needed from user unless they want to prioritize it.

## Notes for Next Run

- code-index.md line numbers will continue to drift with each Plan; consider whether the code-index approach (absolute line numbers) is the right design or whether function-name search is sufficient. If line numbers are kept, a subagent pass to update them would be appropriate after each multi-file Plan.
- catalog item count in INDEX.md is now 53 (was "~50") — worth updating to "~55" or exact count on next INDEX.md refresh.
- All navigation links were clean — no link rot detected.
