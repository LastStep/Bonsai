---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-19
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
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; cross-referenced each item against Status.md In Progress and Pending.
- **Result:** Both P0 items are fully resolved and shipped — neither needs promotion; both need removal.
  - `[bug] Sensor hook commands use $PWD-walk-up` → SHIPPED v0.4.3 (PR #105/#106, 2026-05-13). Commented out with resolution note.
  - `[feature] bonsai init/add non-interactive flags` → SHIPPED v0.4.2 (PR #102, 2026-05-13). Commented out with resolution note.
- **Issues:** No misplaced P0s (unresolved-but-untracked). Both were resolved after the last hygiene run.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md. Compared In Progress, Pending, and Recently Done against Backlog active items.
- **Result:**
  - Two P0 items removed (see Step 1) — both confirmed resolved in Status.md Recently Done.
  - P1 `[feature] Full agent-drivable CLI parity` — Plan 41 shipped 2026-06-16 (PRs #120-#125): all four cmds (init/add/update/remove) now have headless `*Result` cores + JSONL/exit contract. Item commented out with resolution note.
  - Status.md Pending: only "Trial sentrux" (already commented in Backlog — correctly tracked, no change needed).
  - No Backlog items that are blocking Status.md Pending items were found to be resolvable.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md. Checked P2/P3 Backlog items against current phase milestones.
- **Result:**
  - Phase 1 fully complete (all boxes checked, including `bonsai validate` added 2026-05-07).
  - Phase 2 (Extensibility) is the next phase. Phase 2 milestones:
    - "Self-update mechanism" — mapped to P3 Backlog item. No promotion warranted yet; Phase 2 not actively started.
    - "Template variables expansion" — no Backlog entry exists (gap, noted for user if Phase 2 planning begins).
    - "Micro-task fast path" — mapped to P3 Backlog item. No promotion warranted.
  - No items reference deprecated approaches or completed phases.
- **Issues:** No P2/P3 items warrant promotion at this stage. "Template variables expansion" has no Backlog coverage — not flagging as a gap since Phase 2 planning hasn't started.

### Step 4: Flag stale items
- **Action:** Reviewed all active Backlog items for staleness (30+ days at same priority, unclear rationale, near-duplicates).
- **Result:**
  - **Critical stale finding:** P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — The PAT was due for rotation ~2026-07-15; today is 2026-08-19 (~35 days overdue). Updated the entry to flag urgency and add current status context. This must be actioned before the next release.
  - P1 `[ops] Routine bot PR pile-up` — Added 2026-05-07; root fix (direct-to-main or auto-merge) not yet shipped. Symptom was addressed (9 PRs closed) but recurrence risk remains. Flagged for user below.
  - P1 `[debt] Testing infrastructure for triggers and sensors` — Added 2026-04-16; no progress in ~4 months. Still P1. Flagged.
  - P1 `[debt] Stale agent worktrees + branches accumulating` — Added 2026-04-20; recurring per RoutineLog. No systemic fix shipped. Flagged.
  - No near-duplicates found after P1 "Full agent-drivable CLI parity" was removed (it overlapped with the resolved P0 non-interactive flags item).
- **Issues:** HOMEBREW_TAP_TOKEN PAT is overdue by ~35 days — high urgency.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since 2026-05-07 (last backlog-hygiene run).
- **Result:**
  - 2026-06-13 — Plan 40 dispatch: filed P2 items for symlink hardening, validate drift warning, bonsai-lock policy, Plan 40 review nits — all confirmed in Backlog P2. ✓
  - 2026-06-16 — Plan 41 dispatch: filed P2 for unify remove cinematic/headless logic, P2 for website npm vuln — both in Backlog P2. ✓
  - No uncaptured routine findings requiring new Backlog entries.
- **Issues:** None.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed whether any items are approved for immediate implementation.
- **Result:** No items are pre-approved. HOMEBREW_TAP_TOKEN rotation is operational (not a code task). No dispatch warranted autonomously — presenting to user for decision.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to RoutineLog.md.
- **Result:** Entry written.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for Backlog Hygiene.
- **Result:** `Last Ran` → 2026-08-19, `Next Due` → 2026-08-26, `Status` → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | HOMEBREW_TAP_TOKEN PAT expired ~35 days ago; next release will fail at Homebrew step | Backlog P1 | Entry updated to flag urgency; user action required |
| 2 | info | P0 bug (sensor $PWD-walk-up) was resolved by v0.4.3 in May but still listed as active P0 | Backlog P0 | Commented out with resolution note |
| 3 | info | P0 feature (non-interactive flags) was resolved by v0.4.2 in May but still listed as active P0 | Backlog P0 | Commented out with resolution note |
| 4 | info | P1 feature (full agent-drivable CLI parity) was resolved by Plan 41 in June but still listed as active P1 | Backlog P1 | Commented out with resolution note |
| 5 | low | P1 `[ops] Routine bot PR pile-up` — symptom addressed (9 PRs closed May), root fix not shipped | Backlog P1 | Left in place; flagged for user |
| 6 | low | P1 `[debt] Testing infrastructure` — 4+ months at P1 with no progress | Backlog P1 | Left in place; flagged for user |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

1. **HOMEBREW_TAP_TOKEN PAT OVERDUE (high)** — The Homebrew tap PAT expired ~2026-07-15, now 35+ days past due. Rotate immediately via GitHub Settings → Fine-grained tokens, then `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`. The next release will fail the brew formula update step without this.

2. **Routine bot PR pile-up (low)** — The P1 `[ops]` item about cloud routine PRs accumulating without merge has no root fix shipped. The 9 stale PRs were closed in May, but the cron will keep generating new ones. Decide: (a) commit-direct-to-main, (b) auto-merge if green, or (c) stop creating PRs for already-digested ranges. If cloud routines are no longer running, consider closing this item.

3. **Testing infrastructure for triggers/sensors (low)** — P1 since April 2026, no progress in ~4 months. Either promote to a plan for next session, or explicitly reprioritize to P2 if other work takes precedence.

## Notes for Next Run
- P0 section is now clean (all items commented out). If a new critical bug surfaces, it should be added here.
- Check whether any new Plans have shipped since this run and clean up newly resolved Backlog items.
- Verify HOMEBREW_TAP_TOKEN was rotated — if not, escalate again.
- Plan 41 shipped headless cores; check if any downstream Bonsai-Eval or Odysseus items were also resolved that have Backlog entries.
