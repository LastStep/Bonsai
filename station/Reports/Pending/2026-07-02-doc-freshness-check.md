---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-02
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
- **Duration:** ~8 min
- **Files Read:** 9
  - `station/agent/Routines/doc-freshness-check.md`
  - `station/INDEX.md`
  - `station/CLAUDE.md` (via system-reminder context)
  - `station/agent/Core/routines.md`
  - `station/Playbook/Status.md`
  - `station/code-index.md`
  - `station/Logs/RoutineLog.md` (format check)
  - `station/agent/Core/identity.md` (file existence confirmed via glob)
  - `station/Playbook/Plans/Active/` (directory listing)
- **Files Modified:** 2
  - `station/agent/Core/routines.md` (dashboard update)
  - `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** `git log`, `ls` (directory listings), Glob, Read
- **Errors Encountered:** 0

> **Note:** Status is "partial" because findings are flagged for user decision per the routine procedure — updates to stale docs are proposed, not executed. The routine itself completed successfully.

## Procedure Walkthrough

### Step 1: Scan project documentation and compare against recent git history
- **Action:** Read `station/INDEX.md`, `station/CLAUDE.md`, and `station/code-index.md`. Ran `git log --since="2026-05-04"` to identify all commits since the last doc check.
- **Result:** Found 40+ commits since 2026-05-04. Two major feature ships are the primary source of doc drift:
  - **Plan 40 (2026-06-13):** Frozen v1 schemas, root-relative scaffolding, project-level validate pass.
  - **Plan 41 (2026-06-16):** Headless CLI contract — new `internal/nonint/` package, `--non-interactive`/`--yes`/`--from` flags, `list --json`, `ExitConflict=5`, `docs/agent-interface.md`.
  - **v0.4.1 (2026-05-07):** `bonsai completion` command added (external contribution, merged PR #78).
- **Issues:** Multiple doc files have not been updated to reflect these changes. See Findings Summary.

### Step 2: Check INDEX.md accuracy
- **Action:** Compared INDEX.md's Tech Stack, Key Metrics, and Architecture Overview sections against the actual codebase state.
- **Result:**
  - Tech Stack: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS — all correct).
  - Agent types: 6 — accurate.
  - **CLI commands count: STALE** — INDEX.md shows 8; actual count is 9 (`bonsai completion` added v0.4.1).
  - **Catalog items count: borderline** — INDEX.md shows "~50"; actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). Within the ~50 approximation but on the edge.
  - **Architecture Overview: STALE** — `internal/nonint/` (new headless CLI package from Plan 41) not listed.
- **Issues:** 2 stale entries, 1 borderline.

### Step 3: Check navigation links in station/CLAUDE.md and agent/ subdirectories
- **Action:** Verified all linked files in station/CLAUDE.md navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References) exist on disk. Listed files in `station/agent/Workflows/` and `station/agent/Skills/` to check for unlisted files.
- **Result:**
  - All 37 linked files in CLAUDE.md navigation tables resolve correctly — no broken links.
  - All agent/Core/, agent/Protocols/, agent/Workflows/, agent/Skills/, agent/Sensors/, agent/Routines/ files resolve.
  - **Gap found:** `station/agent/Workflows/plan-grilling.md` exists (added 2026-06-13) but is NOT listed in the CLAUDE.md Workflows navigation table.
  - **Gap found:** `station/agent/Skills/critic-agent-prompts.md` exists but is NOT listed in the CLAUDE.md Skills navigation table.
  - **code-index.md gaps:** `bonsai completion` command and entire `internal/nonint/` package are absent from the code index.
- **Issues:** 2 navigation gaps in CLAUDE.md; 2 gaps in code-index.md. No broken links.

### Step 4: Report findings
- **Action:** Compiled all findings into this report. Per procedure, no doc edits were made — all findings are flagged for user decision.
- **Result:** 6 distinct drift items found across 3 files. See Findings Summary.

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for "Doc Freshness Check" — set Last Ran to 2026-07-02, Next Due to 2026-07-09, Status to done.
- **Result:** Dashboard updated.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `internal/nonint/` package (Plan 41 headless CLI) absent from architecture overview | `station/INDEX.md` — Architecture Overview | Flagged for user update |
| 2 | Low | CLI commands count shows 8; `bonsai completion` (v0.4.1) makes it 9 | `station/INDEX.md` — Key Metrics table | Flagged for user update |
| 3 | Low | `plan-grilling.md` workflow file exists but not listed in Workflows nav table | `station/CLAUDE.md` — Workflows section | Flagged for user update |
| 4 | Low | `critic-agent-prompts.md` skill file exists but not listed in Skills nav table | `station/CLAUDE.md` — Skills section | Flagged for user update |
| 5 | Low | `bonsai completion` command not listed in CLI Commands table | `station/code-index.md` | Flagged for user update |
| 6 | Low | `internal/nonint/` package (significant API: runner.go, events.go, result.go, update.go, remove.go) absent from code index | `station/code-index.md` | Flagged for user update |

---

## Proposed Updates

### Finding 1 — INDEX.md: Add `internal/nonint/` to Architecture Overview

In `station/INDEX.md`, Architecture Overview section, add this line after `internal/generate/`:

```
internal/nonint/      ← headless CLI contract — JSONL event streaming, exit codes, pure *Result cores for init/add/update/remove
```

### Finding 2 — INDEX.md: Update CLI commands count

In `station/INDEX.md`, Key Metrics table, change:

```
| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |
```

to:

```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```

### Finding 3 — CLAUDE.md: Add `plan-grilling.md` to Workflows table

In `station/CLAUDE.md`, Workflows table, add a new row:

```
| Running the 6-critic adversarial plan grilling pipeline; Stress-testing plans before dispatch | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

### Finding 4 — CLAUDE.md: Add `critic-agent-prompts.md` to Skills table

In `station/CLAUDE.md`, Skills table, add a new row (review the file first to confirm the trigger condition):

```
| Using the 6 adversarial critic personas to stress-test plans, PRs, or designs | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

### Findings 5 & 6 — code-index.md: Add `completion` and `internal/nonint/`

`station/code-index.md` needs two additions:
- Add `bonsai completion` row to CLI Commands table (pointing to `cmd/completion.go`)
- Add a new `## Non-Interactive / Headless CLI (internal/nonint/)` section documenting the package (runner.go exit constants, events.go event shapes, result.go Result types, update.go/remove.go headless cores)

These are mechanical updates — the dev agent can fill in exact line numbers by reading the files directly.

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **Finding 1 (Medium):** INDEX.md architecture overview omits `internal/nonint/` — the headless CLI package from Plan 41 is a significant architectural addition. Recommend updating before the next planning session.
- **Finding 3 & 4 (Low):** Two ability files (`plan-grilling.md` and `critic-agent-prompts.md`) are invisible in CLAUDE.md navigation. An agent loading CLAUDE.md won't know these exist. Recommend adding nav rows.
- **Findings 5 & 6 (Low):** code-index.md is missing the `completion` command and the entire `nonint` package. If a dev agent consults code-index.md, it will miss these entry points. Recommend batching these updates into a small doc-refresh task (similar to Plan 37).

---

## Notes for Next Run

- If the doc updates proposed here are executed before the next run, findings 1–6 should all be clean.
- Plans 40 and 41 are both still in `Plans/Active/` — check whether they should be archived (both ships are done).
- The `internal/validate/project.go` file (added Plan 40 Phase 2) is referenced in `validate.go` — confirm the code-index.md Validate section stays accurate after any doc refresh.
- Catalog item count will tick up as abilities are added; consider changing the INDEX.md metric to an exact number (`53`) rather than the approximate `~50` to prevent drift.
