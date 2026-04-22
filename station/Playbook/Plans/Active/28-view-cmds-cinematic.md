---
tags: [plan, tier-2, uiux]
description: Plan 28 — cinematic UI for read-only view commands (list, catalog, guide) + hide auto-generated completion command.
---

# Plan 28 — View Commands Cinematic

**Tier:** 2 (Feature)
**Status:** Draft
**Agent:** general-purpose (worktree isolation, 3 sequential PRs)
**Source:** User request 2026-04-22 — parallel to Plan 27 (add flow polish). User dogfooding add/update/remove in a separate track; asked tech-lead to take list/catalog/guide redesign to keep visual family consistent with init/add.

---

## Goal

Bring `bonsai list`, `bonsai catalog`, `bonsai guide` into the same visual family as Plans 22 (init) and 23 (add): persistent `initflow` chrome (header + palette + footer + min-size floor), LipGloss panels, and — where appropriate — BubbleTea interaction. Also hide Cobra's auto-generated `completion` subcommand from `bonsai --help` output (power users can still invoke it).

These are **read-only** commands, so no stage rail / process sequence — chrome + panels + (for `catalog` and `guide`) keyboard navigation only. Kanji/kana labels optional per command; rail segment structure from init/add does NOT apply.

Ship in **3 sequential PRs**. Each phase is standalone — no env gate, no legacy parallel path. Current flat output is simple enough to replace directly.

---

## Context

### Why

Init (Plan 22) and add (Plan 23) shipped a consistent cinematic shell: persistent header chrome, palette tokens, enso rail, responsive layout. The three remaining view commands (`list`, `catalog`, `guide`) and the one-liner `completion` hide are the last visible CLI surface that still renders as flat print output. Without this pass, `bonsai --help` + first-run visual impression stays inconsistent.

User-stated scope constraint (2026-04-22): **"not expecting a lot of changes"** — keep each command's rewrite minimal. Favor static-cinematic (chrome + palette + panels, no keys) where interaction adds no real value; reserve BubbleTea tabs/scroll only for `catalog` (browsing) and `guide` (scrollable body).

### Current state (anchored to code read 2026-04-22, main at `a9df552`)

- `cmd/list.go:25-143` — `runList` renders: `tui.Heading(cfg.ProjectName)` + optional scaffolding line + per-agent `tui.TitledPanel(displayName, tui.CardFields(pairs), tui.Water)` + muted counts footer via `fmt.Println`. Non-interactive. No workspace tree per agent. No min-size fallback.
- `cmd/catalog.go:22-132` — `runCatalog` renders 7 static sections (Agents / Skills / Workflows / Protocols / Sensors / Routines / Scaffolding) via `tui.SectionHeader` + `tui.CatalogTable`. `-a/--agent` flag filters. No interaction, no keyboard navigation, no inline-expand.
- `cmd/guide.go:42-98` — `runGuide` picks topic via `tui.AskSelect` (Huh), renders selected topic via `glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(100))`, `fmt.Print`. Wall-of-text; no pagination; hard-coded width 100 overflows narrow terminals.
- `cmd/root.go:28-40` — `rootCmd`; no `CompletionOptions` customization, so Cobra auto-adds the `completion` subcommand visibly in `bonsai --help`.
- `internal/tui/initflow/` — exports reusable chrome: `RenderHeader(version, projectDir, width, safe)`, `RenderEnsoRail(idx, labels, width, safe)`, `RenderFooter(keys, width)`, `RenderMinSizeFloor(w, h)`, `ClampColumns(termW)`, `Viewport`, `PanelContentWidth=84`, `PanelWidth(termW)`, design tokens (`FocusedNameStyle`, `ValueStyle`, `LabelStyle`, etc.), `Stage` base + `StageContext`, `KeyHint`, `DefaultKeys(canGoBack)`, `WideCharSafe()`, `TerminalTooSmall(w, h)`, `MinTerminalWidth=70`, `MinTerminalHeight=20`. All additive; safe to extend.
- `internal/tui/addflow/` — `BranchesStage`-like tabbed pattern in `internal/tui/initflow/branches.go` is the reference implementation for `catalog`'s tabbed browser (tabs × category × list × inline-expand with `?`).
- `bubbles v1.0.0` dep already in `go.mod:12` — `viewport` sub-package available for `guide`'s scroll body.

### Design decisions locked

| # | Decision | Picked |
|---|----------|--------|
| 1 | Package location | Three new packages: `internal/tui/listflow/`, `internal/tui/catalogflow/`, `internal/tui/guideflow/`. Mirror `addflow/`'s shape (package file + per-component file + test file). Import `initflow` for chrome + tokens. **Do not lift shared code to a common package yet** — three consumers is below the lift threshold; revisit when a 4th command reuses the shape. |
| 2 | Stage rail | **Omitted from all three.** Rail encodes a mutation process; these are read-only. Chrome = header + footer only. `Stage.SetRailHidden(true)` on each stage. |
| 3 | Kanji labels | **Per-command subtitle in header** only (no rail, no per-stage label). `list` → `覧 RAN LIST`, `catalog` → `録 ROKU CATALOG`, `guide` → `導 DŌ GUIDE`. ASCII fallback via `WideCharSafe()` snapshot at ctor: `LIST`, `CATALOG`, `GUIDE` (no kanji). |
| 4 | Header action label | Extend `initflow.RenderHeader` to accept a command-label parameter (currently hardcoded `"INIT"`). New signature: `RenderHeader(version, projectDir string, action string, width int, safe bool)`. Existing init callers pass `"INIT"` verbatim; add/list/catalog/guide pass their own. Backwards-compat: none needed — all callers under our control. |
| 5 | Right-block label | Currently `"PLANTING INTO"` — also hardcoded in `RenderHeader`. Extend signature further: `rightLabel string`. `init` keeps `"PLANTING INTO"`, `list`/`catalog`/`guide` use `"IN"` (neutral). `add` (Plan 27 may rename independently) uses current default. |
| 6 | `list` — interaction | **Static cinematic.** No keyboard. Render chrome header + optional scaffolding row + per-agent `TitledPanel` (reuses `tui.CardFields`) + workspace file-tree via `tui.FileTree` if that workspace is on disk (live scan of installed files) + counts footer row. No tabs, no scroll — output sized to fit (truncation if terminal too small, fallback to `initflow.RenderMinSizeFloor`). |
| 7 | `list` — workspace tree | Under each agent panel, render `tui.FileTree` of the agent's `Workspace/` directory if it exists on disk. If not (never generated yet), skip the tree and show a single-line `tui.Hint("Run: bonsai init")`. This addresses the current gap — `list` today only shows metadata, not what's on disk. |
| 8 | `catalog` — interaction | **Interactive BubbleTea.** Clone `BranchesStage` pattern: 7 tabs (Agents / Skills / Workflows / Protocols / Sensors / Routines / Scaffolding), `← → / h l` tab cycle (wraps), `↑ ↓ / j k` focus clamp (no wrap), `?` toggles inline-expand on focused row. `-a/--agent` flag pre-applies filter by setting `selected agent` state at ctor and greying tabs with zero items for that agent. `enter` or `q` or `esc` or `ctrl-c` exits cleanly. |
| 9 | `catalog` — inline-expand detail block | For the focused row, render a details box below the list. Show the fields the category has: Agents, Required, plus (Sensors) Event/Matcher, (Routines) Frequency, (Scaffolding) If-Removed. No ABOUT/FILE/Content-body rendering in Phase 1 — reserve for future enhancement. |
| 10 | `guide` — interaction | **Hybrid: static picker + BubbleTea viewport.** If arg provided (`bonsai guide quickstart`), skip picker. Render AltScreen with chrome header + 4 topic tabs across top + glamour-rendered body in a `bubbles/viewport` below. `↑ ↓ / j k / pgup / pgdn` scrolls body; `tab / shift+tab / ← →` cycles topic (re-renders body); `q / esc / ctrl-c` exits. |
| 11 | `guide` — glamour width | Bound to live viewport inner width (not hardcoded 100). Re-render glamour output on `WindowSizeMsg` — glamour is fast enough (tens of ms for 5 KB markdown). |
| 12 | `guide` — topic order | Unchanged: `quickstart`, `concepts`, `cli`, `custom-files`. Order drives tab order. |
| 13 | `completion` — hide strategy | `rootCmd.CompletionOptions.HiddenDefaultCmd = true` in `cmd/root.go` init block. Keeps command functional (`bonsai completion zsh > ...` works) but drops it from `bonsai --help` listing. One-liner. |
| 14 | Min-size floor | All three flows call `initflow.RenderMinSizeFloor(w, h)` when `TerminalTooSmall(w, h)`. Same 70×20 threshold as init/add. |
| 15 | `--no-color` support | All three consume existing `tui.DisableColor()` pathway — no new work needed. |
| 16 | Non-TTY fallback | `list` works fine as static output. `catalog` + `guide` are AltScreen programs; if stdout is not a TTY (piped to `less`, CI), fall back to static one-shot rendering (same output shape `list` uses). Detection: `!isatty.IsTerminal(os.Stdout.Fd())` (same pattern init/add use via harness — verify during Phase 1). |

### Non-goals

- Touching `bonsai init` / `bonsai add` / `bonsai remove` / `bonsai update` flows (Plan 27 is the parallel track).
- Rewriting the catalog data model or `meta.yaml` schema.
- Adding search or fuzzy-filter to `catalog` (reserve for a follow-up once base cinematic lands).
- Extending Glamour styles beyond what `glamour.WithAutoStyle()` provides.
- Any content changes to `docs/{quickstart,concepts,cli,custom-files}.md`.

---

## Dispatch Strategy

**3 sequential PRs** — each phase is independent user-facing work but all three extend `initflow.RenderHeader` in the same way (Phase 1 lands that extension; Phase 2 + 3 consume it). Sequential avoids agent-worktree merge friction on `initflow/chrome.go`.

- **Phase 1 — `catalog` cinematic.** New `internal/tui/catalogflow/` + `initflow.RenderHeader` signature extension + `cmd/catalog.go` rewire. Highest-complexity phase (BubbleTea tabbed browser). Also ships the `completion` hide as a one-liner in `cmd/root.go`.
- **Phase 2 — `list` cinematic.** New `internal/tui/listflow/` + `cmd/list.go` rewire. Static cinematic — no BubbleTea.
- **Phase 3 — `guide` cinematic.** New `internal/tui/guideflow/` + `cmd/guide.go` rewire. BubbleTea viewport for scroll + topic tabs.

Each phase is dispatched to a single general-purpose agent with worktree isolation. Agent creates draft PR; tech-lead self-reviews, dispatches independent code-review agent, merges if clean.

Rationale for bundling completion-hide into Phase 1: it's a one-liner touching `cmd/root.go`, doesn't belong in a standalone PR, and naturally lands with the first phase that edits the chrome/CLI surface.

---

## Phase 1 — catalog cinematic + completion hide

**Scope:** `catalogflow/` package with tabbed BubbleTea browser; `cmd/catalog.go` rewire; `initflow.RenderHeader` signature extension; `completion` subcommand hide.

### Files touched

- `internal/tui/catalogflow/catalogflow.go` — **new.** Package doc. Shared types: `Entry struct { Name, DisplayName, Description string; Meta map[string]string; Agents string; Required string }` (generic across categories — per-category extras packed into `Meta` keyed by label, e.g. `"Event"`, `"Frequency"`, `"If Removed"`).
- `internal/tui/catalogflow/browser.go` — **new.** `BrowserStage` BubbleTea model. Embeds `initflow.Stage` (chromeless, railHidden=true). Fields: `categories []category`, `catIdx int`, `itemIdx map[int]int`, `expanded bool`, `viewport initflow.Viewport`. Keys: `← → / h l` tab cycle wraps, `↑ ↓ / j k` clamp, `?` inline-expand toggle, `q / esc / ctrl-c / enter` quit. Header action label `"CATALOG"`. Subtitle kanji `録 ROKU CATALOG` / ASCII `CATALOG`.
- `internal/tui/catalogflow/browser_test.go` — **new.** Cover: tab cycle wraps both directions; focus clamp at bounds; `?` toggles expand; filtered categories (`-a <agent>`) — tabs with 0 items greyed but not removed (shows user what their agent filter excludes); renders correctly under `initflow.TerminalTooSmall`; exit on each quit key.
- `internal/tui/catalogflow/entry.go` — **new.** Row renderer for each category's items. Focus-tinted leaf border (same `FocusedNameStyle` pattern as Branches). Inline-expand block renders the `Meta` dict as labelled rows (LABEL in `LabelStyle`, value in `ValueStyle`).
- `internal/tui/catalogflow/entry_test.go` — **new.** Cover: required glyph rendering, focus border, expand block line count.
- `internal/tui/initflow/chrome.go` — **modified.** Extend `RenderHeader` signature: add `action string` and `rightLabel string` parameters after `projectDir`. Replace two hardcoded strings in the function body. Update the sole init-side caller. Document: `action` is the uppercase command label shown on row 2 (e.g. `"INIT"`, `"CATALOG"`); `rightLabel` is the destination-context label (e.g. `"PLANTING INTO"`, `"IN"`). Empty `rightLabel` hides the right-block row 1 entirely (useful for commands that don't have a "destination").
- `internal/tui/initflow/chrome_test.go` — **modified.** Update existing callers; add two new test cases exercising the new params (custom action, empty rightLabel hides right row 1).
- `internal/tui/addflow/*.go` (any `RenderHeader` callers) — **modified.** Pass the existing action string (`"ADD"` or whatever Plan 27 lands) explicitly. Grep for all current `RenderHeader` call-sites and thread the new args through. **Careful:** Plan 27 is in-flight on a parallel branch; the merge-forward when Phase 1 of this plan lands on main will show these as conflicts against Plan 27's WIP. Coordinate by making Phase 1 of this plan strictly additive to the signature with sensible defaults if the Plan 27 branch has its own pending caller changes — but since we cannot branch off uncommitted Plan 27 work, we ship Phase 1 against current main and Plan 27 rebases when it merges.
- `cmd/catalog.go` — **rewrite.** Replace `runCatalog` body with: (1) load catalog (unchanged), (2) read `-a` flag, (3) check TTY — if non-TTY, call existing static-render path (factor current body into helper `renderCatalogStatic(cat, agentFilter)`), (4) else build `BrowserStage` via `catalogflow.NewBrowser(cat, agentFilter)` and run through a minimal `harness.Run([]harness.Step{stage})`.
- `cmd/catalog_test.go` — **new.** Static-render path: build in-memory catalog, capture stdout, assert all 7 sections + counts render. Optionally a PTY-free BubbleTea test harness call for the interactive path (out of scope Phase 1 — list as follow-up).
- `cmd/root.go` — **modified.** In `init()` at line 38, add after `rootCmd.PersistentFlags().Bool("no-color", false, ...)`: `rootCmd.CompletionOptions.HiddenDefaultCmd = true` + doc comment ("Hide auto-generated `completion` subcommand from --help; it remains functional for users who invoke it directly.").

### Steps

1. **Extend `initflow.RenderHeader`.** Add `action` and `rightLabel` params. Update body to use them instead of `"INIT"` / `"PLANTING INTO"` literals. Update test file to cover the new params. Update `cmd/init_flow.go`'s sole caller to pass `"INIT"` + `"PLANTING INTO"` explicitly. Grep for all other call-sites in `internal/tui/addflow/**` and update them to pass whatever literal they currently produce (preserve behavior).

2. **Create `internal/tui/catalogflow/` package.** Skeleton: `catalogflow.go` with `Entry` + `Category` types + `NewBrowser(cat *catalog.Catalog, agentFilter string) *BrowserStage` ctor; `browser.go` with `BrowserStage` embedding `initflow.Stage`; `entry.go` with the per-row renderer.

3. **Category builder.** Ctor walks all 7 catalog sections, builds `[]Category`. Each category has `name` (e.g. `"Agents"`), `kanji` + `kana` + `ascii` (optional if we decide to label each tab), and `entries []Entry`. Per-category field packing into `Meta`:
   - Agents: `Meta = nil` (just Name + Description).
   - Skills / Workflows / Protocols: `Meta = nil` (Agents + Required render from their own fields).
   - Sensors: `Meta = {"Event": event, "Matcher": matcher}` (matcher hidden if empty).
   - Routines: `Meta = {"Frequency": freq}`.
   - Scaffolding: `Meta = {"If Removed": affects}`.
   Apply `-a <agent>` filter via existing `cat.SkillsFor(agent)` / `cat.WorkflowsFor(agent)` etc. Tabs with zero filtered items render with a greyed label suffix `(0)` — do NOT drop the tab (user needs to see what's being filtered out).

4. **Browser BubbleTea model.** Implement `Init`, `Update`, `View`. View composes:
   - `initflow.RenderHeader(version, cwd, "CATALOG", width, safe)` (rightLabel empty → omits right-row-1; just shows cwd on right-row-2).
   - Tab strip row — 7 tabs separated by two spaces, active tab bold `ColorPrimary`, others `ColorMuted`, tab count suffix `(N)`.
   - List rows via `entry.go` renderer. Focus row gets leaf-border `│ ` prefix. `viewport` scrolls if entries overflow available body rows.
   - Inline-expand block below the focused row when `expanded = true`.
   - `initflow.RenderFooter(keys, width)` — keys = `[← → tabs] [↑ ↓ focus] [? details] [q quit]`.
   Use `SetRailHidden(true)` on the embedded `Stage` so `renderFrame` skips the rail row.

5. **Rewire `cmd/catalog.go`.** Factor current body into `renderCatalogStatic(cat *catalog.Catalog, agentFilter string) error`. New `runCatalog` checks `!isatty.IsTerminal(os.Stdout.Fd())` — if non-TTY, call the static helper and return. Else build the `BrowserStage` and run via `tea.NewProgram(stage, tea.WithAltScreen()).Run()` (NOT the harness — harness is for sequential stages; catalog is single-stage, no advance).

6. **Hide `completion` subcommand.** In `cmd/root.go` init block, add `rootCmd.CompletionOptions.HiddenDefaultCmd = true` with a one-line doc comment. Verify `bonsai --help` no longer lists `completion` but `bonsai completion zsh` still produces output.

### Verification

- [ ] `make build` — binary builds clean.
- [ ] `go test ./...` — all tests pass, including new `catalogflow/*_test.go` and any `initflow/chrome_test.go` updates.
- [ ] `gofmt -l .` — no diff.
- [ ] `go vet ./...` — clean.
- [ ] `bonsai catalog` — opens AltScreen; tab cycle works both directions; `?` expands inline details; `q` exits cleanly.
- [ ] `bonsai catalog -a tech-lead` — filter applies; greyed tabs for categories with zero matches.
- [ ] `bonsai catalog > /tmp/out.txt` — non-TTY fallback renders static 7 sections (current behavior preserved).
- [ ] `bonsai --help` — `completion` no longer appears in Available Commands.
- [ ] `bonsai completion zsh | head` — still emits completion script (command still callable).
- [ ] Terminal <70×20 — renders `RenderMinSizeFloor` instead of a broken tab strip.

---

## Phase 2 — list cinematic

**Scope:** `listflow/` package with static cinematic render; `cmd/list.go` rewire.

### Files touched

- `internal/tui/listflow/listflow.go` — **new.** Package doc + `RenderAll(cfg *config.ProjectConfig, cat *catalog.Catalog, version, projectDir string) string` pure-function entry point. Returns the full rendered string (header chrome + agent panels + counts footer). No BubbleTea, no Update loop — static string builder.
- `internal/tui/listflow/agent_panel.go` — **new.** Per-agent panel renderer. Builds the `CardFields` pair list for the agent + renders `tui.FileTree(scanAgentWorkspace(absPath), root)` under the panel if the workspace dir exists on disk. Walks the workspace dir via `filepath.WalkDir` (skip hidden files, skip node_modules / .git style dirs — define small `skipDir` predicate).
- `internal/tui/listflow/agent_panel_test.go` — **new.** Cover: panel renders with no workspace (no tree, `bonsai init` hint), with empty workspace (tree shows `(empty)`), with populated workspace (tree renders top N files, truncation if over 50 entries).
- `internal/tui/listflow/listflow_test.go` — **new.** Cover: empty config (no agents) renders empty-state panel + `bonsai add` CTA; one agent renders one panel; multi-agent renders all panels sorted alphabetically by agent name; counts footer matches reality.
- `cmd/list.go` — **rewrite.** Replace `runList` body with: (1) `mustCwd()` + `requireConfig` (unchanged), (2) `loadCatalog()` (unchanged), (3) `fmt.Print(listflow.RenderAll(cfg, cat, Version, cwd))` — single write, no intermediate `tui.Blank()` / `tui.Info()` / `fmt.Println` scattering. Remove the per-agent `tui.TitledPanel` + `tui.CardFields` calls (lifted into `listflow/agent_panel.go`).
- `cmd/list_test.go` — **new.** E2E: temp dir with synthetic `.bonsai.yaml` + one agent workspace, capture stdout, assert header renders, agent panel renders, counts footer matches.

### Steps

1. **Create `internal/tui/listflow/` package.** Skeleton with `RenderAll` entry point.

2. **Header chrome.** Call `initflow.RenderHeader(version, projectDir, "LIST", width, safe)` with empty `rightLabel` (project dir renders on its own in row 2). Width comes from `lipgloss.Width(strings.Repeat(" ", initflow.ClampColumns(termWidth).Total))` pattern — factor into a helper if repeated.

3. **Empty-state panel.** When `len(cfg.Agents) == 0`, render `tui.EmptyPanel("No agents installed.\nRun bonsai add to get started.")` verbatim (reuses existing helper).

4. **Scaffolding row.** If `len(cfg.Scaffolding) > 0`, render a single muted row with scaffolding names (comma-separated, display-cased via `catalog.DisplayNameFrom`). Below header, above agent panels.

5. **Per-agent panels.** For each agent (sorted by name), render:
   - `tui.TitledPanel(displayName, tui.CardFields(pairs), tui.Water)` — same as current.
   - Below the panel (not inside it), `tui.FileTree(scanAgentWorkspace(agent.Workspace, projectDir), agent.Workspace)`. Panel and tree are visually stacked with a single blank line between.
   - If the workspace dir does not exist on disk, render `tui.Hint("Run: bonsai init")` below the panel instead.

6. **Counts footer.** Same as current — muted single-line summary with agent/skill/workflow/protocol/sensor/routine counts separated by `tui.GlyphDot`.

7. **Min-size floor.** If `initflow.TerminalTooSmall(termW, termH)`, return `initflow.RenderMinSizeFloor(termW, termH)` and nothing else.

8. **Rewire `cmd/list.go`.** Delete the per-agent inline rendering; single `fmt.Print(listflow.RenderAll(...))` call replaces it. Keep `mustCwd` + `requireConfig` + `loadCatalog` calls unchanged.

### Verification

- [ ] `make build` — clean.
- [ ] `go test ./...` — all tests pass, including new `listflow/*_test.go` and `cmd/list_test.go`.
- [ ] `gofmt -l .` — no diff.
- [ ] `bonsai list` in a real project — header renders; per-agent panels stack cleanly; workspace file-tree renders under each agent with real files; counts footer accurate.
- [ ] `bonsai list` in a brand-new dir with no `.bonsai.yaml` — same `FatalPanel` error as before (unchanged).
- [ ] `bonsai list` with config but zero agents — empty-state panel with `bonsai add` CTA.
- [ ] Terminal <70×20 — `RenderMinSizeFloor` renders; no broken truncation.
- [ ] `bonsai list --no-color` — palette disabled; everything still reads.
- [ ] `bonsai list | cat` — non-TTY output is still readable (no ANSI escape leaks — LipGloss handles this via `DisableColor` when non-TTY, confirm).

---

## Phase 3 — guide cinematic

**Scope:** `guideflow/` package with tabbed BubbleTea viewport + glamour integration; `cmd/guide.go` rewire.

### Files touched

- `internal/tui/guideflow/guideflow.go` — **new.** Package doc + shared types (`Topic struct { Key, Label, Kanji, ASCII, Markdown string }`).
- `internal/tui/guideflow/viewer.go` — **new.** `ViewerStage` BubbleTea model. Embeds `initflow.Stage` (chromeless, railHidden=true). Fields: `topics []Topic`, `idx int`, `viewport viewport.Model` (from `bubbles/viewport`), `rendered map[int]string` (cached glamour output per topic+width). Keys: `tab / shift+tab / ← →` cycle topic; `↑ ↓ / j k / pgup / pgdn` scroll body; `g / home` top, `G / end` bottom; `q / esc / ctrl-c` quit.
- `internal/tui/guideflow/viewer_test.go` — **new.** Cover: topic cycle wraps both directions; `idx` starts at 0 (or at arg-provided topic key); `WindowSizeMsg` triggers glamour re-render; `?` is NOT bound (reserve); quit on each exit key; min-size floor passes through.
- `internal/tui/guideflow/render.go` — **new.** `renderMarkdown(content string, width int) (string, error)` helper — factored from `cmd/guide.go:renderMarkdown`. Strips YAML frontmatter if present, creates `glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(width))`, returns rendered string. Cached in `ViewerStage.rendered` keyed by `fmt.Sprintf("%d:%d", topicIdx, width)` so re-renders only fire on width or topic change.
- `internal/tui/guideflow/render_test.go` — **new.** Cover: frontmatter strip; narrow vs wide render produces different output; invalid glamour config returns error.
- `cmd/guide.go` — **rewrite.** Replace `runGuide` body:
  - If arg provided, look up topic by key (same validation as now).
  - Build `ViewerStage` via `guideflow.NewViewer(topics, initialTopicKey)` where `topics` is built from the `guideContents` map in the same order as `guideTopics` (preserves the existing 4-topic sequence).
  - Check TTY — if non-TTY, fall back to current glamour-print behavior (factor into `renderStatic(key, content) error`).
  - Else run `tea.NewProgram(stage, tea.WithAltScreen()).Run()`.
  - No Huh picker — topic selection is now the tab strip in the AltScreen viewer. Arg-less invocation opens on the first topic.
- `cmd/guide_test.go` — **new.** Cover the static fallback path (non-TTY).

### Steps

1. **Create `internal/tui/guideflow/` package.** Skeleton: `guideflow.go`, `viewer.go`, `render.go`.

2. **Markdown cache.** `renderMarkdown(content, width)` must be called on `WindowSizeMsg` when width changes. Cache prior renders per (topicIdx, width) to avoid re-rendering unchanged combos on every key event. Width bound is `viewport.Width` = terminal width minus padding (ClampColumns math).

3. **Viewer model.** `Init` — build first-render cache entry, return `nil` cmd. `Update`:
   - `tea.WindowSizeMsg` → update `Stage.SetSize(w, h)`, compute body area (terminal height − chrome rows − tab strip row), call `viewport.SetWidth/SetHeight`, re-render current topic through `renderMarkdown`, push output to `viewport.SetContent`.
   - `tea.KeyMsg` → delegate scroll keys to `viewport.Update`; topic-cycle keys update `idx` and re-render.
   `View` — compose: header + tab strip + viewport + footer. Tab strip active tab bold primary, others muted.

4. **Tab strip.** Same visual style as catalog's tab strip. 4 tabs horizontally. No count suffix. Kanji per topic: `quickstart → 発 HATSU QUICKSTART`, `concepts → 念 NEN CONCEPTS`, `cli → 令 REI CLI`, `custom-files → 己 KO CUSTOM`. ASCII fallback = English only.

5. **Rewire `cmd/guide.go`.** Delete `tui.AskSelect` call and `renderMarkdown` function (moved to `guideflow/render.go`). New `runGuide`: validate arg (unchanged), TTY check, launch viewer or static fallback.

6. **Argless invocation.** `bonsai guide` (no arg) opens the viewer on topic index 0 (`quickstart`). User can cycle via tabs. This replaces the Huh picker — user sees the first topic immediately, switches with tab if they want a different one. Net win: faster to first content; all topics discoverable via tab strip.

### Verification

- [ ] `make build` — clean.
- [ ] `go test ./...` — all tests pass including new `guideflow/*_test.go` and `cmd/guide_test.go`.
- [ ] `gofmt -l .` — no diff.
- [ ] `bonsai guide` — opens AltScreen on quickstart; tab cycles to concepts/cli/custom-files and back; up/down arrows scroll long content; `q` exits.
- [ ] `bonsai guide concepts` — opens directly on concepts.
- [ ] `bonsai guide bogus-topic` — errors with current error message unchanged.
- [ ] `bonsai guide > /tmp/out.md` — non-TTY fallback renders current behavior (Huh picker won't launch in non-TTY; decide: error out with "guide requires a TTY or explicit topic arg", or pipe first topic). **Decision:** if no arg + no TTY → error with "bonsai guide: specify a topic when piping output (quickstart, concepts, cli, custom-files)". If arg + no TTY → render that topic as current glamour output, same as today.
- [ ] Terminal <70×20 — `RenderMinSizeFloor` renders.
- [ ] Resize terminal mid-view — glamour re-wraps on `WindowSizeMsg`.

---

## Dependencies

- `internal/tui/initflow/` — consumed by all three phases. Phase 1 extends `RenderHeader` signature; Phases 2 + 3 consume the extended form. Ship Phase 1 first.
- `bubbles v1.0.0` — `viewport` sub-package, already in `go.mod`.
- `glamour` — already in `go.mod`, currently used in `cmd/guide.go`.
- No new external dependencies.
- **Plan 27 coordination:** if Plan 27's branch adds new `initflow.RenderHeader` callers or renames any exports before my Phase 1 lands, I rebase on top of Plan 27's merged state. If I land first, Plan 27 rebases to pick up the new `action` + `rightLabel` params. Either order works — the extension is backwards-safe via sensible defaults on the new params if needed. **Preferred order:** ship Plan 28 Phase 1 first (~small diff, pure additive), Plan 27 rebases.

---

## Security

> [!warning]
> Refer to `Playbook/Standards/SecurityStandards.md` for all security requirements.

Scope is entirely local TUI rendering — no network, no secrets, no new external inputs. Specific security checkpoints:

- **No user input reaches exec/shell.** Tab keys, scroll keys, topic selection — all routed through BubbleTea's own event loop; no `exec.Command` in any new code.
- **Workspace file-tree in `list`** (Phase 2) walks user-owned directories via `filepath.WalkDir`. Must guard against symlink loops (use `fs.SkipDir` on symlinks pointing outside the workspace root) and path-traversal if the workspace path in `.bonsai.yaml` contains `..` (validate via `filepath.Clean` + ancestor-check against `projectDir`). **Failure mode:** malicious `.bonsai.yaml` setting `workspace: ../../etc` → we should refuse to walk outside projectDir. Surface as a warning in the panel, do not walk.
- **Glamour markdown renderer** (Phase 3) operates on embedded bundled content only (`//go:embed docs/*.md`). No user-supplied markdown. No XSS / injection surface.
- **Cobra completion hide** is a display-only change. `completion` subcommand remains callable — no capability loss, no privilege change.
- **No new dependencies** → no new CVE surface.

---

## Verification (cross-phase)

After all three phases merge:

- [ ] `make build && go test ./...` on main — clean.
- [ ] `bonsai --help` — shows init / add / catalog / guide / list / remove / update / version / help. No `completion`.
- [ ] `bonsai completion zsh | head` — still functional.
- [ ] `bonsai list`, `bonsai catalog`, `bonsai guide` — all three render cinematic chrome consistent with `bonsai init`'s visual family.
- [ ] Piped invocations (`bonsai list > x`, `bonsai catalog > x`, `bonsai guide quickstart > x`) — all produce clean non-ANSI static output suitable for non-TTY consumption.
- [ ] No regression in any Plan 27 track functionality — run `bonsai add` / `bonsai init` / `bonsai remove` / `bonsai update` once each post-merge to sanity-check.
