---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-11
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 11 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/CLAUDE.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/CLAUDE.md` (via system-reminder), plus git log and filesystem checks
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls, file existence checks, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Ran `git log --since="7 days ago"` and `git log --since="2026-08-04"` to get commits from the last 7 days; read station/INDEX.md, Playbook/Status.md, Playbook/Roadmap.md.
- **Result:** Only 3 commits in the last 7 days, all routine-related (memory-consolidation, status-hygiene, backlog-hygiene). No source code or catalog changes in the window. Widened comparison to recent shipped features (Plan 41, 2026-06-16; `completion` command, 2026-05-07) that were flagged as not yet reflected in docs.
- **Issues:** Two features shipped since the last doc freshness run (2026-05-04) have not been captured in documentation: (1) `bonsai completion` command (PR #78, 2026-05-07); (2) `internal/nonint/` package from Plan 41 (2026-06-16).

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md`, verified Tech Stack table, Key Metrics table, Architecture Overview, and Document Registry against actual codebase state (`ls cmd/`, `ls internal/`, agent count).
- **Result:**
  - **Tech stack:** Accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea all correct.
  - **Agent types:** 6 confirmed — backend, devops, frontend, fullstack, security, tech-lead. Matches INDEX.md.
  - **Catalog items:** INDEX.md says "~50". Actual count: 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). Within the `~` approximation.
  - **CLI commands:** INDEX.md says "8 (init, add, remove, list, catalog, update, guide, validate)". Actual: 9 — `completion` was added (PR #78, shipped 2026-05-07) and is present at `cmd/completion.go` but not listed.
  - **Document Registry:** `docs/agent-interface.md` (Plan 41, shipped 2026-06-16) is a significant new contract document not listed in the registry.
  - **Architecture diagram:** The diagram's `internal/` section does not show `internal/nonint/` which exists and is the headless CLI package from Plan 41.
- **Issues:** 3 drift items found — CLI command count/list, missing `docs/agent-interface.md` from registry, missing `internal/nonint/` from architecture diagram.

### Step 3: Check navigation links
- **Action:** Verified all links in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References sections) by testing each path with `[ -e "$f" ]` shell check — 48 total paths verified.
- **Result:** All 48 links resolve to real files. Zero broken links.
- **Issues:** None.

### Step 4: Check root CLAUDE.md for drift
- **Action:** Read `/home/user/Bonsai/CLAUDE.md` project structure tree; cross-referenced with `ls cmd/` and `ls internal/`.
- **Result:**
  - `cmd/completion.go` exists but is NOT listed in the root CLAUDE.md's cmd/ structure tree (ends at `validate.go`).
  - `internal/nonint/` package exists (14 files: nonint.go, config.go, events.go, runner.go, result.go, update.go, remove.go, and corresponding tests + contract_test.go) but is completely absent from the root CLAUDE.md's `internal/` section.
- **Issues:** 2 drift items found in root CLAUDE.md.

### Step 5: Check code-index.md accuracy
- **Action:** Read `station/code-index.md` fully; compared CLI Commands table and internal package sections against current `ls cmd/` and `ls internal/`.
- **Result:**
  - CLI Commands table lists 8 commands — `completion` is absent.
  - Sections for catalog, config, generate, validate, wsvalidate, tui all present.
  - No section for `internal/nonint/` package (added Plan 41, the headless CLI contract layer).
- **Issues:** 2 drift items found in code-index.md.

### Step 6: Report findings and propose updates
- **Action:** Compiled all findings; proposals for each are in Findings Summary below.
- **Result:** 6 total drift items across 4 files; no doc edits made (audit-only per procedure).
- **Issues:** None.

### Step 7: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-08-11, Next Due → 2026-08-18, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package (14 files, Plan 41 headless CLI layer) completely absent from internal/ structure | `CLAUDE.md` (root), line ~37–60 internal/ block | Flagged — user to add `nonint/` entry between `generate/` and `validate/` |
| 2 | HIGH | `internal/nonint/` package has no section in code index | `station/code-index.md` | Flagged — user to add nonint section after wsvalidate |
| 3 | MEDIUM | `bonsai completion` command (PR #78, 2026-05-07) absent from cmd/ tree | `CLAUDE.md` (root), line ~36 | Flagged — add `completion.go` entry after `validate.go` (or before it as it's alphabetically between `catalog.go` and `guide.go`) |
| 4 | MEDIUM | CLI command count/list says 8 (omits `completion`) | `station/INDEX.md`, Key Metrics table | Flagged — update count to 9, add `completion` to list |
| 5 | MEDIUM | `bonsai completion` command absent from CLI Commands table | `station/code-index.md`, CLI Commands section | Flagged — add row: `bonsai completion \| cmd/completion.go \| runs Cobra built-in completion for bash/zsh/fish/powershell` |
| 6 | LOW | `docs/agent-interface.md` (Plan 41 headless contract, 2026-06-16) not in Document Registry | `station/INDEX.md`, Document Registry table | Flagged — user to decide whether to add it; it's a contract doc for AI integrators |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

**Finding 1 + 2 (HIGH) — `internal/nonint/` missing from docs:**
The `internal/nonint/` package is the headless CLI layer shipped in Plan 41 (2026-06-16). It contains the `RunInit`, `RunAdd`, `RunUpdate`, `RunRemove` headless cores, the JSONL event emitter, the `runner.go` exit-code constants (source of truth for `ExitConflict=5`), and the cross-command contract tests. Two docs are missing it:
- Root `CLAUDE.md` internal/ tree: add `nonint/` block between `generate/` and `validate/` — key files: `nonint.go` (public API), `events.go` (JSONL emitter), `runner.go` (exit-code constants), `result.go` (result types)
- `station/code-index.md`: add a new `## Nonint (internal/nonint/)` section

**Finding 3 (MEDIUM) — `completion` command missing from root CLAUDE.md:**
`cmd/completion.go` exists and was shipped in PR #78 (external contribution from @mvanhorn, 2026-05-07). Root CLAUDE.md's cmd/ tree ends at `validate.go` with no `completion.go` entry. Proposed fix: add `│   ├── completion.go        ← bonsai completion — shell completion (bash/zsh/fish/powershell)` after `guide.go`.

**Finding 4 + 5 (MEDIUM) — `completion` missing from INDEX.md and code-index.md:**
- INDEX.md Key Metrics: "CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate)" → should be "9 (init, add, remove, list, catalog, update, guide, validate, completion)"
- code-index.md CLI Commands table: add row `bonsai completion | cmd/completion.go:... | Cobra built-in shell completion for bash/zsh/fish/powershell`

**Finding 6 (LOW) — `docs/agent-interface.md` not in Document Registry:**
Plan 41 Phase 5 shipped `docs/agent-interface.md` as the canonical headless CLI contract (per-command flags, JSONL serialization shapes, exit-code table, stream discipline). It is useful for AI integrators and anyone building Plan 42 (MCP server). User to decide whether to add it to station/INDEX.md Document Registry.

---

## Notes for Next Run

- The `completion` command drift is recurring — it was added 2026-05-07 and was missed in the 2026-05-04 run (run was before the merge) and in this run's 7-day window. Check that root CLAUDE.md and code-index.md are updated before the next run.
- The `internal/nonint/` package was shipped 2026-06-16 — nearly 2 months before this check. This is the longest-running undocumented addition.
- Previous run (2026-05-04) had 5 drift items; at least 3 of them were resolved (bonsai-model.md nav link, code-index.md TUI packages, root CLAUDE.md TUI structure tree). Navigation links are now fully clean.
- Catalog item count is 53. INDEX.md's "~50" is within the approximation but update if it diverges further.
