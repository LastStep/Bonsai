---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-18
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
- **Files Read:** 9 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/CLAUDE.md`, `station/agent/Core/routines.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`, `station/Logs/RoutineLog.md`, `go.mod`, directory listings (agent/Core/, Protocols/, Workflows/, Skills/, Sensors/, catalog/, cmd/)
- **Files Modified:** 0 (audit-only routine)
- **Tools Used:** Read, Bash (git log, ls)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation

Reviewed `station/INDEX.md`, `station/Playbook/Roadmap.md`, and `station/code-index.md` against git history from 2026-05-04 to 2026-08-18 (106 days, 63 commits). Key feature commits in the period:

- **2026-05-07 (PR #54):** `feat(cmd): add explicit completion subcommand for bash/zsh/fish/powershell`
- **2026-05-13 (Plan 39 / v0.4.2):** `feat(nonint): bonsai init/add --non-interactive --from-config`
- **2026-06-13 (Plan 40 / v0.5.0):** freeze schemas + root-relative scaffolding, project-level validate pass
- **2026-06-16 (Plan 41):** headless CLI contract — `--yes`/`--from` on `remove`, `--non-interactive`/`--skip-conflicts` on `update`, `--json` on `list`; MCP-ready core architecture

### Step 2 — Check INDEX.md accuracy

- Tech stack: All entries (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template) verified accurate against `go.mod` and `cmd/`.
- Agent types: Confirmed 6 types in `catalog/agents/` (tech-lead, fullstack, backend, frontend, devops, security) — matches INDEX.md.
- **CLI commands count: STALE.** INDEX.md lists "8 (init, add, remove, list, catalog, update, guide, validate)" but `cmd/completion.go` ships a 9th `completion` subcommand (added 2026-05-07). Count and list are both wrong.
- **Catalog item count: SLIGHTLY STALE.** INDEX.md says "~50" but actual count is 53 (skills=18, workflows=10, protocols=4, sensors=13, routines=8). The approximation is still within tolerance but directionally low.
- **New CLI flags not mentioned anywhere in INDEX.md:** `--non-interactive`, `--from-config` (init/add), `--yes`, `--from` (remove), `--non-interactive`, `--skip-conflicts` (update), `--json` (list). These represent a significant headless/MCP usage mode.

### Step 3 — Check navigation links

**station/CLAUDE.md:**
- Core links: All 4 files (identity.md, memory.md, routines.md, self-awareness.md) verified to exist. ✓
- Protocol links: All 4 files verified to exist. ✓
- Workflow links: 9 files listed — all verified to exist. ✓
  - **Unlisted workflow:** `agent/Workflows/plan-grilling.md` exists on disk but is absent from the CLAUDE.md Workflows table.
- Skills links: 6 files listed — all verified to exist. ✓
  - **Unlisted skill:** `agent/Skills/critic-agent-prompts.md` exists on disk but is absent from the CLAUDE.md Skills table.
- Sensor links: All 10 sensors listed — all verified to exist. ✓
- Routine links: All 7 routines listed — all verified to exist. ✓
- External reference links: `station/INDEX.md`, `Playbook/Status.md`, `Playbook/Roadmap.md`, `Playbook/Standards/SecurityStandards.md`, `Playbook/Plans/Active/`, `Playbook/Backlog.md`, `Logs/KeyDecisionLog.md`, `Reports/Pending/` — all checked and exist. ✓

**code-index.md CLI Commands table:**
- Lists 9 commands (init, add, remove, list, catalog, update, guide, validate) but `completion` command is also missing here. Count is listed as 8 implicitly.

### Step 4 — Report findings

See Findings Summary table below.

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` dashboard row for Doc Freshness Check.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | INDEX.md CLI commands count says 8 but there are 9 — `completion` subcommand (added 2026-05-07 via PR #54) is missing from the list | `station/INDEX.md` — Key Metrics table | Flagged for user review |
| 2 | Medium | `agent/Workflows/plan-grilling.md` exists on disk but is not listed in the CLAUDE.md Workflows navigation table | `station/CLAUDE.md` — Workflows table | Flagged for user review |
| 3 | Medium | `agent/Skills/critic-agent-prompts.md` exists on disk but is not listed in the CLAUDE.md Skills navigation table | `station/CLAUDE.md` — Skills table | Flagged for user review |
| 4 | Medium | Plan 41 headless CLI flags (--yes/--from on remove, --non-interactive/--skip-conflicts on update, --json on list) are not reflected anywhere in INDEX.md — significant usage mode omission | `station/INDEX.md` | Flagged for user review |
| 5 | Low | `40-odysseus-platform-integration.md` remains in `station/Playbook/Plans/Active/` but commits from 2026-06-13 indicate Phases 1-3 were shipped and closed out | `station/Playbook/Plans/Active/` | Flagged for user review |
| 6 | Low | code-index.md CLI Commands table does not include the `completion` command (added 2026-05-07) | `station/code-index.md` | Flagged for user review |
| 7 | Info | INDEX.md catalog item count "~50" is now 53 — still within rounding range but slightly low | `station/INDEX.md` — Key Metrics table | Flagged for user review (low priority) |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[HIGH] INDEX.md CLI command count and list are stale** — add `completion` to the list, update count from 8 → 9.
2. **[MEDIUM] CLAUDE.md Workflows table missing `plan-grilling.md`** — add a row with its activation trigger.
3. **[MEDIUM] CLAUDE.md Skills table missing `critic-agent-prompts.md`** — add a row with its activation trigger.
4. **[MEDIUM] INDEX.md has no mention of headless CLI mode** — consider adding a note about `--non-interactive`, `--json`, `--yes` flags shipped in Plans 39 and 41.
5. **[LOW] Plan 40 still in Active Plans** — should be moved to `Playbook/Plans/Shipped/` or equivalent if shipped.
6. **[LOW] code-index.md missing `completion` command row** — add entry pointing to `cmd/completion.go`.

## Notes for Next Run

- The completion subcommand and headless CLI flags (Plans 39, 41) are now 2+ months old — if they remain undocumented by the next run, escalate severity.
- Scan `agent/Workflows/` and `agent/Skills/` at each run for files not listed in CLAUDE.md — the gap between disk reality and the nav table is a recurring failure mode.
- Plan 40 active-plans stale state: if still there next run, note as recurring.
