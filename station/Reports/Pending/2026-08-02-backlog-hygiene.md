---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-02
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
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 1 — `/home/user/Bonsai/station/Playbook/Backlog.md` (3 resolved items commented out)
- **Tools Used:** Read, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s

Found 2 active P0 items in Backlog:

1. **[bug] Sensor hook commands use $PWD-walk-up** — Checked Status.md: NOT in Pending or In Progress, but found in Recently Done as "v0.4.3 hotfix shipped" (2026-05-13). This P0 is **resolved**, not a misplaced escalation.

2. **[feature] bonsai init / bonsai add need non-interactive flags** — Checked Status.md: NOT in Pending or In Progress, but found in Recently Done as "v0.4.2 release shipped" (2026-05-13). This P0 is **resolved**, not a misplaced escalation.

No P0 items require escalation to Status.md — both are already resolved.

### Step 2 — Cross-reference with Status.md

Reviewed Status.md In Progress (empty), Pending (1 item: sentrux trial), and Recently Done.

**Items to remove** (matching Status.md Recently Done):

1. **P0 bug "Sensor hook commands use $PWD-walk-up"** → resolved by v0.4.3 (PR #105/#106, 2026-05-13). Commented out with resolution note.

2. **P0 feature "bonsai init / bonsai add need non-interactive flags"** → resolved by v0.4.2 (PR #102, 2026-05-13). Commented out with resolution note.

3. **P1 feature "Full agent-drivable (non-interactive) CLI parity: init / update / add / remove"** → resolved by Plan 41 (PRs #120/#122/#123/#121/#125, 2026-06-16). Every mutating command now has headless `*Result` cores + JSONL/exit contract. Commented out with resolution note.

**Pending cross-check:** Status.md Pending has one item: "[research] Trial sentrux on Bonsai repo" — blocked on Rust toolchain install. No Backlog item could unblock this (it's a toolchain gap, not a code task).

### Step 3 — Cross-reference with Roadmap.md

**Phase 1** is complete (all items checked, including validate command added in 2026-05-07 routine-digest).

**Phase 2 (Extensibility)** alignment check:
- "Self-update mechanism" (unchecked) → P3 Backlog "[improvement] Self-update mechanism" under Big Bets — no promotion warranted yet.
- "Micro-task fast path" (unchecked) → P3 Backlog "[improvement] Micro-task fast path" — no promotion warranted.
- "Template variables expansion" (unchecked) → no Backlog item found. Not flagged (likely requires user direction).
- "Custom item detection" (checked) → custom ability scan shipped and in Backlog as closed.

**Phase 3/4** items: all in P3 Big Bets (Managed Agents, Greenhouse). No promotion warranted.

**Deprecated approach check:** The two removed P0 items referenced v0.4.2/v0.4.3 work. Those are now removed from the active backlog.

No P2/P3 items warrant promotion to P1 based on current roadmap phase boundaries.

### Step 4 — Flag stale items

**High-priority stale flag (user action required):**

- **P1 "[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder"** — Added 2026-04-22. Reminder was set for ~2026-07-15. Today is 2026-08-02 — the reminder date is **18 days overdue**. No RoutineLog or Status.md entry confirms the PAT was rotated. Fine-grained PATs default to 90-day expiry. If not rotated, the next release will fail with `401 Bad credentials` at the GoReleaser Homebrew tap step. **FLAG: user must verify PAT rotation status immediately.**

**Stale items at same priority 100+ days without progress (P2/P3):**

| Item | Added | Days Old | Note |
|------|-------|----------|------|
| [improvement] Consolidate FieldNotes usage | 2026-04-15 | 109 | Unclear scope; still valid |
| [improvement] Post-update backup merge hint | 2026-04-16 | 108 | Small, ready to pick up |
| [debt] Break up generate.go | 2026-04-16 | 108 | Recurring high-churn file |
| [debt] catalog/ test coverage 0% | 2026-04-16 | 108 | No progress |
| [debt] CLI cmd/ test coverage 0% | 2026-04-16 | 108 | No progress |
| [research] Revisit concept-decisions | 2026-04-16 | 108 | No progress |
| [improvement] Plan archiving | 2026-04-16 | 108 | Plans/Archive/ exists but scaffolding manifest not updated |
| [improvement] OSS polish — demo GIF | 2026-04-16 | 108 | Requires user recording |
| [debt] Plan 26 candidate — skills frontmatter | 2026-04-22 | 102 | Cosmetic, still unresolved |
| [improvement] Plans Index file | 2026-04-21 | 103 | Revisit with Plan archiving |
| [bookkeeping] Trim Backlog to NoteStandards | 2026-04-25 | 99 | Group A, self-referential |

None have clear rationale for staying stale. Flagged for user re-prioritization consideration (not demoted autonomously).

**Near-duplicates:** No new near-duplicates found. The prior CHANGELOG duplication between Group C and Group D was already resolved in earlier cycles.

### Step 5 — Check for routine-generated items since last run (2026-05-07)

Reviewed RoutineLog entries from 2026-05-07 onward. Non-routine session entries found: Plan 40 dispatch (2026-06-13) and Plan 41 (2026-06-16).

Plan 40 filed these Backlog items — verified all present:
- P2 "[security] Harden all scaffolding writes against symlink substitution" ✓
- P2 "[improvement] bonsai validate warn on .bonsai/project.yaml drift" ✓
- P2 "[improvement] Plan 40 review nits" ✓
- P2 "[bug] bonsai validate can't pass on Bonsai repo itself" ✓
- P2 "[feature] Integrate plan-grilling as first-class catalog ability" ✓

Plan 41 filed these Backlog items — verified all present:
- P2 "[security] Website npm vuln tree — astro upgrade breaks build" ✓
- P2 "[debt] Unify remove business logic" ✓

All routine-generated findings are captured. No uncaptured items to flag.

### Step 6 — Promote ready items via issue-to-implementation

No items are pre-approved for implementation. The HOMEBREW_TAP_TOKEN item could be actionable immediately if the user confirms the PAT needs rotation (ops task, not code). Presenting to user for decision — not auto-promoting.

### Steps 7–8

Logged to RoutineLog.md and dashboard updated (handled in post-procedure steps).

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | P0 bug "sensor $PWD-walk-up" — resolved in v0.4.3 but still active in Backlog P0 | Backlog.md P0 | Commented out with resolution note |
| 2 | HIGH | P0 feature "non-interactive flags" — resolved in v0.4.2 but still active in Backlog P0 | Backlog.md P0 | Commented out with resolution note |
| 3 | HIGH | P1 "Full CLI parity" — resolved by Plan 41 but still active in Backlog P1 | Backlog.md P1 | Commented out with resolution note |
| 4 | HIGH | HOMEBREW_TAP_TOKEN PAT reminder overdue by 18 days — no confirmation of rotation | Backlog.md P1 | Flagged for user — no autonomous action |
| 5 | MEDIUM | P0 section now empty (all items resolved or in Status Pending) | Backlog.md | Noted — structure preserved with comments |
| 6 | LOW | 11 items at P2/P3 100+ days old without visible progress | Backlog.md P2/P3 | Flagged for user re-prioritization |
| 7 | INFO | No uncaptured routine findings from Plans 40/41 (2026-06-13/16) | RoutineLog.md | No action needed |

## Errors & Warnings

None.

## Items Flagged for User Review

1. **[URGENT] HOMEBREW_TAP_TOKEN PAT rotation** — The calendar reminder (added 2026-04-22, set for ~2026-07-15) is 18 days overdue. Fine-grained PATs default to 90-day expiry; the PAT was rotated 2026-04-22, putting expiry around 2026-07-21. If not rotated, next release will fail at the Homebrew tap step with `401 Bad credentials`. Verify and rotate immediately if needed. The Backlog P1 item remains; once confirmed rotated, comment it out with the new expiry date.

2. **[INFO] P0 section is now empty** — Both active P0 items were resolved months ago (v0.4.2 on 2026-05-13, v0.4.3 on 2026-05-13). No P0 work is currently blocking. Recommend reviewing whether the sentrux trial (in Status Pending, blocked on Rust toolchain) still warrants P0 designation or should be downgraded.

3. **[LOW] 11 items 100+ days stale at P2/P3** — Oldest items (Group B debt, Group D research, Group E workspace improvements) have been in backlog since 2026-04-13 to 2026-04-25. Consider a backlog triage session to demote, remove, or schedule the most actionable ones.

## Notes for Next Run

- P0 section is now effectively empty; next run should confirm no new P0s have been filed since 2026-08-02.
- HOMEBREW_TAP_TOKEN status should be confirmed — if rotated, the P1 ops item can be commented out.
- The website npm vuln (P2, added 2026-06-16) is now 47 days old — if not addressed, consider promoting to P1 given the HIGH severity esbuild finding.
- Plans 40 and 41 filed several P2 items (2026-06-13/16) — next hygiene run should check if any have been picked up.
- 87 days elapsed between this run and the last (2026-05-07). Three items were resolved in that gap but not cleaned up. Recommend resuming 7-day cadence.
