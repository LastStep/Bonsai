---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-25
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
- **Files Read:** 9
  - `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`
  - `/home/user/Bonsai/station/INDEX.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/agent/Core/memory.md`
  - `/home/user/Bonsai/station/CLAUDE.md`
  - `/home/user/Bonsai/station/code-index.md` (full)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/docs/agent-interface.md` (head)
- **Files Modified:** 2
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry appended)
- **Tools Used:** Read, Bash (git log, ls, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation vs. recent git history

**Git history (last 7 days):**
Only 2 commits in the past 7 days — both routine-only (status-hygiene, backlog-hygiene). No code changes.

Extended to 30-day window to capture meaningful code-change drift. Key relevant commits:
- `2026-06-16` — Plan 41 shipped (5 PRs): headless `*Result` cores, `internal/nonint/`, `list --json`, `docs/agent-interface.md`, exit-code contract
- `2026-05-07` — `bonsai completion` external contribution merged (PR #78)

**Documentation scanned:** `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`, `station/agent/Core/routines.md`, `station/agent/Core/memory.md`, `station/Playbook/Status.md`.

### Step 2 — Check INDEX.md accuracy

Verified tech stack table — accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS/Homebrew).

**CLI command count stale:** INDEX.md says "8 (init, add, remove, list, catalog, update, guide, validate)". The `completion` command was added 2026-05-07 (PR #78). Count should be 9.

**Architecture overview missing `internal/nonint/`:** Plan 41 (2026-06-16) added `internal/nonint/` as a core package (headless CLI contract for init/add/update/remove). The architecture section lists `internal/validate/` and `internal/wsvalidate/` but not `internal/nonint/`.

**ASCII architecture diagram in INDEX.md** also omits `completion` from the `cmd/` row.

**Document Registry missing `docs/agent-interface.md`:** Plan 41 added this contract document as the authoritative reference for headless CLI usage. It is not in the Registry table.

### Step 3 — Check navigation links

**station/CLAUDE.md link verification:**

Core files — all 3 links resolve: `identity.md`, `memory.md`, `self-awareness.md`. Clean.

Protocols — all 4 links resolve: `memory.md`, `scope-boundaries.md`, `security.md`, `session-start.md`. Clean.

Workflows — 8 of 9 links resolve. `plan-grilling.md` exists on disk at `agent/Workflows/plan-grilling.md` but is NOT listed in the Workflows navigation table. (The inverse: unlisted file, not a broken link.) Clean on broken-link test.

Skills — 5 of 6 listed links resolve cleanly. `critic-agent-prompts.md` exists at `agent/Skills/critic-agent-prompts.md` but is NOT in the Skills navigation table. Previously flagged broken link `bonsai-model.md` is now valid (file exists).

Routines — all 7 links resolve. Clean.

Sensors — all 10 links resolve. Clean.

**`station/code-index.md` drift:**
- `bonsai completion` command not in CLI Commands table (added 2026-05-07)
- `internal/nonint/` package section entirely missing (added 2026-06-16, Plan 41)

### Step 4 — Report findings

See Findings Summary below. All findings flagged for user decision — no doc edits applied (per routine procedure).

### Step 5 — Update dashboard

Routine dashboard in `agent/Core/routines.md` updated: Last Ran → 2026-08-25, Next Due → 2026-09-01, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High (persisting) | Root CLAUDE.md project-structure tree severely stale — no `internal/nonint/`, `completion` command, nor all TUI subpackages (addflow/removeflow/updateflow/listflow/hints). First flagged 2026-05-04; still unresolved. | `/home/user/Bonsai/CLAUDE.md` (Project Structure section) | Flagged for user decision |
| 2 | Medium | INDEX.md CLI command count says 8 — should be 9 after `bonsai completion` (PR #78, 2026-05-07). Also missing from the ASCII architecture diagram cmd/ row. | `station/INDEX.md:33` and `:63` | Flagged for user decision |
| 3 | Medium | INDEX.md Architecture overview missing `internal/nonint/` package. Plan 41 (2026-06-16) added it as the headless CLI contract layer between cmd/ and generate/. | `station/INDEX.md:66–71` | Flagged for user decision |
| 4 | Medium | code-index.md missing two items: (a) `bonsai completion` command in CLI Commands table; (b) entire `internal/nonint/` package section (runner.go, nonint.go, result.go, events.go, remove.go, update.go). | `station/code-index.md` (CLI Commands section + internal packages) | Flagged for user decision |
| 5 | Low | INDEX.md Document Registry missing entry for `docs/agent-interface.md` — the headless CLI contract reference added in Plan 41. | `station/INDEX.md` (Document Registry table) | Flagged for user decision |
| 6 | Low | `agent/Workflows/plan-grilling.md` exists but is not listed in station/CLAUDE.md Workflows navigation table. May be intentionally unlisted or overlooked. | `station/CLAUDE.md` (Workflows table) | Flagged for user decision |
| 7 | Low | `agent/Skills/critic-agent-prompts.md` exists but is not listed in station/CLAUDE.md Skills navigation table. May be intentionally unlisted or overlooked. | `station/CLAUDE.md` (Skills table) | Flagged for user decision |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

### F1 (High) — Root CLAUDE.md project-structure tree
**Status:** Persisting from 2026-05-04 (3rd consecutive run without resolution).
**Proposed update:** Add `internal/nonint/` row, update cmd/ to include `completion.go`, update tui/ to list all subpackages. Alternatively, truncate the tree to high-level only and note "see code-index.md for detail."

### F2+F3 (Medium) — INDEX.md CLI count + architecture
**Proposed update (two changes):**
1. Line 33: `CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate)` → `9 (init, add, remove, list, catalog, update, guide, validate, completion)`
2. Line 63: append `completion` to cmd/ row
3. Add `internal/nonint/ ← headless CLI contract — *Result cores for init/add/update/remove (Plan 41)` row to architecture block

### F4 (Medium) — code-index.md missing nonint + completion
**Proposed update:** Add `bonsai completion` row to CLI Commands table; add new `## Nonint (internal/nonint/)` section documenting `runner.go`, `nonint.go`, `result.go`, `events.go`, `remove.go`, `update.go`.

### F5 (Low) — INDEX.md Document Registry entry for agent-interface.md
**Proposed update:** Add row: `docs/agent-interface.md | Headless CLI contract — flags, JSONL/JSON serializations, exit codes for non-interactive use | When driving bonsai from CI, an AI agent, or the planned bonsai mcp server`

### F6+F7 (Low) — Unlisted workflow/skill files
`plan-grilling.md` and `critic-agent-prompts.md` exist on disk but are absent from the CLAUDE.md navigation table. If actively used, add them. If internal/scratch, note them in a comment or delete them.

## Notes for Next Run

- Root CLAUDE.md project-structure tree (F1) is now on its third cycle unfixed. If user doesn't intend to maintain the full tree, consider replacing it with a pointer to `code-index.md` to break the recurring cycle.
- INDEX.md CLI count (F2) was previously fixed 7→8 in the 2026-05-04 routine digest but immediately drifted again when `completion` was merged 3 days later. Consider automating this count from the actual registered Cobra commands.
- The previously broken `agent/Skills/bonsai-model.md` nav link is now resolved — the file exists and the link is valid.
- No broken links found this cycle in any Core/Protocols/Workflows/Skills/Routines/Sensors navigation table.
