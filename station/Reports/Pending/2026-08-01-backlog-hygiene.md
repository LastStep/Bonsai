---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-01
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07 (value from dashboard before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 5 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 1 — `/home/user/Bonsai/station/Playbook/Backlog.md` (3 resolved items converted to HTML comments)
- **Tools Used:** Read, Edit, Write, Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; cross-referenced each item against Status.md In Progress and Pending tables.
- **Result:** Found 2 P0 items that were already SHIPPED and sitting stale in the Backlog:
  1. `[bug] Sensor hook commands use $PWD-walk-up` — Fixed in v0.4.3 (PR #105/#106, 2026-05-13). Status.md "Recently Done" row confirms ship date.
  2. `[feature] bonsai init / bonsai add need non-interactive flags [Plan 38 P2 blocker]` — Fixed in v0.4.2 (PR #102, 2026-05-13). Status.md "v0.4.2 release shipped" row confirms ship date.
  Both items were converted to HTML audit-trail comments in Backlog.md. P0 section is now empty.
- **Issues:** None — both removals are clean.

### Step 2: Cross-reference with Status.md
- **Action:** Reviewed In Progress (empty), Pending (sentrux trial only), and Recently Done tables; looked for Backlog items already resolved by shipped work.
- **Result:**
  - Found 1 additional resolved P1 item: `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` — fully resolved by Plan 41 (PRs #120/#122/#123/#121/#125, 2026-06-16). Status.md confirms all four mutating commands now have headless cores + JSONL/exit contract. Item converted to HTML comment.
  - Pending item (Trial sentrux) is correctly in Status.md Pending; no Backlog overlap.
  - No "Blocked By" items in Status.md that could be unblocked by a Backlog entry (sentrux is blocked on Rust toolchain, not a Backlog item).
  - **Incidental finding:** Plans 40 and 41 are still in `Plans/Active/` despite both being marked "Recently Done" in Status.md. This is a status-hygiene item, not backlog-hygiene — flagged for the next Status Hygiene run.
- **Issues:** None blocking.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; compared Phase 1 completion state and Phase 2 milestones against Backlog priorities.
- **Result:**
  - Phase 1 is fully complete (all checkboxes checked, including Plan 35 `bonsai validate` added in the 2026-05-07 routine-digest quick fix).
  - Phase 2 milestone items that have Backlog entries: `[improvement] Self-update mechanism` (P3 Big Bets) and `[improvement] Micro-task fast path` (P3 Future Platform) both align with Phase 2 goals. Neither warrants immediate promotion to P1 without user direction — flagging as alignment candidates.
  - No Backlog items found that reference deprecated approaches or completed phases (the two P0 stale items removed in Step 1/2 were the main offenders).
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Scanned all P1 items for entries at the same priority 30+ days without progress, and for entries lacking clear rationale.
- **Result:** All six non-removed P1 items were added between 2026-04-16 and 2026-05-07 — between 85 and 107 days ago. No progress markers.

  Critical stale item:
  - **`[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder`** (added 2026-04-22) — The reminder date was ~2026-07-15. Today is 2026-08-01. The PAT is **17 days past its rotation deadline**. Next release will fail the Homebrew tap step with 401 if not rotated. Flagging for immediate user action.

  Standard stale items (85-107 days, no movement):
  - `[ops] Routine bot PR pile-up` (added 2026-05-07) — Still P1, no resolution committed.
  - `[debt] Testing infrastructure for triggers and sensors` (added 2026-04-16, 107 days) — Group B, no plan assigned.
  - `[debt] Stale agent worktrees + branches accumulating` (added 2026-04-20, 103 days) — No cleanup run documented.

  No entries found with no clear rationale — all items have adequate context.
  No near-duplicates found across priority tiers.
- **Issues:** PAT expiry is urgent.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md for all entries since last backlog-hygiene run (2026-05-07). Checked if any routine findings were uncaptured in the Backlog.
- **Result:** The only RoutineLog entry after 2026-05-07 is the 2026-06-13 Plan 40 dispatch entry (not a routine). No routine has run in 85-88 days:
  - Dependency Audit: last ran 2026-05-04 (88 days overdue)
  - Vulnerability Scan: last ran 2026-05-04 (88 days overdue)
  - Doc Freshness Check: last ran 2026-05-04 (88 days overdue)
  - Memory Consolidation: last ran 2026-05-07 (85 days overdue)
  - Roadmap Accuracy: last ran 2026-05-07 (85 days overdue)
  - Status Hygiene: last ran 2026-05-07 (85 days overdue)

  No uncaptured findings to add — the 2026-05-07 Routine Digest processed all prior flags cleanly (false positives confirmed, quick fixes applied). No new auto-add warranted.
- **Issues:** All other routines are massively overdue — flagged for user review. This routine batch needs a dedicated digest session.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Checked if any item warranted route-through the issue-to-implementation workflow.
- **Result:** No items are in a "user-approved for implementation" state. PAT expiry (Step 4) is a user action, not an implementation task. No promotion initiated.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` — Backlog Hygiene row Last Ran → 2026-08-01, Next Due → 2026-08-08, Status → done.
- **Result:** Done.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 bug "[bug] Sensor hook commands use $PWD-walk-up" was shipped in v0.4.3 (2026-05-13) but still in Backlog P0 | Backlog.md P0 | Converted to HTML comment — removed from active backlog |
| 2 | resolved | P0 feature "[feature] bonsai init/add need non-interactive flags" was shipped in v0.4.2 (2026-05-13) but still in Backlog P0 | Backlog.md P0 | Converted to HTML comment — removed from active backlog |
| 3 | resolved | P1 feature "[feature] Full agent-drivable CLI parity" was shipped by Plan 41 (2026-06-16) but still in Backlog P1 | Backlog.md P1 | Converted to HTML comment — removed from active backlog |
| 4 | **URGENT** | HOMEBREW_TAP_TOKEN PAT was due for rotation ~2026-07-15 — now 17 days past expiry date | Backlog.md P1 | Flagged for immediate user action — PAT needs rotation NOW |
| 5 | medium | 3 P1 items at same priority 85-107 days without progress (bot PR pile-up, testing infra, stale worktrees) | Backlog.md P1 | Flagged for user re-prioritization |
| 6 | medium | All 6 other routines overdue by 85-88 days (last ran 2026-05-04/07) | routines.md dashboard | Flagged for immediate routine catch-up session |
| 7 | low | Plans 40 and 41 still in Plans/Active/ despite being in Status.md Recently Done | Plans/Active/ | Flagged — Status Hygiene item, not in scope for this routine |
| 8 | info | P3/Phase 2 alignment: Self-update mechanism + Micro-task fast path in P3 align with current Roadmap Phase 2 goals | Backlog.md P3 | Noted — consider promotion when Phase 2 work begins |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **URGENT — Rotate HOMEBREW_TAP_TOKEN PAT immediately.** The fine-grained PAT on `LastStep/Bonsai` was due for rotation ~2026-07-15 (90-day expiry from 2026-04-22 rotation). Today is 2026-08-01 — it is expired or near-expired. Next release will fail at the Homebrew tap step with `401 Bad credentials`. Action: rotate PAT at GitHub → Settings → Developer Settings → Fine-grained tokens, update `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai`. Consider the backlog item at that point resolved.

2. **Routine catch-up session needed.** All six other routines are 85-88 days overdue. A routine-digest session should process at minimum: Dependency Audit, Vulnerability Scan, Doc Freshness Check (most operationally impactful after 88 days).

3. **Re-prioritize 3 stale P1 items** (testing infra, stale worktrees, routine bot PR pile-up). Each is 85-107 days at P1 without a plan assigned. Decide: assign to a plan sprint, demote to P2, or close with rationale.

4. **Archive Plans 40 and 41** from `Plans/Active/` to `Plans/Archive/` (Status Hygiene item; can be handled during the next Status Hygiene run or in-session).

---

## Notes for Next Run

- P0 section is now empty. If the next run finds a P0, it should be escalated to Status.md immediately.
- All other routines need to run before the next backlog-hygiene — their findings may generate additional Backlog items.
- The stale P1 items have been flagged twice across routine runs (2026-05-07 and now). If they remain at P1 after the next run without a plan, consider an automatic demotion to P2.
- No Backlog entries were auto-added this run — all uncaptured findings from prior routines were previously handled by the 2026-05-07 Routine Digest.
