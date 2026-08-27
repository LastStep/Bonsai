---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-27
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
- **Tools Used:** Read (file reads), Edit (backlog cleanup, dashboard update, log append), Write (this report)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s
Scanned the P0 section of Backlog.md. Found **two P0 items already resolved** by shipped releases, not yet cleaned up:

1. `[bug] Sensor hook commands use $PWD-walk-up` — Fixed in v0.4.3 (shipped 2026-05-13, Status.md). Commented out.
2. `[feature] bonsai init / bonsai add need non-interactive flags` — Fixed in v0.4.2 (shipped 2026-05-13, Status.md). Commented out.

No P0 items remain that are absent from Status.md.

### Step 2 — Cross-reference with Status.md
Read Status.md. Cross-referenced all Backlog items against In Progress (none) and Recently Done rows.

**Resolved items cleaned up:**
- P1 `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` — Plan 41 (shipped 2026-06-16) delivered all 5 phases: every mutating command now has a `*Result` headless core + JSONL/exit contract. Commented out.

**URGENT FLAG — HOMEBREW PAT expiry:**
- P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — PAT was rotated 2026-04-22 with a 90-day default expiry. Calendar reminder target was ~2026-07-15. Today is 2026-08-27 — **43 days past the reminder date**. The PAT has almost certainly expired. Symptom if expired: next release's GoReleaser brew step will fail with `401 Bad credentials` (binaries publish, only Homebrew formula update is missed). **User must rotate HOMEBREW_TAP_TOKEN immediately before attempting any release.**

**Status.md Pending check:**
- `[research] Trial sentrux on Bonsai repo` — still in Status.md Pending (blocked on Rust toolchain install). Already commented out of Backlog P0 in prior run (2026-05-07). No action needed.

**Blocked-by opportunities:**
- No Backlog items clearly unblock any current Status.md Pending entries. The sentrux trial is blocked externally (Rust toolchain), not by a Backlog resolution.

### Step 3 — Cross-reference with Roadmap.md
Read Roadmap.md. Phase 1 is fully checked. Phase 2 items cross-referenced:

- `Self-update mechanism` — P3 Backlog ("Self-update mechanism" in Big Bets). Consistent, no promotion needed.
- `Micro-task fast path` — P3 Backlog ("Micro-task fast path" in Future Platform). Consistent.
- `Template variables expansion` — not explicitly in Backlog; still a gap. Low priority (Phase 2 general goal).

No P2/P3 items align strongly enough with current-phase milestones to warrant promotion. Phase 3 (Managed Agents, Greenhouse) items are long-horizon and correctly remain in P3.

No deprecated approaches or completed-phase references found in current P0–P2 items. The P3 "Integration scaffolding variants" references markdown vs GitHub Issues/Notion/Jira — these are future platform items, not referencing anything deprecated.

### Step 4 — Flag stale items
The last run was 2026-05-07 — this run is 112 days later. Many items have sat untouched:

- **P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry`** — Added 2026-04-22, expiry window long past. Flagged as urgent above.
- **P1 `[ops] Routine bot PR pile-up`** — Added 2026-05-07. No evidence of resolution in RoutineLog or Status.md since. Still valid.
- **P1 `[debt] Testing infrastructure for triggers and sensors`** — Added 2026-04-16. No progress visible. Still valid (Group B).
- **P1 `[debt] Stale agent worktrees + branches accumulating`** — Added 2026-04-20/21. No cleanup log entry since. Pattern likely recurring.
- **P2 `[bug] bonsai validate can't pass on Bonsai repo`** — Added 2026-06-13. User decision required (lock-file policy). Stale with no resolution.
- **P2 `[security] Website npm vuln tree`** — Added 2026-06-16. No resolution in Status.md. Astro upgrade build break still unresolved. Stale 73 days.

No near-duplicate pairs found after removing the resolved P1 CLI parity item and P0 non-interactive flags item.

### Step 5 — Check for routine-generated items
Reviewed RoutineLog entries since 2026-05-07. Found:
- 2026-06-13 Plan 40 dispatch notes: "bonsai validate can't pass on Bonsai repo" → already in Backlog P2.
- 2026-06-16 Plan 41 dispatch notes: "unify remove cinematic/headless logic" → already in Backlog P2.
- No routine flagged findings (after 2026-05-07) appear uncaptured in Backlog.

### Step 6 — Promote ready items via issue-to-implementation
No items are pre-approved for immediate implementation. The HOMEBREW PAT item requires user action directly (rotate a secret), not agent implementation. Flagged for user review.

### Steps 7–8 — Log and dashboard update
(Completed below.)

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | P0 bug `Sensor hook commands` already resolved in v0.4.3 (2026-05-13) — stale entry | Backlog.md P0 | Commented out with resolution note |
| 2 | HIGH | P0 feature `bonsai init/add non-interactive flags` already resolved in v0.4.2 (2026-05-13) — stale entry | Backlog.md P0 | Commented out with resolution note |
| 3 | HIGH | P1 feature `Full agent-drivable CLI parity` already resolved in Plan 41 (2026-06-16) — stale entry | Backlog.md P1 | Commented out with resolution note |
| 4 | URGENT | HOMEBREW_TAP_TOKEN PAT expiry reminder was 2026-07-15 — 43 days overdue. PAT likely expired. Next release Homebrew step will fail. | Backlog.md P1 | Flagged for user — rotate PAT immediately |
| 5 | MEDIUM | P2 `bonsai validate` can't pass on Bonsai repo (lock-file gitignored) — 73 days stale, no user decision | Backlog.md P2 | Flagged for user decision |
| 6 | MEDIUM | P2 `Website npm vuln tree` — astro upgrade build break unresolved, 73 days old | Backlog.md P2 | Flagged for user review |
| 7 | LOW | P1 `Routine bot PR pile-up` — filed 2026-05-07, no resolution action visible | Backlog.md P1 | Noted — still valid |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

1. **URGENT — Rotate HOMEBREW_TAP_TOKEN PAT** — The fine-grained PAT rotated 2026-04-22 had a 90-day expiry (target ~2026-07-15). Today is 2026-08-27, 43 days past expiry. Go to GitHub → Settings → Developer settings → Personal access tokens → Fine-grained tokens, rotate, and update the `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai`. Without this, the next release's GoReleaser Homebrew step will fail with `401 Bad credentials`.

2. **Decision required — bonsai validate lock-file policy** — `bonsai validate` cannot pass on the Bonsai repo itself because `.bonsai-lock.yaml` is gitignored (`.gitignore:15`). Two options: (a) commit the lock file + run `bonsai update` once to re-lock, or (b) accept that validate is unusable for self-hosting. This blocks the Plan 40 dogfood validation gate.

3. **Website npm vuln tree** — Astro 6.1.7→6.3.2 bump fails the website build after rebase. Needs a real upgrade pass. 6 open Dependabot alerts. Overdue for vulnerability-scan routine attention.

## Notes for Next Run
- After user rotates HOMEBREW_TAP_TOKEN, update or remove the P1 PAT expiry item from Backlog (note new expiry date if PAT is rotated with a new 90-day window).
- If the `bonsai validate` lock-file policy decision is made, remove the P2 bug item.
- The P3 section (Bonsai-Eval followups, Validate command followups) is stable — no action needed.
- Gap since last run: 112 days (2026-05-07 → 2026-08-27). The routine ran at normal 7-day frequency via cloud dispatch loop but this is the first execution since May.
