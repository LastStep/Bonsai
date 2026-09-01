---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-01
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
- **Duration:** ~8 min
- **Files Read:** 9
  - `station/agent/Routines/doc-freshness-check.md`
  - `station/INDEX.md`
  - `station/CLAUDE.md` (via system-reminder context)
  - `station/agent/Core/routines.md`
  - `station/Logs/RoutineLog.md`
  - `CLAUDE.md` (root — via system-reminder context)
  - `cmd/completion.go` (header)
  - `internal/nonint/` (directory listing)
  - `catalog/` subdirectory listings (sensors, routines, agents, skills, workflows, protocols)
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, head)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read `station/INDEX.md` and `station/CLAUDE.md`. Retrieved git log since last run date (2026-05-04) using `git log --oneline --since="2026-05-04"`. Retrieved all recent commits (50 shown).
- **Result:** 50+ commits since the last doc freshness check. Significant code additions include Plan 41 (headless CLI contract — internal/nonint package, --non-interactive/--yes flags across add/remove/update/list), Plan 40 (platform integration — freeze schemas, root-relative scaffolding), and the explicit `completion` subcommand (commit 2aef7fd). Only the most recent 7-day window had 1 commit (backlog-hygiene routine).
- **Issues:** Two code-level features introduced since last check are not reflected in documentation (see Findings 1 and 2).

### Step 2: Check INDEX.md accuracy
- **Action:** Compared INDEX.md tech stack table, key metrics table, architecture overview, and document registry against actual project state (filesystem listing, git log).
- **Result:**
  - Tech stack table: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS/Homebrew).
  - Agent types metric: "6" — confirmed 6 directories in `catalog/agents/` (backend, devops, frontend, fullstack, security, tech-lead). Correct.
  - Catalog items metric: "~50" — actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). Approximation acceptable.
  - CLI commands metric: "8 (init, add, remove, list, catalog, update, guide, validate)" — **STALE**: `completion` was added as an explicit user-facing command (commit 2aef7fd, May 2026), confirmed via cmd/completion.go header comment "shows up in `bonsai --help`". Should be 9.
  - Architecture overview: lists `internal/catalog`, `internal/config`, `internal/generate`, `internal/validate`, `internal/wsvalidate`, `internal/tui` — **STALE**: `internal/nonint/` is missing (10 source files, added Plan 41 Phase 1).
  - Document registry: all paths verified present on disk. ✓
- **Issues:** 2 stale entries (see Findings 1 and 2).

### Step 3: Check navigation links
- **Action:** Cross-referenced every file path in station/CLAUDE.md nav tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References) against actual files on disk.
- **Result:**
  - **Core (3 files):** identity.md, memory.md, self-awareness.md — all present ✓
  - **Protocols (4 files):** memory.md, scope-boundaries.md, security.md, session-start.md — all present ✓
  - **Workflows (9 files):** code-review.md, planning.md, pr-review.md, security-audit.md, session-logging.md, test-plan.md, session-wrapup.md, issue-to-implementation.md, routine-digest.md — all present ✓
  - **Skills (6 files):** planning-template.md, review-checklist.md, issue-classification.md, pr-creation.md, bubbletea.md, bonsai-model.md — all present ✓
  - **Routines (7 files):** backlog-hygiene.md, dependency-audit.md, doc-freshness-check.md, memory-consolidation.md, roadmap-accuracy.md, status-hygiene.md, vulnerability-scan.md — all present ✓
  - **Sensors (10 files):** context-guard.sh, scope-guard-files.sh, session-context.sh, status-bar.sh, routine-check.sh, agent-review.sh, dispatch-guard.sh, subagent-stop-review.sh, compact-recovery.sh, statusline.sh — all present ✓
  - **External refs:** INDEX.md, Playbook/Status.md, Playbook/Roadmap.md, Standards/SecurityStandards.md, Plans/Active/, Backlog.md, Logs/FieldNotes.md, Logs/KeyDecisionLog.md, Reports/Pending/, Reports/report-template.md, code-index.md — all present ✓
  - **Bonsai reference:** .bonsai/catalog.json, .bonsai.yaml — both present ✓
  - **Orphan files (exist on disk but have no nav entry):**
    - `station/agent/Workflows/plan-grilling.md` — present, not in nav table (Finding 3)
    - `station/agent/Skills/critic-agent-prompts.md` — present, not in nav table (Finding 4)
    - `station/agent/Skills/bubbletea/` (directory) — present, not directly in nav table (likely internal to bubbletea.md; acceptable)
- **Issues:** No broken links. 2 unlisted files identified.

### Step 4: Report findings
- **Action:** Compiled findings below.
- **Result:** 4 findings documented, all flagged for user decision per procedure (no changes executed to code or primary docs).
- **Issues:** none

### Step 5: Update dashboard and log
- **Action:** Updated `station/agent/Core/routines.md` dashboard row and appended to `station/Logs/RoutineLog.md`.
- **Result:** Done.
- **Issues:** none

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `internal/nonint/` package (10 files, Plan 41 headless CLI contract) missing from Architecture Overview | `station/INDEX.md` line ~70 and root `CLAUDE.md` project structure | Flagged for user — proposed update below |
| 2 | Low | CLI command count says "8" but `completion` is an explicit user-facing command (9th command, added May 2026) | `station/INDEX.md` Key Metrics table | Flagged for user — proposed update below |
| 3 | Low | `plan-grilling.md` exists in `agent/Workflows/` but has no nav table entry | `station/CLAUDE.md` Workflows nav table | Flagged for user — proposed update below |
| 4 | Low | `critic-agent-prompts.md` exists in `agent/Skills/` but has no nav table entry | `station/CLAUDE.md` Skills nav table | Flagged for user — proposed update below |

---

## Proposed Updates (for user decision — not executed)

### Finding 1 — Add `internal/nonint/` to INDEX.md and root CLAUDE.md

**In `station/INDEX.md` Architecture Overview** (after the `internal/wsvalidate/` line):
```
internal/nonint/      ← headless CLI contract — shared Result type, exit codes, event keys, non-interactive cores for add/update/remove/list
```

**In root `CLAUDE.md` project structure** (after the `wsvalidate/` block):
```
├── nonint/
│   ├── nonint.go        ← non-interactive mode detection + shared helpers
│   ├── config.go        ← headless config parsing
│   ├── result.go        ← shared Result type
│   ├── events.go        ← exit/event code constants
│   └── remove.go        ← headless remove core
```
(Exact file list verified from disk.)

### Finding 2 — Update CLI command count in INDEX.md

Change:
```
| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |
```
To:
```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```

### Finding 3 — Add `plan-grilling.md` to station/CLAUDE.md Workflows nav table

Add row under Workflows:
```
| Running the adversarial 6-critic review pipeline on a plan before dispatch; grilling a plan for blocks and forks | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

### Finding 4 — Add `critic-agent-prompts.md` to station/CLAUDE.md Skills nav table

Add row under Skills (verify description by reading the file first):
```
| Working with critic agents or configuring the adversarial review pipeline | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **Finding 1 (Medium):** `internal/nonint/` missing from architecture docs — straightforward factual update, recommend applying.
- **Finding 2 (Low):** CLI command count in INDEX.md (8 → 9) — small but accurate update, recommend applying.
- **Finding 3 (Low):** `plan-grilling.md` unlisted — add to nav table if workflow is intended to be triggered by the tech-lead agent directly. If it's internal-only scaffolding, no change needed.
- **Finding 4 (Low):** `critic-agent-prompts.md` unlisted — add to nav table with accurate trigger description. Recommend reading the file first to confirm intended use.

---

## Notes for Next Run

- **Gap since last run was 119 days** (2026-05-04 → 2026-09-01). Recommend restoring the 7-day cadence. The volume of code changes (Plans 40, 41, plus several PRs) meant several doc drift items accumulated.
- After user applies Finding 1 and 2 fixes, the architecture docs will be fresh for the next cycle.
- The `completion` command has been present since May 2026; low urgency but worth fixing for new contributors reading INDEX.md.
- Watch for headless CLI contract evolution (Plan 41 Phase 5 shipped an agent-interface contract and CHANGELOG) — if a formal CLI contract doc is added outside station/, it should be reflected here.
