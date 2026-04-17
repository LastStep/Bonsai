# BubbleTea + Huh Architectural Patterns

Survey of official Charmbracelet examples (`charmbracelet/bubbletea` and `charmbracelet/huh` repos). Extracted patterns applicable to Bonsai's Phase 4 TUI architecture (multi-step init flow, progressive disclosure, screen lifecycle).

---

## High-Priority Patterns (Phase 4 Blockers)

### 1. Huh Form as BubbleTea Model (THE KEY PATTERN)

**Pattern:** Embedding a Huh form directly inside a BubbleTea tea.Model  
**Solves:** Unified lifecycle management; forms no longer run in isolation but integrate with BubbleTea's event loop, enabling AltScreen, multi-view transitions, and answer persistence.

**Key Idioms:**
- Store `*huh.Form` as a field in the tea.Model struct
- Wire form's Update output into tea.Model.Update: `form, cmd := m.form.Update(msg)` → validate type assertion `ok, f := form.(*huh.Form)` → reassign `m.form = f`
- Check `m.form.State == huh.StateCompleted` to detect form submission (then transition to next step or quit)
- Call `m.form.Init()` in tea.Model.Init() to initialize form keybindings
- Access completed values via `m.form.GetString(key)` / `m.form.GetString(key)` family of methods
- Huh forms auto-handle key events (arrow keys, Tab, Enter) — no manual routing needed

**Applicability to Bonsai:**  
CRITICAL. Replace serial `cmd/init.go` → `huh.Form.Run()` invocations with one BubbleTea model housing multiple Huh forms. Each step is a state transition (`scaffolding` → `agent` → `abilities` → `review` → `generate` → `complete`); form changes trigger state machine instead of exiting and relaunching.

**Source:** `huh/examples/bubbletea/main.go` (7KB, complete integration example)

---

### 2. Huh Form ViewHook for AltScreen Entry

**Pattern:** Use `.WithViewHook()` to set `v.AltScreen = true` before rendering  
**Solves:** Clean screen transition without output stacking. Form takes over the terminal, clearing prior output.

**Key Idioms:**
```go
form := huh.NewForm(...).WithViewHook(func(v tea.View) tea.View {
    v.AltScreen = true
    return v
})
```
- Called on every frame; idempotent
- Alternative to managing AltScreen in tea.Model.View() — useful for form-centric apps

**Applicability to Bonsai:**  
MODERATE. Useful for the scaffolding form and review panel (screen takeover). But when embedded in BubbleTea model, prefer managing AltScreen in Model.View() for finer control over transitions (review → generate should clear prior output).

**Source:** `huh/examples/bubbletea-options/main.go` (415 bytes, minimal example)

---

### 3. Multi-Group Form for Sequential Disclosure

**Pattern:** Single form with multiple `.NewGroup()` blocks rendered sequentially or in scroll-view  
**Solves:** Layering questions into logical sections without multi-step handoff. Groups can be independent or conditional.

**Key Idioms:**
- `huh.NewForm(group1, group2, group3)` — groups render in order, each full-width
- Each group has its own title/description (implicit section headers)
- Tab/Enter navigate *within* the form across groups; form completes only when all groups filled
- Groups scroll if content exceeds height: `.WithHeight()` limits space, enabling multi-screen feel within one form
- *No form state reset* between groups — answers persist

**Applicability to Bonsai:**  
HIGH. init flow is naturally multi-group: scaffolding (project name, description, path), agent type, ability selection, review, confirmation. Can be one form with 5 groups instead of 5 separate prompts. Avoids answer loss and screen stacking.

**Source:** `huh/examples/multiple-groups/main.go` (1.9KB)

---

### 4. Conditional Field Visibility (TitleFunc / OptionsFunc)

**Pattern:** Use `.TitleFunc()` and `.OptionsFunc()` with dependency keys to dynamically compute field label and choices based on prior answers  
**Solves:** Progressive disclosure — next field appears only when its preconditions are met.

**Key Idioms:**
```go
huh.NewSelect[string]().
    TitleFunc(func() string {
        return fmt.Sprintf("Okay, what kind of %s?", category)
    }, &category). // pass dependency keys
    OptionsFunc(func() []huh.Option[string] {
        switch category {
        case "fruit":
            return [...fruitOptions...]
        case "vegetable":
            return [...vegetableOptions...]
        }
    }, &category)
```
- Closure captures model state
- Framework re-evaluates on each Update for any keyed field that changes
- Title and options both react to prior selection

**Applicability to Bonsai:**  
MODERATE. Abilities selection could use this: "Abilities (for Agent X)" title reacts to agent type; ability categories shown depend on scaffolding type (e.g., Python scaffold shows "Python Skills" only). But consider Group F's "progressive disclosure" concern — may still want explicit sequential screens (tabs/separate form steps) for clarity.

**Source:** `huh/examples/conditional/main.go` (1.9KB)

---

### 5. Form Navigation: Skip Groups

**Pattern:** Groups have `.Skip()` / `.SkipFunc()` to conditionally omit sections  
**Solves:** Branching logic without creating parallel forms; required-only paths can auto-skip optional sections.

**Key Idioms:**
- `.SkipFunc(func() bool { return !userNeedsThis })` — group is not shown if Skip returns true
- Form still completes successfully; skipped values are nil/default
- Example: "Buy 1 Get 1" section only shows if user previously selected a burger type

**Applicability to Bonsai:**  
LOW-MODERATE. Useful for "required fields only" fast-path (e.g., if user selects `--quick` flag, skip abilities group). But init flow doesn't have many optional branches yet.

**Source:** `huh/examples/skip/main.go` (884 bytes)

---

### 6. Theming API for Canonical Palette

**Pattern:** Define a `huh.Theme` and pass via `.WithTheme(theme)`. Built-in themes: Dracula, Base16, Charm, Catppuccin.  
**Solves:** Decouples visual identity from form logic; enables dark/light mode and user theme selection.

**Key Idioms:**
```go
themes := map[string]huh.Theme{
    "default": huh.ThemeFunc(huh.ThemeBase),
    "dracula": huh.ThemeFunc(huh.ThemeDracula),
    // ...
}
form.WithTheme(themes["chosen-theme"])
```
- Themes are composable: define custom theme by extending base
- Theme controls cursor style, borders, focus highlight, input colors

**Applicability to Bonsai:**  
HIGH. Group F requirement: "Define a canonical color palette for the whole TUI". Rather than ad-hoc LipGloss colors, define one `huh.Theme` with Bonsai branding (primary, accent, muted, success, warning, error tokens). Apply to all Huh forms + LipGloss panels via a single source.

**Source:** `huh/examples/theme/main.go` (1.5KB)

---

## Medium-Priority Patterns (Phase 4 Enhancements + Polish)

### 7. BubbleTea Multi-View Routing (State Machine)

**Pattern:** Model.View() dispatches to different sub-view functions based on model state; Model.Update() dispatches to different sub-update functions  
**Solves:** Clean separation of concerns for 3+ screens. Each screen has its own Update logic and View rendering.

**Key Idioms:**
```go
type model struct {
    state   viewState // iota: choicesView, loadingView, resultView
    choice  int
    data    interface{}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if !m.Chosen {
        return updateChoices(msg, m)
    }
    return updateChosen(msg, m)
}

func (m model) View() tea.View {
    if !m.Chosen {
        return tea.NewView(choicesView(m))
    }
    return tea.NewView(chosenView(m))
}
```
- State-based dispatch (switch on m.state)
- Each sub-update returns (model, cmd) — compose into main Update
- View functions are pure; take model and return string

**Applicability to Bonsai:**  
HIGH. Bonsai's init flow is a state machine: stateScaffolding → stateAgent → stateAbilities → stateReview → stateGenerate → stateComplete. Each state has distinct Update and View logic. Current serial form approach has no state machine; BubbleTea dispatch pattern is the idiomatic Go TUI way.

**Source:** `bubbletea/examples/views/main.go` (6.6KB)

---

### 8. Composable Sub-Models (Bubble Components)

**Pattern:** Embed smaller models (e.g., timer.Model, spinner.Model from bubbles lib) into main model. Route messages and Update to focused child.  
**Solves:** Reusable, testable component composition; e.g., progress bar, spinner, timer run independently.

**Key Idioms:**
```go
type mainModel struct {
    state    sessionState // timerView, spinnerView
    timer    timer.Model
    spinner  spinner.Model
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    switch m.state {
    case spinnerView:
        m.spinner, cmd := m.spinner.Update(msg)
        cmds = append(cmds, cmd)
    case timerView:
        m.timer, cmd := m.timer.Update(msg)
        cmds = append(cmds, cmd)
    }
    return m, tea.Batch(cmds...)
}

func (m mainModel) View() tea.View {
    // Render focused child with border, unfocused without
    return tea.NewView(renderComposed(m))
}
```
- Children are standard tea.Model implementations
- Parent routes KeyPressMsg and child-specific messages (e.g., timer.TickMsg, spinner.TickMsg)
- Parent View() composes child.View() outputs with JoinHorizontal/JoinVertical

**Applicability to Bonsai:**  
MODERATE. Useful for embedding reusable widgets (ItemTree for ability selection, file tree preview in review panel). But Huh forms are opaque sub-models — better to embed Huh forms directly (Pattern #1) than treat them as generic bubbles components.

**Source:** `bubbletea/examples/composable-views/main.go` (3.8KB)

---

### 9. AltScreen Toggle for Screen Lifecycle

**Pattern:** Set `v.AltScreen = true` in View() to use alternate screen buffer (clears terminal, no history visible); set to false for inline mode  
**Solves:** Explicit control over whether prior output persists; useful for form-heavy apps (init, wizard flow).

**Key Idioms:**
```go
func (m model) View() tea.View {
    v := tea.NewView(myContent)
    v.AltScreen = m.altscreen // toggle via spacebar, for example
    return v
}
```
- AltScreen true: terminal alternate buffer (clean slate, no scroll history)
- AltScreen false: inline mode (output stacks, scrolls with terminal)
- Ctrl+Z (Suspend) preserves AltScreen state on resume

**Applicability to Bonsai:**  
HIGH (Group F dependency). Group F issue: "TUI screen lifecycle — clear prior step output on major transitions". AltScreen = true for review, generate, complete screens. = false for initial banner/scaffolding (so user can scroll back). Or use AltScreen for entire init flow to keep it contained.

**Source:** `bubbletea/examples/altscreen-toggle/main.go` (1.6KB)

---

### 10. Tab-Based UI Navigation

**Pattern:** Implement active/inactive tab styling; dispatch keyboard navigation (left/right, Tab/Shift+Tab) to switch tabs; render tab bar + content  
**Solves:** Progressive disclosure alternative to sequential forms — all tabs visible, user picks focus.

**Key Idioms:**
```go
type model struct {
    Tabs      []string
    activeTab int
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case "right", "tab":
        m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
}

func (m model) View() tea.View {
    var renderedTabs []string
    for i, t := range m.Tabs {
        if i == m.activeTab {
            renderedTabs = append(renderedTabs, activeTabStyle.Render(t))
        } else {
            renderedTabs = append(renderedTabs, inactiveTabStyle.Render(t))
        }
    }
    row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
    content := tabContent[m.activeTab]
    return tea.NewView(row + "\n" + contentStyle.Render(content))
}
```
- Tabs are just strings; content array parallels Tabs array
- Active tab is highlighted; inactive tabs are dimmed
- Borders customized per tab state

**Applicability to Bonsai:**  
MODERATE. Group F: "Progressive disclosure for project scaffolding step — consider tabs". Could implement init as 3 tabs: Scaffolding, Agent, Abilities. User sees all categories but focuses on one at a time. But sequential form groups (Pattern #3) may be more intuitive for first-time users.

**Source:** `bubbletea/examples/tabs/main.go` (3.2KB)

---

### 11. Responsive Window Sizing

**Pattern:** Listen to `tea.WindowSizeMsg` in Update(); cache width/height; apply in View()  
**Solves:** TUI adapts to terminal resize; no hardcoded widths.

**Key Idioms:**
```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height

// In View():
content := lipgloss.NewStyle().Width(m.width - 4).Render(text)
```
- WindowSizeMsg fires on startup and on resize
- Pass width to form: `form.WithWidth(m.width - padding)`
- Render panels with responsive width

**Applicability to Bonsai:**  
HIGH. Current Huh forms are hardcoded `WithWidth(45)`. Should respond to terminal size.

**Source:** `bubbletea/examples/window-size/main.go` (801 bytes)

---

### 12. Keybinding Help UI (key.Binding + help.Model)

**Pattern:** Define keybindings as `key.Binding` objects with short/full help text. Use `help.Model` to render toggle-able help view (? to expand).  
**Solves:** User discovers commands; help reflows to window width.

**Key Idioms:**
```go
var keys = keyMap{
    Up: key.NewBinding(
        key.WithKeys("up", "k"),
        key.WithHelp("↑/k", "move up"),
    ),
    Help: key.NewBinding(
        key.WithKeys("?"),
        key.WithHelp("?", "toggle help"),
    ),
}

func (k keyMap) ShortHelp() []key.Binding { return []key.Binding{k.Help, k.Quit} }
func (k keyMap) FullHelp() [][]key.Binding { return [][]key.Binding{...} }

// In Update():
case key.Matches(msg, m.keys.Help):
    m.help.ShowAll = !m.help.ShowAll

// In View():
return tea.NewView(status + m.help.View(m.keys))
```
- help.Model() handles layout, wrapping, column layout
- Supports short (one-line) and full (multi-column) modes

**Applicability to Bonsai:**  
MODERATE. Bonsai init doesn't need rich help (it's guided, not exploratory). But useful for future commands (e.g., `bonsai manage --help` with interactive keybinding guide).

**Source:** `bubbletea/examples/help/main.go` (3.3KB)

---

### 13. Splash Screen with Gradient Animation

**Pattern:** Use AltScreen, draw full-screen content on startup, animate with frames via `tea.Tick()`  
**Solves:** Professional brand moment; visual feedback while initializing.

**Key Idioms:**
- Store `rate int64` (animation speed) and track time with ticker
- Compute animation frame based on elapsed time
- Render with color gradients via LipGloss foreground/background
- Example: rotating gradient background across screen

**Applicability to Bonsai:**  
LOW-MODERATE. Group F: "Redesign the B O N S A I banner". Could add animated gradient or rotating border. But init flow is task-focused (users want to get started); splash screens add latency. Better to invest in banner redesign (static) and then add animation if UX testing shows users appreciate it.

**Source:** `bubbletea/examples/splash/main.go` (3.9KB)

---

### 14. Command Sequencing (tea.Sequence / tea.Batch)

**Pattern:** Chain commands using `tea.Sequence()` (runs in order) or `tea.Batch()` (concurrent). Compose with nesting.  
**Solves:** Orchestrate multi-step async operations (e.g., scaffolding step 1 → step 2 → generate → complete).

**Key Idioms:**
```go
func (m model) Init() tea.Cmd {
    return tea.Sequence(
        tea.Batch(step1, step2), // run 1 & 2 in parallel
        step3,                   // then run 3
        tea.Quit,
    )
}
```
- `Sequence` runs commands in order; each waits for prior to complete
- `Batch` runs commands concurrently
- Can nest arbitrarily deep

**Applicability to Bonsai:**  
MODERATE. Not needed for init (form updates are already sequenced by user presses). But useful for post-generate (write files, then show success, then next steps).

**Source:** `bubbletea/examples/sequence/main.go` (1.4KB)

---

## Summary: Top Patterns for Bonsai Phase 4

### Ranked by Impact to Unlock Phase 4:

1. **Huh Form as BubbleTea Model (Pattern #1)** — THE critical architectural shift. Enables unified lifecycle, state machine, answer persistence, and screen transitions. All other patterns depend on this.

2. **Multi-Group Form for Sequential Disclosure (Pattern #3)** — Maps directly to init flow (scaffolding → agent → abilities → review). Keeps answers visible across groups, no screen stacking.

3. **BubbleTea Multi-View State Machine (Pattern #7)** — Structured approach to routing init flow stages (stateScaffolding, stateAgent, stateAbilities, stateReview, stateGenerate, stateComplete). Each state has distinct Update/View.

4. **AltScreen Toggle for Screen Lifecycle (Pattern #9)** — Solves Group F: "clear prior step output on major transitions". AltScreen for review/generate/complete; inline for scaffolding.

5. **Theming API for Canonical Palette (Pattern #6)** — Solves Group F prerequisite: "Define a canonical color palette". Create one Bonsai huh.Theme, apply to all forms + LipGloss panels.

---

## Not Directly Applicable

- **Splash Screen (Pattern #13)**: init is task-focused, not entertainment-focused. Skip animated splash.
- **Composable Sub-Models (Pattern #8)**: Huh forms are opaque; don't treat as generic bubbles. Embed directly (Pattern #1).
- **Tab UI (Pattern #10)**: Viable alternative to sequential groups, but sequential groups feel more natural for guided init flow.

---

## References & File Locations

| Pattern | File | Lines |
|---------|------|-------|
| Huh ↔ BubbleTea embedding | `huh/examples/bubbletea/main.go` | 7.0K |
| AltScreen hook | `huh/examples/bubbletea-options/main.go` | 415B |
| Multi-group form | `huh/examples/multiple-groups/main.go` | 1.9K |
| Conditional visibility | `huh/examples/conditional/main.go` | 1.9K |
| Skip groups | `huh/examples/skip/main.go` | 884B |
| Theming | `huh/examples/theme/main.go` | 1.5K |
| Multi-view dispatch | `bubbletea/examples/views/main.go` | 6.6K |
| Composable models | `bubbletea/examples/composable-views/main.go` | 3.8K |
| AltScreen toggle | `bubbletea/examples/altscreen-toggle/main.go` | 1.6K |
| Tabs | `bubbletea/examples/tabs/main.go` | 3.2K |
| Window sizing | `bubbletea/examples/window-size/main.go` | 801B |
| Keybinding help | `bubbletea/examples/help/main.go` | 3.3K |
| Splash + animation | `bubbletea/examples/splash/main.go` | 3.9K |
| Sequence/Batch | `bubbletea/examples/sequence/main.go` | 1.4K |

---

## Group F Mapping

Backlog items addressed by this survey:

- **Line 112 (canonical palette)**: Pattern #6 (Theming API)
- **Line 114 (sleeker cursor + input grouping)**: Pattern #6 (theme offers cursor styles); apply multi-group layout (Pattern #3)
- **Line 115 (persist answered values)**: Pattern #1 + #3 (form state persists across groups when using single huh.Form)
- **Line 116 (go-back navigation)**: Huh supports Shift+Tab for field navigation; embed in BubbleTea model (Pattern #1) for full form state rewinding
- **Line 117 (progressive disclosure for scaffolding)**: Pattern #3 (groups) or Pattern #10 (tabs); Pattern #3 recommended
- **Line 120 (clear prior output on major transitions)**: Pattern #9 (AltScreen toggle)
- **Line 121 (modernize review → generate → complete)**: Combine Pattern #7 (state machine) with Pattern #9 (AltScreen) for clean screen transitions; Pattern #8 (composable views) for file tree preview widget

