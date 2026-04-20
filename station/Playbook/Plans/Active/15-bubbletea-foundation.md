# Plan 15 — BubbleTea Foundation + Theme System

**Tier:** 2 (Feature)
**Status:** Active
**Source:** Plan 14 deferred items (TUI screen lifecycle, progressive disclosure, go-back navigation)
**Agent:** general-purpose

---

## Goal

Stand up a single long-lived BubbleTea harness that owns the screen for the lifetime of an interactive Bonsai command, and migrate `bonsai init` onto it as the first proof.

### Success Criteria (Iter 1)

- New `internal/tui/harness` package exposes a step-stack `tea.Model` plus adapters wrapping the existing Huh widgets (`Text`, `Select`, `MultiSelect`, `Confirm`, `Review`, `Lazy`).
- `bonsai init` runs entirely inside one `tea.Program` with `tea.WithAltScreen()` for the interactive portion.
- A persistent header (banner + action + `[N/M] Title` crumb) and footer (key hints) frame the active step.
- `Esc` / `Shift+Tab` returns to the previous step with the prior answer preserved; `Esc` on step 1 is a no-op.
- `Ctrl-C` exits cleanly with no partial files written.
- After the harness exits AltScreen, the existing spinner / write-result / success banner render to normal stdout exactly as today — terminal scrollback before the command is preserved on exit.
- All existing semantic tokens, glyphs, panels, and trees from Plan 11/12/14 still render correctly.
- No regressions on `bonsai add`, `bonsai remove`, `bonsai update`, `bonsai list`, `bonsai catalog` (unchanged in iter 1).

---

## Context

Plan 14 (UI/UX Phase 3) explicitly deferred screen lifecycle, progressive disclosure, and go-back navigation to Phase 4+ because they need a real architectural pass — those items are blocked by the current TUI being a chain of stateless `fmt.Println` calls and short-lived Huh forms. Each `bonsai init` step prints to scrollback, the next form opens, and prior answers vanish into the wall of text. Plan 14's `Answer()` helper papered over this; we now want the real fix.

Plan 15 introduces a single long-lived BubbleTea harness that owns the screen for the lifetime of an interactive command. Existing Huh widgets continue to do the input collection — they're already `tea.Model`s, so we compose them inside the harness rather than spawning each one with its own `Form.Run()`. Semantic tokens, glyphs, panels, and trees from Plan 11/12/14 stay as-is; only the orchestration layer changes.

**Decisions captured during planning:**

- **One program per command.** A single `tea.Program` with AltScreen runs the interactive portion; we exit AltScreen for spinner/result/success output to keep transcript-style logs.
- **Composition, not replacement.** Huh widgets stay — adapters delegate `Init/Update/View` to an embedded `*huh.Form`.
- **Step stack, not screen graph.** Linear flow with `Esc`-to-pop. Branching (e.g., add-existing vs. add-new agent) is handled via `LazyStep` extending the stack at runtime, not a separate router.
- **Theme tokens unchanged.** Zen Garden palette + semantic tokens + `BonsaiTheme()` stay in place. Only three new harness-specific style vars (`HarnessHeader`, `HarnessCrumb`, `HarnessFooter`) are added.
- **Local-only iteration.** Same model as Plan 14 — three iterations, user drives ship cadence, no PRs per iteration.

---

## Architecture

```
┌──────────────────────── tea.Program (AltScreen) ────────────────────────┐
│ ┌─ Header ────────────────────────────────────────────────────────────┐ │
│ │  BONSAI v0.1.3        Initializing new project    [3/7] Protocols   │ │  ← banner + crumb
│ └─────────────────────────────────────────────────────────────────────┘ │
│ ┌─ Content (active step) ─────────────────────────────────────────────┐ │
│ │                                                                     │ │
│ │   Project name:                                                     │ │  ← huh.Form
│ │   ▌ my-project                                                      │ │     (embedded
│ │                                                                     │ │      tea.Model)
│ │   ▸ Project name   my-project          ← prior Answer() chips       │ │
│ │   ▸ Description    (skipped)                                        │ │
│ │                                                                     │ │
│ └─────────────────────────────────────────────────────────────────────┘ │
│ ┌─ Footer ────────────────────────────────────────────────────────────┐ │
│ │  ↵ continue   esc back   ctrl-c quit                                │ │
│ └─────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘

           Step stack (held by Harness model)
           ┌──────────────────┐
           │  TextStep        │ ← active
           │  TextStep (done) │
           │  TextStep (done) │   esc pops back here, value preserved
           └──────────────────┘
```

**Key contract:** the harness runs one `tea.Program` for the entire interactive command. When the flow completes, the program exits AltScreen and the caller resumes normal stdout for the spinner / write-result / success banner. (Iter 2/3 will fold the spinner into a step too.)

---

## Iterations

| Iter | Scope | Status |
|------|-------|--------|
| 1 | Harness package + theme + `cmd/init.go` migration | Shipped (ui-ux-testing @ 150d1d3) |
| 2 | Migrate `cmd/add.go` (incl. `runAddItems` pivot) + `NoteStep` + `TitledPanelString` + harness `LazyGroup` splice | Shipped (ui-ux-testing @ 4011882) |
| 2.1 | Post-ship reviewer fixes — stale review panel on Esc-back, tech-lead bootstrap, all-installed zero-keystroke, defensive harness guards | Shipped (ui-ux-testing @ d0e6256) |
| 3 | Harness primitives (`SpinnerStep`, `ConditionalStep`, splice WindowSize re-broadcast, Splice/Build panic recovery) + carry-forward reviewer nits (nested-splicer docstring, LazyStep-in-LazyGroup test, workspace validator filepath.Clean) + migrate `cmd/remove.go` + `cmd/update.go` + retro-fit `cmd/init.go`/`cmd/add.go` to use SpinnerStep + conflict-picker LazyGroup so Ctrl-C during generate is the same clean-exit path as the interactive portion | Shipped (ui-ux-testing @ a406908) |
| 3.1 | Reviewer fixes — `cmd/add.go` conflict-picker index regression (literal `3` discarded user picks; now computes from `len(results)-2` when `wr.HasConflicts()`). 4 non-fix nits routed to Backlog Group F. | Shipped (ui-ux-testing) |

---

## Iter 1 — Detailed Steps

### Step 1 — New `internal/tui/harness` package

**New file:** `internal/tui/harness/harness.go`

Defines:

```go
package harness

type Step interface {
    tea.Model               // Init/Update/View
    Title() string          // breadcrumb label
    Result() any            // value produced; nil while pending
    Done() bool             // signal the harness to advance
}

type Harness struct {
    steps    []Step
    cursor   int
    width    int
    height   int
    banner   string         // "BONSAI v0.1.3"
    action   string         // "Initializing new project"
    quitting bool
    aborted  bool           // user pressed ctrl-c
}

func New(banner, action string, steps []Step) *Harness
func (h *Harness) Init() tea.Cmd
func (h *Harness) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (h *Harness) View() string

// Run drives the harness to completion under tea.WithAltScreen.
// Returns the per-step results in declaration order, or ErrAborted.
func Run(banner, action string, steps []Step) ([]any, error)

var ErrAborted = errors.New("flow aborted by user")
```

Update reducer rules:
- `tea.WindowSizeMsg` → store width/height, broadcast to current step.
- `tea.KeyMsg`:
  - `ctrl+c` → set `aborted=true`, return `tea.Quit`.
  - `esc` (or `shift+tab`) → if `cursor > 0`, pop to `cursor-1`; else ignore.
  - else → forward to active step.
- After forwarding, if `step.Done()` is true, advance: `cursor++`. If `cursor == len(steps)`, return `tea.Quit`.

View composition:
- `header(width)` + `\n` + `step.View()` clipped to `height - headerH - footerH` + `\n` + `footer(width)`.
- Header renders banner left, action middle, `[N/M] Title` right — built from existing `tui.StyleTitle`, `tui.StyleMuted`, `tui.HarnessCrumb`.
- Footer renders key hints with `tui.HarnessFooter`.
- Follows BubbleTea Golden Rule #1 — height calculations subtract 2 for borders before rendering bordered panels.

### Step 2 — Step adapters wrapping Huh forms

**New file:** `internal/tui/harness/steps.go`

Six adapters:

| Adapter | Wraps | Result type |
|---------|-------|-------------|
| `TextStep` | `huh.NewInput()` | `string` |
| `SelectStep` | `huh.NewSelect[string]()` | `string` |
| `MultiSelectStep` | `huh.NewMultiSelect[string]()` (with required/optional split logic from existing `tui.PickItems`) | `[]string` |
| `ConfirmStep` | `huh.NewConfirm()` | `bool` |
| `ReviewStep` | static `tui.ItemTree` panel + `huh.NewConfirm()` | `bool` |
| `LazyStep` | builds itself on entry from `func(prev []any) Step` | wrapped step's result |

Each adapter holds a `*huh.Form` and delegates `Init/Update/View` to it. `Done()` returns `form.State == huh.StateCompleted`. `Result()` returns the captured value.

For `MultiSelectStep`, port the required/collapsed-chip-line rendering logic from `internal/tui/prompts.go:175-198` so the visual behavior is preserved.

Apply `tui.BonsaiTheme()` in each adapter via `form.WithTheme(tui.BonsaiTheme())`.

`LazyStep` exists so the review step (and later, branching flows in iter 2/3) can construct themselves with access to prior step results without leaving AltScreen.

### Step 3 — Theme split

**Modified file:** `internal/tui/styles.go`

Add a new section after the `Panels` block (after line 107):

```go
// ─── Harness Styles ──────────────────────────────────────────────────────
var (
    HarnessHeader = lipgloss.NewStyle().Padding(0, 2).Foreground(ColorMuted)
    HarnessCrumb  = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
    HarnessFooter = lipgloss.NewStyle().Padding(0, 2).Foreground(ColorMuted)
)
```

`BonsaiTheme()` stays in `prompts.go` and is still applied by every step adapter (so a non-harness Huh call site keeps working during migration).

### Step 4 — Migrate `cmd/init.go`

**Modified file:** `cmd/init.go`

Replace the body of `runInit` (lines 28–211) so the interactive portion runs through the harness:

```go
func runInit(cmd *cobra.Command, args []string) error {
    cwd := mustCwd()
    configPath := filepath.Join(cwd, configFile)
    if _, err := os.Stat(configPath); err == nil {
        tui.WarningPanel(configFile + " already exists. Skipping init.")
        return nil
    }
    cat := loadCatalog()

    techLeadType := "tech-lead"
    agentDef := cat.GetAgent(techLeadType)
    if agentDef == nil {
        tui.FatalPanel("Tech Lead agent not found", ...)
    }

    steps := []harness.Step{
        harness.NewText("Project name", "Project name:", "", true),
        harness.NewText("Description", "Description (optional):", "", false),
        harness.NewText("Station directory", "Station directory:", "station/", true),
        harness.NewMultiSelect("Scaffolding", "Project Scaffolding", scaffoldingOptions(cat), nil),
        harness.NewMultiSelect("Skills",     "Skills",     toItemOptions(cat.SkillsFor(techLeadType), techLeadType),     agentDef.DefaultSkills),
        harness.NewMultiSelect("Workflows",  "Workflows",  toItemOptions(cat.WorkflowsFor(techLeadType), techLeadType),  agentDef.DefaultWorkflows),
        harness.NewMultiSelect("Protocols",  "Protocols",  toItemOptions(cat.ProtocolsFor(techLeadType), techLeadType),  agentDef.DefaultProtocols),
        harness.NewMultiSelect("Sensors",    "Sensors",    sensorOpts(cat, techLeadType), agentDef.DefaultSensors),
        harness.NewMultiSelect("Routines",   "Routines",   routineOpts(cat, techLeadType), agentDef.DefaultRoutines),
        harness.NewLazy("Review", buildReviewStep(cat, agentDef)),
    }

    results, err := harness.Run(
        fmt.Sprintf("BONSAI v%s", Version),
        "Initializing new project",
        steps,
    )
    if err != nil {
        if errors.Is(err, harness.ErrAborted) {
            return nil
        }
        return err
    }

    // Pull typed results, validate, build config, then run generate as today.
    // Generation, spinner, conflict resolution, success banner stay outside the harness in iter 1.
}
```

`buildReviewStep` returns a `func(prev []any) harness.Step` closure that constructs a `ReviewStep` with the prior selections.

Validation of `docsPath` (currently lines 57–63) moves into the `TextStep` validator so the user can correct in place rather than crashing after.

The post-harness block (config save, spinner, conflict resolve, write-result, success banner) stays unchanged — runs after `harness.Run` returns and the program has exited AltScreen.

### Step 5 — Tests

**New file:** `internal/tui/harness/harness_test.go`

Reducer-only tests (no TTY needed):

1. `TestHarnessAdvancesOnDone` — fake step with `Done()=true` → cursor advances.
2. `TestHarnessQuitsAfterLastStep` — when cursor reaches `len(steps)`, `Update` returns `tea.Quit`.
3. `TestEscPopsCursor` — cursor=2, send `KeyEsc`, expect cursor=1.
4. `TestEscOnFirstStepIgnored` — cursor=0, send `KeyEsc`, expect cursor=0 and no quit.
5. `TestCtrlCSetsAbortedAndQuits` — sets `aborted=true`, returns `tea.Quit`.
6. `TestWindowSizeBroadcasts` — fake step records `WindowSizeMsg` width.
7. `TestLazyStepBuildsOnEntry` — verify the closure is invoked once when the cursor advances onto it, with prior results passed in.

`fakeStep` is a small struct that records messages and exposes `done bool` to flip in-test.

Existing `internal/tui/styles_test.go` stays unchanged.

---

## Iter 2 — Detailed Steps

### Goal

Migrate `cmd/add.go` (`runAdd` + `runAddItems`) onto the harness inside a single `tea.Program`, introduce the two iter-1 follow-up primitives (`NoteStep` adapter + `tui.TitledPanelString` helper), and extend the harness with multi-step branching (`LazyGroup`) so the two add-flow shapes can coexist in one step list.

### Success Criteria (Iter 2)

- `bonsai add` runs entirely inside one `tea.Program` with `tea.WithAltScreen()` for the interactive portion (same model as `bonsai init`).
- Agent-type selection is step 1; a branch point at step 2 splices in either the "configure new agent" sub-sequence or the "add items to existing agent" sub-sequence without leaving AltScreen.
- Tech-lead special-case (workspace auto-derives from `cfg.DocsPath`) renders as a `NoteStep` — no text input, just an info panel the user advances past with Enter.
- For `runAddItems`, categories with zero uninstalled abilities are skipped (no empty-picker step); if **all** categories are empty, the flow short-circuits pre-harness with the existing `EmptyPanel`.
- The review panel uses the proper bordered `TitledPanel` look via a new string-returning helper (`tui.TitledPanelString`); `cmd/init.go`'s lightweight `buildReviewPanel` from iter 1 is refactored to use the same helper.
- `Esc`/`Shift+Tab`, `Ctrl-C`, scrollback preservation, and header/footer behaviour match iter 1 exactly for the add flows.
- No regressions on `bonsai init` (iter 1), `bonsai remove`, `bonsai update`, `bonsai list`, `bonsai catalog`.

---

### Step 1 — Harness `LazyGroup` / Splicer

**Modified file:** `internal/tui/harness/harness.go`
**Modified file:** `internal/tui/harness/steps.go`

Add a new adapter that expands into multiple steps on entry. Needed because the existing `LazyStep` builds a single inner step — insufficient for the `runAdd` vs `runAddItems` fork, each of which is a multi-step sub-sequence.

**New type in `steps.go`:**

```go
// LazyGroup is a placeholder step that, on first entry, expands into a slice of
// steps spliced into the harness at its position. Used for multi-step branches
// (e.g. "configure new agent" vs "add items to existing agent"). The builder
// runs once with prior results in scope.
type LazyGroup struct {
    title string
    build func(prev []any) []Step
    built bool
}

func NewLazyGroup(title string, build func(prev []any) []Step) *LazyGroup
```

`LazyGroup` satisfies `Step` but its `View()`/`Update()` are never actually driven — the harness splices it out before the user sees a frame.

**New interface in `harness.go`:**

```go
// splicer is implemented by LazyGroup. When the harness cursor advances onto a
// splicer, the group is replaced in-place with the steps it returns, and the
// cursor stays at the same index (now pointing at the first of the new steps).
type splicer interface {
    Splice(prev []any) []Step
    Spliced() bool
}
```

**Reducer change:** in the step-advance loop inside `Harness.Update` (lines 172–187 of iter 1's `harness.go`), after the existing `lazyBuilder` check, add:

```go
if sp, ok := h.steps[h.cursor].(splicer); ok && !sp.Spliced() {
    inserted := sp.Splice(h.priorResults())
    // Replace the group in-place with its expansion.
    h.steps = append(append(append([]Step{}, h.steps[:h.cursor]...), inserted...), h.steps[h.cursor+1:]...)
    // cursor already points at the first spliced step; init it.
    if lb, ok := h.steps[h.cursor].(lazyBuilder); ok && !lb.Built() {
        lb.Build(h.priorResults())
    }
    if init := h.steps[h.cursor].Init(); init != nil {
        cmd = tea.Batch(cmd, init)
    }
    continue
}
```

Splice runs exactly once (guarded by `Spliced()`). After splice, the group is gone from the list — re-popping via Esc-back onto an earlier step never re-visits it.

**`priorResults()` stays correct after splice** because the splice happens at `h.cursor`, so only future-cursor steps are rewritten — previously-completed steps keep their results in the same slots.

**Tests (`harness_test.go`):**
- `TestLazyGroupSplicesOnEntry` — 3-step list where step 1 is a `LazyGroup` that returns `[fakeStep, fakeStep]`. Advance past step 0, verify `h.steps` is now 4 long (orig 3 − 1 group + 2 spliced), cursor stays at 1, `Init` of new step 1 was invoked.
- `TestLazyGroupRunsOnce` — flipping `fakeStep.done` twice to re-trigger advance logic; confirm the group's `Build`/`Splice` only fires on the first entry.
- `TestLazyGroupPassesPriorResults` — step 0 is a `fakeStep` with `result="agent-x"`; group's `build(prev)` asserts `prev[0]=="agent-x"`.

---

### Step 2 — `NoteStep` adapter

**Modified file:** `internal/tui/harness/steps.go`

Wraps `huh.NewNote().Next(true)` so the user can press Enter to advance past a static informational block. Used for:
- Tech-lead workspace shortcut (no text input; just shows the auto-derived path).
- `runAddItems` intro line ("X is already installed at Y — showing uninstalled abilities").

```go
// NoteStep wraps huh.NewNote — a static information block the user advances
// past by pressing Enter. Produces no result.
type NoteStep struct {
    title string
    body  string
    form  *huh.Form
}

func NewNote(title, body string) *NoteStep {
    step := &NoteStep{title: title, body: body}
    step.form = step.buildForm()
    return step
}

func (s *NoteStep) buildForm() *huh.Form {
    note := huh.NewNote().
        Title(s.title).
        Description(s.body).
        Next(true)
    return huh.NewForm(huh.NewGroup(note)).WithTheme(tui.BonsaiTheme())
}

func (s *NoteStep) Title() string                              { return s.title }
func (s *NoteStep) Done() bool                                 { return s.form.State == huh.StateCompleted }
func (s *NoteStep) Result() any                                { return nil }
func (s *NoteStep) Init() tea.Cmd                              { return s.form.Init() }
func (s *NoteStep) View() string                               { return s.form.View() }
func (s *NoteStep) Update(msg tea.Msg) (tea.Model, tea.Cmd)    {
    f, cmd := s.form.Update(msg)
    if ff, ok := f.(*huh.Form); ok { s.form = ff }
    return s, cmd
}
func (s *NoteStep) Reset() tea.Cmd                             { s.form = s.buildForm(); return s.form.Init() }
```

Follow the `buildForm()` rebuild-on-Reset pattern already established in iter 1 for every other adapter (see `TextStep.buildForm` docstring in iter 1's `steps.go` for the huh `f.quitting` rationale).

**Tests (`steps_test.go`):**
- `TestNoteStepViewNonEmpty` — after construction, `View()` returns a non-empty string.
- `TestNoteStepResetRenders` — force `form.State = StateCompleted`, call `Reset()`, assert `View()` is still non-empty (same shape as the existing per-adapter Reset guards for TextStep/SelectStep/etc.).

---

### Step 3 — `tui.TitledPanelString` helper

**Modified file:** `internal/tui/styles.go`

Split the existing `TitledPanel` into a string-returning core + a `Println` wrapper so the harness can render the proper bordered review look inside AltScreen (`fmt.Println` is incompatible with AltScreen — that's why iter 1 shipped with a lightweight `buildReviewPanel`).

```go
// TitledPanelString renders the same bordered panel as TitledPanel and returns
// the result as a string. Use inside AltScreen or when composing with other
// styled content.
func TitledPanelString(title, content string, color lipgloss.TerminalColor) string {
    // ... body of current TitledPanel MINUS the final fmt.Println ...
    return "\n" + buf.String()
}

// TitledPanel renders the bordered panel to stdout.
func TitledPanel(title, content string, color lipgloss.TerminalColor) {
    fmt.Println(TitledPanelString(title, content, color))
}
```

Move every line of the current `TitledPanel` body (styles.go lines 291–356) into `TitledPanelString`, except the final `fmt.Println("\n" + buf.String())` which becomes the one-line `TitledPanel` wrapper's body.

**Refactor iter 1's `cmd/init.go::buildReviewPanel`** (lines 250–288 of current iter 1 `init.go`) to use `TitledPanelString` instead of the ad-hoc heading + tree. Replace the final two lines:

```go
// before (iter 1):
heading := tui.StyleAccent.Bold(true).Render("Review")
return "  " + heading + "\n" + tree

// after (iter 2):
return tui.TitledPanelString("Review", tree, tui.Water)
```

**Tests (`styles_test.go`):**
- `TestTitledPanelStringIncludesTitle` — returned string contains the title rendered in color.
- `TestTitledPanelStringMultilineBody` — body with two lines is preserved (at least one line per content line present).
- `TestTitledPanelPrintsSameAsString` — capture `os.Stdout`, call `TitledPanel`, verify it equals `TitledPanelString(...)` + trailing newline.

---

### Step 4 — Migrate `cmd/add.go` `runAdd` + `runAddItems`

**Modified file:** `cmd/add.go`

**Pre-harness (runs before `harness.Run`):**

1. `cwd := mustCwd()`, load config + catalog, early-exit guards (`requireConfig`) unchanged.
2. Remove the standalone `tui.Heading("Add Agent")` call — the harness header replaces it.
3. Build `agentOptions` (same as today).
4. Build `existingWorkspaces` map once (used by the workspace-step validator inside the splice).
5. **Pre-flight: require tech-lead.** If `cfg.Agents["tech-lead"]` doesn't exist, render `tui.ErrorDetail("Tech Lead required", "No tech-lead agent is installed yet.", "Run: bonsai init")` and return `nil` *without entering the harness*. This is the existing behaviour at `cmd/add.go:85-90` — keep it pre-harness so the user isn't invited into AltScreen only to be rejected. Consequence: we never reach the harness without at least tech-lead installed, so the `agentType != "tech-lead" && !hasTechLead` check inside the splice is unnecessary.

**In-harness step list:**

```go
steps := []harness.Step{
    harness.NewSelect("Agent", "Agent type:", agentOptions),
    harness.NewLazyGroup("Agent flow", func(prev []any) []harness.Step {
        agentType := asString(prev[0])
        if _, exists := cfg.Agents[agentType]; exists {
            return buildAddItemsSteps(cat, cfg, agentType) // runAddItems branch
        }
        return buildNewAgentSteps(cat, cfg, agentType, existingWorkspaces)
    }),
}
```

**`buildNewAgentSteps(cat, cfg, agentType, existingWorkspaces)` (runAdd branch):**

1. **Workspace step** — built via `harness.NewLazy` (single-step lazy, already shipped in iter 1) so it can branch on agent type:
   - If `agentType == "tech-lead"`: `NoteStep(title="Workspace", body="Tech Lead workspace: {cfg.DocsPath or 'station/'}")`.
   - Else: `TextStep(title="Workspace", prompt="Workspace directory (e.g. backend/):", default=agentType+"/", required=true, validator=workspaceUniqueValidator(existingWorkspaces))`. The validator rejects any path already in `existingWorkspaces[]` so the user corrects in-place without a crash after collection.
2. `MultiSelectStep` × 5 — Skills, Workflows, Protocols, Sensors (filtered to exclude `routine-check`, same as `cmd/add.go:137-144`), Routines. Same defaults as today (`agentDef.DefaultSkills` etc.).
3. `LazyStep` — Review: builds a `ReviewStep(title="Review", panel=tui.TitledPanelString("Review", tree, tui.Water), prompt="Generate files?", default=true)` where `tree = tui.ItemTree(...)` mirrors `cmd/add.go:168-178`.

**`buildAddItemsSteps(cat, cfg, agentType)` (runAddItems branch):**

1. Compute filtered lists (`filterItems`/`filterSensors`/`filterRoutines` — lift the existing closures from `cmd/add.go:251-282` into package-level helpers).
2. If `len(newSkills)+len(newWorkflows)+len(newProtocols)+len(newSensors)+len(newRoutines) == 0`: return `[]Step{NoteStep("All installed", "All available abilities are already installed.\nBrowse more with: bonsai catalog")}` — the splice becomes a single NoteStep, user presses Enter, flow ends with no writes.
3. Otherwise, return:
   - `NoteStep(title="Adding to {agent}", body="{agentDef.DisplayName} is already installed at {installed.Workspace} — showing uninstalled abilities.")`
   - `MultiSelectStep` for each category *only if the filtered list is non-empty* — zero-item categories omitted from the step list entirely (no empty-picker step).
   - `LazyStep` — Review: builds a `ReviewStep(title="Adding", panel=tui.TitledPanelString("Adding", tree, tui.Water), prompt="Generate files?", default=true)` where tree shows only newly-selected items.

**Post-harness (runs after `harness.Run`):**

1. Handle `ErrAborted` (Ctrl-C) → return `nil`.
2. Branch on whether the agent existed pre-harness (recompute `_, exists := cfg.Agents[agentType]`):
   - **New agent:** extract workspace (string from NoteStep result is `nil`, so for tech-lead use the pre-computed `cfg.DocsPath or "station/"` directly; for others pull from `asString(results[...])`), validate workspace uniqueness against `existingWorkspaces` once more (defence-in-depth — the validator already blocked duplicates, but the check is cheap), extract the 5 picker results, build `InstalledAgent`, call `generate.EnsureRoutineCheckSensor`, save config, run spinner + 4 `generate.*` calls + `resolveConflicts` + `lock.Save` + `showWriteResults` + `tui.Success("Added X at Y")` + `tui.Hint("bonsai list…")` — exactly the pipeline at `cmd/add.go:188-228`.
   - **Existing agent:** if the splice short-circuited with the "All installed" NoteStep, `results[...]` for the review slot is absent (or the review confirm never happened) — detect by checking whether any picker produced a non-empty slice. If all empty, return `nil` with no writes. Otherwise append selections to `installed.*` slices, re-run `EnsureRoutineCheckSensor`, save config, same spinner + generate + conflict + lock + write-results + `tui.Success("Added N abilities to X")` + hint — exactly `cmd/add.go:360-397`.

**Helper extraction:** The `describer` closure (currently inlined twice at `cmd/add.go:155-166` and `:327-338`) should be lifted to a package-level `func newDescriber(cat *catalog.Catalog) func(string) string` so both branches and the review builders share one. Small cleanup, within scope because both flows now call it from the `LazyStep` review builder.

**Shared post-harness pipeline** (same as iter 1 init):
- `spinner.New().Action(...).Run()` for generation
- `resolveConflicts(&wr, lock, cwd)` if conflicts
- `lock.Save(cwd)`
- `showWriteResults(&wr, workspace)`
- `tui.Success(...)`, `tui.Hint(...)`, `tui.Blank()`

**Action label:** Pass `"Adding"` as the harness action label (generic — covers both branches). Breadcrumb `[N/M] Title` surfaces specifics.

---

### Step 5 — Tests

**New: `internal/tui/harness/harness_test.go` additions** — 3 `LazyGroup` tests (listed in Step 1).

**New: `internal/tui/harness/steps_test.go` additions** — 2 `NoteStep` tests (Step 2).

**New: `internal/tui/styles_test.go` additions** — 3 `TitledPanelString` tests (Step 3).

**Reducer-only for everything** — no TTY dependency, same pattern as iter 1.

---

### Iter 2 — Verification

#### Build & Test

- [x] `make build` — clean compile. ✅ (verified @ 4011882 and re-verified @ d0e6256 post iter-2.1)
- [x] `go test ./...` — all tests green, incl. new `LazyGroup`, `NoteStep`, `TitledPanelString` cases. ✅
- [x] `gofmt -s -l .` — no output. ✅
- [x] `go vet ./...` — no issues. ✅

> **Manual smoke sections below are deferred to the whole-branch merge audit before `ui-ux-testing → main`.** No PTY in dispatched environment; tech lead to walk the flows locally at iter-3 completion.

#### Manual — `bonsai add` flow (new agent)

Run in a temp project after `bonsai init`:

- [ ] AltScreen activates; scrollback before the command preserved on exit.
- [ ] Header shows `BONSAI vX.Y.Z` left, `Adding` middle, `[1/N] Agent` right on step 1.
- [ ] Pick `backend` → step 2 shows workspace TextStep with `backend/` default.
- [ ] Pick `tech-lead` (if not yet installed — skip this case for this test) → step 2 shows NoteStep with "Tech Lead workspace: …".
- [ ] Fill each multi-select; required-only sections auto-complete with chip-line (same as init).
- [ ] Review panel renders with the bordered `TitledPanel` look (iter 2 upgrade from iter 1 lightweight).
- [ ] Confirm Yes → harness exits AltScreen → spinner + write-results + success banner render to normal stdout.
- [ ] `Esc` on step 2+ pops back, prior answer visible.
- [ ] `Ctrl-C` exits cleanly, no `.bonsai.yaml` mutation, no partial files.

#### Manual — `bonsai add` flow (existing agent, pivot to add-items)

- [ ] After an agent is installed, re-run `bonsai add` and pick the same agent.
- [ ] Step 2 splices in the add-items branch: NoteStep intro + MultiSelectStep(s) for categories with ≥1 uninstalled ability.
- [ ] Categories with 0 uninstalled items are absent from the breadcrumb (e.g., `[3/5]` not `[3/7]` if two categories were skipped).
- [ ] Review panel shows only newly-picked items under "Adding" (not the full agent inventory).
- [ ] Confirm Yes → new selections append to `.bonsai.yaml`; existing selections untouched.
- [ ] Re-run with nothing left to add → single NoteStep "All installed" renders; pressing Enter exits cleanly with no writes.

#### Manual — regressions

- [ ] `NO_COLOR=1 bonsai add` — zero ANSI escapes in output.
- [ ] `bonsai init` — unchanged iter 1 behaviour (now also uses `TitledPanelString` for the review panel — verify the bordered look is intact and scrollback still preserves).
- [ ] `bonsai remove` — unchanged (still on stateless path; iter 3 migrates it).
- [ ] `bonsai update` — unchanged.
- [ ] `bonsai list` / `bonsai catalog` — unchanged.

---

## Iter 2.1 — Fixes (shipped)

Three independent post-ship reviews of iter 2 (commit `4011882`) surfaced four real regressions / UX misses. All landed as a single follow-up commit `d0e6256` on `ui-ux-testing`.

### Fixes landed

**A. Stale review panel after Esc-back.** `LazyStep.Reset()` now clears `built=false` and drops `inner` so the builder closure re-runs against current prior results on re-entry. The harness Esc reset loop was also extended from `[new_cursor, origCursor)` to `[new_cursor, origCursor]` inclusive — a tail `LazyStep` (review) at `origCursor` needs `Reset()` so its next activation rebuilds the panel after the user edits upstream picks. New test `TestLazyStepRebuildsOnReset`; existing `TestEscPopReinitsActiveStep` updated to reflect the new inclusive bound. Affects both `cmd/init.go` and `cmd/add.go` review steps.

**B. Tech-lead bootstrap regression.** Iter 2's pre-harness "require tech-lead" gate blocked `bonsai add` from running at all without a tech-lead — but the catalog's intent is that users *pick* tech-lead from the list to bootstrap. Removed the pre-harness unconditional block. Non-tech-lead picks without an installed tech-lead now show an in-harness `NoteStep` (cosmetic) plus a post-harness `tui.ErrorDetail` to stdout (durable scrollback record, matches pre-iter-2 error UX).

**C. "All installed" path no longer blocks on stdin.** Filter logic lifted into a shared `availableAddItems` helper (+ `availableAddSet` type). The `LazyGroup` splicer now returns a nil / empty slice when every category filters empty; the post-harness path detects this and renders `tui.EmptyPanel` to stdout with zero keystrokes (matches pre-iter-2 behaviour — iter 2 regressed by forcing the user to Enter-past a NoteStep before printing the empty banner).

**D. Defensive harness guards.** `expandSplicer` filters `nil` steps from splice output before installing them (was a panic if any builder returned a slice with a nil element). `View()` short-circuits with a muted "terminal too small" notice when available body height drops below 3 rows, so tiny terminals get a readable message instead of a broken frame.

### Verification (re-run after 2.1)

- [x] `go build ./...` — clean ✅
- [x] `go vet ./...` — clean ✅
- [x] `gofmt -s -l .` — no output ✅
- [x] `go test ./... -count=1` — all packages pass (4 ok, 0 failed) ✅

### Deferred to iter 3 (or beyond)

Reviewer findings that were NOT hotfixed in 2.1:
- **Nested splicers silently unsupported** (iter-2 nit #1) — docstring note needed on `splicer` interface; no caller hits this yet.
- **Empty-splice re-entry edge case** (iter-2 nit #2) — cursor lands on former `cursor+1` without another `expandSplicer()` call; same shape as above, safe for current callers.
- **`WindowSizeMsg` not re-broadcast after splice** (iter-2 nit #3) — first-frame layout on spliced step *may* be briefly wrong at narrow widths until huh's next update cycle.
- **No unit test for `LazyStep` nested inside a `LazyGroup` result** (iter-2 nit #4) — exercised at runtime via `cmd/add.go` but not isolated.
- **Workspace validator normalization** — current `workspaceUniqueValidator` compares raw user input against existing paths; trailing-slash and `./` variants slip past. Filed to Backlog Group F.
- **Conditional-skip step** — `buildAddItemsSteps` filters zero-item categories manually; a reusable `ConditionalStep` adapter would be cleaner. Filed to Backlog Group F.
- **Panic recovery around `Splice`/`Build`** — a builder that panics today will crash the whole `tea.Program`. Filed to Backlog Group F.
- **Spinner Ctrl-C partial-write window** — pre-existing in `cmd/add.go` and `cmd/init.go`; the `huh/spinner` block runs outside the harness, so Ctrl-C during generation can leave partial files. Iter 3 `SpinnerStep` migration addresses this. Filed to Backlog Group F as a waypoint.

---

## Iter 3 — Detailed Steps

### Goal

Close the harness story for the four interactive commands (`init`, `add`, `remove`, `update`). Add the two missing primitives — `SpinnerStep` and `ConditionalStep` — so the whole flow (input → generate → conflict resolution) runs inside one `tea.Program` per command. Migrate `remove.go` + `update.go` onto the harness, and retro-fit `init.go`/`add.go` to fold their existing `huh/spinner` + `resolveConflicts` calls into the harness so Ctrl-C during generation is the same clean-exit path as the interactive portion. Fold in the small carry-forward reviewer nits from iter 2.1 (panic recovery, splice WindowSize re-broadcast, validator normalization, splicer docstring, nested-lazy unit test) along the way.

### Success Criteria (Iter 3)

- New `SpinnerStep` adapter renders a tick spinner while a blocking generate function runs on a worker goroutine, completing when the worker returns. Ctrl-C while the spinner is active aborts via the harness (no terminal mess, AltScreen exits cleanly). The pre-existing partial-file window remains because `generate.*` is uncontextualised — but the spinner no longer lives outside the harness, so the harness's Ctrl-C handling is the only exit path.
- New `ConditionalStep` adapter wraps another step plus a `predicate(prev []any) bool`. When the predicate returns false the wrapped step never renders — the harness sees `AutoComplete()=true` and `Done()=true` on `Init`, and `Result()=nil`.
- `expandSplicer` (and the LazyStep advance path) re-broadcast a `tea.WindowSizeMsg` carrying the harness's stored `width/height` to the freshly-active step so its first frame layout is correct at narrow widths instead of waiting for huh's next update cycle.
- `Splice()`/`Build()` invocations are wrapped in `recover()` — a panic in a builder closure now exits AltScreen cleanly and the post-harness path renders a `tui.FatalPanel` with the recovered value, instead of dumping a stacktrace mid-AltScreen.
- `splicer` interface docstring explicitly notes that nested splicers (a `LazyGroup` whose `Splice()` returns another `LazyGroup`) are not supported in iter 3 — the second splice would be skipped because the advance loop only calls `expandSplicer()` once per advance. Left as a documented limitation since no caller hits it.
- New unit test exercises a `LazyStep` returned inside a `LazyGroup` splice (the runtime pattern used by `cmd/add.go` and the iter-3 review steps). Covers: build runs after splice, prior results visible, Reset rebuilds inner.
- `cmd/add.go::workspaceUniqueValidator` normalizes via `filepath.Clean` before the trim/append-slash compare so `./backend`, `backend`, and `backend/` all collide if any variant is in use. Same normalization applied to the post-harness `normaliseWorkspace`.
- `bonsai remove` (agent-level) runs entirely inside one `tea.Program`: review panel → confirm → spinner → optional conflict picker. Esc-back from confirm pops to the (auto-completed) review step, which behaves correctly. Ctrl-C aborts cleanly.
- `bonsai remove <type> <name>` runs entirely inside one `tea.Program`: optional agent-picker (skipped via `ConditionalStep` if only one agent has the item) → summary panel → confirm → spinner → optional conflict picker.
- `bonsai update` runs entirely inside one `tea.Program`: per-agent custom-file pickers (each gated by `ConditionalStep` so agents with zero discoveries don't show empty pickers) → spinner → optional conflict picker.
- `bonsai init` and `bonsai add` retro-fitted: their existing review confirm now leads directly into the harness's spinner + conflict-picker steps. The `spinner.New().Action(...).Run()` and `resolveConflicts(...)` calls move from post-harness blocks into harness step builders.
- `showWriteResults` + `tui.Success` + `tui.Hint` stay outside the harness on normal stdout — those are scrollback-friendly transcript output the user expects to keep.
- No regressions on `bonsai list` / `bonsai catalog` (still non-interactive — out of scope).

---

### Step 1 — `SpinnerStep` adapter

**Modified file:** `internal/tui/harness/steps.go`

Wraps `github.com/charmbracelet/bubbles/spinner.Model` and a worker goroutine. The action runs on the worker so the bubbletea event loop keeps ticking the spinner. When the action finishes, a `spinnerDoneMsg` flips the step to `Done()=true` and the harness advances.

```go
// SpinnerStep displays a spinner while a blocking action runs on a worker
// goroutine. When the action returns the step completes; Result() is the
// returned error (nil on success). Ctrl-C is handled by the harness — the
// worker keeps running until the underlying call returns, so this does NOT
// fix the pre-existing "Ctrl-C during generate.* leaves partial files"
// issue at the I/O level. What it does fix is the AltScreen exit path:
// cancelling no longer leaves the terminal in spinner-frame state.
type SpinnerStep struct {
    title  string
    label  string                       // text shown next to the spinner
    action func() error                 // the blocking work
    sp     spinner.Model
    err    error
    done   bool
    started bool
}

type spinnerDoneMsg struct{ err error }

// NewSpinner constructs a SpinnerStep.
//   - title: breadcrumb label.
//   - label: text rendered to the right of the spinner glyph.
//   - action: the blocking work; errors are stored in Result().
func NewSpinner(title, label string, action func() error) *SpinnerStep {
    s := spinner.New()
    s.Spinner = spinner.Dot
    s.Style = lipgloss.NewStyle().Foreground(tui.ColorAccent)
    return &SpinnerStep{title: title, label: label, action: action, sp: s}
}

func (s *SpinnerStep) Title() string { return s.title }
func (s *SpinnerStep) Done() bool    { return s.done }
func (s *SpinnerStep) Result() any   { return s.err }

func (s *SpinnerStep) Init() tea.Cmd {
    s.started = true
    return tea.Batch(
        s.sp.Tick,
        func() tea.Msg { return spinnerDoneMsg{err: s.action()} },
    )
}

func (s *SpinnerStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m := msg.(type) {
    case spinnerDoneMsg:
        s.err = m.err
        s.done = true
        return s, nil
    case spinner.TickMsg:
        var cmd tea.Cmd
        s.sp, cmd = s.sp.Update(msg)
        return s, cmd
    }
    return s, nil
}

func (s *SpinnerStep) View() string {
    return "  " + s.sp.View() + " " + tui.StyleMuted.Render(s.label)
}

// Reset is a no-op — once a SpinnerStep has run, popping back to it must NOT
// re-trigger the action (which would re-write files). The harness's Esc-back
// reset loop will skip past it because Reset returns nil; if a user does pop
// back, the step renders its final "completed" frame and the harness should
// not re-Init it. The simplest enforcement: SpinnerStep returns true from
// AutoComplete() once it's done so the Esc-back skip logic walks past.
func (s *SpinnerStep) Reset() tea.Cmd { return nil }

// AutoComplete reports true once the action has finished, so Esc-back skips
// over a completed spinner instead of trying to re-render its (gone) action.
func (s *SpinnerStep) AutoComplete() bool { return s.done }
```

**Dependency promotion:** add `github.com/charmbracelet/bubbles/spinner` import. Already transitive via huh; just becomes direct. `go mod tidy` will move the require line.

**Tests (`steps_test.go`):**
- `TestSpinnerStepCompletesAction` — construct a SpinnerStep with a quick action returning nil; drive `Init()` cmd, then synthesise `spinnerDoneMsg{}` via `Update`, assert `Done()==true` and `Result()==nil`.
- `TestSpinnerStepReportsActionError` — action returns `errors.New("boom")`; assert `Result()` is that error.
- `TestSpinnerStepResetIsNoop` — call `Reset()` after `Done()=true`, assert returned cmd is nil and `Done()` stays true.

---

### Step 2 — `ConditionalStep` adapter

**Modified file:** `internal/tui/harness/steps.go`

Wraps an inner step + a predicate. If the predicate returns false at the time the harness advances onto the step, the wrapped step never renders.

```go
// ConditionalStep wraps another Step with a predicate. When the predicate
// returns false at the moment the harness advances onto this step, the
// wrapped step never renders — Done()=true on Init, AutoComplete()=true so
// Esc-back skips past, and Result()=nil. When the predicate returns true,
// every Step method delegates to the inner step verbatim.
//
// The predicate evaluates against prior results captured at Init time. If
// the user later Esc-backs and changes upstream picks, Reset() re-evaluates
// the predicate so the conditional re-checks correctly.
type ConditionalStep struct {
    inner     Step
    predicate func(prev []any) bool
    skip      bool       // set at Init based on predicate
    skipDone  bool       // flips true once the harness has seen Done()=true once
    initPrev  []any      // prior results captured for the most recent (re-)Init
}

// NewConditional constructs a ConditionalStep.
func NewConditional(inner Step, predicate func(prev []any) bool) *ConditionalStep {
    return &ConditionalStep{inner: inner, predicate: predicate}
}

func (c *ConditionalStep) Title() string { return c.inner.Title() }

func (c *ConditionalStep) Done() bool {
    if c.skip {
        return c.skipDone
    }
    return c.inner.Done()
}

func (c *ConditionalStep) Result() any {
    if c.skip {
        return nil
    }
    return c.inner.Result()
}

// SetPriorResults is called by the harness before Init so the predicate has
// the up-to-date prior-results snapshot. Distinct from a pure Init() arg
// because tea.Model.Init is fixed-signature; the harness invokes this hook
// when present (analogous to how lazyBuilder.Build is invoked).
type priorAware interface {
    SetPrior(prev []any)
}

func (c *ConditionalStep) SetPrior(prev []any) { c.initPrev = prev }

func (c *ConditionalStep) Init() tea.Cmd {
    c.skip = !c.predicate(c.initPrev)
    if c.skip {
        c.skipDone = true
        return nil
    }
    return c.inner.Init()
}

func (c *ConditionalStep) View() string {
    if c.skip {
        return ""
    }
    return c.inner.View()
}

func (c *ConditionalStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if c.skip {
        return c, nil
    }
    updated, cmd := c.inner.Update(msg)
    if step, ok := updated.(Step); ok {
        c.inner = step
    }
    return c, cmd
}

func (c *ConditionalStep) Reset() tea.Cmd {
    c.skip = false
    c.skipDone = false
    if r, ok := c.inner.(resetter); ok {
        return r.Reset()
    }
    return nil
}

func (c *ConditionalStep) AutoComplete() bool {
    if c.skip {
        return true
    }
    type autoChecker interface{ AutoComplete() bool }
    if a, ok := c.inner.(autoChecker); ok {
        return a.AutoComplete()
    }
    return false
}
```

**Harness change in `harness.go`:** the existing advance loop calls `lazyBuilder.Build(prev)` after splice/advance. Add a sibling call for `priorAware`:

```go
// In the advance loop and Init() and post-splice block:
if pa, ok := h.steps[h.cursor].(priorAware); ok {
    pa.SetPrior(h.priorResults())
}
if lb, ok := h.steps[h.cursor].(lazyBuilder); ok && !lb.Built() {
    lb.Build(h.priorResults())
}
if init := h.steps[h.cursor].Init(); init != nil {
    cmd = tea.Batch(cmd, init)
}
```

Apply the `SetPrior` call wherever the existing code calls `Build` — in `Harness.Init`, in the splice expansion path, and in the advance loop in `Harness.Update`. Three callsites total.

**Tests (`steps_test.go` + `harness_test.go`):**
- `TestConditionalStepSkipsWhenPredicateFalse` — wrap a fakeStep, predicate returns false, advance over it, assert inner was never `Init`'d (use a sentinel field on fakeStep). Cursor advances past in one step.
- `TestConditionalStepDelegatesWhenPredicateTrue` — predicate returns true; assert inner.Init was called and View/Update delegate correctly.
- `TestConditionalStepResetReevaluates` — predicate that flips based on a captured slice; flip the slice between the first Init and a Reset+Init cycle, assert skip changes.

---

### Step 3 — Harness `WindowSizeMsg` re-broadcast after splice/lazy entry

**Modified file:** `internal/tui/harness/harness.go`

After `expandSplicer` runs OR after the advance loop builds a `lazyBuilder`, the harness re-sends a `tea.WindowSizeMsg{Width: h.width, Height: h.height}` to the now-active step so its first frame computes layout against the right dimensions instead of waiting for the next user keystroke.

In `Harness.Update`, inside the advance loop after the existing `Init()` call:

```go
// After init cmd batched in:
if h.width > 0 && h.height > 0 {
    updated, sizeCmd := h.steps[h.cursor].Update(tea.WindowSizeMsg{
        Width:  h.width,
        Height: h.height,
    })
    if step, ok := updated.(Step); ok {
        h.steps[h.cursor] = step
    }
    if sizeCmd != nil {
        cmd = tea.Batch(cmd, sizeCmd)
    }
}
```

Same treatment in `Harness.Init` after the initial step's `Init` cmd is computed — but only if `h.width > 0` (which it won't be on the very first Init because we haven't received the first WindowSizeMsg yet). The check guards against a zero-sized broadcast.

**Test (`harness_test.go`):**
- `TestSpliceRebroadcastsWindowSize` — set `h.width=120, h.height=40`, advance through a step that splices in a `fakeStep`, assert the spliced step's recorded `lastSize == (120, 40)` without sending another WindowSizeMsg.

---

### Step 4 — Panic recovery around `Splice()` / `Build()`

**Modified file:** `internal/tui/harness/harness.go`

Wrap the `Splice()` call in `expandSplicer` and the `Build()` call in the advance loop with `defer/recover`. On panic, store the recovered value on the `Harness` and return a `tea.Quit` so the program exits AltScreen cleanly. The post-harness `Run` returns a typed error containing the recovered panic — the caller renders a `tui.FatalPanel` to normal stdout.

```go
// New error type in harness.go:
type BuilderPanicError struct {
    Step  string // step title
    Value any    // recovered value
    Stack string // captured at recovery time via debug.Stack()
}

func (e *BuilderPanicError) Error() string {
    return fmt.Sprintf("harness: builder for step %q panicked: %v", e.Step, e.Value)
}

// New harness fields:
type Harness struct {
    ...
    builderPanic *BuilderPanicError
}

// Helper:
func (h *Harness) recoverBuilder(stepTitle string) {
    if r := recover(); r != nil {
        h.builderPanic = &BuilderPanicError{
            Step:  stepTitle,
            Value: r,
            Stack: string(debug.Stack()),
        }
        h.aborted = false   // not a user abort — distinct flag
        h.quitting = true
    }
}

// In expandSplicer, before calling sp.Splice:
title := h.steps[h.cursor].Title()
defer h.recoverBuilder(title)
inserted := sp.Splice(h.priorResults())
// ... rest of expandSplicer unchanged ...

// Same defer wrapping the lb.Build(...) call sites in Init and Update.
```

`Harness.Update` checks `h.builderPanic != nil` after the splice/build paths and returns `tea.Quit` if set. `Run` checks `model.builderPanic != nil` after `prog.Run()` and returns a `*BuilderPanicError`.

**Caller usage** (in cmd/init.go, cmd/add.go, cmd/remove.go, cmd/update.go):

```go
results, err := harness.Run(banner, action, steps)
if err != nil {
    if errors.Is(err, harness.ErrAborted) {
        return nil
    }
    var bpe *harness.BuilderPanicError
    if errors.As(err, &bpe) {
        tui.FatalPanel("Harness builder panic",
            fmt.Sprintf("Step %q: %v", bpe.Step, bpe.Value),
            "This is a bug — please report it with the trace below.")
        // FatalPanel exits via os.Exit; the line below is unreachable.
        return nil
    }
    return err
}
```

**Test (`harness_test.go`):**
- `TestSpliceBuilderPanicReturnsTypedError` — splice builder panics with `"boom"`, drive Update once, assert `h.quitting==true` and `h.builderPanic.Value=="boom"`. Drive `Run` shape (using a stub `tea.Program`) to confirm `Run` returns `*BuilderPanicError` — or test the post-Update state directly without Run if mocking the program is intractable.
- `TestLazyBuilderPanicReturnsTypedError` — same but for `lazyBuilder.Build`.

---

### Step 5 — `splicer` interface docstring update + nested-lazy unit test

**Modified file:** `internal/tui/harness/harness.go`

Extend the existing `splicer` doc-comment with an explicit limitation note:

```go
// splicer is implemented by LazyGroup. ...
//
// Limitations:
//   - Nested splicers are NOT supported. If Splice() returns a slice that
//     itself contains another splicer at the cursor position, the inner
//     splicer's Splice() will not run automatically — the harness only calls
//     expandSplicer() once per advance. Either flatten the splice to a
//     single level, or build the splice eagerly in the outer builder.
type splicer interface { ... }
```

**New test in `harness_test.go`:**
- `TestLazyStepInsideLazyGroupBuilds` — a LazyGroup that splices in `[fakeStep, NewLazy("inner", build)]`. Drive past the first fakeStep onto the LazyStep, assert the inner step's `build` closure was invoked exactly once, and that prior results passed in include the first fakeStep's result.

---

### Step 6 — `workspaceUniqueValidator` + `normaliseWorkspace` use `filepath.Clean`

**Modified file:** `cmd/add.go`

Update both helpers so the comparison key handles relative-path variants (`./backend`, `backend`, `backend/`) consistently:

```go
func workspaceUniqueValidator(existing map[string]bool) func(string) error {
    return func(s string) error {
        v := strings.TrimSpace(s)
        if v == "" {
            return nil
        }
        v = strings.TrimRight(filepath.Clean(v), "/") + "/"
        if existing[v] {
            return fmt.Errorf("workspace %q is already in use", v)
        }
        return nil
    }
}

func normaliseWorkspace(s string) string {
    v := strings.TrimSpace(s)
    return strings.TrimRight(filepath.Clean(v), "/") + "/"
}
```

The `existing` map built pre-harness (currently from `agent.Workspace` values) must apply the same normalization when populated:

```go
// Where existingWorkspaces is built (currently in runAdd before the harness):
for _, a := range cfg.Agents {
    key := strings.TrimRight(filepath.Clean(a.Workspace), "/") + "/"
    existingWorkspaces[key] = true
}
```

**Tests (`cmd/add_test.go` if it exists, otherwise add a small `internal/tui/harness/...` style test file in `cmd/`):**

Skip new tests if `cmd/` has no test convention today — verify via the manual smoke section that `./backend` and `backend/` collide. (Search for any existing `cmd/*_test.go` first; if absent, document the gap in the iter-3 completion report.)

---

### Step 7 — Migrate `cmd/remove.go` (agent-level `runRemove`)

**Modified file:** `cmd/remove.go`

Replace the body of `runRemove` (current lines 39–135) so the interactive portion runs through the harness.

**Pre-harness:**
1. `cwd := mustCwd()`, `cfg`, `cat`, `agent` lookup, tech-lead-in-use guard. Unchanged.
2. Build `agentDisplayName`, `preview` tree. Same as today.

**In-harness step list:**

```go
steps := []harness.Step{
    harness.NewReview("Confirm removal",
        tui.TitledPanelString("Remove", preview, tui.Amber),
        "Remove "+agentDisplayName+"?",
        false), // default No

    // Spinner runs only if the user confirmed Yes.
    harness.NewConditional(
        harness.NewSpinner("Removing", "Removing agent...", func() error {
            wsPrefix := agent.Workspace
            for relPath := range lock.Files {
                if strings.HasPrefix(relPath, wsPrefix) {
                    lock.Untrack(relPath)
                }
            }
            delete(cfg.Agents, agentName)
            _ = cfg.Save(configPath)
            _ = generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false)
            return nil
        }),
        func(prev []any) bool { return asBool(prev[0]) }, // confirmed
    ),

    // Conflict picker — splice in only if Yes + conflicts exist.
    harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
        if !asBool(prev[0]) {
            return nil
        }
        spinErr := prev[1] // SpinnerStep result is error
        _ = spinErr
        if !wr.HasConflicts() {
            return nil
        }
        return buildConflictSteps(&wr) // see Step 9 — shared helper
    }),
}
```

`lock` and `wr` are declared **before** the steps slice so the spinner closure and the conflict-picker LazyGroup can both close over them:

```go
lock, _ := config.LoadLockFile(cwd)
var wr generate.WriteResult
```

**Post-harness:**
1. Handle `ErrAborted` (user pressed Ctrl-C anywhere) and `BuilderPanicError` per Step 4.
2. If `!asBool(results[0])` (user declined the confirm), return `nil` with no writes.
3. The conflict-picker LazyGroup, if it spliced in steps, has already called `wr.ForceSelected(...)` via the shared helper (Step 9). All harness-driven mutation is done.
4. Save lock: `if err := lock.Save(cwd); err != nil { tui.Warning(...) }`
5. `--delete-files` handling (current lines 116–130). Unchanged.
6. `tui.Success("Removed " + agentDisplayName)` + `tui.Blank()`. Unchanged.

The harness banner is `BONSAI v...`, action is `"Removing agent"`.

---

### Step 8 — Migrate `cmd/remove.go` (`runRemoveItem`)

**Modified file:** `cmd/remove.go`

Replace the body of `runRemoveItem` (current lines 197–361). The pre-harness portion stays; only the interactive section changes.

**Pre-harness (unchanged):**
1. Block auto-managed sensors (lines 199–202).
2. `cwd`, `cfg`, `cat`. Find `matches []agentMatch`. If empty → `ErrorDetail` + return.
3. Filter required (lines 261–275). If `allowed` empty → `ErrorDetail` + return.
4. Build `displayName`, `fromLabels`, `content`.

**In-harness step list:**

```go
// Pre-compute whether we need an agent picker.
needsPicker := len(matches) > 1
agentOptions := buildAgentOptions(matches, cat) // nil if !needsPicker

steps := []harness.Step{
    // Step 0: optional agent picker (nil-skip via Conditional? No — easier to
    // just include the SelectStep unconditionally and have a synthetic
    // single-option auto-pick OR conditionally include the step. Use
    // ConditionalStep wrapping a SelectStep so a single-match flow has
    // results[0]=nil and the targets-builder logic in step 1 handles both.)
    harness.NewConditional(
        harness.NewSelect("Agent", "Remove from which agent?", agentOptions),
        func(prev []any) bool { return needsPicker },
    ),

    // Step 1: confirm summary panel.
    harness.NewLazy("Confirm removal", func(prev []any) harness.Step {
        targets := resolveTargets(prev[0], matches) // see helper below
        // Re-render content with targets-specific From line:
        // (same shape as the existing tui.CardFields call)
        panel := tui.TitledPanelString("Remove Item",
            buildItemSummary(displayName, it, targets), tui.Amber)
        return harness.NewReview("Confirm removal", panel, "Remove "+displayName+"?", false)
    }),

    // Step 2: spinner — gated by confirm.
    harness.NewConditional(
        harness.NewSpinner("Removing", "Removing "+it.singular+"...", func() error {
            // Re-resolve targets (closure capture). Action body == current
            // body of the spinner.New().Action(...) at lines 302-347 — minus
            // the wrapping spinner machinery.
            return runRemoveItemAction(cwd, cfg, cat, lock, &wr, configPath, name, it, capturedTargets)
        }),
        func(prev []any) bool { return asBool(prev[1]) }, // results[1] is the Review confirm bool
    ),

    // Step 3: conflict picker — gated by confirm + conflicts existing.
    harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
        if !asBool(prev[1]) || !wr.HasConflicts() {
            return nil
        }
        return buildConflictSteps(&wr)
    }),
}
```

**Helpers added in `cmd/remove.go`:**

```go
// buildAgentOptions builds the agent-picker options from matches.
func buildAgentOptions(matches []agentMatch, cat *catalog.Catalog) []huh.Option[string] {
    var options []huh.Option[string]
    for _, m := range matches {
        label := agentDisplayName(cat, m.name)
        options = append(options,
            huh.NewOption(label+" "+tui.StyleMuted.Render(tui.GlyphArrow+" "+m.agent.Workspace), m.name))
    }
    options = append(options, huh.NewOption("All agents", "_all_"))
    return options
}

// resolveTargets converts the agent-picker result + matches into the target
// slice. Handles: nil (single match — auto-pick), "_all_", or a specific name.
func resolveTargets(picked any, matches []agentMatch) []agentMatch {
    if picked == nil {
        return matches // single-match path; ConditionalStep skipped the picker
    }
    selected := asString(picked)
    if selected == "_all_" || len(matches) == 1 {
        return matches
    }
    for _, m := range matches {
        if m.name == selected {
            return []agentMatch{m}
        }
    }
    return matches
}

// runRemoveItemAction is the body of the current spinner.Action closure,
// extracted so it's callable from a SpinnerStep closure (which needs an
// `error`-returning function rather than a bare `func()`).
func runRemoveItemAction(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog,
    lock *config.LockFile, wr *generate.WriteResult, configPath, name string,
    it itemType, targets []agentMatch) error {
    for _, t := range targets {
        // ... body of current lines 304-343, with `return err` paths plumbed
        //     through (today they swallow with `_ =`; preserve that swallow
        //     for parity, and just return nil at the end — the plan does NOT
        //     introduce error handling here, that's a separate Backlog item).
    }
    _ = cfg.Save(configPath)
    _ = generate.SettingsJSON(cwd, cfg, cat, lock, wr, false)
    return nil
}

// buildItemSummary returns the static summary content for the review panel.
func buildItemSummary(displayName string, it itemType, targets []agentMatch) string {
    var fromLabels []string
    for _, t := range targets {
        fromLabels = append(fromLabels, t.name+" ("+t.agent.Workspace+")")
    }
    return tui.CardFields([][2]string{
        {"Item", displayName},
        {"Type", catalog.DisplayNameFrom(it.singular)},
        {"From", strings.Join(fromLabels, ", ")},
    })
}
```

**Targets capture across closures:** the spinner action needs the resolved `targets`, which are computed from `prev[0]` in the LazyStep at step 1. The cleanest plumbing: have the LazyStep's `Result()` continue to be the confirm bool (so step 2/3 predicates work), and capture targets via a heap-allocated `*[]agentMatch` pointer set by the LazyStep build closure:

```go
var capturedTargets []agentMatch

// Step 1 (LazyStep build):
harness.NewLazy("Confirm removal", func(prev []any) harness.Step {
    capturedTargets = resolveTargets(prev[0], matches)
    panel := ...
    return harness.NewReview(...)
}),

// Step 2 (SpinnerStep closure):
harness.NewSpinner("Removing", ..., func() error {
    return runRemoveItemAction(..., capturedTargets)
}),
```

The closure captures `&capturedTargets` by reference, so by the time the spinner runs (step 2), the LazyStep build at step 1 has already populated it.

**Post-harness:**
1. `ErrAborted` / `BuilderPanicError` handling.
2. If `!asBool(results[1])` → return nil.
3. `lock.Save(cwd)` with warning on error.
4. `tui.Success("Removed " + displayName)` + `tui.Blank()`.

---

### Step 9 — Shared `buildConflictSteps` helper

**Modified file:** `cmd/root.go`

Add a new helper alongside `resolveConflicts` (which stays for backward compat / tests, but the four migrated commands no longer call it). Returns the 1-or-2 harness steps that drive the conflict picker.

```go
// buildConflictSteps returns the harness steps for the conflict-resolution
// picker. Caller is responsible for actually applying the picks via
// applyConflictPicks (called from a post-harness block) — the harness only
// captures the picks, since wr.ForceSelected mutates the lock state.
//
// Returns nil when wr.HasConflicts() is false so callers can splice in
// nothing without a wrapper conditional.
func buildConflictSteps(wr *generate.WriteResult) []harness.Step {
    conflicts := wr.Conflicts()
    if len(conflicts) == 0 {
        return nil
    }

    // Build options, all pre-selected for update.
    available := make([]tui.ItemOption, 0, len(conflicts))
    defaults := make([]string, 0, len(conflicts))
    for _, c := range conflicts {
        available = append(available, tui.ItemOption{
            Name:  c.RelPath,
            Value: c.RelPath,
            Desc:  "modified since last generate",
        })
        defaults = append(defaults, c.RelPath)
    }

    return []harness.Step{
        harness.NewMultiSelect("Conflicts",
            fmt.Sprintf("%d file(s) modified since Bonsai generated them. Select which to update — unchecked files keep your changes.", len(conflicts)),
            available, defaults),
        harness.NewConditional(
            harness.NewConfirm("Backup", "Create .bak backups before overwriting?", false),
            func(prev []any) bool {
                // Predicate evaluates with the conflict picker as the most
                // recent prior result. Only ask about backups if the user
                // selected at least one file to overwrite.
                if len(prev) == 0 {
                    return false
                }
                picks := asStringSlice(prev[len(prev)-1])
                return len(picks) > 0
            },
        ),
    }
}

// applyConflictPicks consumes the harness results from buildConflictSteps
// (the trailing two slots in the results slice) and runs the file mutations
// resolveConflicts used to do inline. confIdx is the index of the
// MultiSelectStep result in the results slice; backupIdx is confIdx+1.
//
// Returns true if writes occurred, false otherwise.
func applyConflictPicks(results []any, confIdx int, wr *generate.WriteResult,
    lock *config.LockFile, projectRoot string) bool {
    if confIdx < 0 || confIdx >= len(results) {
        return false
    }
    selected := asStringSlice(results[confIdx])
    if len(selected) == 0 {
        return false
    }
    backupIdx := confIdx + 1
    backup := backupIdx < len(results) && asBool(results[backupIdx])
    if backup {
        for _, relPath := range selected {
            abs := filepath.Join(projectRoot, relPath)
            data, readErr := os.ReadFile(abs)
            if readErr == nil {
                _ = os.WriteFile(abs+".bak", data, 0644)
            }
        }
    }
    wr.ForceSelected(selected, projectRoot, lock)
    return true
}
```

The MultiSelect uses the existing `MultiSelectStep` (the conflict list has no required items, so all are optional with `defaults=allConflicts` for the pre-selected effect — the existing `Selected(true)` mechanic).

`asStringSlice` and `asBool` are currently defined in `cmd/init.go` as package-level helpers (lines 88–119). They're already package-scoped (`asString`, `asStringSlice`, `asBool`) so no relocation needed.

**Caller usage in init/add/remove/update:**

```go
// At the end of the post-harness block, BEFORE showWriteResults:
if applyConflictPicks(results, conflictIdx, &wr, lock, cwd) {
    // applied — wr now has updated state
}
```

The caller knows `conflictIdx` because they constructed the step list. For predictable layout, the conflict steps are always the last two indices when present (after the LazyGroup splice expansion). The `applyConflictPicks` helper tolerates the slot being absent (LazyGroup spliced in nothing) by returning false.

---

### Step 10 — Migrate `cmd/update.go`

**Modified file:** `cmd/update.go`

Replace the body of `runUpdate` (lines 28–203).

**Pre-harness (unchanged):**
1. `cwd`, `cfg`, `cat`, `lock`. Sort `agentNames`.
2. Pre-flight: scan all agents for discovered files. Bail out if every agent has zero discoveries — in that case skip the harness entirely and go directly to the spinner phase (no input needed).

**Pre-flight scan:**

```go
// Map: agentName -> []DiscoveredFile (only valid ones; warnings printed inline).
discoveredByAgent := make(map[string][]generate.DiscoveredFile)
hasAnyDiscoveries := false
for _, agentName := range agentNames {
    installed := cfg.Agents[agentName]
    discovered, scanErr := generate.ScanCustomFiles(cwd, installed, lock)
    if scanErr != nil || len(discovered) == 0 {
        continue
    }
    var valid, invalid []generate.DiscoveredFile
    for _, d := range discovered {
        if d.Error != "" {
            invalid = append(invalid, d)
        } else {
            valid = append(valid, d)
        }
    }
    // Warnings stay on stdout (pre-harness), exactly as today.
    for _, d := range invalid {
        tui.Warning(fmt.Sprintf("Skipping %s: %s", d.RelPath, d.Error))
        tui.Hint("Add frontmatter to track this file. See docs/custom-files.md for format.")
    }
    if len(valid) > 0 {
        discoveredByAgent[agentName] = valid
        hasAnyDiscoveries = true
    }
}
```

**In-harness step list:**

```go
configChanged := false
var wr generate.WriteResult

steps := []harness.Step{
    // One MultiSelectStep per agent that has discoveries, each gated by
    // ConditionalStep so the harness skips agents with no discoveries
    // entirely. Built via LazyGroup so the order is deterministic and the
    // declaration site is compact.
    harness.NewLazyGroup("Custom files", func(prev []any) []harness.Step {
        var out []harness.Step
        for _, agentName := range agentNames {
            valid := discoveredByAgent[agentName]
            if len(valid) == 0 {
                continue // dropped pre-flight; nothing to add
            }
            agentDef := cat.GetAgent(cfg.Agents[agentName].AgentType)
            agentLabel := agentName
            if agentDef != nil {
                agentLabel = agentDef.DisplayName
            }
            options := buildCustomFileOptions(valid)
            defaults := buildCustomFileDefaults(valid)
            out = append(out, harness.NewMultiSelect(
                "Custom files — "+agentLabel,
                fmt.Sprintf("Custom files found — %s", agentLabel),
                options, defaults))
        }
        return out
    }),

    // Spinner — runs unconditionally (re-renders abilities/CLAUDE.md/settings).
    harness.NewSpinner("Syncing", "Syncing workspace...", func() error {
        // Apply selections from the previous steps. Selections are read from
        // the results slice in the post-harness block; here we just re-render.
        for _, agentName := range agentNames {
            installed := cfg.Agents[agentName]
            agentDef := cat.GetAgent(installed.AgentType)
            if agentDef == nil {
                continue
            }
            generate.EnsureRoutineCheckSensor(installed)
            _ = generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false)
        }
        _ = generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false)
        _ = generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false)
        _ = generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false)
        return nil
    }),

    harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
        if !wr.HasConflicts() {
            return nil
        }
        return buildConflictSteps(&wr)
    }),
}
```

**Custom-file selection application** — must run BEFORE the spinner step so the spinner's `generate.AgentWorkspace` sees the updated config. Two options:

**Option A (chosen):** apply selections inside the spinner closure by reading from a shared `selectionsByAgent` map populated by the LazyGroup builder closure. The LazyGroup builder doesn't know the user's picks (those come from the MultiSelectStep results post-completion), so this needs a different shape.

**Option B:** put the per-agent picker steps and the spinner in *one* tea.Program with the spinner closure reading `prev` results. SpinnerStep doesn't see prior results today — it only sees `Init()`'s `tea.Cmd`. Add the same `priorAware` hook used by ConditionalStep so SpinnerStep can capture prior results at Init time:

```go
// In SpinnerStep:
type SpinnerStep struct {
    ...
    initPrev []any
}
func (s *SpinnerStep) SetPrior(prev []any) { s.initPrev = prev }
func (s *SpinnerStep) Init() tea.Cmd {
    s.started = true
    return tea.Batch(
        s.sp.Tick,
        func() tea.Msg { return spinnerDoneMsg{err: s.action()} }, // action closure can read s.initPrev via closure
    )
}
```

But `action` is captured at construction — it doesn't see `s.initPrev`. To wire prior results into the action, the SpinnerStep needs an action signature variant:

```go
// NewSpinner accepts a func() error. NewSpinnerWithPrior accepts a func([]any) error.
type SpinnerStep struct {
    title    string
    label    string
    action   func() error
    actionP  func(prev []any) error // alternative; one of action/actionP is nil
    ...
}
func NewSpinnerWithPrior(title, label string, action func(prev []any) error) *SpinnerStep {
    s := newSpinnerCommon(title, label)
    s.actionP = action
    return s
}
func (s *SpinnerStep) Init() tea.Cmd {
    s.started = true
    runner := s.action
    if s.actionP != nil {
        prev := s.initPrev
        runner = func() error { return s.actionP(prev) }
    }
    return tea.Batch(s.sp.Tick, func() tea.Msg { return spinnerDoneMsg{err: runner()} })
}
```

Then `cmd/update.go` uses `NewSpinnerWithPrior` so the action body can read the per-agent pick results from `prev`. Order: prev[0] is the LazyGroup placeholder (returns nil after splice — safe to ignore), and prev[1..N] are the per-agent MultiSelect results in declaration order (matches `agentNames` filtered by `len(valid) > 0`).

**Note on prev indexing post-splice:** when a `LazyGroup` splices in N steps, `priorResults()` collects the results of the spliced steps (since they completed before the cursor advanced past). The original LazyGroup's `Result()=nil` doesn't appear in `priorResults` because by then the LazyGroup is gone from `h.steps`. So for the spinner step (the next step after the splice), `prev` has `len = number_of_spliced_steps`. Document this in the SpinnerStep prior-aware path.

**Action body** uses prev to apply selections:

```go
harness.NewSpinnerWithPrior("Syncing", "Syncing workspace...", func(prev []any) error {
    // Apply per-agent selections.
    cursor := 0
    for _, agentName := range agentNames {
        valid := discoveredByAgent[agentName]
        if len(valid) == 0 {
            continue
        }
        if cursor >= len(prev) {
            break
        }
        selected := asStringSlice(prev[cursor])
        cursor++
        applyCustomFileSelection(cfg.Agents[agentName], valid, selected, lock, cwd, &configChanged)
    }
    // Re-render (same as before).
    for _, agentName := range agentNames {
        installed := cfg.Agents[agentName]
        agentDef := cat.GetAgent(installed.AgentType)
        if agentDef == nil {
            continue
        }
        generate.EnsureRoutineCheckSensor(installed)
        _ = generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false)
    }
    _ = generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false)
    _ = generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false)
    _ = generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false)
    return nil
}),
```

`applyCustomFileSelection` is a small helper extracting the selection-apply loop currently at `cmd/update.go:107-143`.

**Post-harness:**

```go
// Apply conflict picks if any.
conflictBaseIdx := /* compute based on prev layout */
applyConflictPicks(results, conflictBaseIdx, &wr, lock, cwd)

// Save config + lock.
if configChanged {
    if err := cfg.Save(configPath); err != nil {
        tui.Warning("Could not save config: " + err.Error())
    }
}
if err := lock.Save(cwd); err != nil {
    tui.Warning("Could not save lock file: " + err.Error())
}

created, updated, _, _, conflicts := wr.Summary()
hadChanges := configChanged || created > 0 || updated > 0 || conflicts > 0
if !hadChanges {
    tui.TitledPanel("Up to date",
        "Workspace is in sync with the catalog.\nNo files needed updating.",
        tui.Moss)
    tui.Blank()
    return nil
}

showWriteResults(&wr, ".")
if configChanged {
    tui.Success("Update complete — custom files tracked")
} else {
    tui.Success("Update complete — workspace synced")
}
tui.Hint("Review changes with: bonsai list")
tui.Blank()
return nil
```

**Banner:** `BONSAI v...`, action `"Updating workspace"`.

**Pre-harness short-circuit:** if `!hasAnyDiscoveries` AND no other interactive surface needs the user, the simplest path is still to enter the harness so the spinner runs inside AltScreen. The LazyGroup's empty splice + the spinner step + the conflict LazyGroup is exactly 1 spinner step + maybe a conflict picker — totally appropriate for AltScreen. Don't add a pre-harness escape hatch.

---

### Step 11 — Retro-fit `cmd/init.go` and `cmd/add.go` to use SpinnerStep + conflict-picker LazyGroup

**Modified files:** `cmd/init.go`, `cmd/add.go`

The two iter-2 commands currently run the `huh/spinner` block + `resolveConflicts` block in a *post-harness* tail (cmd/init.go:213-239, cmd/add.go ~similar shape). Both windows are vulnerable to Ctrl-C-during-generate leaving partial files.

For each command, **append** the same two steps to the existing step list (after the review step):

```go
// init.go runInit step list — append:
harness.NewConditional(
    harness.NewSpinnerWithPrior("Generating", "Generating project files...",
        func(prev []any) error {
            // No prior-results dependency for init; ignore prev. Body lifted
            // verbatim from current spinner.Action closure (cmd/init.go:218-224).
            _ = generate.Scaffolding(cwd, cfg, cat, lock, &wr, false)
            _ = generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false)
            _ = generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false)
            _ = generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false)
            _ = generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false)
            return nil
        }),
    func(prev []any) bool {
        // Review step is the previous index; gate on its confirm bool.
        return asBool(prev[len(prev)-1])
    },
),
harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
    // Find the review confirm result. It's the last result before this group.
    if len(prev) == 0 {
        return nil
    }
    if !wr.HasConflicts() {
        return nil
    }
    return buildConflictSteps(&wr)
}),
```

`cfg`, `installed`, `lock`, `&wr` are constructed in the post-harness block today; with the spinner moving in, they need to be constructed BEFORE `harness.Run`. Restructure runInit's post-harness validation/build block (lines 176-211) to run **between** the input-collection harness call and the spinner, OR move the build into the SpinnerStep closure itself.

**Chosen approach:** move config/installed/lock construction into the SpinnerStep closure. The closure runs after the user confirms the review (gated by ConditionalStep on the review bool). All inputs needed (`projectName`, `description`, `docsPath`, etc.) are accessible via captured `prev` results — but the closure receives `prev` only when the spinner runs, which is fine. Refactor `runInit` so:

```go
func runInit(...) error {
    // Pre-harness: cwd, configPath, exists check, cat, agentDef. Same as today.

    // Declare lock + wr + cfg + installed up-front so closures can write.
    lock, _ := config.LoadLockFile(cwd)
    var wr generate.WriteResult
    var cfg *config.ProjectConfig
    var installed *config.InstalledAgent

    steps := []harness.Step{
        // ... existing 9 input steps ...
        // ... existing review step (LazyStep) ...
        harness.NewConditional(
            harness.NewSpinnerWithPrior("Generating", "Generating project files...",
                func(prev []any) error {
                    // Build cfg + installed from prev results.
                    projectName := asString(prev[0])
                    description := asString(prev[1])
                    docsPath := normaliseDocsPath(asString(prev[2]))
                    selectedScaffolding := asStringSlice(prev[3])
                    selectedSkills := asStringSlice(prev[4])
                    selectedWorkflows := asStringSlice(prev[5])
                    selectedProtocols := asStringSlice(prev[6])
                    selectedSensors := asStringSlice(prev[7])
                    selectedRoutines := asStringSlice(prev[8])

                    installed = &config.InstalledAgent{
                        AgentType: techLeadType,
                        Workspace: docsPath,
                        Skills:    selectedSkills,
                        Workflows: selectedWorkflows,
                        Protocols: selectedProtocols,
                        Sensors:   selectedSensors,
                        Routines:  selectedRoutines,
                    }
                    generate.EnsureRoutineCheckSensor(installed)
                    cfg = &config.ProjectConfig{
                        ProjectName: strings.TrimSpace(projectName),
                        Description: strings.TrimSpace(description),
                        DocsPath:    docsPath,
                        Scaffolding: selectedScaffolding,
                        Agents:      map[string]*config.InstalledAgent{techLeadType: installed},
                    }
                    if err := cfg.Save(configPath); err != nil {
                        return err
                    }
                    _ = generate.Scaffolding(cwd, cfg, cat, lock, &wr, false)
                    _ = generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false)
                    _ = generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false)
                    _ = generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false)
                    _ = generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false)
                    return nil
                }),
            func(prev []any) bool { return asBool(prev[len(prev)-1]) },
        ),
        harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
            if !wr.HasConflicts() {
                return nil
            }
            return buildConflictSteps(&wr)
        }),
    }

    results, err := harness.Run(...)
    if err != nil { ... }

    // Post-harness:
    // - Did the user confirm? (review result is at index 9)
    if len(results) <= 9 || !asBool(results[9]) {
        return nil
    }
    // - Surface spinner error (results[10] is spinner result if it ran).
    if len(results) > 10 {
        if errVal := results[10]; errVal != nil {
            if e, ok := errVal.(error); ok && e != nil {
                tui.Warning("Generation error: " + e.Error())
            }
        }
    }
    // - Apply conflict picks if any.
    conflictIdx := indexOfConflictPicker(results) // first slot after spinner that's a []string
    applyConflictPicks(results, conflictIdx, &wr, lock, cwd)
    // - Save lock, showWriteResults, Success, Hint. Same as today.
    ...
}
```

`indexOfConflictPicker` walks the results slice from the spinner index forward looking for the first `[]string` result (the conflict MultiSelect's output). Returns -1 if no conflict steps spliced in.

**Same restructure for `cmd/add.go`** — both branches (`runAdd`/`runAddItems`) get spinner + conflict picker appended. The shape is identical; the action closures lift from the existing post-harness pipeline.

**Important:** the existing `cfg.Save` for `bonsai add` happens after the harness today. With the spinner inside the harness, `cfg.Save` moves into the spinner closure. The "all installed" short-circuit path in iter 2.1 never reached cfg.Save (no changes); ConditionalStep on the spinner preserves that behavior because the predicate gates on review confirm — and the all-installed splice (Step 4 in iter 2.1) leads directly to the splice ending without a Review confirm at all (since the splice returned `[NoteStep("All installed")]` — actually re-check the iter 2.1 path here).

**Re-check iter 2.1 path:** In iter 2.1, the all-installed branch returns either a NoteStep or an empty slice. If it returns the NoteStep, the user presses Enter and the harness exits with no review confirm in results. If it returns empty, the splice produces zero steps and the harness flows directly to the next step (which would be the new spinner appended in step 11).

For the all-installed empty-splice case to NOT trigger the spinner: the predicate on the ConditionalStep wrapping the spinner must check whether a review confirm exists. The simplest predicate:

```go
func(prev []any) bool {
    if len(prev) == 0 {
        return false
    }
    last := prev[len(prev)-1]
    // Review confirm produces a bool. Anything else (nil, []string, string) means
    // no confirm step ran, so don't spin.
    b, ok := last.(bool)
    return ok && b
}
```

This handles both the all-installed-empty-splice path (last result is whatever preceded; not a bool, predicate false) AND the all-installed-NoteStep path (last result is nil from NoteStep; not a bool, predicate false) AND the normal review-confirmed path (last result is true).

---

### Step 12 — Tests

**New tests in `internal/tui/harness/steps_test.go`:**
- 3 SpinnerStep tests (Step 1).
- 3 ConditionalStep tests (Step 2).

**New tests in `internal/tui/harness/harness_test.go`:**
- 1 WindowSize re-broadcast test (Step 3).
- 2 panic-recovery tests (Step 4).
- 1 LazyStep-inside-LazyGroup test (Step 5).

**Total new tests:** 10. All reducer-only; no TTY needed; same pattern as iter 1/2.

**Existing tests to update:**
- `TestEscPopReinitsActiveStep` and similar may need a small adjustment if the WindowSize re-broadcast in Step 3 changes the message ordering they assert on. Re-run after Step 3 lands and patch as needed.

---

### Iter 3 — Verification

#### Build & Test

- [ ] `make build` — clean compile.
- [ ] `go test ./...` — all tests green, incl. 10 new harness tests.
- [ ] `gofmt -s -l .` — no output.
- [ ] `go vet ./...` — no issues.
- [ ] `go mod tidy` — `bubbles/spinner` promoted to direct dep cleanly.

> **Manual smoke sections below are deferred to the whole-branch merge audit before `ui-ux-testing → main`.** No PTY in dispatched environment; tech lead to walk the flows locally at iter-3 completion.

#### Manual — `bonsai remove <agent>` flow

In a temp project with multiple agents installed:

- [ ] AltScreen activates; scrollback before the command preserved on exit.
- [ ] Header shows `BONSAI v...` left, `Removing agent` middle, `[1/3] Confirm removal` right on step 1.
- [ ] Review panel renders with bordered TitledPanel look.
- [ ] Confirm No → harness exits, no writes.
- [ ] Confirm Yes → spinner appears (visible spinner glyph ticking), then either: (a) clean exit + Success, OR (b) conflict picker appears, complete it, then Success.
- [ ] Ctrl-C during input phase → clean AltScreen exit, no writes.
- [ ] Ctrl-C during spinner → AltScreen exits cleanly (note: file write may be partially complete — that's a separate atomicity issue tracked elsewhere; the harness exit path itself is clean).
- [ ] `--delete-files` flag still triggers post-success cleanup of agent dir.

#### Manual — `bonsai remove skill <name>` flow

- [ ] Single-match agent: agent picker step is silently skipped (no flash of an empty Select); harness goes directly to Confirm.
- [ ] Multi-match: agent picker shows; choosing "All agents" routes through correctly.
- [ ] Required-item case: pre-harness ErrorDetail prints; harness never enters AltScreen.

#### Manual — `bonsai update` flow

- [ ] No discoveries anywhere: harness shows just the spinner (and optional conflict picker), no empty MultiSelect screens.
- [ ] One agent has discoveries: one MultiSelect step renders; others skipped.
- [ ] Multiple agents have discoveries: per-agent MultiSelects render in sorted order.
- [ ] Selections apply correctly — `cfg.Save` happens inside the spinner.
- [ ] Conflict picker appears only when conflicts exist.
- [ ] "Up to date" panel still renders post-harness for no-op case.

#### Manual — `bonsai init` and `bonsai add` regression

- [ ] AltScreen now wraps the entire flow including the spinner and conflict picker.
- [ ] Ctrl-C during the spinner exits AltScreen cleanly.
- [ ] All iter-2 / iter-2.1 behaviors preserved (Esc-back, all-installed short-circuit, tech-lead bootstrap NoteStep, etc.).

#### Manual — broader regressions

- [ ] `NO_COLOR=1` for all four commands — zero ANSI escapes.
- [ ] `bonsai list` / `bonsai catalog` — unchanged (still non-interactive).
- [ ] Workspace validator: try to add `./backend/` when `backend/` is already installed — collision detected.

---

## Iter 3 — Outline (superseded by Iter 3 — Detailed Steps above)

- Migrate `cmd/remove.go` (agent-level + per-item subcommands) onto the harness. ✅ planned
- Migrate `cmd/update.go`. ✅ planned
- Move conflict-resolution picker (`resolveConflicts`) into the harness. ✅ planned
- Move `huh/spinner` invocations into a `SpinnerStep` adapter. ✅ planned

---

## Security

> [!warning]
> Refer to [SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

- No new external deps; bubbletea is already transitive via huh/bubbles.
- No user input crosses an exec/shell boundary in the harness — input is passed by value to existing config code.
- AltScreen does not change file I/O surface — same `os.Stat`, `cfg.Save`, `lock.Save` paths.
- Validators run on the same trimmed strings as today; no new injection points.
- No template execution or YAML parsing changes.

---

## Dependencies

- Promote `github.com/charmbracelet/bubbletea` from indirect to direct in `go.mod` (`go mod tidy` will handle).
- No catalog changes.
- No generator changes.
- No changes to `internal/tui/styles.go` color tokens; only adds the three Harness* style vars.

---

## Verification (Iter 1)

### Build & Test

- [ ] `make build` — compiles with no errors or warnings
- [ ] `go test ./...` — all existing + new harness tests pass
- [ ] `gofmt -s -l .` — no formatting issues
- [ ] `go vet ./...` — no issues

### Manual — `bonsai init` flow

Run `mkdir /tmp/bonsai-plan15-iter1 && cd /tmp/bonsai-plan15-iter1 && /path/to/bonsai init`:

- [ ] AltScreen activates — terminal scrollback before the command is preserved when the command exits.
- [ ] Header shows `BONSAI vX.Y.Z` left, `Initializing new project` middle, `[N/M] <step title>` right.
- [ ] Pressing `Esc` on step 2+ returns to the previous step with the prior answer preserved.
- [ ] Pressing `Esc` on step 1 is a no-op (no quit).
- [ ] `Ctrl-C` exits cleanly, no `.bonsai.yaml` or partial workspace written.
- [ ] After the review step, harness exits AltScreen and the spinner / write-result / success banner render to normal stdout exactly as today.
- [ ] Resize terminal mid-flow — header/footer reflow without redraw artefacts.

### Manual — regressions

- [ ] `NO_COLOR=1 bonsai init` — flow still completes, no ANSI escapes in any output.
- [ ] `bonsai init` on a light terminal — colors legible.
- [ ] `bonsai add` — unchanged behavior (still on stateless path).
- [ ] `bonsai remove` — unchanged.
- [ ] `bonsai update` — unchanged.
- [ ] `bonsai list` — unchanged.
- [ ] `bonsai catalog` — unchanged.

---

## Dispatch

| Agent | Isolation | Notes |
|-------|-----------|-------|
| general-purpose | worktree | Go + TUI changes only. No catalog, no generator, no docs. Iter 1 only — iter 2/3 dispatched separately. |

---

## Out of Scope (defer beyond Plan 15)

- Mouse interaction inside the harness (golden rule #3 doesn't apply yet — keyboard-only).
- Streaming long output (e.g. catalog browsing) inside the harness — `bonsai catalog` and `bonsai list` stay non-interactive.
- Rich preview pane / dual-pane layouts — not needed for wizards.
- Replacing the Zen Garden palette or `BonsaiTheme()`.
- Persistent in-flow validation (e.g., live filesystem checks) — validators stay synchronous in iter 1.
