---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-06-17
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
- **Files Read:** 6 — `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/41-headless-cli-contract.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log), Glob, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state

- **Action:** Read `Roadmap.md` in full, cross-referenced every checkbox against `Status.md`, `RoutineLog.md`, and git log since 2026-05-07.
- **Result:**
  - **Phase 1 (Foundation & Polish):** All items correctly marked `[x]`. Previous run (2026-05-07) applied the `bonsai validate` and "Better trigger sections" fixes. Phase 1 appears clean.
  - **Phase 2 (Extensibility):** `[x] Custom item detection` correctly marked. Three remaining items (`[ ] Self-update mechanism`, `[ ] Template variables expansion`, `[ ] Micro-task fast path`) remain unshipped — correct.
  - **Phase 3 / Phase 4:** All unchecked, none have shipped — correct.
  - **MISSING FROM ROADMAP:** Plan 40 (Odysseus Platform Integration, Phases 1–3, shipped 2026-06-13) and Plan 41 (Headless CLI Contract + MCP-Ready Cores, shipped 2026-06-16) represent significant new capabilities with no corresponding Roadmap entry. These are substantial milestones:
    - **Plan 40 deliverables:** frozen v1 schemas, root-relative scaffolding, project-level `validate` pass (adversarial path/symlink hardening), memory-routing protocol, guide Formats page.
    - **Plan 41 deliverables:** pure `*Result` headless cores for all 4 mutating commands (init/add/update/remove), JSONL + JSON exit-code contract, `list --json`, documented `agent-interface.md` contract. MCP server (Plan 42) is now a fast-follow.
- **Issues:** 3 findings (see Findings Summary below).

### Step 2: Check milestone accuracy

- **Action:** Reviewed whether next milestones in each phase remain the right priority. Checked for superseded planned items.
- **Result:**
  - Phase 2 "Self-update mechanism" and "Template variables expansion" are still unstarted and relevant.
  - Phase 2 "Micro-task fast path" — no backlog item tracking this; no decisions to invalidate it, just unstarted.
  - Phase 3 "Managed Agents integration" — the KeyDecisionLog (2026-04-13) says "Defer cloud integration until local foundation is stable." Plan 41 (headless CLI) is the primary local-foundation prerequisite mentioned implicitly. The prerequisite is now substantially met. Phase 3 could realistically begin (Plan 42 MCP server is the bridging step). The Roadmap's Phase 3 entry doesn't reflect the Plan 42 fast-follow context.
  - No roadmap items reference deprecated approaches. All Phase 3–4 items remain valid future work.
- **Issues:** 1 finding — Phase 3 entry point is now reachable (prerequisite met by Plan 41) but Roadmap doesn't signal this readiness.

### Step 3: Cross-check against Key Decision Log

- **Action:** Read `KeyDecisionLog.md` in full; checked all decisions for contradictions with Roadmap.
- **Result:**
  - **2026-04-13 — Defer Managed Agents cloud integration:** Now partially resolved — the "stable local foundation" condition is met (Plan 41 shipped). The decision to defer is no longer blocking Phase 3 start. Roadmap text doesn't need to change (Phase 3 is still correct), but the signal that Phase 3 is unblocked is missing.
  - **All other structural decisions** (Go rewrite, embed.FS, lock file, tech-lead required, etc.) remain consistent with the Roadmap.
  - **No decisions invalidate any Roadmap items.**
- **Issues:** none beyond the Phase 3 readiness observation above.

### Step 4: Report findings

- **Action:** Compiled findings list below; flagging for user review per procedure (no direct Roadmap edits).
- **Issues:** none.

### Step 5: Update dashboard

- **Action:** Updated `agent/Core/routines.md` dashboard row for Roadmap Accuracy: Last Ran → 2026-06-17, Next Due → 2026-07-01, Status → done.
- **Issues:** none.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | Plan 40 deliverables (frozen schemas, root-relative scaffolding, project validate pass, memory-routing protocol, guide Formats page) shipped 2026-06-13 with no Roadmap entry. These fit Phase 2 (Extensibility) or as a Phase 1 addendum. | `Roadmap.md` Phase 1 / Phase 2 | Flagged for user — suggest adding entries or a v0.5.0 annotation |
| 2 | MEDIUM | Plan 41 deliverables (headless CLI contract, MCP-ready `*Result` cores for all 4 mutating cmds, `list --json`, `agent-interface.md` contract doc) shipped 2026-06-16 with no Roadmap entry. Represents the "headless/agent-drivable" milestone that gates Phase 3. | `Roadmap.md` Phase 2 / Phase 3 | Flagged for user — major milestone worth a Roadmap row |
| 3 | LOW | Phase 3 "Managed Agents integration" prerequisite is now met (Plan 41 = stable headless foundation). Plan 42 (MCP server, fast-follow) is the next bridging step. Roadmap Phase 3 doesn't reflect this readiness signal or the Plan 42 interim step. | `Roadmap.md` Phase 3 | Flagged for user — consider adding MCP server as a Phase 3 first item |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**Finding 1 — Plan 40 not on Roadmap (MEDIUM)**

Plans 40 Phases 1–3 shipped significant foundational work (v0.5.0 content, currently untagged):
- Frozen v1 schemas (`.bonsai/project.yaml` + memory manifest)
- Root-relative scaffolding paths (manifest + memory)
- Project-level `bonsai validate` pass with symlink/path hardening
- Memory-routing docs + `bonsai guide` Formats page

Suggested Roadmap addition — append to Phase 1 or Phase 2:
```
- [x] Frozen v1 schemas + project.yaml — stable hub-facing identity + memory manifest
- [x] Project-level validate pass — adversarial path/symlink hardening, root-relative scaffolding
```

**Finding 2 — Plan 41 not on Roadmap (MEDIUM)**

Plan 41 shipped 2026-06-16 (5 PRs, all merged to main, `ab202c3`). This is the biggest architectural milestone since `bonsai validate`:
- Pure `*Result` headless cores for `init`/`add`/`update`/`remove`
- Unified JSONL (mutating) + JSON (read) exit-code contract
- `list --json` structured output
- `docs/agent-interface.md` contract doc (agent/CI/MCP consumable)

Suggested Roadmap addition — append to Phase 2:
```
- [x] Headless CLI contract + MCP-ready cores — every command has a pure *Result core; JSONL/JSON exit-code contract; agent-interface.md spec
```

**Finding 3 — Phase 3 readiness signal missing (LOW)**

The KeyDecisionLog says "Defer Managed Agents cloud integration until local foundation is stable." Plan 41 is that foundation. Phase 3 entry is now unblocked. The MCP server (Plan 42 — not started) is the natural first Phase 3 step, but the Roadmap Phase 3 list starts with "Managed Agents integration — `bonsai deploy`" which is a bigger leap.

Suggested: Add `[ ] MCP server (`bonsai mcp`) — thin wrapper over headless cores, Plan 42 fast-follow` as the first Phase 3 item.

## Notes for Next Run

- If user accepts the Roadmap additions above, the next run should verify they are accurately checked/unchecked.
- Plan 40 tag (v0.5.0) is still held by user — if released before next run, Roadmap should note the version tag.
- Plan 42 (MCP server) was referenced as a "fast-follow" but has not been planned yet (no plan file in Active/). Check if it has been started.
- Phase 4 remains entirely in future — no changes expected.
