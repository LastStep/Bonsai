---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-30
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
- **Files Read:** 14 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/.bonsai.yaml`, `/home/user/Bonsai/.bonsai/catalog.json`, `/home/user/Bonsai/go.mod`, `/home/user/Bonsai/cmd/completion.go`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`, `/home/user/Bonsai/station/agent/Skills/critic-agent-prompts.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`. Ran `git log --since="7 days ago"` and `git log --since="2026-05-04"` (since last run) to identify code changes.
- **Result:** Three maintenance routine commits today (backlog-hygiene, status-hygiene, memory-consolidation). Key feature commits since last run (2026-05-04):
  - **2026-05-07:** `completion` subcommand added (`cmd/completion.go`) — explicit shell completion for bash/zsh/fish/powershell
  - **2026-06-13:** Plan 40 shipped — freeze schemas + root-relative scaffolding, project-level validate pass, memory-routing protocol + guide Formats page; plan-grilling pipeline added to station workspace
  - **2026-06-16:** Plan 41 shipped — headless CLI contract; new `internal/nonint/` package; `--non-interactive`/`--yes`/`--from` flags on init/add/update/remove; `list --json`; agent-interface contract + CHANGELOG
- **Issues:** Found 4 documentation gaps (see Findings Summary)

### Step 2: Check INDEX.md accuracy
- **Action:** Compared INDEX.md tech stack, folder structure, project description, and key metrics against the actual codebase (`cmd/`, `internal/`, `go.mod`).
- **Result:**
  - Tech stack table: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, embed.FS)
  - Agent types: says 6 — confirmed 6 agent dirs in `catalog/agents/`
  - Catalog items: says "~50" — actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). "~50" is acceptable approximation.
  - **DRIFT FOUND:** CLI commands says "8 (init, add, remove, list, catalog, update, guide, validate)" but `completion` was added 2026-05-07 making it 9.
  - **DRIFT FOUND:** Architecture diagram under `internal/` lists `catalog`, `config`, `generate`, `validate`, `wsvalidate`, `tui` — but `internal/nonint/` (Plan 41, shipped 2026-06-16) is missing.
- **Issues:** 2 stale facts flagged

### Step 3: Check navigation links
- **Action:** Verified every link in `station/CLAUDE.md` — Core, Protocols, Workflows, Skills, Routines, Sensors, External References. Also spot-checked `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/`.
- **Result:**
  - Core links (identity.md, memory.md, self-awareness.md): all resolve
  - Protocol links (memory, scope-boundaries, security, session-start): all resolve
  - Workflow links (code-review, planning, pr-review, security-audit, session-logging, test-plan, session-wrapup, issue-to-implementation, routine-digest): all resolve
  - Skill links (planning-template, review-checklist, issue-classification, pr-creation, bubbletea, bonsai-model): all resolve
  - Routine links: all resolve
  - Sensor links: all resolve
  - Bonsai Reference `../.bonsai/catalog.json`: resolves correctly (file exists at `/home/user/Bonsai/.bonsai/catalog.json`)
  - **DRIFT FOUND:** `station/agent/Workflows/plan-grilling.md` exists on disk (added 2026-06-13) but has NO entry in CLAUDE.md Workflows navigation table. Agent cannot discover this workflow via the nav table.
  - **DRIFT FOUND:** `station/agent/Skills/critic-agent-prompts.md` exists on disk (added 2026-06-13) but has NO entry in CLAUDE.md Skills navigation table. Agent cannot discover this skill via the nav table.
- **Issues:** 2 nav gaps flagged

### Step 4: Report findings
- **Action:** Compiled findings table below. Per procedure, proposing updates but not executing — flagging for user decision.
- **Result:** 4 findings documented. All are clear factual updates with no ambiguity about the correct value. Flagged for user review.
- **Issues:** none

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-06-30, Next Due → 2026-07-07, Status → done.
- **Result:** Dashboard updated.
- **Issues:** none

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | low | INDEX.md CLI commands count says "8" but `bonsai completion` was added 2026-05-07 making it 9. Missing from list: `completion`. | `station/INDEX.md` line 33 + Architecture diagram line 63 | Flagged for user — propose: update count to 9, add `completion` to both list and architecture line |
| 2 | low | INDEX.md Architecture Overview missing `internal/nonint/` package (added Plan 41, 2026-06-16) — the headless CLI contract layer | `station/INDEX.md` Architecture section | Flagged for user — propose: add `internal/nonint/ ← headless CLI contract — Result/Event shapes, non-interactive runner` to the diagram |
| 3 | medium | `station/agent/Workflows/plan-grilling.md` exists (custom workflow, added 2026-06-13) but is absent from CLAUDE.md Workflows navigation table. Agent will not discover or use it via the nav table — dead skill. | `station/CLAUDE.md` Workflows table | Flagged for user — propose: add row `\| Adversarially reviewing a drafted plan before dispatch — "grill the plan", "critic pass" \| [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) \|` |
| 4 | medium | `station/agent/Skills/critic-agent-prompts.md` exists (custom skill, added 2026-06-13) but is absent from CLAUDE.md Skills navigation table. Companion to plan-grilling — also undiscoverable. | `station/CLAUDE.md` Skills table | Flagged for user — propose: add row `\| Running plan-grilling critic agents — dispatching adversarial critic suite \| [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) \|` |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

All 4 findings above are flagged for user decision. They are clear factual drift items with unambiguous proposed fixes, but CLAUDE.md is a Bonsai-generated file (has `<!-- BONSAI_START -->` / `<!-- BONSAI_END -->` markers) and may need careful handling to avoid conflicts on next `bonsai update`. Proposed fixes:

**Finding 1 — INDEX.md CLI count (trivial, no Bonsai marker):**
- Line 33: `| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |` → `| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |`
- Line 63 architecture diagram: append `, completion` to the cmd/ line

**Finding 2 — INDEX.md architecture nonint (trivial, no Bonsai marker):**
- Add `internal/nonint/ ← headless CLI contract — non-interactive runner, Result/Event shapes` after the `internal/wsvalidate/` line

**Finding 3 & 4 — CLAUDE.md nav tables (requires care — Bonsai-generated file):**
- These custom workflow/skill files have `source: adapted from ZenGarden` markers and are not in `.bonsai.yaml` custom_items — they were added as raw files, not via `bonsai add`.
- Option A: Add them to `.bonsai.yaml` `custom_items` under `workflows:` and `skills:` respectively, then run `bonsai update` to regenerate CLAUDE.md with the nav entries.
- Option B: Manually add nav rows to CLAUDE.md outside the BONSAI markers (not recommended — will be overwritten on next `bonsai update`).
- Recommendation: Option A — register them as custom_items in `.bonsai.yaml` and run `bonsai update`.

---

## Notes for Next Run

- After user resolves findings 3 & 4 via `bonsai update` or manual wiring, verify nav table entries resolve.
- `code-index.md` also has no section for `internal/nonint/` — consider updating alongside INDEX.md (finding 2) for consistency.
- `completion` subcommand is also absent from `code-index.md` CLI Commands table — low priority but worth a note.
- All other navigation links are clean — no broken refs found.
