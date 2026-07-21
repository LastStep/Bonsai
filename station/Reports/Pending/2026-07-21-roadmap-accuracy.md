---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-07-21
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
- **Duration:** ~10 min
- **Files Read:** 6 — `station/agent/Routines/roadmap-accuracy.md`, `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Backlog.md`, `station/Playbook/Plans/Active/40-odysseus-platform-integration.md` (referenced via Status.md entries)
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read (file reads), Glob (active plans list), Write (report creation), Edit (dashboard + log updates)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state

**Action:** Read `station/Playbook/Roadmap.md` and compared all checkboxes against known shipped work from Status.md and RoutineLog.md.

**Result:**

- **Phase 1 — Foundation & Polish:** All 11 items correctly marked `[x]`. The two items that were previously unresolved (`Better trigger sections` and `bonsai validate`) were both checked and annotated in the 2026-05-07 Routine Digest. Phase 1 is clean.

- **Phase 2 — Extensibility:** `Custom item detection` is correctly `[x]`. The three remaining unchecked items (self-update mechanism, template variables expansion, micro-task fast path) are all in Backlog P3 under "Future Platform (Roadmap Phase 2+)". Accurate — these have been deprioritized. No mismatches in Phase 2 checkboxes.

- **Phase 3 — Cloud & Orchestration:** Both items remain unchecked. Accurate.

- **Phase 4 — Ecosystem:** All three items remain unchecked. Accurate.

**Missing from roadmap:** Two significant milestones completed since the last roadmap-accuracy run are not reflected anywhere in the roadmap:
  1. **Plan 41 — Headless CLI Contract + MCP-ready cores** (shipped 2026-06-16, PRs #120/#122/#123/#121/#125). All mutating commands (init/add/update/remove) now have pure `*Result` headless cores + JSONL/exit contract (ExitConflict=5); `list --json` added; `docs/agent-interface.md` published.
  2. **Plan 42 — MCP server** is explicitly described as "fast-follow" in the Plan 41 Status.md entry. Not in roadmap, but is the identified next Phase 3 deliverable.

**Issues:** 2 medium findings (see Findings Summary)

### Step 2: Check milestone accuracy

**Action:** Assessed whether the next milestones are still the right priority and whether any planned work has been superseded.

**Result:**

- **Phase 2 items deprioritized, not superseded.** The three unchecked Phase 2 items remain in the backlog but have effectively been pushed behind Phase 3 infrastructure work (Plans 40+41). The roadmap has no annotation acknowledging this sequence change — it implies Phase 2 must complete before Phase 3, which is no longer how the project is sequencing work.

- **Phase 3 language may need updating.** The current Phase 3 lists "Managed Agents integration — `bonsai deploy`, session management, outcome rubrics." The actual next Phase 3 deliverable is now Plan 42 (MCP server), which is a different interface than "Managed Agents integration." An MCP server is an API surface for tool-calling agents; Managed Agents is a cloud deployment model. These are related but distinct — the roadmap may need a new row or revised language to reflect the MCP direction.

- **No planned work has been explicitly superseded.** All unchecked items remain valid future work. Nothing in the KeyDecisionLog or Status.md explicitly cancels any roadmap item.

**Issues:** 1 low finding (see Findings Summary)

### Step 3: Cross-check against Key Decision Log

**Action:** Read `station/Logs/KeyDecisionLog.md` and compared all decisions against roadmap items.

**Result:**

- "Defer Managed Agents cloud integration until local foundation is stable" (2026-04-02, Settled) — This decision is honored. However, Plan 41's headless CLI work has substantially matured the local foundation. The "fast-follow Plan 42" (MCP server) signals the deferral window may be closing. No roadmap update needed from this decision, but it's worth noting the context has shifted.

- All other structural, domain-specific, and settled decisions remain aligned with the roadmap. No decisions found that invalidate any roadmap item.

**Issues:** none

### Step 4: Report findings

Per procedure, flagging all mismatches for user review. No direct edits made to `Roadmap.md`.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | Plan 41 (Headless CLI Contract + MCP-ready cores, shipped 2026-06-16) has no corresponding roadmap row. This is a shipped headline feature — 5 PRs, new contract doc — with no Phase 2 or 3 entry | `Roadmap.md` Phase 2 or new Phase 3 precursor row | Flagged for user review — suggest adding `[x]` row to Phase 2 or 3 |
| 2 | medium | Plan 42 (MCP server, "fast-follow") is the identified next Phase 3 deliverable but is absent from the roadmap. Current Phase 3 only lists Managed Agents + Greenhouse | `Roadmap.md` Phase 3 | Flagged for user review — suggest adding MCP server row to Phase 3 |
| 3 | low | Phase 2 unchecked items (self-update mechanism, template variables expansion, micro-task fast path) are effectively deprioritized in favor of Phase 3 infrastructure, but the roadmap implies linear completion. No annotation acknowledges the skip | `Roadmap.md` Phase 2 | Flagged for user review — suggest annotation or reordering if Phase 3 work is intentionally out-of-sequence |
| 4 | low | Phase 3 "Managed Agents integration" description (`bonsai deploy`, session management, outcome rubrics) may need to be updated or supplemented to include the MCP server path now that Plan 42 is the confirmed next step in that direction | `Roadmap.md` Phase 3 | Flagged for user review |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **Roadmap.md — add Plan 41 row:** Plan 41 (Headless CLI Contract + MCP-ready cores) shipped 2026-06-16. Suggest adding `[x] Headless CLI contract + MCP-ready cores — JSONL/exit contract, *Result headless cores for all mutating commands, list --json, agent-interface.md. (Plan 41)` to Phase 2 or as a Phase 3 precursor.

- **Roadmap.md — add Plan 42 / MCP server row:** Status.md describes Plan 42 as a "fast-follow" to Plan 41. Suggest adding `[ ] MCP server — expose Bonsai commands as MCP tools for direct agent-to-agent integration. (Plan 42)` to Phase 3.

- **Roadmap.md — Phase 2 sequencing annotation:** If Phase 3 infrastructure (Plans 40/41/42) is intentionally being built before Phase 2 items complete, consider adding a brief note to the Phase 2 section (e.g., "Phase 3 precursor work (Plans 40-42) in progress; Phase 2 items deprioritized until Phase 3 integration path is clear") or resequencing the phases.

---

## Notes for Next Run

- Phase 1 is stable — no need to re-check.
- Focus next run on: (1) whether Plan 42 (MCP server) shipped and should be checked, (2) whether Phase 2 items remain intentionally deferred, (3) whether new work has been started in Phase 3/4.
- The 2026-05-07 prior run found 2 low-severity items; both were resolved in the 2026-05-07 Routine Digest (Roadmap.md edits applied). Starting clean this cycle.
