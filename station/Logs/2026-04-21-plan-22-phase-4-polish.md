---
date: 2026-04-21
plan: 22
phase: Phase 4 polish run
session_type: dogfood-driven direct-to-main iteration
---

# Plan 22 — Phase 4 Branches polish run

Five direct-to-main commits against the Branches stage, all UX-polish driven by real-time dogfood feedback. No PRs per fast-iter memory ("skip PR flow for taste-heavy design work until visual direction is settled").

## Commits shipped

| Commit | Subject | Why |
|--------|---------|-----|
| `413e360` | polish: details panel + header split + tab intros + row alignment | Initial iteration — move `?`-expand out of the focused row into a fixed-height box below list; stack `[盆] BONSAI / INITIALIZE` on two rows; add per-tab 2-line intro; right-align DEFAULT via `lipgloss.PlaceHorizontal(tagColW, Right, tag)` |
| `399fe08` | polish: density cut + word-wrapped details + centered tab counts | User: "visual attack" → narrow row to 60 cells, word-wrap ABOUT 2 rows × 60, center tab N/Total via `PlaceHorizontal(Center)` |
| `eaee416` | polish: widen Branches layout + strict name-column truncation | User screenshot showed "Issue To Implementation" (22 chars) overflowing nameColW=16 shoving DEFAULT right. Widened to row=84 cells (nameColW 24, descColW 44, tab colW 16, divider trails 60). Added rune-aware name truncation as a belt-and-braces guard. |
| `6bb74e5` | polish: white ABOUT/FILE values + 3-row wrap fits long descriptions | User saw dispatch-guard description clipped: "…before…". 2×60=120 cells wasn't enough for ~111-char descriptions that wrap into 3 lines. Bumped to 3 rows × 70 cells = 210 cells. ABOUT + FILE values switched from ColorSubtle/ColorRule2 → ColorAccent (white) for legibility against dim surround. |
| `fa0ae64` | polish: center kanji inside header brackets + extra blank line after DETAILS | Terminals left-anchor CJK inside their 2-cell slot, so `[盆]` reads off-center. Padded to `[ 盆 ]` (`[ o ]` for ASCII fallback). Extra blank before counter so DETAILS doesn't bleed into summary. |

## New code

- **`wrapToWidth(text, width)`** — word-break helper with rune-fallback hard-wrap for oversized tokens. Added to `internal/tui/initflow/branches.go` bottom. Returns at least one line; empty input → `[""]`.

## Dimensions (final state)

| Element | Cells |
|---------|-------|
| Row total | 84 (border 2 + glyph 1 + space 1 + name 24 + space 1 + desc 44 + space 1 + tag 10) |
| Tab col | 16 (tab row = 5×16 + 4 spaces = 84) |
| Divider trail rule | 60 (after `─── CATEGORIES `) |
| DETAILS panel | 5 fixed lines (header + 3 about + 1 file) |
| ABOUT contentW | 70 cells × 3 rows = 210 cells absorbed |
| FILE truncation | tail-preserving leading `…` at contentW=70 |

## Gotchas learned

1. **Rune-level truncation matters.** Raw `padRight(name, nameColW)` only pads if shorter — oversized names still overflow and shift downstream columns. Always pair with an explicit rune-aware truncate when column discipline is required.
2. **Fixed-height panels prevent viewport jitter.** Any block that toggles between "show" and "hide" states (like `?`-expand) must render the same number of visible lines in both states, or AltScreen flickers and the counter below hops by a row. Reserve rows eagerly, fill with blank indent.
3. **CJK glyphs are left-anchored in their 2-cell slot.** A tight `[盆]` looks off-center because the wide-char draws flush-left within its reserved 2 cells. Padding `[ 盆 ]` reads balanced in every terminal tested.
4. **`lipgloss.PlaceHorizontal` handles ANSI widths correctly.** For column-centering of styled content, pre-render the style then call `PlaceHorizontal(colW, Center, styled)` — better than manual pad math.

## What did NOT change

- Tests (`branches_test.go`) — all 12 still green. They're state-oriented (tab cycle / focus clamp / toggle / Result shape / Reset preservation / defaults), so column-width changes don't touch them.
- Behaviour or keybinds.
- Plan 22 `.md` — plan was already written with ABOUT + FILE only (locked decision 5). No scope creep.

## Next

Phase 5: Observe stage + Generate one-frame reveal + full-screen Planted + flip `BONSAI_REDESIGN=1` default + delete legacy init + docs. Only phase touching `cmd/init.go` flag routing.
