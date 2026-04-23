# Plan 14 — UI/UX Overhaul Phase 3: Visual Identity & Init Polish

**Tier:** 2 (Feature)
**Status:** Complete — shipped 2026-04-17 (merged via PR #24 bundle)
**Source:** Group F backlog (2026-04-17 dogfooding session) — scope-picked Phase 3 slice
**Agent:** general-purpose

---

## Goal

Establish a cohesive visual identity for the TUI and polish the `bonsai init` flow so it feels professional, visible, and responsive instead of flat and stacked.

### Success Criteria

- `internal/tui/styles.go` exposes a **semantic palette layer** (`ColorPrimary`, `ColorAccent`, `ColorSubtle`, etc.) backed by the existing Zen Garden tokens — all style vars and panel borders reference semantic tokens, not raw colors.
- The welcome banner is redesigned — tighter wordmark, version line, and an optional contextual sub-line passed by the caller (e.g., "Initializing new project").
- After each `AskText` prompt in `bonsai init`, a styled summary line is printed so prior answers remain visible as the user advances through the flow.
- `tui.PickItems` emits an explicit "N required items auto-included" feedback line when a category has zero optional items (currently silent).
- `tui.ItemTree` category headers show item counts, e.g. `Skills (3)`, `Workflows (5)`, `Protocols (4)`.
- Huh form theme reduces vertical gap between prompt title and input, and the text cursor is styled for visibility.
- No regressions on existing flows (dark/light terminals, NO_COLOR, non-TTY).

---

## Context

Plan 11 (Phase 1) introduced the adaptive Zen Garden palette, NO_COLOR handling, `FatalPanel`, and a version banner. Plan 12 (Phase 2) added consistency polish (counts in `catalog`, deterministic `list`, `ActionUnchanged`, `ErrorDetail`). Plan 14 continues the UI overhaul with the taste-pass items surfaced during 2026-04-17 dogfooding of `bonsai init`:

Feedback from the user testing the init flow:
- "Everything is giving overall bulky feel"
- "The B O N S A I / agent scaffolder banner doesn't look professional"
- "The project name data gets hidden after you move on to the next question"
- "Required-only sections silently skip — hard to track what's done"
- "Should show counts alongside items like Skills (3)"
- "Typing cursor can we change that, use something sleek"
- "Overall coloring I don't like — set a proper color palette"

This plan addresses the **foundation** (semantic tokens) and **visible quick wins** (banner, answered-prompt persistence, required-only feedback, counts, prompt polish). Larger architectural items — screen lifecycle (AltScreen / redraw between steps), progressive disclosure of the selection step, go-back navigation, and the review → generate → complete flow redesign — are deferred to Phase 4+ because they need their own design pass.

**Decisions captured during planning:**

- **Palette first.** Every subsequent visual change consumes semantic tokens, not inline colors. Without this, later re-theming turns into a global find-replace hazard.
- **Banner signature adds an `action` param**, not environment sniffing — keeps the function pure and callers explicit.
- **`Answer()` is a new helper in `internal/tui/styles.go`**, called from `cmd/init.go` after each AskText. Don't build it into `AskText` itself — other callers (add/remove/update) may want to suppress the summary.
- **Required-only feedback lives in `PickItems`** — emit it unconditionally when optional count is zero AND required count is ≥1.
- **Category counts derive from `len(cat.Items)` in `ItemTree`** — no signature change, just append `(N)` to the header.
- **Prompt polish is theme-only** — no changes to `AskText` / `PickItems` signatures; just `BonsaiTheme()` tweaks.

---

## Steps

### Step 1 — Add Semantic Palette Tokens

**File:** `internal/tui/styles.go`

Add a new section after the `Zen Garden Palette` block (after line 39), before `// ─── Styles ───`:

```go
// ─── Semantic Tokens ──────────────────────────────────────────────────────
//
// Semantic aliases backed by the Zen Garden palette. Prefer these in new code
// and migrate existing callsites on touch. Swap the palette value here to
// re-theme the whole TUI in one place.

var (
	ColorPrimary   = Leaf  // Brand accent — headings, primary action, banner title
	ColorSecondary = Bark  // Field labels, category headers
	ColorAccent    = Petal // Interactive chrome — cursor, selectors, next/prev
	ColorSubtle    = Sand  // Body text, option labels
	ColorMuted     = Stone // Hints, descriptions, at-rest borders
	ColorSuccess   = Moss  // Success states
	ColorDanger    = Ember // Errors
	ColorWarning   = Amber // Warnings
	ColorInfo      = Water // Info panels, review box
)
```

Then rewrite the existing style vars (lines 44-52) to reference semantic tokens:

```go
var (
	StyleTitle   = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	StyleLabel   = lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary)
	StyleMuted   = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleError   = lipgloss.NewStyle().Foreground(ColorDanger)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleAccent  = lipgloss.NewStyle().Foreground(ColorInfo)
	StyleSand    = lipgloss.NewStyle().Foreground(ColorSubtle)
)
```

Rewrite the panel vars (lines 56-77) to use semantic tokens:

```go
var (
	PanelSuccess = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(1, 2)
	PanelError = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDanger).
			Padding(1, 2)
	PanelWarning = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWarning).
			Padding(1, 2)
	PanelInfo = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorInfo).
			Padding(1, 2)
	PanelEmpty = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted)
			Padding(1, 2)
)
```

Update `internal/tui/prompts.go` `BonsaiTheme()` (lines 12-53) to use semantic tokens:
- Replace `Bark` → `ColorSecondary`
- Replace `Stone` → `ColorMuted`
- Replace `Ember` → `ColorDanger`
- Replace `Petal` → `ColorAccent`
- Replace `Sand` → `ColorSubtle`
- Replace `Moss` → `ColorSuccess`
- Replace `Leaf` → `ColorPrimary`

Update `CatalogTable` in `styles.go` (line 415-438) — change `Bark` → `ColorSecondary`, `Leaf` → `ColorPrimary`, `Stone` → `ColorMuted`.

Update `Banner()` Leaf references to `ColorPrimary`.

Update `ItemTree()` and `FileTree()` `Stone` references to `ColorMuted`, `Sand` → `ColorSubtle`.

Update `TitledPanel()` — caller passes raw color, keep as-is but update `cmd/init.go:145` to pass `tui.ColorInfo` instead of `tui.Water`.

**Keep the Zen Garden palette vars (`Leaf`, `Bark`, etc.) — they remain as the underlying color source.** Semantic tokens are aliases, not replacements.

### Step 2 — Redesign the Banner

**File:** `internal/tui/styles.go`

Replace the `Banner()` function (lines 92-109) with:

```go
// Banner prints the Bonsai welcome banner.
// version is the build version (pass "" or "dev" to hide).
// action is an optional contextual sub-line (e.g., "Initializing new project").
// Pass "" for no action line.
func Banner(version, action string) {
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("BONSAI")
	tagline := StyleMuted.Render("agent scaffolder")
	header := title + "  " + tagline

	var lines []string
	lines = append(lines, header)

	if version != "" && version != "dev" {
		ver := StyleMuted.Render("v" + version)
		lines = append(lines, ver)
	}

	if action != "" {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorInfo).Render(action))
	}

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 3).
		Render(content)

	fmt.Println("\n" + indent(box, 2))
}
```

**Caller updates:**

- `cmd/init.go:39` — change `tui.Banner(Version)` to `tui.Banner(Version, "Initializing new project")`.
- Search for other `tui.Banner(` callers. If any exist in `cmd/add.go`, `cmd/update.go`, etc., update them to pass a contextual action string or `""`. Expected callers: only `cmd/init.go` based on Phase 11 introduction. Verify with `grep -rn "tui.Banner" cmd/`.

### Step 3 — Add `Answer()` Helper

**File:** `internal/tui/styles.go`

Add after `Hint()` (line 134):

```go
// Answer prints a compact styled summary of an answered prompt so prior answers
// stay visible as the user advances through a multi-step flow.
// Example output:   ▸ Project name   my-project
func Answer(label, value string) {
	key := StyleLabel.Render(label)
	val := value
	if strings.TrimSpace(value) == "" {
		val = StyleMuted.Render("(skipped)")
	} else {
		val = StyleSuccess.Render(value)
	}
	fmt.Println("  " + StyleMuted.Render(GlyphArrow) + " " + key + "  " + val)
}
```

**Caller updates in `cmd/init.go`:**

After line 45 (after `projectName` capture):
```go
tui.Answer("Project name", projectName)
```

After line 49 (after `description` capture):
```go
tui.Answer("Description", description)
```

After line 53 (after `docsPath` capture, BEFORE the validation/normalization block):
```go
tui.Answer("Station directory", docsPath)
```

(Place the Answer call right after the AskText so the user sees their raw input; validation still runs after.)

### Step 4 — Required-Only Section Feedback

**File:** `internal/tui/prompts.go`

In `PickItems` (lines 142-209), after the required-items display loop (after line 177), and before the "Collect required values" block, add:

```go
// If all items are required (no optional picker to show), give explicit
// feedback so the user knows the section was processed, not silently skipped.
if len(optional) == 0 && len(required) > 0 {
	plural := "s"
	if len(required) == 1 {
		plural = ""
	}
	fmt.Println("    " + StyleMuted.Render(fmt.Sprintf("(%d required item%s auto-included)", len(required), plural)))
}
```

The existing `fmt` import is already present; no new imports needed.

### Step 5 — Category Counts in ItemTree

**File:** `internal/tui/styles.go`

In `ItemTree` (lines 312-354), modify the category header render (around line 331):

Change:
```go
buf.WriteString("  " + bc.Render(branch) + StyleLabel.Render(cat.Name) + "\n")
```

To:
```go
header := StyleLabel.Render(cat.Name) + " " + StyleMuted.Render(fmt.Sprintf("(%d)", len(cat.Items)))
buf.WriteString("  " + bc.Render(branch) + header + "\n")
```

No signature change. `fmt` is already imported.

### Step 6 — Prompt Theme Polish

**File:** `internal/tui/prompts.go`

In `BonsaiTheme()` (lines 12-53), make these tweaks:

**a. Reduce margin between prompt title and input.** Add after `t.Focused.Title = ...` (line 19):
```go
t.Focused.Title = t.Focused.Title.MarginBottom(0)
```

**b. Stronger, more visible cursor.** Replace `t.Focused.TextInput.Cursor` (line 39) with:
```go
t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.
	Foreground(ColorPrimary).
	Bold(true)
```

**c. Tighten option row and focused group spacing.** After existing `t.Group.Description` line (line 50), add:
```go
t.Focused.Base = t.Focused.Base.PaddingLeft(1)
```

(Reduces the default padding that makes the framed prompt feel indented away from the label.)

If any of the above tweaks don't exist on the current Huh API version, fall back to the closest equivalent — keep the behavioral goal (tighter grouping, more visible cursor) and document what you did.

### Step 7 — Tests

**File:** `internal/tui/styles_test.go` (create if missing)

Add tests:

1. **`TestBannerIncludesAction`** — call `Banner("0.1.3", "Initializing new project")` and assert (via capturing stdout) that the output contains "BONSAI", "agent scaffolder", "v0.1.3", and "Initializing new project".
2. **`TestBannerHidesVersionWhenDev`** — `Banner("dev", "")` should not contain "vdev" or "v".
3. **`TestBannerHidesActionWhenEmpty`** — `Banner("0.1.3", "")` should not add an empty action line.
4. **`TestItemTreeShowsCategoryCounts`** — build an `ItemTree` with two categories (Skills=3 items, Workflows=5) and assert the output contains `Skills (3)` and `Workflows (5)`.
5. **`TestAnswerRendersKeyValue`** — call `Answer("Project name", "my-project")` and assert stdout contains "Project name" and "my-project".
6. **`TestAnswerShowsSkippedForEmpty`** — `Answer("Description", "")` should contain `(skipped)`.

Use `os.Pipe()` stdout capture pattern (see existing `generate_test.go` for reference).

**File:** `internal/tui/prompts_test.go` (create if missing)

Add:

7. **`TestPickItemsRequiredOnlyPrintsFeedback`** — This is hard to unit-test because `PickItems` calls `huh` which expects a TTY. Skip if no good seam exists; rely on manual verification (Verification section below).

---

## Security

> [!warning]
> Refer to [SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

- **No user input interpolation into format strings.** The `Answer(label, value)` helper renders `value` through `StyleSuccess.Render()` which doesn't interpret format specifiers — safe.
- **Banner `action` param is rendered, not evaluated.** Safe.
- **No new file I/O, no new dependencies.** Keep `go.mod` unchanged.
- **No shell execution or template interpolation changes.** Safe.

---

## Dependencies

- No new packages. All changes use existing `github.com/charmbracelet/huh`, `github.com/charmbracelet/lipgloss`.
- No catalog changes.
- No generator changes.

---

## Verification

### Build & Test

- [ ] `make build` — compiles with no errors or warnings
- [ ] `go test ./...` — all existing + new tests pass
- [ ] `gofmt -s -l .` — no formatting issues
- [ ] `go vet ./...` — no issues

### Manual — `bonsai init` flow

Run `mkdir /tmp/test-plan14 && cd /tmp/test-plan14 && /path/to/bonsai init`:

- [ ] Banner: shows "BONSAI" wordmark (not "B O N S A I" spaced), "agent scaffolder" tagline, version line if non-dev, "Initializing new project" contextual line
- [ ] After typing project name: a summary line `  → Project name  <value>` appears below the prompt and remains visible
- [ ] Same for description and station directory
- [ ] Protocols section (required-only for tech-lead): prints `(4 required items auto-included)` (or whatever count applies)
- [ ] Review panel: category headers show counts — `Skills (N)`, `Workflows (N)`, `Protocols (N)`, `Sensors (N)`, `Routines (N)`
- [ ] Cursor in text input is visible/bold in primary color
- [ ] Title and input don't have a large vertical gap

### Manual — regressions

- [ ] `NO_COLOR=1 bonsai init` — no ANSI escapes in output; flow still completes
- [ ] `bonsai init` on a light terminal — colors legible (verify on macOS Terminal default or equivalent)
- [ ] `bonsai list` — unaffected
- [ ] `bonsai add` — unaffected, still functional
- [ ] `bonsai catalog` — unaffected

---

## Dispatch

| Agent | Isolation | Notes |
|-------|-----------|-------|
| general-purpose | worktree | Go + TUI changes only. No catalog, no generator, no docs. |

---

## Out of Scope (Phase 4+)

- **TUI screen lifecycle** — clearing prior content on review/generate/complete transitions (needs AltScreen or explicit redraw architecture)
- **Progressive disclosure** for scaffolding + ability selection — break the "wall of text" step into focused sub-screens
- **Go-back navigation** in multi-step init flow
- **Review → generate → complete flow redesign** — rich file tree visual, modern generate confirm, verbose next-steps panel
- **Non-fullscreen panel border rendering bug** — separate P1 bug fix
- **`go install .` binary name bug** — separate P1 build fix
