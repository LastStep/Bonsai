---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-31
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
- **Duration:** ~10 min
- **Files Read:** 13 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/agent/Core/identity.md`, `station/Playbook/Roadmap.md`, `station/CLAUDE.md` (via system-reminder), `station/code-index.md`, `station/Logs/RoutineLog.md`, `internal/nonint/nonint.go`, `internal/nonint/result.go`, `internal/nonint/events.go`, `internal/validate/project.go`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, head)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Retrieved recent git log (last 30 days, then full recent list); identified commits since last doc freshness check (2026-05-04). Compared against docs in `station/`.
- **Result:** No commits in the last 7 days. Since the last check (2026-05-04), significant code changes landed in Plan 40 (2026-06-13) and Plan 41 (2026-06-16). Plan 41 introduced a new `internal/nonint/` package (10+ files) for headless CLI contracts and MCP-ready cores. Plan 40 added `internal/validate/project.go` for project-level validation and a new `docs/` directory (7 guide files). A new `internal/generate/list_snapshot.go` was added for `list --json`. None of these are reflected in `station/code-index.md` or `station/INDEX.md`.
- **Issues:** 4 doc-drift items found (see Findings Summary).

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` in full; verified tech stack table, key metrics table, architecture overview, and document registry against codebase reality.
- **Result:** Tech stack is accurate. CLI command count (8) is accurate. Agent type count (6) is accurate. Architecture diagram is partially stale — it lists `internal/validate/`, `internal/wsvalidate/`, `internal/generate/`, `internal/config/`, `internal/catalog/`, and `internal/tui/` but is missing `internal/nonint/` (new, Plan 41). Document registry does not mention the new `docs/` directory (website guide pages mirrored to `.bonsai/`).
- **Issues:** Architecture diagram drift (missing `internal/nonint/`).

### Step 3: Check navigation links
- **Action:** Cross-referenced every file link in `station/CLAUDE.md` (Core, Protocols, Workflows, Skills, Routines, Sensors, and External References tables) against the actual filesystem contents of `station/agent/`. Also checked for files present on disk but not listed in the navigation tables.
- **Result:** All 40+ linked files resolve correctly — no broken links found. However, two files exist on disk that have no entry in the CLAUDE.md navigation tables:
  - `agent/Skills/critic-agent-prompts.md` (exists; not in Skills table)
  - `agent/Workflows/plan-grilling.md` (exists; not in Workflows table)
- **Issues:** 2 unlisted files in navigation (low severity — agent may not find them via nav).

### Step 4: Report findings
- **Action:** Compiled all findings, classified by severity, flagged for user decision as required by the procedure.
- **Result:** 6 findings documented. No changes to source docs made (procedure requires flag-only, no execution).
- **Issues:** none.

### Step 5: Update dashboard
- **Action:** Updated `Doc Freshness Check` row in `station/agent/Core/routines.md` — `Last Ran` → 2026-07-31, `Next Due` → 2026-08-07, `Status` → `done`.
- **Result:** Dashboard updated.
- **Issues:** none.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium-High | `internal/nonint/` package (Plan 41, 10+ files) has no entry in code-index.md | `station/code-index.md` | Flagged — propose adding a Nonint section |
| 2 | Medium | `station/INDEX.md` architecture diagram is missing `internal/nonint/` | `station/INDEX.md` (line ~66 architecture block) | Flagged — propose one-line addition |
| 3 | Medium | `internal/generate/list_snapshot.go` (Plan 41) not in code-index.md Generator section | `station/code-index.md` — Generator / `catalog_snapshot.go` section | Flagged — propose adding row under catalog_snapshot.go |
| 4 | Medium | `internal/validate/project.go` (Plan 40) not in code-index.md Validate section | `station/code-index.md` — Validate section | Flagged — propose adding row to Validate table |
| 5 | Low | `agent/Skills/critic-agent-prompts.md` exists but has no entry in CLAUDE.md Skills navigation | `station/CLAUDE.md` — Skills table | Flagged — propose adding row or confirm intentional omission |
| 6 | Low | `agent/Workflows/plan-grilling.md` exists but has no entry in CLAUDE.md Workflows navigation | `station/CLAUDE.md` — Workflows table | Flagged — propose adding row or confirm intentional omission |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

### Finding 1 — code-index.md: missing `internal/nonint/` section (Medium-High)

Plan 41 introduced a new top-level package `internal/nonint/` with 10 files: `nonint.go`, `result.go`, `events.go`, `runner.go`, `config.go`, `update.go`, `remove.go`, plus tests. This implements the headless CLI contract (typed `*Result` shape, JSONL exit contract, `ExitConflict=5`) and MCP-ready cores for `init`, `add`, `update`, `remove`. The code-index.md has no section for this package.

**Proposed addition** — new section between Validate and Workspace-path Validation:

```markdown
## Headless CLI / Nonint (`internal/nonint/`) — Plan 41

Drives mutating commands (init, add, update, remove) without TUI prompts. Inputs are typed options (optionally loaded from a YAML config file); each core returns a structured `*Result` plus JSONL event stream. Exit contract: 0 ok, 2 conflict, 3 invalid, 4 user-abort, 5 conflict (ExitConflict).

| Type / Function | File | Purpose |
|-----------------|------|---------|
| `RunInit()` / `RunAdd()` | `runner.go` | Headless init and add cores |
| `RunUpdate()` | `update.go` | Headless update core |
| `RunRemove()` | `remove.go` | Headless remove core |
| `Result` | `result.go` | Structured return value for all headless cores |
| `Event` / event types | `events.go` | JSONL event stream shapes |
| `LoadConfig()` | `config.go` | Load and validate nonint config from YAML |
```

---

### Finding 2 — INDEX.md: architecture diagram missing `internal/nonint/` (Medium)

The architecture block lists 6 internal packages but omits `internal/nonint/`.

**Proposed one-line addition** after the `internal/validate/` line:

```
internal/nonint/      ← headless CLI contract — typed Result + JSONL events for MCP-ready cores
```

---

### Finding 3 — code-index.md: `list_snapshot.go` not in Generator section (Medium)

`internal/generate/list_snapshot.go` was added in Plan 41 for `bonsai list --json` output. The Generator section of code-index.md documents `catalog_snapshot.go` but not `list_snapshot.go`.

**Proposed addition** — new row under catalog_snapshot.go in the Generator table:

```
| `list_snapshot.go` | Plan 41 | Writes structured JSON output for `bonsai list --json` |
```

---

### Finding 4 — code-index.md: `validate/project.go` not in Validate section (Medium)

Plan 40 Phase 2 added `internal/validate/project.go` for project-level validation (lints repo-resident formats). The Validate section in code-index.md only documents `validate.go`.

**Proposed addition** — new row in the Validate section:

```
| `RunProject()` (or equiv) | `project.go` | Project-level validation pass — lints repo-resident format files (Plan 40 Phase 2) |
```

(Exact function name needs verification from the source file.)

---

### Finding 5 — CLAUDE.md: `agent/Skills/critic-agent-prompts.md` not listed (Low)

The file `station/agent/Skills/critic-agent-prompts.md` exists on disk but has no row in the Skills navigation table in `station/CLAUDE.md`. If this skill is intended to be used, the agent cannot discover it via the standard navigation.

**Decision needed:** Add a row to the Skills table, or confirm the file is intentionally off-nav (e.g. it's a reference doc rather than a loadable skill).

---

### Finding 6 — CLAUDE.md: `agent/Workflows/plan-grilling.md` not listed (Low)

The file `station/agent/Workflows/plan-grilling.md` exists on disk but has no row in the Workflows navigation table in `station/CLAUDE.md`. The workflow is not discoverable via standard navigation.

**Decision needed:** Add a row to the Workflows table, or confirm the file is intentionally off-nav (e.g. it is invoked directly by the agent without requiring a nav entry).

---

## Notes for Next Run

- The `internal/nonint/` and `docs/` additions are the largest drift items from this period. Once code-index.md and INDEX.md are updated, future runs will be lighter.
- `critic-agent-prompts.md` and `plan-grilling.md` may be intentionally unlisted (both look like supplementary references). Confirm before adding to nav.
- Plans 40 and 41 are still in `Plans/Active/` — noted by Backlog Hygiene (see `2026-07-31-backlog-hygiene.md`). Archiving them is a separate action (not a doc-freshness concern).
- All existing navigation links in `station/CLAUDE.md` resolve correctly — no broken links.
