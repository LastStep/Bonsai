---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-28
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
- **Duration:** ~6 min
- **Files Read:** 8 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/agent/Skills/critic-agent-prompts.md` (frontmatter), `station/agent/Workflows/plan-grilling.md` (implicit via ls)
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, git diff, ls, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Ran `git log --since=2026-08-21` for the 7-day window; then extended to `--since=2026-04-28` because the last run was 2026-05-04 (3+ months ago — all changes since last run needed review). Reviewed `git diff --stat HEAD~10 HEAD` for changed files.
- **Result:** One significant commit in the 7-day window (`6082356 2026-08-28 routine: run backlog-hygiene`) plus a large body of changes from 2026-05-13 through 2026-06-16 not covered by the previous doc-freshness run. Key unrecorded changes: Plan 41 (headless CLI contract) shipped 2026-06-16 — added `internal/nonint/` package, new flags on remove/update/list, `docs/` folder, and `internal/generate/list_snapshot.go`.
- **Issues:** 5 documentation gaps found — all flagged below.

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` in full. Compared tech stack table, Key Metrics, Architecture Overview, and Document Registry against actual project state.
- **Result:** Tech stack table is accurate. Key Metrics (6 agent types, ~50 catalog items, 8 CLI commands) are still correct — Plan 41 added flags/options to existing commands, not new top-level commands. Architecture Overview diagram is accurate. **Gap found:** The Document Registry does not mention the `docs/` folder added by Plan 40/41, which now contains public-facing docs (quickstart, concepts, CLI reference, agent-interface contract, formats guide, custom-files guide).
- **Issues:** One finding — `docs/` folder absent from INDEX.md Document Registry.

### Step 3: Check navigation links
- **Action:** Listed actual files in `station/agent/Core/`, `station/agent/Protocols/`, `station/agent/Workflows/`, `station/agent/Skills/`, `station/agent/Sensors/`. Cross-referenced against all links in `station/CLAUDE.md`.
- **Result:**
  - All Core links resolve (identity.md, memory.md, self-awareness.md). ✓
  - All Protocol links resolve (memory.md, scope-boundaries.md, security.md, session-start.md). ✓
  - All Sensor links resolve (all 10 sensor .sh files present). ✓
  - **Gap 1:** `station/agent/Workflows/plan-grilling.md` exists on disk (added 2026-06-13) but has no entry in CLAUDE.md's Workflows nav table.
  - **Gap 2:** `station/agent/Skills/critic-agent-prompts.md` exists on disk but has no entry in CLAUDE.md's Skills nav table.
  - `station/agent/Skills/bubbletea/` directory exists alongside `bubbletea.md`; CLAUDE.md links to the .md file correctly. The directory contains component-level sub-docs — likely intentional (not a broken link).
- **Issues:** Two nav gaps found.

### Step 4: Check code-index.md accuracy
- **Action:** Searched code-index.md for references to `internal/nonint/`, `list_snapshot.go`. Verified `cmd/root.go` helper line numbers.
- **Result:**
  - **Gap 3 (HIGH):** `internal/nonint/` package is entirely absent from code-index.md. Plan 41 (2026-06-16) added this as a major new package providing the headless CLI contract — includes config.go, events.go, remove.go, result.go, runner.go, update.go. The package is the backbone of `--non-interactive` mode for init/add/remove/update and `--json` output for list.
  - **Gap 4 (MEDIUM):** `internal/generate/list_snapshot.go` is absent from code-index.md. Adds `SerializeJSON()` for `bonsai list --json`.
  - **Gap 5 (LOW):** `cmd/root.go` helper line numbers are off by 1 (code-index says `:45`, `:53`, `:64` for `loadCatalog`, `requireConfig`, `mustCwd`; actual are `:46`, `:54`, `:65`). Minor drift from Plan 41 edits.
- **Issues:** Three findings (1 high, 1 medium, 1 low).

### Step 5: Report findings and update dashboard
- **Action:** Wrote this report. Updated routines.md dashboard. Appended to RoutineLog.md.
- **Result:** Completed.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package entirely absent from code-index — Plan 41's headless CLI contract (--non-interactive/--json) undocumented | `station/code-index.md` | Flagged for user |
| 2 | MEDIUM | `internal/generate/list_snapshot.go` absent from code-index — `SerializeJSON()` for `bonsai list --json` undocumented | `station/code-index.md` | Flagged for user |
| 3 | MEDIUM | `plan-grilling.md` workflow exists in `agent/Workflows/` but has no CLAUDE.md nav entry — agent cannot discover it | `station/CLAUDE.md` (Workflows table) | Flagged for user |
| 4 | MEDIUM | `critic-agent-prompts.md` skill exists in `agent/Skills/` but has no CLAUDE.md nav entry — agent cannot discover it | `station/CLAUDE.md` (Skills table) | Flagged for user |
| 5 | MEDIUM | `docs/` folder (8 files) not in INDEX.md Document Registry — public docs invisible to agent nav | `station/INDEX.md` | Flagged for user |
| 6 | LOW | `cmd/root.go` helper line numbers off by 1 in code-index (`:45/53/64` vs actual `:46/54/65`) | `station/code-index.md` | Flagged for user |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

All 6 findings require user decision before updates are executed (per routine procedure: "Propose updates but don't execute — flag for user decision").

**Suggested actions:**

1. **(HIGH) Update code-index.md — add `internal/nonint/` section** covering: `nonint.go` (package entry), `runner.go` (core non-interactive runner), `events.go` (event/result contract), `result.go` (result shapes), `remove.go` (headless remove), `update.go` (headless update). Also add `config.go` if it holds shared types.

2. **(MEDIUM) Update code-index.md — add `list_snapshot.go` row** to the Generator section documenting `SerializeJSON()`.

3. **(MEDIUM) Add `plan-grilling.md` to CLAUDE.md Workflows table** — suggested trigger: "Running the 6-critic adversarial plan review pipeline; Stress-testing a plan draft before dispatch."

4. **(MEDIUM) Add `critic-agent-prompts.md` to CLAUDE.md Skills table** — suggested trigger: "Running the plan-grilling pipeline; Loading verbatim critic prompts for adversarial plan review."

5. **(MEDIUM) Add `docs/` folder to INDEX.md Document Registry** — entries for `docs/README.md` (public docs home), `docs/quickstart.md`, `docs/concepts.md`, `docs/cli.md`, `docs/formats.md`, `docs/agent-interface.md`, `docs/custom-files.md`.

6. **(LOW) Correct root.go line numbers in code-index.md** — change `:45` → `:46`, `:53` → `:54`, `:64` → `:65`.

---

## Notes for Next Run

- The gap between last run (2026-05-04) and now (2026-08-28) is 116 days — three months of drift accumulated. The 7-day frequency should be honored going forward.
- The `docs/` folder may be expanding with Plan 40 website work — check it again next run.
- `internal/nonint/` is an active package — re-check line numbers when code-index entry is added.
- The `bubbletea/` sub-directory in `agent/Skills/` was not flagged (no broken link) but worth confirming whether sub-doc files should be individually listed or linked from `bubbletea.md`.
