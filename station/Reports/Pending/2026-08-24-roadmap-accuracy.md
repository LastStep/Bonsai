---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-08-24
status: partial
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~5 min
- **Files Read:** `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Logs/KeyDecisionLog.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log append)
- **Roadmap edits:** none — findings flagged for user review per routine rules

---

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

**Phase 1 — Foundation & Polish:** All 11 items are marked `[x]`. Cross-checked against Status.md, RoutineLog.md, and KeyDecisionLog.md:

| Item | Roadmap Status | Actual State |
|------|---------------|--------------|
| Go rewrite | ✅ done | Shipped early 2026-04 |
| Full catalog | ✅ done | Confirmed |
| Lock file conflict handling | ✅ done | Confirmed |
| Awareness Framework | ✅ done | Confirmed |
| Dogfooding | ✅ done | Confirmed (station/ workspace is live) |
| Better trigger sections | ✅ done | Fixed by 2026-05-07 Routine Digest (annotation added) |
| UI overhaul | ✅ done | Confirmed |
| Usage instructions | ✅ done | Confirmed |
| Release pipeline | ✅ done | GoReleaser + GitHub Actions + Homebrew active |
| Community health files | ✅ done | Confirmed |
| `bonsai validate` | ✅ done | v0.4.0, Plan 35 |

Phase 1 is fully complete and accurately reflected. No drift.

**Phase 2 — Extensibility:**

| Item | Roadmap Status | Actual State |
|------|---------------|--------------|
| Custom item detection | ✅ done | Confirmed (Plan 34) |
| Self-update mechanism | ⬜ pending | No plan exists — still future work |
| Template variables expansion | ⬜ pending | **Partially addressed** — Plan 40 extended scaffolding template context (root-relative paths, memory routing). Broader "richer context in all templates" is not complete. Needs user clarification. |
| Micro-task fast path | ⬜ pending | No plan exists — still future work |

**Phase 3 — Cloud & Orchestration:** Both items remain unchecked and no work has started on Managed Agents or Greenhouse.

**Phase 4 — Ecosystem:** All items unchecked, no work started. Still distant.

### Step 2 — Check milestone accuracy

Two significant bodies of work shipped since the last roadmap accuracy run (2026-05-07) that are **not represented anywhere in Roadmap.md**:

**Plan 40 — Odysseus Integration (v0.5.0, 2026-06-13):**
- Frozen v1 schemas for scaffolding
- Root-relative scaffolding manifest + memory routing
- Project-level `bonsai validate` pass with adversarial path/symlink hardening (security hardening)
- Memory-routing docs + guide Formats page
- Tag held by user as of 2026-06-13; Phase 4 (update-delivery for existing projects) deferred

**Plan 41 — Headless CLI Contract + MCP-ready cores (2026-06-16):**
- Every mutating command (init/add/update/remove) has a pure `*Result` headless core
- JSONL/exit contract with named exit codes (e.g. `ExitConflict=5`)
- `list --json` output
- `docs/agent-interface.md` formal contract document
- "MCP server = fast-follow Plan 42" noted in Status.md — Phase 3-adjacent work
- Debt filed: unify remove cinematic/headless logic (Backlog P2)

Neither of these ships is reflected in the roadmap. This is the most significant drift found.

Additionally, **Plan 42 (MCP server) is referenced as an active fast-follow** from Plan 41 but does not appear in the roadmap. Phase 3 mentions "Managed Agents integration" with a `bonsai deploy` frame, but the actual trajectory appears to be MCP server first (Plan 42), then fuller cloud integration. The roadmap Phase 3 framing may be outdated.

### Step 3 — Cross-check against Key Decision Log

All structural and domain-specific decisions remain consistent with the roadmap:
- "Bonsai is a scaffolding tool, not a runtime orchestrator" — still consistent with Phase 3 framing (deploying workspaces, not running code directly)
- "Defer Managed Agents cloud integration until local foundation is stable" — foundation has matured significantly (v0.4.x → v0.5.0 + headless CLI), but no formal re-evaluation has occurred. The condition "local foundation is stable" is arguably now met.
- No decision in the log contradicts or supersedes any open roadmap item.
- The headless CLI (Plan 41) is an enabler for MCP integration. Its presence accelerates Phase 3 but doesn't invalidate it.

### Step 4 — Report findings

Findings listed below. Not modifying Roadmap.md directly per routine rules — flagging for user review.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | Plan 40 (v0.5.0 Odysseus: schemas, scaffolding, validate hardening, docs) shipped 2026-06-13 has no roadmap representation | Roadmap.md Phase 2 | Flagged for user review |
| 2 | MEDIUM | Plan 41 (Headless CLI + MCP-ready cores, JSONL contract, agent-interface.md) shipped 2026-06-16 has no roadmap representation | Roadmap.md Phase 2 or Phase 3 | Flagged for user review |
| 3 | MEDIUM | Plan 42 (MCP server, "fast-follow" from Plan 41) is an active near-term item with no roadmap row | Roadmap.md Phase 3 | Flagged for user review |
| 4 | LOW | Phase 3 "Managed Agents integration" framing (`bonsai deploy`, session management, rubrics) may be outdated — actual trajectory is MCP server first (Plan 42), then fuller Managed Agents | Roadmap.md Phase 3 | Flagged for user review |
| 5 | LOW | Phase 2 "Template variables expansion" is ambiguous — Plan 40 partially addressed scaffolding template context; unclear if the broader item is still open | Roadmap.md Phase 2 | Flagged for user review |
| 6 | INFO | v0.5.0 tag held (user decision 2026-06-13) — roadmap has no version milestones so this is informational only | Status.md | No action |
| 7 | INFO | Phase 1 fully accurate — no drift | Roadmap.md Phase 1 | None needed |

---

## Errors & Warnings

None — routine completed cleanly.

---

## Items Flagged for User Review

### FLAG 1 (MEDIUM) — Plan 40 + Plan 41 missing from roadmap

Both are significant shipped capabilities:
- **Plan 40**: frozen v1 schemas, root-relative scaffolding, security-hardened validate pass, memory routing + docs (v0.5.0 scope)
- **Plan 41**: headless CLI contract for all mutating commands, JSONL/exit-code API, `docs/agent-interface.md`, MCP-ready architecture

**Suggested action:** Add rows to Phase 2 under a new sub-theme like "Machine-readable API & platform stability" covering schema freezing, headless cores, and the agent interface contract.

### FLAG 2 (MEDIUM) — Plan 42 (MCP server) has no roadmap row

Status.md notes "MCP server = fast-follow Plan 42." This is a Phase 3-adjacent item (cloud/orchestration enabler) but hasn't been planned yet. If Phase 3 is the right home, the current framing there ("Managed Agents integration — `bonsai deploy`") should be updated to reflect MCP server as the first deliverable.

### FLAG 3 (LOW) — Phase 3 framing may be stale

The settled decision says "Defer Managed Agents cloud integration until local foundation is stable." That condition is now closer to met. Consider revisiting Phase 3's sequencing: MCP server (Plan 42) → then Managed Agents integration.

### FLAG 4 (LOW) — Phase 2 "Template variables expansion" — partial or complete?

Plan 40 added scaffolding-specific template context (root-relative paths, memory routing). If this satisfies the intent, check the box with an annotation. If broader template richness (e.g. for skills/workflows/protocols) is still wanted, leave it open.

---

## Notes for Next Run

- If user adds roadmap rows for Plans 40/41/42, verify they're accurate before next cycle.
- Phase 3 framing re-evaluation is pending user decision — carry forward to next run if unresolved.
- Phase 2 remaining open items (self-update mechanism, micro-task fast path) have no plans — still legitimately future work, not abandoned.
- Last routine-digest that processed a roadmap-accuracy report was 2026-05-07 — that digest applied 2 quick fixes directly. Same approach recommended for FLAG 4 if user decides "Template variables expansion" is complete.
