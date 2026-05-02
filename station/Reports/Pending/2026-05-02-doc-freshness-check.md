---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-05-02
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
- **Duration:** ~10 minutes
- **Files Read:** 18 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/CLAUDE.md`, `station/agent/Core/memory.md`, `station/code-index.md`, `station/Playbook/Backlog.md`, `station/agent/Workflows/issue-to-implementation.md`, `station/agent/Workflows/session-wrapup.md`, `station/agent/Protocols/memory.md`, `station/agent/Protocols/security.md`, `station/agent/Protocols/session-start.md`, `station/Playbook/Plans/Archive/32-followup-bundle.md`, `station/Logs/RoutineLog.md`, `internal/config/config.go`, `internal/wsvalidate/wsvalidate.go`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** git log, find, ls, grep, bash file existence checks
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Ran `git log --oneline --format="%ai %s" -15` to identify recent commits since last check (2026-04-21). Also reviewed `station/Playbook/Status.md` for shipped plans.
- **Result:** Most recent commits are from 2026-04-25 (7 days before today). Key changes: Plan 32 shipped (wsvalidate extract, Validate() chokepoint, O_NOFOLLOW snapshot hardening). NoteStandards added to `Playbook/Standards/`. No commits in the last 7 days from today (2026-05-02).
- **Issues:** Three new artifacts from the 2026-04-25 session are not reflected in documentation: `internal/wsvalidate/` package, `ProjectConfig.Validate()` chokepoint in `internal/config/config.go`, and `Playbook/Standards/NoteStandards.md`. See findings below.

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` in full. Verified tech stack, CLI commands (7: init, add, remove, list, catalog, update, guide), agent types (6: tech-lead, fullstack, backend, frontend, devops, security), catalog items ("~50"). Checked architecture overview diagram.
- **Result:** Tech stack, CLI commands, agent type count, and catalog item count are accurate. The architecture overview diagram at line 63–75 lists `internal/catalog/`, `internal/config/`, `internal/generate/`, `internal/tui/` — but does NOT include `internal/wsvalidate/` which was added by Plan 32. The Document Registry at line 39–53 does NOT include `Playbook/Standards/NoteStandards.md` (added 2026-04-25).
- **Issues:** Two stale omissions — see Findings 1 and 2.

### Step 3: Check navigation links
- **Action:** Extracted all markdown links from `station/CLAUDE.md` and batch-verified all 39 file targets exist on disk. Checked links in Protocol files, Workflow files, and `station/agent/Core/memory.md`.
- **Result:**
  - **CLAUDE.md (39 links):** All 39 resolve. All Core, Protocol, Workflow, Skill, Routine, and Sensor links are valid.
  - **Protocol files:** Links in `memory.md`, `security.md`, and `session-start.md` resolve. `memory.md` links to `../../Playbook/Standards/NoteStandards.md` — file exists, OK.
  - **Workflow files:** `issue-to-implementation.md` links to `../Skills/dispatch.md` (3 occurrences) — file does NOT exist. `session-wrapup.md` has `[plan](path)` and `[PR #N](url)` which are template placeholder text in the brevity rule example, not real links — acceptable.
  - **memory.md References section:** Contains 6 links to `../../Research/RESEARCH-*.md` files — the `station/Research/` directory does not exist and these files cannot be found anywhere in the repository or filesystem.
- **Issues:** Two broken link sets — see Findings 3 and 4.

### Step 4: Additional checks
- **Action:** Checked `station/code-index.md` for reflection of Plan 32 changes. Checked `station/agent/Sensors/` for any unlisted sensor files.
- **Result:** `code-index.md` does not list `wsvalidate` package or `ProjectConfig.Validate()` function added by Plan 32. `statusline.sh` exists in `station/agent/Sensors/` but is not in the CLAUDE.md sensors table — however, it is wired as a `statusLine` provider (not a hook) in `station/.claude/settings.json`, so its absence from the hooks table is correct (different mechanism).
- **Issues:** `code-index.md` drift from Plan 32 — see Finding 5.

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for Doc Freshness Check.
- **Result:** Last Ran set to 2026-05-02, Next Due to 2026-05-09, Status to done.
- **Issues:** none.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Low | `internal/wsvalidate/` package added by Plan 32 not listed in INDEX.md architecture overview | `station/INDEX.md` lines 63–75 | Flagged for user — doc update needed |
| 2 | Low | `Playbook/Standards/NoteStandards.md` added 2026-04-25 missing from INDEX.md Document Registry | `station/INDEX.md` lines 39–53 | Flagged for user — doc update needed |
| 3 | Medium | `agent/Skills/dispatch.md` referenced in `issue-to-implementation.md` (3 occurrences) does not exist — dispatch skill is in catalog but not installed | `station/agent/Workflows/issue-to-implementation.md` lines 35, 175, 204 | Flagged for user — install dispatch skill or remove references |
| 4 | Medium | 6 `station/Research/RESEARCH-*.md` files linked in `memory.md` References section do not exist — directory `station/Research/` is absent from disk and git history | `station/agent/Core/memory.md` lines 78–83 | Flagged for user — files may need to be recovered or links removed |
| 5 | Low | `code-index.md` does not reflect Plan 32 additions: `internal/wsvalidate/` package (`InvalidReason`, `Normalise`) and `ProjectConfig.Validate()` in `internal/config/config.go` | `station/code-index.md` lines 127–138 | Flagged for user — code-index update needed |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **Finding 3 (Medium) — Broken dispatch skill link:** `issue-to-implementation.md` references `agent/Skills/dispatch.md` which does not exist. The `dispatch` skill exists in `catalog/skills/dispatch/` but is not installed in `.bonsai.yaml`. Decision needed: (a) `bonsai add` to install the dispatch skill, or (b) update the workflow to inline dispatch guidance or remove the cross-reference. The workflow is currently functional without the skill file — it degrades gracefully at runtime since Claude Code handles missing file reads — but the dead link is misleading.

- **Finding 4 (Medium) — Missing Research files:** `station/agent/Core/memory.md` References section links to 6 `RESEARCH-*.md` files at `station/Research/`. Neither the directory nor any of the files exist anywhere on disk or in git history. The 2026-04-25 memory-consolidation run erroneously reported these as existing. Decision needed: (a) recover files from another source if they exist elsewhere, or (b) remove the broken references from memory.md — the research is likely embedded in design decisions elsewhere.

- **Finding 1 & 2 (Low) — INDEX.md stale omissions:** Two additions from 2026-04-25 not reflected: `internal/wsvalidate/` package in the architecture overview, and `Playbook/Standards/NoteStandards.md` in the Document Registry. Routine doc update — no urgency.

- **Finding 5 (Low) — code-index.md drift:** Plan 32 added `wsvalidate.InvalidReason()`, `wsvalidate.Normalise()`, and `ProjectConfig.Validate()` but `code-index.md` was not updated. Routine code-index update — no urgency.

## Notes for Next Run

- The Research files finding (Finding 4) is critical context: if the next run also sees these links in memory.md, they remain broken unless the user has resolved them.
- The dispatch skill installation status should be checked — if still not installed by next run, the broken link in `issue-to-implementation.md` persists.
- No commits in the 7 days since 2026-04-25. If development resumes, the next doc-freshness check should cross-reference all new plans/PRs against INDEX.md and code-index.md.
