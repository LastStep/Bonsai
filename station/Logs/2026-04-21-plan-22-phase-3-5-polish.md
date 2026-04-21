---
tags: [session-log, plan-22, uiux]
description: Plan 22 Phase 3.5 — dogfood polish pass on `bonsai init` cinematic redesign (direct-to-main, no PR).
---

# 2026-04-21 — Plan 22 Phase 3.5 polish pass

## Context

User dogfooded `BONSAI_REDESIGN=1 bonsai init` after Phase 3 shipped (PR #49). Screenshots compared against design mockup. Many visual issues surfaced — spacing, font hierarchy, colour, centring. Iterated directly in main tree (no PR, no worktree) per durable UX-preference memory: "fast iteration beats process for UX work" / "test locally" / "pick scope pragmatically, foundations first".

## Changes landed

### Palette (`internal/tui/styles.go`)
- `Bark` dark `#D4A76A` → `#D4AF37` (canonical gold, less orange). Light `#8B6914` → `#7A5E10`.
- New `Moon` token `#F5F5F5`/`#1A1A1A` — semantic white.
- `ColorAccent` remapped from `Petal` (pink) → `Moon` — bold emphasis + interactive chrome now white.

### Rail chrome (`enso.go`)
- Dropped row 3 (kana under current stage) entirely — user directive "remove katakana everywhere, keep kanji only".
- Current-stage anchor `[器]` now Bark-gold bold (was green + dim brackets).
- Done-stage anchor `●` now bright Primary green (was `ColorLeafDim`) — user wanted "nice green dot" on completed checkpoints.
- Rail total width capped at `maxRail = 60` cells (was full-width); centred inside terminal via computed `leftPad = (width - railWidth) / 2`. Checkpoints now sit tight + visually balanced rather than sprawled.

### Stage title + divider (`vessel.go`, `soil.go`)
- Title simplified from `器 VESSEL  うつわ  ·  VESSEL` / `土 SOIL  つち  ·  SOIL` → `器 VESSEL` / `土 SOIL`. Kanji Bark-gold, English white-bold. Drops redundant kana append + trailing `· VESSEL/SOIL`.
- Divider re-tinted: green `───` + Bark `FIELDS`/`SCAFFOLDING` + long dim `─` rule (55 cells). Dropped middle `入力`/`足場` kana segment.

### Field-row layout (`vessel.go`)
- 2-column grid: LEFT (LABEL bold-gold / subtitle dim) + RIGHT (`❯ input` / underline). Dropped 3rd verbose hint line entirely (`used as .bonsai.yaml project_name`, etc.) as dev-speak noise.
- Underline under each input: dim Rule2 at rest, bright Primary green on focus — gives clear "which field am I in" signal without ANSI background tricks.
- `labelColW` 14 → 20, `inputW` 48 → 60 for wider form span.
- Input text style now white-bold (was Bark-gold) — reads as active value, not label/helper.
- Placeholder style `ColorMuted` → `ColorRule2` — dimmer/greyer per user ask.

### Copy refresh
- Vessel paragraph: `Every Bonsai begins with a small decision — what will this one carry?` → `Three quick answers — a name, a purpose, a place to grow.`
- Soil headline: `Choose what the project carries.` → `Tend the soil.` (mirrors Vessel's `Shape the vessel.` pattern).
- Subtitle refresh: `required` / `optional` / `default station/` (dropped `where files live` — less self-explanatory).

### Footer (`chrome.go`)
- Added full-width muted rule (ColorRule2) above brand/hints row — subtle separator between stage body and footer per user request.
- New `centerBlock(block, width)` helper — pads every non-blank line so widest line is centred in terminal width.

### Frame pad (`stage.go`)
- Pad math adjusted `-4` → `-5` blank separators. Footer went from 1 row → 2 rows (rule + hints), so AltScreen fills to terminal bottom correctly.

## Gotcha — bubbles/textinput View() width asymmetry

User reported form still shifting left after each keystroke. Tracked down to:

- `textinput.View()` renders **`Width`** cells when value is empty (placeholder mode, `placeholderView()` pads to `m.Width - promptW - cursorW`).
- Renders **`Width + 3`** cells when typed (prompt + value + cursor + internal padding to `m.Width`).

My first fix used `padRight(view, inputW+2)` which early-returns unchanged when `cur >= w`, so empty was padded to 62 but typed stayed at 63. Line1 width fluctuated → `centerBlock` recomputed `maxW` → body shifted.

**Proper fix:** `lipgloss.PlaceHorizontal(inputCellW, lipgloss.Left, input.View())` with `inputCellW = inputW + 4`. Pads *or truncates* to exactly `inputCellW` cells regardless of input state. Underline rule length matched to same 64 cells.

Worth a memory note (added to `notes`) — other `bubbles/textinput` consumers in the codebase (Huh-wrapped inputs in `prompts.go`) likely have the same latent issue when embedded in any centred/fixed-width frame. Not fixing preemptively (no reports elsewhere); revisit if Phase 4 Branches picker or Observe stage adds new centred textinputs.

## Verification

- `go build ./...` clean across all packages
- `go test ./...` green (full suite, including existing `fallback_test.go` with `secondary == "うつわ"` — preserved by leaving `StageLabel.Render` contract untouched; kana simply no longer consumed by callers)
- `make build` — binary rebuilds to `./bonsai`
- Manual dogfood: `BONSAI_REDESIGN=1 ./bonsai init` — user confirmed "looks great" after iteration loop

## Sequencing decision

Chose to ship Phase 3.5 polish *before* Phase 4 (Branches picker) because the changes touch shared chrome (rail, footer, centerBlock) that Phase 4 will inherit. Stacking Branches on top of broken foundations = double rework. User agreed with "foundations first" framing.

## Carry-forward

- Phase 4 next: Branches tabbed picker + inline-expand per `zen-shell.jsx` ZStepBranches. Chrome + palette + title/divider pattern now settled — Phase 4 just slots its picker body into the established frame.
- `lipgloss.PlaceHorizontal` pattern should be used in Observe + any future stage with textinputs (Phase 5 Planted body may have a `bubbles/spinner` which is width-stable, no fix needed).
