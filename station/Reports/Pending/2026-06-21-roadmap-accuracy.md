---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-06-21
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
- **Files Read:** 6 — `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Backlog.md`, `station/Playbook/Plans/Active/40-odysseus-platform-integration.md`, `station/Playbook/Plans/Active/41-headless-cli-contract.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state
- **Action:** Read `Roadmap.md` and cross-checked its phases against `Status.md` recently-done rows and `Plans/Active/` contents.
- **Result:**
  - Phase 1 is fully marked `[x]` — this is accurate. All items shipped per Status.md records.
  - Phase 2 has one `[x]` (Custom item detection) and three `[ ]` items — all `[ ]` items are correctly unchecked (still future work, per Backlog P2/P3).
  - However, two significant capabilities shipped since the last run (2026-05-07) that are not reflected in the Roadmap:
    - **Plan 40 (v0.5.0 Odysseus Integration)** — shipped frozen v1 schemas, `.bonsai/project.yaml` manifest scaffolding, and `validate` lint for both. Phases 1–3 merged on main. Not a roadmap item anywhere.
    - **Plan 41 (Headless CLI Contract)** — shipped pure `*Result` headless cores for all mutating commands, JSONL/exit contract, `list --json`, and `docs/agent-interface.md`. Shipped 2026-06-16 (PRs #120/#122/#123/#121/#125). Not on the Roadmap.
  - The `bonsai validate` item added to Phase 1 in the 2026-05-07 digest is correctly `[x]` in the Roadmap.
- **Issues:** Two shipped milestones absent from roadmap. Phase 2 structure may need a new item for headless/MCP-ready architecture.

### Step 2: Check milestone accuracy
- **Action:** Reviewed Phase 2 remaining `[ ]` items against Backlog and Status.
- **Result:**
  - **Self-update mechanism** (`[ ]`) — Backlog P3, no active plan. Correctly unchecked.
  - **Template variables expansion** (`[ ]`) — No active plan or Backlog item. Correctly unchecked.
  - **Micro-task fast path** (`[ ]`) — Backlog P3 (`improvement`). Correctly unchecked.
  - **New Phase 2 candidate:** Plan 41's headless CLI contract is a direct enabler of the MCP server (Plan 42, planned but not started) and the Managed Agents integration (Phase 3). The Roadmap's Phase 2 "Extensibility" milestone would benefit from a row for "Headless CLI / MCP-ready cores" — shipped and foundational for Phase 3.
  - **New Phase 3 candidate:** `bonsai mcp` server (Plan 42, not started) is a planned fast-follow to Plan 41 — sits between Phase 2 (headless cores) and Phase 3 (Managed Agents). The Roadmap's Phase 3 currently jumps straight to "`bonsai deploy`" without acknowledging the MCP layer.
- **Issues:** Roadmap Phase 2 missing a shipped item (headless CLI cores); Phase 3 narrative skips the MCP intermediary layer.

### Step 3: Cross-check against Key Decision Log
- **Action:** Read `station/Logs/KeyDecisionLog.md` and compared against Roadmap items and recent plans.
- **Result:**
  - No entries in KeyDecisionLog postdate 2026-04-13 — the log was last updated during initial architecture decisions and has not been augmented with decisions from Plans 40 or 41.
  - Key decisions from Plan 40 (Odysseus integration) and Plan 41 (headless contract) were not added to the log — these include significant architectural decisions: frozen v1 schemas as Bonsai standard, exit-code/JSONL output contract, MCP-ready core shape. These should be in the log per the decision-logging protocol.
  - The 2026-04-02 settled decision "Defer Managed Agents cloud integration until local foundation is stable" remains accurate — Plan 41 is now the foundation-stabilization step that precedes it.
  - Phase 3's "Managed Agents integration" rationale in KeyDecisionLog references "see DESIGN-companion-app.md for Greenhouse design" — this file is referenced but its location isn't clear from the station workspace; the connection to Plan 42 (MCP) as an intermediary step is not captured.
- **Issues:** KeyDecisionLog not updated with Plan 40/41 decisions. MCP intermediary layer not acknowledged in settled decisions.

### Step 4: Report findings
- **Action:** Compiled all findings (see below). No modifications to `Roadmap.md` — flagging for user review per procedure.
- **Result:** 5 findings identified; all flagged for user review. Roadmap is healthy for Phase 1/2/3 at a high level, but has two gaps from recent shipped work.
- **Issues:** None — procedure followed correctly.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Roadmap Accuracy row.
- **Result:** `Last Ran` → 2026-06-21, `Next Due` → 2026-07-05, `Status` → `done`.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | Plan 41 (Headless CLI Contract) shipped 2026-06-16 — no Roadmap row exists for this capability | `Roadmap.md` Phase 2 | Flagged for user review |
| 2 | Low | Plan 40 (Odysseus Integration / v0.5.0) shipped manifest + schema — not tracked on Roadmap | `Roadmap.md` Phase 2 | Flagged for user review |
| 3 | Low | Phase 3 Roadmap skips MCP intermediary layer (Plan 42, planned) between headless cores and Managed Agents | `Roadmap.md` Phase 3 | Flagged for user review |
| 4 | Low | KeyDecisionLog has no entries since 2026-04-13 — Plan 40 and Plan 41 architectural decisions not recorded | `station/Logs/KeyDecisionLog.md` | Flagged for user review |
| 5 | Info | Plan 41 is still in `Plans/Active/` despite shipping 2026-06-16 (also flagged by Memory Consolidation 2026-06-21) | `Plans/Active/41-headless-cli-contract.md` | Flagged for user review (pre-existing flag) |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[Medium] Roadmap Phase 2 — add a row for Headless CLI / MCP-ready cores (Plan 41, shipped).** Suggested text: `- [x] Headless CLI contract — pure *Result cores, JSONL/exit-code contract, agent-interface.md; MCP-ready surface (Plan 42 fast-follow)` — this is Phase 2 Extensibility work that shipped and should be acknowledged.

2. **[Low] Roadmap Phase 2 — optionally add a row for Odysseus/hub integration (Plan 40, v0.5.0, shipped).** Covers frozen v1 schemas + `.bonsai/project.yaml` manifest + validate lint. If Odysseus is user-private context (not public roadmap material), skip this one.

3. **[Low] Roadmap Phase 3 — add a `[ ] bonsai mcp server — Plan 42 fast-follow` row between Phase 2 and the "Managed Agents integration" row.** The MCP layer is the bridge between headless cores (Plan 41) and the Managed Agents cloud platform (Phase 3). Without it, the Phase 3 item implies a bigger jump than the actual incremental plan.

4. **[Low] KeyDecisionLog — add entries for Plan 40 and Plan 41 architectural decisions.** Minimum: (a) headless CLI contract shape (JSONL streaming vs single-doc JSON, exit code 5 = conflict, `*Result` core pattern); (b) MCP-ready core philosophy (CLI = thin wrapper over same cores); (c) frozen v1 schemas as Bonsai standard (project.yaml + memory graph).

5. **[Info] Archive Plan 41 from `Plans/Active/` to `Plans/Archive/`.** Plan shipped 2026-06-16, all phases merged. Also flagged by Memory Consolidation routine 2026-06-21 — low urgency.

## Notes for Next Run

- Watch for Plan 42 (MCP server) shipping — when it does, Phase 3 "Managed Agents integration" item should be annotated with the MCP prerequisite.
- If KeyDecisionLog is updated between now and next run, verify that Plan 40/41 decisions were captured.
- Phase 2 remaining items (self-update, template vars, micro-task fast path) should be revisited for prioritization — no active plans and no Backlog promotion in 2+ months suggests they may be deprioritized in favor of Phase 3 work.
- If v0.5.0 tag is cut before next run, confirm Roadmap Phase 2 items are updated accordingly.
