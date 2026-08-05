---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-08-05
status: success
---

# Routine Report — Roadmap Accuracy

## Overview

Full audit of `station/Playbook/Roadmap.md` against actual build state (Status.md, Plans/Active/, KeyDecisionLog.md). Since the last run (2026-05-07), two major plans shipped — Plan 40 (Odysseus, v0.5.0 untagged) and Plan 41 (Headless CLI Contract, all 5 phases merged) — neither of which is reflected in the Roadmap. The "Current Phase" header remains on Phase 1 despite all checkboxes being checked. 7 findings flagged; no changes made (audit-only routine).

---

## Execution Metadata

| Field | Value |
|-------|-------|
| Routine | Roadmap Accuracy |
| Date | 2026-08-05 |
| Last run | 2026-05-07 |
| Execution mode | subagent (loop.md dispatch) |
| Files read | Roadmap.md, Status.md, KeyDecisionLog.md, Backlog.md, Plans/Active/ listing |
| Changes made | None (audit-only) |

---

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

**Phase 1 — Foundation & Polish:** All 11 checkboxes are marked `[x]`. Phase 1 is fully complete. The `bonsai validate` row (added by the 2026-05-07 routine-digest) is present and correct.

**"Current Phase" header:** Still reads "Phase 1 — Foundation & Polish." This is stale — Phase 1 is done, Phase 2 work has begun (custom item detection shipped, Plans 40/41 shipped), and the header should advance to Phase 2.

**Phase 2 — Extensibility:** The four listed items are partially correct:
- `[x] Custom item detection` — correctly marked done.
- `[ ] Self-update mechanism` — unbuilt. Partially overlapped by `bonsai validate` but not the same capability. Status correct.
- `[ ] Template variables expansion` — unbuilt. Status correct.
- `[ ] Micro-task fast path` — unbuilt. Status correct.

**Missing shipped items in Phase 2:** Two major plans shipped since the last roadmap accuracy run that add Phase 2 capability not captured anywhere on the roadmap:

- **Plan 40 (Odysseus, v0.5.0):** Frozen v1 schemas, root-relative scaffolding (manifest + memory), project-level `bonsai validate` pass with adversarial hardening, memory-routing docs. Phases 1–3 merged (PRs #114, #115, #116). Phase 4 (update-delivery) explicitly held by user.
- **Plan 41 (Headless CLI Contract):** All 5 phases merged (PRs #120–125). Every mutating command (init/add/update/remove) has a pure `*Result` headless core + JSONL/exit contract (`ExitConflict=5`); `list --json`; `docs/agent-interface.md` agent interface contract.

**Missing upcoming item:** Plan 42 (MCP server) is described in Status.md as a "fast-follow" to Plan 41 and is "not yet started." It has no roadmap entry.

**Phase 3 and Phase 4:** Items correctly reflect future state. No shipped work touches these items.

### Step 2 — Check milestone accuracy

**Phase 2 remaining items:** The three unchecked items (`self-update mechanism`, `template variables expansion`, `micro-task fast path`) remain unbuilt and are not superseded by recent plans. They are still valid future intentions.

**Phase 2 goal statement:** "Users can create custom catalog items, extend existing ones, and share them." Plans 40/41 don't fit this goal statement — they are about schema stability and machine/agent-drivable CLI rather than user-facing extensibility. The roadmap either needs new items for these capabilities, or its Phase 2 goal statement needs broadening to include "agent/automation extensibility."

**Plan 40 Phase 4 held:** The update-delivery capability (letting `bonsai init --non-interactive` work on existing projects) was explicitly held by the user and is not captured anywhere on the roadmap or in a named backlog item. It remains a gap.

**Plan 42 (MCP server):** Described as the natural bridge between the headless CLI contract (Plan 41) and cloud orchestration (Phase 3). Its placement — Phase 2 vs Phase 3 — is an open decision for the user.

### Step 3 — Cross-check KeyDecisionLog

The KeyDecisionLog has no entries more recent than 2026-04-13. Plans 40 and 41 shipped significant architectural decisions:
- Frozen v1 schema contract (Plan 40) — not logged.
- Headless CLI + JSONL/exit contract as a stable agent interface (Plan 41) — not logged.

No logged decisions invalidate any roadmap item. The "Defer Managed Agents cloud integration until local foundation is stable" decision (2026-04-02) remains consistent with Phase 3 items being unchecked.

**KeyDecisionLog staleness** is a finding but is out of scope for Roadmap Accuracy; noted here as a secondary flag.

### Step 4 — Report findings

See Findings Summary below. No modifications made to Roadmap.md — flagged for user review per routine procedure.

### Step 5 — Update dashboard

Updated `station/agent/Core/routines.md` — Roadmap Accuracy row: Last Ran → 2026-08-05, Next Due → 2026-08-19, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|---------|--------------|
| 1 | HIGH | "Current Phase" header still reads Phase 1 despite all 11 checkboxes checked — project is in Phase 2 | Roadmap.md, line 16 | Flagged for user review |
| 2 | HIGH | Plan 41 (Headless CLI Contract, agent-drivable cores, JSONL/exit contract, `docs/agent-interface.md`) shipped 2026-06-16 — not reflected in Phase 2 items | Roadmap.md Phase 2 | Flagged for user review |
| 3 | HIGH | Plan 40 (Odysseus: frozen v1 schemas, root-relative scaffolding, project-level validate pass) shipped 2026-06-13 — not reflected in Phase 2 items | Roadmap.md Phase 2 | Flagged for user review |
| 4 | MEDIUM | Plan 42 (MCP server) is a near-future "fast-follow" milestone not on the Roadmap; phase placement (P2 vs P3) is an open decision | Roadmap.md — missing | Flagged for user review |
| 5 | MEDIUM | Phase 2 goal statement ("Users can create custom catalog items…") doesn't cover Plans 40/41 scope (schema stability + agent-drivable CLI); may need broadening | Roadmap.md Phase 2 goal | Flagged for user review |
| 6 | LOW | Plan 40 Phase 4 (update-delivery for existing projects) explicitly held by user — gap not captured on Roadmap or in Backlog | Roadmap.md — missing | Flagged for user review |
| 7 | LOW | KeyDecisionLog has no entries since 2026-04-13; Plans 40/41 architectural decisions (v1 schema freeze, headless CLI contract) not logged | KeyDecisionLog.md | Out of scope — noted only |

---

## Errors & Warnings

None. All source files readable, procedure completed fully.

---

## Items Flagged for User Review

**F1 (HIGH) — Advance "Current Phase" to Phase 2**
All Phase 1 items are checked. The header should read "Phase 2 — Extensibility" with Phase 1 moved to a "Completed Phases" section or simply relabeled as done. Recommended action: reorganize Roadmap.md so Phase 2 is the Current Phase and Phase 1 is shown as complete.

**F2 (HIGH) — Add Plan 41 as a shipped Phase 2 item**
Suggested addition to Phase 2:
```
- [x] Agent-drivable CLI contract — headless *Result cores for init/add/update/remove, JSONL/exit contract (ExitConflict=5), list --json, docs/agent-interface.md _(Plan 41, PRs #120–125, 2026-06-16)_
```

**F3 (HIGH) — Add Plan 40 as a shipped Phase 2 item**
Suggested addition to Phase 2:
```
- [x] Platform integration groundwork (Odysseus) — frozen v1 schemas, root-relative scaffolding, project-level validate pass, memory-routing docs _(Plan 40, PRs #114–116, 2026-06-13; Phase 4 update-delivery held)_
```

**F4 (MEDIUM) — Place Plan 42 (MCP server) on the Roadmap**
Placement decision required: Phase 2 (Extensibility, following headless CLI) or Phase 3 (Cloud & Orchestration, as the bridge to Managed Agents)? Once decided, add an unchecked item.

**F5 (MEDIUM) — Phase 2 goal statement scope**
Current: "Users can create custom catalog items, extend existing ones, and share them."  
Consider broadening to: "Users can create custom catalog items, extend existing ones, and share them. Bonsai is programmatically drivable by agents and automation tools."

**F6 (LOW) — Plan 40 Phase 4 (update-delivery) gap**
The held capability — allowing `bonsai init --non-interactive` to work on existing projects — is not tracked anywhere. Options: (a) add to Backlog as a named P2 item, (b) add as an unchecked Phase 2 roadmap item with a "held" annotation, (c) leave as-is (user aware).

---

## Notes for Next Run

- The "Current Phase" header fix and Plans 40/41 additions are straightforward mechanical updates once the user approves.
- Plan 42 (MCP server) progress should be the primary check at the next run — if it ships, update accordingly.
- Phase 2 goal statement broadening should be validated against the user's intent before editing.
- KeyDecisionLog staleness (last entry 2026-04-13) should be raised separately — 14+ months of architectural decisions without log entries is a growing audit gap.
