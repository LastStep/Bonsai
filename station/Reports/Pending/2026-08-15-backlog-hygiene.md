---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-15
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
- **Tools Used:** Read (5 files), Edit (3 edits to Backlog.md), Write (report + log entry), Edit (routines.md dashboard)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section. Found two P0 items. Cross-referenced each against Status.md.
- **Result:**
  - `[bug] Sensor hook commands use $PWD-walk-up` — Status.md "Recently Done" confirms v0.4.3 hotfix shipped the fix (PR #105/#106, 2026-05-13). **RESOLVED — removed from active P0.**
  - `[feature] bonsai init / bonsai add need non-interactive flags` — Status.md confirms v0.4.2 shipped `--non-interactive --from-config` (PR #102, 2026-05-13). **RESOLVED — removed from active P0.**
  - P0 section is now empty. Both items replaced with HTML comments preserving the audit trail.
- **Issues:** None. Both items were legitimately resolved; they predated the previous backlog-hygiene run's date but must have been missed during the 2026-05-07 sweep (both were added 2026-05-08/2026-05-13 — after the last run).

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress, Pending, and Recently Done tables. Cross-referenced all active Backlog items.
- **Result:**
  - **In Progress:** none — no conflicts to resolve.
  - **Pending:** `[research] Trial sentrux on Bonsai repo` is in Status.md Pending and already commented out of the Backlog P0 section (comment present from 2026-05-07 run). Clean.
  - **Recently Done — Plan 41 (Headless CLI Contract, 2026-06-16):** Resolves the Backlog P1 item `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove`. Plan 41 shipped headless cores + JSONL/exit contracts for all four mutating commands. **RESOLVED — removed from active P1.**
  - **Blocked By check:** No Status.md Pending items are blocked by resolvable Backlog items. The sentrux item is blocked by Rust toolchain (external dependency), not a Backlog item.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md. Tagged P2/P3 items against current phase milestones. Checked for deprecated references.
- **Result:**
  - **Current phase:** Phase 1 is fully complete (all items checked, including `bonsai validate` and trigger sections from 2026-05-07 digest fix). Phase 2 (Extensibility) is next.
  - **Phase 2 alignment:** Three P3 items align with Phase 2 milestones: `[improvement] Self-update mechanism`, `[improvement] Micro-task fast path`, `[feature] Custom item creator`. None individually strong enough to auto-promote to P1 — no user signal yet. Flagged for user consideration.
  - **Deprecated approaches:** No backlog items found referencing deprecated approaches or completed Phase 1 items that aren't already resolved/commented.
  - **Plan 40 Phase 4 (held):** The `[improvement] Plan 40 review nits (non-blocking)` P2 item references Phase 2/Phase 4 of Plan 40. Phase 4 was HELD. Nits are still valid; their implementation context remains correct for when Phase 4 ships or is superseded. Keep as-is.
- **Issues:** None requiring immediate action.

### Step 4: Flag stale items
- **Action:** Scanned all items for age without progress, unclear context, and near-duplicates.
- **Result:**
  - **HOMEBREW_TAP_TOKEN PAT expiry (P1):** Added 2026-04-22 with a reminder for ~2026-07-15. Today is 2026-08-15 — rotation is 1 month overdue. PAT may already be expired. **Updated item text with urgency flag.** Flagged for user.
  - **`[debt] Testing infrastructure for triggers and sensors` (P1):** Added 2026-04-16 — 4 months old with no progress. Still valid. No plan filed, no Status.md entry. Flag for re-prioritization.
  - **`[debt] Stale agent worktrees + branches accumulating` (P1):** Added 2026-04-20, updated 2026-04-21 — ~4 months old. Still potentially relevant (worktrees can re-accumulate). Keep as-is; low harm if stale.
  - **`[security] Website npm vuln tree — astro upgrade breaks npm run build` (P2):** Added 2026-06-16, ~2 months old. Contains HIGH-severity esbuild advisory. Has open Dependabot alerts. No routine has verified it since. Flag for user.
  - **Near-duplicates check:** `[improvement] Plans Index file` and `[improvement] Plan archiving` are related but distinct (Index = listing file; Archiving = folder structure). The Plans Index item already notes it should fold into Plan archiving. No change needed — they're intentionally co-listed in Group E.
  - **Incidental finding (not backlog scope):** Plans 40 and 41 are in `Plans/Active/` but marked Done in Status.md. Plan 41 is fully complete; Plan 40 has Phase 4 held. Status-hygiene routine should archive these.
- **Issues:** PAT expiry is the most urgent finding.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07). Verified any flagged findings are captured.
- **Result:**
  - **RoutineLog since 2026-05-07:** No routine executions found. The 2026-06-13 entry is a session-plan dispatch log (Plan 40), not a routine run. All other routines are also overdue (last ran 2026-05-04 or 2026-05-07).
  - **Session-sourced findings (Plan 40, 2026-06-13 + Plan 41, 2026-06-16):** Five findings were filed to Backlog during those sessions — all present and accounted for in Backlog P2 (symlink hardening, validate drift warning, review nits, validate lock policy, website npm vulns, unify remove logic).
  - **Uncaptured findings:** None. All session-flagged items were captured at time of filing.
- **Issues:** None — but the complete absence of routine runs since 2026-05-07 means six other routines (dependency-audit, doc-freshness-check, memory-consolidation, roadmap-accuracy, status-hygiene, vulnerability-scan) haven't run in 100 days. This is an operational gap but outside this routine's scope.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** No items are pre-approved for autonomous promotion. No P0 items remain (section is now empty). Flagged items for user decision.
- **Result:** No workflow launched. User review required before any promotions.
- **Issues:** None.

### Step 7 & 8: Log results and update dashboard
- **Action:** Appended entry to RoutineLog.md, updated routines.md dashboard.
- **Result:** Done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | HOMEBREW_TAP_TOKEN PAT rotation overdue by ~1 month (was due 2026-07-15) | Backlog P1 | Updated item with urgency flag; flagged for user |
| 2 | medium | P0 bug (sensor hook $PWD-walk-up) was resolved in v0.4.3 but not cleaned from Backlog | Backlog P0 | Commented out — audit trail preserved |
| 3 | medium | P0 feature (non-interactive flags) was resolved in v0.4.2 but not cleaned from Backlog | Backlog P0 | Commented out — audit trail preserved |
| 4 | medium | P1 feature (full CLI parity) was resolved by Plan 41 (2026-06-16) but not cleaned | Backlog P1 | Commented out — audit trail preserved |
| 5 | medium | Website npm vulns (esbuild HIGH) open since 2026-06-16 (~60 days), PR #108 build still broken | Backlog P2 | Flagged for user — needs vulnerability-scan routine |
| 6 | low | Plans 40 + 41 still in Plans/Active/ despite being Done in Status.md | Plans/Active/ | Flagged for user — status-hygiene routine should archive |
| 7 | low | `[debt] Testing infrastructure for triggers and sensors` at P1 for 4 months with no progress | Backlog P1 | Flagged for re-prioritization |
| 8 | info | P3 items (self-update, micro-task fast path, custom item creator) align with Phase 2 roadmap | Backlog P3 | No action — flagged for user consideration at next phase boundary |
| 9 | info | No routine runs at all since 2026-05-07 — all six other routines are 100 days overdue | All routines | Out of scope — flagged for user awareness |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **HOMEBREW_TAP_TOKEN PAT rotation (URGENT):** The 90-day PAT rotation calendar date (2026-07-15) has passed. Rotate the fine-grained PAT at `github.com/settings/personal-access-tokens` and update the secret via `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`. Otherwise the next release's Homebrew formula step will fail with `401 Bad credentials`. (`station/Playbook/Backlog.md` P1)

- **Website npm vuln tree (P2):** esbuild HIGH advisory + 5 other Dependabot alerts open since 2026-06-16. PR #108 (astro 6.1.7→6.3.2) fails the website build after rebase and needs a manual upgrade pass. Consider scheduling the vulnerability-scan routine to assess current state. (`station/Playbook/Backlog.md` P2)

- **Plans/Active/ archiving:** Plans 40 and 41 are still in `Plans/Active/` but both are Done per Status.md (Plan 40 Phases 1-3 merged / Phase 4 held; Plan 41 fully merged). Status-hygiene routine should move them to `Plans/Archive/`.

- **Phase 2 roadmap items for consideration:** Three P3 backlog items (`[improvement] Self-update mechanism`, `[improvement] Micro-task fast path`, `[feature] Custom item creator`) align with the current Phase 2 milestones. Consider promoting to P2 at the next session if Phase 2 work is being scoped.

- **All other routines 100 days overdue:** dependency-audit, doc-freshness-check, memory-consolidation, roadmap-accuracy, status-hygiene, and vulnerability-scan have not run since early May 2026. Running a routine-digest after this backlog-hygiene report would be the recommended next step.

## Notes for Next Run

- P0 section is now clean (empty). If it stays empty, consider whether the section should be retained as a placeholder or collapsed.
- The HOMEBREW_TAP_TOKEN PAT item should be either resolved (with a resolution comment) or escalated to P0 before the next release.
- The website npm vuln situation (Backlog P2) is aging — if still unresolved at next run, consider escalating to P1.
- The 100-day routine gap means other routines (dependency-audit, vulnerability-scan, etc.) will have significant catch-up findings on their next run. Be prepared for a heavy routine-digest session.
