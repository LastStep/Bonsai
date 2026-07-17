---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-07-17
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
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/roadmap-accuracy.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Glob, Bash (ls), Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

**Step 1 — Compare Roadmap against current state:**
Read `Roadmap.md`. Phase 1 ("Foundation & Polish") has all items checked `[x]`, including the last-shipped item `bonsai validate` (Plan 35, v0.4.0). The "Better trigger sections" annotation and `bonsai validate` row were confirmed — both were applied by the 2026-05-07 Routine Digest (the previous run's flags were resolved). Phase 1 is fully complete.

The roadmap's `## Current Phase` header still points to Phase 1. Phase 2 has started: `[x] Custom item detection` is checked. Additionally, Plan 41 (Headless CLI Contract + MCP-ready cores) shipped 2026-06-16 — a significant Phase 2/3 milestone with no roadmap entry.

**Step 2 — Check milestone accuracy:**
Cross-referenced `Roadmap.md` against `Status.md` and `Backlog.md`. Key discrepancies:

- The "Current Phase" label is stale — Phase 1 is 100% done.
- Plan 41 (5 phases, PRs #120/#122/#123/#121/#125) shipped a headless CLI contract and MCP-ready cores. The Backlog P1 entry "[feature] Full agent-drivable CLI parity" is commented out as RESOLVED by Plan 41. There is no corresponding roadmap item.
- Status.md notes: "MCP server = fast-follow Plan 42." Phase 3 describes "Managed Agents integration — `bonsai deploy`" but an MCP server is a distinct, closer-term item that doesn't map to any existing roadmap entry.
- v0.5.0 shipped (Plan 40 Phases 1–3) but the tag was held by the user. Minor — doesn't affect roadmap accuracy.

Phase 2's remaining items (`Self-update mechanism`, `Template variables expansion`, `Micro-task fast path`) all appear in the Backlog as P3 "Future Platform" entries — correctly classified, no urgency drift detected.

**Step 3 — Cross-check against Key Decision Log:**
Read `KeyDecisionLog.md`. Structural and Domain-Specific decisions all remain valid — no roadmap items are invalidated by them.

One Settled decision is worth flagging: *"Defer Managed Agents cloud integration until local foundation is stable."* Plan 41 explicitly shipped "MCP-ready cores" and Plan 42 (MCP server) is described as imminent. The local foundation is clearly now considered stable. The decision has not been formally revisited, but Phase 3 work appears to have effectively begun. The roadmap's Phase 3 description may need updating to reflect MCP server as a concrete near-term item alongside the longer-horizon "Managed Agents" integration.

**Step 4 — Report findings:**
Four findings flagged for user review (see table below). Roadmap.md not modified — all changes flagged per procedure.

**Step 5 — Update dashboard:**
Updated `agent/Core/routines.md` dashboard row for Roadmap Accuracy: `Last Ran` → 2026-07-17, `Next Due` → 2026-07-31, `Status` → `done`.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `## Current Phase` header still points to Phase 1, which is 100% complete. Phase 2 is the active phase (custom item detection checked, Plan 41 shipped). | `Roadmap.md` header | Flagged for user — move "Current Phase" section to Phase 2 |
| 2 | Medium | Plan 41 (Headless CLI Contract + MCP-ready cores, shipped 2026-06-16, 5 phases, PRs #120/#122/#123/#121/#125) has no corresponding roadmap entry. It resolved the P1 backlog item for agent-drivable CLI parity. | `Roadmap.md` Phase 2 | Flagged for user — add a new `[x]` item to Phase 2 (e.g. "Headless CLI contract + agent-drivable cores") |
| 3 | Low | MCP server (Plan 42, described as "fast-follow Plan 41" in Status.md) has no roadmap entry. Phase 3 currently lists only "Managed Agents integration" and "Greenhouse", but an MCP server is a distinct near-term item. | `Roadmap.md` Phase 3 | Flagged for user — consider adding `[ ] MCP server` to Phase 3, or update Phase 3 description to clarify the MCP server is the entry point |
| 4 | Low | KeyDecisionLog.md Settled entry "Defer Managed Agents cloud integration until local foundation is stable" has not been formally revisited. Plan 41's explicit "MCP-ready cores" framing signals foundation stability. Phase 3 work is imminent. | `KeyDecisionLog.md` | Flagged for user — formally reopen and resolve the settled decision; it may need to move to Structural or be superseded |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[ROADMAP DRIFT — Medium]** Update `## Current Phase` in `Roadmap.md` to point to Phase 2 ("Extensibility"). Phase 1 is complete.
- **[MISSING ROADMAP ENTRY — Medium]** Add a checked Phase 2 item for Plan 41 (Headless CLI Contract + MCP-ready cores). Suggested text: `- [x] Headless CLI contract + agent-drivable cores — `*Result cores + JSONL/exit contract, agent-interface.md, `list --json`` (Plan 41, v0.5.x)`
- **[MISSING ROADMAP ENTRY — Low]** Add a Phase 3 item for MCP server (Plan 42, fast-follow). Suggested text: `- [ ] MCP server — expose Bonsai CLI operations as MCP tools for direct agent consumption`
- **[STALE SETTLED DECISION — Low]** Revisit `KeyDecisionLog.md` Settled entry re: Managed Agents deferral. Plan 41's "MCP-ready cores" framing makes the "foundation stability" condition clearly met. Formally acknowledge in the log.
- **[BOOKKEEPING — Low]** Plans 40 and 41 remain in `Plans/Active/` despite being shipped. Archive both (previously flagged in 2026-07-17 Memory Consolidation for Plan 41; Plan 40 Phase 4 was held, but Phases 1–3 shipped).

## Notes for Next Run

- Previous run's 2 flags (Phase 1 "Better trigger sections" checkbox, `bonsai validate` missing from Phase 1) were resolved by the 2026-05-07 Routine Digest — no recurrence.
- The "Current Phase" stale label is the most impactful finding. If the user resolves it before the next run, confirm Phase 2 is properly labeled as current.
- Plan 42 (MCP server) may have shipped by the next run (2026-07-31) — if so, verify it has a roadmap entry and check it off.
- If v0.5.0 is tagged before next run, verify Roadmap reflects the version milestone.
- Watch for Plan 40 Phase 4 (update-delivery) — if it ships, it may satisfy the `Self-update mechanism` Phase 2 item.
