---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-01
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
- **Duration:** ~12 min
- **Files Read:** 10 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/CLAUDE.md`, `/home/user/Bonsai/station/CLAUDE.md` (via system-reminder), `/home/user/Bonsai/station/Logs/RoutineLog.md`, directory listings for internal/, cmd/, agent/
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls), Grep, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read station/INDEX.md, station/Playbook/Status.md, station/Playbook/Roadmap.md. Ran `git log --oneline --since="2026-05-04"` and `git log --name-only --since="2026-05-04"` to identify all commits since last run.
- **Result:** 47 commits since 2026-05-04. Significant shipped work: Plan 41 (headless CLI contract — new `internal/nonint/` package, `list --json`, headless remove/update cores; PRs #120–125); Plan 40 Phases 1–3 (frozen schemas, root-relative scaffolding, project-level validate pass, memory-routing docs, guide formats page; PRs #114–116); Plan 39 (`--non-interactive --from-config` flags for init/add; PR #102); Plan 38 (Bonsai-Eval bootstrap); v0.4.2, v0.4.3, v0.5.0 (untagged). `cmd/completion.go` added in PR #78 (2026-05-07, first contribution).
- **Issues:** Multiple new packages/commands not reflected in documentation — detailed in steps 2–3 below.

### Step 2: Check INDEX.md accuracy
- **Action:** Compared `station/INDEX.md` Tech Stack, Key Metrics, Architecture Overview, and Document Registry against current codebase state (file listing + git history).
- **Result:**
  - **Tech Stack table:** Accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, single binary all still correct.
  - **Key Metrics — CLI commands count:** STALE. Says "8 (init, add, remove, list, catalog, update, guide, validate)". Should be 9 — `bonsai completion [bash|zsh|fish|powershell]` was added in PR #78. (Note: completion is hidden from `--help` via `HiddenDefaultCmd = true` per Plan 28, but the command exists.)
  - **Key Metrics — Agent types:** 6 — still accurate.
  - **Key Metrics — Catalog items:** ~50 — reasonable approximation, no major drift.
  - **Architecture Overview:** STALE. Lists `internal/validate/`, `internal/wsvalidate/`, `internal/tui/` but is missing `internal/nonint/` — a new package added in Plan 41 (2026-06-16) containing the headless CLI contract (event types, result types, exit code constants, runner interface, headless cores for update and remove).
  - **Architecture cmd/ description:** STALE. Says "init, add, remove, list, catalog, update, guide, validate" — missing `completion`.
  - **Document Registry:** All referenced paths exist and descriptions are accurate.
- **Issues:** 2 stale entries (CLI count, architecture diagram).

### Step 3: Check navigation links
- **Action:** Verified all links in station/CLAUDE.md navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External refs). Ran `ls` on all agent/ subdirectories and key Playbook/ paths. Checked `.bonsai/catalog.json` and `.bonsai.yaml`.
- **Result:**
  - **agent/Core/:** identity.md, memory.md, self-awareness.md, routines.md — all exist. All links resolve. ✓
  - **agent/Protocols/:** memory.md, scope-boundaries.md, security.md, session-start.md — all exist. ✓
  - **agent/Workflows/:** All 9 linked files (code-review, planning, pr-review, security-audit, session-logging, test-plan, session-wrapup, issue-to-implementation, routine-digest) — all exist. ✓
  - **agent/Skills/:** All 6 linked files (planning-template, review-checklist, issue-classification, pr-creation, bubbletea, bonsai-model) — all exist. ✓
  - **agent/Routines/:** All 7 listed files — all exist. ✓
  - **agent/Sensors/:** All 10 listed sensor scripts — all exist. ✓
  - **External refs:** INDEX.md, Status.md, Roadmap.md, SecurityStandards.md, Backlog.md, KeyDecisionLog.md, Plans/Active/, Reports/Pending/ — all exist. ✓
  - **.bonsai/catalog.json** and **.bonsai.yaml** — both exist. ✓
  - **No broken links found.**
  - **Note:** `agent/Workflows/plan-grilling.md` and `agent/Skills/critic-agent-prompts.md` exist on disk but are not listed in the station/CLAUDE.md nav tables. May be intentional (support files for the grilling pipeline, not primary workflows).
- **Issues:** No broken links. 2 unlisted files (low priority, likely intentional).

### Step 4: Root CLAUDE.md Project Structure check
- **Action:** Read `/home/user/Bonsai/CLAUDE.md` Project Structure section. Cross-referenced against actual `cmd/`, `internal/`, and `catalog/scaffolding/` directory listings.
- **Result:**
  - **cmd/ listing:** Missing `completion.go` (added PR #78, 2026-05-07).
  - **internal/ listing:** Missing `nonint/` package (added Plan 41, 2026-06-16). This is a significant package (~10 files) for headless CLI contract.
  - **internal/validate/ listing:** Shows only `validate.go` + `validate_test.go`. Missing `project.go` + `project_test.go` (added Plan 40 Phase 2, 2026-06-13 — project-level validate pass).
  - **internal/generate/ listing:** Missing `list_snapshot.go` (added Plan 41 Phase 4, 2026-06-16 — `list --json` support). Also missing `catalog_snapshot_unix.go` and `catalog_snapshot_windows.go` (platform-split from v0.4.0 release, 2026-05-04).
  - **station/ Sensors comment:** Lists only 5 of 10 installed sensors ("context-guard, scope-guard-files, session-context, status-bar, routine-check") — missing agent-review, compact-recovery, dispatch-guard, subagent-stop-review, statusline. Note: station/CLAUDE.md (workspace nav) has the full accurate sensor table; this is only the root CLAUDE.md project overview comment.
  - **catalog/scaffolding/:** Not sub-listed in CLAUDE.md (shows only as `← project management infrastructure templates`) — acceptable, no change needed.
- **Issues:** 5 drift items in root CLAUDE.md.

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-07-01, Next Due → 2026-07-08, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package (~10 files, headless CLI contract + exit codes) not in Project Structure | `CLAUDE.md` (root), internal/ section | Flagged for user — propose adding `├── nonint/` block |
| 2 | MEDIUM | `internal/validate/project.go` + `project_test.go` not listed (Plan 40 Phase 2) | `CLAUDE.md` (root), validate/ section | Flagged for user |
| 3 | MEDIUM | `internal/generate/list_snapshot.go` not listed (Plan 41 Phase 4); also missing `catalog_snapshot_unix.go` + `catalog_snapshot_windows.go` (v0.4.0) | `CLAUDE.md` (root), generate/ section | Flagged for user |
| 4 | MEDIUM | `internal/nonint/` missing from architecture diagram | `station/INDEX.md`, Architecture Overview | Flagged for user |
| 5 | LOW | CLI commands count: 8, should be 9 (`completion` added PR #78) | `station/INDEX.md`, Key Metrics table | Flagged for user |
| 6 | LOW | `cmd/completion.go` not listed in cmd/ section | `CLAUDE.md` (root), cmd/ section | Flagged for user |
| 7 | LOW | Sensors listing in root CLAUDE.md station/ comment shows 5 of 10 sensors | `CLAUDE.md` (root), station/ overview | Flagged for user (station/CLAUDE.md nav table is accurate) |
| 8 | INFO | `agent/Workflows/plan-grilling.md` exists but not in station/CLAUDE.md Workflows nav table | `station/CLAUDE.md` | Flagged for user — likely intentional |
| 9 | INFO | `agent/Skills/critic-agent-prompts.md` exists but not in station/CLAUDE.md Skills nav table | `station/CLAUDE.md` | Flagged for user — likely intentional |

---

## Proposed Updates (for user decision — do not auto-apply)

### Finding 1 + 4 — Add `internal/nonint/` to both CLAUDE.md and INDEX.md

In `CLAUDE.md` (root), after the `validate/` block in internal/, add:
```
│   ├── nonint/
│   │   ├── nonint.go        ← package entry point
│   │   ├── events.go        ← JSONL event types (Init, Add, Update, Remove)
│   │   ├── result.go        ← Result struct + ExitCode constants (0/2/3/4/5)
│   │   ├── runner.go        ← headless runner interface
│   │   ├── update.go        ← headless update core
│   │   └── remove.go        ← headless remove core (--yes/--from)
```

In `station/INDEX.md` Architecture Overview, add after `internal/wsvalidate/`:
```
internal/nonint/      ← headless CLI cores (init/add/update/remove), JSONL event contract, exit codes (Plan 41)
```

Also update the cmd/ description line:
- Before: `cmd/ (Cobra)          ← CLI commands: init, add, remove, list, catalog, update, guide, validate`
- After: `cmd/ (Cobra)          ← CLI commands: init, add, remove, list, catalog, update, guide, validate, completion`

### Finding 2 — Add `project.go` to validate/ listing in CLAUDE.md

```
│   ├── validate/
│   │   ├── validate.go      ← read-only audit — orphans, stale lock entries, untracked customs, frontmatter (Plan 35)
│   │   ├── validate_test.go ← tests for validate package
│   │   ├── project.go       ← project-level validate pass (Plan 40)
│   │   └── project_test.go  ← tests for project validate
```

### Finding 3 — Add missing generate/ files to CLAUDE.md

Add `list_snapshot.go` and the platform-split catalog_snapshot files:
```
│   │   ├── list_snapshot.go          ← writes list --json snapshot
│   │   ├── catalog_snapshot_unix.go  ← platform: symlink safety (POSIX)
│   │   └── catalog_snapshot_windows.go ← platform: symlink safety (Windows)
```

### Finding 5 + 6 — CLI command count + completion.go

In `station/INDEX.md` Key Metrics: `9 (init, add, remove, list, catalog, update, guide, validate, completion)`

In `CLAUDE.md` (root) cmd/ section, add before `validate.go`:
```
│   ├── completion.go        ← bonsai completion [bash|zsh|fish|powershell]
```

### Finding 7 — Root CLAUDE.md station/Sensors comment (low priority)

Update the project structure comment from `← context-guard, scope-guard-files, session-context, status-bar, routine-check` to include all 10 sensors. Or just note `← (see station/CLAUDE.md for full list)`.

### Findings 8–9 — Unlisted workflow + skill files (user decision)

- `agent/Workflows/plan-grilling.md` — if this should be user-accessible, add row to station/CLAUDE.md Workflows table under "Starting adversarial plan review" or similar.
- `agent/Skills/critic-agent-prompts.md` — if this is a support file for plan-grilling, no nav entry needed; if it should be discoverable, add to Skills table.

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **Root CLAUDE.md drift is recurring** — this is the 5th+ consecutive cycle flagging root CLAUDE.md project structure staleness. The 2026-05-04 doc-freshness report flagged it as "high" and the 2026-05-07 Routine Digest captured it as a Backlog P2 "root-CLAUDE.md check sub-step". Consider whether a quick doc-refresh plan (similar to Plan 37) would close the backlog debt and reduce noise in future routine runs.

2. **`plan-grilling.md` and `critic-agent-prompts.md`** not in nav tables — user should decide if these should be listed or intentionally undiscovered.

---

## Notes for Next Run

- All navigation links in station/CLAUDE.md are clean — no broken refs.
- Root CLAUDE.md project structure continues to drift faster than doc routines catch it. A pre-ship doc-refresh checklist (embedded in issue-to-implementation workflow) might prevent accumulation.
- INDEX.md architecture section is accurate except for the missing `internal/nonint/` block.
- `bonsai completion` being hidden from `--help` (via `HiddenDefaultCmd = true`) might be intentional — user should decide whether the CLI count in INDEX.md should reflect hidden commands.
- Next run due 2026-07-08.
