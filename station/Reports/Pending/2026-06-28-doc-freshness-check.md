---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-28
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
- **Duration:** ~10 min
- **Files Read:** 12 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/CLAUDE.md`, `station/code-index.md`, `.bonsai.yaml`, `station/agent/Workflows/plan-grilling.md`, `station/agent/Skills/critic-agent-prompts.md`, `station/Playbook/Plans/Active/41-headless-cli-contract.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** git log, ls, grep; Read tool for file inspection
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`. Ran `git log --since="7 days ago"` to get recent commits.
- **Result:** Only 2 commits in the last 7 days (both from today's routine runs: status-hygiene and backlog-hygiene). Checked the last 14 days and found Plan 41 was shipped 2026-06-16 (5 PRs merged, headless CLI contract complete). No new features or services added in the past 7 days that require immediate doc updates. Plan 41 changes (headless cores, `internal/nonint/`, `bonsai completion`) are not yet reflected in root `CLAUDE.md` or `station/INDEX.md`.
- **Issues:** Several documentation gaps found — see findings below.

### Step 2: Check INDEX.md accuracy
- **Action:** Compared `station/INDEX.md` tech stack, CLI command count, and catalog item count against actual codebase state.
- **Result:** Tech stack is accurate. However:
  - CLI commands count says "8" but `bonsai completion` was added (PR #78, first external contribution, 2026-05-07), making it 9.
  - Catalog items says "~50" but actual count is 63 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines + 6 agents = 59 distinct non-agent items + 6 agents). The approximation "~50" is visibly stale at 63 total items.
- **Issues:** Two stale metrics in INDEX.md Key Metrics table.

### Step 3: Check navigation links
- **Action:** Verified every link in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, Bonsai Reference, External References).
- **Result:** All listed links resolve to real files — no broken links found. However, two custom files exist in the workspace that are NOT listed in the station/CLAUDE.md navigation tables:
  - `agent/Workflows/plan-grilling.md` — adversarial plan review workflow (custom, adapted from ZenGarden 2026-06-13); not listed in Workflows table.
  - `agent/Skills/critic-agent-prompts.md` — critic agent prompt templates companion to plan-grilling; not listed in Skills table.
  Also checked root `CLAUDE.md` project structure — it lists only `release.yml` in `.github/workflows/` but actual directory contains `ci.yml`, `codeql.yml`, `docs.yml`, `release.yml`. The `cmd/completion.go` and `internal/nonint/` directory (added by Plan 41) are absent from the project structure listing.
- **Issues:** 4 undocumented files/directories in root CLAUDE.md; 2 unlisted custom abilities in station/CLAUDE.md.

### Step 4: Report findings
- **Action:** Compiled all findings into this report with severity levels.
- **Result:** 6 findings total — 2 medium, 4 low. None are blocking or critical. Flagging for user decision per the procedure (do not execute changes without user approval).
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for "Doc Freshness Check".
- **Result:** Last Ran → 2026-06-28, Next Due → 2026-07-05, Status → done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `plan-grilling.md` workflow exists in `agent/Workflows/` but is NOT listed in station/CLAUDE.md Workflows navigation table | `station/CLAUDE.md` → Workflows section | Flagged for user — add row or accept as unlisted custom |
| 2 | Medium | `critic-agent-prompts.md` skill exists in `agent/Skills/` but is NOT listed in station/CLAUDE.md Skills navigation table | `station/CLAUDE.md` → Skills section | Flagged for user — add row or accept as unlisted custom |
| 3 | Low | Root `CLAUDE.md` project structure lists only `release.yml` under `.github/workflows/` but `ci.yml`, `codeql.yml`, `docs.yml` also exist | `/CLAUDE.md` project structure tree | Flagged for user — update tree listing |
| 4 | Low | Root `CLAUDE.md` project structure missing `cmd/completion.go` and `internal/nonint/` directory (added by Plan 41) | `/CLAUDE.md` project structure tree | Flagged for user — add missing entries |
| 5 | Low | `station/INDEX.md` Key Metrics: CLI commands count says "8" but `bonsai completion` makes it 9 | `station/INDEX.md` Key Metrics table | Flagged for user — update count to 9 |
| 6 | Low | `station/INDEX.md` Key Metrics: Catalog items says "~50" but actual count is 63 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines + 6 agents) | `station/INDEX.md` Key Metrics table | Flagged for user — update to "~60" or exact count |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

All 6 findings above require user decision. These are doc-only drifts — no code changes needed. Suggested actions:

**Findings 1 & 2 — station/CLAUDE.md missing custom abilities:**
Add the following rows to station/CLAUDE.md (Workflows and Skills tables respectively):
- Workflows: `| Adversarially reviewing a drafted plan before dispatch | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |`
- Skills: `| Critic agent prompts for plan-grilling (6 critics — 5 prose + Reality) | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |`

**Findings 3 & 4 — root CLAUDE.md project structure stale:**
Update the `.github/workflows/` tree stanza to include `ci.yml`, `codeql.yml`, and `docs.yml`. Add `cmd/completion.go` and `internal/nonint/` directory entries.

**Findings 5 & 6 — INDEX.md Key Metrics stale:**
Update CLI commands count from "8" to "9 (init, add, remove, list, catalog, update, guide, validate, completion)". Update catalog items from "~50" to "~60" or provide the precise count.

## Notes for Next Run

- The `internal/nonint/` package added by Plan 41 is undocumented in both root CLAUDE.md and code-index.md — the code-index entry is worth adding if the user wants it (it covers headless cores: `config.go`, `events.go`, `nonint.go`, `remove.go`, `result.go`, `runner.go`, `update.go`).
- Plan 41 plan file remains in `Plans/Active/41-headless-cli-contract.md` despite all phases shipped. This was also flagged by status-hygiene. User should archive it.
- `catalog/routines/infra-drift-check` exists in the catalog but is not installed in this station — this is an opt-in item, not a doc gap. No action needed unless the user wants to install it.
- All navigation links in station/CLAUDE.md resolve correctly — no broken links.
