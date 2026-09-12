---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-12
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
- **Duration:** ~10 min
- **Files Read:** 6 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`, `station/agent/Routines/backlog-hygiene.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read P0 section of Backlog.md, compared each item against Status.md.
- **Result:** Both active P0 items are fully resolved:
  - `[bug] Sensor hook commands use $PWD-walk-up` → shipped as v0.4.3 (2026-05-13). Status.md confirms: "v0.4.3 hotfix shipped — sensor hook commands now bake install-time absolute paths."
  - `[feature] bonsai init/add need non-interactive flags` → shipped as v0.4.2 (2026-05-13). Status.md confirms `--non-interactive --from-config` for init+add.
  - P0 section is now empty of live issues (both removed, resolution comments added).
- **Issues:** None — but note the gap (4 months) between resolution and cleanup is significant.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md (In Progress, Pending, Recently Done). Matched against all Backlog items.
- **Result:**
  - P0 sensor bug + P0 non-interactive flags: removed (resolved, noted above).
  - P1 "Full agent-drivable CLI parity": Plan 41 (merged 2026-06-16) shipped headless `*Result` cores + JSONL/exit contract for all four commands (init/add/update/remove). Item removed from Backlog P1; resolution comment added.
  - Pending "Trial sentrux" correctly cross-referenced (in Status.md Pending, Backlog comment already existed).
  - No Status.md Pending items were found to be unblocked by a Backlog item.
- **Issues:** None after cleanup.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md, scanned P2/P3 items against current and future phases.
- **Result:**
  - Phase 1 is fully checked. No Backlog items reference unfinished Phase 1 work (the two removed P0s were Phase 1 deliverables).
  - Phase 2 (Extensibility) has Backlog candidates: `[feature] Custom item creator` (P3), `[improvement] Self-update mechanism` (P3), `[feature] Micro-task fast path` (P3) — all appropriately at P3 given no active Phase 2 push.
  - No items referencing deprecated or completed phases found that need flagging beyond the removed ones.
  - **MCP server (Plan 42)** is in memory.md Work State as the next planned item but has no Backlog entry. Not auto-adding (per procedure), but flagging for user to confirm if it should be tracked here.
- **Issues:** None structural.

### Step 4: Flag stale items
- **Action:** Reviewed all items for 30+ day stale items (today is 2026-09-12; last run was 2026-05-07 — ~4 months gap).
- **Result:** Key stale findings:
  - **[CRITICAL] P1 HOMEBREW_TAP_TOKEN expiry** — added 2026-04-22 noting ~90-day expiry, reminder set for 2026-07-15. Today is 2026-09-12 — the PAT is ~58 days past the reminder date and very likely expired. Next release attempt will fail at the Homebrew formula step with `401 Bad credentials`. Flagged for immediate user attention.
  - **[HIGH] P2 Website npm vuln tree** — added 2026-06-16, ~3 months with no action. Open Dependabot alerts on /website (esbuild HIGH, vite HIGH+MED, js-yaml MED). The astro upgrade build break has been unresolved since it was filed.
  - **[MEDIUM] P2 bonsai validate lock file policy** — added 2026-06-13, ~3 months. Validate can't pass on the Bonsai dogfood repo itself due to gitignored lock file. Blocks proper Plan 40 dogfood.
  - **[LOW] Group A Bookkeeping** — "Retroactively trim Backlog entries to NoteStandards" added 2026-04-25 (~5 months). Low-effort but repeatedly deferred.
  - **[INFO] Plan 41 archive** — memory.md notes "Plan 41 file still in Plans/Active/ — archive to Plans/Archive/ at next wrap-up." Still not archived.
- **Issues:** HOMEBREW_TAP_TOKEN expiry is the most time-sensitive stale finding.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md for entries since last backlog-hygiene (2026-05-07). No routine log entries exist after 2026-05-07.
- **Result:** No routine runs have occurred since the last backlog-hygiene. All routines show 2026-05-04 or 2026-05-07 as Last Ran in the dashboard — all are severely overdue (4+ months). This is a systemic gap: routines have not been dispatched since May. No new routine-generated Backlog items to verify.
- **Issues:** All 7 routines are overdue; recommend batch routine dispatch after this run.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed if any item is ready for immediate promotion.
- **Result:** No items flagged for promotion at this time — P0 section is now clear. The HOMEBREW_TAP_TOKEN item is operational (fix = PAT rotation, not an implementation plan). Flagging it to the user rather than promoting to issue-to-implementation.
- **Issues:** None.

### Step 7 & 8: Log results and update dashboard
- **Action:** Appended to RoutineLog.md; updated routines.md dashboard row.
- **Result:** Done.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | P0 sensor hook bug was resolved (v0.4.3) but still in Backlog | `Backlog.md P0` | Removed — resolution comment added |
| 2 | high | P0 non-interactive flags were resolved (v0.4.2) but still in Backlog | `Backlog.md P0` | Removed — resolution comment added |
| 3 | high | P1 CLI parity was resolved (Plan 41) but still in Backlog | `Backlog.md P1` | Removed — resolution comment added |
| 4 | high | HOMEBREW_TAP_TOKEN PAT expired ~58 days ago (was due 2026-07-15) | `Backlog.md P1` | Flagged for user — PAT rotation needed before next release |
| 5 | medium | Website npm vuln tree (esbuild/vite HIGH alerts) unaddressed 3+ months | `Backlog.md P2` | Flagged for user — security debt accumulating |
| 6 | medium | bonsai validate can't pass on dogfood repo (lock gitignored) | `Backlog.md P2` | Flagged for user — blocks Plan 40 dogfood |
| 7 | medium | All 7 routines overdue (4+ months gap since last dispatch) | `routines.md dashboard` | Flagged for user — recommend batch routine dispatch |
| 8 | low | Plan 41 file still in Plans/Active/ (should be archived) | `Plans/Active/` | Flagged for user |
| 9 | info | MCP server (Plan 42) has no Backlog entry | `memory.md Work State` | Flagged for user — confirm tracking preference |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[URGENT] Rotate HOMEBREW_TAP_TOKEN** — The PAT was set in 2026-04 with ~90-day expiry; reminder date was 2026-07-15. Today is 2026-09-12 — approximately 58 days overdue. The next `goreleaser release` will fail at the Homebrew formula step. See memory.md note on recovery steps. Backlog item `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` is still in P1.

2. **[HIGH] Website npm vuln tree** — esbuild HIGH and vite HIGH+MED Dependabot alerts have been open since June. The astro upgrade (PR #108) that would clear most has a build break requiring a real fix pass. P2 Backlog item `[security] Website npm vuln tree — astro upgrade breaks npm run build` is ~3 months old.

3. **[MEDIUM] bonsai validate dogfood policy** — P2 item `[bug] bonsai validate can't pass on the Bonsai repo itself` blocks the Plan 40 dogfood. Decide: commit `.bonsai-lock.yaml` + run `bonsai update` to re-lock, or accept validate-unusable-here for now.

4. **[LOW] Plan 41 archive** — `Plans/Active/41-headless-cli-contract.md` should be moved to `Plans/Archive/`. Memory noted this at Plan 41 wrap-up; still in Active.

5. **[LOW] Plan 42 MCP server tracking** — memory.md Work State mentions Plan 42 (go-sdk, stdio `bonsai mcp`) as the next major item. No Backlog entry exists. Confirm if you want it captured here or if it's tracked elsewhere.

6. **[INFO] All routines overdue** — No routines ran since 2026-05-07. Dependency Audit, Doc Freshness Check, Vulnerability Scan, Memory Consolidation, Status Hygiene, and Roadmap Accuracy are all 4+ months past due. Recommend dispatching a batch routine run.

## Notes for Next Run

- The 4-month gap between routine runs means multiple items have accumulated. The next backlog-hygiene should be back on schedule (7 days from 2026-09-12 = 2026-09-19).
- After the HOMEBREW_TAP_TOKEN is rotated, consider updating the expiry reminder with the new date.
- The P0 section is now clean. If new P0s arise, they should be in Status.md Pending/In Progress immediately per Backlog format spec.
- NoteStandards trimming (Group A Bookkeeping) is still pending — low effort, consider bundling with next session's housekeeping.
