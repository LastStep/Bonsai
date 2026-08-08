---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-08
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata

| Field | Value |
|-------|-------|
| Subagent run date | 2026-08-08 |
| Commits reviewed | Since 2026-05-04 (16 commits with code changes) |
| Files inspected | INDEX.md, code-index.md, station/CLAUDE.md, station/Playbook/Roadmap.md, agent/Core/routines.md |
| Nav links verified | Core (3), Protocols (4), Workflows (10 in dir), Skills (8 in dir), Sensors (10 in dir) |
| Status | partial — findings flagged for user decision; no edits made (procedure: flag only) |

## Procedure Walkthrough

### Step 1 — Scan project documentation vs. recent git history

Reviewed all commits since 2026-05-04. Major development since last check:
- **v0.4.2 (2026-05-13):** `bonsai init/add --non-interactive --from-config` → new `internal/nonint/` package
- **v0.4.3 (2026-05-13):** Sensor hook absolute-path hotfix
- **bonsai completion (2026-05-07):** `cmd/completion.go` added (PR #78, external contribution)
- **Plan 40 (2026-06-13):** Frozen v1 schemas, project-level validate pass (`internal/validate/project.go`), memory-routing protocol, `docs/formats.md`
- **Plan 41 (2026-06-16):** Full headless CLI contract — all mutating commands get `internal/nonint/` cores; `list --json`; `docs/agent-interface.md` contract doc; `internal/generate/list_snapshot.go`
- **Plan grilling pipeline (2026-06-13):** `station/agent/Workflows/plan-grilling.md` + `station/agent/Skills/critic-agent-prompts.md` added

Result: **9 documentation drift findings identified.**

### Step 2 — Check INDEX.md accuracy

Tech stack table: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template — all still correct).

Folder structure in Architecture Overview: **stale** — missing `internal/nonint/` package row. The `docs/` top-level directory is also unmentioned.

Key Metrics: **stale** — CLI commands count says 8, is now 9 (`bonsai completion` added May 2026).

### Step 3 — Check navigation links

All links in `station/CLAUDE.md` nav tables that exist in the directory structure resolve correctly. However, two files that exist in their directories are **absent from the nav tables**:

- `agent/Workflows/plan-grilling.md` — in directory, not in Workflows table
- `agent/Skills/critic-agent-prompts.md` — in directory, not in Skills table

All other links confirmed resolvable: Core (4 files), Protocols (4), Workflows (9 in table, 10 in dir), Skills (6 in table, 8 in dir), Sensors (10 — all present), Routines (7 — all present).

External links (`.bonsai/catalog.json`, `.bonsai.yaml`) — both files exist.

### Step 4 — Report findings

See Findings Summary below. All are flagged for user decision per procedure. No edits made.

### Step 5 — Update dashboard

Dashboard updated: Doc Freshness Check row → Last Ran 2026-08-08, Next Due 2026-08-15, Status done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package entirely absent from Architecture Overview | `station/INDEX.md` | Flagged — propose add |
| 2 | HIGH | `plan-grilling.md` not in Workflows nav table — agent cannot discover it via CLAUDE.md | `station/CLAUDE.md` | Flagged — propose add |
| 3 | MEDIUM | `critic-agent-prompts.md` not in Skills nav table | `station/CLAUDE.md` | Flagged — propose add |
| 4 | MEDIUM | CLI commands count says 8; `bonsai completion` added 2026-05-07 makes it 9 | `station/INDEX.md` | Flagged — propose update |
| 5 | MEDIUM | `docs/` top-level directory not mentioned anywhere in INDEX.md | `station/INDEX.md` | Flagged — propose add |
| 6 | LOW | `bonsai completion` missing from CLI commands table | `station/code-index.md` | Flagged — propose add |
| 7 | LOW | `internal/nonint/` package has no section | `station/code-index.md` | Flagged — propose add |
| 8 | LOW | `internal/validate/project.go` (Plan 40 Phase 2) not reflected in validate section | `station/code-index.md` | Flagged — propose add |
| 9 | LOW | `internal/generate/list_snapshot.go` (Plan 41 Phase 4) not in generator section | `station/code-index.md` | Flagged — propose add |

---

## Errors & Warnings

None. All files read successfully. No broken nav links found.

---

## Items Flagged for User Review

### HIGH — Fix soon (navigation drift blocks agent discoverability)

**Finding 1 — `internal/nonint/` missing from INDEX.md Architecture Overview**

The `nonint` package is now a substantial part of the codebase (Plans 39/41). Proposed addition to Architecture Overview table, after `internal/wsvalidate/`:

```
internal/nonint/   ← headless CLI contract — RunInit/RunAdd/RunUpdate/RunRemove cores, Result shape, JSONL event emission, exit codes
```

And update the `cmd/` line to note non-interactive flag support:
```
cmd/ (Cobra)  ← CLI commands: init, add, remove, list, catalog, update, guide, validate, completion
```

**Finding 2 — `plan-grilling.md` not in Workflows nav table**

File added 2026-06-13 (`6995d4f`). Proposed row to add to the Workflows table in `station/CLAUDE.md`:

```
| Starting adversarial plan review — running 6-critic grilling pipeline on a draft plan | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

### MEDIUM — Fix when convenient

**Finding 3 — `critic-agent-prompts.md` not in Skills nav table**

File added 2026-06-13 (`6995d4f`). Proposed row to add to the Skills table in `station/CLAUDE.md`:

```
| Constructing adversarial critic prompts for plan or code review grilling | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

**Finding 4 — Key Metrics CLI command count**

`station/INDEX.md` Key Metrics row: `CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate)` → should be `9 (init, add, remove, list, catalog, update, guide, validate, completion)`.

**Finding 5 — `docs/` directory not in INDEX.md**

The `docs/` directory contains public-facing documentation: `quickstart.md`, `concepts.md`, `cli.md`, `custom-files.md`, `formats.md`, `agent-interface.md`. Propose adding a row to the Document Registry or Architecture section:

```
| `docs/` | Public-facing docs — quickstart, CLI reference, concepts, custom-files guide, formats, agent-interface contract | When writing public docs or reading the agent-interface contract |
```

### LOW — Code-index housekeeping

**Findings 6–9** are code-index additions. These don't block agent navigation (CLAUDE.md is the primary nav) but make the code-index incomplete for developer lookups:

- `bonsai completion` row in CLI commands table (cmd/completion.go)
- New section: `## Nonint (`internal/nonint/`) — Plan 39/41` covering RunInit, RunAdd, RunUpdate, RunRemove, Result, ExitCode constants
- `project.go` entry in the Validate section (project-level manifest + memory-dir linting, Plan 40 Phase 2)
- `list_snapshot.go` entry in the Generator section (ListSnapshot type for `bonsai list --json`, Plan 41 Phase 4)

---

## Notes for Next Run

- Roadmap.md accuracy is out of scope for this routine (covered by Roadmap Accuracy routine) — not assessed.
- `docs/` was created under Plan 40/41; if the Index.md is updated, verify the `agent-interface.md` contract is cross-referenced.
- The `plan-grilling.md` workflow and `critic-agent-prompts.md` skill both appear to be working tools (used in Plan 41 grilling), so their absence from the nav table is a genuine discoverability gap.
- Next run (2026-08-15): verify above items were addressed. Check for any new packages or commands introduced since today.
