# Plan 11 — UI/UX Overhaul Phase 1: Foundation

**Tier:** 2 (Feature)
**Status:** Complete — shipped 2026-04-17
**Source:** Roadmap Phase 1 "UI overhaul" + RESEARCH-uiux-overhaul.md (Phase 1 items 1.1–1.7)

---

## Goal

Make the Bonsai CLI render correctly in all terminal environments — light terminals, dark terminals, NO_COLOR, TERM=dumb, piped output — without changing the visual design on dark terminals. Add a centralized error format, consistent spacing, and version in the banner.

### Success Criteria

- Dark terminals: zero visual regression (dark palette values are unchanged)
- Light terminals: all text readable (uses darker palette variants)
- `NO_COLOR=1` / `TERM=dumb` / piped output → zero ANSI escapes
- `--no-color` flag disables color on any command
- No double-spacing between elements; every command ends with one `tui.Blank()`
- All fatal exits use structured `FatalPanel` (title/detail/hint)
- `bonsai init` banner shows CLI version
- `design-guide` catalog skill has Bonsai-specific CLI design rules

---

## Context

All 9 palette colors in `styles.go` are hardcoded hex values — unreadable on light terminals. NO_COLOR standard and non-TTY detection are missing. Error exits use three inconsistent patterns. Some commands lack trailing whitespace. The banner shows no version. The `design-guide` skill has generic frontend content instead of Bonsai CLI rules. Phase 1 fixes all of this with zero new dependencies.

---

## Steps

### Step 1 — Adaptive Color Palette

**Files:** `internal/tui/styles.go`, `internal/tui/prompts.go`

#### 1A. Convert palette to `lipgloss.AdaptiveColor` (`styles.go:13-23`)

Replace the 9 hardcoded `lipgloss.Color` declarations with `lipgloss.AdaptiveColor`. This type implements `lipgloss.TerminalColor` (same interface as `lipgloss.Color`), so all existing callsites (`.Foreground()`, `.BorderForeground()`, `TitledPanel` third arg) work without changes. Verified: 27+ usages across the codebase — all accept `TerminalColor`, none do type assertions.

```go
// Replace lines 13-23
var (
    Leaf  lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#4A9E6F", Light: "#2D7A4B"}
    Bark  lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#D4A76A", Light: "#8B6914"}
    Stone lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#6B7280", Light: "#4B5563"}
    Water lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#7EC8E3", Light: "#1A7FA0"}
    Moss  lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#73D677", Light: "#2D8A3E"}
    Ember lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#E36F6F", Light: "#C53030"}
    Amber lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#E3C16F", Light: "#B7791F"}
    Sand  lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#C4B7A6", Light: "#6B5E4F"}
    Petal lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#D4A0C0", Light: "#9B4D8A"}
)
```

Dark values are identical to current hex strings — zero visual change on dark terminals. `AdaptiveColor.color()` calls `r.HasDarkBackground()` internally (cached via `sync.Once`).

#### 1B. Fix hardcoded colors in `BonsaiTheme()` (`prompts.go`)

**Line 15** — focused base border uses `#3A5F4A` (muted green). Preserve the dark value via inline `AdaptiveColor` — do **not** substitute with `Leaf`, which would alter the border on dark terminals and violate the "zero visual regression" success criterion:
```go
t.Focused.Base = t.Focused.Base.BorderForeground(
    lipgloss.AdaptiveColor{Dark: "#3A5F4A", Light: "#5A7F6A"},
)
```

**Line 32** — blurred button background uses `#2D2D3D` (dark blue-gray, invisible on light). Replace with adaptive:
```go
t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(Sand).Background(
    lipgloss.AdaptiveColor{Dark: "#2D2D3D", Light: "#E5E7EB"},
)
```

**Line 31** (`FocusedButton` white-on-`Leaf`): leave as-is. White text on the adaptive `Leaf` green reads cleanly on both dark and light terminals. No change needed.

---

### Step 2 — NO_COLOR / Non-TTY Detection + `--no-color` Flag

**Files:** `internal/tui/styles.go`, `cmd/root.go`

#### 2A. Add `init()` and `DisableColor()` to `styles.go`

Add imports `"os"`, `"github.com/mattn/go-isatty"`, `"github.com/muesli/termenv"` (both already indirect deps in go.sum — promote to direct).

Add after imports, before palette section:

```go
func init() {
    if !isatty.IsTerminal(os.Stdout.Fd()) || termenv.EnvNoColor() {
        DisableColor()
    }
}

// DisableColor forces all output to plain text (no ANSI escapes).
func DisableColor() {
    lipgloss.SetColorProfile(termenv.Ascii)
}
```

`termenv.EnvNoColor()` handles `NO_COLOR` env var and `CLICOLOR=0`. The `isatty` check handles piped output. `TERM=dumb` is handled by lipgloss's default `ColorProfile()` which falls through to Ascii for unrecognized terminals.

#### 2B. Add `--no-color` persistent flag (`root.go`)

Add `init()` function in root.go (Go allows multiple per package, no existing `init()` in root.go):
```go
func init() {
    rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
}
```

Add `PersistentPreRun` to `rootCmd` definition (after `Short` field, line 29). Verified: no existing `PersistentPreRun` on rootCmd, no `PreRun`/`PreRunE` on any child command.
```go
PersistentPreRun: func(cmd *cobra.Command, args []string) {
    if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
        tui.DisableColor()
    }
},
```

After implementation, run `go mod tidy` to update `go.mod` with the promoted direct imports.

---

### Step 3 — Version in Banner

**Files:** `internal/tui/styles.go`, `cmd/root.go`, `cmd/init.go`

#### 3A. Fix `SetVersion` to update package-level `Version` (`root.go:50-52`)

Currently only sets `rootCmd.Version`. Add `Version = v`:
```go
func SetVersion(v string) {
    Version = v
    rootCmd.Version = v
}
```

#### 3B. Change `Banner()` signature (`styles.go:77-89`)

Single call site confirmed: only `init.go:39`.

```go
func Banner(version string) {
    title := lipgloss.NewStyle().Bold(true).Foreground(Leaf).Render("B O N S A I")
    subtitle := "agent scaffolder"
    if version != "" && version != "dev" {
        subtitle += " " + GlyphDot + " v" + version
    }
    sub := StyleMuted.Render(subtitle)

    box := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(Leaf).
        Padding(1, 5).
        Align(lipgloss.Center).
        Render(title + "\n" + sub)

    fmt.Println("\n" + indent(box, 3))
}
```

#### 3C. Update call site (`init.go:39`)

```go
tui.Banner(Version)
```

---

### Step 4 — Standardize Spacing

**Files:** All `cmd/*.go` — verified line-by-line by audit agent.

**Contract:** Display helpers own their top margin (`\n` prefix). Commands own exactly one trailing `tui.Blank()` **on success paths** — after `tui.Success(...)` and before the final `return nil`. Cancelled paths (user answers "No" at a confirm) and guard paths (`ErrorPanel` + `return nil`) do not add a trailing `Blank()` — the rendered panel already owns its own trailing line. No double-spacing anywhere.

#### Remove double-spacing (TitledPanel + Blank before AskConfirm):

| File | Line | Action |
|------|------|--------|
| `cmd/init.go` | 148 | Remove `tui.Blank()` |
| `cmd/add.go` | 184 | Remove `tui.Blank()` |
| `cmd/add.go` | 356 | Remove `tui.Blank()` |
| `cmd/remove.go` | 84 | Remove `tui.Blank()` |
| `cmd/remove.go` | 292 | Remove `tui.Blank()` |

#### Add missing trailing `tui.Blank()`:

| File | After | Current end of function | Add |
|------|-------|------------------------|-----|
| `cmd/add.go` | line 230 | `tui.Success(...)` return nil (end of `runAdd`) | `tui.Blank()` before return |
| `cmd/add.go` | line 399 | `tui.Success(...)` return nil (end of `runAddItems`) | `tui.Blank()` before return |
| `cmd/remove.go` | line 134 | `tui.Success(...)` return nil (end of `runRemove`) | `tui.Blank()` before return |
| `cmd/remove.go` | line 360 | `tui.Success(...)` return nil (end of `runRemoveItem`) | `tui.Blank()` before return |
| `cmd/update.go` | line 188 | `tui.Success(...)` return nil (end of `runUpdate`) | `tui.Blank()` before return |

Note: `init.go` already has `tui.Hint()` + `tui.Blank()` after Success — no change needed. `list.go` and `catalog.go` already have trailing `tui.Blank()`. `guide.go` uses raw `fmt.Print` — skip.

---

### Step 5 — Centralized `FatalPanel`

**Files:** `internal/tui/styles.go`, `cmd/root.go`, `cmd/init.go`, `cmd/add.go`, `cmd/remove.go`

#### 5A. Add `FatalPanel()` to `styles.go` (after `ErrorPanel`, ~line 151)

Add `"os"` to imports.

```go
// FatalPanel renders a structured error and exits. title: what happened. detail: why. hint: how to fix.
func FatalPanel(title, detail, hint string) {
    content := StyleError.Bold(true).Render(title)
    if detail != "" {
        content += "\n" + detail
    }
    if hint != "" {
        content += "\n" + StyleMuted.Render(hint)
    }
    fmt.Println("\n" + indent(PanelError.Render(content), 2))
    os.Exit(1)
}
```

#### 5B. Migrate fatal exits (only `ErrorPanel`/`Error` + `os.Exit(1)` patterns)

| File | Lines | Current Pattern | New `FatalPanel(title, detail, hint)` |
|------|-------|-----------------|---------------------------------------|
| `cmd/root.go` | 35-37 | `ErrorPanel(...)` + `os.Exit(1)` | `"Failed to load catalog"`, `err.Error()`, `"This is a bug — please report it."` |
| `cmd/root.go` | 43-45 | `ErrorPanel(...)` + `os.Exit(1)` | `"No "+configFile+" found"`, `"This command requires an initialized project."`, `"Run: bonsai init"` |
| `cmd/init.go` | 56-58 | `ErrorPanel(...)` + `os.Exit(1)` | `"Invalid station directory"`, `"Cannot be empty or root."`, `"Use a subdirectory like: station/"` |
| `cmd/init.go` | 89-91 | `ErrorPanel(...)` + `os.Exit(1)` | `"Tech Lead agent not found"`, `"The built-in catalog is missing the tech-lead agent."`, `"This is a bug — please report it."` |
| `cmd/add.go` | 95-97 | `Error(...)` + `os.Exit(1)` | `"Unknown agent type"`, `agentType+" is not in the catalog."`, `"Run: bonsai catalog"` |
| `cmd/add.go` | 123-125 | `ErrorPanel(...)` + `os.Exit(1)` | `"Workspace conflict"`, `workspace+" is already in use by another agent."`, `"Choose a different directory."` |
| `cmd/add.go` | 238-240 | `Error(...)` + `os.Exit(1)` | `"Unknown agent type"`, `agentType+" is not in the catalog."`, `"Run: bonsai catalog"` |
| `cmd/remove.go` | 54-56 | `Error(...)` + `os.Exit(1)` | `"Agent not installed"`, `agentName+" is not in the current project."`, `"Run: bonsai list"` |

**Do NOT migrate** (these use `ErrorPanel` + `return nil` — recoverable flow exits, not fatal errors):
- `add.go:88-89` (Tech Lead required)
- `remove.go:60-61` (Can't remove tech-lead)
- `remove.go:201-202` (routine-check auto-managed)
- `remove.go:229-230` (item not installed in any agent)
- `remove.go:273-274` (item is required)

---

### Step 6 — Design Guide Catalog Skill

**Files:** `catalog/skills/design-guide/meta.yaml`, `catalog/skills/design-guide/design-guide.md`

#### 6A. Replace `meta.yaml`

Current: targets frontend/fullstack with CSS/component paths. New: targets all code-touching agents with Bonsai TUI paths.

```yaml
name: design-guide
description: Bonsai CLI design system — palette, spacing, panels, errors, command flow
agents: [tech-lead, backend, frontend, fullstack]
triggers:
  scenarios:
    - Modifying TUI styles, panels, display helpers, or color definitions
    - Adding or changing CLI output formatting in any command
  paths:
    - "internal/tui/**"
    - "cmd/*.go"
```

#### 6B. Replace `design-guide.md`

Complete rewrite (~100 lines) with Bonsai-specific CLI design rules codified from the research document:

1. **Design Principles** — Respect the Terminal, One Voice, Progressive Disclosure, Feedback is Mandatory, Understated Personality
2. **Adaptive Palette** — 9 tokens, `AdaptiveColor` requirement, semantic roles, never hardcode `lipgloss.Color("#hex")`
3. **Glyph Set** — 6 glyphs, no additions without justification
4. **Panel Vocabulary** — when to use each panel type (Success, Error, Warning, Info, Empty, Titled, Fatal)
5. **Spacing Contract** — helpers own top margin, commands own trailing `tui.Blank()`, no double-spacing
6. **Canonical Command Flow** — Heading → Input → Review → Confirm → Execute → Results → Success + Hint + Blank
7. **Error Format** — `FatalPanel` for `os.Exit(1)` cases; `ErrorPanel` for recoverable; never bare `fmt.Println` for errors
8. **Anti-Patterns** — hardcoded hex colors, emoji, double-spacing, bare `os.Exit` without panel, raw Go errors to users

---

## Dependencies

- Zero new Go dependencies. `termenv` v0.16.0 and `go-isatty` v0.0.20 promoted from indirect to direct.
- Steps are independent. Recommended order: Steps 1+2 first (shared styles.go import block), then 3–6.

---

## Security

> [!warning]
> Refer to SecurityStandards.md for all security requirements.

- No secrets, credentials, or API keys involved
- Promoted deps already in the dependency tree — no supply chain change
- `FatalPanel` calls `os.Exit(1)` — same as existing behavior
- `--no-color` flag has no security implications

---

## Verification

### Build & Test

- [ ] `go build ./...` — compiles with new types and imports
- [ ] `go vet ./...` — no type mismatches
- [ ] `go test ./...` — all existing tests pass
- [ ] `gofmt -s -l .` — no formatting issues
- [ ] `go mod tidy` — only `termenv` and `go-isatty` move from indirect to direct; no unrelated promotions; `go.sum` stays clean

### Manual Testing

- [ ] `NO_COLOR=1 ./bonsai catalog` → no ANSI escape sequences in output
- [ ] `./bonsai --no-color list` → no ANSI escapes
- [ ] `TERM=dumb ./bonsai catalog` → no ANSI escapes
- [ ] `./bonsai catalog > /tmp/out.txt && cat /tmp/out.txt` → no ANSI escapes in file
- [ ] Dark terminal: `./bonsai init` → identical colors to current (dark palette unchanged)
- [ ] Light terminal: `./bonsai catalog` → all text readable, darker palette variants
- [ ] `./bonsai init` (then cancel) → banner shows "agent scaffolder · v{version}"
- [ ] `./bonsai list` in dir without `.bonsai.yaml` → structured FatalPanel error
- [ ] Walk through `bonsai add` flow → no double-gaps between review panel and confirm
- [ ] All mutation commands → output ends with one blank line before shell prompt

---

## Dispatch

| Phase | Agent | Isolation | Notes |
|-------|-------|-----------|-------|
| All | general-purpose | worktree | Single pass — 9 files, no interface changes |
