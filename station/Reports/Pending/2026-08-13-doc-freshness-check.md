---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-13
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (findings flagged — no autonomous fixes applied per procedure)
- **Duration:** ~10 min
- **Files Read:** 10 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/docs/agent-interface.md`, `/home/user/Bonsai/internal/nonint/result.go`, `/home/user/Bonsai/internal/nonint/runner.go`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Glob, Grep, Bash (git log, ls, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation vs. recent git history
- **Action:** Read station/INDEX.md, station/CLAUDE.md, station/Playbook/Status.md, station/Playbook/Roadmap.md, station/code-index.md. Ran `git log --oneline` to identify recent commits. Checked git history for the last 7 days (no commits) and last 30 days (multiple commits from Plan 41).
- **Result:** No commits in the last 7 days. Most recent significant work: **Plan 41 — Headless CLI Contract + MCP-Ready Cores** (shipped 2026-06-16, ~59 days ago). Plan 41 introduced the `internal/nonint/` package (headless cores for all 4 mutating commands) and the `docs/agent-interface.md` contract doc. Neither is reflected in `station/code-index.md` or `station/INDEX.md`.
- **Issues:** Documentation drift from Plan 41 work — see Findings Summary.

### Step 2: Check INDEX.md accuracy
- **Action:** Read station/INDEX.md and cross-checked tech stack, agent count, CLI command count, and architecture diagram against actual code.
- **Result:**
  - Tech Stack: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template).
  - Key Metrics: accurate — 6 agent types (backend, devops, frontend, fullstack, security, tech-lead); 8 CLI commands (init, add, remove, list, catalog, update, guide, validate).
  - Architecture diagram: mentions `internal/validate/` and `internal/wsvalidate/` but **does NOT include `internal/nonint/`** — the new headless-cores layer added in Plan 41.
  - Document Registry: does NOT list `docs/agent-interface.md` — the headless CLI contract doc produced in Plan 41 Phase 5.
- **Issues:** Two drift items in INDEX.md (see Findings 2 and 3).

### Step 3: Check navigation links
- **Action:** Enumerated all files listed in station/CLAUDE.md nav tables (Core, Protocols, Workflows, Skills, Sensors, Routines) and compared against files actually present on disk via Glob. Also checked for files present but NOT listed.
- **Result:**
  - **All linked files resolve.** Every path in station/CLAUDE.md nav tables points to an existing file. No broken links found.
  - **Two unlisted files discovered:** `agent/Workflows/plan-grilling.md` and `agent/Skills/critic-agent-prompts.md` exist on disk but are not in the CLAUDE.md nav tables. (Bubbletea sub-files in `agent/Skills/bubbletea/` are deliberately sub-files of the parent `bubbletea.md` entry — not a gap.)
  - **code-index.md: `internal/nonint/` entirely missing** — the package has 9 source files (`config.go`, `events.go`, `nonint.go`, `remove.go`, `result.go`, `runner.go`, `update.go`, plus test files) but no section in code-index.md.
- **Issues:** Three items (see Findings 1, 4, 5).

### Step 4: Report findings
- **Action:** Compiled findings into the Findings Summary table below, classified by severity. Per procedure, no documentation was modified — all drift items flagged for user decision.
- **Result:** 5 findings identified (1 high, 2 medium, 2 low).
- **Issues:** none.

### Step 5: Update dashboard
- **Action:** Updated the "Doc Freshness Check" row in station/agent/Core/routines.md.
- **Result:** Last Ran → 2026-08-13, Next Due → 2026-08-20, Status → done.
- **Issues:** none.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package (Plan 41 — 9 source files, headless cores for all 4 mutating commands + exit-code contract + Result type) has **no section in code-index.md**. This is the biggest architectural addition since Plan 35 and is completely absent from the developer reference. | `station/code-index.md` | Flagged — propose adding a new `## NonInt / Headless Cores` section covering `result.go`, `runner.go`, `events.go`, `update.go`, `remove.go`, exit constants, and `EmitJSONL`. |
| 2 | MEDIUM | `station/INDEX.md` Architecture Overview diagram omits `internal/nonint/`. The box list shows `internal/validate/` and `internal/wsvalidate/` but not the newer headless-cores layer. | `station/INDEX.md`, Architecture Overview section | Flagged — propose inserting `internal/nonint/` into the diagram between `internal/config/` and `internal/generate/`. |
| 3 | MEDIUM | `docs/agent-interface.md` (Plan 41 Phase 5 contract doc — the canonical source of truth for headless flag/serialization/exit-code contract) is not listed in the `station/INDEX.md` Document Registry. Agents and CI script maintainers need to be able to find this file. | `station/INDEX.md`, Document Registry table | Flagged — propose adding a row: `docs/agent-interface.md` / "Headless CLI contract — flags, JSONL/JSON serializations, exit codes for driving Bonsai non-interactively" / "When integrating Bonsai with CI, MCP, or a headless hub". |
| 4 | LOW | `agent/Workflows/plan-grilling.md` exists on disk but is not listed in the station/CLAUDE.md Workflows nav table. It has no activation trigger visible to the agent at session start. | `station/CLAUDE.md`, Workflows table | Flagged — propose adding a row (activation: "Pressure-testing a plan for weaknesses, risks, or underspecification before dispatching to an agent"). |
| 5 | LOW | `agent/Skills/critic-agent-prompts.md` exists on disk but is not listed in the station/CLAUDE.md Skills nav table. | `station/CLAUDE.md`, Skills table | Flagged — propose adding a row (activation: "Constructing adversarial or stress-test prompts to evaluate agent output or plan resilience"). |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **Finding 1 (HIGH):** Add `internal/nonint/` section to `station/code-index.md`. This package is the core of Plan 41 and has no developer reference. Recommend prioritizing — it's the substrate for the upcoming Plan 42 MCP server.
- **Finding 2 (MEDIUM):** Update `station/INDEX.md` Architecture Overview to include `internal/nonint/` in the package list.
- **Finding 3 (MEDIUM):** Add `docs/agent-interface.md` to the `station/INDEX.md` Document Registry so agents know where the headless CLI contract lives.
- **Finding 4 (LOW):** Decide whether `plan-grilling.md` should be added to the CLAUDE.md Workflows table (it appears active — used in Plan 41 grilling sessions per commit history).
- **Finding 5 (LOW):** Decide whether `critic-agent-prompts.md` should be added to the CLAUDE.md Skills table.

---

## Notes for Next Run

- Last doc refresh cycle (Plan 37, 2026-05-07) synced code-index.md for Plans up to 35. Plans 40 and 41 have both shipped since then with no follow-up doc refresh — the next run should verify whether a Plan 42 MCP server has shipped and needs code-index coverage.
- `Plans/Active/41-headless-cli-contract.md` appears to still be in `Plans/Active/` even though Status.md shows it as Done — the Status Hygiene routine should archive it.
- HOMEBREW_TAP_TOKEN PAT rotation deadline likely passed (~2026-07-15). This was flagged in today's Backlog Hygiene report and remains unresolved — worth re-verifying on next run.
