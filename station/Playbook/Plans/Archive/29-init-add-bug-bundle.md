---
tags: [plan, tier-2, bugs, security]
description: Plan 29 — init + add command bug bundle. Header label fix, conflicts cleanup, Grow redundancy, Yield dedup, Ground/Vessel path-traversal hardening, test gap fills.
---

# Plan 29 — Init + Add Bug Bundle

**Tier:** 2
**Status:** Complete — shipped 2026-04-23 (PR #72 squash 9eb2bff)
**Agent:** general-purpose (single worktree dispatch) + tech-lead
**Target:** 1 PR, same-session merge

## Goal

Knock off all outstanding bugs and test gaps surfaced by deep-dive of `cmd/init_flow.go`, `cmd/add.go`, `internal/tui/addflow/`, and `internal/tui/initflow/`. Includes the Plan 27 cosmetic/test-gap bundle, a header-label bug introduced by Plan 28 Phase 1, and a path-traversal validator gap that was flagged in Plan 27's security callout but never implemented.

## Context

Deep-dive audit found 8 real bugs + 4 test gaps in init/add code. Findings are file-level specific. No architectural judgment required; each fix is mechanical against an exact code location.

**Parallel session note:** list/catalog/guide commands are under active work by a parallel session. Do NOT touch `internal/tui/catalogflow/`, `internal/tui/listflow/` (future), `internal/tui/guideflow/` (future), `cmd/catalog.go`, `cmd/list.go`, `cmd/guide.go`. Scope is strictly init + add.

## Dispatch Strategy

Single dispatch via `isolation: worktree` off current `main` tip. Independent code-review + security-review agents in parallel post-PR (security review warranted for §H path-traversal work). Target same-session squash-merge.

## Dependencies

- Main at `c81654c` (Plan 27 close-out). Clean working tree.
- No inflight conflicts on the target files — parallel session works on list/catalog/guide, which don't overlap.
- Plan 27 shipped rail canon, chromeless stages, `*ForAgent` scope narrowing — this plan builds on those.

## Steps

### Phase A — Header label correctness

**A1.** `cmd/add.go:77-85` — stamp the correct header labels for the add command:

```go
ctx := initflow.StageContext{
    Version:          Version,
    ProjectDir:       cwd,
    StationDir:       "station/",
    AgentDisplay:     "",
    StartedAt:        startedAt,
    HeaderAction:     "ADD",
    HeaderRightLabel: "GRAFTING INTO",
}
```

Rationale: header row 2 reads `<action> · v<version>` and right-block row 1 reads `<rightLabel>`. The current values are inherited verbatim from `cmd/init_flow.go:67-68` and were never corrected when Plan 28 Phase 1 added the per-command header labels. Comment at `cmd/add.go:139` explicitly says stages should inherit "GRAFTING INTO" — contradicts the stamped value today.

Remove the stale comment at `cmd/add.go:72-76` that says "HeaderAction + HeaderRightLabel preserve the current add-flow presentation verbatim" — the current presentation IS the bug.

### Phase B — Conflicts stage cleanup

**B1.** `internal/tui/addflow/conflicts.go:204-212` — drop dead `w` assignment in `View()`. `w` is computed but never threaded to `renderBody()` (which reads `s.Width()` directly). Target shape:

```go
func (s *ConflictsStage) View() string {
    h := s.Height()
    if h <= 0 {
        h = 24
    }
    if initflow.TerminalTooSmall(s.Width(), s.Height()) {
        return initflow.RenderMinSizeFloor(s.Width(), s.Height())
    }
    body := s.renderBody()
    // ... rest unchanged
}
```

**B2.** `internal/tui/addflow/conflicts.go:308-324` — reconcile `listHeight()` comment + const. Count non-list rows in `renderBody()` body slice:

```
intro (3) + "" + "" + divider (1) + "" + list (N) + "" +
batchCaption (1) + batchRow (1) + "" + counter (1) + "" + hint (1)
```

Non-list: 3 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 = **14 rows**. Change `const fixedRows = 15` → `14`. Update comment to match. List now gets `h-14` budget instead of `h-15` (1 additional row on 20-row terminal).

### Phase C — Grow redundancy

**C1.** `internal/tui/addflow/grow.go:32` — drop the `g.SetRailHidden(true)` call. `GenerateStage.View()` at `internal/tui/initflow/generate.go:225-253` branches on `bodyOnly` BEFORE calling `renderFrame`, so `railHidden` is unreachable when `bodyOnly=true`. Keep `SetBodyOnly(true)` — that's the actual mode gate. Update the inline comment if needed.

### Phase D — Yield cleanup

**D1.** `internal/tui/addflow/yield.go:26-27` — doc-comment stale after Plan 27. "rail position 5" → "rail position 3 (StageIdxYield)".

**D2.** `internal/tui/addflow/yield.go:225-242` in `renderSuccess()` — collapse the two identical `totalAbilities` / `totalAbilitiesStat` computations. One `totalAbilities int` assignment used for both the hero-stats line AND the summary row.

### Phase E — Ground cleanup

**E1.** `internal/tui/addflow/ground.go:33` — drop `_  lipgloss.Position // reserved — keeps imports stable` field. `lipgloss` import is used by `lipgloss.PlaceHorizontal` at line 225 and other styles; the reserved field is dead.

### Phase F — Test gaps (Plan 27 backlog)

**F1.** `internal/tui/addflow/conflicts_test.go` — add `TestConflicts_ViewportFollowsFocus`:
- Seed with 10 conflict files
- `SetSize(100, 20)` so `listHeight()` returns small enough budget to force viewport
- Set `s.focus = 8`
- Assert `renderList()` output contains the path of `s.files[8]`
- Assert `renderList()` output does NOT contain the path of `s.files[0]` (scrolled off-top)

**F2.** `internal/tui/addflow/conflicts_test.go` — add `TestConflicts_ColorTonesDifferPerAction`:
- One-row fixture (or reuse `newTestConflicts()` + focus row 0)
- Capture `renderRow(0)` with `s.action[key] = ConflictActionKeep`
- Capture `renderRow(0)` with `s.action[key] = ConflictActionOverwrite`
- Assert the two rendered strings are NOT equal (proves palette tone differs per action, not just layout)

**F3.** `internal/tui/addflow/conflicts_test.go` — add `TestConflicts_LowercaseKMovesFocus`:
- Start at focus=0, press `tea.KeyDown` to move to focus=1, then press rune `k`
- Assert focus returned to 0 (lowercase `k` is `up`, not batch-Keep)
- Assert no `action[]` values changed from initial Keep defaults

**F4.** `internal/tui/initflow/generate_test.go` — add `TestGenerateStage_BodyOnlyDropsChrome`:
- Build a GenerateStage via `NewGenerateStage(ctx, func() error { return nil })`
- Call `s.SetBodyOnly(true)`
- Call `s.SetSize(120, 30)`
- Render `s.View()` and assert output does NOT contain rail glyphs (check a distinctive rail char or the enso tick-mark pattern produced by `RenderEnsoRail`) and does NOT contain the footer's `BONSAI 一` brand.

### Phase G — Conflicts dim reference cleanup

**G1.** `internal/tui/addflow/conflicts.go:186` — `renderBatchRow` declares `bark := initflow.LabelStyle()` which is only used once at line 389. Leave unchanged — not a bug; keeping to track in case the agent notices. No action required.

### Phase H — Workspace path-traversal validator

**H1.** `internal/tui/addflow/ground.go` — harden `Update`'s Enter branch:

```go
if m.String() == "enter" {
    v := strings.TrimSpace(s.input.Value())
    if v == "" {
        s.validateErr = "workspace required"
        s.showError = true
        return s, nil
    }
    norm := NormaliseWorkspace(v)
    if reason := invalidWorkspaceReason(norm); reason != "" {
        s.validateErr = reason
        s.showError = true
        return s, nil
    }
    if s.existingWorkspaces[norm] {
        s.validateErr = fmt.Sprintf("workspace %q is already in use", norm)
        s.showError = true
        return s, nil
    }
    // ...advance
}
```

Helper:

```go
// invalidWorkspaceReason returns a user-facing error string when the
// normalised workspace escapes the project root or is absolute. Returns
// "" when the workspace is a safe project-relative path. Called by
// GroundStage after NormaliseWorkspace cleans the input.
func invalidWorkspaceReason(ws string) string {
    // filepath.IsAbs catches "/etc/" on POSIX and "C:\..." on Windows.
    if filepath.IsAbs(ws) {
        return "workspace must be project-relative (no leading /)"
    }
    // After filepath.Clean, any remaining ".." component means the path
    // escapes the project root. Split on "/" (NormaliseWorkspace always
    // emits forward slashes) and check each segment.
    for _, seg := range strings.Split(strings.TrimRight(ws, "/"), "/") {
        if seg == ".." {
            return "workspace must not escape project root (no ..)"
        }
    }
    return ""
}
```

**H2.** `internal/tui/initflow/vessel.go` — apply the same check to the STATION input in `validate()`:

```go
func (s *VesselStage) validate() bool {
    if strings.TrimSpace(s.inputs[vesselIdxName].Value()) == "" {
        return false
    }
    station := strings.TrimSpace(s.inputs[vesselIdxStation].Value())
    if station == "" {
        return true
    }
    if station == "/" {
        return false
    }
    // After normalising to trailing slash, reject absolute + path-escape.
    norm := station
    if !strings.HasSuffix(norm, "/") {
        norm += "/"
    }
    if filepath.IsAbs(norm) {
        return false
    }
    for _, seg := range strings.Split(strings.TrimRight(norm, "/"), "/") {
        if seg == ".." {
            return false
        }
    }
    return true
}
```

Add `"path/filepath"` to the import block. No new user-facing error string — `validate()` returns bool; the existing `showErrors = true` branch at `Update` handles the UX.

**H3.** Tests:

- `internal/tui/addflow/ground_test.go` — add `TestGround_RejectsAbsolutePath`, `TestGround_RejectsParentEscape`, `TestGround_RejectsHiddenParentEscape` (`nested/../..`). Each presses Enter with an invalid value and asserts `Done()=false` + `validateErr` contains the right phrase.
- `internal/tui/initflow/vessel_test.go` — add `TestVessel_RejectsAbsoluteStation`, `TestVessel_RejectsParentEscapeStation`. Assert `validate()` returns false.

**H4.** Security rationale comment block at the top of `invalidWorkspaceReason` (~4 lines): "Project-relative only. Defence against accidental writes outside the project root when the user types `../...` or a rooted path. Not an adversarial boundary — the user already has write access to their own filesystem — but prevents silent surprises in test harnesses and dogfooding sessions."

### Phase I — Verification

```bash
make build
go test ./...
gofmt -s -l internal/tui/ cmd/
```

Manual smoke (dispatched agent reports in PR body):

1. `bonsai add` on a fresh temp project (after `bonsai init`) — Select → Branches → Observe → Grow → Yield — header row 2 shows "ADD · v<X>" (was "INIT") and right-block row 1 shows "GRAFTING INTO" (was "PLANTING INTO").
2. `bonsai add <agent>` in a project with conflicts — conflict vertical list renders, K/O/B batch works, focus + action per row update, no layout regression.
3. `bonsai add <agent>` with workspace input `../foo` → rejected with "must not escape project root".
4. `bonsai add <agent>` with workspace input `/etc/foo` → rejected with "must be project-relative".
5. `bonsai init` with station input `../bar/` → validation fails, error feedback shown.

## Security

> [!warning]
> Refer to [Playbook/Standards/SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

- **Phase H is the security-sensitive portion.** The new validator must not reject legit project-relative paths like `services/api/`, `./foo`, `nested/dir/`, or `./dir with spaces/`. Re-run Plan 27 PR1's `TestNormaliseWorkspace` after the changes to prove normalisation invariants are preserved.
- **Do not weaken existing validation.** Empty, `"/"`, and duplicate-workspace paths all stay rejected. The new checks are additive.
- **No new dependencies.** `path/filepath` is stdlib.

## Verification

- [ ] `make build` passes
- [ ] `go test ./...` passes — includes 4 new conflict/generate tests + 5 new validator tests
- [ ] `gofmt -s -l internal/tui/ cmd/` emits no paths
- [ ] `bonsai add` header renders "ADD · v<X>" and "GRAFTING INTO"
- [ ] Conflict stage still builds + renders correctly after §B1/B2 changes
- [ ] Grow stage still chromeless after §C1 drop
- [ ] Ground stage rejects absolute + path-escape inputs with user-facing error
- [ ] Vessel stage rejects absolute + path-escape STATION input (silent validation error is OK)
- [ ] 6/6 CI green on draft PR (test/lint/Analyze-Go/govulncheck/CodeQL/GitGuardian)
- [ ] Independent code-review agent returns PASS (minors acceptable — file to Backlog)
- [ ] Independent security-review agent returns PASS on §H

## Notes for the dispatched agent

- Work in a fresh worktree on a branch named `plan-29/init-add-bug-bundle` off origin/main.
- **SCOPE LOCK:** Do NOT touch `internal/tui/catalogflow/`, `cmd/catalog.go`, `cmd/list.go`, `cmd/guide.go`, or any `listflow/`/`guideflow/` directory — a parallel session owns those tracks. If you notice bugs there, file to `station/Playbook/Backlog.md` rather than fixing.
- Phase G is a no-op flag — mentioned only so the agent does not add it to the change list.
- If Phase H's validator behaviour surprises legit inputs in the existing test suite (e.g. `TestNormaliseWorkspace` fixtures), **stop and report** — do not relax the new validator to make the old tests pass. The fix may require adjusting fixture values instead.
- Draft PR body format per `station/agent/Workflows/issue-to-implementation.md`. One PR, all phases.
