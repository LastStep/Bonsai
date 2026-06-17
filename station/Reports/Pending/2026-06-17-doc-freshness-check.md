---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-17
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
- **Files Read:** 13 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/internal/nonint/nonint.go`, `/home/user/Bonsai/internal/nonint/result.go`, `/home/user/Bonsai/internal/nonint/runner.go`, `/home/user/Bonsai/internal/generate/list_snapshot.go`, `/home/user/Bonsai/station/agent/Skills/critic-agent-prompts.md`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, git diff, ls, grep), Glob, Grep, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`, `station/Playbook/Status.md`. Ran `git log --oneline --since="7 days ago"` and `git diff --name-only HEAD~10..HEAD` to identify changed files outside `station/`.
- **Result:** 43 commits in the last 7 days (many today's routine runs + Plan 41 ship on 2026-06-16 + Plan 40 on 2026-06-13). Key non-station file changes: `cmd/list.go`, `cmd/remove.go`, `cmd/root.go`, `cmd/update.go`, `internal/generate/list_snapshot.go`, `internal/nonint/contract_test.go`, `internal/nonint/remove.go`, `internal/nonint/result_test.go`, `internal/nonint/update.go`, `docs/agent-interface.md`, `docs/formats.md`, `internal/tui/updateflow/run.go`. Plan 41 (Headless CLI contract, 5 phases) is the dominant code change since last doc-freshness run.
- **Issues:** Multiple documentation targets identified as stale (see Findings Summary).

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` in full. Compared tech stack, folder structure, CLI command count, and key metrics against actual code state.
- **Result:** Tech stack table is accurate. CLI commands section still shows "8 (init, add, remove, list, catalog, update, guide, validate)" which is correct — `completion` is intentionally hidden (`HiddenDefaultCmd = true` in cmd/root.go). However, INDEX.md has **no mention of headless/non-interactive mode, `--json` flags, `docs/agent-interface.md`, or MCP-readiness** — all are significant Plan 41 deliverables. Architecture diagram does not reference `internal/nonint/` package.
- **Issues:** Medium drift — INDEX.md missing headless CLI capability and `internal/nonint` package from architecture diagram.

### Step 3: Check navigation links
- **Action:** Verified all file paths linked in `station/CLAUDE.md` (Core, Protocols, Workflows, Skills, Routines, Sensors, External References sections). Checked 55 paths total.
- **Result:** All 55 paths resolve to existing files/directories. Zero broken links. However, discovered two **undocumented files** in agent skill/workflow directories that are not referenced in the CLAUDE.md nav tables: `agent/Skills/critic-agent-prompts.md` and `agent/Workflows/plan-grilling.md` — both exist on disk but have no nav entries.
- **Issues:** Two files exist with no nav row. Agent cannot discover them via CLAUDE.md routing table.

### Step 4: Report findings
- **Action:** Consolidated findings into the table below. Per procedure, updates are flagged for user decision only — not executed.
- **Result:** 5 drift items found (1 high, 3 medium, 1 low). All proposed as user-decision items.
- **Issues:** None — audit scope completed cleanly.

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for "Doc Freshness Check": Last Ran → 2026-06-17, Next Due → 2026-06-24, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `internal/nonint/` package (8 files: nonint.go, config.go, result.go, events.go, runner.go, remove.go, update.go, add.go) shipped in Plan 41 (2026-06-16) — not documented anywhere in `code-index.md` | `station/code-index.md` | Flagged — needs new section after Workspace-path Validation |
| 2 | MEDIUM | `internal/generate/list_snapshot.go` (Plan 41 Phase 4) — `ListSnapshot`, `ListAgent`, `SerializeJSON()` types/functions not in `code-index.md` Generator section | `station/code-index.md` → Generator / `catalog_snapshot.go` section | Flagged — needs row in Generate section |
| 3 | MEDIUM | `station/agent/Workflows/plan-grilling.md` exists (added 2026-06-13 via Plan 40 station work) but is absent from CLAUDE.md Workflows nav table | `station/CLAUDE.md` Workflows table | Flagged — needs new row; trigger: "Adversarially reviewing a drafted plan before dispatch" |
| 4 | MEDIUM | `station/agent/Skills/critic-agent-prompts.md` exists but is absent from CLAUDE.md Skills nav table | `station/CLAUDE.md` Skills table | Flagged — needs new row; trigger: "Dispatching 6-critic plan-grilling agents; accessing verbatim prompt templates" |
| 5 | LOW | `station/INDEX.md` architecture diagram and description have no mention of `internal/nonint/` package, headless CLI contract, `--non-interactive` / `--json` flags, or `docs/agent-interface.md` — the biggest feature shipped since last doc-freshness run | `station/INDEX.md` Architecture section + Key Metrics | Flagged — suggest adding a short note under Architecture or a new "Headless Interface" row to Document Registry |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

### Finding 1 (HIGH) — `code-index.md` missing `internal/nonint/` package
The `internal/nonint/` package is the headless core for all mutating commands (init, add, update, remove). It has 8 source files and significant public surface:
- `runner.go` — `RunInit()`, `RunAdd()`, `RunUpdate()`, `RunRemove()`, exit code constants (`ExitOK=0`, `ExitInvalidConfig=2`, `ExitRuntime=3`, `ExitWrongCWDForInit=4`, `ExitConflict=5`)
- `result.go` — `Result` struct (headless return value), `Counts()` method
- `events.go` — `EmitJSONL()`, `EmitFile()`, `EmitSummary()`
- `config.go` — `LoadConfig()`, `applyDefaults()`
- `remove.go` — `RunRemove()` implementation
- `update.go` — `RunUpdate()` implementation

**Proposed addition:** New `## Headless Core (internal/nonint/)` section in `code-index.md`, similar in shape to the `internal/validate/` section.

### Finding 2 (MEDIUM) — `code-index.md` missing `list_snapshot.go`
`internal/generate/list_snapshot.go` contains `ListSnapshot`, `ListAgent`, and `SerializeJSON()` — the JSON contract for `bonsai list --json`. It lives alongside `catalog_snapshot.go` and should be documented in the Generator section under `catalog_snapshot.go`'s entry.

### Finding 3 (MEDIUM) — `plan-grilling.md` not in CLAUDE.md Workflows nav
The workflow exists and is actively used (Plan 40 station work). Suggested nav row:
```
| Adversarially reviewing a drafted plan before dispatch; Running 6 parallel critic agents against a plan to convergence | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

### Finding 4 (MEDIUM) — `critic-agent-prompts.md` not in CLAUDE.md Skills nav
The skill exists as the verbatim prompt source for plan-grilling critic agents. Suggested nav row:
```
| Dispatching 6-critic plan-grilling agents; Accessing verbatim critic prompt templates for plan adversarial review | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

### Finding 5 (LOW) — INDEX.md missing headless interface mention
Plan 41 (headless CLI contract) is described in Status.md as one of the largest milestones. The canonical reference doc is `docs/agent-interface.md`. Suggested addition to INDEX.md Document Registry:
```
| `docs/agent-interface.md` | Headless CLI contract — flags, exit codes, JSONL/JSON serialization for non-interactive use | When driving Bonsai from CI, MCP, or AI agents |
```

And a note to the Architecture section adding `internal/nonint/` to the package diagram.

## Notes for Next Run
- All 5 findings are quick fixes (row additions, one new section). A single focused doc-refresh pass could resolve all of them in one commit without needing a formal plan.
- The `plan-grilling` + `critic-agent-prompts` omission from CLAUDE.md nav has likely been present since 2026-06-13 when they were added. If the user runs plan-grilling again before the nav is updated, they will need to remember the file path manually.
- Code-index drift is recurring (flagged in 2026-05-04 and 2026-04-14 runs). Consider adding a code-index update step to the Plan completion checklist (in `issue-to-implementation.md` or `session-wrapup.md`) to catch this at ship time.
- `docs/agent-interface.md` is an important external-facing document (the MCP-readiness contract). Consider adding it to the Document Registry in INDEX.md as a permanent entry.
