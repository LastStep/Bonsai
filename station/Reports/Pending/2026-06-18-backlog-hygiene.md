---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-06-18
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
- **Duration:** ~6 min
- **Files Read:** 5 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read (5×), Edit (3×), Write (1×)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section and cross-referenced against Status.md In Progress / Pending.
- **Result:** Found 2 stale P0 items that had already been shipped:
  - `[bug] Sensor hook commands use $PWD-walk-up` — RESOLVED by v0.4.3 hotfix (2026-05-13, PRs #105/#106). Fix was baking absolute install-time paths into hook commands.
  - `[feature] bonsai init / bonsai add need non-interactive flags` — RESOLVED by v0.4.2 release (2026-05-13, PR #102). `--non-interactive` + `--from-config` shipped.
  - Both removed from Backlog P0 and replaced with dated HTML comments.
  - `[research] Trial sentrux` was already commented out in P0 and correctly placed in Status.md Pending — no action needed.
- **Issues:** None.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md (In Progress, Pending, Recently Done) and cross-referenced against all Backlog items.
- **Result:** Found 1 resolved P1 item:
  - `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` (added 2026-06-13) — RESOLVED by Plan 41 (shipped 2026-06-16, PRs #120/#122/#123/#121/#125). All four mutating commands now have headless `*Result` cores + JSONL/exit contract (`ExitConflict=5`); `list --json` also added; `docs/agent-interface.md` contract doc shipped.
  - Removed from Backlog P1 and replaced with dated HTML comment.
  - `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — PAT was rotated 2026-04-22, due ~2026-07-15. Still relevant; remains in Backlog P1.
  - `[ops] Routine bot PR pile-up` — 9 PRs closed 2026-05-07 but structural fix (commit-direct or auto-merge) not yet implemented. Remains in Backlog P1.
  - No Pending items in Status.md have "Blocked By" entries that a Backlog item could unblock (sentrux trial is blocked on Rust toolchain install, not a Backlog item).
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md and cross-referenced Backlog P2/P3 items against phase milestones.
- **Result:**
  - Phase 1 is fully checked — all items complete. No Backlog items reference deprecated Phase 1 approaches.
  - Phase 2 milestones (self-update, template vars, micro-task fast path) each have matching Backlog entries in P3.
  - Phase 3 milestones (Managed Agents, Greenhouse app) each have matching Backlog entries in P3 Big Bets.
  - No P2/P3 items identified as ready to promote to P1 based on current roadmap phase (Phase 1 complete, Phase 2 not yet started formally).
  - The `[feature] Integrate plan-grilling as a first-class Bonsai catalog ability` (P2) is a good candidate for Phase 2 work but the user should decide — flagged for awareness, not promoted autonomously.
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Reviewed all items for age, clarity, and near-duplicates.
- **Result:**
  - Several items are 42–65 days old (added 2026-04-13 through 2026-04-16) with no progress. These are mostly P3 ideas and research items (Big Bets, Future Platform, Research). They have clear rationale and intent — not stale due to neglect, but because they are intentionally lower-priority. No items are ambiguous or context-free.
  - The `HOMEBREW_TAP_TOKEN` PAT (P1, added 2026-04-22) has a specific due date of ~2026-07-15 — that deadline is **27 days away**. Flagging for user attention.
  - No true near-duplicates found after P0/P1 cleanup. The earlier `bonsai init / bonsai add non-interactive flags` and `full CLI parity` items were near-duplicates but both are now resolved and commented out.
  - Group A (`[bookkeeping] Retroactively trim Backlog entries`) is itself a stale item (added 2026-04-25) — still valid since this file still has verbose prose in many entries.
- **Issues:** PAT expiry approaching — flagged for user review.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07).
- **Result:** Entries since 2026-05-07:
  - 2026-06-13 Plan 40 dispatch — findings about `bonsai validate` breaking (gitignored lock), symlink hardening gap, and Plan 40 nits: ALL captured in Backlog as P2 items added 2026-06-13 (lines 67-70 of Backlog.md). Verified captured.
  - 2026-06-16 Plan 41 dispatch — `[security] Website npm vuln tree` and `[debt] Unify remove business logic` both filed to Backlog P2 on 2026-06-16. Verified captured.
  - No routine report findings found that are missing from Backlog (all routine-generated findings from the 2026-05-04 and 2026-05-07 cycles were already captured or actioned in prior hygiene runs).
- **Issues:** None uncaptured.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed whether any items are approved or urgent enough to route to issue-to-implementation.
- **Result:** No items are approved for immediate implementation. The `HOMEBREW_TAP_TOKEN` PAT renewal (P1) is approaching but is a user action (rotate PAT in GitHub secrets), not an agent-implementable task. Flagged to user.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` Backlog Hygiene row.
- **Result:** `Last Ran` → 2026-06-18, `Next Due` → 2026-06-25, `Status` → `done`.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | P0 `[bug] Sensor hook $PWD-walk-up` was resolved by v0.4.3 (2026-05-13) but still in Backlog P0 | `Backlog.md` P0 section | Removed — replaced with dated HTML comment |
| 2 | high | P0 `[feature] bonsai init/add non-interactive flags` was resolved by v0.4.2 (2026-05-13) but still in Backlog P0 | `Backlog.md` P0 section | Removed — replaced with dated HTML comment |
| 3 | high | P1 `[feature] Full agent-drivable CLI parity` was resolved by Plan 41 (2026-06-16) but still in Backlog P1 | `Backlog.md` P1 section | Removed — replaced with dated HTML comment |
| 4 | medium | `HOMEBREW_TAP_TOKEN` PAT expiry approaching (~2026-07-15, 27 days away) | `Backlog.md` P1 | Flagged for user — requires manual PAT rotation in GitHub secrets |
| 5 | low | Group A bookkeeping item (`[bookkeeping] Retroactively trim Backlog entries`) is 54 days old with no progress | `Backlog.md` Group A | Left in place — low-priority housekeeping, no urgency |
| 6 | info | P2 `[feature] Integrate plan-grilling` is a natural Phase 2 candidate | `Backlog.md` P2 | Noted for awareness — no autonomous promotion |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **PAT rotation due ~2026-07-15 (27 days):** `HOMEBREW_TAP_TOKEN` fine-grained PAT on `LastStep/Bonsai` was rotated 2026-04-22. 90-day default expiry puts deadline at ~2026-07-15. Action: rotate before that date via GitHub repo Settings > Secrets > `HOMEBREW_TAP_TOKEN`. If missed, GoReleaser brew step will fail on next release (binaries still publish, only Homebrew formula update fails).

## Notes for Next Run

- P0 section is now clear — if it stays clear next run, consider noting in report that P0 section is clean.
- The `[ops] Routine bot PR pile-up` P1 item has no structural fix yet — check whether cloud routine dispatch behavior has changed.
- The P2 website npm vuln (`[security] Website npm vuln tree — astro upgrade breaks npm run build`) was added 2026-06-16 and may have progressed by next run — check PRs #108 status.
- Next run should verify whether the Plan 40 dogfood blocker (`.bonsai-lock.yaml` gitignored) has been resolved.
