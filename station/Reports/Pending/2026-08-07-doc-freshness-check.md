---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-07
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
- **Files Read:** 9 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/agent/Core/routines.md`, `station/CLAUDE.md`, `station/code-index.md`, `docs/agent-interface.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard updated), `station/Logs/RoutineLog.md` (log entry appended)
- **Tools Used:** Read, Bash (git log, file existence checks, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read station/INDEX.md, station/Playbook/Status.md, station/Playbook/Roadmap.md. Retrieved git log for the period 2026-05-04 → 2026-08-07 (the window since last run). Identified 30+ commits spanning Plans 38–41, v0.4.1–v0.4.3, and various station housekeeping.
- **Result:** Key changes since last run: (a) PR #78 added `bonsai completion` command (2026-05-07); (b) Plan 40 Phase 3 created `docs/formats.md`; (c) Plan 41 Phase 5 created `docs/agent-interface.md` and added headless CLI contract (`internal/nonint/`) — these docs-level additions are not reflected in station documentation. Roadmap, tech stack, and architecture description remain accurate.
- **Issues:** 4 documentation drift items found (detailed below).

### Step 2: Check INDEX.md accuracy
- **Action:** Compared INDEX.md Tech Stack table, Key Metrics, Architecture Overview, and Document Registry against the actual codebase state (go.mod, cmd/ listing, catalog/ counts, internal/ structure, docs/ directory).
- **Result:**
  - Tech stack (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS) — **accurate**.
  - Agent types (6) — **accurate** (backend, devops, frontend, fullstack, security, tech-lead).
  - CLI commands claim: **stale** — says "8 (init, add, remove, list, catalog, update, guide, validate)". The `completion` command was added via PR #78 (2026-05-07), making it 9 total.
  - Catalog item count "~50" — **still approximately correct** (actual: 53 = 18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). No update needed.
  - `docs/` directory: **not mentioned anywhere in INDEX.md** — Document Registry and Architecture Overview both omit it entirely.
  - `internal/nonint/` package: **not mentioned in INDEX.md architecture diagram** — added by Plan 41.
- **Issues:** 2 items flagged (CLI count, docs/ directory).

### Step 3: Check navigation links
- **Action:** Verified all 49 hyperlinked paths in station/CLAUDE.md navigation tables — Core, Bonsai Reference, Protocols, Workflows, Skills, Routines, Sensors, and External References. Also checked links referenced from agent/Core/, agent/Protocols/, agent/Workflows/, agent/Skills/ headings.
- **Result:** All 49 links resolve to real files. No broken or dangling references found.
- **Issues:** None.

### Step 4: Report findings + code-index.md check
- **Action:** Reviewed code-index.md CLI Commands table against actual cmd/ directory. Checked for `completion` command and `internal/nonint/` package coverage.
- **Result:**
  - `bonsai completion` command (cmd/completion.go) is **missing** from the CLI Commands table in code-index.md.
  - `internal/nonint/` package (Plan 41, headless CLI contract, pure-function cores) is **not documented** in code-index.md — it's a substantial new package.
  - All other code-index.md entries (catalog, config, generate, validate, wsvalidate, tui sub-packages) remain accurate.
- **Issues:** 2 items flagged (both code-index.md omissions).

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Low | CLI commands count stale — says "8" but `completion` makes it 9; list omits `completion` | `station/INDEX.md` — Key Metrics table | Flagged for user update |
| 2 | Medium | `docs/` directory not referenced — 7 files incl. `agent-interface.md`, `cli.md`, `concepts.md`, `custom-files.md`, `formats.md`, `quickstart.md`, `README.md` | `station/INDEX.md` — Document Registry + Architecture Overview both omit it | Flagged for user update |
| 3 | Low | `bonsai completion` missing from CLI Commands table in code-index.md (`cmd/completion.go` exists) | `station/code-index.md` — CLI Commands section | Flagged for user update |
| 4 | Medium | `internal/nonint/` package (Plan 41 — headless CLI contract, pure-function cores, exit/event contract) missing from code-index.md | `station/code-index.md` — no section for nonint/ | Flagged for user update |

## Proposed Updates (for user decision)

**Finding 1 — INDEX.md Key Metrics:**
```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```

**Finding 2 — INDEX.md Document Registry (add row):**
```
| `docs/` | Published user-facing documentation — agent-interface contract, CLI reference, concepts, quickstart, guide formats | When sharing with users or referencing headless CLI contract |
```
And update the Architecture Overview diagram to include `docs/` between go.mod and cmd/.

**Finding 3 — code-index.md CLI Commands table (add row):**
```
| `bonsai completion` | `cmd/completion.go` | `completionCmd` — shell completion for bash/zsh/fish/powershell |
```

**Finding 4 — code-index.md (add new section after Workspace-path Validation):**
```
## Headless CLI Contract (`internal/nonint/`) — Plan 41

Pure-function cores for all mutating commands. No prompts, no os.Exit, no stdout from inside the core.
CLI adapter serializes Result to JSONL; a future MCP adapter will use the same Result.

| Type / Function | File | Purpose |
|-----------------|------|---------|
| ExitOK / ExitInvalidConfig / ExitRuntime / ExitWrongCWDForInit / ExitConflict | runner.go | Exit-code constants (canonical source of truth) |
| RunInit / RunAdd / RunUpdate / RunRemove | runner.go | Pure-function headless cores for each mutating command |
| Result | runner.go | Structured result type (events, conflicts, errors) returned by each core |
```
(Exact lines to be confirmed by inspecting internal/nonint/ — the above is illustrative.)

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review
All 4 findings above require user decision before updates are applied (procedure: propose only, do not execute). Recommended priority:
1. **Finding 4** (Medium) — Add `internal/nonint/` section to code-index.md — most useful for agent navigation
2. **Finding 2** (Medium) — Add `docs/` to INDEX.md Document Registry and Architecture Overview
3. **Findings 1 & 3** (Low) — Quick one-liner fixes in INDEX.md + code-index.md for `completion`

## Notes for Next Run
- Check whether `bonsai mcp` server has been added (mentioned in Plan 41 as a fast-follow Plan 42) — if shipped, it will need INDEX.md and code-index.md additions.
- `internal/nonint/` section in code-index.md should include exact line numbers once added.
- No broken navigation links observed this run — link health is good.
