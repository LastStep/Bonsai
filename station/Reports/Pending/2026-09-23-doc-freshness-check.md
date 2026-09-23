---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-23
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
- **Files Read:** 8
  - `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`
  - `/home/user/Bonsai/station/INDEX.md`
  - `/home/user/Bonsai/station/CLAUDE.md` (via system-reminder injection)
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/code-index.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/cmd/completion.go`
- **Files Modified:** 2
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard updated)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry appended)
- **Tools Used:** `git log --oneline -20`, `git log --since="7 days ago"`, `git log --since="2026-05-01" --name-only`, `ls` on catalog/agents/, cmd/, internal/, station subdirs
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation and compare against recent git history

No commits in the last 7 days (since 2026-09-16). Checked commits since last routine run (2026-05-04) to capture accumulated drift.

**Significant changes detected:**
- **Plan 41 (2026-06-16)**: Headless CLI contract — added `internal/nonint/` package with pure headless cores for init/add/update/remove, JSONL output (`internal/nonint/events.go`), exit contract (ExitConflict=5), `list --json`, and `docs/agent-interface.md`.
- **PR #78 (2026-05-07)**: External contributor `@mvanhorn` added `bonsai completion [bash|zsh|fish|powershell]` shell completion command.
- **Plan 40 (2026-06-13)**: Added `internal/validate/project.go` (project-level validate pass).
- **v0.4.2 (2026-05-13)**: Added `--non-interactive --from-config` flags to `bonsai init` and `bonsai add`.
- **v0.4.3 hotfix (2026-05-13)**: Absolute paths baked in `.claude/settings.json`.

### Step 2: Check INDEX.md accuracy

Verified tech stack, folder structure, agent count, catalog item count, and CLI command count against the actual codebase.

**Findings:**
- Tech stack (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template): **Accurate**
- Agent types (6): `backend, devops, frontend, fullstack, security, tech-lead` — **Accurate**
- Catalog items (~50): Counted 53 actual items (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines) — **Accurate (within stated approximation)**
- CLI commands (8): INDEX.md lists `init, add, remove, list, catalog, update, guide, validate` — **STALE** — `bonsai completion` was added (PR #78, 2026-05-07), making it 9 commands
- Architecture overview: Missing `internal/nonint/` package — **STALE** — this package was added by Plan 41 and is a significant new internal layer

### Step 3: Check navigation links

Verified all links in `station/CLAUDE.md` navigation tables against the actual filesystem.

**All listed links resolve correctly.** No broken links.

**Extra files on disk not in CLAUDE.md navigation tables (informational):**
- `station/agent/Skills/critic-agent-prompts.md` — present on disk, not in Skills table
- `station/agent/Workflows/plan-grilling.md` — present on disk, not in Workflows table
- `station/agent/Skills/bubbletea/` (directory) — present alongside `bubbletea.md`, not referenced

Also checked `code-index.md`: does not include `internal/nonint/` coverage — last synced in Plan 37 (2026-05-07), before Plan 41 shipped.

### Step 4: Report findings

See Findings Summary table below. All findings are flagged for user decision — no doc edits executed per routine procedure.

### Step 5: Update dashboard

Dashboard row for "Doc Freshness Check" updated in `agent/Core/routines.md`.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | INDEX.md CLI command count says 8 — should be 9 after `bonsai completion` was added (PR #78, 2026-05-07). Fix: change `8 (init, add, remove, list, catalog, update, guide, validate)` to `9 (init, add, remove, list, catalog, update, guide, validate, completion)` | `station/INDEX.md` — Key Metrics table, CLI commands row | Flagged — user decision |
| 2 | Medium | INDEX.md architecture overview missing `internal/nonint/` package. Plan 41 added headless CLI cores here (pure Result shapes, JSONL events, exit codes). Fix: add row `internal/nonint/  ← headless CLI cores — pure Result/event shapes for init/add/update/remove; JSONL output + exit contract` | `station/INDEX.md` — Architecture Overview section | Flagged — user decision |
| 3 | Low | `code-index.md` does not cover `internal/nonint/`. Package was added in Plan 41 with `nonint.go`, `result.go`, `events.go`, `runner.go`, `remove.go`, `update.go` and test data. | `station/code-index.md` — missing section for nonint package | Flagged — user decision |
| 4 | Low | `station/agent/Skills/critic-agent-prompts.md` exists on disk but is not listed in station/CLAUDE.md Skills navigation table | `station/CLAUDE.md` — Skills table | Flagged — user decision |
| 5 | Low | `station/agent/Workflows/plan-grilling.md` exists on disk but is not listed in station/CLAUDE.md Workflows navigation table | `station/CLAUDE.md` — Workflows table | Flagged — user decision |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **[Medium] INDEX.md CLI command count drift** — `bonsai completion` is a registered CLI command (in `cmd/completion.go`, shown in `bonsai --help`) but INDEX.md Key Metrics still shows 8 commands. Recommend bumping to 9 and adding `completion` to the parenthetical list.
- **[Medium] INDEX.md architecture diagram missing `internal/nonint/`** — This package is the headless contract layer enabling MCP integration (the Plan 42 fast-follow). Should be listed alongside the other internal packages in the architecture overview.
- **[Low] `code-index.md` missing nonint coverage** — The code index documents every other internal package but has no section for `internal/nonint/`. Worth a targeted code-index pass (similar to Plan 37) before Plan 42 work begins.
- **[Low] `critic-agent-prompts.md` and `plan-grilling.md` unlisted in CLAUDE.md navigation** — These files are present in the workspace but not in the nav table. Intentional (internal-use-only) or an oversight? If agents should load them, add trigger rows.

---

## Notes for Next Run

- No git commits in the last 7 days — project has been idle since the Plan 41 ship on 2026-06-16. Next run should check if any Plan 42 (MCP server) work has shipped.
- Items 1 and 2 are small fixes that could be batched as a mini doc refresh (similar to Plan 37) rather than a full plan — likely worth doing before Plan 42 kicks off since agent-interface.md already references the nonint package.
- All navigation links in station/CLAUDE.md are intact — no broken-link maintenance needed.
