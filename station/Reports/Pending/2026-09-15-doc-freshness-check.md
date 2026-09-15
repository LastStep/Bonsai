---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-15
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
- **Files Read:** 10
  - `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`
  - `/home/user/Bonsai/station/INDEX.md`
  - `/home/user/Bonsai/station/code-index.md`
  - `/home/user/Bonsai/station/CLAUDE.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/agent/Core/memory.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/.bonsai.yaml`
  - `/home/user/Bonsai/station/agent/Workflows/issue-to-implementation.md` (link target check)
  - `/home/user/Bonsai/station/agent/Skills/` (directory listing)
- **Files Modified:** 2
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** git log, find, grep, ls
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Ran `git log --oneline --after="2026-05-04"` to find all commits since last run. Reviewed commit messages and selected `--stat` for the most significant `feat`/`fix`/`docs` commits.
- **Result:** 20+ commits found. Key feature additions since last run:
  - **Plan 40 Phase 1** (`1e715c7`): `feat(generate)` — freeze schemas + root-relative scaffolding
  - **Plan 40 Phase 2** (`a540fdd`): `feat(validate)` — project-level validate pass
  - **Plan 40 Phase 3** (`2aef7fd`): `docs` — memory-routing protocol + guide Formats page (`docs/formats.md` added)
  - **Plan 41 Phase 1** (`65932b9`): `feat(nonint)` — Result reshape + shared exit/event contract; new `internal/nonint/` package
  - **Plan 41 Phase 2** (`5a4cf3c`): `feat(update)` — headless update core with `--non-interactive`/`--skip-conflicts`
  - **Plan 41 Phase 3** (`332884a`): `feat(remove)` — headless remove with `--yes`/`--from` flags + symlink safety
  - **Plan 41 Phase 4** (`6700848`): `feat(list)` — `list --json` + read-command JSON parity
  - **Plan 41 Phase 5** (`ab202c3`): `docs(contract)` — new `docs/agent-interface.md` (headless CLI contract), contract test sweep
  - **Completion command** (`2eae9d4`): `feat(cmd)` — explicit `completion` subcommand (bash/zsh/fish/powershell)
- **Issues:** Multiple doc drift items found — see findings below.

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md`, compared Tech Stack, Key Metrics, and Architecture Overview against current state.
- **Result:**
  - Tech Stack table: correct (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS)
  - Agent types count: correct (6)
  - **CLI commands count: STALE** — says "8 (init, add, remove, list, catalog, update, guide, validate)" but `completion` is now the 9th command
  - Catalog items: says "~50" — actual count is 53, within tolerance for `~50`
  - Architecture overview diagram: accurate (all packages listed are correct)
- **Issues:** CLI commands count is one behind.

### Step 3: Check navigation links
- **Action:** Extracted all `[...](...)` links from `station/CLAUDE.md` and agent subdirectory files. Resolved each relative link and checked for file existence.
- **Result:**
  - All links in `station/CLAUDE.md`: **all 49 targets resolve correctly** — no broken links
  - All links in `agent/Core/self-awareness.md`, `agent/Core/identity.md`: clean
  - **`agent/Core/memory.md` lines 87–92**: References `../../Research/RESEARCH-*.md` (6 files). Neither `station/Research/` nor `/home/user/Bonsai/Research/` exist anywhere in the repo. These links are broken.
  - **`agent/Workflows/issue-to-implementation.md`**: References `../Skills/dispatch.md` at lines 35, 175, and 204. File `station/agent/Skills/dispatch.md` does not exist. The `dispatch` skill IS present in the catalog (`catalog/skills/dispatch/`) but has never been installed via `bonsai add`.
  - `agent/Protocols/` — two "possibly broken" flags: `../../Playbook/Plans/Active/` (directory, valid) and `../Skills/` (directory, valid). False positives from the checker.
  - `agent/Workflows/` reference to `../../Playbook/Standards/NoteStandards.md`: file exists, valid.
- **Issues:** 2 broken-link clusters found (Research directory, dispatch.md).

### Step 4: Report findings
- **Action:** Compiled 5 findings below. Per procedure, not executing doc updates — flagging for user decision.
- **Result:** 5 findings documented; 3 require user decision.

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for "Doc Freshness Check".
- **Result:** `Last Ran` → 2026-09-15, `Next Due` → 2026-09-22, `Status` → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `internal/nonint/` package (Plan 41 — 13+ files, headless CLI contract infrastructure) is entirely absent from code index | `station/code-index.md` | Flagged for user — propose adding a new section |
| 2 | Low | `bonsai completion` command (`cmd/completion.go`) missing from CLI Commands table | `station/code-index.md` | Flagged for user — propose adding a row |
| 3 | Low | CLI command count says "8" but is now "9" (completion added) | `station/INDEX.md` Key Metrics table | Flagged for user — update count + list |
| 4 | Medium | 6 broken links to `../../Research/RESEARCH-*.md` — Research directory does not exist | `station/agent/Core/memory.md` lines 87–92 | Flagged for user — links cannot be auto-fixed (need to create Research dir or remove links) |
| 5 | Medium | 3 broken links to `../Skills/dispatch.md` — dispatch skill not installed; exists in catalog | `station/agent/Workflows/issue-to-implementation.md` lines 35, 175, 204 | Flagged for user — run `bonsai add` to install `dispatch` skill OR remove references |

---

## Proposed Updates (for user decision)

### Finding 1 — Add `internal/nonint/` section to code-index.md

Insert a new section after `## Workspace-path Validation` (or before `## TUI`):

```markdown
## Headless CLI / Non-interactive (`internal/nonint/`) — Plan 41

Core infrastructure for headless/CI-friendly operation. All mutating commands (init, add,
update, remove) emit JSONL to stdout; read commands (list, catalog, validate) emit single-doc
JSON. Diagnostics go to stderr. Exit codes are documented in `docs/agent-interface.md`.

| File | Purpose |
|------|---------|
| `nonint.go` | Package entry — shared types and interfaces |
| `config.go` | Per-command config structs for headless invocations |
| `events.go` | JSONL event types (file, summary, warning) |
| `result.go` | Result shapes + exit-code constants (ExitOK=0, ExitNoOp=2, ExitConflicts=3, ExitNotFound=4, ExitConflict=5) |
| `runner.go` | Canonical exit-code table + emit discipline — stdout vs stderr routing |
| `update.go` | Headless update core (`--non-interactive`, `--skip-conflicts`) |
| `remove.go` | Headless remove core (`--non-interactive`/`--yes`, `--from`, `--delete-files`) |
```

### Finding 2 — Add `bonsai completion` row to code-index.md CLI Commands table

Add to the CLI Commands table:
```
| `bonsai completion` | `cmd/completion.go:20` | `completionCmd` — generate shell completion for bash/zsh/fish/powershell |
```

### Finding 3 — Update INDEX.md Key Metrics CLI count

Change: `8 (init, add, remove, list, catalog, update, guide, validate)`
To: `9 (init, add, remove, list, catalog, update, guide, validate, completion)`

### Finding 4 — Research directory broken links

Options:
1. **Create the Research directory** and add the 6 RESEARCH-*.md placeholder files (these appear to be planned research artifacts — may be worth creating stubs)
2. **Remove the links from memory.md** lines 87–92 and replace with prose references (no clickable links until files exist)
3. **Leave as-is** if these links are intentional forward-references to planned work (least clean but zero effort)

### Finding 5 — dispatch.md broken links

Options:
1. **Install the dispatch skill**: run `bonsai add` and select `dispatch` from Skills for the tech-lead agent
2. **Remove the three references** in `issue-to-implementation.md` if dispatch is intentionally not installed
3. **Create `station/agent/Skills/dispatch.md`** as a custom file (same content as `catalog/skills/dispatch/dispatch.md`)

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **Finding 4 (medium):** `station/agent/Core/memory.md` references 6 non-existent Research files. Likely planned research documents never created. Decide: create Research directory with stubs, remove links, or leave.
- **Finding 5 (medium):** `issue-to-implementation.md` references `dispatch.md` skill which is not installed. Workflow will send agents to a missing reference. Recommend: `bonsai add` → select `dispatch` skill, OR remove references from the workflow.
- **Finding 1 (medium):** `internal/nonint/` package from Plan 41 (full headless CLI contract layer) is not in `code-index.md`. When navigating the codebase for headless/CI work, this gap will cost time. Recommend updating code-index.md.

---

## Notes for Next Run

- The Research directory link issue has been present since at least this run — if still broken at next check, escalate priority.
- `dispatch` skill availability should be resolved before next run — verify it's installed or references are cleaned up.
- All 7 routines remain severely overdue (4+ month gap since May 2026). Next doc-freshness check in 7 days (2026-09-22).
- code-index.md line numbers for existing functions were not re-verified in this run (out of scope for drift detection, and would require full source reads). If code churn is high, consider a dedicated code-index refresh.
