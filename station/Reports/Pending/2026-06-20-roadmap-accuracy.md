---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-06-20
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
- **Duration:** ~8 min
- **Files Read:** 5 — `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/Playbook/Backlog.md`, `station/agent/Routines/roadmap-accuracy.md`
- **Files Modified:** 3 — `station/Reports/Pending/2026-06-20-roadmap-accuracy.md` (created), `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry appended)
- **Tools Used:** Read, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `station/Playbook/Roadmap.md` in full.

**Phase 1 — Foundation & Polish:** All 11 items are marked `[x]`. Cross-checked each against Status.md and RoutineLog.md history:
- All items verified as genuinely shipped. The two fixes from the last routine-digest (2026-05-07) — "Better trigger sections" `[x]` with annotation and `bonsai validate` row addition — are correctly reflected in the current Roadmap.md.
- Phase 1 is **clean and accurate**.

**Phase 2 — Extensibility:**
- `[x]` Custom item detection — confirmed shipped (Plan 32 / catalog scan).
- `[ ]` Self-update mechanism — confirmed unbuilt; correctly unchecked; in Backlog P3.
- `[ ]` Template variables expansion — confirmed unbuilt; correctly unchecked.
- `[ ]` Micro-task fast path — confirmed unbuilt; correctly unchecked; in Backlog P3.
- Phase 2 checkboxes are accurate. **However, Plan 41 (shipped 2026-06-16) established a headless CLI contract that is explicitly framed as the MCP server enabler ("MCP server = fast-follow Plan 42"). This upcoming Plan 42 is not represented anywhere on the Roadmap.** An MCP server is a significant Phase 2 or Phase 3 deliverable.

**Phase 3 — Cloud & Orchestration:**
- `[ ]` Managed Agents integration — deferred per KeyDecisionLog ("local foundation stable" gate). Plan 41 headless cores may now satisfy that gate — flagged for user decision.
- `[ ]` Greenhouse companion app — design phase, no active work; correctly unchecked; in Backlog Big Bets.

**Phase 4 — Ecosystem:** All unchecked; no active work; correctly unchecked.

### Step 2 — Check milestone accuracy

Assessed the next logical milestones against recent work:

1. **Plan 42 (MCP server)** is explicitly the next planned major feature (Status.md Plan 41 notes), but has no Roadmap entry. It fits naturally as either a Phase 2 item (extensibility — external tooling integration) or Phase 3 (Cloud & Orchestration).

2. **Plan 40 Phase 4 (update-delivery)** remains held. The Roadmap has no row for this, and it maps loosely to Phase 2 "Self-update mechanism" but isn't the same thing. No inaccuracy — it's tracked elsewhere.

3. **Plan 41 headless CLI contract** could be considered the enabling infrastructure for Phase 3 (Managed Agents). The KeyDecisionLog's "defer until local foundation stable" condition may now be satisfied. Worth a user review of whether Phase 3 priority should be reconsidered.

4. **Website security (P2 Backlog)** — astro/vite npm vulns. Not a Roadmap concern.

### Step 3 — Cross-check against KeyDecisionLog

Read `station/Logs/KeyDecisionLog.md` in full.

- No KeyDecisionLog decisions invalidate current Roadmap items.
- The "Defer Managed Agents cloud integration until local foundation is stable" decision may be ready for re-evaluation now that Plan 41's headless CLI contract ships (the primary "local foundation stable" prerequisite cited). This is flagged for user review — not a correction.
- All architectural decisions (Go rewrite, embed.FS, text/template) are consistent with the current Phase 1 complete state.

### Step 4 — Report findings

No direct modifications to `Roadmap.md` per procedure. Findings flagged for user review below.

### Step 5 — Update dashboard

Updated `station/agent/Core/routines.md` dashboard row for Roadmap Accuracy: Last Ran → 2026-06-20, Next Due → 2026-07-04, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | Plan 42 (MCP server) is explicitly the next planned deliverable but has no Roadmap entry. It fits Phase 2 (extensibility) or Phase 3 (Cloud & Orchestration). | `Roadmap.md` — Phase 2 or 3 | Flagged for user review — recommend adding a row |
| 2 | LOW | Plan 41 headless CLI contract may satisfy the KeyDecisionLog "local foundation stable" gate for Phase 3 Managed Agents. Phase 3 priority could be reconsidered. | `Roadmap.md` Phase 3 / `KeyDecisionLog.md` | Flagged for user review — no change warranted without user input |
| 3 | INFO | Phase 1 is fully complete and accurate — all 11 items correctly marked `[x]`. | `Roadmap.md` Phase 1 | No action needed |
| 4 | INFO | Phase 2–4 unchecked items correctly reflect unbuilt state. | `Roadmap.md` Phases 2–4 | No action needed |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**F1 — MEDIUM: Add MCP server (Plan 42) to Roadmap**

Plan 41 notes explicitly state "MCP server = fast-follow Plan 42." The MCP server is the next planned major feature and is not on the Roadmap. Recommend adding a row to Phase 2 or Phase 3:

- Phase 2 option: `[ ] MCP server integration — expose headless CLI cores as an MCP tool server`
- Phase 3 option: Could be co-located with Managed Agents as part of Cloud & Orchestration

Suggested Phase 2 placement (as an extensibility item that enables third-party integrations), but user should decide.

**F2 — LOW: Reconsider Phase 3 (Managed Agents) deferral**

The KeyDecisionLog deferred Managed Agents "until local foundation is stable." Plan 41 (2026-06-16) shipped the headless CLI contract that is the explicit prerequisite. The local foundation may now qualify as stable. If so, Phase 3 could be elevated from "Big Bets backlog" to active roadmap priority. No change recommended without user decision.

## Notes for Next Run

- Watch for Plan 42 (MCP server) to be added to Roadmap — if it ships before next run, it should be checked `[x]`.
- If Phase 3 is re-prioritized, Roadmap Phase 3 description may need updating to reflect MCP server as the integration mechanism rather than a bespoke "Managed Agents integration."
- Phase 1 is solidly complete. No re-checks needed there.
