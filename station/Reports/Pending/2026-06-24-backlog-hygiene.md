---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-06-24
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07 (previous value from dashboard)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 minutes
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Tools Used:** Read, Edit, Write, Grep
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; identified 2 P0 items. Cross-referenced each against Status.md.
- **Result:** Both P0 items are resolved by shipped code:
  - `[bug] Sensor hook commands use $PWD-walk-up` — fixed in v0.4.3 (PRs #105/#106, 2026-05-13). Hook commands now bake install-time absolute paths.
  - `[feature] bonsai init / bonsai add need non-interactive flags` — shipped in v0.4.2 (Plan 39, PR #102, 2026-05-13). Both `--non-interactive` and `--from-config` flags delivered.
  - Neither was in Status.md as Pending or In Progress (both are now in Recently Done via their release entries).
- **Action Taken:** Commented out both P0 items with resolution notes; P0 section is now empty of active items.
- **Issues:** None — clean resolution.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md. Compared Backlog items against In Progress, Pending, and Recently Done.
- **Result:**
  - Status.md Pending: only "Trial sentrux" (already commented out in Backlog P0 — correctly tracked).
  - Status.md In Progress: none.
  - Plan 41 (Headless CLI, shipped 2026-06-16) substantially delivered on the P1 "Full agent-drivable CLI parity" item — all four cmds now have headless cores + JSONL/exit contract. The remaining piece is the MCP server (fast-follow Plan 42).
  - No Backlog items were found duplicated in Status.md Recently Done that still appear as active entries (aside from the two P0s now commented out).
  - No Status.md Pending items appeared blockable by Backlog resolution.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md. Checked Backlog P2/P3 items against current phase milestones.
- **Result:**
  - Phase 1 (Foundation & Polish) is complete — all boxes checked. No P0/P1 items reference deprecated Phase 1 approaches.
  - Phase 2 (Extensibility) milestones: "Self-update mechanism" and "Micro-task fast path" are in P3 Backlog correctly. "Template variables expansion" is not explicitly tracked in Backlog (minor gap — not flagged as P2 in roadmap either).
  - `[feature] Integrate plan-grilling as first-class Bonsai catalog ability` (P2) aligns with Phase 2 Extensibility goals. Could warrant promotion — flagged for user review (low urgency).
  - No items reference deprecated approaches or completed phases that are still filed as active P0/P1 work.
- **Issues:** None blocking.

### Step 4: Flag stale items
- **Action:** Checked item ages. Last run was 2026-05-07; today is 2026-06-24 (48 days gap). Checked items added before 2026-05-25 (30+ days at same priority).
- **Result:** Several P1 items are 60–70+ days old with no movement:
  - `[ops] HOMEBREW_TAP_TOKEN PAT expiry` (added 2026-04-22, 63 days) — **CRITICAL TIME SENSITIVITY**: PAT expires ~2026-07-15, only 21 days away. Needs immediate user action.
  - `[ops] Routine bot PR pile-up` (added 2026-05-07, 48 days) — stale, no fix applied.
  - `[debt] Testing infrastructure for triggers and sensors` (added 2026-04-16, 69 days) — no movement.
  - `[debt] Stale agent worktrees + branches` (added 2026-04-20, 65 days) — no movement.
  - Group B debt items (all 2026-04-16 to 2026-04-24, 60–69 days) — no movement; lower priority debt, acceptable.
  - No clear near-duplicates identified beyond the Plan 41 overlap noted in Step 2.
- **Action Taken:** Added `[URGENT — expires ~2026-07-15, 21 days away]` tag to HOMEBREW_TAP_TOKEN P1 item and updated its note with rotation deadline. Updated P1 "Full agent-drivable CLI parity" entry to reflect Plan 41 delivery and identify MCP server as remaining piece.
- **Issues:** HOMEBREW_TAP_TOKEN expiry is the primary time-sensitive finding.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07). Checked for uncaptured findings.
- **Result:** No routine runs logged between 2026-05-07 and 2026-06-24 (48-day gap with no routine execution). Plan 40 (2026-06-13) and Plan 41 (2026-06-16) execution notes in RoutineLog generated several new Backlog items — all confirmed captured:
  - P2: symlink hardening, validate drift warning, Plan 40 review nits, lockfile gitignore issue — all in Backlog P2.
  - P2: debt unify remove logic, website npm vuln tree — all in Backlog P2.
  - No uncaptured findings requiring new Backlog items.
- **Issues:** Notable gap — no routines have run since 2026-05-04 to 2026-05-07. All 7 routines are significantly overdue (dashboard shows next due dates in May). This routine's own previous run was 2026-05-07 (48 days ago vs 7-day frequency). Flagged for user awareness.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed items for promotion candidates.
- **Result:** No items are user-approved for immediate implementation dispatch. The HOMEBREW_TAP_TOKEN rotation is a manual user action (not implementable via agent). The P1 "Full agent-drivable CLI parity" item needs Plan 42 decision point. Flagging to user for next session.
- **Issues:** None — no promotions dispatched (correct; user must confirm).

### Step 7: Log results
- **Action:** Appended entry to RoutineLog.md.
- **Result:** Entry written.

### Step 8: Update dashboard
- **Action:** Updated routines.md dashboard row for Backlog Hygiene.
- **Result:** Last Ran → 2026-06-24, Next Due → 2026-07-01, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | Both P0 items resolved by shipped code (v0.4.3 + v0.4.2) but still active in Backlog | `Backlog.md` P0 section | Commented out both items with resolution notes; P0 section now empty |
| 2 | high | HOMEBREW_TAP_TOKEN PAT expires ~2026-07-15 — 21 days away. Missing action will break Homebrew formula on next release | `Backlog.md` P1 | Added URGENT tag + rotation deadline annotation; flagged for user |
| 3 | medium | P1 "Full agent-drivable CLI parity" was substantially delivered by Plan 41 (2026-06-16) but Backlog entry not updated | `Backlog.md` P1 | Updated entry to reflect Plan 41 delivery; identified MCP server as remaining piece |
| 4 | medium | All 7 routines significantly overdue — dashboard shows May 2026 due dates; 48-day gap since last any routine ran | `routines.md` dashboard | Flagged for user — all routines need scheduling |
| 5 | low | P1 "Routine bot PR pile-up" (48 days, no fix applied) — unclear if cloud routine behavior has changed | `Backlog.md` P1 | Flagged for user; no action taken (fix requires user decision on cloud routine config) |
| 6 | low | Several Group B debt items 60–70 days old at P1 with no movement | `Backlog.md` P1 | Flagged for awareness; no action taken (debt items, acceptable aging) |
| 7 | info | `[feature] Integrate plan-grilling` (P2) aligns with Roadmap Phase 2 Extensibility — potential promote candidate | `Backlog.md` P2 | Flagged for user; no action taken |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

1. **[URGENT] Rotate HOMEBREW_TAP_TOKEN PAT by 2026-07-15** — The PAT rotated 2026-04-22 expires in ~21 days. Failure to rotate will break GoReleaser's Homebrew formula update on the next release. Go to GitHub → Settings → Developer Settings → Fine-grained tokens and rotate the `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai`.

2. **All 7 routines are significantly overdue** — The dashboard shows all routines next-due in May 2026; today is 2026-06-24. None appear to have run in 48+ days. Consider scheduling a routine-digest session to catch up on: Dependency Audit, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene, Vulnerability Scan. Some may surface actionable findings (website npm vulns are a known P2 item).

3. **Plan 41 delivered headless CLI parity — confirm MCP server path** — The P1 "Full agent-drivable CLI parity" item was substantially resolved by Plan 41 (shipped 2026-06-16). The remaining piece is the MCP server (noted as "fast-follow Plan 42" in Status.md). Confirm: is Plan 42 ready to dispatch, or should this P1 remain open pending user direction?

4. **Routine bot PR pile-up fix** — The P1 entry for eliminating parallel-track bot PRs (9 closed 2026-05-07) hasn't been actioned. Confirm whether cloud routine behavior has changed, or pick one of the three proposed fixes.

## Notes for Next Run

- P0 section is now empty — good state. Next run should verify no new P0s have accumulated.
- HOMEBREW_TAP_TOKEN rotation deadline: 2026-07-15. If not resolved by next backlog-hygiene run, escalate to P0.
- All other routines are overdue — the next session with user should trigger a full routine-digest pass.
- The `[feature] Integrate plan-grilling` P2 item is a legitimate Phase 2 Extensibility candidate — worth discussing if Phase 2 work is starting.
- Website npm vuln tree (P2) has been open since 2026-06-16 — Vulnerability Scan routine should pick it up when that runs.
