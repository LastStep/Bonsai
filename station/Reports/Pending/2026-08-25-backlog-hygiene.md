---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-25
status: partial
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~8 min
- **Files Read:** 5 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

> **Partial status reason:** Step 6 (promote via issue-to-implementation) requires user confirmation before acting. One high-severity item (HOMEBREW_TAP_TOKEN expiry) is flagged for immediate user attention — remediation requires user action outside this subagent's scope.

---

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog P0 section; checked each active P0 item against Status.md In Progress and Recently Done.
- **Result:** Both active P0 items are already resolved and present in Status.md Recently Done:
  - `[bug] Sensor hook commands use $PWD-walk-up` → v0.4.3 hotfix shipped (Status.md Done 2026-05-13, PR #105/#106).
  - `[feature] bonsai init / bonsai add need non-interactive flags` → v0.4.2 shipped (Status.md Done 2026-05-13, PR #102).
  - Neither required escalation — both were already addressed.
- **Issues:** None requiring escalation. However, both items were stale in the Backlog (not cleaned up after their associated releases shipped 3+ months ago).

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress and Pending sections; matched against Backlog items.
- **Result:**
  - **Removed from Backlog P0:** Both resolved items above (converted to comments preserving audit trail).
  - **In Progress:** Empty — no active tasks. No Backlog cross-references needed.
  - **Pending:** "[research] Trial sentrux on Bonsai repo" — already commented out from Backlog P0 (promoted 2026-05-07). Still blocked on Rust toolchain install. No Backlog item can unblock it.
  - P0 section is now empty after cleanup.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; compared Phase 2/3/4 milestones against Backlog items.
- **Result:**
  - **Phase 1:** Complete — all items checked. Backlog aligns.
  - **Phase 2 alignment:** `Self-update mechanism` (P3 Backlog) and `Micro-task fast path` (P3 Backlog) correctly sit at P3 (Phase 2 but not current priority). `Custom item creator` (P3 Backlog) also aligns with Phase 2 extensibility goal. No promotion warranted without user guidance.
  - **Phase 3 alignment:** `Managed Agents integration` and `Greenhouse companion app` correctly sit in P3 Big Bets.
  - **Deprecated approaches:** No Backlog items reference deprecated approaches or completed phases.
  - **P2/P3 promotion candidates:** The P2 item "[feature] Integrate plan-grilling as a first-class Bonsai catalog ability" aligns with Phase 2 extensibility but the project is not actively in Phase 2, so no promotion.
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Reviewed all items for 30+ day staleness, missing rationale, and near-duplicates.
- **Result:**
  - **HIGH — HOMEBREW_TAP_TOKEN expired (P1, 30+ days overdue):** "[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder" added 2026-04-22. The PAT was rotated 2026-04-22 with a 90-day default expiry → expired approximately 2026-07-21, which is 35 days ago. The next release attempt will fail at the Homebrew tap step with `401 Bad credentials`. This is a release-blocking issue and requires immediate user action.
  - **MEDIUM — P1 CLI parity item stale (75+ days):** "[feature] Full agent-drivable (non-interactive) CLI parity" added 2026-06-13 as "main thing" with a note to "Promote to a plan + grill next session." Item has been at P1 for 73 days without a plan being created. Flagged for user re-prioritization.
  - **LOW — Multiple Group B debt items stale (120+ days):** Group B items (generate.go split, catalog test coverage, cmd/ test coverage) have been at P1/P2 since 2026-04-16 (131 days). No near-term plan scheduled. Appropriate for P1/P2 if Phase 2 work resumes.
  - **Near-duplicates:** None found. The now-removed P0 feature item and the P1 "full CLI parity" item are distinct (the resolved P0 was a subset; the P1 is the broader headless contract).
- **Issues:** PAT expiry is time-sensitive and requires user action.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md for entries since the last backlog-hygiene run (2026-05-07).
- **Result:** No routine executions logged between 2026-05-07 and 2026-08-25 (109-day gap). All routines are severely overdue — dashboard shows all "Next Due" dates in May 2026. No new routine-flagged findings exist to verify against the Backlog.
  - Reviewing the 2026-05-07 Backlog Hygiene report's flags (from RoutineLog):
    1. P0 escalation (Trial sentrux) → Already resolved via comment in Backlog.
    2. Roadmap "Better trigger sections" unchecked → Now shows `[x]` in Roadmap.md (resolved).
    3. `code-index.md` staleness (medium drift) → Not present as a standalone Backlog item. Flagged for user review.
    4. Broken nav link `agent/Skills/bonsai-model.md` → Not present as a standalone Backlog item. Flagged for user review.
    5. INDEX.md arch diagram drift → Not present as a standalone Backlog item. Flagged for user review.
  - These 3 uncaptured findings from 2026-05-04 Doc Freshness Check (and 2026-05-07 Backlog Hygiene) have not been captured as Backlog items — per procedure, not auto-adding, flagging for user review.
- **Issues:** 3 uncaptured doc-drift findings. 109-day routine gap noted.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any Backlog items are ready for immediate promotion.
- **Result:** The P1 "[feature] Full agent-drivable CLI parity" item explicitly says "Promote to a plan + grill next session" and was flagged by the user as "main thing." This is a candidate for promotion — but per procedure, user confirmation is required before routing through issue-to-implementation.
- **Issues:** Not actioned — flagged for user review below.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Backlog Hygiene row.
- **Result:** Done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Critical | HOMEBREW_TAP_TOKEN PAT expired ~2026-07-21 (35 days ago) — next release will fail at Homebrew step | `Backlog.md` P1 | Flagged for immediate user action |
| 2 | High | P0 bug "Sensor hook commands" was resolved in v0.4.3 but remained in Backlog | `Backlog.md` P0 | Removed — converted to comment with resolution note |
| 3 | High | P0 feature "non-interactive flags" was resolved in v0.4.2 but remained in Backlog | `Backlog.md` P0 | Removed — converted to comment with resolution note |
| 4 | Medium | P1 "Full agent-drivable CLI parity" at P1 for 73 days, no plan created despite being "main thing" | `Backlog.md` P1 | Flagged for user re-prioritization / plan creation |
| 5 | Low | 3 uncaptured doc-drift findings from 2026-05-04 Doc Freshness Check: code-index.md staleness, broken bonsai-model.md nav link, INDEX.md arch diagram drift | — | Flagged for user review — not auto-added per procedure |
| 6 | Info | 109-day routine gap — all routines severely overdue since 2026-05-07 | `routines.md` | Noted — dashboard will update this routine; others need separate runs |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

### 1. HOMEBREW_TAP_TOKEN PAT Expiry — CRITICAL (Release Blocker)

The `HOMEBREW_TAP_TOKEN` PAT was rotated 2026-04-22 with a 90-day default expiry. It expired approximately **2026-07-21**, 35 days ago. The next `bonsai` release will fail at the Homebrew formula update step with `401 Bad credentials`. Rotate the PAT in GitHub Actions secrets before the next release attempt.

The P1 Backlog item tracks this — consider resolving it now and removing the item once rotated.

### 2. P1 "Full agent-drivable CLI parity" — Ready for Plan Creation

This item was added 2026-06-13 by the user as "main thing" with the note "Promote to a plan + grill next session." It has been at P1 for 73 days without a plan. Confirm whether to start the issue-to-implementation workflow for this item.

### 3. Uncaptured Doc-Drift Findings (from 2026-05-04 routine cycle)

Three findings from the 2026-05-04 Doc Freshness Check have not been captured as Backlog items:
- `code-index.md` staleness (medium drift — validate command missing, line numbers off, 6 TUI packages undocumented)
- Broken nav link: `agent/Skills/bonsai-model.md`
- INDEX.md architecture diagram drift

These were also re-flagged in the 2026-05-07 Backlog Hygiene report. Decide whether to add Backlog entries for each or handle inline.

---

## Notes for Next Run

- P0 section is now empty — confirm this is correct and there are no emergency issues before the next run.
- All other routines (Dependency Audit, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene, Vulnerability Scan) are 90–110+ days overdue. The next routine-digest session should process all of them.
- The HOMEBREW_TAP_TOKEN rotation is time-sensitive and should be done before P0 is clear.
- If the P1 CLI parity item gets promoted to a plan, the Backlog P1 section will need updating.
