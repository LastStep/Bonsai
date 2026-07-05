---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-07-05
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
- **Duration:** ~7 minutes
- **Files Read:** 5
  - `station/agent/Routines/roadmap-accuracy.md`
  - `station/Playbook/Roadmap.md`
  - `station/Playbook/Status.md`
  - `station/Logs/KeyDecisionLog.md`
  - `station/Logs/RoutineLog.md`
- **Files Modified:** 2
  - `station/agent/Core/routines.md` — dashboard row updated (Last Ran, Next Due, Status)
  - `station/Logs/RoutineLog.md` — entry appended
- **Tools Used:** Read, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `station/Playbook/Roadmap.md` in full.

**Phase 1 — Foundation & Polish:** All 11 items are marked `[x]`. Cross-referencing against Status.md "Recently Done":
- The 2026-05-07 Routine Digest applied two quick fixes: checked "Better trigger sections" with annotation, and added a `bonsai validate` row. Both are correctly reflected in the current Roadmap.
- No Phase 1 checkbox drift detected. Phase 1 is accurately marked complete.

**Phase 2 — Extensibility:**
- `[x] Custom item detection` — accurate; shipped via Plan 34, confirmed in prior routine cycles.
- `[ ] Self-update mechanism` — still not shipped. No active work tracked.
- `[ ] Template variables expansion` — still not shipped. No active work tracked.
- `[ ] Micro-task fast path` — still not shipped. No active work tracked.

**Phase 3 — Cloud & Orchestration:** Both items remain `[ ]`. No changes.

**Phase 4 — Ecosystem:** All items remain `[ ]`. No changes.

**New work since 2026-05-07 not represented in the roadmap:**

Two significant deliverables have shipped since the last run that do not map to any existing roadmap line item:

1. **Plan 40 — Odysseus Platform Integration** (v0.5.0, Phases 1–3, 2026-06-13): Shipped frozen v1 schemas, root-relative scaffolding, project-level `validate` pass with adversarial path/symlink hardening, and memory-routing docs + guide Formats page. Phase 4 held. This work is extensibility infrastructure but has no corresponding roadmap item.

2. **Plan 41 — Headless CLI Contract + MCP-ready cores** (2026-06-16): All 5 phases merged. Every mutating command (`init`/`add`/`update`/`remove`) has a pure `*Result` headless core + JSONL/exit contract (`ExitConflict=5`); `list --json`; `docs/agent-interface.md` contract document. Status.md notes: "MCP server = fast-follow Plan 42." This is a major architectural milestone with no roadmap representation.

### Step 2 — Check milestone accuracy

Are the next milestones still the right priority? The remaining Phase 2 items (self-update mechanism, template variables expansion, micro-task fast path) have not been touched and do not appear in Status.md Pending or any active plan. Meanwhile, the actual development trajectory has moved into headless CLI + MCP-server territory (Plans 41/42), which sits between Phase 2 extensibility and Phase 3 cloud/orchestration.

The Plan 42 MCP server (referenced as "fast-follow" in Status.md) would logically belong in Phase 3 ("Cloud & Orchestration"), but the roadmap makes no mention of an MCP server. The Phase 3 item currently reads: "Managed Agents integration — `bonsai deploy`, session management, outcome rubrics" — which is a different surface area from an MCP server.

No roadmap items reference deprecated approaches. No planned work has been superseded by conflicting decisions.

### Step 3 — Cross-check against Key Decision Log

Read `station/Logs/KeyDecisionLog.md` in full.

Relevant decision: **"2026-04-13 — Defer Managed Agents cloud integration until local foundation is stable."**

Status: The headless CLI contract (Plan 41) and forthcoming MCP server (Plan 42) represent foundational work that directly enables Managed Agents integration. This is natural progression — the decision to defer cloud integration until the local foundation is stable is effectively being honoured: Plans 40 and 41 built the local foundation (v1 schemas, headless cores), and Plan 42 (MCP) is the next logical step toward cloud/orchestration. However, the roadmap's Phase 3 description does not reflect the MCP-server implementation path.

No other decisions in the log invalidate or contradict current roadmap items.

### Step 4 — Report findings

Per procedure: findings flagged for user review. No modifications made to `Roadmap.md`.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | Plan 41 (Headless CLI Contract + MCP-ready cores) shipped 2026-06-16 with no corresponding roadmap entry | `Roadmap.md` Phase 2/3 | Flagged for user — recommend adding as Phase 2 `[x]` item or Phase 3 precursor |
| 2 | MEDIUM | MCP server (Plan 42, referenced as "fast-follow" in Status.md) has no roadmap entry | `Roadmap.md` Phase 3 | Flagged for user — recommend adding to Phase 3 alongside or replacing/augmenting "Managed Agents integration" |
| 3 | LOW | Phase 3 "Managed Agents integration" description (`bonsai deploy`, session management, outcome rubrics) does not acknowledge the MCP-server approach now in motion | `Roadmap.md` Phase 3 | Flagged for user — description may need updating once Plan 42 scope is clear |
| 4 | INFO | Plan 40 "Odysseus Platform Integration" (v0.5.0, Phases 1–3) has no named roadmap line item | `Roadmap.md` Phase 2 | Informational — the work is extensibility infrastructure; could be captured as a Phase 2 item or left implicit |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

### Flag 1 (MEDIUM) — Headless CLI Contract not in roadmap

Plan 41 shipped a significant architectural feature that has no roadmap representation:
- Pure `*Result` headless cores for all mutating commands
- JSONL/exit-code contract (`ExitConflict=5`)
- `docs/agent-interface.md` contract document

**Suggested roadmap addition (Phase 2 or as Phase 3 precursor):**
```
- [x] Headless CLI contract — programmatic/agent-facing API with JSONL output and exit-code contract for all mutating commands
```

### Flag 2 (MEDIUM) — MCP server (Plan 42) not in roadmap

Status.md explicitly notes "MCP server = fast-follow Plan 42." This is active near-term work with no roadmap entry. It would logically belong in Phase 3 alongside "Managed Agents integration."

**Suggested roadmap addition (Phase 3):**
```
- [ ] MCP server — Claude Code MCP interface over the headless CLI contract; enables agent-native tool calls into Bonsai
```

### Flag 3 (LOW) — Phase 3 description misalignment

The Phase 3 "Managed Agents integration" item currently describes `bonsai deploy`, session management, and outcome rubrics — which may evolve or be partially superseded by the MCP-server approach. Once Plan 42 ships, the Phase 3 description should be reconciled to reflect the actual implementation path.

No action required now; revisit at the next roadmap accuracy run after Plan 42 is complete.

---

## Notes for Next Run

- Check whether Plan 42 (MCP server) has shipped and add a `[x]` entry to Phase 3.
- If the user has updated the roadmap based on Flags 1 and 2 from this report, verify those additions are accurate.
- Confirm whether any Phase 2 unchecked items (self-update mechanism, template variables expansion, micro-task fast path) have moved to active work.
- v0.5.0 tag was held by user as of 2026-06-13 — check whether it has since been cut.
- HOMEBREW_TAP_TOKEN PAT expires ~2026-07-15 (flagged by backlog-hygiene 2026-07-05) — unrelated to roadmap but worth noting for context.
