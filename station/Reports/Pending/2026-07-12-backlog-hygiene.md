---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-12
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
- **Duration:** ~10 min
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s
Read Backlog.md P0 section and checked each item against Status.md (In Progress or Pending).

**P0 item 1:** `[bug] Sensor hook commands use $PWD-walk-up` — NOT in Status.md active tables. However, Status.md "Recently Done" confirms v0.4.3 shipped the fix (2026-05-13): "sensor hook commands now bake install-time absolute paths in `.claude/settings.json`". The backlog item itself says "Ships v0.4.3." This item is resolved — not a misplaced P0 needing escalation, but a stale closed item. Removed (Step 2 action).

**P0 item 2:** `[feature] bonsai init/add need non-interactive flags` — NOT in Status.md active tables. Status.md "Recently Done" shows v0.4.2 shipped `--non-interactive --from-config` for init + add (2026-05-13), and Plan 41 (2026-06-16) extended full headless contract to all 4 commands. This item is fully resolved. Removed (Step 2 action).

**Sentrux P0 (HTML comment):** Already tracked as Status.md Pending (Blocked By: Rust toolchain). No action needed.

### Step 2 — Cross-reference with Status.md
Read Status.md In Progress (empty) and Recently Done.

**Items resolved and removed from Backlog:**
1. P0 `[bug] Sensor hook $PWD-walk-up` → resolved v0.4.3. Converted to HTML comment.
2. P0 `[feature] bonsai init/add non-interactive flags` → resolved v0.4.2 + Plan 41. Converted to HTML comment.
3. P1 `[feature] Full agent-drivable CLI parity: init / update / add / remove` → resolved by Plan 41 (2026-06-16). Item text stated the gap ("update has NO flags; remove likely TUI-only"); Plan 41 shipped headless `*Result` cores for all 4 commands + JSONL/exit contract. Converted to HTML comment.

**Blocked By check:** Status.md Pending has one item — "Trial sentrux on Bonsai repo" blocked by Rust toolchain (cargo/rustc not installed). No backlog item would unblock this; it requires user action to install Rust.

### Step 3 — Cross-reference with Roadmap.md
Read Roadmap.md.

Phase 1 is fully checked. Phase 2 items (self-update mechanism, template variables expansion, micro-task fast path) are in Backlog P3 at appropriate priority — no promotions needed since no phase boundary crossing is imminent.

No backlog items reference deprecated approaches or completed phases that don't already account for their status.

P2/P3 backlog items aligned with Phase 2+ milestones are correctly prioritized. No promotions recommended autonomously.

### Step 4 — Flag stale items

**URGENT flag (user action required within 3 days):**
- P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — Item was added 2026-04-22 with a calendar reminder for ~**2026-07-15**. Today is 2026-07-12. **The PAT rotation deadline is 3 days away.** If the PAT expires before rotation, the next release's GoReleaser brew step will fail with `401 Bad credentials`. See "Items Flagged for User Review" section.

**Long-stale items (30+ days at same priority, no progress — noted, no autonomous action):**
- P1 `[debt] Testing infrastructure for triggers and sensors` — stale since 2026-04-16 (87 days). Valid, no context change.
- P1 `[debt] Stale agent worktrees + branches` — stale since 2026-04-21 (82 days). Valid, recurring pattern.
- P1 `[ops] Routine bot PR pile-up` — stale since 2026-05-07 (66 days). Root cause (cloud routine PR behavior) not yet fixed.
- Group A `[bookkeeping] Retroactively trim Backlog entries to NoteStandards` — stale since 2026-04-25 (78 days). Still valid.
- Group B items — mostly stale since 2026-04-16/2026-04-24 (range 78–87 days). All still valid, no context change.
- Groups C, D, E, F items — similarly stale. All remain valid.

**Near-duplicates reviewed:** The now-removed P0 non-interactive flags item and the removed P1 full CLI parity item were effectively duplicates (P0 was a specific blocker, P1 was the full-scope version). Both resolved by the same work. Removing both is correct.

### Step 5 — Check for routine-generated items
Read RoutineLog.md for entries since last backlog-hygiene run (2026-05-07).

Since 2026-05-07, no routine reports ran (all routines are significantly overdue — last runs were 2026-05-04/2026-05-07, ~65 days ago). The only log entries after 2026-05-07 are plan dispatch logs (Plan 40 on 2026-06-13, Plan 41 on 2026-06-16).

Plan 40 and Plan 41 dispatch sessions did file new Backlog items (confirmed present):
- P2: Harden scaffolding writes against symlink substitution (2026-06-13)
- P2: bonsai validate warn on project.yaml drift (2026-06-13)
- P2: Plan 40 review nits (2026-06-13)
- P2: bonsai validate can't pass on Bonsai repo (2026-06-13)
- P2: Website npm vuln tree — astro upgrade fails build (2026-06-16)
- P2: Unify remove business logic (2026-06-16)

All plan-session-generated items are present in the backlog. No uncaptured findings.

**Note on overdue routines:** All 7 routines are significantly overdue (65+ days since last run). This is outside the scope of backlog hygiene to resolve, but worth flagging for the user to schedule a routine digest session.

### Step 6 — Promote ready items
No items have explicit user approval for promotion. The sentrux research item is already in Status.md Pending (blocked). No autonomous promotions.

### Steps 7–8 — Log + Dashboard
Appended to RoutineLog.md and updated routines.md dashboard.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 `[bug] Sensor hook $PWD-walk-up` — fixed in v0.4.3, stale backlog entry | Backlog.md P0 | Converted to HTML comment |
| 2 | resolved | P0 `[feature] bonsai init/add non-interactive flags` — fixed in v0.4.2 + Plan 41, stale backlog entry | Backlog.md P0 | Converted to HTML comment |
| 3 | resolved | P1 `[feature] Full agent-drivable CLI parity` — shipped by Plan 41 (2026-06-16) | Backlog.md P1 | Converted to HTML comment |
| 4 | URGENT | P1 `[ops] HOMEBREW_TAP_TOKEN PAT` — expiry reminder set for 2026-07-15, 3 days from today | Backlog.md P1 | Flagged for user (no text change needed — deadline already in item) |
| 5 | info | All 7 routines overdue 65+ days — no routine digest since 2026-05-07 | routines.md dashboard | Flagged for user |
| 6 | info | Many P1/P2/P3 items 30–87 days stale without progress | Backlog.md P1-P3 | No action; normal backlog state |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

### 1. HOMEBREW_TAP_TOKEN PAT Rotation — Due 2026-07-15 (3 days)

The `HOMEBREW_TAP_TOKEN` fine-grained PAT was rotated on 2026-04-22 with a 90-day expiry. The calendar reminder in the backlog targets ~2026-07-15. **Today is 2026-07-12 — rotation is needed in the next 3 days** to avoid breaking the GoReleaser brew step on the next release.

Steps: Go to GitHub → Settings → Developer Settings → Personal access tokens, rotate the PAT, then run `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`.

### 2. All Routines Significantly Overdue

The routines dashboard shows all 7 routines last ran 2026-05-04/2026-05-07 (~65 days ago). Suggested next session: run a full routine digest to process overdue routines. Priority order: Vulnerability Scan, Dependency Audit, Doc Freshness Check (all 65+ days overdue), then Status Hygiene, Memory Consolidation, Roadmap Accuracy.

## Notes for Next Run

- P0 section is now empty (both items resolved and commented out). If the sentrux research item is resolved via Status.md before next run, remove its HTML comment from the P0 area.
- Three newly removed items (2 P0, 1 P1) were all resolved by the same workstream (v0.4.2 → Plan 41 headless CLI contract). Good coverage.
- Consider scheduling a routine digest soon — all routines are 65+ days overdue. A vulnerability-scan + dependency-audit cycle is especially timely given the npm vuln item in P2.
- Website npm vuln (P2, added 2026-06-16) remains open — astro upgrade failing build needs a dedicated session.
