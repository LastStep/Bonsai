---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-11
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
- **Duration:** ~8 min
- **Files Read:** 9 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/CLAUDE.md`, `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/code-index.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Plans/Active/` (listing), `station/Playbook/Plans/Archive/` (listing)
- **Files Modified:** 3 — `station/Reports/Pending/2026-09-11-doc-freshness-check.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, head, grep), file existence checks
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history

Ran `git log --since="7 days ago" --oneline --name-status`. Three commits in the window (all 2026-09-11):

- `ddccba0` — status-hygiene report added
- `ca6a6c4` — status-hygiene routine: Status.md, StatusArchive.md, Plan 41 archived, routines.md updated
- `bbe9c1e` — backlog-hygiene routine: Backlog.md, RoutineLog.md, backlog report added, routines.md updated

All three commits are routine maintenance — no new features, services, or config introduced in the last 7 days. However, comparing the cumulative code state (Plans 39 and 41 shipped since last doc-freshness run on 2026-05-04) reveals three documentation gaps described below.

### Step 2 — Check INDEX.md accuracy

- Tech stack table: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS).
- Folder structure: mostly accurate, but the `internal/` block is **incomplete** — `internal/nonint/` is absent.
- Project description: accurate — still correctly describes the CLI scaffolding tool.
- Key Metrics CLI command count: **stale** — says 8 commands but `completion` was added 2026-05-07.
- Architecture overview diagram: `internal/nonint/` missing from the internal package list.

### Step 3 — Check navigation links

Verified all links in `station/CLAUDE.md` navigation tables (Core, Skills, Protocols, Workflows, Routines, Sensors, External References). Result:

- **All Core, Skills, Protocols, Workflows, Routines, and Sensor files:** 100% present, no broken links.
- **External reference `../.bonsai/catalog.json`:** Present at `/home/user/Bonsai/.bonsai/catalog.json` — OK.
- **External reference `../.bonsai.yaml`:** Present — OK.
- **All Playbook, Logs, Reports references:** Present — OK.

No broken navigation links found.

### Step 4 — Compile findings and flag for user decision

Three documentation gaps identified and documented below. None are critical; all are flag-only per procedure.

### Step 5 — Update dashboard

Updated `station/agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-09-11, Next Due → 2026-09-18, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|---------|--------------|
| 1 | Medium | `internal/nonint/` package (Plan 41, headless CLI contract) absent from INDEX.md architecture block and code-index.md | `station/INDEX.md` arch diagram + `station/code-index.md` | Flagged for user — proposed update below |
| 2 | Low | CLI command count stale: says 8, should be 9 (`completion` added 2026-05-07) | `station/INDEX.md` Key Metrics + `station/code-index.md` CLI Commands table | Flagged for user — proposed update below |
| 3 | Low | `docs/` directory (agent-interface.md, cli.md, concepts.md, etc.) not referenced in station docs | `station/INDEX.md` Document Registry | Flagged for user — proposed update below |

---

## Errors & Warnings

None.

---

## Items Flagged for User Review

### Finding 1 — `internal/nonint/` undocumented (Medium)

**What:** Plan 41 (shipped 2026-06-16) created `internal/nonint/` — the headless CLI contract package containing `RunInit`, `RunAdd`, `LoadConfig`, `Result`, `EmitJSONL`, plus remove and update orchestrators. This is a substantial package (10 files, ~7 modules) fully absent from station docs.

**Where it's missing:**
- `station/INDEX.md` — architecture block `internal/` list (between `internal/wsvalidate/` and `internal/tui/`)
- `station/code-index.md` — no section for the package at all
- Root `CLAUDE.md` Project Structure — `internal/` listing is also incomplete (but that file is out of station/ scope per procedure)

**Proposed addition to `station/INDEX.md` architecture block:**
```
internal/nonint/      ← headless CLI contract — RunInit/RunAdd/LoadConfig, JSONL event emission, exit codes
```

**Proposed new section in `station/code-index.md`** (after the `wsvalidate` section):

```
## Headless CLI Contract (`internal/nonint/`)

Non-interactive (CI/eval) entry points. Implements Plan 41 Phase 1–5.

| Type / Function | File | Purpose |
|-----------------|------|---------|
| `LoadConfig()` | `config.go` | Read `.bonsai.yaml`-shaped YAML, apply defaults, validate |
| `RunInit()` | `runner.go` | Headless init — accepts `*config.ProjectConfig`, returns `*Result` |
| `RunAdd()` | `runner.go` | Headless add — accepts loaded config + catalog, returns `*Result` |
| `Result` / `Counts` | `result.go` | Structured return value (written files + per-category counts) |
| `EmitJSONL()` | `events.go` | Serialize `Result` to JSONL for `--non-interactive` stdout contract |
| `RunUpdate()` | `update.go` | Headless update — re-renders all agents, returns `*Result` |
| `RunRemove()` | `remove.go` | Headless remove — removes agent or item, returns `*Result` |
| Exit codes | `runner.go` | `ExitOK=0`, `ExitInvalidConfig=2`, `ExitRuntime=3`, `ExitWrongCWDForInit=4`, `ExitConflict=5` |
```

---

### Finding 2 — CLI command count stale (Low)

**What:** `station/INDEX.md` Key Metrics says "8 (init, add, remove, list, catalog, update, guide, validate)". The `completion` command was merged 2026-05-07 (first external contribution, @mvanhorn, PR #78). `station/code-index.md` CLI Commands table also omits it.

**Proposed fix for `station/INDEX.md`:**
Change `| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |`
to `| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |`

**Proposed addition to `station/code-index.md` CLI Commands table:**
Add row: `| `bonsai completion` | `cmd/completion.go` | `completionCmd` — shell completion scripts (bash/zsh/fish/powershell) |`

---

### Finding 3 — `docs/` directory undocumented (Low)

**What:** `/home/user/Bonsai/docs/` contains: `README.md`, `agent-interface.md`, `cli.md`, `concepts.md`, `custom-files.md`, `formats.md`, `quickstart.md`, `assets/`. The `agent-interface.md` file is the Plan 41 headless contract reference explicitly named in Status.md Done row ("docs/agent-interface.md contract doc"). None of these appear in `station/INDEX.md`.

**Proposed addition to `station/INDEX.md` Document Registry:**
```
| `docs/agent-interface.md` | Headless CLI contract — JSONL event shapes, exit codes, `--non-interactive` usage | When implementing or reviewing CI/eval integrations |
| `docs/` (other) | End-user guides — quickstart, cli reference, concepts, custom-files, formats | When updating user-facing documentation |
```

---

## Notes for Next Run

- Three findings are all flag-only; none were auto-fixed. Apply proposed updates if user approves.
- All navigation links are clean — no link rot introduced.
- The doc drift accumulated between 2026-05-04 (last run) and now (2026-09-11) — 130 days gap. Plans 39 and 41 both shipped in that window, each adding code with no corresponding station doc update.
- Consider flagging a doc-update step in future plan templates to prevent recurrence.
