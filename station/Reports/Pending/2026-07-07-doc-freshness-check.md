---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-07
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
- **Duration:** ~6 min
- **Files Read:** 11 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/41-headless-cli-contract.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/CLAUDE.md` (root), `/home/user/Bonsai/go.mod`
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, file existence checks, find, grep), Glob, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation vs. recent git history

Ran `git log --since="60 days ago"` to identify code changes since last run (2026-05-04). Key commits found:

| Date | Commit | Change |
|------|--------|--------|
| 2026-05-07 | `9fcf64c` / `2eae9d4` | **completion command added** — PR #78, `cmd/completion.go` |
| 2026-05-07 | `a5a4185` | Plan 37 doc refresh bundle (code-index sync, INDEX Go version) |
| 2026-05-13 | `410a5f1` | Plan 39 — `bonsai init`/`add --non-interactive --from-config` (v0.4.2) |
| 2026-05-13 | `584b82b` | v0.4.3 hotfix — absolute paths baked into sensor hooks |
| 2026-06-13 | `1e715c7` / `a540fdd` / `2aef7fd` | Plan 40 Ph1–3 — frozen schemas, project-validate, memory-routing docs |
| 2026-06-16 | `65932b9` – `ab202c3` | **Plan 41** — headless CLI cores, list --json, exit contract, agent-interface.md |

Three significant feature deliveries since last check: v0.4.2 non-interactive flags, Plan 40 schema freeze + validate improvements, and Plan 41 full headless CLI contract.

### Step 2 — Check INDEX.md accuracy

Read `station/INDEX.md` and compared against codebase:

- **Go version**: INDEX.md says "Go 1.25+" — `go.mod` confirms `go 1.25.0` / `toolchain go1.25.9`. **Accurate.**
- **Agent types**: INDEX.md says "6 (tech-lead, fullstack, backend, frontend, devops, security)" — `ls catalog/agents/` confirms 6 types. **Accurate.**
- **CLI commands count**: INDEX.md says "8 (init, add, remove, list, catalog, update, guide, validate)" — but `ls cmd/` shows `completion.go` also present (added PR #78, 2026-05-07). **DRIFT: should be 9.**
- **Catalog items**: INDEX.md says "~50" — `find catalog -name meta.yaml | wc -l` returns 53. The approximation "~50" is borderline acceptable but could be updated to "~55" for accuracy.
- **Architecture overview**: The architecture box does not mention headless CLI features (Result shapes, list --json, exit codes contract, `docs/agent-interface.md`). These are new Plan 41 surface-area items not captured.

### Step 3 — Check navigation links

Verified all 51 links in `station/CLAUDE.md` navigation tables. Checked:
- agent/Core/ files (identity, memory, self-awareness)
- agent/Protocols/ (4 files)
- agent/Workflows/ (8 files)
- agent/Skills/ (6 files)
- agent/Routines/ (7 files)
- agent/Sensors/ (10 files)
- Playbook/ (Status, Roadmap, Backlog, Standards/SecurityStandards)
- Logs/ (FieldNotes, KeyDecisionLog)
- Reports/ (report-template)
- ../.bonsai/catalog.json, ../.bonsai.yaml, code-index.md

**Result: All 51 links resolve to existing files. No broken links.**

Also checked links referenced in `agent/Core/routines.md`, `agent/Protocols/`, `agent/Workflows/` — no issues found.

### Step 4 — Report findings (flagged for user decision)

See Findings Summary table below. All proposed updates are flagged — not executed.

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` — `Doc Freshness Check` row: Last Ran → 2026-07-07, Next Due → 2026-07-14, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | CLI commands count is "8" but `completion` (PR #78, 2026-05-07) brings total to 9 | `station/INDEX.md` line 33 | Flagged — proposed update: change "8 (init, add, remove, list, catalog, update, guide, validate)" → "9 (init, add, remove, list, catalog, update, guide, validate, completion)" |
| 2 | Medium | `bonsai completion` command absent from CLI Commands table | `station/code-index.md` — CLI Commands section | Flagged — proposed: add row `bonsai completion \| cmd/completion.go:... \| Shell completion for bash/zsh/fish/powershell` |
| 3 | Low | `completion.go` missing from `cmd/` project structure listing | root `CLAUDE.md` — project structure block | Flagged — proposed: add `│   ├── completion.go         ← bonsai completion (shell completion subcommand)` after `guide.go` line |
| 4 | Low | Plan 41 fully shipped (all 5 phases, PRs #120–#125 merged 2026-06-16) but file remains in `Plans/Active/` with `status: ready` | `station/Playbook/Plans/Active/41-headless-cli-contract.md` | Flagged — proposed: move to `Plans/Archive/41-headless-cli-contract.md` and update status to `shipped` |
| 5 | Low | INDEX.md architecture overview does not mention headless CLI surface added in Plan 41 (list --json, exit codes 0/2/3/4/5, headless cores, agent-interface.md contract doc) | `station/INDEX.md` — Architecture Overview | Flagged — low priority since the section is high-level; proposed: add a note or extend Key Metrics to reflect |
| 6 | Info | Catalog item count: INDEX.md says "~50", actual count is 53 | `station/INDEX.md` line 32 | Flagged — minor; proposed: update to "~55" for accuracy |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

**Finding #1 (Medium) — completion command count drift in INDEX.md**
Proposed fix in `station/INDEX.md` line 33:
- Before: `| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |`
- After: `| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |`

**Finding #2 (Medium) — completion command missing from code-index.md**
Add a new row to the CLI Commands table after `bonsai validate`:
```
| `bonsai completion` | `cmd/completion.go` | Shell completion for bash/zsh/fish/powershell |
```

**Finding #3 (Low) — completion.go missing from root CLAUDE.md structure**
Add to `cmd/` block after `guide.go`:
```
│   ├── completion.go        ← bonsai completion (shell completion for bash/zsh/fish/powershell)
```

**Finding #4 (Low) — Plan 41 should be archived**
- `Plans/Active/41-headless-cli-contract.md` → move to `Plans/Archive/41-headless-cli-contract.md`
- Update `status: ready` → `status: shipped` in frontmatter

---

## Notes for Next Run

- The completion command drift (#1/#2/#3) is the highest-impact finding — it's been stale since 2026-05-07 (2 months). Quick 3-file fix.
- Plan 40 (`40-odysseus-platform-integration.md`) correctly remains in Active/ — Phase 4 is still HELD per Status.md.
- All navigation links are healthy — no broken links detected. The link surface is stable.
- Backlog Hygiene ran the same day (2026-07-07) and flagged the HOMEBREW_TAP_TOKEN PAT expiry (~2026-07-15, 8 days). Cross-reference with that report.
- Next run: 2026-07-14. Priority: confirm the 3 completion-command doc fixes have been applied.
