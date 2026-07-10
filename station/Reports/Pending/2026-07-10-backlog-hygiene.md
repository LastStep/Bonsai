---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-10
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07 (before this run — 64 days overdue)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 4 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 1 — `/home/user/Bonsai/station/Playbook/Backlog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s

Read Backlog.md P0 section. Found 2 active P0 items and 1 existing HTML-commented item (sentrux — already in Status.md Pending).

**P0 item 1:** `[bug] Sensor hook commands use $PWD-walk-up` — The item says "Ships v0.4.3." Verified against Status.md: "v0.4.3 hotfix shipped — sensor hook commands now bake install-time absolute paths" (2026-05-13). **Resolved** — not in Status.md In Progress or Pending because it's done. Action: comment out (Step 2 resolution).

**P0 item 2:** `[feature] bonsai init / bonsai add need non-interactive flags` — Verified against Status.md: "v0.4.2 release shipped — `bonsai init`/`add` `--non-interactive --from-config <path>`" (2026-05-13). **Resolved** — action: comment out (Step 2 resolution).

No P0s exist in Backlog that need immediate promotion to Status.md (the only one that should be there — sentrux — already is).

### Step 2: Cross-reference with Status.md

Read Status.md. In Progress: empty. Recently Done: 10+ rows covering v0.4.2, v0.4.3, Plan 38, Plan 39, Plan 40 (P1-3), Plan 41.

Cross-reference findings:

| Backlog Item | Status.md Match | Action |
|---|---|---|
| P0: `[bug] $PWD-walk-up sensor hook` | v0.4.3 hotfix shipped 2026-05-13 | Commented out ✓ |
| P0: `[feature] non-interactive flags for init/add` | v0.4.2 released 2026-05-13 | Commented out ✓ |
| P1: `[feature] Full agent-drivable CLI parity (init/update/add/remove)` | Plan 41 shipped 2026-06-16: all 4 cmds have headless cores + JSONL/exit contract | Commented out ✓ |

After this step, the P0 section contains zero active bullet items (3 HTML comments only).

Checked Status.md Pending for "Blocked By" items: only one Pending row exists — sentrux trial blocked on Rust toolchain. No Backlog item directly unblocks it (it's an external dependency, not a Bonsai feature).

### Step 3: Cross-reference with Roadmap.md

Read Roadmap.md. Phase 1 is fully complete — all checkboxes checked. Current work is moving toward Phase 2 (Extensibility) and Phase 3 (Cloud & Orchestration).

Phase 2 milestones unchecked:
- Self-update mechanism → Backlog P3 `[improvement] Self-update mechanism` — consider promoting to P2 now that Phase 2 is the active horizon
- Template variables expansion → no Backlog tracking item (gap)
- Micro-task fast path → Backlog P3 `[improvement] Micro-task fast path` — consider promoting to P2

Phase 3 milestones:
- Managed Agents integration → Backlog P3 Big Bets `[feature] Managed Agents integration` — aligned
- Greenhouse companion app → Backlog P3 Big Bets `[feature] Greenhouse companion app` — aligned

No items appear to reference deprecated approaches. No completed-phase references found that are misleading.

**Flag for user:** "Template variables expansion" (Phase 2 milestone) has no tracking entry in Backlog. Consider adding one.

### Step 4: Flag stale items

With today's date of 2026-07-10 and the last hygiene run on 2026-05-07 (64 days ago), many items are well past 30 days at the same priority with no recorded progress:

**P1 items — stale (no progress 60+ days):**
- `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — added 2026-04-22, rotation due ~2026-07-15. **URGENT: 5 days away.** The token expires in ~5 days. Flag: rotate before 2026-07-15.
- `[ops] Routine bot PR pile-up` — added 2026-05-07, ~64 days. Still open; no resolution noted.
- `[debt] Testing infrastructure for triggers and sensors` [Group B] — added 2026-04-16, ~85 days. No progress. Consider demoting to P2 or scoping a plan.
- `[debt] Stale agent worktrees + branches accumulating` — added 2026-04-20, ~81 days. Likely partially addressed by session cleanup; consider reassessing.

**P2 items stale at 30+ days (selected highlights):**
- Group E items (Plan archiving, Plans Index, Consolidate FieldNotes, Post-update backup merge hint, Port statusLine) — all added 2026-04-15 to 2026-04-22, ~79-86 days. No progress.
- Group B remaining items (break up generate.go, catalog test coverage, CLI cmd test coverage, PTY smoke test) — all added 2026-04-16 to 2026-04-20, ~81-85 days. No progress.
- Group C items (OSS demo GIF, skills frontmatter convention, validateOverlay placeholder) — all added 2026-04-16 to 2026-05-13, ~58-85 days. No progress.
- Group F items (AltScreen release note, Fill Deviations more aggressively) — added 2026-04-20, ~81 days.

No obvious near-duplicates found in the current Backlog. The changelog near-duplicate flagged in the 2026-04-21 run appears to have been resolved (Group C no longer has a duplicate changelog entry).

### Step 5: Check for routine-generated items

Read RoutineLog.md. Last backlog-hygiene run was 2026-05-07. Relevant entries since then:

- **2026-06-13 (Plan 40 dispatch):** Backlog P2 items added inline during session: symlink hardening, validate identity-drift warning, Plan 40 review nits, validate-can't-pass-on-Bonsai-repo. All confirmed present in Backlog P2.
- **2026-06-16 (Plan 41 dispatch):** "Unify remove cinematic/headless logic" filed to Backlog P2. Confirmed present.

No routine runs between 2026-05-07 and 2026-07-10 (64-day gap). No uncaptured findings — all session-surfaced items were captured inline by the dispatching sessions.

No routine-flagged findings require new Backlog entries.

### Step 6: Promote ready items via issue-to-implementation

No items are approved for immediate implementation by the user. The HOMEBREW_TAP_TOKEN PAT expiry (P1) is time-sensitive but is an ops task (user rotates PAT in GitHub settings), not a code implementation task. Presented to user via this report.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | P0 `[bug] $PWD-walk-up sensor hook` was resolved in v0.4.3 (2026-05-13) but remained as active P0 | Backlog.md L49 | Commented out — resolved |
| 2 | medium | P0 `[feature] non-interactive flags for init/add` was resolved in v0.4.2 (2026-05-13) but remained as active P0 | Backlog.md L53 | Commented out — resolved |
| 3 | medium | P1 `[feature] Full agent-drivable CLI parity` was resolved in Plan 41 (2026-06-16) but remained as active P1 | Backlog.md L57 | Commented out — resolved |
| 4 | high | **URGENT:** `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — rotation due ~2026-07-15, 5 days away | Backlog.md P1 | Flagged for user — rotate PAT now |
| 5 | low | P0 section is now empty of active items (all 3 were resolved) — section is vestigial | Backlog.md P0 | No action needed; section header can remain |
| 6 | low | Phase 2 Roadmap milestone "Template variables expansion" has no Backlog tracking entry | Roadmap.md Phase 2 | Flagged for user — add Backlog entry if desired |
| 7 | low | Multiple P1/P2 items stale 60-85 days with no recorded progress (Groups B, C, E) | Backlog.md | Flagged for user — re-prioritize at next planning session |
| 8 | low | P3 items "Self-update mechanism" and "Micro-task fast path" align with now-active Phase 2 | Backlog.md P3 | Flagged for user — consider promoting to P2 |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **URGENT (5 days): Rotate HOMEBREW_TAP_TOKEN PAT** — The P1 backlog item `[ops] HOMEBREW_TAP_TOKEN PAT expiry` (added 2026-04-22) notes rotation is due ~2026-07-15. Today is 2026-07-10. Go to GitHub → Settings → Secrets → `HOMEBREW_TAP_TOKEN` on `LastStep/Bonsai`, rotate before the deadline. An expired PAT causes GoReleaser to silently skip the Homebrew formula update during releases.

2. **Backlog cleanup: 3 resolved items removed** — The P0 section and one P1 item were commented out as resolved by v0.4.2, v0.4.3, and Plan 41. P0 section now contains zero active items.

3. **Stale P1/P2 items (60-85 days, no progress)** — Groups B, C, and E workspace improvement items have been sitting at their priority levels for 80+ days. Consider scoping a housekeeping plan or demoting to P3. Specific candidates: `break up generate.go`, `catalog test coverage`, `CLI cmd test coverage`, `Plan archiving / Plans Index`.

4. **Phase 2 milestone gap** — Roadmap Phase 2 lists "Template variables expansion" with no corresponding Backlog tracking entry. Consider adding one.

5. **P3 → P2 promotion candidates** — Phase 1 is complete. `[improvement] Self-update mechanism` and `[improvement] Micro-task fast path` align with Phase 2 milestones. Consider promoting both from P3 to P2.

## Notes for Next Run

- P0 section is now empty of active items. If no new P0s emerge, the section can be left as a placeholder for the format.
- The routine is 64 days overdue (ran 2026-05-07, frequency is 7 days). Other routines are similarly overdue (all last ran between 2026-05-04 and 2026-05-07). A routine-digest session is recommended to address the backlog of overdue routines.
- HOMEBREW_TAP_TOKEN PAT expiry is the only time-critical item; everything else can wait for a normal planning session.
