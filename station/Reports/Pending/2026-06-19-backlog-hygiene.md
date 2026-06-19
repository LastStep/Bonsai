---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-06-19
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
- **Files Read:** 5 — `station/agent/Routines/backlog-hygiene.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write, Glob, Grep
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate Misplaced P0s
- **Action:** Read Backlog.md P0 section and checked each item against Status.md.
- **Result:** Found 2 resolved P0 items still in the Backlog:
  1. `[bug] Sensor hook commands use $PWD-walk-up` — Fixed in v0.4.3 (PRs #105/#106, 2026-05-13). Status.md confirms "v0.4.3 hotfix shipped."
  2. `[feature] bonsai init / bonsai add need non-interactive flags` — Fixed in v0.4.2 (PR #102, 2026-05-13). Status.md confirms "v0.4.2 release shipped."
  The third P0 — `[research] Trial sentrux` — was already commented out and promoted to Status.md Pending in the 2026-05-07 routine-digest. No action needed there.
- **Issues:** Both resolved P0s were still live Backlog bullets rather than HTML comments — resolved autonomously.

### Step 2: Cross-Reference with Status.md
- **Action:** Read Status.md In Progress, Pending, and Recently Done sections.
- **Result:**
  - In Progress: empty — no conflicts with Backlog.
  - Pending: only `[research] Trial sentrux` (already properly removed from Backlog P0 via comment).
  - Recently Done (since last hygiene run 2026-05-07): Plan 41 (headless CLI, v0.5.0 prep), Plan 40 (Odysseus integration phases 1-3), v0.4.3 hotfix, Plan 38 handoff, v0.4.2 release, PR triage sweep (9 stale bot PRs closed).
  - Removed 2 resolved P0 Backlog items that appeared in Status.md Recently Done (see Step 1).
  - The `[ops] Routine bot PR pile-up` P1 item: 9 stale PRs were closed (2026-05-07) but the root fix (direct-to-main or auto-merge) has not shipped — item remains valid, left in place.
  - No Status.md "Blocked By" items would be unblocked by resolving a Backlog item (sentrux is blocked on Rust toolchain, not a Backlog dependency).
- **Issues:** None beyond the resolved P0s above.

### Step 3: Cross-Reference with Roadmap.md
- **Action:** Read Roadmap.md and checked P2/P3 Backlog items for phase alignment.
- **Result:**
  - Phase 1 is fully complete (all boxes checked). No Backlog items reference Phase 1 work.
  - Phase 2 (Extensibility): Unchecked items are `Self-update mechanism`, `Template variables expansion`, `Micro-task fast path` — all have corresponding Backlog P3 entries appropriately filed.
  - Phase 3 (Cloud & Orchestration): `Managed Agents integration` and `Greenhouse companion app` are in Backlog P3 Big Bets — appropriate.
  - Phase 4 (Ecosystem): Future items — no near-term Backlog relevance.
  - No P2/P3 Backlog items found that align so closely with current-phase milestones that they warrant promotion to P1. Plan 41 shipped headless CLI (a prerequisite for Phase 3/Odysseus work), but the next natural P1 is the non-interactive CLI parity item already at P1.
  - No Backlog items reference deprecated approaches or completed phases.
- **Issues:** None.

### Step 4: Flag Stale Items
- **Action:** Scanned all Backlog items for age (>30 days without progress) and clarity issues.
- **Result:** Several items are significantly overdue for re-prioritization review:
  1. **P1 — `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder`** (added 2026-04-22, 58 days): The PAT was rotated 2026-04-22 with a reminder note for ~2026-07-15 to rotate again. Today is 2026-06-19 — the 2026-07-15 deadline is now **26 days away**. This is time-sensitive and requires user action before the next release attempt.
  2. **P1 — `[debt] Stale agent worktrees + branches accumulating`** (added 2026-04-20, 60 days): No evidence of any cleanup since filing. RoutineLog mentions the pattern was documented but no sweep has occurred.
  3. **P1 — `[debt] Testing infrastructure for triggers and sensors`** (added 2026-04-16, 64 days): No progress visible. The trigger system expansion that motivated this item has continued with Plans 35-41 shipping. Risk of invisible regressions grows.
  4. **P1 — `[ops] Routine bot PR pile-up`** (added 2026-05-07, 43 days): The 9 PRs were closed but the structural fix (commit-direct-to-main, auto-merge, or skip-on-absorbed-range) has not landed.
  5. **P2 — Group A `[bookkeeping] Retroactively trim Backlog entries to NoteStandards`** (added 2026-04-25, 55 days): Item explicitly deferred but entries continue to grow verbose. NoteStandards rule is well-defined — this sweep keeps getting bypassed.
  6. Near-duplicate check: No new near-duplicates identified. Existing known overlap (`[feature] Changelog generation skill` vs `[feature] Full agent-drivable CLI parity` being the main narrative) is tracked and intentional.
- **Issues:** PAT expiry is the highest-urgency finding — flagged for user review.

### Step 5: Check for Routine-Generated Items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07).
- **Result:** No routine executions have occurred since 2026-05-07. The only post-May-7 RoutineLog entry is the `2026-06-13 — Plan 40 dispatch` session log (not a routine). `Reports/Pending/` is empty — all prior cycle reports were archived. No uncaptured routine findings require Backlog entries.
- **Issues:** None. Note: It has been 43 days since any routine ran (all are overdue). The routine dashboard still shows May dates as Last Ran with status "done" — these are all stale and overdue by several weeks.

### Step 6: Promote Ready Items via Issue-to-Implementation
- **Action:** Assessed whether any Backlog items are approved for immediate implementation.
- **Result:** No user approval was given for any item. No P0s remain after the cleanup. The clearest candidate for promotion is `[feature] Full agent-drivable non-interactive CLI parity: init/update/add/remove` (P1, added 2026-06-13, noted as "the main thing" by user), but promotion requires user confirmation per the routine procedure. Flagging this for user decision.
- **Issues:** None — no autonomous promotions made.

### Step 7: Log Results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Entry written.
- **Issues:** None.

### Step 8: Update Dashboard
- **Action:** Updated `station/agent/Core/routines.md` Backlog Hygiene row.
- **Result:** `Last Ran` → 2026-06-19, `Next Due` → 2026-06-26, `Status` → done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | `[bug] Sensor hook commands $PWD-walk-up` P0 entry still live despite fix shipping in v0.4.3 (2026-05-13) | `Backlog.md P0` | Converted to HTML comment with resolution note |
| 2 | high | `[feature] Non-interactive flags` P0 entry still live despite fix shipping in v0.4.2 (2026-05-13) | `Backlog.md P0` | Converted to HTML comment with resolution note |
| 3 | high | HOMEBREW_TAP_TOKEN PAT rotation deadline is 26 days away (2026-07-15) | `Backlog.md P1` | Flagged for user — no autonomous action possible |
| 4 | medium | All routines overdue — no routine has run since 2026-05-07 (43 days) | `routines.md dashboard` | Flagged for user — dashboard Last Ran dates all stale |
| 5 | medium | `[debt] Stale agent worktrees + branches` (P1) — 60 days old, no cleanup | `Backlog.md P1` | Flagged for user review |
| 6 | medium | `[debt] Testing infrastructure for triggers and sensors` (P1) — 64 days, no progress; risk grows as codebase expands | `Backlog.md P1` | Flagged for user review |
| 7 | low | `[ops] Routine bot PR pile-up` (P1) — structural fix unshipped 43 days after filing | `Backlog.md P1` | Flagged for user review |
| 8 | info | `[feature] Full agent-drivable CLI parity` (P1) — user flagged as "main thing" 2026-06-13, no plan started | `Backlog.md P1` | Flagged for user — awaiting confirmation to start via issue-to-implementation |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **PAT expiry (high-urgency):** `HOMEBREW_TAP_TOKEN` PAT was rotated 2026-04-22 with a ~2026-07-15 rotation reminder. That deadline is now 26 days away. Rotate the PAT in GitHub secrets before the next release attempt to avoid a brew step 401 failure.

2. **All routines overdue (43 days):** Every routine on the dashboard has a "Next Due" date in May. Consider scheduling a routine-digest run or dispatching individual routines. At minimum: Dependency Audit, Doc Freshness Check, Vulnerability Scan, and Status Hygiene are all overdue by 5+ weeks.

3. **`[feature] Full agent-drivable CLI parity`** is noted as "the main thing" by the user (added 2026-06-13). Confirm whether to start this via `/issue-to-implementation` now — it has a clear scope (audit init/update/add/remove for non-interactive completeness, then unified surface + JSONL/exit-code contract).

4. **P1 stale debt items:** `[debt] Stale agent worktrees + branches` (60 days) and `[debt] Testing infrastructure for triggers and sensors` (64 days) have had no visible progress. Consider deferring to P2 if not targeted in the next 1-2 plans, or confirm they remain P1.

5. **`[ops] Routine bot PR pile-up`:** The structural fix (commit-direct-to-main or auto-merge) has not shipped. The 9 stale PRs were cleared, but the symptom will recur unless the root cause is addressed.

## Notes for Next Run

- P0 section is now clean — all items resolved or properly promoted. The next hygiene run should check that the sentrux trial in Status.md Pending has either progressed or been deferred.
- PAT expiry (2026-07-15) will be urgent by the next run (2026-06-26). Follow up immediately if not already rotated.
- The Backlog has grown significantly in scope since NoteStandards was introduced — Group A bookkeeping sweep (`[bookkeeping] Retroactively trim Backlog entries`) remains deferred but increasingly valuable. Consider bundling with a future Plan as a prep step.
- No pending routine reports need backlog items — all prior cycles were fully digested.
