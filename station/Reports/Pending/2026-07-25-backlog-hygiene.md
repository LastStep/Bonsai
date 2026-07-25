---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-25
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
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write, Glob, Bash
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; cross-referenced each P0 item against Status.md In Progress and Recently Done.
- **Result:** Both P0 items are RESOLVED and should not be in the P0 section:
  - `[bug] Sensor hook commands use $PWD-walk-up` — shipped as v0.4.3 (PRs #105/#106, 2026-05-13). Status.md "Recently Done" confirms: "sensor hook commands now bake install-time absolute paths." Commented out with audit trail.
  - `[feature] bonsai init / bonsai add need non-interactive flags` — shipped as v0.4.2 (PR #102, 2026-05-13). Status.md "Recently Done" confirms `--non-interactive --from-config` on init and add. Commented out with audit trail.
  - After removal, P0 section is now empty (only HTML comment audit trail remains). This is a healthy state.
- **Issues:** None — both items cleanly traced to shipped releases.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress, Pending, and Recently Done. Matched against all Backlog items.
- **Result:**
  - **Resolved — P1 CLI parity item removed:** `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` was added 2026-06-13. Plan 41 (shipped 2026-06-16, PRs #120/#122/#123/#121/#125) delivered exactly this: "every mutating cmd (init/add/update/remove) has a pure `*Result` headless core + JSONL/exit contract." Commented out with audit trail.
  - **Status.md Pending cross-check:** Only pending item is `[research] Trial sentrux` (blocked on Rust toolchain). No backlog item can directly unblock this — it requires user/ops action to install rustup. No change needed.
  - **Blocked By chain check:** No backlog items appear to unblock the sentrux trial directly.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md. Compared P2/P3 items against current phase milestones.
- **Result:**
  - Phase 1 is fully complete (all checkboxes marked `[x]`). No stale phase-reference issues in backlog items for Phase 1.
  - Phase 2 (Extensibility) has three unchecked milestones: Self-update mechanism, Template variables expansion, Micro-task fast path. Three P3 backlog items map to these exactly (`[improvement] Self-update mechanism`, `[improvement] Micro-task fast path`, `[feature] Custom item creator`). These warrant promotion consideration as Phase 2 becomes active, but no user signal yet to promote — flagging for awareness rather than acting.
  - No backlog items reference deprecated approaches or completed phases (except the two P0 items now cleaned up).
- **Issues:** None requiring immediate action.

### Step 4: Flag stale items
- **Action:** Scanned all P0–P3 items for items at same priority 30+ days without progress. Last run was 2026-05-07; today is 2026-07-25 (79 days elapsed).
- **Result:** Most items in the backlog are technically stale by age, but staleness is more meaningful when combined with urgency. Key flags:

  1. **URGENT — `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` (P1):** Added 2026-04-22 with explicit reminder date of ~2026-07-15. Today is 2026-07-25 — **10 days past the reminder date**. If the PAT has not been rotated, the next GoReleaser release will fail at the Homebrew brew step with a 401. Requires user action.
  2. **`[ops] Routine bot PR pile-up — eliminate parallel-track artifacts` (P1):** Added 2026-05-07. 9 stale PRs were closed as a cleanup measure but the root cause (cloud routine creating daily PRs) was never fixed. 79 days stale. Low urgency if no cloud routine is actively running, but worth confirming.
  3. **`[security] Website npm vuln tree — astro upgrade breaks npm run build` (P2):** Added 2026-06-16. A security-tagged item with an active build break — 39 days without action. Should be escalated soon.
  4. **Group B items (code quality/testing):** All filed April 2026, no progress visible. These are debt items — stale but not urgent. Noted for next roadmap review.
  5. **`[debt] Stale agent worktrees + branches accumulating` (P1):** Added 2026-04-20. No resolution noted in RoutineLog or Status. The underlying issue (merged PR branches not being auto-deleted) may still be accumulating.
- **Items with no clear rationale:** None — all items carry sufficient context.
- **Near-duplicates:** None identified beyond the previously-noted Plan 24 CHANGELOG item (already filed as refiled good-first-issue).
- **Issues:** PAT expiry flag requires user attention.

### Step 5: Check for routine-generated items
- **Action:** Read recent RoutineLog.md entries since 2026-05-07 (last backlog-hygiene run).
- **Result:** No routine-specific log entries exist after 2026-05-07. The Plan 40 dispatch entry from 2026-06-13 is a plan dispatch (not a routine), and it did produce new backlog items (security hardening P2, improvement P2, Plan 40 review nits P2, validate dogfood P2) — all are already captured in Backlog.md. No uncaptured routine findings.
  - **Note:** All routines show last_ran dates in early May 2026 (79+ days overdue). This is a significant gap — no routines have run since the last dispatch cycle. The Tech Lead should be aware.
- **Issues:** Routine execution gap (all routines overdue) flagged for user awareness.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed whether any items are ready for immediate promotion or implementation.
- **Result:** No items are immediately actionable without user confirmation. Flagging the PAT expiry (Step 4, item 1) as the most time-sensitive item. The HOMEBREW_TAP_TOKEN rotation is a user/ops task, not an agent task. No issue-to-implementation workflow triggered.
- **Issues:** None — no items promoted autonomously.

### Step 7 & 8: Log and dashboard update
- **Action:** Appended entry to RoutineLog.md; updated routines.md dashboard.
- **Result:** Done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | P0 bug (sensor hook $PWD walk-up) already shipped in v0.4.3 — stale entry | Backlog.md P0 | Commented out with audit trail |
| 2 | High | P0 feature (non-interactive flags) already shipped in v0.4.2 — stale entry | Backlog.md P0 | Commented out with audit trail |
| 3 | High | P1 full CLI parity already shipped via Plan 41 — stale entry | Backlog.md P1 | Commented out with audit trail |
| 4 | High | HOMEBREW_TAP_TOKEN PAT reminder date (2026-07-15) has passed — rotation overdue | Backlog.md P1 | Flagged for user — user action required |
| 5 | Medium | Website npm vuln / astro build break (security P2) 39 days without action | Backlog.md P2 | Flagged for user attention |
| 6 | Medium | All routines overdue (79+ days since last run) | routines.md dashboard | Noted — this run is part of the recovery |
| 7 | Low | Phase 2 milestones (self-update, micro-task fast-path, custom item creator) have P3 backlog items eligible for promotion when Phase 2 begins | Backlog.md P3 | Noted — no promotion without user signal |
| 8 | Low | Routine bot PR pile-up fix (P1) — root cause not resolved 79 days after filing | Backlog.md P1 | Flagged for user awareness |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **URGENT — HOMEBREW_TAP_TOKEN PAT rotation:** The reminder date of ~2026-07-15 has passed (today is 2026-07-25). If the PAT has not been rotated, the next GoReleaser release will fail at the Homebrew tap step with a 401. Check and rotate the `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai` via GitHub → Settings → Secrets. The original PAT was rotated 2026-04-22 (90-day default expiry → expired ~2026-07-21).

2. **Website npm vulnerability (P2 security):** `[security] Website npm vuln tree — astro upgrade breaks npm run build` has been in P2 for 39 days. This involves a failing build fix (PR #108 rebase needed) plus vite/js-yaml bumps. Consider scheduling a session to address this.

3. **Routine execution gap:** All seven routines show last_ran in early May 2026. The 79-day gap means no dependency audits, vulnerability scans, doc freshness checks, or status hygiene have run. Consider running all overdue routines in a routine-digest session.

4. **Phase 2 readiness check:** Three P3 items map directly to Phase 2 milestones. When the user is ready to start Phase 2 work, these are the starting candidates: `[improvement] Self-update mechanism`, `[improvement] Micro-task fast path`, `[feature] Custom item creator`.

## Notes for Next Run

- P0 section is now clean (only audit-trail comments remain) — healthy state.
- P1 full CLI parity is resolved — Plan 41 shipped the complete headless contract.
- PAT expiry is the most time-sensitive open item; confirm rotation before cutting next release.
- If all seven routines are run in the coming sessions, the next backlog-hygiene run will have more routine-generated findings to cross-check.
- The backlog is large (50+ items across P1–P3) and many are 90+ days old. Consider a dedicated backlog pruning session to apply NoteStandards (Group A item) and retire items that are no longer relevant.
