---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-01
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (value from dashboard before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~5 minutes
- **Files Read:** 11
  - `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`
  - `/home/user/Bonsai/station/INDEX.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/code-index.md`
  - `/home/user/Bonsai/CLAUDE.md` (via system prompt)
  - `/home/user/Bonsai/station/CLAUDE.md` (via system prompt)
  - `/home/user/Bonsai/station/agent/Core/memory.md` (links extracted)
  - `/home/user/Bonsai/station/agent/Protocols/` (links extracted)
  - `/home/user/Bonsai/station/agent/Workflows/issue-to-implementation.md` (links extracted)
- **Files Modified:** 2
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, file existence checks, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan Project Documentation vs. Recent Git History
- **Action:** Ran `git log --since="7 days ago"` and `git log --since="30 days ago"` to identify recent commits. Read `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, and `station/code-index.md`.
- **Result:** Only 4 commits in the last 7 days (all today, 2026-08-01), all routine maintenance (memory-consolidation, status-hygiene, backlog-hygiene). No new features or services shipped in the last 7 days. The main code changes that caused doc drift occurred earlier: Plan 41 (headless CLI, June 2026) added `internal/nonint/`, the `bonsai completion` command was merged in May 2026. Both remain undocumented in station docs.
- **Issues:** Doc drift from Plan 41 and the completion command addition is still outstanding (flagged below).

### Step 2: Check INDEX.md Accuracy
- **Action:** Compared `station/INDEX.md` tech stack table, key metrics, and architecture overview against actual codebase structure.
- **Result:** Tech stack table is accurate. Architecture overview text is accurate. **Key Metrics row for "CLI commands" is stale**: INDEX.md says "8 (init, add, remove, list, catalog, update, guide, validate)" but the actual count is **9** — `bonsai completion` (shell completions for bash/zsh/fish/powershell) was added in commit `2eae9d4`. The `internal/nonint/` package (added in Plan 41, commit `410a5f1`) is not reflected anywhere in station docs.
- **Issues:** 2 drift items (see Findings table).

### Step 3: Check Navigation Links
- **Action:** Extracted and resolved all relative Markdown links from: `station/CLAUDE.md` (all sections), `agent/Core/*.md`, `agent/Protocols/*.md`, `agent/Workflows/*.md`, `agent/Skills/*.md`.
- **Result:**
  - **station/CLAUDE.md**: All 52 links verified — all resolve to real files.
  - **agent/Core/memory.md**: 6 Research file links are broken (`station/Research/RESEARCH-*.md`) — already noted as stale in memory.md from the 2026-08-01 memory-consolidation run. Not a new finding.
  - **agent/Protocols/**: All links resolve.
  - **agent/Workflows/issue-to-implementation.md**: **3 broken links** to `agent/Skills/dispatch.md` — file does not exist. The `dispatch` catalog skill (`catalog/skills/dispatch/dispatch.md`) exists but is not installed for the tech-lead agent. The workflow file assumes it is.
  - **agent/Skills/**: No outbound relative links.
- **Issues:** 1 broken link finding (dispatch.md, 3 occurrences in issue-to-implementation.md).

### Step 4: Report Findings
- **Action:** Compiled all drift and broken-link findings.
- **Result:** 6 findings total — 1 medium (broken link in workflow), 5 low (stale counts/missing entries across INDEX.md, code-index.md, and root CLAUDE.md). All flagged for user decision — no autonomous edits to doc content made.
- **Issues:** None.

### Step 5: Update Dashboard
- **Action:** Set `Last Ran` → 2026-08-01, `Next Due` → 2026-08-08, `Status` → `done` in `agent/Core/routines.md`.
- **Result:** Dashboard updated.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `agent/Skills/dispatch.md` linked 3× in issue-to-implementation.md but file does not exist (dispatch skill not installed for tech-lead) | `station/agent/Workflows/issue-to-implementation.md` lines 35, 175, 204 | Flagged for user — options: install dispatch skill (`bonsai add`), or update workflow to not reference an uninstalled skill |
| 2 | Low | CLI command count stale: INDEX.md says "8" but actual count is 9 (`bonsai completion` added commit `2eae9d4`, May 2026) | `station/INDEX.md` Key Metrics table | Flagged for user — proposed fix: update to "9 (init, add, remove, list, catalog, update, guide, validate, completion)" |
| 3 | Low | `bonsai completion` command missing from CLI Commands table | `station/code-index.md` | Flagged for user — proposed fix: add row "bonsai completion \| cmd/completion.go \| completionCmd → shell completion scripts for bash/zsh/fish/powershell" |
| 4 | Low | `internal/nonint/` package (Plan 41 headless CLI contract) missing from code index entirely | `station/code-index.md` | Flagged for user — proposed fix: add new section documenting the nonint package (config, events, runner, result, remove, update sub-packages) |
| 5 | Low | `completion.go` missing from cmd/ directory tree listing | `CLAUDE.md` (project root) | Flagged for user — proposed fix: add `├── completion.go ← bonsai completion (shell completions)` to cmd/ tree |
| 6 | Low | `internal/nonint/` missing from internal/ directory tree listing | `CLAUDE.md` (project root) | Flagged for user — proposed fix: add `├── nonint/ ← headless CLI contract (Plan 41) — runner, result, events, config` to internal/ tree |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

### Finding 1 (Medium): Broken dispatch.md links in issue-to-implementation workflow
`station/agent/Workflows/issue-to-implementation.md` references `agent/Skills/dispatch.md` at 3 locations (lines 35, 175, 204). This file does not exist — the `dispatch` skill is in the catalog (`catalog/skills/dispatch/dispatch.md`) but not installed for the tech-lead agent.

**Options for user:**
- **Option A:** Install the dispatch skill via `bonsai add` (select tech-lead → add "dispatch" skill). This would create the file and make links valid.
- **Option B:** Update `issue-to-implementation.md` to reference the dispatch concept inline or via a different existing skill, since dispatch-guard sensor is already installed.

### Findings 2–6 (Low): Code doc drift — completion command and nonint package undocumented
Three doc files (`station/INDEX.md`, `station/code-index.md`, `CLAUDE.md`) have not been updated to reflect:
- The `bonsai completion` subcommand (added May 2026, commit `2eae9d4`)
- The `internal/nonint/` headless CLI package (added Plan 41, commit `410a5f1`)

These are low-risk but affect agent navigation accuracy. If the user approves, these are all small targeted edits (5 rows/lines across 3 files). A single agent pass could apply all 5 fixes.

---

## Notes for Next Run
- Research links in `agent/Core/memory.md` are already marked stale — no need to re-flag.
- If the dispatch skill is installed (Option A above), verify `agent/Skills/dispatch.md` is generated and the links in issue-to-implementation.md become valid.
- Findings 2–6 will reappear as drift until the user approves the doc updates.
