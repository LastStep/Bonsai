---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-30
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
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/CLAUDE.md` (system-injected), `station/code-index.md`, `station/Playbook/Status.md`, `docs/agent-interface.md`, `station/Logs/RoutineLog.md`, `station/Reports/Pending/` (directory listing), `catalog/agents/` (directory listing)
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Reports/Pending/2026-08-30-doc-freshness-check.md` (this file)
- **Tools Used:** Read, Bash (git log, ls, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

**Step 1 — Scan project documentation against recent git history:**
Ran `git log --oneline --since="7 days ago"` and `git log --oneline -10` to understand recent changes. Found 2 commits from this session (backlog-hygiene routine) and several earlier commits from Plan 41 (headless CLI contract, PRs #120–#125, shipped 2026-06-16). Also found Plan 40 (Odysseus platform integration, 2026-06-13). These represent ~116 days of gap since docs were last checked (last ran: 2026-05-04). Scanned INDEX.md, CLAUDE.md, code-index.md, and Status.md.

**Step 2 — Check INDEX.md accuracy:**
Tech stack verified: Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, Go text/template, embed.FS — all correct. CLI commands count: 8 (init/add/remove/list/catalog/update/guide/validate) — correct. Agent types: 6 (tech-lead, fullstack, backend, frontend, devops, security) — confirmed by listing `catalog/agents/`. However: `internal/nonint` package (created by Plan 41) is NOT mentioned in the INDEX.md Architecture Overview. This is a gap — the architecture diagram shows only 6 internal packages but there are now 7.

**Step 3 — Check navigation links:**
Verified all links in `station/CLAUDE.md` navigation tables by checking file existence. All referenced files exist — no broken links found. However, two files in the agent workspace are NOT listed in CLAUDE.md:
- `agent/Skills/critic-agent-prompts.md` — exists in Skills directory, not in Skills nav table
- `agent/Workflows/plan-grilling.md` — exists in Workflows directory, not in Workflows nav table

**Step 4 — Report findings:**
5 findings identified — see table below. All flagged for user decision per routine procedure (audit-only).

**Step 5 — Update dashboard:**
Updated routines.md dashboard: Doc Freshness Check row — Last Ran → 2026-08-30, Next Due → 2026-09-06, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `critic-agent-prompts.md` exists in `agent/Skills/` but is not listed in CLAUDE.md Skills navigation table — undiscoverable via nav | `station/CLAUDE.md` Skills section | Flagged for user decision |
| 2 | Medium | `plan-grilling.md` exists in `agent/Workflows/` but is not listed in CLAUDE.md Workflows navigation table — undiscoverable via nav | `station/CLAUDE.md` Workflows section | Flagged for user decision |
| 3 | Medium | `internal/nonint` package (created by Plan 41 headless CLI contract, PRs #120–#125) is absent from INDEX.md Architecture Overview and missing from `code-index.md` — architecture diagram shows 6 internal packages but there are 7 | `station/INDEX.md`, `station/code-index.md` | Flagged for user decision |
| 4 | Low | `docs/agent-interface.md` (the headless CLI contract doc shipped by Plan 41) is not referenced in INDEX.md Document Registry — agents may not know it exists | `station/INDEX.md` Document Registry | Flagged for user decision |
| 5 | Low | `code-index.md` has stale line numbers for cmd/ functions after Plan 41 expanded those files with headless non-interactive variants. Confirmed stale: `bonsai remove` listed at `:34` (actual: 67), `bonsai list` at `:18` (actual: 39), `bonsai update` at `:19` (actual: 51), `runRemoveItem()` at `:290` (actual: 428), `runRemoveItemAction()` at `:565` (actual: 703). New headless functions unlisted: `runRemoveAgentNonInteractive`, `runRemoveItemNonInteractive`, `runUpdateNonInteractive` | `station/code-index.md` | Flagged for user decision |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **Add `critic-agent-prompts.md` to CLAUDE.md Skills table** — add a row with trigger scenario and link. Skill file: `station/agent/Skills/critic-agent-prompts.md`

2. **Add `plan-grilling.md` to CLAUDE.md Workflows table** — add a row with trigger scenario and link. Workflow file: `station/agent/Workflows/plan-grilling.md`

3. **Document `internal/nonint` in INDEX.md and code-index.md** — add it to the Architecture Overview in INDEX.md (between wsvalidate and tui lines), and add a new section to code-index.md describing the headless runner, Result types, and exit code constants.

4. **Add `docs/agent-interface.md` to INDEX.md Document Registry** — add a row: `docs/agent-interface.md` | Headless CLI contract — flags, JSONL serialization, exit codes for non-interactive/MCP use | When driving Bonsai from an agent, CI, or MCP wrapper.

5. **Refresh code-index.md line numbers for cmd/ files** — run `grep -n "func run" cmd/remove.go cmd/list.go cmd/update.go` and update stale entries. Also add entries for `runRemoveAgentNonInteractive`, `runRemoveItemNonInteractive`, `runUpdateNonInteractive`.

## Notes for Next Run

- Findings 1 and 2 (unlisted skill/workflow) are low-urgency but will accumulate if not addressed — worth batching into a doc-refresh plan.
- Finding 3 and 5 relate to Plan 41 aftermath. A targeted doc-refresh plan (similar to Plan 37 for the v0.4.0 era) would resolve findings 3–5 cleanly.
- Next run due 2026-09-06. If none of the flagged items are resolved by then, they should be escalated.
- The 116-day gap since last doc check (2026-05-04 to 2026-08-30) spans Plans 38–41 and multiple releases (v0.4.1, v0.4.2, v0.4.3) — a backlog of drift accumulated. Recommend running this routine on schedule going forward.
