---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-20
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
- **Duration:** ~5 min
- **Files Read:** 8 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md` (header), `station/CLAUDE.md` (via system context)
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, file existence checks, directory listings)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation vs recent git history
- **Action:** Ran `git log --oneline` to review commit history; read `station/INDEX.md`, `station/Playbook/Status.md`, and `station/Playbook/Roadmap.md`.
- **Result:** Git history returned 30 commits beyond the 7-day window (last activity was Plan 41 ship; commits include Plans 40–41, v0.4.3 hotfix, and external `completion` command contribution). Since the last doc-freshness-check (2026-05-04) significant features shipped: Plan 41 headless CLI contract (5 phases, PRs #120–#125), `bonsai completion` command (PR #78 by external contributor), and `docs/agent-interface.md` contract document.
- **Issues:** Three items of drift identified (see Findings Summary).

### Step 2: Check INDEX.md accuracy
- **Action:** Compared INDEX.md tech stack, folder structure, agent type count, catalog item count, and CLI command count against actual codebase (via `ls cmd/`, `ls catalog/agents/`, item counts per catalog category, `go.mod`).
- **Result:**
  - Tech stack: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS — all confirmed).
  - Agent types: accurate (6 types: tech-lead, fullstack, backend, frontend, devops, security — confirmed).
  - Folder structure in Architecture Overview: accurate — all directories listed exist.
  - **CLI command count: STALE** — INDEX.md says "8 (init, add, remove, list, catalog, update, guide, validate)" but `cmd/completion.go` exists; actual count is 9.
  - **Catalog items count: slightly stale** — INDEX.md says "~50" but actual is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines), or 60 including scaffolding (7). Still within "~50" approximation but worth refreshing.
  - **Document Registry: incomplete** — `docs/agent-interface.md` (shipped in Plan 41) is not listed.
- **Issues:** 3 drift items found.

### Step 3: Check navigation links
- **Action:** Verified all file links in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Sensors, Routines, Bonsai Reference, External References) against actual file system.
- **Result:** All 40+ links resolve. No broken links found.
- **Issues:** None — all links intact.

### Step 4: Report findings
- **Action:** Compiled finding table below; flagged for user decision (no doc edits made per routine rule).
- **Result:** 3 findings identified, all minor to medium severity.
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for "Doc Freshness Check" — Last Ran → 2026-08-20, Next Due → 2026-08-27, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | CLI command count stale: INDEX.md says 8 commands but `bonsai completion` (PR #78, shipped ~2026-05-07) makes 9. Completion command not listed in the count or by name. | `station/INDEX.md` — Key Metrics table, "CLI commands" row | Flagged — update count to 9 and add `completion` to the list |
| 2 | Low | `code-index.md` missing `bonsai completion` command entry. The CLI Commands table covers 8 commands but omits `cmd/completion.go`. | `station/code-index.md` — CLI Commands table | Flagged — add row: `bonsai completion \| cmd/completion.go` |
| 3 | Low | Document Registry in INDEX.md doesn't reference `docs/agent-interface.md`, which was shipped as part of Plan 41 (headless CLI contract doc, main `ab202c3`). | `station/INDEX.md` — Document Registry table | Flagged — add row for `docs/agent-interface.md` |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[Medium] INDEX.md CLI count drift:** Update Key Metrics "CLI commands" from `8 (init, add, remove, list, catalog, update, guide, validate)` to `9 (init, add, remove, list, catalog, update, guide, validate, completion)`.
- **[Low] code-index.md missing completion:** Add entry for `bonsai completion` to the CLI Commands table, pointing to `cmd/completion.go`.
- **[Low] INDEX.md Document Registry gap:** Add a row for `docs/agent-interface.md` — the agent-interface contract doc shipped with Plan 41. Suggested row: `\`docs/agent-interface.md\` | Headless CLI + MCP-ready agent interface contract | When building MCP server or automation on top of Bonsai commands`.

## Notes for Next Run

- All navigation links are clean — no link-rot present.
- Tech stack, agent types, and folder structure in INDEX.md remain accurate.
- Catalog items count (~50 vs actual 53) is within the approximation range; the "~" qualifier covers this, but worth updating if the format is changed.
- No architecture changes detected in recent commits — Plan 41 was additive (headless cores + contract flags) and didn't change the existing architecture diagram.
