---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-09
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
- **Files Read:** 5 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 1 — `/home/user/Bonsai/station/Playbook/Backlog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section. Found 2 live P0 items. Cross-checked each against Status.md In Progress and Pending tables.
- **Result:** Neither P0 item appeared in Status.md as In Progress or Pending. However, both items were found in Status.md "Recently Done" — meaning they are resolved, not merely untracked. (Step 2 handled removal.)
- **Issues:** none beyond confirming the items were stale resolved entries, not escalation candidates.

### Step 2: Cross-reference with Status.md
- **Action:** Compared all Backlog items against Status.md In Progress, Pending, and Recently Done tables.
- **Result:** Identified 3 items resolved by work done since the last hygiene run (2026-05-07):
  1. **P0 `[bug]` Sensor hook `$PWD`-walk-up** — resolved by v0.4.3 hotfix (PRs #105/#106, 2026-05-13). Commented out.
  2. **P0 `[feature]` `bonsai init`/`add` non-interactive flags** — resolved by v0.4.2 (`--non-interactive --from-config` shipped, 2026-05-13). Commented out.
  3. **P1 `[feature]` Full agent-drivable CLI parity** — resolved by Plan 41 (PRs #120–#125, 2026-06-16). Plan 41 delivered headless `*Result` cores + JSONL/exit contract for all four cmds (init/add/update/remove). Commented out.
- **Status.md Pending cross-check:** One Pending item exists — `[research] Trial sentrux on Bonsai repo`, blocked on Rust toolchain. No Backlog item can unblock it directly (the blocker is environment-level: rustup/cargo not installed). Already correctly commented out in Backlog.
- **Issues:** none beyond the three stale items removed.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md. Reviewed Phase 1 (complete), Phase 2, Phase 3, Phase 4.
- **Result:**
  - Phase 1 is fully checked off. No stale Backlog items referencing deprecated Phase 1 work found.
  - Phase 2 "Template variables expansion" has **no corresponding Backlog entry**. This is a Phase 2 milestone with zero tracking. Flagged for user (see Items Flagged for User Review).
  - Phase 2 "Self-update mechanism" and "Micro-task fast path" exist in Backlog P3 "Future Platform" — both appropriately deferred.
  - Phase 2 "Custom item detection" is checked ([x]) and has no live Backlog entry — correct.
  - Phase 3 and Phase 4 items are in Backlog P3 "Big Bets" and "Future Platform" — alignment healthy.
- **Promotion candidates:** The "Micro-task fast path" (P3 Backlog) aligns with Phase 2 but is appropriately low-priority; no promotion recommended without user signal.
- **Issues:** 1 — Phase 2 "Template variables expansion" untracked.

### Step 4: Flag stale items
- **Action:** Scanned all P0–P3 items for age, context clarity, and near-duplicates.
- **Result:**
  - **HOMEBREW_TAP_TOKEN PAT (P1):** Added 2026-04-22. PAT was rotated 2026-04-22 with default 90-day expiry — estimated expiry **~2026-07-21**. The calendar reminder target of ~2026-07-15 has passed without any action logged. Today is 2026-08-09: the PAT is likely already expired. The last release (v0.4.3, 2026-05-13) succeeded while the PAT was still valid. The next release (v0.5.0, tag held) will likely fail at the brew step. **Urgent — flagged for user.**
  - **Stale worktrees item (P1):** Added 2026-04-21. The count of "17+ worktrees" and "20+ stale remote branches" is now 110+ days old and reflects a point-in-time snapshot. The item is still valid as a standing hygiene item but its data is stale. Flagged for refresh.
  - **Go module refresh (P3):** Added 2026-04-21, updated 2026-05-04. Lists specific version ranges (e.g., `x/crypto v0.36 → v0.50`) from a 2026-05-04 scan — 97 days ago. Current deltas will be larger. Flagged for refresh.
  - **Many items 30+ days old without progress:** The bulk of Backlog items (added April–June 2026) have no activity for 54–115 days. This is expected for P2/P3 items during an active P0/P1 execution phase. No individual flags beyond the above.
  - **Near-duplicates:** None found beyond what prior hygiene runs already noted (the CHANGELOG/Group C duplicate from April 2026 was resolved — Group C no longer contains that item).
- **Issues:** 3 items flagged (PAT expiry, stale worktrees data, stale Go module versions).

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since 2026-05-07.
- **Result:** No routine log entries between 2026-05-07 and 2026-08-09 (gap of 94 days — all routines overdue). The only post-May entry in RoutineLog.md is the Plan 40 session dispatch (2026-06-13), which is not a routine. That session did file new Backlog items (6 P2 entries from Plan 40 grill + Plan 41 review), all of which are present in Backlog.md. No uncaptured routine findings to add.
- **Issues:** none — all session-filed items are accounted for.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any item is approved for immediate implementation.
- **Result:** No items have explicit user approval for immediate implementation. The HOMEBREW_TAP_TOKEN PAT requires user action (credential rotation) — not agent-implementable. No autonomous issue-to-implementation routing triggered.
- **Issues:** none.

### Step 7: Log results
- **Action:** Entry appended to `station/Logs/RoutineLog.md`.

### Step 8: Update dashboard
- **Action:** `routines.md` Backlog Hygiene row updated: Last Ran → 2026-08-09, Next Due → 2026-08-16, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 sensor hook `$PWD`-walk-up — stale, fixed in v0.4.3 | Backlog P0 | Commented out |
| 2 | resolved | P0 non-interactive flags — stale, shipped in v0.4.2 | Backlog P0 | Commented out |
| 3 | resolved | P1 Full agent-drivable CLI parity — stale, shipped in Plan 41 | Backlog P1 | Commented out |
| 4 | high | HOMEBREW_TAP_TOKEN PAT expired ~2026-07-21 — next release will fail brew step | Backlog P1 | Flagged for user |
| 5 | medium | Phase 2 "Template variables expansion" has no Backlog entry | Roadmap.md | Flagged for user |
| 6 | low | Stale agent worktree count (110+ days old snapshot) | Backlog P1 | Flagged for user |
| 7 | low | Go module version ranges in P3 item are 97+ days stale | Backlog P3 | Flagged for user |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

### [URGENT] HOMEBREW_TAP_TOKEN PAT has likely expired
The PAT rotated on 2026-04-22 had a default 90-day expiry, putting expiry at ~2026-07-21. The calendar reminder target of ~2026-07-15 passed without any rotation logged in RoutineLog.md. The last Homebrew-touching release was v0.4.3 (2026-05-13, within the 90-day window). The next release (v0.5.0 tag is held) will fail at the GoReleaser brew step with `401 Bad credentials` if the PAT is not rotated first.

**Recommended action:** Rotate the `HOMEBREW_TAP_TOKEN` fine-grained PAT in GitHub repo secrets (`LastStep/Bonsai`) before cutting the next release tag. Set a new calendar reminder ~85 days from rotation date.

### Phase 2 "Template variables expansion" has no Backlog entry
Roadmap.md Phase 2 lists "Template variables expansion — richer context available in templates" as a milestone. No Backlog item tracks this. Consider adding a P2 item or explicitly deferring it to Phase 3.

### Stale worktrees item should be refreshed
The P1 `[debt] Stale agent worktrees + branches accumulating` item cites a 2026-04-21 count of "17+ worktrees, 20+ stale remote branches." This data is 110+ days old. The actual counts have likely changed (worktrees from Plans 38–41 added since then). Recommend refreshing the item with current data, or closing it if the sweep has been done.

### Go module versions in P3 item are outdated
The P3 `[debt] Batch refresh outdated Go modules after toolchain upgrade` item lists specific `go list -m -u all` version delta data from 2026-05-04. That data is 97 days stale. The deltas will be larger now. Update with a fresh scan before scheduling the work.

---

## Notes for Next Run

- P0 section is now empty of live items (only HTML comments remain). If this state persists, consider whether to keep the P0 header or collapse it.
- 94-day gap since last run (2026-05-07 → 2026-08-09) — all other routines are also overdue. A routine-digest pass after this sweep would be valuable.
- The Status.md "Recently Done" table is dense with Plan 40 and Plan 41 work from June 2026. A status-hygiene run should age out items older than 14 days.
- Plan 41 filed the `[debt] Unify remove business logic` P2 item (already in Backlog) — keep tracking.
