---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-26
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
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/CLAUDE.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`, `station/Logs/RoutineLog.md`, `cmd/remove.go`, `cmd/update.go`, `cmd/list.go`, `cmd/completion.go`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, file existence checks, function listing)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read `station/CLAUDE.md`, `station/INDEX.md`, `station/code-index.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`. Reviewed git log for all commits since last run (2026-05-04). There were no commits in the last 7 days, but the gap since last run is ~83 days, so the full git log was examined. Key shipped work: Plan 40 (Phases 1–3, root-relative scaffolding + validate pass), Plan 41 (Headless CLI Contract + MCP-ready cores — all 5 phases, PRs #120/#122/#123/#121/#125, shipped 2026-06-16), `bonsai completion` command (PR #78, 2026-05-07).
- **Result:** Several documentation gaps identified — see Findings Summary.
- **Issues:** `code-index.md` is missing all new headless functions from Plan 41 and the `completion` command from PR #78.

### Step 2: Check INDEX.md accuracy
- **Action:** Compared `station/INDEX.md` Tech Stack, Key Metrics, Architecture Overview, and Document Registry against the actual codebase and recent commits.
- **Result:** Tech Stack is accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea). Architecture Overview is accurate — all packages (`internal/validate/`, `internal/wsvalidate/`) are present. Key Metrics "CLI commands: 8" is still correct if `bonsai completion` is counted as a separate subcommand rather than a top-level command (it wires under root). **One gap found:** the Document Registry does not list `docs/agent-interface.md`, which was created as part of Plan 41 as the machine-readable CLI contract.
- **Issues:** `docs/agent-interface.md` not in Document Registry.

### Step 3: Check navigation links
- **Action:** Verified all 50 navigation links in `station/CLAUDE.md` (Core, Protocols, Workflows, Skills, Routines, Sensors, External References tables). Checked links in the Routines table, Sensors table, and all external references.
- **Result:** All 50 links resolve to real files. Zero broken links found.
- **Issues:** None.

### Step 4: Report findings
- **Action:** Compiled 5 findings ranked by severity. All are flagged for user decision per procedure (no doc edits executed).
- **Result:** 5 findings: 2 medium (code-index drift), 1 low (missing doc registry entry), 1 low (stale plan in Active/), 1 info (Roadmap gap). See Findings Summary below.
- **Issues:** None — all findings are documentation drift, no broken navigation.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | `code-index.md` missing `bonsai completion` command — `completion.go` shipped in PR #78 (2026-05-07) but not listed in CLI Commands table | `station/code-index.md` | Flagged for user |
| 2 | medium | `code-index.md` missing 12+ headless/noninteractive functions from Plan 41 — Remove Helpers (`headlessRequested`, `loadConfigHeadless`, `runRemoveAgentNonInteractive`, `runRemoveItemNonInteractive`, `filterRequired`, `buildRemoveOptions`, `buildTargetOptions`, `resolveRemoveTargets`), Update Helpers (`runUpdateNonInteractive`, `isTerminal`), List (`renderListJSON`, `terminalSize`) | `station/code-index.md` | Flagged for user |
| 3 | low | `INDEX.md` Document Registry missing `docs/agent-interface.md` — the headless CLI + JSONL/exit contract document from Plan 41 | `station/INDEX.md` | Flagged for user |
| 4 | low | `Plans/Active/41-headless-cli-contract.md` should be in Archive — Plan 41 is fully shipped (all 5 phases merged, 2026-06-16) | `station/Playbook/Plans/Active/` | Flagged for user |
| 5 | info | Roadmap doesn't reflect headless CLI / MCP-ready cores milestone from Plan 41 — Phase 3 ("Managed Agents integration") is the goal but the foundational headless contract step is not checked off anywhere | `station/Playbook/Roadmap.md` | Flagged for user |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **code-index.md needs an update sweep** (Findings 1 & 2) — Two categories of missing entries. Could be dispatched as a Plan (similar to Plan 37 doc-refresh-bundle). The `completion` command entry is a one-liner; the headless functions section is a larger sweep across Remove Helpers, Update Helpers, and List sections.

2. **INDEX.md Document Registry** (Finding 3) — Add `docs/agent-interface.md` with a description like "Headless CLI + JSONL/exit contract — for AI agents and CI consumers" and a "When building MCP servers or CI integrations on top of Bonsai" usage note.

3. **Plan 41 archive move** (Finding 4) — Move `station/Playbook/Plans/Active/41-headless-cli-contract.md` to `station/Playbook/Plans/Archive/`. Plan 40 should stay in Active (Phase 4 is still HELD).

4. **Roadmap milestone** (Finding 5) — Consider adding a checked-off item under Phase 3 or Phase 2: "Headless CLI contract — `*Result` cores + JSONL/exit codes + `docs/agent-interface.md` (Plan 41)".

## Notes for Next Run

- All 50 navigation links in `station/CLAUDE.md` resolved cleanly — link check passes.
- The gap since last run was 83 days (should be 7). If loop.md dispatch is working, next run should be 2026-08-02.
- Backlog hygiene (same date) already flagged that HOMEBREW_TAP_TOKEN PAT is likely expired — relevant if a release is triggered before rotation.
- If Plan 37-style doc-refresh is dispatched to close Findings 1 & 2, the next doc freshness check should verify code-index line numbers are still accurate (Plan 41 shifted several function positions in `remove.go`).
