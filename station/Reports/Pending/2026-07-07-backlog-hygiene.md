---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-07
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
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Tools Used:** Read, Edit, Write, Bash
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
Scanned the P0 section of Backlog.md. Found 2 active P0 items (plus 1 HTML-commented resolved item for Trial sentrux). Cross-referenced each against Status.md:

- **P0 `[bug] Sensor hook commands use $PWD-walk-up`** — Status.md Recently Done confirms "v0.4.3 hotfix shipped — sensor hook commands now bake install-time absolute paths." **RESOLVED.** Commented out.
- **P0 `[feature] bonsai init / bonsai add need non-interactive flags`** — Status.md Recently Done confirms "v0.4.2 release shipped — `--non-interactive --from-config <path>` for init/add." **RESOLVED.** Commented out.

No remaining live P0 items in Backlog after cleanup. The Trial sentrux P0 is correctly placed in Status.md Pending (blocked on Rust toolchain).

### Step 2: Cross-reference with Status.md
Read Status.md In Progress (empty), Pending (Trial sentrux — handled), and Recently Done.

Key cross-reference finding: **P1 `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove`** — Plan 41 (PRs #120/#122/#123/#121/#125, main `ab202c3`) delivered exactly this: all four mutating commands have pure `*Result` headless cores, JSONL/exit contract (`ExitConflict=5`), `list --json`, and `docs/agent-interface.md`. **RESOLVED.** Commented out.

Checked Status.md Pending for "Blocked By" items that a Backlog item could unblock: Trial sentrux is blocked by Rust toolchain (`cargo`/`rustc` not installed) — no Backlog item addresses this.

### Step 3: Cross-reference with Roadmap.md
Phase 1 is fully complete (all checkboxes checked). Phase 2 (Extensibility) is the current target phase.

Phase 2 milestone alignment scan:
- "Self-update mechanism" (Phase 2) — Backlog P3 `[improvement] Self-update mechanism` present. Currently P3; Phase 2 focus could warrant P2 promotion but no action taken (presenting to user for decision).
- "Micro-task fast path" (Phase 2) — Backlog P3 `[improvement] Micro-task fast path` present. Same consideration.
- "Template variables expansion" (Phase 2) — No direct Backlog entry. Gap noted.
- "Custom item creator" P3 Backlog — aligns well with Phase 2 Extensibility.

No items found referencing deprecated approaches or completed Phase 1 milestones (beyond the already-resolved items cleaned up in Steps 1/2).

### Step 4: Flag stale items
Today is 2026-07-07; 30-day threshold is 2026-06-07. Most P1/P2 items were added in April 2026 (~80 days ago). The majority are long-standing known items with valid rationale — staleness is expected for P2/P3 given project velocity.

**CRITICAL STALENESS FLAG:** `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` (added 2026-04-22, P1) — The item explicitly notes "set calendar reminder for ~2026-07-15 to rotate before next release." **Today is 2026-07-07 — the PAT expires in approximately 8 days.** If not rotated before the next release, GoReleaser will fail at the Homebrew tap step with `401 Bad credentials`. This needs user action NOW.

Other stale items (30+ days, no progress) that are correctly parked at their priorities and do not need escalation:
- Group B testing debt items (added 2026-04-16, ~82 days) — still accurate P1 debt
- Group C/D/E improvement items — correctly P2/P3
- `[ops] Routine bot PR pile-up` (added 2026-05-07, 61 days) — system issue still unresolved; valid P1

Near-duplicate check: `[improvement] Plans Index file` (Group E) and `[improvement] Plan archiving — Active/Archive folder structure` (Group E) are related and note themselves as such. No action needed — they can be combined when picked up.

Items with no clear context/rationale: none found.

### Step 5: Check for routine-generated items
Read RoutineLog.md entries since last backlog-hygiene (2026-05-07). No routine runs were logged after 2026-05-07 (the 2026-06-13 Plan 40 dispatch entry is a plan execution log, not a routine). All seven routines are massively overdue (44–63 days past their Next Due dates per the dashboard). No uncaptured routine findings to verify. No new Backlog items required from this step.

### Step 6: Promote ready items via issue-to-implementation
No items meet the auto-promote criteria (no user approval on record, no P0 requiring immediate action after cleanup). The PAT expiry flag requires user attention before any implementation route is chosen. Presenting to user in Flags section.

### Steps 7 & 8: Log + Dashboard
Completed as part of report finalization (RoutineLog.md appended, routines.md dashboard updated).

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 sensor hook $PWD bug — resolved v0.4.3 | Backlog P0 | Commented out item |
| 2 | resolved | P0 non-interactive flags — resolved v0.4.2 | Backlog P0 | Commented out item |
| 3 | resolved | P1 non-interactive CLI parity — resolved Plan 41 | Backlog P1 | Commented out item |
| 4 | critical | HOMEBREW_TAP_TOKEN PAT expires ~2026-07-15 (8 days) | Backlog P1 `[ops]` | Flagged for user — no code change |
| 5 | info | All 7 routines overdue 44–63 days since 2026-05-07 | routines.md dashboard | Flagged for user |
| 6 | info | Phase 2 milestone items (self-update, micro-task fast path) at P3 — possible promotion candidates | Backlog P3 | Flagged for user — no change |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

### FLAG 1 — CRITICAL: HOMEBREW_TAP_TOKEN PAT Expiry (8 days)
The `HOMEBREW_TAP_TOKEN` fine-grained PAT was rotated 2026-04-22 and defaults to 90-day expiry. The P1 Backlog item `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` explicitly flagged ~2026-07-15 as the rotation deadline. **Today is 2026-07-07 — 8 days remain.** Symptom of missed rotation: GoReleaser fails at brew step with `401 Bad credentials` on next release (binaries publish fine, only Homebrew formula update is missed). Action: rotate the PAT at `https://github.com/settings/personal-access-tokens` and update via `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`.

### FLAG 2 — INFO: All Routines Massively Overdue
No routines have run since 2026-05-07 (62 days ago). Dashboard shows all 7 routines overdue by 44–63 days. Recommend a routine-digest session to batch-run the most critical ones (dependency-audit and vulnerability-scan first, then others).

### FLAG 3 — INFO: Phase 2 Extensibility Alignment
Three Backlog P3 items align with the active Phase 2 Extensibility milestone: `[improvement] Self-update mechanism`, `[improvement] Micro-task fast path`, and `[feature] Custom item creator`. Consider promoting any of these to P2 when reviewing the roadmap. No change made — presenting for user decision.

## Notes for Next Run
- P0 section is now clean (2 resolved items commented out + 1 previously-commented). The only active P0-equivalent is Trial sentrux in Status.md Pending.
- P1 section had one resolution (Plan 41 non-interactive parity). Remaining P1 items are genuine open work.
- PAT rotation should clear the P1 `[ops] HOMEBREW_TAP_TOKEN` item — remove it from Backlog once rotated.
- If routines run before next backlog-hygiene, their reports may surface new Backlog items to capture.
- RoutineLog has a large gap (2026-05-07 → 2026-07-07); next run should see recent routine entries.
