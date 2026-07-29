---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-29
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
- **Files Read:** 5
  - `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`
  - `/home/user/Bonsai/station/Playbook/Backlog.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 3
  - `/home/user/Bonsai/station/Playbook/Backlog.md` (3 items removed)
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard updated)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (entry appended)
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section. Found 2 active items (plus 1 already-commented sentrux entry).
- **Result:** Both P0 items are RESOLVED by recent releases — neither should be in P0. Both removed (see Step 2).
  - `[bug] Sensor hook commands use $PWD-walk-up` — fixed in v0.4.3 (PR #105/#106, 2026-05-13)
  - `[feature] bonsai init / bonsai add need non-interactive flags` — shipped in v0.4.2 (PR #102, 2026-05-13)
- **Issues:** P0 section was stale. Both items were resolved after the 2026-05-07 hygiene run but never cleared.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress and Recently Done. Cross-referenced each Backlog item.
- **Result:** 3 items resolved and removed from Backlog:
  1. P0 `$PWD-walk-up bug` — matches Status.md "v0.4.3 hotfix shipped" (2026-05-13)
  2. P0 `--non-interactive flags feature` — matches Status.md "v0.4.2 release shipped" (2026-05-13)
  3. P1 `Full agent-drivable CLI parity` — matches Status.md "Plan 41 — Headless CLI Contract SHIPPED" (2026-06-16)
- **Issues:** None. Blocked-by check: Pending item "Trial sentrux" is blocked on Rust toolchain install — no Backlog items that could unblock it (it requires user action to install rustup).

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md phases and compared against P2/P3 Backlog items.
- **Result:** Phase 1 is fully complete (all boxes checked). Phase 2 milestones (Extensibility) align with existing Backlog items appropriately:
  - P3 `Self-update mechanism` aligns with Phase 2 — correct priority, no promotion warranted
  - P3 `Micro-task fast path` aligns with Phase 2 — correct priority, no promotion warranted
  - P3 `Custom item creator` aligns with Phase 2 — correct priority, no promotion warranted
- **Issues:** None found. No Backlog items reference deprecated approaches or completed phases.

### Step 4: Flag stale items
- **Action:** Checked item ages and clarity of rationale.
- **Result:** Several stale items identified:
  1. **URGENT** — `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` (P1, added 2026-04-22) — reminder was set for ~2026-07-15. Today is 2026-07-29, **14 days past the deadline**. PAT is almost certainly expired. Next release will fail at the Homebrew step unless rotated.
  2. `[ops] Routine bot PR pile-up` (P1, added 2026-05-07) — underlying architectural fix never implemented. The 9 stale PRs were closed, but the root cause (cloud routine creating PRs that local digest already absorbs) remains. Ongoing risk.
  3. `[bookkeeping] Retroactively trim Backlog entries to NoteStandards` (Group A P2, added 2026-04-25) — 95 days old with no progress noted. Most entries still exceed the 3-line cap. Low urgency but worth acknowledging.
  4. Group B items (`[Plan-31-cosmetic]`, `[Plan-29-test-gap]`, `[Plan-29-security-hardening]` remaining sub-item) — all 95+ days old. The Unicode lookalike item in Plan-29-security-hardening is explicitly marked "purely speculative" with no real attack vector — candidate for removal on next user review.
- **Issues:** HOMEBREW_TAP_TOKEN requires immediate user action.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since 2026-05-07 (the date of the last backlog-hygiene run).
- **Result:** Two plan dispatch entries logged after 2026-05-07 (Plan 40 on 2026-06-13, Plan 41 on 2026-06-16) each filed Backlog items at the time. Verified all are captured:
  - Plan 40 → P2 items for symlink hardening, validate drift warning, review nits, validate/lock bug — all present in Backlog
  - Plan 41 → P2 items for website npm vulns, remove business-logic duplication — all present in Backlog
- **Notable gap:** No routine execution logs exist between 2026-05-07 and 2026-07-29 (83-day gap). All 7 routines are severely overdue per dashboard. No routine findings were generated during this period, so there are no uncaptured items from that angle — but the gap itself is worth flagging.
- **Issues:** 83-day routine execution gap. Multiple high-priority routines overdue (Dependency Audit, Vulnerability Scan, Doc Freshness, Status Hygiene, Memory Consolidation).

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any item warrants immediate promotion via issue-to-implementation workflow.
- **Result:** No items were approved for implementation in this pass. The P2 `[security] Website npm vuln tree` item is the highest-urgency candidate (6 open Dependabot alerts, HIGH severity), but confirmation is needed before triggering the full workflow.
- **Issues:** Flagging website npm vulns for user review — needs a decision on whether to route through vulnerability-scan routine or promote directly.

### Step 7 & 8: Log and dashboard update
- **Action:** Appended entry to RoutineLog.md and updated routines.md dashboard.
- **Result:** Both completed successfully.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Resolved | P0 `$PWD-walk-up` bug cleared by v0.4.3 | Backlog P0 | Commented out (audit trail preserved) |
| 2 | Resolved | P0 `--non-interactive flags` cleared by v0.4.2 | Backlog P0 | Commented out (audit trail preserved) |
| 3 | Resolved | P1 `Full agent-drivable CLI parity` cleared by Plan 41 | Backlog P1 | Commented out (audit trail preserved) |
| 4 | HIGH | HOMEBREW_TAP_TOKEN rotation 14 days past due (was 2026-07-15) | Backlog P1 | Flagged — user action required |
| 5 | MEDIUM | 83-day routine execution gap — all 7 routines severely overdue | routines.md dashboard | Flagged — user decision needed |
| 6 | MEDIUM | Website npm vulns (6 alerts, incl. HIGH esbuild+vite) unresolved since 2026-06-16 | Backlog P2 | Flagged — candidate for vulnerability-scan routine or direct plan |
| 7 | LOW | `[ops] Routine bot PR pile-up` architectural fix never implemented | Backlog P1 | Flagged — user decision on approach (a/b/c) |
| 8 | LOW | P0 section now fully empty — no critical blockers | Backlog P0 | Informational only |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

### 1. HIGH — HOMEBREW_TAP_TOKEN PAT rotation overdue
The reminder was set for ~2026-07-15 (90 days after 2026-04-22 rotation). Today is 2026-07-29, 14 days past the deadline. Fine-grained PATs default to 90-day expiry. If not rotated, the next release will ship binaries successfully but GoReleaser's Homebrew formula update will fail with `401 Bad credentials`.

**Action needed:** Rotate `HOMEBREW_TAP_TOKEN` on GitHub, update `LastStep/Bonsai` secret, update the Backlog entry with new rotation date.

### 2. MEDIUM — All routines severely overdue (83-day gap)
The dashboard shows all routines last ran on 2026-05-04 or 2026-05-07. None have run since. Overdue by 47–78 days depending on routine. Suggest a routine-digest session to process multiple routines at once:
- Dependency Audit (overdue ~78 days)
- Vulnerability Scan (overdue ~78 days) — especially relevant given open website npm vuln alerts
- Doc Freshness Check (overdue ~78 days)
- Memory Consolidation (overdue ~78 days)
- Status Hygiene (overdue ~78 days)
- Roadmap Accuracy (overdue ~69 days)
- Backlog Hygiene (completed today)

### 3. MEDIUM — Website npm vulnerabilities (6 open Dependabot alerts)
The P2 Backlog item from 2026-06-16 notes 6 open Dependabot alerts including HIGH severity esbuild and vite. The astro bump PR (#108) that would clear most alerts fails the website build after rebase. This has been unresolved for 43 days. Route through vulnerability-scan routine or promote directly to a plan.

### 4. LOW — Routine bot PR pile-up architectural fix deferred
9 PRs were closed in 2026-05-07, but the cloud routine still pushes daily PRs. No fix was chosen from the three options (a: commit-direct, b: auto-merge, c: skip when local absorbed). This will recur.

---

## Notes for Next Run

- P0 section is now empty — verify this is the correct state at next run
- HOMEBREW_TAP_TOKEN rotation: update the Backlog P1 entry with the new rotation date after rotating
- Group B items (Plan-29-test-gap, Plan-29-security-hardening remaining sub-item, Plan-31-cosmetic) are all 95+ days old — consider whether these small items warrant a P3 demotion or a single bundled fix-up plan
- If `[ops] Routine bot PR pile-up` fix is chosen (option a/b/c), remove or resolve that P1 entry
- The 83-day routine gap is structural — consider whether the loop.md dispatch is functioning correctly
