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
| 1 | Harness package + theme + `cmd/init.go` migration | Shipped (ui-ux-testing @ 2d7a947) |
| 2 | Migrate `cmd/add.go` (incl. `runAddItems` pivot) + `NoteStep` + `TitledPanelString` + harness `LazyGroup` splice | Planned |
| 3 | Migrate `cmd/remove.go` + `cmd/update.go` (custom-file scan, conflict picker, spinner step) | Outlined |

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
func (s *NoteStep) Update(msg tea.Msg) (tea.Model, tea.Cmd)    { /* standard form-delegation */ }
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

1. `cwd := mustCwd()`, load config + catalog, early-exit guards (`requireConfig`, unknown agent type guard stays unchanged).
2. Remove the standalone `tui.Heading("Add Agent")` call — the harness header replaces it.
3. The tech-lead pre-check (`if agentType != "tech-lead" && !hasTechLead`) must move to a post-select branch — it depends on the user's selection. Rendering in-harness is awkward; handle it by making the branch return a single `NoteStep` that shows the error, and have the flow terminate after that one note (see Step 4b).

Actually cleaner: keep the tech-lead-required check *inside* the splice so the failure is rendered as a NoteStep with `tui.ErrorDetail`-equivalent copy, and the flow short-circuits by returning a single NoteStep. But ErrorDetail is intrinsically a stdout write. Preferred path: **after `harness.Run` returns**, re-check `agentType == "tech-lead"` vs installed agents and render `ErrorDetail` in normal stdout exactly as today, then return. That means the splicer builds the full `runAdd` flow unconditionally; the tech-lead-required validation happens after harness exit with no state written. (Harness is essentially "gather answers"; execution stays outside.)

**Decision:** Move the tech-lead-required check *after* `harness.Run` returns, before persisting config. User flow: they pick `backend`, walk through the whole add-agent wizard, click Yes on review; harness exits; we then see "no tech-lead installed" and render `ErrorDetail`. Slightly late but only a pathological case (no one runs `bonsai add backend` before `bonsai init`). Acceptable for iter 2. **Alternative** (if objectionable during self-review): short-circuit pre-harness via an agent-catalog-independent check — read `cfg.Agents` right after `requireConfig` and refuse to offer non-tech-lead options when no tech-lead is installed. Cleaner but changes the option list.

**In-harness step list:**

```go
steps := []harness.Step{
    harness.NewSelect("Agent", "Agent type:", agentOptions),
    harness.NewLazyGroup("Agent flow", func(prev []any) []harness.Step {
        agentType := asString(prev[0])
        if _, exists := cfg.Agents[agentType]; exists {
            return buildAddItemsSteps(cat, cfg, agentType) // runAddItems branch
        }
        return buildNewAgentSteps(cat, cfg, agentType)     // runAdd branch
    }),
}
```

**`buildNewAgentSteps(cat, cfg, agentType)` (runAdd branch):**

1. **Workspace step** — built via `harness.NewLazy` (single-step lazy) so it can branch:
   - If `agentType == "tech-lead"`: `NoteStep` with body "Tech Lead workspace: {cfg.DocsPath or 'station/'}".
   - Else: `TextStep` with prompt "Workspace directory (e.g. backend/):", default `agentType + "/"`, required, plus a validator that rejects any path already in `existingWorkspaces[]` (reject in-place so the user can correct without crashing after the form).
2. `MultiSelectStep` × 5 — Skills, Workflows, Protocols, Sensors (filtered to exclude `routine-check`), Routines.
3. `LazyStep` — Review: builds a `ReviewStep` using `tui.TitledPanelString("Review", tree, tui.Water)`.

**`buildAddItemsSteps(cat, cfg, agentType)` (runAddItems branch):**

1. `NoteStep` — "{agentDef.DisplayName} is already installed at {installed.Workspace} — showing uninstalled abilities."
2. `MultiSelectStep` for each of Skills/Workflows/Protocols/Sensors/Routines *only if the filtered list is non-empty* — zero-item categories are omitted from the step list entirely (clean, avoids auto-complete headers with nothing to show).
3. `LazyStep` — Review: builds a `ReviewStep` titled "Adding" with the tree of newly-selected items.

**Pre-harness short-circuit for runAddItems:** after computing filter counts in `buildAddItemsSteps`, if `totalAvailable == 0`, we can't short-circuit inside the splicer (no way to cancel the flow cleanly from there). Instead, do this check **before** entering the harness: before constructing the step list, peek at whether the agent exists AND whether it has any uninstalled abilities; if both true and count is 0, render `EmptyPanel(...)` in normal stdout and return.

This means the pre-harness pipeline in `runAdd` looks like:

```go
func runAdd(cmd *cobra.Command, args []string) error {
    cwd, cfg, cat, configPath, err := setupAdd()  // existing guards
    if err != nil { return err }

    // NOTE: agent-type selection is inside the harness now; we can't
    // short-circuit empty add-items pre-harness without asking the user first.
    // So: let the user pick agent type in-harness, then after harness exit, if
    // the branch was runAddItems AND nothing was selected, show EmptyPanel and
    // return. The splicer still needs to build *some* flow, so when the filter
    // count is 0 it returns a single NoteStep with the "everything installed"
    // message and the review step is skipped. Post-harness sees no selections
    // and returns without writing.

    // ... build step list, call harness.Run ...
    // ... pull results, persist config, generate, etc. ...
}
```

**Decision on empty-add-items UX:** When all filter counts are zero in the add-items branch, `buildAddItemsSteps` returns `[]Step{ NoteStep("All installed", "All available abilities are already installed. Browse more with: bonsai catalog") }`. The user presses Enter, the flow ends, nothing is written. Cleaner than a separate pre-harness EmptyPanel because the flow stays in AltScreen.

**Post-harness (runs after `harness.Run`):**

1. Pull typed results.
2. Branch on agent type + existence:
   - **New agent:** validate workspace uniqueness (delayed tech-lead-required check here too), build `InstalledAgent`, save config, run spinner + generate, render `TitledPanel("Adding") via TitledPanel(...)` — wait, the review is already shown in-harness; post-harness just runs the write pipeline + success banner.
   - **Existing agent:** append new selections to `installed.*`, save config, run spinner + generate, success banner.

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

- [ ] `make build` — clean compile.
- [ ] `go test ./...` — all tests green, incl. new `LazyGroup`, `NoteStep`, `TitledPanelString` cases.
- [ ] `gofmt -s -l .` — no output.
- [ ] `go vet ./...` — no issues.

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

## Iter 3 — Outline (detail in next iteration)

- Migrate `cmd/remove.go` (agent-level + per-item subcommands) onto the harness.
- Migrate `cmd/update.go` — custom-file scan becomes a step that emits a `MultiSelectStep` per agent that has discoveries (via `LazyStep`/`LazyGroup`).
- Move conflict-resolution picker (`resolveConflicts`) into the harness so post-generate prompts stay inside AltScreen.
- Move `huh/spinner` invocations into a `SpinnerStep` adapter so the whole flow is one program.

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
