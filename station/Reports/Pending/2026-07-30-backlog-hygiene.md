---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-30
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
- **Duration:** ~6 min
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write — no Bash commands or external tools
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section; checked each P0 item against `Status.md` In Progress / Pending / Recently Done tables.
- **Result:** Found 2 P0 items. Both appear fully resolved:
  - `[bug] Sensor hook commands use $PWD-walk-up` — confirmed resolved by v0.4.3 hotfix (PRs #105/#106, Status.md 2026-05-13 Recently Done). Backlog entry retained the "Ships v0.4.3" tag but was never cleaned up.
  - `[feature] bonsai init / bonsai add need non-interactive flags` — confirmed resolved by v0.4.2 (PR #102, Status.md 2026-05-13 Recently Done).
  - `[research] Trial sentrux` is correctly commented out (promoted to Status.md Pending 2026-05-07, still blocked on Rust toolchain).
- **Issues:** No live P0 items were found in Backlog that should be in Status.md. Both previously present P0s were resolved, not missing from Status.

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md` (In Progress, Pending, Recently Done). Cross-checked all Backlog items against Status.md entries.
- **Result:**
  - 2 P0 Backlog items matched Recently Done rows → removed from Backlog (commented out with audit trail).
  - 1 P1 item matched Recently Done: `[feature] Full agent-drivable CLI parity` → Plan 41 shipped all headless cores + JSONL/exit contract (Status.md 2026-06-16) → removed from Backlog.
  - Status.md Pending: only `[research] Trial sentrux` — already correctly commented out of P0 in Backlog.
  - No Status.md Pending items were found with "Blocked By" fields resolvable via a Backlog item. The sentrux item is blocked on Rust toolchain (external dependency).
- **Issues:** None — cross-reference was clean after the 3 removals.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md`. Checked Phase 1 (current) and Phase 2 milestones against P2/P3 Backlog items.
- **Result:**
  - Roadmap Phase 1: all items are `[x]` — no open Phase 1 work to cross-reference.
  - Phase 2 items (Extensibility): "Self-update mechanism" has a P3 Big Bets entry; "Micro-task fast path" has a P3 Future Platform entry; "Template variables expansion" has no Backlog entry but is a low-priority Phase 2 item — acceptable.
  - No P2/P3 Backlog items flagged for promotion; current phase (Phase 1) is complete and Phase 2 work is not being prioritized yet.
  - No items referencing deprecated approaches or completed phases found that weren't already handled.
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Reviewed all Backlog items for age and staleness. Last run was 2026-05-07; today is 2026-07-30 (84 days elapsed). Most items were added in April–June 2026.
- **Result:**
  - **HOMEBREW_TAP_TOKEN PAT expiry (P1)** — added 2026-04-22, reminder set for ~2026-07-15. Today is 2026-07-30 = 9 days past the reminder date and ~9 days past estimated 90-day expiry (2026-04-22 + 90d = ~2026-07-21). **This is overdue and should be acted on before the next release.** Flagging for user.
  - **[security] Website npm vuln tree (P2)** — added 2026-06-16 (44 days ago). HIGH severity vulns (esbuild, vite). Has not been addressed. Flagging for user.
  - **[ops] Routine bot PR pile-up (P1)** — added 2026-05-07. The immediate 9-PR crisis was resolved, but the root-cause fix (a/b/c options) was never implemented. 84 days at P1 without progress. Consider downgrading to P2 if bot-driven maintenance is no longer actively pushing PRs, or picking up one of the three fix options.
  - **Group B items** (testing infrastructure, break up generate.go, etc.) — these are mostly P1 debt items from April 2026. No progress in 3+ months. Not flagging individually but noting that Group B represents significant unfunded debt.
  - P3 items are appropriately parked.
  - No items found with genuinely unclear context or rationale; all entries have source attribution and rationale.
  - **Near-duplicate check:** No exact duplicates found. The "Plans Index file" and "Plan archiving" Group E items are related but distinct (one is an index, one is folder structure). The Group C changelog item and Group D changelog skill item remain distinct (the Group C entry is about an OSS process, Group D is about a catalog skill). These were already flagged in a prior run (2026-04-21) — still unresolved but not duplication-critical.
- **Issues:** 3 items flagged (see above).

### Step 5: Check for routine-generated items
- **Action:** Read `Logs/RoutineLog.md` for entries since 2026-05-07 (last backlog-hygiene run).
- **Result:** Only one log entry found after 2026-05-07: "2026-06-13 — Plan 40 dispatch" — this is a plan execution log, not a routine report. No routine reports were filed between 2026-05-07 and today (84-day gap, all routines significantly overdue by Dependency Audit / Vulnerability Scan / Doc Freshness Check / Memory Consolidation / Roadmap Accuracy / Status Hygiene standards). **No routine-generated findings are uncaptured in the Backlog** because no routines ran in this window.
- **Issues:** No uncaptured routine findings to add. However, flagging the 84-day routine gap for user awareness — when other routines next run, they may surface findings that require Backlog entries.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Evaluated whether any Backlog item is ready for promotion to issue-to-implementation workflow.
- **Result:** No items flagged as user-approved for immediate implementation. The user should review the HOMEBREW_TAP_TOKEN urgency (flags above); if confirmed expired, that's an immediate ops action (not a planned implementation). No other items have sufficient urgency or clear approval to route through the workflow autonomously.
- **Issues:** None — deferring to user.

### Step 7 & 8: Log and dashboard update
- **Action:** Appended entry to `Logs/RoutineLog.md` and updated `agent/Core/routines.md` dashboard row.
- **Result:** Completed as part of post-procedure steps.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 "sensor hook $PWD-walk-up" was still in Backlog despite v0.4.3 fix | Backlog.md P0 section | Removed (commented out with audit trail) |
| 2 | resolved | P0 "non-interactive flags" was still in Backlog despite v0.4.2 fix | Backlog.md P0 section | Removed (commented out with audit trail) |
| 3 | resolved | P1 "Full agent-drivable CLI parity" was still in Backlog despite Plan 41 shipping | Backlog.md P1 section | Removed (commented out with audit trail) |
| 4 | high | HOMEBREW_TAP_TOKEN PAT likely expired (~2026-07-21, 90-day window from rotation 2026-04-22) | Backlog.md P1 "[ops] HOMEBREW_TAP_TOKEN" | Flagged for user — no change to entry |
| 5 | medium | Website npm HIGH severity vulns (esbuild, vite) outstanding 44 days without action | Backlog.md P2 "[security] Website npm vuln tree" | Flagged for user — no change to entry |
| 6 | low | P1 "[ops] Routine bot PR pile-up" root cause never fixed, 84 days at P1 | Backlog.md P1 | Flagged for user — consider downgrade or pick up |
| 7 | info | All routines significantly overdue (84+ day gap since last run) | RoutineLog.md | Flagged for user awareness |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **HOMEBREW_TAP_TOKEN PAT likely expired** — The fine-grained PAT was rotated 2026-04-22. At 90-day default expiry, it expired approximately 2026-07-21 (9 days ago). If a release is attempted before rotation, GoReleaser will fail at the Homebrew formula update step with a 401 error. **Recommend rotating immediately** at `LastStep/Bonsai` repo secrets. The Backlog P1 entry `[ops] HOMEBREW_TAP_TOKEN PAT expiry` can be removed after rotation.

2. **Website npm HIGH severity vulnerabilities** — The P2 `[security] Website npm vuln tree` entry (added 2026-06-16) documents HIGH severity esbuild and HIGH+MED vite vulnerabilities outstanding for 44 days. The astro 6.1.7→6.3.2 bump (PR #108) that would clear most fails the website build after rebase. Recommend picking up a fix pass: upgrade astro with build-fix, bump vite/js-yaml, or pin patched transitives.

3. **P1 "Routine bot PR pile-up" staleness** — The root cause fix was never implemented (87 days at P1). If the cloud routine bot is still active, PRs may be accumulating again. Recommend checking the PR list on `LastStep/Bonsai` and deciding: implement one of the three fix options (commit-direct-to-main, auto-merge-if-green, or skip-when-local-digest-absorbed), or downgrade to P2 if the bot frequency has changed.

4. **All routines overdue** — The 84-day gap since last routine execution (2026-05-07) means Dependency Audit, Vulnerability Scan, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene, and Backlog Hygiene are all significantly overdue. Recommend scheduling a routine-digest session to run all pending routines and process findings.

---

## Notes for Next Run

- P0 section in Backlog is now empty (all items resolved or commented out). If new P0s arise during future sessions, the section will repopulate.
- The HOMEBREW_TAP_TOKEN item (P1) should be resolved and its Backlog row removed on the next session where the user rotates the PAT.
- After the website npm vuln fix ships, the P2 `[security] Website npm vuln tree` row should be removed.
- The `[ops] Routine bot PR pile-up` P1 item should be either picked up or explicitly downgraded based on current bot activity.
- The backlog is otherwise healthy — items are well-structured with source attribution, groups are coherent, and priorities appear appropriately assigned given project pace.
