---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-18
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
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; identified 2 P0 items. Checked each against Status.md for In Progress or Pending placement.
- **Result:** Both P0 items are in Status.md **Recently Done** — they are resolved, not missing:
  - `[bug] Sensor hook commands use $PWD-walk-up` → resolved by v0.4.3 hotfix (shipped 2026-05-13)
  - `[feature] bonsai init / bonsai add need non-interactive flags` → resolved by v0.4.2 (shipped 2026-05-13) and superseded by Plan 41 (shipped 2026-06-16)
  - No unescalated P0 items found (all resolved).
- **Issues:** None — P0 section is healthy after cleanup.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress, Pending, and Recently Done tables. Cross-referenced against all Backlog items.
- **Result:**
  - **3 items removed** from Backlog (matched Recently Done in Status.md):
    1. P0 `[bug] Sensor hook commands use $PWD-walk-up` — v0.4.3 hotfix in Recently Done
    2. P0 `[feature] bonsai init / bonsai add need non-interactive flags` — v0.4.2 + Plan 41 in Recently Done
    3. P1 `[feature] Full agent-drivable (non-interactive) CLI parity` — Plan 41 headless CLI contract in Recently Done
  - **Pending cross-check:** `[research] Trial sentrux on Bonsai repo` is already in Status.md Pending (blocked on Rust toolchain). No Backlog items directly unblock it — Rust toolchain install is external.
  - No Backlog items conflict with In Progress items (none currently in progress).
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md. Checked P2/P3 Backlog items against current phase milestones.
- **Result:**
  - Phase 1 is fully complete (all boxes checked). No Backlog items reference deprecated Phase 1 approaches.
  - Phase 2 alignment confirmed for: `[feature] Custom item creator` (P3), `[improvement] Self-update mechanism` (P3), `[improvement] Micro-task fast path` (P3) — all correctly in P3 matching Phase 2's future status.
  - `[feature] Port statusLine to catalog sensor` (P2 Group E) aligns with Phase 2 catalog extensibility — no promotion needed, correct priority.
  - Phase 3 items (`Managed Agents integration`, `Greenhouse companion app`) correctly mapped to Big Bets P3.
  - No items reference completed phases incorrectly or deprecated approaches.
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Reviewed all items for 30+ day staleness (items unchanged since before 2026-08-19).
- **Result:** Several items are 90–150+ days old with no progress noted. Key stale flags:
  1. **P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry`** — added 2026-04-22, reminder date was ~2026-07-15. That date has now PASSED (we are in September 2026). PAT may be expired or expiring soon. **Action required.**
  2. **P1 `[ops] Routine bot PR pile-up`** — added 2026-05-07, ~4 months old. Still relevant (cloud cron behaviour unchanged).
  3. **P1 `[debt] Testing infrastructure for triggers and sensors`** — added 2026-04-16, ~5 months old with no progress.
  4. **P1 `[debt] Stale agent worktrees + branches accumulating`** — added 2026-04-20, ~5 months old. The specific worktrees/branches from that time may or may not still exist.
  5. **Sentrux trial** (Status.md Pending) — blocked 4+ months by Rust toolchain. Not a Backlog item per se, but flagging stagnation.
  - P2 items (2026-06-13/16) are ~3 months old — not yet critically stale, no action needed yet.
  - No near-duplicates found that weren't already noted in the Backlog.
- **Issues:** HOMEBREW_TAP_TOKEN PAT reminder date has passed — flagged for user.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07).
- **Result:** Entries after 2026-05-07 in the log:
  - 2026-06-13: Plan 40 dispatch — Backlog items filed: Plan 40 review nits, validate dogfood blocker, symlink hardening, identity drift warning, plan-grilling integration. All are present in Backlog P2. No uncaptured findings.
  - No dependency-audit, vulnerability-scan, or doc-freshness routines have run since 2026-05-07 — these are now 136+ days overdue. **This is itself a finding** worth noting (the routine log shows a large gap).
- **Issues:** Several sibling routines appear to be significantly overdue (last run 2026-05-04 or 2026-05-07). No specific backlog items need to be added from routine output — but the user should be aware other routines need running.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed whether any item warrants promotion + issue-to-implementation routing.
- **Result:** The P0 section is now empty (all resolved and removed). No items were explicitly marked ready for implementation by the user. The HOMEBREW_TAP_TOKEN flag (Step 4) warrants user attention but is an ops task, not a code change. Presenting for user decision rather than auto-routing.
- **Issues:** None requiring autonomous routing.

### Step 7 & 8: Log and dashboard
- **Action:** Updated RoutineLog.md and routines.md dashboard.
- **Result:** Both updated successfully.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 sensor hook commands bug was already shipped as v0.4.3 | Backlog P0 | Removed from Backlog, replaced with HTML comment |
| 2 | resolved | P0 non-interactive flags feature was already shipped as v0.4.2 + Plan 41 | Backlog P0 | Removed from Backlog, replaced with HTML comment |
| 3 | resolved | P1 full agent-drivable CLI parity was shipped by Plan 41 | Backlog P1 | Removed from Backlog, replaced with HTML comment |
| 4 | high | HOMEBREW_TAP_TOKEN reminder date (~2026-07-15) has passed — PAT may need rotation | Backlog P1 | Flagged for user |
| 5 | medium | P1 stale items (testing infra 5mo, worktrees 5mo, bot PR pile-up 4mo) | Backlog P1 | Flagged for user |
| 6 | medium | Sentrux trial blocked 4+ months in Status.md Pending | Status.md Pending | Flagged for user |
| 7 | low | Multiple sibling routines overdue since 2026-05-04/07 (136+ days) | RoutineLog.md | Flagged for user |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

1. **HOMEBREW_TAP_TOKEN PAT rotation** — The reminder date of ~2026-07-15 has passed (2+ months ago). Check whether the PAT has already expired (symptom: GoReleaser brew step fails with 401 at next release). Rotate via GitHub → Settings → Developer settings → Personal access tokens, then `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`. Set a new calendar reminder ~90 days out.

2. **Sentrux trial (Status.md Pending)** — Has been blocked by "Rust toolchain not installed" since 2026-05-07 (4+ months). Decision needed: install Rust toolchain and run the trial, or close this item (demote to Backlog P3 or drop it). The P0 removal no longer makes this a P0-escalated item; it lives in Status.md Pending.

3. **P1 stale items** — Three P1 items have been dormant for 4–5 months:
   - `[debt] Testing infrastructure for triggers and sensors` — no test infra yet; backlog hygiene can't build it, needs a plan.
   - `[debt] Stale agent worktrees + branches accumulating` — may be partially self-resolved if worktrees were cleaned manually; worth a quick `git worktree list` check.
   - `[ops] Routine bot PR pile-up` — bot-PR accumulation is a recurring pattern; a structural fix (one of the 3 options in the Backlog entry) would be worth prioritizing.

4. **Sibling routines overdue** — Dependency Audit (last: 2026-05-04), Vulnerability Scan (2026-05-04), Doc Freshness Check (2026-05-04), Memory Consolidation (2026-05-07), Status Hygiene (2026-05-07), Roadmap Accuracy (2026-05-07) are all 136–148 days overdue. Consider running a routine digest after backlog hygiene.

## Notes for Next Run

- P0 section is now empty — verify it stays clean. Any new P0 added here should be escalated immediately to Status.md.
- HOMEBREW_TAP_TOKEN PAT rotation flag: if resolved before next run, remove the P1 entry from Backlog.
- The 3-item P1 stale cluster (testing infra, worktrees, bot PRs) is now highlighted — if still present in 7 days, consider demoting testing infra to P2 or converting to a formal plan.
- Sibling routines running again will produce fresh flags — run routine-digest after the backlog-hygiene report is absorbed.
