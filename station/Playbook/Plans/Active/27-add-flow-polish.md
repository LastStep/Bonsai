---
tags: [plan, tier-2, uiux]
description: Plan 27 — bonsai add cinematic polish bundle. Rail canonicalisation, stale-state bug, cross-agent conflict leak, layout parity, conflict UX overhaul.
---

# Plan 27 — Add Flow Polish

**Tier:** 2
**Status:** Draft
**Agent:** general-purpose (single dispatch, worktree-isolated)

## Goal

Land the dogfooding findings on `bonsai add` cinematic: rail matches init's canon, two real bugs fixed, conflict UX redesigned as standalone full-screen vertical picker with batch-resolve controls, final card reaches init parity.

## Context

`bonsai add` cinematic shipped in Plans 23 Phase 1/2/3 (PRs #59, #62, #64). User dogfooded the shipped flow and filed 9 findings spanning rail labelling, stale state after esc-back, cross-agent conflict detection, conflict picker UX, and final-view parity with init. Research confirmed:

- **Init rail = 4 stages** (Vessel · Soil · Branches · Observe) plus a chromeless Planted terminal. No Ground, no Grow. Rail labels baked in `internal/tui/initflow/fallback.go:87` as `var StageLabels = [4]StageLabel{...}`.
- **Add rail = 7 stages** currently (Select · Ground · Graft · Observe · Grow · Conflict · Yield). Labels in `internal/tui/addflow/addflow.go:49-57`.
- **Bug #1 (stale Select state):** LazyGroup.Splice is idempotent via `spliced bool` guard at `internal/tui/harness/steps.go:617-621`. After first branch (add-items vs new-agent), esc-back to Select + pick different agent cannot re-splice. Harness replaces the LazyGroup with children at `expandSplicer` in `harness.go:193-223` — original reference lost, so unsplice is impossible today.
- **Bug #7b (cross-agent conflict leak):** `generate.PathScopedRules`, `generate.WorkflowSkills`, `generate.SettingsJSON` iterate `cfg.Agents` (ALL agents) at `internal/generate/generate.go:1066`, `1104`, `477`. On `bonsai add <new>`, these regenerate files under OTHER already-installed agent workspaces, and `writeFile` → `lock.IsModified` compares disk vs lockfile, flagging user-edited files on unrelated agents as conflicts.
- **Width constants:** all stages use shared `PanelContentWidth = 84` via `PanelWidth(termWidth)` at `internal/tui/initflow/design.go:23-29`. No stage-specific width diverges at the panel level — "Observe feels narrow" is likely an internal-body issue, not a panel issue.
- **Ground stage** collects workspace path + optional display name for new-agent branch. Skipped on add-items branch. Cannot simply delete — must either fold into Select or default-and-inline-edit.

Non-goals for Plan 27:
- Touching `bonsai init` rail or stage set. Init stays at 4 stages.
- Rewriting `bonsai update` conflict flow. Only `bonsai add` is in scope.
- Inline diff previews in Conflict picker (tracked separately — `FileResult` does not carry a diff field yet; Plan 23 TODO at `conflicts.go:320-325`).

## Dispatch Strategy

Split into **2 sequential PRs**:

- **PR1 — Foundations + Bugs:** Phase A (rail canon, Graft→Branches rename, Observe width audit) + Phase B (B1 harness re-splice, B2 cross-agent conflict leak). Verification subset: `make build`, `go test ./...`, `gofmt`, dogfood step #3 (esc-back agent switch) + #2 (no cross-agent conflicts).
- **PR2 — Polish + Verify:** Phase C (C1–C9) + full Phase D dogfood pass. Branches off main AFTER PR1 merges.

PR1 lands the foundation + correctness fixes; PR2 is pure UX polish. Rationale: bug fixes ship sooner, polish benefits from stable rail/rename state, each PR stays under ~1k LoC for reasonable review burden.

## Dependencies

- Plan 23 Phase 3 cutover merged (`788fa6c`). Cinematic add-flow is the only path — legacy removed.
- Current main at `a9df552`. Working tree clean. No in-flight work conflicts.
- Security: any change to conflict detection scope must preserve the invariant "no silent overwrite of user edits". Refer to `Playbook/Standards/SecurityStandards.md`.

## Steps

### Phase A — Foundations

**A1. Rail canon for addflow.**

Target labels in `internal/tui/addflow/addflow.go`:

```go
const (
    StageIdxSelect  = 0
    StageIdxBranches = 1
    StageIdxObserve = 2
    StageIdxYield   = 3
)

var StageLabels = []initflow.StageLabel{
    {Kanji: "選", Kana: "えらぶ", English: "SELECT"},
    {Kanji: "枝", Kana: "えだ", English: "BRANCHES"},
    {Kanji: "観", Kana: "みる", English: "OBSERVE"},
    {Kanji: "結", Kana: "むすぶ", English: "YIELD"},
}
```

- Drop `StageIdxGround`, `StageIdxGraft`, `StageIdxGrow`, `StageIdxConflicts` from the visible rail constants. Rail length = 4.
- Ground / Grow / Conflicts stages remain as steps in `h.steps` but render chromeless (no rail, no header/footer chrome) — see C6, C7, C1. The rail's displayed index stays at OBSERVE while these three render, so the user sees a continuous "still on Observe" until Yield.
- Add a package-level sentinel for each off-rail stage: `const StageIdxOffRail = -1`. Off-rail stages set this in their constructor via `base.SetRailIndex(StageIdxOffRail)`. If the existing `SetRailIndex` does not accept -1 gracefully, extend the `Stage` base to skip rail render when index < 0.

**A2. Graft → Branches rename.**

- Rename `internal/tui/addflow/graft.go` → `internal/tui/addflow/branches.go` via `git mv`.
- Rename `GraftStage` → `BranchesStage` across all callers. Factory ctors `NewNewAgentGraft` / `NewAddItemsGraft` → `NewNewAgentBranches` / `NewAddItemsBranches`. Closure type `GraftContext` → `BranchesContext`.
- Rename `GraftResult` → `BranchesResult` in `addflow.go`. Update `cmd/add.go:buildAddGrowAction` type-scan to read the new type.
- Rename `graft_test.go` → `branches_test.go`. Update test names accordingly.

**A3. Observe body width audit.**

- Open `internal/tui/addflow/observe.go` — find any internal body column that uses a narrower budget than `initflow.PanelWidth(s.Width())`.
- Align body panels to the same `PanelWidth` used by Select / Branches / Yield. Target: visual edge of body column matches across all four main stages at the same terminal width.
- Record before/after widths in the PR body as evidence (screenshots or row-count dump).

### Phase B — Bug fixes

**B1. Stale Select state on esc-back + re-pick (finding #1).**

Root cause (traced in research): after Select → agentA → AddItemsBranches is spliced in, esc-back resets Select's Done flag but the spliced children remain in `h.steps`. Picking agentB on forward advance lands back on the previously-spliced BranchesA instance with agentA's categories baked in. LazyGroup has no way to re-splice because it was replaced in `h.steps` at splice time.

Fix — add harness-level re-splice support:

1. `internal/tui/harness/steps.go` — add `Reset()` on `LazyGroup`:
   ```go
   func (g *LazyGroup) Reset() tea.Cmd {
       g.spliced = false
       return nil
   }
   ```

2. `internal/tui/harness/harness.go` — preserve splice metadata so esc-back can unsplice:
   - Extend Harness with a `splices []spliceRecord` field where:
     ```go
     type spliceRecord struct {
         lg       *LazyGroup // retained ref to the original splicer
         insertAt int        // absolute index where it originally sat
         length   int        // count of inserted children
     }
     ```
   - In `expandSplicer`: before replacing the LazyGroup in `h.steps`, append a `spliceRecord{lg: sp.(*LazyGroup), insertAt: h.cursor, length: len(inserted)}`. (Guard on type assertion — `LazyStep` uses the same splicer iface; only apply the record path to LazyGroup.)
   - In the esc-back branch of `Update` (around `harness.go:268-314`): after computing `h.cursor` backward, iterate `h.splices` in reverse; for any record with `insertAt >= h.cursor`, unsplice: replace `h.steps[record.insertAt : record.insertAt+record.length]` with `[]Step{record.lg}`, call `record.lg.Reset()`, drop the record. Adjust subsequent record offsets if multiple splices exist.
   - Then fall through to the existing SetPrior + Reset loop over `[h.cursor, origCursor]` unchanged.
   - On the next forward advance past the LazyGroup, `expandSplicer` fires again with fresh prev results.

3. Verification: write a harness-level composition test in `internal/tui/harness/steps_test.go`:
   - Construct `[TextStepA, LazyGroup(fn)]` where `fn` reads prev[0] and returns different children for "foo" vs "bar".
   - Run: advance TextStep with "foo" → asserts children for "foo" are spliced.
   - Simulate esc: h.cursor goes back to 0, LazyGroup is restored, children dropped.
   - Change TextStep result to "bar", advance → asserts children for "bar" are spliced.
   - Failing today, passes after fix.

**B2. Cross-agent conflict leak (finding #7b).**

Root cause: three generator functions regenerate files for all installed agents on every `bonsai add` call. Fix — scope the three functions to a single agent when called from `add`:

1. `internal/generate/generate.go`:
   - Add a new exported signature `PathScopedRulesForAgent(projectRoot string, agent *config.InstalledAgent, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, wr *WriteResult, force bool) error`. Body iterates only `agent.Skills` (and any cross-cut rule source) — no `for _, installed := range cfg.Agents` loop. `cfg` is still passed for palette/context.
   - Same for `WorkflowSkillsForAgent` and `SettingsJSONForAgent` at `internal/generate/generate.go:1103` and `:477`.
   - Keep existing `PathScopedRules`, `WorkflowSkills`, `SettingsJSON` intact — they are still correct for `bonsai update` which regenerates all agents. Mark the new `*ForAgent` variants in doc comments as the `bonsai add` path.

2. `cmd/add.go` — at the `buildAddGrowAction` call sites (approx `cmd/add.go:508-511` for new-agent + the add-items branch):
   - Replace `generate.PathScopedRules(...)` with `generate.PathScopedRulesForAgent(cwd, installed, cfg, cat, lock, wr, false)`.
   - Same for `WorkflowSkills` + `SettingsJSON`.
   - `installed` is the `*config.InstalledAgent` already in scope (populated by `buildAddGrowAction`).

3. `internal/generate/generate_test.go` — add a table test `TestPathScopedRulesForAgentScope`:
   - Seed two installed agents (tech-lead, frontend) with distinct Skills/Workflows.
   - Modify a file on disk under tech-lead's workspace.
   - Call `PathScopedRulesForAgent(..., frontendInstalled, ...)`.
   - Assert `WriteResult.Conflicts()` contains ZERO entries pointing under `tech-lead/`.
   - Negative case: calling the legacy `PathScopedRules` with same fixture produces the cross-agent conflict (proves the fix is at the call-site level, not the generator's shared machinery).

### Phase C — Polish

**C1. Conflict standalone full-screen surface (finding #7a).**

- Lift `ConflictsStage` out of the addflow rail. Keep the file at `internal/tui/addflow/conflicts.go` but drop rail-chrome composition — render as a chromeless full-screen stage (similar to initflow Planted).
- The stage should still sit in `h.steps` so the harness drives it, but:
  - Do not call `initflow.RenderEnsoRail` in the View.
  - Do not call `initflow.RenderHeader` / `RenderFooter`. Compose a dedicated conflict header block with its own title row + per-file context; compose an inline CTA row in place of the rail footer.
  - Implement the `harness.Chromeless` interface so the harness knows not to wrap the stage in rail chrome.
- The stage is placed AFTER Observe (conditional on `wr.HasConflicts()`) and BEFORE Grow's final render / Yield. Keep the existing post-harness type-scan in `cmd/add.go:256-262` for the conflict picks map.

**C2. Vertical list layout (finding #7c).**

- Replace the horizontal tab strip in `conflicts.go:renderTabs` with a vertical list — one row per conflict file, focus marker on the active row.
- Row layout: `[focus-glyph] [action-glyph-colored-by-pick] [relative-path] [action-label]`.
- Use a Viewport to wrap when the conflict count exceeds available rows.
- Keybindings update:
  - `↑ ↓ / j k` — move focus row (no wrap).
  - `1 / 2 / 3` — set focused row's action to Keep / Overwrite / Backup.
  - `␣` — cycle focused row's action (Keep → Overwrite → Backup → Keep).
  - `↵` — advance to Grow (or complete stage).
  - `esc / shift+tab` — back to Observe.

**C3. Per-file action color (finding #8).**

- In `renderRow` for the vertical list, select the row text color / glyph color from `initflow.ConflictActionGlyph(tone)` + `ConflictRowStyle(tone)` based on the current action for that row. The color updates live as the user cycles `␣` or hits `1/2/3`.
- Keep the default (Keep) rendered in Success tone; Overwrite in Warning tone; Backup in Danger tone. Reuse existing palette tokens from `internal/tui/initflow/design.go`.

**C4. Batch-resolve row (finding #8).**

- Above the footer (separate block, below the vertical list), render a single row of three buttons: `[Keep all]  [Overwrite all]  [Backup all]`.
- Keybindings `K / O / B` (uppercase) apply the corresponding action to every row in the list. Lowercase `k/o/b` can remain unassigned or map to the focused row only — choose one shape, document in the key hint footer.
- Visually separate the batch row from the per-file list with a single blank line + muted caption: `batch resolve:`.

**C5. Conflict rail drop.**

- Remove `衝 CONFLICT` from `StageLabels` (dropped in A1 already). The Conflict stage is off-rail — do not increment `stageIdx` through it. Keep Observe as the "current" rail anchor while Conflict is showing, OR leave the rail invisible during Conflict since it's chromeless (preferred — less visual churn between Observe and Conflict).

**C6. Drop Ground rail stage (finding #2) — render off-rail chromeless.**

- Remove `地 GROUND` from `StageLabels` (dropped in A1 already).
- Keep `GroundStage` as a separate step in the splicer sub-sequence BUT render it chromeless (no rail, no header/footer chrome). Same pattern as the Conflict stage in C1.
- Add `Chromeless() bool { return true }` to `GroundStage` if not already present.
- The rail must keep SELECT as the "current" anchor while Ground renders (no increment). Simplest: Ground renders a centred full-body form with its own inline key hints; since it's Chromeless, the harness does not wrap it in rail chrome.
- The splicer in `cmd/add.go:120-183` is unchanged in structure: new-agent branch still splices `[Ground, Branches, Observe]`. Add-items branch still skips Ground. The rail length (4) is decoupled from the step count.
- Inline-into-Select expansion is explicitly deferred. Chromeless off-rail Ground is the committed design.

**C7. Drop Grow rail stage (finding #6) — visible flow ends at Observe.**

- Remove `育 GROW` from `StageLabels` (dropped in A1 already).
- Grow still exists as a step — it runs the generate action. Render it chromeless: a centred progress block with the Bonsai tree animation, no rail, no header/footer chrome.
- After Grow succeeds, proceed to Conflict (if any) → Yield. The rail shows Observe as the last visible tab throughout Grow's spinner.

**C8. Agent row cosmetics (finding #3).**

- In `addflow/select.go:renderRow`:
  - Strip the word "Agent" (case-insensitive) from `opt.DisplayName` before rendering. E.g. "Tech Lead Agent" → "Tech Lead". This is a display-only strip; the catalog's `DisplayName` is unchanged, `opt.Name` is still the machine identifier.
  - Show the FULL description — no truncation via `rr[:descColW-1]`. Allow the description to flow to the row's right edge; truncate only if it exceeds `s.Width() - nameColW - badge width - 4`.
  - Push the `(installed)` badge to the rightmost column of the row, AFTER the description, not inside the name column. New row layout: `[border 2] [glyph 1] [sp 1] [name] [sp 2] [description fill...] [sp 2] [installed-badge]`.
  - Badge should right-align via a pad-fill against the row's total width.

**C9. Yield parity with init's Planted (finding #9).**

- Open `internal/tui/addflow/yield.go` vs `internal/tui/initflow/planted.go`. Align:
  - Outer frame padding + vertical centering.
  - Section rhythm: title → hero line → divider → content blocks → blank → key hints.
  - Key-hint row style / glyph spacing.
  - Footer or chromeless-hint parity (Planted is chromeless with inline `↵ exit · q quit`; Yield should match).
- Retain Yield's four variants (success / all-installed / tech-lead-required / unknown-agent). Only the success variant needs full parity with Planted. The three error-panel variants keep their current shape.

### Phase D — Verification

Run before marking done:

```bash
make build
go test ./...
gofmt -s -l internal/tui/ cmd/
```

Manual dogfooding (report screenshots in PR):

1. `bonsai add tech-lead` when already installed — AddItemsBranches path, verify ability list is filtered, Observe + Yield render correctly.
2. `bonsai add frontend` in a project with tech-lead installed — new-agent branch, verify workspace stamp on Select, verify NO conflicts appear under `tech-lead/`.
3. `bonsai add <agent-A>` → press ` ` (toggle) on a few abilities → esc → pick `<agent-B>` → verify BranchesB renders with B's abilities, NOT A's.
4. Trigger a conflict: edit a generated file on disk post-install, re-run `bonsai add <same-agent>` → verify standalone conflict surface, vertical list, per-file color update, batch-resolve row works.
5. Narrow terminal to `70 × 20` — verify all four rail stages render without overflow, Conflict standalone respects `MinSizeFloor`.

## Security

> [!warning]
> Refer to [Playbook/Standards/SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

Specific callouts for this plan:

- **Conflict scope fix (B2) must not regress the "no silent overwrite" invariant.** Backup-failure drop pattern from Plan 23 Phase 3 `applyCinematicConflictPicks` (`cmd/add.go`) MUST still fire for any conflict the user picks Overwrite/Backup on. The scope fix reduces the set of files appearing as conflicts; it does NOT change what happens once a file is flagged.
- **Workspace path input in Select inline form (C6)** inherits the existing Ground-stage validation (non-empty, no `..`, no absolute paths, collision check). Port the validator, do not weaken it.
- **No new dependencies.** All visual changes use existing lipgloss palette tokens + existing Viewport / StageLabel primitives.

## Verification

- [ ] `make build` passes
- [ ] `go test ./...` passes — includes new harness re-splice test + new `PathScopedRulesForAgentScope` test
- [ ] `gofmt -s -l internal/tui/ cmd/` emits no paths
- [ ] Rail shows exactly 4 stages: SELECT · BRANCHES · OBSERVE · YIELD
- [ ] Esc-back from Branches → Select → pick different agent → Branches shows new agent's categories + clean selections
- [ ] `bonsai add frontend` in a project with tech-lead installed produces zero conflicts under `tech-lead/**`
- [ ] Conflict surface is full-screen, vertical, off-rail
- [ ] Per-file row color reflects picked action; `K/O/B` batch row applies to all
- [ ] Select row shows full description + right-aligned `(installed)` badge + no "Agent" suffix
- [ ] Yield success variant matches Planted's section rhythm
- [ ] No regressions on `bonsai init` or `bonsai update`
- [ ] 6/6 CI green on draft PR (test / lint / Analyze-Go / govulncheck / CodeQL / GitGuardian)
- [ ] Independent code-review agent returns PASS (minors acceptable — file to Backlog)

## Notes for the dispatched agent

- Work in a fresh worktree on a branch named `plan-27/add-flow-polish` off origin/main.
- This plan is intentionally single-dispatch (Tier 2 bundle). Do NOT split into multiple PRs unless you hit a build-breaking sequencing problem; then split into 2 — foundations+bugs first, polish second — and report back before proceeding.
- Do NOT touch `internal/tui/initflow/**` beyond the additive exports needed for chromeless composition. Init's rail + stages are out of scope.
- Do NOT change `bonsai update` conflict handling. Only `bonsai add`.
- If any ambiguity blocks a step, stop and report — do not make design decisions. Specifically: if C6 inline expansion proves hairy, fall back to off-rail chromeless Ground (documented option) and note the choice in the PR body.
- Draft PR body format per issue-to-implementation workflow (`station/agent/Workflows/issue-to-implementation.md`).
