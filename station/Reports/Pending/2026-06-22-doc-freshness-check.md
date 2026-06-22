---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-22
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~10 min
- **Files Read:** 12 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/CLAUDE.md`, `station/code-index.md`, `station/Playbook/Status.md`, `station/Logs/RoutineLog.md`, `cmd/remove.go`, `internal/generate/generate.go`, `internal/nonint/` (runner.go, result.go, nonint.go), `docs/agent-interface.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (entry append)
- **Tools Used:** Read, Bash (git log, grep, ls), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation vs recent git history
- **Action:** Ran `git log --since="7 days ago"` with `--name-status` and `--since="30 days ago"` for broader context. Identified all commits in the window.
- **Result:** One commit in the 7-day window: `9940936` (2026-06-22) — backlog-hygiene routine run. Looking back 30 days: heavy Plan 41 activity on 2026-06-16 (5 PRs merged: #120/#122/#123/#121/#125) introducing `internal/nonint/` package, `cmd/remove.go` major expansion, `docs/agent-interface.md`, headless CLI contract. Also Plan 40 Phases 1-3 (2026-06-13) expanding `internal/generate/generate.go` significantly.
- **Issues:** Station docs have not been updated to reflect the `internal/nonint/` package (Plan 41) or the line-number drift in generate.go (Plan 40). Both plans shipped before the last doc-freshness run (2026-05-04) but postdate it — the routine has been overdue since 2026-05-11.

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` in full. Verified tech stack table, key metrics, architecture overview diagram, and document registry.
- **Result:**
  - **Tech Stack table:** Accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, single binary / embed.FS all correct.
  - **Key Metrics — Agent types:** 6 (tech-lead, fullstack, backend, frontend, devops, security) — CORRECT (confirmed via `ls catalog/agents/`).
  - **Key Metrics — Catalog items:** "~50" — actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). Within the approximation; no drift flag needed.
  - **Key Metrics — CLI commands:** 8 (init, add, remove, list, catalog, update, guide, validate) — CORRECT. Note: `bonsai completion` also exists (`cmd/completion.go`) but is a meta/shell-completion command typically omitted from feature counts. No change needed.
  - **Architecture Overview diagram:** Lists `internal/catalog/`, `internal/config/`, `internal/generate/`, `internal/validate/`, `internal/wsvalidate/`, `internal/tui/`. **Missing:** `internal/nonint/` — the headless CLI core package introduced by Plan 41 (Phases 1-3, 2026-06-16). This is a meaningful omission since `nonint/` is now a peer package handling all mutating headless execution paths.
- **Issues:** One drift item — `internal/nonint/` absent from architecture overview.

### Step 3: Check navigation links
- **Action:** Listed files in `station/agent/Core/`, `station/agent/Protocols/`, `station/agent/Workflows/`, `station/agent/Skills/`, `station/agent/Sensors/`, `station/agent/Routines/`. Cross-checked all linked files in `station/CLAUDE.md` navigation tables.
- **Result:** All navigation links in `station/CLAUDE.md` resolve to real files:
  - Core: identity.md, memory.md, self-awareness.md, routines.md — all present.
  - Protocols: memory.md, scope-boundaries.md, security.md, session-start.md — all present.
  - Workflows: code-review.md, issue-to-implementation.md, plan-grilling.md, planning.md, pr-review.md, routine-digest.md, security-audit.md, session-logging.md, session-wrapup.md, test-plan.md — all present.
  - Skills: bonsai-model.md, bubbletea.md, critic-agent-prompts.md, issue-classification.md, planning-template.md, pr-creation.md, review-checklist.md — all present.
  - Sensors: all 10 listed in CLAUDE.md sensors table confirmed present.
  - Routines: all 7 listed routines confirmed present.
  - Bonsai Reference: `../.bonsai/catalog.json`, `../.bonsai.yaml`, `agent/Skills/bonsai-model.md` — all present.
- **Issues:** None — all navigation links are intact.

### Step 4: Check code-index.md accuracy
- **Action:** Compared `station/code-index.md` line references against actual source files for recently changed code (Plan 40 + Plan 41). Focused on `cmd/remove.go`, `internal/generate/generate.go`, and checked for `internal/nonint/` coverage.
- **Result:**
  - **`cmd/remove.go` — Remove Helpers section:** All 5 helper functions have drifted significantly from Plan 41 Phase 3 (which heavily expanded remove.go with headless adapter code):
    - `runRemoveItem()`: indexed `:290`, actual `428` (+138)
    - `runRemoveItemAction()`: indexed `:565`, actual `703` (+138)
    - `agentItemList()`: indexed `:618`, actual `756` (+138)
    - `itemIsRequired()`: indexed `:667`, actual `805` (+138)
    - `itemDisplayName()`: indexed `:693`, actual `831` (+138)
    - The main entry `runRemove()` at `:67` — indexed as `:34`. This is partially correct (`:34` is where the `removeCmd` is registered in `init()`, not where `runRemove()` is defined; actual `runRemove` is at `:67`).
  - **`internal/generate/generate.go` — Core Generation Functions section:** All 5 tracked functions have drifted from Plan 40 Phase 1-3 expansion:
    - `Scaffolding()`: indexed `:360`, actual `401` (+41)
    - `SettingsJSON()`: indexed `:473`, actual `564` (+91)
    - `WorkspaceClaudeMD()`: indexed `:725`, actual `826` (+101)
    - `AgentWorkspace()`: indexed `:1359`, actual `1460` (+101)
    - `RoutineDashboard()`: indexed `:1010`, actual `1111` (+101)
  - **`internal/nonint/` package:** Entirely absent from `code-index.md`. Plan 41 introduced this package with public functions `RunInit()`, `RunAdd()` (runner.go), and `Result`, event types, `config.go`, `remove.go`, `update.go`. This is a significant omission.
  - **`update.go` — `runUpdate()` line:** indexed `:19`, actual `51` (`:19` is `init()`, not `runUpdate()`). Minor — same issue as remove.go entry.
  - **`docs/agent-interface.md`:** New file from Plan 41 Phase 5. Not referenced in `station/code-index.md` (which covers Go source, not docs — this is expected, no flag needed).
- **Issues:** Two significant drifts — stale line numbers in remove.go helpers and generate.go functions; one missing section for `internal/nonint/`.

### Step 5: Update dashboard and log
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for "Doc Freshness Check" (Last Ran → 2026-06-22, Next Due → 2026-06-29, Status → done). Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Both files updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `internal/nonint/` package entirely missing from architecture overview | `station/INDEX.md` — Architecture Overview diagram | Flagged for user — propose adding one line: `internal/nonint/ ← headless mutating cores (init/add/update/remove) — pure *Result functions, JSONL/exit contract` |
| 2 | Medium | `cmd/remove.go` Remove Helpers section — all 5 line numbers stale (+138 lines off from Plan 41 expansion) | `station/code-index.md` lines 67–71 | Flagged for user — propose updating to actual line numbers: runRemoveItem:428, runRemoveItemAction:703, agentItemList:756, itemIsRequired:805, itemDisplayName:831 |
| 3 | Medium | `internal/generate/generate.go` Core Generation Functions — all 5 line numbers stale (+41 to +101 lines off from Plan 40 expansion) | `station/code-index.md` lines 161–165 | Flagged for user — propose updating to: Scaffolding:401, SettingsJSON:564, WorkspaceClaudeMD:826, AgentWorkspace:1460, RoutineDashboard:1111 |
| 4 | Medium | `internal/nonint/` package entirely absent from code-index | `station/code-index.md` — no section exists | Flagged for user — propose adding new section after Validate section documenting RunInit, RunAdd, Result, exit constants |
| 5 | Low | `bonsai remove` entry in CLI commands table points to `:34` (is the `init()` func) not `runRemove()` at `:67` | `station/code-index.md` line 24 | Flagged for user — minor, propose correcting to `:67` |
| 6 | Low | `bonsai update` entry points to `:19` (is the `init()` func) not `runUpdate()` at `:51` | `station/code-index.md` line 27 | Flagged for user — minor, propose correcting to `:51` |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

All findings require user decision since this routine is audit-only (per procedure: "Propose updates but don't execute — flag for user decision").

**Proposed updates for user to execute (or delegate to Plan 37-style doc refresh):**

1. **`station/INDEX.md` — Architecture Overview:** Add `internal/nonint/` line to the architecture diagram:
   ```
   internal/nonint/      ← headless mutating cores (init/add/update/remove) — pure *Result functions, JSONL/exit contract
   ```

2. **`station/code-index.md` — Remove Helpers section (lines 67–71):** Update line numbers:
   - `runRemoveItem()`: `:290` → `:428`
   - `runRemoveItemAction()`: `:565` → `:703`
   - `agentItemList()`: `:618` → `:756`
   - `itemIsRequired()`: `:667` → `:805`
   - `itemDisplayName()`: `:693` → `:831`

3. **`station/code-index.md` — Core Generation Functions (lines 161–165):** Update line numbers:
   - `Scaffolding()`: `:360` → `:401`
   - `SettingsJSON()`: `:473` → `:564`
   - `WorkspaceClaudeMD()`: `:725` → `:826`
   - `AgentWorkspace()`: `:1359` → `:1460`
   - `RoutineDashboard()`: `:1010` → `:1111`

4. **`station/code-index.md` — New section for `internal/nonint/`:** Add after the Validate section:
   ```markdown
   ## Headless CLI Cores (`internal/nonint/`) — Plan 41

   Pure mutating cores for non-interactive / MCP use. Each is a typed function: options in, `*Result` out — no prompts, no os.Exit.

   | Function | File | Purpose |
   |----------|------|---------|
   | `RunInit()` | `runner.go:73` | Headless init — create project config + generate workspace |
   | `RunAdd()` | `runner.go:150` | Headless add — merge overlay config + generate agent |
   | `Run()` (update) | `update.go` | Headless update — re-render all agents |
   | `Run()` (remove) | `remove.go` | Headless remove — uninstall agent or item |

   | Type | Purpose |
   |------|---------|
   | `Result` | Shared output shape — WriteResult + summary + warnings |
   | Exit constants | `ExitOK=0`, `ExitInvalidConfig=2`, `ExitRuntime=3`, `ExitWrongCWDForInit=4`, `ExitConflict=5` (in `runner.go`) |
   ```

5. **`station/code-index.md` — CLI Commands table:** Minor corrections:
   - `bonsai remove`: `:34` → `:67` (where `runRemove()` is defined)
   - `bonsai update`: `:19` → `:51` (where `runUpdate()` is defined)

**Suggested next action:** Either handle these as a quick in-session edit (5 files, all mechanical line-number corrections) or queue as a Plan 37-style doc refresh task in the Backlog.

## Notes for Next Run

- Plan 41 (`internal/nonint/`) introduced a significant new package that needs code-index coverage; the architecture diagram in INDEX.md also needs updating.
- Plan 40 expansion of `generate.go` caused broad line-number drift in code-index — when large features ship, schedule a code-index refresh as a post-ship cleanup.
- The `cmd/remove.go` and `cmd/update.go` entry function line numbers in the CLI Commands table point to `init()` not the actual command function — these have been wrong since at least Plan 31. Worth fixing at next pass.
- All navigation links in CLAUDE.md are intact — no broken links found.
- Routine was overdue by ~7 weeks (last ran 2026-05-04, due 2026-05-11). Consider running doc-freshness-check after each major plan ships rather than waiting for the 7-day cycle.
