---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-06-30
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
- **Duration:** ~8 min
- **Files Read:** 6 — `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/41-headless-cli-contract.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/40-odysseus-platform-integration.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Write, Edit, Glob, Bash
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state
- **Action:** Read `station/Playbook/Roadmap.md` and cross-checked all Phase 1 checkboxes against Status.md recently-done entries and archived log entries.
- **Result:** Phase 1 (Foundation & Polish) is fully complete — all 11 items are `[x]`. The two quick-fixes applied by the 2026-05-07 routine-digest ("Better trigger sections" annotated `[x]`, `bonsai validate` row added) are intact and accurate. The current phase item in Status.md ("Plan 41 SHIPPED 2026-06-16") has no corresponding roadmap row — see Step 2.
- **Issues:** Phase 1 is clean. No stale unchecked boxes or spuriously-checked boxes found.

### Step 2: Check milestone accuracy (Phase 2 and beyond)
- **Action:** Reviewed Phase 2 (Extensibility), Phase 3 (Cloud & Orchestration), and Phase 4 (Ecosystem) items against recent work (Plans 40, 41) and Backlog.
- **Result:**
  - **Phase 2 — `[x] Custom item detection`:** Correctly marked done. Accurate.
  - **Phase 2 — `[ ] Self-update mechanism`:** Open. Matches Backlog P3 entry. Accurate.
  - **Phase 2 — `[ ] Template variables expansion`:** Open. **Gap:** Backlog Hygiene (2026-06-30) flagged this has no Backlog tracking entry despite being a named roadmap item. The Backlog contains no entry for it. Low-severity but worth adding a Backlog P3 entry so it doesn't get lost.
  - **Phase 2 — `[ ] Micro-task fast path`:** Open. Matches Backlog P3 entry. Accurate.
  - **Phase 2 — MISSING ROW:** Plan 41 shipped a full "Headless CLI Contract + MCP-ready cores" (all mutating cmds now have pure `*Result` headless cores, JSONL/exit contract). This is a significant Phase 2 extensibility milestone (enabling AI-agent-driven Bonsai headless) with no roadmap entry. It also directly enables the Phase 3 MCP server (Plan 42 described as "fast-follow").
  - **Phase 2 — MISSING ROW:** Plan 40 (Odysseus integration, v0.5.0 shipped Phases 1–3) introduced `.bonsai/project.yaml` manifest, memory graph (`station/Memory/`), and `validate` lint for both. These are Extensibility milestones not tracked in the roadmap.
  - **Phase 3 — `[ ] Managed Agents integration`:** Deferred per KeyDecisionLog (2026-04-02). Still appropriate — local foundation is now more solid. Plan 42 (MCP server, not started) is the next step on this path.
  - **Phase 3 — `[ ] Greenhouse companion app`:** No change. Matches Backlog "Big Bets." Accurate.
  - **Phase 4 items:** No change. All still open. Accurate.
- **Issues:** Two significant new capabilities (Plan 41 headless contract, Plan 40 Odysseus scaffold) have no roadmap rows. Flagged for user review — per procedure, Roadmap.md is not modified directly.

### Step 3: Cross-check against Key Decision Log
- **Action:** Read `station/Logs/KeyDecisionLog.md` in full.
- **Result:** The KeyDecisionLog has no entries after 2026-04-13. Given that Plans 40 and 41 introduced significant architectural decisions (headless contract shape, JSONL event philosophy, MCP-ready core pattern, Odysseus boundary decisions), there are several decisions that should arguably be in the log. However, those decisions were locked in the plan grilling sessions and recorded in the plan files themselves — whether to backfill them into KeyDecisionLog is a user call.
- **Issues:** No KeyDecisionLog decisions were found that invalidate any current roadmap items. The settled decision to "defer Managed Agents cloud integration until local foundation is stable" remains valid and is being honored (Plan 42 not started yet, still gated behind headless foundation now shipping).

### Step 4: Report findings
- **Action:** Compiled all mismatches and gaps. Per procedure, NOT modifying Roadmap.md — all items flagged for user review.
- **Result:** 3 items flagged (see Findings Summary). Dashboard and log updated.
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for Roadmap Accuracy: `Last Ran` → 2026-06-30, `Next Due` → 2026-07-14, `Status` → `done`.
- **Result:** Done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Low | Phase 2 has no row for headless CLI contract (Plan 41) — MCP-ready headless cores for all mutating cmds shipped 2026-06-16; significant extensibility milestone with no roadmap entry | `Roadmap.md` Phase 2 | Flagged for user — recommend adding `[x] Headless CLI contract — MCP-ready headless cores for init/add/update/remove (Plan 41, v0.4.x)` under Phase 2 |
| 2 | Low | Phase 2 has no row for Odysseus scaffold items (Plan 40 Phases 1–3) — `.bonsai/project.yaml` manifest, memory graph scaffolding, validate lint for both shipped as v0.5.0 (untagged) | `Roadmap.md` Phase 2 | Flagged for user — recommend adding a row or annotation under Phase 2, or noting v0.5.0 scope against Phase 3 if Odysseus is considered the first cloud-integration step |
| 3 | Low | `Template variables expansion` (Phase 2 open item) has no Backlog tracking entry — Backlog Hygiene 2026-06-30 flagged this; confirmed by full Backlog read | `Backlog.md` (absent) | Flagged for user — recommend adding a P3 Backlog entry so this roadmap item has proper tracking |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **Roadmap Phase 2 — Add headless CLI contract row** (Finding #1): Plan 41 shipped all mutating-cmd headless cores + JSONL/exit contract. Recommend `[x]` row: `[x] Headless CLI contract + MCP-ready cores — all mutating commands (init/add/update/remove) have pure Result-returning cores, JSONL streaming, and unified exit-code contract (Plan 41)`. This also makes the path to Plan 42 (MCP server) more legible in the roadmap.

2. **Roadmap Phase 2 (or Phase 3) — Add Odysseus/project.yaml manifest row** (Finding #2): Plan 40 Phases 1–3 shipped `.bonsai/project.yaml`, memory graph scaffold, and validate lint. Suggest user decides whether this belongs under Phase 2 Extensibility or Phase 3 Cloud & Orchestration (since Odysseus is the cloud platform). Either way, a row would prevent a future reader from wondering why v0.5.0 exists with no roadmap anchor.

3. **Backlog — Add `Template variables expansion` entry** (Finding #3): Phase 2 roadmap item `[ ] Template variables expansion` has no Backlog entry. Add a P2 or P3 entry like: `[feature] Template variables expansion — richer context available in templates (roadmap Phase 2 item, not yet started). *(added 2026-06-30, source: roadmap-accuracy routine)*`

4. **KeyDecisionLog — Optional backfill**: Plans 40/41 locked several architectural decisions (headless contract shape, JSONL vs single-doc JSON, MCP-ready core pattern, Odysseus boundary). These were recorded in plan files but not in KeyDecisionLog. User may want to backfill the most structural ones (e.g., "JSONL for streaming mutation progress, single-doc JSON for read snapshots" and "Bonsai = schema authority / Odysseus = hub runtime").

## Notes for Next Run

- Phase 1 is stable — no drift expected. Skip detailed Phase 1 re-audit next run.
- Phase 2 has active work (Plan 42 MCP server described as "fast-follow" to Plan 41). Check if Plan 42 has shipped and whether any new Phase 2/3 rows are needed.
- KeyDecisionLog has been static since 2026-04-13 despite 6+ weeks of significant architectural decisions. Prompt user to consider a backfill pass.
- The HOMEBREW_TAP_TOKEN PAT expires ~2026-07-15 (per Backlog P1 and RoutineLog 2026-06-30 Backlog Hygiene) — not a roadmap issue but worth noting in next planning context.
