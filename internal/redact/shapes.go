package redact

// The rules that take a whole value by its own shape, with no name before it: private-key blocks, webhook URLs,
// credentials inside a URL, and the known token shapes. They run first, one after another in this order, each over
// the text the ones before it left, as the studio's redactor runs them.
//
// Two departures, both so a cut of the output reads the same a second time:
//   - An AWS key id is `AKIA` or `ASIA` and 16 of [0-9A-Z], and any more of [0-9A-Z] glued after it go with it. The
//     studio's rule also wants a word's end after the 16, so `AKIA` and 17 such characters kept all 21; cut after
//     the 16th, a second pass took them. Here the 16 go wherever the word ends, and where the studio's rule took
//     nothing, the whole run goes when it looks random, as its random-run rule took it (findAWS).
//   - A token shape needs a word's start before it (JavaScript's \b). Where an earlier pass's marker stands right
//     before it, the marker's `]` is that start, so the next pass takes it; the passes run until nothing changes.
//
// And one addition, so no part of a secret is left because a shape took the rest: a shape takes part of a run of the
// random run's characters ([A-Za-z0-9_+/=-]), and what it leaves on either side, too short to be judged on its own,
// goes with it when it holds a piece that looks random. It is judged once every shape has run (leftovers, below).

import "strings"

type shapeRule struct {
	kind Kind
	find func(s string, add func(a, b int))
}

// shapeRules is the order the whole-value rules run in.
var shapeRules = []shapeRule{
	{KindPrivateKey, findPEM},
	{KindWebhook, findDiscord},
	{KindWebhook, findSlack},
	{KindURLCredentials, findURLCredentials},
	{KindAnthropicKey, tokenRule("sk-ant-", nil, isTokenRune, 10, false)},
	{KindOpenAIKey, tokenRule("sk-", nil, isTokenRune, 20, false)},
	{KindStripeKey, findStripe},
	{KindGitHubToken, tokenRule("github_pat_", nil, isWord, 20, false)},
	{KindGitHubToken, tokenRule("gh", []string{"p_", "o_", "u_", "s_", "r_"}, isAlnum, 20, false)},
	{KindSlackToken, tokenRule("xox", []string{"a-", "b-", "p-", "o-", "s-", "r-"}, isAlnumDash, 10, false)},
	{KindNPMToken, tokenRule("npm_", nil, isAlnum, 30, false)},
	{KindAWSKeyID, findAWS},
	{KindGoogleAPIKey, tokenRule("AIza", nil, isTokenRune, 35, true)},
	{KindGoogleOAuth, tokenRule("ya29.", nil, isTokenRune, 10, false)},
	{KindJWT, findJWT},
}

// isTokenRune is [A-Za-z0-9_-].
func isTokenRune(r rune) bool { return isWord(r) || r == '-' }

// isAlnumDash is [A-Za-z0-9-].
func isAlnumDash(r rune) bool { return isAlnum(r) || r == '-' }

// wordStart is JavaScript's \b before a word character at i: the character before is not a word character.
func wordStart(s string, i int) bool {
	r, _ := runeBefore(s, i)
	return !isWord(r)
}

// runOf gives the end of the run of bytes from i that body accepts (every rule here reads ASCII).
func runOf(s string, i int, body func(rune) bool) int {
	for i < len(s) && s[i] < 0x80 && body(rune(s[i])) {
		i++
	}
	return i
}

// tokenRule is a token shape: \b, a prefix, one of the second parts if any, then at least min characters body
// accepts (exactly min when exact).
func tokenRule(prefix string, second []string, body func(rune) bool, min int, exact bool) func(string, func(int, int)) {
	return func(s string, add func(a, b int)) {
		for i := 0; i < len(s); {
			j := strings.Index(s[i:], prefix)
			if j < 0 {
				return
			}
			j += i
			if !wordStart(s, j) {
				i = j + 1
				continue
			}
			k := j + len(prefix)
			if second != nil {
				ok := false
				for _, p := range second {
					if strings.HasPrefix(s[k:], p) {
						k, ok = k+len(p), true
						break
					}
				}
				if !ok {
					i = j + 1
					continue
				}
			}
			// An exact shape reads no further than the characters it keeps.
			limit := len(s)
			if exact && k+min < limit {
				limit = k + min
			}
			e := runOf(s[:limit], k, body)
			if e-k < min {
				i = j + 1
				continue
			}
			add(j, e)
			i = e
		}
	}
}

// findStripe is \b[sr]k_(?:live|test)_[A-Za-z0-9]{10,}.
func findStripe(s string, add func(a, b int)) {
	for i := 0; i+3 <= len(s); i++ {
		if (s[i] != 's' && s[i] != 'r') || s[i+1] != 'k' || s[i+2] != '_' || !wordStart(s, i) {
			continue
		}
		k := i + 3
		if !strings.HasPrefix(s[k:], "live_") && !strings.HasPrefix(s[k:], "test_") {
			continue
		}
		k += 5
		e := runOf(s, k, isAlnum)
		if e-k < 10 {
			continue
		}
		add(i, e)
		i = e - 1
	}
}

// findAWS is \b(?:AKIA|ASIA)[0-9A-Z]{16}, and any more of [0-9A-Z] glued after it. The studio's rule also wants a word's
// end after the 16; where a word character follows them, it takes nothing, and its random-run rule then judges the
// whole run: so here, where a word character follows the 16, the whole run goes when it looks random (the key id glued
// to its secret goes with the secret).
func findAWS(s string, add func(a, b int)) {
	upperDigit := func(r rune) bool { return r >= 'A' && r <= 'Z' || isDigit(r) }
	runL, runR, random := 0, 0, false // the run last judged whole
	for i := 0; i+4 <= len(s); i++ {
		if s[i] != 'A' || (s[i+1:i+4] != "KIA" && s[i+1:i+4] != "SIA") || !wordStart(s, i) {
			continue
		}
		e := runOf(s, i+4, upperDigit)
		if e-(i+4) < 16 {
			continue
		}
		if i+20 < len(s) && isWord(rune(s[i+20])) {
			if i >= runR {
				runL, runR = runAround(s, i)
				random = runR-runL >= 32 && randomRun(s[runL:runR])
			}
			if random {
				add(runL, runR)
				i = runR - 1
				continue
			}
		}
		add(i, e)
		i = e - 1
	}
}

// findJWT is \beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}. When a start fails, every `eyJ` up to
// where the reading stopped fails the same way (its part ends where the failed one's did, and is shorter), so the
// search goes on from there and reads each part once.
func findJWT(s string, add func(a, b int)) {
	for i := 0; i < len(s); {
		j := strings.Index(s[i:], "eyJ")
		if j < 0 {
			return
		}
		j += i
		if !wordStart(s, j) {
			i = j + 1
			continue
		}
		e := runOf(s, j+3, isTokenRune)
		if e-(j+3) < 8 || e >= len(s) || s[e] != '.' {
			i = e
			continue
		}
		f := runOf(s, e+1, isTokenRune)
		if f-(e+1) < 8 || f >= len(s) || s[f] != '.' {
			i = f
			continue
		}
		g := runOf(s, f+1, isTokenRune)
		if g-(f+1) < 8 {
			i = g
			continue
		}
		add(j, g)
		i = g
	}
}

// findPEM is a private-key block from its BEGIN line to its END line, or to the end of the text when it has none.
func findPEM(s string, add func(a, b int)) {
	const begin, end = "-----BEGIN ", "-----END "
	for i := 0; i < len(s); {
		j := strings.Index(s[i:], begin)
		if j < 0 {
			return
		}
		j += i
		h, ok := pemLine(s, j+len(begin))
		if !ok {
			i = j + 1
			continue
		}
		stop := len(s)
		for k := h; k < len(s); {
			m := strings.Index(s[k:], end)
			if m < 0 {
				break
			}
			m += k
			if e, ok := pemLine(s, m+len(end)); ok {
				stop = e
				break
			}
			k = m + 1
		}
		add(j, stop)
		i = stop
	}
}

// pemLine reads the rest of a BEGIN or END line from i: [A-Z0-9 ]* ending in `PRIVATE KEY`, then `-----`.
func pemLine(s string, i int) (int, bool) {
	k := runOf(s, i, func(r rune) bool { return r >= 'A' && r <= 'Z' || isDigit(r) || r == ' ' })
	if !strings.HasSuffix(s[i:k], "PRIVATE KEY") || !strings.HasPrefix(s[k:], "-----") {
		return 0, false
	}
	return k + 5, true
}

// hasPrefixFold reports whether p (lower-case ASCII) starts at byte i of s, ignoring ASCII case only, as the studio's
// URL rules read a host and a path (its `i` flag without `u`).
func hasPrefixFold(s string, i int, p string) bool {
	_, ok := matchWord(s, i, p, false)
	return ok
}

// urlScheme reads `http://` or `https://` (any case) at i and gives where it ends.
func urlScheme(s string, i int) (int, bool) {
	if !hasPrefixFold(s, i, "http") {
		return 0, false
	}
	k := i + 4
	if k < len(s) && (s[k] == 's' || s[k] == 'S') {
		k++
	}
	if !strings.HasPrefix(s[k:], "://") {
		return 0, false
	}
	return k + 3, true
}

// isURLRune is [^\s"'<>]: the characters of a webhook URL's secret part.
func isURLRune(r rune) bool { return !isSpace(r) && r != '"' && r != '\'' && r != '<' && r != '>' }

// urlRest gives the end of the run of URL characters from i.
func urlRest(s string, i int) int {
	for i < len(s) {
		r, n := runeAt(s, i)
		if !isURLRune(r) {
			break
		}
		i += n
	}
	return i
}

// findDiscord is https?://(?:[a-z]+\.)?discord(?:app)?\.com/api/webhooks/[^\s"'<>]+, in any case.
func findDiscord(s string, add func(a, b int)) {
	host := func(k int) (int, bool) {
		if !hasPrefixFold(s, k, "discord") {
			return 0, false
		}
		k += len("discord")
		if hasPrefixFold(s, k, "app.com/api/webhooks/") {
			return k + len("app.com/api/webhooks/"), true
		}
		if hasPrefixFold(s, k, ".com/api/webhooks/") {
			return k + len(".com/api/webhooks/"), true
		}
		return 0, false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != 'h' && s[i] != 'H' {
			continue
		}
		k, ok := urlScheme(s, i)
		if !ok {
			continue
		}
		v, ok := host(k)
		if !ok {
			l := runOf(s, k, isLetter)
			if l > k && l < len(s) && s[l] == '.' {
				v, ok = host(l + 1)
			}
		}
		if !ok {
			continue
		}
		if e := urlRest(s, v); e > v {
			add(i, e)
			i = e - 1
		}
	}
}

// findSlack is https?://hooks\.slack\.com/services/[^\s"'<>]+, in any case.
func findSlack(s string, add func(a, b int)) {
	for i := 0; i < len(s); i++ {
		if s[i] != 'h' && s[i] != 'H' {
			continue
		}
		k, ok := urlScheme(s, i)
		if !ok || !hasPrefixFold(s, k, "hooks.slack.com/services/") {
			continue
		}
		v := k + len("hooks.slack.com/services/")
		if e := urlRest(s, v); e > v {
			add(i, e)
			i = e - 1
		}
	}
}

// isSchemeRune is [a-z0-9+.-] in any case: the characters of a URL's scheme and what is glued before it.
func isSchemeRune(r rune) bool { return isAlnum(r) || r == '+' || r == '.' || r == '-' }

// findURLCredentials takes user:pass out of scheme://user:pass@host and keeps the rest. The scheme is read from the
// start of its run of scheme characters, and needs a letter at a word's start inside that run.
func findURLCredentials(s string, add func(a, b int)) {
	userRune := func(r rune) bool { return isURLRune(r) && r != '/' && r != ':' && r != '@' }
	passRune := func(r rune) bool { return isURLRune(r) && r != '/' && r != '@' }
	run := func(i int, ok func(rune) bool) int {
		for i < len(s) {
			r, n := runeAt(s, i)
			if !ok(r) {
				break
			}
			i += n
		}
		return i
	}
	for i := 0; i < len(s); i++ {
		if !isSchemeRune(rune(s[i])) || i > 0 && isSchemeRune(rune(s[i-1])) {
			continue
		}
		r := runOf(s, i, isSchemeRune)
		if !strings.HasPrefix(s[r:], "://") {
			i = r - 1
			continue
		}
		letter := false
		for k := i; k < r && !letter; k++ {
			letter = isLetter(rune(s[k])) && wordStart(s, k)
		}
		u := r + 3
		ue := run(u, userRune)
		if !letter || ue == u || ue >= len(s) || s[ue] != ':' {
			i = r - 1
			continue
		}
		pe := run(ue+1, passRune)
		if pe == ue+1 || pe >= len(s) || s[pe] != '@' {
			i = r - 1
			continue
		}
		add(u, pe)
		i = pe
	}
}

// leftovers gives the parts of their runs that the shapes' spans in s (in order, apart: every span the shapes of one
// pass found, in the text as it stood before them) leave, and that must go with them. A shape takes part of a run of
// the random run's characters (a token with more glued after it, an exact-length key, a JSON web token's first or
// last part, the start of a webhook URL with a word glued before it); the random-run rule then judges what is left on
// each side as a run of its own, which is often under its 32 characters, though the whole run is not: a secret glued
// to a key id (`AKIA...` and AWS's secret after it) would be kept. So each such leftover is judged with its run: when
// the whole run is 32 or more and the leftover holds a piece that looks random (between the run's separators, a piece
// the shape's edge and the leftover share counted whole), it goes. A leftover of names and words (`TOKEN_FOR_CI=`
// before a token, `_token=` or `-backup` after one) has no such piece and stays, as the studio's redactor leaves it;
// and a leftover the random-run rule never takes (one starting with `toolu_`) stays too. Judged with every shape's
// span at once, a leftover never holds another shape's start. Every run and piece is read once, so the time stays
// linear.
func leftovers(s string, spans []Span) []Span {
	var out []Span
	runL, runR := 0, 0 // the run last read: the leftovers of one run share it
	judge := func(x, y int, kind Kind) {
		if x >= runR {
			runL, runR = runAround(s, x)
		}
		if runR-runL < 32 || strings.HasPrefix(s[x:], "toolu_") || strings.HasPrefix(s[x:], "srvtoolu_") {
			return
		}
		// The piece the leftover shares with the shape on either side, when no separator stands between them.
		ps, pe := x, y
		if x > runL && !isPieceSep(s[x-1]) && !isPieceSep(s[x]) {
			for ps > runL && !isPieceSep(s[ps-1]) {
				ps--
			}
		}
		if y < runR && !isPieceSep(s[y-1]) && !isPieceSep(s[y]) {
			for pe < runR && !isPieceSep(s[pe]) {
				pe++
			}
		}
		for start, k := ps, ps; k <= pe; k++ {
			if k == pe || isPieceSep(s[k]) {
				if looksRandom(s[start:k]) {
					out = append(out, Span{x, y, kind})
					return
				}
				start = k + 1
			}
		}
	}
	prevEnd := 0
	for i, sp := range spans {
		// Before the span, back to the run's start or the span before it; when that span's leftover reached this one,
		// it was judged with that span.
		if x := sp.Start; x > prevEnd || i == 0 {
			for x > prevEnd && isRunByte(s[x-1]) {
				x--
			}
			if x < sp.Start && (x > prevEnd || i == 0) {
				judge(x, sp.Start, sp.Kind)
			}
		}
		// After it, to the run's end or the next span.
		next := len(s)
		if i+1 < len(spans) {
			next = spans[i+1].Start
		}
		y := sp.End
		for y < next && isRunByte(s[y]) {
			y++
		}
		if y > sp.End {
			judge(sp.End, y, sp.Kind)
		}
		prevEnd = sp.End
	}
	return out
}
