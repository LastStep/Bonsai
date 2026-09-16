---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-16
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 6 — `station/agent/Routines/backlog-hygiene.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s

Scanned P0 section. Found 2 active P0 items and cross-referenced with Status.md:

1. **[bug] Sensor hook commands use `$PWD`-walk-up** — Status.md Recently Done confirms v0.4.3 hotfix shipped 2026-05-13 (PRs #105/#106). Fully resolved. Commented out with resolution note.

2. **[feature] `bonsai init`/`add` need non-interactive flags** — Status.md Recently Done confirms v0.4.2 shipped 2026-05-13 (PR #102) with `--non-interactive` + `--from-config`. Fully resolved. Commented out with resolution note.

Result: P0 section is now empty. Marked with `_(no active P0 items)_` notice.

### Step 2 — Cross-reference with Status.md

Read Status.md In Progress (none) and Recently Done:

- **Plan 41 shipped (2026-06-16)** — "Every mutating cmd (init/add/update/remove) has a pure `*Result` headless core + JSONL/exit contract." This directly resolves the P1 item "Full agent-drivable (non-interactive) CLI parity: init / update / add / remove". Commented out with resolution note.

- **sentrux trial** — Already commented out in Backlog with promotion notice (promoted to Status.md Pending 2026-05-07). No action needed.

- No Pending items in Status.md with "Blocked By" that could be unblocked via a Backlog resolution. The only Pending item (sentrux trial) is blocked on Rust toolchain installation — a user action, not a Backlog item.

### Step 3 — Cross-reference with Roadmap.md

Roadmap Phase 1 is fully complete (all boxes checked).

Phase 2 — Extensibility milestones:
- Self-update mechanism → P3 in Backlog (appropriate)
- Template variables expansion → not in Backlog (low priority, not flagged)
- Micro-task fast path → P3 in Backlog (appropriate)

No P2/P3 items warranting promotion to P1 based on Phase 2 alignment. Phase 2 work is not yet actively prioritized.

No deprecated approach references found requiring removal.

### Step 4 — Flag stale items

**HOMEBREW_TAP_TOKEN PAT (P1, URGENT):** PAT was rotated 2026-04-22 with a ~2026-07-15 expiry calendar note. As of 2026-09-16, the PAT is approximately 2 months overdue for rotation. A next release (e.g., shipping v0.5.0 tag or Plan 42 MCP server) will fail at the GoReleaser brew step. Updated the Backlog item with an OVERDUE urgency marker.

**Stale P1 items (150+ days, no progress):**
- `[debt] Testing infrastructure for triggers and sensors` — added 2026-04-16, still P1
- `[debt] Stale agent worktrees + branches` — added 2026-04-20/2026-04-21, still P1

Both are legitimate debt items but have shown no movement in 5 months. Flagged for user re-prioritization (demote to P2 or schedule?).

**Stale P1 ops item:**
- `[ops] Routine bot PR pile-up` — added 2026-05-07. 9 stale PRs were closed, but the upstream fix (change cloud routine behavior) has not happened. Still relevant; no change.

**Duplicate entry identified:**
- `[feature] Changelog generation skill + release changelogs` appears in both **Group C: OSS Readiness** and **Group D: Catalog Expansion** with identical text. One of these should be removed. Decision required: keep in C (OSS scope) or D (catalog feature scope)? Flagged for user.

### Step 5 — Check for routine-generated items since last run (2026-05-07)

RoutineLog entries since 2026-05-07:
- 2026-06-13 — Plan 40 dispatch log entry: mentions new Backlog P2 items were filed during that session.

Verified those items now appear in Backlog P2:
- `[security] Harden all scaffolding writes against symlink substitution` ✓
- `[improvement] bonsai validate warn on .bonsai/project.yaml drift` ✓
- `[improvement] Plan 40 review nits (non-blocking)` ✓
- `[bug] bonsai validate can't pass on the Bonsai repo itself` ✓
- `[security] Website npm vuln tree — astro upgrade breaks npm run build` ✓
- `[debt] Unify remove business logic` ✓

All routine-generated findings are captured. No uncaptured routine findings.

**Uncaptured finding from memory.md:** memory.md Work State references "MCP server = Plan 42 (go-sdk, stdio `bonsai mcp`)" as an open follow-up from Plan 41. This is a concrete planned feature with no Backlog entry. Flagged for user — should this be added as P1 or P2?

### Steps 6–8

No items approved for implementation promotion — routing decisions deferred to user.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 bug (sensor hook `$PWD`-walk-up) already shipped as v0.4.3 | Backlog P0 | Commented out with resolution note |
| 2 | resolved | P0 feature (non-interactive flags) already shipped as v0.4.2 | Backlog P0 | Commented out with resolution note |
| 3 | resolved | P1 feature (agent-drivable CLI parity) already shipped as Plan 41 | Backlog P1 | Commented out with resolution note |
| 4 | high | HOMEBREW_TAP_TOKEN PAT ~2 months overdue for rotation | Backlog P1 | Added OVERDUE urgency marker to item |
| 5 | medium | Duplicate "Changelog generation skill" entry in Group C and Group D | Backlog P2 Group C + D | Flagged for user decision |
| 6 | medium | Plan 42 MCP server referenced in memory.md but absent from Backlog | memory.md Work State | Flagged for user — no auto-add |
| 7 | low | P1 [debt] Testing infra — 150+ days stale, no progress | Backlog P1 | Flagged for re-prioritization |
| 8 | low | P1 [debt] Stale worktrees + branches — 150+ days stale, no progress | Backlog P1 | Flagged for re-prioritization |

## Errors & Warnings

None.

## Items Flagged for User Review

1. **[URGENT] Rotate HOMEBREW_TAP_TOKEN PAT immediately** — Due ~2026-07-15, now ~2 months overdue. Next release will fail at GoReleaser brew step with 401 error. Rotate under `LastStep/Bonsai` repo secrets.

2. **[DECISION] Duplicate "Changelog generation skill" entry** — Identical bullet appears in Group C (OSS Readiness) and Group D (Catalog Expansion). Which group should it live in? Recommend removing from Group C (pure catalog feature) and keeping in Group D.

3. **[DECISION] Plan 42 MCP server missing from Backlog** — memory.md Work State names it as a concrete planned next step ("go-sdk, stdio `bonsai mcp`"). Should this be added as a P1 backlog item, or is it already being planned outside the Backlog?

4. **[RE-PRIORITIZE] P1 testing infrastructure (added 2026-04-16, 150+ days stale)** — No movement. Demote to P2, or assign to a plan?

5. **[RE-PRIORITIZE] P1 stale worktrees + branches (added 2026-04-20/04-21, 150+ days stale)** — The one-time sweep was recommended but never executed. Demote to P2 or schedule?

## Notes for Next Run

- P0 section is now empty. If a new P0 surfaces, verify it appears in Status.md Pending within the same session.
- HOMEBREW_TAP_TOKEN rotation is the most urgent actionable item. Check first.
- Several P1 items are aging into P2 territory — if user doesn't act on them by next cycle, consider recommending demotion.
- Watch for Plan 42 MCP server plan creation — once a plan exists, a Backlog entry should be added or the plan reference added to Status.md.
