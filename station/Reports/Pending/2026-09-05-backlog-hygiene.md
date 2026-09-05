---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-05
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
- **Files Modified:** 2 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog P0 section and compared each item against Status.md.
- **Result:** Both P0 items found in Backlog are already resolved — not missing from Status.md, but already in "Recently Done." No unresolved P0s remain.
- **Issues:** none

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress, Pending, and Recently Done tables. Cross-checked all Backlog items.
- **Result:**
  - P0 "[bug] Sensor hook commands use `$PWD`-walk-up" — RESOLVED via v0.4.3 (PR #105/#106, Status.md Recently Done 2026-05-13). **REMOVED from Backlog** with HTML comment.
  - P0 "[feature] `bonsai init`/`bonsai add` need non-interactive flags" — RESOLVED via v0.4.2 (PR #102, Status.md Recently Done 2026-05-13). **REMOVED from Backlog** with HTML comment.
  - P1 "[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove" — Plan 41 shipped `*Result` headless cores for all four commands + `list --json` + `docs/agent-interface.md` contract (Status.md Recently Done 2026-06-16). This item may be substantially resolved. **Flagged for user review.**
  - Status.md Pending: "[research] Trial sentrux" (blocked on Rust toolchain) — already commented out in Backlog P0 as promoted. Correct state.
- **Issues:** none

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md. Compared P2/P3 Backlog items against phase milestones.
- **Result:**
  - Phase 1 is fully complete (all checkboxes checked, per the last roadmap-accuracy routine in May 2026).
  - Phase 2 milestones: "Self-update mechanism" and "Micro-task fast path" are in Backlog P3 (Future Platform). Appropriate — Phase 2 is not the current active workstream.
  - Phase 3/4 milestones (Managed Agents, Greenhouse app, Catalog marketplace) all have matching P3 Backlog entries. No promotion warranted.
  - No Backlog items reference deprecated Phase 1 approaches or completed milestones in an incompatible way.
- **Issues:** none

### Step 4: Flag stale items
- **Action:** Scanned all Backlog items for 30+ day staleness and items with no clear context.
- **Result:** Multiple items are stale (30–142 days without progress). Key findings:
  1. **P1 [ops] HOMEBREW_TAP_TOKEN PAT expiry** (added 2026-04-22) — reminder was for ~2026-07-15; that date has passed. PAT has almost certainly expired again. Needs immediate user action before next release.
  2. **P1 [debt] Testing infrastructure for triggers and sensors** (added 2026-04-16) — 142 days stale, no visible progress. Candidate for re-prioritization or demotion to P2.
  3. **P1 [debt] Stale agent worktrees + branches accumulating** (added 2026-04-20) — 138 days; pattern is recurrent but no structural fix shipped.
  4. **P2 [security] Website npm vuln tree — astro upgrade breaks build** (added 2026-06-16) — 81 days, active Dependabot alerts, astro 6.1.7→6.3.2 bump still failing. Security-relevant, should not age further.
  5. **P2 [bug] `bonsai validate` can't pass on the Bonsai repo itself** (added 2026-06-13) — 84 days; blocks dogfood verification.
  6. **Group A [bookkeeping] Retroactively trim Backlog entries** (added 2026-04-25) — 133 days; cosmetic but keeps Backlog bloated.
- **Issues:** none blocking — all flagged for user review

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07).
- **Result:** No routine log entries exist between 2026-05-07 and 2026-09-05 (RoutineLog is empty for that span). Plans 40 and 41 produced review findings that were captured in Backlog (P2 security items added 2026-06-13 and 2026-06-16) — confirmed captured.
- **Issues:** none; no uncaptured routine findings to flag

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed all items for promotion readiness.
- **Result:** No items confirmed approved by user for immediate implementation. P1 "[feature] Full agent-drivable (non-interactive) CLI parity" is the primary candidate for closure review (Plan 41 may have resolved it). No workflow dispatch initiated — user confirmation required per procedure.
- **Issues:** none

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.

### Step 8: Update dashboard
- **Action:** Updated Backlog Hygiene row in `agent/Core/routines.md` — Last Ran → 2026-09-05, Next Due → 2026-09-12, Status → done.
- **Result:** Done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | P0 "[bug] Sensor hook commands `$PWD`-walk-up" was resolved in v0.4.3 (2026-05-13) but remained in Backlog P0 | Backlog.md P0 | Removed — added HTML comment with resolution note |
| 2 | HIGH | P0 "[feature] non-interactive flags for init/add" was resolved in v0.4.2 (2026-05-13) but remained in Backlog P0 | Backlog.md P0 | Removed — added HTML comment with resolution note |
| 3 | MEDIUM | P1 "[feature] Full agent-drivable CLI parity" may be substantially resolved by Plan 41 (headless `*Result` cores for all four commands) | Backlog.md P1 | Flagged for user review — no autonomous action |
| 4 | MEDIUM | P1 "[ops] HOMEBREW_TAP_TOKEN PAT expiry" — reminder date ~2026-07-15 has passed; PAT likely expired before next release | Backlog.md P1 | Flagged for user — needs PAT rotation and calendar update |
| 5 | MEDIUM | P2 "[security] Website npm vuln tree" — 81 days stale, active security alerts, astro upgrade still failing build | Backlog.md P2 | Flagged for user — security item should not age further |
| 6 | LOW | P2 "[bug] `bonsai validate` can't pass on the Bonsai repo itself" — 84 days stale, blocks dogfood | Backlog.md P2 | Flagged for user |
| 7 | LOW | P1 "[debt] Testing infrastructure for triggers and sensors" — 142 days stale | Backlog.md P1 | Flagged for user — candidate for demotion to P2 |
| 8 | LOW | Multiple P2/P3 Group items (A, B, C, D, E, F) from April 2026 are 120–142 days stale with no visible progress | Backlog.md | Flagged generally — no autonomous action |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **P1 [feature] Full agent-drivable CLI parity** — Plan 41 shipped headless cores for all four commands. Is this item now closed? If so, remove from Backlog.

2. **P1 [ops] HOMEBREW_TAP_TOKEN PAT** — reminder date ~2026-07-15 has passed. Rotate PAT and update the secret on `LastStep/Bonsai` before next release. Update the calendar reminder.

3. **P2 [security] Website npm vuln tree** — 81 days without resolution. The astro 6.1.7→6.3.2 bump (PR #108) fails the build after rebase. Should be prioritized soon — Dependabot alerts accumulate.

4. **P2 [bug] `bonsai validate` can't pass on the Bonsai repo** — 84 days. Gitignored `.bonsai-lock.yaml` blocks the dogfood gate. Decide lock-file policy.

5. **P1 [debt] Testing infrastructure** — 142 days stale. Consider demoting to P2 or setting a concrete milestone.

## Notes for Next Run

- P0 section is now clear — no active P0 items. Future runs should be fast if the user processes flagged items above.
- Gap since last run was 121 days (2026-05-07 → 2026-09-05). All other routines are similarly overdue.
- Plan 41 shipped major functionality (headless CLI contract) that may resolve multiple Backlog items — the user should do a brief Backlog sweep after reviewing this report.
- HOMEBREW_TAP_TOKEN expiry is time-sensitive before any release attempt.
