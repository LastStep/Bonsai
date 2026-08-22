---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-22
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (findings flagged; no updates executed per procedure)
- **Duration:** ~8 min
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/CLAUDE.md` (via system-reminder), `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`, `station/Logs/RoutineLog.md`, `station/Reports/Pending/` (listing), git log output
- **Files Modified:** 2 — `station/Reports/Pending/2026-08-22-doc-freshness-check.md` (this report), `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls), Glob, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history

Ran `git -C /home/user/Bonsai log --since="2026-05-04" --oneline --name-only` and reviewed all commits since last run (2026-05-04 through 2026-08-22). Significant shipped changes identified:

- **`bonsai completion` command** (PR #78, commit `2eae9d4`, 2026-05-07) — Added shell completion subcommand (`bash|zsh|fish|powershell`) via `cmd/completion.go`. First external contribution.
- **Plan 39** (PR #102, v0.4.2, 2026-05-13) — `bonsai init`/`add` `--non-interactive --from-config` flags.
- **Plan 40 Phases 1–3** (PRs #114/#116/#115, v0.5.0, 2026-06-13) — Frozen v1 schemas, root-relative scaffolding, project-level validate pass, memory-routing protocol, new `docs/` directory with multiple published reference files (`docs/formats.md`, `docs/concepts.md`, `docs/quickstart.md`, `docs/cli.md`, `docs/custom-files.md`).
- **Plan 41 — Headless CLI Contract** (PRs #120/#122/#123/#121/#125, commit `ab202c3`, 2026-06-16) — Major new package `internal/nonint/` with headless cores (pure Result types + JSONL/exit contract `ExitConflict=5`) for all mutating commands; `list --json`; `docs/agent-interface.md` contract doc.

### Step 2 — Check INDEX.md accuracy

Read `station/INDEX.md`. Compared stated tech stack, folder structure, and CLI command count against actual codebase. Found two stale entries:
- **CLI commands count "8"** — `bonsai completion` makes it 9. Stale since 2026-05-07.
- **Architecture Overview** — `internal/nonint/` package is absent. Stale since 2026-06-16.
- **Document Registry** — `docs/` directory (with 6 user-facing reference files) is unmentioned. Stale since 2026-06-13.

Tech stack table entries (Go version, library choices) remain accurate.

### Step 3 — Check navigation links

Verified all links in `station/CLAUDE.md` navigation tables:

- **Core/** (identity.md, memory.md, self-awareness.md, routines.md) — all present
- **Protocols/** (memory.md, scope-boundaries.md, security.md, session-start.md) — all present
- **Workflows/** (code-review.md, planning.md, pr-review.md, security-audit.md, session-logging.md, test-plan.md, session-wrapup.md, issue-to-implementation.md, routine-digest.md, plan-grilling.md) — all present
- **Skills/** (planning-template.md, review-checklist.md, issue-classification.md, pr-creation.md, bubbletea.md, bonsai-model.md, critic-agent-prompts.md) — all present
- **Routines/** (backlog-hygiene.md, dependency-audit.md, doc-freshness-check.md, memory-consolidation.md, roadmap-accuracy.md, status-hygiene.md, vulnerability-scan.md) — all present
- **Sensors/** (context-guard.sh, scope-guard-files.sh, session-context.sh, status-bar.sh, routine-check.sh, agent-review.sh, dispatch-guard.sh, subagent-stop-review.sh, compact-recovery.sh, statusline.sh) — all present
- `.bonsai/catalog.json` and `.bonsai.yaml` (referenced as Bonsai Reference) — both present

No broken navigation links found.

Also checked `station/code-index.md`:
- CLI Commands table (Step 2): `bonsai completion` is absent.
- Internal packages section: No entry for `internal/nonint/`.

### Step 4 — Report findings

All findings flagged below. No updates executed per procedure ("flag for user decision").

### Step 5 — Update dashboard

Dashboard row updated: Last Ran → 2026-08-22, Next Due → 2026-08-29, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | CLI commands count "8" is stale — `bonsai completion` added (PR #78, 2026-05-07) makes it 9 | `station/INDEX.md` — Key Metrics table | Flagged for user |
| 2 | Medium | `internal/nonint/` package absent from Architecture Overview — major new package (Plan 41, 2026-06-16) | `station/INDEX.md` — Architecture Overview section | Flagged for user |
| 3 | Low | `docs/` directory (6 reference files: quickstart, concepts, cli, formats, custom-files, agent-interface) absent from Document Registry and Architecture | `station/INDEX.md` — Document Registry + Architecture sections | Flagged for user |
| 4 | Low | `bonsai completion` missing from CLI Commands table in code-index | `station/code-index.md` — CLI Commands table | Flagged for user |
| 5 | Low | `internal/nonint/` package missing from code-index — substantial package with headless cores, Result types, JSONL/exit contract | `station/code-index.md` — Internal packages section | Flagged for user |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[INDEX.md: Fix CLI commands count]** Change "8 (init, add, remove, list, catalog, update, guide, validate)" to "9 (init, add, remove, list, catalog, update, guide, validate, completion)" in Key Metrics table.

- **[INDEX.md: Add `internal/nonint/` to Architecture Overview]** Insert a line in the architecture diagram for `internal/nonint/` with description: "headless cores for all mutating commands — pure Result types + JSONL/exit contract (ExitConflict=5)". Also update the text description below the diagram.

- **[INDEX.md: Add `docs/` to Document Registry]** Add a row for `docs/` noting it contains published user-facing reference docs (agent-interface.md, formats.md, quickstart.md, concepts.md, cli.md, custom-files.md). Verify whether a single row covers the directory or each file warrants its own entry.

- **[code-index.md: Add `bonsai completion` to CLI Commands table]** Add row: `bonsai completion` | `cmd/completion.go` | shell completion subcommand (bash/zsh/fish/powershell).

- **[code-index.md: Add `internal/nonint/` section]** Add a full section documenting the nonint package types (Result, ExitCode, Event, Runner, Config) and key functions/entry points.

## Notes for Next Run

- The gap since last run was ~110 days (2026-05-04 → 2026-08-22); 3 major plans shipped in that interval (Plans 39, 40, 41) — significant doc drift accumulated. More frequent execution of this routine would catch changes closer to ship time.
- All navigation links are clean — no maintenance needed there.
- `Roadmap.md` is accurate; Phase 1 checkboxes match shipped features.
