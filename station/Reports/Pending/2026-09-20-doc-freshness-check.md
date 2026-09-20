---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-09-20
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
- **Duration:** ~8 min
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Status.md`, `station/agent/Core/memory.md`, `station/code-index.md`, `station/CLAUDE.md` (via system-reminder), plus directory listings for `cmd/`, `internal/`, `internal/tui/`, `internal/nonint/`, `internal/generate/`, `station/agent/Skills/`, `station/agent/Workflows/`, `station/agent/Sensors/`, `station/agent/Core/`, `station/agent/Protocols/`, `station/Playbook/Plans/Active/`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Ran `git log --since="7 days ago"` and cross-referenced with INDEX.md, Status.md, and RoutineLog.md to identify features not reflected in docs.
- **Result:** Only 2 commits in the last 7 days (both station-only routine runs: backlog-hygiene and status-hygiene — no code changes). However, this run is catching up from the prior run on 2026-05-04. Since then, significant code landed: Plan 39 (v0.4.2 / `internal/nonint/` package + `--non-interactive` flags), Plan 40 (v0.5.0 Odysseus integration), Plan 41 (headless CLI contract), v0.4.3 hotfix (catalog_snapshot platform split), and first external contribution (`bonsai completion`). Multiple docs have not been updated to reflect these changes.
- **Issues:** 9 drift items found — see Findings Summary.

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` and compared Tech Stack, Key Metrics, and Architecture Overview against the actual file system.
- **Result:** 
  - Tech Stack rows: all accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, single binary. ✓
  - Key Metrics: "CLI commands | 8" is **stale** — `bonsai completion` was added 2026-05-07 (PR #78, first external contribution); count should be 9.
  - Architecture Overview: `cmd/` description lists "init, add, remove, list, catalog, update, guide, validate" — missing `completion`. `internal/nonint/` layer (added Plan 39) not shown in the diagram at all.
- **Issues:** 2 stale entries in INDEX.md (metrics count + arch diagram).

### Step 3: Check navigation links
- **Action:** Cross-referenced all links in `station/CLAUDE.md` nav tables (Core, Protocols, Workflows, Skills, Sensors, Routines, External References, Bonsai Reference) against actual files on disk. Also checked agent/Core/, agent/Protocols/, agent/Workflows/, agent/Skills/ directories.
- **Result:**
  - Core links (identity.md, memory.md, self-awareness.md, routines.md): all resolve ✓
  - Protocol links (memory.md, scope-boundaries.md, security.md, session-start.md): all resolve ✓
  - Workflow links (code-review.md, planning.md, pr-review.md, security-audit.md, session-logging.md, test-plan.md, session-wrapup.md, issue-to-implementation.md, routine-digest.md): all resolve ✓
  - Skills links (bonsai-model.md, bubbletea.md, issue-classification.md, planning-template.md, pr-creation.md, review-checklist.md): all resolve ✓
  - Sensor links (all 10 listed): all resolve ✓
  - Routine links (all 7): all resolve ✓
  - Bonsai Reference: `.bonsai/catalog.json` exists ✓; `agent/Skills/bonsai-model.md` exists ✓
  - External References: Playbook/Status.md ✓, Playbook/Roadmap.md ✓, Playbook/Standards/SecurityStandards.md ✓, Playbook/Plans/Active/ ✓, Playbook/Backlog.md ✓, Logs/KeyDecisionLog.md ✓, Reports/Pending/ ✓, Reports/report-template.md ✓, code-index.md ✓, INDEX.md ✓
  - **Unlisted files** (exist in directory but absent from nav): `agent/Skills/critic-agent-prompts.md` and `agent/Workflows/plan-grilling.md`.
- **Issues:** No broken links found. Two unlisted files noted (info-level — may be intentionally private or under construction).

### Step 4: Report findings
- **Action:** Compiled all findings from Steps 1–3.
- **Result:** 9 findings total — 2 medium, 3 low, 4 info. No changes executed (audit-only routine per procedure). Flagging for user decision.
- **Issues:** none.

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for "Doc Freshness Check" (Last Ran → 2026-09-20, Next Due → 2026-09-27, Status → done).
- **Result:** Dashboard updated ✓
- **Issues:** none.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | INDEX.md Key Metrics: "CLI commands \| 8" — should be 9 (`bonsai completion` added 2026-05-07 via PR #78) | `station/INDEX.md` line ~33 | Flagged — propose update: `8 → 9 (init, add, remove, list, catalog, update, guide, validate, completion)` |
| 2 | Medium | Root `Bonsai/CLAUDE.md` project structure: `cmd/completion.go` missing from tree (added 2026-05-07) | `CLAUDE.md` cmd/ block | Flagged — propose adding entry after validate.go |
| 3 | Medium | Root `Bonsai/CLAUDE.md` project structure: `internal/nonint/` package entirely absent — added in Plan 39/v0.4.2, contains `nonint.go`, `runner.go`, `config.go`, `events.go`, `result.go`, `remove.go`, `update.go`, and tests | `CLAUDE.md` internal/ block | Flagged — propose adding `internal/nonint/` section between `wsvalidate/` and `tui/` |
| 4 | Low | Root `Bonsai/CLAUDE.md` project structure: `internal/generate/` missing platform-split files added in v0.4.0 hotfix + Plan 41: `catalog_snapshot_unix.go`, `catalog_snapshot_unix_test.go`, `catalog_snapshot_windows.go`, `list_snapshot.go` | `CLAUDE.md` internal/generate/ block | Flagged — propose updating generate/ listing |
| 5 | Low | INDEX.md Architecture Overview: `cmd/` list and `internal/` diagram do not show `completion` command or `nonint` layer | `station/INDEX.md` Architecture Overview block | Flagged — propose adding `completion` to cmd list and `internal/nonint/` row to arch table |
| 6 | Low | `station/code-index.md`: no entry for `bonsai completion` command; no section for `internal/nonint/` package (both post-date last code-index refresh in Plan 37 / 2026-05-07) | `station/code-index.md` | Flagged — propose adding completion row to CLI Commands table and a new `internal/nonint/` section |
| 7 | Info | `station/agent/Skills/critic-agent-prompts.md` exists in Skills directory but is not listed in station/CLAUDE.md Skills nav table — may be intentionally unlisted or under development | `station/CLAUDE.md` Skills table | Flagged — user to decide: add to nav table or leave as private |
| 8 | Info | `station/agent/Workflows/plan-grilling.md` exists in Workflows directory but is not listed in station/CLAUDE.md Workflows nav table | `station/CLAUDE.md` Workflows table | Flagged — user to decide: add to nav table or leave as private |
| 9 | Info | `Plans/Active/` contains Plan 40 (`40-odysseus-platform-integration.md`) and Plan 41 (`41-headless-cli-contract.md`) — both fully shipped; memory.md and previous status-hygiene both flag Plan 41 for archiving; Plan 40 phases 1–3 shipped (tag held, phase 4 held) | `station/Playbook/Plans/Active/` | Flagged — already flagged by status-hygiene routine; user to decide archive timing |

---

## Errors & Warnings
No errors encountered.

---

## Items Flagged for User Review

1. **INDEX.md CLI count** — easy one-liner: `8 → 9`. Recommend applying before next release note or doc sweep.
2. **Root CLAUDE.md project structure drift (medium, recurring)** — `cmd/completion.go` and `internal/nonint/` are both absent. This is the recurring category from prior doc-freshness runs (2026-04-14, 2026-04-21, 2026-05-04 all found root-CLAUDE.md drift). The previous routine flagged a P2/P3 backlog item to add a "root-CLAUDE.md check" sub-step to this routine. Recommend bundling these into the next doc-refresh plan (Plan 37 pattern) or fixing in a quick chore commit.
3. **`critic-agent-prompts.md` and `plan-grilling.md`** — unlisted in nav tables. If intentionally private, no action. If they should be discoverable, add rows.

---

## Notes for Next Run
- Root CLAUDE.md project structure drift is a **recurring finding** across the last 4 doc-freshness runs. Consider whether a pre-commit hook or a CI step checking that key source files appear in the project-structure comment would prevent future drift.
- The `internal/nonint/` gap is significant: it's a first-class package introduced in a named plan/release but entirely absent from the only developer-facing project structure reference.
- code-index.md was last refreshed in Plan 37 (2026-05-07). Consider including it explicitly in the next doc-refresh plan.
- No broken nav links were found this cycle — nav health is clean.
