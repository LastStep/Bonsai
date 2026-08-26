---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-08-26
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
- **Duration:** ~6 min
- **Files Read:** 5 — `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `Playbook/Roadmap.md` and compared each item against known shipped work.

**Phase 1 — Foundation & Polish:** All 11 items are `[x]` (checked). Cross-referenced against Status.md Done rows and RoutineLog entries. Every item matches reality:
- Go rewrite: confirmed (Go 1.25+, Cobra/Huh/LipGloss/BubbleTea in CLAUDE.md)
- Full catalog: confirmed (6 agent types, skills/workflows/protocols/sensors/routines)
- Lock file conflict handling: confirmed (lockfile.go in internal/config/)
- Awareness Framework: confirmed (status-bar + context-guard sensors active)
- Dogfooding: confirmed (station/ workspace active)
- Better trigger sections: `[x]` with annotation (Plans 08/17/21 + context-guard regex; Plan 08 C3 deferred to P3 Backlog) — correct
- UI overhaul: confirmed (TUI cinematic flows shipped Plans 22/23/28/31)
- Usage instructions: confirmed (Plan 05, guide command)
- Release pipeline: confirmed (GoReleaser + GitHub Actions + Homebrew Tap)
- Community health files: confirmed (ISSUE_TEMPLATE, CONTRIBUTING.md, CODE_OF_CONDUCT.md)
- bonsai validate: confirmed (Plan 35, v0.4.0 headline, validate.go in internal/)

Phase 1 is clean — no corrections needed.

**Phase 2 — Extensibility:** Custom item detection `[x]` confirmed (scan.go exists in internal/generate/). Three remaining items (`[ ]`) have no active plans and no Status.md placement — still accurately unchecked.

**Phase 3 — Cloud & Orchestration:** Both items remain unchecked. KeyDecisionLog 2026-04-02 explicitly defers Managed Agents until local foundation is stable. Greenhouse companion app has no active work. Accurately unchecked.

**Phase 4 — Ecosystem:** All items unchecked. No active plans or signals. Accurate.

### Step 2 — Check milestone accuracy

Reviewed Status.md Recently Done + Pending for priority signals:
- Plan 41 (Headless CLI Contract + MCP-ready cores) shipped 2026-06-16 on main (`ab202c3`). Status.md notes: "MCP server = fast-follow Plan 42."
- Plan 42 (MCP server) is the implied next major initiative — not yet a plan file but referenced as "fast-follow."
- Phase 3 roadmap currently lists "Managed Agents integration — `bonsai deploy`, session management, outcome rubrics" and "Greenhouse companion app." The MCP server direction is a distinct track not captured in either Phase 3 item.
- **Finding (Low):** Phase 3 does not reflect the MCP server as a near-term milestone. Plan 42 is the stated next step but has no roadmap entry. The roadmap may need a new Phase 3 item for "MCP server — agent-accessible headless interface for tool/IDE integration" to capture this direction.

Reviewed whether Phase 2 items remain the right priorities:
- Self-update mechanism, Template variables expansion, Micro-task fast path — none have active plans, no Status.md placement, no Backlog P1 entries. The Phase 2 framing ("Users can create custom catalog items, extend existing ones, and share them") remains valid as a goal, but the specific unchecked items may have drifted from current priorities.
- **Finding (Low):** Phase 2 remaining items have been static since initial roadmap creation (~April 2026). No concrete plan activity for any of them across the 111-day gap since last run. Worth confirming if these remain the right next-steps or if they've been superseded by the MCP/headless direction.

### Step 3 — Cross-check against Key Decision Log

Read `Logs/KeyDecisionLog.md` in full. No recent decisions (nothing post-2026-04-13 in the log) that could invalidate roadmap items. Key cross-checks:
- "Bonsai is a scaffolding tool, not a runtime orchestrator" — Phase 3 Managed Agents item uses `bonsai deploy` (setup step), consistent with this decision.
- "Defer Managed Agents cloud integration until local foundation is stable" — reflected accurately in Phase 3 remaining unchecked. Plan 41 headless-CLI work is the most recent "local foundation" investment.
- All Structural and Domain-Specific decisions align with the current roadmap structure.

Cross-check: **clean.**

### Step 4 — Report findings

Two low-severity findings. No Roadmap.md edits made — flagged for user review per routine procedure.

### Step 5 — Update dashboard

Updated routines.md (Last Ran → 2026-08-26, Next Due → 2026-09-09, Status → done).

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Low | Phase 3 does not include MCP server, which Status.md identifies as "fast-follow Plan 42" — the most concrete near-term Phase 3 step | Roadmap.md Phase 3 | Flagged for user review — recommend adding a `[ ] MCP server — headless interface for tool/IDE integration` item to Phase 3 |
| 2 | Low | Phase 2 remaining items (Self-update mechanism, Template variables expansion, Micro-task fast path) have no active plans, no Backlog P1 entries, and no Status.md presence after 111-day gap | Roadmap.md Phase 2 | Flagged for user review — confirm these remain the right next-steps or replace with current priorities |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**Finding 1 — Phase 3: MCP server missing from roadmap**

Plan 41 (Headless CLI Contract, shipped 2026-06-16) explicitly identifies Plan 42 (MCP server) as the fast-follow next step. The MCP server is a distinct, concrete deliverable that sits within Phase 3's "Cloud & Orchestration" theme (agent-accessible interface for tool and IDE integration), but no roadmap entry exists for it.

Recommended action: Add to Phase 3:
```
- [ ] MCP server — agent-accessible headless interface; exposes `init`/`add`/`update`/`remove`/`list` operations via MCP protocol for tool and IDE integration (Plan 42)
```

**Finding 2 — Phase 2: Remaining items may have drifted from current priorities**

The three unchecked Phase 2 items have been static since April 2026 with no plan activity. The project has since invested heavily in headless CLI, MCP readiness, and validation infrastructure. It's worth confirming:
- Are Self-update mechanism / Template variables expansion / Micro-task fast path still the right Phase 2 priorities?
- Has any Phase 2 work been subsumed into other plans (e.g., Plan 40's v1 schema freezing as a form of template stabilization)?

Recommended action: Review and either reaffirm these items, swap them for more current priorities, or add a note indicating these are deferred to Phase 4/Ecosystem.

## Notes for Next Run

- Phase 1 is stable — no changes expected.
- If Plan 42 (MCP server) ships before the next run (2026-09-09), check its Phase 3 entry and mark `[x]`.
- KeyDecisionLog has no entries since 2026-04-13 — if architectural decisions were made around Plans 40/41, they should be logged there (not a roadmap issue, but worth noting for the next memory-consolidation or doc-freshness run).
- The "tag held" situation for v0.5.0 (Plan 40, main branch) and no v0.5.x release since v0.4.3 is not a roadmap issue but may be relevant context for a release-timing conversation.
