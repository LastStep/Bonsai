---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-29
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
- **Duration:** ~8 min
- **Files Read:** 5 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 4 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Reports/Pending/2026-08-29-backlog-hygiene.md`
- **Tools Used:** Read, Edit, Write, Grep (file reads and targeted edits)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog P0 section; checked each active P0 against Status.md In Progress and Pending tables.
- **Result:** Both active P0 items are RESOLVED — not missing from Status.md, but already shipped. No unplaced P0s requiring immediate promotion. The sentrux research item was already promoted to Status.md Pending (confirmed as a comment in Backlog).
  - `[bug] Sensor hook commands use $PWD-walk-up` → Fixed in v0.4.3 (Status.md Recently Done 2026-05-13)
  - `[feature] bonsai init/add need non-interactive flags` → Fixed in v0.4.2 (Status.md Recently Done 2026-05-13)
- **Issues:** Both P0s were stale/resolved but still listed as active items. Cleaned up in Step 2.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md; compared all Backlog items against In Progress, Pending, and Recently Done tables.
- **Result:** Found 3 items in Backlog that are fully resolved per Status.md:
  1. P0 `[bug] Sensor hook commands` → v0.4.3 hotfix shipped 2026-05-13 (PR #105/#106)
  2. P0 `[feature] bonsai init/add non-interactive flags` → v0.4.2 shipped 2026-05-13 (PR #102)
  3. P1 `[feature] Full agent-drivable CLI parity` → Plan 41 shipped 2026-06-16 (all 5 phases, PRs #120–#125) — headless *Result cores + JSONL/exit contract for init/add/update/remove + list --json
  
  All three items converted to HTML comments in Backlog.md with resolution notes.
  
  No Pending items with "Blocked By" that could be unblocked by a Backlog item (sentrux remains blocked on Rust toolchain, not a Backlog dependency).
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; tagged P2/P3 backlog items against phase milestones.
- **Result:**
  - Phase 2 milestones vs Backlog:
    - "Self-update mechanism" [x not yet] → P3 backlog item exists (`[improvement] Self-update mechanism`) — no promotion warranted (still Phase 3 territory, not urgent)
    - "Template variables expansion" → no dedicated backlog item (gap noted)
    - "Micro-task fast path" → P3 backlog item exists
  - Phase 3 milestones vs Backlog:
    - "Managed Agents integration" → P3 backlog item exists
    - "Greenhouse companion app" → P3 backlog item exists
  - No deprecated approaches or completed-phase references detected in active P2/P3 items.
  - P2 `[feature] Integrate plan-grilling as first-class catalog ability` aligns with Phase 2 (Extensibility) goals — candidate for P1 promotion when capacity opens.
- **Issues:** None requiring immediate action.

### Step 4: Flag stale items
- **Action:** Checked all items against date added; scanned for near-duplicates; checked for items with no clear rationale.
- **Result:**
  - **CRITICAL STALENESS FLAG:** P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` — added 2026-04-22 with instruction to rotate by ~2026-07-15. Today is 2026-08-29, which is **45 days past the target rotation date.** If the PAT has not been manually rotated, the HOMEBREW_TAP_TOKEN may have expired. A GoReleaser release failure at the brew step would be the first symptom. **Flagged for immediate user review.**
  - **Security staleness:** P2 `[security] Website npm vuln tree — astro upgrade breaks npm run build` — added 2026-06-16 (74 days ago), 6 open Dependabot alerts. Not urgent for Go binary but warrants follow-up.
  - Most P2/P3 items are 90–130+ days old at their current priority level. This is expected for a healthy backlog; none lack rationale or are clearly irrelevant.
  - No near-duplicates detected across priority tiers after review.
- **Issues:** HOMEBREW_TAP_TOKEN likely expired — flagged for user.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07); verified flagged findings are captured.
- **Result:** RoutineLog entries after 2026-05-07 last run: none (no routines ran after 2026-05-07 until this run). Previous flags from 2026-05-07 Backlog Hygiene report:
  - (1) P0 escalation: sentrux — already in Status.md Pending (correct)
  - (2) Roadmap Phase 1 "Better trigger sections" — no separate backlog item; the Roadmap Accuracy report flagged this directly; no action needed in Backlog
  - (3) code-index.md staleness — no new backlog item found; this falls under the Doc Freshness Check routine's domain
  - (4) broken nav link agent/Skills/bonsai-model.md — no separate backlog item; routine finding, not a backlog item
  - (5) INDEX.md arch diagram drift — routine finding, not a backlog item
  
  No uncaptured findings requiring new Backlog items.
- **Issues:** None.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any items qualify for autonomous promotion (P0 needing immediate action, or user-approved items).
- **Result:** No items meet the autonomous promotion criteria. No P0s remain active. No user approval received for promotion. Skipped per procedure.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to station/Logs/RoutineLog.md.
- **Result:** Done.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated station/agent/Core/routines.md — Backlog Hygiene row Last Ran → 2026-08-29, Next Due → 2026-09-05, Status → done.
- **Result:** Done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | HOMEBREW_TAP_TOKEN PAT likely expired — rotation was due 2026-07-15, today is 2026-08-29 (45 days overdue) | P1 ops item in Backlog.md | Flagged for immediate user review |
| 2 | medium | `[security] Website npm vuln tree — astro upgrade` stale 74 days with 6 open Dependabot alerts | P2 security item in Backlog.md | Flagged for user review |
| 3 | info | P0 `[bug] Sensor hook $PWD-walk-up` was resolved in v0.4.3 (2026-05-13) but still listed as active | Backlog.md P0 section | Converted to HTML comment with resolution note |
| 4 | info | P0 `[feature] bonsai init/add non-interactive flags` was resolved in v0.4.2 (2026-05-13) but still listed as active | Backlog.md P0 section | Converted to HTML comment with resolution note |
| 5 | info | P1 `[feature] Full agent-drivable CLI parity` was resolved by Plan 41 (2026-06-16) but still listed as active | Backlog.md P1 section | Converted to HTML comment with resolution note |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

1. **HOMEBREW_TAP_TOKEN PAT — likely expired (HIGH):** The P1 ops item set a reminder to rotate the fine-grained PAT by ~2026-07-15. Today is 2026-08-29, 45 days past that date. If the token has not been rotated manually, the next `bonsai` release will fail at the GoReleaser brew step with `401 Bad credentials`. **Recommended action: rotate the PAT in GitHub → Settings → Developer settings → Fine-grained personal access tokens, then update the `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai`.**

2. **Website npm vulnerabilities — 6 open Dependabot alerts (MEDIUM):** The astro upgrade (PR #108) fails the website build after rebase. 6 alerts open since 2026-06-16. Go binary unaffected. When capacity allows, run a proper astro upgrade-with-build-fix pass rather than auto-merging Dependabot PRs.

## Notes for Next Run

- P0 section is now clean (all three previously active items were resolved). If a new P0 surfaces, it will stand out clearly.
- The gap between routine runs was 114 days (2026-05-07 → 2026-08-29), well past the 7-day cadence. Dashboard shows all other routines also last ran on or before 2026-05-07. The maintenance loop appears to have been inactive for several months. Consider verifying the loop.md dispatch schedule is still active.
- Plan 41 fully resolves the headless CLI story. MCP server (Plan 42 fast-follow) may be in progress — check Status.md next run.
- 3 items removed from active backlog this run. Active P1 count: 4 (down from 5). Active P0 count: 0 (down from 2).
