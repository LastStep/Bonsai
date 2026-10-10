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

// framings puts a string at the start, after a word, before more words, and on a line of its own between a line and
// a CRLF line holding a key with no secret word.
func framings(s string) []string {
	return []string{s, "seen " + s, s + " so the line goes on", "head\n" + s + "\r\nafter: it\n"}
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
		`read Authorization: Token "Tq83vz, and the line goes on`: "read Authorization: [redacted] and the line goes on",
		`X-Upstream-Authorization: digest 'Tq83vz`:                "X-Upstream-Authorization: [redacted]",
		`proxy-authorization: Basic "dXNlcjpwYXNz" fine`:          "proxy-authorization: [redacted] fine",
		"AUTHORIZATION:\tTOKEN = Tq83vz":                          "AUTHORIZATION:\t[redacted] [redacted]",
	} {
		got := Text(in)
		if got != want {
			t.Errorf("Text(%q) = %q, want %q", in, got, want)
		}
		fixedAtEveryCut(t, in, got)
	}
}

// valueTakers are names whose value would start where a second name stands.
var valueTakers = []string{"store:", "PASSWD=", "--secret", "use bearer", "Proxy-Authorization:", "authorization: Basic",
	"Authorization:\ttoken", "http.extraheader=", `"clientSecret":`, "--api-key", "db_passphrase ="}

// seconds are second names, each with a slot for its secret.
var seconds = []string{"passwd: %s", "access_key=%s", `"secret": "%s"`, "--token %s", "proxy-authorization: %s",
	"bearer %s", "extraheader=%s", "PassPhrase = '%s'", "x-api-key:%s"}

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
		for _, mid := range []string{"secret:", "passwd=", "--token", "Bearer", "authorization="} {
			for _, p := range []string{"(", "{", "[", "`", "<", "\""} {
				for _, f := range framings(first + " " + mid + " " + p + "auth_key: %s" + p) {
					s := m.secret(12)
					checkGone(t, fmt.Sprintf(f, s), s)
					n++
				}
			}
		}
		// A name that takes extraheader= as its value, the rest of that value after a key's value.
		for _, f := range framings(first + " extraHeader= passwd:%s;q7%s") {
			s1, s2 := m.secret(10), m.secret(10)
			checkGone(t, fmt.Sprintf(f, s1, s2), s1, s2)
			n++
		}
		// A quote left open before a name at the line's end, the second value on the line below.
		s1, s2 := m.secret(10), m.secret(10)
		checkGone(t, fmt.Sprintf("%s secret: \"%s and then  passwd:\n  %s\n", first, s1, s2), s1, s2)
		n++
	}
	// A header name holding a secret word, a quoted secret word, a separator and a value.
	for _, sep := range []string{"=", ":"} {
		for _, q := range []string{"'", "\""} {
			s := m.secret(12)
			checkGone(t, fmt.Sprintf("my-secret-authorization%s %spasswd%s%s %s", sep, q, q, sep, s), s)
			n++
		}
	}
	// A key's open quote, then a flag and a quoted key glued after it: no cut of the output changes on a second pass.
	for _, first := range []string{"passwd=", "token: ", "--secret "} {
		s := m.secret(12)
		checkGone(t, fmt.Sprintf(`%s"psql --secret "accessKey": "%s"`, first, s), s)
		s = m.secret(12)
		checkGone(t, fmt.Sprintf(`%s"Tq83vz {"authKey": "%s"}`, first, s), s)
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
	parts := []string{"passwd: %s", "--secret %s", "Authorization: Digest %s", "bearer %s", "extraHeader=%s", "\"apiKey\": \"%s\"", "SIGNING_TOKEN=%s"}
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

// A shape glued to a name: a token, a key id or a webhook URL whose run goes straight on into a secret-named name
// takes the name into its own span, and the name's value still goes. Every shape kind is glued to every kind of name,
// with nothing between and with each separator a shape's run may hold, the name in small letters and in capitals;
// wherever a name takes its value with a plain word in the shape's place, it takes it with the shape there too.
func TestShapeGluedToAName(t *testing.T) {
	m := newMaker(15)
	urlsafe := alnum + "_-"
	b64 := alnum + "+/"
	type shape struct{ kind, text, secret string }
	var shapes []shape
	add := func(kind, before, secret, after string) {
		shapes = append(shapes, shape{kind, before + secret + after, secret})
	}
	add("private-key, no END line", "-----BEGIN DSA PRIVATE KEY-----\n", m.from(b64, 48), "")
	add("webhook-url, discord.com", "https://discord.com/api/webhooks/42/", m.secret(24), "")
	add("webhook-url, Slack", "https://hooks.slack.com/services/T07/", m.secret(20), "")
	add("webhook-url, Slack in capitals", "HTTPS://HOOKS.SLACK.COM/services/", m.secret(20), "")
	add("url-credentials", "https://ci:", m.secret(12), "@pkgs.example.org/team")
	add("anthropic-key", "sk-ant-api03-", m.secret(30), "")
	add("openai-key", "sk-proj-", m.secret(26), "")
	add("stripe-key", "rk_live_", m.secret(20), "")
	add("github-token, fine-grained", "github_pat_", m.secret(30), "")
	add("github-token, OAuth", "gho_", m.secret(30), "")
	add("slack-token", "xoxp-2718281828-", m.secret(16), "")
	add("npm-token", "npm_", m.secret(36), "")
	add("aws-key-id", "ASIA", strings.ToUpper(m.secret(16)), "")
	add("google-api-key", "AIza", m.from(urlsafe, 35), "")
	add("google-oauth-token", "ya29.", m.secret(30), "")
	add("jwt", "eyJ"+m.from(urlsafe, 12)+"."+m.from(urlsafe, 16)+".", m.from(alnum, 20), "")
	add("random-run", "", m.secret(40), "")
	names := []string{
		"password: %s", "token=%s", "api_key = '%s'", "secret\":\"%s\"", "--password %s", "-token %s",
		"--client-secret %s", "upstream-authorization: bearer %s", "authorization=%s", "proxy-authorization: token %s",
		"bearer %s", "extraheader=%s", "extraheader= %s",
	}
	glues := []string{"", "_", "-", ".", "/", ",", ";", "&", "?", "=", ":"}
	n, asked := 0, 0
	for _, sh := range shapes {
		// A plain word ending as the shape ends: the reference a name is read against.
		stand := "w" + sh.text[len(sh.text)-1:]
		for _, g := range glues {
			for _, nm := range names {
				for _, form := range []string{nm, strings.Replace(strings.ToUpper(nm), "%S", "%s", 1)} {
					for _, frame := range []string{"%s", "in the log %s and so on"} {
						v := m.secret(12)
						in := fmt.Sprintf(frame, sh.text+g+fmt.Sprintf(form, v))
						out := Text(in)
						n++
						if w := survives(out, sh.secret); w != "" {
							t.Errorf("%s: %q kept %q of the shape: %q", sh.kind, in, w, out)
						}
						ref := fmt.Sprintf(frame, stand+g+fmt.Sprintf(form, v))
						if survives(Text(ref), v) == "" {
							asked++
							if w := survives(out, v); w != "" {
								t.Errorf("%s: %q kept %q of the value: %q", sh.kind, in, w, out)
							}
						}
						fixedAtEveryCut(t, in, out)
					}
				}
			}
		}
	}
	t.Logf("%d strings, %d with a value to take", n, asked)
	if n != len(shapes)*len(glues)*len(names)*4 || asked < n*3/4 {
		t.Errorf("%d strings, %d with a value to take", n, asked)
	}
}

// A shape that takes an Authorization header's name and scheme word (a webhook URL runs on to the next whitespace,
// so `...,authorization=Basic` is all its own) leaves the token after the scheme word, and the token still goes:
// every webhook URL glued by every character it holds to every header form whose separator its scheme word follows
// straight, every scheme word in several folds, the token after each kind of whitespace (a line's end among them),
// bare, quoted or with its quote left open. Wherever the header takes its token with a plain word in the URL's place,
// it takes it with the URL there too, and every output is a fixed point at every cut.
func TestHeaderTokenPastAShape(t *testing.T) {
	m := newMaker(17)
	urls := []string{"https://discord.com/api/webhooks/31/", "http://canary.discordapp.com/api/webhooks/31/",
		"https://hooks.slack.com/services/T5/", "HTTPS://HOOKS.SLACK.COM/SERVICES/"}
	glues := []string{"", ",", ";", "&", "?", "/", "=", "_", "-", ".", ":", "#", "|", "!", "*"}
	headers := []string{"authorization=", "Authorization:", "Proxy-AUTHORIZATION=", "X-Upstream-Authorization:", "AUTHORIZATION ="}
	schemes := []string{"Basic", "basic", "BASIC", "Baſic", "Token", "token", "TOKEN", "ToKen", "Digest",
		"digeſt", "DIGEST", "Bearer", "bearer"}
	between := []string{" ", "\t", "  ", " ", "　", "\n", "\r\n"}
	tokens := []string{"%s", `"%s"`, "'%s'", `"%s more"`, `"%s`, "%s,rest"}
	frames := []string{"%s", "in the log %s and so on", "x\n%s\r\nnext: line"}
	n, asked, cut := 0, 0, 0
	for _, u := range urls {
		for _, g := range glues {
			for _, h := range headers {
				for si, sc := range schemes {
					for bi, b := range between {
						tok := tokens[(si+bi)%len(tokens)]
						frame := frames[(si+bi+len(g))%len(frames)]
						v := m.secret(14)
						header := h + sc + b + fmt.Sprintf(tok, v)
						in := fmt.Sprintf(frame, u+m.secret(20)+g+header)
						out := Text(in)
						n++
						if survives(Text(fmt.Sprintf(frame, "word"+g+header)), v) == "" {
							asked++
							if w := survives(out, v); w != "" {
								t.Errorf("%q kept %q of the token: %q", in, w, out)
							}
						}
						if Text(out) != out {
							t.Errorf("not a fixed point: %q gave %q, then %q", in, out, Text(out))
						}
						if n%5 == 0 {
							cut++
							fixedAtEveryCut(t, in, out)
						}
					}
				}
			}
		}
	}
	t.Logf("%d strings, %d with a token to take, %d checked at every cut", n, asked, cut)
	if asked < n*9/10 {
		t.Errorf("%d strings, only %d with a token to take", n, asked)
	}
}
