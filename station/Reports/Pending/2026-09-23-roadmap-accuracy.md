---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-09-23
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
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 2
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard row updated)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (entry appended)
  - `/home/user/Bonsai/station/Reports/Pending/2026-09-23-roadmap-accuracy.md` (this file)
- **Tools Used:** Read, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `Playbook/Roadmap.md` and compared each item against `Playbook/Status.md` and `Logs/RoutineLog.md`.

**Phase 1 — Foundation & Polish:** All items are checked [x]. The 2026-05-07 Routine Digest confirmed two items were updated in that pass ("Better trigger sections" marked [x] with annotation; `bonsai validate` row added as [x]). All Phase 1 items are confirmed shipped. **ACCURATE.**

**Phase 2 — Extensibility:** Custom item detection [x] is correct (Plan 34 shipped). Three items remain unchecked: self-update mechanism, template variables expansion, micro-task fast path. None of these have been shipped. **PARTIALLY ACCURATE** — see findings #1 and #2 below.

**Phase 3 — Cloud & Orchestration:** Both items remain unchecked. KeyDecisionLog explicitly deferred Managed Agents integration ("until local foundation is stable"). **ACCURATE** for existing items, but see finding #3 (MCP server not represented).

**Phase 4 — Ecosystem:** All unchecked, no work started. **ACCURATE.**

### Step 2 — Check milestone accuracy

Cross-checked the last 4 months of Status.md against roadmap phases:

- **Plan 41 (2026-06-16)** shipped a full headless CLI contract: pure `*Result` cores for all mutating commands, JSONL/exit-code contract (ExitConflict=5), `list --json`, and `docs/agent-interface.md`. Status.md notes "MCP server = fast-follow Plan 42." This significant delivered capability has no roadmap representation.

- **Plan 40 (2026-06-13)** shipped frozen v1 schemas + root-relative scaffolding + project-level validate pass (Phases 1-3). Phase 4 (update-delivery for existing projects) was explicitly HELD by user. The Roadmap item "Self-update mechanism" may relate to this deferred Phase 4 work — the relationship is ambiguous.

- **Plan 42 (MCP server):** Referenced as "fast-follow" in Status.md (Plan 41 entry). No roadmap row exists for it. Phase 3 mentions "Managed Agents integration — `bonsai deploy`" but an MCP server is architecturally distinct.

- **`bonsai completion` command** (PR #78, 2026-05-07): New CLI command from external contributor. Not a major feature — fine to leave off roadmap.

- **Bonsai-Eval offshoot** (Plan 38, 2026-05-13): Separate repository bootstrapped. Not on roadmap, but it's a derivative project, not a core Bonsai feature.

No roadmap items appear to reference deprecated or superseded approaches.

### Step 3 — Cross-check against Key Decision Log

Read `station/Logs/KeyDecisionLog.md`. Last entry dated 2026-04-13 — the log has not been updated in ~5 months despite significant architectural work (headless CLI contract, MCP-ready interfaces, frozen v1 schemas, worktree isolation pattern). No decisions in the log invalidate existing roadmap items. The MCP-readiness direction taken in Plan 41 is architecturally significant and not recorded.

### Step 4 — Report findings

Four items flagged for user review below. Roadmap.md is NOT modified per procedure (flag-only).

### Step 5 — Dashboard + log updated

Dashboard row for Roadmap Accuracy set to `Last Ran: 2026-09-23`, `Next Due: 2026-10-07`, `Status: done`. RoutineLog.md entry appended.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | Plan 41 headless CLI + MCP-ready cores shipped with no roadmap representation | `Playbook/Roadmap.md` Phase 2/3 gap | Flagged for user — recommend adding [x] row to Phase 2 or Phase 3 |
| 2 | Medium | "Self-update mechanism" (Phase 2) may be stale/superseded — Plan 40 Phase 4 (update-delivery) was HELD; relationship to this item is ambiguous | `Playbook/Roadmap.md` Phase 2 | Flagged for user — clarify whether Plan 40 P4 satisfies or replaces this item |
| 3 | Medium | MCP server (Plan 42, imminent fast-follow) has no roadmap row | `Playbook/Roadmap.md` Phase 3 | Flagged for user — recommend adding [ ] row under Phase 3 |
| 4 | Low | KeyDecisionLog not updated since 2026-04-13 — 5 months of architectural decisions (headless CLI, MCP-readiness, frozen v1 schemas) not recorded | `station/Logs/KeyDecisionLog.md` | Flagged for user |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **[HIGH] Add roadmap row for Plan 41 headless CLI contract**: Plan 41 (merged 2026-06-16, PRs #120–#125) delivered a full headless-CLI + MCP-ready interface: pure `*Result` cores for all mutating commands, JSONL/exit-code contract, `list --json`, and `docs/agent-interface.md`. This is a v0.5.x headline capability with no Phase 2 or Phase 3 row. Recommend: add `[x] Headless CLI + agent interface — pure \*Result cores, JSONL contract, MCP-ready (Plan 41)` under Phase 2 (or Phase 3 as a Cloud precursor).

- **[MEDIUM] Clarify "Self-update mechanism" vs Plan 40 Phase 4**: Plan 40 Phase 4 (update-delivery for existing projects via `bonsai init --non-interactive`) was explicitly HELD and moved to Backlog P1 (superseded by headless-CLI workstream). The Phase 2 roadmap item "Self-update mechanism — catalog items can flag when they're stale or have issues" may be the same concept or a different one (per-item staleness detection vs full update-delivery). User should either: (a) reword the item to disambiguate, (b) mark it superseded by Plan 41/42, or (c) confirm it means something different.

- **[MEDIUM] Add MCP server row to Phase 3**: Status.md (Plan 41 entry) explicitly notes "MCP server = fast-follow Plan 42." The MCP server is architecturally distinct from "Managed Agents integration" (which is about `bonsai deploy` + session management). Recommend adding `[ ] MCP server — expose Bonsai commands as MCP tools for Claude Code and other MCP clients (Plan 42)` under Phase 3.

- **[LOW] Update KeyDecisionLog**: No entries since 2026-04-13 despite: headless CLI contract pattern (all mutating commands → pure `*Result` + JSONL), MCP-readiness as a design target, frozen v1 schemas for scaffolding manifest + memory, worktree-isolation-leak pattern (recurring gotcha ~10×). These belong in the Structural or Domain-Specific sections.

---

## Notes for Next Run

- If Plan 42 (MCP server) ships before next run, verify it's checked [x] on the roadmap.
- If Plan 40 Phase 4 (update-delivery) ships before next run, the "Self-update mechanism" ambiguity should be resolved by then.
- KeyDecisionLog staleness is a recurring gap — consider adding a check to the session-start workflow or doc-freshness-check routine to flag it explicitly.
- Phase 1 and Phase 2 (custom item detection) are stable — no re-verification needed next cycle.
