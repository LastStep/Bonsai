---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-20
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
- **Duration:** ~8 min
- **Files Read:** 9 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Status.md`, `station/CLAUDE.md`, `Bonsai/CLAUDE.md`, `station/code-index.md`, `station/agent/Skills/critic-agent-prompts.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Reports/Pending/2026-06-20-doc-freshness-check.md`
- **Tools Used:** Read, Bash (git log, ls, grep, test), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history

Reviewed git log from 2026-05-04 to 2026-06-20 (47 days). Major code changes in this period:

- **Plan 39** (2026-05-13): `bonsai init/add --non-interactive --from-config` (v0.4.2)
- **v0.4.3 hotfix** (2026-05-13): baked absolute paths into sensor hook commands
- **Plan 40 Phases 1–3** (2026-06-13): frozen v1 schemas + root-relative scaffolding, validate pass, memory-routing docs + `docs/` directory created (agent-interface.md, formats.md, concepts.md, etc.)
- **Plan 41** (2026-06-16): headless CLI contract — new `internal/nonint/` package, `cmd/completion.go`, `*Result` headless cores, `list --json`, `ExitConflict=5` exit contract

These represent significant structural additions that are not all reflected in docs.

### Step 2 — Check INDEX.md accuracy

Verified INDEX.md tech stack, folder structure, and project description. Findings:

- **CLI commands count** says `8` — actual user-visible commands are now 9: init, add, remove, list, catalog, update, guide, validate, completion (completion.go added by Plan 39/external contribution, PR #78 merged 2026-05-07)
- **Catalog items** says `~50` — actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines)
- **Architecture diagram** in INDEX.md does not list `internal/nonint/` (added by Plan 41)
- **Document Registry** does not list `docs/` directory (added by Plan 40 Phase 3)
- Tech stack, agent types count (6), and description all correct

### Step 3 — Check navigation links

Verified all links in `station/CLAUDE.md` navigation tables:

- **Core** (identity, memory, self-awareness): all resolve ✓
- **Bonsai Reference** (bonsai-model, catalog.json, .bonsai.yaml): all resolve ✓
- **Protocols** (memory, scope-boundaries, security, session-start): all resolve ✓
- **Workflows** (10 entries): all resolve ✓ — BUT `plan-grilling.md` exists in `agent/Workflows/` and is **not listed** in the nav table
- **Skills** (5 entries + bubbletea): all resolve ✓ — BUT `critic-agent-prompts.md` exists in `agent/Skills/` and is **not listed** in the nav table
- **Routines** (7 entries): all resolve ✓
- **Sensors** (10 entries): all resolve ✓

Also verified `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/` links — all files referenced in tables exist.

**One structural note:** `station/CLAUDE.md` links to `agent/Skills/bubbletea.md` (a file), but the `bubbletea` directory also exists at `agent/Skills/bubbletea/`. Both exist — the file link resolves correctly. No breakage, low drift.

### Step 4 — Report findings

8 drift items identified (see Findings Summary below). All flagged for user decision per routine procedure — no docs edited.

### Step 5 — Update dashboard

Dashboard row for "Doc Freshness Check" updated to Last Ran: 2026-06-20, Next Due: 2026-06-27, Status: done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package missing from project structure tree | `Bonsai/CLAUDE.md` | Flagged for user |
| 2 | MEDIUM | CLI commands count stale: says 8, actual is 9 (completion added 2026-05-07) | `station/INDEX.md` Key Metrics | Flagged for user |
| 3 | MEDIUM | Catalog items count stale: says ~50, actual is 53 | `station/INDEX.md` Key Metrics | Flagged for user |
| 4 | MEDIUM | `plan-grilling.md` workflow not in Skills nav table | `station/CLAUDE.md` Workflows section | Flagged for user |
| 5 | MEDIUM | `critic-agent-prompts.md` skill not in Skills nav table | `station/CLAUDE.md` Skills section | Flagged for user |
| 6 | LOW | `docs/` directory not in Document Registry (added Plan 40 P3) | `station/INDEX.md` Document Registry | Flagged for user |
| 7 | LOW | Architecture diagram missing `internal/nonint/` | `station/INDEX.md` Architecture Overview | Flagged for user |
| 8 | INFO | `code-index.md` has no entries for `internal/nonint/` package (~10 new files from Plan 41) | `station/code-index.md` | Flagged for user |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**HIGH — `internal/nonint/` missing from root CLAUDE.md project structure tree**
Plan 41 added `internal/nonint/` with ~10 source files (config.go, events.go, nonint.go, remove.go, result.go, runner.go, update.go, etc.). This is a first-class internal package and should appear in the project structure tree in `Bonsai/CLAUDE.md`.

**MEDIUM — Two undocumented workspace abilities**
- `station/agent/Workflows/plan-grilling.md` (added 2026-06-13, commit `6995d4f`) — not in `station/CLAUDE.md` Workflows nav table
- `station/agent/Skills/critic-agent-prompts.md` (added 2026-06-13, same commit) — not in `station/CLAUDE.md` Skills nav table
Both have `source: adapted from ZenGarden ZEN/Docs ... full Bonsai-catalog integration pending (Backlog)`. If the Backlog item covers adding them to the nav table, that may already be tracked — recommend confirming or adding both rows to the nav.

**MEDIUM — INDEX.md Key Metrics stale**
- CLI commands: 8 → 9 (completion command shipped as PR #78 from external contributor, 2026-05-07)
- Catalog items: ~50 → 53

**LOW — docs/ directory undocumented in INDEX.md**
`docs/` was created by Plan 40 Phase 3 (PR #115, 2026-06-13) with agent-interface.md, formats.md, concepts.md, quickstart.md, cli.md, custom-files.md. It is a user-facing docs directory not listed in the Document Registry.

**LOW — INDEX.md arch diagram missing internal/nonint/**
Diagram lists internal/catalog, config, generate, validate, wsvalidate, tui — missing nonint.

**INFO — code-index.md missing internal/nonint/**
The `internal/nonint/` package (Plan 41) has no entries in `station/code-index.md`. This package contains the headless CLI contract core — runner.go, result.go, events.go, nonint.go, config.go, update.go, remove.go. Worth documenting for agent navigation.

## Notes for Next Run

- Root `Bonsai/CLAUDE.md` project-structure tree is a recurring drift source — it requires manual updates after every plan that adds packages or files. Consider adding a structured update step to any plan that touches `cmd/` or `internal/`.
- `docs/` directory is growing and not tracked in any nav — may benefit from a Document Registry row in INDEX.md.
- Previous cycle (2026-05-04) flagged: root CLAUDE.md tree drift, code-index.md staleness, INDEX.md CLI count. CLI count remains partially unresolved (8→9); root CLAUDE.md tree was partially updated (Plan 37 doc-refresh-bundle covered Go drift but not structural additions from later plans).
