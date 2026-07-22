---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-22
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
- **Files Read:** 9 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/agent/Core/identity.md`, `station/code-index.md`, `station/Logs/RoutineLog.md`, `station/CLAUDE.md` (via system-reminder)
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log append)
- **Tools Used:** `git log --since`, `git log --format`, `git show --stat`, `git diff --name-only`, `ls` (multiple dirs), file existence checks
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation vs recent git history
Compared station/ docs against commits since last run (2026-05-04). No commits exist in the last 7 days (most recent is 2026-06-16). Total commits since last run: ~44 commits across Plans 37–41 plus dependency bumps and hotfixes.

Key code changes since 2026-05-04:
- **Plan 37 (2026-05-07):** doc refresh bundle — code-index.md sync, INDEX.md Go drift fix
- **v0.4.1 (2026-05-07):** `bonsai completion` command (PR #78, external contribution from @mvanhorn)
- **Plans 38/39/v0.4.2 (2026-05-13):** `bonsai init/add --non-interactive --from-config` flags
- **v0.4.3 (2026-05-13):** sensor hook absolute-path bake fix
- **Plan 40 (2026-06-13):** frozen v1 schemas, root-relative scaffolding, project-level validate, memory-routing docs, guide Formats page; added `plan-grilling.md` workflow + `critic-agent-prompts.md` skill to station
- **Plan 41 (2026-06-16):** headless CLI contract — new `internal/nonint/` package, `docs/agent-interface.md`, `list --json`, `--yes/--from` (remove), `--non-interactive/--skip-conflicts` (update)

### Step 2 — Check INDEX.md accuracy
- Tech stack table: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, Homebrew/GitHub Releases).
- Architecture overview: accurate.
- **Drift found:** Key Metrics row shows "CLI commands: 8 (init, add, remove, list, catalog, update, guide, validate)" — `bonsai completion` (PR #78) makes 9. Count is stale.
- Document registry: accurate for station/ docs. `docs/agent-interface.md` (Plan 41 Phase 5) is a project-root doc not in scope of the registry, but worth noting.

### Step 3 — Check navigation links
Verified all 35 links in station/CLAUDE.md navigation tables (Core, Bonsai Reference, Protocols, Workflows, Skills, Routines, Sensors, External References). All resolve to real files — no broken links found.

Two files exist in agent/Workflows/ and agent/Skills/ that are NOT listed in the CLAUDE.md nav tables:
- `station/agent/Workflows/plan-grilling.md` (added Plan 40 session, 2026-06-13)
- `station/agent/Skills/critic-agent-prompts.md` (added same session)

Both have frontmatter noting "full Bonsai-catalog integration pending (Backlog)" — their omission from the nav table appears intentional, but the files are installed and active.

### Step 4 — Report findings
See Findings Summary below. Proposing updates but not executing — flagged for user decision.

### Step 5 — Update dashboard
Updated `station/agent/Core/routines.md` row for Doc Freshness Check: Last Ran → 2026-07-22, Next Due → 2026-07-29, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | INDEX.md Key Metrics shows 8 CLI commands; `bonsai completion` (PR #78) makes 9 | `station/INDEX.md` line 33 | Flagged for user |
| 2 | Medium | `internal/nonint/` package (Plan 41, 8+ source files) has no section in code-index.md | `station/code-index.md` | Flagged for user |
| 3 | Low | `agent/Workflows/plan-grilling.md` exists but is not listed in station/CLAUDE.md Workflows nav table | `station/CLAUDE.md` Workflows section | Flagged for user |
| 4 | Low | `agent/Skills/critic-agent-prompts.md` exists but is not listed in station/CLAUDE.md Skills nav table | `station/CLAUDE.md` Skills section | Flagged for user |
| 5 | Low | `docs/agent-interface.md` (Plan 41 Phase 5 contract doc) not referenced in INDEX.md document registry | `station/INDEX.md` | Flagged for user (low — registry focuses on station/ docs) |

**Navigation link audit: CLEAN — all 35 nav links resolve to real files.**

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

### F1 (Medium) — INDEX.md CLI command count stale
`station/INDEX.md` line 33: change `8 (init, add, remove, list, catalog, update, guide, validate)` → `9 (init, add, remove, list, catalog, update, guide, validate, completion)`.

### F2 (Medium) — code-index.md missing `internal/nonint/` package section
`station/code-index.md` has no entry for the `internal/nonint/` package introduced in Plan 41. Package contains: `config.go`, `events.go`, `nonint.go`, `remove.go`, `result.go`, `runner.go`, `update.go` (+ tests). Recommend adding a new section between Validate and Workspace-path Validation, covering the headless-core types (`Result`, `Event`, `Config`) and entry functions (`RunAdd`, `RunInit`, `RunRemove`, `RunUpdate`). This is a Plan 42 doc task or can be bundled into the next code-index refresh.

### F3 (Low) — `plan-grilling.md` not in CLAUDE.md nav table
`station/agent/Workflows/plan-grilling.md` is active and used but absent from the Workflows section of `station/CLAUDE.md`. Frontmatter says "full Bonsai-catalog integration pending (Backlog)". Recommend adding a row: `| Adversarial review of a drafted plan via 6 critic agents, looped to convergence | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |`. User decides whether to include now or wait for catalog integration.

### F4 (Low) — `critic-agent-prompts.md` not in CLAUDE.md nav table
Same situation as F3 — `station/agent/Skills/critic-agent-prompts.md` is active but not listed in the Skills table. Recommend adding a row with matching trigger and link. Defer if waiting for Bonsai-catalog integration.

### F5 (Low) — `docs/agent-interface.md` absent from INDEX.md registry
`docs/agent-interface.md` is the headless CLI contract doc (Plan 41, Phase 5). INDEX.md's document registry covers station/ files, so this may be by design. If the Tech Lead wants agents to know the contract doc exists, add a row to INDEX.md: `| docs/agent-interface.md | Headless CLI contract — flags, serializations, exit codes for non-interactive use | Plan 42 or MCP integration work |`.

---

## Notes for Next Run

- All nav links were clean this run — no broken links to repair.
- F1 (command count) and F2 (nonint/ package) are the highest-value fixes; both are straightforward edits.
- If Plan 42 (MCP server) ships before the next run, expect additional drift in INDEX.md (new command or package) and code-index.md.
- The 76-day gap since the last doc-freshness run (noted also in the Backlog Hygiene report) allowed multiple Plans (40 + 41) worth of drift to accumulate. Running on schedule (7 days) should keep findings to 1–2 items per run.
