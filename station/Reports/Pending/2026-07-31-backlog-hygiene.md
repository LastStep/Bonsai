---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-31
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
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section; cross-referenced each item against `Status.md`.
- **Result:** Both original P0 items are RESOLVED and already shipped — they were never cleaned up from the Backlog after delivery. No active P0 items remain without Status.md coverage.
  - `[bug] Sensor hook commands use $PWD-walk-up` → RESOLVED by v0.4.3 (PR #105/#106, 2026-05-13)
  - `[feature] bonsai init / bonsai add need non-interactive flags` → RESOLVED by v0.4.2/Plan 39 (PR #102, 2026-05-13)
- **Issues:** One NEW P0 was identified during the P1 sweep (see Step 4): HOMEBREW_TAP_TOKEN PAT expired ~2026-07-21 — escalated from P1 to P0.

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md` In Progress, Pending, and Recently Done. Compared all Backlog items against recent deliveries.
- **Result:** 3 items confirmed resolved and removed from Backlog:
  1. P0 `[bug] Sensor hook commands` → v0.4.3 (2026-05-13)
  2. P0 `[feature] bonsai init / bonsai add non-interactive` → v0.4.2/Plan 39 (2026-05-13)
  3. P1 `[feature] Full agent-drivable CLI parity (init/update/add/remove)` → Plan 41 (2026-06-16, PRs #120/#122/#123/#121/#125)
- **Also found:** The sentrux trial is correctly listed in Status.md Pending (blocked on Rust toolchain) — no Backlog action needed.
- **Missing item:** `memory.md` Work State explicitly calls out `Plan 42 (go-sdk, stdio bonsai mcp)` as a Backlog P2 follow-up from Plan 41, but no Backlog entry exists. Flagged for user.
- **Issues:** None beyond the 3 removals and 1 missing item.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md`; scanned Phase 1 (complete), Phase 2, Phase 3, Phase 4 against Backlog entries.
- **Result:**
  - Phase 1: All items checked, matches Status.md. Clean.
  - Phase 2: P3 Backlog items "Self-update mechanism" and "Micro-task fast path" align with Phase 2 goals. No promotion warranted without user direction (no milestone pressure).
  - Phase 3: P3 "Managed Agents integration" and "Greenhouse companion app" are correctly P3 and aligned.
  - Phase 4: P3 "Catalog marketplace / Plugin system" items correctly staged.
  - The Plan 41 delivery unblocks the MCP server (Plan 42) which is a Phase 3 prerequisite — no Roadmap change needed, but a Backlog P2 entry for Plan 42 is missing (flagged for user).
- **Issues:** One missing Backlog P2 item (Plan 42 MCP server) — flagged for user, not auto-added per procedure.

### Step 4: Flag stale items
- **Action:** Checked dates on all Backlog items; identified items at same priority for 30+ days.
- **Result:**
  - **CRITICAL — HOMEBREW_TAP_TOKEN PAT OVERDUE:** P1 item added 2026-04-22 with rotation target ~2026-07-15. Today is 2026-07-31 — PAT expired ~2026-07-21 (10 days ago). Promoted to P0 with updated description and action steps.
  - **Mass staleness (100+ days at same priority):** Virtually all Group B (code quality), Group C (OSS readiness), Group D (catalog expansion), Group E (workspace improvements), and P3 items have been at the same priority since April 2026 (~3.5 months). These are all correctly prioritized — no individual promotions warranted, but the backlog as a whole has not been processed in 3 months. Flagged for user awareness.
  - **`[ops] Routine bot PR pile-up`** (added 2026-05-07, 85 days old, P1): Still relevant. No near-term resolution is obvious without a user decision on commit-direct-to-main vs auto-merge.
  - **`[debt] Stale agent worktrees + branches`** (added 2026-04-20, 102 days old, P1): Status unknown. Could be worse than the April audit count. Flagged.
- **Issues:** 1 critical (PAT escalation, resolved autonomously by promotion to P0).

### Step 5: Check for routine-generated items
- **Action:** Reviewed `RoutineLog.md` for entries since last backlog-hygiene run (2026-05-07).
- **Result:** Only one post-2026-05-07 log entry found: `2026-06-13 — Plan 40 dispatch` (not a routine). No routine runs have been logged since 2026-05-07 — all 7 routines are massively overdue (74–85 days past due). This means no routine-generated findings have been captured in the Backlog during this window. Flagging for user: if the dependency audit, vulnerability scan, or doc freshness check was run informally (outside the routine system), any findings from those runs are not captured.
- **Issues:** 7 routines all severely overdue. No findings to verify or add.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any items are ready for implementation promotion.
- **Result:** The newly-escalated P0 PAT rotation is a direct user action (rotate PAT via `gh secret set`) — not an issue-to-implementation workflow item. No other items are ready for autonomous promotion without user approval.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.

### Step 8: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` Backlog Hygiene row: Last Ran → 2026-07-31, Next Due → 2026-08-07, Status → done.
- **Result:** Done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | **P0 — CRITICAL** | HOMEBREW_TAP_TOKEN PAT expired ~2026-07-21 (10 days overdue) — next release will fail at Homebrew step | `Backlog.md` P1 → P0 | Escalated P1 item to P0 with updated action steps |
| 2 | High | P0 `[bug] Sensor hook commands use $PWD-walk-up` was shipped in v0.4.3 (2026-05-13) but still listed as active | `Backlog.md` P0 | Removed (HTML comment with resolution note) |
| 3 | High | P0 `[feature] bonsai init/add non-interactive flags` was shipped in v0.4.2 (2026-05-13) but still listed as active | `Backlog.md` P0 | Removed (HTML comment with resolution note) |
| 4 | High | P1 `[feature] Full agent-drivable CLI parity` was shipped in Plan 41 (2026-06-16) but still listed as active | `Backlog.md` P1 | Removed (HTML comment with resolution note) |
| 5 | Medium | Plan 42 MCP server explicitly noted in `memory.md` Work State as Backlog P2 follow-up — not in Backlog | `memory.md` Work State | Flagged for user to add |
| 6 | Medium | Plan 41 still in `Plans/Active/` — memory.md says "archive to Plans/Archive/ at next wrap-up" | `Plans/Active/41-headless-cli-contract.md` | Flagged for user |
| 7 | Medium | Plan 40 still in `Plans/Active/` (Phases 1–3 shipped, Phase 4 held) | `Plans/Active/40-odysseus-platform-integration.md` | Flagged for user — partially complete plan |
| 8 | Low | All 7 routines massively overdue (74–85 days) — last run 2026-05-07 | `routines.md` dashboard | Flagged for user — no routine has run in ~3 months |
| 9 | Low | Nearly all P2/P3 and Group items unchanged since April 2026 (3+ months) | `Backlog.md` | No promotions warranted without user direction; noted |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **[URGENT] Rotate HOMEBREW_TAP_TOKEN PAT now.** The PAT expired ~2026-07-21. Any release attempt (for v0.5.0 tag or Plan 42 MCP server release) will fail at the GoReleaser Homebrew step. Action: `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai` with a fresh fine-grained PAT scoped to `LastStep/homebrew-tap`. After rotation, update the P0 Backlog item's rotation date.

2. **Add Plan 42 (MCP server) to Backlog P2.** `memory.md` Work State explicitly says "Open follow-ups (Backlog P2): (1) MCP server = Plan 42 (go-sdk, stdio `bonsai mcp`) — the contract was built for this." No Backlog entry exists. Suggest: `[feature] MCP server (Plan 42) — go-sdk, stdio transport, 'bonsai mcp' command; agent-interface.md contract already defines the surface. *(added 2026-06-16, source: Plan 41 follow-up)*`

3. **Archive Plan 41 (and evaluate Plan 40).** Plan 41 is fully shipped (all 5 phases merged, main `ab202c3`). Memory.md already flagged this. Plan 40: Phases 1–3 merged, Phase 4 held indefinitely (superseded by headless parity workstream, which shipped as Plan 41). Decision needed: archive Plan 40 as "partial — phase 4 superseded" or keep in Active.

4. **Schedule overdue routines.** All 7 routines last ran 2026-05-07 (~85 days ago). Recommended order: (1) Vulnerability Scan (most time-sensitive given website npm vulns already flagged), (2) Dependency Audit, (3) Doc Freshness Check, (4) Status Hygiene, (5) Memory Consolidation, (6) Roadmap Accuracy, (7) Backlog Hygiene (next cycle).

---

## Notes for Next Run

- After PAT rotation, update the P0 item with the new rotation date and reset to P1 (or close if a calendar reminder is set elsewhere).
- Plan 42 (MCP server) should be in the Backlog P2 section by next run.
- Plans 40 and 41 should be in Archive (or at least Plan 41 unconditionally).
- Website npm vulnerability situation (esbuild HIGH, vite HIGH+MED) has been unaddressed since 2026-06-16 — the vulnerability-scan routine is overdue and this should be the first routine run after this one.
- If the bot PR pile-up (P1 item) has resumed generating stale PRs since the 2026-05-07 triage sweep, close them before next routine-digest.
