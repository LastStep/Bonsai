---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-11
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
- **Duration:** ~10 minutes
- **Files Read:** 9 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/CLAUDE.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/internal/generate/generate.go` (line number probes), `/home/user/Bonsai/cmd/*.go` (function location probes)
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read `git log --since="14 days ago"` and `--since="60 days ago"` with `--name-only`. Identified commits since last doc-freshness-check (2026-05-04). Relevant code-changing commits: Plans 40 (v0.5.0 Phases 1-3) and Plan 41 (headless CLI contract + non-interactive flags).
- **Result:** Two significant plan cycles shipped since last check — Plan 40 (freeze schemas, validate pass, docs/guide) and Plan 41 (headless CLI contract — non-interactive init/add/update/remove, new `internal/nonint/` package, `--yes`/`--from`/`--non-interactive`/`--skip-conflicts` flags, platform-split catalog_snapshot, list_snapshot). These changes added substantial code to existing files and introduced entirely new packages and files not yet reflected in docs.
- **Issues:** None in execution; multiple doc drift items found (detailed below).

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` and compared tech stack, agent count, catalog count, and CLI command count against actual directory listings.
- **Result:**
  - Tech stack: ACCURATE — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea all correct.
  - Agent types: 6 (tech-lead, fullstack, backend, frontend, devops, security) — ACCURATE.
  - CLI commands: 8 (init, add, remove, list, catalog, update, guide, validate) — ACCURATE.
  - Catalog items: "~50" — actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines) — STILL ACCURATE (~50 is close enough).
  - Architecture diagram: ACCURATE — shows all layers correctly.
- **Issues:** No actionable drift in INDEX.md.

### Step 3: Check navigation links in station/CLAUDE.md
- **Action:** Verified every file referenced in the Core, Protocols, Workflows, Skills, Sensors, and Routines nav tables against the actual directory contents of `station/agent/`.
- **Result:**
  - All Core, Protocol, Skills, Sensors, Routines links resolve — all referenced files exist.
  - Previously-broken `agent/Skills/bonsai-model.md` link now resolves (this was flagged in 2026-05-04 run — it has since been created/fixed).
  - **DRIFT FOUND:** `agent/Workflows/plan-grilling.md` exists on disk but is NOT listed in the Workflows nav table.
  - **DRIFT FOUND:** `agent/Skills/critic-agent-prompts.md` exists on disk but is NOT listed in the Skills nav table.
- **Issues:** 2 low-severity nav table omissions.

### Step 4: Check root Bonsai/CLAUDE.md accuracy
- **Action:** Compared project structure tree in `Bonsai/CLAUDE.md` against actual directory structure.
- **Result:**
  - **DRIFT FOUND:** `internal/nonint/` package entirely absent from project structure tree (added in Plan 41).
  - **DRIFT FOUND:** `cmd/completion.go` not listed in cmd/ section (added during Plan 41 or earlier).
  - **DRIFT FOUND:** `internal/generate/catalog_snapshot_unix.go`, `catalog_snapshot_windows.go`, `list_snapshot.go` not listed (added Plan 40/41).
  - **DRIFT FOUND:** `station/agent/Sensors/` description lists only 5 sensors ("context-guard, scope-guard-files, session-context, status-bar, routine-check") but 10 sensors are actually installed (also: agent-review, compact-recovery, dispatch-guard, subagent-stop-review, statusline).
- **Issues:** 4 items; low-severity structural drift.

### Step 5: Check code-index.md accuracy
- **Action:** Ran `grep -n` probes across all files referenced in code-index.md to verify function locations against documented line numbers.
- **Result:** Significant drift across all files touched by Plan 41 (which added ~100-140 lines to files by inserting non-interactive functions). Specific findings below.
  - `generate.go`: all 8 documented functions are off by approximately +100 lines.
  - CLI entry functions: all 6 of the drifted commands are off by +8 to +138 lines.
  - `remove.go` helpers: all 5 functions off by +138 lines.
  - `add.go` helpers: all 5 functions off by +35 lines.
  - `init_flow.go` helpers: 3 functions off by +88 lines.
  - `showWriteResults()` (listed at root.go `:201`) — function no longer exists in root.go at all.
  - New `ForAgent` variants of `SettingsJSON`, `PathScopedRules`, `WorkflowSkills` not documented.
  - `internal/nonint/` package entirely absent.
  - `list_snapshot.go`, `catalog_snapshot_unix.go`, `catalog_snapshot_windows.go` not documented.
- **Issues:** 2 high-severity blocks (major line drift, missing package), 3 medium-severity items.

### Step 6: Report findings
- **Action:** Compiled all drift items into the findings table below.
- **Issues:** See Findings Summary.

### Step 7: Update dashboard
- **Action:** Updated `routines.md` — Doc Freshness Check row Last Ran → 2026-07-11, Next Due → 2026-07-18, Status → done.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `generate.go` function line numbers all drifted by ~+100 lines — `Scaffolding` (:360→:401), `SettingsJSON` (:473→:564), `WorkspaceClaudeMD` (:725→:826), `RoutineDashboard` (:1010→:1111), `EnsureRoutineCheckSensor` (:972→:1073), `PathScopedRules` (:1164→:1265), `WorkflowSkills` (:1228→:1329), `AgentWorkspace` (:1359→:1460) | `station/code-index.md` — Generator section | Flagged for user — proposed line number refresh in Notes section |
| 2 | HIGH | CLI entry function line numbers drifted: `runInit` (27→35), `runAdd` (56→73), `runRemove` (34→67), `runList` (18→39), `runCatalog` (23→39), `runUpdate` (19→51); `remove.go` helpers all off by +138; `add.go` helpers all off by +35; `init_flow.go` helpers off by +88 | `station/code-index.md` — CLI Commands section | Flagged for user |
| 3 | MEDIUM | `showWriteResults()` listed at root.go `:201` in code-index but function no longer exists in root.go | `station/code-index.md` — Shared Helpers section | Flagged for user |
| 4 | MEDIUM | `internal/nonint/` package (Plan 41 headless-CLI contract) entirely absent from code-index | `station/code-index.md` | Flagged for user |
| 5 | MEDIUM | `internal/generate/list_snapshot.go` not documented in code-index | `station/code-index.md` — Generator section | Flagged for user |
| 6 | MEDIUM | `catalog_snapshot_unix.go` and `catalog_snapshot_windows.go` not documented in code-index | `station/code-index.md` — Generator section | Flagged for user |
| 7 | LOW | `agent/Workflows/plan-grilling.md` exists on disk but not listed in Workflows nav table | `station/CLAUDE.md` — Workflows table | Flagged for user |
| 8 | LOW | `agent/Skills/critic-agent-prompts.md` exists on disk but not listed in Skills nav table | `station/CLAUDE.md` — Skills table | Flagged for user |
| 9 | LOW | `internal/nonint/` package missing from project structure tree | `Bonsai/CLAUDE.md` — Project Structure section | Flagged for user |
| 10 | LOW | `cmd/completion.go` and 3 generate/ files (`catalog_snapshot_unix.go`, `catalog_snapshot_windows.go`, `list_snapshot.go`) missing from project structure tree | `Bonsai/CLAUDE.md` — Project Structure section | Flagged for user |
| 11 | LOW | Sensors listing in root CLAUDE.md station structure block names only 5 of 10 installed sensors | `Bonsai/CLAUDE.md` — Project Structure section | Flagged for user |
| 12 | INFO | Previously-broken `agent/Skills/bonsai-model.md` nav link now resolves — prior flag is cleared | `station/CLAUDE.md` — Skills table | Resolved (no action needed) |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **[HIGH] code-index.md line number refresh** — All generate.go, CLI entry, and helper function line numbers are drifted following Plan 41's non-interactive mode additions. A targeted refresh of the Generator and CLI Commands sections in `station/code-index.md` is recommended. This can be done autonomously via a Tier 1 patch agent dispatch if desired.

2. **[MEDIUM] code-index.md: nonint package documentation** — `internal/nonint/` (Plan 41 headless CLI contract package) is entirely undocumented. Worth adding a section to code-index.md covering: `nonint.go` (entry), `config.go` (LoadConfig), `runner.go`, `events.go` (event shapes), `result.go` (result contract), `remove.go`/`update.go` command-specific helpers.

3. **[MEDIUM] code-index.md: showWriteResults stale entry** — Function no longer in root.go; entry should be removed or the function located elsewhere.

4. **[LOW] station/CLAUDE.md: nav table omissions** — `plan-grilling.md` (Workflows) and `critic-agent-prompts.md` (Skills) are undocumented in nav tables. Both are minor additions — can be added directly by the Tech Lead in a quick edit if desired.

5. **[LOW] root CLAUDE.md: structural drift** — `internal/nonint/`, `cmd/completion.go`, 3 new generate/ files, and expanded Sensors listing are missing from the project structure tree. Low urgency but creates confusion when onboarding or reasoning about the codebase structure.

---

## Notes for Next Run

- The `code-index.md` line numbers will continue to drift as long as non-interactive functions are added above existing functions. Consider whether line numbers are worth maintaining at all vs. just listing function names + file paths. The cost of keeping them accurate is a Tier 1 dispatch after every plan that touches cmd/ or internal/generate/.
- `internal/nonint/` documentation gap is the highest-value addition to code-index.md — this package is the new headless-CLI surface and Tech Lead should be able to navigate it efficiently.
- INDEX.md catalog count "~50" will need an update when count passes ~55+ — currently 53 so still accurate.
- All station/CLAUDE.md nav links pass — no broken links found this cycle.
