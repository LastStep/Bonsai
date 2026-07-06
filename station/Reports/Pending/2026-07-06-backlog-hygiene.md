---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-06
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

---

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s

Scanned the P0 section of Backlog.md. Found **2 P0 items, both already resolved**:

1. `[bug] Sensor hook commands use $PWD-walk-up` — Fixed in v0.4.3 hotfix (Status.md 2026-05-13, PR #105/#106). Sensor hook commands now bake install-time absolute paths into `.claude/settings.json`.
2. `[feature] bonsai init / bonsai add need non-interactive flags` — Fixed in v0.4.2 release (Status.md 2026-05-13). `--non-interactive` + `--from-config` shipped for both commands.

Neither was in Status.md as In Progress or Pending — both appear only in Recently Done as resolved. Both commented out from P0 section.

### Step 2 — Cross-reference with Status.md

- **In Progress:** none (table is empty).
- **Pending:** only `[research] Trial sentrux` — already correctly represented as an HTML comment in the Backlog P0 section; still blocked on Rust toolchain. No change needed.
- **Recently Done cross-check:** Found one P1 item fully resolved:
  - `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` (added 2026-06-13) → resolved by Plan 41 "Headless CLI Contract" (Status.md 2026-06-16). Plan 41 shipped headless cores + JSONL/exit contracts for all four commands (init/add/update/remove) + `list --json` + `docs/agent-interface.md`. Commented out from P1.
- **Blocked-by check:** The only Pending item (sentrux trial) is blocked on Rust toolchain, not on any Backlog item. No Backlog resolutions can unblock it.

### Step 3 — Cross-reference with Roadmap.md

- **Current phase:** Phase 1 is fully complete (all checkboxes `[x]`).
- **Phase 2/3/4 alignment:** P3 items map correctly to future phases — `[improvement] Self-update mechanism` → Phase 2, `[feature] Managed Agents integration` → Phase 3, `[feature] Greenhouse companion app` → Phase 3. No promotions warranted.
- **Deprecated references:** The two removed P0 items referenced versions (v0.4.2, v0.4.3) that have shipped — correctly cleaned up.
- **No Phase 2 milestones warrant P1 promotion** — Phase 2 work has not been prioritized by the user yet.

### Step 4 — Flag stale items

Today is 2026-07-06. 30-day threshold is 2026-06-06. Key findings:

**URGENT — Time-sensitive:**
- P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` — PAT rotated 2026-04-22 with 90-day expiry → expires approximately **2026-07-20** (14 days from now). The item itself set a reminder for **2026-07-15 (9 days away)**. User must rotate the PAT before that date or the next release will fail at the brew step.

**Stale P1 items (60+ days without progress):**
- `[ops] Routine bot PR pile-up` — added 2026-05-07 (60 days). Still relevant if cloud routines resume.
- `[debt] Stale agent worktrees + branches` — added 2026-04-20 (77 days). Accumulation likely worsened since Plan 41 dispatches (5 parallel worktrees).
- `[debt] Testing infrastructure for triggers and sensors` — added 2026-04-16 (81 days). No movement; trigger system expanded significantly since then.

**Near-duplicates:** None found.

**Items without clear context:** None found — all entries have rationale.

### Step 5 — Check for routine-generated items since last run

Last backlog-hygiene ran 2026-05-07. Scanned RoutineLog.md for entries after that date. **No routine log entries exist after 2026-05-07** (60 days of silence). The 2026-06-13 entry is a Plan 40 dispatch, not a routine run.

This means the following routines are **significantly overdue and have not generated new findings to capture**:
- Vulnerability Scan — Next Due 2026-05-11 (56 days overdue)
- Dependency Audit — Next Due 2026-05-11 (56 days overdue)
- Doc Freshness Check — Next Due 2026-05-11 (56 days overdue)
- Memory Consolidation — Next Due 2026-05-12 (55 days overdue)
- Status Hygiene — Next Due 2026-05-12 (55 days overdue)
- Roadmap Accuracy — Next Due 2026-05-21 (46 days overdue)

No uncaptured findings to add to Backlog from routine reports — because no routine reports exist. **Flagged for user review.**

### Step 6 — Promote ready items

No items approved for immediate promotion. The HOMEBREW_TAP_TOKEN expiry requires user action (PAT rotation), not an implementation workflow. Flagged for user attention.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 `[bug] Sensor hook $PWD-walk-up` was already fixed in v0.4.3 | Backlog.md P0 | Commented out |
| 2 | resolved | P0 `[feature] non-interactive flags` was already fixed in v0.4.2 | Backlog.md P0 | Commented out |
| 3 | resolved | P1 `[feature] Full agent-drivable CLI parity` fully resolved by Plan 41 | Backlog.md P1 | Commented out |
| 4 | **urgent** | HOMEBREW_TAP_TOKEN PAT expires ~2026-07-20; reminder date 2026-07-15 is 9 days away | Backlog.md P1 | Flagged for user — no code change needed |
| 5 | **urgent** | 6 overdue routines, none have run since May 2026 (46–56 days overdue) | routines.md dashboard | Flagged for user — routine runs needed |
| 6 | medium | P1 `[ops] Routine bot PR pile-up` — 60 days without progress | Backlog.md P1 | Flagged for re-prioritization |
| 7 | medium | P1 `[debt] Stale agent worktrees + branches` — 77 days without progress | Backlog.md P1 | Flagged for re-prioritization |
| 8 | medium | P1 `[debt] Testing infrastructure for triggers/sensors` — 81 days without progress | Backlog.md P1 | Flagged for re-prioritization |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **URGENT — Rotate HOMEBREW_TAP_TOKEN PAT by 2026-07-15.** The fine-grained PAT rotated 2026-04-22 has a 90-day expiry (~2026-07-20). Symptom if missed: GoReleaser fails at brew step with `401 Bad credentials` — release succeeds but Homebrew formula update is missed. Action: go to GitHub → Settings → Personal Access Tokens → rotate `HOMEBREW_TAP_TOKEN`, update `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`.

2. **URGENT — 6 overdue routines, none have run since 2026-05-07.** Recommend scheduling a routine-digest session to run: Vulnerability Scan, Dependency Audit, Doc Freshness Check, Status Hygiene, Memory Consolidation, and Roadmap Accuracy. A lot has changed since May (Plans 40 and 41 shipped).

3. **P2 website npm vulnerabilities** (`[security] Website npm vuln tree` — added 2026-06-16) — 6 open Dependabot alerts, astro upgrade needed but breaks build. 20 days old and unaddressed.

---

## Notes for Next Run

- P0 section is now clean — both items were resolved production-shipped items that were never cleaned up after the May 2026 releases.
- P1 section has one clean-up (CLI parity resolved by Plan 41). Remaining P1 items are genuinely pending work.
- The 60-day routine gap means next runs of Vulnerability Scan and Dependency Audit should be treated as baseline re-establishes, not incrementals.
- Consider a mini-sweep of stale agent worktrees and branches after the next dispatch cycle — the P1 `[debt] Stale agent worktrees` item is likely worse after Plans 40 and 41's parallel dispatches.
