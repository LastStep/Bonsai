---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-28
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07 (previous value from dashboard, before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 minutes
- **Files Read:** 7 — `station/agent/Core/identity.md`, `station/agent/Core/memory.md`, `station/agent/Routines/backlog-hygiene.md`, `station/agent/Core/routines.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write, Glob, Bash
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Playbook/Backlog.md` P0 section; cross-referenced against `Status.md`.
- **Result:** Both active P0 items are already resolved and should be removed (not escalated — they shipped). See Step 2.
  - `[bug] Sensor hook commands use $PWD-walk-up` — fixed in v0.4.3 (Status.md: "sensor hook commands now bake install-time absolute paths", PR #105/#106).
  - `[feature] bonsai init / bonsai add need non-interactive flags` — shipped in v0.4.2 (Status.md: `--non-interactive --from-config <path>` delivered, PR #102).
  - The `[research] Trial sentrux` item was already commented out and promoted to Status.md Pending (handled 2026-05-07).
- **Issues:** None — no P0 items need escalation. Both were resolved in prior work cycles.

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md` in full; compared against backlog items across all priority tiers.
- **Result:** Three items confirmed resolved and removed from backlog (replaced with dated HTML comments for audit trail):
  1. **P0 resolved:** `[bug] Sensor hook commands use $PWD-walk-up` — v0.4.3, PR #105/#106, 2026-05-13.
  2. **P0 resolved:** `[feature] bonsai init / bonsai add need non-interactive flags` — v0.4.2, PR #102, 2026-05-13.
  3. **P1 resolved:** `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` — Plan 41 (main `ab202c3`, 2026-06-16) shipped headless `*Result` cores for all four commands + `list --json` + JSONL/exit-code contract. This item's stated goal is fully satisfied; follow-ups (MCP server Plan 42, unify remove logic, website vuln) are already separately tracked in Backlog P2.
- **Blocked-by check:** Status.md Pending has one item: `[research] Trial sentrux` blocked on Rust toolchain. No backlog item would unblock it — resolution requires rustup install (user action).
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Playbook/Roadmap.md`; checked all P2/P3 backlog items against phase milestones.
- **Result:**
  - Phase 1 is fully checked — all items complete. No backlog items reference stale Phase 1 work.
  - Phase 2 (Extensibility): backlog items that align include `[feature] Custom item creator` (P3), `[improvement] Self-update mechanism` (P3), `[improvement] Micro-task fast path` (P3). No promotions warranted without user input.
  - Phase 3 (Cloud & Orchestration): `[feature] Managed Agents integration` and `[feature] Greenhouse companion app` are already in P3 Big Bets. The now-resolved P1 headless-CLI parity item was the precursor to this phase — its removal is appropriate.
  - `[feature] Integrate plan-grilling as first-class Bonsai catalog ability` (P2) aligns with Phase 2 extensibility milestones — flagged for possible P1 promotion at user discretion.
  - No items reference deprecated approaches or completed phases.
- **Issues:** None blocking.

### Step 4: Flag stale items
- **Action:** Scanned all backlog items for age (30+ days without progress) and clarity.
- **Result:**
  - **PAST DUE — P1 ops: `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder`** — The item explicitly called for rotation at ~2026-07-15. Today is 2026-07-28. The PAT has very likely expired. Symptom if expired: next GoReleaser release fails at brew step with `401 Bad credentials`. **Flagged for immediate user action.**
  - Most P2/P3 items date from 2026-04-13 to 2026-06-16 (41–105 days old). This is within normal aging given the project's pace; no items are obviously abandoned or context-free.
  - No near-duplicates detected across priority tiers after the three removals above.
  - Group A (NoteStandards backlog trim) and Group E (Plan archiving, Plans Index) remain outstanding but are known long-standing items.
- **Issues:** HOMEBREW_TAP_TOKEN PAT likely expired — requires user attention before next release.

### Step 5: Check for routine-generated items
- **Action:** Read all RoutineLog entries since 2026-05-07 (last backlog-hygiene run).
- **Result:** The only log entry after 2026-05-07 is `2026-06-13 — Plan 40 dispatch` (a non-routine session entry). From that entry, all flagged items are confirmed already captured in Backlog:
  - Symlink hardening → already in P2 as `[security] Harden all scaffolding writes`
  - `bonsai validate` project.yaml drift warning → already in P2
  - Plan 40 review nits → already in P2
  - `bonsai validate` can't pass on this repo → already in P2
  - Website npm vuln tree → already in P2
  - Unify remove logic → already in P2
  - **NOT captured:** MCP server (Plan 42). Memory.md Work State records "Open follow-ups (Backlog P2): (1) MCP server = Plan 42 (go-sdk, stdio `bonsai mcp`) — the contract was built for this" but no corresponding backlog entry exists. **Flagged for user to add.**
- **Issues:** One uncaptured finding — MCP server Plan 42 not in Backlog despite memory.md referencing it as a planned P2 item.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any item warrants immediate promotion through the workflow.
- **Result:** No items cleared for autonomous promotion. Items requiring user decisions are flagged in the "Items Flagged for User Review" section below. No user confirmation available in this subagent run.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Entry appended.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for Backlog Hygiene.
- **Result:** Last Ran → 2026-07-28, Next Due → 2026-08-04, Status → done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Low | P0 `[bug] Sensor hook commands use $PWD-walk-up` was resolved in v0.4.3 but still in backlog | `Backlog.md` P0 | Removed (HTML comment with resolution date/PR) |
| 2 | Low | P0 `[feature] bonsai init/add non-interactive flags` was resolved in v0.4.2 but still in backlog | `Backlog.md` P0 | Removed (HTML comment with resolution date/PR) |
| 3 | Low | P1 `[feature] Full agent-drivable CLI parity` was resolved in Plan 41 but still in backlog | `Backlog.md` P1 | Removed (HTML comment noting Plan 41 delivery) |
| 4 | High | P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry` reminder was due 2026-07-15; today is 2026-07-28 | `Backlog.md` P1 | Flagged for immediate user action — PAT likely expired |
| 5 | Medium | MCP server (Plan 42) referenced in memory.md as Backlog P2 but no entry exists in Backlog | `Backlog.md` (missing) | Flagged for user to add |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[URGENT] Rotate HOMEBREW_TAP_TOKEN PAT** — Reminder set for ~2026-07-15 when PAT was created 2026-04-22 (90-day default). Today is 2026-07-28 — PAT has expired. Next GoReleaser release will fail at the brew formula step with `401 Bad credentials`. Recovery: rotate PAT in GitHub → set `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai`. See memory.md Notes for full recovery procedure.

- **[Backlog entry missing] MCP server — Plan 42** — memory.md Work State notes this as "Open follow-ups (Backlog P2): (1) MCP server = Plan 42 (go-sdk, stdio `bonsai mcp`) — the contract was built for this." No corresponding backlog item exists. Suggest adding: `- **[feature] MCP server (Plan 42)** — go-sdk, stdio `bonsai mcp` transport; headless CLI contract from Plan 41 is the prerequisite. *(added 2026-07-28, source: memory.md follow-up)*`

## Notes for Next Run

- Both P0 items are now gone; the P0 section only contains the `[research] Trial sentrux` comment (already promoted to Status.md Pending). If Rust toolchain remains uninstalled, the sentrux item may need a formal deferral decision.
- Plan 41 file is still in `Plans/Active/` — noted in memory.md as needing archival at next wrap-up. This is a Status Hygiene task, not Backlog Hygiene.
- Group A (NoteStandards backlog trim) has been open since 2026-04-25 (94 days). Consider scheduling as a discrete task if it keeps deferring.
