---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-06-27
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
- **Duration:** ~8 min
- **Files Read:** 4 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Tools Used:** Read, Edit, Write, Bash
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Scanned P0 section of Backlog.md; checked each item against Status.md (In Progress + Pending).
- **Result:** Found 2 P0 items that are actually resolved — both removed (commented out with audit trail):
  1. `[bug] Sensor hook commands use $PWD-walk-up` — shipped in v0.4.3 (PR #105/#106, 2026-05-13). Status.md confirms "v0.4.3 hotfix shipped."
  2. `[feature] bonsai init / bonsai add need non-interactive flags` — shipped in v0.4.2 (PR #102, 2026-05-13). Superseded by the P1 "Full agent-drivable CLI parity" item.
- **Issues:** P0 section is now empty of live items (only HTML comments). This is correct — P0 items should be in Status.md, and both resolved items were already completed before this run.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md; compared In Progress, Pending, and Recently Done against all Backlog items.
- **Result:**
  - No Backlog items match anything currently In Progress (table is empty).
  - Pending item `[research] Trial sentrux` is correctly shown as a comment in the P0 section (promoted 2026-05-07) — no stale Backlog entry persists.
  - Recently Done includes Plan 41 (Headless CLI Contract, all 5 phases merged) — confirmed this supersedes/resolves the P0 non-interactive flags item.
  - No Status.md Pending items with "Blocked By" are unblockable by a current Backlog item (sentrux is blocked on Rust toolchain install — not a Backlog dependency).
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; scanned Phase 2+ milestones against P2/P3 Backlog items.
- **Result:**
  - Phase 1 is fully complete (all checkboxes checked). No stale Phase 1 tracking needed.
  - Phase 2 milestones: "Self-update mechanism" and "Micro-task fast path" both exist in P3 Backlog — appropriate priority given Phase 3/4 is far off.
  - "Template variables expansion" (Phase 2) has no Backlog item — this is a gap, but not a blocking issue. Flagging for user awareness.
  - "Custom item detection" (Phase 2) is `[x]` and no stale Backlog item exists — clean.
  - No items reference deprecated approaches or completed phases.
- **Issues:** Minor — Phase 2 "Template variables expansion" milestone lacks a Backlog tracking entry.

### Step 4: Flag stale items
- **Action:** Scanned all Backlog items for age (last backlog-hygiene: 2026-05-07; today: 2026-06-27 = 51 days elapsed). Checked items present since prior runs.
- **Result:**
  - Most Group B, C, D, E items have been at the same priority since 2026-04-16 (72+ days) with no progress — these are known long-term items awaiting capacity.
  - **CRITICAL FINDING:** P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — PAT rotated 2026-04-22 with ~90-day expiry. As of 2026-06-27 this is approximately 24 days from expiry (~2026-07-21). Updated the item with `[URGENT — expires ~2026-07-21]` label and current-date urgency note. **This requires immediate user action.**
  - P1 `[ops] Routine bot PR pile-up` — added 2026-05-07, still pending. Stale at 51 days.
  - P1 `[debt] Stale agent worktrees + branches accumulating` — added 2026-04-20 (~68 days). Stale, no progress since last run.
  - Near-duplicate confirmed persists: Group C "CHANGELOG + richer release notes" (line in Group B notes) vs Group D "Changelog generation skill" — known from prior runs, flagged again.
- **Issues:** PAT expiry is urgent. Multiple P1 debt items stale 50+ days without progress.

### Step 5: Check for routine-generated items needing Backlog capture
- **Action:** Read RoutineLog.md entries since 2026-05-07 (last backlog-hygiene run).
- **Result:** No routine entries exist in RoutineLog between 2026-05-07 and 2026-06-27. The only entries after 2026-05-07 is the `2026-06-13 — Plan 40 dispatch` session note (not a routine). No uncaptured routine findings.
- **Issues:** None — a 51-day gap in routine execution. This is the longest gap observed. Routine cadence may have lapsed due to the extended Plan 40/41 implementation cycle.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Evaluated whether any items are approved for immediate promotion.
- **Result:** No items have explicit user approval for immediate implementation. The P1 "Full agent-drivable CLI parity" item was described as "Promote to a plan + grill next session" — this is the natural next action for the user but requires user decision. Not auto-promoting.
- **Issues:** None — deferred to user.

### Step 7: Log results
- **Action:** Appended entry to RoutineLog.md.
- **Result:** Done.

### Step 8: Update dashboard
- **Action:** Updated routines.md dashboard row for Backlog Hygiene.
- **Result:** Done — Last Ran → 2026-06-27, Next Due → 2026-07-04, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | P0 bug item `$PWD-walk-up` was already resolved in v0.4.3 (2026-05-13) — stale in Backlog | Backlog P0 | Commented out with resolution note |
| 2 | high | P0 feature item `--non-interactive flags` was already resolved in v0.4.2 — stale in Backlog | Backlog P0 | Commented out with resolution note |
| 3 | critical | `HOMEBREW_TAP_TOKEN` PAT expires ~2026-07-21 (~24 days) — imminent Homebrew publish failure on next release | Backlog P1 | Added URGENT label + expiry date to item; flagging for user |
| 4 | low | Phase 2 Roadmap milestone "Template variables expansion" has no Backlog tracking entry | Roadmap Phase 2 / Backlog | Flagged for user — no action (gap, not urgent) |
| 5 | low | Plans 40 + 41 still in Plans/Active/ despite being shipped | Plans/Active/ | Flagged for user — plan archiving is a known backlog item (Group E) |
| 6 | info | 51-day gap since last routine execution — longest observed lapse | RoutineLog | Noted; no action required |
| 7 | low | P1 debt items stale 50-70+ days without progress (worktrees/branches, bot PR pile-up) | Backlog P1 | Flagged for user review |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **URGENT — Rotate HOMEBREW_TAP_TOKEN PAT now.** PAT rotated 2026-04-22, 90-day default = expires ~2026-07-21 (~24 days from today). Rotate at `https://github.com/settings/personal-access-tokens` and update secret at `https://github.com/LastStep/Bonsai/settings/secrets/actions`. Failure symptom: GoReleaser brew step 401 on next release — binaries publish but Homebrew formula misses the bump.

2. **Phase 2 "Template variables expansion" lacks Backlog tracking.** Decide whether to add a Backlog item or defer indefinitely.

3. **Plans 40 + 41 in Plans/Active/ despite being shipped.** Plan 41 shipped all 5 phases. Plan 40 Phases 1-3 merged but Phase 4 held. Archive Plan 41 now; archive Plan 40 or mark status appropriately.

4. **P1 "Full agent-drivable CLI parity"** — flagged for promotion via `/planning` next session. This was the user's stated "main thing" (2026-06-13).

## Notes for Next Run

- P0 section is now clean (all live items resolved or promoted). If new P0s appear, they should be promoted to Status.md immediately.
- PAT rotation should be confirmed resolved before next run (2026-07-04).
- If Plans 40/41 are not archived by next run, flag again.
- The 51-day gap in routine execution suggests the cloud routine dispatch may have lapsed — verify loop.md dispatch health.
- Near-duplicate: changelog items in Group C vs Group D (known from multiple prior runs) — consider consolidating or explicitly deciding to keep both.
