---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-06-14
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
- **Files Read:** 5
  - `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`
  - `/home/user/Bonsai/station/Playbook/Backlog.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 3
  - `/home/user/Bonsai/station/Playbook/Backlog.md` (P0 cleanup — 2 resolved items commented out)
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard updated)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (entry appended)
- **Tools Used:** Read, Edit, Write, Bash
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s
Scanned the P0 section of Backlog.md. Found **2 resolved P0 items** that had not been commented out after resolution:

1. **`[bug] Sensor hook commands use $PWD-walk-up`** — RESOLVED via v0.4.3 hotfix (PRs #105/#106, 2026-05-13). Status.md confirms: "v0.4.3 hotfix shipped — sensor hook commands now bake install-time absolute paths." The Backlog entry still said "Ships v0.4.3" (future tense) but v0.4.3 has long since shipped. Commented out with resolution note.

2. **`[feature] bonsai init / bonsai add need non-interactive flags`** — RESOLVED via v0.4.2 (PR #102, 2026-05-13). Status.md confirms: "v0.4.2 release shipped — --non-interactive --from-config shipped for both init and add." Commented out with resolution note.

No P0 items remain in the Backlog after cleanup. The `[research] Trial sentrux` was already correctly commented out (promoted to Status.md Pending 2026-05-07).

### Step 2 — Cross-reference with Status.md
Read Status.md. In Progress table is empty (no active work). Pending has one item: `[research] Trial sentrux on Bonsai repo` (blocked on Rust toolchain). Recently Done includes Plan 40 (Phases 1–3), v0.4.3 hotfix, v0.4.2 release, and prior history.

Cross-reference findings:
- The two resolved P0 items (above) matched "Recently Done" entries — cleaned up in Step 1.
- No other Backlog items matched In Progress or Recently Done in Status.md.
- Status.md Pending "Trial sentrux" is blocked on Rust toolchain. No Backlog items would unblock it (it needs external toolchain install, not a Bonsai code change).
- Plan 40 Phase 4 (update-delivery) is HELD per Status.md. The P1 Backlog entry "Full agent-drivable CLI" (added 2026-06-13) correctly represents this as the next major workstream.

### Step 3 — Cross-reference with Roadmap.md
Read Roadmap.md. Phase 1 is complete (all checkboxes checked, including the `bonsai validate` row added 2026-05-07 routine-digest). Current focus is Phase 2 — Extensibility.

P2/P3 Backlog items that align with Phase 2 milestones:
- **`[improvement] Self-update mechanism`** (P3 Big Bets) directly maps to Roadmap Phase 2 "Self-update mechanism" milestone — candidate for promotion to P2 when Phase 2 work begins.
- **`[improvement] Micro-task fast path`** (P3) maps to Roadmap Phase 2 "Micro-task fast path" — same.
- **`[feature] Custom item creator`** (P3) aligns with Phase 2 Extensibility theme.
- **`[feature] Full agent-drivable CLI`** (P1) is the immediate prerequisite that unblocks Phase 4 platform work.

No Backlog items reference deprecated approaches or fully completed phases (Phase 1 is done; no Backlog items still describe Phase 1 work as future).

No promotions performed — flagging the Phase-2-aligned P3 items for user review at phase boundary.

### Step 4 — Flag stale items
Many items are 30+ days old with no progress. Key stale groupings:

**Very stale (50+ days, added 2026-04-13 to 2026-04-25):**
- All of Group B (Code Quality & Testing) — added 2026-04-16, no movement
- All of Group C (OSS Readiness) — added 2026-04-16/17
- All of Group D (Catalog Expansion) — added 2026-04-16
- All of Group E (Workspace Improvements) — added 2026-04-15/16/21
- All of Group F (UI/UX Testing) — added 2026-04-20
- All P3 items — some from 2026-04-13 (61 days)

**Near-duplicate check:**
- P1 "[feature] Full agent-drivable CLI parity" (added 2026-06-13) supersedes the now-commented-out P0 non-interactive flags item. The P1 is broader and correctly positioned.
- Group C "[debt] Plan 26 candidate — skills frontmatter convention decision" and Group D "[research] Revisit concept-decisions research" are distinct enough to keep separate.
- No actionable near-duplicates found (prior cycles have already cleaned the obvious ones).

**Stale items flagged for user review** (not auto-removed — need user decision):
- Group B items (5 debt items, all 50+ days) — no progress; stale re-prioritization candidates
- Group C "[improvement] OSS polish — demo GIF/asciinema" — 50+ days, requires user recording; should confirm if still wanted
- Group D "[feature] Unbuilt catalog items" — 50+ days, phase dependency on concept-decisions review (also stale)
- P3 items added 2026-04-13 (Greenhouse companion app, Managed Agents, Archon analysis) — 61 days; confirm still relevant

No items lack clear context or rationale that would warrant removal without user input.

### Step 5 — Check for routine-generated items
Reviewed RoutineLog.md entries since 2026-05-07 (last backlog-hygiene run). The only substantive entry since then is the 2026-06-13 Plan 40 dispatch note, which added backlog items:
- P1: "Full agent-drivable CLI" — confirmed present in Backlog P1 (line 57, added 2026-06-13) ✓
- P2: symlink hardening security item — confirmed present in Backlog P2 (added 2026-06-13) ✓
- P2: `bonsai validate` bonsai.yaml drift improvement — confirmed present ✓
- P2: Plan 40 review nits — confirmed present ✓
- P2: `bonsai validate` can't pass on Bonsai repo (lock gitignored) — confirmed present ✓
- P2: Integrate plan-grilling as catalog ability — confirmed present ✓

All routine-generated findings from since the last backlog-hygiene run are captured in the Backlog. Nothing is missing.

### Step 6 — Promote ready items via issue-to-implementation
No items currently approved for immediate implementation by user. The P1 "Full agent-drivable CLI" is flagged as the "main thing" but requires a plan first (Backlog entry notes: "Promote to a plan + grill next session (/plan)"). Not auto-routing — presenting to user to confirm whether to kick off planning.

### Steps 7–8 — Log results and update dashboard
Completed as final actions of this run.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | Resolved P0 `$PWD-walk-up` bug still listed as active in Backlog | Backlog.md P0 | Commented out with resolution note (v0.4.3 shipped 2026-05-13) |
| 2 | High | Resolved P0 `--non-interactive` flags still listed as active in Backlog | Backlog.md P0 | Commented out with resolution note (v0.4.2 shipped 2026-05-13) |
| 3 | Low | Group B–F items (20+ entries) stale 50+ days without progress | Backlog.md P1-P2 | Flagged for user review — no auto-removal |
| 4 | Low | P3 items from 2026-04-13 (61 days) — Greenhouse, Managed Agents, Archon, Big Bets | Backlog.md P3 | Flagged for user review |
| 5 | Info | P3 items aligning with Phase 2 Roadmap milestones (Self-update, Micro-task fast path) | Backlog.md P3 / Roadmap Phase 2 | Noted — candidate for promotion when Phase 2 work begins |
| 6 | Info | P1 "Full agent-drivable CLI" ready for planning session | Backlog.md P1 | Flagged for user — suggests `/plan` next session |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **Stale Group B (Code Quality & Testing) items — 50+ days:** All 5 debt items (generate.go split, catalog test coverage, CLI test coverage, PTY smoke test, Plan-29/31 cosmetic nits). No progress. Re-prioritize, defer, or explicitly park as low-priority with a target phase.

2. **Stale Group C (OSS Readiness) items — 50+ days:** Demo GIF requires user recording. Skills frontmatter convention decision is a low-stakes pick. Confirm if both are still wanted.

3. **Stale Group D (Catalog Expansion) items — 50+ days:** "Revisit concept-decisions research" is a prerequisite gate for 3 other items. Until that gate is opened, Group D stays frozen. Is there a target date for this?

4. **P3 Big Bets (60+ days):** Greenhouse companion app, Managed Agents integration, Archon analysis — still in research/design phase. Confirm these are still on the long-term roadmap.

5. **P1 "Full agent-drivable CLI parity" — ready for planning:** User noted this is the "main thing" (added 2026-06-13). Backlog entry suggests `/plan` next session. Confirm when to start the planning workflow.

6. **P3 Phase-2-aligned items (Self-update mechanism, Micro-task fast path, Custom item creator):** When Phase 2 work begins, these should be promoted. Flag for next Roadmap Accuracy routine.

## Notes for Next Run

- P0 section is now clean (all items are commented-out with resolution notes or HTML comments). Next run should find zero active P0s unless a new critical issue surfaces.
- Group B–F stale item counts will grow until user makes a re-prioritization sweep. Consider a dedicated "backlog triage" session.
- P1 "Full agent-drivable CLI" should move to Status.md once planning starts — next backlog-hygiene run should catch it there.
- The gap between last run (2026-05-07) and this run (2026-06-14) was 38 days — more than 5× the 7-day cadence. Several findings (especially the two stale P0s) accumulated during this gap. Shorter cadence would catch these faster.
