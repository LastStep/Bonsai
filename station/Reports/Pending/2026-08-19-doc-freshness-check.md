---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-19
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
- **Duration:** ~8 min
- **Files Read:** 11 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/agent/Core/memory.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/agent/Workflows/issue-to-implementation.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/.bonsai.yaml`, `/home/user/Bonsai/catalog/skills/dispatch/meta.yaml`, `/home/user/Bonsai/station/Playbook/Standards/NoteStandards.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, grep, ls, find)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation and compare against recent git history
- **Action:** Ran `git log --since="7 days ago"` to inspect recent commits; read INDEX.md, Status.md, memory.md, CLAUDE.md.
- **Result:** Last 7 days had 2 commits — both routine executions (status-hygiene, backlog-hygiene). No new code features or architecture changes. Nothing new in code that requires doc updates.
- **Issues:** none

### Step 2: Check INDEX.md accuracy
- **Action:** Verified tech stack table, key metrics, and folder structure against actual codebase.
- **Result:** All accurate. Agent types = 6 (tech-lead, fullstack, backend, frontend, devops, security) ✓. CLI commands = 8 documented (init, add, remove, list, catalog, update, guide, validate) ✓. Catalog items = ~50 (actual: 53 meta.yaml files) — description is accurate. Architecture diagram still matches directory structure.
- **Issues:** none

### Step 3: Check navigation links
- **Action:** Parsed all markdown link targets in `station/CLAUDE.md`, `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, and `agent/Skills/`; tested each path for existence.
- **Result:**
  - `station/CLAUDE.md` — all 50+ links resolve correctly ✓
  - `agent/Core/memory.md` — **6 broken links** to `../../Research/RESEARCH-*.md` (Research/ directory does not exist). Plus 1 false positive (`url` in template example text — not a real link).
  - `agent/Protocols/` — all links resolve correctly ✓
  - `agent/Workflows/issue-to-implementation.md` — **3 broken links** to `../Skills/dispatch.md` (file not installed in workspace)
  - `agent/Workflows/session-wrapup.md` — `path` and `url` are template placeholders in example text, not actual links ✓
  - `agent/Skills/` — no internal links checked (files have no cross-links)
- **Issues:** 2 findings (see Findings Summary)

### Step 4: Report findings
- **Action:** Compiled findings table. No autonomous fixes applied per routine policy — flagged for user decision.
- **Result:** 2 findings documented below.
- **Issues:** none

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-08-19, Next Due → 2026-08-26, Status → done.
- **Result:** Done.
- **Issues:** none

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | `dispatch.md` skill referenced by `issue-to-implementation.md` (3 links: Prerequisites, Phase 7, Phase 8) but not installed in `station/agent/Skills/`. Skill exists in catalog (`catalog/skills/dispatch/`) targeting `tech-lead` and is not in `.bonsai.yaml`. Any agent following the workflow and trying to load the dispatch skill will hit a dead link. | `station/agent/Workflows/issue-to-implementation.md` lines 35, 175, 204 | Flagged for user — proposed fix: run `bonsai add` to install the `dispatch` skill |
| 2 | Medium | 6 broken links in References section pointing to `../../Research/RESEARCH-*.md` files (`landscape-analysis`, `concept-decisions`, `eval-system`, `trigger-system`, `uiux-overhaul`, `proof-of-bonsai-effectiveness`). The `Research/` directory was never created in `station/`. | `station/agent/Core/memory.md` lines 87–92 | Flagged for user — options: (a) remove the references, (b) create stubs in `station/Research/`, or (c) leave as-is if intentionally deferred |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**Finding 1 — Missing `dispatch` skill (High priority):**
The `dispatch` skill is listed in `catalog/skills/dispatch/meta.yaml` with `agents: [tech-lead]` and is not installed. The `issue-to-implementation.md` workflow makes 3 references to it. When the workflow is active and an agent tries to load the skill from the link, it will fail. Proposed fix: `bonsai add` → select `dispatch`. No other skills were found missing.

**Finding 2 — Stale Research/ references in memory.md (Medium priority):**
The References section in `agent/Core/memory.md` lists 6 research documents that don't exist on disk. These appear to be documents that were planned but never materialized (or were created in a different location). The links have been broken since at least the last doc-freshness run (2026-05-04). Proposed options:
- Remove the dead links (memory.md Notes has no active use for them if the files don't exist)
- Create `station/Research/` and add stub files for any that are still relevant
- Leave as-is if these are intentionally deferred and understood to be dead references

## Notes for Next Run

- Both findings were present from at least 2026-05-04 (last run) — they are not caused by recent commits.
- No code changes in the past 7 days means INDEX.md and architecture docs remain fresh.
- If `dispatch` skill is installed before the next run, verify the link resolves and CLAUDE.md Skills navigation table is updated.
- Check if Research/ directory has been created by next run.
