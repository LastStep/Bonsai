---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-10
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~7 min
- **Files Read:** 12 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/CLAUDE.md` (system-reminder), `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`, `/home/user/Bonsai/station/agent/Skills/critic-agent-prompts.md`, `/home/user/Bonsai/cmd/completion.go`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls), Glob, Grep
- **Errors Encountered:** 0

## Procedure Walkthrough

**Step 1 — Scan project documentation and compare against recent git history (last 7 days):**

Git log shows only 2 commits in the last 7 days — both routine maintenance runs (status-hygiene `65a73c3`, backlog-hygiene `d757ec8`). All changes were confined to `station/` files (RoutineLog, Status, StatusArchive, Backlog, routines.md). No application code changes, no new features, services, or catalog items introduced in the window. No doc drift from recent code changes.

**Step 2 — Check INDEX.md accuracy:**

- **Tech stack table:** Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template — all accurate. ✅
- **Agent types count (6):** Verified against `catalog/agents/` — backend, devops, frontend, fullstack, security, tech-lead. ✅
- **Catalog items (~50):** No change since last freshness check. ✅
- **CLI commands count (8):** **DRIFT FOUND** — `cmd/completion.go` exists (added via PR #78, merged 2026-05-07, external contributor @mvanhorn). The `bonsai completion [bash|zsh|fish|powershell]` command is a real, fully-documented user-facing command. INDEX.md Key Metrics still reads "8 (init, add, remove, list, catalog, update, guide, validate)". Should be **9**.

**Step 3 — Check navigation links:**

Verified all links in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References sections). All 35 linked files/directories exist on disk. **Zero broken links found.**

Discovered two files that exist on disk but are not listed in the CLAUDE.md nav tables:
- `agent/Workflows/plan-grilling.md` — adversarial plan review workflow (6-critic loop)
- `agent/Skills/critic-agent-prompts.md` — companion prompt templates for plan-grilling

Both are intentionally omitted: Backlog P2 item tracks "Integrate plan-grilling as a first-class Bonsai catalog ability" — full nav-table wiring is deferred until catalog integration ships. These are not broken links; they are intentional unlisted items.

Also noted: `docs/agent-interface.md` was added as part of Plan 41 (headless CLI contract, 2026-06-16). It is the authoritative contract doc for the JSONL/exit-code interface. It is not listed in INDEX.md Document Registry; the registry only covers `station/` workspace files, so this omission may be intentional — but noting it for awareness.

**Step 4 — Report findings:**

Findings documented below. All flagged for user decision — no doc edits executed.

**Step 5 — Update dashboard:**

Updated `agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-07-10, Next Due → 2026-07-17, Status → done. ✅

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | LOW | INDEX.md CLI command count is 8 but should be 9 — `bonsai completion` added via PR #78 (2026-05-07) | `station/INDEX.md` Key Metrics table | Flagged for user review |
| 2 | LOW | `code-index.md` CLI Commands table omits `bonsai completion` (`cmd/completion.go:20 completionCmd`) | `station/code-index.md` | Flagged for user review |
| 3 | INFO | `agent/Workflows/plan-grilling.md` exists but not in CLAUDE.md Workflows nav table | `station/CLAUDE.md` | Known Backlog P2 (catalog integration pending) — no action |
| 4 | INFO | `agent/Skills/critic-agent-prompts.md` exists but not in CLAUDE.md Skills nav table | `station/CLAUDE.md` | Known Backlog P2 (catalog integration pending) — no action |
| 5 | INFO | `docs/agent-interface.md` (headless CLI contract, Plan 41) not in INDEX.md Document Registry | `station/INDEX.md` | Informational — registry covers station/ only; may be intentional |

## Proposed Updates (for User Decision)

**Finding 1 — INDEX.md CLI count:**
In `station/INDEX.md` Key Metrics table, change:
> `| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |`

to:
> `| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |`

**Finding 2 — code-index.md completion command:**
In `station/code-index.md` CLI Commands table, add a row after `bonsai validate`:
> `| \`bonsai completion\` | \`cmd/completion.go:20\` | \`completionCmd\` — generate shell completion scripts (bash/zsh/fish/powershell) |`

**Finding 5 — docs/agent-interface.md (optional):**
If the tech lead wants the agent to be able to reference the headless CLI contract, add to INDEX.md Document Registry:
> `| \`docs/agent-interface.md\` | Headless CLI contract — JSONL output schema, exit codes, \`--json\`/\`--from-config\` flags | When reviewing Plan 41 headless usage, MCP integration, or non-interactive CLI behavior |`

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[LOW] INDEX.md CLI count 8 → 9** — `bonsai completion` missing from Key Metrics. Quick fix (one-liner). Proposed update in Findings section above.
- **[LOW] code-index.md missing `completion` entry** — `cmd/completion.go` not in CLI Commands table. Quick fix (one row). Proposed update above.
- **[INFO] docs/agent-interface.md not in INDEX.md registry** — Optional; only needed if tech lead wants the agent to reference the headless CLI contract from INDEX.md.

## Notes for Next Run

- Last run was 2026-05-04 — this run was 66 days overdue. No significant drift found despite the gap (no code changes in the last 7 days; the larger Plan 41 changes from June were infrastructure files, not docs-breaking).
- `code-index.md` line numbers should be verified at next run — they drift whenever generate.go or catalog.go are modified. The current run did not do a line-number audit (no code changes to prompt it).
- The `plan-grilling` / `critic-agent-prompts` catalog integration (Backlog P2) remains the main pending nav-table gap. Watch for it to ship and then wire the nav table.
- HOMEBREW_TAP_TOKEN PAT expiry (flagged in backlog-hygiene 2026-07-10) is ~2026-07-15 — unrelated to docs freshness but worth noting as an upcoming ops event.
