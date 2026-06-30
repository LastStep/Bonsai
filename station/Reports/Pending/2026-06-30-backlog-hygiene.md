---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-06-30
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
- **Tools Used:** Read, Edit, Write, Bash
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section; identified 2 P0 items and cross-checked each against `Status.md`.
- **Result:** Both P0 items were already resolved and present in Status.md Recently Done:
  - `[bug] Sensor hook commands use $PWD-walk-up` — resolved 2026-05-13 via v0.4.3 hotfix (PRs #105/#106).
  - `[feature] bonsai init / bonsai add need non-interactive flags` — resolved 2026-05-13 via v0.4.2 (PR #102).
  - Neither required escalation; both required removal (handled in Step 2).
- **Issues:** None (no genuinely unescalated P0s found).

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md` In Progress, Pending, and Recently Done tables; cross-referenced against all Backlog P0/P1 items.
- **Result:**
  - Removed 2 resolved P0 items from Backlog (see Findings). Replaced with HTML comments for audit trail.
  - `[research] Trial sentrux` was already correctly commented out in the P0 section and present in Status.md Pending (blocked on Rust toolchain).
  - No other Backlog items matched "In Progress" work in Status.md (In Progress table is currently empty).
  - No Status.md Pending "Blocked By" items found that could be unblocked by a Backlog item — the only Pending item (sentrux) is blocked on an environment dependency (Rust toolchain), not a Backlog item.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md`; compared Phase 2/3/4 open items against Backlog P2/P3.
- **Result:**
  - Phase 1 is fully complete (all boxes checked).
  - Phase 2 open items:
    - `[ ] Self-update mechanism` → present in Backlog P3 (Big Bets). Priority may warrant promotion to P2 now that Phase 1 is complete.
    - `[ ] Template variables expansion` → NOT present in Backlog at any priority. Potential gap — flagged for user review.
    - `[ ] Micro-task fast path` → present in Backlog P3 (Future Platform).
  - Phase 2 alignment: the top Backlog P1 item (`[feature] Full agent-drivable (non-interactive) CLI parity`) aligns directly with Phase 3 (Cloud & Orchestration) goals and the Odysseus platform. No promotion action needed — already at P1.
  - No Backlog items reference deprecated approaches or completed-phase work (post-removal of the two resolved P0s).
- **Issues:** One gap: `Template variables expansion` (Phase 2 milestone) has no Backlog entry.

### Step 4: Flag stale items
- **Action:** Scanned all priority tiers for items at the same priority 30+ days without progress, items with unclear rationale, and near-duplicates.
- **Result:**
  - **PAT expiry alert (time-sensitive):** P1 item `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` — added 2026-04-22. The PAT was rotated 2026-04-22 with a 90-day expiry; reminder target is ~2026-07-15. Today is 2026-06-30 — the PAT expires in approximately 15 days. **Flagged for immediate user attention.**
  - **Long-stale P1 items (no movement since April):**
    - `[debt] Testing infrastructure for triggers and sensors` — added 2026-04-16 (75 days). No progress noted in any session log.
    - `[debt] Stale agent worktrees + branches accumulating` — added 2026-04-20/21 (70 days). This was flagged repeatedly but never actioned as a sweep.
    - `[ops] Routine bot PR pile-up` — added 2026-05-07 (54 days). No resolution noted.
  - **Group A stale:** `[bookkeeping] Retroactively trim Backlog entries` — added 2026-04-25 (66 days). No progress; each new entry continues the verbose format this item would fix.
  - **Group B:** Most items added 2026-04-16 (75 days). No clear champion session identified.
  - **Group F:** `[docs] Document AltScreen behavior change` — added 2026-04-20 (71 days). Mildly time-sensitive (should accompany a future release).
  - **Near-duplicates found:**
    - `[feature] Full agent-drivable (non-interactive) CLI parity` (P1) supersedes the now-removed `[feature] bonsai init / bonsai add need non-interactive flags` (was P0). The P1 item explicitly notes it supersedes Plan 40 Phase 4. No duplicate remains.
    - No other near-duplicates detected across priority tiers.
- **Issues:** 1 time-sensitive flag (PAT expiry ~15 days), 3 long-stale P1 items worth re-prioritizing.

### Step 5: Check for routine-generated items
- **Action:** Read `RoutineLog.md` entries since 2026-05-07 (last backlog-hygiene run).
- **Result:**
  - Only one RoutineLog entry since 2026-05-07: the 2026-06-13 Plan 40 dispatch session.
  - That session explicitly filed multiple new Backlog items (P2 symlink hardening, validate drift, Plan 40 review nits, bonsai validate lockfile gitignored issue, website npm vuln, unify remove cinematic/headless logic).
  - All items from that session are already present in Backlog P2. No uncaptured findings.
  - No routine subagent runs (dependency audit, vulnerability scan, doc freshness, etc.) have run since 2026-05-07 — all those routines are significantly overdue.
- **Issues:** None in terms of uncaptured findings. The overdue routines are not a backlog item but a session-start flag for the user.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any items are approved for immediate implementation or require routing.
- **Result:** No user authorization for promotion exists in this autonomous run. No item self-evidently meets the "P0 needs immediate action" threshold after the P0 removals. The PAT expiry (Step 4) is flagged to user but is a manual action (rotate secret), not an issue-to-implementation candidate.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Entry appended.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Backlog Hygiene row.
- **Result:** `Last Ran` → 2026-06-30, `Next Due` → 2026-07-07, `Status` → done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | P0 item `[bug] Sensor hook commands` resolved in v0.4.3 but still in Backlog | `Backlog.md` P0 | Removed; replaced with HTML comment audit trail |
| 2 | High | P0 item `[feature] non-interactive flags` resolved in v0.4.2 but still in Backlog | `Backlog.md` P0 | Removed; replaced with HTML comment audit trail |
| 3 | High (time-sensitive) | HOMEBREW_TAP_TOKEN PAT expires ~2026-07-15 (in ~15 days) | `Backlog.md` P1 | Flagged for user — manual secret rotation required |
| 4 | Medium | `Template variables expansion` (Roadmap Phase 2 milestone) has no Backlog entry | `Roadmap.md` Phase 2 | Flagged for user — consider adding P2/P3 entry |
| 5 | Low | `[debt] Testing infrastructure for triggers and sensors` stale 75 days at P1 | `Backlog.md` P1 | Flagged for user — re-prioritize or defer to P2 |
| 6 | Low | `[debt] Stale agent worktrees + branches accumulating` stale 70 days at P1 | `Backlog.md` P1 | Flagged for user — run one-time sweep or lower priority |
| 7 | Low | `[ops] Routine bot PR pile-up` stale 54 days at P1 | `Backlog.md` P1 | Flagged for user — resolve or accept as won't-fix |
| 8 | Low | All other routines overdue (Dependency Audit, Vulnerability Scan, Doc Freshness, Memory Consolidation, Status Hygiene, Roadmap Accuracy) | `routines.md` dashboard | Not a Backlog item — session-start concern for user |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **PAT expiry (urgent ~15 days):** `HOMEBREW_TAP_TOKEN` was rotated 2026-04-22 (90-day default expiry). Rotate the secret at `LastStep/Bonsai` repo settings before 2026-07-15 to avoid GoReleaser brew step failure on next release.

2. **Roadmap gap:** `Template variables expansion` is a Phase 2 milestone with no corresponding Backlog entry. Decide: add a P2 item, fold into another item, or accept as undocumented.

3. **Stale P1s for re-triage:** Three P1 items have had no movement in 54–75 days:
   - `[debt] Testing infrastructure for triggers and sensors` (2026-04-16)
   - `[debt] Stale agent worktrees + branches accumulating` (2026-04-20)
   - `[ops] Routine bot PR pile-up` (2026-05-07)
   Consider re-prioritizing to P2 or scheduling dedicated sessions.

4. **Overdue routines:** All other routines last ran in early May 2026 (~7–8 weeks ago). Session-start hook should be flagging these as overdue. Recommend running a routine-digest session soon.

## Notes for Next Run

- P0 section is now empty of actionable items (two resolved items replaced with HTML comments). If P0 fills again, the next run should catch it.
- The `Template variables expansion` Roadmap gap should be verified — if a backlog item was added by the user between now and the next run, no action needed.
- PAT rotation should occur before next run (due ~2026-07-15, next backlog-hygiene due 2026-07-07).
- All other routines are significantly overdue; if routine-digest hasn't run by the next backlog-hygiene cycle, note that as a systemic flag.
