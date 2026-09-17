---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-17
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (previous last_ran from dashboard)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~4 minutes
- **Files Read:** 10
  - `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`
  - `/home/user/Bonsai/station/INDEX.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/CLAUDE.md` (via system-reminder)
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/code-index.md`
  - `/home/user/Bonsai/station/agent/Core/memory.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/docs/agent-interface.md` (preview — confirming content)
- **Files Modified:** 3
  - `/home/user/Bonsai/station/Reports/Pending/2026-09-17-doc-freshness-check.md` (this report)
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, file existence checks, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history

Ran `git log --oneline --since="7 days ago"` — no commits in the last 7 days (today is 2026-09-17, last commit `c6a6757` is from the June 2026 ship cycle). Extended to `git log --oneline --since="2026-05-04"` to cover everything since the last doc freshness check.

**Significant changes since 2026-05-04:**
- **Plan 40** (Phases 1–3, v0.5.0): frozen v1 schemas, root-relative scaffolding, project-level validate pass, memory-routing docs, guide Formats page
- **Plan 41** (all 5 phases, 2026-06-16): headless CLI contract — `internal/nonint/` package, `--yes`/`--from` on remove, `--non-interactive`/`--skip-conflicts` on update, `list --json`, typed `*Result` shapes, `ExitConflict=5`, `docs/agent-interface.md` contract doc
- Various dependency bumps and the first external contribution (`bonsai completion` command — already reflected in CLAUDE.md)

### Step 2: Check INDEX.md accuracy

Verified tech stack, folder structure, key metrics:
- **Tech stack table:** accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, yaml.v3, text/template, embed.FS. All correct.
- **Key metrics:** "Agent types: 6" — correct. "Catalog items: ~50" — approximate, acceptable. "CLI commands: 8" — correct (same 8 commands; Plan 41 added flags, not new commands).
- **Architecture diagram:** `cmd/`, `internal/catalog/`, `internal/config/`, `internal/generate/`, `internal/validate/`, `internal/wsvalidate/`, `internal/tui/`, `catalog/` — all listed. **`internal/nonint/` is missing.** Plan 41 added this package (8 files) and it is not mentioned anywhere in `INDEX.md`.
- **Document Registry:** Does not include `docs/agent-interface.md`. Plan 41 explicitly called this the "canonical contract doc" and single source of truth for the headless CLI.

### Step 3: Check navigation links

Verified all links in `station/CLAUDE.md` navigation tables by checking file existence for all 52 referenced paths. Results:
- All Core, Protocol, Workflow, Skill, Routine, Sensor, Playbook, Log, and Report links resolve correctly.
- The `../.bonsai/catalog.json` and `../.bonsai.yaml` links (relative to `station/`) correctly resolve to `/home/user/Bonsai/.bonsai/catalog.json` and `/home/user/Bonsai/.bonsai.yaml` — both exist.

Checked links in `agent/Core/memory.md` References section:
- 6 links to `../../Research/RESEARCH-*.md` are broken. Path resolves to `station/Research/RESEARCH-*.md`, but no `station/Research/` directory exists. Backlog line 109 confirms this is a planned scaffolding addition — the links are aspirational. **Low severity — known gap, tracked in Backlog.**

No broken links in `agent/Core/identity.md`, `agent/Core/self-awareness.md`, or `agent/Core/routines.md`.

### Step 4: Report findings

See Findings Summary table below. All findings are flagged for user decision — no doc updates were executed.

### Step 5: Update dashboard

Updated `agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-09-17, Next Due → 2026-09-24, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | `internal/nonint/` package (Plan 41, 8 files) not documented anywhere in station docs — missing from INDEX.md architecture diagram and entirely absent from code-index.md | `station/INDEX.md`, `station/code-index.md` | Flagged for user review — propose adding architecture entry + code-index section |
| 2 | LOW | `docs/agent-interface.md` (Plan 41 canonical contract doc) not in INDEX.md Document Registry | `station/INDEX.md` | Flagged for user review — propose adding a Document Registry row |
| 3 | LOW | 6 broken Research/ file links in `agent/Core/memory.md` — `../../Research/RESEARCH-*.md` paths do not exist (no `station/Research/` directory) | `station/agent/Core/memory.md:87–92` | Flagged for user review — Backlog line 109 confirms Research/ scaffolding is planned; links may be intended as forward-references |
| 4 | INFO | Plan 41 (headless CLI contract) not reflected as a completed milestone in Roadmap.md | `station/Playbook/Roadmap.md` | Informational — Roadmap is a phase-level view; headless CLI is arguably a Phase 2/3 enabler. Flagged but not blocking. |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**Finding 1 (MEDIUM) — code-index.md: Add `internal/nonint/` section**

Proposed additions to `station/code-index.md`:

```markdown
## Headless CLI Cores (`internal/nonint/`) — Plan 41

Pure-function mutating cores for non-interactive use. Each core takes typed options, returns a `Result` — no prompts, no `os.Exit`, no stdout data from inside the core. The CLI adapter serializes to JSONL; a future MCP adapter will serialize the same `Result` to structured content.

| File | Purpose |
|------|---------|
| `nonint.go` | Package-level types and shared option structs |
| `config.go` | Config overlay loading + `validateOverlay()` |
| `events.go` | JSONL event shapes (`file`, `summary` event kinds) |
| `result.go` | `Result` shape + per-command result types |
| `runner.go` | Exit-code constants (`ExitOK=0`, `ExitInvalidConfig=2`, `ExitRuntime=3`, `ExitWrongCWDForInit=4`, `ExitConflict=5`) |
| `remove.go` | `RunRemoveAgent` / `RunRemoveItem` headless cores |
| `update.go` | `RunUpdate` headless core |
```

**Finding 1 (cont.) — INDEX.md: Add `internal/nonint/` to architecture diagram**

Proposed addition to the Architecture Overview box:
```
internal/nonint/      ← headless CLI cores — pure Result-returning functions, no prompts (Plan 41)
```

**Finding 2 (LOW) — INDEX.md: Add `docs/agent-interface.md` to Document Registry**

Proposed new row:
```markdown
| `docs/agent-interface.md` | Headless CLI contract — flags, JSONL serialization, exit codes for non-interactive/MCP use | When building MCP wrapper, CI pipelines, or integrating bonsai programmatically |
```

**Finding 3 (LOW) — memory.md Research/ broken links**

The 6 Research/ links in `station/agent/Core/memory.md` (lines 87–92) point to `station/Research/` which does not exist. Options:
1. Leave as-is — treat as forward-references until Research/ scaffolding is built (Backlog line 109)
2. Note them as aspirational with a `<!-- planned -->` comment
3. Remove or rewrite as plain text until the files exist

## Notes for Next Run

- The big documentation gap is `internal/nonint/` — Plan 41 is the largest ship since Plan 41 ran. If code-index is updated before the next run, clear Finding 1.
- No navigation link regressions. All 52 checked paths exist.
- Research/ link gap is a pre-existing known issue (Backlog line 109). Unless Research/ scaffolding ships, it will appear every run.
- git log shows no commits since early July 2026 — if that holds, next run will only need to check for any work done between now and 2026-09-24.
