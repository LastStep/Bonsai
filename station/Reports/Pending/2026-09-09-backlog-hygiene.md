---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-09
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
- **Files Read:** 6 — `Playbook/Backlog.md`, `Playbook/Status.md`, `Playbook/Roadmap.md`, `Logs/RoutineLog.md`, `agent/Core/routines.md`, `agent/Core/memory.md`
- **Files Modified:** 4 — `Playbook/Backlog.md`, `agent/Core/routines.md`, `Logs/RoutineLog.md`, `Reports/Pending/2026-09-09-backlog-hygiene.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s
Two active P0 items existed in Backlog at run time:
- `[bug] Sensor hook commands use $PWD-walk-up` — resolved by v0.4.3 hotfix (2026-05-13, PRs #105/#106). Present in Status.md Recently Done. **Removed.**
- `[feature] bonsai init / bonsai add need non-interactive flags` — resolved by v0.4.2 (2026-05-13, PR #102). Present in Status.md Recently Done. **Removed.**

P0 section is now empty of active items. No escalation needed.

### Step 2 — Cross-reference with Status.md
Status.md Recently Done reviewed. Key resolutions found:

| Backlog Item | Status.md Match | Action |
|---|---|---|
| P0 sensor hook bug | v0.4.3 hotfix (2026-05-13) | Removed |
| P0 non-interactive flags | v0.4.2 release (2026-05-13) | Removed |
| P1 full agent-drivable CLI parity | Plan 41 SHIPPED (2026-06-16) | Removed |

Pending table reviewed: `[research] Trial sentrux` still blocked on Rust toolchain — no change. No Blocked By dependencies became unblockable by Backlog items.

### Step 3 — Cross-reference with Roadmap.md
- Phase 1: All items confirmed checked. The `bonsai validate` row was added post-2026-05-07 routine-digest as verified in RoutineLog.
- Phase 2: Two items (`Template variables expansion`, `Micro-task fast path`) are unchecked Roadmap milestones. `Micro-task fast path` exists as P3 Backlog. `Template variables expansion` has **no Backlog entry at all** — flagged for user review.
- Phase 3/4: All unchecked items have corresponding P3 Backlog entries. No deprecated references found.

### Step 4 — Flag stale items
Today is 2026-09-09. Last backlog-hygiene ran 2026-05-07 (125 days ago). Items assessed:

**PAT expiry calendar reminder (P1):** The 90-day rotation deadline (2026-07-15) has passed by 55 days. Marked `[STALE]` — user must verify HOMEBREW_TAP_TOKEN was rotated and that GoReleaser brew step still functions.

**Plans/Active/ not archived:** Both Plan 40 and Plan 41 files remain in `Plans/Active/`. Plan 41 is fully shipped; Plan 40 Phase 4 is held. Memory.md explicitly noted "Plan 41 file still in Plans/Active/ — archive to Plans/Archive/ at next wrap-up." Flagged for user action (not in scope of backlog-hygiene to archive).

**Very old items noted (140+ days, no progress):**
- P1 `[debt] Testing infrastructure for triggers and sensors` (2026-04-16, 146 days) — valid debt but no owner, no plan filed
- P1 `[debt] Stale agent worktrees + branches accumulating` (2026-04-20, 142 days) — episodic issue; recurring cleanup more applicable than a single fix
- P2 `[improvement] Plan 40 review nits (non-blocking)` (2026-06-13, 88 days) — minor nits with no plan filed, risk of bit-rot

These are flagged in the "Items Flagged for User Review" section below but not auto-tagged [STALE] since they remain valid backlog items without ambiguous status.

### Step 5 — Check for routine-generated items
RoutineLog entries since 2026-05-07 reviewed. The 2026-06-13 Plan 40 dispatch and Plan 41 dispatch generated Backlog items that are already captured (symlink hardening P2, validate drift check P2, unify-remove P2). One item from memory.md Work State is **not in Backlog**: `[feature] MCP server (Plan 42) — go-sdk, stdio bonsai mcp`. Flagged for user to add manually.

No other uncaptured routine findings. Note: All other routines (Dependency Audit, Vulnerability Scan, Doc Freshness Check, Memory Consolidation, Status Hygiene, Roadmap Accuracy) are severely overdue — last ran 2026-05-04 or 2026-05-07, now 125–128 days ago. No routine-generated Backlog items from that gap exist; those routines need to run.

### Step 6 — Promote ready items (SKIPPED — autonomous mode)
No interactive promotion. Candidates flagged in "Items Flagged for User Review" below.

### Step 7 & 8 — Log + Update dashboard
RoutineLog.md updated. Dashboard `Last Ran` → 2026-09-09, `Next Due` → 2026-09-16, `Status` → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | P0 sensor hook bug already resolved (v0.4.3) | Backlog.md P0 | Removed; replaced with HTML audit comment |
| 2 | High | P0 non-interactive flags already resolved (v0.4.2) | Backlog.md P0 | Removed; replaced with HTML audit comment |
| 3 | High | P1 full CLI parity already resolved (Plan 41) | Backlog.md P1 | Removed; replaced with HTML audit comment |
| 4 | High | PAT expiry deadline (2026-07-15) passed 55 days ago | Backlog.md P1 | Marked [STALE]; user must verify rotation |
| 5 | Medium | `Template variables expansion` Phase 2 roadmap item has no Backlog entry | Roadmap.md Phase 2 | Flagged for user to add |
| 6 | Medium | Plan 41 not archived from Plans/Active/ | Plans/Active/ | Flagged (out of routine scope) |
| 7 | Medium | MCP server (Plan 42) in memory Work State but not in Backlog | memory.md | Flagged for user to add |
| 8 | Low | All other routines overdue by 125–128 days | routines.md | Flagged — dispatch recommended |
| 9 | Low | P1 testing infra (146 days) and worktrees (142 days) long-stale | Backlog.md P1 | Noted; not tagged (status is clear) |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[ACTION REQUIRED] HOMEBREW_TAP_TOKEN PAT rotation:** The 2026-07-15 deadline has passed. If the PAT was not renewed, the next `goreleaser` release will fail at the Homebrew formula update step with a 401. Verify the secret on `LastStep/Bonsai` is current.
- **[ADD TO BACKLOG] MCP server (Plan 42):** Memory Work State references this as a P2 follow-up from Plan 41 ("go-sdk, stdio `bonsai mcp`") but no Backlog item exists for it. Add manually.
- **[ADD TO BACKLOG] `Template variables expansion`:** Phase 2 Roadmap item has no corresponding Backlog entry. Add if still planned.
- **[ARCHIVE] Plan 41 plan file** at `Plans/Active/41-headless-cli-contract.md` — shipped 2026-06-16, move to `Plans/Archive/`. Memory.md already flagged this.
- **[CONSIDER] Run overdue routines:** Dependency Audit, Vulnerability Scan, Doc Freshness Check, Memory Consolidation, Status Hygiene, and Roadmap Accuracy are all 125–128 days overdue. A routine-digest session is recommended.
- **[PROMOTE CANDIDATE] P1 testing infra for triggers/sensors** (146 days old): Longstanding debt. Consider filing a Tier-1 plan or deferring to P3 if no near-term capacity.
- **[PROMOTE CANDIDATE] P2 `bonsai validate` can't pass on Bonsai repo** (88 days): Blocks the dogfood gate established in Plan 40. Consider promoting to P1 and scheduling.

## Notes for Next Run

- P0 section is now clean (all resolved items removed). The only open P0 activity is the sentrux trial in Status.md Pending.
- PAT expiry item updated to [STALE] — user should close it after verifying rotation.
- Two uncaptured items (MCP server, Template variables expansion) should be added by user before next run so they can be tracked properly.
- All other routines severely overdue — coordinate a batch routine-digest session soon.
