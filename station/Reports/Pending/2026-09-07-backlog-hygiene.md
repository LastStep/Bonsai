---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-07
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
- **Files Read:** 6 — `station/agent/Core/identity.md`, `station/agent/Core/memory.md`, `station/agent/Routines/backlog-hygiene.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s
Read `Backlog.md` P0 section. Found **two active P0 items** — both were resolved by shipped releases but never removed from the backlog:

1. `[bug] Sensor hook commands use $PWD-walk-up` — fixed by v0.4.3 hotfix (2026-05-13, PRs #105/#106). The fix baked absolute install-time paths into each hook command. Item had been in backlog since 2026-05-13 with the note "Ships v0.4.3" but was never removed after ship.

2. `[feature] bonsai init / bonsai add need non-interactive flags` — fixed by v0.4.2 (2026-05-13, Plan 39). Both flags shipped with JSONL stdout and exit codes 0/2/3/4. Item had been in backlog since 2026-05-08 and unblocked Plan 38 rung-3.

**Action taken:** Both P0 items replaced with HTML resolution comments in Backlog.md. P0 section is now empty of active items.

Neither item needed escalation to Status.md (both were already shipped — they needed removal, not promotion).

### Step 2 — Cross-reference with Status.md
Read `Status.md`. Found **one additional resolved item** in P1:

- `[feature] Full agent-drivable CLI parity: init/update/add/remove` (added 2026-06-13) — explicitly requested "need a unified non-interactive surface + JSONL/exit-code contract across all four." Plan 41 (2026-06-16, PRs #120-#125, main `ab202c3`) delivered exactly this: every mutating cmd has a pure `*Result` headless core + JSONL/exit contract (`ExitConflict=5`), plus `list --json` and `docs/agent-interface.md`. The item remained in Backlog P1 after Plan 41 merged.

**Action taken:** P1 "Full agent-drivable CLI parity" item replaced with HTML resolution comment.

No Status.md Pending items have "Blocked By" that would be unblocked by a Backlog item. The one Pending item (`[research] Trial sentrux`) is blocked by Rust toolchain unavailability — no Backlog item addresses that.

### Step 3 — Cross-reference with Roadmap.md
Read `Roadmap.md`. Phase 1 is fully complete (all checkboxes checked). Phase 2 (Extensibility) is the next phase; relevant items in the backlog:

- P3 "Self-update mechanism" aligns with Phase 2 — could be promoted but no immediate trigger
- P3 "Micro-task fast path" aligns with Phase 2 — same

No Backlog items reference deprecated approaches or completed phases. The previously-stale "Better trigger sections" and "UI overhaul" items were already checked in Roadmap during the 2026-05-07 routine-digest.

**Action taken:** No promotions made (no user instruction or P0 urgency requiring it). Flagged for user review.

### Step 4 — Flag stale items
The backlog has been untouched since 2026-06-16 (Plan 41 items filed); the last hygiene run was 2026-05-07 — **~4 months gap**. Key staleness finding:

- **HOMEBREW_TAP_TOKEN PAT expiry (P1):** The item notes the PAT was rotated 2026-04-22 with a 90-day default, setting a reminder for ~2026-07-15. Today is 2026-09-07 — the PAT has almost certainly expired (54 days past its expected expiry). The next release will fail at the GoReleaser brew step with "401 Bad credentials." This is the most time-sensitive item in the backlog.

- Multiple Group B/C/D/E items are 4-5 months old with no movement — typical for this project's batch approach. No near-duplicates detected that weren't already flagged in prior runs.

**Action taken:** HOMEBREW_TAP_TOKEN staleness flagged for user review. No item removals for staleness alone.

### Step 5 — Check for routine-generated items
Reviewed `RoutineLog.md` entries since last backlog-hygiene (2026-05-07). Only significant post-date entries:
- 2026-06-13: Plan 40 dispatch — Backlog items filed: P2 symlink hardening, P2 validate drift warning, P2 Plan 40 review nits, P2 validate/lock-file bug, P2 website npm vulns, P2 unify remove logic. All are present in current Backlog.md — no uncaptured findings.

No routine runs occurred after 2026-05-07 (all routines are 100+ days overdue). No findings to capture.

**Action taken:** None needed — all flagged items are already in the Backlog.

### Step 6 — Promote ready items via issue-to-implementation
No item meets the criteria for autonomous promotion. HOMEBREW_TAP_TOKEN is the most urgent but requires the user to rotate the PAT (agent cannot act on GitHub secrets). Flagging for user review.

### Step 7 — Log results
Appended entry to `station/Logs/RoutineLog.md`.

### Step 8 — Update dashboard
Updated `station/agent/Core/routines.md` dashboard row for Backlog Hygiene.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | P0 bug "Sensor hook $PWD-walk-up" was fixed in v0.4.3 (2026-05-13) but never removed from backlog | Backlog.md P0 | Removed — replaced with HTML resolution comment |
| 2 | HIGH | P0 feature "bonsai init/add non-interactive flags" was fixed in v0.4.2 (2026-05-13) but never removed from backlog | Backlog.md P0 | Removed — replaced with HTML resolution comment |
| 3 | MEDIUM | P1 feature "Full agent-drivable CLI parity" was resolved by Plan 41 (2026-06-16) but never removed from backlog | Backlog.md P1 | Removed — replaced with HTML resolution comment |
| 4 | HIGH | HOMEBREW_TAP_TOKEN PAT ~54 days past expected expiry date (~2026-07-15); next release will fail at brew step | Backlog.md P1 | Flagged for user review — cannot auto-rotate |
| 5 | LOW | All routines are 100-119 days overdue (last ran: 2026-05-04 to 2026-05-07) | routines.md dashboard | Noted — outside scope of this routine |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **HOMEBREW_TAP_TOKEN PAT is almost certainly expired.** Added 2026-04-22; 90-day default expiry ~2026-07-15; today is 2026-09-07 (54 days overdue). **Action required:** rotate the fine-grained PAT on GitHub, update the `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai`, and set a new calendar reminder for 90 days out. Symptom of expired PAT: GoReleaser fails at brew step with "GET https://api.github.com/repos/LastStep/homebrew-tap: 401 Bad credentials" — release otherwise succeeds (binaries published, only formula update missed). The Backlog P1 item itself remains valid as a reminder.

2. **All routines are 100-119 days overdue.** Dependency Audit, Doc Freshness Check, Vulnerability Scan, Memory Consolidation, Status Hygiene, and Roadmap Accuracy have not run since early May 2026. Recommend a routine-digest session to catch up.

## Notes for Next Run

- P0 section is now clean. Both items were resolved by shipped releases; the gap between ship and removal was 3-4 months — consider adding a post-release checklist item to sweep resolved Backlog P0/P1s.
- HOMEBREW_TAP_TOKEN: if user rotates the PAT, update the Backlog P1 item's date and set a new ~2026-12-06 reminder note in the item text.
- The "Routine bot PR pile-up" P1 item (bot PRs accumulating without merge) was flagged 2026-05-07 — 4 months with no fix. Consider promoting to plan if the bot is still active.
