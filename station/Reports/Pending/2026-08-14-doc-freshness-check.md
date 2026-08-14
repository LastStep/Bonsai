---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-14
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
- **Files Read:** 9 — `/home/user/Bonsai/CLAUDE.md`, `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/CLAUDE.md`, `station/Playbook/Status.md`, `station/code-index.md`, `station/agent/Workflows/plan-grilling.md` (head), `station/agent/Skills/critic-agent-prompts.md` (head), `internal/generate/list_snapshot.go` (head), `internal/nonint/nonint.go` (head), `internal/nonint/runner.go` (head)
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Files Created:** 1 — `station/Reports/Pending/2026-08-14-doc-freshness-check.md` (this report)
- **Tools Used:** Read, Bash (git log, ls, grep, head), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Ran `git log --oneline --since="2026-05-04"` to identify all commits since the last run (102 days ago). Reviewed git history for code changes that affect documentation.
- **Result:** 52+ commits since 2026-05-04. Major code-impacting events: Plan 41 (headless CLI contract, 5 PRs — `internal/nonint/` package added, `list_snapshot.go` added); Plan 40 (Odysseus integration, `catalog_snapshot_unix/windows.go` added); PR #78 (external contribution — `cmd/completion.go`); v0.4.2/v0.4.3 releases; Plan 39 (non-interactive flags).
- **Issues:** Documentation in root `CLAUDE.md` and `station/code-index.md` has not been updated to reflect these structural changes.

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md`. Compared Tech Stack, Key Metrics, Architecture Overview against current codebase.
- **Result:** Tech stack table is accurate. Agent count (6) is accurate. Catalog items (~50) is approximate and acceptable. CLI commands count says 8 (init/add/remove/list/catalog/update/guide/validate) but `bonsai completion` was added in PR #78 (2026-05-07) — count is now 9. Arch diagram in INDEX.md omits `internal/nonint/` from the `internal/` listing and both the `cmd/` line and metrics row omit `completion`. Overall: 2 low-severity items.
- **Issues:** CLI commands count 8 → 9; `internal/nonint/` and `completion` absent from arch diagram; both low severity.

### Step 3: Check navigation links
- **Action:** Checked `station/CLAUDE.md` Skills, Workflows, and Protocols nav tables against actual files in `station/agent/Skills/`, `station/agent/Workflows/`, `station/agent/Protocols/`, `station/agent/Core/`. Ran `ls` on each directory.
- **Result:**
  - **Protocols** — all 4 linked files exist (memory.md, scope-boundaries.md, security.md, session-start.md). Clean.
  - **Core** — all 3 linked files exist (identity.md, memory.md, self-awareness.md). Clean.
  - **Skills** — 7 files in Skills/; `bonsai-model.md` (previously flagged broken in 2026-05-04 run) now exists. However, `critic-agent-prompts.md` exists in Skills/ but has NO nav table row in station/CLAUDE.md.
  - **Workflows** — 8 files in Workflows/; `plan-grilling.md` exists in Workflows/ but has NO nav table row in station/CLAUDE.md. Added 2026-06-13 during Plan 40 session. This is a meaningful gap — the plan-grilling pipeline is used actively and the tech lead agent has no nav pointer to it.
  - **Sensors** — all 10 sensor files exist and match the Sensors table. Clean.
  - **Routines** — all 7 routine files match the Routines table. Clean.
- **Issues:** 2 missing nav rows in station/CLAUDE.md (plan-grilling workflow, critic-agent-prompts skill).

### Step 4: Report findings
- **Action:** Compiled 8 findings across root `CLAUDE.md`, `station/CLAUDE.md`, `station/INDEX.md`, and `station/code-index.md`.
- **Result:** All findings are documentation drift — no code quality issues found. All findings are proposed for user decision (no edits executed per routine instructions).
- **Issues:** None at execution level.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package entirely absent from root CLAUDE.md project-structure tree. Plan 41 added this as the headless CLI contract package (12+ source files: config, events, nonint, remove, result, runner, update + tests). It owns `RunInit`, `RunAdd`, `RunUpdate`, `RunRemove` headless cores and all exit-code constants. | `CLAUDE.md` (project-structure) | Flagged — propose adding `internal/nonint/` block to the tree |
| 2 | MEDIUM | `cmd/completion.go` absent from `cmd/` listing in root CLAUDE.md project-structure. Added PR #78 (2026-05-07, external contribution). | `CLAUDE.md` (project-structure) | Flagged — propose adding `completion.go` line to cmd/ block |
| 3 | MEDIUM | 3 new `internal/generate/` files absent from root CLAUDE.md: `catalog_snapshot_unix.go` and `catalog_snapshot_windows.go` (v0.4.0 hotfix platform split — POSIX-only `syscall.O_NOFOLLOW`) and `list_snapshot.go` (Plan 41 Phase 4 `list --json` contract, `ListSnapshot` type). | `CLAUDE.md` (project-structure) | Flagged — propose adding these 3 file lines to generate/ block |
| 4 | MEDIUM | `station/CLAUDE.md` Workflows nav table missing `plan-grilling.md` row. File exists at `agent/Workflows/plan-grilling.md` — added 2026-06-13. This is the 6-critic adversarial review pipeline used for Plan 40, 41 grilling. Active workflow, actively used, with no nav pointer. | `station/CLAUDE.md` (Workflows table) | Flagged — propose adding nav row |
| 5 | MEDIUM | `station/code-index.md` has no section for `internal/nonint/` package. The package is the canonical headless runner for all 4 mutating commands. Likely the highest-value missing code-index section. | `station/code-index.md` | Flagged — propose new section after Workspace-path Validation |
| 6 | LOW | `station/CLAUDE.md` Skills nav table missing `critic-agent-prompts.md` row. File exists at `agent/Skills/critic-agent-prompts.md` — added 2026-06-13. Contains verbatim prompt templates for the 6 plan-grilling critics. | `station/CLAUDE.md` (Skills table) | Flagged — propose adding nav row |
| 7 | LOW | `station/code-index.md` Generator section missing `list_snapshot.go` entry. Added Plan 41 Phase 4; exports `ListSnapshot` type and write function for `bonsai list --json` contract. | `station/code-index.md` (Generator section) | Flagged — propose adding row to catalog_snapshot table |
| 8 | LOW | `station/code-index.md` Generator section missing `catalog_snapshot_unix.go` / `catalog_snapshot_windows.go`. Platform-split files from v0.4.0 hotfix. Not critical to document individually, but the existing `catalog_snapshot.go` entry implies a single file when 3 exist. | `station/code-index.md` (Generator section) | Flagged — propose a note in existing catalog_snapshot row |
| 9 | LOW | `station/INDEX.md` CLI commands metric says 8; `bonsai completion` was added in PR #78 making it 9. Also absent from the arch diagram `cmd/` annotation in INDEX.md. | `station/INDEX.md` (Key Metrics, arch diagram) | Flagged — propose updating count to 9 and adding `completion` to annotation |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

All 9 findings are proposed updates — no edits were made to documentation files per the routine's Step 4 instruction. The following need user decision:

**HIGH priority (do first):**
- **Finding #1** — Root `CLAUDE.md` project-structure: add `internal/nonint/` block (6–8 lines describing the package and its key files: config.go, events.go, nonint.go, result.go, runner.go + platform entries for update.go and remove.go).

**MEDIUM priority (batch in one sweep):**
- **Finding #2** — Root `CLAUDE.md` project-structure: add `completion.go ← bonsai completion (bash/zsh/fish/powershell)` to cmd/ block.
- **Finding #3** — Root `CLAUDE.md` project-structure: add the 3 new generate/ files after `catalog_snapshot_test.go`.
- **Finding #4** — `station/CLAUDE.md` Workflows table: add `plan-grilling.md` row with activate-when trigger ("Grilling a drafted plan against 6 adversarial critics before dispatch").
- **Finding #5** — `station/code-index.md`: add new `## Headless CLI Contract (internal/nonint/)` section documenting RunInit, RunAdd, RunUpdate, RunRemove, exit-code constants, and the JSONL emission pattern.

**LOW priority (opportunistic):**
- **Findings #6, #7, #8** — nav rows + code-index entry additions; low-risk, minor value.
- **Finding #9** — INDEX.md metrics count 8 → 9.

**Recommendation:** Scope Findings #1–5 into a single P2 doc-refresh pass (similar to Plan 37). The root CLAUDE.md drift is the highest-impact item — it's the first file any new agent or developer reads and it currently shows a materially incomplete picture of `internal/`.

## Notes for Next Run

- The root `CLAUDE.md` project-structure tree has been persistently stale across multiple doc-freshness cycles. Consider adding a standing sub-step: "verify root CLAUDE.md `internal/` and `cmd/` blocks match `ls internal/` and `ls cmd/*.go`" to catch this automatically.
- `station/CLAUDE.md` Workflows table had 2 new files added in June without a corresponding nav row update. This suggests the workflow for "add a custom workflow to station" does not include a reminder to update the nav table — potential process gap.
- `plan-grilling.md` and `critic-agent-prompts.md` both have `source: adapted from ZenGarden` and note "full Bonsai-catalog integration pending (Backlog)." They are custom files, not catalog-managed. This is noted for context.
- `bonsai-model.md` (flagged broken in 2026-05-04 run) now resolves correctly — resolved.
- Previous cycle's 5 flags: (1) root CLAUDE.md project-structure stale — STILL OUTSTANDING (now deeper); (2) broken nav link bonsai-model.md — RESOLVED; (3) code-index stale — STILL OUTSTANDING (now wider); (4) INDEX.md CLI count 7→8 — was fixed in Routine Digest 2026-05-04 but now needs 8→9; (5) INDEX.md arch diagram drift — partially addressed but nonint still missing.
