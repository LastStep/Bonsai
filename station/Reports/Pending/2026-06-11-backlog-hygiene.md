---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-06-11
status: partial
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (1 P0 escalation flagged for user; all other checks passed)
- **Duration:** ~8 min
- **Files Read:** 6 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`
- **Files Modified:** 3 — `/home/user/Bonsai/station/Reports/Pending/2026-06-11-backlog-hygiene.md` (this file), `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Write, Edit, Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section; identified all P0 items; cross-referenced each against `Status.md` In Progress and Pending tables.
- **Result:** 1 P0 item exists in the Backlog that is NOT in Status.md:
  - **`[bug] Sensor hook commands use $PWD-walk-up — breaks in multi-.bonsai.yaml setups`** (added 2026-05-13) — This P0 bug was filed after the last backlog-hygiene run. It documents that `internal/generate/generate.go:534` bakes `$PWD`-walk-up into hook commands, causing failures when a Claude Code session cd's into a second Bonsai project. Both repos were hotfixed locally, but `bonsai update` would clobber the fix. The item states "Ships v0.4.3" but no plan or Status.md entry exists. **Requires user decision: promote to Status.md Pending or plan a fix.**
  - The previously-flagged P0 (`[research] Trial sentrux`) remains correctly commented out — it was promoted to Status.md Pending 2026-05-07 and is still there (blocked on Rust toolchain).
- **Issues:** P0 escalation required — see Items Flagged for User Review.

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md`; compared all In Progress and Recently Done entries against Backlog items; checked whether any Pending "Blocked By" items could be unblocked by Backlog work.
- **Result:**
  - In Progress: empty (no active work) — nothing to remove from Backlog.
  - Pending: only `[research] Trial sentrux` (blocked on Rust toolchain) — already commented out in Backlog P0 correctly.
  - Recently Done includes: `v0.4.2 release shipped` (Plan 39, 2026-05-13) — the `bonsai init/add --non-interactive --from-config` feature. This is already correctly commented out in Backlog.md with an HTML comment `<!-- "[feature] bonsai init / bonsai add need non-interactive flags" — shipped in v0.4.2 ... Removed from backlog 2026-06-11 (backlog-hygiene). -->` (comment was pre-existing, correctly dated).
  - No Status.md Pending items with "Blocked By" can be unblocked by a Backlog item. The only Pending item is sentrux (blocked on Rust toolchain, not a Backlog item).
- **Issues:** none.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md`; identified current phase and future milestones; cross-referenced P2/P3 Backlog items against phase goals.
- **Result:**
  - Phase 1 is **fully complete** (all boxes checked including `bonsai validate` added by 2026-05-07 routine-digest). No stale Phase 1 references in Backlog.
  - Phase 2 milestones not yet started: `Self-update mechanism`, `Template variables expansion`, `Micro-task fast path`.
    - Backlog P3 contains `[improvement] Self-update mechanism` (added 2026-04-13) — aligns directly with Phase 2. Candidate for P2 promotion.
    - Backlog P3 contains `[improvement] Micro-task fast path` (added 2026-04-15) — aligns directly with Phase 2. Candidate for P2 promotion.
    - No "Template variables expansion" item exists in Backlog — may want to add one.
  - Phase 3 (Cloud & Orchestration): `[feature] Managed Agents integration` and `[feature] Greenhouse companion app` are in Backlog P3 Big Bets — appropriately long-term, no promotion needed.
  - No Backlog items reference deprecated approaches or completed phases.
- **Issues:** 2 P3 items align with Phase 2 milestones; flagged as promotion candidates for user consideration (not auto-promoted).

### Step 4: Flag stale items
- **Action:** Reviewed all items for age (30+ days without progress), missing context/rationale, and near-duplicates.
- **Result:**
  - **Stale P1 items (30+ days, no progress):**
    - `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` (2026-04-22, 50 days) — concrete action still due ~2026-07-15 (35 days away). **Not stale yet — action window is approaching.** Flag: remind user to rotate before 2026-07-15.
    - `[ops] Routine bot PR pile-up` (2026-05-07, 35 days) — fix needed; no progress noted. Stale at P1.
    - `[debt] Testing infrastructure for triggers and sensors` (2026-04-16, 56 days) — Group B item, no progress. Stale.
    - `[debt] Stale agent worktrees + branches accumulating` (2026-04-20/21, ~51 days) — housekeeping; no progress.
  - **Stale P2 items (30+ days):**
    - `[bookkeeping] Retroactively trim Backlog entries to NoteStandards` (2026-04-25, 47 days) — ironically, the Backlog itself violates the standard it advocates. Still valid.
    - `[improvement] Consolidate FieldNotes usage` (2026-04-15, 57 days) — unclear priority; no progress.
    - `[improvement] Post-update backup merge hint` (2026-04-16, 56 days) — small change, still valid.
    - `[feature] Port statusLine to catalog sensor` (2026-04-22, 50 days) — issue #53 filed; still P2.
  - **Near-duplicates found:** None new; the `[feature] Changelog generation skill` (P2 Group D) and `[improvement] OSS polish — demo GIF` (P2 Group C) remain distinct items.
  - **Items without clear context/rationale:** None — all items carry added dates and source.
- **Issues:** Multiple P1/P2 items are stale (30-57 days). No auto-changes made; flagged for user re-prioritization.

### Step 5: Check for routine-generated items
- **Action:** Read `RoutineLog.md` entries since last backlog-hygiene (2026-05-07); checked for uncaptured findings.
- **Result:** No routine runs have occurred since 2026-05-07. The last entries in RoutineLog are all dated 2026-05-07 (routine-digest, roadmap-accuracy, status-hygiene, backlog-hygiene, memory-consolidation). `Reports/Pending/` is empty — all prior reports archived. No uncaptured findings exist.
- **Issues:** None. All other routines (Dependency Audit, Doc Freshness Check, Vulnerability Scan) are overdue (last ran 2026-05-04, next due 2026-05-11 — now 31 days past due). These routines likely have findings that would generate backlog items. **Flagging that a full routine catch-up is overdue.**

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any item is ready for autonomous promotion without user confirmation.
- **Result:** The P0 bug (`$PWD-walk-up`) has a described fix (bake absolute install-time project root into hook commands) and targets v0.4.3. It is a candidate for the issue-to-implementation workflow but requires user confirmation per procedure. No other items are in a state where auto-promotion is appropriate. All P0/P1 items either need user confirmation or are already tracked.
- **Issues:** P0 promotion pending user decision.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Entry written.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `last_ran` and `Next Due` in `agent/Core/routines.md` dashboard row for Backlog Hygiene.
- **Result:** Updated to Last Ran: 2026-06-11, Next Due: 2026-06-18, Status: done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | **P0 / Critical** | `[bug] Sensor hook commands use $PWD-walk-up` added 2026-05-13 is NOT in Status.md — breaks multi-.bonsai.yaml setups | `Backlog.md` P0 section | Flagged for user — requires promotion to Status.md or plan creation |
| 2 | Medium | HOMEBREW_TAP_TOKEN PAT rotation due ~2026-07-15 (35 days away) | `Backlog.md` P1 | Flagged as time-sensitive reminder — no action taken |
| 3 | Low | P1 `[ops] Routine bot PR pile-up` stale 35 days without progress | `Backlog.md` P1 | Noted — no change |
| 4 | Low | P1 `[debt] Testing infrastructure for triggers and sensors` stale 56 days | `Backlog.md` P1 | Noted — no change |
| 5 | Low | P1 `[debt] Stale agent worktrees + branches accumulating` stale 51 days | `Backlog.md` P1 | Noted — no change |
| 6 | Low | P3 `[improvement] Self-update mechanism` aligns with Phase 2 milestone — candidate for P2 promotion | `Backlog.md` P3 | Flagged — no auto-promotion |
| 7 | Low | P3 `[improvement] Micro-task fast path` aligns with Phase 2 milestone — candidate for P2 promotion | `Backlog.md` P3 | Flagged — no auto-promotion |
| 8 | Info | Dependency Audit, Doc Freshness Check, Vulnerability Scan all ~31 days overdue (last ran 2026-05-04) | `routines.md` dashboard | Flagged — other routines need to run |
| 9 | Info | No new routine-generated findings to capture in Backlog since last run | `RoutineLog.md` | No action needed |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[P0 ESCALATION — URGENT]** `[bug] Sensor hook commands use $PWD-walk-up` (Backlog P0, added 2026-05-13) — This bug has been in P0 for 29 days without a Status.md entry or plan. The item notes "Ships v0.4.3" but no plan exists. Fix is described (bake absolute install-time root into hook command). Recommend: create a Tier 1/2 plan and promote to Status.md, or confirm the local hotfixes are sufficient and downgrade priority.

2. **[TIME-SENSITIVE]** HOMEBREW_TAP_TOKEN PAT rotation due ~2026-07-15 (35 days from today). If the PAT expires unnoticed, the next release's brew formula update will silently fail. Set a calendar reminder now if not already done.

3. **[ROUTINE CATCH-UP OVERDUE]** Dependency Audit, Doc Freshness Check, and Vulnerability Scan last ran 2026-05-04 — all ~31 days overdue. Running these would likely surface new backlog items (Go module updates, doc drift, potential CVEs).

4. **[P2 PROMOTION CANDIDATES]** Two P3 items align directly with Phase 2 Roadmap milestones and may warrant promotion to P2: `Self-update mechanism` and `Micro-task fast path`. Consider promoting when planning Phase 2 work.

## Notes for Next Run

- The P0 `$PWD-walk-up` bug should be resolved (either planned + shipped as v0.4.3, or explicitly downgraded) before next hygiene run.
- PAT rotation (2026-07-15) will have passed by next run (2026-06-18) — confirm rotation happened.
- If other routines run before next backlog-hygiene, their flagged findings should be reviewed for uncaptured backlog items.
- The `[bookkeeping] Retroactively trim Backlog entries to NoteStandards` P2 item remains open — the Backlog itself has many verbose multi-paragraph entries that violate NoteStandards.
