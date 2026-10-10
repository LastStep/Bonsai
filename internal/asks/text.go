package asks

// The rules every free-text field of an ask and an answer goes through, in this order (design/plan-5.md, 5.2.5 note
// 2; contract §2.6, §9.1):
//
//  1. CRLF is made LF.
//  2. A hidden character is refused, not stripped, as today: Unicode's control (Cc), format (Cf), private-use (Co),
//     surrogate (Cs) and unassigned (Cn) characters, the line and paragraph separators (Zl, Zp), and the variation
//     selectors (today's rule refuses them too: an emoji written with U+FE0F is refused), with one exception: a line
//     feed in a field that may hold lines (why and an answer's words). Text that is not UTF-8 is refused the same way.
//     A tab is a control character, so it is refused too, as today.
//  3. The redactor (internal/redact's Text) takes every secret it finds out.
//  4. The limit, in characters (code points), on the redacted text: a field over it is refused, never cut.
//
// No NFC normalising, unlike today's: it needs a library outside Go's standard one, and the studio's bridge may
// normalise. A title, an option and every one-line field refuse a line feed with their own sentence.

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/redact"
)

// field is one free-text field's rules.
type field struct {
	name  string // as the command line names it: "--title"
	max   int    // its limit, in characters, after redaction
	lines bool   // it may hold line feeds (why, words)
}

// The free-text fields.
var (
	titleField  = field{name: "--title", max: TitleMax}
	whyField    = field{name: "--why", max: WhyMax, lines: true}
	thenField   = field{name: "--then", max: ThenMax}
	optionField = field{name: "--option", max: OptionMax}
	choiceField = field{name: "--choice", max: OptionMax}
	wordsField  = field{name: "--words", max: WordsMax, lines: true}
	byField     = field{name: "--by", max: ByMax}
	viaField    = field{name: "--via", max: ViaMax}
)

// selectors are the variation selectors: U+180B-U+180D, U+180F, U+FE00-U+FE0F and U+E0100-U+E01EF.
var selectors = &unicode.RangeTable{
	R16: []unicode.Range16{{Lo: 0x180b, Hi: 0x180d, Stride: 1}, {Lo: 0x180f, Hi: 0x180f, Stride: 1}, {Lo: 0xfe00, Hi: 0xfe0f, Stride: 1}},
	R32: []unicode.Range32{{Lo: 0xe0100, Hi: 0xe01ef, Stride: 1}},
}

// Hidden reports whether r is a character a free-text field refuses (step 2 above; a line feed is reported here
// too, and allowed by the fields that hold lines).
func Hidden(r rune) bool {
	// Outside the letters, marks, numbers, punctuation, symbols and separators is category C: Cc, Cf, Co, Cs and the
	// unassigned Cn (Go's tables give C as all five).
	return !unicode.In(r, unicode.L, unicode.M, unicode.N, unicode.P, unicode.S, unicode.Z) ||
		unicode.In(r, unicode.Zl, unicode.Zp, selectors)
}

// clean gives a free-text field as it is stored, or the refusal naming the rule it breaks (bad-value, exit 2).
func clean(f field, s string) (string, *Error) {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if !utf8.ValidString(s) {
		return "", badValue(f.name+" is not UTF-8 text", "give "+f.name+" as UTF-8 text")
	}
	for _, r := range s {
		switch {
		case r == '\n' && f.lines:
		case r == '\n':
			return "", badValue(f.name+" must be one line", "give "+f.name+" on one line, with no line feed")
		case Hidden(r):
			return "", badValue(fmt.Sprintf("%s holds a hidden or control character (U+%04X), which is refused, not stripped", f.name, r),
				"give "+f.name+" again without it")
		}
	}
	s = redact.Text(s)
	if n := utf8.RuneCountInString(s); n > f.max {
		return "", badValue(fmt.Sprintf("%s is %d characters once redacted, over its %d (nothing is cut)", f.name, n, f.max),
			fmt.Sprintf("shorten %s to at most %d characters", f.name, f.max))
	}
	return s, nil
}

// required is clean for a field that must be given: empty, or only spaces, is refused (missing-value).
func required(f field, s string) (string, *Error) {
	if strings.TrimSpace(s) == "" {
		return "", &Error{Code: "missing-value", Exit: ExitInput, What: f.name + " is required", Next: "give " + f.name + " its text"}
	}
	return clean(f, s)
}

// optional is clean for a field that may be left out: "" stays "" (null); only spaces is refused.
func optional(f field, s string) (string, *Error) {
	switch {
	case s == "":
		return "", nil
	case strings.TrimSpace(s) == "":
		return "", badValue(f.name+" holds only spaces", "leave "+f.name+" out, or give it text")
	}
	return clean(f, s)
}

func badValue(what, next string) *Error {
	return &Error{Code: "bad-value", Exit: ExitInput, What: what, Next: next}
}

func badFlag(what, next string) *Error {
	return &Error{Code: "bad-flag", Exit: ExitInput, What: what, Next: next}
}
