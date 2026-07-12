---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-12
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
- **Files Read:** 14
  - `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`
  - `/home/user/Bonsai/station/INDEX.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/CLAUDE.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/code-index.md`
  - `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`
  - `/home/user/Bonsai/station/agent/Skills/critic-agent-prompts.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/docs/agent-interface.md`
  - `/home/user/Bonsai/CHANGELOG.md`
  - `/home/user/Bonsai/cmd/completion.go`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry appended)
- **Tools Used:** Read, Bash (git log, ls, grep), Glob, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history
Ran `git log --since="7 days ago"` and `git log --since="14 days ago"`. Only 2 commits in the last 7 days (both today — routine runs for backlog-hygiene and status-hygiene). Broadened to `git diff --stat HEAD~5 HEAD` to capture substantive code changes. Found major work since last check date (2026-05-04):

- **Plan 40** (2026-06-13): Frozen v1 schemas + root-relative scaffolding (memory + project-manifest), project-level validate pass, `bonsai guide formats` page, `docs/formats.md` updated.
- **Plan 41** (2026-06-16): Headless CLI contract — headless `update`/`remove`/`list --json` cores, `internal/nonint/` package, `docs/agent-interface.md` created (245 lines), exit-code `5` (`ExitConflict`).
- **PR #78** (2026-05-07): `bonsai completion [bash|zsh|fish|powershell]` from external contributor @mvanhorn — new 9th CLI command.

### Step 2 — Check INDEX.md accuracy
Read `station/INDEX.md`. Found 3 items of drift:

1. **CLI commands count**: Says "8 (init, add, remove, list, catalog, update, guide, validate)" — `completion` was merged after last check (PR #78, 2026-05-07), making it 9 commands.
2. **Architecture overview**: Missing `internal/nonint/` layer added in Plan 41. Currently lists `internal/validate/` and `internal/wsvalidate/` but not `internal/nonint/`.
3. **Document registry**: `docs/agent-interface.md` (headless CLI contract, 245 lines) not listed.

Current-phase label ("Dogfooding & Polish") still accurate. Tech stack table still accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea). Agent type count (6) still accurate. Catalog items (~50) still plausible.

### Step 3 — Check navigation links
Verified all links in `station/CLAUDE.md` navigation tables resolve to existing files:

- Core files: identity.md, memory.md, self-awareness.md — all present.
- Protocols: memory.md, scope-boundaries.md, security.md, session-start.md — all present.
- Workflows: 9 listed, all present. BUT `plan-grilling.md` exists in `agent/Workflows/` and is NOT listed.
- Skills: 6 listed, all present. BUT `critic-agent-prompts.md` exists in `agent/Skills/` and is NOT listed.
- Routines: all 7 listed files present.
- Sensors: all 10 listed files present.
- External refs: `.bonsai.yaml` and `.bonsai/catalog.json` both exist at project root.

The two unlisted files (`plan-grilling.md`, `critic-agent-prompts.md`) both carry the note "full Bonsai-catalog integration pending (Backlog)" — they were custom-added during Plan 40/41 grilling sessions and intentionally not yet wired into the nav table.

Also checked `code-index.md`: no entry for `bonsai completion` command or `internal/nonint/` package (both added after last doc check).

### Step 4 — Report findings
5 flagged items requiring user decision. No auto-fixes applied per procedure.

### Step 5 — Update dashboard
Set `last_ran` to 2026-07-12, `next_due` to 2026-07-19 in `agent/Core/routines.md`.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | CLI command count stale: says 8 commands, should be 9 — `completion` added PR #78 (2026-05-07) | `station/INDEX.md`, Key Metrics table | Flagged — propose update: change "8" → "9" and add `completion` to the list |
| 2 | MEDIUM | Architecture overview missing `internal/nonint/` package (added Plan 41, 2026-06-16) | `station/INDEX.md`, Architecture Overview | Flagged — propose adding a row: `internal/nonint/ ← headless CLI cores (pure Result functions, no prompts)` |
| 3 | MEDIUM | Document registry missing `docs/agent-interface.md` — the canonical headless CLI contract (245 lines) | `station/INDEX.md`, Document Registry | Flagged — propose adding a row for `docs/agent-interface.md` |
| 4 | LOW | `code-index.md` missing `bonsai completion` command entry in CLI Commands table | `station/code-index.md` | Flagged — propose adding row: `bonsai completion | cmd/completion.go:20 | completionCmd — generate shell completion scripts (bash/zsh/fish/powershell)` |
| 5 | LOW | `code-index.md` has no section for `internal/nonint/` package | `station/code-index.md` | Flagged — propose adding a new `## Nonint (internal/nonint/)` section covering Result type, RunnerOptions, ExitOK/ExitConflict constants, and per-command headless cores |
| 6 | INFO | `agent/Workflows/plan-grilling.md` exists but not listed in `station/CLAUDE.md` nav table | `station/CLAUDE.md`, Workflows table | Not flagged as a bug — file notes "Bonsai-catalog integration pending (Backlog)". Intentional interim state. |
| 7 | INFO | `agent/Skills/critic-agent-prompts.md` exists but not listed in `station/CLAUDE.md` nav table | `station/CLAUDE.md`, Skills table | Not flagged as a bug — same pending note. Intentional interim state. |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

**Findings 1–5 require user decision** (none of the proposed updates were applied — procedure says flag, don't execute):

1. **INDEX.md Key Metrics** — bump CLI command count 8 → 9, add `completion` to the parenthetical list.
2. **INDEX.md Architecture Overview** — add `internal/nonint/` to the package listing.
3. **INDEX.md Document Registry** — add a row for `docs/agent-interface.md` (headless CLI contract / AI integrator reference).
4. **code-index.md CLI Commands table** — add `bonsai completion` row.
5. **code-index.md** — add `internal/nonint/` section covering the headless core API surface.

Findings 6 and 7 are informational — no action needed unless you want to wire the two custom abilities into the nav table now.

---

## Notes for Next Run

- All 5 substantive drift items trace to Plans 40/41 (June 2026) — a 68-day gap since last check means this is a larger-than-usual backlog.
- If these INDEX.md and code-index.md updates are applied this session, the next run (2026-07-19) should be clean.
- Check if Plan 40 Phase 4 (held at June 2026) has shipped by then — if so, verify any scaffolding/validate changes are reflected in INDEX.md.
- Also confirm v0.5.0 has been tagged before the next run — CHANGELOG shows it as "Unreleased" as of today.
