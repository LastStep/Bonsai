---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-11
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07
- **Gap:** 127 days — routines have not run since early May 2026
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 6 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`, `station/agent/Core/memory.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Glob, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog P0 section and cross-checked each item against Status.md In Progress and Pending.
- **Result:** Both P0 items are RESOLVED (shipped) and should not be in P0:
  - `[bug] Sensor hook commands use $PWD-walk-up` — shipped v0.4.3 (2026-05-13, PR #105/#106). Status.md Recently Done confirms.
  - `[feature] bonsai init/add non-interactive flags` — shipped v0.4.2 (2026-05-13, PR #102). Status.md Recently Done confirms.
  - The commented-out `[research] Trial sentrux` is correctly in Status.md Pending.
- **Issues:** Both P0 items were stale resolved entries left uncleaned since May 2026. Both removed.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md fully; matched Backlog items against In Progress, Pending, and Recently Done.
- **Result:**
  - P0 items 1 and 2: confirmed shipped → removed from Backlog (see Step 1).
  - P1 `[feature] Full agent-drivable CLI parity` — Plan 41 (Status.md Recently Done 2026-06-16) shipped all 4 commands with `*Result` headless cores + JSONL/exit contract. This P1 is fully delivered → removed from Backlog.
  - No Backlog items match the current In Progress table (empty).
  - Status.md Pending `[research] Trial sentrux` — already commented out in Backlog P0 ✓.
  - No Status.md Pending "Blocked By" items could be unblocked by a Backlog item in the current state.
- **Issues:** 1 resolved P1 item removed.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; checked P2/P3 Backlog items for alignment with current phase milestones; checked for deprecated-approach references.
- **Result:**
  - Phase 1 fully shipped ✓ — all checkboxes complete.
  - Phase 2 Extensibility items alignment:
    - "Self-update mechanism" → P3 Backlog has it ✓
    - "Micro-task fast path" → P3 Backlog has it ✓
    - **"Template variables expansion"** → NOT captured in Backlog anywhere — flagged for user.
  - Phase 3/4 items: Managed Agents (P3 ✓), Greenhouse (P3 ✓). Catalog marketplace, plugin system, cross-project coordination not yet in Backlog — appropriate given distance to Phase 4.
  - No items reference deprecated or completed-phase approaches.
- **Issues:** "Template variables expansion" (Phase 2 milestone) uncaptured in Backlog — flagged.

### Step 4: Flag stale items
- **Action:** Reviewed each item's added date vs 2026-09-11. Checked for no-context items and near-duplicates.
- **Result:**
  - **CRITICAL STALE — P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry`**: Added 2026-04-22. PAT set then with 90-day expiry → expired ~2026-07-15. Today is 2026-09-11 — **PAT has been expired for ~58 days**. Any release since July 2026 would fail at the Homebrew formula update step. Immediate user action required.
  - All Group B items (added 2026-04-16): 148 days stale, no Status.md movement. Volume of debt is unchanged — flagged as a category.
  - All Group C, D, E items: 148+ days stale. No progress visible. Still valid work but deeply stale.
  - P1 `[ops] Routine bot PR pile-up` (2026-05-07): 127 days stale, no apparent resolution. Still valid.
  - P1 `[debt] Stale agent worktrees + branches` (2026-04-20): 144 days stale. No cleanup recorded. Still valid; the accumulation has continued.
  - No near-duplicates found (the previously overlapping P0/P1 non-interactive items are now both removed).
- **Issues:** PAT expiry is the top-priority stale flag (operational urgency).

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md for entries since last backlog-hygiene (2026-05-07).
- **Result:** **No routine entries exist after 2026-05-07.** All routines are 127+ days overdue. The dashboard shows Next Due dates of 2026-05-11 to 2026-05-21 — none have run in ~4 months.
  - Five findings from the 2026-05-07 Backlog Hygiene report were flagged: `code-index.md` staleness, broken nav link `agent/Skills/bonsai-model.md`, INDEX.md arch diagram drift. These were NOT captured in the Backlog at the time (audit-only), and no subsequent routine has addressed them. Flagged for user.
  - memory.md Work State references **MCP server = Plan 42** (go-sdk, stdio `bonsai mcp`) as an open follow-up from Plan 41 — NOT in the Backlog. Flagged for user.
  - memory.md Work State references **Plan 41 file still in Plans/Active/** pending archive to Plans/Archive/ — NOT done. Flagged for user.
- **Issues:** No routine-generated Backlog items to verify (no routines ran). 3 items flagged for capture + 1 housekeeping task.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed whether any item is pre-approved or at P0 urgency requiring immediate routing.
- **Result:** No items are pre-approved. The PAT expiry (P1 ops) requires user action (credential rotation), not code implementation. Nothing routed autonomously.
- **Issues:** None — all promotions require user decision.

### Step 7 & 8: Log results and update dashboard
- **Action:** Appended entry to RoutineLog.md; updated routines.md dashboard row for Backlog Hygiene.
- **Result:** Done.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | P0 sensor hook bug item is resolved (shipped v0.4.3) | Backlog.md P0 | Removed — added resolution comment |
| 2 | High | P0 non-interactive flags item is resolved (shipped v0.4.2) | Backlog.md P0 | Removed — added resolution comment |
| 3 | High | P1 full CLI parity item is resolved (shipped Plan 41) | Backlog.md P1 | Removed — added resolution comment |
| 4 | **Critical** | HOMEBREW_TAP_TOKEN PAT expired ~2026-07-15 (~58 days ago) | Backlog.md P1 ops item | Flagged for immediate user action |
| 5 | Medium | All routines overdue by 127+ days — no runs since 2026-05-07 | routines.md dashboard | Flagged for user |
| 6 | Medium | MCP server (Plan 42) in memory Work State but not in Backlog | memory.md | Flagged for user — not auto-added |
| 7 | Medium | Plan 41 file still in Plans/Active/ (should archive) | Plans/Active/ | Flagged for user |
| 8 | Low | "Template variables expansion" (Phase 2 milestone) not in Backlog | Roadmap.md Phase 2 | Flagged for user |
| 9 | Low | code-index.md staleness, broken bonsai-model.md nav link, INDEX.md arch drift uncaptured since 2026-05-07 | RoutineLog 2026-05-07 flags | Flagged for user — not auto-added |
| 10 | Low | Group B/C/D/E items 148+ days stale with no visible progress | Backlog.md | Noted — no action (valid backlog) |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **[IMMEDIATE] HOMEBREW_TAP_TOKEN PAT expired** — Set 2026-04-22 with 90-day expiry. Expired ~2026-07-15. Rotate now at GitHub → Bonsai repo secrets. Any release since July would have silently failed at Homebrew formula publish.

2. **[ACTION] All 7 routines are 127+ days overdue** — Last ran 2026-05-04/07. Dashboard Next Due dates are all in mid-May 2026. Recommend running all overdue routines in a routine-digest session soon (dependency-audit, vulnerability-scan, doc-freshness-check, status-hygiene, roadmap-accuracy, memory-consolidation at minimum).

3. **[CAPTURE] MCP server (Plan 42)** — memory.md Work State references `bonsai mcp` (go-sdk, stdio) as the open follow-up to Plan 41 headless contract. Not yet in Backlog. Confirm priority (P1?) and add entry.

4. **[HOUSEKEEPING] Archive Plan 41 plan file** — `station/Playbook/Plans/Active/41-headless-cli-contract.md` was flagged in memory.md at ship time (2026-06-16) for archiving to `Plans/Archive/`. Still in Active/ 87 days later.

5. **[CAPTURE] "Template variables expansion"** — Phase 2 Roadmap milestone not in Backlog. Add a P3 item if still relevant.

6. **[VERIFY] Stale doc findings from 2026-05-07** — code-index.md staleness, broken `agent/Skills/bonsai-model.md` nav link, INDEX.md arch diagram drift — flagged by last Backlog Hygiene and Doc Freshness Check but never captured in Backlog. Run doc-freshness-check routine to audit current state (may be resolved or worse after 4 months of development).

---

## Notes for Next Run

- The PAT rotation should be verified before the next release attempt.
- Consider running a full routine-digest to process all 7 overdue routines at once.
- After Plan 41 archive + Plan 42 Backlog entry, the P1 section will be leaner and more actionable.
- Group B (code quality + testing) has been at P1 for 148+ days with no movement — consider explicit de-prioritization or scheduling at next session.
