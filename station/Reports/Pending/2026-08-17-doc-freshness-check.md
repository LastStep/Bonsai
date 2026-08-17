---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-17
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
- **Duration:** ~10 min
- **Files Read:** 12
  - `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`
  - `/home/user/Bonsai/station/INDEX.md`
  - `/home/user/Bonsai/station/CLAUDE.md` (via system context)
  - `/home/user/Bonsai/station/Playbook/Status.md`
  - `/home/user/Bonsai/station/Playbook/Roadmap.md`
  - `/home/user/Bonsai/station/Playbook/Backlog.md`
  - `/home/user/Bonsai/station/code-index.md`
  - `/home/user/Bonsai/station/agent/Core/routines.md`
  - `/home/user/Bonsai/station/Logs/RoutineLog.md`
  - `/home/user/Bonsai/docs/agent-interface.md` (existence confirmed)
  - Directory listings: `station/agent/Core/`, `Protocols/`, `Workflows/`, `Skills/`, `Sensors/`, `Routines/`, `cmd/`, `internal/`
- **Files Modified:** 2
  - `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update)
  - `/home/user/Bonsai/station/Logs/RoutineLog.md` (log append)
- **Tools Used:** Read, Bash (git log, ls), Glob, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation vs recent git history

Read `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, and `station/Playbook/Backlog.md`. Ran `git log --oneline --since="2026-05-04"` to capture all commits since the last run (103 days). Found 50+ commits covering Plans 40 and 41, plus several releases and routine runs.

**Key changes since last run:**
- **Plan 41 (2026-06-16)** — Headless CLI contract + MCP-ready cores. All 5 phases merged (PRs #120/122/123/121/125). New `internal/nonint/` package (headless contract layer), JSONL/exit-code contract, `list --json`, and `docs/agent-interface.md` contract doc.
- **Plan 40 (2026-06-13)** — Odysseus platform integration (Phases 1–3). Frozen v1 schemas, root-relative scaffolding, project-level `validate` pass.
- **v0.4.3 hotfix (2026-05-13)** — Sensor hooks bake absolute paths.
- **v0.4.2 (2026-05-13)** — Non-interactive flags for `bonsai init/add`.
- **`bonsai completion` command (2026-05-07)** — Shell completion for bash/zsh/fish/powershell. Added via PR #78.

### Step 2 — Check INDEX.md accuracy

Verified tech stack, folder structure, and project description. Stack is accurate. Architecture overview is largely accurate but missing the new `internal/nonint/` package added in Plan 41. CLI commands count is stale.

**Drift found:**
- Key Metrics: CLI commands count says "8" and lists the 8 original commands. `bonsai completion` was added ~2026-05-07 — count should be 9.
- Architecture overview: `internal/nonint/` package (headless contract layer, Plan 41) is not mentioned.
- Document Registry: `docs/agent-interface.md` (Plan 41 contract doc for MCP/agent integration) is not registered.

### Step 3 — Check navigation links

Checked all links in `station/CLAUDE.md` navigation tables against actual files. Checked `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/`, `agent/Sensors/`, `agent/Routines/` using Glob and ls.

**All links resolve — no broken links found.**

However, two files exist that are NOT in the CLAUDE.md navigation tables:

1. `agent/Workflows/plan-grilling.md` — present in `Workflows/` dir but absent from the Workflows nav table in CLAUDE.md. This workflow was added (commit `6995d4f`) as a 6-critic adversarial plan review pipeline.

2. `agent/Skills/critic-agent-prompts.md` — present in `Skills/` dir but absent from the Skills nav table in CLAUDE.md. This skill contains critic agent prompts used during plan grilling.

These are navigation gaps, not broken links — the files exist and are functional, just undiscoverable via the nav table.

### Step 4 — Report findings

Six findings catalogued below (4 medium, 2 low). Per routine procedure: all are flagged for user decision — no doc updates executed.

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` dashboard row for Doc Freshness Check: `Last Ran → 2026-08-17`, `Next Due → 2026-08-24`, `Status → done`.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `plan-grilling` workflow exists but missing from CLAUDE.md Workflows nav table | `station/CLAUDE.md` | Flagged for user — no edit |
| 2 | Medium | `critic-agent-prompts` skill exists but missing from CLAUDE.md Skills nav table | `station/CLAUDE.md` | Flagged for user — no edit |
| 3 | Medium | `docs/agent-interface.md` (Plan 41 headless contract) not in INDEX.md Document Registry | `station/INDEX.md` | Flagged for user — no edit |
| 4 | Medium | `internal/nonint/` package (Plan 41 headless contract layer) not in INDEX.md architecture overview or code-index.md | `station/INDEX.md`, `station/code-index.md` | Flagged for user — no edit |
| 5 | Low | CLI commands count stale: says 8 + lists 8 commands; `bonsai completion` added ~2026-05-07 → should be 9 | `station/INDEX.md` Key Metrics table | Flagged for user — no edit |
| 6 | Low | `bonsai completion` command (cmd/completion.go) missing from code-index.md CLI Commands table | `station/code-index.md` | Flagged for user — no edit |

---

## Proposed Updates (for user decision)

### Finding 1 — Add plan-grilling to CLAUDE.md Workflows table

Add a row to the Workflows table in `station/CLAUDE.md`:

```
| Running adversarial (6-critic) grilling on a new implementation plan | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

### Finding 2 — Add critic-agent-prompts to CLAUDE.md Skills table

Add a row to the Skills table in `station/CLAUDE.md`:

```
| Using critic agent prompts during plan grilling; Prompts for adversarial review passes | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

### Finding 3 — Add agent-interface.md to INDEX.md Document Registry

Add a row to the Document Registry table in `station/INDEX.md`:

```
| `docs/agent-interface.md` | Headless CLI contract — JSONL event format, exit codes, non-interactive flags for MCP/agent integration | When building MCP server (Plan 42) or writing agent-driven automation |
```

### Finding 4 — Add internal/nonint to INDEX.md and code-index.md

In `station/INDEX.md` Architecture Overview:
- Add `internal/nonint/` to the internal packages listed in the ASCII diagram, annotated as "headless contract layer — result shapes, JSONL events, exit codes".

In `station/code-index.md`:
- Add a new section `## Headless Contract (internal/nonint/) — Plan 41` with the key types and files (config.go, events.go, nonint.go, remove.go, result.go, runner.go, update.go).

### Finding 5 — Update CLI commands count in INDEX.md

In Key Metrics table: `8 (init, add, remove, list, catalog, update, guide, validate)` → `9 (init, add, remove, list, catalog, update, guide, validate, completion)`.

### Finding 6 — Add completion command to code-index.md

Add a row to the CLI Commands table in `station/code-index.md`:

```
| `bonsai completion` | `cmd/completion.go` | `completionCmd` — shell completion for bash/zsh/fish/powershell |
```

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

All 6 findings require user decision before any docs are updated. Priority order:

1. **(Medium) Findings 1 + 2** — CLAUDE.md nav gaps for `plan-grilling` and `critic-agent-prompts`. These are the most operationally impactful: the agent cannot navigate to these files from the standard nav table. Easy fixes — two new table rows.

2. **(Medium) Findings 3 + 4** — INDEX.md and code-index.md do not reflect Plan 41's `internal/nonint/` package or `docs/agent-interface.md`. Important for the upcoming MCP server work (Plan 42). Moderate effort to add.

3. **(Low) Findings 5 + 6** — `bonsai completion` command missing from INDEX.md count and code-index.md. Minor drift — easy one-liners.

---

## Notes for Next Run

- Plans 40 + 41 are now archived. Status.md In Progress and Pending are clean.
- Plans/Active/ is currently empty — check for new plans at next run.
- If Plan 42 (MCP server) has landed by next run, `docs/agent-interface.md` will be in active use — verify it is registered.
- `HOMEBREW_TAP_TOKEN` PAT rotation was flagged as URGENT by backlog-hygiene (target date 2026-07-15 has passed) — not a doc freshness issue but noted for awareness.
