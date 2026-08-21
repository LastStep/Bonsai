---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-21
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
- **Duration:** ~10 minutes
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/Playbook/Status.md`, `station/agent/Core/routines.md`, `station/code-index.md`, `station/CLAUDE.md` (via system context), `station/Logs/RoutineLog.md`, `/home/user/Bonsai/CLAUDE.md` (root), `station/Playbook/Plans/Active/*` (ls), `docs/` (ls)
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history

Read `station/INDEX.md`, `station/Playbook/Status.md`, `station/code-index.md`, and `station/CLAUDE.md`. Ran `git log --oneline --since="2026-05-04"` to enumerate all commits since last run (109 days, 46 commits).

Key shipped changes since 2026-05-04:
- **v0.4.2** — `bonsai init/add --non-interactive --from-config` (Plan 39, 2026-05-13)
- **v0.4.3 hotfix** — absolute path baking for sensor hooks (2026-05-13)
- **`bonsai completion` command** — PR #78 by @mvanhorn, merged 2026-05-07 (`cmd/completion.go`)
- **Plan 40 Phases 1–3** — frozen v1 schemas, root-relative scaffolding, project-level validate pass, memory-routing docs + guide Formats page (2026-06-13)
- **Plan 41 (all 5 phases)** — headless CLI contract: new `internal/nonint/` package, `--non-interactive`/`--yes`/`--from` flags across all mutating commands, `list --json`, `docs/agent-interface.md` contract doc (2026-06-16)

### Step 2 — Check INDEX.md accuracy

**Tech stack:** Still accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS). No drift.

**Agent types:** 6 confirmed (backend, devops, frontend, fullstack, security, tech-lead). Correct.

**Catalog items:** Reported as "~50" — 53 meta.yaml files found via `find`. "~50" remains approximately correct, no change needed.

**CLI command count:** **STALE** — listed as "8 (init, add, remove, list, catalog, update, guide, validate)" but `bonsai completion` was added in PR #78 (2026-05-07). Count should be 9 and the list should include `completion`.

**Architecture overview:** **STALE** — `internal/nonint/` package (added in Plan 41) is not listed in the architecture diagram or text.

### Step 3 — Check navigation links

Verified all links in `station/CLAUDE.md` navigation tables resolve to actual files on disk.

**Core (3/3 OK):** identity.md ✓, memory.md ✓, self-awareness.md ✓

**Protocols (4/4 OK):** memory.md ✓, scope-boundaries.md ✓, security.md ✓, session-start.md ✓

**Workflows (9/9 listed OK):** code-review.md ✓, planning.md ✓, pr-review.md ✓, security-audit.md ✓, session-logging.md ✓, test-plan.md ✓, session-wrapup.md ✓, issue-to-implementation.md ✓, routine-digest.md ✓

**Workflows — unlisted file detected:** `agent/Workflows/plan-grilling.md` exists on disk but is NOT in the CLAUDE.md navigation table. This workflow was added for the plan-grilling pipeline (commit `6995d4f`, 2026-06-13).

**Skills (6/6 listed OK):** planning-template.md ✓, review-checklist.md ✓, issue-classification.md ✓, pr-creation.md ✓, bubbletea.md ✓, bonsai-model.md ✓

**Skills — unlisted file detected:** `agent/Skills/critic-agent-prompts.md` exists on disk but is NOT in the CLAUDE.md navigation table.

**Routines (7/7 OK):** backlog-hygiene.md ✓, dependency-audit.md ✓, doc-freshness-check.md ✓, memory-consolidation.md ✓, roadmap-accuracy.md ✓, status-hygiene.md ✓, vulnerability-scan.md ✓

**Sensors (10/10 OK):** context-guard.sh ✓, scope-guard-files.sh ✓, session-context.sh ✓, status-bar.sh ✓, routine-check.sh ✓, agent-review.sh ✓, dispatch-guard.sh ✓, subagent-stop-review.sh ✓, compact-recovery.sh ✓, statusline.sh ✓

**External references:** `.bonsai/catalog.json` ✓ (exists), `.bonsai.yaml` ✓ (exists), `agent/Skills/bonsai-model.md` ✓

**No broken links found.**

### Step 4 — Report findings

See Findings Summary table below. All findings flagged for user decision — no docs were modified during this audit (audit-only per procedure).

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` — Doc Freshness Check row: `Last Ran` → 2026-08-21, `Next Due` → 2026-08-28, `Status` → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | CLI command count stale: says "8 (init, add, remove, list, catalog, update, guide, validate)" — `completion` added in PR #78 (2026-05-07) makes it 9 | `station/INDEX.md` Key Metrics table | Flagged — user to update |
| 2 | Medium | `internal/nonint/` package missing from architecture overview — added in Plan 41 (2026-06-16); headless CLI cores live there | `station/INDEX.md` Architecture Overview | Flagged — user to update |
| 3 | Low | `agent/Workflows/plan-grilling.md` exists on disk but has no row in the Workflows navigation table — added commit `6995d4f` 2026-06-13 | `station/CLAUDE.md` Workflows table | Flagged — user to add row |
| 4 | Low | `agent/Skills/critic-agent-prompts.md` exists on disk but has no row in the Skills navigation table | `station/CLAUDE.md` Skills table | Flagged — user to add row |
| 5 | Low | `bonsai completion` command not listed in CLI Commands table in code-index.md | `station/code-index.md` CLI Commands | Flagged — user to add row |
| 6 | Low | `internal/nonint/` package undocumented in code-index.md — Plan 41 added config.go, events.go, nonint.go, result.go, runner.go, remove.go, update.go | `station/code-index.md` | Flagged — user to add section |
| 7 | Low | Root CLAUDE.md project structure missing `cmd/completion.go` and `internal/nonint/` — same drift as INDEX.md/code-index.md | `/home/user/Bonsai/CLAUDE.md` Project Structure | Flagged — user to update |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

All 7 findings above are flagged for user decision — this is an audit-only routine. Recommended priority:

**Do now (medium — affects agent understanding of capabilities):**
1. `station/INDEX.md` — update CLI command count from 8 to 9, add `completion` to list, add `internal/nonint/` to architecture section
2. `station/CLAUDE.md` — add rows for `plan-grilling.md` (Workflows) and `critic-agent-prompts.md` (Skills)

**Do next (low — developer reference only):**
3. `station/code-index.md` — add `bonsai completion` CLI row and `internal/nonint/` section
4. `/home/user/Bonsai/CLAUDE.md` (root) — add `cmd/completion.go` and `internal/nonint/` to Project Structure

**Proposed edits for #1 (INDEX.md):**
```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```
Architecture section — add line under `internal/wsvalidate/`:
```
internal/nonint/      ← headless non-interactive cores (init/add/update/remove) + exit/event contract
```

**Proposed edits for #2 (station/CLAUDE.md Workflows table):**
```
| Grilling a plan adversarially before dispatch; Running the 6-critic review pipeline | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

**Proposed edits for #2 (station/CLAUDE.md Skills table):**
Determine trigger description for `critic-agent-prompts.md` first (read file), then add row.

---

## Notes for Next Run

- All navigation links are valid. No broken links to chase.
- `internal/nonint/` is a new architectural layer that should be documented in both INDEX.md and code-index.md for future agent context.
- `docs/agent-interface.md` was added in Plan 41 Phase 5 — a machine-readable CLI contract for MCP/agent consumers. Currently not referenced in station docs; assess whether it belongs in INDEX.md Document Registry once Plan 42 (MCP server) is underway.
- Both Plans 40 and 41 are still in `Plans/Active/` — Plan 40 (Phase 4 HELD) is legitimately active; Plan 41 is shipped (66 days ago). Status Hygiene routine flagged this separately.
