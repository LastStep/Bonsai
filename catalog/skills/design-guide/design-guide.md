# Design Guide

Bonsai CLI design rules. Applies to `internal/tui/**` and `cmd/*.go`.

---

## 1. Design Principles

- **Respect the Terminal** — render correctly on dark, light, NO_COLOR, TERM=dumb, and piped output. Never assume a dark background.
- **One Voice** — every command speaks through the same vocabulary: headings, panels, helpers. Don't invent new shapes.
- **Progressive Disclosure** — show the minimum first. Hints, details, and trees come after the outcome, not before.
- **Feedback is Mandatory** — every mutation ends with `tui.Success(...)` (or a panel that clearly states the outcome).
- **Understated Personality** — muted palette, rounded borders, no emoji, no marketing copy.

## 2. Adaptive Palette

Nine semantic tokens live in `internal/tui/styles.go`. All are `lipgloss.AdaptiveColor` typed as `lipgloss.TerminalColor`.

| Token | Role |
|-------|------|
| `Leaf` | primary brand / success accent |
| `Bark` | labels, headings, table headers |
| `Stone` | muted text, borders, tree branches |
| `Water` | informational accents, review panels |
| `Moss` | success panels and checkmarks |
| `Ember` | errors, fatal panels |
| `Amber` | warnings, removal panels |
| `Sand` | option text, file leaves |
| `Petal` | selection markers, prompt cursors |

**Never** write `lipgloss.Color("#...")` in new code. If a one-off adaptive value is unavoidable (e.g., a theme override), inline it as `lipgloss.AdaptiveColor{Dark: "#...", Light: "#..."}` — never hardcode a single hex.

## 3. Glyph Set

Six glyphs, defined as constants in `styles.go`:

`GlyphCheck` `✓` · `GlyphCross` `✗` · `GlyphWarn` `⚠` · `GlyphArrow` `→` · `GlyphDash` `—` · `GlyphDot` `·`

No emoji. No additional Unicode decorations. If you need a new glyph, justify it in the PR.

## 4. Panel Vocabulary

| Helper | When |
|--------|------|
| `Success` / `SuccessPanel` | terminal success message at end of a command |
| `ErrorPanel` | recoverable error — command returns `nil` and user can retry |
| `FatalPanel(title, detail, hint)` | unrecoverable error — calls `os.Exit(1)`; use for missing config, catalog corruption, argument validation |
| `WarningPanel` / `Warning` | non-blocking issue (skipped file, ignored flag) |
| `InfoPanel` / `Info` | contextual information, not a result |
| `EmptyPanel` | "no items" state in list/catalog views |
| `TitledPanel(title, content, color)` | review/summary block before a confirm prompt |

## 5. Spacing Contract

- **Display helpers own their top margin.** Every `Success`, `Error`, `Heading`, `*Panel`, `TitledPanel` starts with `\n`.
- **Commands own exactly one trailing `tui.Blank()`** on success paths — after `tui.Success(...)` and before `return nil`.
- **Do not add `tui.Blank()` between a `TitledPanel` and an `AskConfirm`.** The panel already owns its gap and Huh owns its own.
- **Cancelled paths** (`!confirmed` → `return nil`) and **guard paths** (`ErrorPanel` + `return nil`) do not emit a trailing `Blank` — the rendered panel already closes the output.

## 6. Canonical Command Flow

```
Heading → Input (AskText / AskSelect / PickItems)
       → Review (TitledPanel)
       → Confirm (AskConfirm)
       → Execute (spinner)
       → Results (showWriteResults / trees)
       → Success (+ optional Hint)
       → Blank
```

Every mutating command (`init`, `add`, `remove`, `update`) follows this shape. Read-only commands (`list`, `catalog`) skip Confirm/Execute.

## 7. Error Format

- **Fatal (exit 1):** `tui.FatalPanel(title, detail, hint)` — title is a short noun phrase, detail explains the cause, hint is one actionable line (usually a command to run).
- **Recoverable:** `tui.ErrorPanel(msg)` followed by `return nil` — the command exits cleanly, the user can fix and retry.
- **Never** use `fmt.Println`, `fmt.Errorf`, or bare `log.Fatal` for user-facing errors. Never surface raw Go error strings without a panel.

## 8. Anti-Patterns

- Hardcoded `lipgloss.Color("#hex")` — breaks light terminals.
- Emoji in any output — breaks terminal fidelity and tone.
- Double blanks between elements (`Blank()` immediately after a `Panel` that already owns its margin).
- `os.Exit(1)` without a preceding `FatalPanel` — the user sees no explanation.
- Raw `error.Error()` strings printed to stdout — wrap them in a panel with a hint.
- Inventing new panel shapes or colors instead of reusing the eight semantic helpers above.
