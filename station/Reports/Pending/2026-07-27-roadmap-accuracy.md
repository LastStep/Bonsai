---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-07-27
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
- **Duration:** ~7 min
- **Files Read:** 7 — `/home/user/Bonsai/station/agent/Routines/roadmap-accuracy.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/40-odysseus-platform-integration.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/41-headless-cli-contract.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read (7x), Glob (2x), Edit (2x), Write (1x)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `Playbook/Roadmap.md`. Compared each roadmap item against known shipped work in `Status.md` and `Plans/Active/`.

**Phase 1 — Foundation & Polish (labeled "Current Phase"):**
All 11 items are checked [x]. Phase 1 is 100% complete. The two items flagged in the 2026-05-07 run ("Better trigger sections" unchecked, missing `bonsai validate` row) have since been resolved — both are now properly checked with annotations. However, the roadmap still labels Phase 1 as the "## Current Phase" section header, which is inaccurate.

**Phase 2 — Extensibility (labeled "Future Phases"):**
One of four items is checked [x] (custom item detection, shipped via Plans 26/34 in Archive). The remaining three (self-update mechanism, template variables expansion, micro-task fast path) are still open. Phase 2 has already begun but is labeled "Future."

**Significant shipped work not in the roadmap:**
- Plan 41 — Headless CLI Contract + MCP-Ready Cores (all 5 phases, PRs #120/#122/#123/#121/#125, main `ab202c3`, 2026-06-16): every mutating command (init/add/update/remove) now has a pure `*Result` headless core + JSONL/exit contract; `list --json`; `docs/agent-interface.md` contract doc. No roadmap entry exists for this work.
- Plan 40 — Odysseus Platform Integration Phases 1-3 (PRs #114/#116/#115, v0.5.0 untagged, 2026-06-13): frozen v1 schemas (memory notes + project manifest), `.bonsai/project.yaml` scaffolding, validate project-level pass, memory routing docs. Phase 4 held and superseded by Plan 41. No roadmap entry exists for this work.

**Upcoming work not in the roadmap:**
- Plan 42 — MCP server (fast-follow to Plan 41, not started but planned). Described in Plan 41 as a thin wrapper calling the same headless cores. This directly implements Phase 3's "Managed Agents integration" goal but is not mentioned in Roadmap.md.

### Step 2 — Check milestone accuracy

**Phase 3 — Cloud & Orchestration** lists two items:
1. "Managed Agents integration — bonsai deploy, session management, outcome rubrics"
2. "Greenhouse companion app — desktop app for managing projects + observing AI agents"

The Key Decision Log records (2026-04-02): "Defer Managed Agents cloud integration until local foundation is stable." Plan 41 has now shipped the headless CLI contract — the local foundation is substantially more stable than when that decision was made. Plan 42 (MCP server, fast-follow) is the next natural step and directly maps to "Managed Agents integration" in Phase 3. The roadmap does not reflect that Phase 3 is becoming more imminent.

**Phase 4 — Ecosystem** (catalog marketplace, plugin system, cross-project coordination): no work started, no decisions affecting these items. These remain accurate far-future items.

### Step 3 — Cross-check against Key Decision Log

Read `Logs/KeyDecisionLog.md`. All entries are dated 2026-04-12 or 2026-04-13. No entries have been added since then despite significant architectural decisions made during Plans 40 and 41, including:
- Two-format serialization split (JSONL for mutating commands, single-doc JSON for reads)
- Headless-CLI contract shape (typed `Result`-returning cores, no TTY dependency)
- Warning stream isolation (warnings → stderr only, never stdout)
- ExitConflict=5 exit code contract
- `.bonsai/project.yaml` manifest as project identity source (documented known drift from `.bonsai.yaml`)
- Memory routing decision (decisions → `Memory/decisions/`, not KeyDecisionLog)

None of the existing KeyDecisionLog entries are invalidated by recent decisions. However, the log is substantially out of date.

### Step 4 — Report findings

Findings compiled below. No changes made to `Roadmap.md` (per procedure: flag, do not modify directly).

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` dashboard row for Roadmap Accuracy: Last Ran → 2026-07-27, Next Due → 2026-08-10, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | Phase 1 labeled "Current Phase" but is 100% complete (all items [x]). Phase 2 has started. Section header is misleading. | `Roadmap.md` header `## Current Phase` | Flagged for user — suggest restructuring: move Phase 1 under a "Completed" section, make Phase 2 the current phase |
| 2 | HIGH | Plan 41 (Headless CLI Contract + MCP-Ready Cores, all 5 phases shipped 2026-06-16) has no roadmap entry. This is the most significant shipped work in the past 2.5 months. | `Roadmap.md` — absent | Flagged for user — suggest adding as a Phase 2 item (headless/agent-drivable CLI) or a Phase 3 prerequisite row |
| 3 | MEDIUM | Plan 40 Phases 1-3 (Odysseus Platform Integration, shipped 2026-06-13, v0.5.0 untagged) has no roadmap entry. Introduces project manifest and memory scaffolding — significant extensibility work. | `Roadmap.md` — absent | Flagged for user — suggest adding under Phase 2 (Extensibility) |
| 4 | MEDIUM | Plan 42 (MCP server, fast-follow to Plan 41) maps directly to Phase 3 "Managed Agents integration" but is not mentioned. The prerequisite work (Plan 41) is now shipped. | `Roadmap.md` Phase 3 | Flagged for user — suggest adding `bonsai mcp` as a Phase 3 item, noting Phase 3 is now actionable |
| 5 | LOW | Phase 2 is labeled "Future Phases" despite being in progress (custom item detection done, Plans 40+41 represent extensibility work). | `Roadmap.md` `## Future Phases` header | Flagged for user — consider relabeling Phase 2 section |
| 6 | LOW | KeyDecisionLog has no entries since 2026-04-13. Plans 40 and 41 produced ~6 significant architectural decisions (serialization contracts, stream hygiene, exit code contract, etc.) that should be logged. | `Logs/KeyDecisionLog.md` | Flagged for user |
| 7 | LOW | Plan 41 is still in `Plans/Active/` despite being fully shipped (all 5 phases, main `ab202c3`). Multiple other routines today flagged this same issue. | `Plans/Active/41-headless-cli-contract.md` | Flagged for user — archive at next opportunity |

## Errors & Warnings

None.

## Items Flagged for User Review

**HIGH:**

1. **Roadmap structural refresh needed** — Phase 1 is complete and Phase 2 has started. The roadmap structure should be updated: Phase 1 → completed section, Phase 2 → current phase. Suggested correction to `Roadmap.md`:
   - Rename `## Current Phase` → `## Current Phase — Phase 2 — Extensibility`
   - Move Phase 1 under a `## Completed` or `## Phase 1 — Foundation & Polish (Done)` section
   - Update `## Future Phases` to start from Phase 3

2. **Plan 41 missing from roadmap** — The headless CLI contract is the most significant shipped work since Plan 35 (`bonsai validate`). Suggested addition under Phase 2:
   ```
   - [x] Headless CLI contract — every mutating command agent-drivable (non-interactive cores, JSONL/exit contract, list --json, documented agent-interface)
   ```

**MEDIUM:**

3. **Plan 40 work missing from roadmap** — The project manifest (`.bonsai/project.yaml`) and memory scaffolding (`station/Memory/`) are shipped extensibility features. Suggested addition under Phase 2:
   ```
   - [x] Project manifest + memory scaffolding — .bonsai/project.yaml, station/Memory/ graph, validate project-level pass
   ```

4. **Phase 3 is now actionable — suggest updating** — Plan 42 (MCP server) is the next planned work item and directly implements the Phase 3 "Managed Agents integration" goal. Consider adding:
   ```
   - [ ] bonsai mcp — stdio MCP server (Plan 42, fast-follow to headless contract)
   ```
   And noting the "Defer cloud integration" decision from 2026-04-02 may be ready to revisit.

**LOW:**

5. **KeyDecisionLog is 3 months stale** — Suggest logging key decisions from Plans 40 and 41 during the next planning session. At minimum: two-format serialization split, headless-CLI contract shape, warning-stderr-only policy, ExitConflict=5.

6. **Plan 41 archive** — Move `Plans/Active/41-headless-cli-contract.md` → `Plans/Archive/` (third routine today to flag this).

## Notes for Next Run

- Previous run (2026-05-07) flagged "Better trigger sections" unchecked — resolved before this run (now [x] with annotation). Good.
- Previous run flagged missing `bonsai validate` row — resolved (now [x] with annotation). Good.
- Next run should verify: (a) Phase 1/2 restructuring done, (b) Plans 40/41 added to roadmap, (c) Plan 42 status (shipped or still pending).
- If Plan 42 (MCP server) ships before the next run, it should appear as [x] under Phase 3 at that point.
- The "hold on v0.5.0 tag" (user, 2026-06-13) was not addressed in this run — worth checking status at next session.
