---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-16
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
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/CLAUDE.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/code-index.md`, `go.mod`, `cmd/completion.go`, `cmd/remove.go`, `cmd/add.go`, `station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `station/Reports/Pending/2026-07-16-doc-freshness-check.md` (created), `station/agent/Core/routines.md` (dashboard row), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, grep, ls)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`; ran `git log --since="2026-05-04"` to capture all commits since last run.
- **Result:** 50+ commits since 2026-05-04, spanning Plans 37–41. Plans 39, 40, and 41 introduced significant new code (headless CLI flags, new generate.go per-agent helpers, completion command, list --json). None of these changes were reflected in `code-index.md` or `station/CLAUDE.md`.
- **Issues:** Multiple stale references and missing entries — detailed in findings below.

### Step 2: Check INDEX.md accuracy
- **Action:** Compared INDEX.md tech stack, folder structure, and key metrics against `go.mod`, `cmd/` directory, and catalog item counts.
- **Result:** Tech stack is accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, embed.FS). Folder structure matches. Agent count (6) is correct. Catalog item count (~50) is approximate but reasonable at 53 actual items. One metric is stale: CLI commands listed as "8" — `completion` makes it 9.
- **Issues:** CLI commands count/list in INDEX.md omits `bonsai completion` (added 2026-04-28, commit `2aef7fd`).

### Step 3: Check navigation links
- **Action:** Verified all file paths linked from `station/CLAUDE.md` nav tables (Core, Protocols, Workflows, Skills, Sensors, Routines sections) against actual file system; also checked for files present in workspace directories but absent from nav.
- **Result:** All linked files resolve to real paths. No broken links found. However, two files added since May 2026 are absent from the nav tables: `agent/Workflows/plan-grilling.md` and `agent/Skills/critic-agent-prompts.md`.
- **Issues:** Two unlisted files identified (see Findings 3 and 4).

### Step 4: Validate code-index.md against source
- **Action:** Spot-checked line numbers in `code-index.md` against live source using `grep -n` on `cmd/add.go`, `cmd/remove.go`, `cmd/list.go`, `cmd/update.go`, `cmd/catalog.go`, `cmd/init_flow.go`, and `internal/generate/generate.go`.
- **Result:** Every checked line number is stale. Plans 39, 40, and 41 inserted significant blocks of code throughout, shifting all subsequent functions by 17–138 lines depending on the file. Additionally, several new functions added by these plans are not listed in code-index at all.
- **Issues:** Pervasive line-number drift across all major source files (see Findings 5 and 6).

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for Doc Freshness Check.
- **Result:** Last Ran set to 2026-07-16, Next Due set to 2026-07-23, Status set to done.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | INDEX.md lists CLI commands as 8 (init/add/remove/list/catalog/update/guide/validate) — `bonsai completion` (cmd/completion.go, added commit 2aef7fd) is missing; count should be 9 | `station/INDEX.md` — Key Metrics table | Flagged for user |
| 2 | Low | INDEX.md architecture diagram omits `completion` command from the `cmd/` annotation | `station/INDEX.md` — Architecture Overview | Flagged for user |
| 3 | Medium | `agent/Workflows/plan-grilling.md` (added ~June 2026, actively used for adversarial plan review) is absent from the Workflows nav table in `station/CLAUDE.md` | `station/CLAUDE.md` — Workflows section | Flagged for user |
| 4 | Low | `agent/Skills/critic-agent-prompts.md` (companion to plan-grilling, added ~June 2026) is absent from the Skills nav table in `station/CLAUDE.md` | `station/CLAUDE.md` — Skills section | Flagged for user |
| 5 | High | `code-index.md` line numbers for `cmd/` functions are stale throughout. Plans 39/40/41 shifted: `add.go` functions by +17–35 lines; `remove.go` functions by +138 lines; `list.go` `runList()` from :18 → :39; `update.go` `runUpdate()` from :19 → :51; `catalog.go` `runCatalog()` from :23 → :39; `init_flow.go` `runInit()` from :27 → :35 | `station/code-index.md` — CLI Commands section | Flagged for user |
| 6 | High | `code-index.md` line numbers for `internal/generate/generate.go` are stale. All exported functions shifted by +41 to +101 lines (e.g. `Scaffolding()` from :360 → :401; `AgentWorkspace()` from :1359 → :1460; `WorkspaceClaudeMD()` from :725 → :826) | `station/code-index.md` — Generator section | Flagged for user |
| 7 | Medium | `code-index.md` is missing Plan 41 headless functions: `runRemoveItemNonInteractive()` (remove.go:317), `runUpdateNonInteractive()` (update.go:100), `renderListJSON()` (list.go:65) | `station/code-index.md` — CLI Commands section | Flagged for user |
| 8 | Low | `code-index.md` missing per-agent generate helpers added by Plan 40/41: `SettingsJSONForAgent()` (generate.go:578), `PathScopedRulesForAgent()` (generate.go:1278), `WorkflowSkillsForAgent()` (generate.go:1338) | `station/code-index.md` — Generator section | Flagged for user |
| 9 | Low | `bonsai completion` subcommand absent from `bonsai add` cinematic flow code-index entry (minor cosmetic note) | `station/code-index.md` | Flagged for user |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

All 9 findings are flagged. None can be resolved autonomously (procedure is audit-only — no doc edits). Recommended priority order:

1. **Findings 5 + 6 (High) — Refresh code-index.md line numbers.** The code-index is the developer agent's primary navigation tool for Go source. With all major function references off by 17–138 lines, it actively misleads. Recommend a full refresh pass using `grep -n` on each source file. This is mechanical work suitable for a code agent.

2. **Findings 3 + 7 (Medium) — Add plan-grilling workflow and headless functions to nav/index.** The plan-grilling workflow is actively used but invisible in the CLAUDE.md nav, meaning a new session agent won't find it via table lookup.

3. **Findings 1 + 2 (Medium) — Update INDEX.md CLI commands count and diagram to include `completion`.**

4. **Findings 4 + 8 (Low) — Add critic-agent-prompts to Skills nav; add per-agent generate helpers to code-index.**

---

## Notes for Next Run

- code-index.md refresh is overdue and accumulating drift with each plan. Consider making code-index refresh a step in the plan closeout workflow rather than catching it in doc-freshness.
- All nav links in station/CLAUDE.md resolve correctly — link rot is not currently a problem.
- INDEX.md tech stack and folder structure are accurate; only the CLI commands count needs updating.
- If `bonsai guide` Formats page (Plan 40 Phase 3) added a new guide topic, the code-index `guideflow/` section may also need a row for it — not checked in this run.
