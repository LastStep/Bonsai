---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-07-12
status: partial
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (2 items flagged requiring user action; audit-only routine with no auto-fixes applied)
- **Duration:** ~5 min
- **Files Read:** 5
  - `/home/user/Bonsai/station/agent/Routines/roadmap-accuracy.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard updated)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (entry appended)
- **Tools Used:** Read, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `station/Playbook/Roadmap.md` and assessed phase status against actual shipped work.

**Phase 1 — Foundation & Polish:** All 11 items are marked `[x]` (including "Better trigger sections" and "`bonsai validate`" which were fixed by the 2026-05-07 Routine Digest). Phase 1 is 100% complete.

**However:** The Roadmap still uses `## Current Phase` as the section header for Phase 1. With every item shipped, this heading is now misleading — Phase 1 is done, not current.

**Phase 2 — Extensibility:** One item is checked (`[x] Custom item detection`). Three items remain open: self-update mechanism, template variables expansion, micro-task fast path. No active plans for any of these are visible in `Status.md`.

**Phase 2 gap identified:** Plan 41 (Headless CLI Contract + MCP-ready cores, shipped 2026-06-16) introduced pure `*Result` headless cores, JSONL/exit-code contract, `list --json`, and `docs/agent-interface.md`. This is a significant new capability — a machine-readable API surface for agent callers — that does not appear in any roadmap phase. It fits squarely in Phase 2 (Extensibility) or as a Phase 3 (Cloud & Orchestration) enabler.

**Status.md alignment check:** Current In Progress = none. Recently Done shows Plan 41 (2026-06-16) and Plan 40 Phases 1–3 (2026-06-13). No active phase is marked current in Roadmap. This is accurate — there's no in-flight milestone right now.

### Step 2 — Check milestone accuracy

**Are next milestones still the right priority?** Phase 2 open items (self-update, template vars, micro-task fast path) remain plausible next targets. However:
- "Micro-task fast path — lightweight protocol bypass for trivial changes" may be partially superseded by Plan 41's headless CLI contract, which already lets agents call Bonsai non-interactively. Whether the original intent differs from what Plan 41 delivered is unclear — flagged for user clarification.

**Has any planned work been superseded?** No Phase 2 or Phase 3 items appear actively invalidated by recent decisions or shipped work. Phase 3 Managed Agents integration and Greenhouse companion app remain future and untouched.

**Deprecated approaches referenced?** No roadmap items reference approaches that have since been deprecated.

### Step 3 — Cross-check against Key Decision Log

Read `station/Logs/KeyDecisionLog.md`. All entries date from 2026-04-02 to 2026-04-13. No new decisions have been logged since. Key observation:

- "Defer Managed Agents cloud integration until local foundation is stable" (2026-04-02) — still in effect. Phase 3 cloud items correctly remain in "Future Phases."
- No decisions in the log invalidate any current roadmap items.

**Side observation (out of scope for this routine):** Plans 40 and 41 made significant architectural decisions (memory routing, headless API contract shape, MCP-ready core design) that are not captured in `KeyDecisionLog.md`. The log has not been updated in ~90 days. This is a separate concern but worth noting.

### Step 4 — Report findings

Per procedure: do not modify `Roadmap.md` directly. All findings are flagged for user review below.

### Step 5 — Update dashboard

Updated `routines.md` dashboard: `Last Ran` → 2026-07-12, `Next Due` → 2026-07-26, `Status` → `done`.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | Phase 1 is 100% complete but Roadmap still uses `## Current Phase` as its section header — misleading now that nothing in Phase 1 is open. Should be reorganized: Phase 1 → "Completed", Phase 2 promoted to "Current Phase" (or a "Between Phases" note if no active milestone). | `Roadmap.md` lines 14–31 | Flagged for user — no auto-edit per procedure |
| 2 | LOW | Plan 41 (Headless CLI Contract, shipped 2026-06-16) — `*Result` headless cores, JSONL/exit-code contract, `list --json`, `docs/agent-interface.md` — is not tracked in any roadmap phase. A new checked `[x]` row should be added (likely Phase 2 or Phase 3). | `Roadmap.md` Phase 2 or Phase 3 section | Flagged for user — no auto-edit per procedure |
| 3 | LOW | "Micro-task fast path" (Phase 2, open) may overlap in intent with Plan 41's headless CLI contract (which enables agent-driven non-interactive calls). Needs user clarification: Is this item still distinct, or should it be closed/reworded? | `Roadmap.md` line 43 | Flagged for user |
| 4 | INFO | `bonsai completion` command (external contribution, PR #78, 2026-05-07) not in any roadmap phase — acceptable omission since it was an unplanned community contribution, not a milestone. | N/A | No action needed |
| 5 | INFO | `KeyDecisionLog.md` has not been updated since 2026-04-13 (~90 days). Plans 40 and 41 both made architectural decisions not captured there. | `station/Logs/KeyDecisionLog.md` | Out of scope for this routine — noting for awareness |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **MEDIUM — Roadmap section headers need reorganization** — Phase 1 is done. The `## Current Phase` label should move to Phase 2 (or a new section "Completed Phases" created for Phase 1). Recommend user edits `Roadmap.md` to restructure and explicitly flag Phase 2 as the active focus.

2. **LOW — Add Plan 41 headless CLI as a roadmap item** — Suggest adding to Phase 2: `- [x] Headless CLI contract — pure \`*Result\` cores, JSONL/exit-code interface, \`list --json\`, \`docs/agent-interface.md\` for agent callers`. Or to Phase 3 as a prerequisite for Managed Agents integration. User should decide placement.

3. **LOW — Clarify "Micro-task fast path" scope** — Does Plan 41's headless CLI satisfy the intent, or is a separate lightweight-protocol-bypass still planned?

---

## Notes for Next Run

- If user restructures Phase headings and adds Plan 41 row, next run will be clean.
- The `KeyDecisionLog.md` staleness is worth flagging to the user separately (not roadmap-routine scope, but 90 days without a new entry while 2 major plans shipped is notable).
- Phase 2 items (self-update, template vars, micro-task fast path) remain open with no active plans — check if any have been scheduled since this run.
