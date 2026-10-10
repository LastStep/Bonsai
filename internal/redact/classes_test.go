package redact

import (
	"fmt"
	"strings"
	"testing"
)

// The three classes of leak the studio's redactor had, by their shapes. With every name found first and every name
// inside another's value taken into it, the three are one rule: every name's value goes, wherever the name stands.
// Each string here plants a distinct made-up secret in every value slot; none may survive, and every output is a
// fixed point at every cut.

// framings puts a string at the start, after a word, before more words, and on a line of its own.
func framings(s string) []string {
	return []string{s, "note " + s, s + " and more words after it", "x\n" + s + "\r\nnext: line\n"}
}

// Class one: what follows an Authorization header's scheme word. A quoted token is a value like any other, and a
// `token` followed by `:` or `=` is a key, not the scheme, so its value goes too.
func TestAfterTheSchemeWord(t *testing.T) {
	m := newMaker(11)
	headers := []string{"Authorization: ", "authorization=", "Proxy-Authorization: ", "x-authorization:\t", "Authorization:\t", "AUTHORIZATION :  "}
	n := 0
	for _, h := range headers {
		for _, scheme := range []string{"Bearer", "bearer", "Token", "token", "Basic", "Digest"} {
			for _, q := range []string{`"`, `'`} {
				for _, f := range framings(h + scheme + " " + q + "%s" + q) {
					s := m.secret(12)
					checkGone(t, fmt.Sprintf(f, s), s)
					n++
				}
			}
		}
		for _, ws := range []string{"", " ", "\t", "  "} {
			for _, sep := range []string{":", "="} {
				for _, f := range framings(h + "token" + ws + sep + " %s") {
					s := m.secret(12)
					checkGone(t, fmt.Sprintf(f, s), s)
					n++
				}
			}
		}
	}
	if n != 6*(6*2*4+4*2*4) {
		t.Errorf("%d strings", n)
	}
	// A quote after the scheme that does not close on its line takes the word it opens: a token is one word, and the
	// prose after it stays.
	for in, want := range map[string]string{
		`see Authorization: Bearer "hunter2, then more words`: "see Authorization: [redacted] then more words",
		`Authorization: token 'hunter2`:                       "Authorization: [redacted]",
		`Authorization: Basic "aHVudGVyMg==" ok`:              "Authorization: [redacted] ok",
		"Authorization: token : hunter2":                      "Authorization: [redacted] : [redacted]",
	} {
		got := Text(in)
		if got != want {
			t.Errorf("Text(%q) = %q, want %q", in, got, want)
		}
		fixedAtEveryCut(t, in, got)
	}
}

// valueTakers are names whose value would start where a second name stands.
var valueTakers = []string{"secrets:", "TOKEN=", "--token", "a bearer", "Authorization:", "Authorization: Bearer",
	"Authorization: token", "git -c http.extraHeader=", `"apiKey":`, "--client-secret", "db_password ="}

// seconds are second names, each with a slot for its secret.
var seconds = []string{"password: %s", "api_key=%s", `"token": "%s"`, "--secret %s", "authorization: %s",
	"bearer %s", "extraHeader=%s", "Password = '%s'", "x-auth-key:%s"}

// Class two: a second name not at the very start of the first name's value: behind punctuation, one character in,
// behind another extraHeader=, or with any whitespace before its `:` or `=`.
func TestSecondNameInsideAValue(t *testing.T) {
	m := newMaker(12)
	before := []string{"{", "(", "[", "<", "*", "|", ">", "`", "\u201c", "\u00ab", "\u00bf", "x", "{\"", "\"", "'", "http.extraHeader= ", "*["}
	n := 0
	for _, first := range valueTakers {
		for _, p := range before {
			for _, second := range seconds {
				if p == "x" && (strings.HasPrefix(second, "-") || strings.HasPrefix(second, "bearer")) {
					continue // a flag and a Bearer start a word: x--secret and xbearer are not names
				}
				for _, f := range framings(first + " " + p + second) {
					s := m.secret(12)
					checkGone(t, fmt.Sprintf(f, s), s)
					n++
				}
			}
		}
	}
	// Any whitespace between the second name and its separator.
	for _, first := range valueTakers {
		for _, sp := range []string{"\v", "\f", "\u00a0", "\u2028", "\ufeff", "\u3000", " \t"} {
			for _, sep := range []string{":", "="} {
				s := m.secret(12)
				checkGone(t, fmt.Sprintf("%s password%s%s %s", first, sp, sep, s), s)
				n++
			}
		}
	}
	// extraHeader= behind extraHeader=.
	for _, f := range framings("git -c http.extraHeader= http.extraHeader= %s push") {
		s := m.secret(12)
		checkGone(t, fmt.Sprintf(f, s), s)
		n++
	}
	if n < 6000 {
		t.Errorf("only %d strings", n)
	}
}

// Class three: three names in a row with the third behind punctuation; a name that takes `extraheader=` as its
// value; a quote left open before a name at a line's end; and the two shapes where a second pass changed the output.
func TestNamesInARow(t *testing.T) {
	m := newMaker(13)
	n := 0
	for _, first := range valueTakers {
		for _, mid := range []string{"password:", "token=", "--secret", "bearer", "Authorization:"} {
			for _, p := range []string{"(", "{", "[", "`", "<", "\""} {
				for _, f := range framings(first + " " + mid + " " + p + "api_key: %s" + p) {
					s := m.secret(12)
					checkGone(t, fmt.Sprintf(f, s), s)
					n++
				}
			}
		}
		// A name that takes extraheader= as its value, the rest of that value after a key's value.
		for _, f := range framings(first + " extraheader= password:%s;x9%s") {
			s1, s2 := m.secret(10), m.secret(10)
			checkGone(t, fmt.Sprintf(f, s1, s2), s1, s2)
			n++
		}
		// A quote left open before a name at the line's end, the second value on the line below.
		s1, s2 := m.secret(10), m.secret(10)
		checkGone(t, fmt.Sprintf("%s password: \"%s more  password:\n  %s\n", first, s1, s2), s1, s2)
		n++
	}
	// A header name holding a secret word, a quoted secret word, a separator and a value.
	for _, sep := range []string{"=", ":"} {
		for _, q := range []string{"'", "\""} {
			s := m.secret(12)
			checkGone(t, fmt.Sprintf("x-token-authorization%s %spassword%s%s %s", sep, q, q, sep, s), s)
			n++
		}
	}
	// A key's open quote, then a flag and a quoted key glued after it: no cut of the output changes on a second pass.
	for _, first := range []string{"token=", "secret: ", "--password "} {
		s := m.secret(12)
		checkGone(t, fmt.Sprintf(`%s"mysql --password "apiKey": "%s"`, first, s), s)
		s = m.secret(12)
		checkGone(t, fmt.Sprintf(`%s"hunter2 {"apiKey": "%s"}`, first, s), s)
		n += 2
	}
	if n < 1300 {
		t.Errorf("only %d strings", n)
	}
}

// The examples of note 1 of the plan's redactor section, exactly.
func TestOuterValueTakesTheInnerName(t *testing.T) {
	for in, want := range map[string]string{
		`a bearer password: "hunter2"`:    "a bearer [redacted]: [redacted]",
		"a bearer db_password=hunter2 ok": "a bearer [redacted] ok",
		"password: mysecret:123":          "password: [redacted]",
	} {
		got := Text(in)
		if got != want {
			t.Errorf("Text(%q) = %q, want %q", in, got, want)
		}
		fixedAtEveryCut(t, in, got)
	}
}

// Names packed together, each with its own made-up value: none survives.
func TestManyNamesTogether(t *testing.T) {
	m := newMaker(14)
	parts := []string{"password: %s", "--token %s", "Authorization: Bearer %s", "bearer %s", "extraheader=%s", "\"secret\": \"%s\"", "TOKEN=%s"}
	// Glued with nothing between, a flag or a Bearer would not start a word, and would be no name.
	glue := []string{" ", ", ", "\n", " (", "\t", "; ", "\r\n"}
	for round := 0; round < 200; round++ {
		var b strings.Builder
		var secrets []string
		for k := 0; k < 6; k++ {
			s := m.secret(12)
			secrets = append(secrets, s)
			b.WriteString(fmt.Sprintf(parts[m.r.Intn(len(parts))], s))
			b.WriteString(glue[m.r.Intn(len(glue))])
		}
		b.WriteString("end\n")
		checkGone(t, b.String(), secrets...)
	}
}
