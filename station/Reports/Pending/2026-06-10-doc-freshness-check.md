---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-10
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
- **Duration:** ~10 min
- **Files Read:** 10
  - `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`
  - `/home/user/Bonsai/station/INDEX.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/code-index.md`
  - `/home/user/Bonsai/station/agent/Core/memory.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/CLAUDE.md`
  - `/home/user/Bonsai/internal/nonint/nonint.go`
- **Files Modified:** 3
  - `/home/user/Bonsai/station/Reports/Pending/2026-06-10-doc-freshness-check.md` (this file)
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** `git log --oneline -30`, `git log --oneline --since="2026-05-04"`, `git diff --name-only HEAD~10 HEAD`, file existence checks via Bash for loops, `grep` via Grep tool
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history

Read `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`, and `station/agent/Core/memory.md`. Ran `git log --oneline --since="2026-05-04"` to get all commits since the last doc freshness check.

**Key commits since 2026-05-04:**
- `584b82b` — `fix(generate): bake absolute paths into sensor hook commands (v0.4.3)` (#106)
- `021da41` — `fix(settings): bake absolute paths into sensor hooks + file upstream` (#105)
- `a4ab5ac` — `chore(station): hand off Plan 38 ownership to Bonsai-Eval tech-lead` (#104)
- `02890de` — `chore(station): Plan 39 wrap — v0.4.2 shipped` (#103)
- `410a5f1` — `feat(nonint): bonsai init/add --non-interactive --from-config (v0.4.2)`
- Multiple station/, Backlog, Status, and dependency changes

**New features not reflected in docs:**
1. `internal/nonint/` package (6 files: `nonint.go`, `config.go`, `events.go`, `runner.go`, and tests) — entirely new package, not in `code-index.md` or `CLAUDE.md` project structure tree
2. `bonsai completion` command (`cmd/completion.go`) — added via PR #78, not in `code-index.md` CLI Commands table
3. `SettingsJSONForAgent()` function added to `internal/generate/generate.go` — not in `code-index.md` Generator section
4. Absolute-path baking behavior change in `SettingsJSON` (v0.4.3) — behavioral change not surfaced in docs

### Step 2 — Check INDEX.md accuracy

INDEX.md tech stack appears accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, yaml.v3, text/template, embed.FS). No toolchain drift.

**Drift found:**
- Key Metrics table: `CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate)` — should be 9, `completion` is now a first-class public command (has `--help` text, is in README under "Shell completion")
- Architecture Overview text block `cmd/ ← CLI commands: init, add, remove, list, catalog, update, guide, validate` — same 8→9 drift

### Step 3 — Check navigation links

Verified all links in `station/CLAUDE.md` navigation tables:
- Core files (identity.md, memory.md, self-awareness.md): all OK
- Protocol files (memory, scope-boundaries, security, session-start): all OK
- Workflow files (code-review, planning, pr-review, security-audit, session-logging, test-plan, session-wrapup, issue-to-implementation, routine-digest): all OK
- Skill files (planning-template, review-checklist, issue-classification, pr-creation, bubbletea, bonsai-model): all OK
- Routine files (all 7): all OK
- Sensor files (all 10): all OK
- External references (.bonsai.yaml, .bonsai/catalog.json, Playbook/*, Logs/*, Reports/*): all OK

**Result: 0 broken links.** All navigation links resolve to real files.

### Step 4 — Report findings

Findings compiled below. Per procedure: flagged for user decision, not executed.

### Step 5 — Update dashboard

`agent/Core/routines.md` Doc Freshness Check row updated: `Last Ran → 2026-06-10`, `Next Due → 2026-06-17`.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `internal/nonint/` package (6 files: `nonint.go`, `config.go`, `events.go`, `runner.go` + 2 tests) not documented — missing from `code-index.md` Generator section and `CLAUDE.md` project structure tree | `station/code-index.md`, `/home/user/Bonsai/CLAUDE.md` | Flagged for user |
| 2 | Low | `bonsai completion` command (`cmd/completion.go`, added PR #78) missing from `code-index.md` CLI Commands table | `station/code-index.md` lines 18–29 | Flagged for user |
| 3 | Low | INDEX.md Key Metrics: `CLI commands | 8` should be `9` (completion added). Architecture overview text block has same count drift | `station/INDEX.md` lines 33, 63 | Flagged for user |
| 4 | Low | `SettingsJSONForAgent()` function (added in v0.4.3 alongside absolute-path fix, used in `cmd/add.go`) not in `code-index.md` Generator section | `station/code-index.md` line 162 | Flagged for user |
| 5 | Info | Navigation links: 0 broken links found. All 31 checked links resolve correctly. | `station/CLAUDE.md` | No action needed |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **Add `internal/nonint/` to `station/code-index.md`** — medium priority. New package introduced in v0.4.2 handles non-interactive mode. Key public API: `RunInit()`, `RunAdd()`, `LoadConfig()`, `LoadOverlay()`, `EmitFile()`, `EmitSummary()`, `EmitWarning()`, and exit code constants (`ExitOK`, `ExitInvalidConfig`, `ExitRuntime`, `ExitWrongCWDForInit`). Also needs a row in `CLAUDE.md`'s project structure tree under `internal/`.

2. **Add `bonsai completion` to `station/code-index.md` CLI Commands table** — low priority. Entry point: `cmd/completion.go`, entry function `completionCmd`. Subcommands: bash, zsh, fish, powershell. Replaces Cobra's auto-generated completion to add install-snippet help text.

3. **Update INDEX.md CLI command count 8 → 9** — low priority. Two-location fix: Key Metrics table (line 33) and Architecture Overview text block (line 63). Both should list `completion` alongside the other 8 commands.

4. **Add `SettingsJSONForAgent()` to `code-index.md` Generator section** — low priority. Signature: `SettingsJSONForAgent(projectRoot string, agent *config.InstalledAgent, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error`. Purpose: single-agent scoped variant of `SettingsJSON()`. Used by `bonsai add` (line 517, 577 in `cmd/add.go`) and non-interactive path.

## Notes for Next Run

- The prior run (2026-05-04) flagged root `CLAUDE.md` project structure tree as "badly stale." Plan 37 (2026-05-07) partially addressed this, but v0.4.2 (nonint) and v0.4.3 (absolute paths in SettingsJSONForAgent) introduced new drift. If findings 1–4 above are addressed, root `CLAUDE.md` should also get the `nonint/` entry in its `internal/` tree.
- The `station/code-index.md` staleness was noted in the 2026-05-07 Backlog Hygiene report as a recurring item — this run confirms it still has drift from v0.4.2/v0.4.3 features.
- No broken navigation links: this is the first clean link-check since routines started. Maintaining 0-broken is the high-value outcome to preserve.
