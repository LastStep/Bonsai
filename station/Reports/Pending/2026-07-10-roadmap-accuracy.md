---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-07-10
status: success
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/roadmap-accuracy.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log append)
- **Tools Used:** Read, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `Roadmap.md` and checked all Phase 1 items against current codebase and Status.md.

**Phase 1 — Foundation & Polish:** All 11 items correctly marked `[x]`. Verified:
- Go rewrite, Full catalog, Lock file handling, Awareness Framework — all confirmed in CLAUDE.md project structure and KeyDecisionLog.
- Dogfooding — station/ workspace in active use.
- Better trigger sections — fixed by 2026-05-07 Routine Digest (quick-fixed to `[x]` with annotation).
- UI overhaul, Usage instructions — fixed by 2026-04-21 Routine Digest.
- Release pipeline — GoReleaser + GitHub Actions + Homebrew Tap confirmed in `.goreleaser.yaml` and `release.yml`.
- Community health files — confirmed shipped.
- `bonsai validate` — added to Phase 1 by 2026-05-07 Routine Digest. Confirmed in `cmd/validate.go` and `internal/validate/`.

Phase 1 is complete and accurately reflected. No issues.

**Alignment with Status.md:** Status.md "Current Phase" section = In Progress: none / Pending: sentrux trial (blocked). Most recent done items are Plan 41 (2026-06-16) and Plan 40 Phases 1–3 (2026-06-13). The roadmap does not yet reflect these major shipped items (see Step 2 findings).

### Step 2 — Check milestone accuracy

Reviewed Phase 2, 3, and 4 items against recent Status.md work.

**Phase 2 — Extensibility:**
- `[x] Custom item detection` — ACCURATE. `internal/generate/scan.go` exists for custom file discovery.
- `[ ] Self-update mechanism` — Still pending. No evidence of work started.
- `[ ] Template variables expansion` — Still pending. Backlog Hygiene (2026-07-10) flagged this has no Backlog tracking entry.
- `[ ] Micro-task fast path` — Still pending.

**GAP — Plan 41 (Headless CLI Contract):** Status.md shows Plan 41 SHIPPED 2026-06-16. This is a significant extensibility milestone: every mutating command (init/add/update/remove) now has a pure `*Result` headless core + JSONL/exit contract (`ExitConflict=5`); `list --json`; and `docs/agent-interface.md` contract doc. This is directly in Phase 2 scope (extensibility / machine-readable output) but has NO roadmap entry. It should appear as `[x]` in Phase 2.

**GAP — Plan 42 (MCP server):** Status.md explicitly states "MCP server = fast-follow Plan 42" as the next step after Plan 41. This is not on the roadmap at all. It belongs in Phase 2 (or Phase 3) as `[ ]`.

**GAP — Shell completions (`bonsai completion`):** Shipped via external contribution PR #78 on 2026-05-07. Not on the roadmap. This is a Phase 1-era capability that's missing from the completed phase checklist. Lower severity since Phase 1 is closed, but worth noting.

**GAP — `--non-interactive --from-config` flag (v0.4.2):** Shipped 2026-05-13. Enables programmatic non-interactive init/add with JSONL stdout. Closely related to Plan 41's headless contract theme. Not on the roadmap. Also Phase 1-era technically since it shipped before Phase 2 work began.

**Phase 3 — Cloud & Orchestration:** Items unchanged (both `[ ]`). The KeyDecisionLog "Settled" section confirms Managed Agents integration is intentionally deferred. Consistent.

**Phase 4 — Ecosystem:** Items unchanged (all `[ ]`). Appropriate — no work started.

### Step 3 — Cross-check against Key Decision Log

Read `KeyDecisionLog.md`. Checked all decisions against roadmap items.

- Go rewrite, embed.FS, text/template, lock file — all reflected in Phase 1 `[x]` items.
- Tech-lead as required agent — not a roadmap item, correctly absent (it's an architectural constant, not a milestone).
- Two-sensor awareness design — reflected as `[x] Awareness Framework` in Phase 1.
- Six agent types — reflected as `[x] Full catalog` in Phase 1.
- "Defer Managed Agents until local foundation stable" (Settled) — Phase 3 items still `[ ]`, CONSISTENT.

No KeyDecisionLog decisions invalidate any current roadmap items. The headless CLI approach (Plan 41) is NOT yet recorded in KeyDecisionLog — that's a separate gap but not this routine's domain.

### Step 4 — Report findings (no direct Roadmap.md edits per procedure)

Findings documented below. Roadmap.md left unmodified — flagged for user review.

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` dashboard row for Roadmap Accuracy.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | Phase 2 missing `[x]` entry for Plan 41 — Headless CLI Contract (headless `*Result` cores + JSONL/exit contract `ExitConflict=5` + `list --json` + `docs/agent-interface.md`). Shipped 2026-06-16. Major extensibility milestone with no roadmap representation. | `Roadmap.md`, Phase 2 | Flagged for user review. Suggested addition: `- [x] Headless CLI contract — machine-readable output (JSONL/exit codes), `*Result` cores for all mutating commands, `list --json`, agent-interface.md contract doc` |
| 2 | Low | Phase 2 (or 3) missing `[ ]` entry for Plan 42 — MCP server. Status.md explicitly names it as "fast-follow" to Plan 41. No planning yet but directionally committed. | `Roadmap.md`, Phase 2 or 3 | Flagged for user review. Suggested addition to Phase 2: `- [ ] MCP server — exposes headless CLI contract as an MCP tool server (Plan 42)` |
| 3 | Low | Phase 1 missing entry for `bonsai completion` (shell completions: bash/zsh/fish/powershell). Shipped via external contribution PR #78 on 2026-05-07. Phase 1 is already closed/complete so impact is cosmetic. | `Roadmap.md`, Phase 1 | Flagged for user review. Optional retroactive addition. |
| 4 | Low | Phase 1 missing entry for `--non-interactive --from-config` flag (v0.4.2 — JSONL stdout, hard-skip conflicts, exit codes 0/2/3/4). Shipped 2026-05-13. Predates headless contract but is thematically related. Phase 1 already closed. | `Roadmap.md`, Phase 1 | Flagged for user review. Optional retroactive addition, or fold into Finding #1 narrative. |
| 5 | Info | Phase 2 `[ ] Template variables expansion` has no Backlog tracking entry (confirmed by 2026-07-10 Backlog Hygiene run). Risk of being forgotten. | `Roadmap.md`, Phase 2 | Flagged for user review. Backlog Hygiene routine already noted this. |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

### Flag 1 (Medium) — Add Plan 41 to Phase 2 roadmap as `[x]`

The Headless CLI Contract (Plan 41) is the most significant shipped work not reflected in the roadmap. Suggested addition to Phase 2:

```
- [x] Headless CLI contract — machine-readable output (JSONL/exit codes), pure `*Result` cores for all mutating commands, `list --json`, `docs/agent-interface.md` contract doc
```

### Flag 2 (Low) — Add MCP server to Phase 2 roadmap as `[ ]`

Plan 42 is named as "fast-follow" in Status.md. Suggest placing in Phase 2:

```
- [ ] MCP server — exposes headless CLI contract as MCP tool server for IDE/agent integration (Plan 42)
```

Or in Phase 3 if you consider it cloud/orchestration work. User judgment call.

### Flag 3 (Low) — Optional Phase 1 retroactive entries

- `bonsai completion` (shell completions — PR #78, 2026-05-07)
- `--non-interactive --from-config` (v0.4.2 programmatic mode)

Both are shipped capabilities that could be added retroactively to Phase 1 for historical completeness. Phase 1 is otherwise clean.

### Flag 4 (Info) — Phase 2 `Template variables expansion` needs a Backlog entry

Per Backlog Hygiene (2026-07-10), this roadmap item has no Backlog tracking entry. Add a Backlog P2 row to avoid losing it.

---

## Notes for Next Run

- Phase 1 is fully accurate — no re-checking needed.
- Next run should focus on: (a) whether Plan 41/42 have been added to the roadmap; (b) whether Phase 2 remaining items (`self-update`, `template vars`, `micro-task`) have any work started.
- Phase 2 `[x] Custom item detection` — verify `bonsai validate`'s `--agent` + untracked-custom scanning mode still matches the roadmap description.
- If Phase 4 HELD of Plan 40 (Odysseus integration) ever resumes, it may affect Phase 2 roadmap items around update-delivery.
