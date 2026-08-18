---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-18
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

> Note: This run is 103 days overdue. All seven routines have last-ran dates in early May 2026 with no subsequent entries in RoutineLog.md, indicating a gap in routine execution since 2026-05-07.

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 5 — `Backlog.md`, `Status.md`, `Roadmap.md`, `Logs/RoutineLog.md`, `agent/Core/routines.md`
- **Files Modified:** 2 — `Backlog.md` (3 resolved items removed), `agent/Core/routines.md` (dashboard updated)
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s

Backlog had two P0 items. Both are fully resolved and were still in the P0 section:

1. **`[bug] Sensor hook commands use $PWD-walk-up`** — Shipped in v0.4.3 (PR #105/#106, 2026-05-13). Status.md "Recently Done" confirms: "sensor hook commands now bake install-time absolute paths." Removed from backlog.

2. **`[feature] bonsai init / bonsai add need non-interactive flags`** — Shipped in v0.4.2 (PR #102, 2026-05-13). Status.md confirms: "`--non-interactive --from-config <path>` (JSONL stdout, hard-skip conflicts, exit codes 0/2/3/4)." Removed from backlog.

The only remaining P0 content is a comment: `[research] Trial sentrux on Bonsai repo` — correctly promoted to Status.md Pending on 2026-05-07 (blocked on Rust toolchain). This is properly placed; no escalation needed.

### Step 2 — Cross-reference with Status.md

**In Progress:** None.
**Pending:** `[research] Trial sentrux` — blocked on Rust toolchain install. Already captured. No change.
**Recently Done (since last run, 2026-05-07 onward):**
- v0.4.3 hotfix (2026-05-13) — clears P0 bug (handled above)
- v0.4.2 release (2026-05-13) — clears P0 feature (handled above)
- Plan 38 handoff (2026-05-13) — no direct backlog impact
- PR triage sweep (2026-05-07) — bot PR pile-up follow-up filed (still in P1 backlog, correctly placed)
- First external contribution merged (2026-05-07) — no backlog impact
- Plan 40 Phases 1–3 (2026-06-13) — debts/nits filed in P2 (already in backlog, correct)
- Plan 41 Headless CLI (2026-06-16) — "unify remove cinematic/headless logic" filed in P2 (correct); also resolves P1 "Full agent-drivable CLI parity" (handled below)

### Step 3 — Cross-reference with Roadmap.md

Phase 1 is complete (all milestones checked). Phase 2 items:
- `Self-update mechanism` — in P3 Backlog (appropriate)
- `Micro-task fast path` — in P3 Backlog (appropriate)
- `Template variables expansion` — not explicitly tracked in backlog (minor gap, P3 range)

Plan 41 delivered headless cores across all mutating commands — this directly enables Phase 3 "Managed Agents integration" and "Cloud & Orchestration." The P1 "Full agent-drivable CLI parity" item was added 2026-06-13 as a prerequisite; Plan 41 shipped 2026-06-16 (3 days later) completing it. Removed from P1.

No P2/P3 items flagged as aligning with deprecated approaches. Phase 2 transition appears to be in progress organically.

### Step 4 — Flag stale items

Last run was 2026-05-07 (103 days ago). Items stale at 30+ days:

| Item | Added | Age | Severity |
|------|-------|-----|----------|
| `[ops] HOMEBREW_TAP_TOKEN PAT expiry` | 2026-04-22 | 118 days | **URGENT** — reminder was 2026-07-15; PAT likely expired |
| `[security] Website npm vuln tree` | 2026-06-16 | 63 days | HIGH — 6 open vulns including esbuild HIGH, vite HIGH+MED |
| `[debt] Testing infrastructure for triggers/sensors` | 2026-04-16 | 124 days | Medium — stale P1 |
| `[debt] Stale agent worktrees + branches` | 2026-04-20 | 120 days | Medium — worktrees may have grown further |
| `[ops] Routine bot PR pile-up` | 2026-05-07 | 103 days | Medium — fix action still pending |
| Group A bookkeeping (trim to NoteStandards) | 2026-04-25 | 115 days | Low — housekeeping |
| Group B items (generate.go split, catalog/cmd coverage) | 2026-04-16 | 124 days | Low — ongoing debt |

### Step 5 — Check for routine-generated items

RoutineLog.md shows no routine entries after 2026-05-07. All seven routines (Dependency Audit, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene, Vulnerability Scan, and this Backlog Hygiene) are massively overdue (103–118 days). No new routine-generated findings to capture since there were no recent runs. The 2026-05-07 findings were already captured in the backlog from prior digests.

### Steps 6–8 — No promotions; log and dashboard updates follow

No items approved for promotion via issue-to-implementation — that requires user decision. Flagged items listed below for user review.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Critical | P0 bug item (`$PWD-walk-up`) still in backlog despite shipping v0.4.3 | Backlog.md P0 | Removed — replaced with resolution comment |
| 2 | Critical | P0 feature item (`non-interactive flags`) still in backlog despite shipping v0.4.2 | Backlog.md P0 | Removed — replaced with resolution comment |
| 3 | High | P1 feature item (`full agent-drivable CLI parity`) resolved by Plan 41 (2026-06-16) but still in backlog | Backlog.md P1 | Removed — replaced with resolution comment |
| 4 | High | HOMEBREW_TAP_TOKEN PAT reminder due 2026-07-15 — now 34 days past due; PAT likely expired | Backlog.md P1 | Flagged for user review (cannot rotate autonomously) |
| 5 | High | Website npm vulns: 6 open Dependabot alerts (esbuild HIGH, vite HIGH+MED, astro LOW) unresolved since 2026-06-16 | Backlog.md P2 | Flagged for user review |
| 6 | Medium | All 7 routines overdue by 60–118 days; no runs since 2026-05-07 | routines.md | Flagged for user review — other routines should be dispatched |
| 7 | Low | `[debt] Stale worktrees/branches` (120 days old) — backlog item itself is stale; situation may have worsened | Backlog.md P1 | Flagged — no autonomous action |
| 8 | Low | `[ops] Routine bot PR pile-up` (103 days old) — fix action not progressed | Backlog.md P1 | Flagged — no autonomous action |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **URGENT: HOMEBREW_TAP_TOKEN PAT** — The calendar reminder added 2026-04-22 targeted 2026-07-15 rotation. Today is 2026-08-18, 34 days past that date. Fine-grained PATs default to 90-day expiry (set 2026-04-22 → expires ~2026-07-21). The PAT has almost certainly expired. Next release will fail at the Homebrew formula step with `401 Bad credentials`. **Rotate now via GitHub Settings → Developer Settings → Personal Access Tokens.**

2. **Website npm vulnerabilities** — 6 open Dependabot alerts on `/website` (esbuild HIGH→0.28.1, vite HIGH+MED→7.3.5, js-yaml MED→4.2.0, astro+esbuild LOW). These have been open since 2026-06-16 (63 days). The astro upgrade PR (#108) fails the website build. Needs a real upgrade-with-build-fix pass.

3. **All other routines overdue** — Dependency Audit, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene, and Vulnerability Scan are all 60–118 days overdue with no log entries since 2026-05-07. Consider scheduling a full routine-digest session.

4. **`[ops] Routine bot PR pile-up`** — The fix (change cloud routine to commit-direct-to-main or auto-merge) was filed 2026-05-07 and is now 103 days old with no action. This will recur if cloud routines resume.

## Notes for Next Run

- P0 section is now clean (zero items; two resolution comments + one promoted-to-status comment).
- P1 count reduced by 1 (full CLI parity resolved by Plan 41).
- Backlog is otherwise accurate but many items are aging. A future sweep could trim verbose entries to NoteStandards (Group A bookkeeping item).
- If all routines are dispatched in a digest session, their findings should be captured in backlog before the next hygiene run.
- Check HOMEBREW_TAP_TOKEN first — it's the most time-sensitive item.
