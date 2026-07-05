---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-05
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
- **Files Read:** 5
  - `station/agent/Routines/backlog-hygiene.md`
  - `station/Playbook/Backlog.md`
  - `station/Playbook/Status.md`
  - `station/Playbook/Roadmap.md`
  - `station/Logs/RoutineLog.md`
- **Files Modified:** 3
  - `station/Playbook/Backlog.md` — 3 resolved items commented out
  - `station/agent/Core/routines.md` — dashboard row updated
  - `station/Logs/RoutineLog.md` — entry appended
- **Tools Used:** Read, Edit, Write, Grep, Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s

Two P0 items were found in the Backlog:

**P0 item A:** `[bug] Sensor hook commands use $PWD-walk-up`
- This was filed 2026-05-13 with the note "Ships v0.4.3."
- Status.md "Recently Done" shows: "v0.4.3 hotfix shipped — sensor hook commands now bake install-time absolute paths in `.claude/settings.json`" (2026-05-13).
- **Finding: RESOLVED.** The bug that prompted this P0 was the fix itself. Commented out.

**P0 item B:** `[feature] bonsai init / bonsai add need non-interactive flags`
- Filed 2026-05-08, was a blocker for Plan 38 P2 rung-3.
- Status.md "Recently Done" shows: "v0.4.2 release shipped — `bonsai init`/`add` `--non-interactive --from-config <path>`" (2026-05-13).
- **Finding: RESOLVED.** Flags shipped. Remaining headless parity work was tracked in P1 (see Step 2). Commented out.

After cleanup: P0 section is empty (only the existing promoted/commented sentrux research item).

### Step 2 — Cross-reference with Status.md

Status.md "In Progress": none.
Status.md "Pending": only `[research] Trial sentrux` — already commented out of Backlog P0 with a promotion note. Correct.

Status.md "Recently Done" key entries checked against Backlog P1:

**P1 item:** `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` (added 2026-06-13)
- Status.md (2026-06-16): "Plan 41 — Headless CLI Contract + MCP-ready cores SHIPPED — all 5 phases merged... Every mutating cmd (init/add/update/remove) has a pure `*Result` headless core + JSONL/exit contract (ExitConflict=5)."
- **Finding: RESOLVED.** The P1 goal was exactly what Plan 41 delivered. Commented out.

**Blocked-by check:** Status.md Pending "sentrux trial" is blocked by missing Rust toolchain (cargo/rustc). No Backlog item exists that would unblock this — it requires a user action (rustup install), not a code change. No action needed.

### Step 3 — Cross-reference with Roadmap.md

Roadmap Phase 1: All items `[x]` — complete. No stale checkboxes found (2026-05-07 routine-digest applied corrections per RoutineLog).

Roadmap Phase 2 alignment with P2/P3 Backlog:
- "Self-update mechanism" → P3 backlog item `[improvement] Self-update mechanism` — correctly at P3, no promotion warranted.
- "Template variables expansion" → no explicit backlog item, but subsumed by custom item creator (P3). Low priority.
- "Micro-task fast path" → P3 backlog item `[improvement] Micro-task fast path` — correctly at P3.

No items reference deprecated approaches or completed phases (Phase 1 items are all checked off and not reflected in active P0/P1 backlog after cleanup).

**No Roadmap-triggered promotions needed.** Phase 2 work remains in research/P3 territory.

### Step 4 — Flag stale items

Today: 2026-07-05. Last hygiene run: 2026-05-07 (~59 days gap).

Items at the same priority for 30+ days without progress:

| Item | Priority | Added | Age | Assessment |
|------|----------|-------|-----|------------|
| `[debt] Testing infrastructure for triggers and sensors` | P1 | 2026-04-16 | 80 days | Stale — no progress logged |
| `[debt] Stale agent worktrees + branches accumulating` | P1 | 2026-04-20 | 76 days | Stale — no cleanup logged |
| `[ops] Routine bot PR pile-up` | P1 | 2026-05-07 | 59 days | Stale — no fix implemented |
| **`[ops] HOMEBREW_TAP_TOKEN PAT expiry`** | **P1** | **2026-04-22** | **~10 days until due!** | **URGENT — calendar reminder was for ~2026-07-15** |
| `[bookkeeping] Retroactively trim Backlog entries to NoteStandards` | Group A | 2026-04-25 | 71 days | Stale — no progress |
| `[Plan-29-test-gap] Inverse-chrome companion test` | Group B | 2026-04-23 | 73 days | Low-risk stale test gap |
| `[improvement] Consolidate FieldNotes usage` | Group E | 2026-04-15 | 81 days | Stale, unclear owner |
| `[research] Revisit concept-decisions research` | Group D | 2026-04-16 | 80 days | Stale research item |

**URGENT flag:** `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — The PAT was rotated 2026-04-22 with a 90-day default expiry. Calendar reminder was set for ~2026-07-15. Today is 2026-07-05 — **the PAT is ~10 days from expiry.** If not rotated before the next release, GoReleaser will fail at the brew step with 401. User must rotate `HOMEBREW_TAP_TOKEN` on `LastStep/Bonsai` repo secrets ASAP.

No items found with no clear context or rationale.

**Near-duplicate check:** No new near-duplicates found. Previously flagged changelog near-duplicate (2026-04-21 run) resolved — only one changelog entry remains in Group D.

### Step 5 — Check for routine-generated items since last run

RoutineLog entries after 2026-05-07: only `2026-06-13 — Plan 40 dispatch` (a plan execution, not a routine). No routine has run between 2026-05-07 and today.

**Observation:** The following routines are all significantly overdue (last ran 2026-05-04 to 2026-05-07, now 2026-07-05):
- Dependency Audit: 62 days overdue (was due 2026-05-11)
- Vulnerability Scan: 62 days overdue (was due 2026-05-11)
- Doc Freshness Check: 62 days overdue (was due 2026-05-11)
- Memory Consolidation: 59 days overdue (was due 2026-05-12)
- Status Hygiene: 59 days overdue (was due 2026-05-12)
- Roadmap Accuracy: 45 days overdue (was due 2026-05-21)
- Backlog Hygiene: 59 days overdue (this run closes the gap)

Three significant code-shipping events occurred since last routine cycle: v0.4.3 hotfix (2026-05-13), Plan 40 Phases 1-3 (2026-06-13), Plan 41 headless contract (2026-06-16). All routines should run soon to catch any drift, new vulns, or stale doc references from these plans.

No uncaptured routine-generated backlog items found (no routine ran to generate findings).

### Step 6 — Promote ready items via issue-to-implementation

No items are flagged for immediate implementation. The cleared P0s were resolved by already-shipped plans. The remaining P0 section is now empty — no urgent escalation needed except the PAT token (ops task, not a code change).

### Steps 7 and 8 — Log results and update dashboard

Completed below.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | P0 bug (sensor walk-up) resolved by v0.4.3 — still in active Backlog P0 | Backlog.md P0 | Commented out with resolution note |
| 2 | Medium | P0 feature (non-interactive flags) resolved by v0.4.2 — still in active Backlog P0 | Backlog.md P0 | Commented out with resolution note |
| 3 | Medium | P1 "Full agent-drivable CLI parity" resolved by Plan 41 — still in active Backlog P1 | Backlog.md P1 | Commented out with resolution note |
| 4 | **HIGH** | HOMEBREW_TAP_TOKEN PAT expires ~2026-07-15 — 10 days away | Backlog.md P1 | Flagged for user — requires PAT rotation in repo secrets |
| 5 | Low | Testing infra for triggers/sensors — 80 days stale at P1 with no progress | Backlog.md P1 Group B | Flagged for re-prioritization |
| 6 | Low | Stale worktrees/branches cleanup — 76 days stale at P1 | Backlog.md P1 | Flagged for re-prioritization |
| 7 | Low | Routine bot PR pile-up fix — 59 days stale at P1 | Backlog.md P1 | Flagged for re-prioritization |
| 8 | Medium | All 6 other routines are 45–62 days overdue | RoutineLog / Dashboard | Flagged — recommend running all routines |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **URGENT (10 days) — Rotate HOMEBREW_TAP_TOKEN PAT.** The fine-grained PAT rotated 2026-04-22 has a ~90-day default expiry. Calendar reminder was ~2026-07-15. Today is 2026-07-05. Go to `LastStep/Bonsai` → Settings → Secrets and rotate `HOMEBREW_TAP_TOKEN` before the next release. Symptom of expiry: GoReleaser fails at brew step with "401 Bad credentials."

2. **Re-prioritize or defer stale P1s.** Three P1 items have been at P1 for 59–80 days with no logged progress: testing infrastructure (Group B), worktree cleanup (housekeeping), and routine bot PR pile-up. If these are not being actively worked, consider demoting to P2 or creating a dedicated cleanup session.

3. **All other routines are significantly overdue.** No routine has run in 59+ days. Plans 40 and 41 shipped substantial changes in this window — Dependency Audit, Vulnerability Scan, Doc Freshness Check, Status Hygiene, Memory Consolidation, and Roadmap Accuracy should all run at the next opportunity.

## Notes for Next Run

- P0 section is now empty (audit-only): both P0s resolved and commented, sentrux promoted to Status.md Pending. If any new P0s are added, check promptly.
- The gap between 2026-05-07 and 2026-07-05 (59 days) is unusually long. Plans 40 and 41 likely generated doc drift and new debt that other routines will surface.
- HOMEBREW_TAP_TOKEN PAT: if next run is after 2026-07-15, check whether it was rotated. If not, GoReleaser brew step will fail on next tag.
