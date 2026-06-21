---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-21
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
- **Duration:** ~5 min
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`, `station/CLAUDE.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `internal/nonint/runner.go`, `docs/agent-interface.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** `git log --oneline --since=2026-05-04`, `ls` (multiple dirs), `grep` (CLAUDE.md links, code-index contents), link resolution check on all CLAUDE.md relative links
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`. Ran `git log --oneline --since=2026-05-04` to enumerate all commits since last run.
- **Result:** 37 commits since 2026-05-04. Major code deliveries: Plan 39 (non-interactive flags, v0.4.2), Plan 40 Phases 1–3 (frozen schemas + root-relative scaffolding + validate pass, v0.5.0 untagged), Plan 41 (headless CLI contract + MCP-ready cores, 5 phases, PRs #120–#125), v0.4.3 hotfix (absolute sensor paths). New packages: `internal/nonint/` (Plans 39/41). New files in `internal/generate/`: `list_snapshot.go`, `catalog_snapshot_unix.go`, `catalog_snapshot_windows.go`. New command: `bonsai completion` (PR #78, shipped v0.4.1 era). New station workflows/skills: `plan-grilling.md`, `critic-agent-prompts.md`.
- **Issues:** Several documentation files have not been updated to reflect these changes (see findings below).

### Step 2: Check INDEX.md accuracy
- **Action:** Compared INDEX.md tech stack, folder structure, CLI command list, and key metrics against the current codebase state.
- **Result:** Tech stack is accurate. Folder structure matches. CLI commands listed as "8 (init, add, remove, list, catalog, update, guide, validate)" — but `bonsai completion` is now a 9th top-level command registered via `cmd/completion.go`. The catalog items metric says "~50" — actual count is 53 (skills: 18, workflows: 10, protocols: 4, sensors: 13, routines: 8), which is within the "~50" approximation. The architecture block does not mention `internal/nonint/` (Plans 39/41 headless package), though its absence is not critical since the block is an overview.
- **Issues:** CLI commands count ("8") and the parenthetical list both omit `completion`. Minor staleness on catalog count (~50 vs 53).

### Step 3: Check navigation links
- **Action:** Extracted all relative links from `station/CLAUDE.md` and resolved each against `station/`. Also checked file listings in `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/`, `agent/Sensors/`, `agent/Routines/`.
- **Result:** All 40+ links in `station/CLAUDE.md` resolve to real files — no broken links detected. However, two files exist in the agent workspace that are NOT listed in the CLAUDE.md navigation tables:
  - `agent/Workflows/plan-grilling.md` — exists on disk (added via Plan 40 station session 2026-06-13), but has no row in the Workflows table in `station/CLAUDE.md`.
  - `agent/Skills/critic-agent-prompts.md` — exists on disk (added alongside plan-grilling), but has no row in the Skills table in `station/CLAUDE.md`.
- **Issues:** Two navigational gaps — files exist but are invisible to the agent via the nav table.

### Step 4: Report findings
- **Action:** Compiled all drift instances into the findings summary below.
- **Result:** 4 findings total, 3 medium severity, 1 low severity. All flagged for user decision per procedure — no doc edits made.
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for "Doc Freshness Check" — set Last Ran to 2026-06-21, Next Due to 2026-06-28, Status to done.
- **Result:** Dashboard updated.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `bonsai completion` command missing from CLI commands list and count ("8" should be "9"). Command shipped in v0.4.1 (PR #78). | `station/INDEX.md` lines 33 and 63 | Flagged for user decision |
| 2 | Medium | `agent/Workflows/plan-grilling.md` exists but has no row in the Workflows navigation table. Added during Plan 40 station session 2026-06-13. | `station/CLAUDE.md` — Workflows section | Flagged for user decision |
| 3 | Medium | `agent/Skills/critic-agent-prompts.md` exists but has no row in the Skills navigation table. Added alongside plan-grilling. | `station/CLAUDE.md` — Skills section | Flagged for user decision |
| 4 | Low | `internal/nonint/` package (Plan 41 headless cores — RunInit, RunAdd, RunUpdate, RunRemove, exit codes 0/2/3/4/5) has no entry in code-index.md. New generate files `list_snapshot.go`, `catalog_snapshot_unix.go`, `catalog_snapshot_windows.go` also absent. | `station/code-index.md` | Flagged for user decision |

---

## Proposed Updates (for user decision — not applied)

### Finding 1: INDEX.md CLI command count
Change line 33:
```
| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |
```
to:
```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```
And update the architecture block on line 63 similarly.

### Finding 2: station/CLAUDE.md — add plan-grilling workflow row
Add to the Workflows table (e.g. between planning and pr-review):
```
| Starting a plan grilling / adversarial review of a draft plan | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```
(Exact trigger text and placement should be confirmed with the file contents.)

### Finding 3: station/CLAUDE.md — add critic-agent-prompts skill row
Add to the Skills table:
```
| Prompts and personas for critic agents in the plan-grilling pipeline | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```
(Exact trigger text should be confirmed with the file contents.)

### Finding 4: code-index.md — add internal/nonint/ section
Add a new section documenting `internal/nonint/` package with entry points (RunInit, RunAdd, RunUpdate, RunRemove, exit code constants) and note the new generate split-files. This was a significant architectural addition via Plan 41 that other agents should be aware of for debugging and extension.

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

4 items flagged — all require user decision before docs are updated:

1. **[INDEX.md] CLI command count drift** — `completion` omitted from count and list. Low-effort fix.
2. **[CLAUDE.md] plan-grilling.md not in Workflows nav** — agent cannot find/load this workflow via the nav table. Recommend adding a row with appropriate trigger text (check the file for its purpose).
3. **[CLAUDE.md] critic-agent-prompts.md not in Skills nav** — skill is invisible to the agent. Recommend adding a row.
4. **[code-index.md] internal/nonint/ package undocumented** — Plan 41 headless package is significant enough to warrant a code-index section. Lower priority than the nav gaps but worth a doc sweep.

---

## Notes for Next Run

- The station/CLAUDE.md Workflows and Skills tables should be audited against `ls agent/Workflows/` and `ls agent/Skills/` at each run — new files added ad-hoc may not make it into the nav tables.
- The gap between runs (48 days) means multiple features shipped without doc updates. Weekly cadence should be restored to keep drift manageable.
- Roadmap.md is accurate — Phase 1 checklist remains valid. Phase 2/3/4 items unchanged.
- All navigation links in CLAUDE.md resolved cleanly — no broken refs.
