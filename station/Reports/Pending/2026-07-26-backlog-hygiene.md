---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-26
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
- **Tools Used:** Read (5×), Edit (3×)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section; checked each P0 item against `Status.md` In Progress and Pending.
- **Result:** Found 2 P0 items in Backlog. Both matched entries in Status.md Recently Done — neither is missing from Status.md, so no escalation needed. However, since both are resolved and shipped, they are candidates for removal (handled in Step 2).
  - `[bug] Sensor hook commands use $PWD-walk-up` → resolved in v0.4.3 (2026-05-13, PRs #105/#106)
  - `[feature] bonsai init / bonsai add need non-interactive flags` → resolved in v0.4.2 (2026-05-13, PR #102)
- **Issues:** None requiring escalation. Both P0s are done, not untracked.

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md`; compared Backlog items against In Progress, Pending, and Recently Done entries.
- **Result:** Found 3 Backlog items matching Status.md Recently Done — removed them (replaced with HTML comments for audit trail):
  1. P0 `[bug] Sensor hook commands use $PWD-walk-up` → shipped v0.4.3
  2. P0 `[feature] bonsai init/add non-interactive flags` → shipped v0.4.2
  3. P1 `[feature] Full agent-drivable CLI parity: init/update/add/remove` → shipped Plan 41 (2026-06-16, PRs #120–#125), every mutating command now has headless cores + JSONL/exit contract
- **Status.md Pending check:** One pending item — `[research] Trial sentrux on Bonsai repo` — blocked on Rust toolchain install. No Backlog item can directly unblock this; it requires the user to install rustup.
- **Issues:** Sentrux trial has been blocked and stale in Pending for ~80 days (since 2026-05-07). Flagged for user review.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md`; checked P2/P3 items against current phase milestones; checked for deprecated/completed phase references.
- **Result:**
  - Phase 1 fully complete (all checkboxes checked). No stale Phase 1 references found in Backlog.
  - Phase 2 "Self-update mechanism" → in Backlog P3 Big Bets. "Micro-task fast path" → in Backlog P3. Both appropriate.
  - Phase 2 "Template variables expansion" → NOT tracked in Backlog at all. Minor gap — worth noting.
  - Phase 3 (Managed Agents, Greenhouse app) → both in Backlog P3 Big Bets. Aligned.
  - Phase 4 items (Catalog marketplace, Plugin system) → not in Backlog; acceptable for far-future items.
  - No items referencing deprecated approaches or completed phases found (the now-removed P0/P1 items were the closest).
- **Issues:** "Template variables expansion" (Roadmap Phase 2 milestone) has no Backlog tracking entry. Low severity — flagged for user.

### Step 4: Flag stale items
- **Action:** Reviewed all items for age, clarity, and near-duplicates. Today's date is 2026-07-26; last hygiene run was 2026-05-07 (80 days gap).
- **Result:**
  - **HOMEBREW_TAP_TOKEN PAT:** P1 item added 2026-04-22 specified calendar reminder for ~2026-07-15. Today is 2026-07-26 — 11 days past the reminder date. The PAT was rotated 2026-04-22 on a 90-day default expiry, putting expiry around 2026-07-21. **The PAT has very likely expired.** Next release will fail at the Homebrew brew step with 401 Bad credentials if not rotated. Flagged as high priority for user.
  - Many P1/P2 items are 90-100+ days old (added 2026-04-16 to 2026-04-25) with no progress. This is expected for a tool under active development with shifting priorities. Items are clear and well-documented; no removal warranted without user input.
  - Near-duplicate check: "Plans Index file" (P2 Group E) and "Plan archiving" (P2 Group E) remain loosely related but are distinct tasks — no merge action taken.
  - All items have clear context and rationale. No items flagged for clarification or removal on content grounds alone.
- **Issues:** HOMEBREW PAT expiry is time-sensitive and flagged for immediate user action.

### Step 5: Check for routine-generated items
- **Action:** Read `station/Logs/RoutineLog.md`; looked for routine entries since last backlog-hygiene run (2026-05-07).
- **Result:** No routine executions logged between 2026-05-07 and today (2026-07-26) — an 80-day gap. The most recent entries are all from 2026-05-07. The 2026-06-13 entry is a session/plan dispatch log, not a routine. No routine flags to capture.
- **Issues:** The 80-day gap in routine execution means several routines (Dependency Audit, Vulnerability Scan, Doc Freshness, Memory Consolidation, Status Hygiene, Roadmap Accuracy) are all significantly overdue. These will likely surface issues when run. Flagged for user awareness.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any items are approved for immediate implementation.
- **Result:** No user-approved promotions in scope for this routine run. No autonomous promotion taken.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for Backlog Hygiene.
- **Result:** `Last Ran` → 2026-07-26, `Next Due` → 2026-08-02, `Status` → done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | HOMEBREW_TAP_TOKEN PAT likely expired (~2026-07-21 based on 90-day rotation from 2026-04-22) — next release will fail Homebrew step with 401 | Backlog P1 `[ops]` item | Flagged for user — rotate PAT via GitHub Settings immediately |
| 2 | medium | 3 Backlog items resolved and shipped but still listed as open | Backlog P0 (×2), P1 (×1) | Removed — replaced with HTML comments noting resolution date and PR/version |
| 3 | medium | Sentrux trial blocked in Status.md Pending for ~80 days (Rust toolchain not installed) | `Status.md` Pending | Flagged for user — decide to defer, install rustup, or drop |
| 4 | medium | No routine executions logged since 2026-05-07 (80-day gap) — 6 other routines all overdue | `Logs/RoutineLog.md` | Flagged for user — schedule a routine-digest session |
| 5 | low | Roadmap Phase 2 "Template variables expansion" has no Backlog tracking entry | `Roadmap.md` Phase 2 | Flagged for user — add to Backlog if desired, or accept as implicit |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[high] Rotate HOMEBREW_TAP_TOKEN PAT immediately.** The PAT was last rotated 2026-04-22 with a likely 90-day expiry. Today is 2026-07-26 — expiry was around 2026-07-21. Go to GitHub Settings → Developer Settings → Personal access tokens, rotate the `HOMEBREW_TAP_TOKEN` fine-grained PAT, and update the repo secret at `LastStep/Bonsai`. Symptom if missed: next GoReleaser release fails at the Homebrew formula push step with `401 Bad credentials`.

2. **[medium] Sentrux trial (Status.md Pending) stale for ~80 days.** Blocked on Rust toolchain not installed. Decide: (a) install rustup and run the trial this session, (b) defer to a later date, or (c) drop the item if sentrux is no longer a priority.

3. **[medium] All 6 other routines are significantly overdue** (Dependency Audit, Vulnerability Scan, Doc Freshness Check, Memory Consolidation, Status Hygiene, Roadmap Accuracy — all last ran 2026-05-04 or 2026-05-07). Consider scheduling a routine-digest session soon to clear the backlog.

4. **[low] Roadmap Phase 2 "Template variables expansion" not tracked in Backlog.** Add an entry if this is a planned feature, or leave it as a roadmap-only milestone.

## Notes for Next Run

- All 3 resolved items are now HTML-commented out in Backlog with resolution notes — do not re-add them.
- HOMEBREW PAT rotation is the single most time-sensitive item; address before any release.
- With 80-day gap now closed, the next run should be in 7 days (2026-08-02). If other routines have run by then, check their flags for backlog-worthy items.
- The P0 section now shows "No active P0 items" — verify this remains accurate; if any new blocking issues surface, add them immediately.
