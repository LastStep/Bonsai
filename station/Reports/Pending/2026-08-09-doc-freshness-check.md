---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-09
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~10 min
- **Files Read:** 10 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/41-headless-cli-contract.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/docs/agent-interface.md`, `/home/user/Bonsai/station/Reports/Pending/2026-08-09-backlog-hygiene.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, grep for function positions, ls for catalog counts and directory listings)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read station/INDEX.md, station/CLAUDE.md (from system context), station/Playbook/Status.md, station/Playbook/Roadmap.md, station/code-index.md. Ran `git log` to collect all commits since last run (2026-05-04) and to identify new features or config changes not yet reflected in docs.
- **Result:** 19 commits since 2026-05-04 — concentrated in two batches: Plan 40 (2026-06-13, freeze schemas + root-relative scaffolding + validate pass + memory-routing docs) and Plan 41 (2026-06-16, headless CLI contract: 5 phases, all mutating commands now have headless cores, `list --json`, `docs/agent-interface.md` contract doc). No activity after 2026-06-16.
- **Issues:** code-index.md has not been updated since Plan 37 (2026-05-07). Plan 41 added significant new function entries and shifted line numbers throughout most cmd/ files and internal/generate/generate.go.

### Step 2: Check INDEX.md accuracy
- **Action:** Cross-checked each section of INDEX.md against the current codebase: tech stack table, key metrics (agent types, catalog item count, CLI command count), architecture diagram, and document registry.
- **Result:**
  - Tech stack (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS): **accurate**.
  - Agent types (6): **accurate** — `ls catalog/agents/` returns 6 entries.
  - CLI commands (8): **accurate** — init, add, remove, list, catalog, update, guide, validate.
  - Catalog items (~50): **slightly low but within range** — actual count is 53 ability items (18 skills, 10 workflows, 4 protocols, 13 sensors, 8 routines) + 6 agent types = 59. "~50" is still a defensible approximation.
  - Architecture diagram: **accurate** — lists all internal packages including internal/validate/ and internal/wsvalidate/.
  - Document Registry: **all entries resolve**.
  - Phase label `**Current phase:** Dogfooding & Polish`: **soft drift** — Plan 41 delivered the headless CLI contract described as MCP infrastructure, positioning the project between Phase 1 and Phase 2.
- **Issues:** Phase label is outdated (see Finding 3). Catalog item count could be updated to "~53" for precision, but within acceptable tolerance.

### Step 3: Check navigation links
- **Action:** Extracted all 50 linked file paths from station/CLAUDE.md navigation tables (Core, Bonsai Reference, Protocols, Workflows, Skills, Routines, Sensors, External References) and ran filesystem existence checks on each.
- **Result:** All 50 links resolve. No broken navigation links found.
- **Issues:** None.

### Step 4: Report findings
- **Action:** Identified 4 findings from comparing code-index.md line numbers and entries against actual source, checking Plans/Active/ against shipped status, and reviewing INDEX.md phase label. All findings are flagged for user decision — no documentation was auto-updated.
- **Issues:** See Findings Summary below.

### Step 5: Update dashboard
- **Action:** Updated routines.md dashboard row for Doc Freshness Check (Last Ran → 2026-08-09, Next Due → 2026-08-16, Status → done). Appended entry to RoutineLog.md.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | code-index.md line numbers are stale throughout cmd/ and internal/generate/ due to Plan 41 headless additions shifting line positions significantly (e.g., `runAdd()` documented at `:56`, actual `:73`; `runRemoveItem()` at `:290`, actual `:428`; `AgentWorkspace()` at `:1359`, actual `:1460`) | `station/code-index.md` | Flagged — not auto-corrected per routine procedure |
| 2 | MEDIUM | code-index.md is missing 7 headless functions added in Plan 41: `runAddNonInteractive()` (add.go:754), `runInitNonInteractive()` (init_flow.go:266), `runRemoveAgentNonInteractive()` (remove.go:288), `runRemoveItemNonInteractive()` (remove.go:317), `runUpdateNonInteractive()` (update.go:100), `renderListJSON()` (list.go:65), `loadConfigHeadless()` (remove.go:270) | `station/code-index.md` | Flagged — not auto-corrected per routine procedure |
| 3 | LOW | Plan 41 shipped 2026-06-16 but `41-headless-cli-contract.md` remains in `Plans/Active/` rather than `Plans/Archive/`; Plan 40 is also in Active/ with "Phase 4 HELD" and no clear archive/continue decision | `station/Playbook/Plans/Active/` | Flagged for user decision |
| 4 | LOW | INDEX.md phase label `**Current phase:** Dogfooding & Polish` predates Plan 41's headless CLI contract + MCP-ready cores; project is entering Phase 2 territory with Plan 42 (MCP server) as the next planned unit of work | `station/INDEX.md` | Flagged for user decision |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[MEDIUM] code-index.md line number refresh needed** — The code-index.md was last updated in Plan 37 (2026-05-07). Plans 40 and 41 added substantial new functions and shifted all existing line numbers. A sweep is needed to restore accuracy. Suggest: file a Plan or Backlog item for a code-index refresh, or run it as part of the next doc-bundle plan.

2. **[MEDIUM] code-index.md missing 7 headless functions from Plan 41** — The headless/non-interactive cores (`runAddNonInteractive`, `runInitNonInteractive`, `runRemoveAgentNonInteractive`, `runRemoveItemNonInteractive`, `runUpdateNonInteractive`, `renderListJSON`, `loadConfigHeadless`) are not in the code-index. These are significant new entry points that agents and maintainers may need to navigate to.

3. **[LOW] Plan 41 should be archived** — Plan 41 is fully shipped (all 5 phases, PRs #120–#125, 2026-06-16). Move `Plans/Active/41-headless-cli-contract.md` → `Plans/Archive/`. Also clarify Plan 40 status (Phase 4 HELD — is this deferred indefinitely or planned?).

4. **[LOW] INDEX.md phase label** — Consider updating from "Dogfooding & Polish" to reflect the current transition (Phase 1 complete, Phase 2 in progress). Low priority — does not affect agent operation.

## Notes for Next Run

- All navigation links are healthy — no broken links as of 2026-08-09.
- The primary ongoing debt is code-index.md staleness. After Plan 41 shipped, code-index.md needs a full line-number sweep + new headless function entries. This is a separate task, not suitable for autonomous execution.
- No new catalog abilities have been added since last run — the ~50 count in INDEX.md remains approximately accurate.
- If a Plan 42 (MCP server) ships before the next doc freshness check, the tech stack and architecture diagram in INDEX.md will need an update.
