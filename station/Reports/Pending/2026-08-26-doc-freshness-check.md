---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-26
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
- **Files Read:** 8 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md` (header only), `/home/user/Bonsai/station/agent/Skills/critic-agent-prompts.md` (header only)
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log append)
- **Tools Used:** Read, Bash (git log, ls, git show), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

**Step 1 — Scan project documentation vs recent git history.**
Ran `git log --oneline -20` to review recent commits. No commits in the last 7 days (last commit was 2026-06-16, shipping Plan 41). Reviewed the two most recent plans:
- **Plan 41** (2026-06-16): Headless CLI contract — added `internal/nonint/` package (~20 files) and `docs/agent-interface.md` + `docs/formats.md`. Major new infrastructure.
- **Plan 40** (2026-06-13): Odysseus Platform Integration — frozen schemas, root-relative scaffolding, validate hardening.

**Step 2 — Check INDEX.md accuracy.**
Tech stack table (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template) is accurate. Key Metrics are accurate (6 agent types verified by `ls catalog/agents/`, 8 CLI commands correct). Architecture overview lists internal packages correctly except `internal/nonint/` is entirely absent — added in Plan 41. The Document Registry has no row for `docs/` directory (also added in Plan 41 timeframe).

**Step 3 — Check navigation links.**
Verified all files referenced in `station/CLAUDE.md` navigation tables:
- Core: `identity.md`, `memory.md`, `self-awareness.md` — all present.
- Protocols: `memory.md`, `scope-boundaries.md`, `security.md`, `session-start.md` — all present.
- Workflows: `code-review.md`, `planning.md`, `pr-review.md`, `security-audit.md`, `session-logging.md`, `test-plan.md`, `session-wrapup.md`, `issue-to-implementation.md`, `routine-digest.md` — all present.
- Skills: `planning-template.md`, `review-checklist.md`, `issue-classification.md`, `pr-creation.md`, `bubbletea.md`, `bonsai-model.md` — all present. (`bonsai-model.md` was flagged as broken in 2026-05-04; it now resolves — resolved.)
- External refs: `../.bonsai/catalog.json`, `../.bonsai.yaml` — both confirmed present.
- All Routines links in dashboard table resolve correctly.
- Two unlisted files found in watched directories: `agent/Workflows/plan-grilling.md` and `agent/Skills/critic-agent-prompts.md` — both self-note "pending full Bonsai-catalog integration (Backlog)"; omission appears intentional.

**Step 4 — Findings compiled and flagged below (no doc edits — audit only).**

**Step 5 — Dashboard updated.**

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `internal/nonint/` package added in Plan 41 (2026-06-16) but absent from architecture overview | `station/INDEX.md` — Architecture Overview | Flagged for user |
| 2 | Medium | `internal/nonint/` package has no entry in code-index.md; package contains ~20 files (headless cores, JSONL contract, exit-code runner) | `station/code-index.md` | Flagged for user |
| 3 | Low | `docs/` directory (`agent-interface.md`, `formats.md`, `cli.md`, `concepts.md`, `custom-files.md`, `quickstart.md`) not referenced in Document Registry | `station/INDEX.md` — Document Registry | Flagged for user |
| 4 | Info | `plan-grilling.md` (Workflows) and `critic-agent-prompts.md` (Skills) exist in agent directories but are not in CLAUDE.md navigation tables — both files self-note Backlog pending integration | `station/CLAUDE.md` | No action needed — Backlog item already tracks this |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

**Finding 1 & 2 — internal/nonint/ undocumented (Medium):**
Plan 41 shipped a full `internal/nonint/` package with headless cores for every mutating command (init/add/update/remove), a JSONL event bus, config, result shapes, exit-code runner, and a contract test sweep. This package is not mentioned anywhere in `station/INDEX.md` or `station/code-index.md`. The code-index drift here is significant — the nonint package is the headless contract surface and would be a primary navigation target for anyone working on Plan 42 (MCP server, fast-follow from Plan 41 per Status.md).

Proposed update for `INDEX.md` Architecture Overview (add one line to the internal packages list):
```
internal/nonint/    ← headless command cores + JSONL event bus + exit-code contract (Plan 41)
```

Proposed new section for `code-index.md` (after the wsvalidate section, before TUI):
```
## Non-Interactive / Headless (`internal/nonint/`) — Plan 41
…table of types and functions from nonint.go, result.go, events.go, runner.go, config.go, update.go, remove.go…
```
(Full table values would require reading each file — recommend delegating to a code agent or updating inline.)

**Finding 3 — docs/ directory unregistered (Low):**
`docs/` now holds the canonical public-facing documentation (agent-interface contract, formats, quickstart, concepts). The Document Registry in INDEX.md has no entry for it. Users navigating via INDEX.md would not discover `docs/agent-interface.md` even though it is the definitive reference for Plan 42 (MCP server) work.

Proposed addition to Document Registry table:
```
| `docs/agent-interface.md` | Headless CLI contract — per-command flags, JSONL/JSON serialization, exit codes | When reasoning about headless/MCP integration |
| `docs/` | Public-facing user documentation (quickstart, concepts, CLI reference, formats) | When writing or reviewing user-facing docs |
```

## Notes for Next Run

- `bonsai-model.md` broken-link from 2026-05-04 is now resolved — no action needed.
- The two previously unflagged issues (code-index drift, INDEX arch drift) from 2026-05-07 Backlog Hygiene round are now formally captured here with specific proposed updates.
- Next run (~2026-09-02): check if `internal/nonint/` section was added to code-index.md and if docs/ was added to Document Registry. Also watch for Plan 42 (MCP server) — will likely add new packages/commands requiring another round of updates.
