---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-07-22
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
- **Files Read:** 7 — `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`, `station/Playbook/Backlog.md`, `station/Playbook/Plans/Active/41-headless-cli-contract.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Write, Edit, Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `Roadmap.md` and compared each item against the current codebase state and Status.md history.

**Phase 1 — Foundation & Polish:** All 11 items correctly marked `[x]`. The two items fixed by the 2026-05-07 routine-digest ("Better trigger sections" annotation + `bonsai validate` row) are present and accurate.

**Phase 2 — Extensibility:**
- `[x] Custom item detection` — correctly marked done (scan.go, shipped well before last run)
- `[ ] Self-update mechanism` — correctly unchecked; item is in Backlog P3 "Future Platform"
- `[ ] Template variables expansion` — correctly unchecked; **however, no Backlog entry exists to track this** (also flagged by Backlog Hygiene 2026-07-22)
- `[ ] Micro-task fast path` — correctly unchecked; item is in Backlog P3

**Phase 3 — Cloud & Orchestration:** Both items (`bonsai deploy` / Managed Agents + Greenhouse companion app) correctly unchecked.

**Phase 4 — Ecosystem:** All items correctly unchecked.

### Step 2 — Check milestone accuracy

Since the last roadmap-accuracy run (2026-05-07), two significant plans shipped that are not reflected anywhere in the roadmap:

**Plan 41 — Headless CLI Contract + MCP-ready cores (shipped 2026-06-16)**
Plan 41 delivered: all 4 mutating commands (`init`, `add`, `update`, `remove`) have pure `*Result`-returning headless cores; JSONL/exit-code contract with `ExitConflict=5`; `list --json`; `docs/agent-interface.md` contract doc. The plan explicitly states: _"future `bonsai mcp` server (Plan 42) is a thin wrapper calling the same functions — zero duplicated work between the CLI and MCP layers."_

This is a concrete, shippable milestone — not just implementation detail. It arguably belongs in Phase 2 (Extensibility — enabling programmatic / agent-driven access to all commands) or as a Phase 3 precursor row. Currently the roadmap skips from "custom item detection" directly to "self-update mechanism" with no acknowledgment of the headless contract work.

**Plan 40 — Odysseus Platform Integration Phases 1-3 (shipped 2026-06-13, Phase 4 held)**
Phases 1-3 delivered frozen v1 schemas, root-relative scaffolding, `validate` pass hardening, and memory-routing docs. Phase 4 (update-delivery) was held and superseded by Plan 41. This work is internal platform plumbing and does not cleanly map to a roadmap row on its own, but it is part of the Phase 3 "Managed Agents integration" groundwork.

**Next milestone accuracy:** The most concrete next milestone (not in roadmap) is Plan 42 = `bonsai mcp` server. With Plan 41 shipped and Plan 42 as the stated fast-follow, the actual trajectory is:
- Phase 2 remnants (template variables expansion, self-update, micro-task fast path) — no active work
- Plan 42 (MCP server) — not yet started, no plan file exists
- Phase 3 cloud/orchestration proper — longer horizon

### Step 3 — Cross-check against Key Decision Log

Reviewed `station/Logs/KeyDecisionLog.md`. Findings:

1. **No new entries since 2026-04-13** — 14 months of significant work (Plans 28-41, v0.2.0-v0.5.0, headless CLI contract, platform integration) without any new architectural decisions logged. Several decisions from Plan 41 in particular are structurally significant:
   - "Layered, not parallel" architecture: MCP is a thin wrapper calling the same headless core (proven by research, this shapes all future MCP/cloud work)
   - JSONL for streaming mutations, single-doc JSON for read snapshots (two-serialization philosophy)
   
2. **Stale file reference in KeyDecisionLog:** The Settled section (line 55) includes: _"See DESIGN-companion-app.md for Greenhouse design."_ This file does not exist anywhere in the repository (confirmed via Glob). The decision itself (defer Managed Agents) is still valid — only the file reference is stale.

3. All other decisions remain accurate and aligned with the current codebase.

### Step 4 — Report findings (no Roadmap.md changes)

Findings documented below. Per procedure, Roadmap.md is not modified directly.

### Step 5 — Update dashboard

Dashboard row updated in `station/agent/Core/routines.md`.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | Plan 41 (Headless CLI Contract + MCP-ready cores, shipped 2026-06-16) has no Roadmap row — meaningful Phase 2/3 progress untracked. MCP server (Plan 42) is explicitly the next step but doesn't appear anywhere in the Roadmap. | `Roadmap.md` Phase 2 / Phase 3 | Flagged for user decision — suggest adding a `[x]` row to Phase 2 or 3 |
| 2 | Low | Phase 2 item "Template variables expansion" has no Backlog tracking entry — no path from Roadmap item to planned work. | `Roadmap.md` Phase 2, `Backlog.md` P3 "Future Platform" | Flagged; also raised by Backlog Hygiene (2026-07-22) — add Backlog entry |
| 3 | Low | `DESIGN-companion-app.md` referenced in `KeyDecisionLog.md` (Settled section, line 55) does not exist in the repository. | `station/Logs/KeyDecisionLog.md` | Flagged; update or remove the stale file reference |
| 4 | Low | `KeyDecisionLog.md` has no entries since 2026-04-13 despite Plans 28–41 and v0.2.0–v0.5.0 shipping. Plan 41's layered MCP architecture decision is particularly worth logging. | `station/Logs/KeyDecisionLog.md` | Flagged for user decision on whether to add a Structural entry |
| 5 | Info | Plans 40 and 41 both remain in `Plans/Active/` despite being shipped — already flagged by Memory Consolidation and Backlog Hygiene (2026-07-22). | `station/Playbook/Plans/Active/` | No action taken (housekeeping; already tracked by other routines) |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **[Medium] Add Plan 41 milestone to Roadmap** — Suggest inserting a `[x]` row in Phase 2 or Phase 3:
   ```
   - [x] Headless CLI contract — all mutating cmds have *Result cores + JSONL/exit protocol; list --json; agent-interface.md. MCP server (Plan 42) is the fast-follow. (Plan 41, v0.5+)
   ```
   Or alternatively, add it as a Phase 3 precursor row under "Cloud & Orchestration."

2. **[Low] Add Backlog entry for Phase 2 "Template variables expansion"** — currently a roadmap item with no Backlog tracking. Without a P3 entry, it has no path to being prioritized or worked.

3. **[Low] Fix stale file reference in KeyDecisionLog** — Update the Settled entry for "Defer Managed Agents cloud integration" to remove or replace the `DESIGN-companion-app.md` reference (file does not exist). If a companion app design still exists in some form (e.g., notes, external doc), add the correct pointer; if abandoned, note that.

4. **[Low] Consider a KeyDecisionLog entry for Plan 41's layered MCP architecture** — Structural impact: _"MCP is a thin wrapper that calls the same non-interactive core, not a second implementation."_ This shapes all future MCP/cloud work and is worth preserving in the decision log.

---

## Notes for Next Run

- Plans 40 and 41 may be archived by then (flagged by 3 routines today); confirm archival
- Plan 42 (MCP server) may be active — if so, verify Phase 3 Roadmap row reflects it
- Check whether KeyDecisionLog has been updated; if still stale after 2 cycles, escalate
- "Template variables expansion" Backlog gap: verify a tracking entry was added
