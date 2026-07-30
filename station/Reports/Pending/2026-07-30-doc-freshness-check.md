---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-30
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
- **Duration:** ~5 min
- **Files Read:** 14 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/CLAUDE.md` (system context), `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/Playbook/Backlog.md` (grep), `/home/user/Bonsai/internal/nonint/nonint.go`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md` (head), `/home/user/Bonsai/station/agent/Skills/critic-agent-prompts.md` (head), `/home/user/Bonsai/cmd/completion.go` (head), directory listings for all station/agent/ subdirs and cmd/, internal/, catalog/
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, grep, head), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against git history (last 7 days)

Ran `git log --since="7 days ago" --oneline --name-only`. Found exactly 2 commits:
- `b6e860a` — backlog-hygiene routine: updated RoutineLog.md, created backlog-hygiene report, updated routines.md dashboard
- `5a55d96` — backlog-hygiene content change: removed 3 resolved items from Backlog.md

No code changes in the last 7 days. All recent commits are routine-maintenance-only. No new features, services, or config were introduced that would require documentation updates.

Cross-checked the full `internal/` package listing and `cmd/` listing against INDEX.md and CLAUDE.md to detect any older drift.

### Step 2 — Check INDEX.md accuracy

Verified INDEX.md tech stack, folder structure, and project description against the actual codebase. The stack (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS) matches reality. The architecture diagram paths all resolve.

**Two accuracy gaps found:**

1. `internal/nonint/` package is not mentioned anywhere in INDEX.md or the project-level CLAUDE.md architecture section. It is a real, substantial package (8 source files: config.go, events.go, nonint.go, remove.go, result.go, runner.go, update.go — plus tests and testdata). Its purpose: headless/non-interactive mode for bonsai CLI (RunInit, RunAdd, config, structured Result output, JSONL emission). This is genuine documentation drift.

2. INDEX.md states "CLI commands: 8 (init, add, remove, list, catalog, update, guide, validate)" but `cmd/completion.go` registers a 9th `completion` command explicitly added to replace Cobra's auto-generated version. The count and command list are slightly stale.

### Step 3 — Check navigation links

Verified every link in station/CLAUDE.md against the filesystem. All explicitly listed files resolve:

- **Core (3/3):** identity.md ✓, memory.md ✓, self-awareness.md ✓
- **Protocols (4/4):** memory.md ✓, scope-boundaries.md ✓, security.md ✓, session-start.md ✓
- **Workflows (9/9):** code-review.md ✓, planning.md ✓, pr-review.md ✓, security-audit.md ✓, session-logging.md ✓, test-plan.md ✓, session-wrapup.md ✓, issue-to-implementation.md ✓, routine-digest.md ✓
- **Skills (6/6):** planning-template.md ✓, review-checklist.md ✓, issue-classification.md ✓, pr-creation.md ✓, bubbletea.md ✓, bonsai-model.md ✓
- **Routines (7/7):** all 7 routine files present ✓
- **Sensors (10/10):** all 10 sensor scripts present ✓
- **External refs:** INDEX.md ✓, Status.md ✓, Roadmap.md ✓, SecurityStandards.md ✓, Plans/Active/ ✓, Backlog.md ✓, KeyDecisionLog.md ✓, Reports/Pending/ ✓, report-template.md ✓, code-index.md ✓, ../.bonsai/catalog.json ✓

Zero broken links found.

**Unlisted files detected in agent directories (not broken links — inverse gap):**

- `agent/Workflows/plan-grilling.md` — exists but not listed in CLAUDE.md Workflows table. File itself notes "full Bonsai-catalog integration pending (Backlog)" and Backlog confirms. Intentional omission.
- `agent/Skills/critic-agent-prompts.md` — exists but not listed in CLAUDE.md Skills table. Same status: in Backlog pending integration. Intentional omission.
- `agent/Skills/bubbletea/` directory (4 files: components.md, emoji-width-fix.md, golden-rules.md, troubleshooting.md) — the nav table links `bubbletea.md` (the top-level file) but the subdirectory's constituent files are not separately linked. Low-priority cosmetic gap.

### Step 4 — Report findings

Three findings flagged for user decision (see Findings Summary below). Per procedure: proposing updates, not executing.

### Step 5 — Update dashboard

Dashboard row for Doc Freshness Check updated: Last Ran → 2026-07-30, Next Due → 2026-08-06, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `internal/nonint/` package undocumented — not mentioned in INDEX.md or CLAUDE.md architecture section, despite being a real multi-file package (headless/non-interactive CLI mode) | `station/INDEX.md`, project `CLAUDE.md` | Flagged for user — proposed add to INDEX.md architecture table and CLAUDE.md project structure |
| 2 | Low | INDEX.md CLI command count reads "8" but `completion` is a 9th registered command | `station/INDEX.md` (Key Metrics table) | Flagged for user — proposed update count to 9 and add `completion` to the list |
| 3 | Low | `agent/Skills/bubbletea/` subdirectory (4 files) unlisted in CLAUDE.md Skills nav — only the parent `bubbletea.md` is linked | `station/CLAUDE.md` (Skills table) | Flagged for user — low priority; subdirectory files are supporting detail; may add footnote or sub-link |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

**Finding 1 (Medium) — Add `internal/nonint/` to INDEX.md architecture section**

Proposed addition to INDEX.md architecture diagram:
```
internal/nonint/      ← headless/non-interactive mode — RunInit/RunAdd orchestrators, structured Result output, JSONL emission
```
And to the CLAUDE.md project-level `internal/` tree:
```
├── nonint/
│   ├── nonint.go    ← package anchor + godoc entry point
│   ├── config.go    ← LoadConfig + applyDefaults
│   ├── result.go    ← Result + Counts (structured headless return value)
│   ├── events.go    ← EmitJSONL + EmitFile + EmitSummary helpers
│   └── runner.go    ← RunInit + RunAdd orchestrators + exit codes
```

**Finding 2 (Low) — Update CLI command count in INDEX.md**

Change:
```
| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |
```
To:
```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```

**Finding 3 (Low) — `bubbletea/` subdirectory in Skills**

The `agent/Skills/bubbletea/` directory holds supplementary BubbleTea reference files. The current nav entry points only to `bubbletea.md`. No action required unless the subdirectory files should be surfaced individually — user's call.

---

## Notes for Next Run

- All navigation links are clean — no broken links found. Next run can focus on code drift rather than link validation.
- If `plan-grilling` Backlog item is actioned before next run, its workflow and skill should be wired into CLAUDE.md nav.
- If `internal/nonint/` gets documented (Finding 1), mark it resolved before next run.
- The 84-day routine gap (noted in backlog-hygiene report) means INDEX.md and CLAUDE.md accumulated moderate drift without detection. Consider running doc-freshness-check more frequently if another long gap occurs.
