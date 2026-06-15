---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-15
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
- **Duration:** ~8 minutes
- **Files Read:** 12 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, `station/code-index.md`, `station/CLAUDE.md`, `/home/user/Bonsai/CLAUDE.md`, `docs/agent-interface.md`, `station/agent/Workflows/plan-grilling.md`, `station/agent/Skills/critic-agent-prompts.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `station/Reports/Pending/2026-06-15-doc-freshness-check.md` (this report), `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log append)
- **Tools Used:** Read, Bash (git log, ls, grep, wc), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation
**Action:** Read `station/INDEX.md`, `station/Playbook/Roadmap.md`, `station/Playbook/Status.md`, and compared against git history from the last 7 days (32 commits).

**Result:** Three significant feature areas from Plans 40 and 41 are not reflected in `station/INDEX.md` or `/home/user/Bonsai/CLAUDE.md`:

1. **Plan 41 — Headless CLI contract (shipped 2026-06-16):** New `internal/nonint` package with headless cores for `init`, `add`, `update`, `remove`. New flags: `--non-interactive`, `--from-config`, `--yes`/`-y`, `--skip-conflicts`, `--from`, `--delete-files`, `--json` (on `list`). New exit code contract. New `docs/agent-interface.md`. Neither `CLAUDE.md` (root or station) nor `code-index.md` mentions `internal/nonint/`.

2. **`completion` command added:** `/home/user/Bonsai/cmd/completion.go` exists. INDEX.md says "8 (init, add, remove, list, catalog, update, guide, validate)" — `completion` is a 9th command. Root `CLAUDE.md` project structure also omits it.

3. **`docs/` directory added:** `docs/agent-interface.md` (Plan 41 Phase 5) exists at `/home/user/Bonsai/docs/`. Not mentioned in `INDEX.md` Document Registry or in root `CLAUDE.md` project structure.

**Issues:** Documentation lags behind 2 shipped plans.

### Step 2 — Check INDEX.md accuracy
**Action:** Verified tech stack table, key metrics, architecture overview.

**Result:**
- **Tech stack table:** Accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS. No drift.
- **Key Metrics — CLI commands:** Says "8" — `completion` is now a 9th command. **Stale.**
- **Key Metrics — Catalog items:** Says "~50" — actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). "~50" is close enough, no action needed.
- **Key Metrics — Agent types:** Says "6" — still correct (tech-lead, fullstack, backend, frontend, devops, security).
- **Architecture overview:** Omits `internal/nonint/` package. **Stale.**
- **Document Registry:** Missing `docs/agent-interface.md`. **Stale.**
- **Phase / current phase:** "Dogfooding & Polish" — accurate given Plan 40 Phase 4 still held and no version tagged.

### Step 3 — Check navigation links
**Action:** Verified all links in `station/CLAUDE.md` navigation tables against real files.

**Result:** One broken link found: `.bonsai/catalog.json` referenced in the Bonsai Reference table as `[../.bonsai/catalog.json](../.bonsai/catalog.json)`. The `.bonsai/` directory does not exist inside `station/`; `catalog.json` lives at `/home/user/Bonsai/.bonsai/catalog.json`. The `../` relative path resolves to `/home/user/Bonsai/.bonsai/catalog.json` from `station/` — this IS the correct path, so the link is actually valid at the filesystem level. (Claude Code's markdown link resolution from `station/CLAUDE.md` at the station root would resolve `../.bonsai/catalog.json` to `/home/user/Bonsai/.bonsai/catalog.json` correctly.) Not broken.

All other links verified present:
- All Core files: identity.md, memory.md, self-awareness.md, routines.md ✓
- All Protocol files: memory.md, scope-boundaries.md, security.md, session-start.md ✓
- All Workflow files (9 entries): code-review, planning, pr-review, security-audit, session-logging, test-plan, session-wrapup, issue-to-implementation, routine-digest ✓
- All Skills files (6 entries): planning-template, review-checklist, issue-classification, pr-creation, bubbletea, bonsai-model ✓
- All Routine files (7): all present ✓
- All Sensors: agent-review, compact-recovery, context-guard, dispatch-guard, routine-check, scope-guard-files, session-context, status-bar, statusline, subagent-stop-review ✓

**Undocumented files found in workspace:**
- `station/agent/Workflows/plan-grilling.md` — installed but not in CLAUDE.md workflows nav table (marked as "Backlog, full Bonsai-catalog integration pending" per file header, so intentionally informal)
- `station/agent/Skills/critic-agent-prompts.md` — installed but not in CLAUDE.md skills nav table (same reason — adapted from ZenGarden, not yet a formal catalog item)

### Step 4 — Report findings
**Action:** Compiled findings. Per the procedure, these are flagged for user decision — not auto-applied.

**Key findings:**
1. `station/code-index.md` — `internal/nonint/` package (Plan 41) missing entirely. Needs a new section.
2. `station/code-index.md` — Line numbers for `cmd/remove.go` helpers are all stale (Plan 41 Phase 3 added ~140 lines before these functions). `runRemoveItem()` is at `:428`, not `:290`; `runRemoveItemAction()` is at `:703`, not `:565`; `agentItemList()` is at `:756`, not `:618`; `itemIsRequired()` is at `:805`, not `:667`; `itemDisplayName()` is at `:831`, not `:693`.
3. `station/code-index.md` — Line numbers for `cmd/add.go` helpers are all stale (Plan 41 added ~80 lines of headless core before these). `runAdd()` is at `:73`, not `:56`; `applyCinematicConflictPicks()` at `:344`, not `:309`; `installedSet()` at `:400`, not `:365`; `buildAddGrowAction()` at `:422`, not `:387`; `distributeAddItemPicks()` at `:605`, not `:570`; `availableAddItems()` at `:690`, not `:655`.
4. `station/code-index.md` — Line numbers for `cmd/init_flow.go` helpers are stale. `runInit()` is at `:35`, not `:27`; `buildGenerateAction()` at `:315`, not `:227`; `plantedSummary()` at `:389`, not `:301`; `scaffoldingToSoilOptions()` at `:404`, not `:316`.
5. `station/code-index.md` — Line numbers for `internal/generate/generate.go` key functions are stale. `Scaffolding()` is at `:401`, not `:360`; `SettingsJSON()` at `:564`, not `:473`; `WorkspaceClaudeMD()` at `:826`, not `:725`; `AgentWorkspace()` at `:1460`, not `:1359`; `RoutineDashboard()` at `:1111`, not `:1010`; `EnsureRoutineCheckSensor()` at `:1073`, not `:972`; `PathScopedRules()` at `:1265`, not `:1164`; `WorkflowSkills()` at `:1329`, not `:1228`.
6. `station/INDEX.md` — CLI commands count says "8" — should be "9 (init, add, remove, list, catalog, update, guide, validate, completion)".
7. `station/INDEX.md` — Architecture overview omits `internal/nonint/` package. Should add: `internal/nonint/ ← headless CLI cores (init/add/update/remove) + exit-code contract`.
8. `station/INDEX.md` — Document Registry missing `docs/agent-interface.md` (Plan 41 Phase 5 contract doc).
9. `/home/user/Bonsai/CLAUDE.md` — Project structure listing omits `internal/nonint/` package and `docs/` directory.

### Step 5 — Update dashboard
**Action:** Dashboard updated per procedure.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `internal/nonint/` package (Plan 41) missing from code index | `station/code-index.md` | Flagged — needs user-triggered update |
| 2 | Low | `cmd/remove.go` helper line numbers stale (off by ~138 lines) | `station/code-index.md` | Flagged — needs user-triggered update |
| 3 | Low | `cmd/add.go` helper line numbers stale (off by ~17-35 lines) | `station/code-index.md` | Flagged — needs user-triggered update |
| 4 | Low | `cmd/init_flow.go` helper line numbers stale (off by 8-88 lines) | `station/code-index.md` | Flagged — needs user-triggered update |
| 5 | Low | `internal/generate/generate.go` function line numbers stale | `station/code-index.md` | Flagged — needs user-triggered update |
| 6 | Low | CLI commands count "8" — should be "9" (completion command added) | `station/INDEX.md` | Flagged — needs user-triggered update |
| 7 | Low | Architecture overview omits `internal/nonint/` package | `station/INDEX.md` | Flagged — needs user-triggered update |
| 8 | Low | Document Registry missing `docs/agent-interface.md` | `station/INDEX.md` | Flagged — needs user-triggered update |
| 9 | Low | Project structure omits `internal/nonint/` and `docs/` | `/home/user/Bonsai/CLAUDE.md` | Flagged — needs user-triggered update |

## Errors & Warnings

None.

## Items Flagged for User Review

The following doc updates are proposed but NOT applied (per procedure — flag only, no execute):

**High value (code-index nonint section):**
- Add `internal/nonint/` section to `station/code-index.md` documenting: exit codes (`ExitOK=0`, `ExitInvalidConfig=2`, `ExitRuntime=3`, `ExitWrongCWDForInit=4`, `ExitConflict=5`), key types (`Result`, `FileEvent`, `SummaryEvent`), and headless runner functions.

**Medium value (INDEX.md updates):**
- Update CLI commands count from "8" to "9" and add `completion` to the list.
- Add `internal/nonint/` to the architecture overview diagram.
- Add `docs/agent-interface.md` to the Document Registry.

**Lower value (CLAUDE.md and line numbers):**
- Update root `CLAUDE.md` project structure to include `internal/nonint/` and `docs/`.
- Update stale line numbers in `code-index.md` for `cmd/remove.go`, `cmd/add.go`, `cmd/init_flow.go`, and `internal/generate/generate.go`.

**Note:** Line number drift in code-index.md is an expected consequence of active development. The semantics (what each function does) remain accurate. Line numbers are guidance, not contracts.

## Notes for Next Run

- Check whether Plan 40 Phase 4 (update-delivery) has shipped — if so, INDEX.md "current phase" and Roadmap.md need review.
- Check if `docs/` directory has grown (more pages added alongside `agent-interface.md`).
- Check if any new catalog items (skills, workflows, sensors, routines) were added since this run — catalog item count was 53.
- Check if the `completion` command was added to CLI commands count in INDEX.md (finding #6).
- Watch for `internal/nonint/` documentation status — whether a formal code-index section was added.
