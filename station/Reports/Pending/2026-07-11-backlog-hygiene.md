---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-11
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
- **Duration:** ~8 minutes
- **Files Read:** 6 — `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md` (3 resolved items replaced with tombstone comments), `station/agent/Core/routines.md` (dashboard updated), `station/Logs/RoutineLog.md` (log entry appended)
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; checked each item against Status.md In Progress and Pending.
- **Result:** Found 2 P0 items remaining in Backlog. Both appear in Status.md "Recently Done" (resolved), not missing from Status.md — so no immediate escalation needed. However, both items are fully resolved and should be removed (handled in Step 2).
- **Issues:** None — no P0 is missing from Status.md entirely.

### Step 2: Cross-reference with Status.md
- **Action:** Compared all Backlog items against Status.md In Progress, Pending, and Recently Done.
- **Result:** Identified 3 Backlog items whose corresponding work shipped:
  1. **P0 `[bug] Sensor hook commands use $PWD-walk-up`** → resolved via v0.4.3 hotfix (PRs #105/#106, 2026-05-13, Status.md "Recently Done")
  2. **P0 `[feature] bonsai init / bonsai add need non-interactive flags`** → resolved via v0.4.2 (--non-interactive/--from-config, 2026-05-13) and extended by Plan 41 (all 4 commands, 2026-06-16)
  3. **P1 `[feature] Full agent-drivable CLI parity: init / update / add / remove`** → resolved via Plan 41 shipped 2026-06-16 (PRs #120/#122/#123/#121/#125; every mutating cmd has headless core + JSONL/exit contract)
  - All 3 items replaced with HTML tombstone comments in Backlog.md.
  - Status.md Pending: `[research] Trial sentrux` remains Pending (blocked on Rust toolchain install) — no Backlog item to remove (it was correctly commented out in a prior run).
  - No Status.md "Blocked By" items could be unblocked by a Backlog item (Rust toolchain is a system dependency, not a code item).
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Checked Roadmap.md phases against P2/P3 Backlog items for alignment.
- **Result:**
  - Phase 1 is complete (all checkboxes checked, including the two that were fixed in the 2026-05-07 routine-digest).
  - Phase 2 milestones "Self-update mechanism" and "Micro-task fast path" map exactly to two P3 Backlog items under "Future Platform (Roadmap Phase 2+)". Since Phase 1 is done, these are candidates for promotion from P3 → P2. Flagged for user decision.
  - No Backlog items reference deprecated Phase 1 approaches (the `$PWD-walk-up` bug entry was already removed under Step 2).
- **Issues:** P3→P2 promotion candidates flagged (see Items Flagged for User Review).

### Step 4: Flag stale items
- **Action:** Scanned all Backlog items for entries 30+ days at same priority without progress, and items lacking clear rationale.
- **Result:**
  - **URGENT flag:** P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` — PAT was rotated 2026-04-22, with ~90-day expiry pointing to ~2026-07-15. Today is 2026-07-11, which is **4 days away**. The PAT needs rotation now to avoid GoReleaser failing at the brew step on the next release.
  - **Stale flag:** P1 `[ops] Routine bot PR pile-up — eliminate parallel-track artifacts` — added 2026-05-07, 65 days stale. 9 PRs were closed but the underlying fix (commit-direct vs PR creation for cloud routines) remains unimplemented.
  - **Stale flag:** P1 `[debt] Stale agent worktrees + branches accumulating` — added 2026-04-20, 82 days stale. Recurring housekeeping pattern; worktrees/branches continue to accumulate.
  - All items have clear rationale. No near-duplicates identified in the current backlog (prior near-duplicate between CHANGELOG items was resolved in earlier runs).
- **Issues:** URGENT PAT expiry escalated (see Items Flagged).

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene (2026-05-07). Checked whether any flagged findings are uncaptured in Backlog.
- **Result:** Routine log since 2026-05-07 contains:
  - Plan 40 dispatch (2026-06-13): generated 7 Backlog P2 entries (security hardening, validate warn, review nits, validate dogfood bug, website npm vulns, unify-remove debt, plan-grilling integration) — all appear captured in Backlog P2.
  - Plan 41 dispatch (2026-06-16): debt item "unify remove cinematic/headless logic" → already in Backlog P2.
  - No uncaptured findings from routines since the last run.
- **Issues:** None.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any item is approved for immediate implementation.
- **Result:** No items approved for immediate dispatch. The sentrux research trial is in Status.md Pending (blocked on Rust toolchain). The HOMEBREW_TAP_TOKEN rotation is a user action (not a code dispatch). All other items remain in backlog pending user prioritization.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Backlog Hygiene row.
- **Result:** `Last Ran` → 2026-07-11, `Next Due` → 2026-07-18, `Status` → done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | P0 `[bug] $PWD-walk-up` resolved in v0.4.3 but still in Backlog | Backlog.md P0 | Replaced with HTML tombstone comment |
| 2 | medium | P0 `[feature] non-interactive flags` resolved in v0.4.2 + Plan 41 but still in Backlog | Backlog.md P0 | Replaced with HTML tombstone comment |
| 3 | medium | P1 `[feature] Full agent-drivable CLI parity` resolved in Plan 41 but still in Backlog | Backlog.md P1 | Replaced with HTML tombstone comment |
| 4 | high | P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — due ~2026-07-15, 4 days away | Backlog.md P1 | Flagged URGENT for user action (rotate PAT now) |
| 5 | low | P1 `[ops] Routine bot PR pile-up` — 65 days stale, no resolution | Backlog.md P1 | Flagged for user re-prioritization |
| 6 | low | P1 `[debt] Stale agent worktrees + branches` — 82 days stale | Backlog.md P1 | Flagged for user re-prioritization |
| 7 | info | P3 `Self-update mechanism` + `Micro-task fast path` align with Phase 2 milestones (Phase 1 complete) | Backlog.md P3, Roadmap Phase 2 | Flagged as P2 promotion candidates |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **URGENT — Rotate HOMEBREW_TAP_TOKEN PAT now.** The fine-grained PAT was rotated 2026-04-22 with ~90-day expiry. Target rotation date was ~2026-07-15, which is 4 days away. If not rotated before the next release, GoReleaser will fail at the Homebrew formula update step with `401 Bad credentials`. Action: rotate PAT at GitHub → Settings → Developer settings → Personal access tokens, then run `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`.

2. **Re-prioritize stale P1s.** Two P1 items have sat untouched for 65–82 days:
   - `[ops] Routine bot PR pile-up` (added 2026-05-07): fix needed in cloud routine configuration (commit-direct vs PR creation) — consider demoting to P2 if not actionable near-term.
   - `[debt] Stale agent worktrees + branches` (added 2026-04-20): recurring pattern; if one-time sweep was already done locally, consider closing; if not, should be picked up as housekeeping before next major release.

3. **Phase 2 promotion candidates.** Phase 1 is complete. Two P3 items directly match Phase 2 Roadmap milestones and could be promoted to P2 now that Phase 2 work could begin:
   - `[improvement] Self-update mechanism` (P3 "Future Platform")
   - `[improvement] Micro-task fast path` (P3 "Future Platform")
   User decision: promote to P2, leave as P3, or defer until Phase 2 is formally started.

## Notes for Next Run

- P0 section is now clean (all 3 entries replaced with tombstones). If any new critical issues arise, they should land here with an `*(added YYYY-MM-DD)*` marker.
- The HOMEBREW_TAP_TOKEN rotation should be confirmed done before the next run — if confirmed, the Backlog P1 item can be removed.
- The two stale P1 ops/debt items should be re-triaged by the user; if still unresolved at the next run (2026-07-18), escalate to the user again.
- Plan 40 Phase 4 (update-delivery/bonsai init re-run for existing projects) remains held per user decision; keep an eye on whether any routine flags the existing-project gap.
- The `bonsai validate` dogfood item (P2 `[bug] bonsai validate can't pass on Bonsai repo`) remains open — blocked on lock-file policy decision (.bonsai-lock.yaml gitignored).
