---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-06-22
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 minutes
- **Files Read:** 5 — `station/agent/Routines/backlog-hygiene.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write, Grep
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section and checked each item against `Status.md` In Progress and Pending.
- **Result:** Found 2 P0 items that are actually RESOLVED and 1 that is already correctly in Status.md Pending:
  - `[bug] Sensor hook commands use $PWD-walk-up` — RESOLVED by v0.4.3 hotfix (PRs #105/#106, 2026-05-13). Was still listed as an active P0.
  - `[feature] bonsai init / bonsai add need non-interactive flags` — RESOLVED by v0.4.2 (PR #102, 2026-05-13). Was still listed as an active P0.
  - `[research] Trial sentrux` — already commented out with note "promoted to Status.md Pending 2026-05-07". Correct.
- **Action Taken:** Commented out both resolved P0 items with resolution notes in Backlog.md. P0 section is now empty of active items.
- **Issues:** None — the P0 section is now clean.

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md` and compared In Progress, Pending, and Recently Done entries against all Backlog items.
- **Result:** No Backlog items match anything currently In Progress (table is empty). The sentrux item is correctly in Pending. No "Blocked By" items in Pending could be unblocked by resolving a current Backlog item (sentrux is blocked on Rust toolchain install — a system dependency, not a Bonsai Backlog item). Recently Done includes Plan 41 (headless CLI), Plan 40 phases 1–3, v0.4.3 hotfix, v0.4.2 — all of which had corresponding Backlog items that were already removed or are now being commented out.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md` and cross-referenced P2/P3 items against current phase milestones.
- **Result:**
  - Phase 1 is fully complete (all items checked).
  - Phase 2 (Extensibility) is current. Two P3 items align with Phase 2 milestones:
    - `[improvement] Self-update mechanism` (P3 Big Bets) aligns with Phase 2 "Self-update mechanism" — candidate for P2 promotion when Phase 2 work begins.
    - `[improvement] Micro-task fast path` (P3) aligns with Phase 2 "Micro-task fast path" — same.
  - The P1 item `[feature] Full agent-drivable (non-interactive) CLI parity` is the primary next workstream and aligns with both Phase 2 (Extensibility) and Phase 3 (Cloud & Orchestration) goals.
  - No items reference deprecated approaches or completed phases that are still active entries.
- **Issues:** None requiring immediate action. Flagging the P3→P2 promotion candidates for user review.

### Step 4: Flag stale items
- **Action:** Audited all items for 30+ day staleness without progress, unclear rationale, and near-duplicates.
- **Result:** Several stale items identified:
  - `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` (P1, added 2026-04-22) — PAT was rotated 2026-04-22, expiry ~2026-07-15. **TODAY IS 2026-06-22 — 23 days until expiry.** This requires user action soon.
  - `[ops] Routine bot PR pile-up` (P1, added 2026-05-07) — ~46 days old with no architectural fix applied. Still relevant.
  - `[debt] Stale agent worktrees + branches` (P1, added 2026-04-20, updated 2026-04-21) — ~62 days old. Housekeeping debt. No progress.
  - `[debt] Testing infrastructure for triggers and sensors` (P1, Group B, added 2026-04-16) — ~67 days old. No progress, still relevant.
  - Group B items generally (added 2026-04-16): `generate.go` split, catalog test coverage, cmd test coverage, PTY smoke tests — all ~67 days old at P1 with no movement.
  - `[bookkeeping] Retroactively trim Backlog entries to NoteStandards` (Group A, added 2026-04-25) — ~58 days stale. The very backlog entries it targets have grown longer since.
  - `[improvement] Plans Index file` (Group E, added 2026-04-21) — ~62 days stale.
  - `[improvement] Consolidate FieldNotes usage` (Group E, added 2026-04-15) — ~68 days old.
  - Near-duplicates: `[feature] Changelog generation skill` (Group D) and the related CHANGELOG work in Group C mention overlapping scope — both are P2 and have been there since April.
- **Issues:** HOMEBREW_TAP_TOKEN expiry is time-sensitive and flagged for user attention.

### Step 5: Check for routine-generated items since last backlog-hygiene (2026-05-07)
- **Action:** Read recent RoutineLog.md entries since 2026-05-07. No routine entries appear between 2026-05-07 and 2026-06-22 (log goes from 2026-05-07 directly to 2026-06-13 Plan 40 dispatch, which is a plan execution, not a routine).
- **Result:** No routine-flagged issues between 2026-05-07 and today that are uncaptured. The Plan 40 and Plan 41 dispatch sessions generated several Backlog entries (2026-06-13 and 2026-06-16 additions) which are all present in the backlog. The website npm vuln finding (`[security] Website npm vuln tree`) was captured. No uncaptured findings.
- **Issues:** None.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any items are approved for immediate implementation.
- **Result:** No items are pre-approved. The `[feature] Full agent-drivable (non-interactive) CLI parity` (P1, added 2026-06-13) was noted by user as "main thing" and is plan-ready (`/plan` noted in the item itself) but not yet formally approved for dispatch. Not routing through issue-to-implementation without user confirmation.
- **Issues:** None — presenting for user review.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` — Backlog Hygiene row: Last Ran → 2026-06-22, Next Due → 2026-06-29, Status → done.
- **Result:** Done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | P0 `[bug] Sensor hook $PWD-walk-up` was still listed as active — resolved by v0.4.3 (2026-05-13) | Backlog.md P0 | Commented out with resolution note |
| 2 | High | P0 `[feature] non-interactive flags` was still listed as active — resolved by v0.4.2 (2026-05-13) | Backlog.md P0 | Commented out with resolution note |
| 3 | High | HOMEBREW_TAP_TOKEN PAT expires ~2026-07-15 (23 days from now) — needs rotation before next release | Backlog.md P1 | Flagged for user action |
| 4 | Medium | 6 Group B P1 debt items 60–67 days old with no movement | Backlog.md P1 Group B | Flagged for re-prioritization |
| 5 | Medium | `[debt] Stale agent worktrees + branches` — 62 days old, recurring housekeeping | Backlog.md P1 | Flagged for re-prioritization |
| 6 | Low | `[ops] Routine bot PR pile-up` — 46 days old, no fix applied | Backlog.md P1 | Flagged for re-prioritization |
| 7 | Low | 2 P3 items align with active Phase 2 milestones (Self-update mechanism, Micro-task fast path) — candidates for P2 promotion | Backlog.md P3 | Flagged for user review |
| 8 | Low | Group A NoteStandards sweep (~58 days old), Group E Plans Index + FieldNotes (~60–68 days old) — stale without progress | Backlog.md | Flagged for re-prioritization |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **HOMEBREW_TAP_TOKEN expiry — action required within 23 days.** PAT was rotated 2026-04-22 with 90-day default expiry → expires ~2026-07-15. Rotate before next release or GoReleaser brew step will silently fail. The Backlog P1 item documents the rotation procedure.

2. **P0 section is now empty.** Both remaining P0 items were confirmed resolved and commented out. If the user disagrees with either resolution, the entries can be restored.

3. **`[feature] Full agent-drivable CLI parity`** (P1, user-noted "main thing") — ready to promote to a plan via `/planning` when the user wants to start. No action taken without confirmation.

4. **Group B P1 debt (testing infrastructure + generate.go split + catalog/cmd coverage)** — 60–67 days old with no movement. Consider either scheduling into a plan or demoting to P2 if the roadmap priority has shifted.

5. **P3 → P2 promotion candidates:** `Self-update mechanism` and `Micro-task fast path` both appear in the active Phase 2 Roadmap milestones. Consider promoting from P3 to P2 to keep Backlog aligned with Roadmap.

## Notes for Next Run

- P0 section is clean after this run. If a new P0 arises it should go to Status.md Pending immediately per the priority guide.
- HOMEBREW_TAP_TOKEN should be rotated before the next run (due 2026-06-29) — check that it was handled.
- Group B P1 staleness will compound further if unaddressed. Consider a dedicated testing-infrastructure session.
- No routine-generated items were missed this cycle. Routine log gap (2026-05-07 → 2026-06-13) reflects a quiet period with no routine runs; confirm this was intentional (loop dormant) vs. missed runs.
