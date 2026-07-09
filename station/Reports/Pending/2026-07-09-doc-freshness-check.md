---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-09
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
- **Duration:** ~8 min
- **Files Read:** 8 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/CLAUDE.md` (via system-reminder), `station/code-index.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `station/Reports/Pending/2026-07-09-doc-freshness-check.md` (created), `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log append)
- **Tools Used:** `git log --oneline --since="7 days ago"`, `git log --oneline --since="7 days ago" --name-only`, `git log --oneline -20`, `git log --oneline -5 --name-status`, `ls /home/user/Bonsai/internal/`, `ls /home/user/Bonsai/cmd/`, `ls /home/user/Bonsai/catalog/agents/`, existence checks via bash for all CLAUDE.md navigation link targets
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Ran `git log --oneline --since="7 days ago"` to identify last 7 days of commits. Also reviewed broader recent history (`git log --oneline -20`) to understand the full landscape since last run (2026-05-04).
- **Result:** Only 1 commit in last 7 days: `0bb3cca chore(routine): run Backlog Hygiene — 10 findings, 3 items resolved` (2026-07-09). That commit only modified `station/` files — no code changes. However, comparing current codebase state against documented state reveals drift from older commits (Plan 41 shipped 2026-06-16, `completion` command added 2026-05-07) that was not captured in prior doc-freshness runs or the 2026-05-04 routine digest.
- **Issues:** 4 drift items identified — see Findings Summary.

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` in full. Compared Tech Stack table, CLI command count, agent type count, catalog item count, and Architecture overview against actual codebase state.
- **Result:**
  - **Tech Stack table** — accurate. Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS all still correct.
  - **Agent types count** — accurate. `ls catalog/agents/` confirms 6 types (tech-lead, fullstack, backend, frontend, devops, security).
  - **Catalog items count** — accurate (approximately). Current counts: 18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines = 53 catalog items. INDEX.md says "~50" — still reasonable.
  - **CLI commands count** — **STALE.** INDEX.md says "8 (init, add, remove, list, catalog, update, guide, validate)". The `completion` command (first external contribution, @mvanhorn, merged 2026-05-07 via PR #78) brings the count to 9.
  - **Architecture overview** — **STALE.** The architecture diagram lists `internal/catalog/`, `internal/config/`, `internal/generate/`, `internal/validate/`, `internal/wsvalidate/`, `internal/tui/` — but omits `internal/nonint/` which was added in Plan 41 (headless CLI contract, PR #120/#122/#123/#121/#125, merged 2026-06-16). This package contains 14 files implementing the headless result/event shapes, runner, and remove/update cores.
- **Issues:** 2 stale items (CLI count, arch diagram).

### Step 3: Check navigation links
- **Action:** Extracted all link targets from `station/CLAUDE.md` navigation tables and verified existence of each file/directory using bash existence checks.
- **Result:** All 44 checked targets resolve correctly:
  - Core files (4/4): identity.md, memory.md, self-awareness.md, routines.md — all present
  - Protocols (4/4): memory.md, scope-boundaries.md, security.md, session-start.md — all present
  - Workflows (9/9): all present including routine-digest.md
  - Skills (6/6): including bonsai-model.md (previously flagged as broken — now confirmed present)
  - Routines (7/7): all present
  - Sensors (10/10): all present including newer sensors (agent-review.sh, dispatch-guard.sh, subagent-stop-review.sh, compact-recovery.sh, statusline.sh)
  - External references: INDEX.md, CLAUDE.md, Playbook/Status.md, Playbook/Roadmap.md, Playbook/Standards/SecurityStandards.md, Playbook/Backlog.md, Logs/KeyDecisionLog.md, Reports/report-template.md, code-index.md, Playbook/Plans/Active/, Reports/Pending/, ../.bonsai/catalog.json, ../.bonsai.yaml — all present
- **Issues:** None. Navigation is completely clean.

### Step 4: Check code-index.md accuracy
- **Action:** Read `station/code-index.md` in full. Compared the CLI Commands table and internal package sections against actual codebase state (`ls /home/user/Bonsai/cmd/`, `ls /home/user/Bonsai/internal/`).
- **Result:**
  - **CLI Commands table** — **STALE.** Lists 8 commands; `completion` command is absent. `cmd/completion.go` exists and registers `bonsai completion [bash|zsh|fish|powershell]`.
  - **`internal/nonint/` section** — **MISSING.** The code-index has no section for the `internal/nonint/` package. This package (added Plan 41) contains: `config.go`, `events.go`, `nonint.go`, `remove.go`, `result.go`, `runner.go`, `update.go` + test files — it is the headless CLI contract core and is referenced in `cmd/add.go`, `cmd/remove.go`, `cmd/update.go`, `cmd/list.go`, `cmd/init.go`.
  - **Existing sections reviewed** — Catalog, Config, Generator, Validate, wsvalidate, TUI sections appear broadly accurate at the function level (line numbers may have drifted but no obvious stale entries found without per-function verification).
- **Issues:** 2 stale/missing items.

### Step 5: Report findings and update dashboard
- **Action:** Compiled findings into this report. Dashboard update and log append follow.
- **Result:** 4 findings total, all flagged for user decision. No doc edits executed per procedure ("propose updates but don't execute — flag for user decision").
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | CLI command count stale: says "8" but should be "9" — `completion` command added May 2026 (PR #78, @mvanhorn) | `station/INDEX.md` Key Metrics table, row "CLI commands" | Flagged for user |
| 2 | Medium | Architecture overview missing `internal/nonint/` package (headless CLI contract, Plan 41 June 2026) | `station/INDEX.md` Architecture Overview section | Flagged for user |
| 3 | Low | CLI Commands table missing `completion` command entry (`cmd/completion.go`) | `station/code-index.md` CLI Commands table | Flagged for user |
| 4 | Medium | No section for `internal/nonint/` package — major Plan 41 addition with 7 source files undocumented | `station/code-index.md` | Flagged for user |

---

## Errors & Warnings

No errors encountered.

**Carry-forward note:** The 2026-05-07 Backlog Hygiene flagged `code-index.md` staleness and INDEX.md arch diagram drift as "uncaptured" from the 2026-05-04 run. Those items are now formally captured here. The previously flagged broken link `agent/Skills/bonsai-model.md` is confirmed resolved — file exists.

---

## Items Flagged for User Review

- **[MEDIUM] `station/INDEX.md` CLI command count**: Update "8 (init, add, remove, list, catalog, update, guide, validate)" → "9 (init, add, remove, list, catalog, update, guide, validate, completion)". Quick fix — 1 line.

- **[MEDIUM] `station/INDEX.md` Architecture overview**: Add `internal/nonint/` to the diagram. Suggested line (after `internal/wsvalidate/`): `internal/nonint/      ← headless CLI contract — Result/Event shapes, runner, remove/update cores`. Quick fix — 1 line.

- **[LOW] `station/code-index.md` CLI Commands table**: Add a row for `bonsai completion` → `cmd/completion.go` → `completionCmd` — Generate shell completion script (bash/zsh/fish/powershell). Quick fix — 1 row.

- **[MEDIUM] `station/code-index.md` `internal/nonint/` section**: Add a new section documenting the package. Key types/functions: `Result` shape, `Event` shape (JSONL line), `Runner`, headless `Remove()`, headless `Update()`. This is large enough to warrant a dispatch to a code agent (one-shot doc write from source scan) or can be handled in a Plan 37-style doc-refresh bundle.

**Suggested approach:** Items 1–3 are quick fixes suitable for a Routine Digest inline apply. Item 4 (nonint section) is a small standalone doc task — bundle into next Plan or a mini doc-refresh dispatch.

---

## Notes for Next Run

- Navigation links are fully clean — skip exhaustive link checks next run unless CLAUDE.md has been regenerated.
- The `internal/nonint/` package is now the main undocumented code surface — once documented, the code-index will be current.
- Root `Bonsai/CLAUDE.md` project structure tree (not `station/CLAUDE.md`) likely also lacks `internal/nonint/` and `cmd/completion.go` entries, but that file is out of scope for this routine (station/ only). It has been a recurring drift target across prior runs — consider adding a root-CLAUDE.md sub-step to this routine.
- Catalog item count (~50 → now 53) is still within the "~50" approximation. No update needed unless count grows past 60.
