---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-18
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
- **Duration:** ~12 min
- **Files Read:** 11 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/CLAUDE.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/agent/Core/self-awareness.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/internal/validate/project.go` (head), `/home/user/Bonsai/CHANGELOG.md` (head)
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against git history
- **Action:** Ran `git log --oneline --since="7 days ago" --name-only` to identify what changed. Also checked broader 45-day history for context. Catalogued all changed files.
- **Result:** 7-day window has significant changes from Plans 40 and 41: Plan 41 (Phases 1–5) added `internal/nonint/` package (headless CLI contract), `cmd/list --json`, headless update/remove flags, `docs/agent-interface.md`, `docs/formats.md`. Plan 40 added `internal/validate/project.go`, new scaffolding items, `docs/formats.md`.
- **Issues:** Documentation in `Bonsai/CLAUDE.md` and `station/code-index.md` did not reflect these additions.

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` in full. Verified tech stack, folder structure, key metrics, architecture overview, and document registry against actual codebase.
- **Result:** Tech stack is accurate. Key metrics: Agent types (6) ✓, CLI commands (8, hidden completion excluded) ✓, Catalog items (~50 — actual count is 53: 18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines) ✓ (within ~50 estimate). Architecture diagram lists `internal/catalog`, `internal/config`, `internal/generate`, `internal/validate`, `internal/wsvalidate`, `internal/tui` — **missing `internal/nonint/`** added by Plan 41. Document registry is complete and accurate.
- **Issues:** 1 low-severity drift — architecture diagram missing `internal/nonint/`.

### Step 3: Check navigation links
- **Action:** Enumerated all linked files in `station/CLAUDE.md` navigation tables (Skills, Workflows, Protocols, Core, Routines), then used `test -f` checks for each.
- **Result:** All 28 links checked — all resolve:
  - Skills (7 files): bonsai-model.md ✓, bubbletea.md ✓, issue-classification.md ✓, planning-template.md ✓, pr-creation.md ✓, review-checklist.md ✓, critic-agent-prompts.md ✓
  - Workflows (10 files): all 10 resolve ✓
  - Protocols (4 files): all 4 resolve ✓
  - Core (4 files): all 4 resolve ✓
  - Routines (7 files): all 7 resolve ✓
  - agent/Core/, agent/Protocols/, agent/Workflows/, agent/Skills/ subdir links all valid ✓
- **Issues:** None — zero broken navigation links. Previous cycles flagged `bonsai-model.md` as broken; this is now resolved.

### Step 4: Report findings
- **Action:** Compiled all drift items found. Per procedure, findings are flagged for user decision rather than auto-applied.
- **Result:** 5 drift items identified (see Findings Summary below). All are in `Bonsai/CLAUDE.md` (root) or `station/code-index.md`. No broken links. No stale scaffolding.
- **Issues:** None — report written as designed.

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for Doc Freshness Check: Last Ran → 2026-06-18, Next Due → 2026-06-25, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package (Plan 41) missing from project structure tree | `Bonsai/CLAUDE.md` lines 37–74 | Flagged for user — propose adding nonint entry after wsvalidate |
| 2 | MEDIUM | `internal/generate/` tree missing `catalog_snapshot_unix.go`, `catalog_snapshot_windows.go`, `list_snapshot.go` | `Bonsai/CLAUDE.md` lines 50–53 | Flagged for user — 3 files to add to generate/ block |
| 3 | MEDIUM | `internal/validate/` tree missing `project.go` (Plan 40 Phase 2) | `Bonsai/CLAUDE.md` lines 54–56 | Flagged for user — project.go + project_test.go to add |
| 4 | MEDIUM | `code-index.md` missing entire `internal/nonint/` package section | `station/code-index.md` | Flagged for user — needs new section after wsvalidate section |
| 5 | LOW | `code-index.md` Validate section missing `project.go` functions | `station/code-index.md` Validate section | Flagged for user — project-level validation functions not documented |
| 6 | LOW | `station/INDEX.md` architecture diagram missing `internal/nonint/` | `station/INDEX.md` lines 66–71 | Flagged for user — add one line to arch overview |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

All 6 findings above require user/agent attention. Priority order for a doc-refresh pass:

1. **Root `Bonsai/CLAUDE.md` project structure tree** (Findings 1–3) — three gaps from Plans 40/41. Recurring drift pattern on this file; previous routine logs (2026-05-04, 2026-04-21) flagged this same file multiple times. Consider promoting the Backlog P2 item for a root-CLAUDE.md sub-step procedure to P1.
2. **`station/code-index.md`** (Findings 4–5) — `internal/nonint/` is a substantial new package (headless CLI contract for all mutating commands). Without a code-index entry, agents navigating the codebase won't know where to find the headless runner, exit contract, or event types.
3. **`station/INDEX.md` architecture diagram** (Finding 6) — minor one-liner addition; lowest priority.

## Notes for Next Run

- Navigation links are now all clean — previous recurring `bonsai-model.md` broken link is resolved (file exists at `station/agent/Skills/bonsai-model.md`).
- `internal/nonint/` drift is likely to recur on next cycle if not addressed — Plan 41 was the largest recent change and touches 3 different doc files.
- Root `Bonsai/CLAUDE.md` project structure tree has been flagged across 5 of the last 6 routine runs — consider whether the backlog P2 "root-CLAUDE.md check sub-step for doc-freshness routine" should be promoted to P1 and executed as a plan.
- Catalog item count is now 53 (18+10+4+13+8). INDEX.md says "~50" — still accurate as an estimate; no update needed.
