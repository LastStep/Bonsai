---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-04-21
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-04-14
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~5 min
- **Files Read:** 13
  - `station/agent/Routines/doc-freshness-check.md`
  - `station/INDEX.md`
  - `station/CLAUDE.md`
  - `Bonsai/CLAUDE.md` (project root)
  - `station/Playbook/Status.md`
  - `station/Playbook/Roadmap.md`
  - `station/Playbook/Backlog.md`
  - `station/code-index.md`
  - `station/agent/Core/routines.md`
  - `station/agent/Core/memory.md`
  - `station/Logs/RoutineLog.md`
  - Directory listings for `agent/`, `catalog/`, `internal/tui/`, `cmd/`
- **Files Modified:** 0 — audit-only; findings flagged for user review, no autonomous edits
- **Tools Used:** Read, Bash (ls + git log), Grep
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read `station/INDEX.md`, `station/CLAUDE.md`, `Bonsai/CLAUDE.md`, `station/Playbook/{Status,Roadmap,Backlog}.md`, `station/code-index.md`, `station/agent/Core/{routines,memory}.md`. Ran `git log --since='7 days ago'` to get the commit window (54 commits, 2026-04-14 through 2026-04-21).
- **Result:** Recent work in the window: Plan 15 (BubbleTea harness PR #26 `2ce63f6`), Plan 18 (bonsai guide multi-topic PR #25), Plan 16 (go install binary name PR #23), Plan 17 (release prep PR #24), Plan 10 (docs site Phases A-D), Plan 11/12/14 (UI/UX overhaul phases), Plan 13 (ActionUnchanged follow-ups). Most are reflected in Status.md "Recently Done"; drift found in three Tier-1 documentation artifacts (detailed below).
- **Issues:** 3 drift items found (see Findings Summary).

### Step 2: Check INDEX.md accuracy
- **Action:** Compared `station/INDEX.md` Tech Stack, Key Metrics, Architecture Overview, and Document Registry against the actual codebase (`catalog/agents/` count, `catalog/skills/` count, `cmd/` file list, `internal/` layout).
- **Result:** Tech stack accurate. Key Metrics accurate (6 agent types verified via `ls catalog/agents/`; 7 CLI commands verified in `cmd/`; ~50 catalog items is in the right ballpark). BubbleTea is listed as a stack component. However, the **Architecture Overview ASCII diagram** still describes `internal/tui/` as "Huh forms + LipGloss styled output" — it does not mention the BubbleTea harness layer added by Plan 15 (`internal/tui/harness/`). Minor drift, not blocking.
- **Issues:** 1 drift item (see finding #1).

### Step 3: Check navigation links in station/CLAUDE.md
- **Action:** Programmatically verified every file link in `station/CLAUDE.md`'s Core / Protocols / Workflows / Skills / Routines / Sensors / External References tables (43 link targets).
- **Result:** 43/43 resolve. No broken links.
- **Issues:** none.

### Step 4: Check navigation link coverage vs actual files
- **Action:** Cross-referenced `station/CLAUDE.md` Skills table against `ls agent/Skills/`. Same for Routines and Sensors.
- **Result:**
  - **Skills table lists 4 entries** (planning-template, review-checklist, issue-classification, pr-creation).
  - **`agent/Skills/` contains 5 non-bak entries:** those 4 plus `bubbletea.md` and a `bubbletea/` subdirectory (components.md, emoji-width-fix.md, golden-rules.md, troubleshooting.md).
  - `bubbletea` appears to be a custom/local skill added outside the catalog — it is not present in `catalog/skills/bubbletea`. It is not registered in the CLAUDE.md nav table. Whether this is intentional (custom skills are private) or a nav-table omission is a user decision.
- **Issues:** 1 item for user decision (see finding #2).

### Step 5: Check Bonsai/CLAUDE.md project structure tree
- **Action:** Compared the `Project Structure` ASCII tree in the root `Bonsai/CLAUDE.md` against the actual repository layout.
- **Result:** `cmd/bonsai/main.go` and `embed.go` both correctly represented (fixed in Plan 16 / PR #23 and Plan 18 / PR #25 respectively). `internal/generate/` listing includes test files accurately. However, the `internal/tui/` block only lists `styles.go` and `prompts.go` — **missing `harness/` subdirectory (Plan 15, 7 files) and `styles_test.go`**. This is the most significant drift found this run.
- **Issues:** 1 drift item (see finding #3).

### Step 6: Check cosmetic table formatting in CLAUDE.md + routines.md
- **Action:** Re-read the Routines table in `station/CLAUDE.md` and the dashboard in `agent/Core/routines.md`.
- **Result:** Both tables have a blank line after the first 3 routines (Backlog Hygiene, Dependency Audit, Doc Freshness Check), which splits the table into two fragments. In Obsidian/GitHub-rendered markdown this breaks the table, producing two separate tables back-to-back with only the first showing headers.
- **Issues:** 1 cosmetic drift item (see finding #4).

### Step 7: Check Roadmap.md + Status.md accuracy
- **Action:** Compared Roadmap Phase 1 checkboxes and Status.md "Pending" entries against the Plan 14/17/18 completion history.
- **Result:** Roadmap Phase 1 still has unchecked boxes for "Better trigger sections" (Phases A+B shipped, Phase C paused — partial), "UI overhaul" (Plans 11/12/14 all shipped — should be checked), and "Usage instructions" (Plans 02/05/18 shipped — likely checkable). Status.md Pending "Better trigger sections — Phase C" still reads "paused while UI/UX Phase 3 ships" — UI/UX Phase 3 shipped 2026-04-17 via PR #24, so the blocker wording is stale. **Note:** These same items were already flagged by the 2026-04-21 Backlog Hygiene routine (see `Reports/Pending/2026-04-21-backlog-hygiene.md` flags 1+2). Not re-flagging, just cross-referencing.
- **Issues:** already captured by Backlog Hygiene routine; noted for coherence.

### Step 8: Log results + update dashboard
- **Action:** Updating `agent/Core/routines.md` dashboard (Last Ran 2026-04-21, Next Due 2026-04-28, Status done), appending to `Logs/RoutineLog.md`, and writing this report.
- **Result:** Completed.
- **Issues:** none.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | low | `INDEX.md` Architecture Overview describes `internal/tui/` as "Huh forms + LipGloss styled output" with no mention of the BubbleTea harness layer (Plan 15, merged 2026-04-20). Tech Stack row already lists BubbleTea, so this is a minor omission in the diagram only. | `station/INDEX.md:58-75` | flagged for user |
| 2 | info | `station/CLAUDE.md` Skills navigation table has 4 entries but `agent/Skills/` contains a 5th skill (`bubbletea`) with a deep subdirectory (4 topic files). Likely an intentional custom skill (not in catalog), but worth confirming whether it should be surfaced in the nav table. | `station/CLAUDE.md:63-70`, `station/agent/Skills/bubbletea*` | flagged for user decision |
| 3 | medium | Root `Bonsai/CLAUDE.md` project-structure tree's `internal/tui/` block lists only `styles.go` and `prompts.go` — missing the entire `harness/` subdirectory added by Plan 15 (contains the BubbleTea step/reducer/splicer machinery that now drives init/add/remove/update) and `styles_test.go`. Also the overall `internal/tui/` one-line description predates the harness migration. This is the most visible drift: the root CLAUDE.md is the codebase orientation doc. | `Bonsai/CLAUDE.md:48-50` | flagged for user (outside station/ scope so I cannot edit) |
| 4 | low | Routines table in both `station/CLAUDE.md` (lines 75-83) and `station/agent/Core/routines.md` dashboard (lines 33-42) has a blank row splitting the routines list into two fragments, breaking markdown table rendering in GitHub/Obsidian. The second fragment (Memory Consolidation → Vulnerability Scan) shows without headers. | `station/CLAUDE.md:75-83`, `station/agent/Core/routines.md:33-42` | flagged for user |
| 5 | info | Many `*.bak` files from the 2026-04-15 `bonsai update` marker-migration still linger across `agent/Core/`, `agent/Protocols/`, `agent/Sensors/`, `agent/Skills/`, `agent/Routines/` (10+ files). Not doc drift per se, but they pollute directory listings and may confuse future doc freshness / status hygiene runs. | `station/agent/{Core,Skills,Sensors,Routines}/*.bak` | flagged for user |

## Errors & Warnings

No errors encountered.

Cross-reference: Findings #1 and #2 from today's Backlog Hygiene report (`Reports/Pending/2026-04-21-backlog-hygiene.md`) overlap with what I'd have otherwise flagged as a #6 here (Roadmap Phase 1 unchecked boxes + stale Status.md Pending blocker for Plan 08 Phase C). Deferring to that report to avoid duplicate items in the user's queue.

## Items Flagged for User Review

1. **[low] `station/INDEX.md` architecture diagram drift** — add a line noting BubbleTea harness + `internal/tui/harness/` alongside the existing "Huh forms + LipGloss styled output" line, or rewrite the tui one-liner to cover all three Charm libs.

2. **[info] `bubbletea` skill navigation visibility** — decide whether to add a row for `agent/Skills/bubbletea.md` in the Skills table of `station/CLAUDE.md`, or leave it as a silent custom skill. If nav-table entry is desired, also decide whether to link the subdirectory topics.

3. **[medium] Root `Bonsai/CLAUDE.md` project-structure tree is stale** — `internal/tui/` listing does not reflect the Plan 15 harness migration (7 files under `harness/`) nor `styles_test.go`. Needs an edit outside my `station/` scope — either user edits directly or dispatches to a code agent. Suggested addition:
   ```
   │   └── tui/
   │       ├── styles.go         ← LipGloss styles, panels, trees, display helpers
   │       ├── styles_test.go    ← tests for styling helpers
   │       ├── prompts.go        ← Huh form wrappers (text, select, multi-select, confirm)
   │       └── harness/          ← BubbleTea reducer/step/splicer machinery driving init/add/remove/update
   ```

4. **[low] Broken routines table formatting** — remove the blank row between the first 3 and last 4 routines in both `station/CLAUDE.md:79` (the blank line) and `station/agent/Core/routines.md:38`. Double-check the underlying template at `catalog/` that produces `routines.md` so the fix sticks on regenerate. (This table is built by `RoutineDashboard()` in `internal/generate/generate.go:884` — the ordering/grouping may be an artifact of how frequencies are iterated.)

5. **[info] Stale `.bak` files across `agent/` subdirectories** — dating back to the 2026-04-15 marker migration (10+ files). Safe to delete if user confirms no rollback needed; would clean up `ls agent/Skills/`, `ls agent/Sensors/`, etc.

## Notes for Next Run

- **Routine digest in 7 days should bundle these with 2026-04-21 Backlog Hygiene findings** — roadmap checkboxes, Status.md Pending stale blocker, routines table formatting, root CLAUDE.md structure tree drift. These are all small text edits that could be batched into one session.
- **Consider adding a sub-step for the root `Bonsai/CLAUDE.md`** — Step 1 of the current procedure only says "Read docs in `station/`" but the root CLAUDE.md is the codebase orientation doc and drifts fastest when the file layout changes. A small procedure tweak would catch it automatically.
- **Watch for a recurring pattern:** plans that change the `cmd/` or `internal/` layout (Plan 15, Plan 16, Plan 18, Plan 09 code-index refresh) consistently leave root CLAUDE.md tree out of date for 1-2 weeks. Could justify a new lightweight "repo-structure tree drift check" — or just fold into this routine permanently.
