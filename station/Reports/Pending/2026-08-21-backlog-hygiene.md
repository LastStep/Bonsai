---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-21
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
- **Duration:** ~8 minutes
- **Files Read:** 5 — `station/agent/Routines/backlog-hygiene.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; cross-referenced each P0 item against Status.md (In Progress + Pending).
- **Result:** Found 2 P0 items that are RESOLVED, not misplaced — neither needs Status.md promotion, both need removal:
  - `[bug] Sensor hook commands use $PWD-walk-up` → fixed by v0.4.3 (PRs #105/#106, 2026-05-13). Status.md confirms: "sensor hook commands now bake install-time absolute paths."
  - `[feature] bonsai init/add need non-interactive flags` → fixed by v0.4.2 (PR #102, 2026-05-13). Status.md confirms `--non-interactive --from-config` shipped.
  - Both removed from Backlog with HTML comments preserving the audit trail.
- **Issues:** None — P0 section is now empty (only HTML comments remain), which is the correct state.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md; compared In Progress + Recently Done against Backlog items.
- **Result:**
  - Identified one additional resolved Backlog item: P1 `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` (added 2026-06-13). Plan 41 shipped all 5 phases (PRs #120/#122/#123/#121/#125, 2026-06-16), delivering `*Result` headless cores + JSONL/exit contract for all four mutating commands plus `list --json` and `docs/agent-interface.md`. Item removed from P1 with HTML comment.
  - Status.md Pending has one item: `[research] Trial sentrux` — already removed from Backlog P0 with an HTML comment (2026-05-07 routine-digest). Consistent.
  - No Status.md Pending "Blocked By" items could be unblocked by a current Backlog item.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; checked P2/P3 Backlog items against current phase milestones.
- **Result:**
  - Roadmap Phase 1 is fully checked [x] — no Backlog items reference incomplete Phase 1 work.
  - Phase 2 (Extensibility) aligns with two P3 items: `[improvement] Self-update mechanism` and `[improvement] Micro-task fast path`. Both are appropriately filed at P3 — Phase 2 is not yet active, no promotion warranted.
  - Phase 3 (Cloud & Orchestration) aligns with P3 `[feature] Managed Agents integration` and `[feature] Greenhouse companion app` — correctly at P3.
  - No Backlog items reference deprecated approaches or completed phases.
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Scanned all Backlog items for age (30+ days at same priority) and clarity issues.
- **Result:** Several items are stale at 100+ days with no progress, flagged below. Key findings:
  - **CRITICAL — HOMEBREW_TAP_TOKEN PAT likely expired:** P1 item `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` (added 2026-04-22) documents that the PAT was rotated 2026-04-22 with a 90-day expiry, and sets a reminder to rotate by ~2026-07-15. Today is 2026-08-21 — **37 days past the rotation deadline**. An expired PAT would cause GoReleaser to fail the Homebrew formula update on the next release (binaries publish but formula is missed). This needs immediate user attention before the next release.
  - P1 `[debt] Testing infrastructure for triggers and sensors` — added 2026-04-16, 127 days stale.
  - P1 `[debt] Stale agent worktrees + branches accumulating` — added 2026-04-20, 123 days stale.
  - P1 `[ops] Routine bot PR pile-up` — added 2026-05-07, 106 days stale.
  - Several Group B/C/D/E P2 items are 120+ days old with no progress — appropriate for P2 holding pattern but noted.
- **Issues:** PAT expiry is an urgent flag requiring user action.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since 2026-05-07 (last backlog-hygiene run).
- **Result:** Only one relevant entry post-2026-05-07: the 2026-06-13 Plan 40 dispatch note (not a routine, but it filed several Backlog items). All items it filed are captured:
  - `[security] Harden scaffolding writes against symlink substitution` → Backlog P2 ✓
  - `[improvement] bonsai validate warn on .bonsai/project.yaml drift` → Backlog P2 ✓
  - `[improvement] Plan 40 review nits` → Backlog P2 ✓
  - `[bug] bonsai validate can't pass on Bonsai repo itself` → Backlog P2 ✓
  - `[debt] Unify remove business logic` → Backlog P2 ✓
  - `[security] Website npm vuln tree — astro upgrade breaks build` → Backlog P2 ✓
  - No uncaptured findings.
- **Issues:** None.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed whether any item warrants immediate promotion with user confirmation.
- **Result:** No items are explicitly approved for promotion by the user. The PAT expiry is urgent but is an ops action (rotate the PAT), not an implementation workflow. Flagged in Step 4 for user review.
- **Issues:** None — no autonomous promotion performed.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Entry appended.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` Backlog Hygiene row.
- **Result:** `Last Ran` → 2026-08-21, `Next Due` → 2026-08-28, `Status` → `done`.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Critical | HOMEBREW_TAP_TOKEN PAT likely expired (37 days past 2026-07-15 rotation deadline) | Backlog P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry` | Flagged for user — no code action possible |
| 2 | Info | P0 `[bug] Sensor hook $PWD-walk-up` was resolved by v0.4.3 but still in Backlog | Backlog P0 | Removed with HTML audit comment |
| 3 | Info | P0 `[feature] bonsai init/add non-interactive flags` resolved by v0.4.2 but still in Backlog | Backlog P0 | Removed with HTML audit comment |
| 4 | Info | P1 `[feature] Full agent-drivable CLI parity` resolved by Plan 41 but still in Backlog | Backlog P1 | Removed with HTML audit comment |
| 5 | Low | P1 `[debt] Testing infrastructure for triggers and sensors` stale 127 days | Backlog P1 | Flagged for re-prioritization |
| 6 | Low | P1 `[debt] Stale agent worktrees + branches` stale 123 days | Backlog P1 | Flagged for re-prioritization |
| 7 | Low | P1 `[ops] Routine bot PR pile-up` stale 106 days | Backlog P1 | Flagged for re-prioritization |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **URGENT — Rotate HOMEBREW_TAP_TOKEN PAT immediately.** The fine-grained PAT rotated 2026-04-22 (90-day expiry) was due for rotation by ~2026-07-15. Today is 2026-08-21. Symptoms if already expired: next `gh release create` / GoReleaser run will fail the Homebrew formula update step with `401 Bad credentials` — binaries still publish, but the tap formula goes stale. Action: rotate the PAT in GitHub Settings, update the `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai`, and update the Backlog item with the new rotation deadline (~2026-11-19 for another 90 days).

2. **Three stale P1 items (127/123/106 days)** — Consider demoting to P2 or setting a concrete next-action if still relevant:
   - `[debt] Testing infrastructure for triggers and sensors`
   - `[debt] Stale agent worktrees + branches accumulating`
   - `[ops] Routine bot PR pile-up — eliminate parallel-track artifacts`

## Notes for Next Run

- P0 section is now empty of live items (only HTML comments remain) — healthy state.
- With Plan 41 shipped (full headless CLI parity), the backlog is in a cleaner state than the last hygiene run.
- The HOMEBREW_TAP_TOKEN flag is the most actionable item from this cycle — confirm rotation before next release.
- Next run due 2026-08-28.
