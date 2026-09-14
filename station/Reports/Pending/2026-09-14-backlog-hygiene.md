---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-14
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 6 — `station/agent/Core/identity.md`, `station/agent/Core/memory.md`, `station/agent/Routines/backlog-hygiene.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s
Scanned the P0 section. Found 2 active P0 items (the sentrux entry was already commented out and in Status.md Pending).

Both P0 items cross-referenced against Status.md:
- `[bug] Sensor hook commands use $PWD-walk-up` → **resolved**: v0.4.3 hotfix shipped 2026-05-13 (bakes absolute install-time paths). No escalation needed — already done.
- `[feature] bonsai init/add need non-interactive flags` → **resolved**: v0.4.2 shipped 2026-05-13 (`--non-interactive --from-config`). No escalation needed — already done.

Result: P0 section cleared of both stale entries (replaced with resolution comments). No un-escalated P0s remain.

### Step 2 — Cross-reference with Status.md
Read Status.md. Key matches found:

**Resolved items removed from Backlog P0:**
1. `[bug] Sensor hook commands use $PWD-walk-up` — in Status.md Recently Done (v0.4.3, 2026-05-13). Removed from P0.
2. `[feature] bonsai init/add need non-interactive flags` — in Status.md Recently Done (v0.4.2, 2026-05-13). Removed from P0.

**Pending → Backlog unblock check:**
- Status.md Pending: `[research] Trial sentrux on Bonsai repo` — Blocked by Rust toolchain install. No Backlog item can unblock this; it's an environment prerequisite.

**Plans 40 & 41 shipped (June 2026) — Backlog items these close or affect:**
- Plan 40 shipped frozen schemas + root-relative scaffolding + `project validate`. P2 `[improvement] bonsai validate warn on bonsai/project.yaml drift` is still open (not in Status.md).
- Plan 41 shipped headless cores. P1 `[feature] Full agent-drivable CLI parity` is partially addressed (init/add/update have headless surfaces); remove coverage is incomplete per that P1 entry — still valid.
- P2 `[debt] Unify remove business logic` was added 2026-06-16 from Plan 41 review — still valid, not in Status.md.

### Step 3 — Cross-reference with Roadmap.md
Phase 1 is essentially complete (all items [x] or accounted for with notes). Two flags carried forward from the 2026-05-07 Roadmap Accuracy routine that were not resolved:
- Phase 1 "Better trigger sections" checkbox remains unchecked in Roadmap despite Plans 08/17/21 + context-guard shipping the bulk. Only deferred piece is Plan 08 C3 (P3 backlog). **Flag for user: update Roadmap checkbox with annotation.**
- Phase 1 has no row for `bonsai validate` (Plan 35, v0.4.0 headline). **Flag for user: add row.**

Phase 2 alignment:
- P1 `[feature] Full agent-drivable CLI parity` directly maps to Phase 2 Extensibility milestone (the headless CLI surface is the enabling layer). Candidate for promotion to a plan once current work clears.
- P3 `[feature] Self-update mechanism` aligns with Phase 2 — currently too speculative to promote.
- P3 `[improvement] Custom item creator` aligns with Phase 2 Extensibility.

No deprecated-approach items detected in P2/P3.

### Step 4 — Flag stale items
Gap since last run: 130 days (2026-05-07 → 2026-09-14). All items from April-June 2026 are 90-130 days without progress markers.

**Critical stale flag:**
- P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` — added 2026-04-22, reminder was for **2026-07-15**. Today is 2026-09-14 — the PAT is **~55 days overdue for rotation**. The fine-grained PAT (90-day default expiry from 2026-04-22 rotation) would have expired ~2026-07-21. Any release attempt since then would fail at the Homebrew formula step with `401 Bad credentials`. **Escalated to user review — potential release blocker.**

**Other stale items (no change made — flagged for awareness):**
- P1 `[ops] Routine bot PR pile-up` (added 2026-05-07, 130 days) — unclear if bot PRs have continued accumulating. Verify PR list on GitHub.
- P1 `[debt] Testing infrastructure for triggers and sensors` (added 2026-04-16, ~150 days) — no progress. Still a gap.
- P1 `[debt] Stale agent worktrees + branches` (added 2026-04-20, ~148 days) — the original worktrees are gone but pattern recurs per memory.md. May need a fresh sweep.
- P2 `[improvement] Plan 40 review nits` (added 2026-06-13, ~93 days) — Plan 40 is shipped. These nits (3 items) are now orphaned from their parent plan. Decide: fix or discard.
- Group A bookkeeping `[bookkeeping] Retroactively trim Backlog entries to NoteStandards` — the backlog itself still has many verbose entries from before the NoteStandards rule. Not cleaned up.

**Near-duplicate check:**
- P1 `[feature] Full agent-drivable CLI parity` and the removed P0 `[feature] bonsai init/add non-interactive` were converging topics. With the P0 removed, the P1 now cleanly covers the full surface — no redundancy.
- P2 `[improvement] Plans Index file` and P2 `[improvement] Plan archiving` are distinct items in Group E but interdependent; consider bundling when either is promoted.

### Step 5 — Check routine-generated items since last backlog-hygiene run (2026-05-07)
Read RoutineLog.md. No routine entries after 2026-05-07 (all routines appear to have gone silent since May; dashboard shows all `Last Ran` dates in May 2026). The 130-day gap means this is the first backlog-hygiene run in that window.

From the 2026-05-07 Backlog Hygiene flags that were NOT captured in Backlog:
1. **code-index.md staleness** — medium drift, unfixed. Plans 41/40 (June 2026) have since shipped substantial changes that would compound the staleness. Not added to Backlog. **Flag for user.**
2. **Broken nav link `agent/Skills/bonsai-model.md`** — flagged 2026-05-07, not added to Backlog. **Flag for user.**
3. **INDEX.md arch diagram drift** — flagged 2026-05-07, not added to Backlog. **Flag for user.**
4. **Phase 1 "Better trigger sections" Roadmap box** — flagged 2026-05-07 and 2026-05-07 Roadmap Accuracy routine both raised this. Still not resolved. **Flag for user.**

From 2026-05-07 Doc Freshness Check flags (high severity: root `Bonsai/CLAUDE.md` project-structure tree badly stale across Plans 22-35) — Plans 40/41 have since added more items. The P2 `[improvement] Add root Bonsai/CLAUDE.md tree-drift check` exists in Backlog but the underlying drift is still present and now worse. Not auto-adding a new entry; existing entry covers it.

### Step 6 — Promote ready items via issue-to-implementation
No items promoted autonomously. The HOMEBREW_TAP_TOKEN item is time-sensitive but requires user action (PAT rotation, not code work). The P1 `[feature] Full agent-drivable CLI parity` is a strong candidate for promotion to a plan but user confirmation is required before starting the workflow.

### Steps 7-8 — Log + dashboard update
Performed as post-procedure actions (RoutineLog.md + routines.md updated).

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | Two P0 items resolved in May 2026 still active in Backlog P0 | `Backlog.md` P0 section | Removed both; added resolution comments |
| 2 | Critical | HOMEBREW_TAP_TOKEN PAT expired ~55 days ago (reminder was 2026-07-15) | `Backlog.md` P1 | Flagged for user; no autonomous action |
| 3 | Medium | code-index.md staleness from May 2026 routine not in Backlog; Plans 40/41 compounded it | Unfiled finding | Flagged for user; not auto-added |
| 4 | Medium | Broken nav link `agent/Skills/bonsai-model.md` not in Backlog | Unfiled finding | Flagged for user; not auto-added |
| 5 | Medium | Roadmap Phase 1 "Better trigger sections" checkbox still unchecked after 130 days | `Roadmap.md` | Flagged for user |
| 6 | Medium | Roadmap Phase 1 missing `bonsai validate` row (Plan 35 headline) | `Roadmap.md` | Flagged for user |
| 7 | Low | Plan 40 review nits (P2, added 2026-06-13) now orphaned from parent plan | `Backlog.md` P2 | Flagged — decide fix or discard |
| 8 | Low | Group A NoteStandards sweep not done — many verbose entries remain | `Backlog.md` | Flagged; low urgency |
| 9 | Low | Routine bot PR pile-up item (P1, 130 days) — unclear current state | `Backlog.md` P1 | Flagged for verification |
| 10 | Info | 130-day gap since last routine runs; all routines overdue | Dashboard | Flagged; other routines need dispatch |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[CRITICAL] Rotate HOMEBREW_TAP_TOKEN now** — Fine-grained PAT rotated 2026-04-22 with 90-day default expiry. Due by ~2026-07-21; now ~55 days overdue. Next release will fail at Homebrew formula step with 401. See memory.md recovery steps.

2. **[MEDIUM] Roadmap.md cleanup needed** — Two unchecked Phase 1 items are effectively done: (a) "Better trigger sections" (Plans 08/17/21 + context-guard shipped; only P3 deferred piece remains — annotate `[x]`); (b) add `bonsai validate` milestone row. Both flagged by 2026-05-07 Roadmap Accuracy routine and unresolved.

3. **[MEDIUM] Unfiled doc-drift items** — Three items from 2026-05-07 routine flags never landed in Backlog: (a) code-index.md medium staleness (Plans 40/41 made it worse), (b) broken nav link `agent/Skills/bonsai-model.md`, (c) INDEX.md arch diagram drift. Add to Backlog or fix in one shot.

4. **[LOW] Plan 40 review nits (P2)** — Three non-blocking items (manifest `created` via yamlScalar, stale doc-comment L~388, missing date-format guard in project.go). Plan 40 is shipped; decide whether to fix in a patch or discard.

5. **[INFO] All other routines overdue** — Last run for all routines was May 2026 (~130 days ago). Recommend dispatching: Dependency Audit, Vulnerability Scan, Doc Freshness Check (all 7-day routines, all ~120 days overdue).

## Notes for Next Run

- After HOMEBREW_TAP_TOKEN rotation, update PAT expiry reminder in Backlog P1 with new target date.
- If Plans 40/41 doc drift is addressed (code-index, root CLAUDE.md), the Doc Freshness Check routine will have a cleaner run.
- P1 `[feature] Full agent-drivable CLI parity` should move to Status.md Pending and get a plan if user confirms it as the next major effort.
- Consider bundling "Plans Index file" + "Plan archiving" items (Group E) into a single small plan — they are interdependent and ~5 months stale.
- The 130-day routine gap suggests the loop.md dispatch mechanism may have been paused. Verify and re-enable if so.
