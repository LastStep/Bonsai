---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-05
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
- **Files Read:** 10 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/code-index.md`, `station/CLAUDE.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Backlog.md`, `CLAUDE.md` (root)
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, grep), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history

Checked git log for the last 7 days: **no commits** in the period 2026-06-28 to 2026-07-05.

Extended the window to the period since the last routine run (2026-05-04). Significant changes found:

**Plan 40 (2026-06-13):** Frozen v1 schemas + root-relative scaffolding, project-level validate pass, guide Formats page added. Shipped 3 phases merged (PRs #114, #115, #116).

**Plan 41 (2026-06-16):** Headless CLI contract — 5 phases merged (PRs #120-#125). Key deliverables:
- New `internal/nonint/` package: pure headless cores for init, add, update, remove-agent, remove-item; exit code constants; JSONL emission functions.
- New flags: `--yes/-y`, `--from`, `--non-interactive`, `--delete-files` on `bonsai remove`; `--non-interactive`, `--skip-conflicts` on `bonsai update`.
- `list --json` parity.
- `docs/agent-interface.md` — canonical headless CLI contract doc.
- `showWriteResults()` / `splitTopSegment()` removed from `cmd/root.go`.
- `ExitConflict=5` exit code added.

**`bonsai completion` (2026-05-07):** Added by external contributor @mvanhorn — shell completion for bash/zsh/fish/powershell (PR #78, commit `2eae9d4`).

### Step 2 — Check INDEX.md accuracy

Read `station/INDEX.md` in full.

**Tech stack:** Still accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, embed.FS). No drift.

**Key Metrics — drift found:**
- "CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate)" — The `completion` command was added in May 2026. Count should be **9** and the list should include `completion`. **(Finding #1)**

**Architecture Overview — drift found:**
- `internal/nonint/` is entirely absent from the package list. This package was added in Plan 41 (June 2026) and contains the canonical headless CLI cores. **(Finding #2)**

### Step 3 — Check navigation links

Verified all links in `station/CLAUDE.md` navigation tables by checking directory contents:

- `agent/Core/` — identity.md, memory.md, self-awareness.md, routines.md — all **PASS**
- `agent/Protocols/` — memory.md, scope-boundaries.md, security.md, session-start.md — all **PASS**
- `agent/Workflows/` — code-review.md, issue-to-implementation.md, plan-grilling.md, planning.md, pr-review.md, routine-digest.md, security-audit.md, session-logging.md, session-wrapup.md, test-plan.md — all **PASS**
- `agent/Skills/` — bonsai-model.md, bubbletea.md, critic-agent-prompts.md, issue-classification.md, planning-template.md, pr-creation.md, review-checklist.md — all **PASS**
- `agent/Routines/` — backlog-hygiene.md, dependency-audit.md, doc-freshness-check.md, memory-consolidation.md, roadmap-accuracy.md, status-hygiene.md, vulnerability-scan.md — all **PASS**
- `agent/Sensors/` — agent-review.sh, compact-recovery.sh, context-guard.sh, dispatch-guard.sh, routine-check.sh, scope-guard-files.sh, session-context.sh, status-bar.sh, statusline.sh, subagent-stop-review.sh — all **PASS**
- `../.bonsai/catalog.json` — exists **PASS**
- `../.bonsai.yaml` — exists **PASS**

**All navigation links resolve. No broken links.**

### Step 4 — Check code-index.md for drift

Read `station/code-index.md` in full and compared against actual codebase.

**drift found:**

- **Missing `completion` command** — `cmd/completion.go` exists but has no row in the CLI Commands table. **(Finding #3)**
- **Stale `showWriteResults()` entry** — Listed in Shared Helpers at line `:201` but was removed from `cmd/root.go` in Plan 41 Phase 2. Confirmed absent from root.go function list. **(Finding #4)**
- **Missing `internal/nonint/` section** — Entire new package from Plan 41 with 5+ public functions and 3 sub-files is not documented in code-index.md. **(Finding #5)**

Also checked CLAUDE.md (root `/home/user/Bonsai/CLAUDE.md`):
- `cmd/completion.go` missing from project structure's `cmd/` listing. **(Finding #6)**
- `internal/nonint/` missing from project structure's `internal/` listing. **(Finding #7)**

### Step 5 — Dashboard update

Updated `agent/Core/routines.md` dashboard row for "Doc Freshness Check" to Last Ran 2026-07-05, Next Due 2026-07-12, Status done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | CLI commands count says 8; `completion` is the 9th command | `station/INDEX.md` Key Metrics table, line 33 | Flagged — awaiting user decision |
| 2 | Medium | `internal/nonint/` package absent from Architecture Overview | `station/INDEX.md` Architecture section | Flagged — awaiting user decision |
| 3 | Medium | `completion` command missing from CLI Commands table | `station/code-index.md` CLI Commands section | Flagged — awaiting user decision |
| 4 | Medium | `showWriteResults()` listed in Shared Helpers but was removed in Plan 41 | `station/code-index.md` line ~39 | Flagged — awaiting user decision |
| 5 | High | `internal/nonint/` package has no section in code-index (RunInit, RunAdd, RunUpdate, RunRemoveAgent, RunRemoveItem, EmitJSONL, Result, etc.) | `station/code-index.md` | Flagged — awaiting user decision |
| 6 | Low | `completion.go` missing from cmd/ in project structure | `/home/user/Bonsai/CLAUDE.md` cmd/ file listing | Flagged — awaiting user decision |
| 7 | Low | `internal/nonint/` missing from internal/ in project structure | `/home/user/Bonsai/CLAUDE.md` internal/ listing | Flagged — awaiting user decision |

## Proposed Updates (not executed — flag for user)

**station/INDEX.md:**
- Line 33: Change "8 (init, add, remove, list, catalog, update, guide, validate)" → "9 (init, add, remove, list, catalog, update, guide, validate, completion)"
- Architecture Overview: Add `internal/nonint/` entry — "headless CLI cores (RunInit/RunAdd/RunUpdate/RunRemoveAgent/RunRemoveItem) + JSONL emission + exit code constants"

**station/code-index.md:**
- Add `completion` row to CLI Commands table (file: `cmd/completion.go`, entry: `completionCmd`, purpose: "Generate shell completion script for bash/zsh/fish/powershell")
- Remove stale `showWriteResults()` row from Shared Helpers table
- Add new `internal/nonint/` section documenting: runner.go (RunInit, RunAdd), update.go (RunUpdate), remove.go (RunRemoveAgent, RunRemoveItem), events.go (EmitJSONL, EmitFile, EmitSummary), result.go (Result, Counts()), nonint.go (config), runner.go (exit code constants: ExitConflict=5)

**CLAUDE.md (root):**
- cmd/ listing: add `completion.go` entry
- internal/ listing: add `nonint/` directory with description

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

7 documentation drift items identified — all require human confirmation before update since they are factual corrections to reference docs. Recommendations:

1. **Prioritize Finding #5** — the missing `internal/nonint/` section in code-index.md is the highest-severity item. Plan 41 added a substantial new package that any developer agent consulting the code-index would miss entirely.
2. **Findings #1-4** are medium-severity factual corrections; straightforward to apply.
3. **Findings #6-7** are low-severity root CLAUDE.md structure updates.

## Notes for Next Run

- The gap since last run was 62 days (2026-05-04 → 2026-07-05). Two major plans shipped in that window (Plans 40 and 41). Consider scheduling the routine closer to its 7-day frequency.
- `docs/agent-interface.md` now exists as the canonical headless CLI contract doc — a reference to this could be added to `station/INDEX.md` Document Registry if the user wants the tech-lead agent to be aware of it.
- Navigation links are clean — no follow-up needed there.
