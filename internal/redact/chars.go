package redact

// The characters the rules read, and how a word is matched in any case.
//
// Whitespace is JavaScript's \s, so a name that a JavaScript reader would take with any whitespace before its `:` or
// `=` is taken here too. A line ends at \n or \r only. A word a rule names (a keyword, `authorization`, a flag word)
// matches in any case under Unicode's simple case folding, so the long s (U+017F) counts as s and the Kelvin sign
// (U+212A) as k: those two are the only characters outside ASCII that fold to an ASCII letter.
//
// Two readings of a name exist side by side. Where a name makes a value be taken out, it is read widely: case folded,
// and its characters with the long s and the Kelvin sign among them. Where a rule holds a value back (a name on the
// line below, a name glued after a quote, a lone word at the end), the name is read narrowly, ASCII only, so a value
// is held back no more often than the narrow reading allows and more is taken out, never less.

import "unicode/utf8"

const (
	longS  = '\u017f' // the long s, which simple case folding makes an s
	kelvin = '\u212a' // the Kelvin sign, which simple case folding makes a k
)

// isSpace is JavaScript's \s.
func isSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0xA0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return r >= 0x2000 && r <= 0x200A
}

// isLineBreak is a line's end: \n or \r.
func isLineBreak(r rune) bool { return r == '\n' || r == '\r' }

// isHSpace is whitespace that does not end a line.
func isHSpace(r rune) bool { return isSpace(r) && !isLineBreak(r) }

// isSpaceTab is a space or a tab: the narrow whitespace of a name on the line below.
func isSpaceTab(r rune) bool { return r == ' ' || r == '\t' }

func isLetter(r rune) bool { return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' }

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

// isAlnum is [A-Za-z0-9].
func isAlnum(r rune) bool { return isLetter(r) || isDigit(r) }

// isWord is JavaScript's \w without the u flag, [A-Za-z0-9_]: what \b reads.
func isWord(r rune) bool { return isAlnum(r) || r == '_' }

// isASCIIName is a name's character in the narrow reading: [A-Za-z0-9_.-].
func isASCIIName(r rune) bool { return isWord(r) || r == '.' || r == '-' }

// isNameRune is a name's character in the wide reading: [A-Za-z0-9_.-], the long s and the Kelvin sign.
func isNameRune(r rune) bool { return isASCIIName(r) || r == longS || r == kelvin }

// isFlagRune is a character of a flag's body: [A-Za-z0-9-], in the wide reading also the long s and the Kelvin sign.
func isFlagRune(r rune, wide bool) bool {
	return isAlnum(r) || r == '-' || wide && (r == longS || r == kelvin)
}

func isQuote(b byte) bool { return b == '"' || b == '\'' }

// runeAt gives the character at byte i and its length, or (-1, 0) at the end. A byte that is not valid UTF-8 is a
// character of its own (utf8.RuneError, length 1), which no rule takes as a name or as whitespace.
func runeAt(s string, i int) (rune, int) {
	if i >= len(s) || i < 0 {
		return -1, 0
	}
	if c := s[i]; c < utf8.RuneSelf {
		return rune(c), 1
	}
	return utf8.DecodeRuneInString(s[i:])
}

// runeBefore gives the character ending at byte i and its length, or (-1, 0) at the start.
func runeBefore(s string, i int) (rune, int) {
	if i <= 0 || i > len(s) {
		return -1, 0
	}
	if c := s[i-1]; c < utf8.RuneSelf {
		return rune(c), 1
	}
	return utf8.DecodeLastRuneInString(s[:i])
}

// foldIs reports whether r is the lower-case ASCII c under simple case folding (wide), or ignoring ASCII case only.
func foldIs(r rune, c byte, wide bool) bool {
	if r == rune(c) {
		return true
	}
	if c >= 'a' && c <= 'z' {
		if r == rune(c-'a'+'A') {
			return true
		}
		if wide && (c == 's' && r == longS || c == 'k' && r == kelvin) {
			return true
		}
	}
	return false
}

// matchWord reports whether word (lower-case ASCII) starts at byte i of s, and where it ends.
func matchWord(s string, i int, word string, wide bool) (int, bool) {
	for k := 0; k < len(word); k++ {
		r, n := runeAt(s, i)
		if n == 0 || !foldIs(r, word[k], wide) {
			return 0, false
		}
		i += n
	}
	return i, true
}

// skipSpace gives the first byte at or after i that is not whitespace (JavaScript's \s*).
func skipSpace(s string, i int) int {
	for i < len(s) {
		r, n := runeAt(s, i)
		if !isSpace(r) {
			break
		}
		i += n
	}
	return i
}

// skipHSpace is skipSpace without crossing a line's end.
func skipHSpace(s string, i int) int {
	for i < len(s) {
		r, n := runeAt(s, i)
		if !isHSpace(r) {
			break
		}
		i += n
	}
	return i
}

// The secret words a key's name holds (`password`, `DB_TOKEN`, `x-api-key`, `aws_secret_access_key`): this list is
// their one home. `credentials` holds `credential`, and `apikey` is `api` and `key` with nothing between.
var keywords = []string{
	"password", "passwd", "passphrase", "pwd", "secret", "token",
	"apikey", "api_key", "api-key", "accesskey", "access_key", "access-key",
	"authkey", "auth_key", "auth-key", "privatekey", "private_key", "private-key",
	"credential",
}

// hasKeyword reports whether s[a:b] holds a secret word.
func hasKeyword(s string, a, b int, wide bool) bool {
	for i := a; i < b; {
		r, n := runeAt(s, i)
		for _, w := range keywords {
			if !foldIs(r, w[0], wide) {
				continue
			}
			if e, ok := matchWord(s, i, w, wide); ok && e <= b {
				return true
			}
		}
		i += n
	}
	return false
}

// The words a flag may end in (`--password`, `--client-secret`, `-token`): their one home.
var flagWords = [...]string{"password", "passwd", "token", "secret", "api-key", "apikey", "auth-token"}

// flagAt reads a flag starting at byte i (its first `-`): `-` or `--`, then words of [A-Za-z0-9] each ending in `-`,
// then a flag word, then a character follow accepts. It gives where the flag ends. It does not look before i.
func flagAt(s string, i int, wide bool, follow func(rune) bool) (int, bool) {
	fr := flagReader{s: s, wide: wide, follow: follow}
	return fr.at(i)
}

// flagReader reads the flags of one text from left to right. Every `-` in one run of flag characters shares that
// run's end, the flag words it ends in and where the last `--` before each stands, so a run is read once however
// many `-` it holds (in the wide reading, the long s and `-` repeated have a `-` that may start a flag every two
// characters, since the long s is a flag's character but not one a flag may follow).
type flagReader struct {
	s      string
	wide   bool
	follow func(rune) bool

	from, end int                 // the run last read: from its first `-` read, to its end
	ok        bool                // whether follow accepts the character at end
	start     [len(flagWords)]int // where each flag word starts at the run's end, or -1
	dd        [len(flagWords)]int // the start of the last `--` that ends before start, or -1
}

// at is flagAt for the reader's text; calls come in ascending order of i.
func (fr *flagReader) at(i int) (int, bool) {
	s := fr.s
	if i >= len(s) || s[i] != '-' {
		return 0, false
	}
	j := i + 1
	if j < len(s) && s[j] == '-' {
		j++
	}
	if i < fr.from || i >= fr.end {
		fr.read(i)
	}
	if !fr.ok {
		return 0, false
	}
	for w := range flagWords {
		start := fr.start[w]
		if start < j {
			continue
		}
		// What comes before the flag word: nothing, or words of [A-Za-z0-9] each ending in one `-`.
		if start == j || s[j] != '-' && s[start-1] == '-' && fr.dd[w] < j {
			return fr.end, true
		}
	}
	return 0, false
}

// read reads the run of flag characters holding the `-` at i.
func (fr *flagReader) read(i int) {
	s := fr.s
	k := i + 1
	for k < len(s) {
		r, n := runeAt(s, k)
		if !isFlagRune(r, fr.wide) {
			break
		}
		k += n
	}
	fr.from, fr.end = i, k
	r, n := runeAt(s, k)
	fr.ok = n > 0 && fr.follow(r)
	for w, word := range flagWords {
		fr.start[w], fr.dd[w] = -1, -1
		start, ok := suffixWord(s, i, k, word, fr.wide)
		if !ok {
			continue
		}
		fr.start[w] = start
		for p := start - 2; p >= i; p-- {
			if s[p] == '-' && s[p+1] == '-' {
				fr.dd[w] = p
				break
			}
		}
	}
}

// suffixWord reports whether s[a:b] ends with word (lower-case ASCII, matched in any case), and where it starts.
func suffixWord(s string, a, b int, word string, wide bool) (int, bool) {
	i := b
	for k := len(word) - 1; k >= 0; k-- {
		if i <= a {
			return 0, false
		}
		r, n := runeBefore(s, i)
		if !foldIs(r, word[k], wide) {
			return 0, false
		}
		i -= n
	}
	if i < a {
		return 0, false
	}
	return i, true
}
