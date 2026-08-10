---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-10
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
- **Duration:** ~10 min
- **Files Read:** 14 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`, `/home/user/Bonsai/station/agent/Skills/critic-agent-prompts.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/CHANGELOG.md`, `/home/user/Bonsai/cmd/root.go`, `/home/user/Bonsai/cmd/completion.go`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** `git log`, `git log --name-only`, directory listing, file reads, link existence checks (bash for-loop)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read station/INDEX.md, station/CLAUDE.md, station/Playbook/Status.md, station/Playbook/Roadmap.md, station/code-index.md. Retrieved git log since 2026-05-04 (last run date) and git log --name-only to enumerate changed files.
- **Result:** Found 51 commits since 2026-05-04, spanning Plans 38–41, v0.4.2, v0.4.3 (hotfix), v0.5.0 (unreleased). Significant new features: `completion` subcommand (PR #78), headless CLI flags across all mutating commands (Plans 39, 41), in-repo memory graph scaffolding, `bonsai validate` project-level pass, `docs/agent-interface.md` contract. Also found two new files in station/agent/Workflows/ and station/agent/Skills/ with no corresponding nav entries in station/CLAUDE.md.
- **Issues:** 4 documentation drift items found (detailed below).

### Step 2: Check INDEX.md accuracy
- **Action:** Compared station/INDEX.md tech stack, folder structure, key metrics, and architecture overview against current codebase state. Checked cmd/ directory listing and CHANGELOG for current version.
- **Result:** Tech stack table is accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, single binary). Architecture overview is accurate. Agent types count (6) is accurate. Catalog items count ("~50") is approximately accurate (53 items counted: 18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). **CLI commands count is stale**: says "8" but is actually 9 — `completion` was added in PR #78 (2026-05-07). `docs/agent-interface.md` (Plan 41 Phase 5) is not listed in Document Registry but may be intentional since it's for external consumers.
- **Issues:** 1 stale metric (CLI commands count), 1 possible omission (agent-interface.md in Document Registry — user decision).

### Step 3: Check navigation links
- **Action:** Verified all file paths referenced in station/CLAUDE.md navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References) against actual filesystem using bash existence check loop.
- **Result:** All 51 link targets resolve to real files. No broken links found. However, two files exist in the workspace that are NOT yet listed in station/CLAUDE.md:
  - `station/agent/Workflows/plan-grilling.md` — added 2026-06-13 (feat(station): add plan-grilling pipeline)
  - `station/agent/Skills/critic-agent-prompts.md` — added 2026-06-13 alongside plan-grilling
  Also: `station/code-index.md` does not list the `completion` command in its CLI Commands table.
- **Issues:** 2 unlisted files (plan-grilling workflow, critic-agent-prompts skill) in station/CLAUDE.md nav tables; `completion` command absent from code-index.md.

### Step 4: Report findings
- **Action:** Compiled 4 drift findings with specific severity, location, and proposed fix for each. All flagged for user decision per routine procedure ("propose updates but don't execute").
- **Result:** Findings documented in Findings Summary below.
- **Issues:** None — findings compiled cleanly.

### Step 5: Update dashboard
- **Action:** Updated the Doc Freshness Check row in `/home/user/Bonsai/station/agent/Core/routines.md` to set Last Ran → 2026-08-10, Next Due → 2026-08-17, Status → done.
- **Result:** Dashboard updated successfully.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | CLI commands metric stale — says "8" but `completion` was added 2026-05-07 (PR #78); count is now 9 | `station/INDEX.md` — Key Metrics table | Flagged; proposed update: change count to "9" and add `completion` to the parenthetical list |
| 2 | Medium | `completion` command absent from CLI Commands table | `station/code-index.md` — CLI Commands section | Flagged; proposed update: add row `bonsai completion \| cmd/completion.go:17 \| completionCmd — generate shell completion script (bash/zsh/fish/powershell)` |
| 3 | Low | `plan-grilling.md` workflow exists but has no nav entry in station/CLAUDE.md Workflows table | `station/CLAUDE.md` — Workflows section; file: `station/agent/Workflows/plan-grilling.md` | Flagged; proposed entry: "Adversarial review of a drafted plan via 6 critic agents before dispatch. Trigger: 'grill the plan'" — activate when grilling a plan |
| 4 | Low | `critic-agent-prompts.md` skill exists but has no nav entry in station/CLAUDE.md Skills table | `station/CLAUDE.md` — Skills section; file: `station/agent/Skills/critic-agent-prompts.md` | Flagged; proposed entry: "Verbatim prompt templates for the 6 plan-grilling critic agents. Load when executing plan-grilling workflow" |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**1. (Medium) station/INDEX.md — CLI commands metric**
- Current: `CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate)`
- Proposed: `CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion)`
- Context: `bonsai completion [bash|zsh|fish|powershell]` merged via PR #78 (2026-05-07). Confirmed in cmd/completion.go, Status.md "First external contribution merged", CHANGELOG v0.4.1.

**2. (Medium) station/code-index.md — completion command missing from CLI Commands table**
- Proposed row to insert after the `bonsai validate` entry:
  `| bonsai completion | cmd/completion.go:17 | completionCmd — generate shell completion script for bash/zsh/fish/powershell |`

**3. (Low) station/CLAUDE.md — plan-grilling workflow missing from Workflows nav table**
- File: `station/agent/Workflows/plan-grilling.md` (added 2026-06-13)
- Description: Adversarial review of a drafted plan via 6 critic agents (5 prose + empirical Reality), looped to convergence before dispatch.
- Proposed row: `| Grilling a Tier-2 (or non-trivial Tier-1) plan before dispatching code agents | agent/Workflows/plan-grilling.md |`

**4. (Low) station/CLAUDE.md — critic-agent-prompts skill missing from Skills nav table**
- File: `station/agent/Skills/critic-agent-prompts.md` (added 2026-06-13)
- Description: Prompt templates for the 6 plan-grilling critic agents (5 prose + Reality). Consumed verbatim by plan-grilling.md.
- Proposed row: `| Dispatching the 6 plan-grilling critic agents; running an adversarial plan review | agent/Skills/critic-agent-prompts.md |`

**5. (Info / user decision) station/INDEX.md — docs/agent-interface.md not in Document Registry**
- `docs/agent-interface.md` was shipped in Plan 41 Phase 5 (2026-06-16) as the canonical headless CLI contract for AI integrators and CI scripts.
- It is currently not listed in station/INDEX.md Document Registry.
- This may be intentional (external-facing doc, not a station doc) but the tech lead may want a pointer when reasoning about headless CLI or MCP integration (Plan 42).
- User decision: add to Document Registry or leave as-is.

## Notes for Next Run

- Last run covered commits from 2026-05-04 through 2026-08-10 (98-day gap). If findings 1–4 are resolved before next run, the next check will start from a clean baseline.
- `v0.5.0` is tagged "Unreleased" in CHANGELOG — once released, station/INDEX.md "Current phase" description may need review.
- The `docs/` directory now has 8 files including `agent-interface.md`, `formats.md` (new in Plan 40 Phase 3) — worth a full docs-dir audit in a future run if more items ship there.
- `bubbletea/` is a subdirectory under `station/agent/Skills/` with no nav entry — likely intentional (loaded via bubbletea.md), but worth confirming with the tech lead.
