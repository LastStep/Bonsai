---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-24
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 13 — `station/agent/Routines/doc-freshness-check.md`, `station/CLAUDE.md`, `station/INDEX.md`, `station/code-index.md`, `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/agent/Workflows/plan-grilling.md`, `station/agent/Skills/critic-agent-prompts.md`, `internal/generate/generate.go` (line numbers), `cmd/add.go`, `cmd/remove.go`, `cmd/root.go`, `cmd/list.go`, `internal/catalog/catalog.go`, `internal/config/config.go`, `internal/config/lockfile.go`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, file checks, grep), Glob, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation vs git history (last 7 days)
Git log shows only 2 commits in the past 7 days (both from today's routine runs — backlog-hygiene and status-hygiene). Expanded to 90-day window to capture Plan 41 (shipped 2026-06-16) which is the most recent feature work. Plan 41 added headless CLI cores (`--json`, `--non-interactive`, `--yes`, `--from`, `--skip-conflicts` flags; exit codes; `docs/agent-interface.md`). This work is not reflected in station documentation.

### Step 2: Check INDEX.md accuracy
- Tech stack table: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS)
- Agent types count (6): accurate
- Catalog items (~50): plausible, not verified exactly
- **CLI commands count: STALE** — listed as 8, but `bonsai completion` (cmd/completion.go) was shipped in Plan 39 / v0.4.2 era (merged PR #78, @mvanhorn external contribution). Count should be 9. The commands list "(init, add, remove, list, catalog, update, guide, validate)" is missing `completion`.
- **Plan 41 additions absent**: `--json` flag on `list`, non-interactive flags (`--yes`, `--from`, `--skip-conflicts`) on `add`/`update`/`remove`, exit codes contract (ExitConflict=5), `docs/agent-interface.md` — none of this appears anywhere in INDEX.md.
- Architecture diagram in INDEX.md: accurate in structure, just missing the headless/JSONL output path.

### Step 3: Check navigation links
Verified 53 links from `station/CLAUDE.md` navigation tables. **All 53 resolve to real files.** No broken links.

However, two files exist in the agent workspace that are NOT listed in the navigation tables:

1. `station/agent/Workflows/plan-grilling.md` — exists, functional, not in CLAUDE.md Workflows table
2. `station/agent/Skills/critic-agent-prompts.md` — exists, functional, not in CLAUDE.md Skills table
3. `station/agent/Skills/bubbletea/` — a subdirectory with 4 reference files (components.md, emoji-width-fix.md, golden-rules.md, troubleshooting.md) exists alongside `bubbletea.md`. Not mentioned in CLAUDE.md or code-index.md.

Also checked `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/`, `agent/Sensors/`, `agent/Routines/` directories — all other files are properly represented.

### Step 4: Report findings (below)

### Step 5: Update dashboard
Updated `agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-07-24, Next Due → 2026-07-31, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | CLI command count stale: INDEX.md says 8, actual is 9 (`completion` missing) | `station/INDEX.md` line 33 | Flagged for user — propose updating count to 9 and adding `completion` to list |
| 2 | Medium | Plan 41 headless CLI features absent from INDEX.md (--json, nonint flags, exit codes, agent-interface.md) | `station/INDEX.md` Architecture section | Flagged for user — propose adding a "Headless / Agent Interface" row to Key Metrics or Architecture section |
| 3 | Low | `plan-grilling.md` workflow not listed in station/CLAUDE.md Workflows nav table | `station/CLAUDE.md` Workflows table | Flagged for user — custom file; add a row or leave as intentionally unlisted |
| 4 | Low | `critic-agent-prompts.md` skill not listed in station/CLAUDE.md Skills nav table | `station/CLAUDE.md` Skills table | Flagged for user — companion to plan-grilling; add a row or leave as intentionally unlisted |
| 5 | Medium | code-index.md has significant line-number drift across generate.go, add.go, remove.go (Plan 41 shift) | `station/code-index.md` | Flagged for user — line numbers off by 30–138 lines; detail below |
| 6 | Low | `bubbletea/` subdirectory (4 reference files) undocumented in CLAUDE.md and code-index.md | `station/agent/Skills/bubbletea/` | Flagged for user — decide whether to add sub-entries to CLAUDE.md Skills table |

---

## Finding 5 Detail: code-index.md Line Drift

Plan 41 (June 2026) added headless cores (`internal/nonint/`) and forAgent variants that shifted all line numbers in `generate.go`, `cmd/add.go`, and `cmd/remove.go`. The previous doc-freshness run (2026-05-04) already flagged code-index drift; it has grown substantially.

### `generate.go` (internal/generate/generate.go)

| Function | Documented | Actual | Delta |
|----------|-----------|--------|-------|
| `FileAction` type | `:141` | `:171` | +30 |
| `FileResult` type | `:153` | `:183` | +30 |
| `WriteResult` type | `:162` | `:192` | +30 |
| `writeFile()` | `:277` | `:307` | +30 |
| `writeFileChmod()` | `:317` | `:347` | +30 |
| `Scaffolding()` | `:360` | `:401` | +41 |
| `SettingsJSON()` | `:473` | `:564` | +91 |
| `WorkspaceClaudeMD()` | `:725` | `:826` | +101 |
| `EnsureRoutineCheckSensor()` | `:972` | `:1073` | +101 |
| `RoutineDashboard()` | `:1010` | `:1111` | +101 |
| `PathScopedRules()` | `:1164` | `:1265` | +101 |
| `WorkflowSkills()` | `:1228` | `:1329` | +101 |
| `AgentWorkspace()` | `:1359` | `:1460` | +101 |

New undocumented functions (Plan 41): `SettingsJSONForAgent` (`:578`), `PathScopedRulesForAgent` (`:1278`), `WorkflowSkillsForAgent` (`:1338`).

### `cmd/add.go`

| Function | Documented | Actual | Delta |
|----------|-----------|--------|-------|
| `runAdd()` | `:56` | `:73` | +17 |
| `applyCinematicConflictPicks()` | `:309` | `:344` | +35 |
| `installedSet()` | `:365` | `:400` | +35 |
| `buildAddGrowAction()` | `:387` | `:422` | +35 |
| `distributeAddItemPicks()` | `:570` | `:605` | +35 |
| `availableAddItems()` | `:655` | `:690` | +35 |

New undocumented function (Plan 41): `runAddNonInteractive` (`:754`).

### `cmd/remove.go`

| Function | Documented | Actual | Delta |
|----------|-----------|--------|-------|
| `runRemoveItem()` | `:290` | `:428` | +138 |
| `runRemoveItemAction()` | `:565` | `:703` | +138 |
| `agentItemList()` | `:618` | `:756` | +138 |
| `itemIsRequired()` | `:667` | `:805` | +138 |
| `itemDisplayName()` | `:693` | `:831` | +138 |

New undocumented function (Plan 41): `runRemoveItemNonInteractive` (`:317`).

### `internal/catalog/catalog.go`

| Function | Documented | Actual | Delta |
|----------|-----------|--------|-------|
| `New(fsys)` | `:242` | `:286` | +44 |

### Accurate (no drift): `cmd/root.go` (±1), `internal/config/config.go`, `internal/config/lockfile.go`

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **INDEX.md CLI count** — Update line 33: "CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate)" → "CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion)". Simple one-line fix.

2. **INDEX.md Plan 41 coverage** — Plan 41 shipped a full headless CLI contract (June 2026). Decision needed: add a "Headless / Agent Interface" entry to Architecture section, or add a note to the CLI commands row mentioning `--json` and `--non-interactive` modes. `docs/agent-interface.md` exists and is the formal contract.

3. **plan-grilling + critic-agent-prompts in CLAUDE.md nav** — Both files are marked "Bonsai-catalog integration pending (Backlog)" so they're intentionally custom. Decision needed: add rows to CLAUDE.md Workflows/Skills tables so they're discoverable, or leave unlisted since they're transitional.

4. **code-index.md refresh** — Significant drift after Plan 41. This file is a developer navigation aid; line numbers this far off degrade its usefulness. Recommend dispatching a code agent to refresh all line numbers. Affects: generate.go (13 functions), add.go (6 functions), remove.go (5 functions), catalog.go (1 function). Also three new forAgent/nonint functions to add.

5. **bubbletea/ sub-directory** — Decide whether to reference `agent/Skills/bubbletea/` sub-files in CLAUDE.md (`components.md`, `emoji-width-fix.md`, `golden-rules.md`, `troubleshooting.md`). Currently only `bubbletea.md` is listed.

---

## Notes for Next Run

- code-index.md staleness was first flagged 2026-05-04 and is now >80 days deferred with significant additional drift from Plan 41. If not fixed before next run, consider promoting to a Backlog P2 item with a concrete plan number.
- The `plan-grilling` workflow and `critic-agent-prompts` skill were added during Plan 40/41 work and are actively used. Adding them to CLAUDE.md nav would make them easier to discover.
- All navigation links checked clean — no maintenance needed there.
- Roadmap Phase 2 / Phase 3 still do not reflect Plan 41's headless CLI work. Consider whether "MCP-ready cores (Plan 41)" should appear as a checked Phase 2 or Phase 3 item.
