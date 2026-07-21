---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-21
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
- **Duration:** ~12 min
- **Files Read:** 11 — `station/agent/Routines/doc-freshness-check.md`, `station/CLAUDE.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`, `station/agent/Workflows/plan-grilling.md`, `station/agent/Skills/critic-agent-prompts.md`, `/home/user/Bonsai/embed.go`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log append)
- **Tools Used:** `git log --oneline --since="2026-05-04"`, `git show --stat ab202c3`, `find catalog -name "meta.yaml" | wc -l`, Python link-checker script against `station/CLAUDE.md`, `ls` on `cmd/`, `internal/`, `docs/`, `agent/Workflows/`, `agent/Skills/`
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Ran `git log --since="2026-05-04" --oneline` to get all commits since the last doc freshness check. Reviewed commit messages and key commit diffs (Plan 40 + Plan 41 phases).
- **Result:** 40+ commits since 2026-05-04. Two major plans shipped:
  - **Plan 40** — Odysseus platform integration (frozen schemas, root-relative scaffolding, project-level validate, memory-routing protocol, guide Formats page)
  - **Plan 41** — Headless CLI Contract (5 phases): Result reshape, headless update/remove cores, `list --json`, agent-interface contract doc
- **Issues:** Several new features, files, and packages introduced without corresponding doc updates (detailed in steps below).

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` and compared Tech Stack, Key Metrics, Architecture Overview, and Document Registry against the actual codebase.
- **Result:** Three drift items found:
  1. **Key Metrics — CLI command count stale:** INDEX.md says "8 (init, add, remove, list, catalog, update, guide, validate)" but `bonsai completion` was added in commit `2eae9d4` (2026-05-07), making it 9 commands.
  2. **Architecture Overview — `internal/nonint/` missing:** The diagram shows 6 internal packages but omits `internal/nonint/` which was introduced in Plan 41 as the headless CLI contract package (config, events, result, runner, update, remove sub-modules).
  3. **Document Registry — `docs/agent-interface.md` not listed:** Plan 41 Phase 5 added `docs/agent-interface.md` as the canonical headless CLI contract reference (245 lines, covers per-command flags, serialization formats, exit codes, stream discipline). It is the primary external-consumer contract for MCP server integration. The Document Registry makes no mention of the `docs/` directory or this file.
- **Issues:** 3 medium-severity drift items.

### Step 3: Check navigation links
- **Action:** Ran a Python link-checker against all 55 markdown links in `station/CLAUDE.md`. Verified existence of all files in `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/`.
- **Result:**
  - **All 55 links resolve** — zero broken links in `station/CLAUDE.md`.
  - **Two unlisted files discovered** via directory listing:
    - `agent/Workflows/plan-grilling.md` — a complete workflow for adversarial 6-critic plan review. Added in commit `6995d4f`. Not listed in `station/CLAUDE.md` Workflows navigation table.
    - `agent/Skills/critic-agent-prompts.md` — companion skill with verbatim prompt templates for the 6 critics. Added alongside plan-grilling. Not listed in `station/CLAUDE.md` Skills navigation table.
  - These files are functional and contain complete `tags`, `description`, and `source` frontmatter but are invisible to the agent navigation.
- **Issues:** 2 high-severity drift items (unlisted navigable files).

### Step 4: Report findings
- **Action:** Compiled all drift into the Findings Summary table below. All items flagged for user decision — no doc edits made.
- **Result:** 8 findings total (2 high, 3 medium, 3 low). Proposed updates specified per finding.
- **Issues:** None — findings compiled correctly.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for "Doc Freshness Check" — `Last Ran` → 2026-07-21, `Next Due` → 2026-07-28, `Status` → done.
- **Result:** Dashboard updated.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | `plan-grilling.md` exists in `agent/Workflows/` but is absent from `station/CLAUDE.md` Workflows navigation table — agent cannot discover or load this workflow by name | `station/CLAUDE.md` Workflows table | Flagged for user — add row: `\| Adversarial plan review via 6-critic agents looped to convergence before dispatch \| [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) \|` |
| 2 | high | `critic-agent-prompts.md` exists in `agent/Skills/` but is absent from `station/CLAUDE.md` Skills navigation table — companion to plan-grilling, invisible to navigation | `station/CLAUDE.md` Skills table | Flagged for user — add row: `\| Verbatim prompt templates for the 6 plan-grilling critic agents. Load alongside plan-grilling.md. \| [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) \|` |
| 3 | medium | `internal/nonint/` package missing from Architecture Overview — Plan 41 introduced this as the headless CLI contract layer (config, events, result, runner) | `station/INDEX.md` Architecture section | Flagged for user — add line to diagram: `internal/nonint/    ← headless CLI contract — Result, JSONL events, exit codes, RunInit/RunAdd orchestrators` |
| 4 | medium | CLI command count stale — INDEX.md says 8 commands but `bonsai completion` (added 2026-05-07) brings the total to 9 | `station/INDEX.md` Key Metrics table | Flagged for user — update row: `\| CLI commands \| 9 (init, add, remove, list, catalog, update, guide, validate, completion) \|` |
| 5 | medium | `docs/agent-interface.md` (Plan 41 Phase 5 canonical headless contract) not in Document Registry — key reference for MCP integration work (Plan 42) | `station/INDEX.md` Document Registry | Flagged for user — add row: `` \| `docs/agent-interface.md` \| Canonical headless CLI contract — per-command flags, JSONL serialization, exit codes, stream discipline \| Plan 42 MCP work; headless CI consumers \| `` |
| 6 | low | `bonsai completion` command missing from code-index.md CLI Commands table | `station/code-index.md` | Flagged for user — add row to CLI Commands table: `` \| `bonsai completion` \| `cmd/completion.go:21` \| `completionCmd` — shell completion for bash/zsh/fish/powershell \| `` |
| 7 | low | `internal/nonint/` package has no section in code-index.md — omits all public types (Config, Result, Counts) and runners (RunInit, RunAdd, exit constants) | `station/code-index.md` | Flagged for user — add new section between Workspace-path Validation and TUI |
| 8 | low | `embed.go` entry point row in code-index.md omits `GuideFormats` — `embed.go:24` adds `GuideFormats` (formats.md), but the code-index only lists 4 guide vars | `station/code-index.md` Entry Point table | Flagged for user — update row: `embed.go:12–25 — GuideCustomFiles, GuideQuickstart, GuideConcepts, GuideCli, GuideFormats` |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

All 8 findings above require a doc edit decision. Suggested priority order:

- **Do now (high):** Add `plan-grilling.md` and `critic-agent-prompts.md` to `station/CLAUDE.md` navigation — these are active, used workflows/skills that the agent cannot currently discover from the nav table.
- **Do soon (medium):** Update `station/INDEX.md` for `internal/nonint/` in the architecture diagram, CLI command count (8 → 9), and add `docs/agent-interface.md` to the Document Registry (relevant for upcoming Plan 42 MCP work).
- **Defer (low):** Update `station/code-index.md` for `completion` command, `internal/nonint/` section, and `GuideFormats` embed entry — useful but not navigational blockers.

---

## Notes for Next Run

- The `docs/` directory now has 8 files (`README.md`, `agent-interface.md`, `assets/`, `cli.md`, `concepts.md`, `custom-files.md`, `formats.md`, `quickstart.md`). None are referenced in `station/INDEX.md` Document Registry. Consider whether the full `docs/` directory warrants a registry entry or just the agent-interface contract.
- `internal/nonint/` has its own test suite (`config_test.go`, `contract_test.go`, `events_test.go`, `result_test.go`, `runner_test.go`, `update_test.go`). If Plan 42 (MCP server) ships, code-index coverage of this package will matter more.
- Plan 42 (MCP server) is the expected next-major plan. When it ships, check: does `station/INDEX.md` CLI commands count need to grow? Does the architecture diagram need an `mcp/` layer?
