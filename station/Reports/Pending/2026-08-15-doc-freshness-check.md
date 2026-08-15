---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-15
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
- **Files Read:** 12 — `station/agent/Routines/doc-freshness-check.md`, `station/CLAUDE.md`, `station/INDEX.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `CLAUDE.md` (root), `embed.go`, `station/Reports/Pending/` (listing), `station/Playbook/Plans/Archive/` (listing)
- **Files Modified:** 2 — `station/agent/Core/routines.md` (dashboard update), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls), Write, Edit
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1: Scan project documentation
- **Action:** Read station/INDEX.md, station/Playbook/Status.md, station/Playbook/Roadmap.md, station/CLAUDE.md, root CLAUDE.md. Ran `git log --oneline --since="2026-05-04"` to identify all commits since the last run.
- **Result:** 50+ commits since 2026-05-04. Key shipped work: Plan 40 (Odysseus platform integration, v0.5.0 untagged), Plan 41 (Headless CLI Contract + MCP-ready cores, PRs #120–#125), v0.4.2 (non-interactive flags, PR #102), v0.4.3 (sensor hook absolute paths, PRs #105–#106), bonsai completion command (PR #78, first external contribution). Identified 3 documentation drift items (see Findings Summary).
- **Issues:** 3 medium/low severity documentation drift items found (details below).

### Step 2: Check INDEX.md accuracy
- **Action:** Read station/INDEX.md. Checked tech stack, Key Metrics (agent types, catalog items, CLI commands), and architecture overview against actual codebase state. Verified catalog item counts via `ls catalog/{skills,workflows,protocols,sensors,routines}/`.
- **Result:**
  - **Tech Stack table:** Accurate — Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, single binary. No drift.
  - **Agent types (6):** Accurate — backend, devops, frontend, fullstack, security, tech-lead. No drift.
  - **Catalog items (~50):** Actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). "~50" is an acceptable approximation; low priority to update.
  - **CLI commands (8):** **DRIFT** — the `completion` command (added v0.4.1, PR #78) makes the true count 9. The listed names also omit `completion`.
  - **Architecture diagram:** Accurate — `internal/validate/` and `internal/wsvalidate/` both listed correctly.
  - **Document Registry:** **DRIFT** — no entry for `docs/agent-interface.md` (the Plan 41 CLI contract published at PR #125), or the broader `docs/` directory.
- **Issues:** 2 drift items (CLI count, docs/ missing from Document Registry).

### Step 3: Check navigation links
- **Action:** Verified all links in station/CLAUDE.md navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References). Checked that all referenced files exist on disk.
- **Result:** All navigation links in station/CLAUDE.md resolve to real files. Specifically:
  - Core (3/3 files exist): identity.md, memory.md, self-awareness.md
  - Protocols (4/4 files exist): memory.md, scope-boundaries.md, security.md, session-start.md
  - Workflows (9/9 files exist): code-review.md, planning.md, pr-review.md, security-audit.md, session-logging.md, test-plan.md, session-wrapup.md, issue-to-implementation.md, routine-digest.md
  - Skills (6/6 files exist): planning-template.md, review-checklist.md, issue-classification.md, pr-creation.md, bubbletea.md, bonsai-model.md
  - Routines (7/7 files exist): all present
  - External references: `station/INDEX.md`, `Playbook/Status.md`, `Playbook/Roadmap.md`, `Playbook/Standards/SecurityStandards.md`, `Playbook/Backlog.md`, `Logs/KeyDecisionLog.md`, `Reports/Pending/` — all valid
  - **Previously-flagged `agent/Skills/bonsai-model.md` broken nav link:** RESOLVED — file exists. Cleared.
  - **DRIFT in root CLAUDE.md:** `cmd/` tree listing missing `completion.go`. `docs/` directory missing entirely from the project structure tree.
  - **DRIFT in station/code-index.md:** embed.go entry says "GuideCustomFiles, GuideQuickstart, GuideConcepts, GuideCli" at `:12–21`; file now also has `GuideFormats` at `:23`. CLI Commands table missing `bonsai completion` row.
- **Issues:** 3 drift items across root CLAUDE.md and code-index.md.

### Step 4: Report findings
- **Action:** Compiled 3 drift findings across 4 files. Per routine procedure, all doc updates are flagged for user decision — not executed autonomously.
- **Result:** Findings documented below with specific locations and proposed updates.
- **Issues:** None (audit-only step executed cleanly).

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for "Doc Freshness Check" — Last Ran → 2026-08-15, Next Due → 2026-08-22, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `bonsai completion` command (added v0.4.1 PR #78) not documented — missing from root CLAUDE.md cmd/ tree, INDEX.md CLI count ("8" should be "9"), and code-index.md CLI Commands table | `/home/user/Bonsai/CLAUDE.md`, `station/INDEX.md`, `station/code-index.md` | Flagged for user |
| 2 | Medium | `docs/` directory not represented in any project doc — no entry in INDEX.md Document Registry, not in root CLAUDE.md project structure tree. `docs/agent-interface.md` (the Plan 41 CLI contract) is a key reference not surfaced to agents | `/home/user/Bonsai/CLAUDE.md`, `station/INDEX.md` | Flagged for user |
| 3 | Low | code-index.md embed.go entry stale — lists 4 guide vars at `:12–21`; `GuideFormats` was added at `:23` in Plan 40 Phase 3 (docs/formats.md) | `station/code-index.md` | Flagged for user |

---

## Proposed Updates (for user decision)

### Finding 1 — Add `completion` command

**Root CLAUDE.md** `cmd/` tree — add after `validate.go` line:
```
│   └── completion.go    ← bonsai completion — shell completion scripts (bash/zsh/fish/powershell)
```

**station/INDEX.md** Key Metrics row:
```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```

**station/code-index.md** CLI Commands table — add row:
```
| `bonsai completion` | `cmd/completion.go` | Shell completion scripts (bash/zsh/fish/powershell) |
```

### Finding 2 — Document `docs/` directory

**Root CLAUDE.md** project structure tree — add `docs/` block before or after `station/`:
```
├── docs/                 ← embedded documentation (ships in binary + published on website)
│   ├── agent-interface.md  ← Plan 41 CLI contract — headless core API, JSONL format, exit codes
│   ├── cli.md              ← CLI reference
│   ├── concepts.md         ← core concepts
│   ├── custom-files.md     ← custom files guide
│   ├── formats.md          ← output format guide (Plan 40 Phase 3)
│   └── quickstart.md       ← quickstart guide
```

**station/INDEX.md** Document Registry — add row:
```
| `docs/agent-interface.md` | Plan 41 CLI contract — headless mutating-command API, JSONL event format, exit codes (0/2/3/4/5) | When building MCP server or any non-interactive Bonsai integration |
```

**station/CLAUDE.md** External References table — add row (optional, low priority):
```
| Headless CLI contract | [docs/agent-interface.md](../docs/agent-interface.md) |
```

### Finding 3 — Fix embed.go entry in code-index.md

**station/code-index.md** Entry Point table row:
```
| Embed guide cheatsheets | `embed.go:11–25` — `GuideCustomFiles`, `GuideQuickstart`, `GuideConcepts`, `GuideCli`, `GuideFormats` |
```

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **Finding 1 (Medium):** 3 files need `bonsai completion` command added — root CLAUDE.md, INDEX.md, code-index.md. Easy 1-line fixes each.
2. **Finding 2 (Medium):** `docs/` directory and `docs/agent-interface.md` (Plan 41 contract) not surfaced in any project doc. Recommend adding `docs/` block to root CLAUDE.md structure tree and `docs/agent-interface.md` row to INDEX.md Document Registry at minimum.
3. **Finding 3 (Low):** code-index.md embed.go entry missing `GuideFormats` — update once, low urgency.

**Cleared from previous run:** `agent/Skills/bonsai-model.md` broken nav link (2026-05-04 flag) — file exists, resolved.

---

## Notes for Next Run

- Root CLAUDE.md project structure drift has been a recurring pattern across multiple doc-freshness runs. After user resolves findings 1 and 2, a single clean confirmation run next cycle should close this pattern.
- `docs/agent-interface.md` is expected to grow when Plan 42 (MCP server) ships — next doc-freshness run should check if contract has been updated there.
- Catalog item count is 53; INDEX.md says "~50". Acceptable approximation; leave unless user wants an exact count.
- All station/CLAUDE.md navigation links resolve cleanly — no dead links for the first time in several cycles.
