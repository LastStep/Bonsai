---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-29
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
- **Duration:** ~6 min
- **Files Read:** 9 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/CLAUDE.md`, `station/code-index.md`, `CHANGELOG.md`, `cmd/completion.go`, `internal/nonint/runner.go`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log append)
- **Tools Used:** Read, Bash (git log, ls, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read station/INDEX.md and CHANGELOG.md; ran `git log --oneline --since="2026-05-04"` to enumerate commits since the last doc freshness check.
- **Result:** 30+ commits landed since 2026-05-04. Key changes were Plan 40 (in-repo memory graph, project manifest, validate project-level pass, `bonsai guide formats`) and Plan 41 (headless CLI contract — `internal/nonint/` package, `--non-interactive`/`--yes`/`--from`/`--skip-conflicts` flags, `list --json`, `docs/agent-interface.md`). Several of these are not reflected in station docs.
- **Issues:** See findings 1, 4, 7 below.

### Step 2: Check INDEX.md accuracy
- **Action:** Compared station/INDEX.md (Project Snapshot, Tech Stack, Key Metrics, Architecture Overview, Document Registry) against actual codebase state.
- **Result:** Tech stack table is accurate. Key Metrics counts are still correct (CLI commands = 8, agent types = 6, catalog items ~50). Architecture diagram is conceptually accurate but misses `internal/nonint/` as a new layer. Document Registry does not reference the newly added `docs/agent-interface.md`. The Project Snapshot does not mention v0.5.0 features shipped in Plans 40 and 41.
- **Issues:** Findings 4 and 7.

### Step 3: Check navigation links
- **Action:** Listed actual files in `station/agent/{Workflows,Skills,Core,Protocols,Sensors,Routines}/` and cross-referenced each against navigation tables in station/CLAUDE.md.
- **Result:**
  - Core links (identity.md, memory.md, self-awareness.md): all resolve.
  - Protocol links (memory.md, scope-boundaries.md, security.md, session-start.md): all resolve.
  - Sensor links (10 sensors): all resolve.
  - Routine links (7 routines): all resolve.
  - **Workflow links:** 9 listed in CLAUDE.md all resolve. But `station/agent/Workflows/plan-grilling.md` EXISTS and is NOT listed.
  - **Skill links:** 6 listed in CLAUDE.md all resolve. But `station/agent/Skills/critic-agent-prompts.md` EXISTS and is NOT listed.
  - External references (STATUS, Roadmap, Backlog, Plans/Active/, Standards/SecurityStandards.md, KeyDecisionLog.md, Reports/Pending/): all resolve.
  - **RESOLVED from 2026-05-04:** `agent/Skills/bonsai-model.md` broken link — file now exists, link resolves.
- **Issues:** Findings 2 and 3.

### Step 4: Check code-index.md accuracy
- **Action:** Compared station/code-index.md against actual `cmd/` and `internal/` directory contents.
- **Result:** `internal/nonint/` package (added by Plan 41 with 7 source files: runner.go, config.go, events.go, nonint.go, remove.go, result.go, update.go) is entirely absent from code-index. `cmd/completion.go` (shell completion command) is also undocumented. Line numbers for existing entries in generate.go and other files are likely stale given Plan 40/41 additions.
- **Issues:** Findings 1 and 5.

### Step 5: Report findings and log
- **Action:** Compiled findings table, wrote this report, updated dashboard, appended to RoutineLog.
- **Result:** 7 findings identified; no edits to docs (audit-only routine; all changes flagged for user decision).
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package (Plan 41 headless core, 7 source files) entirely absent from code-index | `station/code-index.md` | Flagged for user — propose adding a new section documenting nonint package entry points (exit codes, Result shapes, runner, events, config, remove core, update core) |
| 2 | MEDIUM | `plan-grilling.md` workflow exists but is not listed in CLAUDE.md Workflows navigation table | `station/CLAUDE.md` | Flagged for user — propose adding a row: "Running the 6-critic adversarial review pipeline for a plan | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md)" |
| 3 | MEDIUM | `critic-agent-prompts.md` skill exists but is not listed in CLAUDE.md Skills navigation table | `station/CLAUDE.md` | Flagged for user — propose adding a row: "Prompts for critic agent roles in the plan-grilling pipeline | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md)" |
| 4 | MEDIUM | `docs/agent-interface.md` (canonical headless CLI contract, Plan 41 Phase 5) not in Document Registry | `station/INDEX.md` | Flagged for user — propose adding row: "`docs/agent-interface.md` — canonical headless CLI contract: per-command flags, JSONL/JSON outputs, exit-code table — When integrating Bonsai with CI or AI clients" |
| 5 | LOW | `cmd/completion.go` (shell completion command) absent from code-index CLI Commands table | `station/code-index.md` | Flagged for user — low priority; propose adding a row to CLI Commands |
| 6 | LOW | Function line numbers in code-index.md likely stale for generate.go, cmd/remove.go, cmd/list.go, cmd/update.go — Plans 40/41 added significant code | `station/code-index.md` | Flagged for user — line-number refresh needed; defer to planned code-index sweep |
| 7 | LOW | INDEX.md Project Snapshot does not describe v0.5.0 feature additions (memory graph scaffolding, project manifest scaffolding, enhanced validate, headless CLI flags, `list --json`) | `station/INDEX.md` | Flagged for user — low urgency (snapshot-level doc); propose a brief "v0.5.0 highlights" callout or update to Architecture Overview |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

All 7 findings above require user decision before any edits are made. Priority recommendation:

1. **(HIGH)** Add `internal/nonint/` section to code-index.md — this is the most impactful gap for navigating Plan 41 work.
2. **(MEDIUM)** Add `plan-grilling.md` and `critic-agent-prompts.md` to station/CLAUDE.md navigation tables.
3. **(MEDIUM)** Add `docs/agent-interface.md` to station/INDEX.md Document Registry.
4. **(LOW)** Refresh code-index.md line numbers and add `cmd/completion.go` — consolidate with the code-index staleness backlog item from the 2026-05-04 run (still open per Backlog Hygiene 2026-08-29 flags).
5. **(LOW)** Update INDEX.md Project Snapshot to mention v0.5.0 additions.

**Previously open findings resolved this cycle:**
- `agent/Skills/bonsai-model.md` broken nav link — RESOLVED (file exists and link resolves).

## Notes for Next Run

- The code-index staleness finding has now been flagged for 3 consecutive doc freshness runs (2026-04-21, 2026-05-04, 2026-08-29). Consider promoting the code-index sweep to a Backlog P1 or assigning it to the next available plan slot.
- `plan-grilling.md` and `critic-agent-prompts.md` appear to have been intentionally added to agent/Skills and agent/Workflows but were never wired into the CLAUDE.md navigation — this pattern suggests these files may be in "draft" or "soft-installed" state. Confirm intended visibility before adding nav rows.
- v0.5.0 is unreleased as of this run. INDEX.md update can wait until after the release tag.
