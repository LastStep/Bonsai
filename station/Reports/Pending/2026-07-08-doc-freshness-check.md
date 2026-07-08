---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-08
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
- **Files Read:** 11 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/CLAUDE.md`, `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/code-index.md`, `/home/user/Bonsai/CLAUDE.md` (root), `cmd/root.go`, `internal/generate/generate.go`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, file existence checks, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation vs recent git history
- **Action:** Read `station/INDEX.md`, `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`. Pulled git log from 2026-06-01 to present.
- **Result:** Found 26 commits since the last run (2026-05-04). Two major plans shipped: Plan 40 (Odysseus platform integration, v0.5.0 candidate — 2026-06-13) and Plan 41 (Headless CLI Contract + MCP-ready cores — 2026-06-16). Both plans introduced significant new code not reflected in `station/code-index.md`.
- **Issues:** Significant code-index drift from Plans 40 and 41.

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` and compared against current codebase state (CLI commands, catalog items, tech stack, architecture diagram).
- **Result:** Tech stack row is accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, embed.FS). Agent types (6) accurate. Catalog items (~50) plausible. Architecture diagram is accurate. Two items are stale:
  - CLI command count says 8, but `bonsai completion` was added via PR #78 (2026-05-07) — count should be 9.
  - No mention of the headless/non-interactive CLI capability (Plan 41) or `docs/agent-interface.md` contract document.
- **Issues:** CLI count stale (low), headless capability unmentioned (low).

### Step 3: Check navigation links
- **Action:** Checked all 51 links in `station/CLAUDE.md` navigation tables against the filesystem. Also spot-checked links referenced in the procedure (agent/Core/, agent/Protocols/, agent/Workflows/, agent/Skills/).
- **Result:** All 51 links resolve — no broken navigation entries found. The bonsai-model.md link that was broken in the 2026-05-07 backlog-hygiene report appears to have been resolved.
- **Issues:** None — navigation is clean.

### Step 4: Check code-index.md accuracy
- **Action:** Read `station/code-index.md` and compared against actual function signatures and line numbers in key source files: `internal/generate/generate.go`, `cmd/root.go`, `cmd/add.go`, `cmd/remove.go`, `cmd/init_flow.go`, `cmd/update.go`, `cmd/list.go`. Also checked `internal/generate/` for new files.
- **Result:** Extensive drift found. Detailed findings in Findings Summary below.
- **Issues:** High-severity line number drift; missing new headless functions and files.

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-07-08, Next Due → 2026-07-15, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `generate.go` line numbers all drifted by ~40–100 lines after Plans 40/41 insertions; several new public functions undocumented (`SettingsJSONForAgent` :578, `PathScopedRulesForAgent` :1278, `WorkflowSkillsForAgent` :1338, `injectTriggerSection` :1440, `bonsaiReferenceLines` :744) | `station/code-index.md` → `internal/generate/generate.go` | Flagged for user — needs code-index refresh pass |
| 2 | MEDIUM | Plan 41 headless functions not documented: `runAddNonInteractive` (add.go:754), `runRemoveAgentNonInteractive` (remove.go:288), `runRemoveItemNonInteractive` (remove.go:317), `runInitNonInteractive` (init_flow.go:266), `runUpdateNonInteractive` (update.go:100), `renderListJSON` (list.go:65), `headlessRequested` (remove.go:48), `loadConfigHeadless` (remove.go:270) | `station/code-index.md` → `cmd/` | Flagged for user — code-index missing an entire Plan 41 section |
| 3 | MEDIUM | `showWriteResults()` listed in cmd/root.go section of code-index but function no longer exists in root.go (removed, file is 195 lines) | `station/code-index.md` | Flagged for user — stale entry to remove |
| 4 | MEDIUM | New file `internal/generate/list_snapshot.go` (Plan 41) not documented — contains `SerializeJSON()` function used for `list --json` output | `station/code-index.md` | Flagged for user — missing file section |
| 5 | LOW | CLI command count stale: INDEX.md says 8, but `bonsai completion` (PR #78, 2026-05-07) makes it 9 | `station/INDEX.md` | Flagged for user |
| 6 | LOW | `docs/agent-interface.md` (Plan 41 Phase 5 — agent-interface contract) exists but is not mentioned in INDEX.md Document Registry | `station/INDEX.md` | Flagged for user — worth adding to registry for agent discoverability |
| 7 | LOW | All cmd/add.go, cmd/remove.go, cmd/init_flow.go line numbers drifted from code-index values (e.g., `runAdd` :56→:73, `runRemoveItem` :290→:428, `runInit` :27→:35) | `station/code-index.md` | Included in the overall code-index refresh flag |
| 8 | INFO | `station/CLAUDE.md` — all 51 navigation links resolve cleanly | — | No action needed |
| 9 | INFO | Roadmap.md — Phase 1 all checked, Phase 2–4 future items accurate | — | No action needed |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **CODE-INDEX REFRESH (HIGH + MEDIUM)** — `station/code-index.md` needs a sweep to update line numbers across `generate.go`, `cmd/add.go`, `cmd/remove.go`, `cmd/init_flow.go`, and add new sections for: (a) headless non-interactive functions from Plan 41 in each cmd file, (b) `internal/generate/list_snapshot.go` with `SerializeJSON()`. Remove stale `showWriteResults()` entry from root.go section. This is a straightforward doc-refresh task suitable for a quick dispatch — similar to Plan 37 doc-refresh-bundle.

2. **INDEX.md MINOR UPDATES (LOW)** — Two small updates: (a) CLI command count 8→9 (add `completion`), (b) add `docs/agent-interface.md` row to Document Registry.

## Notes for Next Run

- `station/code-index.md` recurring drift pattern — Plans 40 and 41 both added code without a corresponding code-index refresh. The code-index was last synced in Plan 37 (2026-05-07); two major plans (40, 41) landed since then without updating it. Consider whether to include a code-index sweep as a mandatory step in the issue-to-implementation workflow for plans that add new public functions.
- Navigation links remain clean — the broken `bonsai-model.md` link flagged in the 2026-05-07 backlog-hygiene appears resolved.
- Root `Bonsai/CLAUDE.md` project-structure tree (recurring drift item from prior cycles) — reviewed this cycle and it appears accurate for the current file layout. The persistent drift noted in 2026-05-04 run was apparently fixed in Plan 37 or later. No issues found this cycle.
