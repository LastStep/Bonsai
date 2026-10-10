package redact

// The names a value follows, found all at once before any value is read.
//
// A name is one of:
//   - a key: a run of name characters holding a secret word (`password`, `DB_TOKEN`, `x-api-key`), then an optional
//     quote, any whitespace, `:` or `=`, any whitespace (`password: x`, `"apiKey": "x"`, `TOKEN=x`);
//   - a flag: `-` or `--`, words of [A-Za-z0-9] each ending in `-`, a flag word (`--password`, `--client-secret`),
//     starting a word (not after [A-Za-z0-9_-]), then whitespace on its line;
//   - an Authorization header: `authorization` anywhere (`Proxy-Authorization`, `x-authorization`), any whitespace,
//     `:` or `=`, any whitespace; its value's quote and scheme word are read with the value;
//   - a Bearer: `bearer` starting a word, then whitespace;
//   - `extraheader`, any whitespace, `=`, any whitespace (git's http.extraHeader).
//
// Each is read widely (chars.go): any whitespace JavaScript's \s matches, and every word in any case under simple
// case folding. A name may stand inside another name's value; it is found all the same, and its value is taken.

type nameKind uint8

// The kinds in the studio's rule order.
const (
	nameExtra nameKind = iota
	nameAuth
	nameBearer
	nameFlag
	nameKey
)

var nameKinds = [...]Kind{
	nameExtra:  KindExtraHeader,
	nameAuth:   KindAuthorization,
	nameBearer: KindBearer,
	nameFlag:   KindFlag,
	nameKey:    KindKey,
}

// name is one name: where it starts, and where its value may start (past its separator and the whitespace after
// it; for an Authorization header, before its optional quote and scheme word).
type name struct {
	kind  nameKind
	start int
	val   int
}

// findNames gives every name in s, in order of where it starts.
func findNames(s string) []name {
	var out []name
	n := len(s)
	prev := rune(-1)
	flags := flagReader{s: s, wide: true, follow: isHSpace}
	flagEnd, flagVal := -1, 0 // every flag in one run ends where the run does: its value's start is read once
	for i := 0; i < n; {
		r, size := runeAt(s, i)
		if isNameRune(r) {
			if !isNameRune(prev) {
				e := i
				for e < n {
					rr, nn := runeAt(s, e)
					if !isNameRune(rr) {
						break
					}
					e += nn
				}
				if hasKeyword(s, i, e, true) {
					if v, ok := separator(s, e, true, ":="); ok {
						out = append(out, name{nameKey, i, v})
					}
				}
			}
			switch {
			case r == '-':
				if !isASCIIFlagLead(prev) {
					if e, ok := flags.at(i); ok {
						if e != flagEnd {
							flagEnd, flagVal = e, skipHSpace(s, e)
						}
						out = append(out, name{nameFlag, i, flagVal})
					}
				}
			case foldIs(r, 'b', true):
				if !isWord(prev) {
					if e, ok := matchWord(s, i, "bearer", true); ok {
						if rr, _ := runeAt(s, e); isSpace(rr) {
							out = append(out, name{nameBearer, i, skipSpace(s, e)})
						}
					}
				}
			case foldIs(r, 'a', true):
				if e, ok := matchWord(s, i, "authorization", true); ok {
					if v, ok := separator(s, e, false, ":="); ok {
						out = append(out, name{nameAuth, i, v})
					}
				}
			case foldIs(r, 'e', true):
				if e, ok := matchWord(s, i, "extraheader", true); ok {
					if v, ok := separator(s, e, false, "="); ok {
						out = append(out, name{nameExtra, i, v})
					}
				}
			}
		}
		prev = r
		i += size
	}
	return out
}

// isASCIIFlagLead is what a flag may not follow: [A-Za-z0-9_-].
func isASCIIFlagLead(r rune) bool { return isWord(r) || r == '-' }

// separator reads, from e, an optional quote (when quote is true), any whitespace, one of seps, any whitespace, and
// gives where a value may start.
func separator(s string, e int, quote bool, seps string) (int, bool) {
	if quote && e < len(s) && isQuote(s[e]) {
		e++
	}
	e = skipSpace(s, e)
	if e >= len(s) || (s[e] != seps[0] && (len(seps) < 2 || s[e] != seps[1])) {
		return 0, false
	}
	return skipSpace(s, e+1), true
}
