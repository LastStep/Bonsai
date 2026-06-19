---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-19
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
- **Files Read:** 9 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/CLAUDE.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/internal/generate/list_snapshot.go`
- **Files Modified:** 3 — `station/Reports/Pending/2026-06-19-doc-freshness-check.md` (this report), `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls, grep, find)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan Project Documentation / Compare Against Recent Git History
- **Action:** Ran `git log --oneline --since="7 days ago"` and `--since="2026-05-04"` (since last doc freshness run 46 days ago). Reviewed Status.md for recent shipped plans.
- **Result:** Since the last doc freshness check (2026-05-04), the following major plans shipped: Plan 40 (Odysseus platform integration — frozen schemas, root-relative scaffolding, validate pass, memory-routing docs, `docs/formats.md`), Plan 41 (Headless CLI Contract — `internal/nonint/` package, `bonsai list --json`, `docs/agent-interface.md`, PRs #120/#122/#123/#121/#125). New external contribution: `bonsai completion` command (PR #78, `cmd/completion.go`).
- **Issues:** 8 drift items found — see Findings Summary.

### Step 2: Check INDEX.md Accuracy
- **Action:** Read `station/INDEX.md` and compared tech stack, folder structure, and metrics against codebase reality.
- **Result:** Tech stack section accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template). Agent types still 6 (verified `ls catalog/agents/`). CLI commands count says **8** but `bonsai completion` is now a public command (`cmd/completion.go`, registered via `rootCmd.AddCommand`), making it **9**. Catalog items count `~50` — actual count is 53 `meta.yaml` files (within range, acceptable). Architecture diagram does not reference `internal/nonint/` package or `docs/agent-interface.md`.
- **Issues:** (1) CLI commands 8→9; (2) `docs/agent-interface.md` and `docs/formats.md` not in document registry.

### Step 3: Check Navigation Links
- **Action:** Verified all links in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors). Verified existence of each linked file.
- **Result:** All links resolve — 0 broken nav links. However, two files exist in agent subdirs that are NOT linked in nav tables: `agent/Workflows/plan-grilling.md` and `agent/Skills/critic-agent-prompts.md`. Both are custom files added during Plan 40/session work (plan-grilling pipeline, `6995d4f`).
- **Issues:** (1) `plan-grilling.md` missing from Workflows nav; (2) `critic-agent-prompts.md` missing from Skills nav.

### Step 4: Check root CLAUDE.md and code-index.md for drift
- **Action:** Read root `Bonsai/CLAUDE.md` project structure tree and `station/code-index.md`. Checked against actual `ls` output of `internal/` and `cmd/`.
- **Result:**
  - **Root CLAUDE.md project structure tree:** Missing `internal/nonint/` package entirely (14 files — Plan 41's headless CLI core). Missing `internal/generate/list_snapshot.go` (Plan 41 Phase 4). Missing `internal/generate/catalog_snapshot_unix.go` and `catalog_snapshot_windows.go` (v0.4.0 hotfix split). `cmd/completion.go` not listed.
  - **code-index.md:** Has no section for `internal/nonint/` package. `list_snapshot.go` not covered. Both are significant omissions for a developer navigating Plan 42 (MCP server) work.
- **Issues:** HIGH severity — `internal/nonint/` and `list_snapshot.go` undocumented in both dev docs.

### Step 5: Update Dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for "Doc Freshness Check" — Last Ran → 2026-06-19, Next Due → 2026-06-26, Status → done.
- **Result:** Dashboard updated.
- **Issues:** none

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package (Plan 41 — 14 files: config.go, events.go, nonint.go, remove.go, result.go, runner.go, update.go + tests) missing from root `Bonsai/CLAUDE.md` project structure tree AND `station/code-index.md` | `/home/user/Bonsai/CLAUDE.md` (internal/ section), `/home/user/Bonsai/station/code-index.md` | Flagged for user — doc update needed |
| 2 | HIGH | `internal/generate/list_snapshot.go` (Plan 41 Phase 4 — `ListSnapshot` JSON shape for `bonsai list --json`) missing from root `Bonsai/CLAUDE.md` tree AND `code-index.md` Generate section | `/home/user/Bonsai/CLAUDE.md` (generate/ item list), `/home/user/Bonsai/station/code-index.md` | Flagged for user — doc update needed |
| 3 | MEDIUM | `internal/generate/catalog_snapshot_unix.go` + `catalog_snapshot_windows.go` (v0.4.0 hotfix split) absent from root `Bonsai/CLAUDE.md` project structure tree (only `catalog_snapshot.go` listed) | `/home/user/Bonsai/CLAUDE.md` | Flagged for user |
| 4 | MEDIUM | CLI commands count `8` in `station/INDEX.md` — `bonsai completion` (PR #78, `cmd/completion.go`) makes it **9** | `/home/user/Bonsai/station/INDEX.md` line 33 | Flagged for user |
| 5 | MEDIUM | `station/CLAUDE.md` Workflows nav table missing `plan-grilling.md` — file exists at `agent/Workflows/plan-grilling.md`, added during Plan 40 session (`6995d4f`) | `/home/user/Bonsai/station/CLAUDE.md` Workflows section | Flagged for user |
| 6 | LOW | `station/CLAUDE.md` Skills nav table missing `critic-agent-prompts.md` — file exists at `agent/Skills/critic-agent-prompts.md`, added as companion to plan-grilling workflow | `/home/user/Bonsai/station/CLAUDE.md` Skills section | Flagged for user |
| 7 | LOW | `docs/agent-interface.md` (Plan 41 headless contract doc) and `docs/formats.md` (Plan 40 Phase 3) not referenced in `station/INDEX.md` document registry | `/home/user/Bonsai/station/INDEX.md` Document Registry section | Flagged for user |
| 8 | INFO | `cmd/completion.go` not listed in root `Bonsai/CLAUDE.md` cmd/ section of project structure tree | `/home/user/Bonsai/CLAUDE.md` | Flagged for user (low priority — completion is a support command) |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

All 8 findings require user decision — this routine is audit-only per procedure (Step 4: "Propose updates (but don't execute — flag for user decision)").

**Priority order:**
1. (HIGH) Add `internal/nonint/` section to `station/code-index.md` — this is actively needed for Plan 42 (MCP server) navigation. Suggest adding after the Validate section.
2. (HIGH) Add `list_snapshot.go` entry to `code-index.md` Generator section — needed for Plan 42.
3. (MEDIUM) Update root `Bonsai/CLAUDE.md` project structure tree: add `internal/nonint/`, `list_snapshot.go`, `catalog_snapshot_unix.go`, `catalog_snapshot_windows.go`, `cmd/completion.go`.
4. (MEDIUM) Update `station/INDEX.md` CLI commands count: 8→9.
5. (MEDIUM) Add `plan-grilling.md` row to Workflows nav in `station/CLAUDE.md`.
6. (LOW) Add `critic-agent-prompts.md` row to Skills nav in `station/CLAUDE.md`.
7. (LOW) Consider adding `docs/agent-interface.md` to `station/INDEX.md` document registry.

## Notes for Next Run

- Root `CLAUDE.md` project structure tree is the recurring high-drift item — Plan 41 extended `internal/` significantly without doc update. Suggest making root CLAUDE.md update a checklist item in the plan template (or adding it as a doc-refresh step at plan close-out).
- `internal/nonint/` is a pure-headless MCP-ready package — Plan 42 will need it well-documented. Prioritize #1 and #2 before Plan 42 kickoff.
- Routine ran 46 days late (last run 2026-05-04, should have run 2026-05-11). 8 findings vs 5 in the prior cycle. Drift accumulates faster with active plan work.
- All nav links in `station/CLAUDE.md` resolve cleanly — no broken links this cycle.
