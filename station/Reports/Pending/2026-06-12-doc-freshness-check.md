---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-12
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
- **Files Read:** 12 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/40-odysseus-platform-integration.md`, `/home/user/Bonsai/station/agent/Core/memory.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/CLAUDE.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** `git log --oneline --since="14 days ago"`, `git diff HEAD~2 HEAD --name-only`, `git log --oneline --format`, `git show <hash> --stat`, directory listings via `ls`, `grep` patterns for link/content checks
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Ran `git log` for last 14 days; checked files changed in recent commits; read Status.md, INDEX.md, Roadmap.md, Plans/Active/, and memory.md.
- **Result:** Only 2 commits in the window: `b18df20` (docs(station): adopt Plan 40 — Odysseus platform integration) and `08e7f49` (chore(station): commit v0.4.3 session tracker updates). Plan 40 introduces the Odysseus integration boundary (new workstreams: `project.yaml`, `station/Memory/`, `bonsai export`, `bonsai validate` extensions, graphify sensor, plan format spec). These workstreams are planned — not yet implemented — but several docs have gaps relative to the plan adoption.
- **Issues:** Plan 40 is active in `Plans/Active/` but INDEX.md, Roadmap.md, and code-index.md do not reference it.

### Step 2: Check INDEX.md accuracy
- **Action:** Read INDEX.md in full; cross-checked tech stack, CLI command count, catalog item count, and architecture overview against actual codebase.
- **Result:**
  - **CLI commands count:** INDEX.md says "8 (init, add, remove, list, catalog, update, guide, validate)" but `completion` command was shipped in v0.4.1 (PR #78, commit `2eae9d4`) — actual count is **9**.
  - **Catalog items count:** INDEX.md says "~50" — actual count is 53 (skills: 18, workflows: 10, protocols: 4, sensors: 13, routines: 8). Within the stated approximation, not a material drift.
  - **Architecture diagram:** `internal/nonint/` package missing — shipped in v0.4.2 (commit `410a5f1`). Architecture overview in INDEX.md lists `internal/` packages but does not mention `nonint`.
  - **Plan 40 context:** INDEX.md has no mention of Odysseus integration or Plan 40 workstreams.
- **Issues:** CLI count stale (8 vs 9), `internal/nonint` missing from arch diagram.

### Step 3: Check navigation links
- **Action:** Parsed all relative links from `station/CLAUDE.md` and verified each path exists.
- **Result:** All 49 relative links in `station/CLAUDE.md` resolve to existing files/directories. Zero broken links. (The 2026-05-04 `agent/Skills/bonsai-model.md` broken-link flag is resolved — file exists.)
- **Issues:** None. Navigation links are clean.

### Step 4: Report findings
- **Action:** Compiled findings table; assessed severity; flagged items per routine spec (propose updates, don't execute).
- **Result:** 5 findings identified: 2 medium, 2 low, 1 informational.
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for Doc Freshness Check: Last Ran → 2026-06-12, Next Due → 2026-06-19, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | `completion` command missing from CLI commands list (says 8, actual 9; list omits `completion`) | `station/INDEX.md` Key Metrics table, Architecture Overview cmd comment | Flagged for user — propose updating count to 9 and adding `completion` to the list |
| 2 | medium | `internal/nonint/` package not in architecture overview or code-index — shipped v0.4.2, handles `--non-interactive --from-config` mode | `station/INDEX.md` Architecture Overview, `station/code-index.md` (no section), root `CLAUDE.md` project tree | Flagged for user — propose adding nonint to all three location |
| 3 | low | Root `CLAUDE.md` project tree still missing `completion.go` in `cmd/` section (recurring from prior cycles) | `/home/user/Bonsai/CLAUDE.md` cmd/ tree block | Flagged for user — tracked in Backlog P2 ungrouped (add root CLAUDE.md tree-drift check to routine). Propose adding `completion.go` line. |
| 4 | low | Roadmap Phase 3 "Greenhouse companion app" may be superseded by Odysseus (Plan 40). Plan 40 adopts Odysseus as the hub runtime; Greenhouse is not mentioned. Worth confirming if Greenhouse is deprecated/renamed or still a separate future item. | `station/Playbook/Roadmap.md` Phase 3 | Flagged for user decision — unclear if Greenhouse is retired |
| 5 | info | Plan 40 (Odysseus integration) is active but not reflected in INDEX.md document registry or Roadmap Future Phases section. No urgency — plan is design-phase only, no code shipped yet. | `station/INDEX.md`, `station/Playbook/Roadmap.md` | Informational — recommend adding Plan 40 context to Roadmap Phase 2/3 when workstreams begin implementation |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

1. **INDEX.md CLI count**: Update "8 (init, add, remove, list, catalog, update, guide, validate)" → "9 (init, add, remove, list, catalog, update, guide, validate, completion)". Also update the Architecture Overview cmd comment line.

2. **`internal/nonint/` documentation**: Package shipped v0.4.2 but missing from INDEX.md arch diagram, code-index.md, and root CLAUDE.md project tree. Recommend a small doc patch adding a `nonint/` row to each location with description: "non-interactive mode runtime — `--non-interactive --from-config` support for `init`/`add`".

3. **Root CLAUDE.md project tree drift**: `completion.go` still not in cmd/ section. This is the third consecutive cycle flagging root CLAUDE.md drift. The Backlog P2 item to add a sub-step to the doc-freshness-check routine procedure exists — recommend actioning the Backlog item so future cycles catch this automatically.

4. **Roadmap / Greenhouse**: Confirm whether the Greenhouse companion app (Roadmap Phase 3) is superseded by Odysseus or remains a separate long-term item. If superseded, update or remove the row.

## Notes for Next Run
- Previous cycle flags largely resolved: bonsai-model.md nav link fixed, code-index.md refreshed via Plan 37, INDEX.md Go version fixed. Only root CLAUDE.md tree drift (completion.go, nonint/) persists.
- The Backlog P2 item "Add root Bonsai/CLAUDE.md tree-drift check to doc-freshness-check routine" has been promoted from P3 and should be actioned to prevent future recurrence.
- Odysseus integration (Plan 40) will generate doc obligations when implementation begins — flag at that point for INDEX.md, Roadmap, and code-index updates.
