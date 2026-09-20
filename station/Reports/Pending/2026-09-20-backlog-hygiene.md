---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-20
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
- **Files Read:** 7 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`, `station/agent/Core/identity.md`, `station/agent/Core/memory.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Playbook/Backlog.md` P0 section; cross-referenced each P0 against `Status.md` In Progress and Pending.
- **Result:** Found **2 P0 items that are RESOLVED** and should not remain as P0 blockers:
  1. `[bug] Sensor hook commands use $PWD-walk-up` — fixed in v0.4.3 (2026-05-13, PRs #105/#106, commit 584b82b).
  2. `[feature] bonsai init/add need non-interactive flags` — fixed in v0.4.2 (2026-05-13, PR #102, commit 410a5f1).
  The sentrux research item was already commented out (promoted to Status.md Pending on 2026-05-07).
- **Issues:** Both P0 items were resolved after the last backlog-hygiene run (2026-05-07) but never cleaned up. Removed both from P0 section; replaced with HTML audit-trail comments noting resolution date and artifact links.

### Step 2: Cross-reference with Status.md
- **Action:** Read `Playbook/Status.md`; matched Backlog items against In Progress, Pending, and Recently Done.
- **Result:**
  - **2 P0 items removed** (see Step 1 above — both appear in Recently Done via v0.4.2/v0.4.3 ships).
  - **P1 "Full agent-drivable CLI parity"** remains valid — v0.4.2 shipped init/add only; update/remove headless coverage is still open (the P1 explicitly scopes this).
  - **Pending:** sentrux trial is correctly placed in Status.md Pending (blocked on Rust toolchain). No action needed.
  - **No additional promotions or removals** triggered by Status cross-reference.
- **Issues:** None beyond the two resolved P0s already handled.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Playbook/Roadmap.md`; compared P2/P3 items against current and future phase milestones.
- **Result:**
  - Phase 1 is **fully complete** (all items checked). No Backlog items reference Phase 1 gaps.
  - Phase 2 (Extensibility) milestones: "Self-update mechanism" and "Micro-task fast path" both have P3 Backlog entries — correctly placed at P3 (not urgent, Phase 2 work not yet started).
  - Phase 2 "Template variables expansion" has **no Backlog entry** — omission noted but this is an intentional Phase 2 roadmap item, not requiring a Backlog capture unless scoped.
  - No items reference deprecated approaches or completed phases that don't exist.
  - `[feature] Full agent-drivable CLI parity` (P1) is the most Phase-2-aligned active item; already at P1 and being tracked — no promotion needed.
- **Issues:** None. Roadmap alignment is healthy.

### Step 4: Flag stale items
- **Action:** Reviewed all P1–P3 items for age and progress signal. Last backlog-hygiene was 2026-05-07 (4.5 months ago); any item unchanged since then qualifies as stale (>30 days at same priority).
- **Result:** All items are effectively stale by the 30-day threshold since no hygiene run has occurred. Specific high-priority stale findings:
  1. **URGENT — P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry`**: Calendar reminder was set for ~2026-07-15. Today is 2026-09-20 — **the PAT may already be expired** (2 months past the reminder date). If a release is attempted now, the Homebrew formula step will fail. See Findings table.
  2. **P1 `[debt] Stale agent worktrees + branches`** (added 2026-04-20): No progress signal in 5 months. Worth revisiting or accepting as ongoing hygiene.
  3. **Group B items** (Code Quality & Testing): All unchanged since April 2026. The codebase has grown (Plans 34–41 shipped) but these items remain unaddressed. Not escalated — still valid P1/P2, no new urgency.
  4. **`Plans/Active/41-headless-cli-contract.md`**: Memory.md explicitly notes "Plan 41 file still in Plans/Active/ — archive to Plans/Archive/ at next wrap-up." This is a bookkeeping item, not a Backlog item, but flagged here for the Tech Lead.
- **Issues:** No items removed for staleness (staleness alone is not grounds for removal — items remain valid). HOMEBREW_TAP_TOKEN is flagged as urgent for user attention.

### Step 5: Check for routine-generated items
- **Action:** Read `Logs/RoutineLog.md` entries since 2026-05-07 (last backlog-hygiene run).
- **Result:** No routines have been dispatched since 2026-05-07. The only post-date log entry is "2026-06-13 — Plan 40 dispatch" which is a session plan log, not a routine. Therefore **no uncaptured routine findings** exist.
- **Issues:** None. (Note: All 7 routines are significantly overdue — last ran 2026-05-04 to 2026-05-07, all well past their 5–14 day frequency windows given today's date of 2026-09-20. The routine-check sensor will surface this.)

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed P0 and P1 for items approved or urgent enough for immediate promotion.
- **Result:** No items are pre-approved for implementation. The most actionable item is P1 `[feature] Full agent-drivable CLI parity` (init/update/add/remove headless parity) — already flagged in memory.md as "main thing" and scoped as the next plan (`/plan` at next session). No workflow launch triggered (requires user confirmation per procedure).
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Entry written.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Backlog Hygiene row: Last Ran → 2026-09-20, Next Due → 2026-09-27, Status → done.
- **Result:** Updated.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | **High** | P0 `[bug] Sensor hook $PWD-walk-up` resolved in v0.4.3 but not removed from Backlog | `Backlog.md` P0 section | Removed; replaced with HTML audit-trail comment |
| 2 | **High** | P0 `[feature] bonsai init/add non-interactive flags` resolved in v0.4.2 but not removed | `Backlog.md` P0 section | Removed; replaced with HTML audit-trail comment |
| 3 | **High** | P1 `HOMEBREW_TAP_TOKEN PAT` calendar reminder was 2026-07-15 — **2 months overdue** | `Backlog.md` P1 | Flagged for immediate user review — no auto-fix |
| 4 | Low | `Plans/Active/41-headless-cli-contract.md` should be archived per `memory.md` note | `Plans/Active/` | Flagged for Tech Lead — out of routine scope |
| 5 | Info | All 7 routines significantly overdue (last ran 2026-05-04/07, ~4.5 months ago) | `agent/Core/routines.md` | Informational — routine-check sensor will surface this |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **URGENT — HOMEBREW_TAP_TOKEN PAT** (P1): The fine-grained PAT was rotated 2026-04-22 with a 90-day expiry; the reminder date of ~2026-07-15 is now 2 months past. **Action needed:** Rotate the PAT at `https://github.com/settings/tokens`, set `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai`, and update the Backlog item's reminder date. Symptom of expiry: GoReleaser Homebrew step fails with `401 Bad credentials` while binaries publish successfully.

2. **INFORMATIONAL — Plan 41 archive**: `Plans/Active/41-headless-cli-contract.md` should be moved to `Plans/Archive/` per the memory.md note. Can be done in the next wrap-up session.

## Notes for Next Run

- Both P0 items are cleared — the P0 section is now empty (sentrux is in Status.md Pending). Next run should verify sentrux is either progressed or deprioritized.
- All 7 routines are significantly overdue. A routine-digest session to process them all would clear substantial backlog/doc/dependency drift.
- The PAT rotation (finding #3) should be done before the next release attempt, regardless of when backlog-hygiene next runs.
- P1 "Full agent-drivable CLI parity" is the highest-value open item — plan at next session (`/planning`).
