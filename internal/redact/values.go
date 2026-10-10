package redact

// Each name's value, read after every name is found.
//
// The values are read from the last name to the first, and each name's value is read twice:
//   - on the text as it is, as a rule that meets the name first reads it;
//   - on the text with the values already taken (those of the names after it) standing as markers. A name inside
//     another's value has had its own value taken first, so the outer value reads that marker as its own and takes
//     the inner name too: `a bearer db_password=hunter2 ok` becomes `a bearer [redacted] ok`, one marker over both,
//     as the studio's second sweep writes it.
//
// Both readings are taken out, and spans that overlap or touch become one. So no name's value is left because
// another value reached it first, and wherever a name stands, its value goes.
//
// A value's shapes, by name (the studio's, with its three leak classes closed):
//   - a quoted value ends at its closing quote on its line and takes everything glued after the quote, a name too (a
//     shell reads `"a b"c` as one word; a marker left glued to a name would change when a cut shortened the name); a
//     quote that never closes on its line runs to the line's end, and so does one whose next quote opens another
//     key's value (`PASSWORD: "x ... TOKEN="a b"`);
//   - a key's bare value stops at whitespace, a quote, `,` `;` `&` `}` `)`; a flag's and an Authorization header's at
//     whitespace or a quote, and a flag's never starts with `-`; a Bearer's is 8 or more of [A-Za-z0-9._~+/=-];
//     extraheader's is everything to the next whitespace;
//   - a value may sit on a later line than its name (`password:`, then `  hunter2`), except when that line starts
//     with a name on its own line (`secrets:`, then `  password: x`: the name below has the value), or is the last
//     line of the text and one word (it may be a name the recorder's cut left: `secrets:`, then `  passw`);
//   - after an Authorization header, its scheme word (Basic, Bearer, Token, Digest) is the header's own and the
//     value follows it, quoted or not (a quoted token takes the scheme word with it, and when its quote does not
//     close on its line, only the word it opens); `token` followed by `:` or `=` is a key, not the scheme, so the
//     header takes it as its value (with what follows its separator, as after a scheme word) and the key's value
//     goes too;
//   - a value that is only a cut tail of the marker (`password=[reda`), or after an Authorization header of a scheme
//     word (`Authorization: Bea`), with nothing after it but whitespace, is kept: it is what a cut of the redactor's
//     own output leaves, and a second pass must leave it.

import (
	"sort"
	"strings"
)

// cls is a value's character class.
type cls uint8

const (
	clsKey       cls = iota // a key's bare value: [^\s"',;&})]
	clsQuoteless            // a flag's or an Authorization header's: [^\s"']
	clsBearer               // a Bearer's: [A-Za-z0-9._~+/=-]
	clsNonSpace             // extraheader's: \S
	nCls
)

func inCls(c cls, r rune) bool {
	switch c {
	case clsKey:
		return !isSpace(r) && r != '"' && r != '\'' && r != ',' && r != ';' && r != '&' && r != '}' && r != ')'
	case clsQuoteless:
		return !isSpace(r) && r != '"' && r != '\''
	case clsBearer:
		return isAlnum(r) || r == '.' || r == '_' || r == '~' || r == '+' || r == '/' || r == '=' || r == '-'
	default:
		return !isSpace(r)
	}
}

// span is a value taken, in the text's byte offsets.
type span struct {
	a, b int
	kind Kind
}

// state is the reading of one text's values.
type state struct {
	s     string
	n     int
	stack []span // the values taken so far, merged, from the rightmost (index 0) to the leftmost (last)

	stop   [nCls][]int32 // stop[c][i]: the first byte at or after i not in class c (built when first needed)
	nextQ  [2][]int32    // the first `"` (0) or `'` (1), or line end, at or after i
	nextLB []int32       // the first line end at or after i

	runFrom map[int]run // a narrow run of name characters from a byte
	runTo   map[int]run // a narrow run of name characters ending at a byte

	lastNonSpace int // the last byte that starts a character that is not whitespace, or -1
	loneFrom     int // from here to the end the text is one narrow word, a quote, whitespace
}

type run struct {
	start, end int
	keyword    bool
}

// nameValues gives the spans of s that the values of names (in order of where they start) take, in order.
func nameValues(s string, names []name) []Span {
	if len(names) == 0 {
		return nil
	}
	st := newState(s)
	for i := len(names) - 1; i >= 0; i-- {
		nm := names[i]
		kind := nameKinds[nm.kind]
		var got [2]span
		k := 0
		a, b, ok := st.value(nm, false)
		reach := st.lineEnd(nm.val)
		if ok {
			got[k] = span{a, b, kind}
			k++
			reach = max(reach, b, st.lineEnd(a))
		}
		// The second reading, with the values already taken standing as markers, can differ only where one of
		// them lies within what the first reading looked at.
		if l := len(st.stack); l > 0 && st.stack[l-1].a <= reach {
			if a, b, ok := st.value(nm, true); ok {
				got[k] = span{a, b, kind}
				k++
			}
		}
		for _, g := range got[:k] {
			st.add(g)
		}
	}
	out := make([]Span, 0, len(st.stack))
	for i := len(st.stack) - 1; i >= 0; i-- {
		sp := st.stack[i]
		out = append(out, Span{sp.a, sp.b, sp.kind})
	}
	return out
}

func newState(s string) *state {
	st := &state{s: s, n: len(s), runFrom: map[int]run{}, runTo: map[int]run{}, lastNonSpace: -1}
	for i := len(s); i > 0; {
		r, n := runeBefore(s, i)
		if !isSpace(r) {
			st.lastNonSpace = i - n
			break
		}
		i -= n
	}
	i := skipBackHSpace(s, len(s))
	if i > 0 && isQuote(s[i-1]) {
		i--
	}
	for i > 0 {
		r, n := runeBefore(s, i)
		if !isASCIIName(r) {
			break
		}
		i -= n
	}
	st.loneFrom = i
	return st
}

func skipBackHSpace(s string, i int) int {
	for i > 0 {
		r, n := runeBefore(s, i)
		if !isHSpace(r) {
			break
		}
		i -= n
	}
	return i
}

// add takes a span out, joining it with any it overlaps or touches; the joined span keeps its leftmost part's kind.
func (st *state) add(sp span) {
	if sp.b <= sp.a {
		return
	}
	stk := st.stack
	lo := sort.Search(len(stk), func(i int) bool { return stk[i].a <= sp.b })
	hi := sort.Search(len(stk), func(i int) bool { return stk[i].b < sp.a })
	if lo >= hi {
		st.stack = append(stk, span{})
		copy(st.stack[lo+1:], st.stack[lo:])
		st.stack[lo] = sp
		return
	}
	m := span{min(sp.a, stk[hi-1].a), max(sp.b, stk[lo].b), sp.kind}
	if stk[hi-1].a < sp.a {
		m.kind = stk[hi-1].kind
	}
	st.stack = append(append(stk[:lo], m), stk[hi:]...)
}

// spanFrom gives the index of the leftmost span starting at or after pos, or -1.
func (st *state) spanFrom(pos int) int {
	stk := st.stack
	return sort.Search(len(stk), func(i int) bool { return stk[i].a < pos }) - 1
}

// spanAt gives the index of the span starting at pos, or -1.
func (st *state) spanAt(pos int) int {
	if k := st.spanFrom(pos); k >= 0 && st.stack[k].a == pos {
		return k
	}
	return -1
}

// spanEndIn reports whether a span ends in [lo, hi].
func (st *state) spanEndIn(lo, hi int) bool {
	stk := st.stack
	k := sort.Search(len(stk), func(i int) bool { return stk[i].b < lo }) - 1
	return k >= 0 && stk[k].b <= hi
}

func (st *state) stopOf(c cls) []int32 {
	if st.stop[c] != nil {
		return st.stop[c]
	}
	a := make([]int32, st.n+1)
	a[st.n] = int32(st.n)
	start := 0
	for i := 0; i < st.n; {
		r, size := runeAt(st.s, i)
		if !inCls(c, r) {
			for k := start; k <= i; k++ {
				a[k] = int32(i)
			}
			start = i + size
		}
		i += size
	}
	for k := start; k < st.n; k++ {
		a[k] = int32(st.n)
	}
	st.stop[c] = a
	return a
}

func (st *state) nextOf(q byte) []int32 {
	k := 0
	if q == '\'' {
		k = 1
	}
	if st.nextQ[k] == nil {
		st.nextQ[k] = nextByte(st.s, func(c byte) bool { return c == q || c == '\n' || c == '\r' })
	}
	return st.nextQ[k]
}

func nextByte(s string, hit func(byte) bool) []int32 {
	a := make([]int32, len(s)+1)
	a[len(s)] = int32(len(s))
	for i := len(s) - 1; i >= 0; i-- {
		if hit(s[i]) {
			a[i] = int32(i)
		} else {
			a[i] = a[i+1]
		}
	}
	return a
}

// lineEnd gives where the line holding p ends (a value never holds a line end, so the same in both readings).
func (st *state) lineEnd(p int) int {
	if st.nextLB == nil {
		st.nextLB = nextByte(st.s, func(c byte) bool { return c == '\n' || c == '\r' })
	}
	if p >= st.n {
		return st.n
	}
	return int(st.nextLB[p])
}

// classEnd gives the end of the run of class c from p. With virt, a value taken stands as a marker, which every
// class but the Bearer's takes whole.
func (st *state) classEnd(c cls, p int, virt bool) int {
	stop := st.stopOf(c)
	if !virt {
		return int(stop[p])
	}
	i, k := p, st.spanFrom(p)
	for {
		e := int(stop[i])
		if k < 0 || e < st.stack[k].a {
			return e
		}
		if c == clsBearer {
			return st.stack[k].a
		}
		i = st.stack[k].b
		k--
	}
}

// closeQuote finds the next quote q on the line from p.
func (st *state) closeQuote(p int, q byte, virt bool) (int, bool) {
	next := st.nextOf(q)
	i, k := p, -1
	if virt {
		k = st.spanFrom(p)
	}
	for {
		j := int(next[i])
		if k >= 0 && j >= st.stack[k].a {
			i = st.stack[k].b
			k--
			continue
		}
		if j < st.n && st.s[j] == q {
			return j, true
		}
		return 0, false
	}
}

// narrowRunFrom reads the narrow run of name characters from i and whether it holds a secret word (ASCII case).
func (st *state) narrowRunFrom(i int) run {
	if r, ok := st.runFrom[i]; ok {
		return r
	}
	e := i
	for e < st.n {
		r, n := runeAt(st.s, e)
		if !isASCIIName(r) {
			break
		}
		e += n
	}
	r := run{i, e, e > i && hasKeyword(st.s, i, e, false)}
	st.runFrom[i] = r
	return r
}

// narrowRunTo reads the narrow run of name characters ending at e.
func (st *state) narrowRunTo(e int) run {
	if r, ok := st.runTo[e]; ok {
		return r
	}
	i := e
	for i > 0 {
		r, n := runeBefore(st.s, i)
		if !isASCIIName(r) {
			break
		}
		i -= n
	}
	r := run{i, e, e > i && hasKeyword(st.s, i, e, false)}
	st.runTo[e] = r
	return r
}

// keyAhead reads a narrow key name at x, its optional quote, spaces or tabs, and `:` or `=`.
func (st *state) keyAhead(x int) bool {
	rn := st.narrowRunFrom(x)
	if !rn.keyword {
		return false
	}
	i := rn.end
	if i < st.n && isQuote(st.s[i]) {
		i++
	}
	for i < st.n && (st.s[i] == ' ' || st.s[i] == '\t') {
		i++
	}
	return i < st.n && (st.s[i] == ':' || st.s[i] == '=')
}

// nameOnItsLine is a name at q with its separator on q's line, read narrowly.
func (st *state) nameOnItsLine(q int, virt bool) bool {
	if q >= st.n || virt && st.spanAt(q) >= 0 {
		return false
	}
	s := st.s
	r, _ := runeAt(s, q)
	switch {
	case isASCIIName(r) && st.keyAhead(q):
		return true
	case r == '-':
		_, ok := flagAt(s, q, false, isSpaceTab)
		return ok
	case foldIs(r, 'a', false):
		e, ok := matchWord(s, q, "authorization", false)
		if !ok {
			return false
		}
		for e < st.n && (s[e] == ' ' || s[e] == '\t') {
			e++
		}
		return e < st.n && (s[e] == ':' || s[e] == '=')
	case foldIs(r, 'b', false):
		e, ok := matchWord(s, q, "bearer", false)
		return ok && e < st.n && (s[e] == ' ' || s[e] == '\t')
	}
	return false
}

// belowBlocked reports whether a value at p may not be taken: p starts a later line than its name, and that line
// starts (after an optional quote) with a name on its own line, or is the text's last line and one word.
func (st *state) belowBlocked(p int, virt bool) bool {
	r, _ := runeBefore(st.s, skipBackHSpace(st.s, p))
	if !isLineBreak(r) {
		return false
	}
	q := p
	if q < st.n && isQuote(st.s[q]) {
		q++
	}
	if st.nameOnItsLine(q, virt) {
		return true
	}
	if virt && len(st.stack) > 0 && st.stack[0].a >= q {
		return false
	}
	return q >= st.loneFrom
}

// opensNext reports whether the quote at c opens another key's value: what stands before it is a narrow name
// holding a secret word, an optional quote, `:` or `=`, with spaces or tabs around (`TOKEN="`).
func (st *state) opensNext(c int, virt bool) bool {
	s := st.s
	i := c
	for i > 0 && (s[i-1] == ' ' || s[i-1] == '\t') {
		i--
	}
	if i == 0 || s[i-1] != ':' && s[i-1] != '=' {
		return false
	}
	i--
	for i > 0 && (s[i-1] == ' ' || s[i-1] == '\t') {
		i--
	}
	if i > 0 && isQuote(s[i-1]) {
		i--
	}
	rn := st.narrowRunTo(i)
	if !rn.keyword {
		return false
	}
	return !virt || !st.spanEndIn(rn.start+1, c)
}

// quoted reads a quoted value at p, with glue the class of what may be glued after its closing quote.
func (st *state) quoted(p int, glue cls, virt bool) (int, bool) {
	if p < st.n && isQuote(st.s[p]) {
		if c, ok := st.closeQuote(p+1, st.s[p], virt); ok && !st.opensNext(c, virt) {
			return st.classEnd(glue, c+1, virt), true
		}
		return st.lineEnd(p), true
	}
	return 0, false
}

// cutTail reports a value that is kept as a cut tail: the start of the marker (or, for an Authorization header, of
// a scheme word), with nothing after it but whitespace.
func (st *state) cutTail(p, e int, virt, scheme bool) bool {
	if e <= p || e <= st.lastNonSpace {
		return false
	}
	if virt && st.spanFrom(p) >= 0 && st.stack[st.spanFrom(p)].a < e {
		return false
	}
	v := st.s[p:e]
	if len(v) > len(Marker) {
		return false
	}
	if strings.HasPrefix(Marker, v) {
		return true
	}
	if !scheme {
		return false
	}
	low := strings.ToLower(v)
	for _, w := range schemes {
		if strings.HasPrefix(w, low) {
			return true
		}
	}
	return false
}

// schemes are an Authorization header's scheme words.
var schemes = []string{"basic", "bearer", "token", "digest"}

// value reads a name's value: where it starts and ends, and whether there is one. With virt, the values already
// taken stand as markers.
func (st *state) value(nm name, virt bool) (int, int, bool) {
	p := nm.val
	switch nm.kind {
	case nameKey:
		return st.plain(p, clsKey, virt, true)
	case nameExtra:
		return st.plain(p, clsNonSpace, virt, true)
	case nameFlag:
		return st.flag(p, virt)
	case nameBearer:
		if p < st.n && !st.belowBlocked(p, virt) {
			if e := st.classEnd(clsBearer, p, virt); e-p >= 8 {
				return p, e, true
			}
		}
		return 0, 0, false
	default:
		return st.authorization(p, virt)
	}
}

// plain reads a key's or extraheader's value: quoted, or bare of class c.
func (st *state) plain(p int, c cls, virt, below bool) (int, int, bool) {
	if p >= st.n || below && st.belowBlocked(p, virt) {
		return 0, 0, false
	}
	e, ok := st.quoted(p, c, virt)
	if !ok {
		e = st.classEnd(c, p, virt)
		if e == p {
			return 0, 0, false
		}
	}
	if st.cutTail(p, e, virt, false) {
		return 0, 0, false
	}
	return p, e, true
}

// flag reads a flag's value, on the flag's line: quoted, or bare and not starting with `-`.
func (st *state) flag(p int, virt bool) (int, int, bool) {
	if p >= st.n {
		return 0, 0, false
	}
	r, _ := runeAt(st.s, p)
	if isLineBreak(r) {
		return 0, 0, false
	}
	if _, ok := st.quoted(p, clsQuoteless, virt); !ok {
		if r == '-' && !(virt && st.spanAt(p) >= 0) {
			return 0, 0, false
		}
	}
	return st.plain(p, clsQuoteless, virt, false)
}

// authorization reads an Authorization header's value. After its separator: an optional quote, whitespace, then
// with a scheme word and whitespace the value after it; a quoted value there takes the scheme word with it (the
// header's whole value: the quoted token and what is glued after its closing quote, or, when the quote does not
// close on its line, the word it opens), and `token` followed by `:` or `=` is a key, not a scheme. Without a scheme
// word, or when nothing can follow it, the value is what stands there, the scheme word itself among it.
func (st *state) authorization(p0 int, virt bool) (int, int, bool) {
	s := st.s
	p := p0
	if p < st.n && isQuote(s[p]) {
		p++
	}
	p = skipSpace(s, p)
	tail := 0
	for _, w := range schemes {
		e, ok := matchWord(s, p, w, true)
		if !ok {
			continue
		}
		if r, _ := runeAt(s, e); !isSpace(r) {
			break
		}
		v := skipSpace(s, e)
		if v < st.n && w == "token" && (s[v] == ':' || s[v] == '=') {
			// `token :` is a key, so the header's value is the word token; what stands after the separator is
			// taken too, as it would be after a scheme word.
			if !st.belowBlocked(v, virt) {
				tail, _ = st.authCore(v, virt)
			}
			break
		}
		if v >= st.n || st.belowBlocked(v, virt) {
			break
		}
		if isQuote(s[v]) {
			// A token is one word: a quote not closed on its line takes that word, not the rest of the line.
			e := st.classEnd(clsQuoteless, v+1, virt)
			if c, ok := st.closeQuote(v+1, s[v], virt); ok && !st.opensNext(c, virt) {
				e = st.classEnd(clsQuoteless, c+1, virt)
			}
			return p, e, true
		}
		if e, ok := st.authCore(v, virt); ok {
			if st.cutTail(v, e, virt, true) {
				return 0, 0, false
			}
			return v, e, true
		}
		break
	}
	if p >= st.n || st.belowBlocked(p, virt) {
		return 0, 0, false
	}
	e, ok := st.authCore(p, virt)
	if !ok || st.cutTail(p, e, virt, true) {
		return 0, 0, false
	}
	return p, max(e, tail), true
}

// authCore is an Authorization header's value at p: a marker already written that `,` `;` `&` `}` `)` follows is
// whole (a key's value the header's name also holds stops there, so a second pass leaves it); else a run of
// [^\s"']. A value taken in this reading (virt) that stands at p is read through, as any other text.
func (st *state) authCore(p int, virt bool) (int, bool) {
	if strings.HasPrefix(st.s[p:], Marker) {
		e := p + len(Marker)
		whole := true
		if virt {
			if k := st.spanAt(p); k >= 0 && st.stack[k].b != e {
				whole = false
			}
		}
		if whole && e < st.n && strings.IndexByte(",;&})", st.s[e]) >= 0 && !(virt && st.spanAt(e) >= 0) {
			return e, true
		}
	}
	e := st.classEnd(clsQuoteless, p, virt)
	return e, e > p
}
