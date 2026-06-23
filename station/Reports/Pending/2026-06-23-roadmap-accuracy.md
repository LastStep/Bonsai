---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-06-23
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
- **Files Read:** 7 — `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Plans/Active/40-odysseus-platform-integration.md`, `station/Playbook/Plans/Active/41-headless-cli-contract.md`, `station/agent/Core/routines.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Write, Edit, Bash, Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `station/Playbook/Roadmap.md` and cross-referenced against `station/Playbook/Status.md` and recent work context.

**Phase 1 — Foundation & Polish:** All 11 items are marked `[x]`. The last routine run (2026-05-07) applied two quick fixes: marked "Better trigger sections" `[x]` with annotation, and added the `bonsai validate` row. Both are correctly reflected in current Roadmap.md.

**Phase 1 completeness gap:** Phase 1 has no row for the v0.5.0 work shipped in Plans 40+41:
- In-repo memory graph (`station/Memory/`) + project manifest (`.bonsai/project.yaml`) — Plan 40
- Headless CLI contract (pure `*Result` cores, JSONL/exit contract, `docs/agent-interface.md`) — Plan 41

These are significant shipped features (merged via 5 PRs to main at `ab202c3`) with no representation in the roadmap.

**Current Phase header is stale:** The roadmap's `## Current Phase` section still points to Phase 1, which is 100% complete. The project is now operating in Phase 2 territory.

### Step 2 — Check milestone accuracy

**Phase 2 remaining items:**
- `[ ] Self-update mechanism` — Plan 40 Phase 4 (held) is a form of update delivery for new scaffolding items. Not shipped; flag as potentially tracking this intent, but Phase 4 is still held.
- `[ ] Template variables expansion` — not in any active plan or pending work. Still accurate as unbuilt.
- `[ ] Micro-task fast path` — not in any active plan or pending work. Still accurate as unbuilt.

**MCP server gap:** Plan 41's primary stated motivation is enabling Plan 42 (MCP server, "fast-follow"). Plan 42 is not in the roadmap at all. Phase 3 mentions "Managed Agents integration — `bonsai deploy`, session management, outcome rubrics" but this is a different scope than a local MCP server for Claude Code / Cursor / Claude Desktop. The two should be separated in Phase 3 or a new item added.

**v0.5.0 tag held:** Per Status.md, tag is held by user decision. Not a roadmap accuracy issue — the work is done; the release cadence is a deployment decision.

### Step 3 — Cross-check against Key Decision Log

Read `station/Logs/KeyDecisionLog.md` in full.

- **"Defer Managed Agents cloud integration until local foundation is stable" (2026-04-02)** — still correctly represented in Phase 3 as unstarted. No invalidation.
- **Plan 40 locked decisions (2026-06-13)** — introduced `station/Memory/` + `.bonsai/project.yaml` as repo-side standards, and `bonsai` as schema authority for Odysseus. This is an architectural evolution not captured in the roadmap.
- **Plan 41 architecture decision (2026-06-16)** — "each command's core is a pure function: typed-options in → structured `Result` out" — this shapes how Plan 42 (MCP server) will be built. Phase 3 predates this decision and doesn't account for the layered CLI/MCP architecture.
- No KeyDecisionLog entry invalidates any existing roadmap item. All settled decisions remain consistent with current roadmap content.

### Step 4 — Report findings (flagged for user review, no Roadmap.md edits)

Per procedure: findings documented below. Roadmap.md not modified.

### Step 5 — Update dashboard

Updated `station/agent/Core/routines.md` — Roadmap Accuracy row: Last Ran → 2026-06-23, Next Due → 2026-07-07, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | `## Current Phase` still points to Phase 1, which is 100% complete — project is now in Phase 2 | `Roadmap.md` lines 14–31 | Flagged for user — recommend moving Phase 2 to Current Phase, Phase 1 to a "Completed Phases" or "Foundation" section |
| 2 | medium | v0.5.0 features (Plans 40+41) not reflected in roadmap — headless CLI contract, in-repo memory graph, project manifest are significant shipped work with no roadmap row | `Roadmap.md` Phase 1 section | Flagged for user — recommend adding rows to Phase 1 (already complete) or annotating as v0.5.0 milestones |
| 3 | low | Plan 42 (MCP server, "fast-follow" per Plan 41) has no roadmap entry — Phase 3 mentions "Managed Agents integration" which is broader/different scope | `Roadmap.md` Phase 3 | Flagged for user — consider adding a MCP server item distinct from managed-agents cloud integration |
| 4 | low | Plans 40 and 41 still in `Plans/Active/` despite both being in Status.md Recently Done — bookkeeping drift | `station/Playbook/Plans/Active/` | Flagged for user — archive both plan files (also flagged by status-hygiene and doc-freshness-check routines today) |
| 5 | info | Phase 2 "Self-update mechanism" may partially overlap with Plan 40 Phase 4 (held — update-delivery for new scaffolding items) | `Roadmap.md` Phase 2 | No action needed — item still unchecked and accurately reflects that the feature isn't shipped; context noted |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[medium] Transition Current Phase to Phase 2** — All Phase 1 items are `[x]`. The `## Current Phase` header should point to Phase 2, with Phase 1 moved to a completed/archived section or noted inline.

2. **[medium] Add v0.5.0 items to Roadmap** — Plans 40 and 41 shipped significant features (in-repo memory graph, project manifest, headless CLI contract, MCP-ready cores). These aren't in any roadmap phase. Options: (a) add rows to Phase 1 (already complete, annotate as "shipped v0.5.0"), (b) add a Phase 1.5 / v0.5.0 milestone block, (c) add to Phase 2 as completed items.

3. **[low] Add MCP server (Plan 42) to Roadmap Phase 3** — Plan 41 explicitly targets a fast-follow Plan 42 for the MCP server. The current Phase 3 "Managed Agents integration" covers cloud deployment; a local MCP server is a distinct, nearer-term item.

4. **[low] Archive Plans 40 and 41** — Both plan files remain in `Plans/Active/` but both plans are done per Status.md. Three separate routines today (status-hygiene, doc-freshness-check, roadmap-accuracy) have now flagged this.

## Notes for Next Run

- If Phase 1 → Phase 2 transition is made, verify Phase 2 has a clear "Current" milestone to audit against.
- If Plan 42 (MCP server) is added to the roadmap, check for a plan file in Active/ next run.
- The v0.5.0 tag is held by user — if it ships before next run, the roadmap should reflect the release milestone.
- The HOMEBREW_TAP_TOKEN PAT expiry (~2026-07-15) surfaced in backlog-hygiene today — unrelated to roadmap but the v0.5.0 release will require it.
