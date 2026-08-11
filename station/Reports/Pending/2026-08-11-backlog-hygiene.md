---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-11
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~6 min
- **Files Read:** 5
  - `station/Playbook/Backlog.md`
  - `station/Playbook/Status.md`
  - `station/Playbook/Roadmap.md`
  - `station/Logs/RoutineLog.md`
  - `station/agent/Core/routines.md`
- **Files Modified:** 3
  - `station/Playbook/Backlog.md` (3 resolved items commented out, 1 P1 removal)
  - `station/agent/Core/routines.md` (dashboard updated)
  - `station/Logs/RoutineLog.md` (entry appended)
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog P0 section and cross-referenced each item against Status.md.
- **Result:** Both P0 items were found to be fully resolved by shipped releases:
  - `[bug] Sensor hook commands use $PWD-walk-up` — Fixed by v0.4.3 hotfix (PRs #105/#106, shipped 2026-05-13). Status.md confirms: "v0.4.3 hotfix shipped — sensor hook commands now bake install-time absolute paths."
  - `[feature] bonsai init/add need non-interactive flags` — Fixed by v0.4.2 release (shipped 2026-05-13). Status.md confirms: `--non-interactive --from-config <path>` shipped in both commands.
  - Both items replaced with resolved HTML comments in Backlog.md. P0 section now shows no active items.
- **Issues:** None — no true P0 escalation needed. Both were already fixed.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md in full; checked all In Progress, Pending, and Recently Done rows against Backlog items.
- **Result:**
  - **In Progress:** empty — no cross-ref needed.
  - **Pending:** "Trial sentrux" — already commented-out in Backlog P0 with note "promoted to Status.md Pending 2026-05-07." Consistent.
  - **Recently Done items with Backlog impact:**
    - Plan 41 (headless CLI, shipped 2026-06-16) resolves the P1 item "Full agent-drivable (non-interactive) CLI parity: init / update / add / remove." Removed from Backlog with resolved comment.
    - v0.4.3 hotfix and v0.4.2 release handled in Step 1.
    - "Windows cross-compile CI gate" — Status.md notes "Backlog P2 row cleared" (done previously).
    - "Root CLAUDE.md Go drift fix" — Status.md notes "Backlog row cleared" (done previously).
  - **Blocked-by check:** Sentrux Pending blocked on Rust toolchain install. No Backlog item directly unblocks it (toolchain install is a user/ops action).
- **Issues:** None beyond items already actioned.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; checked P2/P3 Backlog items against current phase milestones.
- **Result:**
  - Phase 1 is fully complete (`[x]` on all items). No stale phase-1-referenced Backlog items found.
  - Project is now in Phase 2 (Extensibility). Three P3 Backlog items align with Phase 2 goals and could be candidates for promotion:
    - "Custom item creator" (P3, Future Platform) — directly maps to Phase 2 "custom catalog items"
    - "Self-update mechanism" (P3, Big Bets) — directly listed in Phase 2 roadmap
    - "Micro-task fast path" (P3, Future Platform) — maps to Phase 2 lightweight protocol bypass
  - Flagging these for user consideration — promotion to P2 or P1 depends on priority vs. other active work.
  - No items reference deprecated approaches or completed phases.
- **Issues:** None requiring immediate action.

### Step 4: Flag stale items
- **Action:** Reviewed all Backlog items for age (30+ days without progress) and clarity.
- **Result:**
  - **CRITICAL — HOMEBREW_TAP_TOKEN PAT OVERDUE:** P1 item added 2026-04-22 set a calendar reminder for ~2026-07-15 rotation. Today is 2026-08-11 — 21+ days past the reminder, ~21 days past the estimated 90-day expiry (~2026-07-21). The PAT is very likely expired. Any next release will fail at the Homebrew tap step with "401 Bad credentials." Requires immediate user action.
  - **Long-stale P1 items (100+ days, no progress):**
    - "HOMEBREW_TAP_TOKEN PAT expiry" — 111 days, URGENT (see above)
    - "Routine bot PR pile-up" — 96 days, process fix still pending
    - "Testing infrastructure for triggers and sensors" — 117 days
    - "Stale agent worktrees + branches" — 113 days
  - **Long-stale P2 items (50+ days):**
    - "Harden scaffolding writes against symlink" — 59 days (added post-Plan 40)
    - "bonsai validate warn on project.yaml drift" — 59 days
    - "Plan 40 review nits" — 59 days
    - "bonsai validate can't pass on Bonsai repo" — 59 days (blocks dogfood)
    - "Website npm vuln tree / astro upgrade" — 56 days (6 open Dependabot alerts, build broken)
    - "Unify remove business logic" — 56 days (debt from Plan 41 Phase 3)
    - "Integrate plan-grilling as catalog ability" — 59 days
  - **Historical stale items (Groups A/B/C/D/E/F — 100–120 days):** These have been carried since April 2026. Many require user recording (demo GIF), architectural decisions (integration variants), or significant implementation effort. No individual item is clearly wrong or abandoned — they're simply low-priority queue.
  - **Near-duplicates:** No new near-duplicates found. The CHANGELOG consolidation near-duplicate (Group C vs Group D) was flagged in the 2026-04-21 run and remains unresolved.
- **Issues:** HOMEBREW_TAP_TOKEN PAT expiry is time-sensitive and flagged for immediate user review.

### Step 5: Check routine-generated items since last run
- **Action:** Reviewed RoutineLog.md entries since 2026-05-07 (last backlog-hygiene run).
- **Result:**
  - Only significant session since then: Plan 40 dispatch (2026-06-13). That session added 6 Backlog items directly: P2 symlink hardening, P2 validate drift warn, P2 Plan 40 review nits, P2 validate-can't-pass, P2 website npm vulns, P2 unify-remove logic. All 6 are captured in Backlog.md.
  - No routines have run since 2026-05-04/2026-05-07 (Dependency Audit, Vulnerability Scan, Doc Freshness, Memory Consolidation, Status Hygiene, Roadmap Accuracy — all overdue by 88-99 days). This is a finding: the routine schedule has been idle for ~96 days. Any issues found by those routines since then are uncaptured.
  - One notable gap: no Dependency Audit or Vulnerability Scan has run since 2026-05-04. Given the website npm vuln tree P2 item mentions 6 open Dependabot alerts, and 99 days have passed, new vulnerabilities may exist uncaptured.
- **Issues:** All June-2026 session findings are captured. However, ~96 days of uncaptured routine output is flagged for user awareness.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any items are approved/ready for immediate implementation routing.
- **Result:** No items are pre-approved for autonomous implementation. Flagging the HOMEBREW_TAP_TOKEN PAT rotation to user for priority decision — that's an ops action, not a code dispatch. All other promotions (Phase 2-aligned P3 items, P1 stale items) require user confirmation before routing.
- **Issues:** None.

### Step 7 & 8: Log results and update dashboard
- **Action:** Appending RoutineLog.md entry, updating routines.md dashboard.
- **Result:** Completed.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | HOMEBREW_TAP_TOKEN PAT likely expired (~2026-07-21 expiry, now 2026-08-11 — 21 days overdue). Next release will fail Homebrew step. | Backlog P1 | Flagged for user — immediate rotation required |
| 2 | medium | Both Backlog P0 items were already resolved (v0.4.3 bug fix + v0.4.2 feature). P0 section had stale live items for 96 days. | `Backlog.md` P0 | Commented out with resolution notes |
| 3 | medium | Backlog P1 "Full agent-drivable CLI parity" was resolved by Plan 41 (shipped 2026-06-16) but not cleaned from Backlog. | `Backlog.md` P1 | Commented out with resolution note |
| 4 | medium | Website npm vuln tree P2 item: 6 Dependabot alerts open, astro upgrade breaks build. 56 days unresolved. | `Backlog.md` P2 | Flagged for user — needs dedicated fix pass |
| 5 | medium | All routines idle for ~96 days. Dependency Audit, Vulnerability Scan, Doc Freshness, Memory Consolidation, Status Hygiene, Roadmap Accuracy all overdue by 88-99 days. | `agent/Core/routines.md` | Flagged for user — routine schedule needs catchup session |
| 6 | low | P1 "Routine bot PR pile-up" (added 2026-05-07, 96 days): cloud routine PRs still accumulating without merge strategy. | `Backlog.md` P1 | Flagged — still unresolved, needs process decision |
| 7 | low | Three P3 Backlog items align with current Phase 2 Roadmap milestones and could be promoted: "Custom item creator," "Self-update mechanism," "Micro-task fast path." | `Backlog.md` P3 | Flagged for user consideration |
| 8 | info | 96 days of routine output since last hygiene run. High volume of stale items (Groups A/B/C/D/E/F, 100–120 days old). | `Backlog.md` | Noted — no autonomous action taken |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[URGENT] Rotate HOMEBREW_TAP_TOKEN PAT immediately.** Original PAT rotated 2026-04-22, 90-day expiry means it expired ~2026-07-21. Now 2026-08-11 — 21 days past expiry. Next `goreleaser` run will fail at Homebrew tap step with 401. Go to `LastStep/Bonsai` → Settings → Secrets and rotate the PAT before cutting the next release.

- **Website npm vuln tree (P2):** 6 open Dependabot alerts on `/website` (esbuild HIGH, vite HIGH+MED, js-yaml MED, astro+esbuild LOW). The astro 6.1.7→6.3.2 upgrade PR (#108) fails build after rebase. Needs a dedicated pass to fix the build break + bump vite/js-yaml. This is 56 days old.

- **All routines are overdue.** Dependency Audit (97 days), Vulnerability Scan (99 days), Doc Freshness Check (99 days), Memory Consolidation (96 days), Roadmap Accuracy (96 days), Status Hygiene (96 days). Consider a catchup routine-digest session to process them all.

- **Phase 2 P3 promotion candidates:** Three P3 items align with the current Roadmap Phase 2 (Extensibility) — "Custom item creator," "Self-update mechanism," "Micro-task fast path." User may want to promote one or more to P2/P1 given the project has entered Phase 2.

- **P1 "Routine bot PR pile-up"** (96 days): A process decision is needed — whether to change cloud routines to commit-direct-to-main, auto-merge if green, or skip PR creation when local digest has already absorbed the date range.

## Notes for Next Run

- P0 section is now empty of active items. If a new P0 surfaces before next run, add it directly to the P0 section (the HTML comments are archived-only context).
- The HOMEBREW_TAP_TOKEN PAT situation will recur — consider documenting a recurring calendar reminder or automating PAT expiry tracking.
- After the catchup routine-digest runs, re-check whether new findings need Backlog entries.
- The website npm vuln item is a recurring theme — if not fixed before next hygiene run, consider escalating priority.
