---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-25
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (findings flagged for user decision — procedure requires no autonomous doc edits)
- **Duration:** ~10 min
- **Files Read:** 9 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/Reports/Pending/` (listing)
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, file existence checks, grep), Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read station/INDEX.md, Playbook/Status.md, Playbook/Roadmap.md. Ran `git log --oneline -30` and `git log --oneline --since="7 days ago"` to identify recent commits.
- **Result:** Only one commit in the last 7 days (`2db828f` — backlog-hygiene routine). The gap since last doc-freshness-check (2026-05-04) spans ~82 days and covers Plans 40 and 41, plus the `completion` command. Key changes: Plan 40 (frozen schemas, root-relative scaffolding), Plan 41 (headless CLI contract — `--non-interactive`, `--from-config`, `--skip-conflicts`, `--yes`, `--from` flags; `list --json`; `ExitConflict=5`; new `internal/nonint/` package; `docs/agent-interface.md`). `completion` command added in commit `2eae9d4` (v0.4.x).
- **Issues:** Four doc drift items found (detailed in Findings Summary below).

### Step 2: Check INDEX.md accuracy
- **Action:** Compared INDEX.md "Key Metrics" table and "Architecture Overview" against actual codebase state.
- **Result:**
  - Agent count (6): confirmed correct — backend, devops, frontend, fullstack, security, tech-lead.
  - Catalog items (~50): actual count is 53 meta.yaml files; ~50 is still within reasonable range.
  - CLI commands count: INDEX.md says "8" — actual is **9** (`completion` was added in v0.4.x and is not listed).
  - Tech stack table: accurate. Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML v3, text/template, embed.FS — all confirmed.
  - Architecture overview diagram: does not mention the new `internal/nonint/` headless CLI layer added in Plan 41.
- **Issues:** CLI commands count stale (8 vs 9); `completion` missing from command list; `internal/nonint/` not in architecture overview.

### Step 3: Check navigation links
- **Action:** Verified all 51 file/directory targets linked from station/CLAUDE.md against the filesystem. This covered Core files, Protocols, Workflows, Skills, Routines, Sensors, Playbook, Logs, Reports, and external pointers.
- **Result:** All 51 targets resolve. Zero broken links.
- **Issues:** None.

### Step 4: Report findings
- **Action:** Compiled findings table below. Per procedure, all updates are flagged for user decision — not executed.
- **Result:** 4 findings identified; 3 are low severity, 1 is medium.
- **Issues:** None (procedure followed correctly).

### Step 5: Update dashboard
- **Action:** Updated `routines.md` dashboard row for "Doc Freshness Check" — Last Ran → 2026-07-25, Next Due → 2026-08-01, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | `showWriteResults()` listed in Shared Helpers table but function no longer exists — removed in Plan 41 Phase 2 (commit `5a4cf3c`) | `station/code-index.md` — CLI Commands / Shared Helpers | Flagged for user — remove stale row |
| 2 | low | `internal/nonint/` package (Plan 41 headless CLI infrastructure) not documented in code-index at all | `station/code-index.md` | Flagged for user — add new section |
| 3 | low | CLI commands count says "8" but is "9" — `completion` added in commit `2eae9d4` | `station/INDEX.md` — Key Metrics table | Flagged for user — update count + list |
| 4 | low | `completion` command missing from CLI Commands table | `station/code-index.md` — CLI Commands section | Flagged for user — add row |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**Finding 1 — Remove stale `showWriteResults()` row from code-index.md (medium)**
The Shared Helpers table in code-index.md lists `showWriteResults()` at line `:201`. This function was removed from `cmd/root.go` during Plan 41 Phase 2. A grep across all Go source confirms it no longer exists anywhere. The row should be deleted from the table.

**Finding 2 — Add `internal/nonint/` section to code-index.md (low)**
Plan 41 added an entire new package at `internal/nonint/` with the following files:
- `nonint.go` — shared non-interactive core utilities
- `runner.go` — exit-code constants (`ExitConflict=5`, `ExitPartial=2`, `ExitNotFound=3`, `ExitBadInput=4`) and per-command orchestration
- `events.go` — JSONL event emitters (file events, summary events)
- `result.go` — `Result` struct (counts: created, updated, skipped, conflicts)
- `update.go` / `remove.go` — headless cores for `update` and `remove` commands
This package is the canonical source for the agent-interface contract (`docs/agent-interface.md`) and is entirely absent from code-index.md. Proposed action: add an "## Headless CLI (`internal/nonint/`)" section documenting runner.go exit constants, events.go event shapes, and result.go Result type.

**Finding 3 — Update INDEX.md CLI command count (low)**
`station/INDEX.md` Key Metrics table says "8 (init, add, remove, list, catalog, update, guide, validate)". The `completion` command (`bonsai completion [bash|zsh|fish|powershell]`) was added in v0.4.x (commit `2eae9d4`, merged PR #78). Count should be updated to **9** and `completion` added to the parenthetical list.

**Finding 4 — Add `completion` to code-index.md CLI Commands table (low)**
The CLI Commands table in code-index.md lists 8 bonsai subcommands but omits `completion`. Proposed row:
```
| `bonsai completion` | `cmd/completion.go` | Shell completion generator (bash/zsh/fish/powershell) — auto-wired by Cobra |
```

## Notes for Next Run

- The 82-day gap since the last doc-freshness-check (2026-05-04 → 2026-07-25) allowed three doc-drift items to accumulate. If routines run on schedule (7-day cadence), drift is easier to catch one feature at a time.
- Finding 1 (`showWriteResults()` stale row) is the most actionable — it's incorrect data in a navigation document the agent actively uses. Recommend prioritizing.
- All navigation links in CLAUDE.md are clean — no broken references.
- Next run should verify: (a) headless flags appear in INDEX.md or code-index.md if user chooses to add them, (b) any new Plan 42 (MCP server) additions land in docs.
