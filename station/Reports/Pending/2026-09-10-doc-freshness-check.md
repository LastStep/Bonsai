---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-10
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (previous value from dashboard, before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~5 min
- **Files Read:** 9 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/CLAUDE.md`, `/home/user/Bonsai/cmd/completion.go`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history

- **Action:** Read station/INDEX.md, station/Playbook/Roadmap.md, station/Playbook/Status.md. Ran `git log --since="7 days ago"` and `git log --since="60 days ago"` to identify recent commits.
- **Result:** Only 2 commits in the last 7 days — both routine maintenance runs (backlog-hygiene and status-hygiene on 2026-09-10). Looking at longer history (since 2026-05-04), `bonsai completion` was shipped in commit `2eae9d4` on 2026-05-07 via PR #78 (external contribution). This change is NOT reflected in several documentation files.
- **Issues:** Documentation drift detected — see Findings Summary.

### Step 2: Check INDEX.md accuracy

- **Action:** Read station/INDEX.md. Verified tech stack table, key metrics, folder structure, and architecture diagram against actual codebase (ls catalog/agents/, ls cmd/*.go, ls catalog/skills/ etc.).
- **Result:**
  - Tech stack: Correct (Go, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template).
  - Agent types: Correct — 6 types listed and 6 exist in catalog/agents/.
  - Catalog items: INDEX.md says "~50" — actual total is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). Approximation is acceptable.
  - **CLI commands: STALE** — INDEX.md says "8 (init, add, remove, list, catalog, update, guide, validate)" but there are 9 CLI commands. `bonsai completion` (shipped 2026-05-07) is missing from both the Key Metrics table (line 33) and the Architecture Overview command list (line 63).
  - Folder structure: Accurate. No new top-level directories introduced.
- **Issues:** CLI command count and list stale in two places in INDEX.md.

### Step 3: Check navigation links

- **Action:** Verified all 53 file paths linked from station/CLAUDE.md navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References sections).
- **Result:** All 53 links resolve to real files — 100% pass rate. No broken navigation links found.
- **Secondary finding:** station/code-index.md CLI Commands table lists only 8 commands (missing `bonsai completion`). Root CLAUDE.md project structure listing (cmd/ directory tree) also omits `completion.go`.
- **Issues:** Two additional files have the same `completion` command omission.

### Step 4: Report findings

- **Action:** Compiled 3 concrete drift findings (all variants of the same root cause — `completion` command not reflected post-ship).
- **Result:** Findings documented below. No doc updates applied — flagged for user decision per routine rules.
- **Issues:** None.

### Step 5: Update dashboard

- **Action:** Update routines.md dashboard row for Doc Freshness Check.
- **Result:** Updated Last Ran → 2026-09-10, Next Due → 2026-09-17, Status → done.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `bonsai completion` missing from CLI commands count and list in Key Metrics table | `station/INDEX.md:33` | Flagged — propose changing "8 (init, add, remove, list, catalog, update, guide, validate)" → "9 (init, add, remove, list, catalog, update, guide, validate, completion)" |
| 2 | Low | `bonsai completion` missing from Architecture Overview command list | `station/INDEX.md:63` | Flagged — propose adding `completion` to the inline list in the architecture diagram |
| 3 | Low | `bonsai completion` missing from CLI Commands table | `station/code-index.md` (CLI Commands table) | Flagged — propose adding a row: `bonsai completion \| cmd/completion.go:20 \| completionCmd — generate shell completion script (bash/zsh/fish/powershell)` |
| 4 | Low | `completion.go` missing from cmd/ directory tree listing | `CLAUDE.md` (project root, cmd/ structure block ~line 36) | Flagged — propose inserting `│   ├── completion.go       ← bonsai completion — shell completion (bash/zsh/fish/powershell)` after `validate.go` line |

All findings share the same root cause: `bonsai completion` was shipped 2026-05-07 (PR #78, commit `2eae9d4`, external contribution from @mvanhorn) but was not followed up with documentation updates in INDEX.md, code-index.md, or root CLAUDE.md.

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

**4 stale doc locations** (all low-effort text edits, same root cause):

1. `station/INDEX.md:33` — CLI command count + list (Medium)
2. `station/INDEX.md:63` — Architecture diagram command list (Low)
3. `station/code-index.md` — CLI Commands table missing `completion` row (Low)
4. `CLAUDE.md` (project root) — `cmd/` directory tree missing `completion.go` line (Low)

All 4 can be fixed in a single pass. No architectural decision needed — this is a documentation gap, not a content question.

---

## Notes for Next Run

- All navigation links are currently clean — no broken paths. If new workflows/skills are added between now and the next run, the link check will catch any missing file.
- The `~50 catalog items` approximation in INDEX.md is still defensible at 53 items, but may want to update to `~55` if more items are added this cycle.
- Next run due: 2026-09-17.
