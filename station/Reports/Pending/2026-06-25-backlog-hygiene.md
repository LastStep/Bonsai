---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-06-25
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
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
- **Tools Used:** Read (file reads), Edit (Backlog.md P0/P1 cleanup), Write (this report), Edit (routines.md dashboard)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; cross-referenced each P0 item against Status.md In Progress and Pending tables.
- **Result:** Both P0 items are **fully resolved** — neither belongs in the active P0 section:
  - `[bug] Sensor hook commands use $PWD-walk-up` — shipped as v0.4.3 (PRs #105/#106, 2026-05-13). Hook commands now bake absolute install-time paths.
  - `[feature] bonsai init / bonsai add need non-interactive flags` — shipped as v0.4.2 (PR #102, 2026-05-13). Both flags exist; full headless contract (all 4 commands) subsequently shipped via Plan 41.
- **Action Taken:** Both P0 items commented out with resolution notes. P0 section is now empty of active items — appropriate as no blocking issues are known.
- **Issues:** none

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress + Pending + Recently Done tables; compared against Backlog entries.
- **Result:**
  - **In Progress:** empty — nothing to conflict with.
  - **Pending:** "[research] Trial sentrux on Bonsai repo" — already correctly commented out in Backlog P0 (`promoted to Status.md Pending 2026-05-07`). Clean.
  - **Recently Done:** Plan 41 (Headless CLI Contract) shipped 2026-06-16. This directly resolves the P1 Backlog item "Full agent-drivable (non-interactive) CLI parity: init / update / add / remove" (added 2026-06-13). Item was still active in P1 despite the plan shipping.
- **Action Taken:** P1 "Full CLI parity" item commented out with resolution note pointing to Plan 41.
- **No Status.md Blocked-By items** were found that could be unblocked by resolving a Backlog entry.
- **Issues:** none

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; compared Phase 1/2/3/4 milestones against Backlog P2/P3 items.
- **Result:**
  - Phase 1 is complete — all checkboxes checked. No Backlog P2/P3 items reference Phase 1 milestones that are still open.
  - Phase 2 milestones: "Self-update mechanism" and "Micro-task fast path" have corresponding P3 Backlog entries (appropriate tier).
  - Phase 3 milestones: "Managed Agents integration" and "Greenhouse companion app" have P3 "Big Bets" Backlog entries (appropriate tier).
  - No P2/P3 Backlog items reference deprecated approaches or completed phases that should be flagged.
  - No P2/P3 items appear to warrant promotion to P1 at this time — the primary P1 headroom was just filled by Plan 41 and the PAT rotation.
- **Issues:** none

### Step 4: Flag stale items
- **Action:** Audited all Backlog items for age (30+ days at same priority without progress), missing context, and near-duplicates.
- **Result:**
  - **⚠ HOMEBREW_TAP_TOKEN PAT [P1 — URGENT]:** Added 2026-04-22. PAT was rotated 2026-04-22 with a 90-day expiry — rotation due ~2026-07-15. **Today is 2026-06-25 — only 20 days remain.** This is the most time-sensitive active item in the entire backlog. Annotated with urgency marker `[⚠ DUE ~2026-07-15 — 20 days]`.
  - **[ops] Routine bot PR pile-up [P1]:** Added 2026-05-07. Still open. No duplicate; no near-match. Context is clear (9 stale PRs closed, root cause documented). Not stale enough to flag for removal — still a real unresolved process issue.
  - **[debt] Stale agent worktrees + branches [P1]:** Added 2026-04-20 (65+ days). Persists without resolution. RoutineLog.md shows this pattern hit repeatedly through at least 2026-06-13. Flagging for user as stale P1 needing a decision (accept as known-cost vs. schedule a one-time sweep).
  - **[debt] Testing infrastructure for triggers/sensors [P1]:** Added 2026-04-16 (70+ days). No progress noted. Valid concern; no near-duplicate found.
  - **[bookkeeping] Retroactively trim Backlog entries to NoteStandards [Group A P2]:** Added 2026-04-25 (60+ days). Still valid — many current Backlog entries remain verbose. Low-urgency housekeeping.
  - **Near-duplicate check:** No near-duplicates found across priority tiers. The previously noted overlap between Group C "CHANGELOG.md" and Group D "Changelog generation skill" (flagged 2026-04-21) appears to have been accepted as distinct items.
- **Issues:** PAT expiry requires immediate user attention.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07); checked if findings from those entries are captured in Backlog.
- **Result:** Only one significant entry post-2026-05-07: the 2026-06-13 Plan 40 dispatch log (not a routine — a plan execution). The actual routines (Dependency Audit, Vulnerability Scan, Doc Freshness, etc.) have not been run since 2026-05-04/2026-05-07. **No routine logs exist between 2026-05-07 and 2026-06-25** — a 49-day gap. All routines show overdue status in the dashboard (Next Due dates from 2026-05-11 to 2026-05-21).
  - The Plan 40 and Plan 41 dispatch logs filed new Backlog items directly (security hardening, validate nits, lock-file policy, website npm vulns, remove logic drift) — all captured.
  - No uncaptured routine findings detected.
- **Issues:** The 49-day gap since the last routine execution run is itself a flag for the user — all other routines (Dependency Audit, Vulnerability Scan, Doc Freshness, Memory Consolidation, Roadmap Accuracy, Status Hygiene) are significantly overdue.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed Backlog for items approved for immediate implementation or P0-level urgency requiring workflow routing.
- **Result:** No items meet the bar for autonomous promotion today:
  - The former P0 items are resolved.
  - The PAT rotation (P1 URGENT) requires user action (manual PAT rotation on GitHub), not a code implementation workflow.
  - No user approval on record for any other item.
- **Issues:** none — no workflow dispatch needed.

### Step 7: Log results
- **Action:** Appended entry to station/Logs/RoutineLog.md.
- **Result:** Done.
- **Issues:** none

### Step 8: Update dashboard
- **Action:** Updated routines.md dashboard row for Backlog Hygiene.
- **Result:** Done (Last Ran → 2026-06-25, Next Due → 2026-07-02, Status → done).
- **Issues:** none

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | Both P0 items fully resolved (v0.4.3 + v0.4.2 + Plan 41) but still marked active | Backlog.md P0 section | Commented out both with resolution notes |
| 2 | high | P1 "Full CLI parity" item resolved by Plan 41 (2026-06-16) but not removed from Backlog | Backlog.md P1 | Commented out with Plan 41 resolution note |
| 3 | high | HOMEBREW_TAP_TOKEN PAT expires ~2026-07-15 — only 20 days remain | Backlog.md P1 ops item | Added urgency annotation `[⚠ DUE ~2026-07-15 — 20 days]`; flagged for user |
| 4 | medium | All other routines overdue — 49-day gap since last execution (2026-05-07) | station/agent/Core/routines.md | Flagged for user review — no backlog item needed, but user should schedule routine catch-up |
| 5 | medium | [debt] Stale agent worktrees + branches (P1) — 65+ days without resolution | Backlog.md P1 | Flagged for user decision: sweep now or accept as known cost |
| 6 | low | [bookkeeping] Trim Backlog to NoteStandards (Group A) — 60+ days without action | Backlog.md Group A | Flagged as low-priority stale housekeeping |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[URGENT — 20 days] Rotate HOMEBREW_TAP_TOKEN PAT before ~2026-07-15.** The PAT was issued 2026-04-22 (90-day default). Expired PAT will silently break the Homebrew formula update on the next release — GoReleaser will still publish binaries but the tap formula won't update. Action: go to GitHub → Settings → Developer Settings → Fine-grained tokens, rotate `HOMEBREW_TAP_TOKEN`, update the secret on `LastStep/Bonsai`.

- **All other routines are 35–49 days overdue** (Dependency Audit, Vulnerability Scan, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene — all last ran 2026-05-04 or 2026-05-07). Significant drift may have accumulated. Recommend scheduling a routine-digest session to process them.

- **[debt] Stale agent worktrees + branches (P1, added 2026-04-20)** — 65+ days without resolution. The one-time sweep (`git worktree remove -f -f`, `git branch -D`, `git push origin --delete` for merged-PR branches) could be a quick win. Decide: schedule a housekeeping session, or demote to P2.

## Notes for Next Run

- P0 section is now empty — if it stays empty at next run, it confirms the section header can remain (standard template) without confusion.
- The PAT rotation item (P1 ops) should be checked off or removed at next run — due date is 2026-07-15, which is before the next backlog-hygiene run (2026-07-02). Verify rotation happened.
- All other overdue routines should have run before the next backlog-hygiene — their findings may add new Backlog items to process.
- The Plan 41 headless CLI contract is shipped. Next major P1 item is likely the "Routine bot PR pile-up" ops issue or the stale worktree housekeeping sweep.
