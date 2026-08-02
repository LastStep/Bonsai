---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-08-02
status: success
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 5
  - `/home/user/Bonsai/station/agent/Routines/roadmap-accuracy.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`
  - `/home/user/Bonsai/station/Playbook/Backlog.md`
- **Files Modified:** 2
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `Playbook/Roadmap.md` and cross-checked against recently shipped work in `Status.md`.

**Phase 1 — Foundation & Polish (labeled "Current Phase"):**
All 11 items are checked `[x]`. Phase 1 is fully complete. The previous run (2026-05-07) fixed the two remaining unchecked boxes ("Better trigger sections" and `bonsai validate`). No Phase 1 drift found.

The "## Current Phase" heading now labels a phase that is 100% done. All work since May has been Phase 2/3 preparation. This label is technically stale.

**Status.md vs. Roadmap alignment:**
- In Progress: none
- Pending: sentrux trial (blocked, Rust toolchain not installed) — not roadmap-related
- Recently Done: Plan 41 (Headless CLI Contract + MCP-ready cores, 2026-06-16) and Plan 40 (Odysseus Platform Integration v0.5.0, 2026-06-13) — neither is reflected in the roadmap

### Step 2 — Check milestone accuracy

**Phase 2 — Extensibility (remaining `[ ]` items):**
- `Self-update mechanism` — still valid priority. Notably, Plan 40 Phase 4 (update delivery path for existing projects) was explicitly held pending headless CLI work; this maps directly to this item and could now be started.
- `Template variables expansion` — no recent work; still valid.
- `Micro-task fast path` — no recent work; still valid.

**Phase 3 — Cloud & Orchestration:**
- `Managed Agents integration` — Plans 40+41 have laid significant groundwork: Odysseus v1 schemas, root-relative scaffolding, memory-routing docs, and the headless CLI cores are all shipped. The `[ ]` status is technically correct (`bonsai deploy`, session management, and outcome rubrics not yet shipped), but partial progress exists.
- `Greenhouse companion app` — no work; still future.

**Missing from roadmap — Plan 41 / MCP server:**
Plan 41 shipped a headless CLI API + JSONL/exit contract explicitly described as "MCP-ready cores." Status.md notes "MCP server = fast-follow Plan 42." This significant upcoming feature (Plan 42) is not reflected in any roadmap item. The headless CLI + MCP pathway logically belongs in Phase 3 (Cloud & Orchestration) as an enabling step before "Managed Agents integration."

### Step 3 — Cross-check against Key Decision Log

Read `Logs/KeyDecisionLog.md`. Key findings:

- "Defer Managed Agents cloud integration until local foundation is stable" (2026-04-02, Settled) — This deferral decision was made early. The local foundation is now significantly more mature (Plans 40+41, v0.4.x/v0.5.0 shipped, headless API live). The timing for beginning serious Phase 3 work is approaching. The decision is not invalidated, but its premise (unstable local foundation) has been addressed.
- No decisions directly contradict or invalidate any current roadmap items.
- All Phase 2/3/4 items remain technically uncompleted per their definitions — no decisions deprecated them.

### Step 4 — Report findings (no Roadmap.md edits — flagged for user review)

See Findings Summary below.

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` — Roadmap Accuracy row `Last Ran` → 2026-08-02, `Next Due` → 2026-08-16, `Status` → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Low | "Current Phase" heading labels Phase 1, which is 100% complete — heading is stale | `Roadmap.md` line 14 | Flagged for user review — do not modify |
| 2 | Medium | Plan 41 headless CLI + upcoming Plan 42 MCP server not in roadmap — significant shipped infrastructure and near-term feature absent | `Roadmap.md` Phase 3 | Flagged for user review — recommend adding MCP server as a Phase 3 milestone |
| 3 | Low | Phase 3 "Managed Agents integration" has partial foundation work (Plans 40+41) but remains unchecked — an annotation noting partial progress would be accurate | `Roadmap.md` Phase 3 | Flagged for user review |
| 4 | Info | Phase 2 "Self-update mechanism" maps directly to Plan 40 Phase 4 (update delivery, held) — when Plan 40 Phase 4 resumes, check this box | `Roadmap.md` Phase 2 | No action needed now |
| 5 | Info | Phase 2/3/4 unchecked items remain valid priorities — no decisions supersede them, no deprecated approaches found | `Roadmap.md` Phase 2–4 | No action needed |

---

## Errors & Warnings

None.

---

## Items Flagged for User Review

### F1 (Low) — Phase 1 "Current Phase" label is stale
All 11 Phase 1 items are `[x]`. The `## Current Phase` heading should be updated — either rename to `## Completed Phases` and add a `## Current Phase` section pointing to Phase 2, or restructure the document so Phase 2 becomes the current phase header.

**Suggested edit to Roadmap.md:**
```
## Completed Phases

### Phase 1 — Foundation & Polish ✓
...

## Current Phase

### Phase 2 — Extensibility
```

### F2 (Medium) — MCP server (Plan 42) not in roadmap

Plan 41 shipped "MCP-ready cores" (headless CLI + JSONL/exit contract). Status.md notes "MCP server = fast-follow Plan 42." This is not represented in the roadmap anywhere.

**Suggested addition to Phase 3:**
```
- [ ] MCP server — expose bonsai init/add/update/remove as MCP tools for agent-drivable workspace setup
```

This milestone sits between Phase 2 (Extensibility) and Phase 3 (Cloud & Orchestration) conceptually, but Phase 3 is the better home since MCP is a cloud/orchestration enabler.

### F3 (Low) — Phase 3 partial progress annotation

Plans 40+41 have done significant Phase 3 groundwork. A note on the "Managed Agents integration" item would prevent future confusion:

**Suggested annotation:**
```
- [ ] Managed Agents integration — `bonsai deploy`, session management, outcome rubrics _(Plans 40+41 shipped foundational schemas + headless cores; bonsai deploy + session management remain)_
```

---

## Notes for Next Run

- If Plan 42 (MCP server) has shipped by next run, check the MCP server roadmap item if added.
- If Plan 40 Phase 4 (update delivery) resumes, check "Self-update mechanism" in Phase 2.
- Monitor whether Plan 42 is formally written (it was referenced as "fast-follow" in Plan 41 notes) — if so, confirm it maps to a roadmap item.
- "Current Phase" heading change is a cosmetic fix the user can do in <2 minutes; prompt during routine-digest.
