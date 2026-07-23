---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-07-23
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
- **Files Read:** 7 — `station/agent/Routines/roadmap-accuracy.md`, `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/Playbook/Backlog.md`, `station/Playbook/Plans/Active/40-odysseus-platform-integration.md`, `station/Playbook/Plans/Active/41-headless-cli-contract.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Write, Edit, Glob, Bash
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state

**Action:** Read `Roadmap.md` and compared all Phase 1 items against Status.md and active plans.

**Result:** Phase 1 is 100% complete — all 11 items carry `[x]`. The last items to land were `bonsai validate` (Plan 35, v0.4.0, 2026-05-04) and the release pipeline / community health files. Since the 2026-05-07 roadmap-accuracy run, the roadmap was updated to resolve both prior flags: "Better trigger sections" is now checked with an annotation, and a `bonsai validate` row was added. Both previous findings are closed.

**Issue:** The roadmap still bears the heading `## Current Phase` over Phase 1, despite Phase 1 being entirely done. Active development has moved into Phase 2 territory (Plans 40, 41 shipped). The `Current Phase` label is stale and misdirects anyone reading the roadmap cold.

---

### Step 2: Check milestone accuracy

**Action:** Reviewed Phase 2, 3, and 4 against Status.md, active plans, and Backlog.

**Result (Phase 2 — Extensibility):**

- `[x] Custom item detection` — correctly marked done (Plan 34, v0.4.0).
- `[ ] Self-update mechanism` — correctly unstarted; lives in P3 Backlog.
- `[ ] Template variables expansion` — correctly unstarted; no Backlog entry, no active work.
- `[ ] Micro-task fast path` — correctly unstarted; lives in P3 Backlog.

**Issue — Major gap:** **Plan 41 (Headless CLI Contract + MCP-ready cores, shipped 2026-06-16)** is a Phase 2-class milestone — "every mutating command runs fully non-interactive behind a pure `Result`-returning core; every read command emits structured JSON; all seven share one exit-code contract and a documented event/schema philosophy" — yet it has **no representation anywhere in the roadmap**. The plan was also architected to make `bonsai mcp` (Plan 42) a thin wrapper with zero duplicated work. This is a shipped capability of comparable weight to `bonsai validate` (which was added to the roadmap as a Phase 1 headline).

**Issue — Untracked shipped work:** **Plan 40 Phases 1–3 (v0.5.0, untagged, shipped 2026-06-13)** shipped: frozen v1 schemas for in-repo memory graphs (`station/Memory/`), per-repo project manifest (`.bonsai/project.yaml`), adversarial-hardened project-level `validate` pass, and memory-routing documentation. None of this appears in the roadmap. Plan 40 Phase 4 (update-delivery path for new scaffolding items) is **HELD** — also unrepresented.

**Issue — Planned fast-follow missing:** **Plan 42 (MCP server)** is referenced in Plan 41's context as the immediate next step ("MCP server = fast-follow Plan 42") and in Status.md. Plan 42 is not in the roadmap, not in the Backlog, and has no plan file. For a feature that is explicitly "next" and already architecturally pre-positioned by Plan 41, its absence from planning artifacts is a gap.

**Result (Phase 3 — Cloud & Orchestration):**

- `[ ] Managed Agents integration` — Backlog P3 "Big Bets", correctly deferred. Decision log confirms: "Defer Managed Agents cloud integration until local foundation is stable."
- `[ ] Greenhouse companion app` — Backlog P3 "Big Bets" with design-phase note. Consistent with roadmap.

No Phase 3 or 4 drift. MCP server (Plan 42) would logically belong here if/when it's added.

---

### Step 3: Cross-check against Key Decision Log

**Action:** Read `KeyDecisionLog.md` and checked all decisions against roadmap items.

**Result:** No decisions invalidate any roadmap line. Key cross-checks:

- "Bonsai is a scaffolding tool, not a runtime orchestrator" — consistent with all phases.
- "Defer Managed Agents cloud integration until local foundation is stable" — Phase 3 is correctly future-gated.
- "Catalog embedded via `embed.FS`" / "Go text/template" / "Lock file with SHA-256 hashes" — all Phase 1 foundation items, correctly `[x]`.
- Two-sensor awareness framework — Phase 1 `[x]`, correct.

No decision log entries suggest any roadmap item is deprecated or should be removed. The Phase 3/4 items are still the intended long-term direction.

---

### Step 4: Report findings

Findings reported below. No edits made to `Roadmap.md` — all items flagged for user review.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | **High** | `## Current Phase` header over Phase 1 is stale — all 11 items `[x]`, active work is in Phase 2 | `Roadmap.md` L14 | Flagged for user — suggest relabeling Phase 1 as completed, Phase 2 as "Current Phase" |
| 2 | **High** | Plan 41 (Headless CLI Contract + MCP-ready cores, shipped 2026-06-16) has no roadmap representation | `Roadmap.md` — missing entry | Flagged for user — suggest adding a Phase 2 line: `[x] Headless CLI contract — non-interactive cores for all mutating commands; structured JSON output; JSONL/exit-code contract (agent-interface.md). MCP-ready substrate.` |
| 3 | **Medium** | Plan 40 Phases 1-3 (in-repo memory graph, project manifest, adversarial validate hardening, shipped 2026-06-13) have no roadmap representation | `Roadmap.md` — missing entry | Flagged for user — consider adding a Phase 2 or Phase 1 polish line covering project manifest + memory routing |
| 4 | **Medium** | Plan 42 (MCP server, fast-follow to Plan 41) not in roadmap, not in Backlog, no plan file | Missing from all planning artifacts | Flagged for user — suggest adding to Phase 3 and/or creating a Plan 42 Backlog entry |
| 5 | **Low** | Phase 2 should be promoted to "Current Phase" once Phase 1 is relabeled complete | `Roadmap.md` section structure | Flagged for user |
| 6 | **Low** | v0.5.0 untagged (Plan 40 + Plan 41 represent two significant releases since v0.4.3 2026-05-13) | Status.md / release pipeline | Off-roadmap but noted — release cadence has stalled vs the roadmap's "Release pipeline" completion. No roadmap item tracks version milestones. |

---

## Errors & Warnings

No errors encountered.

**Operational flag (off-roadmap):** The Backlog P1 item for `HOMEBREW_TAP_TOKEN PAT expiry` set a reminder date of 2026-07-15. Today is 2026-07-23 — **8 days past the reminder**. If the PAT has not been rotated, the next GoReleaser release will silently fail at the Homebrew formula update step (`401 Bad credentials`). This was also flagged by the 2026-07-23 Backlog Hygiene run.

---

## Items Flagged for User Review

1. **Relabel Phase 1 complete, promote Phase 2 to "Current Phase"** — Phase 1 is done. The `Current Phase` label is misleading. Suggest: rename heading to `## Completed — Phase 1: Foundation & Polish`, add `## Current Phase — Phase 2: Extensibility`.

2. **Add Plan 41 to Phase 2** — Headless CLI Contract + MCP-ready cores is a significant Phase 2 milestone. Suggested line:
   ```
   - [x] Headless CLI contract — non-interactive cores for all mutating commands, JSONL/exit-code contract, structured JSON reads. MCP-ready substrate. _(Plan 41, 2026-06-16)_
   ```

3. **Add Plan 40 deliverables to Roadmap** — The project manifest (`.bonsai/project.yaml`), in-repo memory routing (`station/Memory/`), and adversarial-hardened validate are shipped features. Consider a Phase 2 line or a Phase 1 expansion entry:
   ```
   - [x] Project manifest + memory routing — `.bonsai/project.yaml`, `station/Memory/` scaffolding, frozen v1 schemas, adversarial validate hardening. _(Plan 40, 2026-06-13)_
   ```

4. **Add Plan 42 (MCP server) to Roadmap Phase 3 + Backlog** — Plan 42 is explicitly "next" per Plan 41's architecture. Suggest adding to Phase 3:
   ```
   - [ ] MCP server — `bonsai mcp` thin wrapper over headless cores; native AI-in-editor integration
   ```
   And a Backlog P1 entry tracking the plan creation.

5. **Rotate HOMEBREW_TAP_TOKEN PAT** — Reminder date 2026-07-15 is 8 days past. Rotate before next release. (Flagged by Backlog Hygiene too.)

---

## Notes for Next Run

- The two flags from the 2026-05-07 run are fully resolved (Better trigger sections checked + annotated, `bonsai validate` row added). No carry-forwards from that run.
- If Plans 40/41 entries are added to the roadmap, the next run's Phase 2 accuracy check will be cleaner.
- If Plan 42 is created and added to Backlog + Roadmap, track whether it has shipped by next run (14 days = 2026-08-06).
- v0.5.0 tag status: if still untagged at next run, this is worth escalating — 60+ days without a version tag creates confusion between "latest stable" and "main".
