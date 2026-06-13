---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-13
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
- **Duration:** ~12 min
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Status.md`, `/home/user/Bonsai/CLAUDE.md` (root), `station/CLAUDE.md`, `station/code-index.md`, `internal/generate/generate.go` (line scan), `internal/catalog/catalog.go` (line scan), `cmd/root.go`, `cmd/add.go`, `cmd/remove.go`, `cmd/init_flow.go`, `cmd/catalog.go`, `cmd/validate.go`, `internal/generate/catalog_snapshot.go`
- **Files Modified:** 0 — audit-only routine (findings flagged for user decision)
- **Tools Used:** Read, Bash (git log, ls, grep, head)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read `station/INDEX.md`, `station/CLAUDE.md`, `station/Playbook/Status.md`, and root `CLAUDE.md`. Compared against last 7 days of git commits (17 commits from v0.4.2/Plan 39, Plan 40 Phases 1-3, and backlog-hygiene routine).
- **Result:** Plan 40 (Odysseus integration) shipped Phases 1-3 on main, adding: frozen v1 schemas, root-relative scaffolding (manifest + memory), project-level `validate` pass (`internal/validate/project.go`), and memory-routing docs. These additions are partially reflected in docs but create drift items.
- **Issues:** See findings below.

### Step 2: Check INDEX.md accuracy
- **Action:** Verified tech stack, folder structure, CLI command count, catalog item count, and agent type count in `station/INDEX.md`.
- **Result:**
  - Tech stack: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS).
  - Agent types: 6 — correct (backend, devops, frontend, fullstack, security, tech-lead).
  - CLI commands: listed as 8 — technically 9 now (completion command added via external contribution v0.5.0 era, merged PR #78). The root CLAUDE.md still lists 8 matching INDEX.md but `cmd/completion.go` exists.
  - Catalog items: listed as "~50" — actual: 18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines = 53 total (not counting scaffolding). Still within "~50" range but could say ~55.
  - Architecture diagram: missing `internal/nonint/` package (added for `--non-interactive` mode, Plan 39 / v0.4.2), and Plan 40's `internal/validate/project.go` addition.
- **Issues:** CLI count drift (8→9), nonint package missing from arch diagram.

### Step 3: Check navigation links
- **Action:** Verified all links in `station/CLAUDE.md` nav tables against actual files on disk. Checked Core, Protocols, Workflows, Skills, Routines, and Sensors.
- **Result:**
  - **Core links:** All 4 resolve (`identity.md`, `memory.md`, `self-awareness.md`, `routines.md`). ✓
  - **Protocol links:** All 4 resolve (`memory.md`, `scope-boundaries.md`, `security.md`, `session-start.md`). ✓
  - **Workflow links:** All 9 listed links resolve. ✓ BUT `agent/Workflows/plan-grilling.md` exists on disk and is NOT listed in the nav (added 2026-06-13 via Plan 40 grilling pipeline).
  - **Skills links:** All 6 listed links resolve including `bonsai-model.md` (previously flagged as broken — now confirmed fixed). ✓ BUT `agent/Skills/critic-agent-prompts.md` exists on disk and is NOT listed in the nav.
  - **Routines links:** All 7 resolve. ✓
  - **Sensor links:** All 10 listed resolve. ✓
  - **Bonsai Reference section:** `.bonsai/catalog.json` link and `.bonsai.yaml` link — both should exist after `bonsai init`. Not validated (not part of station/ proper).
  - **Quick Triggers:** `/plan` and `/grill` slash commands exist at `station/.claude/commands/plan.md` and `station/.claude/commands/grill.md` but are NOT listed in the Quick Triggers table (the table shows `/planning` instead).
- **Issues:** 3 unlisted items: `plan-grilling.md` workflow, `critic-agent-prompts.md` skill, `/plan`+`/grill` slash commands.

### Step 4: Report findings
- **Action:** Compiled all drift items from Steps 1-3. Determined severity. Proposing updates, not executing.
- **Result:** 6 drift items found across 3 severity levels. None blocking. All are additive omissions (new features/files not reflected in docs), not stale or incorrect content.
- **Issues:** none beyond the findings themselves.

### Step 5: Update dashboard
- **Action:** Will update `agent/Core/routines.md` dashboard row for Doc Freshness Check.
- **Result:** Done (Last Ran → 2026-06-13, Next Due → 2026-06-20, Status → done).
- **Issues:** none.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | `internal/nonint/` package missing from INDEX.md Architecture Overview and root `CLAUDE.md` project-structure tree. Package was added for `--non-interactive` CLI mode (Plan 39 / v0.4.2). Contains: `nonint.go` (anchor), `config.go` (LoadConfig + applyDefaults), `events.go` (EmitFile/EmitSummary/EmitWarning), `runner.go` (RunInit + RunAdd). | `station/INDEX.md` arch diagram + root `CLAUDE.md` `internal/` tree | Flagged — not fixed (user decision) |
| 2 | medium | `code-index.md` line numbers for `internal/generate/generate.go` are all off by ~101 lines due to Plan 40's scaffolding additions. Drifted functions: `Scaffolding()` (360→401), `SettingsJSON()` (473→564), `WorkspaceClaudeMD()` (725→826), `AgentWorkspace()` (1359→1460), `RoutineDashboard()` (1010→1111), `EnsureRoutineCheckSensor()` (972→1073), `PathScopedRules()` (1164→1265), `WorkflowSkills()` (1228→1329). `cmd/add.go` lines also shifted (~17 lines). Root.go lines shifted ~2. | `station/code-index.md` | Flagged — not fixed (user decision) |
| 3 | medium | `plan-grilling.md` workflow added to `station/agent/Workflows/` (2026-06-13, Plan 40 grilling pipeline) but not listed in `station/CLAUDE.md` Workflows nav table. | `station/CLAUDE.md` Workflows section | Flagged — not fixed (user decision) |
| 4 | low | `critic-agent-prompts.md` skill added to `station/agent/Skills/` (2026-06-13) but not listed in `station/CLAUDE.md` Skills nav table. | `station/CLAUDE.md` Skills section | Flagged — not fixed (user decision) |
| 5 | low | `/plan` and `/grill` slash commands (at `station/.claude/commands/plan.md` and `grill.md`) exist but are not in the Quick Triggers table in `station/CLAUDE.md`. The Quick Triggers table still shows `/planning` as the planning trigger but `/plan` now exists as the primary command entry point. | `station/CLAUDE.md` Quick Triggers section | Flagged — not fixed (user decision) |
| 6 | low | CLI command count in `station/INDEX.md` and root `CLAUDE.md` shows 8 commands; `completion` command (`cmd/completion.go`) now exists (PR #78 external contribution). Count should be 9. | `station/INDEX.md` Key Metrics + root `CLAUDE.md` arch comment | Flagged — not fixed (user decision) |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

All 6 findings are flagged for user review / decision:

1. **Finding #1 (medium):** Add `internal/nonint/` to INDEX.md arch diagram and root `CLAUDE.md` project tree. Suggest adding a line between `internal/wsvalidate/` and `internal/tui/`: `internal/nonint/ ← non-interactive orchestrators (RunInit, RunAdd) + JSONL event emitters`.

2. **Finding #2 (medium):** `code-index.md` generate.go line numbers are ~101 off. This is recurring drift from Plan 37 (which last synced them). Recommend a doc-refresh sweep similar to Plan 37, or promoting a `code-index` sub-step in this routine to catch line drift earlier. Consider whether code-index line numbers are worth maintaining vs switching to function-name-only references.

3. **Finding #3 (low):** Add `plan-grilling.md` to `station/CLAUDE.md` Workflows nav. Suggested row: `| Running adversarial plan review; Dispatching 6-critic grilling pipeline on a numbered plan | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |`

4. **Finding #4 (low):** Add `critic-agent-prompts.md` to `station/CLAUDE.md` Skills nav. Suggested row: `| Running the plan-grilling workflow; Critic agent prompt templates for the 6 adversarial reviewers | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |`

5. **Finding #5 (low):** Reconcile Quick Triggers: `/plan` now exists at `station/.claude/commands/plan.md` (drives full planning → grilling → confirm flow). Consider updating Quick Triggers to show `/plan` as the primary trigger for planning, and adding a row for grilling (`/grill`).

6. **Finding #6 (low):** Increment CLI command count from 8 to 9 in INDEX.md and optionally add `completion` to the command list in root CLAUDE.md.

## Notes for Next Run

- Findings #3 and #4 (plan-grilling workflow + critic-agent-prompts skill) are marked as "full Bonsai-catalog integration pending" in the Backlog (added 2026-06-13). If that integration ships before the next run, CLAUDE.md nav will be auto-generated and this won't be an issue.
- Finding #2 (code-index line numbers) is a recurring pattern — see 2026-05-04 and 2026-05-07 runs. Consider scoping into a Plan or adding a line-number check substep to this routine.
- `bonsai-model.md` link was previously flagged as broken in 2026-05-04 and 2026-05-07 runs — confirmed RESOLVED (file exists at correct path).
- All previous drift from 2026-05-04 (root CLAUDE.md structure, INDEX arch diagram, validate command) is RESOLVED except nonint (new since then).
