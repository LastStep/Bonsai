package redact

import (
	"math/rand"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// The made-up secrets the tests plant. hunter2 and AWS's documented example key and secret are well known; every
// other secret is a random string the test makes from a fixed seed, so a run is repeatable.
const (
	awsKeyID     = "AKIAIOSFODNN7EXAMPLE"
	awsSecretKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
)

const alnum = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// maker makes random secrets: each has a capital, a small letter and a digit, and starts with a letter.
type maker struct{ r *rand.Rand }

func newMaker(seed int64) *maker { return &maker{rand.New(rand.NewSource(seed))} }

func (m *maker) from(set string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = set[m.r.Intn(len(set))]
	}
	return string(b)
}

// secret is n random characters of [A-Za-z0-9] with at least one of each kind, starting with a letter.
func (m *maker) secret(n int) string {
	if n < 4 {
		n = 4
	}
	s := []byte(m.from(alnum, n))
	s[0] = "abcdefghijklmnopqrstuvwxyz"[m.r.Intn(26)]
	s[1] = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"[m.r.Intn(26)]
	s[2] = "0123456789"[m.r.Intn(10)]
	return string(s)
}

// survives reports what of secret the text still holds: a word of it (a maximal run of letters and digits, three or
// more long) or a run of 6 or more of its characters. "" when nothing survives.
func survives(text, secret string) string {
	for _, w := range wordsOf(secret) {
		if len(w) >= 3 && hasWord(text, w) {
			return w
		}
	}
	r := []rune(secret)
	for i := 0; i+6 <= len(r); i++ {
		if strings.Contains(text, string(r[i:i+6])) {
			return string(r[i : i+6])
		}
	}
	return ""
}

func wordsOf(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

func hasWord(text, w string) bool {
	for _, x := range wordsOf(strings.ReplaceAll(text, Marker, " ")) {
		if x == w {
			return true
		}
	}
	return false
}

// fixedAtEveryCut checks that Text of out, and of out cut at every character, gives it back.
func fixedAtEveryCut(t *testing.T, in, out string) {
	t.Helper()
	if again := Text(out); again != out {
		t.Errorf("not a fixed point: %q gave %q, then %q", in, out, again)
		return
	}
	for k := 1; k < len(out); k++ {
		if !utf8.RuneStart(out[k]) {
			continue
		}
		if c := out[:k]; Text(c) != c {
			t.Errorf("not a fixed point cut at %d: %q gave %q; the cut %q gave %q", k, in, out, c, Text(c))
			return
		}
	}
}

// checkGone redacts in, checks that none of the secrets survives and that the output is a fixed point at every cut.
func checkGone(t *testing.T, in string, secrets ...string) string {
	t.Helper()
	out := Text(in)
	for _, s := range secrets {
		if w := survives(out, s); w != "" {
			t.Errorf("%q kept %q of the secret %q: %q", in, w, s, out)
		}
	}
	if !strings.Contains(out, Marker) {
		t.Errorf("%q: no marker in %q", in, out)
	}
	fixedAtEveryCut(t, in, out)
	return out
}
