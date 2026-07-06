---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-06
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (findings flagged; doc edits deferred to user per procedure — "propose but don't execute")
- **Duration:** ~10 min
- **Files Read:** 11
  - `station/agent/Routines/doc-freshness-check.md`
  - `station/CLAUDE.md`
  - `station/INDEX.md`
  - `station/agent/Core/routines.md`
  - `station/Playbook/Status.md`
  - `station/Playbook/Roadmap.md`
  - `station/Logs/RoutineLog.md`
  - `station/agent/Workflows/plan-grilling.md`
  - `station/agent/Skills/critic-agent-prompts.md`
  - `station/code-index.md`
  - `/home/user/Bonsai/CLAUDE.md` (root project CLAUDE.md)
- **Files Modified:** 2
  - `station/agent/Core/routines.md` (dashboard Last Ran / Next Due / Status)
  - `station/Logs/RoutineLog.md` (appended routine entry)
- **Tools Used:** Read, Bash (git log, ls, grep)
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1 — Scan project documentation + compare against recent git history

Ran `git log --since="60 days ago"` to capture all recent work. Key recent activity:

- **2026-07-06:** routine: run status-hygiene, backlog-hygiene
- **2026-06-16:** Plan 41 — Headless CLI Contract shipped (all 5 phases, PRs #120/#122/#123/#121/#125). Adds headless `*Result` cores for init/add/update/remove, `list --json`, `ExitConflict=5`, `docs/agent-interface.md` contract doc.
- **2026-06-13:** Plan 40 Phases 1–3 merged (frozen v1 schemas, root-relative scaffolding, project-level validate, memory-routing docs, guide Formats page). Phase 4 held.
- **2026-06-13:** `plan-grilling.md` workflow + `critic-agent-prompts.md` skill added to station (6-critic adversarial review pipeline).
- **2026-05-13:** v0.4.3 hotfix (absolute sensor hook paths).
- **2026-05-07:** v0.4.1 shipped. `bonsai completion` command merged (PR #78 — first external contribution).

**Residual items from 2026-05-04 run:**
- Root `Bonsai/CLAUDE.md` project-structure tree was flagged "badly stale" — partially addressed on 2026-05-07 (Go 1.24+ → 1.25+) but structural drift (missing `completion.go`) persists.
- `agent/Skills/bonsai-model.md` broken nav link — link now resolves (file exists), resolved.
- `code-index.md` stale — partially addressed (Plan 35 validate + Plan 32 wsvalidate added on 2026-05-07), but Plan 41 headless content still missing.
- `INDEX.md` CLI count 7→8 — was updated on 2026-05-07, but is now stale again (8→9 with `completion`).

### Step 2 — Check INDEX.md accuracy

- **Tech stack table:** Accurate. Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS — all verified.
- **Agent types count:** 6 — verified (backend, devops, frontend, fullstack, security, tech-lead). Accurate.
- **Catalog item count:** Says "~50". Actual count: Skills 18 + Workflows 10 + Protocols 4 + Sensors 13 + Routines 8 = **53**. Mildly stale; "~50" qualifier softens it.
- **CLI commands count:** Says "8 (init, add, remove, list, catalog, update, guide, validate)". Actual: `completion.go` exists, making it **9**. Stale.
- **Architecture diagram:** Accurate for overall flow. Does not mention the headless cores / `docs/agent-interface.md` contract — worth a note for discoverability.

### Step 3 — Check navigation links

**station/CLAUDE.md nav table verified:**

- Core files (identity, memory, self-awareness): all resolve ✓
- Protocols (memory, scope-boundaries, security, session-start): all resolve ✓
- Workflows listed: all 9 listed entries resolve ✓ — BUT `plan-grilling.md` exists in `agent/Workflows/` and is **not listed**.
- Skills listed: all 6 listed entries resolve ✓ — BUT `critic-agent-prompts.md` exists in `agent/Skills/` and is **not listed**.
- Routines listed: all 7 listed entries resolve ✓
- Sensors listed: all 10 listed entries resolve (not individually verified via stat, but directory list confirms all named files present) ✓

**agent/Core/, agent/Protocols/, agent/Workflows/, agent/Skills/ cross-check:** No broken links found in listed entries.

### Step 4 — Report findings

See Findings Summary below. Per procedure: findings flagged; no doc edits executed.

### Step 5 — Update dashboard

Dashboard updated in `station/agent/Core/routines.md`.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `plan-grilling.md` workflow exists in `agent/Workflows/` but is not listed in the CLAUDE.md Workflows navigation table. Added 2026-06-13. The tech-lead agent won't know this workflow exists. | `station/CLAUDE.md` — Workflows table | Flagged for user. Proposed fix: add row `\| Adversarial review of a drafted plan via 6 critics before dispatch \| [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) \|` under "Reviewing a pull request…" entry. |
| 2 | MEDIUM | `critic-agent-prompts.md` skill exists in `agent/Skills/` but is not listed in the CLAUDE.md Skills navigation table. Added 2026-06-13. | `station/CLAUDE.md` — Skills table | Flagged for user. Proposed fix: add row `\| Dispatching the 6 plan-grilling critic agents; getting critic prompt templates \| [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) \|`. |
| 3 | MEDIUM | `INDEX.md` CLI command count says "8" but `completion.go` exists — 9 commands total. `bonsai completion` shipped via PR #78 (2026-05-07). | `station/INDEX.md` — Key Metrics | Flagged for user. Fix: change `8 (init, add, remove, list, catalog, update, guide, validate)` → `9 (init, add, remove, list, catalog, update, guide, validate, completion)`. |
| 4 | MEDIUM | `code-index.md` CLI Commands table missing `bonsai completion` (added 2026-05-07). Only 8 commands listed. | `station/code-index.md` — CLI Commands table | Flagged for user. Fix: add row `\| \`bonsai completion\` \| \`cmd/completion.go\` \| shell completion for bash/zsh/fish/powershell \|`. |
| 5 | MEDIUM | `code-index.md` missing Plan 41 headless/non-interactive functions (2026-06-16): `runAddNonInteractive`, `runInitNonInteractive`, `runRemoveNonInteractive`, `runUpdateNonInteractive`, `list --json` flag, and `ExitConflict=5` exit code contract. These are significant public-API additions. | `station/code-index.md` — Add/Remove/Update/List Helpers | Flagged for user. Requires adding a "Headless / Non-Interactive Cores" subsection to code-index.md CLI Commands. |
| 6 | LOW | `INDEX.md` catalog item count says "~50". Actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). | `station/INDEX.md` — Key Metrics | Flagged for user. Minor drift; the "~" qualifier softens it. Fix: update to "~53" or exact "53". |
| 7 | LOW | `docs/agent-interface.md` (the headless CLI contract, Plan 41) exists at repo root but is not cross-referenced anywhere in station documentation — not in INDEX.md Document Registry or CLAUDE.md nav. | `station/INDEX.md` — Document Registry | Flagged for user. Fix: add row to Document Registry: `\| \`docs/agent-interface.md\` \| Headless CLI contract — \`\*Result\` cores, JSONL/exit codes, MCP integration spec \| When building MCP server (Plan 42) or integrating headless Bonsai calls \|`. |
| 8 | LOW | Root `/home/user/Bonsai/CLAUDE.md` project structure tree missing `completion.go` in the `cmd/` listing. Last updated 2026-05-07. Plan 41 headless test files (`add_nonint_test.go`, etc.) are also not listed (though test files are commonly omitted from project trees). | `/home/user/Bonsai/CLAUDE.md` — Project Structure | Flagged for user. Fix: add `│   ├── completion.go        ← bonsai completion (shell completions)` after `validate.go` in the cmd/ listing. |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

All 8 findings above require user action — none were auto-resolved because the procedure specifies "propose but don't execute" for doc content updates. In priority order:

1. **[HIGH] Add `plan-grilling.md` to CLAUDE.md Workflows table** — agent will miss this workflow without it.
2. **[MEDIUM] Add `critic-agent-prompts.md` to CLAUDE.md Skills table** — paired with finding #1.
3. **[MEDIUM] Fix INDEX.md CLI count (8 → 9)** — quick one-liner.
4. **[MEDIUM] Add `bonsai completion` to code-index.md** — quick one-liner.
5. **[MEDIUM] Add Plan 41 headless cores to code-index.md** — requires a new subsection; moderate effort.
6. **[LOW] Update INDEX.md catalog count (~50 → ~53)** — cosmetic, quick.
7. **[LOW] Add `docs/agent-interface.md` to INDEX.md Document Registry** — useful for Plan 42 discoverability.
8. **[LOW] Add `completion.go` to root CLAUDE.md project structure tree** — cosmetic housekeeping.

---

## Notes for Next Run

- The two MEDIUM findings about code-index.md (Plan 41 headless cores) may grow if Plan 42 (MCP server) ships before the next run — watch for new `cmd/mcp.go` or `internal/mcp/` entries.
- If the user resolves finding #5 (headless cores in code-index), also check that the `ExitConflict=5` / exit-code contract table is captured.
- `docs/agent-interface.md` is the headless contract spec — once Plan 42 lands, this doc may change; check it next run.
- No broken nav links found this cycle — clean baseline for link hygiene.
