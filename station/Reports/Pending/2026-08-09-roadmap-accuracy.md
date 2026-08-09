---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-08-09
status: success
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~10 min
- **Files Read:** 7 — `/home/user/Bonsai/station/agent/Routines/roadmap-accuracy.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Logs/KeyDecisionLog.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/41-headless-cli-contract.md`, `/home/user/Bonsai/station/Playbook/Plans/Active/40-odysseus-platform-integration.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Write, Edit, Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Compare Roadmap against current state

Read `Playbook/Roadmap.md` and cross-checked all phases against recent work in `Status.md`.

**Phase 1 — Foundation & Polish:** All 11 items are marked `[x]` complete. This matches the actual build state. The two flags from the 2026-05-07 run (marking "Better trigger sections" as done, adding `bonsai validate`) have both been resolved in the roadmap — Phase 1 is accurate.

**Phase 2 — Extensibility:** Only `[x] Custom item detection` is marked shipped. Three significant features have shipped since the last run that are NOT reflected:

1. **v0.4.2 — non-interactive CLI flags** (`--non-interactive --from-config` for `init`/`add`, JSONL stdout, exit codes, Plan 39, 2026-05-13)
2. **Plan 40 Phases 1–3 — platform schemas** (`.bonsai/project.yaml` manifest, frozen memory-note schema v1, root-relative scaffolding, `validate` lint for both, `bonsai guide` Formats page, shipped 2026-06-13, untagged)
3. **Plan 41 — Headless CLI Contract** (all four mutating commands have `*Result` headless cores + JSONL/exit contract with `ExitConflict=5`; `list --json`; `docs/agent-interface.md` contract doc; shipped 2026-06-16, PRs #120/#122/#123/#121/#125)

Plan 41 is the most significant omission — it's the MCP-readiness milestone and represents a complete agent-drivable surface. An AI agent (or CI script) can now drive the full Bonsai lifecycle via stdout/exit-codes with no TTY. This isn't a minor quality-of-life item; it's a Phase 2 Extensibility headline.

**Phase 3 — Cloud & Orchestration:** Both items (`Managed Agents integration`, `Greenhouse companion app`) remain `[ ]` unfilled. However, Plan 41 created the headless cores that make Plan 42 (`bonsai mcp` stdio server) a concrete, fast-follow next step. The roadmap Phase 3 currently names only "Managed Agents integration" (cloud deployment), but the near-term concrete work is an MCP stdio server that would enable native AI-in-editor integration. This item is missing from Phase 3.

**Phase 4 — Ecosystem:** Three items remain `[ ]`. No new work here — still accurate.

### Step 2: Check milestone accuracy

**Next milestones (Phase 2 open items):**

- `[ ] Self-update mechanism` — still in Backlog P3 (improvement/low-priority). Appropriate priority; no supersession.
- `[ ] Template variables expansion` — still pending; no concrete plan or backlog entry exists for this. The Backlog Hygiene routine flagged today that this item has no backlog entry (see Backlog Hygiene report 2026-08-09). May be partially addressed by Plan 40's frozen schema work (which expands the template context) but not explicitly matched.
- `[ ] Micro-task fast path` — still in Backlog P3. Appropriate; no new decisions invalidate it.

**Phase 3 readiness shift:** The Settled decision "Defer Managed Agents cloud integration until local foundation is stable" (KeyDecisionLog, 2026-04-02) was contingent on the local CLI being solid. Plan 41 delivered the headless contract, which means the gating condition for Phase 3 work has materially changed. The concrete next step is Plan 42 (MCP server) — a near-term item that doesn't appear in Phase 3 at all.

**Plan 40 Phase 4 HELD:** `bonsai update` delivery of new scaffolding items (Odysseus manifest + Memory/ scaffolding) is still HELD — this piece of Phase 2 extensibility work is unresolved. The v0.5.0 tag is also still uncut.

### Step 3: Cross-check against Key Decision Log

Read `station/Logs/KeyDecisionLog.md` (Structural, Domain-Specific, Settled sections).

**Decisions that remain valid:** All structural, domain-specific, and settled decisions are consistent with the current build. No existing decisions have been invalidated by Plans 39/40/41.

**Missing decisions (not recorded in KeyDecisionLog):**

1. **Plan 41 core architecture** — "Each command's core is a pure function: typed-options in → structured `Result` out. No Huh prompts, no `os.Exit`, no `fmt.Println` for data. MCP adapter = thin wrapper calling same cores, zero duplicate work." This is a structural decision (shapes how MCP and future CLI features should be built) that belongs in the Structural section of the KeyDecisionLog. Currently it exists only in Plan 41's architecture section.

2. **Plan 40 decisions** — Manifest location (`.bonsai/project.yaml`, host-agnostic), config split (`.bonsai.yaml` = generator-facing, `project.yaml` = hub-facing, seeded once, never reconciled), memory routing (decisions → `Memory/decisions/`, not KeyDecisionLog), schema freeze at v1 — these are architectural decisions locked in June 2026 that are not in the KeyDecisionLog.

### Step 4: Report findings

Per procedure, no changes made to `Roadmap.md`. All findings flagged for user review below.

### Step 5: Update dashboard

Dashboard row for Roadmap Accuracy updated: `Last Ran` → 2026-08-09, `Next Due` → 2026-08-23, `Status` → `done`.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | Phase 2 missing shipped item: Plan 41 — Headless CLI Contract (all 4 mutating cmds have `*Result` cores + JSONL/exit contract + `ExitConflict=5`; `list --json`; `docs/agent-interface.md`; shipped 2026-06-16) | `Roadmap.md` Phase 2 | Flagged for user — recommend adding `[x]` item for agent-drivable CLI surface |
| 2 | MEDIUM | Phase 2 missing shipped item: Plan 40 Phases 1–3 — frozen v1 schemas, `.bonsai/project.yaml` manifest, memory routing, `bonsai guide` Formats page (shipped 2026-06-13, untagged v0.5.0) | `Roadmap.md` Phase 2 | Flagged for user — recommend adding `[x]` item for platform schemas + manifest |
| 3 | MEDIUM | Phase 2 missing shipped item: v0.4.2 non-interactive CLI flags (`--non-interactive --from-config` for `init`/`add`, Plan 39, 2026-05-13) | `Roadmap.md` Phase 2 | Flagged for user — can be folded into finding #1 entry or noted separately |
| 4 | MEDIUM | Phase 3 missing near-term item: Plan 42 (`bonsai mcp` stdio server) is fast-follow to Plan 41 and concrete next step toward Cloud & Orchestration — not in roadmap at all | `Roadmap.md` Phase 3 | Flagged for user — recommend adding `[ ]` item for `bonsai mcp` Plan 42 |
| 5 | LOW | KeyDecisionLog missing Plan 41 structural architecture decision: pure-function headless cores → MCP as thin wrapper, no duplicate implementations | `KeyDecisionLog.md` Structural section | Flagged for user |
| 6 | LOW | KeyDecisionLog missing Plan 40 decisions: manifest location, config split, memory routing, schema v1 freeze | `KeyDecisionLog.md` Domain-Specific section | Flagged for user |

---

## Suggested Roadmap Corrections

These are proposed edits for user review. Do not apply without user approval.

### Phase 2 — additions (all `[x]` complete)

```markdown
- [x] Agent-drivable CLI surface — all mutating commands (`init`/`add`/`update`/`remove`) have pure `*Result` headless cores + JSONL/exit contract (`ExitConflict=5`); `list --json`; `docs/agent-interface.md` contract doc _(v0.4.2 flags for init/add; Plan 41 for update/remove + contract; June 2026)_
- [x] Platform schemas + manifest — `.bonsai/project.yaml` project identity manifest, frozen memory-note schema v1, `bonsai guide` Formats page, `validate` lint for both _(Plan 40 Phases 1–3, untagged v0.5.0, June 2026; Phase 4 delivery-via-update HELD)_
```

### Phase 3 — addition (`[ ]` not yet started)

```markdown
- [ ] `bonsai mcp` stdio server — MCP tool wrapping for all commands using Plan 41 headless cores; native AI-in-editor integration (Claude Code, Cursor, Desktop); `readOnlyHint`/`destructiveHint` + `structuredContent` from `Result` (Plan 42, fast-follow to Plan 41)
```

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **[HIGH] Phase 2 roadmap gap — Headless CLI contract (Plan 41)** — The biggest shipped work since the last roadmap-accuracy run is not in the roadmap at all. Recommend adding the suggested entry above. Plan 41 is still in `Plans/Active/` — the Doc Freshness Check today also flagged it should be archived to `Plans/Archive/`.

2. **[MEDIUM] Phase 2 roadmap gap — Platform schemas (Plan 40)** — Plan 40 Phases 1–3 shipped a significant set of platform-facing features (manifest, memory schema, validate lint, Formats guide). Phase 4 (delivery via `bonsai update`) is still HELD; the v0.5.0 tag is uncut. User should decide whether to reflect only what shipped or wait for Phase 4 + tag before adding to the roadmap.

3. **[MEDIUM] Phase 3 roadmap gap — Plan 42 (MCP server)** — Plan 42 is an explicitly planned fast-follow and the concrete near-term Phase 3 step. Adding it to the roadmap now establishes intent and makes the next planning conversation cleaner.

4. **[LOW] KeyDecisionLog gaps** — Two structural/architectural decisions from Plans 40 and 41 are not in the KeyDecisionLog. User should decide whether to backfill. The Plan 41 decision (pure-function core → MCP shim) is particularly useful for future agents planning Phase 3 work.

5. **[MEDIUM] Phase 2 "Template variables expansion" — no backlog entry** — This roadmap item has no corresponding Backlog entry and no concrete plan. The Backlog Hygiene routine also noted this today. User should either create a Backlog entry with concrete scope, or decide whether Plan 40's schema/context work partially satisfies the intent and reword the roadmap item.

---

## Notes for Next Run

- Verify Plan 40 Phase 4 status (delivery via `bonsai update`) and v0.5.0 tag — if shipped, roadmap needs updating.
- Verify Plan 42 status — if shipped, mark `[ ]` → `[x]` and check Phase 3 progress.
- Check whether Plan 41 has been archived from `Plans/Active/` to `Plans/Archive/`.
- KeyDecisionLog backfill from Plans 40/41 if user approved it this cycle.
- Cross-check whether Phase 2 "Template variables expansion" got a concrete scope or was removed.
