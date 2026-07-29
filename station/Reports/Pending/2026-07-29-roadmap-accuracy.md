---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-07-29
status: success
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 6 — `/home/user/Bonsai/station/agent/Routines/roadmap-accuracy.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/41-headless-cli-contract.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read (×6), Edit (×2), Write (×1)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state
- **Action:** Read `Playbook/Roadmap.md` and cross-checked each phase item against Status.md Recently Done and Plans/Active/.
- **Result:**
  - **Phase 1** — All 11 items marked `[x]`. Verified accurate against Status.md and RoutineLog (all items shipped across Plans 02–37, v0.4.0). The 2026-05-07 routine digest applied the two previously-flagged fixes (checked "Better trigger sections" with annotation; added `bonsai validate` row). Phase 1 is genuinely complete.
  - **Phase 1 structural drift** — The roadmap's `## Current Phase` heading still points to Phase 1, even though all Phase 1 items are done. Phase 2 lives under `## Future Phases`, which misrepresents the project's actual position: Phase 2 work is actively happening (Plan 41 shipped June 2026, Plan 40 v0.5.0 shipped June 2026).
  - **Phase 2** — "Custom item detection" correctly marked `[x]` (Plan 34). Three items remain `[ ]` (self-update, template variables, micro-task fast path) — these are still valid outstanding items.
  - **Phase 2 gap** — Plan 41 (Headless CLI Contract + MCP-ready cores) shipped 2026-06-16 with all 5 phases merged (PRs #120/#122/#123/#121/#125). This is a major Phase 2 deliverable — it added pure `*Result` headless cores for every mutating command, JSONL/exit-code contract, `list --json`, and the `docs/agent-interface.md` contract doc. There is no corresponding row in Roadmap Phase 2.
  - **Phase 2 gap (planned)** — Plan 42 (MCP server, fast-follow to Plan 41) is referenced in Status.md but not started and not in the roadmap at all.
  - **Phase 3/4** — All items correctly pending. KDL "Defer Managed Agents cloud integration" entry is still in force — Phase 3 items are deliberately deferred.
- **Issues:** 3 findings (1 structural, 2 gap).

### Step 2: Check milestone accuracy
- **Action:** Checked whether remaining Phase 2 items are still the right priorities, and whether any planned work has been superseded.
- **Result:**
  - The three remaining Phase 2 items (self-update, template variables expansion, micro-task fast path) are still valid and not superseded.
  - Plan 42 (MCP server) is now the practical next Phase 2/3 candidate based on Status.md and Plan 41's stated goal ("Plan 42 is a thin wrapper calling the same functions"). Its absence from the roadmap creates a navigation gap for any reader trying to understand what's next.
  - No roadmap items reference deprecated approaches. No items have been invalidated by changed priorities.
  - Plan 40 (v0.5.0) shipped "frozen v1 schemas + root-relative scaffolding (manifest + memory)" — this doesn't map to any named roadmap item. It's general platform hardening within Phase 2's extensibility goal. Low severity; could be documented as an unlisted Phase 2 milestone but not strictly wrong to omit.
- **Issues:** 1 finding (MCP server not in roadmap).

### Step 3: Cross-check against Key Decision Log
- **Action:** Read `Logs/KeyDecisionLog.md` in full and compared against roadmap items.
- **Result:**
  - All KDL decisions dated 2026-04-13 or earlier — no new key decisions logged since the last roadmap-accuracy run (2026-05-07).
  - "Defer Managed Agents cloud integration" (Settled, 2026-04-02) — still consistent with Phase 3 remaining `[ ]`.
  - No KDL decision invalidates any roadmap item.
  - Notable: Plan 40's architectural decision (frozen v1 schemas, memory_dir security fix) and Plan 41's architectural decision (MCP-is-a-wrapper pattern, layered CLI/MCP contract) were never logged to the KDL. These are significant structural decisions. Not a roadmap-accuracy issue, but worth flagging for the user.
- **Issues:** 0 roadmap-invalidating decisions; 1 KDL hygiene flag (unlisted structural decisions from Plans 40/41).

### Step 4: Report findings
- **Action:** Documented all findings below. Roadmap.md not modified per procedure — all items flagged for user review.
- **Result:** 4 items flagged (see Findings Summary).
- **Issues:** none.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Roadmap Accuracy row.
- **Result:** `Last Ran` → 2026-07-29, `Next Due` → 2026-08-12, `Status` → `done`.
- **Issues:** none.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `## Current Phase` section still points to Phase 1, which is fully complete. Phase 2 is now the active phase (Plan 40 v0.5.0 + Plan 41 shipped June 2026). | `Roadmap.md` — section heading + Phase 2 under "Future Phases" | Flagged for user — recommend promoting Phase 2 to "Current Phase" and moving Phase 1 to a "Completed" section. Do not modify directly. |
| 2 | Medium | Plan 41 (Headless CLI Contract + MCP-ready cores) shipped 2026-06-16, all 5 phases merged. No corresponding `[x]` row in Roadmap Phase 2. | `Roadmap.md` Phase 2 | Flagged for user — recommend adding: `[x] Headless CLI contract + MCP-ready cores — pure \`*Result\` cores, JSONL/exit-code contract, \`list --json\`, \`docs/agent-interface.md\` _(Plan 41, shipped 2026-06-16)_` |
| 3 | Low | Plan 42 (MCP server, bonsai mcp) is referenced in Status.md as the direct fast-follow to Plan 41 but appears nowhere in the roadmap. | `Roadmap.md` Phase 2/3 | Flagged for user — recommend adding: `[ ] MCP server (\`bonsai mcp\`) — thin wrapper over Plan 41 headless cores _(Plan 42, not started)_` to Phase 2 or Phase 3. |
| 4 | Low | Plans 40 and 41 introduced significant architectural decisions (frozen v1 schemas; MCP-is-a-wrapper pattern) that were never logged to the Key Decision Log. | `Logs/KeyDecisionLog.md` | Flagged for user — not a roadmap accuracy item, but relevant to future planning sessions. Consider adding these to KDL Structural section. |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

**Finding 1 — Roadmap structural drift (Current Phase header):**
Phase 1 is 100% complete. The roadmap should reflect that we are in Phase 2. Suggested edit: rename `## Current Phase` → `## Completed` or `## Phase 1 ✓`, and promote Phase 2 to `## Current Phase`. Low-effort change, high clarity gain.

**Finding 2 — Headless CLI not in roadmap:**
Plan 41 is a major shipped milestone (agent-drivable CLI, structured output contract, MCP-ready cores). It belongs in Phase 2 as a checked item so the roadmap accurately reflects what has been built. Suggested addition to Phase 2:
```
- [x] Headless CLI contract + MCP-ready cores — pure `*Result` cores, JSONL/exit-code
      contract, `list --json`, `docs/agent-interface.md` _(Plan 41, 2026-06-16)_
```

**Finding 3 — MCP server (Plan 42) not in roadmap:**
If Plan 42 is the next major workstream, it should appear in the roadmap. It fits in Phase 2 (extensibility) or Phase 3 (orchestration). User decision needed: which phase?

**Finding 4 — KDL hygiene:**
Consider adding two structural decisions to `Logs/KeyDecisionLog.md`:
- Plan 40: frozen v1 schemas + memory_dir security constraint
- Plan 41: MCP-as-thin-wrapper principle (one core, two serializers)

---

## Notes for Next Run

- Phase 2 should be the "Current Phase" by the next run — check that the structural edit was made.
- If Plan 42 has shipped by next run, add a `[x]` row to the roadmap for it.
- If Plans 40 and 41 are still in `Plans/Active/` at next run, flag for archival — both appear substantially complete per Status.md.
- Previous run (2026-05-07) flagged 2 items, both resolved by the 2026-05-07 routine digest. No carry-forward items from that run.
