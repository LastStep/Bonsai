---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-07
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~5 minutes
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/agent/Core/identity.md`, `station/agent/Core/memory.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/CLAUDE.md`, `station/Playbook/Status.md`, `station/code-index.md`, `station/agent/Workflows/plan-grilling.md`, `station/agent/Skills/critic-agent-prompts.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, file listing, link verification), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history
Read `station/INDEX.md`, `station/CLAUDE.md`, `station/Playbook/Status.md`, `station/code-index.md`. Pulled git log from 2026-05-04 to 2026-09-07 (~4 months). Key merged work since last run:
- **Plan 41 (2026-06-16):** Headless `*Result` cores for init/add/update/remove, `list --json`, `docs/agent-interface.md` contract (all 5 phases shipped, main `ab202c3`)
- **Plan 40 P1-3 (2026-06-13):** Frozen v1 schemas, root-relative scaffolding, project-level `validate` pass, `docs/formats.md` (guide Formats page), memory-routing protocol
- **Plan 40 Phase 4:** Still held pending dogfood delivery path
- **External contrib (2026-05-07):** `bonsai completion` subcommand added via PR #78
- **v0.4.3 hotfix (2026-05-13):** Sensor hook absolute path baking

### Step 2 — Check INDEX.md accuracy
- Tech stack table: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS)
- Folder structure: accurate — all described directories exist
- CLI command count: listed as "8 (init, add, remove, list, catalog, update, guide, validate)"; `bonsai completion` (added PR #78) is an unlisted 9th. Minor gap.
- Catalog item count: listed as "~50 (skills, workflows, protocols, sensors, routines)"; actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). Still within ~50 estimate range.
- Document registry: does not reference `docs/agent-interface.md` (Plan 41) or `docs/formats.md` (Plan 40 Phase 3). Both are user-facing docs shipped to main.

### Step 3 — Check navigation links
Verified all 38 file paths referenced in `station/CLAUDE.md` navigation tables. All resolved successfully. No broken links.

Discovered two files present on disk that are NOT listed in the navigation tables:
- `station/agent/Workflows/plan-grilling.md` — active workflow, used in the grilling pipeline, added 2026-06-13
- `station/agent/Skills/critic-agent-prompts.md` — companion skill for plan-grilling, added 2026-06-13

### Step 4 — Report findings
See Findings Summary below. All items flagged for user review per procedure (no edits to doc content executed).

### Step 5 — Update dashboard
Updated `agent/Core/routines.md` row for Doc Freshness Check: Last Ran → 2026-09-07, Next Due → 2026-09-14, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `plan-grilling.md` workflow exists on disk but is not listed in the CLAUDE.md Workflows navigation table — agent may not know to load it | `station/agent/Workflows/plan-grilling.md` | Flagged for user |
| 2 | Medium | `critic-agent-prompts.md` skill exists on disk but is not listed in the CLAUDE.md Skills navigation table — referenced by plan-grilling workflow | `station/agent/Skills/critic-agent-prompts.md` | Flagged for user |
| 3 | Medium | Plan 41 (all 5 phases shipped 2026-06-16) still lives in `Plans/Active/`. Memory.md note says "archive at next wrap-up." | `station/Playbook/Plans/Active/41-headless-cli-contract.md` | Flagged for user |
| 4 | Low | `docs/agent-interface.md` (Plan 41 Phase 5) and `docs/formats.md` (Plan 40 Phase 3) are shipped user-facing docs not listed in INDEX.md Document Registry | `station/INDEX.md` | Flagged for user |
| 5 | Low | INDEX.md CLI command count reads "8 (init, add, remove, list, catalog, update, guide, validate)" — `bonsai completion` (PR #78, external contrib) is an unlisted 9th command | `station/INDEX.md` | Flagged for user |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

**Finding 1 — Add `plan-grilling.md` to CLAUDE.md Workflows table**

Proposed addition to the Workflows table in `station/CLAUDE.md`:

```
| Grilling a drafted plan adversarially (6-critic loop) before dispatch | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

**Finding 2 — Add `critic-agent-prompts.md` to CLAUDE.md Skills table**

Proposed addition to the Skills table in `station/CLAUDE.md`:

```
| Prompt templates for 6 plan-grilling critic agents — dispatched verbatim by plan-grilling.md | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

**Finding 3 — Archive Plan 41**

Move `station/Playbook/Plans/Active/41-headless-cli-contract.md` → `station/Playbook/Plans/Archive/41-headless-cli-contract.md`.
Already captured in memory.md as a "next wrap-up" action — can be combined with the next session's wrap-up, or done now.

**Finding 4 — Add new docs/ entries to INDEX.md Document Registry**

Proposed additions to `station/INDEX.md` Document Registry table:

```
| `docs/agent-interface.md` | Headless CLI contract — JSONL event format, exit codes, flags for all mutating commands (init/add/update/remove/list) | When scripting Bonsai or wiring up the MCP server (Plan 42) |
| `docs/formats.md` | Guide: supported workspace formats (memory-routing paths, scaffolding layout) | When explaining workspace layout to new agents or users |
```

**Finding 5 — Update CLI command count in INDEX.md**

Change `CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate)` to `9 (init, add, remove, list, catalog, update, guide, validate, completion)` in the Key Metrics table.

---

## Notes for Next Run

- This run covered ~4 months of commits (last run 2026-05-04, today 2026-09-07). Routine should return to 7-day cadence.
- Plan 40 Phase 4 still held — if it ships before the next run, `docs/formats.md` may need updating.
- All 38 navigation links verified clean — no structural link rot detected.
- Consider adding `docs/` directory to INDEX.md as a tracked docs section (currently only `station/` docs are indexed).
