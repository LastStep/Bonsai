---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-15
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (findings require user action — procedure instructs flag-only, no doc edits)
- **Duration:** ~8 min
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/CLAUDE.md` (via system-reminder), `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/Logs/RoutineLog.md`, `station/agent/Workflows/plan-grilling.md`, `station/agent/Skills/critic-agent-prompts.md`, `station/Reports/Pending/` (listing), `station/` (listing)
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Glob, Bash (git log, ls), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Ran `git log --since="2026-05-04" --oneline` to get all commits since last run (72 days). Read `station/INDEX.md` and `station/CLAUDE.md`. Checked `station/Playbook/Status.md` for recent work context.
- **Result:** 48+ commits identified since 2026-05-04. Key changes shipped:
  - Plan 39 (v0.4.2): `bonsai init/add --non-interactive --from-config`
  - External contribution (PR #78): `bonsai completion [bash|zsh|fish|powershell]` added
  - Plan 40 (v0.5.0): freeze schemas, root-relative scaffolding, project-level validate pass, memory-routing docs
  - Plan 41: full headless CLI contract — pure `*Result` cores + JSONL/exit contract for all mutating commands; `list --json`; `docs/agent-interface.md` contract doc
  - Plan grilling pipeline added: 6-critic adversarial review workflow + skill prompts
- **Issues:** 3 documentation gaps found (detailed in Findings Summary)

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` in full. Cross-checked tech stack, folder structure, key metrics, and architecture overview against current codebase state.
- **Result:** Tech stack correct. Architecture overview accurate. Key Metrics table has one stale row: CLI commands count says 8 but `completion` was added (commit `2eae9d4`, PR #78), making it 9.
- **Issues:** 1 stale metric (see Finding #3)

### Step 3: Check navigation links
- **Action:** Globbed all files in `station/agent/Core/`, `station/agent/Protocols/`, `station/agent/Workflows/`, `station/agent/Skills/`, `station/agent/Sensors/`. Cross-checked against `station/CLAUDE.md` navigation tables.
- **Result:**
  - Core (4 files): identity.md, memory.md, self-awareness.md, routines.md — all referenced correctly
  - Protocols (4 files): memory.md, scope-boundaries.md, security.md, session-start.md — all referenced correctly
  - Sensors (10 files): all 10 match the Sensors table in CLAUDE.md exactly
  - Routines (7 files): all match dashboard
  - Workflows: 10 files on disk, 9 in CLAUDE.md table — `plan-grilling.md` NOT listed (Finding #1)
  - Skills: 7 files on disk, 6 in CLAUDE.md table — `critic-agent-prompts.md` NOT listed (Finding #2)
- **Issues:** 2 missing navigation entries (Findings #1 and #2)

### Step 4: Report findings
- **Action:** Compiled findings table below. Per procedure, proposing updates but not executing — flagged for user decision.
- **Result:** 3 findings identified, all low-to-medium severity, no broken links.
- **Issues:** none

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` — Doc Freshness Check row: `Last Ran → 2026-07-15`, `Next Due → 2026-07-22`, `Status → done`.
- **Result:** Dashboard updated.
- **Issues:** none

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | `plan-grilling.md` workflow exists on disk but is absent from the Workflows navigation table in `station/CLAUDE.md` — agents cannot discover it via nav | `station/CLAUDE.md` Workflows table | Flagged — user decision required |
| 2 | LOW | `critic-agent-prompts.md` skill exists and is consumed by plan-grilling.md but is absent from the Skills table in `station/CLAUDE.md` | `station/CLAUDE.md` Skills table | Flagged — user decision required |
| 3 | LOW | INDEX.md Key Metrics row says "CLI commands: 8 (init, add, remove, list, catalog, update, guide, validate)" — `completion` subcommand was added in PR #78 (2026-05-07), making the correct count 9 | `station/INDEX.md` Key Metrics table | Flagged — user decision required |

---

## Proposed Updates (for user approval)

### Finding #1 — Add plan-grilling.md to CLAUDE.md Workflows table

In the Workflows table, add this row after the `routine-digest.md` row:

```
| Adversarially reviewing a drafted Tier-2+ plan before dispatch; Running a critic pass ("grill the plan", "critic pass", "review plan NN") | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

Also consider adding to Quick Triggers:
```
| Running adversarial critic review of a plan | "grill the plan" or /plan-grilling |
```

### Finding #2 — Add critic-agent-prompts.md to CLAUDE.md Skills table

In the Skills table, add this row:

```
| Dispatching critic agents for plan grilling; Verbatim prompt templates for 6-critic suite (consumed by plan-grilling workflow) | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

### Finding #3 — Update INDEX.md CLI command count

Change:
```
| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |
```
To:
```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

Three items require user approval before changes are made:

1. **Add `plan-grilling.md` to CLAUDE.md Workflows table** — workflow is live and actively used (Plans 40+41) but invisible via nav. Recommend approving.
2. **Add `critic-agent-prompts.md` to CLAUDE.md Skills table** — consumed by plan-grilling, low friction fix. Recommend approving.
3. **Update INDEX.md CLI command count 8 → 9** — trivial one-liner, recommend approving.

No other architectural, security, or stack-level drift was found. All other navigation links resolve correctly.

---

## Notes for Next Run

- Gap since last run was 72 days (vs 7-day frequency) — routine has been systematically deferred. Backlog Hygiene report from today also flags this.
- Station docs are otherwise in good shape: all sensor/protocol/core/workflow links valid.
- Watch for Plan 42 (MCP server) — will likely add new CLI surface or external integration docs that need INDEX.md + CLAUDE.md coverage.
- `docs/agent-interface.md` was created in Plan 41 Phase 5 (headless contract doc) — confirm it's referenced somewhere agents can find it if it matters.
