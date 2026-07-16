---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-16
status: partial
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07 (before this run — 69 days gap)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (routine completed; step 6 requires user decision before proceeding)
- **Duration:** ~8 min
- **Files Read:** 5 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; cross-referenced each P0 item against Status.md In Progress and Recently Done.
- **Result:** Both P0 items are fully resolved and already shipped — neither needed escalation, both needed removal (see Step 2 for actions taken).
  - `[bug] Sensor hook commands use $PWD-walk-up` → RESOLVED in v0.4.3 (Status.md 2026-05-13). Hook commands now bake install-time absolute paths.
  - `[feature] bonsai init / bonsai add need non-interactive flags` → RESOLVED in v0.4.2 (Status.md 2026-05-13). `--non-interactive --from-config` shipped for both commands.
- **Issues:** No misplaced P0s — the P0 section was populated with stale-resolved items that should have been cleaned up at the 2026-05-07 run or sooner.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress, Pending, and Recently Done tables; cross-referenced each against Backlog items.
- **Result:** Three items found in Status.md as shipped but still present in Backlog. All three commented out with resolution notes:
  1. P0 `[bug] Sensor hook commands $PWD-walk-up` → resolved v0.4.3, Status.md 2026-05-13
  2. P0 `[feature] bonsai init/add non-interactive flags` → resolved v0.4.2, Status.md 2026-05-13
  3. P1 `[feature] Full agent-drivable CLI parity (init/update/add/remove)` → resolved Plan 41, Status.md 2026-06-16
- **Cross-ref blocking check:** Status.md Pending has one item: `[research] Trial sentrux on Bonsai repo` — blocked on Rust toolchain. No Backlog item would unblock it (requires user install of rustup/cargo).
- **Issues:** None beyond the three stale items removed above.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; scanned Phase 2/3/4 milestones against Backlog P2/P3 items.
- **Result:**
  - Phase 1 is fully shipped (all checkboxes ticked) — no Backlog items reference Phase 1 milestones inappropriately.
  - Phase 2 alignment:
    - `Self-update mechanism` → P3 Backlog entry `[improvement] Self-update mechanism` (added 2026-04-13). Tagged match — no promotion warranted yet (Phase 2 not started).
    - `Template variables expansion` → **no Backlog entry exists**. Flagged for user (see Findings table, finding #7).
    - `Micro-task fast path` → P3 Backlog entry `[improvement] Micro-task fast path` (added 2026-04-15). Match confirmed.
  - Phase 3: `Managed Agents integration` → P3 `[feature] Managed Agents integration`. `Greenhouse companion app` → P3 `[feature] Greenhouse companion app`. Both aligned.
  - Phase 4: `Catalog marketplace` and `Plugin system` → **no Backlog entries**. These are long-horizon Phase 4 items; not flagging as gaps (too early).
  - No Backlog items reference deprecated approaches or completed Phase 1 milestones.
- **Issues:** One gap — Phase 2 `Template variables expansion` has no Backlog entry. Low severity (Phase 2 not actively started).

### Step 4: Flag stale items
- **Action:** Scanned all Backlog items for age (today 2026-07-16 — items added before 2026-06-16 are 30+ days old) and for clarity of context.
- **Result:** The majority of Backlog items are 30-90+ days old with no activity. Specific flags:
  - **P1 URGENT — PAT expiry overdue:** `[ops] HOMEBREW_TAP_TOKEN PAT expiry` (added 2026-04-22). Reminder was for ~2026-07-15 — that was YESTERDAY. The PAT was set 2026-04-22 with 90-day default expiry → expires ~2026-07-21 (5 days from now). Requires immediate user action to rotate before next release.
  - **Group B (Code Quality/Testing):** 8 items added 2026-04-16 through 2026-04-24 (90+ days). No progress on any. Flag for re-prioritization — test infra in particular may have grown more urgent with Plan 41 adding new code paths.
  - **Group C (OSS Readiness):** 3 items from 2026-04-16 through 2026-05-13. Demo GIF (user-action-required), ability-name tab completion (no progress), skills frontmatter convention decision (no progress). Consider re-prioritizing or closing the frontmatter question.
  - **Group E (Workspace Improvements):** 5 items from 2026-04-16 through 2026-04-21. `Plans/Archive` folder structure and Plans Index (still not done — Plans/Active/ does contain active plan 40+41 which should be in Archive by now). `FieldNotes consolidation` is 91 days old.
  - **P1 `[ops] Routine bot PR pile-up`:** (added 2026-05-07) — 70 days old. Structural fix still pending. Cloud routine may still be generating PRs.
  - **P2 `[bug] bonsai validate can't pass on Bonsai repo itself`:** (added 2026-06-13) — directly blocks Plan 40 dogfood. 33 days old. Should be examined before next Plan 40/validate work.
  - **P2 `[security] Website npm vuln tree`:** (added 2026-06-16) — 30 days old. 6 open Dependabot alerts including HIGH severity esbuild/vite issues. Overdue for the vulnerability-scan routine's attention.
  - No near-duplicates found beyond the resolved CLI-parity items (P0/P1 were effectively the same item at different stages — now all resolved).
- **Issues:** PAT expiry is the most time-critical finding from this step.

### Step 5: Check for routine-generated items
- **Action:** Reviewed RoutineLog.md for entries since last Backlog Hygiene run (2026-05-07).
- **Result:** The RoutineLog shows NO routine executions between 2026-05-07 and 2026-07-16 (69 days). The only post-2026-05-07 entry is the "2026-06-13 — Plan 40 dispatch" operational log entry — not a routine execution. All 7 routines are overdue:
  - Backlog Hygiene: due 2026-05-14 (63 days overdue — this run)
  - Dependency Audit: due 2026-05-11 (66 days overdue)
  - Doc Freshness Check: due 2026-05-11 (66 days overdue)
  - Memory Consolidation: due 2026-05-12 (65 days overdue)
  - Roadmap Accuracy: due 2026-05-21 (56 days overdue)
  - Status Hygiene: due 2026-05-12 (65 days overdue)
  - Vulnerability Scan: due 2026-05-11 (66 days overdue)
- No routine-flagged issues from the 69-day gap are currently captured in the Backlog (since no routines ran). Flagging this gap for user attention — Plans 40 and 41 shipped significant new code during this period with no routine coverage.
- **Issues:** Major gap — 2 significant plans shipped without routine oversight. Especially concerning: no vulnerability scan since 2026-05-04 (73 days), no doc freshness check since 2026-05-04, no status hygiene since 2026-05-07.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any items warrant promotion. No items have explicit user approval for implementation. Flagging the PAT expiry (P1) as a candidate for immediate action if user approves.
- **Result:** No autonomous promotion. Presented to user for decision in "Items Flagged for User Review" below.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.

### Step 8: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` Backlog Hygiene row.
- **Result:** `Last Ran` → 2026-07-16, `Next Due` → 2026-07-23, `Status` → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | P0 `[bug] Sensor hook commands $PWD-walk-up` still in Backlog — resolved in v0.4.3 (2026-05-13) | `Backlog.md` P0 | Commented out with resolution note |
| 2 | high | P0 `[feature] non-interactive flags` still in Backlog — resolved in v0.4.2 (2026-05-13) | `Backlog.md` P0 | Commented out with resolution note |
| 3 | high | P1 `[feature] Full CLI parity` still in Backlog — resolved in Plan 41 (2026-06-16) | `Backlog.md` P1 | Commented out with resolution note |
| 4 | high | HOMEBREW_TAP_TOKEN PAT expiry was due 2026-07-15 — 1 day overdue, expires ~2026-07-21 (5 days) | `Backlog.md` P1 | Flagged for immediate user action |
| 5 | high | All 7 routines overdue — no runs since 2026-05-07 (69 days); Plans 40+41 shipped with no routine coverage | `routines.md` dashboard | Flagged for user to schedule catch-up routine runs |
| 6 | medium | P2 `[security] Website npm vuln tree` — 6 open Dependabot alerts incl. HIGH esbuild/vite, 30 days old | `Backlog.md` P2 | Flagged for user (vulnerability-scan routine should handle) |
| 7 | medium | Phase 2 Roadmap item `Template variables expansion` has no Backlog entry | `Roadmap.md` Phase 2 | Flagged for user review |
| 8 | medium | Plans 40 and 41 are in `Plans/Active/` but both shipped — should be in `Plans/Archive/` | `Plans/Active/` | Flagged for user (Status Hygiene routine's job) |
| 9 | low | P1 `[ops] Routine bot PR pile-up` structural fix still pending after 70 days | `Backlog.md` P1 | No change — still valid, flagged for re-prioritization |
| 10 | low | P2 `[bug] bonsai validate can't pass on Bonsai repo itself` — blocks Plan 40 dogfood | `Backlog.md` P2 | No change — flagged as blocking item for next Plan 40/validate work |
| 11 | low | Group B code-quality items (8 items) 90+ days old with no progress | `Backlog.md` Group B | Flagged for re-prioritization |
| 12 | low | `Plans/Archive` folder structure (Group E) unresolved 91+ days — plan files still accumulate in Active/ | `Backlog.md` Group E | No change — flagged |
| 13 | info | P0 section now empty of active items — placeholder note added | `Backlog.md` P0 | Placeholder line added |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[URGENT] HOMEBREW_TAP_TOKEN PAT rotation** — The reminder set in Backlog P1 was for 2026-07-15 (yesterday). The PAT was created 2026-04-22 with 90-day default expiry → expires approximately 2026-07-21 (5 days from now). Rotate now via GitHub → Settings → Developer settings → Fine-grained tokens, then `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`. Symptom of expiry: GoReleaser fails at brew step with 401.

2. **[RECOMMENDED] Run overdue routines** — All 7 routines are 56–66 days overdue. Priority order:
   - Status Hygiene (Plans 40/41 are Done but still in Plans/Active/, Status.md may have items to archive)
   - Vulnerability Scan (no scan in 73 days; website has 6 open Dependabot HIGH/MED alerts)
   - Doc Freshness Check (Plans 40/41 added new code paths, CLAUDE.md tree likely drifted)
   - Dependency Audit, Memory Consolidation, Roadmap Accuracy, then Backlog Hygiene again after those run

3. **[DECISION NEEDED] Roadmap Phase 2 gap** — `Template variables expansion` is a Phase 2 milestone with no Backlog entry. Add a Backlog item, or decide it's not needed.

4. **[REPRIORITIZATION] Group B (Code Quality)** — 8 testing-debt items are 90+ days old. Plan 41 added significant new headless code paths (`nonint/` package) with no tests. Consider picking up at least `cmd/` test coverage or `nonint/` unit tests this cycle.

5. **[REPRIORITIZATION] P1 `[ops] Routine bot PR pile-up`** — Still unresolved after 70 days. If the cloud routine is still generating PRs that accumulate without merge, this should be addressed soon.

## Notes for Next Run
- The P0 section is now empty; if no new P0 items arise before the next run it can stay with the placeholder.
- Both Plan 40 and Plan 41 are in `Plans/Active/` — Status Hygiene routine should move them to Archive.
- After the overdue routines run, the next Backlog Hygiene should clean up any new items those routines surface.
- PAT rotation calendar reminder: next rotation due ~2026-10-14 (90 days from this rotation).
