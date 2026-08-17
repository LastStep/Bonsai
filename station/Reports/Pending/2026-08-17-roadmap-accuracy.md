---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-08-17
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
  - `/home/user/Bonsai/station/Playbook/Backlog.md`
- **Files Modified:** 3
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
  - `/home/user/Bonsai/station/Reports/Pending/2026-08-17-roadmap-accuracy.md` (this report)
- **Tools Used:** Read, Write, Edit, Bash (directory listings)
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `Roadmap.md` and checked each phase against the current archive of Plans 1–41 (all in `Plans/Archive/`) and `Status.md`.

**Phase 1 — Foundation & Polish:** All 11 items are `[x]`. The last two items (`Better trigger sections` and `bonsai validate`) were unresolved in the 2026-05-07 run and appear to have been added since then — both are now correctly marked. Phase 1 is fully complete. However, the Roadmap still lists Phase 1 under "## Current Phase." Since all items are done and Phase 2 work has begun (`[x] Custom item detection`), the label is stale.

**Phase 2 — Extensibility:** One of four items is `[x]` (`Custom item detection`, shipped ~Plan 34). Three items remain open. However, Plan 41 (Headless CLI Contract, shipped 2026-06-16) represents a major capability addition with no corresponding Roadmap entry. This is a gap.

**Phase 3 — Cloud & Orchestration / Phase 4 — Ecosystem:** No changes — all items still open. The `Managed Agents integration` deferral decision in the KeyDecisionLog remains valid.

### Step 2 — Check milestone accuracy

- No next milestones are explicitly labeled. Phase 2 is effectively the active phase but the Roadmap doesn't signal this.
- Plan 42 (`bonsai mcp` stdio server) is filed in Backlog P2 as the concrete next work item, but it does not appear in any Roadmap phase. It bridges Phase 2 (extensibility) and Phase 3 (Cloud) — closest alignment is Phase 3 "Managed Agents integration" direction, but it's a distinct deliverable.
- Phase 2 item `[ ] Template variables expansion` has no Backlog tracking entry — flagged by backlog-hygiene this same session. The Roadmap item is valid but lacks a concrete plan or backlog anchor.
- Phase 2 item `[ ] Self-update mechanism` is in Backlog P3. Roadmap alignment is correct.
- Phase 2 item `[ ] Micro-task fast path` is in Backlog P3. Roadmap alignment is correct.

### Step 3 — Cross-check against Key Decision Log

The KeyDecisionLog was read in full. All entries are dated 2026-04-13 or earlier. No new decisions have been recorded since then, despite Plans 40 and 41 making significant architectural decisions:

- **Plan 40**: Froze v1 schemas for `.bonsai/project.yaml` and the manifest format; chose root-relative scaffolding paths for memory and manifest.
- **Plan 41**: Established the headless CLI contract — JSONL output, `ExitConflict=5`, `*Result` core types as the stable agent-drivable interface; produced `docs/agent-interface.md` as the contract document.

These decisions affect how future features (MCP server, Managed Agents integration) must be designed. The KeyDecisionLog should reflect them.

No existing KDL decisions were found to invalidate current Roadmap items.

### Step 4 — Report findings

Four findings compiled below. Per routine instructions, Roadmap.md was not modified — all findings are flagged for user review.

### Step 5 — Update dashboard and log

Dashboard updated; RoutineLog entry appended.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | "Current Phase" header still shows Phase 1 — all 11 items are `[x]`, Phase 2 has begun | `Roadmap.md` § Current Phase | Flagged for user review |
| 2 | Medium | Headless CLI contract (Plan 41, 2026-06-16) not in any Roadmap phase — major capability gap | `Roadmap.md` § Phase 2 | Flagged for user review |
| 3 | Low | Plan 42 (bonsai mcp stdio server) in Backlog P2 but absent from Roadmap | `Roadmap.md` § Phase 2 or Phase 3 | Flagged for user review |
| 4 | Low | KeyDecisionLog has no entries since 2026-04-13 despite Plans 40+41 making significant architectural decisions | `KeyDecisionLog.md` | Flagged for user review |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

### Finding 1 — Phase transition: Phase 1 → Phase 2 (Medium)

All Phase 1 items are `[x]`. Phase 2 already has one checked item (`Custom item detection`). The Roadmap's "## Current Phase" header still points at Phase 1. **Suggested action:** Move Phase 1 to a "## Completed Phases" (or similar) section, promote Phase 2 to "## Current Phase."

### Finding 2 — Headless CLI contract missing from Roadmap (Medium)

Plan 41 (shipped 2026-06-16) added headless cores for all four mutating commands (`init`, `add`, `update`, `remove`), JSONL output, `ExitConflict=5`, `list --json`, and `docs/agent-interface.md`. This is a Phase 2-grade capability that makes Bonsai machine-drivable — a prerequisite for the MCP server and future Managed Agents integration. It currently appears in Status.md and Plans/Archive but not in the Roadmap. **Suggested action:** Add to Phase 2 as:
```
- [x] Headless CLI contract — JSONL/exit-code API, *Result cores for all mutating commands,
      list --json, docs/agent-interface.md agent contract doc (Plan 41, 2026-06-16)
```

### Finding 3 — Plan 42 (bonsai mcp) not in Roadmap (Low)

Plan 42 (`bonsai mcp` stdio server) is listed in Backlog P2 and described as the natural fast-follow to Plan 41. It uses the headless cores as its substrate. It doesn't map cleanly to any current Roadmap item — Phase 3's "Managed Agents integration" is cloud-focused (bonsai deploy), while Plan 42 is a local MCP transport layer. **Suggested action:** Either add as a Phase 2 item (`[ ] MCP server — bonsai mcp stdio, one tool per headless core`) or as an early Phase 3 enabler. Defer until Plan 42 is formally kicked off.

### Finding 4 — KeyDecisionLog gap since April 2026 (Low)

Plans 40 and 41 produced decisions that affect future architecture: the frozen v1 schema format, the headless core pattern, the JSONL/exit-code contract, and `ExitConflict=5`. These don't appear in the KeyDecisionLog. **Suggested action:** Add 2–3 entries capturing the headless contract decision and the v1 schema freeze. Specifically relevant when designing the MCP server (Plan 42) or any future agent-integration work.

---

## Notes for Next Run

- If Phase 1 → Phase 2 transition happens before next run, verify Phase 2 header and checked items are accurate.
- Plan 42 may be in flight by next run — check if it should be promoted from Backlog to Roadmap.
- KeyDecisionLog entries for Plans 40/41 may have been added — verify at next run.
- The `[ ] Template variables expansion` item in Phase 2 still lacks a Backlog entry — check if one was created.
