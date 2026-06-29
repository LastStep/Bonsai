---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-29
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~12 min
- **Files Read:** 11 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/CLAUDE.md`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`, `/home/user/Bonsai/station/agent/Skills/critic-agent-prompts.md`, `/home/user/Bonsai/.bonsai.yaml`
- **Files Modified:** 3 — `/home/user/Bonsai/station/CLAUDE.md` (2 rows added), `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard updated), `/home/user/Bonsai/station/Logs/RoutineLog.md` (entry appended)
- **Tools Used:** Read, Bash (git log, ls, find, python3/yaml, grep), Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history
Git log for the last 7 days returned 2 commits (both 2026-06-29): `588bbec` (status-hygiene routine) and `ea7c5f7` (backlog-hygiene routine). Both are routine maintenance commits — no new features, services, or config changes from code commits in this window.

However, comparing actual filesystem state against documented state revealed significant drift accumulated since the last doc-freshness run (2026-05-04 — 56 days). This drift is from Plans 39–41 (shipped between 2026-05-04 and 2026-06-16).

### Step 2 — Check INDEX.md accuracy
- **CLI command count:** INDEX.md says 8 commands. Actual count is now **9**: `completion` subcommand was added (Plan 41 era / v0.4.2 contribution). Low severity — INDEX.md lists the names `(init, add, remove, list, catalog, update, guide, validate)` which is still correct as the primary surface; `completion` is a utility subcommand. Flagging for user decision.
- **Tech stack:** Accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea all current.
- **Agent types:** 6 — confirmed accurate.
- **Catalog items ~50:** Actual count is 53 (skills 18 + workflows 10 + protocols 4 + sensors 13 + routines 8). Close enough — no update needed.
- **Architecture diagram in INDEX.md:** References `internal/wsvalidate/` (accurate), `internal/validate/` (accurate). No drift on INDEX.md architecture section.

### Step 3 — Check navigation links

**station/CLAUDE.md:** All 57 links checked (55 internal + 2 parent). **Zero broken links** before this run. However, two custom ability files existed in `agent/Skills/` and `agent/Workflows/` without corresponding entries in the navigation tables:
- `agent/Skills/critic-agent-prompts.md` — not listed in Skills table
- `agent/Workflows/plan-grilling.md` — not listed in Workflows table

Both were added to `station/CLAUDE.md` as part of this routine run (autonomous fix, low risk — adding rows to nav tables).

**agent/Core/ files:** `memory.md` contains 6 broken links to `Research/RESEARCH-*.md` files (paths like `../../Research/RESEARCH-landscape-analysis.md`). The `station/Research/` directory does not exist. These files may have moved or been removed. Flagging for user — cannot resolve autonomously.

Also found 1 false-positive broken link: `[label](url)` in `memory.md` — this is template placeholder text, not a real link.

**agent/Workflows/issue-to-implementation.md:** Contains 3 broken links to `agent/Skills/dispatch.md`. The `dispatch` skill exists in the catalog (`catalog/skills/dispatch/`) but is **not installed** for the tech-lead agent in `.bonsai.yaml`. The file `station/agent/Skills/dispatch.md` therefore does not exist. Flagging for user — needs either skill installation via `bonsai add` or workflow text update.

**agent/Workflows/session-wrapup.md:** Contains 2 false-positive broken links (`[plan](path)` and `[PR #N](url)`) — these are template placeholder examples in the workflow body, not real navigation links. Low severity, no action needed.

**agent/Core/, agent/Protocols/ files:** No broken links found.

### Step 4 — Check code-index.md accuracy

Line numbers in `code-index.md` are **significantly drifted** for `internal/catalog/catalog.go`:

| Function | code-index.md | Actual |
|----------|--------------|--------|
| `DisplayNameFrom()` | `:49` | `:50` |
| `New()` | `:242` | `:286` |
| `loadItems()` | `:346` | `:390` |
| `loadSensors()` | `:397` | `:441` |
| `loadRoutines()` | `:448` | `:492` |
| `loadScaffolding()` | `:499` | `:543` |
| `loadAgents()` | `:516` | `:560` |

For `internal/generate/generate.go`, key functions also drifted:

| Function | code-index.md | Actual |
|----------|--------------|--------|
| `Scaffolding()` | `:360` | `:401` |
| `SettingsJSON()` | `:473` | `:564` |
| `WorkspaceClaudeMD()` | `:725` | `:826` |
| `AgentWorkspace()` | `:1359` | `:1460` |
| `EnsureRoutineCheckSensor()` | `:972` | `:1073` |
| `RoutineDashboard()` | `:1010` | `:1111` |

Additionally, `code-index.md` does not document these new files shipped in Plans 39–41:
- `internal/generate/list_snapshot.go` — `ListSnapshot` type + headless list-JSON serializer (Plan 41, Phase 4)
- `internal/generate/catalog_snapshot_unix.go` / `catalog_snapshot_windows.go` — OS-split platform files (v0.4.0 hotfix PR #95)
- `cmd/completion.go` — `completion [bash|zsh|fish|powershell]` subcommand (external contribution, v0.4.2 era)
- `cmd/add_nonint_test.go`, `cmd/init_nonint_test.go`, `cmd/remove_nonint_test.go`, `cmd/update_nonint_test.go` — non-interactive mode tests (Plan 41)
- `cmd/add_test.go`, `cmd/catalog_test.go`, `cmd/guide_test.go`, `cmd/list_test.go`, `cmd/validate_test.go` — cmd-level integration tests

Root **CLAUDE.md project structure tree** also missing these files.

### Step 5 — Report findings and update dashboard
Dashboard updated; log appended; report written.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | `critic-agent-prompts.md` skill missing from Skills nav table | `station/CLAUDE.md` | Fixed — row added to Skills table |
| 2 | medium | `plan-grilling.md` workflow missing from Workflows nav table | `station/CLAUDE.md` | Fixed — row added to Workflows table |
| 3 | high | 6 broken links to `Research/RESEARCH-*.md` files (directory does not exist) | `station/agent/Core/memory.md` | Flagged for user — cannot resolve autonomously |
| 4 | high | 3 broken links to `agent/Skills/dispatch.md` (skill not installed) | `station/agent/Workflows/issue-to-implementation.md` | Flagged for user — needs `bonsai add` or workflow text update |
| 5 | medium | `code-index.md` line numbers significantly drifted for `catalog.go` and `generate.go` (off by +40–100 lines) | `station/code-index.md` | Flagged for user — line-number refresh needed |
| 6 | medium | `code-index.md` missing: `list_snapshot.go`, `catalog_snapshot_unix/windows.go`, `cmd/completion.go`, 9 new test files | `station/code-index.md` | Flagged for user — entries needed |
| 7 | medium | Root `CLAUDE.md` project tree missing same new files (completion.go, test files, list_snapshot.go, platform split files) | `/home/user/Bonsai/CLAUDE.md` | Flagged for user — tree update needed |
| 8 | low | INDEX.md CLI command count 8 → 9 (`completion` subcommand shipped) | `station/INDEX.md` | Flagged for user |
| 9 | info | `[plan](path)` / `[PR #N](url)` placeholder links in session-wrapup.md — not real navigation links | `station/agent/Workflows/session-wrapup.md` | No action — template placeholders |
| 10 | info | `[label](url)` placeholder in memory.md — not a real link | `station/agent/Core/memory.md` | No action — template placeholder |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

**Finding 3 — Broken Research/ links in memory.md (high):**
`station/agent/Core/memory.md` references 6 research documents via paths like `../../Research/RESEARCH-landscape-analysis.md`. The `station/Research/` directory does not exist. These files may have been deleted, archived, or never created. Options: (a) create/restore the Research directory and files, (b) update the links in memory.md to the correct location, (c) remove the stale link entries if the files are gone.

**Finding 4 — Broken dispatch.md links in issue-to-implementation.md (high):**
`station/agent/Workflows/issue-to-implementation.md` references `agent/Skills/dispatch.md` in 3 places. The `dispatch` skill exists in the catalog but is not installed for the tech-lead agent. Options: (a) run `bonsai add` and select the `dispatch` skill for tech-lead, or (b) update the workflow references to the correct alternative (e.g., inline the dispatch guidance or link to agent-review/dispatch-guard sensors).

**Finding 5+6 — code-index.md staleness (medium):**
Line numbers for key functions in `catalog.go` and `generate.go` have drifted significantly (+40–100 lines). New files from Plans 39–41 are undocumented. A full code-index refresh is warranted — this is a non-trivial update best done as a dedicated task.

**Finding 7 — Root CLAUDE.md project tree drift (medium):**
The project-structure tree in `/home/user/Bonsai/CLAUDE.md` is missing several files added since May 2026. This is a recurring finding (previously flagged in 2026-05-04 run). Consider a bundled doc-refresh task.

**Finding 8 — INDEX.md CLI count drift (low):**
`completion` subcommand was contributed externally and is now shipped. The count of 8 in INDEX.md should become 9.

## Notes for Next Run
- Research/ directory link decay is new this cycle — first appearance. Should be resolved before next run to avoid false positives.
- dispatch.md link decay predates this cycle (not previously flagged — skill may have been added and removed).
- code-index.md line-number drift is a recurring pattern; consider whether the index provides enough value to justify periodic refresh effort, or whether a Backlog P3 item for automation is warranted.
- Two autonomous fixes applied this run: plan-grilling workflow row + critic-agent-prompts skill row added to station/CLAUDE.md.
