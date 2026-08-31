---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-31
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
- **Duration:** ~10 min
- **Files Read:** 7 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`, `station/agent/Core/memory.md`, `station/agent/Routines/backlog-hygiene.md`
- **Files Modified:** 4 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Reports/Pending/2026-08-31-backlog-hygiene.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; checked each item against Status.md.
- **Result:** Both P0 items are RESOLVED — not misplaced escalations, but stale resolved entries that should have been cleaned up months ago.
  - `[bug] Sensor hook commands use $PWD-walk-up` → resolved by v0.4.3 (Status.md 2026-05-13)
  - `[feature] bonsai init/add need non-interactive flags` → resolved by v0.4.2 (Status.md 2026-05-13)
- **Issues:** P0 section was empty of genuine blockers — no active P0s exist. Both entries should have been removed at last hygiene run (2026-05-07) but were added after that run.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress, Pending, and Recently Done tables; matched against all Backlog entries.
- **Result:**
  - Removed 2 resolved P0 items (v0.4.3 + v0.4.2 deliveries; see Step 1).
  - Removed 1 resolved P1 item: `[feature] Full agent-drivable non-interactive CLI parity` — Plan 41 shipped headless `*Result` cores + JSONL/exit contract for all four commands + `docs/agent-interface.md` (Status.md 2026-06-16). This P1 was added 2026-06-13 and immediately superseded by Plan 41.
  - Pending item `[research] Trial sentrux on Bonsai repo` is in Status.md Pending (blocked on Rust toolchain); Backlog comment already references this — no action needed.
  - No "Blocked By" items in Status.md Pending that a Backlog item could unblock (sentrux is blocked on toolchain, not a Backlog item).
- **Issues:** none

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; matched P2/P3 Backlog items against phase milestones.
- **Result:**
  - Phase 1 is fully complete (all boxes checked). No stale P0/P1 Backlog items reference Phase 1 goals (after the removals above).
  - Phase 2 milestones — two P3 Backlog items directly align with unchecked Phase 2 goals:
    - `[improvement] Self-update mechanism` (P3 Big Bets) ↔ Phase 2 "Self-update mechanism"
    - `[improvement] Micro-task fast path` (P3 Future Platform) ↔ Phase 2 "Micro-task fast path"
    - Flagging both for possible P3→P2 promotion now that Phase 1 is complete.
  - No deprecated approaches or completed-phase references found in remaining items.
- **Issues:** Flagged 2 P3 items for potential promotion (see Findings Summary).

### Step 4: Flag stale items
- **Action:** Scanned all items for 30+ day staleness, unclear rationale, and near-duplicates.
- **Result:**
  - **CRITICAL STALE — HOMEBREW_TAP_TOKEN expired:** `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` (P1) was added 2026-04-22; PAT was set 2026-04-22 with 90-day default expiry; reminder date ~2026-07-15. Today is 2026-08-31 — the PAT has been expired for ~41 days. Next release will fail at the brew step with `401 Bad credentials`. Requires immediate user action (rotate PAT, set secret on `LastStep/Bonsai`).
  - **All routines severely overdue:** Last routine runs were 2026-05-04 to 2026-05-07. All routines are 107–116 days overdue. Dependency Audit, Vulnerability Scan, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene are all far past their Next Due dates.
  - **Plan 41 file needs archiving:** Memory flags "Plan 41 file still in Plans/Active/ — archive to Plans/Archive/ at next wrap-up." This is pending housekeeping.
  - **MCP server (Plan 42) not in Backlog:** Memory Work State explicitly lists "MCP server = Plan 42 (go-sdk, stdio `bonsai mcp`)" as an open follow-up and calls it "Backlog P2" — but no corresponding item exists in Backlog.md. Flagged for user to add.
  - Group F items (`[docs] Document AltScreen behavior change`, `[docs] Fill "Deviations from Plan"`) are 4+ months old with no progress. Still relevant as documentation best-practices items — no duplicate found, keeping but flagging.
  - `[bookkeeping] Retroactively trim Backlog entries to NoteStandards` (Group A, P2) has been in the backlog since 2026-04-25 (4+ months). Clear rationale exists, but no progress made. Flagged for re-prioritization.
  - No near-duplicates found across priority tiers.
- **Issues:** 4 items flagged for user (see below).

### Step 5: Check for routine-generated items since last run
- **Action:** Read RoutineLog.md for entries since 2026-05-07.
- **Result:** No routine log entries exist after 2026-05-07. All routines have been dormant for 116 days. The previous routine-digest (2026-05-04) filed 6 backlog items — all confirmed present in current Backlog. No uncaptured findings to add.
- **Issues:** The extended gap between routine runs is itself a notable finding — no routine-originated issues have been captured since May.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any item is ripe for immediate promotion.
- **Result:** No items are pre-approved for immediate implementation. HOMEBREW_TAP_TOKEN is the most time-sensitive and requires user direct action (not an issue-to-implementation workflow — it's a credential rotation). Presenting all flagged items to user for decision.
- **Issues:** none

### Step 7 & 8: Log results and update dashboard
- **Action:** Appending to RoutineLog.md and updating routines.md dashboard.
- **Result:** Done.
- **Issues:** none

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | HOMEBREW_TAP_TOKEN PAT expired ~41 days ago — next release will fail at brew step | Backlog P1 item, repo secrets | Flagged for immediate user action |
| 2 | HIGH | All 6 other routines severely overdue (107–116 days) | routines.md dashboard | Flagged for user scheduling |
| 3 | MEDIUM | P0 section had 2 resolved items (v0.4.3 + v0.4.2) still listed | Backlog P0 | Removed; resolution comments added |
| 4 | MEDIUM | P1 had 1 resolved item (Plan 41 headless cores) still listed | Backlog P1 | Removed; resolution comment added |
| 5 | MEDIUM | MCP server (Plan 42) referenced in memory as Backlog P2 — not in Backlog.md | memory.md Work State | Flagged for user to add |
| 6 | MEDIUM | Plan 41 file still in Plans/Active/ — should be archived | Plans/Active/ | Flagged for user |
| 7 | LOW | 2 Phase 2 Roadmap items (self-update, micro-task fast path) sit at P3 — Phase 1 is complete | Backlog P3 | Flagged for potential P3→P2 promotion |
| 8 | LOW | Group A bookkeeping sweep 4+ months stale | Backlog P2 Group A | Flagged for re-prioritization |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **[URGENT] Rotate HOMEBREW_TAP_TOKEN now.** The PAT was set 2026-04-22 with 90-day default expiry and is ~41 days past due. Rotate at GitHub → Settings → PAT → fine-grained, set on `LastStep/Bonsai` repo secret. Symptom if expired: release pipeline succeeds through binary publish, then fails at brew step with `401 Bad credentials`. See memory.md note on recovery procedure.
- **[HIGH] Schedule overdue routines.** All 6 other routines (Dependency Audit, Vulnerability Scan, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene) are 107–116 days overdue. Recommend running each in priority order: Vulnerability Scan + Dependency Audit first (security), then Status Hygiene + Memory Consolidation, then Doc Freshness Check + Roadmap Accuracy.
- **[MEDIUM] Add MCP server (Plan 42) to Backlog P2.** Memory Work State calls it "Backlog P2" but it has no entry in Backlog.md. Suggested entry: `- **[feature] MCP server — Plan 42 (go-sdk, stdio `bonsai mcp`)** — headless contract from Plan 41 built for this; implement as `bonsai mcp` stdio server using go-sdk. *(added 2026-06-16, source: Plan 41 follow-up)*`
- **[MEDIUM] Archive Plan 41 file.** `Plans/Active/41-headless-cli-contract.md` was flagged in memory for archiving to `Plans/Archive/` at next wrap-up. Has not been done.
- **[LOW] Consider promoting P3 Phase 2 items to P2.** Phase 1 is complete. `[improvement] Self-update mechanism` and `[improvement] Micro-task fast path` both have direct Roadmap Phase 2 entries. Promote if Phase 2 work is the current focus.
- **[LOW] Decide on Group A bookkeeping sweep.** `[bookkeeping] Retroactively trim Backlog entries` has been P2 since 2026-04-25 with no progress. Either schedule it, demote to P3, or remove if the policy is "new entries follow NoteStandards, old ones are grandfathered."

---

## Notes for Next Run

- The gap between 2026-05-07 and 2026-08-31 (116 days) is abnormal — 16 missed 7-day cycles. Consider ensuring the loop.md dispatch is reliably triggered.
- After running all overdue routines, a routine-digest pass will be needed to synthesize findings across all reports.
- After HOMEBREW_TAP_TOKEN rotation, update memory.md with the new rotation date so the next PAT-expiry window is tracked.
- After Plan 41 archiving and MCP server Backlog entry, the Backlog P1/P2 sections will be accurate.
