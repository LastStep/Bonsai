---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-20
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
- **Duration:** ~8 minutes
- **Files Read:** 11 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/CLAUDE.md` (system-reminder), `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`, `station/agent/Skills/critic-agent-prompts.md`, `station/agent/Workflows/plan-grilling.md`, `/home/user/Bonsai/CLAUDE.md`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log append)
- **Tools Used:** `git log --oneline --since="2026-05-04"` (50 commits since last run), `ls` (cmd/, internal/generate/, internal/tui/, agent/Skills/, agent/Workflows/, agent/Sensors/, agent/Core/, agent/Protocols/, agent/Routines/), `grep -n` (line number verification across cmd/add.go, cmd/root.go, cmd/remove.go, cmd/list.go, cmd/update.go, cmd/catalog.go, cmd/guide.go, cmd/validate.go, cmd/init_flow.go, internal/generate/generate.go, internal/catalog/catalog.go, internal/generate/list_snapshot.go, cmd/completion.go)
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read station/INDEX.md, Playbook/Status.md, Playbook/Roadmap.md; compared against `git log --oneline --since="2026-05-04"` (returned 50 commits across Plans 40, 41, and dependency bumps).
- **Result:** INDEX.md tech stack, key metrics, and architecture overview are accurate. Roadmap.md Phase 1 is fully checked; future phases (2–4) match current reality. Status.md is current and accurate (Plan 41 + Plan 40 results correctly recorded). The major codebase change since last run was Plan 41 (headless CLI contract, PRs #120–125) which added noninteractive cores to every mutating command and a `list --json` mode.
- **Issues:** Plan 41 shifted line numbers in cmd/ files and internal/generate/generate.go significantly, and added several new source files not reflected in station/code-index.md. See Step 2 and Finding #1.

### Step 2: Check INDEX.md accuracy
- **Action:** Verified tech stack table (Go version, frameworks), key metrics (agent count, CLI command count, catalog items), and architecture diagram against codebase state.
- **Result:** All INDEX.md content is accurate:
  - Go 1.25+ ✓ (correct, upgraded in Plan 36)
  - Cobra, Huh, LipGloss, BubbleTea ✓
  - 6 agent types ✓ (tech-lead, fullstack, backend, frontend, devops, security — confirmed via `ls catalog/agents/`)
  - 8 CLI commands ✓ (init, add, remove, list, catalog, update, guide, validate — `completion` is present via completion.go but hidden from help output via `HiddenDefaultCmd = true`, so 8 is defensible)
  - ~50 catalog items ✓ (approximate, no precise count needed)
  - Architecture diagram ✓ (references all major internal packages correctly)
- **Issues:** None for INDEX.md itself.

### Step 3: Check navigation links
- **Action:** Verified all nav links in station/CLAUDE.md by listing the referenced directories and checking for presence of every linked file.
- **Result:**
  - Core files (identity.md, memory.md, self-awareness.md) ✓ all exist
  - Protocol files (memory.md, scope-boundaries.md, security.md, session-start.md) ✓ all exist
  - Workflow files in nav table ✓ all exist; but `plan-grilling.md` exists in `agent/Workflows/` and is NOT in the nav table
  - Skill files in nav table ✓ all exist; but `critic-agent-prompts.md` exists in `agent/Skills/` and is NOT in the nav table
  - Bonsai reference files (.bonsai/catalog.json, .bonsai.yaml, agent/Skills/bonsai-model.md) ✓ all exist
  - Routine files ✓ all exist and match the routines table
  - Sensor files ✓ all 10 sensor .sh files exist
- **Issues:** Two unlisted files (see Findings #2 and #3).

### Step 4: Report findings
- **Action:** Compiled all drift findings into this report with severity classifications.
- **Result:** 3 findings total — 1 high, 2 medium/low. All are flag-for-user items; no doc edits executed (audit-only routine).
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for "Doc Freshness Check" — Last Ran → 2026-07-20, Next Due → 2026-07-27, Status → done.
- **Result:** Done.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `code-index.md` — all line numbers stale after Plan 41 + Plan 40; also missing `list_snapshot.go` section and headless nonint functions | `station/code-index.md` | Flagged for user — recommend refresh sweep (similar to Plan 37) |
| 2 | MEDIUM | `station/CLAUDE.md` Workflows nav table missing `plan-grilling.md` (added 2026-06-13) | `station/CLAUDE.md` | Flagged for user — add nav row |
| 3 | MEDIUM | `station/CLAUDE.md` Skills nav table missing `critic-agent-prompts.md` (added 2026-06-13) | `station/CLAUDE.md` | Flagged for user — add nav row |
| 4 | LOW | Root `CLAUDE.md` project structure tree missing 5 new source files added since last sweep: `internal/generate/{catalog_snapshot_unix.go, catalog_snapshot_unix_test.go, catalog_snapshot_windows.go, list_snapshot.go}` + `cmd/completion.go` | `/home/user/Bonsai/CLAUDE.md` | Flagged for user (recurring — also flagged 2026-05-04 run) |

---

## Finding #1 Detail — code-index.md line number drift

Plan 41 (headless CLI contract, PRs #120–125, merged 2026-06-16) added noninteractive core functions to every mutating command, shifting all function bodies down by 17–100+ lines. Plan 40 (Phases 1–3) also added code to generate.go before Plan 41.

**CMD command functions (index → actual):**
| Function | File | Index says | Actual |
|----------|------|-----------|--------|
| `runInit()` | cmd/init_flow.go | `:27` | `:35` |
| `runAdd()` | cmd/add.go | `:56` | `:73` |
| `runList()` | cmd/list.go | `:18` | `:39` |
| `runUpdate()` | cmd/update.go | `:19` | `:51` |
| `runCatalog()` | cmd/catalog.go | `:23` | `:39` |
| `runGuide()` | cmd/guide.go | `:44` | `:44` (no drift) |
| `runValidate()` | cmd/validate.go | `:43` | `:43` (no drift) |

**root.go helpers (minor drift, ~1–2 lines):**
| Function | Index says | Actual |
|----------|-----------|--------|
| `loadCatalog()` | `:45` | `:46` |
| `requireConfig()` | `:53` | `:54` |
| `mustCwd()` | `:64` | `:65` |
| `Execute()` | `:83` | `:84` |
| `buildConflictSteps()` | `:104` | `:105` |
| `applyConflictPicks()` | `:150` | `:151` |

**add.go helpers:**
| Function | Index says | Actual |
|----------|-----------|--------|
| `applyCinematicConflictPicks()` | `:309` | `:344` |
| `installedSet()` | `:365` | `:400` |
| `buildAddGrowAction()` | `:387` | `:422` |
| `distributeAddItemPicks()` | `:570` | `:605` |
| `availableAddItems()` | `:655` | `:690` |

**remove.go helpers:**
| Function | Index says | Actual |
|----------|-----------|--------|
| `runRemoveItem()` | `:290` | `:428` |
| `runRemoveItemAction()` | `:565` | `:703` |
| `agentItemList()` | `:618` | `:756` |
| `itemIsRequired()` | `:667` | `:805` |
| `itemDisplayName()` | `:693` | `:831` |

**generate.go functions:**
| Function | Index says | Actual |
|----------|-----------|--------|
| `Scaffolding()` | `:360` | `:401` |
| `SettingsJSON()` | `:473` | `:564` |
| `WorkspaceClaudeMD()` | `:725` | `:826` |
| `EnsureRoutineCheckSensor()` | `:972` | `:1073` |
| `RoutineDashboard()` | `:1010` | `:1111` |
| `PathScopedRules()` | `:1164` | `:1265` |
| `WorkflowSkills()` | `:1228` | `:1329` |
| `AgentWorkspace()` | `:1359` | `:1460` |

**catalog.go functions:**
| Function | Index says | Actual |
|----------|-----------|--------|
| `DisplayNameFrom()` | `:49` | `:50` |
| `New(fsys)` | `:242` | `:286` |
| `loadItems()` | `:346` | `:390` |
| `loadSensors()` | `:397` | `:441` |
| `loadRoutines()` | `:448` | `:492` |
| `loadScaffolding()` | `:499` | `:543` |
| `loadAgents()` | `:516` | `:560` |

**Missing from code-index entirely:**
- `internal/generate/list_snapshot.go` — `ListSnapshot` / `ListAgent` types + `SerializeJSON()` function (Plan 41 Phase 4 — list --json)
- Headless nonint functions: `runAddNonInteractive()` (add.go:754), `runInitNonInteractive()` (init_flow.go:266), `runUpdateNonInteractive()` (update.go:100), `runRemoveAgentNonInteractive()` (remove.go:288), `runRemoveItemNonInteractive()` (remove.go:317)
- Per-agent generate variants: `SettingsJSONForAgent()` (generate.go:578), `PathScopedRulesForAgent()` (generate.go:1278), `WorkflowSkillsForAgent()` (generate.go:1338)

---

## Finding #2 Detail — plan-grilling.md missing from Workflows nav

`station/agent/Workflows/plan-grilling.md` exists and has a well-formed header (frontmatter tags: `[workflow, planning, grilling]`, source: adapted from ZenGarden 2026-06-13). It is the adversarial plan review workflow — dispatches 6 critic agents and loops to convergence. It was hand-added during Plan 41 prep (plan-grilling commits visible in git log) but never registered in `station/CLAUDE.md`'s Workflows nav table.

Suggested nav row:
```
| Adversarially reviewing a drafted plan before dispatch — running 5 prose critics + 1 empirical Reality check, looping to convergence | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

---

## Finding #3 Detail — critic-agent-prompts.md missing from Skills nav

`station/agent/Skills/critic-agent-prompts.md` exists (frontmatter tags: `[skill, planning, grilling]`, source: adapted from ZenGarden 2026-06-13). It contains the verbatim prompts consumed by `plan-grilling.md` — one per critic agent. Also hand-added and not registered in the Skills nav table.

Suggested nav row:
```
| Dispatching plan-grilling critic agents — verbatim prompts for the 6 critic agent types (5 prose + Reality) | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

---

## Finding #4 Detail — root CLAUDE.md missing new source files

The root `Bonsai/CLAUDE.md` project structure tree (under `internal/generate/`) does not list these files added since the last doc sweep (Plan 37, 2026-05-07):

In `internal/generate/`:
- `catalog_snapshot_unix.go` — Unix-specific snapshot helper (Platform split in Plan 36 hotfix for Windows cross-compile)
- `catalog_snapshot_unix_test.go` — test for above
- `catalog_snapshot_windows.go` — Windows-specific snapshot helper (same split)
- `list_snapshot.go` — `ListSnapshot` / `SerializeJSON` for `bonsai list --json` (Plan 41 Phase 4)

In `cmd/`:
- `completion.go` — `bonsai completion [bash|zsh|fish|powershell]` shell completion (PR #78, @mvanhorn)

Note: This is a recurring finding (also flagged in 2026-04-21 and 2026-05-04 runs). Root CLAUDE.md is not generated by Bonsai (it's the developer-agent instructions), so it requires a manual refresh. This finding was promoted to Backlog P2 after the 2026-05-04 run; check if that item was actioned.

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **[HIGH] code-index.md refresh** — All ~30 line number references are stale (typically 35–100 lines off) and the headless CLI section from Plan 41 is entirely missing. Recommend a code-index refresh sweep (Tier 1 quick patch, similar to Plan 37). Low-risk, purely additive.
- **[MEDIUM] station/CLAUDE.md Workflows nav** — Add row for `agent/Workflows/plan-grilling.md`. The suggested row text is in Finding #2 above. Quick manual edit.
- **[MEDIUM] station/CLAUDE.md Skills nav** — Add row for `agent/Skills/critic-agent-prompts.md`. Suggested row in Finding #3. Quick manual edit.
- **[LOW] Root CLAUDE.md project structure tree** — Add 5 missing source file entries. Recurring item. If the Backlog P2 entry for "root-CLAUDE.md routine procedure tweak" (filed 2026-04-21) has not been actioned, this is the time.

---

## Notes for Next Run

- code-index.md line numbers drift every major plan — consider whether a lighter-weight index format (function names only, no line numbers) would be more durable. Or add code-index.md refresh to the generate.go automation.
- plan-grilling.md and critic-agent-prompts.md are marked "full Bonsai-catalog integration pending (Backlog)" in their frontmatter — once they're in the catalog, the CLAUDE.md nav rows would be auto-generated. Track whether this was promoted.
- The `bubbletea/` directory inside `agent/Skills/` is a slash-command content directory (`.claude/skills/bubbletea/`), not an undocumented skill — `bubbletea.md` is the actual skill file and it's correctly listed in the nav table. No action needed.
