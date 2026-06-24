---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-06-24
status: success
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07 (previous value from dashboard)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~5 min
- **Files Read:** 4 — `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/41-headless-cli-contract.md`
- **Files Modified:** 3 — `/home/user/Bonsai/station/Reports/Pending/2026-06-24-roadmap-accuracy.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state
- **Action:** Read `Roadmap.md` and cross-checked all phase items against `Status.md` recently done entries and active plans.
- **Result:** Phase 1 is fully complete and accurately marked. Since the last run (2026-05-07), the 2026-05-07 Routine Digest had already applied two quick fixes (checked "Better trigger sections" and added `bonsai validate` row to Phase 1). Both of those are now correctly reflected. Phase 2 has `Custom item detection` accurately marked `[x]`. Phase 3 and 4 unchecked items are all still outstanding — no false marks found.
- **Issues:** Plan 41 (Headless CLI Contract + MCP-ready cores), shipped 2026-06-16, has no roadmap representation despite being a significant Phase 2/Phase 3-bridge milestone.

### Step 2: Check milestone accuracy
- **Action:** Evaluated whether next milestones in Phase 2 (`Self-update mechanism`, `Template variables expansion`, `Micro-task fast path`) remain the right priorities and whether any planned work has been superseded.
- **Result:** Those three items remain unbuilt and the user has not indicated any change in priority. No roadmap items reference deprecated approaches. Plan 41 introduced the headless/MCP-ready substrate that is the explicit foundation for a Phase 3 MCP server (Plan 42, described as "fast-follow" in Plan 41) — but Plan 42 does not appear on the roadmap yet.
- **Issues:** Phase 3 currently lists only `Managed Agents integration` and `Greenhouse companion app`. The MCP server path (Plan 42) is architecturally distinct from the Managed Agents integration and should likely be its own roadmap item once scoped.

### Step 3: Cross-check against Key Decision Log
- **Action:** Read `KeyDecisionLog.md` in full, checked all entries dated since the last run (2026-05-07) for any that would invalidate roadmap items.
- **Result:** No new entries in the KeyDecisionLog since 2026-04-13. The standing "Defer Managed Agents cloud integration until local foundation is stable" decision remains in effect and is consistent with Phase 3 items still being unchecked. No roadmap items are invalidated by any logged decision.
- **Issues:** None.

### Step 4: Report findings
- **Action:** Documenting two findings below. Not modifying `Roadmap.md` directly — flagging for user review.
- **Result:** 2 findings identified: 1 low-severity gap (Plan 41 headless CLI unrepresented), 1 informational item (Plan 42 MCP server not yet on roadmap).

### Step 5: Update dashboard
- **Action:** Will update `agent/Core/routines.md` dashboard row for Roadmap Accuracy.
- **Result:** See dashboard update section below.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Low | Plan 41 (Headless CLI Contract + MCP-ready cores, shipped 2026-06-16) has no roadmap representation. It is a significant Phase 2 milestone: all mutating commands have headless pure-Result cores, structured JSONL/JSON output, and a documented exit-code contract. The roadmap reads as though the only Phase 2 completed item is custom item detection — understating actual progress. | `Roadmap.md` Phase 2 | Flagged for user review — recommend adding a `[x] Headless CLI contract — all mutating commands run non-interactively; structured JSON/JSONL output + exit-code contract. MCP-ready cores (Plan 41, v0.5.0+)` row to Phase 2. |
| 2 | Info | Plan 42 (MCP server — `bonsai mcp`) is described as a "fast-follow" to Plan 41 in the active plan file but is not listed in the roadmap. If actively planned, it warrants a Phase 3 row distinct from `Managed Agents integration`. | `Roadmap.md` Phase 3 / `Plans/Active/41-headless-cli-contract.md` | Flagged for user to decide whether to add a Phase 3 row or keep it off-roadmap until the plan is scoped. |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[Low] Roadmap Phase 2 — add `[x]` row for Plan 41 Headless CLI Contract.** Suggested text:
   ```
   - [x] Headless CLI contract — all mutating commands (`init`, `add`, `update`, `remove`) run non-interactively with pure `Result` cores; structured JSONL + JSON output + documented exit-code contract; MCP-ready substrate (Plan 41, v0.5.0+)
   ```

2. **[Info] Roadmap Phase 3 — consider adding a row for Plan 42 MCP server** once scoped. Current Phase 3 `Managed Agents` item refers to a managed cloud platform integration; Plan 42 is a local MCP server wrapper over the headless cores — architecturally distinct. No action required until Plan 42 is formally scoped.

## Notes for Next Run

- Phase 1 is fully settled — no need to re-audit those items.
- Phase 2 has one legitimately completed item (custom item detection) and potentially two after user applies the Plan 41 addition.
- Status.md confirms no currently in-progress work (In Progress table is empty); Pending has only the sentrux research item blocked on Rust toolchain.
- KeyDecisionLog has had no new entries since 2026-04-13 — if that continues, this cross-check step can be brief.
