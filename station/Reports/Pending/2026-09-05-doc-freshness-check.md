---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-05
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
- **Files Read:** 9 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/CLAUDE.md` (via system-reminder), `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, plus directory listings for `cmd/`, `catalog/`, `station/agent/`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

**Step 1 — Scan project documentation vs recent git history:**
Ran `git log --oneline --since="7 days ago"` — returned zero commits. No code changes in the last 7 days. Broadened to recent history: 20 most recent commits all dated 2026-06-16 (Plan 41, headless CLI contract, 5 phases shipped). Gap between last commit (2026-06-16) and today (2026-09-05) is ~81 days — well outside the 7-day window, so no new code drift to detect from the routine's primary comparison window.

**Step 2 — Check INDEX.md accuracy:**
Read `station/INDEX.md`. Verified Tech Stack table (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS) — all correct. Verified agent count: 6 — correct (backend, devops, frontend, fullstack, security, tech-lead). Verified catalog count ("~50"): actual count is 53 items (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines) — "~50" is a slight undercount but within approximation range. **CLI command count: INDEX.md says 8 — actual is 9.** `bonsai completion` was added in PR #78 (May 2026, @mvanhorn contribution) and `completion.go` exists in `cmd/` but is not reflected in INDEX.md or root CLAUDE.md's cmd/ file listing.

**Step 3 — Check navigation links in station/CLAUDE.md:**
Verified all links against actual filesystem:
- Core files (identity.md, memory.md, self-awareness.md, routines.md) — all present
- Protocol files (memory.md, scope-boundaries.md, security.md, session-start.md) — all present
- Workflow files (code-review, planning, pr-review, security-audit, session-logging, test-plan, session-wrapup, issue-to-implementation, routine-digest) — all present
- Skills files (planning-template, review-checklist, issue-classification, pr-creation, bubbletea, bonsai-model) — all present
- Routine files (backlog-hygiene, dependency-audit, doc-freshness-check, memory-consolidation, roadmap-accuracy, status-hygiene, vulnerability-scan) — all present
- Sensor scripts (context-guard, scope-guard-files, session-context, status-bar, routine-check, agent-review, dispatch-guard, subagent-stop-review, compact-recovery, statusline) — all present
- External refs (.bonsai.yaml, .bonsai/catalog.json, Playbook/Status.md, Roadmap.md, Standards/SecurityStandards.md, Plans/Active/, Backlog.md, Logs/KeyDecisionLog.md, Reports/Pending/, Reports/report-template.md) — all present

No broken links found.

**Step 4 — Additional findings noted:**
- `station/agent/Skills/critic-agent-prompts.md` exists but is not listed in station/CLAUDE.md skills nav table
- Plan 41 plan file (`41-headless-cli-contract.md`) remains in `Plans/Active/` — plan is fully shipped (all 5 phases merged 2026-06-16) and should be archived
- Plan 41 capabilities (headless mode, `--json` output, `docs/agent-interface.md` contract) not reflected in INDEX.md architecture overview

**Step 5 — Dashboard and log updated** (see sections below).

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | CLI command count stale: INDEX.md says "8 commands" — `bonsai completion` added in PR #78 (May 2026) brings actual count to 9 | `station/INDEX.md` Key Metrics table + root `CLAUDE.md` cmd/ listing | Flagged for user — proposed fix: update count to 9 and add `completion.go` to root CLAUDE.md listing |
| 2 | LOW | Plan 41 shipped but plan file still in `Plans/Active/` | `station/Playbook/Plans/Active/41-headless-cli-contract.md` | Flagged for user — proposed: move to `Plans/Archive/` |
| 3 | LOW | Catalog item count slightly low: INDEX.md says "~50" but actual is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines) | `station/INDEX.md` Key Metrics table | Flagged for user — minor, "~50" acceptable approximation; proposed: update to "~55" |
| 4 | LOW | `critic-agent-prompts.md` not in station/CLAUDE.md Skills nav table | `station/CLAUDE.md` Skills section, `station/agent/Skills/critic-agent-prompts.md` | Flagged for user — may be intentional (private skill); proposed: add nav entry if this skill should be discoverable |
| 5 | LOW | Headless CLI capabilities from Plan 41 not reflected in INDEX.md overview | `station/INDEX.md` Architecture Overview | Flagged for user — INDEX.md is high-level but Plan 42 (MCP server) may need this documented; proposed: add a note on headless/JSON mode |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[MEDIUM] CLI command count in INDEX.md and root CLAUDE.md is stale (8 → 9).**
   `bonsai completion [bash|zsh|fish|powershell]` was shipped in PR #78 and `cmd/completion.go` exists.
   - Proposed fix for `station/INDEX.md`: change "8 (init, add, remove, list, catalog, update, guide, validate)" to "9 (init, add, remove, list, catalog, update, guide, validate, completion)"
   - Proposed fix for root `CLAUDE.md`: add `├── completion.go  ← bonsai completion (shell autocompletion)` to cmd/ file listing

2. **[LOW] Plan 41 plan file should be archived.**
   `station/Playbook/Plans/Active/41-headless-cli-contract.md` — Plan 41 fully shipped 2026-06-16 (all 5 phases, PRs #120–#125). Move to `Plans/Archive/` alongside Plans 30–39.

3. **[LOW] `critic-agent-prompts.md` in station/agent/Skills/ not in nav table.**
   Verify intent: if this is an active skill the agent should load, add it to the station/CLAUDE.md Skills nav table with trigger description.

4. **[LOW] Consider updating INDEX.md "~50" catalog count to "~55".**
   Actual count is 53 across 5 ability types. Minor but the key metrics table should stay roughly accurate.

## Notes for Next Run

- No commits in the last ~81 days — if the gap continues to grow, the routine may want to extend its lookback window beyond 7 days when checking for code-vs-docs drift
- All navigation links are clean; no broken links to chase
- The 4 flagged items above are the primary backlog for this routine
