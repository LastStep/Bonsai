---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-25
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
- **Files Read:** 9 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/CLAUDE.md`, `CLAUDE.md` (root), `station/agent/Core/memory.md`, `station/code-index.md`, `station/Playbook/Status.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history

Ran `git log --oneline --since="2026-05-04"` — found **50 commits** since the last doc freshness check (2026-05-04). Key code-changing commits:

- **Plan 40 (v0.5.0) — Phases 1–3:** Freeze schemas, root-relative scaffolding, validate pass, memory-routing protocol + guide formats. Landed: PRs #114, #115, #116.
- **Plan 41 — Headless CLI Contract + MCP-ready cores:** 5 phases — `*Result` shapes + JSONL/exit contract, headless update/remove/list, agent-interface contract doc. Landed: PRs #120–#125.
- **PR #54 — Explicit `completion` command** (visible in `bonsai --help`).
- **PR #106 — v0.4.3** sensor hook absolute path bake.
- **PR #123 / #122 / #121 / #120 / #125** — headless cores for remove, update, list, nonint, contract.

New packages added since last check:
- `internal/nonint/` (11+ files) — headless CLI cores for all mutating commands (Plan 40/41)
- `internal/generate/list_snapshot.go` — Plan 41 list --json support
- `internal/generate/catalog_snapshot_unix.go` + `catalog_snapshot_windows.go` — split from v0.4.0 hotfix
- `cmd/completion.go` — explicit shell completion command

### Step 2 — Check INDEX.md accuracy

Read `station/INDEX.md`. Checked Key Metrics and Architecture Overview against current filesystem state.

**Actual counts:**
- Skills: 18, Workflows: 10, Protocols: 4, Sensors: 13, Routines: 8, Agents: 6 → Total: **59 catalog items**
- CLI commands: `init, add, remove, list, catalog, update, guide, validate, completion` = **9 user-facing commands**
- Internal packages: `catalog, config, generate, validate, wsvalidate, tui, nonint` = **7 packages** (nonint is new)

INDEX.md says ~50 catalog items and 8 CLI commands; both are stale. Architecture diagram omits `internal/nonint/`.

### Step 3 — Check navigation links

Verified all links in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References). All **30 links resolve** — no broken links found.

Also spot-checked Core/, Protocols/, Workflows/, Skills/ — no broken internal links detected.

### Step 4 — Report findings

Findings compiled below. Per procedure, updates are flagged for user decision — no documentation edits executed.

### Step 5 — Update dashboard

Updated `station/agent/Core/routines.md` dashboard row for Doc Freshness Check: Last Ran → 2026-06-25, Next Due → 2026-07-02, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | Root `CLAUDE.md` project-structure tree missing `cmd/completion.go`, entire `internal/nonint/` package (11 files), `internal/generate/list_snapshot.go`, `internal/generate/catalog_snapshot_unix.go`, `catalog_snapshot_windows.go`, `catalog_snapshot_unix_test.go`, and 7 new test files under `cmd/` | `Bonsai/CLAUDE.md` lines 27–74 | Flagged for user review |
| 2 | MEDIUM | INDEX.md architecture diagram missing `internal/nonint/` layer — the headless CLI core package added in Plans 40/41 | `station/INDEX.md` lines 60–78 | Flagged for user review |
| 3 | MEDIUM | INDEX.md Key Metrics stale: "Catalog items ~50" should be ~59; "CLI commands 8" should be 9 (completion added in PR #54) | `station/INDEX.md` lines 32–33 | Flagged for user review |
| 4 | LOW | `code-index.md` missing entire `internal/nonint/` section (Plans 40/41 added 11+ files: `nonint.go`, `runner.go`, `events.go`, `result.go`, `config.go`, `update.go`, `remove.go`, plus tests + contract_test) | `station/code-index.md` | Flagged for user review |
| 5 | LOW | Plans/Active/ contains 2 shipped plans that memory.md Work State flags for archiving: `40-odysseus-platform-integration.md` (v0.5.0 shipped, tag-held) and `41-headless-cli-contract.md` (all 5 phases merged 2026-06-16) | `station/Playbook/Plans/Active/` | Flagged for user review |
| 6 | INFO | All navigation links in `station/CLAUDE.md` resolve — no broken links found | `station/CLAUDE.md` | No action needed |

---

## Proposed Updates (for user decision)

### Finding 1 — Root CLAUDE.md tree update (HIGH)

The project-structure tree under `## Project Structure` needs entries for:

```
cmd/
    ├── completion.go            ← explicit shell completion (bash/zsh/fish/powershell)
    ├── add_test.go              ← tests for bonsai add
    ├── add_nonint_test.go       ← headless-mode integration tests for add
    ├── catalog_test.go          ← tests for bonsai catalog
    ├── guide_test.go            ← tests for bonsai guide
    ├── init_nonint_test.go      ← headless-mode integration tests for init
    ├── list_test.go             ← tests for bonsai list
    ├── remove_nonint_test.go    ← headless-mode integration tests for remove
    ├── update_nonint_test.go    ← headless-mode integration tests for update
    └── validate_test.go         ← tests for bonsai validate

internal/nonint/                 ← headless CLI cores + MCP-ready contract (Plan 40/41)
    ├── nonint.go                ← shared non-interactive state + entry
    ├── runner.go                ← canonical exit-code source + JSONL event dispatcher
    ├── result.go                ← *Result shapes for init/add/update/remove
    ├── events.go                ← JSONL event types (file/summary/warning)
    ├── config.go                ← --from-config overlay parsing
    ├── update.go                ← headless update core
    ├── remove.go                ← headless remove core
    └── testdata/                ← golden JSONL fixtures

internal/generate/
    ├── catalog_snapshot_unix.go   ← Unix-specific openSnapshotFile (build tag !windows)
    ├── catalog_snapshot_windows.go ← Windows-specific openSnapshotFile
    └── list_snapshot.go           ← list --json snapshot renderer
```

### Finding 2 — INDEX.md architecture diagram (MEDIUM)

Add `internal/nonint/` row after `internal/wsvalidate/`:

```
internal/nonint/      ← headless CLI cores (init/add/update/remove) + JSONL/exit contract (Plan 41)
```

### Finding 3 — INDEX.md Key Metrics (MEDIUM)

Update:
- `Catalog items | ~50` → `~59 (skills, workflows, protocols, sensors, routines)`
- `CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate)` → `9 (+ completion)`

### Finding 4 — code-index.md nonint section (LOW)

Add a new `## Non-interactive / Headless (`internal/nonint/`) — Plan 41` section covering `nonint.go`, `runner.go`, `result.go`, `events.go`, `config.go`, `update.go`, `remove.go` with key types and functions.

### Finding 5 — Plan archive (LOW)

Move `station/Playbook/Plans/Active/40-odysseus-platform-integration.md` and `41-headless-cli-contract.md` to `station/Playbook/Plans/Archive/`. Already noted in memory.md Work State ("Plan 41 file still in Plans/Active/ — archive to Plans/Archive/ at next wrap-up").

---

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **Finding 1 (HIGH):** Root `Bonsai/CLAUDE.md` project-structure tree — significant drift from Plans 40/41 + PR #54. This is a recurring finding (also flagged 2026-05-04, 2026-04-21). The Backlog has a P2 item for a root-CLAUDE.md doc-freshness sub-step — this run confirms it still accumulates.
- **Finding 2 (MEDIUM):** INDEX.md arch diagram missing nonint package.
- **Finding 3 (MEDIUM):** INDEX.md Key Metrics — catalog count and CLI command count stale.
- **Finding 4 (LOW):** code-index.md missing nonint package section.
- **Finding 5 (LOW):** Plans/Active/ has 2 shipped plans pending archive (already flagged in memory.md).

## Notes for Next Run

- Root CLAUDE.md tree drift is a recurring P2 pattern (3rd consecutive cycle). Worth considering whether to bundle Findings 1+4 into a single doc-refresh Tier-1 dispatch (similar to Plan 37 doc-refresh-bundle) during the next routine-digest session.
- All nav links clean this cycle — no broken refs to chase.
- code-index.md is now materially behind (nonint = 11 files, no section). Should be bundled with root CLAUDE.md fix.
- Plan 40 is v0.5.0-ready but tag-held; plans sit in Active until tag decision. User is aware.
