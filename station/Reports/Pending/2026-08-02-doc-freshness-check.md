---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-02
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
- **Duration:** ~12 minutes
- **Files Read:** 12 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/.bonsai.yaml`, `/home/user/Bonsai/station/agent/Workflows/plan-grilling.md`, `/home/user/Bonsai/station/CLAUDE.md` (via system-reminder), `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/docs/agent-interface.md` (head only), `/home/user/Bonsai/cmd/completion.go` (head only)
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Bash (git log, ls, directory listings), Glob, Grep
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation
Read `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, and `station/code-index.md`. Compared against `git log --oneline --since=2026-05-04` (the prior run date). Identified 46 commits since last run, including Plans 40 and 41 (major feature work) and the `completion` command addition. Flagged documentation drift in INDEX.md and code-index.md.

### Step 2 — Check INDEX.md accuracy
Verified tech stack, folder structure, and project description. Tech stack table is accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template). Architecture overview accurately reflects the internal package layout. Key Metrics have two inaccuracies:
- **CLI command count** says 8 — `bonsai completion` was added in commit `2eae9d4` (2026-05-07), making it 9.
- **Catalog items** says ~50 — current count is 53 abilities (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines); still approximately accurate but slightly low.
- **docs/agent-interface.md** (Plan 41 Phase 5, commit `ab202c3`, 2026-06-16) is not referenced in the Document Registry — this is the canonical headless CLI contract document.

### Step 3 — Check navigation links
Verified all links in `station/CLAUDE.md` navigation tables against actual filesystem contents:
- All Core, Protocols, Workflows (listed), Skills (listed), Routines, Sensors links resolve correctly.
- `../.bonsai/catalog.json` resolves.
- `../.bonsai.yaml` resolves.
- All External References (Status.md, Roadmap.md, SecurityStandards.md, Plans/Active/, Backlog.md, KeyDecisionLog.md, Reports/Pending/, code-index.md) resolve.
- **Two files exist but are NOT listed in CLAUDE.md:**
  - `station/agent/Workflows/plan-grilling.md` — custom plan-grilling workflow (added commit `6995d4f`), not in Workflows table
  - `station/agent/Skills/critic-agent-prompts.md` — the 6 critic-agent prompts referenced by plan-grilling, not in Skills table
- Also verified `agent/Workflows/` links in Core/Protocols/Workflows/Skills directories — no broken links found within listed items.

### Step 4 — Report findings
Seven findings identified, ranging Moderate to Info. All flagged for user review per procedure (propose updates; don't execute). See Findings Summary below.

### Step 5 — Update dashboard
Updated `agent/Core/routines.md` — Doc Freshness Check row: Last Ran → 2026-08-02, Next Due → 2026-08-09, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Moderate | `plan-grilling.md` not in CLAUDE.md Workflows navigation table — agent won't discover this workflow without direct file knowledge | `station/CLAUDE.md` Workflows section | Flagged for user review |
| 2 | Moderate | `critic-agent-prompts.md` not in CLAUDE.md Skills navigation table — required by plan-grilling workflow, not surfaced in nav | `station/CLAUDE.md` Skills section | Flagged for user review |
| 3 | Moderate | code-index.md missing Plan 41 headless CLI coverage — `--json`, `--non-interactive`, `--from-config`, `--yes/--from`, `--skip-conflicts` flags and their internal structures not documented | `station/code-index.md` CLI Commands section | Flagged for user review |
| 4 | Low | INDEX.md CLI command count stale — says 8 commands, should be 9; `bonsai completion` added 2026-05-07 not listed | `station/INDEX.md` line 33 | Flagged for user review |
| 5 | Low | code-index.md missing `bonsai completion` entry — `cmd/completion.go` has no row in the CLI Commands table | `station/code-index.md` CLI Commands table | Flagged for user review |
| 6 | Info | `docs/agent-interface.md` not in INDEX.md Document Registry — canonical headless CLI contract shipped in Plan 41 Phase 5 | `station/INDEX.md` Document Registry table | Flagged for user review |
| 7 | Info | INDEX.md catalog item count says ~50; actual count is 53 abilities (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines) — still approximately accurate | `station/INDEX.md` line 32 | Flagged for user review |

---

## Errors & Warnings

None.

---

## Items Flagged for User Review

### F1 + F2 — station/CLAUDE.md missing two entries (Moderate)

Two custom files shipped since last CLAUDE.md update are not in the navigation table. This means the tech-lead agent won't know to load them unless it discovers the files directly.

**Proposed addition to Workflows table:**

```markdown
| Adversarially reviewing a drafted plan before dispatch; Running 6-critic convergence review (Security/Architecture/Simplicity/Risk/Verification/Reality) | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

**Proposed addition to Skills table:**

```markdown
| Dispatching the 6 plan-grilling critics; referencing the verbatim prompt templates for each critic role | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

Note: since these are custom files not tracked in `.bonsai.yaml`, running `bonsai update` will not add them — they must be added manually to `station/CLAUDE.md`.

### F3 — code-index.md Plan 41 headless coverage missing (Moderate)

Plan 41 shipped a substantial headless-mode feature set (PRs #120–125, merged 2026-06-16). The code-index.md CLI Commands section does not reflect:
- `bonsai list --json` (JSONL output)
- `bonsai init/add --non-interactive --from-config <path>`
- `bonsai remove --yes/--from`
- `bonsai update --non-interactive/--skip-conflicts`
- Exit codes: `ExitConflict=5`, `ExitNotFound=3`, `ExitAlreadyInstalled=4`
- `docs/agent-interface.md` (the canonical written contract)

Recommend a "Headless / Non-Interactive Mode" subsection in code-index.md. Approximate effort: 15–20 lines.

### F4 + F5 — One-liner fixes (Low)

Two small factual corrections:

1. `station/INDEX.md` line 33: `CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate)` → `9 (init, add, remove, list, catalog, update, guide, validate, completion)`

2. `station/code-index.md` CLI Commands table: add row for `bonsai completion`:
   ```
   | `bonsai completion` | `cmd/completion.go:21` | `completionCmd` — shell completion scripts (bash/zsh/fish/powershell) |
   ```

### F6 — docs/agent-interface.md not in Document Registry (Info)

`docs/agent-interface.md` is the canonical headless CLI contract. Adding it to `station/INDEX.md` Document Registry makes it discoverable:

```markdown
| `docs/agent-interface.md` | Headless CLI contract — flags, JSONL serialization, exit codes for driving Bonsai non-interactively from AI agents or CI | When building tooling on top of Bonsai; Plan 42 MCP server work |
```

### F7 — Catalog item count (Info)

`station/INDEX.md` line 32 says `~50`; actual count is 53. Low urgency — `~50` is still defensibly correct. Could update to `~55` or `53` for precision.

---

## Notes for Next Run

- After user applies F1/F2 (CLAUDE.md nav table additions), verify via `bonsai validate` that plan-grilling and critic-agent-prompts are being detected as custom files in the workspace scan.
- F3 (headless code-index coverage) is the highest-effort item — consider dispatching a focused doc-update to a code agent rather than doing inline.
- If Plan 42 (MCP server) has shipped by next run, INDEX.md and code-index.md will need another doc-freshness sweep for new commands/packages.
