---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-16
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
- **Duration:** ~8 minutes
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Status.md`, `station/CLAUDE.md`, `station/code-index.md`, `CLAUDE.md` (root), `station/agent/Skills/critic-agent-prompts.md`, `station/agent/Workflows/plan-grilling.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log append)
- **Tools Used:** Read, Bash (git log, file existence checks, grep, ls)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history (last 7 days)
- **Action:** Ran `git log --oneline --since="7 days ago" --name-status` to enumerate all commits and changed files since 2026-06-09. Cross-referenced against `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`, and root `CLAUDE.md`.
- **Result:** 25 commits found in the window. Key changes from Plans 40/41 that alter the codebase significantly:
  - **Plan 41** (Headless CLI Contract): Added `internal/nonint/` package (8 files), `internal/generate/list_snapshot.go`, `docs/agent-interface.md`, `cmd/list.go` `--json` flag, headless remove/update cores in `cmd/remove.go` and `cmd/update.go`, `docs/formats.md` updated.
  - **Plan 40**: Added `internal/validate/project.go`, `catalog/scaffolding/MEMORY.md.tmpl`, `catalog/scaffolding/.bonsai/project.yaml.tmpl`.
  - Custom workspace items added (not from plans): `station/agent/Workflows/plan-grilling.md`, `station/agent/Skills/critic-agent-prompts.md`.
- **Issues:** Multiple documentation locations are stale relative to these changes (see Findings Summary).

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` Tech Stack table, Key Metrics table, Architecture Overview, and arch diagram. Cross-checked against actual codebase state.
- **Result:**
  - **Tech Stack:** Accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, single binary all correct.
  - **Key Metrics:** Agent types = 6 (correct). Catalog items = ~50 (actual count is 53: 18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines — mild drift, still within "~50" range). CLI commands = 8 (correct).
  - **Architecture diagram:** Does not mention `internal/nonint/` package (new in Plan 41) or `docs/` directory (new in Plans 40/41 with `agent-interface.md`, `formats.md`, etc.). Low-to-medium severity drift.
- **Issues:** INDEX.md arch section missing `internal/nonint/` and `docs/` entries. Flagged for user decision.

### Step 3: Check navigation links
- **Action:** Extracted all markdown link targets from `station/CLAUDE.md` and verified each file exists using a shell loop. Also checked `station/agent/Core/`, `station/agent/Protocols/`, `station/agent/Workflows/`, `station/agent/Skills/` for files not listed in nav tables.
- **Result:**
  - **All 46 navigation links in `station/CLAUDE.md` resolve to existing files** — clean.
  - **Unlisted files found:**
    - `station/agent/Workflows/plan-grilling.md` — exists but not listed in CLAUDE.md Workflows nav table.
    - `station/agent/Skills/critic-agent-prompts.md` — exists but not listed in CLAUDE.md Skills nav table.
    - `station/agent/Skills/bubbletea/` directory — exists alongside `bubbletea.md` (multi-file skill); nav correctly points to `bubbletea.md`.
- **Issues:** 2 unlisted workflow/skill files. Low severity — they are functional but invisible to nav-table routing.

### Step 4: Report findings
- **Action:** Compiled drift findings across root `CLAUDE.md`, `station/code-index.md`, `station/INDEX.md`, and `station/CLAUDE.md`. Categorized by severity.
- **Result:** 5 drift items identified (see Findings Summary). All flagged for user decision — no edits executed (audit-only per procedure).

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` — Doc Freshness Check row: `Last Ran` → 2026-06-16, `Next Due` → 2026-06-23, `Status` → done.
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | **HIGH** | Root `CLAUDE.md` project-structure tree missing `internal/nonint/` package (added Plan 41 — 8 files: headless cores for init/add/update/remove, Result shape, events, runner, config) | `CLAUDE.md` lines 37–74 (internal/ block) | Flagged — propose update |
| 2 | **MEDIUM** | `station/code-index.md` — `internal/nonint/` package entirely undocumented. No section exists for this new package. Also: `internal/generate/list_snapshot.go` undocumented, `internal/validate/project.go` undocumented, remove.go headless helpers (`runRemoveAgentNonInteractive`, `runRemoveItemNonInteractive`, `loadConfigHeadless`) missing, update.go `runUpdateNonInteractive` missing, list.go `renderListJSON` missing. Line-number drift: `runRemoveItem()` listed at `:290` but actual `:428`; `runRemoveItemAction()` listed at `:565` but actual `:703`. | `station/code-index.md` | Flagged — propose code-index refresh (Plan 37-class task) |
| 3 | **MEDIUM** | `station/INDEX.md` architecture diagram missing `internal/nonint/` and `docs/` directory. `docs/` contains `agent-interface.md`, `formats.md`, `quickstart.md`, `concepts.md`, `custom-files.md`, `cli.md`, `README.md` — new human-facing docs added in Plans 40/41. | `station/INDEX.md` lines 59–77 | Flagged — propose minimal arch section update |
| 4 | **LOW** | `station/CLAUDE.md` Workflows nav table missing `plan-grilling.md` (file exists at `agent/Workflows/plan-grilling.md`, added 2026-06-13, in-use per Plan 40 session log). | `station/CLAUDE.md` Workflows section | Flagged — easy 1-row addition |
| 5 | **LOW** | `station/CLAUDE.md` Skills nav table missing `critic-agent-prompts.md` (file exists at `agent/Skills/critic-agent-prompts.md`, added 2026-06-13, consumed by plan-grilling workflow). | `station/CLAUDE.md` Skills section | Flagged — easy 1-row addition |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

All 5 items are flagged for user decision (audit-only routine):

1. **[HIGH] Root CLAUDE.md internal/ tree drift** — Add `internal/nonint/` stanza. Straightforward — mirrors existing package blocks. No ambiguity. Suggested action: add as Backlog P2 or quick-fix next session.

2. **[MEDIUM] code-index.md stale** — Needs `internal/nonint/` section + `list_snapshot.go` row + `project.go` row + updated remove/update/list helper tables + corrected line numbers. This is a Plan-37-class doc refresh task (~30 min agent dispatch). Suggested action: add to Backlog as P2 doc debt, scope into next release prep cycle.

3. **[MEDIUM] INDEX.md arch section drift** — Missing `internal/nonint/` and `docs/` entries. Suggested action: quick-fix inline (2 lines in the arch diagram) — low blast radius.

4. **[LOW] station/CLAUDE.md missing plan-grilling workflow row** — One-liner addition to Workflows nav table. Suggested action: quick-fix inline or bundle with #5.

5. **[LOW] station/CLAUDE.md missing critic-agent-prompts skill row** — One-liner addition to Skills nav table. Suggested action: quick-fix inline or bundle with #4.

## Notes for Next Run

- Catalog items count drifted from ~50 to 53 (still within the "~50" approximation — not flagged as a separate finding but worth noting if count crosses 55+).
- `completion.go` in `cmd/` is not documented in `code-index.md` CLI commands table — this predates the current 7-day window (it was contributed by @mvanhorn, shipped 2026-05-07). Filed here for awareness; not a new drift.
- The pattern of `plan-grilling.md` and `critic-agent-prompts.md` being added directly to `station/` (not via `bonsai add`) means nav tables won't auto-update. Consider adding these to next `station/CLAUDE.md` nav table sweep.
- Root `CLAUDE.md` project-structure drift is a recurring finding (seen in 2026-04-21 and 2026-05-04 runs). Consider promoting the "root-CLAUDE.md check" Backlog item to a bundled quick-fix sweep.
