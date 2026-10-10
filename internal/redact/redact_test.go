package redact

import (
	"fmt"
	"strings"
	"testing"
)

// Every rule that takes a whole value by its shape, each with a made-up secret: none survives, and the output is a
// fixed point at every cut.
func TestShapesAreTakenOut(t *testing.T) {
	m := newMaker(1)
	b64 := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	urlsafe := alnum + "_-"
	pemBody := m.from(b64, 64)
	jwt := "eyJ" + m.from(urlsafe, 20) + "." + m.from(urlsafe, 30) + "." + m.from(urlsafe, 40)
	rows := []struct {
		name, text string
		secrets    []string
	}{
		{"private key", "here:\n-----BEGIN RSA PRIVATE KEY-----\n" + pemBody + "\n-----END RSA PRIVATE KEY-----\nafter", []string{pemBody}},
		{"private key with no END line", "-----BEGIN EC PRIVATE KEY-----\n" + pemBody, []string{pemBody}},
		{"Discord webhook", "send to https://discord.com/api/webhooks/42/" + m.secret(30) + " now", nil},
		{"Discord webhook, other host", "https://canary.discordapp.com/api/webhooks/7/" + m.secret(20), nil},
		{"Slack webhook", "post https://hooks.slack.com/services/" + m.secret(9) + "/" + m.secret(24), nil},
		{"credentials in a URL", "git clone https://builder:" + m.secret(12) + "@example.com/team/repo.git", nil},
		{"Anthropic key", "export it: sk-ant-api03-" + m.secret(40) + " ok", nil},
		{"OpenAI-style key", "key sk-proj-" + m.secret(30), nil},
		{"Stripe-style live key", "sk_live_" + m.secret(24), nil},
		{"Stripe-style restricted test key", "rk_test_" + m.secret(24), nil},
		{"GitHub fine-grained token", "github_pat_" + m.secret(40), nil},
		{"GitHub token", "push with ghp_" + m.secret(36), nil},
		{"GitHub app token", "ghs_" + m.secret(36) + " expires", nil},
		{"Slack token", "xoxb-4815162342-" + m.secret(24), nil},
		{"npm token", "//registry.example.org/:_authToken=npm_" + m.secret(36), nil},
		{"AWS key id", "the id is " + awsKeyID + ".", []string{awsKeyID}},
		{"Google API key", "maps " + "AIza" + m.from(urlsafe, 35), nil},
		{"Google OAuth token", "ya29." + m.secret(40), nil},
		{"JSON web token", "jwt " + jwt, []string{jwt}},
		{"a long random run", "blob " + m.secret(44) + " end", nil},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			secrets := r.secrets
			if secrets == nil {
				secrets = randomParts(r.text)
			}
			checkGone(t, r.text, secrets...)
		})
	}
}

// randomParts picks the made-up secrets out of a row: its words of 9 or more characters with a digit and a capital.
func randomParts(text string) []string {
	var out []string
	for _, w := range wordsOf(text) {
		if len(w) >= 9 && strings.ContainsAny(w, "0123456789") && strings.ContainsAny(w, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			out = append(out, w)
		}
	}
	return out
}

// Every kind of name, and the value after it: the value goes, the name and the words around it stay.
func TestNamesAndTheirValues(t *testing.T) {
	rows := []struct{ in, want string }{
		// keys
		{"retry with password=hunter2 then", "retry with password=[redacted] then"},
		{"DB_PASSWORD: hunter2", "DB_PASSWORD: [redacted]"},
		{`{"apiKey": "hunter2", "user": "pat"}`, `{"apiKey": [redacted], "user": "pat"}`},
		{"x-api-key: hunter2", "x-api-key: [redacted]"},
		{"pwd=hunter2&next=1", "pwd=[redacted]&next=1"},
		{"passphrase = 'a long phrase' here", "passphrase = [redacted] here"},
		{"service.credentials=hunter2;other=1", "service.credentials=[redacted];other=1"},
		{"private_key: hunter2)", "private_key: [redacted])"},
		{"aws_secret_access_key = " + awsSecretKey, "aws_secret_access_key = [redacted]"},
		{"MY_TOKEN=hunter2 make release", "MY_TOKEN=[redacted] make release"},
		{"--db-password=hunter2", "--db-password=[redacted]"},
		// flags
		{"tool --password hunter2 --verbose", "tool --password [redacted] --verbose"},
		{"tool -token hunter2", "tool -token [redacted]"},
		{"login --client-secret hunter2 --region eu", "login --client-secret [redacted] --region eu"},
		{"x --api-key 'two words' y", "x --api-key [redacted] y"},
		{"x --auth-token\thunter2", "x --auth-token\t[redacted]"},
		{"tool --password --verbose", "tool --password --verbose"},
		// Authorization headers
		{"Authorization: Bearer hunter2xyz", "Authorization: Bearer [redacted]"},
		{"authorization: basic aHVudGVyMjpodW50ZXIy", "authorization: basic [redacted]"},
		{"Proxy-Authorization: Digest hunter2", "Proxy-Authorization: Digest [redacted]"},
		{"Authorization=hunter2", "Authorization=[redacted]"},
		{`curl -H "Authorization: token hunter2" x`, `curl -H "Authorization: token [redacted]" x`},
		// Bearer
		{"send it as a bearer hunter2hunter2 please", "send it as a bearer [redacted] please"},
		{"a bearer short", "a bearer short"},
		// extraheader
		{`git -c http.extraheader="AUTHORIZATION: basic aHVudGVyMg==" fetch`, `git -c http.extraheader=[redacted] fetch`},
		{"git -c http.extraHeader=hunter2 pull", "git -c http.extraHeader=[redacted] pull"},
	}
	for _, r := range rows {
		got := Text(r.in)
		if got != r.want {
			t.Errorf("Text(%q)\n got %q\nwant %q", r.in, got, r.want)
		}
		fixedAtEveryCut(t, r.in, got)
	}
}

// A quoted value ends at its closing quote on its line and takes what is glued after it; a quote never closed on
// its line runs to the line's end; a quote straight after another key's separator opens that key's value.
func TestQuotedValues(t *testing.T) {
	rows := []struct{ in, want string }{
		{`token: "two words" after`, `token: [redacted] after`},
		{`token: "two words"glued after`, `token: [redacted] after`},
		{`token: 'never closed and more`, `token: [redacted]`},
		{"token: 'never closed\nnext line", "token: [redacted]\nnext line"},
		{"token: \"closed on a later line\nlater\" x", "token: [redacted]\nlater\" x"},
		{`a SECRET: "x y z, then API_KEY="p q" end`, `a SECRET: [redacted]`},
		{`{"secret": "a, b", "kept": 1}`, `{"secret": [redacted], "kept": 1}`},
		{`--password "x"y z`, `--password [redacted] z`},
	}
	for _, r := range rows {
		got := Text(r.in)
		if got != r.want {
			t.Errorf("Text(%q)\n got %q\nwant %q", r.in, got, r.want)
		}
		fixedAtEveryCut(t, r.in, got)
	}
}

// A value on a later line than its name is taken, unless that line starts with a name on its own line, or is the
// text's last line and one word.
func TestValuesBelowTheirName(t *testing.T) {
	rows := []struct{ in, want string }{
		{"password:\n    hunter2\n", "password:\n    [redacted]\n"},
		{"password:\n\n  \"hunter2 x\"\nok: 1", "password:\n\n  [redacted]\nok: 1"},
		{"use a Bearer\n  hunter2hunter2\n", "use a Bearer\n  [redacted]\n"},
		{"vault:\n  token: hunter2", "vault:\n  token: [redacted]"},
		{"vault:\n  'secret': 'hunter2'\n", "vault:\n  'secret': [redacted]\n"},
		{"Authorization:\n  --password hunter2 x", "Authorization:\n  --password [redacted] x"},
		{"auth_token:\nAuthorization: Basic hunter2", "auth_token:\nAuthorization: Basic [redacted]"},
		// The last line, one word: it may be a name the recorder's cut left, so it is kept.
		{"vault:\n  passw", "vault:\n  passw"},
		{"password:\n  hunter2", "password:\n  hunter2"},
		// A flag takes nothing from the next line.
		{"run --token\nnext: line", "run --token\nnext: line"},
		// A word that only ends in a secret word is not a flag.
		{"see the release-token\nkey: value", "see the release-token\nkey: value"},
	}
	for _, r := range rows {
		got := Text(r.in)
		if got != r.want {
			t.Errorf("Text(%q)\n got %q\nwant %q", r.in, got, r.want)
		}
		fixedAtEveryCut(t, r.in, got)
	}
}

// What the redactor keeps unchanged.
func TestKeeps(t *testing.T) {
	m := newMaker(2)
	rows := []string{
		"merged at 3f1c2a9e8b7d6c5f4e3d2c1b0a9f8e7d6c5b4a3f and 1a2b3c4",
		"session 8d3c1f2e-77aa-4b0c-9e21-5f3a6b7c8d9e",
		"SESSION 8D3C1F2E-77AA-4B0C-9E21-5F3A6B7C8D9E",
		"guid: 0f1e2d3c4b5a69788796a5b4c3d2e1f0",
		"hex " + "C0FFEE0123456789abcdef0123456789ABCDEF01",
		"toolu_01" + m.secret(24),
		"srvtoolu_01" + m.secret(30),
		"tests/Engine.Tests/InventoryStacksMergeWhenTheSameItemArrives3Test.cs",
		"docs/notes/2026-10-10-the-redactor-and-its-tests.md",
		`C:/Users/someone/projects/sample-app/src/main.go`,
		"Two drafts, or three? I have 2 ready and a third half done.",
		"the token budget ran out, and the password field is empty",
		"a release-token list, the build-secret step, my-password notes",
		"x --token\nname: y",
		"password=[redacted]",
		"Authorization: Bearer [redacted]",
		"",
	}
	for _, r := range rows {
		if got := Text(r); got != r {
			t.Errorf("Text(%q) changed it to %q", r, got)
		}
	}
}

// A cut tail of the marker, or after an Authorization header of a scheme word, with only whitespace after it, is
// kept; anywhere else it is a value.
func TestCutTails(t *testing.T) {
	for _, kept := range []string{"pwd=[re", "the password=[redac", "--token [", "http.extraheader=[redacted",
		"Authorization: Ba", "Authorization: Bearer ", "authorization: \"Dig", "Authorization: token [red", "x: password: [r  "} {
		if got := Text(kept); got != kept {
			t.Errorf("Text(%q) = %q, want it kept", kept, got)
		}
	}
	for in, want := range map[string]string{
		"password=[red then":    "password=[redacted] then",
		"--token [r x":          "--token [redacted] x",
		"Authorization: Basics": "Authorization: [redacted]",
		"password=bearer":       "password=[redacted]",
	} {
		if got := Text(in); got != want {
			t.Errorf("Text(%q) = %q, want %q", in, got, want)
		}
	}
}

// A keyword matches in any case under simple case folding: the long s counts as s, the Kelvin sign as k.
func TestCaseFolding(t *testing.T) {
	rows := []struct{ in, want string }{
		{"PASSWORD=hunter2", "PASSWORD=[redacted]"},
		{"PaSsWoRd: hunter2", "PaSsWoRd: [redacted]"},
		{"pa\u017f\u017fword=hunter2", "pa\u017f\u017fword=[redacted]"},
		{"api_\u212aey: hunter2", "api_\u212aey: [redacted]"},
		{"to\u212aen=hunter2", "to\u212aen=[redacted]"},
		{"--pa\u017f\u017fword hunter2", "--pa\u017f\u017fword [redacted]"},
		{"BEARER hunter2hunter2", "BEARER [redacted]"},
		{"AUTHORIZATION: BEARER hunter2", "AUTHORIZATION: BEARER [redacted]"},
	}
	for _, r := range rows {
		got := Text(r.in)
		if got != r.want {
			t.Errorf("Text(%q)\n got %q\nwant %q", r.in, got, r.want)
		}
		fixedAtEveryCut(t, r.in, got)
	}
}

// Any character JavaScript's \s matches may stand between a name and its separator.
func TestEveryWhitespaceBeforeTheSeparator(t *testing.T) {
	spaces := []rune{'\t', '\v', '\f', ' ', 0xA0, 0x1680, 0x2000, 0x2005, 0x200A, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF}
	m := newMaker(3)
	for _, sp := range spaces {
		for _, sep := range []string{":", "="} {
			for _, name := range []string{"password", "x-api-key", "\"token\"", "authorization", "extraheader"} {
				if name == "extraheader" && sep == ":" {
					continue
				}
				secret := m.secret(14)
				in := fmt.Sprintf("set %s%c%s %s now", name, sp, sep, secret)
				checkGone(t, in, secret)
			}
		}
	}
}

// Find gives the spans Text replaces, with their rules' kinds.
func TestFind(t *testing.T) {
	in := "a password=hunter2 and " + awsKeyID + " and Authorization: Bearer hunter2hunter2"
	spans := Find(in)
	var b strings.Builder
	at := 0
	var kinds []string
	for _, sp := range spans {
		if sp.Start < at || sp.End <= sp.Start {
			t.Fatalf("spans out of order or touching: %+v", spans)
		}
		b.WriteString(in[at:sp.Start])
		b.WriteString(Marker)
		at = sp.End
		kinds = append(kinds, string(sp.Kind))
	}
	b.WriteString(in[at:])
	if got := b.String(); got != Text(in) {
		t.Errorf("Find's spans give %q, Text gives %q", got, Text(in))
	}
	if want := "secret-named-key aws-key-id authorization-header"; strings.Join(kinds, " ") != want {
		t.Errorf("kinds %q, want %q", strings.Join(kinds, " "), want)
	}
	if Find("nothing to see") != nil || Find("") != nil {
		t.Error("Find found something in clean text")
	}
	known := map[Kind]bool{}
	for _, k := range Kinds {
		known[k.Kind] = true
	}
	for _, r := range shapeRules {
		if !known[r.kind] {
			t.Errorf("kind %q is not in Kinds", r.kind)
		}
	}
	for _, k := range nameKinds {
		if !known[k] {
			t.Errorf("kind %q is not in Kinds", k)
		}
	}
}

// Text of a cut keeps a cut tail; Capped cuts after redacting, by characters.
func TestCapped(t *testing.T) {
	in := "x password=hunter2 " + strings.Repeat("é", 10)
	got := Capped(in, 15)
	if got != "x password=[red" {
		t.Errorf("Capped = %q", got)
	}
	if Text(got) != got {
		t.Errorf("a capped field changed on a second pass: %q", Text(got))
	}
	if c := Cut("ééé", 2); c != "éé" {
		t.Errorf("Cut by characters gave %q", c)
	}
	if c := Cut("abc", 0); c != "" {
		t.Errorf("Cut to 0 gave %q", c)
	}
}

// Text that is not valid UTF-8 keeps its bytes, and its secrets still go.
func TestInvalidUTF8(t *testing.T) {
	in := "a\xff\xfe password=hunter2 \xc3"
	got := Text(in)
	if got != "a\xff\xfe password=[redacted] \xc3" {
		t.Errorf("Text(%q) = %q", in, got)
	}
}
