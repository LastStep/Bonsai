---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-03
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
- **Files Read:** 12 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`, `CLAUDE.md` (root), `station/CLAUDE.md`, `station/agent/Workflows/plan-grilling.md`, `station/agent/Skills/critic-agent-prompts.md`, `internal/nonint/nonint.go`
- **Files Modified:** 3 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry), `station/Reports/Pending/2026-07-03-doc-freshness-check.md` (this report)
- **Tools Used:** Read, Bash (git log, ls, head), Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation vs recent git history

**7-day git window (2026-06-26 to 2026-07-03):** Only 2 commits — both routine maintenance runs (status-hygiene and backlog-hygiene). No code changes.

**Broader context (last 30 days):** Two major plans shipped since the last doc-freshness-check (2026-05-04):
- **Plan 40 (2026-06-13)** — Odysseus platform integration: frozen schemas + root-relative scaffolding + validate pass + docs updates. Phases 1–3 merged; Phase 4 held.
- **Plan 41 (2026-06-16)** — Headless CLI Contract + MCP-ready cores: all 5 phases merged. New `internal/nonint/` package created; headless cores added to init/add/update/remove; `list --json`; `docs/agent-interface.md` contract doc. Exit codes (ExitConflict=5) + JSONL/exit contract.

**New code since last doc run:**
- `internal/nonint/` package (12 files: config.go, config_test.go, contract_test.go, events.go, events_test.go, nonint.go, remove.go, remove_test.go, result.go, result_test.go, runner.go, runner_test.go + testdata/)
- `cmd/add_nonint_test.go`, `cmd/init_nonint_test.go`, `cmd/remove_nonint_test.go`, `cmd/update_nonint_test.go` — new non-interactive test files
- `docs/agent-interface.md` — headless contract documentation
- `station/agent/Workflows/plan-grilling.md` — adversarial plan review workflow
- `station/agent/Skills/critic-agent-prompts.md` — critic agent prompt templates

### Step 2 — Check INDEX.md accuracy

- **Tech stack table:** Accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS all correct.
- **Key metrics:**
  - Agent types: 6 — verified accurate.
  - Catalog items: ~50 — rough count, plausible.
  - CLI commands: 8 — actually 9 now (includes `bonsai completion` from PR #78). Count is stale by 1.
- **Architecture diagram:** Missing `internal/nonint/` package (Plan 41 deliverable). All 6 listed packages exist, but nonint is undocumented.
- **Document Registry:** Accurate — all referenced files exist. External `LastStep/Bonsai-Eval` link preserved.

### Step 3 — Check navigation links

**station/CLAUDE.md** — all linked files verified:
- Core (4 files): all exist ✓
- Protocols (4 files): all exist ✓
- Workflows (9 listed): all exist ✓ — `plan-grilling.md` exists but not listed (see Finding #5)
- Skills (6 listed): all exist ✓ — `critic-agent-prompts.md` exists but not listed (see Finding #5)
- Routines (7 listed): all exist ✓
- Sensors (10 listed): all exist ✓
- External: `../.bonsai/catalog.json` ✓, `../.bonsai.yaml` ✓

**No broken links found in station/CLAUDE.md.**

**agent/Core/, Protocols/, Workflows/, Skills/:** All files listed in station/CLAUDE.md nav tables exist on disk.

**Root Bonsai/CLAUDE.md:** Project structure tree missing `internal/nonint/` package and `cmd/completion.go` — recurring drift since Plan 41 shipped and completion command was added (PR #78, 2026-04-22).

### Step 4 — Report findings

See Findings Summary below. This is audit-only — no doc edits made. All items flagged for user decision.

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` dashboard row for Doc Freshness Check.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | `code-index.md` missing entire `internal/nonint/` package section — Plan 41 (Jun 2026) introduced 12 source files + testdata, none documented | `station/code-index.md` | Flagged — propose new section between wsvalidate and TUI |
| 2 | MEDIUM | Root `Bonsai/CLAUDE.md` project structure tree missing `internal/nonint/` package (Plan 41) — recurring drift from previous cycles, now compounded | `CLAUDE.md` (root) | Flagged — propose adding nonint block under wsvalidate |
| 3 | MEDIUM | `station/INDEX.md` architecture diagram missing `internal/nonint/` — headless CLI contract package not reflected anywhere in narrative docs | `station/INDEX.md` | Flagged — propose 1-line addition after wsvalidate row |
| 4 | LOW | `station/INDEX.md` CLI command count stale: says 8, actually 9 (completion added via PR #78) | `station/INDEX.md` | Flagged — update count from 8 → 9 |
| 5 | LOW | `station/agent/Workflows/plan-grilling.md` and `station/agent/Skills/critic-agent-prompts.md` exist on disk but are not in the station/CLAUDE.md nav tables — their frontmatter notes "Bonsai-catalog integration pending (Backlog)" | `station/CLAUDE.md` | Flagged — user to decide: add to nav now or track in Backlog until catalog integration |
| 6 | LOW | Plans 40 and 41 both remain in `Plans/Active/` — Plan 41 is fully shipped (all 5 phases, Status.md confirms); Plan 40 is partially shipped (Phase 4 held, legitimately active) | `station/Playbook/Plans/Active/` | Flagged — status-hygiene routine already noted Plan 41 archival; action deferred to user |

**Previously recurring finding now resolved:** `agent/Skills/bonsai-model.md` broken link (flagged 2026-05-07 backlog hygiene) — file exists, link works ✓

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **[USER ACTION] `code-index.md` — add `internal/nonint/` section** (Finding #1, HIGH). Proposed content:
   ```
   ## Headless Contract (`internal/nonint/`) — Plan 41
   
   Pure headless cores for init/add/update/remove + shared Result/Event contract.
   No TUI imports — safe for MCP/CI invocation. Entry via runner.go.
   ```
   Full section to document: `nonint.go`, `config.go`, `runner.go`, `events.go`, `result.go`, `remove.go`, `update.go`.

2. **[USER ACTION] Root `CLAUDE.md` — add `internal/nonint/` to project structure tree** (Finding #2, MEDIUM). One block to add after wsvalidate:
   ```
   │   ├── nonint/
   │   │   └── nonint.go + config/events/result/runner/remove/update.go  ← headless CLI contract + MCP-ready cores (Plan 41)
   ```
   Also add `cmd/completion.go` to the cmd/ block.

3. **[USER ACTION] `station/INDEX.md` — add `internal/nonint/` to arch diagram** (Finding #3, MEDIUM). One line after wsvalidate row:
   ```
   internal/nonint/      ← headless CLI contract cores — pure Result/Event shapes, no TUI
   ```

4. **[USER ACTION] `station/INDEX.md` — fix CLI command count** (Finding #4, LOW). Change "8" → "9" in Key Metrics table (completion command shipped PR #78).

5. **[USER DECISION] `plan-grilling` + `critic-agent-prompts` nav listing** (Finding #5, LOW). Both files exist in agent dirs but are unlisted in station/CLAUDE.md. Frontmatter says integration is pending. Options: (a) add rows to nav now so agent can discover them, (b) leave unlisted until Backlog item is addressed.

6. **[USER ACTION] Archive Plan 41** (Finding #6, LOW). Move `station/Playbook/Plans/Active/41-headless-cli-contract.md` to `Plans/Archive/`. (Plan 40 can remain Active — Phase 4 still held.)

---

## Notes for Next Run

- Root `Bonsai/CLAUDE.md` drift is a recurring issue across multiple cycles. Consider adding an explicit sub-step in the routine procedure to check root CLAUDE.md cmd/ + internal/ blocks against `ls` output — was noted as a Backlog item in 2026-04-21 cycle.
- `code-index.md` drifts whenever a new internal package is added. The nonint package is the first new package since validate (Plan 35). Consider whether code-index.md maintenance should be part of the post-plan checklist.
- No broken navigation links found — workspace nav is healthy.
- All 7 git commits since last routine run (2026-05-04) are doc/routine entries — no unreviewed code changes sitting undocumented (Plan 40+41 were already noted in Status.md).
