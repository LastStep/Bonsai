---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-23
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
- **Files Read:** 5 — `station/agent/Routines/backlog-hygiene.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s

Scanned P0 section. Found 2 items:

1. `[bug] Sensor hook commands use $PWD-walk-up` — Cross-referenced against Status.md. Found "v0.4.3 hotfix shipped — sensor hook commands now bake install-time absolute paths" in Recently Done (2026-05-13). **Resolved. Removed.**

2. `[feature] bonsai init / bonsai add need non-interactive flags [Plan 38 P2 blocker]` — Cross-referenced against Status.md. Found "v0.4.2 release shipped — `bonsai init`/`add` `--non-interactive --from-config <path>` (JSONL stdout, hard-skip conflicts, exit codes 0/2/3/4)" in Recently Done (2026-05-13). **Resolved. Removed.**

P0 section now empty of actionable items. No P0s present that need escalation.

### Step 2 — Cross-reference with Status.md

**In Progress:** empty — no overlap to clean.

**Pending:** `[research] Trial sentrux on Bonsai repo` (blocked on Rust toolchain). Already commented out of Backlog from prior run. No action needed.

**Recently Done cross-check:** Identified one P1 item fully resolved by a recently-shipped plan:
- `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` — Plan 41 (2026-06-16) shipped headless `*Result` cores + JSONL/exit contract for all four mutating commands (init/add/update/remove) plus `list --json` and a `docs/agent-interface.md` contract doc. **Resolved. Removed.**

**Blocked-by check:** Pending "Trial sentrux" is blocked by Rust toolchain, not by any Backlog item. No unblocking candidates.

### Step 3 — Cross-reference with Roadmap.md

Phase 1 is fully complete (all checkboxes marked). Phase 2 (Extensibility) is the active next phase.

- `Self-update mechanism` (Phase 2 unchecked) — matches P3 Backlog item. Could promote to P2 as Phase 2 begins, but project has not formally entered Phase 2 yet. No change this cycle.
- `Micro-task fast path` (Phase 2 unchecked) — matches P3 Backlog item. Same reasoning, no change.
- No P2/P3 Backlog items reference deprecated approaches or completed phases beyond those already removed.

No items promoted. Roadmap alignment healthy.

### Step 4 — Flag stale items

Note: All routines appear to have last run in May 2026 (4+ months ago). Many Backlog items have not been touched in 3-5 months.

**High-urgency stale flag:**
- `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` (P1) — PAT rotated 2026-04-22, reminder set for ~2026-07-15. Today is 2026-09-23. The PAT is ~2 months past its rotation reminder date and is likely expired. **Flagged for immediate user action.**

**Other stale P1 items (30+ days, no progress):**
- `[debt] Testing infrastructure for triggers and sensors` — added 2026-04-16, ~5 months stale.
- `[debt] Stale agent worktrees + branches accumulating` — added 2026-04-20/21, ~5 months stale.
- `[ops] Routine bot PR pile-up` — added 2026-05-07, ~4.5 months stale.

No items with truly unclear context or missing rationale. No near-duplicates introduced since last run.

### Step 5 — Check for routine-generated items

RoutineLog reviewed since 2026-05-07 (last backlog-hygiene run). Only entry found after that date is the 2026-06-13 Plan 40 dispatch (not a routine run). No routine findings between 2026-05-07 and today — all routines are significantly overdue. No uncaptured routine findings to verify.

### Step 6 — Promote ready items via issue-to-implementation

No items approved for immediate implementation. HOMEBREW_TAP_TOKEN rotation requires user action (PAT rotation in GitHub secrets), not an agent code change. No autonomous dispatch triggered.

### Step 7 — Log results

Appended entry to `station/Logs/RoutineLog.md`.

### Step 8 — Update dashboard

Updated `agent/Core/routines.md` Backlog Hygiene row: Last Ran → 2026-09-23, Next Due → 2026-09-30.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 bug `$PWD-walk-up` still in Backlog despite v0.4.3 fix | Backlog.md P0 | Replaced with HTML comment noting resolution |
| 2 | resolved | P0 feature `non-interactive flags` still in Backlog despite v0.4.2 fix | Backlog.md P0 | Replaced with HTML comment noting resolution |
| 3 | resolved | P1 feature `full CLI parity` still in Backlog despite Plan 41 ship | Backlog.md P1 | Replaced with HTML comment noting resolution |
| 4 | high | HOMEBREW_TAP_TOKEN PAT ~2 months past rotation reminder date (~2026-07-15) | Backlog.md P1 | Flagged for user — requires manual PAT rotation in GitHub secrets |
| 5 | low | 3 P1 items stale 3-5 months (testing infra, worktrees, bot PR pile-up) | Backlog.md P1 | No change — items still valid, flagged for awareness |
| 6 | info | All routines overdue by 4+ months (last run 2026-05-04/07) | routines.md | No change — outside scope of this routine |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **URGENT: Rotate HOMEBREW_TAP_TOKEN PAT immediately.** The PAT was rotated 2026-04-22 (90-day expiry), reminder was set for ~2026-07-15. It is now 2026-09-23 — approximately 2 months past the rotation window. An expired PAT will cause GoReleaser to fail the Homebrew formula update step at the next release (symptom: `401 Bad credentials` from the brew step; binaries still publish but formula update is missed). Action: rotate the fine-grained PAT in GitHub account settings, then update the `HOMEBREW_TAP_TOKEN` secret in `LastStep/Bonsai` repo settings.

- **Awareness: all routines are 4+ months overdue.** The routine dashboard shows all 7 routines last ran in May 2026. Consider running the full routine suite (Dependency Audit, Vulnerability Scan, Doc Freshness Check, Status Hygiene, Memory Consolidation, Roadmap Accuracy) at the next session.

## Notes for Next Run

- P0 section is now empty. Both P0 items removed; if a new P0 surfaces, escalate immediately.
- Full CLI parity (Plan 41) is done; the non-interactive surface is complete.
- HOMEBREW_TAP_TOKEN rotation is the most urgent user action item.
- Consider running all other overdue routines before next backlog-hygiene — findings from Dependency Audit / Vulnerability Scan may surface new Backlog items.
