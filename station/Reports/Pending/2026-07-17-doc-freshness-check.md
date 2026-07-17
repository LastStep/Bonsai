---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-17
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
- **Duration:** ~6 min
- **Files Read:** 13
  - `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`
  - `/home/user/Bonsai/station/CLAUDE.md`
  - `/home/user/Bonsai/station/INDEX.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/agent/Core/memory.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/code-index.md`
  - `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md` (header only)
  - `/home/user/Bonsai/station/agent/Core/identity.md` (link check)
  - `/home/user/Bonsai/station/agent/Protocols/session-start.md` (link check)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/station/Reports/Pending/` (directory listing)
- **Files Modified:** 0 — doc updates are flagged for user decision per routine procedure
- **Tools Used:** `git log --oneline --since="30 days ago"`, `git ls-files station/Research/`, `find` for RESEARCH-*.md, `ls` on all agent subdirectories and key paths
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation vs recent git history

Ran `git log --oneline --since="7 days ago"` — found 2 commits (both from the backlog-hygiene routine run today):
- `9162c4c routine(backlog-hygiene): report, dashboard update, log entry`
- `8b23771 routine(backlog-hygiene): mark 3 resolved P0/P1 items as done`

No feature/config changes in the last 7 days. Extended scope to 30 days — same 2 commits only. Last substantive feature commit predates the 30-day window (Plan 41 shipped 2026-06-16, ~31 days ago). Checked Plan 41 changes against docs — found drift (Finding 6 below).

### Step 2 — Check INDEX.md accuracy

Read `station/INDEX.md`. Tech stack table and project description are accurate. Architecture diagram is accurate for the core flow.

**Found drift:**
- Key Metrics row says "CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate)" — the `completion` command (added by @mvanhorn, merged 2026-05-07) is missing. Count should be 9.
- INDEX.md does not reflect Plan 41 additions: headless `*Result` cores, `list --json`, `ExitConflict=5`, and `docs/agent-interface.md` contract. These are significant architectural additions.

### Step 3 — Check navigation links

Verified all agent directory contents via `ls`. Cross-referenced against links in `station/CLAUDE.md` navigation tables. All explicitly listed files resolve to real files.

**Found missing entries (files exist but not in navigation tables):**
- `agent/Workflows/plan-grilling.md` — exists, not in Workflows table in station/CLAUDE.md. File meta header notes "full Bonsai-catalog integration pending (Backlog)".
- `agent/Skills/critic-agent-prompts.md` — exists, not in Skills table. Used by the plan-grilling workflow.

**Found broken links:**
- `station/agent/Core/memory.md` References section (lines 87–92) — 6 links point to `../../Research/RESEARCH-*.md`. The `station/Research/` directory does not exist (confirmed: not in working tree and not in git history). Prior 2026-05-07 memory-consolidation report claimed these files existed — they have since been removed or were local-only and never committed.

**Found outstanding housekeeping item:**
- `station/Playbook/Plans/Active/41-headless-cli-contract.md` — memory.md explicitly notes "Plan 41 file still in Plans/Active/ — archive to Plans/Archive/ at next wrap-up." Plan shipped 2026-06-16, all 5 phases merged. File has not been archived.

### Step 4 — Report findings

7 findings catalogued below. All doc updates flagged for user decision — not executed per routine procedure.

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-07-17, Next Due → 2026-07-24, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `memory.md` References section has 6 broken links to `station/Research/` which does not exist | `station/agent/Core/memory.md` lines 87–92 | Flagged — user decision: remove entries or restore Research/ directory |
| 2 | Medium | `plan-grilling.md` workflow is not in the station/CLAUDE.md Workflows navigation table | `station/CLAUDE.md` Workflows table | Flagged — propose adding row: trigger "grill the plan / critic pass / review plan NN" → file |
| 3 | Low-Medium | INDEX.md architecture section does not reflect Plan 41 headless CLI additions (headless cores, `list --json`, `ExitConflict=5`, `docs/agent-interface.md`) | `station/INDEX.md` Architecture section | Flagged — propose updating architecture description + key metrics |
| 4 | Low | INDEX.md CLI command count says 8, should be 9 (`completion` command missing) | `station/INDEX.md` Key Metrics table | Flagged — propose updating to "9 (init, add, remove, list, catalog, update, guide, validate, completion)" |
| 5 | Low | `code-index.md` CLI Commands table does not list `bonsai completion` | `station/code-index.md` CLI Commands section | Flagged — propose adding row for `completion` command and `cmd/completion.go` |
| 6 | Low | `critic-agent-prompts.md` skill is not in station/CLAUDE.md Skills navigation table | `station/CLAUDE.md` Skills table | Flagged — decide whether to add navigation entry |
| 7 | Low | Plan 41 (`41-headless-cli-contract.md`) still in Plans/Active/ — should be archived per memory.md note | `station/Playbook/Plans/Active/41-headless-cli-contract.md` | Flagged — propose moving to Plans/Archive/ |

---

## Errors & Warnings

No errors encountered during execution.

---

## Items Flagged for User Review

- **[BROKEN LINKS] memory.md Research/ references (6 links)** — `station/Research/` does not exist. Options: (a) remove the entire References section (or just the 6 dead entries), (b) recreate the Research/ directory and restore the files if they exist elsewhere (GitHub history, another machine). Prior report claimed they existed 2026-05-07; they are now gone from the working tree and git index.

- **[NAVIGATION GAP] plan-grilling workflow not in CLAUDE.md** — The plan-grilling workflow and its associated `critic-agent-prompts.md` skill are installed but not routed via the navigation table. The workflow meta says "Bonsai-catalog integration pending (Backlog)" — this was likely a deliberate deferral. Decide: add navigation entries now, or leave as-is (agent can still be pointed there explicitly).

- **[DOC DRIFT] INDEX.md and code-index.md not updated for Plan 41 or `completion`** — Three concrete updates needed: (1) CLI count 8 → 9 in INDEX.md; (2) Add `completion` row in code-index.md; (3) Add headless-CLI note to INDEX.md architecture section. The backlog-hygiene report from today (2026-07-17) also flagged "INDEX.md arch drift" as an uncaptured finding — these are the same items.

- **[HOUSEKEEPING] Plan 41 not archived** — `41-headless-cli-contract.md` is still in Plans/Active/. Move to Plans/Archive/ when convenient.

---

## Notes for Next Run

- The Research/ broken-link situation should be resolved before the next run — otherwise it will show up again. If the files are truly gone, remove the References section entries.
- Consider running `bonsai validate` after any doc sweep that touches workspace structure (per memory.md gotcha).
- The gap between this run (2026-07-17) and last run (2026-05-04) is 74 days — significantly overdue. The dashboard showed the routine as "done" but with a 2026-05-11 next-due date. Multiple plans shipped in the interim without triggering a doc check.
- Cross-reference with today's backlog-hygiene report: it also flagged code-index staleness and INDEX.md arch drift. These findings overlap — coordinate so they are added to the backlog once, not twice.
