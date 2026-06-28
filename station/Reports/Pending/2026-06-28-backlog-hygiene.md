---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-06-28
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07 (previous value from dashboard, before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 5 — `station/agent/Routines/backlog-hygiene.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write, Grep, Bash (ls)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; cross-referenced each item against Status.md In Progress and Pending tables.
- **Result:** Found 2 P0 items that are fully resolved and should be removed:
  - `[bug] Sensor hook commands use $PWD-walk-up` — SHIPPED in v0.4.3 (2026-05-13, PR #105/#106). Absolute install-time paths now baked into hook commands. Still present as an active P0 bullet despite the fix.
  - `[feature] bonsai init / bonsai add need non-interactive flags` — SHIPPED in v0.4.2 (2026-05-13, PR #102). `--non-interactive` + `--from-config` flags exist. Still present as an active P0 bullet.
  - `[research] Trial sentrux` — correctly handled: already promoted to Status.md Pending (HTML comment in Backlog confirmed).
  - **Action taken:** Converted both resolved P0 bullets to HTML audit-trail comments (standard pattern for resolved Backlog items).
- **Issues:** P0 section was left with resolved items for 46+ days after shipping. Cleaned up.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md. Checked In Progress, Pending, and Recently Done tables against Backlog items.
- **Result:**
  - In Progress: empty — no cross-ref needed.
  - Pending: only `Trial sentrux` (already a comment in Backlog — correct).
  - Recently Done: Plans 34–41 all shipped. The two resolved P0 items (v0.4.2 + v0.4.3) appear in Recently Done — this confirmed they were safe to remove from Backlog P0.
  - No Pending items with "Blocked By" that could be unblocked by Backlog items (sentrux is blocked by Rust toolchain, not a Backlog item).
- **Issues:** None beyond the resolved P0s already handled in Step 1.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md. Checked Phase 1 (current), Phase 2, Phase 3, Phase 4 against Backlog items.
- **Result:**
  - Phase 1: All milestones checked. Backlog has no P2/P3 items misaligned with Phase 1 (it's essentially complete).
  - Phase 2 milestones (`Self-update mechanism`, `Template variables expansion`, `Micro-task fast path`): All present in Backlog at P3 or as research items. `Template variables expansion` has no dedicated Backlog entry — but it's a Phase 2 milestone item with no urgency given Phase 1 is still the active phase.
  - Phase 3 milestones (`Managed Agents integration`, `Greenhouse companion app`): Both captured in Backlog P3 Big Bets.
  - No deprecated-approach references found in Backlog items checked.
  - No P2/P3 items require promotion based on current phase alignment (Phase 1 complete, Phase 2 not yet entered).
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Reviewed all Backlog items for age (30+ days at same priority without progress). Today is 2026-06-28.
- **Result:** Key findings:
  1. **`[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder`** (P1, added 2026-04-22, 67 days old) — **TIME-SENSITIVE**: PAT was rotated 2026-04-22 on a ~90-day cycle, meaning it expires ~2026-07-22. With only ~17 days until expiry, this needs immediate user attention before the next release. Added `[TIME-SENSITIVE: due ~2026-07-15]` tag and updated wording to surface urgency.
  2. **`[ops] Routine bot PR pile-up`** (P1, added 2026-05-07, 52 days old) — no progress noted. Still valid concern but non-blocking.
  3. **`[debt] Testing infrastructure for triggers and sensors`** (P1, added 2026-04-16, 73 days old) — long-standing; context still valid. No near-duplicate found.
  4. **`[debt] Stale agent worktrees + branches accumulating`** (P1, added 2026-04-20, 69 days old) — audit was based on a 2026-04-21 snapshot. Post-Plan-41, the worktree situation has likely evolved; the count and specific branches cited may be stale. Flagged for user re-check.
  5. Group B items (added 2026-04-16, 73 days): code quality debt items still valid but no change in priority warranted.
  6. Group C, D, E items similarly aged but all have clear rationale.
- **Issues:** PAT expiry is the highest-urgency stale finding.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07). Identified routine-generated findings from: Plan 40 (2026-06-13), Plan 41 (2026-06-16).
- **Result:** All routine-generated findings since 2026-05-07 that warranted Backlog entries are already captured:
  - Symlink hardening (P2 security) ✓
  - `bonsai validate` identity-drift warning (P2 improvement) ✓
  - Plan 40 review nits (P2 improvement) ✓
  - `bonsai validate` can't pass on Bonsai repo (P2 bug) ✓
  - Website npm vuln tree (P2 security) ✓
  - Unify remove business logic (P2 debt) ✓
  - Plan grilling integration (P2 feature) ✓
  - Full agent-drivable CLI parity (P1 feature) ✓
- **Issues:** None — all findings captured correctly.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed P0 and P1 items for any that are ready for immediate implementation without requiring user confirmation.
- **Result:** No items are approved for immediate autonomous promotion:
  - P0 is now empty of active items (both resolved).
  - P1 items each require user decision: CLI parity needs planning (`/plan`), PAT rotation needs user action, bot PR pile-up needs user config decision, testing debt and worktree cleanup need user prioritization.
- **Issues:** None. PAT expiry item flagged for user review (not routable to issue-to-implementation without user confirmation).

### Step 7: Log results
- **Action:** Appended entry to station/Logs/RoutineLog.md.
- **Result:** Done.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated routines.md dashboard row for Backlog Hygiene.
- **Result:** Done — Last Ran → 2026-06-28, Next Due → 2026-07-05, Status → done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | P0 bug `$PWD-walk-up` resolved in v0.4.3 but still listed as active P0 | Backlog.md P0 | Converted to HTML audit comment |
| 2 | HIGH | P0 feature `non-interactive flags` resolved in v0.4.2 but still listed as active P0 | Backlog.md P0 | Converted to HTML audit comment |
| 3 | HIGH | HOMEBREW_TAP_TOKEN PAT expires ~2026-07-22 — 17 days away, rotate before next release | Backlog.md P1 | Added `[TIME-SENSITIVE]` tag + urgency note; flagged for user |
| 4 | LOW | `[debt] Stale agent worktrees + branches` P1 item cites a 2026-04-21 snapshot — likely outdated after Plan 40/41 | Backlog.md P1 | Flagged for user re-check |
| 5 | LOW | `Template variables expansion` (Phase 2 roadmap milestone) has no Backlog entry | Roadmap.md Phase 2 | No action — Phase 2 not yet entered; low urgency |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **HOMEBREW_TAP_TOKEN PAT rotation (urgent)** — PAT rotated 2026-04-22 on a ~90-day cycle; expiry is approximately 2026-07-22. If a release is cut before rotation, the Homebrew formula update will fail silently (binaries publish but formula stale). Rotate the PAT in GitHub repo secrets before 2026-07-15 to be safe. See Backlog.md P1 `[ops]` item.

2. **Stale worktree/branch P1 item** — The `[debt] Stale agent worktrees + branches accumulating` P1 item was based on a 2026-04-21 audit (69 days ago). Post-Plan-41, the state may have changed. Recommend re-audit or close if already cleaned up. If the item remains valid, add a note with the current count.

## Notes for Next Run

- P0 section is now clean (all resolved items commented out, `Trial sentrux` correctly in Status.md Pending).
- PAT rotation is the most time-sensitive item — confirm it's handled before 2026-07-15.
- Website npm vuln (P2) has been in Backlog since 2026-06-16 — the next Vulnerability Scan routine should pick this up for resolution assessment.
- `[feature] Full agent-drivable CLI parity` (P1) was added 2026-06-13 per user priority ("main thing"). Next session should plan this via `/plan`.
- `Template variables expansion` (Phase 2 roadmap) has no Backlog entry — consider adding one when Phase 2 planning begins.
