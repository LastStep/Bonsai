---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-06
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
- **Files Read:** 12 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `cmd/completion.go`, `docs/formats.md`, `cmd/` (directory listing), `internal/` (directory listing), `internal/nonint/` (directory listing), git log output
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls), Grep
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history

- **Action:** Read `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`. Retrieved git log from `2026-05-04` onwards (since last run) using `git log --since`. Also listed `cmd/` and `internal/` directories to cross-check structure.
- **Result:** 44 commits since last run (2026-05-04). Key feature commits are from two plans:
  - **Plan 40** (2026-06-13): root-relative scaffolding (Phase 1), project-level validate pass (Phase 2), `docs/formats.md` memory-note + manifest format spec (Phase 3).
  - **Plan 41** (2026-06-16): Result reshape + headless exit/event contract (Phase 1), headless update core with `--non-interactive`/`--skip-conflicts` (Phase 2), headless remove core with `--yes`/`--from` (Phase 3), `list --json` (Phase 4), `docs/agent-interface.md` canonical headless contract + CHANGELOG (Phase 5).
  - Also: `feat(cmd): add explicit completion subcommand` (2026-05-07).
- **Issues:** Four documentation gaps identified (see Findings Summary below).

### Step 2: Check INDEX.md accuracy

- **Action:** Verified tech stack table, Key Metrics table, Architecture Overview, and Document Registry in `station/INDEX.md` against actual codebase.
- **Result:**
  - Tech stack (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS) — accurate.
  - Agent types count (6) — accurate.
  - Catalog items (~50) — reasonable approximation, no count change.
  - **CLI commands count is stale:** says `8 (init, add, remove, list, catalog, update, guide, validate)`. A `completion` command (`cmd/completion.go`) was added in May 2026. Should be `9 (init, add, remove, list, catalog, update, guide, validate, completion)`.
  - **Architecture Overview missing `internal/nonint/`:** The section lists `internal/catalog/`, `internal/config/`, `internal/generate/`, `internal/validate/`, `internal/wsvalidate/`, `internal/tui/` — but not `internal/nonint/`, which was added by Plan 41 and is the headless CLI contract implementation layer.
  - Document Registry does not reference `docs/agent-interface.md`, the canonical headless contract added by Plan 41 Phase 5.
- **Issues:** 3 stale items in INDEX.md (CLI count, missing nonint layer, missing agent-interface.md reference).

### Step 3: Check navigation links

- **Action:** Verified all links in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References) against actual files on disk.
- **Result:**
  - Core: identity.md, memory.md, self-awareness.md — all exist ✓
  - Protocols: memory.md, scope-boundaries.md, security.md, session-start.md — all exist ✓
  - Workflows: code-review.md, planning.md, pr-review.md, security-audit.md, session-logging.md, test-plan.md, session-wrapup.md, issue-to-implementation.md, routine-digest.md — all exist ✓
  - Skills: planning-template.md, review-checklist.md, issue-classification.md, pr-creation.md, bubbletea.md, bonsai-model.md — all exist ✓
  - Routines: backlog-hygiene.md, dependency-audit.md, doc-freshness-check.md, memory-consolidation.md, roadmap-accuracy.md, status-hygiene.md, vulnerability-scan.md — all exist ✓
  - Sensors: context-guard.sh, scope-guard-files.sh, session-context.sh, status-bar.sh, routine-check.sh, agent-review.sh, dispatch-guard.sh, subagent-stop-review.sh, compact-recovery.sh, statusline.sh — all exist ✓
  - External references: INDEX.md, Playbook/Status.md, Playbook/Roadmap.md, Playbook/Standards/SecurityStandards.md, Playbook/Plans/Active/, Playbook/Backlog.md, Logs/KeyDecisionLog.md, Reports/Pending/, Reports/report-template.md, code-index.md — all exist ✓
  - CLAUDE.md project structure: `cmd/completion.go` is NOT listed (only 9 files listed, missing completion.go).
  - CLAUDE.md project structure: `internal/nonint/` is NOT listed (6 internal packages listed, missing nonint/).
- **Issues:** 2 stale entries in CLAUDE.md project structure (missing completion.go, missing internal/nonint/).

### Step 4: Report findings

- **Action:** Compiled 4 findings (2 in INDEX.md, 2 in CLAUDE.md). Also noted code-index.md has no section for internal/nonint and doesn't list the completion command — flagged as a secondary gap. Per procedure, not executing updates — flagging for user decision.
- **Result:** See Findings Summary below.
- **Issues:** None in the reporting step.

### Step 5: Update dashboard

- **Action:** Updated `agent/Core/routines.md` — set `Last Ran` → 2026-08-06, `Next Due` → 2026-08-13, `Status` → done for the Doc Freshness Check row.
- **Result:** Dashboard updated.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | CLI commands count stale — says `8`, should be `9`; `completion` command missing from list | `station/INDEX.md` — Key Metrics table | Flagged for user update |
| 2 | Medium | `internal/nonint/` package entirely absent from architecture overview | `station/INDEX.md` — Architecture Overview | Flagged for user update |
| 3 | Medium | `docs/agent-interface.md` not referenced anywhere in station docs | `station/INDEX.md` — Document Registry | Flagged for user update |
| 4 | Low | `cmd/completion.go` and `internal/nonint/` missing from project structure file tree | `station/CLAUDE.md` — Project Structure section | Flagged for user update |
| 5 | Low | `code-index.md` has no section for `internal/nonint/` and `completion` not in CLI Commands table | `station/code-index.md` | Flagged for user update |

---

## Proposed Updates (not executed — for user decision)

### INDEX.md — Key Metrics

Change:
```
| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |
```
To:
```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```

### INDEX.md — Architecture Overview

Add to the internal layers block:
```
internal/nonint/      ← headless CLI contract — Result type, JSONL/JSON serialization, exit codes, runner
```

### INDEX.md — Document Registry

Add row:
```
| `docs/agent-interface.md` | Canonical headless CLI contract — per-command flags, JSONL/JSON shapes, exit-code table | When driving bonsai without a TTY (CI, MCP, scripts) |
```

### CLAUDE.md — Project Structure (cmd/ listing)

Add after `validate.go` line:
```
│   └── completion.go        ← bonsai completion — shell completion (bash/zsh/fish/powershell)
```

### CLAUDE.md — Project Structure (internal/ listing)

Add before or after `validate/`:
```
│   ├── nonint/
│   │   └── nonint.go        ← headless CLI contract — Result, events, runner, exit codes; used by init/add/update/remove --non-interactive
```

### code-index.md — CLI Commands table

Add row:
```
| `bonsai completion` | `cmd/completion.go` | `completionCmd` — generate shell completion scripts (bash/zsh/fish/powershell) |
```

### code-index.md — New section for internal/nonint/

Add a new section (between Workspace-path Validation and TUI sections):

```
## Headless CLI Contract (`internal/nonint/`)

Plan 41 — MCP-ready headless cores for init, add, update, remove. All mutating commands stream JSONL (file/summary events) to stdout; read commands (list/catalog/validate with `--json`) emit indent-2 JSON. Flags: `--non-interactive --from-config` (init/add), `--non-interactive --skip-conflicts` (update), `--non-interactive --yes --from` (remove). Exit codes pinned in `runner.go`.

| File | Purpose |
|------|---------|
| `nonint.go` | Config types for headless input shapes |
| `result.go` | `Result` type wrapping `generate.WriteResult` + Warnings |
| `events.go` | JSONL event shapes (file / summary) |
| `runner.go` | Exit code constants (canonical source) |
| `update.go` | Headless update core |
| `remove.go` | Headless remove core |
```

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

All 5 findings are proposed-update candidates — none are blocking but they create agent confusion risk if INDEX.md or CLAUDE.md are used to navigate the codebase. The `internal/nonint/` omission is the highest-priority gap: any agent reasoning about command architecture will miss the headless layer entirely.

Recommended priority:
1. Add `internal/nonint/` to CLAUDE.md and INDEX.md (highest impact — absent entirely)
2. Update CLI command count and list in INDEX.md (quick fix, high visibility)
3. Add `docs/agent-interface.md` to INDEX.md Document Registry (useful for scripting/MCP context)
4. Update `code-index.md` with nonint section and completion entry (useful for code navigation)
5. Add `completion.go` to CLAUDE.md project structure (cosmetic, lowest priority)

---

## Notes for Next Run

- Plans 40 and 41 were the primary sources of drift in this cycle. Both are now fully reflected in these proposals.
- No commits since 2026-06-16 (the repo has been quiet). If that's still true on the next run, focus can shift to verifying doc accuracy rather than catching up on code changes.
- The `docs/` directory now contains `formats.md` and `agent-interface.md` — consider whether these should be surfaced in `station/CLAUDE.md`'s External References table, not just INDEX.md.
