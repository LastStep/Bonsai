---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-15
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 6 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/agent/Core/memory.md`
- **Files Modified:** 3 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; checked each P0 against Status.md In Progress and Pending tables.
- **Result:** 2 active P0 items found. Neither appears in Status.md as In Progress or Pending — both appear in Status.md Recently Done (resolved). No truly misplaced P0s requiring immediate promotion; both warranted removal (Step 2).
- **Issues:** none

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md; matched all Backlog items against In Progress, Pending, and Recently Done entries.
- **Result:** 2 Backlog P0 items resolved and removed:
  1. `[bug] Sensor hook commands use $PWD-walk-up` — resolved by v0.4.3 hotfix (2026-05-13). Replaced with resolution comment.
  2. `[feature] bonsai init/add need non-interactive flags` — resolved by v0.4.2 (2026-05-13). Replaced with resolution comment noting that broader CLI parity (update/remove) is tracked at P1.
- No Pending items with "Blocked By" that could be unblocked by a Backlog item (sentrux blocked on Rust toolchain install — no Backlog item directly addresses that).
- **Issues:** none

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; tagged P2/P3 items by phase alignment.
- **Result:** Phase 1 is complete (all checkboxes checked, including bonsai validate added per 2026-05-07 routine-digest). No items reference deprecated Phase 1 approaches. Phase 2 alignment: Backlog P3 "[improvement] Self-update mechanism" and "[improvement] Micro-task fast path" both map directly to Phase 2 milestones. Both remain P3 — no promotion recommended without user direction. Phase 3 Big Bets (Managed Agents, Greenhouse) align with Phase 3 as expected.
- **Issues:** none

### Step 4: Flag stale items
- **Action:** Reviewed all items for 30+ day stagnation; checked for unclear rationale and near-duplicates.
- **Result:**
  - **CRITICAL — P1 HOMEBREW_TAP_TOKEN PAT:** Added 2026-04-22, rotation due ~2026-07-15. Today is 2026-09-15 — 2 months past due. Item updated with OVERDUE flag. This warrants immediate user action before any release attempt.
  - **P1 — Full agent-drivable CLI parity (init/update/add/remove):** Added 2026-06-13, marked "main thing" by user. No movement in 3 months. Not stale-in-priority (P1 is correct), but surfacing for visibility.
  - All P2/P3 items are 3-5 months old. This is an extended gap (routines haven't run since May 2026). No items lack rationale. No near-duplicates found beyond the previously documented P0/P1 overlap that was already resolved.
  - Group F items ([docs] AltScreen behavior, [docs] Deviations from Plan) remain valid P2 items despite the Group F phase being "essentially closed" — they are standalone documentation improvements, not phase-dependent.
- **Issues:** PAT expiry overdue — user action required.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since 2026-05-07.
- **Result:** No routine runs occurred between 2026-05-07 and 2026-09-15 (all routines severely overdue — 4+ months). The 2026-06-13 RoutineLog entry is a session/dispatch log (Plan 40), not a routine result. Findings from that dispatch (website npm vuln, unify remove logic) were already captured in Backlog P2 at time of writing. No uncaptured routine-flagged items.
- **Issues:** All routines are 4+ months overdue — routine-check sensor should have flagged this at session start repeatedly. This is a systemic gap (no active sessions in the gap period).

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any item is approved for immediate implementation.
- **Result:** No items have explicit user approval for immediate implementation noted in context. The P1 "Full agent-drivable CLI parity" is the highest-priority candidate but requires user confirmation. Not routing to issue-to-implementation without approval.
- **Issues:** none

### Step 7: Log results
- **Action:** Appended entry to RoutineLog.md.
- **Result:** Done.

### Step 8: Update dashboard
- **Action:** Updated routines.md dashboard row for Backlog Hygiene.
- **Result:** last_ran → 2026-09-15, next_due → 2026-09-22, status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | P0 bug item (sensor hook $PWD-walk-up) still in Backlog; resolved by v0.4.3 | Backlog.md P0 | Removed (replaced with resolution comment) |
| 2 | High | P0 feature item (non-interactive flags) still in Backlog; resolved by v0.4.2 | Backlog.md P0 | Removed (replaced with resolution comment) |
| 3 | High | HOMEBREW_TAP_TOKEN PAT rotation overdue by ~2 months (due 2026-07-15) | Backlog.md P1 | Updated item with OVERDUE flag; flagged for user |
| 4 | Medium | P1 "Full agent-drivable CLI parity" — no movement in 3 months despite user marking as "main thing" | Backlog.md P1 | Surfaced for user visibility; no edit (correct priority) |
| 5 | Low | All routines 4+ months overdue — extended maintenance gap | routines.md dashboard | Noted; this run begins catch-up |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[URGENT] Rotate HOMEBREW_TAP_TOKEN PAT immediately.** The fine-grained PAT was due ~2026-07-15 and is now ~2 months overdue. Before the next `goreleaser` release run, rotate the PAT and update the `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai`. Symptom if expired: GoReleaser brew step fails with `401 Bad credentials`; release binaries publish but Homebrew formula update is missed.

- **[DECISION] Prioritize P1 — Full agent-drivable CLI parity (init/update/add/remove).** User previously called this "the main thing" (2026-06-13). No movement in 3 months. Ready to plan and dispatch when capacity opens — consider `/issue-to-implementation` or `/planning`.

- **[FYI] All 7 routines are severely overdue** (4+ months, last run May 2026). The following are most time-sensitive after this one: Vulnerability Scan, Dependency Audit, Memory Consolidation. Consider a routine-digest session to process pending reports from this catch-up batch.

- **[FYI] Plan 41 file still in Plans/Active/** — memory.md Work State notes it should be archived to Plans/Archive/ at next wrap-up. Out of scope for this routine but flagging for visibility.

## Notes for Next Run

- P0 section is now clear of active items (both items were resolved and removed). If a new P0 surfaces, verify it appears in Status.md within the same routine cycle.
- The 4-month gap between this run and the previous (2026-05-07 → 2026-09-15) means some items may have been resolved in the interim without being cleaned up. A secondary cross-reference pass against git log or PR history would catch additional stale items — deferred to user judgement given scope of this routine.
- Watch HOMEBREW_TAP_TOKEN PAT going forward — add a recurring calendar note with a 2-week lead (rotate by 2026-12-14 if a 90-day PAT is set today).
