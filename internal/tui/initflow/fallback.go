// Package initflow implements the cinematic 4-stage `bonsai init` flow.
//
// The package is intentionally self-contained — its primitives (chrome,
// enso rail, stage base) are only used by the redesigned init path behind
// the BONSAI_REDESIGN env flag while the legacy flow remains default.
//
// fallback.go owns wide-character detection + the stage-label mapping so
// stage primitives can render kanji/kana on safe terminals and ASCII-only
// labels on constrained ones.
package initflow

import (
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
)

// WideCharSafe reports whether the current terminal can reliably render
// 2-wide CJK characters for the enso-rail kanji. The check is deliberately
// conservative — on terminals where we cannot prove wide-char support we
// fall back to the bracketed ASCII rail.
//
// Order of checks:
//  1. BONSAI_ASCII_ONLY=1 — explicit opt-out always wins.
//  2. runewidth reports East-Asian locale — trust the library.
//  3. Windows Terminal (WT_SESSION) — modern conhost handles CJK correctly.
//  4. Allow-listed unix terminals by $TERM prefix.
//  5. UTF-8 locale ($LC_ALL/$LANG) as a loose signal otherwise.
//
// The fall-through returns false only when no positive signal is found —
// better to render ASCII on a rich terminal than misaligned kanji on a
// broken one.
func WideCharSafe() bool {
	if os.Getenv("BONSAI_ASCII_ONLY") == "1" {
		return false
	}
	if runewidth.EastAsianWidth {
		return true
	}
	if os.Getenv("WT_SESSION") != "" {
		return true
	}
	term := os.Getenv("TERM")
	goodTerms := []string{"xterm", "screen", "tmux", "alacritty", "kitty", "wezterm", "ghostty"}
	for _, good := range goodTerms {
		if strings.HasPrefix(term, good) {
			return true
		}
	}
	lang := os.Getenv("LC_ALL")
	if lang == "" {
		lang = os.Getenv("LANG")
	}
	return strings.Contains(strings.ToLower(lang), "utf")
}

// StageLabel is the triple-label (kanji + kana + English) attached to each
// of the four init-flow stages. Render picks between the kanji/kana form
// (when the terminal is wide-char safe) and ASCII-only fallback.
//
// Stage labels (design-locked):
//
//	器 うつわ Vessel
//	土 つち  Soil
//	枝 えだ  Branches
//	観 みる  Observe
type StageLabel struct {
	Kanji   string
	Kana    string
	English string
}

// Render returns (primary, secondary) display strings for this stage label.
// On wide-safe terminals primary is "<kanji> <English>" and secondary is
// the kana reading. On ASCII-only terminals primary is English-only and
// secondary is empty.
func (l StageLabel) Render(safe bool) (primary, secondary string) {
	if safe {
		return l.Kanji + " " + l.English, l.Kana
	}
	return l.English, ""
}

// StageLabels holds the four canonical init-flow stage labels in order.
// Both the enso rail renderer and stage constructors pull from this table
// so the text appears in exactly one place.
var StageLabels = [4]StageLabel{
	{Kanji: "器", Kana: "うつわ", English: "VESSEL"},
	{Kanji: "土", Kana: "つち", English: "SOIL"},
	{Kanji: "枝", Kana: "えだ", English: "BRANCHES"},
	{Kanji: "観", Kana: "みる", English: "OBSERVE"},
}
