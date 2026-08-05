---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-05
status: success
---

# Routine Report — Doc Freshness Check

## Overview

Routine audit of project documentation for drift against the actual codebase. Scope: `station/INDEX.md`, `station/CLAUDE.md`, `station/code-index.md`, root `CLAUDE.md`, and all navigation links in `station/agent/` subdirectories. 9 findings, all documentation-only — no code changes proposed. 1 action taken (dashboard update).

---

## Execution Metadata

| Field | Value |
|-------|-------|
| Date | 2026-08-05 |
| Last run | 2026-05-04 |
| Gap | 93 days (3× overdue) |
| Execution mode | Subagent (loop.md dispatch) |
| Duration | ~8 min |
| Commits reviewed | 31 commits since 2026-05-07 |
| Plans shipped in window | 39, 40 (P1–3), 41 (all 5 phases) |

---

## Procedure Walkthrough

### Step 1 — Scan Project Documentation

Ran `git log --oneline --since="2026-05-07"` — 31 commits found covering:
- **Plan 39**: `bonsai init`/`add` `--non-interactive --from-config` (v0.4.2)
- **Plan 40 P1–3**: frozen v1 schemas + root-relative scaffolding, project-level validate pass, memory-routing docs + guide Formats page (v0.5.0, untagged)
- **Plan 41 all 5 phases**: headless `*Result` cores for init/add/update/remove + `list --json` + `docs/agent-interface.md` contract + ExitConflict=5 (merged 2026-06-16)
- **PR #78**: `bonsai completion [bash|zsh|fish|powershell]` — external contribution, merged 2026-05-07

Key new artifacts that could cause doc drift: `internal/nonint/` package (new), `completion.go` command (new), `list_snapshot.go` in generate, platform-split catalog snapshot files, headless exit code contract, `docs/agent-interface.md`.

### Step 2 — Check INDEX.md Accuracy

Read `station/INDEX.md`. Compared Key Metrics, Tech Stack, and Architecture sections against reality.

- **Tech Stack**: accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea all correct.
- **Key Metrics (CLI commands)**: STALE — says `8 (init, add, remove, list, catalog, update, guide, validate)`. PR #78 added `bonsai completion` as a 9th command on 2026-05-07. Count and list both need updating.
- **Architecture diagram**: STALE — lists 6 internal packages; missing `internal/nonint/` which was added in Plans 39+41 and contains all headless CLI cores (runner.go, events.go, result.go, config.go, remove.go, update.go).
- **Headless features**: not documented anywhere in INDEX.md — `--non-interactive`, `--from-config`, `list --json`, exit code contract, and `docs/agent-interface.md` are significant user-facing changes introduced since last check. INDEX.md is not an API reference but a summary mention would help orientation.

### Step 3 — Check Navigation Links

**station/CLAUDE.md**:
- Core links: all 3 files exist (identity.md, memory.md, self-awareness.md) — OK
- Protocol links: all 4 files exist (memory.md, scope-boundaries.md, security.md, session-start.md) — OK
- Workflow links: 9 files listed, all 9 exist — OK. However `plan-grilling.md` exists in `agent/Workflows/` and is NOT listed.
- Skill links: 6 files listed, all 6 exist — OK. However `critic-agent-prompts.md` exists in `agent/Skills/` and is NOT listed.
- Sensor links: all 10 files listed exist — OK.
- Routine links: all 7 files listed exist — OK.
- Bonsai Reference links: `.bonsai/catalog.json` and `.bonsai.yaml` both exist — OK.

**Root CLAUDE.md (project-level)**:
- cmd/ tree section: Missing `completion.go`. New test files (`*_nonint_test.go`) also not listed (test files are often omitted from tree docs, so this is lower severity).
- internal/ tree section: Missing `internal/nonint/` package entirely. This is a significant omission — nonint is a primary package for Plans 39+41 deliverables.
- internal/generate/ section: Missing `list_snapshot.go`, `catalog_snapshot_unix.go`, and `catalog_snapshot_windows.go` (platform-split files added for cross-compile fix). Lower severity given these are implementation-detail files.

**station/code-index.md**:
- CLI Commands table: `bonsai completion` command missing. Table documents 9 commands but only lists 8; `completion.go` exists in `cmd/` and was shipped in PR #78.

**station/agent/Core/ files**:
- All internal links in identity.md, memory.md, self-awareness.md verified as valid (cross-checked against file system).

### Step 4 — Report Findings

See Findings Summary table below.

### Step 5 — Update Dashboard

Updated `station/agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-08-05, Next Due → 2026-08-12, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | MEDIUM | CLI command count stale — says 8, should be 9 (completion added PR #78, 2026-05-07) | `station/INDEX.md` line 33 | Flagged for user — propose update |
| 2 | MEDIUM | Architecture diagram missing `internal/nonint/` package (Plans 39+41 deliverable) | `station/INDEX.md` architecture section | Flagged for user — propose update |
| 3 | MEDIUM | `internal/nonint/` package missing from project structure tree | Root `CLAUDE.md` internal/ section | Flagged for user — propose update |
| 4 | MEDIUM | `completion.go` missing from cmd/ tree | Root `CLAUDE.md` cmd/ section | Flagged for user — propose update |
| 5 | MEDIUM | `bonsai completion` command missing from CLI Commands table | `station/code-index.md` | Flagged for user — propose update |
| 6 | LOW | `plan-grilling.md` workflow exists but not listed in Workflows nav table | `station/CLAUDE.md` Workflows section | Flagged for user — propose update |
| 7 | LOW | `critic-agent-prompts.md` skill exists but not listed in Skills nav table | `station/CLAUDE.md` Skills section | Flagged for user — propose update |
| 8 | LOW | Plans 40 and 41 still in `Plans/Active/` despite all phases shipped | `station/Playbook/Plans/Active/` | Flagged for user — Status Hygiene already flagged Plan 41 |
| 9 | INFO | Headless CLI features (--non-interactive, list --json, ExitConflict=5, agent-interface.md) not reflected in INDEX.md summary | `station/INDEX.md` | Flagged for user — optional orientation update |

---

## Errors & Warnings

None. All files read successfully. No broken navigation links found.

---

## Items Flagged for User Review

### Finding 1 & 2 — INDEX.md Key Metrics + Architecture (MEDIUM)

**Proposed change to `station/INDEX.md`:**

```diff
-| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |
+| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```

Architecture diagram — add after `internal/wsvalidate/`:
```
internal/nonint/      ← headless CLI cores — *Result types, JSONL event stream, exit code contract (Plans 39, 41)
```

Architecture diagram — update cmd/ line:
```diff
-cmd/ (Cobra)          ← CLI commands: init, add, remove, list, catalog, update, guide, validate
+cmd/ (Cobra)          ← CLI commands: init, add, remove, list, catalog, update, guide, validate, completion
```

### Finding 3 & 4 — Root CLAUDE.md structure (MEDIUM)

**Proposed changes to root `CLAUDE.md`:**

In cmd/ tree, add after `add.go`:
```
│   ├── completion.go        ← bonsai completion (bash/zsh/fish/powershell shell completions)
```

In internal/ tree, add after `wsvalidate/` and before `tui/`:
```
│   ├── nonint/
│   │   ├── nonint.go        ← headless mode entry + shared types
│   │   ├── runner.go        ← RunInit/RunAdd headless cores
│   │   ├── result.go        ← *Result types (InitResult, AddResult, etc.)
│   │   ├── events.go        ← JSONL event stream types
│   │   ├── config.go        ← headless config loading + overlay validation
│   │   ├── remove.go        ← headless remove core
│   │   └── update.go        ← headless update core
```

### Finding 5 — code-index.md (MEDIUM)

**Proposed change to `station/code-index.md` CLI Commands table:**

Add row after `bonsai validate`:
```
| `bonsai completion` | `cmd/completion.go` | `completionCmd` — shell completion scripts (bash/zsh/fish/powershell) |
```

### Finding 6 — plan-grilling.md not in CLAUDE.md (LOW)

`agent/Workflows/plan-grilling.md` exists but is absent from the Workflows navigation table in `station/CLAUDE.md`. Propose adding:

```
| Grilling a plan with adversarial critics before dispatch; Running the 6-critic review pipeline on a draft implementation plan | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

### Finding 7 — critic-agent-prompts.md not in CLAUDE.md (LOW)

`agent/Skills/critic-agent-prompts.md` exists but is absent from the Skills navigation table in `station/CLAUDE.md`. Propose adding:

```
| Using adversarial critic personas to stress-test a plan or proposal before dispatch | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

### Finding 8 — Plans 40 and 41 in Active/ (LOW)

`Plans/Active/40-odysseus-platform-integration.md` and `Plans/Active/41-headless-cli-contract.md` both shipped — all phases merged. Recommend archiving both:
```bash
git mv station/Playbook/Plans/Active/40-odysseus-platform-integration.md station/Playbook/Plans/Archive/
git mv station/Playbook/Plans/Active/41-headless-cli-contract.md station/Playbook/Plans/Archive/
```
Note: Plan 40 (v0.5.0) has a held tag — confirm tag release decision before or independently of archive.

### Finding 9 — Headless CLI features not in INDEX.md (INFO)

Plans 39 and 41 introduced significant user-facing features that could be noted in INDEX.md:
- `--non-interactive --from-config <path>` on init/add
- `--non-interactive --skip-conflicts` on update
- `--yes --from` on remove
- `list --json`
- Exit code contract: 0=ok, 2=conflict, 3=partial, 4=config-error, 5=conflict-on-exit
- `docs/agent-interface.md` machine-readable contract

Propose adding a "Headless / MCP-ready CLI" note to the Key Metrics or architecture section. Optional — INDEX.md is primarily a doc map, not a feature list.

---

## Notes for Next Run

- **Completion command is now a recurring reference point** — PR #78 added it; all three doc files (INDEX.md, root CLAUDE.md, code-index.md) still needed updating as of this run.
- **Check plan-grilling and critic-agent-prompts** — both files existed with no navigation entry. If they remain post-cleanup, add entries.
- **Monitor Plans/Active/** — routine should flag any plan that has been shipped (all phases merged) but not archived.
- **If MCP server (Plan 42) ships**, expect significant new surface area: `bonsai mcp` command, `internal/mcp/` package, docs additions. Run doc-freshness immediately after Plan 42 merges.
- **internal/generate/ platform-split files** (`catalog_snapshot_unix.go`, `catalog_snapshot_windows.go`) not listed in root CLAUDE.md — low priority but worth catching in next sweep if root CLAUDE.md tree is updated.
