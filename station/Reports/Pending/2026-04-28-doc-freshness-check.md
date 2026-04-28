---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-04-28
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-04-21
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 18 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/agent/Core/memory.md`, `/home/user/Bonsai/station/agent/Workflows/issue-to-implementation.md`, `/home/user/Bonsai/station/agent/Workflows/session-wrapup.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/.bonsai.yaml`, `/home/user/Bonsai/.claude/settings.json`, `/home/user/Bonsai/station/.claude/settings.json`, `/home/user/Bonsai/internal/wsvalidate/wsvalidate.go`, plus git log, directory listings, and catalog meta.yaml files
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, find, ls, grep, python3 link checker)
- **Errors Encountered:** 0

## Procedure Walkthrough

**Step 1 — Scan project documentation vs recent git history:**
Read `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`. Cross-referenced against `git log --since="7 days ago"`. The last 7 days of commits (2026-04-22 through 2026-04-25) show Plans 26-33 shipping — including new TUI flow packages (`catalogflow`, `guideflow`, `listflow`, `removeflow`, `updateflow`), the `wsvalidate` extract, and NoteStandards. Found 3 doc-vs-code gaps.

**Step 2 — Check INDEX.md accuracy:**
Tech stack table is accurate (Go 1.24+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, single binary). Agent types count (6) is correct. CLI commands count (7) is correct in the Key Metrics table. However, the Architecture Overview diagram lists only 5 commands (`init, add, remove, list, catalog`) — missing `update` and `guide` from the inline comment.

**Step 3 — Check navigation links:**
Used Python link checker against all markdown links in `station/CLAUDE.md`, `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/`. Results:
- `station/CLAUDE.md`: 50 links — all valid. No broken navigation links.
- Core/Protocols/Workflows/Skills: 12 "broken" reported by the scanner, but investigation shows:
  - `[plan](path)` and `[label](url)` in session-wrapup.md and memory.md — these are **example text / instructional placeholders**, not real links. Not a real breakage.
  - `[Research/RESEARCH-*.md]` links in memory.md (6 links) — these reference `station/Research/` which is gitignored (`station/Research/` in .gitignore). The directory is absent from disk. These references are known (the 2026-04-20 memory consolidation report noted the files existed); the Research/ content was likely on a different machine or is archived. The memory.md References section still documents them as active pointers.
  - `[agent/Skills/dispatch.md]` in issue-to-implementation.md (3 refs) — `dispatch.md` was never created as a file in `station/agent/Skills/`. The skill was referenced as a prerequisite and in Phases 7 and 8 of the workflow, but the file does not exist.

**Step 4 — Report findings:**
Identified 4 actionable findings (see Findings Summary below). The Research/ and dispatch.md issues require user decisions on whether to create files, update references, or accept as-is. The INDEX.md and code-index.md issues are documentation drift from recent code changes.

**Step 5 — Update dashboard:**
Dashboard row for Doc Freshness Check updated to `last_ran: 2026-04-28`, `next_due: 2026-05-05`, `status: done`.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Low | `INDEX.md` architecture diagram lists only 5 CLI commands (`init, add, remove, list, catalog`) — missing `update` and `guide` in the inline comment. Key Metrics table correctly says 7. | `station/INDEX.md` line 62 | Flagged for user — propose adding `, update, guide` to the comment |
| 2 | Low | `code-index.md` missing 5 new TUI flow packages shipped in Plans 28–31: `catalogflow/`, `guideflow/`, `listflow/`, `removeflow/`, `updateflow/`. Also missing `internal/wsvalidate/` (shipped Plan 32). | `station/code-index.md` | Flagged for user — these are documentation-only gaps, not blocking |
| 3 | Medium | `issue-to-implementation.md` references `agent/Skills/dispatch.md` three times (Prerequisites, Phase 7, Phase 8) but the file does not exist. The dispatch instructions are actually inlined in the workflow itself (Phase 8 has full prompt structure). | `station/agent/Workflows/issue-to-implementation.md` lines 35, 175, 204 | Flagged for user — options: (a) create `dispatch.md` to extract the inlined dispatch instructions, or (b) remove the three `dispatch.md` links and leave content inline |
| 4 | Low | `memory.md` References section points to 6 `station/Research/RESEARCH-*.md` files that are gitignored and not on disk. Files may exist on another machine or were archived. The pointers remain accurate as intent but are currently unresolvable on this machine. | `station/agent/Core/memory.md` lines 78–83 | Flagged for user — options: (a) annotate as "off this machine", (b) remove if Research/ content is abandoned, (c) recreate/restore from source |

## Errors & Warnings

No errors encountered.

**Note on false positives:** The Python link checker flagged `[plan](path)` and `[label](url)` in session-wrapup.md and memory.md as broken links. These are intentional example/placeholder text within instructional content (e.g., "new row format: `outcome one-liner. [plan](path) · [PR #N](url)`"). Not a real breakage.

## Items Flagged for User Review

1. **Finding #3 — dispatch.md missing (Medium):** `issue-to-implementation.md` references a `dispatch.md` skill file that doesn't exist. Recommend deciding: create the file (extracting dispatch content from Phase 8), or remove the three dead references. This affects the Prerequisites and Triage/Execute sections of the issue-to-implementation workflow.

2. **Finding #2 — code-index.md stale (Low, cosmetic):** Five new TUI flow packages and `wsvalidate` package not documented. Recommend a code-index update pass to add entries for `catalogflow/`, `guideflow/`, `listflow/`, `removeflow/`, `updateflow/`, and `internal/wsvalidate/`. Can be done in a single small patch.

3. **Finding #4 — Research/ references (Low):** Memory.md has 6 pointers to gitignored Research/ files. Worth confirming whether these are accessible elsewhere or should be annotated/removed.

4. **Finding #1 — INDEX.md diagram (Low, cosmetic):** Architecture diagram inline comment missing `update` and `guide`. One-line fix.

## Notes for Next Run

- The Research/ gitignore (`station/Research/`) and `dispatch.md` findings are likely to recur unless resolved. If unresolved, next run can skip re-investigating them (known open items).
- The code-index stale packages will remain until a code-index update pass is done — check whether those sections have been added before re-flagging.
- station/CLAUDE.md navigation links are clean and well-maintained (50/50 pass). No recurring link rot in this layer.
- settings.json: When launched from `station/`, the effective settings file is `station/.claude/settings.json` which correctly wires all 9 sensors. Root-level `.claude/settings.json` only has 5 sensors (the session-scoped sensors are in the station-level file by design).
