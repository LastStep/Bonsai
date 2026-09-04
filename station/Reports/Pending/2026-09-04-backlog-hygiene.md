---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-04
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
- **Files Read:** 5 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 3 — `Playbook/Backlog.md` (2 resolved P0s commented out), `agent/Core/routines.md` (dashboard updated), `Logs/RoutineLog.md` (entry appended)
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section; checked both P0 items against `Status.md` In Progress and Pending tables.
- **Result:** Both P0 items are **resolved** — not misplaced escalations but stale entries that should have been cleaned up. (1) `[bug] Sensor hook commands use $PWD-walk-up` is resolved in v0.4.3 hotfix (Status.md confirms: "sensor hook commands now bake install-time absolute paths," PR #105/#106, commit `584b82b`). (2) `[feature] bonsai init/add need non-interactive flags` is resolved in v0.4.2 (Status.md confirms: "--non-interactive --from-config shipped," PR #102, commit `410a5f1`). Neither item exists in Status.md In Progress or Pending because both are already in Recently Done (as shipped releases). P0 section is now empty; no live escalations required.
- **Issues:** None — both items resolved, cleaned up by commenting out with resolution notes (matching existing backlog convention).

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md`; cross-referenced all Backlog items against In Progress, Pending, and Recently Done tables.
- **Result:** The two P0 items removed (see Step 1). `[research] Trial sentrux` remains correctly in both Backlog (commented out as promoted) and Status.md Pending. P1 item `[feature] Full agent-drivable non-interactive CLI parity` (added 2026-06-13) is adjacent to Plan 41 (shipped 2026-06-16 — "Every mutating cmd init/add/update/remove has a pure *Result headless core + JSONL/exit contract"). The backlog item may be substantially or fully resolved by Plan 41; scope overlap is high but not identical (the item also calls for a `--from-config` overlay surface for update/remove). Flagged for user review. No Status.md Pending items with "Blocked By" could be unblocked by a Backlog item (Sentrux trial remains blocked on Rust toolchain, a system-level dependency not in the Backlog).
- **Issues:** One P1 potentially resolved by Plan 41 — flagged for user review.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md`; identified current phase and checked P2/P3 backlog alignment.
- **Result:** Phase 1 is fully complete (all checkboxes checked). Project has moved to Phase 2 (Extensibility). Two P3 items directly mirror Phase 2 milestones: (1) `[improvement] Self-update mechanism` maps to Phase 2 "Self-update mechanism" milestone; (2) `[feature] Micro-task fast path` maps to Phase 2 "Micro-task fast path" milestone. Both are candidates for promotion from P3 → P2 given Phase 2 is now active. No Backlog items reference deprecated Phase 1 approaches (all Phase 1-tagged items are either resolved, properly closed, or remain valid). No items reference completed phases incorrectly.
- **Issues:** Two P3 items are Phase 2 milestone candidates — flagged for user review.

### Step 4: Flag stale items
- **Action:** Scanned all items for 30+ day stale indicators (added dates vs 2026-09-04) and items with unclear rationale; checked for near-duplicates.
- **Result:** 
  - **CRITICAL STALE: `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder`** — Added 2026-04-22 with explicit reminder for `~2026-07-15`. Today is 2026-09-04 — the reminder date has passed by **51 days**. If the PAT was not rotated, it has expired (90-day default). Symptom: GoReleaser fails at brew step with `401 Bad credentials`. Immediate user action required before next release.
  - **Stale P1 — `[ops] Routine bot PR pile-up`** — Added 2026-05-07, no resolution in ~4 months. Three fix options were documented (commit-direct-to-main, auto-merge, skip-if-digest-ran). Decision still pending.
  - **Stale P1 — `[debt] Stale agent worktrees + branches`** — Added 2026-04-20, no resolution in ~4.5 months. Suggested one-time sweep with `git worktree remove -f -f` + branch cleanup. The debt may be larger now.
  - **Stale P1 Group B** — All 8 Code Quality & Testing items have been at P1 since 2026-04-16 (4.5 months). No progress noted in Status.md. Collectively represent significant testing debt. Recommend re-prioritization review: either schedule a dedicated testing sprint or demote to P2 to reduce false urgency.
  - No near-duplicates found that weren't already accounted for (the two cleaned-up P0s were the closest, and both are resolved).
- **Issues:** 1 critical stale (HOMEBREW PAT), 2 stale P1s, 1 stale Group B cluster — all flagged for user review.

### Step 5: Check for routine-generated items
- **Action:** Read `Logs/RoutineLog.md`; reviewed entries since last backlog-hygiene run (2026-05-07).
- **Result:** No routine entries exist after 2026-05-07. The RoutineLog shows the last entries are from the 2026-05-07 batch (Status Hygiene, Roadmap Accuracy, Backlog Hygiene, Memory Consolidation) and the 2026-05-04 batch. No routines have run between 2026-05-07 and 2026-09-04. There are therefore no routine-generated findings from the intervening period that need Backlog capture.
- **Issues:** None — no uncaptured routine findings.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any item is approved for immediate implementation.
- **Result:** No items are explicitly approved by the user for promotion. The two resolved P0s are now clean. The P3 Phase 2 candidates (self-update, micro-task fast path) are flagged but require user decision before promotion. No autonomous action taken on this step.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `Logs/RoutineLog.md`.
- **Result:** Entry written.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Backlog Hygiene row.
- **Result:** `Last Ran` → 2026-09-04, `Next Due` → 2026-09-11, `Status` → `done`.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | Both P0 items resolved in v0.4.3 + v0.4.2 — stale entries never cleaned up | `Backlog.md` P0 section | Commented out with resolution notes (autonomous) |
| 2 | high | HOMEBREW_TAP_TOKEN PAT reminder date (2026-07-15) has passed by 51 days — PAT likely expired | `Backlog.md` P1 ops row | Flagged for user — rotate PAT immediately before next release |
| 3 | medium | P1 `[feature] Full agent-drivable CLI parity` may be resolved by Plan 41 (JSONL/exit contract shipped for all 4 cmds) | `Backlog.md` P1 | Flagged for user — confirm scope and close or narrow |
| 4 | medium | Two P3 items (`Self-update mechanism`, `Micro-task fast path`) are direct Phase 2 milestones — Phase 2 now active | `Backlog.md` P3 | Flagged for user — consider promoting to P2 |
| 5 | medium | `[ops] Routine bot PR pile-up` — 4 months stale at P1, decision still pending | `Backlog.md` P1 | Flagged for user — make one of the 3 proposed decisions |
| 6 | low | `[debt] Stale agent worktrees + branches` — 4.5 months at P1, no action taken | `Backlog.md` P1 | Flagged for user — schedule sweep or demote |
| 7 | low | Group B Code Quality & Testing (8 items) at P1 since 2026-04-16 — 4.5 months stale | `Backlog.md` Group B | Flagged for user — re-prioritize as a group or schedule dedicated sprint |
| 8 | info | No routine executions between 2026-05-07 and 2026-09-04 — all other routines also overdue | `Logs/RoutineLog.md` | Noted — other routines need scheduling |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **URGENT: Rotate HOMEBREW_TAP_TOKEN PAT** — reminder date 2026-07-15 has passed. If the PAT expired, the next release will fail at the Homebrew formula step with `401 Bad credentials`. Check GitHub secrets on `LastStep/Bonsai` and rotate if expired (go to GitHub → Settings → Developer settings → Fine-grained tokens).

- **Confirm P1 scope: `[feature] Full agent-drivable non-interactive CLI parity`** — Plan 41 shipped JSONL/exit contract for all four commands. Verify whether this item is now fully resolved and should be removed, or whether there's remaining scope (e.g., `--from-config` for update/remove, unified `--non-interactive` flag parity) that still needs tracking.

- **Promote P3 → P2: `Self-update mechanism` and `Micro-task fast path`** — Both are listed Phase 2 milestones, and Phase 1 is fully complete. Recommend promoting both from P3 to P2 to reflect current project phase.

- **Decide P1: `[ops] Routine bot PR pile-up`** — Choose one of the three documented options: (a) commit direct to main, (b) auto-merge if green, (c) skip PR creation when local digest covers the date range.

- **Re-prioritize Group B Code Quality & Testing** — 8 P1 items at this priority for 4.5 months suggests they are not truly P1. Consider scheduling a dedicated testing sprint or demoting to P2 with a target phase.

- **All other routines are overdue** — Last run date across all 7 routines is 2026-05-04 or 2026-05-07. With a 7-day frequency, all are now ~119 days overdue. Recommend running the full routine cycle.

## Notes for Next Run

- Both P0 items are now cleaned up; the P0 section is empty. If a new genuine P0 surfaces, it should appear here and be reflected in Status.md within the same session.
- The HOMEBREW PAT situation should be resolved before any release — verify PAT rotation happened by checking for a successful Homebrew formula update after next release.
- Group B re-prioritization decision should be recorded in `Logs/KeyDecisionLog.md` when made.
- Plan 41 scope resolution for the "Full agent-drivable CLI parity" item should update or close this P1 entry.
