---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-06-14
status: partial
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~8 min
- **Files Read:** 6
  - `/home/user/Bonsai/station/agent/Routines/roadmap-accuracy.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/station/Playbook/Plans/Active/40-odysseus-platform-integration.md`
- **Files Modified:** 3
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
  - `/home/user/Bonsai/station/Reports/Pending/2026-06-14-roadmap-accuracy.md` (this report)
- **Tools Used:** Read, Write, Edit, Bash, Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `Roadmap.md`. Current state:

**Phase 1 — Foundation & Polish:** All 11 items are `[x]`. This matches reality — the prior routine (2026-05-07) confirmed "Better trigger sections" was annotated and checked, and `bonsai validate` was added. Phase 1 is accurate.

**Phase 2 — Extensibility:** 1 of 4 items checked (`[x] Custom item detection`). The 3 remaining (`[ ] Self-update`, `[ ] Template variables expansion`, `[ ] Micro-task fast path`) are all unshipped. However, **Plan 40 Phases 1–3 shipped significant new capabilities (PRs #114/#115/#116, merged 2026-06-13) that are not represented in Phase 2 or anywhere else on the Roadmap.** Specifically:

- Frozen v1 schemas for memory notes and project manifest (`.bonsai/project.yaml`)
- `validate` lint extended to 12 categories (was 6 — Plan 40 Phase 2 added 6 more covering memory-note and manifest validation)
- `station/Memory/` scaffolding items (project-manifest + memory, both opt-in, required: false)
- Memory routing docs update (howToWorkLines in generate.go)
- New `bonsai guide` "Formats" page

These are additive features shipped as v0.5.0 (untagged, user-held). They don't map cleanly to any existing Phase 2 item. The Roadmap has drift relative to what was built.

**Phase 3 — Cloud & Orchestration:** Both items unshipped. Managed Agents integration deferred per 2026-04-02 Settled decision. No change needed.

**Phase 4 — Ecosystem:** All items unshipped. No change needed.

### Step 2 — Check milestone accuracy

The "Current Phase" header still reads "Phase 1 — Foundation & Polish" despite Phase 1 being fully complete. Phase 2 is underway (one item checked, more shipped under Plan 40 that aren't listed). The "Current Phase" designation is stale — should point to Phase 2.

The remaining Phase 2 items (Self-update, Template variables expansion, Micro-task fast path) are still correct priorities per the Backlog. Self-update was the Phase 4 (HELD) item from Plan 40, still pending.

### Step 3 — Cross-check against Key Decision Log

Read `KeyDecisionLog.md`. Key relevant decisions:

- **2026-04-02 Settled:** "Defer Managed Agents cloud integration until local foundation is stable." → Phase 3 items appropriately remain unchecked.
- No decisions in the log invalidate any Phase 2 or Phase 3 roadmap items.
- The Plan 40 "Config split = separate" and "Manifest location = `.bonsai/project.yaml`" decisions (locked 2026-06-13) represent new architectural choices not reflected in the Roadmap.

No decisions in the KeyDecisionLog actively contradict or deprecate any roadmap item. The gap is additive — new work shipped that has no roadmap entry.

### Step 4 — Report findings

Findings flagged for user review (not modifying Roadmap.md per routine protocol). See Findings Summary below.

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` Roadmap Accuracy row: Last Ran → 2026-06-14, Next Due → 2026-06-28, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | "Current Phase" header still says Phase 1. Phase 1 is fully complete; work is in Phase 2. | `Roadmap.md` line 16 | Flagged — user to update header to "Phase 2 — Extensibility" |
| 2 | MEDIUM | Plan 40 Phases 1–3 (shipped 2026-06-13) introduced frozen v1 memory-note + project-manifest schemas, 6 new `validate` lint categories, `station/Memory/` scaffolding, and a `bonsai guide` Formats page. None of these are listed in Phase 2 (or anywhere on the Roadmap). | `Roadmap.md` Phase 2 section | Flagged — user to add 1–2 bullet(s) capturing what shipped in Plan 40 (or accept that the Roadmap is milestone-level and these are implementation details) |
| 3 | LOW | Phase 4 (update-delivery) from Plan 40 is HELD and has been folded into the "headless-CLI parity" workstream (per Status.md). This may eventually become a new Phase 2 item (Self-update mechanism via `bonsai update`), but the connection isn't explicit. | `Roadmap.md` Phase 2 + Status.md | Flagged — no immediate action needed; watch whether "headless-CLI parity" displaces the existing Phase 2 ordering |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

**Finding 1 (MEDIUM) — Stale "Current Phase" header:**
The roadmap header reads "### Phase 1 — Foundation & Polish" as the active phase, but Phase 1 is complete (`[x]` on all 11 items). Recommended change:

```markdown
## Current Phase

### Phase 2 — Extensibility
```

Remove the "Phase 1" section from "Current Phase" or convert it to a "Completed Phases" block.

**Finding 2 (MEDIUM) — Plan 40 shipped features missing from Phase 2:**
Plan 40 Phases 1–3 (PRs #114/#115/#116, merged 2026-06-13, v0.5.0 untagged) shipped:
- `.bonsai/project.yaml` — project manifest (slug, title, description, permalink, agent type)
- `station/Memory/` scaffolding items (memory-note + manifest opt-in scaffolding)
- `validate` expanded to 12 lint categories (was 6; added memory-note + manifest validation)
- `bonsai guide` Formats page documenting both schemas

These aren't well-captured by any existing Phase 2 bullet. Possible options:
1. Add a `[x]` bullet: "Odysseus integration artifacts — project manifest, memory scaffolding, validate lint (12 categories), guide Formats page" under Phase 2.
2. Accept that the Roadmap operates at milestone abstraction level and leave as-is (these are implementation details of extensibility, not new milestones).
3. Mark a new Phase 2 bullet `[x] Schema standards & in-repo memory graph — project-manifest + memory-note frozen v1 schemas with validate lint`.

User to decide.

**Finding 3 (LOW) — Phase 4 HELD / headless-CLI parity connection:**
Plan 40's Phase 4 (update-delivery via `bonsai update`) was held pending a separate headless-CLI parity workstream (per Status.md 2026-06-13 dispatch note). This held work may eventually map to Phase 2's "Self-update mechanism" bullet. No roadmap change needed now, but worth tracking when headless-CLI planning begins.

---

## Notes for Next Run

- If user acts on Finding 1: confirm "Current Phase" header now points to Phase 2.
- If user acts on Finding 2: confirm new Plan 40 bullet(s) added under Phase 2.
- Phase 2 has 1 checked item + 3 unchecked. Progress is slow but intentional — headless-CLI parity (P1 Backlog) is the next major workstream.
- Phase 3 (Managed Agents / Greenhouse) remains appropriately deferred per 2026-04-02 Settled decision.
- KeyDecisionLog is clean — no decisions contradict current roadmap direction.
