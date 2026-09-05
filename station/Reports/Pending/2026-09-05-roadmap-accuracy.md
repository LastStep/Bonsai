---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-09-05
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
- **Files Read:** 6 — `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/agent/Core/memory.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Write, Edit, Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state

Read `station/Playbook/Roadmap.md`. Phase 1 — Foundation & Polish shows all 11 items checked including `bonsai validate` (Plan 35). The two flags from the previous run (2026-05-07) have been resolved: "Better trigger sections" is now `[x]` with an annotation, and `bonsai validate` was added as the final Phase 1 item.

**Phase 1 is fully complete.** However, the roadmap still labels Phase 1 as "Current Phase" with Phase 2 under a "Future Phases" header. Phase 2 already has one checked item (Custom item detection), meaning active work is happening in Phase 2.

Checking alignment with `Status.md`: Recently Done includes Plan 41 (Headless CLI Contract, shipped 2026-06-16) and Plan 40 (frozen v1 schemas + root-relative scaffolding, merged 2026-06-13). Neither appears in the roadmap.

### Step 2: Check milestone accuracy

Phase 2 next unchecked items: Self-update mechanism, Template variables expansion, Micro-task fast path. None of these are in active plans. Instead, Work State in memory.md points to Plan 42 (MCP server, `bonsai mcp`) as the concrete next item. Plan 42 is not listed in the roadmap at all.

Plan 40 partially touches "Template variables expansion" (frozen v1 schemas, root-relative scaffolding), but it was primarily a platform-integration plan, not template expansion per se.

Phase 3 "Managed Agents integration" — the KeyDecisionLog decision was "Defer until local foundation is stable." Local foundation (Phase 1) is now stable and confirmed shipped. The timing question is open.

### Step 3: Cross-check against Key Decision Log

No recent decisions in `KeyDecisionLog.md` explicitly invalidate existing roadmap items. The Headless CLI Contract (Plan 41) represents a strategic architectural decision — building MCP-ready `*Result` cores and the `agent-interface.md` contract — that isn't reflected in any roadmap phase and wasn't logged as a key decision.

The "Bonsai is a scaffolding tool, not a runtime orchestrator" settled decision is still consistent with Phase 1 deliverables. Phase 3 (Managed Agents) represents intentional future scope creep toward orchestration, which was previously deferred.

### Step 4: Report findings

Four findings flagged (see table below). No Roadmap.md modifications made — flagged for user review.

### Step 5: Update dashboard

`agent/Core/routines.md` Roadmap Accuracy row updated: Last Ran → 2026-09-05, Next Due → 2026-09-19, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | "Current Phase" header is stale — Phase 1 is fully done, Phase 2 has a checked item and is the active phase | `Roadmap.md` §Current Phase | Flagged for user — recommend updating header to Phase 2 and moving Phase 1 to a completed section |
| 2 | MEDIUM | Plan 41 (Headless CLI Contract + MCP-ready cores, shipped 2026-06-16) is not represented in any roadmap phase — significant architectural capability with no roadmap home | `Roadmap.md` Phase 2 | Flagged for user — recommend adding item under Phase 2 (e.g., "Headless CLI contract + MCP-ready interface — `*Result` cores, JSONL/exit contract, `docs/agent-interface.md`") |
| 3 | LOW | Plan 42 (MCP server, `bonsai mcp` via go-sdk stdio) is the concrete next planned item per memory.md Work State but does not appear in any roadmap phase | `Roadmap.md` Phase 2 | Flagged for user — recommend adding item under Phase 2 or Phase 3 depending on scope classification |
| 4 | LOW | Phase 3 prerequisite (stable local foundation) is now met — KeyDecisionLog "defer until foundation stable" condition is satisfied. Timing question: is Phase 3 work now on the horizon? | `Roadmap.md` Phase 3 | Flagged for user — no roadmap change needed, but a priority-check question for the user |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **MEDIUM — Promote Phase 2 to "Current Phase"**: Phase 1 is fully shipped. Phase 2 has at least one checked item. The current header structure is misleading. Suggest moving Phase 1 into a "Completed" section and making Phase 2 the "Current Phase."

2. **MEDIUM — Add Plan 41 to Roadmap Phase 2**: The Headless CLI Contract / MCP-ready interface work (Plan 41) represents a significant shipped capability that should be documented in the roadmap for historical accuracy and to communicate direction to external contributors. Suggested new item: `[x] Headless CLI contract — pure *Result cores, JSONL/exit contract (ExitConflict=5), list --json, docs/agent-interface.md`

3. **LOW — Add MCP server to Roadmap**: Plan 42 (`bonsai mcp`, go-sdk, stdio transport) is the stated next item per Work State. Adding it to Phase 2 or Phase 3 would make the roadmap forward-looking and honest about direction.

4. **LOW — Phase 3 timing check**: The deferral condition for Managed Agents ("until local foundation is stable") has been met. Worth a quick user decision: is Phase 3 still deferred, or is it now active horizon?

## Notes for Next Run

- Previous run flags (Better trigger sections, bonsai validate missing) were resolved — no need to re-check.
- Watch for Plan 42 progress — if it ships before next run, roadmap should reflect the addition.
- If user promotes Phase 2 to current and adds Plan 41/42 items, next run will start from a cleaner baseline.
- Phase 2 "Self-update mechanism" and "Template variables expansion" appear low-priority given active direction toward MCP; worth asking if they're still on the Phase 2 list or deprioritized.
