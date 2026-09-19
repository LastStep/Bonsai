---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-19
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (value from dashboard before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~10 min
- **Files Read:** 8 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation vs recent git history
- **Action:** Read `git log --since="7 days ago"` and compared against station docs.
- **Result:** 2 commits found in last 7 days — both from the `backlog-hygiene` routine run (2026-09-19). No new feature code landed in the last 7 days. Widened to 30 days: same 2 commits. The meaningful doc-relevant code work (Plan 41 headless CLI contract, shipped 2026-06-16) is ~3 months old — predates the last doc-freshness-check run (2026-05-04) by about 6 weeks, meaning the May run pre-dated Plan 41. All Plan 41 drift is new since last check.
- **Issues:** 4 documentation gaps from Plan 41 not reflected in station docs (see Findings Summary).

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` in full and compared Tech Stack, Key Metrics, Architecture Overview, and Document Registry against actual codebase.
- **Result:**
  - **Tech Stack**: Accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template all confirmed correct.
  - **Key Metrics — CLI commands count**: STALE. Lists `8 (init, add, remove, list, catalog, update, guide, validate)` but `cmd/completion.go` exists — there are 9 commands. `bonsai completion [bash|zsh|fish|powershell]` was merged (PR #78, @mvanhorn, 2026-05-07) after Plan 37's doc refresh and was never added to INDEX.md.
  - **Key Metrics — Agent types**: Accurate (6: tech-lead, fullstack, backend, frontend, devops, security confirmed via `ls catalog/agents/`).
  - **Key Metrics — Catalog items**: "~50" — approximate, not verified precisely but plausible.
  - **Architecture Overview diagram**: STALE. `internal/nonint/` package added in Plan 41 is missing from the architecture list. Currently lists: catalog, config, generate, validate, wsvalidate, tui — missing `nonint`.
  - **Document Registry**: STALE. `docs/agent-interface.md` (Plan 41 headless CLI contract doc) exists on disk but is not listed.
- **Issues:** 3 stale items in INDEX.md.

### Step 3: Check navigation links in CLAUDE.md and agent instruction dirs
- **Action:** Read `station/CLAUDE.md` in full and cross-referenced all linked files against actual filesystem.
- **Result:**
  - All Core links resolve: identity.md ✓, memory.md ✓, self-awareness.md ✓
  - All Protocol links resolve: memory.md ✓, scope-boundaries.md ✓, security.md ✓, session-start.md ✓
  - All Workflow links resolve: code-review.md ✓, planning.md ✓, pr-review.md ✓, security-audit.md ✓, session-logging.md ✓, test-plan.md ✓, session-wrapup.md ✓, issue-to-implementation.md ✓, routine-digest.md ✓
  - All Skill links resolve: planning-template.md ✓, review-checklist.md ✓, issue-classification.md ✓, pr-creation.md ✓, bubbletea.md ✓, bonsai-model.md ✓
  - All Routine links resolve: all 7 routine files present ✓
  - All Sensor script files present ✓ (10 sensors all exist)
  - **Unlisted files**: `agent/Workflows/plan-grilling.md` and `agent/Skills/critic-agent-prompts.md` exist on disk but are NOT listed in CLAUDE.md's navigation tables. Both have frontmatter noting "full Bonsai-catalog integration pending (Backlog)" — intentional omission awaiting catalog wiring, but agent cannot easily discover these files without directory scan.
- **Issues:** No broken links found. 2 custom files present but undiscoverable via CLAUDE.md (low priority — intentional pending backlog item).

### Step 4: Check code-index.md accuracy
- **Action:** Scanned `station/code-index.md` for references to recently-added code components.
- **Result:**
  - `internal/nonint/` package: NOT listed anywhere in code-index.md. This package has 14 Go files (config, events, nonint, remove, result, runner, update + tests) and is the entire headless CLI contract layer from Plan 41.
  - `bonsai completion` command: NOT listed in the CLI Commands table in code-index.md. `cmd/completion.go` exists.
  - All other packages and functions that were documented in Plan 37 (2026-05-07) appear to still reference the correct files and functions — line numbers may have drifted slightly but function names are stable.
- **Issues:** 2 missing entries in code-index.md.

### Step 5: Report findings and flag for user decision
- **Action:** Compiled all findings below. Per routine procedure: propose but do not execute doc updates — flag for user decision.
- **Result:** 5 findings documented (4 in INDEX.md/code-index.md, 1 low-severity custom-file discovery gap).
- **Issues:** none (reporting step completed cleanly).

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | `internal/nonint/` package missing from Architecture Overview and Document Registry | `station/INDEX.md` | Flagged — propose adding to Architecture diagram and description |
| 2 | medium | `bonsai completion` command missing — CLI commands count says 8, actually 9 | `station/INDEX.md` Key Metrics | Flagged — propose updating count to 9 and listing completion |
| 3 | medium | `docs/agent-interface.md` not in Document Registry | `station/INDEX.md` | Flagged — propose adding row for agent-interface contract doc |
| 4 | medium | `internal/nonint/` package undocumented in code index | `station/code-index.md` | Flagged — propose adding a Nonint section documenting runner.go, result.go, events.go, config.go |
| 5 | low | `plan-grilling.md` and `critic-agent-prompts.md` present in agent dirs but not discoverable via CLAUDE.md navigation | `station/CLAUDE.md` | Noted — intentional (Backlog: "full Bonsai-catalog integration pending"). Already tracked in Backlog. No action needed now. |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**Finding 1–4** are all medium-severity and relate to Plan 41 work (headless CLI contract + completion command) that landed after the last doc-freshness-check. These are all audit flags — no docs were changed. Suggested user decisions:

1. **INDEX.md Architecture Overview**: Add `internal/nonint/` to the architecture list. Proposed text: `"internal/nonint/ ← headless CLI contract layer — *Result cores, JSONL serialization, exit-code constants"`.
2. **INDEX.md Key Metrics**: Change "8" to "9" and append `completion` to the commands list.
3. **INDEX.md Document Registry**: Add a row for `docs/agent-interface.md` — headless CLI contract (flags, JSONL shapes, exit codes) for MCP/AI integrators.
4. **code-index.md**: Add a `## Nonint (internal/nonint/)` section documenting the key types and functions — particularly `runner.go` (exit-code constants), `result.go` (WriteResult headless shape), `events.go` (JSONL event shapes).

These are all low-risk documentation-only updates with no code impact. Could be batched into a single Plan 43 "doc refresh" or handled in a next session.

## Notes for Next Run

- Plan 41 drift (4 items) is the bulk of this run's findings. If a doc-refresh plan is executed before next run, these should all be cleared.
- Finding 5 (plan-grilling/critic-agent-prompts) will remain until those abilities are wired into the catalog — check Backlog entry status at next run.
- `bonsai completion` was merged 2026-05-07 (PR #78) — nearly 4.5 months ago. Consider whether to add `completion` to root CLAUDE.md project instructions as well (currently only listed in code help output, not docs).
- All navigation links in CLAUDE.md are intact. No broken links found.
- Next run due: 2026-09-26.
