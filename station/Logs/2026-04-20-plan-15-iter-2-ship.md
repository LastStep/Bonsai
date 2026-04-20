---
tags: [log, session, plan-15, bubbletea, dogfooding]
description: Plan 15 iter 2 ship + iter 2.1 reviewer fixes + Bonsai-Test dogfooding findings.
---

# 2026-04-20 — Plan 15 Iter 2 Ship + 2.1 Fixes + Dogfooding

**Plan:** [15-bubbletea-foundation.md](../Playbook/Plans/Active/15-bubbletea-foundation.md)
**Branch:** `ui-ux-testing` (local-only; whole branch merges to main at iter 3 completion)
**Tip:** `38bf035` (docs) · `d0e6256` (iter 2.1 code) · `4011882` (iter 2 code)

## Shipped

### Iter 2 — `cmd/add.go` migration (`4011882`)

Dispatched to `general-purpose` agent with worktree isolation. Scope per plan:
- Migrated `runAdd` + `runAddItems` onto the BubbleTea harness via a single `LazyGroup` splice selecting the branch at runtime.
- New primitives: `NoteStep` adapter, `tui.TitledPanelString` helper, harness `splicer` interface + `expandSplicer` reducer.
- Helper extraction: `newDescriber`, `userSensorOptions`, `asString`/`asStringSlice`/`asBool` lifted from `cmd/init.go` for cross-command reuse.
- 3 reducer tests for `LazyGroup` + 2 for `NoteStep` + 3 for `TitledPanelString`.

Independent reviewer APPROVED with 4 non-blocking nits (nested-splicer docstring, empty-splice re-entry, WindowSizeMsg re-broadcast, LazyStep-in-LazyGroup test) — all unreachable from current callers.

### Iter 2.1 — post-ship reviewer fixes (`d0e6256`)

User asked for a post-ship audit; three independent review agents surfaced **four real regressions** beyond the iter-2 reviewer's non-blocking nits:

- **A. Stale review panel after Esc-back** — `LazyStep.Reset()` didn't clear `built=false` / drop `inner`, so re-entering ran the STALE inner step against new prior results. Extended harness Esc reset loop to `[new_cursor, origCursor]` inclusive so a tail `LazyStep` actually resets. New `TestLazyStepRebuildsOnReset` + updated `TestEscPopReinitsActiveStep`.
- **B. Tech-lead bootstrap regression** — iter 2's pre-harness "require tech-lead" gate blocked `bonsai add` from running at all without tech-lead; but the flow is how users bootstrap. Removed block; non-tech-lead-without-tech-lead path now shows in-harness NoteStep + post-harness `tui.ErrorDetail`.
- **C. "All installed" zero-keystroke** — iter 2 forced user to Enter-past a NoteStep before the empty banner printed. Filter logic lifted to `availableAddItems` helper; splicer returns nil slice when empty; post-harness renders `tui.EmptyPanel` with zero keystrokes (matches pre-iter-2).
- **D. Defensive guards** — `expandSplicer` filters nil steps; `View()` short-circuits with muted "terminal too small" notice when body <3 rows.

6 non-fix follow-ups routed to Backlog Group F (see below). All four iter-2 reviewer nits carried forward to iter 3 scope.

## Dogfooding — Bonsai-Test

User ran manual smoke on `~/Apps/Bonsai-Test` and surfaced three findings:

1. **`src/sda/.claude/settings.json` in Updated panel** — path concat bug in `showWriteResults` (`cmd/root.go:131`). When adding agent X, multi-agent generators (`SettingsJSON`, `PathScopedRules`, `WorkflowSkills`) touch ALL agents' files; the panel renders them under X's workspace root → wrong tree prefix. Actual file on disk is correct. **Pre-existing, not Plan 15.**
2. **`.claude/skills/{name}/SKILL.md` explained** — Claude Code's native skill format; Bonsai auto-generates shims for a curated subset of workflows (`CuratedSlashWorkflows`). Intentionally absent from `bonsai list` (generated artifact, not a catalog selection).
3. **Phantom "Updated" files run-to-run** — `AgentWorkspace` (`generate.go:1213`) iterates `cfg.Agents` (a Go map, randomized iteration order) to build `ctx.OtherAgents`. Templates that `range .OtherAgents` (identity.md.tmpl, scope-guard-files.sh.tmpl, dispatch-guard.sh.tmpl) render different bytes across runs → `writeFile` flags as `ActionUpdated` even when logical state unchanged. **Pre-existing bug, Plan 15 just made it visible via add-flow re-runs.** Fix: sort `OtherAgents` by agent type before return.

Both real bugs filed to Backlog Group F (`38bf035`).

## Verification

Both iter 2 (`4011882`) and iter 2.1 (`d0e6256`) passed:
- `go build ./...` — clean
- `go vet ./...` — clean
- `gofmt -s -l .` — no output
- `go test ./... -count=1` — all 4 packages pass

Manual smoke on user's Bonsai-Test workspace exercised `bonsai init`, `bonsai add` new-agent, `bonsai add` add-items, `bonsai add` re-run — flow works end-to-end; only issues are pre-existing panel/template bugs.

## Commits (chronological on `ui-ux-testing`)

| SHA       | Kind | Summary |
|-----------|------|---------|
| `4011882` | code | iter 2 — `cmd/add.go` harness migration + LazyGroup/NoteStep/TitledPanelString |
| `c5f4265` | docs | iter 2 ship docs + reviewer nits routed |
| `d0e6256` | code | iter 2.1 — 4 reviewer fixes (A/B/C/D) |
| `202d4ca` | docs | iter 2.1 plan reconcile + 6 Backlog Group F follow-ups |
| `38bf035` | docs | 3 Bonsai-Test dogfooding findings (2 bugs + 1 explanation) |

## Backlog Additions (Group F)

**From iter-2.1 review:**
- Spinner Ctrl-C partial-write (waypoint for iter 3 SpinnerStep)
- Workspace validator normalization (`./backend/` vs `backend/`)
- Panic recovery around `Splice`/`Build`
- `ConditionalStep` adapter for empty-picker skip
- AltScreen release-note documentation
- Completion-report deviations-list hygiene

**From dogfooding:**
- `showWriteResults` cross-workspace path concat
- Nondeterministic `OtherAgents` ordering phantom updates

## Status

**Plan 15:** iter 1 shipped (`150d1d3`) + iter 2 shipped (`4011882`) + iter 2.1 shipped (`d0e6256`). Iter 3 outlined: migrate `cmd/remove.go` + `cmd/update.go`, carry-forward nits, `SpinnerStep`, `ConditionalStep`.

**Next session:** Iter 3 dispatch. Whole-branch merge to main after iter 3 ships.
