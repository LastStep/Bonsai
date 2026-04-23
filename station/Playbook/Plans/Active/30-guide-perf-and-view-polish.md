---
Status: Draft — 2026-04-23
Plan: 30
Title: Guide viewer perf + list/catalog/guide polish bundle
Agent: general-purpose
Tier: 2
---

# Plan 30 — Guide Viewer Perf + List/Catalog/Guide Polish Bundle

**Tier:** 2
**Status:** Draft
**Agent:** general-purpose

## Goal

Make `bonsai guide` tab-switches feel instant (currently 4× slow because `glamour.NewTermRenderer(WithAutoStyle())` runs on every switch — each construction triggers `termenv.HasDarkBackground()` OSC query + goldmark+chroma init). Bundle two previously-filed "batch one-shot" backlog cleanups (Plan-28 Phase 1 NITs for catalog; Plan-28 Phase 2+3 NITs for list + guide) that touch the same three packages.

## Context

- User-visible complaint: *"bonsai guide command is very slow. going across guides at start takes a lot of time. this should be instant."*
- Current cache in `internal/tui/guideflow/viewer.go:246` (`refreshViewportContent`) is keyed `idx:width` and skips glamour only on **repeat** visits. First visit to each of the 4 tabs pays the full construction cost.
- Root cause (confirmed from glamour source at `~/go/pkg/mod/github.com/charmbracelet/glamour@v1.0.0/glamour.go:70-99` + `:306-322`):
  1. `WithAutoStyle()` resolves via `getDefaultStyle(AutoStyle)` → `termenv.HasDarkBackground()` OSC 11 query (100ms+ timeout common).
  2. `NewTermRenderer` also builds goldmark extensions, ANSI renderer, chroma dependencies each call.
  3. WordWrap is frozen at construction — a per-width renderer is required, but nothing forces per-topic renderers.

- **Backlog bundling rationale:** Three backlog items all tagged "batch one-shot" and all scoped to the same three packages (`guideflow`, `listflow`, `catalogflow`). Shipping together in one worktree is cheaper than three separate PRs and avoids the worktree-held-branch pattern that has hit 12× this month.
  - `[Plan-28-cosmetic] Phase 1 review NITs` — catalog: `timeZero()` wrapper, stale `filterHidZero` docstring, redundant zero-value field, Agents details metadata gap, greyed-tab style test gap.
  - `[Plan-28-cosmetic] Phase 2+3 post-fix NITs` — list: `itoa` duplication, empty-workspace test gap, empty-agents/non-TTY test gap. Guide: `chromeRows` magic constant, `labelFor`/`shortFor` inconsistency, wasted `NewStage` label args.

## Phases

### Phase A — Guide viewer perf (the user's ask)

Two independent changes in `internal/tui/guideflow/`:

**A1. Cache `*glamour.TermRenderer` per width in `ViewerStage`.**

- Add field `renderers map[int]*glamour.TermRenderer` to `ViewerStage` struct (`viewer.go:25-35`), initialised in `NewViewer` alongside `rendered`.
- New method `(s *ViewerStage) rendererFor(width int) (*glamour.TermRenderer, error)`: returns cached renderer for width or builds+caches a fresh one via `glamour.NewTermRenderer(WithAutoStyle(), WithWordWrap(width))`.
- `renderMarkdown` in `render.go` **split into two functions**:
  - Keep `renderMarkdown(content string, width int)` as the existing ctor-and-render one-shot (still used by `cmd/guide.go:78` `renderStatic` non-TTY path — keep that path unchanged).
  - Add `renderMarkdownWith(content string, renderer *glamour.TermRenderer)` that strips frontmatter and calls `renderer.Render(body)`.
- `refreshViewportContent` (`viewer.go:246`) swaps the direct `renderMarkdown(..., w)` call for:
  ```go
  r, err := s.rendererFor(w)
  if err != nil { ... }
  rendered, err := renderMarkdownWith(s.topics[s.idx].Markdown, r)
  ```
- **Rationale:** per-width renderer construction moves from O(tabs × widths) to O(widths). On a fixed-size terminal that's **1 construction** for the whole session, and every tab-switch drops to ~1-5ms pure goldmark conversion.

**A2. Pre-warm remaining topics on first `WindowSizeMsg` via `tea.Cmd`.**

- New message type `preWarmMsg struct { width int; results map[int]string }` in `viewer.go`.
- New method `(s *ViewerStage) preWarmCmd(width int) tea.Cmd`: returns a `tea.Cmd` that, in a goroutine:
  - Builds/reuses renderer at width via `rendererFor` (safe — renderer cache lookup inside `tea.Cmd` is fine because at this point nothing else races; see note below).
  - Iterates `s.topics`, renders each into a local `map[int]string` keyed by topic idx, returns `preWarmMsg{width, results}`.
- `Update` `WindowSizeMsg` branch: after existing `resizeViewport()` + `refreshViewportContent()` calls, if the width changed since last pre-warm, dispatch `s.preWarmCmd(m.Width)` and return it as the tea.Cmd.
- `Update` handles `preWarmMsg`: for each `(idx, body)` pair, store `s.rendered[fmt.Sprintf("%d:%d", idx, msg.width)] = body`. Do not touch viewport content (current topic already rendered synchronously).
- **Race concern:** `tea.Cmd` runs on a goroutine; the returned msg lands back on the Update loop (single-threaded). The cmd function must NOT mutate `s.rendered` or `s.renderers` directly — it returns results via the msg and Update writes them. `rendererFor` is called once synchronously in the goroutine for the pre-warm width; add field `renderersMu sync.Mutex` guarding `rendererFor` OR construct a fresh local renderer inside the cmd to avoid the lock (simpler, small cost).
- **Simpler approach (preferred):** the `preWarmCmd` constructs a fresh `*glamour.TermRenderer` for its width locally (one extra construction on the goroutine), renders all topics, ships them back. Update loop writes results into `s.rendered`. No shared-state mutation from the goroutine. The `rendererFor` cache on the main goroutine stays lock-free.

**A3. Regression tests** (`internal/tui/guideflow/viewer_test.go` — new tests, don't touch existing):

- `TestViewer_RendererCachedPerWidth` — construct viewer, send two WindowSizeMsg with same width, assert `len(s.renderers) == 1` after both.
- `TestViewer_RendererPerDistinctWidth` — two WindowSizeMsgs with different widths, assert `len(s.renderers) == 2`.
- `TestViewer_PreWarmPopulatesCache` — construct viewer, send WindowSizeMsg, execute the returned cmd, send the resulting `preWarmMsg` back into Update, assert `s.rendered` has entries for all topic idx values at that width.
- `TestViewer_TabSwitchUsesCachedRenderer` — seed `s.renderers` with a sentinel, switch tabs, assert no new renderer constructed (len unchanged).

### Phase B — Guide polish (Plan-28 Phase 3 NITs)

**B1. `chromeRows` magic constant → computed from rendered chrome heights.**

- `internal/tui/guideflow/viewer.go:234` — replace `const chromeRows = 8` with:
  ```go
  header := initflow.RenderHeader(s.Stage.version(), ..., s.width, s.EnsoSafe())
  footer := initflow.RenderFooter(s.keyHints(), s.width)
  chromeRows := lipgloss.Height(header) + lipgloss.Height(footer) + 4 // 2 blank separators + tab strip (1) + blank (1)
  ```
  Pattern mirrors `Stage.renderFrame` (`internal/tui/initflow/stage.go:258-259`).
- **Caveat:** `Stage.version`/`Stage.ensoSafe` are unexported. Options: (a) add a new accessor `Stage.Chrome() (header, footer string)` returning the two pre-rendered strings at current dims, (b) duplicate the render call in `resizeViewport`. Prefer (a) — one new Stage method:
  ```go
  func (s *Stage) ChromeHeights(keys []KeyHint) (headerH, footerH int) {
      width := s.width; if width <= 0 { width = 80 }
      headerH = lipgloss.Height(RenderHeader(s.version, s.projectDir, s.headerAction, s.headerRightLabel, width, s.ensoSafe))
      footerH = lipgloss.Height(RenderFooter(keys, width))
      return
  }
  ```
  In `resizeViewport`:
  ```go
  headerH, footerH := s.ChromeHeights(s.keyHints())
  chromeRows := headerH + footerH + 4
  ```

**B2. `labelFor`/`shortFor` inconsistency for unknown keys.**

- `internal/tui/guideflow/guideflow.go:48-73` — `labelFor` defaults to `strings.ToUpper(strings.ReplaceAll(key, "-", " "))` (space-separated), `shortFor` defaults to `up[:5]` (raw truncation).
- Unify the default fallback so both derive from the same base. New private helper `deriveLabel(key string) string` (hyphen→space, uppercase). `shortFor` default becomes `truncate(deriveLabel(key), 5)`.
- No behaviour change for the 4 known keys (they hit the switch cases before fallback); future additions get consistent fallback.

**B3. Drop wasted `NewStage` label args.**

- `internal/tui/guideflow/viewer.go:59-68` — `NewStage(0, StageLabel{English:"GUIDE"}, "GUIDE", ...)` — both `label` and `title` are superseded by the subsequent `SetHeaderAction("GUIDE")` call and the rail is hidden, so `StageLabel{}` zero-value and `""` title are interchangeable. Change to:
  ```go
  base := initflow.NewStage(0, initflow.StageLabel{}, "", version, projectDir, "", "", time.Time{})
  ```
  Keep the `SetHeaderAction("GUIDE")` call — that's the one that actually renders. Doc-comment the `StageLabel{}` choice: *"Empty label — rail is hidden, title is unused for chromeless viewers. Header text comes from SetHeaderAction."*

### Phase C — List polish (Plan-28 Phase 2 NITs)

**C1. Dedupe `itoa` helper.**

- `internal/tui/listflow/agent_panel.go:364-387` and `internal/tui/initflow/enso.go:247-270` are near-identical (minor buf-size difference). Neither is hot-path; both exist to avoid a `strconv` import.
- **Swap both to `strconv.Itoa`.** Justification: `strconv` is a stdlib package, zero-cost import; keeping two copies of a textbook function because "we don't want to import strconv" is a code-smell. `internal/tui/initflow/chrome.go:202-203` also uses `itoa` — swap that too.
- Call sites:
  - `listflow/agent_panel.go:219` — `"... (" + strconv.Itoa(extra) + " more)"`
  - `initflow/chrome.go:202-203` — swap both `itoa(...)` to `strconv.Itoa(...)`
  - `initflow/enso.go:151` — swap `itoa(i+1)` to `strconv.Itoa(i+1)`
- Add `"strconv"` import, remove the two `itoa` func defs (listflow/agent_panel.go and initflow/enso.go).

**C2. Empty-workspace defensive-branch test.**

- New test in `internal/tui/listflow/agent_panel_test.go`: `TestRenderWorkspaceBlock_EmptyString` — calls `renderWorkspaceBlock("", "/some/project/dir")`, asserts output contains `"Workspace missing"` string (matches the CTA at line 165). Locks the contract that `workspace: ""` never panics and always emits the same hint as missing-directory.

**C3. Empty-agents + non-TTY e2e for `cmd/list`.**

- New test in `cmd/list_test.go`: `TestRunList_NoAgents` — same scaffold as `TestRunList_HappyPath` but with `cfg.Agents = map[string]*config.InstalledAgent{}`. Asserts output still includes "LIST" header + "0 agents" count + exits without error. Ensures the agent-panel loop doesn't panic on empty input.
- `TestRunList_NonTTYFallback` — **skip if happy path already covers it** (non-TTY is the default test environment since stdout is redirected via captureStdout). Leave the test gap note in Backlog if cannot be meaningfully added.

### Phase D — Catalog polish (Plan-28 Phase 1 NITs)

**D1. Drop `timeZero()` wrapper.**

- `internal/tui/catalogflow/browser.go:16-19` — delete the `timeZero()` function + its doc comment.
- `browser.go:126` — replace `timeZero()` call with `time.Time{}`.
- `time` import stays (it's the `time.Time{}` type itself).

**D2. Fix stale `filterHidZero` docstring.**

- `internal/tui/catalogflow/browser.go:21-30` — the `category` struct docstring references a `filterHidZero` field that doesn't exist. Rewrite the docstring to describe only the real fields (`key`, `displayName`, `entries`) and the "(0)" muted rendering behaviour (which lives in the tab-render code, not on the struct).

**D3. Drop redundant `expanded: false`.**

- `browser.go:139` — delete the `expanded: false,` line in the `BrowserStage` struct literal (zero value is false). `internal/tui/initflow/branches.go:234` has the same pattern — drop there too for consistency.

**D4. Agents section entry metadata.**

- `browser.go:156-174` (`buildCategories` Agents block) — entries get `Meta: map[string]string{"Kind": "agent"}` so the details block renders `KIND: agent` instead of `(no extra metadata)`.
- Alternatively: keep `(no extra metadata)` — it's accurate. **Decision (ships in the plan):** add `Meta: map[string]string{"Kind": "agent"}` per the NIT suggestion. Surfaces parity with other sections' metadata blocks.

**D5. Greyed-tab style test.**

- New test in `internal/tui/catalogflow/browser_test.go`: `TestBrowser_ZeroCountTabRendersMuted` — build a `BrowserStage` with an agent filter that yields `(0)` in at least one section, render the tab strip, assert the rendered output contains the muted ANSI escape sequence for `tui.ColorMuted` around the `(0)` suffix. Use `lipgloss.NewStyle().Foreground(tui.ColorMuted).Render("(0)")` as the reference substring OR assert the `(0)` cell is styled differently than the active-count cells via raw-ANSI substring comparison.

**D6. Dead-code sweep post-`RenderFrame` routing.**

- Plan 28 Phase 1 MINOR 3 routed the `View()` body through `Stage.RenderFrame`. Check `internal/tui/catalogflow/browser.go` for any remaining body-padding / truncation helpers that are unused now that `RenderFrame` does it. Run `go vet ./...` + manual grep for unreferenced exported funcs in the package. Drop anything unused.

### Phase E — Verification

Run in worktree after all phase commits land:

```bash
make build
go test ./...
./bonsai guide  # manual smoke test — tab through 4 topics, confirm instant switches
./bonsai list   # manual smoke test — confirm no regressions
./bonsai catalog # manual smoke test — confirm tab strip behaves
./bonsai guide quickstart | head # non-TTY path — confirm unchanged static render
```

All must pass. Any failure is a blocker; fix in the same worktree before PR.

## Dependencies

None. No new Go modules, no schema changes, no cross-package API breaks (the one new exported method `Stage.ChromeHeights` is additive).

## Security

> [!warning]
> Refer to [Playbook/Standards/SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

- **No new input surfaces.** All three commands are read-only (`list`, `catalog`, `guide` never write to disk or execute subprocesses). Plan-28 Phase 2's `filepath.Rel`-based escape guards are untouched.
- **Goroutine safety (Phase A2):** the `preWarmCmd` MUST NOT mutate shared state. It constructs its own renderer locally, returns results as a `tea.Msg`, and Update writes to `s.rendered` on the single-threaded tea loop.
- **No change to glamour style sourcing.** `WithAutoStyle` remains (terminal bg detection is the right UX).

## Verification

- [ ] `make build` — clean
- [ ] `go test ./...` — all pass
- [ ] Phase A tests present and green (4 new tests in `guideflow/viewer_test.go`)
- [ ] Phase B NIT items resolved (chromeRows computed, labelFor/shortFor consistent, NewStage args trimmed)
- [ ] Phase C `itoa` removed from both packages, two new tests in list package
- [ ] Phase D `timeZero` removed, docstring fixed, `expanded: false` removed, Agents Meta added, greyed-tab test added
- [ ] Manual smoke: `./bonsai guide` tab-switching feels instant (< 50ms perceived)
- [ ] Manual smoke: `./bonsai guide quickstart | head` static render unchanged (non-TTY fallback)
- [ ] No new dependencies in `go.mod`
- [ ] `govulncheck ./...` clean
- [ ] `golangci-lint run` (or Makefile `lint` target) clean

## Rollback

Single PR. If issues surface post-merge:
- A1/A2 revert is a pure code swap (per-tab-switch render returns — regression is perf, not correctness).
- B-D are cosmetic — revert any single phase independently via `git revert <phase-commit>` if it causes unexpected behaviour.
