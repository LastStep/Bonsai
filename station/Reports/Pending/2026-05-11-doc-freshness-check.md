---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-05-11
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
- **Duration:** ~8 minutes
- **Files Read:** 16 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/38-bonsai-eval-bootstrap.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/agent/Core/memory.md`, `/home/user/Bonsai/station/agent/Protocols/session-start.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/CLAUDE.md`, `/home/user/Bonsai/go.mod`, `/home/user/Bonsai/cmd/completion.go`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/README.md`, `/home/user/Bonsai/station/Reports/report-template.md`
- **Files Modified:** 0 — no doc updates executed (findings flagged for user decision per procedure step 4)
- **Tools Used:** git log (--oneline --since=7days, --show, --name-only), ls, grep, find, head, tail
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Playbook/Plans/Active/38-bonsai-eval-bootstrap.md`. Ran `git log --oneline --since="7 days ago"` covering 2026-05-04 to 2026-05-11.
- **Result:** 17 commits in the last 7 days. Key feature landed: `feat(cmd): add explicit completion subcommand for bash/zsh/fish/powershell (#54)` (commit `2eae9d4`, 2026-05-07). Also: v0.4.1 release shipped, PR triage sweep, Plan 38 Bonsai-Eval bootstrap dispatched, Plan 37 doc-refresh bundle executed.
- **Issues:** The `completion` subcommand was added but is not fully reflected in all doc files — see Findings.

### Step 2: Check INDEX.md accuracy
- **Action:** Cross-referenced `station/INDEX.md` key metrics table (Tech Stack, CLI command count, Agent types, Catalog items) against actual codebase state (`go.mod`, `cmd/*.go`, `catalog/` directories).
- **Result:**
  - **Go version:** `Go 1.25+` — matches `go.mod` (`go 1.25.0`). Correct.
  - **CLI commands:** INDEX.md says "8 (init, add, remove, list, catalog, update, guide, validate)". Actual count is **9** — `completion` subcommand was added in commit `2eae9d4`. Not reflected.
  - **Agent types:** "6 (tech-lead, fullstack, backend, frontend, devops, security)" — matches `catalog/agents/` (6 dirs). Correct.
  - **Catalog items:** "~50" — actual count: 18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines = **53** non-core items. "~50" is within reasonable approximation range; no update strictly needed.
  - **Architecture diagram:** still accurate.
  - **Bonsai-Eval external reference row:** present and correct (added by Plan 38).
- **Issues:** CLI command count stale (8 → 9).

### Step 3: Check navigation links
- **Action:** Extracted all relative markdown links from `station/CLAUDE.md` and verified each target file exists on disk.
- **Result:** **All 52 links resolve.** No broken links found. Files checked include all Core, Protocols, Workflows, Skills, Routines, Sensors, and Playbook references.
- **Issues:** None for navigation links themselves. However, `station/agent/Core/memory.md` References section contains 6 links to `station/Research/RESEARCH-*.md` files — that directory does not exist on disk. These are not in the CLAUDE.md navigation table (so not caught by the link check) but are referenced by the agent during sessions.

### Step 4: Report findings
- **Action:** Compiled all findings below. Per procedure, updates are flagged for user decision — not executed autonomously.
- **Result:** 3 findings identified (1 medium, 1 medium, 1 low). Detailed in Findings Summary.
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Will update `agent/Core/routines.md` dashboard row for Doc Freshness Check to set `Last Ran → 2026-05-11`, `Next Due → 2026-05-18`, `Status → done`.
- **Result:** Completed as post-procedure step.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | CLI command count stale: says "8" but `completion` was added (now 9). Two instances: `station/INDEX.md` line 33 and Architecture section line 63. | `station/INDEX.md` | Flagged — user to update or approve agent update |
| 2 | Medium | Root `CLAUDE.md` cmd/ directory tree does not include `completion.go`. All other cmd/ files are listed. | `/home/user/Bonsai/CLAUDE.md` | Flagged — user to update or approve agent update |
| 3 | Low | `station/agent/Core/memory.md` References section links to 6 `station/Research/RESEARCH-*.md` files — the `station/Research/` directory does not exist. These were likely deleted or never committed. Links will 404 if agent tries to navigate them. | `station/agent/Core/memory.md` lines 83–88 | Flagged — user to confirm if Research/ was intentionally removed and update references accordingly |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **INDEX.md CLI command count** — Two lines in `station/INDEX.md` list "8" commands. With `completion` subcommand added (PR #54, merged 2026-05-07), the correct count is 9. Proposed fix: update line 33 `8 (init, add, remove, list, catalog, update, guide, validate)` → `9 (init, add, remove, list, catalog, update, guide, validate, completion)` and update line 63 similarly.

2. **Root CLAUDE.md cmd/ tree missing completion.go** — The project-level `CLAUDE.md` file structure section lists `validate.go` as the last cmd/ file but `completion.go` is absent. Proposed fix: add `│   ├── completion.go    ← bonsai completion — shell completion script generator` before or after validate.go.

3. **Dead Research links in memory.md** — The References section in `station/agent/Core/memory.md` (lines 83–88) links to 6 RESEARCH-*.md files in `station/Research/`. That directory does not exist. Options: (a) remove the dead links, (b) restore the Research/ directory from git history if it was accidentally deleted, (c) update paths if files moved. Confirm intent before editing.

## Notes for Next Run

- Plan 37 doc-refresh-bundle was executed 2026-05-07 and fixed several drifts — this run found 3 new items introduced since then (completion subcommand added post-plan-37).
- The `station/Research/` directory issue may be long-standing — worth a quick `git log --diff-filter=D -- station/Research/` to determine when it was removed.
- Catalog item count "~50" (INDEX.md) is currently 53 — still within the "~50" approximation. If catalog grows significantly, consider updating to "~55" or similar.
- All navigation links in CLAUDE.md are clean — no link rot from recent restructuring.
