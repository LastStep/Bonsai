---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-27
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch
- **Gap Since Last Run:** 81 days (significantly overdue — all routines appear last-ran 2026-05-04/05-07)

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 5
  - `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`
  - `/home/user/Bonsai/station/Playbook/Backlog.md`
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 3
  - `/home/user/Bonsai/station/Playbook/Backlog.md` (removed 2 resolved P0 items, added 2 inline flags)
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (new entry appended)
- **Tools Used:** Read (5×), Edit (3×), Write (2×)
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s

Scanned P0 section of Backlog.md. Found **2 P0 items**; cross-referenced each against Status.md.

| P0 Item | In Status.md? | Verdict |
|---------|--------------|---------|
| [bug] Sensor hook commands use `$PWD`-walk-up | Yes — v0.4.3 hotfix shipped 2026-05-13 (PRs #105/#106) | RESOLVED — removed |
| [feature] `bonsai init`/`add` need non-interactive flags | Yes — v0.4.2 shipped 2026-05-13 (PR #102) | RESOLVED — removed |

The `[research] Trial sentrux` item was already commented out (promoted to Status.md Pending 2026-05-07). No misplaced active P0s remain.

### Step 2 — Cross-reference with Status.md

Read Status.md. Key activity since last backlog-hygiene run (2026-05-07):
- **v0.4.2 shipped 2026-05-13** — `--non-interactive`/`--from-config` for init+add
- **v0.4.3 shipped 2026-05-13** — sensor hook absolute-path baking
- **Plan 38 handoff 2026-05-13** — Bonsai-Eval bootstrapped
- **Plan 40 shipped 2026-06-13** — frozen v1 schemas, validate pass, memory routing
- **Plan 41 shipped 2026-06-16** — headless CLI contract for all 4 mutating commands

**Removals applied:**
- P0 sensor hook bug → commented out with resolution note
- P0 non-interactive flags feature → commented out with resolution note

**Additional findings (no removal without user confirmation):**
- P1 "Full agent-drivable CLI parity" (added 2026-06-13) — Plan 41 appears to have shipped exactly this (headless cores + JSONL/exit contract for all 4 cmds + `list --json` + contract doc). Flagged inline for user review.
- No P1/P2/P3 items currently In Progress in Status.md.
- Pending item (sentrux trial) — still blocked on Rust toolchain, no change needed.

**Blocked-by check:** No Status.md Pending items with "Blocked By" that could be unblocked by a Backlog item resolution.

### Step 3 — Cross-reference with Roadmap.md

Phase 1 is **fully complete** (all checkboxes checked, including `bonsai validate` added by 2026-05-07 routine-digest).

Phase 2 milestones and Backlog alignment:
| Phase 2 Milestone | Backlog Entry | Current Priority |
|-------------------|---------------|-----------------|
| Self-update mechanism | "[improvement] Self-update mechanism" | P3 (Big Bets / Future Platform) |
| Template variables expansion | Not in Backlog explicitly | — |
| Micro-task fast path | "[improvement] Micro-task fast path" | P3 (Future Platform) |
| Custom item detection | Already checked (shipped) | — |

**Observation:** Phase 1 is done. Phase 2 milestones include Self-update mechanism and Micro-task fast path which are currently P3. Now that Phase 1 is complete, these could warrant promotion to P2. Flagging for user review — not auto-promoting per routine scope.

**Deprecated references check:**
- No Backlog items reference completed plan phases or deprecated approaches with stale framing, except the P1 CLI parity item (which Plan 41 resolved — already flagged).
- P2 "[bug] `bonsai validate` can't pass on the Bonsai repo itself" references `.bonsai-lock.yaml` being gitignored — still accurate, no change.

### Step 4 — Flag stale items

**Time reference:** Last run 2026-05-07; today 2026-07-27 (81 days gap). All items added before 2026-06-27 and untouched are 30+ days without progress.

**Critical flag — HOMEBREW_TAP_TOKEN likely expired:**
The P1 ops item notes PAT was rotated 2026-04-22 with a 90-day default expiry. 90 days from 2026-04-22 = 2026-07-21. Today is 2026-07-27. **The PAT is 6 days past its likely expiry.** The next release will fail at the Homebrew formula update step. Annotated inline as `[OVERDUE — rotate now]`.

**Stale items (30+ days, no progress noted):**
- Group A: "[bookkeeping] Retroactively trim Backlog entries to NoteStandards" — added 2026-04-25. 93 days old with no action.
- Group B items (all from 2026-04-16): break up generate.go, catalog test coverage, CLI test coverage, PTY smoke test, Plan-29-test-gap, Plan-29-security-hardening, Plan-31-cosmetic, Plan-31-security-hardening. All 8 items 102 days old without noted progress.
- Group C: "OSS polish — demo GIF/asciinema" — added 2026-04-16, 102 days. User-recording item (agent can't do this), but worth reviewing if still wanted.
- Group C: "Ability-name argument completion" — added 2026-05-07, 81 days.
- Group D: "Revisit concept-decisions research", "Unbuilt catalog items", "Changelog generation skill", "Research scaffolding item" — all 2026-04-16 (102 days).
- Group E: "Plan archiving", "Plans Index file", "Consolidate FieldNotes", "Post-update backup merge hint", "Port statusLine to catalog sensor" — 2026-04-16 to 2026-04-22 (97–102 days).
- Group F: Both remaining items (docs items) — 2026-04-20 (98 days).

**Near-duplicates:**
- Group E "Plan archiving — Active/Archive folder structure" and "Plans Index file" are tightly coupled — the Plans Index naturally follows from an Active/Archive split. Consider merging into a single item.
- P3 "[improvement] Self-update mechanism" (Big Bets section) and Phase 2 Roadmap "Self-update mechanism" checkbox are the same concept — consistent, no action needed, but worth noting the phase boundary has now passed.

### Step 5 — Check routine-generated items

Reviewed RoutineLog.md for entries since 2026-05-07 backlog-hygiene run. The only RoutineLog entries after 2026-05-07 are from the same day's sweep (Memory Consolidation, Roadmap Accuracy, Status Hygiene) — all before this run's last-ran date. The 2026-06-13 entry is a Plan dispatch log, not a routine.

**No routine-generated findings in the 2026-05-07 → 2026-07-27 window are uncaptured.** The following routine flags from 2026-05-07 are verified as addressed:
- Roadmap "Better trigger sections" checkbox — addressed by 2026-05-07 Routine Digest (checked `[x]` with annotation)
- code-index.md staleness — addressed by Plan 37 (doc-refresh-bundle, 2026-05-07)
- Broken nav link `agent/Skills/bonsai-model.md` — present in CLAUDE.md Skills table, file reference exists

### Steps 6–8 — Promote / Log / Update dashboard

No items are approved for immediate implementation this run (audit-only). HOMEBREW_TAP_TOKEN expiry and P1 CLI parity resolution require user decision before any promotion. See Items Flagged for User Review below.

Dashboard and log updated per post-procedure instructions.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|-------------|
| 1 | resolved | P0 sensor hook `$PWD`-walk-up bug — shipped v0.4.3 2026-05-13 | Backlog P0 | Removed item, HTML comment with resolution note |
| 2 | resolved | P0 `bonsai init`/`add` non-interactive flags — shipped v0.4.2 2026-05-13 | Backlog P0 | Removed item, HTML comment with resolution note |
| 3 | high | HOMEBREW_TAP_TOKEN PAT expired (~2026-07-21, 6 days ago) — next release brew step will fail | Backlog P1 | Added `[OVERDUE — rotate now]` inline annotation. User must rotate PAT immediately. |
| 4 | medium | P1 "Full agent-drivable CLI parity" likely resolved by Plan 41 (2026-06-16) | Backlog P1 | Added `[REVIEW: likely resolved by Plan 41]` inline annotation. User to confirm and remove. |
| 5 | low | 15+ stale Group B/C/D/E/F items (80–102 days old, no progress) | Backlog P2/P3 | Flagged in report. No inline edits — user to re-prioritize at next roadmap review. |
| 6 | low | Phase 2 milestones (Self-update, Micro-task fast path) are in Backlog as P3 — Phase 1 now done | Backlog P3 | Flagged for user review: promote to P2 or leave as P3. |
| 7 | info | Near-duplicate: Group E "Plan archiving" + "Plans Index file" closely related | Backlog P2 | Noted. Consider merging. |

---

## Errors & Warnings

None.

---

## Items Flagged for User Review

1. **[HIGH — ACTION NEEDED] HOMEBREW_TAP_TOKEN PAT has likely expired (2026-07-21).** Next release will fail at the Homebrew formula update step with `401 Bad credentials`. Rotate the PAT now: GitHub → Developer Settings → Fine-grained tokens → generate new → `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`. Set a new calendar reminder for 90 days out. Also audit any other repo secrets for expiry.

2. **[MEDIUM] P1 "Full agent-drivable CLI parity" — verify resolved by Plan 41.** Plan 41 shipped headless cores (`*Result` types), JSONL/exit contract (`ExitConflict=5`), `list --json`, and `docs/agent-interface.md` for all 4 mutating commands. If this fully satisfies the original requirement, remove the Backlog entry. If Rung-2.5 multi-agent `--from-config` for `update` and `remove` still needs validation against the Bonsai-Eval eval harness, keep and re-scope.

3. **[LOW] Phase 2 milestones now relevant.** Phase 1 is complete. P3 items "[improvement] Self-update mechanism" and "[improvement] Micro-task fast path" map directly to Phase 2 roadmap checkboxes. Consider promoting to P2 if Phase 2 is now the active phase.

4. **[LOW] ~15 stale Group B/C/D/E/F items.** These have been in the backlog 80–102 days with no noted progress. Recommend a brief roadmap-alignment sweep: either assign to a plan, demote to P3, or archive with explanation. Suggests the next routine-digest should include a focused Backlog re-prioritization session.

---

## Notes for Next Run

- HOMEBREW_TAP_TOKEN rotation should be documented with a new expiry date once rotated — update the Backlog P1 entry with the new rotation date.
- If P1 "Full agent-drivable CLI parity" is confirmed resolved and removed, the P0 section will be empty and P1 will be lighter — a good sign.
- Group B (Code Quality & Testing) has 8 items 100+ days old — this group has been continuously deferred. Consider either scheduling a dedicated testing sprint or formally downgrading these to P3 so the backlog accurately reflects real priorities.
- This routine ran after an 81-day gap (all routines appear stalled since 2026-05-07). The 2026-06-13 and 2026-06-16 plan sessions produced Backlog items (Plans 40/41 reviews) but no routine runs were triggered. If the loop.md dispatch is still active, verify the cron schedule is firing correctly.
