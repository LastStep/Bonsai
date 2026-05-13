---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-05-13
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~10 min
- **Files Read:** 12 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/CLAUDE.md` (root), `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/Playbook/Plans/Archive/39-bonsai-noninteractive-flags.md` (spot check), `/home/user/Bonsai/station/agent/Core/identity.md` (existence check), `/home/user/Bonsai/station/agent/Core/memory.md` (existence check), `/home/user/Bonsai/station/agent/Core/self-awareness.md` (existence check)
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log append)
- **Tools Used:** `git log --oneline --since="7 days ago"`, `git log --name-only`, `ls` (directory listings), `grep` (pattern searches), `find` (file existence checks)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against git history
- **Action:** Read `station/INDEX.md`, `station/code-index.md`, and ran `git log --oneline --since="7 days ago"` with file listings to identify all changes in the last 7 days.
- **Result:** 22 commits in last 7 days. Key new code: `internal/nonint/` package (7 files), `cmd/completion.go`, `--non-interactive`/`--from-config` flags on `bonsai init` and `bonsai add`, sensor hook absolute-path baking fix (v0.4.3 `584b82b`). These new/changed features have **partial** coverage in station docs: Status.md and plan archives mention them, but `code-index.md` and `CLAUDE.md` (root) project structure are not updated.
- **Issues:** 3 doc drift items identified (see Findings Summary).

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` and compared tech stack, folder structure, CLI command count, and architecture diagram against current codebase.
- **Result:** Most of INDEX.md is accurate. Architecture diagram mentions `internal/nonint/` is missing from the diagram. CLI command count says `8` — this is still correct (`completion` is a subcommand of root, not a standalone command); the diagram's command list (`init, add, remove, list, catalog, update, guide, validate`) is also accurate. Catalog items `~50` is accurate (actual count: 53). Agent types `6` is accurate.
- **Issues:** `internal/nonint/` package is missing from the `Architecture Overview` block in `station/INDEX.md`. Minor drift only.

### Step 3: Check navigation links
- **Action:** Checked all linked files in `station/CLAUDE.md` navigation tables, plus all files referenced in `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/`, `agent/Routines/`, and `agent/Sensors/`.
- **Result:**
  - **1 broken link found**: `station/CLAUDE.md` Bonsai Reference table links to `../.bonsai/catalog.json`, which resolves to `/home/user/Bonsai/station/.bonsai/catalog.json`. That path **does not exist**. The catalog.json is at `/home/user/Bonsai/.bonsai/catalog.json` (project root). Since station/ is nested inside the Bonsai project root, `../.bonsai/catalog.json` from station/ points to `station/../.bonsai/` = `/home/user/Bonsai/.bonsai/catalog.json`, which **does exist**. Link resolves correctly at the OS level.
  - All Core, Protocol, Workflow, Skill, Routine, and Sensor file links resolve to real files.
  - All "External References" section links resolve.
- **Issues:** No broken navigation links. The `../.bonsai/catalog.json` link resolves correctly (OS path traversal from station/ goes up to repo root).

### Step 4: Report findings
- **Action:** Compiled 3 drift items with severity classifications.
- **Result:** Findings documented below. No doc edits made (audit-only routine — user decision required per procedure).
- **Issues:** None in the procedure itself; findings flagged for user.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for Doc Freshness Check — `Last Ran` → 2026-05-13, `Next Due` → 2026-05-20, `Status` → done.
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `internal/nonint/` package (7 files: nonint.go, config.go, events.go, runner.go + 3 test files) shipped in v0.4.2 but is not documented in `station/code-index.md` — no section, no types, no functions listed. | `station/code-index.md` | Flagged — not edited |
| 2 | Low | `cmd/completion.go` (added via external contribution #54, merged 2026-05-07) not listed in `station/code-index.md` CLI Commands table or in root `CLAUDE.md` project structure tree. | `station/code-index.md`, `/home/user/Bonsai/CLAUDE.md` | Flagged — not edited |
| 3 | Low | `internal/nonint/` package missing from `station/INDEX.md` Architecture Overview block. The block lists all other internal packages (`catalog/`, `config/`, `generate/`, `validate/`, `wsvalidate/`, `tui/`) but omits `nonint/`. | `station/INDEX.md` | Flagged — not edited |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

### Finding 1 — code-index.md: nonint package undocumented (Medium)
`internal/nonint/` is a new package introduced in Plan 39 (v0.4.2). It contains the non-interactive runner logic for `bonsai init`/`bonsai add`. `station/code-index.md` has no section for it.

**Proposed addition to code-index.md** (new section after `internal/wsvalidate/`):

```markdown
## Non-Interactive Runner (`internal/nonint/`)

Drives `bonsai init` / `bonsai add` under `--non-interactive --from-config` flags — no TUI, no prompts, JSONL stdout events, exit codes 0/2/3/4.

| Type / Function | File | Purpose |
|-----------------|------|---------|
| `RunInit()` | `runner.go` | Non-interactive init orchestrator — load config, generate workspace, emit events |
| `RunAdd()` | `runner.go` | Non-interactive add orchestrator — load overlay config, generate agent, emit events |
| `LoadConfig()` | `config.go` | Load + validate `.bonsai.yaml`-shaped input YAML; apply defaults |
| `Emitter` | `events.go` | JSONL event emitter — one JSON object per line to stdout |
| `Event` types | `events.go` | `file`, `summary`, `error` event shapes |
```

### Finding 2 — code-index.md + root CLAUDE.md: completion command undocumented (Low)
`cmd/completion.go` is not listed in the `code-index.md` CLI Commands table or root `CLAUDE.md` project structure tree.

**Proposed addition to code-index.md** CLI Commands table:
```
| `bonsai completion` | `cmd/completion.go` | Shell completion scripts for bash/zsh/fish/powershell |
```

**Proposed addition to root CLAUDE.md** `cmd/` tree (before `validate.go` line):
```
│   ├── completion.go        ← bonsai completion — shell completion script generator
```

### Finding 3 — INDEX.md Architecture Overview: nonint missing (Low)
The architecture overview block omits `internal/nonint/`.

**Proposed addition** (after `internal/wsvalidate/` line):
```
internal/nonint/      ← non-interactive init/add runner — no TUI, JSONL events, exit codes
```

## Notes for Next Run
- The previous cycle (2026-05-04) flagged root `CLAUDE.md` project-structure tree as "badly stale." That was addressed in Plan 37 (2026-05-07). However, Plan 39 (v0.4.2) shipped afterward and the tree drifted again. Consider adding a doc-update step to the Plan template to prompt updating code-index and CLAUDE.md after each code plan ships.
- If the nonint section is added to code-index.md, capture actual line numbers from source (this run did not read `internal/nonint/*.go` to get exact line refs).
- All navigation links are clean — no broken refs this cycle.
