---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-27
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
- **Files Read:** 10 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/CLAUDE.md` (system-reminder), `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, ls commands)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history

Ran `git log --oneline --since="7 days ago" --name-only`. Found 2 commits in the last 7 days:

- `f9a682d` — routine: backlog-hygiene complete — touched `station/Logs/RoutineLog.md`, `station/Reports/Pending/2026-07-27-backlog-hygiene.md`, `station/agent/Core/routines.md`
- `25db10f` — routine: backlog-hygiene partial — touched `station/Playbook/Backlog.md`

Both commits are routine maintenance only. No new features, services, or configuration were added to the codebase in the last 7 days. No code-to-doc drift from recent work.

Expanded the scan to the broader git log to look for pre-existing drift. Identified 3 findings (see below).

### Step 2 — Check INDEX.md accuracy

Read `station/INDEX.md`. Checked tech stack, architecture overview, key metrics, and document registry against current codebase state.

- **Tech stack** — accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template).
- **Architecture overview** — accurate; references all internal packages correctly including `internal/wsvalidate/` and `internal/validate/`.
- **Agent types: 6** — verified against `catalog/agents/`: backend, devops, frontend, fullstack, security, tech-lead. Accurate.
- **Catalog items: ~50** — actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). Approximation close enough; no action required.
- **CLI commands: 8** — `cmd/` directory reveals `completion.go` in addition to the 8 listed. `bonsai completion` was shipped 2026-05-07 (Status.md). **Count is stale — should be 9.** (Finding #1)
- **Document registry** — `docs/agent-interface.md` (shipped with Plan 41, June 2026) is not listed. (Finding #2)

### Step 3 — Check navigation links

Verified all links in `station/CLAUDE.md` navigation tables against actual filesystem:

- **agent/Core/**: identity.md, memory.md, routines.md, self-awareness.md — all present ✓
- **agent/Protocols/**: memory.md, scope-boundaries.md, security.md, session-start.md — all present ✓
- **agent/Workflows/**: code-review.md, issue-to-implementation.md, planning.md, pr-review.md, routine-digest.md, security-audit.md, session-logging.md, session-wrapup.md, test-plan.md — all present ✓
- **agent/Skills/**: bonsai-model.md, bubbletea.md, critic-agent-prompts.md, issue-classification.md, planning-template.md, pr-creation.md, review-checklist.md — all present ✓
- **agent/Sensors/**: agent-review.sh, compact-recovery.sh, context-guard.sh, dispatch-guard.sh, routine-check.sh, scope-guard-files.sh, session-context.sh, status-bar.sh, statusline.sh, subagent-stop-review.sh — all present ✓
- **agent/Routines/**: all 7 routine files present ✓
- **Playbook/**: Backlog.md, Roadmap.md, Status.md, Standards/SecurityStandards.md — all present ✓
- **Playbook/Plans/Active/**: 40-odysseus-platform-integration.md, 41-headless-cli-contract.md — both present ✓
- **Logs/**: FieldNotes.md, KeyDecisionLog.md, RoutineLog.md — all present ✓
- **Reports/Pending/** — directory present ✓
- **.bonsai/catalog.json** — present ✓
- **.bonsai.yaml** — present ✓
- **code-index.md** — present ✓

No broken links found. All navigation targets resolve.

Also noted: `code-index.md` CLI command table lists 8 commands; `bonsai completion` is absent. (Finding #3 — same underlying drift as Finding #1)

### Step 4 — Report findings

3 findings identified, all pre-existing drift (not from last 7 days). None are blockers. All flagged for user decision below.

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` dashboard row for "Doc Freshness Check": Last Ran → 2026-07-27, Next Due → 2026-08-03, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Low | INDEX.md CLI command count is stale: says "8 (init, add, remove, list, catalog, update, guide, validate)" — `bonsai completion` shipped 2026-05-07 makes it 9 | `station/INDEX.md` Key Metrics table | Flagged — propose update to "9" |
| 2 | Low | `docs/agent-interface.md` (Plan 41 headless CLI contract doc) absent from INDEX.md document registry | `station/INDEX.md` Document Registry | Flagged — propose adding row |
| 3 | Low | `code-index.md` CLI Commands table missing `bonsai completion` row | `station/code-index.md` | Flagged — propose adding row |

## Errors & Warnings

None.

## Items Flagged for User Review

### [LOW] Finding #1 — INDEX.md CLI command count stale

**Where:** `station/INDEX.md`, Key Metrics table, "CLI commands" row.

**Current:** `8 (init, add, remove, list, catalog, update, guide, validate)`

**Proposed:** `9 (init, add, remove, list, catalog, update, guide, validate, completion)`

`bonsai completion [bash|zsh|fish|powershell]` was merged 2026-05-07 from @mvanhorn (PR #78). The file `cmd/completion.go` exists.

---

### [LOW] Finding #2 — `docs/agent-interface.md` missing from INDEX.md document registry

**Where:** `station/INDEX.md`, Document Registry table.

**What:** Plan 41 shipped `docs/agent-interface.md` as the headless CLI / MCP-ready contract document (June 2026). This is a key reference for anyone building automation on top of the Bonsai CLI, but it is not listed in the INDEX.md document registry.

**Proposed row to add:**

```
| `docs/agent-interface.md` | Headless CLI contract — JSONL event schema, exit codes, `--non-interactive` flag spec (Plan 41) | When building automation or MCP tooling on top of Bonsai CLI |
```

---

### [LOW] Finding #3 — `code-index.md` CLI Commands table missing `bonsai completion`

**Where:** `station/code-index.md`, "CLI Commands" table.

**Proposed row to add** (after `bonsai validate` row):

```
| `bonsai completion` | `cmd/completion.go` | `completionCmd` — shell completion for bash/zsh/fish/powershell |
```

---

### [INFO] Plan 40 still in Plans/Active/

**Where:** `station/Playbook/Plans/Active/40-odysseus-platform-integration.md`

Status.md notes "Phase 4 HELD" for Plan 40 (Odysseus Platform Integration). The plan file remains in `Plans/Active/`. This may be intentional (keeping it active until Phase 4 resumes), but worth confirming — if the plan is indefinitely deferred, consider moving it to `Plans/Archive/` to keep Active/ clean.

## Notes for Next Run

- Pre-existing drift (Findings #1–3) will reappear unless actioned. Recommend addressing them in the next doc-touch session.
- No new features or code changes from the past 7 days require documentation updates.
- All navigation links healthy — no link rot detected.
- `docs/` directory contains several docs (`cli.md`, `concepts.md`, `custom-files.md`, `formats.md`, `quickstart.md`) — not individually tracked in INDEX.md (likely by design, as they are website-facing docs). No action needed.
