---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-06-13
status: partial
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~10 min
- **Files Read:** 5 — `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state
- **Action:** Read `Roadmap.md` and cross-checked all `[x]`/`[ ]` items against `Status.md` Recently Done table and known shipped versions (v0.4.0–v0.4.3).
- **Result:** All Phase 1 items marked `[x]` are accurate — every listed item has shipped. The `bonsai validate` item (added in last cycle) is correctly marked `[x]`. Phase 2 custom item detection `[x]` is also accurate. However, the roadmap's "Current Phase" heading still says "Phase 1 — Foundation & Polish" despite all Phase 1 items being complete and active work now operating in Phase 2 territory. This is the primary drift.
- **Issues:** Roadmap heading claims Phase 1 is current when it is fully complete. Phase 2 is the active phase.

### Step 2: Check milestone accuracy
- **Action:** Reviewed Status.md Recently Done (last 37 days) and Backlog for priority shifts.
- **Result:** Three significant gaps found:
  1. **Plan 40 (Odysseus platform integration)** — Phases 1–3 merged (2026-06-13). Introduces `.bonsai/project.yaml`, frozen v1 schemas, root-relative scaffolding, memory routing docs, and guide Formats page. None of this is represented in the roadmap. Phase 4 (update-delivery) was HELD. The hub/Odysseus angle is closest to Phase 3 "Managed Agents integration" but is a different scope — it's platform scaffolding plumbing, not full cloud deployment.
  2. **Full agent-drivable CLI parity** — Promoted to P1 Backlog (2026-06-13) as the user's stated "main thing." `init`+`add` have `--non-interactive`/`--from-config` (v0.4.2); `update`/`remove` are TUI-only. This is Phase 2 extensibility scope but not listed in the roadmap's Phase 2 items.
  3. **Phase 2 "Self-update mechanism"** — Roadmap lists this as `[ ]` (unstarted). Plan 40's update-delivery slice (Phase 4, HELD) was scoping exactly this. The roadmap item should be annotated to reflect that design work has begun (partially blocked, HELD in Plan 40 Phase 4).
- **Issues:** Three roadmap gaps flagged — see Findings Summary.

### Step 3: Cross-check against Key Decision Log
- **Action:** Read `KeyDecisionLog.md` in full and compared decision dates against recent work (2026-05-07 through 2026-06-13).
- **Result:** The most recent KDL entry is dated 2026-04-13. In the 37 days since last roadmap-accuracy run, the project shipped Plan 40 (frozen v1 schemas, hub integration model, project.yaml spec, memory-routing protocol), v0.4.2 (non-interactive CLI contract), and v0.4.3 (hook path baking). Each of these involved significant architectural decisions that belong in the KDL but were not logged there. The log is visibly stale — no entries for 2+ months of active development.
- **Issues:** KeyDecisionLog has not been updated since 2026-04-13 despite substantial architectural work. At least 3 new Structural-tier decisions identified.

### Step 4: Report findings
- **Action:** Compiled findings below. Per procedure: not modifying Roadmap.md directly — flagging for user review.
- **Result:** 5 findings across 3 severity levels. All require user action to resolve (roadmap is user-owned).
- **Issues:** None blocking routine completion.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Roadmap Accuracy row — Last Ran → 2026-06-13, Next Due → 2026-06-27, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | "Current Phase" heading still says Phase 1 — all items `[x]`, Phase 2 is now active | `Roadmap.md` L15 | Flagged — not modified |
| 2 | medium | Plan 40 Odysseus integration (hub scaffolding, project.yaml, frozen schemas) not represented in roadmap | `Roadmap.md` Phase 2/3 | Flagged — not modified |
| 3 | medium | Full agent-drivable CLI parity (P1 Backlog, user's "main thing") has no roadmap entry | `Roadmap.md` Phase 2 | Flagged — not modified |
| 4 | low | Phase 2 "Self-update mechanism" `[ ]` doesn't reflect Plan 40 Phase 4 design work (HELD, but scoped) | `Roadmap.md` Phase 2 | Flagged — not modified |
| 5 | low | KeyDecisionLog.md has no entries since 2026-04-13 — 2+ months of architectural work unlogged | `Logs/KeyDecisionLog.md` | Flagged — not modified |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**Finding 1 — Phase 1 → Phase 2 heading transition:**
All Phase 1 items are `[x]`. Suggest moving "Current Phase" to Phase 2 — Extensibility in `Roadmap.md`. Suggested wording change: rename Phase 2 heading to "### Phase 2 — Extensibility _(current)_" and demote Phase 1 to "### Phase 1 — Foundation & Polish _(complete)_".

**Finding 2 — Plan 40 Odysseus platform work:**
Plan 40 Phases 1–3 introduced hub-facing features (`.bonsai/project.yaml`, frozen v1 schemas, memory routing). This doesn't fit cleanly under any existing roadmap item. Options:
- (a) Add a Phase 2 item: `[ ] Hub integration foundation — project.yaml, frozen v1 schemas, Odysseus platform scaffolding (Plan 40 P1–P3 shipped; Phase 4 HELD)`
- (b) Annotate the Phase 3 "Managed Agents integration" item to note Phase 40 partial progress

**Finding 3 — Full agent-drivable CLI parity:**
This is the user's stated "main thing" and the top P1 Backlog item. It doesn't map to any Phase 2 roadmap item. Suggest adding: `[ ] Full non-interactive CLI parity — headless init/add/update/remove with JSONL/exit-code contract for AI agent automation`

**Finding 4 — Self-update mechanism annotation:**
Phase 2 roadmap lists `[ ] Self-update mechanism`. Plan 40 Phase 4 scoped this (update-delivery mechanism) but was held. Suggest annotating: `[ ] Self-update mechanism — catalog items can flag when stale or have issues _(Plan 40 Phase 4 design started; HELD)_`

**Finding 5 — KeyDecisionLog refresh:**
Multiple Structural-tier decisions since 2026-04-13 are unlogged:
- Non-interactive CLI contract (JSONL + exit codes 0/2/3/4, v0.4.2)
- Frozen v1 schema design (Plan 40 — `.bonsai/project.yaml` immutable-key model)
- Hook path baking at install time (v0.4.3 — absolute paths vs `$PWD` walk-up)
Recommend a KDL update pass in the next session.

## Notes for Next Run

- Phase 1 → Phase 2 heading transition is the main structural drift item — recommend resolving before next cycle so the roadmap correctly signals active phase
- If Plan 40 Phase 4 (update-delivery) remains HELD at next run, consider demoting self-update mechanism to `[ ] - deferred` or adding a `[→ Plan 40 Phase 4, HELD]` annotation
- KeyDecisionLog staleness is a recurring pattern — if still unupdated at next cycle, may warrant a dedicated KDL refresh step added to this routine's procedure
- Check whether full agent-drivable CLI parity has been promoted to an active plan — if P1 Backlog item lingers unstarted at next cycle, escalate
