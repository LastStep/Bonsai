---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-21
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
- **Duration:** ~10 min
- **Files Read:** 5 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read (5×), Python string replacement (3×), Write (report), Edit (dashboard + log)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section; cross-referenced each P0 with `Status.md`.
- **Result:** Found 2 P0 items — BOTH already resolved and present in Status.md "Recently Done":
  1. `[bug] Sensor hook commands use $PWD-walk-up` — Fixed in v0.4.3 hotfix (PR #105/#106, 2026-05-13). Backlog still showed this as an active P0.
  2. `[feature] bonsai init / bonsai add need non-interactive flags` — Fixed in v0.4.2 (Plan 39, 2026-05-13). Backlog still showed this as an active P0.
- **Issues:** Both P0s were stale. Both removed (converted to HTML comments with audit trail). P0 section is now empty of active items.

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md` In Progress and Recently Done. Matched against Backlog items.
- **Result:**
  - **Confirmed resolved and removed from backlog:** P1 `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` — Plan 41 shipped all 5 phases on 2026-06-16 (PRs #120/#122/#123/#121/#125); all four commands now have headless `*Result` cores + JSONL/exit contract. This P1 item was added 2026-06-13 as the driver for Plan 41 and is fully resolved.
  - **Status.md Pending:** Only 1 item — `[research] Trial sentrux on Bonsai repo` (blocked on Rust toolchain). Already commented out in Backlog P0 (promoted 2026-05-07). No action needed.
  - **No Blocked-By unblocking opportunities** found between Pending items and Backlog items. The sentrux item remains blocked on Rust toolchain (no Backlog item resolves that).
- **Issues:** None — all cross-references clean after removals.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md` and compared P2/P3 Backlog items against current phase milestones.
- **Result:**
  - Phase 1 is **fully complete** (all checkboxes checked). No stale phase-1 references found in Backlog that weren't already cleaned up in prior runs.
  - Phase 2 (Extensibility) has 3 open items: self-update mechanism, template variables expansion, micro-task fast path. Matching Backlog items exist at P3 (Big Bets / Future Platform). None warrant promotion to P1 without user direction — Phase 2 has no active timeline set.
  - No deprecated approach references found. The P0/P1 items removed above referenced completed work but were active bullets, not deprecated-approach references.
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Reviewed all Backlog items for age, lack of progress, and clarity.
- **Result:**
  - **URGENT — HOMEBREW_TAP_TOKEN PAT likely expired (P1 item):** The PAT was rotated 2026-04-22 with a 90-day default expiry. 90 days = expiry ~2026-07-21, which is TODAY. The calendar reminder was set for ~2026-07-15 and has passed with no visible action. If a release is attempted now, GoReleaser will fail at the Homebrew formula step with `401 Bad credentials`. **Flagged for immediate user action.**
  - **Stale: P1 `[debt] Stale agent worktrees + branches accumulating`** — Added 2026-04-20, updated 2026-04-21. The "Suggested" one-time sweep was documented but no entry in RoutineLog or Status confirms it was performed. 90+ days old. May be partially resolved or recurring. Flagged for user review.
  - No items with no clear context or rationale found.
  - **Near-duplicates check:** No new near-duplicates found. Prior routine runs already noted the changelog near-duplicate (Group C/D). Still present but intentionally kept per previous review.
- **Issues:** 1 urgent (PAT expiry), 1 low (stale worktrees item age).

### Step 5: Check for routine-generated items
- **Action:** Read recent RoutineLog entries since last backlog-hygiene run (2026-05-07).
- **Result:** No routine executions recorded in RoutineLog between 2026-05-07 and 2026-07-21 (10+ weeks). All routines are significantly overdue per the dashboard. The only log entries in that window were plan dispatches (Plan 40, Plan 41), not routine executions. No routine-generated findings are pending capture in the Backlog.
- **Issues:** The gap in routine execution (10+ weeks) is itself notable but is not a Backlog item — it is the state of the routine system, and these runs are now executing to catch up.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed whether any Backlog items are approved or ready for immediate promotion.
- **Result:** No items currently approved for autonomous implementation. The HOMEBREW_TAP_TOKEN PAT issue requires user action (manual secret rotation on GitHub). All other flagged items require user decision before promotion.
- **Issues:** None autonomous-actionable.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for Backlog Hygiene.
- **Result:** `Last Ran` → 2026-07-21, `Next Due` → 2026-07-28, `Status` → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | **high** | HOMEBREW_TAP_TOKEN PAT rotated 2026-04-22; 90-day default expiry = today (2026-07-21). Releases will fail Homebrew formula step with 401 until PAT is rotated. | `station/Playbook/Backlog.md` P1 item; GitHub repo secrets | Flagged for user — PAT rotation requires manual user action on GitHub |
| 2 | **low** | P0 bug `[sensor hook $PWD-walk-up]` still in Backlog as active item despite being fixed in v0.4.3 (2026-05-13) | `station/Playbook/Backlog.md` P0 | Removed — converted to HTML comment with audit trail |
| 3 | **low** | P0 feature `[bonsai init/add --non-interactive flags]` still in Backlog as active item despite being fixed in v0.4.2 (2026-05-13) | `station/Playbook/Backlog.md` P0 | Removed — converted to HTML comment with audit trail |
| 4 | **low** | P1 `[Full agent-drivable CLI parity]` still in Backlog despite Plan 41 shipping full headless contract 2026-06-16 | `station/Playbook/Backlog.md` P1 | Removed — converted to HTML comment with audit trail |
| 5 | **info** | P1 `[Stale agent worktrees + branches]` added 2026-04-20 — 90+ days old with no Status entry confirming the one-time sweep was performed | `station/Playbook/Backlog.md` P1 | Flagged for user — confirm whether sweep was done; update or remove item if resolved |
| 6 | **info** | No routines have run since 2026-05-07 (10+ weeks). RoutineLog shows no entries in the gap. | `station/Logs/RoutineLog.md` | Context only — this backlog-hygiene run is part of the catch-up cycle |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **URGENT — Rotate HOMEBREW_TAP_TOKEN PAT now.** The PAT was rotated 2026-04-22 with a default 90-day expiry. It may have expired today (2026-07-21). Go to GitHub → `LastStep/Bonsai` → Settings → Secrets → Rotate `HOMEBREW_TAP_TOKEN`. Generate a new fine-grained PAT, set a calendar reminder for the next rotation (~90 days out = ~2026-10-19). If you are planning a release soon, do this before the release attempt.
- **Confirm stale-worktrees sweep status.** The P1 `[debt] Stale agent worktrees + branches accumulating` item (added 2026-04-20) documents a one-time sweep as the fix. If that sweep was done, remove the item. If not done, consider scheduling it — the accumulation may have grown over 3 months.

## Notes for Next Run

- P0 section is now entirely HTML comments (audit trail only). If a new critical bug surfaces, this section will be repopulated.
- P1 section is healthier post-cleanup. The two remaining P1 items ([ops] Homebrew PAT, [ops] Routine bot PR pile-up) are ongoing operational items that don't resolve by shipping code.
- Routine coverage gap (10+ weeks) should normalize now that the loop is running. Next backlog-hygiene is due 2026-07-28.
- If the HOMEBREW_TAP_TOKEN PAT is rotated, update or remove the P1 Backlog item.
