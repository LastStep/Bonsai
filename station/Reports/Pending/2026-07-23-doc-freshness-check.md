---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-23
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
- **Duration:** ~8 minutes
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/CLAUDE.md` (via system context), `station/code-index.md`, `station/Playbook/Status.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/identity.md` (via dir listing), filesystem listing of cmd/, internal/, catalog/agents/, station/agent/ subdirectories
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log append)
- **Tools Used:** Read, Bash (git log, ls, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read station/INDEX.md, station/code-index.md, station/Playbook/Status.md, station/agent/Core/routines.md, station/Logs/RoutineLog.md. Ran `git -C /home/user/Bonsai log --oneline --since="80 days ago"` to capture commits since last run (2026-05-04).
- **Result:** 47 commits since last run. Key structural changes: Plan 39 (`bonsai init/add --non-interactive`, PR #102), PR #78 (`bonsai completion` subcommand added, commit `2eae9d4`), Plan 40 (Odysseus platform integration — frozen schemas, root-relative scaffolding, validate pass, guide Formats page), Plan 41 (Headless CLI contract — `internal/nonint/` package added, all mutating commands get `*Result` headless cores, `list --json`, exit code contract `ExitConflict=5`, `docs/agent-interface.md`). Also: plan-grilling pipeline added to station (commit `6995d4f`).
- **Issues:** None in data gathering.

### Step 2: Check INDEX.md accuracy
- **Action:** Compared station/INDEX.md tech stack, key metrics, architecture overview, and document registry against actual codebase (`ls cmd/`, `ls internal/`, `ls catalog/agents/`).
- **Result:** Tech stack row (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS) — **accurate**. Agent types (6) — **accurate** (tech-lead, fullstack, backend, frontend, devops, security confirmed in catalog/agents/). Architecture overview (cmd → internal/catalog, config, generate, validate, wsvalidate, tui → catalog embed.FS) — **accurate at the high level but missing `internal/nonint/`**. Document registry — **all paths resolve**.
- **Issues:** Key Metrics row "CLI commands: 8 (init, add, remove, list, catalog, update, guide, validate)" is stale — `bonsai completion` (PR #78, 2026-05-07) makes it 9.

### Step 3: Check navigation links
- **Action:** Verified all file paths referenced in station/CLAUDE.md navigation tables against actual filesystem listings of `station/agent/` subdirectories.
- **Result:**
  - Core files (identity.md, memory.md, self-awareness.md, routines.md) — all exist ✓
  - Bonsai reference (bonsai-model.md, ../.bonsai/catalog.json, ../.bonsai.yaml) — all exist ✓
  - Protocols (memory.md, scope-boundaries.md, security.md, session-start.md) — all exist ✓
  - Workflows (code-review, planning, pr-review, security-audit, session-logging, test-plan, session-wrapup, issue-to-implementation, routine-digest) — all 9 listed files exist ✓
  - Skills (planning-template, review-checklist, issue-classification, pr-creation, bubbletea, bonsai-model) — all 6 listed files exist ✓
  - Routines (all 7 listed) — all exist ✓
  - Sensors (all 10 listed) — all exist ✓
  - External References (Status.md, Roadmap.md, Standards/SecurityStandards.md, Plans/Active/, Backlog.md, KeyDecisionLog.md, Reports/Pending/) — all resolve ✓
  - **No broken navigation links found.**
- **Issues:** Two unlisted files found on disk: `agent/Workflows/plan-grilling.md` (added commit `6995d4f`, plan-grilling pipeline) and `agent/Skills/critic-agent-prompts.md` — both exist but have no entry in the CLAUDE.md navigation tables.

### Step 4: Check code-index.md accuracy
- **Action:** Read station/code-index.md in full. Cross-referenced CLI commands table against actual `cmd/` directory listing. Checked internal/ package list.
- **Result:** CLI Commands table has 9 named entries (root, init, add, remove, list, catalog, update, guide, validate) but `cmd/completion.go` exists on disk with no corresponding row. `internal/nonint/` package has 14 files (config.go, events.go, nonint.go, remove.go, result.go, runner.go, update.go plus tests) added during Plans 39 and 41 — no section exists in code-index.md for this package. This drift was previously flagged in the 2026-05-07 Backlog Hygiene log as "medium drift, uncaptured from 2026-05-04 Doc Freshness."
- **Issues:** Two gaps in code-index.md — missing completion command entry, missing nonint package section.

### Step 5: Report findings (flag for user)
- **Action:** Compiled all drift items into findings table below. No auto-edits made — all items flagged for user decision per procedure.
- **Result:** 5 findings across 4 files. No broken links. All drift is additive (new content not yet documented).

### Step 6: Update dashboard
- **Action:** Updated station/agent/Core/routines.md dashboard row for "Doc Freshness Check" (Last Ran → 2026-07-23, Next Due → 2026-07-30, Status → done).
- **Result:** Done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | Root CLAUDE.md project structure tree missing `cmd/completion.go` (added PR #78, 2026-05-07) and `internal/nonint/` package (added Plans 39/41). This was flagged HIGH in 2026-05-04 run for different files (Plans 22–35 drift); now further stale. | `/home/user/Bonsai/CLAUDE.md` | Flagged for user — no edit |
| 2 | Medium | station/code-index.md missing: (a) `bonsai completion` CLI entry, (b) `internal/nonint/` package section. Drift first flagged 2026-05-07. | `station/code-index.md` | Flagged for user — no edit |
| 3 | Low | station/INDEX.md Key Metrics shows "CLI commands: 8" — should be 9 after `bonsai completion` added May 2026 | `station/INDEX.md` | Flagged for user — no edit |
| 4 | Low | station/CLAUDE.md Workflows table has no entry for `plan-grilling.md` (file exists at `agent/Workflows/plan-grilling.md`, added commit `6995d4f`) | `station/CLAUDE.md` | Flagged for user — no edit |
| 5 | Low | station/CLAUDE.md Skills table has no entry for `critic-agent-prompts.md` (file exists at `agent/Skills/critic-agent-prompts.md`) | `station/CLAUDE.md` | Flagged for user — no edit |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

- **Root CLAUDE.md drift (carry-forward, compounding):** The project structure tree in `/home/user/Bonsai/CLAUDE.md` was flagged HIGH in the May 2026 run and has not been updated. Two new items have now compounded the gap: `cmd/completion.go` and `internal/nonint/`. Recommend targeting this in a near-term doc-refresh plan (or at minimum a quick edit pass — the completion entry is a one-liner and nonint needs a new table block).

- **code-index.md gaps:** `station/code-index.md` is the developer quick-nav for Go source. The `bonsai completion` command and the entire `internal/nonint/` package (headless result types, config, events, runner, remove, update cores — ~14 files) are undocumented there. Since code agents rely on this for navigation, these gaps reduce their accuracy when working with nonint or completion code. Recommend a targeted code-index refresh as a Plan or quick inline update.

- **station/INDEX.md CLI count:** One-line fix — "8" → "9" and add `completion` to the parenthetical list.

- **plan-grilling.md / critic-agent-prompts.md unlisted in CLAUDE.md:** Both files exist on disk. If the plan-grilling workflow is intentionally part of the agent's toolkit, it should have a row in the Workflows table. `critic-agent-prompts.md` likely supports the plan-grilling pipeline and may warrant a Skills entry. Alternatively, if these are internal/support files not meant for direct loading, that should be noted somewhere. User decision needed.

- **Backlog carry-forward (from 2026-07-23 Backlog Hygiene):** HOMEBREW_TAP_TOKEN PAT reminder date (2026-07-15) is 8 days past — rotate before next release.

## Notes for Next Run
- Root CLAUDE.md and code-index.md drift are now 80+ days old and compounding with each plan shipped. If still unresolved at next check (2026-07-30), recommend escalating to P1 backlog item.
- Check if `plan-grilling.md` / `critic-agent-prompts.md` nav decision has been made; close out finding 4/5 if resolved.
- `internal/nonint/` is a new first-class package — worth adding to the architecture diagram in INDEX.md as well when doing the code-index refresh.
- No broken links found this cycle — navigation table health is good.
