---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-09-11
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
- **Files Read:** 6 — `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`, `station/agent/Routines/roadmap-accuracy.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard row), `station/Logs/RoutineLog.md` (new entry)
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `Roadmap.md` and compared against `Status.md` and `RoutineLog.md`.

**Phase 1 — Foundation & Polish:** All 11 checkboxes are marked `[x]` and match what has been shipped. The `bonsai validate` row and the annotated "Better trigger sections" row were both added correctly by the 2026-05-07 routine digest. Phase 1 is complete and accurate.

**Phase 2 — Extensibility:** "Custom item detection" is correctly marked `[x]`. Three items remain unchecked (`Self-update mechanism`, `Template variables expansion`, `Micro-task fast path`). These are still valid future goals. However, **Plan 41 (Headless CLI Contract, shipped 2026-06-16) is a significant Phase 2-level capability with no roadmap row.** Plan 41 delivered: `*Result` headless cores for all four mutating commands, JSONL/exit-code contract (`ExitConflict=5`), `list --json`, and `docs/agent-interface.md`. This is the machine-readable interface layer that enables automation and MCP integration — clearly an extensibility milestone.

**Phase 3 — Cloud & Orchestration:** Both items remain future/unchecked. **Plan 42 (MCP server, `bonsai mcp` via go-sdk stdio) is the named next initiative** (referenced in memory.md Work State) and is not represented on the roadmap. Phase 3 currently references "Claude's Managed Agents platform" as the integration target; the actual implementation trajectory is MCP server first. This may require a Phase 3 rewrite or a new item prepended before Managed Agents.

**Phase 4 — Ecosystem:** No changes. All three items remain future and are still the right long-term goals.

### Step 2 — Check milestone accuracy

- **Plan 41 (Headless CLI)** shipped June 2026 — not on roadmap, no tracking row. This should be added to Phase 2 as a completed item, or acknowledged as a prerequisite layer that Phase 3 builds on.
- **Plan 42 (MCP server)** is the live next priority. Absent from the roadmap entirely. Should appear under Phase 3.
- **Non-interactive mode** (`--non-interactive --from-config`, Plan 39, v0.4.2, shipped May 2026) — also not on roadmap. Less critical to represent, but it enabled the Bonsai-Eval bootstrap (Plan 38) — a meaningful automation capability.
- **`bonsai completion`** (external PR #78, May 2026) — minor capability addition, not roadmap-worthy on its own.
- **Template variables expansion** — flagged by Backlog Hygiene (2026-09-11) as a Phase 2 milestone with no Backlog entry. Roadmap has it as `[ ]`, which is correct, but it also has no tracking in the Backlog. Dual-gap.

### Step 3 — Cross-check against Key Decision Log

- "Bonsai is a scaffolding tool, not a runtime orchestrator" — still holds. Plan 41 adds a headless API layer but Bonsai still generates files and steps away. No conflict.
- "Defer Managed Agents cloud integration until local foundation is stable" — the headless contract (Plan 41) and MCP server (Plan 42) represent that the foundation is now stable enough to advance toward Phase 3. The decision to defer still makes sense for the Managed Agents / cloud deployment milestone specifically.
- No decisions in the KeyDecisionLog explicitly invalidate any current roadmap items.

### Step 4 — Report findings

See Findings Summary below. No Roadmap.md edits made — all findings flagged for user review per procedure.

### Step 5 — Update dashboard

Updated `station/agent/Core/routines.md` Roadmap Accuracy row: Last Ran → 2026-09-11, Next Due → 2026-09-25, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | Plan 41 (Headless CLI Contract, shipped 2026-06-16) has no roadmap row — a major extensibility milestone is invisible | `Roadmap.md` Phase 2 | Flagged for user review — recommend adding `[x] Headless CLI contract — machine-readable *Result cores + JSONL/exit-code API (docs/agent-interface.md)` to Phase 2 |
| 2 | Medium | Plan 42 (MCP server, `bonsai mcp`) is the live next priority and is absent from the roadmap | `Roadmap.md` Phase 3 | Flagged for user review — recommend adding `[ ] MCP server — stdio `bonsai mcp` command (go-sdk) for agent-to-tool integration` to Phase 3, ahead of Managed Agents |
| 3 | Medium | Phase 3 references "Claude's Managed Agents platform" but the actual integration trajectory is MCP server first; "Managed Agents" may no longer be the primary Phase 3 target | `Roadmap.md` Phase 3 | Flagged for user review — may want to reword or reorder Phase 3 milestones to reflect MCP-first approach |
| 4 | Low | Non-interactive mode (`--non-interactive --from-config`, Plan 39, v0.4.2) is a shipped automation capability not represented in the roadmap | `Roadmap.md` Phase 2 | Flagged for user review — lower priority; could be rolled into the Plan 41 headless row or noted separately |
| 5 | Low | "Template variables expansion" (Phase 2, unchecked) also has no Backlog tracking entry, per 2026-09-11 Backlog Hygiene findings | `Roadmap.md` Phase 2 + `Backlog.md` | Flagged for user review — Backlog Hygiene already surfaced this; confirm dual tracking is desired |

---

## Errors & Warnings

None.

---

## Items Flagged for User Review

1. **[High] Add Plan 41 to Roadmap Phase 2** — "Headless CLI contract" should be a `[x]` completed item in Phase 2. Suggested wording: `[x] Headless CLI contract — *Result cores for all commands + JSONL/exit-code API; docs/agent-interface.md (Plan 41, v0.5.x)`.

2. **[Medium] Add MCP server to Roadmap Phase 3** — Plan 42 is the live next initiative. Suggest prepending to Phase 3: `[ ] MCP server — stdio \`bonsai mcp\` via go-sdk for agent-tool integration (Plan 42)`.

3. **[Medium] Revisit Phase 3 "Managed Agents" wording** — The platform may have shifted. If MCP is now the primary cloud integration path, "Managed Agents integration — `bonsai deploy`, session management, outcome rubrics" may need to be reworded, reordered, or scoped differently.

4. **[Low] Non-interactive mode (`--non-interactive`)** — Consider adding to Phase 2 (or folding into the Plan 41 row). Shipped May 2026, enabled the Bonsai-Eval bootstrap.

5. **[Low] Template variables expansion tracking gap** — Phase 2 item has no Backlog entry. Decide whether to add it or treat Roadmap as the sole tracker.

---

## Notes for Next Run

- Phase 1 is fully done and stable — no need to re-audit next cycle.
- Primary focus next run: verify Plan 42 (MCP server) has either shipped or been represented in the Roadmap, and confirm Phase 3 wording has been updated to reflect the MCP-first trajectory.
- If the user adds a new phase or reorganizes phases between now and next run, re-read Roadmap.md fresh before comparing.
