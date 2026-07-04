---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-04
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
- **Duration:** ~8 min
- **Files Read:** 10 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/CLAUDE.md`, `/home/user/Bonsai/internal/generate/list_snapshot.go`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read git history (commits since 2026-05-04 — only 1 commit in last 7 days: `f7730ec` backlog-hygiene run). Reviewed Status.md to identify plans shipped since last doc-freshness run (Plans 40 and 41). Scanned project structure via `ls` on cmd/, internal/, internal/tui/, internal/generate/, docs/.
- **Result:** Two major plans shipped since last run: Plan 40 (Odysseus integration — schemas, validate pass, memory-routing docs + `docs/` dir) and Plan 41 (Headless CLI Contract — `internal/nonint/` package, `list --json`, `internal/generate/list_snapshot.go`, per-agent variants in generate.go). These introduced new packages and files not yet reflected in docs.
- **Issues:** Multiple docs have drifted. See Findings Summary.

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` in full. Verified tech stack, architecture diagram, CLI command count (8), catalog items (~50), agent types (6), document registry.
- **Result:** Tech stack accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template). Architecture overview accurate. CLI count 8 is still correct (completion is a meta-command, not a primary user-facing command). Catalog item count "~50" is approximate and still in range. Agent types 6 is correct.
- **Issues:** Low — `docs/agent-interface.md` (created in Plan 41) is not listed in the Document Registry table. It's an external-facing contract document that agents/CI scripts use.

### Step 3: Check navigation links
- **Action:** Verified all 42 links in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, and External References sections). Used bash loop to check file existence for each path. Also verified all 10 sensor links.
- **Result:** All 52 links resolve to real files. No broken navigation links.
- **Issues:** None — clean.

### Step 4: Report findings
- **Action:** Compiled findings across root CLAUDE.md, code-index.md, and INDEX.md. Categorized by severity. All findings flagged for user decision (no doc edits made per autonomous-run policy).
- **Result:** 3 documents with drift identified. 4 specific findings. Root CLAUDE.md has structural drift (high, recurring). code-index.md has line-number drift plus missing package (medium). INDEX.md has one missing entry (low).
- **Issues:** None in report generation.

### Step 5: Update dashboard
- **Action:** Updated dashboard row for "Doc Freshness Check" in `station/agent/Core/routines.md`.
- **Result:** Last Ran → 2026-07-04, Next Due → 2026-07-11, Status → done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | Root `CLAUDE.md` project structure tree missing `cmd/completion.go`, entire `internal/nonint/` package (Plan 41 headless CLI), `internal/generate/list_snapshot.go`, and OS-split files `catalog_snapshot_unix.go` / `catalog_snapshot_windows.go` | `/home/user/Bonsai/CLAUDE.md` lines 26–53 | Flagged for user — no edit (recurring, 3rd consecutive cycle) |
| 2 | MEDIUM | `code-index.md` generate.go section: all ~19 line number references have drifted by +40–100 lines each due to Plan 41 additions. Actual: `Scaffolding()` is at :401 (index says :360), `AgentWorkspace()` at :1460 (says :1359), etc. | `station/code-index.md` generate.go section | Flagged for user — no edit |
| 3 | MEDIUM | `code-index.md` missing `internal/nonint/` package section entirely — Plan 41's headless CLI contract layer. Key public API: `LoadConfig()`, `LoadOverlay()`, `EmitJSONL()`, `EmitFile()`, `EmitSummary()`, `RunRemoveAgent()`, `RunRemoveItem()`. Also missing: `internal/generate/list_snapshot.go` (`ListSnapshot` type, `SerializeJSON()`) and `cmd/completion.go` command. | `station/code-index.md` | Flagged for user — no edit |
| 4 | LOW | `station/INDEX.md` Document Registry table missing `docs/agent-interface.md` — the headless CLI contract document created by Plan 41 that AI agents and CI scripts use directly | `station/INDEX.md` lines 39–53 | Flagged for user — no edit |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**Finding 1 — HIGH: Root CLAUDE.md project structure drift (recurring)**

`/home/user/Bonsai/CLAUDE.md` lines 26–53: The project structure tree needs updating to reflect:
- Add `cmd/completion.go` — `bonsai completion [bash|zsh|fish|powershell]` (PR #78, external contribution)
- Add `internal/nonint/` package block — headless CLI contract (Plan 41): `config.go`, `events.go`, `nonint.go`, `remove.go`, `result.go`, `runner.go`, `update.go` + test files
- Add `internal/generate/list_snapshot.go` and `catalog_snapshot_unix.go` / `catalog_snapshot_windows.go` to the generate/ block
- Note: This is the 3rd consecutive cycle this has been flagged. Recommend promoting to P1 Backlog if not already a P2.

**Finding 2 — MEDIUM: code-index.md generate.go line numbers all drifted**

Every line number reference in the generate.go section of `station/code-index.md` is stale. Confirmed actual values (selected):

| Function | Index says | Actual |
|----------|-----------|--------|
| `Scaffolding()` | :360 | :401 |
| `SettingsJSON()` | :473 | :564 |
| `WorkspaceClaudeMD()` | :725 | :826 |
| `AgentWorkspace()` | :1359 | :1460 |
| `RoutineDashboard()` | :1010 | :1111 |
| `EnsureRoutineCheckSensor()` | :972 | :1073 |
| `PathScopedRules()` | :1164 | :1265 |
| `WorkflowSkills()` | :1228 | :1329 |
| `howToWorkLines()` | :575 | :676 |
| `quickTriggersLines()` | :683 | :784 |

Also: three new per-agent variants added by Plan 41 are undocumented: `SettingsJSONForAgent()` (:578), `PathScopedRulesForAgent()` (:1278), `WorkflowSkillsForAgent()` (:1338).

**Finding 3 — MEDIUM: code-index.md missing nonint package + list_snapshot + completion**

`station/code-index.md` needs a new section for `internal/nonint/` (Plan 41 headless CLI contract layer), an entry for `internal/generate/list_snapshot.go` (`ListSnapshot` + `SerializeJSON()`), and a row in the CLI Commands table for `cmd/completion.go` (`bonsai completion`).

**Finding 4 — LOW: INDEX.md missing docs/agent-interface.md in Document Registry**

Add a row to the Document Registry table: `docs/agent-interface.md` | Headless CLI contract — JSONL event format, exit codes, `--from-config` spec | When driving Bonsai headlessly from CI or an agent script.

## Notes for Next Run

- Root CLAUDE.md project-structure drift is a persistent recurring issue (3rd cycle in a row). This doc is outside the station/ workspace scope boundary so it's not auto-updated by any sensor. Consider whether a root-CLAUDE.md update should be included in the standard doc-refresh-bundle plan process.
- code-index.md line numbers drift whenever generate.go grows significantly. Line number references age out fast. If the index is primarily used for navigation, consider whether approximate "section" references or function-name-only entries would be more durable than exact line numbers.
- All station/CLAUDE.md navigation links are clean — no broken links for the first time in several cycles.
- Roadmap.md is accurate: Phase 1 complete, Phase 2 partially done, Phase 3 and 4 items appropriate.
