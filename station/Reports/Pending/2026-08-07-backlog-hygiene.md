---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-07
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
- **Files Read:** 5 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read (file reads), Edit (backlog cleanup)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section. Two active P0 items found. Cross-referenced each against `Status.md` In Progress and Pending tables.
- **Result:**
  - `[bug] Sensor hook commands use $PWD-walk-up` — NOT in Status.md In Progress/Pending because it is RESOLVED. Status.md "Recently Done" confirms v0.4.3 hotfix shipped 2026-05-13 (PRs #105/#106) with absolute-path baking. This P0 should have been removed at ship time.
  - `[feature] bonsai init / bonsai add need non-interactive flags [Plan 38 P2 blocker]` — NOT in Status.md In Progress/Pending because it is RESOLVED. Status.md "Recently Done" confirms v0.4.2 shipped 2026-05-13 (PR #102) with exactly these flags.
  - Both P0 items are stale resolved entries, not escalation candidates. Commented out in Backlog.md with resolution notes.
- **Issues:** None — both items were resolved, not missing from Status.md due to oversight.

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md` fully. Checked In Progress, Pending, and Recently Done against all Backlog entries.
- **Result:**
  - **In Progress:** empty — no conflicts.
  - **Pending:** `[research] Trial sentrux on Bonsai repo` — already reflected in Backlog as an HTML comment (promoted 2026-05-07). No action needed. Blocked by Rust toolchain — no Backlog item addresses this blocker directly; user must install rustup manually.
  - **Recently Done matches found:**
    - Plan 41 (shipped 2026-06-16) resolves P1 `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove`. Plan 41 delivered JSONL/exit-code contract across all four mutating cmds (`init/add/update/remove`) plus `list --json`. Item commented out in Backlog.md.
    - v0.4.3 resolves P0 bug #1 (handled in Step 1).
    - v0.4.2 resolves P0 feature #2 (handled in Step 1).
  - **Blocked-by cross-check:** Sentrux trial is blocked by Rust toolchain install. No Backlog item would unblock it.
- **Issues:** None beyond stale items cleaned up.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md`. Cross-referenced P2/P3 Backlog items against phase milestones.
- **Result:**
  - Phase 1 is fully complete — all items checked `[x]`. No stale unchecked boxes found (previous routine-digest fixed these in 2026-05-07).
  - Phase 2 `[ ] Self-update mechanism` → Backlog P3 entry exists (`[improvement] Self-update mechanism`). Alignment noted but no promotion recommended without user input.
  - Phase 2 `[ ] Micro-task fast path` → Backlog P3 entry exists (`[improvement] Micro-task fast path`). Same.
  - Phase 2 `[ ] Template variables expansion` → No explicit Backlog item found. Minor gap — this Phase 2 milestone has no tracking entry.
  - Phase 3 `[ ] Managed Agents integration` + `[ ] Greenhouse companion app` → Backlog P3 Big Bets entries exist. Aligned.
  - No items referencing deprecated approaches or completed phases found.
- **Issues:** Phase 2 "Template variables expansion" milestone has no Backlog tracking entry. Flagged for user review.

### Step 4: Flag stale items
- **Action:** Scanned all Backlog entries for age (30+ days since 2026-07-08), items lacking context, and near-duplicates.
- **Result:**
  - **URGENT — HOMEBREW_TAP_TOKEN PAT rotation overdue:** P1 item set a reminder for ~2026-07-15. Today is 2026-08-07 — 3+ weeks past deadline. PAT may be expired. Added `[URGENT 2026-08-07]` tag and expanded rotation instructions in Backlog.md entry.
  - **Stale cluster (4+ months old, April 2026 additions):** Multiple P1 items have had no progress in 4+ months:
    - `[ops] Routine bot PR pile-up` (2026-05-07, P1) — 3 months with no fix to cloud routine behavior.
    - `[debt] Testing infrastructure for triggers and sensors` (2026-04-16, P1) — 4 months.
    - `[debt] Stale agent worktrees + branches` (2026-04-20, P1) — 4 months (worktrees may have grown).
    - Group A bookkeeping, Group B debt items, Group C OSS items, Group D catalog expansion, Group E workspace improvements — most added April 2026, no recent progress.
  - **Near-duplicates:** None found that weren't already noted in previous hygiene runs. The CHANGELOG duplication between Group C and Group D (flagged 2026-04-21) is still present but unchanged.
  - **Items lacking context:** None — all entries have adequate rationale.
- **Issues:** HOMEBREW_TAP_TOKEN PAT urgency flagged (see Findings Summary). Stale item cluster noted but not actioned without user direction.

### Step 5: Check for routine-generated items
- **Action:** Reviewed `RoutineLog.md` for entries since last backlog-hygiene run (2026-05-07).
- **Result:**
  - RoutineLog entries since 2026-05-07: 2026-05-07 batch (Roadmap Accuracy, Status Hygiene, Backlog Hygiene, Routine Digest, Memory Consolidation) + 2026-06-13 Plan 40 session + 2026-06-16 Plan 41 session items.
  - Plan 40 session (2026-06-13) filed P2 items to Backlog: symlink hardening, validate drift, review nits, validate lockfile policy. All confirmed present in Backlog.md P2 section.
  - Plan 41 session (2026-06-16) filed P2 items: website npm vuln, remove-logic duplication. Both confirmed present in Backlog.md P2 section.
  - **Note:** No routine executions between 2026-05-07 and today (2026-08-07) — 3+ months gap. All routines are significantly overdue per the dashboard (all Next Due dates are in May 2026). This is a systemic observation, not a backlog item.
  - No uncaptured routine findings found.
- **Issues:** None — all routine-generated findings are captured.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed Backlog for items approved for promotion or requiring immediate action.
- **Result:** No user-approved promotions to action. The HOMEBREW_TAP_TOKEN item (P1) may need immediate user attention before any release, but this is a manual action (rotate PAT in GitHub UI) not an implementation workflow item.
- **Issues:** None autonomously actioned. User must confirm before any promotion.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.

### Step 8: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` Backlog Hygiene row: Last Ran → 2026-08-07, Next Due → 2026-08-14, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | P0 bug "sensor hook $PWD-walk-up" is stale — resolved by v0.4.3 (2026-05-13) but never removed | `Backlog.md` P0 | Commented out with resolution note |
| 2 | HIGH | P0 feature "non-interactive flags" is stale — resolved by v0.4.2 (2026-05-13) but never removed | `Backlog.md` P0 | Commented out with resolution note |
| 3 | HIGH | P1 "Full agent-drivable CLI parity" is stale — resolved by Plan 41 (2026-06-16) but never removed | `Backlog.md` P1 | Commented out with resolution note |
| 4 | HIGH | HOMEBREW_TAP_TOKEN PAT rotation deadline (~2026-07-15) has passed — 3+ weeks overdue | `Backlog.md` P1 | Added URGENT tag + rotation instructions; **flagged for user** |
| 5 | MEDIUM | Phase 2 Roadmap milestone "Template variables expansion" has no Backlog tracking entry | `Roadmap.md` Phase 2 | **Flagged for user** — no auto-add per procedure |
| 6 | LOW | Sentrux trial (Status Pending) blocked by Rust toolchain — no Backlog item addresses the blocker | `Status.md` Pending | **Flagged for user** — installing rustup is out of scope for this routine |
| 7 | INFO | All 7 routines are 3+ months overdue (last ran 2026-05-04/05-07, Next Due dates all in May 2026) | `routines.md` dashboard | Noted — systemic observation, not a backlog item |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[URGENT] HOMEBREW_TAP_TOKEN PAT rotation overdue** — deadline 2026-07-15 passed. Rotate immediately via GitHub → Settings → Developer settings → Fine-grained PATs, then `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`. Next release will fail brew step if expired.

2. **[LOW] Phase 2 "Template variables expansion" has no Backlog entry** — user should decide: add a P2/P3 tracking entry for this Roadmap milestone, or remove the milestone if it's no longer planned.

3. **[INFO] Sentrux trial blocker** — the Pending item in Status.md requires `rustup` + `cargo` installed. No autonomous path to unblock.

4. **[INFO] All other routines are 3+ months overdue** — consider running a routine-digest sweep across the overdue batch (Dependency Audit, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene, Vulnerability Scan).

## Notes for Next Run

- P0 section is now effectively empty (only HTML comments remain). If a new critical issue emerges, it should be the only P0 entry.
- HOMEBREW_TAP_TOKEN urgency flag in P1 section — once rotated, the URGENT tag should be downgraded back to a periodic reminder.
- The 4-month stale cluster in P1 (testing infra, worktrees, bot PR pile-up) should be triaged at the next planning session — either promote to plans or demote to P2.
- Plan 41 resolved the headless CLI parity work — any remaining CLI polish items should be checked against the shipped `docs/agent-interface.md` contract doc to confirm they're not already covered.
