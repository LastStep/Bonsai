---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-14
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (41 days ago, 6 cycles overdue)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (findings identified, flagged for user; no doc edits made — procedure is read-only, propose-only)
- **Duration:** ~8 minutes
- **Files Read:** 19
  - `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`
  - `/home/user/Bonsai/station/INDEX.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/CLAUDE.md`
  - `/home/user/Bonsai/station/code-index.md`
  - `/home/user/Bonsai/station/CLAUDE.md`
  - `/home/user/Bonsai/internal/catalog/catalog.go`
  - `/home/user/Bonsai/internal/generate/generate.go`
  - `/home/user/Bonsai/internal/validate/validate.go`
  - `/home/user/Bonsai/internal/tui/harness/harness.go`
  - `/home/user/Bonsai/internal/tui/styles.go`
  - `/home/user/Bonsai/catalog/scaffolding/manifest.yaml`
  - `/home/user/Bonsai/station/Playbook/Plans/Archive/37-doc-refresh-bundle.md`
  - `/home/user/Bonsai/station/agent/Skills/bonsai-model.md`
  - `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`
  - `/home/user/Bonsai/station/agent/Skills/critic-agent-prompts.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/cmd/completion.go`
  - `/home/user/Bonsai/cmd/validate.go`
- **Files Modified:** 2
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, grep, ls), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation vs recent git history

Ran `git log --since="2026-05-04"` to identify commits since the last run (41 days). Found 7 source code commits of note. The dominant change is **Plan 40 (Odysseus platform integration, v0.5.0)** executed 2026-06-13, which shipped in 3 phases:

- **Phase 1** (`1e715c7`): Added root-relative scaffolding, `Slugify()`, `yamlScalar()`, new `ScaffoldingItem.RootRelative` field, and new catalog scaffolding items (`project-manifest`, `memory`). Added ~40 lines to `catalog.go` and ~120 lines to `generate.go`.
- **Phase 2** (`a540fdd`): Added project-level validate pass to `validate.Run()` with 6 new `Category` constants (`invalid_manifest`, `missing_manifest`, `invalid_note`, `unresolved_relation`, `symlink_escape`, `memory_index_too_large`).
- **Phase 3** (`2aef7fd`): Added `docs/formats.md` (memory-note schema + project manifest spec), updated memory protocol, added Formats tab to `bonsai guide`.

Also found:
- `2eae9d4` (2026-05-07): Added `completion` subcommand (shell completions for bash/zsh/fish/powershell).
- `584b82b` (2026-05-13): `fix(generate)` absolute path baking for sensor hooks (v0.4.3).
- `a5a4185` (2026-05-07): Plan 37 doc-refresh-bundle executed — refreshed `code-index.md` line numbers. **This resolved the prior HIGH finding on code-index being stale.**

### Step 2 — Check INDEX.md accuracy

Read `station/INDEX.md`. Findings:
- **Tech stack** — all entries accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, embed.FS).
- **CLI commands count** — still says `8 (init, add, remove, list, catalog, update, guide, validate)`. The `completion` subcommand was added 2026-05-07 and is user-visible (`bonsai completion [bash|zsh|fish|powershell]`). **Drift: count should be 9.**
- **Agent types** — 6. Verified: backend, devops, frontend, fullstack, security, tech-lead. Accurate.
- **Catalog items** — `~50`. Actual: 18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines = 53. The `~50` approximation is still defensible; no material drift.
- **Architecture diagram** — `cmd/` annotation still lists 8 commands (missing `completion`). Minor drift.
- **New scaffolding items** (`project-manifest`, `memory`) not mentioned anywhere in INDEX.md or CLAUDE.md Key Concepts — this is new capability from Plan 40.

### Step 3 — Check navigation links

**station/CLAUDE.md navigation table:**
- All Core links verified: `identity.md`, `memory.md`, `self-awareness.md` — all exist.
- **Bonsai Reference:** `agent/Skills/bonsai-model.md` link — **RESOLVED**. File exists at `station/agent/Skills/bonsai-model.md` (size 11239 bytes, last modified 2026-06-10). This was the MEDIUM broken-link finding from the prior run.
- All Protocol links verified: `memory.md`, `scope-boundaries.md`, `security.md`, `session-start.md` — all exist.
- All Workflow links verified — all 9 files exist.
- All Skills links verified — all 6 files exist. **Note:** `plan-grilling.md` (workflow) and `critic-agent-prompts.md` (skill) exist on disk in their respective directories but are NOT in the nav table. These are custom station items adapted from ZenGarden, with full Bonsai-catalog integration pending (per Backlog 2026-06-13 entry). **Not a broken link; missing nav entries** are a separate finding.
- All Sensor links verified — all 10 sensor `.sh` files exist.
- All Routine links verified — all 7 routine `.md` files exist.

**Root CLAUDE.md Key Concepts:**
- `Scaffolding` key concept says: `(INDEX, Playbook, Logs, Reports)`. Since Plan 40, two new optional scaffolding items exist: `project-manifest` and `memory`. Description is now incomplete.

### Step 4 — Report findings

See Findings Summary table below.

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` Doc Freshness Check row: Last Ran → 2026-06-14, Next Due → 2026-06-21, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `code-index.md` line numbers stale — `catalog.go` and `generate.go` functions shifted by ~40–100 lines due to Plan 40 Phase 1 additions. All 7 catalog.go entries wrong; 12 generate.go entries wrong. | `station/code-index.md` — `## Catalog` and `## Generator` sections | Flagged. Propose new Plan or inline fix. |
| 2 | MEDIUM | `code-index.md` validate section describes "Six detection categories" but Plan 40 Phase 2 added 6 more, bringing the total to 12 categories. Also missing mention of project-level audit pass. | `station/code-index.md` — `## Validate` section | Flagged. Should update category count and list new categories. |
| 3 | MEDIUM | `station/INDEX.md` and `station/code-index.md` file layout do not mention new scaffolding items `project-manifest` (`.bonsai/project.yaml`) and `memory` (`MEMORY.md`, `Memory/decisions/`, `Memory/notes/`). | `station/INDEX.md` (key metrics / arch), `station/code-index.md` (file layout) | Flagged. New opt-in scaffold items from Plan 40. |
| 4 | MEDIUM | `station/INDEX.md` CLI command count is `8`; completion was added 2026-05-07. Should be `9`. Architecture diagram also missing `completion`. | `station/INDEX.md` line 33 and line 63 | Flagged. Minor correction needed. |
| 5 | MEDIUM | Root `CLAUDE.md` Key Concepts section describes Scaffolding as `(INDEX, Playbook, Logs, Reports)` — omits the two new Plan 40 items (`project-manifest`, `memory`). Also omits `root_relative` field in the bullet. | `/home/user/Bonsai/CLAUDE.md` line 129 | Flagged. Needs update to reflect new scaffolding scope. |
| 6 | LOW | `station/CLAUDE.md` Workflows nav table does not list `plan-grilling.md`; Skills nav table does not list `critic-agent-prompts.md`. Both exist on disk and are functional. Backlog already tracks full catalog integration (2026-06-13). | `station/CLAUDE.md` Workflows + Skills tables | Flagged. Nav gap until catalog integration shipped. User should decide: add temporary entries or wait for catalog integration. |
| 7 | RESOLVED | Broken nav link `agent/Skills/bonsai-model.md` — flagged HIGH in prior run. | Resolved — file exists (2026-06-10). | No action needed. |
| 8 | RESOLVED | `code-index.md` line numbers stale across cmd/ + generate.go (pre-Plan 37) — flagged HIGH in prior run. Plan 37 executed 2026-05-07 and refreshed them. | Resolved (Plan 37 archived). | No action needed (new drift introduced by Plan 40). |
| 9 | RESOLVED | `INDEX.md` CLI count and arch diagram drift — flagged LOW in prior run. Partially resolved (Plan 37 bumped Go 1.24+ → 1.25+). Remaining drift is `completion` command (Finding #4). | Partially resolved. | Finding #4 covers remaining gap. |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[HIGH] code-index.md line number drift (catalog.go + generate.go)** — Plan 40 Phase 1 shifted both files by ~40-100 lines. The prior Plan 37 refresh is now stale. Recommend a new inline Tier-1 plan (same shape as Plan 37) to re-grep and update. Impact: developer navigation jumps to wrong lines.

2. **[MEDIUM] code-index.md validate section outdated** — Now describes 6 categories; actual is 12. Should add new Plan 40 Phase 2 categories and note the project-level audit pass.

3. **[MEDIUM] New scaffolding items undocumented** — `project-manifest` and `memory` scaffolding items from Plan 40 not reflected in INDEX.md, CLAUDE.md Key Concepts, or code-index file layout. These are opt-in items so no user projects are broken, but docs are incomplete.

4. **[MEDIUM] CLI command count drift** — `completion` command exists since 2026-05-07 but INDEX.md still says 8 commands. Simple one-line fix.

5. **[LOW] plan-grilling + critic-agent-prompts missing from nav** — Files exist and are usable but not discoverable from CLAUDE.md nav. User can decide: add temporary nav entries now, or wait for Backlog item "Integrate plan-grilling as first-class Bonsai catalog ability."

## Notes for Next Run

- Plan 40 Phase 4 (update-delivery path for existing projects) is still on hold — deferred by user. When shipped, it will add new CLI behavior to document.
- The `completion` subcommand exists but is not in the architecture diagram or key metrics — suggest bundling the fix with the next doc sweep.
- Line numbers in code-index.md will continue drifting as long as Plan 40 additions sit unaccounted. A fresh Plan 37-style refresh (catalog.go + generate.go + validate.go sections) is the appropriate fix.
- If the `plan-grilling` catalog integration ships (Backlog P2 Group A 2026-06-13), station/CLAUDE.md nav tables will need updating and the LOW finding here will auto-resolve.
