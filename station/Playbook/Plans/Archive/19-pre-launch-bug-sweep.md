---
tags: [plan, tier-2, oss-release, bugfix, harness]
description: Pre-launch bug sweep — Tier 1 fresh-install blockers + Tier 2 harness polish, bundled for OSS release readiness.
---

# Plan 19 — Pre-launch Bug Sweep

**Tier:** 2
**Status:** Draft
**Agent:** general-purpose (worktree-isolated)

## Goal

Ship eight targeted fixes as one PR so the next `go install github.com/.../bonsai` → `bonsai init` flow is clean for first-time OSS users: four fresh-install blockers + four harness polish items. After merge, manual `init → add → remove → update` on a fresh temp dir must produce no visible artifacts of the known bugs (CRLF, duplicate tech-lead message, tree root mislabel, silent generator failures).

## Context

**Why this batch, why now.** OSS launch is imminent. The Backlog's **Group F** (fresh-install blockers) + half of **Group B** (harness polish) is small, mechanical, and well-scoped — exactly the kind of cleanup that's risky to do *after* launch (first-user-experience regressions are the most expensive bugs to ship with). Items chosen:

| # | Backlog Group | Severity | File footprint |
|---|---|---|---|
| 1 | F | P1 | `.gitattributes` (new), `internal/generate/generate.go`, `internal/generate/generate_test.go` |
| 2 | F | — | `cmd/root.go` (`showWriteResults`) |
| 3 | F | — | `cmd/update.go` (`applyCustomFileSelection`) |
| 4 | B | P1 | 5 files in `cmd/` — ~29 sites |
| 5 | B | — | `internal/tui/harness/steps.go` (`SpinnerStep.Init`) |
| 6 | B | — | `internal/tui/harness/steps.go` (`NewConditional`) |
| 7 | B | — | `internal/tui/harness/steps.go` + `harness.go` (Esc-back re-eval) |
| 8 | B | — | `cmd/add.go` (drop NoteStep, keep ErrorDetail) |

**Deferred** (not in this plan — tracked separately in Backlog): pre-release docs audit, demo GIF, `generate.go` split, catalog/cmd test coverage, trigger test infra, PTY smoke test, GO-2026-4602 stdlib monitor, Group D catalog expansion, Group E workspace QoL.

**Key prior art**
- Plan 15 (BubbleTea harness, merged `2ce63f6`) introduced `recoverBuilder` covering Build/Splice panics but left SpinnerStep action goroutine without `recover`.
- Plan 15 also introduced `SpinnerStep`'s `func() error` return path and `spinnerDoneMsg` propagation. Callers still `_ =` the inner generator errors, making the mechanism useless until call sites migrate.
- `internal/tui/harness/harness.go` Esc-back loop (L284-295) iterates steps + calls `Reset()` but does **not** call `SetPrior()` first, so `ConditionalStep.Reset()` can't re-evaluate its predicate against the user's new upstream picks.

## Dependencies

- Clean working tree on `main` at `5e9255f` (confirmed via pre-flight commit `84479a9`).
- No blocking merges queued — `git log origin/main..HEAD` empty.
- Go 1.24 toolchain available (verified via `go.mod`).
- `make build && go test ./...` green on main (verified during Plan 15 merge).

## Steps

Execution order matters: Tier 1 bug fixes first (each independently verifiable), harness changes second (they depend on spinner refactor landing first), then add.go cleanup.

### Step 1 — CRLF-proof installed shell scripts

**Reproduction (defensive — current machine is LF):** fresh clone on Windows with `core.autocrlf=true` → clone rewrites `.sh.tmpl` sources to CRLF → `bonsai init` emits installed sensor scripts with CRLF → bash refuses `#!/bin/bash\r` → SessionStart hooks silently fail on the user's first session.

**1a. Create `.gitattributes`** at repo root:

```gitattributes
# Force LF on anything that ends up in a generated shell hook or gets shelled out.
*.sh         text eol=lf
*.sh.tmpl    text eol=lf

# Keep lockfiles + generated-preview goldens predictable across checkouts.
go.sum       text eol=lf
*.yaml       text eol=lf
```

**1b. Defensive LF normalization in the generator.** Edit `internal/generate/generate.go` `writeFile` (L262) and `writeFileChmod` (L301). For any target path where `strings.HasSuffix(rel, ".sh")` is true (so the post-`.tmpl`-strip output path), normalize bytes: replace every `\r\n` with `\n` and strip any standalone `\r` before writing. Rationale: belt-and-braces — even if `.gitattributes` gets dropped or a user clones through an old git that ignores it, the binary still writes LF.

Prefer a single helper `normalizeShellLF(data []byte, rel string) []byte` placed immediately above `writeFile`. Call it in both `writeFile` and `writeFileChmod` before the `os.WriteFile` call. No-op for non-`.sh` paths.

**1c. Generation-time test.** Add `TestShellScriptLF` to `internal/generate/generate_test.go`:
- Build a minimal `catalog.Catalog` with one sensor (the test can construct it directly or re-use a helper — match `TestAgentWorkspaceNewFiles` style).
- Run `AgentWorkspace` into a tmp dir.
- For every file in `agent/Sensors/` ending in `.sh`, read bytes, assert zero `\r` characters.
- Test should fail if a future change inlines a template with `\r\n` literal or removes the normalizer.

**Acceptance:** after merge, `find agent/Sensors -name '*.sh' | xargs file` on a fresh `bonsai init` reports ASCII text with LF terminators; new test passes.

### Step 2 — Fix `showWriteResults` cross-workspace tree root

**Reproduction:** `bonsai init` with workspace `station/` → install a code agent `backend` with workspace `src/` → during `bonsai add backend`, the post-harness tree at `cmd/root.go:166-198` is called with `rootLabel = "src"`. But `wr.Files` includes files under `station/` too (because `PathScopedRules` and `WorkflowSkills` touch both workspaces). The prefix-strip path `strings.TrimPrefix(f.RelPath, "src/")` no-ops for `station/...` paths → they render *under* the `src` tree root → user sees nonsense like `src/station/agent/...`.

**2a. Drop the single-root assumption.** Change `showWriteResults(wr *generate.WriteResult, rootLabel string)` signature behaviour:

- Group files by their **top-level segment** (first component before the first `/`).
- For each group, render a separate `FileTree` with that segment as the root label.
- Keep `ActionCreated` / `ActionUpdated`/`ActionForced` / `ActionConflict` bucketing as today, but bucket *within* each top-level group.

**2b. Keep call sites simple.** Drop the `rootLabel` parameter entirely — the function derives roots from paths. Update:
- `cmd/init.go` L274: `showWriteResults(&wr)`
- `cmd/update.go` L197: `showWriteResults(&wr)`
- `cmd/add.go` — check for existing call (likely `showWriteResults(wr, …)`) and drop the label.
- `cmd/remove.go` — same.

**2c. Preserve stable panel ordering.** Sort top-level groups alphabetically so `bonsai init`'s sole workspace produces deterministic output and tests on top of this remain stable. Within each group, `Created` → `Updated` → `Skipped (user modified)`.

**Acceptance:** running `bonsai add backend` in a `station/`-rooted project renders two trees rooted at `src` and `station` (or whichever labels apply) with no cross-root leaf leakage.

### Step 3 — Dedup in `runUpdate.applyCustomFileSelection`

**Reproduction:** user runs `bonsai update`, picks the same custom file twice across agents, or re-runs `update` after a prior successful run — `applyCustomFileSelection` appends to `installed.Skills` without checking for membership. Over time, `.bonsai.yaml` accumulates duplicate entries.

**Fix:** `cmd/update.go:245-285` (`applyCustomFileSelection`). In the switch statement, for each type (`skill`, `workflow`, `protocol`, `sensor`, `routine`), build a `seen := make(map[string]struct{})` from the existing slice before the loop, then `if _, dup := seen[d.Name]; dup { continue }` before the append. Mark `seen[d.Name] = struct{}{}` after append so subsequent iterations in the same call de-dupe too.

Refactor into a small helper to avoid five copies:

```go
func appendUnique(slice []string, name string) []string {
    for _, existing := range slice {
        if existing == name {
            return slice
        }
    }
    return append(slice, name)
}
```

Call `installed.Skills = appendUnique(installed.Skills, d.Name)` etc.

**Acceptance:** `bonsai update` twice in a row, picking the same files each time, produces a `.bonsai.yaml` where each list has no duplicates. Add unit test `TestApplyCustomFileSelectionDedupes` — construct two calls with overlapping selections, assert post-state slice has `len(unique) == len(result)`.

### Step 4 — Spinner error propagation (errors.Join at 29 sites)

**Goal:** wrap every `_ = generate.X(...)` inside a spinner action closure with `errs = append(errs, generate.X(...))` and return `errors.Join(errs...)`. Harness already propagates via `spinnerDoneMsg.err` → `SpinnerStep.Result()`; callers need to actually check it.

**4a. Migrate spinner actions (5 sites).**

- `cmd/init.go` L167-206 — gather `errs` at top of closure, wrap L200-204 generators, return `errors.Join(cfg.Save err if non-nil, errs...)`. Preserve existing early-return on `cfg.Save` err — that's critical (no partial `.bonsai.yaml`). Actually: switch to `errs = append(errs, cfg.Save(configPath))` and keep early return pattern where needed (Scaffolding depends on cfg existing — keep `if err := cfg.Save(...); err != nil { return err }` first).
- `cmd/add.go` L166-169 — `runAddSpinner` body. Find all `_ =` sites there (L334-337, L375-378 based on grep), collect errors, return `errors.Join`.
- `cmd/remove.go` L95-106 (agent removal) — wrap L103 (`cfg.Save`) + L104 (`SettingsJSON`).
- `cmd/remove.go` L315-317 — `runRemoveItemAction` body lives at L451+ (os.Remove calls that currently discard errors). Keep `os.Remove` as `_ =` (file-not-found after lock update is expected); wrap the *generator* calls at L472 (`RoutineDashboard`), L483 (`WorkspaceClaudeMD`), L487 (`cfg.Save`), L488 (`SettingsJSON`) with error collection, return `errors.Join`.
- `cmd/update.go` L113-140 — wrap L134, L136, L137, L138 generators, return `errors.Join`.

**4b. Caller error surfaces (4 sites).**

- `cmd/init.go` L249-257 — already checks `results[10]`, keep.
- `cmd/add.go` L234 — already checks `addOutcome.spinnerErr`. Verify `runAddSpinner` plumbs the errors.Join result into `addOutcome.spinnerErr`.
- `cmd/remove.go` after L126 Run call, before L141 `if !asBool(results[0])` — extract spinner result at index 1 (after the initial Review step), if it's a non-nil error, `tui.Warning("Removal error: " + err.Error())` and `return nil`.
- `cmd/remove.go` after L338 Run call — extract spinner result at index 2 (after optional Select + Lazy). Same `tui.Warning` path.
- `cmd/update.go` after L155 Run call, before L170 conflict picker — extract spinner result at index `len(agentsWithDiscoveries)` (the spinner slot). Same `tui.Warning`.

**4c. Keep `os.Remove` swallows where they're legitimate.** `cmd/remove.go:451,457,463,464` — these are best-effort filesystem cleanups where ENOENT is expected; don't aggregate. Leave as `_ =`. Similarly `cmd/root.go:157` (backup file write — best effort).

**4d. Import `errors` where needed.** Every cmd file touched here should already have `errors` imported (they use `errors.Is`/`errors.As`). Double-check after edits.

**Acceptance:** injected error in a generator (e.g. wrong path permissions) now produces a `Warning: Generation error: <err>` panel instead of silent success + broken output. Update `memory.md` note about ~30 callsites — mark "migrated to errors.Join" after this step.

### Step 5 — SpinnerStep action goroutine recover()

**File:** `internal/tui/harness/steps.go:746-757` (`SpinnerStep.Init`).

Current code runs the action in a `tea.Cmd` closure that panics propagate through the BubbleTea event loop → harness crashes with no panel. Plan 15 iter 3 added `recoverBuilder` for Build/Splice; this is the gap.

**Fix:** wrap the action invocation in a `defer recover()` that converts panic → `spinnerDoneMsg{err: fmt.Errorf("spinner action panic: %v", r)}`:

```go
func (s *SpinnerStep) Init() tea.Cmd {
    runner := s.action
    prev := s.initPrev
    if s.actionP != nil {
        runner = func() error { return s.actionP(prev) }
    }
    return tea.Batch(
        s.sp.Tick,
        func() (msg tea.Msg) {
            defer func() {
                if r := recover(); r != nil {
                    msg = spinnerDoneMsg{err: fmt.Errorf("spinner action panic: %v", r)}
                }
            }()
            return spinnerDoneMsg{err: runner()}
        },
    )
}
```

**Acceptance:** unit test `TestSpinnerStepRecoversFromPanic` — action closure calls `panic("boom")`, harness runs to completion, result is an error matching `spinner action panic:`. Add to `internal/tui/harness/steps_test.go`.

### Step 6 — NewConditional nil-predicate guard

**File:** `internal/tui/harness/steps.go:807`.

Current code accepts a nil predicate; `Init()` at L832 calls `c.predicate(c.initPrev)` → nil deref panic.

**Fix:** in `NewConditional`, if `predicate == nil`, substitute `func(prev []any) bool { return true }` (default to NOT skip — safer than silently skipping). Document with a one-line comment: `// nil predicate = always show (safer default than skip)`.

**Acceptance:** unit test `TestConditionalNilPredicateDefaultsToShow` — build with `NewConditional(inner, nil)`, call SetPrior(nil), call Init, assert `Done()` mirrors inner's Done (i.e. predicate defaulted to show path, not skip).

### Step 7 — ConditionalStep predicate re-evaluation on Esc-back

**Files:** `internal/tui/harness/harness.go:284-295`, `internal/tui/harness/steps.go:863-870`.

**Problem:** Esc-back loop calls `Reset()` on each step from new cursor → origCursor, but does NOT call `SetPrior()` first. So when the new cursor lands on a `ConditionalStep`, its `initPrev` still holds the stale snapshot from the original forward traversal. If the user changed upstream picks, the predicate should re-evaluate against the new picks — but the step has no way to see them.

Compounding: `ConditionalStep.Reset()` zeros `skip`/`skipDone` but doesn't re-evaluate the predicate. A forward `Done()` check after Reset returns the stale state until some later path re-invokes Init.

**Fix 7a — harness.go Esc-back loop:** before `Reset()` each step, call `SetPrior(h.priorResults())` if the step implements `priorAware`. That way `initPrev` holds fresh results when `Reset()` fires.

```go
for i := h.cursor; i <= origCursor && i < len(h.steps); i++ {
    if pa, ok := h.steps[i].(priorAware); ok {
        pa.SetPrior(h.priorResults())
    }
    if r, ok := h.steps[i].(resetter); ok {
        if cmd := r.Reset(); cmd != nil && i == h.cursor {
            cmds = append(cmds, cmd)
        }
    }
}
```

**Fix 7b — ConditionalStep.Reset():** re-evaluate predicate using the (now fresh) `initPrev`, and set `skip`/`skipDone` accordingly so `Done()` immediately reflects the re-evaluation. If predicate now says skip → `skip = true`, `skipDone = true`. If predicate now says show → `skip = false`, `skipDone = false`, delegate to inner resetter.

```go
func (c *ConditionalStep) Reset() tea.Cmd {
    c.skip = !c.predicate(c.initPrev)
    if c.skip {
        c.skipDone = true
        return nil
    }
    c.skipDone = false
    if r, ok := c.inner.(resetter); ok {
        return r.Reset()
    }
    return nil
}
```

**Acceptance:** new test `TestEscBackReevaluatesConditional` in `harness_test.go` — steps: [Select, Conditional(inner=Text, predicate=first-selection-was-'a')]. User picks 'a' → Cond shows → Text gets input → Esc-back to Select → picks 'b' → advances forward. Expect: Conditional skips this time. Test the whole harness run, not just the Reset method.

Also keep existing `TestConditionalStepResetReevaluates` passing (it exercises the manual SetPrior→Reset→Init path; with 7b, the Init re-eval at L832 still works).

### Step 8 — Drop duplicate Tech-Lead message in `bonsai add`

**File:** `cmd/add.go:148-154` (in-AltScreen NoteStep) + `cmd/add.go:219-226` (post-harness ErrorDetail).

**Fix:** delete the NoteStep branch entirely at L148-154. Let the LazyGroup return `nil` in the tech-lead-required case — the spinner's predicate (which checks `prev[len-1].(bool)`) will see a string (the agent type) at the tail and return false → spinner skips → harness ends clean. Post-harness ErrorDetail at L219-226 already handles this case on stdout.

Concretely replace L145-155:

```go
// Require tech-lead before adding other agents. The user can still
// pick "tech-lead" here to bootstrap — we only block when the pick
// is a non-tech-lead agent and no tech-lead is installed yet. The
// error surfaces post-harness on stdout (see L219) so it persists
// after AltScreen exits.
if agentType != "tech-lead" {
    if _, hasTechLead := cfg.Agents["tech-lead"]; !hasTechLead {
        return nil
    }
}
```

**Acceptance:** `bonsai add backend` on a project with no tech-lead → user sees exactly ONE "Tech Lead required" panel (the post-harness ErrorDetail), not two.

## Security

> [!warning]
> Refer to `Playbook/Standards/SecurityStandards.md` for all security requirements.

- **Step 1 (CRLF fix):** no security impact. Line-ending normalization is a correctness fix, not a trust boundary.
- **Step 2 (tree labels):** purely cosmetic output — no user input crosses a trust boundary.
- **Step 3 (dedup):** output is `.bonsai.yaml` which the user already owns. No new I/O paths.
- **Step 4 (errors.Join):** **net security improvement** — silent generator failures currently mask permission errors (e.g. `.claude/settings.json` write fails → user never sees hooks are broken). Surfacing errors makes tampering or permission-downgrade attacks more visible.
- **Step 5 (panic recover):** no new attack surface — panic messages in Go can include format-string-interpolated content; the `fmt.Errorf("spinner action panic: %v", r)` interpolates via `%v` which is safe for arbitrary values. Do NOT use `%s` with unsanitized input.
- **Steps 6–8:** no external input handled.

No new secrets, no new network calls, no new file-write paths outside what the generator already writes.

## Verification

Run all of these, each must pass:

- [ ] `make build` — binary compiles clean.
- [ ] `go test ./...` — all tests green, including new tests added in Steps 1, 3, 5, 6, 7.
- [ ] `go vet ./...` — clean.
- [ ] `golangci-lint run` (CI v1.64.8) — clean. **Note:** local `golangci-lint` on this machine is v2.11.4 and errors out with `unsupported version`. Trust CI, or install v1.64.8 locally first.
- [ ] Manual CRLF check: `grep -l $'\r' agent/Sensors/*.sh` on freshly-init'd project returns no matches.
- [ ] Manual tree check: `bonsai init` in a temp dir with `docs_path: station`, then `bonsai add backend` with workspace `src`. Confirm the post-add "Created"/"Updated" panels render two trees rooted at `src/` and `station/` (no cross-leakage).
- [ ] Manual dedup check: `bonsai update` twice in a row on the same project, pick same files. Confirm `.bonsai.yaml` lists have no duplicates.
- [ ] Manual error surfacing: `chmod 000 .claude/` then `bonsai add skill <name>`. Confirm user sees a Warning panel, not silent success.
- [ ] Manual Esc-back: `bonsai add`, pick an agent with Conditional downstream steps, Esc all the way back to the first select, change selection, forward-advance. Confirm skipped/shown state matches new selection.
- [ ] Manual duplicate-message check: fresh project with no tech-lead, `bonsai add backend`. Exactly one "Tech Lead required" panel visible.
- [ ] `git diff --stat` covers every file the plan says it should. No stray formatting churn outside the intended lines.
- [ ] PR description enumerates all 8 items with bullet per file change.
