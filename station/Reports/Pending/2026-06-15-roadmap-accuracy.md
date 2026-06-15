---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-06-15
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
- **Files Read:** 7 — `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/Playbook/Backlog.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Plans/Archive/41-headless-cli-contract.md`, `station/Playbook/Plans/Active/40-odysseus-platform-integration.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Write, Edit, Glob, Bash (git log)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

**Read `station/Playbook/Roadmap.md`:** All Phase 1 items remain checked `[x]`. Phase 2 has one checked item (Custom item detection) and three unchecked. Phases 3 and 4 are fully unchecked.

**What's actually been built since last run (Plans 40 + 41 shipped):**

**Plan 40 — Odysseus/v0.5.0 (Phases 1–3 shipped, Phase 4 held):**
- `.bonsai/project.yaml` project manifest (new scaffolding item, opt-in)
- `station/Memory/` memory graph scaffolding (new scaffolding item, opt-in)
- Validate project-level pass — `bonsai validate` now audits manifest + memory notes
- Frozen v1 schemas for memory notes and project manifest
- `bonsai guide` "Formats" page for memory routing docs

**Plan 41 — Headless CLI Contract + MCP-Ready Cores (all 5 phases shipped, PRs #120/#122/#123/#121/#125):**
- Pure `*Result`-returning headless cores for all four mutating commands (init, add, update, remove)
- JSONL streaming contract for mutating commands; `list --json` for read commands
- Unified exit-code contract (`ExitConflict=5`)
- `docs/agent-interface.md` — formal contract doc
- Architecture: MCP server (Plan 42) will be a thin wrapper over these same cores — zero duplication

**Roadmap alignment check:**

| Phase | Item | Status | Assessment |
|-------|------|--------|------------|
| Phase 1 | All items | `[x]` | Correct — all shipped |
| Phase 2 | Custom item detection | `[x]` | Correct — shipped |
| Phase 2 | Self-update mechanism | `[ ]` | Correct — not built |
| Phase 2 | Template variables expansion | `[ ]` | Correct — not built |
| Phase 2 | Micro-task fast path | `[ ]` | Correct — not built |
| Phase 3 | Managed Agents integration | `[ ]` | Correct — not built; but Plan 41 shipped direct prerequisite |
| Phase 3 | Greenhouse companion app | `[ ]` | Correct — not built |
| Phase 4 | All items | `[ ]` | Correct — not built |

**Key gaps identified — items not reflected in Roadmap:**

1. **Plan 40 shipped extensible scaffolding (project manifest + memory graph)** — This is a Phase 2-tier capability (users can extend their workspace with Odysseus-compatible structured memory and project identity). The Roadmap Phase 2 goal says "users can create custom catalog items, extend existing ones, and share them" — the manifest/memory additions are directly in this spirit but are not reflected as a shipped item.

2. **Plan 41's headless CLI contract is a significant Phase 2/3 milestone** — Agent-drivable, CI-scriptable, MCP-ready surface. This is arguably the biggest prerequisite step toward Phase 3 ("Managed Agents integration"), but it's also a Phase 2 extensibility feature in its own right. The roadmap has no entry for this shipped capability.

3. **v0.5.0 is untagged** — Plan 40 Phases 1–3 are on `main` as "v0.5.0 (untagged)" per Status.md. The tag is held by user decision. The roadmap does not track version milestones, so no change needed here, but worth noting that the distinction between "shipped on main" and "released" is a user decision already noted in Status.md.

### Step 2 — Check milestone accuracy

**Are next milestones still right priority?**

Phase 2 remaining items (`Self-update mechanism`, `Template variables expansion`, `Micro-task fast path`) are on Backlog P3, consistent with their `[ ]` status. No priority drift.

Phase 3 (`Managed Agents integration`) has Plan 42 (MCP server) identified as the fast-follow to Plan 41. This is the next logical step toward Phase 3. The roadmap correctly shows this as not started, but the foundation is now substantially more advanced than the roadmap indicates.

**Has any planned work been superseded?**

- Plan 40 Phase 4 (update-delivery for scaffolding) was superseded/absorbed by Plan 41. The headless update core (`bonsai update --non-interactive`) ships the delivery mechanism Plan 40 Phase 4 intended. Status.md documents this. No roadmap item references Plan 40 Phase 4 specifically, so no correction needed.

**Deprecated approaches flagged:**

None found. All current roadmap items reference approaches that remain valid.

### Step 3 — Cross-check against Key Decision Log

**Decisions reviewed:**

- "Rewrite from Python to Go" — Phase 1 complete, no roadmap impact
- "Catalog embedded via embed.FS" — still accurate, no change
- "Defer Managed Agents cloud integration until local foundation is stable" — the Settled section still holds. Plan 41's MCP-ready cores make this decision potentially revisitable sooner (the local foundation is now substantially more stable), but the decision is explicitly "settled" and should not be relitigated without user direction.
- "Bonsai is a scaffolding tool, not a runtime orchestrator" — still accurate. Plan 41's headless cores do not change this — Bonsai generates workspaces; the MCP server (Plan 42) would be how AI drives it, consistent with "generates files and steps away."

**No recent decisions invalidate roadmap items.** No corrections required from decision log cross-check.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | Phase 2 missing shipped item: Plan 40 delivered project manifest + memory graph scaffolding (new opt-in scaffolding items, validate project pass). Not reflected in Roadmap Phase 2. | `Roadmap.md` Phase 2 | Flagged for user review — suggest adding row under Phase 2 |
| 2 | Medium | Phase 2/3 missing shipped milestone: Plan 41 delivered headless CLI contract + MCP-ready cores — major prerequisite for Phase 3 Managed Agents integration. Not reflected anywhere in Roadmap. | `Roadmap.md` Phase 2 / Phase 3 | Flagged for user review — suggest adding row(s) |
| 3 | Low | v0.5.0 untagged — Phases 1–3 of Plan 40 are on main but the tag is held. Roadmap does not track version tags (correct by design), but it's worth confirming the user still intends to hold. | `Status.md` note | No roadmap change needed; flagged as informational |
| 4 | Info | Phase 3 prerequisite substantially advanced — Plan 41's MCP-ready cores mean Plan 42 (MCP server) is now a thin-wrapper task. Phase 3 "Managed Agents integration" is closer than the roadmap implies. | `Roadmap.md` Phase 3 | No structural change needed; informational |

## Errors & Warnings

None.

## Items Flagged for User Review

### Flag 1 — Phase 2: Add shipped Plan 40 item
**Suggested addition to Phase 2:**
```
- [x] Project manifest + memory graph scaffolding — `.bonsai/project.yaml` (hub-facing project identity) + `station/Memory/` (structured memory notes with frozen v1 schema), both opt-in scaffolding items; `bonsai validate` project-level pass audits both. v0.5.0 (Plans 40/41).
```

User may prefer shorter phrasing or to split into two separate checkboxes. The point is that two new scaffolding items and a validate project pass shipped and are not reflected.

### Flag 2 — Phase 2 or Phase 3: Add shipped Plan 41 item
**Suggested addition (Phase 2 or a new "Phase 2.5 — Headless/MCP Foundation" note):**
```
- [x] Headless CLI contract + MCP-ready cores — all four mutating commands (init/add/update/remove) return pure `*Result` structs; JSONL streaming + unified exit-code contract; `list --json`; `docs/agent-interface.md` formal contract. MCP server (Plan 42) will be a thin wrapper. v0.5.0 (Plan 41).
```

This could live in Phase 2 (it's an extensibility feature enabling AI/CI consumers) or as a preamble note to Phase 3 (it's the direct prerequisite for Managed Agents). User's call on placement.

### Flag 3 — v0.5.0 tag decision
v0.5.0 remains untagged on `main`. Status.md says "tag held (user)." If the user has made a decision since, the tag status should be updated.

## Notes for Next Run

- If the user adds the two suggested Phase 2 items, next run will find them checked `[x]` and Phase 2 will show 3/5 items done.
- Plan 42 (MCP server) is the next likely roadmap-impacting plan. Watch for it to appear in Status.md Active/Pending.
- Phase 2 still has 3 unchecked items (`Self-update mechanism`, `Template variables expansion`, `Micro-task fast path`) — all Backlog P3. These are not in active planning; monitor for priority shift.
- The `[research] Trial sentrux` P0 item in Status.md Pending is blocked on Rust toolchain — if it ships, it's orthogonal to roadmap (ops/security, not a roadmap milestone).
