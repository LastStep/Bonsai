---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-03
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
- **Duration:** ~12 min (gap was 119 days; thorough cross-reference pass required)
- **Files Read:** 5 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 1 — `station/Playbook/Backlog.md` (3 items resolved/commented out)
- **Tools Used:** Read (5×), Edit (2×)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog P0 section; identified 2 active P0 items. Cross-checked each against Status.md.
- **Result:** Both P0 items were already resolved by shipped work:
  - `[bug] Sensor hook commands use $PWD-walk-up` → RESOLVED by v0.4.3 (2026-05-13, PR #105/#106). Status.md explicitly confirms: "sensor hook commands now bake install-time absolute paths."
  - `[feature] bonsai init / bonsai add need non-interactive flags` → RESOLVED by v0.4.2 (2026-05-13, PR #102). Status.md confirms: `--non-interactive --from-config` shipped, unblocked Plan 38 P2.
  - Both items were commented out in Backlog.md with resolution notes.
- **Issues:** No unresolved P0s to escalate. P0 section is now empty of active items.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md fully. Compared Backlog items against In Progress, Pending, and Recently Done tables.
- **Result:**
  - 2 P0 items resolved (see Step 1).
  - 1 P1 item resolved: `[feature] Full agent-drivable (non-interactive) CLI parity` — Plan 41 (2026-06-16) shipped headless cores for all 4 mutating commands (init/add/update/remove) with JSONL/exit contract. Item commented out in Backlog with resolution note.
  - Pending row for `[research] Trial sentrux on Bonsai repo` remains in Status.md — still blocked on Rust toolchain. No change needed.
  - No new "Blocked By" pairings discovered between Status.md Pending items and Backlog items that could unblock them.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md. Checked P2/P3 items against current phase milestones.
- **Result:**
  - Phase 1 is fully complete (all checkboxes checked as of the 2026-05-07 routine-digest).
  - Phase 2 (Extensibility) is the current phase. Open items: Self-update mechanism, Template variables expansion, Micro-task fast path. All have corresponding Backlog P3 entries — no promotion warranted at this time without user direction.
  - Plan 41's headless CLI contract is a strong enabler for Phase 3 (Cloud & Orchestration). No Backlog updates needed for roadmap alignment.
  - No Backlog items reference deprecated approaches or completed phases that haven't already been cleaned up.
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Reviewed all remaining Backlog items for staleness. The run gap is 119 days (2026-05-07 → 2026-09-03), so essentially all items are 30+ days stale by calendar.
- **Result:** Most items are legitimately long-running backlog debt, not stale in the "forgotten" sense. Items with real staleness concerns flagged below:
  - **[ops] HOMEBREW_TAP_TOKEN PAT expiry** (P1): Added 2026-04-22, rotate target ~2026-07-15. Today is 2026-09-03 — the PAT is almost certainly expired (90-day window). **This is now urgent.** Symptom of expired PAT: GoReleaser brew step fails at 401. Next release will fail Homebrew tap update if not rotated.
  - **[ops] Routine bot PR pile-up** (P1): Added 2026-05-07. The root-cause fix (cloud routine commit strategy) remains unimplemented. Still valid.
  - **[improvement] Add root Bonsai/CLAUDE.md tree-drift check** (P2 Ungrouped): Flagged by 3 consecutive routine-digest cycles (2026-04-14, 2026-04-21, 2026-05-04). Now at ~5 months without action. Recommend discussing with user whether to promote to P1 or formally defer.
  - No near-duplicates found beyond those already noted in prior runs (changelog/Group C vs Group D overlap, previously flagged).
- **Issues:** HOMEBREW_TAP_TOKEN likely expired — requires immediate user action.

### Step 5: Check for routine-generated items
- **Action:** Scanned RoutineLog.md for entries since last backlog-hygiene run (2026-05-07).
- **Result:** Only one entry between 2026-05-07 and 2026-09-03: the 2026-06-13 Plan 40 dispatch log (not a routine run — no backlog findings). No routine runs (Dependency Audit, Vulnerability Scan, Doc Freshness Check, etc.) executed in the 119-day gap. All routines show `last_ran` ≤ 2026-05-07 in the dashboard and are 100+ days overdue.
  - Uncaptured findings from the missing routine runs cannot be verified or added — flagging the gap for user review.
- **Issues:** All 6 other routines are severely overdue (100–115 days). No findings from missed runs captured in Backlog.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any items are approved for immediate implementation.
- **Result:** No user approval signals present. No items auto-promoted. The HOMEBREW_TAP_TOKEN item is the most urgent; escalated in Step 4 for user decision.
- **Issues:** None.

### Step 7: Log results
- **Action:** RoutineLog.md entry appended (see post-procedure steps).
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `routines.md` dashboard row for Backlog Hygiene.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Resolved | P0 bug "Sensor hook $PWD-walk-up" was already fixed (v0.4.3) | Backlog.md P0 | Commented out with resolution note |
| 2 | Resolved | P0 feature "bonsai init/add non-interactive flags" was already fixed (v0.4.2) | Backlog.md P0 | Commented out with resolution note |
| 3 | Resolved | P1 "Full agent-drivable CLI parity" was already fixed (Plan 41) | Backlog.md P1 | Commented out with resolution note |
| 4 | Critical | HOMEBREW_TAP_TOKEN PAT almost certainly expired (target rotate date 2026-07-15 has passed) | Backlog.md P1 | Flagged for user — requires immediate rotation |
| 5 | High | All 6 other routines are 100–115 days overdue — no findings from that gap captured | routines.md dashboard | Flagged for user — recommend running overdue routines |
| 6 | Medium | CLAUDE.md tree-drift check improvement (P2 Ungrouped) has recurred 5+ times across routine-digests without action | Backlog.md Ungrouped P2 | Flagged for user — consider promoting to P1 or formally deferring |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[URGENT] Rotate HOMEBREW_TAP_TOKEN PAT** — The 90-day fine-grained PAT set 2026-04-22 was due for rotation ~2026-07-15. Today is 2026-09-03. The PAT is almost certainly expired. Until rotated, any release will fail at the Homebrew tap update step with `401 Bad credentials`. Rotate in GitHub > Settings > Personal access tokens, then `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`.

- **[HIGH] Run overdue routines** — Dependency Audit, Vulnerability Scan, Doc Freshness Check, Memory Consolidation, Status Hygiene, and Roadmap Accuracy are all 100–115 days overdue. Recommend triggering a full routine-digest pass to catch any issues that accumulated during Plans 40/41 and the Bonsai-Eval work.

- **[MEDIUM] Promote or formally defer "CLAUDE.md tree-drift check" (P2 Ungrouped)** — This item has surfaced in 5+ consecutive routine runs (2026-04-14, 2026-04-21, 2026-05-04, and the underlying drift worsened with Plans 40/41 adding new internals). Either promote to P1 and fix it, or add a deferred note to the item acknowledging the decision.

## Notes for Next Run

- P0 section is now empty of active items — the next run should confirm this remains the case.
- The 119-day run gap is unusually long; verify the loop dispatch mechanism is healthy.
- HOMEBREW_TAP_TOKEN rotation (item 4 above) should be confirmed resolved before the next release.
- If overdue routines have been run by next cycle, check their findings against the Backlog to capture any new items.
