// Package redact is Bonsai's redactor (spec §8; contract §2.6): what may leave the machine in a record, and in
// what form. It is text in and text out, with no file system, clock or network, and it gives the same on every
// system. Every secret pattern Bonsai has lives here; no other package holds one.
//
// It has two jobs:
//   - Text takes every secret it finds out of a string and writes Marker in its place. Find gives the same as spans,
//     each with the kind of rule that took it, so a caller can say where a secret is (a memory note and its line)
//     without holding it. Capped is Text, then the cut to a field's cap.
//   - The reductions (reduce.go) bring a tool call down to the one string a record keeps: Target (a path inside the
//     checkout, a command's head, a host, an MCP tool, a subagent type, a skill), CommandHead, InsidePath and
//     Question. Each passes Text and its cap. What is never kept needs no redacting, so they are the first line of
//     defence.
//
// How Text reads a string, in the order it runs:
//  1. shapes.go: the rules that take a whole value by its shape: private-key blocks, webhook URLs, credentials
//     inside a URL, and the known token shapes.
//  2. names.go and values.go: every name a value follows is found first (a secret-named key before `:` or `=`, a
//     flag ending in a secret word, an Authorization header, a Bearer, extraheader=), then each name's value. A name
//     that stands inside another name's value has its own value taken, and the outer value takes the inner name
//     too, so no value is left because another rule reached it first. The studio's redactor took a name's value
//     before the next rule could see the name inside it, and leaked in three classes of shape; read this way the
//     three are one rule: every name's value goes, wherever the name stands. A name a shape took into its span
//     (`AKIA...PASSWORD: x`, a token or a webhook URL with `_token=` or `,password:` glued on) is still a name: it
//     is found in the text as it stood before the shapes, and its value goes too (lost, below).
//  3. random.go: a long run of token characters that looks random.
//
// The three run again over their own output until it no longer changes (one pass is nearly always enough).
//
// What it promises:
//   - At least what the studio's redactor takes out, and its leak classes closed: that redactor is the floor its
//     differential is measured against, not an oracle of what to write.
//   - A fixed point, also cut: Text of its own output gives that output back, and so does Text of that output cut at
//     any character. The recorder cuts a field to its cap after redacting, and the studio's bridge redacts again.
//   - Linear time: every rule reads a run once, so no input makes it slow (time_test.go).
//   - What it keeps: git SHAs and other pure hex, UUIDs and GUIDs, toolu_ ids, CamelCase names, prose that says
//     "token" or "password" with no value after it, words that only end in -secret or -token, a flag at a line's end,
//     and text already redacted.
//
// What it is not: a proof that nothing secret remains. A secret with no shape and no name before it (a bare word on
// a line of its own) is not found; the reductions keep most such text out of a record before Text sees it. It reads
// text, not a shell: quotes are paired by the rules above, not as a shell would pair them.
package redact

import "sort"

// Marker is what stands in for every value taken out. The studio's bridge and Desk know it.
const Marker = "[redacted]"

// Kind names the rule that took a span out.
type Kind string

// The kinds of span Find gives. Kinds is their one home.
const (
	KindPrivateKey     Kind = "private-key"
	KindWebhook        Kind = "webhook-url"
	KindURLCredentials Kind = "url-credentials"
	KindAnthropicKey   Kind = "anthropic-key"
	KindOpenAIKey      Kind = "openai-key"
	KindStripeKey      Kind = "stripe-key"
	KindGitHubToken    Kind = "github-token"
	KindSlackToken     Kind = "slack-token"
	KindNPMToken       Kind = "npm-token"
	KindAWSKeyID       Kind = "aws-key-id"
	KindGoogleAPIKey   Kind = "google-api-key"
	KindGoogleOAuth    Kind = "google-oauth-token"
	KindJWT            Kind = "jwt"
	KindKey            Kind = "secret-named-key"
	KindFlag           Kind = "secret-flag"
	KindAuthorization  Kind = "authorization-header"
	KindBearer         Kind = "bearer-token"
	KindExtraHeader    Kind = "extra-header"
	KindRandom         Kind = "random-run"
)

// Kinds lists every kind with what it takes out, in the order the rules run.
var Kinds = []struct {
	Kind  Kind
	Means string
}{
	{KindPrivateKey, "a private-key block, from its BEGIN line to its END line or the text's end"},
	{KindWebhook, "a Discord or Slack webhook URL, whole"},
	{KindURLCredentials, "the user:password inside scheme://user:password@host"},
	{KindAnthropicKey, "an sk-ant- key"},
	{KindOpenAIKey, "an sk- key"},
	{KindStripeKey, "an sk_live_, sk_test_, rk_live_ or rk_test_ key"},
	{KindGitHubToken, "a github_pat_, ghp_, gho_, ghu_, ghs_ or ghr_ token"},
	{KindSlackToken, "an xoxa-, xoxb-, xoxp-, xoxo-, xoxs- or xoxr- token"},
	{KindNPMToken, "an npm_ token"},
	{KindAWSKeyID, "an AKIA or ASIA key id"},
	{KindGoogleAPIKey, "an AIza key"},
	{KindGoogleOAuth, "a ya29. token"},
	{KindJWT, "a JSON web token (eyJ..., three dotted parts)"},
	{KindKey, "the value after a secret-named key and its : or = (password: x, DB_TOKEN=x, \"apiKey\": \"x\")"},
	{KindFlag, "the value after a flag ending in a secret word (--password x, --client-secret x)"},
	{KindAuthorization, "the value of an Authorization header, after its scheme word"},
	{KindBearer, "a Bearer value of eight or more token characters"},
	{KindExtraHeader, "the value after extraheader= (git's http.extraHeader)"},
	{KindRandom, "a run of 32 or more token characters that looks random"},
}

// Span is a part of a text the redactor takes out: bytes [Start, End) of the text it was given, and the kind of the
// rule that took it. A span is never its value: a caller that names a finding names where it is, not what it holds.
type Span struct {
	Start, End int
	Kind       Kind
}

// Text gives s with every secret it finds replaced by Marker. Text of its own output gives that output back, and so
// does Text of that output cut at any length (see the package comment).
func Text(s string) string {
	if s == "" {
		return s
	}
	d := redactDoc(s)
	return d.text
}

// Find gives the spans Text replaces, in order, none overlapping or touching another.
func Find(s string) []Span {
	if s == "" {
		return nil
	}
	d := redactDoc(s)
	return append([]Span(nil), d.spans...)
}

// Capped redacts s and then cuts it to at most max characters (code points), the order a record's field is written
// in: the redactor reads the whole field, and a cut tail of its marker is kept by any later pass.
func Capped(s string, max int) string {
	return Cut(Text(s), max)
}

// Cut cuts s to at most max characters (code points), never inside one.
func Cut(s string, max int) string {
	if max <= 0 {
		return ""
	}
	n := 0
	for i := range s {
		if n == max {
			return s[:i]
		}
		n++
	}
	return s
}

// maxPasses bounds the passes over a text. One pass is almost always enough and a second finds nothing; a third is
// needed only where a pass's marker makes a name of what stood glued to the secret before it.
const maxPasses = 8

// doc is a text with the spans taken out of it so far, in the original's byte offsets.
type doc struct {
	orig  string
	spans []Span // ascending, disjoint, none touching
	text  string // orig with every span replaced by Marker
	segs  []seg  // how text maps onto orig
}

// seg is a piece of the current text: kept from orig, or a marker standing for orig[o:o+ol].
type seg struct {
	t, o, ol int
	marker   bool
}

func redactDoc(s string) *doc {
	d := &doc{orig: s}
	d.render()
	for pass := 0; pass < maxPasses; pass++ {
		changed, shaped := false, false
		preText, preSegs := d.text, d.segs // render makes new ones, so these stay the text before the shapes
		for _, r := range shapeRules {
			if d.apply(r.kind, r.find) {
				changed, shaped = true, true
			}
		}
		names := findNames(d.text)
		if shaped {
			names = d.withLost(names, preText, preSegs)
		}
		if d.merge(d.mapSpans(nameValues(d.text, names))) {
			d.render()
			changed = true
		}
		if d.apply(KindRandom, findRandom) {
			changed = true
		}
		if !changed {
			break
		}
	}
	return d
}

// withLost adds to names, the names of the current text, the names this pass's shapes took into their spans: those
// found in the text as it stood before the shapes (preText, mapped onto orig by preSegs) and not now, whose values
// start outside every span. Each stands in the current text where it stood, a part of it inside a marker at the
// marker's start, so its value is read as any other name's. A value that starts inside the span that took its name
// is that span's: the password of a URL's credentials ends at its `@`, and the host after it stays.
func (d *doc) withLost(names []name, preText string, preSegs []seg) []name {
	pre := findNames(preText)
	if len(pre) == 0 {
		return names
	}
	type at struct {
		kind nameKind
		val  int
	}
	now := make(map[at]bool, len(names))
	for _, nm := range names {
		now[at{nm.kind, origAt(d.segs, nm.val)}] = true
	}
	added := false
	for _, nm := range pre {
		v := origAt(preSegs, nm.val)
		if now[at{nm.kind, v}] {
			continue
		}
		if k := sort.Search(len(d.spans), func(i int) bool { return d.spans[i].End > v }); k < len(d.spans) && d.spans[k].Start < v {
			continue
		}
		names = append(names, name{nm.kind, d.textAt(origAt(preSegs, nm.start)), d.textAt(v)})
		added = true
	}
	if added {
		sort.SliceStable(names, func(i, j int) bool { return names[i].start < names[j].start })
	}
	return names
}

// origAt gives the byte of orig that byte pos of a text stands for, by the text's map: in a marker, the start of
// what it stands for; at the text's end, orig's end.
func origAt(segs []seg, pos int) int {
	k := sort.Search(len(segs), func(i int) bool { return segs[i].t > pos }) - 1
	if k < 0 {
		return 0
	}
	g := segs[k]
	switch {
	case !g.marker:
		return g.o + min(pos-g.t, g.ol)
	case pos >= g.t+len(Marker):
		return g.o + g.ol
	default:
		return g.o
	}
}

// textAt gives the byte of the current text that stands for byte o of orig: inside a span, its marker's start.
func (d *doc) textAt(o int) int {
	k := sort.Search(len(d.segs), func(i int) bool { return d.segs[i].o > o }) - 1
	if k < 0 {
		return 0
	}
	g := d.segs[k]
	switch {
	case !g.marker:
		return g.t + min(o-g.o, g.ol)
	case o >= g.o+g.ol:
		return g.t + len(Marker)
	default:
		return g.t
	}
}

// apply runs one rule over the current text and takes out what it finds; it reports whether anything changed.
func (d *doc) apply(kind Kind, find func(string, func(int, int))) bool {
	var found []Span
	find(d.text, func(a, b int) { found = append(found, Span{a, b, kind}) })
	if !d.merge(d.mapSpans(found)) {
		return false
	}
	d.render()
	return true
}

// render rebuilds the current text and its map from orig and the spans.
func (d *doc) render() {
	if len(d.spans) == 0 {
		d.text = d.orig
		d.segs = []seg{{0, 0, len(d.orig), false}}
		return
	}
	b := make([]byte, 0, len(d.orig))
	d.segs = make([]seg, 0, 2*len(d.spans)+1)
	at := 0
	for _, sp := range d.spans {
		if sp.Start > at {
			d.segs = append(d.segs, seg{len(b), at, sp.Start - at, false})
			b = append(b, d.orig[at:sp.Start]...)
		}
		d.segs = append(d.segs, seg{len(b), sp.Start, sp.End - sp.Start, true})
		b = append(b, Marker...)
		at = sp.End
	}
	if at < len(d.orig) {
		d.segs = append(d.segs, seg{len(b), at, len(d.orig) - at, false})
		b = append(b, d.orig[at:]...)
	}
	d.text = string(b)
}

// tlen is a piece's length in the current text.
func (g seg) tlen() int {
	if g.marker {
		return len(Marker)
	}
	return g.ol
}

// mapSpans turns spans of the current text into spans of orig, sorted. A span that starts or ends inside a marker
// takes the whole of what that marker stands for.
func (d *doc) mapSpans(in []Span) []Span {
	if len(in) == 0 {
		return nil
	}
	out := make([]Span, 0, len(in))
	k := 0
	find := func(pos int) int {
		// The piece holding text byte pos; the spans come in order, so the search walks forward.
		for k+1 < len(d.segs) && d.segs[k+1].t <= pos {
			k++
		}
		for k > 0 && d.segs[k].t > pos {
			k--
		}
		return k
	}
	sort.SliceStable(in, func(i, j int) bool { return in[i].Start < in[j].Start })
	for _, sp := range in {
		if sp.End <= sp.Start {
			continue
		}
		ga := d.segs[find(sp.Start)]
		a := ga.o
		if !ga.marker {
			a = ga.o + sp.Start - ga.t
		}
		gb := d.segs[find(sp.End-1)]
		b := gb.o + gb.ol
		if !gb.marker {
			b = gb.o + sp.End - gb.t
		}
		out = append(out, Span{a, b, sp.Kind})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Start < out[j].Start })
	return out
}

// merge adds spans (sorted by start) to the doc's, joining any that overlap or touch; a joined span keeps the kind of
// its leftmost part. It reports whether the doc's spans changed.
func (d *doc) merge(add []Span) bool {
	if len(add) == 0 {
		return false
	}
	out := make([]Span, 0, len(d.spans)+len(add))
	i, j := 0, 0
	push := func(sp Span) {
		if n := len(out); n > 0 && sp.Start <= out[n-1].End {
			if sp.End > out[n-1].End {
				out[n-1].End = sp.End
			}
			return
		}
		out = append(out, sp)
	}
	for i < len(d.spans) || j < len(add) {
		if j >= len(add) || i < len(d.spans) && d.spans[i].Start <= add[j].Start {
			push(d.spans[i])
			i++
		} else {
			push(add[j])
			j++
		}
	}
	changed := len(out) != len(d.spans)
	for k := 0; !changed && k < len(out); k++ {
		changed = out[k].Start != d.spans[k].Start || out[k].End != d.spans[k].End
	}
	d.spans = out
	return changed
}
