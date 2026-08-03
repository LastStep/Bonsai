---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-03
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
- **Duration:** ~8 min
- **Files Read:** 12 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`, `/home/user/Bonsai/station/agent/Skills/critic-agent-prompts.md`, `/home/user/Bonsai/internal/nonint/nonint.go`, `/home/user/Bonsai/internal/nonint/result.go`, `/home/user/Bonsai/internal/nonint/config.go`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, head), Glob, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation vs recent git history
- **Action:** Ran `git log --oneline --since="2026-07-27"` and then `git log --oneline -20` to capture recent history. Read INDEX.md, Playbook/Status.md.
- **Result:** No commits in the last 7 days (last activity: 2026-06-16, Plan 41). Most recent changes were Plan 41 (headless CLI contract) and Plan 40 (frozen schemas + validate pass). These shipped 48 days ago — no new activity since then. Checked documentation against Plan 40 and Plan 41 changes to find any unreflected additions.
- **Issues:** Two documentation gaps found relative to Plan 41: (1) `internal/nonint/` package absent from code-index.md; (2) `bonsai completion` command count still says 8 in INDEX.md and code-index.md.

### Step 2: Check INDEX.md accuracy
- **Action:** Verified tech stack, agent types, catalog item count, and CLI command count against actual codebase.
- **Result:**
  - Tech stack: correct (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template). ✓
  - Agent types: 6 (tech-lead, fullstack, backend, frontend, devops, security) — matches `catalog/agents/` directory. ✓
  - Catalog items: "~50" — actual count is 53 meta.yaml files (18 skills + 9 workflows + 4 protocols + 12 sensors + 8 routines + 6 agents + scaffolding). Acceptable approximation. ✓
  - CLI commands: "8" — incorrect. `bonsai completion` (`cmd/completion.go`) was added by external contributor in 2026-05-07 (#78). Actual count is 9. **DRIFT.**
- **Issues:** CLI command count off by one (8 listed, 9 actual).

### Step 3: Check navigation links in station/CLAUDE.md and agent/ directories
- **Action:** Compared all links in station/CLAUDE.md navigation tables against actual file listings in `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/`, `agent/Sensors/`, `agent/Routines/`, and all external Playbook/Logs/Reports references.
- **Result:**
  - Core (4 files): all 4 linked files exist. ✓
  - Protocols (4 files): all 4 linked files exist. ✓
  - Workflows: 9 files listed in nav, 10 files exist on disk. `plan-grilling.md` is present at `agent/Workflows/plan-grilling.md` but NOT listed in the station/CLAUDE.md Workflows table. **DRIFT.**
  - Skills: 6 files listed in nav, 8 items exist on disk. `critic-agent-prompts.md` is present at `agent/Skills/critic-agent-prompts.md` but NOT listed in the station/CLAUDE.md Skills table. (`bubbletea/` is a sub-directory of additional bubbletea reference files — expected, not a navigation item.) **DRIFT.**
  - Sensors (10 files): all 10 linked files exist. ✓
  - Routines (7 files): all 7 linked files exist. ✓
  - External refs (Playbook/, Logs/, Reports/, code-index.md, .bonsai/catalog.json, .bonsai.yaml): all resolve. ✓
- **Issues:** Two unlisted files in navigation table (plan-grilling workflow + critic-agent-prompts skill).

### Step 4: Report findings
- **Action:** Consolidated all drift items into the findings table below. Per procedure, findings are flagged for user decision — no doc edits executed.
- **Result:** 3 findings total (1 medium, 2 low). All are doc gaps, no broken links.
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-08-03, Next Due → 2026-08-10, Status → done.
- **Result:** Done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `plan-grilling.md` (Workflow) exists but is NOT listed in the station/CLAUDE.md Workflows navigation table. It has defined trigger phrases ("grill the plan", "review plan NN", "critic pass", "team of agents review this") that are invisible without a nav entry. | `station/CLAUDE.md` → Workflows section; file: `agent/Workflows/plan-grilling.md` | Flagged for user. Proposed fix: add row under Workflows with trigger "Adversarial review of a drafted plan via critic agents; grilling a plan before dispatch" |
| 2 | Low | `critic-agent-prompts.md` (Skill) exists but is NOT listed in the station/CLAUDE.md Skills navigation table. It is consumed verbatim by `plan-grilling.md`. | `station/CLAUDE.md` → Skills section; file: `agent/Skills/critic-agent-prompts.md` | Flagged for user. Proposed fix: add row under Skills with trigger "Dispatching the 6 plan-grilling critic agents; plan-grilling workflow in progress" |
| 3 | Low | CLI command count in INDEX.md and code-index.md is "8" but `bonsai completion` was added as a 9th command in 2026-05-07 (external contribution #78). Both files need updating. | `station/INDEX.md` line 33; `station/code-index.md` CMD table | Flagged for user. Proposed fix: update count to "9" and add `bonsai completion` row to code-index.md CLI Commands table. |
| 4 | Low | `internal/nonint/` package (Plan 41, 2026-06-16) is entirely absent from code-index.md. This is the headless CLI core — contains `RunInit`, `RunAdd`, `RunUpdate`, `RunRemove` orchestrators, `Result`/`EmitJSONL`/`LoadConfig`, and exit codes (`ExitOK`/`ExitConflict`/`ExitRuntime` etc.). | `station/code-index.md` — no section for `internal/nonint/` | Flagged for user. Proposed fix: add `internal/nonint/` section to code-index.md documenting the package surface (Result, RunInit/RunAdd/RunUpdate/RunRemove, EmitJSONL, LoadConfig, exit codes). |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

All 4 findings are doc gaps requiring user decision. None are breaking — no links are broken, no files are missing. The most user-visible gap is **Finding #1** (plan-grilling not in navigation), which means the agent won't see its trigger in the Quick Triggers table. Recommend addressing #1 and #2 together (they're companion files), then #3 and #4 together (both are Plan 41/completion drift in code-index.md).

**Proposed station/CLAUDE.md additions:**

Workflows table — add row:
```
| Adversarial review of a drafted plan via 6 critic agents before dispatch; grilling a plan for blocks, risks, and verification gaps | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

Skills table — add row:
```
| Dispatching the 6 plan-grilling critic agents (Security, Architecture, Simplicity, Risk, Verification, Reality); consumed verbatim by plan-grilling.md | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

## Notes for Next Run

- No broken links found. All sensor, routine, protocol, core, and workflow files listed in CLAUDE.md resolve.
- All 4 findings are low-effort fixes — could be addressed in a single session pass.
- The `bubbletea/` sub-directory in `agent/Skills/` is intentional (referenced by bubbletea.md itself, not a navigation gap).
- Next run should also check if `docs/agent-interface.md` (Plan 41 contract doc) is referenced anywhere in station docs — it was mentioned in Status.md but not in the station navigation.
