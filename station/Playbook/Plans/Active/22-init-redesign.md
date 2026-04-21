# Plan 22 — `bonsai init` Cinematic Redesign

**Tier:** 2 (Feature)
**Status:** Draft
**Agent:** general-purpose (worktree isolation)
**Source:** User request 2026-04-21 — handoff bundle from claude.ai/design (`Bonsai init - redesign.html` + JSX/CSS + chat transcript). Full design inputs preserved in the chat transcript of this session.

**Review refinements (2026-04-21):** Tech Lead plan review pass applied 7 fixes before dispatch — (1) lock widget name to `RenderFileTree` + `TreeNode` (drop stale `FileTree2`); (2) swap broken `//go:build ignore` demo binary for a `TestRenderFileTree_Demo` that prints via `t.Log`; (3) add missing `strings` import to `fallback.go` example; (4) thread `agentDef.DisplayName` through stage ctors so Observe can show AGENT; (5) capture `startedAt := time.Now()` at `runInitRedesign` entry so PlantedStage ELAPSED clock has a source; (6) Phase 5 test sweep — grep + update/delete tests referencing deleted legacy helpers; (7) clarify BranchesStage `Reset()` is no-op so Esc-and-return preserves picks.

---

## Goal

Replace the `bonsai init` TUI end-to-end with a custom full-screen cinematic flow that matches the handoff design:

- 10 sequential Huh prompts → **4 stages** (器 Vessel · 土 Soil · 枝 Branches · 観 Observe).
- Persistent chrome every frame: top-left enso-ringed brand mark, top-right `PLANTING INTO ~/code/<project>/station/` breadcrumb, 4-dot enso progress rail, footer hairline + context keys.
- Branches merges the five ability pickers (Skills · Workflows · Protocols · Sensors · Routines) into one **tabbed picker** with **inline-expand on the focused row**.
- Generate is a **one-frame reveal**: `max(actual, 600ms)` hold on a drawing-in enso with kanji 生, then cross-fade to Planted.
- Planted replaces the post-init stdout `Success + Hint` with a full-screen view: **file tree of what was written** + summary + 3 next-step commands.
- Kanji + kana labels (器/土/枝/観 + 技/流/律/感/習) with runewidth-safe fallback.
- Palette tightened for this flow: Leaf + Bark as only accents, everything else muted; other semantic tokens (Success/Warning/Info) remain available where a distinction helps users (e.g. conflict counts in Amber, NEW badges in Leaf, REQUIRED in Bark).

Ship in **6 PRs** (5 phases; Phase 5 split into 5A + 5B after initial scoping). Legacy flow stays the default via env flag `BONSAI_REDESIGN=1` until Phase 5B flips default and deletes the legacy path.

---

## Context

### Why

User is dissatisfied with the current `bonsai init` TUI. Handoff bundle from claude.ai/design (chat transcript attached in session) settled on: cinematic linear (Variation A) collapsed to 4 stages, with Variation B's target-location breadcrumb and focused-row detail pane folded in, and a dashboard mock spun off as a future command. Design iterated through: (a) 3 variations, (b) polish pass with zen vocabulary + inline-expand on Branches + `max(actual, 600ms)` one-frame-reveal on generate, (c) ASCII bonsai replaced by file trees (Planted + Observe Soil preview), (d) Observe layout corrected — Vessel+Soil left, Branches right; Soil tree shows scaffolding files only.

### Current state (anchored to code read 2026-04-21)

- `cmd/init.go` declares 10 `harness.Step`s backed by `NewText`/`NewMultiSelect`/`NewLazy`/`NewConditional(NewSpinnerWithPrior)`/`NewLazyGroup`. Each Huh form owns the viewport inside `harness.Run`'s `tea.WithAltScreen` program.
- `internal/tui/harness/harness.go` frames each step with `renderHeader` (banner + action + `[N/M] Title` crumb) and `renderFooter` (↵ continue · esc back · ctrl-c quit). Step `View()` is the body.
- `internal/tui/harness/steps.go` step types (TextStep, MultiSelectStep, ConfirmStep, ReviewStep, LazyStep, LazyGroup, NoteStep, SpinnerStep, ConditionalStep) all wrap `huh.Form` except SpinnerStep. Huh's full-screen takeover precludes the design's persistent chrome + multi-field stages — confirmed by reading steps.go:62-92 (Huh form owns input rendering).
- `internal/tui/styles.go` has the Zen Garden palette (Leaf/Bark/Stone/Water/Moss/Ember/Amber/Sand/Petal) + semantic tokens (ColorPrimary=Leaf, ColorSecondary=Bark, ColorAccent=Petal, ColorInfo=Water, etc.) + `FileTree(files, rootLabel)` flat renderer. Existing `FileTree` is unstyled-per-node and has no status badges — insufficient for Planted/Observe needs.
- `internal/generate/generate.go:161` `WriteResult.Files []FileResult{RelPath, Action, Source}` is the source of truth for Planted's file tree.
- `go.mod` already pulls `github.com/charmbracelet/bubbles v1.0.0` (direct — via huh) and `github.com/mattn/go-runewidth v0.0.23` (indirect — via lipgloss). No new external dependencies required.

### Design decisions locked (from Q&A 2026-04-21)

| # | Decision | Picked |
|---|----------|--------|
| 1 | Scope | **B** — `init` first, other commands in follow-up plans |
| 2 | Framework | **Custom BubbleTea + `charmbracelet/bubbles` primitives (textinput, list, paginator, key.Binding, spinner, viewport)** — Huh dropped from `init`, kept elsewhere |
| 3 | Branches | **A** — one tabbed picker, five categories |
| 4 | Kanji + kana | **A** — full kanji w/ runewidth fallback to ASCII labels when wide chars unsafe |
| 5 | Enso progress | **A** — unicode glyphs `○ ● ◉` + kanji-in-box for current; boxed circle/rail layout (no sixel, no Kitty graphics) |
| 6 | Generate duration | **A** — `max(actualDuration, 600ms)` hold with drawing-in enso ring |
| 7 | Inline `?` expand on Branches | **A** — ships with tabs |
| 8 | Palette tighten | **A** — `init` scope only; other commands keep full palette; extra color permitted where a distinction helps (e.g. warnings, conflict badges, hints) |
| 9 | File tree widget | **A** — build `tui.RenderFileTree` + `tui.TreeNode` (new) — box-drawing + status badges + inline notes + dense mode. Existing flat `tui.FileTree(files, rootLabel)` at `styles.go:462` untouched — it is still used by `cmd/root.go` for `add`/`remove`/`update` panels |
| 10 | `bonsai dashboard` | **A** — separate later plan, not in this rewrite |
| 11 | Ship strategy | **B** — phased (5 PRs) |
| 12 | Esc semantics | **A** — stage-level (Esc at stage N → stage N-1) |
| 13 | Git-remote inheritance for NAME | **B** — plain input, no inference |

---

## Phases

Each phase is an independent PR. Phases 1–4 + 5A land behind `BONSAI_REDESIGN=1` so the default binary runs the legacy flow throughout development; Phase 5B flips the default and deletes the legacy path.

### Phase 1 — FileTree widget + palette audit

Scope: build the reusable file-tree renderer and tighten the `init`-scope palette tokens. No behavioural change to any command yet.

#### Files touched

- `internal/tui/filetree.go` — **new**
- `internal/tui/filetree_test.go` — **new**
- `internal/tui/styles.go` — add Leaf-dim / Rule / Rule-2 accent tokens (for enso rail, borders)

#### Steps

1. **Create `internal/tui/filetree.go`.** Shape driven by design's `file-tree.jsx`.

   ```go
   package tui

   type NodeKind int
   const (
       NodeFile NodeKind = iota
       NodeDir
   )

   type NodeStatus int
   const (
       NodeNormal NodeStatus = iota
       NodeNew                     // Leaf-colored name, NEW badge right-aligned
       NodeRequired                // Bold name, REQUIRED badge right-aligned
       NodeCurrent                 // 2-col leaf-tinted left border + highlight bg
   )

   type TreeNode struct {
       Name     string
       Kind     NodeKind
       Status   NodeStatus
       Note     string          // one-line caption rendered after the name in muted
       Children []TreeNode      // nil for files
   }

   type FileTreeOpts struct {
       Root     *TreeNode       // optional root label rendered above first line
       Dense    bool            // reduces row padding (design uses `dense` on Observe)
       MaxWidth int             // 0 = auto from terminal; otherwise cap
   }

   // RenderFileTree renders the tree to a string. Safe to embed in AltScreen views.
   func RenderFileTree(nodes []TreeNode, opts FileTreeOpts) string
   ```

   Rendering rules (match design exactly):
   - Glyphs: `├─ ` and `└─ ` branches; `│  ` and `   ` continuation prefixes. Box-drawing renders unchanged on any UTF-8 terminal; no fallback needed.
   - Name color: `NodeDir.NodeCurrent` → Leaf bold; `NodeDir.*` → Bark bold; `NodeFile.NodeNew` → Leaf; `NodeFile.NodeRequired` → Subtle bold; else Subtle.
   - Note: muted, rendered after the name with a 2-space gap. Truncated with `…` when line would overflow `opts.MaxWidth`.
   - Status badges right-aligned with 1-col padding: `NEW` (Leaf) and `REQUIRED` (Bark) — each 9 visible columns (badge + trailing space).
   - `NodeCurrent`: prepend a 2-col left-border `│ ` in Leaf, with a leaf-tint background via lipgloss `.Background(color-mix-equivalent)` — use `lipgloss.AdaptiveColor` so it degrades to no-background on 256-color terminals.
   - `opts.Dense`: row padding drops from 1 line to 0 (no trailing newline-between-items).
   - `opts.Root`: when non-nil, renders a prefix line with name + note, then the children below with `│  ` continuation.

2. **Tests in `internal/tui/filetree_test.go`.** Cover:
   - Flat file list renders branch glyphs correctly (first vs last item).
   - Nested dir renders `│  ` continuation prefix.
   - NodeNew + note renders `NEW` badge and Leaf-colored name.
   - NodeCurrent prepends the leaf border and background (check via ANSI presence — skip when lipgloss profile is ASCII).
   - Dense mode output has no blank lines between siblings.
   - MaxWidth trims note with `…`.
   - Root label renders above children.

3. **Add palette tokens to `internal/tui/styles.go`.** Under `Semantic Tokens`:
   ```go
   var (
       // Enso / rule chrome tokens — used by the init-flow chrome. Dimmer
       // shades of Leaf/Stone for at-rest rail segments and thin dividers.
       ColorLeafDim = lipgloss.AdaptiveColor{Dark: "#3A7253", Light: "#3D6D53"}
       ColorRule    = lipgloss.AdaptiveColor{Dark: "#3B4049", Light: "#D4D0CA"}
       ColorRule2   = lipgloss.AdaptiveColor{Dark: "#4A4F58", Light: "#B9B5AF"}
   )
   ```
   Do NOT touch existing Leaf/Bark/Moss/etc. values — other commands rely on them. The tighten happens implicitly in init-flow code via which tokens we use, not by redefining tokens globally.

#### Verification

- [ ] `make build` — compiles.
- [ ] `go test ./internal/tui/...` — all new FileTree tests pass.
- [ ] `gofmt -s -l .` — clean.
- [ ] `golangci-lint run` via CI — clean.
- [ ] Manual visual check: write a `TestRenderFileTree_Demo` in `filetree_test.go` that prints a fully-populated tree (every NodeStatus, nested dirs, Dense on/off, root label) via `t.Log(...)`. Run `go test ./internal/tui/... -v -run TestRenderFileTree_Demo` and eyeball against design reference. Keep the test — it doubles as a regression aid when styles change.
- [ ] No behavioural change to `bonsai init` / `add` / `remove` / `update` — existing e2e smoke tests for these commands still pass.

---

### Phase 2 — Custom init-flow package + chrome primitives + env-flag entrypoint

Scope: create `internal/tui/initflow/` subpackage, implement the persistent chrome (header + enso progress + footer), add a `Chromeless` optional Step capability to the harness, and wire `cmd/init.go` to route to a new stub `runInitRedesign` when `BONSAI_REDESIGN=1`. The new path renders empty 4-stage scaffolds that cycle on Enter but do nothing useful — the shell is verifiable in isolation before stage logic lands.

#### Files touched

- `internal/tui/harness/harness.go` — add `Chromeless` optional interface
- `internal/tui/harness/harness_test.go` — cover chromeless render path
- `internal/tui/initflow/chrome.go` — **new**
- `internal/tui/initflow/enso.go` — **new**
- `internal/tui/initflow/stage.go` — **new**
- `internal/tui/initflow/fallback.go` — **new** (runewidth detection + ASCII fallbacks)
- `internal/tui/initflow/fallback_test.go` — **new**
- `cmd/init.go` — env-flag branch + stub `runInitRedesign`

#### Steps

1. **Add `Chromeless` capability to the harness.** In `internal/tui/harness/harness.go`:

   ```go
   // Chromeless is an optional Step capability. When a step returns true,
   // the harness skips rendering its default header/footer and yields the full
   // frame to Step.View(). The step is responsible for drawing its own chrome
   // (banner, breadcrumb, progress, footer keys).
   type Chromeless interface {
       Chromeless() bool
   }
   ```

   In `View()`:
   ```go
   if c, ok := h.steps[h.cursor].(Chromeless); ok && c.Chromeless() {
       return h.steps[h.cursor].View()
   }
   // existing header/body/footer composition
   ```

   Add a test: a mock chromeless step returns a known-string View; assert Harness.View returns it verbatim (no header/footer prepended).

2. **Create `internal/tui/initflow/fallback.go`.** Detect unsafe wide-char terminals:

   ```go
   package initflow

   import (
       "os"
       "strings"
       "github.com/mattn/go-runewidth"
   )

   // WideCharSafe reports whether the terminal can reliably render 2-wide CJK
   // characters. Relies on runewidth's East-Asian detection; also checks for
   // known-bad terminals via $TERM and Windows consoles without UTF-8.
   func WideCharSafe() bool {
       if os.Getenv("BONSAI_ASCII_ONLY") == "1" {
           return false
       }
       // Runewidth only reports correctly on East-Asian locales; we want the
       // inverse — most non-East-Asian locales on modern terminals still
       // render CJK at 2-wide correctly. Explicit allow-list + opt-out.
       if runewidth.EastAsianWidth {
           return true
       }
       term := os.Getenv("TERM")
       // mintty / conhost pre-1809 misalign CJK. Modern Windows Terminal sets
       // WT_SESSION; conhost from Win10 1809+ also handles CJK.
       if os.Getenv("WT_SESSION") != "" {
           return true
       }
       // Known-good unix terminals
       goodTerms := []string{"xterm", "screen", "tmux", "alacritty", "kitty", "wezterm", "ghostty"}
       for _, good := range goodTerms {
           if strings.HasPrefix(term, good) {
               return true
           }
       }
       // Unknown — assume safe on any UTF-8 locale; false only when LANG/LC_ALL
       // clearly lacks UTF-8.
       lang := os.Getenv("LC_ALL")
       if lang == "" { lang = os.Getenv("LANG") }
       return strings.Contains(strings.ToLower(lang), "utf")
   }

   // Label picks between kanji+kana (safe) and ascii fallback. Matching
   // design labels: 器 Vessel うつわ; 土 Soil つち; 枝 Branches えだ; 観 Observe みる.
   type StageLabel struct { Kanji, Kana, English string }

   func (l StageLabel) Render(safe bool) (primary, secondary string) {
       if safe {
           return l.Kanji + " " + l.English, l.Kana
       }
       return l.English, ""
   }
   ```

   Tests in `fallback_test.go`: toggle `BONSAI_ASCII_ONLY=1` to force false; assert `StageLabel.Render(false)` returns English-only with empty secondary.

3. **Create `internal/tui/initflow/enso.go`.** Enso progress rail:

   ```go
   package initflow

   import (
       "strings"
       "github.com/charmbracelet/lipgloss"
       "github.com/LastStep/Bonsai/internal/tui"
   )

   const (
       ensoDone    = "●"
       ensoPending = "○"
       railChar    = "─"
   )

   // RenderEnsoRail draws the 4-stage enso progress rail, current stage styled
   // larger with a boxed kanji interior + leaf tint, done stages as filled ●,
   // pending as hollow ○. Rail segments between stages colored Leaf-to-LeafDim
   // up to current, Rule beyond.
   //
   // Layout (approx; width depends on terminal):
   //
   //    ●───────●───────╔═════╗───────○
   //                    ║ 枝 │
   //                    ╚═════╝
   //    VESSEL  SOIL   BRANCHES  OBSERVE
   //                     えだ
   //
   // On non-wide-safe terminals, falls back to a bracketed rail:
   //
   //    [x]─────[x]─────[ 3 ]─────[ ]
   //    VESSEL  SOIL   BRANCHES  OBSERVE
   //
   func RenderEnsoRail(stageIdx int, width int, safe bool) string
   ```

   Rules:
   - 4 fixed stages (labels const in fallback.go).
   - Current stage inner kanji rendered via a 3-row lipgloss box (`╭─╮`, `│K│`, `╰─╯`) with `.Background(tint)` using a color-mix approximation (adaptive color). Other stages are 1-row dots.
   - Rail lengths computed to roughly centre the rail within `width`. Pad left/right equally.
   - Labels rendered below in a second line, aligned under each stage's centre column.
   - For `stageIdx == i`: label styled Leaf bold, kana rendered small underneath (third line) — only for the current stage.

4. **Create `internal/tui/initflow/chrome.go`.** Top banner + "PLANTING INTO" breadcrumb + footer:

   ```go
   // RenderHeader renders the two top rows of every stage:
   //   row 1 (left): enso-ringed 盆 + "BONSAI" wordmark + "INITIALIZE · v<Version>"
   //   row 1 (right): "PLANTING INTO\n~/.../voyager-api/station/"
   // Station segment is colored Leaf; project segment is Bark; prefix muted.
   func RenderHeader(version, projectDir, stationSubdir string, width int, safe bool) string

   // RenderFooter renders the bottom row: "一 BONSAI 一" (muted) on the left,
   // context-specific key hints on the right.
   func RenderFooter(hints []KeyHint, width int) string

   type KeyHint struct {
       Key  string  // e.g. "↵", "␣", "?", "esc"
       Desc string  // e.g. "continue", "toggle", "details"
   }
   ```

5. **Create `internal/tui/initflow/stage.go`.** Base `Stage` step type:

   ```go
   // Stage is the shared base for Vessel/Soil/Branches/Observe. Each stage
   // composes its body into chromefull() -> string. Satisfies harness.Step and
   // harness.Chromeless so the harness yields the whole frame.
   type Stage struct {
       title        string
       idx          int              // 0..3 — which stage in the rail
       label        StageLabel
       projectDir   string           // filled from runInit context
       stationDir   string           // starts as "station/", updated from Vessel
       version      string
       agentDisplay string           // from agentDef.DisplayName — rendered in Observe's AGENT row
       startedAt    time.Time        // captured at runInitRedesign entry — Planted uses time.Since for its ELAPSED clock
       width, height int
       done         bool
       ensoSafe     bool
   }

   func (s *Stage) Chromeless() bool { return true }
   func (s *Stage) Title() string    { return s.title }
   func (s *Stage) Done() bool       { return s.done }
   func (s *Stage) Result() any      { return nil } // overridden in subclasses

   // renderFrame composes header + enso + body + footer.
   func (s *Stage) renderFrame(body string, keys []KeyHint) string
   ```

   In Phase 2 the body is a placeholder: each stage's `View()` returns `renderFrame("  (stage body goes here)", default-keys)`. Enter advances.

6. **Wire env-flag branch in `cmd/init.go`.**

   At the top of `runInit`:
   ```go
   if os.Getenv("BONSAI_REDESIGN") == "1" {
       return runInitRedesign(cmd, args)
   }
   ```

   Create `runInitRedesign` in a new file `cmd/init_redesign.go` that:
   - Captures `startedAt := time.Now()` at entry — threaded through every Stage via ctor so PlantedStage can compute `time.Since(startedAt)`.
   - Computes `cwd`, `configPath`, loads catalog + agent (same as legacy).
   - Pulls `agentDisplay := agentDef.DisplayName` (fallback to `catalog.DisplayNameFrom(agentDef.Name)` if empty).
   - Early-exits if config exists (same warning path).
   - Builds the 4 stage Steps with placeholder results, passing `startedAt` + `agentDisplay` + `version` + `projectDir` into each stage ctor.
   - Uses `harness.Run` (unchanged) to drive them.
   - Stubs the generate + planted + conflict flow to no-op — in Phase 2 the redesign path only paints chrome and cycles stages. Actual file writes happen in Phase 5.

#### Verification

- [ ] `make build` — compiles both flags.
- [ ] `go test ./...` — passes.
- [ ] `BONSAI_REDESIGN=1 ./bonsai init` in a tmpdir renders the 4-stage chrome + rail; Enter cycles through; Esc pops back one stage. Ctrl-C aborts with no config written.
- [ ] Legacy: `./bonsai init` in a tmpdir runs the old flow unchanged.
- [ ] Terminal resize mid-flow re-wraps chrome without crashes.
- [ ] `BONSAI_ASCII_ONLY=1 BONSAI_REDESIGN=1 ./bonsai init` uses the ASCII fallback rail.
- [ ] Snapshot the chrome for a 120x32 terminal and paste it into the PR description for review.

---

### Phase 3 — Vessel + Soil stages (real input)

Scope: Vessel (3 textinputs on one page) and Soil (custom multi-select list). Branches/Observe remain chrome-only stubs. Continues behind `BONSAI_REDESIGN=1`. Also strips the misleading `station/` segment from the header (station subdir doesn't exist at render time — only the project root does).

#### Files touched

- `internal/tui/initflow/chrome.go` — drop `stationSubdir` parameter from `RenderHeader`; header renders project path only
- `internal/tui/initflow/chrome_test.go` — **new** — covers the project-only header render + `~`-collapse behavior
- `internal/tui/initflow/stage.go` — drop `stationDir` arg from `renderFrame`'s `RenderHeader` call (keep the `stationDir` field on `Stage` and `StageContext` — Vessel writes it; Phase 5 Planted body will render it)
- `internal/tui/initflow/vessel.go` — **new**
- `internal/tui/initflow/vessel_test.go` — **new**
- `internal/tui/initflow/soil.go` — **new**
- `internal/tui/initflow/soil_test.go` — **new**
- `cmd/init_redesign.go` — wire real steps

#### Steps

1. **Create `VesselStage`** — 3× `textinput.Model` (name, description, station dir) on one page.

   Layout matches design (`zen-shell.jsx` ZStepProject):
   ```
     器 うつわ · VESSEL
     Shape the vessel.
     Every Bonsai begins with a small decision...

     ─── FIELDS ─── 入力 ──────────────

     NAME          ❯ voyager-api             [caret]
     required      ↳ inherited from git remote

     DESCRIPTION   ❯ Internal voyager service...
     optional      one line · shown in agent prompts

     STATION       ❯ station/
     where files   default · subdirectory under project root
     live
   ```

   Keybindings:
   - `↑ ↓ / tab shift-tab`: cycle focus between the three inputs (matches bubbles.textinput's blur-on-tab pattern).
   - `↵`: on the last focused field, submit + advance stage. On other fields, move focus down.
   - `esc` / `shift+tab` on first field: propagated to harness (pops back — no-op at stage 0).

   Validation:
   - `Name`: required — trim-empty errors inline ("required") in Leaf-Dim text under the field. Re-enter edit mode on ↵ until valid.
   - `Description`: optional.
   - `Station`: required — reject empty / `"/"` per existing `stationDirValidator`. Trim + append `/` per `normaliseDocsPath`.

   `Result()` returns `map[string]string{"name": ..., "description": ..., "station": ...}` — one bag for the stage, to keep `prev[]` indexing stable across stages vs today's per-field indexing.

2. **Create `SoilStage`** — custom multi-select list based on `bubbles/list` (or hand-rolled — judge at implementation; hand-rolled gives us exact row layout control from the design).

   Layout (`zen-shell.jsx` ZStepScaffolding):
   ```
     土 つち · SOIL
     Choose what the project carries.
     Shared files every agent can see. ...

     ─── SCAFFOLDING ─── 足場 ─────────

     ◆ CLAUDE.md          Root-level agent directive...     REQUIRED
     ◆ agents-index       Directory of every agent...       REQUIRED
     ◆ session-log        Rolling log of what each...
     ◆ readme-stub        A starter README...
     ◇ editor-config      Editorconfig file with your...
     ◇ git-hooks          Pre-commit + pre-push hooks...

     4 of 6 selected · 2 required, always on
   ```

   Interaction:
   - `↑ ↓`: move focus. Focused row has Leaf border-left + 7% leaf-tint background.
   - `␣`: toggle selection. Required items ignore toggle.
   - `↵`: advance stage.
   - Badge: `REQUIRED` in Bark right-aligned.
   - Glyph: `◆` selected (Leaf), `◇` unselected (muted2).

   Result: `[]string` — selected item `Name` values (matches current `asStringSlice(prev[3])` contract from legacy `runInit`).

3. **Wire real results into `runInitRedesign`.**

   Stage slice in `runInitRedesign`:
   - `NewVesselStage(ctx StageContext)` — stage 0
   - `NewSoilStage(ctx StageContext, options []ScaffoldingOption)` — stage 1
   - `newBranchesStageStub(ctx StageContext)` — stage 2 (empty scaffold, returns empty picks)
   - `newObserveStageStub(ctx StageContext)` — stage 3 (returns `false` confirm)
   - Legacy generate/conflict tail is still skipped in Phase 3.

4. **Strip `station/` from `RenderHeader`.**

   User-reported bug: header's right block reads `PLANTING INTO / ~/.../project/station/`, but `station/` doesn't exist yet at any point before Phase 5 generate. Showing it is misleading — it claims the path exists when it doesn't.

   - Change `func RenderHeader(version, projectDir, stationSubdir string, width int, safe bool) string` → `func RenderHeader(version, projectDir string, width int, safe bool) string`.
   - Inside: compute `projectDisplay := collapseHome(projectDir)`; render `parent` muted + `projectName` in Bark + trailing `/` muted. Drop the `station` segment and the `stationSubdir` normalization.
   - `Stage.renderFrame` call site: `RenderHeader(s.version, s.projectDir, width, s.ensoSafe)`.
   - Keep `Stage.stationDir`, `StageContext.StationDir`, and Vessel's station input — still needed: Phase 5 Planted body renders the generated file tree rooted at `station/`.

5. **Tests.**

   - `chrome_test.go`: header with `projectDir="/home/alice/voyager-api"` collapses to `~/voyager-api/`; `projectDir="/tmp/p"` with `HOME=/home/bob` keeps absolute; no `station` substring appears in rendered output.
   - `vessel_test.go`: TextInput focus cycling; required-empty validation blocks submit; Description empty → Result contains `""`; Station default (`station/`) applied when input empty; `Result()` returns `map[string]string` with keys `name`/`description`/`station`.
   - `soil_test.go`: Hand-rolled list (not `bubbles/list` — overkill for a 4–8 item list; hand-roll gives exact row layout from `zen-shell.jsx`). Required items pre-selected and cannot be toggled off; arrow-key focus advances/wraps; Space toggles optional items; Result order matches input order; empty selection permitted only if all optional items are unselected and required items covered.

#### Design decisions resolved (pre-dispatch refinement, 2026-04-21)

- **Vessel `Result()` shape = `map[string]string`** (keys `name`, `description`, `station`). Single stage replaces the three legacy steps, so `prev[]` becomes per-stage not per-field: `runInitRedesign` reads `vessel := prev[0].(map[string]string); name := vessel["name"]`. Matches the "one bag per stage" note at §Vessel·Result.
- **Soil implementation = hand-rolled list, not `bubbles/list`.** Scaffolding catalog is ~4–8 items; `bubbles/list` ships fuzzy filter + pagination + its own row styling (hard to match `zen-shell.jsx` exact layout). Hand-roll is ~120 lines, single file, full control over focus highlight, required-badge placement, and `◆/◇` glyphs.
- **Header station strip**: removed for clarity — confirmed not a breaking change since `RenderHeader` has one caller (`Stage.renderFrame`) inside the same package.

#### Verification

- [ ] `BONSAI_REDESIGN=1 ./bonsai init` — header right block shows `~/.../project/` only, no `station/` segment.
- [ ] `BONSAI_REDESIGN=1 ./bonsai init` — Vessel page accepts 3 inputs, tab cycles focus, required validation works, station field receives `station/` default.
- [ ] Soil page lists scaffolding items from catalog, required pinned, `␣` toggles optional, counter updates ("X of N selected").
- [ ] Esc at Soil pops back to Vessel with prior values preserved (TextInputs still show entered strings).
- [ ] `go test ./internal/tui/initflow/...` passes.
- [ ] `make build && go test ./...` — full suite clean.

---

### Phase 4 — Branches tabbed picker + inline-expand

Scope: replace the stub with the real Branches stage — tabbed category picker, per-category item list, inline-expand on focused row.

#### Files touched

- `internal/tui/initflow/branches.go` — **new**
- `internal/tui/initflow/branches_test.go` — **new**
- `cmd/init_redesign.go` — wire `BranchesStage`
- `internal/catalog/catalog.go` — expose per-item metadata helpers if needed (`.Description`, `.Affects`, cross-links) — verify shape before touching; most fields already present on `CatalogItem`

#### Steps

1. **Read the full catalog item shape** in `internal/catalog/catalog.go` before writing BranchesStage. Verify these fields exist on the ability item structs (CatalogItem, SensorItem, RoutineItem): `Name`, `DisplayName`, `Description`. Check for `Affects` / cross-link fields. If absent, leave them out of the inline-expand panel (show only Description + file path) — the design's "affects / cross-links" metadata is a stretch, and best deferred to a separate catalog-metadata plan if the fields don't exist yet.

2. **Create `BranchesStage`.** State:

   ```go
   type BranchesStage struct {
       Stage                    // embedded chrome
       categories []branchCat   // 5 tabs: Skills/Workflows/Protocols/Sensors/Routines
       catIdx     int           // current tab
       expanded   bool          // inline-expand on/off (? toggles)
       itemIdx    map[int]int   // cat.index -> focused item row per tab
       selected   map[int]map[string]bool // cat.index -> set of machine-names
   }

   type branchCat struct {
       key         string         // "skills" etc.
       displayName string
       kanji       string         // 技 流 律 感 習
       items       []branchItem
   }

   type branchItem struct {
       name, displayName, description string
       required, isDefault           bool
       affects, crossLinks, filePath string // empty-string-safe
   }
   ```

   Layout (matches `zen-shell.jsx` ZStepAbilities):
   ```
     枝 えだ · BRANCHES
     Shape the branches of the Tech Lead.

     ┌──────────────┬──────────────┬──────────────┬──────────────┬──────────────┐
     │ 技 skills ◆  │ 流 workflows │ 律 protocols │ 感 sensors   │ 習 routines  │
     │ 4 / 17       │ 4 / 10       │ 4 / 4        │ 3 / 11       │ 3 / 8        │
     ├──────────────┴──────────────┴──────────────┴──────────────┴──────────────┤

     ◆ api-design       REST + OpenAPI conventions...              DEFAULT  ·
     ◆ auth-patterns    Session / JWT / OAuth2 flows...            DEFAULT  ·
     ◆ coding-standards Style guide, naming, error handling...     DEFAULT  ·
     ◆ testing-strategy Test pyramid, coverage thresholds...       DEFAULT  ·
     ◇ database-patterns Schema design, migration ordering...                ▾
       ABOUT     Conventions for schema design, migration...
       AFFECTS   planning · code-review · schema-migration
       CROSS     testing-strategy · observability
       FILE      station/skills/database-patterns.md
     ◇ observability    Structured logging, metrics naming...
     ...

     18 abilities selected · across 5 categories          [?] toggle details
   ```

   Keybindings:
   - `← →` / `h l`: switch tab (cycles).
   - `↑ ↓` / `j k`: move focus within current tab (clamp, no wrap).
   - `␣`: toggle item (ignored for required items — they show as pinned with `◆` + `(required)` note, no toggle).
   - `?`: toggle `expanded` (global — when true, the focused row shows the expand block; other rows stay compact).
   - `↵`: advance stage with `Result() any` returning `BranchesResult{Skills, Workflows, Protocols, Sensors, Routines []string}`.
   - `esc` / `shift+tab`: propagate to harness — pops back to Soil.

   Rendering details:
   - Tab row: `[selected-count] / [total]` subtitle under each kanji+label; current tab has Leaf border-bottom + 6% leaf-tint bg.
   - Required items: always selected (`◆` Leaf), toggle is a no-op; label includes muted `(required)` inline.
   - Default items: show `DEFAULT` in muted2 right-aligned above the `·` / `▾` caret.
   - Focused row: leaf border-left + 7% leaf-tint bg.
   - Inline expand: only renders when `expanded && focused`. 4-row key/value block with labels ABOUT / AFFECTS / CROSS / FILE; empty values skipped.
   - Selection counter summed across all 5 tabs.

3. **Seed defaults from `agentDef.Default{Skills,Workflows,Protocols,Sensors,Routines}`.** Mark required items as pre-selected and immutable.

4. **Tests.**

   - `branches_test.go`: tab cycling; items within a tab scroll; ␣ toggles non-required; ␣ on required no-op; `?` toggles expand; Result collects per-tab selections; defaults applied on first render; required always in Result.

#### Verification

- [ ] `BONSAI_REDESIGN=1 ./bonsai init` — all 5 ability tabs render, selection state persists when switching tabs, counts update live.
- [ ] `?` toggles the inline-expand block on the focused row; arrow keys move the focus; expand stays on across row moves (global toggle).
- [ ] Escape pops back to Soil with all current Branches picks preserved (re-entering Branches restores the same selections). Harness contract: `Reset()` on BranchesStage is a **no-op** — the step's `selected`/`catIdx`/`itemIdx`/`expanded` state must NOT be cleared when the user Esc's back and then returns. Confirm this by reading `harness.harness.go` cursor-rewind behavior before implementation and mirroring what VesselStage/SoilStage do (they similarly preserve textinput values and list cursors).
- [ ] `go test ./internal/tui/initflow/...` passes.

---

### Phase 5A — Responsive resize pass + Observe/Generate/Planted stages

**Scope:** (1) add responsive width/height handling across every initflow stage so `bonsai init` no longer clips on narrow/short terminals; (2) build the three final stages (Observe, Generate, Planted). Still gated by `BONSAI_REDESIGN=1` — default flow unchanged.

**Why split:** Phase 5 originally bundled stages + wiring + legacy delete into one PR. Adding the resize overhaul pushes the LoC count past the ~1000-line comfort threshold and mixes three independent concerns. 5A lands stages + resize (net additive, no default-flow change); 5B is a small, surgical PR that flips the default and deletes legacy code.

**User-reported issue (2026-04-21):** On narrow terminals (<100 cols) the Branches stage clips: `nameColW=24` + `descColW=44` + tag column (10) + borders = 84-cell fixed row. Below ~90 cols the tag column is cut off; below ~70 cols the description column is partially eaten. DETAILS box (3 rows × 70 cells) clips identically. Vessel and Soil have the same disease at smaller scale. Height: lists render flat, no viewport — long tab contents (≥8 items) push counter + details off the bottom of 24-row terminals.

#### Files touched

- `internal/tui/initflow/layout.go` — **new** — shared responsive helpers (`clampColumn`, `viewport` hand-rolled scroll, min-size floor panel)
- `internal/tui/initflow/layout_test.go` — **new**
- `internal/tui/initflow/branches.go` — responsive widths + vertical scroll for list + details block clamp
- `internal/tui/initflow/soil.go` — responsive widths (row padding derived from `s.width`, description truncate follows)
- `internal/tui/initflow/vessel.go` — input width derived from `s.width` (currently fixed 60)
- `internal/tui/initflow/chrome.go` — add `RenderMinSizeFloor(width, height)` helper that renders a centred "please enlarge your terminal" panel when `width < 70 || height < 20`
- `internal/tui/initflow/stage.go` — route `renderFrame` through the floor when dims are too small
- `internal/tui/initflow/observe.go` — **new**
- `internal/tui/initflow/observe_test.go` — **new**
- `internal/tui/initflow/generate.go` — **new**
- `internal/tui/initflow/generate_test.go` — **new**
- `internal/tui/initflow/planted.go` — **new**
- `internal/tui/initflow/planted_test.go` — **new**
- `cmd/init_redesign.go` — swap `NewStubStage(3, …)` for `NewObserveStage(…)` only; Generate + Planted remain uninstalled in 5A (the conditional-splice + `BONSAI_REDESIGN` removal lives in 5B)
- Existing `branches_test.go` / `soil_test.go` / `vessel_test.go` — add responsive assertions (narrow-width render does not truncate tag/required columns; long lists scroll)

#### Steps

##### 1. Shared responsive helpers (`internal/tui/initflow/layout.go`)

Centralise the resize primitives so every stage uses the same rules and tests stay proportional to the surface.

```go
// Min dimensions below which every stage shows a "please enlarge terminal"
// floor instead of attempting to lay out its body.
const (
    MinTerminalWidth  = 70
    MinTerminalHeight = 20
)

// TerminalTooSmall reports whether the given dims fall below the floor.
func TerminalTooSmall(width, height int) bool { ... }

// ClampColumns returns the per-column cell budget for a given row width.
// For Branches: (nameW, descW, tagW) scaled proportionally, floors applied,
// cumulative width <= availableWidth.
func ClampColumns(availableWidth int) (nameW, descW, tagW int) { ... }

// Viewport is a minimal hand-rolled vertical scroll: holds a slice of lines
// (pre-rendered) and an offset. Focus-follows-cursor: caller supplies the
// focused-line index, viewport clamps offset so that line is visible.
// Rationale: matches Soil precedent (hand-roll list vs. bubbles/list).
type Viewport struct { lines []string; offset, height int }
func (v *Viewport) SetLines(lines []string)
func (v *Viewport) SetHeight(h int)
func (v *Viewport) Follow(focusIdx int)  // adjusts offset
func (v *Viewport) View() string         // returns v.height joined lines
```

The viewport is tiny (~60 LoC) and sidesteps the `bubbles/viewport` dep — matches the existing "hand-roll small things" precedent from Soil.

##### 2. Min-size floor panel (`chrome.go`)

```go
// RenderMinSizeFloor renders a centred panel telling the user to enlarge the
// terminal. No stage body is attempted below the floor so we never paint a
// broken layout. Uses RenderHeader at a best-effort width (clamped to >=40)
// for a recognisable brand even on tiny terminals.
func RenderMinSizeFloor(width, height int) string { ... }
```

`stage.go:renderFrame` checks `TerminalTooSmall(s.width, s.height)` at the top and returns the floor render early if true. Stage-specific bodies never run with degenerate dims.

##### 3. Branches responsive refactor

- Replace `const nameColW = 24`, `const descColW = 44`, `const colW = 16` with values derived from `ClampColumns(availableW)` where `availableW = s.width - 4` (side padding).
- Tag column (`DEFAULT` / `(required)`) pinned at `tagW = 12` to prevent squeeze; description absorbs the remaining space after name (`nameW = min(24, availableW / 3)`) and tag. Floor at `descW = 20`; below that, drop description entirely (render name + tag only).
- Wrap `renderList` with a `Viewport` so when catalog entries exceed `listH := s.height - chromeRows`, the focused row stays visible and up/down scrolls. Focus behavior unchanged.
- DETAILS block `contentW` clamped to `min(s.width - 10, 70)`.
- Per-tab 2-line intro clamp to `s.width - 4`.

##### 4. Vessel + Soil responsive refactor

- Vessel: `inputW = clamp(s.width - labelColW - 4, 30, 60)`. `labelColW` stays 20; inputs shrink proportionally. Underline tracks `inputCellW = inputW + 4`.
- Soil: `namePad = min(20, s.width / 4)`; description truncate cap = `max(30, s.width - namePad - 20)`.
- Neither stage needs vertical scroll today (scaffolding is 4–8 items). Add a TODO note in each file referencing the Viewport helper if ever needed.

##### 5. Observe stage (`observe.go`)

Layout (from `zen-shell.jsx` ZStepReview — left = Vessel summary + Soil tree, right = Branches summary by kanji-group):

```
  観 OBSERVE
  One last look before planting.

  ─── VESSEL ─── 器 ──             ─── BRANCHES ─── 枝 · 18 abilities ──
  NAME         voyager-api        技 SKILLS    api-design · auth-patterns · …
  DESCRIPTION  Internal API …     流 FLOWS     planning · code-review · pr-review
  STATION      station/           律 RULES     memory · security · scope
  AGENT        Tech Lead          感 SENSE     scope-guard · dispatch-validator
                                  習 HABIT     backlog-hygiene · doc-freshness

  ─── SOIL ─── 土 · scaffolding ──
  station/
  ├─ agents-index.md     REQUIRED   directory of every agent
  ├─ session-log.md      NEW        rolling per-session log
  └─ readme.md           NEW        starter README if one doesn't exist

  ┌─ Plant N files into <name>?   0 CONFLICTS  ─────────────────┐
  │ Existing files will be offered for merge · nothing          │
  │ overwritten without your say-so                             │
  │                           [ CANCEL ]  [ ⏎  PLANT ]          │
  └─────────────────────────────────────────────────────────────┘
```

- Soil tree: preview-only. Renders scaffolding picks (`prev[1]` — SoilStage Result) via `tui.RenderFileTree`. No ability subdirs (those land at generate time).
- Branches summary: groups `prev[2]` (BranchesResult) into five kanji-labelled rows.
- Vessel summary: reads `prev[0]` map.
- **Responsive:** Two-column grid (left VESSEL+SOIL · right BRANCHES) when `s.width >= 100`; single-column stacked layout below that. Summary rows word-wrap via the existing `wrapToWidth` helper.
- **CANCEL / PLANT CTA:** `y` / `Y` / `↵` with PLANT focused → confirm (Result = `true`). `n` / `N` → cancel (Result = `false`). `tab` / `← →` → toggle button focus.
- Result: `bool`.
- File count for CTA banner: read `wr` via a sentinel — no generation has run yet; use `len(prev[1]) + len(prev[2])` as an upper-bound approximation + a `…` suffix. Exact count is shown on Planted. (Scope-light: don't pre-simulate the generate pipeline.)

##### 6. Generate stage (`generate.go`)

Full-screen custom view; runs the generate action in a goroutine and holds for `max(actualDuration, 600ms)`.

State machine (carried verbatim from the original Phase 5 plan):

```
stateRunning   →  action goroutine active, arc draws from 0° to 360°
stateMinHold   →  action done but elapsed < 600ms, continue drawing
stateDone      →  action done and 600ms elapsed; Done() = true → harness advances
stateError     →  action returned error; show InfoPanel, wait for key
```

Arc: 12-row × 24-col box of `●` / `○` / `◐` glyphs, lit progressively by tick. Centred kanji 生. Progress label animates 種 SEED → 苗 SPROUT → 盆栽 BONSAI on a hairline below the arc. `tea.Tick` at 24fps (42ms interval).

Pipeline inputs: the generate stage wraps the same generator calls that today live inside `runInit`'s spinner closure — `cfg.Save(configPath)` first (legacy safety invariant), then `generate.Scaffolding / AgentWorkspace / PathScopedRules / WorkflowSkills / SettingsJSON`. The stage is constructed with: `cat`, `agentDef`, `cwd`, `configPath`, `lock`, plus pointers to `wr` / `cfg` / `installed` populated from Vessel/Soil/Branches results during its Init.

**Responsive:** arc scales 12x24 → 8x16 below `width < 90`; below floor → not reached (min-size panel catches it).

##### 7. Planted stage (`planted.go`)

Full-screen view post-generate:

```
[ 盆 ]  BONSAI                             ELAPSED  00:04.8
        PLANTED · <project>

                   生 · PLANTED
                   <project> is ready.
           N files written · K conflicts · lock synced

  ────────────────────────────────────────────────────────────────────

  WRITTEN · 書                  SUMMARY · 概要
  ~/<path>/                     AGENT      Tech Lead → station/
  ├─ CLAUDE.md   NEW            ABILITIES  N wired
  ├─ .bonsai.yaml  NEW          ···        X skills · Y flows · Z rules · …
  └─ station/
     ├─ agents-index.md  NEW    ────────────
     ├─ protocols/             NEXT · 次へ
     │  ├─ memory.md  NEW
     │  ...                     一  $ claude            open the workspace
     ...                            Say "hi, get started" — the Tech Lead self-orients.

                                二  $ bonsai add        add a code agent
                                    Backend, frontend, devops — each with its own workspace.

                                三  $ bonsai dashboard  tend the garden
                                    Inspect and adjust abilities after the fact.

  一 BONSAI 一 planted with care                               [⏎] exit
```

File tree sourced from `WriteResult.Files`:
- Group by directory (split `RelPath` on `/`).
- NodeStatus mapping: `ActionCreated` → NodeNew; `ActionUpdated` / `ActionForced` → NodeNew with badge "UPDATED" (Bark); `ActionUnchanged` → NodeNormal; `ActionSkipped` / `ActionConflict` omitted from tree.
- Agent workspace subtree (under `station/`) marked `NodeCurrent` (leaf border + tint).
- **Scope-light on REQUIRED badge:** first ship uses `NEW` on all created files; skip REQUIRED for scaffolding-required items. Follow-up plan item (add to Backlog Group F after 5A ships).

**Responsive:** two-column grid (WRITTEN left · SUMMARY right) when `s.width >= 100`; single-column stacked below. File tree wrapped in a `Viewport` when its line count exceeds available body height.

Keys: `↵` or `q` → exits.

Done() returns `true` only after the user acknowledges — the flow ends here, so the harness exiting is the exit.

##### 8. Wire Observe only (keep stubs for Generate/Planted in 5A)

`cmd/init_redesign.go` line 79 currently installs `NewStubStage(3, ...)` — replace with `NewObserveStage(ctx, cat, agentDef)`. Generate + Planted are NOT wired into `runInitRedesign` in 5A — they exist as packaged types with full tests but are not in the step chain. 5B does the full splice.

Rationale: keeps 5A self-contained (can ship + merge independently), lets 5B be a small wiring/deletion PR that's trivial to review.

#### Verification (5A)

- [ ] `BONSAI_REDESIGN=1 ./bonsai init` with `tput cols; tput lines` at 120×40 → full flow through Observe; terminal returns clean on Ctrl-C.
- [ ] Same at 80×24 → all stages render without clipping; Branches tag column always visible; descriptions truncate with `…` instead of running past the edge.
- [ ] Same at 70×20 (floor) → last column may be tight but everything fits.
- [ ] Same at 60×16 → every stage shows the "please enlarge terminal" floor panel; `ctrl-c` still exits cleanly.
- [ ] Branches list with every routine enabled (≥15 items in Routines tab) scrolls vertically when focus moves past visible height; counter + DETAILS stay on screen.
- [ ] Observe: `y` / `↵` with PLANT focused → Result = true; `n` → Result = false; tab toggles button focus.
- [ ] Observe: Vessel facts reflect prior inputs; Soil tree shows scaffolding picks; Branches summary correctly kanji-grouped.
- [ ] `go test ./internal/tui/initflow/...` passes (Observe/Generate/Planted unit tests + Viewport tests + layout clamp tests).
- [ ] `make build && go test ./...` passes.
- [ ] No behavioural change to default `bonsai init` (env flag not set) — legacy flow unchanged.

---

### Phase 5B — Wire Generate + Planted, flip default, delete legacy

**Scope:** small surgical PR — splice Generate + Planted + conflict-picker into `runInitRedesign`, flip `BONSAI_REDESIGN=1` default, delete legacy `runInit` body and helpers, handoff artifacts.

#### Files touched

- `cmd/init_redesign.go` — full step splice (Generate + conflict picker + Planted); remove env-flag routing
- `cmd/init.go` — rename `runInitRedesign` → `runInit`; delete legacy `runInit` body + `buildReviewPanel`; drop `BONSAI_REDESIGN` env-flag branch
- `cmd/init_test.go` (if present) — update or delete tests referencing legacy helpers
- `station/Playbook/Plans/Active/22-init-redesign.md` — mark Complete + archive
- `station/Playbook/Status.md` — move Plan 22 to Recently Done
- `station/agent/Core/memory.md` — update Work State; add durable learnings from the rewrite
- `station/Reports/Pending/` — final rewrite report

#### Steps

1. **Splice Generate + conflict + Planted into `runInitRedesign`.**

   ```go
   steps := []harness.Step{
       initflow.NewVesselStage(ctx),
       initflow.NewSoilStage(ctx, soilOptions),
       initflow.NewBranchesStage(ctx, cat, agentDef),
       initflow.NewObserveStage(ctx, cat, agentDef),          // prev[3] Result() = bool
       harness.NewConditional(
           initflow.NewGenerateStage(ctx, cat, agentDef, cwd, configPath, lock, &wr, &cfg, &installed),
           plantedConfirmed,                                  // reads prev[3]
       ),
       harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
           if !wr.HasConflicts() {
               return nil
           }
           return buildConflictSteps(&wr)
       }),
       harness.NewConditional(
           initflow.NewPlantedStage(ctx, &wr, cfg, installed),
           plantedConfirmed,
       ),
   }
   ```

   `plantedConfirmed := func(prev []any) bool { b, _ := prev[3].(bool); return b }`.

   Conflict-picker LazyGroup splices zero or more `MultiSelect` steps in; Planted follows unconditionally if Observe confirm was `true`.

2. **Flip default + delete legacy.**

   - `cmd/init.go`: drop the `BONSAI_REDESIGN` env-flag branch; rename `runInitRedesign` → `runInit`; delete the old `runInit` body and helpers exclusive to it (`buildReviewPanel`, `scaffoldingOptions`).
   - **Keep if still referenced elsewhere:** `stationDirValidator`, `normaliseDocsPath`, `asString` / `asStringSlice` / `asBool`, `userSensorOptions`, `toItemOptions`, `toRoutineOptions`, `toSensorOptions`, `buildConflictSteps`, `applyConflictPicks`. Grep each before deletion (`grep -rn "<name>" --include='*.go' .`).
   - **Test sweep:** `grep -rn 'buildReviewPanel\|scaffoldingOptions' --include='*_test.go' .` — update or delete any tests referencing deleted helpers. Any test that imports `cmd` and exercises init behaviour must be re-scored against the redesigned flow or deleted.
   - Run `go test ./...` after deletion — catches anything still referencing the old path.

3. **Final audit.**

   - `grep -rn "BONSAI_REDESIGN" .` → should return nothing outside `station/` (plan file + memory).
   - `grep -rn "runInitRedesign" .` → should return nothing (rename complete).
   - Manual smoke at 120×40 and 80×24 — full flow works without env flag.

4. **Handoff + docs.**

   - Update `station/Playbook/Status.md`: move Plan 22 row from In Progress to Recently Done with today's date + final PR list.
   - Update `station/agent/Core/memory.md` Work State: plan complete, main sha.
   - Flush any durable UX-preference learnings into `memory.md` Feedback under the 2026-04-17 section.
   - Submit a final report to `station/Reports/Pending/` summarising the rewrite (phases shipped, PR numbers, notable decisions).

#### Verification (5B)

- [ ] Default `bonsai init` (no env flag) runs the redesigned flow.
- [ ] `BONSAI_REDESIGN` env var is no longer consulted anywhere in `cmd/` or `internal/`.
- [ ] `grep -rn "runInitRedesign" --include='*.go' .` returns nothing.
- [ ] Generate always holds ≥600ms; with a real catalog (~23 files) total elapsed ≈ 0.6–1.0s.
- [ ] Generate surfaces errors in an InfoPanel and does not advance to Planted on failure.
- [ ] Conflict picker still fires when existing files are detected (manual test: `touch station/CLAUDE.md` before re-running init).
- [ ] Planted shows the file tree rooted at project dir, NEW badges on created files, summary + 3 next-command rows.
- [ ] `↵` / `q` exits Planted cleanly; terminal returns to normal stdout without residue.
- [ ] `make build && go test ./...` passes.
- [ ] `golangci-lint` CI clean.

---

## Dependencies

- `github.com/charmbracelet/bubbles` — already indirect in `go.mod` via huh; Phase 2 promotes to direct.
- `github.com/mattn/go-runewidth` — already indirect; Phase 2 promotes to direct.
- `charmbracelet/bubbletea` — already direct.
- `charmbracelet/lipgloss` — already direct.
- No new external dependencies. No catalog schema changes (Phase 4 verifies `Description`/`Affects` availability before depending on them).

Run `go mod tidy` at the end of Phase 2 to promote the two indirects.

---

## Security

> [!warning]
> Refer to [Playbook/Standards/SecurityStandards.md](../Standards/SecurityStandards.md) for all security requirements.

Scope of changes is pure TUI + orchestration. No network. No new file-writing code (reuses existing `generate.*` pipeline verbatim). No new shell / exec calls. No user input fed to eval/templates. `ProjectName` and `Description` pass through the same trim + save path as today (`cfg.Save(configPath)`) — no shell injection surface added.

Specific checks each PR:
- [ ] No hardcoded credentials / tokens in new files.
- [ ] No `exec.Command` invocations.
- [ ] No secrets exfiltrated via the Planted screen's file tree — it only reads `WriteResult.Files` (relative paths written by our own generators).
- [ ] GitGuardian / gitleaks CI remains green.

---

## Out of Scope

- `bonsai add` / `bonsai remove` / `bonsai update` — untouched. Palette changes isolated to init-flow code paths.
- `bonsai dashboard` command — separate future plan per user decision.
- Catalog metadata additions (`Affects`, cross-links) — if they don't exist as fields yet, render what's available and file a follow-up plan rather than bolting them on here.
- Snapshot/golden-file tests for exact rendered frames — expensive to maintain; rely on unit tests of individual components + manual visual verification per PR.
- Windows Terminal / mintty wide-char testing beyond runewidth detection — verified by `BONSAI_ASCII_ONLY=1` fallback path only; pixel-perfect Windows rendering deferred.
- Animations beyond the Generate enso draw-in — transitions between stages are instant (subtle per design's "animation_tolerance" answer).
- The `mustCwd` path-error surfacing UX from Plan 14 — unchanged.
- Sixel / Kitty graphics protocol fallbacks for a true-circle enso — explicitly rejected in Q5 answer; unicode rail only.

---

## Risk Register

| Risk | Mitigation |
|------|------------|
| CJK characters misalign on Windows conhost <1809 | `BONSAI_ASCII_ONLY=1` opt-out + auto-fallback via `WideCharSafe()`. |
| Terminal truecolor absent → color mix via `AdaptiveColor` degrades ugly | Test with `TERM=xterm-256color` locally; ensure at-rest rail is readable on 256-color. |
| Bubble Tea tick interval too slow for 600ms animation | 24fps (42ms) is well inside BubbleTea's typical cadence; verified by existing `SpinnerStep`. |
| Generate runs fast → min-hold feels artificial | Design agreed min-hold is the preferred compromise; if user pushback, flip to straight-to-planted with zero hold behind a subseq. env flag. |
| Legacy flow drift during Phase 2–4, 5A | Env-flag branch isolates new code; legacy gets no changes until Phase 5B. |
| Big scope / plan creep | Each phase is an independent PR with its own Verification gate. If a phase reveals missing catalog data, spin out a dependency plan rather than growing the phase. |
| FileTree diff in Planted doesn't match Observe Soil preview because `WriteResult` omits unchanged scaffolding | Confirm `WriteResult` includes `ActionCreated` for scaffolding files on fresh init; if not, augment `WriteResult.Files` during generate to always include attempted writes. Check in Phase 5A Planted step before shipping. |
| Responsive refactor breaks previously-nominal 120-col layout | Clamp thresholds calibrated around the current nominal width (120 col); `ClampColumns(120)` must produce the existing widths (name 24, desc 44, tag 12). Regression test asserts this. |
| Viewport scroll breaks focus-follows behaviour | Unit test covers `Viewport.Follow(focusIdx)`: verifies offset clamps so the focused line is always within `[offset, offset+height)`. |
| Min-size floor triggers on legitimate 80×24 terminals | Floor thresholds set to 70×20 — below minimum POSIX vt100 compatibility; 80×24 is comfortably above the floor. Verified in 5A Verification smoke at 80×24. |

---

## Verification (master gate)

- [ ] All phases merged and in main (Phases 1, 2, 3, 4, 5A, 5B — 6 PRs total).
- [ ] `bonsai init` in a fresh tmpdir renders the full cinematic flow: Vessel → Soil → Branches → Observe → Generate (≥600ms) → Planted. No env flag required.
- [ ] `bonsai init` with an existing `.bonsai.yaml` prints the "already exists" warning unchanged.
- [ ] Conflict picker fires correctly on pre-existing scaffolding files.
- [ ] `go test ./...` passes.
- [ ] `golangci-lint run` CI clean.
- [ ] `gitleaks` CI clean.
- [ ] `BONSAI_ASCII_ONLY=1 bonsai init` uses ASCII fallback throughout.
- [ ] Legacy code (`runInitLegacy`, `BONSAI_REDESIGN` env check) removed; grep returns nothing.
- [ ] Manual visual smoke: 70-col (floor), 80-col, 120-col, 200-col widths render without border drift.
- [ ] Sub-floor terminals (<70×20) show the "please enlarge" panel instead of a broken layout.
- [ ] Branches + Observe + Planted lists scroll vertically when contents exceed available height; focused row stays visible.
- [ ] Ctrl-C at any stage exits cleanly with no partial `.bonsai.yaml`.
- [ ] Esc pops stage-level (Branches → Soil, etc.) with prior selections preserved.
- [ ] `station/Playbook/Status.md` and `station/agent/Core/memory.md` updated.
