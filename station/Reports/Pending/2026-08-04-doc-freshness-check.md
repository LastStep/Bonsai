---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-04
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
- **Files Read:** 15 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/CLAUDE.md` (via system-reminder), `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`, `/home/user/Bonsai/station/agent/Skills/critic-agent-prompts.md`, `/home/user/Bonsai/station/agent/Skills/bubbletea.md`, `/home/user/Bonsai/internal/nonint/nonint.go`, `/home/user/Bonsai/internal/nonint/result.go`, `/home/user/Bonsai/internal/nonint/runner.go`, `/home/user/Bonsai/cmd/completion.go`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, grep, test, find, head)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read station/INDEX.md, station/Playbook/Status.md, station/Playbook/Roadmap.md. Ran `git log --since="7 days ago"` (empty — last commit 2026-06-16, 49 days ago). Checked broader git history (`git log --oneline -30`) to assess what has shipped since last run (2026-05-04).
- **Result:** Major work since last run: Plan 41 (Headless CLI Contract, all 5 phases shipped 2026-06-16) added `internal/nonint` package, headless cores for all mutating commands, `--json`/`--yes`/`--from`/`--non-interactive`/`--skip-conflicts` flags, `docs/agent-interface.md` contract, JSONL/exit contract (`ExitConflict=5`). Plan 40 Phases 1–3 also shipped. `bonsai completion` command added via PR #78 (2026-05-07). None of these are reflected in code-index.md.
- **Issues:** Multiple documentation gaps identified — see findings below.

### Step 2: Check INDEX.md accuracy
- **Action:** Read INDEX.md and cross-referenced Tech Stack table and Key Metrics against the codebase.
- **Result:** Tech stack is accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea). Agent types count (6) is correct. Catalog items (~50) is approximately correct (53 items: 18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). CLI commands count says "8" — but `bonsai completion` (PR #78, 2026-05-07) was added, making it 9 commands. The list in the metric still reads "(init, add, remove, list, catalog, update, guide, validate)" without "completion".
- **Issues:** CLI command count is stale (8 → 9); "completion" missing from the list.

### Step 3: Check navigation links
- **Action:** Enumerated all files in `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/`, `agent/Routines/`, `agent/Sensors/` and compared against links in station/CLAUDE.md navigation tables.
- **Result:** All links in CLAUDE.md resolve to real files. However, two files exist in the workspace that are NOT linked in the navigation table:
  1. `agent/Workflows/plan-grilling.md` — active workflow (used for Plans 40 and 41), not listed in Workflows table.
  2. `agent/Skills/critic-agent-prompts.md` — companion skill consumed by plan-grilling, not listed in Skills table.
  The `agent/Skills/bubbletea/` subdirectory files (components.md, emoji-width-fix.md, golden-rules.md, troubleshooting.md) are referenced from within bubbletea.md itself and do not require top-level navigation entries.
- **Issues:** 2 unlisted files in navigation; plan-grilling trigger phrases ("grill the plan", "review plan NN", "critic pass") also absent from CLAUDE.md Quick Triggers table.

### Step 4: Report findings
- **Action:** Compiled all drift findings with severity ratings and proposed updates.
- **Result:** 4 distinct findings documented below. Per procedure, no edits executed — all flagged for user decision.
- **Issues:** None procedural.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for Doc Freshness Check — `Last Ran` set to 2026-08-04, `Next Due` set to 2026-08-11, `Status` set to `done`.
- **Result:** Dashboard updated.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `internal/nonint` package entirely missing from code-index.md — Plan 41 (2026-06-16) added a new package with 7 files and significant public API (Result, RunInit, RunAdd, RunRemoveAgent, RunRemoveItem, ExitConflict, EmitJSONL). Developer has no navigation reference for this package. | `station/code-index.md` | Flagged for user — proposed update below |
| 2 | Medium | `plan-grilling.md` exists in `agent/Workflows/` but is absent from the CLAUDE.md Workflows navigation table. Active workflow (used for Plans 40 and 41). Trigger phrases defined in the file but not in Quick Triggers. | `station/CLAUDE.md` (Workflows table + Quick Triggers) | Flagged for user — proposed update below |
| 3 | Medium | `critic-agent-prompts.md` exists in `agent/Skills/` but is absent from the CLAUDE.md Skills navigation table. Companion to plan-grilling; currently undiscoverable via navigation. | `station/CLAUDE.md` (Skills table) | Flagged for user — proposed update below |
| 4 | Low | `bonsai completion` command (PR #78, merged 2026-05-07) not in code-index.md CLI commands table. INDEX.md Key Metrics still says "8 CLI commands (init, add, remove, list, catalog, update, guide, validate)" — should be 9 with "completion" added. | `station/code-index.md`, `station/INDEX.md` | Flagged for user — proposed update below |

---

## Proposed Updates

### Finding 1 — Add `internal/nonint/` section to code-index.md

Add a new section after `internal/wsvalidate/`:

```markdown
## Non-Interactive / Headless Core (`internal/nonint/`) — Plan 41

Pure headless implementations of all mutating commands. No TUI dependencies — safe for MCP/CI use.
Exit codes: ExitOK=0, ExitInvalidConfig=1, ExitRuntime=2, ExitWrongCWDForInit=3, ExitConflict=5.

| Type / Function | File | Purpose |
|-----------------|------|---------|
| `Result` | `result.go` | Structured return value — file operation counts + JSONL event list |
| `Result.Counts()` | `result.go` | Decompose Result into (created, updated, unchanged, skipped, conflicts) |
| `RunInit()` | `runner.go` | Headless `bonsai init` orchestrator |
| `RunAdd()` | `runner.go` | Headless `bonsai add` orchestrator (overlay-merge strategy) |
| `RunRemoveAgent()` | `remove.go` | Headless `bonsai remove <agent>` core |
| `RunRemoveItem()` | `remove.go` | Headless `bonsai remove <type> <name>` core |
| `EmitJSONL()` | `events.go` | Serialize Result to JSONL on provided writer |
| `KindFor()` | `remove.go` | Map singular item-type string to `ItemKind` |
| `LoadConfig()` | `config.go` | Load + validate `.bonsai.yaml` for headless use |
```

### Finding 2 — Add plan-grilling to CLAUDE.md Workflows table and Quick Triggers

Add to Workflows table:
```
| Adversarially reviewing a drafted plan with 6 critic agents before dispatch; Running a "grill the plan" / critic pass on a Tier-2 plan | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

Add to Quick Triggers table:
```
| Adversarially reviewing a plan before dispatch | "grill the plan", "review plan NN", "critic pass", "team of agents review this" |
```

### Finding 3 — Add critic-agent-prompts to CLAUDE.md Skills table

Add to Skills table:
```
| Verbatim dispatch prompts for the 6 plan-grilling critic agents. Load when dispatching critic agents via plan-grilling.md. | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

### Finding 4 — Update code-index.md and INDEX.md for completion command

In code-index.md CLI commands table, add:
```
| `bonsai completion` | `cmd/completion.go` | Generate shell completion scripts (bash/zsh/fish/powershell) |
```

In INDEX.md Key Metrics table, update:
```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **Findings 1–4 above** all require user approval before executing changes to CLAUDE.md, code-index.md, and INDEX.md. Proposed updates are in the section above — any can be applied independently.
- **Plan 41 in Plans/Active/** — 41-headless-cli-contract.md is still in Active/ despite being fully shipped. Already flagged by Status Hygiene routine (2026-08-04) as a known Backlog P2 item. No new action needed here unless the user wants to prioritize archiving.

---

## Notes for Next Run

- Check whether the `docs/agent-interface.md` contract doc (created Plan 41) should be listed in the INDEX.md Document Registry.
- Verify code-index.md accurately reflects headless command helpers (`runRemoveAgentNonInteractive`, `runRemoveItemNonInteractive`, `runUpdateNonInteractive`) added to existing cmd files as part of Plan 41.
- If Plan 42 (MCP server) ships before the next run, the architecture overview in INDEX.md will need a new layer added.
