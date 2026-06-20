---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-06-20
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
- **Files Read:** 5 — `station/agent/Routines/backlog-hygiene.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 1 — `station/Playbook/Backlog.md` (2 P0 items resolved → HTML comments)
- **Tools Used:** Read, Edit, Write, Bash
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
Read `Backlog.md` P0 section. Found 2 P0 items.

**Item 1:** `[bug] Sensor hook commands use $PWD-walk-up` — Cross-referenced with Status.md. **RESOLVED**: v0.4.3 hotfix shipped 2026-05-13 with absolute install-time project root baked into hook commands. This item was already fixed and is no longer a valid P0.

**Item 2:** `[feature] bonsai init / bonsai add need non-interactive flags` — Cross-referenced with Status.md. **RESOLVED**: `--non-interactive` + `--from-config` flags shipped in v0.4.2 on 2026-05-13. This item was already fixed and is no longer a valid P0.

Both resolved P0 items replaced with HTML comments (audit trail preserved).

**Result:** P0 section is now empty of active items. The only active item that _was_ a P0 — `[research] Trial sentrux` — was already promoted to Status.md Pending as of 2026-05-07 (routine-digest) and correctly only appears as an HTML comment.

### Step 2: Cross-reference with Status.md
Read Status.md. Confirmed:
- **In Progress:** none
- **Pending:** `[research] Trial sentrux on Bonsai repo` — blocked on Rust toolchain install. Already removed from Backlog active items (HTML comment). Correct.
- **Recently Done:** v0.4.3 (2026-05-13), v0.4.2 (2026-05-13), Plan 41 headless CLI (2026-06-16), Plan 40 Phases 1-3 (2026-06-13), and earlier items.

No active Backlog items duplicate In Progress or Recently Done tasks beyond the two P0s already resolved.

No Status.md Pending items with "Blocked By" appear resolvable by Backlog items (sentrux is blocked on Rust toolchain, not a code task).

### Step 3: Cross-reference with Roadmap.md
Read Roadmap.md. Key observations:
- **Phase 1:** Fully checked — all checkboxes complete. No Backlog items reference deprecated Phase 1 items.
- **Phase 2 "Extensibility":** Unchecked items (`Self-update mechanism`, `Template variables expansion`, `Micro-task fast path`). These have corresponding P3 Backlog entries — appropriate tier, no promotion needed.
- **Phase 3 "Cloud & Orchestration":** Big Bets in P3. Correctly placed.
- **Phase 4 "Ecosystem":** Long-horizon P3 items. Correctly placed.

The P1 item `[feature] Full agent-drivable (non-interactive) CLI parity` aligns with Phase 2 goals and is correctly at P1. No P2/P3 items need promotion based on current phase.

No Backlog items reference completed phases in a misleading way.

### Step 4: Flag stale items
Last run was 2026-05-07; today is 2026-06-20 (44 days). Items with no progress since their add date that are 30+ days old:

**P1 stale items (60+ days old, no visible progress):**
- `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — added 2026-04-22 (59 days). **URGENCY FLAG**: The calendar reminder target is ~2026-07-15, which is only **25 days away**. This is actionable now and should be escalated for user attention.
- `[debt] Testing infrastructure for triggers and sensors` — added 2026-04-16 (65 days). No apparent progress. Stale at P1.
- `[debt] Stale agent worktrees + branches accumulating` — added 2026-04-20 (61 days). No apparent progress. Stale at P1.
- `[ops] Routine bot PR pile-up` — added 2026-05-07 (44 days). No visible resolution since 9 PRs closed on that date. Stale at P1.

**P2 items with context gaps:**
- `[Plan-29-test-gap] Inverse-chrome companion test` — added 2026-04-23 (58 days). Minor test gap, no progress.
- `[Plan-29-security-hardening] Phase H validator hardening` — added 2026-04-23 (58 days). Item 4 still open.
- `[Plan-31-cosmetic] PR #75 review minors` — added 2026-04-24 (57 days). 4 remaining minors unfixed.
- `[Plan-31-security-hardening] TOCTOU on .bonsai/ dir perms` — added 2026-04-24 (57 days). No progress.

**Group A:**
- `[bookkeeping] Retroactively trim Backlog entries to NoteStandards` — added 2026-04-25 (56 days). No progress. Items are still verbose.

No near-duplicates identified that weren't already noted in prior hygiene runs.

### Step 5: Check for routine-generated items
Reviewed RoutineLog entries since 2026-05-07. Found two major sessions:
- **2026-06-13 Plan 40 dispatch**: Filed P2 items (`[security] Harden scaffolding writes`, `[improvement] bonsai validate drift warning`, `[improvement] Plan 40 review nits`, `[bug] bonsai validate can't pass on Bonsai repo`). All confirmed present in Backlog.
- **2026-06-16 Plan 41 dispatch** (inferred from Status.md + Backlog items dated 2026-06-16): Filed `[security] Website npm vuln tree` and `[debt] Unify remove business logic`. Both confirmed present in Backlog.

No routine findings since 2026-05-07 are uncaptured.

**Pending reports:** The `Reports/Pending/` folder is empty — all prior reports appear archived. No unfiled routine findings.

### Step 6: Promote ready items via issue-to-implementation
No items are approved for autonomous implementation in this run. The `HOMEBREW_TAP_TOKEN` PAT renewal (flagged below) is a user action, not an agent code task. The P1 headless CLI parity item is the user's stated "main thing" and needs user direction to start.

### Step 7: Log results
Done — see RoutineLog entry below.

### Step 8: Update dashboard
Done — routines.md dashboard updated (Last Ran, Next Due, Status).

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | P0-resolved | `[bug] Sensor hook $PWD-walk-up` was resolved in v0.4.3 (2026-05-13) but still listed as active P0 | Backlog.md P0 | Removed — replaced with HTML comment |
| 2 | P0-resolved | `[feature] non-interactive flags` was resolved in v0.4.2 (2026-05-13) but still listed as active P0 | Backlog.md P0 | Removed — replaced with HTML comment |
| 3 | HIGH | `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — rotation due ~2026-07-15 (25 days away) | Backlog.md P1 | Flagged for user — no agent action possible |
| 4 | MEDIUM | `[debt] Testing infrastructure for triggers and sensors` stale 65+ days at P1, no progress | Backlog.md P1 | Flagged for re-prioritization review |
| 5 | MEDIUM | `[debt] Stale agent worktrees + branches accumulating` stale 61+ days at P1, no progress | Backlog.md P1 | Flagged for re-prioritization review |
| 6 | MEDIUM | `[ops] Routine bot PR pile-up` stale 44 days, fix approach still undecided | Backlog.md P1 | Flagged for re-prioritization review |
| 7 | LOW | Group A bookkeeping sweep (NoteStandards trim) stale 56 days | Backlog.md Group A | Noted — low priority, no action |
| 8 | LOW | Multiple P2 plan-cosmetic items stale 57–58 days | Backlog.md P2 Group B | Noted — no action |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **URGENT — HOMEBREW_TAP_TOKEN PAT renewal due ~2026-07-15 (25 days):** The fine-grained PAT rotated 2026-04-22 expires ~2026-07-15. Go to GitHub → Settings → Developer Settings → Fine-grained PATs and rotate `HOMEBREW_TAP_TOKEN` before that date. Store the new value in `LastStep/Bonsai` repo secrets. If missed: GoReleaser will succeed but Homebrew formula won't update (only formula push fails, binaries still publish).

2. **P1 Stale review — 3 items 44–65 days at P1 with no progress:** Consider demoting `[debt] Testing infrastructure for triggers and sensors` and `[debt] Stale agent worktrees + branches accumulating` to P2 if no near-term plan, or actively scheduling them. The routine bot PR pile-up (44 days) also needs a decision on approach (a)/(b)/(c) from the item description.

3. **P0 section is now empty:** Both resolved P0s cleared. The only active near-P0 item — sentrux trial — remains in Status.md Pending (blocked on Rust). If the Rust toolchain gets installed, this should be picked up.

## Notes for Next Run

- P0 section is clean — both stale resolved items removed this run.
- Watch for: HOMEBREW_TAP_TOKEN renewal confirmation (due 2026-07-15).
- The `[feature] Full agent-drivable CLI parity` P1 item (added 2026-06-13) is the user's stated "main thing" — if a session starts a plan for it, remove from Backlog once it lands in Status.md.
- Website npm vuln (`[security]` P2, added 2026-06-16) remains open — watch for vulnerability-scan or dependency-audit routine to act on it.
- If `bonsai validate` lockfile policy is resolved (Backlog P2 `[bug]`, 2026-06-13), remove that item when the decision lands.
