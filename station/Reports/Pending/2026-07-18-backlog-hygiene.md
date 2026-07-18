---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-18
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07 (before this run — 72 days overdue)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 6 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`, `station/agent/Routines/backlog-hygiene.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write, Glob, Bash
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; cross-referenced each item against Status.md.
- **Result:** Found 2 resolved P0 items still listed:
  1. `[bug] Sensor hook commands use $PWD-walk-up` — fixed in v0.4.3 (Status.md 2026-05-13, PRs #105/#106). Item had note "Ships v0.4.3" but was never removed.
  2. `[feature] bonsai init / bonsai add need non-interactive flags` — fixed in v0.4.2 (Status.md 2026-05-13, PR #102). Item added 2026-05-08, resolved 5 days later but never cleaned up.
  Both items commented out with resolution notes. No items found in Backlog P0 that are NOT in Status.md — the only current P0-tier item ("Trial sentrux") is correctly in Status.md Pending.
- **Issues:** None blocking. Both removals were clean.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress + Pending + Recently Done; checked all Backlog items against each.
- **Result:**
  - **Resolved item found in P1:** `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` — Plan 41 (PRs #120/#122/#123/#121/#125, 2026-06-16) shipped all mutating commands with pure `*Result` headless cores + JSONL/exit contract. Item was added 2026-06-13 and resolved 3 days later but never removed. Commented out with resolution note.
  - **Pending cross-check:** Only pending item is "Trial sentrux" — blocked by Rust toolchain install. No Backlog item can directly unblock this (the blocker is an ops decision by the user).
  - **No items currently In Progress.**
  - **Debt item filed by Plan 41:** `[debt] Unify remove business logic` (P2) was filed by Plan 41 review — confirmed present in Backlog P2. Good.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; cross-referenced P2/P3 items against current phase milestones.
- **Result:**
  - **Phase 1:** All items checked. Phase 1 is complete per Roadmap checkboxes. No Backlog items reference Phase 1 milestones as pending.
  - **Phase 2 milestones vs Backlog:**
    - `Self-update mechanism` → P3 Backlog item exists (vague — "Skills and workflows should be able to self-heal or flag when they have issues"). Phase 2 milestone; currently P3. Flagged for consideration.
    - `Template variables expansion` → No Backlog item exists. Could capture; flagging for user.
    - `Micro-task fast path` → P3 Backlog item exists in "Future Platform". Aligns with Phase 2; could promote to P2. Flagging.
  - **Phase 3 milestones:** Both (`Managed Agents integration`, `Greenhouse companion app`) have P3 Backlog items — appropriate.
  - **Phase 4 milestones:** No Backlog items for catalog marketplace, plugin system, or cross-project coordination — appropriate (long-range, not yet worth capturing in detail).
  - **No deprecated-approach references found.**
- **Issues:** Phase 2 items `Template variables expansion` and `Micro-task fast path` are underweighted relative to roadmap prominence. Flagged for user review.

### Step 4: Flag stale items
- **Action:** Identified items unchanged in priority for 30+ days without progress. Today is 2026-07-18; last routine run was 2026-05-07 (72 days ago).
- **Result:** Multiple stale P1 items identified:
  1. `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` — added 2026-04-22, rotation due ~2026-07-15. **NOW 3 DAYS OVERDUE.** PAT was set with 90-day default expiry from 2026-04-22, expiry date ~2026-07-21. Rotation window is closing. Urgent flag.
  2. `[ops] Routine bot PR pile-up` — added 2026-05-07, 72 days at P1. No progress captured.
  3. `[debt] Testing infrastructure for triggers and sensors` — added 2026-04-16, 93 days at P1. Large scope; no plan drafted yet.
  4. `[debt] Stale agent worktrees + branches accumulating` — added 2026-04-20, 89 days at P1. Tactical debt; straightforward to execute.
  - P2 security item `[security] Website npm vuln tree — astro upgrade breaks npm run build` — added 2026-06-16, 32 days at P2. 6 active Dependabot alerts. The Astro 6.1.7→6.3.2 PR (#108) has a build break; needs dedicated fix pass. Security items at P2 with active CVEs warrant attention.
  - P2 item `[bug] bonsai validate can't pass on the Bonsai repo itself` — added 2026-06-13, blocks Plan 40 dogfood. Consider P1 promotion.
- **Issues:** HOMEBREW_TAP_TOKEN is urgent and requires immediate user action.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since 2026-05-07 (last backlog-hygiene run).
- **Result:** No routine executions in RoutineLog after 2026-05-07. Plan 40 (2026-06-13) and Plan 41 (2026-06-16) dispatch entries exist but are plan executions, not routine runs. All other routines (Dependency Audit, Vulnerability Scan, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene) are significantly overdue. No routine-generated findings to capture in Backlog — no uncaptured items from routines.
- **Issues:** All routines are 60-70+ days overdue. This is flagged as context; dispatch is at user/loop discretion.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Checked for user-approved or P0-immediate items requiring promotion.
- **Result:** No user approvals present. Resolved P0s were removed (not promoted). No autonomous promotion executed per procedure rule ("Present the item to the user and confirm before starting the workflow"). HOMEBREW_TAP_TOKEN is flagged for user action (not a code implementation item).
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to RoutineLog.md.
- **Result:** Done.

### Step 8: Update dashboard
- **Action:** Updated routines.md dashboard row for Backlog Hygiene.
- **Result:** Done — Last Ran → 2026-07-18, Next Due → 2026-07-25, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | HOMEBREW_TAP_TOKEN PAT rotation overdue — due 2026-07-15, today 2026-07-18; PAT expiry ~2026-07-21 | Backlog P1 | Flagged for user — rotate immediately |
| 2 | medium | P0 sensor hook bug item still in Backlog despite v0.4.3 fix (2026-05-13) | Backlog P0 | Commented out with resolution note |
| 3 | medium | P0 non-interactive flags item still in Backlog despite v0.4.2 fix (2026-05-13) | Backlog P0 | Commented out with resolution note |
| 4 | medium | P1 "Full agent-drivable CLI parity" still in Backlog despite Plan 41 resolution (2026-06-16) | Backlog P1 | Commented out with resolution note |
| 5 | medium | P2 "[security] Website npm vuln tree" — 6 active Dependabot alerts, 32 days no action, PR #108 has build break | Backlog P2 | Flagged for user review |
| 6 | low | P2 "bonsai validate can't pass on Bonsai repo" — blocks Plan 40 dogfood; consider P1 promotion | Backlog P2 | Flagged for user consideration |
| 7 | low | P1 "Routine bot PR pile-up" — 72 days at P1 with no progress | Backlog P1 | Flagged for re-prioritization |
| 8 | low | P1 "Testing infrastructure for triggers and sensors" — 93 days at P1, no plan drafted | Backlog P1 | Flagged for re-prioritization |
| 9 | low | P1 "Stale agent worktrees + branches" — 89 days at P1 | Backlog P1 | Flagged for re-prioritization |
| 10 | info | Phase 2 roadmap items "Template variables expansion" and "Micro-task fast path" have no P1/P2 Backlog entries matching their roadmap weight | Roadmap Phase 2 | Flagged for user consideration |
| 11 | info | All other routines 60-70+ days overdue (last ran 2026-05-04/07) | routines.md dashboard | Flagged — dispatch at user discretion |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **URGENT — HOMEBREW_TAP_TOKEN PAT rotation:** The fine-grained PAT set on 2026-04-22 with 90-day default expiry is due for rotation NOW (reminder date was 2026-07-15; today is 2026-07-18; estimated hard expiry ~2026-07-21). If not rotated before the next release attempt, GoReleaser will fail at the brew tap step with `401 Bad credentials`. Action: rotate the PAT in GitHub → Developer Settings → Fine-grained PATs, then update the `HOMEBREW_TAP_TOKEN` secret in `LastStep/Bonsai` repo settings.

2. **Website npm vulns (P2 security):** 6 active Dependabot alerts on `/website` (esbuild HIGH, vite HIGH+MED, js-yaml MED, astro, astro+esbuild LOW). PR #108 (Astro 6.1.7→6.3.2 bump) fails the website build after rebase — needs a real upgrade-with-build-fix pass, not auto-merge. Recommend scheduling a dedicated vulnerability-scan routine run.

3. **Stale P1 items for re-prioritization:** Items 7/8/9 above have been at P1 for 72-93 days with no progress. Consider: (a) drafting plans to make them actionable, (b) demoting to P2 if deprioritized, or (c) accepting current status.

4. **Phase 2 roadmap coverage:** "Template variables expansion" has no Backlog entry. If it's actively planned for Phase 2, capture it. "Micro-task fast path" and "Self-update mechanism" are in P3 but are Phase 2 milestones — consider promoting to P2 to match roadmap.

5. **bonsai validate on Bonsai repo (P2 bug):** `.bonsai-lock.yaml` is gitignored, making validate report 38 orphan errors on the self-hosted station. Blocks dogfood for Plan 40. Consider P1 promotion since it affects daily development workflow.

## Notes for Next Run

- Both P0 slots are now empty (all resolved). P0 section contains only HTML comments. If new critical bugs arise, add directly.
- HOMEBREW_TAP_TOKEN PAT — verify rotation was done. If another 90-day PAT was set, next reminder date should be ~Oct 2026.
- The 72-day gap since last routine run means multiple other routines are also significantly overdue. Dependency Audit and Vulnerability Scan are highest priority given the active npm security items.
- Plan 41 is still in `Plans/Active/` despite being shipped (Status.md 2026-06-16). Status Hygiene routine should archive it.
- Plan 40 is also in `Plans/Active/` with Phase 4 HELD — correct status, but worth noting Phase 4 is blocked on lock-file policy decision.
