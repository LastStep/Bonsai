---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-13
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07 (previous value from dashboard before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 5
  - `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`
  - `/home/user/Bonsai/station/Playbook/Backlog.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 3
  - `/home/user/Bonsai/station/Playbook/Backlog.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section; cross-referenced each item against `Status.md` In Progress and Pending tables.
- **Result:** Two active P0 items found. Neither appeared in Status.md — however, both have been resolved and shipped (confirmed via Status.md Recently Done):
  - `[bug] Sensor hook commands use $PWD-walk-up` — resolved in v0.4.3 (PRs #105/#106, 2026-05-13).
  - `[feature] bonsai init / bonsai add need non-interactive flags` — resolved in v0.4.2 (PR #102, 2026-05-13).
  - These were added after the last backlog-hygiene run (2026-05-07) and had never been reviewed by this routine.
- **Issues:** Both P0 items were already resolved — they should not have been in P0 at all. Commented out with audit trail.

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md`. Scanned In Progress, Pending, and Recently Done for matches to Backlog items.
- **Result:**
  - **In Progress:** empty — no overlap.
  - **Pending:** `[research] Trial sentrux` — already a commented-out HTML entry in Backlog P0. Correct state.
  - **Recently Done (matching Backlog items):**
    - Plan 41 (2026-06-16) shipped full headless CLI parity for all four commands — resolves P1 `[feature] Full agent-drivable (non-interactive) CLI parity`. Commented out with audit trail.
    - v0.4.3 hotfix (2026-05-13) — resolves P0 sensor hook bug (handled in Step 1).
    - v0.4.2 (2026-05-13) — resolves P0 non-interactive flags (handled in Step 1).
  - No Status.md Pending items with "Blocked By" are unblocked by a Backlog item.
- **Issues:** None beyond the resolved items already actioned.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md`. Checked P2/P3 Backlog items against current-phase milestones. Checked for deprecated-approach references.
- **Result:**
  - Phase 1 is complete — all checkboxes ticked (including `bonsai validate` row and "Better trigger sections" annotation added by 2026-05-07 digest).
  - Phase 2 items `self-update mechanism` and `micro-task fast path` exist in P3 Backlog (Big Bets / Future Platform). Phase 2 is the "next up" phase but has no active plan. No promotion warranted autonomously; these remain correctly categorized as P3.
  - `[docs] Document AltScreen behavior change in release notes` (Group F) — original context was Plan 15 (shipped April 2026); v0.4, v0.4.3, v0.5.0 have since shipped with no mention. This item is 87+ days old and references a behavioral change from 3+ versions ago. Flagged as likely stale for user review.
  - No Backlog items reference deprecated approaches or completed Phase 1 line items other than those already resolved above.
- **Issues:** One stale Group F item flagged for user.

### Step 4: Flag stale items
- **Action:** Scanned all items for 30+ days at same priority without progress, items with no clear context, and near-duplicates.
- **Result:**
  - Most P2/P3 items are 80–90 days old (added April 2026). P2/P3 aging without progress is acceptable; no escalation warranted.
  - **URGENT FLAG:** `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` — notes PAT rotated 2026-04-22, reminder for ~2026-07-15. Today is 2026-07-13 — this deadline is 2 days away. The PAT must be rotated before the next release or the Homebrew tap step will fail with 401. This requires immediate user action.
  - Near-duplicates: `[improvement] Plans Index file` and `[improvement] Plan archiving` are already noted as related/sub-tasks in the text. No change needed.
  - `[feature] Changelog generation skill + release changelogs` and the Group A "CHANGELOG.md" sub-item reference the same gap. Both are clearly scoped; no de-duplication needed.
- **Issues:** 1 urgent PAT expiry flag; 1 stale Group F doc item flagged.

### Step 5: Check routine-generated items
- **Action:** Read `RoutineLog.md` entries since last backlog-hygiene run (2026-05-07).
- **Result:** No routine runs appear between 2026-05-07 and 2026-07-13 (log only shows the 2026-06-13 Plan 40 dispatch, which is a session log, not a routine). No pending reports exist in `Reports/Pending/`. No uncaptured routine findings to check.
- **Issues:** None.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed items for user-approval-to-implement signals.
- **Result:** No item has explicit user approval in Backlog text or Status.md. No promotion initiated (requires user confirmation). The HOMEBREW_TAP_TOKEN item is flagged as urgent but is an ops action (PAT rotation), not a code item — handled in flags below.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `Last Ran`, `Next Due`, and `Status` for Backlog Hygiene row in `agent/Core/routines.md`.
- **Result:** Done — 2026-07-13 / 2026-07-20 / done.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 sensor hook bug (`$PWD`-walk-up) already shipped in v0.4.3 | Backlog.md P0 | Commented out with audit trail |
| 2 | resolved | P0 non-interactive flags already shipped in v0.4.2 | Backlog.md P0 | Commented out with audit trail |
| 3 | resolved | P1 full agent-drivable CLI parity resolved by Plan 41 (2026-06-16) | Backlog.md P1 | Commented out with audit trail |
| 4 | **URGENT** | HOMEBREW_TAP_TOKEN PAT expires ~2026-07-15 (2 days) | Backlog.md P1 | Flagged for user — requires immediate PAT rotation |
| 5 | low | `[docs] Document AltScreen behavior change` (Group F) — 87+ days old, multiple releases since | Backlog.md Group F | Flagged for user — consider removing |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **URGENT — HOMEBREW_TAP_TOKEN PAT rotation due 2026-07-15 (2 days)**
   The `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai` was rotated 2026-04-22 with a ~90-day expiry. The reminder target is ~2026-07-15 — that is in 2 days. Rotate this PAT before the next release or the Homebrew formula step will fail with `401 Bad credentials`. Steps: create a new fine-grained PAT on GitHub with write access to `LastStep/homebrew-tap`, then run `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`.

2. **Stale Group F item — AltScreen release note**
   `[docs] Document AltScreen behavior change in release notes` was added during Plan 15 (April 2026). Three major versions have since shipped (v0.2, v0.3, v0.4, v0.4.3, v0.5) without this note. The window for a meaningful release note has likely passed. Consider removing this item.

## Notes for Next Run
- P0 section is now empty (all items resolved or promoted/commented). If any new critical issues surface they should be added directly.
- P1 section: `[ops] Routine bot PR pile-up` and `[ops] HOMEBREW_TAP_TOKEN PAT expiry` remain active. Verify PAT rotation was completed.
- Next routine gap since 2026-05-07 was 67 days — unusually long. Consider whether the loop dispatch schedule is healthy.
- No routine runs between May and July 2026 means the routine check sensor may not be firing reliably in the current setup.
