---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-09-04
status: success
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~5 min
- **Files Read:** 5 — `Playbook/Roadmap.md`, `Playbook/Status.md`, `Logs/KeyDecisionLog.md`, `agent/Core/routines.md`, `Logs/RoutineLog.md`
- **Files Modified:** 2 — `agent/Core/routines.md`, `Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state
- **Action:** Read `Playbook/Roadmap.md` and compared each item's checked/unchecked status against `Playbook/Status.md` history.
- **Result:** Phase 1 — all 11 items are checked [x] as complete. This is accurate: Status.md confirms all Phase 1 work was shipped (last was `bonsai validate` in v0.4.0). Phase 2 has one item checked [x] (custom item detection via `generate/scan.go`), which is also accurate per CLAUDE.md project structure.
- **Issues:** Phase 1 is labeled "Current Phase" but is fully complete. Phase 2 work has already begun (custom item detection done). The section header is misleading.

### Step 2: Check milestone accuracy
- **Action:** Reviewed upcoming milestones against recent Status.md entries and context note (last shipped: Plan 41, 2026-06-16; next identified: Plan 42 MCP server).
- **Result:** Plan 42 (MCP server) is the next confirmed planned item per Status.md ("MCP server = fast-follow Plan 42"), but it does not appear anywhere in the Roadmap. The headless CLI contract (Plan 41) was explicitly designed to be "MCP-ready cores," yet the MCP server build-out has no Roadmap entry. This is the most significant gap found.
- **Issues:** MCP server milestone missing from Roadmap.

### Step 3: Cross-check against Key Decision Log
- **Action:** Read `Logs/KeyDecisionLog.md` and checked for decisions that could invalidate or shift Roadmap items.
- **Result:** No decisions directly invalidate any Roadmap item. The decision "Defer Managed Agents cloud integration until local foundation is stable" (2026-04-02) remains consistent with Phase 3 being future-phased. Plan 41's headless CLI + MCP-ready cores could be interpreted as the "stable local foundation" threshold being approached, so Phase 3 may warrant a future re-evaluation conversation.
- **Issues:** One settled decision references `RESEARCH.md section 7` ("Where does Bonsai end?") — the Research/ directory does not exist in the repo. This is a stale reference in the KeyDecisionLog, not a Roadmap item.

### Step 4: Report findings
- **Action:** Compiled findings; per procedure, Roadmap.md is not modified — all items flagged for user review.
- **Result:** 3 findings documented. No Roadmap edits made.
- **Issues:** none

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Roadmap Accuracy row: Last Ran → 2026-09-04, Next Due → 2026-09-18, Status → done.
- **Result:** Done.
- **Issues:** none

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | Plan 42 (MCP server) is the confirmed next planned item but does not appear anywhere in the Roadmap | `Playbook/Roadmap.md` — Phase 2 items | Flagged for user review |
| 2 | MEDIUM | Phase 1 is labeled "Current Phase" but all 11 items are complete; Phase 2 work has already begun | `Playbook/Roadmap.md` — "Current Phase" section header | Flagged for user review |
| 3 | LOW | Settled decision references `RESEARCH.md section 7` which does not exist in the repo | `Logs/KeyDecisionLog.md` — Settled section | Flagged for user review |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[HIGH] Add MCP server (Plan 42) to Roadmap.md** — Status.md shows "MCP server = fast-follow Plan 42" as the confirmed next planned item. Suggest adding it under Phase 2 or as a bridge item. Candidate wording: `[ ] MCP server layer — expose headless CLI cores via MCP protocol (Plan 42)`.

2. **[MEDIUM] Phase 1 "Current Phase" label** — Phase 1 is fully complete; Phase 2 has begun. Options: (a) Relabel Phase 1 as "Phase 1 — Complete" and promote Phase 2 to "Current Phase", or (b) Add a brief "Phase 2 (in progress)" marker. Either brings the Roadmap into sync with reality.

3. **[LOW] Stale RESEARCH.md reference in KeyDecisionLog** — The settled decision "Bonsai is a scaffolding tool, not a runtime orchestrator" cites "RESEARCH.md section 7 — 'Where does Bonsai end?'" but the Research/ directory does not exist. Consider replacing the cite with a brief inline rationale or removing the dead link.

## Notes for Next Run

- If Plan 42 is shipped before next run (2026-09-18), confirm it's checked [x] in Roadmap.md.
- Phase 3 (Managed Agents) re-evaluation may be warranted now that the headless CLI contract is shipped — worth raising with user as a conversation topic, not a roadmap change to make unilaterally.
- Dependency Audit and Vulnerability Scan were last run 2026-05-04 (now ~123 days overdue) — these are outside this routine's scope but worth flagging here since they were visible in the dashboard.
