---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-04-30
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-04-21
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 10 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/internal/generate/catalog_snapshot.go`, `/home/user/Bonsai/.claude/settings.json`
- **Files Modified:** 3 — `station/Reports/Pending/2026-04-30-doc-freshness-check.md` (this report), `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** `git log --oneline --since="7 days ago"`, `git log --name-only`, `ls /home/user/Bonsai/internal/tui/`, `ls /home/user/Bonsai/internal/`, `grep` patterns on CLAUDE.md/code-index.md, `find catalog -name meta.yaml | wc -l`, link resolution loop
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read `station/INDEX.md`, `station/CLAUDE.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, and `station/code-index.md`. Ran `git log --oneline --since="7 days ago" --name-only` to get 21 commits and their changed files over the past 7 days.
- **Result:** 21 commits since 2026-04-23. Major new features shipped: Plan 31 (removeflow cinematic, updateflow cinematic, hints 3-layer, catalog --json/NO_COLOR), Plan 32 (wsvalidate extraction, config.Validate, catalog_snapshot hardening), Plan 33 (website rewrite), NoteStandards addition. New packages exist that are not reflected in several docs.
- **Issues:** none in execution; several doc drift findings (see Findings Summary)

### Step 2: Check INDEX.md accuracy
- **Action:** Verified tech stack table, key metrics, folder structure, and architecture overview against actual codebase state.
- **Result:** Tech stack is accurate. Key metrics: agent types = 6 (correct), CLI commands = 7 (correct), catalog items = ~50 (actual 53 — acceptable approximation). Architecture diagram (`INDEX.md`) is a simplified overview and does not enumerate tui subpackages, so no update needed there. INDEX.md is fundamentally accurate.
- **Issues:** none requiring update to INDEX.md

### Step 3: Check navigation links
- **Action:** Extracted all link targets from `station/CLAUDE.md` and verified each resolves to a real file or directory. Checked all Core, Protocols, Workflows, Skills, Routines, Sensors links.
- **Result:** 36 links checked — all 36 resolve to real files/directories. No broken links found. However, `statusline.sh` exists in `agent/Sensors/` but is NOT listed in the Sensors table — it is a personal script (not a registered hook per `.claude/settings.json`) and CLAUDE.md omits it. Also, `station/Playbook/Standards/NoteStandards.md` (added this week) is not referenced in the External References table of `station/CLAUDE.md`.
- **Issues:** 2 minor gaps (see findings #3 and #4)

### Step 4: Report findings
- **Action:** Compiled all drift items discovered from git log cross-reference and link verification.
- **Result:** 5 findings identified: 2 medium (root CLAUDE.md project structure tree, code-index.md missing new packages), 2 low (statusline.sh undocumented, NoteStandards not in External References), 1 info (generate/ new files not in code-index). All flagged for user decision — no auto-edits executed per routine design (audit-only).
- **Issues:** none

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for Doc Freshness Check: Last Ran → 2026-04-30, Next Due → 2026-05-07, Status → done.
- **Result:** Dashboard updated.
- **Issues:** none

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | Root `Bonsai/CLAUDE.md` project-structure tree `internal/tui/` block only lists `harness/` and `initflow/` — missing `addflow/`, `catalogflow/`, `guideflow/`, `listflow/`, `removeflow/`, `updateflow/`, `hints/` (shipped in Plans 23, 28, 31, 32) | `Bonsai/CLAUDE.md` lines 48–55 | Flagged for user update |
| 2 | medium | `station/code-index.md` has no sections for new packages: `removeflow/`, `updateflow/`, `hints/`, `wsvalidate/`, `catalogflow/`, `listflow/`, `guideflow/`; also missing `catalog_snapshot.go` exports (`SerializeCatalog`, `WriteCatalogSnapshot`) | `station/code-index.md` | Flagged for user update |
| 3 | low | `statusline.sh` present in `agent/Sensors/` but not listed in Sensors nav table in `station/CLAUDE.md` — appears to be a personal/custom script not registered in `.claude/settings.json` hooks | `station/CLAUDE.md` Sensors table | Flagged for user decision (intentional omission vs oversight) |
| 4 | low | `station/Playbook/Standards/NoteStandards.md` (added 2026-04-30 per recent commits) not referenced in External References table in `station/CLAUDE.md` | `station/CLAUDE.md` External References | Flagged for user update |
| 5 | info | `internal/wsvalidate/` package (workspace-path validation, extracted in Plan 32) not documented in `code-index.md` alongside generate/ and config/ sections | `station/code-index.md` | Flagged for user update |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[medium] Root `Bonsai/CLAUDE.md` tui structure drift** — The project-structure tree still shows only `harness/` and `initflow/` under `internal/tui/`. At minimum, `addflow/`, `removeflow/`, `updateflow/`, `hints/`, `catalogflow/`, `listflow/`, `guideflow/` should be added. This is the same class of finding as the 2026-04-21 run (which flagged `harness/` and `styles_test.go` missing at the time). Recommend updating during next code-index refresh pass.

2. **[medium] `station/code-index.md` missing new package sections** — 7 new TUI flow packages and 2 new non-TUI packages (`wsvalidate`, generate's `catalog_snapshot.go`) shipped since the last code-index refresh (which covered through Plan 22/add-flow). A code-index refresh pass would close all of: removeflow, updateflow, hints, wsvalidate, catalogflow, listflow, guideflow, and catalog_snapshot. This is P2 Backlog work (Group E has existing "code-index drift" entries).

3. **[low] `statusline.sh` sensor documentation** — File exists in `agent/Sensors/` and is not a Bonsai-managed hook (not in `.claude/settings.json`). Either: (a) it's an intentional personal script and the Sensors table correctly omits it, or (b) it should be listed as a note in the nav. User should confirm intent.

4. **[low] NoteStandards.md not in External References** — `station/Playbook/Standards/NoteStandards.md` was wired into Status.md, Backlog.md, session-logging, and memory protocol workflows this week but the External References table in `station/CLAUDE.md` only lists `SecurityStandards.md`. Consider adding a row: `| Note formatting standards | [station/Playbook/Standards/NoteStandards.md](Playbook/Standards/NoteStandards.md) |`.

## Notes for Next Run

- Root `Bonsai/CLAUDE.md` tui structure drift is a persistent pattern — updated after Plan 15, flagged again after Plan 22, flagged again now. Consider adding a sub-step to the doc-freshness procedure: "compare `ls internal/tui/` against CLAUDE.md project structure entry."
- code-index.md refresh should be treated as a quarterly or per-major-release task; current drift accumulated across Plans 23–32.
- All nav links clean — no link rot this cycle.
