---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-07-11
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
- **Duration:** ~8 minutes
- **Files Read:** 6 — `/home/user/Bonsai/station/agent/Routines/roadmap-accuracy.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `/home/user/Bonsai/station/Reports/Pending/2026-07-11-roadmap-accuracy.md` (this report), `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state
- **Action:** Read `station/Playbook/Roadmap.md` and cross-referenced against `station/Playbook/Status.md` and recent RoutineLog entries to determine what has actually shipped since the last run (2026-05-07).
- **Result:** Phase 1 is fully complete — all checkboxes are `[x]`, consistent with the last Routine Digest (2026-05-07) which applied the `bonsai validate` row addition and the "Better trigger sections" annotation. Phase 1 is accurate.
  - Phase 2 shows `[x] Custom item detection` and three open items. However, two major features have shipped since 2026-05-07 that are not captured in any roadmap phase:
    - **Plan 41 (2026-06-16):** Headless CLI Contract + MCP-ready cores — all 5 phases merged (PRs #120/#122/#123/#121/#125). Every mutating command has a pure `*Result` headless core + JSONL/exit contract (`ExitConflict=5`); `list --json`; `docs/agent-interface.md` contract doc. This is a major architectural milestone with no roadmap entry.
    - **Plan 40 (2026-06-13):** Odysseus Platform Integration Phases 1–3 (v0.5.0) — frozen v1 schemas, root-relative scaffolding, project-level `validate` pass. Phase 4 HELD by user; dogfood deferred; tag held. No roadmap entry.
  - Phase 3 and Phase 4 are unchanged — all items open, no work started.
- **Issues:** Two major shipped features are not reflected in the roadmap. One active next-priority item (MCP server, Plan 42) is not on the roadmap.

### Step 2: Check milestone accuracy
- **Action:** Reviewed Phase 2 open items for priority alignment against current Status.md and recent RoutineLog entries.
- **Result:**
  - Phase 2 open items — `Self-update mechanism`, `Template variables expansion`, `Micro-task fast path` — are still valid future items, but the Backlog Hygiene routine run today (2026-07-11) flagged that `Self-update mechanism` and `Micro-task fast path` are P3 Backlog items being considered for P2 promotion. This means the roadmap's implied ordering (these come before Phase 3) may not reflect actual development priorities.
  - The active next priority from Status.md is **Plan 42 (MCP Server)** referenced as "fast-follow" to Plan 41. This is not on the roadmap at all.
  - Phase 3 "Managed Agents integration" (`bonsai deploy`) and "Greenhouse companion app" (Tauri + Svelte + SQLite) have no recent progress and remain appropriately future.
  - No roadmap items reference deprecated approaches.
- **Issues:** Phase 2 ordering may be misleading — the actual next-priority work (MCP server) is unlisted, while the three open Phase 2 items are lower priority in practice.

### Step 3: Cross-check against Key Decision Log
- **Action:** Read `station/Logs/KeyDecisionLog.md` in full; checked all Structural, Domain-Specific, and Settled decisions against roadmap items.
- **Result:**
  - **No decision invalidates any existing roadmap item.**
  - Structural decisions (Go rewrite, embed.FS catalog, text/template, lock file, tech-lead always required) are all consistent with Phase 1 being marked done.
  - The Settled decision "Defer Managed Agents cloud integration until local foundation is stable" is still consistent with Phase 3 being open. The Plan 41 headless CLI + MCP-ready cores work represents maturation of the local foundation, but Phase 3 has not been triggered/started.
  - The Settled decision "Bonsai is a scaffolding tool, not a runtime orchestrator" remains consistent with all roadmap phases.
- **Issues:** None — KeyDecisionLog is clean relative to the roadmap.

### Step 4: Report findings
- **Action:** Compiled findings below. Per procedure, no edits to Roadmap.md — flagged for user review.
- **Result:** 4 findings identified (2 HIGH, 2 MEDIUM). All flagged for user decision.
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Roadmap Accuracy row (Last Ran, Next Due, Status).
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | **Headless CLI Contract (Plan 41) not on roadmap** — All 5 phases merged 2026-06-16 (PRs #120/#122/#123/#121/#125). Every mutating command has a headless `*Result` core + JSONL/exit contract + `docs/agent-interface.md`. This is a major shipped architectural milestone with no roadmap entry in any phase. | `station/Playbook/Roadmap.md` — Phase 2 | Flagged for user. Suggest adding: `- [x] Headless CLI + agent-drivable interface — JSONL/exit contract, pure Result cores, agent-interface.md contract doc (Plan 41)` under Phase 2 |
| 2 | HIGH | **MCP Server (Plan 42) active next-priority but not on roadmap** — Status.md explicitly says "MCP server = fast-follow Plan 42." This is the identified next milestone, with no roadmap representation. | `station/Playbook/Roadmap.md` — Phase 2 or Phase 3 | Flagged for user. Suggest adding: `- [ ] MCP server — expose CLI operations via Model Context Protocol (Plan 42)` under Phase 2 or as Phase 3 prerequisite |
| 3 | MEDIUM | **Phase 2 open items out of sync with actual priority order** — `Self-update mechanism` and `Micro-task fast path` are P3 Backlog items (today's Backlog Hygiene routine flagged them as P2 promotion candidates but not yet promoted). MCP server is the actual next priority but isn't listed. Roadmap implies a priority order that doesn't match current work. | `station/Playbook/Roadmap.md` — Phase 2 open items | Flagged for user. No edit needed until user decides on MCP server roadmap placement and Backlog P2 promotions |
| 4 | MEDIUM | **v0.5.0 work not marked in roadmap** — Plan 40 (Odysseus: frozen schemas + root-relative scaffolding + project-level validate) and Plan 41 (headless CLI) together constitute a v0.5.0 release. Roadmap only has a version marker for `v0.4.0 headline` on `bonsai validate`. Note also: Plan 40 Phase 4 is HELD and tag is still held by user. | `station/Playbook/Roadmap.md` — Phase 2 | Flagged for user. Consider adding `v0.5.0 headline` annotation to `bonsai validate` project-level variant or headless CLI row once user decides to cut the tag |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[HIGH] Should Plan 41 (Headless CLI Contract) appear in Roadmap Phase 2 as a completed milestone?** Suggested text: `- [x] Headless CLI + agent-drivable interface — JSONL/exit contract, pure Result cores, agent-interface.md contract doc (Plan 41)`

2. **[HIGH] Should MCP Server (Plan 42) be added to the roadmap?** It's the active next-priority. Suggest: `- [ ] MCP server — expose CLI operations via Model Context Protocol (Plan 42)` — decision needed on whether this belongs in Phase 2 (extensibility tooling) or Phase 3 (cloud/orchestration bridge).

3. **[MEDIUM] Should Phase 2 open items be reordered or annotated?** `Self-update mechanism` and `Micro-task fast path` are P3 Backlog items that may be lower priority than MCP server. User may want to annotate or restructure Phase 2 ordering.

4. **[MEDIUM] Is v0.5.0 tag still held?** Plan 40 RoutineLog notes "tag held (user)." Once the tag is cut, consider adding a `v0.5.0 headline` annotation to the appropriate roadmap item(s).

## Notes for Next Run

- Phase 1 is stable and accurate — no re-check needed.
- Watch for Plan 42 (MCP server) progress: if shipped, it will need a roadmap entry + `[x]` marker.
- Watch for v0.5.0 tag cut: once released, roadmap may need version annotation.
- Phase 2 open items (`Self-update mechanism`, `Template variables expansion`, `Micro-task fast path`) have been open since the project's Phase 2 definition. If MCP server lands and Phase 3 starts accelerating, these may need to be explicitly deferred or reordered.
- KeyDecisionLog remains accurate relative to roadmap — no re-check needed unless new architectural decisions are logged.
