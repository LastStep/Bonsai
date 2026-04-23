---
tags: [log, plan-30, perf, guideflow, catalogflow, listflow]
date: 2026-04-23
---

# 2026-04-23 — Plan 30 guide viewer perf + list/catalog/guide polish bundle

## What shipped

Plan 30 (PR [#74](https://github.com/LastStep/Bonsai/pull/74) squash `7733ae2`) — user-reported guide perf regression fixed + 2 pre-filed Plan-28 cosmetic NIT bundles cleared in same ship.

### Phase A — Guide viewer perf (the user's ask)

User: *"bonsai guide command is very slow. going across guides at start takes a lot of time. this should be instant."*

**Root cause** (confirmed from `~/go/pkg/mod/github.com/charmbracelet/glamour@v1.0.0/glamour.go:70-99` + `:306-322`):

- `refreshViewportContent` called `renderMarkdown` per tab switch.
- `renderMarkdown` called `glamour.NewTermRenderer(WithAutoStyle(), WithWordWrap(width))` on every invocation.
- `WithAutoStyle` → `getDefaultStyle(AutoStyle)` → `termenv.HasDarkBackground()` → OSC 11 terminal query, ~100-200ms on many terminals.
- Plus goldmark extension wiring + chroma init per construction.
- Existing `idx:width` rendered-output cache skipped the glamour call only on **repeat** visits — first visit to each of 4 tabs paid the full construction cost.

**Fix shipped:**

- **A1.** `ViewerStage` gains `renderers map[int]*glamour.TermRenderer` cache keyed by viewport width. New `rendererFor(width)` accessor returns cached renderer or builds-and-caches. `render.go` split: `renderMarkdown` (one-shot, kept for tests) + new `renderMarkdownWith(content, renderer)`. `refreshViewportContent` swaps to the cached path. Per-width construction now O(widths), not O(tabs × widths) — fixed-size terminal session does **1** construction total.
- **A2.** Pre-warm via `tea.Cmd` dispatched on first `WindowSizeMsg` per distinct width (tracked via new `preWarmedWidth int` field). Cmd builds a **local** renderer on the goroutine (no shared-state mutation), renders all topics, returns `preWarmMsg{width, results map[int]string}`. `Update` writes results to `s.rendered` on the tea loop — race-free by construction.
- **A3.** 5 new regression tests: `TestViewer_RendererCachedPerWidth`, `TestViewer_RendererPerDistinctWidth`, `TestViewer_PreWarmPopulatesCache`, `TestViewer_TabSwitchUsesCachedRenderer`, `TestViewer_WidthChangeDispatchesPreWarm` (last one added post-review).

### Phases B-D — Bundled NIT cleanups

Two pre-filed Plan-28 backlog bundles both tagged "batch one-shot" and scoped to the same three packages (`guideflow`, `listflow`, `catalogflow`) — shipping together avoided 3 separate PRs + 3 worktree-held-branch cleanups.

- **B1.** `internal/tui/guideflow/viewer.go` `const chromeRows = 8` magic replaced with `headerH + footerH + 4` computed via new exported `Stage.ChromeHeights(keys []KeyHint) (headerH, footerH int)` on `initflow/stage.go` — mirrors `Stage.renderFrame`'s own chrome math. No more fragility on header/footer row-count changes.
- **B2.** `labelFor`/`shortFor` default fallback unified via new private `deriveLabel(key)` helper. No behaviour change for 4 known keys.
- **B3.** `NewStage(0, StageLabel{English:"GUIDE"}, "GUIDE", ...)` trimmed to `StageLabel{}` + `""` title since `SetHeaderAction("GUIDE")` supersedes both (rail hidden, title unused).
- **C1.** Dedupe `itoa` — dropped from `listflow/agent_panel.go` + `initflow/enso.go`; 4 call sites (`agent_panel.go:219`, `initflow/chrome.go:202-203`, `initflow/enso.go:151`) swapped to `strconv.Itoa`. "Avoid strconv import" rationale was thin; `strconv` is stdlib.
- **C2.** New `TestRenderWorkspaceBlock_EmptyString` locks empty-string → "Workspace missing" hint contract.
- **C3.** New `TestRunList_NoAgents` — zero-agent config renders "LIST" + "0 agents" count without panic.
- **D1.** `catalogflow/browser.go` `timeZero()` wrapper deleted (single call site); `time.Time{}` inlined. `time` import kept (type still needed).
- **D2.** `category` struct docstring rewritten — stale `filterHidZero` reference pointed at a field that never existed.
- **D3.** Redundant `expanded: false` dropped from `browser.go:139` + `initflow/branches.go:234` (zero value).
- **D4.** Agents section `buildCategories` entries gain `Meta: map[string]string{"Kind": "agent"}` — details block renders `KIND: agent` instead of `(no extra metadata)`.
- **D5.** New `TestBrowser_ZeroCountTabRendersMuted` forces truecolor profile via `termenv.SetDefaultOutput`, asserts muted styling around `(0)` suffix.
- **D6.** Dead-code sweep via `go vet` + `deadcode` tool — catalogflow clean, all package-level funcs referenced.

## Process

- Plan 30 spec committed to main (`d49a08e`) before dispatch so worktree agent would see the plan in `station/Playbook/Plans/Active/`.
- Single general-purpose agent dispatched via `isolation: worktree` off main. Returned 4-commit draft PR with `make build && go test ./...` green first pass.
- **CI lint failure** on first push: trailing blank lines at `initflow/enso.go:247` + `listflow/agent_panel.go:364` — leftover from `itoa` func removal. Tech-lead ran `gofmt -w` in worktree + pushed `e3e76fa`. CI re-run 6/6 green (test/lint/Analyze-Go/govulncheck/CodeQL/GitGuardian).
- Independent code-review agent returned **PASS-WITH-MINORS** — 3 non-blocking: (1) stale `renderMarkdown` doc comment claimed usage by `cmd/guide.go renderStatic` which has its own inline renderer, (2) unnecessary `maxInt` helper in new `browser_test.go` (Go 1.25 has builtin `max`), (3) missing width-change dispatch test.
- Fix-agent shipped `f28dd7b` addressing all 3. Subtle test-width pick: `PanelWidth` clamps at 84, so viewer widths 120 and 100 both clamp to 84 — distinct-width branch doesn't fire. Agent used 120 (→84) and 80 (→76) to actually trigger the distinct branch; documented in a test comment.
- Post-merge: worktree-held-branch pattern 13× this month. Manual cleanup: `git worktree remove -f -f .claude/worktrees/agent-a6d462d3` + `git branch -D plan-30-guide-perf-polish worktree-agent-a6d462d3` + `git push origin --delete plan-30-guide-perf-polish`.

Net: +624/−86 across 14 files, 6 commits on the branch (4 phase + 1 gofmt + 1 review-minor fix).

## Open items

None from Plan 30 directly. Two Backlog items cleared:

- `[Plan-28-cosmetic] Phase 1 review NITs` (catalog)
- `[Plan-28-cosmetic] Phase 2+3 post-fix NITs` (list + guide)

## Key decisions

- **Chose shared-renderer cache over single-renderer design.** Word wrap is locked at glamour ctor time — can't change post-construction. Per-width cache is the minimal change that actually works.
- **Chose local renderer in pre-warm goroutine over shared `rendererFor` + mutex.** One extra construction on the pre-warm goroutine vs adding a mutex to the main-loop hot path. Renderer construction is ~100ms once; mutex contention would be forever. Correctness > minor CPU savings.
- **Kept `WithAutoStyle`.** Auto-detecting terminal background is the right UX even if it costs an OSC query. The cache eliminates the repeat-cost; the first-time cost is unavoidable without forcing users to set `GLAMOUR_STYLE`.
