---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-24
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (previous value from dashboard)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 minutes
- **Files Read:** 16 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/CLAUDE.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/docs/agent-interface.md`, `/home/user/Bonsai/cmd/remove.go`, `/home/user/Bonsai/cmd/add.go`, `/home/user/Bonsai/cmd/update.go`, `/home/user/Bonsai/cmd/list.go`, `/home/user/Bonsai/cmd/init_flow.go`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, grep, ls), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan Project Documentation
- **Action:** Read `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, and ran `git log --since="2026-05-04"` to capture all commits since last doc-freshness run.
- **Result:** Found 48 commits since 2026-05-04. Major changes: Plan 40 (Odysseus — freeze schemas + root-relative scaffolding, v0.5.0 Phases 1–3), Plan 41 (Headless CLI Contract — `internal/nonint/` package, `--non-interactive` flags on all mutating commands, `list --json`, `docs/agent-interface.md`). Also: `bonsai completion` command (Plan 37 era, PR #78 `2eae9d4`).
- **Issues:** Multiple documentation artifacts have not been updated to reflect Plans 40 and 41.

### Step 2: Check INDEX.md Accuracy
- **Action:** Compared `station/INDEX.md` tech stack, key metrics, and architecture block against actual codebase.
- **Result:** INDEX.md is **accurate** — Tech Stack (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS) all correct. Agent types count (6) correct. Catalog items (~50) reasonable. CLI commands (8) correct — `bonsai completion` is a subcommand, not a top-level command, so the count of 8 is still right. Architecture diagram names (`internal/nonint` is new but not necessarily needed in the high-level diagram).
- **Issues:** None in INDEX.md itself.

### Step 3: Check Navigation Links
- **Action:** Verified every link in `station/CLAUDE.md` nav tables (Core, Protocols, Workflows, Skills, Routines, Sensors, Bonsai Reference, External References) against actual files on disk.
- **Result:** All 42 links in the nav tables resolve to real files. Zero broken links.
- **Issues:** Two files exist in agent workspace directories but are **missing from the CLAUDE.md nav tables** — see Finding #1.

### Step 4: Report Findings
- **Action:** Compiled all drift findings below.
- **Result:** 4 findings across 3 documents. Two are flag-for-user (nav table gaps, CLAUDE.md root tree drift). Two are informational (code-index.md line-number drift, code-index.md missing sections).
- **Issues:** See Findings Summary.

### Step 5: Update Dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for Doc Freshness Check: Last Ran → 2026-06-24, Next Due → 2026-07-01, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | **Medium** | `plan-grilling.md` and `critic-agent-prompts.md` exist in `agent/Workflows/` and `agent/Skills/` respectively but are not in `station/CLAUDE.md` nav tables. These were added in commit `6995d4f` (2026-06-13) as custom (non-catalog) workflow+skill. The agent cannot route to them from CLAUDE.md. | `station/CLAUDE.md` Workflows and Skills nav tables | Flagged for user — pending catalog integration decision (Backlog P2 line 74). Recommend adding interim nav entries so the agent can load them while full catalog integration is built. |
| 2 | **Medium** | `station/code-index.md` has significant line-number drift across `cmd/` and `internal/generate/generate.go` — all functions shifted +17–145 lines due to Plan 41 headless additions. Example: `runRemoveItem()` documented at `:290`, actual `:428`; `runRemoveItemAction()` at `:565`, actual `:703`; `Scaffolding()` at `:360`, actual `:401`; `AgentWorkspace()` at `:1359`, actual `:1460`. | `station/code-index.md` — all `cmd/` and `generate.go` sections | Flagged for user — entire code-index.md needs a refresh sweep. Propose as a Plan 42 sub-task or standalone micro-plan. |
| 3 | **Medium** | `station/code-index.md` is missing the `internal/nonint/` package section (Plan 41, 15 files), the `completion.go` CLI command entry, and the `docs/agent-interface.md` contract document. These represent significant new surface area added since the last code-index sweep. | `station/code-index.md` | Flagged for user — add sections covering `internal/nonint/` (EmitJSONL, EmitFile, EmitSummary, LoadConfig, LoadOverlay, Result shape) and `completion.go`, plus a note on `docs/agent-interface.md`. |
| 4 | **Low** | Root `Bonsai/CLAUDE.md` project structure tree is missing `completion.go` in `cmd/` listing and `internal/nonint/` in `internal/` listing. These were both added after the last tree update. | `/home/user/Bonsai/CLAUDE.md` Project Structure block | Flagged for user — matches existing Backlog P2 item (line 133: "Add root Bonsai/CLAUDE.md tree-drift check to doc-freshness-check routine"). |

---

## Errors & Warnings

No errors encountered. The `showWriteResults()` function referenced in code-index.md at `:201` could not be found in `cmd/root.go` — it appears to have been removed or refactored. This is a sub-finding under Finding #2 (line-number drift).

---

## Items Flagged for User Review

1. **Finding #1 — CLAUDE.md nav gaps for plan-grilling.md + critic-agent-prompts.md:** Recommend adding these two rows to the nav tables now (interim measure) while the full catalog integration tracks in Backlog. Low-effort edit. The agent can't auto-route to these files without nav entries.

2. **Finding #2 + #3 — code-index.md refresh:** All `cmd/` function line numbers are stale (Plan 41 shifted everything). Missing `internal/nonint/` section and `completion.go`. Recommend a dedicated code-index sweep — either as part of Plan 42 kickoff or as a standalone micro-task. The code-index is a navigation aid; stale line numbers actively mislead the agent.

3. **Finding #4 — Root CLAUDE.md tree drift:** Minor; consistent with existing Backlog item. Can batch with the code-index sweep.

---

## Notes for Next Run

- The Backlog item (P2 line 133) requesting that `doc-freshness-check` add a sub-step to diff the root CLAUDE.md tree block is still open. If that sub-step were built in, Finding #4 would be caught automatically each cycle.
- Plans 40 and 41 are the two biggest sources of drift this cycle. After code-index is refreshed, the next cycles should be clean unless Plan 42 (MCP server) ships.
- `showWriteResults()` in root.go appears removed — this should be confirmed and the code-index entry deleted.
