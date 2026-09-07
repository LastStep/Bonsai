---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-09-07
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
- **Files Read:** 6 — `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/agent/Core/memory.md`, `station/Playbook/Backlog.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Write, Edit, Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

**Step 1 — Compare Roadmap against current state:**
Read `Roadmap.md`. Phase 1 has all items marked `[x]` — fully complete. The roadmap still labels Phase 1 "## Current Phase" despite all items being done. Phase 2 has one item completed (`[x] Custom item detection`) and three still open. Cross-checked Phase 1 items against Status.md and confirmed accuracy — both flags from the previous run (2026-05-07) were resolved: "Better trigger sections" is now `[x]` with annotation, and `bonsai validate` was added as a Phase 1 item.

**Step 2 — Check milestone accuracy:**
Compared roadmap items against Status.md "Recently Done" for the period 2026-05-07 to 2026-09-07. Identified three significant shipped or in-progress deliverables that have no roadmap representation: (1) Plan 41 headless CLI contract, (2) Plan 42 MCP server (next priority per memory.md), (3) Plan 40 Odysseus hub schema phases 1–3. Also noted `bonsai completion` command (community contribution, shipped 2026-05-07) is absent from the roadmap — low-severity since Phase 1 is already complete. Phase 2 future items (self-update mechanism, template variables expansion, micro-task fast path) show no progress and remain correctly un-checked.

**Step 3 — Cross-check against Key Decision Log:**
Read `KeyDecisionLog.md`. The Settled decision "Defer Managed Agents cloud integration until local foundation is stable" (2026-04-02) was noted. Plan 41 has now shipped headless MCP-ready cores — the local foundation is materially more stable than when the deferral was logged. The decision is not invalidated (Phase 3 is still future), but may be worth the user revisiting whether the deferral condition has been met. No other key decisions invalidate current roadmap items.

**Step 4 — Report findings:**
Five findings identified; none trigger a direct edit to `Roadmap.md`. All flagged for user review below.

**Step 5 — Update dashboard:**
Updated `agent/Core/routines.md` Roadmap Accuracy row: `Last Ran → 2026-09-07`, `Next Due → 2026-09-21`, `Status → done`.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | Phase 1 is fully complete but still labeled "## Current Phase" — stale heading | `Roadmap.md` L16 | Flagged for user; no edit made |
| 2 | MEDIUM | Plan 41 headless CLI contract shipped 2026-06-16 (agent-drivable `*Result` cores, JSONL/exit contract, `agent-interface.md`) — no roadmap item exists for it | `Roadmap.md` Phase 2 | Flagged for user; suggest adding `[x]` item to Phase 2 |
| 3 | MEDIUM | Plan 42 MCP server (`bonsai mcp`, go-sdk, stdio) is explicitly next priority in memory.md but absent from Roadmap.md — Phase 3 has "Managed Agents integration" but no MCP server step | `Roadmap.md` Phase 3 | Flagged for user; suggest adding an MCP server milestone |
| 4 | LOW | Plan 40 Phases 1–3 shipped a `.bonsai/project.yaml` hub-facing schema freeze (frozen v1 schemas, root-relative scaffolding, project-level validate with path hardening) — new concept with no roadmap entry | `Roadmap.md` Phase 2/3 | Flagged for user; may belong in Phase 2 or as a Phase 3 prerequisite |
| 5 | LOW | `bonsai completion [bash|zsh|fish|powershell]` command shipped 2026-05-07 (community contribution, PR #78) — not on roadmap | `Roadmap.md` Phase 1 | Flagged for user; Phase 1 already complete so low urgency |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**[1] MEDIUM — Phase 1 heading is stale**
`Roadmap.md` labels Phase 1 "## Current Phase" but all 11 items are `[x]`. Phase 2 is the active phase (1 of 4 items done). Suggest renaming Phase 1 to "## Phase 1 — Foundation & Polish (Complete)" and Phase 2 to "## Current Phase". No functional impact, but misleading for any agent or user reading the roadmap at a glance.

**[2] MEDIUM — Plan 41 headless CLI contract not on roadmap**
Plan 41 shipped 2026-06-16 (Status.md): `*Result` headless cores for init/add/update/remove, `list --json`, `docs/agent-interface.md` machine contract, `ExitConflict=5`. This is a Phase 2 capability (enables agent-driven Bonsai / MCP readiness) with no roadmap entry. Suggest adding under Phase 2:
```
- [x] Headless CLI contract — *Result cores, JSONL/exit codes, agent-interface.md; enables MCP integration
```

**[3] MEDIUM — Plan 42 MCP server absent from roadmap**
memory.md explicitly lists MCP server as the next work item: "MCP server = Plan 42 (go-sdk, stdio `bonsai mcp`) — the contract was built for this." Phase 3 has "Managed Agents integration — `bonsai deploy`, session management, outcome rubrics" but no MCP server item. An MCP server (`bonsai mcp` stdio) is a distinct, concrete deliverable that makes Bonsai drivable from any MCP-compatible orchestrator. Suggest adding to Phase 3:
```
- [ ] MCP server — `bonsai mcp` (go-sdk, stdio transport); exposes init/add/update/remove/validate via MCP protocol
```

**[4] LOW — Plan 40 Odysseus hub schema not on roadmap**
Plan 40 Phases 1–3 (shipped 2026-06-13) introduced: frozen v1 `.bonsai/project.yaml` schema (hub-facing name/slug/description), root-relative scaffolding manifest, project-level validate with adversarial path/symlink hardening. Phase 4 is HELD. The `.bonsai/project.yaml` concept is infrastructure for a future hub/platform but has no roadmap entry. Likely belongs as a Phase 2 completed item or a Phase 3 prerequisite. Backlog has a related item: "bonsai validate warn on .bonsai/project.yaml ↔ .bonsai.yaml identity drift." User to decide placement.

**[5] LOW — `bonsai completion` not on roadmap**
Community-contributed shell completion (bash/zsh/fish/powershell) shipped 2026-05-07 (PR #78). Phase 1 is fully complete so this doesn't need a retroactive checkbox, but if the user wants the roadmap to be a complete historical record, it could be added. Otherwise no action needed.

## Notes for Next Run

- Previous run flags (2026-05-07): both were resolved before this run. "Better trigger sections" is now `[x]` in roadmap; `bonsai validate` row added to Phase 1. Clean slate for this run.
- If user adds Plan 41 / Plan 42 roadmap entries before next run, Finding #2 and #3 will resolve automatically.
- Plan 40 Phase 4 is still HELD — if it ships before next run, Finding #4 may need a roadmap entry update.
- The "Defer Managed Agents cloud integration" decision in KeyDecisionLog.md may be worth revisiting now that headless cores are shipped (Plan 41). Flag at next planning session if MCP server (Plan 42) is in flight.
- Plans 40 and 41 remain in `Plans/Active/` — memory.md flags this. Suggest archiving both at next wrap-up session.
