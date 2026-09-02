---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-02
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (findings flagged; doc updates require user decision per procedure)
- **Duration:** ~5 minutes
- **Files Read:** 8 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/code-index.md`, `station/CLAUDE.md` (via system reminder), `station/agent/Core/memory.md`, `station/agent/Core/routines.md`, `docs/agent-interface.md` (head), `station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls, file-existence checks, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Ran `git log --oneline -20` and `git log --since="7 days ago"` to identify recent changes. Identified Plan 41 ("Headless CLI Contract") as the most recent shipped plan — 5 PRs merged, last commit `c6a6757` (2026-06-16). No commits in the last 7 days; coverage extended to 14+ days to capture the most recent material changes.
- **Result:** Plan 41 shipped `internal/nonint/` (new package), `docs/agent-interface.md` (new canonical reference), and headless flags across init/add/remove/update/list commands. Two documentation files (`station/INDEX.md` and `station/code-index.md`) were identified as not reflecting these changes.
- **Issues:** 2 documentation drift items found (detailed below).

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` in full. Verified tech stack table, CLI command count, agent type count, and document registry against actual codebase state.
- **Result:** Tech stack (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template) — accurate. CLI commands (8: init, add, remove, list, catalog, update, guide, validate) — accurate. Agent types (6) — accurate. **Document Registry — missing the `docs/` directory entirely.** Plan 41 created `docs/` with 7 files including `docs/agent-interface.md` (the canonical headless CLI contract). None are listed in the registry.
- **Issues:** 1 drift item — docs/ directory and its contents absent from INDEX.md Document Registry.

### Step 3: Check navigation links
- **Action:** Enumerated all 49 linked files from `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References) and checked each with a file-existence test.
- **Result:** 49/49 links resolve to existing files. No broken navigation links.
- **Issues:** None.

### Step 4: Report findings
- **Action:** Compiled 2 documentation drift findings and 1 housekeeping flag. Per procedure: findings listed below, updates proposed but not executed.
- **Issues:** None in execution.

### Step 5: Update dashboard
- **Action:** Updated Doc Freshness Check row in `station/agent/Core/routines.md` — Last Ran → 2026-09-02, Next Due → 2026-09-09, Status → done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `docs/` directory (7 files incl. `agent-interface.md`) not referenced in Document Registry — Plan 41 shipped this as the canonical headless/MCP contract; Plan 42 depends on it | `station/INDEX.md` | Flagged for user decision — proposed row: `docs/agent-interface.md` — headless CLI contract (flags, exit codes, serialization) — *When integrating Bonsai non-interactively or building MCP wrapper* |
| 2 | Medium | `internal/nonint/` package (7 source files + tests, exit-code constants, Result/Event types, headless runners) entirely absent from code-index — shipped in Plan 41 | `station/code-index.md` | Flagged for user decision — needs a new "Headless CLI (`internal/nonint/`)" section covering: `runner.go` exit-code constants, `result.go` Result shape, `events.go` event stream types, `nonint.go` entry, `remove.go`/`update.go` per-command cores |
| 3 | Low | Plan 41 plan file remains in `Plans/Active/` despite being fully shipped (2026-06-16) — memory explicitly flags this for archiving | `station/Playbook/Plans/Active/41-headless-cli-contract.md` | Flagged for user decision — move to `Plans/Archive/` |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**Finding 1 — INDEX.md Document Registry gap (`docs/`)**

Proposed addition to the Document Registry table in `station/INDEX.md`:

```
| `docs/agent-interface.md` | Headless CLI contract — flags, exit codes, JSONL serialization for non-interactive/MCP use | When integrating Bonsai non-interactively or building the MCP wrapper (Plan 42) |
```

Optionally add a general `docs/` row covering the user-facing guides (cli.md, concepts.md, quickstart.md, custom-files.md, formats.md) if you want the registry to be the single lookup for all documentation.

**Finding 2 — code-index.md missing `internal/nonint/`**

Plan 41 added a full new package. Proposed new section for `station/code-index.md` (after the Workspace-path Validation section):

```markdown
## Headless CLI (`internal/nonint/`) — Plan 41

Pure-function cores for non-interactive driving — typed options in, structured Result out.
No prompts, no os.Exit, no data on stdout from inside the core.
Exit-code constants defined in `runner.go` are canonical (docs/agent-interface.md references them).

| Type / Function | File | Purpose |
|-----------------|------|---------|
| ExitOK=0 / ExitInvalidConfig=2 / ExitRuntime=3 / ExitWrongCWDForInit=4 / ExitConflict=5 | runner.go | Exit-code contract (canonical — wins over docs/agent-interface.md if they disagree) |
| Result / Warnings | result.go | Structured output shape for all mutating commands |
| FileEvent / SummaryEvent | events.go | JSONL event types emitted to stdout on headless path |
| RunInit / RunAdd | nonint.go | Headless init + add cores |
| RunRemove | remove.go | Headless remove core (--yes / --from flags) |
| RunUpdate | update.go | Headless update core (--non-interactive / --skip-conflicts flags) |
```

**Finding 3 — Plan 41 archive**

Move `station/Playbook/Plans/Active/41-headless-cli-contract.md` → `station/Playbook/Plans/Archive/41-headless-cli-contract.md`. This was already noted in memory as a deferred wrap-up action.

## Notes for Next Run

- If Plan 42 (MCP server) ships before the next run, check whether `docs/agent-interface.md` is still accurate as the canonical reference and whether any new docs were added.
- The `docs/` directory is growing — consider whether it warrants a standing row in INDEX.md beyond just `agent-interface.md`.
- Plan 40 (`40-odysseus-platform-integration.md`) is also in `Plans/Active/` — memory says phases 1-3 shipped as v0.5.0, phase 4 was held and superseded by Plan 41. Check whether this plan also needs archiving at next status-hygiene or backlog-hygiene run.
