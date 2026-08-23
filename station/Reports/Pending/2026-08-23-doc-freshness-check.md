---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-23
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
- **Duration:** ~6 min
- **Files Read:** 14 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md` (header only); plus glob/ls scans of `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/`, `agent/Routines/`, `agent/Sensors/`, `Playbook/`, `Logs/`, `Reports/`, `cmd/`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Glob, Bash (git log, ls)
- **Errors Encountered:** 0

## Procedure Walkthrough

**Step 1 — Scan project documentation vs recent git history:**
Read station/INDEX.md, station/CLAUDE.md, and station/Playbook/Status.md. Retrieved git log (last 25 commits). Key commits since last run (2026-05-04): Plan 40 Phases 1–3 (frozen schemas, root-relative scaffolding, validate pass, memory-routing docs), Plan 41 (headless CLI contract across all mutating commands, list --json, exit-code contract). The `bonsai completion` command was added via PR #78 (2026-05-07), one day after the last routine run. One new workflow file (`plan-grilling.md`) and one new skill file (`critic-agent-prompts.md`) appeared on disk without corresponding navigation entries in CLAUDE.md.

**Step 2 — Check INDEX.md accuracy:**
Tech stack table (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template) is accurate. Architecture diagram references 8 CLI commands but `cmd/completion.go` now exists. Key Metrics row "CLI commands: 8 (init, add, remove, list, catalog, update, guide, validate)" is stale — there are 9 commands. All other INDEX.md content (agent types, catalog count, architecture flow) appears accurate.

**Step 3 — Check navigation links:**
All links in station/CLAUDE.md were verified against the filesystem:
- Core files: all 3 present (identity.md, memory.md, self-awareness.md) — OK
- Protocols: all 4 present — OK
- Workflows listed in nav: all 9 present — OK; but `plan-grilling.md` exists on disk and is NOT listed
- Skills listed in nav: all 6 present — OK; but `critic-agent-prompts.md` exists on disk and is NOT listed
- Routines: all 7 present — OK
- Sensors: all 10 present — OK
- External References (Playbook/, Logs/, Reports/, code-index.md, .bonsai.yaml, .bonsai/catalog.json): all targets exist — OK

**Step 4 — Report findings:**
4 findings flagged below. None executed autonomously (procedure: flag for user decision).

**Step 5 — Update dashboard:**
Done — `routines.md` Last Ran → 2026-08-23, Next Due → 2026-08-30, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `CMD_COUNT_STALE` — CLI command count says 8; `cmd/completion.go` exists (PR #78, 2026-05-07) making it 9. Both INDEX.md Key Metrics row and Architecture diagram list need updating. | `station/INDEX.md` | Flagged for user — proposed fix: update count to 9, add `completion` to the cmd list in Key Metrics and Architecture diagram. |
| 2 | Low | `WORKFLOW_MISSING_FROM_NAV` — `agent/Workflows/plan-grilling.md` exists (imported 2026-06-13 from ZenGarden) but is absent from the Workflows navigation table in station/CLAUDE.md. | `station/CLAUDE.md` (Workflows table) | Flagged for user — proposed fix: add a row for plan-grilling.md with trigger "Running adversarial grill review on a drafted plan". |
| 3 | Low | `SKILL_MISSING_FROM_NAV` — `agent/Skills/critic-agent-prompts.md` exists but is not listed in the Skills navigation table in station/CLAUDE.md. | `station/CLAUDE.md` (Skills table) | Flagged for user — confirm intentional omission (supporting reference, not a trigger-based skill) or add a row. |
| 4 | Low | `PLANS_NOT_ARCHIVED` — Plans 40 and 41 are in `Playbook/Plans/Active/` but both are marked done in Status.md. (Also flagged by Backlog Hygiene routine today.) | `station/Playbook/Plans/Active/` | Flagged for user — move to Plans/Archive/ when convenient. |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[Medium] INDEX.md CLI command count is stale (8 vs 9)** — `bonsai completion` added PR #78. Update Key Metrics row and Architecture diagram in `station/INDEX.md`. Low effort fix.
2. **[Low] plan-grilling.md missing from CLAUDE.md Workflows nav** — File exists and has proper frontmatter. Add a row with activation condition.
3. **[Low] critic-agent-prompts.md missing from CLAUDE.md Skills nav** — Confirm intentional omission or add a navigation row.
4. **[Low] Plans 40 and 41 in Active/ despite being done** — Archive them when convenient (also in Backlog Hygiene report from today).

## Notes for Next Run

- All navigation link targets verified clean — no broken links as of this run.
- Plans/Active/ archival (finding #4) is a recurring flag — if still not done by next run, escalate to P1 in the Backlog.
- The `bubbletea/` subdirectory under Skills has 4 sub-files (components.md, golden-rules.md, troubleshooting.md, emoji-width-fix.md). They aren't individually listed in CLAUDE.md nav — this appears intentional (main bubbletea.md links them). Verified no drift there.
- Plan 42 (MCP server fast-follow) was mentioned in Status.md as "next" after Plan 41 — watch for new cmd/internal files when it ships.
