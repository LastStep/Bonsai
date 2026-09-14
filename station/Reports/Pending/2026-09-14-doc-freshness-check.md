---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-14
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
- **Duration:** ~10 min
- **Files Read:** 9 — `station/agent/Routines/doc-freshness-check.md`, `station/CLAUDE.md`, `station/INDEX.md`, `station/Playbook/Status.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/code-index.md`, `cmd/` (directory listing), `docs/` (directory listing)
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

**Step 1 — Scan project documentation:**
Read `station/INDEX.md`, `station/CLAUDE.md`, and `station/code-index.md`. Compared against `git log --since="2026-05-04"` (last routine run date). Found ~30 commits since last check, including Plan 40 (platform integration, validate pass), Plan 41 (headless CLI contract, MCP-ready cores), `bonsai completion` command (PR #78), and v0.4.2/v0.4.3 releases. Recent routine commits (backlog-hygiene, status-hygiene) from today are routine maintenance and do not affect doc accuracy.

**Step 2 — Check INDEX.md accuracy:**
Tech stack table is accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS).  
Architecture diagram is accurate.  
**Drift found:** Key Metrics row "CLI commands: 8" is stale — `bonsai completion` was added in PR #78 (2026-05-07), making the real count 9.  
**Drift found:** Document Registry is missing `docs/agent-interface.md`, which was shipped as the headless/MCP contract document in Plan 41 (June 2026).

**Step 3 — Check navigation links:**
Verified all links in `station/CLAUDE.md` — Core, Protocols, Workflows, Skills, Routines, Sensors, and External References sections. All referenced files exist on disk. No broken links found.

Additionally checked agent subdirectory listings for files present but unlisted:
- `agent/Workflows/plan-grilling.md` exists (added ~2026-05-04 with the 6-critic adversarial plan review pipeline) but is NOT in the Workflows nav table.
- `agent/Skills/critic-agent-prompts.md` exists but is NOT in the Skills nav table.

**Step 4 — Report findings:**
See Findings Summary below. All items are flagged for user decision — no doc edits executed.

**Step 5 — Update dashboard:**
Dashboard row updated in `agent/Core/routines.md`.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `CLI commands: 8` is stale — `bonsai completion` was added (PR #78, 2026-05-07), count should be 9. Also update command list to include `completion`. | `station/INDEX.md` — Key Metrics table | Flagged — propose updating count to 9 and adding `completion` to the list |
| 2 | Low | `docs/agent-interface.md` (headless/MCP CLI contract, Plan 41) missing from Document Registry | `station/INDEX.md` — Document Registry | Flagged — propose adding row: `docs/agent-interface.md` \| Headless CLI + MCP contract — Result types, exit codes, JSONL contract for all mutating commands \| When planning MCP server (Plan 42) or evaluating headless integration |
| 3 | Low | `bonsai completion` command missing from CLI Commands table | `station/code-index.md` | Flagged — propose adding row for `cmd/completion.go` |
| 4 | Low | `agent/Workflows/plan-grilling.md` exists (6-critic adversarial plan review pipeline, active) but absent from Workflows nav table | `station/CLAUDE.md` | Flagged — propose adding row: activate when "Running adversarial plan review / grilling a plan before dispatch" |
| 5 | Low | `agent/Skills/critic-agent-prompts.md` exists but absent from Skills nav table | `station/CLAUDE.md` | Flagged — verify if intentionally unlisted (internal-only) or should be added |
| 6 | Low | `cmd/` structure listing omits `completion.go` (added PR #78, 2026-05-07) | Root `CLAUDE.md` — Project Structure | Flagged — propose adding `completion.go` entry to cmd/ listing |

## Errors & Warnings

No errors encountered.

**Note from prior backlog-hygiene run (2026-09-14):** Three related doc-drift items were already filed in the backlog — "code-index.md", "bonsai-model.md nav link", and "INDEX.md arch diagram". Findings 1, 2, and 3 above overlap and expand on those backlog items. The nav link for bonsai-model.md was checked and found to be valid; the concern may have been about content staleness, not a broken link.

## Items Flagged for User Review

1. **INDEX.md CLI count** (Finding 1) — quick one-liner fix; recommend updating now.
2. **INDEX.md missing `docs/agent-interface.md`** (Finding 2) — add a Document Registry row.
3. **`code-index.md` missing `completion` command** (Finding 3) — add row to CLI Commands table.
4. **CLAUDE.md plan-grilling workflow** (Finding 4) — add to Workflows nav table so the pipeline is discoverable.
5. **CLAUDE.md critic-agent-prompts skill** (Finding 5) — decision needed: add to nav or confirm intentionally internal-only.
6. **Root CLAUDE.md `completion.go`** (Finding 6) — add to cmd/ structure listing.

All items are low-risk doc maintenance. Items 1–4 and 6 are clear fixes; item 5 needs a decision.

## Notes for Next Run

- Line numbers in `code-index.md` were not re-validated against actual source (out of scope for drift check — would require full reindex). If Plan 41 reshuffled `cmd/root.go`, `cmd/add.go`, or `cmd/remove.go` significantly, line references may be stale. Consider a targeted re-index run.
- Next due: 2026-09-21. If Plan 42 (MCP server) ships before then, a mid-cycle check on `docs/agent-interface.md` and INDEX.md may be worthwhile.
