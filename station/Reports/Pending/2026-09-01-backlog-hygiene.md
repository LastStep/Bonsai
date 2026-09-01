---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-01
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
- **Duration:** ~5 min
- **Files Read:** 5 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section; cross-referenced each item against `Status.md`.
- **Result:** Both P0 items were already RESOLVED in the codebase but had not been cleaned from the backlog:
  - **[bug] Sensor hook commands use `$PWD`-walk-up** — Fixed in v0.4.3 (PRs #105/#106, 2026-05-13). Status.md confirms the hotfix shipped. Commented out.
  - **[feature] `bonsai init`/`add` non-interactive flags** — Fixed in v0.4.2 (`--non-interactive` + `--from-config` shipped, 2026-05-13). Status.md confirms. Commented out.
- **Issues:** Both P0 items were stale resolved work. Removed from active P0 list; HTML comments preserve the audit trail.

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md` In Progress, Pending, and Recently Done tables; matched against all Backlog entries.
- **Result:**
  - **P1 "Full agent-drivable CLI parity (init/update/add/remove)"** — Plan 41 (PRs #120/#122/#123/#121/#125, 2026-06-16) shipped all five phases: headless cores for all four commands, JSONL/exit-code contract (`ExitConflict=5`), `list --json`, and `docs/agent-interface.md`. This P1 is fully resolved. Commented out.
  - **P1 "[research] Trial sentrux"** — Still in Status.md Pending, blocked on Rust toolchain. No change.
  - No other Backlog items matched In Progress or Recently Done rows that would indicate additional resolved work.
- **Issues:** None after cleanup.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md`; checked P2/P3 Backlog items against current and future phase milestones.
- **Result:**
  - Phase 1 is complete (all items checked). No stale Backlog references to Phase 1 uncompleted items.
  - Phase 2 milestones ("Self-update mechanism", "Micro-task fast path") map to existing P3 Backlog items — appropriate priority, no promotion needed at this time.
  - No Backlog items reference deprecated approaches or completed phases.
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Reviewed items for 30+ day staleness without progress, unclear rationale, and near-duplicates.
- **Result:**
  - **HOMEBREW_TAP_TOKEN PAT expiry (P1)** — Added 2026-04-22, rotation due ~2026-07-15. Today is 2026-09-01. The PAT is 47+ days past its expected expiry. Entry updated with urgency: "OVERDUE, rotate immediately." Flagged for user action.
  - Most P3 items (Big Bets, Research, Future Platform) are intentionally long-horizon — staleness expected and acceptable for this tier.
  - P2 items added 2026-06-13 (Plan 40 review nits, validate dog-food blocker, symlink hardening, etc.) are 79 days old but are non-blocking tech debt — no action warranted without user direction.
  - **No near-duplicates found** — the CHANGELOG generation vs Group D changelog item was already noted as Group D; no new conflicts found.
- **Issues:** PAT expiry escalated.

### Step 5: Check for routine-generated items
- **Action:** Read `RoutineLog.md` entries since last backlog-hygiene (2026-05-07): found 2026-05-07 routine batch entries and 2026-06-13 Plan 40 dispatch entry. Verified all flagged items were captured in Backlog.
- **Result:** All Plan 40/41 dispatch-generated backlog items are accounted for in Backlog.md (P2 symlink hardening, validate drift warning, review nits, validate dog-food blocker, website npm vuln tree, remove business logic unification). No uncaptured findings.
- **Issues:** None — routine coverage is complete.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed items that could be implementation-ready; no user confirmation is available in this subagent context.
- **Result:** The HOMEBREW_TAP_TOKEN PAT expiry is the most urgent actionable item — flagged for user review and immediate action. All other items require user prioritization before workflow routing.
- **Issues:** Cannot initiate issue-to-implementation without user approval.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Entry written.

### Step 8: Update dashboard
- **Action:** Updated `Backlog Hygiene` row in `station/agent/Core/routines.md` dashboard.
- **Result:** `Last Ran` → 2026-09-01, `Next Due` → 2026-09-08, `Status` → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | P0 "[bug] Sensor hook $PWD-walk-up" is RESOLVED (v0.4.3, 2026-05-13) but still listed as active P0 | `Backlog.md` P0 | Commented out with resolution note |
| 2 | high | P0 "[feature] non-interactive flags" is RESOLVED (v0.4.2, 2026-05-13) but still listed as active P0 | `Backlog.md` P0 | Commented out with resolution note |
| 3 | high | P1 "Full agent-drivable CLI parity" is RESOLVED (Plan 41, 2026-06-16) but still listed as active P1 | `Backlog.md` P1 | Commented out with resolution note |
| 4 | high | HOMEBREW_TAP_TOKEN PAT expired ~2026-07-15 — now 47+ days overdue, next release will fail at brew step | `Backlog.md` P1 | Entry updated with urgency; flagged for user |
| 5 | info | No uncaptured routine-generated findings from 2026-05-07 → 2026-09-01 period | `RoutineLog.md` scan | No action needed |
| 6 | info | Roadmap alignment healthy — P3 items for Phase 2 milestones at correct priority | `Roadmap.md` cross-ref | No change |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

- **[URGENT] HOMEBREW_TAP_TOKEN PAT has expired.** Was due for rotation ~2026-07-15; today is 2026-09-01. Rotate the fine-grained PAT in GitHub `LastStep/Bonsai` repo secrets immediately, then document next rotation due date (~2026-12-01) in Backlog. Next release will fail at the Homebrew formula update step with `401 Bad credentials` until this is rotated.

- **[ADVISORY] 117-day gap since last backlog-hygiene.** The routine was last run 2026-05-07 and has not run since. Several resolved items accumulated stale P0/P1 entries. Recommend running this routine at its nominal 7-day cadence going forward.

## Notes for Next Run

- Three items cleaned from P0/P1 this run. The P0 section is now empty of active items (all entries are HTML comments). If a new P0 surfaces, it will stand out clearly.
- HOMEBREW_TAP_TOKEN rotation should be verified by next run — confirm the P1 entry has been addressed or updated with new expiry date.
- P2 items added 2026-06-13 (Plan 40/41 review debt) are now ~80+ days old. If the user does not prioritize them by the next run, consider re-evaluating whether they should be P3 or removed.
- `bonsai validate` cannot pass on the Bonsai repo itself (lock file gitignored — P2 backlog item) — worth raising to user as a blocking hygiene issue.
