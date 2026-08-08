---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-08-08
status: partial
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07 (before this run) — gap of 93 days
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Files read:** `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`
- **Files modified:** `station/agent/Core/routines.md` (dashboard row), `station/Logs/RoutineLog.md` (append)
- **Roadmap edits:** None — per procedure, findings are flagged for user review only

---

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

**Phase 1 — Foundation & Polish:** All 11 items are marked `[x]`. Cross-checking against Status.md and the archive, every deliverable is confirmed shipped:
- Go rewrite, full catalog, lock file, awareness framework, dogfooding — all shipped before last run.
- UI overhaul, usage instructions, release pipeline, community health files, `bonsai validate` — confirmed shipped.
- "Better trigger sections" — annotated with the deferred Plan 08 C3 caveat. Annotation is accurate.

**Phase 1 conclusion:** Complete and accurate.

**Phase 2 — Extensibility status check:**
- `[x]` Custom item detection — confirmed shipped (catalog scan + untracked-custom detection).
- `[ ]` Self-update mechanism — not shipped. Backlog P3 entry matches.
- `[ ]` Template variables expansion — not shipped. No Backlog tracking entry found; may be an informal placeholder.
- `[ ]` Micro-task fast path — not shipped. Backlog P3 entry matches.

**Phase 2 conclusion:** 1 of 4 items done. The unchecked items are accurately marked incomplete. However, the "Current Phase" header still reads "Phase 1 — Foundation & Polish" despite Phase 1 being fully complete. The Roadmap's current-phase framing is stale.

**Phase 3 — Cloud & Orchestration:**
- Both items (`Managed Agents integration`, `Greenhouse companion app`) remain incomplete. Accurate.

**Phase 4 — Ecosystem:** All 3 items incomplete. Accurate.

### Step 2 — Check milestone accuracy

Since the last roadmap accuracy run (2026-05-07), two significant plans shipped that are **not reflected anywhere in the Roadmap**:

**Plan 41 — Headless CLI Contract + MCP-ready cores (shipped 2026-06-16):**
- All four mutating commands (`init`/`add`/`update`/`remove`) now have pure `*Result` headless cores.
- JSONL/exit contract (ExitConflict=5), `list --json`, `docs/agent-interface.md` contract doc.
- Status.md notes: "MCP server = fast-follow Plan 42."
- This is a material Phase 2/Phase 3 deliverable (unlocks agent-drivability and MCP integration) with no Roadmap entry.

**Plan 40 — Odysseus Platform Integration Phases 1–3 (shipped 2026-06-13):**
- Frozen v1 schemas, root-relative scaffolding (manifest + memory), project-level `bonsai validate` pass with adversarial path/symlink hardening, memory-routing docs, guide Formats page.
- Phase 4 is HELD; dogfood deferred (locked .bonsai-lock.yaml gitignore issue); tag held (user).
- None of these deliverables map to a named Roadmap item.

**bonsai completion command (shipped 2026-05-07, PR #78):**
- Shell completion for bash/zsh/fish/powershell — a usability milestone not tracked in the Roadmap. Minor omission.

**Milestone accuracy conclusion:**
- The "Current Phase" annotation is stale — Phase 1 is done; the project has transitioned into Phase 2 work.
- Two major shipped plans (40, 41) are invisible in the Roadmap.
- Phase 3 preconditions (headless contract enabling MCP) are now materially closer but Roadmap does not reflect this progress.

### Step 3 — Cross-check against Key Decision Log

Reviewed all three sections (Structural, Domain-Specific, Settled). Findings:

- **"Defer Managed Agents cloud integration until local foundation is stable"** (Settled, 2026-04-02) — Plan 41's headless CLI contract and MCP-ready cores were explicitly designed as a Phase 3 precondition ("MCP server = fast-follow Plan 42"). The "local foundation is stable" condition may now be satisfied. No KDL entry supersedes or updates this decision.
- No KDL decisions invalidate any current Roadmap item.
- No KDL decisions conflict with Phase 2 unchecked items.

**KDL conclusion:** No invalidations. One settled decision worth user re-evaluation given Plan 41's progress.

### Step 4 — Findings (see below)

### Step 5 — Dashboard + log updated

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | Roadmap "Current Phase" still reads "Phase 1" despite Phase 1 being 100% complete | Roadmap.md header | Flagged for user — no edit made |
| 2 | MEDIUM | Plan 41 (Headless CLI contract + MCP-ready cores, shipped 2026-06-16) has no Roadmap entry | Roadmap.md Phase 2/3 | Flagged for user — no edit made |
| 3 | MEDIUM | Plan 40 Phases 1–3 (v1 schema freeze, root-relative scaffolding, project-level validate, shipped 2026-06-13) has no Roadmap entry | Roadmap.md Phase 2 | Flagged for user — no edit made |
| 4 | LOW | "bonsai completion" shell completion (shipped 2026-05-07) not tracked in any Roadmap phase | Roadmap.md | Flagged for user — minor omission |
| 5 | LOW | "Template variables expansion" in Phase 2 has no Backlog tracking entry to anchor it | Backlog.md / Roadmap.md | Flagged for user — may be a vague placeholder |
| 6 | INFO | KDL "defer Managed Agents until local foundation stable" may warrant revisiting — Plan 41 delivered the headless contract that unblocks MCP (Plan 42) | KeyDecisionLog.md Settled | Flagged for user consideration |

---

## Errors & Warnings

- **None.** All referenced files read successfully. Procedure completed as specified.

---

## Items Flagged for User Review

### 1. Update "Current Phase" header in Roadmap.md
Phase 1 is 100% done. The roadmap still presents Phase 1 as the "Current Phase" section. Suggested: move Phase 2 — Extensibility under the `## Current Phase` heading and push Phase 1 into a `## Completed Phases` or `## Phase History` section. Decision is yours — some teams prefer to keep the done phase visible at the top as a completed record.

### 2. Add Plan 41 to the Roadmap
Plan 41 shipped a significant architectural layer: all four mutating commands now have headless `*Result` cores, JSONL/exit contract (ExitConflict=5), `list --json`, and `docs/agent-interface.md`. This is the foundation for the MCP server (Plan 42). Suggested entry somewhere in Phase 2 or Phase 3:
```
- [x] Headless CLI contract — all mutating commands expose pure *Result cores + JSONL/exit contract; `docs/agent-interface.md` defines the stable agent interface. Unblocks MCP server.
```

### 3. Add Plan 40 deliverables to the Roadmap (or acknowledge as infrastructure)
Plan 40 Phases 1–3 shipped: v1 schema freeze, root-relative scaffolding, project.yaml hub integration, hardened `bonsai validate` with adversarial path guards. Phase 4 (dogfood gate) is on hold. Suggested: either add as a Phase 2 row or acknowledge in a "Released" section.

### 4. Reconsider "Template variables expansion" wording
The Phase 2 item `[ ] Template variables expansion — richer context available in templates` has no Backlog entry and no plan reference. If this refers to template context vars (`.ProjectName`, `.OtherAgents`, etc.), the available set has grown organically. Recommend either: (a) define what "richer" means and add a Backlog item with scope, or (b) mark it `[x]` with an annotation if you consider the current TemplateContext sufficient.

### 5. Revisit Managed Agents "defer" decision (Optional)
Plan 41's headless CLI contract was explicitly designed as the MCP server precondition. The "local foundation is stable" condition cited in the KDL for deferring Phase 3 appears to be met. This may be the right time to revisit the Settled decision and open Phase 3 planning.

---

## Notes for Next Run
- Re-run is due 2026-08-22. If the user acts on any of the flagged items before then, those can be cleared at the top of the next run.
- The 93-day gap since the last run resulted in two major plans (40, 41) being untracked in the Roadmap. 14-day cadence should prevent this level of drift going forward.
- If "bonsai completion" warrants a Roadmap entry, consider adding a "Phase 1 Supplemental" or noting it under Phase 1 as part of the OSS Polish milestone.
