---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-03
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
- **Duration:** ~5 minutes
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/code-index.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/CLAUDE.md` (via system-reminder), `station/agent/Core/routines.md`, `catalog/scaffolding/manifest.yaml`, `cmd/root.go`, `cmd/completion.go`
- **Files Modified:** 3 — `station/Reports/Pending/2026-09-03-doc-freshness-check.md` (this report), `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, file existence checks)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation and compare against recent git history

Read `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`, `station/Playbook/Status.md`, and `station/Playbook/Roadmap.md`. Ran `git log --oneline --since="2026-05-04"` to capture all commits since the last run.

**Recent commits (last 7 days, 2026-08-27 to 2026-09-03):** 4 commits, all routine maintenance (backlog-hygiene and status-hygiene). No code changes — no new drift introduced in the last 7 days.

**Commits since last run (2026-05-04 to 2026-09-03):** ~47 commits. Key code changes that affect documentation:
- `2eae9d4` 2026-05-07: `feat(cmd): add explicit completion subcommand` (PR #78) — added `bonsai completion`
- Plans 40 and 41 (2026-06-13 to 2026-06-16): added project-manifest and memory scaffolding items, headless CLI contract, list --json, agent-interface.md contract doc, guide Formats page

### Step 2 — Check INDEX.md accuracy

- **Tech stack**: Correct (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, embed.FS).
- **Agent types**: 6 — correct (tech-lead, fullstack, backend, frontend, devops, security).
- **Catalog items**: Actual count is 53; INDEX says "~50" — within the approximation range, acceptable.
- **CLI commands**: INDEX says "8 (init, add, remove, list, catalog, update, guide, validate)" — **STALE**. Actual is 9: the `completion` subcommand was added 2026-05-07 (PR #78) and is not reflected.
- **Architecture Overview**: Accurate for the core pipeline. Does not mention the two new opt-in scaffolding items added in Plan 40 (`project-manifest`, `memory`), though these are minor additions and the overview isn't exhaustive.

### Step 3 — Check navigation links

Verified all 51 linked paths from `station/CLAUDE.md` (Core, Protocols, Workflows, Skills, Routines, Sensors, Playbook, Logs, Reports, code-index, INDEX, .bonsai/, external refs). **All 51 paths resolve to real files or directories — no broken links.**

### Step 4 — Report findings

See Findings Summary below.

### Step 5 — Update dashboard

Updated `station/agent/Core/routines.md` — Doc Freshness Check row: `last_ran` → 2026-09-03, `next_due` → 2026-09-10, `status` → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | INDEX.md CLI command count stale — says 8, actual is 9 (missing `bonsai completion` added PR #78 2026-05-07) | `station/INDEX.md` Key Metrics table | Flagged for user decision — not updated |
| 2 | Low | code-index.md CLI Commands table missing `bonsai completion` entry | `station/code-index.md` CLI Commands section | Flagged for user decision — not updated |
| 3 | Info | INDEX.md Architecture Overview does not mention new opt-in scaffolding items `project-manifest` and `memory` added in Plan 40 | `station/INDEX.md` Architecture Overview | Flagged for user decision — not updated (overview is not exhaustive) |
| 4 | OK | All 51 navigation links in station/CLAUDE.md resolve to real files | station/CLAUDE.md | No action needed |
| 5 | OK | Tech stack, agent types, folder structure, roadmap phases all accurate | station/INDEX.md, station/Playbook/Roadmap.md | No action needed |

---

## Proposed Updates (flagged, not applied)

### Finding #1 — INDEX.md Key Metrics table
Change:
```
| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |
```
To:
```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```

### Finding #2 — code-index.md CLI Commands table
Add a row after `bonsai validate`:
```
| `bonsai completion` | `cmd/completion.go` | `completionCmd` — shell completion for bash/zsh/fish/powershell |
```

### Finding #3 — INDEX.md Architecture Overview (optional)
Could add a note under the `catalog/ (embed.FS)` line:
```
catalog/scaffolding/  ← opt-in project infrastructure (index, playbook, logs, reports, project-manifest, memory)
```
This is low priority — the overview is intentionally high-level.

---

## Errors & Warnings

None.

## Items Flagged for User Review

- **Finding #1 and #2** (CLI count + code-index): Both are minor factual drift — recommend applying both updates in a single pass, or deferring to next `bonsai update` cycle. These are station/ docs, not generated files, so they need manual edits.
- **Finding #3**: Very low priority — only update if doing a broader INDEX.md refresh.

## Notes for Next Run

- This routine was last run 2026-05-04 — a 122-day gap. The routine-check sensor should flag this at session start. Consider whether the loop.md dispatch cadence needs adjustment if this gap recurs.
- All accumulated drift from Plans 40 and 41 (2026-05-07 to 2026-06-16) was captured in this run. If user applies the proposed updates, the next run should be clean.
