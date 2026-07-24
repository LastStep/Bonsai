---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-07-24
status: success
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 8
  - `station/agent/Routines/roadmap-accuracy.md`
  - `station/agent/Core/identity.md`
  - `station/Playbook/Roadmap.md`
  - `station/Playbook/Status.md`
  - `station/Logs/KeyDecisionLog.md`
  - `station/agent/Core/routines.md`
  - `station/Logs/RoutineLog.md`
  - `station/Playbook/Backlog.md`
- **Files Modified:** 3
  - `station/Reports/Pending/2026-07-24-roadmap-accuracy.md` (created — this file)
  - `station/agent/Core/routines.md` (dashboard updated)
  - `station/Logs/RoutineLog.md` (entry appended)
- **Tools Used:** Read, Write, Edit, Bash (ls), Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

**Phase 1 — Foundation & Polish:** All 11 items are marked `[x]`. Cross-checked each against Status.md, RoutineLog.md, and code structure. Every item aligns with shipped work:
- `bonsai validate` (Plan 35, v0.4.0) ✓ — correctly reflects the "Better trigger sections" annotation added in the 2026-05-07 routine-digest
- Release pipeline, community health files, dogfooding — all confirmed ✓
- Phase 1 is correctly declared complete.

**Noteworthy omission in Phase 1:** Since the last roadmap-accuracy run (2026-05-07), two major features shipped that are not represented in Phase 1 even as informational rows:
- `--non-interactive --from-config` flags (Plan 39, v0.4.2, shipped 2026-05-13)
- Headless CLI Contract + MCP-ready cores (Plan 41, shipped 2026-06-16) — every mutating command now has a pure `*Result` headless core + JSONL/exit contract + `docs/agent-interface.md`

Phase 1 is complete, so these can't be added as open items. However, they represent significant capability additions that are invisible in the Roadmap.

**Phase 2 — Extensibility:** One item checked (`[x] Custom item detection`), three items unchecked:
- `[ ] Self-update mechanism`
- `[ ] Template variables expansion`
- `[ ] Micro-task fast path`

All three open items exist in Backlog at P3 ("Future Platform" section). No recent work addressed any of these. Development has effectively deprioritized Phase 2 remaining items in favor of Phase 3 pre-work. The Roadmap does not reflect this implicit priority decision.

**Phase 3 — Cloud & Orchestration:** Both items show `[ ]`:
- `[ ] Managed Agents integration — bonsai deploy, session management, outcome rubrics`
- `[ ] Greenhouse companion app`

**Key drift found here:** The Roadmap shows zero Phase 3 progress, but substantial Phase 3 groundwork has shipped since 2026-05-07:
- **Plan 40 (2026-06-13):** Frozen v1 schemas (`project.yaml`), root-relative scaffolding, project-level validate pass with adversarial hardening, memory-routing docs. Labeled "Odysseus platform integration" — directly targeting Managed Agents readiness.
- **Plan 41 (2026-06-16):** Pure `*Result` headless cores + JSONL/exit contract (`ExitConflict=5`) + `docs/agent-interface.md` for all mutating commands. Status.md explicitly states: "MCP server = fast-follow Plan 42."

Neither Plan 40 nor Plan 41 checks the Phase 3 box (correct — `bonsai deploy` hasn't shipped), but they represent pre-work that should be acknowledged in the Roadmap. The precedent for this pattern is the "Better trigger sections" item in Phase 1, which was checked and annotated _(shipped via Plans 08/17/21 + context-guard regex; intent-classification prompt-hook deferred to P3 Backlog as Plan 08 C3)_.

**Phase 4 — Ecosystem:** Both items show `[ ]`. No work has been done here. Correct.

### Step 2 — Check milestone accuracy

The next planned milestones as implied by the Roadmap are Phase 2 remaining items (self-update, template variables expansion, micro-task fast path). However, Status.md and RoutineLog show the actual next step is Plan 42 (MCP server), which is a Phase 3 item. This gap exists because:
1. The user has intentionally deprioritized Phase 2 remaining items (all at P3 in Backlog)
2. Plan 42 is referenced in Status.md's done row for Plan 41 but has no plan file in Plans/Active and no Backlog entry

Plan 42 is effectively "in flight" as an intent but invisible in all tracking artifacts. If development direction continues toward MCP/Managed Agents, Plan 42 should be in Backlog (at minimum P2) or Pending.

No roadmap items reference deprecated approaches. All Phase 3/4 items remain plausible targets.

### Step 3 — Cross-check against Key Decision Log

Reviewed all decisions in KeyDecisionLog.md:

- **"Defer Managed Agents cloud integration until local foundation is stable" (Settled, 2026-04-02):** Plans 40 and 41 are exactly the kind of "local foundation becoming stable" work this decision anticipated. The cloud integration (`bonsai deploy`, session management, outcome rubrics) still hasn't shipped — so the settled decision technically still holds. However, the phrasing "defer until stable" implies a future check-in point. The foundation is now arguably stable enough to begin Phase 3 proper. No formal decision has been logged that "the foundation is now stable." The Settled section should be updated to note that Plans 40/41 represent the foundation stabilization.
- All other Key Decisions were checked — none invalidate any Roadmap item.

### Step 4 — Report findings

See Findings Summary below. No Roadmap.md edits made (audit-only per procedure).

### Step 5 — Update dashboard

Done. Roadmap Accuracy row in `agent/Core/routines.md` updated: Last Ran → 2026-07-24, Next Due → 2026-08-07.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | Phase 3 "Managed Agents integration" shows no progress, but Plans 40+41 (Odysseus schemas + headless CLI contract) constitute substantial Phase 3 pre-work. Roadmap doesn't reflect this — gives a false picture that Phase 3 is untouched. | `Roadmap.md` Phase 3, first bullet | Flagged for user review — recommend annotation like Phase 1's "Better trigger sections" note, e.g. _(Plans 40/41 shipped headless CLI contract + agent-interface.md as MCP-readiness pre-work. `bonsai deploy` + session management not yet shipped.)_ |
| 2 | Low | Plan 42 (MCP server) referenced as "fast-follow" in Plan 41's Status.md done row, but has no plan file in Plans/Active and no Backlog entry. Risk of falling through the cracks as organizational memory. | `Status.md` Plan 41 done row | Flagged for user review — recommend adding a P2 Backlog entry: `[feature] MCP server (Plan 42) — fast-follow to Plan 41 headless CLI contract. Exposes init/add/update/remove as MCP tools for Managed Agents integration.` |
| 3 | Low | Phase 2 remaining items (self-update, template vars, micro-task fast path) are all at P3 in Backlog while development has moved to Phase 3 pre-work. The Roadmap implies sequential phases but actual trajectory skipped Phase 2 remainder. | `Roadmap.md` Phase 2, `Backlog.md` P3 | Flagged for user review — no correction required unless user wants to add a note to Phase 2 clarifying that the phase is "partially complete, remainder deferred." |
| 4 | Info | Phase 1 is correctly and completely marked done. All 11 items verified against Status.md, RoutineLog, and codebase. | `Roadmap.md` Phase 1 | No action needed. |
| 5 | Info | KeyDecisionLog "Defer Managed Agents" settled decision (2026-04-02) is functionally superseded — Plans 40/41 are the "foundation stabilization" that decision was waiting for. No formal successor decision logged. | `Logs/KeyDecisionLog.md` Settled section | Informational — user may want to add a note in Settled section that Plans 40/41 represent the anticipated foundation stabilization milestone. |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **Roadmap Phase 3 annotation** (Medium): Add a note under `[ ] Managed Agents integration` acknowledging that Plans 40 (Odysseus schemas) and 41 (headless CLI contract + `docs/agent-interface.md`) have shipped as Phase 3 pre-work. This mirrors the precedent set by the "Better trigger sections" item in Phase 1. Suggested annotation: _(Plans 40/41 shipped headless CLI contract + MCP-ready cores as pre-work. `bonsai deploy` + session management remain.)_

2. **Plan 42 tracking gap** (Low): The "fast-follow MCP server" is referenced but untracked. Recommend filing in Backlog as P2 `[feature]` item to prevent it from becoming orphaned intent.

3. **Phase 2 sequencing** (Low): Consider adding a parenthetical to Phase 2 header or open items noting that these are deferred in favor of Phase 3 groundwork. Optional — the Roadmap is aspirational, but clarity helps future routine runs.

## Notes for Next Run

- Phase 1 fully done — verified for the last time; no need to re-audit Phase 1 checkboxes.
- The main drift to watch going forward is Phase 3 progress: if Plan 42 ships and `bonsai deploy` advances, the Phase 3 Roadmap row should be updated.
- If the user adds an annotation to Phase 3 addressing Finding 1, confirm it in the next run.
- Check whether Plan 40 Phase 4 (update-delivery, currently HELD) is ever revived — if so, the Roadmap should reflect it.
- Plans 40 + 41 are still in Plans/Active (also flagged by Status Hygiene today, 2026-07-24) — verify those moved to Plans/Archive/ next run.
