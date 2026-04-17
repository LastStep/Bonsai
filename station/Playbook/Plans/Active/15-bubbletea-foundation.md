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
| 1 | Harness package + theme + `cmd/init.go` migration | Planned |
| 2 | Migrate `cmd/add.go` (incl. `runAddItems` pivot) | Planned |
| 3 | Migrate `cmd/remove.go` + `cmd/update.go` (custom-file scan, conflict picker, spinner step) | Planned |

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

## Iter 2 — Outline (detail in next iteration)

- Migrate `cmd/add.go` `runAdd` and `runAddItems` onto the harness.
- Add agent-type selection step using `SelectStep`.
- Branch the step list at runtime when `cfg.Agents[agentType]` exists (handled by `LazyStep` extending the stack with the "add items" sub-flow).
- Tech-lead workspace shortcut (info-line step instead of text input).

## Iter 3 — Outline

- Migrate `cmd/remove.go` (agent-level + per-item subcommands) onto the harness.
- Migrate `cmd/update.go` — custom-file scan becomes a step that emits a `MultiSelectStep` per agent that has discoveries (via `LazyStep`).
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
