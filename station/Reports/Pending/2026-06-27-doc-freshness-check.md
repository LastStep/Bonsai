---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-27
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
- **Duration:** ~10 min
- **Files Read:** 10 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/CLAUDE.md`, `/home/user/Bonsai/station/agent/Core/self-awareness.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/CLAUDE.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation + compare against recent git history
- **Action:** Read `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`, root `Bonsai/CLAUDE.md`; ran `git log --oneline --since="54 days ago"` to capture all commits since last run (2026-05-04).
- **Result:** 54 commits since last run spanning Plans 39, 40, 41, v0.4.1–v0.4.3 releases, PR triage, and two other routines today. Key code changes: Plan 39 added `internal/nonint/` package; Plan 41 added `list_snapshot.go`, `catalog_snapshot_unix.go`, `catalog_snapshot_windows.go` in `internal/generate/`; PR #78 added `cmd/completion.go`. Multiple docs do not reflect these changes.
- **Issues:** 4 doc drift items found (see Findings Summary).

### Step 2: Check INDEX.md accuracy
- **Action:** Verified tech stack, folder structure, CLI count, catalog count, and agent type count in `station/INDEX.md`.
- **Result:**
  - **Tech stack:** accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea — all correct).
  - **Agent types:** 6 — accurate.
  - **Catalog items:** `~50` listed; actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). Within range — `~50` acceptable as approximate.
  - **CLI commands:** `8 (init, add, remove, list, catalog, update, guide, validate)` — **low-severity drift**: `completion.go` (PR #78) adds an explicit `bonsai completion` subcommand making it 9. However the completion command was always latent via Cobra's auto-generated command; this is cosmetic.
  - **Architecture diagram:** missing `internal/nonint/` package — **high-severity drift** since nonint is a significant new internal package (Plans 39+41, headless CLI contract).
- **Issues:** 2 items (nonint missing from architecture diagram — HIGH; CLI count — LOW).

### Step 3: Check navigation links
- **Action:** Verified all links in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills) and linked Bonsai reference files.
- **Result:** All 17 navigation links verified as resolving to real files:
  - Core (3/3): identity.md, memory.md, self-awareness.md — all exist.
  - Protocols (4/4): memory.md, scope-boundaries.md, security.md, session-start.md — all exist.
  - Workflows (9/9): code-review.md, planning.md, pr-review.md, security-audit.md, session-logging.md, test-plan.md, session-wrapup.md, issue-to-implementation.md, routine-digest.md — all exist.
  - Skills (6/6): bonsai-model.md, planning-template.md, review-checklist.md, issue-classification.md, pr-creation.md, bubbletea.md — all exist.
  - Bonsai reference links (`.bonsai/catalog.json`, `.bonsai.yaml`) — both exist.
  - Previous finding (broken `bonsai-model.md` link) — **RESOLVED**: file exists at the correct path.
- **Issues:** none — all links clean.

### Step 4: Report findings + flag for user decision
- **Action:** Compiled drift items; all items are doc-only drift (no code changes needed). Following routine procedure: flag for user decision, do not execute doc updates autonomously.
- **Result:** 4 drift findings identified (see Findings Summary). Root `Bonsai/CLAUDE.md` tree drift is a recurring pattern (4th cycle). `internal/nonint/` is the largest new undocumented item from recent plans.
- **Issues:** none in execution; 4 findings to report.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard — Doc Freshness Check row: Last Ran → 2026-06-27, Next Due → 2026-07-04, Status → done.
- **Result:** Dashboard updated.
- **Issues:** none.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package (Plans 39+41, headless CLI contract) entirely absent from root `Bonsai/CLAUDE.md` project-structure tree, `station/INDEX.md` architecture diagram, and `station/code-index.md` | `Bonsai/CLAUDE.md`, `station/INDEX.md`, `station/code-index.md` | Flagged for user — doc update needed |
| 2 | MEDIUM | `internal/generate/list_snapshot.go` (`ListSnapshot`, `SerializeJSON` — Plan 41 Phase 4) absent from root `Bonsai/CLAUDE.md` generate/ tree and `station/code-index.md` | `Bonsai/CLAUDE.md`, `station/code-index.md` | Flagged for user |
| 3 | MEDIUM | `internal/generate/catalog_snapshot_unix.go` and `catalog_snapshot_windows.go` (OS-split openSnapshotFile — v0.4.0 hotfix) absent from root `Bonsai/CLAUDE.md` generate/ tree and `station/code-index.md` | `Bonsai/CLAUDE.md`, `station/code-index.md` | Flagged for user |
| 4 | LOW | `cmd/completion.go` (explicit `bonsai completion` command — PR #78) absent from root `Bonsai/CLAUDE.md` cmd/ tree and `station/code-index.md`; `station/INDEX.md` CLI count lists 8 commands (should be 9) | `Bonsai/CLAUDE.md`, `station/code-index.md`, `station/INDEX.md` | Flagged for user |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**All 4 findings are documentation-only updates** — no code changes required. Per routine procedure, proposing updates but not executing.

**Finding 1 (HIGH) — `internal/nonint/` undocumented everywhere:**

The `internal/nonint/` package was added by Plan 39 (non-interactive flags) and significantly expanded by Plan 41 (headless CLI contract). It now contains 11 files (config.go, config_test.go, contract_test.go, events.go, events_test.go, nonint.go, remove.go, remove_test.go, result.go, result_test.go, runner.go, runner_test.go, update.go, update_test.go) and is the foundation of the headless MCP-ready interface.

Proposed updates:
- Root `Bonsai/CLAUDE.md` project-structure tree — add `nonint/` under `internal/` with description `← headless CLI contract — Result, JSONL events, RunInit/RunAdd/RunRemove/RunUpdate orchestrators (Plans 39+41)`
- `station/INDEX.md` architecture diagram — add `internal/nonint/  ← headless command cores + JSONL event contract (Plan 41)` in the diagram block
- `station/code-index.md` — add a new `## Nonint (internal/nonint/)` section documenting the key types (Result, Counts, RunInit, RunAdd, RunRemove, RunUpdate, EmitJSONL, exit codes)

Note: this is a recurring pattern — the Backlog P2 item `[improvement] Add root Bonsai/CLAUDE.md tree-drift check to doc-freshness-check routine` (Backlog line 133) was promoted from P3→P2 after the 3rd cycle. This is the 4th cycle of the same class of drift. **Consider prioritizing that P2 item** to prevent future cycles from accumulating silently.

**Findings 2+3 (MEDIUM) — new generate/ files undocumented:**

- `list_snapshot.go`: defines `ListSnapshot` (stable JSON shape for `bonsai list --json`), `ListAgent`, `SerializeJSON`. Proposed code-index addition: under Generator section → `### list_snapshot.go — Headless List Serialization` with `ListSnapshot / ListAgent`, `SerializeJSON`.
- `catalog_snapshot_unix.go` + `catalog_snapshot_windows.go`: OS-split implementation of `openSnapshotFile` (uses `O_NOFOLLOW` on unix, plain create on windows — the v0.4.0 Windows cross-compile hotfix). Proposed code-index addition: note in the catalog_snapshot.go section that the OS-split is via `_unix.go`/`_windows.go` build tags.

**Finding 4 (LOW) — completion command and CLI count:**

- Root `Bonsai/CLAUDE.md` cmd/ tree: add `├── completion.go    ← bonsai completion [bash|zsh|fish|powershell]` entry.
- `station/code-index.md` CLI Commands table: add row `bonsai completion | cmd/completion.go | completionCmd`.
- `station/INDEX.md` Key Metrics: update `8 (init, add, remove, list, catalog, update, guide, validate)` → `9 (init, add, remove, list, catalog, update, guide, validate, completion)`.

## Notes for Next Run

1. **Recurring drift class:** Root `Bonsai/CLAUDE.md` project-structure tree continues to drift with each plan that adds new files. The Backlog P2 item (line 133) to add a tree-diff sub-step to this routine procedure template is the correct fix — recommend the user prioritize it or fold it into the next doc-refresh plan.
2. **Navigation links are clean:** All 17 station/CLAUDE.md nav links resolve — this is a good signal that the workspace navigation layer is healthy. No action needed there.
3. **bonsai-model.md link resolved:** The broken nav link flagged in the 2026-05-04 cycle is confirmed resolved (file exists).
4. **Plans 40+41 in Active/:** Still present in `Plans/Active/` despite being in Status.md Recently Done. This was already flagged by status-hygiene today — carry-forward, not a new finding here.
5. **INDEX.md catalog count (~50):** Actual is 53. The approximation is still reasonable and not worth updating every cycle — only update when next plan adds substantial new abilities.
